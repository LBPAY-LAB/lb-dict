# DEV-006: Docker Images Specification - DICT

**Projeto**: DICT - Diretorio de Identificadores de Contas Transacionais (LBPay)
**Componente**: Docker Images & Container Strategy
**Versao**: 1.0
**Data**: 2025-10-25
**Autor**: DEVOPS (AI Agent - DevOps Engineer)
**Revisor**: [Aguardando]
**Aprovador**: Head de DevOps, CTO (Jose Luis Silva)

---

## Controle de Versao

| Versao | Data | Autor | Descricao das Mudancas |
|--------|------|-------|------------------------|
| 1.0 | 2025-10-25 | DEVOPS | Versao inicial - Dockerfiles e estrategia de container para DICT |

---

## Sumario Executivo

### Visao Geral

Especificacao completa de **Docker Images** para o **DICT Stack**, incluindo:
- Dockerfiles multi-stage otimizados para todos os servicos
- Base images minimais (Alpine, Distroless)
- Security scanning (Trivy, Grype)
- Image signing (Cosign)
- Registry strategy (AWS ECR)
- Versioning & tagging strategy

### Objetivos

- Imagens minimas: < 50MB para servicos Go
- Seguranca: Zero vulnerabilidades HIGH/CRITICAL
- Performance: Build time < 5 minutos
- Reproducibilidade: Builds deterministicos
- Compliance: Image signing obrigatorio para producao

### Container Stack

| Service | Language | Base Image | Size Target | Build Time |
|---------|----------|------------|-------------|------------|
| **Core DICT** | Go 1.24.5 | distroless/static | < 50MB | < 3 min |
| **Connect API** | Go 1.24.5 | distroless/static | < 50MB | < 3 min |
| **Connect Worker** | Go 1.24.5 | distroless/static | < 50MB | < 3 min |
| **RSFN Bridge** | Go 1.24.5 + Java 21 | distroless/java21 | < 200MB | < 5 min |

---

## Indice

1. [Build Strategy](#1-build-strategy)
2. [Core DICT Dockerfile](#2-core-dict-dockerfile)
3. [Connect API Dockerfile](#3-connect-api-dockerfile)
4. [Connect Worker Dockerfile](#4-connect-worker-dockerfile)
5. [RSFN Bridge Dockerfile](#5-rsfn-bridge-dockerfile)
6. [Security Scanning](#6-security-scanning)
7. [Image Signing](#7-image-signing)
8. [Registry Strategy](#8-registry-strategy)
9. [Rastreabilidade](#9-rastreabilidade)

---

## 1. Build Strategy

### Multi-Stage Build Pattern

```dockerfile
# Pattern: 3-stage build

Stage 1: Builder
  - Full SDK image (golang:alpine, maven:jdk21)
  - Install dependencies
  - Build binary/JAR
  - Optimize binary (strip symbols, UPX compression)

Stage 2: Runtime (optional)
  - Prepare runtime dependencies
  - Copy certificates, configs

Stage 3: Final
  - Minimal base (distroless/static, distroless/java)
  - Copy binary from builder
  - Non-root user
  - Minimal attack surface
```

### Build Optimization Techniques

```yaml
Techniques:
  1. Layer caching:
     - COPY go.mod go.sum before COPY . .
     - Leverage Docker BuildKit cache

  2. Dependency caching:
     - go mod download in separate layer
     - Maven dependency:go-offline

  3. Binary optimization:
     - CGO_ENABLED=0 (static linking)
     - -ldflags="-w -s" (strip debug symbols)
     - UPX compression (optional, 30-50% size reduction)

  4. Multi-platform builds:
     - docker buildx (linux/amd64, linux/arm64)

  5. BuildKit secrets:
     - --mount=type=secret for credentials
     - Never bake secrets into image layers
```

### .dockerignore

```dockerignore
# .dockerignore (all services)

# Git
.git
.gitignore

# Documentation
*.md
docs/

# Tests
*_test.go
test/
coverage.out

# Build artifacts
bin/
dist/
target/

# IDE
.vscode/
.idea/
*.swp

# CI/CD
.github/
.circleci/

# Configs (use ConfigMaps)
*.yaml
*.yml
!Dockerfile

# Temporary files
tmp/
*.tmp
*.log
```

---

## 2. Core DICT Dockerfile

### Dockerfile

```dockerfile
# core-dict/Dockerfile

# ============================================
# Stage 1: Builder
# ============================================
FROM golang:1.24.5-alpine3.19 AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    make \
    ca-certificates \
    tzdata

# Set working directory
WORKDIR /build

# Copy go.mod and go.sum first (better caching)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build arguments
ARG VERSION=dev
ARG BUILD_DATE
ARG VCS_REF
ARG GOOS=linux
ARG GOARCH=amd64

# Build binary
RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
    go build \
    -ldflags="-w -s \
      -X main.Version=${VERSION} \
      -X main.BuildDate=${BUILD_DATE} \
      -X main.GitCommit=${VCS_REF}" \
    -o core-dict \
    ./cmd/server

# Optional: UPX compression (30-50% size reduction)
# RUN apk add --no-cache upx && upx --best --lzma core-dict

# Verify binary
RUN ./core-dict --version

# ============================================
# Stage 2: Final (Distroless)
# ============================================
FROM gcr.io/distroless/static-debian12:nonroot

# Metadata labels (OCI standard)
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.authors="LBPay DevOps <devops@lbpay.io>"
LABEL org.opencontainers.image.url="https://github.com/lbpay/dict-core"
LABEL org.opencontainers.image.documentation="https://docs.lbpay.io/dict"
LABEL org.opencontainers.image.source="https://github.com/lbpay/dict-core"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.revision="${VCS_REF}"
LABEL org.opencontainers.image.vendor="LBPay"
LABEL org.opencontainers.image.title="Core DICT"
LABEL org.opencontainers.image.description="Domain service for DICT operations"

# Copy binary from builder
COPY --from=builder /build/core-dict /core-dict

# Copy CA certificates (for HTTPS)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Expose ports
EXPOSE 8080 9090 50051

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/core-dict", "healthcheck"]

# Run as non-root user (distroless default: nobody)
USER nonroot:nonroot

# Entrypoint
ENTRYPOINT ["/core-dict"]

# Default command (can be overridden)
CMD ["serve"]
```

### Build & Push

```bash
#!/bin/bash
# build-core-dict.sh

set -euo pipefail

# Variables
VERSION=${VERSION:-$(git describe --tags --always)}
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
VCS_REF=$(git rev-parse --short HEAD)
IMAGE_NAME="lbpay/core-dict"
ECR_REGISTRY="123456789012.dkr.ecr.us-east-1.amazonaws.com"

# Build with BuildKit
DOCKER_BUILDKIT=1 docker build \
  --build-arg VERSION="${VERSION}" \
  --build-arg BUILD_DATE="${BUILD_DATE}" \
  --build-arg VCS_REF="${VCS_REF}" \
  --tag "${IMAGE_NAME}:${VERSION}" \
  --tag "${IMAGE_NAME}:latest" \
  --file Dockerfile \
  --progress=plain \
  .

# Tag for ECR
docker tag "${IMAGE_NAME}:${VERSION}" "${ECR_REGISTRY}/${IMAGE_NAME}:${VERSION}"
docker tag "${IMAGE_NAME}:latest" "${ECR_REGISTRY}/${IMAGE_NAME}:latest"

# Security scan
trivy image --severity HIGH,CRITICAL "${IMAGE_NAME}:${VERSION}"

# Push to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin "${ECR_REGISTRY}"

docker push "${ECR_REGISTRY}/${IMAGE_NAME}:${VERSION}"
docker push "${ECR_REGISTRY}/${IMAGE_NAME}:latest"

echo "Image pushed: ${ECR_REGISTRY}/${IMAGE_NAME}:${VERSION}"
```

---

## 3. Connect API Dockerfile

### Dockerfile

```dockerfile
# connect-api/Dockerfile

# ============================================
# Stage 1: Builder
# ============================================
FROM golang:1.24.5-alpine3.19 AS builder

RUN apk add --no-cache git make ca-certificates tzdata

WORKDIR /build

# Dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Source
COPY . .

# Build args
ARG VERSION=dev
ARG BUILD_DATE
ARG VCS_REF
ARG GOOS=linux
ARG GOARCH=amd64

# Build
RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
    go build \
    -ldflags="-w -s \
      -X main.Version=${VERSION} \
      -X main.BuildDate=${BUILD_DATE} \
      -X main.GitCommit=${VCS_REF}" \
    -o connect-api \
    ./cmd/api

# ============================================
# Stage 2: Final
# ============================================
FROM gcr.io/distroless/static-debian12:nonroot

LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.revision="${VCS_REF}"
LABEL org.opencontainers.image.title="Connect API"
LABEL org.opencontainers.image.description="REST API for RSFN Connect orchestration"

COPY --from=builder /build/connect-api /connect-api
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

EXPOSE 8080 9090

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD ["/connect-api", "healthcheck"]

USER nonroot:nonroot

ENTRYPOINT ["/connect-api"]
CMD ["serve"]
```

---

## 4. Connect Worker Dockerfile

### Dockerfile

```dockerfile
# connect-worker/Dockerfile

# ============================================
# Stage 1: Builder
# ============================================
FROM golang:1.24.5-alpine3.19 AS builder

RUN apk add --no-cache git make ca-certificates tzdata

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

ARG VERSION=dev
ARG BUILD_DATE
ARG VCS_REF
ARG GOOS=linux
ARG GOARCH=amd64

RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
    go build \
    -ldflags="-w -s \
      -X main.Version=${VERSION} \
      -X main.BuildDate=${BUILD_DATE} \
      -X main.GitCommit=${VCS_REF}" \
    -o connect-worker \
    ./cmd/worker

# ============================================
# Stage 2: Final
# ============================================
FROM gcr.io/distroless/static-debian12:nonroot

LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.revision="${VCS_REF}"
LABEL org.opencontainers.image.title="Connect Worker"
LABEL org.opencontainers.image.description="Temporal worker for DICT workflows"

COPY --from=builder /build/connect-worker /connect-worker
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

EXPOSE 9090

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD ["/connect-worker", "healthcheck"]

USER nonroot:nonroot

ENTRYPOINT ["/connect-worker"]
CMD ["run"]
```

---

## 5. RSFN Bridge Dockerfile

### Dockerfile

```dockerfile
# rsfn-bridge/Dockerfile

# ============================================
# Stage 1: Go Builder
# ============================================
FROM golang:1.24.5-alpine3.19 AS go-builder

RUN apk add --no-cache git make ca-certificates tzdata

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

ARG VERSION=dev
ARG BUILD_DATE
ARG VCS_REF
ARG GOOS=linux
ARG GOARCH=amd64

RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
    go build \
    -ldflags="-w -s \
      -X main.Version=${VERSION} \
      -X main.BuildDate=${BUILD_DATE} \
      -X main.GitCommit=${VCS_REF}" \
    -o rsfn-bridge \
    ./cmd/bridge

# ============================================
# Stage 2: Java Builder (XML Signer)
# ============================================
FROM maven:3.9-eclipse-temurin-21-alpine AS java-builder

WORKDIR /build

# Copy Maven project
COPY xml-signer/pom.xml ./
RUN mvn dependency:go-offline

COPY xml-signer/src ./src
RUN mvn clean package -DskipTests

# ============================================
# Stage 3: Final (Distroless Java)
# ============================================
FROM gcr.io/distroless/java21-debian12:nonroot

LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.revision="${VCS_REF}"
LABEL org.opencontainers.image.title="RSFN Bridge"
LABEL org.opencontainers.image.description="SOAP/XML adapter for Bacen RSFN API"

# Copy Go binary
COPY --from=go-builder /build/rsfn-bridge /rsfn-bridge

# Copy Java JAR (XML signer)
COPY --from=java-builder /build/target/xml-signer.jar /opt/signer/xml-signer.jar

# Copy CA certificates
COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=go-builder /usr/share/zoneinfo /usr/share/zoneinfo

EXPOSE 50051 9090

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD ["/rsfn-bridge", "healthcheck"]

USER nonroot:nonroot

ENTRYPOINT ["/rsfn-bridge"]
CMD ["serve"]
```

### Notes

```yaml
RSFN Bridge Complexity:
  - Requires both Go runtime and JVM
  - Go binary: gRPC server
  - Java JAR: XML signing with ICP-Brasil certificate
  - Image size: ~150-200MB (due to JVM)

Optimization:
  - Use distroless/java21 (minimal JVM)
  - Only include xml-signer.jar (not full Maven dependencies)
  - Consider GraalVM native-image for Java (reduces size to ~50MB)
```

---

## 6. Security Scanning

### Trivy Scan

```bash
#!/bin/bash
# trivy-scan.sh

set -euo pipefail

IMAGE=$1
SEVERITY=${SEVERITY:-HIGH,CRITICAL}

echo "Scanning image: ${IMAGE}"

# Scan for vulnerabilities
trivy image \
  --severity "${SEVERITY}" \
  --exit-code 1 \
  --no-progress \
  --format json \
  --output trivy-report.json \
  "${IMAGE}"

# Generate SARIF report (for GitHub Security tab)
trivy image \
  --severity "${SEVERITY}" \
  --format sarif \
  --output trivy-results.sarif \
  "${IMAGE}"

# Generate SBOM (Software Bill of Materials)
trivy image \
  --format cyclonedx \
  --output sbom.json \
  "${IMAGE}"

echo "Scan complete. Reports:"
echo "  - trivy-report.json"
echo "  - trivy-results.sarif"
echo "  - sbom.json"

# Parse results
VULNERABILITIES=$(jq '.Results[].Vulnerabilities | length' trivy-report.json)
echo "Total vulnerabilities found: ${VULNERABILITIES}"

if [ "${VULNERABILITIES}" -gt 0 ]; then
  echo "ERROR: Vulnerabilities found!"
  jq '.Results[].Vulnerabilities[] | {ID: .VulnerabilityID, Severity: .Severity, Package: .PkgName}' trivy-report.json
  exit 1
fi

echo "No vulnerabilities found!"
```

### Grype Scan (Alternative)

```bash
#!/bin/bash
# grype-scan.sh

IMAGE=$1

grype "${IMAGE}" \
  --scope all-layers \
  --output json \
  --file grype-report.json

grype "${IMAGE}" \
  --scope all-layers \
  --output sarif \
  --file grype-results.sarif

# Fail on HIGH or CRITICAL
grype "${IMAGE}" \
  --fail-on high \
  --scope all-layers
```

### Vulnerability Remediation

```yaml
Remediation Process:

1. Identify vulnerability:
   - CVE-XXXX-YYYY
   - Package: libfoo v1.2.3
   - Severity: HIGH

2. Check if exploitable:
   - Review CVE details
   - Assess impact on DICT services
   - Check if package is actually used

3. Remediation options:
   a. Update base image:
      - FROM gcr.io/distroless/static-debian12:nonroot
      - Check for newer version

   b. Update dependency:
      - go get -u vulnerable/package@latest
      - mvn dependency:tree

   c. Accept risk (document):
      - Add to .trivyignore with justification
      - # CVE-XXXX-YYYY: Not exploitable because...

4. Re-scan:
   - trivy image ...
   - Verify fix

5. Deploy:
   - Build new image
   - Push to ECR
   - Update K8s manifests
```

---

## 7. Image Signing

### Cosign Setup

```bash
#!/bin/bash
# setup-cosign.sh

# Install cosign
curl -LO https://github.com/sigstore/cosign/releases/latest/download/cosign-linux-amd64
sudo mv cosign-linux-amd64 /usr/local/bin/cosign
sudo chmod +x /usr/local/bin/cosign

# Generate key pair (once)
cosign generate-key-pair

# Output:
#   - cosign.key (private key - store in AWS Secrets Manager)
#   - cosign.pub (public key - distribute to clusters)
```

### Sign Image

```bash
#!/bin/bash
# sign-image.sh

set -euo pipefail

IMAGE=$1
COSIGN_KEY=${COSIGN_KEY:-cosign.key}

echo "Signing image: ${IMAGE}"

# Sign image
cosign sign \
  --key "${COSIGN_KEY}" \
  --tlog-upload=false \
  "${IMAGE}"

# Verify signature
cosign verify \
  --key cosign.pub \
  "${IMAGE}"

echo "Image signed successfully!"
```

### Verify Signature (Kubernetes Admission Controller)

```yaml
# Use Kyverno or OPA Gatekeeper to verify signatures

apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: verify-image-signature
spec:
  validationFailureAction: enforce
  background: false
  webhookTimeoutSeconds: 30
  failurePolicy: Fail
  rules:
  - name: verify-signature
    match:
      any:
      - resources:
          kinds:
          - Pod
    verifyImages:
    - imageReferences:
      - "123456789012.dkr.ecr.us-east-1.amazonaws.com/lbpay/*:*"
      attestors:
      - count: 1
        entries:
        - keys:
            publicKeys: |
              -----BEGIN PUBLIC KEY-----
              <cosign.pub content>
              -----END PUBLIC KEY-----
```

---

## 8. Registry Strategy

### AWS ECR Configuration

```bash
#!/bin/bash
# setup-ecr.sh

set -euo pipefail

AWS_REGION="us-east-1"
AWS_ACCOUNT_ID="123456789012"

# Create repositories
REPOS=(
  "lbpay/core-dict"
  "lbpay/connect-api"
  "lbpay/connect-worker"
  "lbpay/rsfn-bridge"
)

for REPO in "${REPOS[@]}"; do
  echo "Creating repository: ${REPO}"

  aws ecr create-repository \
    --repository-name "${REPO}" \
    --region "${AWS_REGION}" \
    --image-scanning-configuration scanOnPush=true \
    --encryption-configuration encryptionType=AES256 \
    --tags \
      Key=Project,Value=DICT \
      Key=Team,Value=LBPay \
      Key=Environment,Value=production

  # Set lifecycle policy (keep last 30 images)
  aws ecr put-lifecycle-policy \
    --repository-name "${REPO}" \
    --region "${AWS_REGION}" \
    --lifecycle-policy-text '{
      "rules": [
        {
          "rulePriority": 1,
          "description": "Keep last 30 images",
          "selection": {
            "tagStatus": "any",
            "countType": "imageCountMoreThan",
            "countNumber": 30
          },
          "action": {
            "type": "expire"
          }
        }
      ]
    }'

  # Set image tag immutability
  aws ecr put-image-tag-mutability \
    --repository-name "${REPO}" \
    --region "${AWS_REGION}" \
    --image-tag-mutability IMMUTABLE
done

echo "ECR repositories created!"
```

### Image Tagging Strategy

```yaml
Tagging Convention:

1. Git SHA (immutable):
   - lbpay/core-dict:abc123def456
   - Used for: Production deployments
   - Benefit: Reproducible, traceable

2. Semantic Version (immutable):
   - lbpay/core-dict:v1.2.3
   - lbpay/core-dict:v1.2
   - lbpay/core-dict:v1
   - Used for: Releases
   - Benefit: Human-readable

3. Branch name (mutable):
   - lbpay/core-dict:main
   - lbpay/core-dict:develop
   - lbpay/core-dict:feature-abc
   - Used for: Development
   - Benefit: Easy testing

4. Latest (mutable):
   - lbpay/core-dict:latest
   - Used for: Local development only
   - Avoid in production!

Example:
  Git commit: abc123def456
  Git tag: v1.2.3
  Branch: main

  Tags created:
    - lbpay/core-dict:abc123def456
    - lbpay/core-dict:v1.2.3
    - lbpay/core-dict:v1.2
    - lbpay/core-dict:v1
    - lbpay/core-dict:main
    - lbpay/core-dict:latest
```

### Pull Image from ECR

```bash
#!/bin/bash
# pull-from-ecr.sh

AWS_REGION="us-east-1"
ECR_REGISTRY="123456789012.dkr.ecr.us-east-1.amazonaws.com"
IMAGE="lbpay/core-dict:v1.2.3"

# Login to ECR
aws ecr get-login-password --region "${AWS_REGION}" | \
  docker login --username AWS --password-stdin "${ECR_REGISTRY}"

# Pull image
docker pull "${ECR_REGISTRY}/${IMAGE}"

# Verify signature (if using cosign)
cosign verify \
  --key cosign.pub \
  "${ECR_REGISTRY}/${IMAGE}"
```

---

## 9. Rastreabilidade

### Documentos Relacionados

| ID | Documento | Relacao |
|----|-----------|---------|
| **DEV-001** | [CI/CD Pipeline Core](./Pipelines/DEV-001_CI_CD_Pipeline_Core.md) | Pipeline que usa Dockerfiles |
| **DEV-004** | [Kubernetes Manifests](./DEV-004_Kubernetes_Manifests.md) | Manifests que referenciam imagens |
| **DEV-005** | [Monitoring & Observability](./DEV-005_Monitoring_Observability.md) | Observabilidade de containers |
| **SEC-001** | [mTLS Configuration](../13_Seguranca/SEC-001_mTLS_Configuration.md) | Certificados em containers |

### Metricas de Sucesso

```yaml
Image Quality:
  - Image size: < 50MB (Go services)
  - Image size: < 200MB (RSFN Bridge with JVM)
  - Build time: < 5 minutos
  - Vulnerabilities: 0 HIGH/CRITICAL

Security:
  - All images signed with cosign
  - All images scanned with Trivy
  - SBOM generated for all images
  - Image tag immutability enabled

Registry:
  - ECR lifecycle policy: keep 30 images
  - Image scanning on push: enabled
  - Encryption: AES256
  - Replication: multi-region (optional)
```

---

**Ultima Atualizacao**: 2025-10-25
**Versao**: 1.0
**Status**: Completo
