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
	"time"

	"github.com/minio/selfupdate"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	releasesURL = "https://api.github.com/repos/Keyfactor/kfutil/releases"
	binaryName  = "kfutil"

	// maxBinaryBytes caps the extracted binary size to prevent OOM on malformed archives.
	maxBinaryBytes = 100 * 1024 * 1024 // 100 MiB
)

// apiClient is used for GitHub API metadata calls (short, bounded latency).
var apiClient = &http.Client{Timeout: 30 * time.Second}

// downloadClient is used for binary asset downloads (larger payloads).
var downloadClient = &http.Client{Timeout: 5 * time.Minute}

// allowedTokenHosts are the only hosts to which GITHUB_TOKEN may be forwarded.
var allowedTokenHosts = map[string]bool{
	"api.github.com":                true,
	"github.com":                    true,
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
		log.Warn().Err(err).
			Str("event", "upgrade.operator_resolution_failed").
			Msg("could not resolve OS user identity — audit logs will use 'unknown'")
		return "unknown"
	}
	return u.Username
}

// normalizeTag returns the tag as-is, or "latest" when the tag is empty,
// so log fields are unambiguous when no --version flag was passed.
func normalizeTag(tag string) string {
	if tag == "" {
		return "latest"
	}
	return tag
}

// sanitizeURL strips query-string parameters before a URL is written to a log
// field or console output. This prevents presigned CDN URLs (which carry
// time-limited credentials in their query strings) from appearing in logs.
func sanitizeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

// currentExecutable resolves the path to the running binary for audit log fields.
func currentExecutable(operator string) string {
	exe, err := os.Executable()
	if err != nil {
		log.Warn().Err(err).
			Str("event", "upgrade.executable_resolution_failed").
			Str("operator", operator).
			Msg("could not resolve executable path — audit logs will record 'kfutil' as executable")
		return "kfutil"
	}
	return exe
}

// Run fetches the target release, verifies the checksum, and replaces the
// running binary. targetVersion may be any valid GitHub tag (e.g. "v1.9.0")
// or empty to use the latest release.
func Run(currentVersion, targetVersion string, dryRun, force bool) error {
	operator := resolveOperator()

	log.Info().
		Str("event", "upgrade.run_started").
		Str("operator", operator).
		Str("current_version", currentVersion).
		Str("target_version", normalizeTag(targetVersion)).
		Bool("force", force).
		Bool("dry_run", dryRun).
		Msg("upgrade run initiated")

	release, err := fetchRelease(targetVersion, operator)
	if err != nil {
		log.Error().Err(err).
			Str("event", "upgrade.fetch_release_failed").
			Str("operator", operator).
			Str("tag", normalizeTag(targetVersion)).
			Msg("failed to fetch release metadata")
		return err
	}

	// Strip leading "v" for comparison with version.VERSION which has no prefix.
	releaseVer := strings.TrimPrefix(release.TagName, "v")
	currentVer := strings.TrimPrefix(currentVersion, "v")

	fmt.Printf("Current version : %s\n", currentVer)
	fmt.Printf("Target version  : %s\n", releaseVer)

	if !force && targetVersion == "" && releaseVer == currentVer {
		log.Info().
			Str("event", "upgrade.already_current").
			Str("operator", operator).
			Str("version", currentVer).
			Msg("binary is already at the latest version — no action taken")
		fmt.Println("Already at the latest version.")
		return nil
	}

	// Log whenever --force is set so the flag is always traceable in the audit record.
	if force {
		log.Warn().
			Str("event", "upgrade.force_override").
			Str("operator", operator).
			Str("current_version", currentVer).
			Str("target_version", releaseVer).
			Msg("--force flag set: safety checks may be bypassed")
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
		log.Error().
			Str("event", "upgrade.asset_not_found").
			Str("operator", operator).
			Str("tag", release.TagName).
			Str("asset_name", assetName).
			Str("os", runtime.GOOS).
			Str("arch", runtime.GOARCH).
			Msg("no matching release asset found for current platform")
		return fmt.Errorf("no release asset found for %s/%s (looked for %q)\navailable assets:\n%s",
			runtime.GOOS, runtime.GOARCH, assetName, listAssets(release.Assets))
	}

	if sumsURL == "" {
		log.Error().
			Str("event", "upgrade.sums_missing").
			Str("operator", operator).
			Str("tag", release.TagName).
			Msg("release has no SHA256SUMS asset — upgrade aborted")
		return fmt.Errorf("release %s has no SHA256SUMS asset — upgrade aborted", release.TagName)
	}

	exe := currentExecutable(operator)

	if dryRun {
		log.Info().
			Str("event", "upgrade.dry_run").
			Str("operator", operator).
			Str("from_version", currentVer).
			Str("to_version", releaseVer).
			Str("executable", exe).
			Str("source_url", sanitizeURL(archiveURL)).
			Msg("dry-run: no changes applied")
		fmt.Printf("\n[dry-run] Would download : %s\n", sanitizeURL(archiveURL))
		fmt.Printf("[dry-run] Would verify   : %s\n", sanitizeURL(sumsURL))
		fmt.Printf("[dry-run] Would replace  : %s\n", exe)
		return nil
	}

	fmt.Printf("Downloading %s ...\n", assetName)
	archiveData, err := download(archiveURL, operator)
	if err != nil {
		log.Error().Err(err).
			Str("event", "upgrade.download_failed").
			Str("operator", operator).
			Str("source_url", sanitizeURL(archiveURL)).
			Msg("archive download failed")
		return fmt.Errorf("download failed: %w", err)
	}

	binary, err := extractBinary(archiveData, binaryName)
	if err != nil {
		log.Error().Err(err).
			Str("event", "upgrade.extract_failed").
			Str("operator", operator).
			Str("source_url", sanitizeURL(archiveURL)).
			Msg("binary extraction from archive failed")
		return fmt.Errorf("extract failed: %w", err)
	}

	fmt.Println("Verifying checksum ...")
	sumsData, err := download(sumsURL, operator)
	if err != nil {
		log.Error().Err(err).
			Str("event", "upgrade.checksum_download_failed").
			Str("operator", operator).
			Str("sums_url", sanitizeURL(sumsURL)).
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
			Str("source_url", sanitizeURL(archiveURL)).
			Msg("checksum verification failed")
		return err
	}
	log.Info().
		Str("event", "upgrade.checksum_verified").
		Str("operator", operator).
		Str("asset", assetName).
		Str("source_url", sanitizeURL(archiveURL)).
		Msg("SHA-256 checksum verified successfully")
	fmt.Println("Checksum OK.")

	log.Info().
		Str("event", "upgrade.applying").
		Str("from_version", currentVer).
		Str("to_version", releaseVer).
		Str("executable", exe).
		Str("operator", operator).
		Str("source_url", sanitizeURL(archiveURL)).
		Bool("force", force).
		Msg("applying binary replacement")

	fmt.Println("Applying update ...")
	if err := apply(bytes.NewReader(binary), operator, currentVer, releaseVer, exe, sanitizeURL(archiveURL)); err != nil {
		failureReason := "apply_error"
		if os.IsPermission(err) {
			failureReason = "permission_denied"
		}
		log.Error().Err(err).
			Str("event", "upgrade.apply_failed").
			Str("from_version", currentVer).
			Str("to_version", releaseVer).
			Str("executable", exe).
			Str("operator", operator).
			Str("source_url", sanitizeURL(archiveURL)).
			Str("failure_reason", failureReason).
			Msg("binary replacement failed")
		return err
	}

	log.Info().
		Str("event", "upgrade.applied").
		Str("from_version", currentVer).
		Str("to_version", releaseVer).
		Str("executable", exe).
		Str("operator", operator).
		Str("source_url", sanitizeURL(archiveURL)).
		Msg("binary replacement complete")

	fmt.Printf("Upgraded to %s.\n", release.TagName)
	return nil
}

// fetchRelease returns the latest release or a specific tagged release.
func fetchRelease(tag, operator string) (*GitHubRelease, error) {
	return fetchReleaseFrom(releasesURL, tag, operator)
}

// fetchReleaseFrom is the testable core of fetchRelease; baseURL replaces releasesURL.
func fetchReleaseFrom(baseURL, tag, operator string) (*GitHubRelease, error) {
	var reqURL string
	if tag == "" {
		reqURL = baseURL + "/latest"
	} else {
		// URL-encode the tag so path-separator or query characters in user
		// input cannot alter the request URL structure.
		reqURL = baseURL + "/tags/" + url.PathEscape(tag)
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		log.Warn().Err(err).
			Str("event", "upgrade.github_api_request_build_failed").
			Str("operator", operator).
			Str("url", sanitizeURL(reqURL)).
			Msg("failed to construct GitHub API request")
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	tokenPresent := false
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		tokenPresent = true
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	start := time.Now()
	resp, err := apiClient.Do(req)
	if err != nil {
		log.Error().Err(err).
			Str("event", "upgrade.github_api_network_error").
			Str("operator", operator).
			Str("url", sanitizeURL(reqURL)).
			Str("method", http.MethodGet).
			Int64("latency_ms", time.Since(start).Milliseconds()).
			Bool("github_token_present", tokenPresent).
			Msg("GitHub API network request failed")
		return nil, fmt.Errorf("GitHub API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Two events fire for non-200 responses: github_api_response (Warn) captures
	// transport telemetry (latency, headers); github_api_rejected (Error) captures
	// the control decision. Both are intentional — removing the Warn event would
	// drop latency_ms from the audit trail.
	var ev *zerolog.Event
	if resp.StatusCode >= 400 {
		ev = log.Warn()
	} else {
		ev = log.Info()
	}
	ev.Str("event", "upgrade.github_api_response").
		Str("url", sanitizeURL(reqURL)).
		Str("method", http.MethodGet).
		Int("status_code", resp.StatusCode).
		Int64("latency_ms", time.Since(start).Milliseconds()).
		Bool("github_token_present", tokenPresent).
		Str("operator", operator).
		Msg("GitHub API response received")

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		var errMsg string
		if tag != "" {
			errMsg = fmt.Sprintf("release tag %q not found", tag)
		} else {
			errMsg = "no releases found for this repository"
		}
		log.Error().
			Str("event", "upgrade.github_api_rejected").
			Str("operator", operator).
			Int("status_code", resp.StatusCode).
			Str("url", sanitizeURL(reqURL)).
			Msg("GitHub API rejected the upgrade request")
		return nil, fmt.Errorf("%s", errMsg)
	case http.StatusForbidden, http.StatusTooManyRequests:
		log.Error().
			Str("event", "upgrade.github_api_rejected").
			Str("operator", operator).
			Int("status_code", resp.StatusCode).
			Str("url", sanitizeURL(reqURL)).
			Msg("GitHub API rejected the upgrade request")
		return nil, fmt.Errorf("GitHub API rate limited (HTTP %d) — set GITHUB_TOKEN to increase limits", resp.StatusCode)
	default:
		log.Error().
			Str("event", "upgrade.github_api_rejected").
			Str("operator", operator).
			Int("status_code", resp.StatusCode).
			Str("url", sanitizeURL(reqURL)).
			Msg("GitHub API rejected the upgrade request")
		return nil, fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var rel GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		log.Error().Err(err).
			Str("event", "upgrade.release_parse_failed").
			Str("operator", operator).
			Str("url", sanitizeURL(reqURL)).
			Int("status_code", resp.StatusCode).
			Msg("failed to parse GitHub release response body")
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
func download(rawURL, operator string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		log.Warn().Err(err).
			Str("event", "upgrade.http_request_build_failed").
			Str("operator", operator).
			Str("url", sanitizeURL(rawURL)).
			Msg("failed to construct HTTP request")
		return nil, err
	}

	tokenPresent := false
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		if parsed, err := url.Parse(rawURL); err == nil && allowedTokenHosts[parsed.Hostname()] {
			tokenPresent = true
			req.Header.Set("Authorization", "Bearer "+tok)
		}
	}

	start := time.Now()
	resp, err := downloadClient.Do(req)
	if err != nil {
		log.Error().Err(err).
			Str("event", "upgrade.http_network_error").
			Str("operator", operator).
			Str("url", sanitizeURL(rawURL)).
			Str("method", http.MethodGet).
			Int64("latency_ms", time.Since(start).Milliseconds()).
			Bool("github_token_present", tokenPresent).
			Msg("HTTP network request failed")
		return nil, err
	}
	defer resp.Body.Close()

	// Two events fire for non-200 responses: http_response (Warn) captures
	// transport telemetry (latency, headers); http_request_failed (Error) captures
	// the control decision. Both are intentional — removing the Warn event would
	// drop latency_ms from the audit trail.
	var ev *zerolog.Event
	if resp.StatusCode >= 400 {
		ev = log.Warn()
	} else {
		ev = log.Info()
	}
	ev.Str("event", "upgrade.http_response").
		Str("url", sanitizeURL(rawURL)).
		Str("method", http.MethodGet).
		Int("status_code", resp.StatusCode).
		Int64("latency_ms", time.Since(start).Milliseconds()).
		Bool("github_token_present", tokenPresent).
		Str("operator", operator).
		Msg("HTTP response received")

	if resp.StatusCode != http.StatusOK {
		log.Error().
			Str("event", "upgrade.http_request_failed").
			Str("operator", operator).
			Int("status_code", resp.StatusCode).
			Str("url", sanitizeURL(rawURL)).
			Msg("HTTP download request returned non-success status")
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, sanitizeURL(rawURL))
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBinaryBytes))
	if err != nil {
		return nil, err
	}
	// LimitReader silently caps at maxBinaryBytes; detect and reject truncated payloads.
	if int64(len(data)) >= maxBinaryBytes {
		log.Error().
			Str("event", "upgrade.download_size_limit_reached").
			Str("operator", operator).
			Str("url", sanitizeURL(rawURL)).
			Int64("limit_bytes", maxBinaryBytes).
			Msg("HTTP response body reached size cap — download rejected")
		return nil, fmt.Errorf("response body from %s exceeded maximum allowed size (%d bytes)", sanitizeURL(rawURL), maxBinaryBytes)
	}
	return data, nil
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
				// LimitReader prevents OOM on malformed or zip-bomb archives.
				data, err := io.ReadAll(io.LimitReader(rc, maxBinaryBytes))
				if err != nil {
					return nil, err
				}
				if int64(len(data)) >= maxBinaryBytes {
					return nil, fmt.Errorf("binary %q exceeds maximum allowed size (%d bytes)", name, maxBinaryBytes)
				}
				return data, nil
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
func apply(binary io.Reader, operator, fromVersion, toVersion, exe, sourceURL string) error {
	log.Info().
		Str("event", "upgrade.apply_started").
		Str("operator", operator).
		Str("from_version", fromVersion).
		Str("to_version", toVersion).
		Str("executable", exe).
		Str("source_url", sourceURL).
		Msg("binary write commencing via selfupdate")
	err := selfupdate.Apply(binary, selfupdate.Options{})
	if err != nil {
		if rbErr := selfupdate.RollbackError(err); rbErr != nil {
			log.Error().Err(rbErr).
				Str("event", "upgrade.rollback_failed").
				Str("operator", operator).
				Str("from_version", fromVersion).
				Str("to_version", toVersion).
				Str("executable", exe).
				Str("source_url", sourceURL).
				Msg("binary replacement failed and rollback also failed — binary may be corrupted")
			return fmt.Errorf("upgrade failed and rollback also failed: %w", rbErr)
		}
		// Rollback succeeded — log at Warn so auditors can distinguish a
		// recovery action from routine Info events.
		log.Warn().
			Str("event", "upgrade.rollback_succeeded").
			Str("operator", operator).
			Str("from_version", fromVersion).
			Str("to_version", toVersion).
			Str("executable", exe).
			Str("source_url", sourceURL).
			Msg("apply failed but binary was successfully rolled back to prior version")
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied writing to %s\nTry re-running with elevated privileges (sudo kfutil upgrade)", exe)
		}
		return fmt.Errorf("upgrade failed: %w", err)
	}
	return nil
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
