package runner

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/TerrorSquad/forge/internal/forge/config"
)

type cacheEntry struct {
	Passed    bool      `json:"passed"`
	Timestamp time.Time `json:"timestamp"`
}

type toolCache map[string]cacheEntry

const cacheFile = ".forge/cache.json"

func loadCache(repoRoot string) toolCache {
	p := filepath.Join(repoRoot, cacheFile)
	data, err := os.ReadFile(p)
	if err != nil {
		return toolCache{}
	}
	var tc toolCache
	if err := json.Unmarshal(data, &tc); err != nil {
		_ = os.Remove(p)
		return toolCache{}
	}
	return tc
}

func saveCache(repoRoot string, tc toolCache) {
	p := filepath.Join(repoRoot, cacheFile)
	_ = os.MkdirAll(filepath.Dir(p), 0755)

	data, err := json.MarshalIndent(tc, "", "  ")
	if err != nil {
		return
	}

	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return
	}
	_ = os.Rename(tmp, p)
}

// ClearCache deletes the cache file.
func ClearCache(repoRoot string) error {
	p := filepath.Join(repoRoot, cacheFile)
	err := os.Remove(p)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func toolCacheKey(tool config.ToolConfig, files []string) (string, error) {
	cfgBytes, err := json.Marshal(tool)
	if err != nil {
		return "", err
	}

	type fileHash struct {
		path string
		hash string
	}
	fhList := make([]fileHash, 0, len(files))
	for _, f := range files {
		h, err := hashFile(f)
		if err != nil {
			h = "deleted"
		}
		fhList = append(fhList, fileHash{f, h})
	}
	sort.Slice(fhList, func(i, j int) bool { return fhList[i].path < fhList[j].path })

	h := sha256.New()
	h.Write(cfgBytes)
	for _, fh := range fhList {
		fmt.Fprintf(h, "%s:%s\n", fh.path, fh.hash)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

func isCacheHit(tc toolCache, key string) bool {
	entry, ok := tc[key]
	return ok && entry.Passed
}

func updateCacheEntry(tc toolCache, key string) {
	tc[key] = cacheEntry{Passed: true, Timestamp: time.Now()}
}
