# rm

Remove files or directories.

## Synopsis

```bash
tofu rm <files...> [flags]
```

## Description

Remove (unlink) the specified files or directories.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--recursive` | `-r` | Remove directories and contents recursively | `false` |
| `--force` | `-f` | Ignore nonexistent files, never prompt | `false` |
| `--interactive` | `-i` | Prompt before every removal | `false` |
| `--dir` | `-d` | Remove empty directories | `false` |
| `--verbose` | `-v` | Explain what is being done | `false` |

## Examples

Remove a file:

```bash
tofu rm file.txt
```

Remove multiple files:

```bash
tofu rm file1.txt file2.txt file3.txt
```

Remove directory recursively:

```bash
tofu rm -r directory/
```

Force remove (no prompts, ignore missing):

```bash
tofu rm -rf directory/
```

Interactive mode:

```bash
tofu rm -i file.txt
# Output: rm: remove 'file.txt'? y
```

Remove empty directory:

```bash
tofu rm -d empty_dir/
```

Verbose output:

```bash
tofu rm -v file.txt
# Output: removed 'file.txt'
```
