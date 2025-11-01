# SPECS-DEPLOYMENT.md - Deployment & Operations Specification

**Projeto**: DICT Rate Limit Monitoring System
**Stack**: Kubernetes + Helm + Goose + Docker
**Status**: ✅ ESPECIFICAÇÃO COMPLETA - Production-Ready

---

## 🎯 Objetivo

Especificação técnica completa de **deployment e operações** do sistema:

1. **Kubernetes Manifests**: Deployments, Services, ConfigMaps, Secrets
2. **Helm Charts**: Templates parametrizados para ambientes
3. **Database Migrations**: Goose para migrations SQL
4. **Docker Images**: Multi-stage builds otimizados
5. **CI/CD Pipeline**: GitHub Actions para deploy automatizado
6. **Disaster Recovery**: Backup, restore e rollback procedures

---

## 📋 Tabela de Conteúdos

- [1. Kubernetes Architecture](#1-kubernetes-architecture)
- [2. Helm Charts](#2-helm-charts)
- [3. Database Migrations](#3-database-migrations)
- [4. Docker Images](#4-docker-images)
- [5. CI/CD Pipeline](#5-cicd-pipeline)
- [6. Runbooks & Operations](#6-runbooks--operations)

---

## 1. Kubernetes Architecture

### Namespace & Resources

```yaml
# Location: k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: dict-rate-limit
  labels:
    app: dict-rate-limit
    environment: production
```

### Dict API Deployment

```yaml
# Location: k8s/dict-api-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dict-api
  namespace: dict-rate-limit
spec:
  replicas: 3
  selector:
    matchLabels:
      app: dict-api
  template:
    metadata:
      labels:
        app: dict-api
        version: v1.0.0
    spec:
      containers:
      - name: dict-api
        image: lbpay/dict-api:1.0.0
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: PORT
          value: "8080"
        - name: REDIS_ADDR
          valueFrom:
            configMapKeyRef:
              name: dict-config
              key: redis.addr
        - name: BRIDGE_GRPC_ADDRESS
          valueFrom:
            configMapKeyRef:
              name: dict-config
              key: bridge.grpc.address
        - name: BRIDGE_TLS_CERT_FILE
          value: /certs/client.crt
        - name: BRIDGE_TLS_KEY_FILE
          value: /certs/client.key
        volumeMounts:
        - name: tls-certs
          mountPath: /certs
          readOnly: true
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
      volumes:
      - name: tls-certs
        secret:
          secretName: bridge-tls-certs
---
apiVersion: v1
kind: Service
metadata:
  name: dict-api
  namespace: dict-rate-limit
spec:
  selector:
    app: dict-api
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
  type: ClusterIP
```

### Orchestration Worker Deployment

```yaml
# Location: k8s/orchestration-worker-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: orchestration-worker
  namespace: dict-rate-limit
spec:
  replicas: 2
  selector:
    matchLabels:
      app: orchestration-worker
  template:
    metadata:
      labels:
        app: orchestration-worker
    spec:
      containers:
      - name: worker
        image: lbpay/orchestration-worker:1.0.0
        env:
        - name: TEMPORAL_HOST
          valueFrom:
            configMapKeyRef:
              name: dict-config
              key: temporal.host
        - name: TEMPORAL_NAMESPACE
          value: "dict"
        - name: PULSAR_BROKER_URL
          valueFrom:
            configMapKeyRef:
              name: dict-config
              key: pulsar.broker.url
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: dict-secrets
              key: database.url
        resources:
          requests:
            memory: "256Mi"
            cpu: "200m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
```

### ConfigMap

```yaml
# Location: k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: dict-config
  namespace: dict-rate-limit
data:
  redis.addr: "redis-master.redis.svc.cluster.local:6379"
  bridge.grpc.address: "bridge.lbpay.svc.cluster.local:9090"
  temporal.host: "temporal-frontend.temporal.svc.cluster.local:7233"
  pulsar.broker.url: "pulsar://pulsar-broker.pulsar.svc.cluster.local:6650"
  cache.ttl: "60s"
  workflow.cron.schedule: "*/5 * * * *"
```

### Secrets

```yaml
# Location: k8s/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: dict-secrets
  namespace: dict-rate-limit
type: Opaque
stringData:
  database.url: "postgres://user:password@postgres.db.svc.cluster.local:5432/dict?sslmode=require"
---
apiVersion: v1
kind: Secret
metadata:
  name: bridge-tls-certs
  namespace: dict-rate-limit
type: kubernetes.io/tls
data:
  client.crt: <base64-encoded-cert>
  client.key: <base64-encoded-key>
  ca.crt: <base64-encoded-ca>
```

---

## 2. Helm Charts

### Chart Structure

```
helm/
├── Chart.yaml
├── values.yaml
├── values-staging.yaml
├── values-production.yaml
└── templates/
    ├── deployment-dict-api.yaml
    ├── deployment-orchestration-worker.yaml
    ├── service-dict-api.yaml
    ├── configmap.yaml
    ├── secrets.yaml
    ├── hpa.yaml
    ├── pdb.yaml
    └── servicemonitor.yaml
```

### Chart.yaml

```yaml
# Location: helm/Chart.yaml
apiVersion: v2
name: dict-rate-limit
description: DICT Rate Limit Monitoring System
type: application
version: 1.0.0
appVersion: "1.0.0"
keywords:
  - dict
  - rate-limit
  - bacen
maintainers:
  - name: LBPay Engineering
    email: engineering@lbpay.com
```

### values.yaml

```yaml
# Location: helm/values.yaml
# Default values for dict-rate-limit chart

replicaCount:
  dictApi: 3
  orchestrationWorker: 2

image:
  repository: lbpay
  pullPolicy: IfNotPresent
  dictApi:
    tag: "1.0.0"
  orchestrationWorker:
    tag: "1.0.0"

service:
  type: ClusterIP
  port: 80

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: dict-api.lbpay.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: dict-api-tls
      hosts:
        - dict-api.lbpay.com

resources:
  dictApi:
    requests:
      memory: "128Mi"
      cpu: "100m"
    limits:
      memory: "512Mi"
      cpu: "500m"
  orchestrationWorker:
    requests:
      memory: "256Mi"
      cpu: "200m"
    limits:
      memory: "1Gi"
      cpu: "1000m"

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

redis:
  addr: "redis-master.redis.svc.cluster.local:6379"
  cacheTTL: "60s"

bridge:
  grpcAddress: "bridge.lbpay.svc.cluster.local:9090"

temporal:
  host: "temporal-frontend.temporal.svc.cluster.local:7233"
  namespace: "dict"

pulsar:
  brokerURL: "pulsar://pulsar-broker.pulsar.svc.cluster.local:6650"
  topic: "persistent://lb-conn/dict/core-events"

database:
  url: "postgres://user:password@postgres.db.svc.cluster.local:5432/dict?sslmode=require"

monitoring:
  serviceMonitor:
    enabled: true
    interval: 30s
```

### Deployment Template

```yaml
# Location: helm/templates/deployment-dict-api.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "dict-rate-limit.fullname" . }}-dict-api
  labels:
    {{- include "dict-rate-limit.labels" . | nindent 4 }}
    app.kubernetes.io/component: dict-api
spec:
  replicas: {{ .Values.replicaCount.dictApi }}
  selector:
    matchLabels:
      {{- include "dict-rate-limit.selectorLabels" . | nindent 6 }}
      app.kubernetes.io/component: dict-api
  template:
    metadata:
      labels:
        {{- include "dict-rate-limit.selectorLabels" . | nindent 8 }}
        app.kubernetes.io/component: dict-api
    spec:
      containers:
      - name: dict-api
        image: "{{ .Values.image.repository }}/dict-api:{{ .Values.image.dictApi.tag }}"
        imagePullPolicy: {{ .Values.image.pullPolicy }}
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: REDIS_ADDR
          value: {{ .Values.redis.addr }}
        - name: BRIDGE_GRPC_ADDRESS
          value: {{ .Values.bridge.grpcAddress }}
        - name: CACHE_TTL
          value: {{ .Values.redis.cacheTTL }}
        resources:
          {{- toYaml .Values.resources.dictApi | nindent 10 }}
```

### HPA (Horizontal Pod Autoscaler)

```yaml
# Location: helm/templates/hpa.yaml
{{- if .Values.autoscaling.enabled }}
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{ include "dict-rate-limit.fullname" . }}-dict-api
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ include "dict-rate-limit.fullname" . }}-dict-api
  minReplicas: {{ .Values.autoscaling.minReplicas }}
  maxReplicas: {{ .Values.autoscaling.maxReplicas }}
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: {{ .Values.autoscaling.targetCPUUtilizationPercentage }}
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: {{ .Values.autoscaling.targetMemoryUtilizationPercentage }}
{{- end }}
```

---

## 3. Database Migrations

### Goose Configuration

```bash
# Location: db/migrations/goose.sh
#!/bin/bash

# Goose migration script
GOOSE_DRIVER="postgres"
GOOSE_DBSTRING="${DATABASE_URL}"

goose -dir ./db/migrations ${GOOSE_DRIVER} "${GOOSE_DBSTRING}" $@
```

### Migration Files

```sql
-- Location: db/migrations/001_create_dict_rate_limit_policies.sql
-- +goose Up
CREATE TABLE IF NOT EXISTS dict_rate_limit_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_name VARCHAR(100) NOT NULL UNIQUE,
    category VARCHAR(1) CHECK (category IN ('A','B','C','D','E','F','G','H')),
    capacity_max INTEGER NOT NULL CHECK (capacity_max > 0),
    refill_tokens INTEGER NOT NULL CHECK (refill_tokens > 0),
    refill_period_sec INTEGER NOT NULL CHECK (refill_period_sec > 0),
    warning_threshold_pct DECIMAL(5,2) DEFAULT 25.00,
    critical_threshold_pct DECIMAL(5,2) DEFAULT 10.00,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_policies_category ON dict_rate_limit_policies(category);
CREATE INDEX idx_policies_enabled ON dict_rate_limit_policies(enabled);

-- +goose Down
DROP TABLE IF EXISTS dict_rate_limit_policies CASCADE;
```

```sql
-- Location: db/migrations/002_create_dict_rate_limit_states.sql
-- +goose Up
CREATE TABLE IF NOT EXISTS dict_rate_limit_states (
    id BIGSERIAL,
    policy_name VARCHAR(100) NOT NULL,
    available_tokens INTEGER NOT NULL,
    capacity INTEGER NOT NULL,
    utilization_pct DECIMAL(5,2) NOT NULL,
    category VARCHAR(1),
    checked_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (id, checked_at)
) PARTITION BY RANGE (checked_at);

-- Create initial partitions (current month + next 2 months)
CREATE TABLE dict_rate_limit_states_2025_10 PARTITION OF dict_rate_limit_states
    FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');

CREATE TABLE dict_rate_limit_states_2025_11 PARTITION OF dict_rate_limit_states
    FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');

CREATE TABLE dict_rate_limit_states_2025_12 PARTITION OF dict_rate_limit_states
    FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');

-- Indexes
CREATE INDEX idx_states_policy_time ON dict_rate_limit_states(policy_name, checked_at DESC);
CREATE INDEX idx_states_utilization ON dict_rate_limit_states(utilization_pct DESC);

-- +goose Down
DROP TABLE IF EXISTS dict_rate_limit_states CASCADE;
```

### Kubernetes Migration Job

```yaml
# Location: k8s/migration-job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: dict-db-migration
  namespace: dict-rate-limit
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
      - name: migrate
        image: lbpay/dict-migrations:1.0.0
        command: ["/bin/sh", "-c"]
        args:
          - |
            goose -dir /migrations postgres "${DATABASE_URL}" up
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: dict-secrets
              key: database.url
```

---

## 4. Docker Images

### Multi-Stage Dockerfile (Dict API)

```dockerfile
# Location: apps/dict/Dockerfile
# Stage 1: Build
FROM golang:1.24.5-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/dict-api \
    ./apps/dict/cmd/main.go

# Stage 2: Runtime
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/dict-api /app/dict-api

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8080 9090

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/dict-api"]
```

### Multi-Stage Dockerfile (Orchestration Worker)

```dockerfile
# Location: apps/orchestration-worker/Dockerfile
FROM golang:1.24.5-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build \
    -ldflags="-w -s" \
    -o /app/orchestration-worker \
    ./apps/orchestration-worker/cmd/main.go

FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/orchestration-worker /app/orchestration-worker

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

ENTRYPOINT ["/app/orchestration-worker"]
```

---

## 5. CI/CD Pipeline

### GitHub Actions Workflow

```yaml
# Location: .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    tags:
      - 'v*'

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: lbpay/dict-rate-limit

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}

      - name: Build and push Dict API
        uses: docker/build-push-action@v5
        with:
          context: .
          file: apps/dict/Dockerfile
          push: true
          tags: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}/dict-api:${{ steps.meta.outputs.version }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

      - name: Build and push Orchestration Worker
        uses: docker/build-push-action@v5
        with:
          context: .
          file: apps/orchestration-worker/Dockerfile
          push: true
          tags: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}/orchestration-worker:${{ steps.meta.outputs.version }}

  deploy:
    needs: build-and-push
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup kubectl
        uses: azure/setup-kubectl@v3

      - name: Setup Helm
        uses: azure/setup-helm@v3
        with:
          version: 'v3.12.0'

      - name: Configure kubeconfig
        run: |
          echo "${{ secrets.KUBECONFIG }}" | base64 -d > kubeconfig
          export KUBECONFIG=./kubeconfig

      - name: Run database migrations
        run: |
          kubectl apply -f k8s/migration-job.yaml
          kubectl wait --for=condition=complete --timeout=300s job/dict-db-migration -n dict-rate-limit

      - name: Deploy with Helm
        run: |
          helm upgrade --install dict-rate-limit ./helm \
            --namespace dict-rate-limit \
            --create-namespace \
            --values helm/values-production.yaml \
            --set image.dictApi.tag=${{ github.ref_name }} \
            --set image.orchestrationWorker.tag=${{ github.ref_name }} \
            --wait \
            --timeout 10m

      - name: Verify deployment
        run: |
          kubectl rollout status deployment/dict-api -n dict-rate-limit
          kubectl rollout status deployment/orchestration-worker -n dict-rate-limit
```

---

## 6. Runbooks & Operations

### Runbook: Deploy New Version

```markdown
# Runbook: Deploy New Version

## Pre-Deployment Checklist
- [ ] All tests passing (CI)
- [ ] Code review approved
- [ ] Database migrations reviewed
- [ ] Rollback plan documented
- [ ] Stakeholders notified

## Deployment Steps

1. **Tag Release**
   ```bash
   git tag -a v1.0.1 -m "Release 1.0.1"
   git push origin v1.0.1
   ```

2. **Monitor CI/CD Pipeline**
   - Watch GitHub Actions workflow
   - Verify Docker images built
   - Confirm migration job completes

3. **Verify Deployment**
   ```bash
   kubectl get pods -n dict-rate-limit
   kubectl logs -f deployment/dict-api -n dict-rate-limit
   ```

4. **Smoke Tests**
   ```bash
   curl https://dict-api.lbpay.com/health
   curl https://dict-api.lbpay.com/api/v1/rate-limit/policies
   ```

5. **Monitor Metrics**
   - Check Grafana dashboards
   - Verify no alerts firing
   - Confirm error rate <1%

## Rollback Procedure

```bash
helm rollback dict-rate-limit -n dict-rate-limit
```
```

### Runbook: Database Maintenance

```markdown
# Runbook: Database Partition Management

## Monthly Partition Creation

```bash
# Connect to PostgreSQL
psql $DATABASE_URL

# Create next month partition
CREATE TABLE dict_rate_limit_states_2025_11
PARTITION OF dict_rate_limit_states
FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
```

## Partition Cleanup (Drop old partitions)

```bash
# Drop partitions older than 13 months
DROP TABLE IF EXISTS dict_rate_limit_states_2024_09;
```

## Vacuum & Analyze

```bash
VACUUM ANALYZE dict_rate_limit_states;
VACUUM ANALYZE dict_rate_limit_policies;
```
```

---

**Última Atualização**: 2025-10-31
**Versão**: 1.0.0
**Status**: ✅ ESPECIFICAÇÃO COMPLETA - Production-Ready
