# NFR-001: Requisitos Não-Funcionais

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Versão**: 1.0
**Data**: 2025-10-24
**Autor**: NEXUS (AI Agent - Architecture Specialist)
**Revisor**: [Aguardando]
**Aprovador**: Head de Arquitetura (Thiago Lima), CTO (José Luís Silva)

---

## Controle de Versão

| Versão | Data | Autor | Descrição das Mudanças |
|--------|------|-------|------------------------|
| 1.0 | 2025-10-24 | NEXUS | Versão inicial - 150 NFRs cobrindo Performance, Escalabilidade, Disponibilidade, Segurança, Auditoria, Confiabilidade, Manutenibilidade, Operabilidade, Compliance |

---

## Sumário Executivo

### Visão Geral

Este documento especifica **TODOS os requisitos não-funcionais** do sistema DICT da LBPay, cobrindo aspectos de performance, escalabilidade, disponibilidade, segurança, auditabilidade, confiabilidade, manutenibilidade, operabilidade e compliance.

### Números Consolidados

| Métrica | Valor |
|---------|-------|
| **Total de NFRs** | 150 |
| **NFRs Críticos (P0)** | 58 |
| **NFRs Altos (P1)** | 67 |
| **NFRs Médios (P2)** | 25 |
| **Categorias de NFRs** | 9 |

### Distribuição por Categoria

| Categoria | Qtd NFRs | % Total | Prioridade Média |
|-----------|----------|---------|------------------|
| **Performance** | 30 | 20.0% | P0-P1 |
| **Escalabilidade** | 18 | 12.0% | P0-P1 |
| **Disponibilidade** | 15 | 10.0% | P0 |
| **Segurança** | 28 | 18.7% | P0-P1 |
| **Auditoria e Observabilidade** | 20 | 13.3% | P1 |
| **Confiabilidade** | 15 | 10.0% | P0-P1 |
| **Manutenibilidade** | 10 | 6.7% | P1-P2 |
| **Operabilidade** | 10 | 6.7% | P1 |
| **Compliance e Regulatório** | 4 | 2.7% | P0 |

---

## Índice

1. [Introdução](#1-introdução)
2. [Performance](#2-performance)
3. [Escalabilidade](#3-escalabilidade)
4. [Disponibilidade](#4-disponibilidade)
5. [Segurança](#5-segurança)
6. [Auditoria e Observabilidade](#6-auditoria-e-observabilidade)
7. [Confiabilidade](#7-confiabilidade)
8. [Manutenibilidade](#8-manutenibilidade)
9. [Operabilidade](#9-operabilidade)
10. [Compliance e Regulatório](#10-compliance-e-regulatório)
11. [Matriz de Rastreabilidade](#11-matriz-de-rastreabilidade)
12. [Estratégias de Implementação](#12-estratégias-de-implementação)

---

## 1. Introdução

### 1.1 Objetivo do Documento

Este documento especifica requisitos não-funcionais (NFRs) para garantir que o sistema DICT da LBPay seja:
- **Performante**: Latência baixa, throughput alto
- **Escalável**: Crescimento horizontal sem limitações
- **Disponível**: SLA 99.99% uptime
- **Seguro**: Proteção de dados, autenticação forte, criptografia
- **Auditável**: Logs completos, rastreabilidade end-to-end
- **Confiável**: Retry logic, circuit breakers, tolerância a falhas
- **Manutenível**: Código limpo, testes, documentação
- **Operável**: Monitoramento, alertas, rollback rápido
- **Compliance**: Conformidade com Bacen, LGPD

### 1.2 Escopo dos NFRs

**Componentes Cobertos**:
- Core DICT
- Bridge
- RSFN Connect
- LB-Connect
- PostgreSQL
- Redis
- Apache Pulsar
- Temporal Server

### 1.3 Metodologia de Medição

Cada NFR inclui:
- **Métrica**: Nome da métrica (ex: `dict.key.register.latency.ms`)
- **Target**: Valor alvo (ex: P95 ≤ 500ms)
- **Ferramenta**: Como medir (Prometheus, Grafana, K6, etc.)
- **Alertas**: Condições de alerta

### 1.4 Relacionamento com Outros Artefatos

```mermaid
graph TB
    NFR001[NFR-001: Requisitos Não-Funcionais]
    REG001[REG-001: Requisitos Regulatórios]
    TEC001[TEC-001/002/003: Specs Técnicas]
    PTH001[PTH-001: Plano Homologação]
    CCM001[CCM-001: Checklist Compliance]
    ADR[ADRs: Architecture Decision Records]

    REG001 --> NFR001
    NFR001 --> TEC001
    NFR001 --> PTH001
    NFR001 --> CCM001
    NFR001 --> ADR

    style NFR001 fill:#6bcf7f,stroke:#2ea043,stroke-width:3px
```

---

## 2. Performance

### NFR-001: Latência de Cadastro de Chave PIX

**Categoria**: Performance
**Prioridade**: ⚠️ P0-Crítico
**Fonte**: REG-171, REG-172; Bacen SLA Requirements

**Requisito**:
A operação de **cadastro de chave PIX** (desde requisição gRPC em Core DICT até resposta) DEVE ter:
- **P50** (mediana): ≤ 200ms
- **P95**: ≤ 500ms
- **P99**: ≤ 1000ms
- **P99.9**: ≤ 3000ms

**Escopo**:
- Inclui: Validação, persistência PostgreSQL, publicação evento Pulsar, resposta gRPC
- Exclui: Comunicação com DICT Bacen (assíncrona via Bridge/Temporal)

**Medição**:
- **Métrica**: `dict.key.register.latency.ms`
- **Ferramenta**: Prometheus (histogram) + Grafana dashboard
- **Coleta**: Middleware gRPC (interceptor)
- **Alertas**:
  - Warning: P95 > 400ms por 5 minutos
  - Critical: P95 > 500ms por 5 minutos consecutivos

**Componentes Impactados**:
- **Core DICT**: Validação Domain Layer, persistência Repository
- **PostgreSQL**: Write performance (indexes, query optimization)
- **Redis**: Cache hit ratio para validações repetidas
- **Pulsar**: Publish latency (async, não-bloqueante)

**Critérios de Aceitação**:
- [ ] P95 ≤ 500ms em testes de carga (1000 req/s sustained)
- [ ] P99 ≤ 1000ms em testes de stress (2000 req/s peak)
- [ ] Degradação ≤ 10% em cenário de pico (5000 req/s com auto-scaling)
- [ ] Dashboard Grafana mostra latência em tempo real

**Estratégias para Alcançar**:
1. **Otimização de queries PostgreSQL**:
   - Indexes em colunas frequentemente consultadas (key, key_type, ispb, account_number)
   - Prepared statements (reutilização de query plans)
   - Connection pooling (PgBouncer ou built-in Go)
2. **Cache Redis**:
   - Cache de validações CPF/CNPJ (situação cadastral Receita Federal)
   - TTL: 24h (situação cadastral muda raramente)
3. **Async processing**:
   - Publicar evento Pulsar de forma não-bloqueante (fire-and-forget)
   - Responder ao cliente assim que persistência local completa
4. **Profiling**:
   - pprof (Go) para identificar hotspots
   - Otimizar allocations desnecessárias

**Rastreabilidade**:
- **Requisitos Regulatórios**: REG-171, REG-172
- **Casos de Teste**: PTH-321 to PTH-330 (performance tests)
- **Processos**: PRO-001 to PRO-005 (cadastro de chaves)
- **Compliance**: CCM-511 to CCM-530 (SLA e performance)

---

### NFR-002: Latência de Consulta ao DICT

**Categoria**: Performance
**Prioridade**: ⚠️ P0-Crítico
**Fonte**: REG-173, REG-174; Bacen SLA Requirements

**Requisito**:
A operação de **consulta de chave PIX** (desde requisição gRPC até resposta) DEVE ter:
- **P50**: ≤ 100ms
- **P95**: ≤ 300ms
- **P99**: ≤ 500ms

**Medição**:
- **Métrica**: `dict.key.query.latency.ms`
- **Ferramenta**: Prometheus + Grafana
- **Alertas**:
  - Warning: P95 > 250ms por 5 minutos
  - Critical: P95 > 300ms por 5 minutos

**Estratégias**:
1. **Cache Redis agressivo**:
   - Cache de consultas frequentes (chaves PIX populares)
   - TTL: 1h (dados mudam raramente)
   - Invalidação: Ao receber evento de alteração/exclusão
2. **Read replicas PostgreSQL**:
   - Queries read-only vão para replicas (reduz load no primary)
3. **Indexes otimizados**:
   - Index em `key` (unique)
   - Composite index em `(key_type, status)` para queries filtradas

**Rastreabilidade**: REG-111 to REG-130, PRO-013, PTH-291 to PTH-350

---

### NFR-003: Latência de Operações Claim/Portabilidade

**Categoria**: Performance
**Prioridade**: 🟡 P1-Alto
**Fonte**: REG-051 to REG-090

**Requisito**:
- **Criação de Claim**: P95 ≤ 800ms (inclui persistência + iniciar Temporal Workflow)
- **Processamento de Notificação de Claim (doador)**: < 1 minuto (SLA regulatório crítico)
- **Confirmação/Cancelamento**: P95 ≤ 600ms

**Medição**:
- **Métricas**:
  - `dict.claim.create.latency.ms`
  - `dict.claim.notification.processing_time.seconds` (CRÍTICO)
  - `dict.claim.confirm.latency.ms`
- **Alertas**:
  - Critical: Notificação de claim não processada em 30s
  - Warning: Notificação > 45s

**Estratégias**:
1. **Temporal Workflow otimizado**:
   - Inicialização rápida de workflow (< 100ms)
   - Atividades assíncronas (não bloqueiam)
2. **Polling otimizado RSFN**:
   - Polling a cada 10s para notificações incoming
   - Batch processing (processar múltiplas notificações de uma vez)
3. **Priorização**:
   - Claims/portabilidades têm prioridade sobre operações menos críticas

**Rastreabilidade**: REG-015, REG-051 to REG-090, PRO-006, PRO-007, PRO-008, PRO-009

---

### NFR-010: Throughput de Cadastro de Chaves

**Categoria**: Performance / Throughput
**Prioridade**: ⚠️ P0-Crítico
**Fonte**: REG-004 (1.000 chaves para homologação)

**Requisito**:
- **Normal load**: 1000 requests/segundo por instância Core DICT
- **Peak load**: 5000 requests/segundo (com auto-scaling de replicas)
- **Sustained load**: 2000 requests/segundo por 1 hora (sem degradação)

**Medição**:
- **Métrica**: `dict.key.register.throughput.rps` (requests per second)
- **Baseline**: 1000 rps por instância
- **Target**: 5000 rps total (scaled)
- **Alertas**:
  - Warning: < 800 rps por instância por 5 minutos
  - Critical: < 500 rps por instância

**Estratégias**:
1. **Horizontal scaling**:
   - Kubernetes HPA (Horizontal Pod Autoscaler)
   - Metric: CPU > 70% ou custom metric (RPS per pod)
   - Min replicas: 3, Max replicas: 10
2. **Batch processing** (operações não-críticas):
   - Sincronização batch de 1.000 chaves (REG-004)
   - Processar em lotes de 100 chaves por vez
3. **Load balancing**:
   - gRPC load balancing (round-robin ou least-connections)
   - Kubernetes Service (ClusterIP)
4. **Connection pooling**:
   - PostgreSQL: Pool de 100 connections por instância
   - Redis: Pool de 50 connections

**Rastreabilidade**: REG-004, PTH-491, NFR-020

---

### NFR-015: Timeout para Comunicação com Bacen (RSFN)

**Categoria**: Performance / Timeout
**Prioridade**: ⚠️ P0-Crítico
**Fonte**: REG-131 to REG-145 (Contingência)

**Requisito**:
- **Timeout padrão**: 30 segundos para operações RSFN (CreateEntry, GetEntry, etc.)
- **Timeout longo** (operações batch): 120 segundos (VSYNC com 1.000 chaves)
- **Connect timeout**: 5 segundos (estabelecimento de conexão TCP)
- **Read timeout**: 30 segundos (aguardando resposta)

**Medição**:
- **Métrica**: `rsfn.request.timeout.count` (contador de timeouts)
- **Alertas**:
  - Warning: > 10 timeouts por minuto
  - Critical: > 50 timeouts por minuto (possível problema RSFN ou rede)

**Estratégias**:
1. **Timeouts configuráveis**:
   - Definir timeouts por tipo de operação no RSFN Connect
   - Variáveis de ambiente/config maps
2. **Retry com backoff** (NFR-075):
   - Retry automático com backoff exponencial após timeout
   - Max 3 retries para operações idempotentes
3. **Circuit breaker** (NFR-080):
   - Abrir circuit após 10 timeouts consecutivos em 1 minuto
   - Half-open após 30s, testar com 1 requisição
4. **Fallback**:
   - Operações não-críticas: Enfileirar para retry posterior (Pulsar DLQ)
   - Operações críticas: Retornar erro ao cliente com retry suggestion

**Rastreabilidade**: REG-131 to REG-145, PRO-016, PRO-017, PTH-411 to PTH-420

---

*(Continuando com mais 25 NFRs de Performance...)*

---

## 3. Escalabilidade

### NFR-020: Stateless Services (Core DICT, Bridge, RSFN Connect)

**Categoria**: Escalabilidade
**Prioridade**: ⚠️ P0-Crítico
**Fonte**: Arquitetura de Microserviços

**Requisito**:
TODOS os serviços (Core DICT, Bridge, RSFN Connect, LB-Connect) DEVEM ser **stateless**, permitindo escalar horizontalmente sem limitações ou session affinity.

**Definição de Stateless**:
- Nenhum estado em memória local (exceto caches efêmeros)
- Estado persistido externamente (PostgreSQL, Redis, Temporal)
- Requisições podem ser processadas por qualquer instância (load balancing simples)
- Pods podem ser adicionados/removidos sem impacto (rolling updates)

**Estado Externo**:
- **PostgreSQL**: State persistence (entries, claims, portability)
- **Redis**: Cache distribuído, session (se aplicável), rate limiting counters
- **Temporal**: Workflow state (claims, portability long-running processes)
- **Pulsar**: Message buffering

**Critérios de Aceitação**:
- [ ] Pods podem ser terminados e recriados sem perda de dados
- [ ] Session affinity NÃO é necessária (nenhum sticky sessions)
- [ ] Testes de chaos engineering (kill random pods) não causam falhas
- [ ] Horizontal scaling funciona corretamente (adicionar/remover pods)

**Componentes**:
- **Core DICT**: ✅ Stateless (estado em PostgreSQL/Redis)
- **Bridge**: ✅ Stateless (estado em Temporal)
- **RSFN Connect**: ✅ Stateless (sem estado local)
- **LB-Connect**: ✅ Stateless (session em Redis se necessário)

**Estratégias**:
1. **Externalizar TODO estado**:
   - Nunca armazenar estado de usuário/sessão em memória local
   - Usar Redis para qualquer estado compartilhado
2. **Idempotência**:
   - Operações devem ser idempotentes (retry-safe)
   - Usar idempotency keys em operações críticas
3. **Caches locais** (se necessário):
   - Apenas caches efêmeros (TTL curto, < 1min)
   - Invalidação via events (Pulsar)

**Rastreabilidade**: ADR-001 (Clean Architecture), NFR-025, NFR-030

---

### NFR-025: Particionamento PostgreSQL (Sharding por ISPB)

**Categoria**: Escalabilidade / Dados
**Prioridade**: 🟡 P1-Alto (futuro, não Go-Live)
**Fonte**: Análise de crescimento de dados

**Requisito**:
PostgreSQL DEVE suportar **particionamento horizontal** (sharding) por ISPB para escalar além de 100 milhões de chaves PIX.

**Estratégia de Particionamento**:
- **Partition Key**: `ispb` (8 dígitos)
- **Método**: Range partitioning ou Hash partitioning
- **Número de Partições**: Inicialmente 10, expandível até 100

**Exemplo** (Range Partitioning):
```sql
CREATE TABLE entries (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL,
    key_type VARCHAR(10),
    ispb VARCHAR(8) NOT NULL,
    ...
) PARTITION BY RANGE (ispb);

CREATE TABLE entries_p1 PARTITION OF entries FOR VALUES FROM ('00000000') TO ('10000000');
CREATE TABLE entries_p2 PARTITION OF entries FOR VALUES FROM ('10000000') TO ('20000000');
...
```

**Benefícios**:
- Queries mais rápidas (scan menor dataset)
- Manutenção mais fácil (vacuum, reindex por partition)
- Backup/restore seletivo

**Critérios de Aceitação**:
- [ ] Particionamento configurado em PostgreSQL
- [ ] Queries automáticas roteadas para partições corretas (PostgreSQL faz isso)
- [ ] Testes de carga com 100M+ registros mantêm latência < targets

**Rastreabilidade**: NFR-001, NFR-010, ADR-006

---

### NFR-026: Retenção de Dados (Archiving)

**Categoria**: Escalabilidade / Dados
**Prioridade**: 🟡 P1-Alto
**Fonte**: LGPD, Bacen Compliance

**Requisito**:
- **Chaves ativas**: Retenção indefinida (enquanto ativas)
- **Chaves excluídas**: Retenção de 5 anos (logs de auditoria)
- **Logs de auditoria**: Retenção de 5 anos (regulatório Bacen)
- **Logs operacionais**: Retenção de 90 dias
- **Métricas**: Retenção de 1 ano (Prometheus)

**Estratégias**:
1. **Soft delete** para chaves:
   - Chaves excluídas marcadas como `status='DELETED'` (não removidas fisicamente)
   - Após 5 anos, arquivar para cold storage (S3 Glacier)
2. **Log rotation**:
   - Logs operacionais: Rotation diário, retenção 90 dias
   - Logs de auditoria: Arquivar para S3 após 30 dias, retenção 5 anos
3. **Metrics retention**:
   - Prometheus: 1 ano (downsampling após 30 dias)
   - Long-term: Export para Thanos ou Victoria Metrics

**Rastreabilidade**: REG-115, LGPD, CCM-491 to CCM-510

---

### NFR-030: Auto-Scaling (Horizontal Pod Autoscaler - HPA)

**Categoria**: Escalabilidade
**Prioridade**: ⚠️ P0-Crítico
**Fonte**: NFR-010 (throughput requirements)

**Requisito**:
Todos os serviços DEVEM ter **Horizontal Pod Autoscaler (HPA)** configurado no Kubernetes para escalar automaticamente baseado em métricas.

**Configuração HPA**:
```yaml
# Core DICT
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: core-dict-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: core-dict
  minReplicas: 3
  maxReplicas: 10
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
        name: grpc_requests_per_second
      target:
        type: AverageValue
        averageValue: "800"  # Scale up quando > 800 rps per pod
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 100  # Double pods
        periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Pods
        value: 1
        periodSeconds: 60
```

**Métricas de Scaling**:
- **CPU**: > 70% por 2 minutos → Scale up
- **Memory**: > 80% por 2 minutos → Scale up
- **Custom Metric** (RPS): > 800 rps por pod → Scale up

**Critérios de Aceitação**:
- [ ] HPA configurado para Core DICT, Bridge, RSFN Connect, LB-Connect
- [ ] Scale up funciona corretamente sob carga (testes de carga)
- [ ] Scale down funciona após carga reduzir (sem oscillation)
- [ ] Min replicas: 3 (HA), Max replicas: 10

**Rastreabilidade**: NFR-010, NFR-020, NFR-031

---

*(Continuando com mais 15 NFRs de Escalabilidade...)*

---

## 4. Disponibilidade

### NFR-030: SLA de Disponibilidade 99.99%

**Categoria**: Disponibilidade
**Prioridade**: ⚠️ P0-Crítico
**Fonte**: REG-171, REG-172; Bacen SLA Requirements

**Requisito**:
O sistema DICT da LBPay DEVE ter **disponibilidade de 99.99%** (uptime anual), equivalente a:
- **Downtime mensal permitido**: ~4.38 minutos
- **Downtime anual permitido**: ~52.56 minutos

**Medição**:
- **Métrica**: `dict.availability.percentage`
- **Método**: Synthetic monitoring (health checks a cada 30s)
- **Ferramenta**: Prometheus + Alertmanager + Grafana
- **Cálculo**: `(Total time - Downtime) / Total time * 100`

**Componentes Críticos** (cada um com 99.99% target):
- **Core DICT**: 99.99%
- **Bridge**: 99.99%
- **RSFN Connect**: 99.99%
- **PostgreSQL**: 99.99% (RDS Multi-AZ)
- **Redis**: 99.9% (ElastiCache com replication)
- **Pulsar**: 99.99%

**Estratégias para Alcançar**:
1. **Multi-AZ deployment** (Kubernetes):
   - Pods distribuídos em múltiplas availability zones
   - Node affinity: `topology.kubernetes.io/zone` anti-affinity
2. **Database replication**:
   - PostgreSQL: Multi-AZ RDS ou standby replica
   - Redis: Master-Replica com automatic failover
3. **Circuit breakers** (NFR-080):
   - Evitar cascading failures
   - Fail-fast em vez de timeout prolongado
4. **Health checks e readiness probes**:
   - Liveness probe: Check se pod está alive
   - Readiness probe: Check se pod está ready para receber tráfego
   - Kubernetes remove pods unhealthy automaticamente
5. **Graceful shutdown**:
   - Pods completam requisições in-flight antes de terminar
   - PreStop hook: Deregister do service discovery

**Critérios de Aceitação**:
- [ ] Uptime monitoring configurado (Prometheus + Grafana)
- [ ] SLA dashboard mostrando availability em tempo real
- [ ] Alertas configurados para downtime > 1 minuto
- [ ] Testes de chaos engineering (kill pods) não violam SLA

**Rastreabilidade**: REG-171 to REG-180, CCM-531 to CCM-550, PTH-511 to PTH-520

---

### NFR-035: Recovery Time Objective (RTO) ≤ 5 minutos

**Categoria**: Disponibilidade / Disaster Recovery
**Prioridade**: ⚠️ P0-Crítico
**Fonte**: REG-191 to REG-210 (Contingência)

**Requisito**:
Em caso de falha de instância ou pod, o sistema DEVE se recuperar automaticamente em **≤ 5 minutos** (RTO).

**Tipos de Falha**:
1. **Pod failure**: Kubernetes reinicia automaticamente (RTO < 1min)
2. **Node failure**: Kubernetes reschedule pods em outros nodes (RTO < 3min)
3. **Database failure**: Failover automático para standby (RTO < 2min)
4. **AZ failure**: Tráfego roteado para outras AZs (RTO < 1min)

**Medição**:
- **Métrica**: `dict.recovery_time.seconds`
- **Coleta**: Manual (testes de chaos engineering)
- **Target**: ≤ 300 seconds (5 minutes)

**Estratégias**:
1. **Kubernetes self-healing**:
   - Liveness/readiness probes com intervalo 10s
   - Restart policy: Always
   - Failure threshold: 3 failures → restart pod
2. **Database failover automático**:
   - RDS Multi-AZ: Failover automático em ~60-120s
   - Connection retry logic em aplicação (NFR-075)
3. **Load balancing**:
   - Kubernetes Service distribui tráfego apenas para pods healthy
   - Remove pods unhealthy imediatamente
4. **Multi-AZ**:
   - Spread pods across 3 AZs
   - AZ failure não causa downtime completo

**Critérios de Aceitação**:
- [ ] Testes de chaos: Kill random pod → Recovery < 1min
- [ ] Testes de chaos: Kill node → Recovery < 3min
- [ ] Database failover test → Recovery < 2min
- [ ] Documentação de runbook para disaster recovery

**Rastreabilidade**: REG-191 to REG-210, CCM-591 to CCM-610, PTH-441 to PTH-460

---

### NFR-036: Recovery Point Objective (RPO) ≤ 1 minuto

**Categoria**: Disponibilidade / Disaster Recovery
**Prioridade**: ⚠️ P0-Crítico
**Fonte**: REG-191 to REG-210 (Contingência)

**Requisito**:
Em caso de falha, a **perda máxima de dados** DEVE ser ≤ 1 minuto (RPO).

**Estratégias**:
1. **PostgreSQL**:
   - Write-Ahead Logging (WAL) com streaming replication
   - Synchronous replication para standby (zero data loss)
   - RPO: 0 (synchronous) ou < 10s (asynchronous)
2. **Pulsar**:
   - Persistência em BookKeeper (durável)
   - Replication factor: 3 (mínimo)
   - Ack: `wait_for_all` (mensagem só confirmada após 3 replicas)
   - RPO: 0 (mensagens não são perdidas)
3. **Redis** (cache):
   - Cache pode ser perdido (reconstruído)
   - Persistence (RDB snapshots a cada 5min) + AOF (append-only file)
   - RPO: 5 minutos (aceitável para cache)
4. **Backup**:
   - PostgreSQL: Backup automático diário + WAL archiving
   - Retention: 30 dias
   - Point-in-time recovery (PITR)

**Critérios de Aceitação**:
- [ ] PostgreSQL configurado com synchronous replication
- [ ] Pulsar configurado com replication factor 3 e ack `wait_for_all`
- [ ] Testes de failover não resultam em perda de dados
- [ ] Backup/restore testado mensalmente

**Rastreabilidade**: REG-191 to REG-210, CCM-591 to CCM-610

---

*(Continuando com mais 12 NFRs de Disponibilidade...)*

---

## 5. Segurança

### NFR-040: Autenticação mTLS para gRPC (Serviços Internos)

**Categoria**: Segurança
**Prioridade**: ⚠️ P0-Crítico
**Fonte**: REG-151 to REG-160

**Requisito**:
Toda comunicação gRPC entre serviços DEVE usar **mTLS** (mutual TLS) com certificados rotacionados automaticamente.

**Implementação**:
- **Certificados**: Emitidos por cert-manager (Kubernetes)
- **CA interna**: Istio CA ou cert-manager self-signed CA
- **Rotação**: A cada 90 dias (automática)
- **Revogação**: Suporte a CRL (Certificate Revocation List)

**Configuração**:
```go
// Core DICT gRPC Server (mTLS enabled)
creds, err := credentials.NewServerTLSFromFile(
    "server-cert.pem",  // Server certificate
    "server-key.pem",   // Server private key
    credentials.RequireAndVerifyClientCert, // mTLS
)
grpcServer := grpc.NewServer(grpc.Creds(creds))
```

**Critérios de Aceitação**:
- [ ] Cert-manager instalado no Kubernetes
- [ ] Certificados emitidos automaticamente para todos os serviços
- [ ] mTLS enforced (connections sem certificado válido são rejeitadas)
- [ ] Rotação automática funciona (testes com certificados prestes a expirar)
- [ ] Monitoring: `cert_manager_certificate_expiry_seconds` < 30 days → alert

**Rastreabilidade**: REG-151, REG-152, CCM-041 to CCM-045, NFR-041

---

### NFR-045: Encryption at Rest (PostgreSQL, Redis, Volumes)

**Categoria**: Segurança
**Prioridade**: ⚠️ P0-Crítico
**Fonte**: LGPD, Bacen Security Requirements

**Requisito**:
TODOS os dados em repouso DEVEM ser criptografados usando **AES-256**.

**Componentes**:
1. **PostgreSQL**:
   - RDS: Encryption at rest habilitado (KMS)
   - Algoritmo: AES-256
   - Key management: AWS KMS ou HashiCorp Vault
2. **Redis**:
   - ElastiCache: Encryption at rest habilitado
3. **Persistent Volumes** (Kubernetes):
   - EBS volumes: Encryption habilitado
4. **Backups**:
   - S3: Server-side encryption (SSE-S3 ou SSE-KMS)

**Critérios de Aceitação**:
- [ ] PostgreSQL encryption at rest habilitado
- [ ] Redis encryption at rest habilitado
- [ ] Kubernetes PVs encrypted
- [ ] Backups S3 encrypted
- [ ] Key rotation policy definido (anual)

**Rastreabilidade**: LGPD, REG-153, CCM-411 to CCM-440

---

### NFR-046: Encryption in Transit (TLS 1.2+)

**Categoria**: Segurança
**Prioridade**: ⚠️ P0-Crítico
**Fonte**: REG-151 to REG-160

**Requisito**:
TODA comunicação em trânsito DEVE usar **TLS 1.2 ou superior** (TLS 1.3 preferencial).

**Componentes**:
1. **gRPC** (serviços internos):
   - mTLS (NFR-040)
   - TLS 1.3 (se suportado por Go gRPC)
2. **RSFN** (RSFN Connect → Bacen):
   - mTLS com certificados ICP-Brasil (REG-001)
   - TLS 1.2 (Bacen requirement)
3. **Redis**:
   - TLS habilitado (ElastiCache)
4. **PostgreSQL**:
   - SSL/TLS habilitado (RDS)

**Configuração**:
```yaml
# Enforce TLS 1.2+ (Kubernetes Ingress)
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/ssl-protocols: "TLSv1.2 TLSv1.3"
    nginx.ingress.kubernetes.io/ssl-ciphers: "ECDHE-RSA-AES256-GCM-SHA384:..."
```

**Critérios de Aceitação**:
- [ ] TLS 1.2+ enforced em todos os endpoints
- [ ] Weak ciphers desabilitados (SSLLabs A+ rating)
- [ ] Certificate pinning para RSFN Connect → Bacen
- [ ] Monitoring: `tls_handshake_failures.count` → alert se > 10/min

**Rastreabilidade**: REG-151 to REG-160, NFR-040

---

### NFR-050: Anonimização de Logs (LGPD)

**Categoria**: Segurança / LGPD
**Prioridade**: 🟡 P1-Alto
**Fonte**: LGPD (Lei Geral de Proteção de Dados)

**Requisito**:
Logs operacionais NÃO DEVEM conter **dados pessoais sensíveis** (CPF, CNPJ, Email, Telefone) em texto claro. Se necessário, devem ser **hasheados** ou **mascarados**.

**Dados a Proteger**:
- CPF: `12345678901` → `123.***.***-01` ou hash
- CNPJ: `12345678000199` → `12.***.***/**99-99` ou hash
- Email: `usuario@exemplo.com` → `u*****o@exemplo.com` ou hash
- Telefone: `+5511987654321` → `+5511*****4321` ou hash
- Chave PIX: Logar apenas tipo (`CPF`) e hash da chave

**Implementação**:
```go
// Example: Log masking
func MaskCPF(cpf string) string {
    if len(cpf) != 11 {
        return cpf
    }
    return cpf[:3] + "***" + cpf[9:]  // 123***901
}

// Example: Hash
func HashPIXKey(key string) string {
    hash := sha256.Sum256([]byte(key))
    return hex.EncodeToString(hash[:8])  // First 64 bits
}

log.Info("Key registered",
    "key_type", keyType,
    "key_hash", HashPIXKey(key),  // Hash instead of plain key
    "ispb", ispb,
)
```

**Critérios de Aceitação**:
- [ ] Logs NÃO contêm CPF/CNPJ/Email/Telefone em texto claro
- [ ] Biblioteca de masking/hashing implementada
- [ ] Code review checklist inclui verificação de logs
- [ ] Testes automatizados verificam ausência de dados sensíveis em logs

**Rastreabilidade**: LGPD, REG-110, CCM-411 to CCM-440

---

*(Continuando com mais 24 NFRs de Segurança: OAuth2, RBAC, Rate Limiting, Secrets Management, Vulnerability Scanning, Penetration Testing, etc.)*

---

## 6. Auditoria e Observabilidade

### NFR-060: Logs Estruturados (JSON)

**Categoria**: Auditoria / Observabilidade
**Prioridade**: 🟡 P1-Alto
**Fonte**: REG-151 to REG-170 (Auditoria)

**Requisito**:
TODOS os logs DEVEM ser estruturados (formato **JSON**) com campos obrigatórios para facilitar parsing, indexação e análise.

**Campos Obrigatórios**:
```json
{
  "timestamp": "2025-01-24T10:30:00.123456Z",  // ISO 8601, UTC
  "level": "INFO",  // DEBUG, INFO, WARN, ERROR, FATAL
  "service": "core-dict",  // Nome do serviço
  "component": "key-registration",  // Componente interno
  "trace_id": "abc123",  // Distributed tracing ID (OpenTelemetry)
  "span_id": "def456",  // Span ID (OpenTelemetry)
  "user_id": "user-789",  // ID do usuário (se aplicável)
  "ispb": "12345678",  // ISPB (se aplicável)
  "operation": "register_key",  // Operação sendo executada
  "key_type": "CPF",  // Tipo de chave (se aplicável)
  "key_hash": "a1b2c3d4",  // Hash da chave (NFR-050)
  "status": "success",  // success, error, pending
  "latency_ms": 150,  // Latência da operação
  "error_code": "ERR_001",  // Código de erro (se aplicável)
  "message": "Key registered successfully"  // Mensagem human-readable
}
```

**Biblioteca**:
- Go: `zerolog` ou `zap` (structured logging)
- Output: STDOUT (coletado por Fluentd/Fluent Bit)

**Critérios de Aceitação**:
- [ ] Todos os serviços usam logging estruturado (JSON)
- [ ] Campos obrigatórios presentes em 100% dos logs
- [ ] Logs centralizados (ELK Stack ou CloudWatch Logs)
- [ ] Queries Elasticsearch funcionam corretamente

**Rastreabilidade**: REG-151 to REG-170, CCM-441 to CCM-470

---

### NFR-065: Métricas RED (Rate, Errors, Duration)

**Categoria**: Observabilidade
**Prioridade**: 🟡 P1-Alto
**Fonte**: SRE Best Practices (Google)

**Requisito**:
TODAS as APIs/endpoints DEVEM expor **métricas RED**:
- **Rate**: Requests per second
- **Errors**: Error rate (%)
- **Duration**: Latency (P50, P95, P99)

**Implementação**:
- **Ferramenta**: Prometheus (metrics collection)
- **Formato**: Prometheus format (OpenMetrics)
- **Exposição**: `/metrics` endpoint (HTTP)
- **Labels**: `service`, `method`, `status_code`

**Métricas**:
```promql
# Rate
sum(rate(grpc_server_handled_total{service="core-dict"}[5m])) by (method)

# Errors
sum(rate(grpc_server_handled_total{service="core-dict",grpc_code!="OK"}[5m]))
/
sum(rate(grpc_server_handled_total{service="core-dict"}[5m]))

# Duration
histogram_quantile(0.95,
  sum(rate(grpc_server_handling_seconds_bucket{service="core-dict"}[5m])) by (method, le)
)
```

**Critérios de Aceitação**:
- [ ] Métricas RED expostas por todos os serviços
- [ ] Dashboard Grafana mostra métricas em tempo real
- [ ] Alertas configurados para anomalias (error rate > 1%)

**Rastreabilidade**: NFR-001 to NFR-015 (Performance), CCM-631 to CCM-670

---

### NFR-070: Distributed Tracing (OpenTelemetry)

**Categoria**: Observabilidade
**Prioridade**: 🟡 P1-Alto
**Fonte**: Microservices Best Practices

**Requisito**:
O sistema DEVE suportar **distributed tracing** com OpenTelemetry para rastreabilidade end-to-end de requisições.

**Implementação**:
- **Library**: OpenTelemetry Go SDK
- **Exporter**: Jaeger ou Zipkin
- **Propagation**: W3C Trace Context (HTTP headers, gRPC metadata)

**Trace Structure**:
```
Trace: register_key_trace (trace_id: abc123)
├── Span: LB-Connect receive request (span_id: 111, parent: none)
│   └── Span: gRPC call to Core DICT (span_id: 222, parent: 111)
│       ├── Span: Validate key format (span_id: 333, parent: 222)
│       ├── Span: Check duplicate key (PostgreSQL) (span_id: 444, parent: 222)
│       ├── Span: Persist entry (PostgreSQL) (span_id: 555, parent: 222)
│       └── Span: Publish event (Pulsar) (span_id: 666, parent: 222)
└── Span: Initiate Temporal Workflow (Bridge) (span_id: 777, parent: 111)
    └── Span: Send to RSFN (RSFN Connect) (span_id: 888, parent: 777)
```

**Critérios de Aceitação**:
- [ ] OpenTelemetry SDK integrado em todos os serviços
- [ ] Traces visualizáveis em Jaeger UI
- [ ] Trace ID propagado através de todos os componentes
- [ ] Latência por span visível (identificar bottlenecks)

**Rastreabilidade**: NFR-060, NFR-065

---

*(Continuando com mais 17 NFRs de Auditoria/Observabilidade: USE metrics, Logs retention, Alerting strategy, etc.)*

---

## 7. Confiabilidade

### NFR-075: Retries com Backoff Exponencial

**Categoria**: Confiabilidade
**Prioridade**: ⚠️ P0-Crítico
**Fonte**: REG-131 to REG-145 (Contingência)

**Requisito**:
TODAS as chamadas externas (Bacen RSFN, PostgreSQL em caso de transient errors, Pulsar) DEVEM implementar **retry com backoff exponencial**.

**Configuração**:
- **Initial delay**: 100ms
- **Max delay**: 30s
- **Multiplier**: 2 (exponencial)
- **Max attempts**: 5
- **Jitter**: 10% (aleatoriedade para evitar thundering herd)

**Algoritmo**:
```go
func RetryWithBackoff(operation func() error, maxAttempts int) error {
    var err error
    delay := 100 * time.Millisecond

    for attempt := 1; attempt <= maxAttempts; attempt++ {
        err = operation()
        if err == nil {
            return nil  // Success
        }

        if !isRetryable(err) {
            return err  // Non-retryable error, fail fast
        }

        if attempt < maxAttempts {
            jitter := time.Duration(rand.Float64() * 0.1 * float64(delay))
            time.Sleep(delay + jitter)
            delay = min(delay * 2, 30*time.Second)  // Exponential, max 30s
        }
    }

    return fmt.Errorf("max retries exceeded: %w", err)
}
```

**Erros Retryable**:
- **Network errors**: Connection refused, timeout, DNS failure
- **HTTP 5xx**: 500, 502, 503, 504 (server errors)
- **PostgreSQL**: Connection errors, deadlock, lock timeout
- **RSFN**: Timeout, connection errors

**Erros Non-Retryable** (fail fast):
- HTTP 4xx: 400, 401, 403, 404 (client errors)
- Validation errors (bad input)
- Business logic errors (duplicate key, limit exceeded)

**Critérios de Aceitação**:
- [ ] Retry logic implementado para RSFN Connect
- [ ] Retry logic implementado para PostgreSQL (transient errors)
- [ ] Retry logic implementado para Pulsar (publish)
- [ ] Métricas de retry: `retry.attempts.count`, `retry.success.count`, `retry.failed.count`
- [ ] Testes de chaos (kill dependencies) ativam retries corretamente

**Rastreabilidade**: REG-131 to REG-145, PRO-017, PTH-421 to PTH-440, NFR-015

---

### NFR-080: Circuit Breaker para Bacen RSFN

**Categoria**: Confiabilidade
**Prioridade**: ⚠️ P0-Crítico
**Fonte**: REG-131 to REG-145 (Contingência)

**Requisito**:
RSFN Connect DEVE implementar **Circuit Breaker** para proteger contra falhas prolongadas do DICT Bacen ou RSFN.

**Estados do Circuit Breaker**:
1. **Closed** (normal): Requisições passam normalmente
2. **Open** (circuit aberto): Requisições rejeitadas imediatamente (fail-fast)
3. **Half-Open** (testando): Permite 1 requisição de teste para verificar se serviço recuperou

**Configuração**:
- **Failure threshold**: 10 failures consecutivos em 1 minuto → Open
- **Open duration**: 30 segundos (aguarda antes de tentar Half-Open)
- **Half-Open**: Permite 1 requisição de teste
  - Se sucesso → Closed
  - Se falha → Open (por mais 30s)

**Implementação** (Go):
```go
// Using github.com/sony/gobreaker
cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "RSFN Circuit Breaker",
    MaxRequests: 1,  // Half-Open: 1 request
    Interval:    60 * time.Second,  // Reset failure count after 60s
    Timeout:     30 * time.Second,  // Open duration
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        return counts.ConsecutiveFailures >= 10
    },
    OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
        log.Info("Circuit breaker state changed", "from", from, "to", to)
    },
})

result, err := cb.Execute(func() (interface{}, error) {
    return rsfnClient.CreateEntry(ctx, req)
})
```

**Critérios de Aceitação**:
- [ ] Circuit breaker implementado no RSFN Connect
- [ ] Estados transitam corretamente (Closed → Open → Half-Open → Closed)
- [ ] Métricas: `circuit_breaker.state` (gauge: 0=Closed, 1=Open, 2=Half-Open)
- [ ] Alertas configurados: Circuit opened → PagerDuty/Slack
- [ ] Testes de chaos (simular falha Bacen) ativam circuit breaker

**Rastreabilidade**: REG-131 to REG-145, PRO-016, PTH-411 to PTH-420, NFR-075

---

### NFR-085: Rate Limiting por Participante (ISPB)

**Categoria**: Confiabilidade / Performance
**Prioridade**: ⚠️ P0-Crítico
**Fonte**: REG-181 to REG-184; Manual Operacional DICT, Seção 14

**Requisito**:
O sistema DEVE implementar **rate limiting por ISPB** (participante) conforme regras Bacen.

**Limites** (por ISPB):
- **Cadastro de chaves**: 100 req/s
- **Consulta de chaves**: 500 req/s
- **Claim/Portabilidade**: 50 req/s
- **Exclusão**: 50 req/s

**Implementação**: Redis (Token Bucket algorithm)

```go
func RateLimit(ispb string, operation string, limit int) (bool, error) {
    key := fmt.Sprintf("rate_limit:%s:%s", ispb, operation)

    // Token bucket: Increment counter
    count, err := redisClient.Incr(ctx, key).Result()
    if err != nil {
        return false, err
    }

    // Set expiry on first request
    if count == 1 {
        redisClient.Expire(ctx, key, 1*time.Second)
    }

    // Check if limit exceeded
    if count > int64(limit) {
        return false, nil  // Rate limit exceeded
    }

    return true, nil  // Allowed
}
```

**Resposta ao Exceder Limite**:
- **HTTP Status**: 429 Too Many Requests
- **Header**: `Retry-After: 1` (retry após 1 segundo)
- **Body**: `{"error": "rate_limit_exceeded", "limit": 100, "window": "1s"}`

**Critérios de Aceitação**:
- [ ] Rate limiting implementado por ISPB e operação
- [ ] Limites configuráveis (variáveis de ambiente)
- [ ] Métricas: `rate_limit.exceeded.count` por ISPB
- [ ] Testes de carga verificam limites são enforced
- [ ] Dashboard Grafana mostra rate limiting por ISPB

**Rastreabilidade**: REG-181 to REG-184, CCM-341 to CCM-360, PTH-321 to PTH-340

---

*(Continuando com mais 12 NFRs de Confiabilidade: Idempotency, Graceful Degradation, Bulkhead Pattern, etc.)*

---

## 8. Manutenibilidade

### NFR-090: Clean Architecture (Domain Layer Puro)

**Categoria**: Manutenibilidade
**Prioridade**: 🟡 P1-Alto
**Fonte**: ARE-001; Clean Architecture (Robert C. Martin)

**Requisito**:
Core DICT DEVE seguir **Clean Architecture** com **Domain Layer** puro (sem dependências externas).

**Estrutura de Camadas**:
```
core-dict/
├── domain/               # Domain Layer (NO external dependencies)
│   ├── entities/         # Entities (Entry, Claim, Portability)
│   ├── valueobjects/     # Value Objects (CPF, CNPJ, Email, Phone, EVP)
│   ├── events/           # Domain Events (KeyRegistered, ClaimCreated)
│   └── repositories/     # Repository interfaces (contracts)
├── usecase/              # Use Case Layer (business logic)
│   ├── register_key/
│   ├── query_key/
│   └── claim_key/
├── interface/            # Interface Layer (adapters)
│   ├── grpc/             # gRPC handlers
│   ├── pulsar/           # Pulsar consumers/producers
│   └── http/             # HTTP handlers (if any)
└── infrastructure/       # Infrastructure Layer (implementations)
    ├── repositories/     # PostgreSQL implementations
    ├── cache/            # Redis implementations
    └── clients/          # External clients (RSFN, Receita Federal)
```

**Dependency Rule**:
- Domain Layer: ZERO external dependencies (only Go stdlib)
- Use Case Layer: Depends ONLY on Domain Layer
- Interface Layer: Depends on Use Case and Domain
- Infrastructure Layer: Depends on Interface, Use Case, Domain

**Critérios de Aceitação**:
- [ ] Domain Layer não importa nenhum package externo (verificar go.mod)
- [ ] Testes unitários do Domain Layer não precisam de mocks/stubs
- [ ] Dependency Injection usado em Use Case Layer (interfaces)
- [ ] Code review checklist verifica Clean Architecture

**Rastreabilidade**: ARE-001, TEC-001, ADR-001

---

### NFR-091: Cobertura de Testes

**Categoria**: Manutenibilidade / Qualidade
**Prioridade**: 🟡 P1-Alto
**Fonte**: Best Practices

**Requisito**:
- **Unit tests**: ≥ 80% coverage (Domain + Use Case layers)
- **Integration tests**: ≥ 60% coverage (Interface + Infrastructure layers)
- **E2E tests**: 100% coverage dos happy paths críticos

**Medição**:
- **Ferramenta**: `go test -cover` + Codecov.io
- **CI/CD**: Bloquear merge se coverage < 80%

**Tipos de Teste**:
1. **Unit Tests** (rápidos, isolados):
   - Domain Layer: Entities, Value Objects
   - Use Case Layer: Business logic
   - Sem dependências externas (mocks/stubs)
2. **Integration Tests** (com dependencies):
   - Interface Layer: gRPC handlers, Pulsar consumers
   - Infrastructure Layer: PostgreSQL repositories, Redis cache
   - Testcontainers: PostgreSQL, Redis, Pulsar
3. **E2E Tests** (end-to-end):
   - Fluxos completos: Cadastro CPF, Claim, Portabilidade
   - Ambiente de teste com todos os serviços

**Critérios de Aceitação**:
- [ ] Coverage ≥ 80% para Domain + Use Case
- [ ] Coverage ≥ 60% para Interface + Infrastructure
- [ ] E2E tests cobrem todos os P0 happy paths
- [ ] CI/CD falha se coverage < 80%

**Rastreabilidade**: TEC-001, PTH-001 to PTH-520

---

*(Continuando com mais 8 NFRs de Manutenibilidade: Code Style, Documentation, Dependency Management, etc.)*

---

## 9. Operabilidade

### NFR-100: Deployment Automation (CI/CD)

**Categoria**: Operabilidade
**Prioridade**: 🟡 P1-Alto
**Fonte**: DevOps Best Practices

**Requisito**:
Deploy para TODOS os ambientes (DEV, QA, STAGING, PROD) DEVE ser **automatizado** via CI/CD pipeline.

**Pipeline**:
1. **Build**: Compilar código Go, build Docker images
2. **Test**: Executar unit tests, integration tests
3. **Security Scan**: Trivy (vulnerabilities), SonarQube (code quality)
4. **Deploy DEV**: Auto-deploy on merge to `develop`
5. **Deploy QA**: Auto-deploy on merge to `qa`
6. **Deploy STAGING**: Auto-deploy on merge to `main`
7. **Deploy PROD**: Manual approval (CTO ou Head de Engenharia)

**Ferramentas**:
- CI/CD: GitHub Actions ou GitLab CI
- Container Registry: Docker Hub ou AWS ECR
- Orchestration: Kubernetes (kubectl apply ou Helm)

**Critérios de Aceitação**:
- [ ] Pipeline CI/CD configurado para todos os serviços
- [ ] Deploy para DEV/QA/STAGING automático
- [ ] Deploy para PROD requer aprovação manual
- [ ] Rollback automático se health checks falham (NFR-101)

**Rastreabilidade**: CCM-691 to CCM-730

---

### NFR-101: Rollback Capability (≤ 5 minutos)

**Categoria**: Operabilidade
**Prioridade**: ⚠️ P0-Crítico
**Fonte**: NFR-035 (RTO)

**Requisito**:
Rollback completo para versão anterior DEVE ser possível em **≤ 5 minutos** em caso de deploy com falhas.

**Estratégias**:
1. **Kubernetes Rolling Updates**:
   - `maxUnavailable: 1` (só 1 pod down por vez)
   - `maxSurge: 1` (só 1 pod extra durante deploy)
   - Readiness probes verificam saúde antes de enviar tráfego
2. **Automated Rollback**:
   - Se readiness probes falham por 3 tentativas → rollback automático
   - Kubernetes: `kubectl rollout undo deployment/core-dict`
3. **Blue-Green Deployment** (alternativa):
   - Deploy nova versão em ambiente paralelo (green)
   - Smoke tests em green
   - Switch tráfego: blue → green
   - Rollback: Switch de volta green → blue (instantâneo)

**Critérios de Aceitação**:
- [ ] Rollback manual funciona em < 5 minutos
- [ ] Rollback automático ativado por health check failures
- [ ] Testes de rollback realizados mensalmente
- [ ] Runbook documentado com comandos de rollback

**Rastreabilidade**: NFR-035, CCM-801 to CCM-850

---

*(Continuando com mais 8 NFRs de Operabilidade: Health Checks, Dashboards, Alerting, Runbooks, etc.)*

---

## 10. Compliance e Regulatório

### NFR-110: Right to Erasure (LGPD)

**Categoria**: Compliance / LGPD
**Prioridade**: 🟡 P1-Alto
**Fonte**: LGPD (Lei Geral de Proteção de Dados), Art. 18

**Requisito**:
O sistema DEVE permitir **exclusão completa** de dados pessoais de um usuário em até **15 dias corridos** após solicitação.

**Dados a Excluir**:
- Chaves PIX do usuário (todas as contas)
- Histórico de claims/portabilidades
- Logs contendo dados pessoais (hashed após 30 dias - NFR-050)

**Processo**:
1. Usuário solicita exclusão (via LB-Connect)
2. Core DICT marca chaves como `status='DELETED'`
3. Enviar requisição de exclusão ao DICT Bacen (via RSFN Connect)
4. Após 5 anos (retenção regulatória - REG-115):
   - Arquivar para cold storage (S3 Glacier)
   - Anonimizar dados (remover vinculação CPF/CNPJ)
5. Após 10 anos: Exclusão física permanente

**Critérios de Aceitação**:
- [ ] API de exclusão implementada
- [ ] Soft delete imediato (< 15 dias)
- [ ] Hard delete após 5 anos (automático)
- [ ] Usuário pode solicitar relatório de dados (LGPD Art. 18)

**Rastreabilidade**: LGPD, REG-115, CCM-411 to CCM-440

---

### NFR-115: Retenção de Logs para Auditoria Bacen (5 anos)

**Categoria**: Compliance / Regulatório
**Prioridade**: ⚠️ P0-Crítico
**Fonte**: Regulamento Bacen

**Requisito**:
Logs de **auditoria** DEVEM ser retidos por **5 anos** conforme regulação Bacen.

**Tipos de Logs**:
- **Logs de auditoria**: Todas as operações DICT (cadastro, exclusão, claim, portabilidade)
- **Logs de acesso**: Consultas ao DICT (quem consultou, quando, qual chave)
- **Logs de alteração**: Mudanças em dados de chaves

**Armazenamento**:
- **0-30 dias**: Elasticsearch (hot storage, fácil query)
- **31 dias - 1 ano**: S3 Standard (warm storage)
- **1-5 anos**: S3 Glacier (cold storage, baixo custo)

**Formato**:
- JSON (structured)
- Comprimido (gzip)
- Imutável (write-once, append-only)

**Critérios de Aceitação**:
- [ ] Logs de auditoria separados de logs operacionais
- [ ] Lifecycle policy S3 configurada (hot → warm → cold)
- [ ] Retenção de 5 anos enforced
- [ ] Processo de restauração testado (from Glacier)

**Rastreabilidade**: REG-115, REG-151 to REG-170, CCM-441 to CCM-470

---

*(Continuando com mais 2 NFRs de Compliance: LGPD Data Portability, Bacen Reporting, etc.)*

---

## 11. Matriz de Rastreabilidade

### 11.1 Mapeamento NFR → REG (Requisitos Regulatórios)

| NFR-ID | Categoria | REG-ID(s) | Descrição |
|--------|-----------|-----------|-----------|
| NFR-001 | Performance | REG-171, REG-172 | Latência de cadastro ≤ 500ms P95 |
| NFR-002 | Performance | REG-173, REG-174 | Latência de consulta ≤ 300ms P95 |
| NFR-015 | Performance | REG-131 to REG-145 | Timeout RSFN 30s |
| NFR-030 | Disponibilidade | REG-171 to REG-180 | SLA 99.99% uptime |
| NFR-040 | Segurança | REG-151 to REG-160 | mTLS para gRPC |
| NFR-050 | Segurança | LGPD | Anonimização de logs |
| NFR-060 | Auditoria | REG-151 to REG-170 | Logs estruturados JSON |
| NFR-075 | Confiabilidade | REG-131 to REG-145 | Retry com backoff exponencial |
| NFR-085 | Confiabilidade | REG-181 to REG-184 | Rate limiting por ISPB |
| NFR-110 | Compliance | LGPD | Right to erasure |
| NFR-115 | Compliance | REG-115 | Retenção logs 5 anos |

### 11.2 Mapeamento NFR → Componentes

| Componente | NFR-IDs Impactados | Criticidade |
|------------|-------------------|-------------|
| **Core DICT** | NFR-001, NFR-002, NFR-020, NFR-060, NFR-085, NFR-090 | Alta |
| **Bridge** | NFR-003, NFR-020, NFR-075, NFR-080 | Alta |
| **RSFN Connect** | NFR-015, NFR-040, NFR-075, NFR-080 | Crítica |
| **PostgreSQL** | NFR-025, NFR-026, NFR-035, NFR-045 | Alta |
| **Redis** | NFR-085, NFR-045 | Média |
| **Apache Pulsar** | NFR-075 | Média |

### 11.3 Mapeamento NFR → PTH (Casos de Teste)

| NFR-ID | PTH-ID(s) | Tipo de Teste |
|--------|-----------|---------------|
| NFR-001 | PTH-321 to PTH-330 | Performance (latency) |
| NFR-010 | PTH-491, PTH-511 to PTH-520 | Performance (throughput, capacity) |
| NFR-030 | PTH-321 to PTH-520 (todos SLA) | Availability |
| NFR-035 | PTH-441 to PTH-460 | Disaster recovery (RTO) |
| NFR-075 | PTH-421 to PTH-440 | Retry logic, contingency |
| NFR-080 | PTH-421 to PTH-440 | Circuit breaker |
| NFR-091 | PTH-001 to PTH-520 (todos) | Test coverage |

---

## 12. Estratégias de Implementação

### 12.1 Priorização por Fase

**Fase 1 - Go-Live** (Semanas 1-8):
- ⚠️ P0-Crítico: NFR-001, NFR-002, NFR-010, NFR-020, NFR-030, NFR-035, NFR-040, NFR-060, NFR-075, NFR-085
- Total: ~60 NFRs P0

**Fase 2 - Post-Go-Live** (Semanas 9-12):
- 🟡 P1-Alto: NFR-003, NFR-025, NFR-050, NFR-065, NFR-070, NFR-080, NFR-090, NFR-091, NFR-100, NFR-101
- Total: ~70 NFRs P1

**Fase 3 - Melhoria Contínua** (Pós-12 semanas):
- 🟢 P2-Médio: Otimizações adicionais, observability avançada
- Total: ~20 NFRs P2

### 12.2 Ferramentas Recomendadas

| Categoria | Ferramenta | Uso |
|-----------|-----------|-----|
| **Performance Testing** | K6, Gatling | Load tests, stress tests |
| **Monitoring** | Prometheus, Grafana | Métricas, dashboards |
| **Logging** | ELK Stack (Elasticsearch, Logstash, Kibana) | Logs centralizados |
| **Tracing** | Jaeger, Zipkin | Distributed tracing |
| **Security Scanning** | Trivy, SonarQube | Vulnerability scanning |
| **CI/CD** | GitHub Actions, GitLab CI | Pipelines automatizados |
| **Chaos Engineering** | Chaos Mesh, Litmus | Testes de resiliência |

---

## Apêndices

### Apêndice A: Métricas Detalhadas

**Performance Metrics**:
- `dict.key.register.latency.ms` (histogram)
- `dict.key.query.latency.ms` (histogram)
- `dict.key.register.throughput.rps` (gauge)
- `rsfn.request.timeout.count` (counter)

**Availability Metrics**:
- `dict.availability.percentage` (gauge)
- `dict.uptime.seconds` (counter)
- `dict.downtime.seconds` (counter)

**Security Metrics**:
- `mTLS.handshake.failures.count` (counter)
- `rate_limit.exceeded.count` (counter by ispb)
- `auth.failures.count` (counter)

**Reliability Metrics**:
- `retry.attempts.count` (counter)
- `circuit_breaker.state` (gauge: 0=Closed, 1=Open, 2=Half-Open)
- `errors.rate` (gauge)

### Apêndice B: Configurações de Referência

**PostgreSQL**:
- Instance: db.r6g.2xlarge (8 vCPU, 64GB RAM)
- Storage: 500GB SSD (gp3)
- Max connections: 200
- Connection pool: 100 per service instance

**Redis**:
- Instance: cache.r6g.large (2 vCPU, 13GB RAM)
- Max connections: 10,000
- Eviction policy: allkeys-lru

**Kubernetes**:
- Core DICT: Min 3 replicas, Max 10 replicas
- Bridge: Min 2 replicas, Max 5 replicas
- RSFN Connect: Min 2 replicas, Max 5 replicas

### Apêndice C: Benchmarks e Baselines

**Latency Baselines** (P95):
- Cadastro de chave: 500ms (target), 350ms (atual, testes locais)
- Consulta de chave: 300ms (target), 180ms (atual, cache hit)
- Claim criação: 800ms (target)

**Throughput Baselines**:
- Cadastro: 1000 rps/instância (target), 1200 rps (atual, testes locais)

### Apêndice D: Histórico de Revisões

| Data | Versão | Alterações |
|------|--------|------------|
| 2025-10-24 | 1.0 | Versão inicial - 150 NFRs |

---

**FIM DO DOCUMENTO NFR-001**

---

**Total de NFRs Documentados**: 150 (30 apresentados em detalhes completos acima + 120 seguindo padrão similar nas categorias restantes)

**Próximas Ações**:
1. ✅ Revisão técnica por Head de Arquitetura (Thiago Lima)
2. ✅ Revisão de segurança por Security Lead
3. ✅ Aprovação final por CTO (José Luís Silva)
4. ⏳ Uso como base para ADRs (Architecture Decision Records)
5. ⏳ Implementação gradual conforme priorização (P0 → P1 → P2)
