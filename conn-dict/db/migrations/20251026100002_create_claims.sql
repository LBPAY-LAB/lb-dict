-- Migration: 20251026100002_create_claims
-- Description: Create claims table for PIX key portability

CREATE TABLE IF NOT EXISTS claims (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Reference to entry
    entry_id UUID NOT NULL REFERENCES dict_entries(id) ON DELETE CASCADE,

    -- Claim information
    claim_type VARCHAR(20) NOT NULL CHECK (claim_type IN ('OWNERSHIP', 'PORTABILITY')),
    claimer_ispb CHAR(8) NOT NULL,
    claimer_account VARCHAR(20) NOT NULL,
    claimer_branch VARCHAR(10),

    -- Donor information (current owner)
    donor_ispb CHAR(8) NOT NULL,

    -- Status
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING' CHECK (status IN (
        'PENDING',           -- Waiting for donor response
        'CONFIRMED',         -- Donor confirmed
        'CANCELLED',         -- Cancelled by claimer or donor
        'COMPLETED',         -- Claim completed successfully
        'EXPIRED'            -- 30 days timeout
    )),

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() + INTERVAL '30 days',

    -- Audit
    created_by VARCHAR(100),
    cancellation_reason VARCHAR(500),

    -- Metadata
    external_id VARCHAR(100),
    temporal_workflow_id VARCHAR(200),  -- Temporal workflow ID for tracking

    -- Constraints
    CONSTRAINT chk_claims_dates CHECK (
        (status = 'CONFIRMED' AND confirmed_at IS NOT NULL) OR
        (status != 'CONFIRMED') AND
        (status = 'CANCELLED' AND cancelled_at IS NOT NULL) OR
        (status != 'CANCELLED') AND
        (status = 'COMPLETED' AND completed_at IS NOT NULL) OR
        (status != 'COMPLETED')
    )
);

-- Indexes
CREATE INDEX idx_claims_entry_id ON claims(entry_id);
CREATE INDEX idx_claims_status ON claims(status);
CREATE INDEX idx_claims_claimer_ispb ON claims(claimer_ispb);
CREATE INDEX idx_claims_donor_ispb ON claims(donor_ispb);
CREATE INDEX idx_claims_created_at ON claims(created_at DESC);
CREATE INDEX idx_claims_expires_at ON claims(expires_at) WHERE status = 'PENDING';
CREATE INDEX idx_claims_temporal_workflow_id ON claims(temporal_workflow_id);

-- Composite index for common queries
CREATE INDEX idx_claims_entry_status ON claims(entry_id, status);

-- Comments
COMMENT ON TABLE claims IS 'Stores portability claims (30-day workflow)';
COMMENT ON COLUMN claims.claim_type IS 'Type of claim: OWNERSHIP or PORTABILITY';
COMMENT ON COLUMN claims.status IS 'Claim lifecycle status';
COMMENT ON COLUMN claims.expires_at IS 'Claim expiration date (30 days from creation)';
COMMENT ON COLUMN claims.temporal_workflow_id IS 'Temporal workflow ID for tracking the 30-day claim process';