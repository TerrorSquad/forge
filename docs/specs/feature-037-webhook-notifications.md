# Feature 037: Webhook & Notification Integrations

## Summary
Post hook results to external systems — Slack, Microsoft Teams, generic
webhooks — so team leads get real-time visibility into pre-push failures on
shared branches (protect, main, release/*) without watching every developer's
terminal.

## Motivation
Pre-commit hooks are local and invisible to the team. When a developer pushes
a failing build to a shared branch, the first the team knows is CI going red
minutes later. Webhook notifications on pre-push failure bring shared-branch
health into the team's communication channel in real time.

## Config Design

```toml
[notifications.slack]
enabled    = true
webhook    = "https://hooks.slack.com/services/T.../B.../..."
on         = ["pre-push"]         # which hooks to notify
only_when  = ["fail"]             # "pass", "fail", or both
branches   = ["main", "develop"]  # restrict to specific branches (glob)

[notifications.teams]
enabled  = true
webhook  = "https://outlook.office.com/webhook/..."
on       = ["pre-push"]
only_when = ["fail"]

[notifications.webhook]
enabled  = true
url      = "https://my-internal-tool.company.com/hooks/forge"
method   = "POST"
headers  = { "Authorization" = "Bearer ${HOOK_TOKEN}" }
on       = ["pre-push", "pre-commit"]
only_when = ["fail", "pass"]
```

Environment variable interpolation (`${VAR}`) is supported in `webhook`,
`url`, and `headers` values to avoid secrets in the config file.

## Payload Formats

### Slack (Block Kit)
```json
{
  "blocks": [
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "❌ *forge pre-push failed* on `develop` by @alex\n>  `tests` failed after 12.3s\n> 3 passed · 1 failed"
      }
    }
  ]
}
```

### Generic Webhook (JSON)
```json
{
  "event":    "hook.failed",
  "hook":     "pre-push",
  "branch":   "develop",
  "user":     "alex",
  "hostname": "dev-machine",
  "duration_ms": 12300,
  "tools": [
    { "name": "tests", "status": "fail", "duration_ms": 12100 }
  ]
}
```

## Functional Requirements

1. Notifications fire asynchronously after the hook exits — they never delay
   or block the result.
2. Delivery is best-effort: network errors are logged to stderr (not as hook
   failures).
3. `branches` filtering uses the same glob/regex syntax as feature-032
   profiles.
4. `only_when` controls whether to notify on pass, fail, or both.
5. Sensitive values (webhook URLs, tokens) may be set via env vars:
   `${ENV_VAR_NAME}` syntax in any string config value.
6. `forge doctor` validates all configured webhook URLs are reachable
   (HEAD request with 5s timeout).
7. `FORGE_NO_NOTIFY=1` suppresses all notifications (see feature-029).

## Security Considerations
- Never log webhook URLs or tokens.
- Always use HTTPS (reject `http://` unless `allow_insecure = true` is set).
- `Content-Type: application/json` only; no user-controlled MIME types.

## Out of Scope
- OAuth flows or API key rotation.
- PagerDuty / OpsGenie integrations (use the generic webhook).
- Notification history / read receipts.
