# Feature 012: Config Migration Tool

## Summary
Provide a `forge migrate` command that reads a
`.git-hooks.config.json` (the custom Husky/zx format) and outputs a valid
`forge.toml`.

## Motivation
Existing repositories that use the predecessor TypeScript-based hook engine
should be able to adopt forge without manually transcribing configuration.

## Source Format
The source config follows the JSON schema at:
`https://raw.githubusercontent.com/TerrorSquad/php-forge/main/forge/.husky/.git-hooks.config.schema.json`

Key fields to map:

| Source (JSON)                   | Target (TOML)                                  |
|---------------------------------|------------------------------------------------|
| `hooks.preCommit.tools.<Name>`  | `[hooks.pre-commit.tools.<name>]`              |
| `tools.<Name>.command`          | `command`                                      |
| `tools.<Name>.args`             | `args`                                         |
| `tools.<Name>.type`             | `type`                                         |
| `tools.<Name>.extensions`       | `extensions`                                   |
| `tools.<Name>.stagesFilesAfter` | `restage`                                      |
| `tools.<Name>.passFiles`        | `pass_files`                                   |
| `tools.<Name>.runForEachFile`   | `run_per_file`                                 |
| `tools.<Name>.onFailure`        | `on_failure`                                   |
| `tools.<Name>.group`            | `group`                                        |
| `skip.preCommit`                | `hooks.pre-commit.enabled = false`             |

## Functional Requirements
1. Default source path: `.git-hooks.config.json` in repo root.
2. Accept `--from <path>` flag.
3. Print generated `forge.toml` to stdout by default.
4. `--write` flag writes to `forge.toml` (fails if file exists; use
   `--force` to overwrite).
5. Emit warnings for fields that cannot be automatically mapped.

## CLI
```text
forge migrate [--from .git-hooks.config.json] [--write] [--force]
```

## Out of Scope
- Husky v9 native format migration.
- lint-staged configuration migration.
