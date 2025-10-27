# Guia de Refatoração: Unificação de Interfaces - Commands Layer

**Data**: 2025-10-27
**Status**: Em Progresso
**Arquivos Completos**: 3/9

---

## Padrão de Refatoração

### 1. Imports Obrigatórios

Adicionar no topo de TODOS os arquivos de comando:

```go
import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/repositories"
	"github.com/lbpay-lab/core-dict/internal/application/services"
	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects" // Apenas se usar Claim
)
```

### 2. Substituições Globais

| **Antes (Commands Layer)** | **Depois (Domain Layer)** |
|----------------------------|---------------------------|
| `KeyType` | `entities.KeyType` |
| `*Entry` | `*entities.Entry` |
| `*Claim` | `*entities.Claim` |
| `ClaimType` | `valueobjects.ClaimType` |
| `EntryRepository` | `repositories.EntryRepository` |
| `ClaimRepository` | `repositories.ClaimRepository` |
| `AccountRepository` | `repositories.AccountRepository` |
| `CacheService` | `services.CacheService` |
| `ConnectClient` | `services.ConnectClient` |
| `entry.Status = "PENDING"` | `entry.Status = entities.KeyStatusPending` |
| `entry.Status = "ACTIVE"` | `entry.Status = entities.KeyStatusActive` |
| `entry.Status = "BLOCKED"` | `entry.Status = entities.KeyStatusBlocked` |
| `entry.Status = "DELETED"` | `entry.Status = entities.KeyStatusDeleted` |
| `claim.Status = "PENDING"` | `claim.Status = valueobjects.ClaimStatusPending` |

### 3. Estrutura de Entry

**ANTES (commands.Entry - NESTED)**:
```go
entry := &Entry{
    ID:        uuid.New(),
    KeyType:   cmd.KeyType,
    KeyValue:  cmd.KeyValue,
    Status:    "PENDING",
    AccountID: cmd.AccountID,
    Account: Account{
        ISPB:          cmd.AccountISPB,
        Branch:        cmd.AccountBranch,
        AccountNumber: cmd.AccountNumber,
        AccountType:   cmd.AccountType,
    },
    Owner: Owner{
        Type:  cmd.OwnerType,
        TaxID: cmd.OwnerTaxID,
        Name:  cmd.OwnerName,
    },
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}
```

**DEPOIS (entities.Entry - FLAT)**:
```go
now := time.Now()
entry := &entities.Entry{
    ID:            uuid.New(),
    KeyType:       cmd.KeyType,
    KeyValue:      cmd.KeyValue,
    Status:        entities.KeyStatusPending,
    AccountID:     cmd.AccountID,
    ISPB:          cmd.AccountISPB,
    Branch:        cmd.AccountBranch,
    AccountNumber: cmd.AccountNumber,
    AccountType:   cmd.AccountType,
    OwnerName:     cmd.OwnerName,
    OwnerTaxID:    cmd.OwnerTaxID,
    OwnerType:     cmd.OwnerType,
    CreatedAt:     now,
    UpdatedAt:     now,
}
```

### 4. Estrutura de Claim

**ANTES (commands.Claim)**:
```go
claim := &Claim{
    ID:            uuid.New(),
    EntryID:       entry.ID,
    Type:          cmd.ClaimType,
    Status:        "PENDING",
    ClaimerISPB:   cmd.ClaimerISPB,
    ClaimedISPB:   cmd.ClaimedISPB,
    NewAccountID:  cmd.AccountID,
    BacenClaimID:  cmd.BacenClaimID,
    RequestedAt:   now,
    DeadlineAt:    deadline,
    CreatedAt:     now,
    UpdatedAt:     now,
}
```

**DEPOIS (entities.Claim)**:
```go
now := time.Now()
deadline := now.Add(7 * 24 * time.Hour)

claim := &entities.Claim{
    ID:                    uuid.New(),
    ClaimType:             cmd.ClaimType, // já é valueobjects.ClaimType
    Status:                valueobjects.ClaimStatusPending,
    ClaimerISPB:           cmd.ClaimerISPB,
    OwnerISPB:             cmd.ClaimedISPB,
    ClaimerAccountID:      &cmd.AccountID,
    OwnerAccountID:        &entry.AccountID,
    BacenClaimID:          &cmd.BacenClaimID,
    CompletionPeriodDays:  7,
    ExpiresAt:             &deadline,
    EntryKey:              cmd.KeyValue,
    CreatedAt:             now,
    UpdatedAt:             now,
}
```

### 5. CacheService: Métodos Padronizados

**Substituir**:
```go
h.cacheService.InvalidateKey(ctx, "entry:"+keyValue)
```

**Por**:
```go
h.cacheService.Delete(ctx, "entry:"+keyValue)
```

**Substituir**:
```go
h.cacheService.InvalidatePattern(ctx, "entries:*")
```

**Por**:
```go
h.cacheService.Invalidate(ctx, "entries:*")
```

### 6. EntryRepository: Métodos

**Substituir**:
```go
entry, err := h.entryRepo.FindByKeyValue(ctx, keyValue)
```

**Por**:
```go
entry, err := h.entryRepo.FindByKey(ctx, keyValue)
```

### 7. Acessar campos de Entry

**ANTES (nested)**:
```go
entry.Account.ISPB
entry.Account.Branch
entry.Account.AccountNumber
entry.Owner.Name
entry.Owner.TaxID
```

**DEPOIS (flat)**:
```go
entry.ISPB
entry.Branch
entry.AccountNumber
entry.OwnerName
entry.OwnerTaxID
```

### 8. Remover Interfaces Duplicadas

**REMOVER do final dos arquivos**:
```go
// Temporary interfaces
type Entry struct { ... }
type Account struct { ... }
type Owner struct { ... }
type Claim struct { ... }
type EntryRepository interface { ... }
type ClaimRepository interface { ... }
type CacheService interface { ... }
type ConnectClient interface { ... }
```

**MANTER apenas**:
```go
// DomainEvent (ainda não movido para Domain Layer)
type DomainEvent struct {
    EventType     string
    AggregateID   string
    AggregateType string
    OccurredAt    time.Time
    Payload       map[string]interface{}
}

// Service interfaces específicas do Application Layer
type EventPublisher interface {
    Publish(ctx context.Context, event DomainEvent) error
}

type KeyValidatorService interface {
    ValidateFormat(keyType entities.KeyType, keyValue string) error
    ValidateLimits(ctx context.Context, keyType entities.KeyType, ownerTaxID string) error
}

type OwnershipService interface {
    ValidateOwnership(ctx context.Context, keyType entities.KeyType, keyValue, ownerTaxID string) error
}

type DuplicateCheckerService interface {
    IsDuplicate(ctx context.Context, keyValue string) (bool, error)
}

type EntryEventProducer interface {
    PublishCreated(ctx context.Context, entry interface{}, userID string) error
    PublishUpdated(ctx context.Context, entry interface{}, userID string) error
    PublishDeleted(ctx context.Context, entryID, keyValue, reason, userID string) error
}
```

---

## Arquivos que PRECISAM ser Refatorados

### ✅ Completos (3)
1. ✅ `create_entry_command.go` - Completo
2. ✅ `create_claim_command.go` - Completo
3. ✅ `delete_entry_command.go` - Completo

### 🔄 Pendentes (6)
4. ⏳ `update_entry_command.go`
5. ⏳ `block_entry_command.go`
6. ⏳ `unblock_entry_command.go`
7. ⏳ `confirm_claim_command.go`
8. ⏳ `cancel_claim_command.go`
9. ⏳ `complete_claim_command.go`

### ⚠️ Atenção Especial
10. ⚠️ `create_infraction_command.go` - Requer entities.Infraction (verificar se existe)

---

## Checklist de Refatoração por Arquivo

Para cada arquivo, seguir:

- [ ] 1. Adicionar imports corretos (entities, repositories, services, valueobjects)
- [ ] 2. Substituir tipos de comando (KeyType → entities.KeyType)
- [ ] 3. Atualizar struct do Handler (usar repositories.*, services.*)
- [ ] 4. Atualizar constructor do Handler
- [ ] 5. Substituir criação de entidades (usar entities.Entry/Claim)
- [ ] 6. Corrigir acessos a campos (nested → flat)
- [ ] 7. Substituir métodos de repositório (FindByKeyValue → FindByKey)
- [ ] 8. Substituir métodos de cache (InvalidateKey → Delete)
- [ ] 9. Substituir status constants (string → entities.KeyStatus/valueobjects.ClaimStatus)
- [ ] 10. Remover interfaces duplicadas do final do arquivo
- [ ] 11. Compilar e verificar erros

---

## Comando de Verificação

Após refatorar todos os arquivos:

```bash
cd /Users/jose.silva.lb/LBPay/IA_Dict/core-dict
go build ./internal/application/commands/...
```

**Esperado**: Compilação 100% sucesso, 0 erros.

---

## Próximos Passos Após Commands

1. **Implementar métodos faltantes na Infrastructure Layer**:
   - `EntryRepository.CountByOwnerAndType()`
   - `ClaimRepository.FindActiveByEntryID()`

2. **Atualizar cmd/grpc/real_handler_init.go**:
   - Descomentar inicialização de handlers
   - Passar repositories corretos
   - Testar compilação completa

3. **Executar testes unitários**:
   - Ajustar mocks
   - Atualizar testes para usar entities.Entry/Claim

---

**Autor**: Interface Unification Specialist
**Última Atualização**: 2025-10-27
