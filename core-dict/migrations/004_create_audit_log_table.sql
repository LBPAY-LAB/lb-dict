-- Migration: 004_create_audit_log_table
-- Description: Create audit log table with partitioning by month
-- Author: data-specialist-core
-- Date: 2025-10-27

-- +goose Up
-- +goose StatementBegin

-- Create audit.entry_events table
CREATE TABLE audit.entry_events (
    id                  BIGSERIAL,
    event_id            UUID NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    entity_type         VARCHAR(50) NOT NULL,
    entity_id           UUID NOT NULL,
    event_type          VARCHAR(100) NOT NULL,
    event_subtype       VARCHAR(100),
    old_values          JSONB,
    new_values          JSONB,
    diff                JSONB,
    user_id             UUID REFERENCES core_dict.users(id),
    ip_address          INET,
    user_agent          TEXT,
    occurred_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata            JSONB,

    PRIMARY KEY (id, occurred_at)
) PARTITION BY RANGE (occurred_at);

COMMENT ON TABLE audit.entry_events IS 'Complete audit log for Core DICT (LGPD/Bacen compliance)';
COMMENT ON COLUMN audit.entry_events.entity_type IS 'Type: ENTRY, CLAIM, PORTABILITY, ACCOUNT';
COMMENT ON COLUMN audit.entry_events.event_type IS 'Event: CREATED, UPDATED, DELETED, SYNCED';
COMMENT ON COLUMN audit.entry_events.diff IS 'Computed difference between old_values and new_values';

-- Create partitions for current year and next year
CREATE TABLE audit.entry_events_2025_10 PARTITION OF audit.entry_events
    FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');

CREATE TABLE audit.entry_events_2025_11 PARTITION OF audit.entry_events
    FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');

CREATE TABLE audit.entry_events_2025_12 PARTITION OF audit.entry_events
    FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');

CREATE TABLE audit.entry_events_2026_01 PARTITION OF audit.entry_events
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE audit.entry_events_2026_02 PARTITION OF audit.entry_events
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

CREATE TABLE audit.entry_events_2026_03 PARTITION OF audit.entry_events
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

-- Create default partition for future dates
CREATE TABLE audit.entry_events_default PARTITION OF audit.entry_events DEFAULT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS audit.entry_events CASCADE;

-- +goose StatementEnd
