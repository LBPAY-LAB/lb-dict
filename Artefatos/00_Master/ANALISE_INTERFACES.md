# Análise de Incompatibilidades de Interfaces - Core DICT

**Data**: 2025-10-27
**Status**: Análise Completa
**Responsável**: Interface Unification Specialist

---

## Executive Summary

O código Real Mode está **95% pronto** mas completamente bloqueado devido a **incompatibilidades de interfaces** entre as 3 camadas da Clean Architecture. As implementações de repositórios e services estão funcionais, mas não podem ser conectadas aos handlers devido a diferenças nos tipos e assinaturas de métodos.

**Impacto**: 0 de 19 handlers funcionais (9 commands + 10 queries).

---

## Análise Detalhada das Incompatibilidades

### 1. Entry Repository

#### Domain Layer (`internal/domain/repositories/entry_repository.go`)
```go
type EntryRepository interface {
    Create(ctx context.Context, entry *entities.Entry) error
    FindByKey(ctx context.Context, keyValue string) (*entities.Entry, error)
    FindByID(ctx context.Context, id uuid.UUID) (*entities.Entry, error)
    Update(ctx context.Context, entry *entities.Entry) error
    Delete(ctx context.Context, entryID uuid.UUID) error
    UpdateStatus(ctx context.Context, entryID uuid.UUID, status entities.KeyStatus) error
    List(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*entities.Entry, error)
    CountByAccount(ctx context.Context, accountID uuid.UUID) (int64, error)
}
```

#### Application/Commands Layer (`internal/application/commands/create_entry_command.go`)
```go
type EntryRepository interface {
    Create(ctx context.Context, entry *Entry) error
    FindByID(ctx context.Context, id uuid.UUID) (*Entry, error)
    FindByKeyValue(ctx context.Context, keyValue string) (*Entry, error)
    Update(ctx context.Context, entry *Entry) error
    Delete(ctx context.Context, id uuid.UUID) error
    CountByOwnerAndType(ctx context.Context, ownerTaxID string, keyType KeyType) (int, error)
}

// Note: commands.Entry é DIFERENTE de entities.Entry
type Entry struct {
    ID        uuid.UUID
    KeyType   KeyType
    KeyValue  string
    Status    string
    AccountID uuid.UUID
    Account   Account  // commands.Account (nested struct)
    Owner     Owner    // commands.Owner (nested struct)
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**Incompatibilidades**:
1. ❌ **Tipo Entry**: `*entities.Entry` vs `*commands.Entry`
2. ❌ **Nome método**: `FindByKey()` vs `FindByKeyValue()`
3. ❌ **Método ausente**: `CountByOwnerAndType()` não existe no Domain
4. ❌ **Estruturas aninhadas**: commands.Entry tem structs aninhados, entities.Entry tem UUIDs

---

### 2. Claim Repository

#### Domain Layer (`internal/domain/repositories/claim_repository.go`)
```go
type ClaimRepository interface {
    Create(ctx context.Context, claim *entities.Claim) error
    Update(ctx context.Context, claim *entities.Claim) error
    FindByID(ctx context.Context, claimID uuid.UUID) (*entities.Claim, error)
    FindByEntryKey(ctx context.Context, entryKey string) ([]*entities.Claim, error)
    FindByStatus(ctx context.Context, status valueobjects.ClaimStatus, limit, offset int) ([]*entities.Claim, error)
    FindExpired(ctx context.Context, limit int) ([]*entities.Claim, error)
    FindByWorkflowID(ctx context.Context, workflowID string) (*entities.Claim, error)
    ExistsActiveClaim(ctx context.Context, entryKey string) (bool, error)
    List(ctx context.Context, filters ClaimFilters) ([]*entities.Claim, error)
}
```

#### Application/Commands Layer (`internal/application/commands/create_claim_command.go`)
```go
type ClaimRepository interface {
    Create(ctx context.Context, claim *Claim) error
    FindByID(ctx context.Context, id uuid.UUID) (*Claim, error)
    FindActiveByEntryID(ctx context.Context, entryID uuid.UUID) (*Claim, error)
    Update(ctx context.Context, claim *Claim) error
    ListPendingClaims(ctx context.Context, limit int) ([]*Claim, error)
}

// Note: commands.Claim é DIFERENTE de entities.Claim
type Claim struct {
    ID             uuid.UUID
    EntryID        uuid.UUID
    Type           ClaimType  // commands.ClaimType (string)
    Status         string
    ClaimerISPB    string
    ClaimedISPB    string
    NewAccountID   uuid.UUID
    BacenClaimID   string
    RequestedAt    time.Time
    DeadlineAt     time.Time
    ResolvedAt     *time.Time
    ResolutionNote string
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

**Incompatibilidades**:
1. ❌ **Tipo Claim**: `*entities.Claim` vs `*commands.Claim`
2. ❌ **Método ausente**: `FindActiveByEntryID()` não existe no Domain
3. ❌ **Método ausente**: `ListPendingClaims()` não existe no Domain
4. ❌ **ClaimType**: `valueobjects.ClaimType` vs `commands.ClaimType` (string)

---

### 3. Account Repository

#### Domain Layer (`internal/domain/repositories/account_repository.go`)
```go
type AccountRepository interface {
    Create(ctx context.Context, account *entities.Account) error
    Update(ctx context.Context, account *entities.Account) error
    FindByID(ctx context.Context, accountID uuid.UUID) (*entities.Account, error)
    FindByAccountNumber(ctx context.Context, ispb, branch, accountNumber string) (*entities.Account, error)
    FindByOwnerTaxID(ctx context.Context, taxID string) ([]*entities.Account, error)
    FindByISPB(ctx context.Context, ispb string, limit, offset int) ([]*entities.Account, error)
    ExistsByAccountNumber(ctx context.Context, ispb, branch, accountNumber string) (bool, error)
    List(ctx context.Context, filters AccountFilters) ([]*entities.Account, error)
    Count(ctx context.Context, filters AccountFilters) (int64, error)
}
```

**Commands Layer não define AccountRepository** → Depende diretamente de Domain (✅ correto!)

---

### 4. Cache Service

#### Application/Services Layer (`internal/application/services/cache_service.go`)
```go
type CacheService interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    Invalidate(ctx context.Context, pattern string) error
}
```

#### Application/Commands Layer (`internal/application/commands/create_entry_command.go`)
```go
type CacheService interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    InvalidateKey(ctx context.Context, key string) error
    InvalidatePattern(ctx context.Context, pattern string) error
}
```

**Incompatibilidades**:
1. ❌ **Método ausente**: `InvalidateKey()` vs `Delete()`
2. ❌ **Método ausente**: `Invalidate()` vs `InvalidatePattern()`

---

### 5. Query Handlers (Queries Layer)

#### Get Entry Query (`internal/application/queries/get_entry_query.go`)
```go
type GetEntryQueryHandler struct {
    entryRepo     repositories.EntryRepository  // ✅ USA DOMAIN
    cache         services.CacheService         // ✅ USA SERVICES
    connectClient services.ConnectClient        // ✅ USA SERVICES
}
```

**Status**: ✅ **CORRETO** - Queries já usam interfaces do Domain!

#### Get Claim Query (`internal/application/queries/get_claim_query.go`)
```go
type GetClaimQueryHandler struct {
    claimRepo repositories.ClaimRepository  // ✅ USA DOMAIN
    cache     services.CacheService         // ✅ USA SERVICES
}
```

**Status**: ✅ **CORRETO** - Queries já usam interfaces do Domain!

---

## Tabela Comparativa: Interfaces Duplicadas

| Componente | Domain Layer | Commands Layer | Queries Layer | Infrastructure |
|------------|-------------|----------------|---------------|----------------|
| **Entry** | `entities.Entry` | `commands.Entry` ❌ | USA Domain ✅ | Implementa Domain ✅ |
| **Claim** | `entities.Claim` | `commands.Claim` ❌ | USA Domain ✅ | Implementa Domain ✅ |
| **EntryRepository** | Define interface | Redefine interface ❌ | USA Domain ✅ | Implementa Domain ✅ |
| **ClaimRepository** | Define interface | Redefine interface ❌ | USA Domain ✅ | Implementa Domain ✅ |
| **CacheService** | N/A | Redefine interface ❌ | USA Services ✅ | N/A |

---

## Raiz do Problema

### Problema 1: Duplicação de Entidades
Commands Layer define suas próprias entidades (`commands.Entry`, `commands.Claim`) em vez de reutilizar `entities.Entry` e `entities.Claim`.

**Por quê?**
Provavelmente para evitar dependência direta do Domain Layer, mas **isso quebra Clean Architecture**.

### Problema 2: Duplicação de Interfaces
Commands Layer redefine interfaces de repositórios com assinaturas DIFERENTES.

**Por quê?**
Tentativa de simplificar interfaces, mas **quebra compatibilidade**.

### Problema 3: Inconsistência entre Commands e Queries
- **Queries Layer**: ✅ Usa `repositories.EntryRepository` (Domain)
- **Commands Layer**: ❌ Define `commands.EntryRepository` (próprio)

---

## Soluções Propostas

### Opção A: Unificação Total (RECOMENDADO)

**Princípio**: Application Layer SEMPRE usa interfaces do Domain Layer.

#### Mudanças Necessárias

**1. Remover duplicação de entidades em Commands**
```go
// ANTES (commands/create_entry_command.go)
type Entry struct {
    ID        uuid.UUID
    KeyType   KeyType
    KeyValue  string
    // ...
}

// DEPOIS
import "github.com/lbpay-lab/core-dict/internal/domain/entities"
// Usar entities.Entry diretamente
```

**2. Remover duplicação de interfaces em Commands**
```go
// ANTES (commands/create_entry_command.go)
type EntryRepository interface {
    Create(ctx context.Context, entry *Entry) error
    // ...
}

// DEPOIS
import "github.com/lbpay-lab/core-dict/internal/domain/repositories"
// Usar repositories.EntryRepository diretamente
```

**3. Adicionar métodos faltantes no Domain Layer**
```go
// Adicionar em domain/repositories/entry_repository.go
CountByOwnerAndType(ctx context.Context, ownerTaxID string, keyType entities.KeyType) (int, error)

// Adicionar em domain/repositories/claim_repository.go
FindActiveByEntryID(ctx context.Context, entryID uuid.UUID) (*entities.Claim, error)
ListPendingClaims(ctx context.Context, limit int) ([]*entities.Claim, error)
```

**4. Unificar CacheService**
```go
// Escolher uma única interface (services.CacheService)
// Mapear InvalidateKey → Delete
// Mapear InvalidatePattern → Invalidate
```

#### Vantagens
✅ Solução permanente e correta (Clean Architecture)
✅ Elimina duplicação de código
✅ Facilita manutenção futura
✅ Queries e Commands usam as mesmas interfaces

#### Desvantagens
⚠️ Requer refatoração de ~10 arquivos
⚠️ Pode quebrar testes existentes

---

### Opção B: Camada de Adapters (NÃO RECOMENDADO)

Criar `internal/infrastructure/adapters/` para traduzir entre interfaces:

```go
// adapters/entry_repository_adapter.go
type EntryRepositoryAdapter struct {
    domainRepo repositories.EntryRepository
}

func (a *EntryRepositoryAdapter) Create(ctx context.Context, entry *commands.Entry) error {
    domainEntry := toDomainEntry(entry)
    return a.domainRepo.Create(ctx, domainEntry)
}

func (a *EntryRepositoryAdapter) FindByKeyValue(ctx context.Context, keyValue string) (*commands.Entry, error) {
    domainEntry, err := a.domainRepo.FindByKey(ctx, keyValue)
    if err != nil {
        return nil, err
    }
    return toCommandsEntry(domainEntry), nil
}
```

#### Vantagens
✅ Não quebra código existente
✅ Isola Commands de Domain

#### Desvantagens
❌ Adiciona complexidade desnecessária
❌ Duplicação de tipos permanece
❌ Mais código para manter
❌ Violar princípio Clean Architecture (Application depende de Domain)

---

## Decisão Recomendada

### ✅ OPÇÃO A: UNIFICAÇÃO TOTAL

**Motivos**:
1. **Clean Architecture**: Application Layer DEVE depender do Domain Layer
2. **Queries já fazem isso**: Queries usam `repositories.EntryRepository` corretamente
3. **Simplicidade**: Menos código, menos bugs
4. **Manutenibilidade**: Uma única fonte de verdade

**Esforço estimado**: 2-3 horas

**Arquivos a modificar**:
1. `internal/application/commands/create_entry_command.go` (remover duplicações)
2. `internal/application/commands/update_entry_command.go` (remover duplicações)
3. `internal/application/commands/delete_entry_command.go` (remover duplicações)
4. `internal/application/commands/create_claim_command.go` (remover duplicações)
5. `internal/application/commands/confirm_claim_command.go` (remover duplicações)
6. `internal/application/commands/cancel_claim_command.go` (remover duplicações)
7. `internal/application/commands/complete_claim_command.go` (remover duplicações)
8. `internal/application/commands/block_entry_command.go` (remover duplicações)
9. `internal/application/commands/unblock_entry_command.go` (remover duplicações)
10. `internal/domain/repositories/entry_repository.go` (adicionar métodos faltantes)
11. `internal/domain/repositories/claim_repository.go` (adicionar métodos faltantes)
12. `internal/infrastructure/database/entry_repository_impl.go` (implementar métodos faltantes)
13. `internal/infrastructure/database/claim_repository_impl.go` (implementar métodos faltantes)

---

## Próximos Passos

1. ✅ **Análise completa** (este documento)
2. 🔄 **Implementar Opção A**: Unificar interfaces
3. ⏳ **Testar compilação**: `go build ./...`
4. ⏳ **Atualizar testes unitários**
5. ⏳ **Documentar solução final**

---

## Conclusão

O bloqueio é 100% solucionável com refatoração sistemática. A solução correta é **Opção A (Unificação)**, que alinha o código com Clean Architecture e elimina duplicação desnecessária.

**Tempo estimado para Real Mode funcional**: 3-4 horas após implementação da solução.

---

**Autor**: Interface Unification Specialist
**Revisado por**: Project Manager
**Aprovação**: Pendente
