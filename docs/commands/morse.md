# morse

Encode/decode Morse code.

## Synopsis

```bash
tofu morse [text] [flags]
```

## Description

Convert text to Morse code or decode Morse code back to text. Optionally plays audio beeps.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--decode` | `-d` | Decode Morse code to text | `false` |
| `--beep` | `-b` | Play audio beeps while encoding | `false` |
| `--wpm` | `-w` | Words per minute for audio | `15` |

## Examples

Encode text to Morse:

```bash
tofu morse "Hello World"
```

Decode Morse to text:

```bash
tofu morse -d ".... . .-.. .-.. --- / .-- --- .-. .-.. -.."
```

Encode from stdin:

```bash
echo "SOS" | tofu morse
```

Encode with audio:

```bash
tofu morse -b "SOS"
```

Faster audio playback:

```bash
tofu morse -b -w 25 "Hello"
```

## Sample Output

Encoding:
```
$ tofu morse "Hello World"
.... . .-.. .-.. --- / .-- --- .-. .-.. -..
```

Decoding:
```
$ tofu morse -d ".... . .-.. .-.. ---"
HELLO
```

## Morse Code Reference

| Character | Code | Character | Code |
|-----------|------|-----------|------|
| A | .- | N | -. |
| B | -... | O | --- |
| C | -.-. | P | .--. |
| D | -.. | Q | --.- |
| E | . | R | .-. |
| F | ..-. | S | ... |
| G | --. | T | - |
| H | .... | U | ..- |
| I | .. | V | ...- |
| J | .--- | W | .-- |
| K | -.- | X | -..- |
| L | .-.. | Y | -.-- |
| M | -- | Z | --.. |

Numbers: 0-9 follow standard Morse patterns.

## Notes

- Words are separated by `/` in Morse code
- Letters are separated by spaces
- Audio playback requires CGO on Linux
