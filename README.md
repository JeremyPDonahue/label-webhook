# AppID Labeling Webhook

A simple Kubernetes mutating webhook that automatically applies `appid` labels to pods based on namespace annotations.

## What it does

This webhook looks for an `appid` annotation or label on namespaces and applies it to all pods created in that namespace. That's it.

## Why this exists

We needed a way to automatically tag pods with application IDs for cost tracking and resource management. Instead of manually labeling every pod, this webhook does it automatically when pods are created.

## Quick Start

### Prerequisites

- OpenShift 4.8+ or Kubernetes 1.21+
- `kubectl` or `oc` CLI configured
- Docker or Podman for building images
- `make` for automation

### Deploy

1. Clone this repo
2. Build and push the image to your registry
3. Update the image reference in `k8s/base/deployment.yaml`
4. Deploy: `kubectl apply -k k8s/base/`

The webhook will deploy in the `kube-system` namespace.

## How to use it

### 1. Set an appid on your namespace

```bash
# Add appid as annotation (preferred)
kubectl annotate namespace my-app appid=my-application-123

# Or as a label (fallback)
kubectl label namespace my-app appid=my-application-123
```

### 2. Create pods in that namespace

All new pods will automatically get the label `managed-by/appid=my-application-123`

### Configuration

The webhook has a few environment variables you can tweak:

| Variable | Default | Description |
|----------|---------|-------------|
| `ENABLE_LABELING` | `true` | Turn the webhook on/off |
| `LABEL_PREFIX` | `managed-by` | Prefix for the appid label |
| `DRY_RUN` | `false` | Log what would happen without doing it |

## Example

If you have a namespace with `appid=my-app-123`, new pods will look like:

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    managed-by/appid: my-app-123
    # ... other labels that were already there
```

That's it. One label gets added.

## Monitoring

The webhook exposes Prometheus metrics on port 9090 and has health checks at `/healthz` and `/readyz`.

## Notes

- The webhook skips system namespaces automatically
- If no appid is found in a namespace, nothing happens
- The webhook uses `failurePolicy: Ignore` so pod creation won't break if the webhook is down
- Only affects new pod creation, not existing pods

## Building

```bash
make build
make docker-build
```

## License

MIT License
