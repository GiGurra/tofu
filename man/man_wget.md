# `tofu wget`

Download files from the web with progress indication and resume support.

## Synopsis

```
tofu wget [flags] <url>
```

## Description

Download files from URLs. Features automatic filename detection from URLs, a progress bar showing download status, and the ability to resume partially downloaded files.

By default, redirects are followed automatically and the output filename is derived from the URL.

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-O` | Write output to specified file (use `-` for stdout) |
| `--continue` | `-c` | Resume a partially downloaded file |
| `--quiet` | `-q` | Quiet mode - no progress output |
| `--no-progress` | | Disable progress bar but show other output |
| `--insecure` | `-k` | Allow insecure server connections (skip TLS verification) |
| `--timeout` | `-T` | Set timeout in seconds (default: 30) |
| `--retries` | `-t` | Set number of retries (default: 3, use 0 for infinite) |
| `--user-agent` | `-U` | Set custom User-Agent header |
| `--header` | `-H` | Add custom header(s) |

## Examples

Download a file (auto-detect filename):

```
> tofu wget https://example.com/file.zip
Downloading: https://example.com/file.zip
Saving to: file.zip
[==============================] 100.0% 1.2 MB/s 10.5 MB/10.5 MB
Downloaded: file.zip (10.5 MB)
```

Download with custom output filename:

```
> tofu wget -O myfile.zip https://example.com/file.zip
```

Resume a partially downloaded file:

```
> tofu wget -c https://example.com/large-file.iso
Downloading: https://example.com/large-file.iso
Resuming from byte 524288000
Saving to: large-file.iso
[==============================] 100.0% 2.5 MB/s 1.0 GB/1.0 GB
Downloaded: large-file.iso (1.0 GB)
```

Quiet mode (no output except errors):

```
> tofu wget -q https://example.com/file.txt
```

Download to stdout:

```
> tofu wget -O - https://example.com/data.json | jq .
```

Download with custom headers:

```
> tofu wget -H "Authorization: Bearer token123" https://api.example.com/file.zip
```

Skip TLS certificate verification:

```
> tofu wget -k https://self-signed.example.com/file.txt
```

Set custom timeout and retries:

```
> tofu wget -T 60 -t 5 https://slow-server.example.com/large-file.zip
```

## Progress Bar

The progress bar shows:
- Visual progress indicator
- Percentage complete (when file size is known)
- Download speed
- Downloaded bytes / Total bytes

When the server doesn't provide Content-Length, the progress shows downloaded bytes and speed without percentage.

## Resume Support

The `-c` flag enables resume support for partially downloaded files:

1. If the output file already exists, wget sends a `Range` header
2. The server responds with `206 Partial Content` if it supports ranges
3. Download continues from where it left off
4. If the server doesn't support ranges, download starts from the beginning

## Comparison with GNU wget

| GNU wget | tofu wget | Description |
|----------|-----------|-------------|
| `wget URL` | `tofu wget URL` | Basic download |
| `wget -O file URL` | `tofu wget -O file URL` | Custom output file |
| `wget -c URL` | `tofu wget -c URL` | Resume download |
| `wget -q URL` | `tofu wget -q URL` | Quiet mode |
| `wget --header "H: V"` | `tofu wget -H "H: V"` | Custom header |
| `wget -U agent` | `tofu wget -U agent` | Custom User-Agent |
| `wget --no-check-certificate` | `tofu wget -k` | Skip TLS verification |

## Exit Codes

- `0` - Download successful
- `1` - Error occurred (network error, file not found, etc.)

## Notes

- HTTPS URLs are used by default when no scheme is specified
- Redirects are followed automatically (up to 10 redirects)
- The User-Agent defaults to `tofu/wget` if not specified
- Retries use exponential backoff (1s, 2s, 3s, etc.)
