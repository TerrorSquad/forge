# Feature 008: Environment Variable Skip Controls

## Summary
Allow developers to skip the entire hook pipeline or individual tools at runtime
via environment variables, without editing config files.

## Motivation
During hot-fixes, bisects, or manual quality-check sessions developers need a way
to commit quickly while still keeping the hook pipeline active for normal flow.
Editing `booster.toml` for a temporary bypass is risky and creates noisy diffs.

## Functional Requirements

### Hook-level skips
| Variable         | Behaviour                           |
|------------------|-------------------------------------|
| `SKIP_PRECOMMIT` | Skip entire pre-commit hook         |
| `SKIP_COMMITMSG` | Skip entire commit-msg hook         |
| `SKIP_PREPUSH`   | Skip entire pre-push hook           |

### Tool-level skips
- Variable name: `SKIP_<NORMALIZED_TOOL_NAME>=1`
- Normalisation rule: uppercase the tool name, replace every non-alphanumeric
  character sequence with a single underscore, strip leading/trailing underscores.
- Examples:
  - `ESLint` → `SKIP_ESLINT`
  - `PHP Syntax Check` → `SKIP_PHP_SYNTAX_CHECK`
  - `markdownlint-cli2` → `SKIP_MARKDOWNLINT_CLI2`

### Group filtering
- `HOOKS_ONLY=<group1>,<group2>` runs only tools whose `group` matches one of the
  listed values (case-insensitive). Tools with no group are always executed.

## Accepted truthy values
`1`, `true`, `yes`, `on` (case-insensitive).

## Logging
- Every skipped tool must emit a `skipped (<reason>)` line on stdout.

## Security
- Skip variables must never be loaded from `booster.toml`; they are runtime-only.
