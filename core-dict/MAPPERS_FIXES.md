# Mapper Fixes - gRPC Proto ↔ Domain/Commands/Queries

**Data**: 2025-10-27
**Status**: ✅ COMPILANDO COM SUCESSO

---

## Resumo das Correções

Os mappers em `internal/infrastructure/grpc/mappers/` foram corrigidos para alinhar com as estruturas reais dos Commands, Queries e Entities do projeto.

### Arquivos Corrigidos

1. **key_mapper.go** - Mappers de Keys (Entry)
2. **claim_mapper.go** - Mappers de Claims
3. **error_mapper.go** - Mappers de Errors (comentados errors não existentes)

---

## Principais Mudanças

### 1. Ajustes em CreateKeyRequest → CreateEntryCommand

**Proto**: `CreateKeyRequest` tem apenas:
- `key_type` (enum)
- `key_value` (string)
- `account_id` (string UUID)

**Solução**: Mapper simplificado. Handler irá buscar detalhes da conta (ISPB, Branch, Owner, etc.) usando o `account_id`.

```go
return commands.CreateEntryCommand{
    KeyType:     keyType,
    KeyValue:    req.GetKeyValue(),
    AccountID:   accountID,
    RequestedBy: requestedBy,
    // AccountISPB, AccountBranch, etc. - populated by handler
}
```

### 2. Ajustes em StartClaimRequest → CreateClaimCommand

**Proto**: `StartClaimRequest` tem apenas:
- `key` (DictKey)
- `account_id` (string UUID)

**Solução**: Mapper simplificado. Handler irá determinar ClaimType, ISPBs, OwnerTaxID a partir do contexto.

```go
return commands.CreateClaimCommand{
    KeyValue:    req.GetKey().GetKeyValue(),
    AccountID:   accountID,
    RequestedBy: requestedBy,
    // ClaimType, ClaimerISPB, ClaimedISPB, OwnerTaxID - determined by handler
}
```

### 3. Ajustes em RespondToClaimRequest

**Proto**: `RespondToClaimRequest` tem:
- `claim_id` (string UUID)
- `response` (enum: ACCEPT/REJECT)
- `reason` (optional string)

**Commands** esperavam:
- `TwoFactorCode` - virá do auth context/header
- `ConfirmedBy` - virá do auth context (nome do usuário)

**Solução**: Mappers simplificados, campos de auth virão do contexto.

### 4. Ajustes em ListKeysRequest e ListClaimsRequest

**Proto**: Usa `page_token` (cursor-based pagination)
**Query**: Usa `page` (int, 1-indexed)

**Solução Temporária**: `page = 1` (TODO: extrair de `page_token`)

### 5. Conversão de Tipos entities → proto

**Problema**: Entry entity usa `entities.KeyType` e `entities.KeyStatus` (string-based), não `valueobjects`.

**Solução**: Criados helper functions:
- `mapEntityKeyTypeToProto(entities.KeyType) → commonv1.KeyType`
- `mapEntityStatusToProto(entities.KeyStatus) → commonv1.EntryStatus`

### 6. Conversão UUID → string

**Problema**: Proto usa `string` para IDs, entities usam `uuid.UUID`.

**Solução**: Usar `.String()` ao mapear:
```go
KeyId: entry.ID.String(),
AccountId: entry.AccountID.String(),
```

### 7. Mapeamento de Status Proto ↔ Domain

**Proto EntryStatus** (common.proto):
- `ENTRY_STATUS_ACTIVE`
- `ENTRY_STATUS_PORTABILITY_PENDING`
- `ENTRY_STATUS_PORTABILITY_CONFIRMED`
- `ENTRY_STATUS_CLAIM_PENDING`
- `ENTRY_STATUS_DELETED`

**Domain KeyStatus** (entities):
- `PENDING` → `ENTRY_STATUS_PORTABILITY_PENDING`
- `ACTIVE` → `ENTRY_STATUS_ACTIVE`
- `BLOCKED` → `ENTRY_STATUS_DELETED` (mapeado para deleted)
- `DELETED` → `ENTRY_STATUS_DELETED`

**Nota**: Proto não tem `PENDING`, `BLOCKED`, `FAILED` genéricos - apenas estados específicos de portabilidade/claim.

### 8. Claim Entity Mapping

**Problema**: Claim entity tem estrutura diferente:
- `EntryKey` (string) - não tem `KeyType` separado
- `ClaimerParticipant`/`DonorParticipant` (valueobjects.Participant) - não ISPBs diretos
- Não armazena `EntryID` separadamente

**Solução**: Mappers comentaram campos não disponíveis com TODO:
```go
EntryId: "", // TODO: Not stored in Claim entity - needs lookup
KeyType: commonv1.KeyType_KEY_TYPE_UNSPECIFIED, // TODO: Parse from EntryKey
ClaimerIspb: claim.ClaimerParticipant.ISPB, // OK
```

### 9. Account Entity Mapping

**Problema**: Account entity usa:
- `Branch` (não `BranchCode`)
- `AccountType` como `entities.AccountType` (string constants: "CACC", "SVGS", "SLRY", "TRAN")

**Solução**: Helper function `mapEntityAccountTypeToProto`:
```go
case entities.AccountTypeCACC: return commonv1.AccountType_ACCOUNT_TYPE_CHECKING
case entities.AccountTypeSVGS: return commonv1.AccountType_ACCOUNT_TYPE_SAVINGS
...
```

### 10. Error Mapper

**Problema**: Referências a domain errors não existentes:
- `domain.ErrClaimNotFound`
- `domain.ErrDuplicateKeyGlobal`
- `domain.ErrNotOwner`
- `domain.ErrCannotDeleteActiveKey`
- etc.

**Solução**: Comentados com TODO até serem implementados em `domain/errors.go`.

---

## TODOs Identificados

### Alta Prioridade

1. **Pagination Token**: Implementar extração de `page` a partir de `page_token` (cursor-based)
2. **2FA/Auth Context**: Passar `TwoFactorCode` e `ConfirmedBy` via auth middleware/context
3. **Account Lookup**: Handler deve buscar detalhes completos da conta usando `account_id`
4. **Claim KeyType**: Adicionar `KeyType` ao Claim entity ou parser para extrair de `EntryKey`

### Média Prioridade

5. **Claim EntryID**: Adicionar `EntryID` ao Claim entity ou implementar lookup
6. **Domain Errors**: Adicionar errors faltantes em `domain/errors.go`
7. **Account Proto Fields**: Mapear campos completos do proto Account (check_digit, holder_name, etc.)

### Baixa Prioridade

8. **ClaimType Inference**: Handler deve inferir ClaimType (OWNERSHIP vs PORTABILITY) do contexto
9. **Portability History**: Implementar mapeamento de histórico de portabilidades

---

## Assinaturas dos Mappers Alterados

### key_mapper.go

```go
// Return types mudaram de T para (T, error)
func MapProtoCreateKeyRequestToCommand(req, userID) (commands.CreateEntryCommand, error)
func MapProtoListKeysRequestToQuery(req, accountID) (queries.ListEntriesQuery, error)
func MapProtoDeleteKeyRequestToCommand(req, userID) (commands.DeleteEntryCommand, error)
```

### claim_mapper.go

```go
// Return types mudaram de T para (T, error)
func MapProtoStartClaimRequestToCommand(req, userID) (commands.CreateClaimCommand, error)
func MapProtoRespondToClaimRequestToConfirmCommand(req, userID) (commands.ConfirmClaimCommand, error)
func MapProtoRespondToClaimRequestToCancelCommand(req, userID) (commands.CancelClaimCommand, error)
func MapProtoCancelClaimRequestToCommand(req, userID) (commands.CancelClaimCommand, error)

// Parâmetro adicional `ispb` (de auth context)
func MapProtoListIncomingClaimsRequestToQuery(req, ispb) queries.ListClaimsQuery
func MapProtoListOutgoingClaimsRequestToQuery(req, ispb) queries.ListClaimsQuery
```

---

## Teste de Compilação

```bash
cd /Users/jose.silva.lb/LBPay/IA_Dict/core-dict
go build ./internal/infrastructure/grpc/...
# ✅ SUCCESS - No errors
```

---

## Próximos Passos

1. **Implementar gRPC Server**: Usar os mappers corrigidos nos handlers do gRPC server
2. **Adicionar Auth Middleware**: Extrair user_id, account_id, ispb, 2FA do contexto
3. **Implementar Handlers**: Preencher campos faltantes (account details, claim type, etc.)
4. **Adicionar Testes**: Unit tests para cada mapper function
5. **Pagination**: Implementar cursor-based pagination (page_token)

---

## Arquitetura da Conversão

```
┌──────────────┐
│ Proto (gRPC) │
│  CreateKey   │
│  Request     │
└──────┬───────┘
       │ MapProtoCreateKeyRequestToCommand()
       ↓
┌──────────────────┐
│ CreateEntry      │
│ Command          │ ← Partial (only key_type, key_value, account_id)
└──────┬───────────┘
       │ Handler.Handle()
       │ - Lookup Account details
       │ - Fetch Owner info
       │ - Validate permissions
       ↓
┌──────────────────┐
│ CreateEntry      │
│ Command (Full)   │ ← Complete command with all fields
└──────┬───────────┘
       │ Domain Logic
       ↓
┌──────────────────┐
│ Entry (Entity)   │
└──────────────────┘
```

---

**Status Final**: ✅ Mappers compilando e prontos para uso nos handlers gRPC
