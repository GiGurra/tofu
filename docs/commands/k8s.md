# k8s

Kubernetes utilities.

## Synopsis

```bash
tofu k8s <subcommand> [flags]
```

## Subcommands

- [`tail pods`](#tail-pods) - Tail logs from Kubernetes pods

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
