# `tofu find`

Finds items on the filesystem with a modern and user-friendly interface.

## Interface

```
> tofu find --help
Find file system items matching a search term

Usage:
  tofu find <search-term> [flags]

Flags:
  -s, --search-type string   Type of search to perform (exact,contains,prefix,suffix,regex). (default "contains")
  -i, --ignore-case          Perform a case-insensitive search. (default false)
  -c, --work-dir string      The working directory to start the search from. (default ".")
  -t, --types strings        Types of file system items to search for (file, dir, all). (default [all])
  -q, --quiet                Suppress error messages. (default false)
  -h, --help                 help for find
```

### Examples

From powershell:

```
> tofu find -i baldur -c / -q
\Program Files (x86)\Steam\steamapps\common\Baldur's Gate Enhanced Edition
\Program Files (x86)\Steam\steamapps\common\Baldur's Gate Enhanced Edition\Baldur.exe
\Program Files (x86)\Steam\steamapps\common\Baldur's Gate II Enhanced Edition
\Program Files (x86)\Steam\steamapps\common\Baldur's Gate II Enhanced Edition\Baldur.exe
\Program Files (x86)\Steam\steamapps\common\Baldurs Gate 3
```

From fish shell on wsl/ubuntu24:

```
> tofu find -i baldur -c / -q
/mnt/c/Program Files (x86)/Steam/steamapps/common/Baldur's Gate Enhanced Edition
/mnt/c/Program Files (x86)/Steam/steamapps/common/Baldur's Gate Enhanced Edition/Baldur.exe
/mnt/c/Program Files (x86)/Steam/steamapps/common/Baldur's Gate II Enhanced Edition
/mnt/c/Program Files (x86)/Steam/steamapps/common/Baldur's Gate II Enhanced Edition/Baldur.exe
/mnt/c/Program Files (x86)/Steam/steamapps/common/Baldurs Gate 3
```