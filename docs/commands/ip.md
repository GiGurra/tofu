# ip

Show local and public IP addresses.

## Synopsis

```bash
tofu ip [flags]
```

## Description

Display network interface information including local IP addresses, public IP, DNS servers, and default gateway.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--local-only` | `-l` | Only show local interfaces, skip public IP lookup | `false` |
| `--json` | `-j` | Output in JSON format | `false` |

## Examples

Show all network info:

```bash
tofu ip
```

Only local interfaces:

```bash
tofu ip -l
```

JSON output:

```bash
tofu ip -j
```

## Sample Output

```
Local Interfaces:
  lo0:
    127.0.0.1/8
    ::1/128
  en0:
    192.168.1.100/24
    fe80::1234:5678:abcd:ef01/64

Public IP:
  203.0.113.42

DNS Servers:
  8.8.8.8
  8.8.4.4

Default Gateway:
  192.168.1.1
```

## JSON Output

```json
{
  "interfaces": {
    "en0": ["192.168.1.100/24", "fe80::1234:5678:abcd:ef01/64"],
    "lo0": ["127.0.0.1/8", "::1/128"]
  },
  "public_ip": "203.0.113.42",
  "dns_servers": ["8.8.8.8", "8.8.4.4"],
  "gateways": ["192.168.1.1"]
}
```
