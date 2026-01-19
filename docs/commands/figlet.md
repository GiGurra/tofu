# figlet

ASCII art text banners.

## Synopsis

```bash
tofu figlet [text] [flags]
```

## Description

Render text as large ASCII art banners. Similar to the classic FIGlet program.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--font` | `-f` | Font: `standard`, `small`, `mini`, `block` | `standard` |

## Examples

Render text:

```bash
tofu figlet "HELLO"
```

Use small font:

```bash
tofu figlet -f small "HI"
```

Use block font:

```bash
tofu figlet -f block "GO"
```

From stdin:

```bash
echo "TEXT" | tofu figlet
```

## Sample Output

Standard font:
```
█   █ █████ █     █      ███
█   █ █     █     █     █   █
█████ ████  █     █     █   █
█   █ █     █     █     █   █
█   █ █████ █████ █████  ███
```

Small font:
```
█ █ ██▀ █   █   ▄█▄
███ █▄  █   █   █ █
█ █ ██▄ ███ ███ ▀█▀
```

Block font:
```
███ ███ ███ ███ ███
███ ███ ███ ███ ███
███ ███ ███ ███ ███
```

## Supported Characters

- A-Z (uppercase letters)
- 0-9 (digits)
- Space, period, comma, exclamation, question mark
- Hyphen, underscore, colon
