# find

Find file system items, optionally matching a search term.

## Synopsis

```bash
tofu find [search-term] [flags]
```

## Description

Search for files and directories by name. If no search term is provided, all matching items are listed.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--search-type` | `-s` | Type of search: `exact`, `contains`, `prefix`, `suffix`, `regex` | `contains` |
| `--ignore-case` | `-i` | Perform case-insensitive search | `false` |
| `--work-dir` | `-c` | Directory to start the search from | `.` |
| `--types` | `-t` | Types to search for: `file`, `dir`, `all` | `all` |
| `--quiet` | `-q` | Suppress error messages | `false` |

## Examples

Find all files containing "test" in their name:

```bash
tofu find test
```

Find files with exact name match:

```bash
tofu find config.json -s exact
```

Find files starting with "main":

```bash
tofu find main -s prefix
```

Find files ending with ".go":

```bash
tofu find .go -s suffix
```

Find using regex pattern:

```bash
tofu find "test_.*\.go$" -s regex
```

Search case-insensitively:

```bash
tofu find README -i
```

Find only directories:

```bash
tofu find src -t dir
```

Find only files:

```bash
tofu find .md -s suffix -t file
```

Search in a specific directory:

```bash
tofu find test -c /path/to/project
```
