# `tofu cp`

Copy files and directories.

## Synopsis

```
tofu cp <sources> [flags]
```

## Description

Copy SOURCE to DEST, or multiple SOURCE(s) to DIRECTORY.

## Options

- `-r, --recursive`: Copy directories recursively
- `-f, --force`: If an existing destination file cannot be opened, remove it and try again
- `-i, --interactive`: Prompt before overwriting
- `-n, --no-clobber`: Do not overwrite an existing file
- `-v, --verbose`: Explain what is being done
- `-p, --preserve`: Preserve mode, ownership, and timestamps

## Examples

Copy a file:

```
tofu cp file.txt backup.txt
```

Copy a directory recursively:

```
tofu cp -r src/ dest/
```

Copy with verbose output:

```
tofu cp -v file.txt dest/
```

Copy preserving attributes:

```
tofu cp -p important.txt backup/
```

Interactive copy (prompt before overwrite):

```
tofu cp -i file.txt existing.txt
```
