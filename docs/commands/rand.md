# rand

Generate random data.

## Synopsis

```bash
tofu rand [flags]
```

## Description

Generate random data in various formats: strings, integers, hex, base64, passwords, or passphrases.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--type` | `-t` | Type: `str`, `int`, `hex`, `base64`, `password`, `phrase` | `str` |
| `--length` | `-l` | Length of output (chars for str/hex/base64/password, words for phrase) | `16` |
| `--count` | `-n` | Number of values to generate | `1` |
| `--min` | | Minimum value for int type | `0` |
| `--max` | | Maximum value for int type | `100` |
| `--charset` | `-c` | Custom character set for str type | |

## Examples

Random string:

```bash
tofu rand
# Output: xK9mPq2nL8vR4wYj
```

Random integer:

```bash
tofu rand -t int --min 1 --max 100
# Output: 42
```

Random hex string:

```bash
tofu rand -t hex -l 32
# Output: a1b2c3d4e5f6789012345678abcdef01
```

Random base64:

```bash
tofu rand -t base64 -l 24
# Output: SGVsbG8gV29ybGQhIQ==
```

Generate a password:

```bash
tofu rand -t password -l 20
# Output: K9$mP@2n!L8vR#4wYj&x
```

Generate a passphrase:

```bash
tofu rand -t phrase -l 4
# Output: correct horse battery staple
```

Multiple values:

```bash
tofu rand -n 5
# Output:
# xK9mPq2nL8vR4wYj
# Hy7tGb3fNk9sWz1x
# ...
```

Custom character set:

```bash
tofu rand -c "abc123" -l 10
# Output: a1b2c3a1b2
```

## Types

| Type | Description |
|------|-------------|
| `str` | Alphanumeric string (A-Z, a-z, 0-9) |
| `int` | Integer within min/max range |
| `hex` | Hexadecimal string (0-9, a-f) |
| `base64` | Base64-encoded random bytes |
| `password` | String with letters, numbers, and symbols |
| `phrase` | Random words (passphrase style) |
