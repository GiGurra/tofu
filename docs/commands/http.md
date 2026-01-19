# http

Make HTTP requests (like curl).

## Synopsis

```bash
tofu http <url> [flags]
```

## Description

Perform HTTP requests with support for different methods, headers, and data.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--method` | `-X` | HTTP method | `GET` |
| `--headers` | `-H` | Custom headers (can repeat) | |
| `--data` | `-d` | POST data | |
| `--output` | `-o` | Write output to file | |
| `--follow-redirects` | `-L` | Follow redirects | `false` |
| `--verbose` | `-v` | Verbose output | `false` |
| `--insecure` | `-k` | Allow insecure SSL connections | `false` |

## Examples

Simple GET request:

```bash
tofu http https://api.example.com/data
```

POST with data:

```bash
tofu http -X POST -d '{"name":"test"}' https://api.example.com/users
```

POST with data (auto-detects POST method):

```bash
tofu http -d '{"key":"value"}' https://api.example.com
```

Set custom headers:

```bash
tofu http -H "Authorization: Bearer token123" -H "Content-Type: application/json" https://api.example.com
```

Save response to file:

```bash
tofu http -o response.json https://api.example.com/data
```

Follow redirects:

```bash
tofu http -L https://example.com
```

Verbose mode (show headers):

```bash
tofu http -v https://example.com
```

Allow self-signed certificates:

```bash
tofu http -k https://localhost:8443
```

## Verbose Output

```
> GET /api/users HTTP/1.1
> Host: api.example.com
> User-Agent: tofu/http
>
< HTTP/1.1 200 OK
< Content-Type: application/json
< Content-Length: 42
<
{"users": [...]}
```
