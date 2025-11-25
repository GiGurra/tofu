# `tofu diff`

Compare two files line by line and show differences with optional color output.

## Interface

```
> tofu diff --help
Compare files line by line

Usage:
  tofu diff <file1> <file2> [flags]

Flags:
  -u, --unified int     Output NUM lines of unified context. (default 3)
  -c, --context int     Output NUM lines of context.
  -y, --side-by-side    Output in two columns side by side.
  -W, --width int       Output at most NUM columns (for side-by-side). (default 130)
      --color string    Color output (auto, always, never). (default "auto")
      --no-color        Disable color output.
  -q, --brief           Report only when files differ.
  -i, --ignore-case     Ignore case differences.
  -b, --ignore-space    Ignore changes in whitespace.
  -B, --ignore-blank    Ignore blank lines.
  -s, --stats           Show statistics summary.
  -h, --help            help for diff
```

## Output Format

By default, `tofu diff` outputs in unified diff format with colored output (when terminal supports it):

- **Red** lines with `-` prefix: deleted from file1
- **Green** lines with `+` prefix: added in file2
- **Cyan** hunk headers: `@@ -start,count +start,count @@`
- **Yellow** file headers: `--- file1` and `+++ file2`

## Examples

Basic diff between two files:

```
> tofu diff old.txt new.txt
--- old.txt
+++ new.txt
@@ -1,4 +1,4 @@
 line 1
-line 2
+modified line 2
 line 3
 line 4
```

Brief mode (only report if different):

```
> tofu diff -q file1.txt file2.txt
Files file1.txt and file2.txt differ
```

Side-by-side comparison:

```
> tofu diff -y old.txt new.txt
line 1                           line 1
line 2                         <
                               > modified line 2
line 3                           line 3
```

With statistics:

```
> tofu diff -s old.txt new.txt
--- old.txt
+++ new.txt
@@ -1,3 +1,3 @@
 line 1
-old content
+new content
 line 3

1 insertion(s), 1 deletion(s)
```

Ignore case differences:

```
> tofu diff -i file1.txt file2.txt
```

Ignore whitespace changes:

```
> tofu diff -b file1.txt file2.txt
```

Ignore blank lines:

```
> tofu diff -B file1.txt file2.txt
```

More context lines:

```
> tofu diff -u 5 file1.txt file2.txt
```

Force color output (e.g., when piping):

```
> tofu diff --color=always file1.txt file2.txt | less -R
```

Disable color:

```
> tofu diff --no-color file1.txt file2.txt
```

## Color Modes

| Mode | Description |
|------|-------------|
| `auto` | Color when output is a terminal (default) |
| `always` | Always use color |
| `never` | Never use color |

## Exit Codes

- `0` - Files are identical (or no errors)
- `1` - Files are different or an error occurred

## Comparison with GNU diff

| GNU diff | tofu diff | Description |
|----------|-----------|-------------|
| `diff -u` | `tofu diff -u` | Unified format |
| `diff -y` | `tofu diff -y` | Side-by-side |
| `diff -q` | `tofu diff -q` | Brief (report if different) |
| `diff -i` | `tofu diff -i` | Ignore case |
| `diff -b` | `tofu diff -b` | Ignore whitespace |
| `diff -B` | `tofu diff -B` | Ignore blank lines |
| `diff --color` | `tofu diff --color` | Colored output |
