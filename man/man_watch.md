# `tofu watch`

Watch files and execute a command on change.

## Interface

```
> tofu watch --help
Watch files and execute a command on change

Usage:
  tofu watch [dirs]... [flags]

Flags:
  -e, --execute string            Command to execute when files change. (Required)
  -p, --patterns strings          File patterns to watch (optional, watches all files if not specified).
      --pattern-type string       Type of pattern matching (regex, literal, glob). (default "glob")
  -r, --recursive                 Watch directories recursively. (default true)
      --include-hidden            Include hidden files and directories. (default false)
      --exclude strings           Patterns to exclude (glob style).
      --previous-process string   Action for previous process (kill, wait). (default "kill")
      --handle-shutdown string    Action when process exits (restart, ignore). (default "ignore")
      --restart-policy string     Restart policy (exponential-backoff). (default "exponential-backoff")
      --min-backoff-millis int    Minimum backoff duration in milliseconds. (default 1000)
      --max-backoff-millis int    Maximum backoff duration in milliseconds. (default 10000)
      --max-restarts int          Maximum number of automatic restarts. (default 10)
  -h, --help                      help for watch
```

### Examples

Watch the current directory and run tests on change:

```
> tofu watch -e "go test ./..."
```

Watch specific source files and rebuild:

```
> tofu watch -p "*.go" -p "*.mod" -e "go build"
```

Watch a specific directory, excluding `tmp` and `.git`:

```
> tofu watch ./src --exclude "*/tmp/*" --exclude "*/.git/*" -e "npm start"
```

Watch with regex pattern:

```
> tofu watch --pattern-type regex -p ".*_test\.go$" -e "echo 'Test file changed'"
```

Wait for the previous process to finish before starting a new one (instead of killing it):

```
> tofu watch -e "long-running-script.sh" --previous-process wait
```

Restart the process if it crashes:

```
> tofu watch -e "my-server" --handle-shutdown restart
```
