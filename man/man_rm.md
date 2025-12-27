# `tofu rm`

Remove files or directories.

## Synopsis

```
tofu rm <files> [flags]
```

## Description

Remove (unlink) the FILE(s).

## Options

- `-r, --recursive`: Remove directories and their contents recursively
- `-f, --force`: Ignore nonexistent files and arguments, never prompt
- `-i, --interactive`: Prompt before every removal
- `-d, --dir`: Remove empty directories
- `-v, --verbose`: Explain what is being done

## Examples

Remove a file:

```
tofu rm file.txt
```

Remove multiple files:

```
tofu rm file1.txt file2.txt
```

Remove a directory recursively:

```
tofu rm -r mydir/
```

Force remove without prompts:

```
tofu rm -rf mydir/
```

Interactive removal:

```
tofu rm -i file.txt
```

Verbose removal:

```
tofu rm -rv mydir/
```
