# 🥢 TOFU

[![CI Status](https://github.com/GiGurra/tofu/actions/workflows/ci.yml/badge.svg)](https://github.com/GiGurra/tofu/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/GiGurra/tofu)](https://goreportcard.com/report/github.com/GiGurra/tofu)
[![Docs](https://img.shields.io/badge/docs-gigurra.github.io%2Ftofu-blue)](https://gigurra.github.io/tofu/)

**[Read the serious documentation](https://gigurra.github.io/tofu/)**

---

**ARE YOU TIRED** of your command line tools not working the same way on Windows, Mac, and Linux?

**ARE YOU SICK** of cryptic flags that were designed when disco was still cool?

**DO YOU LIE AWAKE AT NIGHT** wondering why `grep` works differently in PowerShell?

## WELL LOOK NO FURTHER!

```shell
go install github.com/gigurra/tofu@latest
```

Introducing **TOFU** - *Tools Of Fantastic Utility!*

For the LOW LOW PRICE of **absolutely nothing**, you get not one, not two, but **OVER 50 COMMAND LINE TOOLS** that work EVERYWHERE!

---

## 🤖 Featured: Claude Code Session Management

**Manage your Claude Code sessions like a pro!**

![Claude Session Demo](docs/claude/demo.gif)

- **Interactive watch modes** - Browse sessions and conversations with search, filtering, and sorting
- **Tmux integration** - Run Claude in persistent sessions, attach/detach anytime
- **Status tracking** - See when Claude is working, idle, or waiting for input
- **Session indicators** - Know which conversations have active sessions (⚡/○)

```bash
# Start Claude in a tmux session
tofu claude session new

# Interactive session browser (with search!)
tofu claude session ls -w

# Interactive conversation browser
tofu claude conv ls -w

# Global search across all projects
tofu claude conv ls -g -w
```

**[Full Claude documentation →](docs/claude/README.md)**

---

### But wait - there's MORE!

You get `cat`! You get `grep`! You get `find`! You get `ls`!

**ALL WORKING IN POWERSHELL WITHOUT MAKING YOU WANT TO FLIP YOUR DESK!**

---

### But wait - there's EVEN MORE!

| Tool | What It Does | Wow Factor |
|------|--------------|------------|
| `cat` | Concatenates files | Works in PowerShell! |
| `grep` | Searches text | Actually readable output! |
| `find` | Finds files | No more `-name` confusion! |
| `ls`, `ll`, `la` | Lists directories | Colors! Sorting! Sanity! |
| `tree` | Shows directory trees | Respects .gitignore! |
| `sed2` | Stream editing | Syntax from THIS century! |
| `diff` | Compares files | With actual colors! |
| `head` / `tail` | Shows file parts | Cross-platform! |
| `cp`, `mv`, `rm` | File operations | The classics, everywhere! |
| `mkdir`, `rmdir`, `touch` | More file ops | You know what these do! |
| `watch` | Watches for changes | Like nodemon but better! |
| `serve` | HTTP file server | One command, done! |
| `http` | HTTP client | Like curl but readable! |
| `ping` | Network ping | Real ICMP packets! |
| `nc` | Netcat | TCP/UDP connections! |
| `dns` | DNS lookups | All record types! |
| `ip` | Network info | Local AND public IPs! |
| `port` | Port management | Find and kill by port! |
| `ps` | Process list | With filtering! |
| `free` | Memory info | Human readable! |
| `du`, `df` | Disk usage | Storage stats! |
| `uuid` | UUID generator | v1 through v7! |
| `rand` | Random generator | Strings, passwords, passphrases! |
| `hash` | Cryptographic hashes | MD5, SHA1, SHA256, SHA512! |
| `base64` | Base64 encode/decode | For your encoding needs! |
| `jwt` | JWT decoder | Inspect those tokens! |
| `qr` | QR code generator | In your terminal! |
| `time` | Time utilities | Parse and format! |
| `cron` | Cron explainer | Human readable schedules! |
| `env` | Environment vars | Cross-platform management! |
| `count` | Line/word counter | Modern wc replacement! |
| `clip` | Clipboard | Copy/paste from CLI! |
| `archive` | Archive tool | tar, zip, 7z, rar! |
| `wget` | File downloader | With progress bars! |
| `reverse` | Reverse lines | Because why not! |
| `tee` | Output splitter | To files AND stdout! |
| `sponge` | Soak up stdin | Release after EOF! |
| `which` | Locate programs | Find it in PATH! |
| `gh` | GitHub utilities | PRs, issues, repos! |
| `git` | Git utilities | Sync, open, and more! |
| `k8s` | Kubernetes utilities | Tail pod logs! |

---

### But wait - there's SILLY MORE!

**BECAUSE PRODUCTIVITY ISN'T EVERYTHING:**

| Tool | What It Does | Usefulness Level |
|------|--------------|------------------|
| `excuse` | Generate programmer excuses | "It works on my machine!" |
| `blame` | Randomly blame something | It's clearly cosmic rays! |
| `dice` | Roll dice (D&D style) | `2d20+5` = settled debate! |
| `coin` | Flip a coin | With animation! |
| `magic8` | Ask the Magic 8-Ball | For architecture decisions! |
| `pick` | Random selection | `ls \| tofu pick` |
| `uwu` | UwU-ify text | Pipe your errors through it! |
| `flip` | Flip text upside down | (╯°□°)╯︵ ┻━┻ |
| `lolcat` | Rainbow text | Makes everything better! |
| `morse` | Morse code | ... --- ... |
| `busy` | Fake progress bar | Look productive! |
| `lorem` | Lorem ipsum generator | For your placeholder needs! |
| `name` | Random project names | "Operation Thundering Gopher" |
| `pomodoro` | Pomodoro timer | 🍅 Get stuff done! |
| `standup` | Stand up reminders | Your spine will thank you! |
| `clock` | Analog terminal clock | Tick tock! |
| `bird` | Flappy Tofu game | Procrastinate in style! |
| `screensaver` | Animated tofu bowl | 🍜 Mesmerizing! |
| `fortune` | Tech fortune cookies | Programming wisdom! |
| `cowsay` | ASCII animal speech | Moo! |
| `figlet` | ASCII art banners | BIG TEXT! |
| `weather` | ASCII weather | Wraps wttr.in! |
| `calendar` | Terminal calendar | Today highlighted! |
| `stopwatch` | Stopwatch with laps | Space for lap! |
| `typing` | Typing speed test | How fast are you? |

---

### But wait - there's STILL MORE!

**ORDER NOW** and we'll throw in:

- 🌈 Colored output *(where applicable)*
- 📝 Shell completions for bash, zsh, fish, AND PowerShell
- ⚡ Blazing fast Go performance
- 🎯 Smart defaults that actually make sense
- 🤖 Partially written by LLMs *(for that authentic 2020s vibe)*

---

## Installation

```shell
go install github.com/gigurra/tofu@latest
```

That's it. That's the whole thing. **NO ASSEMBLY REQUIRED!**

---

## Documentation

*For the fine print and actual documentation, see the [official docs](https://gigurra.github.io/tofu/).*

*We have to put the real docs somewhere, and lawyers said we can't just yell at people forever.*

---

## FAQ

**Q: Is this a joke?**
A: The README is. The tools actually work.

**Q: Why is it called tofu?**
A: Because it adapts to any environment, just like tofu absorbs any flavor. Also we couldn't get the domain for "tools".

**Q: Does this require sudo?**
A: Only `ping`. We're not animals.

**Q: Who made this?**
A: The Satire Coders Collective, with generous assistance from our robot overlords.

---

## License

MIT - *Because sharing is caring!*

---

**CALL NOW!** *(Just kidding, just run `go install`)*

*Operators are standing by!* *(They're not, it's just a GitHub repo)*

*Silky smooth commands, firm reliable results.* 🥢
