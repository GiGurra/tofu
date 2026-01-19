# TOFU - Tools Of Fantastic Utility

[![CI Status](https://github.com/GiGurra/tofu/actions/workflows/ci.yml/badge.svg)](https://github.com/GiGurra/tofu/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/GiGurra/tofu)](https://goreportcard.com/report/github.com/GiGurra/tofu)

Cross-platform CLI tools written in Go. Over 50 Unix-like commands that work consistently on Windows, macOS, and Linux.

## Quick Start

```bash
go install github.com/gigurra/tofu@latest
```

## Commands Overview

### File Operations

| Command | Description |
|---------|-------------|
| [cat](commands/cat.md) | Concatenate and display files |
| [find](commands/find.md) | Find files by name or pattern |
| [grep](commands/grep.md) | Search text using patterns |
| [sed2](commands/sed2.md) | Stream editor with modern syntax |
| [tree](commands/tree.md) | Display directory tree |
| [ls / ll / la](commands/ls.md) | List directory contents |
| [mkdir](commands/mkdir.md) | Create directories |
| [mv](commands/mv.md) | Move or rename files |
| [cp](commands/cp.md) | Copy files and directories |
| [rm](commands/rm.md) | Remove files |
| [rmdir](commands/rmdir.md) | Remove empty directories |
| [touch](commands/touch.md) | Create or update file timestamps |
| [head](commands/head.md) | Display beginning of files |
| [tail](commands/tail.md) | Display end of files |
| [diff](commands/diff.md) | Compare files |
| [du](commands/du.md) | Disk usage |
| [df](commands/df.md) | Disk free space |

### Network

| Command | Description |
|---------|-------------|
| [serve](commands/serve.md) | HTTP file server |
| [port](commands/port.md) | Port management |
| [ip](commands/ip.md) | Network interface info |
| [dns](commands/dns.md) | DNS lookups |
| [http](commands/http.md) | HTTP client |
| [nc](commands/nc.md) | Netcat (TCP/UDP) |
| [ping](commands/ping.md) | Network ping |
| [wget](commands/wget.md) | File downloader |
| [weather](commands/weather.md) | ASCII weather display |

### Encoding & Crypto

| Command | Description |
|---------|-------------|
| [base64](commands/base64.md) | Base64 encode/decode |
| [jwt](commands/jwt.md) | JWT decoder |
| [hash](commands/hash.md) | Cryptographic hashes |
| [qr](commands/qr.md) | QR code generator |
| [uuid](commands/uuid.md) | UUID generator |
| [morse](commands/morse.md) | Morse code converter |

### System

| Command | Description |
|---------|-------------|
| [free](commands/free.md) | Memory info |
| [ps](commands/ps.md) | Process list |
| [time](commands/time.md) | Time utilities |
| [env](commands/env.md) | Environment variables |
| [watch](commands/watch.md) | Watch for file changes |
| [which](commands/which.md) | Locate programs |
| [clip](commands/clip.md) | Clipboard operations |
| [cron](commands/cron.md) | Cron schedule explainer |
| [archive](commands/archive.md) | Archive/compression tool |

### Text Processing

| Command | Description |
|---------|-------------|
| [count](commands/count.md) | Line/word/char counter |
| [reverse](commands/reverse.md) | Reverse lines |
| [tee](commands/tee.md) | Output splitter |
| [sponge](commands/sponge.md) | Buffer stdin |
| [lolcat](commands/lolcat.md) | Rainbow text |
| [uwu](commands/uwu.md) | UwU-ify text |
| [figlet](commands/figlet.md) | ASCII art banners |

### Randomness

| Command | Description |
|---------|-------------|
| [rand](commands/rand.md) | Random generator |
| [coin](commands/coin.md) | Coin flip |
| [dice](commands/dice.md) | Dice roller |
| [pick](commands/pick.md) | Random selection |
| [name](commands/name.md) | Random project names |
| [lorem](commands/lorem.md) | Lorem ipsum generator |

### Fun & Silly

| Command | Description |
|---------|-------------|
| [bird](commands/bird.md) | Flappy Tofu game |
| [blame](commands/blame.md) | Random blame generator |
| [excuse](commands/excuse.md) | Programmer excuses |
| [magic8](commands/magic8.md) | Magic 8-Ball |
| [flip](commands/flip.md) | Flip text upside down |
| [busy](commands/busy.md) | Fake progress bar |
| [fortune](commands/fortune.md) | Fortune cookies |
| [cowsay](commands/cowsay.md) | ASCII animal speech |
| [screensaver](commands/screensaver.md) | Animated tofu bowl |
| [typing](commands/typing.md) | Typing speed test |

### Productivity

| Command | Description |
|---------|-------------|
| [pomodoro](commands/pomodoro.md) | Pomodoro timer |
| [standup](commands/standup.md) | Stand-up reminders |
| [calendar](commands/calendar.md) | Terminal calendar |
| [clock](commands/clock.md) | Analog terminal clock |
| [stopwatch](commands/stopwatch.md) | Stopwatch with laps |

### DevOps

| Command | Description |
|---------|-------------|
| [k8s](commands/k8s.md) | Kubernetes utilities |
| [gh](commands/gh.md) | GitHub utilities |
| [git](commands/git.md) | Git utilities |
| [claude](commands/claude.md) | Claude AI integration |

## Features

- Works consistently on Windows, macOS, and Linux
- Colored output where applicable
- Shell completions for bash, zsh, fish, and PowerShell
- Smart defaults that make sense
- Fast Go performance
