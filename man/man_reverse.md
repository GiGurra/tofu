# `tofu reverse`

Output lines in reverse order.

## Synopsis

```
tofu reverse [files] [flags]
```

## Description

Output lines in reverse order. Reads from files or standard input if no files are specified.

## Examples

Reverse lines in a file:

```
tofu reverse file.txt
```

Reverse piped input:

```
echo -e "line1\nline2\nline3" | tofu reverse
```

Reverse and save to new file:

```
tofu reverse input.txt > reversed.txt
```
