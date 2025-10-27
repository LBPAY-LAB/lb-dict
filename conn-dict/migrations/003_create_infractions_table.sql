-- +goose Up
-- +goose StatementBegin
-- Infractions table: stores fraud reports and infractions
CREATE TABLE infractions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    infraction_id VARCHAR(50) UNIQUE NOT NULL,

    -- Related entry/claim
    entry_id VARCHAR(50),
    claim_id VARCHAR(50),
    key VARCHAR(255) NOT NULL,

    -- Infraction details
    type VARCHAR(50) NOT NULL CHECK (type IN (
        'FRAUD',
        'ACCOUNT_CLOSED',
        'INCORRECT_DATA',
        'UNAUTHORIZED_USE',
        'DUPLICATE_KEY',
        'OTHER'
    )),

    -- Reporter information
    reporter_participant VARCHAR(8) NOT NULL,  -- ISPB that reported
    reported_participant VARCHAR(8),            -- ISPB being reported (if applicable)

    -- Status
    status VARCHAR(30) NOT NULL CHECK (status IN (
        'OPEN',
        'UNDER_INVESTIGATION',
        'RESOLVED',
        'DISMISSED',
        'ESCALATED_TO_BACEN'
    )) DEFAULT 'OPEN',

    -- Details
    description TEXT NOT NULL,
    evidence_urls TEXT[],  -- Array of evidence URLs
    resolution_notes TEXT,

    -- Timestamps
    reported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    investigated_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,

    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- Foreign keys
    CONSTRAINT fk_entry FOREIGN KEY (entry_id) REFERENCES entries(entry_id) ON DELETE SET NULL,
    CONSTRAINT valid_reporter CHECK (LENGTH(reporter_participant) = 8)
);

-- Indexes
CREATE INDEX idx_infractions_infraction_id ON infractions(infraction_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_infractions_key ON infractions(key) WHERE deleted_at IS NULL;
CREATE INDEX idx_infractions_status ON infractions(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_infractions_reporter ON infractions(reporter_participant) WHERE deleted_at IS NULL;
CREATE INDEX idx_infractions_type ON infractions(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_infractions_reported_at ON infractions(reported_at DESC);

-- Trigger
CREATE TRIGGER update_infractions_updated_at
    BEFORE UPDATE ON infractions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE infractions IS 'Stores fraud reports and infractions related to DICT entries';
COMMENT ON COLUMN infractions.evidence_urls IS 'Array of URLs pointing to evidence documents';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_infractions_updated_at ON infractions;
DROP TABLE IF EXISTS infractions;
-- +goose StatementEnd