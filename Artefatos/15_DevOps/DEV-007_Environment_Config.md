# DEV-007: Environment Configuration - DICT

**Projeto**: DICT - Diretorio de Identificadores de Contas Transacionais (LBPay)
**Componente**: Environment Configuration & Secrets Management
**Versao**: 1.0
**Data**: 2025-10-25
**Autor**: DEVOPS (AI Agent - DevOps Engineer)
**Revisor**: [Aguardando]
**Aprovador**: Head de DevOps, CTO (Jose Luis Silva)

---

## Controle de Versao

| Versao | Data | Autor | Descricao das Mudancas |
|--------|------|-------|------------------------|
| 1.0 | 2025-10-25 | DEVOPS | Versao inicial - Configuracao de ambientes e gestao de secrets |

---

## Sumario Executivo

### Visao Geral

Especificacao completa de **configuracao de ambientes** para o **DICT**, incluindo:
- Configuracao para 3 ambientes: dev, staging, prod
- ConfigMaps e Secrets do Kubernetes
- Integracao com AWS Secrets Manager via External Secrets Operator
- Vault integration para secrets sensíveis
- Feature flags (LaunchDarkly ou similar)
- Estrategia de promocao entre ambientes

### Objetivos

- Separacao clara de ambientes (dev, staging, prod)
- Secrets nunca em codigo ou manifests
- Configuracao versionada no Git (exceto secrets)
- Rotacao automatica de secrets
- Auditoria completa de acesso a secrets

### Stack de Configuracao

| Componente | Ferramenta | Proposito |
|------------|------------|-----------|
| **ConfigMaps** | Kubernetes | Configuracoes nao-sensiveis |
| **Secrets** | Kubernetes + External Secrets Operator | Secrets dinamicos do AWS Secrets Manager |
| **Vault** | HashiCorp Vault | Secrets sensíveis (certificados ICP-Brasil) |
| **Feature Flags** | LaunchDarkly / Unleash | Controle de features |
| **Environment Variables** | Kubernetes env / envFrom | Injecao de config em pods |

---

## Indice

1. [Environment Overview](#1-environment-overview)
2. [ConfigMaps](#2-configmaps)
3. [Secrets Management](#3-secrets-management)
4. [AWS Secrets Manager Integration](#4-aws-secrets-manager-integration)
5. [HashiCorp Vault Integration](#5-hashicorp-vault-integration)
6. [Feature Flags](#6-feature-flags)
7. [Promotion Strategy](#7-promotion-strategy)
8. [Rastreabilidade](#8-rastreabilidade)

---

## 1. Environment Overview

### Environments

```yaml
Environments:
  1. Development (dev):
     - Purpose: Desenvolvimento ativo, testes de desenvolvedores
     - Cluster: eks-lbpay-dev
     - Namespace: dict-dev
     - Stability: Baixa (pode quebrar)
     - Data: Sintetico (nao real)
     - Replicas: 2 (sem HPA)

  2. Staging (staging):
     - Purpose: Pre-producao, testes de QA, homologacao
     - Cluster: eks-lbpay-staging
     - Namespace: dict-staging
     - Stability: Media
     - Data: Copia anonimizada de producao
     - Replicas: 3 (com HPA)

  3. Production (prod):
     - Purpose: Ambiente de producao real
     - Cluster: eks-lbpay-prod
     - Namespace: dict-prod
     - Stability: Alta (SLO 99.9%)
     - Data: Real (dados de clientes)
     - Replicas: 5 (com HPA, PDB)
```

### Resource Allocation

```yaml
# DEV Environment
Resources (dev):
  core-dict:
    requests: {cpu: 250m, memory: 256Mi}
    limits: {cpu: 500m, memory: 512Mi}
  connect-api:
    requests: {cpu: 250m, memory: 256Mi}
    limits: {cpu: 500m, memory: 512Mi}
  connect-worker:
    requests: {cpu: 250m, memory: 256Mi}
    limits: {cpu: 500m, memory: 512Mi}
  rsfn-bridge:
    requests: {cpu: 250m, memory: 512Mi}
    limits: {cpu: 500m, memory: 1Gi}

# STAGING Environment
Resources (staging):
  core-dict:
    requests: {cpu: 500m, memory: 512Mi}
    limits: {cpu: 1000m, memory: 1Gi}
  connect-api:
    requests: {cpu: 500m, memory: 512Mi}
    limits: {cpu: 1000m, memory: 1Gi}
  connect-worker:
    requests: {cpu: 500m, memory: 512Mi}
    limits: {cpu: 1000m, memory: 1Gi}
  rsfn-bridge:
    requests: {cpu: 500m, memory: 1Gi}
    limits: {cpu: 1000m, memory: 2Gi}

# PRODUCTION Environment
Resources (prod):
  core-dict:
    requests: {cpu: 1000m, memory: 1Gi}
    limits: {cpu: 2000m, memory: 2Gi}
  connect-api:
    requests: {cpu: 1000m, memory: 1Gi}
    limits: {cpu: 2000m, memory: 2Gi}
  connect-worker:
    requests: {cpu: 1000m, memory: 1Gi}
    limits: {cpu: 2000m, memory: 2Gi}
  rsfn-bridge:
    requests: {cpu: 1000m, memory: 2Gi}
    limits: {cpu: 2000m, memory: 4Gi}
```

---

## 2. ConfigMaps

### Core DICT ConfigMap

```yaml
# configmap-core-dict-dev.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: core-dict-config
  namespace: dict-dev
  labels:
    app: core-dict
    environment: dev
data:
  # Application
  ENV: "dev"
  LOG_LEVEL: "debug"
  LOG_FORMAT: "json"

  # Server
  GRPC_PORT: "50051"
  HTTP_PORT: "8080"
  METRICS_PORT: "9090"
  GRPC_MAX_CONN: "100"

  # Database
  DB_MAX_OPEN_CONNS: "25"
  DB_MAX_IDLE_CONNS: "10"
  DB_CONN_MAX_LIFETIME: "5m"

  # Cache (Redis)
  CACHE_TTL: "300"
  CACHE_KEY_PREFIX: "dict:dev:"

  # Pulsar
  PULSAR_TOPIC_PREFIX: "persistent://lbpay/dict-dev/"
  PULSAR_SUBSCRIPTION: "dict-dev-sub"
  PULSAR_CONSUMER_NAME: "core-dict-dev"

  # Features
  FEATURE_FLAG_URL: "https://featureflags.lbpay.io"
  FEATURE_FLAG_ENV: "dev"

  # Observability
  OTEL_EXPORTER_OTLP_ENDPOINT: "http://jaeger-collector.monitoring.svc.cluster.local:14268/api/traces"
  OTEL_SERVICE_NAME: "core-dict"
  OTEL_SERVICE_VERSION: "1.0.0"
  OTEL_DEPLOYMENT_ENVIRONMENT: "dev"

  # Rate Limiting
  RATE_LIMIT_ENABLED: "true"
  RATE_LIMIT_RPS: "1000"
  RATE_LIMIT_BURST: "2000"

---
# configmap-core-dict-staging.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: core-dict-config
  namespace: dict-staging
  labels:
    app: core-dict
    environment: staging
data:
  ENV: "staging"
  LOG_LEVEL: "info"
  LOG_FORMAT: "json"
  GRPC_PORT: "50051"
  HTTP_PORT: "8080"
  METRICS_PORT: "9090"
  GRPC_MAX_CONN: "500"
  DB_MAX_OPEN_CONNS: "50"
  DB_MAX_IDLE_CONNS: "25"
  DB_CONN_MAX_LIFETIME: "10m"
  CACHE_TTL: "600"
  CACHE_KEY_PREFIX: "dict:staging:"
  PULSAR_TOPIC_PREFIX: "persistent://lbpay/dict-staging/"
  PULSAR_SUBSCRIPTION: "dict-staging-sub"
  PULSAR_CONSUMER_NAME: "core-dict-staging"
  FEATURE_FLAG_URL: "https://featureflags.lbpay.io"
  FEATURE_FLAG_ENV: "staging"
  OTEL_EXPORTER_OTLP_ENDPOINT: "http://jaeger-collector.monitoring.svc.cluster.local:14268/api/traces"
  OTEL_SERVICE_NAME: "core-dict"
  OTEL_SERVICE_VERSION: "1.0.0"
  OTEL_DEPLOYMENT_ENVIRONMENT: "staging"
  RATE_LIMIT_ENABLED: "true"
  RATE_LIMIT_RPS: "5000"
  RATE_LIMIT_BURST: "10000"

---
# configmap-core-dict-prod.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: core-dict-config
  namespace: dict-prod
  labels:
    app: core-dict
    environment: production
data:
  ENV: "production"
  LOG_LEVEL: "info"
  LOG_FORMAT: "json"
  GRPC_PORT: "50051"
  HTTP_PORT: "8080"
  METRICS_PORT: "9090"
  GRPC_MAX_CONN: "1000"
  DB_MAX_OPEN_CONNS: "100"
  DB_MAX_IDLE_CONNS: "50"
  DB_CONN_MAX_LIFETIME: "15m"
  CACHE_TTL: "600"
  CACHE_KEY_PREFIX: "dict:prod:"
  PULSAR_TOPIC_PREFIX: "persistent://lbpay/dict-prod/"
  PULSAR_SUBSCRIPTION: "dict-prod-sub"
  PULSAR_CONSUMER_NAME: "core-dict-prod"
  FEATURE_FLAG_URL: "https://featureflags.lbpay.io"
  FEATURE_FLAG_ENV: "production"
  OTEL_EXPORTER_OTLP_ENDPOINT: "http://jaeger-collector.monitoring.svc.cluster.local:14268/api/traces"
  OTEL_SERVICE_NAME: "core-dict"
  OTEL_SERVICE_VERSION: "1.0.0"
  OTEL_DEPLOYMENT_ENVIRONMENT: "production"
  RATE_LIMIT_ENABLED: "true"
  RATE_LIMIT_RPS: "10000"
  RATE_LIMIT_BURST: "20000"
```

### RSFN Bridge ConfigMap

```yaml
# configmap-rsfn-bridge-dev.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: rsfn-bridge-config
  namespace: dict-dev
data:
  ENV: "dev"
  LOG_LEVEL: "debug"

  # Bacen API (Homologacao)
  BACEN_API_URL: "https://api-hom.rsfn.bcb.gov.br/dict/api/v1"
  BACEN_TIMEOUT: "30s"
  BACEN_RETRY_MAX: "3"
  BACEN_RETRY_DELAY: "1s"

  # Circuit Breaker
  CIRCUIT_BREAKER_THRESHOLD: "5"
  CIRCUIT_BREAKER_TIMEOUT: "60s"
  CIRCUIT_BREAKER_MAX_REQUESTS: "10"

  # XML Signing
  XML_SIGNER_JAR_PATH: "/opt/signer/xml-signer.jar"
  XML_SIGNER_TIMEOUT: "5s"

  # mTLS
  MTLS_ENABLED: "true"
  MTLS_CERT_PATH: "/certs/client-cert.pem"
  MTLS_KEY_PATH: "/certs/client-key.pem"
  MTLS_CA_PATH: "/certs/ca.pem"

  # SOAP
  SOAP_NAMESPACE: "http://www.bcb.gov.br/pi/dict/v1"
  SOAP_ACTION_PREFIX: "http://www.bcb.gov.br/pi/dict/v1/"

---
# configmap-rsfn-bridge-prod.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: rsfn-bridge-config
  namespace: dict-prod
data:
  ENV: "production"
  LOG_LEVEL: "info"

  # Bacen API (Producao)
  BACEN_API_URL: "https://api.rsfn.bcb.gov.br/dict/api/v1"
  BACEN_TIMEOUT: "30s"
  BACEN_RETRY_MAX: "3"
  BACEN_RETRY_DELAY: "2s"

  CIRCUIT_BREAKER_THRESHOLD: "5"
  CIRCUIT_BREAKER_TIMEOUT: "60s"
  CIRCUIT_BREAKER_MAX_REQUESTS: "10"

  XML_SIGNER_JAR_PATH: "/opt/signer/xml-signer.jar"
  XML_SIGNER_TIMEOUT: "10s"

  MTLS_ENABLED: "true"
  MTLS_CERT_PATH: "/certs/client-cert.pem"
  MTLS_KEY_PATH: "/certs/client-key.pem"
  MTLS_CA_PATH: "/certs/ca.pem"

  SOAP_NAMESPACE: "http://www.bcb.gov.br/pi/dict/v1"
  SOAP_ACTION_PREFIX: "http://www.bcb.gov.br/pi/dict/v1/"
```

---

## 3. Secrets Management

### Kubernetes Secrets (Manual - deprecated)

```yaml
# AVOID: Manual secrets (not recommended)
# Use External Secrets Operator instead

apiVersion: v1
kind: Secret
metadata:
  name: core-dict-secrets
  namespace: dict-prod
type: Opaque
stringData:
  database-url: "postgres://user:password@postgres-svc:5432/dict_prod?sslmode=require"
  redis-url: "redis://:password@redis-svc:6379/0"
  pulsar-url: "pulsar://pulsar-svc:6650"
  pulsar-token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### External Secrets Operator

```yaml
# Install External Secrets Operator (once per cluster)
# helm repo add external-secrets https://charts.external-secrets.io
# helm install external-secrets external-secrets/external-secrets -n kube-system

# ClusterSecretStore (AWS Secrets Manager)
apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata:
  name: aws-secrets-manager
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
      auth:
        jwt:
          serviceAccountRef:
            name: external-secrets-sa
            namespace: kube-system

---
# ExternalSecret: Core DICT (DEV)
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: core-dict-secrets
  namespace: dict-dev
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: ClusterSecretStore
  target:
    name: core-dict-secrets
    creationPolicy: Owner
  data:
  - secretKey: database-url
    remoteRef:
      key: /lbpay/dict/dev/database-url
  - secretKey: redis-url
    remoteRef:
      key: /lbpay/dict/dev/redis-url
  - secretKey: pulsar-url
    remoteRef:
      key: /lbpay/dict/dev/pulsar-url
  - secretKey: pulsar-token
    remoteRef:
      key: /lbpay/dict/dev/pulsar-token

---
# ExternalSecret: Core DICT (PROD)
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
  - secretKey: database-url
    remoteRef:
      key: /lbpay/dict/prod/database-url
  - secretKey: redis-url
    remoteRef:
      key: /lbpay/dict/prod/redis-url
  - secretKey: pulsar-url
    remoteRef:
      key: /lbpay/dict/prod/pulsar-url
  - secretKey: pulsar-token
    remoteRef:
      key: /lbpay/dict/prod/pulsar-token
```

---

## 4. AWS Secrets Manager Integration

### Create Secrets in AWS

```bash
#!/bin/bash
# create-aws-secrets.sh

set -euo pipefail

AWS_REGION="us-east-1"

# Function to create secret
create_secret() {
  local NAME=$1
  local VALUE=$2
  local DESCRIPTION=$3

  aws secretsmanager create-secret \
    --name "${NAME}" \
    --description "${DESCRIPTION}" \
    --secret-string "${VALUE}" \
    --region "${AWS_REGION}" \
    --tags Key=Project,Value=DICT Key=Team,Value=LBPay

  echo "Created secret: ${NAME}"
}

# DEV Secrets
create_secret \
  "/lbpay/dict/dev/database-url" \
  "postgres://dict_user:dev_password@postgres-dev.us-east-1.rds.amazonaws.com:5432/dict_dev?sslmode=require" \
  "PostgreSQL connection URL for DICT dev"

create_secret \
  "/lbpay/dict/dev/redis-url" \
  "redis://:dev_redis_password@redis-dev.us-east-1.cache.amazonaws.com:6379/0" \
  "Redis connection URL for DICT dev"

create_secret \
  "/lbpay/dict/dev/pulsar-url" \
  "pulsar://pulsar-dev.us-east-1.amazonaws.com:6650" \
  "Pulsar connection URL for DICT dev"

create_secret \
  "/lbpay/dict/dev/pulsar-token" \
  "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.dev_token..." \
  "Pulsar authentication token for DICT dev"

# STAGING Secrets
create_secret \
  "/lbpay/dict/staging/database-url" \
  "postgres://dict_user:staging_password@postgres-staging.us-east-1.rds.amazonaws.com:5432/dict_staging?sslmode=require" \
  "PostgreSQL connection URL for DICT staging"

create_secret \
  "/lbpay/dict/staging/redis-url" \
  "redis://:staging_redis_password@redis-staging.us-east-1.cache.amazonaws.com:6379/0" \
  "Redis connection URL for DICT staging"

# PROD Secrets
create_secret \
  "/lbpay/dict/prod/database-url" \
  "postgres://dict_user:STRONG_PROD_PASSWORD@postgres-prod.us-east-1.rds.amazonaws.com:5432/dict_prod?sslmode=require" \
  "PostgreSQL connection URL for DICT production"

create_secret \
  "/lbpay/dict/prod/redis-url" \
  "redis://:STRONG_REDIS_PASSWORD@redis-prod.us-east-1.cache.amazonaws.com:6379/0" \
  "Redis connection URL for DICT production"

create_secret \
  "/lbpay/dict/prod/pulsar-url" \
  "pulsar://pulsar-prod.us-east-1.amazonaws.com:6650" \
  "Pulsar connection URL for DICT production"

create_secret \
  "/lbpay/dict/prod/pulsar-token" \
  "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.prod_token..." \
  "Pulsar authentication token for DICT production"

echo "All secrets created successfully!"
```

### Rotate Secrets

```bash
#!/bin/bash
# rotate-secret.sh

set -euo pipefail

SECRET_NAME=$1
NEW_VALUE=$2
AWS_REGION="us-east-1"

echo "Rotating secret: ${SECRET_NAME}"

# Update secret
aws secretsmanager update-secret \
  --secret-id "${SECRET_NAME}" \
  --secret-string "${NEW_VALUE}" \
  --region "${AWS_REGION}"

# Wait for External Secrets Operator to sync (default: 1h, can force sync)
echo "Secret rotated. External Secrets Operator will sync within 1 hour."
echo "To force immediate sync, restart pods or use kubectl patch."

# Force sync (optional)
kubectl rollout restart deployment/core-dict -n dict-prod
```

---

## 5. HashiCorp Vault Integration

### Vault Configuration

```yaml
# Install Vault (Helm)
# helm repo add hashicorp https://helm.releases.hashicorp.com
# helm install vault hashicorp/vault -n vault --create-namespace

# ClusterSecretStore (Vault)
apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata:
  name: vault-backend
spec:
  provider:
    vault:
      server: "https://vault.lbpay.io"
      path: "secret"
      version: "v2"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "dict-role"
          serviceAccountRef:
            name: external-secrets-sa
            namespace: kube-system

---
# ExternalSecret: ICP-Brasil Certificates (Vault)
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: icp-brasil-certs
  namespace: dict-prod
spec:
  refreshInterval: 24h
  secretStoreRef:
    name: vault-backend
    kind: ClusterSecretStore
  target:
    name: icp-brasil-certs
    creationPolicy: Owner
  data:
  - secretKey: client-cert.pem
    remoteRef:
      key: secret/data/lbpay/dict/prod/icp-brasil
      property: client-cert
  - secretKey: client-key.pem
    remoteRef:
      key: secret/data/lbpay/dict/prod/icp-brasil
      property: client-key
  - secretKey: ca.pem
    remoteRef:
      key: secret/data/lbpay/dict/prod/icp-brasil
      property: ca-cert
  - secretKey: signing-cert.pem
    remoteRef:
      key: secret/data/lbpay/dict/prod/icp-brasil
      property: signing-cert
  - secretKey: signing-key.pem
    remoteRef:
      key: secret/data/lbpay/dict/prod/icp-brasil
      property: signing-key
```

### Store Secrets in Vault

```bash
#!/bin/bash
# store-vault-secrets.sh

set -euo pipefail

VAULT_ADDR="https://vault.lbpay.io"
VAULT_TOKEN="<your-token>"

export VAULT_ADDR
export VAULT_TOKEN

# Enable KV v2 secrets engine
vault secrets enable -path=secret kv-v2

# Store ICP-Brasil certificates (PROD)
vault kv put secret/lbpay/dict/prod/icp-brasil \
  client-cert="$(cat certs/prod/client-cert.pem)" \
  client-key="$(cat certs/prod/client-key.pem)" \
  ca-cert="$(cat certs/prod/ca.pem)" \
  signing-cert="$(cat certs/prod/signing-cert.pem)" \
  signing-key="$(cat certs/prod/signing-key.pem)"

echo "Secrets stored in Vault successfully!"
```

---

## 6. Feature Flags

### LaunchDarkly Integration

```yaml
# LaunchDarkly SDK Configuration

# Environment Variables (ConfigMap)
FEATURE_FLAG_URL: "https://app.launchdarkly.com"
FEATURE_FLAG_SDK_KEY: "<from-aws-secrets-manager>"
FEATURE_FLAG_ENV: "production"

# Feature Flags (examples)
Features:
  - enable-portability:
      description: "Enable DICT portability feature"
      default: false
      prod: true
      staging: true
      dev: true

  - enable-claim:
      description: "Enable DICT claim feature"
      default: false
      prod: true
      staging: true
      dev: true

  - enable-rate-limiting:
      description: "Enable API rate limiting"
      default: true
      prod: true
      staging: true
      dev: false

  - bacen-timeout-seconds:
      description: "Bacen API timeout in seconds"
      default: 30
      prod: 30
      staging: 60
      dev: 120

  - maintenance-mode:
      description: "Enable maintenance mode (read-only)"
      default: false
      prod: false
      staging: false
      dev: false
```

### Feature Flag Usage (Go)

```go
// feature_flags.go

import (
    ld "github.com/launchdarkly/go-server-sdk/v6"
    "github.com/launchdarkly/go-server-sdk/v6/ldcomponents"
)

type FeatureFlags struct {
    client *ld.LDClient
}

func NewFeatureFlags(sdkKey string) (*FeatureFlags, error) {
    client, err := ld.MakeClient(sdkKey, 5*time.Second)
    if err != nil {
        return nil, err
    }

    return &FeatureFlags{client: client}, nil
}

func (f *FeatureFlags) IsPortabilityEnabled(ctx context.Context, userID string) bool {
    user := ld.NewUser(userID)
    return f.client.BoolVariation("enable-portability", user, false)
}

func (f *FeatureFlags) GetBacenTimeout(ctx context.Context) int {
    user := ld.NewAnonymousUser()
    return f.client.IntVariation("bacen-timeout-seconds", user, 30)
}

func (f *FeatureFlags) IsMaintenanceMode(ctx context.Context) bool {
    user := ld.NewAnonymousUser()
    return f.client.BoolVariation("maintenance-mode", user, false)
}

// Usage in handler
func (s *Server) CreateEntry(ctx context.Context, req *pb.CreateEntryRequest) (*pb.CreateEntryResponse, error) {
    if s.featureFlags.IsMaintenanceMode(ctx) {
        return nil, status.Error(codes.Unavailable, "System in maintenance mode")
    }

    // ... create entry logic
}
```

---

## 7. Promotion Strategy

### Environment Promotion Flow

```yaml
Promotion Flow:

DEV → STAGING → PROD

Step 1: DEV Deployment
  Trigger: Push to develop branch
  Action: Auto-deploy to DEV
  Approval: None
  Tests: Smoke tests

Step 2: STAGING Deployment
  Trigger: Push to main branch OR Manual
  Action: Auto-deploy to STAGING
  Approval: 1 tech lead
  Tests: Integration tests, E2E tests, Performance tests

Step 3: PROD Deployment
  Trigger: Manual (workflow_dispatch)
  Action: Canary deployment (10% → 50% → 100%)
  Approval: 2 approvers (tech lead + ops manager)
  Tests: Smoke tests, SLO verification

Rollback:
  DEV: kubectl rollout undo
  STAGING: ArgoCD rollback
  PROD: ArgoCD rollback + PagerDuty incident
```

### Configuration Promotion

```bash
#!/bin/bash
# promote-config.sh

set -euo pipefail

SOURCE_ENV=$1  # dev, staging
TARGET_ENV=$2  # staging, prod

echo "Promoting config from ${SOURCE_ENV} to ${TARGET_ENV}"

# Validate
if [[ "${SOURCE_ENV}" == "prod" ]]; then
  echo "ERROR: Cannot promote from prod"
  exit 1
fi

if [[ "${SOURCE_ENV}" == "staging" && "${TARGET_ENV}" != "prod" ]]; then
  echo "ERROR: Can only promote staging to prod"
  exit 1
fi

# Copy ConfigMaps (review before applying)
kubectl get configmap -n "dict-${SOURCE_ENV}" -o yaml > /tmp/configmap-${SOURCE_ENV}.yaml

echo "Review ConfigMap at /tmp/configmap-${SOURCE_ENV}.yaml"
echo "Update namespace to dict-${TARGET_ENV}"
echo "Adjust values for ${TARGET_ENV} environment"
echo ""
echo "Then apply:"
echo "  kubectl apply -f /tmp/configmap-${SOURCE_ENV}.yaml"
```

### Secret Promotion

```bash
#!/bin/bash
# promote-secrets.sh

set -euo pipefail

SOURCE_ENV=$1
TARGET_ENV=$2

echo "Promoting secrets from ${SOURCE_ENV} to ${TARGET_ENV}"

# Secrets are managed in AWS Secrets Manager
# Simply copy values between environments

aws secretsmanager get-secret-value \
  --secret-id "/lbpay/dict/${SOURCE_ENV}/database-url" \
  --query SecretString \
  --output text > /tmp/db-url.txt

echo "Review /tmp/db-url.txt"
echo "Adjust for ${TARGET_ENV} (host, credentials, etc.)"
echo ""
echo "Then update:"
echo "  aws secretsmanager put-secret-value \\"
echo "    --secret-id /lbpay/dict/${TARGET_ENV}/database-url \\"
echo "    --secret-string \"\$(cat /tmp/db-url.txt)\""
```

---

## 8. Rastreabilidade

### Documentos Relacionados

| ID | Documento | Relacao |
|----|-----------|---------|
| **DEV-001** | [CI/CD Pipeline Core](./Pipelines/DEV-001_CI_CD_Pipeline_Core.md) | Pipeline usa environment configs |
| **DEV-004** | [Kubernetes Manifests](./DEV-004_Kubernetes_Manifests.md) | Manifests referenciam ConfigMaps/Secrets |
| **SEC-001** | [mTLS Configuration](../13_Seguranca/SEC-001_mTLS_Configuration.md) | Certificados em Vault |

### Metricas de Sucesso

```yaml
Configuration Management:
  - Secrets: 0 hardcoded em codigo ou manifests
  - Secret rotation: Automatica (90 dias)
  - Config drift: 0 (GitOps via ArgoCD)
  - Feature flags: >= 5 features controlled

Security:
  - Secrets encrypted at rest (AWS KMS)
  - Secrets encrypted in transit (TLS)
  - Audit logs: 100% dos acessos a secrets
  - RBAC: Least privilege (ninguem tem acesso direto a secrets)

Environments:
  - DEV uptime: >= 95%
  - STAGING uptime: >= 99%
  - PROD uptime: >= 99.9%
  - Promotion time: < 30 minutos (dev → staging → prod)
```

---

**Ultima Atualizacao**: 2025-10-25
**Versao**: 1.0
**Status**: Completo
