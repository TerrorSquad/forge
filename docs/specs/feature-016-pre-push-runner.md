# Feature 016: Pre-Push Runner

## Summary
Wire the `pre-push` hook to actually execute configured tools, receiving the
remote name and URL from git's stdin payload.

## Motivation
`pre-push` is installed as a shim but currently runs nothing. Teams want to
run `go test ./...` or `phpunit` before a push without writing custom shell
scripts.

## Git Protocol
Git passes remote info on stdin, one line per ref being pushed:
```
<local-ref> <local-sha1> <remote-ref> <remote-sha1>
```
And sets env vars `$1` = remote name, `$2` = remote URL.

## Functional Requirements

1. forge reads stdin lines and exposes them to tools as env vars:
   - `FORGE_PUSH_REMOTE` — remote name (e.g. `origin`)
   - `FORGE_PUSH_URL` — remote URL
   - `FORGE_PUSH_BRANCH` — local branch being pushed
2. Tools configured under `[hooks.pre-push]` run sequentially (or parallel if
   enabled).
3. `pass_files` defaults to `false` for pre-push tools (no staged-file list).
4. Any non-zero exit aborts the push.

## Example Config

```toml
[hooks.pre-push]
enabled = true

[hooks.pre-push.tools.tests]
command = "go"
args    = ["test", "./..."]
pass_files = false
group  = "test"

[hooks.pre-push.tools.govet]
command = "go"
args    = ["vet", "./..."]
pass_files = false
group  = "lint"
```

## Out of Scope
- Filtering push refs (e.g. skip tests when pushing to non-main).
- Pre-receive server-side hooks.
