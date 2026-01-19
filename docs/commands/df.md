# df

Report filesystem disk space usage.

## Synopsis

```bash
tofu df [paths...] [flags]
```

## Description

Report filesystem disk space usage. Similar to the Unix `df` command but cross-platform.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--all` | `-a` | Include pseudo filesystems | `false` |
| `--human` | `-h` | Print sizes in human readable format | `false` |
| `--inode` | `-i` | List inode information instead of block usage | `false` |
| `--local` | `-l` | Limit to local filesystems | `false` |
| `--type` | `-t` | Limit to filesystems of specific type | |
| `--sort` | `-S` | Sort by: `name`, `used`, `available`, `percent` | `name` |
| `--reverse` | `-r` | Reverse the sort order | `false` |

## Examples

Show disk space for all filesystems:

```bash
tofu df
```

Human-readable output:

```bash
tofu df -h
```

Show inode usage:

```bash
tofu df -i
```

Show only local filesystems:

```bash
tofu df -l
```

Filter by filesystem type:

```bash
tofu df -t ext4
```

Sort by percent used:

```bash
tofu df -S percent
```

Show space for specific path:

```bash
tofu df /home
```

Include all filesystems:

```bash
tofu df -a
```

## Sample Output

```
Filesystem                       1K-blocks        Used   Available Use%  Mounted on
-----------------------------------------------------------------------------------------------
/dev/sda1                        102400000    45678912    51234567  45%  /
/dev/sdb1                        512000000   123456789   300000000  29%  /home
```

Human-readable:

```
Filesystem                           Size     Used    Avail Use%  Mounted on
-----------------------------------------------------------------------------------------------
/dev/sda1                            97.7G    43.5G    48.9G  45%  /
/dev/sdb1                             488G     118G     286G  29%  /home
```
