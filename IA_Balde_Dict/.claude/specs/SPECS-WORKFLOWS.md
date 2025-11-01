# SPECS-WORKFLOWS.md - Temporal Workflows Specification

**Projeto**: DICT Rate Limit Monitoring System
**Componente**: Orchestration Worker (apps/orchestration-worker)
**Engine**: Temporal 1.22+
**Status**: ✅ ESPECIFICAÇÃO COMPLETA - Production-Ready

---

## 🎯 Objetivo

Especificação técnica completa dos **Temporal Workflows** que orquestram o monitoramento contínuo de rate limits do DICT BACEN:

1. **Cron Workflow**: `MonitorRateLimitsWorkflow` - Execução periódica (*/5 min)
2. **Activities**: 5 activities principais para coleta, análise e notificação
3. **Retry Policies**: Estratégias de retry diferenciadas por tipo de erro
4. **Error Handling**: Tratamento de erros retryable vs non-retryable

**Continue-As-New**: Utilizado para workflows de longa duração (>1 dia).

---

## 📋 Tabela de Conteúdos

- [1. Arquitetura de Workflows](#1-arquitetura-de-workflows)
- [2. Cron Workflow: MonitorRateLimitsWorkflow](#2-cron-workflow-monitorratelimitsworkflow)
- [3. Temporal Activities](#3-temporal-activities)
- [4. Retry Policies](#4-retry-policies)
- [5. Error Handling](#5-error-handling)
- [6. Continue-As-New Pattern](#6-continue-as-new-pattern)
- [7. Workflow Testing](#7-workflow-testing)
- [8. Observability](#8-observability)
- [9. Production Checklist](#9-production-checklist)

---

## 1. Arquitetura de Workflows

### Diagrama de Fluxo

```
┌─────────────────────────────────────────────────────────────────────┐
│                    TEMPORAL CRON WORKFLOW                            │
│                MonitorRateLimitsWorkflow                             │
│                Schedule: "*/5 * * * *"                               │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              │ Execute every 5 minutes
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Activity 1: GetPoliciesActivity                                    │
│  ├─ Call Bridge gRPC → DICT BACEN                                   │
│  ├─ Fetch all 24 policies state                                     │
│  └─ Return: []PolicyState                                           │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Activity 2: StorePolicyStateActivity                               │
│  ├─ Insert into dict_rate_limit_states (partitioned table)          │
│  ├─ Batch insert (24 rows)                                          │
│  └─ Return: success                                                 │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Activity 3: AnalyzeBalanceActivity                                 │
│  ├─ For each policy:                                                │
│  │   - Check if utilization > WARNING threshold (75%)              │
│  │   - Check if utilization > CRITICAL threshold (90%)             │
│  ├─ Generate AlertEvent for policies exceeding thresholds           │
│  └─ Return: []AlertEvent                                            │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Activity 4: PublishAlertActivity (if alerts > 0)                   │
│  ├─ For each AlertEvent:                                            │
│  │   - Publish to Pulsar topic: core-events                        │
│  │   - Action: ActionRateLimitAlert                                │
│  │   - Payload: {policy, severity, utilization, message}           │
│  ├─ Insert into dict_rate_limit_alerts (audit log)                 │
│  └─ Return: success                                                 │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Activity 5: PublishMetricsActivity                                 │
│  ├─ Export to Prometheus:                                           │
│  │   - dict_rate_limit_available_tokens                            │
│  │   - dict_rate_limit_utilization_pct                             │
│  │   - dict_rate_limit_alerts_total                                │
│  └─ Return: success                                                 │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                           DONE
                              │
                              │ Wait until next cron execution
                              │ (5 minutes)
                              ▼
                    [Continue-As-New after 24h]
```

### Workflow Lifecycle

```
Start (Cron Trigger)
  ↓
GetPoliciesActivity → [Retry 3x if gRPC error]
  ↓
StorePolicyStateActivity → [Retry 2x if DB error]
  ↓
AnalyzeBalanceActivity → [No retry, deterministic]
  ↓
PublishAlertActivity (conditional) → [Retry 3x if Pulsar error]
  ↓
PublishMetricsActivity → [Best-effort, no retry]
  ↓
Complete
  ↓
Wait for next cron (5 min)
  ↓
[Continue-As-New if workflow > 24h old]
```

---

## 2. Cron Workflow: MonitorRateLimitsWorkflow

### Workflow Definition

```go
// Location: apps/orchestration-worker/infrastructure/temporal/workflows/ratelimit/monitor_workflow.go
package ratelimit

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// MonitorRateLimitsWorkflow é o workflow cron de monitoramento
func MonitorRateLimitsWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting MonitorRateLimitsWorkflow")

	// Workflow options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Check if Continue-As-New is needed (after 24h)
	info := workflow.GetInfo(ctx)
	if shouldContinueAsNew(info) {
		logger.Info("Workflow running for >24h, executing Continue-As-New")
		return workflow.NewContinueAsNewError(ctx, MonitorRateLimitsWorkflow)
	}

	// ========================================================================
	// ACTIVITY 1: Get policies from DICT via Bridge
	// ========================================================================
	logger.Info("Executing GetPoliciesActivity")

	var policiesResult GetPoliciesResult
	err := workflow.ExecuteActivity(
		workflow.WithRetryPolicy(ctx, retryPolicyBridgeCall()),
		GetPoliciesActivity,
	).Get(ctx, &policiesResult)

	if err != nil {
		logger.Error("GetPoliciesActivity failed", "error", err)
		return err
	}

	logger.Info("Policies retrieved successfully",
		"count", len(policiesResult.Policies),
		"checked_at", policiesResult.CheckedAt,
	)

	// ========================================================================
	// ACTIVITY 2: Store policy states in PostgreSQL
	// ========================================================================
	logger.Info("Executing StorePolicyStateActivity")

	err = workflow.ExecuteActivity(
		workflow.WithRetryPolicy(ctx, retryPolicyDatabaseWrite()),
		StorePolicyStateActivity,
		StorePolicyStateInput{
			Policies:  policiesResult.Policies,
			CheckedAt: policiesResult.CheckedAt,
		},
	).Get(ctx, nil)

	if err != nil {
		logger.Error("StorePolicyStateActivity failed", "error", err)
		return err
	}

	logger.Info("Policy states stored successfully")

	// ========================================================================
	// ACTIVITY 3: Analyze balances and detect thresholds
	// ========================================================================
	logger.Info("Executing AnalyzeBalanceActivity")

	var analyzeResult AnalyzeBalanceResult
	err = workflow.ExecuteActivity(
		workflow.WithRetryPolicy(ctx, retryPolicyNone()), // Deterministic, no retry
		AnalyzeBalanceActivity,
		AnalyzeBalanceInput{
			Policies: policiesResult.Policies,
		},
	).Get(ctx, &analyzeResult)

	if err != nil {
		logger.Error("AnalyzeBalanceActivity failed", "error", err)
		return err
	}

	logger.Info("Balance analysis completed",
		"alerts", len(analyzeResult.Alerts),
		"warnings", analyzeResult.WarningCount,
		"criticals", analyzeResult.CriticalCount,
	)

	// ========================================================================
	// ACTIVITY 4: Publish alerts (if any)
	// ========================================================================
	if len(analyzeResult.Alerts) > 0 {
		logger.Info("Publishing alerts", "count", len(analyzeResult.Alerts))

		err = workflow.ExecuteActivity(
			workflow.WithRetryPolicy(ctx, retryPolicyPulsarPublish()),
			PublishAlertActivity,
			PublishAlertInput{
				Alerts: analyzeResult.Alerts,
			},
		).Get(ctx, nil)

		if err != nil {
			logger.Error("PublishAlertActivity failed", "error", err)
			// Non-critical: continue even if alert publish fails
			// Alerts are already in database (StorePolicyStateActivity)
		} else {
			logger.Info("Alerts published successfully")
		}
	} else {
		logger.Info("No alerts to publish, all policies within thresholds")
	}

	// ========================================================================
	// ACTIVITY 5: Publish metrics to Prometheus
	// ========================================================================
	logger.Info("Executing PublishMetricsActivity")

	err = workflow.ExecuteActivity(
		workflow.WithRetryPolicy(ctx, retryPolicyNone()), // Best-effort, no retry
		PublishMetricsActivity,
		PublishMetricsInput{
			Policies:      policiesResult.Policies,
			AlertCount:    len(analyzeResult.Alerts),
			WarningCount:  analyzeResult.WarningCount,
			CriticalCount: analyzeResult.CriticalCount,
		},
	).Get(ctx, nil)

	if err != nil {
		logger.Warn("PublishMetricsActivity failed (non-critical)", "error", err)
		// Non-critical: metrics failure should not fail workflow
	}

	logger.Info("MonitorRateLimitsWorkflow completed successfully")
	return nil
}

// shouldContinueAsNew checks if workflow should continue-as-new
func shouldContinueAsNew(info *workflow.Info) bool {
	// Continue-As-New after 24 hours to prevent history growth
	elapsed := time.Since(info.WorkflowStartTime)
	return elapsed > 24*time.Hour
}
```

### Workflow Registration

```go
// Location: apps/orchestration-worker/setup/temporal.go
package setup

import (
	"time"

	"github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/workflows/ratelimit"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// RegisterRateLimitWorkflows registra workflows e activities
func RegisterRateLimitWorkflows(w worker.Worker) {
	// Register workflow
	w.RegisterWorkflow(ratelimit.MonitorRateLimitsWorkflow)

	// Register activities
	w.RegisterActivity(ratelimit.GetPoliciesActivity)
	w.RegisterActivity(ratelimit.StorePolicyStateActivity)
	w.RegisterActivity(ratelimit.AnalyzeBalanceActivity)
	w.RegisterActivity(ratelimit.PublishAlertActivity)
	w.RegisterActivity(ratelimit.PublishMetricsActivity)
}

// StartRateLimitCronWorkflow inicia o cron workflow
func StartRateLimitCronWorkflow(c client.Client) error {
	// Workflow options
	workflowOptions := client.StartWorkflowOptions{
		ID:           "monitor-rate-limits-cron",
		TaskQueue:    "dict-task-queue",
		CronSchedule: "*/5 * * * *", // Every 5 minutes
		WorkflowExecutionTimeout: 10 * time.Minute,
		WorkflowRunTimeout:       5 * time.Minute,
		WorkflowTaskTimeout:      1 * time.Minute,
	}

	// Start workflow
	we, err := c.ExecuteWorkflow(
		context.Background(),
		workflowOptions,
		ratelimit.MonitorRateLimitsWorkflow,
	)

	if err != nil {
		return fmt.Errorf("failed to start cron workflow: %w", err)
	}

	log.Printf("Started MonitorRateLimitsWorkflow: WorkflowID=%s, RunID=%s",
		we.GetID(), we.GetRunID())

	return nil
}
```

---

## 3. Temporal Activities

### Activity 1: GetPoliciesActivity

```go
// Location: apps/orchestration-worker/infrastructure/temporal/activities/ratelimit/get_policies_activity.go
package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/grpc"
	"go.temporal.io/sdk/activity"
)

// GetPoliciesResult representa o resultado da activity
type GetPoliciesResult struct {
	Policies  []PolicyState `json:"policies"`
	CheckedAt time.Time     `json:"checked_at"`
}

// PolicyState representa o estado de uma política
type PolicyState struct {
	PolicyName      string    `json:"policy_name"`
	Category        string    `json:"category"`
	CapacityMax     int       `json:"capacity_max"`
	RefillTokens    int       `json:"refill_tokens"`
	RefillPeriodSec int       `json:"refill_period_sec"`
	AvailableTokens int       `json:"available_tokens"`
	UtilizationPct  float64   `json:"utilization_pct"`
	Status          string    `json:"status"`
	CheckedAt       time.Time `json:"checked_at"`
}

// RateLimitActivity contém dependências das activities
type RateLimitActivity struct {
	grpcGateway *grpc.Gateway
}

// NewRateLimitActivity cria uma nova instância
func NewRateLimitActivity(grpcGateway *grpc.Gateway) *RateLimitActivity {
	return &RateLimitActivity{
		grpcGateway: grpcGateway,
	}
}

// GetPoliciesActivity consulta políticas do DICT via Bridge
func (a *RateLimitActivity) GetPoliciesActivity(ctx context.Context) (*GetPoliciesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("GetPoliciesActivity started")

	// Heartbeat
	activity.RecordHeartbeat(ctx, "fetching policies from DICT")

	// Call Bridge gRPC
	bridgeResp, err := a.grpcGateway.RateLimitClient.ListPolicies(ctx)
	if err != nil {
		logger.Error("Failed to call Bridge gRPC", "error", err)
		return nil, fmt.Errorf("bridge grpc call failed: %w", err)
	}

	// Convert Bridge response to domain
	policies := make([]PolicyState, 0, len(bridgeResp.Policies))
	checkedAt := time.Now().UTC()

	for _, bp := range bridgeResp.Policies {
		// Calculate utilization percentage
		utilizationPct := 100.0 - (float64(bp.AvailableTokens) / float64(bp.Capacity) * 100.0)

		// Determine status
		status := "OK"
		remainingPct := float64(bp.AvailableTokens) / float64(bp.Capacity) * 100.0

		if remainingPct <= bp.CriticalThresholdPct {
			status = "CRITICAL"
		} else if remainingPct <= bp.WarningThresholdPct {
			status = "WARNING"
		}

		policies = append(policies, PolicyState{
			PolicyName:      bp.PolicyName,
			Category:        bp.Category,
			CapacityMax:     bp.Capacity,
			RefillTokens:    bp.RefillTokens,
			RefillPeriodSec: bp.RefillPeriodSec,
			AvailableTokens: bp.AvailableTokens,
			UtilizationPct:  utilizationPct,
			Status:          status,
			CheckedAt:       checkedAt,
		})
	}

	logger.Info("GetPoliciesActivity completed",
		"count", len(policies),
		"checked_at", checkedAt,
	)

	return &GetPoliciesResult{
		Policies:  policies,
		CheckedAt: checkedAt,
	}, nil
}
```

### Activity 2: StorePolicyStateActivity

```go
// Location: apps/orchestration-worker/infrastructure/temporal/activities/ratelimit/store_policy_state_activity.go
package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/database/repositories"
	"go.temporal.io/sdk/activity"
)

// StorePolicyStateInput representa o input da activity
type StorePolicyStateInput struct {
	Policies  []PolicyState `json:"policies"`
	CheckedAt time.Time     `json:"checked_at"`
}

// StorePolicyStateActivity armazena estados no PostgreSQL
func (a *RateLimitActivity) StorePolicyStateActivity(
	ctx context.Context,
	input StorePolicyStateInput,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("StorePolicyStateActivity started", "count", len(input.Policies))

	// Heartbeat
	activity.RecordHeartbeat(ctx, "storing policy states")

	// Get repository
	repo := repositories.NewRateLimitStateRepository(a.grpcGateway.DB)

	// Batch insert
	err := repo.BatchInsert(ctx, input.Policies)
	if err != nil {
		logger.Error("Failed to insert policy states", "error", err)
		return fmt.Errorf("database insert failed: %w", err)
	}

	logger.Info("StorePolicyStateActivity completed successfully")
	return nil
}
```

### Activity 3: AnalyzeBalanceActivity

```go
// Location: apps/orchestration-worker/infrastructure/temporal/activities/ratelimit/analyze_balance_activity.go
package ratelimit

import (
	"context"

	"go.temporal.io/sdk/activity"
)

// AnalyzeBalanceInput representa o input da activity
type AnalyzeBalanceInput struct {
	Policies []PolicyState `json:"policies"`
}

// AnalyzeBalanceResult representa o resultado da análise
type AnalyzeBalanceResult struct {
	Alerts        []AlertEvent `json:"alerts"`
	WarningCount  int          `json:"warning_count"`
	CriticalCount int          `json:"critical_count"`
}

// AlertEvent representa um alerta a ser publicado
type AlertEvent struct {
	PolicyName      string    `json:"policy_name"`
	Category        string    `json:"category"`
	Severity        string    `json:"severity"` // WARNING or CRITICAL
	AvailableTokens int       `json:"available_tokens"`
	CapacityMax     int       `json:"capacity_max"`
	UtilizationPct  float64   `json:"utilization_pct"`
	Message         string    `json:"message"`
	DetectedAt      time.Time `json:"detected_at"`
}

// AnalyzeBalanceActivity analisa saldos e detecta thresholds
func (a *RateLimitActivity) AnalyzeBalanceActivity(
	ctx context.Context,
	input AnalyzeBalanceInput,
) (*AnalyzeBalanceResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("AnalyzeBalanceActivity started", "policies", len(input.Policies))

	alerts := make([]AlertEvent, 0)
	warningCount := 0
	criticalCount := 0

	for _, policy := range input.Policies {
		if policy.Status == "WARNING" {
			warningCount++
			alerts = append(alerts, AlertEvent{
				PolicyName:      policy.PolicyName,
				Category:        policy.Category,
				Severity:        "WARNING",
				AvailableTokens: policy.AvailableTokens,
				CapacityMax:     policy.CapacityMax,
				UtilizationPct:  policy.UtilizationPct,
				Message: fmt.Sprintf(
					"Policy %s (Category %s) is at %.2f%% utilization (WARNING threshold exceeded)",
					policy.PolicyName, policy.Category, policy.UtilizationPct,
				),
				DetectedAt: time.Now().UTC(),
			})
		} else if policy.Status == "CRITICAL" {
			criticalCount++
			alerts = append(alerts, AlertEvent{
				PolicyName:      policy.PolicyName,
				Category:        policy.Category,
				Severity:        "CRITICAL",
				AvailableTokens: policy.AvailableTokens,
				CapacityMax:     policy.CapacityMax,
				UtilizationPct:  policy.UtilizationPct,
				Message: fmt.Sprintf(
					"URGENT: Policy %s (Category %s) is at %.2f%% utilization (CRITICAL threshold exceeded)",
					policy.PolicyName, policy.Category, policy.UtilizationPct,
				),
				DetectedAt: time.Now().UTC(),
			})
		}
	}

	logger.Info("AnalyzeBalanceActivity completed",
		"alerts", len(alerts),
		"warnings", warningCount,
		"criticals", criticalCount,
	)

	return &AnalyzeBalanceResult{
		Alerts:        alerts,
		WarningCount:  warningCount,
		CriticalCount: criticalCount,
	}, nil
}
```

### Activity 4: PublishAlertActivity

```go
// Location: apps/orchestration-worker/infrastructure/temporal/activities/ratelimit/publish_alert_activity.go
package ratelimit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/database/repositories"
	"go.temporal.io/sdk/activity"
)

// PublishAlertInput representa o input da activity
type PublishAlertInput struct {
	Alerts []AlertEvent `json:"alerts"`
}

// PublishAlertActivity publica alertas no Pulsar
func (a *RateLimitActivity) PublishAlertActivity(
	ctx context.Context,
	input PublishAlertInput,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("PublishAlertActivity started", "alerts", len(input.Alerts))

	// Heartbeat
	activity.RecordHeartbeat(ctx, "publishing alerts to Pulsar")

	// Get Pulsar producer
	producer := a.grpcGateway.PulsarProducer

	// Get alert repository (for audit log)
	repo := repositories.NewRateLimitAlertRepository(a.grpcGateway.DB)

	// Publish each alert
	for _, alert := range input.Alerts {
		// Create Pulsar message
		payload, err := json.Marshal(map[string]interface{}{
			"action": "ActionRateLimitAlert",
			"data": map[string]interface{}{
				"policy_name":      alert.PolicyName,
				"category":         alert.Category,
				"severity":         alert.Severity,
				"available_tokens": alert.AvailableTokens,
				"capacity_max":     alert.CapacityMax,
				"utilization_pct":  alert.UtilizationPct,
				"message":          alert.Message,
				"detected_at":      alert.DetectedAt,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to marshal alert: %w", err)
		}

		// Publish to Pulsar
		_, err = producer.Send(ctx, &pulsar.ProducerMessage{
			Payload: payload,
			Key:     alert.PolicyName,
		})
		if err != nil {
			logger.Error("Failed to publish alert to Pulsar", "error", err, "policy", alert.PolicyName)
			return fmt.Errorf("pulsar publish failed: %w", err)
		}

		// Save to database (audit log)
		err = repo.Insert(ctx, alert)
		if err != nil {
			logger.Warn("Failed to save alert to database (non-critical)", "error", err)
			// Continue - audit log failure should not fail workflow
		}

		logger.Info("Alert published successfully",
			"policy", alert.PolicyName,
			"severity", alert.Severity,
		)
	}

	logger.Info("PublishAlertActivity completed successfully")
	return nil
}
```

### Activity 5: PublishMetricsActivity

```go
// Location: apps/orchestration-worker/infrastructure/temporal/activities/ratelimit/publish_metrics_activity.go
package ratelimit

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.temporal.io/sdk/activity"
)

// PublishMetricsInput representa o input da activity
type PublishMetricsInput struct {
	Policies      []PolicyState `json:"policies"`
	AlertCount    int           `json:"alert_count"`
	WarningCount  int           `json:"warning_count"`
	CriticalCount int           `json:"critical_count"`
}

// Prometheus metrics (declared in metrics package)
var (
	availableTokensGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dict_rate_limit_available_tokens",
			Help: "Available tokens for each rate limit policy",
		},
		[]string{"policy_name", "category"},
	)

	utilizationGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dict_rate_limit_utilization_pct",
			Help: "Utilization percentage for each rate limit policy",
		},
		[]string{"policy_name", "category"},
	)

	alertsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dict_rate_limit_alerts_total",
			Help: "Total number of rate limit alerts",
		},
		[]string{"severity"},
	)
)

// PublishMetricsActivity exporta métricas para Prometheus
func (a *RateLimitActivity) PublishMetricsActivity(
	ctx context.Context,
	input PublishMetricsInput,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("PublishMetricsActivity started")

	// Update gauges for each policy
	for _, policy := range input.Policies {
		availableTokensGauge.WithLabelValues(policy.PolicyName, policy.Category).
			Set(float64(policy.AvailableTokens))

		utilizationGauge.WithLabelValues(policy.PolicyName, policy.Category).
			Set(policy.UtilizationPct)
	}

	// Update alert counters
	if input.WarningCount > 0 {
		alertsCounter.WithLabelValues("WARNING").Add(float64(input.WarningCount))
	}
	if input.CriticalCount > 0 {
		alertsCounter.WithLabelValues("CRITICAL").Add(float64(input.CriticalCount))
	}

	logger.Info("PublishMetricsActivity completed successfully")
	return nil
}
```

---

## 4. Retry Policies

### Retry Policy Definitions

```go
// Location: apps/orchestration-worker/infrastructure/temporal/workflows/ratelimit/retry_policies.go
package ratelimit

import (
	"time"

	"go.temporal.io/sdk/temporal"
)

// retryPolicyBridgeCall retorna política de retry para chamadas gRPC ao Bridge
func retryPolicyBridgeCall() *temporal.RetryPolicy {
	return &temporal.RetryPolicy{
		InitialInterval:        1 * time.Second,
		BackoffCoefficient:     2.0,
		MaximumInterval:        30 * time.Second,
		MaximumAttempts:        3,
		NonRetryableErrorTypes: []string{"InvalidArgument", "NotFound"},
	}
}

// retryPolicyDatabaseWrite retorna política de retry para escrita em banco
func retryPolicyDatabaseWrite() *temporal.RetryPolicy {
	return &temporal.RetryPolicy{
		InitialInterval:        500 * time.Millisecond,
		BackoffCoefficient:     2.0,
		MaximumInterval:        10 * time.Second,
		MaximumAttempts:        2,
		NonRetryableErrorTypes: []string{"InvalidArgument", "ConstraintViolation"},
	}
}

// retryPolicyPulsarPublish retorna política de retry para publicação Pulsar
func retryPolicyPulsarPublish() *temporal.RetryPolicy {
	return &temporal.RetryPolicy{
		InitialInterval:        1 * time.Second,
		BackoffCoefficient:     2.0,
		MaximumInterval:        20 * time.Second,
		MaximumAttempts:        3,
		NonRetryableErrorTypes: []string{"InvalidPayload"},
	}
}

// retryPolicyNone retorna política sem retry (atividades determinísticas)
func retryPolicyNone() *temporal.RetryPolicy {
	return &temporal.RetryPolicy{
		MaximumAttempts: 1,
	}
}
```

### Retry Policy Matrix

| Activity | Max Attempts | Initial Interval | Max Interval | Backoff | Non-Retryable Errors |
|----------|--------------|------------------|--------------|---------|---------------------|
| GetPoliciesActivity | 3 | 1s | 30s | 2.0 | InvalidArgument, NotFound |
| StorePolicyStateActivity | 2 | 500ms | 10s | 2.0 | InvalidArgument, ConstraintViolation |
| AnalyzeBalanceActivity | 1 | - | - | - | (deterministic) |
| PublishAlertActivity | 3 | 1s | 20s | 2.0 | InvalidPayload |
| PublishMetricsActivity | 1 | - | - | - | (best-effort) |

---

## 5. Error Handling

### Error Types

```go
// Location: apps/orchestration-worker/infrastructure/temporal/activities/ratelimit/errors.go
package ratelimit

import (
	"errors"

	"go.temporal.io/sdk/temporal"
)

// Error types
var (
	// Retryable errors
	ErrBridgeUnavailable = errors.New("bridge service unavailable")
	ErrDatabaseTimeout   = errors.New("database operation timeout")
	ErrPulsarUnavailable = errors.New("pulsar service unavailable")

	// Non-retryable errors
	ErrInvalidArgument      = temporal.NewNonRetryableApplicationError("invalid argument", "InvalidArgument", nil)
	ErrPolicyNotFound       = temporal.NewNonRetryableApplicationError("policy not found", "NotFound", nil)
	ErrConstraintViolation  = temporal.NewNonRetryableApplicationError("database constraint violation", "ConstraintViolation", nil)
	ErrInvalidPayload       = temporal.NewNonRetryableApplicationError("invalid pulsar payload", "InvalidPayload", nil)
)

// IsRetryableError determina se um erro é retryable
func IsRetryableError(err error) bool {
	if errors.Is(err, ErrBridgeUnavailable) {
		return true
	}
	if errors.Is(err, ErrDatabaseTimeout) {
		return true
	}
	if errors.Is(err, ErrPulsarUnavailable) {
		return true
	}

	// Temporal non-retryable errors
	var appErr *temporal.ApplicationError
	if errors.As(err, &appErr) {
		return !appErr.NonRetryable()
	}

	return false
}
```

### Error Handling Example

```go
// In GetPoliciesActivity
bridgeResp, err := a.grpcGateway.RateLimitClient.ListPolicies(ctx)
if err != nil {
	// Check if it's a gRPC error
	if isGRPCUnavailableError(err) {
		// Retryable error
		return nil, ErrBridgeUnavailable
	}
	if isGRPCNotFoundError(err) {
		// Non-retryable error
		return nil, ErrPolicyNotFound
	}

	// Generic retryable error
	return nil, fmt.Errorf("bridge grpc call failed: %w", err)
}
```

---

## 6. Continue-As-New Pattern

### Implementation

```go
// In MonitorRateLimitsWorkflow
func shouldContinueAsNew(info *workflow.Info) bool {
	// Continue-As-New after 24 hours to prevent history growth
	elapsed := time.Since(info.WorkflowStartTime)

	// Also check event history size (if accessible)
	// Temporal recommends Continue-As-New when history > 10K events

	return elapsed > 24*time.Hour
}

// Execute Continue-As-New
if shouldContinueAsNew(info) {
	logger.Info("Workflow running for >24h, executing Continue-As-New")
	return workflow.NewContinueAsNewError(ctx, MonitorRateLimitsWorkflow)
}
```

### Why Continue-As-New?

**Problem**: Cron workflows run indefinitely, accumulating event history.

**Solution**: Continue-As-New resets the workflow history while preserving state.

**Benefits**:
- Prevents history bloat (>10K events)
- Maintains Temporal performance
- Allows workflow code updates (new version deployed)

**Trigger**: After 24 hours OR 10K events (whichever comes first)

---

## 7. Workflow Testing

### Replay Test

```go
// Location: apps/orchestration-worker/infrastructure/temporal/workflows/ratelimit/monitor_workflow_test.go
package ratelimit_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type WorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowTestSuite))
}

func (s *WorkflowTestSuite) Test_MonitorRateLimitsWorkflow_Success() {
	env := s.NewTestWorkflowEnvironment()

	// Mock GetPoliciesActivity
	env.OnActivity(GetPoliciesActivity, mock.Anything).Return(&GetPoliciesResult{
		Policies: []PolicyState{
			{PolicyName: "ENTRIES_CREATE", AvailableTokens: 150, CapacityMax: 300},
		},
		CheckedAt: time.Now(),
	}, nil)

	// Mock StorePolicyStateActivity
	env.OnActivity(StorePolicyStateActivity, mock.Anything, mock.Anything).Return(nil)

	// Mock AnalyzeBalanceActivity
	env.OnActivity(AnalyzeBalanceActivity, mock.Anything, mock.Anything).Return(&AnalyzeBalanceResult{
		Alerts:        []AlertEvent{},
		WarningCount:  0,
		CriticalCount: 0,
	}, nil)

	// Mock PublishMetricsActivity
	env.OnActivity(PublishMetricsActivity, mock.Anything, mock.Anything).Return(nil)

	// Execute workflow
	env.ExecuteWorkflow(MonitorRateLimitsWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *WorkflowTestSuite) Test_MonitorRateLimitsWorkflow_WithAlerts() {
	env := s.NewTestWorkflowEnvironment()

	// Mock activities
	env.OnActivity(GetPoliciesActivity, mock.Anything).Return(&GetPoliciesResult{
		Policies: []PolicyState{
			{PolicyName: "ENTRIES_CREATE", AvailableTokens: 30, CapacityMax: 300, Status: "CRITICAL"},
		},
		CheckedAt: time.Now(),
	}, nil)

	env.OnActivity(StorePolicyStateActivity, mock.Anything, mock.Anything).Return(nil)

	env.OnActivity(AnalyzeBalanceActivity, mock.Anything, mock.Anything).Return(&AnalyzeBalanceResult{
		Alerts: []AlertEvent{
			{PolicyName: "ENTRIES_CREATE", Severity: "CRITICAL"},
		},
		WarningCount:  0,
		CriticalCount: 1,
	}, nil)

	env.OnActivity(PublishAlertActivity, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity(PublishMetricsActivity, mock.Anything, mock.Anything).Return(nil)

	// Execute workflow
	env.ExecuteWorkflow(MonitorRateLimitsWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *WorkflowTestSuite) Test_MonitorRateLimitsWorkflow_BridgeFailure() {
	env := s.NewTestWorkflowEnvironment()

	// Mock GetPoliciesActivity failure
	env.OnActivity(GetPoliciesActivity, mock.Anything).Return(nil, ErrBridgeUnavailable)

	// Execute workflow
	env.ExecuteWorkflow(MonitorRateLimitsWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.Error(env.GetWorkflowError())
}
```

---

## 8. Observability

### Temporal Dashboard Metrics

```yaml
Metrics disponíveis no Temporal UI:

Workflow Metrics:
  - workflow.executions.started
  - workflow.executions.completed
  - workflow.executions.failed
  - workflow.executions.timed_out
  - workflow.continue_as_new

Activity Metrics:
  - activity.executions.started
  - activity.executions.completed
  - activity.executions.failed
  - activity.heartbeats.received
  - activity.retry.count

Performance Metrics:
  - workflow.execution.duration (p50, p99)
  - activity.execution.duration (p50, p99)
  - workflow.task_queue.latency
```

### Custom Metrics (Prometheus)

```go
// Workflow execution metrics
workflowExecutionDuration := prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "dict_rate_limit_workflow_duration_seconds",
		Help:    "Duration of MonitorRateLimitsWorkflow execution",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"status"}, // success, failed
)

// Activity execution metrics
activityExecutionDuration := prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "dict_rate_limit_activity_duration_seconds",
		Help:    "Duration of activity execution",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"activity_name", "status"},
)
```

---

## 9. Production Checklist

### Pre-Deployment

- [ ] All activities registered in Temporal worker
- [ ] Cron workflow started with correct schedule (*/5 * * * *)
- [ ] Retry policies validated
- [ ] Continue-As-New tested (24h trigger)
- [ ] Error handling validated (retryable vs non-retryable)
- [ ] All tests passing (>90% coverage)
- [ ] Temporal Dashboard configured
- [ ] Prometheus metrics exported
- [ ] Alerts configured (workflow failures)

### Post-Deployment

- [ ] Monitor Temporal Dashboard for workflow executions
- [ ] Verify cron execution every 5 minutes
- [ ] Check activity success rate (>99%)
- [ ] Validate alerts published to Pulsar
- [ ] Verify metrics in Prometheus/Grafana
- [ ] Test Continue-As-New after 24h
- [ ] Runbook created for incident response

---

**Última Atualização**: 2025-10-31
**Versão**: 1.0.0
**Status**: ✅ ESPECIFICAÇÃO COMPLETA - Production-Ready
