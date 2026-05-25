# Feature 033: CI Output Mode

## Summary
Detect CI environments and emit machine-readable output: GitHub Actions
workflow commands (problem matchers / annotations), GitLab CI section markers,
and JUnit XML. Gives inline code annotations on pull requests automatically.

## Motivation
In CI, forge today produces the same coloured terminal output as locally.
GitHub Actions and GitLab can parse structured output to create inline PR
annotations ("PHPStan: src/Foo.php:42 — error: ..."). JUnit XML is consumed
by test reporting plugins (Allure, Jenkins, GitLab test summary). Structured
output makes CI results dramatically more actionable.

## CI Environment Detection

Forge detects CI mode from standard env vars:
- `CI=true` → generic CI (plain text, no colour)
- `GITHUB_ACTIONS=true` → GitHub Actions annotations
- `GITLAB_CI=true` → GitLab CI section markers
- `FORGE_OUTPUT=<format>` → explicit override (`github`, `gitlab`, `junit`,
  `text`, `json`)

## GitHub Actions Output

When `GITHUB_ACTIONS=true`, failed tool output is re-emitted using [workflow
commands](https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions):

```
::group::forge pre-commit
::error file=src/Foo/Handler/FooHandler.php,line=42,title=phpstan::Parameter #1 expects string, int given.
::endgroup::
```

Rules:
- Parse tool output for common error patterns (PHPStan, Psalm, ESLint, PHP CS
  Fixer, golangci-lint) to extract file/line/message.
- Lines that don't match a known pattern are emitted as plain `::error::` with
  the raw message.
- `::group::` / `::endgroup::` wrap each tool's output for collapsible display.
- Passing tools emit `::notice` with timing.

## GitLab CI Section Markers

```
\e[0Ksection_start:TIMESTAMP:phpstan[collapsed=true]\r\e[0K📋 phpstan
... output ...
\e[0Ksection_end:TIMESTAMP:phpstan\r\e[0K
```

## JUnit XML

`forge run pre-commit --output=junit > forge-results.xml`

```xml
<?xml version="1.0" encoding="UTF-8"?>
<testsuites name="forge" time="9.3" tests="4" failures="1">
  <testsuite name="pre-commit" tests="4" failures="1" time="9.3">
    <testcase name="ecs"    time="0.34" classname="pre-commit"/>
    <testcase name="phpstan" time="4.2"  classname="pre-commit">
      <failure message="src/Foo.php:42 — Parameter #1 expects string, int given."/>
    </testcase>
    <testcase name="psalm"  time="1.8"  classname="pre-commit"/>
    <testcase name="deptrac" time="2.9" classname="pre-commit"/>
  </testsuite>
</testsuites>
```

## JSON Output

`forge run pre-commit --output=json`

```json
{
  "hook": "pre-commit",
  "passed": false,
  "duration_ms": 9300,
  "tools": [
    { "name": "ecs",     "status": "pass", "duration_ms": 340,  "output": "" },
    { "name": "phpstan", "status": "fail", "duration_ms": 4200, "output": "..." }
  ]
}
```

## Functional Requirements

1. Auto-detect CI environment on startup; no config change required.
2. `--output` flag overrides auto-detection for any `forge run` invocation.
3. Colours are disabled in all non-TTY / CI modes.
4. JUnit and JSON outputs write to stdout; annotations write to stdout (GHA
   reads stdout for workflow commands).
5. Error parsers are built-in for: PHPStan, Psalm, ESLint, PHP CS Fixer,
   golangci-lint, rustc, tsc, pytest. Extensible via regex patterns in config.

## Config

```toml
[ci]
output = "github"   # override auto-detection
error_parsers = [
  { name = "custom", pattern = "^(?P<file>[^:]+):(?P<line>\\d+): (?P<msg>.+)$" }
]
```

## Out of Scope
- Posting PR review comments via API (that's the webhook feature-037 territory).
- Parsing arbitrary tool outputs beyond the built-in list without config.
