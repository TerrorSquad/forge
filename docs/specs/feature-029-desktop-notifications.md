# Feature 029: Desktop Notifications

## Summary
Send a native desktop notification when a hook completes, showing a pass/fail
summary. Especially valuable for slow pre-push hooks where the developer has
switched context while waiting.

## Motivation
Pre-push hooks can take 10–60 seconds (tests, linting, spec generation). Today
developers either stare at the terminal or miss the result because they tabbed
away. A desktop notification brings the result to them instantly without
requiring terminal focus.

## Visual Design

**Success (macOS / Linux)**
```
🎉 booster · pre-push passed
4 passed · 9.3s
```

**Failure**
```
❌ booster · pre-push failed
tests failed · 3 passed · 1 failed · 12.1s
```

Clicking the notification focuses the terminal that ran the hook (if supported
by the OS notification system).

## Functional Requirements

1. Notification is sent after every hook run (pass or fail) when the hook takes
   longer than a configurable threshold.
2. Global config in `booster.toml`:
   ```toml
   [notifications]
   enabled   = true
   threshold = "5s"   # only notify when hook duration exceeds this
   on_pass   = true   # notify on success (default: true)
   on_fail   = true   # notify on failure (default: true)
   ```
3. Per-hook override:
   ```toml
   [hooks.pre-push]
   notify = true
   ```
4. Notification delivery is best-effort — failure to deliver (no daemon,
   missing binary) must never fail the hook.
5. `BOOSTER_NO_NOTIFY=1` suppresses all notifications unconditionally.

## Platform Support

| Platform | Mechanism | Binary |
|----------|-----------|--------|
| macOS    | `osascript` display notification | built-in |
| Linux (GNOME/KDE) | `notify-send` | `libnotify-bin` |
| Linux (fallback) | `wall` broadcast | built-in |
| WSL2 | `powershell.exe New-BurntToastNotification` | requires BurntToast module |

Detection order: `osascript` → `notify-send` → `wall` → silent.

## Non-Functional Requirements
- Notification delivery is fire-and-forget (goroutine, non-blocking).
- No external Go dependencies (use `os/exec` to call system binaries).

## Out of Scope
- Rich notification actions (e.g. "Re-run" button) — OS support is inconsistent.
- Email or webhook notifications (see feature-037).
- Notification history / persistence.
