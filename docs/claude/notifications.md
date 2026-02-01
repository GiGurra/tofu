# OS Notifications

Get notified when Claude sessions need attention.

## Overview

Tofu can send OS notifications when Claude sessions transition to states that require user attention (idle, awaiting permission, awaiting input). This is useful when running multiple sessions or working in a different window.

**Disabled by default** - requires explicit configuration to enable.

## Configuration

Create `~/.tofu/config.json`:

```json
{
  "notifications": {
    "enabled": true,
    "transitions": [
      {"from": "working", "to": "idle"},
      {"from": "working", "to": "awaiting_permission"},
      {"from": "working", "to": "awaiting_input"},
      {"from": "*", "to": "exited"}
    ],
    "cooldown_seconds": 5
  }
}
```

### Options

| Field | Description | Default |
|-------|-------------|---------|
| `enabled` | Master switch for notifications | `false` |
| `transitions` | List of state transitions that trigger notifications | See below |
| `cooldown_seconds` | Minimum seconds between notifications per session | `5` |

### Transitions

Each transition rule has `from` and `to` fields. Use `*` as a wildcard to match any state.

**Default transitions:**
- `working` → `idle` - Claude finished processing
- `working` → `awaiting_permission` - Claude needs permission to proceed
- `working` → `awaiting_input` - Claude is asking a question
- `*` → `exited` - Session ended

**Available states:** `working`, `idle`, `awaiting_permission`, `awaiting_input`, `exited`

### Examples

Only notify when permission is needed:
```json
{
  "notifications": {
    "enabled": true,
    "transitions": [
      {"from": "working", "to": "awaiting_permission"}
    ]
  }
}
```

Notify on any state change from working:
```json
{
  "notifications": {
    "enabled": true,
    "transitions": [
      {"from": "working", "to": "*"}
    ]
  }
}
```

## Notification Content

Notifications display:
- **Title:** `Claude: <state>` (e.g., "Claude: Idle")
- **Body:** Session ID, project name, and conversation title/prompt

## Platform Support

| Platform | Method |
|----------|--------|
| Linux | `notify-send` via libnotify |
| macOS | Native notification center |
| Windows | Toast notifications |
| WSL | PowerShell toast notifications (automatic fallback) |

## Troubleshooting

### Notifications not appearing

1. Check that `~/.tofu/config.json` exists and has valid JSON
2. Verify `"enabled": true` is set
3. Ensure Claude hooks are installed: `tofu claude session install-hooks`
4. Check that the session state transition matches your configured rules

### WSL-specific issues

WSL requires PowerShell access for notifications. Tofu automatically falls back to PowerShell when running in WSL. If notifications still don't work:

1. Verify PowerShell is accessible: `ls /mnt/c/Windows/System32/WindowsPowerShell/v1.0/powershell.exe`
2. Check Windows notification settings allow toast notifications

### Cooldown

If you're not seeing notifications for rapid state changes, it's likely the cooldown. Notifications are rate-limited per session to prevent spam. Adjust `cooldown_seconds` if needed.
