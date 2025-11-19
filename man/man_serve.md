# `tofu serve`

Instant static file server.

## Interface

```
> tofu serve --help
Instant static file server

Usage:
  tofu serve [dir] [flags]

Flags:
  -p, --port int          Port to listen on. (default 8080)
      --host string       Host interface to bind to. (default "localhost")
      --spa-mode          Enable Single Page Application mode (redirect 404 to index.html). (default false)
      --no-cache          Disable browser caching. (default false)
  -h, --help              help for serve
```

### Examples

Serve current directory on port 8080:

```
> tofu serve .
```

Serve specific directory on port 3000:

```
> tofu serve ./public -p 3000
```

Serve a Single Page Application (SPA) with 404 redirection:

```
> tofu serve -d ./build --spa-mode
```

Serve without caching (good for development):

```
> tofu serve --no-cache
```

Expose to local network:

```
> tofu serve --host 0.0.0.0
```
