#!/bin/bash

# Certificate generation script for Custom Labels Webhook
# Usage: ./generate-certs.sh <namespace> <service-name>

set -euo pipefail

NAMESPACE="${1:-openshift-webhook}"
SERVICE_NAME="${2:-custom-labels-webhook}"
SECRET_NAME="webhook-certs"

echo "Generating certificates for webhook in namespace: $NAMESPACE"

# Create temporary directory
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

cd "$TEMP_DIR"

# Generate CA private key
openssl genrsa -out ca.key 4096

# Generate CA certificate
openssl req -new -x509 -days 365 -key ca.key -out ca.crt -subj "/CN=webhook-ca/O=acme-corp"

# Generate server private key
openssl genrsa -out server.key 4096

# Create certificate signing request configuration
cat > csr.conf <<EOF
[req]
default_bits = 4096
prompt = no
distinguished_name = dn
req_extensions = v3_req

[dn]
CN = $SERVICE_NAME.$NAMESPACE.svc
O = acme-corp

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = $SERVICE_NAME
DNS.2 = $SERVICE_NAME.$NAMESPACE
DNS.3 = $SERVICE_NAME.$NAMESPACE.svc
DNS.4 = $SERVICE_NAME.$NAMESPACE.svc.cluster
DNS.5 = $SERVICE_NAME.$NAMESPACE.svc.cluster.local
IP.1 = 127.0.0.1
EOF

# Generate certificate signing request
openssl req -new -key server.key -out server.csr -config csr.conf

# Generate server certificate signed by CA
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365 -extensions v3_req -extfile csr.conf

# Verify certificate
echo "Verifying generated certificate..."
openssl x509 -in server.crt -text -noout | grep -A1 "Subject Alternative Name"

# Create Kubernetes secret
echo "Creating Kubernetes secret: $SECRET_NAME in namespace: $NAMESPACE"
kubectl create secret tls "$SECRET_NAME" \
    --cert=server.crt \
    --key=server.key \
    --namespace="$NAMESPACE" \
    --dry-run=client -o yaml | kubectl apply -f -

# Get CA bundle for webhook configuration
CA_BUNDLE=$(base64 < ca.crt | tr -d '\n')

echo "Certificate generation completed!"
echo "CA Bundle for webhook configuration:"
echo "$CA_BUNDLE"

# Update webhook configuration with CA bundle
if command -v yq &> /dev/null; then
    echo "Updating webhook configuration with CA bundle..."
    yq eval ".webhooks[0].clientConfig.caBundle = \"$CA_BUNDLE\"" -i "../k8s/webhook.yaml"
    echo "Webhook configuration updated!"
else
    echo "yq not found. Please manually update the caBundle in k8s/webhook.yaml with:"
    echo "$CA_BUNDLE"
fi

echo "Certificates successfully generated and deployed!"
