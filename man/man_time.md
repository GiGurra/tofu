# `tofu time`

Show current time or parse a timestamp.

## Interface

```
> tofu time --help
Show current time or parse a timestamp

Usage:
  tofu time [timestamp] [flags]

Flags:
  -f, --format string  Explicit input format (e.g. '2006-01-02' or 'unix', 'unixmilli').
  -u, --utc            Show output in UTC only (suppress Local).
  -h, --help           help for time
```

## Description

Display the current time in various formats, or parse a provided timestamp.

If no argument is provided, the current system time is displayed.
If an argument is provided, it attempts to parse it as:
1. Unix timestamp (seconds, milliseconds, or nanoseconds).
2. Standard date/time formats (RFC3339, RFC1123, DateOnly, DateTime, etc.).

You can force a specific input format using the `--format` / `-f` flag.
Supported special format names: `unix`, `unixmilli`, `unixmicro`, `unixnano`.
Otherwise, provide a Go reference time layout (e.g. `2006-01-02 15:04`).

## Examples

Show current time:

```
> tofu time
Local:      2023-10-27 10:00:00.000 +0200 CEST
UTC:        2023-10-27 08:00:00.000 +0000 UTC
Unix:       1698393600
UnixMilli:  1698393600000
RFC3339:    2023-10-27T10:00:00+02:00
ISO8601:    2023-10-27T10:00:00.000Z+02:00
```

Parse a Unix timestamp:

```
> tofu time 1698393600
Local:      2023-10-27 10:00:00.000 +0200 CEST
...
```

Parse with explicit format:

```
> tofu time "27/10/2023" -f "02/01/2006"
...
```

Force unix milliseconds interpretation (even if number looks small):

```
> tofu time 123456789 -f unixmilli
```

Parse an ISO8601 string:

```
> tofu time "2023-10-27T10:00:00Z"
...
```

Parse a simpler date string:

```
> tofu time "2023-10-27 10:00"
...
```
