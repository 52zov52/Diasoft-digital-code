# syntax=docker/dockerfile:1
# Stage 1: Build
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -extldflags '-static'" -o /diploma-verify ./cmd/server/

# Stage 2: Runtime (Hardened Alpine)
FROM alpine:3.19 AS runtime
RUN apk --no-cache add ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /diploma-verify .
COPY db/migrations ./db/migrations/
COPY docs/api/openapi.yaml ./docs/api/openapi.yaml

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser:appgroup

EXPOSE 8080
# Healthcheck использует встроенный wget busybox
HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/diploma-verify"]