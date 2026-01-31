# TOFU - Tools Of Fantastic Utility

[![CI Status](https://github.com/GiGurra/tofu/actions/workflows/ci.yml/badge.svg)](https://github.com/GiGurra/tofu/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/GiGurra/tofu)](https://goreportcard.com/report/github.com/GiGurra/tofu)

Cross-platform CLI tools written in Go. 50+ commands that work consistently on Windows, macOS, and Linux.

```bash
go install github.com/gigurra/tofu@latest
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

### DevOps Utilities

**Kubernetes** - Tail logs across pods:
```bash
tofu k8s logs -f -l app=myservice    # Tail all pods matching label
tofu k8s logs -f deployment/myapp    # Tail deployment pods
```

**GitHub** - Quick access to PRs and issues:
```bash
tofu gh pr list                      # List PRs
tofu gh pr open 123                  # Open PR in browser
tofu gh issue list --mine            # Your issues
```

**Git** - Handy shortcuts:
```bash
tofu git sync                        # Pull + push
tofu git open                        # Open repo in browser
```

---

### Security & Encoding

**JWT** - Decode and inspect tokens:
```bash
tofu jwt decode <token>              # Decode JWT payload
echo $TOKEN | tofu jwt decode -      # From stdin
```

**Hash** - Quick cryptographic hashes:
```bash
tofu hash sha256 file.txt            # SHA256 of file
echo "data" | tofu hash md5 -        # MD5 from stdin
```

**Base64** - Encode/decode:
```bash
tofu base64 encode "hello"
tofu base64 decode "aGVsbG8="
```

---

### Network Tools

**HTTP** - Readable HTTP client:
```bash
tofu http get https://api.example.com/data
tofu http post https://api.example.com/data '{"key": "value"}'
```

**Serve** - Instant file server:
```bash
tofu serve .                         # Serve current directory on :8080
tofu serve -p 3000 ./dist            # Custom port
```

**DNS** - Quick lookups:
```bash
tofu dns example.com                 # A records
tofu dns example.com MX              # MX records
```

---

### Everyday Essentials

**Cross-platform classics** that work in PowerShell:
```bash
tofu cat file.txt                    # View files
tofu grep "pattern" *.go             # Search with colors
tofu find . -name "*.json"           # Find files
tofu tree                            # Directory tree (respects .gitignore)
```

**Clipboard**:
```bash
echo "copy this" | tofu clip         # Copy to clipboard
tofu clip                            # Paste from clipboard
```

**Random generation**:
```bash
tofu uuid                            # Generate UUID
tofu rand password 16                # Random password
tofu rand passphrase 4               # Memorable passphrase
```

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

## All Commands

Browse the sidebar for the full command reference, organized by category:

- **File Operations** - cat, grep, find, ls, tree, diff, etc.
- **Network** - http, serve, ping, dns, wget, etc.
- **System** - ps, free, watch, env, etc.
- **Encoding** - base64, jwt, hash, qr, uuid
- **DevOps** - k8s, gh, git
- **Fun** - bird, cowsay, fortune, and more

---

## Features

- Works on Windows, macOS, and Linux
- Colored output
- Shell completions (bash, zsh, fish, PowerShell)
- Smart defaults
- Fast Go performance
