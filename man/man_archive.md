# `tofu archive`

Create, extract, and list archive files in various formats.

## Synopsis

```
tofu archive <command> [flags]
```

## Commands

- `create` (alias: `c`) - Create an archive from files and directories
- `extract` (alias: `x`) - Extract files from an archive
- `list` (alias: `l`, `ls`) - List contents of an archive

## Supported Formats

| Format | Extension(s) | Create | Extract |
|--------|--------------|--------|---------|
| tar | `.tar` | Yes | Yes |
| tar+gzip | `.tar.gz`, `.tgz` | Yes | Yes |
| tar+bzip2 | `.tar.bz2`, `.tbz2`, `.tbz` | Yes | Yes |
| tar+xz | `.tar.xz`, `.txz` | Yes | Yes |
| tar+zstd | `.tar.zst`, `.tar.zstd` | Yes | Yes |
| tar+lz4 | `.tar.lz4` | Yes | Yes |
| tar+brotli | `.tar.br` | Yes | Yes |
| zip | `.zip` | Yes | Yes |
| 7-Zip | `.7z` | No | Yes |
| RAR | `.rar` | No | Yes |

## Create Command

```
tofu archive create -o <output> [flags] <files...>
```

### Create Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Output archive file name (required) |
| `--format` | `-f` | Archive format (overrides extension detection) |
| `--verbose` | `-v` | List files as they are added |
| `--strip-path` | `-s` | Strip leading path components |

### Create Examples

Create a gzip-compressed tar archive:

```
> tofu archive create -o backup.tar.gz documents/ photos/
```

Create a zip archive with verbose output:

```
> tofu archive create -v -o project.zip src/ README.md LICENSE
a src/
a src/main.go
a src/util.go
a README.md
a LICENSE
```

Create using explicit format (ignores extension):

```
> tofu archive create -f tar.zst -o backup.tar.zst data/
```

Strip paths (files appear at archive root):

```
> tofu archive create -s -o files.tar /path/to/file1.txt /other/path/file2.txt
```

## Extract Command

```
tofu archive extract [flags] <archive>
```

### Extract Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Output directory (default: current directory) |
| `--verbose` | `-v` | List files as they are extracted |

### Extract Examples

Extract to current directory:

```
> tofu archive extract backup.tar.gz
```

Extract to specific directory:

```
> tofu archive extract -o /tmp/restore backup.tar.gz
```

Extract with verbose output:

```
> tofu archive extract -v project.zip
x src/
x src/main.go
x src/util.go
x README.md
```

Extract a 7z archive:

```
> tofu archive extract archive.7z
```

## List Command

```
tofu archive list [flags] <archive>
```

### List Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--long` | `-l` | Long listing format (show size and permissions) |

### List Examples

List archive contents:

```
> tofu archive list backup.tar.gz
src/
src/main.go
src/util.go
README.md
```

Long format listing:

```
> tofu archive list -l backup.tar.gz
drwxr-xr-x          0  src/
-rw-r--r--       1234  src/main.go
-rw-r--r--        567  src/util.go
-rw-r--r--        890  README.md
```

## Format Detection

The archive format is automatically detected:

- **When creating**: Based on the output filename extension
- **When extracting/listing**: Based on file contents (magic bytes)

You can override format detection when creating by using the `--format` flag.

## Comparison with Traditional Tools

| Traditional | tofu archive | Description |
|------------|--------------|-------------|
| `tar czf a.tar.gz dir/` | `tofu archive create -o a.tar.gz dir/` | Create tar.gz |
| `tar xzf a.tar.gz` | `tofu archive extract a.tar.gz` | Extract tar.gz |
| `tar tzf a.tar.gz` | `tofu archive list a.tar.gz` | List tar.gz |
| `zip -r a.zip dir/` | `tofu archive create -o a.zip dir/` | Create zip |
| `unzip a.zip` | `tofu archive extract a.zip` | Extract zip |
| `unzip -l a.zip` | `tofu archive list a.zip` | List zip |
| `7z x a.7z` | `tofu archive extract a.7z` | Extract 7z |

## Notes

- 7-Zip and RAR formats support extraction only (not creation)
- The format auto-detection reads file headers, so renamed files still work
- Symbolic links are preserved when the archive format supports them
- File permissions are preserved on Unix systems
