# DICT LBPay - Squad de Agentes Especializados

Esta pasta contém os **prompts e contextos** de 8 agentes especializados que trabalham em paralelo para criar a documentação técnica do projeto DICT LBPay.

---

## 🤖 Agentes Disponíveis

| Agente | Especialidade | Documentos | Status |
|--------|--------------|------------|--------|
| **architect** | Diagramas C4, ADRs, TechSpecs | DIA-XXX, TSP-XXX | ✅ Ativo |
| **backend** | APIs REST/gRPC, Schemas DB | API-XXX, DAT-XXX, GRPC-XXX | ✅ Ativo |
| **security** | Segurança, mTLS, LGPD | SEC-XXX, CMP-XXX | ✅ Ativo |
| **qa** | Test cases, Test plans | TST-XXX | ✅ Ativo |
| **devops** | CI/CD, Kubernetes, Monitoring | DEV-XXX | ✅ Ativo |
| **frontend** | Wireframes, Componentes, UX | FE-XXX | ✅ Ativo |
| **product** | User stories, Backlog | US-XXX, BP-XXX, PM-001 | ✅ Ativo |
| **scrum** | Sprints, Retros, Checklists | PM-002, PM-003, PM-004 | ✅ Ativo |

---

## 📁 Estrutura de Diretórios

```
agents/
├── README.md                    # Este arquivo
├── architect/
│   ├── prompt.md                # Prompt do agente
│   ├── context.md               # Contexto específico (a criar)
│   └── templates/               # Templates de documentos (a criar)
├── backend/
│   ├── prompt.md
│   └── ...
├── security/
│   ├── prompt.md
│   └── ...
├── qa/
│   ├── prompt.md
│   └── ...
├── devops/
│   ├── prompt.md
│   └── ...
├── frontend/
│   ├── prompt.md
│   └── ...
├── product/
│   ├── prompt.md
│   └── ...
└── scrum/
    ├── prompt.md
    └── ...
```

---

## 🚀 Como Usar os Agentes

### Método 1: Claude Code (Task Tool)

Execute um único agente usando o Task tool:

```typescript
// Exemplo: Architect cria DIA-004
await claude.task({
  agent: 'architect',
  task: 'Create DIA-004: C4 Component Diagram for RSFN Connect'
});
```

### Método 2: Múltiplos Agentes em Paralelo (Recomendado!)

Execute 4-6 agentes simultaneamente para máxima eficiência:

```typescript
// Executar 4 agentes em paralelo
await Promise.all([
  claude.task({
    agent: 'architect',
    task: 'Create DIA-004 and DIA-005'
  }),
  claude.task({
    agent: 'backend',
    task: 'Create API-002 and TSP-001'
  }),
  claude.task({
    agent: 'devops',
    task: 'Create DEV-001, DEV-002, DEV-003'
  }),
  claude.task({
    agent: 'qa',
    task: 'Create TST-001, TST-002, TST-003'
  })
]);

// Resultado: 11 documentos criados em paralelo! 🚀
```

---

## 📋 Workflow Recomendado

### 1. **Planning** (PO + Scrum Master)

```bash
# Product Owner cria/atualiza backlog
Task: product agent "Update product backlog based on PROGRESSO_FASE_2.md"

# Scrum Master cria sprint plan
Task: scrum agent "Create Sprint 4 plan selecting high-priority items from backlog"
```

### 2. **Sprint Execution** (Todos os agentes em paralelo)

```bash
# DIA 1: Arquitetura + Backend + DevOps + QA (4 agentes)
Task: architect "Create DIA-004, DIA-005, DIA-007"
Task: backend "Create API-002, TSP-001, TSP-002"
Task: devops "Create DEV-001, DEV-002, DEV-003"
Task: qa "Create TST-001, TST-002"

# Resultado Dia 1: 11 documentos criados

# DIA 2: Security + Frontend + Integration (3 agentes)
Task: security "Create CMP-001, CMP-002, CMP-003"
Task: frontend "Create FE-001, FE-002, FE-003"
Task: backend "Create INT-002, INT-003, INT-004"

# Resultado Dia 2: 10 documentos criados
```

### 3. **Review** (Tech Lead)

```bash
# Architect revisa todos os documentos criados
Task: architect "Review all documents created in Sprint 4, check consistency and cross-references"
```

### 4. **Retrospective** (Scrum Master)

```bash
# Scrum Master facilita retrospectiva
Task: scrum "Create Sprint 4 retrospective, analyze velocity, identify improvements"
```

---

## 🎯 Prioridades por Agente (Sprint 3-4)

### Architect (Alta Prioridade)
- [ ] DIA-004: C4 Component Diagram Connect
- [ ] DIA-005: C4 Component Diagram Bridge
- [ ] DIA-007: Sequence CreateEntry
- [ ] DIA-008: Flow VSYNC Daily
- [ ] DIA-009: Deployment Kubernetes

### Backend (Alta Prioridade)
- [ ] API-002: Core DICT REST API
- [ ] API-003: Connect Admin API
- [ ] TSP-001: Temporal Workflow Engine
- [ ] TSP-002: Apache Pulsar Messaging
- [ ] INT-002: Flow ClaimWorkflow E2E
- [ ] INT-003: Flow VSYNC E2E

### Security (Alta Prioridade)
- [ ] CMP-001: Audit Logs Specification
- [ ] CMP-002: LGPD Compliance Checklist
- [ ] CMP-003: Bacen Regulatory Compliance

### QA (Alta Prioridade)
- [ ] TST-001: Test Cases CreateEntry
- [ ] TST-002: Test Cases ClaimWorkflow
- [ ] TST-003: Test Cases Bridge mTLS

### DevOps (Alta Prioridade)
- [ ] DEV-001: CI/CD Pipeline Core
- [ ] DEV-002: CI/CD Pipeline Connect
- [ ] DEV-003: CI/CD Pipeline Bridge
- [ ] DEV-004: Kubernetes Manifests

---

## 📊 Métricas de Performance

### Sem Paralelização (1 agente por vez)
- **Velocidade**: 6 docs/dia
- **Sprint 3 (2 semanas)**: ~60 docs

### Com Paralelização (4-6 agentes simultâneos)
- **Velocidade**: 20-30 docs/dia 🚀
- **Sprint 3 (2 semanas)**: ~200+ docs (muito além do necessário!)

**Ganho de Eficiência**: 3-5x mais rápido!

---

## ✅ Checklist de Qualidade

Antes de considerar um documento completo, verificar:

- [ ] Documento segue template do agente
- [ ] Todas as seções estão preenchidas
- [ ] Exemplos de código/diagramas incluídos
- [ ] Referências cruzadas corretas
- [ ] Peer review realizado (outro agente revisa)
- [ ] Checklist de validação preenchido
- [ ] Rastreabilidade para requisitos/ADRs

---

## 🔄 Processo de Peer Review

Cada documento deve ser revisado por **outro agente** antes de ser considerado completo:

| Autor | Revisor | Justificativa |
|-------|---------|---------------|
| Architect | Backend | Verificar viabilidade técnica das decisões |
| Backend | Security | Verificar aspectos de segurança das APIs |
| DevOps | Backend | Verificar alinhamento infra ↔ aplicação |
| QA | Backend | Verificar testabilidade das specs |
| Frontend | Product | Verificar alinhamento UX ↔ requisitos |

---

## 📖 Leitura Adicional

- [Claude.md](../Claude.md) - Visão geral do projeto e Squad
- [PROGRESSO_FASE_2.md](../Artefatos/00_Master/PROGRESSO_FASE_2.md) - Status dos documentos
- [PM-001: Product Backlog](../Artefatos/17_Gestao/Backlog/PM-001_Product_Backlog.md)
- [Sprint 3 Plan](../Artefatos/17_Gestao/Sprints/Sprint_03_Plan.md)

---

**Última Atualização**: 2025-10-25
**Mantido por**: Scrum Master Agent
