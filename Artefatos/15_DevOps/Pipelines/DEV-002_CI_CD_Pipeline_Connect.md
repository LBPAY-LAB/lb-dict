# DEV-002: CI/CD Pipeline - RSFN Connect

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: RSFN Connect (Orchestration Service) CI/CD Pipeline
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: DEVOPS (AI Agent - DevOps Engineer)
**Revisor**: [Aguardando]
**Aprovador**: Head de DevOps, CTO (José Luís Silva)

---

## Controle de Versão

| Versão | Data | Autor | Descrição das Mudanças |
|--------|------|-------|------------------------|
| 1.0 | 2025-10-25 | DEVOPS | Versão inicial - Pipeline CI/CD para RSFN Connect com Temporal Worker |

---

## Sumário Executivo

### Visão Geral

Pipeline CI/CD para o **RSFN Connect** (multi-app architecture), cobrindo:
- ✅ **Dual Build**: API (`apps/dict`) + Temporal Worker (`apps/orchestration-worker`)
- ✅ **Workflow Testing**: Temporal workflow integration tests
- ✅ **Redis Integration Tests**: Cache layer validation
- ✅ **Security Scanning**: gosec, trivy, dependency check
- ✅ **Multi-Container Build**: 2 Docker images (API + Worker)
- ✅ **Temporal Worker Deployment**: StatefulSet + Worker registration
- ✅ **Multi-Environment Deploy**: dev, staging, prod

### Arquitetura do Connect

```
rsfn-connect/
├── apps/
│   ├── dict/                          # API REST (Fiber + Huma)
│   │   └── main.go
│   └── orchestration-worker/          # Temporal Worker
│       └── main.go
├── pkg/
│   ├── workflows/                     # Temporal Workflows (ClaimWorkflow, VSYNC, OTP)
│   ├── activities/                    # Temporal Activities
│   └── repositories/                  # Database access
└── configs/
```

### Stack Tecnológica

| Componente | Tecnologia | Versão |
|------------|------------|--------|
| **Linguagem** | Go | 1.24.5 |
| **API Framework** | Fiber v2 + Huma v2 | v2.52.9 / v2.34.1 |
| **Orchestration** | Temporal Workflow | v1.36.0 |
| **Cache** | Redis | v9.14.1 |
| **Message Broker** | Apache Pulsar | 3.0+ |
| **Database** | PostgreSQL | 16+ |

---

## Índice

1. [Workflow Overview](#1-workflow-overview)
2. [Pipeline Stages](#2-pipeline-stages)
3. [GitHub Actions Workflow](#3-github-actions-workflow)
4. [Temporal Worker Build](#4-temporal-worker-build)
5. [Environment Configuration](#5-environment-configuration)
6. [Deployment Strategy](#6-deployment-strategy)
7. [Monitoring & Alerts](#7-monitoring--alerts)
8. [Rastreabilidade](#8-rastreabilidade)

---

## 1. Workflow Overview

### Trigger Strategy

```yaml
Triggers:
  - push: [main, develop, release/*]
  - pull_request: [main, develop]
  - workflow_dispatch: Manual trigger

Path Filters:
  - rsfn-connect/**
  - .github/workflows/rsfn-connect-ci.yml
```

### Pipeline Flow

```
┌──────────────────────────────────────────────────────────────────┐
│                   RSFN Connect CI/CD Pipeline                     │
├──────────────────────────────────────────────────────────────────┤
│                                                                    │
│  Stage 1: Code Quality (Lint + Format)                           │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                 │
│  │   Lint     │→ │   Format   │→ │   Vet      │                 │
│  └────────────┘  └────────────┘  └────────────┘                 │
│         ↓                                                          │
│  Stage 2: Testing (Unit + Integration + Workflows)               │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                 │
│  │ Unit Tests │→ │Integration │→ │ Temporal   │                 │
│  │            │  │Redis+Pulsar│  │ Workflows  │                 │
│  └────────────┘  └────────────┘  └────────────┘                 │
│         ↓                                                          │
│  Stage 3: Security Scanning                                       │
│  ┌────────────┐  ┌────────────┐                                  │
│  │   gosec    │→ │   trivy    │                                  │
│  └────────────┘  └────────────┘                                  │
│         ↓                                                          │
│  Stage 4: Multi-Container Build                                   │
│  ┌────────────┐  ┌────────────┐                                  │
│  │ Build API  │  │Build Worker│                                  │
│  │  (dict)    │  │(orchestr.) │                                  │
│  └────────────┘  └────────────┘                                  │
│         ↓              ↓                                           │
│  ┌────────────┐  ┌────────────┐                                  │
│  │ Push ECR   │  │ Push ECR   │                                  │
│  │ connect-api│  │connect-wrkr│                                  │
│  └────────────┘  └────────────┘                                  │
│         ↓                                                          │
│  Stage 5: Deploy (API + Worker + Temporal Registration)          │
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

Same as Core DICT (golangci-lint, gofmt, go vet).

---

### Stage 2: Testing (15 min)

**Objetivos**:
- Executar unit tests
- Executar integration tests (Redis, Pulsar, PostgreSQL)
- Executar **Temporal workflow tests**

**Jobs**:

```yaml
unit-tests:
  - go test -v -race ./...
  - Coverage threshold: 80%

integration-tests:
  Services:
    - PostgreSQL 16
    - Redis 7
    - Pulsar 3.0
    - Temporal Server (dev mode)

  Tests:
    - API REST endpoints
    - Pulsar producer/consumer
    - Redis cache operations
    - Database repositories

temporal-workflow-tests:
  Setup:
    - Start Temporal dev server
    - Register worker
    - Start test workflows

  Tests:
    - ClaimWorkflow (30-day monitoring)
    - VSYNC Workflow (planned - mock)
    - OTP Workflow (planned - mock)
    - Activity retries
    - Workflow cancellation
    - Signal handling
```

**Temporal Test Example**:

```go
// Test ClaimWorkflow
func TestClaimWorkflow(t *testing.T) {
    testSuite := &testsuite.WorkflowTestSuite{}
    env := testSuite.NewTestWorkflowEnvironment()

    // Register workflow and activities
    env.RegisterWorkflow(ClaimWorkflow)
    env.RegisterActivity(SendClaimToBackofficeActivity)
    env.RegisterActivity(MonitorClaimStatusActivity)

    // Execute workflow
    env.ExecuteWorkflow(ClaimWorkflow, ClaimInput{
        ClaimID: "CLAIM-123",
        KeyType: "CPF",
        KeyValue: "12345678900",
    })

    require.True(t, env.IsWorkflowCompleted())
    require.NoError(t, env.GetWorkflowError())
}
```

---

### Stage 3: Security Scanning (8 min)

Same as Core DICT (gosec + trivy).

---

### Stage 4: Multi-Container Build (20 min)

**Objetivos**:
- Build 2 Docker images: API + Worker
- Scan both images
- Push both to ECR

**API Dockerfile** (`apps/dict/Dockerfile`):

```dockerfile
# Stage 1: Builder
FROM golang:1.24.5-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o connect-api ./apps/dict

# Stage 2: Runtime
FROM gcr.io/distroless/static-debian11
COPY --from=builder /build/connect-api /
COPY --from=builder /build/configs /configs
EXPOSE 8080 9090
USER nonroot:nonroot
ENTRYPOINT ["/connect-api"]
```

**Worker Dockerfile** (`apps/orchestration-worker/Dockerfile`):

```dockerfile
# Stage 1: Builder
FROM golang:1.24.5-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o orchestration-worker ./apps/orchestration-worker

# Stage 2: Runtime
FROM gcr.io/distroless/static-debian11
COPY --from=builder /build/orchestration-worker /
COPY --from=builder /build/configs /configs
USER nonroot:nonroot
ENTRYPOINT ["/orchestration-worker"]
```

**Build Strategy**:

```yaml
Build Matrix:
  - App: dict
    Image: connect-api
    Dockerfile: apps/dict/Dockerfile
    Context: .

  - App: orchestration-worker
    Image: connect-worker
    Dockerfile: apps/orchestration-worker/Dockerfile
    Context: .
```

---

### Stage 5: Deploy (10 min)

**Components to Deploy**:
1. **API Deployment** (Kubernetes Deployment)
2. **Temporal Worker StatefulSet** (StatefulSet for worker registration)
3. **Services** (API Service, metrics)
4. **ConfigMaps/Secrets**

**Deployment Order**:
1. Deploy API
2. Deploy Temporal Worker
3. Register Worker with Temporal Server
4. Verify worker is polling tasks

---

## 3. GitHub Actions Workflow

### Arquivo: `.github/workflows/rsfn-connect-ci.yml`

```yaml
name: RSFN Connect CI/CD

on:
  push:
    branches: [main, develop, release/**]
    paths:
      - 'rsfn-connect/**'
      - '.github/workflows/rsfn-connect-ci.yml'
  pull_request:
    branches: [main, develop]
    paths:
      - 'rsfn-connect/**'
  workflow_dispatch:

env:
  GO_VERSION: '1.24.5'
  DOCKER_BUILDKIT: 1
  ECR_REGISTRY: 123456789012.dkr.ecr.us-east-1.amazonaws.com
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
          working-directory: rsfn-connect
          args: --timeout=5m

      - name: Check formatting
        working-directory: rsfn-connect
        run: |
          gofmt -s -l . | tee fmt_errors.txt
          if [ -s fmt_errors.txt ]; then
            echo "::error::Code not formatted"
            exit 1
          fi

      - name: Run go vet
        working-directory: rsfn-connect
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
          POSTGRES_DB: connect_test
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

      pulsar:
        image: apachepulsar/pulsar:3.0.0
        command: bin/pulsar standalone
        ports:
          - 6650:6650
          - 8080:8080

      temporal:
        image: temporalio/auto-setup:1.22.0
        env:
          DB: postgresql
          DB_PORT: 5432
          POSTGRES_USER: postgres
          POSTGRES_PWD: testpass
          POSTGRES_SEEDS: postgres
        ports:
          - 7233:7233

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Wait for services
        run: |
          timeout 60 bash -c 'until curl -f http://localhost:8080/admin/v2/clusters; do sleep 2; done'
          timeout 60 bash -c 'until nc -z localhost 7233; do sleep 2; done'

      - name: Run unit tests
        working-directory: rsfn-connect
        env:
          DATABASE_URL: postgres://postgres:testpass@localhost:5432/connect_test?sslmode=disable
          REDIS_URL: redis://localhost:6379
          PULSAR_URL: pulsar://localhost:6650
          TEMPORAL_HOST: localhost:7233
        run: |
          go test -v -race -coverprofile=coverage.out ./...

      - name: Check coverage
        working-directory: rsfn-connect
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: $coverage%"
          if (( $(echo "$coverage < 80" | bc -l) )); then
            exit 1
          fi

      - name: Run Temporal workflow tests
        working-directory: rsfn-connect
        env:
          TEMPORAL_HOST: localhost:7233
        run: |
          go test -v -tags=workflow ./pkg/workflows/...

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
          args: '-no-fail -fmt sarif -out gosec.sarif ./rsfn-connect/...'

      - name: Upload gosec SARIF
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: gosec.sarif

      - name: Run Trivy FS scan
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: 'rsfn-connect'
          format: 'sarif'
          output: 'trivy-fs.sarif'
          severity: 'HIGH,CRITICAL'
          exit-code: '1'

  # ====================================
  # Stage 4: Multi-Container Build
  # ====================================

  build:
    name: Build & Push Docker Images
    runs-on: ubuntu-latest
    needs: security
    strategy:
      matrix:
        app:
          - name: dict
            image: connect-api
            dockerfile: apps/dict/Dockerfile
          - name: orchestration-worker
            image: connect-worker
            dockerfile: apps/orchestration-worker/Dockerfile
    steps:
      - uses: actions/checkout@v4

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
          images: ${{ env.ECR_REGISTRY }}/lbpay/${{ matrix.app.image }}
          tags: |
            type=ref,event=branch
            type=sha,prefix={{branch}}-
            type=raw,value=latest,enable={{is_default_branch}}

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Build image
        uses: docker/build-push-action@v5
        with:
          context: rsfn-connect
          file: rsfn-connect/${{ matrix.app.dockerfile }}
          push: false
          load: true
          tags: ${{ steps.meta.outputs.tags }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

      - name: Scan image with Trivy
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ env.ECR_REGISTRY }}/lbpay/${{ matrix.app.image }}:${{ github.sha }}
          format: 'sarif'
          output: 'trivy-${{ matrix.app.name }}.sarif'
          severity: 'HIGH,CRITICAL'
          exit-code: '1'

      - name: Push to ECR
        uses: docker/build-push-action@v5
        with:
          context: rsfn-connect
          file: rsfn-connect/${{ matrix.app.dockerfile }}
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
      url: https://connect-api.dev.lbpay.io
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

      - name: Deploy API
        run: |
          kubectl set image deployment/connect-api \
            connect-api=${{ env.ECR_REGISTRY }}/lbpay/connect-api:${{ github.sha }} \
            -n dict-dev
          kubectl rollout status deployment/connect-api -n dict-dev --timeout=5m

      - name: Deploy Temporal Worker
        run: |
          kubectl set image statefulset/connect-worker \
            connect-worker=${{ env.ECR_REGISTRY }}/lbpay/connect-worker:${{ github.sha }} \
            -n dict-dev
          kubectl rollout status statefulset/connect-worker -n dict-dev --timeout=5m

      - name: Verify Temporal Worker registration
        run: |
          kubectl exec -n dict-dev connect-worker-0 -- \
            curl -f http://localhost:9090/metrics | grep temporal_worker_registered || exit 1

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
      url: https://connect-api.staging.lbpay.io
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

      - name: Sync Connect API (Staging)
        run: |
          argocd app set connect-api-staging \
            --kustomize-image connect-api=${{ env.ECR_REGISTRY }}/lbpay/connect-api:${{ github.sha }}
          argocd app sync connect-api-staging --prune --force
          argocd app wait connect-api-staging --health --timeout 600

      - name: Sync Connect Worker (Staging)
        run: |
          argocd app set connect-worker-staging \
            --kustomize-image connect-worker=${{ env.ECR_REGISTRY }}/lbpay/connect-worker:${{ github.sha }}
          argocd app sync connect-worker-staging --prune --force
          argocd app wait connect-worker-staging --health --timeout 600

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
      url: https://connect-api.lbpay.io
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

      - name: Deploy Connect API (PROD)
        run: |
          argocd app set connect-api-prod \
            --kustomize-image connect-api=${{ env.ECR_REGISTRY }}/lbpay/connect-api:${{ github.sha }}
          argocd app sync connect-api-prod --prune --force
          argocd app wait connect-api-prod --health --timeout 600

      - name: Deploy Connect Worker (PROD)
        run: |
          argocd app set connect-worker-prod \
            --kustomize-image connect-worker=${{ env.ECR_REGISTRY }}/lbpay/connect-worker:${{ github.sha }}
          argocd app sync connect-worker-prod --prune --force
          argocd app wait connect-worker-prod --health --timeout 600

      - name: Verify Temporal Worker health
        run: |
          kubectl exec -n dict-prod connect-worker-0 -- \
            curl -f http://localhost:9090/health || exit 1

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
              "text": "RSFN Connect Deployment",
              "blocks": [
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": "*RSFN Connect CI/CD*\n*Status*: ${{ job.status }}\n*Branch*: ${{ github.ref_name }}\n*Images*: connect-api, connect-worker"
                  }
                }
              ]
            }
```

---

## 4. Temporal Worker Build

### Worker Configuration

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: connect-worker
  namespace: dict-prod
spec:
  serviceName: connect-worker
  replicas: 3
  selector:
    matchLabels:
      app: connect-worker
  template:
    metadata:
      labels:
        app: connect-worker
    spec:
      containers:
      - name: connect-worker
        image: 123456789012.dkr.ecr.us-east-1.amazonaws.com/lbpay/connect-worker:latest
        env:
        - name: TEMPORAL_HOST
          value: "temporal-frontend.temporal:7233"
        - name: TEMPORAL_NAMESPACE
          value: "lbpay-dict"
        - name: WORKER_TASK_QUEUE
          value: "dict-workflows"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: connect-secrets
              key: database-url
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: connect-secrets
              key: redis-url
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 1000m
            memory: 1Gi
        livenessProbe:
          httpGet:
            path: /health
            port: 9090
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 9090
          initialDelaySeconds: 10
          periodSeconds: 5
```

### Worker Registration

```go
// Worker startup code
func main() {
    c, err := client.Dial(client.Options{
        HostPort:  os.Getenv("TEMPORAL_HOST"),
        Namespace: os.Getenv("TEMPORAL_NAMESPACE"),
    })
    if err != nil {
        log.Fatal("Failed to create Temporal client", err)
    }
    defer c.Close()

    w := worker.New(c, "dict-workflows", worker.Options{})

    // Register workflows
    w.RegisterWorkflow(workflows.ClaimWorkflow)
    w.RegisterWorkflow(workflows.VSYNCWorkflow)
    w.RegisterWorkflow(workflows.OTPWorkflow)

    // Register activities
    w.RegisterActivity(activities.SendClaimToBackofficeActivity)
    w.RegisterActivity(activities.MonitorClaimStatusActivity)
    w.RegisterActivity(activities.SyncAccountsActivity)

    log.Println("Temporal worker started, polling task queue: dict-workflows")
    err = w.Run(worker.InterruptCh())
    if err != nil {
        log.Fatal("Worker failed", err)
    }
}
```

---

## 5. Environment Configuration

### DEV Environment

```yaml
API:
  Replicas: 2
  Resources:
    requests: {cpu: 250m, memory: 256Mi}
    limits: {cpu: 500m, memory: 512Mi}

Worker:
  Replicas: 1
  Resources:
    requests: {cpu: 250m, memory: 256Mi}
    limits: {cpu: 500m, memory: 512Mi}

Temporal:
  Namespace: lbpay-dict-dev
  TaskQueue: dict-workflows-dev
```

### STAGING Environment

```yaml
API:
  Replicas: 3
  Resources:
    requests: {cpu: 500m, memory: 512Mi}
    limits: {cpu: 1000m, memory: 1Gi}

Worker:
  Replicas: 2
  Resources:
    requests: {cpu: 500m, memory: 512Mi}
    limits: {cpu: 1000m, memory: 1Gi}
```

### PRODUCTION Environment

```yaml
API:
  Replicas: 5
  HPA: {min: 5, max: 20}
  Resources:
    requests: {cpu: 1000m, memory: 1Gi}
    limits: {cpu: 2000m, memory: 2Gi}

Worker:
  Replicas: 3
  HPA: {min: 3, max: 10}
  Resources:
    requests: {cpu: 1000m, memory: 1Gi}
    limits: {cpu: 2000m, memory: 2Gi}

Temporal:
  Namespace: lbpay-dict-prod
  TaskQueue: dict-workflows-prod
  Workers: 3 (stateful, sticky sessions)
```

---

## 6. Deployment Strategy

### API Deployment

Same as Core DICT:
- DEV: Rolling Update
- Staging: Blue-Green
- PROD: Canary (10% → 50% → 100%)

### Worker Deployment (StatefulSet)

```yaml
Strategy: RollingUpdate

Rolling Update:
  1. Update worker-0
  2. Wait for worker-0 to register with Temporal
  3. Update worker-1
  4. Repeat for all workers

Challenges:
  - Workers must gracefully drain in-flight workflows
  - New workers must register before old workers terminate
  - Zero downtime for workflow execution

Graceful Shutdown:
  1. Stop polling new tasks
  2. Wait for in-flight activities to complete (timeout: 5 minutes)
  3. Terminate worker
```

---

## 7. Monitoring & Alerts

### Temporal Metrics

```yaml
Metrics:
  - temporal_workflow_start_total
  - temporal_workflow_completed_total
  - temporal_workflow_failed_total
  - temporal_activity_execution_latency
  - temporal_worker_task_slots_available
  - temporal_worker_registered

Dashboards:
  - Temporal Workflows Dashboard (Grafana)
  - Worker Health Dashboard
  - Activity Execution Times

Alerts:
  - Workflow failure rate > 5%
  - Activity timeout rate > 1%
  - Worker not registered for > 2 minutes
  - Task queue backlog > 1000
```

---

## 8. Rastreabilidade

### Documentos Relacionados

| ID | Documento | Relação |
|----|-----------|---------|
| **TEC-003** | [RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md) | Especificação técnica do Connect |
| **ADR-002** | [Temporal Workflow](../../02_Arquitetura/ADRs/ADR-002_Orchestrator_Temporal_Workflow.md) | Justificativa do Temporal |
| **ADR-004** | [Redis Cache](../../02_Arquitetura/ADRs/ADR-004_Cache_Redis.md) | Justificativa do Redis |
| **DEV-004** | [Kubernetes Manifests](#) | Manifests K8s do Connect |

---

**Última Atualização**: 2025-10-25
**Versão**: 1.0
**Status**: ✅ Completo
