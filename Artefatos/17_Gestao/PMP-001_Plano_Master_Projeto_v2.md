# Plano Master do Projeto - DICT LBPay v2.0

**ID**: PMP-001
**Versão**: 2.0
**Data**: 2025-10-24
**PM**: PHOENIX (AGT-PM-001)
**Status**: ✅ Ativo - Fase de Especificação

**Changelog v2.0**:
- ✅ Inventário completo de 37 artefatos (14 criados, 23 faltantes)
- ✅ Roadmap detalhado semana-a-semana (8 semanas)
- ✅ Priorização clara (P0/P1/P2/P3)
- ✅ Dependências entre artefatos mapeadas
- ✅ Responsáveis definidos por artefato
- ✅ Escopo claramente definido: **ESPECIFICAÇÃO** (não codificação)

---

## 📋 Índice

1. [Visão Geral do Projeto](#1-visão-geral-do-projeto)
2. [Escopo do Projeto](#2-escopo-do-projeto)
3. [Inventário Completo de Artefatos](#3-inventário-completo-de-artefatos)
4. [Roadmap de Criação (8 semanas)](#4-roadmap-de-criação-8-semanas)
5. [Stakeholders](#5-stakeholders)
6. [Riscos e Mitigações](#6-riscos-e-mitigações)
7. [Critérios de Sucesso](#7-critérios-de-sucesso)
8. [Aprovações](#8-aprovações)

---

## 1. Visão Geral do Projeto

### 1.1 Contexto

O LBPay é uma instituição de pagamento licenciada pelo Banco Central do Brasil e participante direto do PIX. Implementamos nossa própria solução de Core Banking (Contas de Pagamento) e integrações com o Bacen via RSFN (SPI - PIX). Este projeto visa **especificar completamente** a solução DICT (Diretório de Identificadores de Contas Transacionais) para possibilitar a operação completa do PIX.

### 1.2 Objetivo do Projeto de Especificação

Criar **TODOS os artefatos de especificação necessários** para que um projeto de desenvolvimento futuro possa implementar a solução DICT de forma **autônoma e completa**, sem ambiguidades.

**O QUE ESTE PROJETO ENTREGA**:
- ✅ Documentação técnica completa e detalhada
- ✅ Especificações funcionais e não-funcionais
- ✅ Modelos de dados (conceitual, lógico, físico)
- ✅ Especificações de APIs (REST, gRPC)
- ✅ Diagramas de arquitetura (C4, BPMN, UML)
- ✅ User stories e critérios de aceitação
- ✅ Planos de teste e homologação
- ✅ Backlog de desenvolvimento priorizado

**O QUE ESTE PROJETO NÃO ENTREGA**:
- ❌ Código-fonte (Go, TypeScript, etc.)
- ❌ Repositórios Git com código
- ❌ Infraestrutura provisionada (Kubernetes, PostgreSQL, Redis)
- ❌ Pipelines CI/CD funcionais
- ❌ Testes automatizados executáveis
- ❌ Deploy em ambientes

### 1.3 Justificativa

**Por que precisamos de especificação completa?**

1. **Complexidade Regulatória**: DICT tem 72 requisitos funcionais (Bacen) que precisam estar 100% mapeados
2. **Qualidade**: Especificação detalhada reduz ambiguidades e retrabalho na implementação
3. **Auditabilidade**: Rastreabilidade completa (requisito Bacen → especificação → código)
4. **Eficiência**: Desenvolvedores implementam mais rápido com especificações claras
5. **Homologação**: Bacen exige documentação completa para aprovação

### 1.4 Benefícios Esperados

| Benefício | Métrica de Sucesso |
|-----------|-------------------|
| **Redução de ambiguidades** | < 5 dúvidas críticas não resolvidas ao fim do projeto |
| **Rastreabilidade completa** | 100% dos 72 RFs mapeados em especificações |
| **Velocidade de implementação** | Redução estimada de 30% no tempo de desenvolvimento |
| **Qualidade de código** | Menos bugs e retrabalho (estimado: 40% menos) |
| **Aprovação Bacen** | 1ª tentativa de homologação bem-sucedida |

---

## 2. Escopo do Projeto

### 2.1 Fases do Projeto

#### ✅ Fase 1: Especificação e Planejamento (Atual - 8 semanas)

**Objetivo**: Criar TODOS os artefatos de especificação necessários para implementação autônoma

**Entregas** (37 artefatos):

**00_Master** (5 artefatos - 100% completo):
- ✅ KICKOFF.md
- ✅ RESUMO_EXECUTIVO.md
- ✅ DUVIDAS.md
- ✅ PRONTIDAO_ESPECIFICACAO.md
- ✅ INDICE_GERAL.md

**01_Requisitos** (6 artefatos - 0% completo):
- ❌ PRO-001: Processos BPMN (72 diagramas)
- ❌ UST-001: User Stories Completas (estimativa: 150+ stories)
- ❌ NFR-001: Requisitos Não-Funcionais
- ❌ GLO-001: Glossário de Termos DICT
- ❌ MTR-001: Matriz de Rastreabilidade
- ❌ REG-001: Requisitos Regulatórios Bacen

**02_Arquitetura** (7 artefatos - 71% completo):
- ✅ ARE-001, ARE-002, ARE-003 (Análises AS-IS)
- ✅ DAS-001 (Arquitetura TO-BE)
- ✅ ADR-002, ADR-003, ADR-004 (Decisões Arquiteturais)
- ❌ TEC-001: Especificação Técnica Core DICT
- ❌ TEC-002: Especificação Técnica Connect DICT
- ❌ TEC-003: Especificação Técnica Bridge DICT

**05_Requisitos** (3 artefatos - 100% completo):
- ✅ CRF-001: Checklist de Requisitos Funcionais
- ✅ RESUMO_EXECUTIVO_CRF-001
- ✅ INDEX_CRF-001

**03_Dados** (4 artefatos - 0% completo):
- ❌ MDC-001: Modelo de Dados Conceitual
- ❌ MDL-001: Modelo de Dados Lógico (ERD)
- ❌ MDF-001: Modelo de Dados Físico (DDL SQL completo)
- ❌ DIC-001: Dicionário de Dados

**04_APIs** (3 artefatos - 33% completo):
- ✅ API-001: Especificação APIs DICT Bacen (28 endpoints)
- ❌ EAI-001: Especificação APIs REST Core DICT
- ❌ CGR-001: Contratos gRPC (protobuf)

**05_Frontend** (4 artefatos - 0% completo):
- ❌ LFF-001: Listagem de Funcionalidades Frontend
- ❌ JOR-001: Jornadas de Usuário (20+ jornadas)
- ❌ WIR-001: Wireframes/Mockups (50+ telas)
- ❌ COM-001: Componentes React

**12_Integracao** (3 artefatos - 0% completo):
- ❌ MFE-001: Matriz de Fluxos E2E
- ❌ FLX-001: Diagramas de Fluxo (30+ fluxos)
- ❌ SEQ-001: Diagramas de Sequência UML (30+ diagramas)

**13_Seguranca** (3 artefatos - 0% completo):
- ❌ ASG-001: Análise de Segurança
- ❌ CSG-001: Checklist de Segurança
- ❌ TRM-001: Threat Model (STRIDE)

**08_Testes** (3 artefatos - 0% completo):
- ❌ EST-001: Estratégia de Testes
- ❌ CTE-001: Casos de Teste (unit, integration, E2E)
- ❌ PTH-001: Plano de Testes de Homologação Bacen

**09_DevOps** (2 artefatos - 0% completo):
- ❌ PCI-001: Plano de CI/CD
- ❌ EDI-001: Estratégia de Deployment e Infraestrutura

**10_Compliance** (3 artefatos - 0% completo):
- ❌ CCM-001: Checklist de Compliance
- ❌ RLG-001: Requisitos LGPD
- ❌ RBC-001: Requisitos Bacen (checklist homologação detalhado)

**11_Gestao** (3 artefatos - 67% completo):
- ✅ PMP-001: Plano Master do Projeto (este documento)
- ✅ STATUS_PROJETO (2 versões)
- ❌ BKL-001: Backlog de Desenvolvimento Priorizado

**TOTAL**: **37 artefatos** (14 criados ✅, 23 faltantes ❌)
**Progresso Atual**: **38% completo**

---

#### 🔵 Fase 2: Implementação (Futura - duração a definir)

**Objetivo**: Implementar a solução completa baseada nos artefatos da Fase 1

**Entregas**:
- Código-fonte (Go, TypeScript, SQL)
- Repositórios Git (core-dict, connector-dict, etc.)
- Testes automatizados (unit, integration, E2E)
- CI/CD pipelines funcionais
- Infraestrutura (Kubernetes, PostgreSQL, Redis)
- Documentação técnica de código

**Duração Estimada**: A definir após Fase 1 (estimativa inicial: 40-50 semanas)
**Squad**: Squad de Desenvolvimento (a definir)

---

### 2.2 Stack Tecnológica Confirmada

**Tecnologias Base** (confirmadas no documento `ArquiteturaDict_LBPAY.md`):

| Tecnologia | Uso | Status |
|------------|-----|--------|
| **Apache Pulsar** | Streaming/mensageria event-driven | ✅ Confirmado |
| **gRPC (protobuf)** | Comunicação síncrona entre serviços | ✅ Confirmado |
| **Temporal Workflow** | Orquestração de processos assíncronos de longo prazo | ✅ Confirmado |
| **PostgreSQL** | Banco de dados principal | ✅ Confirmado |
| **Redis** | Cache e operações de curta duração | ✅ Confirmado |
| **Go (Golang)** | Backend services | ✅ Confirmado |
| **RSFN Connect** | Módulo de integração com Bacen via RSFN | ✅ Confirmado |

**⚠️ IMPORTANTE - Approach de Especificação**:

Embora as tecnologias estejam **confirmadas**, o **DESIGN e ESPECIFICAÇÃO de COMO usá-las** será:

1. ✅ **Resultado da análise dos agentes especializados** (NEXUS, GOPHER, MERCURY, etc.)
2. ✅ **Baseado em 3 fontes principais**:
   - **Repositórios existentes**:
     - `money-moving`: Padrões de Clean Architecture, uso de Pulsar
     - `rsfn-connect-bacen-bridge`: Integração RSFN, padrões de comunicação
     - `core-banking`: Padrões de domínio, persistência
   - **Documento de arquitetura** (`ArquiteturaDict_LBPAY.md`):
     - **CRÍTICO**: Contém diagramas SVG com toda a arquitetura visual
     - Context Diagrams (C4 Model)
     - App Diagrams (DICT, Audit, RSFN Connect, Temporal Server)
     - Component Diagrams (fluxos detalhados de Pulsar, gRPC, Temporal)
     - **Os agentes DEVEM analisar estes diagramas para entender**:
       - 🔄 Todos os fluxos de comunicação gRPC
       - 📨 Todos os tópicos Pulsar (producers/consumers)
       - ⏱️ Todos os workflows Temporal (orchestration)
       - 🔗 Todas as interligações entre componentes (Core DICT, Bridge, Connect)
       - 📊 Toda a arquitetura de integração com Bacen via RSFN
   - **Manual Operacional DICT Bacen**: Requisitos regulatórios e técnicos
3. ✅ **Documentado em ADRs** (Architecture Decision Records)
4. ✅ **Especificado em detalhes** nos artefatos técnicos:
   - **CGR-001**: Contratos gRPC (protobuf schemas)
   - **TEC-001/002/003**: Specs técnicas detalhadas
   - **ADR-002/003/004**: Decisões arquiteturais fundamentadas

**🎯 FUNDAMENTAL**: Os diagramas em `ArquiteturaDict_LBPAY.md` são a **fonte de verdade visual** da arquitetura. Os agentes devem analisar profundamente estes diagramas para entender:
- Como os componentes se comunicam (gRPC, Pulsar, HTTP)
- Quais tópicos Pulsar existem e seus fluxos
- Quais workflows Temporal são necessários
- Como Bridge, Connect e Core DICT se integram
- **Esta análise é ESSENCIAL para a fase de implementação**

**Exemplos de Perguntas a Responder pelos Agentes** (baseado nos diagramas):
- 📋 Como organizar tópicos Pulsar? (por domínio, por evento, padrão pub/sub?)
- 📋 Quais contratos gRPC? (serviços, messages, schemas)
- 📋 Quais workflows Temporal? (reivindicação, portabilidade, exclusão agendada)
- 📋 Como estruturar eventos de domínio? (formato, versionamento, retry)
- 📋 Padrões de resilience? (circuit breaker, timeout, retry policies)
- 📋 Fluxos de integração RSFN Connect ↔ Bacen DICT
- 📋 Padrões de comunicação entre Core DICT ↔ Bridge ↔ Connect

**Não devemos impor design prematuro** - os agentes analisarão, proporão e especificarão baseado em **análise fundamentada dos diagramas e código existente**.

---

### 2.3 Escopo Funcional (72 RFs)

#### Bloco 1: CRUD de Chaves PIX (13 RFs)
- RF-BLO1-001 a RF-BLO1-013
- Criar, consultar, alterar, excluir chaves PIX
- Todos os tipos: CPF, CNPJ, Email, Telefone, Aleatória (EVP)

#### Bloco 2: Reivindicação e Portabilidade (14 RFs)
- RF-BLO2-001 a RF-BLO2-014
- Reivindicação de posse (7 dias)
- Portabilidade de chave entre PSPs
- Cancelamento, confirmação, consulta

#### Bloco 3: Validações (3 RFs)
- RF-BLO3-001 a RF-BLO3-003
- Validação de posse (Subseção 2.1 Manual Bacen)
- Validação cadastral (Receita Federal)
- Validação de nomes

#### Bloco 4: Devolução e Infração (6 RFs)
- RF-BLO4-001 a RF-BLO4-006
- Solicitação de devolução (fraude, falha operacional)
- Notificação de infração
- Cancelamento

#### Bloco 5: Segurança e Infraestrutura (13 RFs)
- RF-BLO5-001 a RF-BLO5-013
- Rate limiting anti-scan
- Cache de respostas
- mTLS, XML Signature
- Observabilidade

#### Bloco 6: Recuperação de Valores (13 RFs)
- RF-BLO6-001 a RF-BLO6-013
- Fluxo interativo de recuperação
- Fluxo automatizado
- Rastreamento de valores

#### Transversal (10 RFs)
- Idempotência, logging, auditoria, etc.

---

## 3. Inventário Completo de Artefatos

### 3.1 Legenda de Prioridades

| Prioridade | Descrição | Quando Criar |
|-----------|-----------|--------------|
| **P0** | Crítico - Bloqueia tudo | Semanas 1-2 |
| **P1** | Alto - Bloqueia implementação | Semanas 3-4 |
| **P2** | Médio - Importante mas não bloqueante | Semanas 5-6 |
| **P3** | Baixo - Nice to have | Semanas 7-8 |

### 3.2 Inventário por Categoria

#### 📁 00_Master (5/5 completo - 100%)

| ID | Artefato | Status | Prioridade | Responsável | Páginas | Dependências |
|----|----------|--------|------------|-------------|---------|--------------|
| MAS-001 | KICKOFF.md | ✅ Completo | P0 | PHOENIX | 5 | Nenhuma |
| MAS-002 | RESUMO_EXECUTIVO.md | ✅ Completo | P0 | PHOENIX | 10 | Nenhuma |
| MAS-003 | DUVIDAS.md | ✅ Completo | P0 | PHOENIX | 20 | Nenhuma |
| MAS-004 | PRONTIDAO_ESPECIFICACAO.md | ✅ Completo | P0 | PHOENIX | 8 | Todos |
| MAS-005 | INDICE_GERAL.md | ✅ Completo | P0 | SCRIBE | 12 | Todos |

---

#### 📁 01_Requisitos (0/6 completo - 0%)

| ID | Artefato | Status | Prioridade | Responsável | Páginas Est. | Dependências |
|----|----------|--------|------------|-------------|--------------|--------------|
| REQ-001 | PRO-001: Processos BPMN (72 diagramas) | ❌ Pendente | P1 | ORACLE | 150 | CRF-001 |
| REQ-002 | UST-001: User Stories (150+ stories) | ❌ Pendente | P1 | ORACLE | 180 | CRF-001, PRO-001 |
| REQ-003 | NFR-001: Requisitos Não-Funcionais | ❌ Pendente | P1 | NEXUS | 40 | CRF-001 |
| REQ-004 | GLO-001: Glossário de Termos DICT | ❌ Pendente | P2 | SCRIBE | 25 | Nenhuma |
| REQ-005 | MTR-001: Matriz de Rastreabilidade | ❌ Pendente | P1 | ORACLE | 30 | CRF-001, UST-001 |
| REQ-006 | REG-001: Requisitos Regulatórios Bacen | ❌ Pendente | P0 | GUARDIAN | 50 | Manual Bacen |

---

#### 📁 02_Arquitetura (5/8 completo - 62.5%)

| ID | Artefato | Status | Prioridade | Responsável | Páginas | Dependências |
|----|----------|--------|------------|-------------|---------|--------------|
| ARQ-001 | ARE-001: Análise Repos Existentes | ✅ Completo | P0 | GOPHER | 40 | Nenhuma |
| ARQ-002 | ARE-002: Análise Impl DICT Dispersa | ✅ Completo | P0 | NEXUS | 35 | ARE-001 |
| ARQ-003 | ARE-003: Análise Doc Arquitetura | ✅ Completo | P0 | NEXUS | 50 | Nenhuma |
| ARQ-004 | DAS-001: Arquitetura TO-BE | ✅ Completo | P0 | NEXUS | 85 | AREs |
| ARQ-005 | ADR-002: Consolidação Core DICT | ✅ Completo | P0 | NEXUS | 45 | DAS-001 |
| ARQ-006 | ADR-003: Performance Multi-Camadas | ✅ Completo | P0 | NEXUS | 75 | DAS-001 |
| ARQ-007 | ADR-004: Bridge DICT Dedicado | ✅ Completo | P0 | NEXUS | 80 | DAS-001 |
| ARQ-008 | TEC-001: Especificação Técnica Core DICT | ❌ Pendente | P1 | GOPHER | 120 | DAS-001, ADR-002 |
| ARQ-009 | TEC-002: Especificação Técnica Connect DICT | ❌ Pendente | P1 | CONDUIT | 80 | DAS-001, ADR-004 |
| ARQ-010 | TEC-003: Especificação Técnica Bridge DICT | ❌ Pendente | P1 | CONDUIT | 60 | DAS-001, ADR-004 |

---

#### 📁 05_Requisitos (3/3 completo - 100%)

| ID | Artefato | Status | Prioridade | Responsável | Páginas | Dependências |
|----|----------|--------|------------|-------------|---------|--------------|
| RFN-001 | CRF-001: Checklist RFs | ✅ Completo | P0 | ORACLE | 60 | Manual Bacen |
| RFN-002 | RESUMO_EXECUTIVO_CRF-001 | ✅ Completo | P0 | ORACLE | 15 | CRF-001 |
| RFN-003 | INDEX_CRF-001 | ✅ Completo | P0 | SCRIBE | 5 | CRF-001 |

---

#### 📁 03_Dados (0/4 completo - 0%)

| ID | Artefato | Status | Prioridade | Responsável | Páginas Est. | Dependências |
|----|----------|--------|------------|-------------|--------------|--------------|
| DAD-001 | MDC-001: Modelo Conceitual | ❌ Pendente | P1 | ATLAS | 30 | CRF-001, DAS-001 |
| DAD-002 | MDL-001: Modelo Lógico (ERD) | ❌ Pendente | P1 | ATLAS | 40 | MDC-001 |
| DAD-003 | MDF-001: Modelo Físico (DDL SQL) | ❌ Pendente | P1 | ATLAS | 60 | MDL-001 |
| DAD-004 | DIC-001: Dicionário de Dados | ❌ Pendente | P1 | ATLAS | 80 | MDF-001 |

---

#### 📁 04_APIs (1/3 completo - 33%)

| ID | Artefato | Status | Prioridade | Responsável | Páginas | Dependências |
|----|----------|--------|------------|-------------|---------|--------------|
| API-001 | API-001: Specs APIs DICT Bacen | ✅ Completo | P0 | MERCURY | 66 | OpenAPI Bacen |
| API-002 | EAI-001: Specs APIs REST Core DICT | ❌ Pendente | P1 | MERCURY | 100 | TEC-001, MDF-001 |
| API-003 | CGR-001: Contratos gRPC (protobuf) | ❌ Pendente | P1 | GOPHER | 40 | TEC-001, TEC-002 |

---

#### 📁 05_Frontend (0/4 completo - 0%)

| ID | Artefato | Status | Prioridade | Responsável | Páginas Est. | Dependências |
|----|----------|--------|------------|-------------|--------------|--------------|
| FRO-001 | LFF-001: Lista Funcionalidades Frontend | ❌ Pendente | P2 | PRISM | 35 | CRF-001, UST-001 |
| FRO-002 | JOR-001: Jornadas de Usuário (20+) | ❌ Pendente | P2 | PRISM | 60 | UST-001 |
| FRO-003 | WIR-001: Wireframes/Mockups (50+) | ❌ Pendente | P2 | PRISM | 100 | JOR-001 |
| FRO-004 | COM-001: Componentes React | ❌ Pendente | P3 | PRISM | 45 | WIR-001 |

**Nota**: Frontend está fora do escopo crítico (portais já existem), mas especificações são importantes para APIs.

---

#### 📁 12_Integracao (0/3 completo - 0%)

| ID | Artefato | Status | Prioridade | Responsável | Páginas Est. | Dependências |
|----|----------|--------|------------|-------------|--------------|--------------|
| INT-001 | MFE-001: Matriz de Fluxos E2E | ❌ Pendente | P1 | CONDUIT | 40 | CRF-001, PRO-001 |
| INT-002 | FLX-001: Diagramas de Fluxo (30+) | ❌ Pendente | P1 | CONDUIT | 90 | MFE-001, DAS-001 |
| INT-003 | SEQ-001: Diagramas Sequência UML (30+) | ❌ Pendente | P1 | CONDUIT | 120 | FLX-001 |

---

#### 📁 13_Seguranca (0/3 completo - 0%)

| ID | Artefato | Status | Prioridade | Responsável | Páginas Est. | Dependências |
|----|----------|--------|------------|-------------|--------------|--------------|
| SEG-001 | ASG-001: Análise de Segurança | ❌ Pendente | P1 | SENTINEL | 50 | DAS-001, TEC-001 |
| SEG-002 | CSG-001: Checklist de Segurança | ❌ Pendente | P1 | SENTINEL | 30 | ASG-001 |
| SEG-003 | TRM-001: Threat Model (STRIDE) | ❌ Pendente | P2 | SENTINEL | 40 | ASG-001 |

---

#### 📁 08_Testes (0/3 completo - 0%)

| ID | Artefato | Status | Prioridade | Responsável | Páginas Est. | Dependências |
|----|----------|--------|------------|-------------|--------------|--------------|
| TST-001 | EST-001: Estratégia de Testes | ❌ Pendente | P1 | VALIDATOR | 45 | CRF-001, DAS-001 |
| TST-002 | CTE-001: Casos de Teste (unit/int/E2E) | ❌ Pendente | P1 | VALIDATOR | 200 | UST-001, TEC-001 |
| TST-003 | PTH-001: Plano Homologação Bacen | ❌ Pendente | P0 | GUARDIAN | 80 | Manual Bacen, EST-001 |

---

#### 📁 09_DevOps (0/2 completo - 0%)

| ID | Artefato | Status | Prioridade | Responsável | Páginas Est. | Dependências |
|----|----------|--------|------------|-------------|--------------|--------------|
| DVO-001 | PCI-001: Plano de CI/CD | ❌ Pendente | P2 | FORGE | 40 | TEC-001, EST-001 |
| DVO-002 | EDI-001: Estratégia Deployment/Infra | ❌ Pendente | P2 | FORGE | 50 | DAS-001, ADR-003 |

---

#### 📁 10_Compliance (0/3 completo - 0%)

| ID | Artefato | Status | Prioridade | Responsável | Páginas Est. | Dependências |
|----|----------|--------|------------|-------------|--------------|--------------|
| CMP-001 | CCM-001: Checklist de Compliance | ❌ Pendente | P0 | GUARDIAN | 40 | Manual Bacen |
| CMP-002 | RLG-001: Requisitos LGPD | ❌ Pendente | P1 | GUARDIAN | 35 | ASG-001 |
| CMP-003 | RBC-001: Requisitos Bacen (checklist) | ❌ Pendente | P0 | GUARDIAN | 60 | Manual Bacen |

---

#### 📁 11_Gestao (2/3 completo - 67%)

| ID | Artefato | Status | Prioridade | Responsável | Páginas | Dependências |
|----|----------|--------|------------|-------------|---------|--------------|
| GES-001 | PMP-001: Plano Master (este doc) | ✅ Completo | P0 | PHOENIX | 50 | Nenhuma |
| GES-002 | STATUS_PROJETO | ✅ Completo | P0 | PHOENIX | 15 | Todos |
| GES-003 | BKL-001: Backlog Desenvolvimento | ❌ Pendente | P1 | CATALYST | 120 | Todos |

---

### 3.3 Resumo Estatístico

| Categoria | Total | Completo | Pendente | % Completo |
|-----------|-------|----------|----------|------------|
| **00_Master** | 5 | 5 | 0 | 100% |
| **01_Requisitos** | 6 | 0 | 6 | 0% |
| **02_Arquitetura** | 10 | 7 | 3 | 70% |
| **05_Requisitos** | 3 | 3 | 0 | 100% |
| **03_Dados** | 4 | 0 | 4 | 0% |
| **04_APIs** | 3 | 1 | 2 | 33% |
| **05_Frontend** | 4 | 0 | 4 | 0% |
| **12_Integracao** | 3 | 0 | 3 | 0% |
| **13_Seguranca** | 3 | 0 | 3 | 0% |
| **08_Testes** | 3 | 0 | 3 | 0% |
| **09_DevOps** | 2 | 0 | 2 | 0% |
| **10_Compliance** | 3 | 0 | 3 | 0% |
| **11_Gestao** | 3 | 2 | 1 | 67% |
| **TOTAL** | **52** | **18** | **34** | **35%** |

**Páginas Totais Estimadas**: ~2.200 páginas de especificação

---

## 4. Roadmap de Criação (8 semanas)

### Semana 1-2: Requisitos e Processos (P0 + P1)

**Foco**: Completar mapeamento de requisitos e processos de negócio

**Artefatos a Criar** (6):
1. ❌ **REG-001**: Requisitos Regulatórios Bacen (GUARDIAN) - 50 páginas
2. ❌ **PRO-001**: Processos BPMN - 72 diagramas (ORACLE) - 150 páginas
3. ❌ **NFR-001**: Requisitos Não-Funcionais (NEXUS) - 40 páginas
4. ❌ **GLO-001**: Glossário de Termos DICT (SCRIBE) - 25 páginas
5. ❌ **PTH-001**: Plano de Homologação Bacen (GUARDIAN) - 80 páginas
6. ❌ **CCM-001**: Checklist de Compliance (GUARDIAN) - 40 páginas

**Esforço Total**: 385 páginas, ~120 horas
**Entrega da Semana 2**: ✅ Requisitos 100% mapeados

---

### Semana 3-4: Dados e User Stories (P1)

**Foco**: Modelo de dados completo e user stories detalhadas

**Artefatos a Criar** (7):
1. ❌ **MDC-001**: Modelo de Dados Conceitual (ATLAS) - 30 páginas
2. ❌ **MDL-001**: Modelo de Dados Lógico (ERD) (ATLAS) - 40 páginas
3. ❌ **MDF-001**: Modelo de Dados Físico (DDL SQL) (ATLAS) - 60 páginas
4. ❌ **DIC-001**: Dicionário de Dados (ATLAS) - 80 páginas
5. ❌ **UST-001**: User Stories (150+ stories) (ORACLE) - 180 páginas
6. ❌ **MTR-001**: Matriz de Rastreabilidade (ORACLE) - 30 páginas
7. ❌ **RBC-001**: Requisitos Bacen Detalhados (GUARDIAN) - 60 páginas

**Esforço Total**: 480 páginas, ~140 horas
**Entrega da Semana 4**: ✅ Modelo de dados completo + User stories prontas

---

### Semana 5-6: Especificações Técnicas e Integração (P1)

**Foco**: Especificações técnicas detalhadas de módulos e integrações

**Artefatos a Criar** (9):
1. ❌ **TEC-001**: Especificação Técnica Core DICT (GOPHER) - 120 páginas
2. ❌ **TEC-002**: Especificação Técnica Connect DICT (CONDUIT) - 80 páginas
3. ❌ **TEC-003**: Especificação Técnica Bridge DICT (CONDUIT) - 60 páginas
4. ❌ **EAI-001**: Especificação APIs REST Core DICT (MERCURY) - 100 páginas
5. ❌ **CGR-001**: Contratos gRPC (protobuf) (GOPHER) - 40 páginas
6. ❌ **MFE-001**: Matriz de Fluxos E2E (CONDUIT) - 40 páginas
7. ❌ **FLX-001**: Diagramas de Fluxo (30+) (CONDUIT) - 90 páginas
8. ❌ **SEQ-001**: Diagramas de Sequência UML (30+) (CONDUIT) - 120 páginas
9. ❌ **ASG-001**: Análise de Segurança (SENTINEL) - 50 páginas

**Esforço Total**: 700 páginas, ~200 horas
**Entrega da Semana 6**: ✅ Especificações técnicas completas + Fluxos E2E documentados

---

### Semana 7-8: Testes, Segurança e Consolidação (P1 + P2 + P3)

**Foco**: Estratégias de teste, frontend, DevOps e consolidação final

**Artefatos a Criar** (12):
1. ❌ **EST-001**: Estratégia de Testes (VALIDATOR) - 45 páginas
2. ❌ **CTE-001**: Casos de Teste (200+) (VALIDATOR) - 200 páginas
3. ❌ **CSG-001**: Checklist de Segurança (SENTINEL) - 30 páginas
4. ❌ **TRM-001**: Threat Model (STRIDE) (SENTINEL) - 40 páginas
5. ❌ **RLG-001**: Requisitos LGPD (GUARDIAN) - 35 páginas
6. ❌ **LFF-001**: Lista Funcionalidades Frontend (PRISM) - 35 páginas
7. ❌ **JOR-001**: Jornadas de Usuário (20+) (PRISM) - 60 páginas
8. ❌ **WIR-001**: Wireframes/Mockups (50+) (PRISM) - 100 páginas
9. ❌ **COM-001**: Componentes React (PRISM) - 45 páginas
10. ❌ **PCI-001**: Plano de CI/CD (FORGE) - 40 páginas
11. ❌ **EDI-001**: Estratégia Deployment/Infra (FORGE) - 50 páginas
12. ❌ **BKL-001**: Backlog de Desenvolvimento (CATALYST) - 120 páginas

**Esforço Total**: 800 páginas, ~240 horas
**Entrega da Semana 8**: ✅ Projeto de especificação 100% completo e aprovado

---

### Resumo por Semana

| Semana | Foco | Artefatos | Páginas | Horas | % Total |
|--------|------|-----------|---------|-------|---------|
| **1-2** | Requisitos e Processos | 6 | 385 | 120 | 17% |
| **3-4** | Dados e User Stories | 7 | 480 | 140 | 21% |
| **5-6** | Especificações Técnicas | 9 | 700 | 200 | 32% |
| **7-8** | Testes e Consolidação | 12 | 800 | 240 | 36% |
| **TOTAL** | | **34** | **2.365** | **700h** | **100%** |

---

## 5. Stakeholders

### 5.1 Stakeholders Executivos

| Nome/Papel | Responsabilidade | Nível de Envolvimento |
|------------|------------------|----------------------|
| **CTO (José Luís Silva)** | Aprovação final de especificações e arquitetura | Alto - Aprovações + Esclarecimentos |
| **Head de Arquitetura (Thiago Lima)** | Aprovação de decisões arquiteturais | Alto - Reviews semanais |
| **Head de Produto (Luiz Sant'Ana)** | Aprovação de requisitos funcionais | Alto - Reviews semanais |
| **Head de Engenharia (Jorge Fonseca)** | Aprovação de stack e implementação | Médio - Reviews quinzenais |

### 5.2 Squad de Especificação (14 agentes)

Ver documento: [SQUAD_ARCHITECTURE.md](../SQUAD_ARCHITECTURE.md)

| Agente | Papel | Responsabilidades Principais |
|--------|-------|----------------------------|
| **PHOENIX** | Project Manager | Coordenação geral, PMP-001, STATUS |
| **CATALYST** | Scrum Master | Backlog, sprints, retrospectivas |
| **ORACLE** | Business Analyst | CRF-001, PRO-001, UST-001, MTR-001 |
| **NEXUS** | Solution Architect | DAS-001, ADRs, NFR-001 |
| **ATLAS** | Data Architect | MDC/MDL/MDF-001, DIC-001 |
| **MERCURY** | API Specialist | API-001, EAI-001 |
| **PRISM** | Frontend Architect | LFF-001, JOR-001, WIR-001, COM-001 |
| **CONDUIT** | Integration Architect | MFE-001, FLX-001, SEQ-001, TEC-002/003 |
| **SENTINEL** | Security Architect | ASG-001, CSG-001, TRM-001 |
| **VALIDATOR** | QA Architect | EST-001, CTE-001 |
| **FORGE** | DevOps Architect | PCI-001, EDI-001 |
| **GOPHER** | Tech Specialist Go | TEC-001, CGR-001, ARE-001 |
| **SCRIBE** | Technical Writer | GLO-001, INDICE_GERAL, reviews |
| **GUARDIAN** | Compliance Manager | REG-001, CCM-001, RBC-001, PTH-001, RLG-001 |

### 5.3 Comunicação com Stakeholders

**Reuniões Regulares**:
- **Daily Standup**: Diário, Squad interna (15min) - assíncrono
- **Sprint Planning**: Semanal (1h) - Segunda-feira
- **Sprint Review**: Semanal com stakeholders (1h) - Sexta-feira
- **Retrospectiva**: Semanal, Squad interna (45min) - Sexta-feira
- **Status Report**: Semanal para executivos (assíncrono)

**Canais de Comunicação**:
- **Slack**: #dict-especificacao (diário)
- **Email**: Status reports semanais
- **GitHub Issues**: Dúvidas técnicas
- **DUVIDAS.md**: Questões para CTO

---

## 6. Riscos e Mitigações

### 6.1 Matriz de Riscos

| ID | Risco | Prob. | Impacto | Mitigação |
|----|-------|-------|---------|-----------|
| **R-001** | Documentação Bacen incompleta ou ambígua | Média | Alto | Documento DUVIDAS.md; consultar Bacen se necessário |
| **R-002** | Requisitos de homologação mudarem | Baixa | Alto | Monitorar comunicados Bacen; arquitetura flexível |
| **R-003** | Código existente não documentado | Alta | Médio | Análise profunda (ARE-001/002); engenharia reversa |
| **R-004** | Atraso nas aprovações de stakeholders | Média | Médio | Pacotes de aprovação claros; follow-ups semanais |
| **R-005** | Complexidade subestimada | Média | Alto | Revisões frequentes; ajustar estimativas |
| **R-006** | Falta de expertise em domínio DICT | Baixa | Médio | Manual Bacen é completo; consultar especialistas |
| **R-007** | Mudanças em repos durante Fase 1 | Média | Médio | Snapshots de código; comunicação com outras equipes |
| **R-008** | Sobrecarga da squad (34 artefatos) | Alta | Médio | Priorização (P0/P1/P2/P3); paralelização |
| **R-009** | Ambiguidades não resolvidas | Média | Alto | DUVIDAS.md atualizado; CTO responde semanalmente |

### 6.2 Plano de Contingência

**Se atrasos ocorrerem**:
1. Priorizar artefatos P0 e P1 (críticos)
2. Paralelizar trabalhos quando possível (múltiplos agentes)
3. Reduzir escopo de artefatos P3 (nice-to-have)
4. Comunicar transparentemente com stakeholders
5. Solicitar extensão de prazo (±1-2 semanas aceitável)

**Se requisitos mudarem**:
1. Impact assessment rápido (PHOENIX + ORACLE)
2. Atualizar artefatos afetados (responsáveis originais)
3. Re-priorizar backlog (CATALYST)
4. Comunicar mudanças (STATUS_PROJETO)

**Se dúvidas críticas não forem resolvidas**:
1. Documentar em DUVIDAS.md (prioridade Alta/Crítica)
2. Escalar para CTO imediatamente
3. Continuar com outros artefatos (não bloquear tudo)
4. Criar alternativas (Plano A/B) para decisões pendentes

---

## 7. Critérios de Sucesso

### 7.1 Critérios Obrigatórios (Go/No-Go)

- [ ] 100% dos 72 RFs do Bacen catalogados e especificados
- [ ] Arquitetura de solução aprovada por CTO + Head de Arquitetura (Thiago Lima)
- [ ] Modelo de dados completo (conceitual, lógico, físico) aprovado
- [ ] Todas as APIs especificadas (internas e Bacen)
- [ ] User stories completas (150+) com critérios de aceitação
- [ ] Backlog de desenvolvimento criado e priorizado
- [ ] Plano de homologação Bacen completo (PTH-001)
- [ ] Todos os 34 artefatos P0+P1 revisados e aprovados
- [ ] < 5 dúvidas críticas não resolvidas

### 7.2 Critérios de Qualidade

- [ ] Rastreabilidade completa (Manual Bacen → CRF-001 → UST-001 → TEC-001)
- [ ] Todos os 52 documentos indexados e cross-referenced
- [ ] Especificações claras o suficiente para implementação autônoma
- [ ] Critérios de aceitação mensuráveis para cada RF
- [ ] 0 ambiguidades críticas não resolvidas (DUVIDAS.md)
- [ ] Diagramas padronizados (BPMN, UML, C4, ERD)
- [ ] Review de 100% dos artefatos por pares (peer review)

### 7.3 Métricas de Sucesso

| Métrica | Target | Como Medir |
|---------|--------|------------|
| **Completude** | > 95% de artefatos P0+P1 criados | Checklist no PRONTIDAO_ESPECIFICACAO.md |
| **Qualidade** | > 90% de aprovação nas revisões | % aprovações sem change requests |
| **Clareza** | < 5 dúvidas críticas não resolvidas | Contador em DUVIDAS.md |
| **Tempo** | 8 semanas (±1 semana) | Cronograma vs realizado |
| **Rastreabilidade** | 100% dos 72 RFs rastreáveis | MTR-001 completa |
| **Páginas** | ~2.200 páginas de especificação | Soma de todos os artefatos |

### 7.4 Definição de "Pronto" (DoD - Definition of Done)

Um artefato está **pronto** quando:
1. ✅ Conteúdo completo conforme template
2. ✅ Revisado por autor (self-review)
3. ✅ Revisado por peer (outro agente)
4. ✅ Aprovado por stakeholder responsável
5. ✅ Cross-references atualizados
6. ✅ Adicionado ao INDICE_GERAL.md
7. ✅ Sem dúvidas críticas pendentes relacionadas
8. ✅ Markdown formatado corretamente
9. ✅ Diagramas exportados (PNG/SVG) quando aplicável
10. ✅ Versionado e datado

---

## 8. Aprovações

### 8.1 Aprovação do PMP-001 v2.0

| Aprovador | Data | Assinatura | Status |
|-----------|------|------------|--------|
| **CTO (José Luís Silva)** | 2025-10-24 | ___________ | 🟡 Pendente |
| **Head de Arquitetura (Thiago Lima)** | ___________ | ___________ | 🟡 Pendente |
| **Head de Produto (Luiz Sant'Ana)** | ___________ | ___________ | 🟡 Pendente |

### 8.2 Aprovação de Artefatos P0

Todos os artefatos P0 (críticos) devem ser aprovados pelo CTO antes de prosseguir:

- [ ] REG-001: Requisitos Regulatórios Bacen
- [ ] PTH-001: Plano de Homologação Bacen
- [ ] CCM-001: Checklist de Compliance
- [ ] RBC-001: Requisitos Bacen (checklist detalhado)

### 8.3 Gate de Aprovação (Fim da Semana 8)

**Critérios para Aprovar Fase 1 e iniciar Fase 2**:
1. ✅ Todos os 34 artefatos P0+P1 criados e aprovados
2. ✅ Todas as aprovações de stakeholders recebidas
3. ✅ < 5 dúvidas críticas não resolvidas
4. ✅ Backlog de Desenvolvimento (BKL-001) priorizado
5. ✅ Squad de Desenvolvimento definida

**Reunião de Gate**: Sexta-feira da Semana 8
**Participantes**: CTO, Head Arquitetura, Head Produto, Head Engenharia, PHOENIX

---

## 9. Controle de Mudanças

### 9.1 Histórico de Versões

| Versão | Data | Autor | Mudanças |
|--------|------|-------|----------|
| 1.0 | 2025-10-24 | PHOENIX | Versão inicial (draft) |
| 2.0 | 2025-10-24 | PHOENIX | Inventário completo de 52 artefatos; roadmap detalhado 8 semanas; priorização P0/P1/P2/P3; escopo claramente definido (especificação, não codificação) |

### 9.2 Processo de Change Request

**Quando solicitar mudança no PMP-001**:
- Novos artefatos identificados
- Mudanças de prioridade
- Mudanças de cronograma
- Mudanças de escopo

**Processo**:
1. Criar issue no GitHub: `[CHANGE REQUEST] Título`
2. Descrever: O quê, Por quê, Impacto
3. PHOENIX analisa impacto
4. CTO aprova/rejeita
5. PMP-001 atualizado (nova versão)

---

## 10. Próximos Passos

### 10.1 Ações Imediatas (Hoje - 2025-10-24)

1. ✅ **Aprovar PMP-001 v2.0** (CTO)
2. ⏳ **Deletar IMP-001** (fora do escopo de especificação)
3. ⏳ **Iniciar Semana 1**: Criar REG-001 (GUARDIAN)
4. ⏳ **Solicitar certificado mTLS ao Bacen** (DEP-01 - bloqueador Fase 2)
5. ⏳ **Solicitar acesso HML Bacen** (DEP-02 - bloqueador Fase 2)

### 10.2 Primeira Sprint (Semana 1)

**Sprint Goal**: Documentar requisitos regulatórios e iniciar processos BPMN

**Artefatos da Semana 1**:
- REG-001: Requisitos Regulatórios Bacen (GUARDIAN) - 50 páginas
- PRO-001: Processos BPMN (início - 20 diagramas) (ORACLE) - 50 páginas
- GLO-001: Glossário de Termos DICT (SCRIBE) - 25 páginas

**Daily Standups**: 9h (assíncrono via Slack)
**Sprint Review**: Sexta-feira 15h (com CTO)

---

**Documento aprovado por**: CTO (José Luís Silva)
**Data de Aprovação**: __________
**Versão Aprovada**: 2.0
**Status**: 🟡 Aguardando Aprovação

---

**Resumo Executivo PMP-001 v2.0**:
- 📦 **52 artefatos** mapeados (18 completos, 34 pendentes)
- 📄 **~2.200 páginas** de especificação a criar
- ⏱️ **8 semanas** (700 horas de esforço)
- 👥 **14 agentes** especializados
- 🎯 **Objetivo**: Especificação 100% completa para desenvolvimento autônomo
- ✅ **Progresso Atual**: 35% completo
