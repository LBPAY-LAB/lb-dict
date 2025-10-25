# GRPC-002: Core DICT gRPC Service

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Versão**: 1.0
**Data**: 2025-10-25
**Status**: ✅ Especificação Completa (Futuro - Core DICT ainda não existe)
**Responsável**: ARCHITECT (AI Agent - Technical Architect)

---

## 📋 Resumo Executivo

Este documento especifica o **serviço gRPC Core DICT**, que será exposto para o FrontEnd (aplicativo mobile/web) permitindo que usuários gerenciem suas chaves PIX, visualizem status de claims, e realizem operações de portabilidade.

**Objetivo**: Definir contrato gRPC completo para comunicação FrontEnd ↔ Core DICT, separando responsabilidades (Core DICT = API pública, Connect = orquestração interna).

**Importante**: Core DICT ainda **NÃO foi implementado**. Este documento é uma especificação para implementação futura.

---

## 🎯 Contexto Arquitetural

### Fluxo de Comunicação

```
[FrontEnd Web/Mobile]
        │
        │ HTTPS/REST (público)
        ▼
[Core DICT API Gateway]
        │
        │ gRPC (interno)
        ▼
[RSFN Connect]
        │
        │ gRPC (interno)
        ▼
[Bridge DICT]
        │
        │ SOAP/mTLS (externo)
        ▼
[Bacen DICT]
```

### Responsabilidades

| Componente | Responsabilidade |
|------------|------------------|
| **FrontEnd** | UI/UX, validação de input, autenticação JWT |
| **Core DICT** | API REST pública, transformação REST ↔ gRPC, rate limiting, autorização |
| **Connect** | Orquestração, workflows Temporal, persistência PostgreSQL, eventos Pulsar |
| **Bridge** | Adaptação gRPC ↔ SOAP, mTLS, assinatura XML, comunicação com Bacen |

---

## 📄 Service Definition (gRPC)

### Arquivo: `proto/core/v1/core.proto`

```protobuf
syntax = "proto3";

package rsfn.core.v1;

option go_package = "github.com/lbpay/dict/proto/core/v1;corev1";

import "common/v1/types.proto";
import "common/v1/errors.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/empty.proto";

// ====================================================================
// CORE DICT SERVICE - FrontEnd → Core DICT
// ====================================================================
service CoreDictService {
  // ========== Key Operations (Chaves PIX) ==========

  // Criar nova chave PIX para usuário autenticado
  rpc CreateKey(CreateKeyRequest) returns (CreateKeyResponse);

  // Listar todas as chaves do usuário autenticado
  rpc ListKeys(ListKeysRequest) returns (ListKeysResponse);

  // Obter detalhes de uma chave específica
  rpc GetKey(GetKeyRequest) returns (GetKeyResponse);

  // Deletar chave PIX
  rpc DeleteKey(DeleteKeyRequest) returns (DeleteKeyResponse);

  // ========== Claim Operations (Reivindicações - 30 dias) ==========

  // Iniciar reivindicação de chave de outro usuário
  rpc StartClaim(StartClaimRequest) returns (StartClaimResponse);

  // Verificar status de uma claim
  rpc GetClaimStatus(GetClaimStatusRequest) returns (GetClaimStatusResponse);

  // Listar claims recebidas (onde sou o dono da chave)
  rpc ListIncomingClaims(ListIncomingClaimsRequest) returns (ListIncomingClaimsResponse);

  // Listar claims enviadas (onde sou o reivindicador)
  rpc ListOutgoingClaims(ListOutgoingClaimsRequest) returns (ListOutgoingClaimsResponse);

  // Responder a uma claim (aceitar ou rejeitar)
  rpc RespondToClaim(RespondToClaimRequest) returns (RespondToClaimResponse);

  // Cancelar claim enviada (apenas reivindicador)
  rpc CancelClaim(CancelClaimRequest) returns (CancelClaimResponse);

  // ========== Portability Operations (Portabilidade de Conta) ==========

  // Iniciar portabilidade de chave para nova conta
  rpc StartPortability(StartPortabilityRequest) returns (StartPortabilityResponse);

  // Confirmar portabilidade
  rpc ConfirmPortability(ConfirmPortabilityRequest) returns (ConfirmPortabilityResponse);

  // Cancelar portabilidade
  rpc CancelPortability(CancelPortabilityRequest) returns (CancelPortabilityResponse);

  // ========== Query Operations (Consultas) ==========

  // Consultar chave DICT de terceiros (para transações PIX)
  rpc LookupKey(LookupKeyRequest) returns (LookupKeyResponse);

  // ========== Health Check ==========

  // Health check do Core DICT
  rpc HealthCheck(google.protobuf.Empty) returns (HealthCheckResponse);
}

// ====================================================================
// KEY OPERATIONS - Messages
// ====================================================================

message CreateKeyRequest {
  // Tipo de chave PIX
  rsfn.common.v1.KeyType key_type = 1;

  // Valor da chave (opcional se EVP - será gerado)
  string key_value = 2;

  // Conta a vincular à chave
  string account_id = 3;  // ID da conta no sistema LBPay
}

message CreateKeyResponse {
  // ID da key criada
  string key_id = 1;

  // Tipo e valor da chave
  rsfn.common.v1.DictKey key = 2;

  // Status (sempre ACTIVE ao criar)
  rsfn.common.v1.EntryStatus status = 3;

  // Timestamp de criação
  google.protobuf.Timestamp created_at = 4;
}

message ListKeysRequest {
  // Paginação
  int32 page_size = 1;  // Default: 20, Max: 100
  string page_token = 2;

  // Filtros opcionais
  optional rsfn.common.v1.KeyType key_type = 3;
  optional rsfn.common.v1.EntryStatus status = 4;
}

message ListKeysResponse {
  // Lista de chaves do usuário
  repeated KeySummary keys = 1;

  // Token para próxima página (vazio se última página)
  string next_page_token = 2;

  // Total de keys do usuário
  int32 total_count = 3;
}

message KeySummary {
  string key_id = 1;
  rsfn.common.v1.DictKey key = 2;
  rsfn.common.v1.EntryStatus status = 3;
  string account_id = 4;
  google.protobuf.Timestamp created_at = 5;
  google.protobuf.Timestamp updated_at = 6;
}

message GetKeyRequest {
  // ID da key ou valor da key
  oneof identifier {
    string key_id = 1;
    rsfn.common.v1.DictKey key = 2;
  }
}

message GetKeyResponse {
  string key_id = 1;
  rsfn.common.v1.DictKey key = 2;
  rsfn.common.v1.Account account = 3;
  rsfn.common.v1.EntryStatus status = 4;
  google.protobuf.Timestamp created_at = 5;
  google.protobuf.Timestamp updated_at = 6;

  // Histórico de portabilidades (se houver)
  repeated PortabilityHistory portability_history = 7;
}

message PortabilityHistory {
  string portability_id = 1;
  rsfn.common.v1.Account old_account = 2;
  rsfn.common.v1.Account new_account = 3;
  google.protobuf.Timestamp confirmed_at = 4;
}

message DeleteKeyRequest {
  string key_id = 1;
}

message DeleteKeyResponse {
  bool deleted = 1;
  google.protobuf.Timestamp deleted_at = 2;
}

// ====================================================================
// CLAIM OPERATIONS - Messages (30 dias)
// ====================================================================

message StartClaimRequest {
  // Chave a ser reivindicada
  rsfn.common.v1.DictKey key = 1;

  // Conta do reivindicador (destino)
  string account_id = 2;  // ID da conta no sistema LBPay
}

message StartClaimResponse {
  // ID da claim criada
  string claim_id = 1;

  // ID da entry reivindicada
  string entry_id = 2;

  // Status inicial (sempre OPEN)
  rsfn.common.v1.ClaimStatus status = 3;

  // Data de expiração (created_at + 30 dias)
  google.protobuf.Timestamp expires_at = 4;

  // Timestamp de criação
  google.protobuf.Timestamp created_at = 5;

  // Mensagem para o usuário
  string message = 6;  // "Claim criada. O dono tem 30 dias para responder"
}

message GetClaimStatusRequest {
  string claim_id = 1;
}

message GetClaimStatusResponse {
  string claim_id = 1;
  string entry_id = 2;
  rsfn.common.v1.DictKey key = 3;

  // Status atual
  rsfn.common.v1.ClaimStatus status = 4;

  // ISPBs envolvidos
  string claimer_ispb = 5;  // Quem está reivindicando
  string owner_ispb = 6;    // Dono atual

  // Timestamps
  google.protobuf.Timestamp created_at = 7;
  google.protobuf.Timestamp expires_at = 8;
  optional google.protobuf.Timestamp completed_at = 9;

  // Tempo restante (para exibir no frontend)
  int32 days_remaining = 10;
}

message ListIncomingClaimsRequest {
  // Filtros
  optional rsfn.common.v1.ClaimStatus status = 1;

  // Paginação
  int32 page_size = 2;
  string page_token = 3;
}

message ListIncomingClaimsResponse {
  repeated ClaimSummary claims = 1;
  string next_page_token = 2;
  int32 total_count = 3;
}

message ListOutgoingClaimsRequest {
  // Filtros
  optional rsfn.common.v1.ClaimStatus status = 1;

  // Paginação
  int32 page_size = 2;
  string page_token = 3;
}

message ListOutgoingClaimsResponse {
  repeated ClaimSummary claims = 1;
  string next_page_token = 2;
  int32 total_count = 3;
}

message ClaimSummary {
  string claim_id = 1;
  string entry_id = 2;
  rsfn.common.v1.DictKey key = 3;
  rsfn.common.v1.ClaimStatus status = 4;
  google.protobuf.Timestamp created_at = 5;
  google.protobuf.Timestamp expires_at = 6;
  int32 days_remaining = 7;
}

message RespondToClaimRequest {
  string claim_id = 1;

  // Resposta (aceitar ou rejeitar)
  enum ClaimResponse {
    CLAIM_RESPONSE_UNSPECIFIED = 0;
    CLAIM_RESPONSE_ACCEPT = 1;
    CLAIM_RESPONSE_REJECT = 2;
  }
  ClaimResponse response = 2;

  // Razão (opcional, para rejeição)
  optional string reason = 3;
}

message RespondToClaimResponse {
  string claim_id = 1;
  rsfn.common.v1.ClaimStatus new_status = 2;  // CONFIRMED ou CANCELLED
  google.protobuf.Timestamp responded_at = 3;
  string message = 4;  // "Claim aceita com sucesso" ou "Claim rejeitada"
}

message CancelClaimRequest {
  string claim_id = 1;
  optional string reason = 2;
}

message CancelClaimResponse {
  string claim_id = 1;
  rsfn.common.v1.ClaimStatus status = 2;  // CANCELLED
  google.protobuf.Timestamp cancelled_at = 3;
}

// ====================================================================
// PORTABILITY OPERATIONS - Messages
// ====================================================================

message StartPortabilityRequest {
  // Chave a sofrer portabilidade
  string key_id = 1;

  // Nova conta de destino
  string new_account_id = 2;
}

message StartPortabilityResponse {
  string portability_id = 1;
  string key_id = 2;
  rsfn.common.v1.Account new_account = 3;
  google.protobuf.Timestamp started_at = 4;
  string message = 5;  // "Portabilidade iniciada. Aguarde confirmação"
}

message ConfirmPortabilityRequest {
  string portability_id = 1;
}

message ConfirmPortabilityResponse {
  string portability_id = 1;
  string key_id = 2;
  rsfn.common.v1.EntryStatus status = 3;  // ACTIVE (com nova conta)
  google.protobuf.Timestamp confirmed_at = 4;
}

message CancelPortabilityRequest {
  string portability_id = 1;
  optional string reason = 2;
}

message CancelPortabilityResponse {
  string portability_id = 1;
  google.protobuf.Timestamp cancelled_at = 2;
}

// ====================================================================
// QUERY OPERATIONS - Messages
// ====================================================================

message LookupKeyRequest {
  // Chave a consultar
  rsfn.common.v1.DictKey key = 1;
}

message LookupKeyResponse {
  // Dados públicos da chave (para transação PIX)
  rsfn.common.v1.DictKey key = 1;
  rsfn.common.v1.Account account = 2;  // Apenas dados públicos (ISPB, agência, conta)
  string account_holder_name = 3;

  // Status (se ACTIVE, pode receber PIX)
  rsfn.common.v1.EntryStatus status = 4;
}

// ====================================================================
// HEALTH CHECK
// ====================================================================

message HealthCheckResponse {
  enum HealthStatus {
    HEALTH_STATUS_UNSPECIFIED = 0;
    HEALTH_STATUS_HEALTHY = 1;
    HEALTH_STATUS_DEGRADED = 2;
    HEALTH_STATUS_UNHEALTHY = 3;
  }
  HealthStatus status = 1;

  // Conectividade com Connect (gRPC)
  bool connect_reachable = 2;

  // Timestamp do health check
  google.protobuf.Timestamp checked_at = 3;
}
```

---

## 🔐 Autenticação e Autorização

### Autenticação: JWT Bearer Token

**FrontEnd** envia token JWT no metadata gRPC:
```
Authorization: Bearer <jwt-token>
```

**Core DICT** valida token e extrai user_id:
```go
// Pseudocódigo
func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    // Extrair token do metadata
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "missing metadata")
    }

    authHeader := md.Get("authorization")
    if len(authHeader) == 0 {
        return nil, status.Error(codes.Unauthenticated, "missing authorization header")
    }

    // Validar JWT
    token := strings.TrimPrefix(authHeader[0], "Bearer ")
    claims, err := validateJWT(token)
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, "invalid token")
    }

    // Adicionar user_id ao contexto
    ctx = context.WithValue(ctx, "user_id", claims.UserID)

    return handler(ctx, req)
}
```

### Autorização: RBAC (Role-Based Access Control)

**Roles**:
- `user`: Usuário normal (pode gerenciar apenas suas próprias chaves)
- `admin`: Administrador (pode ver todas as chaves, mas não modificar)
- `support`: Suporte (pode ver chaves de clientes para troubleshooting)

**Regras**:
- `CreateKey`, `DeleteKey`: Apenas `user` (próprio usuário)
- `ListKeys`: `user` (próprias chaves), `admin`/`support` (todas)
- `RespondToClaim`: Apenas `user` (dono da chave)
- `LookupKey`: Qualquer usuário autenticado (dados públicos)

---

## 🔄 Mapeamento REST → gRPC

Core DICT expõe REST API para FrontEnd, mas internamente chama Connect via gRPC.

### Exemplo: POST /api/v1/keys

**REST Request** (FrontEnd → Core DICT):
```http
POST /api/v1/keys
Authorization: Bearer <jwt>
Content-Type: application/json

{
  "keyType": "CPF",
  "keyValue": "12345678900",
  "accountId": "acc-550e8400"
}
```

**gRPC Call** (Core DICT → Connect):
```go
response, err := coreClient.CreateKey(ctx, &corev1.CreateKeyRequest{
    KeyType:   pb.KeyType_KEY_TYPE_CPF,
    KeyValue:  "12345678900",
    AccountId: "acc-550e8400",
})
```

**REST Response** (Core DICT → FrontEnd):
```http
HTTP/1.1 201 Created
Content-Type: application/json

{
  "keyId": "key-550e8400",
  "key": {
    "keyType": "CPF",
    "keyValue": "12345678900"
  },
  "status": "ACTIVE",
  "createdAt": "2025-10-25T10:00:00Z"
}
```

---

## 📊 Rate Limiting

**Limites por Usuário** (para evitar abuso):

| Operação | Limite | Janela |
|----------|--------|--------|
| CreateKey | 5 keys/dia | 24 horas |
| DeleteKey | 10 keys/dia | 24 horas |
| StartClaim | 3 claims/hora | 1 hora |
| LookupKey | 100 lookups/min | 1 minuto |

**Implementação**:
- Usar Redis para counters
- Retornar `codes.ResourceExhausted` quando limite excedido
- Incluir `RetryInfo` com tempo de espera

---

## 📋 Checklist de Implementação

**Importante**: Core DICT **NÃO existe ainda**. Esta checklist é para implementação futura.

- [ ] Criar repo `dict-core` (novo repositório)
- [ ] Implementar serviço gRPC `CoreDictService`
- [ ] Implementar API Gateway REST (Gin, Echo, ou Chi)
- [ ] Configurar autenticação JWT
- [ ] Implementar autorização RBAC
- [ ] Configurar rate limiting (Redis)
- [ ] Criar gRPC client para Connect
- [ ] Implementar mapeamento REST ↔ gRPC
- [ ] Adicionar logging e métricas (Prometheus)
- [ ] Criar testes unitários e E2E
- [ ] Documentar API REST (Swagger/OpenAPI)
- [ ] Deploy em Kubernetes (staging + prod)

---

## 📚 Referências

### Documentos Internos
- [GRPC-001: Bridge gRPC Service](GRPC-001_Bridge_gRPC_Service.md) - Contrato Connect ↔ Bridge
- [GRPC-003: Proto Files Specification](GRPC-003_Proto_Files_Specification.md) - Common types
- [GRPC-004: Error Handling gRPC](GRPC-004_Error_Handling_gRPC.md) - Error handling
- [TEC-001: Core DICT Specification](../../11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md) - (quando criado)
- [SEC-004: API Authentication](../../13_Seguranca/SEC-004_API_Authentication.md)

### Documentação Externa
- [gRPC Gateway](https://github.com/grpc-ecosystem/grpc-gateway) - REST → gRPC proxy
- [JWT Authentication](https://jwt.io/introduction)
- [Google API Design Guide](https://cloud.google.com/apis/design)

---

**Versão**: 1.0
**Status**: ✅ Especificação Completa (Aguardando criação do Core DICT)
**Próxima Revisão**: Quando Core DICT for implementado

---

**IMPORTANTE**: Este documento especifica o **Core DICT que ainda não existe**. A implementação do Core DICT é uma **Fase futura** (provavelmente Fase 3+), após Connect e Bridge estarem prontos e testados.
