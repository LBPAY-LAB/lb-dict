# IMP-005: Database Migration Guide

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: Database Migration Management with Goose
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: BACKEND (AI Agent - Backend Developer)

---

## Sumário Executivo

Este documento fornece o guia completo para criação, execução e gerenciamento de **migrações de banco de dados** usando **Goose**, incluindo best practices, estratégias de rollback e versionamento.

**Baseado em**:
- [IMP-001: Manual de Implementação Core DICT](./IMP-001_Manual_Implementacao_Core_DICT.md)
- [DAT-001: Schema Database Core DICT](../03_Dados/DAT-001_Schema_Database_Core_DICT.md)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | BACKEND | Versão inicial - Database Migration Guide |

---

## Índice

1. [Visão Geral do Goose](#1-visão-geral-do-goose)
2. [Setup e Instalação](#2-setup-e-instalação)
3. [Criando Migrações](#3-criando-migrações)
4. [Executando Migrações](#4-executando-migrações)
5. [Migration Best Practices](#5-migration-best-practices)
6. [Rollback Procedures](#6-rollback-procedures)
7. [Database Versioning Strategy](#7-database-versioning-strategy)
8. [Common Migration Patterns](#8-common-migration-patterns)
9. [Troubleshooting](#9-troubleshooting)

---

## 1. Visão Geral do Goose

### 1.1. O que é Goose?

**Goose** é uma ferramenta de migração de banco de dados para Go que permite:
- Versionamento incremental do schema
- Migrações up (aplicar) e down (reverter)
- Suporte para SQL e Go migrations
- Controle de versão via tabela `goose_db_version`

### 1.2. Por que Goose?

| Vantagem | Descrição |
|----------|-----------|
| **Simplicidade** | SQL puro, fácil de entender |
| **Versionamento** | Migrações numeradas sequencialmente |
| **Rollback** | Suporte nativo para reverter migrações |
| **Go Native** | Integração natural com aplicações Go |
| **Multi-DB** | Suporte para PostgreSQL, MySQL, SQLite, etc. |

### 1.3. Estrutura de Migração

```sql
-- +goose Up
-- SQL para aplicar migração (forward)
CREATE TABLE users (...);

-- +goose Down
-- SQL para reverter migração (backward)
DROP TABLE users;
```

---

## 2. Setup e Instalação

### 2.1. Instalar Goose

```bash
# Via Go install
go install github.com/pressly/goose/v3/cmd/goose@latest

# Verificar instalação
goose -version
```

### 2.2. Configurar Variáveis de Ambiente

**Criar arquivo `.env.goose`**:

```bash
# .env.goose
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="host=localhost port=5432 user=dict_app password=secure_password_here dbname=lbpay_core_dict sslmode=disable"
export GOOSE_MIGRATION_DIR=./db/migrations
```

**Carregar variáveis**:
```bash
source .env.goose
```

### 2.3. Estrutura de Diretórios

```
core-dict/
├── db/
│   ├── migrations/
│   │   ├── 00001_create_schema.sql
│   │   ├── 00002_create_tables.sql
│   │   ├── 00003_create_indexes.sql
│   │   └── 00004_add_column_to_entries.sql
│   └── seeds/
│       └── 00001_seed_users.sql
└── scripts/
    ├── migrate.sh
    └── rollback.sh
```

---

## 3. Criando Migrações

### 3.1. Criar Nova Migração

**Comando**:
```bash
cd db/migrations
goose create <migration_name> sql
```

**Exemplo**:
```bash
goose create add_status_to_claims sql
```

**Output**:
```
Created new file: 00005_add_status_to_claims.sql
```

### 3.2. Template de Migração

**Arquivo**: `db/migrations/00005_add_status_to_claims.sql`

```sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE dict.claims
ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT 'OPEN'
CHECK (status IN ('OPEN', 'WAITING_RESOLUTION', 'CONFIRMED', 'CANCELLED', 'COMPLETED', 'EXPIRED'));

CREATE INDEX idx_claims_status ON dict.claims (status);

COMMENT ON COLUMN dict.claims.status IS 'Current claim status';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS dict.idx_claims_status;

ALTER TABLE dict.claims
DROP COLUMN IF EXISTS status;
-- +goose StatementEnd
```

### 3.3. Tipos de Migrações

#### A. Migrações SQL (Recomendado)

**Quando usar**: Para operações DDL (CREATE, ALTER, DROP)

```sql
-- +goose Up
CREATE TABLE dict.new_table (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS dict.new_table;
```

#### B. Migrações Go

**Quando usar**: Para operações complexas, seeding, transformações de dados

```go
package migrations

import (
    "database/sql"
    "github.com/pressly/goose/v3"
)

func init() {
    goose.AddMigration(upDataMigration, downDataMigration)
}

func upDataMigration(tx *sql.Tx) error {
    // Complex data transformation
    rows, err := tx.Query("SELECT id, old_field FROM table")
    if err != nil {
        return err
    }
    defer rows.Close()

    for rows.Next() {
        var id string
        var oldField string
        if err := rows.Scan(&id, &oldField); err != nil {
            return err
        }

        newField := transformData(oldField)
        if _, err := tx.Exec("UPDATE table SET new_field = $1 WHERE id = $2", newField, id); err != nil {
            return err
        }
    }

    return nil
}

func downDataMigration(tx *sql.Tx) error {
    // Reverse transformation
    return nil
}
```

---

## 4. Executando Migrações

### 4.1. Comandos Básicos

```bash
# Ver status de migrações
goose status

# Aplicar todas as migrações pendentes
goose up

# Aplicar próxima migração
goose up-by-one

# Aplicar até versão específica
goose up-to 20250101000000

# Reverter última migração
goose down

# Reverter todas as migrações
goose reset

# Verificar versão atual
goose version
```

### 4.2. Script de Migração Automatizado

**Arquivo**: `scripts/migrate.sh`

```bash
#!/bin/bash
set -e

# Load environment variables
source .env.goose

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Database Migration Script ===${NC}"

# Check Goose installation
if ! command -v goose &> /dev/null; then
    echo -e "${RED}ERROR: Goose is not installed${NC}"
    echo "Install with: go install github.com/pressly/goose/v3/cmd/goose@latest"
    exit 1
fi

# Check database connection
echo -e "${YELLOW}Testing database connection...${NC}"
if ! goose version &> /dev/null; then
    echo -e "${RED}ERROR: Cannot connect to database${NC}"
    echo "Check your GOOSE_DBSTRING environment variable"
    exit 1
fi
echo -e "${GREEN}Database connection OK${NC}"

# Show current status
echo -e "${YELLOW}Current migration status:${NC}"
goose status

# Ask for confirmation
echo ""
read -p "Apply pending migrations? (y/n): " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Applying migrations...${NC}"
    goose up
    echo -e "${GREEN}Migrations applied successfully!${NC}"

    echo -e "${YELLOW}Final status:${NC}"
    goose status
else
    echo -e "${YELLOW}Migration cancelled${NC}"
fi
```

**Tornar executável**:
```bash
chmod +x scripts/migrate.sh
```

**Executar**:
```bash
./scripts/migrate.sh
```

### 4.3. Script de Rollback

**Arquivo**: `scripts/rollback.sh`

```bash
#!/bin/bash
set -e

source .env.goose

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}=== Database Rollback Script ===${NC}"

# Show current status
echo -e "${YELLOW}Current migration status:${NC}"
goose status

echo ""
echo -e "${RED}WARNING: This will rollback the last migration!${NC}"
read -p "Are you sure? (y/n): " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Rolling back last migration...${NC}"
    goose down
    echo -e "${GREEN}Rollback completed!${NC}"

    echo -e "${YELLOW}Final status:${NC}"
    goose status
else
    echo -e "${YELLOW}Rollback cancelled${NC}"
fi
```

---

## 5. Migration Best Practices

### 5.1. Naming Conventions

**Pattern**: `<timestamp>_<descriptive_name>.sql`

**Examples**:
```
00001_create_schema.sql
00002_create_tables.sql
00003_create_indexes.sql
00004_add_status_to_claims.sql
00005_add_external_id_to_entries.sql
```

### 5.2. Migration Guidelines

#### DO

- ✅ Keep migrations **small and focused** (one change per migration)
- ✅ Always provide **down migrations** (rollback)
- ✅ Test migrations on **development** before production
- ✅ Use **transactions** for multi-statement migrations
- ✅ Add **comments** explaining complex operations
- ✅ Use **IF EXISTS** / **IF NOT EXISTS** for idempotency
- ✅ Create **indexes concurrently** in PostgreSQL

#### DON'T

- ❌ Don't modify existing migrations (create new ones)
- ❌ Don't delete old migrations
- ❌ Don't use `SELECT *` in migrations
- ❌ Don't depend on application code in migrations
- ❌ Don't mix DDL and DML in same migration
- ❌ Don't create migrations without testing down path

### 5.3. Idempotent Migrations

**Always use IF EXISTS / IF NOT EXISTS**:

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS dict.new_table (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL
);

ALTER TABLE dict.entries
ADD COLUMN IF NOT EXISTS new_field VARCHAR(100);

CREATE INDEX IF NOT EXISTS idx_entries_new_field
ON dict.entries (new_field);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS dict.idx_entries_new_field;

ALTER TABLE dict.entries
DROP COLUMN IF EXISTS new_field;

DROP TABLE IF EXISTS dict.new_table;
-- +goose StatementEnd
```

### 5.4. Transactional Migrations

**Wrap multiple statements in transactions**:

```sql
-- +goose Up
-- +goose StatementBegin
BEGIN;

CREATE TABLE dict.temp_table (
    id UUID PRIMARY KEY,
    data JSONB
);

INSERT INTO dict.temp_table (id, data)
SELECT id, row_to_json(entries.*) FROM dict.entries;

ALTER TABLE dict.entries ADD COLUMN metadata JSONB;

UPDATE dict.entries e
SET metadata = t.data
FROM dict.temp_table t
WHERE e.id = t.id;

DROP TABLE dict.temp_table;

COMMIT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE dict.entries DROP COLUMN IF EXISTS metadata;
-- +goose StatementEnd
```

### 5.5. Online Migrations (Zero Downtime)

**For large tables, use concurrent index creation**:

```sql
-- +goose Up
-- +goose StatementBegin
-- Create index concurrently (doesn't lock table)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_entries_key_value
ON dict.entries (key_value);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS dict.idx_entries_key_value;
-- +goose StatementEnd
```

---

## 6. Rollback Procedures

### 6.1. Rollback Strategy

**Levels of rollback**:

1. **Single migration**: Rollback last migration
2. **Multiple migrations**: Rollback to specific version
3. **Full rollback**: Rollback all migrations (DANGER)

### 6.2. Safe Rollback Checklist

Before rollback:

- [ ] **Backup database**
- [ ] Verify rollback SQL is correct
- [ ] Test rollback on staging
- [ ] Check for data dependencies
- [ ] Notify team of rollback
- [ ] Have restore plan ready

### 6.3. Rollback Commands

**Rollback last migration**:
```bash
goose down
```

**Rollback to specific version**:
```bash
goose down-to 00003
```

**Rollback all migrations** (DANGER):
```bash
goose reset
```

### 6.4. Emergency Rollback Procedure

**Step 1: Stop application**
```bash
# Stop all application instances
systemctl stop core-dict-api
```

**Step 2: Backup database**
```bash
pg_dump -h localhost -U dict_app -d lbpay_core_dict > backup_$(date +%Y%m%d_%H%M%S).sql
```

**Step 3: Rollback migration**
```bash
goose down
```

**Step 4: Verify database state**
```bash
goose status
psql -U dict_app -d lbpay_core_dict -c "\dt dict.*"
```

**Step 5: Restart application**
```bash
systemctl start core-dict-api
```

### 6.5. Rollback Testing

**Always test rollback before deploying**:

```bash
# Apply migration
goose up-by-one

# Test application
./test_app.sh

# Rollback migration
goose down

# Test application still works
./test_app.sh

# Re-apply migration
goose up-by-one
```

---

## 7. Database Versioning Strategy

### 7.1. Version Tracking

**Goose uses table `goose_db_version`**:

```sql
SELECT * FROM goose_db_version ORDER BY id DESC;
```

**Output**:
```
 id |   version_id   |      is_applied      | tstamp
----+----------------+----------------------+-------------------------
  1 | 00001          | t                    | 2025-10-25 10:00:00
  2 | 00002          | t                    | 2025-10-25 10:05:00
  3 | 00003          | t                    | 2025-10-25 10:10:00
```

### 7.2. Semantic Versioning

**Migration naming strategy**:

```
<sequence>_<version>_<description>.sql

Examples:
00001_v1.0_create_schema.sql
00002_v1.0_create_tables.sql
00003_v1.0_create_indexes.sql
00004_v1.1_add_status_to_claims.sql
00005_v1.2_add_external_id.sql
```

### 7.3. Environment-Specific Migrations

**Directory structure**:

```
db/
├── migrations/
│   ├── 00001_create_schema.sql        # Shared
│   ├── 00002_create_tables.sql        # Shared
│   └── 00003_create_indexes.sql       # Shared
└── migrations-env/
    ├── dev/
    │   └── 00100_seed_test_data.sql   # Dev only
    ├── staging/
    │   └── 00200_staging_config.sql   # Staging only
    └── prod/
        └── 00300_prod_config.sql      # Prod only
```

**Run environment-specific migrations**:
```bash
# Production migrations
export GOOSE_MIGRATION_DIR=./db/migrations
goose up

# Dev seeds
export GOOSE_MIGRATION_DIR=./db/migrations-env/dev
goose up
```

---

## 8. Common Migration Patterns

### 8.1. Add Column

```sql
-- +goose Up
ALTER TABLE dict.entries
ADD COLUMN external_id VARCHAR(100);

CREATE INDEX idx_entries_external_id ON dict.entries (external_id);

-- +goose Down
DROP INDEX IF EXISTS dict.idx_entries_external_id;

ALTER TABLE dict.entries
DROP COLUMN IF EXISTS external_id;
```

### 8.2. Rename Column

```sql
-- +goose Up
ALTER TABLE dict.entries
RENAME COLUMN old_name TO new_name;

-- +goose Down
ALTER TABLE dict.entries
RENAME COLUMN new_name TO old_name;
```

### 8.3. Change Column Type

```sql
-- +goose Up
-- Step 1: Add new column
ALTER TABLE dict.entries
ADD COLUMN status_new VARCHAR(50);

-- Step 2: Migrate data
UPDATE dict.entries
SET status_new = status::VARCHAR;

-- Step 3: Drop old column
ALTER TABLE dict.entries
DROP COLUMN status;

-- Step 4: Rename new column
ALTER TABLE dict.entries
RENAME COLUMN status_new TO status;

-- +goose Down
-- Reverse process
ALTER TABLE dict.entries
ADD COLUMN status_old VARCHAR(20);

UPDATE dict.entries
SET status_old = status;

ALTER TABLE dict.entries
DROP COLUMN status;

ALTER TABLE dict.entries
RENAME COLUMN status_old TO status;
```

### 8.4. Add Foreign Key

```sql
-- +goose Up
ALTER TABLE dict.entries
ADD CONSTRAINT fk_entries_account
FOREIGN KEY (account_id)
REFERENCES dict.accounts(id)
ON DELETE RESTRICT;

-- +goose Down
ALTER TABLE dict.entries
DROP CONSTRAINT IF EXISTS fk_entries_account;
```

### 8.5. Create Enum Type

```sql
-- +goose Up
CREATE TYPE dict.entry_status AS ENUM (
    'PENDING',
    'ACTIVE',
    'DELETED',
    'CLAIM_PENDING'
);

ALTER TABLE dict.entries
ALTER COLUMN status TYPE dict.entry_status USING status::dict.entry_status;

-- +goose Down
ALTER TABLE dict.entries
ALTER COLUMN status TYPE VARCHAR(20);

DROP TYPE IF EXISTS dict.entry_status;
```

### 8.6. Add Check Constraint

```sql
-- +goose Up
ALTER TABLE dict.entries
ADD CONSTRAINT chk_key_value_length
CHECK (LENGTH(key_value) >= 3);

-- +goose Down
ALTER TABLE dict.entries
DROP CONSTRAINT IF EXISTS chk_key_value_length;
```

---

## 9. Troubleshooting

### 9.1. Common Issues

#### Issue: Migration version mismatch

**Error**:
```
Error: version mismatch: database is at version 5, but migrations go up to 3
```

**Solution**:
```bash
# Check database version
goose version

# Check migration files
ls -la db/migrations/

# Reset to correct version
goose down-to 00003
```

#### Issue: Migration fails mid-way

**Error**:
```
Error: migration failed: pq: duplicate key value violates unique constraint
```

**Solution**:
```bash
# Check current state
goose status

# Manual fix in database
psql -U dict_app -d lbpay_core_dict

# Mark migration as failed in goose_db_version
UPDATE goose_db_version SET is_applied = false WHERE version_id = 5;

# Fix data issue manually
-- (run SQL to fix)

# Retry migration
goose up-by-one
```

#### Issue: Cannot connect to database

**Error**:
```
Error: failed to connect to database
```

**Solution**:
```bash
# Check connection string
echo $GOOSE_DBSTRING

# Test connection manually
psql "$GOOSE_DBSTRING"

# Verify credentials
cat .env.goose
```

### 9.2. Recovery Procedures

**Restore from backup**:
```bash
# Stop application
systemctl stop core-dict-api

# Restore database
psql -U dict_app -d lbpay_core_dict < backup_20251025_100000.sql

# Reset migration version
psql -U dict_app -d lbpay_core_dict -c "DELETE FROM goose_db_version WHERE version_id > 3;"

# Restart application
systemctl start core-dict-api
```

---

## Rastreabilidade

### Requisitos Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RF-MIG-001 | Criar e executar migrações com Goose | [IMP-001](./IMP-001_Manual_Implementacao_Core_DICT.md) | ✅ Especificado |
| RF-MIG-002 | Rollback de migrações | Best Practices | ✅ Especificado |
| RF-MIG-003 | Versionamento de database | Best Practices | ✅ Especificado |

---

## Próximas Revisões

**Pendências**:
- [ ] Adicionar suporte para migrações multi-tenant
- [ ] Implementar CI/CD integration para migrações
- [ ] Adicionar health checks pós-migração
- [ ] Criar scripts de validação de schema

---

**Referências**:
- [Goose Documentation](https://github.com/pressly/goose)
- [PostgreSQL ALTER TABLE](https://www.postgresql.org/docs/current/sql-altertable.html)
- [IMP-001: Manual de Implementação Core DICT](./IMP-001_Manual_Implementacao_Core_DICT.md)
- [DAT-001: Schema Database Core DICT](../03_Dados/DAT-001_Schema_Database_Core_DICT.md)
