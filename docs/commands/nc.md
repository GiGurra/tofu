# nc

Netcat clone - Connect to or listen on sockets.

## Synopsis

```bash
tofu nc <host> <port> [flags]
tofu nc -l <port> [flags]
```

## Description

A netcat clone for TCP/UDP connections. Can act as a client connecting to a server, or as a server listening for connections.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--listen` | `-l` | Listen mode for inbound connections | `false` |
| `--udp` | `-u` | Use UDP instead of TCP | `false` |
| `--verbose` | `-v` | Verbose mode | `false` |

## Examples

Connect to a server:

```bash
tofu nc localhost 8080
```

Start a TCP server:

```bash
tofu nc -l 8080
```

Use UDP:

```bash
tofu nc -u localhost 5353
```

Simple chat between two terminals:

Terminal 1 (server):
```bash
tofu nc -l 9000
```

Terminal 2 (client):
```bash
tofu nc localhost 9000
```

Port scanner (connect test):

```bash
echo "" | tofu nc -v localhost 22
```

Send data to server:

```bash
echo "Hello, server!" | tofu nc localhost 8080
```

Receive data and save to file:

```bash
tofu nc -l 9000 > received.txt
```

Transfer a file:

Receiver:
```bash
tofu nc -l 9000 > file.txt
```

Sender:
```bash
tofu nc localhost 9000 < file.txt
```
