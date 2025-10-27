# Sessão: Implementação de Mappers Proto ↔ Domain

**Data**: 2025-10-27
**Duração**: 30 minutos
**Status**: ✅ Mappers Implementados

---

## 🎯 Objetivo

Implementar mappers para converter entre gRPC Proto messages e Domain models, permitindo integração híbrida (mock + real) nos handlers.

---

## ✅ Arquivos Criados

### 1. `key_mapper.go` (269 linhas)

**Funções implementadas** (16):

#### Proto → Domain (5)
- `MapProtoKeyTypeToDomain()` - CPF, CNPJ, Email, Phone, EVP
- `MapProtoStatusToDomain()` - Pending, Active, Blocked, Deleted, ClaimPending, Failed
- `MapProtoAccountTypeToDomain()` - Checking, Savings, Payment
- `MapProtoAccountToDomain()` - Account struct completo
- `MapProtoLookupKeyRequestToQuery()` - Lookup query

#### Domain → Proto (5)
- `MapDomainKeyTypeToProto()`
- `MapDomainStatusToProto()`
- `MapDomainAccountTypeToProto()`
- `MapDomainAccountToProto()`
- `MapDomainEntryToProtoKeySummary()` - Para ListKeys

#### Complex Mappings (4)
- `MapProtoCreateKeyRequestToCommand()` - CreateEntryCommand
- `MapProtoListKeysRequestToQuery()` - ListEntriesQuery (com filtros)
- `MapProtoDeleteKeyRequestToCommand()` - DeleteEntryCommand
- `MapDomainEntryToProtoGetKeyResponse()` - GetKey response completo

#### Helpers (2)
- `TimeToTimestamppb()` - Time → protobuf Timestamp
- Validação de page_size (default 20, max 100)

---

### 2. `claim_mapper.go` (305 linhas)

**Funções implementadas** (15):

#### Proto → Domain (2)
- `MapProtoClaimTypeToDomain()` - Ownership, Portability
- `MapProtoClaimStatusToDomain()` - Open, WaitingResolution, Confirmed, Cancelled, Completed

#### Domain → Proto (2)
- `MapDomainClaimTypeToProto()`
- `MapDomainClaimStatusToProto()`

#### Request → Command (5)
- `MapProtoStartClaimRequestToCommand()` - CreateClaimCommand
- `MapProtoRespondToClaimRequestToConfirmCommand()` - ConfirmClaimCommand
- `MapProtoRespondToClaimRequestToCancelCommand()` - CancelClaimCommand
- `MapProtoCancelClaimRequestToCommand()` - CancelClaimCommand
- `MapProtoListIncomingClaimsRequestToQuery()` - direction="incoming"
- `MapProtoListOutgoingClaimsRequestToQuery()` - direction="outgoing"

#### Domain → Response (3)
- `MapDomainClaimToProtoSummary()` - ClaimSummary para listas
- `MapDomainClaimToProtoGetClaimStatusResponse()` - GetClaimStatus completo
- `MapDomainClaimToProtoStartClaimResponse()` - StartClaim response
- `MapDomainClaimToProtoRespondToClaimResponse()` - RespondToClaim response

#### Helpers (3)
- `CalculateDaysRemaining()` - Calcula dias até expiração (30 dias)
- `FormatClaimMessage()` - Mensagem user-friendly por status
- `FormatClaimResponseMessage()` - Mensagem após aceitar/rejeitar

---

### 3. `error_mapper.go` (127 linhas)

**Funções implementadas** (3):

#### Domain Errors → gRPC (1)
`MapDomainErrorToGRPC()` - Mapeia todos os domain errors:

| Domain Error | gRPC Code | User Message |
|--------------|-----------|--------------|
| `ErrInvalidKeyType` | InvalidArgument | "Invalid key type..." |
| `ErrInvalidKeyValue` | InvalidArgument | "Invalid key value..." |
| `ErrEntryNotFound` | NotFound | "Entry not found..." |
| `ErrClaimNotFound` | NotFound | "Claim not found..." |
| `ErrDuplicateKey` | AlreadyExists | "Key already registered. You may initiate a portability claim..." |
| `ErrDuplicateKeyGlobal` | AlreadyExists | "Key already registered in RSFN..." |
| `ErrUnauthorized` | PermissionDenied | "Unauthorized..." |
| `ErrNotOwner` | PermissionDenied | "You are not the owner..." |
| `ErrMaxKeysExceeded` | ResourceExhausted | "Maximum number of keys exceeded..." |
| `ErrCannotDeleteActiveKey` | FailedPrecondition | "Cannot delete active key..." |
| `ErrClaimExpired` | DeadlineExceeded | "Claim has expired (>30 days)..." |
| default | Internal | "Internal server error..." |

#### gRPC → User-Friendly (1)
`FormatUserFriendlyError()` - Formata mensagens amigáveis

#### Context Errors (1)
`MapContextError()` - Timeout e cancelamento

---

## 📊 Métricas

| Arquivo | LOC | Funções | Tipos Mapeados |
|---------|-----|---------|----------------|
| `key_mapper.go` | 269 | 16 | KeyType, KeyStatus, AccountType, Account, Entry |
| `claim_mapper.go` | 305 | 15 | ClaimType, ClaimStatus, Claim |
| `error_mapper.go` | 127 | 3 | 12 domain errors |
| **TOTAL** | **701** | **34** | **20** |

---

## 🎯 Cobertura de Funcionalidades

### Keys (100%) ✅
- ✅ CreateKey: Proto request → CreateEntryCommand
- ✅ ListKeys: Proto request → ListEntriesQuery (com filtros)
- ✅ GetKey: Domain Entry → Proto GetKeyResponse
- ✅ DeleteKey: Proto request → DeleteEntryCommand
- ✅ LookupKey: Proto request → GetEntryByKeyQuery

### Claims (100%) ✅
- ✅ StartClaim: Proto request → CreateClaimCommand
- ✅ GetClaimStatus: Domain Claim → Proto GetClaimStatusResponse
- ✅ ListIncoming/Outgoing: Proto request → ListClaimsQuery (direction)
- ✅ RespondToClaim: Proto request → Confirm/CancelClaimCommand
- ✅ CancelClaim: Proto request → CancelClaimCommand

### Portability (0%) ⚠️
- ⏳ StartPortability: Não implementado (TODO)
- ⏳ ConfirmPortability: Não implementado (TODO)
- ⏳ CancelPortability: Não implementado (TODO)

**Nota**: Portability mappers serão implementados quando portability commands forem criados na Application Layer.

### Error Handling (100%) ✅
- ✅ 12 domain errors mapeados
- ✅ User-friendly messages
- ✅ Context errors (timeout, cancel)

---

## 🚀 Próximos Passos

### Fase 1: Atualizar Handler (AGORA) ⏳

**Objetivo**: Adicionar feature flag e dependencies injection

**Arquivo**: `core_dict_service_handler.go`

**Mudanças**:
```go
type CoreDictServiceHandler struct {
    corev1.UnimplementedCoreDictServiceServer

    // ========== Feature Flag ==========
    useMockMode bool  // true = mock (para testes Front-End), false = real

    // ========== Command Handlers ==========
    createEntryCmd    *application.CreateEntryCommandHandler
    deleteEntryCmd    *application.DeleteEntryCommandHandler
    startClaimCmd     *application.CreateClaimCommandHandler
    confirmClaimCmd   *application.ConfirmClaimCommandHandler
    cancelClaimCmd    *application.CancelClaimCommandHandler
    // ... (mais 7 command handlers)

    // ========== Query Handlers ==========
    getEntryQuery     *application.GetEntryQueryHandler
    listEntriesQuery  *application.ListEntriesQueryHandler
    getClaimQuery     *application.GetClaimQueryHandler
    listClaimsQuery   *application.ListClaimsQueryHandler
    // ... (mais 6 query handlers)

    // ========== Logger ==========
    logger *slog.Logger
}
```

**Constructor**:
```go
func NewCoreDictServiceHandler(
    useMockMode bool,  // Feature flag
    createEntryCmd *application.CreateEntryCommandHandler,
    // ... all handlers
    logger *slog.Logger,
) *CoreDictServiceHandler {
    return &CoreDictServiceHandler{
        useMockMode: useMockMode,
        createEntryCmd: createEntryCmd,
        // ...
        logger: logger,
    }
}
```

**Estimativa**: 1 hora

---

### Fase 2: Implementar Método CreateKey (EXEMPLO)

**Padrão Híbrido**:
```go
func (h *CoreDictServiceHandler) CreateKey(ctx context.Context, req *corev1.CreateKeyRequest) (*corev1.CreateKeyResponse, error) {
    // 1. Validação (sempre, mock ou real)
    if req.GetKeyType() == commonv1.KeyType_KEY_TYPE_UNSPECIFIED {
        return nil, status.Error(codes.InvalidArgument, "key_type is required")
    }

    // 2. MOCK MODE (para testes Front-End)
    if h.useMockMode {
        h.logger.Info("CreateKey: MOCK MODE")
        return &corev1.CreateKeyResponse{
            KeyId: fmt.Sprintf("mock-key-%d", time.Now().Unix()),
            Key: &commonv1.DictKey{
                KeyType:  req.GetKeyType(),
                KeyValue: req.GetKeyValue(),
            },
            Status:    commonv1.EntryStatus_ENTRY_STATUS_ACTIVE,
            CreatedAt: timestamppb.Now(),
        }, nil
    }

    // 3. REAL MODE (lógica de negócio completa)
    h.logger.Info("CreateKey: REAL MODE", "key_type", req.GetKeyType(), "key_value", req.GetKeyValue())

    // 3a. Extract user_id from context
    userID, ok := ctx.Value("user_id").(string)
    if !ok || userID == "" {
        return nil, status.Error(codes.Unauthenticated, "user not authenticated")
    }

    // 3b. Map proto → domain command
    cmd := mappers.MapProtoCreateKeyRequestToCommand(req, userID)

    // 3c. Execute command handler
    entry, err := h.createEntryCmd.Handle(ctx, cmd)
    if err != nil {
        h.logger.Error("CreateKey failed", "error", err, "user_id", userID)
        return nil, mappers.MapDomainErrorToGRPC(err)
    }

    // 3d. Map domain → proto response
    return &corev1.CreateKeyResponse{
        KeyId:     entry.ID,
        Key:       mappers.MapDomainKeyToProto(&domain.DictKey{KeyType: entry.KeyType, KeyValue: entry.KeyValue}),
        Status:    mappers.MapDomainStatusToProto(entry.Status),
        CreatedAt: timestamppb.New(entry.CreatedAt),
    }, nil
}
```

**Estimativa por método**: 30-40 min
**Total 15 métodos**: 8 horas

---

### Fase 3: Configuração via ENV

**Arquivo**: `.env` ou `config.yaml`

```env
# Feature Flags
CORE_DICT_USE_MOCK_MODE=false  # false = real business logic

# gRPC Server
GRPC_PORT=9090
GRPC_MAX_CONCURRENT=1000

# Logging
LOG_LEVEL=info
```

**Código**:
```go
// cmd/server/main.go
useMockMode := os.Getenv("CORE_DICT_USE_MOCK_MODE") == "true"

handler := grpc.NewCoreDictServiceHandler(
    useMockMode,
    createEntryCmd,
    // ...
    logger,
)
```

---

## ✅ Vantagens da Abordagem Híbrida

### Para Front-End Squad
1. ✅ **Pode começar hoje** com `MOCK_MODE=true`
2. ✅ **Testa estrutura** de Request/Response
3. ✅ **Valida paginação**, filtros, error handling
4. ✅ **Desenvolve UI** sem depender de backend real

### Para Backend Squad
1. ✅ **Implementa lógica real** sem bloquear Front-End
2. ✅ **Testa ambos os modos** (mock + real)
3. ✅ **Feature flag fácil** (1 env var)
4. ✅ **Rollback rápido** se bugs aparecerem

### Para Testes
1. ✅ **Unit tests** usam mock mode
2. ✅ **Integration tests** usam real mode
3. ✅ **E2E tests** podem testar ambos

---

## 📋 Checklist Atualizado

| Item | Status | Tempo |
|------|--------|-------|
| ✅ Mappers Proto ↔ Domain | Completo | 30min |
| ⏳ Feature flag + dependencies | Próximo | 1h |
| ⏳ Implementar 15 métodos híbridos | Próximo | 8h |
| ⏳ Testar mock mode | Próximo | 30min |
| ⏳ Testar real mode | Próximo | 1h |
| ⏳ Documentação | Próximo | 1h |

**Total restante**: ~12 horas (1.5 dias)

---

## 🎯 Timeline Ajustado

### Segunda-feira (Restante: 6.5h)
**Agora**: Finalizar mappers + feature flag (1.5h)
**14:00-18:00**: Implementar 5 métodos (CreateKey, ListKeys, GetKey, DeleteKey, HealthCheck) - 4h
**18:00-19:00**: Testar mock mode - 1h

### Terça-feira (8h)
**09:00-13:00**: Implementar 7 métodos de Claims - 4h
**14:00-18:00**: Implementar 3 métodos restantes + Portability - 4h

### Quarta-feira (3h)
**09:00-10:00**: Testar real mode - 1h
**10:00-12:00**: Documentação + ajustes - 2h

**Total**: 17.5 horas (~2.5 dias)

---

**Última Atualização**: 2025-10-27 23:30 BRT
**Próximo Step**: Atualizar `core_dict_service_handler.go` com feature flag
**Status**: ✅ **MAPPERS COMPLETOS - PRONTO PARA INTEGRAÇÃO**
