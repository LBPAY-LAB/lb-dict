-- Migration: 002_create_entries_table
-- Description: Create dict_entries table (PIX keys) with partitioning by ISPB
-- Author: data-specialist-core
-- Date: 2025-10-27

-- +goose Up
-- +goose StatementBegin

-- Create users table first (referenced by entries)
CREATE TABLE core_dict.users (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username            VARCHAR(100) UNIQUE NOT NULL,
    email               VARCHAR(255) UNIQUE NOT NULL,
    password_hash       VARCHAR(255),
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    is_admin            BOOLEAN NOT NULL DEFAULT FALSE,
    first_name          VARCHAR(100),
    last_name           VARCHAR(100),
    role                VARCHAR(50) CHECK (role IN ('ADMIN', 'OPERATOR', 'VIEWER', 'AUDITOR')),
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_login_at       TIMESTAMP WITH TIME ZONE,
    sso_provider        VARCHAR(50),
    sso_external_id     VARCHAR(255)
);

COMMENT ON TABLE core_dict.users IS 'Users authorized to operate Core DICT';

-- Create accounts table (referenced by entries)
CREATE TABLE core_dict.accounts (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id             VARCHAR(100) UNIQUE,
    account_number          VARCHAR(20) NOT NULL,
    branch_code             VARCHAR(10) NOT NULL,
    account_type            VARCHAR(20) NOT NULL CHECK (
                                account_type IN ('CACC', 'SVGS', 'SLRY', 'TRAN')
                            ),
    account_status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (
                                account_status IN ('ACTIVE', 'BLOCKED', 'CLOSED', 'PENDING_CLOSURE')
                            ),
    holder_document         VARCHAR(14) NOT NULL,
    holder_document_type    VARCHAR(10) NOT NULL CHECK (holder_document_type IN ('CPF', 'CNPJ')),
    holder_name             VARCHAR(255) NOT NULL,
    holder_name_encrypted   BYTEA,
    participant_ispb        VARCHAR(8) NOT NULL,
    opened_at               TIMESTAMP WITH TIME ZONE NOT NULL,
    closed_at               TIMESTAMP WITH TIME ZONE,
    created_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMP WITH TIME ZONE,
    last_sync_at            TIMESTAMP WITH TIME ZONE,
    sync_source             VARCHAR(50),
    created_by              UUID REFERENCES core_dict.users(id),
    updated_by              UUID REFERENCES core_dict.users(id),

    -- Constraints
    CONSTRAINT chk_holder_document_format CHECK (
        (holder_document_type = 'CPF' AND LENGTH(holder_document) = 11) OR
        (holder_document_type = 'CNPJ' AND LENGTH(holder_document) = 14)
    ),
    CONSTRAINT unique_account UNIQUE (participant_ispb, branch_code, account_number, deleted_at)
);

COMMENT ON TABLE core_dict.accounts IS 'CID accounts (Conta de Identificação de Depósito) linked to PIX keys';
COMMENT ON COLUMN core_dict.accounts.account_type IS 'CACC=Checking, SVGS=Savings, SLRY=Salary, TRAN=Transactional';

-- Create dict_entries table (PIX keys)
CREATE TABLE core_dict.dict_entries (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id             VARCHAR(100) UNIQUE,
    key_type                VARCHAR(20) NOT NULL CHECK (
                                key_type IN ('CPF', 'CNPJ', 'EMAIL', 'PHONE', 'EVP')
                            ),
    key_value               VARCHAR(255) NOT NULL,
    key_hash                VARCHAR(64) NOT NULL,
    account_id              UUID NOT NULL REFERENCES core_dict.accounts(id) ON DELETE RESTRICT,
    participant_ispb        VARCHAR(8) NOT NULL,
    participant_branch      VARCHAR(10),
    status                  VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (
                                status IN ('PENDING', 'ACTIVE', 'PORTABILITY_REQUESTED',
                                         'OWNERSHIP_CONFIRMED', 'DELETED', 'CLAIM_PENDING')
                            ),
    ownership_type          VARCHAR(20) NOT NULL CHECK (
                                ownership_type IN ('NATURAL_PERSON', 'LEGAL_ENTITY')
                            ),
    claim_id                UUID,
    portability_id          UUID,
    created_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMP WITH TIME ZONE,
    last_sync_at            TIMESTAMP WITH TIME ZONE,
    sync_status             VARCHAR(20) CHECK (
                                sync_status IN ('SYNCED', 'PENDING_SYNC', 'SYNC_ERROR', 'NOT_SYNCED')
                            ),
    sync_error_message      TEXT,
    created_by              UUID REFERENCES core_dict.users(id),
    updated_by              UUID REFERENCES core_dict.users(id),

    -- Constraints
    CONSTRAINT chk_participant_ispb_format CHECK (
        LENGTH(participant_ispb) = 8 AND participant_ispb ~ '^[0-9]+$'
    ),
    CONSTRAINT chk_key_value_format CHECK (
        (key_type = 'CPF' AND LENGTH(key_value) = 11) OR
        (key_type = 'CNPJ' AND LENGTH(key_value) = 14) OR
        (key_type = 'EMAIL' AND key_value ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$') OR
        (key_type = 'PHONE' AND key_value ~ '^\+55[1-9]{2}9?[0-9]{8}$') OR
        (key_type = 'EVP' AND LENGTH(key_value) = 36)
    ),
    CONSTRAINT unique_active_key UNIQUE NULLS NOT DISTINCT (key_type, key_value, deleted_at),
    CONSTRAINT unique_key_hash UNIQUE (key_hash)
);

COMMENT ON TABLE core_dict.dict_entries IS 'PIX keys managed by Core DICT';
COMMENT ON COLUMN core_dict.dict_entries.key_hash IS 'SHA-256 hash of key_value for LGPD-compliant searches';
COMMENT ON COLUMN core_dict.dict_entries.external_id IS 'ID returned by Bacen DICT after creation';

-- Enable Row Level Security
ALTER TABLE core_dict.dict_entries ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Users can only see entries from their ISPB
CREATE POLICY entries_tenant_isolation ON core_dict.dict_entries
    FOR ALL
    TO PUBLIC
    USING (participant_ispb = COALESCE(current_setting('app.current_ispb', true), participant_ispb));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS core_dict.dict_entries CASCADE;
DROP TABLE IF EXISTS core_dict.accounts CASCADE;
DROP TABLE IF EXISTS core_dict.users CASCADE;

-- +goose StatementEnd
