# INT-004: Sequence Diagram - Error Handling

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Tipo**: Sequence Diagram - Error Handling Across Services
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: BACKEND (AI Agent - Backend Specialist)

---

## Sumário Executivo

Este documento apresenta **diagramas de sequência** detalhados para **error handling** (tratamento de erros) em todo o sistema DICT, cobrindo:
- Retry policies
- Circuit breaker patterns
- Dead letter queues (DLQ)
- Compensating transactions
- Idempotency handling

**Baseado em**:
- [TEC-001: Core DICT Specification](../../11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md)
- [TEC-003: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | BACKEND | Versão inicial - Error Handling Sequences |

---

## Índice

1. [Retry Policy Flow](#1-retry-policy-flow)
2. [Circuit Breaker Pattern](#2-circuit-breaker-pattern)
3. [Dead Letter Queue (DLQ)](#3-dead-letter-queue-dlq)
4. [Compensating Transaction](#4-compensating-transaction)
5. [Idempotency Handling](#5-idempotency-handling)
6. [Error Recovery Strategies](#6-error-recovery-strategies)

---

## 1. Retry Policy Flow

### 1.1. Exponential Backoff with Retry

**Scenario**: Entry creation fails due to temporary network error. System retries with exponential backoff.

```mermaid
sequenceDiagram
    participant Client
    participant CoreDICT
    participant Connect
    participant Temporal
    participant Bacen

    Note over Client,Bacen: Attempt 1 - Initial Request

    Client->>CoreDICT: POST /api/v1/keys (Create Entry)
    CoreDICT->>CoreDICT: Validate Input
    CoreDICT->>Connect: Publish to Pulsar (Request)
    Connect->>Temporal: Start EntryCreateWorkflow
    Temporal->>Connect: Execute SyncToBacenActivity
    Connect->>Bacen: CreateEntryRequest (ISPB)

    Note over Bacen: Network Timeout (5s)

    Bacen--xConnect: Timeout Error
    Connect->>Connect: Detect Error (Attempt 1 Failed)

    Note over Connect,Temporal: Retry Policy: max_attempts=3, initial_interval=1s, backoff=2.0

    Connect->>Connect: Wait 1s (Exponential Backoff)

    Note over Client,Bacen: Attempt 2 - First Retry

    Connect->>Bacen: CreateEntryRequest (ISPB) [Retry 1]
    Bacen--xConnect: Connection Refused
    Connect->>Connect: Detect Error (Attempt 2 Failed)
    Connect->>Connect: Wait 2s (Exponential Backoff: 1s * 2^1)

    Note over Client,Bacen: Attempt 3 - Second Retry

    Connect->>Bacen: CreateEntryRequest (ISPB) [Retry 2]
    Bacen->>Connect: 200 OK (Entry Created)
    Connect->>Temporal: Activity Success
    Temporal->>Connect: Workflow Completed
    Connect->>CoreDICT: Publish to Pulsar (Response)
    CoreDICT->>CoreDICT: Update Entry Status (ACTIVE)
    CoreDICT->>Client: 201 Created

    Note over Client,Bacen: Success after 2 retries (total 3 attempts)
```

### 1.2. Retry Exhausted - Move to DLQ

**Scenario**: All retry attempts fail. Message moved to Dead Letter Queue.

```mermaid
sequenceDiagram
    participant CoreDICT
    participant Connect
    participant Temporal
    participant Bacen
    participant DLQ
    participant AlertSystem

    CoreDICT->>Connect: Publish to Pulsar (Request)
    Connect->>Temporal: Start EntryCreateWorkflow
    Temporal->>Connect: Execute SyncToBacenActivity

    Note over Connect,Bacen: Attempt 1

    Connect->>Bacen: CreateEntryRequest
    Bacen--xConnect: 503 Service Unavailable
    Connect->>Connect: Wait 1s

    Note over Connect,Bacen: Attempt 2

    Connect->>Bacen: CreateEntryRequest [Retry 1]
    Bacen--xConnect: 503 Service Unavailable
    Connect->>Connect: Wait 2s

    Note over Connect,Bacen: Attempt 3

    Connect->>Bacen: CreateEntryRequest [Retry 2]
    Bacen--xConnect: 503 Service Unavailable

    Note over Connect: All retries exhausted (3 attempts)

    Connect->>DLQ: Move message to DLQ
    DLQ->>DLQ: Store failed message
    Connect->>Temporal: Activity Failed (Non-Retryable)
    Temporal->>Connect: Workflow Failed
    Connect->>AlertSystem: Send Alert (Workflow Failed)
    AlertSystem->>AlertSystem: Notify on-call engineer

    Connect->>CoreDICT: Publish to Pulsar (Response: ERROR)
    CoreDICT->>CoreDICT: Update Entry Status (PENDING)
    CoreDICT->>CoreDICT: Log error with trace_id

    Note over CoreDICT,AlertSystem: Manual intervention required
```

---

## 2. Circuit Breaker Pattern

### 2.1. Circuit Breaker - CLOSED State (Normal)

**Scenario**: Circuit breaker allows requests through. All requests succeed.

```mermaid
sequenceDiagram
    participant Client
    participant CoreDICT
    participant Connect
    participant CircuitBreaker
    participant Bacen

    Note over CircuitBreaker: State: CLOSED (Healthy)

    Client->>CoreDICT: POST /api/v1/keys
    CoreDICT->>Connect: Publish Request
    Connect->>CircuitBreaker: Check state
    CircuitBreaker->>CircuitBreaker: State = CLOSED
    CircuitBreaker->>Bacen: Forward request
    Bacen->>CircuitBreaker: 200 OK
    CircuitBreaker->>CircuitBreaker: Record success
    CircuitBreaker->>Connect: Return response
    Connect->>CoreDICT: Publish Response
    CoreDICT->>Client: 201 Created

    Note over CircuitBreaker: Success count: 1, Failure count: 0
```

### 2.2. Circuit Breaker - OPEN State (Failed)

**Scenario**: Circuit breaker detects failures and opens. Subsequent requests fail fast.

```mermaid
sequenceDiagram
    participant Client1 as Client 1
    participant Client2 as Client 2
    participant CoreDICT
    participant Connect
    participant CircuitBreaker
    participant Bacen

    Note over CircuitBreaker: State: CLOSED, Failure threshold: 5

    loop 5 consecutive failures
        Client1->>CoreDICT: POST /api/v1/keys
        CoreDICT->>Connect: Publish Request
        Connect->>CircuitBreaker: Check state
        CircuitBreaker->>Bacen: Forward request
        Bacen--xCircuitBreaker: 503 Service Unavailable
        CircuitBreaker->>CircuitBreaker: Record failure
        CircuitBreaker->>Connect: Return error
        Connect->>CoreDICT: Publish Error Response
        CoreDICT->>Client1: 502 Bad Gateway
    end

    Note over CircuitBreaker: Failures: 5, Threshold reached!

    CircuitBreaker->>CircuitBreaker: TRANSITION: CLOSED -> OPEN
    CircuitBreaker->>CircuitBreaker: Start timeout (60s)

    Note over CircuitBreaker: State: OPEN (Blocking requests)

    Client2->>CoreDICT: POST /api/v1/keys
    CoreDICT->>Connect: Publish Request
    Connect->>CircuitBreaker: Check state
    CircuitBreaker->>CircuitBreaker: State = OPEN
    CircuitBreaker->>Connect: FAST FAIL (Circuit Open)
    Connect->>CoreDICT: Publish Error Response
    CoreDICT->>Client2: 503 Service Unavailable (Circuit Open)

    Note over Client2,Bacen: Request rejected immediately (no attempt to Bacen)
    Note over CircuitBreaker: Wait 60s before transitioning to HALF_OPEN
```

### 2.3. Circuit Breaker - HALF_OPEN State (Testing)

**Scenario**: Circuit breaker allows limited requests to test if service recovered.

```mermaid
sequenceDiagram
    participant Client
    participant CoreDICT
    participant Connect
    participant CircuitBreaker
    participant Bacen

    Note over CircuitBreaker: State: OPEN (60s timeout expired)

    CircuitBreaker->>CircuitBreaker: TRANSITION: OPEN -> HALF_OPEN
    CircuitBreaker->>CircuitBreaker: Allow limited requests

    Note over CircuitBreaker: State: HALF_OPEN (Testing recovery)

    Client->>CoreDICT: POST /api/v1/keys
    CoreDICT->>Connect: Publish Request
    Connect->>CircuitBreaker: Check state
    CircuitBreaker->>CircuitBreaker: State = HALF_OPEN
    CircuitBreaker->>Bacen: Forward request (Test)
    Bacen->>CircuitBreaker: 200 OK
    CircuitBreaker->>CircuitBreaker: Record success

    Note over CircuitBreaker: Success in HALF_OPEN!

    CircuitBreaker->>CircuitBreaker: TRANSITION: HALF_OPEN -> CLOSED
    CircuitBreaker->>CircuitBreaker: Reset failure count

    Note over CircuitBreaker: State: CLOSED (Recovered)

    CircuitBreaker->>Connect: Return response
    Connect->>CoreDICT: Publish Response
    CoreDICT->>Client: 201 Created

    Note over Client,Bacen: Circuit breaker recovered, normal operation resumed
```

---

## 3. Dead Letter Queue (DLQ)

### 3.1. DLQ Processing and Reprocessing

**Scenario**: Failed messages in DLQ are manually reviewed and reprocessed.

```mermaid
sequenceDiagram
    participant Admin
    participant AdminAPI
    participant Connect
    participant DLQ
    participant Temporal
    participant Bacen
    participant CoreDICT

    Note over DLQ: Contains 3 failed messages

    Admin->>AdminAPI: GET /admin/v1/metrics/queues
    AdminAPI->>DLQ: Query DLQ metrics
    DLQ->>AdminAPI: pending_messages: 3
    AdminAPI->>Admin: Show DLQ stats

    Admin->>Admin: Investigate errors
    Admin->>AdminAPI: POST /admin/v1/maintenance/reprocess-dlq
    Note right of Admin: {max_messages: 10, filter: {...}}

    AdminAPI->>DLQ: Get messages from DLQ (max 10)
    DLQ->>AdminAPI: Return 3 messages

    loop For each message in DLQ
        AdminAPI->>AdminAPI: Validate message
        AdminAPI->>Connect: Republish message
        Connect->>Temporal: Start workflow
        Temporal->>Connect: Execute activities

        alt Bacen service recovered
            Connect->>Bacen: CreateEntryRequest
            Bacen->>Connect: 200 OK
            Connect->>Temporal: Activity Success
            Temporal->>Connect: Workflow Completed
            Connect->>CoreDICT: Publish Response
            CoreDICT->>CoreDICT: Update entry status
            AdminAPI->>AdminAPI: Mark as SUCCESS
        else Bacen still unavailable
            Connect->>Bacen: CreateEntryRequest
            Bacen--xConnect: 503 Service Unavailable
            Connect->>DLQ: Move back to DLQ
            AdminAPI->>AdminAPI: Mark as FAILED
        end
    end

    AdminAPI->>Admin: Reprocessing results
    Note right of Admin: successful_count: 2<br/>failed_count: 1

    Admin->>Admin: Review failed messages
```

---

## 4. Compensating Transaction

### 4.1. Claim Creation with Rollback

**Scenario**: Claim created successfully in Bacen, but local DB update fails. Compensating transaction cancels claim.

```mermaid
sequenceDiagram
    participant Client
    participant CoreDICT
    participant CoreDB
    participant Connect
    participant Temporal
    participant Bacen

    Note over Client,Bacen: Forward Transaction - Create Claim

    Client->>CoreDICT: POST /api/v1/claims (Create Claim)
    CoreDICT->>CoreDICT: Validate input
    CoreDICT->>CoreDB: BEGIN TRANSACTION
    CoreDB->>CoreDICT: OK

    CoreDICT->>CoreDB: INSERT INTO claims (...)
    CoreDB->>CoreDICT: Claim ID: claim-123

    CoreDICT->>Connect: Publish to Pulsar (Request)
    Connect->>Temporal: Start ClaimCreateWorkflow
    Temporal->>Connect: Execute SyncClaimToBacenActivity
    Connect->>Bacen: CreateClaimRequest (ISPB)
    Bacen->>Connect: 200 OK (Claim Created)
    Note right of Bacen: External Claim ID: bacen-claim-456

    Connect->>CoreDICT: Publish to Pulsar (Response)

    CoreDICT->>CoreDB: UPDATE claims SET external_id = 'bacen-claim-456'

    Note over CoreDB: Database connection lost!

    CoreDB--xCoreDICT: Connection Error

    Note over CoreDICT,Bacen: Compensating Transaction Required!

    CoreDICT->>CoreDB: ROLLBACK TRANSACTION
    CoreDB->>CoreDICT: OK

    CoreDICT->>Connect: Publish to Pulsar (Compensate)
    Connect->>Temporal: Start ClaimCancelWorkflow (Compensate)
    Temporal->>Connect: Execute CancelClaimActivity
    Connect->>Bacen: CancelClaimRequest (bacen-claim-456)
    Bacen->>Connect: 200 OK (Claim Cancelled)
    Connect->>CoreDICT: Publish Response (Compensated)

    CoreDICT->>Client: 500 Internal Server Error
    Note right of Client: error: "Failed to create claim.<br/>Compensating transaction executed."

    Note over Client,Bacen: System state consistent (claim cancelled in both systems)
```

### 4.2. Saga Pattern - Distributed Transaction

**Scenario**: Multi-step distributed transaction with compensation.

```mermaid
sequenceDiagram
    participant Client
    participant CoreDICT
    participant Connect
    participant Temporal
    participant Bacen
    participant AuditService

    Note over Client,AuditService: Saga: Entry Creation (3 steps)

    Client->>CoreDICT: POST /api/v1/keys
    CoreDICT->>Connect: Publish Request
    Connect->>Temporal: Start EntryCreateSaga

    Note over Temporal: Step 1: Create local entry

    Temporal->>CoreDICT: CreateLocalEntry
    CoreDICT->>CoreDICT: INSERT INTO entries
    CoreDICT->>Temporal: Success

    Note over Temporal: Step 2: Sync to Bacen

    Temporal->>Connect: SyncToBacenActivity
    Connect->>Bacen: CreateEntryRequest
    Bacen->>Connect: 200 OK
    Connect->>Temporal: Success

    Note over Temporal: Step 3: Log audit event

    Temporal->>AuditService: LogEntryCreatedEvent
    AuditService--xTemporal: Service Unavailable

    Note over Temporal: Step 3 FAILED - Trigger compensation

    Note over Temporal,Bacen: Compensate Step 2

    Temporal->>Connect: CompensateSyncToBacen
    Connect->>Bacen: DeleteEntryRequest
    Bacen->>Connect: 200 OK (Entry Deleted)
    Connect->>Temporal: Compensated

    Note over Temporal,CoreDICT: Compensate Step 1

    Temporal->>CoreDICT: CompensateCreateLocalEntry
    CoreDICT->>CoreDICT: DELETE FROM entries
    CoreDICT->>Temporal: Compensated

    Temporal->>Connect: Saga Failed (All compensated)
    Connect->>CoreDICT: Publish Error Response
    CoreDICT->>Client: 500 Internal Server Error

    Note over Client,AuditService: All steps compensated - system consistent
```

---

## 5. Idempotency Handling

### 5.1. Duplicate Request Detection

**Scenario**: Client sends duplicate request with same idempotency key. System detects and returns cached response.

```mermaid
sequenceDiagram
    participant Client
    participant CoreDICT
    participant IdempotencyCache
    participant CoreDB
    participant Connect

    Note over Client,Connect: Request 1 - Original Request

    Client->>CoreDICT: POST /api/v1/keys<br/>{idempotency_key: "key-123"}
    CoreDICT->>IdempotencyCache: Check key-123
    IdempotencyCache->>CoreDICT: NOT FOUND

    CoreDICT->>IdempotencyCache: Lock key-123 (30s TTL)
    IdempotencyCache->>CoreDICT: Locked

    CoreDICT->>CoreDB: INSERT INTO entries
    CoreDB->>CoreDICT: Entry created (entry-uuid-1)

    CoreDICT->>Connect: Publish Request
    Connect->>Connect: Process workflow
    Connect->>CoreDICT: Publish Response (Success)

    CoreDICT->>IdempotencyCache: Store result for key-123
    Note right of IdempotencyCache: {status: 201, entry_id: entry-uuid-1}<br/>TTL: 24h

    CoreDICT->>Client: 201 Created {entry_id: entry-uuid-1}

    Note over Client,Connect: Request 2 - Duplicate Request (Network Retry)

    Client->>CoreDICT: POST /api/v1/keys<br/>{idempotency_key: "key-123"}
    CoreDICT->>IdempotencyCache: Check key-123
    IdempotencyCache->>CoreDICT: FOUND (cached result)

    Note over CoreDICT: Duplicate detected!

    CoreDICT->>Client: 201 Created {entry_id: entry-uuid-1}
    Note right of Client: Same response as original

    Note over Client,Connect: No duplicate creation, idempotency preserved
```

### 5.2. Idempotency with Conflict

**Scenario**: Different request with same idempotency key. System rejects as conflict.

```mermaid
sequenceDiagram
    participant Client
    participant CoreDICT
    participant IdempotencyCache
    participant CoreDB

    Note over Client,CoreDB: Request 1 - Original Request

    Client->>CoreDICT: POST /api/v1/keys<br/>{idempotency_key: "key-456", key_value: "111"}
    CoreDICT->>IdempotencyCache: Check key-456
    IdempotencyCache->>CoreDICT: NOT FOUND

    CoreDICT->>IdempotencyCache: Store request hash
    Note right of IdempotencyCache: {request_hash: hash("111")}<br/>TTL: 24h

    CoreDICT->>CoreDB: INSERT INTO entries (key_value: "111")
    CoreDB->>CoreDICT: Entry created
    CoreDICT->>Client: 201 Created

    Note over Client,CoreDB: Request 2 - Different Request, Same Idempotency Key

    Client->>CoreDICT: POST /api/v1/keys<br/>{idempotency_key: "key-456", key_value: "222"}
    CoreDICT->>IdempotencyCache: Check key-456
    IdempotencyCache->>CoreDICT: FOUND

    CoreDICT->>CoreDICT: Compare request hash
    Note right of CoreDICT: Stored hash: hash("111")<br/>New hash: hash("222")<br/>MISMATCH!

    CoreDICT->>Client: 409 Conflict<br/>{error: "Idempotency key conflict"}
    Note right of Client: Same idempotency key<br/>but different request data

    Note over Client,CoreDB: Request rejected to maintain idempotency guarantee
```

---

## 6. Error Recovery Strategies

### 6.1. Timeout Handling with Graceful Degradation

**Scenario**: Bacen DICT is slow. System times out and returns partial success.

```mermaid
sequenceDiagram
    participant Client
    participant CoreDICT
    participant CoreDB
    participant Connect
    participant Temporal
    participant Bacen

    Note over Client,Bacen: Context: Bacen DICT is slow (>30s response time)

    Client->>CoreDICT: POST /api/v1/keys<br/>(timeout: 30s)
    CoreDICT->>CoreDB: INSERT INTO entries (status: PENDING)
    CoreDB->>CoreDICT: Entry created (entry-uuid-1)

    CoreDICT->>Connect: Publish Request (async)
    Connect->>Temporal: Start EntryCreateWorkflow<br/>(timeout: 300s)

    Note over CoreDICT: Return early to client (degraded mode)

    CoreDICT->>Client: 202 Accepted<br/>{entry_id: entry-uuid-1, status: PENDING}
    Note right of Client: Client receives partial success<br/>Can poll for final status

    Note over Connect,Bacen: Background processing continues

    Temporal->>Connect: Execute SyncToBacenActivity
    Connect->>Bacen: CreateEntryRequest

    Note over Bacen: Slow response (45s)

    Bacen->>Connect: 200 OK (after 45s)
    Connect->>Temporal: Activity Success
    Temporal->>Connect: Workflow Completed

    Connect->>CoreDICT: Publish Response (async)
    CoreDICT->>CoreDB: UPDATE entries SET status = ACTIVE
    CoreDB->>CoreDICT: Updated

    Note over Client,Bacen: Client can poll: GET /api/v1/keys/entry-uuid-1
```

### 6.2. Fallback Strategy

**Scenario**: Primary operation fails. System uses fallback approach.

```mermaid
sequenceDiagram
    participant Client
    participant CoreDICT
    participant PrimaryCRUD
    participant CacheService
    participant FallbackCRUD

    Note over Client,FallbackCRUD: Primary path: Database query

    Client->>CoreDICT: GET /api/v1/keys/12345678901
    CoreDICT->>CacheService: Check cache
    CacheService->>CoreDICT: MISS

    CoreDICT->>PrimaryCRUD: FindByKeyValue("12345678901")
    PrimaryCRUD--xCoreDICT: Database Connection Error

    Note over CoreDICT: Primary path failed - use fallback

    CoreDICT->>FallbackCRUD: FindByKeyValueFromReadReplica
    FallbackCRUD->>CoreDICT: Entry found (slightly stale data)

    Note over CoreDICT: Log degraded mode

    CoreDICT->>CoreDICT: Log warning: Using read replica

    CoreDICT->>Client: 200 OK {entry, X-Data-Source: replica}
    Note right of Client: Response includes header<br/>indicating fallback was used

    Note over Client,FallbackCRUD: Request succeeded with fallback strategy
```

---

## Rastreabilidade

### Requisitos Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RF-ERR-001 | Retry policy com exponential backoff | [TEC-003](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md) | ✅ Especificado |
| RF-ERR-002 | Circuit breaker pattern | [TEC-003](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md) | ✅ Especificado |
| RF-ERR-003 | Dead letter queue processing | [TEC-003](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md) | ✅ Especificado |
| RF-ERR-004 | Compensating transactions | [TEC-001](../../11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md) | ✅ Especificado |
| RF-ERR-005 | Idempotency handling | [API-002](../../04_APIs/REST/API-002_Core_DICT_REST_API.md) | ✅ Especificado |

---

## Próximas Revisões

**Pendências**:
- [ ] Adicionar sequence para rate limiting
- [ ] Adicionar sequence para chaos engineering tests
- [ ] Adicionar sequence para disaster recovery

---

**Referências**:
- [TEC-001: Core DICT Specification](../../11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md)
- [TEC-003: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [API-002: Core DICT REST API](../../04_APIs/REST/API-002_Core_DICT_REST_API.md)
- [Temporal Retry Policies](https://docs.temporal.io/retry-policies)
- [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html)
- [Saga Pattern](https://microservices.io/patterns/data/saga.html)
