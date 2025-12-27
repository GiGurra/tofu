# `tofu ping`

Send ICMP ECHO_REQUEST to network hosts.

## Synopsis

```
tofu ping <host> [flags]
```

## Description

ping uses the ICMP protocol's mandatory ECHO_REQUEST datagram to elicit an ICMP ECHO_RESPONSE from a host or gateway. Requires root/sudo on most systems.

## Options

- `-c, --count int`: Stop after sending count ECHO_REQUEST packets
- `-i, --interval float`: Wait interval seconds between sending each packet (default 1)
- `-W, --timeout float`: Time to wait for a response, in seconds (default 5)
- `-4, --ipv4`: Use IPv4 only
- `-6, --ipv6`: Use IPv6 only

## Examples

Ping a host continuously:

```
sudo tofu ping google.com
```

Ping with a count limit:

```
sudo tofu ping -c 5 google.com
```

Ping with shorter interval:

```
sudo tofu ping -i 0.5 google.com
```

Force IPv4:

```
sudo tofu ping -4 google.com
```

## Notes

- Requires root/sudo privileges on most systems
- Use Ctrl+C to stop continuous pinging
