# `tofu ls`

List directory contents.

## Synopsis

```
tofu ls [paths] [flags]
```

## Description

List information about the FILEs (the current directory by default). Sort entries alphabetically unless other sort options are specified.

## Aliases

- `tofu ll`: Alias for `tofu ls -l` (long format)
- `tofu la`: Alias for `tofu ls -la` (long format, all files)

## Options

- `-a, --all`: Do not ignore entries starting with .
- `-A, --almost-all`: Do not list implied . and ..
- `-l, --long`: Use a long listing format
- `-h, --human-readable`: With -l, print sizes like 1K 234M 2G etc.
- `-1, --one-per-line`: List one file per line
- `-r, --reverse`: Reverse order while sorting
- `-t, --sort-by-time`: Sort by time, newest first
- `-S, --sort-by-size`: Sort by file size, largest first
- `-U, --no-sort`: Do not sort; list entries in directory order
- `-F, --classify`: Append indicator (one of */=>@|) to entries
- `-d, --directory`: List directories themselves, not their contents
- `-R, --recursive`: List subdirectories recursively
- `-i, --inode`: Print the index number of each file
- `-s, --size`: Print the allocated size of each file, in blocks
- `-c, --color string`: Colorize the output: 'always', 'auto', or 'never' (default "auto")
- `-g, --group-dirs-first`: Group directories before files
- `-G, --no-group`: In a long listing, don't print group names
- `-n, --numeric-uid-gid`: Like -l, but list numeric user and group IDs
- `-f, --full-group`: Show full group identifier (e.g., Windows SID)

## Examples

List current directory:

```
tofu ls
```

Long format with human-readable sizes:

```
tofu ls -lh
```

List all files including hidden:

```
tofu ls -a
```

List sorted by modification time:

```
tofu ls -lt
```

Recursive listing:

```
tofu ls -R
```

Group directories first:

```
tofu ls -g
```
