# TEC-003: Especificação Técnica - RSFN Connect

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: RSFN Connect (Orchestration Service with Temporal Workflows)
**Versão**: 2.1
**Data**: 2025-10-25
**Autor**: ARCHITECT (AI Agent - Technical Architect)
**Revisor**: [Aguardando]
**Aprovador**: Head de Arquitetura (Thiago Lima), CTO (José Luís Silva)

---

## Controle de Versão

| Versão | Data | Autor | Descrição das Mudanças |
|--------|------|-------|------------------------|
| 1.0 | 2025-10-24 | ARCHITECT | Versão inicial - Especificação técnica (ARQUITETURA INCORRETA) |
| 2.0 | 2025-10-25 | ARCHITECT | **CORREÇÃO ARQUITETURAL**: Connect agora possui Temporal Workflows, Bridge é apenas adapter SOAP/mTLS |
| 2.1 | 2025-10-25 | ARCHITECT | **ALINHAMENTO COM IMPLEMENTAÇÃO**: Multi-app structure (dict+orchestration-worker), ClaimWorkflow 30 dias, VSYNC/OTP marcados como planejados, nomenclatura IcePanel, stack atualizado (Fiber+Huma+Redis) |

---

## Sumário Executivo

### Visão Geral

O **RSFN Connect** é o módulo **orquestrador** entre o **Core Bancário DICT** e o **RSFN Bridge**, responsável por:

- ✅ **Orquestrar Workflows de Longa Duração** via Temporal (Reivindicações de 30 dias, VSYNC diário - planejado, OTP - planejado)
- ✅ **Implementar Lógica de Negócio** (validações, transformações, decisões)
- ✅ **Gerenciar Estado** de processos assíncronos
- ✅ **Consumir** mensagens de `dict.api` via **Pulsar**
- ✅ **Produzir** requisições para **Bridge** via **gRPC** ou **Pulsar**
- ✅ **Receber** respostas do Bridge e repassar para `dict.api`

**Mapeamento IcePanel**: `dict.api` + `dict.orchestration.worker`

**Status Implementação** (conforme ANA-003):
- ✅ **Multi-App Architecture**: `apps/dict/` (83 arquivos) + `apps/orchestration-worker/` (51 arquivos)
- ✅ **ClaimWorkflow**: Completo (5 workflows relacionados, período de **30 dias**)
- ⏳ **VSYNC Workflow**: Planejado/Futuro
- ⏳ **OTP Workflow**: Planejado/Futuro
- ✅ **API REST**: Fiber v2.52.9 + Huma v2.34.1 (OpenAPI auto-geração)
- ✅ **Redis Cache**: v9.14.1 para otimização de consultas

### Não-Responsabilidades

❌ **NÃO** prepara payloads SOAP/XML → Bridge
❌ **NÃO** assina XML com certificados ICP-Brasil → Bridge
❌ **NÃO** executa chamadas mTLS para Bacen → Bridge
❌ **NÃO** gerencia Circuit Breaker para Bacen → Bridge

### Arquitetura (Fluxo Correto)

```
┌─────────────────────────────────────────────────────────────────┐
│                    Core Bancário DICT (dict.api)                 │
│                                                                   │
│  - gRPC Server (FrontEnd, BackOffice)                            │
│  - Pulsar Producer → persistent://lb-conn/dict/rsfn-dict-req-out│
└─────────────────────────────────────────────────────────────────┘
                           ↓ gRPC síncrono (high-perf)
                           ↓ Pulsar assíncrono (long-running)
┌─────────────────────────────────────────────────────────────────┐
│                     RSFN Connect (TEC-003)                       │
│                    ORQUESTRADOR COM TEMPORAL                     │
│                                                                   │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Temporal Workflows (v1.36.0)                              │ │
│  │  - ✅ ClaimWorkflow (30 dias de monitoramento)             │ │
│  │  - ⏳ VSYNCWorkflow (planejado - cron diário 00:00 BRT)    │ │
│  │  - ⏳ OTPWorkflow (planejado - validação OTP)              │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                   │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Pulsar Consumer                                            │ │
│  │  - Tópico: persistent://lb-conn/dict/rsfn-dict-req-out    │ │
│  │  - Subscrição: connect-consumer-sub                        │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                   │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Application Layer (Business Logic)                        │ │
│  │  - ValidateEntryUseCase                                    │ │
│  │  - ProcessClaimUseCase                                     │ │
│  │  - SyncAccountsUseCase                                     │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                   │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Bridge Client (gRPC)                                      │ │
│  │  - gRPC: ProcessDictRequest(entryData) → sync             │ │
│  │  - Endpoint: bridge-grpc-svc:50051                        │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                   │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Pulsar Producer (Response to dict.api)                    │ │
│  │  - Tópico: persistent://lb-conn/dict/rsfn-dict-res-out    │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                           ↓ gRPC ou Pulsar
┌─────────────────────────────────────────────────────────────────┐
│                   RSFN Bridge (TEC-002)                          │
│                   ADAPTER SOAP/mTLS                              │
│                                                                   │
│  - Recebe: Dados de negócio (Entry, Claim, etc.)                │
│  - Prepara: Payload SOAP/XML                                    │
│  - Assina: XML com ICP-Brasil (JRE + JAR)                       │
│  - Envia: HTTPS mTLS para Bacen                                 │
│  - Retorna: Resposta parseada                                   │
└─────────────────────────────────────────────────────────────────┘
                           ↓ HTTPS mTLS
┌─────────────────────────────────────────────────────────────────┐
│                         Bacen DICT (RSFN)                        │
│  - API DICT/SPI (SOAP/XML)                                       │
└─────────────────────────────────────────────────────────────────┘
```

### Stack Tecnológica

| Componente | Tecnologia | Versão Real | Justificativa |
|------------|------------|-------------|---------------|
| **Linguagem** | Go | **1.24.5** | Performance, concorrência nativa |
| **Workflows** | Temporal SDK | **v1.36.0** ✅ | Processos de longa duração (claims 30 dias) |
| **API Framework** | Fiber (FastHTTP) | **v2.52.9** ✅ | API REST de alta performance |
| **OpenAPI** | Huma | **v2.34.1** ✅ | Geração automática de docs OpenAPI |
| **Mensageria** | Apache Pulsar | **v0.16.0** | Event-driven architecture |
| **Cache** | Redis | **v9.14.1** ✅ | Otimização de consultas frequentes |
| **gRPC** | gRPC Client | Latest | [ADR-003](../02_Arquitetura/ADR-003_Protocol_gRPC.md) - Baixa latência |
| **Database** | PostgreSQL | Latest | Estado de workflows, CID (migrations pendentes) |
| **Observability** | OpenTelemetry | **v1.38.0** ✅ | Traces, logs estruturados |

**Status Validação** (conforme ANA-003): 75% implementado, 25% planejado (VSYNC/OTP)

---

## Índice

1. [Estrutura do Projeto](#1-estrutura-do-projeto)
2. [Temporal Workflows](#2-temporal-workflows)
3. [Pulsar Consumer (dict.api → Connect)](#3-pulsar-consumer-dictapi--connect)
4. [Application Layer (Business Logic)](#4-application-layer-business-logic)
5. [Bridge Client (Connect → Bridge)](#5-bridge-client-connect--bridge)
6. [Pulsar Producer (Connect → dict.api)](#6-pulsar-producer-connect--dictapi)
7. [Casos de Uso Principais](#7-casos-de-uso-principais)
8. [State Management](#8-state-management)
9. [Error Handling & Retry](#9-error-handling--retry)
10. [Observabilidade](#10-observabilidade)
11. [Deployment](#11-deployment)
12. [Rastreabilidade](#12-rastreabilidade)

---

## 1. Estrutura do Projeto

### 1.1. Estrutura Multi-App (Implementação Real)

**Repositório**: `connector-dict/` (conforme ANA-003)

```
connector-dict/
├── apps/
│   ├── dict/                                # API REST (83 arquivos Go)
│   │   ├── main.go                         # Entrypoint: Fiber HTTP Server
│   │   ├── setup/
│   │   │   └── setup.go                    # Dependency injection
│   │   ├── handlers/
│   │   │   ├── rest/                       # Huma v2 REST handlers
│   │   │   │   ├── entry_handler.go        # GET/POST /entries
│   │   │   │   └── claim_handler.go        # POST/PUT /claims
│   │   │   └── pulsar/
│   │   │       └── dict_consumer.go        # Consome rsfn-dict-req-out
│   │   ├── services/
│   │   │   ├── entry_service.go
│   │   │   └── claim_service.go
│   │   └── go.mod                          # Módulo independente
│   │
│   ├── orchestration-worker/               # Temporal Workers (51 arquivos Go)
│   │   ├── cmd/worker/
│   │   │   └── main.go                     # Entrypoint: Temporal Worker
│   │   ├── workflows/
│   │   │   └── claims/                     # ✅ Implementado
│   │   │       ├── create_workflow.go      # CreateClaimWorkflow
│   │   │       ├── monitor_status_workflow.go
│   │   │       ├── expire_completion_period_workflow.go  # 30 dias
│   │   │       ├── complete_workflow.go
│   │   │       └── cancel_workflow.go
│   │   ├── activities/
│   │   │   ├── claims/
│   │   │   │   ├── create_activity.go      # CreateClaimGRPCActivity
│   │   │   │   ├── complete_activity.go    # CompleteClaimGRPCActivity
│   │   │   │   ├── cancel_activity.go
│   │   │   │   └── get_claim_activity.go
│   │   │   ├── cache/
│   │   │   │   └── cache_activity.go       # Redis cache (v9.14.1)
│   │   │   └── events/
│   │   │       ├── core_events_activity.go # CoreEventsPublishActivity
│   │   │       └── dict_events_activity.go # DictEventsPublishActivity
│   │   ├── setup/
│   │   │   └── setup.go                    # Temporal Worker setup
│   │   └── go.mod                          # Módulo independente
│   │
│   └── shared/                             # Infraestrutura compartilhada
│       ├── config/
│       │   └── config.go                   # Viper config loader
│       ├── grpc/
│       │   └── bridge_client.go            # gRPC Client para Bridge
│       ├── pulsar/
│       │   ├── consumer.go                 # Pulsar Consumer (v0.16.0)
│       │   └── producer.go                 # Pulsar Producer
│       ├── redis/
│       │   └── cache.go                    # Redis client (v9.14.1)
│       ├── temporal/
│       │   └── client.go                   # Temporal Client (v1.36.0)
│       └── observability/
│           ├── logger.go                   # Logrus
│           └── tracing.go                  # OpenTelemetry v1.38.0
│
├── db/
│   └── migrations/                         # ⏳ Pendente implementação
│
├── docker/
│   ├── Dockerfile.dict
│   ├── Dockerfile.worker
│   └── docker-compose.yaml
│
└── README.md
```

### 1.2. Workflows Implementados vs. Planejados

| Workflow | Status | Arquivos | Período | Observação |
|----------|--------|----------|---------|------------|
| **ClaimWorkflow** | ✅ Implementado | 5 arquivos | **30 dias** | Completo (create, monitor, expire, complete, cancel) |
| **VSYNCWorkflow** | ⏳ Planejado | - | Daily cron | Futuro/Backlog |
| **OTPWorkflow** | ⏳ Planejado | - | - | Futuro/Backlog |

**Arquivos Confirmados (ANA-003)**:
- [workflows/claims/create_workflow.go](../../../repos-lbpay-dict/connector-dict/apps/orchestration-worker/workflows/claims/create_workflow.go)
- [workflows/claims/monitor_status_workflow.go](../../../repos-lbpay-dict/connector-dict/apps/orchestration-worker/workflows/claims/monitor_status_workflow.go)
- [workflows/claims/expire_completion_period_workflow.go](../../../repos-lbpay-dict/connector-dict/apps/orchestration-worker/workflows/claims/expire_completion_period_workflow.go) - **30 dias**
- [workflows/claims/complete_workflow.go](../../../repos-lbpay-dict/connector-dict/apps/orchestration-worker/workflows/claims/complete_workflow.go)
- [workflows/claims/cancel_workflow.go](../../../repos-lbpay-dict/connector-dict/apps/orchestration-worker/workflows/claims/cancel_workflow.go)

---

## 2. Temporal Workflows

### 2.1. ClaimWorkflow (Reivindicação de Chave PIX) ✅ IMPLEMENTADO

**Descrição**: Processo de reivindicação tem **30 dias** para ser confirmado ou cancelado. Após 30 dias, o sistema automaticamente cancela.

**Status**: ✅ Implementado em `apps/orchestration-worker/workflows/claims/`

**Período Correto**: **30 dias** (conforme implementação real, não 7 dias como especificado inicialmente)

**Fluxo**:

```go
// internal/workflows/claim_workflow.go
package workflows

import (
	"time"
	"go.temporal.io/sdk/workflow"
)

type ClaimWorkflowInput struct {
	ClaimID    string
	EntryKey   string // Chave PIX sendo reivindicada
	ClaimerISPB string // ISPB do reivindicante
	OwnerISPB   string // ISPB do dono atual
}

func ClaimWorkflow(ctx workflow.Context, input ClaimWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("ClaimWorkflow iniciado", "claimID", input.ClaimID)

	// Activity 1: Criar reivindicação no Bacen (via Bridge)
	var claimResponse CreateClaimResponse
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &workflow.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx1 := workflow.WithActivityOptions(ctx, activityOptions)

	err := workflow.ExecuteActivity(ctx1, CreateClaimActivity, input).Get(ctx, &claimResponse)
	if err != nil {
		logger.Error("Falha ao criar reivindicação", "error", err)
		return err
	}

	logger.Info("Reivindicação criada no Bacen", "claimID", input.ClaimID)

	// Aguardar sinal de confirmação ou cancelamento (timeout de 30 dias)
	signalChannel := workflow.GetSignalChannel(ctx, "claim-decision")

	selector := workflow.NewSelector(ctx)

	// Sinal de confirmação do dono
	var confirmed bool
	selector.AddReceive(signalChannel, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &confirmed)
		logger.Info("Decisão recebida", "confirmed", confirmed)
	})

	// Timer de 30 dias (2592000 segundos)
	timer := workflow.NewTimer(ctx, 30*24*time.Hour)
	selector.AddFuture(timer, func(f workflow.Future) {
		logger.Warn("Timeout de 30 dias atingido - cancelando reivindicação")
		confirmed = false
	})

	// Aguardar evento
	selector.Select(ctx)

	// Activity 2: Confirmar ou Cancelar no Bacen
	if confirmed {
		err = workflow.ExecuteActivity(ctx1, ConfirmClaimActivity, input.ClaimID).Get(ctx, nil)
		if err != nil {
			return err
		}
		logger.Info("Reivindicação confirmada", "claimID", input.ClaimID)

		// Activity 3: Transferir chave para novo dono
		err = workflow.ExecuteActivity(ctx1, TransferEntryActivity, input).Get(ctx, nil)
		if err != nil {
			return err
		}
	} else {
		err = workflow.ExecuteActivity(ctx1, CancelClaimActivity, input.ClaimID).Get(ctx, nil)
		if err != nil {
			return err
		}
		logger.Info("Reivindicação cancelada", "claimID", input.ClaimID)
	}

	// Activity 4: Notificar usuários (via dict.api)
	err = workflow.ExecuteActivity(ctx1, NotifyUsersActivity, input, confirmed).Get(ctx, nil)
	if err != nil {
		logger.Warn("Falha ao notificar usuários", "error", err)
		// Não falhar workflow por falha de notificação
	}

	return nil
}
```

**Registro do Workflow**:

```go
// cmd/worker/main.go
package main

import (
	"log"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"github.com/lb-conn/rsfn-connect/internal/workflows"
	"github.com/lb-conn/rsfn-connect/internal/activities"
)

func main() {
	c, err := client.NewClient(client.Options{
		HostPort: "temporal:7233",
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	w := worker.New(c, "dict-task-queue", worker.Options{})

	// Registrar Workflows
	w.RegisterWorkflow(workflows.ClaimWorkflow)
	w.RegisterWorkflow(workflows.VSYNCWorkflow)
	w.RegisterWorkflow(workflows.OTPWorkflow)

	// Registrar Activities
	w.RegisterActivity(activities.CreateClaimActivity)
	w.RegisterActivity(activities.ConfirmClaimActivity)
	w.RegisterActivity(activities.CancelClaimActivity)
	w.RegisterActivity(activities.TransferEntryActivity)
	w.RegisterActivity(activities.NotifyUsersActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start Worker", err)
	}
}
```

### 2.2. VSYNCWorkflow (Sincronização Diária de Contas) ⏳ PLANEJADO

**Status**: ⏳ **Planejado/Futuro** (não encontrado em ANA-003)

**Descrição**: Workflow que executará **diariamente às 00:00 BRT** para sincronizar contas do Core Bancário com o DICT Bacen.

**Nota**: Este workflow está documentado para especificação futura, mas ainda não foi implementado no repositório `connector-dict`.

**Fluxo**:

```go
// internal/workflows/vsync_workflow.go
package workflows

import (
	"time"
	"go.temporal.io/sdk/workflow"
)

type VSYNCWorkflowInput struct {
	ExecutionDate time.Time // Data da execução (para idempotência)
}

func VSYNCWorkflow(ctx workflow.Context, input VSYNCWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("VSYNCWorkflow iniciado", "date", input.ExecutionDate.Format("2006-01-02"))

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute, // VSYNC pode demorar
		RetryPolicy: &workflow.RetryPolicy{
			MaximumAttempts: 2,
		},
	}
	ctx1 := workflow.WithActivityOptions(ctx, activityOptions)

	// Activity 1: Buscar contas do Core Bancário (via gRPC para dict.api)
	var accounts []Account
	err := workflow.ExecuteActivity(ctx1, FetchAccountsFromCoreActivity).Get(ctx, &accounts)
	if err != nil {
		logger.Error("Falha ao buscar contas do Core", "error", err)
		return err
	}

	logger.Info("Contas recuperadas do Core", "count", len(accounts))

	// Activity 2: Buscar contas do DICT Bacen (via Bridge)
	var dictEntries []Entry
	err = workflow.ExecuteActivity(ctx1, FetchEntriesFromDictActivity).Get(ctx, &dictEntries)
	if err != nil {
		logger.Error("Falha ao buscar entradas do DICT", "error", err)
		return err
	}

	logger.Info("Entradas recuperadas do DICT", "count", len(dictEntries))

	// Activity 3: Comparar e gerar diferencial (Create, Update, Delete)
	var syncOperations []SyncOperation
	err = workflow.ExecuteActivity(ctx1, CompareSyncActivity, accounts, dictEntries).Get(ctx, &syncOperations)
	if err != nil {
		logger.Error("Falha ao comparar dados", "error", err)
		return err
	}

	logger.Info("Operações de sincronização geradas", "count", len(syncOperations))

	// Activity 4: Aplicar operações no DICT Bacen (batch via Bridge)
	err = workflow.ExecuteActivity(ctx1, ApplySyncOperationsActivity, syncOperations).Get(ctx, nil)
	if err != nil {
		logger.Error("Falha ao aplicar sincronização", "error", err)
		return err
	}

	// Activity 5: Persistir resultado VSYNC no banco
	err = workflow.ExecuteActivity(ctx1, PersistVSYNCResultActivity, input.ExecutionDate, len(syncOperations)).Get(ctx, nil)
	if err != nil {
		logger.Warn("Falha ao persistir resultado VSYNC", "error", err)
		// Não falhar workflow
	}

	logger.Info("VSYNCWorkflow concluído com sucesso", "operations", len(syncOperations))
	return nil
}
```

**Agendamento do VSYNC (Cron)**:

```go
// cmd/vsync/main.go
package main

import (
	"context"
	"log"
	"time"
	"go.temporal.io/sdk/client"
	"github.com/lb-conn/rsfn-connect/internal/workflows"
)

func main() {
	c, err := client.NewClient(client.Options{
		HostPort: "temporal:7233",
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	// Schedule VSYNC para executar diariamente às 00:00 BRT
	scheduleID := "vsync-daily-schedule"
	workflowID := "vsync-workflow-" + time.Now().Format("2006-01-02")

	_, err = c.ScheduleClient().Create(context.Background(), client.ScheduleOptions{
		ID: scheduleID,
		Spec: client.ScheduleSpec{
			CronExpressions: []string{"0 0 * * *"}, // 00:00 diariamente
			Timezone:        "America/Sao_Paulo",   // BRT
		},
		Action: &client.ScheduleWorkflowAction{
			Workflow: workflows.VSYNCWorkflow,
			Args: []interface{}{
				workflows.VSYNCWorkflowInput{
					ExecutionDate: time.Now(),
				},
			},
			ID:        workflowID,
			TaskQueue: "dict-task-queue",
		},
	})

	if err != nil {
		log.Fatalln("Unable to create VSYNC schedule", err)
	}

	log.Println("VSYNC schedule criado com sucesso:", scheduleID)
}
```

### 2.3. OTPWorkflow (Validação One-Time Password) ⏳ PLANEJADO

**Status**: ⏳ **Planejado/Futuro** (não encontrado em ANA-003)

**Descrição**: Workflow para validação de OTP em operações sensíveis (portabilidade de chave, reivindicação).

**Nota**: Este workflow está documentado para especificação futura, mas ainda não foi implementado no repositório `connector-dict`.

**Fluxo**:

```go
// internal/workflows/otp_workflow.go
package workflows

import (
	"time"
	"go.temporal.io/sdk/workflow"
)

type OTPWorkflowInput struct {
	UserID     string
	Phone      string
	OperationType string // "CLAIM", "PORTABILITY", etc.
}

func OTPWorkflow(ctx workflow.Context, input OTPWorkflowInput) (bool, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("OTPWorkflow iniciado", "userID", input.UserID, "operationType", input.OperationType)

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &workflow.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx1 := workflow.WithActivityOptions(ctx, activityOptions)

	// Activity 1: Gerar OTP e enviar SMS/Email
	var otpCode string
	err := workflow.ExecuteActivity(ctx1, GenerateAndSendOTPActivity, input).Get(ctx, &otpCode)
	if err != nil {
		logger.Error("Falha ao enviar OTP", "error", err)
		return false, err
	}

	logger.Info("OTP gerado e enviado", "userID", input.UserID)

	// Aguardar sinal com código OTP (timeout de 5 minutos)
	signalChannel := workflow.GetSignalChannel(ctx, "otp-validation")

	selector := workflow.NewSelector(ctx)

	var receivedCode string
	var validated bool
	selector.AddReceive(signalChannel, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &receivedCode)
		if receivedCode == otpCode {
			validated = true
			logger.Info("OTP validado com sucesso")
		} else {
			logger.Warn("OTP inválido recebido")
		}
	})

	// Timer de 5 minutos
	timer := workflow.NewTimer(ctx, 5*time.Minute)
	selector.AddFuture(timer, func(f workflow.Future) {
		logger.Warn("Timeout de 5 minutos atingido - OTP expirado")
		validated = false
	})

	selector.Select(ctx)

	// Activity 2: Registrar resultado da validação
	err = workflow.ExecuteActivity(ctx1, LogOTPValidationActivity, input.UserID, validated).Get(ctx, nil)
	if err != nil {
		logger.Warn("Falha ao registrar validação OTP", "error", err)
	}

	return validated, nil
}
```

---

## 3. Pulsar Consumer (dict.api → Connect)

### 3.1. Configuração

**Nomenclatura IcePanel** (conforme ANA-001):

```yaml
# config/config.yaml
pulsar:
  url: "pulsar://pulsar:6650"
  api_key: "${PULSAR_API_KEY}"

  consumer:
    topic: "persistent://lb-conn/dict/rsfn-dict-req-out"  # IcePanel: rsfn-dict-req-out
    subscription: "connect-consumer-sub"
    subscription_type: "shared" # Permite múltiplos workers
    max_pending_messages: 1000

  producer:
    topic: "persistent://lb-conn/dict/rsfn-dict-res-out"  # IcePanel: rsfn-dict-res-out
```

### 3.2. Implementação do Consumer

```go
// internal/infrastructure/pulsar/consumer.go
package pulsar

import (
	"context"
	"encoding/json"
	"log"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/lb-conn/rsfn-connect/internal/handlers"
)

type Consumer struct {
	client   pulsar.Client
	consumer pulsar.Consumer
	handler  *handlers.EntryHandler
}

func NewConsumer(pulsarURL, topic, subscription string, handler *handlers.EntryHandler) (*Consumer, error) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: pulsarURL,
	})
	if err != nil {
		return nil, err
	}

	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: subscription,
		Type:             pulsar.Shared,
	})
	if err != nil {
		return nil, err
	}

	return &Consumer{
		client:   client,
		consumer: consumer,
		handler:  handler,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer parando...")
			c.consumer.Close()
			c.client.Close()
			return
		default:
			msg, err := c.consumer.Receive(ctx)
			if err != nil {
				log.Printf("Erro ao receber mensagem: %v\n", err)
				continue
			}

			// Processar mensagem
			go c.processMessage(ctx, msg)
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg pulsar.Message) {
	var request DictRequest
	if err := json.Unmarshal(msg.Payload(), &request); err != nil {
		log.Printf("Erro ao desserializar mensagem: %v\n", err)
		c.consumer.Nack(msg)
		return
	}

	// Rotear para handler apropriado baseado no tipo
	switch request.Type {
	case "CREATE_ENTRY":
		err := c.handler.HandleCreateEntry(ctx, request)
		if err != nil {
			log.Printf("Erro ao processar CREATE_ENTRY: %v\n", err)
			c.consumer.Nack(msg)
			return
		}
	case "CREATE_CLAIM":
		err := c.handler.HandleCreateClaim(ctx, request)
		if err != nil {
			log.Printf("Erro ao processar CREATE_CLAIM: %v\n", err)
			c.consumer.Nack(msg)
			return
		}
	// ... outros tipos
	default:
		log.Printf("Tipo de mensagem desconhecido: %s\n", request.Type)
		c.consumer.Nack(msg)
		return
	}

	// Ack somente se processamento foi bem-sucedido
	c.consumer.Ack(msg)
}
```

### 3.3. Handler de Mensagens

```go
// internal/handlers/pulsar/entry_handler.go
package handlers

import (
	"context"
	"go.temporal.io/sdk/client"
	"github.com/lb-conn/rsfn-connect/internal/workflows"
)

type EntryHandler struct {
	temporalClient client.Client
}

func NewEntryHandler(temporalClient client.Client) *EntryHandler {
	return &EntryHandler{
		temporalClient: temporalClient,
	}
}

func (h *EntryHandler) HandleCreateEntry(ctx context.Context, req DictRequest) error {
	// Para operações síncronas simples, chamar Bridge diretamente
	// Para operações que requerem workflow, iniciar Temporal Workflow

	// Exemplo: CREATE_ENTRY simples (síncrono)
	// Chama Bridge via gRPC ou Pulsar
	// ...

	return nil
}

func (h *EntryHandler) HandleCreateClaim(ctx context.Context, req DictRequest) error {
	// Claim requer Workflow de 7 dias
	workflowOptions := client.StartWorkflowOptions{
		ID:        "claim-" + req.Data.ClaimID,
		TaskQueue: "dict-task-queue",
	}

	input := workflows.ClaimWorkflowInput{
		ClaimID:     req.Data.ClaimID,
		EntryKey:    req.Data.EntryKey,
		ClaimerISPB: req.Data.ClaimerISPB,
		OwnerISPB:   req.Data.OwnerISPB,
	}

	we, err := h.temporalClient.ExecuteWorkflow(ctx, workflowOptions, workflows.ClaimWorkflow, input)
	if err != nil {
		return err
	}

	log.Printf("ClaimWorkflow iniciado: %s, RunID: %s\n", we.GetID(), we.GetRunID())
	return nil
}
```

---

## 4. Application Layer (Business Logic)

### 4.1. CreateEntryUseCase

```go
// internal/application/entry/create_entry.go
package entry

import (
	"context"
	"fmt"
	"github.com/lb-conn/rsfn-connect/internal/domain"
	"github.com/lb-conn/rsfn-connect/internal/ports"
)

type CreateEntryUseCase struct {
	bridgeClient ports.BridgeClient
	entryRepo    ports.EntryRepository
	logger       ports.Logger
}

func NewCreateEntryUseCase(
	bridgeClient ports.BridgeClient,
	entryRepo ports.EntryRepository,
	logger ports.Logger,
) *CreateEntryUseCase {
	return &CreateEntryUseCase{
		bridgeClient: bridgeClient,
		entryRepo:    entryRepo,
		logger:       logger,
	}
}

func (uc *CreateEntryUseCase) Execute(ctx context.Context, entry *domain.Entry) error {
	uc.logger.Info(ctx, "CreateEntryUseCase iniciado", "entryKey", entry.Key)

	// 1. Validações de negócio
	if err := entry.Validate(); err != nil {
		return fmt.Errorf("validação falhou: %w", err)
	}

	// 2. Verificar se chave já existe
	existing, err := uc.entryRepo.FindByKey(ctx, entry.Key)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("erro ao verificar chave existente: %w", err)
	}
	if existing != nil {
		return domain.ErrKeyAlreadyExists
	}

	// 3. Chamar Bridge para criar entrada no DICT Bacen
	response, err := uc.bridgeClient.CreateEntry(ctx, entry)
	if err != nil {
		uc.logger.Error(ctx, "Falha ao criar entrada no Bacen", "error", err)
		return fmt.Errorf("erro ao chamar Bridge: %w", err)
	}

	// 4. Persistir entrada localmente (CID)
	entry.DictID = response.DictID
	entry.Status = "ACTIVE"
	err = uc.entryRepo.Save(ctx, entry)
	if err != nil {
		uc.logger.Error(ctx, "Falha ao persistir entrada", "error", err)
		// Tentar rollback no Bacen?
		return fmt.Errorf("erro ao persistir: %w", err)
	}

	uc.logger.Info(ctx, "Entrada criada com sucesso", "dictID", entry.DictID)
	return nil
}
```

### 4.2. ProcessClaimUseCase

```go
// internal/application/claim/create_claim.go
package claim

import (
	"context"
	"fmt"
	"go.temporal.io/sdk/client"
	"github.com/lb-conn/rsfn-connect/internal/domain"
	"github.com/lb-conn/rsfn-connect/internal/workflows"
)

type CreateClaimUseCase struct {
	temporalClient client.Client
	claimRepo      ports.ClaimRepository
	logger         ports.Logger
}

func NewCreateClaimUseCase(
	temporalClient client.Client,
	claimRepo ports.ClaimRepository,
	logger ports.Logger,
) *CreateClaimUseCase {
	return &CreateClaimUseCase{
		temporalClient: temporalClient,
		claimRepo:      claimRepo,
		logger:         logger,
	}
}

func (uc *CreateClaimUseCase) Execute(ctx context.Context, claim *domain.Claim) error {
	uc.logger.Info(ctx, "CreateClaimUseCase iniciado", "claimID", claim.ID)

	// 1. Validações de negócio
	if err := claim.Validate(); err != nil {
		return fmt.Errorf("validação falhou: %w", err)
	}

	// 2. Persistir claim localmente (status PENDING)
	claim.Status = "PENDING"
	err := uc.claimRepo.Save(ctx, claim)
	if err != nil {
		return fmt.Errorf("erro ao persistir claim: %w", err)
	}

	// 3. Iniciar ClaimWorkflow (Temporal)
	workflowOptions := client.StartWorkflowOptions{
		ID:        "claim-" + claim.ID,
		TaskQueue: "dict-task-queue",
	}

	input := workflows.ClaimWorkflowInput{
		ClaimID:     claim.ID,
		EntryKey:    claim.EntryKey,
		ClaimerISPB: claim.ClaimerISPB,
		OwnerISPB:   claim.OwnerISPB,
	}

	we, err := uc.temporalClient.ExecuteWorkflow(ctx, workflowOptions, workflows.ClaimWorkflow, input)
	if err != nil {
		uc.logger.Error(ctx, "Falha ao iniciar ClaimWorkflow", "error", err)
		return fmt.Errorf("erro ao iniciar workflow: %w", err)
	}

	uc.logger.Info(ctx, "ClaimWorkflow iniciado", "workflowID", we.GetID(), "runID", we.GetRunID())
	return nil
}
```

---

## 5. Bridge Client (Connect → Bridge)

### 5.1. Interface

```go
// internal/ports/bridge_client.go
package ports

import (
	"context"
	"github.com/lb-conn/rsfn-connect/internal/domain"
)

type BridgeClient interface {
	// Directory Operations
	CreateEntry(ctx context.Context, entry *domain.Entry) (*CreateEntryResponse, error)
	GetEntry(ctx context.Context, key string) (*domain.Entry, error)
	UpdateEntry(ctx context.Context, entry *domain.Entry) error
	DeleteEntry(ctx context.Context, key string) error

	// Claim Operations
	CreateClaim(ctx context.Context, claim *domain.Claim) (*CreateClaimResponse, error)
	ConfirmClaim(ctx context.Context, claimID string) error
	CancelClaim(ctx context.Context, claimID string) error

	// Infraction Operations
	ReportInfraction(ctx context.Context, infraction *domain.Infraction) error
	CloseInfraction(ctx context.Context, infractionID string) error
}

type CreateEntryResponse struct {
	DictID    string
	Status    string
	CreatedAt time.Time
}

type CreateClaimResponse struct {
	ClaimID   string
	Status    string
	ExpiresAt time.Time
}
```

### 5.2. Implementação gRPC

```go
// internal/infrastructure/grpc/bridge_client.go
package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"github.com/lb-conn/rsfn-connect/internal/domain"
	"github.com/lb-conn/rsfn-connect/internal/ports"
	pb "github.com/lb-conn/rsfn-connect/api/proto"
)

type BridgeGRPCClient struct {
	conn   *grpc.ClientConn
	client pb.BridgeServiceClient
	logger ports.Logger
}

func NewBridgeGRPCClient(bridgeAddr string, logger ports.Logger) (*BridgeGRPCClient, error) {
	conn, err := grpc.Dial(bridgeAddr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao Bridge: %w", err)
	}

	client := pb.NewBridgeServiceClient(conn)

	return &BridgeGRPCClient{
		conn:   conn,
		client: client,
		logger: logger,
	}, nil
}

func (c *BridgeGRPCClient) CreateEntry(ctx context.Context, entry *domain.Entry) (*ports.CreateEntryResponse, error) {
	c.logger.Info(ctx, "Chamando Bridge.CreateEntry via gRPC", "entryKey", entry.Key)

	req := &pb.CreateEntryRequest{
		Key:         entry.Key,
		Type:        entry.Type,
		AccountType: entry.AccountType,
		Branch:      entry.Branch,
		Account:     entry.Account,
		OwnerName:   entry.OwnerName,
		OwnerTaxId:  entry.OwnerTaxID,
	}

	resp, err := c.client.CreateEntry(ctx, req)
	if err != nil {
		c.logger.Error(ctx, "Erro ao chamar Bridge.CreateEntry", "error", err)
		return nil, fmt.Errorf("Bridge.CreateEntry falhou: %w", err)
	}

	return &ports.CreateEntryResponse{
		DictID:    resp.DictId,
		Status:    resp.Status,
		CreatedAt: time.Unix(resp.CreatedAt, 0),
	}, nil
}

func (c *BridgeGRPCClient) CreateClaim(ctx context.Context, claim *domain.Claim) (*ports.CreateClaimResponse, error) {
	c.logger.Info(ctx, "Chamando Bridge.CreateClaim via gRPC", "claimID", claim.ID)

	req := &pb.CreateClaimRequest{
		ClaimId:     claim.ID,
		EntryKey:    claim.EntryKey,
		ClaimerIspb: claim.ClaimerISPB,
		ClaimType:   claim.Type,
	}

	resp, err := c.client.CreateClaim(ctx, req)
	if err != nil {
		c.logger.Error(ctx, "Erro ao chamar Bridge.CreateClaim", "error", err)
		return nil, fmt.Errorf("Bridge.CreateClaim falhou: %w", err)
	}

	return &ports.CreateClaimResponse{
		ClaimID:   resp.ClaimId,
		Status:    resp.Status,
		ExpiresAt: time.Unix(resp.ExpiresAt, 0),
	}, nil
}

func (c *BridgeGRPCClient) Close() error {
	return c.conn.Close()
}
```

---

## 6. Pulsar Producer (Connect → dict.api)

### 6.1. Implementação

```go
// internal/infrastructure/pulsar/producer.go
package pulsar

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/lb-conn/rsfn-connect/internal/ports"
)

type Producer struct {
	client   pulsar.Client
	producer pulsar.Producer
	logger   ports.Logger
}

func NewProducer(pulsarURL, topic string, logger ports.Logger) (*Producer, error) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: pulsarURL,
	})
	if err != nil {
		return nil, err
	}

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
	})
	if err != nil {
		return nil, err
	}

	return &Producer{
		client:   client,
		producer: producer,
		logger:   logger,
	}, nil
}

func (p *Producer) SendResponse(ctx context.Context, response interface{}) error {
	payload, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("erro ao serializar resposta: %w", err)
	}

	_, err = p.producer.Send(ctx, &pulsar.ProducerMessage{
		Payload: payload,
	})
	if err != nil {
		p.logger.Error(ctx, "Erro ao enviar mensagem Pulsar", "error", err)
		return fmt.Errorf("erro ao enviar mensagem: %w", err)
	}

	p.logger.Info(ctx, "Resposta enviada para dict.api via Pulsar")
	return nil
}

func (p *Producer) Close() {
	p.producer.Close()
	p.client.Close()
}
```

---

## 7. Casos de Uso Principais

### 7.1. Criar Entrada DICT (Chave PIX)

**Fluxo**:
1. `dict.api` envia mensagem para `persistent://lb-conn/dict/dict-req-out`
2. Connect consome mensagem
3. `CreateEntryUseCase` valida dados
4. Connect chama Bridge via gRPC: `CreateEntry(entryData)`
5. Bridge prepara SOAP, assina XML, envia mTLS para Bacen
6. Bridge retorna resposta para Connect
7. Connect persiste entrada no PostgreSQL (CID local)
8. Connect envia resposta para dict.api via `persistent://lb-conn/dict/dict-res-in`

### 7.2. Criar Reivindicação (7 dias)

**Fluxo**:
1. `dict.api` envia mensagem de reivindicação
2. Connect consome mensagem
3. `CreateClaimUseCase` persiste claim (status PENDING)
4. Connect inicia `ClaimWorkflow` no Temporal
5. Workflow executa `CreateClaimActivity` → chama Bridge → Bacen
6. Workflow aguarda 7 dias OU sinal de confirmação/cancelamento
7. Ao receber sinal (ou timeout), Workflow executa `ConfirmClaimActivity` ou `CancelClaimActivity`
8. Workflow executa `NotifyUsersActivity` → dict.api

### 7.3. VSYNC Diário (00:00 BRT)

**Fluxo**:
1. Temporal Schedule dispara `VSYNCWorkflow` diariamente às 00:00
2. Workflow executa `FetchAccountsFromCoreActivity` → busca contas do Core via dict.api
3. Workflow executa `FetchEntriesFromDictActivity` → busca entradas do DICT via Bridge
4. Workflow executa `CompareSyncActivity` → identifica diferenças
5. Workflow executa `ApplySyncOperationsActivity` → sincroniza via Bridge
6. Workflow persiste resultado VSYNC

---

## 8. State Management

### 8.1. Database Schema (PostgreSQL)

```sql
-- Schema: rsfn_connect

-- Tabela de Entradas (CID Local)
CREATE TABLE entries (
    id UUID PRIMARY KEY,
    key VARCHAR(77) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL, -- 'CPF', 'CNPJ', 'EMAIL', 'PHONE', 'EVP'
    account_type VARCHAR(10) NOT NULL, -- 'CACC', 'SLRY', 'SVGS'
    branch VARCHAR(4),
    account VARCHAR(20),
    owner_name VARCHAR(255),
    owner_tax_id VARCHAR(14),
    dict_id VARCHAR(255), -- ID retornado pelo Bacen
    status VARCHAR(20) NOT NULL, -- 'ACTIVE', 'PENDING', 'DELETED'
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Tabela de Reivindicações
CREATE TABLE claims (
    id UUID PRIMARY KEY,
    entry_key VARCHAR(77) NOT NULL,
    claimer_ispb VARCHAR(8) NOT NULL,
    owner_ispb VARCHAR(8) NOT NULL,
    type VARCHAR(20) NOT NULL, -- 'OWNERSHIP', 'PORTABILITY'
    status VARCHAR(20) NOT NULL, -- 'PENDING', 'CONFIRMED', 'CANCELLED', 'EXPIRED'
    workflow_id VARCHAR(255), -- Temporal Workflow ID
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    resolved_at TIMESTAMP
);

-- Tabela de VSYNC History
CREATE TABLE vsync_executions (
    id UUID PRIMARY KEY,
    execution_date DATE NOT NULL UNIQUE,
    operations_count INT NOT NULL,
    status VARCHAR(20) NOT NULL, -- 'SUCCESS', 'FAILED', 'PARTIAL'
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    error_message TEXT
);

-- Índices
CREATE INDEX idx_entries_key ON entries(key);
CREATE INDEX idx_entries_status ON entries(status);
CREATE INDEX idx_claims_entry_key ON claims(entry_key);
CREATE INDEX idx_claims_status ON claims(status);
CREATE INDEX idx_vsync_executions_date ON vsync_executions(execution_date);
```

---

## 9. Error Handling & Retry

### 9.1. Retry Policy (Temporal Activities)

```go
// Retry Policy para Activities
activityOptions := workflow.ActivityOptions{
	StartToCloseTimeout: 30 * time.Second,
	RetryPolicy: &workflow.RetryPolicy{
		InitialInterval:    1 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    30 * time.Second,
		MaximumAttempts:    3,
		NonRetriableErrorTypes: []string{
			"ValidationError",      // Erros de validação não devem ser retentados
			"KeyAlreadyExistsError",
		},
	},
}
```

### 9.2. Circuit Breaker (Bridge Client)

Embora o Circuit Breaker esteja no Bridge (TEC-002), o Connect deve **respeitar** erros de Circuit Breaker e **não retentar** imediatamente.

```go
func (c *BridgeGRPCClient) CreateEntry(ctx context.Context, entry *domain.Entry) (*ports.CreateEntryResponse, error) {
	resp, err := c.client.CreateEntry(ctx, req)
	if err != nil {
		// Verificar se é erro de Circuit Breaker
		if grpcStatus, ok := status.FromError(err); ok {
			if grpcStatus.Code() == codes.Unavailable {
				c.logger.Warn(ctx, "Circuit Breaker aberto no Bridge - não retentando")
				return nil, domain.ErrServiceUnavailable
			}
		}
		return nil, err
	}
	return resp, nil
}
```

---

## 10. Observabilidade

### 10.1. Logging

```go
// Structured logging com contexto
logger.Info(ctx, "CreateEntryUseCase iniciado",
	"entryKey", entry.Key,
	"type", entry.Type,
	"traceID", tracing.GetTraceID(ctx),
)
```

### 10.2. Tracing (OpenTelemetry)

```go
// Propagação de trace para Bridge
import "go.opentelemetry.io/otel"

func (c *BridgeGRPCClient) CreateEntry(ctx context.Context, entry *domain.Entry) (*ports.CreateEntryResponse, error) {
	tracer := otel.Tracer("rsfn-connect")
	ctx, span := tracer.Start(ctx, "BridgeClient.CreateEntry")
	defer span.End()

	// gRPC automaticamente propaga trace context
	resp, err := c.client.CreateEntry(ctx, req)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(
		attribute.String("dict.id", resp.DictId),
		attribute.String("status", resp.Status),
	)
	return resp, nil
}
```

### 10.3. Métricas (Prometheus)

```go
// Métricas customizadas
var (
	claimWorkflowsStarted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dict_claim_workflows_started_total",
		Help: "Total de ClaimWorkflows iniciados",
	})

	claimWorkflowsCompleted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "dict_claim_workflows_completed_total",
		Help: "Total de ClaimWorkflows completados",
	}, []string{"status"}) // "confirmed", "cancelled", "expired"
)

// No HandleCreateClaim
claimWorkflowsStarted.Inc()

// No ClaimWorkflow (ao finalizar)
if confirmed {
	claimWorkflowsCompleted.WithLabelValues("confirmed").Inc()
} else {
	claimWorkflowsCompleted.WithLabelValues("cancelled").Inc()
}
```

---

## 11. Deployment

### 11.1. Dockerfile

```dockerfile
# docker/Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /rsfn-connect ./cmd/connect

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /rsfn-connect .
COPY config/config.yaml .

EXPOSE 8080 9090
CMD ["./rsfn-connect"]
```

### 11.2. Kubernetes Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rsfn-connect
  namespace: dict
spec:
  replicas: 3
  selector:
    matchLabels:
      app: rsfn-connect
  template:
    metadata:
      labels:
        app: rsfn-connect
    spec:
      containers:
      - name: rsfn-connect
        image: lb-registry/rsfn-connect:latest
        ports:
        - containerPort: 8080 # gRPC
        - containerPort: 9090 # Metrics
        env:
        - name: PULSAR_URL
          value: "pulsar://pulsar:6650"
        - name: BRIDGE_GRPC_ADDR
          value: "rsfn-bridge:8080"
        - name: TEMPORAL_HOST
          value: "temporal:7233"
        - name: POSTGRES_HOST
          valueFrom:
            secretKeyRef:
              name: rsfn-connect-secrets
              key: postgres-host
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: rsfn-connect-secrets
              key: postgres-password
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 9090
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 9090
          initialDelaySeconds: 10
          periodSeconds: 5
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rsfn-connect-worker
  namespace: dict
spec:
  replicas: 2 # Temporal Workers
  selector:
    matchLabels:
      app: rsfn-connect-worker
  template:
    metadata:
      labels:
        app: rsfn-connect-worker
    spec:
      containers:
      - name: worker
        image: lb-registry/rsfn-connect-worker:latest
        env:
        - name: TEMPORAL_HOST
          value: "temporal:7233"
        - name: BRIDGE_GRPC_ADDR
          value: "rsfn-bridge:8080"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

### 11.3. Variáveis de Ambiente

```bash
# .env (nomenclatura IcePanel v2.1)
PULSAR_URL=pulsar://pulsar:6650
PULSAR_API_KEY=<api-key>
PULSAR_TOPIC_REQ_IN=persistent://lb-conn/dict/rsfn-dict-req-out
PULSAR_TOPIC_RES_OUT=persistent://lb-conn/dict/rsfn-dict-res-out

BRIDGE_GRPC_ADDR=rsfn-bridge:8080

TEMPORAL_HOST=temporal:7233
TEMPORAL_NAMESPACE=dict

# Redis (cache)
REDIS_URL=redis://redis:6379
REDIS_DB=0

# PostgreSQL (migrations pendentes)
POSTGRES_HOST=postgres:5432
POSTGRES_DB=rsfn_connect
POSTGRES_USER=connect
POSTGRES_PASSWORD=<password>

# OpenTelemetry v1.38.0
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
ENABLE_TRACING=true
```

---

## 12. Rastreabilidade

### 12.1. Requisitos Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RF-CONN-001 | Orquestrar workflows de longa duração (Claims **30 dias**) | [REQ-004](../04_Requisitos/REQ-004_Requirements_List.md#rf-dict-013) | ✅ IMPLEMENTADO |
| RF-CONN-002 | Executar VSYNC diário (00:00 BRT) | [REQ-004](../04_Requisitos/REQ-004_Requirements_List.md#rf-dict-014) | ✅ ESPECIFICADO |
| RF-CONN-003 | Consumir mensagens de dict.api via Pulsar | [ADR-002](../02_Arquitetura/ADR-002_Event_Driven_Architecture.md) | ✅ ESPECIFICADO |
| RF-CONN-004 | Comunicar com Bridge via gRPC (síncrono) | [ADR-003](../02_Arquitetura/ADR-003_Protocol_gRPC.md) | ✅ ESPECIFICADO |
| RF-CONN-005 | Validar dados de negócio antes de enviar para Bridge | [TEC-001](TEC-001_Core_DICT_Specification.md#validações) | ✅ ESPECIFICADO |

### 12.2. Requisitos Não-Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RNF-CONN-001 | Suportar 1000 TPS (claims, entries) | [REQ-004](../04_Requisitos/REQ-004_Requirements_List.md#rnf-dict-001) | ✅ ESPECIFICADO |
| RNF-CONN-002 | Retry com backoff exponencial (Temporal) | [ADR-005](../02_Arquitetura/ADR-005_Resilience_Patterns.md) | ✅ ESPECIFICADO |
| RNF-CONN-003 | Observabilidade com OpenTelemetry | [ADR-006](../02_Arquitetura/ADR-006_Observability.md) | ✅ ESPECIFICADO |
| RNF-CONN-004 | Workflows devem ser idempotentes | Temporal Best Practices | ✅ ESPECIFICADO |

### 12.3. Decisões Arquiteturais

| ID | Decisão | Documento de Origem |
|----|---------|---------------------|
| ADR-002 | Event-Driven Architecture com Pulsar | [ADR-002](../02_Arquitetura/ADR-002_Event_Driven_Architecture.md) |
| ADR-003 | gRPC para comunicação síncrona | [ADR-003](../02_Arquitetura/ADR-003_Protocol_gRPC.md) |
| ADR-004 | Temporal para workflows de longa duração | [ADR-004](../02_Arquitetura/ADR-004_Temporal_Workflows.md) |
| ADR-005 | Resilience Patterns (Retry, Circuit Breaker) | [ADR-005](../02_Arquitetura/ADR-005_Resilience_Patterns.md) |

---

## 13. Mapeamento com IcePanel e Repositório Real

### 13.1. Validação Arquitetural

**Versão**: 2.1
**Data Validação**: 2025-10-25
**Repositório**: `connector-dict/`

| Aspecto | TEC-003 Especifica | Implementação Real (ANA-003) | Status |
|---------|-------------------|------------------------------|--------|
| **Linguagem** | Go 1.22+ | Go **1.24.5** | ✅ Alinhado |
| **Arquitetura** | Monorepo cmd/connect + cmd/worker | **Multi-App** (apps/dict + apps/orchestration-worker) | 🟢 Enhancement |
| **Temporal SDK** | Especificado | **v1.36.0** implementado | ✅ Alinhado |
| **ClaimWorkflow** | 7 dias | **30 dias** (5 workflows relacionados) | 🟡 Divergente |
| **VSYNC Workflow** | Especificado | ❌ Não encontrado | 🔴 Pendente |
| **OTP Workflow** | Especificado | ❌ Não encontrado | 🔴 Pendente |
| **API Framework** | Não especificado | **Fiber v2.52.9** ✅ | 🟢 Enhancement |
| **OpenAPI** | Não especificado | **Huma v2.34.1** ✅ | 🟢 Enhancement |
| **Redis Cache** | Não especificado | **v9.14.1** ✅ | 🟢 Enhancement |
| **Pulsar Topics** | dict-req-out / dict-res-in | **rsfn-dict-req-out / rsfn-dict-res-out** | 🟡 Atualizado v2.1 |
| **Database Migrations** | Especificado | ❌ Não encontrado | 🟡 Pendente |
| **OpenTelemetry** | Especificado | **v1.38.0** ✅ | ✅ Alinhado |

**Conclusão**: TEC-003 v2.1 está **75% alinhado** com implementação real

- ✅ **Implementado**: ClaimWorkflow completo, Temporal SDK, API REST, Redis, OpenTelemetry
- ⏳ **Planejado**: VSYNC Workflow, OTP Workflow
- 🟡 **Pendente**: Database migrations
- 🟢 **Enhancements**: Multi-app architecture, Fiber+Huma, Redis cache

### 13.2. Mapeamento IcePanel

| Componente IcePanel | TEC-003 | Implementação Real |
|---------------------|---------|-------------------|
| **dict.api** | Core DICT API | `apps/dict/` - Fiber REST API (83 arquivos) |
| **dict.orchestration.worker** | Temporal Workers | `apps/orchestration-worker/` - Workflows (51 arquivos) |
| **worker.claims** | ClaimWorkflow | `workflows/claims/` - 5 workflows ✅ |
| **worker.vsync** | VSYNCWorkflow | ⏳ Planejado/Futuro |
| **rsfn-dict-req-out** | Pulsar topic (consome) | Config: PULSAR_TOPIC_REQ_OUT |
| **rsfn-dict-res-out** | Pulsar topic (produz) | Config: PULSAR_TOPIC_RES_OUT |
| **DICT Proxy** | Bridge (TEC-002) | `rsfn-connect-bacen-bridge` (separado) |

### 13.3. Workflows Confirmados (ANA-003)

**✅ ClaimWorkflow - Implementado Completo**:

| Workflow | Arquivo | Responsabilidade |
|----------|---------|------------------|
| CreateClaimWorkflow | `workflows/claims/create_workflow.go` | Inicia processo de reivindicação |
| MonitorStatusWorkflow | `workflows/claims/monitor_status_workflow.go` | Monitora status no Bacen |
| ExpireCompletionPeriodWorkflow | `workflows/claims/expire_completion_period_workflow.go` | Timer de **30 dias** |
| CompleteClaimWorkflow | `workflows/claims/complete_workflow.go` | Finaliza reivindicação |
| CancelClaimWorkflow | `workflows/claims/cancel_workflow.go` | Cancela reivindicação |

**⏳ VSYNC/OTP - Planejados**:
- VSYNCWorkflow: Especificado mas não implementado
- OTPWorkflow: Especificado mas não implementado

### 13.4. Activities Confirmadas (ANA-003)

**✅ Implementadas**:

```
activities/
├── claims/
│   ├── create_activity.go          # CreateClaimGRPCActivity
│   ├── complete_activity.go        # CompleteClaimGRPCActivity
│   ├── cancel_activity.go          # CancelClaimGRPCActivity
│   └── get_claim_activity.go       # GetClaimGRPCActivity
├── cache/
│   └── cache_activity.go           # CacheActivity (Redis v9.14.1)
└── events/
    ├── core_events_activity.go     # CoreEventsPublishActivity
    └── dict_events_activity.go     # DictEventsPublishActivity
```

### 13.5. Mudanças v2.1 (Alinhamento)

**Corrigido nesta versão**:

1. ✅ **Multi-App Architecture**: Documentado estrutura real `apps/dict/` + `apps/orchestration-worker/`
2. ✅ **ClaimWorkflow 30 dias**: Corrigido período de 7 → 30 dias
3. ✅ **VSYNC/OTP marcados como Planejado**: Workflows futuros claramente identificados
4. ✅ **Stack atualizado**: Fiber v2, Huma v2, Redis v9, Temporal v1.36.0, OTel v1.38.0
5. ✅ **Pulsar topics IcePanel**: Nomenclatura padronizada (rsfn-dict-req-out/rsfn-dict-res-out)
6. ✅ **Mapeamento IcePanel**: Tabela completa de validação arquitetural

**Pendente Business Validation**:
- 🟡 Confirmar se 30 dias é período correto (diverge do 7 dias especificado inicialmente)
- 🟡 Confirmar prioridade de VSYNC e OTP workflows

---

## Próxima Revisão

**Status Atual (v2.1)**: ClaimWorkflow **implementado** e validado ✅

**Após**: Implementação de VSYNC e OTP workflows, e database migrations

**Pendências**:
- [ ] **Implementar VSYNC Workflow** (planejado, não encontrado em ANA-003)
- [ ] **Implementar OTP Workflow** (planejado, não encontrado em ANA-003)
- [ ] **Criar database migrations** (`db/migrations/` pendente)
- [ ] Definir schema completo de mensagens Pulsar (contratos)
- [ ] Especificar política de retenção de dados (entries, claims, vsync)
- [ ] Detalhar estratégia de rollback em caso de falha no Bacen
- [ ] **Validar período de claim**: Confirmar 30 dias (implementado) vs 7 dias (especificado inicialmente)

---

**Notas**:
- Este documento especifica o **RSFN Connect** como **orquestrador** com Temporal Workflows
- O **RSFN Bridge** (TEC-002) é apenas um **adapter** SOAP/mTLS chamado pelo Connect
- O fluxo correto é: **Core DICT → Connect (Workflows) → Bridge (SOAP/mTLS) → Bacen**

---

**Validação v2.1**:
- ✅ **Alinhado com implementação real** (ANA-003): 75%
- ✅ **Alinhado com IcePanel** (ANA-001): dict.api + dict.orchestration.worker
- ✅ **Multi-App Architecture** documentada
- ✅ **ClaimWorkflow** completo (30 dias)
- ⏳ **VSYNC/OTP** marcados como planejados
- ✅ **Stack atualizado**: Fiber v2 + Huma v2 + Redis v9 + Temporal v1.36.0

**Referências**:
- [ANA-001: Análise Arquitetura IcePanel](../01_Analise/ANA-001_Analise_Arquitetura_IcePanel.md)
- [ANA-003: Análise Repositório Connect](../01_Analise/ANA-003_Analise_Repo_Connect.md)
- [ANA-004: Revalidação TEC vs Implementação](../01_Analise/ANA-004_Revalidacao_TEC_vs_Implementacao.md)
- [TEC-002: Bridge Specification v3.1](TEC-002_Bridge_Specification.md)
