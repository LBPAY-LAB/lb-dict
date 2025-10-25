# IMP-004: Developer Guidelines

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: Development Standards and Best Practices
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: BACKEND (AI Agent - Backend Developer)

---

## Sumário Executivo

Este documento define os **padrões de desenvolvimento** para o projeto DICT LBPay, incluindo convenções de código Go, estrutura de projetos, padrões de erro, logging, testes e revisão de código.

**Baseado em**:
- [IMP-001: Manual de Implementação Core DICT](./IMP-001_Manual_Implementacao_Core_DICT.md)
- [Effective Go](https://go.dev/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | BACKEND | Versão inicial - Developer Guidelines |

---

## Índice

1. [Go Coding Standards](#1-go-coding-standards)
2. [Project Structure](#2-project-structure)
3. [Naming Conventions](#3-naming-conventions)
4. [Error Handling Patterns](#4-error-handling-patterns)
5. [Logging Standards](#5-logging-standards)
6. [Testing Practices](#6-testing-practices)
7. [Code Review Checklist](#7-code-review-checklist)
8. [Performance Guidelines](#8-performance-guidelines)
9. [Security Best Practices](#9-security-best-practices)

---

## 1. Go Coding Standards

### 1.1. Code Formatting

**Rule**: Use `gofmt` and `goimports` for all code formatting.

```bash
# Format code automatically
gofmt -w .

# Organize imports
goimports -w .
```

**Pre-commit Hook**:
```bash
#!/bin/bash
# .git/hooks/pre-commit

# Run gofmt
unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
    echo "These files are not formatted:"
    echo "$unformatted"
    exit 1
fi

# Run goimports
go install golang.org/x/tools/cmd/goimports@latest
goimports -w .

# Run go vet
go vet ./...
```

### 1.2. Line Length

**Rule**: Limit lines to **120 characters** maximum.

**Bad**:
```go
func CreateEntry(ctx context.Context, keyType string, keyValue string, accountNumber string, branchCode string, holderName string) error {
    // Too long
}
```

**Good**:
```go
func CreateEntry(
    ctx context.Context,
    keyType string,
    keyValue string,
    accountNumber string,
    branchCode string,
    holderName string,
) error {
    // Multi-line function signature
}
```

### 1.3. Import Organization

**Rule**: Organize imports in three groups:
1. Standard library
2. External packages
3. Internal packages

**Good**:
```go
import (
    // Standard library
    "context"
    "fmt"
    "time"

    // External packages
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "gorm.io/gorm"

    // Internal packages
    "github.com/lbpay/core-dict/internal/domain"
    "github.com/lbpay/core-dict/pkg/logger"
)
```

### 1.4. Variable Declaration

**Rule**: Use short variable declarations (`:=`) when possible.

**Bad**:
```go
var entry domain.Entry
entry = domain.Entry{ID: uuid.New()}
```

**Good**:
```go
entry := domain.Entry{ID: uuid.New()}
```

**Exception**: Use `var` for zero values:
```go
var count int  // Explicitly zero
var exists bool
```

### 1.5. Error Handling

**Rule**: Always check errors. Never ignore them with `_`.

**Bad**:
```go
result, _ := repository.FindByID(id)  // Never ignore errors
```

**Good**:
```go
result, err := repository.FindByID(id)
if err != nil {
    return fmt.Errorf("failed to find entry: %w", err)
}
```

### 1.6. Context Usage

**Rule**: Always pass `context.Context` as the first parameter.

**Bad**:
```go
func GetEntry(id string, ctx context.Context) (*Entry, error)
```

**Good**:
```go
func GetEntry(ctx context.Context, id string) (*Entry, error)
```

### 1.7. Struct Initialization

**Rule**: Use named fields for struct initialization.

**Bad**:
```go
entry := Entry{uuid.New(), "CPF", "12345678901", time.Now()}
```

**Good**:
```go
entry := Entry{
    ID:        uuid.New(),
    KeyType:   "CPF",
    KeyValue:  "12345678901",
    CreatedAt: time.Now(),
}
```

---

## 2. Project Structure

### 2.1. Hexagonal Architecture (Ports & Adapters)

```
core-dict/
├── cmd/
│   └── api/
│       └── main.go                    # Application entrypoint
├── internal/
│   ├── domain/                        # Domain layer (business logic)
│   │   ├── entry.go                   # Domain models
│   │   ├── account.go
│   │   ├── claim.go
│   │   └── errors.go                  # Domain errors
│   ├── application/                   # Application layer (use cases)
│   │   ├── entry/
│   │   │   ├── create_entry.go
│   │   │   ├── get_entry.go
│   │   │   └── list_entries.go
│   │   └── claim/
│   │       ├── create_claim.go
│   │       └── confirm_claim.go
│   ├── infrastructure/                # Infrastructure layer (external adapters)
│   │   ├── database/
│   │   │   ├── postgres.go
│   │   │   ├── entry_repository.go
│   │   │   └── account_repository.go
│   │   ├── fiber/                     # HTTP adapter
│   │   │   ├── server.go
│   │   │   ├── routes.go
│   │   │   └── handlers/
│   │   │       ├── entry_handler.go
│   │   │       └── claim_handler.go
│   │   └── pulsar/                    # Messaging adapter
│   │       ├── producer.go
│   │       └── consumer.go
│   └── ports/                         # Ports (interfaces)
│       ├── repositories.go            # Repository interfaces
│       ├── messaging.go               # Messaging interfaces
│       └── logger.go                  # Logger interface
├── pkg/                               # Public packages (reusable)
│   ├── logger/
│   │   └── logger.go
│   └── validator/
│       └── validator.go
├── db/
│   └── migrations/                    # Database migrations
├── config/
│   └── config.yaml                    # Configuration
└── tests/
    ├── integration/                   # Integration tests
    └── e2e/                           # End-to-end tests
```

### 2.2. Layer Responsibilities

| Layer | Responsibility | Imports |
|-------|---------------|---------|
| **Domain** | Business logic, models, validation | No external dependencies |
| **Application** | Use cases, orchestration | Domain + Ports |
| **Infrastructure** | External systems (DB, HTTP, messaging) | Domain + Application + Ports + External libs |
| **Ports** | Interfaces (contracts) | Domain only |

**Dependency Rule**: Inner layers cannot depend on outer layers.

```
Infrastructure → Application → Domain
       ↓              ↓
     Ports  ←  ←  ←  ←
```

---

## 3. Naming Conventions

### 3.1. Packages

**Rule**: Package names should be **lowercase, single-word**, without underscores.

**Bad**:
```go
package entry_repository
package EntryRepository
```

**Good**:
```go
package repository
package domain
```

### 3.2. Files

**Rule**: Use **snake_case** for file names.

**Bad**:
```go
EntryRepository.go
entryRepository.go
```

**Good**:
```go
entry_repository.go
create_entry_use_case.go
```

### 3.3. Variables and Functions

**Rule**: Use **camelCase** for unexported, **PascalCase** for exported.

**Good**:
```go
// Exported (public)
type Entry struct {}
func CreateEntry() {}

// Unexported (private)
var entryCache = make(map[string]*Entry)
func validateEntry() {}
```

### 3.4. Constants

**Rule**: Use **PascalCase** for exported constants, **camelCase** for unexported.

**Good**:
```go
// Exported
const MaxEntriesPerUser = 20
const DefaultTimeout = 30 * time.Second

// Unexported
const defaultPageSize = 20
```

### 3.5. Interfaces

**Rule**: Interface names should describe behavior, ending in `-er` when possible.

**Bad**:
```go
type EntryInterface interface {}
type IEntry interface {}
```

**Good**:
```go
type EntryRepository interface {}
type Logger interface {}
type Validator interface {}
```

### 3.6. Test Functions

**Rule**: Use `Test` prefix + function name + scenario.

**Good**:
```go
func TestCreateEntry_ValidInput_Success(t *testing.T) {}
func TestCreateEntry_DuplicateKey_ReturnsError(t *testing.T) {}
func TestGetEntry_NotFound_ReturnsError(t *testing.T) {}
```

---

## 4. Error Handling Patterns

### 4.1. Domain Errors

**Create custom domain errors**:

```go
// internal/domain/errors.go
package domain

import "errors"

var (
    ErrEntryNotFound      = errors.New("entry not found")
    ErrKeyAlreadyExists   = errors.New("key already exists")
    ErrInvalidKeyFormat   = errors.New("invalid key format")
    ErrMaxKeysExceeded    = errors.New("max keys exceeded")
    ErrClaimNotFound      = errors.New("claim not found")
    ErrClaimAlreadyExists = errors.New("claim already exists")
)

// Custom error with context
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}
```

### 4.2. Error Wrapping

**Rule**: Always wrap errors with context using `fmt.Errorf` and `%w`.

**Bad**:
```go
if err != nil {
    return err  // Lost context
}
```

**Good**:
```go
if err != nil {
    return fmt.Errorf("failed to create entry: %w", err)
}
```

### 4.3. Error Checking

**Use `errors.Is()` and `errors.As()`**:

```go
// Check error type
if errors.Is(err, domain.ErrEntryNotFound) {
    return fiber.NewError(fiber.StatusNotFound, "Entry not found")
}

// Extract custom error
var validationErr *domain.ValidationError
if errors.As(err, &validationErr) {
    return fiber.NewError(fiber.StatusBadRequest, validationErr.Error())
}
```

### 4.4. Application Layer Error Handling

```go
// internal/application/entry/create_entry.go
package entry

import (
    "context"
    "fmt"
    "github.com/lbpay/core-dict/internal/domain"
)

type CreateEntryUseCase struct {
    repository domain.EntryRepository
    logger     domain.Logger
}

func (uc *CreateEntryUseCase) Execute(ctx context.Context, input CreateEntryInput) (*domain.Entry, error) {
    // Validate input
    if err := input.Validate(); err != nil {
        uc.logger.Warnf("Invalid input: %v", err)
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    // Check if key already exists
    exists, err := uc.repository.ExistsByKey(ctx, input.KeyType, input.KeyValue)
    if err != nil {
        uc.logger.Errorf("Failed to check key existence: %v", err)
        return nil, fmt.Errorf("failed to check key existence: %w", err)
    }
    if exists {
        return nil, domain.ErrKeyAlreadyExists
    }

    // Create entry
    entry := domain.NewEntry(input.KeyType, input.KeyValue, input.AccountID)
    if err := uc.repository.Create(ctx, entry); err != nil {
        uc.logger.Errorf("Failed to create entry: %v", err)
        return nil, fmt.Errorf("failed to create entry: %w", err)
    }

    uc.logger.Infof("Entry created: %s", entry.ID)
    return entry, nil
}
```

---

## 5. Logging Standards

### 5.1. Log Levels

| Level | Usage | Example |
|-------|-------|---------|
| **DEBUG** | Detailed debugging info | `logger.Debugf("Entry validated: %+v", entry)` |
| **INFO** | General informational | `logger.Infof("Entry created: %s", entry.ID)` |
| **WARN** | Warning, but not error | `logger.Warnf("Retry attempt %d failed", attempt)` |
| **ERROR** | Error occurred | `logger.Errorf("Failed to connect to DB: %v", err)` |
| **FATAL** | Fatal error, app exits | `logger.Fatalf("Cannot start server: %v", err)` |

### 5.2. Structured Logging

**Use structured logging with fields**:

```go
import "github.com/sirupsen/logrus"

logger.WithFields(logrus.Fields{
    "entry_id":   entry.ID,
    "key_type":   entry.KeyType,
    "key_value":  entry.KeyValue,
    "user_id":    userID,
    "request_id": requestID,
}).Info("Entry created successfully")
```

### 5.3. Logging Best Practices

**DO**:
- Always log errors with context
- Log important business events (entry created, claim confirmed)
- Log request start/end with duration
- Use trace IDs for distributed tracing

**DON'T**:
- Log sensitive data (passwords, tokens, full CPF/CNPJ)
- Log in hot paths (inside tight loops)
- Use `fmt.Println()` for logging

**Example**:
```go
// Bad
fmt.Println("User logged in")
logger.Infof("CPF: %s", cpf)  // Sensitive data

// Good
logger.WithField("user_id", userID).Info("User logged in")
logger.Infof("CPF masked: %s", maskCPF(cpf))  // ***.***.901-**
```

### 5.4. Request Logging Middleware

```go
// internal/infrastructure/fiber/middleware/logger.go
package middleware

import (
    "time"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/sirupsen/logrus"
)

func RequestLogger(logger *logrus.Logger) fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()

        // Generate trace ID
        traceID := uuid.New().String()
        c.Locals("trace_id", traceID)

        // Log request
        logger.WithFields(logrus.Fields{
            "trace_id": traceID,
            "method":   c.Method(),
            "path":     c.Path(),
            "ip":       c.IP(),
        }).Info("Request started")

        // Continue request
        err := c.Next()

        // Log response
        duration := time.Since(start)
        logger.WithFields(logrus.Fields{
            "trace_id":    traceID,
            "method":      c.Method(),
            "path":        c.Path(),
            "status":      c.Response().StatusCode(),
            "duration_ms": duration.Milliseconds(),
        }).Info("Request completed")

        return err
    }
}
```

---

## 6. Testing Practices

### 6.1. Test Structure

**Follow AAA pattern**: Arrange, Act, Assert

```go
func TestCreateEntry_ValidInput_Success(t *testing.T) {
    // Arrange
    repository := mock.NewEntryRepository()
    logger := mock.NewLogger()
    useCase := NewCreateEntryUseCase(repository, logger)

    input := CreateEntryInput{
        KeyType:   "CPF",
        KeyValue:  "12345678901",
        AccountID: uuid.New(),
    }

    // Act
    entry, err := useCase.Execute(context.Background(), input)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, entry)
    assert.Equal(t, "CPF", entry.KeyType)
    assert.Equal(t, "12345678901", entry.KeyValue)
}
```

### 6.2. Table-Driven Tests

**Use table-driven tests for multiple scenarios**:

```go
func TestValidateCPF(t *testing.T) {
    tests := []struct {
        name    string
        cpf     string
        wantErr bool
    }{
        {
            name:    "Valid CPF",
            cpf:     "12345678901",
            wantErr: false,
        },
        {
            name:    "Invalid CPF length",
            cpf:     "123",
            wantErr: true,
        },
        {
            name:    "Invalid CPF format",
            cpf:     "abc12345678",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateCPF(tt.cpf)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateCPF() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 6.3. Mocking

**Use interfaces for mocking**:

```go
// internal/ports/repositories.go
type EntryRepository interface {
    Create(ctx context.Context, entry *domain.Entry) error
    FindByID(ctx context.Context, id uuid.UUID) (*domain.Entry, error)
    ExistsByKey(ctx context.Context, keyType, keyValue string) (bool, error)
}

// tests/mocks/entry_repository_mock.go
type MockEntryRepository struct {
    mock.Mock
}

func (m *MockEntryRepository) Create(ctx context.Context, entry *domain.Entry) error {
    args := m.Called(ctx, entry)
    return args.Error(0)
}

func (m *MockEntryRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Entry, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.Entry), args.Error(1)
}
```

### 6.4. Integration Tests

**Tag integration tests**:

```go
// +build integration

package database_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestEntryRepository_Create_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer teardownTestDB(t, db)

    repository := NewEntryRepository(db)

    entry := &domain.Entry{
        ID:       uuid.New(),
        KeyType:  "CPF",
        KeyValue: "12345678901",
    }

    // Act
    err := repository.Create(context.Background(), entry)

    // Assert
    assert.NoError(t, err)

    // Verify entry was created
    found, err := repository.FindByID(context.Background(), entry.ID)
    assert.NoError(t, err)
    assert.Equal(t, entry.KeyValue, found.KeyValue)
}
```

**Run integration tests**:
```bash
# Run only integration tests
go test -tags=integration ./...

# Run all tests
go test ./...
```

### 6.5. Test Coverage

**Aim for >80% coverage**:

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# Check coverage percentage
go test -cover ./...
```

---

## 7. Code Review Checklist

### 7.1. General

- [ ] Code follows Go formatting standards (gofmt, goimports)
- [ ] No linting errors (`golangci-lint run`)
- [ ] All tests pass (`go test ./...`)
- [ ] Code coverage is >80%
- [ ] No TODOs or FIXMEs without GitHub issues

### 7.2. Code Quality

- [ ] Functions are small and focused (single responsibility)
- [ ] No code duplication
- [ ] Naming is clear and descriptive
- [ ] Comments explain "why", not "what"
- [ ] No magic numbers (use constants)
- [ ] Proper error handling (no ignored errors)

### 7.3. Performance

- [ ] No unnecessary allocations in hot paths
- [ ] Database queries are optimized (use indexes)
- [ ] No N+1 query problems
- [ ] Caching is implemented where appropriate
- [ ] Goroutines are properly managed (no leaks)

### 7.4. Security

- [ ] No SQL injection vulnerabilities (use parameterized queries)
- [ ] No sensitive data logged
- [ ] Input validation is present
- [ ] Authentication/authorization is enforced
- [ ] Secrets are not hardcoded

### 7.5. Testing

- [ ] Unit tests cover all business logic
- [ ] Integration tests cover critical paths
- [ ] Edge cases are tested
- [ ] Error scenarios are tested
- [ ] Mocks are used appropriately

### 7.6. Documentation

- [ ] Public functions have GoDoc comments
- [ ] README is updated (if needed)
- [ ] API documentation is updated
- [ ] Migration guide is provided (for breaking changes)

---

## 8. Performance Guidelines

### 8.1. Avoid Unnecessary Allocations

**Bad**:
```go
// Creates new slice on every iteration
for i := 0; i < n; i++ {
    slice := []int{}
    slice = append(slice, i)
}
```

**Good**:
```go
// Pre-allocate slice
slice := make([]int, 0, n)
for i := 0; i < n; i++ {
    slice = append(slice, i)
}
```

### 8.2. Use Buffered Channels

**Bad**:
```go
ch := make(chan int)  // Unbuffered
```

**Good**:
```go
ch := make(chan int, 100)  // Buffered
```

### 8.3. Use sync.Pool for Temporary Objects

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func processRequest() {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()

    // Use buffer
}
```

### 8.4. Database Connection Pooling

```go
// Configure connection pool
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
sqlDB, err := db.DB()

// Set connection pool settings
sqlDB.SetMaxOpenConns(100)
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(time.Hour)
```

### 8.5. Use Context Timeouts

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := repository.FindByID(ctx, id)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        return nil, fmt.Errorf("operation timed out")
    }
    return nil, err
}
```

---

## 9. Security Best Practices

### 9.1. Input Validation

**Always validate input**:

```go
func ValidateCreateEntryInput(input CreateEntryInput) error {
    if input.KeyType == "" {
        return &ValidationError{Field: "key_type", Message: "required"}
    }

    if !isValidKeyType(input.KeyType) {
        return &ValidationError{Field: "key_type", Message: "invalid type"}
    }

    if err := validateKeyValue(input.KeyType, input.KeyValue); err != nil {
        return err
    }

    return nil
}
```

### 9.2. SQL Injection Prevention

**Always use parameterized queries**:

**Bad**:
```go
query := fmt.Sprintf("SELECT * FROM entries WHERE key_value = '%s'", keyValue)
db.Raw(query).Scan(&entry)
```

**Good**:
```go
db.Where("key_value = ?", keyValue).First(&entry)
```

### 9.3. Sensitive Data Masking

```go
func maskCPF(cpf string) string {
    if len(cpf) != 11 {
        return "***"
    }
    return fmt.Sprintf("***.***.%s-**", cpf[7:10])
}

func maskEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return "***"
    }
    return fmt.Sprintf("%s***@%s", parts[0][:1], parts[1])
}

// Usage
logger.Infof("Entry created for CPF: %s", maskCPF(cpf))
```

### 9.4. JWT Validation

```go
func validateJWT(tokenString string) (*jwt.Token, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Validate signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(jwtSecret), nil
    })

    if err != nil {
        return nil, err
    }

    if !token.Valid {
        return nil, errors.New("invalid token")
    }

    return token, nil
}
```

---

## Rastreabilidade

### Requisitos Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RF-DEV-001 | Go coding standards | Best Practices | ✅ Especificado |
| RF-DEV-002 | Project structure guidelines | [IMP-001](./IMP-001_Manual_Implementacao_Core_DICT.md) | ✅ Especificado |
| RF-DEV-003 | Error handling patterns | Best Practices | ✅ Especificado |
| RF-DEV-004 | Testing practices | Best Practices | ✅ Especificado |

---

## Próximas Revisões

**Pendências**:
- [ ] Adicionar guidelines para gRPC development
- [ ] Adicionar performance benchmarking practices
- [ ] Definir observability standards (OpenTelemetry)
- [ ] Adicionar CI/CD pipeline guidelines

---

**Referências**:
- [Effective Go](https://go.dev/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [IMP-001: Manual de Implementação Core DICT](./IMP-001_Manual_Implementacao_Core_DICT.md)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
