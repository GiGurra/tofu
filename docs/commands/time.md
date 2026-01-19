# time

Show current time or parse a timestamp.

## Synopsis

```bash
tofu time [timestamp] [flags]
```

## Description

Display the current time in various formats, or parse a provided timestamp. Supports Unix timestamps and common date/time formats.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--format` | `-f` | Explicit input format (e.g., `2006-01-02`, `unix`, `unixmilli`) | |
| `--utc` | `-u` | Show output in UTC only (suppress local time) | `false` |

## Examples

Show current time:

```bash
tofu time
```

Parse Unix timestamp:

```bash
tofu time 1698393600
```

Parse Unix milliseconds:

```bash
tofu time 1698393600000
```

Parse ISO 8601 date:

```bash
tofu time "2023-10-27T10:00:00Z"
```

Parse date only:

```bash
tofu time 2023-10-27
```

Parse with custom format:

```bash
tofu time "27/10/2023" -f "02/01/2006"
```

Output in UTC only:

```bash
tofu time -u
```

## Supported Input Formats

The tool auto-detects these formats:

- Unix timestamp (seconds, milliseconds, microseconds, nanoseconds)
- RFC3339: `2006-01-02T15:04:05Z07:00`
- Date and time: `2006-01-02 15:04:05`
- Date only: `2006-01-02`
- RFC1123: `Mon, 02 Jan 2006 15:04:05 MST`

Special format names for `-f`:

- `unix` - Unix seconds
- `unixmilli` - Unix milliseconds
- `unixmicro` - Unix microseconds
- `unixnano` - Unix nanoseconds

## Sample Output

```
Local:      2023-10-27 12:00:00.000 +0200 CEST
UTC:        2023-10-27 10:00:00.000 +0000 UTC
Unix:       1698393600
UnixMilli:  1698393600000
RFC3339:    2023-10-27T10:00:00Z
ISO8601:    2023-10-27T10:00:00.000Z
```
