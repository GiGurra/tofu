# tail

Output the last part of files.

## Synopsis

```bash
tofu tail [files...] [flags]
```

## Description

Print the last N lines of each FILE to standard output. If no files are specified, read from standard input. With the `-f` option, follow file changes in real-time.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--lines` | `-n` | Output the last N lines | `10` |
| `--follow` | `-f` | Output appended data as file grows | `false` |
| `--quiet` | `-q` | Never output headers giving file names | `false` |
| `--verbose` | `-v` | Always output headers giving file names | `false` |

## Examples

Show last 10 lines of a file:

```bash
tofu tail file.txt
```

Show last 20 lines:

```bash
tofu tail -n 20 file.txt
```

Follow file changes (like `tail -f`):

```bash
tofu tail -f /var/log/app.log
```

Follow multiple files:

```bash
tofu tail -f file1.log file2.log
```

Show last lines of multiple files:

```bash
tofu tail file1.txt file2.txt
```

Read from stdin:

```bash
cat bigfile.txt | tofu tail -n 5
```

## Sample Output

With multiple files and follow mode:

```
==> file1.log <==
Latest log entry 1
Latest log entry 2

==> file2.log <==
Another log entry
```
