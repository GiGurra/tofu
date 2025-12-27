# `tofu rmdir`

Remove empty directories.

## Synopsis

```
tofu rmdir <dirs> [flags]
```

## Description

Remove the DIRECTORY(ies), if they are empty.

## Options

- `-p, --parents`: Remove DIRECTORY and its ancestors
- `-v, --verbose`: Output a diagnostic for every directory processed

## Examples

Remove an empty directory:

```
tofu rmdir emptydir
```

Remove directory and empty parents:

```
tofu rmdir -p path/to/empty/dir
```

Verbose removal:

```
tofu rmdir -v emptydir
```

Remove multiple empty directories:

```
tofu rmdir dir1 dir2 dir3
```
