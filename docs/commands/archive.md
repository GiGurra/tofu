# archive

Create and extract archive files.

## Synopsis

```bash
tofu archive create -o <output> <files...> [flags]
tofu archive extract <archive> [flags]
tofu archive list <archive> [flags]
```

## Description

Create, extract, and list archive files in various formats including tar, gzip, bzip2, xz, zstd, zip, 7z, and rar.

## Supported Formats

| Format | Extension | Create | Extract | Password |
|--------|-----------|--------|---------|----------|
| tar | .tar | Yes | Yes | No |
| tar.gz | .tar.gz, .tgz | Yes | Yes | No |
| tar.bz2 | .tar.bz2, .tbz2 | Yes | Yes | No |
| tar.xz | .tar.xz, .txz | Yes | Yes | No |
| tar.zst | .tar.zst | Yes | Yes | No |
| tar.lz4 | .tar.lz4 | Yes | Yes | No |
| zip | .zip | Yes | Yes | Yes (AES) |
| 7z | .7z | No | Yes | Yes |
| rar | .rar | No | Yes | Yes |

## Commands

### create

Create an archive from files and directories.

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output archive file name | (required) |
| `--verbose` | `-v` | List files as they are added | `false` |
| `--format` | `-f` | Archive format (overrides extension) | |
| `--password` | `-p` | Password for encrypted ZIP | |
| `--encryption` | `-e` | ZIP encryption: `legacy`, `aes128`, `aes192`, `aes256` | `aes256` |

### extract

Extract files from an archive.

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output directory | `.` |
| `--verbose` | `-v` | List files as extracted | `false` |
| `--password` | `-p` | Password for encrypted archives | |

### list

List contents of an archive.

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--long` | `-l` | Long listing format | `false` |
| `--password` | `-p` | Password for encrypted archives | |

## Examples

Create a tar.gz archive:

```bash
tofu archive create -o backup.tar.gz file1.txt dir1/
```

Create a zip archive:

```bash
tofu archive create -o project.zip src/ README.md
```

Create encrypted zip:

```bash
tofu archive create -o secret.zip -p mypassword file.txt
```

Create with specific format:

```bash
tofu archive create -f tar.zst -o backup.tar.zst data/
```

Extract archive:

```bash
tofu archive extract backup.tar.gz
```

Extract to specific directory:

```bash
tofu archive extract -o /tmp/output project.zip
```

Extract encrypted archive:

```bash
tofu archive extract -p mypassword secret.zip
```

List archive contents:

```bash
tofu archive list backup.tar.gz
```

Long listing:

```bash
tofu archive list -l project.zip
```

## Aliases

- `tofu archive c` - alias for `create`
- `tofu archive x` - alias for `extract`
- `tofu archive l` or `ls` - alias for `list`
