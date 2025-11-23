# `tofu rand`

Generates random data (strings, integers, hex, base64, passwords, passphrases).

## Interface

```
> tofu rand --help
Generate random data

Usage:
  tofu rand [flags]

Flags:
  -t, --type string       Type of random data (str, int, hex, base64, password, phrase). (default "str")
  -l, --length int        Length (chars for str/password/hex/base64, words for phrase). (default 16)
  -n, --count int         Number of items to generate. (default 1)
      --min int           Minimum value for integer generation. (default 0)
      --max int           Maximum value for integer generation. (default 100)
  -c, --charset string    Custom character set for string generation.
      --separator string  Separator for phrases. (default " ")
      --capitalize string Capitalization for phrases (none, first, all, random, one). (default "none")
  -h, --help              help for rand
```

## Types

- **str**: Random string using a charset (default: alphanumeric).
- **int**: Random integer within a range [min, max].
- **hex**: Random bytes encoded as a hexadecimal string.
- **base64**: Random bytes encoded as a Base64 string.
- **password**: Strong random password using 1Password's SPG (Letters, Digits, Symbols).
- **phrase**: Memorable passphrase (e.g., "correct horse battery staple").

### Examples

Generate a random string (default):

```
> tofu rand
abc123XYZ...
```

Generate a strong password (20 chars):

```
> tofu rand -t password -l 20
8*3kL@9#...
```

Generate a passphrase (4 words):

```
> tofu rand -t phrase -l 4
correct horse battery staple
```

Generate a passphrase with custom separator and capitalization:

```
> tofu rand -t phrase -l 4 --separator "-" --capitalize all
Correct-Horse-Battery-Staple
```

Generate 5 random integers between 1 and 10:

```
> tofu rand -t int --min 1 --max 10 -n 5
7
3
10
...
```

Generate a random hex string (16 bytes -> 32 hex chars):

```
> tofu rand -t hex -l 16
a1b2c3d4...
```