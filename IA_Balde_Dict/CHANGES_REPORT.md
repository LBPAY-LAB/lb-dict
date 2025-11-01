# Relatório de Mudanças - DICT Rate Limit Monitoring
**Data**: 2025-11-01
**Base**: Decisões consolidadas em DUVIDAS.md
**Impacto**: Atualização do CLAUDE.md

---

## 📊 Resumo Executivo

| Categoria | Quantidade |
|-----------|------------|
| ✅ Decisões Confirmadas | 11 |
| 🔴 Mudanças Críticas | 3 |
| ➕ Novos Requisitos | 5 |
| ➖ Funcionalidades Removidas | 3 |
| ⚠️ Pendências | 2 |

---

## 🔴 MUDANÇAS CRÍTICAS

### 1. Threshold WARNING Corrigido (CRÍTICO)
**Impacto**: Alto - Afeta lógica de alertas

**ANTES (CLAUDE.md original)**:
```go
WARNING_THRESHOLD := Capacity * 0.25  // 25% de fichas restantes (75% utilizado)
CRITICAL_THRESHOLD := Capacity * 0.10 // 10% de fichas restantes (90% utilizado)
```

**DEPOIS (baseado em DUVIDAS.md Q3)**:
```go
WARNING_THRESHOLD := Capacity * 0.20  // 20% de fichas restantes (80% utilizado)
CRITICAL_THRESHOLD := Capacity * 0.10 // 10% de fichas restantes (90% utilizado)
```

**Razão**: Decisão do stakeholder - WARNING mais conservador para dar mais tempo de reação.

**Arquivos Impactados**:
- `internal/domain/ratelimit/threshold_calculator.go`
- `internal/temporal/activities/ratelimit/check_threshold_activity.go`
- Testes unitários relacionados

---

### 2. Remoção do Cache Redis (CRÍTICO)
**Impacto**: Alto - Simplifica arquitetura, aumenta latência

**ANTES (CLAUDE.md original)**:
```
┌──────────────────┐
│  Cache Redis     │  TTL: 60s
│  (opcional)      │  Reduz chamadas ao Bridge/DICT
└──────────────────┘
```

**DEPOIS (baseado em DUVIDAS.md Q4)**:
```
❌ REMOVIDO - Sempre consultar DICT via Bridge (dados sempre frescos)
```

**Razão**:
- Decisão do stakeholder: Dados de rate limit devem ser sempre atualizados
- Intervalos de 5 minutos já são suficientes para não sobrecarregar DICT
- Simplicidade arquitetural (menos componentes)

**Arquivos Impactados**:
- `apps/orchestration-worker/infrastructure/cache/` ❌ NÃO CRIAR
- `internal/application/usecases/ratelimit/application.go` - Remover lógica de cache
- `setup/setup.go` - Remover inicialização de Redis para rate limit

---

### 3. Observabilidade Reduzida no Escopo Inicial (CRÍTICO)
**Impacto**: Médio - Afeta entrega inicial, mas não bloqueia

**ANTES (CLAUDE.md original)**:
```yaml
Observabilidade:
  - Prometheus metrics (MANTIDO)
  - Grafana dashboards predefinidos ❌ REMOVIDO
  - PagerDuty integration ❌ REMOVIDO
  - Slack alerts ❌ REMOVIDO
```

**DEPOIS (baseado em DUVIDAS.md Q11)**:
```yaml
Observabilidade (Escopo Inicial):
  - Prometheus metrics ✅ MANTIDO
  - Prometheus AlertManager (local) ✅ MANTIDO
  - Logs estruturados (slog) ✅ MANTIDO

Pós-lançamento (Time de Infra):
  - Grafana dashboards (time de infra cria depois)
  - PagerDuty (se necessário, configurado por SRE)
  - Slack webhooks (se necessário, configurado por SRE)
```

**Razão**:
- Métricas Prometheus são suficientes para monitoramento inicial
- Time de infraestrutura já tem expertise em Grafana
- Evita over-engineering na primeira entrega

**Arquivos Impactados**:
- `SPECS-OBSERVABILITY.md` - Remover seções de Grafana/PagerDuty detalhadas
- `internal/infrastructure/observability/grafana/` ❌ NÃO CRIAR
- `internal/infrastructure/observability/pagerduty/` ❌ NÃO CRIAR
- `SPECS-DEPLOYMENT.md` - Remover Helm values para Grafana

---

## ➕ NOVOS REQUISITOS

### 4. Métricas de Erros 404 (Anti-Scan)
**Impacto**: Médio - Nova funcionalidade de monitoramento

**Requisito (DUVIDAS.md Q6)**:
```go
// Nova métrica para rastrear rate de erros 404 por política
dict_rate_limit_404_rate{policy="ENTRIES_READ"} 0.15  // 15% de erros 404

// Alerta se 404_rate > 20% (possível ataque de scanning)
if (rate(dict_rate_limit_404_errors_total[5m]) / rate(dict_rate_limit_requests_total[5m])) > 0.20 {
    alert("High 404 rate - possible anti-scan penalty")
}
```

**Arquivos Impactados**:
- `internal/infrastructure/observability/metrics/ratelimit_metrics.go` - Adicionar counter de 404s
- `internal/temporal/activities/ratelimit/check_policy_activity.go` - Registrar 404s do DICT
- `SPECS-OBSERVABILITY.md` - Documentar nova métrica
- AlertManager rules - Adicionar alerta de 404_rate

---

### 5. Cálculo de ETA (Estimated Time to Recovery)
**Impacto**: Médio - Melhora alertas e dashboards

**Requisito (DUVIDAS.md Q8.1)**:
```go
// Calcular tempo até recuperação completa (100% de fichas)
func CalculateRecoveryETA(available, capacity, refillTokens, refillPeriodSec int) time.Duration {
    tokensNeeded := capacity - available
    if tokensNeeded <= 0 {
        return 0  // Já está cheio
    }

    periodsNeeded := (tokensNeeded + refillTokens - 1) / refillTokens  // Ceiling division
    secondsToRecovery := periodsNeeded * refillPeriodSec

    return time.Duration(secondsToRecovery) * time.Second
}

// Exemplo:
// Available: 50, Capacity: 1000, RefillTokens: 60, RefillPeriodSec: 60
// Tokens needed: 950
// Periods: ceil(950/60) = 16 períodos
// ETA: 16 * 60s = 16 minutos
```

**Arquivos Impactados**:
- `internal/domain/ratelimit/eta_calculator.go` - Nova entidade
- `internal/temporal/activities/ratelimit/check_policy_activity.go` - Calcular ETA e incluir no alerta
- `internal/infrastructure/observability/metrics/ratelimit_metrics.go` - Gauge de ETA
- Testes unitários

---

### 6. Projeção de Esgotamento
**Impacto**: Médio - Previsão proativa de problemas

**Requisito (DUVIDAS.md Q8.2)**:
```go
// Projetar quando as fichas se esgotarão baseado no rate de consumo atual
func ProjectExhaustion(currentTokens, consumptionRatePerMinute int) (time.Duration, bool) {
    if consumptionRatePerMinute <= 0 {
        return 0, false  // Sem consumo, sem esgotamento
    }

    minutesToExhaustion := currentTokens / consumptionRatePerMinute
    return time.Duration(minutesToExhaustion) * time.Minute, true
}

// Alerta preventivo se esgotamento projetado < 30 minutos
if exhaustionTime, hasConsumption := ProjectExhaustion(available, rate); hasConsumption && exhaustionTime < 30*time.Minute {
    alert("Tokens will exhaust in %v minutes at current rate", exhaustionTime.Minutes())
}
```

**Observação**: Requer tracking de consumo (diferença entre checks consecutivos).

**Arquivos Impactados**:
- `internal/domain/ratelimit/projection_calculator.go` - Nova entidade
- `internal/database/repositories/ratelimit/` - Armazenar histórico de consumo
- Tabela `dict_rate_limit_history` - Adicionar coluna `consumption_rate_per_minute`
- Testes de regressão

---

### 7. Monitoramento de Mudança de Categoria PSP
**Impacto**: Baixo - Edge case, mas importante para compliance

**Requisito (DUVIDAS.md Q5)**:
```go
// Detectar mudança de categoria do PSP (A-H) entre verificações
type RateLimitSnapshot struct {
    CheckedAt   time.Time
    Policy      string
    Category    string  // "A", "B", "C", etc. ← NOVO CAMPO
    Available   int
    Capacity    int
    // ...
}

// Ao processar resposta DICT
if previousCategory != currentCategory {
    logger.Warn("PSP category changed",
        "policy", policyName,
        "oldCategory", previousCategory,
        "newCategory", currentCategory,
        "oldCapacity", previousCapacity,
        "newCapacity", currentCapacity,
    )

    // Publicar evento Pulsar para Core-Dict
    publishEvent("ActionPSPCategoryChanged", CategoryChangeEvent{
        Policy:       policyName,
        OldCategory:  previousCategory,
        NewCategory:  currentCategory,
        ChangedAt:    time.Now(),
    })
}
```

**Arquivos Impactados**:
- Tabela `dict_rate_limit_snapshots` - Adicionar coluna `psp_category VARCHAR(1)`
- `internal/domain/ratelimit/category_monitor.go` - Nova entidade
- Schema Pulsar - Novo evento `ActionPSPCategoryChanged`
- Migration SQL

---

### 8. Uso de Timestamp do DICT para Auditoria
**Impacto**: Baixo - Melhora rastreabilidade

**Requisito (DUVIDAS.md Q9.1)**:
```go
// Usar <ResponseTime> do DICT como timestamp oficial
type BridgeRateLimitResponse struct {
    ResponseTime time.Time  // De <ResponseTime>2025-11-01T10:30:00Z</ResponseTime>
    Policies     []PolicySnapshot
}

// No activity
func (a *RateLimitActivity) CheckPolicy(ctx context.Context, policyName string) (*Snapshot, error) {
    resp, err := a.bridge.GetRateLimitStatus(ctx, policyName)
    if err != nil {
        return nil, err
    }

    snapshot := &Snapshot{
        CheckedAt: resp.ResponseTime,  // ← Timestamp do DICT, não time.Now()
        Policy:    policyName,
        Available: resp.AvailableTokens,
        // ...
    }

    return snapshot, nil
}
```

**Razão**: Auditoria rastreável - timestamp do DICT é autoridade para troubleshooting.

**Arquivos Impactados**:
- `internal/infrastructure/grpc/ratelimit/bridge_client.go` - Parsear `<ResponseTime>`
- Todos os activities que salvam snapshots
- Documentação de auditoria

---

## ➖ FUNCIONALIDADES REMOVIDAS

### Removido 1: Cache Redis para Rate Limit
- **Razão**: Ver "Mudança Crítica #2"
- **Escopo**: Toda a layer de cache Redis específica para rate limit
- **Impacto**: Simplificação arquitetural

### Removido 2: Grafana Dashboards Predefinidos
- **Razão**: Ver "Mudança Crítica #3"
- **Escopo**: Dashboards JSON, templates Helm para Grafana
- **Impacto**: Movido para pós-lançamento (time de infra)

### Removido 3: Integração PagerDuty/Slack Inicial
- **Razão**: Ver "Mudança Crítica #3"
- **Escopo**: Webhooks, API clients para PagerDuty/Slack
- **Impacto**: Prometheus AlertManager local é suficiente inicialmente

---

## ✅ CONFIRMAÇÕES (SEM MUDANÇA)

### Confirmado 1: Bridge gRPC Endpoints Existem (Q1)
- ✅ `GetRateLimitStatus()` disponível
- ✅ Mappers XML ↔ gRPC existentes
- ✅ Sem bloqueadores de integração

### Confirmado 2: Core-Dict Consumer (Q2)
- ✅ Time do Core-Dict implementa consumer de `core-events`
- ✅ Apenas publicar eventos Pulsar (schema `ActionRateLimitAlert`)

### Confirmado 3: Intervalo de Monitoramento (Q4.1)
- ✅ 5 minutos (via Temporal cron)
- ✅ Sem cache (sempre consultar DICT)

### Confirmado 4: Políticas CIDS_FILES (Q7)
- ✅ Monitorar normalmente (não precisa tratamento especial)

### Confirmado 5: Clock e Timezone (Q9)
- ✅ Forçar UTC em todos os componentes (`TZ=UTC`)
- ✅ Usar timestamp do DICT (`<ResponseTime>`) para auditoria

### Confirmado 6: Data Retention (Q10)
- ✅ 13 meses de histórico (particionamento mensal)

### Confirmado 7: Deployment Tools (Q12)
- ✅ Goose para migrations SQL
- ✅ Kubernetes manifests diretos (sem Helm)
- ✅ Prometheus AlertManager local

---

## ⚠️ PENDÊNCIAS

### ✅ RESOLVIDAS (2/2)

1. **Bridge gRPC Integration** (Q1) - ✅ **CONCLUÍDO**
   - Endpoints existem e estão prontos
   - Proto definitions validadas
   - Mappers XML ↔ gRPC disponíveis
   - Documentação completa: [BRIDGE_ENDPOINTS_RATE_LIMIT.md](./.claude/BRIDGE_ENDPOINTS_RATE_LIMIT.md)

2. **Secrets Management** (Q12.3) - ✅ **CONCLUÍDO**
   - AWS Secrets Manager definido
   - Estrutura de secrets documentada
   - Integração com mTLS especificada

### ⚠️ PENDENTE (1/2)

### Pendência 1: Categoria PSP do LBPay (Q5.1)
**Status**: ⚠️ Bloqueador para testes realistas

**Ação Necessária**:
```bash
# Executar query real ao DICT para descobrir categoria
curl -X POST https://dict.pi.rsfn.net.br/api/v2/rate-limit/status \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"policy": "ENTRIES_READ"}'

# Resposta esperada:
<Category>A</Category>  # ou B, C, D, E, F, G, H
```

**Impacto**:
- Sem saber a categoria, não sabemos os limites reais (Capacity, RefillTokens)
- Testes usarão valores mockados (categoria "A" por padrão)

**Responsável**: Time de integrações + BACEN (consulta real)

---

### Pendência 2: Secrets Management (Q12.3)
**Status**: ✅ **RESOLVIDO** - AWS Secrets Manager

**Decisão**: AWS Secrets Manager (infraestrutura AWS existente)

**Secrets gerenciados**:
1. **mTLS Certificates** (`lb-conn/dict/bridge/mtls`)
   - `client_cert`: Certificado do PSP
   - `client_key`: Chave privada do PSP
   - `ca_cert`: CA do DICT BACEN
   - `server_name`: dict.pi.rsfn.net.br

2. **Bridge Endpoint** (`lb-conn/dict/bridge/endpoint`)
   - `host`: bridge.lb-conn.svc.cluster.local
   - `port`: 50051

3. **Database Credentials** (`lb-conn/dict/database/credentials`)
   - `username`: PostgreSQL user
   - `password`: PostgreSQL password
   - `host`: PostgreSQL host
   - `port`: 5432
   - `database`: connector_dict

**Implementação**: Ver [BRIDGE_ENDPOINTS_RATE_LIMIT.md](./.claude/BRIDGE_ENDPOINTS_RATE_LIMIT.md#autenticação-e-segurança)

**Responsável**: DevOps Engineer

---

## 📝 ATUALIZAÇÕES NECESSÁRIAS NO CLAUDE.MD

### Seção "Algoritmo Token Bucket"
```diff
- WARNING_THRESHOLD := Capacity * 0.25
+ WARNING_THRESHOLD := Capacity * 0.20
```

### Seção "Arquitetura de Integração"
```diff
- ┌──────────────────┐
- │  Cache Redis     │
- └──────────────────┘
+ ❌ Cache removido - sempre consultar DICT
```

### Seção "Stack Tecnológica"
```diff
- Cache: Redis (TTL 60s)
+ Cache: ❌ Removido (sempre consultar DICT)
```

### Seção "Métricas de Sucesso"
```diff
+ | Error 404 Rate | <20% | Prometheus counter |
+ | Recovery ETA | Calculado | Domain logic |
+ | Exhaustion Projection | <30min alert | Domain logic |
+ | Category Changes | Logged | Pulsar events |
```

### Seção "Observability"
```diff
- Grafana dashboards predefinidos
- PagerDuty/Slack integration
+ Prometheus AlertManager (local)
+ Grafana: Pós-lançamento (time de infra)
```

### Seção "Deployment"
```diff
+ Migration Tool: Goose
+ Kubernetes: Manifests diretos (sem Helm)
+ Secrets: ⚠️ TBD (Vault vs AWS Secrets Manager)
```

### Nova Seção "Timezone & Clock"
```yaml
Timezone Configuration:
  - Environment: TZ=UTC (forçar UTC)
  - Timestamp Source: DICT <ResponseTime> (autoridade)
  - All times: UTC (no local timezone conversions)
```

### Nova Seção "Anti-Scan Monitoring"
```go
// Monitoramento de erros 404 (anti-scan BACEN)
dict_rate_limit_404_rate{policy="ENTRIES_READ"} 0.15

// Alert rule
- alert: HighRateLimitErrorRate
  expr: rate(dict_rate_limit_404_errors_total[5m]) / rate(dict_rate_limit_requests_total[5m]) > 0.20
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "High 404 rate on {{ $labels.policy }} - possible anti-scan penalty"
```

---

## 🎯 IMPACTO POR FASE

### Fase 1: Database Layer
- ✅ Sem mudanças (schema já correto)
- ➕ Adicionar coluna `psp_category` em `dict_rate_limit_snapshots`
- ➕ Adicionar coluna `consumption_rate_per_minute` em histórico

### Fase 2: Domain & Algorithms
- 🔴 Corrigir threshold WARNING (0.25 → 0.20)
- ➕ Implementar `CalculateRecoveryETA()`
- ➕ Implementar `ProjectExhaustion()`
- ➕ Implementar `CategoryMonitor`

### Fase 3: Bridge Integration
- ✅ Sem mudanças (endpoints confirmados)
- ➕ Parsear `<ResponseTime>` do DICT

### Fase 4: Temporal Workflows
- ➖ Remover lógica de cache Redis
- ➕ Adicionar cálculo de ETA no CheckPolicyActivity
- ➕ Adicionar projeção de esgotamento

### Fase 5: Observability
- ➖ Remover Grafana dashboards do escopo inicial
- ➖ Remover PagerDuty/Slack integration
- ➕ Adicionar métrica de 404 rate
- ➕ Adicionar gauge de recovery ETA

### Fase 6: Testing
- 🔴 Atualizar testes de threshold (25% → 20%)
- ➕ Testes de ETA calculation
- ➕ Testes de projection
- ➕ Testes de category change detection

### Fase 7: Deployment
- ✅ Goose confirmado
- ✅ K8s manifests confirmados
- ⚠️ Secrets management pendente

---

## 📊 RESUMO DE ESFORÇO

| Mudança | Complexidade | Esforço Estimado | Prioridade |
|---------|--------------|------------------|------------|
| Threshold WARNING fix | Baixa | 2h | 🔴 Alta |
| Remover cache Redis | Média | 4h | 🔴 Alta |
| Remover Grafana/PagerDuty | Baixa | 1h | 🟡 Média |
| Métrica 404 rate | Baixa | 3h | 🟡 Média |
| ETA calculation | Média | 6h | 🟡 Média |
| Projection calculation | Média | 6h | 🟢 Baixa |
| Category monitoring | Baixa | 4h | 🟢 Baixa |
| DICT timestamp usage | Baixa | 2h | 🟢 Baixa |
| **TOTAL** | - | **28h** | - |

**Impacto no cronograma**: +1 semana (considerando refinamentos e testes adicionais)

---

## ✅ CHECKLIST DE VALIDAÇÃO

- [ ] CLAUDE.md atualizado com todas as mudanças
- [ ] Threshold WARNING corrigido (0.20)
- [ ] Referências a cache Redis removidas
- [ ] Grafana/PagerDuty movidos para pós-lançamento
- [ ] Métrica de 404 rate documentada
- [ ] ETA calculation documentada
- [ ] Projection documentada
- [ ] Category monitoring documentada
- [ ] DICT timestamp usage documentada
- [ ] Timezone UTC documentado
- [ ] Pendências claramente marcadas (Categoria PSP, Secrets)
- [ ] Squad notificada das mudanças
- [ ] Cronograma ajustado (+1 semana)

---

**Responsável pela Revisão**: Tech Lead
**Próximo Passo**: Atualizar CLAUDE.md e notificar squad
**Data Limite**: 2025-11-02 (antes do início da Fase 1)
