# SPECS-OBSERVABILITY.md - Observability & Monitoring Specification

**Projeto**: DICT Rate Limit Monitoring System
**Componentes**: OpenTelemetry + Prometheus + Grafana + AlertManager
**Status**: ✅ ESPECIFICAÇÃO COMPLETA - Production-Ready

---

## 🎯 Objetivo

Especificação técnica completa da **estratégia de observabilidade** do sistema:

1. **OpenTelemetry**: Instrumentação (logs + traces + metrics)
2. **Prometheus**: Coleta e armazenamento de métricas
3. **Grafana**: Dashboards visuais
4. **AlertManager**: Alertas críticos
5. **SLIs/SLOs**: Indicadores e objetivos de nível de serviço

**Meta**: Visibilidade completa do sistema em produção (zero blind spots).

---

## 📋 Tabela de Conteúdos

- [1. Arquitetura de Observabilidade](#1-arquitetura-de-observabilidade)
- [2. OpenTelemetry Instrumentation](#2-opentelemetry-instrumentation)
- [3. Prometheus Metrics](#3-prometheus-metrics)
- [4. Grafana Dashboards](#4-grafana-dashboards)
- [5. AlertManager Rules](#5-alertmanager-rules)
- [6. SLIs & SLOs](#6-slis--slos)
- [7. Logging Strategy](#7-logging-strategy)
- [8. Distributed Tracing](#8-distributed-tracing)
- [9. Runbooks](#9-runbooks)

---

## 1. Arquitetura de Observabilidade

### Stack de Observabilidade

```
┌─────────────────────────────────────────────────────────────────────┐
│                    DICT RATE LIMIT MONITORING                        │
│          (apps/dict + apps/orchestration-worker)                     │
└─────────────────────────────────────────────────────────────────────┘
                              │
            ┌─────────────────┼─────────────────┐
            │                 │                 │
            ▼                 ▼                 ▼
    ┌─────────────┐   ┌─────────────┐   ┌─────────────┐
    │   LOGS      │   │   TRACES    │   │  METRICS    │
    │  (stdout)   │   │  (OTLP)     │   │ (Prometheus)│
    └─────────────┘   └─────────────┘   └─────────────┘
            │                 │                 │
            │                 │                 │
            ▼                 ▼                 ▼
    ┌─────────────┐   ┌─────────────┐   ┌─────────────┐
    │   Loki      │   │   Tempo     │   │ Prometheus  │
    │ (log aggr)  │   │  (tracing)  │   │ (TSDB)      │
    └─────────────┘   └─────────────┘   └─────────────┘
            │                 │                 │
            └─────────────────┴─────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │    GRAFANA      │
                    │  (Dashboards)   │
                    └─────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │  ALERTMANAGER   │
                    │  (Slack/PagerDuty)│
                    └─────────────────┘
```

### Data Flow

```
Application Code
  ├─ OpenTelemetry SDK
  │   ├─ Logs → stdout → Loki
  │   ├─ Traces → OTLP Exporter → Tempo
  │   └─ Metrics → Prometheus Exporter → Prometheus
  │
  └─ Business Metrics
      └─ Prometheus Client → /metrics endpoint → Prometheus

Prometheus
  ├─ Scrape /metrics every 15s
  ├─ Evaluate alert rules every 1m
  └─ Send alerts → AlertManager → Slack/PagerDuty

Grafana
  ├─ Query Prometheus (metrics)
  ├─ Query Loki (logs)
  ├─ Query Tempo (traces)
  └─ Visualize in dashboards
```

---

## 2. OpenTelemetry Instrumentation

### OpenTelemetry Setup

```go
// Location: shared/telemetry/otel.go
package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// Config representa a configuração do OpenTelemetry
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
}

// Setup configura OpenTelemetry (traces + metrics)
func Setup(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	// Resource (service metadata)
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// ========================================================================
	// TRACE PROVIDER
	// ========================================================================

	// OTLP trace exporter
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Trace provider
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// ========================================================================
	// METRIC PROVIDER
	// ========================================================================

	// Prometheus exporter
	metricExporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// Metric provider
	metricProvider := metric.NewMeterProvider(
		metric.WithReader(metricExporter),
		metric.WithResource(res),
	)

	otel.SetMeterProvider(metricProvider)

	// ========================================================================
	// SHUTDOWN FUNCTION
	// ========================================================================

	shutdown := func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := traceProvider.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown trace provider: %w", err)
		}

		if err := metricProvider.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown metric provider: %w", err)
		}

		return nil
	}

	return shutdown, nil
}

// Tracer retorna um tracer global
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// Meter retorna um meter global
func Meter(name string) metric.Meter {
	return otel.Meter(name)
}
```

### HTTP Handler Instrumentation

```go
// Location: apps/dict/handlers/http/middleware/telemetry.go
package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// TelemetryMiddleware adiciona instrumentação OpenTelemetry
func TelemetryMiddleware(next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "http.request",
		otelhttp.WithServerName("dict-api"),
		otelhttp.WithSpanOptions(trace.WithSpanKind(trace.SpanKindServer)),
	)
}
```

---

## 3. Prometheus Metrics

### Business Metrics

```go
// Location: apps/dict/metrics/rate_limit.go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ========================================================================
	// POLICY METRICS (GAUGE)
	// ========================================================================

	// Available tokens por política
	AvailableTokensGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dict_rate_limit_available_tokens",
			Help: "Available tokens for each rate limit policy",
		},
		[]string{"policy_name", "category"},
	)

	// Utilização percentual por política
	UtilizationGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dict_rate_limit_utilization_pct",
			Help: "Utilization percentage for each rate limit policy (0-100)",
		},
		[]string{"policy_name", "category"},
	)

	// ========================================================================
	// ALERT METRICS (COUNTER)
	// ========================================================================

	// Total de alertas disparados
	AlertsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dict_rate_limit_alerts_total",
			Help: "Total number of rate limit alerts triggered",
		},
		[]string{"policy_name", "severity"}, // severity: WARNING, CRITICAL
	)

	// ========================================================================
	// WORKFLOW METRICS (HISTOGRAM)
	// ========================================================================

	// Duração da execução do workflow de monitoramento
	WorkflowDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dict_rate_limit_workflow_duration_seconds",
			Help:    "Duration of MonitorRateLimitsWorkflow execution",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
		},
		[]string{"status"}, // status: success, failed
	)

	// Duração de execução de cada activity
	ActivityDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dict_rate_limit_activity_duration_seconds",
			Help:    "Duration of Temporal activity execution",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"activity_name", "status"},
	)

	// ========================================================================
	// BRIDGE METRICS (HISTOGRAM + COUNTER)
	// ========================================================================

	// Latência de chamadas ao Bridge
	BridgeCallDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dict_bridge_call_duration_seconds",
			Help:    "Duration of Bridge gRPC calls",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"method", "status"}, // method: ListPolicies, GetPolicy
	)

	// Total de erros em chamadas ao Bridge
	BridgeErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dict_bridge_errors_total",
			Help: "Total number of Bridge gRPC errors",
		},
		[]string{"method", "error_code"}, // error_code: Unavailable, DeadlineExceeded, etc
	)

	// ========================================================================
	// PULSAR METRICS (COUNTER + HISTOGRAM)
	// ========================================================================

	// Total de eventos publicados no Pulsar
	PulsarPublishedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dict_pulsar_published_total",
			Help: "Total number of events published to Pulsar",
		},
		[]string{"action", "status"}, // action: ActionRateLimitAlert
	)

	// Latência de publicação no Pulsar
	PulsarPublishDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dict_pulsar_publish_duration_seconds",
			Help:    "Duration of Pulsar event publish",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
		},
		[]string{"action"},
	)

	// ========================================================================
	// DATABASE METRICS (HISTOGRAM + COUNTER)
	// ========================================================================

	// Latência de queries ao PostgreSQL
	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dict_database_query_duration_seconds",
			Help:    "Duration of PostgreSQL queries",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
		},
		[]string{"query_name", "status"},
	)

	// Total de erros em operações de banco
	DatabaseErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dict_database_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"operation", "error_type"},
	)

	// ========================================================================
	// CACHE METRICS (COUNTER)
	// ========================================================================

	// Total de cache hits/misses
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dict_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_key", "status"}, // status: hit, miss
	)
)
```

### Metrics Exporter (HTTP Endpoint)

```go
// Location: apps/dict/handlers/http/metrics/handler.go
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler retorna o handler de métricas Prometheus
func Handler() http.Handler {
	return promhttp.Handler()
}

// RegisterRoutes registra o endpoint /metrics
func RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/metrics", Handler())
}
```

---

## 4. Grafana Dashboards

### Dashboard 1: Rate Limit Overview

```json
{
  "dashboard": {
    "title": "DICT Rate Limit Monitoring - Overview",
    "tags": ["dict", "rate-limit", "bacen"],
    "timezone": "UTC",
    "refresh": "30s",
    "panels": [
      {
        "id": 1,
        "title": "Available Tokens per Policy",
        "type": "graph",
        "targets": [
          {
            "expr": "dict_rate_limit_available_tokens",
            "legendFormat": "{{policy_name}} ({{category}})",
            "refId": "A"
          }
        ],
        "yaxes": [
          {
            "label": "Tokens",
            "format": "short"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "Utilization % per Policy",
        "type": "graph",
        "targets": [
          {
            "expr": "dict_rate_limit_utilization_pct",
            "legendFormat": "{{policy_name}}",
            "refId": "A"
          }
        ],
        "yaxes": [
          {
            "label": "Utilization %",
            "format": "percent",
            "max": 100,
            "min": 0
          }
        ],
        "thresholds": [
          {"value": 75, "color": "yellow"},
          {"value": 90, "color": "red"}
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0}
      },
      {
        "id": 3,
        "title": "Alerts by Severity (Last 24h)",
        "type": "stat",
        "targets": [
          {
            "expr": "sum(increase(dict_rate_limit_alerts_total{severity=\"WARNING\"}[24h]))",
            "refId": "A"
          },
          {
            "expr": "sum(increase(dict_rate_limit_alerts_total{severity=\"CRITICAL\"}[24h]))",
            "refId": "B"
          }
        ],
        "gridPos": {"h": 4, "w": 6, "x": 0, "y": 8}
      },
      {
        "id": 4,
        "title": "Policies by Status",
        "type": "piechart",
        "targets": [
          {
            "expr": "count(dict_rate_limit_utilization_pct < 75)",
            "legendFormat": "OK",
            "refId": "A"
          },
          {
            "expr": "count(dict_rate_limit_utilization_pct >= 75 < 90)",
            "legendFormat": "WARNING",
            "refId": "B"
          },
          {
            "expr": "count(dict_rate_limit_utilization_pct >= 90)",
            "legendFormat": "CRITICAL",
            "refId": "C"
          }
        ],
        "gridPos": {"h": 4, "w": 6, "x": 6, "y": 8}
      },
      {
        "id": 5,
        "title": "Top 5 Most Utilized Policies",
        "type": "table",
        "targets": [
          {
            "expr": "topk(5, dict_rate_limit_utilization_pct)",
            "format": "table",
            "refId": "A"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 12}
      },
      {
        "id": 6,
        "title": "Workflow Execution Rate (per minute)",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(dict_rate_limit_workflow_duration_seconds_count[5m])",
            "legendFormat": "{{status}}",
            "refId": "A"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 12}
      }
    ]
  }
}
```

### Dashboard 2: Performance & SLIs

```json
{
  "dashboard": {
    "title": "DICT Rate Limit - Performance & SLIs",
    "tags": ["dict", "performance", "sli"],
    "timezone": "UTC",
    "refresh": "30s",
    "panels": [
      {
        "id": 1,
        "title": "API Response Time (p99)",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{handler=\"list-policies\"}[5m]))",
            "legendFormat": "p99",
            "refId": "A"
          },
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{handler=\"list-policies\"}[5m]))",
            "legendFormat": "p95",
            "refId": "B"
          }
        ],
        "yaxes": [
          {
            "label": "Response Time (s)",
            "format": "s"
          }
        ],
        "alert": {
          "name": "API Response Time High",
          "conditions": [
            {
              "evaluator": {"type": "gt", "params": [0.2]},
              "query": {"params": ["A", "5m", "now"]},
              "reducer": {"type": "avg"}
            }
          ]
        },
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "Bridge gRPC Latency (p99)",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.99, rate(dict_bridge_call_duration_seconds_bucket[5m]))",
            "legendFormat": "{{method}}",
            "refId": "A"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0}
      },
      {
        "id": 3,
        "title": "Database Query Latency (p99)",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.99, rate(dict_database_query_duration_seconds_bucket[5m]))",
            "legendFormat": "{{query_name}}",
            "refId": "A"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 8}
      },
      {
        "id": 4,
        "title": "Cache Hit Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(dict_cache_hits_total{status=\"hit\"}[5m])) / sum(rate(dict_cache_hits_total[5m])) * 100",
            "legendFormat": "Hit Rate %",
            "refId": "A"
          }
        ],
        "yaxes": [
          {
            "label": "Hit Rate %",
            "format": "percent",
            "max": 100,
            "min": 0
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 8}
      },
      {
        "id": 5,
        "title": "Workflow Success Rate (Last 1h)",
        "type": "stat",
        "targets": [
          {
            "expr": "sum(rate(dict_rate_limit_workflow_duration_seconds_count{status=\"success\"}[1h])) / sum(rate(dict_rate_limit_workflow_duration_seconds_count[1h])) * 100",
            "refId": "A"
          }
        ],
        "thresholds": [
          {"value": 99, "color": "green"},
          {"value": 95, "color": "yellow"},
          {"value": 0, "color": "red"}
        ],
        "gridPos": {"h": 4, "w": 6, "x": 0, "y": 16}
      },
      {
        "id": 6,
        "title": "Error Rate (Last 1h)",
        "type": "stat",
        "targets": [
          {
            "expr": "sum(rate(dict_bridge_errors_total[1h])) + sum(rate(dict_database_errors_total[1h]))",
            "refId": "A"
          }
        ],
        "thresholds": [
          {"value": 0, "color": "green"},
          {"value": 1, "color": "yellow"},
          {"value": 10, "color": "red"}
        ],
        "gridPos": {"h": 4, "w": 6, "x": 6, "y": 16}
      }
    ]
  }
}
```

---

## 5. AlertManager Rules

### Alert Rules Configuration

```yaml
# Location: configs/prometheus/alert_rules.yml
groups:
  - name: dict_rate_limit_alerts
    interval: 1m
    rules:
      # ======================================================================
      # CRITICAL ALERTS (Page immediately)
      # ======================================================================

      - alert: DictPolicyCriticalUtilization
        expr: dict_rate_limit_utilization_pct >= 90
        for: 5m
        labels:
          severity: critical
          service: dict-rate-limit
        annotations:
          summary: "CRITICAL: Policy {{ $labels.policy_name }} at {{ $value }}% utilization"
          description: "Policy {{ $labels.policy_name }} (Category {{ $labels.category }}) has exceeded 90% utilization for 5 minutes. Immediate action required."
          runbook_url: "https://runbooks.lbpay.com/dict/rate-limit-critical"

      - alert: DictWorkflowFailureRate
        expr: |
          sum(rate(dict_rate_limit_workflow_duration_seconds_count{status="failed"}[5m])) /
          sum(rate(dict_rate_limit_workflow_duration_seconds_count[5m])) > 0.05
        for: 5m
        labels:
          severity: critical
          service: dict-rate-limit
        annotations:
          summary: "CRITICAL: Workflow failure rate > 5%"
          description: "MonitorRateLimitsWorkflow is failing at {{ $value | humanizePercentage }} rate for 5 minutes."
          runbook_url: "https://runbooks.lbpay.com/dict/workflow-failures"

      - alert: DictBridgeUnavailable
        expr: sum(rate(dict_bridge_errors_total{error_code="Unavailable"}[5m])) > 0
        for: 2m
        labels:
          severity: critical
          service: dict-rate-limit
        annotations:
          summary: "CRITICAL: Bridge gRPC service unavailable"
          description: "Bridge service is returning Unavailable errors for 2 minutes. Rate limit monitoring is blocked."
          runbook_url: "https://runbooks.lbpay.com/dict/bridge-unavailable"

      # ======================================================================
      # WARNING ALERTS (Notify but no page)
      # ======================================================================

      - alert: DictPolicyWarningUtilization
        expr: dict_rate_limit_utilization_pct >= 75 and dict_rate_limit_utilization_pct < 90
        for: 10m
        labels:
          severity: warning
          service: dict-rate-limit
        annotations:
          summary: "WARNING: Policy {{ $labels.policy_name }} at {{ $value }}% utilization"
          description: "Policy {{ $labels.policy_name }} has exceeded 75% utilization for 10 minutes."

      - alert: DictAPIResponseTimeSlow
        expr: |
          histogram_quantile(0.99,
            rate(http_request_duration_seconds_bucket{handler=~"list-policies|get-policy"}[5m])
          ) > 0.2
        for: 5m
        labels:
          severity: warning
          service: dict-rate-limit
        annotations:
          summary: "WARNING: API p99 response time > 200ms"
          description: "Dict API response time (p99) is {{ $value }}s for 5 minutes. Target is < 200ms."

      - alert: DictCacheHitRateLow
        expr: |
          sum(rate(dict_cache_hits_total{status="hit"}[5m])) /
          sum(rate(dict_cache_hits_total[5m])) < 0.9
        for: 10m
        labels:
          severity: warning
          service: dict-rate-limit
        annotations:
          summary: "WARNING: Cache hit rate < 90%"
          description: "Redis cache hit rate is {{ $value | humanizePercentage }}. Target is > 90%."

      - alert: DictDatabaseQuerySlow
        expr: |
          histogram_quantile(0.99,
            rate(dict_database_query_duration_seconds_bucket[5m])
          ) > 0.05
        for: 5m
        labels:
          severity: warning
          service: dict-rate-limit
        annotations:
          summary: "WARNING: Database query p99 > 50ms"
          description: "PostgreSQL query latency (p99) is {{ $value }}s. Target is < 50ms."

      # ======================================================================
      # INFO ALERTS (Log only)
      # ======================================================================

      - alert: DictMultipleAlertsTriggered
        expr: increase(dict_rate_limit_alerts_total[1h]) > 10
        labels:
          severity: info
          service: dict-rate-limit
        annotations:
          summary: "INFO: {{ $value }} alerts triggered in last hour"
          description: "Multiple rate limit alerts have been triggered. Review policy thresholds."
```

### AlertManager Configuration

```yaml
# Location: configs/alertmanager/alertmanager.yml
global:
  resolve_timeout: 5m
  slack_api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK/URL'

route:
  receiver: 'default'
  group_by: ['severity', 'service']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h

  routes:
    # Critical alerts → PagerDuty + Slack
    - match:
        severity: critical
      receiver: 'pagerduty-critical'
      continue: true

    - match:
        severity: critical
      receiver: 'slack-critical'

    # Warning alerts → Slack only
    - match:
        severity: warning
      receiver: 'slack-warnings'

    # Info alerts → Slack (separate channel)
    - match:
        severity: info
      receiver: 'slack-info'

receivers:
  - name: 'default'
    slack_configs:
      - channel: '#dict-alerts'
        title: 'DICT Rate Limit Alert'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'

  - name: 'pagerduty-critical'
    pagerduty_configs:
      - service_key: 'YOUR_PAGERDUTY_SERVICE_KEY'
        severity: 'critical'

  - name: 'slack-critical'
    slack_configs:
      - channel: '#dict-critical-alerts'
        title: '🚨 CRITICAL: {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}\n{{ .Annotations.runbook_url }}{{ end }}'
        color: 'danger'

  - name: 'slack-warnings'
    slack_configs:
      - channel: '#dict-alerts'
        title: '⚠️ WARNING: {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
        color: 'warning'

  - name: 'slack-info'
    slack_configs:
      - channel: '#dict-info'
        title: 'ℹ️ INFO: {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
```

---

## 6. SLIs & SLOs

### Service Level Indicators (SLIs)

```yaml
SLIs:
  # Availability SLI
  - name: API Availability
    metric: |
      sum(rate(http_requests_total{handler=~"list-policies|get-policy",code=~"2.."}[5m])) /
      sum(rate(http_requests_total{handler=~"list-policies|get-policy"}[5m]))
    target: ">= 99.9%"
    measurement: "Percentage of successful HTTP requests (2xx status)"

  # Latency SLI
  - name: API Response Time (p99)
    metric: |
      histogram_quantile(0.99,
        rate(http_request_duration_seconds_bucket{handler=~"list-policies|get-policy"}[5m])
      )
    target: "< 200ms"
    measurement: "99th percentile response time"

  # Quality SLI
  - name: Workflow Success Rate
    metric: |
      sum(rate(dict_rate_limit_workflow_duration_seconds_count{status="success"}[1h])) /
      sum(rate(dict_rate_limit_workflow_duration_seconds_count[1h]))
    target: ">= 99%"
    measurement: "Percentage of successful workflow executions"

  # Freshness SLI
  - name: Data Freshness
    metric: |
      time() - max(dict_rate_limit_states_checked_at_timestamp)
    target: "< 5 minutes"
    measurement: "Time since last policy state update"
```

### Service Level Objectives (SLOs)

```yaml
SLOs:
  - name: Monthly API Availability SLO
    sli: API Availability
    objective: 99.9%
    period: 30 days
    error_budget: 0.1% (43.2 minutes/month)

  - name: Monthly Workflow Reliability SLO
    sli: Workflow Success Rate
    objective: 99%
    period: 30 days
    error_budget: 1% (288 failed workflows/month)

  - name: Real-time Response Time SLO
    sli: API Response Time (p99)
    objective: "< 200ms"
    period: 1 hour
    violation_threshold: "> 200ms for > 5 minutes"
```

---

## 7. Logging Strategy

### Structured Logging

```go
// Location: shared/logger/logger.go
package logger

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// Logger interface
type Logger interface {
	InfoContext(ctx context.Context, msg string, args ...interface{})
	WarnContext(ctx context.Context, msg string, args ...interface{})
	ErrorContext(ctx context.Context, msg string, args ...interface{})
}

// slogLogger implementa Logger usando slog
type slogLogger struct {
	logger *slog.Logger
}

// NewLogger cria um novo logger
func NewLogger(serviceName string) Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler).With(
		slog.String("service", serviceName),
	)

	return &slogLogger{logger: logger}
}

// InfoContext implementa Logger
func (l *slogLogger) InfoContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.InfoContext(ctx, msg, l.withTrace(ctx, args...)...)
}

// WarnContext implementa Logger
func (l *slogLogger) WarnContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.WarnContext(ctx, msg, l.withTrace(ctx, args...)...)
}

// ErrorContext implementa Logger
func (l *slogLogger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.ErrorContext(ctx, msg, l.withTrace(ctx, args...)...)
}

// withTrace adiciona trace ID ao log
func (l *slogLogger) withTrace(ctx context.Context, args ...interface{}) []interface{} {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		args = append(args,
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}
	return args
}
```

### Log Format Example

```json
{
  "time": "2025-10-31T10:30:00Z",
  "level": "INFO",
  "msg": "policies retrieved successfully",
  "service": "dict-api",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "count": 24,
  "cached": true
}
```

---

## 8. Distributed Tracing

### Trace Example (Tempo/Jaeger)

```
Trace: MonitorRateLimitsWorkflow
├─ Span: workflow.execute (60s)
│  ├─ Span: GetPoliciesActivity (2s)
│  │  └─ Span: Bridge.gRPC.ListPolicies (1.5s)
│  │     └─ Span: DICT.HTTP.POST /api/v1/policies (1.2s)
│  │
│  ├─ Span: StorePolicyStateActivity (500ms)
│  │  └─ Span: PostgreSQL.BatchInsert (450ms)
│  │
│  ├─ Span: AnalyzeBalanceActivity (50ms)
│  │
│  ├─ Span: PublishAlertActivity (200ms)
│  │  └─ Span: Pulsar.Publish (150ms)
│  │
│  └─ Span: PublishMetricsActivity (10ms)
```

---

## 9. Runbooks

### Runbook 1: Policy Critical Utilization

```markdown
# Runbook: Policy Critical Utilization

**Alert**: DictPolicyCriticalUtilization
**Severity**: CRITICAL
**Trigger**: Policy utilization >= 90% for 5 minutes

## Symptoms
- Specific DICT policy has < 10% tokens remaining
- Risk of 429 errors from DICT BACEN

## Investigation
1. Check Grafana dashboard "Rate Limit Overview"
2. Identify which policy is critical
3. Check recent traffic spike in application logs
4. Verify if legitimate traffic or attack

## Resolution
### Immediate (< 5 min)
1. Review application traffic patterns
2. If attack: Enable rate limiting at API Gateway
3. If legitimate: Request BACEN quota increase (manual process)

### Short-term (< 1 hour)
1. Implement request throttling in application
2. Add caching for duplicate requests
3. Optimize batch operations

### Long-term
1. Negotiate higher quota with BACEN
2. Implement auto-throttling based on bucket state
3. Add predictive alerts (trend analysis)

## Escalation
- L1: DevOps on-call
- L2: Backend Tech Lead
- L3: Architecture Team + BACEN Contact
```

---

**Última Atualização**: 2025-10-31
**Versão**: 1.0.0
**Status**: ✅ ESPECIFICAÇÃO COMPLETA - Production-Ready
