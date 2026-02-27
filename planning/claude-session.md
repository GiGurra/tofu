# Claude Session Manager

Multiplex and manage multiple Claude Code sessions with detach/reattach, status tracking, and notifications.

## Commands

```bash
tclaude session new [dir]       # Start new detached session (optionally in dir)
tclaude session new --resume <conv-id>  # Resume existing conversation
tclaude session ls              # List all sessions with status
tclaude session attach <id>     # Attach to session (Ctrl+B D to detach)
tclaude session kill <id>       # Kill a session
tclaude session watch           # Live dashboard (like top)
```

## Architecture

### Dependencies
- **tmux**: Required for session management. Error with install instructions if missing.

### State Storage

Directory: `~/.tofu/claude-sessions/`

Each session has a state file: `<session-id>.json`
```json
{
  "id": "abc123",
  "tmuxSession": "tofu-claude-abc123",
  "pid": 12345,
  "cwd": "/home/user/project",
  "convId": "0789725a-bc71-47dd-9ca5-1b4fe7aead9b",
  "status": "waiting_input",
  "statusDetail": "",
  "created": "2026-01-31T12:00:00Z",
  "updated": "2026-01-31T12:05:00Z"
}
```

### Status Values
| Status | Meaning |
|--------|---------|
| `running` | Claude is actively working (tool execution, thinking) |
| `waiting_input` | Waiting for user to type next prompt |
| `waiting_permission` | Waiting for user to approve a tool/command |
| `exited` | Session ended (Claude exited) |

### State Detection via Claude Code Hooks

Configure hooks when spawning session. Claude Code hooks can run shell commands on events.

Relevant hooks (in `~/.claude/settings.json` or project `.claude/settings.json`):
- `PreToolUse` - about to run a tool (could detect permission prompts)
- `PostToolUse` - tool finished
- `Stop` - session stopped

The hook script updates the session state file:
```bash
#!/bin/bash
# ~/.tofu/bin/claude-session-hook.sh
echo '{"status": "'"$1"'", "updated": "'"$(date -Iseconds)"'"}' > ~/.tofu/claude-sessions/$SESSION_ID.status
```

### Preventing Duplicate Sessions

Before starting a session for a conversation:
1. Check all session state files
2. If any active session has the same `convId`, refuse to start
3. Verify the session is actually alive (check PID)

### Dashboard (`watch` command)

- Poll state files every 1-2 seconds (or use fsnotify)
- Display table similar to `top`:
  ```
  ID       PROJECT              STATUS             UPDATED
  abc123   /home/user/proj1     waiting_input      2m ago
  def456   /home/user/proj2     running            5s ago
  ghi789   /home/user/proj3     waiting_permission 30s ago
  ```
- Highlight sessions needing attention (waiting_input, waiting_permission)
- Could add keyboard shortcuts: a=attach, k=kill, q=quit

### Notifications

When status changes to `waiting_input` or `waiting_permission`:
- Linux: D-Bus
- macOS: `osascript -e 'display notification ...'`
- Could be opt-in via config or flag

## Implementation Phases

### Phase 1: Basic session management (this session)
- [x] Planning doc
- [ ] `session new` - spawn tmux session with claude
- [ ] `session ls` - list sessions from state files
- [ ] `session attach` - attach to tmux session
- [ ] `session kill` - kill session and cleanup
- [ ] State file creation/cleanup

### Phase 2: Status tracking
- [ ] Hook script for status updates
- [ ] Auto-configure hooks when spawning
- [ ] Status detection and state file updates
- [ ] PID validation for stale sessions

### Phase 3: Dashboard and notifications
- [ ] `session watch` with live updates
- [ ] Desktop notifications on status change
- [ ] Keyboard shortcuts in watch mode

### Phase 4: Polish
- [ ] `--resume <conv-id>` support
- [ ] Duplicate session prevention
- [ ] Better error handling
- [ ] Session naming/labeling

## Open Questions

1. How to reliably detect "waiting for permission"? Need to investigate Claude Code hooks.
2. Should we support Windows? (No tmux, would need different approach)
3. Session timeout/auto-cleanup for dead sessions?
