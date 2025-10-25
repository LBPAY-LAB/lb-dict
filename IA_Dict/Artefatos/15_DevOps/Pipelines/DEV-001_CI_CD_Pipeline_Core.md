# DEV-001: CI/CD Pipeline - Core DICT

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: Core DICT CI/CD Pipeline
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: DEVOPS (AI Agent - DevOps Engineer)
**Revisor**: [Aguardando]
**Aprovador**: Head de DevOps, CTO (José Luís Silva)

---

## Controle de Versão

| Versão | Data | Autor | Descrição das Mudanças |
|--------|------|-------|------------------------|
| 1.0 | 2025-10-25 | DEVOPS | Versão inicial - Pipeline CI/CD completo para Core DICT |

---

## Sumário Executivo

### Visão Geral

Pipeline CI/CD para o **Core DICT** implementado em **GitHub Actions**, cobrindo:
- ✅ **Lint & Format** (golangci-lint, gofmt)
- ✅ **Unit Tests** (go test + coverage)
- ✅ **Integration Tests** (testcontainers)
- ✅ **Security Scanning** (gosec, trivy)
- ✅ **Build & Push Docker** (multi-stage build)
- ✅ **Deploy Multi-Environment** (dev, staging, prod)
- ✅ **Notification** (Slack alerts)

### Tecnologias

| Componente | Tecnologia | Versão |
|------------|------------|--------|
| **CI/CD** | GitHub Actions | v4 |
| **Linguagem** | Go | 1.24.5 |
| **Container** | Docker | 24+ |
| **Registry** | AWS ECR | - |
| **Orchestration** | Kubernetes | 1.28+ |
| **Deployment** | kubectl + ArgoCD | - |

---

## Índice

1. [Workflow Overview](#1-workflow-overview)
2. [Pipeline Stages](#2-pipeline-stages)
3. [GitHub Actions Workflow](#3-github-actions-workflow)
4. [Environment Configuration](#4-environment-configuration)
5. [Security Scanning](#5-security-scanning)
6. [Deployment Strategy](#6-deployment-strategy)
7. [Rollback Strategy](#7-rollback-strategy)
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
  - core-dict/**
  - .github/workflows/core-dict-ci.yml
```

### Pipeline Flow

```
┌──────────────────────────────────────────────────────────────────┐
│                      GitHub Actions Workflow                      │
├──────────────────────────────────────────────────────────────────┤
│                                                                    │
│  Stage 1: Code Quality                                            │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                 │
│  │   Lint     │→ │   Format   │→ │   Vet      │                 │
│  │ golangci   │  │   gofmt    │  │   go vet   │                 │
│  └────────────┘  └────────────┘  └────────────┘                 │
│         ↓                                                          │
│  Stage 2: Testing                                                 │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                 │
│  │ Unit Tests │→ │Integration │→ │  Coverage  │                 │
│  │  go test   │  │testcontainer│  │   >80%    │                 │
│  └────────────┘  └────────────┘  └────────────┘                 │
│         ↓                                                          │
│  Stage 3: Security                                                │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                 │
│  │   gosec    │→ │   trivy    │→ │  Dependency│                 │
│  │   SAST     │  │  Container │  │   Check    │                 │
│  └────────────┘  └────────────┘  └────────────┘                 │
│         ↓                                                          │
│  Stage 4: Build                                                   │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                 │
│  │Docker Build│→ │ Image Scan │→ │ Push to ECR│                 │
│  │multi-stage │  │   trivy    │  │            │                 │
│  └────────────┘  └────────────┘  └────────────┘                 │
│         ↓                                                          │
│  Stage 5: Deploy                                                  │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                 │
│  │    DEV     │→ │  Staging   │→ │    PROD    │                 │
│  │   Auto     │  │ Auto+Approval│ │Manual Only │                 │
│  └────────────┘  └────────────┘  └────────────┘                 │
│                                                                    │
└──────────────────────────────────────────────────────────────────┘
```

---

## 2. Pipeline Stages

### Stage 1: Code Quality (5 min)

**Objetivos**:
- Validar código Go segue padrões
- Detectar code smells
- Garantir formatação consistente

**Jobs**:

```yaml
lint:
  - golangci-lint v1.55+
  - Checks: errcheck, gosimple, govet, ineffassign, staticcheck

format:
  - gofmt -s -w .
  - Fail se houver alterações

vet:
  - go vet ./...
  - Detecta bugs em potencial
```

**Exit Criteria**:
- ✅ Sem erros de linting
- ✅ Código formatado corretamente
- ✅ go vet sem warnings

---

### Stage 2: Testing (10 min)

**Objetivos**:
- Executar unit tests
- Executar integration tests
- Garantir coverage > 80%

**Jobs**:

```yaml
unit-tests:
  - go test -v -race -coverprofile=coverage.out ./...
  - Coverage threshold: 80%
  - Upload coverage to Codecov

integration-tests:
  - Uses: testcontainers (PostgreSQL, Redis, Pulsar)
  - Tests: gRPC APIs, Database operations, Event publishing
  - Timeout: 10 minutes
```

**Exit Criteria**:
- ✅ Todos os testes passam
- ✅ Coverage >= 80%
- ✅ Sem race conditions

---

### Stage 3: Security Scanning (8 min)

**Objetivos**:
- Detectar vulnerabilidades no código (SAST)
- Escanear dependências Go
- Validar conformidade de segurança

**Jobs**:

```yaml
gosec:
  Tool: securego/gosec
  Config: .gosec.yaml
  Rules:
    - G101: Hardcoded credentials
    - G102: Bind to all interfaces
    - G104: Unhandled errors
    - G401: Weak crypto
    - G402: TLS InsecureSkipVerify

dependency-check:
  Tool: nancy (Sonatype)
  Command: go list -json -m all | nancy sleuth
  Action: Fail on HIGH/CRITICAL

trivy-fs:
  Tool: aquasecurity/trivy
  Target: Filesystem scan
  Severity: HIGH,CRITICAL
  Action: Fail pipeline if found
```

**Exit Criteria**:
- ✅ Sem vulnerabilidades HIGH/CRITICAL
- ✅ Dependências atualizadas
- ✅ Conformidade com CWE Top 25

---

### Stage 4: Build & Push (12 min)

**Objetivos**:
- Build de imagem Docker otimizada
- Escanear imagem final
- Push para AWS ECR

**Dockerfile Multi-Stage**:

```dockerfile
# Stage 1: Builder
FROM golang:1.24.5-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o core-dict ./cmd/server

# Stage 2: Runtime
FROM gcr.io/distroless/static-debian11
COPY --from=builder /build/core-dict /
COPY --from=builder /build/configs /configs
EXPOSE 8080 9090 50051
USER nonroot:nonroot
ENTRYPOINT ["/core-dict"]
```

**Jobs**:

```yaml
docker-build:
  - docker build -t core-dict:${{ github.sha }} .
  - docker tag core-dict:${{ github.sha }} core-dict:latest

trivy-image:
  - trivy image --severity HIGH,CRITICAL core-dict:${{ github.sha }}
  - Generate SBOM (Software Bill of Materials)

push-ecr:
  - aws ecr get-login-password | docker login
  - docker push $ECR_REGISTRY/core-dict:${{ github.sha }}
  - docker push $ECR_REGISTRY/core-dict:${{ github.ref_name }}
```

**Image Tags**:
- `{sha}`: Commit SHA (immutable)
- `{branch}`: Branch name (mutable)
- `latest`: Latest build (mutable)
- `v{semver}`: Release tags (immutable)

**Exit Criteria**:
- ✅ Build sem erros
- ✅ Imagem < 50MB
- ✅ Sem vulnerabilidades na imagem
- ✅ Push para ECR com sucesso

---

### Stage 5: Deploy (5-10 min)

**Deployment Strategy por Ambiente**:

| Ambiente | Trigger | Approval | Strategy | Rollback |
|----------|---------|----------|----------|----------|
| **DEV** | Auto (push develop) | Não | Rolling Update | Auto (health check fail) |
| **Staging** | Auto (push main) + Manual | 1 aprovador | Blue-Green | Manual |
| **PROD** | Manual | 2 aprovadores | Canary (10%→50%→100%) | Manual |

**Deployment Flow**:

```yaml
deploy-dev:
  Environment: dev
  Cluster: eks-lbpay-dev
  Namespace: dict-dev
  Method: kubectl apply

deploy-staging:
  Environment: staging
  Cluster: eks-lbpay-staging
  Namespace: dict-staging
  Method: ArgoCD sync
  Approval: 1 tech lead

deploy-prod:
  Environment: production
  Cluster: eks-lbpay-prod
  Namespace: dict-prod
  Method: ArgoCD sync + Canary
  Approval: 2 (tech lead + ops manager)
```

**Exit Criteria**:
- ✅ Deploy completo sem erros
- ✅ Health checks passando
- ✅ Pods em estado Running
- ✅ Smoke tests passando

---

## 3. GitHub Actions Workflow

### Arquivo: `.github/workflows/core-dict-ci.yml`

```yaml
name: Core DICT CI/CD

on:
  push:
    branches: [main, develop, release/**]
    paths:
      - 'core-dict/**'
      - '.github/workflows/core-dict-ci.yml'
  pull_request:
    branches: [main, develop]
    paths:
      - 'core-dict/**'
  workflow_dispatch:
    inputs:
      environment:
        description: 'Target environment'
        required: true
        type: choice
        options:
          - dev
          - staging
          - production

env:
  GO_VERSION: '1.24.5'
  DOCKER_BUILDKIT: 1
  ECR_REGISTRY: 123456789012.dkr.ecr.us-east-1.amazonaws.com
  ECR_REPOSITORY: lbpay/core-dict
  SLACK_WEBHOOK: ${{ secrets.SLACK_WEBHOOK_DEVOPS }}

jobs:
  # ====================================
  # Stage 1: Code Quality
  # ====================================

  lint:
    name: Lint & Format Check
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: v1.55
          working-directory: core-dict
          args: --timeout=5m --config=.golangci.yml

      - name: Check formatting
        working-directory: core-dict
        run: |
          gofmt -s -l . | tee fmt_errors.txt
          if [ -s fmt_errors.txt ]; then
            echo "::error::Code is not formatted. Run 'gofmt -s -w .'"
            exit 1
          fi

      - name: Run go vet
        working-directory: core-dict
        run: go vet ./...

  # ====================================
  # Stage 2: Testing
  # ====================================

  test:
    name: Unit & Integration Tests
    runs-on: ubuntu-latest
    needs: lint
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_PASSWORD: testpass
          POSTGRES_DB: core_dict_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Download dependencies
        working-directory: core-dict
        run: go mod download

      - name: Run unit tests
        working-directory: core-dict
        env:
          DATABASE_URL: postgres://postgres:testpass@localhost:5432/core_dict_test?sslmode=disable
          REDIS_URL: redis://localhost:6379
        run: |
          go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
          go tool cover -func=coverage.out | tail -n 1

      - name: Check coverage threshold
        working-directory: core-dict
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: $coverage%"
          if (( $(echo "$coverage < 80" | bc -l) )); then
            echo "::error::Coverage $coverage% is below 80% threshold"
            exit 1
          fi

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          files: core-dict/coverage.out
          flags: core-dict
          name: core-dict-coverage

  # ====================================
  # Stage 3: Security Scanning
  # ====================================

  security:
    name: Security Scanning
    runs-on: ubuntu-latest
    needs: test
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Run gosec (SAST)
        uses: securego/gosec@master
        with:
          args: '-no-fail -fmt sarif -out gosec-results.sarif ./core-dict/...'

      - name: Upload gosec SARIF
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: gosec-results.sarif

      - name: Check dependencies (nancy)
        working-directory: core-dict
        run: |
          go install github.com/sonatype-nexus-community/nancy@latest
          go list -json -m all | nancy sleuth

      - name: Run Trivy filesystem scan
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: 'core-dict'
          format: 'sarif'
          output: 'trivy-results.sarif'
          severity: 'HIGH,CRITICAL'
          exit-code: '1'

      - name: Upload Trivy SARIF
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: trivy-results.sarif

  # ====================================
  # Stage 4: Build & Push
  # ====================================

  build:
    name: Build & Push Docker Image
    runs-on: ubuntu-latest
    needs: security
    outputs:
      image-tag: ${{ steps.meta.outputs.tags }}
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1

      - name: Login to Amazon ECR
        id: login-ecr
        uses: aws-actions/amazon-ecr-login@v2

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.ECR_REGISTRY }}/${{ env.ECR_REPOSITORY }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha,prefix={{branch}}-
            type=raw,value=latest,enable={{is_default_branch}}

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Build Docker image
        uses: docker/build-push-action@v5
        with:
          context: core-dict
          file: core-dict/Dockerfile
          push: false
          load: true
          tags: ${{ steps.meta.outputs.tags }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          build-args: |
            GO_VERSION=${{ env.GO_VERSION }}
            BUILD_DATE=${{ github.event.head_commit.timestamp }}
            VCS_REF=${{ github.sha }}

      - name: Run Trivy image scan
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ env.ECR_REGISTRY }}/${{ env.ECR_REPOSITORY }}:${{ github.sha }}
          format: 'sarif'
          output: 'trivy-image-results.sarif'
          severity: 'HIGH,CRITICAL'
          exit-code: '1'

      - name: Generate SBOM
        uses: anchore/sbom-action@v0
        with:
          image: ${{ env.ECR_REGISTRY }}/${{ env.ECR_REPOSITORY }}:${{ github.sha }}
          format: spdx-json
          output-file: sbom.spdx.json

      - name: Push to ECR
        uses: docker/build-push-action@v5
        with:
          context: core-dict
          file: core-dict/Dockerfile
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          cache-from: type=gha

      - name: Upload SBOM artifact
        uses: actions/upload-artifact@v4
        with:
          name: sbom
          path: sbom.spdx.json

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
      url: https://core-dict.dev.lbpay.io
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

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
        run: |
          aws eks update-kubeconfig --name eks-lbpay-dev --region us-east-1

      - name: Deploy to DEV
        working-directory: k8s/overlays/dev
        run: |
          kubectl set image deployment/core-dict \
            core-dict=${{ env.ECR_REGISTRY }}/${{ env.ECR_REPOSITORY }}:${{ github.sha }} \
            -n dict-dev
          kubectl rollout status deployment/core-dict -n dict-dev --timeout=5m

      - name: Run smoke tests
        run: |
          kubectl run smoke-test-${{ github.run_id }} \
            --image=curlimages/curl:latest \
            --restart=Never \
            --rm -i -n dict-dev \
            -- curl -f http://core-dict-svc:8080/health || exit 1

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
      url: https://core-dict.staging.lbpay.io
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1

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

      - name: Sync ArgoCD app (Staging)
        run: |
          argocd app set core-dict-staging \
            --kustomize-image ${{ env.ECR_REGISTRY }}/${{ env.ECR_REPOSITORY }}:${{ github.sha }}
          argocd app sync core-dict-staging --prune --force
          argocd app wait core-dict-staging --health --timeout 600

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
      url: https://core-dict.lbpay.io
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

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

      - name: Canary Deploy (10%)
        run: |
          argocd app set core-dict-prod \
            --kustomize-image ${{ env.ECR_REGISTRY }}/${{ env.ECR_REPOSITORY }}:${{ github.sha }}
          argocd app patch core-dict-prod \
            --patch '{"spec":{"source":{"helm":{"parameters":[{"name":"canary.weight","value":"10"}]}}}}'
          argocd app sync core-dict-prod

      - name: Wait 5 minutes (monitoring)
        run: sleep 300

      - name: Canary Deploy (50%)
        run: |
          argocd app patch core-dict-prod \
            --patch '{"spec":{"source":{"helm":{"parameters":[{"name":"canary.weight","value":"50"}]}}}}'
          argocd app sync core-dict-prod

      - name: Wait 5 minutes (monitoring)
        run: sleep 300

      - name: Full Deploy (100%)
        run: |
          argocd app patch core-dict-prod \
            --patch '{"spec":{"source":{"helm":{"parameters":[{"name":"canary.weight","value":"100"}]}}}}'
          argocd app sync core-dict-prod --prune --force
          argocd app wait core-dict-prod --health --timeout 600

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
              "text": "Core DICT Deployment",
              "blocks": [
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": "*Core DICT CI/CD Pipeline*\n*Status*: ${{ job.status }}\n*Branch*: ${{ github.ref_name }}\n*Commit*: ${{ github.sha }}\n*Author*: ${{ github.actor }}"
                  }
                },
                {
                  "type": "actions",
                  "elements": [
                    {
                      "type": "button",
                      "text": {
                        "type": "plain_text",
                        "text": "View Workflow"
                      },
                      "url": "${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}"
                    }
                  ]
                }
              ]
            }
```

---

## 4. Environment Configuration

### DEV Environment

```yaml
Environment: dev
Cluster: eks-lbpay-dev
Namespace: dict-dev
Domain: core-dict.dev.lbpay.io

Resources:
  requests:
    cpu: 250m
    memory: 256Mi
  limits:
    cpu: 500m
    memory: 512Mi

Replicas: 2
HPA: Disabled

Database: PostgreSQL (RDS - dev instance)
Cache: Redis (ElastiCache - dev cluster)
Broker: Pulsar (dev namespace)

Secrets:
  - DB_PASSWORD (AWS Secrets Manager)
  - REDIS_PASSWORD (AWS Secrets Manager)
  - PULSAR_TOKEN (AWS Secrets Manager)
```

### STAGING Environment

```yaml
Environment: staging
Cluster: eks-lbpay-staging
Namespace: dict-staging
Domain: core-dict.staging.lbpay.io

Resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 1000m
    memory: 1Gi

Replicas: 3
HPA: Enabled (min: 3, max: 6)

Database: PostgreSQL (RDS - staging instance)
Cache: Redis (ElastiCache - staging cluster)
Broker: Pulsar (staging namespace)
```

### PRODUCTION Environment

```yaml
Environment: production
Cluster: eks-lbpay-prod
Namespace: dict-prod
Domain: core-dict.lbpay.io

Resources:
  requests:
    cpu: 1000m
    memory: 1Gi
  limits:
    cpu: 2000m
    memory: 2Gi

Replicas: 5
HPA: Enabled (min: 5, max: 20)
PDB: maxUnavailable: 1

Database: PostgreSQL (RDS - prod multi-AZ)
Cache: Redis (ElastiCache - prod cluster with replica)
Broker: Pulsar (prod namespace with replication)

High Availability:
  - Multi-AZ deployment
  - PodDisruptionBudget
  - Anti-affinity rules
  - Cross-zone load balancing
```

---

## 5. Security Scanning

### gosec Configuration (`.gosec.yaml`)

```yaml
global:
  nosec: false
  show-ignored: true
  confidence: medium
  severity: medium

rules:
  # Enabled rules
  G101: true  # Look for hard coded credentials
  G102: true  # Bind to all interfaces
  G103: true  # Audit the use of unsafe block
  G104: true  # Audit errors not checked
  G106: true  # Audit the use of ssh.InsecureIgnoreHostKey
  G107: true  # Url provided to HTTP request as taint input
  G108: true  # Profiling endpoint automatically exposed on /debug/pprof
  G109: true  # Potential Integer overflow made by strconv.Atoi
  G110: true  # Potential DoS vulnerability via decompression bomb
  G201: true  # SQL query construction using format string
  G202: true  # SQL query construction using string concatenation
  G203: true  # Use of unescaped data in HTML templates
  G204: true  # Audit use of command execution
  G301: true  # Poor file permissions used when creating a directory
  G302: true  # Poor file permissions used when creation file or using chmod
  G303: true  # Creating tempfile using a predictable path
  G304: true  # File path provided as taint input
  G305: true  # File traversal when extracting zip archive
  G306: true  # Poor file permissions used when writing to a new file
  G307: true  # Deferring a method which returns an error
  G401: true  # Detect the usage of DES, RC4, MD5 or SHA1
  G402: true  # Look for bad TLS connection settings
  G403: true  # Ensure minimum RSA key length of 2048 bits
  G404: true  # Insecure random number source (rand)
  G501: true  # Import blacklist: crypto/md5
  G502: true  # Import blacklist: crypto/des
  G503: true  # Import blacklist: crypto/rc4
  G504: true  # Import blacklist: net/http/cgi

exclude-rules:
  - G104: # Allow unhandled errors in tests
      patterns:
        - "*_test.go"
```

### Trivy Scan Policy

```yaml
# .trivyignore

# Accepted vulnerabilities (with justification)

# CVE-XXXX-YYYY: [Justification for acceptance]
# Add ignored CVEs here with business justification
```

**Severity Levels**:
- **CRITICAL**: Block pipeline immediately
- **HIGH**: Block pipeline, require security review
- **MEDIUM**: Warning only, manual review
- **LOW**: Informational

---

## 6. Deployment Strategy

### Rolling Update (DEV)

```yaml
Strategy: RollingUpdate
maxSurge: 1
maxUnavailable: 0

Steps:
  1. Create new pod with new image
  2. Wait for pod to be Ready (health checks pass)
  3. Terminate old pod
  4. Repeat for each replica

Rollback Trigger:
  - Health check fails for 3 consecutive times
  - Pod crashes within 60 seconds
  - Automated rollback via kubectl rollout undo
```

### Blue-Green (Staging)

```yaml
Strategy: Blue-Green
Tools: ArgoCD + Kubernetes Services

Steps:
  1. Deploy GREEN environment (new version)
  2. Run smoke tests on GREEN
  3. Switch Service selector to GREEN
  4. Monitor for 10 minutes
  5. Decommission BLUE environment

Rollback:
  - Switch Service selector back to BLUE
  - Zero downtime rollback
```

### Canary (Production)

```yaml
Strategy: Canary
Phases: 10% → 50% → 100%
Tools: ArgoCD + Istio/Nginx Ingress

Phase 1 (10%):
  - Deploy canary with 10% traffic
  - Monitor for 5 minutes
  - Check metrics: error_rate, latency_p99, success_rate

Phase 2 (50%):
  - Increase to 50% traffic
  - Monitor for 5 minutes
  - Compare canary vs stable metrics

Phase 3 (100%):
  - Full rollout
  - Terminate stable version
  - Promote canary to stable

Rollback Triggers:
  - Error rate > 1%
  - P99 latency > 500ms
  - Success rate < 99%
  - Manual intervention
```

---

## 7. Rollback Strategy

### Automated Rollback

```yaml
Triggers:
  - Health check failures (3 consecutive)
  - High error rate (> 1%)
  - Memory/CPU limits exceeded
  - Pod crash loop backoff

Actions:
  DEV:
    - kubectl rollout undo deployment/core-dict -n dict-dev
    - Alert Slack channel #dev-deployments

  Staging:
    - ArgoCD rollback to previous sync
    - Alert Slack channel #staging-deployments

  PROD:
    - Manual approval required
    - PagerDuty incident creation
    - Alert Slack channel #prod-incidents
```

### Manual Rollback

```bash
# Rollback to previous revision
kubectl rollout undo deployment/core-dict -n dict-prod

# Rollback to specific revision
kubectl rollout undo deployment/core-dict --to-revision=3 -n dict-prod

# Check rollback status
kubectl rollout status deployment/core-dict -n dict-prod

# Verify pods
kubectl get pods -n dict-prod -l app=core-dict
```

### Rollback via ArgoCD

```bash
# Rollback to previous sync
argocd app rollback core-dict-prod

# Rollback to specific revision
argocd app rollback core-dict-prod <revision-id>

# Check sync status
argocd app get core-dict-prod
```

---

## 8. Monitoring & Alerts

### Deployment Metrics

```yaml
Metrics to Monitor:
  - Deployment success rate
  - Deployment duration
  - Rollback rate
  - Build time
  - Test coverage
  - Security vulnerabilities found

Tools:
  - Prometheus: Scrape deployment metrics
  - Grafana: Deployment dashboard
  - Slack: Real-time notifications
  - PagerDuty: Production incidents
```

### Health Check Endpoints

```yaml
Liveness Probe:
  path: /health/live
  port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

Readiness Probe:
  path: /health/ready
  port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3

Startup Probe:
  path: /health/startup
  port: 8080
  initialDelaySeconds: 0
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 30
```

### Alerts

```yaml
Critical Alerts:
  - Deployment failed (PROD)
  - Rollback triggered (PROD)
  - Security vulnerability CRITICAL found
  - Health checks failing > 5 minutes

Warning Alerts:
  - Deployment slow (> 15 minutes)
  - Test coverage below 80%
  - Security vulnerability HIGH found
  - Staging deployment failed
```

---

## 9. Rastreabilidade

### Documentos Relacionados

| ID | Documento | Relação |
|----|-----------|---------|
| **TEC-001** | [Core DICT Specification](../../11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md) | Especificação técnica do componente |
| **ADR-003** | [Protocol gRPC](../../02_Arquitetura/ADRs/ADR-003_Protocol_gRPC.md) | Justificativa do protocolo gRPC |
| **ADR-005** | [Database PostgreSQL](../../02_Arquitetura/ADRs/ADR-005_Database_PostgreSQL.md) | Justificativa do banco de dados |
| **SEC-001** | [mTLS Configuration](../../13_Seguranca/SEC-001_mTLS_Configuration.md) | Configuração de segurança |
| **DEV-004** | [Kubernetes Manifests](#) | Manifests K8s do Core DICT |

### Métricas de Sucesso

```yaml
Pipeline Performance:
  - Total duration: < 30 minutes (all stages)
  - Test coverage: >= 80%
  - Security scan: 0 HIGH/CRITICAL vulnerabilities
  - Deployment success rate: >= 99%

Quality Gates:
  ✅ All tests pass
  ✅ Code coverage >= 80%
  ✅ No security vulnerabilities (HIGH/CRITICAL)
  ✅ Docker image size < 50MB
  ✅ Build time < 12 minutes
  ✅ Deploy time < 5 minutes
```

---

## Anexos

### A. golangci-lint Configuration

```yaml
# .golangci.yml
run:
  timeout: 5m
  tests: true
  modules-download-mode: readonly

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - typecheck
    - unused
    - gofmt
    - goimports
    - misspell
    - gocritic
    - revive
    - gosec

linters-settings:
  errcheck:
    check-type-assertions: true
    check-blank: true
  govet:
    check-shadowing: true
  gofmt:
    simplify: true
  gosec:
    confidence: medium
    severity: medium
```

### B. Docker Build Arguments

```dockerfile
ARG GO_VERSION=1.24.5
ARG BUILD_DATE
ARG VCS_REF

LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.revision="${VCS_REF}"
LABEL org.opencontainers.image.source="https://github.com/lbpay/dict-core"
LABEL org.opencontainers.image.title="Core DICT"
LABEL org.opencontainers.image.vendor="LBPay"
```

### C. Secrets Management

```yaml
AWS Secrets Manager:
  - /lbpay/dict/dev/db-password
  - /lbpay/dict/dev/redis-password
  - /lbpay/dict/dev/pulsar-token
  - /lbpay/dict/staging/db-password
  - /lbpay/dict/prod/db-password

External Secrets Operator:
  apiVersion: external-secrets.io/v1beta1
  kind: ExternalSecret
  metadata:
    name: core-dict-secrets
    namespace: dict-prod
  spec:
    refreshInterval: 1h
    secretStoreRef:
      name: aws-secrets-manager
      kind: ClusterSecretStore
    target:
      name: core-dict-secrets
      creationPolicy: Owner
    data:
      - secretKey: db-password
        remoteRef:
          key: /lbpay/dict/prod/db-password
```

---

**Última Atualização**: 2025-10-25
**Versão**: 1.0
**Status**: ✅ Completo
