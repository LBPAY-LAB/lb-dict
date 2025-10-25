# Squad de Arquitetura - Projeto DICT LBPay

## Código da Squad: SQARCH-001

---

## Missão da Squad
Responsável pela idealização, especificação detalhada e planejamento completo do projeto DICT, gerando todos os artefatos necessários para implementação autônoma por agentes Claude Code.

---

## Agentes da Squad

### 1. AGT-PM-001: Project Manager (Gestor de Projeto)
**Nome de Código**: PHOENIX

**Responsabilidades**:
- Gestão geral do projeto e coordenação entre agentes
- Manutenção do backlog master e documentação do projeto
- Controle de prazos, riscos e dependências
- Interface com stakeholders (CTO, Heads)
- Aprovação e validação de artefatos antes de submissão
- Geração de relatórios de progresso e status

**Competências**:
- Metodologias ágeis (Scrum/Kanban)
- Gestão de riscos e impedimentos
- Documentação de projetos complexos
- Comunicação com stakeholders técnicos

**Artefatos Produzidos**:
- Plano Master do Projeto (PMP-001)
- Relatórios de Status (RST-XXX)
- Matriz de Riscos (MRK-001)
- Cronograma Detalhado (CRN-001)

---

### 2. AGT-SM-001: Scrum Master
**Nome de Código**: CATALYST

**Responsabilidades**:
- Facilitação de processos ágeis
- Gestão de sprints e cerimônias
- Remoção de impedimentos
- Métricas de produtividade e velocity
- Gestão do backlog de desenvolvimento
- Coordenação de dailies e retrospectivas

**Competências**:
- Scrum avançado
- Facilitação de equipes distribuídas
- Métricas ágeis
- Gestão de conflitos e impedimentos

**Artefatos Produzidos**:
- Backlog Priorizado (BKL-XXX)
- Sprint Planning (SPL-XXX)
- Velocity Reports (VEL-XXX)
- Retrospectivas (RET-XXX)

---

### 3. AGT-BA-001: Business Analyst (Analista de Negócios)
**Nome de Código**: ORACLE

**Responsabilidades**:
- Análise detalhada do Manual Operacional DICT Bacen
- Extração e catalogação de requisitos funcionais
- Criação de user stories e casos de uso
- Mapeamento de processos de negócio
- Validação de requisitos com documentação Bacen
- Criação de critérios de aceitação

**Competências**:
- Análise de requisitos regulatórios
- Modelagem de processos (BPMN)
- Escrita de user stories e casos de uso
- Domain-Driven Design (DDD)

**Artefatos Produzidos**:
- Checklist de Requisitos Funcionais (CRF-001)
- Catálogo de User Stories (UST-XXX)
- Mapeamento de Processos de Negócio (MPN-XXX)
- Matriz de Rastreabilidade (MTR-001)
- Critérios de Aceitação (CAC-XXX)

---

### 4. AGT-SA-001: Solution Architect (Arquiteto de Solução)
**Nome de Código**: NEXUS

**Responsabilidades**:
- Análise da arquitetura C4 existente (IcePanel)
- Design de arquitetura end-to-end
- Definição de integrações entre módulos
- Evolução da arquitetura Bridge e Connect (abstração)
- Especificação de interfaces (gRPC, REST)
- Decisões tecnológicas e padrões arquiteturais

**Competências**:
- Arquitetura C4 Model
- Microserviços e APIs
- Padrões de integração
- gRPC e Protocol Buffers
- Event-driven architecture

**Artefatos Produzidos**:
- Documento de Arquitetura de Solução (DAS-001)
- Diagramas C4 (Level 1-4) (C4-XXX)
- Especificação de Interfaces (EIF-XXX)
- Mapa de Integrações (MIG-001)
- ADRs - Architecture Decision Records (ADR-XXX)

---

### 5. AGT-DA-001: Data Architect (Arquiteto de Dados)
**Nome de Código**: ATLAS

**Responsabilidades**:
- Modelagem de dados do domínio DICT
- Design de schemas de banco de dados
- Estratégias de cache e persistência
- Definição de modelos de dados (entities, aggregates)
- Mapeamento de eventos de domínio
- Estratégias de migração de dados

**Competências**:
- Modelagem de dados (ER, DDD)
- PostgreSQL, MongoDB
- Event Sourcing e CQRS
- Cache strategies (Redis)
- Data migration

**Artefatos Produzidos**:
- Modelo de Dados Conceitual (MDC-001)
- Modelo de Dados Lógico (MDL-001)
- Modelo de Dados Físico (MDF-001)
- Especificação de Eventos (SEV-XXX)
- Estratégia de Cache (ECA-001)

---

### 6. AGT-API-001: API Specialist (Especialista em APIs)
**Nome de Código**: MERCURY

**Responsabilidades**:
- Análise da OpenAPI DICT Bacen
- Mapeamento de endpoints necessários
- Design de APIs internas (Core DICT)
- Especificação de contratos gRPC
- Definição de estratégias sync/async
- Documentação de APIs (OpenAPI/Swagger)

**Competências**:
- OpenAPI/Swagger
- gRPC e Protocol Buffers
- REST API design
- API versioning
- Async messaging patterns

**Artefatos Produzidos**:
- Catálogo de APIs Bacen (CAB-001)
- Especificação de APIs Internas (EAI-XXX)
- Contratos gRPC (CGR-XXX)
- Matriz Sync/Async (MSA-001)
- Documentação OpenAPI (OAI-XXX)

---

### 7. AGT-FE-001: Frontend Architect (Arquiteto Frontend)
**Nome de Código**: PRISM

**Responsabilidades**:
- Definição de funcionalidades de frontend
- Mapeamento de jornadas de usuário
- Especificação de componentes e telas
- Definição de integrações frontend-backend
- Wireframes e especificações de UI
- Requisitos de UX para chaves PIX

**Competências**:
- Frontend architecture (React, Next.js)
- UX/UI design principles
- Component-driven development
- State management
- API integration patterns

**Artefatos Produzidos**:
- Lista de Funcionalidades Frontend (LFF-001)
- Mapeamento de Jornadas (MJU-XXX)
- Especificação de Componentes (ECO-XXX)
- Wireframes e Mockups (WFR-XXX)
- Matriz Frontend-Backend (MFB-001)

---

### 8. AGT-INT-001: Integration Architect (Arquiteto de Integração)
**Nome de Código**: CONDUIT

**Responsabilidades**:
- Design do módulo Connect DICT (abstrato)
- Design do módulo Bridge DICT (abstrato)
- Definição de padrões de integração com Bacen
- Estratégias de retry, circuit breaker, timeout
- Mapeamento de fluxos de dados end-to-end
- Especificação de middleware e adapters

**Competências**:
- Integration patterns (EIP)
- Message brokers (RabbitMQ, Kafka)
- Resiliency patterns
- API Gateway patterns
- Protocol adapters

**Artefatos Produzidos**:
- Especificação Connect DICT (ECD-001)
- Especificação Bridge DICT (EBD-001)
- Mapeamento de Fluxos E2E (MFE-XXX)
- Padrões de Resiliência (PDR-001)
- Diagrama de Sequência (DSQ-XXX)

---

### 9. AGT-SEC-001: Security Architect (Arquiteto de Segurança)
**Nome de Código**: SENTINEL

**Responsabilidades**:
- Análise de requisitos de segurança Bacen
- Definição de mecanismos de autenticação/autorização
- Estratégias de prevenção a fraudes
- Proteção contra ataques de leitura
- Criptografia e gestão de certificados
- Compliance e auditoria

**Competências**:
- Security by design
- OAuth2, JWT, mTLS
- Fraud detection
- Rate limiting e throttling
- Security compliance (LGPD, PCI)

**Artefatos Produzidos**:
- Análise de Segurança (ASG-001)
- Requisitos de Segurança (RSG-XXX)
- Matriz de Controles (MCS-001)
- Políticas de Rate Limiting (PRL-001)
- Plano de Prevenção a Fraudes (PPF-001)

---

### 10. AGT-QA-001: Quality Assurance Architect (Arquiteto de Qualidade)
**Nome de Código**: VALIDATOR

**Responsabilidades**:
- Definição de estratégia de testes
- Criação de plano de testes para homologação Bacen
- Especificação de testes automatizados
- Critérios de aceitação técnicos
- Definição de ambientes de teste
- Casos de teste baseados em requisitos Bacen

**Competências**:
- Test strategy design
- Test automation (unit, integration, e2e)
- BDD/TDD
- Performance testing
- Compliance testing

**Artefatos Produzidos**:
- Estratégia de Testes (EST-001)
- Plano de Testes Homologação (PTH-001)
- Casos de Teste (CTS-XXX)
- Especificação de Testes Automatizados (ETA-XXX)
- Matriz de Cobertura (MCO-001)

---

### 11. AGT-DV-001: DevOps Architect (Arquiteto DevOps)
**Nome de Código**: FORGE

**Responsabilidades**:
- Estratégia de CI/CD
- Definição de ambientes (dev, staging, prod)
- Infraestrutura como código
- Estratégia de deploy e rollback
- Monitoramento e observabilidade
- Gestão de repositórios e branches

**Competências**:
- CI/CD (GitHub Actions, Jenkins)
- Containerization (Docker, Kubernetes)
- Infrastructure as Code (Terraform)
- Monitoring (Prometheus, Grafana)
- Git workflow strategies

**Artefatos Produzidos**:
- Estratégia de CI/CD (ECD-001)
- Especificação de Ambientes (EAM-XXX)
- Pipeline Definitions (PPL-XXX)
- Estratégia de Monitoramento (EMO-001)
- Git Workflow (GWF-001)

---

### 12. AGT-TS-001: Technical Specialist Golang/gRPC
**Nome de Código**: GOPHER

**Responsabilidades**:
- Análise de código existente (repos LBPay)
- Identificação de padrões de desenvolvimento
- Definição de padrões de código (linting, formatação)
- Avaliação de stack tecnológica
- Recomendações de bibliotecas e frameworks
- Code standards e best practices

**Competências**:
- Golang avançado
- gRPC e Protocol Buffers
- Clean Architecture
- SOLID principles
- Go patterns e idioms

**Artefatos Produzidos**:
- Análise de Stack Tecnológica (AST-001)
- Padrões de Código (PDC-001)
- Bibliotecas Recomendadas (BRE-XXX)
- Code Guidelines (CGU-001)
- Análise de Repositórios (ARE-XXX)

---

### 13. AGT-DOC-001: Technical Writer (Documentador Técnico)
**Nome de Código**: SCRIBE

**Responsabilidades**:
- Consolidação de toda documentação produzida
- Padronização de formatos e templates
- Criação de índices e cross-references
- Gestão de versionamento de documentos
- Review de clareza e completude
- Preparação de documentos para aprovação

**Competências**:
- Technical writing
- Markdown, Mermaid, PlantUML
- Documentation as Code
- Information architecture
- Version control

**Artefatos Produzidos**:
- Índice Master de Documentação (IMD-001)
- Templates de Documentos (TPL-XXX)
- Glossário de Termos (GLO-001)
- Guias de Referência Rápida (GRR-XXX)
- Pacotes de Aprovação (PAP-XXX)

---

### 14. AGT-CM-001: Compliance Manager (Gestor de Compliance)
**Nome de Código**: GUARDIAN

**Responsabilidades**:
- Análise de requisitos de homologação Bacen
- Mapeamento de checklist de compliance
- Validação de conformidade com regulamentação
- Identificação de gaps de conformidade
- Rastreabilidade de requisitos regulatórios
- Preparação para auditoria

**Competências**:
- Regulamentação Bacen (PIX, DICT)
- Compliance e auditoria
- Risk assessment
- Regulatory requirements mapping
- Documentation compliance

**Artefatos Produzidos**:
- Checklist de Homologação (CHO-001)
- Matriz de Conformidade (MCF-001)
- Análise de Gaps (AGA-001)
- Requisitos Regulatórios (RRE-XXX)
- Plano de Auditoria (PAU-001)

---

## Workflow da Squad

### Fase 1: Análise e Descoberta (Semanas 1-2)
**Agentes Ativos**: ORACLE, MERCURY, GUARDIAN, GOPHER
- Análise de documentação Bacen
- Análise de código existente
- Identificação de requisitos e gaps

### Fase 2: Design e Arquitetura (Semanas 3-4)
**Agentes Ativos**: NEXUS, ATLAS, CONDUIT, SENTINEL
- Design de arquitetura de solução
- Modelagem de dados
- Especificação de integrações

### Fase 3: Especificação Detalhada (Semanas 5-6)
**Agentes Ativos**: PRISM, VALIDATOR, FORGE
- Especificação de frontend
- Estratégia de testes
- Pipeline CI/CD

### Fase 4: Consolidação e Planejamento (Semanas 7-8)
**Agentes Ativos**: PHOENIX, CATALYST, SCRIBE
- Consolidação de artefatos
- Criação de backlog de implementação
- Preparação para aprovação

---

## Matriz de Responsabilidade (RACI)

| Artefato | PM | SM | BA | SA | DA | API | FE | INT | SEC | QA | DVO | TS | DOC | CM |
|----------|----|----|----|----|----|----|----|----|-----|----|----|----|----|-----|
| Requisitos Funcionais | A | C | R | C | I | C | C | I | C | C | I | I | R | C |
| Arquitetura Solução | A | I | C | R | C | C | C | R | R | C | C | C | R | C |
| Modelo de Dados | A | I | C | C | R | I | I | C | C | I | I | C | R | I |
| APIs Internas | A | C | C | R | I | R | C | C | C | C | I | C | R | C |
| Frontend Specs | A | C | R | C | I | C | R | I | C | C | I | I | R | I |
| Integração E2E | A | C | C | R | I | C | C | R | C | C | C | C | R | C |
| Segurança | A | C | C | C | I | C | I | C | R | C | C | C | R | R |
| Testes | A | C | C | C | I | C | C | C | C | R | C | C | R | C |
| CI/CD | A | C | I | C | I | I | I | I | C | C | R | C | R | I |
| Backlog | A | R | C | C | C | C | C | C | C | C | C | C | C | C |

**Legenda RACI**:
- R: Responsible (Responsável pela execução)
- A: Accountable (Autoridade final / Aprovador)
- C: Consulted (Consultado)
- I: Informed (Informado)

---

## Convenções de Nomenclatura

### Códigos de Agentes
- Formato: `AGT-[ÁREA]-[NÚMERO]`
- Exemplo: `AGT-PM-001`, `AGT-BA-001`

### Códigos de Artefatos
- Formato: `[TIPO]-[NÚMERO]`
- Tipos:
  - CRF: Checklist Requisitos Funcionais
  - DAS: Documento Arquitetura Solução
  - MDC/MDL/MDF: Modelo de Dados (Conceitual/Lógico/Físico)
  - UST: User Story
  - EIF: Especificação de Interface
  - ADR: Architecture Decision Record
  - E muitos outros...

### Códigos de Tarefas
- Formato: `TSK-[SQUAD]-[AGENTE]-[NÚMERO]`
- Exemplo: `TSK-SQARCH-PM-001`

---

## Comunicação e Sincronização

### Daily Sync
- **Frequência**: Diária
- **Responsável**: CATALYST (AGT-SM-001)
- **Participantes**: Todos os agentes ativos
- **Artefato**: Daily Report (DLY-XXX)

### Sprint Planning
- **Frequência**: Início de cada sprint (semanal)
- **Responsável**: CATALYST + PHOENIX
- **Artefato**: Sprint Plan (SPL-XXX)

### Sprint Review
- **Frequência**: Fim de cada sprint
- **Responsável**: PHOENIX
- **Participantes**: Squad + Stakeholders
- **Artefato**: Sprint Review (SRV-XXX)

### Retrospectiva
- **Frequência**: Fim de cada sprint
- **Responsável**: CATALYST
- **Artefato**: Retrospective (RET-XXX)

---

## Repositório de Artefatos

Todos os artefatos produzidos pela Squad de Arquitetura serão organizados em:

```
/Artefatos/
  /00_Master/              # Documentos master e índices
  /01_Requisitos/          # Requisitos funcionais e não-funcionais
  /02_Arquitetura/         # Documentos de arquitetura
  /03_Dados/               # Modelos de dados
  /04_APIs/                # Especificações de APIs
  /05_Frontend/            # Especificações de frontend
  /12_Integracao/          # Especificações de integração
  /13_Seguranca/           # Documentos de segurança
  /08_Testes/              # Estratégia e planos de teste
  /09_DevOps/              # CI/CD e infraestrutura
  /10_Compliance/          # Compliance e homologação
  /11_Gestao/              # Gestão de projeto
  /99_Templates/           # Templates reutilizáveis
```

---

## Status Atual

- **Data de Criação**: 2025-10-24
- **Status**: SQUAD_DEFINED
- **Próximo Marco**: Início da Fase 1 - Análise e Descoberta
- **Aprovação Pendente**: Sim (CTO, Head Arquitetura, Head Produto, Head Engenharia)

---

## Notas

1. Esta squad é específica para a Fase 1 (Especificação)
2. Uma nova Squad de Desenvolvimento será definida para a Fase 2 (Implementação)
3. Todos os agentes operam de forma autônoma mas coordenada
4. A indexação e rastreabilidade são fundamentais em todos os artefatos
5. Aprovações formais são obrigatórias antes de avançar para implementação
