# Tutorial: Converting Mutating Webhook to Production-Ready Multi-Environment Setup

This tutorial will guide you through transforming the original mutating webhook repository into a production-ready, multi-environment deployment system specifically designed for applying custom labels to workloads in OpenShift clusters.

## Table of Contents

1. [Understanding the Original Repository](#understanding-the-original-repository)
2. [Overview of Changes](#overview-of-changes)
3. [Step-by-Step Transformation](#step-by-step-transformation)
4. [Testing Your Changes](#testing-your-changes)
5. [Deployment Guide](#deployment-guide)

## Understanding the Original Repository

The original repository was a basic mutating webhook focused on:
- Image registry mutation for DockerHub images
- Simple configuration via environment variables
- Single deployment scenario
- Basic logging and error handling

### Problems with the Original Setup

1. **Environment-Specific Hardcoding**: Values like namespaces, organization names, and cluster details were hardcoded
2. **Limited Observability**: No metrics, health checks, or structured monitoring
3. **Security Gaps**: Minimal security hardening for production use
4. **Single Environment**: No support for different deployment environments
5. **Poor Automation**: Manual deployment processes without environment flexibility

## Overview of Changes

We will transform the webhook to:

1. **Apply Custom Labels**: Change from image mutation to label application
2. **Multi-Environment Support**: Support sandbox, staging, and production deployments
3. **Production Hardening**: Add security, monitoring, and operational features
4. **Flexible Configuration**: Environment-specific configurations using Kustomize
5. **Automation**: Comprehensive build and deployment automation

## Step-by-Step Transformation

### Step 1: Update Go Dependencies for Production

**Why**: Modern Kubernetes versions, security patches, and additional features like metrics.

**Action**: Update `go.mod` to include newer dependencies and monitoring capabilities.

```bash
# Replace the existing go.mod content
cat > go.mod << 'EOF'
module mutating-webhook

go 1.21

require (
	github.com/hashicorp/logutils v1.0.0
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.28.4
	k8s.io/apimachinery v0.28.4
	k8s.io/client-go v0.28.4
	github.com/prometheus/client_golang v1.17.0
	github.com/stretchr/testify v1.8.4
)
EOF
```

### Step 2: Enhance Configuration for Multi-Environment Support

**Why**: The original config was designed for a single use case. We need flexible configuration for different environments and organizations.

**Action**: Update the configuration structure to support custom labels and environment-specific settings.

```bash
# Update internal/config/config.go
# Replace the Config struct to include new fields for custom labeling
```

Key additions to the Config struct:
- `CustomLabels map[string]string`: For organization-specific labels
- `LabelPrefix string`: Configurable label prefix
- `Organization`, `Environment`, `ClusterName`: Environment identification
- `EnableLabeling`, `DryRun`: Operational controls
- `EnableMetrics`, `MetricsPort`: Monitoring support
- `ExcludedNamespaces []string`: Flexible namespace exclusion

**Why Each Field**:
- **CustomLabels**: Allows organizations to define their own labeling strategy
- **LabelPrefix**: Prevents label conflicts with existing systems
- **Environment fields**: Enables environment-specific labeling and tracking
- **Operational controls**: Allows safe testing and gradual rollout
- **Metrics**: Essential for production monitoring and debugging

### Step 3: Create Prometheus Metrics Module

**Why**: Production systems require observability. Metrics help with debugging, capacity planning, and alerting.

**Action**: Create `internal/metrics/metrics.go` with comprehensive webhook metrics.

```bash
mkdir -p internal/metrics
# Create the metrics module with counters, histograms, and gauges
```

**Key Metrics Added**:
- `webhook_admission_requests_total`: Track all admission requests
- `webhook_labels_applied_total`: Count labels applied
- `webhook_mutations_total`: Track mutation operations
- `webhook_errors_total`: Monitor error rates
- `webhook_up`: Health status indicator
- `webhook_certificate_expiry_timestamp`: Certificate monitoring

### Step 4: Replace Image Mutation with Custom Label Logic

**Why**: The original webhook focused on image registry changes. We need to apply organizational labels for compliance, monitoring, and management.

**Action**: Completely rewrite `internal/operations/podsMutation.go` to focus on label application.

```bash
# Replace the existing podsMutation.go with label-focused logic
```

**Key Changes**:
1. **Replace `podMutationCreate()`** with `podLabelingMutation()`
2. **Add namespace exclusion logic** for system namespaces
3. **Implement label generation** based on pod metadata and configuration
4. **Add admin exemption** via annotations
5. **Include dry-run support** for safe testing

**New Functions Added**:
- `isNamespaceExcluded()`: Comprehensive system namespace detection
- `generateCustomLabels()`: Creates labels based on workload context
- `createLabelPatches()`: Generates JSON patches for label application
- `sanitizeLabelKey/Value()`: Ensures Kubernetes-compliant labels

### Step 5: Add Production-Grade HTTP Server Features

**Why**: The original server lacked health checks, metrics endpoints, and proper error handling needed for production.

**Action**: Enhance `cmd/webhook/httpServer.go` with production features.

**Key Additions**:
1. **Health endpoints**: `/healthz` and `/readyz` for Kubernetes probes
2. **Metrics server**: Separate HTTP server for Prometheus metrics
3. **Certificate monitoring**: Automatic certificate expiry tracking
4. **Enhanced error handling**: Better error responses with metrics recording
5. **Request timing**: Track admission request duration

### Step 6: Improve Main Application with Graceful Shutdown

**Why**: Production applications need to handle shutdown signals gracefully to avoid disrupting ongoing requests.

**Action**: Update `cmd/webhook/main.go` with signal handling and graceful shutdown.

**Key Features**:
- **Signal handling**: Proper SIGTERM/SIGINT handling
- **Graceful shutdown**: Allow in-flight requests to complete
- **Startup logging**: Clear indication of configuration and status
- **Panic recovery**: Prevent crashes from bringing down the service

### Step 7: Create Multi-Environment Kubernetes Manifests

**Why**: Different environments need different configurations, resource limits, and behaviors. Kustomize provides a clean way to manage this.

**Action**: Restructure Kubernetes manifests using Kustomize base and overlays.

```bash
# Create directory structure
mkdir -p k8s/base k8s/overlays/{sandbox,staging,production}

# Move existing manifests to base
mv k8s/*.yaml k8s/base/

# Create environment-specific overlays
```

**Base Configuration** (`k8s/base/`):
- Common resources that apply to all environments
- Default values that can be overridden
- Core functionality without environment-specific customizations

**Sandbox Overlay** (`k8s/overlays/sandbox/`):
- **Dry-run mode enabled**: Safe for testing without affecting workloads
- **Reduced resources**: Lower CPU/memory for cost efficiency
- **Debug logging**: Higher log levels for troubleshooting
- **Single replica**: Minimal footprint for testing

**Production Overlay** (`k8s/overlays/production/`):
- **High availability**: Multiple replicas with anti-affinity
- **Resource limits**: Production-appropriate CPU/memory limits
- **Monitoring enabled**: Full metrics and health checking
- **Strict security**: Minimal logging, no dry-run mode

### Step 8: Add Security Hardening

**Why**: Production workloads in OpenShift require security best practices to meet enterprise security requirements.

**Action**: Update all manifests with security contexts, RBAC, and network policies.

**Security Features Added**:

1. **Pod Security**:
   ```yaml
   securityContext:
     runAsNonRoot: true
     runAsUser: 10001
     runAsGroup: 10001
     fsGroup: 10001
     readOnlyRootFilesystem: true
   ```

2. **Container Security**:
   ```yaml
   capabilities:
     drop: ["ALL"]
   allowPrivilegeEscalation: false
   ```

3. **RBAC**: Minimal permissions following principle of least privilege
4. **Network Policies**: Restrict communication to necessary endpoints only

### Step 9: Create Production-Ready Dockerfile

**Why**: The original Dockerfile wasn't optimized for production use in OpenShift with proper security and minimal attack surface.

**Action**: Replace Dockerfile with multi-stage build and security hardening.

**Key Improvements**:
- **Multi-stage build**: Separate build and runtime stages
- **Minimal base image**: Use scratch for smallest attack surface
- **Non-root user**: OpenShift-compatible user ID (10001)
- **Health check**: Container-level health verification
- **Security flags**: Static compilation with security flags

### Step 10: Add Comprehensive Automation

**Why**: Manual deployment processes are error-prone and don't scale across environments. Automation ensures consistency.

**Action**: Create sophisticated Makefile with environment support.

**Key Features**:
- **Environment detection**: Automatic loading of environment-specific configurations
- **Build automation**: Consistent binary and container builds
- **Deployment automation**: One-command deployment to any environment
- **Certificate management**: Automated TLS certificate generation
- **Testing and verification**: Built-in testing and validation commands

### Step 11: Add Environment Configuration System

**Why**: Hardcoded values make it impossible to deploy the same codebase to different environments with different requirements.

**Action**: Create `.env.example` template for environment-specific configuration.

**Configuration Categories**:

1. **Deployment Configuration**:
   - Container registry and image settings
   - Kubernetes namespace configuration
   - Service and webhook naming

2. **Organization Configuration**:
   - Company/organization details
   - Environment identification
   - Custom labels to apply

3. **Operational Configuration**:
   - Resource limits and requests
   - Replica counts
   - Feature toggles (dry-run, metrics, etc.)

### Step 12: Create Certificate Management System

**Why**: Admission webhooks require TLS certificates. Manual certificate management is error-prone and doesn't scale.

**Action**: Create `scripts/generate-certs.sh` for automated certificate generation.

**Features**:
- **Automated generation**: Creates CA and server certificates
- **Kubernetes integration**: Automatically creates secrets
- **OpenShift compatible**: Proper DNS names and SAN entries
- **Webhook configuration**: Automatically updates webhook manifest with CA bundle

### Step 13: Add Documentation and Tutorial

**Why**: Complex systems need comprehensive documentation for maintenance and onboarding.

**Action**: Create detailed README.md with all necessary information.

**Documentation Sections**:
- **Quick start guide**: Get up and running quickly
- **Configuration reference**: All available options explained
- **Operations guide**: Day-to-day management tasks
- **Troubleshooting**: Common issues and solutions
- **Security considerations**: Important security information

## Testing Your Changes

### 1. Validate Configuration Changes

```bash
# Check Go module updates
go mod tidy
go mod verify

# Verify configuration can be loaded
go run cmd/webhook/main.go -help
```

### 2. Test Kustomize Manifests

```bash
# Validate base configuration
kubectl kustomize k8s/base

# Validate sandbox overlay
kubectl kustomize k8s/overlays/sandbox

# Validate production overlay
kubectl kustomize k8s/overlays/production
```

### 3. Test Container Build

```bash
# Build container image
make docker-build

# Test container runs
docker run --rm ${IMAGE_FULL} -help
```

### 4. Deploy to Test Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your settings
vim .env

# Deploy to sandbox
make deploy ENV=sandbox

# Verify deployment
make status ENV=sandbox
```

## Deployment Guide

### Prerequisites

1. **Container Registry Access**: Push access to your container registry
2. **Kubernetes/OpenShift Cluster**: Admin access to create namespaces and webhooks
3. **Tools**: `kubectl`/`oc`, `docker`/`podman`, `make`, `kustomize`

### Environment Setup

1. **Configure Environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your specific values
   ```

2. **Set Registry Credentials**:
   ```bash
   docker login ${IMAGE_REGISTRY}
   ```

### Deployment Process

1. **Sandbox Deployment** (Safe testing):
   ```bash
   make deploy ENV=sandbox
   make verify ENV=sandbox
   ```

2. **Production Deployment**:
   ```bash
   # Update .env for production values
   make deploy ENV=production
   make verify ENV=production
   ```

### Post-Deployment Verification

1. **Check Webhook Status**:
   ```bash
   make status ENV=production
   ```

2. **Test Label Application**:
   ```bash
   # Create a test pod
   kubectl run test-pod --image=nginx --namespace=test
   
   # Check applied labels
   kubectl get pod test-pod -o jsonpath='{.metadata.labels}' | jq
   ```

3. **Monitor Metrics**:
   ```bash
   # Port-forward to metrics
   kubectl port-forward svc/prod-custom-labels-webhook 9090:9090
   
   # Check metrics
   curl http://localhost:9090/metrics
   ```

## Key Benefits of These Changes

1. **Production Ready**: Security hardening, monitoring, and operational features
2. **Multi-Environment**: Support for sandbox, staging, and production with different configurations
3. **Flexible Configuration**: Easy customization for different organizations and use cases
4. **Automated Operations**: Comprehensive automation for build, deploy, and management
5. **Observable**: Full metrics and logging for debugging and monitoring
6. **Secure**: Security best practices for enterprise environments
7. **Maintainable**: Clean code structure with comprehensive documentation

## Troubleshooting Common Issues

### Build Issues

```bash
# Clear Go cache
go clean -cache -modcache

# Rebuild with verbose output
make build VERBOSE=1
```

### Deployment Issues

```bash
# Check webhook logs
make logs ENV=sandbox

# Debug deployment
make debug ENV=sandbox

# Check certificate status
kubectl get secret webhook-certs -o yaml
```

### Configuration Issues

```bash
# Validate Kustomize
kubectl kustomize k8s/overlays/sandbox --validate

# Check ConfigMap
kubectl get configmap webhook-config -o yaml
```

This transformation converts a basic webhook into a production-ready, enterprise-grade admission controller that can be safely deployed across multiple environments with comprehensive monitoring, security, and operational features.
