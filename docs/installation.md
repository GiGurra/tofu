# Installation

## Using Go Install

The simplest way to install tofu:

```bash
go install github.com/gigurra/tofu/cmd/tofu@latest
go install github.com/gigurra/tofu/cmd/tclaude@latest
```

This requires Go 1.21 or later.

## From Source

```bash
git clone https://github.com/GiGurra/tofu.git
cd tofu
go build .
```

### Linux Dependencies

On Linux, the `bird` command requires audio libraries:

```bash
sudo apt-get install -y pkg-config libasound2-dev
```

## Shell Completions

Generate shell completions using the built-in completion command:

=== "Bash"

    ```bash
    tofu completion bash > /etc/bash_completion.d/tofu
    ```

=== "Zsh"

    ```bash
    tofu completion zsh > "${fpath[1]}/_tofu"
    ```

=== "Fish"

    ```bash
    tofu completion fish > ~/.config/fish/completions/tofu.fish
    ```

=== "PowerShell"

    ```powershell
    tofu completion powershell | Out-String | Invoke-Expression
    ```

## Verify Installation

```bash
tofu --version
tofu --help
```
