# tee

Copy standard input to files and stdout.

## Synopsis

```bash
tofu tee [files...] [flags]
```

## Description

Read from standard input and write to standard output and files simultaneously. Useful for capturing output while still displaying it.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--append` | `-a` | Append to files instead of overwriting | `false` |
| `--ignore-interrupts` | `-i` | Ignore SIGINT signals | `false` |
| `--silent` | `-s` | Silent mode: only write to files | `false` |

## Examples

Write to file and stdout:

```bash
echo "hello" | tofu tee output.txt
```

Append to file:

```bash
echo "more data" | tofu tee -a output.txt
```

Write to multiple files:

```bash
echo "data" | tofu tee file1.txt file2.txt
```

Silent mode (no stdout):

```bash
echo "data" | tofu tee -s output.txt
```

In a pipeline:

```bash
./build.sh | tofu tee build.log
```

Ignore interrupts:

```bash
./long-running.sh | tofu tee -i output.log
```

## Use Cases

- Logging command output while watching it
- Duplicating data to multiple destinations
- Creating audit trails of command output
