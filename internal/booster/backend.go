package booster

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Backend represents an execution environment for tool commands.
type Backend interface {
	// Name returns the backend identifier.
	Name() string
	// Exec runs cmd with args, streaming stdout/stderr to the process stdout/stderr.
	Exec(dir string, cmd []string) error
	// ExecWithWriter runs cmd with args, writing combined stdout+stderr to w.
	ExecWithWriter(dir string, cmd []string, w io.Writer) error
	// ExecWithContext runs cmd with args (respecting ctx cancellation), writing to w when non-nil.
	// env contains additional environment variables merged on top of the parent process env; nil means inherit unchanged.
	ExecWithContext(ctx context.Context, dir string, cmd []string, env map[string]string, w io.Writer) error
	// BinaryExists checks whether the named binary is available in this backend.
	BinaryExists(dir, binary string) bool
}

// HostBackend runs commands directly on the host.
type HostBackend struct{}

func (b *HostBackend) Name() string { return "host" }

func (b *HostBackend) Exec(dir string, cmd []string) error {
	return b.ExecWithWriter(dir, cmd, nil)
}

func (b *HostBackend) ExecWithWriter(dir string, cmd []string, w io.Writer) error {
	return b.ExecWithContext(context.Background(), dir, cmd, nil, w)
}

func (b *HostBackend) ExecWithContext(ctx context.Context, dir string, cmd []string, env map[string]string, w io.Writer) error {
	c := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	c.Dir = dir
	if len(env) > 0 {
		c.Env = os.Environ()
		for k, v := range env {
			c.Env = append(c.Env, k+"="+v)
		}
	}
	if w != nil {
		c.Stdout = w
		c.Stderr = w
	} else {
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
	}
	c.Stdin = os.Stdin
	return c.Run()
}

func (b *HostBackend) BinaryExists(dir, binary string) bool {
	// Check project-local paths first (vendor/bin, node_modules/.bin)
	local := []string{
		filepath.Join(dir, "vendor", "bin", binary),
		filepath.Join(dir, "node_modules", ".bin", binary),
	}
	for _, p := range local {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	_, err := exec.LookPath(binary)
	return err == nil
}

// DdevBackend routes commands through `ddev exec`.
type DdevBackend struct{}

func (b *DdevBackend) Name() string { return "ddev" }

func (b *DdevBackend) Exec(dir string, cmd []string) error {
	return b.ExecWithWriter(dir, cmd, nil)
}

func (b *DdevBackend) ExecWithWriter(dir string, cmd []string, w io.Writer) error {
	return b.ExecWithContext(context.Background(), dir, cmd, nil, w)
}

func (b *DdevBackend) ExecWithContext(ctx context.Context, dir string, cmd []string, env map[string]string, w io.Writer) error {
	ddevArgs := []string{"ddev", "exec"}
	for k, v := range env {
		ddevArgs = append(ddevArgs, "--env", k+"="+v)
	}
	ddevArgs = append(ddevArgs, "--")
	ddevCmd := append(ddevArgs, cmd...)
	c := exec.CommandContext(ctx, ddevCmd[0], ddevCmd[1:]...)
	c.Dir = dir
	if w != nil {
		c.Stdout = w
		c.Stderr = w
	} else {
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
	}
	c.Stdin = os.Stdin
	return c.Run()
}

func (b *DdevBackend) BinaryExists(dir, binary string) bool {
	// For DDEV, check vendor/bin and node_modules/.bin inside the project
	local := []string{
		filepath.Join(dir, "vendor", "bin", binary),
		filepath.Join(dir, "node_modules", ".bin", binary),
	}
	for _, p := range local {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	// Also accept system-level tools the container likely has (php, composer, node)
	system := map[string]bool{"php": true, "composer": true, "node": true, "npm": true}
	return system[binary]
}

// ResolveBackend returns the appropriate backend for a tool in the given repo root.
// Priority: per-tool override → global config default → DDEV auto-detect → host
func ResolveBackend(repoRoot string, tool ToolConfig, globalDefault string) Backend {
	name := tool.Backend
	if name == "" {
		name = globalDefault
	}

	// Explicit backend requested
	switch name {
	case "ddev":
		return &DdevBackend{}
	case "host":
		return &HostBackend{}
	}

	// Auto-detect DDEV
	if isDdevRunning(repoRoot) {
		return &DdevBackend{}
	}

	return &HostBackend{}
}

// isDdevRunning returns true if a DDEV project is active in repoRoot.
func isDdevRunning(repoRoot string) bool {
	cfg := filepath.Join(repoRoot, ".ddev", "config.yaml")
	if _, err := os.Stat(cfg); err != nil {
		return false
	}
	var out bytes.Buffer
	c := exec.Command("ddev", "status", "--json-output")
	c.Dir = repoRoot
	c.Stdout = &out
	if err := c.Run(); err != nil {
		return false
	}
	return bytes.Contains(out.Bytes(), []byte(`"running"`))
}

// resolveCommandForBackend resolves a vendor/node_modules binary path relative
// to repo root depending on the tool type and active backend.
func resolveCommandForBackend(repoRoot string, tool ToolConfig, backend Backend) string {
	cmd := tool.Command
	switch tool.Type {
	case "php":
		local := filepath.Join(repoRoot, "vendor", "bin", cmd)
		if _, err := os.Stat(local); err == nil {
			return local
		}
	case "node":
		local := filepath.Join(repoRoot, "node_modules", ".bin", cmd)
		if _, err := os.Stat(local); err == nil {
			return local
		}
	}
	return cmd
}

// BackendAvailabilityError is returned when no suitable backend can execute a tool.
type BackendAvailabilityError struct {
	Tool    string
	Backend string
}

func (e *BackendAvailabilityError) Error() string {
	return fmt.Sprintf("tool %q not available via backend %q", e.Tool, e.Backend)
}

// toolBinaryAvailable reports whether the resolved command path is accessible.
// For DdevBackend, the check is always skipped — the container is assumed to
// have every tool it is configured to run.
// resolvedCmd may be:
//   - a relative vendor/bin or node_modules/.bin path (checked relative to repoRoot)
//   - a plain binary name (checked on system PATH)
//   - an absolute path
func toolBinaryAvailable(repoRoot, resolvedCmd string, backend Backend) bool {
	// DDEV container is authoritative — don't try to resolve host paths.
	if _, isDdev := backend.(*DdevBackend); isDdev {
		return true
	}
	if filepath.IsAbs(resolvedCmd) {
		_, err := os.Stat(resolvedCmd)
		return err == nil
	}
	// Relative local binary (vendor/bin/*, node_modules/.bin/*)
	if strings.Contains(resolvedCmd, "/") {
		_, err := os.Stat(filepath.Join(repoRoot, resolvedCmd))
		return err == nil
	}
	// System/PATH binary
	_, err := exec.LookPath(resolvedCmd)
	return err == nil
}
