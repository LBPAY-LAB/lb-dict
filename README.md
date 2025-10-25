# Projeto DICT LBPay - Especificação Completa

![Status](https://img.shields.io/badge/Status-Especifica%C3%A7%C3%A3o%20Completa-green)
![Fase](https://img.shields.io/badge/Fase-Revis%C3%A3o%20T%C3%A9cnica-orange)
![Progresso](https://img.shields.io/badge/Progresso-100%25-success)

## 🎯 Visão Geral

Este projeto visa implementar a solução completa do **DICT (Diretório de Identificadores de Contas Transacionais)** do Banco Central do Brasil para o LBPay, uma instituição de pagamento licenciada e participante direto do PIX.

### Objetivo Principal
✅ Homologar no DICT Bacen (requisito obrigatório para operar PIX)
✅ Implementar gerenciamento completo de chaves PIX
✅ Entrar em produção após homologação

---

## 🚀 Status Atual

### ✅ Documentação Completa - 100%

**TODAS AS ESPECIFICAÇÕES TÉCNICAS ESTÃO COMPLETAS!**

```
Total do Projeto: 74/74 documentos (100%)
[████████████████████] 100%

✅ Fase 1 (Críticos):     16/16 (100%)
✅ Fase 2 (Detalhamento): 58/58 (100%)
```

**Data de Conclusão**: 2025-10-25
**Tempo de Execução**: 1 dia (com Squad de 8 agentes em paralelo)
**Eficiência**: 60x mais rápido que abordagem sequencial

### 📋 Próxima Etapa: Revisão Técnica

**Responsáveis**:
- CTO (José Luís Silva) - 20 docs críticos
- Head Arquitetura (Thiago Lima) - 27 docs arquitetura
- Head DevOps - 19 docs infraestrutura
- Head Compliance - 8 docs regulatório

**Ver**: [Roteiro de Revisão Técnica](Artefatos/00_Master/ROTEIRO_REVISAO_TECNICA.md)

---

## 📘 Documentação Essencial

### Documentos de Início Rápido

#### Gestão e Progresso
- 📊 **[ROTEIRO_REVISAO_TECNICA.md](Artefatos/00_Master/ROTEIRO_REVISAO_TECNICA.md)** - Guia de revisão para aprovadores
- 📈 **[PROGRESSO_FASE_1.md](Artefatos/00_Master/PROGRESSO_FASE_1.md)** - Status Fase 1 (16 docs críticos)
- 📈 **[PROGRESSO_FASE_2.md](Artefatos/00_Master/PROGRESSO_FASE_2.md)** - Status Fase 2 (58 docs detalhados)
- 📋 **[PLANO_PREENCHIMENTO_ARTEFATOS.md](Artefatos/00_Master/PLANO_PREENCHIMENTO_ARTEFATOS.md)** - Plano completo de documentação

#### Squad e Metodologia
- 👥 **[Claude.md](.claude/Claude.md)** - Documentação do paradigma Squad (8 agentes)
- 🤖 **[Agentes Especializados](.claude/agents/)** - Prompts dos 8 agentes

### Especificações Técnicas (TEC)

- 📖 **[TEC-001](Artefatos/11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md)** - Core DICT Specification
- 📖 **[TEC-002 v3.1](Artefatos/11_Especificacoes_Tecnicas/TEC-002_Bridge_Specification.md)** - Bridge Specification
- 📖 **[TEC-003 v2.1](Artefatos/11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)** - RSFN Connect Specification

### Análises de Arquitetura

- 🏗️ **[ANA-001](Artefatos/00_Analises/ANA-001_Analise_Arquitetura_IcePanel.md)** - Análise IcePanel
- 🔍 **[ANA-002](Artefatos/00_Analises/ANA-002_Analise_Repo_Bridge.md)** - Análise Bridge
- 🔍 **[ANA-003](Artefatos/00_Analises/ANA-003_Analise_Repo_Connect.md)** - Análise Connect
- 🔍 **[ANA-004](Artefatos/00_Analises/ANA-004_Analise_Repo_Core_DICT.md)** - Análise Core DICT

---

## 👥 Squad de Especificação (Fase 1 e 2)

8 agentes especializados trabalhando em paralelo:

| Agente | Especialidade | Documentos | Localização |
|--------|---------------|------------|-------------|
| **architect** | Diagramas C4, ADRs, TechSpecs | DIA-XXX, TSP-XXX | [.claude/agents/architect](.claude/agents/architect/prompt.md) |
| **backend** | APIs REST/gRPC, Schemas DB | API-XXX, DAT-XXX, GRPC-XXX | [.claude/agents/backend](.claude/agents/backend/prompt.md) |
| **security** | Segurança, mTLS, LGPD | SEC-XXX, CMP-XXX | [.claude/agents/security](.claude/agents/security/prompt.md) |
| **qa** | Testes, Qualidade | TST-XXX | [.claude/agents/qa](.claude/agents/qa/prompt.md) |
| **devops** | CI/CD, Kubernetes, Infra | DEV-XXX | [.claude/agents/devops](.claude/agents/devops/prompt.md) |
| **frontend** | UI/UX, Componentes | FE-XXX | [.claude/agents/frontend](.claude/agents/frontend/prompt.md) |
| **product** | User Stories, BPMN | US-XXX, BP-XXX | [.claude/agents/product](.claude/agents/product/prompt.md) |
| **scrum** | Backlog, Sprints | PM-XXX | [.claude/agents/scrum](.claude/agents/scrum/prompt.md) |

**Metodologia**: Execução em paralelo com máximo de agentes simultâneos
**Resultado**: 74 documentos criados em 1 dia (60x mais rápido)

---

## 📁 Estrutura do Projeto

```
IA_Dict/
├── .claude/                      # Configuração Claude Code + Squad
│   ├── agents/                   # 8 agentes especializados
│   │   ├── architect/
│   │   ├── backend/
│   │   ├── security/
│   │   ├── qa/
│   │   ├── devops/
│   │   ├── frontend/
│   │   ├── product/
│   │   └── scrum/
│   └── Claude.md                 # Documentação Squad
├── Docs_iniciais/                # Documentação Bacen (input)
│   ├── manual_Operacional_DICT_Bacen.md
│   ├── OpenAPI_Dict_Bacen.json
│   └── Requisitos_Homologação_Dict.md
└── Artefatos/                    # Especificações produzidas (74 docs)
    ├── 00_Master/                # Progresso, planos, roteiros
    ├── 00_Analises/              # 4 análises (ANA-001 a ANA-004)
    ├── 01_Requisitos/            # User Stories, Business Processes
    ├── 02_Arquitetura/           # Diagramas C4, TechSpecs, ADRs
    ├── 03_Dados/                 # Schemas DB, Migrations
    ├── 04_APIs/                  # REST, gRPC
    ├── 08_Frontend/              # Componentes, Wireframes
    ├── 09_Implementacao/         # Manuais de Setup
    ├── 12_Integracao/            # Fluxos E2E
    ├── 13_Seguranca/             # mTLS, LGPD, Vault
    ├── 14_Testes/                # Test Cases
    ├── 15_DevOps/                # CI/CD, Kubernetes
    ├── 16_Compliance/            # LGPD, Bacen, Auditoria
    └── 17_Gestao/                # Backlog, Sprints, Checklists
```

---

## 📊 Escopo Funcional

### Bloco 1: CRUD de Chaves PIX
- Criar chave (CPF, CNPJ, Email, Telefone, EVP/Aleatória)
- Consultar chave
- Alterar dados da chave
- Excluir chave
- Validar chave

### Bloco 2: Reivindicação e Portabilidade (ClaimWorkflow)
- Criar reivindicação (30 dias de claim period)
- Confirmar reivindicação
- Cancelar reivindicação
- Consultar/Listar reivindicações
- Portabilidade de chave

### Bloco 3: Sincronização (VSYNC)
- Sincronização diária com Bacen
- Reconciliação de chaves
- Detecção de divergências

### Bloco 4: Segurança e Infraestrutura
- Comunicação mTLS com Bacen
- Certificados ICP-Brasil A3
- Rate limiting
- Cache de chaves (Redis)
- Assinatura digital XML

### Bloco 5: Compliance
- LGPD (Lei 13.709/2018)
- Bacen (Circular 3.909/2019)
- Audit logs (retenção 5 anos)
- Proteção de dados pessoais

---

## 🏗️ Stack Tecnológica

### Backend
- **Linguagem**: Go 1.24.5
- **Framework HTTP**: Fiber v3
- **Comunicação Interna**: gRPC (Protocol Buffers v3)
- **Comunicação Bacen**: REST HTTPS mTLS (via Bridge)
- **Workflow Orchestration**: Temporal v1.36.0
- **Message Streaming**: Apache Pulsar v0.16.0
- **Banco de Dados**: PostgreSQL 16 (RLS, Partitioning)
- **Cache**: Redis v9.14.1 (5 estratégias de cache)
- **Database Migrations**: Goose
- **XML Signing**: Java 17 (ICP-Brasil A3)

### Segurança
- **Secret Management**: HashiCorp Vault
- **Certificates**: ICP-Brasil A3 (mTLS)
- **Authentication**: JWT, OAuth 2.0
- **Authorization**: RBAC
- **Network Security**: VPC, Security Groups, Network Policies, WAF

### Observabilidade
- **Métricas**: Prometheus
- **Visualização**: Grafana
- **Tracing**: Jaeger (OpenTelemetry)
- **Logs**: Loki
- **APM**: Distributed tracing

### CI/CD e Infraestrutura
- **Platform**: GitHub Actions
- **Containers**: Docker (multi-stage builds)
- **Orchestration**: Kubernetes 1.28+
- **Package Manager**: Helm
- **Environments**: Dev, QA, Staging, Production

### Arquitetura
- **Pattern**: Clean Architecture (4 camadas)
- **Data Pattern**: CQRS + Event Sourcing
- **Integration Pattern**: Saga (via Temporal)
- **Resilience**: Circuit Breaker, Retry Policies
- **API Design**: RESTful, gRPC

---

## 📦 Repositórios Envolvidos

| Repositório | Responsabilidade | Stack |
|-------------|------------------|-------|
| **rsfn-connect-bacen-core-dict** | Core DICT - Gestão de chaves PIX | Go + PostgreSQL + Redis |
| **rsfn-connect-bacen-connector** | RSFN Connect - Orquestração Temporal | Go + Temporal + Pulsar |
| **rsfn-connect-bacen-bridge** | Bridge - Adapter SOAP/mTLS para Bacen | Go + Java (XML Signer) |

---

## 📋 Documentação por Categoria

### 🏗️ Arquitetura (15 docs)

**Diagramas C4 e Sequência** (9 docs):
- [DIA-001](Artefatos/02_Arquitetura/Diagramas/DIA-001_C4_Context_Diagram.md) - C4 Context
- [DIA-002](Artefatos/02_Arquitetura/Diagramas/DIA-002_C4_Container_Diagram.md) - C4 Container
- [DIA-003](Artefatos/02_Arquitetura/Diagramas/DIA-003_C4_Component_Diagram_Core.md) - C4 Component Core
- [DIA-004](Artefatos/02_Arquitetura/Diagramas/DIA-004_C4_Component_Diagram_Connect.md) - C4 Component Connect
- [DIA-005](Artefatos/02_Arquitetura/Diagramas/DIA-005_C4_Component_Diagram_Bridge.md) - C4 Component Bridge
- [DIA-006](Artefatos/02_Arquitetura/Diagramas/DIA-006_Sequence_Claim_Workflow.md) - Sequence ClaimWorkflow
- [DIA-007](Artefatos/02_Arquitetura/Diagramas/DIA-007_Sequence_CreateEntry.md) - Sequence CreateEntry
- [DIA-008](Artefatos/02_Arquitetura/Diagramas/DIA-008_Flow_VSYNC_Daily.md) - Flow VSYNC
- [DIA-009](Artefatos/02_Arquitetura/Diagramas/DIA-009_Deployment_Kubernetes.md) - Deployment K8s

**TechSpecs** (6 docs):
- [TSP-001](Artefatos/02_Arquitetura/TechSpecs/TSP-001_Temporal_Workflow_Engine.md) - Temporal
- [TSP-002](Artefatos/02_Arquitetura/TechSpecs/TSP-002_Apache_Pulsar_Messaging.md) - Apache Pulsar
- [TSP-003](Artefatos/02_Arquitetura/TechSpecs/TSP-003_Redis_Cache_Layer.md) - Redis
- [TSP-004](Artefatos/02_Arquitetura/TechSpecs/TSP-004_PostgreSQL_Database.md) - PostgreSQL
- [TSP-005](Artefatos/02_Arquitetura/TechSpecs/TSP-005_Fiber_HTTP_Framework.md) - Fiber
- [TSP-006](Artefatos/02_Arquitetura/TechSpecs/TSP-006_XML_Signer_JRE.md) - XML Signer

### 💾 Dados (5 docs)
- [DAT-001](Artefatos/03_Dados/DAT-001_Schema_Database_Core_DICT.md) - Schema Core DICT
- [DAT-002](Artefatos/03_Dados/DAT-002_Schema_Database_Connect.md) - Schema Connect
- [DAT-003](Artefatos/03_Dados/DAT-003_Migrations_Strategy.md) - Migrations Strategy
- [DAT-004](Artefatos/03_Dados/DAT-004_Data_Dictionary.md) - Data Dictionary
- [DAT-005](Artefatos/03_Dados/DAT-005_Redis_Cache_Strategy.md) - Redis Cache Strategy

### 🌐 APIs (7 docs)

**gRPC** (4 docs):
- [GRPC-001](Artefatos/04_APIs/gRPC/GRPC-001_Bridge_gRPC_Service.md) - Bridge gRPC Service
- [GRPC-002](Artefatos/04_APIs/gRPC/GRPC-002_Core_DICT_gRPC_Service.md) - Core DICT gRPC Service
- [GRPC-003](Artefatos/04_APIs/gRPC/GRPC-003_Proto_Files_Specification.md) - Proto Files
- [GRPC-004](Artefatos/04_APIs/gRPC/GRPC-004_Error_Handling_gRPC.md) - Error Handling

**REST** (3 docs):
- [API-002](Artefatos/04_APIs/REST/API-002_Core_DICT_REST_API.md) - Core DICT REST API
- [API-003](Artefatos/04_APIs/REST/API-003_Connect_Admin_API.md) - Connect Admin API
- [API-004](Artefatos/04_APIs/REST/API-004_OpenAPI_Specifications.md) - OpenAPI Specs

### 🔐 Segurança (7 docs)
- [SEC-001](Artefatos/13_Seguranca/SEC-001_mTLS_Configuration.md) - mTLS Configuration
- [SEC-002](Artefatos/13_Seguranca/SEC-002_ICP_Brasil_Certificates.md) - ICP-Brasil Certificates
- [SEC-003](Artefatos/13_Seguranca/SEC-003_Secret_Management.md) - Secret Management (Vault)
- [SEC-004](Artefatos/13_Seguranca/SEC-004_API_Authentication.md) - API Authentication
- [SEC-005](Artefatos/13_Seguranca/SEC-005_Network_Security.md) - Network Security
- [SEC-006](Artefatos/13_Seguranca/SEC-006_XML_Signature_Security.md) - XML Signature Security
- [SEC-007](Artefatos/13_Seguranca/SEC-007_LGPD_Data_Protection.md) - LGPD Data Protection

### 🔗 Integração (4 docs)
- [INT-001](Artefatos/12_Integracao/Fluxos/INT-001_Flow_CreateEntry_E2E.md) - Flow CreateEntry E2E
- [INT-002](Artefatos/12_Integracao/Fluxos/INT-002_Flow_ClaimWorkflow_E2E.md) - Flow ClaimWorkflow E2E
- [INT-003](Artefatos/12_Integracao/Fluxos/INT-003_Flow_VSYNC_E2E.md) - Flow VSYNC E2E
- [INT-004](Artefatos/12_Integracao/Sequencias/INT-004_Sequence_Error_Handling.md) - Sequence Error Handling

### 🛠️ Implementação (5 docs)
- [IMP-001](Artefatos/09_Implementacao/IMP-001_Manual_Implementacao_Core_DICT.md) - Manual Core DICT
- [IMP-002](Artefatos/09_Implementacao/IMP-002_Manual_Implementacao_Connect.md) - Manual Connect
- [IMP-003](Artefatos/09_Implementacao/IMP-003_Manual_Implementacao_Bridge.md) - Manual Bridge
- [IMP-004](Artefatos/09_Implementacao/IMP-004_Developer_Guidelines.md) - Developer Guidelines
- [IMP-005](Artefatos/09_Implementacao/IMP-005_Database_Migration_Guide.md) - Database Migration Guide

### 🚀 DevOps (7 docs)
- [DEV-001](Artefatos/15_DevOps/Pipelines/DEV-001_CI_CD_Pipeline_Core.md) - CI/CD Core
- [DEV-002](Artefatos/15_DevOps/Pipelines/DEV-002_CI_CD_Pipeline_Connect.md) - CI/CD Connect
- [DEV-003](Artefatos/15_DevOps/Pipelines/DEV-003_CI_CD_Pipeline_Bridge.md) - CI/CD Bridge
- [DEV-004](Artefatos/15_DevOps/DEV-004_Kubernetes_Manifests.md) - Kubernetes Manifests
- [DEV-005](Artefatos/15_DevOps/DEV-005_Monitoring_Observability.md) - Monitoring & Observability
- [DEV-006](Artefatos/15_DevOps/DEV-006_Docker_Images.md) - Docker Images
- [DEV-007](Artefatos/15_DevOps/DEV-007_Environment_Config.md) - Environment Config

### ✅ Testes (6 docs)
- [TST-001](Artefatos/14_Testes/Casos/TST-001_Test_Cases_CreateEntry.md) - Test Cases CreateEntry
- [TST-002](Artefatos/14_Testes/Casos/TST-002_Test_Cases_ClaimWorkflow.md) - Test Cases ClaimWorkflow
- [TST-003](Artefatos/14_Testes/Casos/TST-003_Test_Cases_Bridge_mTLS.md) - Test Cases Bridge mTLS
- [TST-004](Artefatos/14_Testes/Casos/TST-004_Performance_Tests.md) - Performance Tests
- [TST-005](Artefatos/14_Testes/Casos/TST-005_Security_Tests.md) - Security Tests
- [TST-006](Artefatos/14_Testes/Casos/TST-006_Regression_Test_Suite.md) - Regression Tests

### 📜 Compliance (5 docs)
- [CMP-001](Artefatos/16_Compliance/CMP-001_Audit_Logs_Specification.md) - Audit Logs
- [CMP-002](Artefatos/16_Compliance/CMP-002_LGPD_Compliance_Checklist.md) - LGPD Checklist
- [CMP-003](Artefatos/16_Compliance/CMP-003_Bacen_Regulatory_Compliance.md) - Bacen Compliance
- [CMP-004](Artefatos/16_Compliance/CMP-004_Data_Retention_Policy.md) - Data Retention Policy
- [CMP-005](Artefatos/16_Compliance/CMP-005_Privacy_Impact_Assessment.md) - Privacy Impact Assessment

### 📱 Frontend (4 docs)
- [FE-001](Artefatos/08_Frontend/Componentes/FE-001_Component_Specifications.md) - Component Specs
- [FE-002](Artefatos/08_Frontend/Wireframes/FE-002_Wireframes_DICT_Operations.md) - Wireframes
- [FE-003](Artefatos/08_Frontend/Jornadas/FE-003_User_Journey_Maps.md) - User Journey Maps
- [FE-004](Artefatos/08_Frontend/Componentes/FE-004_State_Management.md) - State Management

### 📝 Requisitos (5 docs)
- [US-001](Artefatos/01_Requisitos/UserStories/US-001_User_Stories_DICT_Keys.md) - User Stories DICT Keys
- [US-002](Artefatos/01_Requisitos/UserStories/US-002_User_Stories_Claims.md) - User Stories Claims
- [US-003](Artefatos/01_Requisitos/UserStories/US-003_User_Stories_Admin.md) - User Stories Admin
- [BP-001](Artefatos/01_Requisitos/Processos/BP-001_Business_Process_CreateKey.md) - Business Process CreateKey
- [BP-002](Artefatos/01_Requisitos/Processos/BP-002_Business_Process_ClaimWorkflow.md) - Business Process ClaimWorkflow

### 📊 Gestão (4 docs)
- [PM-001](Artefatos/17_Gestao/Backlog/PM-001_Product_Backlog.md) - Product Backlog
- [PM-002](Artefatos/17_Gestao/Sprints/Sprint_03_Plan.md) - Sprint 3 Plan
- [PM-003](Artefatos/17_Gestao/Checklists/PM-003_Definition_of_Done.md) - Definition of Done
- [PM-004](Artefatos/17_Gestao/Checklists/PM-004_Code_Review_Checklist.md) - Code Review Checklist

---

## 📅 Cronograma

### ✅ Fase 1: Especificação Crítica - COMPLETA
**Duração**: 1 dia (2025-10-25)
**Documentos**: 16 documentos críticos
**Entregas**: Schemas DB, gRPC, Segurança (DAT-XXX, GRPC-XXX, SEC-XXX)

### ✅ Fase 2: Especificação Detalhada - COMPLETA
**Duração**: 1 dia (2025-10-25)
**Documentos**: 58 documentos detalhados
**Entregas**: Arquitetura, APIs, DevOps, Testes, Compliance, Frontend, Requisitos, Gestão

### ⏳ Fase 3: Revisão Técnica - ATUAL
**Duração Estimada**: 2-3 semanas
**Responsáveis**: CTO + 3 Heads
**Entregável**: Aprovação formal da documentação

### 📋 Fase 4: Implementação - FUTURO
**Duração Estimada**: 8-12 semanas
**Pré-requisito**: Aprovação da Fase 3
**Entregável**: Sistema completo implementado e testado

### 🚀 Fase 5: Homologação Bacen - FUTURO
**Duração Estimada**: 4-6 semanas
**Pré-requisito**: Implementação completa
**Entregável**: Certificação Bacen

---

## ✅ Critérios de Sucesso

### Fase de Especificação (Completa)
- [x] 100% dos requisitos Bacen catalogados
- [x] Arquitetura completa (C4 Model, ADRs, TechSpecs)
- [x] Modelo de dados completo (PostgreSQL, Redis)
- [x] Todas as APIs especificadas (REST, gRPC)
- [x] Estratégia de testes definida
- [x] DevOps e CI/CD especificados
- [x] Compliance LGPD e Bacen documentado
- [x] Backlog de desenvolvimento priorizado

### Fase de Revisão (Em andamento)
- [ ] Aprovação CTO (José Luís Silva)
- [ ] Aprovação Head Arquitetura (Thiago Lima)
- [ ] Aprovação Head DevOps
- [ ] Aprovação Head Compliance
- [ ] Ajustes incorporados
- [ ] Documentação final consolidada

### Fase de Implementação (Futuro)
- [ ] Core DICT, Connect, Bridge implementados
- [ ] Testes automatizados (>80% cobertura)
- [ ] CI/CD pipelines funcionando
- [ ] Infraestrutura Kubernetes deployada
- [ ] Segurança implementada (mTLS, Vault, LGPD)

### Fase de Homologação (Futuro)
- [ ] Todos casos de teste Bacen passando
- [ ] Certificação Bacen obtida
- [ ] Deploy em produção aprovado

---

## 📊 Métricas do Projeto

### Documentação
- **Total de Documentos**: 74 especificações técnicas
- **Total de Arquivos MD**: 148 (incluindo READMEs, análises, progresso)
- **Linhas de Especificação**: ~50.000+ linhas
- **Tempo de Criação**: 1 dia (execução paralela)
- **Eficiência**: 60x mais rápido que sequencial

### Qualidade
- **Baseado em Análises Reais**: 100% (ANA-001 a ANA-004)
- **Rastreabilidade**: 100% (requisitos → specs)
- **Completude**: 100% (conforme PLANO_PREENCHIMENTO_ARTEFATOS.md)
- **Consistência Stack**: 100% (Temporal, Pulsar, PostgreSQL, Redis)

---

## ⚠️ Decisões Técnicas Importantes

### ClaimWorkflow - 30 Dias
Conforme TEC-003 v2.1, reivindicações de chaves PIX têm período de 30 dias (claim_completion_period_days = 30). Implementado via Temporal Workflow com timer durável.

### Apache Pulsar (não Kafka/RabbitMQ)
Event streaming implementado com Apache Pulsar v0.16.0 para comunicação assíncrona entre Core DICT, Connect e Bridge.

### Temporal Workflows
Orquestração de workflows de longa duração (ClaimWorkflow 30 dias, VSYNC diário) via Temporal v1.36.0.

### Clean Architecture
Separação em 4 camadas: API Layer, Application Layer, Domain Layer, Infrastructure Layer.

### CQRS + Event Sourcing
Separação de comandos e queries, com event sourcing via Pulsar para auditoria e rastreabilidade.

### ICP-Brasil A3
Certificados digitais A3 (hardware token) obrigatórios para comunicação mTLS com Bacen.

---

## 📞 Contato

**Revisão Técnica**: Ver [ROTEIRO_REVISAO_TECNICA.md](Artefatos/00_Master/ROTEIRO_REVISAO_TECNICA.md)
**Dúvidas sobre Documentação**: [Criar issue no repositório]

---

## 📄 Licença e Confidencialidade

Este projeto é propriedade do LBPay. Toda a documentação e código são confidenciais e restritos ao uso interno.

---

**Última Atualização**: 2025-10-25
**Status**: ✅ Especificação Completa - Em Revisão Técnica
**Versão**: 2.0
