# Feature 040: `forge explain` — Tool Documentation

## Summary
Add `forge explain <tool>` and `forge explain <hook>` commands that print
human-readable documentation about a configured tool: what it does, which env
vars skip it, what SKIP controls exist, and why it was added (from an optional
`description` field in config).

## Motivation
New developers joining a project see opaque tool names in the hook output
(`✗ deptrac`) but don't know what deptrac is, why it's there, or how to
temporarily skip it. `forge explain` turns the config into living
documentation — answering "what is this?" and "how do I get past it?" without
leaving the terminal.

## CLI Interface

```
forge explain <tool>            # explain a specific tool across all hooks
forge explain pre-commit        # explain all tools in a hook
forge explain                   # explain everything (full reference)
forge explain --hook pre-push   # explain a specific hook's tools
```

### Example Output

```
$ forge explain phpstan

╭─ phpstan (pre-commit · group: analysis) ──────────────────────────────────╮
│ PHPStan — PHP static analysis tool                                         │
│                                                                             │
│ Detects type errors, undefined methods, unreachable code, and other bugs   │
│ at compile time without executing the code.                                 │
│                                                                             │
│ Config  command: vendor/bin/phpstan                                         │
│         args:    analyse                                                    │
│         level:   6 (via phpstan.neon)                                       │
│         cache:   enabled                                                    │
│                                                                             │
│ Skip    SKIP_PHPSTAN=1   (per-tool)                                         │
│         SKIP_GROUP_ANALYSIS=1   (entire analysis group)                     │
│                                                                             │
│ Docs    https://phpstan.org                                                 │
╰─────────────────────────────────────────────────────────────────────────────╯
```

```
$ forge explain pre-commit

pre-commit hook  (7 tools configured)

  format group
    rector              Automated refactoring and code modernization (PHP).
                        Skip: SKIP_RECTOR=1
    ecs                 PHP Coding Standards Fixer.   Skip: SKIP_ECS=1
    multiline-attrs     Normalize PHP attribute formatting.  Skip: SKIP_MULTILINE_ATTRIBUTES=1

  analysis group
    phpstan             Static type analysis.   Skip: SKIP_PHPSTAN=1
    psalm               Deep type inference.    Skip: SKIP_PSALM=1
    deptrac             Architecture layer enforcement.   Skip: SKIP_DEPTRAC=1

  artifacts group
    deptrac-image       Regenerate architecture diagram (PNG).  Skip: SKIP_DEPTRAC_IMAGE=1
```

## Config: `description` and `docs_url`

```toml
[hooks.pre-commit.tools.phpstan]
command     = "vendor/bin/phpstan"
args        = ["analyse"]
description = "PHPStan: static type analysis. Catches type errors at commit time."
docs_url    = "https://phpstan.org"
group       = "analysis"
```

When `description` is absent, forge falls back to a built-in knowledge base
of well-known tools (phpstan, psalm, eslint, prettier, golangci-lint, etc.).

## Functional Requirements

1. `forge explain <tool>` searches all configured hooks for the named tool.
   If found in multiple hooks, shows all.
2. `forge explain <hook>` lists all tools with one-line summaries.
3. SKIP variable names are derived automatically from tool names
   (`SKIP_<UPPER_SNAKE(name)>=1`) — no extra config needed.
4. Group skip vars (feature-035) are listed alongside per-tool vars.
5. `forge explain` with no args prints a condensed reference of every tool
   in every hook.
6. The output respects terminal width (wraps at 80 chars if narrower than 100).

## Built-in Tool Knowledge Base

| Tool name pattern | Description |
|-------------------|-------------|
| `phpstan`         | PHPStan — PHP static analysis |
| `psalm`           | Psalm — PHP type inference and static analysis |
| `ecs`             | Easy Coding Standard — PHP code style fixer |
| `rector`          | Rector — automated PHP refactoring |
| `deptrac`         | Deptrac — architecture layer enforcement |
| `prettier`        | Prettier — opinionated code formatter |
| `eslint`          | ESLint — JavaScript/TypeScript linter |
| `golangci-lint`   | golangci-lint — Go linter aggregator |
| `gofmt`           | gofmt — Go code formatter |
| `govet`           | go vet — Go static analysis |
| `spectral`        | Spectral — OpenAPI/AsyncAPI linter |

## Out of Scope
- Fetching live docs from tool websites.
- Interactive search / fuzzy matching.
- Generating HTML documentation.
