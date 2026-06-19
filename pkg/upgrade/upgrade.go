// Copyright 2026 Keyfactor
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package upgrade

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"runtime"
	"strings"

	"github.com/minio/selfupdate"
	"github.com/rs/zerolog/log"
)

const (
	releasesURL = "https://api.github.com/repos/Keyfactor/kfutil/releases"
	binaryName  = "kfutil"
)

// allowedTokenHosts are the only hosts to which GITHUB_TOKEN may be forwarded.
var allowedTokenHosts = map[string]bool{
	"api.github.com":               true,
	"github.com":                   true,
	"objects.githubusercontent.com": true,
}

// GitHubRelease is the minimal shape we need from the GitHub releases API.
type GitHubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []GitHubAsset `json:"assets"`
}

type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// resolveOperator returns the current OS user name for audit log fields.
func resolveOperator() string {
	u, err := user.Current()
	if err != nil {
		return "unknown"
	}
	return u.Username
}

// Run fetches the target release, verifies the checksum, and replaces the
// running binary. targetVersion may be any valid GitHub tag (e.g. "v1.9.0")
// or empty to use the latest release.
func Run(currentVersion, targetVersion string, dryRun, force bool) error {
	operator := resolveOperator()

	release, err := fetchRelease(targetVersion)
	if err != nil {
		log.Error().Err(err).
			Str("event", "upgrade.fetch_release_failed").
			Str("operator", operator).
			Msg("failed to fetch release metadata")
		return err
	}

	// Strip leading "v" for comparison with version.VERSION which has no prefix.
	releaseVer := strings.TrimPrefix(release.TagName, "v")
	currentVer := strings.TrimPrefix(currentVersion, "v")

	fmt.Printf("Current version : %s\n", currentVer)
	fmt.Printf("Target version  : %s\n", releaseVer)

	if !force && targetVersion == "" && releaseVer == currentVer {
		fmt.Println("Already at the latest version.")
		return nil
	}

	// Log when --force bypasses the version-match safety check.
	if force && targetVersion == "" && releaseVer == currentVer {
		log.Warn().
			Str("event", "upgrade.force_override").
			Str("operator", operator).
			Str("version", releaseVer).
			Msg("version-match check bypassed via --force")
	}

	assetName := archiveAssetName(release.TagName)
	sumsName := fmt.Sprintf("kfutil_%s_SHA256SUMS", strings.TrimPrefix(release.TagName, "v"))

	archiveURL, sumsURL := "", ""
	for _, a := range release.Assets {
		switch a.Name {
		case assetName:
			archiveURL = a.BrowserDownloadURL
		case sumsName:
			sumsURL = a.BrowserDownloadURL
		}
	}

	if archiveURL == "" {
		return fmt.Errorf("no release asset found for %s/%s (looked for %q)\navailable assets:\n%s",
			runtime.GOOS, runtime.GOARCH, assetName, listAssets(release.Assets))
	}

	if sumsURL == "" {
		return fmt.Errorf("release %s has no SHA256SUMS asset — upgrade aborted", release.TagName)
	}

	if dryRun {
		fmt.Printf("\n[dry-run] Would download : %s\n", archiveURL)
		fmt.Printf("[dry-run] Would verify   : %s\n", sumsURL)
		fmt.Printf("[dry-run] Would replace  : %s\n", currentExecutable())
		return nil
	}

	fmt.Printf("Downloading %s ...\n", assetName)
	archiveData, err := download(archiveURL)
	if err != nil {
		log.Error().Err(err).
			Str("event", "upgrade.download_failed").
			Str("url", archiveURL).
			Str("operator", operator).
			Msg("archive download failed")
		return fmt.Errorf("download failed: %w", err)
	}

	binary, err := extractBinary(archiveData, binaryName)
	if err != nil {
		log.Error().Err(err).
			Str("event", "upgrade.extract_failed").
			Str("operator", operator).
			Msg("binary extraction from archive failed")
		return fmt.Errorf("extract failed: %w", err)
	}

	fmt.Println("Verifying checksum ...")
	sumsData, err := download(sumsURL)
	if err != nil {
		log.Error().Err(err).
			Str("event", "upgrade.checksum_download_failed").
			Str("url", sumsURL).
			Str("operator", operator).
			Msg("SHA256SUMS download failed")
		return fmt.Errorf("checksum download failed: %w", err)
	}
	// Verify the hash of the zip archive, not the extracted binary —
	// goreleaser's SHA256SUMS records hashes of the zip archives.
	if err := verifyChecksum(archiveData, assetName, sumsData); err != nil {
		log.Error().Err(err).
			Str("event", "upgrade.checksum_mismatch").
			Str("asset", assetName).
			Str("operator", operator).
			Msg("checksum verification failed")
		return err
	}
	fmt.Println("Checksum OK.")

	exe := currentExecutable()
	log.Info().
		Str("event", "upgrade.applying").
		Str("from_version", currentVer).
		Str("to_version", releaseVer).
		Str("executable", exe).
		Str("operator", operator).
		Str("source_url", archiveURL).
		Msg("applying binary replacement")

	fmt.Println("Applying update ...")
	if err := apply(bytes.NewReader(binary)); err != nil {
		log.Error().Err(err).
			Str("event", "upgrade.apply_failed").
			Str("from_version", currentVer).
			Str("to_version", releaseVer).
			Str("executable", exe).
			Str("operator", operator).
			Msg("binary replacement failed")
		return err
	}

	log.Info().
		Str("event", "upgrade.applied").
		Str("from_version", currentVer).
		Str("to_version", releaseVer).
		Str("executable", exe).
		Str("operator", operator).
		Str("source_url", archiveURL).
		Msg("binary replacement complete")

	fmt.Printf("Upgraded to %s.\n", release.TagName)
	return nil
}

// fetchRelease returns the latest release or a specific tagged release.
func fetchRelease(tag string) (*GitHubRelease, error) {
	return fetchReleaseFrom(releasesURL, tag)
}

// fetchReleaseFrom is the testable core of fetchRelease; baseURL replaces releasesURL.
func fetchReleaseFrom(baseURL, tag string) (*GitHubRelease, error) {
	var reqURL string
	if tag == "" {
		reqURL = baseURL + "/latest"
	} else {
		reqURL = baseURL + "/tags/" + tag
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub API request failed: %w", err)
	}
	defer resp.Body.Close()

	log.Debug().
		Str("event", "upgrade.github_api_response").
		Str("url", reqURL).
		Str("method", http.MethodGet).
		Int("status_code", resp.StatusCode).
		Msg("GitHub API response received")

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		if tag != "" {
			return nil, fmt.Errorf("release tag %q not found", tag)
		}
		return nil, fmt.Errorf("no releases found for this repository")
	case http.StatusForbidden, http.StatusTooManyRequests:
		return nil, fmt.Errorf("GitHub API rate limited (HTTP %d) — set GITHUB_TOKEN to increase limits", resp.StatusCode)
	default:
		return nil, fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var rel GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("failed to parse release response: %w", err)
	}
	return &rel, nil
}

// archiveAssetName returns the goreleaser archive name for the current platform.
// Pattern: kfutil_<version>_<os>_<arch>.zip  (version without leading "v")
func archiveAssetName(tag string) string {
	ver := strings.TrimPrefix(tag, "v")
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	return fmt.Sprintf("kfutil_%s_%s_%s.zip", ver, goos, goarch)
}

// download fetches a URL and returns the body bytes. GITHUB_TOKEN is only
// forwarded to hosts in allowedTokenHosts to prevent token exfiltration via
// a tampered BrowserDownloadURL.
func download(rawURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		if parsed, err := url.Parse(rawURL); err == nil && allowedTokenHosts[parsed.Hostname()] {
			req.Header.Set("Authorization", "Bearer "+tok)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Debug().
		Str("event", "upgrade.http_response").
		Str("url", rawURL).
		Str("method", http.MethodGet).
		Int("status_code", resp.StatusCode).
		Msg("HTTP response received")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, rawURL)
	}
	return io.ReadAll(resp.Body)
}

// extractBinary unpacks the binary named <name> or <name>.exe from a zip archive.
func extractBinary(zipData []byte, name string) ([]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("invalid zip archive: %w", err)
	}

	targets := []string{name, name + ".exe"}
	for _, f := range r.File {
		for _, t := range targets {
			if f.Name == t {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				defer rc.Close()
				return io.ReadAll(rc)
			}
		}
	}
	return nil, fmt.Errorf("binary %q not found in archive", name)
}

// verifyChecksum checks the SHA-256 of archiveData against the goreleaser SUMS file.
// The SUMS file has lines: "<hex>  <assetname.zip>"
// A missing entry is an error — the SUMS file must cover the target asset.
func verifyChecksum(archiveData []byte, assetName string, sumsData []byte) error {
	h := sha256.Sum256(archiveData)
	got := hex.EncodeToString(h[:])

	for _, line := range strings.Split(string(sumsData), "\n") {
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == assetName {
			if parts[0] != got {
				return fmt.Errorf("checksum mismatch for %s\n  expected: %s\n  got:      %s", assetName, parts[0], got)
			}
			return nil
		}
	}
	return fmt.Errorf("no checksum entry found for %s in SHA256SUMS", assetName)
}

// apply replaces the running binary using minio/selfupdate.
func apply(binary io.Reader) error {
	err := selfupdate.Apply(binary, selfupdate.Options{})
	if err != nil {
		if rbErr := selfupdate.RollbackError(err); rbErr != nil {
			return fmt.Errorf("upgrade failed and rollback also failed: %w", rbErr)
		}
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied writing to %s\nTry re-running with elevated privileges (sudo kfutil upgrade)", currentExecutable())
		}
		return fmt.Errorf("upgrade failed: %w", err)
	}
	return nil
}

func currentExecutable() string {
	exe, err := os.Executable()
	if err != nil {
		return "kfutil"
	}
	return exe
}

func listAssets(assets []GitHubAsset) string {
	var sb strings.Builder
	for _, a := range assets {
		sb.WriteString("  ")
		sb.WriteString(a.Name)
		sb.WriteString("\n")
	}
	return sb.String()
}
