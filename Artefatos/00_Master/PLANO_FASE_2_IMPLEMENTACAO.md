# Plano - Fase 2: Implementação (3 Repositórios)

**Data Início**: 2025-10-26
**Data Conclusão Prevista**: 2026-01-17 (12 semanas)
**Status**: 🚀 Iniciado
**Versão**: 1.0

---

## 🎯 Objetivo

Implementar **3 repositórios** em paralelo:
1. **core-dict**: Core DICT (Business Logic + PostgreSQL + Redis)
2. **conn-dict**: RSFN Connect (Temporal + Pulsar)
3. **conn-bridge**: RSFN Bridge (XML Signer + mTLS)

**Meta**: 3 repos funcionais, testados (>80% cobertura), performantes (>1000 TPS) e prontos para homologação Bacen em **12 semanas**.

---

## 📊 Resumo Executivo

| Métrica | Valor |
|---------|-------|
| **Repos a Implementar** | 3 + 1 contratos + 1 e2e tests = 5 repos |
| **Duração Total** | 12 semanas (6 sprints de 2 semanas) |
| **Squad** | 12 agentes (1 PM, 1 Squad Lead, 10 especialistas) |
| **Execução** | **Máximo paralelismo** (6-8 agentes simultâneos) |
| **Ordem** | Bottom-Up: Bridge + Connect (Sprint 1-3) → Core (Sprint 4-6) |
| **Linhas de Código Estimadas** | ~15k LOC Go + ~2k LOC Java |
| **Testes Estimados** | ~200 unit tests + ~50 integration tests + ~20 e2e tests |

---

## 🏗️ Arquitetura dos 3 Repos

### 1. core-dict (Core DICT)

**Tecnologias**: Go 1.24.5, Fiber v3, PostgreSQL 16, Redis v9.14.1, gRPC, Pulsar

**Responsabilidade**: Business logic (CRUD chaves PIX, validações Bacen)

**Clean Architecture (4 camadas)**:
```
cmd/
├── api/              # Entrypoint REST API
└── grpc/             # Entrypoint gRPC

internal/
├── api/              # API Layer (HTTP handlers, gRPC services)
├── application/      # Application Layer (Use Cases, CQRS Commands/Queries)
├── domain/           # Domain Layer (Entities, Value Objects, Domain Services)
└── infrastructure/   # Infrastructure Layer (PostgreSQL, Redis, Pulsar)

pkg/
└── contracts/        # Proto files (gerados de dict-contracts)
```

**Features**:
- CRUD chaves PIX (CPF, CNPJ, Email, Telefone, EVP)
- Validações de negócio (Bacen rules)
- CQRS + Event Sourcing (Pulsar)
- Cache (Redis - 5 estratégias)
- REST API + gRPC Server
- gRPC Client para Connect

---

### 2. conn-dict (RSFN Connect)

**Tecnologias**: Go 1.24.5, Temporal v1.36.0, Pulsar v0.16.0, gRPC

**Responsabilidade**: Orquestração de workflows de longa duração

**Estrutura**:
```
cmd/
├── worker/           # Temporal worker
└── api/              # Admin API

internal/
├── workflows/        # ClaimWorkflow, VSYNCWorkflow
├── activities/       # Temporal activities
├── pulsar/           # Consumer/Producer
└── grpc/             # gRPC Client (Bridge) + Server (Core)

pkg/
└── contracts/        # Proto files
```

**Features**:
- **ClaimWorkflow**: 30 dias, 3 cenários (confirm, cancel, expire)
- **VSYNCWorkflow**: Sincronização diária com Bacen
- Pulsar Consumer (eventos do Core)
- Pulsar Producer (respostas para Core)
- gRPC Client (chama Bridge)
- gRPC Server (recebe chamadas do Core)
- Error handling, retry policies, circuit breaker

---

### 3. conn-bridge (RSFN Bridge)

**Tecnologias**: Go 1.24.5, Java 17 (XML Signer), gRPC, mTLS

**Responsabilidade**: Adapter SOAP/REST com assinatura XML e mTLS

**Estrutura**:
```
cmd/
├── bridge/           # Main service
└── xml-signer/       # Java XML Signer (separado)

internal/
├── grpc/             # gRPC Server (recebe chamadas do Connect)
├── soap/             # SOAP envelope generator
├── bacen/            # REST client com mTLS
└── xmlsigner/        # Client para XML Signer (Java)

xml-signer/           # Java 17 + ICP-Brasil A3
├── src/
├── pom.xml
└── Dockerfile

pkg/
└── contracts/        # Proto files
```

**Features**:
- gRPC Server (GRPC-001)
- SOAP envelope generation
- Chamar XML Signer (Java) via REST
- REST client com mTLS para Bacen
- ICP-Brasil A3 certificates
- Validação de XML

---

### 4. dict-contracts (Proto Files Compartilhados)

**Estrutura**:
```
proto/
├── core_dict.proto   # Contrato Core ↔ Frontend
├── bridge.proto      # Contrato Connect ↔ Bridge
└── common.proto      # Tipos compartilhados

gen/
├── go/               # Código Go gerado
└── README.md
```

**Importância**: Único source of truth para contratos gRPC entre os 3 repos.

---

### 5. dict-e2e-tests (Criado na Sprint 6)

**Estrutura**:
```
tests/
├── createentry/
├── claimworkflow/
└── vsync/

docker-compose.yml    # Sobe 3 repos + infraestrutura
```

**Features**: Testes E2E completos (Frontend → Core → Connect → Bridge → Bacen Simulator)

---

## 📐 Ordem de Implementação (Bottom-Up)

### **Fase A: Sprint 1-3 (Semanas 1-6) - Bridge + Connect em Paralelo**

**Por quê Bridge e Connect primeiro?**
- São chamados pelo Core (dependencies)
- Core pode testar contra serviços reais
- Reduz risco de desalinhamento de contratos
- Permite paralelismo máximo (2 repos simultâneos)

### **Fase B: Sprint 4-6 (Semanas 7-12) - Core DICT**

**Por quê Core depois?**
- Pode testar contra Bridge e Connect já funcionais
- Validar contratos gRPC com serviços reais
- E2E tests podem ser executados desde Sprint 4

---

## 🗓️ Cronograma Detalhado

---

## **Sprint 1 (Semanas 1-2): Setup Bridge + Connect**

### **Objetivos Sprint**
- ✅ Repos conn-bridge e conn-dict deployáveis
- ✅ dict-contracts com proto files completos
- ✅ Docker Compose funcionando
- ✅ CI/CD básico configurado

### **Tarefas (MÁXIMO PARALELISMO - 8 agentes)**

#### **backend-bridge + xml-specialist** (2 agentes)
```yaml
Semana 1:
- [ ] Setup repo conn-bridge (estrutura Go)
- [ ] Copiar XML Signer dos repos existentes (via MCP)
- [ ] Dockerfile XML Signer (Java 17)
- [ ] gRPC server skeleton (GRPC-001)

Semana 2:
- [ ] SOAP envelope generator
- [ ] Integração XML Signer (REST client)
- [ ] Unit tests (>50% cobertura)
```

#### **backend-connect + temporal-specialist** (2 agentes)
```yaml
Semana 1:
- [ ] Setup repo conn-dict (estrutura Go)
- [ ] Temporal server setup (docker-compose.yml)
- [ ] ClaimWorkflow skeleton (sem timer)
- [ ] gRPC client para Bridge (skeleton)

Semana 2:
- [ ] Pulsar setup (docker-compose.yml)
- [ ] Pulsar consumer básico
- [ ] gRPC server skeleton (recebe chamadas do Core)
- [ ] Unit tests (>50% cobertura)
```

#### **api-specialist** (1 agente)
```yaml
Semana 1:
- [ ] Criar repo dict-contracts
- [ ] Proto files completos (core_dict.proto, bridge.proto, common.proto)
- [ ] Gerar código Go para os 3 repos
- [ ] README com instruções de uso

Semana 2:
- [ ] Validação de contratos (buf lint)
- [ ] Versionamento proto files
- [ ] Publicar gen/go/ para importação
```

#### **data-specialist** (1 agente)
```yaml
Semana 1:
- [ ] PostgreSQL schema para Connect (DAT-002)
- [ ] Migrations (Goose) para Connect
- [ ] Redis setup (docker-compose.yml)

Semana 2:
- [ ] Índices PostgreSQL
- [ ] Redis cache config
- [ ] Scripts de seed (dados de teste)
```

#### **devops-lead** (1 agente)
```yaml
Semana 1:
- [ ] Dockerfile conn-bridge (multi-stage)
- [ ] Dockerfile conn-dict (multi-stage)
- [ ] docker-compose.yml conn-bridge
- [ ] docker-compose.yml conn-dict

Semana 2:
- [ ] CI/CD básico (GitHub Actions) - lint, test, build
- [ ] Makefile para cada repo
- [ ] Scripts de setup (.env.example)
```

#### **security-specialist** (1 agente)
```yaml
Semana 1:
- [ ] mTLS config (SEC-001) - dev mode
- [ ] Certificados ICP-Brasil (self-signed para dev)
- [ ] Vault setup (docker-compose.yml)

Semana 2:
- [ ] Secret management (Vault integration)
- [ ] JWT config (SEC-004)
- [ ] Network security (docker networks)
```

#### **qa-lead** (1 agente)
```yaml
Semana 1:
- [ ] Test cases Bridge (TST-003)
- [ ] Test cases Connect (TST-002)
- [ ] Setup test framework (Go testing, testify)

Semana 2:
- [ ] Unit tests Bridge (helpers)
- [ ] Unit tests Connect (helpers)
- [ ] Integration test skeleton
```

### **Entregáveis Sprint 1**
- ✅ conn-bridge deployável com Docker
- ✅ conn-dict deployável com Docker
- ✅ dict-contracts com proto files
- ✅ PostgreSQL + Redis + Temporal + Pulsar + Vault rodando
- ✅ CI/CD básico (lint, test, build)
- ✅ >50% cobertura unit tests

### **Definition of Done Sprint 1**
- [ ] `make build` funciona em ambos repos
- [ ] `docker-compose up` sobe todos serviços
- [ ] gRPC servers respondem a health checks
- [ ] Tests passam: `make test`
- [ ] Lint passa: `make lint`
- [ ] CI/CD verde no GitHub Actions

---

## **Sprint 2 (Semanas 3-4): Bridge + Connect Funcionais**

### **Objetivos Sprint**
- ✅ XML Signer funcional (ICP-Brasil A3)
- ✅ ClaimWorkflow completo (30 dias, 3 cenários)
- ✅ VSYNC workflow básico
- ✅ Integration tests passando

### **Tarefas (MÁXIMO PARALELISMO - 7 agentes)**

#### **backend-bridge + xml-specialist** (2 agentes)
```yaml
Semana 3:
- [ ] XML Signer funcional (copiar de repos existentes)
- [ ] ICP-Brasil A3 integration (dev certs)
- [ ] mTLS com Bacen dev simulator
- [ ] SOAP → REST adapter completo

Semana 4:
- [ ] Error handling (GRPC-004)
- [ ] Retry policies (circuit breaker)
- [ ] Logs estruturados (slog)
- [ ] Unit tests (>80% cobertura)
```

#### **backend-connect + temporal-specialist** (2 agentes)
```yaml
Semana 3:
- [ ] ClaimWorkflow completo (3 cenários):
  - ConfirmClaim (cancela timer)
  - CancelClaim (cancela timer)
  - ExpireClaim (timer 30 dias expira)
- [ ] VSYNCWorkflow básico (sem lógica Bacen ainda)

Semana 4:
- [ ] Pulsar integration completa (consumer + producer)
- [ ] Error handling workflows
- [ ] Temporal UI configurado
- [ ] Unit tests workflows (>70% cobertura)
```

#### **api-specialist** (1 agente)
```yaml
Semana 3-4:
- [ ] Validação contratos gRPC (contract testing)
- [ ] OpenAPI specs (API-004)
- [ ] gRPC interceptors (logging, metrics)
- [ ] Documentação APIs (README)
```

#### **devops-lead** (1 agente)
```yaml
Semana 3:
- [ ] CI/CD completo (DEV-001, DEV-003):
  - Build multi-arch (amd64, arm64)
  - Docker push (registry)
  - Deploy dev (opcional)

Semana 4:
- [ ] Kubernetes manifests básicos (DEV-004)
- [ ] Helm charts (opcional)
- [ ] Monitoring básico (Prometheus)
```

#### **security-specialist** (1 agente)
```yaml
Semana 3:
- [ ] Secret rotation (Vault)
- [ ] LGPD compliance check (SEC-007)
- [ ] Security scan (Trivy, gosec)

Semana 4:
- [ ] Network policies (Kubernetes)
- [ ] Security headers (APIs)
- [ ] Audit logs (estrutura)
```

#### **qa-lead** (1 agente)
```yaml
Semana 3:
- [ ] Integration tests Bridge ↔ Bacen Simulator
- [ ] Integration tests Connect ↔ Bridge

Semana 4:
- [ ] Integration tests ClaimWorkflow (Temporal)
- [ ] Performance tests básicos (load testing)
- [ ] Code coverage report (>80%)
```

### **Entregáveis Sprint 2**
- ✅ Bridge funcional com XML Signer + mTLS
- ✅ Connect com ClaimWorkflow 30 dias + VSYNC
- ✅ Integration tests passando
- ✅ >80% cobertura unit tests
- ✅ CI/CD completo (build + test + push)

### **Definition of Done Sprint 2**
- [ ] XML Signer assina XML corretamente
- [ ] mTLS com Bacen Simulator funciona
- [ ] ClaimWorkflow completa 3 cenários (testados)
- [ ] Integration tests passam (Bridge ↔ Connect)
- [ ] Performance: >100 TPS (Bridge + Connect)
- [ ] CI/CD: Build + Test + Push automático

---

## **Sprint 3 (Semanas 5-6): Bridge + Connect Prontos**

### **Objetivos Sprint**
- ✅ Bridge + Connect prontos para integração com Core
- ✅ E2E tests passando (Bridge + Connect)
- ✅ Observability completa
- ✅ Documentação completa

### **Tarefas (PARALELISMO - 6 agentes)**

#### **backend-bridge + backend-connect** (2 agentes)
```yaml
Semana 5-6:
- [ ] Ajustes finais (code review)
- [ ] Performance tuning
- [ ] Refactoring (se necessário)
- [ ] Observability (Prometheus metrics, Jaeger tracing)
```

#### **temporal-specialist** (1 agente)
```yaml
Semana 5-6:
- [ ] Temporal UI completo (workflow visualization)
- [ ] Monitoring workflows (alertas)
- [ ] Error recovery tests
```

#### **xml-specialist** (1 agente)
```yaml
Semana 5-6:
- [ ] XML validation completa (casos reais Bacen)
- [ ] Performance XML Signer (>50 assinaturas/seg)
- [ ] Testes com certificados A3 reais (se disponível)
```

#### **devops-lead** (1 agente)
```yaml
Semana 5:
- [ ] Observability stack (DEV-005):
  - Prometheus + Grafana
  - Jaeger (tracing)
  - Loki (logs)

Semana 6:
- [ ] Helm charts completos
- [ ] Disaster recovery (backup/restore)
```

#### **qa-lead** (1 agente)
```yaml
Semana 5:
- [ ] E2E tests Bridge + Connect
- [ ] Performance tests (TST-004):
  - >500 TPS (Bridge + Connect)
  - Latency P99 < 500ms

Semana 6:
- [ ] Security tests (TST-005):
  - Pen testing básico
  - Vulnerability scan
- [ ] Load testing (stress test)
```

### **Entregáveis Sprint 3**
- ✅ **Bridge + Connect PRONTOS para Core**
- ✅ E2E tests Bridge + Connect passando
- ✅ Performance: >500 TPS
- ✅ Observability completa (Prometheus, Grafana, Jaeger)
- ✅ Documentação completa (README, API docs)

### **Definition of Done Sprint 3**
- [ ] E2E tests passam (Bridge + Connect sem Core)
- [ ] Performance: >500 TPS, P99 < 500ms
- [ ] Security scan: 0 vulnerabilidades críticas
- [ ] Code coverage: >80%
- [ ] Documentação: README completo, API docs
- [ ] Observability: Dashboards Grafana funcionando

---

## **Sprint 4 (Semanas 7-8): Core DICT Setup + Integração**

### **Objetivos Sprint**
- ✅ Core DICT deployável
- ✅ CRUD chaves PIX funcionando
- ✅ Integração Core → Connect básica

### **Tarefas (MÁXIMO PARALELISMO - 8 agentes)**

#### **backend-core** (1 agente)
```yaml
Semana 7:
- [ ] Setup repo core-dict (Clean Architecture)
- [ ] Domain layer:
  - Entities: Entry, Claim, Account
  - Value Objects: KeyType, KeyValue, Status
  - Domain Services

Semana 8:
- [ ] API layer:
  - REST API (Fiber) - API-002
  - gRPC Server (recebe Frontend)
  - gRPC Client (chama Connect)
```

#### **data-specialist** (1 agente)
```yaml
Semana 7:
- [ ] PostgreSQL schema Core (DAT-001):
  - entries, claims, accounts
  - Índices, particionamento
  - RLS (Row Level Security)

Semana 8:
- [ ] Migrations (Goose) - DAT-003
- [ ] Redis cache (DAT-005):
  - 5 estratégias (read-through, write-through, etc.)
- [ ] Scripts de seed
```

#### **api-specialist** (1 agente)
```yaml
Semana 7:
- [ ] REST API spec (API-002):
  - POST /entries
  - GET /entries/{id}
  - DELETE /entries/{id}
  - POST /claims

Semana 8:
- [ ] gRPC Server (GRPC-002)
- [ ] OpenAPI complete
- [ ] Contract testing (Core ↔ Connect)
```

#### **backend-connect** (1 agente - ajustes)
```yaml
Semana 7-8:
- [ ] Ajustes para receber chamadas do Core
- [ ] gRPC Server (recebe Core)
- [ ] Pulsar Consumer (eventos do Core)
```

#### **security-specialist** (1 agente)
```yaml
Semana 7:
- [ ] JWT/OAuth (SEC-004)
- [ ] RBAC implementation

Semana 8:
- [ ] API authentication (JWT tokens)
- [ ] Authorization (RBAC)
```

#### **devops-lead** (1 agente)
```yaml
Semana 7:
- [ ] Dockerfile Core DICT
- [ ] docker-compose.yml Core

Semana 8:
- [ ] CI/CD Core (DEV-001)
- [ ] Kubernetes deployment Core
```

#### **qa-lead** (1 agente)
```yaml
Semana 7:
- [ ] Test cases Core (TST-001)
- [ ] Unit tests (business logic)

Semana 8:
- [ ] Integration tests Core ↔ Connect
- [ ] Cache tests (Redis)
```

#### **squad-lead** (1 agente - code review)
```yaml
Semana 7-8:
- [ ] Code review Core (Clean Architecture)
- [ ] Validar padrões Go
- [ ] Validar SOLID, DDD
```

### **Entregáveis Sprint 4**
- ✅ Core DICT deployável
- ✅ CRUD chaves PIX funcionando
- ✅ Integração Core → Connect básica
- ✅ >70% cobertura unit tests

### **Definition of Done Sprint 4**
- [ ] `make build` funciona (core-dict)
- [ ] REST API responde: POST /entries, GET /entries/{id}
- [ ] gRPC Server funciona (health check)
- [ ] Integration test passa: Core → Connect → Bridge
- [ ] Tests passam: `make test`
- [ ] Code coverage: >70%

---

## **Sprint 5 (Semanas 9-10): Core DICT Completo (CQRS)**

### **Objetivos Sprint**
- ✅ Core DICT completo (CQRS + Event Sourcing)
- ✅ Business rules (validações Bacen)
- ✅ Integration tests Core ↔ Connect ↔ Bridge

### **Tarefas (PARALELISMO - 7 agentes)**

#### **backend-core** (1 agente)
```yaml
Semana 9:
- [ ] Application layer (CQRS):
  - Command handlers (CreateEntry, DeleteEntry, CreateClaim)
  - Query handlers (GetEntry, ListEntries, ListClaims)
  - Use Cases

Semana 10:
- [ ] Infrastructure layer:
  - PostgreSQL repository (completo)
  - Redis cache (completo)
  - Pulsar producer (eventos)
- [ ] Business rules (validações Bacen):
  - Limite de chaves por CPF/CNPJ
  - Validação de formatos
  - Regras de Claim
```

#### **data-specialist** (1 agente)
```yaml
Semana 9:
- [ ] Índices otimizados (PostgreSQL)
- [ ] Particionamento (por data/tipo)
- [ ] Query optimization

Semana 10:
- [ ] RLS (Row Level Security) completo
- [ ] Performance tuning PostgreSQL
- [ ] Cache optimization (Redis)
```

#### **backend-connect** (1 agente)
```yaml
Semana 9-10:
- [ ] Pulsar Consumer para eventos do Core
- [ ] Processar eventos assíncronos
- [ ] Enviar respostas via Pulsar para Core
```

#### **api-specialist** (1 agente)
```yaml
Semana 9-10:
- [ ] OpenAPI complete (API-002)
- [ ] Contract testing (todos contratos validados)
- [ ] API versioning (v1, v2)
```

#### **devops-lead** (1 agente)
```yaml
Semana 9:
- [ ] Kubernetes deployment Core (completo)
- [ ] Auto-scaling config (HPA)

Semana 10:
- [ ] Observability Core (Prometheus, Grafana)
- [ ] Alertas (high latency, errors)
```

#### **qa-lead** (1 agente)
```yaml
Semana 9:
- [ ] Integration tests Core ↔ Connect ↔ Bridge
- [ ] Cache tests (invalidação, TTL)

Semana 10:
- [ ] E2E tests parciais (sem Frontend ainda)
- [ ] Performance tests Core (>500 TPS)
```

#### **squad-lead** (1 agente)
```yaml
Semana 9-10:
- [ ] Code review final Core
- [ ] Validar CQRS implementation
- [ ] Validar Event Sourcing
```

### **Entregáveis Sprint 5**
- ✅ Core DICT completo (CQRS + Event Sourcing)
- ✅ Business rules Bacen implementadas
- ✅ Integration tests Core ↔ Connect ↔ Bridge passando
- ✅ >80% cobertura unit tests

### **Definition of Done Sprint 5**
- [ ] CQRS funcionando (Command/Query separados)
- [ ] Eventos publicados no Pulsar
- [ ] Integration test passa: Core → Connect → Bridge → Bacen Simulator
- [ ] Performance: >500 TPS (Core)
- [ ] Code coverage: >80%

---

## **Sprint 6 (Semanas 11-12): E2E + Performance + Finalização**

### **Objetivos Sprint**
- ✅ **3 REPOS COMPLETOS E TESTADOS**
- ✅ E2E tests completo (Frontend → Bacen)
- ✅ Performance: >1000 TPS
- ✅ Documentação final
- ✅ Prontos para homologação Bacen

### **Tarefas (PARALELISMO TOTAL - 10 agentes)**

#### **backend-core + backend-connect + backend-bridge** (3 agentes)
```yaml
Semana 11-12:
- [ ] Ajustes finais cross-repo
- [ ] Performance optimization
- [ ] Refactoring final
- [ ] Code freeze (após code review final)
```

#### **api-specialist** (1 agente)
```yaml
Semana 11-12:
- [ ] Validação final de contratos
- [ ] Versionamento APIs (v1 estável)
- [ ] API documentation final
```

#### **data-specialist** (1 agente)
```yaml
Semana 11-12:
- [ ] Tuning PostgreSQL (final)
- [ ] Cache optimization (final)
- [ ] Backup/Restore strategy
```

#### **temporal-specialist** (1 agente)
```yaml
Semana 11-12:
- [ ] Workflow monitoring (alertas)
- [ ] Error recovery tests (chaos engineering)
- [ ] Temporal UI polimento
```

#### **xml-specialist** (1 agente)
```yaml
Semana 11-12:
- [ ] Validação XML compliance Bacen (final)
- [ ] Performance XML Signer (>100 assinaturas/seg)
- [ ] Testes com certificados A3 reais
```

#### **security-specialist** (1 agente)
```yaml
Semana 11-12:
- [ ] Security audit final
- [ ] LGPD compliance final (SEC-007)
- [ ] Pen testing completo
- [ ] Vulnerability scan (0 críticas)
```

#### **devops-lead** (1 agente)
```yaml
Semana 11:
- [ ] Observability completa (Prometheus, Grafana, Jaeger, Loki)
- [ ] Dashboards Grafana finalizados
- [ ] Alertas configurados

Semana 12:
- [ ] Disaster recovery (backup/restore testados)
- [ ] Runbooks (playbooks operacionais)
- [ ] Infrastructure as Code (Terraform - opcional)
```

#### **qa-lead** (1 agente)
```yaml
Semana 11:
- [ ] Criar repo dict-e2e-tests
- [ ] E2E tests completo (Frontend → Bacen):
  - CreateEntry E2E
  - ClaimWorkflow E2E (30 dias simulados)
  - VSYNC E2E
- [ ] Performance tests (TST-004):
  - >1000 TPS (3 repos)
  - Latency P99 < 200ms

Semana 12:
- [ ] Security tests (TST-005):
  - Pen testing
  - Vulnerability scan
  - OWASP Top 10
- [ ] Load testing (stress test):
  - 2000 TPS (stress)
  - Avaliar bottlenecks
```

#### **squad-lead** (1 agente)
```yaml
Semana 11-12:
- [ ] Code review final (todos repos)
- [ ] Documentação final (README, API docs, runbooks)
- [ ] Preparar pacote de entrega (release notes)
```

### **Entregáveis Sprint 6**
- ✅ **3 REPOS COMPLETOS**:
  - core-dict v1.0.0
  - conn-dict v1.0.0
  - conn-bridge v1.0.0
- ✅ dict-contracts v1.0.0
- ✅ dict-e2e-tests v1.0.0
- ✅ E2E tests passando (>95%)
- ✅ Performance: >1000 TPS, P99 < 200ms
- ✅ Security: 0 vulnerabilidades críticas
- ✅ Code coverage: >80%
- ✅ Documentação completa
- ✅ CI/CD funcionando (todos repos)
- ✅ Observability completa

### **Definition of Done Sprint 6**
- [ ] E2E tests passam (Frontend → Bacen)
- [ ] Performance: >1000 TPS, P99 < 200ms
- [ ] Load test: 2000 TPS sem falhas
- [ ] Security scan: 0 críticas, 0 altas
- [ ] Code coverage: >80% (todos repos)
- [ ] Documentação: README, API docs, runbooks
- [ ] Observability: Dashboards funcionando, alertas configurados
- [ ] **PRONTOS PARA HOMOLOGAÇÃO BACEN**

---

## 📊 Métricas de Progresso

### Métricas por Sprint

| Sprint | Repos | LOC Estimado | Testes Estimados | Cobertura Meta |
|--------|-------|--------------|------------------|----------------|
| Sprint 1 | Bridge, Connect (setup) | ~2k | ~20 | >50% |
| Sprint 2 | Bridge, Connect (funcional) | ~5k | ~60 | >80% |
| Sprint 3 | Bridge, Connect (pronto) | ~6k | ~80 | >80% |
| Sprint 4 | Core (setup) | ~8k | ~100 | >70% |
| Sprint 5 | Core (completo) | ~12k | ~150 | >80% |
| Sprint 6 | Core (final) + E2E | ~15k | ~200 | >80% |

### Métricas Finais Esperadas

| Métrica | Meta | Como Medir |
|---------|------|------------|
| **LOC Go** | ~15k | `cloc --include-lang=Go core-dict/ conn-dict/ conn-bridge/` |
| **LOC Java** | ~2k | `cloc --include-lang=Java conn-bridge/xml-signer/` |
| **Unit Tests** | ~200 | `go test ./... -v | grep -c "PASS"` |
| **Integration Tests** | ~50 | Contar testes em `*_integration_test.go` |
| **E2E Tests** | ~20 | Contar testes em `dict-e2e-tests/` |
| **Code Coverage** | >80% | `go test -cover ./...` |
| **Performance (TPS)** | >1000 | k6, JMeter ou hey |
| **Latency P99** | <200ms | Prometheus + Grafana |
| **Vulnerabilities** | 0 críticas | Trivy, gosec, Snyk |

---

## 🚀 Princípio de Máximo Paralelismo

### Como Garantir Paralelismo

**1. Identificar Dependências**:
- Tarefas sem dependências → **executar em paralelo**
- Tarefas com dependências → executar sequencialmente

**Exemplo Sprint 1**:
```
Independentes (PARALELO):
- backend-bridge: Setup repo
- backend-connect: Setup repo
- api-specialist: Criar dict-contracts
- data-specialist: PostgreSQL schemas
- devops-lead: Dockerfiles
- security-specialist: mTLS config
- qa-lead: Test cases

Dependente (SEQUENCIAL):
- api-specialist cria proto files
  → DEPOIS → backend-bridge/connect importam proto files
```

**2. Balanceamento de Carga**:
- Distribuir tarefas entre agentes
- Evitar agentes ociosos
- Ajustar se algum agente estiver sobrecarregado

**3. Sincronização Diária**:
- Daily standups (async)
- Revisar progresso
- Rebalancear tarefas
- Resolver bloqueios imediatamente

---

## ✅ Critérios de Sucesso Fase 2

### Por Sprint
- **Sprint 1**: Bridge + Connect deployáveis ✅
- **Sprint 2**: Bridge + Connect funcionais ✅
- **Sprint 3**: Bridge + Connect prontos ✅
- **Sprint 4**: Core deployável ✅
- **Sprint 5**: Core completo (CQRS) ✅
- **Sprint 6**: **3 REPOS PRONTOS** ✅

### Finais
- ✅ 3 repos funcionais, testados, performantes
- ✅ E2E tests: >95% passando
- ✅ Performance: >1000 TPS, P99 < 200ms
- ✅ Security: 0 vulnerabilidades críticas
- ✅ Code coverage: >80%
- ✅ Documentação completa
- ✅ CI/CD funcionando
- ✅ Observability completa
- ✅ **PRONTOS PARA HOMOLOGAÇÃO BACEN**

---

## 📝 Gestão de Progresso

### Documentos de Tracking (em `00_Master/`)

**PROGRESSO_IMPLEMENTACAO.md**:
- Atualizado **DIARIAMENTE** pelo Project Manager
- Status por sprint, por repo
- Tarefas completadas, em progresso, bloqueios
- Métricas (LOC, testes, cobertura)

**BACKLOG_IMPLEMENTACAO.md**:
- Todas as tarefas pendentes
- Priorização (P0, P1, P2)
- Dependências entre tarefas
- Atribuição de agentes

**METRICAS_IMPLEMENTACAO.md**:
- Métricas consolidadas por sprint
- Gráficos de progresso (burndown)
- Comparação meta vs real

---

## 🔗 Referências

### Especificações (Fase 1)
- [TEC-001](../11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md)
- [TEC-002 v3.1](../11_Especificacoes_Tecnicas/TEC-002_Bridge_Specification.md)
- [TEC-003 v2.1](../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)

### Manuais de Implementação
- [IMP-001](../09_Implementacao/IMP-001_Manual_Implementacao_Core_DICT.md)
- [IMP-002](../09_Implementacao/IMP-002_Manual_Implementacao_Connect.md)
- [IMP-003](../09_Implementacao/IMP-003_Manual_Implementacao_Bridge.md)

### Dados, APIs, DevOps, Testes
- [DAT-001 a DAT-005](../03_Dados/)
- [GRPC-001 a GRPC-004](../04_APIs/gRPC/)
- [API-002 a API-004](../04_APIs/REST/)
- [DEV-001 a DEV-007](../15_DevOps/)
- [TST-001 a TST-006](../14_Testes/Casos/)

---

**Próximo Passo**: Executar Sprint 1 com **MÁXIMO PARALELISMO** (8 agentes simultâneos).

**Última Atualização**: 2025-10-26
**Status**: 🚀 Pronto para Iniciar Sprint 1
**Versão**: 1.0
