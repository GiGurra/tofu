# `tofu tee`

Copy standard input to each FILE, and also to standard output.

## Synopsis

```
tofu tee [files] [flags]
```

## Description

Read from standard input and write to standard output and files simultaneously. Useful for capturing output while still displaying it.

## Options

- `-a, --append`: Append to the given FILEs, do not overwrite
- `-i, --ignore-interrupts`: Ignore interrupt signals (SIGINT)
- `-s, --silent`: Silent mode: do not write to stdout, only to files

## Examples

Write to file while displaying:

```
echo "hello" | tofu tee output.txt
```

Append to file:

```
echo "new line" | tofu tee -a log.txt
```

Write to multiple files:

```
echo "data" | tofu tee file1.txt file2.txt
```

Silent mode (only write to file):

```
echo "quiet" | tofu tee -s output.txt
```

Use in a pipeline:

```
cat input.txt | tofu tee backup.txt | grep "pattern"
```
