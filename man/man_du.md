# `tofu du`

Estimate file and directory space usage.

## Synopsis

```
tofu du [paths] [flags]
```

## Description

Estimate file and directory space usage, like the Unix du command but cross-platform. Summarize disk usage of each FILE, recursively for directories.

## Options

- `-s, --summarize`: Display only the total for each path
- `-a, --all`: Write counts for all files, not just directories
- `-h, --human`: Print sizes in human readable format (B, KB, MB, GB, etc.)
- `-d, --max-depth int`: Print the total for a directory only if it is N or fewer levels deep
- `-b, --bytes`: Apparent size in bytes (equivalent to --apparent-size --block-size=1)
- `--apparent-size`: Print apparent sizes rather than disk usage
- `-k, --kilobytes`: Print in kilobytes
- `-S, --sort string`: Sort by: 'size' (largest last), 'name', or 'none' (fastest, streams output)
- `-r, --reverse`: Reverse the sort order
- `-i, --ignore-git`: Respect .gitignore files

## Examples

Show disk usage for current directory:

```
tofu du
```

Human-readable summary:

```
tofu du -sh .
```

Show sizes up to 2 levels deep:

```
tofu du -h -d 2
```

Show all files and directories:

```
tofu du -ah
```

Sort by size (largest first):

```
tofu du -h -S size -r
```

Respect .gitignore:

```
tofu du -h -i
```
