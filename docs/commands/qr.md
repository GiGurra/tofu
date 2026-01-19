# qr

Render QR codes in the terminal.

## Synopsis

```bash
tofu qr <text> [flags]
```

## Description

Generate and display QR codes directly in the terminal using ANSI colors.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--recovery-level` | `-r` | Error recovery: `low`, `medium`, `high`, `highest` | `medium` |
| `--invert` | `-i` | Invert colors (white on black) | `false` |

## Examples

Generate a QR code for a URL:

```bash
tofu qr "https://example.com"
```

Generate from stdin:

```bash
echo "Hello World" | tofu qr
```

High error recovery:

```bash
tofu qr -r high "https://example.com"
```

Inverted colors (for dark terminals):

```bash
tofu qr -i "https://example.com"
```

Encode WiFi credentials:

```bash
tofu qr "WIFI:T:WPA;S:MyNetwork;P:MyPassword;;"
```

## Error Recovery Levels

| Level | Recovery Capacity |
|-------|-------------------|
| `low` | ~7% |
| `medium` | ~15% |
| `high` | ~25% |
| `highest` | ~30% |

Higher recovery levels create larger QR codes but can recover from more damage.

## Notes

- The QR code is rendered using ANSI background colors
- Works best in terminals that support ANSI color codes
- Use `-i` (invert) if scanning is difficult with your terminal's color scheme
