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

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `conventional_commits` | bool | `false` | Enforce [Conventional Commits](https://www.conventionalcommits.org/) format |
| `append_ticket_footer` | bool | `false` | Append `Closes: PRJ-123` from branch name |
| `require_ticket` | bool | `false` | Fail if the current branch has no ticket ID |

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

Branch naming convention: any branch containing `PRJ-123` or `PRJ_123`.

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
