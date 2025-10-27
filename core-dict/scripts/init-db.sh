#!/bin/bash
# ============================================
# Core DICT - Database Initialization Script
# ============================================
# This script initializes the PostgreSQL database for Core DICT.
# It creates the database, schemas, and runs migrations.

set -e  # Exit on error
set -u  # Exit on undefined variable

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values (can be overridden by environment variables)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${POSTGRES_DB:-core_dict}"
DB_USER="${POSTGRES_USER:-postgres}"
DB_PASSWORD="${POSTGRES_PASSWORD:-postgres}"
MIGRATION_PATH="${MIGRATION_PATH:-./migrations}"

# Function to print colored messages
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to wait for PostgreSQL to be ready
wait_for_postgres() {
    print_info "Waiting for PostgreSQL to be ready..."

    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c '\q' 2>/dev/null; then
            print_success "PostgreSQL is ready!"
            return 0
        fi

        print_warning "Attempt $attempt/$max_attempts: PostgreSQL not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done

    print_error "PostgreSQL did not become ready in time"
    return 1
}

# Function to check if database exists
database_exists() {
    PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'" | grep -q 1
}

# Function to create database
create_database() {
    print_info "Creating database '$DB_NAME'..."

    if database_exists; then
        print_warning "Database '$DB_NAME' already exists. Skipping creation."
    else
        PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "CREATE DATABASE $DB_NAME;"
        print_success "Database '$DB_NAME' created successfully!"
    fi
}

# Function to create schemas
create_schemas() {
    print_info "Creating schemas..."

    PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" <<-EOSQL
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
EOSQL

    print_success "Schemas created successfully!"
}

# Function to run migrations
run_migrations() {
    print_info "Running database migrations..."

    if [ ! -d "$MIGRATION_PATH" ]; then
        print_warning "Migration path '$MIGRATION_PATH' not found. Skipping migrations."
        return 0
    fi

    # Check if goose is installed
    if ! command -v goose &> /dev/null; then
        print_warning "Goose migration tool not found. Skipping migrations."
        print_info "To install goose: go install github.com/pressly/goose/v3/cmd/goose@latest"
        return 0
    fi

    # Run goose migrations
    export GOOSE_DRIVER=postgres
    export GOOSE_DBSTRING="host=$DB_HOST port=$DB_PORT user=$DB_USER password=$DB_PASSWORD dbname=$DB_NAME sslmode=disable"

    goose -dir "$MIGRATION_PATH" up

    print_success "Migrations completed successfully!"
}

# Function to verify database setup
verify_setup() {
    print_info "Verifying database setup..."

    local schema_count=$(PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name IN ('core_dict', 'audit', 'config')")

    if [ "$schema_count" -eq 3 ]; then
        print_success "All schemas are present!"
    else
        print_error "Expected 3 schemas, found $schema_count"
        return 1
    fi

    # Check tables
    local table_count=$(PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'core_dict'")
    print_info "Found $table_count tables in core_dict schema"

    print_success "Database verification completed!"
}

# Function to display connection info
display_connection_info() {
    echo ""
    print_success "Database initialization completed!"
    echo ""
    print_info "Connection details:"
    echo "  Host:     $DB_HOST"
    echo "  Port:     $DB_PORT"
    echo "  Database: $DB_NAME"
    echo "  User:     $DB_USER"
    echo ""
    print_info "Connection string:"
    echo "  postgres://$DB_USER:***@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"
    echo ""
    print_info "To connect with psql:"
    echo "  PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME"
    echo ""
}

# Main execution
main() {
    echo ""
    print_info "===== Core DICT Database Initialization ====="
    echo ""

    # Wait for PostgreSQL
    if ! wait_for_postgres; then
        exit 1
    fi

    # Create database
    create_database

    # Create schemas
    create_schemas

    # Run migrations
    run_migrations

    # Verify setup
    verify_setup

    # Display connection info
    display_connection_info
}

# Run main function
main
