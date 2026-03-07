# TOFU

[![CI Status](https://github.com/GiGurra/tofu/actions/workflows/ci.yml/badge.svg)](https://github.com/GiGurra/tofu/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/GiGurra/tofu)](https://goreportcard.com/report/github.com/GiGurra/tofu)
[![Docs](https://img.shields.io/badge/docs-gigurra.github.io%2Ftofu-blue)](https://gigurra.github.io/tofu/)

This repo started out as a joke, where I gave claude code the instruction to "add more silly commands", repeatedly.
Now though, it does have some use with `tofu k8s|gh|git`.

50+ Unix-like commands that work consistently on Windows, macOS, and Linux.

> **Note:** The Claude Code extensions (`tclaude`) have moved to their own repo: **[tofutools/tclaude](https://github.com/tofutools/tclaude)**.
> The `tofu claude` subcommand and `cmd/tclaude` in this repo are deprecated and will be removed in a future release.
> Please install from the new repo: `go install github.com/tofutools/tclaude/cmd/tclaude@latest`

```bash
go install github.com/gigurra/tofu/cmd/tofu@latest
```

## Documentation

**[gigurra.github.io/tofu](https://gigurra.github.io/tofu/)**

---

## License

MIT
