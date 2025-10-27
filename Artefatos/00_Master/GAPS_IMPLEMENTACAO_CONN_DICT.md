# GAPs de Implementação - conn-dict

**Data**: 2025-10-27 10:45 BRT
**Status**: ✅ Análise Completa
**Baseado em**: Validação de código + ANALISE_SYNC_VS_ASYNC_OPERATIONS.md

---

## 📊 Status Atual

### ✅ **Já Implementado** (95% completo)

| Componente | Arquivos | LOC Total | Status |
|------------|----------|-----------|--------|
| **Workflows** | 9 arquivos (7 prod + 2 test) | ~2,027 LOC | ✅ 95% |
| **Activities** | 6 arquivos | ~1,875 LOC | ✅ 100% |
| **Repositories** | 4 arquivos | ~1,443 LOC | ✅ 100% |
| **Entities** | 5 arquivos | ~835 LOC | ✅ 100% |
| **gRPC Interceptors** | 4 arquivos | ~680 LOC | ✅ 100% |
| **gRPC Services** | 1 arquivo (EntryService) | 315 LOC | ⚠️ 33% |
| **Pulsar Consumer** | 1 arquivo (producer OK) | 0 LOC | ❌ 0% |
| **Commands/Queries** | 18 arquivos | ~1,800 LOC | ✅ 100% |

**Total**: **~9,000 LOC implementados**

---

## 🔴 GAP #1: Remover Workflows Desnecessários

### **Problema**:
Workflows Temporal criados para operações **< 2s** que **NÃO** deveriam usar Temporal.

### **Arquivos a REMOVER**:

1. ❌ **`entry_create_workflow.go`** (194 LOC)
   - **Razão**: CreateEntry é operação < 1.5s
   - **Substituir por**: Pulsar Consumer → Bridge gRPC direto

2. ❌ **`entry_update_workflow.go`** (223 LOC)
   - **Razão**: UpdateEntry é operação < 1s
   - **Substituir por**: Pulsar Consumer → Bridge gRPC direto

### **Arquivos a REFATORAR**:

3. ⚠️ **`entry_delete_workflow.go`** (261 LOC)
   - **Razão**: Delete pode ser imediato OU com período 30 dias
   - **Ação**:
     - Renomear para `entry_delete_with_waiting_period_workflow.go`
     - Remover lógica de delete imediato
     - Manter APENAS timer 30 dias
   - **Adicionar**: `handleEntryDeleteImmediate()` no Pulsar Consumer

---

## 🔴 GAP #2: Criar Pulsar Consumer Completo

### **Problema**:
**NÃO existe** `conn-dict/internal/infrastructure/pulsar/consumer.go` funcional.

**Existe apenas**:
- `producer.go` (135 LOC) ✅
- `event_publisher.go` (98 LOC) ✅

### **Arquivo a CRIAR**:

**`conn-dict/internal/infrastructure/pulsar/consumer.go`**

**Funcionalidades**:
1. **Subscribe** em 3 topics:
   - `dict.entries.created`
   - `dict.entries.updated`
   - `dict.entries.deleted.immediate`

2. **Handlers**:
   ```go
   func (c *Consumer) handleEntryCreated(ctx context.Context, msg pulsar.Message) error {
       // 1. Parse event
       // 2. Call Bridge gRPC CreateEntry
       // 3. Update Entry status (ACTIVE or FAILED)
   }

   func (c *Consumer) handleEntryUpdated(ctx context.Context, msg pulsar.Message) error {
       // 1. Parse event
       // 2. Call Bridge gRPC UpdateEntry
       // 3. Update Entry status
   }

   func (c *Consumer) handleEntryDeleteImmediate(ctx context.Context, msg pulsar.Message) error {
       // 1. Parse event
       // 2. Call Bridge gRPC DeleteEntry
       // 3. Update Entry status (DELETED)
   }
   ```

3. **Dependencies**:
   - Bridge gRPC Client (`bridgepb.BridgeServiceClient`)
   - Entry Repository (`repositories.EntryRepository`)
   - Logger (logrus)

**LOC Estimado**: ~350 LOC

---

## 🟡 GAP #3: Completar gRPC Services

### **Problema**:
Apenas **EntryService** implementado (315 LOC). Faltam:

1. ❌ **ClaimService** (0 LOC)
   - `CreateClaim` → Inicia `ClaimWorkflow` no Temporal
   - `ConfirmClaim` → Signal `ConfirmClaimSignal`
   - `CancelClaim` → Signal `CancelClaimSignal`
   - `GetClaim` → Query DB
   - `ListClaims` → Query DB com paginação

2. ❌ **InfractionService** (0 LOC)
   - `CreateInfraction` → Inicia `InfractionWorkflow` no Temporal
   - `InvestigateInfraction` → Signal `InvestigationDecisionSignal`
   - `ResolveInfraction` → Signal
   - `DismissInfraction` → Signal
   - `GetInfraction` → Query DB
   - `ListInfractions` → Query DB

**LOC Estimado**:
- ClaimService: ~400 LOC
- InfractionService: ~450 LOC
- **Total**: ~850 LOC

---

## 🟡 GAP #4: Implementar gRPC Server Setup

### **Problema**:
**Existe** `internal/grpc/server.go` (skeleto), mas falta:

1. **Registrar Services**:
   ```go
   // Falta registrar
   pb.RegisterClaimServiceServer(grpcServer, claimService)
   pb.RegisterInfractionServiceServer(grpcServer, infractionService)
   ```

2. **Registrar Interceptors**:
   ```go
   // Já existem os interceptors, mas falta chain
   grpc.ChainUnaryInterceptor(
       interceptors.LoggingInterceptor,
       interceptors.MetricsInterceptor,
       interceptors.RecoveryInterceptor,
       interceptors.TracingInterceptor,
   )
   ```

3. **Health Check**:
   ```go
   healthpb.RegisterHealthServer(grpcServer, healthServer)
   ```

**LOC Estimado**: ~150 LOC (updates em `server.go`)

---

## 🟡 GAP #5: cmd/server/main.go

### **Problema**:
**NÃO existe** `cmd/server/main.go` (apenas `cmd/worker/main.go`).

**Precisa criar**:
- `cmd/server/main.go` (gRPC server entrypoint)
- Inicializar:
  - PostgreSQL client
  - Redis client
  - Temporal client
  - Pulsar producer
  - Bridge gRPC client
  - EntryService, ClaimService, InfractionService
  - gRPC server com interceptors

**LOC Estimado**: ~300 LOC

---

## 🟢 GAP #6: FetchBacenEntriesActivity (VSYNC)

### **Problema**:
`vsync_activities.go` tem placeholder:
```go
func (a *VSyncActivities) FetchBacenEntriesActivity(ctx context.Context, input FetchBacenEntriesInput) (*FetchBacenEntriesOutput, error) {
    // TODO: Implement integration with conn-bridge
    // For now, return empty list
    return &FetchBacenEntriesOutput{Entries: []VSyncEntry{}}, nil
}
```

**Precisa**:
- Chamar **Bridge gRPC** `ListEntriesFromBacen(participant_ispb, last_sync_date)`
- Parsear resposta
- Retornar lista de `VSyncEntry`

**LOC Estimado**: ~80 LOC (update em `vsync_activities.go`)

---

## 🟢 GAP #7: Integration Tests

### **Problema**:
Testes unitários existem (8 arquivos `*_test.go`), mas faltam **integration tests**.

**Precisa criar**:
1. `tests/integration/entry_integration_test.go`
   - Test: CreateEntry → Pulsar → Bridge → DB update
   - Requires: Testcontainers (PostgreSQL, Pulsar, Bridge mock)

2. `tests/integration/claim_integration_test.go`
   - Test: CreateClaim → Temporal Workflow → Timer 30 dias
   - Requires: Temporal test suite

3. `tests/integration/vsync_integration_test.go`
   - Test: VSyncWorkflow → FetchBacen → CompareEntries → GenerateReport
   - Requires: Bridge mock com 1k entries

**LOC Estimado**: ~600 LOC

---

## 📋 Resumo de GAPs

| GAP | Tipo | LOC Estimado | Prioridade | Tempo Estimado |
|-----|------|--------------|------------|----------------|
| **#1: Remover Workflows desnecessários** | Refactor | -417 LOC | 🔴 P0 | 2h |
| **#2: Pulsar Consumer completo** | Novo | +350 LOC | 🔴 P0 | 4h |
| **#3: gRPC Services (Claim + Infraction)** | Novo | +850 LOC | 🟡 P1 | 6h |
| **#4: gRPC Server Setup** | Update | +150 LOC | 🟡 P1 | 2h |
| **#5: cmd/server/main.go** | Novo | +300 LOC | 🟡 P1 | 3h |
| **#6: FetchBacenEntriesActivity** | Update | +80 LOC | 🟢 P2 | 1h |
| **#7: Integration Tests** | Novo | +600 LOC | 🟢 P2 | 8h |
| **TOTAL** | | **+1,913 LOC** | | **26h** |

---

## ✅ Plano de Execução (Ordem de Implementação)

### **Sprint 3 - Finalização (Esta semana)**

#### **Fase 1: Refatoração (P0) - 6h**
1. ✅ Criar documento `ANALISE_SYNC_VS_ASYNC_OPERATIONS.md`
2. ❌ Remover `entry_create_workflow.go` (-194 LOC)
3. ❌ Remover `entry_update_workflow.go` (-223 LOC)
4. ⚠️ Refatorar `entry_delete_workflow.go` → `entry_delete_with_waiting_period_workflow.go`
5. ✅ Criar `Pulsar Consumer` completo (+350 LOC)

#### **Fase 2: gRPC Services (P1) - 11h**
6. ✅ Criar `ClaimService` completo (+400 LOC)
7. ✅ Criar `InfractionService` completo (+450 LOC)
8. ✅ Atualizar `internal/grpc/server.go` (+150 LOC)
9. ✅ Criar `cmd/server/main.go` (+300 LOC)

#### **Fase 3: Finishing Touches (P2) - 9h**
10. ✅ Implementar `FetchBacenEntriesActivity` real (+80 LOC)
11. ✅ Criar integration tests (+600 LOC)
12. ✅ Atualizar documentação de gestão
13. ✅ Build final + validação

---

## 🎯 Próxima Ação

**Executar com MÁXIMO PARALELISMO**:

### **6 Agentes em Paralelo**:

1. **refactor-agent**:
   - Remover workflows desnecessários
   - Refatorar delete workflow

2. **pulsar-agent**:
   - Criar Pulsar Consumer completo
   - Implementar 3 handlers

3. **claim-service-agent**:
   - Criar ClaimService completo
   - 5 RPCs

4. **infraction-service-agent**:
   - Criar InfractionService completo
   - 5 RPCs

5. **grpc-server-agent**:
   - Atualizar server.go
   - Criar cmd/server/main.go

6. **vsync-agent**:
   - Implementar FetchBacenEntriesActivity
   - Bridge integration

**Tempo Total**: ~6h (com paralelismo)

---

**Autor**: Claude Sonnet 4.5 (Project Manager)
**Data**: 2025-10-27 10:45 BRT
**Status**: ✅ GAPs identificados e priorizados
