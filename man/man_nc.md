# `tofu nc`

Netcat clone: Connect to or listen on sockets (TCP/UDP).

## Interface

```
> tofu nc --help
Netcat clone: Connect to or listen on sockets (TCP/UDP)

Usage:
  tofu nc [args] [flags]

Flags:
  -l, --listen      Listen mode, for inbound connects. (default false)
  -u, --udp         Use UDP instead of default TCP. (default false)
  -v, --verbose     Verbose mode. (default false)
  -h, --help        help for nc
```

### Examples

**Client Mode:**
Connect to localhost port 8080:
```
> tofu nc localhost 8080
```

**Server Mode:**
Listen on port 8080:
```
> tofu nc -l 8080
```

**File Transfer:**
Send file:
```
> tofu cat file.txt | tofu nc localhost 8080
```

Receive file:
```
> tofu nc -l 8080 > file_copy.txt
```

**Chat:**
Server:
```
> tofu nc -l 8080
```
Client:
```
> tofu nc localhost 8080
```
Then type messages back and forth.
