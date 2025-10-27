-- +goose Up
-- +goose StatementBegin
-- Audit log table: comprehensive audit trail for all DICT operations
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Entity information
    entity_type VARCHAR(50) NOT NULL CHECK (entity_type IN ('CLAIM', 'ENTRY', 'INFRACTION', 'ACCOUNT', 'USER')),
    entity_id VARCHAR(50) NOT NULL,

    -- Operation details
    operation VARCHAR(50) NOT NULL CHECK (operation IN (
        'CREATE', 'UPDATE', 'DELETE', 'STATUS_CHANGE',
        'OWNERSHIP_TRANSFER', 'PORTABILITY', 'FRAUD_REPORT',
        'ACTIVATION', 'DEACTIVATION', 'BLOCK', 'UNBLOCK'
    )),

    -- Actor information
    actor_type VARCHAR(30) CHECK (actor_type IN ('USER', 'SYSTEM', 'WORKFLOW', 'API')),
    actor_id VARCHAR(50),
    actor_participant VARCHAR(8),  -- ISPB if applicable

    -- Change tracking
    old_values JSONB,  -- Previous state
    new_values JSONB,  -- New state
    changes JSONB,     -- Diff of changes

    -- Context
    reason TEXT,
    ip_address INET,
    user_agent TEXT,
    request_id VARCHAR(50),  -- Trace request across services

    -- Timestamp
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Metadata
    metadata JSONB  -- Additional context-specific data
) PARTITION BY RANGE (occurred_at);

-- Create partitions for current and next 12 months
CREATE TABLE audit_logs_2025_10 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');

CREATE TABLE audit_logs_2025_11 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');

CREATE TABLE audit_logs_2025_12 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');

CREATE TABLE audit_logs_2026_01 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

-- Indexes on partitioned table
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id, occurred_at DESC);
CREATE INDEX idx_audit_logs_actor ON audit_logs(actor_id, occurred_at DESC);
CREATE INDEX idx_audit_logs_participant ON audit_logs(actor_participant, occurred_at DESC) WHERE actor_participant IS NOT NULL;
CREATE INDEX idx_audit_logs_operation ON audit_logs(operation, occurred_at DESC);
CREATE INDEX idx_audit_logs_occurred_at ON audit_logs(occurred_at DESC);
CREATE INDEX idx_audit_logs_request_id ON audit_logs(request_id) WHERE request_id IS NOT NULL;

-- GIN index for JSONB columns (for querying JSON fields)
CREATE INDEX idx_audit_logs_old_values ON audit_logs USING GIN (old_values);
CREATE INDEX idx_audit_logs_new_values ON audit_logs USING GIN (new_values);
CREATE INDEX idx_audit_logs_changes ON audit_logs USING GIN (changes);
CREATE INDEX idx_audit_logs_metadata ON audit_logs USING GIN (metadata);

-- Event logs table: stores all domain events published to Pulsar
CREATE TABLE event_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Event identification
    event_id VARCHAR(50) UNIQUE NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_version VARCHAR(10) NOT NULL DEFAULT 'v1',

    -- Aggregate information
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id VARCHAR(50) NOT NULL,

    -- Event data
    payload JSONB NOT NULL,

    -- Causation and correlation
    causation_id VARCHAR(50),  -- ID of command that caused this event
    correlation_id VARCHAR(50), -- ID to correlate related events

    -- Publishing status
    published BOOLEAN NOT NULL DEFAULT FALSE,
    published_at TIMESTAMPTZ,
    publish_attempts INT NOT NULL DEFAULT 0,
    last_publish_error TEXT,

    -- Metadata
    metadata JSONB,

    -- Timestamp
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (occurred_at);

-- Create partitions for event logs
CREATE TABLE event_logs_2025_10 PARTITION OF event_logs
    FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');

CREATE TABLE event_logs_2025_11 PARTITION OF event_logs
    FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');

CREATE TABLE event_logs_2025_12 PARTITION OF event_logs
    FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');

CREATE TABLE event_logs_2026_01 PARTITION OF event_logs
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

-- Indexes for event logs
CREATE INDEX idx_event_logs_event_id ON event_logs(event_id);
CREATE INDEX idx_event_logs_aggregate ON event_logs(aggregate_type, aggregate_id, occurred_at DESC);
CREATE INDEX idx_event_logs_type ON event_logs(event_type, occurred_at DESC);
CREATE INDEX idx_event_logs_published ON event_logs(published, occurred_at) WHERE NOT published;
CREATE INDEX idx_event_logs_correlation ON event_logs(correlation_id) WHERE correlation_id IS NOT NULL;
CREATE INDEX idx_event_logs_occurred_at ON event_logs(occurred_at DESC);

-- GIN index for event payload
CREATE INDEX idx_event_logs_payload ON event_logs USING GIN (payload);

-- Function to automatically create next month partition
CREATE OR REPLACE FUNCTION create_next_partition(
    table_name TEXT,
    start_date DATE
) RETURNS VOID AS $$
DECLARE
    partition_name TEXT;
    end_date DATE;
BEGIN
    partition_name := table_name || '_' || TO_CHAR(start_date, 'YYYY_MM');
    end_date := start_date + INTERVAL '1 month';

    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF %I FOR VALUES FROM (%L) TO (%L)',
        partition_name,
        table_name,
        start_date,
        end_date
    );
END;
$$ LANGUAGE plpgsql;

-- Comments
COMMENT ON TABLE audit_logs IS 'Comprehensive audit trail for all DICT operations (partitioned by month)';
COMMENT ON TABLE event_logs IS 'Domain events log for event sourcing and Pulsar publishing (partitioned by month)';
COMMENT ON COLUMN audit_logs.old_values IS 'JSON snapshot of entity state before change';
COMMENT ON COLUMN audit_logs.new_values IS 'JSON snapshot of entity state after change';
COMMENT ON COLUMN audit_logs.changes IS 'JSON diff showing only what changed';
COMMENT ON COLUMN event_logs.published IS 'Whether event was successfully published to Pulsar';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS create_next_partition(TEXT, DATE);
DROP TABLE IF EXISTS event_logs_2025_10;
DROP TABLE IF EXISTS event_logs_2025_11;
DROP TABLE IF EXISTS event_logs_2025_12;
DROP TABLE IF EXISTS event_logs_2026_01;
DROP TABLE IF EXISTS event_logs;
DROP TABLE IF EXISTS audit_logs_2025_10;
DROP TABLE IF EXISTS audit_logs_2025_11;
DROP TABLE IF EXISTS audit_logs_2025_12;
DROP TABLE IF EXISTS audit_logs_2026_01;
DROP TABLE IF EXISTS audit_logs;
-- +goose StatementEnd