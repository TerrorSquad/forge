package update

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	defaultReleasesURL = "https://api.github.com/repos/TerrorSquad/forge/releases/latest"
	envUpdateURL       = "FORGE_UPDATE_URL"
	httpTimeout        = 30 * time.Second
)

// Release holds the subset of the GitHub Releases API response that we need.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents one downloadable file in a release.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Options controls the behaviour of Run.
type Options struct {
	// CheckOnly prints the latest version without downloading anything.
	CheckOnly bool
	// Version pins to a specific release tag (e.g. "v1.3.2"). Empty = latest.
	Version string
	// Rollback restores the previous binary backup.
	Rollback bool
	// CurrentVersion is the version string of the running binary.
	CurrentVersion string
}

// Run is the entry-point for the update command.
func Run(opts Options, w io.Writer) error {
	if opts.Rollback {
		return rollback(w)
	}

	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine binary path: %w", err)
	}

	release, err := fetchRelease(opts.Version)
	if err != nil {
		return fmt.Errorf("fetching release info: %w", err)
	}

	latest := release.TagName

	if opts.CheckOnly {
		current := opts.CurrentVersion
		if current == "" {
			current = "unknown"
		}
		if current == latest {
			fmt.Fprintf(w, "%s is up to date\n", latest)
		} else {
			fmt.Fprintf(w, "%s available (you have %s)\n", latest, current)
			return ErrUpdateAvailable
		}
		return nil
	}

	if opts.CurrentVersion != "" && opts.CurrentVersion == latest {
		fmt.Fprintf(w, "Already at %s — nothing to do.\n", latest)
		return nil
	}

	fmt.Fprintf(w, "Current version: %s\n", opts.CurrentVersion)
	fmt.Fprintf(w, "Latest version:  %s\n\n", latest)

	assetName := platformAssetName()
	assetURL, checksumURL := findAssetURLs(release.Assets, assetName)
	if assetURL == "" {
		return fmt.Errorf("no release asset found for %s (looking for %q)", runtime.GOOS+"/"+runtime.GOARCH, assetName)
	}

	fmt.Fprintf(w, "Downloading forge %s for %s/%s... ", latest, runtime.GOOS, runtime.GOARCH)
	tmpFile, err := downloadToTemp(assetURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(tmpFile) //nolint:errcheck
	fmt.Fprintln(w, "✓")

	if checksumURL != "" {
		fmt.Fprintf(w, "Verifying checksum (sha256)... ")
		if err := verifyChecksum(tmpFile, checksumURL, assetName); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
		fmt.Fprintln(w, "✓")
	}

	backupPath := binaryPath + ".prev"
	fmt.Fprintf(w, "Backing up current binary to %s\n", backupPath)
	if err := copyFile(binaryPath, backupPath); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	fmt.Fprintf(w, "Installing to %s... ", binaryPath)
	if err := atomicReplace(tmpFile, binaryPath); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}
	fmt.Fprintln(w, "✓")

	fmt.Fprintf(w, "\nforge %s installed successfully.\n", latest)
	return nil
}

// ErrUpdateAvailable is returned by Run when --check finds a newer version.
var ErrUpdateAvailable = fmt.Errorf("update available")

// ---------- internal helpers ----------

func fetchRelease(version string) (*Release, error) {
	url := os.Getenv(envUpdateURL)
	if url == "" {
		if version != "" {
			url = fmt.Sprintf("https://api.github.com/repos/TerrorSquad/forge/releases/tags/%s", version)
		} else {
			url = defaultReleasesURL
		}
	}

	client := &http.Client{Timeout: httpTimeout}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "forge-updater")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

// platformAssetName returns the expected binary name in the release archive,
// e.g. "forge_linux_amd64" or "forge_darwin_arm64".
func platformAssetName() string {
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "amd64"
	}
	return fmt.Sprintf("forge_%s_%s", runtime.GOOS, arch)
}

// findAssetURLs returns the download URL for the binary asset and (if present)
// a URL to the checksums file. It matches loosely so archive suffixes don't
// block the match.
func findAssetURLs(assets []Asset, wantName string) (binaryURL, checksumURL string) {
	for _, a := range assets {
		lower := strings.ToLower(a.Name)
		if strings.Contains(lower, "checksum") || strings.Contains(lower, "sha256") {
			checksumURL = a.BrowserDownloadURL
			continue
		}
		if strings.HasPrefix(lower, strings.ToLower(wantName)) {
			binaryURL = a.BrowserDownloadURL
		}
	}
	return
}

func downloadToTemp(url string) (string, error) {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url) //nolint:noctx
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned %s", resp.Status)
	}

	tmp, err := os.CreateTemp("", "forge-update-*")
	if err != nil {
		return "", err
	}
	defer tmp.Close() //nolint:errcheck

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return tmp.Name(), err
	}
	return tmp.Name(), nil
}

// verifyChecksum fetches the checksum file and compares against localFile.
// The checksum file is expected to be a text file with lines: "<hex>  <filename>".
func verifyChecksum(localFile, checksumURL, assetName string) error {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(checksumURL) //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	expectedHex := ""
	for _, line := range strings.Split(string(body), "\n") {
		parts := strings.Fields(line)
		if len(parts) == 2 && strings.Contains(parts[1], assetName) {
			expectedHex = parts[0]
			break
		}
	}
	if expectedHex == "" {
		// No entry for this asset — skip verification rather than abort.
		return nil
	}

	f, err := os.Open(localFile)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actual := hex.EncodeToString(h.Sum(nil))

	if !strings.EqualFold(actual, expectedHex) {
		return fmt.Errorf("checksum mismatch: expected %s got %s", expectedHex, actual)
	}
	return nil
}

// atomicReplace replaces dst with src using a rename (atomic on POSIX).
// On Windows, the existing binary must be moved first.
func atomicReplace(src, dst string) error {
	info, err := os.Stat(dst)
	if err != nil {
		return err
	}

	if err := os.Chmod(src, info.Mode()); err != nil {
		return err
	}

	// On Windows os.Rename fails if dst is held open; use a two-step move.
	if runtime.GOOS == "windows" {
		old := dst + ".old"
		_ = os.Remove(old)
		if err := os.Rename(dst, old); err != nil {
			return err
		}
	}

	return os.Rename(src, dst)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close() //nolint:errcheck

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close() //nolint:errcheck

	_, err = io.Copy(out, in)
	return err
}

func rollback(w io.Writer) error {
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine binary path: %w", err)
	}
	backupPath := binaryPath + ".prev"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup found at %s", backupPath)
	}
	if err := atomicReplace(backupPath, binaryPath); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}
	fmt.Fprintf(w, "Rolled back to previous version (%s).\n", backupPath)
	return nil
}
