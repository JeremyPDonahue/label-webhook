# Step 1 - Certificate Container
####
FROM registry.c.test-chamber-13.lan/library/alpine:latest as certHost
RUN addgroup -S -g 1000 app && \
    adduser --disabled-password -G app --gecos "application account" --home "/home/app" --shell "/sbin/nologin" --no-create-home --uid 1000 app

# Step 2 - Build Container
####
FROM registry.c.test-chamber-13.lan/dockerhub/library/golang:alpine as builder

COPY . /go/src/app

WORKDIR /go/src/app

RUN apk add --no-cache git && \
    git config --global --add safe.directory /go/src/app && \
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -ldflags="-s -w" -tags timetzdata -o webhook ./cmd/webhook

# Step 3 - Running Container
####
FROM scratch

COPY --from=certHost /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=certHost /etc/passwd /etc/group /etc/
COPY --from=builder --chown=app:app /go/src/app/webhook /app/webhook
COPY html/ /app/html/

USER app:app
WORKDIR /app/

ENTRYPOINT ["/app/webhook"]
