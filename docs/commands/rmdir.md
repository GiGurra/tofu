# rmdir

Remove empty directories.

## Synopsis

```bash
tofu rmdir <directories...> [flags]
```

## Description

Remove the specified directories if they are empty.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--parents` | `-p` | Remove directory and its ancestors | `false` |
| `--verbose` | `-v` | Output diagnostic for every directory processed | `false` |

## Examples

Remove an empty directory:

```bash
tofu rmdir empty_dir
```

Remove multiple directories:

```bash
tofu rmdir dir1 dir2 dir3
```

Remove directory and parent directories:

```bash
tofu rmdir -p path/to/empty/dir
# Removes: dir, empty, to, path (if all are empty)
```

Verbose output:

```bash
tofu rmdir -v mydir
# Output: rmdir: removing directory, 'mydir'
```
