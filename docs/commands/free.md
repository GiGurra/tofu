# free

Display amount of free and used memory.

## Synopsis

```bash
tofu free [flags]
```

## Description

Display the total, used, and free amount of physical and swap memory in the system.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--megabytes` | `-m` | Display output in megabytes | `false` |
| `--gigabytes` | `-g` | Display output in gigabytes | `false` |

## Examples

Display memory in kilobytes (default):

```bash
tofu free
```

Display memory in megabytes:

```bash
tofu free -m
```

Display memory in gigabytes:

```bash
tofu free -g
```

## Sample Output

```
              total        used        free      shared  buff/cache   available
Mem:       16384000     8192000     4096000       51200     4096000     7680000
Swap:       4194304      524288     3670016
```
