package forge

import (
	"path/filepath"
	"testing"
)

func TestSaveAndLoadCache_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	tc := toolCache{"abc123": {Passed: true}}
	saveCache(dir, tc)

	loaded := loadCache(dir)
	if !loaded["abc123"].Passed {
		t.Error("expected cache hit after save+load")
	}
}

func TestLoadCache_Missing(t *testing.T) {
	dir := t.TempDir()
	tc := loadCache(dir)
	if len(tc) != 0 {
		t.Errorf("expected empty cache for missing file, got %d entries", len(tc))
	}
}

func TestLoadCache_Corrupted(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, cacheFile), "not valid json{{{{")
	tc := loadCache(dir)
	if len(tc) != 0 {
		t.Errorf("expected empty cache for corrupted file, got %d entries", len(tc))
	}
}

func TestIsCacheHit_Miss(t *testing.T) {
	tc := toolCache{}
	if isCacheHit(tc, "unknown-key") {
		t.Error("expected cache miss for unknown key")
	}
}

func TestIsCacheHit_Hit(t *testing.T) {
	tc := toolCache{}
	updateCacheEntry(tc, "mykey")
	if !isCacheHit(tc, "mykey") {
		t.Error("expected cache hit after update")
	}
}

func TestToolCacheKey_Deterministic(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "a.go")
	f2 := filepath.Join(dir, "b.go")
	writeFile(t, f1, "package main\n")
	writeFile(t, f2, "package main\n")

	tool := ToolConfig{Command: "gofmt", Args: []string{"-w"}}
	k1, err1 := toolCacheKey(tool, []string{f1, f2})
	k2, err2 := toolCacheKey(tool, []string{f2, f1}) // reversed order
	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v, %v", err1, err2)
	}
	if k1 != k2 {
		t.Errorf("cache key should be order-independent, got %s vs %s", k1, k2)
	}
}

func TestToolCacheKey_ChangesOnFileChange(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	writeFile(t, f, "package main\n")
	tool := ToolConfig{Command: "gofmt"}

	k1, _ := toolCacheKey(tool, []string{f})
	writeFile(t, f, "package main // changed\n")
	k2, _ := toolCacheKey(tool, []string{f})

	if k1 == k2 {
		t.Error("cache key should change when file content changes")
	}
}

func TestClearCache(t *testing.T) {
	dir := t.TempDir()
	tc := toolCache{"x": {Passed: true}}
	saveCache(dir, tc)

	if err := ClearCache(dir); err != nil {
		t.Fatalf("ClearCache failed: %v", err)
	}

	loaded := loadCache(dir)
	if len(loaded) != 0 {
		t.Error("expected empty cache after clear")
	}
}

func TestClearCache_NoFile(t *testing.T) {
	dir := t.TempDir()
	// Clearing a non-existent cache should not error.
	if err := ClearCache(dir); err != nil {
		t.Errorf("ClearCache should not error when file is absent: %v", err)
	}
}
