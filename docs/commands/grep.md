# grep

Search for patterns in files.

## Synopsis

```bash
tofu grep <pattern> [files...] [flags]
```

## Description

Search for PATTERN in each FILE. If no files are specified or `-` is used, read from standard input. Matches are highlighted in color by default.

## Flags

### Pattern Matching

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--pattern-type` | `-t` | Pattern type: `basic`, `extended`, `fixed`, `perl` | `extended` |
| `--ignore-case` | `-i` | Case-insensitive matching | `false` |
| `--invert-match` | `-v` | Select non-matching lines | `false` |
| `--word-regexp` | `-w` | Match only whole words | `false` |
| `--line-regexp` | `-x` | Match only whole lines | `false` |

### Output Control

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--line-number` | `-n` | Print line numbers | `false` |
| `--with-filename` | `-H` | Print filename with output | `false` |
| `--no-filename` | | Suppress filename prefix | `false` |
| `--count` | `-c` | Print only count of matches | `false` |
| `--files-with-match` | `-l` | Print only names of files with matches | `false` |
| `--files-without-match` | `-L` | Print only names of files without matches | `false` |
| `--only-matching` | `-o` | Show only matched parts | `false` |
| `--quiet` | `-q` | Suppress all output | `false` |
| `--ignore-binary` | | Suppress output for binary files | `false` |
| `--max-count` | `-m` | Stop after NUM matches per file | `0` |

### Context Control

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--before-context` | `-B` | Print NUM lines before match | `0` |
| `--after-context` | `-A` | Print NUM lines after match | `0` |
| `--context` | `-C` | Print NUM lines before and after | `0` |

### File/Directory Handling

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--recursive` | `-r` | Search directories recursively | `false` |
| `--include` | `-I` | Search only files matching pattern | |
| `--exclude` | `-e` | Skip files matching pattern | |
| `--exclude-dir` | | Skip directories matching pattern | |
| `--no-messages` | `-s` | Suppress error messages | `false` |

## Examples

Search for a pattern in a file:

```bash
tofu grep "error" log.txt
```

Case-insensitive search:

```bash
tofu grep -i "warning" log.txt
```

Search with line numbers:

```bash
tofu grep -n "TODO" *.go
```

Recursive search in directory:

```bash
tofu grep -r "import" ./src
```

Show context around matches:

```bash
tofu grep -C 3 "panic" main.go
```

Count matches:

```bash
tofu grep -c "test" *_test.go
```

List files containing pattern:

```bash
tofu grep -l "Copyright" *.go
```

Search for whole word:

```bash
tofu grep -w "log" main.go
```

Use fixed string (not regex):

```bash
tofu grep -t fixed "user.name" config.json
```

Include only Go files:

```bash
tofu grep -r -I "*.go" "func main"
```

Exclude test files:

```bash
tofu grep -r -e "*_test.go" "func"
```
