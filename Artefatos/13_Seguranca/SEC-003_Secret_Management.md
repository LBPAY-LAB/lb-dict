# SEC-003: Secret Management

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Versão**: 1.0
**Data**: 2025-10-25
**Status**: ✅ Especificação Completa
**Responsável**: DevOps Lead + Security Lead

---

## 📋 Resumo Executivo

Este documento especifica a **gestão completa de secrets** (senhas, API keys, certificados, tokens) no sistema DICT, incluindo armazenamento seguro, rotação, acesso controlado, e auditoria.

**Objetivo**: Garantir que NENHUM secret seja armazenado em plaintext em código, configuração, ou repositório Git, e que todos os secrets sejam gerenciados de forma centralizada e segura.

---

## 🎯 Tipos de Secrets

### Inventário de Secrets no DICT

| Secret | Tipo | Usado Por | Rotação | Criticidade |
|--------|------|-----------|---------|-------------|
| **PostgreSQL password** | Database credential | Connect, Core | 90 dias | 🔴 Crítico |
| **Redis password** | Database credential | Connect | 90 dias | 🟡 Alto |
| **Temporal API key** | API credential | Connect, Worker | 90 dias | 🟡 Alto |
| **Pulsar auth token** | API credential | Connect, Worker | 90 dias | 🟡 Alto |
| **ICP-Brasil certificate** | X.509 certificate | Bridge | 1 ano | 🔴 Crítico |
| **ICP-Brasil private key** | Private key (HSM) | Bridge | 1 ano | 🔴 Crítico |
| **JWT signing key** | Symmetric key | Core DICT | 90 dias | 🔴 Crítico |
| **Bacen API credentials** | API credential | Bridge | Manual | 🔴 Crítico |
| **GitHub token (CI/CD)** | API credential | GitHub Actions | 90 dias | 🟡 Alto |
| **Slack webhook URL** | Webhook URL | Alertas | Manual | 🟢 Baixo |
| **Encryption keys (AES)** | Symmetric key | Backup encryption | 1 ano | 🔴 Crítico |

---

## 🔐 Solução de Secret Management

### Opções Avaliadas

| Solução | Pros | Cons | Recomendação |
|---------|------|------|--------------|
| **HashiCorp Vault** | ✅ Open-source<br>✅ Dynamic secrets<br>✅ Audit log | ❌ Complexidade de setup<br>❌ Requer infraestrutura dedicada | ✅ **Recomendado** para produção |
| **AWS Secrets Manager** | ✅ Managed service<br>✅ Integração AWS | ❌ Vendor lock-in<br>❌ Custo ($0.40/secret/mês) | ⚠️ Alternativa (se já em AWS) |
| **Kubernetes Secrets** | ✅ Nativo do K8s<br>✅ Fácil setup | ❌ Secrets em base64 (não criptografados por padrão)<br>❌ Sem rotação automática | ❌ NÃO recomendado (apenas para dev) |
| **Google Secret Manager** | ✅ Managed service<br>✅ Integração GCP | ❌ Vendor lock-in | ⚠️ Alternativa (se já em GCP) |

**Decisão**: **HashiCorp Vault** (open-source, multi-cloud, dynamic secrets)

---

## 🏗️ Arquitetura Vault

### Deployment

```
                    [Vault Cluster]
                    (3 nodes - HA)
                           │
                ┌──────────┼──────────┐
                │          │          │
           [Vault 1]  [Vault 2]  [Vault 3]
                │          │          │
                └──────────┼──────────┘
                           │
                    [Consul Backend]
                    (Storage + HA)
                           │
            ┌──────────────┼──────────────┐
            │              │              │
      [Connect Pods]  [Bridge Pods]  [Core Pods]
```

### Vault Configuration

```hcl
# vault-config.hcl (especificação)

# Storage backend (Consul para HA)
storage "consul" {
  address = "consul.vault.svc.cluster.local:8500"
  path    = "vault/"
}

# Listener (HTTPS only)
listener "tcp" {
  address     = "0.0.0.0:8200"
  tls_cert_file = "/vault/tls/tls.crt"
  tls_key_file  = "/vault/tls/tls.key"
}

# Seal (auto-unseal com AWS KMS ou GCP KMS)
seal "awskms" {
  region     = "us-east-1"
  kms_key_id = "arn:aws:kms:us-east-1:123456789:key/vault-unseal"
}

# Telemetry
telemetry {
  prometheus_retention_time = "30s"
  disable_hostname = true
}

# UI
ui = true

# API address
api_addr = "https://vault.dict.svc.cluster.local:8200"
cluster_addr = "https://vault.dict.svc.cluster.local:8201"
```

---

## 🔑 Secret Engines

### 1. Database Secrets (Dynamic)

**Propósito**: Gerar credenciais PostgreSQL/Redis dinamicamente com TTL curto

```bash
# Habilitar database secrets engine
vault secrets enable database

# Configurar conexão com PostgreSQL
vault write database/config/postgres \
  plugin_name=postgresql-database-plugin \
  allowed_roles="dict-connect-role" \
  connection_url="postgresql://{{username}}:{{password}}@postgres.dict.svc:5432/dict?sslmode=require" \
  username="vault_admin" \
  password="<admin_password>"

# Criar role com permissões limitadas
vault write database/roles/dict-connect-role \
  db_name=postgres \
  creation_statements="
    CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';
    GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA dict TO \"{{name}}\";
  " \
  default_ttl="1h" \
  max_ttl="24h"

# Aplicação obtém credenciais temporárias
vault read database/creds/dict-connect-role
# Output:
# Key                Value
# lease_id           database/creds/dict-connect-role/abc123
# lease_duration     1h
# username           v-dict-connect-XYZ123
# password           A1Bb2Cc3Dd4Ee5Ff
```

**Vantagens**:
- ✅ Credenciais temporárias (1h TTL)
- ✅ Rotação automática
- ✅ Revogação imediata (revoke lease)
- ✅ Auditoria de quem acessou o banco

---

### 2. KV Secrets v2 (Static)

**Propósito**: Armazenar secrets estáticos (API keys, webhooks, etc.)

```bash
# Habilitar KV v2 engine
vault secrets enable -path=dict kv-v2

# Armazenar secret
vault kv put dict/connect/pulsar \
  broker_url="pulsar://pulsar.pulsar.svc:6650" \
  auth_token="<token>"

# Ler secret
vault kv get dict/connect/pulsar

# Versioning (KV v2 mantém histórico)
vault kv get -version=1 dict/connect/pulsar  # Versão anterior
vault kv metadata get dict/connect/pulsar    # Metadata (criado em, atualizado em)
```

---

### 3. PKI (Certificates)

**Propósito**: Gerar certificados TLS internos (para mTLS entre serviços)

```bash
# Habilitar PKI engine
vault secrets enable pki

# Configurar CA raiz
vault write pki/root/generate/internal \
  common_name="DICT Internal CA" \
  ttl=87600h  # 10 anos

# Configurar URLs de CRL
vault write pki/config/urls \
  issuing_certificates="https://vault.dict.svc:8200/v1/pki/ca" \
  crl_distribution_points="https://vault.dict.svc:8200/v1/pki/crl"

# Criar role para emitir certificados
vault write pki/roles/dict-services \
  allowed_domains="dict.svc.cluster.local" \
  allow_subdomains=true \
  max_ttl="720h"  # 30 dias

# Gerar certificado
vault write pki/issue/dict-services \
  common_name="connect.dict.svc.cluster.local" \
  ttl="720h"
```

---

## 🔐 Authentication Methods

### 1. Kubernetes Auth (Para Pods)

```bash
# Habilitar Kubernetes auth
vault auth enable kubernetes

# Configurar
vault write auth/kubernetes/config \
  kubernetes_host="https://kubernetes.default.svc:443" \
  kubernetes_ca_cert=@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt \
  token_reviewer_jwt=@/var/run/secrets/kubernetes.io/serviceaccount/token

# Criar policy para Connect
vault policy write dict-connect-policy - <<EOF
path "database/creds/dict-connect-role" {
  capabilities = ["read"]
}

path "dict/data/connect/*" {
  capabilities = ["read"]
}
EOF

# Vincular ServiceAccount → Policy
vault write auth/kubernetes/role/dict-connect \
  bound_service_account_names=dict-connect \
  bound_service_account_namespaces=dict-prod \
  policies=dict-connect-policy \
  ttl=1h
```

**Uso no Pod** (Pseudocódigo Go):
```go
// Connect pod autentica com Vault usando ServiceAccount token
func authenticateWithVault() (*vault.Client, error) {
    client, err := vault.NewClient(&vault.Config{
        Address: "https://vault.dict.svc.cluster.local:8200",
    })
    if err != nil {
        return nil, err
    }

    // Ler ServiceAccount token
    jwt, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
    if err != nil {
        return nil, err
    }

    // Autenticar com Vault
    secret, err := client.Logical().Write("auth/kubernetes/login", map[string]interface{}{
        "role": "dict-connect",
        "jwt":  string(jwt),
    })
    if err != nil {
        return nil, err
    }

    // Obter token Vault
    client.SetToken(secret.Auth.ClientToken)

    return client, nil
}
```

---

### 2. AppRole (Para CI/CD)

```bash
# Habilitar AppRole auth
vault auth enable approle

# Criar policy para GitHub Actions
vault policy write github-actions-policy - <<EOF
path "dict/data/ci/*" {
  capabilities = ["read"]
}
EOF

# Criar AppRole
vault write auth/approle/role/github-actions \
  secret_id_ttl=10m \
  token_ttl=20m \
  token_max_ttl=30m \
  policies=github-actions-policy

# Obter role_id (armazenar como GitHub Secret)
vault read auth/approle/role/github-actions/role-id

# Gerar secret_id (temporário, para cada build)
vault write -f auth/approle/role/github-actions/secret-id
```

---

## 🔄 Rotação de Secrets

### Rotação Automática

**Database Credentials** (Dynamic):
- ✅ Rotação automática via Vault (renovar lease)
- TTL: 1h
- Aplicação renova lease automaticamente a cada 30min

**Certificates** (PKI):
- ✅ Renovar 7 dias antes do vencimento
- Cronjob Kubernetes para renovação

### Rotação Manual

**ICP-Brasil Certificate**:
- Processo manual (ver [SEC-002](SEC-002_ICP_Brasil_Certificates.md))
- Renovar 90 dias antes do vencimento
- Blue-green deployment para troca sem downtime

**JWT Signing Key**:
```bash
# 1. Gerar nova key
openssl rand -base64 32 > jwt-signing-key-v2

# 2. Armazenar no Vault
vault kv put dict/core/jwt signing_key=@jwt-signing-key-v2 version=v2

# 3. Atualizar aplicação para usar nova key (rolling update)
kubectl set env deployment/dict-core JWT_KEY_VERSION=v2

# 4. Aguardar 7 dias (para tokens antigos expirarem)
# 5. Deletar key antiga
vault kv delete dict/core/jwt-signing-key-v1
```

---

## 📊 Auditoria de Secrets

### Vault Audit Log

```bash
# Habilitar audit logging
vault audit enable file file_path=/vault/logs/audit.log

# Formato de log (JSON)
{
  "time": "2025-10-25T10:00:00Z",
  "type": "response",
  "auth": {
    "token_type": "service",
    "policies": ["dict-connect-policy"]
  },
  "request": {
    "operation": "read",
    "path": "database/creds/dict-connect-role",
    "client_token_accessor": "abc123"
  },
  "response": {
    "data": {
      "username": "v-dict-connect-XYZ123"
    }
  }
}
```

### Monitoramento

```yaml
# Prometheus metrics (Vault exporter)
vault_secret_kv_count{mount="dict"} 42
vault_secret_lease_creation{role="dict-connect-role"} 158
vault_secret_lease_expiration{role="dict-connect-role"} 142

# Alertas
groups:
  - name: vault_secrets
    rules:
      - alert: VaultSecretLeaseExpiringSoon
        expr: vault_secret_lease_expiration < 300  # < 5min
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Vault lease expiring soon"

      - alert: VaultUnsealedNodes
        expr: vault_core_unsealed < 2  # Menos de 2 nodes unsealed
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Vault cluster degradado"
```

---

## 🚨 Segurança do Vault

### 1. Acesso ao Vault

**Princípio**: Least Privilege

```hcl
# Policy para Connect (apenas read secrets específicos)
path "database/creds/dict-connect-role" {
  capabilities = ["read"]
}

path "dict/data/connect/*" {
  capabilities = ["read"]
}

# Policy para Admin (pode criar/atualizar secrets)
path "dict/data/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Policy para Auditor (apenas read e list)
path "dict/data/*" {
  capabilities = ["read", "list"]
}
```

---

### 2. Unseal Keys

**Problema**: Vault inicia em estado "sealed" (criptografado)

**Soluções**:

#### Opção A: Shamir Secret Sharing (Manual)
```bash
# Inicializar Vault com 5 key shares (threshold 3)
vault operator init -key-shares=5 -key-threshold=3

# Output:
# Unseal Key 1: abc...
# Unseal Key 2: def...
# Unseal Key 3: ghi...
# Unseal Key 4: jkl...
# Unseal Key 5: mno...
# Initial Root Token: s.xyz...

# Distribuir keys entre 5 pessoas diferentes
# Necessário 3 keys para unseal
vault operator unseal <key1>
vault operator unseal <key2>
vault operator unseal <key3>
```

**Desvantagens**: Manual, requer intervenção humana ao reiniciar Vault

#### Opção B: Auto-Unseal (Recomendado)
```hcl
# vault-config.hcl
seal "awskms" {
  region     = "us-east-1"
  kms_key_id = "arn:aws:kms:us-east-1:123456789:key/vault-unseal"
}
```

**Vantagens**:
- ✅ Unseal automático ao reiniciar
- ✅ Sem intervenção manual
- ✅ Keys gerenciadas pelo Cloud Provider (AWS KMS, GCP KMS)

---

### 3. Root Token

**Problema**: Root token tem acesso total (equivalente a root em Linux)

**Boas Práticas**:
- ❌ **NUNCA** usar root token em produção
- ✅ Revogar root token inicial: `vault token revoke <root-token>`
- ✅ Gerar novo root token apenas em emergências (processo de recovery)
- ✅ Usar policies específicas para cada serviço

---

## 📋 Checklist de Implementação

- [ ] Provisionar Vault cluster (3 nodes HA)
- [ ] Configurar Consul backend para storage
- [ ] Configurar auto-unseal (AWS KMS ou GCP KMS)
- [ ] Habilitar audit logging
- [ ] Configurar Kubernetes auth method
- [ ] Criar policies para cada componente (Connect, Bridge, Core)
- [ ] Habilitar database secrets engine (PostgreSQL, Redis)
- [ ] Habilitar KV v2 engine
- [ ] Migrar secrets de Kubernetes Secrets para Vault
- [ ] Atualizar pods para autenticar com Vault
- [ ] Configurar lease renewal automático
- [ ] Criar cronjobs para rotação de certificates
- [ ] Configurar monitoramento (Prometheus metrics)
- [ ] Configurar alertas (lease expiration, unseal status)
- [ ] Documentar processo de emergency access (root token generation)
- [ ] Treinar equipe DevOps em operação do Vault

---

## 🔧 Integração com Aplicações

### Go Example (Connect)

```go
// Pseudocódigo (especificação)
package secrets

import (
    "github.com/hashicorp/vault/api"
)

// VaultClient encapsula acesso ao Vault
type VaultClient struct {
    client *api.Client
}

// NewVaultClient cria cliente autenticado
func NewVaultClient() (*VaultClient, error) {
    config := api.DefaultConfig()
    config.Address = "https://vault.dict.svc.cluster.local:8200"

    client, err := api.NewClient(config)
    if err != nil {
        return nil, err
    }

    // Autenticar com Kubernetes ServiceAccount
    jwt, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
    if err != nil {
        return nil, err
    }

    secret, err := client.Logical().Write("auth/kubernetes/login", map[string]interface{}{
        "role": "dict-connect",
        "jwt":  string(jwt),
    })
    if err != nil {
        return nil, err
    }

    client.SetToken(secret.Auth.ClientToken)

    return &VaultClient{client: client}, nil
}

// GetDatabaseCreds obtém credenciais temporárias do PostgreSQL
func (v *VaultClient) GetDatabaseCreds() (*DatabaseCreds, error) {
    secret, err := v.client.Logical().Read("database/creds/dict-connect-role")
    if err != nil {
        return nil, err
    }

    return &DatabaseCreds{
        Username:      secret.Data["username"].(string),
        Password:      secret.Data["password"].(string),
        LeaseDuration: secret.LeaseDuration,
        LeaseID:       secret.LeaseID,
    }, nil
}

// RenewLease renova lease antes de expirar
func (v *VaultClient) RenewLease(leaseID string) error {
    _, err := v.client.Sys().Renew(leaseID, 0)  // 0 = incremento padrão
    return err
}

// GetStaticSecret obtém secret estático do KV
func (v *VaultClient) GetStaticSecret(path string) (map[string]interface{}, error) {
    secret, err := v.client.Logical().Read("dict/data/" + path)
    if err != nil {
        return nil, err
    }

    return secret.Data["data"].(map[string]interface{}), nil
}
```

---

## 📚 Referências

### Documentos Internos
- [SEC-002: ICP-Brasil Certificates](SEC-002_ICP_Brasil_Certificates.md) - Certificados
- [SEC-001: mTLS Configuration](SEC-001_mTLS_Configuration.md)
- [DevOps Pipelines](../../15_DevOps/Pipelines/)

### Documentação Externa
- [HashiCorp Vault Documentation](https://www.vaultproject.io/docs)
- [Vault Kubernetes Auth](https://www.vaultproject.io/docs/auth/kubernetes)
- [Dynamic Database Secrets](https://www.vaultproject.io/docs/secrets/databases)
- [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/)
- [OWASP Secret Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)

---

**Versão**: 1.0
**Status**: ✅ Especificação Completa (Aguardando implementação)
**Próxima Revisão**: Após setup do Vault cluster

---

**IMPORTANTE**: Este é um documento de **especificação técnica e operacional**. A implementação será feita pela equipe de DevOps em fase posterior, baseando-se neste documento.
