# `tofu cat`

Concatenate files to standard output.

## Synopsis

```
tofu cat [files] [flags]
```

## Description

Concatenate FILE(s) to standard output. If no FILE is given, read from standard input.

## Options

- `-A, --show-all`: Equivalent to -vET (show all non-printing chars, ends, and tabs)
- `-b, --number-nonblank`: Number non-empty output lines, overrides -n
- `-E, --show-ends`: Display $ at end of each line
- `-n, --number`: Number all output lines
- `-s, --squeeze-blank`: Suppress repeated empty output lines
- `-T, --show-tabs`: Display TAB characters as ^I
- `-v, --show-non-printing`: Use ^ and M- notation for non-printing characters (except LFD and TAB)

## Examples

Display file contents:

```
tofu cat file.txt
```

Number all lines:

```
tofu cat -n file.txt
```

Show line endings and tabs:

```
tofu cat -ET file.txt
```

Concatenate multiple files:

```
tofu cat file1.txt file2.txt > combined.txt
```
