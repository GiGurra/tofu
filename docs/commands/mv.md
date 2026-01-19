# mv

Move (rename) files.

## Synopsis

```bash
tofu mv <source...> <destination> [flags]
```

## Description

Rename SOURCE to DEST, or move SOURCE(s) to DIRECTORY.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--force` | `-f` | Do not prompt before overwriting | `false` |
| `--interactive` | `-i` | Prompt before overwriting | `false` |
| `--no-clobber` | `-n` | Do not overwrite an existing file | `false` |
| `--verbose` | `-v` | Explain what is being done | `false` |

## Examples

Rename a file:

```bash
tofu mv old.txt new.txt
```

Move file to directory:

```bash
tofu mv file.txt /path/to/directory/
```

Move multiple files to directory:

```bash
tofu mv file1.txt file2.txt file3.txt /destination/
```

Interactive mode (prompt before overwrite):

```bash
tofu mv -i source.txt dest.txt
```

Don't overwrite existing files:

```bash
tofu mv -n source.txt dest.txt
```

Verbose output:

```bash
tofu mv -v file.txt newdir/
# Output: renamed 'file.txt' -> 'newdir/file.txt'
```
