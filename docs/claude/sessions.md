# Session Management 📺

Run Claude Code in persistent tmux sessions with status tracking.

## Prerequisites

- **tmux** - Required for session management
- **Claude hooks** - For status tracking (install with `tofu claude session install-hooks`)

## Commands

### session new

Start Claude in a new tmux session.

```bash
# Start a new session in current directory
tofu claude session new

# Start in a specific directory
tofu claude session new /path/to/project

# Resume an existing conversation
tofu claude session new --resume <conv-id>

# Start detached (don't attach immediately)
tofu claude session new -d
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-d, --detach` | Start session without attaching |
| `--resume <id>` | Resume an existing conversation |

### session ls

List active sessions.

```bash
# List sessions
tofu claude session ls

# Interactive watch mode
tofu claude session ls -w

# Include exited sessions
tofu claude session ls -a

# Filter by status
tofu claude session ls --status idle
tofu claude session ls --hide exited
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-w, --watch` | Interactive watch mode |
| `-a, --all` | Include exited sessions |
| `--status <s>` | Show only this status |
| `--hide <s>` | Hide this status |
| `--sort <col>` | Sort by: id, dir, status, age, updated |

**Status values:** `idle`, `working`, `awaiting-permission`, `awaiting-input`, `exited`

### session attach

Attach to an existing session.

```bash
# Attach by session ID
tofu claude session attach <id>

# Force attach (detach other clients)
tofu claude session attach -d <id>
```

### session kill

Kill one or more sessions.

```bash
# Kill a specific session
tofu claude session kill <id>

# Kill all sessions
tofu claude session kill --all

# Kill only idle sessions
tofu claude session kill --idle

# Force (no confirmation)
tofu claude session kill -f <id>
```

### session install-hooks

Install Claude hooks for status tracking.

```bash
tofu claude session install-hooks
```

This modifies `~/.claude/settings.json` to add hooks that report:
- When Claude starts/stops working
- When Claude is waiting for permission or input
- Session status changes

## Interactive Watch Mode

Press `w` or use `-w` flag to enter interactive mode.

### Navigation

| Key | Action |
|-----|--------|
| `↑`/`k` | Move up |
| `↓`/`j` | Move down |
| `Enter` | Attach to session |
| `q`/`Esc` | Quit |

### Search

| Key | Action |
|-----|--------|
| `/` | Start search |
| `Esc` | Clear search / exit search mode |
| `Ctrl+U` | Clear search input |
| `↑`/`↓` | Exit search and navigate |

### Actions

| Key | Action |
|-----|--------|
| `Del`/`x` | Kill session (with confirmation) |
| `r` | Refresh list |
| `h`/`?` | Show help |

### Filtering

| Key | Action |
|-----|--------|
| `f` | Open filter menu |
| `Space` | Toggle filter option |
| `Enter` | Apply filter |

### Sorting

| Key | Action |
|-----|--------|
| `1`/`F1` | Sort by ID |
| `2`/`F2` | Sort by Directory |
| `3`/`F3` | Sort by Status |
| `4`/`F4` | Sort by Age |
| `5`/`F5` | Sort by Updated |

Press the same key again to toggle ascending/descending/off.

## Session Status 🔮

Sessions report their status via Claude hooks:

| Status | Color | Description |
|--------|-------|-------------|
| `idle` | 🟡 Yellow | Claude is waiting for input |
| `working` | 🟢 Green | Claude is processing |
| `awaiting-permission` | 🔴 Red | Needs permission approval |
| `awaiting-input` | 🔴 Red | Waiting for user input |
| `exited` | ⚫ Gray | Session has ended |

## Tmux Integration

Sessions run in tmux with the naming convention `tofu-claude-<id>`.

```bash
# List all tofu tmux sessions
tmux ls | grep tofu-claude

# Manually attach
tmux attach -t tofu-claude-abc123

# Detach from inside tmux
Ctrl+B D
```
