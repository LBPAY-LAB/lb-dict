# ADR-002: Escolha de Orchestrator - Temporal Workflow

**Status**: ✅ Aceito
**Data**: 2025-10-24
**Decisores**: Thiago Lima (Head de Arquitetura), José Luís Silva (CTO)
**Contexto Técnico**: Projeto DICT - LBPay

---

## Controle de Versão

| Versão | Data | Autor | Descrição das Mudanças |
|--------|------|-------|------------------------|
| 1.0 | 2025-10-24 | ARCHITECT | Versão inicial - Documentação da decisão de usar Temporal como orchestrator |

---

## Status

**✅ ACEITO** - Temporal Workflow já é tecnologia confirmada e em uso no LBPay

---

## Contexto

O projeto DICT da LBPay requer um **orchestrator de workflows assíncronos** para gerenciar processos de longa duração (long-running workflows) com alta confiabilidade. O sistema precisa de:

### Requisitos Funcionais

1. **Orquestração de Processos Assíncronos**:
   - **Cadastro de chave PIX**: Enviar CreateEntry ao Bacen → Aguardar confirmação (até 2 min)
   - **Reivindicação (Claim)**: Enviar CreateClaim → Aguardar resposta PSP (até 7 dias corridos)
   - **Portabilidade**: Enviar CreatePortability → Aguardar resposta PSP (até 7 dias corridos)
   - **VSYNC**: Executar sincronização periódica (agendada diariamente)

2. **Retry e Compensação**:
   - Retry automático em caso de falhas transientes (timeout, network error)
   - Exponential backoff (1s, 2s, 4s, 8s, ...)
   - Circuit breaker para falhas persistentes
   - Compensação (SAGA pattern) em caso de falha irreversível

3. **Timeouts e Deadlines**:
   - Timeout por activity (ex: SendCreateEntry = 30s)
   - Timeout total do workflow (ex: RegisterKeyWorkflow = 5 min)
   - Deadline para claim/portability (7 dias) com auto-confirmação/cancelamento

4. **Visibilidade e Debugging**:
   - UI para visualizar workflows em execução
   - Histórico completo de execuções (auditoria)
   - Replay de workflows para debugging

### Requisitos Não-Funcionais

| ID | Requisito | Target | Fonte |
|----|-----------|--------|-------|
| **NFR-090** | Durabilidade de workflows | 100% (zero perda de estado) | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-091** | Disponibilidade | ≥ 99.99% | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-092** | Latência (workflow start) | ≤ 100ms | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-093** | Throughput | ≥ 1.000 workflows/sec | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-094** | Retenção de histórico | 5 anos (compliance) | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-095** | Suporte a long-running workflows | Até 7 dias (claims) | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |

### Contexto Organizacional

- **LBPay já utiliza Temporal** em outros sistemas de produção
- Equipe de backend possui expertise em Temporal (Go SDK)
- Redução de custo operacional (não introduzir nova tecnologia)
- Consistência tecnológica entre projetos LBPay

---

## Decisão

**Escolhemos Temporal Workflow como orchestrator de workflows assíncronos para o projeto DICT.**

### Justificativa

Temporal foi escolhido pelos seguintes motivos:

#### 1. **Já em Uso no LBPay**

✅ **Temporal já é tecnologia estabelecida no LBPay**:
- Utilizado no projeto Money Moving (transações PIX)
- Cluster Temporal provisionado e operacional
- Equipe treinada e experiente (Go SDK)
- **Menor Time-to-Market** (não precisa provisionar nova stack)
- **Menor risco operacional** (tecnologia conhecida)

#### 2. **Arquitetura de Workflows Durável**

**Temporal vs Alternativas**:

| Aspecto | Temporal | AWS Step Functions | Apache Airflow | Camunda |
|---------|----------|-------------------|----------------|---------|
| **Durabilidade** | ✅ **100%** (event sourcing) | ✅ 100% (managed) | ⚠️ Depende de DB | ⚠️ Depende de DB |
| **Long-running workflows** | ✅ **Sem limites** (anos) | ❌ Máx 1 ano | ⚠️ Sim (não otimizado) | ✅ Sim |
| **Retry automático** | ✅ **Built-in** (exp backoff) | ✅ Built-in | ⚠️ Manual | ⚠️ Manual |
| **Compensação (SAGA)** | ✅ **Nativo** | ⚠️ Manual | ⚠️ Manual | ✅ Nativo |
| **Code-first** | ✅ **Go, Java, Python, etc.** | ❌ JSON DSL | ⚠️ Python DAGs | ❌ BPMN XML |
| **Visibilidade** | ✅ **UI rica** | ✅ Console AWS | ⚠️ UI simples | ✅ Cockpit |
| **Vendor lock-in** | ✅ **Open-source** | ❌ AWS-only | ✅ Open-source | ✅ Open-source |
| **Performance** | ✅ **10k+ workflows/sec** | ⚠️ 2k/sec (throttling) | ❌ Batch-focused | ⚠️ 100s/sec |
| **Custo** | ✅ **Self-hosted** (controle) | ❌ Pay-per-execution | ✅ Self-hosted | ✅ Self-hosted |

#### 3. **Workflows como Código (Code-First)**

**Temporal = Código Golang**:
- Workflows escritos em Go (mesma linguagem do projeto)
- Type-safe (erros detectados em compile-time)
- Unit tests de workflows (mock activities)
- Versionamento de workflows (deploy sem breaking changes)

**Exemplo Workflow Temporal (Go)**:
```go
func RegisterKeyWorkflow(ctx workflow.Context, req RegisterKeyRequest) error {
    // Configurar timeouts e retry policies
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 30 * time.Second,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval:    1 * time.Second,
            BackoffCoefficient: 2.0,
            MaximumInterval:    60 * time.Second,
            MaximumAttempts:    3,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    // Activity 1: Enviar CreateEntry ao RSFN
    var entryID string
    err := workflow.ExecuteActivity(ctx, SendCreateEntryActivity, req).Get(ctx, &entryID)
    if err != nil {
        // Falha: publicar evento de erro
        workflow.ExecuteActivity(ctx, PublishEventActivity, KeyRegistrationFailedEvent{...})
        return err
    }

    // Activity 2: Aguardar confirmação Bacen (polling até 2 min)
    var confirmation CreateEntryConfirmation
    err = workflow.ExecuteActivity(ctx, WaitForConfirmationActivity, entryID).Get(ctx, &confirmation)
    if err != nil {
        // Timeout ou erro Bacen
        workflow.ExecuteActivity(ctx, PublishEventActivity, KeyRegistrationFailedEvent{...})
        return err
    }

    // Activity 3: Atualizar status local
    err = workflow.ExecuteActivity(ctx, UpdateKeyStatusActivity, entryID, StatusActive).Get(ctx, nil)
    if err != nil {
        // Inconsistência crítica - compensar
        workflow.ExecuteActivity(ctx, SendAlertActivity, "Key registered in Bacen but failed to update local DB")
        return err
    }

    // Activity 4: Publicar evento de sucesso
    workflow.ExecuteActivity(ctx, PublishEventActivity, KeyRegisteredEvent{...})

    return nil
}
```

**Vantagens**:
- ✅ Lógica de negócio em código (não JSON/XML)
- ✅ IDE support (autocomplete, refactoring)
- ✅ Unit tests com mocks
- ✅ Versionamento via Git

#### 4. **Durabilidade e Confiabilidade**

**Event Sourcing Built-in**:
- Temporal armazena **TODOS os eventos** de um workflow (event log)
- Se um worker falha, outro worker continua de onde parou (sem perda de estado)
- Replay determinístico (mesmos inputs = mesmo output)
- Auditoria completa (histórico de execuções)

**Zero Data Loss**:
- Workflows NÃO são perdidos em caso de falha do worker
- Estado persistido em PostgreSQL/Cassandra (durabilidade)
- Retry automático até sucesso ou max attempts

**Exemplo de Falha e Recuperação**:
1. Workflow `RegisterKeyWorkflow` iniciado
2. Activity `SendCreateEntryActivity` executada com sucesso
3. Worker falha (crash, OOM, etc.)
4. Temporal detecta falha e reassigna workflow para outro worker
5. Novo worker continua do último ponto (Activity 2: `WaitForConfirmation`)
6. **Zero reprocessamento desnecessário** (idempotência garantida)

#### 5. **Long-Running Workflows (7 dias)**

**Temporal suporta workflows de longa duração**:
- **Timers**: Aguardar 7 dias (claim/portability) sem usar recursos
- **Signals**: Receber resposta externa (ex: PSP aceita/recusa claim)
- **Queries**: Consultar estado do workflow em tempo real
- **Cancellation**: Cancelar workflow manualmente (ops)

**Exemplo: ClaimWorkflow (7 dias)**:
```go
func ClaimWorkflow(ctx workflow.Context, req ClaimRequest) error {
    // Enviar CreateClaim ao Bacen
    var claimID string
    err := workflow.ExecuteActivity(ctx, SendCreateClaimActivity, req).Get(ctx, &claimID)
    if err != nil {
        return err
    }

    // Aguardar resposta por 7 dias (timer)
    timer := workflow.NewTimer(ctx, 7*24*time.Hour)

    // Selector: aguarda signal OU timeout
    selector := workflow.NewSelector(ctx)

    var response ClaimResponse
    responseChan := workflow.GetSignalChannel(ctx, "claim_response")

    selector.AddReceive(responseChan, func(c workflow.ReceiveChannel, more bool) {
        c.Receive(ctx, &response)
    })

    selector.AddFuture(timer, func(f workflow.Future) {
        // Timeout: confirmar automaticamente
        response = ClaimResponse{Status: "AUTO_CONFIRMED"}
    })

    selector.Select(ctx) // Bloqueia até signal ou timeout (7 dias)

    // Processar resposta
    if response.Status == "CONFIRMED" || response.Status == "AUTO_CONFIRMED" {
        workflow.ExecuteActivity(ctx, UpdateClaimStatusActivity, claimID, StatusConfirmed)
    } else {
        workflow.ExecuteActivity(ctx, UpdateClaimStatusActivity, claimID, StatusCancelled)
    }

    return nil
}
```

**Signal External (quando PSP responde)**:
```go
// Quando RSFN Connect recebe resposta Bacen, envia signal ao workflow
temporalClient.SignalWorkflow(ctx, workflowID, runID, "claim_response", ClaimResponse{Status: "CONFIRMED"})
```

**Vantagens**:
- ✅ Workflow aguarda 7 dias **sem usar CPU/memória** (apenas timer no DB)
- ✅ Signal desbloqueia workflow instantaneamente
- ✅ Timeout auto-confirma/cancela claim (regulatório)

#### 6. **Retry e Compensação (SAGA Pattern)**

**Retry Automático**:
- Configuração granular por activity
- Exponential backoff (1s, 2s, 4s, 8s, ...)
- Max attempts (ex: 3 tentativas)
- Non-retryable errors (ex: validação falhou = não retry)

**Compensação (SAGA)**:
```go
func RegisterKeyWithSAGA(ctx workflow.Context, req RegisterKeyRequest) error {
    // Activity 1: Criar entry local (PostgreSQL)
    var entryID string
    err := workflow.ExecuteActivity(ctx, CreateLocalEntryActivity, req).Get(ctx, &entryID)
    if err != nil {
        return err
    }

    // Activity 2: Enviar ao Bacen
    err = workflow.ExecuteActivity(ctx, SendCreateEntryActivity, req).Get(ctx, nil)
    if err != nil {
        // COMPENSAÇÃO: Deletar entry local (rollback)
        workflow.ExecuteActivity(ctx, DeleteLocalEntryActivity, entryID)
        return err
    }

    return nil
}
```

**Vantagens**:
- ✅ Compensação automática em caso de falha
- ✅ Consistency garantida (distributed transactions)
- ✅ Rollback explícito (não depende de DB transactions)

#### 7. **Visibilidade e Debugging**

**Temporal Web UI**:
- Visualizar workflows em execução (running, completed, failed)
- Drill-down em cada activity (input, output, erro, latência)
- Replay de workflows (debugging determinístico)
- Métricas: success rate, latência, throughput

**CLI para Ops**:
```bash
# Listar workflows
temporal workflow list --namespace lbpay-dict

# Descrever workflow específico
temporal workflow describe --workflow-id "register-key-123"

# Cancelar workflow
temporal workflow cancel --workflow-id "claim-456"

# Query workflow (estado atual)
temporal workflow query --workflow-id "claim-456" --name "getStatus"
```

**Observabilidade**:
- Métricas Prometheus exportadas (Temporal SDK)
- Logs estruturados (JSON) com correlation IDs
- Tracing distribuído (OpenTelemetry integration)

#### 8. **Cron/Scheduled Workflows**

**VSYNC Agendado**:
```go
// Registrar cron workflow (diariamente às 03:00 AM)
_, err := temporalClient.StartWorkflow(ctx, client.StartWorkflowOptions{
    ID:           "vsync-daily",
    TaskQueue:    "dict-task-queue",
    CronSchedule: "0 3 * * *",  // Diariamente às 03:00
}, VSYNCWorkflow)
```

**Vantagens**:
- ✅ Cron built-in (não precisa de cron job externo)
- ✅ Timezone-aware
- ✅ Histórico de execuções (auditoria)

---

## Consequências

### Positivas ✅

1. **Time-to-Market Reduzido**:
   - Cluster Temporal já provisionado no LBPay
   - Equipe já treinada (Go SDK)
   - Não precisa approval de nova tecnologia

2. **Consistência Tecnológica**:
   - Mesma stack do Money Moving (sinergia)
   - Redução de complexidade operacional
   - Reutilização de padrões (retry, compensação)

3. **Durabilidade 100%**:
   - Event sourcing built-in
   - Zero data loss em caso de falha
   - Replay determinístico

4. **Code-First**:
   - Workflows em Go (type-safe)
   - Unit tests de workflows
   - Versionamento via Git

5. **Long-Running Workflows**:
   - Suporte a workflows de 7 dias (claims)
   - Timers eficientes (não usam CPU/memória)
   - Signals para resposta externa

6. **Visibilidade**:
   - UI rica (drill-down em activities)
   - Histórico completo (5 anos - compliance)
   - Métricas Prometheus

7. **Retry e Compensação**:
   - Retry automático (exp backoff)
   - SAGA pattern (compensação)
   - Circuit breaker

### Negativas ❌

1. **Complexidade Operacional**:
   - Mais componentes que solução serverless (AWS Step Functions)
   - Requer expertise para tuning e troubleshooting
   - **Mitigação**: Equipe LBPay já possui expertise

2. **Curva de Aprendizado (Novos Devs)**:
   - Conceitos avançados: event sourcing, determinism, idempotency
   - Diferente de frameworks tradicionais (Go goroutines ≠ Temporal workflows)
   - **Mitigação**: Documentação interna + treinamentos

3. **Overhead de Configuração**:
   - Namespace creation
   - Task queue configuration
   - Worker deployment (Kubernetes)
   - **Mitigação**: Terraform/IaC para automação

4. **Non-Determinism Bugs**:
   - Workflows devem ser determinísticos (sem `time.Now()`, `rand.Rand()`, etc.)
   - Erros sutis se não seguir best practices
   - **Mitigação**: Linters (go-linter para Temporal), code reviews

### Riscos e Mitigações

| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| **Falha de cluster Temporal** | Baixa | Alto | Multi-node deployment, monitoramento 24/7, runbooks |
| **Performance degradation** | Média | Médio | Capacity planning, horizontal scaling (workers) |
| **Non-determinism bugs** | Média | Alto | Linters, code reviews, replay tests em staging |
| **Perda de histórico** | Muito Baixa | Crítico | Backups regulares PostgreSQL, replicação |

---

## Alternativas Consideradas

### Alternativa 1: AWS Step Functions

**Prós**:
- ✅ Managed service (zero ops)
- ✅ Integração nativa AWS
- ✅ Retry automático
- ✅ Durabilidade 100%

**Contras**:
- ❌ **Vendor lock-in** (AWS-only)
- ❌ **Não usado no LBPay** (introduziria nova stack)
- ❌ Workflows em JSON DSL (não code-first)
- ❌ Limites: máx 1 ano de duração, 25k eventos por workflow
- ❌ Custos variáveis (pay-per-execution)
- ❌ Debugging limitado (logs CloudWatch)
- ❌ Long-running workflows = caro (charged por state transition)

**Decisão**: ❌ **Rejeitado** - Lock-in, custos, Temporal já em uso

### Alternativa 2: Apache Airflow

**Prós**:
- ✅ Open-source
- ✅ UI rica
- ✅ DAGs em Python

**Contras**:
- ❌ **Não usado no LBPay**
- ❌ Focado em **batch processing** (não real-time workflows)
- ❌ Retry manual (não automático por padrão)
- ❌ Compensação manual (sem SAGA built-in)
- ❌ Long-running workflows não otimizados
- ❌ Estado em banco (não event sourcing)
- ❌ Performance limitada (100s workflows/sec)

**Decisão**: ❌ **Rejeitado** - Não adequado para real-time workflows

### Alternativa 3: Camunda

**Prós**:
- ✅ Open-source
- ✅ BPMN 2.0 (padrão)
- ✅ UI rica (Cockpit)
- ✅ SAGA pattern suportado

**Contras**:
- ❌ **Não usado no LBPay**
- ❌ Workflows em XML/BPMN (não code-first)
- ❌ Performance limitada (100s workflows/sec)
- ❌ Curva de aprendizado (BPMN modeling)
- ❌ Estado em banco (não event sourcing nativo)
- ❌ Java-centric (Go SDK limitado)

**Decisão**: ❌ **Rejeitado** - XML-based, Temporal já em uso

### Alternativa 4: Celery (Python)

**Prós**:
- ✅ Open-source
- ✅ Simples de usar
- ✅ Python

**Contras**:
- ❌ **Não usado no LBPay** (LBPay = Go)
- ❌ Task queue (não orchestrator)
- ❌ Retry limitado (não exp backoff avançado)
- ❌ Sem compensação (SAGA manual)
- ❌ Sem long-running workflows (timers limitados)
- ❌ Sem UI (visibilidade limitada)
- ❌ Estado em Redis/RabbitMQ (não event sourcing)

**Decisão**: ❌ **Rejeitado** - Inadequado para orchestration complexa

### Alternativa 5: Custom Solution (gRPC + Pulsar)

**Prós**:
- ✅ Controle total
- ✅ Go nativo

**Contras**:
- ❌ **Reinventar a roda** (event sourcing, retry, compensação)
- ❌ Tempo de desenvolvimento alto
- ❌ Bugs sutis (durabilidade, determinism)
- ❌ Sem UI (precisa construir)
- ❌ Sem tooling (debug, replay)
- ❌ Manutenção contínua

**Decisão**: ❌ **Rejeitado** - Temporal já resolve todos os requisitos

---

## Implementação

### Arquitetura Temporal para Projeto DICT

```
Temporal Cluster (LBPay Production)
│
├── Namespace: lbpay-dict
│   ├── Task Queue: dict-core-task-queue
│   │   └── Workers: Bridge Service (Go)
│   │
│   ├── Workflows:
│   │   ├── RegisterKeyWorkflow (timeout: 5 min)
│   │   ├── ClaimWorkflow (timeout: 7 dias)
│   │   ├── PortabilityWorkflow (timeout: 7 dias)
│   │   ├── DeleteKeyWorkflow (timeout: 5 min)
│   │   └── VSYNCWorkflow (cron: diariamente 03:00)
│   │
│   └── Activities:
│       ├── SendCreateEntryActivity (timeout: 30s)
│       ├── WaitForConfirmationActivity (timeout: 2 min)
│       ├── UpdateKeyStatusActivity (timeout: 5s)
│       ├── PublishEventActivity (timeout: 5s)
│       ├── SendClaimActivity (timeout: 30s)
│       └── ...
│
└── Storage: PostgreSQL (event log + state)
```

### Configuração de Workflows

#### Workflow: `RegisterKeyWorkflow`

**Descrição**: Orquestra cadastro de chave PIX (envio ao Bacen + confirmação)

**Timeout Total**: 5 minutos

**Activities**:
1. `ValidateKeyActivity` (5s)
2. `SendCreateEntryActivity` (30s, retry 3x)
3. `WaitForConfirmationActivity` (2 min, retry 3x)
4. `UpdateKeyStatusActivity` (5s, retry 3x)
5. `PublishEventActivity` (5s, retry 3x)

**Código**:
```go
package workflows

import (
    "time"
    "go.temporal.io/sdk/workflow"
)

func RegisterKeyWorkflow(ctx workflow.Context, req RegisterKeyRequest) error {
    // Configurar options
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 30 * time.Second,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval:    1 * time.Second,
            BackoffCoefficient: 2.0,
            MaximumInterval:    60 * time.Second,
            MaximumAttempts:    3,
            NonRetryableErrorTypes: []string{"ValidationError"},
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    // Activity 1: Validar chave
    err := workflow.ExecuteActivity(ctx, ValidateKeyActivity, req).Get(ctx, nil)
    if err != nil {
        return err
    }

    // Activity 2: Enviar CreateEntry ao RSFN
    var entryID string
    err = workflow.ExecuteActivity(ctx, SendCreateEntryActivity, req).Get(ctx, &entryID)
    if err != nil {
        workflow.ExecuteActivity(ctx, PublishEventActivity, KeyRegistrationFailedEvent{...})
        return err
    }

    // Activity 3: Aguardar confirmação Bacen
    var confirmation CreateEntryConfirmation
    err = workflow.ExecuteActivity(ctx, WaitForConfirmationActivity, entryID).Get(ctx, &confirmation)
    if err != nil {
        workflow.ExecuteActivity(ctx, PublishEventActivity, KeyRegistrationFailedEvent{...})
        return err
    }

    // Activity 4: Atualizar status local
    err = workflow.ExecuteActivity(ctx, UpdateKeyStatusActivity, entryID, StatusActive).Get(ctx, nil)
    if err != nil {
        workflow.ExecuteActivity(ctx, SendAlertActivity, "Key registered in Bacen but failed to update local DB")
        return err
    }

    // Activity 5: Publicar evento de sucesso
    workflow.ExecuteActivity(ctx, PublishEventActivity, KeyRegisteredEvent{...})

    return nil
}
```

#### Workflow: `ClaimWorkflow` (7 dias)

**Descrição**: Orquestra reivindicação de chave (aguarda resposta PSP por até 7 dias)

**Timeout Total**: 7 dias

**Activities**:
1. `ValidateClaimActivity` (5s)
2. `SendCreateClaimActivity` (30s, retry 3x)
3. **Timer**: Aguardar até 7 dias
4. **Signal**: Receber resposta externa (`claim_response`)
5. `UpdateClaimStatusActivity` (5s, retry 3x)

**Código**:
```go
func ClaimWorkflow(ctx workflow.Context, req ClaimRequest) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 30 * time.Second,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 3,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    // Activity 1: Validar claim
    err := workflow.ExecuteActivity(ctx, ValidateClaimActivity, req).Get(ctx, nil)
    if err != nil {
        return err
    }

    // Activity 2: Enviar CreateClaim ao Bacen
    var claimID string
    err = workflow.ExecuteActivity(ctx, SendCreateClaimActivity, req).Get(ctx, &claimID)
    if err != nil {
        workflow.ExecuteActivity(ctx, PublishEventActivity, ClaimFailedEvent{...})
        return err
    }

    // Timer: Aguardar resposta por 7 dias
    timer := workflow.NewTimer(ctx, 7*24*time.Hour)

    // Selector: aguarda signal OU timeout
    selector := workflow.NewSelector(ctx)

    var response ClaimResponse
    responseChan := workflow.GetSignalChannel(ctx, "claim_response")

    selector.AddReceive(responseChan, func(c workflow.ReceiveChannel, more bool) {
        c.Receive(ctx, &response)
    })

    selector.AddFuture(timer, func(f workflow.Future) {
        // Timeout: confirmar automaticamente (regulatório)
        response = ClaimResponse{Status: "AUTO_CONFIRMED"}
    })

    selector.Select(ctx)

    // Processar resposta
    if response.Status == "CONFIRMED" || response.Status == "AUTO_CONFIRMED" {
        workflow.ExecuteActivity(ctx, UpdateClaimStatusActivity, claimID, StatusConfirmed)
        workflow.ExecuteActivity(ctx, PublishEventActivity, ClaimConfirmedEvent{...})
    } else {
        workflow.ExecuteActivity(ctx, UpdateClaimStatusActivity, claimID, StatusCancelled)
        workflow.ExecuteActivity(ctx, PublishEventActivity, ClaimCancelledEvent{...})
    }

    return nil
}
```

#### Workflow: `VSYNCWorkflow` (Cron)

**Descrição**: Executar sincronização VSYNC diariamente às 03:00 AM

**Cron**: `0 3 * * *`

**Código**:
```go
func VSYNCWorkflow(ctx workflow.Context) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 1,  // Não retry VSYNC (executa no próximo dia)
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    // Activity 1: Calcular hash local
    var localHash string
    err := workflow.ExecuteActivity(ctx, CalculateLocalHashActivity).Get(ctx, &localHash)
    if err != nil {
        return err
    }

    // Activity 2: Enviar VSYNC ao Bacen
    var bacenHash string
    err = workflow.ExecuteActivity(ctx, SendVSYNCActivity, localHash).Get(ctx, &bacenHash)
    if err != nil {
        return err
    }

    // Activity 3: Comparar hashes
    if localHash != bacenHash {
        // Discrepância detectada
        workflow.ExecuteActivity(ctx, ReconcileDiscrepanciesActivity)
        workflow.ExecuteActivity(ctx, PublishEventActivity, VSYNCDiscrepancyDetectedEvent{...})
    } else {
        workflow.ExecuteActivity(ctx, PublishEventActivity, VSYNCCompletedEvent{...})
    }

    return nil
}
```

### Deployment de Workers

**Worker Service (Bridge)**:
```go
package main

import (
    "log"
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
)

func main() {
    // Conectar ao Temporal cluster
    c, err := client.NewClient(client.Options{
        HostPort:  "temporal.lbpay.internal:7233",
        Namespace: "lbpay-dict",
    })
    if err != nil {
        log.Fatalln("Unable to create Temporal client", err)
    }
    defer c.Close()

    // Criar worker
    w := worker.New(c, "dict-core-task-queue", worker.Options{})

    // Registrar workflows
    w.RegisterWorkflow(RegisterKeyWorkflow)
    w.RegisterWorkflow(ClaimWorkflow)
    w.RegisterWorkflow(PortabilityWorkflow)
    w.RegisterWorkflow(DeleteKeyWorkflow)
    w.RegisterWorkflow(VSYNCWorkflow)

    // Registrar activities
    w.RegisterActivity(ValidateKeyActivity)
    w.RegisterActivity(SendCreateEntryActivity)
    w.RegisterActivity(WaitForConfirmationActivity)
    w.RegisterActivity(UpdateKeyStatusActivity)
    w.RegisterActivity(PublishEventActivity)
    // ... outras activities

    // Iniciar worker
    err = w.Run(worker.InterruptCh())
    if err != nil {
        log.Fatalln("Unable to start worker", err)
    }
}
```

**Kubernetes Deployment**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dict-bridge-worker
  namespace: lbpay-dict
spec:
  replicas: 3  # 3 workers para high availability
  selector:
    matchLabels:
      app: dict-bridge-worker
  template:
    metadata:
      labels:
        app: dict-bridge-worker
    spec:
      containers:
      - name: worker
        image: lbpay/dict-bridge:v1.0.0
        env:
        - name: TEMPORAL_HOST
          value: "temporal.lbpay.internal:7233"
        - name: TEMPORAL_NAMESPACE
          value: "lbpay-dict"
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
          limits:
            cpu: 2000m
            memory: 2Gi
```

### Monitoramento

**Métricas Prometheus**:
- `temporal_workflow_start_total` (workflows iniciados)
- `temporal_workflow_completed_total` (workflows completados)
- `temporal_workflow_failed_total` (workflows falhados)
- `temporal_workflow_latency_ms` (latência end-to-end)
- `temporal_activity_execution_latency_ms` (latência por activity)
- `temporal_activity_execution_failed_total` (falhas por activity)

**Alertas**:
- Workflow failure rate > 5%
- Activity timeout > 10% (network issues)
- Workflow backlog > 1000 (workers insufficient)
- Workflow latency P95 > 10s (performance degradation)

**Dashboards Grafana**:
- Workflows por tipo (RegisterKey, Claim, etc.)
- Success rate por workflow
- Latency percentiles (P50, P95, P99)
- Activity failure breakdown

---

## Rastreabilidade

### Requisitos Funcionais Impactados

| CRF | Descrição | Workflow Temporal |
|-----|-----------|-------------------|
| [CRF-001](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-001) | Cadastrar Chave CPF | `RegisterKeyWorkflow` |
| [CRF-020](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-020) | Solicitar Claim | `ClaimWorkflow` |
| [CRF-030](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-030) | Solicitar Portabilidade | `PortabilityWorkflow` |
| [CRF-040](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-040) | Excluir Chave | `DeleteKeyWorkflow` |
| [CRF-060](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-060) | VSYNC | `VSYNCWorkflow` (cron) |

### NFRs Impactados

| NFR | Descrição | Como Temporal Atende |
|-----|-----------|---------------------|
| [NFR-090](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-090) | Durabilidade 100% | Event sourcing built-in ✅ |
| [NFR-091](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-091) | Disponibilidade 99.99% | Multi-node cluster, failover ✅ |
| [NFR-092](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-092) | Latência < 100ms (start) | Temporal: ~10ms ✅ |
| [NFR-093](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-093) | Throughput ≥ 1k wf/sec | Temporal: 10k+ wf/sec ✅ |
| [NFR-095](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-095) | Long-running (7 dias) | Timers eficientes ✅ |

### Processos BPMN Impactados

| PRO | Descrição | Workflow Temporal |
|-----|-----------|-------------------|
| [PRO-001](../04_Processos/PRO-001_Processos_BPMN.md#pro-001) | Cadastro CPF | `RegisterKeyWorkflow` |
| [PRO-006](../04_Processos/PRO-001_Processos_BPMN.md#pro-006) | Reivindicação | `ClaimWorkflow` |
| [PRO-008](../04_Processos/PRO-001_Processos_BPMN.md#pro-008) | Portabilidade | `PortabilityWorkflow` |
| [PRO-015](../04_Processos/PRO-001_Processos_BPMN.md#pro-015) | VSYNC | `VSYNCWorkflow` |

---

## Referências

### Documentação Técnica

- [Temporal Documentation](https://docs.temporal.io/)
- [Temporal Go SDK](https://github.com/temporalio/sdk-go)
- [Temporal Best Practices](https://docs.temporal.io/dev-guide/go/best-practices)
- [Event Sourcing in Temporal](https://docs.temporal.io/concepts/what-is-event-sourcing)

### Case Studies

- [DoorDash: Temporal for Order Management](https://doordash.engineering/2020/08/14/workflows-cadence-event-driven-processing/)
- [Netflix: Temporal for Media Processing](https://netflixtechblog.com/distributed-workflow-patterns-with-temporal-1a5e5e3d6c9e)
- [Stripe: Workflow Orchestration](https://stripe.com/blog/workflow-automation)

### Arquitetura LBPay

- [ArquiteturaDict_LBPAY.md](../../Docs_iniciais/ArquiteturaDict_LBPAY.md): Diagramas SVG mostrando Temporal workflows
- [ARE-001](./ARE-001_Analise_Repositorios_Existentes.md): Análise do uso de Temporal em repos existentes

---

## Aprovação

- [x] **Thiago Lima** (Head de Arquitetura) - 2025-10-24
- [x] **José Luís Silva** (CTO) - 2025-10-24

**Rationale**: Temporal Workflow já é tecnologia confirmada e em uso no LBPay. Esta ADR documenta a decisão e fundamenta o uso técnico no projeto DICT.

---

**FIM DO DOCUMENTO ADR-002**
