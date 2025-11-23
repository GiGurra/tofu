# `tofu ps`

Report a snapshot of the current processes.

## Interface

```
> tofu ps --help
Displays information about a selection of the active processes.
By default, it lists all processes with a minimal set of columns.
Use -f for a full format listing.
Filters can be combined (AND logic). Use -v to invert the filter.

Usage:
  tofu ps [flags]

Flags:
  -c, --current       Show only processes owned by the current user.
  -f, --full          Display full format listing.
  -h, --help          help for ps
  -v, --invert        Invert filtering (matches non-matching processes).
  -n, --name string   Filter by command name (substring).
  -p, --pids ints     Filter by PID(s).
  -u, --users strings Filter by username(s).
```

## Description

The `ps` command displays a list of currently running processes. By default, it shows the PID and the command name. The `--full` (or `-f`) flag adds more detailed columns including PPID, User, Status, CPU usage, Memory usage, and the full command line.

You can filter the list of processes by:
- **PID**: Using `-p` or `--pids` (comma-separated list).
- **User**: Using `-u` or `--users` (comma-separated list).
- **Current User**: Using `-c` or `--current` to show only your own processes.
- **Name**: Using `-n` or `--name` to match a substring in the command name.

All filters are combined using AND logic (a process must match all specified criteria). The `-v` or `--invert` flag inverts the final result (shows processes that *do not* match the criteria).

## Examples

Display basic process list:

```
> tofu ps
PID   COMMAND
1     systemd
2     kthreadd
...
```

Display full process details for the current user:

```
> tofu ps -f -c
PID   PPID   USER     STATUS   %CPU   %MEM   COMMAND
1000  1      gigur    S        0.1    0.5    /usr/bin/zsh
...
```

Find processes by name (e.g., "chrome"):

```
> tofu ps -n chrome
PID   COMMAND
1234  chrome
1235  chrome
...
```

Find specific PIDs:

```
> tofu ps -p 1234,5678
PID   COMMAND
1234  chrome
5678  slack
```

Exclude processes owned by "root":

```
> tofu ps -u root -v
PID   COMMAND
1000  zsh
...
```