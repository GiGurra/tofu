# diff

Compare files line by line.

## Synopsis

```bash
tofu diff <file1> <file2> [flags]
```

## Description

Compare two files and show differences with optional color output. Uses a unified diff format by default.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--unified` | `-u` | Output NUM lines of unified context | `3` |
| `--context` | `-c` | Output NUM lines of context | `0` |
| `--side-by-side` | `-y` | Output in two columns side by side | `false` |
| `--width` | `-W` | Output at most NUM columns (side-by-side) | `130` |
| `--color` | | Color output: `auto`, `always`, `never` | `auto` |
| `--no-color` | | Disable color output | `false` |
| `--brief` | `-q` | Report only when files differ | `false` |
| `--ignore-case` | `-i` | Ignore case differences | `false` |
| `--ignore-space` | `-b` | Ignore changes in whitespace | `false` |
| `--ignore-blank` | `-B` | Ignore blank lines | `false` |
| `--stats` | `-s` | Show statistics summary | `false` |

## Examples

Compare two files:

```bash
tofu diff old.txt new.txt
```

Side-by-side comparison:

```bash
tofu diff -y old.txt new.txt
```

Brief mode (just report if different):

```bash
tofu diff -q file1.txt file2.txt
```

Ignore whitespace changes:

```bash
tofu diff -b file1.txt file2.txt
```

Ignore case:

```bash
tofu diff -i file1.txt file2.txt
```

Show more context:

```bash
tofu diff -u 5 old.txt new.txt
```

Show statistics:

```bash
tofu diff -s old.txt new.txt
# Output includes: 5 insertion(s), 3 deletion(s)
```

No color output:

```bash
tofu diff --no-color old.txt new.txt
```

## Sample Output

```diff
--- old.txt
+++ new.txt
@@ -1,4 +1,4 @@
 Line 1
-Line 2 old
+Line 2 new
 Line 3
 Line 4
```
