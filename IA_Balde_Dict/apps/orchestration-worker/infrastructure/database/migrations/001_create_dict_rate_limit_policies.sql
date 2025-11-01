-- +goose Up
-- +goose StatementBegin
-- Table: dict_rate_limit_policies
-- Purpose: Store static reference data for DICT API rate limit policies
-- Source: Retrieved from DICT BACEN via Bridge gRPC GetRateLimitPolicies()
-- Retention: Updated periodically, no partitioning needed

CREATE TABLE IF NOT EXISTS dict_rate_limit_policies (
    -- Primary identifier for the rate-limited endpoint
    endpoint_id VARCHAR(100) PRIMARY KEY,

    -- HTTP endpoint details
    endpoint_path VARCHAR(255) NOT NULL,
    http_method VARCHAR(10) NOT NULL CHECK (http_method IN ('GET', 'POST', 'PUT', 'DELETE', 'PATCH')),

    -- Token bucket configuration
    capacity INTEGER NOT NULL CHECK (capacity > 0),
    refill_tokens INTEGER NOT NULL CHECK (refill_tokens > 0),
    refill_period_sec INTEGER NOT NULL CHECK (refill_period_sec > 0),

    -- PSP category (A-H) - may vary by endpoint or be global
    psp_category VARCHAR(2) CHECK (psp_category ~ '^[A-H]$'),

    -- Audit timestamps (all in UTC)
    created_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),

    -- Unique constraint to prevent duplicate endpoint definitions
    CONSTRAINT unique_endpoint UNIQUE (endpoint_path, http_method)
);

-- Index for queries by category
CREATE INDEX idx_policies_category ON dict_rate_limit_policies(psp_category) WHERE psp_category IS NOT NULL;

-- Index for queries by endpoint path (for lookup)
CREATE INDEX idx_policies_path ON dict_rate_limit_policies(endpoint_path);

-- Trigger to auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_dict_rate_limit_policies_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW() AT TIME ZONE 'UTC';
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_dict_rate_limit_policies_updated_at
    BEFORE UPDATE ON dict_rate_limit_policies
    FOR EACH ROW
    EXECUTE FUNCTION update_dict_rate_limit_policies_updated_at();

-- Table comments for documentation
COMMENT ON TABLE dict_rate_limit_policies IS 'DICT API rate limit policy configurations retrieved from BACEN';
COMMENT ON COLUMN dict_rate_limit_policies.endpoint_id IS 'Unique identifier for the rate-limited endpoint';
COMMENT ON COLUMN dict_rate_limit_policies.capacity IS 'Maximum number of tokens in the bucket';
COMMENT ON COLUMN dict_rate_limit_policies.refill_tokens IS 'Number of tokens added per refill period';
COMMENT ON COLUMN dict_rate_limit_policies.refill_period_sec IS 'Refill period in seconds';
COMMENT ON COLUMN dict_rate_limit_policies.psp_category IS 'PSP category (A-H) if endpoint-specific limits exist';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trigger_update_dict_rate_limit_policies_updated_at ON dict_rate_limit_policies;
DROP FUNCTION IF EXISTS update_dict_rate_limit_policies_updated_at();
DROP TABLE IF EXISTS dict_rate_limit_policies;
-- +goose StatementEnd
