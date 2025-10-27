# Análise: Operações Síncronas vs Assíncronas - DICT LBPay

**Data**: 2025-10-27
**Autor**: Project Manager + Squad Lead
**Status**: ✅ Análise Completa
**Baseado em**: INT-001, INT-002, TSP-001

---

## 🎯 Objetivo

Documentar **quais operações do DICT são síncronas vs assíncronas** para evitar implementação incorreta de Workflows Temporal desnecessários.

**REGRA DE OURO**:
- **Temporal Workflows** = Apenas operações de **longa duração** (horas/dias)
- **gRPC direto** = Operações **rápidas** (< 2s)

---

## 📊 Classificação de Operações

### ✅ **SÍNCRONAS** (Sem Temporal Workflow)

Operações que **devem retornar imediatamente** (< 2s):

| Operação | API | Descrição | Tempo | Justificativa |
|----------|-----|-----------|-------|---------------|
| **GetEntry** | `GET /keys/:id` | Consulta chave PIX | 10-50ms | Read-only, cache Redis |
| **ListEntries** | `GET /keys` | Lista chaves por ISPB | 50-200ms | Read-only, paginação |
| **CheckKeyAvailability** | `GET /keys/check?key=X` | Verifica se chave está disponível | 10-30ms | Read-only, cache |
| **GetClaim** | `GET /claims/:id` | Consulta status de claim | 10-50ms | Read-only, DB query |
| **ListClaims** | `GET /claims` | Lista claims | 50-200ms | Read-only, paginação |
| **GetInfraction** | `GET /infractions/:id` | Consulta infração | 10-50ms | Read-only |
| **ListInfractions** | `GET /infractions` | Lista infrações | 50-200ms | Read-only |

**Implementação**:
- **gRPC Service** direto (EntryService, ClaimService, InfractionService)
- **Queries CQRS** (`internal/application/queries/`)
- **Repository** read methods
- **Redis cache** para reads frequentes

---

### ⏱️ **SEMI-SÍNCRONAS** (API retorna imediato + Processamento Assíncrono)

Operações que **retornam 201 Created IMEDIATAMENTE**, mas **processamento continua em background**:

| Operação | API | Resposta Imediata | Processamento Async | Tempo Total | Usa Temporal? |
|----------|-----|-------------------|---------------------|-------------|---------------|
| **CreateEntry** | `POST /keys` | `201 Created` + `status: PENDING` | Pulsar → Bacen via Bridge | 800ms-1.5s | **NÃO** (< 2s) |
| **UpdateEntry** | `PUT /keys/:id` | `200 OK` + `status: UPDATING` | Pulsar → Bacen via Bridge | 500ms-1s | **NÃO** |
| **DeleteEntry** | `DELETE /keys/:id` | `202 Accepted` + `status: DELETING` | Pulsar → Bacen via Bridge | 500ms-1s | **NÃO** |

**IMPORTANTE**:
- Core DICT **retorna imediatamente** com `status: PENDING`
- **Pulsar Consumer** pega evento e processa
- **RSFN Connect** chama Bridge (gRPC) → Bridge chama Bacen (SOAP/mTLS)
- **Workflow Temporal NÃO é necessário** (processamento < 2s)

**Implementação**:
- **Core DICT**: POST → INSERT DB (status PENDING) → Publish Pulsar event → Return 201
- **Pulsar Consumer** (`conn-dict/internal/infrastructure/pulsar/consumer.go`):
  - Escuta `dict.entries.created`
  - Chama **gRPC Bridge** diretamente (sem Temporal)
  - Update status para ACTIVE ou FAILED
- **Activities Temporal**: APENAS se precisar retry > 3x ou timeout > 2min

---

### ⏳ **ASSÍNCRONAS** (Requerem Temporal Workflow - Longa Duração)

Operações que **duram horas ou dias**:

| Operação | API | Duração | Workflow Temporal | Justificativa |
|----------|-----|---------|-------------------|---------------|
| **CreateClaim** | `POST /claims` | **30 dias** | `ClaimWorkflow` | Timer durável de 30 dias (regra Bacen) |
| **VSyncDailyJob** | Cron | **15-30 min** | `VSyncWorkflow` | Sincronização batch (10k+ entries) |
| **InvestigateInfraction** | `POST /infractions/:id/investigate` | **Dias/semanas** | `InfractionWorkflow` | Processo manual de investigação |

**Implementação**:
- **Temporal Workflow** completo
- **Activities** com retry policies robustas
- **Timer durável** (30 dias para Claim)
- **Human-in-the-loop** (Infraction investigation)

---

## 🔍 Análise do Código Atual (conn-dict)

### ✅ **Já Implementado Corretamente**

1. **ClaimWorkflow** ✅
   - Path: `internal/workflows/claim_workflow.go`
   - Timer 30 dias: ✅
   - Activities: CreateClaimActivity, MonitorClaimActivity, etc.
   - **CORRETO**: Claim precisa de workflow durável

2. **VSyncWorkflow** ✅
   - Path: `internal/workflows/vsync_workflow.go`
   - Batch processing: ✅
   - Activities: FetchBacenEntriesActivity, CompareEntriesActivity, GenerateSyncReportActivity
   - **CORRETO**: VSYNC é operação batch longa

3. **InfractionWorkflow** ✅
   - Path: `internal/workflows/infraction_workflow.go`
   - Human-in-the-loop: ✅
   - Signal `InvestigationDecisionSignal`
   - **CORRETO**: Infração requer processo manual

---

### ❌ **Implementação INCORRETA Identificada**

#### ⚠️ **entry_create_workflow.go** - **NÃO DEVERIA EXISTIR**

**Arquivo**: `internal/workflows/entry_create_workflow.go`

**Problema**:
```go
// CreateEntryWorkflow - ISSO NÃO É NECESSÁRIO!
// CreateEntry é operação RÁPIDA (< 2s), não precisa de Temporal Workflow
func CreateEntryWorkflow(ctx workflow.Context, input CreateEntryWorkflowInput) (*CreateEntryWorkflowResult, error) {
    // ...
}
```

**Razão**:
- CreateEntry leva **800ms-1.5s** (rápido!)
- Core DICT retorna `201 Created` imediatamente
- Pulsar Consumer → Bridge → Bacen (tudo < 2s)
- **Temporal Workflow é OVERKILL** para operações < 2s

**Solução Correta**:
```go
// Core DICT (internal/application/commands/create_entry_command.go)
func (h *CreateEntryHandler) Handle(ctx context.Context, cmd CreateEntryCommand) error {
    // 1. Validate
    // 2. INSERT DB (status: PENDING)
    // 3. Publish Pulsar event "dict.entries.created"
    // 4. Return imediato
}

// Pulsar Consumer (conn-dict/internal/infrastructure/pulsar/consumer.go)
func (c *Consumer) handleEntryCreated(msg pulsar.Message) error {
    // 1. Consume event
    // 2. Call Bridge gRPC directly (NO TEMPORAL!)
    // 3. Update status ACTIVE or FAILED
}
```

---

#### ⚠️ **entry_update_workflow.go** - **NÃO DEVERIA EXISTIR**

**Arquivo**: `internal/workflows/entry_update_workflow.go`

**Problema**: Mesma lógica de CreateEntry - operação rápida (<1s) não precisa de Temporal.

**Solução**: Pulsar Consumer direto.

---

#### ⚠️ **entry_delete_workflow.go** - **TALVEZ NECESSÁRIO (depende do cenário)**

**Arquivo**: `internal/workflows/entry_delete_workflow.go`

**Análise**:
```go
const (
    // DeleteWaitingPeriodDays is the number of days to wait before permanent deletion
    DeleteWaitingPeriodDays = 30
)
```

**Cenários**:
1. **Delete imediato** (admin/compliance): **NÃO precisa de Temporal** (< 2s)
2. **Delete com período de 30 dias** (soft delete): **PRECISA de Temporal** (timer durável)

**Solução**:
- Se `immediate_delete = true`: Pulsar Consumer direto
- Se `immediate_delete = false`: Temporal Workflow com timer 30 dias

**Refactor**:
```go
// entry_delete_workflow.go - APENAS para delete com período de espera
func DeleteEntryWithWaitingPeriodWorkflow(ctx workflow.Context, input DeleteEntryWorkflowInput) (*DeleteEntryWorkflowResult, error) {
    // 1. Soft delete (status = DEACTIVATED)
    // 2. Timer 30 dias
    // 3. Hard delete (status = DELETED)
}

// Pulsar Consumer - Para delete imediato
func (c *Consumer) handleEntryDeleteImmediate(msg pulsar.Message) error {
    // 1. Consume event
    // 2. Call Bridge gRPC directly
    // 3. Update status DELETED
}
```

---

## 📋 Decisão Final: O Que Fazer?

### ✅ **Manter Workflows Existentes**:
1. **ClaimWorkflow** ✅ (30 dias - durável)
2. **VSyncWorkflow** ✅ (batch longo - 15-30min)
3. **InfractionWorkflow** ✅ (human-in-the-loop - dias/semanas)
4. **DeleteEntryWithWaitingPeriodWorkflow** �� (30 dias - durável) - **Renomear e refatorar**

### ❌ **Remover Workflows Desnecessários**:
1. **CreateEntryWorkflow** ❌ - Substituir por Pulsar Consumer direto
2. **UpdateEntryWorkflow** ❌ - Substituir por Pulsar Consumer direto

### 🔄 **Implementar Corretamente**:

#### **Pulsar Consumer** (`conn-dict/internal/infrastructure/pulsar/consumer.go`)

```go
package pulsar

import (
    "context"
    "encoding/json"

    "github.com/apache/pulsar-client-go/pulsar"
    "github.com/sirupsen/logrus"

    "github.com/lbpay-lab/conn-dict/internal/domain/entities"
    "github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
    bridgepb "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
    "google.golang.org/grpc"
)

type Consumer struct {
    client          pulsar.Client
    consumer        pulsar.Consumer
    logger          *logrus.Logger
    entryRepo       *repositories.EntryRepository
    bridgeClient    bridgepb.BridgeServiceClient  // gRPC client para Bridge
}

// Consome eventos de criação de Entry e chama Bridge DIRETAMENTE (sem Temporal)
func (c *Consumer) handleEntryCreated(ctx context.Context, msg pulsar.Message) error {
    var event struct {
        EntryID string `json:"entry_id"`
        Key     string `json:"key"`
        // ... outros campos
    }

    if err := json.Unmarshal(msg.Payload(), &event); err != nil {
        return err
    }

    // 1. Busca Entry no DB
    entry, err := c.entryRepo.GetByEntryID(ctx, event.EntryID)
    if err != nil {
        return err
    }

    // 2. Chama Bridge gRPC DIRETAMENTE (SEM Temporal!)
    req := &bridgepb.CreateEntryRequest{
        EntryId: entry.EntryID,
        Key:     entry.Key,
        KeyType: string(entry.KeyType),
        // ... outros campos
    }

    resp, err := c.bridgeClient.CreateEntry(ctx, req)
    if err != nil {
        // 3. Update status FAILED
        entry.Status = entities.EntryStatusFailed
        c.entryRepo.UpdateStatus(ctx, entry.EntryID, entities.EntryStatusFailed)
        return err
    }

    // 4. Update status ACTIVE + Bacen Entry ID
    entry.Status = entities.EntryStatusActive
    entry.BacenEntryID = &resp.BacenEntryId
    return c.entryRepo.Update(ctx, entry)
}
```

---

## 📊 Tabela Resumo - Decisões de Implementação

| Operação | Tempo | API Response | Processamento | Usa Temporal? | Implementação |
|----------|-------|-------------|---------------|---------------|---------------|
| **GetEntry** | 10-50ms | Síncrono (200 OK) | Nenhum | ❌ | gRPC Service direto |
| **ListEntries** | 50-200ms | Síncrono (200 OK) | Nenhum | ❌ | gRPC Service direto |
| **CreateEntry** | 800ms-1.5s | `201 Created` (imediato) | **Pulsar Consumer → Bridge gRPC** | ❌ | **Pulsar Consumer direto** |
| **UpdateEntry** | 500ms-1s | `200 OK` (imediato) | **Pulsar Consumer → Bridge gRPC** | ❌ | **Pulsar Consumer direto** |
| **DeleteEntry (imediato)** | 500ms-1s | `202 Accepted` (imediato) | **Pulsar Consumer → Bridge gRPC** | ❌ | **Pulsar Consumer direto** |
| **DeleteEntry (30 dias)** | **30 dias** | `202 Accepted` (imediato) | **Temporal Workflow** | ✅ | **DeleteWithWaitingPeriodWorkflow** |
| **CreateClaim** | **30 dias** | `201 Created` (imediato) | **Temporal Workflow (durável)** | ✅ | **ClaimWorkflow** (já implementado) |
| **VSyncDailyJob** | **15-30 min** | N/A (cron job) | **Temporal Workflow (batch)** | ✅ | **VSyncWorkflow** (já implementado) |
| **InvestigateInfraction** | **Dias/semanas** | `202 Accepted` (imediato) | **Temporal Workflow (human-in-the-loop)** | ✅ | **InfractionWorkflow** (já implementado) |

---

## ✅ Próximas Ações

### 1️⃣ **Refatorar Entry Operations**
- [ ] **Remover** `entry_create_workflow.go`
- [ ] **Remover** `entry_update_workflow.go`
- [ ] **Renomear** `entry_delete_workflow.go` → `entry_delete_with_waiting_period_workflow.go`
- [ ] **Criar** `pulsar_consumer.go` completo com:
  - `handleEntryCreated()`
  - `handleEntryUpdated()`
  - `handleEntryDeleteImmediate()`

### 2️⃣ **Implementar gRPC Services Corretamente**
- [ ] **EntryService**: Queries síncronas (GetEntry, ListEntries)
- [ ] **ClaimService**: Queries síncronas + CreateClaim (inicia Temporal Workflow)
- [ ] **InfractionService**: Queries síncronas + InvestigateInfraction (inicia Temporal Workflow)

### 3️⃣ **Validar Arquitetura com Artefatos**
- [x] ✅ Ler INT-001 (CreateEntry E2E) - **Validado**: Processamento assíncrono < 2s
- [x] ✅ Ler INT-002 (ClaimWorkflow E2E) - **Validado**: Workflow durável 30 dias
- [x] ✅ Ler TSP-001 (Temporal Spec) - **Validado**: Workflows apenas para operações longas

---

## 📝 Conclusão

**RESUMO**:
- ✅ **Temporal Workflows**: Apenas para operações **> 2 minutos** (Claims, VSYNC, Infractions)
- ❌ **NÃO usar Temporal**: Para operações **< 2s** (CreateEntry, UpdateEntry, DeleteEntry imediato)
- 🔄 **Pulsar Consumer**: Para processamento assíncrono **< 2s** (Entry CRUD via Bridge)

**AGRADECIMENTO**:
Obrigado por questionar a implementação! Isso evitou uma arquitetura **over-engineered** com Temporal Workflows desnecessários para operações rápidas. 🙏

---

**Próximo Passo**: Criar plano de refatoração e implementação correta com agentes paralelos.

**Autor**: Claude Sonnet 4.5 (Project Manager)
**Data**: 2025-10-27 10:30 BRT
**Status**: ✅ Análise completa e validada
