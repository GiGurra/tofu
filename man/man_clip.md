# `tofu clip`

Clipboard copy and paste utility.

## Interface

```
> tofu clip --help
Clipboard copy and paste

Usage:
  tofu clip [text] [flags]

Flags:
  -p, --paste   Paste from clipboard to standard output. (default false)
  -h, --help    help for clip
```

### Examples

**Copy to clipboard:**

```bash
echo "Hello World" | tofu clip
```

or

```bash
tofu clip "Hello World"
```

**Paste from clipboard:**

```bash
tofu clip -p
```

**Copy file content:**

```bash
tofu cat file.txt | tofu clip
```
