# du

Estimate file and directory space usage.

## Synopsis

```bash
tofu du [paths...] [flags]
```

## Description

Estimate file and directory space usage. Similar to the Unix `du` command but cross-platform.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--summarize` | `-s` | Display only total for each path | `false` |
| `--all` | `-a` | Write counts for all files, not just directories | `false` |
| `--human` | `-h` | Print sizes in human readable format | `false` |
| `--max-depth` | `-d` | Print total only if N or fewer levels deep | `-1` |
| `--bytes` | `-b` | Apparent size in bytes | `false` |
| `--apparent-size` | | Print apparent sizes rather than disk usage | `false` |
| `--kilobytes` | `-k` | Print in kilobytes | `false` |
| `--sort` | `-S` | Sort by: `size`, `name`, `none` | `size` |
| `--reverse` | `-r` | Reverse the sort order | `false` |
| `--ignore-git` | | Respect .gitignore files | `false` |

## Examples

Show disk usage of current directory:

```bash
tofu du
```

Human-readable output:

```bash
tofu du -h
```

Summary only (total):

```bash
tofu du -s
```

Limit depth:

```bash
tofu du -d 1
```

Show all files (not just directories):

```bash
tofu du -a
```

Sort by size (largest last):

```bash
tofu du -S size
```

Sort by name:

```bash
tofu du -S name
```

Fastest mode (no sorting, stream output):

```bash
tofu du -S none
```

Show apparent size in bytes:

```bash
tofu du -b
```

Specific paths:

```bash
tofu du -h /path/to/dir1 /path/to/dir2
```

## Sample Output

```
4K      ./cmd/cat
8K      ./cmd/grep
12K     ./cmd
16K     .
```
