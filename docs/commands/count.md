# count

Count lines, words, and characters.

## Synopsis

```bash
tofu count [files...] [flags]
```

## Description

Count lines, words, characters, and bytes in files. Similar to `wc` but with clearer flags.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--lines` | `-l` | Print the line count | `false` |
| `--words` | `-w` | Print the word count | `false` |
| `--chars` | `-c` | Print the character count | `false` |
| `--bytes` | `-b` | Print the byte count | `false` |
| `--max-line` | `-L` | Print the length of the longest line | `false` |
| `--total-only` | `-t` | Print only the total (for multiple files) | `false` |
| `--no-filename` | `-n` | Never print filenames | `false` |

## Examples

Count lines, words, and characters (default):

```bash
tofu count file.txt
```

Count only lines:

```bash
tofu count -l file.txt
```

Count words:

```bash
tofu count -w file.txt
```

Count bytes:

```bash
tofu count -b file.txt
```

Find longest line:

```bash
tofu count -L file.txt
```

Count from stdin:

```bash
cat file.txt | tofu count
```

Count multiple files:

```bash
tofu count file1.txt file2.txt file3.txt
```

Show only total for multiple files:

```bash
tofu count -t *.txt
```

## Sample Output

Default output (lines, words, chars):
```
     42     256    1542 file.txt
```

Multiple files:
```
     42     256    1542 file1.txt
     18     102     612 file2.txt
     60     358    2154 total
```
