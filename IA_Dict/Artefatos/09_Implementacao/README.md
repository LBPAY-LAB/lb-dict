# Implementação

**Propósito**: Guias e padrões de implementação de código para desenvolvedores

## 📋 Conteúdo

Esta pasta armazenará:

- **Coding Standards**: Padrões de código (Go, TypeScript, etc.)
- **Design Patterns**: Padrões de design aplicados no projeto
- **Best Practices**: Melhores práticas de implementação
- **Code Examples**: Exemplos de código para referência
- **Migration Guides**: Guias de migração entre versões

## 📁 Estrutura Esperada

```
Implementacao/
├── Coding_Standards/
│   ├── Go_Style_Guide.md
│   ├── TypeScript_Style_Guide.md
│   ├── SQL_Best_Practices.md
│   └── Git_Commit_Guidelines.md
├── Design_Patterns/
│   ├── Repository_Pattern.md
│   ├── Factory_Pattern.md
│   ├── Saga_Pattern_Temporal.md
│   └── CQRS_Pattern.md
├── Examples/
│   ├── Clean_Architecture_Example.md
│   ├── gRPC_Client_Example.md
│   ├── Temporal_Workflow_Example.md
│   └── Unit_Test_Example.md
└── Migration_Guides/
    ├── v1.0_to_v2.0.md
    └── Database_Migration_Checklist.md
```

## 🎯 Go Coding Standards

### Project Structure

```
apps/connect/
├── cmd/
│   ├── server/          # Entrypoint (main.go)
│   └── worker/          # Temporal worker
├── internal/
│   ├── api/             # API Layer (gRPC handlers)
│   ├── domain/          # Domain Layer (entities, use cases)
│   ├── application/     # Application Layer (services)
│   └── infrastructure/  # Infrastructure (repos, clients)
├── pkg/                 # Shared libraries
├── db/
│   └── migrations/      # Goose migrations
├── proto/               # Protocol Buffers
├── config/              # Configuration files
└── tests/               # Integration tests
```

### Naming Conventions

```go
// Package names: lowercase, singular
package user

// Interface names: -er suffix
type UserRepository interface {
    Create(ctx context.Context, user *User) error
}

// Struct names: PascalCase
type PostgresUserRepository struct {
    db *sql.DB
}

// Function names: PascalCase (exported), camelCase (internal)
func (r *PostgresUserRepository) Create(ctx context.Context, user *User) error {
    return r.insert(ctx, user)
}

func (r *PostgresUserRepository) insert(ctx context.Context, user *User) error {
    // Implementation
}

// Variable names: camelCase
var userRepo UserRepository

// Constants: PascalCase or UPPER_SNAKE_CASE
const MaxRetries = 3
const DEFAULT_TIMEOUT = 5 * time.Second
```

### Error Handling

```go
// ALWAYS return errors, never panic in production code
func CreateEntry(ctx context.Context, entry *Entry) error {
    if err := validateEntry(entry); err != nil {
        return fmt.Errorf("validation failed: %w", err)  // Wrap errors
    }

    if err := repo.Save(ctx, entry); err != nil {
        return fmt.Errorf("failed to save entry: %w", err)
    }

    return nil
}

// Use custom error types for domain errors
type DuplicateKeyError struct {
    KeyValue string
}

func (e *DuplicateKeyError) Error() string {
    return fmt.Sprintf("key %s already exists", e.KeyValue)
}

// Check error types
if errors.Is(err, &DuplicateKeyError{}) {
    // Handle duplicate key
}
```

### Logging

```go
// Use structured logging (zap)
import "go.uber.org/zap"

func CreateEntry(ctx context.Context, entry *Entry) error {
    logger := zap.L().With(
        zap.String("entry_id", entry.ID),
        zap.String("key_type", entry.KeyType),
        zap.String("request_id", getRequestID(ctx)),
    )

    logger.Info("creating entry")

    if err := repo.Save(ctx, entry); err != nil {
        logger.Error("failed to save entry", zap.Error(err))
        return err
    }

    logger.Info("entry created successfully")
    return nil
}
```

### Testing

```go
// Test function naming: Test<FunctionName>_<Scenario>
func TestCreateEntry_Success(t *testing.T) {
    // Arrange
    entry := &Entry{
        KeyType:  "CPF",
        KeyValue: "12345678900",
    }

    // Act
    err := CreateEntry(context.Background(), entry)

    // Assert
    assert.NoError(t, err)
}

func TestCreateEntry_DuplicateKey_ReturnsError(t *testing.T) {
    // ...
}

// Table-driven tests for multiple scenarios
func TestValidateEntry(t *testing.T) {
    tests := []struct {
        name    string
        entry   *Entry
        wantErr bool
    }{
        {
            name:    "valid CPF",
            entry:   &Entry{KeyType: "CPF", KeyValue: "12345678900"},
            wantErr: false,
        },
        {
            name:    "invalid CPF",
            entry:   &Entry{KeyType: "CPF", KeyValue: "123"},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateEntry(tt.entry)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## 🏗️ Design Patterns

### 1. Repository Pattern

```go
// domain/repository.go
type EntryRepository interface {
    Create(ctx context.Context, entry *Entry) error
    GetByID(ctx context.Context, id string) (*Entry, error)
    GetByKey(ctx context.Context, keyType, keyValue string) (*Entry, error)
    Update(ctx context.Context, entry *Entry) error
    Delete(ctx context.Context, id string) error
}

// infrastructure/postgres_entry_repository.go
type PostgresEntryRepository struct {
    db *sql.DB
}

func NewPostgresEntryRepository(db *sql.DB) *PostgresEntryRepository {
    return &PostgresEntryRepository{db: db}
}

func (r *PostgresEntryRepository) Create(ctx context.Context, entry *domain.Entry) error {
    query := `
        INSERT INTO dict.entries (id, key_type, key_value, status)
        VALUES ($1, $2, $3, $4)
    `
    _, err := r.db.ExecContext(ctx, query, entry.ID, entry.KeyType, entry.KeyValue, entry.Status)
    return err
}
```

### 2. Factory Pattern

```go
// Create different types of workflows
type WorkflowFactory interface {
    CreateWorkflow(workflowType string) (temporal.Workflow, error)
}

type TemporalWorkflowFactory struct{}

func (f *TemporalWorkflowFactory) CreateWorkflow(workflowType string) (temporal.Workflow, error) {
    switch workflowType {
    case "claim":
        return &ClaimWorkflow{}, nil
    case "portability":
        return &PortabilityWorkflow{}, nil
    default:
        return nil, fmt.Errorf("unknown workflow type: %s", workflowType)
    }
}
```

### 3. Saga Pattern (Temporal)

```go
// Compensating transactions for distributed operations
func ClaimWorkflow(ctx workflow.Context, input ClaimInput) error {
    // Step 1: Create claim in Bacen
    var externalID string
    err := workflow.ExecuteActivity(ctx, CreateClaimInBacen, input).Get(ctx, &externalID)
    if err != nil {
        return err  // No compensation needed (first step)
    }

    // Step 2: Save claim locally
    err = workflow.ExecuteActivity(ctx, SaveClaimLocally, externalID).Get(ctx, nil)
    if err != nil {
        // Compensate: Cancel claim in Bacen
        workflow.ExecuteActivity(ctx, CancelClaimInBacen, externalID)
        return err
    }

    // Step 3: Wait 30 days (timer)
    err = workflow.Sleep(ctx, 30*24*time.Hour)
    if err != nil {
        return err
    }

    // Step 4: Complete claim
    return workflow.ExecuteActivity(ctx, CompleteClaim, externalID).Get(ctx, nil)
}
```

## 📚 Code Examples

### gRPC Client Example

```go
// Create gRPC connection to Bridge
conn, err := grpc.Dial(
    "bridge.dict.svc.cluster.local:8081",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(
        grpc_retry.WithMax(3),
        grpc_retry.WithBackoff(grpc_retry.BackoffLinear(100*time.Millisecond)),
    )),
)
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

client := pb.NewBridgeServiceClient(conn)

// Call CreateEntry RPC
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

response, err := client.CreateEntry(ctx, &pb.CreateEntryRequest{
    Key: &pb.DictKey{
        KeyType:  pb.KeyType_KEY_TYPE_CPF,
        KeyValue: "12345678900",
    },
    Account: &pb.Account{
        Ispb:          "00000000",
        AccountNumber: "12345-6",
    },
})
if err != nil {
    log.Fatalf("CreateEntry failed: %v", err)
}

log.Printf("Entry created: %s", response.EntryId)
```

## 🔧 Git Commit Guidelines

### Conventional Commits

```bash
# Format: <type>(<scope>): <subject>

# Types:
# - feat: Nova funcionalidade
# - fix: Correção de bug
# - docs: Documentação
# - style: Formatação (não muda lógica)
# - refactor: Refatoração
# - test: Testes
# - chore: Manutenção (build, CI, etc.)

# Examples:
git commit -m "feat(connect): add CreateEntry gRPC handler"
git commit -m "fix(bridge): handle mTLS certificate expiration"
git commit -m "docs(readme): update installation instructions"
git commit -m "test(claim): add unit tests for ClaimWorkflow"
git commit -m "refactor(repo): extract common query logic"
```

## 📋 Definition of Done

Para considerar uma tarefa completada:

- [ ] Código desenvolvido seguindo padrões deste guia
- [ ] Testes unitários escritos (> 80% coverage)
- [ ] Testes de integração (se aplicável)
- [ ] Code review aprovado por 2+ desenvolvedores
- [ ] Documentação atualizada (README, Swagger, comentários)
- [ ] Migrations criadas (se mudança de schema)
- [ ] Logs estruturados implementados
- [ ] Métricas Prometheus adicionadas
- [ ] Pipeline CI passou (lint + tests + build)
- [ ] Deploy em staging validado

## 📚 Referências

- [Effective Go](https://go.dev/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Clean Architecture (Uncle Bob)](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [12 Factor App](https://12factor.net/)

---

**Status**: 🔴 Pasta vazia (será preenchida na Fase 2)
**Fase de Preenchimento**: Fase 2 (início do desenvolvimento)
**Responsável**: Tech Lead + Desenvolvedores Sênior
**Revisão**: Semestral (atualizar boas práticas)
