# Feature 005: Commit Message Policy

## Summary
Implement policy checks/mutations for the `commit-msg` hook, including conventional commit validation and optional ticket footer enrichment.

## Motivation
Teams need reliable commit metadata for changelogs, release automation, and traceability to ticketing systems.

## Functional Requirements
1. Read commit message file from:
   - `--edit` flag
   - positional fallback argument
2. When enabled, validate first line against conventional commit format.
3. Detect ticket identifiers (`[A-Z]+-[0-9]+`) from current branch name.
4. When enabled, append footer `Closes: <ticket>` if missing.
5. When `require_ticket = true`, reject commit if branch has no ticket.

## Edge Cases
- Empty commit message should fail.
- Existing footer should not be duplicated.
- Footer append should preserve existing content and ensure newline termination.

## Future Extensions
- Configurable regex per organization.
- JIRA/YouTrack style footer templates.
