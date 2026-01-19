# serve

Instant static file server.

## Synopsis

```bash
tofu serve [directory] [flags]
```

## Description

Start an HTTP server to serve static files from a directory. Useful for development and quick file sharing.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--port` | `-p` | Port to listen on | `8080` |
| `--host` | | Host interface to bind to | `localhost` |
| `--spa-mode` | | Enable SPA mode (redirect 404 to index.html) | `false` |
| `--no-cache` | | Disable browser caching | `false` |
| `--read-timeout-millis` | | Max duration for reading request (ms) | `5000` |
| `--write-timeout-millis` | | Max duration for writing response (ms) | `10000` |
| `--idle-timeout-millis` | | Max idle time for keep-alive (ms) | `120000` |
| `--max-header-bytes` | | Max bytes for request headers | `1048576` |

## Examples

Serve current directory on port 8080:

```bash
tofu serve
```

Serve specific directory:

```bash
tofu serve /path/to/files
```

Serve on different port:

```bash
tofu serve -p 3000
```

Serve on all interfaces:

```bash
tofu serve --host 0.0.0.0
```

Enable SPA mode (for React/Vue/Angular apps):

```bash
tofu serve --spa-mode ./dist
```

Disable caching for development:

```bash
tofu serve --no-cache
```

## Output

```
Serving /path/to/files at http://localhost:8080
[200] GET /index.html (1.234ms)
[200] GET /styles.css (0.567ms)
[404] GET /missing.html (0.123ms)
```
