# Task 3: Existing Patterns Analysis

## Overview
Analysis of existing architectural patterns in connector-dict to ensure the VSync implementation follows established conventions.

## Architecture Overview

### Application Structure
```
connector-dict/
├── apps/
│   ├── dict/                    # REST API application
│   │   ├── application/         # Use cases and business logic
│   │   ├── domain/              # Domain entities
│   │   ├── handlers/            # HTTP handlers
│   │   ├── infrastructure/      # External integrations
│   │   └── setup/               # App configuration
│   └── orchestration-worker/    # Async processing with Temporal
│       ├── application/         # Use cases
│       ├── handlers/            # Pulsar message handlers
│       ├── infrastructure/      # Temporal workflows & activities
│       └── setup/               # Worker configuration
└── shared/                      # Shared libraries
    └── infrastructure/          # Common infra (cache, observability)
```

## Key Architectural Patterns

### 1. Clean Architecture Pattern

**Layers:**
```
Handlers → Application → Domain → Infrastructure
```

**Layer Responsibilities:**
- **Handlers**: Request/response handling, input validation
- **Application**: Business logic orchestration, use cases
- **Domain**: Core business entities and rules
- **Infrastructure**: External services, databases, message brokers

**Example Flow (Claim):**
```go
// Handler Layer
func (c *Controller) CreateClaim(ctx context.Context, input CreateClaimRequest) (*CreateClaimResponse, error)
    ↓
// Application Layer
func (app *Application) CreateClaim(ctx context.Context, claim ClaimRequest) (*ClaimResponse, error)
    ↓
// Infrastructure Layer
func (g *Gateway) CreateClaim(ctx context.Context, in CreateClaimRequest) (*CreateClaimResponse, error)
```

### 2. Event-Driven Pattern

**Pulsar Message Flow:**
```
Producer → Topic → Consumer → Handler → Application → Temporal
```

**Handler Pattern:**
```go
type Handler struct {
    obsProvider observability.Provider
    claimApp    *claim.Application
}

func (h *Handler) CreateHandler(ctx context.Context, message pubsub.Message) error {
    // 1. Parse message properties
    props, err := pkg.ParseMessageProperties(message.Properties)

    // 2. Decode message payload
    var request claimsdk.CreateClaimRequest
    if err := message.Decode(&request); err != nil {
        return err
    }

    // 3. Delegate to application
    return h.claimApp.CreateClaim(ctx, props.CorrelationID, &request)
}
```

### 3. Temporal Workflow Pattern

**Workflow Structure:**
```go
// Input structure
type CreateClaimWorkflowInput struct {
    Request *pkgClaim.CreateClaimRequest
    Hash    string  // Correlation ID
}

// Workflow implementation
func CreateClaimWorkflow(ctx workflow.Context, input CreateClaimWorkflowInput) error {
    // 1. Execute main activity
    resp, err := executeActivity(ctx, input)

    // 2. Cache response
    workflows.ExecuteCacheActivity(ctx, input.Hash, resp)

    // 3. Publish events
    workflows.ExecuteDictEventsPublishActivity(ctx, input.Hash, action, resp)

    // 4. Start child workflows if needed
    startMonitorWorkflow(ctx, resp)
}
```

**Activity Pattern:**
```go
func (a *Activity) Execute(ctx context.Context, input Input) error {
    // Business logic here
    return nil
}
```

### 4. Dependency Injection Pattern

**Setup Pattern:**
```go
// Setup dependencies
func Setup(cfg *Config) (*App, error) {
    // 1. Infrastructure
    pulsarClient := NewPulsar(cfg)
    temporalClient := NewTemporal(cfg)
    cache := NewRedisCache(cfg)

    // 2. Application services
    claimApp := claim.NewApplication(observer, claimService)

    // 3. Handlers
    handler := claim.NewHandler(observer, claimApp)

    // 4. Wire together
    consumer := NewPulsarConsumer(pulsarClient, handlers)

    return &App{consumer, temporal}, nil
}
```

### 5. Configuration Pattern

**Environment-based Config:**
```go
type Config struct {
    // Service
    ServiceName    string `env:"SERVICE_NAME"`
    ServiceVersion string `env:"SERVICE_VERSION"`

    // Pulsar
    PulsarURL             string `env:"PULSAR_URL"`
    PulsarTopicDictEvents string `env:"PULSAR_TOPIC_DICT_EVENTS"`

    // Temporal
    TemporalURL string `env:"TEMPORAL_URL"`

    // Redis
    RedisAddr string `env:"REDIS_ADDR"`
}
```

### 6. Error Handling Pattern

**Structured Errors:**
```go
// Domain error
type RFC9457Error struct {
    Status int
    Title  string
    Detail string
}

// Workflow error handling
if err != nil {
    if notifyErr := workflows.NotifyFailure(ctx, input.Hash, action, err); notifyErr != nil {
        workflow.GetLogger(ctx).Error("Failed to notify", "error", notifyErr)
    }
    return err
}
```

### 7. Observability Pattern

**Structured Logging:**
```go
logger := app.observer.Logger()
logger.Info(ctx, fmt.Sprintf("Creating entry: %v", entry.Key))
logger.Error(ctx, "Failed to create entry", err)
```

**Tracing:**
```go
tracer := observer.Tracer()
ctx, span := tracer.Start(ctx, "CreateEntry")
defer span.End()
```

## Pattern Recommendations for VSync

### 1. Container Structure
```
apps/dict.vsync/
├── application/
│   ├── usecases/
│   │   ├── cid/              # CID generation use cases
│   │   └── sync/             # Sync orchestration
│   └── ports/                # Interface definitions
├── domain/
│   ├── cid/                  # CID domain logic
│   └── entry/                # Reuse from dict
├── handlers/
│   └── pulsar/               # Event consumers
├── infrastructure/
│   ├── database/             # PostgreSQL repositories
│   ├── temporal/             # Workflow definitions
│   │   ├── activities/
│   │   └── workflows/
│   └── pulsar/               # Publishers if needed
└── setup/                    # Application bootstrap
```

### 2. Handler Implementation
```go
// Following existing pattern
type VSyncHandler struct {
    observer observability.Provider
    vsyncApp *vsync.Application
}

func (h *VSyncHandler) HandleEntryEvent(ctx context.Context, message pubsub.Message) error {
    // Parse properties
    props, err := pkg.ParseMessageProperties(message.Properties)

    // Decode Entry event
    var entry domain.Entry
    if err := message.Decode(&entry); err != nil {
        return err
    }

    // Process based on action
    switch props.Action {
    case "key.created", "key.updated":
        return h.vsyncApp.ProcessEntry(ctx, props.CorrelationID, &entry)
    case "key.deleted":
        return h.vsyncApp.DeleteEntry(ctx, props.CorrelationID, entry.Key)
    }
}
```

### 3. Workflow Pattern for VSync
```go
type VSyncWorkflowInput struct {
    Entry         *domain.Entry
    Action        string
    CorrelationID string
}

func VSyncWorkflow(ctx workflow.Context, input VSyncWorkflowInput) error {
    // 1. Generate CID
    cid := activities.GenerateCID(ctx, input.Entry)

    // 2. Store in database
    activities.StoreCID(ctx, cid)

    // 3. Send to BACEN if needed
    if shouldSync(input.Action) {
        activities.SendToBACEN(ctx, cid)
    }

    return nil
}
```

### 4. Repository Pattern
```go
type CIDRepository interface {
    Create(ctx context.Context, cid *domain.CID) error
    Update(ctx context.Context, cid *domain.CID) error
    Delete(ctx context.Context, key string) error
    GetByKey(ctx context.Context, key string) (*domain.CID, error)
    GetBatch(ctx context.Context, limit int) ([]*domain.CID, error)
}
```

## Testing Patterns

### Unit Tests
```go
func TestVSyncHandler_HandleEntryEvent(t *testing.T) {
    // Arrange
    mockApp := &mockApplication{}
    handler := NewVSyncHandler(observer, mockApp)

    // Act
    err := handler.HandleEntryEvent(ctx, message)

    // Assert
    assert.NoError(t, err)
    assert.True(t, mockApp.ProcessEntryCalled)
}
```

### Integration Tests
```go
func TestVSyncWorkflow_Integration(t *testing.T) {
    testSuite := &testsuite.WorkflowTestSuite{}
    env := testSuite.NewTestWorkflowEnvironment()

    env.ExecuteWorkflow(VSyncWorkflow, input)

    require.True(t, env.IsWorkflowCompleted())
    require.NoError(t, env.GetWorkflowError())
}
```

## Key Patterns to Follow

1. **Clean Architecture**: Maintain layer separation
2. **Dependency Injection**: Use interfaces, inject dependencies
3. **Event-Driven**: Async processing via Pulsar
4. **Workflow Orchestration**: Complex flows via Temporal
5. **Observability First**: Logging, tracing, metrics
6. **Error Handling**: Structured errors, proper propagation
7. **Configuration**: Environment-based, typed config
8. **Testing**: Unit and integration tests

## Anti-Patterns to Avoid

1. **Direct Database Access**: Always use repositories
2. **Business Logic in Handlers**: Keep in application layer
3. **Synchronous Long Operations**: Use Temporal for long-running tasks
4. **Tight Coupling**: Use interfaces, avoid concrete dependencies
5. **Missing Error Handling**: Always handle and log errors
6. **Hardcoded Values**: Use configuration
7. **Missing Observability**: Always add logs and traces

## Conclusion

The connector-dict follows clean, well-established patterns that should be replicated in the VSync implementation:

1. **Clean Architecture** for clear separation of concerns
2. **Event-Driven** for scalable async processing
3. **Temporal Workflows** for complex orchestration
4. **Dependency Injection** for testability
5. **Observability** for production readiness

Following these patterns will ensure the VSync system integrates seamlessly with the existing architecture.