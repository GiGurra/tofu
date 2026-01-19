# lolcat

Rainbow colorize text.

## Synopsis

```bash
tofu lolcat [text] [flags]
```

## Description

Output text in rainbow colors using ANSI true color. Makes everything better.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--freq` | `-f` | Rainbow frequency (color change rate) | `0.1` |
| `--spread` | `-p` | Rainbow spread (how much colors spread across lines) | `3.0` |
| `--seed` | `-s` | Rainbow seed (starting point in the rainbow) | `0` |

## Examples

Colorize text:

```bash
tofu lolcat "Hello, World!"
```

Colorize from stdin:

```bash
echo "Rainbow text" | tofu lolcat
```

Colorize a file:

```bash
cat README.md | tofu lolcat
```

Faster color cycling:

```bash
tofu lolcat -f 0.3 "Speedy rainbow"
```

Different starting color:

```bash
tofu lolcat -s 10 "Different start"
```

Wider color spread:

```bash
tofu lolcat -p 5.0 "Spread out colors"
```

## Fun Uses

Colorize cowsay:
```bash
tofu cowsay "Hello!" | tofu lolcat
```

Rainbow build output:
```bash
make 2>&1 | tofu lolcat
```

Rainbow fortune:
```bash
tofu fortune | tofu lolcat
```

## Notes

- Requires a terminal that supports ANSI true color (24-bit color)
- Spaces and tabs are not colorized
