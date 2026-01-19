# dns

Lookup DNS records.

## Synopsis

```bash
tofu dns <hostname> [flags]
```

## Description

Perform DNS lookups for various record types. Can use the OS resolver or a specific DNS server.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--server` | `-s` | DNS server to use (`os` for system resolver, or IP) | `os` |
| `--types` | `-t` | Record types: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `NS`, `PTR`, `all` | `A,AAAA,CNAME` |
| `--timeout` | | Timeout in seconds | `2` |
| `--json` | `-j` | Output in JSON format | `false` |

## Examples

Basic lookup:

```bash
tofu dns google.com
```

Query all record types:

```bash
tofu dns -t all google.com
```

Query specific record types:

```bash
tofu dns -t MX,TXT gmail.com
```

Use specific DNS server:

```bash
tofu dns -s 8.8.8.8 google.com
```

Use Cloudflare DNS:

```bash
tofu dns -s 1.1.1.1 example.com
```

JSON output:

```bash
tofu dns -j google.com
```

Reverse lookup (PTR):

```bash
tofu dns -t PTR 8.8.8.8
```

## Sample Output

```
Server:  OS
Address: google.com

A Records:
  142.250.80.46

AAAA Records:
  2607:f8b0:4004:800::200e

MX Records:
  10 smtp.google.com.
  20 smtp2.google.com.

TXT Records:
  v=spf1 include:_spf.google.com ~all
```
