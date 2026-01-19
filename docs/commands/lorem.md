# lorem

Generate Lorem Ipsum text.

## Synopsis

```bash
tofu lorem [flags]
```

## Description

Generate placeholder Lorem Ipsum text. Useful for testing layouts, filling templates, or creating mock content.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--paragraphs` | `-p` | Number of paragraphs | `1` |
| `--sentences` | `-s` | Number of sentences (overrides paragraphs) | `0` |
| `--words` | `-w` | Number of words (overrides sentences) | `0` |
| `--lorem` | `-l` | Start with "Lorem ipsum dolor sit amet" | `true` |

## Examples

One paragraph:

```bash
tofu lorem
# Output: Lorem ipsum dolor sit amet, consectetur adipiscing elit...
```

Multiple paragraphs:

```bash
tofu lorem -p 3
```

Specific number of sentences:

```bash
tofu lorem -s 5
```

Specific number of words:

```bash
tofu lorem -w 50
```

Without "Lorem ipsum" start:

```bash
tofu lorem -l=false
# Output: Consectetur adipiscing elit, sed do eiusmod tempor...
```

## Use Cases

- Filling page layouts during design
- Testing text rendering
- Creating placeholder content
- Generating mock data for demos
- Testing typography and line heights

## Notes

- When using `-s` (sentences), the `-p` flag is ignored
- When using `-w` (words), both `-p` and `-s` are ignored
- By default, text starts with the classic "Lorem ipsum dolor sit amet"
