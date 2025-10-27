-- Migration: 005_create_triggers
-- Description: Create triggers for updated_at and audit logging
-- Author: data-specialist-core
-- Date: 2025-10-27

-- +goose Up
-- +goose StatementBegin

-- Function: Automatically update updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function: Audit entry changes
CREATE OR REPLACE FUNCTION audit_entry_changes()
RETURNS TRIGGER AS $$
DECLARE
    event_type_val VARCHAR(100);
BEGIN
    -- Determine event type
    IF TG_OP = 'INSERT' THEN
        event_type_val := 'CREATED';
    ELSIF TG_OP = 'UPDATE' THEN
        event_type_val := 'UPDATED';
    ELSIF TG_OP = 'DELETE' THEN
        event_type_val := 'DELETED';
    END IF;

    -- Insert audit event
    INSERT INTO audit.entry_events (
        entity_type,
        entity_id,
        event_type,
        old_values,
        new_values,
        user_id,
        occurred_at
    ) VALUES (
        TG_TABLE_NAME,
        COALESCE(NEW.id, OLD.id),
        event_type_val,
        CASE WHEN TG_OP != 'INSERT' THEN row_to_json(OLD) ELSE NULL END,
        CASE WHEN TG_OP != 'DELETE' THEN row_to_json(NEW) ELSE NULL END,
        COALESCE(NEW.updated_by, NEW.created_by, OLD.updated_by),
        NOW()
    );

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Function: Expire old claims (to be called by cron)
CREATE OR REPLACE FUNCTION expire_old_claims()
RETURNS TABLE (expired_count INT) AS $$
DECLARE
    count INT;
BEGIN
    UPDATE core_dict.claims
    SET
        status = 'EXPIRED',
        resolution_type = 'TIMEOUT',
        resolution_date = NOW(),
        updated_at = NOW()
    WHERE
        status = 'OPEN'
        AND expires_at < NOW();

    GET DIAGNOSTICS count = ROW_COUNT;
    RETURN QUERY SELECT count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION expire_old_claims IS 'Expire claims with expires_at < NOW() (run daily via cron)';

-- Apply updated_at trigger to all main tables
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON core_dict.users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_accounts_updated_at
    BEFORE UPDATE ON core_dict.accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_entries_updated_at
    BEFORE UPDATE ON core_dict.dict_entries
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_claims_updated_at
    BEFORE UPDATE ON core_dict.claims
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_portabilities_updated_at
    BEFORE UPDATE ON core_dict.portabilities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Apply audit trigger to critical tables
CREATE TRIGGER audit_entries
    AFTER INSERT OR UPDATE OR DELETE ON core_dict.dict_entries
    FOR EACH ROW EXECUTE FUNCTION audit_entry_changes();

CREATE TRIGGER audit_claims
    AFTER INSERT OR UPDATE OR DELETE ON core_dict.claims
    FOR EACH ROW EXECUTE FUNCTION audit_entry_changes();

CREATE TRIGGER audit_portabilities
    AFTER INSERT OR UPDATE OR DELETE ON core_dict.portabilities
    FOR EACH ROW EXECUTE FUNCTION audit_entry_changes();

CREATE TRIGGER audit_accounts
    AFTER INSERT OR UPDATE OR DELETE ON core_dict.accounts
    FOR EACH ROW EXECUTE FUNCTION audit_entry_changes();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop triggers
DROP TRIGGER IF EXISTS audit_accounts ON core_dict.accounts;
DROP TRIGGER IF EXISTS audit_portabilities ON core_dict.portabilities;
DROP TRIGGER IF EXISTS audit_claims ON core_dict.claims;
DROP TRIGGER IF EXISTS audit_entries ON core_dict.dict_entries;

DROP TRIGGER IF EXISTS update_portabilities_updated_at ON core_dict.portabilities;
DROP TRIGGER IF EXISTS update_claims_updated_at ON core_dict.claims;
DROP TRIGGER IF EXISTS update_entries_updated_at ON core_dict.dict_entries;
DROP TRIGGER IF EXISTS update_accounts_updated_at ON core_dict.accounts;
DROP TRIGGER IF EXISTS update_users_updated_at ON core_dict.users;

-- Drop functions
DROP FUNCTION IF EXISTS expire_old_claims();
DROP FUNCTION IF EXISTS audit_entry_changes();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- +goose StatementEnd
