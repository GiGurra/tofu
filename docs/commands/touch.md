# touch

Change file access and modification times.

## Synopsis

```bash
tofu touch <files...> [flags]
```

## Description

Update the access and modification times of each FILE to the current time. A FILE argument that does not exist is created empty, unless `-c` is supplied.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--no-create` | `-c` | Do not create any files | `false` |
| `--access-only` | `-a` | Change only the access time | `false` |
| `--modify-only` | `-m` | Change only the modification time | `false` |
| `--reference` | `-r` | Use this file's times instead of current time | |
| `--date` | `-d` | Parse date string and use instead of current time | |

## Examples

Create an empty file or update timestamps:

```bash
tofu touch file.txt
```

Create multiple files:

```bash
tofu touch file1.txt file2.txt file3.txt
```

Don't create if doesn't exist:

```bash
tofu touch -c maybe_exists.txt
```

Use timestamps from another file:

```bash
tofu touch -r reference.txt target.txt
```

Set specific date:

```bash
tofu touch -d "2024-01-15 10:30:00" file.txt
```

Change only access time:

```bash
tofu touch -a file.txt
```

Change only modification time:

```bash
tofu touch -m file.txt
```

## Supported Date Formats

- `2006-01-02T15:04:05` (ISO 8601)
- `2006-01-02 15:04:05`
- `2006-01-02T15:04`
- `2006-01-02 15:04`
- `2006-01-02`
- `Jan 2, 2006`
- `Jan 2 2006`
- `02 Jan 2006`
