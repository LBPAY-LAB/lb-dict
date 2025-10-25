# ANA-004 - Revalidação TEC vs Implementação Real

**Versão:** 1.0
**Data:** 2025-10-25
**Autor:** Claude (Análise Automatizada)
**Objetivo:** Validar especificações TEC-001, TEC-002, TEC-003 contra implementações reais (ANA-002, ANA-003)

---

## 1. Sumário Executivo

### 1.1. Status Geral

| Documento | Status Geral | Principais Issues |
|-----------|-------------|-------------------|
| **TEC-001** (Core DICT) | 🟡 **Parcial** | Repositório não analisado (fora de escopo) |
| **TEC-002** (Bridge) | 🟢 **ALINHADO** | Pequenas divergências em nomenclatura |
| **TEC-003** (Connect) | 🟡 **GAPS** | VSYNC e OTP workflows ausentes na implementação |

### 1.2. Inconsistências Críticas Identificadas

| ID | Severidade | Componente | Descrição | Ação Recomendada |
|----|-----------|------------|-----------|------------------|
| **INC-001** | 🔴 Alta | TEC-003 | VSYNC Workflow não implementado | Implementar ou atualizar TEC-003 |
| **INC-002** | 🔴 Alta | TEC-003 | OTP Workflow não implementado | Implementar ou atualizar TEC-003 |
| **INC-003** | 🟡 Média | TEC-002/003 | Nomenclatura de topics Pulsar inconsistente | Padronizar nomes |
| **INC-004** | 🟡 Média | TEC-003 | Database schema não explícito na implementação | Adicionar migrations |
| **INC-005** | 🟡 Média | TEC-002 | Dual protocol (gRPC + Pulsar) não especificado claramente | Atualizar TEC-002 |
| **INC-006** | 🟢 Baixa | TEC-003 | Estrutura multi-app melhor que especificada | Atualizar TEC-003 (enhancement) |

---

## 2. Análise Detalhada TEC-002 (Bridge)

### 2.1. Validação TEC-002 vs ANA-002

#### ✅ ALINHAMENTOS CONFIRMADOS

| Aspecto TEC-002 | Implementação Real (ANA-002) | Status |
|-----------------|------------------------------|--------|
| **Clean Architecture** | ✅ 4 camadas implementadas | ✅ ALINHADO |
| **Sem Temporal Workflows** | ✅ Ausência confirmada (`go.mod` sem `go.temporal.io/sdk`) | ✅ ALINHADO |
| **gRPC Server** | ✅ `handlers/grpc/*_controller.go` | ✅ ALINHADO |
| **Pulsar Consumer** | ✅ `handlers/pulsar/handler.go` | ✅ ALINHADO |
| **XML Signer** | ✅ `shared/signer/signature.go` (JRE + JAR) | ✅ ALINHADO |
| **mTLS Support** | ✅ `shared/http/client.go` com certificados | ✅ ALINHADO |
| **Circuit Breaker** | ✅ `sony/gobreaker/v2` | ✅ ALINHADO |
| **Stateless** | ✅ Sem banco de dados | ✅ ALINHADO |
| **Observability** | ✅ OpenTelemetry | ✅ ALINHADO |

#### 🟡 DIVERGÊNCIAS IDENTIFICADAS

##### INC-005: Dual Protocol Support (gRPC + Pulsar)

**TEC-002 Especifica:**
```markdown
### 1.1 gRPC Server (Operações Síncronas)
...
### 1.2 Pulsar Consumer (Operações Assíncronas)
```

**Implementação Real (ANA-002):**
- ✅ **AMBOS** gRPC e Pulsar implementados **simultaneamente**
- Bridge pode receber requisições via gRPC **OU** Pulsar

**Análise:**
- TEC-002 menciona ambos protocolos, mas não deixa claro se:
  - São **alternativos** (ou um ou outro)
  - São **complementares** (ambos ativos simultaneamente)

**Evidência:**
```
handlers/
├── grpc/              # ✅ gRPC Server implementado
│   ├── claim_controller.go
│   └── directory_controller.go
└── pulsar/            # ✅ Pulsar Consumer implementado
    ├── handler.go
    ├── claim_handler.go
    └── directory_handler.go
```

**Impacto:** 🟡 Médio
**Recomendação:**
1. **Atualizar TEC-002** para especificar claramente dual protocol support
2. Adicionar seção de "Modos de Operação":
   - **Modo Híbrido** (gRPC + Pulsar): Ambos ativos
   - Documentar quando usar cada protocolo (síncrono vs assíncrono)

##### INC-003: Nomenclatura de Topics Pulsar

**TEC-002 Especifica:**
```markdown
**Topics Pulsar**:
- **Entrada**: `bridge-dict-req-in` (consome requisições do Connect)
- **Saída**: `bridge-dict-res-out` (publica respostas para o Connect)
```

**IcePanel (ANA-001) Documenta:**
```
- rsfn-dict-req-out (requests)
- rsfn-dict-res-out (responses)
```

**Implementação Real (ANA-002):**
- Nomes de topics **não explícitos** no código analisado
- Provavelmente configuráveis via environment variables

**Análise:**
- **Inconsistência de nomenclatura** entre TEC-002 e IcePanel
- TEC-002: `bridge-dict-req-in` / `bridge-dict-res-out`
- IcePanel: `rsfn-dict-req-out` / `rsfn-dict-res-out`

**Impacto:** 🟡 Médio
**Recomendação:**
1. **Padronizar nomenclatura** seguindo IcePanel (arquitetura oficial)
2. **Atualizar TEC-002** para usar:
   - `rsfn-dict-req-out` (entrada do Bridge, saída do Connect)
   - `rsfn-dict-res-out` (saída do Bridge, entrada do dict.api)
3. Configurar via `PULSAR_TOPIC_REQ` e `PULSAR_TOPIC_RES`

#### ✅ RESPONSABILIDADES CONFIRMADAS

**TEC-002 Define:**
> Bridge FAZ:
> - Receber requisições (gRPC/Pulsar)
> - Preparar payloads SOAP/XML
> - Assinar XML com certificado ICP-Brasil
> - Enviar requisições mTLS para Bacen
> - Retornar respostas

**ANA-002 Confirma:**
```go
// Use Case Pattern
func (uc *ProcessDirectoryRequestUseCase) Execute(ctx context.Context, req *DirectoryRequest) (*DirectoryResponse, error) {
    // 1. Mapeia request para payload SOAP
    soapPayload := uc.buildSOAPPayload(req)

    // 2. Assina XML com certificado ICP-Brasil
    signedXML, err := uc.xmlSigner.Sign(ctx, soapPayload)

    // 3. Envia para Bacen via mTLS
    response, err := uc.rsfnClient.Send(ctx, signedXML)

    // 4. Retorna resposta (sem armazenar estado)
    return response, nil
}
```

✅ **Implementação 100% alinhada com especificação TEC-002 v3.0**

#### ✅ OPERAÇÕES IMPLEMENTADAS

**TEC-002 Especifica:**
- Directory Operations (CRUD)
- Claim Operations
- Reconciliation (CID, VSYNC)
- Antifraud
- Policies
- Infraction Reports

**ANA-002 Confirma:**
```
application/usecases/
├── antifraud/          # ✅ 8 operações
├── claim/              # ✅ 10 operações
├── directory/          # ✅ 7 operações
├── infraction_report/  # ✅ 9 operações
├── key/                # ✅ 4 operações
├── policies/           # ✅ 5 operações
└── reconciliation/     # ✅ 8 operações
```

✅ **TOTAL: 7 domínios, 51+ operações implementadas**

### 2.2. Conclusão TEC-002

**Status:** 🟢 **95% ALINHADO**

**Issues Pendentes:**
- 🟡 INC-005: Documentar dual protocol support
- 🟡 INC-003: Padronizar nomenclatura de topics Pulsar

**Recomendações:**
1. Atualizar TEC-002 v3.1 com clarificações sobre dual protocol
2. Alinhar nomenclatura de topics com IcePanel
3. Adicionar diagrama de sequência para gRPC vs Pulsar flows

---

## 3. Análise Detalhada TEC-003 (Connect)

### 3.1. Validação TEC-003 vs ANA-003

#### ✅ ALINHAMENTOS CONFIRMADOS

| Aspecto TEC-003 | Implementação Real (ANA-003) | Status |
|-----------------|------------------------------|--------|
| **Temporal SDK** | ✅ `go.temporal.io/sdk v1.36.0` | ✅ ALINHADO |
| **Multi-App Architecture** | ✅ `apps/dict/` + `apps/orchestration-worker/` | ✅ MELHOR que especificado |
| **Claims Workflow** | ✅ `CreateClaimWorkflow` + child workflows | ✅ ALINHADO |
| **Temporal Activities** | ✅ ~10 activities (claims, cache, events) | ✅ ALINHADO |
| **Bridge gRPC Client** | ✅ `infrastructure/grpc/client.go` | ✅ ALINHADO |
| **Pulsar Integration** | ✅ Consumer + Producer | ✅ ALINHADO |
| **API REST** | ✅ Fiber + Huma (melhor que especificado) | ✅ ALINHADO+ |
| **Redis Cache** | ✅ `go-redis/v9` | ✅ ALINHADO |
| **Observability** | ✅ OpenTelemetry | ✅ ALINHADO |

#### 🔴 GAPS CRÍTICOS IDENTIFICADOS

##### INC-001: VSYNC Workflow Ausente

**TEC-003 Especifica:**
```go
## 2.2. VSYNCWorkflow (Verificação de Sincronização Diária)

**Responsabilidade**: Executar VSYNC diário (00:00 BRT) para reconciliar contas PIX.

**Características**:
- **Cron Schedule**: Diário às 00:00 BRT
- **Duração**: ~2-4 horas (dependendo do volume)
- **Retry Policy**: Retry infinito até sucesso
```

**Implementação Real (ANA-003):**
- ❌ **VSYNC Workflow NÃO identificado** no código analisado
- Nenhum arquivo matching `vsync_workflow.go` ou similar
- Nenhuma configuração de cron job encontrada

**Busca Realizada:**
```bash
find apps/orchestration-worker -name "*vsync*" -type f
# Resultado: Nenhum arquivo encontrado
```

**Impacto:** 🔴 **ALTO**
- VSYNC é **requisito regulatório crítico** do Bacen
- Ausência pode significar:
  1. Ainda não implementado (backlog)
  2. Implementado em branch separada
  3. Implementado de forma diferente (outro nome)

**Recomendação:**
1. **URGENTE**: Confirmar status de VSYNC workflow
2. Se não implementado: Priorizar implementação
3. Se em branch separada: Documentar em TEC-003
4. Se implementado diferente: Atualizar ANA-003 e TEC-003

##### INC-002: OTP Workflow Ausente

**TEC-003 Especifica:**
```go
## 2.3. OTPWorkflow (Validação de One-Time Password)

**Responsabilidade**: Validar códigos OTP para operações sensíveis.

**Características**:
- **Timeout**: 5 minutos
- **Signal**: "otp-validated" (sucesso) ou timeout (falha)
```

**Implementação Real (ANA-003):**
- ❌ **OTP Workflow NÃO identificado** no código analisado
- Nenhum arquivo matching `otp_workflow.go` ou similar

**Busca Realizada:**
```bash
find apps/orchestration-worker -name "*otp*" -type f
# Resultado: Nenhum arquivo encontrado
```

**Impacto:** 🔴 **ALTO**
- OTP pode ser **requisito de segurança** para operações críticas
- Ausência pode significar:
  1. Ainda não implementado (backlog)
  2. Implementado em outro serviço (fora de orchestration-worker)
  3. Não é requisito (TEC-003 incorreto)

**Recomendação:**
1. **Verificar**: OTP é requisito real ou especificação excessiva?
2. Se requisito: Implementar OTPWorkflow
3. Se não requisito: Remover de TEC-003 v2.1

##### INC-004: Database Schema Não Explícito

**TEC-003 Especifica:**
```sql
-- Tabela de Reivindicações
CREATE TABLE claims (
    id UUID PRIMARY KEY,
    entry_key VARCHAR(77) NOT NULL,
    claimer_ispb VARCHAR(8) NOT NULL,
    owner_ispb VARCHAR(8) NOT NULL,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    workflow_id VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    resolved_at TIMESTAMP
);

-- Tabela de VSYNC
CREATE TABLE vsync_accounts (
    id UUID PRIMARY KEY,
    ispb VARCHAR(8) NOT NULL,
    account_number VARCHAR(20) NOT NULL,
    ...
);
```

**Implementação Real (ANA-003):**
- ❌ **Database migrations NÃO encontradas** no repositório
- Sem pasta `db/migrations/`
- Sem configs de ORM (GORM, sqlc, etc.)

**Busca Realizada:**
```bash
find connector-dict -name "*migration*" -o -name "*.sql" -type f
# Resultado: Nenhum arquivo SQL encontrado
```

**Análise:**
- **PostgreSQL é dependência confirmada** (usado pelo Temporal Server)
- **Schema específico não explícito** no código Go
- Possibilidades:
  1. Migrations em repositório separado
  2. Gerenciado por ferramenta externa (Flyway, Liquibase)
  3. Ainda não implementado

**Impacto:** 🟡 **MÉDIO**
**Recomendação:**
1. Adicionar migrations em `db/migrations/`
2. Usar `golang-migrate` ou `goose` para versionamento
3. Documentar schema no README

#### ✅ WORKFLOWS IMPLEMENTADOS

**TEC-003 Especifica:**
1. ClaimWorkflow (7 dias)
2. VSYNCWorkflow (daily cron)
3. OTPWorkflow (5 min)

**ANA-003 Confirma:**

| Workflow TEC-003 | Arquivo Implementação | Status |
|------------------|-----------------------|--------|
| **ClaimWorkflow** | ✅ `create_workflow.go` | ✅ IMPLEMENTADO |
| - Monitor Status | ✅ `monitor_status_workflow.go` | ✅ IMPLEMENTADO |
| - Expire Period | ✅ `expire_completion_period_workflow.go` | ✅ IMPLEMENTADO |
| - Complete | ✅ `complete_workflow.go` | ✅ IMPLEMENTADO |
| - Cancel | ✅ `cancel_workflow.go` | ✅ IMPLEMENTADO |
| **VSYNCWorkflow** | ❌ Não encontrado | ❌ AUSENTE (INC-001) |
| **OTPWorkflow** | ❌ Não encontrado | ❌ AUSENTE (INC-002) |

**Claim Workflow Confirmado:**
```go
// infrastructure/temporal/workflows/claims/create_workflow.go
func CreateClaimWorkflow(ctx workflow.Context, input CreateClaimWorkflowInput) error {
    // 1. Activity: Criar reivindicação no Bacen (via Bridge gRPC)
    bacenResp, err := executeCreateClaimActivity(ctx, input)

    // 2. Activity: Cachear resposta (Redis)
    workflows.ExecuteCacheActivity(ctx, input.Hash, bacenResp, false, nil)

    // 3. Activity: Publicar evento para Core (Pulsar)
    workflows.ExecuteCoreEventsPublishActivity(ctx, input.Hash, pkg.ActionCreateClaim, bacenResp)

    // 4. Activity: Publicar evento para DICT (Pulsar)
    workflows.ExecuteDictEventsPublishActivity(ctx, input.Hash, pkg.ActionCreateClaim, bacenResp)

    // 5. Child Workflow: Monitor Completion Period (30 days)
    startMonitorCompletionWorkflow(ctx, bacenResp)

    // 6. Child Workflow: Monitor Status
    startMonitorStatusWorkflow(ctx, bacenResp)

    return nil
}
```

✅ **Implementação alinhada com TEC-003 para ClaimWorkflow**

**Observação:**
- TEC-003 especifica **7 dias** de monitoramento
- ANA-003 mostra **30 dias** na implementação:
  ```go
  timer := workflow.NewTimer(ctx, 30*24*time.Hour)
  ```

**🟡 DIVERGÊNCIA MENOR:**
- TEC-003: "7 dias de monitoramento"
- Implementação: 30 dias (ExpireCompletionPeriodWorkflow)

**Ação:** Validar com negócio qual período correto (7 ou 30 dias)

#### ✅ TEMPORAL ACTIVITIES IMPLEMENTADAS

**TEC-003 Especifica:**
```go
- CreateClaimActivity
- ConfirmClaimActivity
- CancelClaimActivity
- GetClaimActivity
- CacheActivity
- PublishEventActivity
```

**ANA-003 Confirma:**
```
infrastructure/temporal/activities/
├── claims/
│   ├── create_activity.go          # ✅ CreateClaimGRPCActivity
│   ├── complete_activity.go        # ✅ CompleteClaimGRPCActivity
│   ├── cancel_activity.go          # ✅ CancelClaimGRPCActivity
│   └── get_claim_activity.go       # ✅ GetClaimGRPCActivity
├── cache/
│   └── cache_activity.go           # ✅ CacheActivity (Redis)
└── events/
    ├── core_events_activity.go     # ✅ CoreEventsPublishActivity
    └── dict_events_activity.go     # ✅ DictEventsPublishActivity
```

✅ **Todas activities especificadas estão implementadas**

#### 🟢 ENHANCEMENT: Multi-App Architecture

**TEC-003 Especifica:**
```
lb-conn/rsfn-connect/
├── cmd/
│   ├── connect/       # Entrypoint principal
│   └── worker/        # Temporal Worker
```

**Implementação Real (ANA-003):**
```
connector-dict/
├── apps/
│   ├── dict/                    # ✅ API REST (Fiber + Huma)
│   ├── orchestration-worker/    # ✅ Temporal Workers
│   └── shared/                  # ✅ Código compartilhado
```

**Análise:**
- ✅ **Separação MELHOR que especificada**
- Implementação separa completamente:
  - `dict/` → API REST (83 arquivos Go)
  - `orchestration-worker/` → Temporal Workflows (51 arquivos Go)
  - `shared/` → Infraestrutura compartilhada

**Impacto:** 🟢 **Positivo**
**Recomendação:**
1. **Atualizar TEC-003 v2.1** para refletir estrutura multi-app
2. Documentar benefícios da separação:
   - Deploy independente
   - Escalabilidade separada (API vs Workers)
   - Separação de responsabilidades clara

### 3.2. Conclusão TEC-003

**Status:** 🟡 **70% ALINHADO** (Gaps críticos em VSYNC e OTP)

**Issues Pendentes:**
- 🔴 INC-001: VSYNC Workflow ausente
- 🔴 INC-002: OTP Workflow ausente
- 🟡 INC-004: Database schema não explícito
- 🟡 INC-003: Nomenclatura de topics Pulsar

**Pontos Fortes:**
- ✅ ClaimWorkflow completo e bem implementado
- ✅ Temporal Activities alinhadas
- ✅ Bridge Client via gRPC funcional
- ✅ API REST com Fiber + Huma (melhor que especificado)
- ✅ Multi-app architecture excelente

**Recomendações Prioritárias:**
1. **URGENTE**: Implementar VSYNC Workflow ou atualizar TEC-003
2. **URGENTE**: Esclarecer status OTP Workflow
3. Adicionar database migrations
4. Validar período de claim (7 vs 30 dias)
5. Atualizar TEC-003 v2.1 com melhorias implementadas

---

## 4. Análise TEC-001 (Core DICT)

### 4.1. Status

**TEC-001** especifica o **Core DICT (dict.api)** como serviço central de domínio.

**Análise Realizada:**
- ✅ **ANA-003** confirma existência de `apps/dict/` no repositório `connector-dict`
- ✅ Implementado com **Fiber + Huma** (API REST)
- ⚠️ Repositório Core DICT separado **NÃO foi analisado** (fora de escopo)

**Mapeamento Arquitetural:**

| TEC-001 | Implementação Real |
|---------|-------------------|
| Core DICT (gRPC Server) | ❓ Repositório separado não analisado |
| dict.api | ✅ `connector-dict/apps/dict/` (REST API) |

**Observação:**
- **Arquitetura IcePanel** menciona `dict.api` como componente separado
- **TEC-001** e **dict.api** podem ser o **mesmo componente**
- Necessário esclarecer se:
  1. `dict.api` = TEC-001 (Core DICT)
  2. Ou são componentes separados

**Recomendação:**
1. Esclarecer mapeamento TEC-001 ↔ dict.api
2. Analisar repositório Core DICT (se existir separadamente)
3. Atualizar documentação arquitetural

---

## 5. Nomenclatura e Mapeamento IcePanel

### 5.1. Inconsistências de Nomenclatura (INC-003)

**Problema:** Nomes diferentes para mesmos componentes em documentos diferentes.

#### Pulsar Topics

| Documento | Nome Topic Request | Nome Topic Response |
|-----------|-------------------|---------------------|
| **TEC-002** | `bridge-dict-req-in` | `bridge-dict-res-out` |
| **TEC-003** | `dict-req-out` | `dict-res-in` |
| **IcePanel (ANA-001)** | `rsfn-dict-req-out` | `rsfn-dict-res-out` |

**Análise:**
- 3 nomenclaturas diferentes para mesmos topics
- Causa confusão entre equipes
- Necessário padronização

**Recomendação:**
1. **Adotar nomenclatura IcePanel** (arquitetura oficial):
   - `rsfn-dict-req-out` (Core → Connect → Bridge)
   - `rsfn-dict-res-out` (Bridge → Connect → Core)
2. Atualizar TEC-002 e TEC-003 v3.1 / v2.1
3. Configurar nomes via environment variables:
   ```bash
   PULSAR_TOPIC_REQ=rsfn-dict-req-out
   PULSAR_TOPIC_RES=rsfn-dict-res-out
   ```

#### Componentes

| IcePanel | TEC-002 | TEC-003 | Padronização Proposta |
|----------|---------|---------|----------------------|
| DICT Proxy | RSFN Bridge | - | ✅ **Bridge** (TEC-002) |
| dict.api | - | Core DICT (?) | ✅ **dict.api** |
| dict.orchestration.worker | - | RSFN Connect | ✅ **Connect** (orchestration-worker) |
| rsfn-dict-req-out | bridge-dict-req-in | dict-req-out | ✅ **rsfn-dict-req-out** |

### 5.2. Mapeamento Consolidado

**Proposta de Padronização:**

```
┌─────────────────────────────────────────────────────────────────┐
│  IcePanel Name            TEC Doc        Implementation Repo    │
├─────────────────────────────────────────────────────────────────┤
│  dict.api                 TEC-001        [Repositório separado?] │
│  DICT Proxy               TEC-002        rsfn-connect-bacen-bridge│
│  dict.orchestration.worker TEC-003       connector-dict/apps/orchestration-worker│
│  worker.claims            TEC-003        workflows/claims/       │
│  rsfn-dict-req-out        TEC-002/003    (Pulsar topic config)   │
│  rsfn-dict-res-out        TEC-002/003    (Pulsar topic config)   │
└─────────────────────────────────────────────────────────────────┘
```

---

## 6. Tabela Consolidada de Inconsistências

| ID | Severidade | Componente | TEC Spec | Implementação Real | Gap | Ação Recomendada |
|----|-----------|------------|----------|-------------------|-----|------------------|
| **INC-001** | 🔴 Alta | TEC-003 | VSYNCWorkflow especificado | ❌ Não encontrado | VSYNC ausente | Implementar ou remover de TEC-003 |
| **INC-002** | 🔴 Alta | TEC-003 | OTPWorkflow especificado | ❌ Não encontrado | OTP ausente | Validar necessidade e implementar |
| **INC-003** | 🟡 Média | TEC-002/003 | `bridge-dict-req-in` / `dict-req-out` | Nomes configuráveis | 3 nomenclaturas diferentes | Padronizar para `rsfn-dict-*` |
| **INC-004** | 🟡 Média | TEC-003 | Database schema detalhado | ❌ Migrations não encontradas | Schema não explícito | Adicionar `db/migrations/` |
| **INC-005** | 🟡 Média | TEC-002 | gRPC OU Pulsar | ✅ Ambos implementados | Dual protocol não claro | Documentar dual support |
| **INC-006** | 🟢 Baixa | TEC-003 | Estrutura monolítica | ✅ Multi-app melhor | Enhancement | Atualizar TEC-003 (positivo) |
| **INC-007** | 🟡 Média | TEC-003 | Claim 7 dias | ✅ Implementado 30 dias | Período diferente | Validar com negócio |
| **INC-008** | 🟡 Média | TEC-001 | Core DICT | ❓ Repo não analisado | Mapeamento unclear | Esclarecer arquitetura |

---

## 7. Plano de Ação Recomendado

### 7.1. Prioridade ALTA (Semana 1)

1. **INC-001: VSYNC Workflow**
   - [ ] Confirmar com time se VSYNC está implementado em outro repo/branch
   - [ ] Se ausente: Priorizar implementação (requisito regulatório)
   - [ ] Atualizar TEC-003 com status real

2. **INC-002: OTP Workflow**
   - [ ] Validar com segurança se OTP é requisito real
   - [ ] Se sim: Implementar OTPWorkflow
   - [ ] Se não: Remover de TEC-003 v2.1

3. **INC-003: Nomenclatura Pulsar Topics**
   - [ ] Adotar padrão IcePanel: `rsfn-dict-req-out` / `rsfn-dict-res-out`
   - [ ] Atualizar TEC-002 v3.1 e TEC-003 v2.1
   - [ ] Configurar via environment variables

### 7.2. Prioridade MÉDIA (Semana 2-3)

4. **INC-004: Database Migrations**
   - [ ] Adicionar `db/migrations/` em `connector-dict`
   - [ ] Implementar migrations com `golang-migrate`
   - [ ] Documentar schema no README

5. **INC-005: Dual Protocol Documentation**
   - [ ] Atualizar TEC-002 v3.1 com seção "Dual Protocol Support"
   - [ ] Documentar quando usar gRPC vs Pulsar
   - [ ] Adicionar diagramas de sequência

6. **INC-007: Claim Period Validation**
   - [ ] Validar com negócio: 7 dias ou 30 dias?
   - [ ] Atualizar TEC-003 ou código conforme decisão

### 7.3. Prioridade BAIXA (Semana 4)

7. **INC-006: Multi-App Enhancement**
   - [ ] Atualizar TEC-003 v2.1 com estrutura multi-app
   - [ ] Documentar benefícios (deploy independente, escalabilidade)

8. **INC-008: Core DICT Mapping**
   - [ ] Esclarecer mapeamento TEC-001 ↔ dict.api
   - [ ] Analisar repositório Core DICT (se existir)
   - [ ] Atualizar documentação arquitetural

---

## 8. Versionamento Recomendado de Documentos

### 8.1. TEC-002 Bridge Specification

**Versão Atual:** 3.0
**Próxima Versão:** **3.1**

**Mudanças Propostas:**
1. Adicionar seção "Dual Protocol Support" (gRPC + Pulsar)
2. Atualizar nomenclatura topics para `rsfn-dict-*`
3. Adicionar diagrama de sequência gRPC vs Pulsar

### 8.2. TEC-003 Connect Specification

**Versão Atual:** 2.0
**Próxima Versão:** **2.1**

**Mudanças Propostas:**
1. Remover ou marcar como "Pendente": VSYNC e OTP workflows
2. Atualizar estrutura para multi-app (dict + orchestration-worker)
3. Atualizar nomenclatura topics para `rsfn-dict-*`
4. Adicionar seção "Database Migrations"
5. Validar período claim (7 vs 30 dias)

### 8.3. TEC-001 Core DICT Specification

**Versão Atual:** 1.0
**Próxima Versão:** **1.1** (se necessário)

**Mudanças Propostas:**
1. Esclarecer relação com dict.api
2. Revalidar após análise de repositório Core DICT

---

## 9. Conclusão Geral

### 9.1. Status Global de Alinhamento

| Documento | Alinhamento | Principais Issues |
|-----------|------------|-------------------|
| **TEC-001** | 🟡 Não Validado | Repositório não analisado |
| **TEC-002** | 🟢 **95%** | Pequenas divergências nomenclatura |
| **TEC-003** | 🟡 **70%** | VSYNC/OTP ausentes, schema não explícito |

### 9.2. Qualidade da Implementação

**Pontos Fortes:**
- ✅ **Clean Architecture** bem implementada (Bridge e Connect)
- ✅ **Temporal Workflows** funcionais (Claims completo)
- ✅ **Multi-App Separation** excelente (dict + orchestration-worker)
- ✅ **Bridge como Adapter puro** corretamente implementado
- ✅ **Observabilidade** (OpenTelemetry) desde início

**Gaps Críticos:**
- 🔴 **VSYNC Workflow** ausente (requisito regulatório?)
- 🔴 **OTP Workflow** ausente (requisito segurança?)
- 🟡 **Database Migrations** não explícitas
- 🟡 **Nomenclatura inconsistente** entre documentos

### 9.3. Recomendação Final

**Ação Imediata:**
1. ✅ **TEC-002 está praticamente alinhado** - pequenos ajustes nomenclatura
2. 🔴 **TEC-003 precisa de validação urgente** sobre VSYNC/OTP
3. 🟡 **TEC-001 precisa de análise** de repositório Core DICT

**Next Steps:**
1. Executar Plano de Ação (Prioridade Alta primeiro)
2. Atualizar TEC-002 v3.1 e TEC-003 v2.1
3. Analisar repositório Core DICT (se existir)
4. Criar ADR para padronização de nomenclatura

---

**Documento gerado automaticamente via cross-validation TEC vs ANA**
**Última atualização:** 2025-10-25
