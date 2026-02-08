# proxy

TCP proxy/port forwarder.

## Synopsis

```bash
tofu proxy <listen-addr> <target-addr> [flags]
```

## Description

Forward TCP connections from a listen address to a target address. Each incoming connection is proxied bidirectionally to the target, with proper half-close handling.

Supports connect timeouts, idle timeouts, automatic retries with configurable intervals, connection limiting, and verbose transfer statistics.

Useful for exposing WSL services on Windows LAN interfaces, or any TCP port forwarding scenario.

## Arguments

| Argument | Description |
|----------|-------------|
| `listen-addr` | Address to listen on (e.g. `0.0.0.0:8443`) |
| `target-addr` | Address to forward to (e.g. `localhost:8443`) |

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--connect-timeout` | `-t` | Connect timeout in ms (0=no timeout) | `5000` |
| `--idle-timeout` | `-i` | Idle timeout in ms, close if no data (0=no timeout) | `0` |
| `--retries` | `-r` | Connection retries (-1=infinite, 0=no retry) | `0` |
| `--retry-interval` | | Retry interval in ms | `1000` |
| `--max-conns` | `-m` | Max concurrent connections (0=unlimited) | `0` |
| `--verbose` | `-v` | Verbose logging | `false` |

## Examples

Expose a local service on all interfaces:

```bash
tofu proxy 0.0.0.0:8080 localhost:8080
```

Forward WSL service to Windows LAN:

```bash
tofu proxy 0.0.0.0:8443 localhost:8443
```

With timeouts and retries (10s connect, 60s idle, 3 retries):

```bash
tofu proxy -t 10000 -i 60000 -r 3 0.0.0.0:8443 localhost:8443
```

Limit to 10 concurrent connections with verbose output:

```bash
tofu proxy -m 10 -v 0.0.0.0:8080 localhost:8080
```

Retry indefinitely until target is available (1s interval):

```bash
tofu proxy -r -1 --retry-interval 1000 0.0.0.0:3000 localhost:3000
```

## Sample Output

```
Proxying 0.0.0.0:8443 -> localhost:8443
[1] 192.168.1.42:51234 connected (active: 1)
[1] disconnected (active: 0)
[2] 192.168.1.42:51240 connected (active: 1)
[2] disconnected (active: 0)
```

Verbose mode:

```
Proxying 0.0.0.0:8443 -> localhost:8443
  connect-timeout: 5000ms, idle-timeout: 60000ms, retries: 3, retry-interval: 1000ms, max-conns: 0
[1] 192.168.1.42:51234 connected (active: 1)
[1] sent 1.2 KB, received 45.3 KB
[1] disconnected after 3420ms (active: 0)
```
