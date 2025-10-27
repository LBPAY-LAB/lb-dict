# Clean Architecture Structure

This project follows Clean Architecture principles with 4 distinct layers.

## Layer Overview

```
internal/
├── api/                    # API Layer (Presentation)
│   └── grpc/              # gRPC server and handlers
│       ├── server.go      # gRPC server setup
│       └── handlers/      # gRPC request handlers
│
├── application/            # Application Layer (Use Cases)
│   └── usecases/          # Business logic orchestration
│       ├── create_entry.go
│       ├── query_entry.go
│       ├── delete_entry.go
│       └── create_claim.go
│
├── domain/                 # Domain Layer (Business Rules)
│   ├── entities/          # Core business entities
│   │   ├── dict_entry.go
│   │   └── claim.go
│   ├── interfaces/        # Port definitions
│   │   ├── bacen_client.go
│   │   ├── message_publisher.go
│   │   └── circuit_breaker.go
│   └── valueobjects/      # Immutable value objects
│       ├── bacen_request.go
│       └── bacen_response.go
│
├── infrastructure/         # Infrastructure Layer (External)
│   ├── bacen/             # Bacen HTTP client implementation
│   │   └── http_client.go
│   ├── circuitbreaker/    # Circuit breaker implementation
│   │   └── gobreaker.go
│   └── pulsar/            # Apache Pulsar publisher
│       └── publisher.go
│
└── di/                     # Dependency Injection
    └── container.go       # DI container for wiring dependencies
```

## Dependency Flow

```
API Layer → Application Layer → Domain Layer ← Infrastructure Layer
```

- **API Layer** depends on Application Layer
- **Application Layer** depends on Domain Layer interfaces
- **Infrastructure Layer** implements Domain Layer interfaces
- **Domain Layer** has no dependencies (Pure business logic)

## Key Principles

### 1. Dependency Rule
Dependencies point inward. Inner layers know nothing about outer layers.

### 2. Interface Segregation
Infrastructure implements interfaces defined in the domain layer.

### 3. Single Responsibility
Each layer has a clear, distinct responsibility.

### 4. Testability
Business logic can be tested independently of infrastructure.

## Layer Responsibilities

### Domain Layer (`domain/`)
- **entities/**: Core business objects with behavior
- **interfaces/**: Port definitions (what we need from infrastructure)
- **valueobjects/**: Immutable objects representing concepts

**Rules:**
- No external dependencies
- Pure business logic
- Framework-agnostic

### Application Layer (`application/`)
- **usecases/**: Orchestrate domain entities and infrastructure
- Coordinate between domain and infrastructure
- Transaction boundaries
- Business workflows

**Rules:**
- Can use domain entities and interfaces
- Can call infrastructure through interfaces
- No framework-specific code

### Infrastructure Layer (`infrastructure/`)
- **bacen/**: HTTP client for Bacen DICT API
- **circuitbreaker/**: Circuit breaker for resilience
- **pulsar/**: Message broker for event publishing

**Rules:**
- Implements domain interfaces
- Contains all external dependencies
- Adapts external libraries to our interfaces

### API Layer (`api/`)
- **grpc/**: gRPC server and handlers
- Receives requests from outside
- Converts requests to use case calls
- Returns responses

**Rules:**
- Thin layer, minimal logic
- Protocol-specific code only
- Delegates to use cases

## Usage Example

```go
// 1. Initialize DI Container
config := di.DefaultConfig()
config.BacenBaseURL = "https://api.bacen.gov.br"
config.GRPCPort = 50051

container, err := di.NewContainer(config)
if err != nil {
    log.Fatal(err)
}
defer container.Close()

// 2. Start gRPC Server
if err := container.GRPCServer.Start(); err != nil {
    log.Fatal(err)
}
```

## Testing Strategy

### Unit Tests
- **Domain entities**: Test business rules in isolation
- **Use cases**: Mock infrastructure dependencies
- **Infrastructure**: Test adapters independently

### Integration Tests
- Test use cases with real infrastructure
- Test API handlers end-to-end

### Example:
```go
// Mock infrastructure for use case testing
mockBacenClient := &MockBacenClient{}
mockPublisher := &MockPublisher{}
mockCircuitBreaker := &MockCircuitBreaker{}

useCase := usecases.NewCreateEntryUseCase(
    mockBacenClient,
    mockPublisher,
    mockCircuitBreaker,
)

// Test use case logic without real external dependencies
result, err := useCase.Execute(ctx, request)
```

## Adding New Features

### 1. Define Domain Entity
```go
// internal/domain/entities/new_entity.go
type NewEntity struct {
    // Define business object
}
```

### 2. Define Interface (if needed)
```go
// internal/domain/interfaces/new_service.go
type NewService interface {
    DoSomething(ctx context.Context) error
}
```

### 3. Create Use Case
```go
// internal/application/usecases/new_use_case.go
type NewUseCase struct {
    service interfaces.NewService
}
```

### 4. Implement Infrastructure
```go
// internal/infrastructure/newservice/implementation.go
type ServiceImpl struct {}

func (s *ServiceImpl) DoSomething(ctx context.Context) error {
    // Implementation
}
```

### 5. Add API Handler
```go
// internal/api/grpc/handlers/new_handler.go
func (h *Handler) NewOperation(ctx context.Context, req *pb.Request) (*pb.Response, error) {
    // Call use case
}
```

### 6. Wire in DI Container
```go
// internal/di/container.go
func (c *Container) initApplication() error {
    c.NewUseCase = usecases.NewUseCase(c.NewService)
}
```

## Best Practices

1. **Keep domain pure**: No external dependencies in domain layer
2. **Use interfaces**: Define contracts in domain, implement in infrastructure
3. **Thin handlers**: API handlers should just translate and delegate
4. **Rich use cases**: Business logic lives in use cases and entities
5. **Single responsibility**: Each file has one clear purpose
6. **Testable code**: Mock dependencies at boundaries
7. **Explicit dependencies**: Use constructor injection, no globals

## Next Steps

- [ ] Implement proto files for gRPC contracts
- [ ] Add logging and observability
- [ ] Implement comprehensive error handling
- [ ] Add unit tests for each layer
- [ ] Add integration tests
- [ ] Implement retry logic for Bacen calls
- [ ] Add metrics and health checks
- [ ] Document API contracts
