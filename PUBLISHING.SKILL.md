# Forge Publishing Skill

## Commit conventions

- Use conventional commit messages: `fix(...)`, `feat(...)`, `docs(...)`, `chore(...)`.
- Keep commit descriptions short and focused.
- Include enough context for reviewers when the change affects user-facing behavior.

## Pull request workflow

1. Create a feature branch for the change.
2. Run tests locally: `go test ./internal/forge/...`.
3. Update docs if behavior or install expectations change.
4. Commit and push the branch.
5. Open a PR with a clear title and description.

## Review guidance

- Confirm tests pass and relevant docs were updated.
- For hook and install changes, verify the `forge install` / `forge doctor` flow is described clearly.
- If the change affects repo setup, mention whether `.forge` should be committed and whether `forge install` is still required.

## Release readiness

- `forge` behavior is production-ready once tests pass and install/setup docs are clear.
- For user-facing changes, ensure quick-start and README instructions are aligned.
