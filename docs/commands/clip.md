# clip

Clipboard copy and paste.

## Synopsis

```bash
tofu clip [text]        # Copy text to clipboard
tofu clip -p            # Paste from clipboard
```

## Description

Copy to or paste from the system clipboard. Works cross-platform on Windows, macOS, and Linux.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--paste` | `-p` | Paste from clipboard to stdout | `false` |

## Examples

Copy text to clipboard:

```bash
tofu clip "Hello, World!"
```

Copy from stdin:

```bash
echo "Hello" | tofu clip
```

Copy file contents:

```bash
tofu clip < file.txt
```

Paste from clipboard:

```bash
tofu clip -p
```

Paste to file:

```bash
tofu clip -p > output.txt
```

Pipe clipboard contents:

```bash
tofu clip -p | grep "pattern"
```

## Notes

- When copying, if arguments are provided they are joined with spaces
- When no arguments are provided, reads from standard input
- Clipboard operations work with the system clipboard (same as Ctrl+C/Ctrl+V)
- On Linux, may require xclip or xsel to be installed
