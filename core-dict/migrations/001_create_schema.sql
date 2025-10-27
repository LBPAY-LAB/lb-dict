-- Migration: 001_create_schema
-- Description: Create core_dict schema and enable required extensions
-- Author: data-specialist-core
-- Date: 2025-10-27

-- +goose Up
-- +goose StatementBegin

-- Create schemas
CREATE SCHEMA IF NOT EXISTS core_dict;
CREATE SCHEMA IF NOT EXISTS audit;
CREATE SCHEMA IF NOT EXISTS config;

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";          -- UUID generation
CREATE EXTENSION IF NOT EXISTS "pg_trgm";            -- Full-text search (trigram)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";           -- Encryption functions
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements"; -- Query performance monitoring

-- Comments
COMMENT ON SCHEMA core_dict IS 'Core DICT business data - PIX keys, accounts, claims';
COMMENT ON SCHEMA audit IS 'Audit logs for compliance (LGPD/Bacen)';
COMMENT ON SCHEMA config IS 'System configuration tables';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop schemas (cascade to drop all tables)
DROP SCHEMA IF EXISTS config CASCADE;
DROP SCHEMA IF EXISTS audit CASCADE;
DROP SCHEMA IF EXISTS core_dict CASCADE;

-- Note: Extensions are not dropped to avoid breaking other databases

-- +goose StatementEnd
