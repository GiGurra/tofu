# `tofu env`

Cross-platform environment variable management. List, get, set, or filter environment variables with consistent behavior across Windows, macOS, and Linux.

## Interface

```
> tofu env --help
Cross-platform environment variable management

Usage:
  tofu env [flags] [-- command [args...]]

Flags:
  -f, --format string   Output format (plain, json, shell, powershell). (default "plain")
      --filter string   Filter variables by prefix (case-insensitive).
  -s, --sort            Sort variables alphabetically. (default true)
  -k, --keys            Show only variable names (keys).
  -v, --values          Show only variable values.
  -g, --get string      Get a specific environment variable.
      --set string      Set an environment variable (format: KEY=VALUE) and run command.
  -u, --unset string    Unset an environment variable and run command.
  -e, --export          Output in export format for shell sourcing.
      --no-empty        Hide variables with empty values.
  -h, --help            help for env
```

## Features

- **List**: Display all environment variables
- **Get**: Retrieve a specific variable's value
- **Set**: Set a variable and optionally run a command with it
- **Unset**: Remove a variable and optionally run a command without it
- **Filter**: Show only variables matching a prefix
- **Format**: Output in plain, JSON, shell export, or PowerShell format

## Examples

List all environment variables (sorted):

```
> tofu env
HOME=/home/user
PATH=/usr/bin:/bin
...
```

Get a specific variable:

```
> tofu env -g HOME
/home/user
```

Filter variables by prefix:

```
> tofu env --filter PATH
PATH=/usr/bin:/bin
```

Show only keys:

```
> tofu env -k --filter USER
USER
USERNAME
```

Output as JSON:

```
> tofu env -f json --filter HOME
{
  "HOME": "/home/user"
}
```

Output for shell sourcing (bash/zsh):

```
> tofu env -f shell --filter MY_
export MY_VAR='value'
export MY_OTHER='another'
```

Output for PowerShell:

```
> tofu env -f powershell --filter MY_
$env:MY_VAR = 'value'
$env:MY_OTHER = 'another'
```

Auto-detect export format based on OS:

```
> tofu env -e --filter MY_
# On Linux/macOS: export MY_VAR='value'
# On Windows: $env:MY_VAR = 'value'
```

Set a variable and run a command:

```
> tofu env --set DEBUG=true -- ./my-program
```

Unset a variable and run a command:

```
> tofu env -u SECRET_KEY -- ./my-program
```

Hide empty variables:

```
> tofu env --no-empty
```

## Use Cases

### Debugging environment issues

```
> tofu env --filter PATH
> tofu env -g LD_LIBRARY_PATH
```

### Exporting for scripts

```
# Save current env for later
> tofu env -f shell > env_backup.sh

# Restore
> source env_backup.sh
```

### Running commands with modified environment

```
# Run with a specific variable set
> tofu env --set NODE_ENV=production -- npm start

# Run without a specific variable
> tofu env -u DEBUG -- ./release-build
```

### Cross-platform scripting

```
# Works the same on Windows, macOS, and Linux
> tofu env -g HOME
> tofu env --filter USER
```
