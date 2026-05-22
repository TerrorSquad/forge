# Feature 006: Doctor and Observability

## Summary
Provide a `doctor` command that reports setup status, config resolution, hook installation state, and missing tool binaries.

## Motivation
Git hook failures are frequently environment/setup issues. A deterministic diagnostic command reduces support burden and time to fix.

## Functional Requirements
1. Print path to active `booster` binary.
2. Detect and print git repository root.
3. Print selected config path or explain why config is missing.
4. Print current `core.hooksPath` value.
5. If using `.booster/hooks`, report status per managed hook file.
6. Identify configured tool commands missing from PATH.

## UX Requirements
- Output must be human-readable in CI logs.
- Missing items should be grouped clearly.

## Out of Scope
- JSON output mode in v1.
- Automatic remediation commands.
