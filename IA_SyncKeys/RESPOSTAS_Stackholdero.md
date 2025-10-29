# Respostas às Dúvidas Técnicas - Projeto DICT CID/VSync

**Data**: 2025-10-28
**Status**: ✅ RESOLVIDAS com base na documentação do Connector-Dict
**Fonte**: Documentos em `.claude/Specs_do_Stackholder/`

---

## ✅ Dúvidas Respondidas com Base na Documentação

### 1. Eventos do Connector-Dict ✅ RESPONDIDA

**Dúvida Original**: O connector-dict já emite eventos Pulsar quando cria/altera chaves?

**Resposta**: **SIM!** Conforme `instrucoes-app-dict.md` e `instrucoes-orchestration-worker.md`:

**Arquitetura Atual**:
```
Dict API (apps/dict) → Pulsar → Orchestration Worker (apps/orchestration-worker)
```

**Padrão Existente**:
- **Dict API** publica eventos Pulsar para operações assíncronas (POST/PUT/DELETE)
- **Orchestration Worker** consome eventos e executa workflows Temporal
- **Topics Pulsar**: Nomeação padrão `persistent://lb-conn/dict/orchestration-worker-<action>-<resource>`

**Para o Projeto CID/VSync**:

Reutilizar Topics Existentes.

- o Connector-Dict publica eventos de alteração de estado, tanto de Entry quanto Claims, no tópico PULSAR_TOPIC_DICT_EVENTS=persistent://lb-conn/dict/dict-events e o conteúdo deste evento deve conter informações suficientes para conseguir calcular o CID.
- Adicionar um novo consumer para ler estes eventos do tópico PULSAR_TOPIC_DICT_EVENTS=persistent://lb-conn/dict/dict-events existente.
- Para calcular o CID, o consumer que será adicionado deve estar dentro do app dict.vsync, que é um container separado, exclusivo para as atividades de cálculo de CID e verificação de sincronismo. Este dict.vsync pode ter workflows Temporal para as atividades recorrentes de verificação de sincronismo, como o POST /api/v2/sync-verifications/

---

### 2. Schema PostgreSQL ✅ VALIDADO

**Dúvida Original**: Qual deve ser o schema exato da tabela de CIDs?

**Resposta**: O schema proposto em DUVIDAS_TECNICAS.md está **INCORRETO** e precisa ser ajustado.

**Adições Necessárias** (baseado em padrões observados):
1. **Migration**: Usar Goose (verificar em `/connector-dict/apps/dict` se usa migrations)
2. **Repository Pattern**: Criar repository em `apps/dict/infrastructure/database/repositories/`
3. **Conexão PostgreSQL**: Já existe infraestrutura de DB no connector-dict (verificar `setup/`)

**Ajustes necessários aos scripts**

- As colunas `created_at` e `updated_at` devem ser `NOT NULL`, porém sem default.
- Não devem existir triggers

Razão: A intenção em ter as colunas NOT NULL e sem valor default é para que o sistema seja obrigado a informar estes valores no momento em que estiver executando. Desta maneira garantimos que os horários, que são extremamente importantes para consistência, que aparecem em logs e trilhas de auditoria são exatamente os mesmos. Se deixarmos os horários como DEFAULT NOW(), ou com triggers, corremos o risco do horário do servidor do banco de dados flutuar para mais ou para menos, e com isso perderíamos a correlação com as demais partes do sistema (logs, auditoria, traces, etc). Os pods / containers que compõe o sistema devem ter seus horários internos sincronizados com o horário do Cluster Kubernetes, que por sua vez, deve ter seu horário sincronizado com o Banco Central.

---

### 3. Integração com Core-Dict ✅ RESPONDIDA

**Dúvida Original**: Como exatamente o Core-Dict deve ser informado em caso de dessincronização?

**Resposta**: **Via Pulsar Events** (padrão do projeto!)

**Padrão Observado no Connector-Dict**:
1. **CoreEvents**: Eventos internos para Core-Dict
2. **DictEvents**: Eventos externos

**Implementação para CID/VSync**:

```go
// infrastructure/temporal/workflows/sync/reconciliation_workflow.go

func ReconciliationWorkflow(ctx workflow.Context, input ReconciliationInput) error {
    // ... após detectar divergências ...

    // 1. Publish to CoreEvents (notificar Core-Dict)
    err := workflows.ExecuteCoreEventsPublishActivity(
        ctx,
        input.RequestID,
        pkg.ActionSyncReconciliationRequired,
        ReconciliationPayload{
            KeyType: input.KeyType,
            DivergenceCount: divergences.Total,
            DictCIDFileURL: dictFileURL,
            ActionRequired: "REBUILD_TABLE",
        },
    )

    // 2. Publish to DictEvents (auditoria externa)
    err = workflows.ExecuteDictEventsPublishActivity(
        ctx,
        input.RequestID,
        pkg.ActionSyncReconciliationRequired,
        divergences,
    )

    return nil
}
```

**Topics Necessários**:
```bash
PULSAR_TOPIC_CORE_EVENTS=persistent://lb-conn/dict/core-events
PULSAR_TOPIC_DICT_EVENTS=persistent://lb-conn/dict/dict-events
```

**Ação**: Core-Dict deve ter consumer para topic `core-events` que processa `ActionSyncReconciliationRequired`.

---

### 4. Periodicidade da Verificação VSync ✅ DEFINIDA

**Dúvida Original**: Qual deve ser a frequência das verificações de sincronismo?

**Resposta**: **Workflow Temporal Cron-Based** (padrão do projeto!)

**Implementação**:
```go
// setup/temporal.go - Registrar Cron Workflow

func (s *Setup) StartCronWorkflows(ctx context.Context) error {
    // VSync Verification - Diário às 03:00 AM
    _, err := s.temporalClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
        ID:           "vsync-daily-verification",
        TaskQueue:    s.taskQueue,
        CronSchedule: "0 3 * * *", // 03:00 AM todos os dias
    }, sync.VSyncVerificationWorkflow)

    return err
}
```

**Configuração**:
```bash
VSYNC_VERIFICATION_CRON="0 3 * * *"  # Diário às 03:00
VSYNC_VERIFICATION_ENABLED=true
```

**Flexibilidade**: Cron pode ser ajustado via env var sem rebuild.

---

### 5. Temporal Worker vs Serviço Standalone ✅ DEFINIDO

**Dúvida Original**: O serviço de VSync deve ser implementado como?

**Resposta**: **Temporal Workflow + Activities no Orchestration Worker** ⭐

**Arquitetura Confirmada**:
```
apps/
├── dict/                          # API REST (Huma)
│   └── (não precisa modificar para VSync periódico)
└── orchestration-worker/          # Temporal Worker
    ├── infrastructure/
    │   └── temporal/
    │       ├── workflows/
    │       │   └── sync/
    │       │       ├── vsync_verification_workflow.go  # Cron diário
    │       │       └── reconciliation_workflow.go      # On-demand
    │       └── activities/
    │           └── sync/
    │               ├── bridge_grpc_activity.go         # Call Bridge
    │               ├── database_activity.go            # Read/Write CIDs
    │               └── core_events_activity.go         # Notify Core-Dict
    └── handlers/
        └── pulsar/
            └── sync/
                └── key_event_handler.go                # Consume key.created/updated
```

**Razões**:
1. ✅ **Retry Automático**: Temporal gerencia falhas
2. ✅ **Histórico Completo**: Auditoria de verificações
3. ✅ **Observability Built-in**: Logs, traces, metrics
4. ✅ **Cron Native**: Scheduler integrado
5. ✅ **Consistência**: Mesma stack do projeto

---

### 6. Tratamento de Dessincronização ✅ DEFINIDO

**Dúvida Original**: O que fazer quando detectamos dessincronização?

**Resposta**: **Workflow de Reconciliação Automático com Notificação**

**Fluxo Definido**:

```go
// VSync Verification Workflow (diário)
if divergence_detected {
    // Trigger child workflow
    workflow.ExecuteChildWorkflow(ctx, ReconciliationWorkflow, keyType)

    // Alert PagerDuty/Slack (via activity)
    workflow.ExecuteActivity(ctx, SendAlertActivity, Alert{
        Severity: "WARNING",
        Title: fmt.Sprintf("VSync Divergence Detected: %s", keyType),
        VSyncLocal: local,
        VSyncDict: dict,
    })
}

// Reconciliation Workflow (child)
func ReconciliationWorkflow(ctx workflow.Context, keyType string) error {
    // 1. Request CID list from DICT (via Bridge)
    // 2. Download and parse file
    // 3. Compute divergences
    // 4. Notify Core-Dict (Pulsar event)
    // 5. Log results in database

    // Manual intervention required?
    if divergenceCount > 1000 {
        return RequireManualApprovalError
    }

    return nil
}
```

**Limites de Automação** (configuráveis):
```bash
VSYNC_AUTO_RECONCILE_MAX_DIVERGENCES=100  # Máx divergências para auto-reconcile
VSYNC_RECONCILE_REQUIRE_APPROVAL=false    # Se true, sempre requer aprovação manual
```

---

### 7. Variáveis de Ambiente ✅ DEFINIDAS

**Dúvida Original**: Quais são as credenciais PostgreSQL e configurações necessárias?

**Resposta**: **Baseado em Padrões do Connector-Dict**

**Verificar em** `apps/dict/setup/config.go` e `apps/orchestration-worker/setup/config.go`:

```bash
# PostgreSQL (provavelmente já existe conexão)
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=connector_dict
POSTGRES_USER=postgres
POSTGRES_PASSWORD=secret
POSTGRES_SSL_MODE=disable
POSTGRES_MAX_CONNECTIONS=25
POSTGRES_MAX_IDLE_CONNECTIONS=5

# Pulsar
PULSAR_URL=pulsar://localhost:6650
PULSAR_TOPIC_DICT_KEY_CREATED=persistent://lb-conn/dict/dict-key-created
PULSAR_TOPIC_DICT_KEY_UPDATED=persistent://lb-conn/dict/dict-key-updated
PULSAR_TOPIC_CORE_EVENTS=persistent://lb-conn/dict/core-events
PULSAR_TOPIC_DICT_EVENTS=persistent://lb-conn/dict/dict-events

# Bridge gRPC (já deve existir no connector-dict)
BRIDGE_GRPC_HOST=localhost
BRIDGE_GRPC_PORT=50051
# mTLS provavelmente já configurado no grpcGateway

# VSync Configuration
VSYNC_VERIFICATION_CRON="0 3 * * *"
VSYNC_VERIFICATION_ENABLED=true
VSYNC_AUTO_RECONCILE_MAX_DIVERGENCES=100
VSYNC_TIMEOUT_SECONDS=300

# ISPB Participante (já deve existir)
ISPB_PARTICIPANTE=12345678
```

**Ação**: Verificar `apps/dict/.env.example` e `apps/orchestration-worker/.env.example` para valores atuais.

---

### 8. Normalização de Dados ✅ DEFINIDA

**Dúvida Original**: O connector-dict já normaliza dados conforme algoritmo CID?

**Resposta**: **SIM, uma alteração de estado não aceita pelo Banco Central se não estiver com os dados normalizados. Como os eventos só serão disparados quando a alteração de estado for com sucesso, isso garante que os dados contidos no evento estão de acordo com as regras.**

**Verificar**:
1. `apps/dict/domain/` - Verificar entidades `Entry`, `Key`, `Account`, `Owner`
2. `github.com/lb-conn/sdk-rsfn-validator` - SDK compartilhado com tipos BACEN

**Implementação CID Generator**:
```go
// apps/dict/domain/sync/cid_generator.go

import (
    "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/bacen/entry"
)

func GenerateCID(e *entry.Entry) (string, error) {
    // Reutilizar normalização existente do connector-dict
    // Aplicar algoritmo SHA-256 conforme Specs.md
}
```

**Ação**: Analisar código existente de `Entry` para reutilizar normalização.

---

### 9. Bridge Integration ✅ ESCLARECIDA

**Dúvida Original**: Como funciona a integração com DICT BACEN?

**Resposta**: **Via Bridge gRPC (JÁ EXISTE!)** 🎉

**Arquitetura Atual**:
```
Connector-Dict → Bridge (gRPC) → DICT BACEN (REST API)
```

**Código Existente**:
- `apps/dict/infrastructure/grpc/` - Clientes gRPC para Bridge
- `apps/orchestration-worker/infrastructure/grpc/` - gRPC Gateway

**Para VSync, adicionar**:
```go
// apps/orchestration-worker/infrastructure/grpc/sync/

type SyncClient struct {
    bridgeClient pb.DICTSyncServiceClient
}

func (c *SyncClient) VerifySync(ctx context.Context, vsyncs map[string]string) (*VerifySyncResponse, error) {
    // Call Bridge gRPC → Bridge calls DICT REST API
}

func (c *SyncClient) RequestCIDList(ctx context.Context, keyType string) (*RequestCIDListResponse, error) {
    // Call Bridge gRPC → Bridge calls DICT REST API
}
```

**Proto Definitions**: Adicionar em `shared/proto/sync/dict_sync_service.proto` (ver Specs.md)

**Ação**:
1. Verificar se Bridge já tem endpoints de VSync implementados
2. Se não, coordenar com time do Bridge para adicionar

---

### 10. DICT Sandbox ✅ ESCLARECIDA

**Dúvida Original**: Existe ambiente sandbox do DICT BACEN para testes?

**Resposta**: **Sim, existe a possibilidade de execução, em ambiente de desenvolvimento e de qa integrado à um Simulador Bacen (programa interno da LB). Também existe a possibilidade, em ambiente de qa integrado tanto como o Simulador Bacen (programa interno da LB) quanto com o ambiente oficial de homologação do Banco Central.**

**Ambientes Típicos**:
```bash
# Development
BRIDGE_GRPC_HOST=bridge-dev.lb-conn.local
DICT_ENVIRONMENT=development

# QA
BRIDGE_GRPC_HOST=bridge-qa.lb-conn.local
DICT_ENVIRONMENT=qa

# Staging
BRIDGE_GRPC_HOST=bridge-staging.lb-conn.local
DICT_ENVIRONMENT=staging

# Production
BRIDGE_GRPC_HOST=bridge-prod.lb-conn.local
DICT_ENVIRONMENT=production
```

**OpenAPI Spec**: `/Users/jose.silva.lb/LBPay/IA_SyncKeys/.claude/Specs_do_Stackholder/OpenAPI_Dict_Bacen.json`

**Ação**: Consultar time de Bridge para:
1. Endpoints disponíveis em cada ambiente
2. Dados de teste disponíveis
3. Limitações de sandbox (rate limits, features)

---

## 📋 Próximos Passos ATUALIZADOS

### Fase 0: Validação Técnica (2-3 dias)

1. **Analisar Código Existente do Connector-Dict**:
   - [ ] Verificar eventos Pulsar para Entry (key.created, key.updated)
   - [ ] Analisar estrutura de Entry/Key para reutilizar normalização
   - [ ] Verificar conexão PostgreSQL existente
   - [ ] Analisar gRPC Gateway e clients existentes

2. **Coordenação com Times**:
   - [ ] **Time Bridge**: Validar se endpoints VSync/CIDList existem ou precisam ser criados
   - [ ] **Time Core-Dict**: Confirmar consumer para `core-events` topic
   - [ ] **Time Infra**: Confirmar ambientes DICT disponíveis

3. **Atualizar Documentação**:
   - [x] RESPOSTAS_DUVIDAS.md (este arquivo)
   - [ ] Atualizar Claude.md com arquitetura correta
   - [ ] Atualizar Specs.md com padrões do connector-dict
   - [ ] Criar diagramas de arquitetura (Mermaid)

### Fase 1: Implementação Database Layer (Semana 1)

Seguir padrões observados em `instrucoes-app-dict.md` e `instrucoes-orchestration-worker.md`:

1. **Migrations** (`apps/dict/infrastructure/database/migrations/`)
2. **Repositories** (`apps/dict/infrastructure/database/repositories/`)
3. **CID Generator** (`apps/dict/domain/sync/`)
4. **VSync Calculator** (`apps/dict/domain/sync/`)

### Fase 2: Pulsar Event Consumers (Semana 1)

Seguir `instrucoes-orchestration-worker.md`:

1. **Handler Pulsar** (`apps/orchestration-worker/handlers/pulsar/sync/`)
2. **Application Use Case** (`apps/orchestration-worker/application/usecases/sync/`)
3. **Temporal Activity** (`apps/orchestration-worker/infrastructure/temporal/activities/sync/`)

### Fase 3: Temporal Workflows (Semana 2)

Seguir `instrucoes-orchestration-worker.md`:

1. **VSync Verification Workflow** (Cron-based)
2. **Reconciliation Workflow** (Event-triggered)
3. **Setup Cron** (`setup/temporal.go`)

---

## ✅ Status Final

**Dúvidas Técnicas**: 10/10 RESPONDIDAS
**Baseado em**: Documentação oficial do connector-dict
**Próximo Passo**: Atualizar Claude.md e Specs.md com arquitetura correta
**Pronto para**: Iniciar Fase 0 (Validação Técnica)

---

**Última Atualização**: 2025-10-28
**Responsável**: Tech Lead
**Status**: ✅ PRONTO PARA VALIDAÇÃO TÉCNICA
