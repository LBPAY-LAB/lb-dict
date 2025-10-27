# ============================================
# Core DICT - Multi-Stage Dockerfile
# ============================================

# ============================================
# Stage 1: Builder
# ============================================
FROM golang:1.24.5-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    make \
    ca-certificates \
    tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first (for better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 for static binary
# -trimpath removes file system paths from the resulting executable
# -ldflags="-s -w" strips debug information to reduce binary size
ARG VERSION=1.0.0
ARG BUILD_TIME
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
    -o /app/bin/core-dict-grpc \
    ./cmd/grpc/

# Verify the binary was built
RUN ls -lh /app/bin/core-dict-grpc

# ============================================
# Stage 2: Runtime
# ============================================
FROM alpine:3.20 AS runtime

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    && update-ca-certificates

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/core-dict-grpc /app/core-dict-grpc

# Copy migrations (if needed at runtime)
COPY --from=builder /app/migrations /app/migrations

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Set timezone (can be overridden via ENV)
ENV TZ=America/Sao_Paulo

# Create directories for logs and temp files
RUN mkdir -p /app/logs /app/tmp && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose ports
# gRPC API
EXPOSE 9090
# Metrics (Prometheus)
EXPOSE 9091

# Health check (gRPC health check - requires grpc-health-probe or custom script)
# For now, commented out until grpc-health-probe is available
# HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
#     CMD /app/grpc-health-probe -addr=:9090 || exit 1

# Set environment variables
ENV APP_ENV=production \
    GRPC_PORT=9090 \
    LOG_LEVEL=info \
    LOG_FORMAT=json

# Run the application
ENTRYPOINT ["/app/core-dict-grpc"]

# ============================================
# Labels (OCI standard)
# ============================================
LABEL org.opencontainers.image.title="Core DICT" \
      org.opencontainers.image.description="Core DICT - Business Logic Service for PIX DICT" \
      org.opencontainers.image.vendor="LBPay" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.source="https://github.com/lbpay-lab/core-dict" \
      org.opencontainers.image.licenses="Proprietary"

# ============================================
# Build Instructions
# ============================================
# Build:
#   docker build -t lbpay/core-dict:latest .
#
# Build with version:
#   docker build --build-arg VERSION=1.0.0 --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) -t lbpay/core-dict:1.0.0 .
#
# Run:
#   docker run --rm -it --env-file .env -p 8080:8080 -p 9090:9090 lbpay/core-dict:latest
#
# Run with custom env vars:
#   docker run --rm -it \
#     -e DATABASE_URL=postgres://user:pass@host:5432/db \
#     -e REDIS_URL=redis://host:6379 \
#     -e PULSAR_URL=pulsar://host:6650 \
#     -p 8080:8080 -p 9090:9090 \
#     lbpay/core-dict:latest
