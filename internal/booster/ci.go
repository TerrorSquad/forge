package booster

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// ciMode returns the detected CI output format: "github", "gitlab", or "".
func ciMode() string {
	if override := strings.ToLower(strings.TrimSpace(os.Getenv("BOOSTER_OUTPUT"))); override != "" {
		return override
	}
	if isTruthy(os.Getenv("GITHUB_ACTIONS")) {
		return "github"
	}
	if isTruthy(os.Getenv("GITLAB_CI")) {
		return "gitlab"
	}
	return ""
}

// isCI reports whether we are running in any CI environment.
func isCI() bool {
	return isTruthy(os.Getenv("CI")) || ciMode() != ""
}

// --- GitHub Actions output helpers ---

// ghaOpenGroup emits a collapsible ::group:: block.
func ghaOpenGroup(title string) {
	fmt.Fprintf(UI, "::group::%s\n", title)
}

// ghaCloseGroup closes an open ::group:: block.
func ghaCloseGroup() {
	fmt.Fprintf(UI, "::endgroup::\n")
}

// ghaEmitAnnotations parses tool output for file:line error patterns and emits
// GitHub Actions ::error:: / ::warning:: workflow commands.
// Lines that don't match known patterns are emitted as plain ::error:: messages.
func ghaEmitAnnotations(output, toolName string) {
	if strings.TrimSpace(output) == "" {
		return
	}
	for _, line := range strings.Split(strings.TrimRight(output, "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if ann := parseErrorAnnotation(line); ann != "" {
			fmt.Fprintf(UI, "%s\n", ann)
		} else {
			fmt.Fprintf(UI, "::error title=%s::%s\n", toolName, escapeGHAValue(line))
		}
	}
}

// errorPattern represents a regex to extract file/line/col/message from a tool's output.
type errorPattern struct {
	re    *regexp.Regexp
	file  string // named group
	line  string
	col   string
	msg   string
	level string // "error" or "warning"
}

// knownErrorPatterns maps common tool output formats to GitHub Actions annotations.
var knownErrorPatterns = []errorPattern{
	// PHPStan: " src/Foo.php:42 - Parameter #1 expects string"
	{
		re:   regexp.MustCompile(`^\s*(?P<file>[^:]+\.php):(?P<line>\d+)\s+-\s+(?P<msg>.+)$`),
		file: "file", line: "line", msg: "msg", level: "error",
	},
	// Psalm: "ERROR: InvalidArgument - src/Foo.php:42:10 - ..."
	{
		re:   regexp.MustCompile(`^(?P<level>ERROR|INFO|WARNING|SUGGESTION): \w+ - (?P<file>[^:]+):(?P<line>\d+):\d+ - (?P<msg>.+)$`),
		file: "file", line: "line", msg: "msg", level: "level",
	},
	// ESLint: "  42:10  error  message  rule-name"
	{
		re:   regexp.MustCompile(`^\s*(?P<line>\d+):(?P<col>\d+)\s+(?P<level>error|warning)\s+(?P<msg>.+?)\s+\S+$`),
		file: "", line: "line", col: "col", msg: "msg", level: "level",
	},
	// golangci-lint / go vet: "src/foo.go:42:10: message"
	{
		re:   regexp.MustCompile(`^(?P<file>[^:]+\.go):(?P<line>\d+):(?P<col>\d+):\s*(?P<msg>.+)$`),
		file: "file", line: "line", col: "col", msg: "msg", level: "error",
	},
	// Generic "file.ext:line: message"
	{
		re:   regexp.MustCompile(`^(?P<file>[^:]+\.\w+):(?P<line>\d+):\s*(?P<msg>.+)$`),
		file: "file", line: "line", msg: "msg", level: "error",
	},
}

// parseErrorAnnotation tries to parse a line into a GHA workflow command.
// Returns empty string if no pattern matches.
func parseErrorAnnotation(line string) string {
	for _, p := range knownErrorPatterns {
		m := p.re.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		idx := p.re.SubexpIndex

		get := func(name string) string {
			if name == "" {
				return ""
			}
			i := idx(name)
			if i < 0 || i >= len(m) {
				return ""
			}
			return m[i]
		}

		level := strings.ToLower(get(p.level))
		if level == "" {
			level = p.level // static value
		}
		if level != "warning" {
			level = "error"
		}

		file := escapeGHAParam(get(p.file))
		lineN := get(p.line)
		col := get(p.col)
		msg := escapeGHAValue(get(p.msg))

		params := ""
		if file != "" {
			params += "file=" + file
		}
		if lineN != "" {
			if params != "" {
				params += ","
			}
			params += "line=" + lineN
		}
		if col != "" {
			params += ",col=" + col
		}
		if params != "" {
			return fmt.Sprintf("::%s %s::%s", level, params, msg)
		}
		return fmt.Sprintf("::%s::%s", level, msg)
	}
	return ""
}

func escapeGHAValue(s string) string {
	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, "\r", "%0D")
	s = strings.ReplaceAll(s, "\n", "%0A")
	return s
}

func escapeGHAParam(s string) string {
	s = escapeGHAValue(s)
	s = strings.ReplaceAll(s, ":", "%3A")
	s = strings.ReplaceAll(s, ",", "%2C")
	return s
}

// --- CI-aware print functions ---

// PrintHookHeaderCI emits CI-appropriate hook header.
func PrintHookHeaderCI(hookName string) {
	mode := ciMode()
	switch mode {
	case "github":
		ghaOpenGroup("booster · " + hookName)
	default:
		PrintHookHeader(hookName)
	}
}

// PrintToolResultCI emits CI-appropriate tool result.
func PrintToolResultCI(r ToolResult, toolName string) {
	mode := ciMode()
	if mode != "github" {
		PrintToolResult(r)
		return
	}
	// In GHA mode: always print the plain result line, then emit annotations for failures.
	PrintToolResult(r)
	if r.Status == "fail" || r.Status == "would-fail" {
		ghaEmitAnnotations(r.Output, toolName)
	}
}

// PrintSummaryCI emits CI-appropriate summary and closes any open groups.
func PrintSummaryCI(results []ToolResult, total time.Duration, checkMode bool) {
	if checkMode {
		PrintCheckSummary(results, total)
	} else {
		PrintSummary(results, total)
	}
	if ciMode() == "github" {
		ghaCloseGroup()
	}
}
