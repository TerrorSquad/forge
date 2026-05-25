# Feature 021: prepare-commit-msg Hook Support

## Summary
Support the `prepare-commit-msg` git hook to auto-populate commit message
templates with a ticket prefix derived from the branch name.

## Motivation
Teams with ticket-prefixed branches (e.g. `feat/PRJ-123-add-widget`) want the
ticket number injected automatically as the commit message prefix without
relying on the developer to type it.

## Git Protocol
Git invokes `prepare-commit-msg <file> [<source> [<sha1>]]` where `source` is
`message`, `template`, `merge`, `squash`, or `commit`.

## Functional Requirements

1. forge reads `[hooks.prepare-commit-msg.policy]`:
   ```toml
   [hooks.prepare-commit-msg]
   enabled = true

   [hooks.prepare-commit-msg.policy]
   prepend_ticket  = true     # insert "PRJ-123: " at start of subject
   skip_on_merge   = true     # do nothing for merge/squash commits
   skip_if_present = true     # do nothing if ticket already in message
   ```
2. Ticket is extracted from branch name using the same `ticketRegex`
   (`[A-Z]+-[0-9]+`) as commit-msg policy.
3. When `prepend_ticket = true` and a ticket is found, the message file is
   rewritten to prepend `TICKET: ` to the subject line.
4. `skip_on_merge` suppresses the hook when `source` is `merge` or `squash`.
5. `skip_if_present` suppresses rewrite if the ticket already appears in
   the message.
6. Custom tools can also be configured under
   `[hooks.prepare-commit-msg.tools]` for advanced templating.

## Out of Scope
- AI-generated commit messages.
- Interactive editor integration.
