# Repo Overview: Mutating Webhook

This document provides an overview of the files and structure within the Mutating Webhook repository, acting as a guide for understanding and discussing the functionality and configuration of the webhook.

## Code Structure

### `cmd/webhook`
- **`main.go`**: The entry point for the webhook application. It initializes the configuration, sets up signal handling for graceful shutdown, and starts the HTTP server.
- **`httpServer.go`**: Contains the setup and configuration for the webhook's HTTP server, handling incoming requests, loading TLS certificates, and defining service endpoints.
- **`httpServerTemplates.go`**: Manages the HTTP response templates for the webhook, including health check and root path responses.

### `internal/operations`
- **`podsMutation.go`**: Implements the core logic for mutating incoming pod admission requests. It handles applying the `appid` label based on namespace annotations or labels.
- **`parsers.go`**: Provides utility functions for parsing Kubernetes resources, such as pods and deployments.
- **`patch.go`**: Defines JSON Patch operations used to modify pod resources.

### `internal/config`
- **`config.go`**: Defines the configuration structure for the application, specifying default values and environment variables.
- **`initialize.go`**: Handles initialization of configuration, loading from environment and config files, and certificate management.
- **`configFile.go`**: Parses and loads configuration files into the application.

### `internal/metrics`
- **`metrics.go`**: Implements metrics collection and reporting using Prometheus, tracking admission requests, applied labels, errors, and health status.

## Kubernetes Configuration

### `k8s/base`
- **`deployment.yaml`**: Defines the deployment of the webhook, specifying replicas, container settings, and resources.
- **`service.yaml`**: Exposes the webhook as a Kubernetes Service on port `443` for the webhook and `9090` for metrics.
- **`service-monitor.yaml`**: Sets up a ServiceMonitor resource to scrape metrics with Prometheus.
- **`rbac.yaml`**: Configures role-based access control for the webhook, defining roles, role bindings, and service accounts.
- **`webhook.yaml`**: Defines the MutatingWebhookConfiguration, specifying rules for mutating pod creation requests.

### `k8s/overlays`
- **`production/`** and **`sandbox/`**: Overlays for deploying in different environments, using `kustomization.yaml` to modify base configurations.

## Documentation and Utility

- **`WEBHOOK_SIMPLIFICATION_SUMMARY.md`**: A detailed summary of recent simplifications made to the webhook, focusing on the AppID label's application.
- **Supporting Files**: Includes `go.mod` and `go.sum` for dependency management.


## How to Use
1. **Deploy**: Use Kubernetes manifests in the `k8s/base` directory to deploy the webhook and configure it according to your Kubernetes cluster.
2. **Monitor**: Ensure Prometheus is set up to scrape metrics using the `service-monitor.yaml` configuration.
3. **Configure**: Modify `config.yaml` and environment variables as needed to match your deployment environment.

### Key Points
- **Focus**: The webhook exclusively applies the `appid` label to pods in supported namespaces.
- **Metrics**: Provides robust metrics collection to monitor performance and operations.
- **Customization**: Easily adapted for different environments through kustomize overlays.
