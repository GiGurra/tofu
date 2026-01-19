# ping

Send ICMP ECHO_REQUEST to network hosts.

## Synopsis

```bash
tofu ping <host> [flags]
```

## Description

Send ICMP echo request packets to network hosts. Requires root/sudo privileges on most systems.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--count` | `-c` | Stop after sending N packets (0 for infinite) | `0` |
| `--interval` | `-i` | Wait N seconds between packets | `1` |
| `--timeout` | `-W` | Time to wait for response in seconds | `5` |
| `--ipv4` | `-4` | Use IPv4 only | `false` |
| `--ipv6` | `-6` | Use IPv6 only | `false` |

## Examples

Ping a host continuously:

```bash
sudo tofu ping google.com
```

Ping with limited count:

```bash
sudo tofu ping -c 5 google.com
```

Faster ping (0.2 second interval):

```bash
sudo tofu ping -i 0.2 google.com
```

Force IPv4:

```bash
sudo tofu ping -4 google.com
```

Force IPv6:

```bash
sudo tofu ping -6 google.com
```

## Sample Output

```
PING google.com (142.250.80.46): 56 data bytes
64 bytes from 142.250.80.46: icmp_seq=0 time=15.234 ms
64 bytes from 142.250.80.46: icmp_seq=1 time=14.567 ms
64 bytes from 142.250.80.46: icmp_seq=2 time=16.891 ms
^C
--- google.com ping statistics ---
3 packets transmitted, 3 packets received, 0.0% packet loss
round-trip min/avg/max = 14.567/15.564/16.891 ms
```

## Notes

- Requires root/sudo on most Unix systems due to raw socket requirements
- Press Ctrl+C to stop and see statistics
