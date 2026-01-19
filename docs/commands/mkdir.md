# mkdir

Create directories.

## Synopsis

```bash
tofu mkdir <directories...> [flags]
```

## Description

Create the specified directories if they do not already exist.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--parents` | `-p` | Make parent directories as needed, no error if existing | `false` |
| `--mode` | `-m` | Set file mode (octal) | `0755` |
| `--verbose` | `-v` | Print message for each created directory | `false` |

## Examples

Create a directory:

```bash
tofu mkdir mydir
```

Create multiple directories:

```bash
tofu mkdir dir1 dir2 dir3
```

Create nested directories:

```bash
tofu mkdir -p path/to/nested/directory
```

Create with specific permissions:

```bash
tofu mkdir -m 0700 private_dir
```

Verbose output:

```bash
tofu mkdir -v newdir
# Output: mkdir: created directory 'newdir'
```
