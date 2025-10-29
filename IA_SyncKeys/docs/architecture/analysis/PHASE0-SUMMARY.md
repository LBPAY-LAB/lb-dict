# Phase 0: Technical Analysis Summary

## Executive Summary

Completed comprehensive analysis of connector-dict codebase to understand patterns and prepare for VSync implementation. Key finding: **Entry events are NOT currently being published to the dict-events topic**, requiring implementation of event publishing mechanism before VSync can consume them.

## Critical Findings

### 🔴 Entry Events Not Published
- Current state: Dict API makes direct gRPC calls to Bridge for Entry operations
- No Temporal workflows exist for Entry operations
- No events published to `persistent://lb-conn/dict/dict-events` for Entry operations
- Only Claim operations currently publish events

### 🟡 Database Architecture Different Than Expected
- No direct PostgreSQL usage in connector-dict (only for Temporal)
- All data persistence through BACEN Bridge (source of truth)
- Redis used only for caching and idempotency
- VSync will need its own PostgreSQL database for CID storage

### 🟢 Data Model Complete for CID
- Entry domain model contains all required fields
- Data already normalized and validated
- Direct mapping to BACEN SDK types
- Can be reused in VSync implementation

## Architecture Decisions Required

### Decision 1: How to Publish Entry Events?

**Option A: Modify Dict API (Recommended)**
```go
// In dict/application/entry/application.go
func (app *Application) CreateEntry(ctx context.Context, entry CreateEntryRequest) (*CreateEntryResponse, error) {
    // Existing: Call Bridge via gRPC
    resp, err := app.dir.CreateEntry(ctx, payload)

    // NEW: Publish event
    if err == nil {
        app.publisher.Publish(ctx, "key.created", resp)
    }

    return resp, err
}
```

**Option B: Create Entry Workflows**
- Add to orchestration-worker like Claims
- More complex, but consistent with existing async patterns
- Requires significant refactoring

**Option C: Publish from Bridge Response Handler**
- Intercept Bridge responses
- Add event publishing layer
- Minimal changes to existing code

### Decision 2: VSync Container Architecture

**Approved Structure:**
```
apps/dict.vsync/
├── cmd/
│   └── main.go                    # Application entry point
├── application/
│   ├── usecases/
│   │   ├── process_entry.go      # Process Entry events
│   │   ├── generate_cid.go       # CID generation logic
│   │   └── sync_batch.go         # Daily batch sync
│   └── ports/
│       ├── repository.go         # Database interface
│       └── publisher.go          # Event publisher interface
├── domain/
│   ├── cid/
│   │   ├── cid.go               # CID entity
│   │   └── generator.go         # CID generation algorithm
│   └── entry/                    # Import from dict
├── handlers/
│   └── pulsar/
│       └── entry_handler.go      # Pulsar event consumer
├── infrastructure/
│   ├── database/
│   │   ├── postgres.go          # PostgreSQL connection
│   │   ├── migrations/          # Database migrations
│   │   └── repository.go        # CID repository impl
│   ├── temporal/
│   │   ├── activities/
│   │   │   ├── generate_cid.go
│   │   │   └── sync_bacen.go
│   │   └── workflows/
│   │       └── vsync_workflow.go
│   └── pulsar/
│       └── consumer.go           # Event consumer setup
└── setup/
    ├── config.go                 # Configuration
    └── setup.go                  # Dependency injection
```

## Implementation Roadmap

### Phase 1: Enable Entry Event Publishing (Priority 1)
1. **Modify Dict API** to publish Entry events
2. **Define event schema** for Entry operations
3. **Test event publishing** with mock consumer
4. **Deploy changes** to Dict API

### Phase 2: VSync Core Implementation
1. **Setup container structure** (`apps/dict.vsync/`)
2. **Implement Pulsar consumer** for Entry events
3. **Create CID generation** algorithm per BACEN spec
4. **Setup PostgreSQL** database and migrations
5. **Implement repositories** for CID persistence

### Phase 3: Temporal Integration
1. **Create VSyncWorkflow** for orchestration
2. **Implement activities** (GenerateCID, StoreCID, SyncBACEN)
3. **Add retry policies** and error handling
4. **Setup monitoring** and alerting

### Phase 4: Batch Synchronization
1. **Implement daily batch** job for BACEN sync
2. **Create batch workflow** in Temporal
3. **Add reconciliation** logic
4. **Implement audit trail**

## Technical Stack Confirmation

### Approved Technologies
- **Language**: Go 1.24.5
- **Message Broker**: Apache Pulsar
- **Database**: PostgreSQL 15+ with pgx/v5
- **Cache**: Redis 7.2
- **Workflow Engine**: Temporal
- **Migration Tool**: golang-migrate
- **Observability**: OpenTelemetry

### Environment Configuration
```env
# VSync Service
VSYNC_SERVICE_NAME=dict-vsync
VSYNC_SERVICE_VERSION=1.0.0

# Pulsar
PULSAR_URL=pulsar://localhost:6650
PULSAR_TOPIC_DICT_EVENTS=persistent://lb-conn/dict/dict-events
PULSAR_SUBSCRIPTION=vsync-subscription

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=vsync
DB_PASSWORD=vsync123
DB_NAME=vsync
DB_SSL_MODE=disable

# Redis
REDIS_ADDR=localhost:6379
REDIS_DB=1
REDIS_PREFIX=vsync:

# Temporal
TEMPORAL_URL=localhost:7233
TEMPORAL_NAMESPACE=default
TEMPORAL_TASK_QUEUE=vsync-tasks
```

## Database Schema

### Core Tables
```sql
-- CID storage (main table)
CREATE TABLE cids (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_value       VARCHAR(77) NOT NULL,
    key_type        VARCHAR(10) NOT NULL,
    cid             VARCHAR(255) NOT NULL UNIQUE,
    ispb            VARCHAR(8) NOT NULL,
    branch          VARCHAR(10),
    account_number  VARCHAR(20) NOT NULL,
    account_type    VARCHAR(4) NOT NULL,
    tax_id          VARCHAR(14) NOT NULL,
    person_type     VARCHAR(20) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    sync_status     VARCHAR(20) NOT NULL DEFAULT 'pending',
    synced_at       TIMESTAMP,
    deleted_at      TIMESTAMP,
    UNIQUE(key_value, ispb)
);

-- Event processing tracking
CREATE TABLE processed_events (
    correlation_id  VARCHAR(255) PRIMARY KEY,
    action          VARCHAR(50) NOT NULL,
    processed_at    TIMESTAMP NOT NULL DEFAULT NOW()
);
```

## Event Schema Definition

### Entry Event Structure
```json
{
  "properties": {
    "correlation_id": "uuid",
    "action": "key.created|key.updated|key.deleted"
  },
  "payload": {
    "entry": {
      "key": "string",
      "keyType": "CPF|CNPJ|PHONE|EMAIL|EVP",
      "account": {
        "participant": "string(8)",
        "branch": "string|null",
        "accountNumber": "string",
        "accountType": "CACC|SVGS",
        "openingDate": "ISO8601"
      },
      "owner": {
        "type": "NATURAL_PERSON|LEGAL_PERSON",
        "taxIDNumber": "string",
        "name": "string",
        "tradeName": "string|null"
      },
      "creationDate": "ISO8601",
      "keyOwnershipDate": "ISO8601"
    },
    "correlationId": "uuid",
    "responseTime": "ISO8601"
  }
}
```

## CID Generation Algorithm

Based on BACEN specification:
```go
func GenerateCID(entry *domain.Entry) string {
    // Concatenate normalized fields
    data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%d",
        entry.Key.Value,
        entry.Key.Type,
        entry.Account.Participant,
        entry.Account.Branch,
        entry.Account.AccountNumber,
        entry.Account.AccountType,
        entry.Owner.TaxIDNumber,
        entry.CreationDate.Unix())

    // Generate SHA-256 hash
    hash := sha256.Sum256([]byte(data))

    // Return hex-encoded CID
    return hex.EncodeToString(hash[:])
}
```

## Risk Assessment

### High Risk Items
1. **Entry events not published** - Must be implemented first
2. **No existing database layer** - Need to build from scratch
3. **BACEN integration unclear** - Need clarification on sync API

### Medium Risk Items
1. **Event schema changes** - May affect other consumers
2. **Performance at scale** - Need load testing
3. **Temporal workflow complexity** - Learning curve

### Low Risk Items
1. **Domain model reuse** - Well-defined and tested
2. **Pulsar integration** - Pattern already established
3. **Redis caching** - Proven pattern

## Success Criteria

### Phase 0 Complete ✅
- [x] Event schema analyzed
- [x] Entry domain model understood
- [x] Existing patterns documented
- [x] Database requirements defined
- [x] Architecture decisions documented

### Next Steps
1. **Get stakeholder approval** on Entry event publishing approach
2. **Create detailed technical specification** for VSync
3. **Set up development environment** with all dependencies
4. **Begin Phase 1 implementation** (Entry event publishing)

## Recommendations

### Immediate Actions
1. **Priority 1**: Implement Entry event publishing in Dict API
2. **Priority 2**: Set up VSync container structure
3. **Priority 3**: Implement basic event consumer

### Architecture Guidelines
1. **Follow existing patterns** from connector-dict
2. **Use clean architecture** with clear layer separation
3. **Implement comprehensive observability** from day one
4. **Write tests alongside implementation**
5. **Document all decisions** and trade-offs

### Performance Considerations
1. **Batch CID generation** for efficiency
2. **Use connection pooling** for PostgreSQL
3. **Implement caching** for frequently accessed CIDs
4. **Async processing** via Temporal for heavy operations
5. **Monitor and alert** on performance metrics

## Conclusion

The technical analysis reveals that while the connector-dict provides excellent patterns to follow, the VSync implementation faces a critical dependency: **Entry events are not currently being published**. This must be resolved before VSync development can proceed.

Once Entry events are available, the VSync system can be built following the established patterns with confidence. The architecture is sound, the data model is complete, and the implementation path is clear.

**Recommended Next Step**: Schedule meeting with stakeholders to decide on Entry event publishing strategy (Option A, B, or C) and get approval to proceed.

---

**Analysis Completed**: 2024-10-29
**Analyst**: Backend System Architect
**Status**: Ready for stakeholder review