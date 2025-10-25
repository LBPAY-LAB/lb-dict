# Pipelines CI/CD

**Propósito**: Configuração de pipelines de Integração Contínua e Deploy Contínuo

## 📋 Conteúdo

Esta pasta armazenará:

- **GitHub Actions Workflows**: Pipelines CI/CD em YAML
- **Jenkins Pipelines**: Jenkinsfiles (se aplicável)
- **Build Scripts**: Scripts de build customizados
- **Deployment Strategies**: Estratégias de deployment (blue-green, canary, rolling)

## 📁 Estrutura Esperada

```
Pipelines/
├── GitHub_Actions/
│   ├── ci-backend.yml
│   ├── ci-frontend.yml
│   ├── cd-staging.yml
│   ├── cd-production.yml
│   └── security-scan.yml
├── Docker/
│   ├── Dockerfile.connect
│   ├── Dockerfile.bridge
│   ├── Dockerfile.core
│   └── docker-compose.yml
├── Kubernetes/
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── ingress.yaml
│   └── configmap.yaml
└── Scripts/
    ├── build.sh
    ├── test.sh
    ├── deploy-staging.sh
    └── deploy-prod.sh
```

## 🎯 Exemplo: GitHub Actions CI Pipeline

```yaml
# .github/workflows/ci-backend.yml

name: CI - Backend (Connect + Bridge)

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

env:
  GO_VERSION: '1.22'

jobs:
  lint:
    name: Lint Code
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout=5m

  test:
    name: Unit Tests
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: dict_test
          POSTGRES_USER: dict
          POSTGRES_PASSWORD: dict_pass
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
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Run migrations
        run: |
          go install github.com/pressly/goose/v3/cmd/goose@latest
          goose -dir db/migrations postgres "postgres://dict:dict_pass@localhost:5432/dict_test?sslmode=disable" up

      - name: Run tests with coverage
        run: |
          go test ./... -v -race -coverprofile=coverage.out -covermode=atomic

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          fail_ci_if_error: true

  build:
    name: Build Docker Images
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build and push Connect
        uses: docker/build-push-action@v5
        with:
          context: ./apps/connect
          file: ./apps/connect/Dockerfile
          push: true
          tags: ghcr.io/${{ github.repository }}/connect:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

      - name: Build and push Bridge
        uses: docker/build-push-action@v5
        with:
          context: ./apps/bridge
          file: ./apps/bridge/Dockerfile
          push: true
          tags: ghcr.io/${{ github.repository }}/bridge:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  security:
    name: Security Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'

      - name: Upload Trivy results to GitHub Security
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'

      - name: Run gosec Security Scanner
        uses: securego/gosec@master
        with:
          args: './...'
```

## 🚀 Exemplo: CD Pipeline (Staging)

```yaml
# .github/workflows/cd-staging.yml

name: CD - Deploy to Staging

on:
  workflow_run:
    workflows: ["CI - Backend"]
    types:
      - completed
    branches: [develop]

jobs:
  deploy:
    name: Deploy to Staging
    runs-on: ubuntu-latest
    if: ${{ github.event.workflow_run.conclusion == 'success' }}
    environment:
      name: staging
      url: https://dict-staging.lbpay.com.br

    steps:
      - uses: actions/checkout@v4

      - name: Configure kubectl
        uses: azure/k8s-set-context@v3
        with:
          method: kubeconfig
          kubeconfig: ${{ secrets.KUBE_CONFIG_STAGING }}

      - name: Deploy Connect to Staging
        run: |
          kubectl set image deployment/dict-connect \
            connect=ghcr.io/${{ github.repository }}/connect:${{ github.sha }} \
            -n dict-staging

          kubectl rollout status deployment/dict-connect -n dict-staging

      - name: Deploy Bridge to Staging
        run: |
          kubectl set image deployment/dict-bridge \
            bridge=ghcr.io/${{ github.repository }}/bridge:${{ github.sha }} \
            -n dict-staging

          kubectl rollout status deployment/dict-bridge -n dict-staging

      - name: Run smoke tests
        run: |
          curl -f https://dict-staging.lbpay.com.br/health || exit 1
          curl -f https://dict-staging.lbpay.com.br/api/v1/health || exit 1

      - name: Notify Slack
        uses: slackapi/slack-github-action@v1
        with:
          payload: |
            {
              "text": "Deploy to Staging completed successfully",
              "blocks": [
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": ":rocket: *Deploy to Staging*\n*Commit*: ${{ github.sha }}\n*Status*: Success"
                  }
                }
              ]
            }
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```

## 🐳 Exemplo: Dockerfile Multi-stage

```dockerfile
# Dockerfile.connect

# Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/connect \
    ./cmd/connect

# Stage 2: Runtime
FROM alpine:3.19

# Install CA certificates (for mTLS)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/connect /app/connect

# Copy migrations (for runtime migration)
COPY --from=builder /app/db/migrations /app/db/migrations

# Create non-root user
RUN addgroup -g 1000 dict && \
    adduser -D -u 1000 -G dict dict

USER dict

EXPOSE 8080 9090

ENTRYPOINT ["/app/connect"]
```

## 📊 Deployment Strategies

### 1. Rolling Update (Default Kubernetes)
```yaml
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxUnavailable: 1
    maxSurge: 1
```

### 2. Blue-Green Deployment
```bash
# 1. Deploy green version
kubectl apply -f deployment-green.yaml

# 2. Wait for health check
kubectl wait --for=condition=ready pod -l version=green

# 3. Switch traffic to green
kubectl patch service dict-connect -p '{"spec":{"selector":{"version":"green"}}}'

# 4. Remove blue version
kubectl delete deployment dict-connect-blue
```

### 3. Canary Deployment (Argo Rollouts)
```yaml
strategy:
  canary:
    steps:
    - setWeight: 10   # 10% traffic to canary
    - pause: {duration: 5m}
    - setWeight: 50   # 50% traffic
    - pause: {duration: 5m}
    - setWeight: 100  # 100% traffic (full rollout)
```

## 🔐 Secrets Management

### GitHub Secrets (Required)

| Secret | Descrição |
|--------|-----------|
| `KUBE_CONFIG_STAGING` | Kubeconfig para cluster staging |
| `KUBE_CONFIG_PROD` | Kubeconfig para cluster produção |
| `DATABASE_URL_STAGING` | Connection string PostgreSQL staging |
| `DATABASE_URL_PROD` | Connection string PostgreSQL produção |
| `REDIS_URL_STAGING` | Connection string Redis staging |
| `REDIS_URL_PROD` | Connection string Redis produção |
| `SLACK_WEBHOOK_URL` | Webhook para notificações Slack |
| `CODECOV_TOKEN` | Token para upload de coverage |

### Kubernetes Secrets

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: dict-secrets
  namespace: dict-prod
type: Opaque
data:
  database-url: <base64-encoded>
  redis-url: <base64-encoded>
  bacen-mtls-cert: <base64-encoded>
  bacen-mtls-key: <base64-encoded>
```

## 📋 Pipeline Checklist

### CI Pipeline
- [ ] Lint code (golangci-lint)
- [ ] Run unit tests (> 80% coverage)
- [ ] Run integration tests (Testcontainers)
- [ ] Build Docker images
- [ ] Security scan (Trivy, gosec)
- [ ] Push images to registry

### CD Pipeline (Staging)
- [ ] Deploy to staging cluster
- [ ] Run smoke tests
- [ ] Notify team (Slack)

### CD Pipeline (Production)
- [ ] Manual approval required
- [ ] Canary deployment (10% → 50% → 100%)
- [ ] Health checks passed
- [ ] Rollback plan ready
- [ ] Notify stakeholders

## 📚 Referências

- [DevOps Best Practices](../DevOps_Best_Practices.md)
- [Kubernetes Manifests](../Kubernetes/)
- [Monitoramento](../Monitoramento/)
- [DAT-003: Migrations Strategy](../../03_Dados/DAT-003_Migrations_Strategy.md)

---

**Status**: 🔴 Pasta vazia (será preenchida na Fase 2)
**Fase de Preenchimento**: Fase 2 (setup de infraestrutura)
**Responsável**: DevOps Lead + SRE
**Ferramenta**: GitHub Actions, ArgoCD, Kubernetes
