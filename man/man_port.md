# `tofu port`

List or kill processes listening on network ports.

## Interface

```
> tofu port --help
List or kill processes by port

Usage:
  tofu port [port-num] [flags]

Flags:
  -k, --kill              Kill the process listening on the specified port. (default false)
  -u, --udp               Include UDP ports (TCP is default). (default false)
  -a, --all               Show all ports (not just listening). (default false)
  -h, --help              help for port
```

## Details

This tool is a cross-platform alternative to `lsof -i`, `netstat`, or `Get-NetTCPConnection`.
It works on Linux, Windows, and macOS.

### Examples

List all listening TCP ports:

```
> tofu port
PROTO   PORT    PID     PROCESS         STATUS  ADDRESS
TCP     8080    1234    node            LISTEN  ::
TCP     22      890     sshd            LISTEN  0.0.0.0
```

Check who is on port 8080:

```
> tofu port 8080
PROTO   PORT    PID     PROCESS         STATUS  ADDRESS
TCP     8080    1234    node            LISTEN  ::
```

Kill the process on port 8080:

```
> tofu port 8080 --kill
Killing process 1234 (node) on port 8080...
Killed PID 1234
```

Show UDP ports too:

```
> tofu port -u
```
