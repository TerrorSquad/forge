package backend

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TerrorSquad/forge/internal/forge/config"
)

// ---------- helpers ----------

func makeDdevConfig(t *testing.T, projectName string) string {
	t.Helper()
	dir := t.TempDir()
	ddevDir := filepath.Join(dir, ".ddev")
	if err := os.MkdirAll(ddevDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "name: " + projectName + "\ntype: php\n"
	if err := os.WriteFile(filepath.Join(ddevDir, "config.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// ---------- ddevProjectName ----------

func TestDdevProjectName_ReadsName(t *testing.T) {
	dir := makeDdevConfig(t, "my-project")
	name, err := ddevProjectName(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "my-project" {
		t.Errorf("got %q, want %q", name, "my-project")
	}
}

func TestDdevProjectName_QuotedName(t *testing.T) {
	dir := t.TempDir()
	ddevDir := filepath.Join(dir, ".ddev")
	_ = os.MkdirAll(ddevDir, 0o755)
	_ = os.WriteFile(filepath.Join(ddevDir, "config.yaml"), []byte(`name: "quoted-project"`+"\n"), 0o644)

	name, err := ddevProjectName(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "quoted-project" {
		t.Errorf("got %q, want %q", name, "quoted-project")
	}
}

func TestDdevProjectName_MissingFile(t *testing.T) {
	_, err := ddevProjectName(t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing .ddev/config.yaml")
	}
}

func TestDdevProjectName_MissingNameField(t *testing.T) {
	dir := t.TempDir()
	ddevDir := filepath.Join(dir, ".ddev")
	_ = os.MkdirAll(ddevDir, 0o755)
	_ = os.WriteFile(filepath.Join(ddevDir, "config.yaml"), []byte("type: php\n"), 0o644)

	_, err := ddevProjectName(dir)
	if err == nil {
		t.Fatal("expected error when name field is absent")
	}
}

// ---------- ddevContainerName ----------

func TestDdevContainerName_Format(t *testing.T) {
	dir := makeDdevConfig(t, "my-project")
	name, err := ddevContainerName(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "ddev-my-project-web" {
		t.Errorf("got %q, want %q", name, "ddev-my-project-web")
	}
}

func TestDdevContainerName_MissingConfig(t *testing.T) {
	_, err := ddevContainerName(t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

// ---------- HostBackend ----------

func TestHostBackend_Name(t *testing.T) {
	b := &HostBackend{}
	if b.Name() != "host" {
		t.Errorf("got %q, want %q", b.Name(), "host")
	}
}

func TestHostBackend_ExecWithContext_Echo(t *testing.T) {
	b := &HostBackend{}
	var buf bytes.Buffer
	err := b.ExecWithContext(context.Background(), t.TempDir(), []string{"echo", "hello"}, nil, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "hello") {
		t.Errorf("expected output to contain 'hello', got %q", buf.String())
	}
}

func TestHostBackend_ExecWithContext_PassesEnv(t *testing.T) {
	b := &HostBackend{}
	var buf bytes.Buffer
	err := b.ExecWithContext(context.Background(), t.TempDir(), []string{"sh", "-c", "echo $FORGE_TEST_VAR"}, map[string]string{"FORGE_TEST_VAR": "sentinel"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "sentinel") {
		t.Errorf("expected output to contain 'sentinel', got %q", buf.String())
	}
}

func TestHostBackend_ExecWithContext_Cancelled(t *testing.T) {
	b := &HostBackend{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := b.ExecWithContext(ctx, t.TempDir(), []string{"sleep", "10"}, nil, nil)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestHostBackend_BinaryExists_SystemBinary(t *testing.T) {
	b := &HostBackend{}
	// "echo" should be on PATH on any Unix host
	if !b.BinaryExists(t.TempDir(), "echo") {
		t.Error("expected 'echo' to be found via LookPath")
	}
}

func TestHostBackend_BinaryExists_LocalVendorBin(t *testing.T) {
	dir := t.TempDir()
	vendorBin := filepath.Join(dir, "vendor", "bin")
	_ = os.MkdirAll(vendorBin, 0o755)
	_ = os.WriteFile(filepath.Join(vendorBin, "phpstan"), []byte("#!/bin/sh"), 0o755)

	b := &HostBackend{}
	if !b.BinaryExists(dir, "phpstan") {
		t.Error("expected local vendor/bin/phpstan to be found")
	}
}

func TestHostBackend_BinaryExists_LocalNodeBin(t *testing.T) {
	dir := t.TempDir()
	nodeBin := filepath.Join(dir, "node_modules", ".bin")
	_ = os.MkdirAll(nodeBin, 0o755)
	_ = os.WriteFile(filepath.Join(nodeBin, "eslint"), []byte("#!/bin/sh"), 0o755)

	b := &HostBackend{}
	if !b.BinaryExists(dir, "eslint") {
		t.Error("expected local node_modules/.bin/eslint to be found")
	}
}

func TestHostBackend_BinaryExists_Missing(t *testing.T) {
	b := &HostBackend{}
	if b.BinaryExists(t.TempDir(), "this-binary-does-not-exist-forge-test") {
		t.Error("expected missing binary to return false")
	}
}

// ---------- DdevBackend ----------

func TestDdevBackend_Name(t *testing.T) {
	b := &DdevBackend{}
	if b.Name() != "ddev" {
		t.Errorf("got %q, want %q", b.Name(), "ddev")
	}
}

func TestDdevBackend_ExecWithContext_NoDdevConfig(t *testing.T) {
	b := &DdevBackend{}
	err := b.ExecWithContext(context.Background(), t.TempDir(), []string{"echo", "hi"}, nil, nil)
	if err == nil {
		t.Fatal("expected error when no .ddev/config.yaml present")
	}
	if !strings.Contains(err.Error(), "ddev backend") {
		t.Errorf("expected error to mention 'ddev backend', got: %v", err)
	}
}

func TestDdevBackend_BinaryExists_LocalVendorBin(t *testing.T) {
	dir := t.TempDir()
	vendorBin := filepath.Join(dir, "vendor", "bin")
	_ = os.MkdirAll(vendorBin, 0o755)
	_ = os.WriteFile(filepath.Join(vendorBin, "phpstan"), []byte("#!/bin/sh"), 0o755)

	b := &DdevBackend{}
	if !b.BinaryExists(dir, "phpstan") {
		t.Error("expected local vendor/bin/phpstan to be found")
	}
}

func TestDdevBackend_BinaryExists_KnownSystemBinary(t *testing.T) {
	b := &DdevBackend{}
	for _, bin := range []string{"php", "composer", "node", "npm"} {
		if !b.BinaryExists(t.TempDir(), bin) {
			t.Errorf("expected known system binary %q to return true", bin)
		}
	}
}

func TestDdevBackend_BinaryExists_UnknownBinary(t *testing.T) {
	b := &DdevBackend{}
	// unknown binary not in vendor/bin and not in the known list
	if b.BinaryExists(t.TempDir(), "this-binary-does-not-exist-forge-test") {
		t.Error("expected unknown binary to return false")
	}
}

// ---------- ResolveBackend ----------

func TestResolveBackend_ExplicitHost(t *testing.T) {
	tool := config.ToolConfig{Backend: "host"}
	b := ResolveBackend(t.TempDir(), tool, "")
	if _, ok := b.(*HostBackend); !ok {
		t.Errorf("expected HostBackend, got %T", b)
	}
}

func TestResolveBackend_ExplicitDdev(t *testing.T) {
	tool := config.ToolConfig{Backend: "ddev"}
	b := ResolveBackend(t.TempDir(), tool, "")
	if _, ok := b.(*DdevBackend); !ok {
		t.Errorf("expected DdevBackend, got %T", b)
	}
}

func TestResolveBackend_GlobalDefault_Host(t *testing.T) {
	tool := config.ToolConfig{}
	b := ResolveBackend(t.TempDir(), tool, "host")
	if _, ok := b.(*HostBackend); !ok {
		t.Errorf("expected HostBackend from global default, got %T", b)
	}
}

func TestResolveBackend_GlobalDefault_Ddev(t *testing.T) {
	tool := config.ToolConfig{}
	b := ResolveBackend(t.TempDir(), tool, "ddev")
	if _, ok := b.(*DdevBackend); !ok {
		t.Errorf("expected DdevBackend from global default, got %T", b)
	}
}

func TestResolveBackend_AutoDetect_NoDdev_FallsBackToHost(t *testing.T) {
	// No .ddev/config.yaml → isDdevRunning → false → HostBackend
	tool := config.ToolConfig{}
	b := ResolveBackend(t.TempDir(), tool, "")
	if _, ok := b.(*HostBackend); !ok {
		t.Errorf("expected HostBackend when DDEV is absent, got %T", b)
	}
}

func TestResolveBackend_ToolOverridesGlobal(t *testing.T) {
	tool := config.ToolConfig{Backend: "host"}
	b := ResolveBackend(t.TempDir(), tool, "ddev")
	if _, ok := b.(*HostBackend); !ok {
		t.Errorf("expected HostBackend (tool override), got %T", b)
	}
}

// ---------- ResolveCommandForBackend ----------

func TestResolveCommandForBackend_PHP_LocalVendor(t *testing.T) {
	dir := t.TempDir()
	vendorBin := filepath.Join(dir, "vendor", "bin")
	_ = os.MkdirAll(vendorBin, 0o755)
	_ = os.WriteFile(filepath.Join(vendorBin, "phpstan"), []byte("#!/bin/sh"), 0o755)

	tool := config.ToolConfig{Command: "phpstan", Type: "php"}
	got := ResolveCommandForBackend(dir, tool, &HostBackend{})
	if got != filepath.Join(dir, "vendor", "bin", "phpstan") {
		t.Errorf("got %q, want vendor/bin/phpstan path", got)
	}
}

func TestResolveCommandForBackend_Node_LocalNodeModules(t *testing.T) {
	dir := t.TempDir()
	nodeBin := filepath.Join(dir, "node_modules", ".bin")
	_ = os.MkdirAll(nodeBin, 0o755)
	_ = os.WriteFile(filepath.Join(nodeBin, "eslint"), []byte("#!/bin/sh"), 0o755)

	tool := config.ToolConfig{Command: "eslint", Type: "node"}
	got := ResolveCommandForBackend(dir, tool, &HostBackend{})
	if got != filepath.Join(dir, "node_modules", ".bin", "eslint") {
		t.Errorf("got %q, want node_modules/.bin/eslint path", got)
	}
}

func TestResolveCommandForBackend_PHP_NotLocalFallsBackToCommand(t *testing.T) {
	tool := config.ToolConfig{Command: "phpstan", Type: "php"}
	got := ResolveCommandForBackend(t.TempDir(), tool, &HostBackend{})
	if got != "phpstan" {
		t.Errorf("got %q, want %q", got, "phpstan")
	}
}

func TestResolveCommandForBackend_SystemType_ReturnsCommandAsIs(t *testing.T) {
	tool := config.ToolConfig{Command: "make", Type: "system"}
	got := ResolveCommandForBackend(t.TempDir(), tool, &HostBackend{})
	if got != "make" {
		t.Errorf("got %q, want %q", got, "make")
	}
}

// ---------- ToolBinaryAvailable ----------

func TestToolBinaryAvailable_DdevBackend_AlwaysTrue(t *testing.T) {
	b := &DdevBackend{}
	if !ToolBinaryAvailable(t.TempDir(), "anything", b) {
		t.Error("DdevBackend should always return true for ToolBinaryAvailable")
	}
}

func TestToolBinaryAvailable_HostBackend_AbsolutePath_Exists(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "mytool")
	_ = os.WriteFile(bin, []byte("#!/bin/sh"), 0o755)

	if !ToolBinaryAvailable(dir, bin, &HostBackend{}) {
		t.Error("expected true for existing absolute path")
	}
}

func TestToolBinaryAvailable_HostBackend_AbsolutePath_Missing(t *testing.T) {
	if ToolBinaryAvailable(t.TempDir(), "/nonexistent/path/forge-test-tool", &HostBackend{}) {
		t.Error("expected false for non-existent absolute path")
	}
}

func TestToolBinaryAvailable_HostBackend_SystemBinary(t *testing.T) {
	if !ToolBinaryAvailable(t.TempDir(), "echo", &HostBackend{}) {
		t.Error("expected true for 'echo' on system PATH")
	}
}

// ---------- BackendAvailabilityError ----------

func TestBackendAvailabilityError_Error(t *testing.T) {
	e := &BackendAvailabilityError{Tool: "phpstan", Backend: "host"}
	msg := e.Error()
	if !strings.Contains(msg, "phpstan") || !strings.Contains(msg, "host") {
		t.Errorf("unexpected error message: %q", msg)
	}
}
