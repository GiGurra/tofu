# k8s

Kubernetes utilities.

## Synopsis

```bash
tofu k8s <subcommand> [flags]
```

## Subcommands

- [`port-forward`](#port-forward) - Port-forward to pods with auto-reconnect
- [`tail pods`](#tail-pods) - Tail logs from Kubernetes pods

---

## port-forward

Port-forward to a running pod from a deployment, statefulset, daemonset, or service. Automatically reconnects when the connection is lost or the pod terminates.

### Synopsis

```bash
tofu k8s port-forward [flags] <ports...>
```

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--from-deploy` | `-d` | Deployment name to port-forward to | |
| `--from-sts` | | StatefulSet name to port-forward to | |
| `--from-ds` | | DaemonSet name to port-forward to | |
| `--from-svc` | | Service name to port-forward to | |
| `--namespace` | `-n` | Kubernetes namespace | current context |
| `--keepalive` | `-k` | Keep trying to reconnect when connection is lost | `true` |

!!! note
    Exactly one of `--from-deploy`, `--from-sts`, `--from-ds`, or `--from-svc` must be specified.

### Port Format

Ports can be specified as:

- `local:remote` - Forward local port to remote port (e.g., `8080:80`)
- `remote` - Use same port for local and remote (e.g., `80`)

Multiple ports can be forwarded simultaneously.

### Examples

Port-forward from a deployment:

```bash
tofu k8s port-forward -d nginx 8080:80
```

Port-forward from a StatefulSet:

```bash
tofu k8s port-forward --from-sts redis 6379
```

Port-forward from a DaemonSet:

```bash
tofu k8s port-forward --from-ds fluentd 24224:24224
```

Port-forward from a Service:

```bash
tofu k8s port-forward --from-svc my-service 8080:80
```

Multiple ports:

```bash
tofu k8s port-forward -d my-app 8080:80 9090:9090 5432:5432
```

Specific namespace:

```bash
tofu k8s port-forward -d api-server -n production 8080:80
```

Disable auto-reconnect:

```bash
tofu k8s port-forward -d nginx --keepalive=false 8080:80
```

### Features

- **Auto-reconnect**: Automatically finds a new pod and reconnects when:
    - The current pod is deleted or crashes
    - A deployment/statefulset rolls out new pods
    - The connection is lost for any reason

- **Proactive monitoring**: Checks pod status every 2 seconds and reconnects before the connection fails (no need to wait for traffic to discover dead connections)

- **Quick reconnect backoff**: If connection is lost within 5 seconds, waits 1 second before reconnecting to avoid rapid reconnection loops

- **Shell completion**: Tab completion for resource names and namespaces

### Output

```
Port-forwarding to pod nginx-abc123-xyz (ports: 8080:80)
Forwarding from 127.0.0.1:8080 -> 80
Pod nginx-abc123-xyz is no longer running, triggering reconnect...
Connection lost quickly, waiting 1s before reconnecting...
Port-forwarding to pod nginx-def456-uvw (ports: 8080:80)
Forwarding from 127.0.0.1:8080 -> 80
```

### Notes

- Requires `kubectl` to be installed and configured
- Press Ctrl+C to stop port-forwarding
- Works with any workload type that creates pods with selector labels

---

## tail pods

Continuously tail logs from Kubernetes pods matching the specified criteria. Automatically discovers new pods and handles pod restarts.

### Synopsis

```bash
tofu k8s tail pods [flags]
```

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--from-deploy` | `-d` | Filter pods by deployment name (can be repeated) | |
| `--labels` | `-l` | Label selector (can be repeated, AND logic) | |
| `--names` | `-n` | Pod name pattern filter (substring match, OR logic) | |
| `--namespace` | | Kubernetes namespace | current context |
| `--all-namespaces` | `-A` | Search pods in all namespaces | `false` |
| `--max-pods` | | Maximum pods to tail simultaneously | `10` |
| `--tail` | | Number of lines to initially read | `20` |
| `--since` | | Only return logs newer than duration (e.g., 5m, 1h) | |
| `--interval` | | Pod discovery poll interval in milliseconds | `500` |

### Examples

Tail pods from a deployment:

```bash
tofu k8s tail pods -d my-deployment
```

Tail pods by label:

```bash
tofu k8s tail pods -l app=nginx
```

Tail pods by name pattern:

```bash
tofu k8s tail pods -n api-server
```

Tail from specific namespace:

```bash
tofu k8s tail pods -d my-app --namespace production
```

Tail from all namespaces:

```bash
tofu k8s tail pods -l app=frontend -A
```

Only logs from last 5 minutes:

```bash
tofu k8s tail pods -d my-app --since 5m
```

Multiple filters:

```bash
tofu k8s tail pods -d web-server -d api-server -l env=prod
```

### Output

```
[my-deployment-abc123] 2024-01-15T10:30:00Z INFO Starting server
[my-deployment-def456] 2024-01-15T10:30:01Z INFO Connection established
[my-deployment-abc123] 2024-01-15T10:30:02Z DEBUG Processing request
```

### Features

- Auto-discovers new pods matching criteria
- Handles pod restarts automatically
- Prefixes each line with pod name
- Supports namespace/pod format with `-A` flag
- Shell completion for deployments and namespaces

### Notes

- Requires `kubectl` to be installed and configured
- Press Ctrl+C to stop tailing
- Maximum 10 pods by default (configurable with `--max-pods`)
