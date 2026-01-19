# uuid

Generate UUIDs.

## Synopsis

```bash
tofu uuid [flags]
```

## Description

Generate Universally Unique Identifiers (UUIDs) of various versions.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--count` | `-n` | Number of UUIDs to generate | `1` |
| `--version` | `-v` | UUID version (1, 3, 4, 5, 6, 7) | `4` |
| `--namespace` | `-s` | Namespace for v3/v5: `dns`, `url`, `oid`, `x500`, or UUID | |
| `--name` | `-d` | Name/data for v3/v5 | |

## Examples

Generate a random UUID (v4):

```bash
tofu uuid
```

Generate multiple UUIDs:

```bash
tofu uuid -n 5
```

Generate UUID v1 (time-based):

```bash
tofu uuid -v 1
```

Generate UUID v7 (time-ordered random):

```bash
tofu uuid -v 7
```

Generate UUID v5 (SHA-1 namespace):

```bash
tofu uuid -v 5 -s dns -d "example.com"
```

Generate UUID v3 (MD5 namespace):

```bash
tofu uuid -v 3 -s url -d "https://example.com"
```

Use custom namespace:

```bash
tofu uuid -v 5 -s "6ba7b810-9dad-11d1-80b4-00c04fd430c8" -d "mydata"
```

## UUID Versions

| Version | Description |
|---------|-------------|
| 1 | Time-based (MAC address + timestamp) |
| 3 | MD5 hash of namespace + name |
| 4 | Random (most common) |
| 5 | SHA-1 hash of namespace + name |
| 6 | Reordered time-based (sortable) |
| 7 | Unix timestamp + random (sortable, recommended) |

## Sample Output

```
f47ac10b-58cc-4372-a567-0e02b2c3d479
```
