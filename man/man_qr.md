# `tofu qr`

Render QR codes directly in your terminal.

## Interface

```
> tofu qr --help
Render QR codes in the terminal

Usage:
  tofu qr [text] [flags]

Flags:
  -r, --recovery-level      Error recovery level (low, medium, high, highest). (default "medium")
  -i, --invert              Invert colors (white on black). Default is standard black on white. (default false)
  -h, --help                help for qr
```

### Examples

**Standard QR Code:**
```
> tofu qr "https://example.com"
[QR Code Displayed]
```

**High Error Recovery:**
Useful for complex data or if the QR code might be damaged/partially obscured.
```
> tofu qr "Important Data" -r high
```

**Inverted Colors:**
If you prefer white modules on black background (or if your terminal/scanner prefers it).
```
> tofu qr "Inverted" --invert
```

**Pipe from Stdin:**
```
> echo "https://example.com" | tofu qr
```
