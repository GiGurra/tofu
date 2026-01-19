# sponge

Soak up stdin and release to stdout after EOF.

## Synopsis

```bash
tofu sponge [flags]
```

## Description

Read all input from stdin until EOF, storing it in memory (or a temporary file if too large). Only after all input is read does it output to stdout. Useful for in-place file modifications in pipelines.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--max-size` | `-m` | Maximum size in memory before buffering to temp file | `10m` |

## Examples

In-place file modification:

```bash
cat file.txt | grep "pattern" | tofu sponge > file.txt
```

Sort a file in place:

```bash
sort file.txt | tofu sponge > file.txt
```

With custom memory limit:

```bash
cat largefile.txt | tofu sponge -m 100m > output.txt
```

## Why Use Sponge?

Without sponge, this would fail:
```bash
# DON'T DO THIS - file gets truncated before reading!
cat file.txt | grep "pattern" > file.txt
```

With sponge:
```bash
# This works correctly
cat file.txt | grep "pattern" | tofu sponge > file.txt
```

## Size Format

The `--max-size` flag accepts:
- `10m` or `10M` - 10 megabytes
- `1g` or `1G` - 1 gigabyte
- `512k` or `512K` - 512 kilobytes

## Notes

- Input is stored in memory up to the max size, then spills to a temporary file
- Output only starts after all input is read
- Temporary files are automatically cleaned up
