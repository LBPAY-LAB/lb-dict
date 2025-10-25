# GRPC-001: Bridge gRPC Service Specification

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Serviço**: Bridge gRPC Service (Connect → Bridge)
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: ARCHITECT (AI Agent - Technical Architect)
**Revisor**: [Aguardando]
**Aprovador**: Tech Lead, Head de Arquitetura

---

## Sumário Executivo

Este documento especifica o **contrato gRPC** entre o **RSFN Connect** (cliente) e o **RSFN Bridge** (servidor), definindo todos os RPCs (Remote Procedure Calls), mensagens, tipos de dados e tratamento de erros.

**Baseado em**:
- [TEC-002 v3.1: Bridge Specification](../../11_Especificacoes_Tecnicas/TEC-002_Bridge_Specification.md)
- [TEC-003 v2.1: Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [ANA-002: Análise Repositório Bridge](../../00_Analises/ANA-002_Analise_Repo_Bridge.md)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | ARCHITECT | Versão inicial - Contrato gRPC Bridge |

---

## Índice

1. [Visão Geral](#1-visão-geral)
2. [Service Definition](#2-service-definition)
3. [Message Types](#3-message-types)
4. [RPCs Detalhados](#4-rpcs-detalhados)
5. [Error Handling](#5-error-handling)
6. [Security](#6-security)
7. [Performance](#7-performance)

---

## 1. Visão Geral

### 1.1. Arquitetura

```
┌─────────────────────────────────────┐
│      RSFN Connect (gRPC Client)     │
│                                     │
│  - ClaimWorkflow Activities         │
│  - Entry Management Activities      │
│  - VSYNC Activities (futuro)        │
└─────────────────┬───────────────────┘
                  │ gRPC
                  │ TLS 1.2+
                  │ Proto3
                  ↓
┌─────────────────────────────────────┐
│      RSFN Bridge (gRPC Server)      │
│                                     │
│  - BridgeService                    │
│  - SOAP/XML Adapter                 │
│  - mTLS → Bacen DICT                │
└─────────────────────────────────────┘
```

### 1.2. Características

| Aspecto | Valor |
|---------|-------|
| **Protocol** | gRPC (HTTP/2) |
| **Serialization** | Protocol Buffers v3 |
| **Transport** | TLS 1.2+ |
| **Port** | 50051 (padrão gRPC) |
| **Timeout** | 30s (síncrono), 60s (claims) |
| **Retry Policy** | Exponential backoff |
| **Load Balancing** | Round-robin (K8s) |

---

## 2. Service Definition

### 2.1. BridgeService (Principal)

```protobuf
syntax = "proto3";

package rsfn.bridge.v1;

option go_package = "github.com/lbpay/rsfn-bridge/api/v1;bridgev1";

import "google/protobuf/timestamp.proto";
import "google/protobuf/empty.proto";

// BridgeService: Serviço principal para comunicação com Bacen DICT
service BridgeService {
  // ==========================================
  // Entry Management (Chaves PIX)
  // ==========================================

  // CreateEntry: Criar chave PIX no Bacen
  rpc CreateEntry(CreateEntryRequest) returns (CreateEntryResponse);

  // GetEntry: Consultar chave PIX no Bacen
  rpc GetEntry(GetEntryRequest) returns (GetEntryResponse);

  // UpdateEntry: Atualizar dados de chave PIX
  rpc UpdateEntry(UpdateEntryRequest) returns (UpdateEntryResponse);

  // DeleteEntry: Deletar chave PIX (soft delete)
  rpc DeleteEntry(DeleteEntryRequest) returns (DeleteEntryResponse);

  // ==========================================
  // Claim Management (Reivindicações - 30 dias)
  // ==========================================

  // CreateClaim: Iniciar reivindicação de chave
  rpc CreateClaim(CreateClaimRequest) returns (CreateClaimResponse);

  // GetClaim: Consultar status de reivindicação
  rpc GetClaim(GetClaimRequest) returns (GetClaimResponse);

  // CompleteClaim: Finalizar reivindicação (aprovar)
  rpc CompleteClaim(CompleteClaimRequest) returns (CompleteClaimResponse);

  // CancelClaim: Cancelar reivindicação
  rpc CancelClaim(CancelClaimRequest) returns (CancelClaimResponse);

  // ==========================================
  // Portability Management (Portabilidade)
  // ==========================================

  // InitiatePortability: Iniciar portabilidade de chave
  rpc InitiatePortability(InitiatePortabilityRequest) returns (InitiatePortabilityResponse);

  // ConfirmPortability: Confirmar portabilidade
  rpc ConfirmPortability(ConfirmPortabilityRequest) returns (ConfirmPortabilityResponse);

  // CancelPortability: Cancelar portabilidade
  rpc CancelPortability(CancelPortabilityRequest) returns (CancelPortabilityResponse);

  // ==========================================
  // Directory Queries (Consultas DICT)
  // ==========================================

  // GetDirectory: Consultar diretório completo
  rpc GetDirectory(GetDirectoryRequest) returns (GetDirectoryResponse);

  // SearchEntries: Buscar chaves por critérios
  rpc SearchEntries(SearchEntriesRequest) returns (SearchEntriesResponse);

  // ==========================================
  // Health & Monitoring
  // ==========================================

  // HealthCheck: Verificar saúde do Bridge e conectividade com Bacen
  rpc HealthCheck(google.protobuf.Empty) returns (HealthCheckResponse);
}
```

---

## 3. Message Types

### 3.1. Common Types

```protobuf
// KeyType: Tipos de chave PIX suportados
enum KeyType {
  KEY_TYPE_UNSPECIFIED = 0;
  KEY_TYPE_CPF = 1;
  KEY_TYPE_CNPJ = 2;
  KEY_TYPE_EMAIL = 3;
  KEY_TYPE_PHONE = 4;
  KEY_TYPE_EVP = 5;  // Chave aleatória (UUID)
}

// EntryStatus: Status de uma chave PIX
enum EntryStatus {
  ENTRY_STATUS_UNSPECIFIED = 0;
  ENTRY_STATUS_PENDING = 1;
  ENTRY_STATUS_ACTIVE = 2;
  ENTRY_STATUS_PORTABILITY_REQUESTED = 3;
  ENTRY_STATUS_OWNERSHIP_CONFIRMED = 4;
  ENTRY_STATUS_DELETED = 5;
  ENTRY_STATUS_CLAIM_PENDING = 6;
}

// ClaimStatus: Status de reivindicação (30 dias)
enum ClaimStatus {
  CLAIM_STATUS_UNSPECIFIED = 0;
  CLAIM_STATUS_OPEN = 1;
  CLAIM_STATUS_WAITING_RESOLUTION = 2;
  CLAIM_STATUS_CONFIRMED = 3;
  CLAIM_STATUS_CANCELLED = 4;
  CLAIM_STATUS_COMPLETED = 5;
  CLAIM_STATUS_EXPIRED = 6;  // Após 30 dias
}

// AccountType: Tipo de conta CID
enum AccountType {
  ACCOUNT_TYPE_UNSPECIFIED = 0;
  ACCOUNT_TYPE_CACC = 1;  // Conta Corrente
  ACCOUNT_TYPE_SVGS = 2;  // Poupança
  ACCOUNT_TYPE_SLRY = 3;  // Salário
  ACCOUNT_TYPE_TRAN = 4;  // Transacional
}

// Account: Dados de conta CID
message Account {
  string account_number = 1;
  string branch_code = 2;
  AccountType account_type = 3;
  string holder_document = 4;  // CPF ou CNPJ
  string holder_name = 5;
  string participant_ispb = 6;
}

// Entry: Representação de uma chave PIX
message Entry {
  string entry_id = 1;  // ID interno (UUID)
  string external_id = 2;  // ID do Bacen
  KeyType key_type = 3;
  string key_value = 4;
  Account account = 5;
  EntryStatus status = 6;
  google.protobuf.Timestamp created_at = 7;
  google.protobuf.Timestamp updated_at = 8;
}
```

---

## 4. RPCs Detalhados

### 4.1. CreateEntry

**Descrição**: Criar nova chave PIX no Bacen DICT.

```protobuf
message CreateEntryRequest {
  // Dados da chave
  KeyType key_type = 1;
  string key_value = 2;

  // Conta vinculada
  Account account = 3;

  // Metadados
  string idempotency_key = 4;  // Para evitar duplicação
  string correlation_id = 5;   // Rastreamento E2E
}

message CreateEntryResponse {
  // Entry criado
  Entry entry = 1;

  // Resposta do Bacen
  string bacen_transaction_id = 2;
  google.protobuf.Timestamp bacen_timestamp = 3;
}
```

**Casos de Uso**:
- Usuário cadastra nova chave PIX
- FrontEnd → Core DICT → Connect → **Bridge → Bacen**

**Timeout**: 30s

---

### 4.2. GetEntry

**Descrição**: Consultar chave PIX existente no Bacen.

```protobuf
message GetEntryRequest {
  oneof identifier {
    string entry_id = 1;      // ID interno
    string external_id = 2;   // ID Bacen
    KeyQuery key_query = 3;   // Por tipo + valor
  }

  string correlation_id = 4;
}

message KeyQuery {
  KeyType key_type = 1;
  string key_value = 2;
}

message GetEntryResponse {
  Entry entry = 1;
  bool found = 2;
}
```

**Casos de Uso**:
- Consulta de chave antes de transferência PIX
- Validação de chave em cadastro

**Timeout**: 15s (query rápida)

---

### 4.3. CreateClaim (⭐ Crítico - 30 dias)

**Descrição**: Iniciar reivindicação de chave PIX com período de 30 dias.

```protobuf
message CreateClaimRequest {
  // Chave a ser reivindicada
  string entry_id = 1;
  KeyType key_type = 2;
  string key_value = 3;

  // Reivindicante
  Account claimer_account = 4;
  string claimer_ispb = 5;

  // Dono atual
  string owner_ispb = 6;

  // Período de resolução (sempre 30 dias conforme TEC-003 v2.1)
  int32 completion_period_days = 7;  // Deve ser 30

  // Metadados
  string idempotency_key = 8;
  string correlation_id = 9;
}

message CreateClaimResponse {
  // Claim criado
  Claim claim = 1;

  // Bacen response
  string bacen_claim_id = 2;
  google.protobuf.Timestamp expires_at = 3;  // created_at + 30 dias
}

message Claim {
  string claim_id = 1;
  string external_id = 2;  // ID Bacen
  string entry_id = 3;
  ClaimStatus status = 4;
  int32 completion_period_days = 5;  // 30
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp expires_at = 7;
  string claimer_ispb = 8;
  string owner_ispb = 9;
}
```

**Validações**:
- `completion_period_days` DEVE ser 30 (enforcement)
- `expires_at` = `created_at` + 30 dias
- `claimer_ispb` ≠ `owner_ispb`

**Timeout**: 60s (operação complexa)

---

### 4.4. GetClaim

**Descrição**: Consultar status de reivindicação.

```protobuf
message GetClaimRequest {
  oneof identifier {
    string claim_id = 1;
    string external_id = 2;
  }

  string correlation_id = 3;
}

message GetClaimResponse {
  Claim claim = 1;
  bool found = 2;

  // Informações adicionais
  int32 days_remaining = 3;  // Dias até expiração
  bool expired = 4;          // Se já passou dos 30 dias
}
```

**Casos de Uso**:
- MonitorStatusWorkflow consulta status periodicamente
- Dashboard mostra claims pendentes

**Timeout**: 15s

---

### 4.5. CompleteClaim

**Descrição**: Finalizar reivindicação (aprovação).

```protobuf
message CompleteClaimRequest {
  string claim_id = 1;
  string external_id = 2;

  // Decisão
  string resolution_reason = 3;  // Motivo da aprovação

  // Metadados
  string correlation_id = 4;
}

message CompleteClaimResponse {
  Claim claim = 1;  // Status = COMPLETED

  // Entry atualizado
  Entry updated_entry = 2;  // Agora pertence ao claimer

  // Bacen response
  string bacen_transaction_id = 3;
  google.protobuf.Timestamp completed_at = 4;
}
```

**Efeito Colateral**:
- Entry transferido para claimer_account
- Status do Entry → OWNERSHIP_CONFIRMED

**Timeout**: 60s

---

### 4.6. CancelClaim

**Descrição**: Cancelar reivindicação.

```protobuf
message CancelClaimRequest {
  string claim_id = 1;
  string external_id = 2;

  // Razão do cancelamento
  string cancellation_reason = 3;  // "USER_REQUESTED", "TIMEOUT", "ERROR"

  string correlation_id = 4;
}

message CancelClaimResponse {
  Claim claim = 1;  // Status = CANCELLED

  string bacen_transaction_id = 2;
  google.protobuf.Timestamp cancelled_at = 3;
}
```

**Casos de Uso**:
- Usuário cancela reivindicação manualmente
- ExpireCompletionPeriodWorkflow cancela após 30 dias
- Erro no processamento

**Timeout**: 60s

---

### 4.7. HealthCheck

**Descrição**: Verificar saúde do Bridge e conectividade com Bacen.

```protobuf
message HealthCheckResponse {
  HealthStatus status = 1;

  // Detalhes
  BacenConnectionStatus bacen_status = 2;
  CertificateStatus certificate_status = 3;

  // Latências
  int64 bacen_latency_ms = 4;
  google.protobuf.Timestamp last_check = 5;
}

enum HealthStatus {
  HEALTH_STATUS_UNSPECIFIED = 0;
  HEALTH_STATUS_HEALTHY = 1;
  HEALTH_STATUS_DEGRADED = 2;
  HEALTH_STATUS_UNHEALTHY = 3;
}

enum BacenConnectionStatus {
  BACEN_CONNECTION_UNSPECIFIED = 0;
  BACEN_CONNECTION_OK = 1;
  BACEN_CONNECTION_TIMEOUT = 2;
  BACEN_CONNECTION_AUTH_FAILED = 3;
  BACEN_CONNECTION_TLS_ERROR = 4;
}

enum CertificateStatus {
  CERTIFICATE_STATUS_UNSPECIFIED = 0;
  CERTIFICATE_STATUS_VALID = 1;
  CERTIFICATE_STATUS_EXPIRING_SOON = 2;  // < 30 dias
  CERTIFICATE_STATUS_EXPIRED = 3;
}
```

**Casos de Uso**:
- Kubernetes liveness/readiness probes
- Monitoring e alertas
- Dashboard de saúde

**Timeout**: 5s

---

## 5. Error Handling

### 5.1. gRPC Status Codes

| gRPC Code | Situação | Retry? | Exemplo |
|-----------|----------|--------|---------|
| `OK` (0) | Sucesso | - | Entry criado |
| `INVALID_ARGUMENT` (3) | Request inválido | ❌ Não | CPF formato errado |
| `NOT_FOUND` (5) | Recurso não existe | ❌ Não | Entry não encontrado |
| `ALREADY_EXISTS` (6) | Duplicação | ❌ Não | Chave já existe |
| `PERMISSION_DENIED` (7) | Sem permissão | ❌ Não | ISPB não autorizado |
| `RESOURCE_EXHAUSTED` (8) | Rate limit | ✅ Sim | Muitas requests |
| `UNAUTHENTICATED` (16) | Falha autenticação | ❌ Não | TLS certificate inválido |
| `UNAVAILABLE` (14) | Serviço indisponível | ✅ Sim | Bacen timeout |
| `DEADLINE_EXCEEDED` (4) | Timeout | ✅ Sim | Operação > 60s |
| `INTERNAL` (13) | Erro interno | ✅ Sim | SOAP parsing error |

### 5.2. Error Details (google.rpc.Status)

```protobuf
import "google/rpc/status.proto";
import "google/rpc/error_details.proto";

// Exemplo de erro detalhado
google.rpc.Status {
  code: 3,  // INVALID_ARGUMENT
  message: "Invalid CPF format",
  details: [
    google.rpc.BadRequest {
      field_violations: [
        {
          field: "key_value",
          description: "CPF must have 11 digits"
        }
      ]
    }
  ]
}
```

### 5.3. Retry Policy

```yaml
# gRPC Retry Config (Connect side)
retry_policy:
  max_attempts: 3
  initial_backoff: 1s
  max_backoff: 30s
  backoff_multiplier: 2.0
  retryable_status_codes:
    - UNAVAILABLE
    - DEADLINE_EXCEEDED
    - RESOURCE_EXHAUSTED
    - INTERNAL
```

---

## 6. Security

### 6.1. Transport Security

```yaml
# TLS 1.2+ obrigatório
tls:
  min_version: TLS_1_2
  cipher_suites:
    - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
    - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256

# Mutual TLS (opcional entre Connect e Bridge)
mtls:
  enabled: false  # Connect → Bridge não precisa (mesma rede interna)
  # Bridge → Bacen SIM (ver SEC-001)
```

### 6.2. Authentication

```protobuf
// Metadata para autenticação (opcional)
metadata {
  "authorization": "Bearer <jwt-token>",
  "x-api-key": "<api-key>",
  "x-correlation-id": "<uuid>",
  "x-idempotency-key": "<uuid>"
}
```

---

## 7. Performance

### 7.1. Timeouts Recomendados

| RPC | Timeout | Justificativa |
|-----|---------|---------------|
| `GetEntry` | 15s | Query simples |
| `CreateEntry` | 30s | Criação + validação Bacen |
| `CreateClaim` | 60s | Processo complexo (30 dias) |
| `CompleteClaim` | 60s | Transferência de ownership |
| `HealthCheck` | 5s | Deve ser rápido |

### 7.2. Connection Pooling

```yaml
# gRPC Connection Pool (Connect side)
connection_pool:
  max_connections: 10
  max_idle_connections: 5
  idle_timeout: 90s
  max_lifetime: 3600s
```

### 7.3. Observability

```protobuf
// Metadata para tracing
metadata {
  "x-trace-id": "<trace-id>",
  "x-span-id": "<span-id>",
  "x-parent-span-id": "<parent-span-id>"
}
```

**Integração**: OpenTelemetry v1.38.0 (conforme TEC-003)

---

## Próximas Revisões

**Pendências**:
- [ ] Definir RPCs para VSYNC (quando implementado)
- [ ] Adicionar RPCs para OTP validation (quando implementado)
- [ ] Validar timeouts em ambiente real
- [ ] Implementar rate limiting no Bridge

---

**Referências**:
- [TEC-002 v3.1: Bridge Specification](../../11_Especificacoes_Tecnicas/TEC-002_Bridge_Specification.md)
- [TEC-003 v2.1: Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [GRPC-003: Proto Files Specification](GRPC-003_Proto_Files_Specification.md) (pendente)
- [SEC-001: mTLS Configuration](../../13_Seguranca/SEC-001_mTLS_Configuration.md) (pendente)
