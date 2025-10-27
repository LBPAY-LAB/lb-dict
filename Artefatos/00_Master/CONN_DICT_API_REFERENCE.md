# conn-dict API Reference

**Version**: 1.0
**Date**: 2025-10-27
**Status**: PRODUCTION-READY

---

## Table of Contents

1. [Overview](#overview)
2. [Service Endpoints](#service-endpoints)
3. [gRPC Services](#grpc-services)
   - [EntryService](#entryservice)
   - [ClaimService](#claimservice)
   - [InfractionService](#infractionservice)
4. [Pulsar Topics](#pulsar-topics)
   - [Input Topics](#input-topics)
   - [Output Topics](#output-topics)
5. [Temporal Workflows](#temporal-workflows)
6. [Usage Examples](#usage-examples)
7. [Health & Metrics](#health--metrics)
8. [Error Handling](#error-handling)
9. [Security & Authentication](#security--authentication)

---

## Overview

**conn-dict** is the RSFN Connect service responsible for synchronizing DICT PIX key operations with the Central Bank of Brazil (Bacen) RSFN network. It acts as an orchestrator between core-dict (business logic) and conn-bridge (XML adapter).

**Purpose**:
- Process PIX key operations (create, update, delete) asynchronously
- Manage 30-day Claims and Infraction workflows
- Synchronize state with Bacen via VSYNC daily batches
- Provide query APIs for entries, claims, and infractions

**Architecture**:
- **gRPC** for synchronous queries and workflow triggers
- **Pulsar** for asynchronous event-driven operations
- **Temporal** for durable long-running workflows (30+ days)
- **PostgreSQL** for persistent state storage

---

## Service Endpoints

### Network Configuration

| Service | Port | Protocol | Purpose |
|---------|------|----------|---------|
| gRPC Server | **9092** | gRPC | API calls (16 RPCs) |
| Health Check | **8080** | HTTP | Liveness/Readiness probes |
| Metrics | **9091** | HTTP | Prometheus metrics |
| Temporal Worker | N/A | Internal | Workflow/Activity execution |
| Pulsar Consumer | N/A | Internal | Event consumption |

### Environment Variables

```bash
# gRPC Server
GRPC_PORT=9092

# HTTP Server (Health + Metrics)
HTTP_PORT=8080
METRICS_PORT=9091

# Database
DATABASE_URL=postgres://user:pass@localhost:5432/conn_dict?sslmode=disable

# Pulsar
PULSAR_URL=pulsar://localhost:6650
PULSAR_SUBSCRIPTION=conn-dict-consumer

# Temporal
TEMPORAL_HOST=localhost:7233
TEMPORAL_NAMESPACE=default
TEMPORAL_TASK_QUEUE=dict-task-queue

# Bridge gRPC Client
BRIDGE_GRPC_URL=localhost:9094

# Logging
LOG_LEVEL=info
```

---

## gRPC Services

conn-dict exposes **16 gRPC RPCs** across 3 services:

### EntryService

Handles PIX key entry queries. **Note**: Entry creation/update/delete are asynchronous via Pulsar, not RPC calls.

#### RPCs (3 Query Methods)

##### 1. GetEntry

Retrieve a single entry by ID or key.

**Request**:
```protobuf
{
  "entry_id": "uuid-string",  // OR
  "key": "email@example.com"
}
```

**Response**:
```json
{
  "id": "uuid",
  "entry_id": "external-uuid",
  "key": "email@example.com",
  "key_type": "EMAIL",
  "participant": "12345678",
  "account_branch": "0001",
  "account_number": "123456",
  "account_type": "CACC",
  "owner_type": "NATURAL_PERSON",
  "owner_name": "John Doe",
  "owner_tax_id": "12345678900",
  "status": "ACTIVE",
  "bacen_entry_id": "bacen-id",
  "created_at": "2025-10-27T10:00:00Z",
  "activated_at": "2025-10-27T10:00:02Z",
  "updated_at": "2025-10-27T10:00:02Z"
}
```

**Error Codes**:
- `INVALID_ARGUMENT`: Neither entry_id nor key provided
- `NOT_FOUND`: Entry does not exist
- `INTERNAL`: Database query failed

---

##### 2. GetEntryByKey

Convenience method to retrieve entry by PIX key (alias for GetEntry with key parameter).

**Request**:
```protobuf
{
  "key": "+5511987654321"
}
```

**Response**: Same as GetEntry

---

##### 3. ListEntries

List entries for a specific participant (bank) with pagination.

**Request**:
```protobuf
{
  "participant_ispb": "12345678",  // Required (8 digits)
  "limit": 20,                      // Optional (default: 20, max: 100)
  "offset": 0                       // Optional (default: 0)
}
```

**Response**:
```json
{
  "entries": [
    {
      "id": "uuid",
      "entry_id": "external-uuid",
      "key": "key1@example.com",
      "status": "ACTIVE",
      ...
    },
    {
      "id": "uuid2",
      "entry_id": "external-uuid2",
      "key": "key2@example.com",
      "status": "PENDING",
      ...
    }
  ],
  "total_count": 2,
  "limit": 20,
  "offset": 0
}
```

**Error Codes**:
- `INVALID_ARGUMENT`: participant_ispb missing or invalid
- `INTERNAL`: Database query failed

**Security Note**: Users can ONLY query entries for their own ISPB (enforced by authentication middleware).

---

### ClaimService

Handles PIX key portability/ownership claims (30-day workflows).

#### RPCs (5 Methods)

##### 1. CreateClaim

Initiate a new claim for PIX key ownership or portability. Starts a durable 30-day Temporal workflow.

**Request**:
```protobuf
{
  "claim_id": "uuid",              // Optional (auto-generated if not provided)
  "entry_id": "entry-uuid",        // Required
  "claim_type": "OWNERSHIP",       // Required: OWNERSHIP or PORTABILITY
  "claimer_ispb": "87654321",      // Required (8 digits)
  "donor_ispb": "12345678",        // Required (8 digits, must differ from claimer)
  "claimer_account": "0001-123456", // Optional
  "requested_by": "user@bank.com"  // Optional
}
```

**Response**:
```json
{
  "claim_id": "uuid",
  "entry_id": "entry-uuid",
  "workflow_id": "claim-workflow-uuid",
  "run_id": "temporal-run-id",
  "status": "OPEN",
  "claim_type": "OWNERSHIP",
  "claimer_ispb": "87654321",
  "donor_ispb": "12345678",
  "expires_at": "2025-11-26T10:00:00Z",  // 30 days from now
  "message": "Claim created successfully. Donor has 30 days to respond."
}
```

**Workflow Behavior**:
- **0-7 days**: Donor can confirm or reject claim
- **7 days (no response)**: Auto-approve claim
- **30 days**: Claim expires if not completed

**Error Codes**:
- `INVALID_ARGUMENT`: Missing required fields, invalid ISPB, same claimer/donor
- `ALREADY_EXISTS`: Active claim already exists for this entry
- `INTERNAL`: Temporal workflow start failed

---

##### 2. ConfirmClaim

Donor confirms the claim (transfers PIX key ownership to claimer). Sends a Signal to Temporal workflow.

**Request**:
```protobuf
{
  "claim_id": "uuid",              // Required
  "confirmed_by": "user@bank.com"  // Optional
}
```

**Response**:
```json
{
  "claim_id": "uuid",
  "status": "CONFIRMED",
  "message": "Claim confirmation signal sent successfully. Workflow will complete the claim."
}
```

**Error Codes**:
- `INVALID_ARGUMENT`: claim_id missing
- `NOT_FOUND`: Claim does not exist
- `FAILED_PRECONDITION`: Claim already completed/expired or not in confirmable state
- `INTERNAL`: Temporal signal failed

---

##### 3. CancelClaim

Cancel an active claim (can be called by claimer or donor). Sends a Signal to Temporal workflow.

**Request**:
```protobuf
{
  "claim_id": "uuid",                       // Required
  "reason": "Customer request",             // Required
  "cancelled_by": "user@bank.com"           // Optional
}
```

**Response**:
```json
{
  "claim_id": "uuid",
  "status": "CANCELLED",
  "reason": "Customer request",
  "message": "Claim cancelled by user@bank.com. Reason: Customer request"
}
```

**Error Codes**:
- `INVALID_ARGUMENT`: claim_id or reason missing
- `NOT_FOUND`: Claim does not exist
- `FAILED_PRECONDITION`: Claim cannot be cancelled (already completed)
- `INTERNAL`: Temporal signal failed

---

##### 4. GetClaim

Retrieve claim details by claim ID.

**Request**:
```protobuf
{
  "claim_id": "uuid"
}
```

**Response**:
```json
{
  "id": "uuid",
  "claim_id": "external-uuid",
  "type": "OWNERSHIP",
  "status": "OPEN",
  "key": "+5511987654321",
  "key_type": "PHONE",
  "donor_participant": "12345678",
  "claimer_participant": "87654321",
  "claimer_account_branch": "0001",
  "claimer_account_number": "123456",
  "claimer_account_type": "CACC",
  "completion_period_end": "2025-11-03T10:00:00Z",  // 7 days from creation
  "claim_expiry_date": "2025-11-26T10:00:00Z",      // 30 days from creation
  "created_at": "2025-10-27T10:00:00Z",
  "updated_at": "2025-10-27T10:00:00Z",
  "is_expired": false,
  "is_active": true,
  "can_be_cancelled": true
}
```

**Error Codes**:
- `INVALID_ARGUMENT`: claim_id missing
- `NOT_FOUND`: Claim does not exist
- `INTERNAL`: Database query failed

---

##### 5. ListClaims

List claims with optional filtering.

**Request**:
```protobuf
{
  "key": "+5511987654321",  // Required (for now)
  "limit": 20,               // Optional (default: 20, max: 100)
  "offset": 0                // Optional (default: 0)
}
```

**Response**:
```json
{
  "claims": [
    {
      "id": "uuid",
      "claim_id": "external-uuid",
      "type": "PORTABILITY",
      "status": "CONFIRMED",
      ...
    }
  ],
  "total_count": 1,
  "limit": 20,
  "offset": 0
}
```

**Error Codes**:
- `INVALID_ARGUMENT`: key filter missing (required for security)
- `INTERNAL`: Database query failed

**Note**: Full listing (without key filter) requires ListByParticipant method (not yet implemented).

---

### InfractionService

Handles fraud reports and infraction investigations.

#### RPCs (6 Methods)

##### 1. CreateInfraction

Report a fraud or infraction. Starts an investigation workflow (human-in-the-loop).

**Request**:
```protobuf
{
  "infraction_id": "uuid",           // Optional (auto-generated)
  "key": "+5511987654321",           // Required
  "type": "FRAUD",                   // Required: FRAUD, ACCOUNT_CLOSED, INCORRECT_DATA,
                                     //           UNAUTHORIZED_USE, DUPLICATE_KEY, OTHER
  "description": "Fraudulent key",   // Required
  "reporter_ispb": "12345678",       // Required (8 digits)
  "reported_ispb": "87654321",       // Optional (8 digits, must differ from reporter)
  "evidence_urls": [                 // Optional
    "https://evidence.com/file1.pdf"
  ],
  "related_entry_id": "entry-uuid",  // Optional
  "related_claim_id": "claim-uuid"   // Optional
}
```

**Response**:
```json
{
  "infraction_id": "uuid",
  "workflow_id": "infraction-uuid",
  "run_id": "temporal-run-id",
  "status": "OPEN",
  "message": "Infraction created successfully. Investigation workflow started."
}
```

**Infraction Types**:
- **FRAUD**: Fraudulent PIX key usage
- **ACCOUNT_CLOSED**: Account associated with key is closed
- **INCORRECT_DATA**: Key data is incorrect or outdated
- **UNAUTHORIZED_USE**: Unauthorized use of PIX key
- **DUPLICATE_KEY**: Duplicate key detected
- **OTHER**: Other infraction types

**Error Codes**:
- `INVALID_ARGUMENT`: Missing required fields, invalid type/ISPB
- `INTERNAL`: Temporal workflow start failed

---

##### 2. InvestigateInfraction

Submit investigation decision (RESOLVE, DISMISS, ESCALATE). Sends a Signal to workflow.

**Request**:
```protobuf
{
  "infraction_id": "uuid",           // Required
  "decision": "RESOLVE",             // Required: RESOLVE, DISMISS, ESCALATE
  "notes": "Investigation findings"  // Required
}
```

**Response**:
```json
{
  "infraction_id": "uuid",
  "decision": "RESOLVE",
  "message": "Investigation decision 'RESOLVE' sent successfully"
}
```

**Error Codes**:
- `INVALID_ARGUMENT`: Missing required fields, invalid decision
- `NOT_FOUND`: Workflow not found
- `INTERNAL`: Temporal signal failed

---

##### 3. ResolveInfraction

Convenience RPC to resolve an infraction (alias for InvestigateInfraction with decision=RESOLVE).

**Request**:
```protobuf
{
  "infraction_id": "uuid",
  "resolution_notes": "Resolved after investigation"
}
```

**Response**:
```json
{
  "infraction_id": "uuid",
  "status": "RESOLVED",
  "message": "Infraction resolved successfully"
}
```

---

##### 4. DismissInfraction

Convenience RPC to dismiss an infraction (alias for InvestigateInfraction with decision=DISMISS).

**Request**:
```protobuf
{
  "infraction_id": "uuid",
  "dismissal_notes": "Not a valid infraction"
}
```

**Response**:
```json
{
  "infraction_id": "uuid",
  "status": "DISMISSED",
  "message": "Infraction dismissed successfully"
}
```

---

##### 5. GetInfraction

Retrieve infraction details by ID.

**Request**:
```protobuf
{
  "infraction_id": "uuid"
}
```

**Response**:
```json
{
  "id": "uuid",
  "infraction_id": "external-uuid",
  "key": "+5511987654321",
  "type": "FRAUD",
  "description": "Fraudulent key usage detected",
  "evidence_urls": ["https://..."],
  "reporter_participant": "12345678",
  "reported_participant": "87654321",
  "status": "UNDER_INVESTIGATION",
  "entry_id": "entry-uuid",
  "claim_id": "claim-uuid",
  "resolution_notes": null,
  "reported_at": "2025-10-27T10:00:00Z",
  "investigated_at": null,
  "resolved_at": null,
  "created_at": "2025-10-27T10:00:00Z",
  "updated_at": "2025-10-27T10:00:00Z"
}
```

**Error Codes**:
- `INVALID_ARGUMENT`: infraction_id missing
- `NOT_FOUND`: Infraction does not exist
- `INTERNAL`: Database query failed

---

##### 6. ListInfractions

List infractions with optional filtering.

**Request**:
```protobuf
{
  "key": "+5511987654321",      // Optional (filter by key)
  "reporter_ispb": "12345678",  // Optional (filter by reporter, 8 digits)
  "status": "OPEN",             // Optional (filter by status)
  "limit": 20,                  // Optional (default: 20, max: 100)
  "offset": 0                   // Optional (default: 0)
}
```

**Response**:
```json
{
  "infractions": [
    {
      "id": "uuid",
      "infraction_id": "external-uuid",
      "key": "+5511987654321",
      "type": "FRAUD",
      "status": "OPEN",
      ...
    }
  ],
  "total_count": 1,
  "limit": 20,
  "offset": 0
}
```

**Default Behavior**: If no filters provided, returns open infractions (status=OPEN).

**Error Codes**:
- `INVALID_ARGUMENT`: Invalid reporter_ispb format
- `INTERNAL`: Database query failed

---

## Pulsar Topics

conn-dict uses Apache Pulsar for asynchronous event-driven communication with core-dict.

### Input Topics (core-dict → conn-dict)

These topics are **consumed** by conn-dict's Pulsar Consumer.

#### 1. dict.entries.created

Triggered when core-dict creates a new PIX key entry.

**Event Schema**:
```json
{
  "entry_id": "uuid",
  "key": "email@example.com",
  "key_type": "EMAIL",
  "participant": "12345678",
  "account_branch": "0001",
  "account_number": "123456",
  "account_type": "CACC",
  "account_opened_date": "2020-01-01",
  "owner_type": "NATURAL_PERSON",
  "owner_name": "John Doe",
  "owner_tax_id": "12345678900",
  "idempotency_key": "unique-key",
  "request_id": "request-uuid",
  "timestamp": "2025-10-27T10:00:00Z"
}
```

**Consumer Behavior**:
1. Receive event from Pulsar
2. Fetch full entry from database
3. Call Bridge gRPC `CreateEntry` (synchronous, <1.5s)
4. Update entry status to ACTIVE or INACTIVE
5. Store Bacen Entry ID
6. ACK message (or NACK on failure for redelivery)

**Why Pulsar instead of Temporal?**: This is a fast operation (<1.5s) - no need for durable workflow overhead.

---

#### 2. dict.entries.updated

Triggered when core-dict updates an existing PIX key entry (account info change).

**Event Schema**:
```json
{
  "entry_id": "uuid",
  "key": "email@example.com",
  "key_type": "EMAIL",
  "participant": "12345678",
  "account_branch": "0002",        // Changed
  "account_number": "654321",      // Changed
  "account_type": "SLRY",          // Changed
  "account_opened_date": "2020-01-01",
  "owner_type": "NATURAL_PERSON",
  "owner_name": "John Doe Updated",
  "owner_tax_id": "12345678900",
  "idempotency_key": "unique-key",
  "request_id": "request-uuid",
  "timestamp": "2025-10-27T11:00:00Z"
}
```

**Consumer Behavior**:
1. Receive event from Pulsar
2. Fetch entry from database
3. Call Bridge gRPC `UpdateEntry` (synchronous, <1s)
4. Update entry in database
5. ACK message

---

#### 3. dict.entries.deleted.immediate

Triggered when core-dict deletes a PIX key entry **immediately** (no 30-day waiting period).

**Event Schema**:
```json
{
  "entry_id": "uuid",
  "key": "email@example.com",
  "key_type": "EMAIL",
  "reason": "Customer request",
  "idempotency_key": "unique-key",
  "request_id": "request-uuid",
  "timestamp": "2025-10-27T12:00:00Z"
}
```

**Consumer Behavior**:
1. Receive event from Pulsar
2. Fetch entry from database
3. Call Bridge gRPC `DeleteEntry` (synchronous, <1s)
4. Soft delete entry in database (deleted_at timestamp)
5. ACK message

**Note**: For deletions with 30-day waiting period, use `DeleteEntryWithWaitingPeriodWorkflow` (Temporal) instead.

---

### Output Topics (conn-dict → core-dict)

These topics are **published** by conn-dict to notify core-dict of state changes.

#### 1. dict.entries.status.changed

Published when an entry's status changes (e.g., PENDING → ACTIVE, ACTIVE → INACTIVE).

**Event Schema**:
```json
{
  "id": "event-uuid",
  "type": "EntryStatusChanged",
  "timestamp": "2025-10-27T10:00:02Z",
  "version": "1.0",
  "entry_id": "uuid",
  "old_status": "PENDING",
  "new_status": "ACTIVE",
  "bacen_entry_id": "bacen-id",
  "reason": "Successfully registered with Bacen"
}
```

**When Published**:
- After Bridge CreateEntry success → PENDING → ACTIVE
- After Bridge call failure → PENDING → INACTIVE
- After validation failure → ACTIVE → INACTIVE

---

#### 2. dict.claims.created

Published when a new claim is created via CreateClaim RPC.

**Event Schema**:
```json
{
  "id": "event-uuid",
  "type": "claim.created",
  "timestamp": "2025-10-27T10:00:00Z",
  "version": "1.0",
  "aggregate": "claim-uuid",
  "claim_id": "claim-uuid",
  "key": "+5511987654321",
  "key_type": "PHONE",
  "ispb": "87654321",
  "branch": "0001",
  "account": "123456",
  "account_type": "CACC",
  "owner_name": "Jane Doe",
  "owner_document": "98765432100"
}
```

**When Published**:
- Immediately after ClaimWorkflow starts

---

#### 3. dict.claims.completed

Published when a claim is confirmed, cancelled, or expired.

**Event Schema**:
```json
{
  "id": "event-uuid",
  "type": "claim.completed",
  "timestamp": "2025-10-27T12:00:00Z",
  "version": "1.0",
  "aggregate": "claim-uuid",
  "claim_id": "claim-uuid",
  "status": "CONFIRMED",  // or CANCELLED, EXPIRED
  "confirmed_by": "user@bank.com",
  "confirmed_at": "2025-10-27T12:00:00Z"
}
```

**When Published**:
- After claim is confirmed (ConfirmClaim RPC + workflow completion)
- After claim is cancelled (CancelClaim RPC)
- After 30 days elapse (auto-expiry)

---

#### 4. dict.infractions.created

Published when a new infraction is reported via CreateInfraction RPC.

**Event Schema**:
```json
{
  "id": "event-uuid",
  "type": "infraction.created",
  "timestamp": "2025-10-27T10:00:00Z",
  "version": "1.0",
  "aggregate": "infraction-uuid",
  "infraction_id": "infraction-uuid",
  "key": "+5511987654321",
  "type": "FRAUD",
  "reporter_ispb": "12345678",
  "reported_ispb": "87654321"
}
```

---

## Temporal Workflows

conn-dict uses Temporal for durable long-running workflows (>30 days).

### 1. ClaimWorkflow

**Purpose**: Manage 30-day claim lifecycle (OPEN → CONFIRMED/CANCELLED/EXPIRED).

**Workflow ID**: `claim-workflow-{claim_id}`

**Input**:
```go
type ClaimWorkflowInput struct {
    ClaimID        string
    EntryID        string
    ClaimType      string  // OWNERSHIP or PORTABILITY
    ClaimerISPB    string
    DonorISPB      string
    ClaimerAccount string
    RequestedBy    string
}
```

**Duration**: 30 days (auto-expire if no response)

**Signals Accepted**:
- `confirm`: Donor confirms claim → transition to CONFIRMED
- `cancel`: Claimer/donor cancels claim → transition to CANCELLED

**Activities**:
- CreateClaimActivity: Persist claim to database
- NotifyClaimDonorActivity: Send notification to donor bank
- ConfirmClaimActivity: Transfer key ownership
- CancelClaimActivity: Revert claim
- ExpireClaimActivity: Auto-expire after 30 days

**State Transitions**:
```
OPEN → (7 days, no response) → AUTO_APPROVED → CONFIRMED
OPEN → (confirm signal) → CONFIRMED
OPEN → (cancel signal) → CANCELLED
OPEN → (30 days) → EXPIRED
```

**How to Start** (via gRPC RPC):
```bash
# Call CreateClaim RPC
grpcurl -plaintext -d '{
  "entry_id": "uuid",
  "claim_type": "OWNERSHIP",
  "claimer_ispb": "87654321",
  "donor_ispb": "12345678"
}' localhost:9092 conndict.v1.ClaimService/CreateClaim
```

---

### 2. DeleteEntryWithWaitingPeriodWorkflow

**Purpose**: Soft delete entry after 30-day waiting period (Bacen regulation).

**Workflow ID**: `delete-entry-waiting-{entry_id}`

**Input**:
```go
type DeleteEntryWithWaitingPeriodInput struct {
    EntryID string
    Reason  string
}
```

**Duration**: 30 days (timer)

**Activities**:
- MarkEntryForDeletionActivity: Set entry status to MARKED_FOR_DELETION
- WaitThirtyDaysActivity: Sleep for 30 days (Temporal timer)
- DeleteEntryActivity: Soft delete after waiting period

**When to Use**:
- User requests account closure → 30-day waiting period required by Bacen
- Fraudulent account → immediate delete (use dict.entries.deleted.immediate Pulsar topic instead)

---

### 3. VSyncWorkflow

**Purpose**: Daily batch synchronization with Bacen RSFN to detect drift.

**Workflow ID**: `vsync-workflow-{date}`

**Schedule**: Daily at 2:00 AM BRT (Cron: `0 2 * * *`)

**Input**:
```go
type VSyncInput struct {
    SyncDate       time.Time
    BatchSize      int     // Default: 1000
    ParticipantISPB string // Optional (sync all if empty)
}
```

**Activities**:
- FetchBacenEntriesActivity: Fetch all entries from Bacen via Bridge
- FetchLocalEntriesActivity: Fetch all entries from local database
- CompareAndReportDriftActivity: Detect missing/extra entries
- ReconcileEntriesActivity: Auto-fix discrepancies

**Output**:
- SyncReport stored in database (sync_reports table)
- Metrics published to Prometheus

**How it Works**:
1. Fetch all entries from Bacen (paginated, batch of 1000)
2. Fetch all entries from local database
3. Compare: Identify entries missing in Bacen or extra in local DB
4. Auto-reconcile: Re-create missing entries, soft-delete extra entries
5. Store sync report with drift count

---

### 4. InvestigateInfractionWorkflow

**Purpose**: Human-in-the-loop workflow for infraction investigation.

**Workflow ID**: `infraction-{infraction_id}`

**Input**:
```go
type InvestigateInfractionInput struct {
    InfractionID   string
    Key            string
    Type           string
    Description    string
    ReporterISPB   string
    ReportedISPB   string
    EvidenceURLs   []string
    RelatedEntryID string
    RelatedClaimID string
}
```

**Duration**: Variable (waits for human decision, max 30 days)

**Signals Accepted**:
- `investigation_complete`: Human submits investigation decision (RESOLVE, DISMISS, ESCALATE)

**Activities**:
- CreateInfractionActivity: Persist infraction to database
- NotifyReportedParticipantActivity: Notify reported bank
- ResolveInfractionActivity: Mark as resolved
- DismissInfractionActivity: Mark as dismissed
- EscalateInfractionActivity: Escalate to Bacen

**State Transitions**:
```
OPEN → (investigation_complete signal: RESOLVE) → RESOLVED
OPEN → (investigation_complete signal: DISMISS) → DISMISSED
OPEN → (investigation_complete signal: ESCALATE) → ESCALATED
OPEN → (30 days, no decision) → AUTO_DISMISSED
```

---

## Usage Examples

### Example 1: Query Entry by Key (gRPC)

**Using grpcurl**:
```bash
grpcurl -plaintext -d '{
  "key": "email@example.com"
}' localhost:9092 conndict.v1.EntryService/GetEntry
```

**Using Go gRPC client**:
```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    pb "github.com/lbpay-lab/dict-contracts/gen/proto/conndict/v1"
)

func main() {
    conn, err := grpc.Dial("localhost:9092", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewEntryServiceClient(conn)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    resp, err := client.GetEntry(ctx, &pb.GetEntryRequest{
        Key: "email@example.com",
    })
    if err != nil {
        log.Fatalf("GetEntry failed: %v", err)
    }

    fmt.Printf("Entry ID: %s\n", resp.EntryId)
    fmt.Printf("Status: %s\n", resp.Status)
    fmt.Printf("Participant: %s\n", resp.Participant)
}
```

---

### Example 2: Create Claim (gRPC)

**Using grpcurl**:
```bash
grpcurl -plaintext -d '{
  "entry_id": "entry-uuid",
  "claim_type": "PORTABILITY",
  "claimer_ispb": "87654321",
  "donor_ispb": "12345678",
  "claimer_account": "0001-123456",
  "requested_by": "user@bank.com"
}' localhost:9092 conndict.v1.ClaimService/CreateClaim
```

**Using Go gRPC client**:
```go
resp, err := client.CreateClaim(ctx, &pb.CreateClaimRequest{
    EntryId:        "entry-uuid",
    ClaimType:      pb.ClaimType_CLAIM_TYPE_PORTABILITY,
    ClaimerIspb:    "87654321",
    DonorIspb:      "12345678",
    ClaimerAccount: "0001-123456",
    RequestedBy:    "user@bank.com",
})
if err != nil {
    log.Fatalf("CreateClaim failed: %v", err)
}

fmt.Printf("Claim ID: %s\n", resp.ClaimId)
fmt.Printf("Workflow ID: %s\n", resp.WorkflowId)
fmt.Printf("Expires At: %s\n", resp.ExpiresAt)
```

---

### Example 3: Publish Entry Created Event (Pulsar)

**Using Go Pulsar client**:
```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "time"

    "github.com/apache/pulsar-client-go/pulsar"
)

type EntryCreatedEvent struct {
    EntryID           string    `json:"entry_id"`
    Key               string    `json:"key"`
    KeyType           string    `json:"key_type"`
    Participant       string    `json:"participant"`
    AccountBranch     *string   `json:"account_branch,omitempty"`
    AccountNumber     *string   `json:"account_number,omitempty"`
    AccountType       string    `json:"account_type"`
    OwnerType         string    `json:"owner_type"`
    OwnerName         *string   `json:"owner_name,omitempty"`
    OwnerTaxID        *string   `json:"owner_tax_id,omitempty"`
    IdempotencyKey    string    `json:"idempotency_key"`
    RequestID         string    `json:"request_id"`
    Timestamp         time.Time `json:"timestamp"`
}

func main() {
    client, err := pulsar.NewClient(pulsar.ClientOptions{
        URL: "pulsar://localhost:6650",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    producer, err := client.CreateProducer(pulsar.ProducerOptions{
        Topic: "dict.entries.created",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer producer.Close()

    branch := "0001"
    account := "123456"
    ownerName := "John Doe"
    ownerTaxID := "12345678900"

    event := EntryCreatedEvent{
        EntryID:        "entry-uuid",
        Key:            "email@example.com",
        KeyType:        "EMAIL",
        Participant:    "12345678",
        AccountBranch:  &branch,
        AccountNumber:  &account,
        AccountType:    "CACC",
        OwnerType:      "NATURAL_PERSON",
        OwnerName:      &ownerName,
        OwnerTaxID:     &ownerTaxID,
        IdempotencyKey: "unique-key",
        RequestID:      "request-uuid",
        Timestamp:      time.Now(),
    }

    payload, _ := json.Marshal(event)

    _, err = producer.Send(context.Background(), &pulsar.ProducerMessage{
        Payload: payload,
        Key:     event.EntryID,
    })
    if err != nil {
        log.Fatalf("Failed to publish: %v", err)
    }

    log.Println("Event published successfully")
}
```

---

### Example 4: Start Temporal Workflow (Go SDK)

**Start ClaimWorkflow from code**:
```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "go.temporal.io/sdk/client"
)

type ClaimWorkflowInput struct {
    ClaimID        string
    EntryID        string
    ClaimType      string
    ClaimerISPB    string
    DonorISPB      string
    ClaimerAccount string
    RequestedBy    string
}

func main() {
    c, err := client.Dial(client.Options{
        HostPort: "localhost:7233",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer c.Close()

    input := ClaimWorkflowInput{
        ClaimID:        "claim-uuid",
        EntryID:        "entry-uuid",
        ClaimType:      "OWNERSHIP",
        ClaimerISPB:    "87654321",
        DonorISPB:      "12345678",
        ClaimerAccount: "0001-123456",
        RequestedBy:    "user@bank.com",
    }

    workflowOptions := client.StartWorkflowOptions{
        ID:        fmt.Sprintf("claim-workflow-%s", input.ClaimID),
        TaskQueue: "dict-claims-queue",
        WorkflowExecutionTimeout: 31 * 24 * time.Hour,
    }

    we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, "ClaimWorkflow", input)
    if err != nil {
        log.Fatalf("Failed to start workflow: %v", err)
    }

    fmt.Printf("Workflow started: %s (Run ID: %s)\n", we.GetID(), we.GetRunID())
}
```

---

## Health & Metrics

### Health Check Endpoints

#### GET /health (Liveness)

Check if service is alive.

**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2025-10-27T10:00:00Z"
}
```

**Status Codes**:
- `200 OK`: Service is alive
- `503 Service Unavailable`: Service is down

**Usage**:
```bash
curl http://localhost:8080/health
```

---

#### GET /ready (Readiness)

Check if service is ready to handle requests (database + Temporal + Pulsar connections healthy).

**Response**:
```json
{
  "status": "ready",
  "timestamp": "2025-10-27T10:00:00Z",
  "checks": {
    "database": "ok",
    "temporal": "ok",
    "pulsar": "ok",
    "bridge_grpc": "ok"
  }
}
```

**Status Codes**:
- `200 OK`: Service is ready
- `503 Service Unavailable`: One or more dependencies unavailable

**Usage**:
```bash
curl http://localhost:8080/ready
```

---

### Prometheus Metrics

#### GET /metrics

Expose Prometheus metrics for monitoring.

**Metrics Exposed**:
- `grpc_requests_total`: Total gRPC requests (label: method, status)
- `grpc_request_duration_seconds`: gRPC request latency histogram
- `pulsar_messages_consumed_total`: Total Pulsar messages consumed (label: topic)
- `pulsar_messages_published_total`: Total Pulsar messages published (label: topic)
- `temporal_workflow_started_total`: Total Temporal workflows started (label: workflow_type)
- `temporal_workflow_completed_total`: Total Temporal workflows completed (label: workflow_type, status)
- `database_query_duration_seconds`: Database query latency histogram
- `entry_status_total`: Total entries by status (label: status)
- `claim_status_total`: Total claims by status (label: status)
- `infraction_status_total`: Total infractions by status (label: status)

**Usage**:
```bash
curl http://localhost:9091/metrics
```

**Prometheus Scrape Config**:
```yaml
scrape_configs:
  - job_name: 'conn-dict'
    static_configs:
      - targets: ['localhost:9091']
```

---

## Error Handling

### gRPC Status Codes

conn-dict follows standard gRPC error conventions:

| gRPC Code | HTTP Equivalent | When Used |
|-----------|-----------------|-----------|
| `OK` | 200 | Success |
| `INVALID_ARGUMENT` | 400 | Missing/invalid required fields |
| `NOT_FOUND` | 404 | Entry/claim/infraction not found |
| `ALREADY_EXISTS` | 409 | Duplicate entry/claim |
| `FAILED_PRECONDITION` | 412 | Operation not allowed in current state |
| `UNIMPLEMENTED` | 501 | RPC not yet implemented |
| `INTERNAL` | 500 | Database/Temporal/Pulsar errors |
| `UNAVAILABLE` | 503 | Service temporarily unavailable |
| `DEADLINE_EXCEEDED` | 504 | Request timeout |

---

### Retry Policies

**gRPC Client Recommendations**:
```go
retryPolicy := `{
    "methodConfig": [{
        "name": [{"service": "conndict.v1.EntryService"}],
        "retryPolicy": {
            "maxAttempts": 3,
            "initialBackoff": "0.1s",
            "maxBackoff": "1s",
            "backoffMultiplier": 2,
            "retryableStatusCodes": ["UNAVAILABLE", "DEADLINE_EXCEEDED"]
        }
    }]
}`

conn, err := grpc.Dial(
    "localhost:9092",
    grpc.WithDefaultServiceConfig(retryPolicy),
)
```

**Pulsar Consumer Retry**:
- ACK timeout: 20s
- NACK redelivery delay: 60s
- Max delivery attempts: 5
- Dead letter queue: `dict.entries.created-dlq`

**Temporal Retry**:
- Activities auto-retry with exponential backoff
- Max attempts: 5
- Initial interval: 1s
- Max interval: 60s

---

### Timeout Recommendations

| Operation | Recommended Timeout |
|-----------|---------------------|
| GetEntry, GetClaim, GetInfraction | 5s |
| ListEntries, ListClaims, ListInfractions | 10s |
| CreateClaim, CreateInfraction | 30s (workflow start overhead) |
| ConfirmClaim, CancelClaim | 10s (signal send) |
| Pulsar publish | 30s |
| Bridge gRPC calls | 10s |

---

## Security & Authentication

### mTLS (Mutual TLS)

**Production**: All gRPC connections MUST use mTLS with ICP-Brasil A3 certificates.

**Configuration**:
```bash
# Server TLS
TLS_CERT_FILE=/path/to/server-cert.pem
TLS_KEY_FILE=/path/to/server-key.pem
TLS_CA_FILE=/path/to/ca-cert.pem

# Client TLS (for Bridge calls)
BRIDGE_TLS_CERT_FILE=/path/to/client-cert.pem
BRIDGE_TLS_KEY_FILE=/path/to/client-key.pem
BRIDGE_TLS_CA_FILE=/path/to/ca-cert.pem
```

**Go Client Example with mTLS**:
```go
creds, err := credentials.NewClientTLSFromFile("ca-cert.pem", "")
if err != nil {
    log.Fatal(err)
}

conn, err := grpc.Dial(
    "localhost:9092",
    grpc.WithTransportCredentials(creds),
)
```

---

### Authentication

**JWT Token** (passed in gRPC metadata):
```go
ctx := metadata.AppendToOutgoingContext(
    context.Background(),
    "authorization", "Bearer <jwt-token>",
)

resp, err := client.GetEntry(ctx, req)
```

**Token Claims**:
- `ispb`: Bank ISPB (8 digits)
- `role`: User role (ADMIN, OPERATOR, VIEWER)
- `exp`: Expiration timestamp

**Authorization**:
- Users can ONLY query entries/claims/infractions for their own ISPB
- ADMIN role required for ListInfractions without filters

---

## Appendix: Quick Reference

### All gRPC RPCs (16 Total)

**EntryService** (3 RPCs):
1. GetEntry
2. GetEntryByKey
3. ListEntries

**ClaimService** (5 RPCs):
4. CreateClaim
5. ConfirmClaim
6. CancelClaim
7. GetClaim
8. ListClaims

**InfractionService** (6 RPCs):
9. CreateInfraction
10. InvestigateInfraction
11. ResolveInfraction
12. DismissInfraction
13. GetInfraction
14. ListInfractions

---

### All Pulsar Topics (6 Total)

**Input** (core-dict → conn-dict):
1. dict.entries.created
2. dict.entries.updated
3. dict.entries.deleted.immediate

**Output** (conn-dict → core-dict):
4. dict.entries.status.changed
5. dict.claims.created
6. dict.claims.completed

---

### All Temporal Workflows (4 Total)

1. ClaimWorkflow (30 days)
2. DeleteEntryWithWaitingPeriodWorkflow (30 days)
3. VSyncWorkflow (cron daily)
4. InvestigateInfractionWorkflow (human-in-the-loop)

---

## Contact & Support

**Maintainer**: conn-dict Team
**Slack Channel**: #dict-support
**Issue Tracker**: GitHub Issues

**Documentation Version**: 1.0
**Last Updated**: 2025-10-27
**Status**: PRODUCTION-READY
