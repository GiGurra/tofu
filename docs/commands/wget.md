# wget

Download files from the web.

## Synopsis

```bash
tofu wget <url> [flags]
```

## Description

Download files from URLs with progress indication and resume support.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-O` | Write output to file (use `-` for stdout) | |
| `--continue` | `-c` | Resume partially downloaded file | `false` |
| `--quiet` | `-q` | Quiet mode - no progress output | `false` |
| `--no-progress` | | Disable progress bar but show other output | `false` |
| `--insecure` | `-k` | Allow insecure SSL connections | `false` |
| `--timeout` | `-T` | Timeout in seconds | `30` |
| `--retries` | `-t` | Number of retries (0 for infinite) | `3` |
| `--user-agent` | `-U` | Custom User-Agent header | |
| `--headers` | `-H` | Custom headers (can repeat) | |

## Examples

Download a file:

```bash
tofu wget https://example.com/file.zip
```

Download with custom output name:

```bash
tofu wget -O myfile.zip https://example.com/file.zip
```

Resume a partial download:

```bash
tofu wget -c https://example.com/large-file.iso
```

Quiet mode:

```bash
tofu wget -q https://example.com/file.txt
```

Download to stdout:

```bash
tofu wget -O - https://example.com/data.json
```

Custom headers:

```bash
tofu wget -H "Authorization: Bearer token" https://api.example.com/file
```

Allow self-signed certificates:

```bash
tofu wget -k https://localhost:8443/file.txt
```

## Sample Output

```
Downloading: https://example.com/file.zip
Saving to: file.zip
[============================  ]  85.3%  1.2 MB/s  45.6MB/53.4MB
Downloaded: file.zip (53.4 MB)
```

## Features

- Auto-detect output filename from URL
- Progress bar with speed and ETA
- Resume partially downloaded files with `-c`
- Follow redirects automatically
- Retry on connection failures
