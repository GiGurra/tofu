# TOFU - Tools Of Fantastic Utility

[![CI Status](https://github.com/GiGurra/tofu/actions/workflows/ci.yml/badge.svg)](https://github.com/GiGurra/tofu/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/GiGurra/tofu)](https://goreportcard.com/report/github.com/GiGurra/tofu)

Cross-platform CLI tools written in Go. 50+ commands that work consistently on Windows, macOS, and Linux.

```bash
go install github.com/gigurra/tofu@latest
# Install Claude hooks (optional, enables status tracking)
tofu claude session install-hooks
```

---

## Highlights

### Claude Code Integration

![Claude Demo](claude/demo.gif)

Session and conversation management for [Claude Code](https://claude.ai/code):

```bash
tofu claude session new          # Start Claude in a tmux session
tofu claude session ls -w        # Interactive session browser with search
tofu claude conv ls -w           # Interactive conversation browser
tofu claude conv ls -g -w        # Search across all projects
```

Features: tmux sessions, status tracking, interactive watch modes, search/filter/sort.

**[Full documentation →](claude/index.md)**

---

## All Commands

Browse the sidebar for the full command reference, organized by category:

- **File Operations** - cat, grep, find, ls, tree, diff, etc.
- **Network** - http, serve, ping, dns, wget, etc.
- **System** - ps, free, watch, env, etc.
- **Encoding** - base64, jwt, hash, qr, uuid
- **DevOps** - k8s, gh, git
- **Fun** - bird, cowsay, fortune, and more

---

### Fun Stuff

Because productivity isn't everything:

```bash
tofu bird                            # Flappy Tofu game
tofu excuse                          # Programmer excuses
tofu blame                           # Random blame generator
tofu cowsay "Hello!"                 # ASCII cow
tofu fortune                         # Programming wisdom
tofu busy                            # Fake progress bar
```

---

## Features

- Works on Windows, macOS, and Linux
- Colored output
- Shell completions (bash, zsh, fish, PowerShell)
- Smart defaults
- Fast Go performance
