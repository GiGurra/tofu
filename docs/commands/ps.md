# ps

Report a snapshot of current processes.

## Synopsis

```bash
tofu ps [flags]
```

## Description

Display information about active processes. By default, lists all processes with minimal columns. Use filters to narrow results.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--full` | `-f` | Display full format listing | `false` |
| `--users` | `-u` | Filter by username(s) | |
| `--pids` | `-p` | Filter by PID(s) | |
| `--name` | `-n` | Filter by command name (substring) | |
| `--current` | `-c` | Show only processes owned by current user | `false` |
| `--invert` | `-v` | Invert filtering (show non-matching) | `false` |
| `--no-truncate` | `-N` | Do not truncate command line output | `false` |

## Examples

List all processes:

```bash
tofu ps
```

Full format listing:

```bash
tofu ps -f
```

Filter by current user:

```bash
tofu ps -c
```

Filter by process name:

```bash
tofu ps -n node
```

Filter by specific PID:

```bash
tofu ps -p 1234
```

Filter by username:

```bash
tofu ps -u root
```

Show all except matching processes:

```bash
tofu ps -n chrome -v
```

Full output without truncation:

```bash
tofu ps -f -N
```

## Sample Output

Simple format:
```
PID      COMMAND
1        systemd
1234     node
5678     bash
```

Full format (`-f`):
```
PID      PPID     USER     STATUS   %CPU    %MEM    COMMAND
1        0        root     S        0.0     0.1     systemd
1234     5678     user     R        5.2     2.3     node server.js
5678     1        user     S        0.1     0.5     /bin/bash
```
