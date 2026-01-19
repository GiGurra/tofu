# watch

Watch files and execute a command on change.

## Synopsis

```bash
tofu watch -e <command> [directories...] [flags]
```

## Description

Monitor files for changes and automatically execute a command when changes are detected. Supports recursive watching, file pattern filtering, and automatic process restart.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--execute` | `-e` | Command to execute when files change | (required) |
| `--patterns` | `-p` | File patterns to watch (watches all if not specified) | |
| `--pattern-type` | | Pattern type: `regex`, `literal`, `glob` | `glob` |
| `--recursive` | `-r` | Watch directories recursively | `true` |
| `--include-hidden` | | Include hidden files and directories | `false` |
| `--exclude` | | Patterns to exclude (glob style) | |
| `--previous-process` | | Action for previous process: `kill`, `wait` | `kill` |
| `--handle-shutdown` | | Action when process exits: `restart`, `ignore` | `ignore` |
| `--restart-policy` | | Restart policy: `exponential-backoff` | `exponential-backoff` |
| `--min-backoff-millis` | | Minimum backoff in milliseconds | `1000` |
| `--max-backoff-millis` | | Maximum backoff in milliseconds | `10000` |
| `--max-restarts` | | Maximum automatic restarts | `10` |

## Examples

Watch current directory and run tests:

```bash
tofu watch -e "go test ./..."
```

Watch specific directory:

```bash
tofu watch -e "npm run build" ./src
```

Watch only Go files:

```bash
tofu watch -e "go build" -p "*.go"
```

Watch with multiple patterns:

```bash
tofu watch -e "make" -p "*.c" -p "*.h"
```

Exclude directories:

```bash
tofu watch -e "npm test" --exclude "node_modules" --exclude ".git"
```

Use regex patterns:

```bash
tofu watch -e "pytest" --pattern-type regex -p ".*_test\.py$"
```

Auto-restart on crash:

```bash
tofu watch -e "./myserver" --handle-shutdown restart
```

Wait for previous process before restart:

```bash
tofu watch -e "./build.sh" --previous-process wait
```

Include hidden files:

```bash
tofu watch -e "make" --include-hidden
```

## Sample Output

```
Watching 15 directories...
Running: go test ./...

File change detected: main.go
Running: go test ./...
PASS
ok      mypackage    0.342s
```
