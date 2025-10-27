# ===================================
# Stage 1: Build Stage
# ===================================
FROM maven:3.9.5-eclipse-temurin-17 AS builder

# Set working directory
WORKDIR /app

# Copy pom.xml
COPY pom.xml .

# Download dependencies (cached layer)
RUN mvn dependency:go-offline -B

# Copy source code
COPY src ./src

# Build application
RUN mvn clean package -DskipTests -B

# ===================================
# Stage 2: Runtime Stage
# ===================================
FROM eclipse-temurin:17-jre-alpine

# Install runtime dependencies
RUN apk add --no-cache \
    curl \
    tzdata \
    openssl

# Create non-root user
RUN addgroup -g 1001 appuser && \
    adduser -u 1001 -G appuser -D -h /app appuser

# Set working directory
WORKDIR /app

# Copy JAR from builder
COPY --from=builder /app/target/xml-signer-1.0.0.jar /app/xml-signer.jar

# Create directories for certs and logs
RUN mkdir -p /app/certs /app/logs && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/actuator/health || exit 1

# Set JVM options
ENV JAVA_OPTS="-Xms256m -Xmx512m -XX:+UseG1GC -XX:MaxGCPauseMillis=200"

# Run application
ENTRYPOINT ["sh", "-c", "java $JAVA_OPTS -jar /app/xml-signer.jar"]
