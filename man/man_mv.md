# `tofu mv`

Move (rename) files.

## Synopsis

```
tofu mv <sources> [flags]
```

## Description

Rename SOURCE to DEST, or move SOURCE(s) to DIRECTORY.

## Options

- `-f, --force`: Do not prompt before overwriting
- `-i, --interactive`: Prompt before overwriting
- `-n, --no-clobber`: Do not overwrite an existing file
- `-v, --verbose`: Explain what is being done

## Examples

Rename a file:

```
tofu mv old.txt new.txt
```

Move file to directory:

```
tofu mv file.txt dest/
```

Move multiple files to directory:

```
tofu mv file1.txt file2.txt dest/
```

Move with verbose output:

```
tofu mv -v file.txt dest/
```

Interactive move (prompt before overwrite):

```
tofu mv -i file.txt existing.txt
```
