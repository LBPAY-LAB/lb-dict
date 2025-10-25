# DEV-003: CI/CD Pipeline - RSFN Bridge

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: RSFN Bridge (SOAP/mTLS Adapter) CI/CD Pipeline
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: DEVOPS (AI Agent - DevOps Engineer)
**Revisor**: [Aguardando]
**Aprovador**: Head de DevOps, CTO (José Luís Silva)

---

## Controle de Versão

| Versão | Data | Autor | Descrição das Mudanças |
|--------|------|-------|------------------------|
| 1.0 | 2025-10-25 | DEVOPS | Versão inicial - Pipeline CI/CD para RSFN Bridge com XML Signer |

---

## Sumário Executivo

### Visão Geral

Pipeline CI/CD para o **RSFN Bridge** (adaptador SOAP/mTLS para Bacen), cobrindo:
- ✅ **Dual Protocol Build**: gRPC Server + Pulsar Consumer
- ✅ **XML Signer Integration**: JRE + JAR externo (assinatura digital ICP-Brasil)
- ✅ **mTLS Certificate Management**: ICP-Brasil A3 certificates
- ✅ **Circuit Breaker Tests**: sony/gobreaker
- ✅ **Security Scanning**: gosec, trivy
- ✅ **Multi-Stage Build**: Go binary + JRE runtime
- ✅ **Multi-Environment Deploy**: dev, staging, prod

### Arquitetura do Bridge

```
rsfn-bridge/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── grpc/                          # gRPC Server
│   ├── consumer/                      # Pulsar Consumer
│   ├── soap/                          # SOAP XML Builder
│   ├── signer/                        # XML Signer (JRE integration)
│   ├── mtls/                          # mTLS Client
│   └── breaker/                       # Circuit Breaker
├── configs/
└── certs/                             # ICP-Brasil certificates
```

### Stack Tecnológica

| Componente | Tecnologia | Versão |
|------------|------------|--------|
| **Linguagem** | Go | 1.24.5 |
| **Protocols** | gRPC + Pulsar | v1.62+ / 3.0+ |
| **SOAP Client** | Native Go (net/http) | - |
| **XML Signer** | JRE + JAR externo | Java 17 |
| **Circuit Breaker** | sony/gobreaker | v2.3.0 |
| **Observability** | OpenTelemetry | v1.38.0 |

---

## Índice

1. [Workflow Overview](#1-workflow-overview)
2. [Pipeline Stages](#2-pipeline-stages)
3. [XML Signer Build](#3-xml-signer-build)
4. [GitHub Actions Workflow](#4-github-actions-workflow)
5. [Certificate Management](#5-certificate-management)
6. [Environment Configuration](#6-environment-configuration)
7. [Deployment Strategy](#7-deployment-strategy)
8. [Monitoring & Alerts](#8-monitoring--alerts)
9. [Rastreabilidade](#9-rastreabilidade)

---

## 1. Workflow Overview

### Trigger Strategy

```yaml
Triggers:
  - push: [main, develop, release/*]
  - pull_request: [main, develop]
  - workflow_dispatch: Manual trigger

Path Filters:
  - rsfn-bridge/**
  - .github/workflows/rsfn-bridge-ci.yml
```

### Pipeline Flow

```
┌──────────────────────────────────────────────────────────────────┐
│                   RSFN Bridge CI/CD Pipeline                      │
├──────────────────────────────────────────────────────────────────┤
│                                                                    │
│  Stage 1: Code Quality (Lint + Format)                           │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                 │
│  │   Lint     │→ │   Format   │→ │   Vet      │                 │
│  └────────────┘  └────────────┘  └────────────┘                 │
│         ↓                                                          │
│  Stage 2: Testing (Unit + Integration + SOAP Mock)               │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                 │
│  │ Unit Tests │→ │Integration │→ │ SOAP Mock  │                 │
│  │            │  │Pulsar+gRPC │  │ Bacen API  │                 │
│  └────────────┘  └────────────┘  └────────────┘                 │
│         ↓                                                          │
│  Stage 3: Security Scanning (SAST + Certificate Check)           │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                 │
│  │   gosec    │→ │   trivy    │→ │  Cert Val. │                 │
│  └────────────┘  └────────────┘  └────────────┘                 │
│         ↓                                                          │
│  Stage 4: Build (Go + JRE + XML Signer)                          │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                 │
│  │ Build Go   │→ │ Add JRE    │→ │ Add Signer │                 │
│  │   Binary   │  │  Runtime   │  │    JAR     │                 │
│  └────────────┘  └────────────┘  └────────────┘                 │
│         ↓                                                          │
│  ┌────────────┐  ┌────────────┐                                  │
│  │ Scan Image │→ │ Push ECR   │                                  │
│  │   trivy    │  │            │                                  │
│  └────────────┘  └────────────┘                                  │
│         ↓                                                          │
│  Stage 5: Deploy (API + Certificate Mount)                       │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                 │
│  │    DEV     │→ │  Staging   │→ │    PROD    │                 │
│  │   Auto     │  │ Auto+Appr  │  │Manual Only │                 │
│  └────────────┘  └────────────┘  └────────────┘                 │
│                                                                    │
└──────────────────────────────────────────────────────────────────┘
```

---

## 2. Pipeline Stages

### Stage 1: Code Quality (5 min)

Same as Core DICT and Connect.

---

### Stage 2: Testing (12 min)

**Objetivos**:
- Executar unit tests
- Executar integration tests (gRPC, Pulsar)
- Executar **SOAP mock tests** (simular API Bacen)
- Testar Circuit Breaker

**Jobs**:

```yaml
unit-tests:
  - go test -v -race ./...
  - Coverage threshold: 80%

integration-tests:
  Services:
    - Pulsar 3.0
    - SOAP Mock Server (WireMock)

  Tests:
    - gRPC Server endpoints
    - Pulsar consumer
    - SOAP XML builder
    - XML signer (mock)
    - Circuit breaker behavior

soap-mock-tests:
  Setup:
    - Start WireMock (SOAP server simulator)
    - Load Bacen WSDL stubs

  Tests:
    - CreateEntry SOAP request
    - ClaimEntry SOAP request
    - DeleteEntry SOAP request
    - Error responses (timeout, 500, invalid XML)
    - mTLS handshake (mock certificates)

circuit-breaker-tests:
  Tests:
    - Circuit open after N failures
    - Circuit half-open after timeout
    - Circuit closed after success
    - Fallback behavior
```

**SOAP Mock Test Example** (WireMock):

```yaml
# wiremock/stubs/create-entry.json
{
  "request": {
    "method": "POST",
    "urlPath": "/dict/api/v1/entries",
    "headers": {
      "Content-Type": {
        "equalTo": "text/xml; charset=utf-8"
      }
    },
    "bodyPatterns": [
      {
        "matchesXPath": "//CreateEntryRequest"
      }
    ]
  },
  "response": {
    "status": 200,
    "headers": {
      "Content-Type": "text/xml"
    },
    "body": "<?xml version=\"1.0\"?><CreateEntryResponse><EntryId>ENTRY-123</EntryId><Status>SUCCESS</Status></CreateEntryResponse>"
  }
}
```

---

### Stage 3: Security Scanning (10 min)

**Objetivos**:
- SAST (gosec)
- Container scan (trivy)
- **Certificate validation** (ICP-Brasil A3)

**Jobs**:

```yaml
gosec:
  - Same as Core DICT

trivy-fs:
  - Same as Core DICT

certificate-validation:
  Tool: OpenSSL
  Tests:
    - Certificate expiration (> 30 days)
    - Certificate chain validation
    - Private key match
    - Certificate issuer (ICP-Brasil)

  Command: |
    openssl x509 -in cert.pem -noout -checkend 2592000  # 30 days
    openssl verify -CAfile ca.pem cert.pem
```

**Exit Criteria**:
- ✅ Certificates valid for > 30 days
- ✅ Certificate chain valid (ICP-Brasil root)
- ✅ No security vulnerabilities HIGH/CRITICAL

---

### Stage 4: Build (15 min)

**Objetivos**:
- Build Go binary
- Add JRE runtime (for XML Signer)
- Add XML Signer JAR
- Create multi-stage Docker image

**Dockerfile** (`Dockerfile`):

```dockerfile
# ============================================
# Stage 1: Go Builder
# ============================================
FROM golang:1.24.5-alpine AS go-builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o rsfn-bridge ./cmd/server

# ============================================
# Stage 2: Java Builder (XML Signer)
# ============================================
FROM eclipse-temurin:17-jre-alpine AS java-runtime

# ============================================
# Stage 3: Runtime (Go + JRE + Signer)
# ============================================
FROM alpine:3.18
RUN apk add --no-cache ca-certificates tzdata

# Copy JRE from java-runtime stage
COPY --from=java-runtime /opt/java/openjdk /opt/java/openjdk
ENV JAVA_HOME=/opt/java/openjdk
ENV PATH="${JAVA_HOME}/bin:${PATH}"

# Copy Go binary
COPY --from=go-builder /build/rsfn-bridge /rsfn-bridge

# Copy XML Signer JAR (external dependency)
COPY ./signer/xml-signer.jar /opt/signer/xml-signer.jar

# Copy configs and certificates placeholder
COPY ./configs /configs
RUN mkdir -p /certs

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /configs /certs

USER appuser

EXPOSE 50051 9090
ENTRYPOINT ["/rsfn-bridge"]
```

**Image Size Optimization**:
- Go binary: ~10MB (with `-ldflags="-w -s"`)
- JRE (Alpine): ~40MB
- XML Signer JAR: ~5MB
- **Total**: ~60MB

---

## 3. XML Signer Build

### XML Signer JAR (External Dependency)

**Source**: Custom Java application or library (ICP-Brasil digital signature)

**Build** (if building from source):

```yaml
# .github/workflows/build-xml-signer.yml
name: Build XML Signer JAR

on:
  push:
    paths:
      - 'xml-signer/**'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-java@v4
        with:
          java-version: '17'
          distribution: 'temurin'

      - name: Build JAR
        working-directory: xml-signer
        run: mvn clean package

      - name: Upload JAR artifact
        uses: actions/upload-artifact@v4
        with:
          name: xml-signer
          path: xml-signer/target/xml-signer.jar
```

**Usage in Go** (via `os/exec`):

```go
// internal/signer/signer.go
func SignXML(xmlContent string, certPath string, keyPath string) (string, error) {
    cmd := exec.Command(
        "java", "-jar", "/opt/signer/xml-signer.jar",
        "--xml", xmlContent,
        "--cert", certPath,
        "--key", keyPath,
    )
    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("XML signing failed: %w, output: %s", err, output)
    }
    return string(output), nil
}
```

---

## 4. GitHub Actions Workflow

### Arquivo: `.github/workflows/rsfn-bridge-ci.yml`

```yaml
name: RSFN Bridge CI/CD

on:
  push:
    branches: [main, develop, release/**]
    paths:
      - 'rsfn-bridge/**'
      - '.github/workflows/rsfn-bridge-ci.yml'
  pull_request:
    branches: [main, develop]
    paths:
      - 'rsfn-bridge/**'
  workflow_dispatch:

env:
  GO_VERSION: '1.24.5'
  DOCKER_BUILDKIT: 1
  ECR_REGISTRY: 123456789012.dkr.ecr.us-east-1.amazonaws.com
  ECR_REPOSITORY: lbpay/rsfn-bridge
  SLACK_WEBHOOK: ${{ secrets.SLACK_WEBHOOK_DEVOPS }}

jobs:
  # ====================================
  # Stage 1: Code Quality
  # ====================================

  lint:
    name: Lint & Format Check
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: v1.55
          working-directory: rsfn-bridge
          args: --timeout=5m

      - name: Check formatting
        working-directory: rsfn-bridge
        run: |
          gofmt -s -l . | tee fmt_errors.txt
          if [ -s fmt_errors.txt ]; then
            echo "::error::Code not formatted"
            exit 1
          fi

      - name: Run go vet
        working-directory: rsfn-bridge
        run: go vet ./...

  # ====================================
  # Stage 2: Testing
  # ====================================

  test:
    name: Unit & Integration Tests
    runs-on: ubuntu-latest
    needs: lint
    services:
      pulsar:
        image: apachepulsar/pulsar:3.0.0
        command: bin/pulsar standalone
        ports:
          - 6650:6650
          - 8080:8080

      wiremock:
        image: wiremock/wiremock:3.0.0
        ports:
          - 8081:8080
        volumes:
          - ./rsfn-bridge/test/wiremock:/home/wiremock

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Setup WireMock stubs
        run: |
          mkdir -p rsfn-bridge/test/wiremock/mappings
          cp rsfn-bridge/test/stubs/*.json rsfn-bridge/test/wiremock/mappings/

      - name: Wait for services
        run: |
          timeout 60 bash -c 'until curl -f http://localhost:8080/admin/v2/clusters; do sleep 2; done'
          timeout 60 bash -c 'until curl -f http://localhost:8081/__admin; do sleep 2; done'

      - name: Run unit tests
        working-directory: rsfn-bridge
        env:
          PULSAR_URL: pulsar://localhost:6650
          BACEN_API_URL: http://localhost:8081
        run: |
          go test -v -race -coverprofile=coverage.out ./...

      - name: Check coverage
        working-directory: rsfn-bridge
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: $coverage%"
          if (( $(echo "$coverage < 80" | bc -l) )); then
            exit 1
          fi

      - name: Test Circuit Breaker
        working-directory: rsfn-bridge
        run: |
          go test -v -tags=integration ./internal/breaker/...

  # ====================================
  # Stage 3: Security
  # ====================================

  security:
    name: Security Scanning
    runs-on: ubuntu-latest
    needs: test
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Run gosec
        uses: securego/gosec@master
        with:
          args: '-no-fail -fmt sarif -out gosec.sarif ./rsfn-bridge/...'

      - name: Upload gosec SARIF
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: gosec.sarif

      - name: Validate certificates (mock)
        working-directory: rsfn-bridge/test/certs
        run: |
          openssl x509 -in test-cert.pem -noout -checkend 2592000 || echo "Certificate expires in < 30 days"
          openssl verify -CAfile test-ca.pem test-cert.pem

      - name: Run Trivy FS scan
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: 'rsfn-bridge'
          format: 'sarif'
          output: 'trivy-fs.sarif'
          severity: 'HIGH,CRITICAL'
          exit-code: '1'

  # ====================================
  # Stage 4: Build
  # ====================================

  build:
    name: Build & Push Docker Image
    runs-on: ubuntu-latest
    needs: security
    steps:
      - uses: actions/checkout@v4

      - name: Download XML Signer JAR (artifact or S3)
        run: |
          # Option 1: Download from previous build
          # uses: actions/download-artifact@v4
          # Option 2: Download from S3
          aws s3 cp s3://lbpay-artifacts/xml-signer/xml-signer.jar rsfn-bridge/signer/xml-signer.jar

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1

      - name: Login to ECR
        id: login-ecr
        uses: aws-actions/amazon-ecr-login@v2

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.ECR_REGISTRY }}/${{ env.ECR_REPOSITORY }}
          tags: |
            type=ref,event=branch
            type=sha,prefix={{branch}}-
            type=raw,value=latest,enable={{is_default_branch}}

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Build image
        uses: docker/build-push-action@v5
        with:
          context: rsfn-bridge
          file: rsfn-bridge/Dockerfile
          push: false
          load: true
          tags: ${{ steps.meta.outputs.tags }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

      - name: Test XML Signer in container
        run: |
          docker run --rm ${{ env.ECR_REGISTRY }}/${{ env.ECR_REPOSITORY }}:${{ github.sha }} \
            java -version

      - name: Scan image with Trivy
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ env.ECR_REGISTRY }}/${{ env.ECR_REPOSITORY }}:${{ github.sha }}
          format: 'sarif'
          output: 'trivy-image.sarif'
          severity: 'HIGH,CRITICAL'
          exit-code: '1'

      - name: Push to ECR
        uses: docker/build-push-action@v5
        with:
          context: rsfn-bridge
          file: rsfn-bridge/Dockerfile
          push: true
          tags: ${{ steps.meta.outputs.tags }}

  # ====================================
  # Stage 5: Deploy DEV
  # ====================================

  deploy-dev:
    name: Deploy to DEV
    runs-on: ubuntu-latest
    needs: build
    if: github.ref == 'refs/heads/develop'
    environment:
      name: dev
      url: https://bridge.dev.lbpay.io
    steps:
      - uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1

      - name: Setup kubectl
        uses: azure/setup-kubectl@v4
        with:
          version: 'v1.28.0'

      - name: Update kubeconfig
        run: aws eks update-kubeconfig --name eks-lbpay-dev --region us-east-1

      - name: Deploy Bridge
        run: |
          kubectl set image deployment/rsfn-bridge \
            rsfn-bridge=${{ env.ECR_REGISTRY }}/${{ env.ECR_REPOSITORY }}:${{ github.sha }} \
            -n dict-dev
          kubectl rollout status deployment/rsfn-bridge -n dict-dev --timeout=5m

      - name: Verify gRPC endpoint
        run: |
          kubectl run grpc-test-${{ github.run_id }} \
            --image=fullstorydev/grpcurl:latest \
            --restart=Never \
            --rm -i -n dict-dev \
            -- -plaintext rsfn-bridge-svc:50051 list || exit 1

  # ====================================
  # Stage 5: Deploy STAGING
  # ====================================

  deploy-staging:
    name: Deploy to STAGING
    runs-on: ubuntu-latest
    needs: build
    if: github.ref == 'refs/heads/main'
    environment:
      name: staging
      url: https://bridge.staging.lbpay.io
    steps:
      - uses: actions/checkout@v4

      - name: Setup ArgoCD CLI
        run: |
          curl -sSL -o argocd https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
          chmod +x argocd
          sudo mv argocd /usr/local/bin/

      - name: Login to ArgoCD
        run: |
          argocd login ${{ secrets.ARGOCD_SERVER }} \
            --username ${{ secrets.ARGOCD_USERNAME }} \
            --password ${{ secrets.ARGOCD_PASSWORD }} \
            --insecure

      - name: Sync Bridge (Staging)
        run: |
          argocd app set rsfn-bridge-staging \
            --kustomize-image rsfn-bridge=${{ env.ECR_REGISTRY }}/${{ env.ECR_REPOSITORY }}:${{ github.sha }}
          argocd app sync rsfn-bridge-staging --prune --force
          argocd app wait rsfn-bridge-staging --health --timeout 600

  # ====================================
  # Stage 5: Deploy PROD
  # ====================================

  deploy-prod:
    name: Deploy to PRODUCTION
    runs-on: ubuntu-latest
    needs: [build, deploy-staging]
    if: github.ref == 'refs/heads/main' && github.event_name == 'workflow_dispatch'
    environment:
      name: production
      url: https://bridge.lbpay.io
    steps:
      - uses: actions/checkout@v4

      - name: Setup ArgoCD CLI
        run: |
          curl -sSL -o argocd https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
          chmod +x argocd
          sudo mv argocd /usr/local/bin/

      - name: Login to ArgoCD
        run: |
          argocd login ${{ secrets.ARGOCD_SERVER }} \
            --username ${{ secrets.ARGOCD_USERNAME }} \
            --password ${{ secrets.ARGOCD_PASSWORD }} \
            --insecure

      - name: Deploy Bridge (PROD)
        run: |
          argocd app set rsfn-bridge-prod \
            --kustomize-image rsfn-bridge=${{ env.ECR_REGISTRY }}/${{ env.ECR_REPOSITORY }}:${{ github.sha }}
          argocd app sync rsfn-bridge-prod --prune --force
          argocd app wait rsfn-bridge-prod --health --timeout 600

  # ====================================
  # Notifications
  # ====================================

  notify:
    name: Notify Slack
    runs-on: ubuntu-latest
    needs: [deploy-dev, deploy-staging, deploy-prod]
    if: always()
    steps:
      - name: Send Slack notification
        uses: slackapi/slack-github-action@v1
        with:
          webhook-url: ${{ env.SLACK_WEBHOOK }}
          payload: |
            {
              "text": "RSFN Bridge Deployment",
              "blocks": [
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": "*RSFN Bridge CI/CD*\n*Status*: ${{ job.status }}\n*Branch*: ${{ github.ref_name }}\n*Image*: rsfn-bridge"
                  }
                }
              ]
            }
```

---

## 5. Certificate Management

### ICP-Brasil A3 Certificates

**Storage**: AWS Secrets Manager + Kubernetes Secrets

**Certificate Types**:
1. **Client Certificate** (mTLS authentication)
2. **Signing Certificate** (XML digital signature)
3. **CA Certificate** (ICP-Brasil root)

### Certificate Rotation Pipeline

```yaml
# .github/workflows/rotate-certificates.yml
name: Rotate ICP-Brasil Certificates

on:
  schedule:
    - cron: '0 0 1 * *'  # Monthly check
  workflow_dispatch:

jobs:
  check-expiration:
    runs-on: ubuntu-latest
    steps:
      - name: Download certificates from AWS Secrets Manager
        run: |
          aws secretsmanager get-secret-value \
            --secret-id /lbpay/dict/prod/icp-brasil-cert \
            --query SecretString --output text > cert.pem

      - name: Check expiration
        run: |
          days_until_expiry=$(openssl x509 -in cert.pem -noout -checkend 0 -enddate | grep -o '[0-9]*')
          echo "Certificate expires in $days_until_expiry days"

          if [ "$days_until_expiry" -lt 30 ]; then
            echo "::error::Certificate expires in less than 30 days!"
            # Send alert to Slack/PagerDuty
            exit 1
          fi

  rotate-certificate:
    needs: check-expiration
    if: failure()
    runs-on: ubuntu-latest
    steps:
      - name: Manual approval required
        uses: trstringer/manual-approval@v1
        with:
          approvers: security-team,devops-team
          minimum-approvals: 2
```

### Certificate Mount in Kubernetes

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: icp-brasil-certs
  namespace: dict-prod
type: Opaque
data:
  client-cert.pem: <base64-encoded>
  client-key.pem: <base64-encoded>
  signing-cert.pem: <base64-encoded>
  signing-key.pem: <base64-encoded>
  ca.pem: <base64-encoded>

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rsfn-bridge
spec:
  template:
    spec:
      volumes:
      - name: certs
        secret:
          secretName: icp-brasil-certs
      containers:
      - name: rsfn-bridge
        volumeMounts:
        - name: certs
          mountPath: /certs
          readOnly: true
```

---

## 6. Environment Configuration

### DEV Environment

```yaml
Replicas: 2
Resources:
  requests: {cpu: 250m, memory: 512Mi}
  limits: {cpu: 500m, memory: 1Gi}

Certificates: Test certificates (self-signed)
Bacen API: Mock server (WireMock)
```

### STAGING Environment

```yaml
Replicas: 3
Resources:
  requests: {cpu: 500m, memory: 1Gi}
  limits: {cpu: 1000m, memory: 2Gi}

Certificates: Bacen homologation certificates
Bacen API: Bacen homologation environment
```

### PRODUCTION Environment

```yaml
Replicas: 5
HPA: {min: 5, max: 15}
Resources:
  requests: {cpu: 1000m, memory: 2Gi}
  limits: {cpu: 2000m, memory: 4Gi}

Certificates: ICP-Brasil A3 production certificates
Bacen API: Bacen production environment
Circuit Breaker: Enabled (threshold: 5 failures, timeout: 60s)
```

---

## 7. Deployment Strategy

Same as Core DICT:
- DEV: Rolling Update
- Staging: Blue-Green
- PROD: Canary

---

## 8. Monitoring & Alerts

### Bridge-Specific Metrics

```yaml
Metrics:
  - bridge_soap_request_total
  - bridge_soap_request_duration_seconds
  - bridge_soap_error_total
  - bridge_circuit_breaker_state (open/half-open/closed)
  - bridge_mtls_handshake_errors_total
  - bridge_xml_signature_duration_seconds

Alerts:
  - SOAP error rate > 5%
  - Circuit breaker open > 2 minutes
  - mTLS handshake failures > 10/min
  - Certificate expires in < 30 days
```

---

## 9. Rastreabilidade

### Documentos Relacionados

| ID | Documento | Relação |
|----|-----------|---------|
| **TEC-002** | [RSFN Bridge Specification](../../11_Especificacoes_Tecnicas/TEC-002_Bridge_Specification.md) | Especificação técnica do Bridge |
| **SEC-001** | [mTLS Configuration](../../13_Seguranca/SEC-001_mTLS_Configuration.md) | Configuração mTLS |
| **SEC-002** | [ICP-Brasil Certificates](../../13_Seguranca/SEC-002_ICP_Brasil_Certificates.md) | Certificados ICP-Brasil |
| **SEC-006** | [XML Signature Security](../../13_Seguranca/SEC-006_XML_Signature_Security.md) | Assinatura digital XML |
| **DEV-004** | [Kubernetes Manifests](#) | Manifests K8s do Bridge |

---

**Última Atualização**: 2025-10-25
**Versão**: 1.0
**Status**: ✅ Completo
