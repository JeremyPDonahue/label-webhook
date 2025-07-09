# Custom Labels Mutating Webhook for OpenShift

A production-ready Kubernetes mutating admission webhook designed for OpenShift clusters that automatically applies custom labels to all workloads (pods, deployments, etc.) for improved organization, compliance, and monitoring.

## Features

- **Automatic Labeling**: Applies custom labels to all workloads across the cluster
- **OpenShift Compatible**: Designed specifically for RedHat OpenShift environments
- **Production Ready**: Includes metrics, monitoring, security hardening, and graceful shutdown
- **Configurable**: Extensive configuration options via ConfigMaps and environment variables
- **Secure**: Non-root containers, security contexts, network policies, and RBAC
- **Observable**: Prometheus metrics, structured logging, and health checks
- **Namespace Exclusion**: Automatically excludes system namespaces and supports custom exclusions
- **Dry Run Mode**: Test label application without actual mutations

## Quick Start

### Prerequisites

- OpenShift 4.8+ or Kubernetes 1.21+
- `kubectl` or `oc` CLI configured
- Docker or Podman for building images
- `make` for automation

### Deploy to OpenShift

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd mutating-webhook
   ```

2. **Configure the webhook**:
   ```bash
   # Edit k8s/configmap.yaml to set your organization details
   vim k8s/configmap.yaml
   ```

3. **Build and deploy**:
   ```bash
   # Deploy everything with default settings
   make deploy
   
   # Or step by step:
   make docker-build
   make docker-push
   make deploy-namespace
   make deploy-config
   make deploy-webhook
   make deploy-monitoring
   make deploy-admission
   ```

4. **Verify the deployment**:
   ```bash
   make verify
   make status
   ```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NAMESPACE` | `openshift-webhook` | Webhook namespace |
| `ORGANIZATION` | `default` | Organization name for labels |
| `ENVIRONMENT` | `production` | Environment name for labels |
| `CLUSTER_NAME` | `openshift-cluster` | Cluster name for labels |
| `LABEL_PREFIX` | `managed-by` | Prefix for all applied labels |
| `ENABLE_LABELING` | `true` | Enable/disable label application |
| `ENABLE_METRICS` | `true` | Enable Prometheus metrics |
| `DRY_RUN` | `false` | Dry run mode |
| `LOG_LEVEL` | `60` | Logging verbosity (0-100) |

### ConfigMap Configuration

The webhook uses a ConfigMap (`webhook-config`) for advanced configuration:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: webhook-config
  namespace: openshift-webhook
data:
  organization: "acme-corp"
  environment: "production"
  cluster-name: "openshift-prod-cluster"
  config.yaml: |
    # Custom labels to apply to all workloads
    custom-labels:
      company: "acme-corp"
      cost-center: "engineering"
      data-classification: "internal"
      compliance: "required"
      backup: "enabled"
      monitoring: "enabled"
    
    # Additional excluded namespaces
    excluded-namespaces:
      - "custom-system-namespace"
```

## Applied Labels

The webhook automatically applies the following labels to all pods:

- `{prefix}/webhook`: `custom-labels-mutator`
- `{prefix}/organization`: Organization name
- `{prefix}/environment`: Environment name
- `{prefix}/cluster`: Cluster name
- `{prefix}/namespace`: Pod namespace
- `{prefix}/timestamp`: Creation timestamp
- `{prefix}/created-by`: Username who created the resource
- `{prefix}/workload-type`: Type of workload (deployment, daemonset, etc.)
- `{prefix}/workload-name`: Name of the parent workload
- `{prefix}/appid`: Application ID sourced from namespace annotations or labels
- Custom labels from configuration

### Example Applied Labels

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    managed-by/webhook: custom-labels-mutator
    managed-by/organization: acme-corp
    managed-by/environment: production
    managed-by/cluster: openshift-prod-cluster
    managed-by/namespace: default
    managed-by/timestamp: "2024-01-15T10:30:00Z"
    managed-by/created-by: admin
    managed-by/workload-type: deployment
    managed-by/workload-name: my-app
    managed-by/appid: my-application-123
    company: acme-corp
    cost-center: engineering
    data-classification: internal
```

## Monitoring and Observability

### Prometheus Metrics

The webhook exposes metrics on port 9090:

- `webhook_admission_requests_total`: Total admission requests
- `webhook_admission_request_duration_seconds`: Request duration
- `webhook_labels_applied_total`: Total labels applied
- `webhook_mutations_total`: Total mutations performed
- `webhook_errors_total`: Total errors encountered
- `webhook_up`: Webhook health status
- `webhook_certificate_expiry_timestamp`: Certificate expiry time

### Health Checks

- **Liveness**: `GET /healthz`
- **Readiness**: `GET /readyz`
- **Metrics**: `GET /metrics` (port 9090)

### Logs

```bash
# View logs
make logs

# Debug issues
make debug
```

## Security

### Security Features

- Non-root container execution (UID 10001)
- Read-only root filesystem
- Dropped capabilities
- Security contexts and seccomp profiles
- Network policies for restricted communication
- RBAC with minimal required permissions
- TLS encryption for all communications

### Excluded Namespaces

The webhook automatically excludes system namespaces:

- All `kube-*` namespaces
- All `openshift-*` namespaces
- Custom excluded namespaces from configuration

### Admin Exemption

Pods can be exempted from labeling by adding an annotation:

```yaml
metadata:
  annotations:
    webhook.openshift.io/exempt: "true"
```

## Operations

### Building

```bash
# Build binary
make build

# Build Docker image
make docker-build

# Push to registry
make docker-push
```

### Deployment

```bash
# Full deployment
make deploy

# Update existing deployment
make update

# Remove webhook
make undeploy
```

### Certificate Management

```bash
# Generate certificates
make cert-setup

# Certificates are automatically managed
# Certificate expiry is monitored via metrics
```

### Troubleshooting

```bash
# Check status
make status

# View logs
make logs

# Debug deployment
make debug

# Test functionality
make verify
```

## Development

### Setup Development Environment

```bash
# Install dependencies
make dev-setup

# Run tests
make test

# Run linters
make lint

# Format code
make fmt
```

### Project Structure

```
├── cmd/webhook/           # Main application
├── internal/
│   ├── config/           # Configuration management
│   ├── metrics/          # Prometheus metrics
│   ├── operations/       # Admission logic
│   └── certificate/      # Certificate management
├── k8s/                  # Kubernetes manifests
├── scripts/              # Helper scripts
├── Dockerfile            # Container image
├── Makefile             # Build automation
└── README.md            # This file
```

## Application ID (appid) Configuration

The webhook automatically extracts the application ID from the namespace where the pod is created. This allows for automatic application identification and labeling.

### Setting appid in Namespace

To set the appid for a namespace, add it as an annotation:

```bash
# Set appid in namespace annotations
kubectl annotate namespace my-app-namespace appid=my-application-123
```

Or as a label (fallback option):

```bash
# Set appid in namespace labels
kubectl label namespace my-app-namespace appid=my-application-123
```

### Example Namespace Configuration

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: my-app-namespace
  annotations:
    appid: my-application-123
  labels:
    environment: production
    team: engineering
```

**Note**: The webhook will first check annotations, then labels as a fallback. If no appid is found, the appid label will not be applied to pods.

## Customization

### Adding Custom Labels

Edit the ConfigMap to add custom labels:

```yaml
custom-labels:
  your-label: "your-value"
  department: "engineering"
  project: "my-project"
```

### Excluding Additional Namespaces

```yaml
excluded-namespaces:
  - "my-system-namespace"
  - "third-party-system"
```

### Modifying Label Logic

Edit `internal/operations/podsMutation.go` to customize label generation logic.

## FAQ

**Q: Does this affect existing workloads?**
A: No, the webhook only affects new pod creations. Existing pods are not modified.

**Q: Can I disable the webhook temporarily?**
A: Yes, set `DRY_RUN=true` or scale the deployment to 0 replicas.

**Q: What happens if the webhook is down?**
A: The webhook uses `failurePolicy: Ignore`, so pod creation continues without labels.

**Q: How do I update the configuration?**
A: Update the ConfigMap and restart the webhook pods.

**Q: Is this compatible with other admission webhooks?**
A: Yes, multiple admission webhooks can coexist.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Production Gaps

While the webhook is robust and contains significant production-ready features, the following gaps should be addressed for a truly enterprise-grade deployment:

### Testing (Major Gap)
- No unit tests for critical mutation logic
- No integration tests for webhook functionality
- No end-to-end tests for deployment scenarios
- No performance testing under load

### Error Handling (Minor Gap)
- Limited error scenarios covered
- No circuit breaker patterns
- No retry logic for transient failures

### Monitoring (Minor Gap)
- No alerting rules defined
- No runbooks for incident response
- No SLO/SLI definitions

### Documentation (Minor Gap)
- No troubleshooting guides
- No operational runbooks
- No disaster recovery procedures

To fully realize the production potential, focusing on these areas will ensure a comprehensive, reliable deployment.

## Support

For issues and questions:

1. Check the troubleshooting section
2. Review logs with `make debug`
3. Open an issue in the repository

## Changelog

### v1.0.0
- Initial production release
- Custom label application
- OpenShift compatibility
- Prometheus metrics
- Security hardening
- Comprehensive documentation
