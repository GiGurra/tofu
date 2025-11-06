# �豆腐 tofu

> Tools Of Fantastic Utility - the universal cross-platform protein for your command line

**tofu** is a modern toolkit of CLI utilities that work seamlessly across Windows, macOS, and Linux. Like its namesake,
tofu adapts to any environment while providing essential nutrition for your daily command-line workflows.

⚠️ WARNING ⚠️ : This readme was written mostly by claude sonnet 4.5. Expect weirdness and occasional poetry.

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

### tofu find

A modern file finder that makes `find` feel like it's from this century.
See [man/man_find.md](man/man_find.md) for details.

**Coming soon:**

- `tofu grep` - search that doesn't make you look up regex every time
- `tofu watch` - file watching done right
- `tofu serve` - instant static file server
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

Plus, it's fun to say you're "serving fresh tofu" when you ship a new tool. 🍴

## Development Status

🚧 **Early Development** - Currently building the first tool (`tofu find`). Watch this space!

## Contributing

Ideas, issues, and PRs welcome! Let's make CLI tools that spark joy.

## License

MIT

---

*Silky smooth commands, firm reliable results.* 🥢