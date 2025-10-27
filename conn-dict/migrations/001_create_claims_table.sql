-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Claims table: stores DICT portability and ownership claims
CREATE TABLE claims (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    claim_id VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('PORTABILITY', 'OWNERSHIP')),

    -- Key information
    key VARCHAR(255) NOT NULL,
    key_type VARCHAR(20) NOT NULL CHECK (key_type IN ('CPF', 'CNPJ', 'EMAIL', 'PHONE', 'EVP')),

    -- Status management
    status VARCHAR(30) NOT NULL CHECK (status IN (
        'OPEN',           -- Claim created, awaiting donor response
        'WAITING_RESOLUTION', -- Donor has 7 days to confirm/deny
        'CONFIRMED',      -- Donor confirmed, waiting completion
        'CANCELLED',      -- Claim cancelled by claimer or donor
        'COMPLETED',      -- Claim completed, ownership transferred
        'EXPIRED'         -- Claim expired after 30 days
    )) DEFAULT 'OPEN',

    -- Participants
    donor_participant VARCHAR(8) NOT NULL,      -- Donor ISPB (8 digits)
    claimer_participant VARCHAR(8) NOT NULL,    -- Claimer ISPB (8 digits)

    -- Account information (claimer's account to receive the key)
    claimer_account_branch VARCHAR(10),
    claimer_account_number VARCHAR(20),
    claimer_account_type VARCHAR(20) CHECK (claimer_account_type IN ('CACC', 'SLRY', 'SVGS', 'TRAN')),

    -- Timestamps
    completion_period_end TIMESTAMPTZ,  -- 7 days from creation for donor response
    claim_expiry_date TIMESTAMPTZ,      -- 30 days from creation
    confirmed_at TIMESTAMPTZ,           -- When donor confirmed
    completed_at TIMESTAMPTZ,           -- When portability completed
    cancelled_at TIMESTAMPTZ,           -- When claim was cancelled
    expired_at TIMESTAMPTZ,             -- When claim expired

    -- Metadata
    cancellation_reason TEXT,
    notes TEXT,

    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,  -- Soft delete

    -- Indexes
    CONSTRAINT fk_donor CHECK (LENGTH(donor_participant) = 8),
    CONSTRAINT fk_claimer CHECK (LENGTH(claimer_participant) = 8),
    CONSTRAINT different_participants CHECK (donor_participant != claimer_participant)
);

-- Indexes for performance
CREATE INDEX idx_claims_claim_id ON claims(claim_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_claims_key ON claims(key) WHERE deleted_at IS NULL;
CREATE INDEX idx_claims_status ON claims(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_claims_donor ON claims(donor_participant) WHERE deleted_at IS NULL;
CREATE INDEX idx_claims_claimer ON claims(claimer_participant) WHERE deleted_at IS NULL;
CREATE INDEX idx_claims_created_at ON claims(created_at DESC);
CREATE INDEX idx_claims_expiry ON claims(claim_expiry_date) WHERE status IN ('OPEN', 'WAITING_RESOLUTION') AND deleted_at IS NULL;

-- Composite indexes for common queries
CREATE INDEX idx_claims_key_status ON claims(key, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_claims_participant_status ON claims(donor_participant, status) WHERE deleted_at IS NULL;

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_claims_updated_at
    BEFORE UPDATE ON claims
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments for documentation
COMMENT ON TABLE claims IS 'Stores DICT key portability and ownership claims (BACEN DICT requirement)';
COMMENT ON COLUMN claims.claim_id IS 'Unique claim identifier from BACEN DICT';
COMMENT ON COLUMN claims.type IS 'Claim type: PORTABILITY (change ISPB) or OWNERSHIP (change account)';
COMMENT ON COLUMN claims.completion_period_end IS 'Donor has 7 days to respond';
COMMENT ON COLUMN claims.claim_expiry_date IS 'Claim expires after 30 days if not completed';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_claims_updated_at ON claims;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS claims;
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd