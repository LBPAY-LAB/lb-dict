# Clean Architecture Setup - conn-dict

## Task: CONNECT-002 - Clean Architecture Structure

**Status**: COMPLETED

**Date**: 2025-10-26

---

## Overview

Successfully implemented a 4-layer Clean Architecture structure with Domain-Driven Design (DDD) patterns for the conn-dict service. The architecture follows hexagonal/ports-and-adapters principles with clear separation of concerns.

## Statistics

- **Total Go Files**: 25
- **Total Directories**: 16
- **Lines of Code**: 2,383
- **Packages**: 11

---

## Architecture Layers

### 1. API Layer (`internal/api/`)
**Purpose**: Handles external communication (gRPC)

**Components**:
- `grpc/server.go` - Main gRPC server with graceful shutdown
- `grpc/handlers/claim_handler.go` - DICT claim operations handler
- `grpc/handlers/vsync_handler.go` - DICT vsync operations handler

**Responsibilities**:
- Request/response handling
- Protocol translation (gRPC to domain)
- Input validation
- Error mapping

---

### 2. Application Layer (`internal/application/`)
**Purpose**: Orchestrates business logic and use cases

**Components**:

**Use Cases**:
- `usecases/claim_usecase.go` - Claim creation, confirmation, cancellation
- `usecases/vsync_usecase.go` - Entry synchronization, batch processing

**Commands**:
- `commands/claim_commands.go` - CreateClaim, ConfirmClaim, CancelClaim
- `commands/vsync_commands.go` - SyncEntry, ProcessVsyncBatch

**Responsibilities**:
- Use case orchestration
- Transaction boundaries
- Command validation
- Event publishing coordination

---

### 3. Domain Layer (`internal/domain/`)
**Purpose**: Core business logic and rules (Framework-agnostic)

**Components**:

**Aggregates** (Aggregate Roots):
- `aggregates/claim.go` - Claim aggregate with status transitions
  - States: PENDING, CONFIRMED, CANCELLED, EXPIRED
  - Methods: Confirm(), Cancel(), MarkAsExpired()
  - Invariants: Business rule enforcement

- `aggregates/vsync_entry.go` - Vsync entry aggregate
  - States: PENDING, IN_PROGRESS, COMPLETED, FAILED
  - Methods: StartSync(), CompleteSync(), FailSync()
  - Retry logic: Max 3 attempts

**Events** (Domain Events):
- `events/claim_events.go`
  - ClaimCreatedEvent
  - ClaimConfirmedEvent
  - ClaimCancelledEvent
  - ClaimExpiredEvent

- `events/vsync_events.go`
  - VsyncEntryCreatedEvent
  - VsyncStartedEvent
  - VsyncCompletedEvent
  - VsyncFailedEvent

**Interfaces** (Ports):
- `interfaces/repositories.go` - ClaimRepository, VsyncRepository
- `interfaces/temporal.go` - TemporalClient
- `interfaces/event_publisher.go` - EventPublisher
- `interfaces/cache.go` - CacheRepository

**Responsibilities**:
- Business rule enforcement
- State management
- Domain event generation
- Aggregate consistency

---

### 4. Infrastructure Layer (`internal/infrastructure/`)
**Purpose**: External service adapters and technical implementations

**Components**:

**Temporal**:
- `temporal/client.go` - Workflow execution client
- `temporal/worker.go` - Workflow worker registration
- `temporal/config.go` - Configuration
- `temporal/logger.go` - Temporal logger adapter

**Database**:
- `database/postgres_claim_repository.go` - PostgreSQL claim persistence
- `database/postgres_vsync_repository.go` - PostgreSQL vsync persistence

**Cache**:
- `cache/redis_repository.go` - Redis cache implementation

**Message Broker**:
- `pulsar/event_publisher.go` - Apache Pulsar event publishing

**Responsibilities**:
- External service integration
- Data persistence
- Message publishing
- Caching
- Observability

---

## Workflows Layer (`workflows/`)
**Purpose**: Temporal workflow orchestration

**Components**:

**Workflows**:
- `claim_workflow.go`
  - ClaimWorkflow - Main claim processing (7-day timeout)
  - CancelClaimWorkflow - Claim cancellation handling
  - Signal handling: confirm-claim, cancel-claim

- `vsync_workflow.go`
  - VsyncWorkflow - Single entry synchronization
  - VsyncBatchWorkflow - Parallel batch processing
  - ScheduledVsyncWorkflow - Periodic sync scheduler

**Features**:
- Retry policies with exponential backoff
- Parallel execution for batch operations
- Signal-based event handling
- Timeout management
- Child workflow orchestration

---

## DDD Patterns Implemented

### Aggregates
- **Claim**: Manages claim lifecycle with invariants
- **VsyncEntry**: Manages sync state with retry logic
- Both enforce consistency boundaries
- Event sourcing via domain events

### Domain Events
- Base event structure with metadata
- Event ID, Type, Aggregate ID, Timestamp
- Payload encapsulation
- Event clearing after retrieval

### Commands
- Explicit intent representation
- Validation hooks
- Immutable once created
- Clear naming (CreateX, ConfirmX, CancelX)

### Repositories
- Interface segregation
- Aggregate-focused operations
- No query methods in domain
- Infrastructure implementation

### Ports and Adapters
- Domain defines interfaces (ports)
- Infrastructure provides implementations (adapters)
- Clean dependency inversion
- Testability through mocking

---

## Directory Structure

```
conn-dict/
├── internal/
│   ├── api/                    # API Layer
│   │   └── grpc/
│   │       ├── server.go
│   │       └── handlers/
│   │           ├── claim_handler.go
│   │           └── vsync_handler.go
│   │
│   ├── application/            # Application Layer
│   │   ├── usecases/
│   │   │   ├── claim_usecase.go
│   │   │   └── vsync_usecase.go
│   │   └── commands/
│   │       ├── claim_commands.go
│   │       └── vsync_commands.go
│   │
│   ├── domain/                 # Domain Layer
│   │   ├── aggregates/
│   │   │   ├── claim.go
│   │   │   └── vsync_entry.go
│   │   ├── events/
│   │   │   ├── claim_events.go
│   │   │   └── vsync_events.go
│   │   └── interfaces/
│   │       ├── repositories.go
│   │       ├── temporal.go
│   │       ├── event_publisher.go
│   │       └── cache.go
│   │
│   └── infrastructure/         # Infrastructure Layer
│       ├── temporal/
│       │   ├── client.go
│       │   ├── worker.go
│       │   ├── config.go
│       │   └── logger.go
│       ├── database/
│       │   ├── postgres_claim_repository.go
│       │   └── postgres_vsync_repository.go
│       ├── cache/
│       │   └── redis_repository.go
│       └── pulsar/
│           └── event_publisher.go
│
└── workflows/                   # Workflows Layer
    ├── claim_workflow.go
    └── vsync_workflow.go
```

---

## Dependencies

Updated `go.mod` with dict-contracts:

```go
require (
    github.com/lbpay-lab/dict-contracts v0.1.0  // NEW: Proto contracts
    github.com/google/uuid v1.6.0
    github.com/redis/go-redis/v9 v9.14.1
    go.temporal.io/sdk v1.36.0
    github.com/apache/pulsar-client-go v0.16.0
    google.golang.org/grpc v1.70.0
    google.golang.org/protobuf v1.36.3
)
```

---

## Key Features

### Clean Architecture Benefits
1. **Testability**: Each layer can be tested independently
2. **Maintainability**: Clear separation of concerns
3. **Flexibility**: Easy to swap implementations
4. **Scalability**: Layers can evolve independently
5. **Framework Independence**: Domain is pure Go

### DDD Benefits
1. **Business-Centric**: Code reflects business language
2. **Consistency**: Aggregates enforce invariants
3. **Traceability**: Domain events track state changes
4. **Intent**: Commands express clear intentions
5. **Modularity**: Bounded contexts isolation

### Temporal Integration
1. **Durability**: Workflows survive failures
2. **Observability**: Full execution history
3. **Reliability**: Automatic retries
4. **Scalability**: Parallel execution
5. **Timeouts**: Long-running process management

---

## Next Steps

### Integration Points
1. Implement gRPC service registration with dict-contracts proto
2. Wire up dependency injection in `cmd/server/main.go`
3. Add PostgreSQL migrations for aggregates
4. Configure Pulsar topics and subscriptions
5. Implement activity functions for workflows

### Testing
1. Unit tests for aggregates (business logic)
2. Integration tests for repositories
3. Contract tests for gRPC handlers
4. Workflow replay tests for Temporal

### Observability
1. Add structured logging with correlation IDs
2. Implement OpenTelemetry tracing
3. Define Prometheus metrics
4. Create health check endpoints

---

## Validation

All acceptance criteria met:

- [x] 4 layers (api, application, domain, infrastructure) created
- [x] workflows/ directory with Temporal workflows
- [x] Skeleton files with proper interfaces
- [x] DDD patterns: Aggregates, Events, Commands
- [x] Repository interfaces and implementations
- [x] Temporal client and worker setup
- [x] go.mod updated with dict-contracts dependency
- [x] 25 Go files, 2,383 lines of code
- [x] Clean separation of concerns
- [x] Hexagonal architecture (ports and adapters)

---

## Architecture Principles Applied

1. **Dependency Rule**: Dependencies point inward (domain has no dependencies)
2. **Single Responsibility**: Each layer has one reason to change
3. **Interface Segregation**: Small, focused interfaces
4. **Dependency Inversion**: Abstractions over implementations
5. **Open/Closed**: Open for extension, closed for modification

---

**Setup Complete**: Clean Architecture + DDD + Temporal ready for DICT operations.
