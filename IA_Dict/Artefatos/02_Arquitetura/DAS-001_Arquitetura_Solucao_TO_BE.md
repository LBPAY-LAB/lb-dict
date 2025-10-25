# DAS-001 - Documento de Arquitetura de Solução TO-BE

**Agente Responsável**: NEXUS (AGT-ARC-001) - Solution Architect
**Data de Criação**: 2025-10-24
**Versão**: 1.0
**Status**: Em Elaboração

---

## 📋 Índice

1. [Informações Gerais](#1-informações-gerais)
2. [Contexto e Drivers Arquiteturais](#2-contexto-e-drivers-arquiteturais)
3. [Visão Geral da Solução TO-BE](#3-visão-geral-da-solução-to-be)
4. [Arquitetura C4 - Nível 1: Contexto](#4-arquitetura-c4---nível-1-contexto)
5. [Arquitetura C4 - Nível 2: Containers](#5-arquitetura-c4---nível-2-containers)
6. [Arquitetura C4 - Nível 3: Componentes](#6-arquitetura-c4---nível-3-componentes)
7. [Stack Tecnológica](#7-stack-tecnológica)
8. [Fluxos de Dados Principais](#8-fluxos-de-dados-principais)
9. [Estratégia de Performance](#9-estratégia-de-performance)
10. [Estratégia de Resiliência](#10-estratégia-de-resiliência)
11. [Segurança](#11-segurança)
12. [Observabilidade](#12-observabilidade)
13. [Estratégia de Deployment](#13-estratégia-de-deployment)
14. [Decisões Arquiteturais (ADRs)](#14-decisões-arquiteturais-adrs)
15. [Roadmap de Implementação](#15-roadmap-de-implementação)
16. [Referências](#16-referências)

---

## 1. Informações Gerais

### 1.1 Objetivo do Documento

Este documento define a **arquitetura TO-BE** (target) do sistema DICT LBPay, consolidando toda a lógica de negócio DICT em um único repositório (**core-dict**) e estabelecendo padrões claros de integração com o Bacen DICT.

### 1.2 Contexto

**Situação Atual (AS-IS)**:
- ❌ Lógica DICT dispersa em múltiplos repositórios (money-moving, orchestration-go, operation)
- ❌ Duplicação de código
- ❌ Difícil manutenção e evolução
- ❌ Sem implementação completa dos 72 RFs do DICT

**Situação Desejada (TO-BE)**:
- ✅ **Um único repositório** para Core DICT (lógica de negócio)
- ✅ Implementação completa dos 72 Requisitos Funcionais
- ✅ Clean Architecture bem definida
- ✅ Performance para alto volume (dezenas de queries/segundo)
- ✅ Resiliência e observabilidade

### 1.3 Princípios Arquiteturais

1. **Single Responsibility**: Cada componente tem uma responsabilidade clara
2. **Separation of Concerns**: Core DICT (negócio) separado de Connect DICT (infraestrutura)
3. **Clean Architecture**: Domain, Application, Handlers, Infrastructure
4. **Performance First**: Cache multi-camadas, connection pooling
5. **Resilience**: Circuit breakers, retries, timeouts
6. **Observability**: Metrics, traces, logs estruturados
7. **Security by Design**: mTLS, assinatura digital, rate limiting
8. **Cloud Native**: Containerizado, orquestrado (Kubernetes)

---

## 2. Contexto e Drivers Arquiteturais

### 2.1 Requisitos Funcionais

**72 Requisitos Funcionais** mapeados no [CRF-001](../05_Requisitos/CRF-001_Checklist_Requisitos_Funcionais.md):

| Bloco | RFs | Prioridade | Status Atual |
|-------|-----|------------|--------------|
| **Bloco 1 - CRUD de Chaves** | 13 | Must Have | 15.4% implementado |
| **Bloco 2 - Reivindicação/Portabilidade** | 14 | Should Have | 0% implementado |
| **Bloco 3 - Validação** | 3 | Must Have | 33.3% implementado |
| **Bloco 4 - Devolução/Infração** | 6 | Should Have | 0% implementado |
| **Bloco 5 - Segurança** | 13 | Should Have | 23.1% implementado |
| **Bloco 6 - Recuperação de Valores** | 13 | Nice to Have | 0% implementado |
| **Transversal** | 10 | Variado | 10% implementado |

**Gap total**: **91.6%** dos RFs não implementados (66 de 72).

### 2.2 Requisitos Não-Funcionais

#### 2.2.1 Performance
- **RNF-001**: Latência P99 < 1s para consultas (GET /entries/{Key})
- **RNF-002**: Throughput de **dezenas de queries/segundo** (requisito crítico do usuário)
- **RNF-003**: Cache hit ratio > 70% para consultas recorrentes
- **RNF-004**: Connection pool reutilização > 90%

#### 2.2.2 Disponibilidade
- **RNF-005**: SLA 99.9% (downtime < 43min/mês)
- **RNF-006**: Zero-downtime deployments
- **RNF-007**: Recovery Time Objective (RTO) < 5min

#### 2.2.3 Escalabilidade
- **RNF-008**: Horizontal scaling (Kubernetes HPA)
- **RNF-009**: Suportar crescimento de 10x em volume de chaves
- **RNF-010**: Auto-scaling baseado em métricas (CPU, latência, throughput)

#### 2.2.4 Segurança
- **RNF-011**: mTLS obrigatório (Bacen DICT)
- **RNF-012**: Assinatura digital XML (XML Signature)
- **RNF-013**: Rate limiting local (prevenir 429 do Bacen)
- **RNF-014**: Dados sensíveis criptografados at-rest (PostgreSQL TDE)
- **RNF-015**: Logs auditáveis (LGPD compliance)

#### 2.2.5 Observabilidade
- **RNF-016**: Traces distribuídos (OpenTelemetry)
- **RNF-017**: Métricas RED (Rate, Errors, Duration) + USE (Utilization, Saturation, Errors)
- **RNF-018**: Logs estruturados (JSON) com trace correlation
- **RNF-019**: Alertas proativos (latência, erros, rate limiting)

### 2.3 Restrições

1. **Tecnologia**: Golang 1.24.5+ (padrão LBPay)
2. **Infraestrutura**: Kubernetes (GKE, EKS ou AKS)
3. **Banco de Dados**: PostgreSQL 14+ (já em uso)
4. **Cache**: Redis 7+ (já em uso)
5. **Message Broker**: Apache Pulsar (já em uso)
6. **Workflow Engine**: Temporal (já em uso)
7. **Compliance**: Regulamento PIX Bacen, LGPD

---

## 3. Visão Geral da Solução TO-BE

### 3.1 Arquitetura de Alto Nível

```
┌────────────────────────────────────────────────────────────────────────────┐
│                          EXTERNAL SYSTEMS                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ DICT Bacen   │  │  SPI Bacen   │  │ Receita Fed  │  │ Frontend App │  │
│  │  (API v2.6)  │  │              │  │  (CPF/CNPJ)  │  │  (out scope) │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘  │
└────────────────────────────────────────────────────────────────────────────┘
         ↑                    ↑                ↑                  ↑
         │ mTLS/REST          │ mTLS/ISO       │ HTTPS            │
         │ XML Signed         │ 20022          │                  │
         ↓                    ↓                ↓                  ↓
┌────────────────────────────────────────────────────────────────────────────┐
│                         LBPay DICT SYSTEM (TO-BE)                          │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────┐    │
│  │                     API GATEWAY (Kong/NGINX)                       │    │
│  │  • Authentication (JWT)                                            │    │
│  │  • Rate Limiting (external APIs)                                   │    │
│  │  • Load Balancing                                                  │    │
│  └───────────────────────────────────────────────────────────────────┘    │
│                                  ↓                                          │
│  ┌───────────────────────────────────────────────────────────────────┐    │
│  │                     CORE DICT (novo repositório)                   │    │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐  │    │
│  │  │  gRPC API  │  │  REST API  │  │ Pulsar Sub │  │ Temporal   │  │    │
│  │  │  (interno) │  │ (externo)  │  │ (eventos)  │  │ (workflows)│  │    │
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘  │    │
│  │         ↓              ↓                ↓              ↓          │    │
│  │  ┌──────────────────────────────────────────────────────────┐    │    │
│  │  │              APPLICATION LAYER (Use Cases)                │    │    │
│  │  │  • CreatePixKey     • GetPixKey      • DeletePixKey      │    │    │
│  │  │  • UpdatePixKey     • CreateClaim    • ConfirmClaim      │    │    │
│  │  │  • CreateRefund     • GetStatistics  • SyncVerification  │    │    │
│  │  └──────────────────────────────────────────────────────────┘    │    │
│  │         ↓                                                         │    │
│  │  ┌──────────────────────────────────────────────────────────┐    │    │
│  │  │                DOMAIN LAYER (Business Logic)              │    │    │
│  │  │  • PixKey      • Claim       • InfractionReport          │    │    │
│  │  │  • Refund      • FraudMarker • Statistics                │    │    │
│  │  │  • Validators  • Policies    • Rules                     │    │    │
│  │  └──────────────────────────────────────────────────────────┘    │    │
│  │         ↓                                                         │    │
│  │  ┌──────────────────────────────────────────────────────────┐    │    │
│  │  │            INFRASTRUCTURE LAYER (Adapters)                │    │    │
│  │  │  • PostgreSQL Repos  • Redis Caches  • Pulsar Publisher  │    │    │
│  │  │  • Temporal Client   • Connect DICT  • Validators        │    │    │
│  │  └──────────────────────────────────────────────────────────┘    │    │
│  └───────────────────────────────────────────────────────────────────┘    │
│                                  ↓                                          │
│  ┌───────────────────────────────────────────────────────────────────┐    │
│  │              CONNECT DICT (rsfn-connect-bacen-bridge)              │    │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐  │    │
│  │  │ REST Client│  │ XML Signer │  │ mTLS Setup │  │ Rate Limiter│  │    │
│  │  │ (pool)     │  │ (P12 cert) │  │            │  │ (local)     │  │    │
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘  │    │
│  │  ┌─────────────────────────────────────────────────────────────┐  │    │
│  │  │          28 API Clients (entries, claims, refunds, ...)     │  │    │
│  │  └─────────────────────────────────────────────────────────────┘  │    │
│  └───────────────────────────────────────────────────────────────────┘    │
│                                  ↓                                          │
│  ┌───────────────────────────────────────────────────────────────────┐    │
│  │                     SHARED INFRASTRUCTURE                          │    │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐  │    │
│  │  │ PostgreSQL │  │  5x Redis  │  │   Pulsar   │  │  Temporal  │  │    │
│  │  │  (stores)  │  │  (caches)  │  │  (async)   │  │ (workflows)│  │    │
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘  │    │
│  └───────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────┐    │
│  │                    OBSERVABILITY STACK                             │    │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐  │    │
│  │  │ Prometheus │  │   Jaeger   │  │    Loki    │  │  Grafana   │  │    │
│  │  │ (metrics)  │  │  (traces)  │  │   (logs)   │  │ (dashboards)│  │    │
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘  │    │
│  └───────────────────────────────────────────────────────────────────┘    │
└────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Principais Componentes

| Componente | Responsabilidade | Repositório | Linguagem | Status |
|------------|------------------|-------------|-----------|--------|
| **Core DICT** | Lógica de negócio DICT (72 RFs) | `core-dict` (novo) | Go 1.24.5+ | ⚠️ A criar |
| **Connect DICT** | Cliente REST para DICT Bacen | `rsfn-connect-bacen-bridge` | Go 1.24.5+ | ⚠️ Refatorar |
| **Validator Service** | Validação CPF/CNPJ (Receita Federal) | `sdk-rsfn-validator` | Go | ✅ Existente |
| **Temporal Workflows** | Orquestração de processos longos (Claims) | `core-dict/workflows` | Go | ⚠️ A criar |
| **PostgreSQL** | Persistência (CID, VSync, Statistics) | - | - | ✅ Existente |
| **Redis (5 caches)** | Performance multi-camadas | - | - | ✅ Existente |
| **Apache Pulsar** | Mensageria assíncrona | - | - | ✅ Existente |

---

## 4. Arquitetura C4 - Nível 1: Contexto

### 4.1 Diagrama de Contexto

```
                    ┌──────────────────────────────┐
                    │                              │
                    │      Usuário Final LBPay     │
                    │   (Pessoa Física/Jurídica)   │
                    │                              │
                    └──────────────┬───────────────┘
                                   │
                                   │ Cria/Consulta/
                                   │ Exclui chaves PIX
                                   ↓
┌───────────────────────────────────────────────────────────────────┐
│                                                                   │
│                  LBPay Frontend Application                       │
│                     (out of scope)                                │
│                                                                   │
└───────────────────────────────┬───────────────────────────────────┘
                                │ HTTPS/REST
                                │ JWT Auth
                                ↓
        ┌───────────────────────────────────────────────┐
        │                                               │
        │         LBPay DICT System (TO-BE)             │
        │                                               │
        │  • Gerencia chaves PIX                        │
        │  • Reivindicações/Portabilidade               │
        │  • Devoluções/Infrações                       │
        │  • Estatísticas antifraude                    │
        │                                               │
        └───────────────────────────────────────────────┘
                │              │              │
                │              │              │
     ┌──────────┘              │              └──────────┐
     │                         │                         │
     ↓ mTLS/REST               ↓ HTTPS                   ↓ mTLS/ISO20022
     │ XML Signed              │                         │
┌────────────────┐      ┌──────────────┐      ┌──────────────────┐
│                │      │              │      │                  │
│  DICT Bacen    │      │ Receita Fed  │      │    SPI Bacen     │
│   (API v2.6)   │      │  (CPF/CNPJ)  │      │  (Liquidação)    │
│                │      │              │      │                  │
└────────────────┘      └──────────────┘      └──────────────────┘
```

### 4.2 Atores Externos

| Ator | Descrição | Protocolo | Autenticação |
|------|-----------|-----------|--------------|
| **Usuário Final** | PF/PJ que cria/gerencia chaves PIX | HTTPS/REST | JWT (via frontend) |
| **DICT Bacen** | Diretório centralizado de chaves PIX | mTLS/REST/XML | Certificado X.509 |
| **Receita Federal** | Validação de CPF/CNPJ | HTTPS/REST | API Key |
| **SPI Bacen** | Sistema de Pagamentos Instantâneos | mTLS/ISO20022 | Certificado X.509 |

---

## 5. Arquitetura C4 - Nível 2: Containers

### 5.1 Diagrama de Containers

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         LBPay DICT System                                 │
│                                                                           │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                        API Gateway                               │    │
│  │  Kong API Gateway / NGINX Ingress Controller                     │    │
│  │  • Authentication (JWT)                                          │    │
│  │  • Rate Limiting (external)                                      │    │
│  │  • Load Balancing                                                │    │
│  └──────────────────────────┬───────────────────────────────────────┘    │
│                             │                                             │
│  ┌──────────────────────────┴───────────────────────────────────────┐    │
│  │                    Core DICT Service                              │    │
│  │  Container: core-dict                                             │    │
│  │  Technology: Go 1.24.5+, gRPC, REST                               │    │
│  │                                                                   │    │
│  │  Responsibilities:                                                │    │
│  │  • Business logic (72 RFs)                                        │    │
│  │  • CRUD de chaves PIX                                             │    │
│  │  • Reivindicações/Portabilidade                                   │    │
│  │  • Devoluções/Infrações                                           │    │
│  │  • Estatísticas antifraude                                        │    │
│  │  • Reconciliação (VSync, CID)                                     │    │
│  │                                                                   │    │
│  │  Ports:                                                           │    │
│  │  • 8080 (REST API external)                                       │    │
│  │  • 9090 (gRPC API internal)                                       │    │
│  │  • 8081 (health/metrics)                                          │    │
│  └───────────────────────────────────────────────────────────────────┘    │
│                             │                                             │
│  ┌──────────────────────────┴───────────────────────────────────────┐    │
│  │                  Connect DICT Bridge                              │    │
│  │  Container: rsfn-connect-bacen-bridge                             │    │
│  │  Technology: Go 1.24.5+, REST Client                              │    │
│  │                                                                   │    │
│  │  Responsibilities:                                                │    │
│  │  • REST client para DICT Bacen (28 endpoints)                     │    │
│  │  • mTLS configuration                                             │    │
│  │  • XML signing/validation                                         │    │
│  │  • Connection pooling                                             │    │
│  │  • Rate limiting local                                            │    │
│  │  • Retry logic                                                    │    │
│  │                                                                   │    │
│  │  Ports:                                                           │    │
│  │  • 9091 (gRPC API - called by Core DICT)                          │    │
│  └───────────────────────────────────────────────────────────────────┘    │
│                                                                           │
│  ┌───────────────────────────────────────────────────────────────────┐    │
│  │                    Temporal Worker                                 │    │
│  │  Container: core-dict-temporal-worker                              │    │
│  │  Technology: Go 1.24.5+, Temporal SDK                              │    │
│  │                                                                   │    │
│  │  Responsibilities:                                                │    │
│  │  • Claim workflows (7-day process)                                │    │
│  │  • Validation workflows (SMS/Email)                               │    │
│  │  • Reconciliation workflows                                       │    │
│  └───────────────────────────────────────────────────────────────────┘    │
│                                                                           │
│  ┌───────────────────────────────────────────────────────────────────┐    │
│  │                     Pulsar Consumer                                │    │
│  │  Container: core-dict-pulsar-consumer                              │    │
│  │  Technology: Go 1.24.5+, Pulsar Client                             │    │
│  │                                                                   │    │
│  │  Responsibilities:                                                │    │
│  │  • Consume CID events (reconciliation)                            │    │
│  │  • Consume SPI events (liquidation → refill rate limit)           │    │
│  │  • Consume fraud events                                           │    │
│  └───────────────────────────────────────────────────────────────────┘    │
│                                                                           │
└──────────────────────────────────────────────────────────────────────────┘
                   │              │              │              │
        ┌──────────┘              │              │              └──────────┐
        │                         │              │                         │
        ↓                         ↓              ↓                         ↓
┌────────────────┐      ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│                │      │              │  │              │  │              │
│  PostgreSQL    │      │ Redis Cluster│  │Apache Pulsar │  │   Temporal   │
│                │      │  (5 caches)  │  │              │  │   Server     │
│  • CID Store   │      │              │  │ • Topic:     │  │              │
│  • VSync Store │      │ • Response   │  │   dict-cids  │  │ • Workflows  │
│  • Statistics  │      │ • Account    │  │ • Topic:     │  │ • Activities │
│  • PIX Keys    │      │ • Validation │  │   dict-events│  │              │
│                │      │ • Dedup      │  │              │  │              │
│                │      │ • RateLimit  │  │              │  │              │
└────────────────┘      └──────────────┘  └──────────────┘  └──────────────┘
    Port: 5432             Ports:            Port: 6650      Port: 7233
                           7001-7005
```

### 5.2 Responsabilidades dos Containers

#### 5.2.1 Core DICT Service
**Propósito**: Serviço principal com toda lógica de negócio DICT.

**Responsabilidades**:
- ✅ Implementar 72 Requisitos Funcionais
- ✅ Validação de regras de negócio
- ✅ Orquestração de use cases
- ✅ Gerenciamento de transações
- ✅ Publicação de eventos (Pulsar)
- ✅ Iniciar workflows (Temporal)

**APIs**:
- **REST** (porta 8080): Para clientes externos (Frontend, Orchestration)
- **gRPC** (porta 9090): Para clientes internos (outros microservices)

**Tecnologias**:
- Go 1.24.5+
- gRPC-Go
- Chi/Gin (REST framework)
- GORM (ORM)
- Pulsar Go Client
- Temporal Go SDK

#### 5.2.2 Connect DICT Bridge
**Propósito**: Adaptador de infraestrutura para comunicação com DICT Bacen.

**Responsabilidades**:
- ✅ Cliente REST para 28 endpoints DICT Bacen
- ✅ Configuração de mTLS (certificado X.509)
- ✅ Assinatura digital XML (envelopada)
- ✅ Validação de assinaturas de resposta
- ✅ Connection pooling (keep-alive)
- ✅ Rate limiting local (prevenir 429)
- ✅ Retry com exponential backoff
- ✅ Circuit breaker

**API**:
- **gRPC** (porta 9091): Chamado pelo Core DICT

**Tecnologias**:
- Go 1.24.5+
- gRPC-Go
- net/http (com mTLS)
- XML encoding/decoding
- Redis (rate limiting local)

#### 5.2.3 Temporal Worker
**Propósito**: Executar workflows de longa duração.

**Workflows**:
1. **ClaimWorkflow**: Reivindicação/Portabilidade (7 dias)
2. **ValidationWorkflow**: Validação de posse (SMS/Email)
3. **ReconciliationWorkflow**: Sincronização periódica (VSync)

**Tecnologias**:
- Go 1.24.5+
- Temporal Go SDK

#### 5.2.4 Pulsar Consumer
**Propósito**: Processar eventos assíncronos.

**Topics consumidos**:
- `dict-cids`: Eventos de CID (criado, atualizado, deletado)
- `dict-events`: Eventos gerais (claim confirmado, refund criado)
- `spi-liquidation`: Eventos de liquidação SPI (repõe fichas de rate limit)

**Tecnologias**:
- Go 1.24.5+
- Pulsar Go Client

### 5.3 Data Stores

#### 5.3.1 PostgreSQL (Porta 5432)
**Databases**:
- `dict_db`: Database principal

**Schemas**:
- `pix_keys`: Tabela de chaves PIX
- `claims`: Tabela de reivindicações
- `infraction_reports`: Tabela de notificações de infração
- `refunds`: Tabela de devoluções
- `fraud_markers`: Tabela de marcações de fraude
- `cids`: Tabela de CIDs (reconciliação)
- `vsync`: Tabela de VSync por tipo de chave
- `statistics`: Tabela de estatísticas agregadas

**Backup**: Diário (retenção 30 dias)

#### 5.3.2 Redis Cluster (5 instâncias)

| Porta | Cache | Propósito | TTL |
|-------|-------|-----------|-----|
| 7001 | `cache-dict-response` | Respostas completas de consultas DICT | 5min |
| 7002 | `cache-dict-account` | Dados de contas transacionais | 15min |
| 7003 | `cache-dict-key-validation` | Validações de chaves PIX | 10min |
| 7004 | `cache-dict-dedup` | Deduplicação de requisições (RequestId) | 1min |
| 7005 | `cache-dict-rate-limit` | Controle local de rate limiting | Variável |

**Eviction Policy**: `allkeys-lru` (Least Recently Used)

#### 5.3.3 Apache Pulsar (Porta 6650)

**Topics**:
- `persistent://lbpay/dict/cids`: Eventos de CID
- `persistent://lbpay/dict/events`: Eventos gerais DICT
- `persistent://lbpay/spi/liquidation`: Eventos de liquidação SPI

**Retention**: 7 dias

#### 5.3.4 Temporal Server (Porta 7233)

**Namespaces**:
- `lbpay-dict-prod`: Produção
- `lbpay-dict-hml`: Homologação

---

## 6. Arquitetura C4 - Nível 3: Componentes

### 6.1 Core DICT - Component Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Core DICT Service                               │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │                     HANDLERS LAYER                              │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │    │
│  │  │  REST API    │  │  gRPC API    │  │ Pulsar Sub   │         │    │
│  │  │  Handler     │  │  Handler     │  │  Handler     │         │    │
│  │  │              │  │              │  │              │         │    │
│  │  │ • Routes     │  │ • Services   │  │ • Consumers  │         │    │
│  │  │ • Middleware │  │ • Interceptors│ │              │         │    │
│  │  └──────────────┘  └──────────────┘  └──────────────┘         │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                             ↓                                            │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │                   APPLICATION LAYER                             │    │
│  │  ┌─────────────────────────────────────────────────────────┐   │    │
│  │  │                    Use Cases                             │   │    │
│  │  │                                                          │   │    │
│  │  │  Bloco 1 - CRUD:                                        │   │    │
│  │  │  • CreatePixKeyUseCase                                  │   │    │
│  │  │  • GetPixKeyUseCase                                     │   │    │
│  │  │  • UpdatePixKeyUseCase                                  │   │    │
│  │  │  • DeletePixKeyUseCase                                  │   │    │
│  │  │  • ValidatePixKeyUseCase                                │   │    │
│  │  │                                                          │   │    │
│  │  │  Bloco 2 - Claim:                                       │   │    │
│  │  │  • CreateClaimUseCase                                   │   │    │
│  │  │  • AcknowledgeClaimUseCase                              │   │    │
│  │  │  • ConfirmClaimUseCase                                  │   │    │
│  │  │  • CancelClaimUseCase                                   │   │    │
│  │  │  • CompleteClaimUseCase                                 │   │    │
│  │  │  • ListClaimsUseCase                                    │   │    │
│  │  │                                                          │   │    │
│  │  │  Bloco 3 - Validation:                                  │   │    │
│  │  │  • ValidatePossessionUseCase (SMS/Email)                │   │    │
│  │  │  • ValidateCPFCNPJUseCase (Receita Federal)             │   │    │
│  │  │                                                          │   │    │
│  │  │  Bloco 4 - Refund/Infraction:                           │   │    │
│  │  │  • CreateInfractionReportUseCase                        │   │    │
│  │  │  • CreateRefundUseCase                                  │   │    │
│  │  │  • CloseRefundUseCase                                   │   │    │
│  │  │                                                          │   │    │
│  │  │  Bloco 5 - Security:                                    │   │    │
│  │  │  • GetEntryStatisticsUseCase                            │   │    │
│  │  │  • CreateFraudMarkerUseCase                             │   │    │
│  │  │  • SyncVerificationUseCase                              │   │    │
│  │  │                                                          │   │    │
│  │  └─────────────────────────────────────────────────────────┘   │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                             ↓                                            │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │                      DOMAIN LAYER                               │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │    │
│  │  │  Aggregates  │  │  Entities    │  │ Value Objects│         │    │
│  │  │              │  │              │  │              │         │    │
│  │  │ • PixKey     │  │ • Claim      │  │ • Key        │         │    │
│  │  │ • Account    │  │ • Refund     │  │ • Owner      │         │    │
│  │  │              │  │ • Infraction │  │ • Participant│         │    │
│  │  └──────────────┘  └──────────────┘  └──────────────┘         │    │
│  │                                                                 │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │    │
│  │  │  Validators  │  │   Policies   │  │    Rules     │         │    │
│  │  │              │  │              │  │              │         │    │
│  │  │ • KeyVal     │  │ • RateLimit  │  │ • LimitRule  │         │    │
│  │  │ • OwnerVal   │  │ • Retry      │  │ • ClaimRule  │         │    │
│  │  │ • AccountVal │  │              │  │              │         │    │
│  │  └──────────────┘  └──────────────┘  └──────────────┘         │    │
│  │                                                                 │    │
│  │  ┌──────────────────────────────────────────────────────┐     │    │
│  │  │               Domain Services                         │     │    │
│  │  │  • CIDCalculator                                     │     │    │
│  │  │  • VSyncCalculator                                   │     │    │
│  │  │  • ClaimOrchestrator                                 │     │    │
│  │  └──────────────────────────────────────────────────────┘     │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                             ↓                                            │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │                 INFRASTRUCTURE LAYER                            │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │    │
│  │  │ Repositories │  │    Caches    │  │   Clients    │         │    │
│  │  │              │  │              │  │              │         │    │
│  │  │ • PixKeyRepo │  │ • ResponseC  │  │ • ConnectDICT│         │    │
│  │  │ • ClaimRepo  │  │ • AccountC   │  │ • Validator  │         │    │
│  │  │ • RefundRepo │  │ • ValidationC│  │ • Temporal   │         │    │
│  │  │ • InfractionR│  │ • DedupC     │  │ • Pulsar     │         │    │
│  │  │ • CIDRepo    │  │ • RateLimitC │  │              │         │    │
│  │  └──────────────┘  └──────────────┘  └──────────────┘         │    │
│  └────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
```

### 6.2 Componentes Principais

#### 6.2.1 Handlers Layer

**REST API Handler** (porta 8080):
```go
// pkg/handlers/rest/
├── router.go           // Chi/Gin router setup
├── middleware/
│   ├── auth.go         // JWT authentication
│   ├── logging.go      // Request logging
│   ├── tracing.go      // OpenTelemetry
│   └── ratelimit.go    // External rate limiting
├── v1/
│   ├── pixkey.go       // CRUD endpoints
│   ├── claim.go        // Claim endpoints
│   ├── refund.go       // Refund endpoints
│   └── statistics.go   // Statistics endpoints
```

**gRPC API Handler** (porta 9090):
```go
// pkg/handlers/grpc/
├── server.go           // gRPC server setup
├── interceptors/
│   ├── auth.go         // mTLS validation
│   ├── logging.go      // Call logging
│   └── tracing.go      // OpenTelemetry
├── services/
│   ├── pixkey_service.go
│   ├── claim_service.go
│   └── refund_service.go
```

**Pulsar Subscriber Handler**:
```go
// pkg/handlers/pulsar/
├── consumer.go         // Pulsar consumer setup
├── handlers/
│   ├── cid_events.go   // CID event handler
│   ├── dict_events.go  // General DICT events
│   └── spi_events.go   // SPI liquidation events
```

#### 6.2.2 Application Layer

**Use Cases** (72 RFs implementados):
```go
// pkg/application/usecases/
├── pixkey/
│   ├── create_pixkey.go
│   ├── get_pixkey.go
│   ├── update_pixkey.go
│   ├── delete_pixkey.go
│   └── validate_pixkey.go
├── claim/
│   ├── create_claim.go
│   ├── acknowledge_claim.go
│   ├── confirm_claim.go
│   ├── cancel_claim.go
│   ├── complete_claim.go
│   └── list_claims.go
├── validation/
│   ├── validate_possession.go  // SMS/Email
│   └── validate_cpf_cnpj.go    // Receita Federal
├── refund/
│   ├── create_refund.go
│   ├── close_refund.go
│   └── cancel_refund.go
├── infraction/
│   ├── create_infraction.go
│   └── close_infraction.go
├── security/
│   ├── get_entry_statistics.go
│   ├── get_person_statistics.go
│   ├── create_fraud_marker.go
│   └── sync_verification.go
```

Cada Use Case segue o padrão:
```go
type CreatePixKeyUseCase struct {
    repo       domain.PixKeyRepository
    dictClient infrastructure.ConnectDICTClient
    cache      infrastructure.CacheRepository
    publisher  infrastructure.EventPublisher
    validator  domain.PixKeyValidator
}

func (uc *CreatePixKeyUseCase) Execute(ctx context.Context, input CreatePixKeyInput) (*CreatePixKeyOutput, error) {
    // 1. Validação de input
    // 2. Validação de regras de negócio (domain)
    // 3. Chamada ao DICT Bacen (via Connect DICT)
    // 4. Persistência (PostgreSQL)
    // 5. Atualização de cache
    // 6. Publicação de evento (Pulsar)
    // 7. Retorno
}
```

#### 6.2.3 Domain Layer

**Aggregates**:
```go
// pkg/domain/aggregates/
├── pixkey.go           // PixKey aggregate root
├── account.go          // Account aggregate
├── claim.go            // Claim aggregate
└── refund.go           // Refund aggregate
```

Exemplo de Aggregate:
```go
type PixKey struct {
    Key           Key           // Value object
    KeyType       KeyType       // Enum
    Account       Account       // Entity
    Owner         Owner         // Value object
    Status        KeyStatus     // Enum
    CreatedAt     time.Time
    UpdatedAt     time.Time
    CID           CID           // Value object (256-bit hex)
}

func (pk *PixKey) Validate() error {
    // Business rules validation
}

func (pk *PixKey) CanBeDeleted() bool {
    // Business rule: Cannot delete if locked by claim
}

func (pk *PixKey) CalculateCID(requestId RequestID) CID {
    // CID calculation algorithm
}
```

**Value Objects**:
```go
// pkg/domain/valueobjects/
├── key.go              // PIX key (CPF, CNPJ, EMAIL, PHONE, EVP)
├── owner.go            // Owner (Name, TaxIdNumber, Type)
├── participant.go      // Participant (ISPB)
├── cid.go              // Content Identifier (256-bit)
├── vsync.go            // VSync (256-bit XOR of CIDs)
```

**Domain Services**:
```go
// pkg/domain/services/
├── cid_calculator.go   // CID = HMAC-SHA256(requestId, entryAttributes)
├── vsync_calculator.go // VSync = XOR(cid1, cid2, ..., cidN)
├── claim_orchestrator.go // Orchestrates claim state machine
```

**Validators**:
```go
// pkg/domain/validators/
├── key_validator.go    // Validates PIX key format (regex)
├── owner_validator.go  // Validates owner data
├── account_validator.go// Validates account data
├── limit_validator.go  // Validates 5/20 key limits
```

**Policies**:
```go
// pkg/domain/policies/
├── rate_limit_policy.go // Rate limiting rules
├── retry_policy.go      // Retry logic (exponential backoff)
├── cache_policy.go      // Cache TTL rules
```

#### 6.2.4 Infrastructure Layer

**Repositories** (PostgreSQL):
```go
// pkg/infrastructure/persistence/
├── pixkey_repository.go
├── claim_repository.go
├── refund_repository.go
├── infraction_repository.go
├── cid_repository.go
└── vsync_repository.go
```

Interface example:
```go
type PixKeyRepository interface {
    Save(ctx context.Context, key *domain.PixKey) error
    FindByKey(ctx context.Context, key string) (*domain.PixKey, error)
    Delete(ctx context.Context, key string) error
    CountByAccount(ctx context.Context, accountID string) (int, error)
}
```

**Caches** (Redis):
```go
// pkg/infrastructure/cache/
├── response_cache.go       // Port 7001
├── account_cache.go        // Port 7002
├── validation_cache.go     // Port 7003
├── dedup_cache.go          // Port 7004
└── ratelimit_cache.go      // Port 7005
```

**Clients** (External Services):
```go
// pkg/infrastructure/clients/
├── connect_dict_client.go  // gRPC client to Connect DICT
├── validator_client.go     // HTTP client to Receita Federal
├── temporal_client.go      // Temporal workflow client
└── pulsar_client.go        // Pulsar publisher
```

---

## 7. Stack Tecnológica

### 7.1 Core Technologies

| Componente | Tecnologia | Versão | Motivo |
|------------|------------|--------|--------|
| **Linguagem** | Go | 1.24.5+ | Padrão LBPay, performance, concorrência |
| **REST Framework** | Chi / Gin | latest | Simplicidade, performance |
| **gRPC Framework** | gRPC-Go | v1.60+ | Comunicação interna eficiente |
| **ORM** | GORM | v2+ | Abstração de PostgreSQL |
| **Database** | PostgreSQL | 14+ | Já em uso, ACID compliant |
| **Cache** | Redis | 7+ | Já em uso, performance |
| **Message Broker** | Apache Pulsar | 3.x | Já em uso, async messaging |
| **Workflow Engine** | Temporal | 1.x | Já em uso, long-running processes |

### 7.2 Observability Stack

| Componente | Tecnologia | Versão | Propósito |
|------------|------------|--------|-----------|
| **Metrics** | Prometheus | 2.x | Time-series metrics |
| **Traces** | Jaeger | 1.x | Distributed tracing |
| **Logs** | Loki | 2.x | Log aggregation |
| **Dashboards** | Grafana | 10+ | Visualization |
| **Instrumentation** | OpenTelemetry | 1.x | Unified observability |

### 7.3 Infrastructure

| Componente | Tecnologia | Versão | Motivo |
|------------|------------|--------|--------|
| **Container Runtime** | Docker | 24+ | Containerização |
| **Orchestration** | Kubernetes | 1.28+ | Orquestração de containers |
| **Service Mesh** | Istio (opcional) | 1.20+ | Traffic management, mTLS |
| **API Gateway** | Kong / NGINX | latest | Gateway de entrada |
| **CI/CD** | GitHub Actions | - | Automação de deploy |
| **IaC** | Terraform | 1.6+ | Infrastructure as Code |

### 7.4 Security

| Componente | Tecnologia | Versão | Propósito |
|------------|------------|--------|-----------|
| **TLS/mTLS** | OpenSSL | 3.x | Certificados X.509 |
| **XML Signature** | xmlsec | 1.x | Assinatura digital XML |
| **Secrets Management** | HashiCorp Vault | 1.x | Gerenciamento de secrets |
| **Certificate Manager** | cert-manager | 1.x | Automação de certificados |

---

## 8. Fluxos de Dados Principais

### 8.1 Fluxo 1: Criar Chave PIX (CREATE)

**RF Atendido**: RF-BLO1-001
**Pré-requisito**: Validação de Posse (Manual Bacen Subseção 2.1)

#### 8.1.1 Fase 1: Validação de Posse (Obrigatória para PHONE/EMAIL)

**⚠️ Este fluxo DEVE ocorrer ANTES do registro no DICT**

```
┌───────────┐
│  Cliente  │
│ (Portal)  │
└─────┬─────┘
      │ 1. POST /api/v1/pixkeys/validate-ownership
      │    {keyType: "PHONE", key: "+5561988880000"}
      ↓
┌─────────────────────────┐
│  Core DICT (REST API)   │
│                         │
│  ValidateOwnershipUC    │
└─────┬───────────────────┘
      │ 2. Check keyType
      │    keyType == PHONE or EMAIL? ✅
      ↓
      │ 3. Generate 6-digit code
      │    code = "123456"
      │    token = sha256(key+code+salt)
      ↓
      │ 4. Store token in Redis (TTL 30 min)
      │    SET ownership:+5561988880000 = {token, expiresAt, validatedAt: null}
      ↓
┌─────────────────┐
│ Redis (dedup)   │
│   Port 7004     │
└─────┬───────────┘
      │ 5. Stored ✅ (expires in 30 min)
      ↓
      │ 6. Send SMS via Gateway
      │    "Seu código PIX: 123456"
      ↓
┌─────────────────┐
│  SMS Gateway    │
└─────┬───────────┘
      │ 7. SMS sent ✅
      ↓
      │ 8. Return token to frontend
      │    {token: "abc123...", expiresIn: 1800}
      ↓
┌───────────┐
│  Cliente  │
└─────┬─────┘
      │ 9. Show "Enter code" screen
      │    User inputs: "123456"
      ↓
      │ 10. POST /api/v1/pixkeys/confirm-ownership
      │     {key: "+5561988880000", code: "123456"}
      ↓
┌─────────────────────────┐
│  Core DICT              │
│  ConfirmOwnershipUC     │
└─────┬───────────────────┘
      │ 11. Fetch token from Redis
      ↓
┌─────────────────┐
│ Redis (dedup)   │
└─────┬───────────┘
      │ 12. Token found ✅
      │     {token, expiresAt, validatedAt: null}
      ↓
      │ 13. Validate code
      │     sha256(key+"123456"+salt) == token? ✅
      ↓
      │ 14. Mark as validated
      │     UPDATE ownership:+5561988880000 SET validatedAt = now()
      ↓
┌─────────────────┐
│ Redis (dedup)   │
└─────┬───────────┘
      │ 15. Updated ✅
      ↓
      │ 16. Return success + validated token
      │     {validated: true, token: "abc123..."}
      ↓
┌───────────┐
│  Cliente  │
└───────────┘
```

**⏱️ Timeout**: 30 minutos (configurável)
**🔄 Retry**: Se timeout expirar, usuário deve solicitar novo código

---

#### 8.1.2 Fase 2: Registro no DICT (Após Validação de Posse)

```
┌───────────┐
│  Cliente  │
└─────┬─────┘
      │ 1. POST /api/v1/pixkeys
      │    {key, account, owner, validationToken: "abc123..."}
      ↓
┌─────────────────┐
│   API Gateway   │
└─────┬───────────┘
      │ 2. JWT validation
      │    Rate limiting
      ↓
┌─────────────────────────┐
│  Core DICT (REST API)   │
│                         │
│  CreatePixKeyUseCase    │
└─────┬───────────────────┘
      │ 3. Check ownership validation (Subseção 2.1)
      │    If keyType == PHONE or EMAIL:
      │      ownershipValidator.IsValidated(key, token)? ✅
      ↓
┌─────────────────┐
│ Redis (dedup)   │
└─────┬───────────┘
      │ 4. Validation confirmed ✅
      │    (validatedAt != null)
      ↓
      │ 5. Validate input
      │    (KeyValidator)
      ↓
      │ 6. Check limits
      │    (LimitValidator)
      │    SELECT COUNT(*) FROM pixkeys WHERE account_id = ?
      ↓
┌─────────────────┐
│   PostgreSQL    │
└─────────────────┘
      │ 7. Count = 4 (PF) ✅
      ↓
      │ 8. Check dedup cache
      │    (RequestId already used?)
      ↓
┌─────────────────┐
│ Redis (dedup)   │
│   Port 7004     │
└─────────────────┘
      │ 9. Not found ✅
      ↓
      │ 10. Call Connect DICT
      │     (gRPC)
      ↓
┌──────────────────────┐
│   Connect DICT       │
│  (RSFN Bridge)       │
└─────┬────────────────┘
      │ 11. Check rate limit (local)
      ↓
┌─────────────────┐
│ Redis (ratelimit│
│   Port 7005     │
└─────────────────┘
      │ 12. Tokens available ✅
      ↓
      │ 13. Sign XML
      │     (XML Signer)
      ↓
      │ 14. POST https://dict.pi.rsfn.net.br:16422/api/v2/entries/
      │     (mTLS, XML Signed)
      ↓
┌─────────────────┐
│   DICT Bacen    │
└─────┬───────────┘
      │ 15. 201 Created
      │     (XML Signed Response)
      ↓
      │ 16. Validate signature
      │     Parse response
      ↓
      │ 17. Return to Core DICT
      ↓
┌─────────────────────────┐
│  Core DICT              │
│  CreatePixKeyUseCase    │
└─────┬───────────────────┘
      │ 18. Calculate CID
      │     (CIDCalculator)
      ↓
      │ 19. Save to PostgreSQL
      │     INSERT INTO pixkeys (key, account, owner, cid, ...)
      ↓
┌─────────────────┐
│   PostgreSQL    │
└─────────────────┘
      │ 20. Saved ✅
      ↓
      │ 21. Clear ownership token (security)
      │     DEL ownership:+5561988880000
      ↓
┌─────────────────┐
│ Redis (dedup)   │
└─────────────────┘
      │ 22. Token cleared ✅
      ↓
      │ 23. Update cache
      │     SET cache-dict-response:key = response (TTL 5min)
      ↓
┌─────────────────┐
│ Redis (response)│
│   Port 7001     │
└─────────────────┘
      │ 24. Cached ✅
      ↓
      │ 25. Publish event
      │     Topic: dict-events
      │     Event: PixKeyCreated
      ↓
┌─────────────────┐
│  Apache Pulsar  │
└─────────────────┘
      │ 26. Published ✅
      ↓
      │ 27. Return 201 Created
      ↓
┌───────────┐
│  Cliente  │
└───────────┘
```

**Latência esperada**:
- Fase 1 (Validação de Posse): 200-300ms (envio SMS/e-mail)
- Fase 2 (Registro no DICT): 300-500ms (P99)
- **Total do fluxo completo**: ~1 segundo (incluindo interação do usuário)

---

### 8.2 Fluxo 2: Consultar Chave PIX (READ) - COM CACHE HIT

**RF Atendido**: RF-BLO1-008, RF-BLO5-003 (crítico)

```
┌───────────┐
│  Cliente  │
└─────┬─────┘
      │ 1. GET /api/v1/pixkeys/{key}?taxIdNumber=11122233300
      ↓
┌─────────────────┐
│   API Gateway   │
└─────┬───────────┘
      │ 2. JWT validation
      ↓
┌─────────────────────────┐
│  Core DICT (REST API)   │
│  GetPixKeyUseCase       │
└─────┬───────────────────┘
      │ 3. Check response cache
      │    GET cache-dict-response:{key}:{taxId}
      ↓
┌─────────────────┐
│ Redis (response)│
│   Port 7001     │
└─────┬───────────┘
      │ 4. CACHE HIT ✅
      │    {key, account, owner, ...}
      ↓
┌─────────────────────────┐
│  Core DICT              │
│  GetPixKeyUseCase       │
└─────┬───────────────────┘
      │ 5. Return 200 OK
      │    (from cache)
      ↓
┌───────────┐
│  Cliente  │
└───────────┘
```

**Latência esperada**: 5-20ms (P99) ⚡

**Cache Hit Rate esperado**: 70-90%

---

### 8.3 Fluxo 3: Consultar Chave PIX (READ) - CACHE MISS

```
┌───────────┐
│  Cliente  │
└─────┬─────┘
      │ 1. GET /api/v1/pixkeys/{key}?taxIdNumber=11122233300
      ↓
┌─────────────────┐
│   API Gateway   │
└─────┬───────────┘
      │ 2. JWT validation
      ↓
┌─────────────────────────┐
│  Core DICT (REST API)   │
│  GetPixKeyUseCase       │
└─────┬───────────────────┘
      │ 3. Check response cache
      │    GET cache-dict-response:{key}:{taxId}
      ↓
┌─────────────────┐
│ Redis (response)│
│   Port 7001     │
└─────┬───────────┘
      │ 4. CACHE MISS ❌
      ↓
┌─────────────────────────┐
│  Core DICT              │
│  GetPixKeyUseCase       │
└─────┬───────────────────┘
      │ 5. Check validation cache
      │    GET cache-dict-key-validation:{key}
      ↓
┌─────────────────┐
│ Redis (validation)
│   Port 7003     │
└─────┬───────────┘
      │ 6. CACHE MISS ❌
      ↓
      │ 7. Call Connect DICT (gRPC)
      ↓
┌──────────────────────┐
│   Connect DICT       │
└─────┬────────────────┘
      │ 8. Check rate limit (local)
      │    Anti-scan policy (USER scope)
      ↓
┌─────────────────┐
│ Redis (ratelimit│
│   Port 7005     │
└─────┬───────────┘
      │ 9. PF: 100 tokens available ✅
      ↓
      │ 10. GET https://dict.pi.rsfn.net.br:16422/api/v2/entries/{key}?TaxIdNumber=xxx
      │     (mTLS, NOT signed)
      ↓
┌─────────────────┐
│   DICT Bacen    │
└─────┬───────────┘
      │ 11. 200 OK (XML Signed Response)
      ↓
      │ 12. Validate signature
      │     Parse response
      │     Subtract 1 token (status 200)
      ↓
      │ 13. Return to Core DICT
      ↓
┌─────────────────────────┐
│  Core DICT              │
│  GetPixKeyUseCase       │
└─────┬───────────────────┘
      │ 14. Update response cache
      │     SET cache-dict-response:{key}:{taxId} = response (TTL 5min)
      ↓
┌─────────────────┐
│ Redis (response)│
│   Port 7001     │
└─────────────────┘
      │ 15. Cached ✅
      ↓
      │ 16. Return 200 OK
      ↓
┌───────────┐
│  Cliente  │
└───────────┘
```

**Latência esperada**: 200-400ms (P99)

---

### 8.4 Fluxo 4: Reivindicação de Portabilidade (CLAIM - 7 dias)

**RF Atendido**: RF-BLO2-007, RF-BLO2-008

```
┌───────────────┐
│ PSP Reivindicador │
└─────┬─────────┘
      │ 1. POST /api/v1/claims
      │    {key, type: PORTABILITY, claimer: {...}}
      ↓
┌─────────────────────────┐
│  Core DICT              │
│  CreateClaimUseCase     │
└─────┬───────────────────┘
      │ 2. Validate input
      │    Check if key exists
      │    Check if claim already exists for key
      ↓
      │ 3. Call Connect DICT
      │    POST /claims/
      ↓
┌──────────────────────┐
│   Connect DICT       │
└─────┬────────────────┘
      │ 4. POST https://dict.pi.rsfn.net.br:16422/api/v2/claims/
      │    (mTLS, XML Signed)
      ↓
┌─────────────────┐
│   DICT Bacen    │
└─────┬───────────┘
      │ 5. 201 Created
      │    {claimId, status: OPEN, ...}
      ↓
      │ 6. Return to Core DICT
      ↓
┌─────────────────────────┐
│  Core DICT              │
│  CreateClaimUseCase     │
└─────┬───────────────────┘
      │ 7. Save to PostgreSQL
      │    INSERT INTO claims (claim_id, status: OPEN, ...)
      ↓
      │ 8. Start Temporal Workflow
      │    ClaimWorkflow (7-day process)
      ↓
┌─────────────────┐
│  Temporal       │
└─────┬───────────┘
      │ 9. Workflow started ✅
      │    Activity: MonitorClaimStatus (polling)
      ↓
      │ [7 DAYS OF MONITORING]
      │
      │ PSP Doador acknowledges claim (OPEN → WAITING_RESOLUTION)
      │ PSP Doador confirms claim (WAITING_RESOLUTION → CONFIRMED)
      │
      ↓
      │ 10. Activity: CompleteClaimActivity
      │     POST /api/v1/claims/{claimId}/complete
      ↓
┌─────────────────────────┐
│  Core DICT              │
│  CompleteClaimUseCase   │
└─────┬───────────────────┘
      │ 11. Call Connect DICT
      │     POST /claims/{claimId}/complete
      ↓
┌──────────────────────┐
│   Connect DICT       │
└─────┬────────────────┘
      │ 12. POST https://dict.pi.rsfn.net.br:16422/api/v2/claims/{claimId}/complete
      ↓
┌─────────────────┐
│   DICT Bacen    │
└─────┬───────────┘
      │ 13. 200 OK
      │     Status: COMPLETED
      │     New entry created automatically
      ↓
      │ 14. Return to Core DICT
      ↓
┌─────────────────────────┐
│  Core DICT              │
│  CompleteClaimUseCase   │
└─────┬───────────────────┘
      │ 15. Update claim status (COMPLETED)
      │     Create new PixKey in PostgreSQL
      ↓
      │ 16. Invalidate caches
      │     DEL cache-dict-response:{key}:*
      ↓
      │ 17. Publish event
      │     Topic: dict-events
      │     Event: ClaimCompleted
      ↓
      │ 18. Workflow completes ✅
      ↓
┌───────────────┐
│ PSP Reivindicador │
└───────────────┘
```

**Duração total**: 7 dias (máximo regulamentar)

---

## 9. Estratégia de Performance

### 9.1 Cache Multi-Camadas (5 Redis)

**Objetivo**: Reduzir 70-90% das chamadas ao DICT Bacen.

#### 9.1.1 cache-dict-response (Port 7001)

**Propósito**: Cache de respostas completas de consultas.

**Keys**:
```
cache-dict-response:{keyType}:{key}:{taxIdNumber}
```

**Value**: Resposta XML completa do DICT Bacen

**TTL**: 5 minutos

**Invalidação**:
- Evento CID (chave atualizada/deletada)
- Manual (via API admin)

**Benefício**: 70-80% de cache hit rate esperado.

#### 9.1.2 cache-dict-account (Port 7002)

**Propósito**: Cache de dados de conta transacional.

**Keys**:
```
cache-dict-account:{participant}:{branch}:{accountNumber}
```

**Value**: Dados de conta (JSON)

**TTL**: 15 minutos

**Benefício**: Reduz queries a PostgreSQL.

#### 9.1.3 cache-dict-key-validation (Port 7003)

**Propósito**: Cache de validações de chaves (CPF/CNPJ).

**Keys**:
```
cache-dict-key-validation:{taxIdNumber}
```

**Value**: Status de validação (JSON)

**TTL**: 10 minutos

**Benefício**: Reduz chamadas à Receita Federal.

#### 9.1.4 cache-dict-dedup (Port 7004)

**Propósito**: Deduplicação de requisições (idempotência).

**Keys**:
```
cache-dict-dedup:{requestId}
```

**Value**: Hash do request body

**TTL**: 1 minuto

**Benefício**: Previne requisições duplicadas.

#### 9.1.5 cache-dict-rate-limit (Port 7005)

**Propósito**: Rate limiting local (client-side).

**Keys**:
```
cache-dict-rate-limit:{policy}:{scope}
```

**Value**: Número de tokens disponíveis

**TTL**: Variável (conforme taxa de reposição)

**Benefício**: Previne 429 do DICT Bacen.

### 9.2 Connection Pooling

**Configuração** (Connect DICT):
```go
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
    TLSClientConfig:     tlsConfig, // mTLS
    DisableKeepAlives:   false,
}
```

**Benefício**: Reduz latência de mTLS handshake (100-300ms → 5-20ms).

### 9.3 Compressão

**Request**: `Accept-Encoding: gzip`

**Benefício**: Reduz largura de banda em 60-80%.

### 9.4 Batch Operations

**Use Case**: Reconciliação (validar 1000s de chaves).

**Abordagem**:
- `/keys/check` (até 1000 chaves por requisição)
- `/cids/files/` (download de arquivo completo de CIDs)

### 9.5 Horizontal Scaling

**Auto-scaling** (Kubernetes HPA):
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: core-dict
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: core-dict
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Pods
    pods:
      metric:
        name: http_request_duration_seconds_p99
      target:
        type: AverageValue
        averageValue: "1000m"  # 1s
```

---

## 10. Estratégia de Resiliência

### 10.1 Circuit Breaker

**Implementação**: [sony/gobreaker](https://github.com/sony/gobreaker)

**Configuração** (Connect DICT → DICT Bacen):
```go
cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "dict-bacen",
    MaxRequests: 3,        // Half-open state
    Interval:    60 * time.Second,
    Timeout:     30 * time.Second,
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
        return counts.Requests >= 10 && failureRatio >= 0.5
    },
})
```

**Estados**:
- **Closed**: Normal operation
- **Open**: Falhas > 50% (não faz requisições)
- **Half-Open**: Testa 3 requisições após 30s

### 10.2 Retry Logic

**Algoritmo**: Exponential Backoff with Jitter

**Configuração**:
```go
retry.Do(
    func() error {
        return callDICTBacen()
    },
    retry.Attempts(3),
    retry.Delay(1 * time.Second),
    retry.MaxDelay(60 * time.Second),
    retry.DelayType(retry.BackOffDelay),
    retry.RetryIf(func(err error) bool {
        return isRetryable(err) // 429, 500, 503, timeout
    }),
)
```

### 10.3 Timeouts

| Operação | Timeout | Motivo |
|----------|---------|--------|
| **REST request (DICT Bacen)** | 5s | SLA Bacen |
| **gRPC call (internal)** | 2s | Internal network |
| **PostgreSQL query** | 3s | Database |
| **Redis operation** | 500ms | Cache |
| **Temporal workflow** | 7 days | Claim process |

### 10.4 Bulkhead Pattern

**Implementação**: Thread pools separados para cada tipo de operação.

```go
// Separate semaphores for different operations
var (
    readSemaphore  = semaphore.NewWeighted(100)  // 100 concurrent reads
    writeSemaphore = semaphore.NewWeighted(50)   // 50 concurrent writes
    claimSemaphore = semaphore.NewWeighted(20)   // 20 concurrent claims
)
```

**Benefício**: Evita que operações lentas (writes) bloqueiem operações rápidas (reads).

### 10.5 Graceful Shutdown

**Implementação**:
```go
func main() {
    // ...
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // 1. Stop accepting new requests
    // 2. Finish in-flight requests (max 30s)
    // 3. Close connections (PostgreSQL, Redis, Pulsar)
    // 4. Exit
}
```

---

## 11. Segurança

### 11.1 Segurança em Camadas

| Camada | Controle | Implementação |
|--------|----------|---------------|
| **Network** | mTLS | Certificado X.509 (Bacen) |
| **Application** | JWT | Authentication (external APIs) |
| **Data** | Encryption | TDE (PostgreSQL), TLS (Redis) |
| **Message** | Signature | XML Digital Signature (SHA-256 + RSA) |
| **Rate Limiting** | Token Bucket | Local + Bacen |

### 11.2 mTLS (Mutual TLS)

**Certificados**:
- **Client**: Emitido pelo Bacen (ICP-Brasil)
- **Server**: Validado pelo Bacen

**Rotação**: Automática (cert-manager + Vault)

**Configuração** (Go):
```go
cert, err := tls.LoadX509KeyPair("client.crt", "client.key")
caCert, err := ioutil.ReadFile("ca.crt")
caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)

tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{cert},
    RootCAs:      caCertPool,
    MinVersion:   tls.VersionTLS12,
}
```

### 11.3 XML Digital Signature

**Padrão**: XML Signature (envelopada)

**Algoritmos**:
- Hash: SHA-256
- Assinatura: RSA-2048 (mínimo)

**Certificado**: P12 (mesma chain do mTLS)

**Implementação**:
- Opção 1: Go puro (crypto/rsa, encoding/xml)
- Opção 2: Java Signer Service (já existente)

### 11.4 Secrets Management

**HashiCorp Vault**:
```
/secret/lbpay/dict/prod/
├── db-password
├── redis-password
├── pulsar-token
├── temporal-token
├── dict-cert-p12-password
└── jwt-secret
```

**Acesso**: Kubernetes Service Account (via Vault Agent Injector)

### 11.5 Compliance LGPD

**Dados sensíveis**:
- CPF/CNPJ (Owner.TaxIdNumber)
- Nome (Owner.Name)
- Telefone (Key quando KeyType=PHONE)
- Email (Key quando KeyType=EMAIL)

**Controles**:
- ✅ Encryption at-rest (PostgreSQL TDE)
- ✅ Encryption in-transit (TLS 1.2+)
- ✅ Logs auditáveis (quem acessou, quando, o quê)
- ✅ Data masking (logs não expõem CPF completo)
- ✅ Retention policy (7 anos conforme regulamento Bacen)

---

## 12. Observabilidade

### 12.1 Métricas (Prometheus)

#### 12.1.1 Métricas RED (Request)

```prometheus
# Rate (throughput)
dict_http_requests_total{endpoint, method, status}

# Errors (error rate)
dict_http_errors_total{endpoint, error_type}

# Duration (latency)
dict_http_request_duration_seconds{endpoint, method, quantile="0.5|0.95|0.99"}
```

#### 12.1.2 Métricas USE (Resource)

```prometheus
# Utilization
dict_cpu_usage_percent
dict_memory_usage_bytes
dict_db_connections_active

# Saturation
dict_http_requests_queued
dict_db_connections_waiting

# Errors
dict_db_errors_total
dict_cache_errors_total
```

#### 12.1.3 Métricas de Negócio

```prometheus
# DICT operations
dict_pixkeys_created_total
dict_pixkeys_deleted_total
dict_claims_created_total{type="OWNERSHIP|PORTABILITY"}
dict_claims_completed_total

# Cache performance
dict_cache_hits_total{cache_name}
dict_cache_misses_total{cache_name}
dict_cache_hit_ratio{cache_name}

# Rate limiting
dict_rate_limited_requests_total{policy}
dict_rate_limit_tokens_available{policy}
```

### 12.2 Traces (Jaeger + OpenTelemetry)

**Instrumentação**:
```go
import "go.opentelemetry.io/otel"

func CreatePixKey(ctx context.Context, input CreatePixKeyInput) (*CreatePixKeyOutput, error) {
    ctx, span := otel.Tracer("core-dict").Start(ctx, "CreatePixKey")
    defer span.End()

    // Business logic...
    // Automatically propagates context to:
    // - gRPC calls (Connect DICT)
    // - PostgreSQL queries
    // - Redis operations
    // - Pulsar publish
}
```

**Trace Context Propagation**: W3C Trace Context

**Sampling**: 10% em produção (100% em homologação)

### 12.3 Logs (Loki)

**Formato**: JSON estruturado

**Exemplo**:
```json
{
  "timestamp": "2023-10-24T10:00:00.123Z",
  "level": "INFO",
  "service": "core-dict",
  "trace_id": "abc123def456...",
  "span_id": "789ghi012jkl...",
  "endpoint": "POST /api/v1/pixkeys",
  "method": "CreatePixKey",
  "key": "+55619****0000",  // masked
  "account_id": "12345678-0001-0007654321",
  "duration_ms": 450,
  "status": "success"
}
```

**Masking**: CPF/CNPJ, telefone, email são parcialmente mascarados.

**Retention**: 90 dias (comprimido)

### 12.4 Dashboards (Grafana)

#### 12.4.1 Dashboard: Core DICT Overview

**Panels**:
- Throughput (requests/sec) - Line chart
- Latency P50/P95/P99 - Line chart
- Error Rate (%) - Line chart
- Cache Hit Ratio (%) - Gauge
- Top 10 Endpoints by Latency - Table
- Alert Status - Status panel

#### 12.4.2 Dashboard: DICT Bacen Integration

**Panels**:
- Requests to DICT Bacen (by endpoint) - Bar chart
- Latency to DICT Bacen P99 - Line chart
- Rate Limited Requests (429) - Counter
- Circuit Breaker Status - Status panel
- Connection Pool Utilization - Gauge

#### 12.4.3 Dashboard: Business Metrics

**Panels**:
- PIX Keys Created (last 24h) - Counter
- Claims Created (by type) - Pie chart
- Refunds Requested (last 7d) - Line chart
- Top 10 Accounts by Key Count - Table

### 12.5 Alertas

**Plataforma**: Prometheus Alertmanager + PagerDuty

**Alertas Críticos** (P1 - Page 24/7):
```yaml
- alert: HighErrorRate
  expr: rate(dict_http_errors_total[5m]) > 0.05
  for: 5m
  severity: critical
  annotations:
    summary: "Error rate > 5% for 5 minutes"

- alert: HighLatency
  expr: histogram_quantile(0.99, dict_http_request_duration_seconds) > 1.0
  for: 5m
  severity: critical
  annotations:
    summary: "P99 latency > 1s for 5 minutes"

- alert: ServiceDown
  expr: up{job="core-dict"} == 0
  for: 1m
  severity: critical
  annotations:
    summary: "Core DICT service is down"
```

**Alertas de Warning** (P2 - Slack):
```yaml
- alert: CacheHitRatioLow
  expr: dict_cache_hit_ratio{cache_name="response"} < 0.5
  for: 15m
  severity: warning
  annotations:
    summary: "Cache hit ratio < 50% for 15 minutes"

- alert: RateLimitedRequests
  expr: rate(dict_rate_limited_requests_total[5m]) > 0.01
  for: 10m
  severity: warning
  annotations:
    summary: "Rate limited requests > 1% for 10 minutes"
```

---

## 13. Estratégia de Deployment

### 13.1 Ambientes

| Ambiente | Namespace | Propósito | Dados |
|----------|-----------|-----------|-------|
| **Development** | `lbpay-dict-dev` | Desenvolvimento local | Mock/Simulador |
| **Homologation** | `lbpay-dict-hml` | Testes integrados | DICT Bacen HML |
| **Production** | `lbpay-dict-prod` | Produção | DICT Bacen PROD |

### 13.2 CI/CD Pipeline (GitHub Actions)

```yaml
name: Core DICT CI/CD

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24.5'
      - name: Run tests
        run: make test
      - name: Code coverage
        run: make coverage
      - name: Upload coverage
        uses: codecov/codecov-action@v3

  lint:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3

  build:
    needs: [test, lint]
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
      - name: Build Docker image
        run: docker build -t core-dict:${{ github.sha }} .
      - name: Push to registry
        run: docker push core-dict:${{ github.sha }}

  deploy-hml:
    needs: build
    if: github.ref == 'refs/heads/develop'
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to Homologation
        run: |
          kubectl set image deployment/core-dict \
            core-dict=core-dict:${{ github.sha }} \
            -n lbpay-dict-hml

  deploy-prod:
    needs: build
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to Production (Blue-Green)
        run: |
          # 1. Deploy green (new version)
          kubectl apply -f k8s/deployment-green.yaml
          # 2. Wait for green to be healthy
          kubectl wait --for=condition=available deployment/core-dict-green
          # 3. Switch traffic (update service selector)
          kubectl patch service core-dict -p '{"spec":{"selector":{"version":"green"}}}'
          # 4. Wait 5min monitoring
          sleep 300
          # 5. Delete blue (old version)
          kubectl delete deployment core-dict-blue
```

### 13.3 Deployment Strategy

**Blue-Green Deployment**:
```
┌─────────────┐
│   Service   │
│  (Load Bal) │
└──────┬──────┘
       │
       ├──> 100% ─┐
       │          │
       │          ↓
       │    ┌──────────┐
       │    │  BLUE    │  ← Old version (v1.0)
       │    │ (3 pods) │
       │    └──────────┘
       │
       │
       └──> 0% ───┐
                  │
                  ↓
            ┌──────────┐
            │  GREEN   │  ← New version (v1.1)
            │ (3 pods) │  (deploying...)
            └──────────┘

[After health check passed]

       ├──> 0% ───┐
       │          │
       │          ↓
       │    ┌──────────┐
       │    │  BLUE    │  (will be deleted)
       │    │ (3 pods) │
       │    └──────────┘
       │
       │
       └──> 100% ─┐
                  │
                  ↓
            ┌──────────┐
            │  GREEN   │  ← Now serving traffic
            │ (3 pods) │
            └──────────┘
```

**Rollback**: Switch service selector back to blue (instant).

### 13.4 Kubernetes Manifests

#### 13.4.1 Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: core-dict
  namespace: lbpay-dict-prod
spec:
  replicas: 3
  selector:
    matchLabels:
      app: core-dict
  template:
    metadata:
      labels:
        app: core-dict
        version: v1.0.0
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8081"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: core-dict
      containers:
      - name: core-dict
        image: core-dict:latest
        ports:
        - name: http
          containerPort: 8080
        - name: grpc
          containerPort: 9090
        - name: metrics
          containerPort: 8081
        env:
        - name: ENV
          value: "production"
        - name: DB_HOST
          value: "postgresql.lbpay-infra.svc.cluster.local"
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: dict-secrets
              key: db-password
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8081
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8081
          initialDelaySeconds: 10
          periodSeconds: 5
```

#### 13.4.2 Service
```yaml
apiVersion: v1
kind: Service
metadata:
  name: core-dict
  namespace: lbpay-dict-prod
spec:
  selector:
    app: core-dict
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: grpc
    port: 9090
    targetPort: 9090
  type: ClusterIP
```

#### 13.4.3 Ingress
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: core-dict
  namespace: lbpay-dict-prod
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/rate-limit: "100"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - dict-api.lbpay.com
    secretName: dict-api-tls
  rules:
  - host: dict-api.lbpay.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: core-dict
            port:
              number: 80
```

---

## 14. Decisões Arquiteturais (ADRs)

As seguintes ADRs devem ser criadas para documentar decisões importantes:

### 14.1 ADR-001: Clean Architecture

**Status**: ✅ Aprovado

**Contexto**: Necessidade de separar lógica de negócio de infraestrutura.

**Decisão**: Adotar Clean Architecture (Domain, Application, Handlers, Infrastructure).

**Consequências**:
- ✅ Testabilidade alta
- ✅ Independência de frameworks
- ✅ Facilita manutenção
- ⚠️ Curva de aprendizado inicial

### 14.2 ADR-002: Consolidação em Repositório Único

**Status**: ⏳ A criar (documento separado)

**Contexto**: Lógica DICT dispersa em múltiplos repos.

**Decisão**: Criar `core-dict` único para toda lógica de negócio DICT.

**Documento**: [ADR-002_Consolidacao_Core_DICT.md](#)

### 14.3 ADR-003: Performance Multi-Camadas

**Status**: ⏳ A criar (documento separado)

**Contexto**: Requisito de alto volume (dezenas de queries/segundo).

**Decisão**: 5 caches Redis especializados + connection pooling.

**Documento**: [ADR-003_Performance_Multi_Camadas.md](#)

### 14.4 ADR-004: Bridge DICT Dedicado

**Status**: ⏳ A criar (documento separado)

**Contexto**: Comunicação com DICT Bacen requer mTLS, XML signing, rate limiting.

**Decisão**: Componente dedicado (Connect DICT) reutilizando padrões do RSFN Bridge.

**Documento**: [ADR-004_Bridge_DICT_Dedicado.md](#)

### 14.5 ADR-005: Temporal para Workflows Longos

**Status**: ✅ Aprovado

**Contexto**: Claims podem durar 7 dias (processo assíncrono).

**Decisão**: Usar Temporal Workflow Engine.

**Consequências**:
- ✅ Garantia de execução (durável)
- ✅ Transparência (visualização de estado)
- ✅ Retry automático
- ⚠️ Complexidade adicional

---

## 15. Roadmap de Implementação

### 15.1 Fases de Implementação

Conforme [CRF-001](../05_Requisitos/CRF-001_Checklist_Requisitos_Funcionais.md):

#### Fase 1: Fundação (Semanas 1-12, 480h)
**Objetivo**: Infraestrutura e endpoints críticos.

**Entregas**:
- ✅ Repositório `core-dict` criado (estrutura Clean Architecture)
- ✅ Connect DICT refatorado (28 API clients)
- ✅ RF-BLO5-003 (Interface Communication - **crítico path**)
- ✅ RF-BLO1-001, RF-BLO1-008 (CRUD básico)
- ✅ 5 caches Redis configurados
- ✅ Observabilidade básica (métricas, traces, logs)
- ✅ Testes unitários (cobertura 80%)
- ✅ Homologação Bacen iniciada

**Bloqueadores**:
- Certificado mTLS (emissão Bacen)
- Acesso ambiente Homologação Bacen

#### Fase 2: CRUD Completo (Semanas 13-24, 480h)
**Objetivo**: Implementar 100% do Bloco 1 (CRUD).

**Entregas**:
- ✅ RF-BLO1-001 a RF-BLO1-013 (100% do Bloco 1)
- ✅ RF-BLO3-001, RF-BLO3-002 (Validações)
- ✅ RF-BLO5-009, RF-BLO5-010 (Reconciliação)
- ✅ Testes de integração (simulador DICT)
- ✅ Homologação Bacen (Bloco 1) ✅

#### Fase 3: Reivindicações (Semanas 25-39, 600h)
**Objetivo**: Implementar Bloco 2 (Reivindicações).

**Entregas**:
- ✅ RF-BLO2-001 a RF-BLO2-014 (100% do Bloco 2)
- ✅ Temporal Workflows (ClaimWorkflow 7 dias)
- ✅ Testes E2E de reivindicação
- ✅ Homologação Bacen (Bloco 2) ✅

#### Fase 4: Devoluções/Infrações (Semanas 40-51, 480h)
**Objetivo**: Implementar Bloco 4 (Devoluções).

**Entregas**:
- ✅ RF-BLO4-001 a RF-BLO4-006 (100% do Bloco 4)
- ✅ RF-BLO5-008 (Estatísticas antifraude)
- ✅ Testes de performance (load testing)
- ✅ Homologação Bacen completa ✅
- ✅ Go-Live em Produção 🚀

**Total**: **51 semanas**, **2.040 horas**

### 15.2 Dependências Críticas

1. **Certificado mTLS** (Bacen) - Bloqueia Fase 1
2. **Acesso Homologação Bacen** - Bloqueia Fase 1
3. **API Receita Federal** (CPF/CNPJ) - Bloqueia RF-BLO3-002
4. **Aprovação Checklist Bacen** - Bloqueia Go-Live

---

## 16. Referências

### 16.1 Documentos do Projeto

1. **CRF-001** - Checklist de Requisitos Funcionais
   [Artefatos/05_Requisitos/CRF-001_Checklist_Requisitos_Funcionais.md](../05_Requisitos/CRF-001_Checklist_Requisitos_Funcionais.md)

2. **API-001** - Especificação de APIs DICT Bacen
   [Artefatos/04_APIs/API-001_Especificacao_APIs_DICT_Bacen.md](../04_APIs/API-001_Especificacao_APIs_DICT_Bacen.md)

3. **ARE-003** - Análise Documento Arquitetura DICT
   [Artefatos/02_Arquitetura/ARE-003_Analise_Documento_Arquitetura_DICT.md](ARE-003_Analise_Documento_Arquitetura_DICT.md)

4. **ArquiteturaDict_LBPAY.md** (Docs iniciais)
   `/Users/jose.silva.lb/LBPay/IA_Dict/Docs_iniciais/ArquiteturaDict_LBPAY.md`

### 16.2 Documentação Bacen

1. **Manual Operacional do DICT**
   https://www.bcb.gov.br/content/estabilidadefinanceira/pix/Regulamento_Pix/X_ManualOperacionaldoDICT.pdf

2. **Manual de Segurança PIX**
   https://www.bcb.gov.br/content/estabilidadefinanceira/cedsfn/Manual_de_Seguranca_PIX.pdf

3. **Manual de Tempos do Pix**
   https://www.bcb.gov.br/content/estabilidadefinanceira/pix/Regulamento_Pix/IX_ManualdeTemposdoPix.pdf

### 16.3 Arquitetura e Padrões

1. **Clean Architecture** - Robert C. Martin
   https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html

2. **C4 Model** - Simon Brown
   https://c4model.com/

3. **Circuit Breaker Pattern**
   https://martinfowler.com/bliki/CircuitBreaker.html

4. **Token Bucket Algorithm**
   https://en.wikipedia.org/wiki/Token_bucket

---

**Documento criado por**: NEXUS (AGT-ARC-001) - Solution Architect
**Data**: 2025-10-24
**Versão**: 1.0
**Status**: ✅ Completo - Aguardando Review

---

**Estatísticas do Documento**:
- **8 diagramas** (Contexto, Containers, Componentes, Fluxos)
- **5 caches Redis** especificados
- **72 RFs** mapeados à arquitetura
- **4 fases** de implementação (51 semanas)
- **20+ decisões arquiteturais** documentadas
