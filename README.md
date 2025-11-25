# �豆腐 tofu

[![CI Status](https://github.com/GiGurra/tofu/actions/workflows/ci.yml/badge.svg)](https://github.com/GiGurra/tofu/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/GiGurra/tofu)](https://goreportcard.com/report/github.com/GiGurra/tofu)

> Tools Of Fantastic Utility - the universal cross-platform protein for your command line

**tofu** is a modern toolkit of CLI utilities that work seamlessly across Windows, macOS, and Linux. Like its namesake,
tofu adapts to any environment while providing essential nutrition for your daily command-line workflows.

⚠️ WARNING ⚠️ : This is just a satire coding hack vibe coded into even more stupidity. But at least now, I can use find
and grep in powershell :P.

⚠️ WARNING ⚠️ : Large parts of this project, including this readme, are likely written by some LLM.
Expect weirdness and occasional poetry.

## Philosophy

Traditional Unix tools are powerful but often:

- Have cryptic flags and syntax
- Don't work consistently across platforms
- Lack modern CLI niceties (autocomplete, colored output, progress bars)
- Have interfaces designed in the 1970s

**tofu** reimagines these essential tools with:

- ✨ Modern, intuitive interfaces
- 🎯 Smart defaults that just work
- 🌈 Beautiful, colored output (maybe, easy there claude sonnet, maybe in the future)
- ⚡ Blazing fast performance (written in Go)
- 🔄 True cross-platform support (Windows, macOS, Linux)
- 📝 Built-in autocomplete and help
- 🎨 Consistent UX across all tools

## Tools

### tofu cat

Regular-ish `cat`, but works in powershell without making you want to scream.

See [man/man_cat.md](man/man_cat.md) for details.

### tofu clip

Clipboard copy and paste utility.

See [man/man_clip.md](man/man_clip.md) for details.

### tofu find

A modern file finder that makes `find` feel like it's from this century.

See [man/man_find.md](man/man_find.md) for details.

### tofu grep

Like gnu grep, but shittier - although, it works in powershell.

See [man/man_grep.md](man/man_grep.md) for details.

### tofu sed2

Like gnu sed, but at the same time not. This one has an actually useful syntax.

See [man/man_sed2.md](man/man_sed2.md) for details.

### tofu watch

Watches files for changes and executes commands. Like `watch` or `nodemon` but generic and simple.

See [man/man_watch.md](man/man_watch.md) for details.

### tofu serve

Instant static file server. Perfect for previewing sites or sharing files.

See [man/man_serve.md](man/man_serve.md) for details.

### tofu uuid

Generates Universally Unique Identifiers (UUIDs) with support for v1, v3, v4, v5, v6, and v7.

See [man/man_uuid.md](man/man_uuid.md) for details.

### tofu port

List or kill processes listening on network ports. Cross-platform support for Linux, Windows, and macOS.

See [man/man_port.md](man/man_port.md) for details.

### tofu tree

List contents of directories in a tree-like format, with options for depth and hidden files, and respecting `.gitignore`.

See [man/man_tree.md](man/man_tree.md) for details.

### tofu which

Locates executable files associated with a given command.

See [man/man_which.md](man/man_which.md) for details.

### tofu ip

Show local and public IP addresses, DNS servers, and default gateway.

See [man/man_ip.md](man/man_ip.md) for details.

### tofu dns

Lookup DNS records (A, AAAA, MX, etc.) using custom or OS resolvers.

See [man/man_dns.md](man/man_dns.md) for details.

### tofu http

Human-friendly HTTP client (curl/httpie alternative).

See [man/man_http.md](man/man_http.md) for details.

### tofu nc

Netcat clone: Connect to or listen on sockets (TCP/UDP).

See [man/man_nc.md](man/man_nc.md) for details.

### tofu qr

Render QR codes in the terminal.

See [man/man_qr.md](man/man_qr.md) for details.

### tofu base64

Base64 encode/decode files or stdin.

See [man/man_base64.md](man/man_base64.md) for details.

### tofu jwt

Decode and inspect JWT tokens (Header, Claims/Payload).

See [man/man_jwt.md](man/man_jwt.md) for details.

### tofu hash

Calculate cryptographic hashes (MD5, SHA1, SHA256, SHA512) for files or stdin.

See [man/man_hash.md](man/man_hash.md) for details.

### tofu free

Display amount of free and used memory in the system.

See [man/man_free.md](man/man_free.md) for details.

### tofu ps

Report a snapshot of the current processes with filtering options.

See [man/man_ps.md](man/man_ps.md) for details.

### tofu time

Show current time in various formats or parse a provided timestamp, with optional input format specification.

See [man/man_time.md](man/man_time.md) for details.

### tofu rand

Generate random strings, integers, hex, base64, passwords, and passphrases.

See [man/man_rand.md](man/man_rand.md) for details.

### tofu env

Cross-platform environment variable management. List, get, set, filter, and export environment variables.

See [man/man_env.md](man/man_env.md) for details.

### tofu cron

Explain and validate cron expressions. Parses standard 5-field and extended 6-field cron expressions with human-readable output.

See [man/man_cron.md](man/man_cron.md) for details.

**Coming soon:**

- `tofu json` - Simple JSON formatter and query tool (jq lite)
- `tofu url` - URL encoder/decoder and parser
- `tofu archive` - Create or extract common archive formats (zip, tar, gz)
- `tofu semver` - Semantic versioning utilities (bump, sort)
- `tofu count` - Line, word, and character counter
- `tofu diff` - Modern file diff with color output
- `tofu tail` - Real-time log following (tail -f replacement)
- `tofu yaml` - YAML validation and conversion
- ...and more based on what we're all tired of fighting with

## Installation

```shell
# Coming soon
go install github.com/gigurra/tofu@latest
```

### Shell completion

Shell completion is auto generated and available for:

- bash
- zsh
- fish
- powershell

see:

```shell
tofu completions <shell> [--help]
```

## Why "tofu"?

Like tofu, these tools are:

- **Universal** - Works in any environment (platform)
- **Adaptable** - Fits into any workflow
- **Essential** - The protein your CLI diet needs
- **Modern** - A contemporary take on traditional ingredients

(And perhaps, in a very specific light, almost as divinely inspired as TempleOS. Almost.)

Plus, it's fun to say you're "serving fresh tofu" when you ship a new tool. 🍴

## Development Status

🚧 **Early Development** - a.k.a. who the heck knows.

## Contributing

Ideas, issues, and PRs welcome! Let's make CLI tools that spark joy.

## License

MIT

## Original author(s?)

Satire Coders Collective

---

*Silky smooth commands, firm reliable results.* 🥢