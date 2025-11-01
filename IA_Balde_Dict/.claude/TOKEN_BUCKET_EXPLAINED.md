# Token Bucket - Explicação Completa

## 🎯 Objetivo deste Documento

Explicar de forma clara e visual como funciona o **algoritmo Token Bucket** usado pelo DICT BACEN para rate limiting, incluindo:
- Como as fichas evoluem ao longo do tempo
- Fórmulas matemáticas de reposição e consumo
- Cenários práticos de uso
- Estratégias de monitoramento

---

## 📚 Conceito Base

### O que é Token Bucket?

Imagine um **balde físico** que contém **fichas** (tokens):

```
         ┌─────────────────────────┐
         │     BALDE DE FICHAS     │
         │                         │
         │   🪙🪙🪙🪙🪙🪙🪙🪙🪙🪙   │
         │   🪙🪙🪙🪙🪙🪙🪙🪙🪙🪙   │
         │   🪙🪙🪙🪙🪙🪙🪙🪙🪙🪙   │
         │   🪙🪙🪙🪙🪙🪙🪙🪙🪙🪙   │
         │                         │
         └─────────────────────────┘
              Capacity: 40 fichas
```

**Regras**:
1. ✅ **Reabastecimento automático**: A cada período (ex: 60s), novas fichas são adicionadas
2. ❌ **Consumo por operação**: Cada requisição à API consome 1 ficha
3. 🔴 **Bloqueio quando vazio**: Se não há fichas, operação é negada (HTTP 429)
4. 📊 **Capacidade máxima**: O balde não pode ter mais que N fichas

---

## 🔢 Parâmetros do DICT BACEN

### Estrutura de Resposta

Quando consultamos `GET /policies/{policy}`, o DICT retorna:

```xml
<GetPolicyResponse>
    <Category>A</Category>
    <Policy>
        <AvailableTokens>35000</AvailableTokens>     ← Fichas disponíveis AGORA
        <Capacity>36000</Capacity>                   ← Máximo de fichas no balde
        <RefillTokens>1200</RefillTokens>            ← Fichas adicionadas por período
        <RefillPeriodSec>60</RefillPeriodSec>       ← Período de reposição (segundos)
        <Name>ENTRIES_WRITE</Name>
    </Policy>
</GetPolicyResponse>
```

### Significado de Cada Parâmetro

| Parâmetro | Tipo | Descrição | Exemplo |
|-----------|------|-----------|---------|
| **AvailableTokens** | int | Fichas disponíveis no momento da consulta | 35,000 |
| **Capacity** | int | Capacidade máxima do balde | 36,000 |
| **RefillTokens** | int | Quantidade de fichas repostas por período | 1,200 |
| **RefillPeriodSec** | int | Período de reposição em segundos | 60 |

---

## ⏱️ Evolução das Fichas ao Longo do Tempo

### 1. Reposição Automática (Refill)

#### Fórmula Base
```
Taxa de reposição = RefillTokens / RefillPeriodSec

Exemplo ENTRIES_WRITE:
Taxa = 1,200 fichas / 60 segundos = 20 fichas/segundo
```

#### Implementação (Pseudo-código)
```go
func RefillBucket(bucket *TokenBucket) {
    // A cada RefillPeriodSec segundos
    ticker := time.NewTicker(bucket.RefillPeriodSec * time.Second)

    for range ticker.C {
        // Adicionar fichas
        newTokens := bucket.AvailableTokens + bucket.RefillTokens

        // Respeitar capacidade máxima
        if newTokens > bucket.Capacity {
            bucket.AvailableTokens = bucket.Capacity
        } else {
            bucket.AvailableTokens = newTokens
        }
    }
}
```

#### Timeline de Reposição

**Cenário**: Balde com 34,000 fichas (faltam 2,000 para o máximo)

```
t=0s:      AvailableTokens = 34,000
           [████████████████████████████████████░░░░] 94.4%

t=60s:     + RefillTokens (1,200)
           AvailableTokens = 34,000 + 1,200 = 35,200
           [██████████████████████████████████████░░] 97.8%

t=120s:    + RefillTokens (1,200)
           AvailableTokens = 35,200 + 1,200 = 36,400
           EXCEDEU Capacity!
           AvailableTokens = 36,000 (descarta 400 fichas)
           [████████████████████████████████████████] 100%

t=180s:    + RefillTokens (1,200)
           Já está no máximo, descarta 1,200 fichas
           AvailableTokens = 36,000
           [████████████████████████████████████████] 100%
```

**Conclusão**: Fichas excedentes são **descartadas** quando o balde atinge a capacidade máxima.

---

### 2. Consumo por Requisição

#### Regras de Consumo

Cada operação no DICT consome fichas:

| Operação | Consumo | Condição |
|----------|---------|----------|
| POST /entries | -1 ficha | Status ≠ 500 |
| GET /entries/{key} (sucesso) | -1 ficha | Status 200 |
| GET /entries/{key} (não encontrado) | **-3 fichas** | Status 404 (anti-scan) |
| POST /claims | -1 ficha | Status ≠ 500 |
| POST /refunds | -1 ficha | Status ≠ 500 |

**Observação**: Erro 404 consome **3x mais fichas** para desincentivar varredura (anti-scan).

#### Implementação (Pseudo-código)
```go
func ConsumeToken(bucket *TokenBucket, requestStatus int) error {
    tokensToConsume := 1

    // Anti-scan: erro 404 consome 3 fichas
    if requestStatus == 404 {
        tokensToConsume = 3
    }

    // Verificar disponibilidade
    if bucket.AvailableTokens < tokensToConsume {
        return errors.New("HTTP 429 - Too Many Requests")
    }

    // Consumir fichas
    bucket.AvailableTokens -= tokensToConsume
    return nil
}
```

#### Timeline de Consumo

**Cenário**: PSP fazendo requisições contínuas

```
t=0s:      AvailableTokens = 36,000
           PSP faz 1 requisição POST /entries
           AvailableTokens = 36,000 - 1 = 35,999
           [████████████████████████████████████████] 99.9%

t=1s:      PSP faz 20 requisições/segundo
           AvailableTokens = 35,999 - 20 = 35,979
           [████████████████████████████████████████] 99.9%

t=60s:     + RefillTokens (1,200)
           - Consumo médio (1,200 requisições em 60s)
           AvailableTokens = 35,979 + 1,200 - 1,200 = 35,979
           [████████████████████████████████████████] 99.9%

Equilíbrio: Consumo = Reposição → AvailableTokens estável
```

---

### 3. Cenário Crítico: Esgotamento do Balde

#### Situação: Burst de Requisições

```
Policy: ENTRIES_WRITE
Capacity: 36,000
RefillTokens: 1,200/min

t=0min:    AvailableTokens = 36,000
           [████████████████████████████████████████] 100%

           PSP inicia burst de 30,000 requisições

t=0.5min:  AvailableTokens = 36,000 - 15,000 = 21,000
           [█████████████████████░░░░░░░░░░░░░░░░░░░] 58.3%

t=1min:    AvailableTokens = 21,000 - 15,000 = 6,000
           + RefillTokens (1,200)
           AvailableTokens = 6,000 + 1,200 = 7,200
           [████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░] 20%
           ⚠️ WARNING: Apenas 20% disponível!

t=1.5min:  PSP tenta mais 10,000 requisições
           AvailableTokens = 7,200 - 7,200 = 0
           [░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░] 0%
           🔴 CRITICAL: Balde esgotado!

           Requisições seguintes retornam HTTP 429

t=2min:    + RefillTokens (1,200)
           AvailableTokens = 0 + 1,200 = 1,200
           [███░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░] 3.3%

           PSP recupera operação parcial (apenas 1,200 requisições)
```

#### Impacto no Negócio

Quando `AvailableTokens = 0`:

```xml
HTTP/1.1 429 Too Many Requests
Content-Type: application/xml

<?xml version="1.0" encoding="UTF-8" ?>
<Problem>
    <Type>https://dict.pi.rsfn.net.br/api/v2/error/TooManyRequests</Type>
    <Title>Too Many Requests</Title>
    <Status>429</Status>
    <Detail>Policy ENTRIES_WRITE exceeded rate limit. Retry after 60 seconds.</Detail>
</Problem>
```

**Consequências**:
- ❌ **Operações bloqueadas**: Chaves PIX não podem ser criadas
- 💰 **Perda de receita**: Transações PIX rejeitadas
- 📉 **SLA degradado**: Disponibilidade do serviço comprometida
- 👥 **Experiência do usuário**: Falhas na criação de chaves
- 🚨 **Incidentes**: Alertas para equipe de operações

**Tempo de recuperação**:
```
Tempo para recuperação total = (Capacity / RefillTokens) * RefillPeriodSec

Exemplo ENTRIES_WRITE:
Tempo = (36,000 / 1,200) * 60s = 30 * 60s = 1,800s = 30 minutos
```

Se o balde esvaziar completamente, leva **30 minutos** para reabastecer totalmente!

---

## 📊 Cálculo de Utilização e Thresholds

### Fórmula de Utilização

```go
utilization_pct := ((Capacity - AvailableTokens) / Capacity) * 100

// Alternativamente
utilization_pct := (1 - (AvailableTokens / Capacity)) * 100
```

### Exemplos de Cálculo

**Policy: ENTRIES_WRITE (Capacity = 36,000)**

| AvailableTokens | Utilização | Status |
|-----------------|------------|--------|
| 36,000 | 0% | ✅ Ótimo |
| 27,000 | 25% | ✅ Normal |
| 18,000 | 50% | ✅ Normal |
| 9,000 | 75% | ⚠️ WARNING |
| 5,000 | 86% | 🔴 CRITICAL |
| 3,600 | 90% | 🔴 CRITICAL |
| 1,000 | 97% | 🔴 CRITICAL |
| 0 | 100% | 💥 ESGOTADO |

### Thresholds Definidos

```go
const (
    WARNING_THRESHOLD  = 0.25  // 25% de capacidade restante
    CRITICAL_THRESHOLD = 0.10  // 10% de capacidade restante
)

func CalculateAlertLevel(available, capacity int) string {
    pct := float64(available) / float64(capacity)

    if pct < CRITICAL_THRESHOLD {
        return "CRITICAL"
    } else if pct < WARNING_THRESHOLD {
        return "WARNING"
    }
    return "NORMAL"
}

// Exemplo
available := 3000
capacity := 36000
level := CalculateAlertLevel(available, capacity)  // "CRITICAL"
```

---

## 🔍 Políticas do DICT BACEN

### 24 Políticas Monitoradas

#### Categoria A: Alto Volume (Operações Core)

| Política | Capacity | RefillTokens | Período | Taxa/min |
|----------|----------|--------------|---------|----------|
| ENTRIES_WRITE | 36,000 | 1,200 | 60s | 1,200/min |
| CLAIMS_WRITE | 36,000 | 1,200 | 60s | 1,200/min |
| REFUNDS_WRITE | 72,000 | 2,400 | 60s | 2,400/min |
| CIDS_ENTRIES_READ | 36,000 | 1,200 | 60s | 1,200/min |

#### Categoria B: Médio Volume (Consultas)

| Política | Capacity | RefillTokens | Período | Taxa/min |
|----------|----------|--------------|---------|----------|
| ENTRIES_UPDATE | 600 | 600 | 60s | 600/min |
| CLAIMS_READ | 18,000 | 600 | 60s | 600/min |
| INFRACTION_REPORTS_READ | 18,000 | 600 | 60s | 600/min |
| FRAUD_MARKERS_READ | 18,000 | 600 | 60s | 600/min |

#### Categoria C: Baixo Volume (Listagens)

| Política | Capacity | RefillTokens | Período | Taxa/min |
|----------|----------|--------------|---------|----------|
| CLAIMS_LIST_WITH_ROLE | 200 | 40 | 60s | 40/min |
| CLAIMS_LIST_WITHOUT_ROLE | 50 | 10 | 60s | 10/min |
| SYNC_VERIFICATIONS_WRITE | 50 | 10 | 60s | 10/min |
| CIDS_FILES_READ | 50 | 10 | 60s | 10/min |

#### Categoria D: Especial (Anti-Scan)

| Política | Capacity | RefillTokens | Período | Categoria PSP |
|----------|----------|--------------|---------|---------------|
| ENTRIES_READ_PARTICIPANT_ANTISCAN | 50,000 | 25,000 | 60s | A |
| ENTRIES_READ_PARTICIPANT_ANTISCAN | 40,000 | 20,000 | 60s | B |
| ENTRIES_READ_PARTICIPANT_ANTISCAN | 30,000 | 15,000 | 60s | C |
| ENTRIES_READ_PARTICIPANT_ANTISCAN | 16,000 | 8,000 | 60s | D |
| ENTRIES_READ_PARTICIPANT_ANTISCAN | 5,000 | 2,500 | 60s | E |
| ENTRIES_READ_PARTICIPANT_ANTISCAN | 500 | 250 | 60s | F |
| ENTRIES_READ_PARTICIPANT_ANTISCAN | 250 | 25 | 60s | G |
| ENTRIES_READ_PARTICIPANT_ANTISCAN | 50 | 2 | 60s | H |

### Volume Diário por Categoria (Anti-Scan)

```
Volume diário = (RefillTokens * 60min/h * 24h/dia) + Capacity inicial

Categoria A: (25,000 * 1,440) + 50,000 = 36,050,000 fichas/dia
Categoria B: (20,000 * 1,440) + 40,000 = 28,840,000 fichas/dia
Categoria C: (15,000 * 1,440) + 30,000 = 21,630,000 fichas/dia
Categoria D: (8,000 * 1,440) + 16,000 = 11,536,000 fichas/dia
Categoria E: (2,500 * 1,440) + 5,000 = 3,605,000 fichas/dia
Categoria F: (250 * 1,440) + 500 = 360,500 fichas/dia
Categoria G: (25 * 1,440) + 250 = 36,250 fichas/dia
Categoria H: (2 * 1,440) + 50 = 2,930 fichas/dia
```

---

## 🚨 Estratégia de Monitoramento

### Por que monitorar a cada 5 minutos?

#### Cenário 1: Policy com baixa capacidade

```
Policy: CLAIMS_LIST_WITHOUT_ROLE
Capacity: 50 fichas
RefillTokens: 10 fichas/min

Consumo médio: 8 fichas/min (uso normal)
Pico de consumo: 50 fichas em 1 minuto

Monitoramento a cada 5min:
  t=0min:   AvailableTokens = 50
  t=1min:   PICO! 50 - 50 = 0 (esgotado)
  t=5min:   Monitoramento detecta: 0 + (10 * 4) = 40 fichas
            Alerta disparado TARDE (4min após esgotamento)

Solução: Monitoramento + Thresholds preventivos
  t=0min:   AvailableTokens = 50 (100%)
  t=1min:   50 - 30 = 20 (40% disponível)
            ⚠️ ALERTA: Abaixo de WARNING (25%)
  t=2min:   20 - 15 = 5 (10% disponível)
            🔴 ALERTA CRÍTICO: Abaixo de CRITICAL (10%)
  t=3min:   5 - 5 = 0 (esgotado)
            💥 ALERTA URGENTE: Balde vazio
```

### Configuração de Alertas

```yaml
alerting:
  WARNING:
    threshold: 25%  # 25% de capacidade restante
    actions:
      - log_warning
      - publish_pulsar_event
      - update_prometheus_gauge

  CRITICAL:
    threshold: 10%  # 10% de capacidade restante
    actions:
      - log_error
      - publish_pulsar_event
      - update_prometheus_gauge
      - notify_pagerduty
      - send_slack_alert

  EXHAUSTED:
    threshold: 0%  # Balde vazio
    actions:
      - log_critical
      - publish_pulsar_event
      - update_prometheus_gauge
      - notify_pagerduty_urgent
      - send_slack_alert
      - create_incident
```

### Temporal Cron Workflow

```go
// Execução a cada 5 minutos
workflow := client.StartWorkflowOptions{
    ID: "monitor-rate-limits-cron",
    TaskQueue: "dict-task-queue",
    CronSchedule: "*/5 * * * *",
}

func MonitorRateLimitsWorkflow(ctx workflow.Context) error {
    // 1. Consultar DICT via Bridge
    var policies GetPoliciesResult
    workflow.ExecuteActivity(ctx, GetPoliciesActivity).Get(ctx, &policies)

    // 2. Armazenar em PostgreSQL
    workflow.ExecuteActivity(ctx, StorePolicyStateActivity, policies).Get(ctx, nil)

    // 3. Analisar thresholds
    var alerts AnalyzeBalanceResult
    workflow.ExecuteActivity(ctx, AnalyzeBalanceActivity, policies).Get(ctx, &alerts)

    // 4. Se há alertas, publicar
    if len(alerts.Alerts) > 0 {
        workflow.ExecuteActivity(ctx, PublishAlertActivity, alerts).Get(ctx, nil)
    }

    return nil
}
```

---

## 📈 Dashboards e Métricas

### Prometheus Metrics

```go
// Gauge: Fichas disponíveis por policy
dict_rate_limit_available_tokens{policy="ENTRIES_WRITE"} 35000

// Gauge: Capacidade máxima
dict_rate_limit_capacity{policy="ENTRIES_WRITE"} 36000

// Gauge: Utilização em percentual
dict_rate_limit_utilization_pct{policy="ENTRIES_WRITE"} 2.8

// Counter: Total de alertas
dict_rate_limit_alerts_total{policy="ENTRIES_WRITE",severity="WARNING"} 5
dict_rate_limit_alerts_total{policy="ENTRIES_WRITE",severity="CRITICAL"} 2
```

### Grafana Dashboard

```
┌────────────────────────────────────────────────────────────────┐
│  DICT Rate Limit Monitoring                                   │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  ENTRIES_WRITE                                                │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │                                                          │ │
│  │  Fichas: 35,000 / 36,000  (97.2%)                       │ │
│  │  [████████████████████████████████████████░░]           │ │
│  │                                                          │ │
│  │  Status: ✅ Normal                                       │ │
│  │  Última atualização: 2025-11-01 10:30:00                │ │
│  └──────────────────────────────────────────────────────────┘ │
│                                                                │
│  CLAIMS_WRITE                                                 │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │                                                          │ │
│  │  Fichas: 3,600 / 36,000  (90%)                          │ │
│  │  [█████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]            │ │
│  │                                                          │ │
│  │  Status: 🔴 CRITICAL                                     │ │
│  │  Última atualização: 2025-11-01 10:30:00                │ │
│  └──────────────────────────────────────────────────────────┘ │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

---

## 🔗 Referências

- **BACEN Manual Operacional**: Capítulo 19 - Consulta de Baldes
- **Wikipedia**: [Token Bucket Algorithm](https://en.wikipedia.org/wiki/Token_bucket)
- **RFC**: Network Traffic Shaping Algorithms
- **CLAUDE.md**: Especificação completa do projeto
- **SPECS-DATABASE.md**: Schema de armazenamento de estados
- **SPECS-WORKFLOWS.md**: Temporal workflows de monitoramento

---

**Última Atualização**: 2025-11-01
**Versão**: 1.0.0
**Autor**: Tech Lead - DICT Rate Limit Monitoring Project
