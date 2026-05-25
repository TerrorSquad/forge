package forge

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// cacheEntry represents a single tool cache record.
type cacheEntry struct {
	Passed    bool      `json:"passed"`
	Timestamp time.Time `json:"timestamp"`
}

// toolCache is the on-disk structure.
type toolCache map[string]cacheEntry // key → entry

const cacheFile = ".forge/cache.json"

// loadCache reads the cache from disk; returns an empty cache on any error (corrupted = silently reset).
func loadCache(repoRoot string) toolCache {
	p := filepath.Join(repoRoot, cacheFile)
	data, err := os.ReadFile(p)
	if err != nil {
		return toolCache{}
	}
	var tc toolCache
	if err := json.Unmarshal(data, &tc); err != nil {
		// Corrupted — silently delete and start fresh.
		_ = os.Remove(p)
		return toolCache{}
	}
	return tc
}

// saveCache writes the cache atomically via a temp file + rename.
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

// toolCacheKey computes the content-addressed key for a tool run.
// key = sha256(serialised tool config + sorted "path:filehash" list)
func toolCacheKey(tool ToolConfig, files []string) (string, error) {
	cfgBytes, err := json.Marshal(tool)
	if err != nil {
		return "", err
	}

	// Compute per-file hashes and sort.
	type fileHash struct {
		path string
		hash string
	}
	fhList := make([]fileHash, 0, len(files))
	for _, f := range files {
		h, err := hashFile(f)
		if err != nil {
			// If a file doesn't exist (deleted), use the path as placeholder.
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

// hashFile returns the hex sha256 of the file at the given path.
func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

// isCacheHit returns true if the tool's inputs are unchanged since last pass.
func isCacheHit(tc toolCache, key string) bool {
	entry, ok := tc[key]
	return ok && entry.Passed
}

// updateCacheEntry marks a key as passing.
func updateCacheEntry(tc toolCache, key string) {
	tc[key] = cacheEntry{Passed: true, Timestamp: time.Now()}
}
