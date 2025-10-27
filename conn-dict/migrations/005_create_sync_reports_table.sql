-- +goose Up
-- +goose StatementBegin
-- VSYNC Reports table: stores daily synchronization results with Bacen DICT
CREATE TABLE sync_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Sync execution details
    sync_id VARCHAR(50) UNIQUE NOT NULL,  -- Temporal workflow execution ID
    sync_type VARCHAR(20) NOT NULL CHECK (sync_type IN ('FULL', 'INCREMENTAL')),
    sync_timestamp TIMESTAMPTZ NOT NULL,

    -- Participant information
    participant_ispb VARCHAR(8) NOT NULL,  -- ISPB code of the participant

    -- Statistics
    entries_fetched INTEGER NOT NULL DEFAULT 0,     -- Total entries fetched from Bacen
    entries_compared INTEGER NOT NULL DEFAULT 0,    -- Total entries compared
    entries_synced INTEGER NOT NULL DEFAULT 0,      -- Total entries synced
    entries_created INTEGER NOT NULL DEFAULT 0,     -- New entries created locally
    entries_updated INTEGER NOT NULL DEFAULT 0,     -- Existing entries updated
    entries_deleted INTEGER NOT NULL DEFAULT 0,     -- Entries deleted (rare, manual approval)

    -- Discrepancy tracking
    discrepancies_found INTEGER NOT NULL DEFAULT 0,        -- Total discrepancies detected
    discrepancies_missing_local INTEGER NOT NULL DEFAULT 0,   -- Missing in local DB
    discrepancies_outdated_local INTEGER NOT NULL DEFAULT 0,  -- Outdated in local DB
    discrepancies_missing_bacen INTEGER NOT NULL DEFAULT 0,   -- Missing in Bacen (needs review)

    -- Execution details
    status VARCHAR(20) NOT NULL CHECK (status IN ('COMPLETED', 'PARTIAL', 'FAILED')),
    duration_ms INTEGER,  -- Duration in milliseconds

    -- Error tracking
    error_message TEXT,
    error_code VARCHAR(50),

    -- Metadata
    metadata JSONB,  -- Additional context (e.g., filters used, retry attempts)

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (sync_timestamp);

-- Create partitions for current and next 12 months
CREATE TABLE sync_reports_2025_10 PARTITION OF sync_reports
    FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');

CREATE TABLE sync_reports_2025_11 PARTITION OF sync_reports
    FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');

CREATE TABLE sync_reports_2025_12 PARTITION OF sync_reports
    FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');

CREATE TABLE sync_reports_2026_01 PARTITION OF sync_reports
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE sync_reports_2026_02 PARTITION OF sync_reports
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

CREATE TABLE sync_reports_2026_03 PARTITION OF sync_reports
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

-- Indexes on partitioned table
CREATE INDEX idx_sync_reports_sync_id ON sync_reports(sync_id);
CREATE INDEX idx_sync_reports_participant ON sync_reports(participant_ispb, sync_timestamp DESC);
CREATE INDEX idx_sync_reports_status ON sync_reports(status, sync_timestamp DESC);
CREATE INDEX idx_sync_reports_sync_type ON sync_reports(sync_type, sync_timestamp DESC);
CREATE INDEX idx_sync_reports_timestamp ON sync_reports(sync_timestamp DESC);
CREATE INDEX idx_sync_reports_discrepancies ON sync_reports(discrepancies_found DESC, sync_timestamp DESC)
    WHERE discrepancies_found > 0;

-- GIN index for metadata JSONB
CREATE INDEX idx_sync_reports_metadata ON sync_reports USING GIN (metadata);

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_sync_reports_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update updated_at on UPDATE
CREATE TRIGGER trigger_sync_reports_updated_at
    BEFORE UPDATE ON sync_reports
    FOR EACH ROW
    EXECUTE FUNCTION update_sync_reports_updated_at();

-- View for quick sync report summary (last 30 days)
CREATE VIEW sync_reports_summary AS
SELECT
    participant_ispb,
    sync_type,
    COUNT(*) as total_syncs,
    SUM(entries_synced) as total_entries_synced,
    SUM(discrepancies_found) as total_discrepancies,
    AVG(duration_ms) as avg_duration_ms,
    MAX(sync_timestamp) as last_sync_timestamp,
    SUM(CASE WHEN status = 'COMPLETED' THEN 1 ELSE 0 END) as completed_syncs,
    SUM(CASE WHEN status = 'FAILED' THEN 1 ELSE 0 END) as failed_syncs
FROM sync_reports
WHERE sync_timestamp >= NOW() - INTERVAL '30 days'
GROUP BY participant_ispb, sync_type
ORDER BY participant_ispb, sync_type;

-- Comments
COMMENT ON TABLE sync_reports IS 'VSYNC workflow execution results - daily reconciliation with Bacen DICT (partitioned by month, retained for 7 years per Bacen compliance)';
COMMENT ON COLUMN sync_reports.sync_id IS 'Temporal workflow execution ID for traceability';
COMMENT ON COLUMN sync_reports.sync_type IS 'FULL = complete sync, INCREMENTAL = only changed entries since last sync';
COMMENT ON COLUMN sync_reports.discrepancies_missing_bacen IS 'Critical: entries in local DB but not in Bacen - requires manual investigation';
COMMENT ON COLUMN sync_reports.duration_ms IS 'Total sync duration in milliseconds (includes fetch, compare, and reconcile phases)';
COMMENT ON COLUMN sync_reports.metadata IS 'Additional context: last_sync_date for INCREMENTAL, filter criteria, retry count, etc.';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS sync_reports_summary;
DROP TRIGGER IF EXISTS trigger_sync_reports_updated_at ON sync_reports;
DROP FUNCTION IF EXISTS update_sync_reports_updated_at();
DROP TABLE IF EXISTS sync_reports_2025_10;
DROP TABLE IF EXISTS sync_reports_2025_11;
DROP TABLE IF EXISTS sync_reports_2025_12;
DROP TABLE IF EXISTS sync_reports_2026_01;
DROP TABLE IF EXISTS sync_reports_2026_02;
DROP TABLE IF EXISTS sync_reports_2026_03;
DROP TABLE IF EXISTS sync_reports;
-- +goose StatementEnd