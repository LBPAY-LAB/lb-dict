# ===================================
# Stage 1: Build Stage
# ===================================
FROM golang:1.24.5-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always --dirty) -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o /app/bin/rsfn-bridge \
    cmd/server/main.go

# ===================================
# Stage 2: Runtime Stage
# ===================================
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    openssl

# Create non-root user
RUN addgroup -g 1001 appuser && \
    adduser -u 1001 -G appuser -D -h /app appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/rsfn-bridge /app/rsfn-bridge

# Copy configuration files (if needed)
COPY --from=builder /app/config /app/config

# Create directories for certs and logs
RUN mkdir -p /app/certs /app/logs && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 9094 8082

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8082/health || exit 1

# Set entrypoint
ENTRYPOINT ["/app/rsfn-bridge"]

# Default command (can be overridden)
CMD ["--config", "/app/config/config.yaml"]
