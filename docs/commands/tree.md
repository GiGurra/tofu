# tree

List contents of directories in a tree-like format.

## Synopsis

```bash
tofu tree [directory] [flags]
```

## Description

Display the directory structure in a visual tree format. Shows files and subdirectories with their hierarchical relationships.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--depth` | `-L` | Descend only N levels deep (-1 for unlimited) | `-1` |
| `--all` | `-a` | Show entries starting with `.` | `false` |
| `--exclude` | | Exclude files matching pattern | |

## Examples

Show tree of current directory:

```bash
tofu tree
```

Show tree of specific directory:

```bash
tofu tree /path/to/project
```

Limit depth to 2 levels:

```bash
tofu tree -L 2
```

Show hidden files:

```bash
tofu tree -a
```

Exclude certain patterns:

```bash
tofu tree --exclude "node_modules" --exclude ".git"
```

## Sample Output

```
.
├── cmd
│   ├── cat
│   │   └── cat.go
│   └── grep
│       └── grep.go
├── go.mod
├── go.sum
└── main.go

3 directories, 5 files
```
