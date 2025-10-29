---
description: Execute comprehensive test suite with coverage analysis
---

Execute the complete test suite for CID/VSync implementation.

## Test Execution Plan

### 1. Unit Tests
```bash
cd connector-dict/apps/orchestration-worker

# Run all unit tests with race detector
go test -v -race ./internal/domain/... ./internal/application/...

# With coverage
go test -v -race -coverprofile=coverage-unit.out ./internal/domain/... ./internal/application/...
```

### 2. Integration Tests
```bash
# Start test dependencies (PostgreSQL, Redis)
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test -v -race -tags=integration ./tests/integration/...

# With coverage
go test -v -race -tags=integration -coverprofile=coverage-integration.out ./tests/integration/...

# Cleanup
docker-compose -f docker-compose.test.yml down -v
```

### 3. Temporal Workflow Tests
```bash
# Workflow tests (with mocked activities)
go test -v ./internal/infrastructure/temporal/workflows/...

# Activity tests
go test -v ./internal/infrastructure/temporal/activities/...
```

### 4. E2E Tests
```bash
# Full end-to-end tests (requires all services)
go test -v -tags=e2e ./tests/e2e/...
```

### 5. Coverage Analysis
```bash
# Merge all coverage profiles
go tool covdata merge \
  -i coverage-unit.out,coverage-integration.out \
  -o coverage-total.out

# Generate HTML report
go tool cover -html=coverage-total.out -o coverage-report.html

# Check coverage threshold (must be >80%)
coverage=$(go tool cover -func=coverage-total.out | grep total | awk '{print $3}' | sed 's/%//')
if (( $(echo "$coverage < 80" | bc -l) )); then
  echo "❌ Coverage $coverage% is below 80%"
  exit 1
fi
echo "✅ Coverage $coverage% meets threshold"
```

### 6. Lint & Static Analysis
```bash
# golangci-lint
golangci-lint run --timeout 5m ./...

# Go vet
go vet ./...

# Security scan
gosec -exclude=G104 ./...
```

## Expected Output

All tests should pass with output similar to:
```
=== RUN   TestCID_Generate
=== RUN   TestCID_Generate/CPF_key_with_all_fields
=== RUN   TestCID_Generate/missing_required_field
--- PASS: TestCID_Generate (0.01s)
    --- PASS: TestCID_Generate/CPF_key_with_all_fields (0.00s)
    --- PASS: TestCID_Generate/missing_required_field (0.00s)
PASS
coverage: 92.5% of statements
ok      github.com/lb-conn/connector-dict/internal/domain/cid  0.123s  coverage: 92.5% of statements
```

## Troubleshooting

### Tests failing with "connection refused"
```bash
# Check if test dependencies are running
docker-compose -f docker-compose.test.yml ps

# Check logs
docker-compose -f docker-compose.test.yml logs postgres
```

### Coverage below threshold
```bash
# Identify files with low coverage
go tool cover -func=coverage-total.out | grep -v 100.0% | sort -k3 -n

# Focus on files with <80% coverage
```

### Flaky tests
```bash
# Run tests multiple times to identify flaky tests
go test -count=10 -v ./...
```

## CI/CD Integration

This command is automatically run in GitHub Actions pipeline:
```yaml
# .github/workflows/test.yml
- name: Run tests
  run: |
    make test
    make test-integration
    make coverage-check
```
