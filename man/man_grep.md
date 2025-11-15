# `tofu grep`

Like gnu grep, but suckier - although, it works in powershell.

## Interface

```
> tofu grep --help
Search for patterns in files

Usage:
  tofu grep <pattern> [files] [flags]

Flags:
  -t, --pattern-type string   Type of pattern matching (basic,extended,fixed,perl). (default "extended")
  -i, --ignore-case           Perform case-insensitive matching. (default false)
  -v, --invert-match          Select non-matching lines. (default false)
  -w, --word-regexp           Match only whole words. (default false)
  -x, --line-regexp           Match only whole lines. (default false)
  -n, --line-number           Print line number with output lines. (default false)
  -H, --with-filename         Print filename with output lines. (default false)
      --no-filename           Suppress filename prefix on output. (default false)
  -c, --count                 Print only a count of matching lines per file. (default false)
  -l, --files-with-match      Print only names of files with matches. (default false)
  -L, --files-without-match   Print only names of files without matches. (default false)
  -o, --only-matching         Show only the matched parts of lines. (default false)
  -q, --quiet                 Suppress all normal output. (default false)
      --ignore-binary         Suppress output for binary files. (default false)
  -m, --max-count int         Stop after NUM matches per file.
  -B, --before-context int    Print NUM lines of leading context.
  -A, --after-context int     Print NUM lines of trailing context.
  -C, --context int           Print NUM lines of output context.
  -r, --recursive             Search directories recursively. (default false)
  -I, --include strings       Search only files matching pattern (glob).
  -e, --exclude strings       Skip files matching pattern (glob).
      --exclude-dir strings   Skip directories matching pattern (glob).
  -s, --no-messages           Suppress error messages. (default false)
  -h, --help                  help for grep
```

```
> tofu grep -ir tofu . --ignore-binary
.git/FETCH_HEAD:5b78fa752023582da92cf39f0628ac487d95c0d0                branch 'main' of github.com:gigurra/tofu
.git/config:    url = git@github.com:gigurra/tofu
.git/logs/HEAD:9aff1e1b4758809e547f1cedacce460c776e73c3 51e8912bdd4b764fd2c1d1b728500c229ac68440 Johan Kjölhede <johan.kjolhede@gmail.com> 1762637551 +0100 commit: fix tofu find
.git/logs/refs/heads/main:9aff1e1b4758809e547f1cedacce460c776e73c3 51e8912bdd4b764fd2c1d1b728500c229ac68440 Johan Kjölhede <johan.kjolhede@gmail.com> 1762637551 +0100      commit: fix tofu find
.idea/modules.xml:      <module fileurl="file://$PROJECT_DIR$/.idea/tofu.iml" filepath="$PROJECT_DIR$/.idea/tofu.iml" />
.idea/workspace.xml:    &quot;last_opened_file_path&quot;: &quot;/home/gigur/git/tofu/cmd&quot;,
.idea/workspace.xml:    <task id="LOCAL-00017" summary="fix tofu find">
.idea/workspace.xml:    <MESSAGE value="fix tofu find" />
README.md:# �豆腐 tofu
README.md:**tofu** is a modern toolkit of CLI utilities that work seamlessly across Windows, macOS, and Linux. Like its namesake,
README.md:tofu adapts to any environment while providing essential nutrition for your daily command-line workflows.
README.md:**tofu** reimagines these essential tools with:
README.md:### tofu find
README.md:- `tofu grep` - search that doesn't make you look up regex every time
README.md:- `tofu watch` - file watching done right
README.md:- `tofu serve` - instant static file server
README.md:go install github.com/gigurra/tofu@latest
README.md:tofu completions <shell> [--help]
README.md:## Why "tofu"?
README.md:Like tofu, these tools are:
README.md:Plus, it's fun to say you're "serving fresh tofu" when you ship a new tool. 🍴
README.md:🚧 **Early Development** - Currently building the first tool (`tofu find`). Watch this space!
go.mod:module github.com/gigurra/tofu
main.go:        "github.com/gigurra/tofu/cmd"
main.go:                Use:     "tofu",
man/man_find.md:# `tofu find`
man/man_find.md:> tofu find --help
man/man_find.md:  tofu find <search-term> [flags]
man/man_find.md:> tofu find -i baldur -c / -q
man/man_find.md:> tofu find -i baldur -c / -q
```