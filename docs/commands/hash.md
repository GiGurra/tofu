# hash

Calculate cryptographic hashes.

## Synopsis

```bash
tofu hash [files...] [flags]
```

## Description

Calculate cryptographic hashes for files or standard input. Supports MD5, SHA-1, SHA-256, and SHA-512.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--algo` | `-a` | Hash algorithm: `md5`, `sha1`, `sha256`, `sha512` | `sha256` |

## Examples

Hash a file (SHA-256 by default):

```bash
tofu hash file.txt
```

Hash with MD5:

```bash
tofu hash -a md5 file.txt
```

Hash with SHA-512:

```bash
tofu hash -a sha512 file.txt
```

Hash from stdin:

```bash
echo "Hello, World!" | tofu hash
```

Hash multiple files:

```bash
tofu hash file1.txt file2.txt file3.txt
```

Verify a file hash:

```bash
tofu hash -a sha256 download.zip
# Compare output with expected hash
```

## Sample Output

```
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  file.txt
```

## Notes

- Output format matches standard tools (`sha256sum`, `md5sum`, etc.)
- The hash is displayed in hexadecimal format
- Use SHA-256 or SHA-512 for security-sensitive applications
- MD5 and SHA-1 are provided for compatibility but are not recommended for security
