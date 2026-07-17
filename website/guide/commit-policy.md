# Commit-message Policy

forge can validate and mutate commit messages via the `commit-msg` hook.

## Enable

```toml
[hooks.commit-msg]
enabled = true

[hooks.commit-msg.policy]
conventional_commits = true
```

## Policy fields

These apply on the **`commit-msg`** hook:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `conventional_commits` | bool | `false` | Enforce [Conventional Commits](https://www.conventionalcommits.org/) format |
| `allowed_types` | `[]string` | *(default set)* | Override the allowed Conventional Commits types (see below) |
| `append_ticket_footer` | bool | `false` | Append `Closes: PRJ-123` from the branch name |
| `footer_label` | string | `Closes` | Label used for the appended footer (e.g. `Refs`) |
| `require_ticket` | bool | `false` | Fail if the current branch has no ticket ID |
| `ticket_pattern` | string | `([A-Z]+-[0-9]+)` | Regex (with a capture group) used to extract the ticket ID from the branch name |
| `validate_branch_name` | bool | `false` | Fail if the branch name doesn't match `branch_pattern` |
| `branch_pattern` | string | — | Regex the branch name must match when `validate_branch_name = true` |
| `skipped_branches` | `[]string` | `[]` | Exact branch names to skip the policy on (e.g. `main`, `develop`) |

These apply on the **`prepare-commit-msg`** hook (enable `[hooks.prepare-commit-msg]`):

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `prepend_ticket` | bool | `false` | Prepend `PRJ-123: ` to the subject before the editor opens |
| `skip_if_present` | bool | `false` | Skip prepending if the ticket is already in the message |
| `skip_on_merge` | bool | `false` | Skip prepending on merge/squash commits |

## Conventional Commits

When `conventional_commits = true`, commit messages must start with one of:

```
feat | fix | chore | docs | style | refactor | perf | test | build | ci | revert
```

Optionally with a scope: `feat(auth): add OAuth2 support`

## Ticket footer appending

When `append_ticket_footer = true`, forge reads the current branch name, extracts a JIRA-style ticket ID (e.g. `PRJ-123`), and appends a footer:

```
feat(auth): add OAuth2 support

Closes: PRJ-123
```

The ticket is extracted from the branch name using `ticket_pattern` (default `([A-Z]+-[0-9]+)`), so a branch like `feature/PRJ-123-oauth` yields `PRJ-123`. For GitHub issues, set `ticket_pattern = "(#[0-9]+)"`.

## Requiring a ticket

When `require_ticket = true`, the commit fails if no ticket ID is found in the branch name:

```
✗  commit-msg: branch 'fix/typo' contains no ticket ID
```

## Skip policy check once

```sh
SKIP_COMMITMSG=1 git commit -m "wip: quick save"
```

## See also

- [Hooks](/guide/hooks)
- [forge.toml reference](/reference/config)
