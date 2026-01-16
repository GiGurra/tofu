# tofu

Cross-platform CLI tools written in Go. Provides 50+ Unix-like commands that work consistently on Windows, macOS, and Linux.

## Development

```bash
# Build
go build .

# Run from source
go run . --help

# Run tests
go test ./...

# Linux: requires libasound2-dev for audio (bird command)
sudo apt-get install -y pkg-config libasound2-dev
```

## Code Structure

```
main.go              # Entry point, registers all subcommands
cmd/
  <command>/         # Each command in its own package
    <command>.go     # Exports Cmd() returning *cobra.Command
    <command>_test.go
    <command>_linux.go    # Platform-specific (optional)
    <command>_darwin.go
    <command>_windows.go
  common/            # Shared utilities
man/                 # Documentation
```

## Adding a New Command

1. Create `cmd/<name>/<name>.go`
2. Define a `Params` struct with CLI flags/args using struct tags
3. Export a `Cmd()` function returning `*cobra.Command` using `boa.CmdT`
4. Register in `main.go` by importing and adding to `SubCmds` slice

Example pattern (see `cmd/cat/cat.go`):
```go
type Params struct {
    Files []string `pos:"true" optional:"true" help:"..." default:"-"`
    Flag  bool     `short:"f" help:"..."`
}

func Cmd() *cobra.Command {
    return boa.CmdT[Params]{
        Use:     "name",
        Short:   "Description",
        RunFunc: func(params *Params, cmd *cobra.Command, args []string) { ... },
    }.ToCobra()
}
```

## Key Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/GiGurra/boa` - Typed params wrapper for cobra

## Feature Backlog

See `TODO.md` for the list of planned features.

**Important**: Always read `TODO.md` at the start of a session. When implementing a new feature from the list, mark it as complete (`[x]`) in TODO.md after the implementation is done and tested.
