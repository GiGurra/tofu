# crypt

Encrypt and decrypt files.

## Synopsis

```bash
tofu crypt encrypt <files...> [flags]
tofu crypt decrypt <files...> [flags]
```

## Description

Encrypt and decrypt files using modern authenticated encryption. Supports two formats for maximum compatibility.

## Supported Formats

| Format | Extension | Description | Interoperable With |
|--------|-----------|-------------|-------------------|
| age | .age | Modern encryption (default) | `age` CLI |
| openssl | .enc | Legacy compatible | `openssl enc` |

### age (default)

- **Key derivation**: scrypt (memory-hard, GPU-resistant)
- **Encryption**: ChaCha20-Poly1305 (authenticated)
- **Recommended for**: New projects, security-focused use cases

### openssl

- **Key derivation**: PBKDF2-SHA256 (600,000 iterations)
- **Encryption**: AES-256-CBC
- **Recommended for**: Compatibility with systems that only have OpenSSL

## Commands

### encrypt

Encrypt one or more files.

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output file (single input only) | `<input>.age` or `<input>.enc` |
| `--password` | `-p` | Encryption password | (prompted) |
| `--format` | `-f` | Output format: `age`, `openssl` | `age` |
| `--keep` | `-k` | Keep original files | `false` |
| `--force` | `-F` | Overwrite existing output | `false` |
| `--verbose` | `-v` | Verbose output | `false` |

### decrypt

Decrypt one or more files.

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output file (single input only) | (removes .age/.enc) |
| `--password` | `-p` | Decryption password | (prompted) |
| `--format` | `-f` | Input format: `auto`, `age`, `openssl` | `auto` |
| `--keep` | `-k` | Keep encrypted files | `false` |
| `--force` | `-F` | Overwrite existing output | `false` |
| `--verbose` | `-v` | Verbose output | `false` |

## Examples

Encrypt a file (age format, default):

```bash
tofu crypt encrypt secret.txt
```

Encrypt with password on command line:

```bash
tofu crypt encrypt -p mypassword document.pdf
```

Encrypt multiple files:

```bash
tofu crypt encrypt -k file1.txt file2.txt file3.txt
```

Encrypt in OpenSSL-compatible format:

```bash
tofu crypt encrypt -f openssl secret.txt
```

Decrypt a file (auto-detects format):

```bash
tofu crypt decrypt secret.txt.age
```

Decrypt to specific output:

```bash
tofu crypt decrypt -o recovered.txt secret.txt.age
```

Decrypt keeping the encrypted file:

```bash
tofu crypt decrypt -k secret.txt.age
```

## Interoperability

### With age CLI

```bash
# Encrypt with tofu, decrypt with age
tofu crypt encrypt -p secret file.txt
age -d file.txt.age  # enter "secret" when prompted

# Encrypt with age, decrypt with tofu
age -p -o file.age file.txt
tofu crypt decrypt -p <password> file.age
```

### With OpenSSL CLI

```bash
# Encrypt with tofu, decrypt with openssl
tofu crypt encrypt -f openssl -p secret file.txt
openssl enc -d -aes-256-cbc -pbkdf2 -iter 600000 -in file.txt.enc -out file.txt -pass pass:secret

# Encrypt with openssl, decrypt with tofu
openssl enc -aes-256-cbc -pbkdf2 -iter 600000 -in file.txt -out file.enc -pass pass:secret
tofu crypt decrypt -p secret file.enc
```

## Aliases

- `tofu crypt e` or `enc` - alias for `encrypt`
- `tofu crypt d` or `dec` - alias for `decrypt`

## Security Notes

- The `age` format is recommended for security (authenticated encryption)
- Passwords are prompted interactively with confirmation when encrypting
- Original files are deleted by default after encryption (use `-k` to keep)
- The OpenSSL format uses CBC mode without authentication; prefer `age` when possible
