# Technical Specifications - DICT CID/VSync Sync System

**Document Version**: 2.0.0
**Last Updated**: 2025-10-28
**Status**: ✅ ALIGNED WITH CONNECTOR-DICT PATTERNS
**Responsible**: Tech Lead

---

## 📋 Executive Summary

Este documento define as especificações técnicas para implementação do sistema de **Sincronização CID/VSync** seguindo **estritamente os padrões arquiteturais do Connector-Dict** (Clean Architecture + Event-Driven + Temporal Workflows).

**Decisões Arquiteturais Principais**:
1. ✅ **Implementar APENAS no Orchestration Worker** (não modificar Dict API)
2. ✅ **Consumir eventos Pulsar** de key.created/key.updated (já existem)
3. ✅ **Usar Temporal Cron Workflow** para verificação diária de VSync
4. ✅ **Integrar via Bridge gRPC** para chamadas ao DICT BACEN
5. ✅ **Notificar Core-Dict via Pulsar** (core-events topic)

---

## 🏗️ Database Schema (PostgreSQL)

### Migration 001: `dict_cids` Table

**File**: `apps/orchestration-worker/infrastructure/database/migrations/001_create_dict_cids.sql`

```sql
-- Migration: Create dict_cids table for storing Content Identifiers
-- Author: Sync Squad
-- Date: 2025-10-28

-- Enable extensions if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Main table: dict_cids
CREATE TABLE IF NOT EXISTS dict_cids (
    -- Primary Key
    id                          BIGSERIAL PRIMARY KEY,

    -- CID (Content Identifier) - SHA-256 hex (256 bits = 64 chars)
    cid                         VARCHAR(64) NOT NULL UNIQUE,

    -- Key Identification
    key_type                    VARCHAR(10) NOT NULL,           -- CPF|CNPJ|PHONE|EMAIL|EVP
    key_value                   VARCHAR(255) NOT NULL,          -- Normalized key value

    -- Account Information (from Entry event)
    ispb                        VARCHAR(8) NOT NULL,
    branch                      VARCHAR(10),
    account_number              VARCHAR(20) NOT NULL,
    account_type                VARCHAR(4) NOT NULL,            -- CACC|SVGS|TRAN
    account_opened_at           TIMESTAMP NOT NULL,

    -- Owner Information (from Entry event)
    owner_type                  VARCHAR(2) NOT NULL,            -- PF|PJ
    owner_tax_id                VARCHAR(14) NOT NULL,
    owner_name                  VARCHAR(255) NOT NULL,
    owner_trade_name            VARCHAR(255),

    -- Registration Metadata
    registered_at               TIMESTAMP NOT NULL,
    participant_registered_at   TIMESTAMP NOT NULL,
    request_id                  UUID NOT NULL,

    -- Algorithm Version
    algorithm_version           VARCHAR(10) NOT NULL DEFAULT '1.0',

    -- Audit Timestamps
    created_at                  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT dict_cids_key_unique UNIQUE (key_type, key_value)
);

-- Indexes for performance (based on connector-dict patterns)
CREATE INDEX IF NOT EXISTS idx_dict_cids_key_type ON dict_cids(key_type);
CREATE INDEX IF NOT EXISTS idx_dict_cids_key_value ON dict_cids(key_value);
CREATE INDEX IF NOT EXISTS idx_dict_cids_cid ON dict_cids(cid);
CREATE INDEX IF NOT EXISTS idx_dict_cids_owner_tax_id ON dict_cids(owner_tax_id);
CREATE INDEX IF NOT EXISTS idx_dict_cids_created_at ON dict_cids(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_dict_cids_ispb ON dict_cids(ispb);

-- Composite index for common queries
CREATE INDEX IF NOT EXISTS idx_dict_cids_key_type_created ON dict_cids(key_type, created_at DESC);

-- Trigger for updated_at (connector-dict pattern)
CREATE OR REPLACE FUNCTION update_dict_cids_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_dict_cids_updated_at
    BEFORE UPDATE ON dict_cids
    FOR EACH ROW
    EXECUTE FUNCTION update_dict_cids_updated_at();

-- Comments
COMMENT ON TABLE dict_cids IS 'Stores Content Identifiers (CIDs) for all PIX keys as per BACEN Chapter 9';
COMMENT ON COLUMN dict_cids.cid IS 'SHA-256 hash of normalized Entry data (64-char hex)';
COMMENT ON COLUMN dict_cids.key_type IS 'PIX key type: CPF, CNPJ, PHONE, EMAIL, or EVP';
COMMENT ON COLUMN dict_cids.algorithm_version IS 'CID generation algorithm version for future compatibility';
```

### Migration 002: `dict_vsyncs` Table

**File**: `apps/orchestration-worker/infrastructure/database/migrations/002_create_dict_vsyncs.sql`

```sql
-- Migration: Create dict_vsyncs table for storing VSync values
-- Author: Sync Squad
-- Date: 2025-10-28

CREATE TABLE IF NOT EXISTS dict_vsyncs (
    id                      SERIAL PRIMARY KEY,

    -- VSync per Key Type
    key_type                VARCHAR(10) NOT NULL UNIQUE,        -- CPF|CNPJ|PHONE|EMAIL|EVP
    vsync_value             VARCHAR(64) NOT NULL,               -- XOR cumulative (256 bits hex)

    -- Statistics
    total_keys              BIGINT NOT NULL DEFAULT 0,

    -- Status
    last_calculated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    last_verified_at        TIMESTAMP,
    synchronized            BOOLEAN NOT NULL DEFAULT TRUE,

    -- Audit
    updated_at              TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT vsync_key_type_check CHECK (key_type IN ('CPF', 'CNPJ', 'PHONE', 'EMAIL', 'EVP'))
);

-- Index
CREATE INDEX IF NOT EXISTS idx_dict_vsyncs_key_type ON dict_vsyncs(key_type);
CREATE INDEX IF NOT EXISTS idx_dict_vsyncs_synchronized ON dict_vsyncs(synchronized);

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_dict_vsyncs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_dict_vsyncs_updated_at
    BEFORE UPDATE ON dict_vsyncs
    FOR EACH ROW
    EXECUTE FUNCTION update_dict_vsyncs_updated_at();

-- Initialize 5 rows (one per key type) with zero VSyncs
INSERT INTO dict_vsyncs (key_type, vsync_value, total_keys, synchronized) VALUES
    ('CPF',   '0000000000000000000000000000000000000000000000000000000000000000', 0, TRUE),
    ('CNPJ',  '0000000000000000000000000000000000000000000000000000000000000000', 0, TRUE),
    ('PHONE', '0000000000000000000000000000000000000000000000000000000000000000', 0, TRUE),
    ('EMAIL', '0000000000000000000000000000000000000000000000000000000000000000', 0, TRUE),
    ('EVP',   '0000000000000000000000000000000000000000000000000000000000000000', 0, TRUE)
ON CONFLICT (key_type) DO NOTHING;

-- Comments
COMMENT ON TABLE dict_vsyncs IS 'Stores VSync (Verification Synchronizer) for each PIX key type';
COMMENT ON COLUMN dict_vsyncs.vsync_value IS 'XOR cumulative of all CIDs for this key type';
COMMENT ON COLUMN dict_vsyncs.synchronized IS 'FALSE when local VSync != DICT VSync';
```

### Migration 003: `dict_sync_verifications` Table

**File**: `apps/orchestration-worker/infrastructure/database/migrations/003_create_dict_sync_verifications.sql`

```sql
-- Migration: Create dict_sync_verifications audit log
-- Author: Sync Squad
-- Date: 2025-10-28

CREATE TABLE IF NOT EXISTS dict_sync_verifications (
    id                          BIGSERIAL PRIMARY KEY,

    -- Verification Context
    key_type                    VARCHAR(10) NOT NULL,
    verification_type           VARCHAR(20) NOT NULL DEFAULT 'SCHEDULED',  -- SCHEDULED|MANUAL|TRIGGERED

    -- VSyncs
    vsync_local                 VARCHAR(64) NOT NULL,
    vsync_dict                  VARCHAR(64),

    -- Result
    synchronized                BOOLEAN NOT NULL,
    total_keys_local            BIGINT NOT NULL,
    total_keys_dict             BIGINT,

    -- Divergences
    divergence_count            BIGINT DEFAULT 0,

    -- Audit
    verified_at                 TIMESTAMP NOT NULL DEFAULT NOW(),
    verified_by                 VARCHAR(100),                               -- Workflow ID

    -- Follow-up
    reconciliation_triggered    BOOLEAN DEFAULT FALSE,
    reconciliation_workflow_id  VARCHAR(255),

    -- Constraints
    CONSTRAINT verification_key_type_check CHECK (key_type IN ('CPF', 'CNPJ', 'PHONE', 'EMAIL', 'EVP')),
    CONSTRAINT verification_type_check CHECK (verification_type IN ('SCHEDULED', 'MANUAL', 'TRIGGERED'))
);

-- Indexes for audit queries
CREATE INDEX IF NOT EXISTS idx_sync_verifications_key_type ON dict_sync_verifications(key_type);
CREATE INDEX IF NOT EXISTS idx_sync_verifications_verified_at ON dict_sync_verifications(verified_at DESC);
CREATE INDEX IF NOT EXISTS idx_sync_verifications_synchronized ON dict_sync_verifications(synchronized);
CREATE INDEX IF NOT EXISTS idx_sync_verifications_reconciliation ON dict_sync_verifications(reconciliation_triggered);

-- Comments
COMMENT ON TABLE dict_sync_verifications IS 'Audit log of all VSync verification runs';
COMMENT ON COLUMN dict_sync_verifications.verified_by IS 'Temporal Workflow ID that ran verification';
```

### Migration 004: `dict_reconciliations` Table

**File**: `apps/orchestration-worker/infrastructure/database/migrations/004_create_dict_reconciliations.sql`

```sql
-- Migration: Create dict_reconciliations tracking table
-- Author: Sync Squad
-- Date: 2025-10-28

CREATE TABLE IF NOT EXISTS dict_reconciliations (
    id                          BIGSERIAL PRIMARY KEY,

    -- Context
    key_type                    VARCHAR(10) NOT NULL,
    verification_id             BIGINT REFERENCES dict_sync_verifications(id),

    -- Status
    status                      VARCHAR(20) NOT NULL,           -- REQUESTED|DOWNLOADING|PROCESSING|COMPLETED|FAILED

    -- DICT CID List
    dict_file_url               TEXT,
    dict_file_size_bytes        BIGINT,
    dict_total_cids             BIGINT,

    -- Divergences
    missing_local               BIGINT DEFAULT 0,
    missing_dict                BIGINT DEFAULT 0,
    total_divergences           BIGINT DEFAULT 0,

    -- Actions Taken
    keys_added                  BIGINT DEFAULT 0,
    keys_removed                BIGINT DEFAULT 0,

    -- Core-Dict Notification
    core_dict_notified          BOOLEAN DEFAULT FALSE,
    core_dict_notification_at   TIMESTAMP,

    -- Audit
    started_at                  TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at                TIMESTAMP,
    error_message               TEXT,

    -- Workflow
    workflow_id                 VARCHAR(255),

    -- Constraints
    CONSTRAINT reconciliation_key_type_check CHECK (key_type IN ('CPF', 'CNPJ', 'PHONE', 'EMAIL', 'EVP')),
    CONSTRAINT reconciliation_status_check CHECK (status IN ('REQUESTED', 'DOWNLOADING', 'PROCESSING', 'COMPLETED', 'FAILED'))
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_reconciliations_key_type ON dict_reconciliations(key_type);
CREATE INDEX IF NOT EXISTS idx_reconciliations_status ON dict_reconciliations(status);
CREATE INDEX IF NOT EXISTS idx_reconciliations_started_at ON dict_reconciliations(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_reconciliations_workflow_id ON dict_reconciliations(workflow_id);

-- Comments
COMMENT ON TABLE dict_reconciliations IS 'Tracking table for reconciliation processes';
COMMENT ON COLUMN dict_reconciliations.dict_file_url IS 'URL to download full CID list from DICT BACEN';
```

---

## 📡 Pulsar Integration (Connector-Dict Patterns)

### Topics Configuration

**File**: `apps/orchestration-worker/setup/config.go`

```go
type Config struct {
    // ... existing fields ...

    // Sync CID/VSync Topics
    PulsarTopicDictKeyCreated string `mapstructure:"PULSAR_TOPIC_DICT_KEY_CREATED"`
    PulsarTopicDictKeyUpdated string `mapstructure:"PULSAR_TOPIC_DICT_KEY_UPDATED"`

    // VSync Configuration
    VSyncVerificationCron     string `mapstructure:"VSYNC_VERIFICATION_CRON"`
    VSyncVerificationEnabled  bool   `mapstructure:"VSYNC_VERIFICATION_ENABLED"`
    VSyncAutoReconcileMaxDiv  int    `mapstructure:"VSYNC_AUTO_RECONCILE_MAX_DIVERGENCES"`
}

func NewConfigurationFromEnv() Config {
    // ... existing code ...

    viper.SetDefault("PULSAR_TOPIC_DICT_KEY_CREATED", "persistent://lb-conn/dict/dict-key-created")
    viper.SetDefault("PULSAR_TOPIC_DICT_KEY_UPDATED", "persistent://lb-conn/dict/dict-key-updated")
    viper.SetDefault("VSYNC_VERIFICATION_CRON", "0 3 * * *")  // Daily at 03:00 AM
    viper.SetDefault("VSYNC_VERIFICATION_ENABLED", true)
    viper.SetDefault("VSYNC_AUTO_RECONCILE_MAX_DIVERGENCES", 100)

    return Config{
        // ... existing fields ...
        PulsarTopicDictKeyCreated:    viper.GetString("PULSAR_TOPIC_DICT_KEY_CREATED"),
        PulsarTopicDictKeyUpdated:    viper.GetString("PULSAR_TOPIC_DICT_KEY_UPDATED"),
        VSyncVerificationCron:         viper.GetString("VSYNC_VERIFICATION_CRON"),
        VSyncVerificationEnabled:      viper.GetBool("VSYNC_VERIFICATION_ENABLED"),
        VSyncAutoReconcileMaxDiv:      viper.GetInt("VSYNC_AUTO_RECONCILE_MAX_DIVERGENCES"),
    }
}
```

### Event Schema (Expected from Dict API)

```go
// pkg.MessageProperties (already exists in connector-dict)
type MessageProperties struct {
    Action        Action
    CorrelationID string
}

// Event payload (from dict API when key is created/updated)
type KeyEventPayload struct {
    // Key
    Key     string `json:"key"`
    KeyType string `json:"key_type"`  // CPF|CNPJ|PHONE|EMAIL|EVP

    // Account
    ISPB               string    `json:"ispb"`
    Branch             *string   `json:"branch"`
    AccountNumber      string    `json:"account_number"`
    AccountType        string    `json:"account_type"`
    AccountOpenedAt    time.Time `json:"account_opened_at"`

    // Owner
    OwnerType          string    `json:"owner_type"`
    OwnerTaxID         string    `json:"owner_tax_id"`
    OwnerName          string    `json:"owner_name"`
    OwnerTradeName     *string   `json:"owner_trade_name"`

    // Metadata
    RegisteredAt               time.Time `json:"registered_at"`
    ParticipantRegisteredAt    time.Time `json:"participant_registered_at"`
    RequestID                  string    `json:"request_id"`
}
```

### Pulsar Handler Implementation

**File**: `apps/orchestration-worker/handlers/pulsar/sync/sync_handler.go`

```go
package sync

import (
    "github.com/lb-conn/connector-dict/apps/orchestration-worker/application/usecases/sync"
    "github.com/lb-conn/connector-dict/shared/infrastructure/observability/interfaces"
)

type Handler struct {
    syncApp     *sync.Application
    obsProvider interfaces.Provider
}

func NewHandler(syncApp *sync.Application, obsProvider interfaces.Provider) *Handler {
    return &Handler{
        syncApp:     syncApp,
        obsProvider: obsProvider,
    }
}
```

**File**: `apps/orchestration-worker/handlers/pulsar/sync/key_created_handler.go`

```go
package sync

import (
    "context"
    "github.com/lb-conn/libutils/pubsub"
    "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg"
)

func (h *Handler) KeyCreatedHandler(ctx context.Context, message pubsub.Message) error {
    logger := h.obsProvider.Logger()

    // Parse message properties
    props, err := pkg.ParseMessageProperties(message.Properties)
    if err != nil {
        logger.Error(ctx, "failed to parse message properties", err)
        return err
    }

    // Decode message payload
    var payload KeyEventPayload
    if err := message.Decode(&payload); err != nil {
        logger.Error(ctx, "failed to decode key.created event", err)
        return err
    }

    logger.Info(ctx, "processing key.created event",
        "correlation_id", props.CorrelationID,
        "key_type", payload.KeyType,
        "key", payload.Key,
    )

    // Delegate to application use case
    return h.syncApp.ProcessKeyCreated(ctx, props.CorrelationID, &payload)
}
```

---

## ⏱️ Temporal Workflows (Following Connector-Dict Patterns)

### Workflow 1: VSync Verification (Cron-Based)

**File**: `apps/orchestration-worker/infrastructure/temporal/workflows/sync/vsync_verification_workflow.go`

```go
package sync

import (
    "fmt"
    "time"
    "go.temporal.io/sdk/workflow"
    "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/activities"
    syncActivities "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/activities/sync"
)

const VSyncVerificationWorkflowName = "VSyncVerificationWorkflow"

// VSyncVerificationWorkflow verifies synchronization daily
// Cron: "0 3 * * *" (daily at 03:00 AM)
func VSyncVerificationWorkflow(ctx workflow.Context) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Starting VSync verification workflow")

    // Step 1: Read all VSyncs from PostgreSQL
    var localVSyncs map[string]string  // key_type → vsync_value
    ctx = workflow.WithActivityOptions(ctx, activities.DBOptions)
    err := workflow.ExecuteActivity(ctx, syncActivities.ReadVSyncsActivityName).Get(ctx, &localVSyncs)
    if err != nil {
        logger.Error("Failed to read local VSyncs", "error", err)
        return err
    }

    logger.Info("Read local VSyncs", "count", len(localVSyncs))

    // Step 2: Call Bridge gRPC → DICT API
    var dictVSyncs map[string]SyncResult  // key_type → {status, vsync, total_keys}
    ctx = workflow.WithActivityOptions(ctx, activities.GRPCOptions)
    err = workflow.ExecuteActivity(ctx, syncActivities.BridgeVerifySyncActivityName, localVSyncs).Get(ctx, &dictVSyncs)
    if err != nil {
        logger.Error("Failed to verify sync with DICT", "error", err)
        return err
    }

    // Step 3: Compare VSyncs and identify divergences
    divergences := []string{}
    for keyType, localVSync := range localVSyncs {
        dictResult, exists := dictVSyncs[keyType]
        if !exists || localVSync != dictResult.VSync {
            divergences = append(divergences, keyType)
            logger.Warn("Desynchronization detected",
                "key_type", keyType,
                "local_vsync", localVSync,
                "dict_vsync", dictResult.VSync,
            )
        }
    }

    // Step 4: Log verification result
    ctx = workflow.WithActivityOptions(ctx, activities.DBOptions)
    err = workflow.ExecuteActivity(ctx, syncActivities.LogVerificationActivityName, VerificationLog{
        VSyncsLocal: localVSyncs,
        VSyncsDict:  dictVSyncs,
        Divergences: divergences,
    }).Get(ctx, nil)
    if err != nil {
        logger.Error("Failed to log verification result", "error", err)
        return err
    }

    // Step 5: Trigger reconciliation if needed
    if len(divergences) > 0 {
        logger.Info("Triggering reconciliation workflows", "divergences", len(divergences))

        for _, keyType := range divergences {
            // Start child workflow for each key type
            childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
                WorkflowID:        fmt.Sprintf("%s_%s_%d", ReconciliationWorkflowName, keyType, workflow.Now(ctx).Unix()),
                ParentClosePolicy: enums.PARENT_CLOSE_POLICY_ABANDON,
            })

            childWorkflow := workflow.ExecuteChildWorkflow(childCtx, ReconciliationWorkflow, ReconciliationInput{
                KeyType: keyType,
            })

            // Don't wait for child workflow (async)
            var execution workflow.Execution
            _ = childWorkflow.GetChildWorkflowExecution().Get(ctx, &execution)
            logger.Info("Started reconciliation workflow", "key_type", keyType, "workflow_id", execution.ID)
        }
    }

    logger.Info("VSync verification completed",
        "total_types", len(localVSyncs),
        "synchronized", len(localVSyncs)-len(divergences),
        "divergences", len(divergences),
    )

    return nil
}
```

### Workflow 2: Reconciliation (Child Workflow)

**File**: `apps/orchestration-worker/infrastructure/temporal/workflows/sync/reconciliation_workflow.go`

```go
package sync

import (
    "fmt"
    "time"
    "go.temporal.io/sdk/workflow"
    "go.temporal.io/sdk/temporal"
    "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/activities"
    syncActivities "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/activities/sync"
    "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg"
)

const ReconciliationWorkflowName = "ReconciliationWorkflow"

type ReconciliationInput struct {
    KeyType string
}

func ReconciliationWorkflow(ctx workflow.Context, input ReconciliationInput) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Starting reconciliation workflow", "key_type", input.KeyType)

    // Step 1: Request CID list from DICT
    var requestResp RequestCIDListResponse
    ctx = workflow.WithActivityOptions(ctx, activities.GRPCOptions)
    err := workflow.ExecuteActivity(ctx, syncActivities.BridgeRequestCIDListActivityName, input.KeyType).Get(ctx, &requestResp)
    if err != nil {
        logger.Error("Failed to request CID list", "error", err)
        return err
    }

    logger.Info("CID list requested", "request_id", requestResp.RequestID, "status", requestResp.Status)

    // Step 2: Poll for completion (max 5 minutes)
    var statusResp GetCIDListStatusResponse
    maxPolls := 60  // 60 * 5s = 5 minutes
    for i := 0; i < maxPolls; i++ {
        if err := workflow.Sleep(ctx, 5*time.Second); err != nil {
            return err
        }

        err = workflow.ExecuteActivity(ctx, syncActivities.BridgeGetCIDListStatusActivityName, requestResp.RequestID).Get(ctx, &statusResp)
        if err != nil {
            logger.Error("Failed to get CID list status", "error", err)
            return err
        }

        if statusResp.Status == "COMPLETED" {
            break
        } else if statusResp.Status == "FAILED" {
            return fmt.Errorf("DICT failed to generate CID list: %s", statusResp.ErrorMessage)
        }
    }

    if statusResp.Status != "COMPLETED" {
        return temporal.NewNonRetryableApplicationError(
            "CID list generation timed out after 5 minutes",
            "CIDListTimeout",
            nil,
        )
    }

    logger.Info("CID list ready", "url", statusResp.DownloadURL, "total_cids", statusResp.TotalCIDs)

    // Step 3: Download and parse CID list
    var dictCIDs []string
    err = workflow.ExecuteActivity(ctx, syncActivities.BridgeDownloadCIDListActivityName, statusResp.DownloadURL).Get(ctx, &dictCIDs)
    if err != nil {
        logger.Error("Failed to download CID list", "error", err)
        return err
    }

    // Step 4: Compare CIDs
    var divergences Divergences
    ctx = workflow.WithActivityOptions(ctx, activities.DBOptions)
    err = workflow.ExecuteActivity(ctx, syncActivities.CompareCIDsActivityName, CompareCIDsInput{
        KeyType:  input.KeyType,
        DictCIDs: dictCIDs,
    }).Get(ctx, &divergences)
    if err != nil {
        logger.Error("Failed to compare CIDs", "error", err)
        return err
    }

    logger.Info("CID comparison completed",
        "missing_local", divergences.MissingLocal,
        "missing_dict", divergences.MissingDict,
        "total", divergences.Total,
    )

    // Step 5: Check threshold for auto-reconcile
    // TODO: Read threshold from config (injected via activity)
    if divergences.Total > 100 {
        logger.Warn("Divergences exceed threshold - manual approval required",
            "divergences", divergences.Total,
            "threshold", 100,
        )

        // Send alert activity
        _ = workflow.ExecuteActivity(ctx, syncActivities.SendAlertActivityName, Alert{
            Severity: "CRITICAL",
            Title:    fmt.Sprintf("VSync Reconciliation Requires Manual Approval: %s", input.KeyType),
            Message:  fmt.Sprintf("Found %d divergences (threshold: 100)", divergences.Total),
        }).Get(ctx, nil)

        return temporal.NewNonRetryableApplicationError(
            "Manual approval required",
            "ManualApprovalRequired",
            map[string]interface{}{
                "divergences": divergences.Total,
                "threshold":   100,
            },
        )
    }

    // Step 6: Notify Core-Dict (Pulsar event)
    ctx = workflow.WithActivityOptions(ctx, activities.PublishEventOptions)
    err = workflow.ExecuteActivity(ctx, syncActivities.NotifyCoreDictActivityName, CoreDictNotification{
        KeyType:          input.KeyType,
        DivergenceCount:  divergences.Total,
        DictCIDFileURL:   statusResp.DownloadURL,
        ActionRequired:   "REBUILD_TABLE",
    }).Get(ctx, nil)
    if err != nil {
        logger.Error("Failed to notify Core-Dict", "error", err)
        return err
    }

    // Step 7: Apply corrections (optional, based on policy)
    // For now, just log - Core-Dict will rebuild based on DICT file

    // Step 8: Recalculate VSyncs
    err = workflow.ExecuteActivity(ctx, syncActivities.RecalculateVSyncsActivityName, input.KeyType).Get(ctx, nil)
    if err != nil {
        logger.Error("Failed to recalculate VSyncs", "error", err)
        return err
    }

    // Step 9: Save reconciliation log
    ctx = workflow.WithActivityOptions(ctx, activities.DBOptions)
    err = workflow.ExecuteActivity(ctx, syncActivities.SaveReconciliationLogActivityName, ReconciliationLog{
        KeyType:        input.KeyType,
        Status:         "COMPLETED",
        Divergences:    divergences,
        DictFileURL:    statusResp.DownloadURL,
        DictTotalCIDs:  statusResp.TotalCIDs,
    }).Get(ctx, nil)
    if err != nil {
        logger.Error("Failed to save reconciliation log", "error", err)
        return err
    }

    logger.Info("Reconciliation completed successfully", "key_type", input.KeyType)
    return nil
}
```

---

## 🔌 gRPC Integration (Bridge Client)

### Proto Definitions

**File**: `shared/proto/sync/dict_sync_service.proto`

```protobuf
syntax = "proto3";

package dict.sync.v1;

option go_package = "github.com/lb-conn/connector-dict/shared/proto/sync;syncpb";

// DICT Sync Service (to be implemented in Bridge)
service DICTSyncService {
    // Verify synchronization using VSyncs
    rpc VerifySync(VerifySyncRequest) returns (VerifySyncResponse);

    // Request full CID list for reconciliation
    rpc RequestCIDList(RequestCIDListRequest) returns (RequestCIDListResponse);

    // Check status of CID list generation
    rpc GetCIDListStatus(GetCIDListStatusRequest) returns (GetCIDListStatusResponse);

    // Download CID list (optional if using direct URL)
    rpc DownloadCIDList(DownloadCIDListRequest) returns (stream DownloadCIDListResponse);
}

message VerifySyncRequest {
    string participant_ispb = 1;
    map<string, string> vsyncs = 2;  // key_type → vsync_value
}

message VerifySyncResponse {
    map<string, SyncResult> results = 1;  // key_type → result
}

message SyncResult {
    string status = 1;  // OK | DESYNC
    string vsync = 2;   // DICT VSync value (if desync)
    int64 total_keys = 3;
}

message RequestCIDListRequest {
    string participant_ispb = 1;
    string key_type = 2;  // CPF|CNPJ|PHONE|EMAIL|EVP
    string format = 3;    // CSV|JSON
}

message RequestCIDListResponse {
    string request_id = 1;
    string status = 2;  // ACCEPTED | PROCESSING
}

message GetCIDListStatusRequest {
    string request_id = 1;
}

message GetCIDListStatusResponse {
    string status = 1;  // PROCESSING | COMPLETED | FAILED
    string download_url = 2;  // If completed
    int64 total_cids = 3;
    int64 file_size_bytes = 4;
    string error_message = 5;  // If failed
}

message DownloadCIDListRequest {
    string url = 1;
}

message DownloadCIDListResponse {
    bytes chunk = 1;
}
```

---

## ✅ Acceptance Criteria (Updated)

### Fase 0: Análise Técnica
- [ ] Documentar eventos Pulsar existentes para Entry
- [ ] Identificar Entry domain entities reutilizáveis
- [ ] Validar conexão PostgreSQL existente
- [ ] Confirmar com Bridge team: endpoints VSync disponíveis?
- [ ] Confirmar com Core-Dict team: consumer core-events existe?

### Database Layer
- [ ] 4 migrations SQL criadas e testadas
- [ ] Repository interfaces definidas (domain/sync/)
- [ ] Repository implementations (infrastructure/database/repositories/sync/)
- [ ] Unit tests >90% coverage
- [ ] Performance test: 10M CIDs insert <10s

### Algorithms
- [ ] CID Generator: 100% determinístico
- [ ] VSync Calculator: propriedades XOR verificadas
- [ ] Unit tests validados contra spec BACEN

### Pulsar Integration
- [ ] Handler implementado seguindo padrão connector-dict
- [ ] Integration tests com Testcontainers Pulsar
- [ ] Idempotency handling testado

### Temporal Workflows
- [ ] VSyncVerificationWorkflow: cron testado (replay)
- [ ] ReconciliationWorkflow: child workflow testado
- [ ] Todas activities implementadas com retry policies
- [ ] Workflow tests: replay, time skip, failure scenarios

### Bridge Integration
- [ ] Proto definitions (coordenado com Bridge team)
- [ ] gRPC client implementado
- [ ] Error handling: retryable vs non-retryable
- [ ] Integration tests com mock Bridge

### BACEN Compliance
- [ ] 100% conformidade Manual Cap. 9
- [ ] CID algorithm validado
- [ ] VSync algorithm validado
- [ ] Audit trail completo

---

**Status**: ✅ SPECS ATUALIZADAS COM PADRÕES CONNECTOR-DICT
**Next Step**: Iniciar Fase 0 - Análise do código existente
**Target Start Date**: Após validação técnica com stakeholders
