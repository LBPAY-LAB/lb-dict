# DICT CID/VSync Synchronization System

🎯 **Sistema de Sincronização CID/VSync** conforme BACEN Capítulo 9 - Implementação com Squad Especializada de Agentes Claude

---

## ✅ Status do Projeto

**Setup**: 🟢 **COMPLETO E PRONTO PARA IMPLEMENTAÇÃO**

- ✅ 8 agentes especializados configurados
- ✅ 3 comandos de orquestração criados
- ✅ Estrutura de documentação preparada
- ✅ 7 fases de implementação planejadas
- ✅ Integração com plugins existentes

---

## 🚀 Quick Start

### Iniciar Implementação

```bash
# No Claude Code, execute:
/orchestrate-implementation
```

Este comando irá:
1. Analisar complexidade da tarefa atual
2. Determinar níveis de pensamento apropriados
3. Coordenar agentes em paralelo/sequência
4. Executar as 7 fases de implementação

### Executar Testes

```bash
/run-tests
```

### Revisar Código

```bash
/review-code
```

---

## 📋 Squad de Agentes

### Planning Squad (Opus - Pensamento Profundo)
- **ultra-architect-planner**: Arquitetura, BACEN compliance, validação de padrões

### Development Squad (Sonnet - Implementação)
- **go-backend-specialist**: Domain/Application layers
- **temporal-workflow-engineer**: Workflows & activities
- **integration-specialist**: Pulsar, gRPC, Redis

### Quality Squad (Opus - Garantia de Qualidade)
- **qa-lead-test-architect**: Testes com >80% coverage
- **security-compliance-auditor**: BACEN/LGPD compliance

### Operations Squad (Sonnet - Infraestrutura)
- **devops-engineer**: CI/CD, Docker, Kubernetes
- **technical-writer**: Documentação técnica

**Total**: 8 agentes (2 Opus, 6 Sonnet)

---

## 📁 Estrutura do Projeto

```
IA_SyncKeys/
├── .claude/
│   ├── agents/implementacao/
│   │   ├── planning/
│   │   │   └── ultra-architect-planner.md
│   │   ├── development/
│   │   │   ├── go-backend-specialist.md
│   │   │   ├── temporal-workflow-engineer.md
│   │   │   └── integration-specialist.md
│   │   ├── quality/
│   │   │   ├── qa-lead-test-architect.md
│   │   │   └── security-compliance-auditor.md
│   │   └── operations/
│   │       ├── devops-engineer.md
│   │       └── technical-writer.md
│   ├── commands/
│   │   ├── orchestrate-implementation.md
│   │   ├── run-tests.md
│   │   └── review-code.md
│   ├── Claude.md (configuração principal)
│   └── Specs.md (especificações técnicas)
├── docs/
│   ├── architecture/thinking-logs/
│   ├── requirements/
│   ├── reviews/
│   ├── security/thinking-logs/
│   └── SQUAD_SETUP_COMPLETE.md
├── connector-dict/ (repositório clonado)
└── README.md (este arquivo)
```

---

## 🎯 Fases de Implementação

| Fase | Duração | Status | Descrição |
|------|---------|--------|-----------|
| **Fase 0** | 2-3 dias | 🟡 READY | Análise técnica do connector-dict |
| **Fase 1** | 3-4 dias | ⏸️ PENDING | Domain & Application layers |
| **Fase 2** | 2-3 dias | ⏸️ PENDING | PostgreSQL repositories |
| **Fase 3** | 4-5 dias | ⏸️ PENDING | Temporal workflows |
| **Fase 4** | 3-4 dias | ⏸️ PENDING | Integração Pulsar/gRPC/Redis |
| **Fase 5** | 2-3 dias | ⏸️ PENDING | Quality assurance |
| **Fase 6** | 2-3 dias | ⏸️ PENDING | DevOps & documentação |
| **Fase 7** | 2-3 dias | ⏸️ PENDING | Production readiness |

**Duração Total**: 20-28 dias (~4-6 semanas)

---

## 🧠 Níveis de Pensamento

| Nível | Palavra-chave | Uso |
|-------|--------------|-----|
| Básico | `think` | Tarefas simples, refatorações menores |
| Profundo | `think hard` | Features complexas, decisões de design |
| Muito Profundo | `think harder` | Problemas complexos, debugging difícil |
| Ultra Profundo | `ultrathink` | Arquitetura crítica, BACEN compliance |

### Triggers Automáticos

- **BACEN compliance** → `ultrathink`
- **Security/LGPD** → `ultrathink`
- **Temporal workflows** → `think harder`
- **Database schema** → `think hard`
- **Integration patterns** → `think hard`
- **Simple implementation** → `think`

---

## 📊 Requisitos de Qualidade

### Coverage
- Domain layer: >90%
- Application layer: >85%
- Infrastructure layer: >75%
- **Overall: >80%**

### Code Quality
- golangci-lint: **Score A**
- Cyclomatic complexity: <10
- Max function length: <50 lines

### Compliance
- BACEN Chapter 9: **100%**
- LGPD: **100%**
- Security (OWASP Top 10): **0 vulnerabilities**

---

## 🔧 Comandos Úteis

### Desenvolvimento Local

```bash
# Iniciar dependências
docker-compose up -d postgres redis pulsar temporal

# Aplicar migrations
cd connector-dict/apps/orchestration-worker
go run ./cmd/migrate

# Executar worker
go run ./cmd/worker

# Testes
make test
make test-integration
make coverage-check

# Lint
golangci-lint run --timeout 5m ./...
```

### CI/CD

```bash
# Build Docker image
docker build -t orchestration-worker:dev .

# Deploy staging
kubectl apply -f k8s/staging/

# Deploy production
kubectl apply -f k8s/production/
```

---

## 📚 Documentação

### Principais Documentos

- **[SQUAD_SETUP_COMPLETE.md](docs/SQUAD_SETUP_COMPLETE.md)**: Status completo da squad
- **[Claude.md](.claude/Claude.md)**: Configuração do projeto
- **[Specs.md](.claude/Specs.md)**: Especificações técnicas
- **[Squad Guide](.claude/agents/claude-code-agent-squad-guide.md)**: Guia completo de uso

### Documentação BACEN

- **RF_Dict_Bacen.md**: Especificação BACEN Capítulo 9
- **instrucoes-orchestration-worker.md**: Padrões connector-dict
- **instrucoes-app-dict.md**: Padrões Dict API

---

## 🎓 Como Usar a Squad

### Exemplo 1: Iniciar Fase 0

```
"ultra-architect-planner, think hard about Phase 0 - Technical Analysis:

Analyze connector-dict repository and document:
1. Pulsar events for Entry operations
2. Entry fields needed for CID generation
3. PostgreSQL connection patterns
4. Bridge gRPC endpoint status
5. Core-Dict event consumer status

Output: /docs/architecture/analysis/phase0-findings.md"
```

### Exemplo 2: Implementar Domain Layer

```
"go-backend-specialist, think hard about implementing domain/cid/:

Study connector-dict/internal/domain/claim/ for patterns.

Implement:
- CID entity with GenerateCID() method
- VSync value object
- EntryData value object
- Repository interface

Follow Clean Architecture - NO infrastructure dependencies.
Write unit tests first (TDD).
Target: >90% coverage.

Output: internal/domain/cid/*.go + *_test.go"
```

### Exemplo 3: Security Audit

```
"security-compliance-auditor, ultrathink about BACEN compliance audit:

Review all implemented code for:
1. BACEN Chapter 9 compliance (CID/VSync spec)
2. LGPD compliance (no PII in logs)
3. Security vulnerabilities (OWASP Top 10)
4. Audit trail completeness

Output: /docs/security/bacen-compliance-audit.md"
```

---

## 🚦 Próximos Passos

### 1. Iniciar Fase 0 (Recomendado)

```bash
/orchestrate-implementation
```

Ou manualmente:

```
"Orchestrator, start Phase 0 - Technical Analysis with ultra-architect-planner"
```

### 2. Após Fase 0 Completa

```bash
# Review findings
cat docs/architecture/analysis/phase0-findings.md

# Proceed to Phase 1
"Orchestrator, proceed to Phase 1 - Domain & Application Layer"
```

### 3. Continuous Quality Checks

```bash
# After each phase
/run-tests
/review-code
```

---

## 🛡️ Segurança e Compliance

### BACEN Chapter 9
- ✅ CID = SHA-256(Entry fields per spec)
- ✅ VSync = XOR cumulative of all CIDs
- ✅ Daily verification (3 AM cron)
- ✅ Reconciliation on mismatch
- ✅ Audit trail (5 years retention)

### LGPD
- ✅ No PII in logs (CPF/phone/email masked)
- ✅ Encryption at rest
- ✅ Access control
- ✅ Retention policies

---

## 📞 Suporte

### Problemas Comuns

**Q: Como adicionar um novo agente?**
A: Crie arquivo `.md` em `.claude/agents/implementacao/{squad}/` seguindo o template do guia.

**Q: Como ajustar níveis de pensamento?**
A: Edite o campo `thinking_level` no frontmatter do agente ou use prompts com "think harder".

**Q: Como adicionar um novo comando?**
A: Crie arquivo `.md` em `.claude/commands/` com frontmatter `description`.

### Recursos

- Squad Guide: [claude-code-agent-squad-guide.md](.claude/agents/claude-code-agent-squad-guide.md)
- Claude Code Docs: https://docs.claude.com/claude-code
- Temporal Docs: https://docs.temporal.io
- Pulsar Docs: https://pulsar.apache.org/docs

---

## 📈 Métricas de Sucesso

| Métrica | Target | Status |
|---------|--------|--------|
| Test Coverage | >80% | 🔄 In Progress |
| BACEN Compliance | 100% | 🔄 In Progress |
| Code Quality | Score A | 🔄 In Progress |
| Documentation | 100% | 🔄 In Progress |
| Performance | <100ms p99 | 🔄 In Progress |
| Security | 0 vulns | 🔄 In Progress |

---

## 🎉 Acknowledgments

Este projeto utiliza:
- **Claude Code** by Anthropic
- **Connector-Dict** patterns by LBPay
- **BACEN Chapter 9** specification
- **Squad Guide** best practices

---

**🚀 Ready to build! Execute `/orchestrate-implementation` to start Phase 0.**

*Última atualização: 2024-10-28*
