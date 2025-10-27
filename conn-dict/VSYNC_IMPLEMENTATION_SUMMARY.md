# VSYNC Workflow Implementation Summary

**Date**: 2025-10-27
**Repository**: conn-dict
**Component**: RSFN Connect - Data Synchronization
**Status**: ✅ Workflow Created, Activities Stubbed

---

## 📁 Files Created

| File | Lines | Purpose |
|------|-------|---------|
| `internal/workflows/vsync_workflow.go` | 332 | Main VSYNC workflow + scheduler |
| `internal/workflows/vsync_workflow_test.go` | 312 | Comprehensive unit tests |
| `internal/activities/vsync_activities.go` | 268 | Activity stubs + types |
| `internal/workflows/VSYNC_README.md` | 513 | Complete documentation |
| **Total** | **1,425** | **4 files** |

---

## ✅ What Was Implemented

### 1. VSyncWorkflow (vsync_workflow.go)

**Core Workflow**: Periodic data synchronization with Bacen DICT

**Features**:
- ✅ Five-step workflow execution
- ✅ Three discrepancy types (MISSING_LOCAL, OUTDATED_LOCAL, MISSING_BACEN)
- ✅ Comprehensive error handling (COMPLETED, PARTIAL, FAILED statuses)
- ✅ Activity timeout configuration using existing ActivityOptions
- ✅ Audit trail via GenerateSyncReportActivity
- ✅ Event publishing to Pulsar

**Input**:
```go
type VSyncInput struct {
    ParticipantISPB string     // ISPB to sync (empty = all)
    SyncType        string     // "FULL" or "INCREMENTAL"
    LastSyncDate    *time.Time // For incremental sync
}
```

**Output**:
```go
type VSyncResult struct {
    EntriesSynced    int
    EntriesCreated   int
    EntriesUpdated   int
    EntriesDeleted   int
    Discrepancies    int
    Duration         time.Duration
    SyncTimestamp    time.Time
    ReportID         string
    Status           string // "COMPLETED", "PARTIAL", "FAILED"
    ErrorMessage     string
}
```

---

### 2. VSyncSchedulerWorkflow (vsync_workflow.go)

**Purpose**: Cron scheduler that runs VSYNC every 24 hours

**Features**:
- ✅ ContinueAsNew pattern to prevent history bloat
- ✅ Child workflow execution
- ✅ Graceful error handling (continues even if sync fails)
- ✅ Configurable sleep duration (default: 24 hours)

**Cron Pattern**:
```bash
# Run at 2 AM daily
--cron "0 2 * * *"
```

**Why ContinueAsNew?**
- Prevents unbounded workflow history growth
- Restarts workflow with fresh history after each iteration
- Essential for long-running cron workflows

---

### 3. VSYNC Activities (vsync_activities.go)

**Activity Stubs Created**:

#### FetchBacenEntriesActivity
- **Purpose**: Fetch entries from Bacen DICT API
- **Type**: External API call
- **Timeout**: 30s (ExternalAPI options)
- **Retry**: 10 attempts
- **TODO**: Implement Bacen API client (use conn-bridge gRPC)

#### CompareEntriesActivity
- **Purpose**: Compare Bacen data vs local database
- **Type**: Database query
- **Timeout**: 10s (Database options)
- **Retry**: 5 attempts
- **TODO**: Implement hash map comparison logic

#### GenerateSyncReportActivity
- **Purpose**: Create audit report for compliance
- **Type**: Database insert
- **Timeout**: 10s (Database options)
- **Retry**: 5 attempts
- **TODO**: Create `sync_reports` table and repository

**Data Types**:
```go
type BacenEntry struct {
    Key             string
    KeyType         string
    ParticipantISPB string
    AccountBranch   string
    AccountNumber   string
    AccountType     string
    OwnerType       string
    OwnerName       string
    OwnerTaxID      string
    Status          string
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type EntryDiscrepancy struct {
    Type        string // MISSING_LOCAL, OUTDATED_LOCAL, MISSING_BACEN
    Key         string
    EntryID     string
    Reason      string
    CreateInput CreateEntryInput
    UpdateInput UpdateEntryInput
    BacenData   *BacenEntry
    LocalData   interface{}
}
```

---

### 4. Unit Tests (vsync_workflow_test.go)

**Test Coverage**: 6 comprehensive test scenarios

| Test | Scenario | Assertions |
|------|----------|------------|
| `TestVSyncWorkflow_Success` | Normal sync with 1 discrepancy | Status=COMPLETED, 1 created |
| `TestVSyncWorkflow_NoDiscrepancies` | Database already in sync | Status=COMPLETED, 0 discrepancies |
| `TestVSyncWorkflow_PartialFailure` | Some activities fail | Status=PARTIAL, error message |
| `TestVSyncWorkflow_InvalidInput` | Invalid sync type | Workflow error |
| `TestVSyncWorkflow_IncrementalWithoutDate` | Missing LastSyncDate | Validation error |
| `TestVSyncWorkflow_MultipleDiscrepancyTypes` | All 3 discrepancy types | Correct handling |

**Run Tests**:
```bash
cd /Users/jose.silva.lb/LBPay/IA_Dict/conn-dict
go test -v ./internal/workflows -run TestVSyncWorkflow
```

---

### 5. Documentation (VSYNC_README.md)

**Comprehensive guide covering**:
- ✅ Architecture and workflow steps
- ✅ Discrepancy types and handling
- ✅ Cron scheduling pattern
- ✅ Error handling strategy
- ✅ Compliance considerations (LGPD, audit trail)
- ✅ Performance optimization (batching, indexes)
- ✅ Monitoring and alerting
- ✅ Troubleshooting guide
- ✅ Implementation checklist

---

## 🎯 Activity Dependencies

### Activities Reused from Existing Code

These activities already exist in `conn-dict`:

- ✅ `CreateEntryActivity` (entry_activities.go)
- ✅ `UpdateEntryActivity` (entry_activities.go)
- ✅ `PublishClaimEventActivity` (claim_activities.go)

### Activities Requiring Implementation

These activities are stubbed and need implementation:

- ⚠️ `FetchBacenEntriesActivity` (vsync_activities.go)
- ⚠️ `CompareEntriesActivity` (vsync_activities.go)
- ⚠️ `GenerateSyncReportActivity` (vsync_activities.go)

---

## 🔧 Error Handling Strategy

### Workflow-Level

| Scenario | Status | Action |
|----------|--------|--------|
| All fixes succeed | `COMPLETED` | Normal completion |
| Some fixes fail | `PARTIAL` | Log errors, continue workflow |
| Bacen API fails | `FAILED` | Return error, retry on next schedule |
| Database unreachable | `FAILED` | Return error, retry on next schedule |

### Activity-Level

| Activity | Error | Strategy |
|----------|-------|----------|
| FetchBacenEntriesActivity | Timeout | Retry 10x with backoff |
| FetchBacenEntriesActivity | 4xx (auth) | Fail immediately |
| CompareEntriesActivity | DB error | Retry 5x |
| CreateEntryActivity | Validation error | Skip entry, log warning |
| UpdateEntryActivity | DB error | Retry 5x |
| GenerateSyncReportActivity | Fail | Log warning (non-critical) |
| PublishClaimEventActivity | Fail | Log warning (non-critical) |

### Discrepancy-Specific

**MISSING_BACEN (Critical Decision)**:
- **DO NOT** auto-delete entries
- **Reason**: Entry could be pending registration, network error, or Bacen data issue
- **Action**: Flag for manual review by compliance team

---

## 📊 Compliance Considerations

### Audit Trail

Every VSYNC execution generates a detailed audit report:

**Database Schema** (TODO):
```sql
CREATE TABLE sync_reports (
    id UUID PRIMARY KEY,
    sync_timestamp TIMESTAMPTZ NOT NULL,
    entries_synced INT NOT NULL,
    entries_created INT NOT NULL,
    entries_updated INT NOT NULL,
    entries_deleted INT NOT NULL,
    discrepancies INT NOT NULL,
    status VARCHAR(20) NOT NULL,
    duration_ms BIGINT NOT NULL,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Retention: 7 years (Bacen requirement)
CREATE INDEX idx_sync_reports_timestamp ON sync_reports(sync_timestamp DESC);
```

### Manual Review Queue

Entries flagged as `MISSING_BACEN` require manual review:

**Database Schema** (TODO):
```sql
CREATE TABLE manual_review_queue (
    id UUID PRIMARY KEY,
    entry_id UUID NOT NULL,
    key VARCHAR(77) NOT NULL,
    reason TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING',
    reviewed_by UUID,
    reviewed_at TIMESTAMPTZ,
    action_taken TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### LGPD Compliance

- ✅ Log only `entry_id` and `key` (PIX key is pseudonymized)
- ✅ Avoid logging PII in standard logs
- ✅ Store detailed data in encrypted `sync_reports` table with RBAC

---

## 🚀 Next Steps (Implementation Checklist)

### Phase 1: Activity Implementation (High Priority)

- [ ] **FetchBacenEntriesActivity**
  - [ ] Create Bacen DICT API client (use conn-bridge gRPC)
  - [ ] Implement pagination handling (Bacen max 1000 entries/page)
  - [ ] Add mTLS authentication
  - [ ] Handle XML/JSON response parsing

- [ ] **CompareEntriesActivity**
  - [ ] Query local database using EntryRepository
  - [ ] Implement hash map comparison (O(n) efficiency)
  - [ ] Detect all 3 discrepancy types
  - [ ] Generate EntryDiscrepancy objects

- [ ] **GenerateSyncReportActivity**
  - [ ] Create SyncReportRepository
  - [ ] Create `sync_reports` database table
  - [ ] Insert report with all statistics
  - [ ] Optionally generate PDF/JSON export

### Phase 2: Database Schema (High Priority)

- [ ] Create migration for `sync_reports` table
- [ ] Create migration for `manual_review_queue` table
- [ ] Add indexes for performance
- [ ] Configure retention policy (7 years)

### Phase 3: Workflow Registration (Medium Priority)

- [ ] Update `cmd/worker/main.go`
- [ ] Register `VSyncWorkflow` with Temporal worker
- [ ] Register `VSyncSchedulerWorkflow` with Temporal worker
- [ ] Register new VSYNC activities
- [ ] Test registration

### Phase 4: Deployment (Medium Priority)

- [ ] Deploy worker with VSYNC workflows
- [ ] Start `VSyncSchedulerWorkflow` cron job
  ```bash
  temporal workflow start \
    --task-queue dict-task-queue \
    --type VSyncSchedulerWorkflow \
    --workflow-id vsync-scheduler \
    --cron "0 2 * * *"
  ```
- [ ] Verify first execution
- [ ] Monitor workflow metrics

### Phase 5: Monitoring (Low Priority)

- [ ] Add Prometheus metrics
  - [ ] `vsync_duration_seconds`
  - [ ] `vsync_discrepancies_total`
  - [ ] `vsync_status`
- [ ] Configure Grafana dashboards
- [ ] Set up AlertManager rules
- [ ] Create runbook for compliance team

---

## 📈 Performance Considerations

### Batching (for large datasets)

```go
const BatchSize = 1000

for i := 0; i < len(discrepancies); i += BatchSize {
    batch := discrepancies[i:min(i+BatchSize, len(discrepancies))]
    // Process batch
    workflow.RecordHeartbeat(ctx, i)
}
```

### Database Indexes (TODO)

```sql
-- Optimize local entry lookup
CREATE INDEX idx_entries_key ON entries(key);
CREATE INDEX idx_entries_ispb ON entries(participant_ispb);
CREATE INDEX idx_entries_updated_at ON entries(updated_at DESC);

-- Optimize comparison query
CREATE INDEX idx_entries_ispb_key ON entries(participant_ispb, key);
```

### Parallel Execution

For multiple ISPBs, run workflows in parallel:
```bash
temporal workflow start --type VSyncWorkflow --input '{"participant_ispb": "12345678", ...}'
temporal workflow start --type VSyncWorkflow --input '{"participant_ispb": "87654321", ...}'
```

---

## 📝 Testing Strategy

### Unit Tests (Completed)

- ✅ 6 test scenarios covering happy path, errors, edge cases
- ✅ Mock all activities using Temporal test suite
- ✅ Assert workflow results and status

**Run**:
```bash
go test -v ./internal/workflows -run TestVSyncWorkflow
```

### Integration Tests (TODO)

- [ ] Test against real Temporal server
- [ ] Test with stub Bacen API
- [ ] Test database transactions
- [ ] Test event publishing to Pulsar

### End-to-End Tests (TODO)

- [ ] Run full VSYNC against test environment
- [ ] Verify report generation
- [ ] Verify manual review queue population
- [ ] Verify metrics and alerts

---

## 🔍 Monitoring and Alerts

### Key Metrics

| Metric | Threshold | Alert Level |
|--------|-----------|-------------|
| `vsync_duration_seconds` | > 3600 (1 hour) | Warning |
| `vsync_discrepancies_total` | > 100 | Warning |
| `vsync_error_rate` | > 5% | Warning |
| `vsync_failed_count` | > 0 | Critical |
| `vsync_missing_bacen_count` | > 10 | Warning |

### Alerting Rules (TODO)

```yaml
# Prometheus AlertManager
groups:
  - name: vsync
    rules:
      - alert: VSyncFailed
        expr: vsync_status{status="failed"} > 0
        severity: critical

      - alert: VSyncHighDiscrepancies
        expr: vsync_discrepancies_total > 100
        severity: warning
```

---

## 📚 References

- **Temporal Docs**: https://docs.temporal.io/workflows
- **ContinueAsNew Pattern**: https://docs.temporal.io/workflows#continue-as-new
- **Bacen DICT Spec**: https://www.bcb.gov.br/estabilidadefinanceira/pix
- **LGPD Guidelines**: https://www.gov.br/cidadania/pt-br/acesso-a-informacao/lgpd

---

## 📞 Support

**Team**: Backend Connect Team
**Slack**: #dict-backend
**On-Call**: PagerDuty rotation
**Documentation**: `/conn-dict/internal/workflows/VSYNC_README.md`

---

**Implementation Status**: 🟡 In Progress (Workflow complete, activities stubbed)
**Next Milestone**: Implement activity logic + database schema
**Target Completion**: Sprint 4

---

**Last Updated**: 2025-10-27
**Maintainer**: Backend Connect Team
**Version**: 1.0