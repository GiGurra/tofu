# `tofu touch`

Change file access and modification times.

## Synopsis

```
tofu touch <files> [flags]
```

## Description

Update the access and modification times of each FILE to the current time. A FILE argument that does not exist is created empty, unless -c is supplied.

## Options

- `-c, --no-create`: Do not create any files
- `-a, --access-only`: Change only the access time
- `-m, --modify-only`: Change only the modification time
- `-r, --reference string`: Use this file's times instead of current time
- `-d, --date string`: Parse date string and use it instead of current time

## Examples

Create a new empty file or update timestamps:

```
tofu touch file.txt
```

Create multiple files:

```
tofu touch file1.txt file2.txt file3.txt
```

Update only if file exists:

```
tofu touch -c maybe-exists.txt
```

Use another file's timestamp:

```
tofu touch -r reference.txt target.txt
```

Set specific date:

```
tofu touch -d "2024-01-15 10:30:00" file.txt
```

Change only access time:

```
tofu touch -a file.txt
```
