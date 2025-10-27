-- Migration: 003_create_claims_table
-- Description: Create claims and portabilities tables
-- Author: data-specialist-core
-- Date: 2025-10-27

-- +goose Up
-- +goose StatementBegin

-- Create claims table
CREATE TABLE core_dict.claims (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id             VARCHAR(100) UNIQUE,
    workflow_id             VARCHAR(255),
    entry_id                UUID NOT NULL REFERENCES core_dict.dict_entries(id) ON DELETE RESTRICT,
    claim_type              VARCHAR(50) NOT NULL CHECK (claim_type IN ('OWNERSHIP', 'PORTABILITY')),
    claimer_ispb            VARCHAR(8) NOT NULL,
    claimer_account_id      UUID REFERENCES core_dict.accounts(id),
    owner_ispb              VARCHAR(8) NOT NULL,
    owner_account_id        UUID NOT NULL REFERENCES core_dict.accounts(id),
    status                  VARCHAR(50) NOT NULL DEFAULT 'OPEN' CHECK (
                                status IN ('OPEN', 'WAITING_RESOLUTION', 'CONFIRMED',
                                         'CANCELLED', 'COMPLETED', 'EXPIRED')
                            ),
    completion_period_days  INT NOT NULL DEFAULT 30,
    expires_at              TIMESTAMP WITH TIME ZONE NOT NULL,
    resolution_type         VARCHAR(50) CHECK (
                                resolution_type IN ('APPROVED', 'REJECTED', 'TIMEOUT', 'CANCELLED')
                            ),
    resolution_reason       TEXT,
    resolution_date         TIMESTAMP WITH TIME ZONE,
    created_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMP WITH TIME ZONE,
    last_sync_at            TIMESTAMP WITH TIME ZONE,
    sync_status             VARCHAR(20),
    created_by              UUID REFERENCES core_dict.users(id),
    updated_by              UUID REFERENCES core_dict.users(id),

    -- Constraints
    CHECK (expires_at > created_at),
    CHECK (completion_period_days > 0)
);

COMMENT ON TABLE core_dict.claims IS 'PIX key claims (30-day resolution period per TEC-003 v2.1)';
COMMENT ON COLUMN core_dict.claims.completion_period_days IS 'Resolution period in days (default: 30 days per TEC-003 v2.1)';
COMMENT ON COLUMN core_dict.claims.workflow_id IS 'Temporal Workflow ID (ClaimWorkflow in RSFN Connect)';

-- Create portabilities table
CREATE TABLE core_dict.portabilities (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id             VARCHAR(100) UNIQUE,
    workflow_id             VARCHAR(255),
    entry_id                UUID NOT NULL REFERENCES core_dict.dict_entries(id) ON DELETE RESTRICT,
    origin_ispb             VARCHAR(8) NOT NULL,
    origin_account_id       UUID NOT NULL REFERENCES core_dict.accounts(id),
    destination_ispb        VARCHAR(8) NOT NULL,
    destination_account_id  UUID NOT NULL REFERENCES core_dict.accounts(id),
    status                  VARCHAR(50) NOT NULL DEFAULT 'INITIATED' CHECK (
                                status IN ('INITIATED', 'PENDING_APPROVAL', 'APPROVED',
                                         'REJECTED', 'COMPLETED', 'CANCELLED', 'FAILED')
                            ),
    initiated_at            TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at            TIMESTAMP WITH TIME ZONE,
    requires_otp            BOOLEAN NOT NULL DEFAULT TRUE,
    otp_validated_at        TIMESTAMP WITH TIME ZONE,
    created_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by              UUID REFERENCES core_dict.users(id)
);

COMMENT ON TABLE core_dict.portabilities IS 'PIX key portabilities between institutions';

-- Add foreign keys to dict_entries (deferred because of circular dependency)
ALTER TABLE core_dict.dict_entries
    ADD CONSTRAINT fk_claim_id FOREIGN KEY (claim_id)
        REFERENCES core_dict.claims(id) ON DELETE SET NULL;

ALTER TABLE core_dict.dict_entries
    ADD CONSTRAINT fk_portability_id FOREIGN KEY (portability_id)
        REFERENCES core_dict.portabilities(id) ON DELETE SET NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop foreign keys from dict_entries first
ALTER TABLE core_dict.dict_entries DROP CONSTRAINT IF EXISTS fk_claim_id;
ALTER TABLE core_dict.dict_entries DROP CONSTRAINT IF EXISTS fk_portability_id;

DROP TABLE IF EXISTS core_dict.portabilities CASCADE;
DROP TABLE IF EXISTS core_dict.claims CASCADE;

-- +goose StatementEnd
