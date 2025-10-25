# ARE-003: Análise Documento de Arquitetura DICT

**Documento**: Análise em profundidade do documento de arquitetura DICT (ArquiteturaDict_LBPAY.md)
**Data**: 2025-10-24
**Versão**: 1.0
**Status**: COMPLETO
**Autor**: NEXUS (AGT-SA-001)

---

## RESUMO EXECUTIVO

### Principais Descobertas

O documento de arquitetura DICT (`ArquiteturaDict_LBPAY.md`, 1.1MB, 1760 linhas) é um diagrama estrutural completo baseado no **IcePanel**, contendo a arquitetura AS-IS (atual) do sistema DICT LBPay. Principais achados:

#### 1. **RSFN Connect: O "Bridge" Atual**
- **NÃO existe menção a "Bridge" no documento**
- O componente equivalente é o **RSFN Connect** (Rede do Sistema Financeiro Nacional)
- É um **sistema específico para DICT**, não genérico
- Função: integração direta com Banco Central (BC DICT)

#### 2. **Resposta para DUV-003 (Nível de Abstração do Bridge)**
O documento evidencia que a arquitetura atual utiliza um **Bridge ESPECÍFICO DICT**:
- **RSFN Connect** é dedicado exclusivamente ao DICT
- Possui componentes especializados (DICT Proxy, Producer/Consumer RSFN)
- Não há abstração genérica para outros sistemas do Bacen
- **Recomendação**: Manter Bridge específico DICT no core-dict

#### 3. **Resposta para DUV-012 (Performance para Alto Volume)**
Estratégias de performance identificadas:
- **5 caches Redis dedicados**: Respostas, Contas, Validação Chave, Dedup, Rate Limit
- **Apache Pulsar** para mensageria assíncrona (6+ filas/topics)
- **Temporal Workflows** para orquestração (workers especializados)
- **Rate Limiting** com token bucket (Worker Rate Limit)
- **PostgreSQL** para persistência de longa duração (CID, VSync, Statistics)

### Impacto nas Dúvidas Arquiteturais

| Dúvida | Status | Resposta do Documento |
|--------|--------|------------------------|
| **DUV-003**: Nível de abstração do Bridge | ✅ RESOLVIDA | Bridge específico DICT (RSFN Connect) |
| **DUV-012**: Performance para alto volume | ✅ RESOLVIDA | Caching Redis + Pulsar + Temporal + Rate Limiting |

---

## 1. ESTRATÉGIA DE PERSISTÊNCIA

### 1.1. Banco de Dados

**Decisão identificada**: Uso de **PostgreSQL** para persistência estruturada

#### Stores PostgreSQL (2):

```
1. CID e VSync (9.6)
   - Descrição: Base de dados principal para DICT
   - Uso: Diretório de Identificadores e Verificação de Sincronismo
   - Status: live

2. db_statistics (9.7)
   - Descrição: informações de segurança - PersonStatistics e EntryStatistics
   - Uso: Persistência de métricas e indicadores de fraude
   - Status: live
   - Integra com: dict.dashboard, dict.statistics
```

**Observação**: Não há menção a schemas ou tabelas específicas. O documento foca em componentes e fluxos.

### 1.2. Estratégia de Cache

**Decisão identificada**: Cache distribuído com **Redis** (5 instâncias dedicadas)

#### Caches Redis Identificados:

```ascii
┌─────────────────────────────────────────────────────────────┐
│                  ESTRATÉGIA DE CACHE DICT                    │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  1. Cache Contas (9.1) - live                                │
│     - Dados sincronizados das contas                         │
│     - Usado por: dict.api, Sync de Contas                    │
│                                                               │
│  2. Cache de Resposta (9.2) - live                           │
│     - Respostas de requisições (por hash)                    │
│     - Usado por: dict.api (consulta + persistência)          │
│     - Crítico para performance                               │
│                                                               │
│  3. Cache Dedup (9.3) - live                                 │
│     - cache claims_seen (deduplicação)                       │
│     - Usado por: dict.orchestration.monitor                  │
│                                                               │
│  4. Cache Rate Limit (9.4) - future                          │
│     - Controle de rate limiting (token bucket)               │
│     - Usado por: DICT Proxy, Worker Rate Limit               │
│                                                               │
│  5. Cache Validacao Chave (9.5) - live                       │
│     - Validação de chaves PIX                                │
│     - Usado por: dict.api (validações)                       │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

**Padrão de Uso**: Consulta antes de processar, persistência após processar

### 1.3. Persistência de Longa Duração

**Sistema Identificado**: **Audit** (System 7.1)

```yaml
Audit System:
  Descrição: Mantém trilha de auditoria, gerencia persistência longa
  Grupo: Data Lake
  Integração: Consome eventos de domain_events via Pulsar
  Status: live
```

**Conclusão**: Separação clara entre dados operacionais (PostgreSQL) e auditoria (Data Lake).

---

## 2. ARQUITETURA DO "BRIDGE" (RSFN CONNECT)

### 2.1. Descoberta Crítica

**O documento NÃO menciona "Bridge"**. O equivalente arquitetural é:

```
RSFN Connect (System 7.8)
└── Descrição: Rede do Sistema Financeiro Nacional
└── Grupo: LB-Connect
└── Status: live
└── Função: Sistema de integração com Banco Central
```

### 2.2. Componentes do RSFN Connect

```ascii
┌──────────────────────────────────────────────────────────────────┐
│                    RSFN CONNECT ARCHITECTURE                      │
├──────────────────────────────────────────────────────────────────┤
│                                                                    │
│  ┌────────────────┐      ┌──────────────────────┐                │
│  │   DICT Proxy   │◄────►│ DICT Proxy App       │                │
│  │  (Component)   │      │ (Application Layer)  │                │
│  └────────┬───────┘      └──────────────────────┘                │
│           │                                                        │
│           ├─► Consumer DICT Proxy (10.2)                          │
│           ├─► Producer DICT Proxy (10.17)                         │
│           ├─► Producer Rate Limit (10.18)                         │
│           └─► Validação Rate Limit (10.21)                        │
│                                                                    │
│  ┌─────────────────────────────────────────────┐                 │
│  │          Topics Pulsar (RSFN)                │                 │
│  ├─────────────────────────────────────────────┤                 │
│  │  • rsfn-dict-req-out (9.12) - requisições   │                 │
│  │  • rsfn-dict-res-out (9.13) - respostas     │                 │
│  └─────────────────────────────────────────────┘                 │
│                                                                    │
│  Integração com:                                                  │
│  - BC DICT (External System)                                      │
│  - DICT (Internal System)                                         │
│  - Cipher (Certificados/mTLS)                                     │
│  - Open Telemetry Collector (Observabilidade)                    │
│                                                                    │
└──────────────────────────────────────────────────────────────────┘
```

### 2.3. Nível de Abstração

**Resposta para DUV-003**: O "Bridge" atual é **ESPECÍFICO DICT**

Evidências:
1. Nome do sistema: **RSFN Connect** (não genérico)
2. Componentes especializados: `DICT Proxy`, `dict.rsfn.producer`, `dict.rsfn.consumer`
3. Topics Pulsar específicos: `rsfn-dict-req-out`, `rsfn-dict-res-out`
4. Integração direta com `BC DICT` (não abstrata)

**Implicação**: A arquitetura atual NÃO prevê um bridge genérico. Se core-dict criar um Bridge genérico, será uma **decisão arquitetural nova**, não uma continuação do padrão existente.

### 2.4. Integração com Bacen

**Protocolo**: Não especificado no documento (apenas diagramas de fluxo)

**Componentes de Integração**:
```
1. DICT Proxy (3.1 - Component Diagram)
   └── Faz proxy das requisições DICT para BC

2. Cipher (System 7.3)
   └── Mantém certificados, atualizações automáticas
   └── Implica: uso de mTLS para comunicação com Bacen

3. Fluxo de Comunicação:
   dict.api → dict.rsfn.producer → rsfn-dict-req-out →
   DICT Proxy → BC DICT → DICT Proxy → rsfn-dict-res-out →
   dict.rsfn.consumer → dict.api
```

### 2.5. Padrões de Extensibilidade

**NÃO identificados** no documento. O RSFN Connect é implementação concreta, não framework extensível.

---

## 3. PERFORMANCE E ESCALA

### 3.1. SLAs Definidos

**Parcialmente identificado**: Apenas para **Statistics**

```
dict.statistics (8.8):
  Descrição: "O DICT coleta e fornece informações de segurança"
  SLA de Atualização: Máximo de 12 horas
  Campo watermark: Indica data do último evento
  Fonte: Seção 8.8 (linhas ~812-886)
```

**Gap**: Não há SLAs de latência ou throughput para operações CRUD.

### 3.2. Estratégias de Cache

**Resposta para DUV-012**: Múltiplas camadas de caching

#### Padrão de Cache na Requisição CRUD (Seção 4.1)

```
Fluxo de Requisição:
1. dict.api.controller recebe requisição
2. Autenticação e Autorização (mTLS)
3. dict.api.model.validation: valida payload (estrutura)
4. dict.api.application: consulta Cache de Resposta (hash) ◄─ CACHE L1
5. dict.api.validation:
   - Valida chave via Cache Validacao Chave ◄─ CACHE L2
   - Valida conta via Cache Contas ◄─ CACHE L2
6. Processa requisição
7. Persiste resultado no Cache de Resposta ◄─ CACHE L1
```

**Estratégia**: Cache-Aside Pattern (Lazy Loading)

#### Performance Workers

```
Worker Rate Limit (8.16):
  - Repõe tokens conforme programado
  - Consome eventos de Rate Limit
  - Usa: Cache Rate Limit (Redis)
  - Padrão: Token Bucket Algorithm

Worker Respostas Orfãs (10.22):
  - Olha fila rsfn-dict-res-out
  - Consome mensagens (timeout + x ms)
  - Coloca no Cache de Respostas
  - Evita: perda de respostas assíncronas
```

### 3.3. Connection Pooling

**NÃO mencionado** explicitamente no documento.

**Inferência**: Aplicações Golang (dict.api, workers) provavelmente usam connection pooling padrão do Go para:
- Redis (via go-redis)
- PostgreSQL (via pgx ou lib/pq)
- Pulsar (via pulsar-client-go)

### 3.4. Separação Leitura/Escrita

**NÃO identificada** no documento.

**Observação**: Arquitetura atual parece usar mesma base para leitura e escrita (PostgreSQL CID e VSync).

**Gap**: Possível otimização futura (CQRS pattern não evidenciado).

---

## 4. COMUNICAÇÃO ASSÍNCRONA

### 4.1. Temporal Workflows

**Sistema**: Temporal Server (7.9) + Apps workers

```yaml
Temporal Server (System 7.9):
  Descrição: Orquestração com workflows, agendamentos, triggers
  Status: live

Temporal App (8.15):
  Tecnologia: Temporal (https://docs.temporal.io)
  Status: live

Workers Temporal:
  - dict.orchestration.monitor (8.6)
  - dict.orchestration.worker (8.7)
  - dict.vsync (8.9)
```

### 4.2. Workflows Identificados

```ascii
┌──────────────────────────────────────────────────────────────┐
│              TEMPORAL WORKFLOWS ARCHITECTURE                  │
├──────────────────────────────────────────────────────────────┤
│                                                                │
│  1. dict.orchestration.monitor (8.6)                          │
│     ├─ Descrição: Polling no DICT BC em busca de notificações│
│     ├─ Componentes:                                           │
│     │  └─ workflow polling (10.26)                            │
│     ├─ Stores:                                                │
│     │  └─ locks/dict-orchestration-monitor (Pulsar)           │
│     │  └─ Cache Dedup (Redis)                                 │
│     └─ Status: live                                           │
│                                                                │
│  2. dict.orchestration.worker (8.7)                           │
│     ├─ Descrição: Contém workflows Temporal de processamento │
│     ├─ Components (Activities):                               │
│     │  ├─ worker.claims (10.23) - reivindicações             │
│     │  ├─ worker.entries (10.24) - vínculo de chaves         │
│     │  └─ worker respostas orfãs (10.22) - timeout handling  │
│     └─ Status: live                                           │
│                                                                │
│  3. dict.vsync (8.9)                                          │
│     ├─ Descrição: Verificação de sincronismo                 │
│     ├─ Components:                                            │
│     │  ├─ dic.vsync.domainevents.consumer (10.4)             │
│     │  └─ worker.vsync (10.25) - activities de sincronismo   │
│     ├─ Stores:                                                │
│     │  └─ CID e VSync (PostgreSQL)                            │
│     └─ Status: live                                           │
│                                                                │
└──────────────────────────────────────────────────────────────┘
```

### 4.3. Apache Pulsar

**Infraestrutura de Mensageria**: Apache Pulsar (6 filas/topics identificados)

```ascii
┌────────────────────────────────────────────────────────────┐
│               APACHE PULSAR TOPICS/QUEUES                   │
├────────────────────────────────────────────────────────────┤
│                                                              │
│  DOMAIN EVENTS                                              │
│  ├─ nome_da_fila_domain_events (9.9) - live                │
│  │  └─ Eventos de alteração de estado                       │
│  │  └─ Criação, alteração, exclusão                         │
│  │  └─ Consumidores: Audit, Statistics, VSync               │
│  │                                                           │
│  RATE LIMITING                                              │
│  ├─ nome_da_fila_rate_limit (9.10) - live                  │
│  │  └─ Consumidor: Worker Rate Limit                        │
│  │                                                           │
│  SYNC CONTAS                                                │
│  ├─ nome_da_fila_sync_contas (9.11) - live                 │
│  │  └─ Sincronização de dados de contas                     │
│  │                                                           │
│  RSFN COMMUNICATION                                         │
│  ├─ rsfn-dict-req-out (9.12) - live                        │
│  │  └─ Tópico para requisições ao BC DICT                   │
│  │                                                           │
│  ├─ rsfn-dict-res-out (9.13) - live                        │
│  │  └─ Tópico para respostas do BC DICT                     │
│  │                                                           │
│  ORCHESTRATION LOCKS                                        │
│  └─ locks/dict-orchestration-monitor (9.8) - future        │
│     └─ Controle de locks para polling                       │
│                                                              │
└────────────────────────────────────────────────────────────┘
```

### 4.4. Padrões de Mensageria

**Padrões Identificados**:

1. **Event-Driven Architecture (EDA)**
   - Producer: `dict.domainevents.producer` (10.12)
   - Descrição: "toda alteração de estado é publicada"
   - Consumidores: Audit, Statistics, VSync

2. **Request-Response Assíncrono**
   - Requisição: `dict.rsfn.producer` → `rsfn-dict-req-out`
   - Resposta: `rsfn-dict-res-out` → `dict.rsfn.consumer`
   - Timeout handling: Worker respostas orfãs

3. **Queue-Based Load Leveling**
   - Worker Rate Limit consome eventos de rate limit
   - Processa em background, não bloqueia request path

---

## 5. ESTRUTURA DO CORE DICT

### 5.1. Clean Architecture

**Status**: NÃO explicitamente mencionada, mas evidências de separação de camadas

#### Camadas Identificadas (dict.api - 8.3):

```ascii
┌────────────────────────────────────────────────────────┐
│            DICT.API COMPONENT STRUCTURE                 │
├────────────────────────────────────────────────────────┤
│                                                          │
│  CONTROLLER LAYER                                       │
│  ├─ dict.api.controller (10.7)                         │
│  │  └─ "recebe requisições para serem processadas"     │
│  │  └─ Tech: Golang                                     │
│  │                                                       │
│  VALIDATION LAYER                                       │
│  ├─ dict.api.model.validation (10.8)                   │
│  │  └─ "validação estrutural do payload"               │
│  │  └─ Campos obrigatórios, tamanhos, limites          │
│  │  └─ Calcula hash                                     │
│  │                                                       │
│  ├─ dict.api.validation (10.9)                         │
│  │  └─ "realiza validações da chave PIX"               │
│  │  └─ Posse, situação na RF, Nomes                     │
│  │                                                       │
│  APPLICATION LAYER                                      │
│  ├─ dict.api.application (10.6)                        │
│  │  └─ "centraliza as regras de negócio"               │
│  │  └─ "orquestra o processamento das requisições"     │
│  │                                                       │
│  INTEGRATION LAYER                                      │
│  ├─ dict.rsfn.producer (10.14)                         │
│  ├─ dict.rsfn.consumer (10.13)                         │
│  ├─ dict.domainevents.producer (10.12)                 │
│  │                                                       │
│  SECURITY LAYER                                         │
│  └─ Autenticação e Autorização (10.1)                  │
│     └─ "mTLS"                                            │
│                                                          │
└────────────────────────────────────────────────────────┘
```

### 5.2. Mapeamento para Clean Architecture

| Camada Clean Arch | Componentes dict.api |
|-------------------|---------------------|
| **Presentation** | dict.api.controller |
| **Application** | dict.api.application, dict.api.validation |
| **Domain** | *(não explícito - provável dentro de application)* |
| **Infrastructure** | dict.rsfn.*, dict.domainevents.producer, caches |

**Gap**: Documento não detalha estrutura de domínio (entities, value objects, repositories).

### 5.3. Padrões de Código

**Não mencionados** no documento.

**Inferência** (baseado em stack Golang):
- Provavelmente: Dependency Injection, Repository Pattern
- Provável estrutura: `/internal/app/`, `/internal/domain/`, `/pkg/`

---

## 6. INTEGRAÇÕES

### 6.1. Core DICT ↔ Connector

**NÃO existe "Connector" mencionado no documento.**

**Equivalente funcional**: `dict.api` (System 8.3)

```
dict.api (8.3):
  Descrição: "thin API para receber requisições do DICT"
  Tecnologia: Golang + API
  Status: live

  Integração com Core:
    - Recebe requisições de Core (System 7.4)
    - Relação: Core → dict.api (usa)
```

### 6.2. Connector ↔ Bridge

**Traduzindo para nomenclatura do documento**: `dict.api` ↔ `RSFN Connect`

```ascii
┌──────────────────────────────────────────────────────────┐
│        INTEGRAÇÃO DICT.API ↔ RSFN CONNECT                │
├──────────────────────────────────────────────────────────┤
│                                                            │
│  dict.api (8.3)                                           │
│    │                                                       │
│    ├─► dict.rsfn.producer (10.14)                        │
│    │   └─ Publica mensagem em: rsfn-dict-req-out         │
│    │                                                       │
│    │   [ Apache Pulsar - Message Broker ]                 │
│    │                                                       │
│    │   ┌──────────────────────────────────┐              │
│    │   │  RSFN Connect (DICT Proxy)       │              │
│    │   │  ├─ Consumer DICT Proxy (10.2)   │              │
│    │   │  │  └─ Consome: rsfn-dict-req-out│              │
│    │   │  │                                 │              │
│    │   │  ├─ DICT Proxy Application (10.5) │              │
│    │   │  │  └─ Processa requisição         │              │
│    │   │  │  └─ Envia para BC DICT          │              │
│    │   │  │                                 │              │
│    │   │  └─ Producer DICT Proxy (10.17)   │              │
│    │   │     └─ Publica: rsfn-dict-res-out │              │
│    │   └──────────────────────────────────┘              │
│    │                                                       │
│    │   [ Apache Pulsar - Message Broker ]                 │
│    │                                                       │
│    └─► dict.rsfn.consumer (10.13)                        │
│        └─ Consome: rsfn-dict-res-out                      │
│                                                            │
└──────────────────────────────────────────────────────────┘
```

### 6.3. Protocolos de Integração

| Integração | Protocolo | Evidência |
|------------|-----------|-----------|
| **Core → dict.api** | HTTP/REST ou gRPC | Mencionado como "API" (8.3) |
| **dict.api ↔ RSFN Connect** | Apache Pulsar | Topics rsfn-dict-req/res-out |
| **RSFN Connect → BC DICT** | mTLS (provável) | Cipher mantém certificados |
| **Interno (Workers)** | Temporal gRPC | Padrão Temporal |

**Gap**: Documento não especifica se dict.api é REST ou gRPC.

---

## 7. OBSERVABILIDADE

### 7.1. Ferramentas Definidas

**Sistema**: Observabilidade (7.7)

```ascii
┌─────────────────────────────────────────────────────────┐
│             OBSERVABILITY STACK                          │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  METRICS                                                 │
│  ├─ Prometheus (8.12) - live                            │
│  │  └─ Store: store_prometheus                           │
│  │  └─ Usado por: dict.dashboard                         │
│  │                                                        │
│  TRACING & LOGS                                          │
│  ├─ Signoz (8.13) - live                                │
│  │                                                        │
│  COLLECTION                                              │
│  ├─ Open Telemetry Collector (8.11) - live              │
│  │  └─ Descrição: "coletor de telemetria"               │
│  │  └─ "persiste em uma base Prometheus"                │
│  │                                                        │
│  VISUALIZATION                                           │
│  └─ Grafana (8.10) - live                               │
│     └─ Usado por: dict.dashboard                         │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

### 7.2. Logging

**NÃO detalhado** no documento.

**Inferência**: Apps Golang provavelmente usam:
- Structured logging (JSON)
- Integração com OpenTelemetry Collector

### 7.3. Tracing

**Sistema**: Signoz (8.13)

```
Signoz:
  Função: Tracing distribuído + logs
  Integração: OpenTelemetry Collector
  Status: live
```

**Gap**: Documento não mostra como tracing é instrumentado nos apps.

### 7.4. Métricas

**Sistema**: Prometheus (8.12) + Grafana (8.10)

```
Prometheus:
  - Coleta métricas via OpenTelemetry Collector
  - Store: store_prometheus (9.14)

Grafana:
  - Visualização de métricas
  - Integrado com dict.dashboard (3.3)
```

**Métricas específicas**:
- `dict.statistics` (8.8): Persiste métricas de segurança e fraude
- Components: `dict.statistics.collector` (10.15), `dict.statistics.summarizer` (10.16)

---

## 8. RESPOSTAS PARA DÚVIDAS PENDENTES

### 8.1. DUV-003: Qual o nível de abstração do Bridge?

**RESPOSTA BASEADA NO DOCUMENTO**:

O "Bridge" atual (**RSFN Connect**) é **ESPECÍFICO DICT**, com os seguintes fundamentos:

#### Evidências de Especialização:

1. **Nomenclatura Específica**
   - Sistema: "RSFN Connect" (Rede do Sistema Financeiro Nacional)
   - Componentes: `DICT Proxy`, `dict.rsfn.producer`, `dict.rsfn.consumer`
   - Naming convention indica acoplamento ao domínio DICT

2. **Topics Pulsar Especializados**
   ```
   - rsfn-dict-req-out (requisições DICT)
   - rsfn-dict-res-out (respostas DICT)
   ```
   Não há abstração genérica (ex: `rsfn-req-out`, `rsfn-res-out`)

3. **Integração Direta com BC DICT**
   - External System: "BC DICT" (7.2)
   - Descrição: "Diretório de Identificadores Contas Transacionais"
   - Sem camada de abstração para outros sistemas Bacen

4. **Componentes Não Reutilizáveis**
   - DICT Proxy Application (10.5): lógica específica
   - Producer/Consumer RSFN: não genéricos

#### Implicações para core-dict:

**Opção A: Bridge Genérico (DUV-003-OPT-A)**
- ✅ Permite extensão futura (PIX, CIP, etc.)
- ❌ Não segue padrão da arquitetura atual
- ❌ Requer ADR para justificar mudança arquitetural

**Opção B: Bridge Específico DICT (DUV-003-OPT-B)** ⭐ **RECOMENDADA**
- ✅ Mantém consistência com arquitetura atual
- ✅ Simplicidade: foco no domínio DICT
- ✅ Menos over-engineering
- ❌ Menos flexível para futuro

**Decisão Sugerida**: **Bridge Específico DICT**, alinhado com padrão RSFN Connect.

---

### 8.2. DUV-012: Quais estratégias de performance estão definidas?

**RESPOSTA BASEADA NO DOCUMENTO**:

#### Estratégias Identificadas:

**1. Caching Multi-Camada (Redis)**

```yaml
Cache L1 - Respostas (Cache de Resposta):
  Uso: Evitar reprocessamento de requisições idênticas
  Trigger: Hash do payload
  Padrão: Cache-Aside (Lazy Loading)
  Status: live

Cache L2 - Dados de Referência:
  - Cache Contas: dados sincronizados
  - Cache Validacao Chave: validações PIX
  Status: live

Cache L3 - Controle:
  - Cache Dedup: claims_seen (deduplicação)
  - Cache Rate Limit: token bucket
  Status: live/future
```

**2. Mensageria Assíncrona (Apache Pulsar)**

```yaml
Padrão: Request-Response Assíncrono
Benefícios:
  - Desacoplamento temporal
  - Absorção de picos de carga
  - Retry automático
  - Não bloqueia request path

Topics:
  - rsfn-dict-req-out (requisições)
  - rsfn-dict-res-out (respostas)
```

**3. Orquestração com Temporal**

```yaml
Workflows:
  - Polling assíncrono (dict.orchestration.monitor)
  - Processamento background (dict.orchestration.worker)
  - Sincronização (dict.vsync)

Benefícios:
  - Execução assíncrona longa duração
  - Retry automático com backoff
  - State management distribuído
```

**4. Rate Limiting (Token Bucket)**

```yaml
Worker: Worker Rate Limit (8.16)
Componente: Validação Rate Limit (10.21)
Storage: Cache Rate Limit (Redis)
Padrão: Token Bucket Algorithm
Descrição: "repõe tokens conforme programado"
```

**5. Timeout Handling**

```yaml
Worker: worker respostas orfãs (10.22)
Descrição: "Olha fila rsfn-dict-res-out"
Timeout: "timeout + x ms"
Ação: Coloca no Cache de Respostas
```

#### Performance SLAs:

**Identificado**: Apenas para Statistics
```
dict.statistics:
  - Atualização: Máximo 12 horas
  - Watermark: Indica data do último evento
```

**Gap**: Sem SLAs de latência/throughput para CRUD.

#### Recomendações para core-dict:

1. **Adotar caching em camadas** (L1=Respostas, L2=Referência)
2. **Usar Pulsar** para comunicação assíncrona
3. **Implementar Rate Limiting** desde início
4. **Definir SLAs** de latência (ex: p99 < 500ms)

---

## 9. GAPS IDENTIFICADOS

### 9.1. Gaps de Documentação

| Gap | Categoria | Impacto | Prioridade |
|-----|-----------|---------|------------|
| **Schemas PostgreSQL** | Persistência | Alto | P0 |
| **Definição de API (REST/gRPC)** | Integração | Alto | P0 |
| **SLAs de Latência/Throughput** | Performance | Alto | P0 |
| **Estrutura de Domínio (Entities, VOs)** | Arquitetura | Médio | P1 |
| **Padrões de Código** | Implementação | Médio | P1 |
| **Connection Pooling** | Performance | Médio | P2 |
| **CQRS (Leitura/Escrita)** | Escala | Baixo | P3 |
| **Instrumentação Tracing** | Observabilidade | Médio | P2 |

### 9.2. Gaps de Decisão Arquitetural

1. **Bridge Genérico vs Específico** (DUV-003)
   - Documento mostra: Específico
   - core-dict pode divergir (requer ADR)

2. **Segregação de Responsabilidades**
   - Documento não deixa claro: dict.api é Core ou Connector?
   - Necessário: definir boundary entre core-dict e outros módulos

3. **Padrão de Erros**
   - Como erros são propagados assincronamente?
   - Dead Letter Queues em Pulsar?

### 9.3. Gaps de Implementação

1. **Autenticação/Autorização**
   - Documento menciona mTLS
   - Gap: como Core autentica chamadas ao dict.api?
   - Keycloak (7.6) está no diagrama mas sem fluxo detalhado

2. **Health Checks**
   - Não mencionado no documento
   - Necessário: /health, /ready endpoints

3. **Circuit Breaker**
   - Não mencionado
   - Necessário: proteção contra falhas em BC DICT

---

## 10. ADRs PROPOSTOS

Com base nas decisões identificadas no documento, propõe-se criar os seguintes ADRs:

### 10.1. ADRs de Arquitetura Core

| ID | Título | Decisão Derivada do Documento |
|----|--------|-------------------------------|
| **ADR-001** | Estratégia de Caching Multi-Camada | Redis com 5 caches especializados |
| **ADR-002** | Mensageria com Apache Pulsar | Comunicação assíncrona desacoplada |
| **ADR-003** | Orquestração com Temporal | Workflows para processos longos |
| **ADR-004** | Persistência PostgreSQL | CID, VSync, Statistics |
| **ADR-005** | Separação Auditoria (Data Lake) | Audit system separado |

### 10.2. ADRs de Integração

| ID | Título | Decisão Derivada do Documento |
|----|--------|-------------------------------|
| **ADR-006** | RSFN Connect como Bridge DICT | Bridge específico vs genérico |
| **ADR-007** | Protocolo mTLS com Bacen | Cipher gerencia certificados |
| **ADR-008** | Pattern Request-Response Assíncrono | Topics Pulsar req-out/res-out |

### 10.3. ADRs de Performance

| ID | Título | Decisão Derivada do Documento |
|----|--------|-------------------------------|
| **ADR-009** | Rate Limiting com Token Bucket | Worker Rate Limit + Redis |
| **ADR-010** | Timeout Handling Assíncrono | Worker respostas orfãs |
| **ADR-011** | Cache-Aside Pattern | Consulta antes, persiste depois |

### 10.4. ADRs de Observabilidade

| ID | Título | Decisão Derivada do Documento |
|----|--------|-------------------------------|
| **ADR-012** | Stack Observability (Prometheus/Grafana/Signoz) | OpenTelemetry padrão |
| **ADR-013** | Métricas de Segurança | dict.statistics (fraude) |

### 10.5. ADRs de Estrutura (Novos - Gaps)

| ID | Título | Necessidade Identificada |
|----|--------|--------------------------|
| **ADR-014** | Clean Architecture Layers | Definir camadas domain/application/infra |
| **ADR-015** | API Contract (REST vs gRPC) | Documento não especifica |
| **ADR-016** | SLAs de Performance | Definir p50, p95, p99 |
| **ADR-017** | Circuit Breaker Pattern | Proteção contra falhas BC DICT |
| **ADR-018** | Health Check Strategy | /health, /ready, /metrics |

---

## 11. DIAGRAMA DE ARQUITETURA CONSOLIDADO

Com base no documento, a arquitetura DICT pode ser representada como:

```ascii
┌─────────────────────────────────────────────────────────────────────────────┐
│                         ARQUITETURA DICT LBPAY                               │
└─────────────────────────────────────────────────────────────────────────────┘

┌────────────────┐
│   LB CORE      │
│                │
│  ┌──────────┐  │
│  │   Core   │  │
│  │ (7.4)    │  │
│  └────┬─────┘  │
│       │        │
└───────┼────────┘
        │ usa
        ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  LB-CONNECT (DICT)                                                          │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────┐      │
│  │  dict.api (8.3) - "thin API"                                      │      │
│  ├──────────────────────────────────────────────────────────────────┤      │
│  │  Controller → Model.Validation → Application → Validation         │      │
│  │      ▲              │                  │              │            │      │
│  │      │              │                  │              ▼            │      │
│  │      │              │                  │      ┌────────────────┐  │      │
│  │      │              │                  │      │ Cache Validacao│  │      │
│  │      │              │                  │      │ Chave (Redis)  │  │      │
│  │      │              │                  │      └────────────────┘  │      │
│  │      │              │                  │      ┌────────────────┐  │      │
│  │      │              │                  │      │ Cache Contas   │  │      │
│  │      │              │                  │      │    (Redis)     │  │      │
│  │      │              │                  │      └────────────────┘  │      │
│  │      │              │                  │                           │      │
│  │      │              │                  └──► dict.rsfn.producer     │      │
│  │      │              │                                 │            │      │
│  │      │              └─────► Hash ──► Cache de Resposta (Redis)    │      │
│  │      │                                       (check/save)          │      │
│  │      │                                                              │      │
│  │      └────────────────────────── dict.rsfn.consumer                │      │
│  │                                         │                           │      │
│  └─────────────────────────────────────────┼───────────────────────────┘      │
│                                            │                                  │
│                       ┌────────────────────┴────────────────────┐            │
│                       │      Apache Pulsar (Message Bus)        │            │
│                       │  ┌─────────────────────────────────┐    │            │
│                       │  │ rsfn-dict-req-out               │    │            │
│                       │  │ rsfn-dict-res-out               │    │            │
│                       │  │ nome_da_fila_domain_events      │    │            │
│                       │  │ nome_da_fila_rate_limit         │    │            │
│                       │  │ locks/dict-orchestration-monitor│    │            │
│                       │  └─────────────────────────────────┘    │            │
│                       └─────────────┬───────────────────────────┘            │
│                                     │                                        │
│  ┌──────────────────────────────────┼────────────────────────────────┐      │
│  │  RSFN Connect (7.8)              │                                 │      │
│  ├──────────────────────────────────┴─────────────────────────────────┤      │
│  │  ┌───────────────────────────────────────────────────────────┐    │      │
│  │  │  DICT Proxy (3.1)                                          │    │      │
│  │  │  ├─ Consumer DICT Proxy ◄── rsfn-dict-req-out             │    │      │
│  │  │  ├─ DICT Proxy Application (lógica de integração)         │    │      │
│  │  │  │  └─ Validação Rate Limit ──► Cache Rate Limit (Redis)  │    │      │
│  │  │  ├─ Producer DICT Proxy ──► rsfn-dict-res-out             │    │      │
│  │  │  └─ Producer Rate Limit ──► nome_da_fila_rate_limit       │    │      │
│  │  └───────────────────────────────────────────────────────────┘    │      │
│  │                            │                                       │      │
│  │                            │ mTLS (Cipher 7.3)                     │      │
│  │                            ▼                                       │      │
│  │                 ┌─────────────────────┐                            │      │
│  │                 │   BC DICT (7.2)     │ (External)                 │      │
│  │                 │ Banco Central DICT  │                            │      │
│  │                 └─────────────────────┘                            │      │
│  └────────────────────────────────────────────────────────────────────┘      │
│                                                                               │
│  ┌───────────────────────────────────────────────────────────────────┐      │
│  │  Temporal Workflows                                                │      │
│  ├───────────────────────────────────────────────────────────────────┤      │
│  │  ┌─────────────────────────────────────────────────────────────┐  │      │
│  │  │ Temporal Server (7.9)                                        │  │      │
│  │  │  ├─ dict.orchestration.monitor (8.6)                        │  │      │
│  │  │  │  └─ workflow polling ──► BC DICT polling                 │  │      │
│  │  │  ├─ dict.orchestration.worker (8.7)                         │  │      │
│  │  │  │  ├─ worker.claims (reivindicações)                       │  │      │
│  │  │  │  ├─ worker.entries (vínculo de chaves)                   │  │      │
│  │  │  │  └─ worker respostas orfãs (timeout handling)            │  │      │
│  │  │  └─ dict.vsync (8.9)                                        │  │      │
│  │  │     └─ worker.vsync (sincronização)                         │  │      │
│  │  └─────────────────────────────────────────────────────────────┘  │      │
│  └───────────────────────────────────────────────────────────────────┘      │
│                                                                               │
│  ┌───────────────────────────────────────────────────────────────────┐      │
│  │  Background Workers                                                │      │
│  ├───────────────────────────────────────────────────────────────────┤      │
│  │  ├─ Worker Rate Limit (8.16) - token bucket refill               │      │
│  │  ├─ Sync de Contas (8.14) - sincroniza dados de contas           │      │
│  │  └─ dict.statistics (8.8) - coleta métricas de segurança          │      │
│  └───────────────────────────────────────────────────────────────────┘      │
│                                                                               │
└───────────────────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────────────┐
│  PERSISTÊNCIA                                                               │
├────────────────────────────────────────────────────────────────────────────┤
│  Redis (5 instâncias):                                                      │
│  ├─ Cache de Resposta                                                       │
│  ├─ Cache Contas                                                            │
│  ├─ Cache Validacao Chave                                                   │
│  ├─ Cache Dedup                                                             │
│  └─ Cache Rate Limit                                                        │
│                                                                              │
│  PostgreSQL (2 bases):                                                      │
│  ├─ CID e VSync (diretório principal)                                       │
│  └─ db_statistics (métricas de segurança)                                   │
│                                                                              │
│  Apache Pulsar (6 filas/topics)                                             │
└────────────────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────────────┐
│  OBSERVABILITY                                                              │
├────────────────────────────────────────────────────────────────────────────┤
│  ├─ OpenTelemetry Collector (8.11) - coleta telemetria                     │
│  ├─ Prometheus (8.12) - métricas                                           │
│  ├─ Grafana (8.10) - visualização                                          │
│  ├─ Signoz (8.13) - tracing + logs                                         │
│  └─ dict.dashboard (8.4) - UI de config/monitoramento                      │
└────────────────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────────────┐
│  DATA LAKE                                                                  │
├────────────────────────────────────────────────────────────────────────────┤
│  Audit (7.1) - consome nome_da_fila_domain_events                          │
│  └─ Persistência longa duração + trilha de auditoria                       │
└────────────────────────────────────────────────────────────────────────────┘
```

---

## 12. CONCLUSÕES E PRÓXIMOS PASSOS

### 12.1. Principais Conclusões

1. **Arquitetura AS-IS está bem estruturada**
   - Event-Driven Architecture
   - Caching multi-camada
   - Orquestração com Temporal
   - Observabilidade completa

2. **Bridge DICT é específico, não genérico**
   - RSFN Connect é implementação concreta
   - Não há abstração para outros sistemas Bacen
   - core-dict deve decidir: seguir padrão ou inovar

3. **Performance está preparada para alto volume**
   - 5 caches Redis
   - Comunicação assíncrona (Pulsar)
   - Rate limiting
   - Timeout handling

4. **Gaps críticos para core-dict**
   - Schemas PostgreSQL não documentados
   - SLAs de latência não definidos
   - Estrutura de domínio não explícita
   - API contract (REST/gRPC) indefinido

### 12.2. Próximos Passos Sugeridos

#### Fase Imediata (Sprint 1):

1. **Criar ADRs Críticos** (P0)
   - ADR-006: Bridge específico vs genérico (resolver DUV-003)
   - ADR-015: API Contract REST vs gRPC
   - ADR-016: SLAs de Performance (resolver DUV-012)

2. **Documentar Schemas PostgreSQL**
   - Reverter engenharia de CID e VSync
   - Documentar entities e relationships

3. **Definir Boundaries**
   - O que é core-dict?
   - O que é connector?
   - O que é bridge?

#### Fase Seguinte (Sprint 2):

4. **Criar ADRs de Arquitetura**
   - ADR-001 a ADR-005 (persistência, caching, mensageria)
   - ADR-014: Clean Architecture layers

5. **Implementar PoCs**
   - PoC: Caching Redis (validar padrão)
   - PoC: Pulsar integration
   - PoC: Temporal workflow básico

6. **Definir Observabilidade**
   - Instrumentação OpenTelemetry
   - Dashboards Grafana
   - Alertas Prometheus

---

## ANEXOS

### A. Glossário de Sistemas

| Sistema | ID | Descrição | Status |
|---------|-----|-----------|--------|
| **Audit** | 7.1 | Trilha de auditoria, persistência longa | live |
| **BC DICT** | 7.2 | Diretório DICT Banco Central (externo) | live |
| **Cipher** | 7.3 | Gerenciamento de certificados mTLS | live |
| **Core** | 7.4 | Sistema Core LBPay | live |
| **DICT** | 7.5 | Sistema DICT LBPay | live |
| **Keycloak** | 7.6 | Autenticação/Autorização | live |
| **Observabilidade** | 7.7 | Stack de observabilidade | live |
| **RSFN Connect** | 7.8 | Integração com Bacen (Bridge) | live |
| **Temporal Server** | 7.9 | Orquestração workflows | live |

### B. Glossário de Apps

| App | ID | Tecnologia | Descrição |
|-----|-----|-----------|-----------|
| **dict.api** | 8.3 | Golang | Thin API para requisições DICT |
| **dict.dashboard** | 8.4 | Next.js | UI de configuração/monitoramento |
| **dict.orchestration.monitor** | 8.6 | Golang | Polling notificações BC |
| **dict.orchestration.worker** | 8.7 | Golang | Workflows Temporal |
| **dict.statistics** | 8.8 | Golang | Persistência de estatísticas |
| **dict.vsync** | 8.9 | Golang | Verificação sincronismo |
| **Worker Rate Limit** | 8.16 | Golang | Refill token bucket |
| **Sync de Contas** | 8.14 | Golang | Sincronização dados contas |

### C. Glossário de Stores

| Store | ID | Tecnologia | Descrição | Status |
|-------|-----|-----------|-----------|--------|
| **Cache Contas** | 9.1 | Redis | Dados contas sincronizadas | live |
| **Cache de Resposta** | 9.2 | Redis | Respostas por hash | live |
| **Cache Dedup** | 9.3 | Redis | claims_seen | live |
| **Cache Rate Limit** | 9.4 | Redis | Token bucket | future |
| **Cache Validacao Chave** | 9.5 | Redis | Validações PIX | live |
| **CID e VSync** | 9.6 | PostgreSQL | Base principal DICT | live |
| **db_statistics** | 9.7 | PostgreSQL | Estatísticas segurança | live |
| **nome_da_fila_domain_events** | 9.9 | Pulsar | Eventos de estado | live |
| **nome_da_fila_rate_limit** | 9.10 | Pulsar | Eventos rate limit | live |
| **rsfn-dict-req-out** | 9.12 | Pulsar | Requisições BC DICT | live |
| **rsfn-dict-res-out** | 9.13 | Pulsar | Respostas BC DICT | live |

### D. Referências

1. **Documento Original**: `/Users/jose.silva.lb/LBPay/IA_Dict/Docs_iniciais/ArquiteturaDict_LBPAY.md`
2. **IcePanel**: Ferramenta de diagramação arquitetural usada
3. **Temporal**: https://docs.temporal.io
4. **Apache Pulsar**: https://pulsar.apache.org/docs
5. **OpenTelemetry**: https://opentelemetry.io

---

**FIM DO DOCUMENTO ARE-003**

---

**Metadados do Documento**:
- Total de linhas analisadas: 1760
- Total de sistemas identificados: 9
- Total de apps identificados: 16
- Total de stores identificados: 14
- Total de componentes identificados: 26
- Total de ADRs propostos: 18
- Tempo de análise: ~45 minutos (agente NEXUS)
