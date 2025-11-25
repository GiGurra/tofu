# `tofu count`

Count lines, words, characters, and bytes in files. A modern alternative to `wc` with clearer flags.

## Interface

```
> tofu count --help
Count lines, words, and characters

Usage:
  tofu count [files...] [flags]

Flags:
  -l, --lines        Print the line count.
  -w, --words        Print the word count.
  -c, --chars        Print the character count.
  -b, --bytes        Print the byte count.
  -L, --max-line     Print the length of the longest line.
  -t, --total-only   Print only the total (when multiple files).
  -n, --no-filename  Never print filenames.
  -h, --help         help for count
```

## Default Behavior

When no flags are specified, `tofu count` shows lines, words, and characters (similar to `wc`):

```
> tofu count file.txt
      10      50     300 file.txt
```

## Examples

Count lines only (like `wc -l`):

```
> tofu count -l file.txt
      10
```

Count words only:

```
> tofu count -w file.txt
      50
```

Count characters (Unicode-aware):

```
> tofu count -c file.txt
     300
```

Count bytes:

```
> tofu count -b file.txt
     312
```

Find the longest line length:

```
> tofu count -L file.txt
      80
```

Multiple files:

```
> tofu count file1.txt file2.txt
      10      50     300 file1.txt
      20     100     600 file2.txt
      30     150     900 total
```

Only show total for multiple files:

```
> tofu count -t file1.txt file2.txt
      30     150     900 total
```

Read from stdin:

```
> cat file.txt | tofu count
      10      50     300
```

Or explicitly:

```
> tofu count - < file.txt
      10      50     300
```

Combine flags:

```
> tofu count -lw file.txt
      10      50
```

## Comparison with wc

| wc | tofu count | Description |
|----|------------|-------------|
| `wc -l` | `tofu count -l` | Line count |
| `wc -w` | `tofu count -w` | Word count |
| `wc -c` | `tofu count -b` | Byte count |
| `wc -m` | `tofu count -c` | Character count (Unicode) |
| `wc -L` | `tofu count -L` | Max line length |

## Unicode Support

`tofu count` properly handles Unicode text:

- `-c` (chars) counts Unicode code points (runes)
- `-b` (bytes) counts raw bytes
- `-L` (max-line) measures line length in characters, not bytes

```
> echo "日本語" | tofu count -c -b
       4       10
```
(3 characters + newline = 4 chars, but 9 bytes for UTF-8 + 1 newline = 10 bytes)
