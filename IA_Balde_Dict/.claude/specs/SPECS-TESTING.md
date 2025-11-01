# SPECS-TESTING.md - Testing Strategy Specification

**Projeto**: DICT Rate Limit Monitoring System
**Framework**: Testify + MockGen + Testcontainers + Temporal SDK
**Target**: >85% Code Coverage
**Status**: ✅ ESPECIFICAÇÃO COMPLETA - Production-Ready

---

## 🎯 Objetivo

Especificação técnica completa da **estratégia de testes** do sistema:

1. **Unit Tests**: Testes unitários (>90% coverage por camada)
2. **Integration Tests**: Testes de integração (PostgreSQL, Redis, Pulsar)
3. **Temporal Tests**: Workflow replay tests e activity mocking
4. **E2E Tests**: Testes end-to-end completos
5. **Performance Tests**: Load testing e benchmarks
6. **Contract Tests**: Validação de contratos gRPC e Pulsar

**Meta**: Garantir qualidade production-ready com test automation completa.

---

## 📋 Tabela de Conteúdos

- [1. Test Pyramid](#1-test-pyramid)
- [2. Unit Tests](#2-unit-tests)
- [3. Integration Tests](#3-integration-tests)
- [4. Temporal Workflow Tests](#4-temporal-workflow-tests)
- [5. E2E Tests](#5-e2e-tests)
- [6. Performance Tests](#6-performance-tests)
- [7. Contract Tests](#7-contract-tests)
- [8. Test Data & Fixtures](#8-test-data--fixtures)
- [9. CI/CD Integration](#9-cicd-integration)

---

## 1. Test Pyramid

### Testing Strategy Overview

```
                    ┌─────────────┐
                    │   E2E (5%)  │  ← 20 tests (full system)
                    └─────────────┘
                  ┌───────────────────┐
                  │ Integration (15%) │  ← 60 tests (DB, gRPC, Pulsar)
                  └───────────────────┘
              ┌─────────────────────────────┐
              │      Unit Tests (80%)       │  ← 320 tests (logic, domain)
              └─────────────────────────────┘

Total: ~400 tests
Target Coverage: >85%
Execution Time: <3 minutes (CI pipeline)
```

### Test Distribution by Layer

| Layer | Unit Tests | Integration Tests | E2E Tests | Total |
|-------|------------|-------------------|-----------|-------|
| Domain | 60 | 0 | 0 | 60 |
| Application | 80 | 20 | 0 | 100 |
| Infrastructure | 100 | 30 | 5 | 135 |
| Handlers (HTTP) | 60 | 10 | 5 | 75 |
| Workflows (Temporal) | 20 | 0 | 10 | 30 |
| **Total** | **320** | **60** | **20** | **400** |

---

## 2. Unit Tests

### Domain Layer Tests

```go
// Location: apps/orchestration-worker/domain/ratelimit/policy_test.go
package ratelimit_test

import (
	"testing"

	"github.com/lb-conn/connector-dict/apps/orchestration-worker/domain/ratelimit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicy_CalculateUtilization(t *testing.T) {
	tests := []struct {
		name            string
		availableTokens int
		capacityMax     int
		want            float64
	}{
		{
			name:            "50% utilization",
			availableTokens: 150,
			capacityMax:     300,
			want:            50.0,
		},
		{
			name:            "90% utilization (critical)",
			availableTokens: 30,
			capacityMax:     300,
			want:            90.0,
		},
		{
			name:            "0% utilization (full)",
			availableTokens: 300,
			capacityMax:     300,
			want:            0.0,
		},
		{
			name:            "100% utilization (empty)",
			availableTokens: 0,
			capacityMax:     300,
			want:            100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &ratelimit.Policy{
				AvailableTokens: tt.availableTokens,
				CapacityMax:     tt.capacityMax,
			}

			got := policy.CalculateUtilization()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPolicy_DetermineStatus(t *testing.T) {
	tests := []struct {
		name                 string
		availableTokens      int
		capacityMax          int
		warningThresholdPct  float64
		criticalThresholdPct float64
		want                 string
	}{
		{
			name:                 "OK status (50% remaining)",
			availableTokens:      150,
			capacityMax:          300,
			warningThresholdPct:  25.0,
			criticalThresholdPct: 10.0,
			want:                 "OK",
		},
		{
			name:                 "WARNING status (20% remaining)",
			availableTokens:      60,
			capacityMax:          300,
			warningThresholdPct:  25.0,
			criticalThresholdPct: 10.0,
			want:                 "WARNING",
		},
		{
			name:                 "CRITICAL status (5% remaining)",
			availableTokens:      15,
			capacityMax:          300,
			warningThresholdPct:  25.0,
			criticalThresholdPct: 10.0,
			want:                 "CRITICAL",
		},
		{
			name:                 "CRITICAL status (0% remaining)",
			availableTokens:      0,
			capacityMax:          300,
			warningThresholdPct:  25.0,
			criticalThresholdPct: 10.0,
			want:                 "CRITICAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &ratelimit.Policy{
				AvailableTokens:      tt.availableTokens,
				CapacityMax:          tt.capacityMax,
				WarningThresholdPct:  tt.warningThresholdPct,
				CriticalThresholdPct: tt.criticalThresholdPct,
			}

			got := policy.DetermineStatus()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPolicy_Validate(t *testing.T) {
	tests := []struct {
		name    string
		policy  *ratelimit.Policy
		wantErr bool
	}{
		{
			name: "valid policy",
			policy: &ratelimit.Policy{
				PolicyName:      "ENTRIES_CREATE",
				Category:        "A",
				CapacityMax:     300,
				RefillTokens:    5,
				RefillPeriodSec: 60,
				AvailableTokens: 150,
			},
			wantErr: false,
		},
		{
			name: "invalid - empty policy name",
			policy: &ratelimit.Policy{
				PolicyName:  "",
				CapacityMax: 300,
			},
			wantErr: true,
		},
		{
			name: "invalid - negative capacity",
			policy: &ratelimit.Policy{
				PolicyName:  "ENTRIES_CREATE",
				CapacityMax: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid - available > capacity",
			policy: &ratelimit.Policy{
				PolicyName:      "ENTRIES_CREATE",
				CapacityMax:     300,
				AvailableTokens: 400,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
```

### Application Layer Tests

```go
// Location: apps/dict/application/usecases/ratelimit/service_test.go
package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/lb-conn/connector-dict/apps/dict/application/usecases/ratelimit"
	"github.com/lb-conn/connector-dict/apps/dict/handlers/http/ratelimit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockBridgeClient é um mock do cliente gRPC
type MockBridgeClient struct {
	mock.Mock
}

func (m *MockBridgeClient) ListPolicies(ctx context.Context) ([]bridge.PolicyState, error) {
	args := m.Called(ctx)
	return args.Get(0).([]bridge.PolicyState), args.Error(1)
}

func (m *MockBridgeClient) GetPolicy(ctx context.Context, policyName string) (*bridge.PolicyState, error) {
	args := m.Called(ctx, policyName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*bridge.PolicyState), args.Error(1)
}

// MockCache é um mock do Redis cache
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) GetWithTTL(ctx context.Context, key string, dest interface{}) (time.Duration, error) {
	args := m.Called(ctx, key, dest)
	return args.Get(0).(time.Duration), args.Error(1)
}

func TestService_ListPolicies_CacheHit(t *testing.T) {
	// Setup
	mockBridge := new(MockBridgeClient)
	mockCache := new(MockCache)

	service := ratelimit.NewService(mockBridge, mockCache, nil, 60*time.Second)

	// Mock cache hit
	cachedPolicies := []ratelimit.PolicySummary{
		{PolicyName: "ENTRIES_CREATE", AvailableTokens: 150},
	}
	mockCache.On("GetWithTTL", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			dest := args.Get(2).(*[]ratelimit.PolicySummary)
			*dest = cachedPolicies
		}).
		Return(45*time.Second, nil)

	// Execute
	policies, cached, ttl, err := service.ListPolicies(context.Background(), nil, nil)

	// Assert
	require.NoError(t, err)
	assert.True(t, cached)
	assert.Equal(t, 45*time.Second, ttl)
	assert.Len(t, policies, 1)
	assert.Equal(t, "ENTRIES_CREATE", policies[0].PolicyName)

	// Bridge should NOT be called on cache hit
	mockBridge.AssertNotCalled(t, "ListPolicies")
	mockCache.AssertExpectations(t)
}

func TestService_ListPolicies_CacheMiss(t *testing.T) {
	// Setup
	mockBridge := new(MockBridgeClient)
	mockCache := new(MockCache)

	service := ratelimit.NewService(mockBridge, mockCache, nil, 60*time.Second)

	// Mock cache miss
	mockCache.On("GetWithTTL", mock.Anything, mock.Anything, mock.Anything).
		Return(time.Duration(0), fmt.Errorf("cache miss"))

	// Mock bridge response
	bridgePolicies := []bridge.PolicyState{
		{
			PolicyName:      "ENTRIES_CREATE",
			AvailableTokens: 150,
			Capacity:        300,
		},
	}
	mockBridge.On("ListPolicies", mock.Anything).Return(bridgePolicies, nil)

	// Mock cache set
	mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, 60*time.Second).
		Return(nil)

	// Execute
	policies, cached, ttl, err := service.ListPolicies(context.Background(), nil, nil)

	// Assert
	require.NoError(t, err)
	assert.False(t, cached)
	assert.Equal(t, 60*time.Second, ttl)
	assert.Len(t, policies, 1)

	mockBridge.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestService_ListPolicies_WithFilters(t *testing.T) {
	// Setup
	mockBridge := new(MockBridgeClient)
	mockCache := new(MockCache)

	service := ratelimit.NewService(mockBridge, mockCache, nil, 60*time.Second)

	// Mock cache miss
	mockCache.On("GetWithTTL", mock.Anything, mock.Anything, mock.Anything).
		Return(time.Duration(0), fmt.Errorf("cache miss"))

	// Mock bridge response (3 policies)
	bridgePolicies := []bridge.PolicyState{
		{PolicyName: "ENTRIES_CREATE", Category: "A", AvailableTokens: 150, Capacity: 300}, // 50% util
		{PolicyName: "CLAIMS_CREATE", Category: "B", AvailableTokens: 50, Capacity: 1000},  // 95% util
		{PolicyName: "ACCOUNT_LIST", Category: "A", AvailableTokens: 100, Capacity: 200},   // 50% util
	}
	mockBridge.On("ListPolicies", mock.Anything).Return(bridgePolicies, nil)
	mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Execute with filters: category=A, minUtilization=90
	category := "A"
	minUtil := 90.0
	policies, _, _, err := service.ListPolicies(context.Background(), &category, &minUtil)

	// Assert - should return 0 policies (no Category A with >90% utilization)
	require.NoError(t, err)
	assert.Len(t, policies, 0)
}
```

---

## 3. Integration Tests

### PostgreSQL Integration Tests

```go
// Location: apps/orchestration-worker/infrastructure/database/repositories/integration_test.go
package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/database/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupPostgresContainer inicia container PostgreSQL para testes
func setupPostgresContainer(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get connection string
	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")
	connStr := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())

	// Connect
	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	// Run migrations
	runMigrations(t, pool)

	cleanup := func() {
		pool.Close()
		container.Terminate(ctx)
	}

	return pool, cleanup
}

func runMigrations(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	// Create tables (simplified - use actual migration files in production)
	_, err := pool.Exec(ctx, `
		CREATE TABLE dict_rate_limit_states (
			id BIGSERIAL,
			policy_name VARCHAR(100) NOT NULL,
			available_tokens INTEGER NOT NULL,
			capacity INTEGER NOT NULL,
			utilization_pct DECIMAL(5,2) NOT NULL,
			category VARCHAR(1),
			checked_at TIMESTAMP WITH TIME ZONE NOT NULL,
			PRIMARY KEY (id, checked_at)
		) PARTITION BY RANGE (checked_at);

		CREATE TABLE dict_rate_limit_states_2025_10 PARTITION OF dict_rate_limit_states
			FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
	`)
	require.NoError(t, err)
}

func TestRateLimitStateRepository_BatchInsert(t *testing.T) {
	// Setup
	pool, cleanup := setupPostgresContainer(t)
	defer cleanup()

	repo := repositories.NewRateLimitStateRepository(pool)

	// Test data
	policies := []ratelimit.PolicyState{
		{
			PolicyName:      "ENTRIES_CREATE",
			Category:        "A",
			AvailableTokens: 150,
			CapacityMax:     300,
			UtilizationPct:  50.0,
			CheckedAt:       time.Now(),
		},
		{
			PolicyName:      "CLAIMS_CREATE",
			Category:        "B",
			AvailableTokens: 500,
			CapacityMax:     1000,
			UtilizationPct:  50.0,
			CheckedAt:       time.Now(),
		},
	}

	// Execute
	err := repo.BatchInsert(context.Background(), policies)

	// Assert
	require.NoError(t, err)

	// Verify insertion
	var count int
	err = pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM dict_rate_limit_states").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestRateLimitStateRepository_GetLatestStates(t *testing.T) {
	// Setup
	pool, cleanup := setupPostgresContainer(t)
	defer cleanup()

	repo := repositories.NewRateLimitStateRepository(pool)

	// Insert test data
	now := time.Now()
	policies := []ratelimit.PolicyState{
		{
			PolicyName:     "ENTRIES_CREATE",
			AvailableTokens: 150,
			CapacityMax:    300,
			CheckedAt:      now.Add(-5 * time.Minute), // Old
		},
		{
			PolicyName:     "ENTRIES_CREATE",
			AvailableTokens: 100,
			CapacityMax:    300,
			CheckedAt:      now, // Latest
		},
	}
	err := repo.BatchInsert(context.Background(), policies)
	require.NoError(t, err)

	// Execute
	latest, err := repo.GetLatestStates(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Len(t, latest, 1)
	assert.Equal(t, 100, latest[0].AvailableTokens)
}
```

### Redis Cache Integration Tests

```go
// Location: apps/dict/infrastructure/cache/integration_test.go
package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/lb-conn/connector-dict/apps/dict/infrastructure/cache"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupRedisContainer(t *testing.T) (*redis.Client, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "6379")

	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port.Port()),
	})

	cleanup := func() {
		client.Close()
		container.Terminate(ctx)
	}

	return client, cleanup
}

func TestRedisCache_SetAndGet(t *testing.T) {
	// Setup
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	cache := cache.NewRedisCache(client)

	// Test data
	type TestData struct {
		Name  string
		Value int
	}
	data := TestData{Name: "test", Value: 123}

	// Execute Set
	err := cache.Set(context.Background(), "test:key", data, 60*time.Second)
	require.NoError(t, err)

	// Execute Get
	var result TestData
	ttl, err := cache.GetWithTTL(context.Background(), "test:key", &result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, data, result)
	assert.Greater(t, ttl, 50*time.Second)
	assert.LessOrEqual(t, ttl, 60*time.Second)
}

func TestRedisCache_GetMiss(t *testing.T) {
	// Setup
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	cache := cache.NewRedisCache(client)

	// Execute Get on non-existent key
	var result string
	_, err := cache.GetWithTTL(context.Background(), "nonexistent:key", &result)

	// Assert
	assert.Error(t, err)
}
```

---

## 4. Temporal Workflow Tests

### Workflow Replay Tests

```go
// Location: apps/orchestration-worker/infrastructure/temporal/workflows/ratelimit/monitor_workflow_test.go
package ratelimit_test

import (
	"testing"
	"time"

	"github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/workflows/ratelimit"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type MonitorWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestMonitorWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(MonitorWorkflowTestSuite))
}

func (s *MonitorWorkflowTestSuite) Test_MonitorRateLimitsWorkflow_Success() {
	env := s.NewTestWorkflowEnvironment()

	// Mock GetPoliciesActivity
	env.OnActivity(ratelimit.GetPoliciesActivity, mock.Anything).Return(&ratelimit.GetPoliciesResult{
		Policies: []ratelimit.PolicyState{
			{
				PolicyName:      "ENTRIES_CREATE",
				AvailableTokens: 150,
				CapacityMax:     300,
				Status:          "OK",
			},
		},
		CheckedAt: time.Now(),
	}, nil)

	// Mock StorePolicyStateActivity
	env.OnActivity(ratelimit.StorePolicyStateActivity, mock.Anything, mock.Anything).Return(nil)

	// Mock AnalyzeBalanceActivity
	env.OnActivity(ratelimit.AnalyzeBalanceActivity, mock.Anything, mock.Anything).Return(&ratelimit.AnalyzeBalanceResult{
		Alerts:        []ratelimit.AlertEvent{},
		WarningCount:  0,
		CriticalCount: 0,
	}, nil)

	// Mock PublishMetricsActivity
	env.OnActivity(ratelimit.PublishMetricsActivity, mock.Anything, mock.Anything).Return(nil)

	// Execute workflow
	env.ExecuteWorkflow(ratelimit.MonitorRateLimitsWorkflow)

	// Assert
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *MonitorWorkflowTestSuite) Test_MonitorRateLimitsWorkflow_WithCriticalAlerts() {
	env := s.NewTestWorkflowEnvironment()

	// Mock GetPoliciesActivity with critical policy
	env.OnActivity(ratelimit.GetPoliciesActivity, mock.Anything).Return(&ratelimit.GetPoliciesResult{
		Policies: []ratelimit.PolicyState{
			{
				PolicyName:      "ENTRIES_CREATE",
				AvailableTokens: 30,
				CapacityMax:     300,
				Status:          "CRITICAL",
			},
		},
		CheckedAt: time.Now(),
	}, nil)

	env.OnActivity(ratelimit.StorePolicyStateActivity, mock.Anything, mock.Anything).Return(nil)

	// Mock AnalyzeBalanceActivity with alerts
	env.OnActivity(ratelimit.AnalyzeBalanceActivity, mock.Anything, mock.Anything).Return(&ratelimit.AnalyzeBalanceResult{
		Alerts: []ratelimit.AlertEvent{
			{
				PolicyName: "ENTRIES_CREATE",
				Severity:   "CRITICAL",
			},
		},
		WarningCount:  0,
		CriticalCount: 1,
	}, nil)

	// Mock PublishAlertActivity
	env.OnActivity(ratelimit.PublishAlertActivity, mock.Anything, mock.Anything).Return(nil)

	env.OnActivity(ratelimit.PublishMetricsActivity, mock.Anything, mock.Anything).Return(nil)

	// Execute workflow
	env.ExecuteWorkflow(ratelimit.MonitorRateLimitsWorkflow)

	// Assert
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	// Verify PublishAlertActivity was called
	env.AssertCalled(s.T(), ratelimit.PublishAlertActivity, mock.Anything, mock.Anything)
}

func (s *MonitorWorkflowTestSuite) Test_MonitorRateLimitsWorkflow_BridgeFailure_Retry() {
	env := s.NewTestWorkflowEnvironment()

	// Mock GetPoliciesActivity failure then success
	callCount := 0
	env.OnActivity(ratelimit.GetPoliciesActivity, mock.Anything).Return(func() (*ratelimit.GetPoliciesResult, error) {
		callCount++
		if callCount == 1 {
			return nil, fmt.Errorf("bridge unavailable")
		}
		return &ratelimit.GetPoliciesResult{
			Policies:  []ratelimit.PolicyState{},
			CheckedAt: time.Now(),
		}, nil
	})

	env.OnActivity(ratelimit.StorePolicyStateActivity, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity(ratelimit.AnalyzeBalanceActivity, mock.Anything, mock.Anything).Return(&ratelimit.AnalyzeBalanceResult{}, nil)
	env.OnActivity(ratelimit.PublishMetricsActivity, mock.Anything, mock.Anything).Return(nil)

	// Execute workflow
	env.ExecuteWorkflow(ratelimit.MonitorRateLimitsWorkflow)

	// Assert - workflow should succeed after retry
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	s.Equal(2, callCount) // Activity called twice (1 failure + 1 success)
}

func (s *MonitorWorkflowTestSuite) Test_MonitorRateLimitsWorkflow_ContinueAsNew() {
	env := s.NewTestWorkflowEnvironment()

	// Set workflow start time to 25 hours ago (trigger Continue-As-New)
	env.SetStartTime(time.Now().Add(-25 * time.Hour))

	// Execute workflow
	env.ExecuteWorkflow(ratelimit.MonitorRateLimitsWorkflow)

	// Assert - should have Continue-As-New error
	s.True(env.IsWorkflowCompleted())
	err := env.GetWorkflowError()
	s.Error(err)
	s.Contains(err.Error(), "ContinueAsNew")
}
```

---

## 5. E2E Tests

### Complete System E2E Test

```go
// Location: tests/e2e/rate_limit_monitoring_test.go
package e2e_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_RateLimitMonitoring_FullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup: Ensure all services are running
	// - PostgreSQL
	// - Redis
	// - Pulsar
	// - Bridge (mock or real)
	// - Dict API
	// - Orchestration Worker

	ctx := context.Background()

	// Step 1: Verify Dict API is healthy
	resp, err := http.Get("http://localhost:8080/health")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Step 2: List policies (should trigger Bridge call)
	resp, err = http.Get("http://localhost:8080/api/v1/rate-limit/policies")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var policies struct {
		Policies []map[string]interface{} `json:"policies"`
		Total    int                      `json:"total"`
		Cached   bool                     `json:"cached"`
	}
	err = json.NewDecoder(resp.Body).Decode(&policies)
	require.NoError(t, err)
	assert.Equal(t, 24, policies.Total)
	assert.False(t, policies.Cached) // First call should be cache miss

	// Step 3: List policies again (should hit cache)
	resp, err = http.Get("http://localhost:8080/api/v1/rate-limit/policies")
	require.NoError(t, err)
	err = json.NewDecoder(resp.Body).Decode(&policies)
	require.NoError(t, err)
	assert.True(t, policies.Cached) // Second call should be cache hit

	// Step 4: Wait for Temporal workflow to execute (next cron run)
	time.Sleep(6 * time.Minute) // Wait for one cron cycle (5 min + buffer)

	// Step 5: Verify database has policy states
	// (connect to PostgreSQL and query dict_rate_limit_states)

	// Step 6: Verify metrics are exported
	resp, err = http.Get("http://localhost:8080/metrics")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse Prometheus metrics
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "dict_rate_limit_available_tokens")
	assert.Contains(t, string(body), "dict_rate_limit_utilization_pct")
}
```

---

## 6. Performance Tests

### Load Testing (k6)

```javascript
// Location: tests/performance/rate_limit_load_test.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '1m', target: 100 },  // Ramp up to 100 users
    { duration: '3m', target: 100 },  // Stay at 100 users
    { duration: '1m', target: 500 },  // Ramp up to 500 users
    { duration: '3m', target: 500 },  // Stay at 500 users
    { duration: '1m', target: 1000 }, // Spike to 1000 users
    { duration: '2m', target: 1000 }, // Stay at 1000 users
    { duration: '1m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(99)<200'], // 99% of requests must complete within 200ms
    http_req_failed: ['rate<0.01'],   // Error rate must be < 1%
  },
};

export default function () {
  // Test GET /api/v1/rate-limit/policies
  let response = http.get('http://localhost:8080/api/v1/rate-limit/policies');

  check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
    'has policies': (r) => JSON.parse(r.body).total === 24,
  });

  sleep(1);
}
```

### Benchmark Tests (Go)

```go
// Location: apps/dict/handlers/http/ratelimit/handler_bench_test.go
package ratelimit_test

import (
	"context"
	"testing"

	"github.com/lb-conn/connector-dict/apps/dict/handlers/http/ratelimit"
)

func BenchmarkHandler_ListPolicies_CacheHit(b *testing.B) {
	// Setup
	mockService := new(MockRateLimitService)
	handler := ratelimit.NewHandler(mockService, nil)

	policies := make([]ratelimit.PolicySummary, 24)
	mockService.On("ListPolicies", mock.Anything, (*string)(nil), (*float64)(nil)).
		Return(policies, true, 60*time.Second, nil)

	// Benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.ListPolicies(context.Background(), &ratelimit.ListPoliciesRequest{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRepository_BatchInsert_1000Rows(b *testing.B) {
	// Setup PostgreSQL container
	pool, cleanup := setupPostgresContainer(b)
	defer cleanup()

	repo := repositories.NewRateLimitStateRepository(pool)

	// Generate 1000 policy states
	policies := make([]ratelimit.PolicyState, 1000)
	for i := 0; i < 1000; i++ {
		policies[i] = ratelimit.PolicyState{
			PolicyName:      fmt.Sprintf("POLICY_%d", i),
			AvailableTokens: 100,
			CapacityMax:     200,
			CheckedAt:       time.Now(),
		}
	}

	// Benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := repo.BatchInsert(context.Background(), policies)
		if err != nil {
			b.Fatal(err)
		}
	}
}
```

---

## 7. Contract Tests

### gRPC Contract Tests

```go
// Location: tests/contract/bridge_grpc_test.go
package contract_test

import (
	"context"
	"testing"

	pb "github.com/lb-conn/rsfn-connect-bacen-bridge/proto/ratelimit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestBridgeContract_ListPoliciesResponse(t *testing.T) {
	// Create sample response matching contract
	resp := &pb.ListPoliciesResponse{
		Policies: []*pb.Policy{
			{
				PolicyName:           "ENTRIES_CREATE",
				Category:             "A",
				Capacity:             300,
				RefillTokens:         5,
				RefillPeriodSec:      60,
				AvailableTokens:      150,
				WarningThresholdPct:  25.0,
				CriticalThresholdPct: 10.0,
				CheckedAt:            timestamppb.Now(),
			},
		},
		CheckedAt: timestamppb.Now(),
	}

	// Verify proto can be marshaled/unmarshaled
	data, err := proto.Marshal(resp)
	require.NoError(t, err)

	var decoded pb.ListPoliciesResponse
	err = proto.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Assert contract fields
	assert.Len(t, decoded.Policies, 1)
	assert.Equal(t, "ENTRIES_CREATE", decoded.Policies[0].PolicyName)
	assert.Equal(t, int32(300), decoded.Policies[0].Capacity)
}
```

### Pulsar Event Contract Tests

```go
// Location: tests/contract/pulsar_event_test.go
package contract_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

func TestPulsarContract_ActionRateLimitAlert(t *testing.T) {
	// Load JSON schema
	schemaLoader := gojsonschema.NewReferenceLoader("file://./schemas/action_rate_limit_alert.json")

	// Sample event payload
	event := map[string]interface{}{
		"action": "ActionRateLimitAlert",
		"data": map[string]interface{}{
			"policy_name":      "ENTRIES_CREATE",
			"category":         "A",
			"severity":         "CRITICAL",
			"available_tokens": 30,
			"capacity_max":     300,
			"utilization_pct":  90.0,
			"message":          "Test alert",
			"detected_at":      "2025-10-31T10:30:00Z",
		},
		"metadata": map[string]interface{}{
			"source":     "dict-rate-limit-monitoring",
			"version":    "1.0.0",
			"created_at": "2025-10-31T10:30:00Z",
		},
	}

	// Marshal to JSON
	eventJSON, err := json.Marshal(event)
	require.NoError(t, err)

	// Validate against schema
	documentLoader := gojsonschema.NewStringLoader(string(eventJSON))
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	require.NoError(t, err)

	// Assert validation passed
	assert.True(t, result.Valid(), "Event should match JSON schema")
	if !result.Valid() {
		for _, err := range result.Errors() {
			t.Logf("Schema validation error: %s", err)
		}
	}
}
```

---

## 8. Test Data & Fixtures

### Test Fixtures

```go
// Location: tests/fixtures/policies.go
package fixtures

import (
	"time"

	"github.com/lb-conn/connector-dict/apps/orchestration-worker/domain/ratelimit"
)

// PolicyStates retorna dados de teste de políticas
func PolicyStates() []ratelimit.PolicyState {
	now := time.Now()

	return []ratelimit.PolicyState{
		{
			PolicyName:           "ENTRIES_CREATE",
			Category:             "A",
			CapacityMax:          300,
			RefillTokens:         5,
			RefillPeriodSec:      60,
			AvailableTokens:      150,
			UtilizationPct:       50.0,
			WarningThresholdPct:  25.0,
			CriticalThresholdPct: 10.0,
			Status:               "OK",
			CheckedAt:            now,
		},
		{
			PolicyName:           "CLAIMS_CREATE",
			Category:             "B",
			CapacityMax:          1000,
			RefillTokens:         300,
			RefillPeriodSec:      60,
			AvailableTokens:      50,
			UtilizationPct:       95.0,
			WarningThresholdPct:  25.0,
			CriticalThresholdPct: 10.0,
			Status:               "CRITICAL",
			CheckedAt:            now,
		},
		// ... (22 more policies for complete BACEN set)
	}
}

// AlertEvents retorna dados de teste de alertas
func AlertEvents() []ratelimit.AlertEvent {
	return []ratelimit.AlertEvent{
		{
			PolicyName:      "CLAIMS_CREATE",
			Category:        "B",
			Severity:        "CRITICAL",
			AvailableTokens: 50,
			CapacityMax:     1000,
			UtilizationPct:  95.0,
			Message:         "URGENT: Policy CLAIMS_CREATE at 95% utilization",
			DetectedAt:      time.Now(),
		},
	}
}
```

---

## 9. CI/CD Integration

### GitHub Actions Workflow

```yaml
# Location: .github/workflows/test.yml
name: Test Suite

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Run unit tests
        run: |
          go test ./... -short -race -coverprofile=coverage.txt -covermode=atomic

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.txt

  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Run integration tests
        run: |
          go test ./... -run Integration -race -coverprofile=coverage-integration.txt

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Start services (docker-compose)
        run: |
          docker-compose up -d

      - name: Run E2E tests
        run: |
          go test ./tests/e2e/... -timeout 10m

      - name: Stop services
        run: |
          docker-compose down
```

### Test Coverage Report

```bash
# Generate HTML coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# View coverage summary
go tool cover -func=coverage.out

# Expected output:
# github.com/lb-conn/connector-dict/apps/dict/handlers/http/ratelimit/handler.go:           88.5%
# github.com/lb-conn/connector-dict/apps/dict/application/usecases/ratelimit/service.go:   92.3%
# github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/workflows/ratelimit/monitor_workflow.go: 90.1%
# total:                                                                                    87.2%
```

---

**Última Atualização**: 2025-10-31
**Versão**: 1.0.0
**Status**: ✅ ESPECIFICAÇÃO COMPLETA - Production-Ready
