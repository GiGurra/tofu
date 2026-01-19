# env

Cross-platform environment variable management.

## Synopsis

```bash
tofu env [flags] [command]
```

## Description

List, get, set, or filter environment variables. Works consistently across Windows, macOS, and Linux.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--format` | `-f` | Output format: `plain`, `json`, `shell`, `powershell` | `plain` |
| `--filter` | | Filter variables by prefix (case-insensitive) | |
| `--sort` | `-s` | Sort variables alphabetically | `true` |
| `--keys` | `-k` | Show only variable names | `false` |
| `--values` | `-v` | Show only variable values | `false` |
| `--get` | `-g` | Get a specific environment variable | |
| `--set` | | Set variable (KEY=VALUE) and run command | |
| `--unset` | `-u` | Unset variable and run command | |
| `--export` | `-e` | Output in export format for shell sourcing | `false` |
| `--no-empty` | | Hide variables with empty values | `false` |

## Examples

List all environment variables:

```bash
tofu env
```

Get a specific variable:

```bash
tofu env -g PATH
```

Filter by prefix:

```bash
tofu env --filter HOME
```

JSON output:

```bash
tofu env -f json
```

Shell export format:

```bash
tofu env -e
# Or explicitly:
tofu env -f shell
```

PowerShell format:

```bash
tofu env -f powershell
```

Set variable and run command:

```bash
tofu env --set "NODE_ENV=production" npm start
```

Unset variable and run command:

```bash
tofu env -u DEBUG npm start
```

Show only keys:

```bash
tofu env -k
```

Hide empty values:

```bash
tofu env --no-empty
```

## Sample Output

Plain format:
```
HOME=/home/user
PATH=/usr/bin:/bin
USER=johndoe
```

JSON format:
```json
{
  "HOME": "/home/user",
  "PATH": "/usr/bin:/bin",
  "USER": "johndoe"
}
```

Shell export format:
```bash
export HOME='/home/user'
export PATH='/usr/bin:/bin'
export USER='johndoe'
```
