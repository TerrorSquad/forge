package update

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ---------- helpers ----------

func fakeRelease(tag string, assets []Asset) []byte {
	r := Release{TagName: tag, Assets: assets}
	b, _ := json.Marshal(r)
	return b
}

func serveRelease(t *testing.T, body []byte) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// ---------- platformAssetName ----------

func TestPlatformAssetName_ContainsOS(t *testing.T) {
	name := platformAssetName()
	if !strings.Contains(name, runtime.GOOS) {
		t.Errorf("expected OS %q in asset name %q", runtime.GOOS, name)
	}
}

func TestPlatformAssetName_ContainsArch(t *testing.T) {
	name := platformAssetName()
	if !strings.Contains(name, runtime.GOARCH) {
		t.Errorf("expected arch %q in asset name %q", runtime.GOARCH, name)
	}
}

// ---------- findAssetURLs ----------

func TestFindAssetURLs_FindsBinaryAndChecksum(t *testing.T) {
	assets := []Asset{
		{Name: "forge_linux_amd64", BrowserDownloadURL: "https://example.com/forge_linux_amd64"},
		{Name: "forge_checksums.txt", BrowserDownloadURL: "https://example.com/checksums.txt"},
	}
	bin, cksum := findAssetURLs(assets, "forge_linux_amd64")
	if bin != "https://example.com/forge_linux_amd64" {
		t.Errorf("unexpected binary URL: %q", bin)
	}
	if cksum != "https://example.com/checksums.txt" {
		t.Errorf("unexpected checksum URL: %q", cksum)
	}
}

func TestFindAssetURLs_NoMatch(t *testing.T) {
	assets := []Asset{
		{Name: "forge_windows_amd64.zip", BrowserDownloadURL: "https://example.com/forge_windows_amd64.zip"},
	}
	bin, _ := findAssetURLs(assets, "forge_linux_amd64")
	if bin != "" {
		t.Errorf("expected no match, got %q", bin)
	}
}

func TestFindAssetURLs_ArchiveSuffixAllowed(t *testing.T) {
	assets := []Asset{
		{Name: "forge_linux_amd64.tar.gz", BrowserDownloadURL: "https://example.com/forge_linux_amd64.tar.gz"},
	}
	bin, _ := findAssetURLs(assets, "forge_linux_amd64")
	if bin == "" {
		t.Error("expected match for archive with .tar.gz suffix")
	}
}

// ---------- fetchRelease ----------

func TestFetchRelease_UsesEnvOverride(t *testing.T) {
	body := fakeRelease("v9.9.9", nil)
	srv := serveRelease(t, body)
	t.Setenv(envUpdateURL, srv.URL)

	rel, err := fetchRelease("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rel.TagName != "v9.9.9" {
		t.Errorf("got tag %q, want v9.9.9", rel.TagName)
	}
}

func TestFetchRelease_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	t.Setenv(envUpdateURL, srv.URL)

	_, err := fetchRelease("")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestFetchRelease_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	t.Cleanup(srv.Close)
	t.Setenv(envUpdateURL, srv.URL)

	_, err := fetchRelease("")
	if err == nil {
		t.Fatal("expected JSON decode error")
	}
}

// ---------- checksum verification ----------

func TestVerifyChecksum_Match(t *testing.T) {
	content := []byte("hello forge")
	tmp := filepath.Join(t.TempDir(), "binary")
	if err := os.WriteFile(tmp, content, 0o644); err != nil {
		t.Fatal(err)
	}
	checksumBody := sha256Hex(content) + "  forge_linux_amd64\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(checksumBody))
	}))
	t.Cleanup(srv.Close)
	if err := verifyChecksum(tmp, srv.URL, "forge_linux_amd64"); err != nil {
		t.Errorf("expected checksum match, got error: %v", err)
	}
}

func TestVerifyChecksum_Mismatch(t *testing.T) {
	content := []byte("hello forge")
	tmp := filepath.Join(t.TempDir(), "binary")
	if err := os.WriteFile(tmp, content, 0o644); err != nil {
		t.Fatal(err)
	}
	checksumBody := "0000000000000000000000000000000000000000000000000000000000000000  forge_linux_amd64\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(checksumBody))
	}))
	t.Cleanup(srv.Close)
	err := verifyChecksum(tmp, srv.URL, "forge_linux_amd64")
	if err == nil {
		t.Fatal("expected checksum mismatch error")
	}
}

func TestVerifyChecksum_NoEntrySkips(t *testing.T) {
	content := []byte("hello forge")
	tmp := filepath.Join(t.TempDir(), "binary")
	if err := os.WriteFile(tmp, content, 0o644); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("abc123  forge_darwin_arm64\n"))
	}))
	t.Cleanup(srv.Close)
	if err := verifyChecksum(tmp, srv.URL, "forge_linux_amd64"); err != nil {
		t.Errorf("expected skip (no entry), got: %v", err)
	}
}

// ---------- Run --check ----------

func TestRun_CheckOnly_UpToDate(t *testing.T) {
	body := fakeRelease("v1.2.3", nil)
	srv := serveRelease(t, body)
	t.Setenv(envUpdateURL, srv.URL)

	var buf bytes.Buffer
	err := Run(Options{CheckOnly: true, CurrentVersion: "v1.2.3"}, &buf)
	if err != nil {
		t.Errorf("expected no error when up to date, got: %v", err)
	}
	if !strings.Contains(buf.String(), "up to date") {
		t.Errorf("expected 'up to date' in output, got: %q", buf.String())
	}
}

func TestRun_CheckOnly_UpdateAvailable(t *testing.T) {
	body := fakeRelease("v2.0.0", nil)
	srv := serveRelease(t, body)
	t.Setenv(envUpdateURL, srv.URL)

	var buf bytes.Buffer
	err := Run(Options{CheckOnly: true, CurrentVersion: "v1.0.0"}, &buf)
	if err == nil {
		t.Fatal("expected ErrUpdateAvailable")
	}
	if !strings.Contains(buf.String(), "v2.0.0 available") {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestRun_AlreadyLatest(t *testing.T) {
	body := fakeRelease("v1.5.0", nil)
	srv := serveRelease(t, body)
	t.Setenv(envUpdateURL, srv.URL)

	var buf bytes.Buffer
	err := Run(Options{CurrentVersion: "v1.5.0"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "nothing to do") {
		t.Errorf("expected 'nothing to do', got: %q", buf.String())
	}
}

func TestRun_NoAssetForPlatform(t *testing.T) {
	body := fakeRelease("v2.0.0", []Asset{
		{Name: "forge_plan9_mips64", BrowserDownloadURL: "https://example.com/forge_plan9_mips64"},
	})
	srv := serveRelease(t, body)
	t.Setenv(envUpdateURL, srv.URL)

	var buf bytes.Buffer
	err := Run(Options{CurrentVersion: "v1.0.0"}, &buf)
	if err == nil {
		t.Fatal("expected error for missing platform asset")
	}
	if !strings.Contains(err.Error(), "no release asset found") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- ErrUpdateAvailable ----------

func TestErrUpdateAvailable_IsDistinct(t *testing.T) {
	if ErrUpdateAvailable == nil {
		t.Fatal("ErrUpdateAvailable must not be nil")
	}
	if !strings.Contains(ErrUpdateAvailable.Error(), "update available") {
		t.Errorf("unexpected error string: %v", ErrUpdateAvailable)
	}
}
