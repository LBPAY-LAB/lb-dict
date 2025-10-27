-- Migration: 20251026100001_create_dict_entries
-- Description: Create dict_entries table for storing PIX keys

CREATE TABLE IF NOT EXISTS dict_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Key information
    key_type VARCHAR(20) NOT NULL CHECK (key_type IN ('CPF', 'CNPJ', 'EMAIL', 'PHONE', 'EVP')),
    key_value VARCHAR(77) NOT NULL,

    -- Account information
    ispb CHAR(8) NOT NULL,
    account_number VARCHAR(20) NOT NULL,
    account_type VARCHAR(10) NOT NULL CHECK (account_type IN ('CACC', 'SLRY', 'SVGS', 'TRAN')),
    branch VARCHAR(10),

    -- Owner information
    owner_type VARCHAR(10) NOT NULL CHECK (owner_type IN ('NATURAL', 'LEGAL')),
    owner_document VARCHAR(14) NOT NULL,
    owner_name VARCHAR(200) NOT NULL,

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'PORTABILITY_PENDING', 'BLOCKED', 'DELETED')),

    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100),

    -- Metadata
    reason VARCHAR(500),
    external_id VARCHAR(100),

    -- Unique constraint
    CONSTRAINT uk_dict_entries_key UNIQUE (key_type, key_value)
);

-- Indexes for performance
CREATE INDEX idx_dict_entries_ispb ON dict_entries(ispb);
CREATE INDEX idx_dict_entries_owner_document ON dict_entries(owner_document);
CREATE INDEX idx_dict_entries_status ON dict_entries(status);
CREATE INDEX idx_dict_entries_created_at ON dict_entries(created_at DESC);
CREATE INDEX idx_dict_entries_key_value ON dict_entries(key_value);

-- Composite index for common queries
CREATE INDEX idx_dict_entries_ispb_status ON dict_entries(ispb, status);

-- Trigger to update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_dict_entries_updated_at
BEFORE UPDATE ON dict_entries
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE dict_entries IS 'Stores PIX dictionary entries (keys)';
COMMENT ON COLUMN dict_entries.key_type IS 'Type of PIX key: CPF, CNPJ, EMAIL, PHONE, EVP';
COMMENT ON COLUMN dict_entries.key_value IS 'The actual key value (hashed for EVP)';
COMMENT ON COLUMN dict_entries.ispb IS 'Bank identifier (ISPB)';
COMMENT ON COLUMN dict_entries.status IS 'Entry status: ACTIVE, PORTABILITY_PENDING, BLOCKED, DELETED';