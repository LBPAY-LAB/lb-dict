-- +goose Up
-- +goose StatementBegin
-- Table: dict_rate_limit_states
-- Purpose: Time-series snapshots of rate limit token bucket states
-- Source: Retrieved every 5 minutes from DICT via Bridge gRPC GetRateLimitPolicy()
-- Retention: 13 months with monthly partitioning for performance
-- Partitioning: RANGE on created_at (monthly partitions)

-- Parent table (partitioned)
CREATE TABLE IF NOT EXISTS dict_rate_limit_states (
    -- Primary key
    id BIGSERIAL,

    -- Foreign key to policy
    endpoint_id VARCHAR(100) NOT NULL,

    -- Token bucket state at snapshot time
    available_tokens INTEGER NOT NULL CHECK (available_tokens >= 0),
    capacity INTEGER NOT NULL CHECK (capacity > 0),
    refill_tokens INTEGER NOT NULL CHECK (refill_tokens > 0),
    refill_period_sec INTEGER NOT NULL CHECK (refill_period_sec > 0),

    -- PSP category at snapshot time (may change over time)
    psp_category VARCHAR(2) CHECK (psp_category ~ '^[A-H]$'),

    -- Calculated metrics (NEW REQUIREMENTS)
    consumption_rate_per_minute DECIMAL(10,2) CHECK (consumption_rate_per_minute >= 0),
    recovery_eta_seconds INTEGER CHECK (recovery_eta_seconds >= 0),
    exhaustion_projection_seconds INTEGER,
    error_404_rate DECIMAL(5,2) CHECK (error_404_rate >= 0 AND error_404_rate <= 100),

    -- Timestamps (UTC enforced)
    response_timestamp TIMESTAMPTZ NOT NULL, -- From DICT <ResponseTime> XML field
    created_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'), -- When we stored this record

    -- Partition key constraint (will be set by partitions)
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Create 13 monthly partitions (current month + 12 previous months)
-- These will be created dynamically in production, but we create initial set here

-- Function to create monthly partition
CREATE OR REPLACE FUNCTION create_dict_rate_limit_state_partition(partition_date DATE)
RETURNS VOID AS $$
DECLARE
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
BEGIN
    partition_name := 'dict_rate_limit_states_' || TO_CHAR(partition_date, 'YYYY_MM');
    start_date := DATE_TRUNC('month', partition_date);
    end_date := start_date + INTERVAL '1 month';

    EXECUTE FORMAT(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF dict_rate_limit_states
         FOR VALUES FROM (%L) TO (%L)',
        partition_name,
        start_date,
        end_date
    );

    -- Create indexes on partition
    EXECUTE FORMAT(
        'CREATE INDEX IF NOT EXISTS %I ON %I (endpoint_id, created_at DESC)',
        partition_name || '_endpoint_time_idx',
        partition_name
    );

    EXECUTE FORMAT(
        'CREATE INDEX IF NOT EXISTS %I ON %I (psp_category, created_at DESC) WHERE psp_category IS NOT NULL',
        partition_name || '_category_time_idx',
        partition_name
    );
END;
$$ LANGUAGE plpgsql;

-- Create initial 13 partitions (current month + 12 previous months)
DO $$
DECLARE
    i INTEGER;
    partition_date DATE;
BEGIN
    FOR i IN 0..12 LOOP
        partition_date := DATE_TRUNC('month', NOW() AT TIME ZONE 'UTC') - (i || ' months')::INTERVAL;
        PERFORM create_dict_rate_limit_state_partition(partition_date);
    END LOOP;
END;
$$;

-- Foreign key to policies table
ALTER TABLE dict_rate_limit_states
    ADD CONSTRAINT fk_state_policy
    FOREIGN KEY (endpoint_id)
    REFERENCES dict_rate_limit_policies(endpoint_id)
    ON DELETE CASCADE;

-- Table comments
COMMENT ON TABLE dict_rate_limit_states IS 'Time-series snapshots of DICT rate limit states (5-min frequency, 13-month retention)';
COMMENT ON COLUMN dict_rate_limit_states.available_tokens IS 'Current available tokens in bucket at snapshot time';
COMMENT ON COLUMN dict_rate_limit_states.consumption_rate_per_minute IS 'Calculated consumption rate (tokens per minute)';
COMMENT ON COLUMN dict_rate_limit_states.recovery_eta_seconds IS 'Estimated seconds until full token recovery';
COMMENT ON COLUMN dict_rate_limit_states.exhaustion_projection_seconds IS 'Estimated seconds until token exhaustion if trend continues';
COMMENT ON COLUMN dict_rate_limit_states.error_404_rate IS 'Percentage of 404 errors in recent requests';
COMMENT ON COLUMN dict_rate_limit_states.response_timestamp IS 'Timestamp from DICT response (authoritative time)';
COMMENT ON COLUMN dict_rate_limit_states.created_at IS 'Timestamp when we stored this snapshot (partition key)';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop function
DROP FUNCTION IF EXISTS create_dict_rate_limit_state_partition(DATE);

-- Drop all partitions (will be done automatically with parent table drop)
-- Drop parent table (cascades to all partitions)
DROP TABLE IF EXISTS dict_rate_limit_states;
-- +goose StatementEnd
