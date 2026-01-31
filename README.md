# TOFU

[![CI Status](https://github.com/GiGurra/tofu/actions/workflows/ci.yml/badge.svg)](https://github.com/GiGurra/tofu/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/GiGurra/tofu)](https://goreportcard.com/report/github.com/GiGurra/tofu)
[![Docs](https://img.shields.io/badge/docs-gigurra.github.io%2Ftofu-blue)](https://gigurra.github.io/tofu/)

**Tools Of Fantastic Utility** - Cross-platform CLI tools written in Go.

50+ Unix-like commands that work consistently on Windows, macOS, and Linux.

```bash
go install github.com/gigurra/tofu@latest
```

---

## Featured: Claude Code Session Management

![Claude Demo](docs/claude/demo.gif)

Powerful session and conversation management for [Claude Code](https://claude.ai/code):

```bash
tofu claude session new          # Start Claude in a tmux session
tofu claude session ls -w        # Interactive session browser with search
tofu claude conv ls -w           # Interactive conversation browser
tofu claude conv ls -g -w        # Search across all projects
```

- **Tmux integration** - Run Claude in persistent sessions, attach/detach anytime
- **Status tracking** - See when Claude is working, idle, or waiting for input
- **Interactive watch modes** - Browse with search, filtering, and sorting
- **Session indicators** - Know which conversations have active sessions

**[Full documentation →](https://gigurra.github.io/tofu/claude/)**

---

## Highlights

### DevOps

```bash
tofu k8s logs -f -l app=myservice    # Tail logs across pods
tofu gh pr list                       # List GitHub PRs
tofu gh pr open 123                   # Open PR in browser
tofu git sync                         # Pull + push
```

### Security & Encoding

```bash
tofu jwt decode <token>               # Decode JWT payload
tofu hash sha256 file.txt             # Cryptographic hashes
tofu base64 encode "hello"            # Base64 encode/decode
```

### Network

```bash
tofu http get https://api.example.com # HTTP client
tofu serve .                          # Instant file server
tofu dns example.com MX               # DNS lookups
```

### Everyday Tools

```bash
tofu cat file.txt                     # Works in PowerShell!
tofu grep "pattern" *.go              # Search with colors
tofu find . -name "*.json"            # Find files
tofu tree                             # Directory tree
tofu clip                             # Clipboard operations
tofu uuid                             # Generate UUIDs
```

### Fun

```bash
tofu bird                             # Flappy Tofu game
tofu cowsay "Hello!"                  # ASCII art
tofu fortune                          # Programming wisdom
tofu excuse                           # Programmer excuses
```

---

## Why TOFU?

- **Cross-platform** - Same commands on Windows, macOS, and Linux
- **PowerShell-friendly** - No more `grep` not found
- **Sensible defaults** - Commands that just work
- **Shell completions** - For bash, zsh, fish, and PowerShell
- **Single binary** - One `go install` and you're done

---

## Documentation

**[gigurra.github.io/tofu](https://gigurra.github.io/tofu/)**

---

## License

MIT
