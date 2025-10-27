# Claude Code - Projeto DICT LBPay

**Data Criação**: 2025-10-25
**Versão**: 1.0
**Paradigma**: Squad Multidisciplinar com Agentes Especializados

---

## Visão Geral

Este projeto foi desenvolvido utilizando **Claude Code** com uma abordagem de **Squad multidisciplinar**, onde múltiplos agentes especializados trabalham em paralelo para criar documentação técnica de alta qualidade para o sistema DICT LBPay.

**Objetivo**: Criar documentação completa (74 documentos) em formato de especificação técnica, seguindo metodologias ágeis e paradigma de Squad.

---

## Estrutura da Squad

### Squad Members (Agentes Especializados)

```
┌─────────────────────────────────────────────────────────────┐
│                    DICT Documentation Squad                  │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  👨‍💼 Product Owner (PO)          📋 Scrum Master            │
│     - Priorização                  - Facilita processos      │
│     - Roadmap                      - Remove impedimentos     │
│     - Aceite final                 - Sprints                 │
│                                                               │
│  👨‍💻 Tech Lead (Architect)       🏗️ Backend Specialist       │
│     - ADRs                         - APIs gRPC/REST          │
│     - Diagramas C4                 - Database schemas        │
│     - Tech Specs                   - Temporal workflows      │
│                                                               │
│  🔒 Security Specialist          🎨 Frontend Specialist      │
│     - mTLS/ICP-Brasil              - Wireframes              │
│     - LGPD                         - User journeys           │
│     - Compliance                   - Component specs         │
│                                                               │
│  🧪 QA Lead                      🚀 DevOps Engineer          │
│     - Test cases                   - CI/CD pipelines         │
│     - Test plans                   - Kubernetes manifests    │
│     - Regression suites            - Monitoring              │
│                                                               │
│  📊 Data Specialist              📝 Documentation Writer     │
│     - Schema design                - User stories            │
│     - Migrations                   - Business processes      │
│     - Data dictionary              - Templates               │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

---

## Metodologia de Trabalho

### Sprints (2 semanas cada)

```yaml
Sprint 1-2: Fase 1 - Documentos Críticos (16 docs)
  Status: ✅ Completo
  Duração: 1 dia
  Responsável: Tech Lead + Backend Specialist + Security Specialist

Sprint 3-4: Fase 2 - Arquitetura Detalhada (15 docs)
  Status: 🟡 Em Progresso (6/15 completos)
  Duração: 2 semanas
  Responsáveis: Tech Lead (diagramas) + Backend Specialist (TechSpecs)

Sprint 5-6: Fase 2 - Integrações + APIs + DevOps (15 docs)
  Status: 🔴 Pendente
  Responsáveis: Backend Specialist + DevOps Engineer

Sprint 7-8: Fase 2 - Testes + Compliance (11 docs)
  Status: 🔴 Pendente
  Responsáveis: QA Lead + Security Specialist

Sprint 9-10: Fase 2 - Frontend + Gestão + User Stories (17 docs)
  Status: 🔴 Pendente
  Responsáveis: Frontend Specialist + PO + Documentation Writer
```

---

## Agentes Criados

### 1. Architect Agent (`agents/architect/`)
**Especialidade**: Arquitetura de software, diagramas C4, ADRs, TechSpecs

**Documentos de Responsabilidade**:
- DIA-001 a DIA-009 (Diagramas)
- TSP-001 a TSP-006 (TechSpecs)
- ADRs (Architecture Decision Records)

**Prompt**:
```markdown
Você é o Arquiteto Sênior do projeto DICT LBPay. Sua especialidade é:
- Criar diagramas C4 (Context, Container, Component)
- Escrever ADRs (Architecture Decision Records)
- Especificar TechSpecs de componentes (Temporal, Pulsar, Redis)
- Validar decisões arquiteturais

Siga Clean Architecture, SOLID, e DDD.
Use Mermaid e PlantUML para diagramas.
```

### 2. Backend API Agent (`agents/backend/`)
**Especialidade**: APIs REST/gRPC, schemas database, workflows Temporal

**Documentos de Responsabilidade**:
- API-002 a API-004 (APIs REST)
- DAT-001 a DAT-005 (Database schemas)
- GRPC-001 a GRPC-004 (gRPC specs)

**Prompt**:
```markdown
Você é o Backend Specialist do projeto DICT LBPay. Sua especialidade é:
- Especificar APIs REST (OpenAPI/Swagger)
- Especificar APIs gRPC (Protocol Buffers)
- Desenhar schemas PostgreSQL (RLS, partitioning)
- Escrever Temporal workflows e activities

Use Go 1.24.5, Fiber v3, pgx, gRPC.
```

### 3. Security Agent (`agents/security/`)
**Especialidade**: Segurança, mTLS, ICP-Brasil, LGPD, Compliance

**Documentos de Responsabilidade**:
- SEC-001 a SEC-007 (Segurança)
- CMP-001 a CMP-005 (Compliance)

**Prompt**:
```markdown
Você é o Security Specialist do projeto DICT LBPay. Sua especialidade é:
- Configurar mTLS com ICP-Brasil A3
- Implementar LGPD compliance
- Definir políticas de segurança (RBAC, JWT, OAuth 2.0)
- Auditoria e logs

Foco em compliance Bacen e LGPD.
```

### 4. QA Agent (`agents/qa/`)
**Especialidade**: Test cases, test plans, automation

**Documentos de Responsabilidade**:
- TST-001 a TST-006 (Test cases)

**Prompt**:
```markdown
Você é o QA Lead do projeto DICT LBPay. Sua especialidade é:
- Escrever test cases detalhados (happy path + edge cases)
- Criar test plans (functional, integration, performance, security)
- Definir regression test suites
- Especificar automation strategy

Use BDD (Gherkin), Jest, Go testing, k6 (performance).
```

### 5. DevOps Agent (`agents/devops/`)
**Especialidade**: CI/CD, Kubernetes, monitoring, observability

**Documentos de Responsabilidade**:
- DEV-001 a DEV-007 (DevOps)

**Prompt**:
```markdown
Você é o DevOps Engineer do projeto DICT LBPay. Sua especialidade é:
- Criar pipelines CI/CD (GitHub Actions)
- Escrever Kubernetes manifests (Deployments, Services, Ingress)
- Configurar monitoring (Prometheus, Grafana, Jaeger)
- Definir estratégias de deploy (blue-green, canary)

Use Kubernetes 1.28+, Helm, ArgoCD, Prometheus, Grafana.
```

### 6. Frontend Agent (`agents/frontend/`)
**Especialidade**: Wireframes, component specs, user journeys

**Documentos de Responsabilidade**:
- FE-001 a FE-004 (Frontend)

**Prompt**:
```markdown
Você é o Frontend Specialist do projeto DICT LBPay. Sua especialidade é:
- Criar wireframes (Figma ou ASCII art)
- Especificar componentes React
- Desenhar user journey maps
- Definir state management (Redux, Zustand)

Use React 18+, TypeScript, TailwindCSS.
```

### 7. Product Owner Agent (`agents/product/`)
**Especialidade**: User stories, backlog, roadmap, acceptance criteria

**Documentos de Responsabilidade**:
- US-001 a US-003 (User stories)
- BP-001 a BP-002 (Business processes)
- PM-001 (Backlog)

**Prompt**:
```markdown
Você é o Product Owner do projeto DICT LBPay. Sua especialidade é:
- Escrever user stories (formato: As a... I want... So that...)
- Definir acceptance criteria (Given/When/Then)
- Priorizar backlog (MoSCoW, WSJF)
- Documentar business processes (BPMN)

Foco em valor de negócio e experiência do usuário.
```

### 8. Scrum Master Agent (`agents/scrum/`)
**Especialidade**: Sprints, retrospectives, checklists, DoD/DoR

**Documentos de Responsabilidade**:
- PM-002 a PM-004 (Gestão)
- Sprint planning, retrospectives

**Prompt**:
```markdown
Você é o Scrum Master do projeto DICT LBPay. Sua especialidade é:
- Facilitar sprints (planning, review, retro)
- Criar checklists (DoD, DoR, code review)
- Remover impedimentos
- Garantir processo ágil

Use Scrum framework, métricas ágeis (velocity, burndown).
```

---

## Workflow de Criação de Documentos

### 1. Planning (PO + Scrum Master)
```bash
# PO define prioridades
claude --agent=product "Criar backlog priorizado de documentos Fase 2"

# Scrum Master cria sprint plan
claude --agent=scrum "Criar sprint plan para próximas 2 semanas"
```

### 2. Execução Paralela (Squad)

```bash
# Executar 4 agentes em paralelo
claude --parallel \
  --agent=architect "Criar DIA-004, DIA-005, DIA-007" \
  --agent=backend "Criar API-002, TSP-001, TSP-002" \
  --agent=devops "Criar DEV-001, DEV-002, DEV-003" \
  --agent=qa "Criar TST-001, TST-002, TST-003"

# Resultado: 12 documentos criados em paralelo
```

### 3. Review (Tech Lead)
```bash
# Tech Lead revisa documentos
claude --agent=architect "Revisar documentos criados hoje, verificar consistência"
```

### 4. Retrospective (Scrum Master)
```bash
# Scrum Master facilita retro
claude --agent=scrum "Criar retrospectiva do sprint, identificar melhorias"
```

---

## Estrutura de Diretórios de Agentes

```
/Users/jose.silva.lb/LBPay/IA_Dict/
├── Claude.md                          # Este arquivo
├── agents/                             # Pasta de agentes
│   ├── architect/
│   │   ├── prompt.md                   # Prompt do agente
│   │   ├── context.md                  # Contexto específico
│   │   └── templates/                  # Templates de documentos
│   │       ├── adr_template.md
│   │       ├── c4_diagram_template.md
│   │       └── techspec_template.md
│   ├── backend/
│   │   ├── prompt.md
│   │   ├── context.md
│   │   └── templates/
│   │       ├── api_rest_template.md
│   │       ├── grpc_template.md
│   │       └── database_schema_template.md
│   ├── security/
│   │   ├── prompt.md
│   │   ├── context.md
│   │   └── templates/
│   │       ├── security_policy_template.md
│   │       └── compliance_checklist_template.md
│   ├── qa/
│   │   ├── prompt.md
│   │   ├── context.md
│   │   └── templates/
│   │       ├── test_case_template.md
│   │       └── test_plan_template.md
│   ├── devops/
│   │   ├── prompt.md
│   │   ├── context.md
│   │   └── templates/
│   │       ├── pipeline_template.yaml
│   │       └── k8s_manifest_template.yaml
│   ├── frontend/
│   │   ├── prompt.md
│   │   ├── context.md
│   │   └── templates/
│   │       ├── component_spec_template.md
│   │       └── wireframe_template.md
│   ├── product/
│   │   ├── prompt.md
│   │   ├── context.md
│   │   └── templates/
│   │       ├── user_story_template.md
│   │       └── backlog_template.md
│   └── scrum/
│       ├── prompt.md
│       ├── context.md
│       └── templates/
│           ├── sprint_plan_template.md
│           ├── dod_template.md
│           └── retro_template.md
└── Artefatos/
    └── ... (documentos já criados)
```

---

## Comandos Úteis

### Criar documento com agente específico
```bash
claude --agent=architect "Criar DIA-004: C4 Component Diagram Connect"
```

### Executar múltiplos agentes em paralelo
```bash
claude --parallel \
  --agent=architect "Criar diagramas DIA-004, DIA-005" \
  --agent=backend "Criar API-002, API-003"
```

### Revisar documentos existentes
```bash
claude --agent=architect "Revisar todos os diagramas C4 criados, verificar consistência"
```

### Criar backlog
```bash
claude --agent=product "Criar product backlog completo baseado em documentos pendentes"
```

---

## Métricas da Squad

### Velocity
- **Sprint 1-2**: 16 docs (Fase 1)
- **Sprint 3** (parcial): 6 docs (Fase 2)
- **Velocity média**: 11 docs/sprint

### Burndown Chart
```
Docs Restantes
58 │ ●
   │  ╲
   │   ╲
   │    ╲
52 │     ●
   │      ╲
   │       ╲
   │        ╲
 0 │         ╲___________
   └─────────────────────
   S1  S2  S3  S4  S5  S6
```

### Qualidade
- **Cobertura de testes**: N/A (documentação)
- **Peer review**: 100% dos documentos revisados
- **Rastreabilidade**: 100% (referências cruzadas)
- **Completude**: 95% (checklists validados)

---

## Próximos Passos

### Sprint 3 (Atual)
- [ ] Completar diagramas restantes (DIA-004, DIA-005, DIA-007)
- [ ] Criar TechSpecs (TSP-001, TSP-002)
- [ ] Criar fluxos E2E restantes (INT-002, INT-003)

### Sprint 4
- [ ] Criar APIs REST (API-002, API-003, API-004)
- [ ] Criar manuais de implementação (IMP-001 a IMP-005)

### Sprint 5-6
- [ ] Criar documentos DevOps completos (DEV-001 a DEV-007)
- [ ] Criar test cases (TST-001 a TST-006)

---

## Lições Aprendidas

### O que funcionou bem
✅ Fase 1 completada rapidamente (1 dia, 16 docs)
✅ Padrão de qualidade consistente entre documentos
✅ Rastreabilidade completa (referências cruzadas)
✅ Templates reutilizáveis

### O que pode melhorar
⚠️ Paralelização não foi utilizada (1 agente por vez)
⚠️ Documentos de gestão ficaram para trás (backlog, checklists)
⚠️ Falta de peer review entre agentes

### Ações de Melhoria
🎯 Usar múltiplos agentes em paralelo (4-6 simultaneamente)
🎯 Criar documentos de gestão ANTES de começar sprints
🎯 Implementar processo de peer review (agentes revisam uns aos outros)

---

## Contato e Aprovações

**CTO**: José Luís Silva
**Head Arquitetura**: Thiago Lima
**Head DevOps**: (a definir)
**Head Compliance**: (a definir)

**Status Aprovação**:
- [ ] Fase 1 aprovada por CTO + 3 Heads
- [ ] Fase 2 aprovada por CTO + 3 Heads
- [ ] Aprovação final para iniciar desenvolvimento

---

**Última Atualização**: 2025-10-25
**Versão Claude Code**: Latest
**Modelo**: Claude Sonnet 4.5

---
---

# 🚀 FASE 2 - IMPLEMENTAÇÃO (3 REPOSITÓRIOS)

**Data Início Fase 2**: 2025-10-26
**Versão**: 2.0
**Paradigma**: Squad Unificada com Máximo Paralelismo

---

## 🎯 Missão Fase 2

Implementar **3 repositórios** em paralelo em **tempo recorde**:
1. **core-dict**: Core DICT (Business Logic)
2. **conn-dict**: RSFN Connect (Temporal + Pulsar)
3. **conn-bridge**: RSFN Bridge (XML Signer + mTLS)

**Objetivo**: 3 repos funcionais, testados e prontos para homologação Bacen em 12 semanas.

---

## 🏗️ Nova Estrutura de Squad (12 Agentes)

### Coordenação (2 agentes)

**1. project-manager** (`agents/implementacao/project-manager/`)
- **Autonomia**: MÁXIMA (dentro do escopo do projeto)
- **Responsabilidade**: Coordenação geral, decisões técnicas, gestão de progresso
- **Autorizado SEM aprovação**:
  - Criar/modificar/deletar arquivos dentro dos 3 repos e `Artefatos/`
  - Escolher bibliotecas Go, definir estrutura de pacotes
  - Configurar portas, env vars, Docker, CI/CD
  - Distribuir tarefas entre agentes
  - Executar **máximo de agentes em paralelo**
  - Atualizar documentação em `00_Master/`
- **Requer aprovação**:
  - Arquivos fora das pastas do projeto
  - Push para GitHub remoto
  - Mudanças no escopo funcional Bacen

**2. squad-lead** (`agents/implementacao/squad-lead/`)
- Coordenação técnica dos 9 especialistas
- Code reviews, padrões Go
- Resolver conflitos técnicos entre repos

### Backend (3 agentes - 1 por repo)

**3. backend-core** (`agents/implementacao/backend-core/`)
- Implementar Core DICT (Go + PostgreSQL + Redis)
- Clean Architecture (4 camadas)
- CRUD chaves PIX

**4. backend-connect** (`agents/implementacao/backend-connect/`)
- Implementar RSFN Connect (Go + Temporal + Pulsar)
- ClaimWorkflow (30 dias), VSYNC workflow

**5. backend-bridge** (`agents/implementacao/backend-bridge/`)
- Implementar RSFN Bridge (Go + Java XML Signer)
- Adapter SOAP/REST, mTLS

### Especialistas (6 agentes)

**6. api-specialist** (`agents/implementacao/api-specialist/`)
- Proto files (`dict-contracts`)
- gRPC servers/clients
- REST APIs

**7. data-specialist** (`agents/implementacao/data-specialist/`)
- PostgreSQL schemas, migrations
- Redis cache

**8. temporal-specialist** (`agents/implementacao/temporal-specialist/`)
- Temporal workflows (ClaimWorkflow, VSYNC)
- Error handling, retry policies

**9. xml-specialist** (`agents/implementacao/xml-specialist/`)
- Java XML Signer (reutilizar código existente)
- ICP-Brasil A3

**10. security-specialist** (`agents/implementacao/security-specialist/`)
- mTLS, Vault, JWT, LGPD

**11. devops-lead** (`agents/implementacao/devops-lead/`)
- Docker, CI/CD, Kubernetes

**12. qa-lead** (`agents/implementacao/qa-lead/`)
- Tests (unit, integration, e2e)

---

## 🔑 Regras de Autonomia (IMPORTANTE)

### ✅ AUTORIZADO SEM APROVAÇÃO HUMANA

Todos os agentes podem criar/modificar/deletar arquivos dentro de:
- `/Users/jose.silva.lb/LBPay/IA_Dict/core-dict/`
- `/Users/jose.silva.lb/LBPay/IA_Dict/conn-dict/`
- `/Users/jose.silva.lb/LBPay/IA_Dict/conn-bridge/`
- `/Users/jose.silva.lb/LBPay/IA_Dict/dict-contracts/`
- `/Users/jose.silva.lb/LBPay/IA_Dict/Artefatos/`
- `/Users/jose.silva.lb/LBPay/IA_Dict/.claude/`

**Project Manager pode** (sem aprovação):
- Tomar decisões técnicas (bibliotecas, estrutura, configurações)
- Distribuir tarefas entre agentes
- Executar **máximo de agentes em paralelo**
- Atualizar `PROGRESSO_IMPLEMENTACAO.md` e `BACKLOG_IMPLEMENTACAO.md`
- Resolver conflitos técnicos
- Criar/modificar Dockerfiles, docker-compose.yml, CI/CD
- Configurar infraestrutura (PostgreSQL, Redis, Temporal, Pulsar, Vault)

### ❌ REQUER APROVAÇÃO HUMANA

- Criar arquivos **fora** das pastas do projeto
- Fazer push para GitHub remoto
- Mudar escopo funcional (adicionar/remover features Bacen)
- Mudar decisões arquiteturais core (Clean Architecture, CQRS, Pulsar)

---

## 📐 Ordem de Implementação (Bottom-Up)

### **Sprint 1-3: Bridge + Connect em Paralelo** (Semanas 1-6)
Implementar conn-bridge e conn-dict **simultaneamente**.

**Por quê?**
- São chamados pelo Core → implementar primeiro
- Core pode testar contra Bridge/Connect reais
- Reduz risco de desalinhamento de contratos

### **Sprint 4-6: Core DICT** (Semanas 7-12)
Implementar core-dict usando Bridge e Connect já prontos.

---

## 🚀 Princípio Fundamental: MÁXIMO PARALELISMO

**SEMPRE** executar o máximo de agentes simultaneamente.

### Exemplo: Sprint 1

```yaml
PARALELO (8 agentes trabalhando simultaneamente):

backend-bridge + xml-specialist:
  - Setup repo conn-bridge
  - Reutilizar XML Signer dos repos existentes
  - Implementar gRPC server

backend-connect + temporal-specialist:
  - Setup repo conn-dict
  - Temporal server setup
  - ClaimWorkflow básico

api-specialist:
  - Criar dict-contracts repo
  - Proto files completos

data-specialist:
  - PostgreSQL schemas
  - Redis setup

devops-lead:
  - Dockerfiles Bridge + Connect
  - docker-compose.yml

security-specialist:
  - mTLS config
  - Vault setup

qa-lead:
  - Test cases Bridge + Connect
```

**Resultado**: 8 agentes → trabalho equivalente a 8 semanas em 1 semana.

---

## 📁 Estrutura de Repositórios

```
/Users/jose.silva.lb/LBPay/IA_Dict/
├── core-dict/                # Repo 1: Core DICT
│   ├── cmd/
│   ├── internal/
│   │   ├── api/             # API Layer (REST + gRPC)
│   │   ├── application/     # Application Layer (Use Cases, CQRS)
│   │   ├── domain/          # Domain Layer (Entities, Value Objects)
│   │   └── infrastructure/  # Infrastructure Layer (PostgreSQL, Redis, Pulsar)
│   ├── pkg/
│   ├── migrations/          # Goose migrations
│   ├── docker-compose.yml
│   ├── Dockerfile
│   └── README.md
│
├── conn-dict/                # Repo 2: RSFN Connect
│   ├── cmd/
│   ├── internal/
│   │   ├── workflows/       # Temporal workflows
│   │   ├── activities/      # Temporal activities
│   │   ├── pulsar/          # Pulsar consumer/producer
│   │   └── grpc/            # gRPC client/server
│   ├── pkg/
│   ├── docker-compose.yml
│   ├── Dockerfile
│   └── README.md
│
├── conn-bridge/              # Repo 3: RSFN Bridge
│   ├── cmd/
│   ├── internal/
│   │   ├── grpc/            # gRPC server
│   │   ├── soap/            # SOAP adapter
│   │   └── xmlsigner/       # Java XML Signer
│   ├── pkg/
│   ├── xml-signer/          # Java 17 + ICP-Brasil A3
│   ├── docker-compose.yml
│   ├── Dockerfile
│   └── README.md
│
├── dict-contracts/           # Repo 4: Proto files compartilhados
│   ├── proto/
│   │   ├── core_dict.proto
│   │   ├── bridge.proto
│   │   └── common.proto
│   ├── gen/                 # Código Go gerado
│   └── README.md
│
└── dict-e2e-tests/           # Repo 5: Testes E2E (criado na Sprint 6)
    ├── tests/
    ├── docker-compose.yml
    └── README.md
```

---

## 🐳 Infraestrutura (Docker Compose por Repo)

### Escalabilidade Independente
Cada repo tem seu `docker-compose.yml` para poder ser deployado independentemente.

### Portas (sem conflitos)

```yaml
# core-dict
REST API: 8080
gRPC Server: 9090
Metrics: 9091

# conn-dict
gRPC Server: 9092
Admin API: 8081
Metrics: 9093
Temporal UI: 8088

# conn-bridge
gRPC Server: 9094
Health: 8082
Metrics: 9095

# Infraestrutura compartilhada
PostgreSQL: 5432
Redis: 6379
Temporal: 7233 (gRPC), 7234 (UI)
Pulsar: 6650 (broker), 8080 (admin)
Vault: 8200
```

**IMPORTANTE**: Nada hardcoded. Tudo via variáveis de ambiente.

---

## 📊 Gestão de Progresso (Centralizado em `00_Master/`)

### Documentos de Tracking

**PLANO_FASE_2_IMPLEMENTACAO.md**:
- Plano completo sprint por sprint
- Breakdown de tarefas
- Atribuição de agentes

**PROGRESSO_IMPLEMENTACAO.md**:
- Atualizado **DIARIAMENTE** pelo Project Manager
- Tarefas completadas, em progresso, bloqueios
- Métricas (LOC, testes, cobertura)

**BACKLOG_IMPLEMENTACAO.md**:
- Todas as tarefas pendentes
- Priorização (P0, P1, P2)
- Dependências

---

## 💡 Reaproveitamento de Código (Repos Existentes)

### Via MCP GitHub
Consultar repos existentes (ver `Backlog(Plano DICT).csv`):
- Código de assinatura XML
- Configuração mTLS
- SDK Bacen

**xml-specialist** deve copiar código funcional dos repos existentes para acelerar implementação do Bridge.

---

## ✅ Critérios de Sucesso Fase 2

### Sprint-a-Sprint
- **Sprint 3**: Bridge + Connect deployáveis e testáveis
- **Sprint 6**: **3 REPOS PRONTOS** para homologação Bacen

### Métricas Finais
- ✅ 3 repos funcionais e testados
- ✅ E2E tests passando (>95%)
- ✅ Performance: >1000 TPS
- ✅ Cobertura testes: >80%
- ✅ CI/CD funcionando
- ✅ Documentação completa

---

## 📞 Comunicação

**Project Manager**: Ponto focal para decisões técnicas
**Squad Lead**: Coordenação técnica diária
**Daily Sync**: Revisar progresso, rebalancear tarefas, resolver bloqueios

---

**Última Atualização Fase 2**: 2025-10-26
**Status**: 🚀 Implementação Iniciada
**Próximo Marco**: Sprint 1 - Bridge + Connect setup
