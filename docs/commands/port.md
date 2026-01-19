# port

List or kill processes by port.

## Synopsis

```bash
tofu port [port-number] [flags]
```

## Description

List processes listening on ports, or find and kill the process using a specific port.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--kill` | `-k` | Kill the process listening on the port | `false` |
| `--udp` | `-u` | Include UDP ports | `false` |
| `--all` | `-a` | Show all ports (not just listening) | `false` |

## Examples

List all listening ports:

```bash
tofu port
```

Find process on specific port:

```bash
tofu port 8080
```

Kill process on port:

```bash
tofu port -k 8080
```

Include UDP ports:

```bash
tofu port -u
```

Show all connections (not just listening):

```bash
tofu port -a
```

## Sample Output

```
PROTO   PORT   PID     PROCESS      STATUS   ADDRESS
TCP     22     1234    sshd         LISTEN   0.0.0.0
TCP     80     5678    nginx        LISTEN   0.0.0.0
TCP     8080   9012    node         LISTEN   127.0.0.1
```
