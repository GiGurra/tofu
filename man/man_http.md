# `tofu http`

Make HTTP requests (like curl).

## Interface

```
> tofu http --help
Make HTTP requests (like curl)

Usage:
  tofu http [url] [flags]

Flags:
  -X, --method string       HTTP method to use (GET, POST, PUT, DELETE, etc.). Default is GET. (default "GET")
  -H, --headers strings     Pass custom header(s) to server.
  -d, --data string         HTTP POST data.
  -o, --output-file string  Write to file instead of stdout.
  -L, --follow-redirects    Follow redirects. (default false)
  -v, --verbose             Make the operation more talkative. (default false)
  -k, --insecure            Allow insecure server connections when using SSL. (default false)
  -h, --help                help for http
```

### Examples

```
> tofu http https://example.com
... content of example.com ...

> tofu http https://example.com -v
> GET / HTTP/1.1
> User-Agent: tofu/http
>
< HTTP/1.1 200 OK
< ...
... content ...
```
