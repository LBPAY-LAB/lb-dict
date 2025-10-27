-- Migration: 20251026100000_create_extensions
-- Description: Create PostgreSQL extensions required for DICT

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable pgcrypto for encryption functions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enable btree_gist for advanced indexing
CREATE EXTENSION IF NOT EXISTS "btree_gist";