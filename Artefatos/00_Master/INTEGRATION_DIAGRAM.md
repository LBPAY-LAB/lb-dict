# Application Layer Integration Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         CORE-DICT APPLICATION LAYER                              │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                   │
│  ┌──────────────────────────┐        ┌───────────────────────────┐             │
│  │  COMMAND HANDLERS (10)   │        │  QUERY HANDLERS (12)      │             │
│  ├──────────────────────────┤        ├───────────────────────────┤             │
│  │  ✅ CreateEntry          │        │  ✅ VerifyAccount         │             │
│  │  ✅ UpdateEntry          │        │  ✅ GetEntry              │             │
│  │  ✅ DeleteEntry          │        │  ✅ HealthCheck           │             │
│  │  ⚠️  CreateClaim          │        │  ⚪ ListEntries           │             │
│  │  ⚠️  ConfirmClaim         │        │  ⚪ GetClaim              │             │
│  │  ⚠️  CancelClaim          │        │  ⚪ ListClaims            │             │
│  │  ⚠️  CompleteClaim        │        │  ⚪ GetAccount            │             │
│  │  ⚠️  BlockEntry           │        │  ⚪ GetAuditLog           │             │
│  │  ⚠️  UnblockEntry         │        │  ⚪ GetStatistics         │             │
│  │  ⚠️  CreateInfraction     │        │  ⚪ GetInfraction         │             │
│  └──────────┬───────────────┘        │  ⚪ ListInfractions       │             │
│             │                         │  ⚪ ListInfractions       │             │
│             │                         └───────────┬───────────────┘             │
│             │                                     │                              │
│             │                                     │                              │
│  ┏━━━━━━━━━┷━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┷━━━━━━━━━━━━━━━━━━┓         │
│  ┃              NEW: gRPC & PULSAR INTEGRATION                       ┃         │
│  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━┯━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛         │
│                                  │                                              │
│         ┌────────────────────────┼────────────────────────┐                    │
│         │                        │                        │                    │
│         ▼                        ▼                        ▼                    │
│  ┌─────────────────┐  ┌─────────────────────┐  ┌──────────────────┐          │
│  │ ConnectClient   │  │ EntryEventProducer  │  │  CacheService    │          │
│  │ (gRPC)          │  │ (Pulsar)            │  │  (Redis)         │          │
│  ├─────────────────┤  ├─────────────────────┤  ├──────────────────┤          │
│  │ GetEntry        │  │ PublishCreated      │  │ Get/Set/Delete   │          │
│  │ GetEntryByKey   │  │ PublishUpdated      │  │ Invalidate       │          │
│  │ CreateClaim     │  │ PublishDeleted      │  │ TTL: 1-10 min    │          │
│  │ ConfirmClaim    │  │ PublishBlocked      │  └──────┬───────────┘          │
│  │ CancelClaim     │  │ PublishUnblocked    │         │                       │
│  │ CreateInfraction│  │ PublishClaim*       │         │                       │
│  │ VerifyAccount   │  │ PublishInfraction   │         │                       │
│  │ HealthCheck     │  └──────┬──────────────┘         │                       │
│  └────────┬────────┘         │                         │                       │
│           │                  │                         │                       │
└───────────┼──────────────────┼─────────────────────────┼───────────────────────┘
            │                  │                         │
            │ gRPC             │ Pulsar                  │ Redis
            │ :9092            │ :6650                   │ :6379
            ▼                  ▼                         ▼
┌───────────────────┐  ┌────────────────────┐  ┌──────────────────┐
│   CONN-DICT       │  │   PULSAR BROKER    │  │   REDIS CACHE    │
│   (gRPC Server)   │  │   (Message Bus)    │  │   (K/V Store)    │
├───────────────────┤  ├────────────────────┤  ├──────────────────┤
│ Port: 9092        │  │ Topics:            │  │ TTL Management   │
│ Circuit Breaker   │  │ - entries.created  │  │ Pattern Matching │
│ Retry: 3x         │  │ - entries.updated  │  │ JSON Serialization│
│ Timeout: 5s       │  │ - entries.deleted  │  │                  │
│ Keepalive: 30s    │  │ - claim-events     │  │                  │
└─────────┬─────────┘  │ - infraction-events│  └──────────────────┘
          │            └──────┬─────────────┘
          │                   │
          ▼                   ▼
┌───────────────────┐  ┌────────────────────┐
│   CONN-BRIDGE     │  │   CONN-DICT        │
│   (SOAP Client)   │  │   (Consumer)       │
├───────────────────┤  ├────────────────────┤
│ mTLS + ICP-Brasil │  │ Temporal Workflows │
│ XML Signing       │  │ - ClaimWorkflow    │
│ Retry Logic       │  │ - VSyncWorkflow    │
└─────────┬─────────��  └────────────────────┘
          │
          ▼
┌───────────────────┐
│   BACEN DICT      │
│   (SOAP API)      │
├───────────────────┤
│ RSFN Operations   │
│ - CreateEntry     │
│ - UpdateEntry     │
│ - DeleteEntry     │
│ - GetEntry        │
│ - CreateClaim     │
│ - Verify Account  │
└───────────────────┘
```

## Legend

- ✅ **Updated with Connect + Pulsar integration**
- ⚠️ **Interfaces added, but not yet updated** (TODO)
- ⚪ **No changes needed** (no external integration required)

## Flow Example: CreateEntry

```
User Request
    │
    ▼
CreateEntryCommand.Handle()
    │
    ├─── [1] Validate Format (local)
    ├─── [2] Validate Ownership (local)
    ├─── [3] Check Duplicate LOCAL (PostgreSQL)
    │
    ├─── [3a] Check Duplicate GLOBAL ◄─── NEW
    │         │
    │         └─── ConnectClient.GetEntryByKey()
    │              │
    │              └─── gRPC → conn-dict:9092
    │                   │
    │                   └─── Bridge → Bacen DICT
    │                        │
    │                        └─── Returns: exists/not exists
    │                             Latency: ~500ms
    │
    ├─── [4] Validate Limits (max 5 CPF)
    ├─── [5] Create Entity
    ├─── [6] Save to PostgreSQL (local)
    │
    ├─── [7] Publish Event ◄─── UPDATED
    │         │
    │         └─── EntryEventProducer.PublishCreated()
    │              │
    │              └─── Pulsar Topic: dict.entries.created
    │                   │
    │                   └─── conn-dict Consumer
    │                        │
    │                        └─── Temporal ClaimWorkflow
    │                             │
    │                             └─── Bridge → Bacen DICT
    │                                  │
    │                                  └─── Status Update → Core
    │              
    │              (Async, non-blocking)
    │              Latency: ~10ms
    │
    └─── [8] Invalidate Cache
         │
         └─── Redis.Delete("entry:12345678900")
    
Return CreateEntryResult to User
Total Latency: ~516ms (500ms Connect + 16ms local)
```

## Integration Patterns

### Pattern 1: Synchronous gRPC Call (Global Duplicate Check)

```go
// Used in: CreateEntry, VerifyAccount
if h.connectClient != nil {
    existingEntry, err := h.connectClient.GetEntryByKey(ctx, keyValue)
    if err == nil && existingEntry != nil {
        return ErrDuplicateKeyGlobal
    }
}
```

**When to use**: When immediate response is required from RSFN

### Pattern 2: Asynchronous Pulsar Event (State Changes)

```go
// Used in: CreateEntry, UpdateEntry, DeleteEntry
if h.entryProducer != nil {
    go func() {
        bgCtx := context.Background()
        h.eventPublisher.Publish(bgCtx, DomainEvent{...})
    }()
}
```

**When to use**: When RSFN update can be eventual consistent

### Pattern 3: Cache-Aside with Remote Fallback (Query Optimization)

```go
// Used in: GetEntry, VerifyAccount
// 1. Try cache
if cached, found := h.cache.Get(ctx, key); found {
    return cached
}
// 2. Try local DB
entry, err := h.entryRepo.FindByKey(ctx, key)
if err == nil {
    return entry
}
// 3. Fallback to RSFN (optional)
if h.connectClient != nil {
    return h.connectClient.GetEntryByKey(ctx, key)
}
```

**When to use**: When data might exist in RSFN but not locally

---

**Created**: 2025-10-27
**Author**: API Specialist Agent
**Version**: 1.0
