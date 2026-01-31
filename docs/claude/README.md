# Claude Session & Conversation Management

Documentation and demos for the `tofu claude` commands.

## Demo Recording

The `demo.tape` file is a [VHS](https://github.com/charmbracelet/vhs) script that records a demo of the session and conversation management features.

### Prerequisites

```bash
# Install VHS
go install github.com/charmbracelet/vhs@latest

# Install dependencies (Ubuntu/Debian)
sudo apt-get install -y ffmpeg ttyd
```

### Recording the Demo

```bash
cd docs/claude
vhs demo.tape
```

This generates:
- `demo.gif` - Animated GIF for README/docs
- `demo.mp4` - Video with timeline for reviewing

### What the Demo Shows

1. **Conversation watch mode** (`conv ls -w`) - Interactive list with search
2. **Global conversation search** (`conv ls -g -w`) - Search across all projects
3. **Session creation** (`session new -d`) - Create detached tmux sessions
4. **Session watch mode** (`session ls -w`) - Interactive list with search, filtering, sorting
5. **Session indicators** - Shows which conversations have active sessions

## Features

### Session Management

- `tofu claude session new` - Start Claude in a tmux session
- `tofu claude session ls` - List sessions (add `-w` for interactive mode)
- `tofu claude session attach` - Attach to a session
- `tofu claude session kill` - Kill sessions

### Conversation Management

- `tofu claude conv ls` - List conversations (add `-w` for interactive mode, `-g` for global)
- `tofu claude conv search` - Search conversation content
- `tofu claude conv resume` - Resume a conversation in a new session

### Interactive Watch Mode Keys

| Key | Action |
|-----|--------|
| `/` | Start search |
| `↑`/`↓` | Navigate |
| `Enter` | Attach/create session |
| `Del`/`x` | Delete/kill (with confirmation) |
| `h` | Show help |
| `f` | Filter menu (sessions only) |
| `q` | Quit |
