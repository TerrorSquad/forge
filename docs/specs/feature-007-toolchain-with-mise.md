# Feature 007: Toolchain Management with mise

## Summary
Use `mise` to pin and install Go version at repository level for reproducible local builds.

## Motivation
Contributors need deterministic compiler behavior and minimal setup drift across machines.

## Functional Requirements
1. Repository must include `mise.toml` with pinned Go version.
2. Setup steps should support:
   - `mise install`
   - `mise exec -- go version`
3. Build docs should prefer `mise exec --` for commands requiring pinned toolchain.

## Versioning Policy
- Pin a specific major/minor/patch version (for example `1.23.8`), not floating `latest`.
- Update intentionally through reviewed change.

## Team Workflow
- New contributors run `mise install` once.
- CI may either use mise or native Go setup, but local experience should remain consistent.

## Out of Scope
- Automatic installation of project-level linting binaries.
