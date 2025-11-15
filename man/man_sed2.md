# `tofu sed2`

A sed-like-but-different stream editor for filtering and transforming text.

## Interface

```
> tofu sed2 --help
sed-like-but-different stream editor for filtering and transforming text

Usage:
  tofu sed2 <from> <to> [files] [flags]

Flags:
  -t, --search-type string   Type of pattern to search for (literal, regex). (default "regex")
  -i, --in-place             Edit files in place. (default false)
  -I, --ignore-case          Perform a case-insensitive search. (default false)
  -g, --global               Replace all occurrences on each line (not just first). (default false)
  -h, --help                 help for sed2
```

## Examples

### Basic replacement

Replace the first occurrence of "foo" with "bar" in a file:

```bash
> tofu sed2 foo bar input.txt
```

### Global replacement

Replace all occurrences of "foo" with "bar" on each line:

```bash
> tofu sed2 -g foo bar input.txt
```

### In-place editing

Edit a file in place (modifies the original file):

```bash
> tofu sed2 -i -g foo bar input.txt
```

### Case-insensitive replacement

Replace "foo" with "bar" ignoring case:

```bash
> tofu sed2 -I foo bar input.txt
# Matches: foo, Foo, FOO, fOo, etc.
```

### Literal pattern matching

Use literal strings instead of regex (useful for special characters):

```bash
> tofu sed2 -t literal "foo.*" bar input.txt
# Matches the literal string "foo.*", not as a regex
```

### Using regex patterns

Replace with capture groups:

```bash
> tofu sed2 '(\w+) (\w+)' '$2 $1' input.txt
# Swaps the first two words on each line
```

Replace dates in format YYYY-MM-DD to DD/MM/YYYY:

```bash
> tofu sed2 '(\d{4})-(\d{2})-(\d{2})' '$3/$2/$1' input.txt
```

### Reading from stdin

Process piped input:

```bash
> echo "hello world" | tofu sed2 world universe
hello universe

> cat file.txt | tofu sed2 -g old new
```

### Multiple files

Process multiple files (outputs to stdout):

```bash
> tofu sed2 foo bar file1.txt file2.txt file3.txt
```

Edit multiple files in place:

```bash
> tofu sed2 -i -g old new *.txt
```

## Differences from traditional sed

- Simpler syntax focused on search-and-replace operations
- Regex by default (use `-t literal` for literal matching)
- Clear flag names (`-g` for global, `-i` for in-place)
- Works consistently across Windows, macOS, and Linux
- Supports large files (up to 10MB per line)
- Built-in support for capture groups in replacements

## Notes

- When using `-i` (in-place), the entire file is read into memory before writing
- Without `-g`, only the first match on each line is replaced
- Regex patterns use Go's RE2 syntax (similar to PCRE)
- Capture groups in replacements use `$1`, `$2`, etc.
- Files are created with permissions `0644` when edited in-place
