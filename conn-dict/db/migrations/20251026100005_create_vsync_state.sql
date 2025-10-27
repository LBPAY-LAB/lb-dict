-- Migration: 20251026100005_create_vsync_state
-- Description: Create vsync_state table for tracking daily VSYNC operations

CREATE TABLE IF NOT EXISTS vsync_state (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Sync information
    sync_date DATE NOT NULL UNIQUE,
    sync_type VARCHAR(20) NOT NULL DEFAULT 'FULL' CHECK (sync_type IN ('FULL', 'INCREMENTAL')),

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN (
        'PENDING',           -- Sync scheduled
        'RUNNING',           -- Sync in progress
        'COMPLETED',         -- Sync completed successfully
        'FAILED',            -- Sync failed
        'PARTIAL'            -- Partial sync (some errors)
    )),

    -- Metrics
    total_entries INTEGER DEFAULT 0,
    processed_entries INTEGER DEFAULT 0,
    successful_entries INTEGER DEFAULT 0,
    failed_entries INTEGER DEFAULT 0,

    -- Timestamps
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,

    -- Error tracking
    error_message TEXT,
    error_count INTEGER DEFAULT 0,

    -- Temporal workflow tracking
    temporal_workflow_id VARCHAR(200),

    -- Metadata
    bacen_sync_id VARCHAR(100),
    checksum VARCHAR(64),
    metadata JSONB
);

-- Indexes
CREATE INDEX idx_vsync_state_sync_date ON vsync_state(sync_date DESC);
CREATE INDEX idx_vsync_state_status ON vsync_state(status);
CREATE INDEX idx_vsync_state_temporal_workflow_id ON vsync_state(temporal_workflow_id) WHERE temporal_workflow_id IS NOT NULL;

-- Comments
COMMENT ON TABLE vsync_state IS 'Tracks daily VSYNC operations with Bacen';
COMMENT ON COLUMN vsync_state.sync_date IS 'Date of the sync operation (one per day)';
COMMENT ON COLUMN vsync_state.total_entries IS 'Total entries to sync';
COMMENT ON COLUMN vsync_state.checksum IS 'Checksum for data integrity verification';