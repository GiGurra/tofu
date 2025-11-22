# `tofu ip`

Show local and public IP addresses.

## Interface

```
> tofu ip --help
Show local and public IP addresses

Usage:
  tofu ip [flags]

Flags:
  -l, --local-only   Only show local interfaces, do not attempt to discover public IP. (default false)
  -h, --help         help for ip
```

### Examples

**Show all IPs (local and public)**

```
> tofu ip
Local Interfaces:
  lo:
    127.0.0.1/8
    ::1/128
  eth0:
    192.168.1.100/24
    fe80::a00:27ff:fe4e:66a1/64

Public IP:
  203.0.113.42
```

**Show only local interfaces**

```
> tofu ip --local-only
Local Interfaces:
  lo:
    127.0.0.1/8
    ::1/128
  eth0:
    192.168.1.100/24
    fe80::a00:27ff:fe4e:66a1/64
```
