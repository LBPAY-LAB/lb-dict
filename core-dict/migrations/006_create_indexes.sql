-- Migration: 006_create_indexes
-- Description: Create optimized indexes for performance
-- Author: data-specialist-core
-- Date: 2025-10-27

-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- DICT_ENTRIES INDEXES
-- ============================================================================

-- Primary lookup: by key type and value
CREATE INDEX idx_entries_key_type_value
    ON core_dict.dict_entries (key_type, key_value)
    WHERE deleted_at IS NULL;

-- Lookup by hash (LGPD-compliant search)
CREATE INDEX idx_entries_key_hash
    ON core_dict.dict_entries (key_hash)
    WHERE deleted_at IS NULL;

-- Lookup by account
CREATE INDEX idx_entries_account_id
    ON core_dict.dict_entries (account_id)
    WHERE deleted_at IS NULL;

-- Status filtering
CREATE INDEX idx_entries_status
    ON core_dict.dict_entries (status)
    WHERE deleted_at IS NULL;

-- Sync status (for background sync jobs)
CREATE INDEX idx_entries_sync_status
    ON core_dict.dict_entries (sync_status, last_sync_at)
    WHERE sync_status != 'SYNCED' AND deleted_at IS NULL;

-- ISPB filtering (multi-tenant queries)
CREATE INDEX idx_entries_participant_ispb
    ON core_dict.dict_entries (participant_ispb, status)
    WHERE deleted_at IS NULL;

-- External ID lookup
CREATE INDEX idx_entries_external_id
    ON core_dict.dict_entries (external_id)
    WHERE external_id IS NOT NULL;

-- Claim/portability lookup
CREATE INDEX idx_entries_claim_id
    ON core_dict.dict_entries (claim_id)
    WHERE claim_id IS NOT NULL;

CREATE INDEX idx_entries_portability_id
    ON core_dict.dict_entries (portability_id)
    WHERE portability_id IS NOT NULL;

-- ============================================================================
-- ACCOUNTS INDEXES
-- ============================================================================

-- Lookup by holder document (CPF/CNPJ)
CREATE INDEX idx_accounts_holder_document
    ON core_dict.accounts (holder_document_type, holder_document)
    WHERE deleted_at IS NULL;

-- Lookup by participant and status
CREATE INDEX idx_accounts_participant_status
    ON core_dict.accounts (participant_ispb, account_status)
    WHERE deleted_at IS NULL;

-- External ID lookup
CREATE INDEX idx_accounts_external_id
    ON core_dict.accounts (external_id)
    WHERE external_id IS NOT NULL;

-- Full-text search on holder name (pg_trgm)
CREATE INDEX idx_accounts_holder_name_trgm
    ON core_dict.accounts USING gin (holder_name gin_trgm_ops);

-- Account number lookup
CREATE INDEX idx_accounts_number_branch
    ON core_dict.accounts (participant_ispb, branch_code, account_number)
    WHERE deleted_at IS NULL;

-- ============================================================================
-- CLAIMS INDEXES
-- ============================================================================

-- Active claims lookup
CREATE INDEX idx_claims_status
    ON core_dict.claims (status, created_at DESC)
    WHERE status IN ('OPEN', 'WAITING_RESOLUTION');

-- Expiration check (for cron job)
CREATE INDEX idx_claims_expires_at
    ON core_dict.claims (expires_at, status)
    WHERE status = 'OPEN';

-- Workflow ID lookup
CREATE INDEX idx_claims_workflow_id
    ON core_dict.claims (workflow_id)
    WHERE workflow_id IS NOT NULL;

-- Entry lookup
CREATE INDEX idx_claims_entry_id
    ON core_dict.claims (entry_id);

-- ISPB filtering
CREATE INDEX idx_claims_claimer_ispb
    ON core_dict.claims (claimer_ispb, status);

CREATE INDEX idx_claims_owner_ispb
    ON core_dict.claims (owner_ispb, status);

-- ============================================================================
-- PORTABILITIES INDEXES
-- ============================================================================

-- Status lookup
CREATE INDEX idx_portabilities_status
    ON core_dict.portabilities (status, created_at DESC);

-- Workflow ID lookup
CREATE INDEX idx_portabilities_workflow_id
    ON core_dict.portabilities (workflow_id)
    WHERE workflow_id IS NOT NULL;

-- Entry lookup
CREATE INDEX idx_portabilities_entry_id
    ON core_dict.portabilities (entry_id);

-- ISPB filtering
CREATE INDEX idx_portabilities_origin_ispb
    ON core_dict.portabilities (origin_ispb, status);

CREATE INDEX idx_portabilities_destination_ispb
    ON core_dict.portabilities (destination_ispb, status);

-- ============================================================================
-- USERS INDEXES
-- ============================================================================

-- Login lookup
CREATE INDEX idx_users_username
    ON core_dict.users (username)
    WHERE is_active = TRUE;

CREATE INDEX idx_users_email
    ON core_dict.users (email)
    WHERE is_active = TRUE;

-- SSO lookup
CREATE INDEX idx_users_sso
    ON core_dict.users (sso_provider, sso_external_id)
    WHERE sso_provider IS NOT NULL;

-- ============================================================================
-- AUDIT.ENTRY_EVENTS INDEXES
-- ============================================================================

-- Time-based lookup (most common query pattern)
CREATE INDEX idx_entry_events_occurred_at
    ON audit.entry_events (occurred_at DESC);

-- Entity lookup
CREATE INDEX idx_entry_events_entity
    ON audit.entry_events (entity_type, entity_id, occurred_at DESC);

-- User activity
CREATE INDEX idx_entry_events_user
    ON audit.entry_events (user_id, occurred_at DESC)
    WHERE user_id IS NOT NULL;

-- Event type filtering
CREATE INDEX idx_entry_events_event_type
    ON audit.entry_events (event_type, occurred_at DESC);

-- JSONB indexes for metadata search
CREATE INDEX idx_entry_events_metadata
    ON audit.entry_events USING gin (metadata);

CREATE INDEX idx_entry_events_new_values
    ON audit.entry_events USING gin (new_values);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop all indexes (in reverse order)
DROP INDEX IF EXISTS audit.idx_entry_events_new_values;
DROP INDEX IF EXISTS audit.idx_entry_events_metadata;
DROP INDEX IF EXISTS audit.idx_entry_events_event_type;
DROP INDEX IF EXISTS audit.idx_entry_events_user;
DROP INDEX IF EXISTS audit.idx_entry_events_entity;
DROP INDEX IF EXISTS audit.idx_entry_events_occurred_at;

DROP INDEX IF EXISTS core_dict.idx_users_sso;
DROP INDEX IF EXISTS core_dict.idx_users_email;
DROP INDEX IF EXISTS core_dict.idx_users_username;

DROP INDEX IF EXISTS core_dict.idx_portabilities_destination_ispb;
DROP INDEX IF EXISTS core_dict.idx_portabilities_origin_ispb;
DROP INDEX IF EXISTS core_dict.idx_portabilities_entry_id;
DROP INDEX IF EXISTS core_dict.idx_portabilities_workflow_id;
DROP INDEX IF EXISTS core_dict.idx_portabilities_status;

DROP INDEX IF EXISTS core_dict.idx_claims_owner_ispb;
DROP INDEX IF EXISTS core_dict.idx_claims_claimer_ispb;
DROP INDEX IF EXISTS core_dict.idx_claims_entry_id;
DROP INDEX IF EXISTS core_dict.idx_claims_workflow_id;
DROP INDEX IF EXISTS core_dict.idx_claims_expires_at;
DROP INDEX IF EXISTS core_dict.idx_claims_status;

DROP INDEX IF EXISTS core_dict.idx_accounts_number_branch;
DROP INDEX IF EXISTS core_dict.idx_accounts_holder_name_trgm;
DROP INDEX IF EXISTS core_dict.idx_accounts_external_id;
DROP INDEX IF EXISTS core_dict.idx_accounts_participant_status;
DROP INDEX IF EXISTS core_dict.idx_accounts_holder_document;

DROP INDEX IF EXISTS core_dict.idx_entries_portability_id;
DROP INDEX IF EXISTS core_dict.idx_entries_claim_id;
DROP INDEX IF EXISTS core_dict.idx_entries_external_id;
DROP INDEX IF EXISTS core_dict.idx_entries_participant_ispb;
DROP INDEX IF EXISTS core_dict.idx_entries_sync_status;
DROP INDEX IF EXISTS core_dict.idx_entries_status;
DROP INDEX IF EXISTS core_dict.idx_entries_account_id;
DROP INDEX IF EXISTS core_dict.idx_entries_key_hash;
DROP INDEX IF EXISTS core_dict.idx_entries_key_type_value;

-- +goose StatementEnd
