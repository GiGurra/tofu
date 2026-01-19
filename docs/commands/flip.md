# flip

Flip text upside down.

## Synopsis

```bash
tofu flip [text] [flags]
```

## Description

Transform text to appear upside down. Perfect for expressing frustration or just having fun. (╯°□°)╯︵ ┻━┻

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--table` | `-t` | Add table flip emote | `false` |

## Examples

Flip text:

```bash
tofu flip "Hello World"
# Output: plɹoM ollǝH
```

Table flip mode:

```bash
tofu flip -t "This code"
# Output: (╯°□°)╯︵ ǝpoɔ sᴉɥ⊥
```

From stdin:

```bash
echo "Frustration" | tofu flip
# Output: uoᴉʇɐɹʇsnɹℲ
```

## Character Mappings

| Original | Flipped |
|----------|---------|
| a | ɐ |
| b | q |
| d | p |
| e | ǝ |
| f | ɟ |
| g | ƃ |
| h | ɥ |
| i | ᴉ |
| m | ɯ |
| n | u |
| r | ɹ |
| t | ʇ |
| w | ʍ |
| y | ʎ |
| ! | ¡ |
| ? | ¿ |
| . | ˙ |

## Notes

- Text is reversed and characters are flipped
- Unsupported characters remain unchanged
- Works with stdin for piping
