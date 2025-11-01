# Índice de Especificações Técnicas - DICT Rate Limit Monitoring

## 📚 Estrutura de Documentação

Este projeto utiliza uma arquitetura modular de documentação para permitir máxima profundidade técnica e facilitar a orquestração da squad de desenvolvimento.

### Documentos Principais

#### 🎯 [CLAUDE.md](./CLAUDE.md)
**Documento mestre** - Visão geral do projeto, requisitos, squad e planejamento.
- Visão executiva
- **🪣 Algoritmo Token Bucket** (seção detalhada)
- Escopo (In/Out)
- Arquitetura de integração
- Squad especializada (10 agentes)
- Metodologia de trabalho
- Fases de execução

---

#### 🪣 [TOKEN_BUCKET_EXPLAINED.md](./TOKEN_BUCKET_EXPLAINED.md)
**Explicação Didática do Token Bucket** - Fundamentos do Rate Limiting do DICT BACEN.
- Conceito base (analogia do balde de fichas)
- Parâmetros fundamentais (AvailableTokens, Capacity, RefillTokens, RefillPeriodSec)
- Evolução das fichas ao longo do tempo
  - Reposição automática (refill)
  - Consumo por requisição
  - Cenários de esgotamento
- Cálculo de utilização e thresholds
- 24 políticas do DICT BACEN (tabela completa)
- Estratégia de monitoramento (por que 5 minutos?)
- Dashboards e métricas
- Exemplos práticos com timelines

**Responsável**: Tech Lead
**Dependências**: Nenhuma
**Status**: ✅ **COMPLETO**

---

### Especificações Técnicas Detalhadas

#### 🗄️ [SPECS-DATABASE.md](./SPECS-DATABASE.md)
**Database Schema & Repositories**
- Schema PostgreSQL completo (3 tabelas + views)
- Migrations SQL (up/down)
- Indexes e partitioning strategy
- Repository pattern (interfaces + implementations)
- Queries otimizados
- Data retention policies

**Responsável**: DB & Domain Engineer
**Dependências**: Nenhuma
**Status**: 📝 Pendente

---

#### 🔌 [SPECS-API.md](./SPECS-API.md)
**Dict API - REST Endpoints**
- OpenAPI 3.1 specification
- Schemas Huma (request/response)
- Controllers e handlers HTTP
- Application layer (use cases)
- Error handling (RFC 9457)
- Cache strategy (Redis)
- Rate limiting interno

**Responsável**: Dict API Engineer
**Dependências**: SPECS-DATABASE, SPECS-INTEGRATION
**Status**: 📝 Pendente

---

#### ⚡ [SPECS-WORKFLOWS.md](./SPECS-WORKFLOWS.md)
**Temporal Workflows & Activities**
- MonitorPoliciesWorkflow (cron)
- AlertLowBalanceWorkflow (child)
- Todas as activities (6+)
- Retry policies detalhadas
- Continue-As-New strategy
- Workflow testing (replay)
- Error handling avançado

**Responsável**: Temporal Workflow Engineer
**Dependências**: SPECS-DATABASE, SPECS-INTEGRATION
**Status**: 📝 Pendente

---

#### 🔗 [SPECS-INTEGRATION.md](./SPECS-INTEGRATION.md)
**Bridge gRPC & Pulsar Events**
- Proto definitions (se necessário)
- gRPC client implementation
- Mappers (Bacen ↔ gRPC)
- mTLS configuration
- Pulsar topics (schemas Avro/JSON)
- Event publishers
- Schema evolution

**Responsável**: gRPC Engineer + Pulsar Specialist
**Dependências**: Coordenação com time Bridge
**Status**: 📝 Pendente (CRÍTICO - verificar Bridge)

---

#### 📊 [SPECS-OBSERVABILITY.md](./SPECS-OBSERVABILITY.md)
**Metrics, Dashboards & Alerts**
- Prometheus metrics (gauges, counters, histograms)
- Grafana dashboards (JSON templates)
- Alert rules (Prometheus AlertManager)
- OpenTelemetry traces
- Logging strategy
- SLIs/SLOs

**Responsável**: DevOps & SRE Engineer
**Dependências**: SPECS-API, SPECS-WORKFLOWS
**Status**: 📝 Pendente

---

#### 🧪 [SPECS-TESTING.md](./SPECS-TESTING.md)
**Testing Strategy & Implementation**
- Unit tests (>85% coverage)
- Integration tests (Testcontainers)
- Temporal workflow replay tests
- Load tests (simular latência DICT)
- Mock strategies
- Test data generation
- CI/CD integration

**Responsável**: QA Lead & Test Architect
**Dependências**: Todos os SPECS acima
**Status**: 📝 Pendente

---

#### 🚀 [SPECS-DEPLOYMENT.md](./SPECS-DEPLOYMENT.md)
**DevOps, Operations & Infrastructure**
- Kubernetes manifests
- Helm charts
- Database migrations (Goose/Flyway)
- Temporal cron configuration
- Pulsar topic provisioning
- Environment variables
- Secrets management
- Disaster recovery
- Runbooks operacionais

**Responsável**: DevOps & SRE Engineer
**Dependências**: Todos os SPECS acima
**Status**: 📝 Pendente

---

#### 🔐 [SPECS-SECURITY.md](./SPECS-SECURITY.md)
**Security & Compliance**
- BACEN Manual Cap. 19 compliance checklist
- Input validation strategies
- SQL injection prevention
- Secrets management (Vault/AWS)
- mTLS configuration
- LGPD compliance
- Audit trail
- Penetration testing

**Responsável**: Security & BACEN Compliance Auditor
**Dependências**: Todos os SPECS acima
**Status**: 📝 Pendente

---

### Documentos de Apoio

#### ❓ [DUVIDAS.md](./DUVIDAS.md)
**Questões para Stakeholder**
- Dúvidas técnicas
- Decisões arquiteturais
- Coordenação com times externos (Bridge, Core-Dict)
- Validações de requisitos

**Responsável**: Tech Lead
**Status**: 📝 Pendente

---

#### 📋 [SQUAD-AGENTS.md](./SQUAD-AGENTS.md)
**Squad Detalhada & Orquestração**
- Definição completa de cada agente
- Responsabilidades específicas
- Dependências entre agentes
- Workflow de colaboração
- Templates de prompts
- Critérios de qualidade

**Responsável**: Tech Lead
**Status**: 📝 Pendente

---

## 🔄 Fluxo de Desenvolvimento

### Fase 0: Especificação (Atual)
```
CLAUDE.md (Visão Geral) ─┐
                          ├─> SPECS-INDEX.md ─┐
                          │                    ├─> SPECS-DATABASE.md
                          │                    ├─> SPECS-API.md
                          │                    ├─> SPECS-WORKFLOWS.md
                          │                    ├─> SPECS-INTEGRATION.md
                          │                    ├─> SPECS-OBSERVABILITY.md
                          │                    ├─> SPECS-TESTING.md
                          │                    ├─> SPECS-DEPLOYMENT.md
                          │                    └─> SPECS-SECURITY.md
                          │
                          ├─> DUVIDAS.md
                          └─> SQUAD-AGENTS.md
```

### Fase 1-8: Implementação
Cada agente da squad consultará os SPECS relevantes para sua área de responsabilidade.

---

## 📊 Matriz de Dependências

| Documento | Depende de | Consumido por |
|-----------|-----------|---------------|
| SPECS-DATABASE.md | - | API, Workflows, Testing |
| SPECS-API.md | DATABASE, INTEGRATION | Workflows, Testing, Deployment |
| SPECS-WORKFLOWS.md | DATABASE, INTEGRATION | Testing, Deployment |
| SPECS-INTEGRATION.md | - | API, Workflows, Testing |
| SPECS-OBSERVABILITY.md | API, Workflows | Deployment |
| SPECS-TESTING.md | Todos acima | Deployment |
| SPECS-DEPLOYMENT.md | Todos acima | - |
| SPECS-SECURITY.md | Todos acima | Deployment |

---

## 🎯 Convenções de Nomenclatura

### Arquivos
- `SPECS-*.md` - Especificações técnicas detalhadas
- `CLAUDE.md` - Documento mestre de requisitos e planejamento
- `DUVIDAS.md` - Questões pendentes
- `SQUAD-AGENTS.md` - Definição da squad

### Seções nos SPECS
- **📋 Overview** - Resumo executivo
- **🎯 Objetivos** - O que este componente resolve
- **🏗️ Arquitetura** - Diagramas e estrutura
- **📝 Especificação Detalhada** - Código, schemas, configs
- **🧪 Testing** - Estratégia de testes específica
- **📊 Métricas** - KPIs e observabilidade
- **🚀 Deployment** - Como deployar este componente
- **🔗 Referências** - Links para outros SPECS

---

## 📖 Como Usar Esta Documentação

### Para Tech Lead
1. Ler [CLAUDE.md](./CLAUDE.md) para visão geral
2. Revisar todos SPECS-* para validação técnica
3. Coordenar com times externos usando [DUVIDAS.md](./DUVIDAS.md)
4. Definir squad em [SQUAD-AGENTS.md](./SQUAD-AGENTS.md)

### Para Desenvolvedores
1. Consultar [CLAUDE.md](./CLAUDE.md) para contexto
2. Ler SPECS específicos da sua área:
   - DB Engineer → SPECS-DATABASE.md
   - API Engineer → SPECS-API.md
   - Temporal Engineer → SPECS-WORKFLOWS.md
   - etc.
3. Seguir padrões definidos nos SPECS
4. Validar com testes conforme SPECS-TESTING.md

### Para QA
1. Revisar [SPECS-TESTING.md](./SPECS-TESTING.md)
2. Validar cobertura contra todos SPECS-*
3. Criar test plans baseados em cada SPEC

### Para DevOps
1. Revisar [SPECS-DEPLOYMENT.md](./SPECS-DEPLOYMENT.md)
2. Consultar SPECS-OBSERVABILITY.md para dashboards
3. Configurar infraestrutura conforme specs

---

## ✅ Status Geral

| Documento | Status | Progresso | Responsável |
|-----------|--------|-----------|-------------|
| CLAUDE.md | ✅ Completo | 100% | Tech Lead |
| SPECS-INDEX.md | ✅ Completo | 100% | Tech Lead |
| SPECS-DATABASE.md | 📝 Pendente | 0% | DB Engineer |
| SPECS-API.md | 📝 Pendente | 0% | API Engineer |
| SPECS-WORKFLOWS.md | 📝 Pendente | 0% | Temporal Engineer |
| SPECS-INTEGRATION.md | 📝 Pendente | 0% | gRPC/Pulsar Engineers |
| SPECS-OBSERVABILITY.md | 📝 Pendente | 0% | DevOps Engineer |
| SPECS-TESTING.md | 📝 Pendente | 0% | QA Lead |
| SPECS-DEPLOYMENT.md | 📝 Pendente | 0% | DevOps Engineer |
| SPECS-SECURITY.md | 📝 Pendente | 0% | Security Auditor |
| DUVIDAS.md | 📝 Pendente | 0% | Tech Lead |
| SQUAD-AGENTS.md | 📝 Pendente | 0% | Tech Lead |

---

**Última Atualização**: 2025-10-31
**Versão**: 1.0.0
**Responsável**: Tech Lead
**Próximo Passo**: Criar SPECS-DATABASE.md
