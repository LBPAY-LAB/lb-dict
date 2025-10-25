# DEV-005: Monitoring & Observability Stack - DICT

**Projeto**: DICT - Diretorio de Identificadores de Contas Transacionais (LBPay)
**Componente**: Complete Monitoring & Observability Stack
**Versao**: 1.0
**Data**: 2025-10-25
**Autor**: DEVOPS (AI Agent - DevOps Engineer)
**Revisor**: [Aguardando]
**Aprovador**: Head de DevOps, CTO (Jose Luis Silva)

---

## Controle de Versao

| Versao | Data | Autor | Descricao das Mudancas |
|--------|------|-------|------------------------|
| 1.0 | 2025-10-25 | DEVOPS | Versao inicial - Stack completo de observabilidade para DICT |

---

## Sumario Executivo

### Visao Geral

Stack completo de **Monitoring & Observability** para o **DICT**, implementando os tres pilares:
- **Metrics**: Prometheus + Grafana (dashboards, queries, alertas)
- **Traces**: Jaeger (distributed tracing)
- **Logs**: Loki + Promtail (log aggregation)

### Objetivos

- Garantir visibilidade completa do sistema em producao
- Detectar problemas antes que afetem usuarios
- Fornecer dados para troubleshooting rapido
- Medir SLOs (Service Level Objectives)
- Compliance com requisitos do Bacen (auditoria, rastreabilidade)

### SLOs (Service Level Objectives)

| Metrica | Objetivo | Medicao |
|---------|----------|---------|
| **Availability** | 99.9% | Uptime mensal |
| **Latency P95** | < 2s | Requisicoes DICT |
| **Latency P99** | < 5s | Requisicoes DICT |
| **Error Rate** | < 1% | Erros 5xx / Total |
| **Bacen Timeout** | < 0.1% | Timeout em chamadas RSFN |
| **Workflow Success** | > 99% | Workflows concluidos com sucesso |

---

## Indice

1. [Arquitetura de Observabilidade](#1-arquitetura-de-observabilidade)
2. [Prometheus Stack](#2-prometheus-stack)
3. [Grafana Dashboards](#3-grafana-dashboards)
4. [Jaeger Distributed Tracing](#4-jaeger-distributed-tracing)
5. [Loki Log Aggregation](#5-loki-log-aggregation)
6. [Alerting Strategy](#6-alerting-strategy)
7. [SLOs & SLIs](#7-slos--slis)
8. [Incident Response](#8-incident-response)
9. [Rastreabilidade](#9-rastreabilidade)

---

## 1. Arquitetura de Observabilidade

### Stack Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                    Observability Stack (EKS)                      │
├──────────────────────────────────────────────────────────────────┤
│                                                                    │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  Applications (dict-prod namespace)                        │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐                 │  │
│  │  │Core DICT │  │ Connect  │  │  Bridge  │                 │  │
│  │  │          │  │          │  │          │                 │  │
│  │  │ Exports: │  │ Exports: │  │ Exports: │                 │  │
│  │  │ - Metrics│  │ - Metrics│  │ - Metrics│                 │  │
│  │  │ - Traces │  │ - Traces │  │ - Traces │                 │  │
│  │  │ - Logs   │  │ - Logs   │  │ - Logs   │                 │  │
│  │  └──────────┘  └──────────┘  └──────────┘                 │  │
│  └────────────────────────────────────────────────────────────┘  │
│                          │                                        │
│                          ▼                                        │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  Collection Layer (monitoring namespace)                   │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐                 │  │
│  │  │Prometheus│  │  Jaeger  │  │ Promtail │                 │  │
│  │  │  Server  │  │Collector │  │ DaemonSet│                 │  │
│  │  │          │  │          │  │          │                 │  │
│  │  │  Scrapes │  │ Receives │  │ Collects │                 │  │
│  │  │  /metrics│  │  Traces  │  │   Logs   │                 │  │
│  │  └──────────┘  └──────────┘  └──────────┘                 │  │
│  └────────────────────────────────────────────────────────────┘  │
│                          │                                        │
│                          ▼                                        │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  Storage Layer                                             │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐                 │  │
│  │  │Prometheus│  │  Jaeger  │  │   Loki   │                 │  │
│  │  │  TSDB    │  │  Storage │  │  Storage │                 │  │
│  │  │ (15 days)│  │ (S3 7d)  │  │ (S3 30d) │                 │  │
│  │  └──────────┘  └──────────┘  └──────────┘                 │  │
│  └────────────────────────────────────────────────────────────┘  │
│                          │                                        │
│                          ▼                                        │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  Visualization & Alerting                                  │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐                 │  │
│  │  │ Grafana  │  │Jaeger UI │  │AlertMgr  │                 │  │
│  │  │Dashboards│  │ Tracing  │  │  Slack   │                 │  │
│  │  │ Queries  │  │  Queries │  │ PagerDuty│                 │  │
│  │  └──────────┘  └──────────┘  └──────────┘                 │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                    │
└──────────────────────────────────────────────────────────────────┘
```

### Pilares de Observabilidade

```yaml
Metrics (Prometheus):
  - Metricas de negocio: transacoes, latencia, throughput
  - Metricas de sistema: CPU, memoria, disco, rede
  - Metricas de infraestrutura: Kubernetes, containers
  Retencao: 15 dias (TSDB local)

Traces (Jaeger):
  - Distributed tracing end-to-end
  - Visualizacao de chamadas entre servicos
  - Analise de latencia por componente
  Retencao: 7 dias (S3)

Logs (Loki):
  - Logs estruturados (JSON)
  - Agregacao de logs de todos os pods
  - Correlacao com traces via trace_id
  Retencao: 30 dias (S3)
```

---

## 2. Prometheus Stack

### Deployment

```yaml
# prometheus-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus
  namespace: monitoring
  labels:
    app: prometheus
spec:
  replicas: 2
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
    spec:
      serviceAccountName: prometheus-sa
      containers:
      - name: prometheus
        image: prom/prometheus:v2.48.0
        args:
          - '--config.file=/etc/prometheus/prometheus.yml'
          - '--storage.tsdb.path=/prometheus'
          - '--storage.tsdb.retention.time=15d'
          - '--web.enable-lifecycle'
          - '--web.enable-admin-api'
        ports:
        - name: web
          containerPort: 9090
        resources:
          requests:
            cpu: 1000m
            memory: 2Gi
          limits:
            cpu: 2000m
            memory: 4Gi
        volumeMounts:
        - name: config
          mountPath: /etc/prometheus
        - name: storage
          mountPath: /prometheus
      volumes:
      - name: config
        configMap:
          name: prometheus-config
      - name: storage
        persistentVolumeClaim:
          claimName: prometheus-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: prometheus-svc
  namespace: monitoring
spec:
  type: ClusterIP
  selector:
    app: prometheus
  ports:
  - name: web
    port: 9090
    targetPort: 9090
```

### Prometheus Configuration

```yaml
# prometheus-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  namespace: monitoring
data:
  prometheus.yml: |
    global:
      scrape_interval: 30s
      scrape_timeout: 10s
      evaluation_interval: 30s
      external_labels:
        cluster: eks-lbpay-prod
        environment: production

    rule_files:
      - /etc/prometheus/rules/*.yml

    alerting:
      alertmanagers:
      - static_configs:
        - targets:
          - alertmanager-svc.monitoring.svc.cluster.local:9093

    scrape_configs:
      # Core DICT metrics
      - job_name: 'core-dict'
        kubernetes_sd_configs:
        - role: pod
          namespaces:
            names:
            - dict-prod
        relabel_configs:
        - source_labels: [__meta_kubernetes_pod_label_app]
          action: keep
          regex: core-dict
        - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
          action: keep
          regex: true
        - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_port]
          action: replace
          target_label: __address__
          regex: ([^:]+)(?::\d+)?;(\d+)
          replacement: $1:$2
        - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
          action: replace
          target_label: __metrics_path__
          regex: (.+)
        - source_labels: [__meta_kubernetes_namespace]
          action: replace
          target_label: namespace
        - source_labels: [__meta_kubernetes_pod_name]
          action: replace
          target_label: pod

      # Connect API metrics
      - job_name: 'connect-api'
        kubernetes_sd_configs:
        - role: pod
          namespaces:
            names:
            - dict-prod
        relabel_configs:
        - source_labels: [__meta_kubernetes_pod_label_app]
          action: keep
          regex: connect-api

      # Connect Worker metrics
      - job_name: 'connect-worker'
        kubernetes_sd_configs:
        - role: pod
          namespaces:
            names:
            - dict-prod
        relabel_configs:
        - source_labels: [__meta_kubernetes_pod_label_app]
          action: keep
          regex: connect-worker

      # RSFN Bridge metrics
      - job_name: 'rsfn-bridge'
        kubernetes_sd_configs:
        - role: pod
          namespaces:
            names:
            - dict-prod
        relabel_configs:
        - source_labels: [__meta_kubernetes_pod_label_app]
          action: keep
          regex: rsfn-bridge

      # Temporal metrics
      - job_name: 'temporal'
        kubernetes_sd_configs:
        - role: pod
          namespaces:
            names:
            - infrastructure
        relabel_configs:
        - source_labels: [__meta_kubernetes_pod_label_app]
          action: keep
          regex: temporal

      # Kubernetes nodes
      - job_name: 'kubernetes-nodes'
        kubernetes_sd_configs:
        - role: node
        relabel_configs:
        - action: labelmap
          regex: __meta_kubernetes_node_label_(.+)

      # Kubernetes pods
      - job_name: 'kubernetes-pods'
        kubernetes_sd_configs:
        - role: pod
        relabel_configs:
        - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
          action: keep
          regex: true
```

### Prometheus Metrics Examples

```yaml
# Metricas DICT Custom

# 1. Business Metrics
dict_entries_total{operation="create"}
dict_entries_total{operation="update"}
dict_entries_total{operation="delete"}
dict_entries_total{operation="query"}

dict_operation_duration_seconds{operation="create"}
dict_operation_duration_seconds{operation="query"}

dict_portability_requests_total{status="pending"}
dict_portability_requests_total{status="confirmed"}
dict_portability_requests_total{status="cancelled"}

dict_claim_requests_total{type="ownership"}
dict_claim_requests_total{type="portability"}

# 2. RSFN Bridge Metrics
rsfn_requests_total{operation="CreateEntry", status="success"}
rsfn_requests_total{operation="CreateEntry", status="error"}
rsfn_request_duration_seconds{operation="CreateEntry"}

rsfn_bacen_timeout_total{operation="CreateEntry"}
rsfn_circuit_breaker_state{service="bacen"} # 0=closed, 1=open, 2=half-open

# 3. Temporal Workflow Metrics
temporal_workflow_started_total{workflow_type="CreateEntryWorkflow"}
temporal_workflow_completed_total{workflow_type="CreateEntryWorkflow"}
temporal_workflow_failed_total{workflow_type="CreateEntryWorkflow"}
temporal_workflow_duration_seconds{workflow_type="CreateEntryWorkflow"}

temporal_activity_execution_total{activity_type="ValidateEntry"}
temporal_activity_failed_total{activity_type="SendToBacen"}

# 4. Infrastructure Metrics
postgres_connections_active
postgres_transaction_duration_seconds
redis_cache_hits_total
redis_cache_misses_total
pulsar_messages_published_total
pulsar_messages_consumed_total

# 5. gRPC Metrics (auto-instrumented)
grpc_server_started_total{grpc_service="dict.v1.DictService", grpc_method="CreateEntry"}
grpc_server_handled_total{grpc_service="dict.v1.DictService", grpc_method="CreateEntry", grpc_code="OK"}
grpc_server_handling_seconds{grpc_service="dict.v1.DictService", grpc_method="CreateEntry"}
```

### Example Prometheus Queries

```promql
# 1. Taxa de requisicoes por segundo (QPS)
rate(grpc_server_started_total{namespace="dict-prod"}[5m])

# 2. Latencia P95 das requisicoes gRPC
histogram_quantile(0.95,
  rate(grpc_server_handling_seconds_bucket{namespace="dict-prod"}[5m])
)

# 3. Taxa de erro (5xx)
sum(rate(grpc_server_handled_total{grpc_code!="OK"}[5m]))
/
sum(rate(grpc_server_handled_total[5m])) * 100

# 4. Availability (uptime)
avg(up{job="core-dict"}) * 100

# 5. Timeout do Bacen (ultimos 5 minutos)
sum(rate(rsfn_bacen_timeout_total[5m]))

# 6. Workflows travados (em execucao > 10 minutos)
temporal_workflow_started_total - temporal_workflow_completed_total - temporal_workflow_failed_total
> 0 for 10m

# 7. Cache hit rate (Redis)
sum(rate(redis_cache_hits_total[5m]))
/
(sum(rate(redis_cache_hits_total[5m])) + sum(rate(redis_cache_misses_total[5m]))) * 100

# 8. CPU usage por pod
sum(rate(container_cpu_usage_seconds_total{namespace="dict-prod"}[5m])) by (pod)

# 9. Memory usage por pod
sum(container_memory_working_set_bytes{namespace="dict-prod"}) by (pod) / 1024 / 1024 / 1024

# 10. Disk I/O
rate(container_fs_writes_bytes_total{namespace="dict-prod"}[5m])
```

---

## 3. Grafana Dashboards

### Deployment

```yaml
# grafana-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  namespace: monitoring
spec:
  replicas: 2
  selector:
    matchLabels:
      app: grafana
  template:
    metadata:
      labels:
        app: grafana
    spec:
      containers:
      - name: grafana
        image: grafana/grafana:10.2.0
        ports:
        - name: web
          containerPort: 3000
        env:
        - name: GF_SECURITY_ADMIN_PASSWORD
          valueFrom:
            secretKeyRef:
              name: grafana-secrets
              key: admin-password
        - name: GF_SERVER_ROOT_URL
          value: "https://grafana.lbpay.io"
        - name: GF_AUTH_ANONYMOUS_ENABLED
          value: "false"
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
          limits:
            cpu: 1000m
            memory: 2Gi
        volumeMounts:
        - name: storage
          mountPath: /var/lib/grafana
        - name: datasources
          mountPath: /etc/grafana/provisioning/datasources
        - name: dashboards-config
          mountPath: /etc/grafana/provisioning/dashboards
        - name: dashboards
          mountPath: /var/lib/grafana/dashboards
      volumes:
      - name: storage
        persistentVolumeClaim:
          claimName: grafana-pvc
      - name: datasources
        configMap:
          name: grafana-datasources
      - name: dashboards-config
        configMap:
          name: grafana-dashboards-config
      - name: dashboards
        configMap:
          name: grafana-dashboards
---
apiVersion: v1
kind: Service
metadata:
  name: grafana-svc
  namespace: monitoring
spec:
  type: ClusterIP
  selector:
    app: grafana
  ports:
  - name: web
    port: 3000
    targetPort: 3000
```

### Datasources Configuration

```yaml
# grafana-datasources.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-datasources
  namespace: monitoring
data:
  datasources.yaml: |
    apiVersion: 1
    datasources:
      - name: Prometheus
        type: prometheus
        access: proxy
        url: http://prometheus-svc.monitoring.svc.cluster.local:9090
        isDefault: true
        editable: false

      - name: Loki
        type: loki
        access: proxy
        url: http://loki-svc.monitoring.svc.cluster.local:3100
        editable: false

      - name: Jaeger
        type: jaeger
        access: proxy
        url: http://jaeger-query-svc.monitoring.svc.cluster.local:16686
        editable: false
```

### Dashboard: DICT Overview

```json
{
  "dashboard": {
    "title": "DICT - Overview",
    "tags": ["dict", "overview"],
    "timezone": "America/Sao_Paulo",
    "refresh": "30s",
    "panels": [
      {
        "id": 1,
        "title": "QPS (Queries per Second)",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(grpc_server_started_total{namespace=\"dict-prod\"}[5m]))",
            "legendFormat": "Total QPS"
          }
        ]
      },
      {
        "id": 2,
        "title": "Latency P95 (seconds)",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(grpc_server_handling_seconds_bucket{namespace=\"dict-prod\"}[5m]))",
            "legendFormat": "P95 Latency"
          },
          {
            "expr": "histogram_quantile(0.99, rate(grpc_server_handling_seconds_bucket{namespace=\"dict-prod\"}[5m]))",
            "legendFormat": "P99 Latency"
          }
        ],
        "alert": {
          "conditions": [
            {
              "evaluator": {
                "params": [2],
                "type": "gt"
              },
              "query": {
                "params": ["A", "5m", "now"]
              },
              "reducer": {
                "type": "avg"
              },
              "type": "query"
            }
          ]
        }
      },
      {
        "id": 3,
        "title": "Error Rate (%)",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(grpc_server_handled_total{grpc_code!=\"OK\",namespace=\"dict-prod\"}[5m])) / sum(rate(grpc_server_handled_total{namespace=\"dict-prod\"}[5m])) * 100",
            "legendFormat": "Error Rate"
          }
        ],
        "alert": {
          "conditions": [
            {
              "evaluator": {
                "params": [1],
                "type": "gt"
              }
            }
          ]
        }
      },
      {
        "id": 4,
        "title": "Active Pods",
        "type": "stat",
        "targets": [
          {
            "expr": "count(up{job=\"core-dict\"} == 1)"
          }
        ]
      },
      {
        "id": 5,
        "title": "Bacen Timeout Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(rsfn_bacen_timeout_total[5m]))",
            "legendFormat": "Timeouts/sec"
          }
        ]
      },
      {
        "id": 6,
        "title": "Workflow Success Rate",
        "type": "gauge",
        "targets": [
          {
            "expr": "sum(rate(temporal_workflow_completed_total[5m])) / sum(rate(temporal_workflow_started_total[5m])) * 100"
          }
        ],
        "thresholds": [
          {"value": 99, "color": "green"},
          {"value": 95, "color": "yellow"},
          {"value": 0, "color": "red"}
        ]
      }
    ]
  }
}
```

### Dashboard: RSFN Bridge

```json
{
  "dashboard": {
    "title": "RSFN Bridge - Bacen Integration",
    "panels": [
      {
        "id": 1,
        "title": "Bacen Requests (by operation)",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(rsfn_requests_total{status=\"success\"}[5m])) by (operation)",
            "legendFormat": "{{operation}} - Success"
          },
          {
            "expr": "sum(rate(rsfn_requests_total{status=\"error\"}[5m])) by (operation)",
            "legendFormat": "{{operation}} - Error"
          }
        ]
      },
      {
        "id": 2,
        "title": "Bacen Latency P95 (by operation)",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(rsfn_request_duration_seconds_bucket[5m])) by (operation)"
          }
        ]
      },
      {
        "id": 3,
        "title": "Circuit Breaker State",
        "type": "stat",
        "targets": [
          {
            "expr": "rsfn_circuit_breaker_state{service=\"bacen\"}"
          }
        ],
        "valueMappings": [
          {"value": 0, "text": "CLOSED"},
          {"value": 1, "text": "OPEN"},
          {"value": 2, "text": "HALF-OPEN"}
        ]
      },
      {
        "id": 4,
        "title": "XML Signing Duration",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(xml_signing_duration_seconds_bucket[5m]))"
          }
        ]
      }
    ]
  }
}
```

### Dashboard: Temporal Workflows

```json
{
  "dashboard": {
    "title": "Temporal - Workflows & Activities",
    "panels": [
      {
        "id": 1,
        "title": "Workflow Executions",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(temporal_workflow_started_total[5m])) by (workflow_type)",
            "legendFormat": "{{workflow_type}} - Started"
          },
          {
            "expr": "sum(rate(temporal_workflow_completed_total[5m])) by (workflow_type)",
            "legendFormat": "{{workflow_type}} - Completed"
          },
          {
            "expr": "sum(rate(temporal_workflow_failed_total[5m])) by (workflow_type)",
            "legendFormat": "{{workflow_type}} - Failed"
          }
        ]
      },
      {
        "id": 2,
        "title": "Stuck Workflows (running > 10 min)",
        "type": "stat",
        "targets": [
          {
            "expr": "temporal_workflow_started_total - temporal_workflow_completed_total - temporal_workflow_failed_total"
          }
        ],
        "alert": {
          "conditions": [
            {
              "evaluator": {
                "params": [10],
                "type": "gt"
              }
            }
          ]
        }
      },
      {
        "id": 3,
        "title": "Activity Execution Time",
        "type": "heatmap",
        "targets": [
          {
            "expr": "rate(temporal_activity_execution_duration_seconds_bucket[5m])"
          }
        ]
      }
    ]
  }
}
```

---

## 4. Jaeger Distributed Tracing

### Deployment

```yaml
# jaeger-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: jaeger-all-in-one
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: jaeger
  template:
    metadata:
      labels:
        app: jaeger
    spec:
      containers:
      - name: jaeger
        image: jaegertracing/all-in-one:1.51
        env:
        - name: COLLECTOR_ZIPKIN_HOST_PORT
          value: ":9411"
        - name: SPAN_STORAGE_TYPE
          value: "s3"
        - name: S3_BUCKET
          value: "lbpay-jaeger-traces-prod"
        - name: S3_REGION
          value: "us-east-1"
        - name: SPAN_RETENTION_DAYS
          value: "7"
        ports:
        - name: jaeger-ui
          containerPort: 16686
        - name: collector-grpc
          containerPort: 14250
        - name: collector-http
          containerPort: 14268
        - name: zipkin
          containerPort: 9411
        - name: agent-compact
          containerPort: 6831
          protocol: UDP
        - name: agent-binary
          containerPort: 6832
          protocol: UDP
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
          limits:
            cpu: 1000m
            memory: 2Gi
---
apiVersion: v1
kind: Service
metadata:
  name: jaeger-query-svc
  namespace: monitoring
spec:
  type: ClusterIP
  selector:
    app: jaeger
  ports:
  - name: jaeger-ui
    port: 16686
    targetPort: 16686
---
apiVersion: v1
kind: Service
metadata:
  name: jaeger-collector
  namespace: monitoring
spec:
  type: ClusterIP
  selector:
    app: jaeger
  ports:
  - name: grpc
    port: 14250
    targetPort: 14250
  - name: http
    port: 14268
    targetPort: 14268
```

### Trace Context Propagation

```go
// Exemplo: Instrumentacao Go (Core DICT)

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/trace"
)

func InitTracer() (*trace.TracerProvider, error) {
    exporter, err := jaeger.New(
        jaeger.WithCollectorEndpoint(
            jaeger.WithEndpoint("http://jaeger-collector.monitoring.svc.cluster.local:14268/api/traces"),
        ),
    )
    if err != nil {
        return nil, err
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceName("core-dict"),
            semconv.ServiceVersion("1.0.0"),
            semconv.DeploymentEnvironment("production"),
        )),
    )

    otel.SetTracerProvider(tp)
    return tp, nil
}

// Exemplo de uso em handler gRPC
func (s *Server) CreateEntry(ctx context.Context, req *pb.CreateEntryRequest) (*pb.CreateEntryResponse, error) {
    tracer := otel.Tracer("dict-service")
    ctx, span := tracer.Start(ctx, "CreateEntry")
    defer span.End()

    span.SetAttributes(
        attribute.String("entry.key", req.Key),
        attribute.String("entry.type", req.Type),
    )

    // Chamada ao repositorio (child span automatico)
    entry, err := s.repo.Create(ctx, req)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }

    // Publicar evento (child span)
    err = s.publisher.Publish(ctx, "dict.entry.created", entry)

    span.SetStatus(codes.Ok, "Entry created successfully")
    return &pb.CreateEntryResponse{Entry: entry}, nil
}
```

### Trace Example Queries

```yaml
# Jaeger UI Queries

# 1. Traces com latencia > 2s
service=core-dict duration>2s

# 2. Traces com erro
service=core-dict error=true

# 3. Traces de workflow especifico
service=connect-worker operation=CreateEntryWorkflow

# 4. Traces de chamadas ao Bacen
service=rsfn-bridge operation=SendToBacen

# 5. Traces end-to-end (multi-service)
service=connect-api operation=POST /dict/entries
```

---

## 5. Loki Log Aggregation

### Deployment

```yaml
# loki-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: loki
  namespace: monitoring
spec:
  replicas: 2
  selector:
    matchLabels:
      app: loki
  template:
    metadata:
      labels:
        app: loki
    spec:
      containers:
      - name: loki
        image: grafana/loki:2.9.0
        args:
          - -config.file=/etc/loki/loki.yaml
        ports:
        - name: http
          containerPort: 3100
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
          limits:
            cpu: 1000m
            memory: 2Gi
        volumeMounts:
        - name: config
          mountPath: /etc/loki
        - name: storage
          mountPath: /loki
      volumes:
      - name: config
        configMap:
          name: loki-config
      - name: storage
        persistentVolumeClaim:
          claimName: loki-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: loki-svc
  namespace: monitoring
spec:
  type: ClusterIP
  selector:
    app: loki
  ports:
  - name: http
    port: 3100
    targetPort: 3100
```

### Loki Configuration

```yaml
# loki-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: loki-config
  namespace: monitoring
data:
  loki.yaml: |
    auth_enabled: false

    server:
      http_listen_port: 3100

    ingester:
      lifecycler:
        ring:
          kvstore:
            store: inmemory
          replication_factor: 1
      chunk_idle_period: 5m
      chunk_retain_period: 30s

    schema_config:
      configs:
      - from: 2024-01-01
        store: boltdb-shipper
        object_store: s3
        schema: v11
        index:
          prefix: loki_index_
          period: 24h

    storage_config:
      boltdb_shipper:
        active_index_directory: /loki/index
        cache_location: /loki/cache
        shared_store: s3
      aws:
        s3: s3://us-east-1/lbpay-loki-logs-prod
        s3forcepathstyle: false

    chunk_store_config:
      max_look_back_period: 720h  # 30 days

    table_manager:
      retention_deletes_enabled: true
      retention_period: 720h  # 30 days

    limits_config:
      enforce_metric_name: false
      reject_old_samples: true
      reject_old_samples_max_age: 168h
      ingestion_rate_mb: 10
      ingestion_burst_size_mb: 20
```

### Promtail DaemonSet

```yaml
# promtail-daemonset.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: promtail
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: promtail
  template:
    metadata:
      labels:
        app: promtail
    spec:
      serviceAccountName: promtail-sa
      containers:
      - name: promtail
        image: grafana/promtail:2.9.0
        args:
          - -config.file=/etc/promtail/promtail.yaml
        ports:
        - name: http
          containerPort: 3101
        env:
        - name: HOSTNAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
        volumeMounts:
        - name: config
          mountPath: /etc/promtail
        - name: varlog
          mountPath: /var/log
          readOnly: true
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: promtail-config
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
```

### Promtail Configuration

```yaml
# promtail-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: promtail-config
  namespace: monitoring
data:
  promtail.yaml: |
    server:
      http_listen_port: 3101

    clients:
      - url: http://loki-svc.monitoring.svc.cluster.local:3100/loki/api/v1/push

    positions:
      filename: /tmp/positions.yaml

    scrape_configs:
      - job_name: kubernetes-pods
        kubernetes_sd_configs:
        - role: pod
        relabel_configs:
        - source_labels: [__meta_kubernetes_pod_node_name]
          target_label: node
        - source_labels: [__meta_kubernetes_namespace]
          target_label: namespace
        - source_labels: [__meta_kubernetes_pod_name]
          target_label: pod
        - source_labels: [__meta_kubernetes_pod_label_app]
          target_label: app
        - source_labels: [__meta_kubernetes_pod_container_name]
          target_label: container
        - replacement: /var/log/pods/*$1/*.log
          separator: /
          source_labels:
          - __meta_kubernetes_pod_uid
          - __meta_kubernetes_pod_container_name
          target_label: __path__

        pipeline_stages:
        # Parse JSON logs
        - json:
            expressions:
              level: level
              message: msg
              trace_id: trace_id
              timestamp: ts
        # Extract level
        - labels:
            level:
            trace_id:
        # Drop debug logs in production
        - match:
            selector: '{level="debug"}'
            action: drop
```

### LogQL Queries

```logql
# 1. Logs de erro (ultimos 5 minutos)
{namespace="dict-prod", level="error"}

# 2. Logs do Core DICT com trace_id
{namespace="dict-prod", app="core-dict"} |= "trace_id" | json

# 3. Logs de timeout do Bacen
{namespace="dict-prod", app="rsfn-bridge"} |~ "timeout|TIMEOUT"

# 4. Logs de workflow falhado
{namespace="dict-prod", app="connect-worker"} |~ "workflow.*failed"

# 5. Top 10 erros por mensagem
topk(10, sum by (message) (rate({namespace="dict-prod", level="error"}[5m])))

# 6. Correlacao trace + logs
{namespace="dict-prod"} | json | trace_id="abc123..."

# 7. Latencia de queries (parsed)
{namespace="dict-prod", app="core-dict"}
  | json
  | latency > 2000
  | line_format "{{.message}} - latency={{.latency}}ms"
```

---

## 6. Alerting Strategy

### AlertManager Deployment

```yaml
# alertmanager-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 2
  selector:
    matchLabels:
      app: alertmanager
  template:
    metadata:
      labels:
        app: alertmanager
    spec:
      containers:
      - name: alertmanager
        image: prom/alertmanager:v0.26.0
        args:
          - '--config.file=/etc/alertmanager/alertmanager.yml'
          - '--storage.path=/alertmanager'
        ports:
        - name: web
          containerPort: 9093
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
        volumeMounts:
        - name: config
          mountPath: /etc/alertmanager
        - name: storage
          mountPath: /alertmanager
      volumes:
      - name: config
        configMap:
          name: alertmanager-config
      - name: storage
        emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  name: alertmanager-svc
  namespace: monitoring
spec:
  type: ClusterIP
  selector:
    app: alertmanager
  ports:
  - name: web
    port: 9093
    targetPort: 9093
```

### AlertManager Configuration

```yaml
# alertmanager-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: alertmanager-config
  namespace: monitoring
data:
  alertmanager.yml: |
    global:
      resolve_timeout: 5m
      slack_api_url: 'https://hooks.slack.com/services/XXX/YYY/ZZZ'

    route:
      group_by: ['alertname', 'cluster', 'service']
      group_wait: 10s
      group_interval: 10s
      repeat_interval: 12h
      receiver: 'slack-default'
      routes:
      - match:
          severity: critical
        receiver: 'pagerduty-critical'
        continue: true
      - match:
          severity: warning
        receiver: 'slack-warnings'

    receivers:
    - name: 'slack-default'
      slack_configs:
      - channel: '#dict-alerts'
        title: 'DICT Alert'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
        send_resolved: true

    - name: 'slack-warnings'
      slack_configs:
      - channel: '#dict-warnings'
        title: 'DICT Warning'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'

    - name: 'pagerduty-critical'
      pagerduty_configs:
      - service_key: 'YOUR_PAGERDUTY_SERVICE_KEY'
        description: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'

    inhibit_rules:
    - source_match:
        severity: 'critical'
      target_match:
        severity: 'warning'
      equal: ['alertname', 'cluster', 'service']
```

### Alert Rules

```yaml
# prometheus-rules.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-rules
  namespace: monitoring
data:
  alerts.yml: |
    groups:
    - name: dict_alerts
      interval: 30s
      rules:

      # HIGH LATENCY ALERT
      - alert: HighLatency
        expr: |
          histogram_quantile(0.95,
            rate(grpc_server_handling_seconds_bucket{namespace="dict-prod"}[5m])
          ) > 2
        for: 5m
        labels:
          severity: warning
          service: dict
        annotations:
          summary: "High latency detected"
          description: "P95 latency is {{ $value }}s (threshold: 2s) for {{ $labels.grpc_method }}"

      # CRITICAL LATENCY ALERT
      - alert: CriticalLatency
        expr: |
          histogram_quantile(0.95,
            rate(grpc_server_handling_seconds_bucket{namespace="dict-prod"}[5m])
          ) > 5
        for: 2m
        labels:
          severity: critical
          service: dict
        annotations:
          summary: "CRITICAL: Very high latency"
          description: "P95 latency is {{ $value }}s (threshold: 5s)"

      # HIGH ERROR RATE ALERT
      - alert: HighErrorRate
        expr: |
          sum(rate(grpc_server_handled_total{grpc_code!="OK",namespace="dict-prod"}[5m]))
          /
          sum(rate(grpc_server_handled_total{namespace="dict-prod"}[5m])) * 100 > 1
        for: 5m
        labels:
          severity: critical
          service: dict
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }}% (threshold: 1%)"

      # BACEN TIMEOUT ALERT
      - alert: BacenTimeout
        expr: |
          sum(rate(rsfn_bacen_timeout_total[5m])) > 0.1
        for: 2m
        labels:
          severity: critical
          service: rsfn-bridge
        annotations:
          summary: "Bacen timeouts detected"
          description: "Timeout rate: {{ $value }} timeouts/sec"

      # WORKFLOW STUCK ALERT
      - alert: WorkflowStuck
        expr: |
          (temporal_workflow_started_total - temporal_workflow_completed_total - temporal_workflow_failed_total) > 10
        for: 10m
        labels:
          severity: warning
          service: temporal
        annotations:
          summary: "Workflows stuck in running state"
          description: "{{ $value }} workflows running for >10 minutes"

      # POD DOWN ALERT
      - alert: PodDown
        expr: |
          up{job="core-dict"} == 0
        for: 1m
        labels:
          severity: critical
          service: dict
        annotations:
          summary: "Pod is down"
          description: "Pod {{ $labels.pod }} is down"

      # HIGH CPU ALERT
      - alert: HighCPUUsage
        expr: |
          sum(rate(container_cpu_usage_seconds_total{namespace="dict-prod"}[5m])) by (pod)
          /
          sum(container_spec_cpu_quota{namespace="dict-prod"}/container_spec_cpu_period{namespace="dict-prod"}) by (pod) * 100 > 80
        for: 5m
        labels:
          severity: warning
          service: dict
        annotations:
          summary: "High CPU usage"
          description: "CPU usage is {{ $value }}% for pod {{ $labels.pod }}"

      # HIGH MEMORY ALERT
      - alert: HighMemoryUsage
        expr: |
          sum(container_memory_working_set_bytes{namespace="dict-prod"}) by (pod)
          /
          sum(container_spec_memory_limit_bytes{namespace="dict-prod"}) by (pod) * 100 > 80
        for: 5m
        labels:
          severity: warning
          service: dict
        annotations:
          summary: "High memory usage"
          description: "Memory usage is {{ $value }}% for pod {{ $labels.pod }}"

      # CIRCUIT BREAKER OPEN
      - alert: CircuitBreakerOpen
        expr: |
          rsfn_circuit_breaker_state{service="bacen"} == 1
        for: 1m
        labels:
          severity: critical
          service: rsfn-bridge
        annotations:
          summary: "Circuit breaker OPEN"
          description: "Circuit breaker for Bacen is OPEN - all requests are failing"

      # LOW AVAILABILITY
      - alert: LowAvailability
        expr: |
          avg(up{job="core-dict"}) * 100 < 99.9
        for: 5m
        labels:
          severity: critical
          service: dict
        annotations:
          summary: "SLO violation: Low availability"
          description: "Availability is {{ $value }}% (SLO: 99.9%)"
```

---

## 7. SLOs & SLIs

### Service Level Indicators (SLIs)

```yaml
# SLIs Definition

# 1. Availability SLI
sli_availability:
  query: avg(up{job="core-dict"}) * 100
  target: 99.9
  window: 30d

# 2. Latency SLI (P95)
sli_latency_p95:
  query: histogram_quantile(0.95, rate(grpc_server_handling_seconds_bucket{namespace="dict-prod"}[5m]))
  target: 2  # seconds
  window: 30d

# 3. Error Rate SLI
sli_error_rate:
  query: |
    sum(rate(grpc_server_handled_total{grpc_code!="OK",namespace="dict-prod"}[5m]))
    /
    sum(rate(grpc_server_handled_total{namespace="dict-prod"}[5m])) * 100
  target: 1  # percent
  window: 30d

# 4. Bacen Timeout SLI
sli_bacen_timeout:
  query: |
    sum(rate(rsfn_bacen_timeout_total[5m]))
    /
    sum(rate(rsfn_requests_total[5m])) * 100
  target: 0.1  # percent
  window: 30d

# 5. Workflow Success SLI
sli_workflow_success:
  query: |
    sum(rate(temporal_workflow_completed_total[5m]))
    /
    sum(rate(temporal_workflow_started_total[5m])) * 100
  target: 99  # percent
  window: 30d
```

### Error Budget Calculation

```yaml
# Error Budget (mensal)

Availability SLO: 99.9%
Error Budget: 0.1%
Monthly Downtime Allowed: 43.2 minutes

Calculation:
  Total minutes in month: 43200 (30 days)
  Allowed downtime: 43200 * 0.001 = 43.2 minutes

Monitoring:
  current_uptime = avg_over_time(up{job="core-dict"}[30d]) * 100
  error_budget_remaining = (99.9 - current_uptime) / 0.1 * 100

  # Example: If current_uptime = 99.95%
  # error_budget_remaining = (99.9 - 99.95) / 0.1 * 100 = -50%
  # (exceeded error budget by 50%)
```

---

## 8. Incident Response

### Incident Runbooks

```yaml
# Runbook: High Latency

Title: High Latency (P95 > 2s)
Severity: Warning
Team: DevOps, Backend

Steps:
  1. Check Grafana dashboard: DICT Overview
  2. Identify which operation has high latency
  3. Check Jaeger traces for slow operations
  4. Investigate:
     - Database slow queries (pg_stat_statements)
     - Redis cache miss rate
     - Bacen API latency
     - Temporal workflow stuck
  5. Mitigation:
     - Scale up pods (HPA should auto-scale)
     - Clear cache if stale data
     - Contact Bacen if external issue
  6. Post-mortem: Document root cause

---

# Runbook: Bacen Timeout

Title: Bacen Timeout Rate > 0.1%
Severity: Critical
Team: DevOps, Integration

Steps:
  1. Check circuit breaker state (Grafana dashboard)
  2. Verify Bacen API status: https://status.bcb.gov.br
  3. Check mTLS certificates expiration
  4. Review recent Bacen API changes (changelog)
  5. Investigate:
     - Network connectivity to Bacen
     - XML signing issues
     - SOAP envelope format
  6. Mitigation:
     - If Bacen is down: Wait for recovery (circuit breaker handles this)
     - If certificate issue: Renew certificate
     - If network issue: Check AWS security groups, NAT gateway
  7. Escalate to Bacen support if needed

---

# Runbook: Workflow Stuck

Title: Workflows stuck (>10 running for >10 min)
Severity: Warning
Team: DevOps, Backend

Steps:
  1. Check Temporal UI for stuck workflows
  2. Identify workflow type and activity
  3. Check Jaeger traces for the workflow execution
  4. Investigate:
     - Activity timeout configuration
     - Worker capacity (are workers running?)
     - Database deadlock
     - External API dependency (Bacen)
  5. Mitigation:
     - Manually complete/cancel stuck workflows (Temporal UI)
     - Scale up workers
     - Fix activity code bug
  6. Prevent:
     - Review timeout configuration
     - Add retry policy
     - Add idempotency checks
```

### On-Call Rotation

```yaml
# PagerDuty Integration

Schedule:
  Team: LBPay DICT DevOps
  Rotation: Weekly

Escalation Policy:
  Level 1: On-call engineer (5 minutes)
  Level 2: Tech lead (10 minutes)
  Level 3: Manager (15 minutes)

Critical Alerts (PagerDuty):
  - HighErrorRate
  - BacenTimeout
  - PodDown
  - CircuitBreakerOpen
  - LowAvailability

Warning Alerts (Slack only):
  - HighLatency
  - HighCPUUsage
  - HighMemoryUsage
  - WorkflowStuck
```

---

## 9. Rastreabilidade

### Documentos Relacionados

| ID | Documento | Relacao |
|----|-----------|---------|
| **DEV-001** | [CI/CD Pipeline Core](./Pipelines/DEV-001_CI_CD_Pipeline_Core.md) | Pipeline do Core DICT |
| **DEV-004** | [Kubernetes Manifests](./DEV-004_Kubernetes_Manifests.md) | Manifests K8s com ServiceMonitors |
| **TEC-001** | [Core DICT Specification](../11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md) | Especificacao do Core |
| **SEC-001** | [mTLS Configuration](../13_Seguranca/SEC-001_mTLS_Configuration.md) | Configuracao de seguranca |

### Metricas de Sucesso

```yaml
Observability Stack:
  - Metrics coverage: 100% dos servicos
  - Traces sampling: 100% em producao
  - Logs retention: 30 dias
  - Alert response time: < 5 minutos (P95)

SLO Compliance:
  - Availability: >= 99.9%
  - Latency P95: < 2s
  - Error rate: < 1%
  - Bacen timeout: < 0.1%
  - Workflow success: > 99%

Incident Response:
  - MTTD (Mean Time To Detect): < 2 minutos
  - MTTR (Mean Time To Resolve): < 30 minutos (P95)
  - Post-mortem: 100% dos incidentes criticos
```

---

**Ultima Atualizacao**: 2025-10-25
**Versao**: 1.0
**Status**: Completo
