-- +goose Up
-- +goose StatementBegin
-- Entries table: stores DICT key entries (PIX keys)
CREATE TABLE entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entry_id VARCHAR(50) UNIQUE NOT NULL,

    -- Key information
    key VARCHAR(255) UNIQUE NOT NULL,
    key_type VARCHAR(20) NOT NULL CHECK (key_type IN ('CPF', 'CNPJ', 'EMAIL', 'PHONE', 'EVP')),

    -- Account information
    participant VARCHAR(8) NOT NULL,  -- Current owner ISPB
    account_branch VARCHAR(10),
    account_number VARCHAR(20),
    account_type VARCHAR(20) CHECK (account_type IN ('CACC', 'SLRY', 'SVGS', 'TRAN')),
    account_opened_date DATE,

    -- Owner information (person or company)
    owner_type VARCHAR(20) CHECK (owner_type IN ('NATURAL_PERSON', 'LEGAL_PERSON')),
    owner_name VARCHAR(255),
    owner_tax_id VARCHAR(14),  -- CPF (11) or CNPJ (14)

    -- Status management
    status VARCHAR(30) NOT NULL CHECK (status IN (
        'ACTIVE',                  -- Entry active and operational
        'INACTIVE',                -- Entry deactivated
        'BLOCKED',                 -- Entry temporarily blocked
        'PORTABILITY_PENDING',     -- Portability claim in progress
        'OWNERSHIP_CHANGE_PENDING' -- Ownership claim in progress
    )) DEFAULT 'ACTIVE',

    -- Timestamps
    registered_at TIMESTAMPTZ,  -- When key was first registered
    activated_at TIMESTAMPTZ,   -- When key was activated
    deactivated_at TIMESTAMPTZ, -- When key was deactivated

    -- Metadata
    reason_for_status_change TEXT,
    bacen_entry_id VARCHAR(50),  -- BACEN's internal entry ID

    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,  -- Soft delete

    -- Constraints
    CONSTRAINT valid_participant CHECK (LENGTH(participant) = 8),
    CONSTRAINT valid_tax_id CHECK (LENGTH(owner_tax_id) IN (11, 14))
);

-- Indexes for performance
CREATE UNIQUE INDEX idx_entries_key ON entries(key) WHERE deleted_at IS NULL;
CREATE INDEX idx_entries_entry_id ON entries(entry_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_entries_participant ON entries(participant) WHERE deleted_at IS NULL;
CREATE INDEX idx_entries_status ON entries(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_entries_key_type ON entries(key_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_entries_owner_tax_id ON entries(owner_tax_id) WHERE deleted_at IS NULL;

-- Composite indexes
CREATE INDEX idx_entries_participant_status ON entries(participant, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_entries_key_status ON entries(key, status) WHERE deleted_at IS NULL;

-- Trigger for updated_at
CREATE TRIGGER update_entries_updated_at
    BEFORE UPDATE ON entries
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE entries IS 'Stores DICT key entries (PIX keys) registered in the system';
COMMENT ON COLUMN entries.key IS 'The PIX key (CPF, CNPJ, email, phone, or random key)';
COMMENT ON COLUMN entries.participant IS 'ISPB of the institution that owns this entry';
COMMENT ON COLUMN entries.status IS 'Current status of the entry';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_entries_updated_at ON entries;
DROP TABLE IF EXISTS entries;
-- +goose StatementEnd