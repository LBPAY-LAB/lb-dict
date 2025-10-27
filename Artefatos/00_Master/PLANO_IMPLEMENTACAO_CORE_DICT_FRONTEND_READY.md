# Plano de Implementação: Core-Dict Front-End Ready

**Data**: 2025-10-27
**Versão**: 2.0
**Status**: 🚀 Em Execução
**Objetivo**: Implementação completa do Core-Dict para integração com Front-End

---

## 🎯 Objetivo

Completar a implementação do **Core-Dict** para atender **100% das necessidades do Front-End**, integrando:
- ✅ 15 RPCs gRPC (já definidos)
- ✅ Command/Query Handlers (já implementados em Application Layer)
- ✅ Requisitos funcionais do Manual DICT Bacen
- ✅ User Stories do Front-End

**Timeline**: **2-3 dias** (Segunda a Quarta-feira)

---

## 📋 Contexto Atual

### ✅ O que JÁ está implementado (95%)

#### 1. Domain Layer (100%) ✅
- Entities: Entry, Claim, Account, Infraction
- Value Objects: KeyType, KeyStatus, ClaimStatus, ClaimType, Participant
- Repositories Interfaces
- Domain Events
- **176 testes** (100% passando)

#### 2. Application Layer (100%) ✅
- **10 Command Handlers**:
  - CreateEntryCommand
  - DeleteEntryCommand
  - CreateClaimCommand
  - ConfirmClaimCommand
  - CancelClaimCommand
  - CompleteClaimCommand
  - BlockEntryCommand
  - UnblockEntryCommand
  - CreateInfractionCommand
  - UpdateEntryCommand (via claim)

- **10 Query Handlers**:
  - GetEntryQuery
  - ListEntriesQuery
  - GetClaimQuery
  - ListClaimsQuery
  - GetStatisticsQuery
  - HealthCheckQuery
  - VerifyAccountQuery
  - GetAuditLogQuery
  - ListClaimsByEntryQuery
  - ListExpiredClaimsQuery

- **6 Services**:
  - KeyValidatorService (CPF, CNPJ, Email, Phone, EVP)
  - AccountOwnershipService
  - DuplicateKeyChecker
  - EventPublisherService
  - CacheService (5 strategies)
  - NotificationService (webhook/email/slack)

- **73 testes** (88% cobertura)

#### 3. Infrastructure Layer (100%) ✅
- **4 Repositories** (PostgreSQL + pgx):
  - EntryRepository
  - ClaimRepository
  - AccountRepository
  - AuditRepository

- **Database** (5 migrations):
  - dict_entries (partitioned by month, RLS)
  - claims (30-day tracking)
  - accounts
  - audit_log
  - sync_reports

- **Cache** (Redis):
  - 5 cache strategies implemented
  - Rate limiting (token bucket)

- **Messaging** (Pulsar):
  - EntryEventProducer (3 topics)
  - EntryEventConsumer (5 handlers)
  - Event streaming E2E

- **gRPC Client** (ConnectClient):
  - 17 RPCs para conn-dict
  - Circuit Breaker
  - Retry Policy com exponential backoff

#### 4. gRPC Server (80%) ⚠️
- **CoreDictServiceHandler** (15 métodos implementados)
- **Status**:
  - ✅ Skeleton 100% (validações, mocks)
  - ✅ Interceptors (auth, logging, metrics, rate limit, recovery)
  - ⚠️ **Falta**: Integração com Application Layer

---

## ⚠️ O que FALTA implementar (5%)

### Gap 1: Integração gRPC Handlers ↔ Application Layer (P0 - CRÍTICO)

**Problema**: Handlers retornam mocks, não executam business logic real

**Solução**: Conectar os 15 métodos gRPC com command/query handlers

**Impacto**: **BLOQUEANTE** para Front-End

---

## 🚀 Plano de Implementação Detalhado

### Fase 1: Mappers Proto ↔ Domain (4 horas)

**Objetivo**: Criar funções de mapeamento entre gRPC Proto e Domain models

#### Arquivo 1: `internal/infrastructure/grpc/mappers/key_mapper.go`

**Funções** (16 total):
```go
// Proto → Domain
func MapProtoKeyTypeToDomain(kt commonv1.KeyType) domain.KeyType
func MapProtoStatusToDomain(st commonv1.EntryStatus) domain.EntryStatus
func MapProtoAccountToDomain(acc *commonv1.Account) *domain.Account
func MapProtoOwnerToDomain(owner *commonv1.Owner) *domain.Owner

// Domain → Proto
func MapDomainKeyToProto(key *domain.DictKey) *commonv1.DictKey
func MapDomainKeyTypeToProto(kt domain.KeyType) commonv1.KeyType
func MapDomainStatusToProto(st domain.EntryStatus) commonv1.EntryStatus
func MapDomainAccountToProto(acc *domain.Account) *commonv1.Account
func MapDomainOwnerToProto(owner *domain.Owner) *commonv1.Owner
func MapDomainEntryToProtoKeySummary(entry *domain.Entry) *corev1.KeySummary
func MapDomainEntryToProtoGetKeyResponse(entry *domain.Entry) *corev1.GetKeyResponse

// Complex mappings
func MapProtoCreateKeyRequestToCommand(req *corev1.CreateKeyRequest, userID string) application.CreateEntryCommand
func MapProtoStartClaimRequestToCommand(req *corev1.StartClaimRequest, userID string) application.CreateClaimCommand
func MapProtoRespondToClaimRequestToCommand(req *corev1.RespondToClaimRequest, userID string) application.ConfirmClaimCommand or CancelClaimCommand
```

**Estimativa**: 2 horas
**LOC**: ~400 linhas
**Testes**: 16 funções × 2 testes = 32 testes (~200 LOC)

#### Arquivo 2: `internal/infrastructure/grpc/mappers/claim_mapper.go`

**Funções** (12 total):
```go
// Proto → Domain
func MapProtoClaimTypeToDomain(ct commonv1.ClaimType) domain.ClaimType
func MapProtoClaimStatusToDomain(cs commonv1.ClaimStatus) domain.ClaimStatus

// Domain → Proto
func MapDomainClaimToProto(claim *domain.Claim) *corev1.ClaimSummary
func MapDomainClaimToProtoGetClaimStatusResponse(claim *domain.Claim) *corev1.GetClaimStatusResponse
func MapDomainClaimStatusToProto(cs domain.ClaimStatus) commonv1.ClaimStatus
func MapDomainClaimTypeToProto(ct domain.ClaimType) commonv1.ClaimType

// Helpers
func CalculateDaysRemaining(expiresAt time.Time) int32
func FormatClaimMessage(status domain.ClaimStatus) string
```

**Estimativa**: 1 hora
**LOC**: ~300 linhas
**Testes**: 12 funções × 2 testes = 24 testes (~150 LOC)

#### Arquivo 3: `internal/infrastructure/grpc/mappers/error_mapper.go`

**Funções** (10 total):
```go
// Domain Errors → gRPC Status Codes
func MapDomainErrorToGRPC(err error) error {
    switch {
    case errors.Is(err, domain.ErrInvalidKeyType):
        return status.Error(codes.InvalidArgument, err.Error())
    case errors.Is(err, domain.ErrDuplicateKey):
        return status.Error(codes.AlreadyExists, err.Error())
    case errors.Is(err, domain.ErrEntryNotFound):
        return status.Error(codes.NotFound, err.Error())
    case errors.Is(err, domain.ErrUnauthorized):
        return status.Error(codes.PermissionDenied, err.Error())
    case errors.Is(err, domain.ErrMaxKeysExceeded):
        return status.Error(codes.ResourceExhausted, err.Error())
    case errors.Is(err, domain.ErrInvalidStatus):
        return status.Error(codes.FailedPrecondition, err.Error())
    case errors.Is(err, domain.ErrClaimExpired):
        return status.Error(codes.DeadlineExceeded, err.Error())
    default:
        return status.Error(codes.Internal, "Internal server error")
    }
}

// gRPC Status → User-Friendly Message
func FormatUserFriendlyError(code codes.Code, msg string) string
```

**Estimativa**: 1 hora
**LOC**: ~200 linhas
**Testes**: 10 funções × 2 testes = 20 testes (~120 LOC)

**Total Fase 1**: 4 horas, ~900 LOC, 76 testes (~470 LOC)

---

### Fase 2: Injetar Dependencies no Handler (1 hora)

**Objetivo**: Adicionar command/query handlers ao CoreDictServiceHandler struct

#### Arquivo: `internal/infrastructure/grpc/core_dict_service_handler.go`

**Modificação**:
```go
type CoreDictServiceHandler struct {
    corev1.UnimplementedCoreDictServiceServer

    // Command Handlers (10)
    createEntryCmd    *application.CreateEntryCommandHandler
    deleteEntryCmd    *application.DeleteEntryCommandHandler
    startClaimCmd     *application.CreateClaimCommandHandler
    respondClaimCmd   *application.ConfirmClaimCommandHandler  // or CancelClaimCommandHandler
    cancelClaimCmd    *application.CancelClaimCommandHandler
    completeClaimCmd  *application.CompleteClaimCommandHandler
    blockEntryCmd     *application.BlockEntryCommandHandler
    unblockEntryCmd   *application.UnblockEntryCommandHandler
    createInfractCmd  *application.CreateInfractionCommandHandler
    startPortCmd      *application.StartPortabilityCommandHandler
    confirmPortCmd    *application.ConfirmPortabilityCommandHandler
    cancelPortCmd     *application.CancelPortabilityCommandHandler

    // Query Handlers (10)
    getEntryQuery       *application.GetEntryQueryHandler
    listEntriesQuery    *application.ListEntriesQueryHandler
    getClaimQuery       *application.GetClaimQueryHandler
    listClaimsQuery     *application.ListClaimsQueryHandler
    verifyAccountQuery  *application.VerifyAccountQueryHandler
    getStatsQuery       *application.GetStatisticsQueryHandler
    healthCheckQuery    *application.HealthCheckQueryHandler
    getAuditLogQuery    *application.GetAuditLogQueryHandler

    // Services
    keyValidator       *application.KeyValidatorService
    ownershipService   *application.AccountOwnershipService
    duplicateChecker   *application.DuplicateKeyChecker

    // Logger
    logger *slog.Logger
}

// Constructor
func NewCoreDictServiceHandler(
    createEntryCmd *application.CreateEntryCommandHandler,
    deleteEntryCmd *application.DeleteEntryCommandHandler,
    startClaimCmd *application.CreateClaimCommandHandler,
    // ... all 20 handlers
    logger *slog.Logger,
) *CoreDictServiceHandler {
    return &CoreDictServiceHandler{
        createEntryCmd: createEntryCmd,
        deleteEntryCmd: deleteEntryCmd,
        // ... assign all
        logger: logger,
    }
}
```

**Estimativa**: 1 hora
**LOC**: ~100 linhas (constructor + struct fields)

---

### Fase 3: Implementar Integração dos 15 Métodos (8 horas)

**Objetivo**: Conectar cada método gRPC com handlers reais

#### Padrão de Implementação:

```go
func (h *CoreDictServiceHandler) CreateKey(ctx context.Context, req *corev1.CreateKeyRequest) (*corev1.CreateKeyResponse, error) {
    // 1. Extract user_id from context (set by auth interceptor)
    userID, ok := ctx.Value("user_id").(string)
    if !ok || userID == "" {
        return nil, status.Error(codes.Unauthenticated, "user not authenticated")
    }

    // 2. Validate request (already done, keep it)
    if req.GetKeyType() == commonv1.KeyType_KEY_TYPE_UNSPECIFIED {
        return nil, status.Error(codes.InvalidArgument, "key_type is required")
    }

    // 3. Map proto -> domain command
    cmd := mappers.MapProtoCreateKeyRequestToCommand(req, userID)

    // 4. Execute command handler
    entry, err := h.createEntryCmd.Handle(ctx, cmd)
    if err != nil {
        h.logger.Error("Failed to create entry", "error", err, "user_id", userID)
        return nil, mappers.MapDomainErrorToGRPC(err)
    }

    // 5. Map domain -> proto response
    return &corev1.CreateKeyResponse{
        KeyId:     entry.ID,
        Key:       mappers.MapDomainKeyToProto(entry.Key),
        Status:    mappers.MapDomainStatusToProto(entry.Status),
        CreatedAt: timestamppb.New(entry.CreatedAt),
    }, nil
}
```

#### Breakdown por Método (8h total):

| Método | Command/Query | Complexidade | Tempo | LOC |
|--------|---------------|--------------|-------|-----|
| **CreateKey** | CreateEntryCommand | Média | 40min | ~60 |
| **ListKeys** | ListEntriesQuery | Baixa | 30min | ~50 |
| **GetKey** | GetEntryQuery | Baixa | 20min | ~40 |
| **DeleteKey** | DeleteEntryCommand | Baixa | 20min | ~30 |
| **StartClaim** | CreateClaimCommand | Média | 40min | ~60 |
| **GetClaimStatus** | GetClaimQuery | Baixa | 20min | ~40 |
| **ListIncomingClaims** | ListClaimsQuery (filter) | Média | 30min | ~50 |
| **ListOutgoingClaims** | ListClaimsQuery (filter) | Média | 30min | ~50 |
| **RespondToClaim** | ConfirmClaim / CancelClaim | Alta | 50min | ~80 |
| **CancelClaim** | CancelClaimCommand | Baixa | 20min | ~30 |
| **StartPortability** | StartPortabilityCmd | Média | 40min | ~60 |
| **ConfirmPortability** | ConfirmPortabilityCmd | Média | 40min | ~60 |
| **CancelPortability** | CancelPortabilityCmd | Baixa | 20min | ~30 |
| **LookupKey** | GetEntryQuery (public) | Média | 30min | ~50 |
| **HealthCheck** | HealthCheckQuery | Baixa | 10min | ~20 |

**Total**: 8 horas, ~710 LOC

---

### Fase 4: Criar Testes de Handlers (4 horas)

**Objetivo**: Testar integração end-to-end de cada handler

#### Padrão de Teste:

```go
func TestCreateKey_Success(t *testing.T) {
    // Arrange
    mockCmd := new(MockCreateEntryCommandHandler)
    mockCmd.On("Handle", mock.Anything, mock.Anything).Return(&domain.Entry{
        ID: "entry-123",
        KeyType: domain.KeyTypeCPF,
        KeyValue: "12345678900",
        Status: domain.EntryStatusActive,
        CreatedAt: time.Now(),
    }, nil)

    handler := grpc.NewCoreDictServiceHandler(mockCmd, /* ... */)

    req := &corev1.CreateKeyRequest{
        KeyType:  commonv1.KeyType_KEY_TYPE_CPF,
        KeyValue: "12345678900",
        AccountId: "acc-123",
    }

    ctx := context.WithValue(context.Background(), "user_id", "user-456")

    // Act
    resp, err := handler.CreateKey(ctx, req)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, "entry-123", resp.KeyId)
    assert.Equal(t, commonv1.KeyType_KEY_TYPE_CPF, resp.Key.KeyType)
    assert.Equal(t, commonv1.EntryStatus_ENTRY_STATUS_ACTIVE, resp.Status)
    mockCmd.AssertExpectations(t)
}

func TestCreateKey_Unauthorized(t *testing.T) {
    // Test without user_id in context
}

func TestCreateKey_DuplicateKey(t *testing.T) {
    // Test with domain.ErrDuplicateKey
}
```

**Breakdown**:
- 15 handlers × 3 testes cada (success, error, edge case) = **45 testes**
- ~60 LOC por teste
- **Total**: 45 testes, ~2.700 LOC, 4 horas

---

### Fase 5: Integration Testing E2E (2 horas)

**Objetivo**: Testar fluxo completo com todos os layers

#### Teste E2E Example:

```go
func TestE2E_CreateKeyFlow(t *testing.T) {
    // Setup: Real PostgreSQL (testcontainers), Real Redis, Mock Pulsar, Mock Connect
    env := setupE2EEnvironment(t)

    // 1. Start gRPC server
    server := startGRPCServer(env)

    // 2. Create gRPC client
    conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
    require.NoError(t, err)
    client := corev1.NewCoreDictServiceClient(conn)

    // 3. Call CreateKey
    ctx := context.Background()
    ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+testJWT)

    req := &corev1.CreateKeyRequest{
        KeyType:  commonv1.KeyType_KEY_TYPE_CPF,
        KeyValue: "12345678900",
        AccountId: "acc-123",
    }

    resp, err := client.CreateKey(ctx, req)

    // 4. Assert
    require.NoError(t, err)
    assert.NotEmpty(t, resp.KeyId)
    assert.Equal(t, commonv1.EntryStatus_ENTRY_STATUS_ACTIVE, resp.Status)

    // 5. Verify in database
    entry, err := env.EntryRepo.FindByID(ctx, resp.KeyId)
    require.NoError(t, err)
    assert.Equal(t, "12345678900", entry.KeyValue)

    // 6. Verify event published to Pulsar
    assert.Eventually(t, func() bool {
        return env.PulsarMock.ReceivedEvent("dict.entries.created")
    }, 5*time.Second, 100*time.Millisecond)
}
```

**Testes E2E** (10 testes):
1. CreateKey → GetKey
2. CreateKey → ListKeys
3. CreateKey → DeleteKey → GetKey (should 404)
4. CreateKey duplicate → should fail
5. StartClaim → GetClaimStatus
6. StartClaim → RespondToClaim (accept) → CompleteClaimCommand
7. StartClaim → RespondToClaim (reject) → CancelClaimCommand
8. StartClaim → CancelClaim (by claimer)
9. StartPortability → ConfirmPortability
10. LookupKey (public access)

**Total**: 10 testes, ~1.500 LOC, 2 horas

---

### Fase 6: Documentação de API (2 horas)

**Objetivo**: Documentar APIs para Front-End Squad

#### Arquivo 1: `docs/API_GRPC_FRONTEND.md`

**Conteúdo**:
- Descrição de cada RPC
- Request/Response examples
- Error codes
- Rate limiting
- Authentication
- Examples com grpcurl

**Estimativa**: 1 hora, ~1.000 linhas

#### Arquivo 2: `examples/grpcurl_examples.sh`

**Conteúdo**:
```bash
#!/bin/bash

# Health Check
grpcurl -plaintext localhost:9090 dict.core.v1.CoreDictService/HealthCheck

# Create CPF Key
grpcurl -plaintext \
  -H "authorization: Bearer $JWT_TOKEN" \
  -d '{
    "key_type": 1,
    "key_value": "12345678900",
    "account_id": "acc-123"
  }' \
  localhost:9090 dict.core.v1.CoreDictService/CreateKey

# List Keys
grpcurl -plaintext \
  -H "authorization: Bearer $JWT_TOKEN" \
  -d '{
    "page_size": 20
  }' \
  localhost:9090 dict.core.v1.CoreDictService/ListKeys

# ... (15 examples)
```

**Estimativa**: 1 hora, ~300 linhas

---

## 📊 Resumo de Esforço

| Fase | Objetivo | Duração | LOC | Testes |
|------|----------|---------|-----|--------|
| **1. Mappers** | Proto ↔ Domain | 4h | 900 | 76 |
| **2. Dependencies** | Injetar handlers | 1h | 100 | - |
| **3. Integração** | 15 métodos gRPC | 8h | 710 | - |
| **4. Unit Tests** | Testar handlers | 4h | 2.700 | 45 |
| **5. E2E Tests** | Fluxo completo | 2h | 1.500 | 10 |
| **6. Documentação** | Guias + exemplos | 2h | 1.300 | - |
| **TOTAL** | **Core-Dict 100%** | **21h** | **7.210** | **131** |

**Timeline**: **3 dias úteis** (7h/dia)

---

## 🗓️ Cronograma Detalhado

### Segunda-feira (2025-10-28) - 7 horas
**Objetivo**: Mappers + Dependencies

**Manhã** (4h):
- ⏳ 09:00-11:00: `key_mapper.go` (2h)
- ⏳ 11:00-12:00: `claim_mapper.go` (1h)
- ⏳ 12:00-13:00: `error_mapper.go` (1h)

**Tarde** (3h):
- ⏳ 14:00-15:00: Injetar dependencies no handler (1h)
- ⏳ 15:00-17:00: Testes de mappers (2h)

**Entregável**: Mappers completos (76 testes passando)

---

### Terça-feira (2025-10-29) - 8 horas
**Objetivo**: Integração dos 15 Métodos gRPC

**Manhã** (4h):
- ⏳ 09:00-10:00: CreateKey, ListKeys, GetKey, DeleteKey (4 métodos)
- ⏳ 10:00-11:30: StartClaim, GetClaimStatus, ListIncomingClaims, ListOutgoingClaims (4 métodos)
- ⏳ 11:30-13:00: RespondToClaim, CancelClaim (2 métodos complexos)

**Tarde** (4h):
- ⏳ 14:00-15:30: StartPortability, ConfirmPortability, CancelPortability (3 métodos)
- ⏳ 15:30-16:30: LookupKey, HealthCheck (2 métodos)
- ⏳ 16:30-18:00: Build + testes manuais com grpcurl

**Entregável**: 15 handlers integrados e testáveis

---

### Quarta-feira (2025-10-30) - 6 horas
**Objetivo**: Testes + Documentação

**Manhã** (4h):
- ⏳ 09:00-13:00: Unit tests dos 15 handlers (45 testes)

**Tarde** (2h):
- ⏳ 14:00-16:00: E2E tests (10 testes)

**Noite** (2h - opcional):
- ⏳ 18:00-20:00: Documentação (API_GRPC_FRONTEND.md + examples)

**Entregável**: Core-Dict 100% funcional com testes

---

## ✅ Critérios de Aceitação

### Funcional
- ✅ Todos os 15 RPCs gRPC executam business logic real (não mocks)
- ✅ Front-End pode criar chaves PIX (CPF, CNPJ, Email, Phone, EVP)
- ✅ Front-End pode listar, buscar e deletar chaves
- ✅ Front-End pode gerenciar claims (30 dias)
- ✅ Front-End pode fazer portabilidade
- ✅ Front-End pode consultar chaves de terceiros (LookupKey)

### Técnico
- ✅ 131 testes adicionais criados (76 mappers + 45 handlers + 10 E2E)
- ✅ Cobertura de código >85% nas novas implementações
- ✅ Build sem erros: `go build ./...`
- ✅ Todos os testes passando: `go test ./...`
- ✅ gRPC server inicia sem erros
- ✅ Interceptors funcionando (auth, logging, metrics)

### Qualidade
- ✅ Error handling consistente (domain errors → gRPC codes)
- ✅ Logging estruturado (slog)
- ✅ Métricas exportadas (Prometheus)
- ✅ Documentação completa (API guide + examples)

### Performance
- ✅ CreateKey: <50ms (p95)
- ✅ ListKeys: <100ms (p95)
- ✅ GetKey: <30ms (p95) - cache hit
- ✅ LookupKey: <200ms (p95) - via Connect → Bridge → Bacen

---

## 🚀 Após Implementação

### Para o Front-End Squad

**O que terão disponível**:
1. ✅ **15 APIs gRPC** funcionais (não mocks)
2. ✅ **Documentação completa** (API_GRPC_FRONTEND.md)
3. ✅ **Exemplos práticos** (grpcurl_examples.sh)
4. ✅ **Proto contracts** (dict-contracts)
5. ✅ **Servidor gRPC** rodando em `localhost:9090` (dev) ou `core-dict.lbpay.svc.cluster.local:9090` (k8s)

**Como testar**:
```bash
# 1. Iniciar Core-Dict
cd /Users/jose.silva.lb/LBPay/IA_Dict/core-dict
docker-compose up -d  # PostgreSQL, Redis, Pulsar
go run cmd/server/main.go

# 2. Testar CreateKey
grpcurl -plaintext \
  -H "authorization: Bearer $JWT_TOKEN" \
  -d '{"key_type": 1, "key_value": "12345678900", "account_id": "acc-123"}' \
  localhost:9090 dict.core.v1.CoreDictService/CreateKey

# 3. Testar ListKeys
grpcurl -plaintext \
  -H "authorization: Bearer $JWT_TOKEN" \
  -d '{"page_size": 20}' \
  localhost:9090 dict.core.v1.CoreDictService/ListKeys
```

### Próximos Passos (Após Quarta-feira)

1. **Front-End Integration** (Quinta-Sexta)
   - Front-End Squad integra com Core-Dict
   - Testes de integração Front-End ↔ Core-Dict
   - Validação de fluxos de UI

2. **Conn-Dict + Conn-Bridge** (Próxima Semana)
   - Completar implementação dos outros 2 repos
   - Testar fluxo E2E: Core → Connect → Bridge → Bacen

3. **Homologação Bacen** (2 semanas)
   - Certificado ICP-Brasil A3
   - Homologação em ambiente Bacen (https://dict-h.pi.rsfn.net.br:16522)
   - Validação de conformidade

---

## 📞 Comunicação

### Daily Standup (9h diariamente)
- ✅ O que foi feito ontem?
- ✅ O que será feito hoje?
- ⚠️ Há bloqueios?

### Comunicação com Front-End Squad
- **Slack**: #dict-integration
- **Reunião síncrona**: Quarta 16h (demo do Core-Dict funcionando)

---

## 🎯 Metas de Qualidade

| Métrica | Meta | Como Medir |
|---------|------|------------|
| **Cobertura de Código** | >85% | `go test -cover ./internal/infrastructure/grpc/...` |
| **Latência p95** | <100ms | Prometheus metrics `grpc_server_handling_seconds` |
| **Taxa de Erro** | <1% | Prometheus metrics `grpc_server_handled_total{grpc_code!="OK"}` |
| **Build Time** | <2min | `time go build ./...` |
| **Test Time** | <5min | `time go test ./...` |

---

## 📚 Referências

### Documentos Base
1. [API-001: Especificação APIs DICT Bacen](../04_APIs/API-001_Especificacao_APIs_DICT_Bacen.md)
2. [US-001: User Stories - DICT Keys](../01_Requisitos/UserStories/US-001_User_Stories_DICT_Keys.md)
3. [US-002: User Stories - Claims](../01_Requisitos/UserStories/US-002_User_Stories_Claims.md)
4. [FE-001: Component Specifications](../08_Frontend/Componentes/FE-001_Component_Specifications.md)

### Código Fonte
- Proto Contracts: `dict-contracts/proto/core_dict.proto`
- Handler Atual: `core-dict/internal/infrastructure/grpc/core_dict_service_handler.go`
- Application Layer: `core-dict/internal/application/`
- Domain Layer: `core-dict/internal/domain/`

---

**Última Atualização**: 2025-10-27 22:00 BRT
**Status**: ✅ **PLANO APROVADO - INICIAR SEGUNDA-FEIRA**
**Responsável**: Backend Squad + Project Manager
**Revisor**: Tech Lead + Front-End Squad Lead
