# Core-Dict → Conn-Dict: Guia de Integração gRPC

**Data**: 2025-10-27
**Versão dict-contracts**: v0.2.0
**Status**: Production-Ready
**Maintainer**: DICT Implementation Squad

---

## Índice

1. [Overview](#1-overview)
2. [Setup](#2-setup)
3. [Entry Operations (Read-Only)](#3-entry-operations-read-only)
4. [Claim Operations](#4-claim-operations)
5. [Infraction Operations](#5-infraction-operations)
6. [Health Check](#6-health-check)
7. [Error Handling](#7-error-handling)
8. [Timeouts e Retry](#8-timeouts-e-retry)
9. [Observability](#9-observability)
10. [Testing](#10-testing)
11. [Production Checklist](#11-production-checklist)
12. [Configuration](#12-configuration)
13. [Complete Production Example](#13-complete-production-example)
14. [Quick Reference](#14-quick-reference)

---

## 1. Overview

Este guia documenta **EXATAMENTE** como `core-dict` deve chamar `conn-dict` via gRPC.

**Arquitetura**:
```
┌───────────┐                      ┌────────────┐
│ core-dict │ ──── gRPC (9092) ──> │ conn-dict  │
│ (API)     │                      │ (Connect)  │
└───────────┘                      └────────────┘
```

**ConnectService** expõe **15 RPCs**:
- **3 Entry RPCs**: GetEntry, GetEntryByKey, ListEntries (read-only queries)
- **5 Claim RPCs**: CreateClaim, ConfirmClaim, CancelClaim, GetClaim, ListClaims
- **6 Infraction RPCs**: CreateInfraction, InvestigateInfraction, ResolveInfraction, DismissInfraction, GetInfraction, ListInfractions
- **1 Health RPC**: HealthCheck

**Quando usar gRPC vs Pulsar**:
- **gRPC (síncrono)**: Queries (GetEntry, ListClaims), workflow triggers (CreateClaim, ConfirmClaim)
- **Pulsar (assíncrono)**: Entry mutations (create, update, delete via eventos)

---

## 2. Setup

### 2.1 Dependências (go.mod)

```go
module github.com/lbpay-lab/core-dict

go 1.24.5

require (
    github.com/lbpay-lab/dict-contracts v0.2.0
    google.golang.org/grpc v1.58.3
    google.golang.org/protobuf v1.31.0
    go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.46.0
    github.com/prometheus/client_golang v1.18.0
)

// Development: local replace (remover em produção)
replace github.com/lbpay-lab/dict-contracts => ../dict-contracts
```

### 2.2 Imports Necessários

```go
import (
    "context"
    "fmt"
    "time"

    // gRPC Core
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/keepalive"

    // Proto Contracts
    connectv1 "github.com/lbpay-lab/dict-contracts/gen/proto/conn_dict/v1"
    commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
    "google.golang.org/protobuf/types/known/emptypb"
    "google.golang.org/protobuf/types/known/timestamppb"

    // Observability
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
    "github.com/prometheus/client_golang/prometheus"
)
```

### 2.3 Criar Cliente gRPC (Desenvolvimento)

```go
package infrastructure

import (
    "context"
    "fmt"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    connectv1 "github.com/lbpay-lab/dict-contracts/gen/proto/conn_dict/v1"
)

// ConnectClient é o cliente gRPC para conn-dict
type ConnectClient struct {
    conn   *grpc.ClientConn
    client connectv1.ConnectServiceClient
}

// NewConnectClient cria um novo cliente conn-dict (dev)
func NewConnectClient(address string) (*ConnectClient, error) {
    conn, err := grpc.Dial(
        address,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithDefaultCallOptions(
            grpc.MaxCallRecvMsgSize(10*1024*1024), // 10MB
            grpc.MaxCallSendMsgSize(10*1024*1024), // 10MB
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to dial conn-dict: %w", err)
    }

    client := connectv1.NewConnectServiceClient(conn)

    return &ConnectClient{
        conn:   conn,
        client: client,
    }, nil
}

// Close fecha a conexão gRPC
func (c *ConnectClient) Close() error {
    return c.conn.Close()
}

// Client retorna o cliente gRPC
func (c *ConnectClient) Client() connectv1.ConnectServiceClient {
    return c.client
}
```

**Uso**:
```go
// Criar cliente
client, err := NewConnectClient("localhost:9092")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Usar client
ctx := context.Background()
resp, err := client.Client().GetEntry(ctx, &connectv1.GetEntryRequest{
    EntryId: "entry-uuid-123",
})
```

---

## 3. Entry Operations (Read-Only)

Entry mutations (create, update, delete) são feitas via **Pulsar events**, não gRPC.
Use gRPC apenas para **queries**.

### 3.1 GetEntry - Buscar por ID

Retorna uma entry específica por ID.

**Request**:
```go
resp, err := client.GetEntry(ctx, &connectv1.GetEntryRequest{
    EntryId:   "entry-uuid-123",
    RequestId: "req-456", // Opcional (tracing)
})
if err != nil {
    st, ok := status.FromError(err)
    if ok && st.Code() == codes.NotFound {
        // Entry não existe
        return nil, ErrEntryNotFound
    }
    return nil, err
}

// Verificar se foi encontrado
if !resp.Found {
    return nil, ErrEntryNotFound
}

entry := resp.Entry
fmt.Printf("Entry: %s, Status: %s\n", entry.KeyValue, entry.Status)
```

**Response Fields**:
```go
type Entry struct {
    EntryId         string    // UUID interno
    ParticipantIspb string    // ISPB do participante (8 dígitos)
    KeyType         KeyType   // KEY_TYPE_CPF, KEY_TYPE_EMAIL, etc.
    KeyValue        string    // Valor da chave
    Account         *Account  // Conta vinculada
    Status          EntryStatus // ENTRY_STATUS_ACTIVE, etc.
    CreatedAt       *timestamppb.Timestamp
    UpdatedAt       *timestamppb.Timestamp
    DeletedAt       *timestamppb.Timestamp // Opcional (soft delete)
}
```

**Error Codes**:
- `INVALID_ARGUMENT`: `entry_id` vazio
- `NOT_FOUND`: Entry não existe
- `INTERNAL`: Erro database

**Latência Esperada**: p50 < 10ms, p99 < 50ms

---

### 3.2 GetEntryByKey - Buscar por Chave DICT

Busca entry por chave DICT (CPF, email, phone, etc.).

**Request**:
```go
resp, err := client.GetEntryByKey(ctx, &connectv1.GetEntryByKeyRequest{
    Key: &commonv1.DictKey{
        KeyType:  commonv1.KeyType_KEY_TYPE_EMAIL,
        KeyValue: "user@example.com",
    },
    RequestId: "req-789",
})
if err != nil {
    return nil, err
}

if !resp.Found {
    return nil, ErrEntryNotFound
}

entry := resp.Entry
```

**Key Types**:
```go
commonv1.KeyType_KEY_TYPE_CPF      // CPF (11 dígitos): "12345678900"
commonv1.KeyType_KEY_TYPE_CNPJ     // CNPJ (14 dígitos): "12345678000199"
commonv1.KeyType_KEY_TYPE_EMAIL    // Email: "user@example.com"
commonv1.KeyType_KEY_TYPE_PHONE    // Phone: "+5511987654321"
commonv1.KeyType_KEY_TYPE_EVP      // Chave aleatória (UUID): "123e4567-e89b-..."
```

**Error Codes**:
- `INVALID_ARGUMENT`: `key` vazio ou tipo inválido
- `NOT_FOUND`: Entry não existe
- `INTERNAL`: Erro database

**Latência Esperada**: p50 < 10ms, p99 < 50ms

---

### 3.3 ListEntries - Listar por Participante

Lista entries de um participante (ISPB) com paginação.

**Request**:
```go
resp, err := client.ListEntries(ctx, &connectv1.ListEntriesRequest{
    ParticipantIspb: "12345678", // Required (8 dígitos)
    KeyType:         &keyType,   // Optional filter
    Status:          &status,    // Optional filter
    Limit:           100,        // Default: 100, Max: 1000
    Offset:          0,          // Default: 0
    RequestId:       "req-101",
})
if err != nil {
    return nil, err
}

for _, entry := range resp.Entries {
    fmt.Printf("Key: %s, Status: %s\n", entry.KeyValue, entry.Status)
}

fmt.Printf("Total: %d, HasMore: %v\n", resp.TotalCount, resp.HasMore)
```

**Paginação**:
```go
// Buscar página 1 (primeiras 100 entries)
resp1, _ := client.ListEntries(ctx, &connectv1.ListEntriesRequest{
    ParticipantIspb: "12345678",
    Limit:           100,
    Offset:          0,
})

// Buscar página 2 (próximas 100 entries)
if resp1.HasMore {
    resp2, _ := client.ListEntries(ctx, &connectv1.ListEntriesRequest{
        ParticipantIspb: "12345678",
        Limit:           100,
        Offset:          100,
    })
}
```

**Filtros Opcionais**:
```go
// Filtrar por tipo de chave
keyType := commonv1.KeyType_KEY_TYPE_CPF
resp, _ := client.ListEntries(ctx, &connectv1.ListEntriesRequest{
    ParticipantIspb: "12345678",
    KeyType:         &keyType, // Apenas CPFs
})

// Filtrar por status
status := commonv1.EntryStatus_ENTRY_STATUS_ACTIVE
resp, _ := client.ListEntries(ctx, &connectv1.ListEntriesRequest{
    ParticipantIspb: "12345678",
    Status:          &status, // Apenas ativos
})
```

**Error Codes**:
- `INVALID_ARGUMENT`: `participant_ispb` inválido ou vazio
- `INTERNAL`: Erro database

**Latência Esperada**: p50 < 50ms, p99 < 200ms

**Segurança**: Usuário só pode listar entries do próprio ISPB (enforced pelo middleware de autenticação).

---

## 4. Claim Operations

Claims gerenciam portabilidade/posse de chaves PIX com workflow de 30 dias.

### 4.1 CreateClaim - Iniciar Reivindicação

Inicia um **ClaimWorkflow** (30 dias) no Temporal.

**Request**:
```go
resp, err := client.CreateClaim(ctx, &connectv1.CreateClaimRequest{
    EntryId:     "entry-uuid-123", // Required
    ClaimerIspb: "87654321",       // Required (8 dígitos)
    OwnerIspb:   "12345678",       // Required (8 dígitos, diferente de claimer)
    ClaimerAccount: &commonv1.Account{
        Ispb:                  "87654321",
        AccountType:           commonv1.AccountType_ACCOUNT_TYPE_CHECKING,
        AccountNumber:         "123456",
        AccountCheckDigit:     "7",
        BranchCode:            "0001",
        AccountHolderName:     "Jane Doe",
        AccountHolderDocument: "98765432100",
        DocumentType:          commonv1.DocumentType_DOCUMENT_TYPE_CPF,
    },
    ClaimType: connectv1.CreateClaimRequest_CLAIM_TYPE_PORTABILITY,
    RequestId: "req-202",
})
if err != nil {
    return nil, err
}

fmt.Printf("Claim ID: %s\n", resp.ClaimId)
fmt.Printf("Status: %s\n", resp.Status)           // Sempre "OPEN"
fmt.Printf("Expires At: %s\n", resp.ExpiresAt)    // created_at + 30 dias
fmt.Printf("Message: %s\n", resp.Message)
```

**Claim Types**:
```go
connectv1.CreateClaimRequest_CLAIM_TYPE_OWNERSHIP     // Reivindicação de posse
connectv1.CreateClaimRequest_CLAIM_TYPE_PORTABILITY  // Portabilidade de conta
```

**Workflow Behavior** (30 dias):
```
Day 0:  CreateClaim → Status: OPEN → Owner notificado
Day 0-7: Owner pode confirmar/rejeitar
Day 7:  (se sem resposta) → Auto-approve → Status: CONFIRMED
Day 30: (se ainda aberto) → Auto-expire → Status: EXPIRED
```

**Error Codes**:
- `INVALID_ARGUMENT`: Campo obrigatório faltando, ISPB inválido, claimer == owner
- `ALREADY_EXISTS`: Claim ativa já existe para esta entry
- `NOT_FOUND`: Entry não existe
- `INTERNAL`: Temporal workflow start falhou

**Latência Esperada**: p50 < 100ms, p99 < 500ms (overhead de start workflow)

---

### 4.2 ConfirmClaim - Owner Aceita

Owner confirma a claim → transfere posse da chave ao claimer.
Envia um **Signal** ao ClaimWorkflow.

**Request**:
```go
resp, err := client.ConfirmClaim(ctx, &connectv1.ConfirmClaimRequest{
    ClaimId:   "claim-uuid-456", // Required
    Reason:    "Customer request", // Optional
    RequestId: "req-303",
})
if err != nil {
    st, ok := status.FromError(err)
    if ok && st.Code() == codes.FailedPrecondition {
        // Claim não pode ser confirmada (já expirou, já foi cancelada, etc.)
        return nil, ErrClaimNotConfirmable
    }
    return nil, err
}

fmt.Printf("Status: %s\n", resp.Status)           // CONFIRMED
fmt.Printf("Entry: %v\n", resp.UpdatedEntry)       // Entry agora pertence ao claimer
fmt.Printf("Confirmed At: %s\n", resp.ConfirmedAt)
```

**Side Effects**:
1. ClaimWorkflow recebe signal "confirm"
2. Workflow executa `TransferKeyOwnershipActivity`
3. Entry ownership transferida (participant_ispb atualizado)
4. Claim status → CONFIRMED
5. Evento `dict.claims.completed` publicado no Pulsar

**Error Codes**:
- `INVALID_ARGUMENT`: `claim_id` vazio
- `NOT_FOUND`: Claim não existe
- `FAILED_PRECONDITION`: Claim não pode ser confirmada (já expirou, já cancelada, já confirmada)
- `INTERNAL`: Temporal signal falhou

**Latência Esperada**: p50 < 50ms, p99 < 200ms

---

### 4.3 CancelClaim - Cancelar Reivindicação

Claimer ou owner cancela a claim. Envia **Signal** ao workflow.

**Request**:
```go
resp, err := client.CancelClaim(ctx, &connectv1.CancelClaimRequest{
    ClaimId:   "claim-uuid-456", // Required
    Reason:    "Customer changed mind", // Required
    RequestId: "req-404",
})
if err != nil {
    return nil, err
}

fmt.Printf("Status: %s\n", resp.Status)      // CANCELLED
fmt.Printf("Reason: %s\n", resp.Message)
```

**Error Codes**:
- `INVALID_ARGUMENT`: `claim_id` ou `reason` vazio
- `NOT_FOUND`: Claim não existe
- `FAILED_PRECONDITION`: Claim não pode ser cancelada (já completa)
- `INTERNAL`: Temporal signal falhou

**Latência Esperada**: p50 < 50ms, p99 < 200ms

---

### 4.4 GetClaim - Buscar por ID

Retorna claim específica por ID.

**Request**:
```go
resp, err := client.GetClaim(ctx, &connectv1.GetClaimRequest{
    ClaimId:   "claim-uuid-456",
    RequestId: "req-505",
})
if err != nil {
    return nil, err
}

if !resp.Found {
    return nil, ErrClaimNotFound
}

claim := resp.Claim
fmt.Printf("Key: %s\n", claim.KeyValue)
fmt.Printf("Status: %s\n", claim.Status)
fmt.Printf("Days Remaining: %d\n", claim.DaysRemaining)
```

**Response Fields**:
```go
type Claim struct {
    ClaimId      string    // UUID
    EntryId      string    // Entry reivindicada
    KeyType      KeyType   // Tipo da chave
    KeyValue     string    // Valor da chave
    ClaimerIspb  string    // Quem está reivindicando
    OwnerIspb    string    // Dono atual
    ClaimType    ClaimType // OWNERSHIP ou PORTABILITY
    Status       ClaimStatus // OPEN, CONFIRMED, CANCELLED, EXPIRED
    CreatedAt    *timestamppb.Timestamp
    ExpiresAt    *timestamppb.Timestamp // created_at + 30 dias
    CompletedAt  *timestamppb.Timestamp // Opcional
    DaysRemaining int32    // Dias restantes (0-30)
}
```

**Error Codes**:
- `INVALID_ARGUMENT`: `claim_id` vazio
- `NOT_FOUND`: Claim não existe
- `INTERNAL`: Erro database

---

### 4.5 ListClaims - Listar com Filtros

Lista claims com filtros opcionais e paginação.

**Request**:
```go
// Listar claims por entry
resp, _ := client.ListClaims(ctx, &connectv1.ListClaimsRequest{
    EntryId: &entryID,
    Limit:   50,
    Offset:  0,
})

// Listar claims por claimer
resp, _ := client.ListClaims(ctx, &connectv1.ListClaimsRequest{
    ClaimerIspb: &claimerIspb,
    Status:      &status,
    Limit:       50,
})

// Listar claims por owner
resp, _ := client.ListClaims(ctx, &connectv1.ListClaimsRequest{
    OwnerIspb: &ownerIspb,
    Limit:     50,
})
```

**Filtros Disponíveis**:
- `entry_id`: Filtrar por entry
- `claimer_ispb`: Filtrar por claimer (8 dígitos)
- `owner_ispb`: Filtrar por owner (8 dígitos)
- `status`: Filtrar por status (OPEN, CONFIRMED, CANCELLED, EXPIRED)

**Paginação**: Igual a `ListEntries` (limit, offset, has_more)

**Error Codes**:
- `INVALID_ARGUMENT`: ISPB inválido
- `INTERNAL`: Erro database

---

## 5. Infraction Operations

Infrações gerenciam fraudes e violações com workflow de investigação.

### 5.1 CreateInfraction - Reportar Infração

Inicia um **InvestigateInfractionWorkflow** (human-in-the-loop).

**Request**:
```go
resp, err := client.CreateInfraction(ctx, &connectv1.CreateInfractionRequest{
    Key: &commonv1.DictKey{
        KeyType:  commonv1.KeyType_KEY_TYPE_PHONE,
        KeyValue: "+5511999999999",
    },
    ParticipantIspb: "12345678",      // ISPB sob suspeita
    InfractionType:  connectv1.CreateInfractionRequest_INFRACTION_TYPE_FRAUD,
    Description:     "Suspected fraudulent activity detected",
    ReporterIspb:    "87654321",      // Quem reportou
    RequestId:       "req-606",
})
if err != nil {
    return nil, err
}

fmt.Printf("Infraction ID: %s\n", resp.InfractionId)
fmt.Printf("Status: %s\n", resp.Status) // Sempre "REPORTED"
```

**Infraction Types**:
```go
connectv1.CreateInfractionRequest_INFRACTION_TYPE_FRAUD                // Fraude
connectv1.CreateInfractionRequest_INFRACTION_TYPE_ACCOUNT_CLOSED       // Conta encerrada
connectv1.CreateInfractionRequest_INFRACTION_TYPE_INVALID_ACCOUNT      // Conta inválida
connectv1.CreateInfractionRequest_INFRACTION_TYPE_DUPLICATE_KEY        // Chave duplicada
connectv1.CreateInfractionRequest_INFRACTION_TYPE_INCORRECT_OWNERSHIP  // Titularidade incorreta
```

**Workflow Behavior**:
```
1. CreateInfraction → Status: REPORTED
2. Analista investiga (human decision)
3. InvestigateInfraction signal → Status: UNDER_INVESTIGATION
4. Decision:
   - RESOLVE → Status: RESOLVED (ação corretiva aplicada)
   - DISMISS → Status: DISMISSED (sem fundamento)
```

**Error Codes**:
- `INVALID_ARGUMENT`: Campo obrigatório faltando, tipo inválido
- `INTERNAL`: Temporal workflow start falhou

---

### 5.2 InvestigateInfraction - Analista Decide

Analista submete decisão de investigação.

**Request**:
```go
resp, err := client.InvestigateInfraction(ctx, &connectv1.InvestigateInfractionRequest{
    InfractionId: "infr-uuid-789",
    Decision:     connectv1.InvestigateInfractionRequest_INVESTIGATION_DECISION_PROCEED,
    AnalystNotes: "Evidence confirmed fraud. Proceeding with resolution.",
    RequestId:    "req-707",
})
if err != nil {
    return nil, err
}

fmt.Printf("Status: %s\n", resp.Status)
```

**Investigation Decisions**:
```go
connectv1.InvestigateInfractionRequest_INVESTIGATION_DECISION_PROCEED  // Proceder com resolução
connectv1.InvestigateInfractionRequest_INVESTIGATION_DECISION_DISMISS  // Descartar (sem fundamento)
```

**Error Codes**:
- `INVALID_ARGUMENT`: `infraction_id`, `decision`, ou `analyst_notes` faltando
- `NOT_FOUND`: Infraction não existe
- `INTERNAL`: Temporal signal falhou

---

### 5.3 ResolveInfraction - Resolver

Marca infração como resolvida (ação corretiva aplicada).

**Request**:
```go
resp, err := client.ResolveInfraction(ctx, &connectv1.ResolveInfractionRequest{
    InfractionId: "infr-uuid-789",
    Resolution:   "Key deleted and participant notified",
    RequestId:    "req-808",
})
if err != nil {
    return nil, err
}

fmt.Printf("Status: %s\n", resp.Status) // RESOLVED
```

**Error Codes**:
- `INVALID_ARGUMENT`: `infraction_id` ou `resolution` faltando
- `NOT_FOUND`: Infraction não existe
- `INTERNAL`: Erro

---

### 5.4 DismissInfraction - Descartar

Marca infração como sem fundamento.

**Request**:
```go
resp, err := client.DismissInfraction(ctx, &connectv1.DismissInfractionRequest{
    InfractionId: "infr-uuid-789",
    Reason:       "Investigation showed no fraud",
    RequestId:    "req-909",
})
if err != nil {
    return nil, err
}

fmt.Printf("Status: %s\n", resp.Status) // DISMISSED
```

---

### 5.5 GetInfraction - Buscar por ID

**Request**:
```go
resp, err := client.GetInfraction(ctx, &connectv1.GetInfractionRequest{
    InfractionId: "infr-uuid-789",
    RequestId:    "req-1010",
})
if err != nil {
    return nil, err
}

if !resp.Found {
    return nil, ErrInfractionNotFound
}

infraction := resp.Infraction
fmt.Printf("Type: %s, Status: %s\n", infraction.InfractionType, infraction.Status)
```

---

### 5.6 ListInfractions - Listar com Filtros

**Request**:
```go
// Listar infrações abertas por participante
resp, _ := client.ListInfractions(ctx, &connectv1.ListInfractionsRequest{
    ParticipantIspb: &ispb,
    Status:          &status,
    Limit:           50,
    Offset:          0,
})

for _, infr := range resp.Infractions {
    fmt.Printf("ID: %s, Type: %s\n", infr.InfractionId, infr.InfractionType)
}
```

**Filtros**:
- `participant_ispb`: Filtrar por participante sob investigação
- `status`: Filtrar por status

---

## 6. Health Check

Verifica saúde do conn-dict e dependências.

**Request**:
```go
resp, err := client.HealthCheck(ctx, &emptypb.Empty{})
if err != nil {
    return nil, err
}

fmt.Printf("Status: %s\n", resp.Status)
fmt.Printf("PostgreSQL: %v (latency: %dms)\n", resp.Postgresql.Reachable, resp.Postgresql.LatencyMs)
fmt.Printf("Redis: %v (latency: %dms)\n", resp.Redis.Reachable, resp.Redis.LatencyMs)
fmt.Printf("Temporal: %v (latency: %dms)\n", resp.Temporal.Reachable, resp.Temporal.LatencyMs)
fmt.Printf("Pulsar: %v (latency: %dms)\n", resp.Pulsar.Reachable, resp.Pulsar.LatencyMs)
fmt.Printf("Bridge: %v (latency: %dms)\n", resp.Bridge.Reachable, resp.Bridge.LatencyMs)
```

**Health Status**:
```go
connectv1.HealthCheckResponse_HEALTH_STATUS_HEALTHY    // Tudo OK
connectv1.HealthCheckResponse_HEALTH_STATUS_DEGRADED   // Algumas deps down
connectv1.HealthCheckResponse_HEALTH_STATUS_UNHEALTHY  // Serviço down
```

**Usage**:
```go
// Verificar antes de operações críticas
resp, err := client.HealthCheck(ctx, &emptypb.Empty{})
if err != nil || resp.Status != connectv1.HealthCheckResponse_HEALTH_STATUS_HEALTHY {
    return fmt.Errorf("conn-dict unhealthy: %v", resp)
}

// Prosseguir com operação
```

---

## 7. Error Handling

### 7.1 gRPC Status Codes

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

resp, err := client.GetEntry(ctx, req)
if err != nil {
    st, ok := status.FromError(err)
    if !ok {
        // Erro não é gRPC (network, timeout, etc.)
        return fmt.Errorf("network error: %w", err)
    }

    switch st.Code() {
    case codes.InvalidArgument:
        // 400 Bad Request: Validação falhou
        return ErrInvalidRequest

    case codes.NotFound:
        // 404 Not Found: Entry/Claim/Infraction não existe
        return ErrNotFound

    case codes.AlreadyExists:
        // 409 Conflict: Claim/Entry já existe
        return ErrAlreadyExists

    case codes.FailedPrecondition:
        // 412 Precondition Failed: Operação não permitida no estado atual
        return ErrNotAllowed

    case codes.Internal:
        // 500 Internal Error: Database, Temporal, Pulsar falhou
        log.Error("conn-dict internal error", "message", st.Message())
        return ErrInternalServer

    case codes.Unavailable:
        // 503 Service Unavailable: conn-dict down
        return ErrServiceUnavailable

    case codes.DeadlineExceeded:
        // 504 Timeout: Request timeout
        return ErrTimeout

    default:
        return fmt.Errorf("unexpected grpc error: %v", st)
    }
}
```

### 7.2 Structured Error Details

conn-dict pode retornar **erro estruturado** em metadata:

```go
resp, err := client.CreateClaim(ctx, req)
if err != nil {
    st, ok := status.FromError(err)
    if ok {
        // Extrair detalhes do erro
        for _, detail := range st.Details() {
            switch v := detail.(type) {
            case *commonv1.ValidationError:
                fmt.Printf("Validation failed: field=%s, constraint=%s\n", v.Field, v.Constraint)
            case *commonv1.BusinessError:
                fmt.Printf("Business error: code=%s, message=%s\n", v.ErrorCode, v.Message)
            case *commonv1.InfrastructureError:
                fmt.Printf("Infra error: component=%s, retriable=%v\n", v.Component, v.Retriable)
            }
        }
    }
}
```

---

## 8. Timeouts e Retry

### 8.1 Timeouts Recomendados

```go
// Query rápido (GetEntry, GetClaim)
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := client.GetEntry(ctx, req)

// Query lento (ListEntries, ListClaims)
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

resp, err := client.ListEntries(ctx, req)

// Workflow start (CreateClaim, CreateInfraction)
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := client.CreateClaim(ctx, req)
```

**Tabela de Timeouts**:

| Operação | Timeout Recomendado | Justificativa |
|----------|---------------------|---------------|
| GetEntry, GetEntryByKey | 5s | Query database simples |
| GetClaim, GetInfraction | 5s | Query database simples |
| ListEntries, ListClaims, ListInfractions | 10s | Query com paginação |
| CreateClaim, CreateInfraction | 30s | Start Temporal workflow (overhead) |
| ConfirmClaim, CancelClaim | 10s | Send Temporal signal |
| HealthCheck | 3s | Health check rápido |

### 8.2 Retry Policy (Automático)

Configure retry automático no cliente gRPC:

```go
retryPolicy := `{
    "methodConfig": [{
        "name": [{"service": "dict.connect.v1.ConnectService"}],
        "retryPolicy": {
            "maxAttempts": 3,
            "initialBackoff": "0.1s",
            "maxBackoff": "1s",
            "backoffMultiplier": 2,
            "retryableStatusCodes": ["UNAVAILABLE", "DEADLINE_EXCEEDED"]
        }
    }]
}`

conn, err := grpc.Dial(
    "localhost:9092",
    grpc.WithDefaultServiceConfig(retryPolicy),
    grpc.WithTransportCredentials(insecure.NewCredentials()),
)
```

**Retry Behavior**:
- Retry apenas em `UNAVAILABLE` (503) e `DEADLINE_EXCEEDED` (504)
- Máximo 3 tentativas
- Backoff exponencial: 100ms → 200ms → 400ms

### 8.3 Circuit Breaker (Produção)

Implementar circuit breaker para evitar cascading failures:

```go
import "github.com/sony/gobreaker"

// Circuit breaker settings
cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "conn-dict",
    MaxRequests: 3,
    Interval:    10 * time.Second,
    Timeout:     30 * time.Second,
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
        return counts.Requests >= 10 && failureRatio >= 0.6
    },
})

// Wrapp gRPC call
result, err := cb.Execute(func() (interface{}, error) {
    return client.GetEntry(ctx, req)
})
```

**Circuit States**:
- **Closed**: Normal operation
- **Open**: Too many failures → reject requests (30s)
- **Half-Open**: Test if service recovered (3 requests)

---

## 9. Observability

### 9.1 Metrics (Prometheus)

Instrumentar chamadas gRPC com Prometheus:

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    grpcCallsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "core_dict_grpc_calls_total",
            Help: "Total gRPC calls to conn-dict",
        },
        []string{"method", "status"},
    )

    grpcCallDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "core_dict_grpc_call_duration_seconds",
            Help:    "gRPC call duration",
            Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
        },
        []string{"method"},
    )
)

// Instrumentar chamada
func (c *ConnectClient) GetEntryWithMetrics(ctx context.Context, req *connectv1.GetEntryRequest) (*connectv1.GetEntryResponse, error) {
    start := time.Now()

    resp, err := c.client.GetEntry(ctx, req)

    duration := time.Since(start).Seconds()
    grpcCallDuration.WithLabelValues("GetEntry").Observe(duration)

    status := "success"
    if err != nil {
        status = "error"
    }
    grpcCallsTotal.WithLabelValues("GetEntry", status).Inc()

    return resp, err
}
```

### 9.2 Tracing (OpenTelemetry)

Habilitar tracing automático:

```go
import (
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/propagation"
)

// Inicializar OpenTelemetry
func initTracer() {
    // Setup exporter (Jaeger, Zipkin, etc.)
    // ...
}

// Criar cliente com tracing
conn, err := grpc.Dial(
    "localhost:9092",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
    grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
)
```

**Resultado**: Todas as chamadas gRPC são automaticamente traced com:
- Span ID, Trace ID
- Request/response size
- Error status
- Latency

### 9.3 Logging Estruturado

```go
import "log/slog"

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

resp, err := client.GetEntry(ctx, req)
if err != nil {
    logger.Error("GetEntry failed",
        "entry_id", req.EntryId,
        "error", err.Error(),
    )
    return nil, err
}

logger.Info("GetEntry success",
    "entry_id", resp.Entry.EntryId,
    "status", resp.Entry.Status,
)
```

---

## 10. Testing

### 10.1 Mock Client (Unit Tests)

```go
package infrastructure_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    connectv1 "github.com/lbpay-lab/dict-contracts/gen/proto/conn_dict/v1"
)

// MockConnectServiceClient é um mock do cliente gRPC
type MockConnectServiceClient struct {
    mock.Mock
}

func (m *MockConnectServiceClient) GetEntry(ctx context.Context, req *connectv1.GetEntryRequest, opts ...grpc.CallOption) (*connectv1.GetEntryResponse, error) {
    args := m.Called(ctx, req)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*connectv1.GetEntryResponse), args.Error(1)
}

// Teste unitário
func TestGetEntry(t *testing.T) {
    mockClient := new(MockConnectServiceClient)

    // Setup mock
    mockClient.On("GetEntry", mock.Anything, mock.Anything).Return(
        &connectv1.GetEntryResponse{
            Entry: &connectv1.Entry{
                EntryId:  "entry-123",
                KeyValue: "test@example.com",
            },
            Found: true,
        },
        nil,
    )

    // Testar
    resp, err := mockClient.GetEntry(context.Background(), &connectv1.GetEntryRequest{
        EntryId: "entry-123",
    })

    assert.NoError(t, err)
    assert.True(t, resp.Found)
    assert.Equal(t, "test@example.com", resp.Entry.KeyValue)

    mockClient.AssertExpectations(t)
}
```

### 10.2 Integration Test (conn-dict rodando)

```go
func TestConnectIntegration(t *testing.T) {
    // Assumir conn-dict rodando em localhost:9092
    client, err := NewConnectClient("localhost:9092")
    require.NoError(t, err)
    defer client.Close()

    ctx := context.Background()

    // Test GetEntry
    t.Run("GetEntry", func(t *testing.T) {
        resp, err := client.Client().GetEntry(ctx, &connectv1.GetEntryRequest{
            EntryId: "test-entry-123",
        })

        // Verificar resposta
        if err != nil {
            st, ok := status.FromError(err)
            require.True(t, ok)
            assert.Equal(t, codes.NotFound, st.Code())
        } else {
            assert.NotNil(t, resp)
        }
    })

    // Test CreateClaim
    t.Run("CreateClaim", func(t *testing.T) {
        resp, err := client.Client().CreateClaim(ctx, &connectv1.CreateClaimRequest{
            EntryId:     "test-entry-456",
            ClaimerIspb: "87654321",
            OwnerIspb:   "12345678",
            ClaimType:   connectv1.CreateClaimRequest_CLAIM_TYPE_OWNERSHIP,
        })

        require.NoError(t, err)
        assert.NotEmpty(t, resp.ClaimId)
        assert.Equal(t, commonv1.ClaimStatus_CLAIM_STATUS_OPEN, resp.Status)
    })
}
```

### 10.3 E2E Test (Docker Compose)

```yaml
# docker-compose.test.yml
version: '3.9'

services:
  conn-dict:
    image: conn-dict:latest
    ports:
      - "9092:9092"
    environment:
      - DATABASE_URL=postgres://test:test@postgres:5432/test
      - TEMPORAL_HOST=temporal:7233
    depends_on:
      - postgres
      - temporal

  core-dict-test:
    image: core-dict-test:latest
    command: go test -v ./tests/e2e/...
    depends_on:
      - conn-dict
```

```bash
# Rodar E2E tests
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

---

## 11. Production Checklist

Antes de deploy em produção:

- [ ] **Connection Pool** configurado (max connections, idle timeout)
- [ ] **Timeouts** definidos (5s queries, 30s mutations)
- [ ] **Retry policy** configurado (max 3 tentativas)
- [ ] **Circuit breaker** implementado
- [ ] **Metrics** instrumented (Prometheus)
- [ ] **Tracing** enabled (OpenTelemetry/Jaeger)
- [ ] **Logging** estruturado (JSON, com trace IDs)
- [ ] **Health check** monitoring (alertar se conn-dict down)
- [ ] **Graceful shutdown** implementado
- [ ] **TLS/mTLS** configurado (produção)
- [ ] **Load balancing** configurado (múltiplas instâncias conn-dict)
- [ ] **Rate limiting** configurado (evitar sobrecarga)
- [ ] **Integration tests** passando (CI/CD)
- [ ] **Performance tests** passando (>1000 TPS)
- [ ] **Security scan** passando (vulnerabilidades)

---

## 12. Configuration

### 12.1 Development

```go
// config/config.go
type Config struct {
    ConnectGRPCAddress string
    ConnectTimeout     time.Duration
    ConnectRetryMax    int
}

func LoadConfig() *Config {
    return &Config{
        ConnectGRPCAddress: getEnv("CONNECT_GRPC_ADDRESS", "localhost:9092"),
        ConnectTimeout:     parseDuration(getEnv("CONNECT_TIMEOUT", "10s")),
        ConnectRetryMax:    parseInt(getEnv("CONNECT_RETRY_MAX", "3")),
    }
}
```

```bash
# .env (development)
CONNECT_GRPC_ADDRESS=localhost:9092
CONNECT_TIMEOUT=10s
CONNECT_RETRY_MAX=3
```

### 12.2 Production (mTLS)

```go
import (
    "google.golang.org/grpc/credentials"
)

// Criar cliente com mTLS
func NewConnectClientWithTLS(address string, certFile, keyFile, caFile string) (*ConnectClient, error) {
    // Carregar certificado do cliente
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return nil, fmt.Errorf("failed to load client cert: %w", err)
    }

    // Carregar CA
    caCert, err := os.ReadFile(caFile)
    if err != nil {
        return nil, fmt.Errorf("failed to load CA: %w", err)
    }

    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)

    // Criar TLS config
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
        MinVersion:   tls.VersionTLS13,
    }

    creds := credentials.NewTLS(tlsConfig)

    // Criar conexão
    conn, err := grpc.Dial(
        address,
        grpc.WithTransportCredentials(creds),
    )
    if err != nil {
        return nil, err
    }

    return &ConnectClient{
        conn:   conn,
        client: connectv1.NewConnectServiceClient(conn),
    }, nil
}
```

```bash
# .env (production)
CONNECT_GRPC_ADDRESS=conn-dict.prod.lbpay.com.br:9092
CONNECT_TLS_CERT_FILE=/etc/certs/client-cert.pem
CONNECT_TLS_KEY_FILE=/etc/certs/client-key.pem
CONNECT_TLS_CA_FILE=/etc/certs/ca-cert.pem
CONNECT_TIMEOUT=10s
```

### 12.3 Kubernetes (Service Discovery)

```yaml
# kubernetes/core-dict-deployment.yaml
env:
  - name: CONNECT_GRPC_ADDRESS
    value: "conn-dict-service.dict-namespace.svc.cluster.local:9092"
  - name: CONNECT_TIMEOUT
    value: "10s"
```

---

## 13. Complete Production Example

```go
package infrastructure

import (
    "context"
    "crypto/tls"
    "crypto/x509"
    "fmt"
    "log/slog"
    "os"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/sony/gobreaker"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/credentials"
    "google.golang.org/grpc/keepalive"
    "google.golang.org/grpc/status"
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

    connectv1 "github.com/lbpay-lab/dict-contracts/gen/proto/conn_dict/v1"
    commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
)

// Metrics
var (
    grpcCallsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "core_dict_grpc_calls_total",
            Help: "Total gRPC calls to conn-dict",
        },
        []string{"method", "status"},
    )

    grpcCallDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "core_dict_grpc_call_duration_seconds",
            Help:    "gRPC call duration",
            Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
        },
        []string{"method"},
    )
)

// ConnectClient é o cliente production-ready para conn-dict
type ConnectClient struct {
    conn   *grpc.ClientConn
    client connectv1.ConnectServiceClient
    cb     *gobreaker.CircuitBreaker
    logger *slog.Logger
}

// Config contém configuração do cliente
type Config struct {
    Address    string
    CertFile   string
    KeyFile    string
    CAFile     string
    Timeout    time.Duration
    MaxRetries int
}

// NewConnectClientProduction cria cliente production-ready
func NewConnectClientProduction(cfg *Config) (*ConnectClient, error) {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    // Setup TLS
    var opts []grpc.DialOption
    if cfg.CertFile != "" {
        creds, err := loadTLSCredentials(cfg.CertFile, cfg.KeyFile, cfg.CAFile)
        if err != nil {
            return nil, fmt.Errorf("failed to load TLS: %w", err)
        }
        opts = append(opts, grpc.WithTransportCredentials(creds))
    } else {
        opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
    }

    // Retry policy
    retryPolicy := fmt.Sprintf(`{
        "methodConfig": [{
            "name": [{"service": "dict.connect.v1.ConnectService"}],
            "retryPolicy": {
                "maxAttempts": %d,
                "initialBackoff": "0.1s",
                "maxBackoff": "1s",
                "backoffMultiplier": 2,
                "retryableStatusCodes": ["UNAVAILABLE", "DEADLINE_EXCEEDED"]
            }
        }]
    }`, cfg.MaxRetries)

    opts = append(opts,
        grpc.WithDefaultServiceConfig(retryPolicy),
        grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
        grpc.WithKeepaliveParams(keepalive.ClientParameters{
            Time:                10 * time.Second,
            Timeout:             3 * time.Second,
            PermitWithoutStream: true,
        }),
    )

    // Connect
    conn, err := grpc.Dial(cfg.Address, opts...)
    if err != nil {
        return nil, fmt.Errorf("failed to dial: %w", err)
    }

    client := connectv1.NewConnectServiceClient(conn)

    // Circuit breaker
    cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
        Name:        "conn-dict",
        MaxRequests: 3,
        Interval:    10 * time.Second,
        Timeout:     30 * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= 10 && failureRatio >= 0.6
        },
        OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
            logger.Warn("circuit breaker state changed",
                "name", name,
                "from", from.String(),
                "to", to.String(),
            )
        },
    })

    return &ConnectClient{
        conn:   conn,
        client: client,
        cb:     cb,
        logger: logger,
    }, nil
}

// GetEntry busca entry por ID (production-ready)
func (c *ConnectClient) GetEntry(ctx context.Context, entryID string) (*connectv1.Entry, error) {
    start := time.Now()
    method := "GetEntry"

    defer func() {
        duration := time.Since(start).Seconds()
        grpcCallDuration.WithLabelValues(method).Observe(duration)
    }()

    // Circuit breaker
    result, err := c.cb.Execute(func() (interface{}, error) {
        ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
        defer cancel()

        resp, err := c.client.GetEntry(ctx, &connectv1.GetEntryRequest{
            EntryId: entryID,
        })
        if err != nil {
            return nil, err
        }

        if !resp.Found {
            return nil, status.Error(codes.NotFound, "entry not found")
        }

        return resp.Entry, nil
    })

    if err != nil {
        grpcCallsTotal.WithLabelValues(method, "error").Inc()
        c.logger.Error("GetEntry failed",
            "entry_id", entryID,
            "error", err.Error(),
        )
        return nil, err
    }

    grpcCallsTotal.WithLabelValues(method, "success").Inc()
    c.logger.Info("GetEntry success",
        "entry_id", entryID,
    )

    return result.(*connectv1.Entry), nil
}

// CreateClaim inicia claim workflow (production-ready)
func (c *ConnectClient) CreateClaim(ctx context.Context, req *connectv1.CreateClaimRequest) (*connectv1.CreateClaimResponse, error) {
    start := time.Now()
    method := "CreateClaim"

    defer func() {
        duration := time.Since(start).Seconds()
        grpcCallDuration.WithLabelValues(method).Observe(duration)
    }()

    result, err := c.cb.Execute(func() (interface{}, error) {
        ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
        defer cancel()

        return c.client.CreateClaim(ctx, req)
    })

    if err != nil {
        grpcCallsTotal.WithLabelValues(method, "error").Inc()
        c.logger.Error("CreateClaim failed",
            "entry_id", req.EntryId,
            "error", err.Error(),
        )
        return nil, err
    }

    grpcCallsTotal.WithLabelValues(method, "success").Inc()
    resp := result.(*connectv1.CreateClaimResponse)
    c.logger.Info("CreateClaim success",
        "claim_id", resp.ClaimId,
        "entry_id", req.EntryId,
    )

    return resp, nil
}

// Close fecha a conexão
func (c *ConnectClient) Close() error {
    return c.conn.Close()
}

// loadTLSCredentials carrega credenciais mTLS
func loadTLSCredentials(certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return nil, err
    }

    caCert, err := os.ReadFile(caFile)
    if err != nil {
        return nil, err
    }

    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)

    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
        MinVersion:   tls.VersionTLS13,
    }

    return credentials.NewTLS(tlsConfig), nil
}
```

**Uso**:
```go
func main() {
    cfg := &Config{
        Address:    os.Getenv("CONNECT_GRPC_ADDRESS"),
        CertFile:   os.Getenv("CONNECT_TLS_CERT_FILE"),
        KeyFile:    os.Getenv("CONNECT_TLS_KEY_FILE"),
        CAFile:     os.Getenv("CONNECT_TLS_CA_FILE"),
        Timeout:    10 * time.Second,
        MaxRetries: 3,
    }

    client, err := NewConnectClientProduction(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Usar client
    entry, err := client.GetEntry(context.Background(), "entry-123")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Entry: %v\n", entry)
}
```

---

## 14. Quick Reference

### 14.1 All 15 gRPC RPCs

| # | RPC | Purpose | Latency | Timeout |
|---|-----|---------|---------|---------|
| 1 | GetEntry | Buscar entry por ID | <10ms | 5s |
| 2 | GetEntryByKey | Buscar entry por chave | <10ms | 5s |
| 3 | ListEntries | Listar entries do participante | <50ms | 10s |
| 4 | CreateClaim | Iniciar claim (30 dias) | <100ms | 30s |
| 5 | ConfirmClaim | Owner aceita claim | <50ms | 10s |
| 6 | CancelClaim | Cancelar claim | <50ms | 10s |
| 7 | GetClaim | Buscar claim por ID | <10ms | 5s |
| 8 | ListClaims | Listar claims com filtros | <50ms | 10s |
| 9 | CreateInfraction | Reportar infração | <100ms | 30s |
| 10 | InvestigateInfraction | Analista decide | <50ms | 10s |
| 11 | ResolveInfraction | Resolver infração | <50ms | 10s |
| 12 | DismissInfraction | Descartar infração | <50ms | 10s |
| 13 | GetInfraction | Buscar infração por ID | <10ms | 5s |
| 14 | ListInfractions | Listar infrações | <50ms | 10s |
| 15 | HealthCheck | Health check | <50ms | 3s |

### 14.2 Error Codes Summary

| gRPC Code | HTTP | Quando Usar |
|-----------|------|-------------|
| OK | 200 | Success |
| INVALID_ARGUMENT | 400 | Validação falhou |
| NOT_FOUND | 404 | Entry/Claim/Infraction não existe |
| ALREADY_EXISTS | 409 | Claim/Entry já existe |
| FAILED_PRECONDITION | 412 | Operação não permitida no estado atual |
| INTERNAL | 500 | Erro database/Temporal/Pulsar |
| UNAVAILABLE | 503 | Serviço down |
| DEADLINE_EXCEEDED | 504 | Timeout |

### 14.3 Environment Variables

```bash
# Development
CONNECT_GRPC_ADDRESS=localhost:9092
CONNECT_TIMEOUT=10s
CONNECT_RETRY_MAX=3

# Production (mTLS)
CONNECT_GRPC_ADDRESS=conn-dict.prod:9092
CONNECT_TLS_CERT_FILE=/etc/certs/client-cert.pem
CONNECT_TLS_KEY_FILE=/etc/certs/client-key.pem
CONNECT_TLS_CA_FILE=/etc/certs/ca-cert.pem
CONNECT_TIMEOUT=10s
CONNECT_RETRY_MAX=3
```

### 14.4 SLA Esperado (conn-dict)

| Métrica | Target | Alert Threshold |
|---------|--------|-----------------|
| Disponibilidade | 99.9% | < 99.5% |
| Latência p50 (queries) | < 10ms | > 50ms |
| Latência p99 (queries) | < 50ms | > 200ms |
| Latência p50 (mutations) | < 100ms | > 500ms |
| Error rate | < 0.1% | > 1% |

---

## Appendix A: Proto Files Reference

**Location**: `/Users/jose.silva.lb/LBPay/IA_Dict/dict-contracts/proto/conn_dict/v1/connect_service.proto`

**Generated Go Code**: `/Users/jose.silva.lb/LBPay/IA_Dict/dict-contracts/gen/proto/conn_dict/v1/`

**Common Types**: `/Users/jose.silva.lb/LBPay/IA_Dict/dict-contracts/proto/common.proto`

---

## Appendix B: Related Documentation

- **conn-dict API Reference**: `/Users/jose.silva.lb/LBPay/IA_Dict/Artefatos/00_Master/CONN_DICT_API_REFERENCE.md`
- **Temporal Workflows**: Ver ClaimWorkflow, InfractionWorkflow, VSyncWorkflow em conn-dict
- **Pulsar Events**: Ver dict.entries.created, dict.claims.completed em conn-dict
- **Bridge Integration**: conn-dict chama conn-bridge via gRPC para comunicação com Bacen

---

## Contact & Support

**Maintainer**: DICT Implementation Squad
**Slack Channel**: #dict-support
**Issue Tracker**: GitHub Issues

**Documentation Version**: 1.0
**Last Updated**: 2025-10-27
**Status**: PRODUCTION-READY

---

**FIM DO GUIA**
