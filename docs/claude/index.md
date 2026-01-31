# Claude Code Integration

Powerful session and conversation management for [Claude Code](https://claude.ai/code).

![Demo](demo.gif)

## Features

- **Session Management** - Run Claude in tmux sessions, attach/detach anytime
- **Status Tracking** - See when Claude is working, idle, or waiting for input
- **Interactive Watch Modes** - Browse sessions and conversations with search, filtering, sorting
- **Session Indicators** - Know which conversations have active sessions (⚡ attached, ○ active)

## Installation

After installing tofu, install the Claude hooks for status tracking:

```bash
# Install tofu
go install github.com/gigurra/tofu@latest

# Install Claude hooks (enables status tracking)
tofu claude session install-hooks
```

This adds hooks to `~/.claude/settings.json` that report Claude's status (working, idle, awaiting input) to tofu.

## Quick Start

```bash
# Start Claude in a new tmux session
tofu claude session new

# Or resume an existing conversation
tofu claude session new --resume <conv-id>

# Interactive session browser
tofu claude session ls -w

# Interactive conversation browser
tofu claude conv ls -w
```

## Commands

| Command | Description |
|---------|-------------|
| `session new` | Start Claude in a tmux session |
| `session ls` | List sessions (`-w` for interactive) |
| `session attach` | Attach to a session |
| `session kill` | Kill sessions |
| `conv ls` | List conversations (`-w` for interactive, `-g` for global) |
| `conv search` | Search conversation text |
| `conv resume` | Resume a conversation |
| `conv delete` | Delete a conversation |
| `conv prune-empty` | Delete empty conversations |

## Interactive Watch Mode Keys

Both `session ls -w` and `conv ls -w` support these keys:

| Key | Action |
|-----|--------|
| `/` | Start search |
| `↑`/`↓` or `j`/`k` | Navigate |
| `Enter` | Attach/create session |
| `Del`/`x` | Delete/kill (with confirmation) |
| `h` or `?` | Show help |
| `Esc` | Clear search / quit |
| `q` | Quit |

Session watch also supports:

| Key | Action |
|-----|--------|
| `f` | Filter menu (by status) |
| `1`-`5` | Sort by column |

## Documentation

- [Session Management](sessions.md) - Detailed session commands
- [Conversation Management](conversations.md) - Detailed conversation commands

## Recording a Demo

The `demo.tape` file is a [VHS](https://github.com/charmbracelet/vhs) script:

```bash
# Install VHS and dependencies
go install github.com/charmbracelet/vhs@latest
sudo apt-get install -y ffmpeg ttyd

# Record
cd docs/claude
vhs demo.tape  # Outputs demo.gif and demo.mp4
```
