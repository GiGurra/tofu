# sed2

Stream editor for filtering and transforming text.

## Synopsis

```bash
tofu sed2 <from> <to> [files...] [flags]
```

## Description

A sed-like stream editor with modern, readable syntax. Replace occurrences of a pattern with a replacement string.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--search-type` | `-t` | Pattern type: `literal`, `regex` | `regex` |
| `--in-place` | `-i` | Edit files in place | `false` |
| `--ignore-case` | `-I` | Case-insensitive search | `false` |
| `--global` | `-g` | Replace all occurrences on each line | `false` |

## Examples

Replace first occurrence on each line:

```bash
tofu sed2 "old" "new" file.txt
```

Replace all occurrences (global):

```bash
tofu sed2 -g "old" "new" file.txt
```

Case-insensitive replacement:

```bash
tofu sed2 -I "error" "warning" log.txt
```

Edit file in place:

```bash
tofu sed2 -i "foo" "bar" config.txt
```

Use literal string (not regex):

```bash
tofu sed2 -t literal "user.name" "user.email" file.txt
```

Regex with capture groups:

```bash
tofu sed2 "(\w+)@example.com" "$1@newdomain.com" emails.txt
```

Read from stdin:

```bash
echo "hello world" | tofu sed2 "world" "universe"
```

Process multiple files:

```bash
tofu sed2 -g "TODO" "DONE" *.md
```
