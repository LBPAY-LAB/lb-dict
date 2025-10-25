# GRPC-004: Error Handling gRPC

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Versão**: 1.0
**Data**: 2025-10-25
**Status**: ✅ Especificação Completa
**Responsável**: ARCHITECT (AI Agent - Technical Architect)

---

## 📋 Resumo Executivo

Este documento especifica a **estratégia completa de tratamento de erros em comunicações gRPC** entre os componentes do sistema DICT (Core ↔ Connect ↔ Bridge), incluindo códigos de status, detalhes de erro, retry policies, e propagação de erros.

**Objetivo**: Padronizar tratamento de erros em todas as comunicações gRPC para facilitar debugging, melhorar observabilidade, e garantir resiliência do sistema.

---

## 🎯 Princípios de Error Handling

### 1. Fail Fast
- Retornar erros imediatamente quando detectados
- Não mascarar erros com valores default
- Validar inputs no início do processamento

### 2. Contexto Rico
- Incluir informações suficientes para debugging
- Adicionar request_id para rastreamento
- Incluir stack trace em desenvolvimento (NÃO em produção)

### 3. Idempotência
- Operações devem ser safe para retry
- Usar idempotency keys quando necessário
- Documentar quais operações são idempotentes

### 4. Graceful Degradation
- Sistema deve continuar funcionando parcialmente se possível
- Não propagar falhas para componentes não afetados
- Usar circuit breakers para proteger downstream services

---

## 📊 Códigos de Status gRPC

### Mapeamento de Erros

| gRPC Status | HTTP Equiv | Quando Usar | Retry? |
|-------------|------------|-------------|--------|
| **OK** (0) | 200 OK | Sucesso | N/A |
| **CANCELLED** (1) | 499 | Cliente cancelou request | ❌ Não |
| **UNKNOWN** (2) | 500 | Erro desconhecido/inesperado | ⚠️ Talvez |
| **INVALID_ARGUMENT** (3) | 400 | Input inválido (validação falhou) | ❌ Não |
| **DEADLINE_EXCEEDED** (4) | 504 | Timeout excedido | ✅ Sim |
| **NOT_FOUND** (5) | 404 | Recurso não encontrado | ❌ Não |
| **ALREADY_EXISTS** (6) | 409 | Recurso já existe (duplicate key) | ❌ Não |
| **PERMISSION_DENIED** (7) | 403 | Sem permissão | ❌ Não |
| **RESOURCE_EXHAUSTED** (8) | 429 | Rate limit excedido | ✅ Sim (com backoff) |
| **FAILED_PRECONDITION** (9) | 400 | Pré-condição falhada (ex: claim expirada) | ❌ Não |
| **ABORTED** (10) | 409 | Operação abortada (conflito) | ✅ Sim |
| **OUT_OF_RANGE** (11) | 400 | Valor fora do range válido | ❌ Não |
| **UNIMPLEMENTED** (12) | 501 | RPC não implementado | ❌ Não |
| **INTERNAL** (13) | 500 | Erro interno do servidor | ⚠️ Talvez |
| **UNAVAILABLE** (14) | 503 | Serviço indisponível | ✅ Sim |
| **DATA_LOSS** (15) | 500 | Perda de dados | ❌ Não |
| **UNAUTHENTICATED** (16) | 401 | Não autenticado | ❌ Não |

---

## 🔧 Implementação em Go

### 1. Retornar Erros Estruturados

```go
// Pseudocódigo (especificação, NÃO implementar agora)
package grpcerrors

import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/genproto/googleapis/rpc/errdetails"
)

// NewInvalidArgumentError cria erro de validação
func NewInvalidArgumentError(field, message string) error {
    st := status.New(codes.InvalidArgument, "Validation failed")

    // Adicionar detalhes de erro
    br := &errdetails.BadRequest{
        FieldViolations: []*errdetails.BadRequest_FieldViolation{
            {
                Field:       field,
                Description: message,
            },
        },
    }

    st, err := st.WithDetails(br)
    if err != nil {
        return status.Error(codes.InvalidArgument, message)
    }

    return st.Err()
}

// NewNotFoundError cria erro de recurso não encontrado
func NewNotFoundError(resourceType, resourceID string) error {
    st := status.New(codes.NotFound, "Resource not found")

    ri := &errdetails.ResourceInfo{
        ResourceType: resourceType,
        ResourceName: resourceID,
        Description:  fmt.Sprintf("%s with ID %s not found", resourceType, resourceID),
    }

    st, _ = st.WithDetails(ri)
    return st.Err()
}

// NewAlreadyExistsError cria erro de duplicate key
func NewAlreadyExistsError(resourceType, resourceID string) error {
    st := status.New(codes.AlreadyExists, "Resource already exists")

    ri := &errdetails.ResourceInfo{
        ResourceType: resourceType,
        ResourceName: resourceID,
        Description:  fmt.Sprintf("%s with ID %s already exists", resourceType, resourceID),
    }

    st, _ = st.WithDetails(ri)
    return st.Err()
}

// NewUnavailableError cria erro de serviço indisponível
func NewUnavailableError(service string, retryDelay time.Duration) error {
    st := status.New(codes.Unavailable, "Service temporarily unavailable")

    retryInfo := &errdetails.RetryInfo{
        RetryDelay: durationpb.New(retryDelay),
    }

    st, _ = st.WithDetails(retryInfo)
    return st.Err()
}

// NewRateLimitError cria erro de rate limit
func NewRateLimitError(quota string, retryDelay time.Duration) error {
    st := status.New(codes.ResourceExhausted, "Rate limit exceeded")

    qi := &errdetails.QuotaFailure{
        Violations: []*errdetails.QuotaFailure_Violation{
            {
                Subject:     quota,
                Description: "Rate limit exceeded for " + quota,
            },
        },
    }

    retryInfo := &errdetails.RetryInfo{
        RetryDelay: durationpb.New(retryDelay),
    }

    st, _ = st.WithDetails(qi, retryInfo)
    return st.Err()
}
```

---

### 2. Exemplo de Uso no Handler

```go
// Pseudocódigo
func (s *ConnectServiceServer) CreateEntry(ctx context.Context, req *pb.CreateEntryRequest) (*pb.CreateEntryResponse, error) {
    // 1. Validação de input
    if req.Key == nil {
        return nil, grpcerrors.NewInvalidArgumentError("key", "key is required")
    }

    if req.Key.KeyType == pb.KeyType_KEY_TYPE_UNSPECIFIED {
        return nil, grpcerrors.NewInvalidArgumentError("key.keyType", "keyType must be specified")
    }

    if !isValidCPF(req.Key.KeyValue) {
        return nil, grpcerrors.NewInvalidArgumentError("key.keyValue", "invalid CPF format")
    }

    // 2. Verificar se entry já existe
    existing, err := s.entryRepo.GetByKey(ctx, req.Key.KeyType.String(), req.Key.KeyValue)
    if err == nil && existing != nil {
        return nil, grpcerrors.NewAlreadyExistsError("entry", existing.ID)
    }

    // 3. Chamar Bridge (pode falhar)
    bridgeResp, err := s.bridgeClient.CreateEntry(ctx, &bridgepb.CreateEntryRequest{
        Key:     req.Key,
        Account: req.Account,
    })
    if err != nil {
        // Propagar erro do Bridge (já está no formato correto)
        return nil, err
    }

    // 4. Salvar no banco (pode falhar)
    entry := &domain.Entry{
        ID:         uuid.New().String(),
        KeyType:    req.Key.KeyType.String(),
        KeyValue:   req.Key.KeyValue,
        ExternalID: bridgeResp.ExternalId,
        Status:     "ACTIVE",
    }

    if err := s.entryRepo.Create(ctx, entry); err != nil {
        // Erro interno do PostgreSQL
        return nil, status.Error(codes.Internal, "failed to save entry")
    }

    // 5. Sucesso
    return &pb.CreateEntryResponse{
        EntryId:   entry.ID,
        Status:    pb.EntryStatus_ENTRY_STATUS_ACTIVE,
        CreatedAt: timestamppb.Now(),
    }, nil
}
```

---

### 3. Extrair Detalhes de Erro no Cliente

```go
// Pseudocódigo
func handleCreateEntryError(err error) {
    st, ok := status.FromError(err)
    if !ok {
        // Erro não é gRPC
        log.Errorf("Non-gRPC error: %v", err)
        return
    }

    // Log código de status
    log.Errorf("gRPC error: code=%s, message=%s", st.Code(), st.Message())

    // Extrair detalhes
    for _, detail := range st.Details() {
        switch t := detail.(type) {
        case *errdetails.BadRequest:
            // Erro de validação
            for _, violation := range t.FieldViolations {
                log.Errorf("Validation error: field=%s, description=%s",
                    violation.Field, violation.Description)
            }

        case *errdetails.ResourceInfo:
            // Recurso não encontrado ou já existe
            log.Errorf("Resource error: type=%s, name=%s, description=%s",
                t.ResourceType, t.ResourceName, t.Description)

        case *errdetails.RetryInfo:
            // Informação de retry
            log.Infof("Retry after: %v", t.RetryDelay.AsDuration())

        case *errdetails.QuotaFailure:
            // Rate limit
            for _, violation := range t.Violations {
                log.Errorf("Quota exceeded: subject=%s, description=%s",
                    violation.Subject, violation.Description)
            }
        }
    }
}
```

---

## 🔄 Retry Policies

### Estratégia de Retry

```go
// Pseudocódigo
import "github.com/grpc-ecosystem/go-grpc-middleware/retry"

// Configuração de retry para gRPC client
opts := []grpc_retry.CallOption{
    grpc_retry.WithMax(3),  // Máximo 3 retries
    grpc_retry.WithPerRetryTimeout(5 * time.Second),  // Timeout por tentativa
    grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),  // Backoff exponencial
    grpc_retry.WithCodes(  // Apenas retry em códigos específicos
        codes.Unavailable,
        codes.DeadlineExceeded,
        codes.ResourceExhausted,
        codes.Aborted,
    ),
}

// Criar cliente com retry
conn, err := grpc.Dial(
    "bridge.dict.svc.cluster.local:8081",
    grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(opts...)),
)
```

### Backoff Exponencial

```
Tentativa 1: Imediato
Tentativa 2: 100ms + jitter
Tentativa 3: 200ms + jitter
Tentativa 4: 400ms + jitter

Jitter: ±25% (evitar thundering herd)
```

---

## 🛡️ Circuit Breaker

### Quando Usar

- Proteger serviços downstream (Bridge, Bacen)
- Evitar sobrecarga em serviços falhos
- Fail fast quando serviço está indisponível

### Implementação (Pseudocódigo)

```go
import "github.com/sony/gobreaker"

// Configuração do circuit breaker
cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "BridgeClient",
    MaxRequests: 3,  // Requests permitidos em half-open state
    Interval:    10 * time.Second,  // Janela para contar falhas
    Timeout:     30 * time.Second,  // Tempo em open state antes de half-open
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        // Abrir circuit se 50% das requests falharem (mínimo 10 requests)
        failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
        return counts.Requests >= 10 && failureRatio >= 0.5
    },
    OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
        log.Infof("Circuit breaker %s changed state: %s -> %s", name, from, to)
    },
})

// Wrapper para chamar Bridge com circuit breaker
func (c *BridgeClient) CreateEntryWithCircuitBreaker(ctx context.Context, req *pb.CreateEntryRequest) (*pb.CreateEntryResponse, error) {
    result, err := cb.Execute(func() (interface{}, error) {
        return c.bridgeClient.CreateEntry(ctx, req)
    })

    if err != nil {
        if err == gobreaker.ErrOpenState {
            // Circuit breaker está aberto (serviço indisponível)
            return nil, grpcerrors.NewUnavailableError("Bridge", 30*time.Second)
        }
        return nil, err
    }

    return result.(*pb.CreateEntryResponse), nil
}
```

### Estados do Circuit Breaker

```
CLOSED (normal) ──[50% falhas]──> OPEN (fail fast)
      ^                                  │
      │                                  │ [30s timeout]
      │                                  ▼
      └──────[3 sucessos]────── HALF-OPEN (testando)
```

---

## 📊 Cenários de Erro

### Cenário 1: Entry Não Encontrada

**Situação**: Cliente solicita GetEntry com ID inexistente

**Response**:
```go
return nil, status.Error(codes.NotFound, "entry not found")

// Ou com detalhes:
st := status.New(codes.NotFound, "entry not found")
ri := &errdetails.ResourceInfo{
    ResourceType: "entry",
    ResourceName: entryID,
    Description:  "Entry with ID " + entryID + " not found",
}
st, _ = st.WithDetails(ri)
return nil, st.Err()
```

**Cliente deve**:
- ❌ NÃO fazer retry
- Tratar como erro de negócio (exibir mensagem ao usuário)

---

### Cenário 2: Duplicate Key (Chave Já Existe)

**Situação**: Cliente tenta criar entry com CPF que já existe

**Response**:
```go
return nil, grpcerrors.NewAlreadyExistsError("entry", cpf)
```

**Cliente deve**:
- ❌ NÃO fazer retry
- Exibir mensagem: "Esta chave PIX já está cadastrada"

---

### Cenário 3: Validação de Input Falhou

**Situação**: CPF inválido enviado no request

**Response**:
```go
return nil, grpcerrors.NewInvalidArgumentError("key.keyValue", "invalid CPF format: must be 11 digits")
```

**Cliente deve**:
- ❌ NÃO fazer retry
- Exibir erro de validação no campo correspondente

---

### Cenário 4: Timeout (Bacen Demorou Demais)

**Situação**: Bacen não respondeu em 5 segundos

**Response**:
```go
return nil, status.Error(codes.DeadlineExceeded, "request to Bacen timed out after 5s")
```

**Cliente deve**:
- ✅ Fazer retry (com backoff exponencial)
- Limite: 3 tentativas

---

### Cenário 5: Rate Limit Excedido

**Situação**: ISPB excedeu limite de 100 requests/min

**Response**:
```go
return nil, grpcerrors.NewRateLimitError("CreateEntry:ispb:00000000", 60*time.Second)
```

**Cliente deve**:
- ✅ Fazer retry após 60 segundos (conforme RetryInfo)
- Exibir mensagem: "Limite de requisições excedido. Tente novamente em 1 minuto"

---

### Cenário 6: Bridge Indisponível

**Situação**: Bridge não está respondendo (connection refused)

**Response**:
```go
return nil, grpcerrors.NewUnavailableError("Bridge", 30*time.Second)
```

**Cliente deve**:
- ✅ Fazer retry (com backoff exponencial)
- Circuit breaker pode abrir após múltiplas falhas

---

### Cenário 7: Erro Interno (PostgreSQL Down)

**Situação**: PostgreSQL não está acessível

**Response**:
```go
// NÃO expor detalhes internos
return nil, status.Error(codes.Internal, "internal server error")

// Logar detalhes internamente
log.Errorf("PostgreSQL connection failed: %v", err)
```

**Cliente deve**:
- ⚠️ Pode tentar retry (mas provavelmente falhará novamente)
- Escalar para equipe de ops se persistir

---

## 🔍 Observabilidade

### Logging de Erros

```go
// Pseudocódigo
import "go.uber.org/zap"

// Interceptor para logging de erros gRPC
func ErrorLoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    resp, err := handler(ctx, req)

    if err != nil {
        st, _ := status.FromError(err)

        logger := zap.L().With(
            zap.String("method", info.FullMethod),
            zap.String("grpc_code", st.Code().String()),
            zap.String("error_message", st.Message()),
            zap.String("request_id", getRequestID(ctx)),
        )

        // Log level baseado em código de status
        switch st.Code() {
        case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists:
            // Erros de cliente (não são graves)
            logger.Info("Client error")

        case codes.Internal, codes.DataLoss:
            // Erros internos (graves)
            logger.Error("Internal error", zap.Error(err))

        case codes.Unavailable, codes.DeadlineExceeded:
            // Erros de disponibilidade (warning)
            logger.Warn("Availability error")

        default:
            logger.Info("gRPC error")
        }
    }

    return resp, err
}
```

### Métricas Prometheus

```go
// Pseudocódigo
var (
    grpcErrors = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "grpc_errors_total",
            Help: "Total gRPC errors by code",
        },
        []string{"method", "code"},
    )

    grpcRetries = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "grpc_retries_total",
            Help: "Total gRPC retries",
        },
        []string{"method"},
    )
)

// Incrementar em cada erro
if err != nil {
    st, _ := status.FromError(err)
    grpcErrors.WithLabelValues(method, st.Code().String()).Inc()
}
```

### Alertas

```yaml
# Prometheus alert rules
groups:
  - name: grpc_errors
    rules:
      - alert: HighGRPCErrorRate
        expr: |
          rate(grpc_errors_total{code!="OK"}[5m]) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High gRPC error rate"
          description: "{{ $labels.method }} has {{ $value }} errors/sec"

      - alert: GRPCInternalErrors
        expr: |
          rate(grpc_errors_total{code="INTERNAL"}[5m]) > 1
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "gRPC internal errors detected"

      - alert: CircuitBreakerOpen
        expr: |
          circuit_breaker_state{state="open"} == 1
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Circuit breaker {{ $labels.name }} is OPEN"
```

---

## 📋 Checklist de Implementação

- [ ] Criar helpers para erros estruturados (grpcerrors package)
- [ ] Implementar interceptor de logging de erros
- [ ] Configurar retry policies em todos os gRPC clients
- [ ] Implementar circuit breakers para serviços críticos (Bridge, Bacen)
- [ ] Adicionar métricas Prometheus para erros gRPC
- [ ] Configurar alertas para high error rate
- [ ] Documentar códigos de erro em Swagger/OpenAPI (para APIs REST)
- [ ] Criar runbook de troubleshooting de erros comuns
- [ ] Testar todos os cenários de erro (testes E2E)
- [ ] Validar propagação de request_id em toda a stack

---

## 📚 Referências

### Documentos Internos
- [GRPC-001: Bridge gRPC Service](GRPC-001_Bridge_gRPC_Service.md) - RPC definitions
- [GRPC-003: Proto Files Specification](GRPC-003_Proto_Files_Specification.md) - Error messages
- [TEC-002 v3.1: Bridge Specification](../../11_Especificacoes_Tecnicas/TEC-002_Bridge_Specification.md)
- [TEC-003 v2.1: Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)

### Documentação Externa
- [gRPC Error Handling](https://grpc.io/docs/guides/error/)
- [gRPC Status Codes](https://grpc.github.io/grpc/core/md_doc_statuscodes.html)
- [Google API Error Model](https://cloud.google.com/apis/design/errors)
- [grpc-ecosystem/go-grpc-middleware](https://github.com/grpc-ecosystem/go-grpc-middleware)
- [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html)

---

**Versão**: 1.0
**Status**: ✅ Especificação Completa (Aguardando implementação)
**Próxima Revisão**: Após implementação (validar error rates em produção)

---

**IMPORTANTE**: Este é um documento de **especificação técnica**. A implementação será feita pelos desenvolvedores em fase posterior, baseando-se neste documento.
