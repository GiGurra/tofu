# cp

Copy files and directories.

## Synopsis

```bash
tofu cp <source...> <destination> [flags]
```

## Description

Copy SOURCE to DEST, or multiple SOURCE(s) to DIRECTORY.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--recursive` | `-r` | Copy directories recursively | `false` |
| `--force` | `-f` | Remove destination file if cannot open, then retry | `false` |
| `--interactive` | `-i` | Prompt before overwriting | `false` |
| `--no-clobber` | `-n` | Do not overwrite an existing file | `false` |
| `--verbose` | `-v` | Explain what is being done | `false` |
| `--preserve` | `-p` | Preserve mode, ownership, and timestamps | `false` |

## Examples

Copy a file:

```bash
tofu cp source.txt dest.txt
```

Copy file to directory:

```bash
tofu cp file.txt /path/to/directory/
```

Copy multiple files:

```bash
tofu cp file1.txt file2.txt /destination/
```

Copy directory recursively:

```bash
tofu cp -r source_dir/ dest_dir/
```

Preserve file attributes:

```bash
tofu cp -p important.txt backup.txt
```

Interactive mode:

```bash
tofu cp -i source.txt dest.txt
```

Verbose output:

```bash
tofu cp -v file.txt backup/
# Output: 'file.txt' -> 'backup/file.txt'
```
