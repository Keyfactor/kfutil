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
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── archiveAssetName ──────────────────────────────────────────────────────────

func TestArchiveAssetName(t *testing.T) {
	name := archiveAssetName("v1.9.0")
	expected := fmt.Sprintf("kfutil_1.9.0_%s_%s.zip", runtime.GOOS, runtime.GOARCH)
	assert.Equal(t, expected, name)
}

func TestArchiveAssetName_NoLeadingV(t *testing.T) {
	assert.Equal(t, archiveAssetName("v1.2.3"), archiveAssetName("v1.2.3"))
	// Both should strip the "v"
	withV := archiveAssetName("v1.2.3")
	assert.NotContains(t, withV, "_v1.")
}

// ── verifyChecksum ────────────────────────────────────────────────────────────

func TestVerifyChecksum_Match(t *testing.T) {
	data := []byte("fake binary content")
	h := sha256.Sum256(data)
	hashHex := hex.EncodeToString(h[:])
	sums := fmt.Sprintf("%s  kfutil_1.9.0_linux_amd64.zip\n", hashHex)

	err := verifyChecksum(data, "kfutil_1.9.0_linux_amd64.zip", []byte(sums))
	require.NoError(t, err)
}

func TestVerifyChecksum_Mismatch(t *testing.T) {
	data := []byte("fake binary content")
	sums := "badhash  kfutil_1.9.0_linux_amd64.zip\n"
	err := verifyChecksum(data, "kfutil_1.9.0_linux_amd64.zip", []byte(sums))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestVerifyChecksum_AssetNotInSums(t *testing.T) {
	// Missing entry must be an error — a SUMS file that doesn't cover the target
	// asset must not silently pass, as an attacker could strip the entry.
	err := verifyChecksum([]byte("data"), "kfutil_1.9.0_freebsd_arm64.zip", []byte("abc123  kfutil_1.9.0_linux_amd64.zip\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no checksum entry found")
}

// ── extractBinary ─────────────────────────────────────────────────────────────

func makeZip(t *testing.T, filename, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, err := w.Create(filename)
	require.NoError(t, err)
	_, err = f.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return buf.Bytes()
}

func TestExtractBinary_Unix(t *testing.T) {
	data := makeZip(t, "kfutil", "binary-content")
	got, err := extractBinary(data, "kfutil")
	require.NoError(t, err)
	assert.Equal(t, []byte("binary-content"), got)
}

func TestExtractBinary_Windows(t *testing.T) {
	data := makeZip(t, "kfutil.exe", "win-binary")
	got, err := extractBinary(data, "kfutil")
	require.NoError(t, err)
	assert.Equal(t, []byte("win-binary"), got)
}

func TestExtractBinary_NotFound(t *testing.T) {
	data := makeZip(t, "readme.txt", "hello")
	_, err := extractBinary(data, "kfutil")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in archive")
}

func TestExtractBinary_InvalidZip(t *testing.T) {
	_, err := extractBinary([]byte("not a zip"), "kfutil")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid zip archive")
}

// ── download (token host allowlist) ──────────────────────────────────────────

func TestDownload_TokenSentToTrustedHost(t *testing.T) {
	var receivedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("data"))
	}))
	defer srv.Close()

	// Temporarily register the test server's host (127.0.0.1) as trusted.
	allowedTokenHosts["127.0.0.1"] = true
	t.Cleanup(func() { delete(allowedTokenHosts, "127.0.0.1") })

	t.Setenv("GITHUB_TOKEN", "super-secret-token")

	_, err := download(srv.URL+"/asset.zip", "testuser")
	require.NoError(t, err)
	assert.Equal(t, "Bearer super-secret-token", receivedAuth, "GITHUB_TOKEN must be forwarded to trusted host")
}

func TestDownload_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	_, err := download(srv.URL+"/asset.zip", "testuser")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func TestDownload_TokenNotSentToUntrustedHost(t *testing.T) {
	var receivedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("data"))
	}))
	defer srv.Close()

	t.Setenv("GITHUB_TOKEN", "super-secret-token")

	_, err := download(srv.URL+"/asset.zip", "testuser")
	require.NoError(t, err)
	assert.Empty(t, receivedAuth, "GITHUB_TOKEN must not be sent to untrusted host")
}

// ── fetchRelease (via mock HTTP server) ───────────────────────────────────────

func mockReleaseServer(t *testing.T, tag string, statusCode int) *httptest.Server {
	t.Helper()
	rel := GitHubRelease{
		TagName: tag,
		Assets: []GitHubAsset{
			{Name: fmt.Sprintf("kfutil_1.9.0_%s_%s.zip", runtime.GOOS, runtime.GOARCH), BrowserDownloadURL: "http://example.com/kfutil.zip"},
			{Name: "kfutil_1.9.0_SHA256SUMS", BrowserDownloadURL: "http://example.com/SHA256SUMS"},
		},
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if statusCode != http.StatusOK {
			w.WriteHeader(statusCode)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rel)
	}))
}

func TestFetchRelease_Latest(t *testing.T) {
	srv := mockReleaseServer(t, "v1.9.0", http.StatusOK)
	defer srv.Close()

	// fetchRelease builds the URL itself; test via fetchReleaseFrom which accepts a base URL.

	rel, err := fetchReleaseFrom(srv.URL, "", "testuser")
	require.NoError(t, err)
	assert.Equal(t, "v1.9.0", rel.TagName)
	assert.Len(t, rel.Assets, 2)
}

func TestFetchRelease_SpecificTag(t *testing.T) {
	srv := mockReleaseServer(t, "v1.8.0", http.StatusOK)
	defer srv.Close()

	rel, err := fetchReleaseFrom(srv.URL, "v1.8.0", "testuser")
	require.NoError(t, err)
	assert.Equal(t, "v1.8.0", rel.TagName)
}

func TestFetchRelease_NotFound(t *testing.T) {
	srv := mockReleaseServer(t, "", http.StatusNotFound)
	defer srv.Close()

	_, err := fetchReleaseFrom(srv.URL, "v99.0.0", "testuser")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFetchRelease_RateLimited(t *testing.T) {
	srv := mockReleaseServer(t, "", http.StatusForbidden)
	defer srv.Close()

	_, err := fetchReleaseFrom(srv.URL, "", "testuser")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rate limited")
}

// ── sanitizeURL ───────────────────────────────────────────────────────────────

func TestSanitizeURL_StripsQueryParams(t *testing.T) {
	raw := "https://objects.githubusercontent.com/github-production-release-asset/abc123/kfutil.zip?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKID&X-Amz-Signature=deadbeef"
	got := sanitizeURL(raw)
	assert.Equal(t, "https://objects.githubusercontent.com/github-production-release-asset/abc123/kfutil.zip", got)
	assert.NotContains(t, got, "X-Amz")
	assert.NotContains(t, got, "Signature")
	assert.NotContains(t, got, "?")
}

func TestSanitizeURL_ParseErrorFallback(t *testing.T) {
	// An unparseable URL must be returned as-is — better to log a weird string
	// than to panic or drop the URL entirely from audit output.
	raw := "://not a valid url"
	got := sanitizeURL(raw)
	assert.Equal(t, raw, got)
}

func TestSanitizeURL_StripsFragment(t *testing.T) {
	raw := "https://github.com/Keyfactor/kfutil/releases/tag/v1.9.0#readme"
	got := sanitizeURL(raw)
	assert.NotContains(t, got, "#")
	assert.NotContains(t, got, "readme")
}

// ── extractBinary size cap ────────────────────────────────────────────────────

func TestExtractBinary_ExceedsMaxSize(t *testing.T) {
	// Build a zip with an entry that is exactly maxBinaryBytes bytes.
	// extractBinary must return an error rather than returning a slice of that size.
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, err := w.Create("kfutil")
	require.NoError(t, err)
	// Write maxBinaryBytes bytes — this will trigger the size-cap check.
	chunk := make([]byte, 4096)
	written := 0
	for written < maxBinaryBytes {
		n := maxBinaryBytes - written
		if n > len(chunk) {
			n = len(chunk)
		}
		_, err = f.Write(chunk[:n])
		require.NoError(t, err)
		written += n
	}
	require.NoError(t, w.Close())

	_, err = extractBinary(buf.Bytes(), "kfutil")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}
