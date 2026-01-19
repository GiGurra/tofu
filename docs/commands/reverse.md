# reverse

Output lines in reverse order.

## Synopsis

```bash
tofu reverse [files...]
```

## Description

Output the lines of each file in reverse order (last line first, first line last). Similar to the `tac` command.

## Examples

Reverse a file:

```bash
tofu reverse file.txt
```

Reverse from stdin:

```bash
echo -e "first\nsecond\nthird" | tofu reverse
```

Reverse multiple files:

```bash
tofu reverse file1.txt file2.txt
```

## Sample Output

Input:
```
line 1
line 2
line 3
```

Output:
```
line 3
line 2
line 1
```

## Use Cases

- Reading log files from newest to oldest
- Reversing command history
- Processing files from bottom to top
