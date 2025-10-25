# ANA-001: Análise Detalhada - ArquiteturaDict_LBPAY.md (IcePanel)

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Documento Analisado**: `Docs_iniciais/ArquiteturaDict_LBPAY.md`
**Ferramenta de Origem**: IcePanel (https://icepanel.io)
**Data da Análise**: 2025-10-25
**Analista**: ARCHITECT (AI Agent - Technical Architect)
**Versão**: 1.0

---

## Sumário Executivo

Este documento apresenta uma **análise detalhada** do documento de arquitetura `ArquiteturaDict_LBPAY.md`, gerado pela ferramenta **IcePanel**. O documento original contém diagramas SVG embutidos que representam a arquitetura completa do sistema DICT da LBPay.

### Principais Descobertas

1. ✅ **Confirmação Arquitetural**: A arquitetura descrita no IcePanel **confirma** o fluxo correto identificado anteriormente:
   - `Core Bancário DICT (dict.api)` → `RSFN Connect` → `DICT Proxy (Bridge)` → `Banco Central (BC DICT)`

2. ✅ **Temporal Workflows**: O documento confirma que os **workflows Temporal** estão no componente `dict.orchestration.worker`, que faz parte do sistema **DICT**, não do Bridge.

3. ✅ **Separação de Responsabilidades**:
   - **DICT Proxy**: Adapter para mTLS com Bacen (confirmado como "Proxy (adapter) para conexão segura (mTLS) e robusta com o DICT no BCB")
   - **RSFN Connect**: Sistema dentro do grupo LB-Connect
   - **dict.orchestration.worker**: Contém workflows Temporal de processamento

4. ⚠️ **Nomenclatura Inconsistente**: Há sobreposição/confusão entre:
   - "RSFN Connect" (sistema)
   - "DICT Proxy" (app)
   - "dict.api" (app)
   - Necessário clarificar qual é qual no contexto TEC-002 e TEC-003

---

## Índice

1. [Estrutura do Documento](#1-estrutura-do-documento)
2. [Análise de Context Diagrams](#2-análise-de-context-diagrams)
3. [Análise de App Diagrams](#3-análise-de-app-diagrams)
4. [Análise de Component Diagrams](#4-análise-de-component-diagrams)
5. [Análise de Actors e Groups](#5-análise-de-actors-e-groups)
6. [Análise de Systems](#6-análise-de-systems)
7. [Análise de Apps](#7-análise-de-apps)
8. [Análise de Stores (Filas, Caches, DBs)](#8-análise-de-stores-filas-caches-dbs)
9. [Análise de Components](#9-análise-de-components)
10. [Mapeamento: IcePanel → TEC-002/TEC-003](#10-mapeamento-icepanel--tec-002tec-003)
11. [Conclusões e Recomendações](#11-conclusões-e-recomendações)

---

## 1. Estrutura do Documento

O documento `ArquiteturaDict_LBPAY.md` foi gerado pela ferramenta **IcePanel** e possui a seguinte estrutura:

### 1.1. Metadados

- **Landscape**: Rui Holdorf's landscape
- **Data de Geração**: 2025-10-23 22:18:49
- **Ferramenta**: IcePanel (https://icepanel.io)

### 1.2. Estrutura de Conteúdo

```
1. Context Diagrams (2 diagramas)
   ├─ 1.1. Context Diagram
   └─ 1.2. Context Diagram 2

2. App Diagrams (6 diagramas)
   ├─ 2.1. Audit App Diagram
   ├─ 2.2. BC DICT App Diagram 2
   ├─ 2.3. DICT App Diagram
   ├─ 2.4. Observabilidade App Diagram
   ├─ 2.5. RSFN Connect App Diagram ⭐ CRÍTICO
   └─ 2.6. Temporal Server App Diagram ⭐ CRÍTICO

3. Component Diagrams (9 diagramas)
   ├─ 3.1. DICT Proxy Component Diagram ⭐ CRÍTICO
   ├─ 3.2. dict.api Component Diagram ⭐ CRÍTICO
   ├─ 3.3. dict.dashboard Component Diagram
   ├─ 3.4. dict.orchestration.monitor Component Diagram ⭐ CRÍTICO
   ├─ 3.5. dict.orchestration.worker Component Diagram ⭐ CRÍTICO
   ├─ 3.6. dict.statistics Component Diagram
   ├─ 3.7. dict.vsync Component Diagram ⭐ CRÍTICO
   ├─ 3.8. Sync de Contas Component Diagram
   └─ 3.9. Worker Rate Limit Component Diagram

4. Component Diagram Flows (1 fluxo)
   └─ 4.1. Requisição CRUD ⭐ CRÍTICO

5. Actors (1 ator)
   └─ 5.1. Outras IPs

6. Groups (4 grupos)
   ├─ 6.1. Banco Central ⭐ CRÍTICO
   ├─ 6.2. Data Lake
   ├─ 6.3. LB Core ⭐ CRÍTICO
   └─ 6.4. LB-Connect ⭐ CRÍTICO

7. Systems (9 sistemas)
   ├─ 7.1. Audit
   ├─ 7.2. BC DICT ⭐ CRÍTICO
   ├─ 7.3. Cipher
   ├─ 7.4. Core
   ├─ 7.5. DICT ⭐ CRÍTICO
   ├─ 7.6. Keycloak
   ├─ 7.7. Observabilidade
   ├─ 7.8. RSFN Connect ⭐ CRÍTICO
   └─ 7.9. Temporal Server ⭐ CRÍTICO

8. Apps (16 aplicações)
   ├─ 8.1. API DICT ⭐ CRÍTICO
   ├─ 8.2. DICT Proxy ⭐ CRÍTICO
   ├─ 8.3. dict.api ⭐ CRÍTICO
   ├─ 8.4. dict.dashboard
   ├─ 8.5. dict.monitor
   ├─ 8.6. dict.orchestration.monitor ⭐ CRÍTICO
   ├─ 8.7. dict.orchestration.worker ⭐ CRÍTICO
   ├─ 8.8. dict.statistics
   ├─ 8.9. dict.vsync ⭐ CRÍTICO
   ├─ 8.10. Grafana
   ├─ 8.11. Open Telemetry Collector
   ├─ 8.12. Prometheus
   ├─ 8.13. Signoz
   ├─ 8.14. Sync de Contas
   ├─ 8.15. Temporal ⭐ CRÍTICO
   └─ 8.16. Worker Rate Limit

9. Stores (14 stores)
   ├─ 9.1. Cache Contas
   ├─ 9.2. Cache de Resposta
   ├─ 9.3. Cache Dedup
   ├─ 9.4. Cache Rate Limit
   ├─ 9.5. Cache Validacao Chave
   ├─ 9.6. CID e VSync ⭐ CRÍTICO
   ├─ 9.7. db_statistics
   ├─ 9.8. locks/dict-orchestration-monitor
   ├─ 9.9. nome_da_fila_domain_events
   ├─ 9.10. nome_da_fila_rate_limit
   ├─ 9.11. nome_da_fila_sync_contas
   ├─ 9.12. rsfn-dict-req-out ⭐ CRÍTICO
   ├─ 9.13. rsfn-dict-res-out ⭐ CRÍTICO
   └─ 9.14. store_prometheus

10. Components (26 componentes)
    ├─ 10.1. Autenticação e Autorização ⭐ CRÍTICO
    ├─ 10.2. Consumer DICT Proxy
    ├─ 10.3. Consumer Sync Contas
    ├─ 10.4. dic.vsync.domainevents.consumer
    ├─ 10.5. DICT Proxy Applicaton
    ├─ 10.6. dict.api.application ⭐ CRÍTICO
    ├─ 10.7. dict.api.controller ⭐ CRÍTICO
    ├─ 10.8. dict.api.model.validation
    ├─ 10.9. dict.api.validation
    ├─ 10.10. dict.dashboard.api
    ├─ 10.11. dict.dashboard.web
    ├─ 10.12. dict.domainevents.producer
    ├─ 10.13. dict.rsfn.consumer ⭐ CRÍTICO
    ├─ 10.14. dict.rsfn.producer ⭐ CRÍTICO
    ├─ 10.15. dict.statistics.collector
    ├─ 10.16. dict.statistics.summarizer
    ├─ 10.17. Producer DICT Proxy
    ├─ 10.18. Producer Rate Limit
    ├─ 10.19. ratelimit.consumer
    ├─ 10.20. ratelimit.refill
    ├─ 10.21. Validação Rate Limit
    ├─ 10.22. worker respostas orfãs
    ├─ 10.23. worker.claims ⭐ CRÍTICO
    ├─ 10.24. worker.entries ⭐ CRÍTICO
    ├─ 10.25. worker.vsync ⭐ CRÍTICO
    └─ 10.26. workflow polling ⭐ CRÍTICO
```

**⭐ Itens Marcados como CRÍTICO**: São componentes/diagramas essenciais para entender a arquitetura DICT e a separação entre Connect e Bridge.

---

## 2. Análise de Context Diagrams

### 2.1. Context Diagram (Nível 1)

**Propósito**: Visão de alto nível dos sistemas e suas interações.

**Componentes Identificados** (extraídos dos SVGs embutidos):

Embora os SVGs estejam codificados em base64, baseado na estrutura do documento, este diagrama provavelmente mostra:

1. **DICT** (Sistema interno LBPay)
2. **RSFN Connect** (Sistema interno LBPay no grupo LB-Connect)
3. **BC DICT** (Sistema externo - Banco Central)
4. **Temporal Server** (Sistema interno para workflows)
5. **Observabilidade** (Sistema interno para monitoramento)

**Interações Esperadas**:
- DICT ↔ RSFN Connect
- RSFN Connect ↔ BC DICT
- DICT ↔ Temporal Server
- Sistemas → Observabilidade

### 2.2. Context Diagram 2 (Nível 2)

**Propósito**: Refinamento do Context Diagram 1 com mais detalhes.

**Observação**: Sem acesso direto aos SVGs renderizados, baseio-me na estrutura textual que indica componentes adicionais como Data Lake, Audit, Core.

---

## 3. Análise de App Diagrams

### 3.1. Audit App Diagram

**Sistema**: Audit
**Status**: live
**Propósito**: Auditoria de operações

**Tecnologias**: Não especificadas no texto

### 3.2. BC DICT App Diagram 2

**Sistema**: BC DICT (Banco Central)
**Status**: live (sistema externo)
**Descrição**: "Diretório de Identificadores Contas Transacionais"

**Componente Interno**:
- **API DICT**: API externa fornecida pelo Banco Central

**Observação**: Este é o sistema EXTERNO do Bacen que recebe requisições SOAP/XML via mTLS.

### 3.3. DICT App Diagram ⭐

**Sistema**: DICT (LBPay interno)
**Status**: live
**Descrição**: "Diretório de Identificadores de Contas Transacionais"

**Apps Internos** (confirmados):
1. **dict.api** (Golang)
2. **dict.dashboard** (Golang)
3. **dict.monitor** (Golang)
4. **dict.orchestration.monitor** (Golang)
5. **dict.orchestration.worker** (Golang) ⭐ **Contém workflows Temporal**
6. **dict.statistics** (Golang)
7. **dict.vsync** (Golang)
8. **Sync de Contas** (Golang)
9. **Worker Rate Limit** (Golang)

**Tecnologias**: Golang (Go 1.22+)

**Análise Crítica**:
- ✅ **dict.orchestration.worker** é explicitamente descrito como "Contém os workflows Temporal de processamento"
- ✅ Confirma que Temporal Workflows estão dentro do sistema **DICT**, não no Bridge

### 3.4. Observabilidade App Diagram

**Sistema**: Observabilidade
**Status**: live

**Apps Internos**:
1. **Grafana** (Visualização)
2. **Open Telemetry Collector** (Coleta de telemetria)
3. **Prometheus** (Time-series database)
4. **Signoz** (Observabilidade)

**Stores**:
- **store_prometheus**: Armazenamento Prometheus

**Função**: Coletor de telemetria, persiste em base Prometheus

### 3.5. RSFN Connect App Diagram ⭐⭐⭐ CRÍTICO

**Sistema**: RSFN Connect
**Status**: live
**Grupo**: LB-Connect
**Descrição**: "Rede do Sistema Financeiro Nacional"

**Apps Visualizados no Diagrama**:

Baseado na análise dos textos e referências cruzadas, este diagrama mostra:

1. **DICT Proxy** (Golang)
   - Descrição: "Proxy (adapter) para conexão segura (mTLS) e robusta com o DICT no BCB"
   - **ESTA É A CONFIRMAÇÃO**: DICT Proxy = Bridge (TEC-002)

2. **Worker Rate Limit** (Golang)
   - Descrição: "repõe tokens conforme programado, consome eventos de Rate Limit"
   - Função: Gerenciamento de Rate Limiting (Token Bucket)

**Relações Identificadas** (baseado em referências cruzadas):
- DICT (sistema) → RSFN Connect (sistema)
- RSFN Connect → BC DICT (externo)
- DICT Proxy ↔ API DICT (Bacen)

**Observações Arquiteturais**:

🔴 **INCONSISTÊNCIA CRÍTICA DETECTADA**:

O diagrama "RSFN Connect App Diagram" parece mostrar apenas:
- DICT Proxy (Bridge)
- Worker Rate Limit

**MAS NÃO MOSTRA** explicitamente um componente chamado "RSFN Connect" como aplicação.

**Hipótese de Interpretação**:
1. **"RSFN Connect" (Sistema)** = Nome do grupo/contexto que contém o Bridge
2. **"DICT Proxy" (App)** = O que chamamos de Bridge (TEC-002)
3. **"dict.api + dict.orchestration.worker" (Apps)** = O que chamamos de Connect (TEC-003)?

**Necessita Clarificação**: A nomenclatura "RSFN Connect" no IcePanel não corresponde 1:1 com "Connect" em TEC-003.

### 3.6. Temporal Server App Diagram ⭐⭐

**Sistema**: Temporal Server
**Status**: live
**Descrição**: "Orquestração com workflows, agendamentos, triggers"

**Apps**:
- **Temporal** (Temporal Server standalone)

**Diagramas Relacionados**:
- dict.orchestration.monitor Component Diagram
- dict.orchestration.worker Component Diagram
- dict.vsync Component Diagram

**Confirmação**: Temporal Server é usado para orquestração de workflows.

---

## 4. Análise de Component Diagrams

### 4.1. DICT Proxy Component Diagram ⭐⭐⭐

**App**: DICT Proxy (Golang)
**Descrição**: "Proxy (adapter) para conexão segura (mTLS) e robusta com o DICT no BCB"

**Componentes Internos Identificados**:
1. **DICT Proxy Application**
2. **Consumer DICT Proxy** (Pulsar Consumer)
3. **Producer DICT Proxy** (Pulsar Producer)
4. **Autenticação e Autorização** (mTLS)

**Stores Conectados**:
- **rsfn-dict-req-out** (Pulsar topic - requisições)
- **rsfn-dict-res-out** (Pulsar topic - respostas)

**Sistemas Externos Conectados**:
- **API DICT** (BC DICT - Bacen)

**Fluxo Identificado**:
```
dict.api → rsfn-dict-req-out (Pulsar)
    → Consumer DICT Proxy
    → DICT Proxy Application
    → API DICT (Bacen mTLS)
    → Producer DICT Proxy
    → rsfn-dict-res-out (Pulsar)
    → dict.api
```

**Análise Crítica**:
- ✅ **DICT Proxy** é claramente um **adapter/proxy** para mTLS com Bacen
- ✅ Usa Pulsar para comunicação assíncrona com dict.api
- ✅ **CORRESPONDE EXATAMENTE AO BRIDGE (TEC-002)**

### 4.2. dict.api Component Diagram ⭐⭐⭐

**App**: dict.api (Golang)
**Descrição**: (API principal do DICT LBPay)

**Componentes Internos Identificados**:
1. **dict.api.controller** (Controladores gRPC/REST)
2. **dict.api.application** (Lógica de aplicação)
3. **dict.api.validation** (Validações de negócio)
4. **dict.api.model.validation** (Validações de modelo)
5. **Autenticação e Autorização** (mTLS para frontends)
6. **dict.rsfn.producer** (Pulsar Producer para RSFN)
7. **dict.rsfn.consumer** (Pulsar Consumer de respostas RSFN)

**Stores Conectados**:
- **rsfn-dict-req-out** (Pulsar topic - envia requisições)
- **rsfn-dict-res-out** (Pulsar topic - recebe respostas)
- **Cache de Resposta**
- **Cache Validacao Chave**
- **Cache Dedup**
- **nome_da_fila_domain_events** (Pulsar - eventos de domínio)

**Sistemas Externos Conectados**:
- **RSFN Connect** (sistema)
- **DICT Proxy** (app - via Pulsar)

**Fluxo Identificado**:
```
Frontend (gRPC/REST)
    → dict.api.controller
    → dict.api.validation
    → dict.api.application
    → dict.rsfn.producer
    → rsfn-dict-req-out (Pulsar)
    → [DICT Proxy processa via Bacen]
    → rsfn-dict-res-out (Pulsar)
    → dict.rsfn.consumer
    → dict.api.application
    → dict.api.controller
    → Frontend (response)
```

**Análise Crítica**:
- ✅ **dict.api** é o **Core Bancário DICT** (ponto de entrada)
- ✅ Produz mensagens para **rsfn-dict-req-out** (Pulsar)
- ✅ Consome respostas de **rsfn-dict-res-out** (Pulsar)
- ✅ **NÃO possui Temporal Workflows** (apenas produz/consome Pulsar)

### 4.3. dict.dashboard Component Diagram

**App**: dict.dashboard (Golang)

**Componentes Internos**:
1. **dict.dashboard.api** (Backend API)
2. **dict.dashboard.web** (Frontend Web)

**Stores Conectados**:
- **store_prometheus** (Prometheus database)

**Função**: Dashboard de monitoramento e visualização.

### 4.4. dict.orchestration.monitor Component Diagram ⭐⭐

**App**: dict.orchestration.monitor (Golang)

**Componentes Internos**:
1. **workflow polling** (Polling de workflows Temporal)

**Stores Conectados**:
- **locks/dict-orchestration-monitor** (Distributed locks)

**Sistemas Externos Conectados**:
- **RSFN Connect** (sistema)
- **Temporal Server** (sistema)

**Função**: Monitora workflows Temporal, realiza polling no Bacen para encontrar:
- Retornos
- Inícios de reivindicação
- Estatísticas

**Análise Crítica**:
- ✅ Confirma que existe um componente de **polling** para workflows Temporal
- ✅ Interage com RSFN Connect e Temporal Server
- ⚠️ Necessita entender se este polling é para Claims (7 dias) ou VSYNC

### 4.5. dict.orchestration.worker Component Diagram ⭐⭐⭐ CRÍTICO

**App**: dict.orchestration.worker (Golang)
**Descrição**: "Contém os workflows Temporal de processamento"

**Componentes Internos (Workers Temporal)**:
1. **worker.claims** ⭐
   - Descrição: "activities relacionadas à reivindicação"
   - Tecnologia: Golang
   - **FUNÇÃO**: Activities para ClaimWorkflow (7 dias)

2. **worker.entries** ⭐
   - Descrição: "activities relacionadas à vinculo de chaves"
   - Tecnologia: Golang
   - **FUNÇÃO**: Activities para criação/atualização de entries

3. **worker.vsync**
   - Descrição: "activities relacionadas à sincronismo"
   - **FUNÇÃO**: Activities para VSYNCWorkflow

**Sistemas Externos Conectados**:
- **RSFN Connect** (sistema)
- **Temporal Server** (sistema)

**Análise Crítica**:
- ✅ **CONFIRMAÇÃO DEFINITIVA**: dict.orchestration.worker contém os **Workers Temporal**
- ✅ Possui workers separados para Claims, Entries e VSYNC
- ✅ Estes são os **Temporal Activities** executados pelos Workflows
- ❌ **NÃO está no DICT Proxy (Bridge)**, está em **dict.orchestration.worker**

**Implicação Arquitetural**:
```
dict.orchestration.worker (Temporal Workers + Activities)
    ↓ chama
DICT Proxy / RSFN Connect (Bridge - Adapter SOAP/mTLS)
    ↓ chama
BC DICT (Bacen)
```

### 4.6. dict.statistics Component Diagram

**App**: dict.statistics (Golang)
**Descrição**: "persiste informações de estatística e indicadores"

**Componentes Internos**:
1. **dict.statistics.collector** (Coletor)
2. **dict.statistics.summarizer** (Sumarizador)

**Stores Conectados**:
- **db_statistics** (Database PostgreSQL)

**Função**: Coleta e fornece informações de segurança do DICT Bacen:

**Tipos de Estatísticas** (conforme especificação detalhada no documento):

1. **getEntryStatistics** (vinculada a chave PIX):
   - Quantidade de liquidações como recebedor no SPI
   - Notificações de infração confirmadas por tipo de fraude:
     - Falsidade ideológica (ApplicationFrauds)
     - Conta laranja (MuleAccounts)
     - Conta fraudador (ScammerAccounts)
     - Outros (OtherFrauds)
   - Valor total de notificações confirmadas
   - Quantidade de participantes reportando fraudes
   - Notificações abertas/não fechadas
   - Notificações rejeitadas

2. **getPersonStatistics** (vinculada a CPF/CNPJ):
   - Mesmas informações de getEntryStatistics
   - Quantidade de contas vinculadas a chaves PIX

**Períodos de Agregação**:
- Últimos 90 dias
- Últimos 12 meses (sem considerar mês atual)
- Últimos 60 meses (sem considerar mês atual)

**Watermark**: Informações atualizadas com atraso máximo de **12 horas**.

**Análise Crítica**:
- ✅ Sistema crítico para **análise de fraude** e **segurança**
- ✅ Consome dados do Bacen via DICT Proxy
- ✅ Persiste em banco PostgreSQL local

### 4.7. dict.vsync Component Diagram ⭐⭐

**App**: dict.vsync (Golang)
**Descrição**: "Verificação de sincronismo"

**Componentes Internos**:
1. **worker.vsync** (Activities Temporal para VSYNC)
2. **dic.vsync.domainevents.consumer** (Consumer de eventos de domínio)

**Stores Conectados**:
- **CID e VSync** (Database PostgreSQL)
- **rsfn-dict-req-out** (Pulsar - requisições)
- **rsfn-dict-res-out** (Pulsar - respostas)
- **nome_da_fila_domain_events** (Pulsar - eventos de domínio)

**Sistemas Externos Conectados**:
- **RSFN Connect** (sistema)
- **Temporal Server** (sistema)

**Função**: Sincronização diária entre base local (CID) e DICT Bacen.

**Fluxo VSYNC** (inferido):
```
Temporal Schedule (cron diário)
    → VSYNCWorkflow (Temporal)
    → worker.vsync (Activities)
    → rsfn-dict-req-out (Pulsar - solicita todas as entries do Bacen)
    → DICT Proxy (busca entries via Bacen)
    → rsfn-dict-res-out (Pulsar - retorna entries)
    → worker.vsync (compara com CID local)
    → CID e VSync (Database - atualiza diferenças)
    → nome_da_fila_domain_events (Pulsar - publica eventos de divergência)
```

**Análise Crítica**:
- ✅ **VSYNC é um Temporal Workflow** executado por dict.vsync
- ✅ Usa **worker.vsync** (Temporal Activities)
- ✅ Comunica com DICT Proxy via Pulsar (rsfn-dict-req-out / rsfn-dict-res-out)
- ✅ Persiste resultados em **CID e VSync** (PostgreSQL)

### 4.8. Sync de Contas Component Diagram

**App**: Sync de Contas (Golang)

**Componentes Internos**:
1. **Consumer Sync Contas** (Consumer Pulsar)

**Stores Conectados**:
- **nome_da_fila_sync_contas** (Pulsar)
- **Cache Contas** (Redis)

**Função**: Sincronização de contas do Core Bancário com cache local.

### 4.9. Worker Rate Limit Component Diagram

**App**: Worker Rate Limit (Golang)
**Descrição**: "repõe tokens conforme programado, consome eventos de Rate Limit"

**Componentes Internos**:
1. **ratelimit.consumer** (Consumer de eventos)
2. **ratelimit.refill** (Refill de tokens - Token Bucket)
3. **Validação Rate Limit** (Validação de limites)
4. **Producer Rate Limit** (Producer Pulsar)

**Stores Conectados**:
- **nome_da_fila_rate_limit** (Pulsar)
- **Cache Rate Limit** (Redis)

**Função**: Implementação de **Rate Limiting** usando algoritmo **Token Bucket**:
- Consumir eventos de requisições
- Validar limites de taxa
- Refazer tokens periodicamente
- Bloquear requisições que excedem limite

**Diagrama Relacionado**:
- RSFN Connect App Diagram

**Análise Crítica**:
- ✅ Rate Limiting está implementado como **worker separado**
- ✅ Usa **Token Bucket algorithm** conforme API-001
- ✅ Faz parte do **RSFN Connect App Diagram**

---

## 5. Análise de Actors e Groups

### 5.1. Actor: Outras IPs

**Tipo**: External actor
**Descrição**: Outras instituições de pagamento

**Função**: Representa outras IPs que interagem com o sistema DICT (consultas, reivindicações, etc.)

### 5.2. Groups

#### 6.1. Banco Central ⭐

**Sistemas Internos**:
- **BC DICT** (Sistema externo - Diretório DICT do Bacen)
  - **API DICT** (API SOAP/XML com mTLS)

**Função**: Grupo representando o Banco Central do Brasil e seu sistema DICT.

#### 6.2. Data Lake

**Função**: Armazenamento de dados históricos e analytics.

**Observação**: Não há detalhes suficientes no documento sobre apps/stores dentro deste grupo.

#### 6.3. LB Core ⭐

**Sistemas Internos**:
- **Core** (Core Bancário LBPay)
- **Cipher** (Serviço de criptografia)
- **Keycloak** (Autenticação e autorização)

**Função**: Grupo representando o Core Bancário da LBPay.

#### 6.4. LB-Connect ⭐⭐

**Sistemas Internos**:
- **RSFN Connect** (Sistema de conexão com RSFN)

**Função**: Grupo representando componentes de conexão com o RSFN (Rede do Sistema Financeiro Nacional).

**Análise Crítica**:
- ✅ **LB-Connect** é um **grupo** que contém o sistema **RSFN Connect**
- ✅ RSFN Connect parece ser o **contexto/sistema** que engloba DICT Proxy (Bridge)
- ⚠️ **Inconsistência**: No TEC-003, chamamos de "Connect" o orquestrador com Temporal, mas aqui "RSFN Connect" parece ser apenas o grupo que contém o Bridge

---

## 6. Análise de Systems

### 6.1. Audit

**Status**: live
**Descrição**: Sistema de auditoria

**Diagramas**: Audit App Diagram

### 6.2. BC DICT ⭐⭐⭐

**Tipo**: External system (Banco Central)
**Status**: live
**Descrição**: "Diretório de Identificadores Contas Transacionais"

**Grupo**: Banco Central

**Apps Internos**:
- **API DICT** (API SOAP/XML fornecida pelo Bacen)

**Diagramas**:
- BC DICT App Diagram 2
- DICT Proxy Component Diagram
- RSFN Connect App Diagram

**Análise Crítica**:
- ✅ Sistema **EXTERNO** do Bacen
- ✅ Recebe chamadas SOAP/XML via mTLS do **DICT Proxy**

### 6.3. Cipher

**Status**: live
**Grupo**: LB Core

**Função**: Serviço de criptografia/descriptografia.

### 6.4. Core

**Status**: live
**Grupo**: LB Core

**Função**: Core Bancário da LBPay.

### 6.5. DICT ⭐⭐⭐

**Tipo**: Internal system
**Status**: live
**Descrição**: "Diretório de Identificadores de Contas Transacionais"

**Apps Internos**:
1. **dict.api** (API principal)
2. **dict.dashboard** (Dashboard web)
3. **dict.monitor** (Monitor)
4. **dict.orchestration.monitor** (Monitor de workflows Temporal)
5. **dict.orchestration.worker** ⭐ (Workers Temporal)
6. **dict.statistics** (Estatísticas e fraude)
7. **dict.vsync** ⭐ (Verificação de sincronismo)
8. **Sync de Contas** (Sincronização de contas)
9. **Worker Rate Limit** (Rate limiting)

**Diagramas**:
- Context Diagram
- DICT App Diagram
- dict.api Component Diagram
- dict.dashboard Component Diagram
- dict.orchestration.monitor Component Diagram
- dict.orchestration.worker Component Diagram
- dict.statistics Component Diagram
- dict.vsync Component Diagram
- Sync de Contas Component Diagram
- Worker Rate Limit Component Diagram

**Análise Crítica**:
- ✅ **Sistema DICT LBPay** é o sistema **PRINCIPAL**
- ✅ Contém **dict.orchestration.worker** (Temporal Workflows)
- ✅ Contém **dict.api** (Core Bancário DICT - ponto de entrada)
- ✅ **NÃO contém DICT Proxy** (DICT Proxy está em RSFN Connect)

### 6.6. Keycloak

**Status**: live
**Grupo**: LB Core

**Função**: Autenticação e autorização (OAuth2, OpenID Connect).

### 6.7. Observabilidade

**Status**: live

**Apps Internos**:
- Grafana
- Open Telemetry Collector
- Prometheus
- Signoz

**Função**: Monitoramento, observabilidade, telemetria.

### 6.8. RSFN Connect ⭐⭐⭐

**Tipo**: Internal system
**Status**: live
**Descrição**: "Rede do Sistema Financeiro Nacional"
**Grupo**: LB-Connect

**Apps Esperados** (baseado em RSFN Connect App Diagram):
- **DICT Proxy** (Golang) - "Proxy (adapter) para conexão segura (mTLS) e robusta com o DICT no BCB"
- **Worker Rate Limit** (Golang)

**Diagramas**:
- Context Diagram
- dict.api Component Diagram
- dict.orchestration.monitor Component Diagram
- dict.orchestration.worker Component Diagram
- dict.vsync Component Diagram
- RSFN Connect App Diagram

**Análise Crítica**:
- ⚠️ **NOMENCLATURA INCONSISTENTE**: "RSFN Connect" no IcePanel parece ser apenas o **grupo/contexto** que contém o Bridge
- ✅ **DICT Proxy** (dentro de RSFN Connect) = **Bridge (TEC-002)**
- ❌ **NÃO há app chamado "Connect"** dentro de RSFN Connect
- 🔍 **Hipótese**: "Connect" em TEC-003 pode ser **dict.orchestration.worker + dict.api** combinados?

### 6.9. Temporal Server ⭐⭐

**Tipo**: Internal system
**Status**: live
**Descrição**: "Orquestração com workflows, agendamentos, triggers"

**Apps Internos**:
- **Temporal** (Temporal Server)

**Diagramas**:
- Context Diagram
- dict.orchestration.monitor Component Diagram
- dict.orchestration.worker Component Diagram
- dict.vsync Component Diagram
- Temporal Server App Diagram

**Análise Crítica**:
- ✅ Temporal Server é usado por **dict.orchestration.worker** e **dict.vsync**
- ✅ Workflows executam em **dict.orchestration.worker**, não no Bridge

---

## 7. Análise de Apps

### 7.1. API DICT (Bacen)

**Tipo**: External app
**Status**: live
**Sistema**: BC DICT (Banco Central)

**Tecnologias**: API (SOAP/XML)

**Diagramas**:
- BC DICT App Diagram 2
- DICT Proxy Component Diagram
- RSFN Connect App Diagram

**Função**: API do DICT fornecida pelo Bacen (endpoint SOAP/XML com mTLS).

### 7.2. DICT Proxy ⭐⭐⭐

**Tipo**: Internal app
**Status**: live
**Sistema**: RSFN Connect
**Descrição**: "Proxy (adapter) para conexão segura (mTLS) e robusta com o DICT no BCB"

**Tecnologias**: Golang

**Diagramas**:
- DICT Proxy Component Diagram
- dict.api Component Diagram
- dict.orchestration.monitor Component Diagram
- dict.orchestration.worker Component Diagram
- RSFN Connect App Diagram
- Requisição CRUD (flow)

**Função**: **Adapter SOAP/mTLS** para Bacen.

**Análise Crítica**:
- ✅ **DICT Proxy = Bridge (TEC-002)** - CONFIRMADO
- ✅ Descrição confirma: "Proxy (adapter) para conexão segura (mTLS)"
- ✅ Usa Golang
- ✅ Conecta com API DICT (Bacen)

### 7.3. dict.api ⭐⭐⭐

**Tipo**: Internal app
**Status**: live
**Sistema**: DICT

**Tecnologias**: Golang

**Diagramas**:
- DICT App Diagram
- dict.api Component Diagram

**Componentes**:
- dict.api.controller
- dict.api.application
- dict.api.validation
- dict.api.model.validation
- Autenticação e Autorização
- dict.rsfn.producer
- dict.rsfn.consumer

**Função**: **API principal do DICT LBPay** (ponto de entrada via gRPC/REST).

**Análise Crítica**:
- ✅ **Core Bancário DICT** (recebe requisições de frontends)
- ✅ **Produz** mensagens para Pulsar (rsfn-dict-req-out)
- ✅ **Consome** respostas de Pulsar (rsfn-dict-res-out)
- ❌ **NÃO possui Temporal Workflows** (apenas Pulsar Producer/Consumer)

### 7.4. dict.dashboard

**Sistema**: DICT
**Tecnologias**: Golang

**Componentes**:
- dict.dashboard.api (Backend)
- dict.dashboard.web (Frontend)

**Função**: Dashboard de visualização e monitoramento.

### 7.5. dict.monitor

**Sistema**: DICT
**Tecnologias**: Golang

**Função**: Monitor do sistema DICT.

### 7.6. dict.orchestration.monitor ⭐⭐

**Sistema**: DICT
**Tecnologias**: Golang

**Componentes**:
- workflow polling

**Função**: Monitor de workflows Temporal, realiza polling no Bacen.

**Diagramas**:
- DICT App Diagram
- dict.orchestration.monitor Component Diagram

### 7.7. dict.orchestration.worker ⭐⭐⭐

**Sistema**: DICT
**Status**: live
**Descrição**: "Contém os workflows Temporal de processamento"
**Tecnologias**: Golang

**Componentes (Workers Temporal)**:
- worker.claims ⭐ (Reivindicações)
- worker.entries ⭐ (Vínculo de chaves)
- worker.vsync (Sincronismo)

**Diagramas**:
- DICT App Diagram
- dict.orchestration.worker Component Diagram

**Análise Crítica**:
- ✅ **Este é o CORAÇÃO dos Temporal Workflows**
- ✅ Contém workers para Claims, Entries e VSYNC
- ✅ Executa **Temporal Activities** chamadas pelos Workflows
- ✅ **NÃO está no Bridge, está no sistema DICT**

### 7.8. dict.statistics

**Sistema**: DICT
**Tecnologias**: Golang
**Descrição**: "persiste informações de estatística e indicadores"

**Componentes**:
- dict.statistics.collector
- dict.statistics.summarizer

**Função**: Coleta e persiste estatísticas de fraude e segurança do DICT Bacen.

### 7.9. dict.vsync ⭐⭐

**Sistema**: DICT
**Status**: live
**Descrição**: "Verificação de sincronismo"
**Tecnologias**: Golang

**Componentes**:
- worker.vsync (Activities Temporal)
- dic.vsync.domainevents.consumer

**Função**: VSYNC diário (sincronização entre CID local e DICT Bacen).

**Diagramas**:
- DICT App Diagram
- dict.vsync Component Diagram

### 7.10. Grafana

**Sistema**: Observabilidade
**Tecnologias**: Grafana

**Função**: Visualização de métricas.

### 7.11. Open Telemetry Collector

**Sistema**: Observabilidade
**Descrição**: "coletor de telemetria, persiste em uma base Prometheus"
**Tecnologias**: OpenTelemetry

**Função**: Coleta de telemetria (traces, metrics, logs).

### 7.12. Prometheus

**Sistema**: Observabilidade
**Tecnologias**: Prometheus

**Função**: Time-series database para métricas.

### 7.13. Signoz

**Sistema**: Observabilidade

**Função**: Plataforma de observabilidade (alternativa a Jaeger/Zipkin).

### 7.14. Sync de Contas

**Sistema**: DICT
**Tecnologias**: Golang

**Função**: Sincronização de contas do Core Bancário.

### 7.15. Temporal ⭐⭐

**Sistema**: Temporal Server
**Tecnologias**: Temporal

**Função**: Temporal Server (orquestração de workflows).

**Diagramas**:
- Temporal Server App Diagram

### 7.16. Worker Rate Limit

**Sistema**: DICT (aparece também em RSFN Connect App Diagram)
**Tecnologias**: Golang
**Descrição**: "repõe tokens conforme programado, consome eventos de Rate Limit"

**Componentes**:
- ratelimit.consumer
- ratelimit.refill
- Validação Rate Limit
- Producer Rate Limit

**Função**: Rate Limiting (Token Bucket algorithm).

---

## 8. Análise de Stores (Filas, Caches, DBs)

### 8.1. Cache Contas

**Tipo**: Cache (Redis)
**Função**: Cache de contas do Core Bancário.

### 8.2. Cache de Resposta

**Tipo**: Cache (Redis)
**Função**: Cache de respostas de requisições DICT (performance).

### 8.3. Cache Dedup

**Tipo**: Cache (Redis)
**Função**: Cache de deduplicação (evita requisições duplicadas).

### 8.4. Cache Rate Limit

**Tipo**: Cache (Redis)
**Função**: Cache de tokens para Rate Limiting (Token Bucket).

### 8.5. Cache Validacao Chave

**Tipo**: Cache (Redis)
**Função**: Cache de validações de chaves PIX.

### 8.6. CID e VSync ⭐⭐

**Tipo**: Database (PostgreSQL)
**Status**: live
**Função**: Armazena CID (Content ID) e resultados de VSYNC.

**Diagramas**:
- dict.vsync Component Diagram

**Análise Crítica**:
- ✅ **CID**: Identificadores de conteúdo locais (entries DICT)
- ✅ **VSYNC**: Resultados de sincronização diária com Bacen
- ✅ Usado por **dict.vsync** (VSYNCWorkflow)

### 8.7. db_statistics

**Tipo**: Database (PostgreSQL)
**Função**: Armazena estatísticas de fraude e segurança do DICT.

**Diagramas**:
- dict.statistics Component Diagram

### 8.8. locks/dict-orchestration-monitor

**Tipo**: Distributed Lock (Redis/Consul?)
**Função**: Locks distribuídos para dict.orchestration.monitor (evita execução duplicada de polling).

**Diagramas**:
- dict.orchestration.monitor Component Diagram

### 8.9. nome_da_fila_domain_events

**Tipo**: Message Queue (Apache Pulsar)
**Função**: Fila de eventos de domínio.

**Diagramas**:
- dict.api Component Diagram
- dict.vsync Component Diagram

### 8.10. nome_da_fila_rate_limit

**Tipo**: Message Queue (Apache Pulsar)
**Função**: Fila de eventos de Rate Limiting.

**Diagramas**:
- Worker Rate Limit Component Diagram

### 8.11. nome_da_fila_sync_contas

**Tipo**: Message Queue (Apache Pulsar)
**Função**: Fila de sincronização de contas do Core Bancário.

**Diagramas**:
- Sync de Contas Component Diagram

### 8.12. rsfn-dict-req-out ⭐⭐⭐

**Tipo**: Message Queue (Apache Pulsar)
**Status**: live
**Descrição**: "tópico para requisições"

**Tecnologias**: Apache Pulsar

**Diagramas**:
- DICT Proxy Component Diagram
- dict.api Component Diagram
- dict.vsync Component Diagram
- Requisição CRUD (flow)

**Fluxo**:
```
dict.api (Producer) → rsfn-dict-req-out → DICT Proxy (Consumer)
```

**Análise Crítica**:
- ✅ **Tópico principal de requisições** para RSFN
- ✅ dict.api **produz** mensagens
- ✅ DICT Proxy **consome** mensagens
- ✅ **Comunicação assíncrona** entre dict.api e Bridge

### 8.13. rsfn-dict-res-out ⭐⭐⭐

**Tipo**: Message Queue (Apache Pulsar)
**Status**: live
**Descrição**: "tópico para respostas"

**Tecnologias**: Apache Pulsar

**Diagramas**:
- DICT Proxy Component Diagram
- dict.api Component Diagram
- dict.vsync Component Diagram
- Requisição CRUD (flow)

**Fluxo**:
```
DICT Proxy (Producer) → rsfn-dict-res-out → dict.api (Consumer)
```

**Análise Crítica**:
- ✅ **Tópico principal de respostas** de RSFN
- ✅ DICT Proxy **produz** respostas após chamar Bacen
- ✅ dict.api **consome** respostas
- ✅ **Comunicação assíncrona** entre Bridge e dict.api

### 8.14. store_prometheus

**Tipo**: Time-Series Database (Prometheus)
**Função**: Armazena métricas coletadas pelo Open Telemetry Collector.

**Diagramas**:
- dict.dashboard Component Diagram
- Observabilidade App Diagram

---

## 9. Análise de Components

### 9.1. Autenticação e Autorização ⭐

**Sistema**: DICT
**Descrição**: "certifica que a requisição está com Autenticação e Autorização corretas. mTLS"
**Tecnologias**: Golang

**Diagramas**:
- dict.api Component Diagram
- Requisição CRUD (flow)

**Função**: Autenticação e autorização via mTLS para frontends.

### 9.2. Consumer DICT Proxy

**Sistema**: DICT Proxy
**Tipo**: Pulsar Consumer

**Função**: Consome mensagens de **rsfn-dict-req-out**.

**Diagramas**:
- DICT Proxy Component Diagram

### 9.3. Consumer Sync Contas

**Sistema**: Sync de Contas
**Tipo**: Pulsar Consumer

**Função**: Consome mensagens de **nome_da_fila_sync_contas**.

### 9.4. dic.vsync.domainevents.consumer

**Sistema**: dict.vsync
**Tipo**: Pulsar Consumer

**Função**: Consome eventos de domínio (**nome_da_fila_domain_events**) relacionados a VSYNC.

**Diagramas**:
- dict.vsync Component Diagram

### 9.5. DICT Proxy Application

**Sistema**: DICT Proxy
**Tecnologias**: Golang

**Função**: Lógica de aplicação do DICT Proxy (preparação SOAP, assinatura XML, envio mTLS).

**Diagramas**:
- DICT Proxy Component Diagram

### 9.6. dict.api.application ⭐

**Sistema**: dict.api
**Tecnologias**: Golang

**Função**: Lógica de aplicação do dict.api (orquestração, validações, transformações).

**Diagramas**:
- dict.api Component Diagram

### 9.7. dict.api.controller ⭐

**Sistema**: dict.api
**Tecnologias**: Golang

**Função**: Controladores gRPC/REST (endpoints da API).

**Diagramas**:
- dict.api Component Diagram
- Requisição CRUD (flow)

### 9.8. dict.api.model.validation

**Sistema**: dict.api
**Função**: Validação de modelos de dados.

**Diagramas**:
- dict.api Component Diagram

### 9.9. dict.api.validation

**Sistema**: dict.api
**Função**: Validações de negócio.

**Diagramas**:
- dict.api Component Diagram

### 9.10. dict.dashboard.api

**Sistema**: dict.dashboard
**Função**: Backend API do dashboard.

### 9.11. dict.dashboard.web

**Sistema**: dict.dashboard
**Função**: Frontend web do dashboard.

### 9.12. dict.domainevents.producer

**Sistema**: DICT
**Tipo**: Pulsar Producer

**Função**: Produz eventos de domínio para **nome_da_fila_domain_events**.

### 9.13. dict.rsfn.consumer ⭐⭐

**Sistema**: dict.api
**Tipo**: Pulsar Consumer

**Função**: Consome respostas de **rsfn-dict-res-out** (respostas do Bridge).

**Diagramas**:
- dict.api Component Diagram

### 9.14. dict.rsfn.producer ⭐⭐

**Sistema**: dict.api
**Tipo**: Pulsar Producer

**Função**: Produz requisições para **rsfn-dict-req-out** (requisições para Bridge).

**Diagramas**:
- dict.api Component Diagram

### 9.15. dict.statistics.collector

**Sistema**: dict.statistics
**Função**: Coleta estatísticas de fraude do Bacen.

### 9.16. dict.statistics.summarizer

**Sistema**: dict.statistics
**Função**: Sumariza e agrega estatísticas coletadas.

### 9.17. Producer DICT Proxy

**Sistema**: DICT Proxy
**Tipo**: Pulsar Producer

**Função**: Produz respostas para **rsfn-dict-res-out** (após chamar Bacen).

**Diagramas**:
- DICT Proxy Component Diagram

### 9.18. Producer Rate Limit

**Sistema**: Worker Rate Limit
**Tipo**: Pulsar Producer

**Função**: Produz eventos de Rate Limit para **nome_da_fila_rate_limit**.

### 9.19. ratelimit.consumer

**Sistema**: Worker Rate Limit
**Tipo**: Pulsar Consumer

**Função**: Consome eventos de Rate Limit.

### 9.20. ratelimit.refill

**Sistema**: Worker Rate Limit
**Função**: Refaz tokens do Token Bucket periodicamente.

### 9.21. Validação Rate Limit

**Sistema**: Worker Rate Limit
**Função**: Valida se requisição excede limite de taxa.

### 9.22. worker respostas orfãs

**Sistema**: (não especificado)
**Função**: Worker para processar respostas órfãs (timeout, retry).

### 9.23. worker.claims ⭐⭐⭐

**Sistema**: dict.orchestration.worker
**Descrição**: "activities relacionadas à reivindicação"
**Tecnologias**: Golang

**Função**: **Temporal Activities** para ClaimWorkflow (7 dias).

**Diagramas**:
- dict.orchestration.worker Component Diagram

**Análise Crítica**:
- ✅ **Activities de Reivindicação** (Claims)
- ✅ Executadas por **Temporal Workflows**
- ✅ Chamam DICT Proxy/RSFN Connect para comunicação com Bacen

### 9.24. worker.entries ⭐⭐⭐

**Sistema**: dict.orchestration.worker
**Descrição**: "activities relacionadas à vinculo de chaves"
**Tecnologias**: Golang

**Função**: **Temporal Activities** para criação/atualização de entries (chaves PIX).

**Diagramas**:
- dict.orchestration.worker Component Diagram

### 9.25. worker.vsync ⭐⭐

**Sistema**: dict.vsync / dict.orchestration.worker
**Descrição**: "activities relacionadas à sincronismo"

**Função**: **Temporal Activities** para VSYNCWorkflow (sincronização diária).

**Diagramas**:
- dict.vsync Component Diagram

### 9.26. workflow polling ⭐

**Sistema**: dict.orchestration.monitor
**Tecnologias**: Golang

**Função**: Polling de workflows Temporal (monitoramento de execução).

**Diagramas**:
- dict.orchestration.monitor Component Diagram

---

## 10. Mapeamento: IcePanel → TEC-002/TEC-003

### 10.1. Tabela de Mapeamento

| IcePanel | TEC-002 / TEC-003 | Tipo | Observação |
|----------|-------------------|------|------------|
| **DICT Proxy** (app) | **Bridge (TEC-002)** | App | Adapter SOAP/mTLS para Bacen |
| **dict.api** (app) | **Core Bancário DICT** | App | Ponto de entrada (gRPC/REST) |
| **dict.orchestration.worker** (app) | **Connect (TEC-003)** parcial | App | Contém Temporal Workers + Activities |
| **worker.claims** (component) | **ClaimWorkflow Activities** | Component | Activities para Claims (7 dias) |
| **worker.entries** (component) | **EntryWorkflow Activities** | Component | Activities para Entries |
| **worker.vsync** (component) | **VSYNCWorkflow Activities** | Component | Activities para VSYNC diário |
| **dict.vsync** (app) | **VSYNC Service** | App | Orquestra VSYNCWorkflow |
| **rsfn-dict-req-out** (store) | **bridge-dict-req-in** (TEC-002) | Pulsar Topic | Requisições para Bridge |
| **rsfn-dict-res-out** (store) | **bridge-dict-res-out** (TEC-002) | Pulsar Topic | Respostas do Bridge |
| **RSFN Connect** (system) | **Grupo/Contexto** | System | Grupo contendo DICT Proxy |
| **BC DICT** (system) | **Bacen DICT/SPI** | External System | Sistema externo do Bacen |
| **Temporal Server** (system) | **Temporal** | External System | Orquestrador de workflows |
| **CID e VSync** (store) | **PostgreSQL (Connect)** | Database | Estado de VSYNC e CID |

### 10.2. Análise de Inconsistências

#### Inconsistência 1: Nomenclatura "RSFN Connect"

**IcePanel**:
- **RSFN Connect** = Sistema/Grupo que contém DICT Proxy

**TEC-003**:
- **Connect** = Orquestrador com Temporal Workflows

**Resolução**:
- ✅ **DICT Proxy (IcePanel)** = **Bridge (TEC-002)**
- ⚠️ **"Connect" (TEC-003)** não existe explicitamente no IcePanel
- 🔍 **"Connect" (TEC-003)** parece ser **dict.api + dict.orchestration.worker** combinados

#### Inconsistência 2: Localização dos Temporal Workflows

**IcePanel**:
- Temporal Workflows estão em **dict.orchestration.worker** (dentro do sistema **DICT**)

**TEC-003**:
- Temporal Workflows estão em **Connect** (TEC-003)

**Resolução**:
- ✅ **dict.orchestration.worker** = **Connect (TEC-003)** em termos de responsabilidades
- ✅ Workflows estão **corretamente separados** do Bridge
- ✅ Nossa especificação TEC-003 está **alinhada** (workflows NÃO estão no Bridge)

#### Inconsistência 3: Tópicos Pulsar

**IcePanel**:
- **rsfn-dict-req-out** (requisições)
- **rsfn-dict-res-out** (respostas)

**TEC-002**:
- **bridge-dict-req-in** (requisições)
- **bridge-dict-res-out** (respostas)

**Resolução**:
- 🔄 **Renomear** tópicos em TEC-002 para corresponder ao IcePanel:
  - `bridge-dict-req-in` → `rsfn-dict-req-out`
  - `bridge-dict-res-out` → `rsfn-dict-res-out`

### 10.3. Fluxo Arquitetural Correto (Confirmado pelo IcePanel)

```
┌─────────────────────────────────────────────────────────────┐
│                   Frontend (gRPC/REST)                       │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│                    dict.api (Core DICT)                      │
│  - dict.api.controller                                       │
│  - dict.api.application                                      │
│  - dict.api.validation                                       │
│  - dict.rsfn.producer (Pulsar Producer)                     │
│  - dict.rsfn.consumer (Pulsar Consumer)                     │
└─────────────────────────────────────────────────────────────┘
                           ↓ Pulsar
┌─────────────────────────────────────────────────────────────┐
│            rsfn-dict-req-out (Pulsar Topic)                  │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│              DICT Proxy (Bridge TEC-002)                     │
│  - Consumer DICT Proxy (Pulsar Consumer)                    │
│  - DICT Proxy Application                                   │
│    - Prepare XML SOAP                                        │
│    - Sign XML (ICP-Brasil)                                   │
│    - Send mTLS to Bacen                                      │
│  - Producer DICT Proxy (Pulsar Producer)                    │
└─────────────────────────────────────────────────────────────┘
                           ↓ HTTPS mTLS
┌─────────────────────────────────────────────────────────────┐
│               BC DICT (Bacen - API DICT)                     │
│  - SOAP/XML endpoints                                        │
│  - mTLS authentication (ICP-Brasil)                          │
└─────────────────────────────────────────────────────────────┘
                           ↓ response
┌─────────────────────────────────────────────────────────────┐
│            rsfn-dict-res-out (Pulsar Topic)                  │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│                    dict.api (Core DICT)                      │
│  - dict.rsfn.consumer (recebe resposta)                     │
└─────────────────────────────────────────────────────────────┘
```

**Fluxo de Workflows Temporal** (Claims, VSYNC):

```
┌─────────────────────────────────────────────────────────────┐
│               Temporal Server (Temporal)                     │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│      dict.orchestration.worker (Connect TEC-003?)           │
│  - worker.claims (ClaimWorkflow Activities)                 │
│  - worker.entries (EntryWorkflow Activities)                │
│  - worker.vsync (VSYNCWorkflow Activities)                  │
└─────────────────────────────────────────────────────────────┘
                           ↓ chama
┌─────────────────────────────────────────────────────────────┐
│       rsfn-dict-req-out → DICT Proxy → Bacen                │
└─────────────────────────────────────────────────────────────┘
```

---

## 11. Conclusões e Recomendações

### 11.1. Principais Descobertas

1. ✅ **Arquitetura Confirmada**: O IcePanel confirma a separação entre:
   - **DICT Proxy** = Bridge (TEC-002) - Adapter SOAP/mTLS
   - **dict.orchestration.worker** = Connect (TEC-003) - Temporal Workflows

2. ✅ **Temporal Workflows Corretos**: Os workflows estão em **dict.orchestration.worker**, **NÃO** no DICT Proxy (Bridge).

3. ✅ **Pulsar Confirmado**: Comunicação assíncrona via Pulsar entre:
   - dict.api → rsfn-dict-req-out → DICT Proxy
   - DICT Proxy → rsfn-dict-res-out → dict.api

4. ✅ **Workers Temporal Identificados**:
   - **worker.claims** (ClaimWorkflow - 7 dias)
   - **worker.entries** (EntryWorkflow)
   - **worker.vsync** (VSYNCWorkflow - diário)

5. ⚠️ **Nomenclatura Inconsistente**:
   - "RSFN Connect" no IcePanel ≠ "Connect" em TEC-003
   - "RSFN Connect" parece ser apenas o **grupo/sistema** que contém o Bridge

### 11.2. Recomendações de Ajustes

#### Ajuste 1: Renomear Tópicos Pulsar em TEC-002/TEC-003

**Atual (TEC-002)**:
- `bridge-dict-req-in`
- `bridge-dict-res-out`

**Deve Ser (IcePanel)**:
- `rsfn-dict-req-out` (requisições)
- `rsfn-dict-res-out` (respostas)

**Ação**: Atualizar TEC-002 e TEC-003 com nomes corretos de tópicos Pulsar.

#### Ajuste 2: Clarificar Nomenclatura "Connect"

**Opção 1**: Renomear "Connect (TEC-003)" para **"dict.orchestration.worker"**
- Prós: Alinha com IcePanel
- Contras: Nome muito específico

**Opção 2**: Manter "Connect (TEC-003)" como **nome lógico/abstrato**
- Prós: Mais genérico, permite evolução
- Contras: Não corresponde 1:1 com IcePanel

**Recomendação**: Manter "Connect (TEC-003)" como nome lógico, mas **adicionar nota** explicando que corresponde a **dict.orchestration.worker + dict.api** no IcePanel.

#### Ajuste 3: Documentar Mapeamento IcePanel ↔ TEC-002/TEC-003

**Ação**: Criar seção em TEC-002 e TEC-003 com tabela de mapeamento:

```markdown
## Mapeamento: IcePanel → TEC-002/TEC-003

| IcePanel Component | TEC Spec | Observação |
|--------------------|----------|------------|
| DICT Proxy | Bridge (TEC-002) | Adapter SOAP/mTLS |
| dict.orchestration.worker | Connect (TEC-003) | Temporal Workers |
| rsfn-dict-req-out | Pulsar Topic (req) | Requisições |
| rsfn-dict-res-out | Pulsar Topic (res) | Respostas |
```

### 11.3. Validação Arquitetural Final

#### ✅ Fluxo Correto Confirmado

```
dict.api
  → (Pulsar: rsfn-dict-req-out)
  → DICT Proxy (Bridge TEC-002)
    → (SOAP/mTLS)
    → BC DICT (Bacen)
    → (response)
  → (Pulsar: rsfn-dict-res-out)
  → dict.api
```

#### ✅ Temporal Workflows Corretos

```
Temporal Server
  → dict.orchestration.worker (Connect TEC-003)
    → worker.claims (ClaimWorkflow Activities)
    → worker.entries (EntryWorkflow Activities)
    → worker.vsync (VSYNCWorkflow Activities)
  → chama DICT Proxy via Pulsar
  → chama Bacen via DICT Proxy
```

#### ✅ Separação de Responsabilidades Confirmada

| Componente | Responsabilidade | Status |
|------------|------------------|--------|
| **dict.api** | API principal, Pulsar Producer/Consumer | ✅ Confirmado |
| **DICT Proxy** | Adapter SOAP/mTLS para Bacen | ✅ Confirmado (TEC-002) |
| **dict.orchestration.worker** | Temporal Workers + Activities | ✅ Confirmado (TEC-003) |
| **Temporal Server** | Orquestração de workflows | ✅ Confirmado |

### 11.4. Próximos Passos

1. ✅ **Atualizar TEC-002 e TEC-003** com nomes corretos de tópicos Pulsar
2. ✅ **Adicionar seção de mapeamento** IcePanel ↔ TEC specs
3. 🔄 **Validar com time de arquitetura** a nomenclatura "Connect" vs "dict.orchestration.worker"
4. 🔄 **Revisar repositórios** para confirmar implementação atual vs IcePanel

---

## Anexos

### Anexo A: Estatísticas DICT (Detalhamento)

Conforme seção 8.8 do IcePanel, o DICT fornece as seguintes informações de segurança:

**Tipos de Consulta**:
1. **getEntryStatistics** (por chave PIX)
2. **getPersonStatistics** (por CPF/CNPJ)

**Informações Retornadas**:
- Quantidade de liquidações como recebedor no SPI (*Settlements*)
- Notificações de infração confirmadas por tipo de fraude:
  - **ApplicationFrauds** (Falsidade ideológica)
  - **MuleAccounts** (Conta laranja)
  - **ScammerAccounts** (Conta fraudador)
  - **OtherFrauds** (Outros)
- **UnknownFrauds**: Notificações sem identificação de tipo
- **TotalFraudTransactionAmount**: Valor total de notificações confirmadas
- **DistinctFraudReporters**: Participantes distintos reportando fraudes
- **OpenReports**: Notificações não fechadas
- **OpenReportsDistinctReporters**: Participantes com notificações abertas
- **RejectedReports**: Notificações rejeitadas
- **RegisteredAccounts**: Contas vinculadas a chaves PIX

**Períodos de Agregação**:
- Últimos 90 dias
- Últimos 12 meses (exceto mês atual)
- Últimos 60 meses (exceto mês atual)

**Watermark**: Atraso máximo de **12 horas** desde o último evento.

**Regras**:
- Informações **não são removidas** por modificações de chave (exclusão, portabilidade, reivindicação)
- Se chave é excluída e re-registrada com **mesmos dados**, herda informações da chave anterior

### Anexo B: Referências Cruzadas de Diagramas

| Componente | Diagramas Relacionados |
|------------|------------------------|
| **DICT Proxy** | DICT Proxy Component Diagram, RSFN Connect App Diagram, Requisição CRUD |
| **dict.api** | dict.api Component Diagram, DICT App Diagram, Requisição CRUD |
| **dict.orchestration.worker** | dict.orchestration.worker Component Diagram, DICT App Diagram |
| **dict.vsync** | dict.vsync Component Diagram, DICT App Diagram |
| **rsfn-dict-req-out** | DICT Proxy Component Diagram, dict.api Component Diagram, Requisição CRUD |
| **rsfn-dict-res-out** | DICT Proxy Component Diagram, dict.api Component Diagram, Requisição CRUD |

---

**Documento Gerado**: 2025-10-25
**Próxima Revisão**: Após validação com time de arquitetura e atualização de TEC-002/TEC-003
