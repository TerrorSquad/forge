package backend

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/TerrorSquad/forge/internal/forge/config"
)

// Backend represents an execution environment for tool commands.
type Backend interface {
	// Name returns the backend identifier.
	Name() string
	// Exec runs cmd, streaming stdout/stderr to the process stdout/stderr.
	Exec(dir string, cmd []string) error
	// ExecWithWriter runs cmd, writing combined stdout+stderr to w.
	ExecWithWriter(dir string, cmd []string, w io.Writer) error
	// ExecWithContext runs cmd respecting ctx cancellation, writing to w when non-nil.
	// env contains additional environment variables merged on top of the parent process env.
	ExecWithContext(ctx context.Context, dir string, cmd []string, env map[string]string, w io.Writer) error
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

// DdevBackend routes commands through `docker exec` into the DDEV web container.
// This is faster than `ddev exec` because it skips the DDEV CLI overhead.
type DdevBackend struct{}

type DockerBackend struct {
	container string
}

func (b *DdevBackend) Name() string { return "ddev" }

func (b *DdevBackend) Exec(dir string, cmd []string) error {
	return b.ExecWithWriter(dir, cmd, nil)
}

func (b *DdevBackend) ExecWithWriter(dir string, cmd []string, w io.Writer) error {
	return b.ExecWithContext(context.Background(), dir, cmd, nil, w)
}

func (b *DdevBackend) ExecWithContext(ctx context.Context, dir string, cmd []string, env map[string]string, w io.Writer) error {
	container, err := ddevContainerName(dir)
	if err != nil {
		return fmt.Errorf("ddev backend: %w", err)
	}
	containerDir := ddevContainerDir(dir)
	dockerArgs := []string{"docker", "exec", "-i", "-w", containerDir}
	for k, v := range env {
		dockerArgs = append(dockerArgs, "-e", k+"="+v)
	}
	dockerArgs = append(dockerArgs, container)
	dockerArgs = append(dockerArgs, cmd...)
	c := exec.CommandContext(ctx, dockerArgs[0], dockerArgs[1:]...)
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

func (b *DockerBackend) Name() string { return b.container }

func (b *DockerBackend) Exec(dir string, cmd []string) error {
	return b.ExecWithWriter(dir, cmd, nil)
}

func (b *DockerBackend) ExecWithWriter(dir string, cmd []string, w io.Writer) error {
	return b.ExecWithContext(context.Background(), dir, cmd, nil, w)
}

func (b *DockerBackend) ExecWithContext(ctx context.Context, dir string, cmd []string, env map[string]string, w io.Writer) error {
	dockerArgs := []string{"docker", "exec", "-i"}
	for k, v := range env {
		dockerArgs = append(dockerArgs, "-e", k+"="+v)
	}
	dockerArgs = append(dockerArgs, b.container)
	dockerArgs = append(dockerArgs, cmd...)
	c := exec.CommandContext(ctx, dockerArgs[0], dockerArgs[1:]...)
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

// ResolveBackend returns the appropriate backend for a tool in the given repo root.
// Priority: per-tool override → global config default → DDEV auto-detect → host
func ResolveBackend(repoRoot string, tool config.ToolConfig, globalDefault string) Backend {
	name := tool.Backend
	if name == "" {
		name = globalDefault
	}

	switch name {
	case "ddev":
		return &DdevBackend{}
	case "host":
		return &HostBackend{}
	case "":
		if isDdevRunning(repoRoot) {
			return &DdevBackend{}
		}
		return &HostBackend{}
	default:
		return &DockerBackend{container: name}
	}
}

// isDdevRunning returns true if the DDEV web container for repoRoot is running.
// Uses `docker inspect` directly — faster than `ddev status`.
func isDdevRunning(repoRoot string) bool {
	container, err := ddevContainerName(repoRoot)
	if err != nil {
		return false
	}
	return isDockerContainerRunning(container)
}

func isDockerContainerRunning(container string) bool {
	out, err := exec.Command("docker", "inspect", "--format", "{{.State.Running}}", container).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}

// ddevProjectName reads the DDEV project name from .ddev/config.yaml.
func ddevProjectName(repoRoot string) (string, error) {
	data, err := os.ReadFile(filepath.Join(repoRoot, ".ddev", "config.yaml"))
	if err != nil {
		return "", fmt.Errorf("no .ddev/config.yaml found: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name:") {
			name := strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			name = strings.Trim(name, `"'`)
			if name != "" {
				return name, nil
			}
		}
	}
	return "", fmt.Errorf(".ddev/config.yaml has no 'name:' field")
}

// ddevContainerName returns the Docker container name for the DDEV web service.
func ddevContainerName(repoRoot string) (string, error) {
	name, err := ddevProjectName(repoRoot)
	if err != nil {
		return "", err
	}
	return "ddev-" + name + "-web", nil
}

// ddevProjectRoot walks up from startDir until it finds a directory containing
// a .ddev/ subdirectory, returning the DDEV project root.
func ddevProjectRoot(startDir string) (string, error) {
	dir := filepath.Clean(startDir)
	for {
		if _, err := os.Stat(filepath.Join(dir, ".ddev")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .ddev directory found above %s", startDir)
		}
		dir = parent
	}
}

// ddevContainerDir translates a host absolute path to the equivalent path
// inside the DDEV web container, which mounts the project root at /var/www/html.
// If the project root cannot be determined, /var/www/html is returned.
func ddevContainerDir(hostDir string) string {
	root, err := ddevProjectRoot(hostDir)
	if err != nil {
		return "/var/www/html"
	}
	rel, err := filepath.Rel(root, hostDir)
	if err != nil || rel == "." {
		return "/var/www/html"
	}
	return "/var/www/html/" + filepath.ToSlash(rel)
}

// ResolveCommandForBackend resolves a vendor/node_modules binary path relative
// to repo root depending on the tool type and active backend.
func ResolveCommandForBackend(repoRoot string, tool config.ToolConfig, backend Backend) string {
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

// ToolBinaryAvailable reports whether the resolved command path is accessible.
func ToolBinaryAvailable(repoRoot, resolvedCmd string, backend Backend) bool {
	if _, isDdev := backend.(*DdevBackend); isDdev {
		return true
	}
	if _, isDocker := backend.(*DockerBackend); isDocker {
		return true
	}
	if filepath.IsAbs(resolvedCmd) {
		_, err := os.Stat(resolvedCmd)
		return err == nil
	}
	if strings.Contains(resolvedCmd, "/") {
		_, err := os.Stat(filepath.Join(repoRoot, resolvedCmd))
		return err == nil
	}
	_, err := exec.LookPath(resolvedCmd)
	return err == nil
}

// BackendIssue describes a backend availability problem found during doctor.
type BackendIssue struct {
	Backend string
	Message string
}

// CheckBackendAvailability inspects every backend referenced in cfg and returns
// a list of issues (empty slice means all backends are reachable).
func CheckBackendAvailability(repoRoot string, cfg *config.Config) []BackendIssue {
	// Collect all unique backend names explicitly referenced.
	names := map[string]struct{}{}
	if cfg.Execution.DefaultBackend != "" {
		names[cfg.Execution.DefaultBackend] = struct{}{}
	}
	for _, hook := range cfg.Hooks {
		for _, tool := range hook.Tools {
			if tool.Backend != "" {
				names[tool.Backend] = struct{}{}
			}
		}
	}

	var issues []BackendIssue
	for name := range names {
		switch name {
		case "ddev":
			if !isDdevRunning(repoRoot) {
				issues = append(issues, BackendIssue{
					Backend: "ddev",
					Message: "DDEV web container is not running — tools with backend=ddev will fail",
				})
			}
		case "host", "":
			// always available
		default:
			if !isDockerContainerRunning(name) {
				issues = append(issues, BackendIssue{
					Backend: name,
					Message: fmt.Sprintf("docker container %q is not running or inaccessible", name),
				})
			}
		}
	}
	return issues
}
