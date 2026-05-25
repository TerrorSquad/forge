package forge

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	toml "github.com/pelletier/go-toml/v2"
)

const (
	remoteFetchTimeout  = 10 * time.Second
	remoteMaxRedirects  = 3
	remoteMaxContentLen = 1 << 20 // 1 MB sanity cap
)

// fetchRemotePreset downloads a TOML config from a https:// URL, validates it,
// and returns the content with a provenance header.
// When yes=true or CI env is set, no confirmation is prompted.
func fetchRemotePreset(rawURL string, yes bool) (string, error) {
	if !strings.HasPrefix(rawURL, "https://") {
		return "", fmt.Errorf("remote presets require an https:// URL, got %q", rawURL)
	}

	fmt.Printf("Fetching: %s\n", rawURL)

	client := &http.Client{
		Timeout: remoteFetchTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= remoteMaxRedirects {
				return fmt.Errorf("too many redirects (max %d)", remoteMaxRedirects)
			}
			return nil
		},
	}

	resp, err := client.Get(rawURL) //nolint:noctx
	if err != nil {
		return "", fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch failed: HTTP %d from %s", resp.StatusCode, rawURL)
	}

	limited := io.LimitReader(resp.Body, remoteMaxContentLen)
	body, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	// Validate as TOML before offering to write.
	var tmp interface{}
	if err := toml.Unmarshal(body, &tmp); err != nil {
		return "", fmt.Errorf("fetched content is not valid TOML: %w", err)
	}

	// Preview + confirmation (skip when yes or CI).
	autoYes := yes || os.Getenv("CI") != ""
	if !autoYes {
		fmt.Println("--- preview ---")
		preview := string(body)
		if len(preview) > 800 {
			preview = preview[:800] + "\n... (truncated)\n"
		}
		fmt.Println(preview)
		fmt.Println("---------------")
		fmt.Print("Write to forge.toml? [y/N] ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer != "y" && answer != "yes" {
			return "", fmt.Errorf("aborted by user")
		}
	}

	header := fmt.Sprintf("# Fetched from %s on %s\n", rawURL, time.Now().Format("2006-01-02"))
	return header + string(body), nil
}
