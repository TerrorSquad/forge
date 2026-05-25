# Feature 027: post-commit Hook Support

## Summary
Support the `post-commit` git hook for lightweight side-effects that run
after a successful commit (notifications, logging, tag suggestions).

## Motivation
Teams want to trigger actions after a commit without blocking it. Common
use cases: desktop notifications, updating a local changelog, printing a
reminder to push.

## Git Protocol
`post-commit` takes no arguments and receives no stdin. It runs after the
commit object is created. A non-zero exit does NOT abort the commit (git
ignores the exit code) but forge will still print the failure.

## Functional Requirements

1. `post-commit` is added to `supportedHooks` and a shim is installed.
2. Tools run sequentially with `pass_files = false` by default (no staged
   files exist at this point).
3. forge propagates non-zero exit codes to the terminal for visibility,
   but the commit is already done — no rollback is possible or attempted.
4. A banner line makes it clear this runs post-commit:
   ```
   post-commit (informational — commit already saved)
   ```
5. `SKIP_POSTCOMMIT=1` skips the hook.

## Example Config

```toml
[hooks.post-commit]
enabled = true

[hooks.post-commit.tools.notify]
command    = "notify-send"
args       = ["Committed!", "forge: commit hook passed"]
pass_files = false
on_failure = "continue"     # never block on notification failure
```

## Out of Scope
- Rolling back or amending commits from post-commit.
- post-merge, post-rewrite hooks (separate features).
