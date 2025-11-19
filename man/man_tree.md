# `tofu tree`

List contents of directories in a tree-like format.

## Interface

```
> tofu tree --help
List contents of directories in a tree-like format

Usage:
  tofu tree [dir] [flags]

Flags:
  -L, --depth int           Descend only level directories deep. (default -1) (use -1 for infinite depth)
  -a, --all                 Do not ignore entries starting with . (default false)
      --ignore-gitignore    Do not ignore files specified in .gitignore. (default false)
      --exclude strings     Exclude files matching the pattern.
  -h, --help                help for tree
```

### Examples

List the current directory with default settings (excludes hidden files):

```
> tofu tree .
.
├── cmd
│   ├── cmd_cat.go
│   ├── cmd_find.go
│   ├── cmd_grep.go
│   ├── cmd_sed2.go
│   ├── cmd_serve.go
│   ├── cmd_watch.go
│   └── common.go
├── go.mod
├── go.sum
├── LICENSE
├── main.go
├── man
│   ├── man_cat.md
│   ├── man_find.md
│   ├── man_grep.md
│   ├── man_sed2.md
│   ├── man_serve.md
│   ├── man_watch.md
│   └── man_uuid.md
├── README.md
└── test
```

List current directory including hidden files and dotfiles:

```
> tofu tree . -a
.
├── .git
│   ├── HEAD
│   └── config
├── .gitignore
└── cmd
    └── cmd_cat.go
...
```

List current directory up to depth 1:

```
> tofu tree . -L 1
.
├── cmd
├── go.mod
├── go.sum
├── LICENSE
├── main.go
├── man
├── README.md
└── test
```

Exclude files matching a pattern:

```
> tofu tree . --exclude "*.mod" --exclude "*.sum"
.
└── cmd
    └── cmd_cat.go
...
```
