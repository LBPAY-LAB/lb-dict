# VSYNC Workflow - DICT Data Synchronization

## Overview

The **VSYNC (Verification and Synchronization) Workflow** is a critical Temporal workflow that ensures data consistency between the local DICT database and Bacen's authoritative DICT registry.

### Purpose

- **Compliance**: Maintain regulatory compliance with Bacen DICT requirements
- **Data Integrity**: Detect and fix data inconsistencies automatically
- **Audit Trail**: Generate detailed reports for compliance and troubleshooting
- **Reconciliation**: Periodic reconciliation prevents data drift

### Execution Schedule

- **Frequency**: Every 24 hours (daily)
- **Recommended Time**: 2 AM (low traffic period)
- **Scheduler**: `VSyncSchedulerWorkflow` (cron-based)

---

## Architecture

### Workflow Structure

```
VSyncWorkflow
├── Step 1: FetchBacenEntriesActivity (External API)
│   └── Fetch all entries from Bacen DICT API
├── Step 2: CompareEntriesActivity (Database)
│   └── Compare Bacen data vs local database
├── Step 3: Apply Fixes (Database)
│   ├── CreateEntryActivity (for MISSING_LOCAL)
│   ├── UpdateEntryActivity (for OUTDATED_LOCAL)
│   └── Flag for review (for MISSING_BACEN)
├── Step 4: GenerateSyncReportActivity (Database)
│   └── Create audit report
└── Step 5: PublishClaimEventActivity (Messaging)
    └── Publish sync completion event to Pulsar
```

### Activity Dependencies

| Activity | Type | Timeout | Retry Policy | Dependencies |
|----------|------|---------|--------------|--------------|
| `FetchBacenEntriesActivity` | External API | 30s | 10 retries | Bacen DICT API Client, mTLS |
| `CompareEntriesActivity` | Database | 10s | 5 retries | EntryRepository |
| `CreateEntryActivity` | Database | 10s | 5 retries | EntryRepository, Pulsar |
| `UpdateEntryActivity` | Database | 10s | 5 retries | EntryRepository, Pulsar |
| `GenerateSyncReportActivity` | Database | 10s | 5 retries | SyncReportRepository |
| `PublishClaimEventActivity` | Messaging | 15s | 7 retries | PulsarProducer |

---

## Discrepancy Types

### 1. MISSING_LOCAL

**Description**: Entry exists in Bacen DICT but not in local database

**Action**: Create entry locally

**Example**:
```
Bacen:   Key: +5511999999999, ISPB: 12345678
Local:   (not found)
Fix:     CreateEntryActivity(+5511999999999)
```

### 2. OUTDATED_LOCAL

**Description**: Entry exists in both systems but data differs

**Action**: Update local entry with Bacen data

**Example**:
```
Bacen:   Key: +5511999999999, Status: ACTIVE, Owner: João Silva
Local:   Key: +5511999999999, Status: ACTIVE, Owner: João S.
Fix:     UpdateEntryActivity(+5511999999999, {OwnerName: "João Silva"})
```

### 3. MISSING_BACEN

**Description**: Entry exists locally but not in Bacen DICT

**Action**: **FLAG FOR MANUAL REVIEW** (do NOT auto-delete)

**Reason**: Entry could be:
- Pending registration (not yet synced to Bacen)
- Network error during Bacen fetch
- Bacen data issue
- Recently deleted at Bacen (need to verify before deleting)

**Example**:
```
Bacen:   (not found)
Local:   Key: +5511777777777, ISPB: 12345678
Fix:     Log warning, flag for compliance team review
```

---

## Workflow Input/Output

### VSyncInput

```go
type VSyncInput struct {
    ParticipantISPB string     // ISPB to sync (empty = all)
    SyncType        string     // "FULL" or "INCREMENTAL"
    LastSyncDate    *time.Time // For incremental sync
}
```

**Sync Types**:
- **FULL**: Sync all entries (slower, comprehensive)
- **INCREMENTAL**: Sync only entries updated since `LastSyncDate` (faster, daily use)

### VSyncResult

```go
type VSyncResult struct {
    EntriesSynced    int           // Total entries processed
    EntriesCreated   int           // Missing entries created
    EntriesUpdated   int           // Outdated entries updated
    EntriesDeleted   int           // Entries flagged for review
    Discrepancies    int           // Total discrepancies found
    Duration         time.Duration // Execution time
    SyncTimestamp    time.Time     // When sync started
    ReportID         string        // Audit report ID
    Status           string        // "COMPLETED", "PARTIAL", "FAILED"
    ErrorMessage     string        // Error details (if any)
}
```

**Status Values**:
- **COMPLETED**: All discrepancies fixed successfully
- **PARTIAL**: Some fixes failed (check logs and report)
- **FAILED**: Critical error prevented sync (Bacen API down, DB unavailable)

---

## Cron Scheduling

### VSyncSchedulerWorkflow

The scheduler runs VSYNC daily using the **ContinueAsNew** pattern to prevent workflow history bloat.

#### Pattern

```go
func VSyncSchedulerWorkflow(ctx workflow.Context) error {
    // 1. Execute VSYNC as child workflow
    result := ExecuteChildWorkflow(VSyncWorkflow, ...)

    // 2. Sleep 24 hours
    Sleep(24 * time.Hour)

    // 3. Restart (ContinueAsNew)
    return workflow.NewContinueAsNewError(ctx, VSyncSchedulerWorkflow)
}
```

#### Why ContinueAsNew?

Without `ContinueAsNew`, Temporal would keep ALL workflow events in history forever, causing:
- **Memory issues**: History grows unbounded (365 days = 365 executions)
- **Performance degradation**: Loading gigantic history slows down workflow
- **Cost**: Larger history = more storage costs

With `ContinueAsNew`, workflow restarts with fresh history after each iteration.

#### Starting the Scheduler

```bash
# Method 1: Temporal CLI (recommended)
temporal workflow start \
  --task-queue dict-task-queue \
  --type VSyncSchedulerWorkflow \
  --workflow-id vsync-scheduler \
  --cron "0 2 * * *"  # Run at 2 AM daily

# Method 2: Start manually (will run indefinitely)
temporal workflow start \
  --task-queue dict-task-queue \
  --type VSyncSchedulerWorkflow \
  --workflow-id vsync-scheduler
```

#### Stopping the Scheduler

```bash
# Gracefully cancel the scheduler workflow
temporal workflow cancel --workflow-id vsync-scheduler
```

---

## Error Handling

### Strategy

| Error Type | Strategy | Rationale |
|------------|----------|-----------|
| Bacen API timeout | Retry 10x with exponential backoff | Transient network issue |
| Bacen API 5xx | Retry 10x | Server error, may recover |
| Bacen API 4xx | Fail immediately | Client error (auth, invalid request) |
| Database error | Retry 5x | Connection pool, deadlock |
| Validation error | Skip entry, log warning | Don't fail entire sync |
| Event publish failure | Log warning, continue | Non-critical (can replay later) |

### Partial Failure Handling

VSYNC uses **best-effort** approach:
- If some fixes fail, workflow completes with `PARTIAL` status
- Failed fixes are logged with details
- Sync report includes error count and messages
- Next sync will retry failed fixes

### Critical Failures

If `FetchBacenEntriesActivity` or `CompareEntriesActivity` fail, workflow returns `FAILED` status.

---

## Compliance Considerations

### Audit Trail

Every VSYNC execution generates a detailed audit report:

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

### LGPD Compliance

- **Logs**: Log only `entry_id` and `key` (PIX key is pseudonymized)
- **Reports**: Avoid logging PII (names, CPF, addresses) in standard logs
- **Detailed data**: Store in encrypted `sync_reports` table with access control

### Manual Review Queue

Entries flagged as `MISSING_BACEN` are added to a manual review queue:

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

Compliance team reviews and takes action:
- **Confirm deletion**: Entry was legitimately deleted at Bacen → delete locally
- **Keep entry**: Entry is pending registration → keep locally
- **Investigate**: Data discrepancy → contact Bacen support

---

## Performance Optimization

### Batching

For large ISPB counts (>10,000 entries), use batching:

```go
const BatchSize = 1000

for i := 0; i < len(discrepancies); i += BatchSize {
    batch := discrepancies[i:min(i+BatchSize, len(discrepancies))]

    // Process batch
    for _, disc := range batch {
        // Apply fix
    }

    // Heartbeat every batch
    workflow.RecordHeartbeat(ctx, i)
}
```

### Database Indexes

```sql
-- Optimize local entry lookup
CREATE INDEX idx_entries_key ON entries(key);
CREATE INDEX idx_entries_ispb ON entries(participant_ispb);
CREATE INDEX idx_entries_updated_at ON entries(updated_at DESC);

-- Optimize comparison query
CREATE INDEX idx_entries_ispb_key ON entries(participant_ispb, key);
```

### Parallel Processing

For multiple ISPBs, run VSYNC workflows in parallel:

```bash
# Start VSYNC for ISPB 12345678
temporal workflow start \
  --type VSyncWorkflow \
  --workflow-id vsync-12345678 \
  --input '{"participant_ispb": "12345678", "sync_type": "FULL"}'

# Start VSYNC for ISPB 87654321
temporal workflow start \
  --type VSyncWorkflow \
  --workflow-id vsync-87654321 \
  --input '{"participant_ispb": "87654321", "sync_type": "FULL"}'
```

---

## Monitoring and Alerts

### Metrics to Track

| Metric | Threshold | Alert |
|--------|-----------|-------|
| `vsync_duration_seconds` | > 3600 (1 hour) | Slow sync - investigate |
| `vsync_discrepancies` | > 100 | High discrepancy rate |
| `vsync_error_rate` | > 5% | Partial failures |
| `vsync_failed_count` | > 0 | Critical - sync failed |
| `vsync_missing_bacen_count` | > 10 | Many entries missing in Bacen |

### Prometheus Metrics

```go
var (
    vsyncDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
        Name: "vsync_duration_seconds",
        Help: "VSYNC workflow execution duration",
    })

    vsyncDiscrepancies = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "vsync_discrepancies_total",
        Help: "Total discrepancies found in last sync",
    })

    vsyncStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
        Name: "vsync_status",
        Help: "VSYNC status (1=completed, 0.5=partial, 0=failed)",
    }, []string{"status"})
)
```

### Alerting Rules

```yaml
# Prometheus AlertManager rules
groups:
  - name: vsync
    rules:
      - alert: VSyncFailed
        expr: vsync_status{status="failed"} > 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "VSYNC workflow failed"

      - alert: VSyncHighDiscrepancies
        expr: vsync_discrepancies_total > 100
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High discrepancy count in VSYNC"
```

---

## Testing

### Unit Tests

Run workflow tests:

```bash
cd /Users/jose.silva.lb/LBPay/IA_Dict/conn-dict
go test -v ./internal/workflows -run TestVSyncWorkflow
```

### Integration Tests

Test against real Temporal server:

```bash
# Start Temporal dev server
temporal server start-dev

# Run workflow
temporal workflow start \
  --type VSyncWorkflow \
  --task-queue dict-task-queue \
  --input '{"participant_ispb": "12345678", "sync_type": "FULL"}'

# Check workflow status
temporal workflow describe --workflow-id <workflow-id>
```

### Test Scenarios

1. **Happy path**: No discrepancies, all in sync
2. **Missing local**: Bacen has 10 entries, local has 5 → create 5
3. **Outdated local**: 5 entries have stale data → update 5
4. **Missing Bacen**: 3 entries not in Bacen → flag for review
5. **Partial failure**: Some CreateEntry calls fail → PARTIAL status
6. **Critical failure**: Bacen API down → FAILED status

---

## Implementation Checklist

### Phase 1: Activity Stubs (Completed)

- [x] `FetchBacenEntriesActivity` stub created
- [x] `CompareEntriesActivity` stub created
- [x] `GenerateSyncReportActivity` stub created
- [x] Data types defined (`BacenEntry`, `EntryDiscrepancy`)

### Phase 2: Activity Implementation (TODO)

- [ ] Implement Bacen API client (use conn-bridge gRPC)
- [ ] Implement database comparison logic
- [ ] Implement sync report generation
- [ ] Create `sync_reports` and `manual_review_queue` tables

### Phase 3: Workflow Registration (TODO)

- [ ] Register workflows and activities in Temporal worker
- [ ] Update `cmd/worker/main.go`

### Phase 4: Deployment (TODO)

- [ ] Deploy worker with VSYNC workflows
- [ ] Start `VSyncSchedulerWorkflow` cron
- [ ] Configure monitoring and alerts
- [ ] Document runbooks for compliance team

---

## Troubleshooting

### Common Issues

**Issue**: VSYNC workflow stuck in "Running" state

**Solution**: Check worker logs, ensure activities are registered and worker is running

---

**Issue**: High discrepancy count

**Solution**:
1. Check Bacen API availability
2. Verify local database backups are not being restored
3. Review manual review queue for patterns

---

**Issue**: Slow VSYNC execution (>1 hour)

**Solution**:
1. Add database indexes
2. Enable batching for large ISPB counts
3. Increase worker concurrency

---

**Issue**: Entries missing in Bacen (MISSING_BACEN)

**Solution**:
1. Check if entries are pending registration (recently created)
2. Verify Bacen API response completeness
3. Contact Bacen support if persistent

---

## References

- [Temporal Workflows Documentation](https://docs.temporal.io/workflows)
- [Temporal ContinueAsNew Pattern](https://docs.temporal.io/workflows#continue-as-new)
- [Bacen DICT API Specification](https://www.bcb.gov.br/estabilidadefinanceira/pix)
- [LGPD Compliance Guidelines](https://www.gov.br/cidadania/pt-br/acesso-a-informacao/lgpd)

---

**Last Updated**: 2025-10-27
**Version**: 1.0
**Maintainer**: Backend Connect Team