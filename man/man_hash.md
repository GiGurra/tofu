# `tofu hash`

Calculate cryptographic hashes for files or standard input.

## Interface

```
> tofu hash --help
Calculate file hashes

Usage:
  tofu hash [flags] [files...]

Flags:
  -a, --algo string   Hash algorithm (md5, sha1, sha256, sha512). (default "sha256")
  -h, --help          help for hash
```

## Description

The `hash` command computes and displays the cryptographic hash of specified files or data read from standard input. It mimics the behavior of tools like `sha256sum` or `md5sum`.

Supported algorithms:
- md5
- sha1
- sha256 (default)
- sha512

## Examples

Hash a file using SHA-256 (default):

```
> tofu hash myfile.txt
2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824  myfile.txt
```

Hash a string from stdin using MD5:

```
> echo -n "hello" | tofu hash -a md5
5d41402abc4b2a76b9719d911017c592  -
```

Hash multiple files with SHA-512:

```
> tofu hash -a sha512 file1.txt file2.txt
...  file1.txt
...  file2.txt
```
