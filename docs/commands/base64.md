# base64

Base64 encode or decode data.

## Synopsis

```bash
tofu base64 [files...] [flags]
```

## Description

Encode data to base64 or decode base64 data. Supports standard and URL-safe alphabets, with or without padding.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--decode` | `-d` | Decode data | `false` |
| `--url-safe` | `-u` | Use URL-safe character set | `false` |
| `--no-padding` | `-r` | No padding characters (raw) | `false` |
| `--alphabet` | `-a` | Alphabet: `standard`, `url`, or custom 64-char string | `standard` |

## Examples

Encode a string:

```bash
echo "Hello, World!" | tofu base64
```

Decode base64:

```bash
echo "SGVsbG8sIFdvcmxkIQ==" | tofu base64 -d
```

Encode a file:

```bash
tofu base64 image.png > image.b64
```

URL-safe encoding:

```bash
echo "data?foo=bar" | tofu base64 -u
```

Raw encoding (no padding):

```bash
echo "test" | tofu base64 -r
```

Decode from multiple files:

```bash
tofu base64 -d file1.b64 file2.b64
```

## Notes

- Standard alphabet uses `+` and `/`
- URL-safe alphabet uses `-` and `_`
- When decoding, the tool handles both padded and unpadded input
