# TSP-001: Temporal Workflow Engine - Technical Specification

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: Temporal Workflow Engine
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: BACKEND (AI Agent - Backend Specialist)
**Revisor**: [Aguardando]
**Aprovador**: Tech Lead, Head de Arquitetura

---

## Sumário Executivo

Este documento especifica a implementação do **Temporal Workflow Engine** (v1.36.0) para o projeto DICT LBPay, cobrindo deployment em Kubernetes, configuração de workflows de longa duração (ClaimWorkflow de 30 dias), activities, retry policies, e estratégias de monitoramento.

**Baseado em**:
- [TEC-003: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [ANA-003: Análise Repositório Connect](../../00_Analises/ANA-003_Analise_Repo_Connect.md)
- [ADR-004: Temporal Workflows](../ADR-004_Temporal_Workflows.md) (pendente)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | BACKEND | Versão inicial - Temporal v1.36.0 specification |

---

## Índice

1. [Visão Geral](#1-visão-geral)
2. [Deployment Kubernetes](#2-deployment-kubernetes)
3. [Configuration](#3-configuration)
4. [Workflows](#4-workflows)
5. [Activities](#5-activities)
6. [Retry Policies](#6-retry-policies)
7. [Monitoring & Observability](#7-monitoring--observability)
8. [High Availability](#8-high-availability)
9. [Disaster Recovery](#9-disaster-recovery)

---

## 1. Visão Geral

### 1.1. Temporal Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Temporal Cluster (v1.36.0)                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Frontend Service (gRPC Gateway)                            │ │
│  │  - Port: 7233 (gRPC)                                       │ │
│  │  - Port: 7243 (Metrics)                                    │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           ↓                                       │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  History Service (Workflow Execution)                       │ │
│  │  - StatefulSet (3 replicas)                                │ │
│  │  - Manages workflow state                                  │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           ↓                                       │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Matching Service (Task Routing)                           │ │
│  │  - Routes tasks to workers                                 │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           ↓                                       │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Worker Service (Task Execution)                           │ │
│  │  - Executes workflow logic                                 │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│               PostgreSQL (Temporal Persistence)                  │
│  - Schema: temporal                                              │
│  - Tables: executions, tasks, histories                          │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│                  DICT Orchestration Workers                      │
│  - ClaimWorkflow (30 days)                                       │
│  - MonitorStatusWorkflow                                         │
│  - ExpireCompletionPeriodWorkflow                                │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2. Key Features

| Feature | Value | Justification |
|---------|-------|---------------|
| **Version** | v1.36.0 | Latest stable (as of 2025-10-25) |
| **Deployment** | Kubernetes StatefulSet | High availability, persistence |
| **Persistence** | PostgreSQL 16 | ACID compliance, durability |
| **Workers** | Go SDK v1.36.0 | Native Go integration |
| **Task Queue** | `dict-task-queue` | Centralized queue for all DICT workflows |
| **Namespace** | `dict` | Logical isolation |
| **Max Workflow Duration** | 30 days (Claims) | Business requirement |
| **HA** | 3 History replicas | 99.9% availability |

---

## 2. Deployment Kubernetes

### 2.1. Namespace

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: temporal
  labels:
    name: temporal
    environment: production
```

### 2.2. Temporal Server StatefulSet

```yaml
# k8s/temporal-server.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: temporal-server
  namespace: temporal
spec:
  serviceName: temporal-headless
  replicas: 3
  selector:
    matchLabels:
      app: temporal-server
  template:
    metadata:
      labels:
        app: temporal-server
    spec:
      containers:
      - name: temporal-server
        image: temporalio/server:1.36.0
        ports:
        - name: grpc
          containerPort: 7233
          protocol: TCP
        - name: metrics
          containerPort: 7243
          protocol: TCP
        env:
        - name: TEMPORAL_CLI_ADDRESS
          value: "temporal-frontend:7233"
        - name: DB
          value: "postgresql"
        - name: DB_PORT
          value: "5432"
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: temporal-secrets
              key: postgres-user
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: temporal-secrets
              key: postgres-password
        - name: POSTGRES_SEEDS
          value: "postgres:5432"
        - name: DYNAMIC_CONFIG_FILE_PATH
          value: "/etc/temporal/config/dynamicconfig.yaml"
        - name: SERVICES
          value: "history,matching,worker,frontend"
        volumeMounts:
        - name: config
          mountPath: /etc/temporal/config
        - name: data
          mountPath: /var/lib/temporal
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
        livenessProbe:
          tcpSocket:
            port: 7233
          initialDelaySeconds: 60
          periodSeconds: 10
        readinessProbe:
          tcpSocket:
            port: 7233
          initialDelaySeconds: 30
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: temporal-config
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 10Gi
```

### 2.3. Temporal Frontend Service

```yaml
# k8s/temporal-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: temporal-frontend
  namespace: temporal
spec:
  type: ClusterIP
  ports:
  - name: grpc
    port: 7233
    targetPort: 7233
    protocol: TCP
  - name: metrics
    port: 7243
    targetPort: 7243
    protocol: TCP
  selector:
    app: temporal-server
---
apiVersion: v1
kind: Service
metadata:
  name: temporal-headless
  namespace: temporal
spec:
  type: ClusterIP
  clusterIP: None
  ports:
  - name: grpc
    port: 7233
    targetPort: 7233
  selector:
    app: temporal-server
```

### 2.4. Temporal ConfigMap

```yaml
# k8s/temporal-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: temporal-config
  namespace: temporal
data:
  dynamicconfig.yaml: |
    # Retention
    system.enableNamespaceNotActiveAutoForwarding:
      - value: true

    # Workflow execution limits
    limit.maxIDLength:
      - value: 255
    limit.maxWorkflowIDSize:
      - value: 1000
    limit.workflowExecutionRate:
      - value: 5000  # 5000 workflows/sec

    # History service
    history.maxPageSize:
      - value: 1000
    history.defaultActivityRetryPolicy:
      - value:
          InitialIntervalInSeconds: 1
          MaximumIntervalInSeconds: 30
          BackoffCoefficient: 2.0
          MaximumAttempts: 3

    # Matching service
    matching.numTaskqueueWritePartitions:
      - value: 10
    matching.numTaskqueueReadPartitions:
      - value: 10

    # Frontend service
    frontend.rps:
      - value: 10000
    frontend.globalRPS:
      - value: 50000
```

### 2.5. PostgreSQL for Temporal

```yaml
# k8s/temporal-postgres.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres-temporal
  namespace: temporal
spec:
  serviceName: postgres
  replicas: 1
  selector:
    matchLabels:
      app: postgres-temporal
  template:
    metadata:
      labels:
        app: postgres-temporal
    spec:
      containers:
      - name: postgres
        image: postgres:16
        ports:
        - containerPort: 5432
          name: postgres
        env:
        - name: POSTGRES_DB
          value: "temporal"
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: temporal-secrets
              key: postgres-user
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: temporal-secrets
              key: postgres-password
        - name: PGDATA
          value: /var/lib/postgresql/data/pgdata
        volumeMounts:
        - name: postgres-data
          mountPath: /var/lib/postgresql/data
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
  volumeClaimTemplates:
  - metadata:
      name: postgres-data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 50Gi
---
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: temporal
spec:
  type: ClusterIP
  ports:
  - port: 5432
    targetPort: 5432
  selector:
    app: postgres-temporal
```

### 2.6. Secrets

```yaml
# k8s/temporal-secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: temporal-secrets
  namespace: temporal
type: Opaque
stringData:
  postgres-user: temporal
  postgres-password: <REPLACE_WITH_SECURE_PASSWORD>
```

---

## 3. Configuration

### 3.1. Temporal Client Configuration

```go
// internal/infrastructure/temporal/client.go
package temporal

import (
	"go.temporal.io/sdk/client"
	"log"
)

type Config struct {
	HostPort  string
	Namespace string
}

func NewClient(cfg Config) (client.Client, error) {
	c, err := client.NewClient(client.Options{
		HostPort:  cfg.HostPort,
		Namespace: cfg.Namespace,
	})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
		return nil, err
	}
	return c, nil
}
```

### 3.2. Worker Configuration

```go
// apps/orchestration-worker/cmd/worker/main.go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"github.com/lb-conn/rsfn-connect/apps/orchestration-worker/workflows/claims"
	"github.com/lb-conn/rsfn-connect/apps/orchestration-worker/activities/claims"
)

func main() {
	// Create Temporal client
	c, err := client.NewClient(client.Options{
		HostPort:  getEnv("TEMPORAL_HOST", "temporal-frontend:7233"),
		Namespace: getEnv("TEMPORAL_NAMESPACE", "dict"),
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	// Create worker
	w := worker.New(c, "dict-task-queue", worker.Options{
		MaxConcurrentActivityExecutionSize:     100,
		MaxConcurrentWorkflowTaskExecutionSize: 50,
	})

	// Register Workflows
	w.RegisterWorkflow(claims.CreateClaimWorkflow)
	w.RegisterWorkflow(claims.MonitorStatusWorkflow)
	w.RegisterWorkflow(claims.ExpireCompletionPeriodWorkflow)
	w.RegisterWorkflow(claims.CompleteClaimWorkflow)
	w.RegisterWorkflow(claims.CancelClaimWorkflow)

	// Register Activities
	w.RegisterActivity(claims.CreateClaimGRPCActivity)
	w.RegisterActivity(claims.CompleteClaimGRPCActivity)
	w.RegisterActivity(claims.CancelClaimGRPCActivity)
	w.RegisterActivity(claims.GetClaimGRPCActivity)

	// Start worker
	go func() {
		err = w.Run(worker.InterruptCh())
		if err != nil {
			log.Fatalln("Unable to start Worker", err)
		}
	}()

	log.Println("Temporal Worker started successfully")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Worker stopping...")
	w.Stop()
	log.Println("Worker stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
```

---

## 4. Workflows

### 4.1. ClaimWorkflow (30 days)

**File**: `apps/orchestration-worker/workflows/claims/create_workflow.go`

**Description**: Orchestrates a 30-day claim resolution period.

**Implementation**:

```go
package claims

import (
	"time"
	"go.temporal.io/sdk/workflow"
)

type CreateClaimWorkflowInput struct {
	ClaimID     string
	EntryKey    string
	ClaimerISPB string
	OwnerISPB   string
}

func CreateClaimWorkflow(ctx workflow.Context, input CreateClaimWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("CreateClaimWorkflow started", "claimID", input.ClaimID)

	// Activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
			NonRetriableErrorTypes: []string{
				"ValidationError",
				"KeyAlreadyExistsError",
			},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Activity 1: Create claim in Bacen via Bridge
	var claimResponse CreateClaimResponse
	err := workflow.ExecuteActivity(ctx, CreateClaimGRPCActivity, input).Get(ctx, &claimResponse)
	if err != nil {
		logger.Error("Failed to create claim in Bacen", "error", err)
		return err
	}

	logger.Info("Claim created in Bacen", "externalID", claimResponse.ExternalID)

	// Wait for signal or 30-day timeout
	signalChannel := workflow.GetSignalChannel(ctx, "claim-decision")

	selector := workflow.NewSelector(ctx)

	var confirmed bool
	selector.AddReceive(signalChannel, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &confirmed)
		logger.Info("Claim decision received", "confirmed", confirmed)
	})

	// 30-day timer (2592000 seconds)
	timer := workflow.NewTimer(ctx, 30*24*time.Hour)
	selector.AddFuture(timer, func(f workflow.Future) {
		logger.Warn("30-day timeout reached - auto-cancelling claim")
		confirmed = false
	})

	// Wait for event
	selector.Select(ctx)

	// Activity 2: Confirm or Cancel claim
	if confirmed {
		err = workflow.ExecuteActivity(ctx, CompleteClaimGRPCActivity, input.ClaimID).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to complete claim", "error", err)
			return err
		}
		logger.Info("Claim completed successfully", "claimID", input.ClaimID)

		// Activity 3: Transfer entry ownership
		err = workflow.ExecuteActivity(ctx, TransferEntryActivity, input).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to transfer entry", "error", err)
			return err
		}
	} else {
		err = workflow.ExecuteActivity(ctx, CancelClaimGRPCActivity, input.ClaimID).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to cancel claim", "error", err)
			return err
		}
		logger.Info("Claim cancelled", "claimID", input.ClaimID)
	}

	// Activity 4: Notify users
	err = workflow.ExecuteActivity(ctx, NotifyUsersActivity, input, confirmed).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to notify users", "error", err)
		// Don't fail workflow for notification errors
	}

	logger.Info("CreateClaimWorkflow completed", "claimID", input.ClaimID)
	return nil
}
```

### 4.2. MonitorStatusWorkflow

**File**: `apps/orchestration-worker/workflows/claims/monitor_status_workflow.go`

**Description**: Periodically checks claim status in Bacen.

```go
package claims

import (
	"time"
	"go.temporal.io/sdk/workflow"
)

type MonitorStatusWorkflowInput struct {
	ClaimID string
}

func MonitorStatusWorkflow(ctx workflow.Context, input MonitorStatusWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("MonitorStatusWorkflow started", "claimID", input.ClaimID)

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 15 * time.Second,
		RetryPolicy: &workflow.RetryPolicy{
			MaximumAttempts: 2,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Monitor every 6 hours for 30 days
	for i := 0; i < 120; i++ { // 30 days * 4 checks/day = 120
		// Check status in Bacen
		var status ClaimStatus
		err := workflow.ExecuteActivity(ctx, GetClaimGRPCActivity, input.ClaimID).Get(ctx, &status)
		if err != nil {
			logger.Error("Failed to get claim status", "error", err)
			// Continue monitoring despite errors
		} else {
			logger.Info("Claim status checked", "status", status)

			// If claim is resolved, stop monitoring
			if status == "COMPLETED" || status == "CANCELLED" || status == "EXPIRED" {
				logger.Info("Claim resolved - stopping monitoring", "status", status)
				return nil
			}
		}

		// Wait 6 hours before next check
		err = workflow.Sleep(ctx, 6*time.Hour)
		if err != nil {
			return err
		}
	}

	logger.Info("MonitorStatusWorkflow completed after 30 days", "claimID", input.ClaimID)
	return nil
}
```

### 4.3. ExpireCompletionPeriodWorkflow

**File**: `apps/orchestration-worker/workflows/claims/expire_completion_period_workflow.go`

**Description**: Auto-cancels claim after 30 days if not resolved.

```go
package claims

import (
	"time"
	"go.temporal.io/sdk/workflow"
)

type ExpireCompletionPeriodWorkflowInput struct {
	ClaimID   string
	ExpiresAt time.Time
}

func ExpireCompletionPeriodWorkflow(ctx workflow.Context, input ExpireCompletionPeriodWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("ExpireCompletionPeriodWorkflow started", "claimID", input.ClaimID, "expiresAt", input.ExpiresAt)

	// Calculate time until expiration
	now := workflow.Now(ctx)
	timeUntilExpiration := input.ExpiresAt.Sub(now)

	if timeUntilExpiration <= 0 {
		logger.Warn("Claim already expired", "claimID", input.ClaimID)
		return nil
	}

	// Wait until expiration
	err := workflow.Sleep(ctx, timeUntilExpiration)
	if err != nil {
		return err
	}

	// Activity: Cancel claim
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &workflow.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	err = workflow.ExecuteActivity(ctx, CancelClaimGRPCActivity, input.ClaimID).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to auto-cancel expired claim", "error", err)
		return err
	}

	logger.Info("Claim auto-cancelled due to 30-day expiration", "claimID", input.ClaimID)
	return nil
}
```

---

## 5. Activities

### 5.1. CreateClaimGRPCActivity

**File**: `apps/orchestration-worker/activities/claims/create_activity.go`

```go
package claims

import (
	"context"
	"go.temporal.io/sdk/activity"
	pb "github.com/lb-conn/rsfn-connect/api/proto"
)

type CreateClaimGRPCActivity struct {
	bridgeClient pb.BridgeServiceClient
}

func NewCreateClaimGRPCActivity(bridgeClient pb.BridgeServiceClient) *CreateClaimGRPCActivity {
	return &CreateClaimGRPCActivity{
		bridgeClient: bridgeClient,
	}
}

func (a *CreateClaimGRPCActivity) Execute(ctx context.Context, input CreateClaimWorkflowInput) (*CreateClaimResponse, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("CreateClaimGRPCActivity started", "claimID", input.ClaimID)

	req := &pb.CreateClaimRequest{
		ClaimId:     input.ClaimID,
		EntryKey:    input.EntryKey,
		ClaimerIspb: input.ClaimerISPB,
		OwnerIspb:   input.OwnerISPB,
	}

	resp, err := a.bridgeClient.CreateClaim(ctx, req)
	if err != nil {
		logger.Error("Failed to call Bridge.CreateClaim", "error", err)
		return nil, err
	}

	logger.Info("Claim created in Bacen via Bridge", "externalID", resp.ExternalId)

	return &CreateClaimResponse{
		ExternalID: resp.ExternalId,
		Status:     resp.Status,
		ExpiresAt:  resp.ExpiresAt.AsTime(),
	}, nil
}
```

### 5.2. Activity Retry Configuration

All activities use this default retry policy:

```go
activityOptions := workflow.ActivityOptions{
	StartToCloseTimeout: 30 * time.Second,
	RetryPolicy: &workflow.RetryPolicy{
		InitialInterval:    1 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    30 * time.Second,
		MaximumAttempts:    3,
		NonRetriableErrorTypes: []string{
			"ValidationError",
			"KeyAlreadyExistsError",
			"NotFoundError",
		},
	},
}
```

---

## 6. Retry Policies

### 6.1. Workflow Retry Policy

```go
workflowOptions := client.StartWorkflowOptions{
	ID:        "claim-" + claimID,
	TaskQueue: "dict-task-queue",
	RetryPolicy: &temporal.RetryPolicy{
		InitialInterval:    5 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    5 * time.Minute,
		MaximumAttempts:    5,
	},
	WorkflowExecutionTimeout: 31 * 24 * time.Hour, // 31 days (30 + buffer)
}
```

### 6.2. Activity Retry Matrix

| Activity | InitialInterval | MaxInterval | MaxAttempts | NonRetriable Errors |
|----------|----------------|-------------|-------------|---------------------|
| CreateClaimGRPCActivity | 1s | 30s | 3 | ValidationError, KeyAlreadyExistsError |
| CompleteClaimGRPCActivity | 1s | 30s | 3 | NotFoundError, AlreadyResolvedError |
| CancelClaimGRPCActivity | 1s | 30s | 3 | NotFoundError |
| GetClaimGRPCActivity | 1s | 15s | 2 | None (always retry) |
| NotifyUsersActivity | 2s | 60s | 5 | None (best effort) |

### 6.3. Backoff Strategy

**Exponential Backoff**:
- Attempt 1: 1s
- Attempt 2: 2s (1s * 2.0)
- Attempt 3: 4s (2s * 2.0)
- Attempt 4: 8s (4s * 2.0)
- ...
- Max: 30s

---

## 7. Monitoring & Observability

### 7.1. Prometheus Metrics

**Exposed Metrics** (Port 7243):

```yaml
# Workflow metrics
temporal_workflow_started_total
temporal_workflow_completed_total
temporal_workflow_failed_total
temporal_workflow_timeout_total
temporal_workflow_duration_seconds

# Activity metrics
temporal_activity_started_total
temporal_activity_completed_total
temporal_activity_failed_total
temporal_activity_retry_total

# Task queue metrics
temporal_taskqueue_depth
temporal_taskqueue_backlog_duration_seconds

# History service metrics
temporal_history_execution_latency_seconds
temporal_history_shard_count
```

### 7.2. ServiceMonitor (Prometheus Operator)

```yaml
# k8s/temporal-servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: temporal-server
  namespace: temporal
spec:
  selector:
    matchLabels:
      app: temporal-server
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
```

### 7.3. Grafana Dashboard

**Key Panels**:
- Workflow Execution Rate (per minute)
- Workflow Success/Failure Ratio
- Activity Retry Rate
- Task Queue Depth
- Average Workflow Duration
- ClaimWorkflow 30-day expiration countdown

**Grafana Dashboard JSON**: (pendente criação em TSP-003)

### 7.4. Alerting Rules

```yaml
# prometheus/temporal-alerts.yaml
groups:
  - name: temporal
    interval: 30s
    rules:
      - alert: TemporalWorkflowFailureRateHigh
        expr: rate(temporal_workflow_failed_total[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High workflow failure rate"
          description: "Workflow failure rate is {{ $value }} (> 0.1/s)"

      - alert: TemporalTaskQueueBacklogHigh
        expr: temporal_taskqueue_depth > 1000
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Task queue backlog is high"
          description: "Task queue depth is {{ $value }} tasks"

      - alert: TemporalHistoryServiceDown
        expr: up{job="temporal-server"} == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Temporal History Service is down"
          description: "History service has been down for 2 minutes"

      - alert: ClaimWorkflowNearing30DayExpiration
        expr: temporal_workflow_duration_seconds{workflow_type="CreateClaimWorkflow"} > 2419200  # 28 days
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "Claim workflow nearing 30-day expiration"
          description: "Claim {{ $labels.workflow_id }} has been running for {{ $value }} seconds"
```

### 7.5. Temporal Web UI

**Deployment**:

```yaml
# k8s/temporal-web.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: temporal-web
  namespace: temporal
spec:
  replicas: 2
  selector:
    matchLabels:
      app: temporal-web
  template:
    metadata:
      labels:
        app: temporal-web
    spec:
      containers:
      - name: temporal-web
        image: temporalio/ui:2.21.3
        ports:
        - containerPort: 8080
        env:
        - name: TEMPORAL_ADDRESS
          value: "temporal-frontend:7233"
        - name: TEMPORAL_CORS_ORIGINS
          value: "*"
---
apiVersion: v1
kind: Service
metadata:
  name: temporal-web
  namespace: temporal
spec:
  type: ClusterIP
  ports:
  - port: 8080
    targetPort: 8080
  selector:
    app: temporal-web
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: temporal-web
  namespace: temporal
  annotations:
    kubernetes.io/ingress.class: "nginx"
spec:
  rules:
  - host: temporal-ui.lbpay.com.br
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: temporal-web
            port:
              number: 8080
```

---

## 8. High Availability

### 8.1. Replication

**History Service**: 3 replicas (StatefulSet)

**PostgreSQL**: Master-Slave replication (future: use managed PostgreSQL)

### 8.2. Disaster Recovery

**Backup Strategy**:

```yaml
# k8s/temporal-backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: temporal-postgres-backup
  namespace: temporal
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:16
            command:
            - /bin/sh
            - -c
            - |
              pg_dump -h postgres -U temporal temporal > /backups/temporal_$(date +%Y%m%d_%H%M%S).sql
              aws s3 cp /backups/temporal_$(date +%Y%m%d_%H%M%S).sql s3://lbpay-backups/temporal/
            env:
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: temporal-secrets
                  key: postgres-password
            volumeMounts:
            - name: backups
              mountPath: /backups
          volumes:
          - name: backups
            emptyDir: {}
          restartPolicy: OnFailure
```

### 8.3. Scaling Strategy

**Horizontal Pod Autoscaler**:

```yaml
# k8s/temporal-hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: temporal-server-hpa
  namespace: temporal
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: StatefulSet
    name: temporal-server
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

---

## 9. Disaster Recovery

### 9.1. Backup Retention

| Backup Type | Frequency | Retention |
|-------------|-----------|-----------|
| PostgreSQL Full Dump | Daily | 30 days |
| PostgreSQL WAL Archives | Continuous | 7 days |
| Workflow History Snapshots | Weekly | 90 days |

### 9.2. Recovery Procedures

**Full Recovery** (RTO: 2 hours):

1. Restore PostgreSQL from S3 backup
2. Apply WAL archives
3. Redeploy Temporal StatefulSet
4. Verify workflow continuity

**Point-in-Time Recovery** (RTO: 4 hours):

1. Restore PostgreSQL to specific timestamp
2. Replay workflows from last checkpoint
3. Validate data consistency

---

## Rastreabilidade

### Requisitos Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RF-TSP-001 | Orquestrar ClaimWorkflow (30 dias) | [TEC-003](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md) | ✅ Especificado |
| RF-TSP-002 | Monitorar status de claims periodicamente | [TEC-003](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md) | ✅ Especificado |
| RF-TSP-003 | Auto-cancelar claims após 30 dias | [TEC-003](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md) | ✅ Especificado |

### Requisitos Não-Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RNF-TSP-001 | HA: 99.9% availability | SLA requirement | ✅ Especificado |
| RNF-TSP-002 | Retry with exponential backoff | Best Practices | ✅ Especificado |
| RNF-TSP-003 | Prometheus metrics exposition | Observability | ✅ Especificado |
| RNF-TSP-004 | PostgreSQL persistence (ACID) | Durability | ✅ Especificado |

---

## Próximas Revisões

**Pendências**:
- [ ] Implementar VSYNC workflows (daily cron)
- [ ] Implementar OTP workflows
- [ ] Configurar PostgreSQL replication (master-slave)
- [ ] Criar Grafana dashboards completos
- [ ] Validar alerting rules em ambiente real
- [ ] Implementar workflow versioning strategy

---

**Referências**:
- [TEC-003: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [Temporal Documentation v1.36.0](https://docs.temporal.io/)
- [Temporal Go SDK v1.36.0](https://github.com/temporalio/sdk-go)
- [Kubernetes StatefulSets](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)
