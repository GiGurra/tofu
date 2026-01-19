# cat

Concatenate files to standard output.

## Synopsis

```bash
tofu cat [files...] [flags]
```

## Description

Concatenate FILE(s) to standard output. If no files are specified or `-` is used, read from standard input.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--show-all` | `-A` | Equivalent to -vET (show all non-printing chars, ends, and tabs) | `false` |
| `--number-nonblank` | `-b` | Number non-empty output lines, overrides -n | `false` |
| `--show-ends` | `-E` | Display $ at end of each line | `false` |
| `--number` | `-n` | Number all output lines | `false` |
| `--squeeze-blank` | `-s` | Suppress repeated empty output lines | `false` |
| `--show-tabs` | `-T` | Display TAB characters as ^I | `false` |
| `--show-nonprinting` | `-v` | Use ^ and M- notation for non-printing characters | `false` |

## Examples

Display a file:

```bash
tofu cat file.txt
```

Concatenate multiple files:

```bash
tofu cat file1.txt file2.txt file3.txt
```

Number all lines:

```bash
tofu cat -n file.txt
```

Show line endings and tabs:

```bash
tofu cat -ET file.txt
```

Read from stdin and pipe to another command:

```bash
echo "hello" | tofu cat -n
```

Suppress repeated empty lines:

```bash
tofu cat -s file_with_blanks.txt
```
