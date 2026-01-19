# ls / ll / la

List directory contents.

## Synopsis

```bash
tofu ls [paths...] [flags]
tofu ll [paths...] [flags]   # Alias for ls -lh
tofu la [paths...] [flags]   # Alias for ls -lah
```

## Description

List information about files and directories. By default, lists the current directory. The `ll` command is an alias for `ls -lh` (long format with human-readable sizes), and `la` is an alias for `ls -lah` (long format including hidden files).

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--all` | `-a` | Do not ignore entries starting with `.` | `false` |
| `--almost-all` | `-A` | Do not list implied `.` and `..` | `false` |
| `--long` | `-l` | Use long listing format | `false` |
| `--human-readable` | `-h` | Print sizes like 1K, 234M, 2G | `false` |
| `--one-per-line` | `-1` | List one file per line | `false` |
| `--reverse` | `-r` | Reverse sort order | `false` |
| `--sort-by-time` | `-t` | Sort by time, newest first | `false` |
| `--sort-by-size` | `-S` | Sort by size, largest first | `false` |
| `--no-sort` | `-U` | Do not sort; list in directory order | `false` |
| `--classify` | `-F` | Append indicator (*/=>@\|) to entries | `false` |
| `--directory` | `-d` | List directories themselves, not contents | `false` |
| `--recursive` | `-R` | List subdirectories recursively | `false` |
| `--inode` | `-i` | Print inode number of each file | `false` |
| `--size` | `-s` | Print allocated size in blocks | `false` |
| `--color` | | Colorize output: `always`, `auto`, `never` | `auto` |
| `--group-dirs-first` | | Group directories before files | `false` |
| `--no-group` | `-G` | Don't print group names in long listing | `false` |
| `--numeric-uid-gid` | `-n` | List numeric user and group IDs | `false` |

## Examples

List current directory:

```bash
tofu ls
```

Long format with human-readable sizes:

```bash
tofu ls -lh
# Or use the alias:
tofu ll
```

Show all files including hidden:

```bash
tofu ls -la
# Or use the alias:
tofu la
```

Sort by modification time:

```bash
tofu ls -lt
```

Sort by size (largest first):

```bash
tofu ls -lS
```

Recursive listing:

```bash
tofu ls -R
```

Group directories first:

```bash
tofu ls --group-dirs-first
```

List specific files/directories:

```bash
tofu ls file1.txt dir1/ dir2/
```

Show file type indicators:

```bash
tofu ls -F
```
