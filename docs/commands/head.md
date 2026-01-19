# head

Output the first part of files.

## Synopsis

```bash
tofu head [files...] [flags]
```

## Description

Print the first N lines of each FILE to standard output. If no files are specified, read from standard input. With more than one FILE, precede each with a header giving the file name.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--lines` | `-n` | Output the first N lines | `10` |
| `--quiet` | `-q` | Never output headers giving file names | `false` |
| `--verbose` | `-v` | Always output headers giving file names | `false` |

## Examples

Show first 10 lines of a file:

```bash
tofu head file.txt
```

Show first 20 lines:

```bash
tofu head -n 20 file.txt
```

Show first lines of multiple files:

```bash
tofu head file1.txt file2.txt
```

Suppress file headers:

```bash
tofu head -q file1.txt file2.txt
```

Read from stdin:

```bash
cat bigfile.txt | tofu head -n 5
```

## Sample Output

With multiple files:

```
==> file1.txt <==
Line 1 of file1
Line 2 of file1
...

==> file2.txt <==
Line 1 of file2
Line 2 of file2
...
```
