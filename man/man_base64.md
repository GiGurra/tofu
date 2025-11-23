# `tofu base64`

Base64 encode or decode data.

## Interface

```
> tofu base64 --help
Base64 encode or decode data

Usage:
  tofu base64 [flags] [files...]

Flags:
  -a, --alphabet string   Custom 64-character alphabet or predefined set (standard, url). (default "standard")
  -d, --decode            Decode data.
  -h, --help              help for base64
  -r, --no-padding        Do not write padding characters (raw) when encoding. Handle unpadded input when decoding.
  -u, --url-safe          Use URL-safe character set (alias for --alphabet url).
```

## Description

The `base64` command encodes or decodes data using the Base64 representation. It processes input from the specified files or standard input if no files are provided.

## Options

- **--alphabet, -a**: Specify the character set to use.
  - `standard`: The standard Base64 alphabet (RFC 4648).
  - `url`: The URL-safe Base64 alphabet (RFC 4648).
  - `<custom>`: A string of exactly 64 unique characters to use as the alphabet.
- **--decode, -d**: Decode the input data.
- **--url-safe, -u**: Shortcut for `--alphabet url`.
- **--no-padding, -r**: Omit padding characters (`=`) when encoding. When decoding, it allows unpadded input.

## Examples

**Encode a string:**

```bash
> echo -n "hello" | tofu base64
aGVsbG8=
```

**Decode a string:**

```bash
> echo "aGVsbG8=" | tofu base64 -d
hello
```

**URL-safe encoding (no padding):**

```bash
> echo -n "hello world?" | tofu base64 -u -r
aGVsbG8gd29ybGQ_
```

**Custom alphabet:**

```bash
> echo -n "hello" | tofu base64 -a "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+-"
aGVsbG8=
```
