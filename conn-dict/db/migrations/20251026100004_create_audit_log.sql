-- Migration: 20251026100004_create_audit_log
-- Description: Create audit_log table for compliance and tracking

CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Entity information
    entity_type VARCHAR(50) NOT NULL CHECK (entity_type IN ('ENTRY', 'CLAIM', 'INFRACTION')),
    entity_id UUID NOT NULL,

    -- Action information
    action VARCHAR(50) NOT NULL CHECK (action IN (
        'CREATE',
        'UPDATE',
        'DELETE',
        'BLOCK',
        'UNBLOCK',
        'CLAIM_CREATED',
        'CLAIM_CONFIRMED',
        'CLAIM_CANCELLED',
        'CLAIM_COMPLETED',
        'CLAIM_EXPIRED',
        'INFRACTION_REPORTED',
        'INFRACTION_ACKNOWLEDGED',
        'INFRACTION_RESOLVED'
    )),

    -- Actor information
    actor_type VARCHAR(50) NOT NULL CHECK (actor_type IN ('USER', 'SYSTEM', 'BACEN', 'EXTERNAL')),
    actor_id VARCHAR(100),
    actor_ispb CHAR(8),

    -- Change details
    old_values JSONB,
    new_values JSONB,
    changes JSONB,

    -- Context
    ip_address INET,
    user_agent TEXT,
    request_id VARCHAR(100),
    correlation_id VARCHAR(100),

    -- Timestamp
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Metadata
    reason TEXT,
    metadata JSONB
);

-- Partitioning by month (for performance with large datasets)
-- This is a placeholder - actual partitioning should be set up per requirements

-- Indexes
CREATE INDEX idx_audit_log_entity ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_log_action ON audit_log(action);
CREATE INDEX idx_audit_log_actor_id ON audit_log(actor_id);
CREATE INDEX idx_audit_log_actor_ispb ON audit_log(actor_ispb) WHERE actor_ispb IS NOT NULL;
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at DESC);
CREATE INDEX idx_audit_log_request_id ON audit_log(request_id) WHERE request_id IS NOT NULL;

-- GIN index for JSONB columns (for fast JSON queries)
CREATE INDEX idx_audit_log_old_values ON audit_log USING GIN (old_values) WHERE old_values IS NOT NULL;
CREATE INDEX idx_audit_log_new_values ON audit_log USING GIN (new_values) WHERE new_values IS NOT NULL;
CREATE INDEX idx_audit_log_metadata ON audit_log USING GIN (metadata) WHERE metadata IS NOT NULL;

-- Comments
COMMENT ON TABLE audit_log IS 'Comprehensive audit log for LGPD compliance and tracking';
COMMENT ON COLUMN audit_log.entity_type IS 'Type of entity being audited';
COMMENT ON COLUMN audit_log.old_values IS 'JSON snapshot of values before change';
COMMENT ON COLUMN audit_log.new_values IS 'JSON snapshot of values after change';
COMMENT ON COLUMN audit_log.changes IS 'JSON object showing what changed (diff)';