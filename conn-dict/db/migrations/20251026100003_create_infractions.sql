-- Migration: 20251026100003_create_infractions
-- Description: Create infractions table for DICT rule violations

CREATE TABLE IF NOT EXISTS infractions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Reference to entry (optional, can be NULL for general infractions)
    entry_id UUID REFERENCES dict_entries(id) ON DELETE SET NULL,

    -- Infraction information
    infraction_type VARCHAR(50) NOT NULL CHECK (infraction_type IN (
        'FRAUD',
        'ACCOUNT_CLOSED',
        'INCORRECT_DATA',
        'DUPLICATE_KEY',
        'UNAUTHORIZED_USE',
        'OTHER'
    )),

    -- Reporter information
    reporter_ispb CHAR(8) NOT NULL,
    reported_ispb CHAR(8) NOT NULL,

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (status IN (
        'OPEN',              -- Infraction reported
        'ACKNOWLEDGED',      -- Acknowledged by reported party
        'UNDER_REVIEW',      -- Being reviewed by Bacen
        'RESOLVED',          -- Resolved
        'REJECTED'           -- Rejected (invalid infraction)
    )),

    -- Details
    description TEXT NOT NULL,
    evidence_url VARCHAR(500),
    bacen_protocol VARCHAR(100),

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,

    -- Resolution
    resolution TEXT,
    resolved_by VARCHAR(100),

    -- Metadata
    external_id VARCHAR(100),
    severity VARCHAR(20) CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL'))
);

-- Indexes
CREATE INDEX idx_infractions_entry_id ON infractions(entry_id) WHERE entry_id IS NOT NULL;
CREATE INDEX idx_infractions_reporter_ispb ON infractions(reporter_ispb);
CREATE INDEX idx_infractions_reported_ispb ON infractions(reported_ispb);
CREATE INDEX idx_infractions_status ON infractions(status);
CREATE INDEX idx_infractions_created_at ON infractions(created_at DESC);
CREATE INDEX idx_infractions_type ON infractions(infraction_type);

-- Composite index
CREATE INDEX idx_infractions_reported_status ON infractions(reported_ispb, status);

-- Comments
COMMENT ON TABLE infractions IS 'Stores DICT infractions and rule violations';
COMMENT ON COLUMN infractions.infraction_type IS 'Type of infraction reported';
COMMENT ON COLUMN infractions.bacen_protocol IS 'Bacen protocol number for tracking';