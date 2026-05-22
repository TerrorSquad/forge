# Feature 029: booster publish (Preset Sharing)

## Summary
`booster publish` packages the current `booster.toml` as a shareable URL or
gist so other teams can reference it via `booster init --preset <url>`.

## Motivation
Pairs with feature-025 (remote presets). Authoring a shared preset should be
as easy as running one command.

## Functional Requirements

### Publish to GitHub Gist

```sh
booster publish --gist              # requires GITHUB_TOKEN or gh CLI
```

1. Reads `booster.toml` from the repo root.
2. Strips any local-only comments or secrets (prompts for confirmation of
   what will be published).
3. Creates a public GitHub Gist named `booster.toml` with a description
   `booster preset`.
4. Prints the raw Gist URL that can be passed to `booster init --preset`.

### Publish to clipboard / stdout

```sh
booster publish --stdout            # print to stdout
booster publish --copy              # copy raw URL to clipboard (requires xclip/pbcopy)
```

### Update existing preset

```sh
booster publish --gist --update <gist-id>    # update existing gist
```

## Security

- Content review step before publishing: displays the TOML and requires
  explicit confirmation.
- No secrets scanning (out of scope for v1 of this feature) — operator
  responsibility.
- `--gist` requires either `GITHUB_TOKEN` env var or `gh` CLI logged in.

## Out of Scope
- Private gists.
- Version tagging for published presets.
- A central booster preset registry.
