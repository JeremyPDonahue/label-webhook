# Multi-stage build for production OpenShift deployment

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Create app user for security
RUN addgroup -g 10001 -S webhook && \
    adduser -u 10001 -S webhook -G webhook

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with security and optimization flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -tags "timetzdata netgo" \
    -o webhook ./cmd/webhook

# Final stage - minimal runtime image
FROM scratch

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy the binary
COPY --from=builder /build/webhook /webhook

# Create directories that may be needed
COPY --from=builder --chown=10001:10001 /tmp /tmp

# Use non-root user (OpenShift compatible)
USER 10001:10001

# Set working directory
WORKDIR /

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/webhook", "-help"]

# Expose ports
EXPOSE 8443 9090

# Run the webhook
ENTRYPOINT ["/webhook"]
