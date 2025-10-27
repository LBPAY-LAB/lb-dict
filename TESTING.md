# Testing Guide - DICT LBPay Project

This guide describes the testing framework and practices for the DICT LBPay project.

## Table of Contents

- [Overview](#overview)
- [Test Framework](#test-framework)
- [Test Structure](#test-structure)
- [Running Tests](#running-tests)
- [Writing Tests](#writing-tests)
- [Coverage](#coverage)
- [Best Practices](#best-practices)
- [CI/CD Integration](#cicd-integration)

## Overview

The DICT LBPay project uses a comprehensive testing strategy with:
- **testify** for assertions and test utilities
- **testcontainers-go** for integration tests with real dependencies
- **Temporal test suite** for workflow/activity testing
- Target: >80% test coverage across all repositories

## Test Framework

### Dependencies

All repositories use:
```go
github.com/stretchr/testify v1.11.1
github.com/testcontainers/testcontainers-go v0.27.0
```

### Test Types

1. **Unit Tests**: Test individual components in isolation
2. **Integration Tests**: Test component interactions with real dependencies
3. **Workflow Tests**: Test Temporal workflows and activities (conn-dict only)

## Test Structure

### Directory Layout

```
repo/
├── internal/
│   └── domain/
│       ├── entities.go
│       └── entities_test.go          # Unit tests
├── tests/
│   ├── integration/
│   │   ├── grpc_test.go             # Integration tests
│   │   └── database_test.go
│   └── helpers/
│       ├── fixtures.go               # Test data
│       └── mocks.go                  # Mock implementations
```

### Test Files

#### conn-bridge
- `internal/domain/entities/dict_entry_test.go` - Entity validation tests
- `internal/application/usecases/create_entry_test.go` - Use case tests with mocks
- `internal/infrastructure/bacen/http_client_test.go` - HTTP client tests
- `tests/integration/bridge_grpc_test.go` - gRPC integration tests
- `tests/helpers/bacen_mock.go` - Mock Bacen API server
- `tests/helpers/test_helpers.go` - Common test utilities

#### conn-dict
- `internal/domain/aggregates/claim_test.go` - Aggregate logic tests
- `internal/activities/claim_activities_test.go` - Temporal activity tests
- `tests/integration/temporal_test.go` - Temporal workflow integration tests
- `tests/helpers/temporal_test_env.go` - Temporal test environment setup
- `tests/helpers/test_helpers.go` - Database and Redis test containers

#### dict-contracts
- `tests/proto_validation_test.go` - Protobuf definition validation tests

## Running Tests

### All Tests

```bash
make test
```

### Unit Tests Only

```bash
make test-unit
```

### Integration Tests

```bash
make test-integration
```

### Coverage Report

```bash
make coverage
```

This generates:
- `coverage.out` - Coverage data
- `coverage.html` - HTML coverage report
- Terminal output with total coverage percentage

### Repository-Specific Commands

#### conn-bridge
```bash
cd conn-bridge
make test              # All tests with coverage
make test-unit         # Unit tests only (fast)
make test-integration  # Integration tests
make coverage          # Generate coverage report
```

#### conn-dict
```bash
cd conn-dict
make test              # All tests with coverage
make test-unit         # Unit tests only (fast)
make test-integration  # Integration tests
make coverage          # Generate coverage report
```

#### dict-contracts
```bash
cd dict-contracts
make test              # All tests
make test-unit         # Unit tests only
make coverage          # Coverage report
```

## Writing Tests

### Unit Test Example

```go
package entities

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDictEntry_Validate(t *testing.T) {
	tests := []struct {
		name    string
		entry   *DictEntry
		wantErr bool
	}{
		{
			name: "valid entry",
			entry: &DictEntry{
				Key:    "test@example.com",
				Type:   KeyTypeEmail,
				Status: StatusActive,
			},
			wantErr: false,
		},
		{
			name:    "invalid entry",
			entry:   &DictEntry{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entry.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
```

### Mock Example

```go
type MockBacenClient struct {
	mock.Mock
}

func (m *MockBacenClient) SendRequest(ctx context.Context, req *Request) (*Response, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Response), args.Error(1)
}

func TestUseCase_WithMock(t *testing.T) {
	mockClient := new(MockBacenClient)
	mockClient.On("SendRequest", mock.Anything, mock.Anything).Return(&Response{}, nil)

	uc := NewUseCase(mockClient)
	// ... test logic

	mockClient.AssertExpectations(t)
}
```

### Integration Test Example

```go
package integration

import (
	"testing"
	"github.com/lbpay-lab/conn-bridge/tests/helpers"
)

func TestDatabaseIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test database with testcontainers
	db := helpers.SetupTestDB(t)

	// Test logic here

	// Cleanup happens automatically via t.Cleanup()
}
```

### Temporal Test Example

```go
package workflows

import (
	"testing"
	"go.temporal.io/sdk/testsuite"
)

func TestClaimWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(ClaimWorkflow)
	env.RegisterActivity(CreateClaimActivity)

	env.ExecuteWorkflow(ClaimWorkflow, input)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
```

## Coverage

### Coverage Goals

- Overall: >80% coverage
- Critical paths: >90% coverage
- New features: Must include tests

### Checking Coverage

```bash
# Generate coverage report
make coverage

# View in browser
open coverage.html

# Command line summary
go tool cover -func=coverage.out
```

### Coverage by Package

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -E '^total:'
```

## Best Practices

### 1. Table-Driven Tests

Use table-driven tests for multiple scenarios:

```go
tests := []struct {
	name    string
	input   Input
	want    Output
	wantErr bool
}{
	{"scenario 1", input1, output1, false},
	{"scenario 2", input2, output2, true},
}

for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		// test logic
	})
}
```

### 2. Parallel Tests

Run independent tests in parallel:

```go
func TestSomething(t *testing.T) {
	t.Parallel()
	// test logic
}
```

### 3. Use Subtests

Organize related tests:

```go
func TestFeature(t *testing.T) {
	t.Run("success case", func(t *testing.T) { /* ... */ })
	t.Run("error case", func(t *testing.T) { /* ... */ })
}
```

### 4. Test Helpers

Use helpers from `tests/helpers/`:

```go
// Create test context with timeout
ctx := helpers.CreateTestContext(t)

// Load fixtures
entry := helpers.LoadFixture(t, "valid_cpf_entry")

// Setup test infrastructure
db := helpers.SetupTestDB(t)
redis := helpers.SetupTestRedis(t)
```

### 5. Mock External Dependencies

Always mock external services:
- Bacen API
- Temporal server (use test suite for local tests)
- Pulsar (mock producer/consumer)

### 6. Clean Up Resources

Use `t.Cleanup()`:

```go
func TestSomething(t *testing.T) {
	resource := setupResource()
	t.Cleanup(func() {
		resource.Close()
	})
}
```

### 7. Skip Long Tests

Use `-short` flag for quick tests:

```go
func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	// integration test logic
}
```

## CI/CD Integration

### GitHub Actions

Tests run automatically on:
- Pull requests
- Pushes to main branch
- Manual workflow dispatch

### Workflow Configuration

See `.github/workflows/ci.yml` in each repository.

Example:
```yaml
- name: Run tests
  run: make test

- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    files: ./coverage.out
```

### Local Pre-commit

Run before committing:

```bash
# In each repo
make check  # Runs lint + test
```

## Troubleshooting

### Tests Fail with "Permission Denied"

Docker might not be running (needed for testcontainers):
```bash
# Start Docker
docker info
```

### Integration Tests Hang

Check if containers are running:
```bash
docker ps
docker logs <container-id>
```

### Coverage Too Low

Find uncovered code:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -v "100.0%"
```

## Resources

- [testify Documentation](https://github.com/stretchr/testify)
- [testcontainers-go Documentation](https://golang.testcontainers.org/)
- [Temporal Testing Guide](https://docs.temporal.io/develop/go/testing-suite)
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)

## Support

For questions or issues with tests:
1. Check this guide
2. Review existing tests in the codebase
3. Ask the team in Slack #dict-dev
