# Multi-stage Dockerfile for conn-dict
# Build images for both Connect API and Temporal Worker

# Stage 1: Builder
FROM golang:1.24.5-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build Connect binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /build/bin/connect \
    cmd/connect/main.go

# Build Worker binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /build/bin/worker \
    cmd/worker/main.go

# Stage 2: Connect Runtime
FROM alpine:3.20 AS connect

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bin/connect /app/connect

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Set timezone
ENV TZ=America/Sao_Paulo

# Change ownership
RUN chown -R app:app /app

USER app

# Expose ports
EXPOSE 9092 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8081/health || exit 1

ENTRYPOINT ["/app/connect"]

# Stage 3: Worker Runtime
FROM alpine:3.20 AS worker

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bin/worker /app/worker

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Set timezone
ENV TZ=America/Sao_Paulo

# Change ownership
RUN chown -R app:app /app

USER app

# Temporal Worker doesn't expose HTTP ports by default
# But we can add a health check endpoint if needed

ENTRYPOINT ["/app/worker"]
