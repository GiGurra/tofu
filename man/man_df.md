# `tofu df`

Report filesystem disk space usage.

## Synopsis

```
tofu df [paths] [flags]
```

## Description

Report filesystem disk space usage, like the Unix df command but cross-platform. Shows total, used, and available space for mounted filesystems.

## Options

- `-a, --all`: Include all filesystems, including pseudo filesystems
- `-h, --human`: Print sizes in human readable format (e.g., 1K 234M 2G)
- `-i, --inode`: List inode information instead of block usage
- `-l, --local`: Limit listing to local filesystems
- `-t, --type string`: Limit listing to filesystems of a specific type
- `-S, --sort string`: Sort by: 'used', 'available', 'percent', or 'name' (default)
- `-r, --reverse`: Reverse the sort order

## Examples

Show disk space for all mounted filesystems:

```
tofu df
```

Human-readable output:

```
tofu df -h
```

Show specific path:

```
tofu df /home
```

Sort by usage percentage:

```
tofu df -S percent
```

Show only local filesystems:

```
tofu df -l
```
