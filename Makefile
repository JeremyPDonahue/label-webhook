# Custom Labels Mutating Webhook for OpenShift
# Multi-environment deployment automation

.PHONY: all build test clean deploy undeploy cert-setup help

# Load environment configuration
include .env
export

# Default environment if not specified
ENV ?= sandbox

# Variables (can be overridden by .env file)
IMAGE_REGISTRY ?= quay.io/your-org
IMAGE_NAME ?= custom-labels-webhook
IMAGE_TAG ?= develop
IMAGE_FULL = $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

NAMESPACE ?= webhook-$(ENV)
WEBHOOK_NAME ?= $(ENV)-custom-labels-mutator
ORGANIZATION ?= your-company
CLUSTER_NAME ?= $(ENV)-cluster

# Go variables
GO_MODULE = mutating-webhook
GO_VERSION = 1.21
GOOS = linux
GOARCH = amd64

# Build flags
LDFLAGS = -w -s -extldflags "-static"
BUILD_FLAGS = -a -installsuffix cgo -tags "timetzdata netgo"

# Kustomize overlay path
OVERLAY_PATH = k8s/overlays/$(ENV)

# Default target
all: test build

## help: Display this help message
help:
	@echo "Custom Labels Mutating Webhook - Production Deployment"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## clean: Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f webhook
	@docker rmi $(IMAGE_FULL) 2>/dev/null || true

## test: Run unit tests
test:
	@echo "Running unit tests..."
	@go mod download
	@go test -v ./...
	@go vet ./...
	@gofmt -l .

## build: Build the webhook binary
build:
	@echo "Building webhook binary..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags="$(LDFLAGS)" \
		$(BUILD_FLAGS) \
		-o webhook ./cmd/webhook

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image: $(IMAGE_FULL)"
	@docker build -t $(IMAGE_FULL) .

## docker-push: Push Docker image to registry
docker-push: docker-build
	@echo "Pushing Docker image: $(IMAGE_FULL)"
	@docker push $(IMAGE_FULL)

## cert-setup: Setup certificates for the webhook
cert-setup:
	@echo "Setting up certificates..."
	@kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	@./scripts/generate-certs.sh $(NAMESPACE) custom-labels-webhook

## deploy-namespace: Deploy namespace and RBAC
deploy-namespace:
	@echo "Deploying namespace and RBAC..."
	@kubectl apply -f k8s/namespace.yaml
	@kubectl apply -f k8s/rbac.yaml

## deploy-config: Deploy configuration
deploy-config:
	@echo "Deploying configuration..."
	@kubectl apply -f k8s/configmap.yaml

## deploy-webhook: Deploy webhook service and deployment
deploy-webhook:
	@echo "Deploying webhook..."
	@kubectl apply -f k8s/service.yaml
	@kubectl apply -f k8s/daemonset.yaml
	@kubectl apply -f k8s/network-policy.yaml

## deploy-monitoring: Deploy monitoring resources
deploy-monitoring:
	@echo "Deploying monitoring resources..."
	@kubectl apply -f k8s/service-monitor.yaml

## deploy-admission: Deploy admission webhook configuration
deploy-admission:
	@echo "Deploying admission webhook configuration..."
	@kubectl apply -f k8s/webhook.yaml

## deploy: Full deployment
deploy: docker-push deploy-namespace deploy-config deploy-webhook deploy-monitoring
	@echo "Waiting for webhook to be ready..."
	@kubectl wait --for=condition=available --timeout=300s deployment/custom-labels-webhook -n $(NAMESPACE)
	@make deploy-admission
	@echo "Deployment completed successfully!"

## undeploy: Remove all webhook resources
undeploy:
	@echo "Removing webhook resources..."
	@kubectl delete -f k8s/webhook.yaml --ignore-not-found=true
	@kubectl delete -f k8s/service-monitor.yaml --ignore-not-found=true
	@kubectl delete -f k8s/network-policy.yaml --ignore-not-found=true
	@kubectl delete -f k8s/daemonset.yaml --ignore-not-found=true
	@kubectl delete -f k8s/service.yaml --ignore-not-found=true
	@kubectl delete -f k8s/configmap.yaml --ignore-not-found=true
	@kubectl delete -f k8s/rbac.yaml --ignore-not-found=true
	@kubectl delete namespace $(NAMESPACE) --ignore-not-found=true

## logs: View webhook logs
logs:
	@kubectl logs -f deployment/custom-labels-webhook -n $(NAMESPACE)

## status: Check webhook status
status:
	@echo "Webhook deployment status:"
	@kubectl get all -n $(NAMESPACE)
	@echo ""
	@echo "Admission webhook configuration:"
	@kubectl get mutatingwebhookconfigurations $(WEBHOOK_NAME) -o yaml

## debug: Debug webhook issues
debug:
	@echo "Debugging webhook..."
	@kubectl describe deployment custom-labels-webhook -n $(NAMESPACE)
	@kubectl describe pods -l app=custom-labels-webhook -n $(NAMESPACE)
	@kubectl logs deployment/custom-labels-webhook -n $(NAMESPACE) --tail=50

## verify: Verify webhook is working
verify:
	@echo "Verifying webhook functionality..."
	@kubectl create namespace webhook-test --dry-run=client -o yaml | kubectl apply -f -
	@kubectl run test-pod --image=nginx --namespace=webhook-test --rm -it --restart=Never -- /bin/bash -c "echo 'Test completed'"
	@kubectl get pod test-pod -n webhook-test -o jsonpath='{.metadata.labels}' | jq .
	@kubectl delete namespace webhook-test

## update: Update webhook image
update: docker-push
	@echo "Updating webhook deployment..."
	@kubectl set image deployment/custom-labels-webhook webhook=$(IMAGE_FULL) -n $(NAMESPACE)
	@kubectl rollout status deployment/custom-labels-webhook -n $(NAMESPACE)

# Development targets
dev-setup:
	@echo "Setting up development environment..."
	@go mod download
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint:
	@echo "Running linters..."
	@golangci-lint run

fmt:
	@echo "Formatting code..."
	@gofmt -w .
	@go mod tidy
