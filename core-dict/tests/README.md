# Core-Dict Tests

This directory contains comprehensive integration and end-to-end tests for the Core-Dict service.

## Test Structure

```
tests/
├── integration/              # Integration tests (35 tests)
│   ├── entry_lifecycle_test.go      # 10 tests - Entry CRUD operations
│   ├── claim_workflow_test.go       # 12 tests - Claim lifecycle
│   ├── database_test.go             # 8 tests - PostgreSQL features
│   └── cache_test.go                # 5 tests - Redis caching
│
├── e2e/                      # End-to-end tests (15 tests)
│   ├── create_entry_e2e_test.go              # 5 tests - Entry creation E2E
│   ├── claim_workflow_e2e_test.go            # 5 tests - Claim workflow E2E
│   ├── integration_connect_bridge_test.go    # 3 tests - Full stack integration
│   └── performance_test.go                   # 2 tests - Performance benchmarks
│
├── testhelpers/              # Test utilities
│   ├── test_environment.go   # Integration test setup
│   ├── e2e_environment.go    # E2E test setup
│   ├── pulsar_mock.go        # Pulsar simulator
│   ├── connect_mock.go       # Connect gRPC mock
│   └── fixtures.go           # Test data fixtures
│
└── mocks/                    # Mock configurations
    └── bacen-expectations.json   # Bacen mock responses
```

## Total Tests: 50

### Integration Tests (35 tests)
- **Entry Lifecycle**: 10 tests
  - Create, Read, Update, Delete
  - Duplicate check
  - Cache invalidation
  - Soft delete
  - Block/Unblock
  - Ownership transfer
  - Max keys validation

- **Claim Workflow**: 12 tests
  - Ownership claims
  - Portability claims
  - Auto-confirm after 30 days
  - Cancel claim
  - Complete claim
  - Filter by status
  - gRPC Connect integration

- **Database**: 8 tests
  - RLS (Row Level Security)
  - Partitioning by month
  - Transactions and rollback
  - Indexes and performance
  - Migrations up/down
  - Constraint violations
  - Soft delete behavior
  - Audit log tracking

- **Cache**: 5 tests
  - Cache-Aside pattern
  - Write-Through pattern
  - Rate limiter (100 RPS)
  - Pattern-based invalidation
  - TTL expiration

### E2E Tests (15 tests)
- **Create Entry E2E**: 5 tests
  - CPF entry with Bacen simulation
  - EVP generation
  - Global duplicate check (Core → Connect → Bridge → Bacen)
  - Max keys validation
  - LGPD compliance (SHA256 hashing)

- **Claim Workflow E2E**: 5 tests
  - Ownership claim complete 30-day flow
  - Portability donor to recipient
  - Auto-confirm simulation
  - Cancel before confirm
  - Full gRPC stack (Core → Connect/Temporal → Bridge → Bacen)

- **Connect/Bridge Integration**: 3 tests
  - Core → Connect → Bridge → Bacen SOAP flow
  - VSYNC workflow via Temporal
  - Pulsar event propagation end-to-end

- **Performance**: 2 tests
  - 1000 TPS sustained load
  - 100 concurrent claims

## Running Tests

### Prerequisites

```bash
# Install dependencies
go mod download

# Install testcontainers (for integration tests)
go get github.com/testcontainers/testcontainers-go@latest
```

### Unit Tests
```bash
make test-unit
```

### Integration Tests
```bash
# Run all integration tests (uses testcontainers)
make test-integration

# Run specific integration test
go test -v ./tests/integration -run TestIntegration_CreateEntry_CompleteFlow
```

### E2E Tests
```bash
# Start services via docker-compose
make test-e2e-setup

# Run E2E tests
make test-e2e

# Run specific E2E test
go test -v ./tests/e2e -run TestE2E_CreateEntry_CPF_Success

# Cleanup
make test-e2e-teardown
```

### Performance Tests
```bash
# Run performance tests (longer duration)
make test-performance

# Or run individually
go test -v ./tests/e2e -run TestE2E_Performance_CreateEntry_1000TPS
```

### All Tests
```bash
# Run all tests with coverage
make test-all

# Generate coverage report
make test-coverage
```

## Test Coverage Goals

- **Target**: 80%+ overall coverage
- **Integration Tests**: Cover all critical business logic
- **E2E Tests**: Cover all major user flows
- **Performance Tests**: Validate SLAs (1000 TPS, <100ms latency)

## Test Configuration

### Environment Variables

Integration tests (via testcontainers):
```bash
# No manual configuration needed
# Testcontainers automatically starts PostgreSQL, Redis, etc.
```

E2E tests (requires services running):
```bash
export CORE_DICT_URL=http://localhost:8080
export CONN_DICT_URL=http://localhost:8081
export CONN_BRIDGE_URL=http://localhost:8082
```

### Docker Compose

E2E tests use `docker-compose.test.yml`:
```bash
docker-compose -f docker-compose.test.yml up -d
```

This starts:
- PostgreSQL
- Redis
- Pulsar
- Temporal
- Core-Dict
- Conn-Dict
- Conn-Bridge
- Bacen Mock

## Continuous Integration

GitHub Actions workflow (`.github/workflows/test.yml`):
```yaml
name: Tests

on: [push, pull_request]

jobs:
  integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24.5'
      - run: make test-integration

  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24.5'
      - run: make test-e2e-ci
```

## Debugging Tests

### Verbose output
```bash
go test -v ./tests/...
```

### Run single test
```bash
go test -v ./tests/integration -run TestIntegration_CreateEntry_CompleteFlow
```

### Keep containers running on failure (integration tests)
```bash
# Manually inspect testcontainers
docker ps -a | grep testcontainers
```

### View E2E logs
```bash
docker-compose -f docker-compose.test.yml logs -f core-dict
```

## Test Patterns

### Integration Test Pattern
```go
func TestIntegration_FeatureName(t *testing.T) {
    env := testhelpers.SetupIntegrationTest(t)
    defer env.CleanAll()

    // Arrange
    // ... setup test data

    // Act
    // ... execute operation

    // Assert
    // ... verify results
}
```

### E2E Test Pattern
```go
func TestE2E_FeatureName(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    env := testhelpers.SetupE2ETest(t)

    // Arrange
    // Act
    // Assert (including cross-service verification)
}
```

## Performance Benchmarks

Target SLAs:
- **Throughput**: 1000 TPS (sustained)
- **Latency**:
  - P50: <50ms
  - P95: <100ms
  - P99: <200ms
- **Error Rate**: <1%
- **Concurrent Operations**: 100+ parallel claims

## Troubleshooting

### Integration tests fail with "port already in use"
```bash
# Kill existing containers
docker kill $(docker ps -q)
```

### E2E tests timeout
```bash
# Increase health check timeout in docker-compose.test.yml
# Check service logs for startup errors
docker-compose -f docker-compose.test.yml logs
```

### Performance tests don't achieve target TPS
- Check system resources (CPU, memory)
- Verify database connection pool settings
- Check for rate limiting in services
- Review logs for errors

## Contributing

When adding new tests:
1. Follow existing test patterns
2. Use testhelpers for common setup
3. Clean up resources in test cleanup
4. Add test to appropriate category (integration/e2e)
5. Update this README with new test counts
6. Ensure tests pass in CI

## Test Metrics

After running tests, metrics are available:
```bash
# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Benchmark results
go test -bench=. -benchmem ./tests/...
```

## License

Copyright (c) 2025 LBPay Lab
