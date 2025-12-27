# `tofu mkdir`

Create directories.

## Synopsis

```
tofu mkdir <dirs> [flags]
```

## Description

Create the DIRECTORY(ies), if they do not already exist.

## Options

- `-p, --parents`: Make parent directories as needed, no error if existing
- `-m, --mode string`: Set file mode (as in chmod), not a=rwx - umask (default "0755")
- `-v, --verbose`: Print a message for each created directory

## Examples

Create a directory:

```
tofu mkdir mydir
```

Create nested directories:

```
tofu mkdir -p path/to/nested/dir
```

Create with specific permissions:

```
tofu mkdir -m 0700 private
```

Create with verbose output:

```
tofu mkdir -v newdir
```

Create multiple directories:

```
tofu mkdir dir1 dir2 dir3
```
