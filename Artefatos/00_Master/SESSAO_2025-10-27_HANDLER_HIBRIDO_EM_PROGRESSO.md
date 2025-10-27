# Sessão: Implementação Handler Híbrido (Mock + Real)

**Data**: 2025-10-27 (continuação)
**Duração**: 1h
**Status**: ⏳ EM PROGRESSO - Corrigindo erros de compilação

---

## 🎯 Objetivo

Implementar handler híbrido com feature flag para alternar entre mock e real business logic, permitindo:
- Front-End testar com mock enquanto backend finaliza implementação
- Alternar para real mode quando pronto (via ENV variable)

---

## ✅ Progresso Atual

### 1. Handler Atualizado com Feature Flag ✅

**Arquivo**: `core_dict_service_handler.go`

**Mudanças Implementadas**:

```go
type CoreDictServiceHandler struct {
    corev1.UnimplementedCoreDictServiceServer

    // ========== Feature Flag ==========
    useMockMode bool  // true = mock, false = real

    // ========== Command Handlers ==========
    createEntryCmd   *commands.CreateEntryCommandHandler
    updateEntryCmd   *commands.UpdateEntryCommandHandler
    deleteEntryCmd   *commands.DeleteEntryCommandHandler
    blockEntryCmd    *commands.BlockEntryCommandHandler
    unblockEntryCmd  *commands.UnblockEntryCommandHandler
    createClaimCmd   *commands.CreateClaimCommandHandler
    confirmClaimCmd  *commands.ConfirmClaimCommandHandler
    cancelClaimCmd   *commands.CancelClaimCommandHandler
    completeClaimCmd *commands.CompleteClaimCommandHandler

    // ========== Query Handlers ==========
    getEntryQuery       *queries.GetEntryQueryHandler
    listEntriesQuery    *queries.ListEntriesQueryHandler
    getClaimQuery       *queries.GetClaimQueryHandler
    listClaimsQuery     *queries.ListClaimsQueryHandler
    getAccountQuery     *queries.GetAccountQueryHandler
    verifyAccountQuery  *queries.VerifyAccountQueryHandler
    healthCheckQuery    *queries.HealthCheckQueryHandler
    getStatisticsQuery  *queries.GetStatisticsQueryHandler
    listInfractionsQuery *queries.ListInfractionsQueryHandler
    getAuditLogQuery    *queries.GetAuditLogQueryHandler

    // ========== Logger ==========
    logger *slog.Logger
}
```

**Constructor** com dependency injection completa:
```go
func NewCoreDictServiceHandler(
    useMockMode bool,
    // 9 command handlers
    createEntryCmd *commands.CreateEntryCommandHandler,
    // ...
    // 10 query handlers
    getEntryQuery *queries.GetEntryQueryHandler,
    // ...
    logger *slog.Logger,
) *CoreDictServiceHandler
```

---

### 2. Método CreateKey Híbrido Implementado ✅

**Padrão Implementado**:

```go
func (h *CoreDictServiceHandler) CreateKey(ctx context.Context, req *corev1.CreateKeyRequest) (*corev1.CreateKeyResponse, error) {
    // ========== 1. VALIDATION (sempre, mock ou real) ==========
    if req.GetKeyType() == commonv1.KeyType_KEY_TYPE_UNSPECIFIED {
        return nil, status.Error(codes.InvalidArgument, "key_type is required")
    }
    // ... validações

    // ========== 2. MOCK MODE ==========
    if h.useMockMode {
        h.logger.Info("CreateKey: MOCK MODE", ...)
        return &corev1.CreateKeyResponse{
            KeyId: fmt.Sprintf("mock-key-%d", now.Unix()),
            // ... mock response
        }, nil
    }

    // ========== 3. REAL MODE ==========
    h.logger.Info("CreateKey: REAL MODE", ...)

    // 3a. Extract user_id from context
    userID, ok := ctx.Value("user_id").(string)

    // 3b. Map proto → domain command
    cmd, err := mappers.MapProtoCreateKeyRequestToCommand(req, userID)

    // 3c. Execute command handler
    result, err := h.createEntryCmd.Handle(ctx, cmd)

    // 3d. Map domain result → proto response
    return &corev1.CreateKeyResponse{
        KeyId: result.EntryID.String(),
        // ... real response
    }, nil
}
```

---

### 3. Mapper Helper Adicionado ✅

**Arquivo**: `key_mapper.go`

**Função**:
```go
// MapStringStatusToProto converts string status (from command result) to proto EntryStatus
func MapStringStatusToProto(status string) commonv1.EntryStatus {
	switch status {
	case "pending":
		return commonv1.EntryStatus_ENTRY_STATUS_PENDING
	case "active":
		return commonv1.EntryStatus_ENTRY_STATUS_ACTIVE
	// ... 6 cases
	default:
		return commonv1.EntryStatus_ENTRY_STATUS_UNSPECIFIED
	}
}
```

---

## ⚠️ Problemas Encontrados

### 1. Erros de Compilação nos Mappers

**Arquivo**: `claim_mapper.go`

**Erros**:
```
undefined: commonv1.ClaimType
undefined: commonv1.ClaimType_CLAIM_TYPE_OWNERSHIP
undefined: commonv1.ClaimType_CLAIM_TYPE_PORTABILITY
```

**Causa**: Possível inconsistência entre os mappers e o proto file. Precisa verificar se:
- `ClaimType` existe no proto `common.proto`
- Enums estão corretos

**Arquivo**: `key_mapper.go`

**Erros**:
```
undefined: valueobjects.AccountType
undefined: queries.GetEntryByKeyQuery
```

**Causa**:
- `AccountType` pode não existir no domain
- `GetEntryByKeyQuery` pode não ter sido criado ainda

---

### 2. Imports de Módulo

**Problema inicial**: Imports usando `core-dict/internal/...` em vez de `github.com/lbpay-lab/core-dict/internal/...`

**Solução aplicada**:
```bash
sed -i '' 's|"core-dict/internal/|"github.com/lbpay-lab/core-dict/internal/|g' mappers/*.go
```

**Problema adicional**: Import genérico `application` não funciona (Go tentou buscar no GitHub)

**Solução aplicada**: Trocar para imports específicos:
```go
// ANTES
"github.com/lbpay-lab/core-dict/internal/application"

// DEPOIS
"github.com/lbpay-lab/core-dict/internal/application/commands"
"github.com/lbpay-lab/core-dict/internal/application/queries"
```

---

## 🔧 Próximos Passos (PENDENTE)

### Passo 1: Corrigir Erros de Compilação ⏳

**a) Verificar proto files**:
```bash
grep -r "ClaimType" dict-contracts/proto/
```

**b) Verificar domain types**:
```bash
grep -r "AccountType" core-dict/internal/domain/valueobjects/
```

**c) Verificar queries existentes**:
```bash
ls core-dict/internal/application/queries/ | grep -i entry
```

**d) Corrigir mappers** baseado nos achados

---

### Passo 2: Implementar Restante dos Métodos Híbridos (14 métodos) ⏳

Seguindo o mesmo padrão do CreateKey:

#### Keys (3 restantes)
- [ ] ListKeys (mock + real)
- [ ] GetKey (mock + real)
- [ ] DeleteKey (mock + real)

#### Claims (6 métodos)
- [ ] StartClaim (mock + real)
- [ ] GetClaimStatus (mock + real)
- [ ] ListIncoming/Outgoing (mock + real)
- [ ] RespondToClaim (mock + real)
- [ ] CancelClaim (mock + real)

#### Portability (3 métodos)
- [ ] StartPortability (mock + real)
- [ ] ConfirmPortability (mock + real)
- [ ] CancelPortability (mock + real)

#### Query/Health (2 métodos)
- [ ] LookupKey (mock + real)
- [ ] HealthCheck (mock + real)

**Estimativa**: 30 minutos/método × 14 = 7 horas

---

### Passo 3: Configuração ENV ⏳

**Criar**: `.env.example`
```env
# Feature Flags
CORE_DICT_USE_MOCK_MODE=false  # false = real business logic

# gRPC Server
GRPC_PORT=9090

# Logging
LOG_LEVEL=info
```

**Atualizar**: `cmd/server/main.go`
```go
useMockMode := os.Getenv("CORE_DICT_USE_MOCK_MODE") == "true"

handler := grpc.NewCoreDictServiceHandler(
    useMockMode,
    createEntryCmd,  // injetar todos os handlers
    // ...
    logger,
)
```

**Estimativa**: 30 minutos

---

### Passo 4: Testes ⏳

**Mock Mode**:
```bash
# Configurar
export CORE_DICT_USE_MOCK_MODE=true

# Testar com grpcurl
grpcurl -plaintext -d '{"key_type": "KEY_TYPE_CPF", "key_value": "12345678900", "account_id": "acc-123"}' \
  localhost:9090 core.v1.CoreDictService/CreateKey

# Deve retornar: mock-key-1234567890
```

**Real Mode**:
```bash
# Configurar
export CORE_DICT_USE_MOCK_MODE=false

# Subir infraestrutura
docker-compose up -d postgres redis

# Testar
grpcurl -plaintext -d '{"key_type": "KEY_TYPE_CPF", "key_value": "12345678900", "account_id": "acc-123"}' \
  -H "user_id: user-123" \
  localhost:9090 core.v1.CoreDictService/CreateKey

# Deve executar lógica real e persistir no PostgreSQL
```

**Estimativa**: 1 hora

---

## 📊 Métricas de Progresso

| Item | Status | LOC | Tempo |
|------|--------|-----|-------|
| Handler struct + constructor | ✅ Completo | 98 | 30min |
| CreateKey híbrido | ✅ Completo | 64 | 20min |
| MapStringStatusToProto | ✅ Completo | 18 | 5min |
| Correção imports | ✅ Completo | - | 15min |
| **SUBTOTAL** | **Parcial** | **180** | **1h10min** |
| Correção erros compilação | ⏳ Pendente | - | 30min |
| 14 métodos restantes | ⏳ Pendente | ~900 | 7h |
| Configuração ENV | ⏳ Pendente | ~50 | 30min |
| Testes | ⏳ Pendente | - | 1h |
| **TOTAL RESTANTE** | - | **~950** | **~9h** |

---

## 🎯 Timeline Atualizado

### Sessão Atual (2025-10-27)
- ✅ 00:00-01:10: Handler + CreateKey + fixes
- ⏳ 01:10-02:00: Corrigir erros compilação (PRÓXIMO)

### Segunda-feira (Restante: 5h)
- ⏳ 14:00-18:00: Implementar 7 métodos (Keys + Claims) - 4h
- ⏳ 18:00-19:00: Implementar 4 métodos (Portability + Query) - 1h

### Terça-feira (4h)
- ⏳ 09:00-10:00: Implementar 3 métodos restantes - 1h
- ⏳ 10:00-11:00: Configuração ENV + main.go - 1h
- ⏳ 11:00-13:00: Testes (mock + real) - 2h

**Total previsto**: 11 horas (~1.5 dias)

---

## 🚀 Vantagens da Abordagem Híbrida

### Para Front-End
✅ Pode começar **HOJE** com `MOCK_MODE=true`
✅ Testa estrutura Request/Response sem backend real
✅ Valida paginação, filtros, error handling
✅ Desenvolve UI sem bloqueio

### Para Backend
✅ Implementa lógica real **em paralelo** ao Front-End
✅ Testa ambos os modos (mock + real)
✅ Feature flag fácil (1 env var)
✅ Rollback rápido se bugs aparecerem

### Para Testes
✅ Unit tests usam mock mode
✅ Integration tests usam real mode
✅ E2E tests podem testar ambos

---

## 📝 Notas Técnicas

### Padrão de Validação
- **Sempre** valida requests (mesmo em mock mode)
- Validações incluem: campos required, tipos corretos, limites

### Extração de user_id
- Real mode: extrai `user_id` do context (setado por auth interceptor)
- Mock mode: não precisa de autenticação

### Error Handling
- Real mode: usa `mappers.MapDomainErrorToGRPC(err)` para converter domain errors em gRPC status codes
- Mock mode: retorna sucesso sempre (sem errors)

### Logging
- Ambos os modos logam: `logger.Info("CreateKey: MOCK/REAL MODE", ...)`
- Real mode loga adicionalmente: erros, user_id, entry_id

---

## 🐛 Bugs/Blockers

### Blocker 1: ClaimType undefined
**Impacto**: Alto - impede compilação de claim_mapper.go
**Próxima ação**: Verificar proto files para confirmar se ClaimType existe

### Blocker 2: AccountType undefined
**Impacto**: Médio - impede compilação de key_mapper.go
**Próxima ação**: Verificar domain valueobjects

### Blocker 3: GetEntryByKeyQuery undefined
**Impacto**: Médio - impede LookupKey implementation
**Próxima ação**: Criar query se não existir

---

## 🔍 Erros Identificados e Próximas Correções

### 1. ClaimType removido ✅
**Problema**: claim_mapper.go usava `commonv1.ClaimType` que não existe
**Causa**: ClaimType só existe em conn_dict events, não em common.proto
**Solução aplicada**: Removidas funções `MapProtoClaimTypeToDomain` e `MapDomainClaimTypeToProto`

### 2. AccountType undefined ⏳
**Erro**: `undefined: valueobjects.AccountType`
**Verificar**: Se AccountType existe no domain ou apenas no proto

### 3. GetEntryByKeyQuery undefined ⏳
**Erro**: `undefined: queries.GetEntryByKeyQuery`
**Solução**: Criar query ou usar GetEntryQuery com KeyValue

### 4. Commands/Queries struct mismatch ⏳
**Erros**:
```
unknown field UserID in struct literal of type commands.CreateClaimCommand
unknown field KeyType in struct literal of type commands.CreateClaimCommand
cannot use req.GetAccountId() (value of type string) as uuid.UUID value
```

**Causa**: Os mappers foram escritos assumindo estrutura de Commands/Queries, mas não verificamos a estrutura real

**Solução necessária**:
1. Ler cada Command/Query struct
2. Ajustar mappers para usar os campos corretos
3. Fazer conversões string → uuid.UUID onde necessário

---

## 📋 Plano de Ação (Próxima Sessão)

### Passo 1: Ler estruturas reais ⏳

```bash
# Ler Commands
cat internal/application/commands/create_claim_command.go
cat internal/application/commands/confirm_claim_command.go
cat internal/application/commands/cancel_claim_command.go

# Ler Queries
ls internal/application/queries/ | grep -i entry
```

### Passo 2: Ajustar mappers ⏳

Para cada mapper, verificar:
- Campos corretos nos structs
- Conversões de tipos (string → uuid, etc.)
- Imports necessários

### Passo 3: Testar compilação ⏳

```bash
go build ./internal/infrastructure/grpc/...
```

### Passo 4: Continuar implementação dos 14 métodos restantes ⏳

---

**Última Atualização**: 2025-10-27 02:00 BRT
**Próxima Ação**: Ler estruturas reais de Commands/Queries e ajustar mappers
**Status**: ⏳ **EM PROGRESSO - HANDLER ESTRUTURA PRONTA, MAPPERS PRECISAM AJUSTES**
**Tempo investido**: 1h50min
**Estimativa para completar**: 8-10h
