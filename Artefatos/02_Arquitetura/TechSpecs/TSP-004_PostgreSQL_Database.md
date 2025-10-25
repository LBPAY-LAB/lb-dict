# TSP-004: PostgreSQL Database - Technical Specification

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: PostgreSQL Database Layer
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: BACKEND (AI Agent - Backend Specialist)
**Revisor**: [Aguardando]
**Aprovador**: Tech Lead, Head de Arquitetura

---

## Sumário Executivo

Este documento especifica a implementação do **PostgreSQL 16** para o projeto DICT LBPay, cobrindo deployment em Kubernetes (primary + standby), Row-Level Security (RLS) policies, particionamento de tabelas (audit logs), estratégias de backup/restore, e connection pooling com PgBouncer.

**Baseado em**:
- [DAT-001: Schema Database Core DICT](../../03_Dados/DAT-001_Schema_Database_Core_DICT.md)
- [DAT-002: Schema Database Connect](../../03_Dados/DAT-002_Schema_Database_Connect.md)
- [TEC-003: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | BACKEND | Versão inicial - PostgreSQL 16 specification |

---

## Índice

1. [Visão Geral](#1-visão-geral)
2. [Deployment Kubernetes](#2-deployment-kubernetes)
3. [Database Schema](#3-database-schema)
4. [Row-Level Security (RLS)](#4-row-level-security-rls)
5. [Table Partitioning](#5-table-partitioning)
6. [Connection Pooling (PgBouncer)](#6-connection-pooling-pgbouncer)
7. [Backup & Restore](#7-backup--restore)
8. [Monitoring & Observability](#8-monitoring--observability)
9. [Performance Tuning](#9-performance-tuning)
10. [High Availability](#10-high-availability)

---

## 1. Visão Geral

### 1.1. PostgreSQL Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    PostgreSQL 16 Cluster                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Primary Database (Read/Write)                             │ │
│  │  - StatefulSet Pod 0                                       │ │
│  │  - Port: 5432                                              │ │
│  │  - Handles all writes                                      │ │
│  │  - Streaming replication to standby                        │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           ↓ (replication)                         │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Standby Database (Read-Only)                              │ │
│  │  - StatefulSet Pod 1                                       │ │
│  │  - Port: 5432                                              │ │
│  │  - Handles read queries (load balancing)                   │ │
│  │  - Hot standby (can be promoted to primary)                │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│                    PgBouncer (Connection Pool)                   │
│  - Deployment (3 replicas)                                       │
│  - Max connections: 1000                                         │
│  - Pool mode: transaction                                        │
│  - Routes writes to primary, reads to standby                    │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│                  DICT Applications                               │
│  - Core DICT Service (CRUD operations)                           │
│  - Connect Service (Bridge integration)                          │
│  - Temporal Workers (workflow persistence)                       │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2. Key Features

| Feature | Value | Justification |
|---------|-------|---------------|
| **PostgreSQL Version** | 16.2 | Latest stable (as of 2025-10-25) |
| **Deployment** | Kubernetes StatefulSet | Persistent storage, stable network IDs |
| **HA Mode** | Primary + Hot Standby | Read scalability, failover |
| **Replication** | Streaming (async) | Low latency, near-real-time |
| **Connection Pool** | PgBouncer | Reduce connection overhead |
| **Partitioning** | Range (by date) | Efficient audit log queries |
| **RLS** | Row-Level Security | Multi-tenant isolation |
| **Backup** | WAL-E + pg_dump | Point-in-time recovery |
| **Storage** | 100Gi SSD | Fast I/O for high throughput |

---

## 2. Deployment Kubernetes

### 2.1. Namespace

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: postgres
  labels:
    name: postgres
    environment: production
```

### 2.2. PostgreSQL ConfigMap

```yaml
# k8s/postgres-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: postgres-config
  namespace: postgres
data:
  postgresql.conf: |
    # Connection settings
    listen_addresses = '*'
    port = 5432
    max_connections = 200
    superuser_reserved_connections = 3

    # Memory settings
    shared_buffers = 2GB
    effective_cache_size = 6GB
    maintenance_work_mem = 512MB
    work_mem = 16MB

    # WAL settings
    wal_level = replica
    max_wal_size = 4GB
    min_wal_size = 1GB
    wal_compression = on
    wal_buffers = 16MB

    # Replication settings
    max_wal_senders = 10
    max_replication_slots = 10
    hot_standby = on
    hot_standby_feedback = on

    # Query planning
    random_page_cost = 1.1
    effective_io_concurrency = 200
    default_statistics_target = 100

    # Logging
    logging_collector = on
    log_directory = 'log'
    log_filename = 'postgresql-%Y-%m-%d_%H%M%S.log'
    log_rotation_age = 1d
    log_rotation_size = 100MB
    log_min_duration_statement = 1000
    log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
    log_checkpoints = on
    log_connections = on
    log_disconnections = on
    log_lock_waits = on
    log_temp_files = 0

    # Autovacuum
    autovacuum = on
    autovacuum_max_workers = 4
    autovacuum_naptime = 30s
    autovacuum_vacuum_scale_factor = 0.1
    autovacuum_analyze_scale_factor = 0.05

    # Performance
    checkpoint_completion_target = 0.9
    default_toast_compression = lz4

  pg_hba.conf: |
    # TYPE  DATABASE        USER            ADDRESS                 METHOD
    local   all             all                                     trust
    host    all             all             0.0.0.0/0               md5
    host    all             all             ::0/0                   md5
    host    replication     replicator      0.0.0.0/0               md5
```

### 2.3. PostgreSQL StatefulSet (Primary)

```yaml
# k8s/postgres-primary.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres-primary
  namespace: postgres
spec:
  serviceName: postgres-headless
  replicas: 1
  selector:
    matchLabels:
      app: postgres
      role: primary
  template:
    metadata:
      labels:
        app: postgres
        role: primary
    spec:
      containers:
      - name: postgres
        image: postgres:16.2-alpine
        ports:
        - name: postgres
          containerPort: 5432
        env:
        - name: POSTGRES_DB
          value: "dict"
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: postgres-secrets
              key: postgres-user
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secrets
              key: postgres-password
        - name: PGDATA
          value: /var/lib/postgresql/data/pgdata
        - name: POSTGRES_INITDB_ARGS
          value: "--encoding=UTF8 --locale=en_US.UTF-8"
        volumeMounts:
        - name: postgres-data
          mountPath: /var/lib/postgresql/data
        - name: config
          mountPath: /etc/postgresql
        - name: init-scripts
          mountPath: /docker-entrypoint-initdb.d
        resources:
          requests:
            memory: "4Gi"
            cpu: "2000m"
          limits:
            memory: "8Gi"
            cpu: "4000m"
        livenessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - $(POSTGRES_USER)
            - -d
            - $(POSTGRES_DB)
          initialDelaySeconds: 60
          periodSeconds: 10
          timeoutSeconds: 5
        readinessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - $(POSTGRES_USER)
            - -d
            - $(POSTGRES_DB)
          initialDelaySeconds: 30
          periodSeconds: 5
          timeoutSeconds: 3
      volumes:
      - name: config
        configMap:
          name: postgres-config
      - name: init-scripts
        configMap:
          name: postgres-init-scripts
  volumeClaimTemplates:
  - metadata:
      name: postgres-data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "fast-ssd"
      resources:
        requests:
          storage: 100Gi
```

### 2.4. PostgreSQL StatefulSet (Standby)

```yaml
# k8s/postgres-standby.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres-standby
  namespace: postgres
spec:
  serviceName: postgres-standby-headless
  replicas: 1
  selector:
    matchLabels:
      app: postgres
      role: standby
  template:
    metadata:
      labels:
        app: postgres
        role: standby
    spec:
      initContainers:
      - name: setup-replication
        image: postgres:16.2-alpine
        command:
        - /bin/sh
        - -c
        - |
          if [ ! -f /var/lib/postgresql/data/pgdata/PG_VERSION ]; then
            pg_basebackup -h postgres-primary -U replicator -D /var/lib/postgresql/data/pgdata -P -v -R -X stream -C -S standby_slot
          fi
        env:
        - name: PGPASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secrets
              key: replicator-password
        volumeMounts:
        - name: postgres-data
          mountPath: /var/lib/postgresql/data
      containers:
      - name: postgres
        image: postgres:16.2-alpine
        ports:
        - name: postgres
          containerPort: 5432
        env:
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: postgres-secrets
              key: postgres-user
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secrets
              key: postgres-password
        - name: PGDATA
          value: /var/lib/postgresql/data/pgdata
        volumeMounts:
        - name: postgres-data
          mountPath: /var/lib/postgresql/data
        - name: config
          mountPath: /etc/postgresql
        resources:
          requests:
            memory: "4Gi"
            cpu: "1000m"
          limits:
            memory: "8Gi"
            cpu: "2000m"
        livenessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - $(POSTGRES_USER)
          initialDelaySeconds: 60
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - $(POSTGRES_USER)
          initialDelaySeconds: 30
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: postgres-config
  volumeClaimTemplates:
  - metadata:
      name: postgres-data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "fast-ssd"
      resources:
        requests:
          storage: 100Gi
```

### 2.5. PostgreSQL Services

```yaml
# k8s/postgres-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: postgres-primary
  namespace: postgres
spec:
  type: ClusterIP
  ports:
  - name: postgres
    port: 5432
    targetPort: 5432
  selector:
    app: postgres
    role: primary
---
apiVersion: v1
kind: Service
metadata:
  name: postgres-standby
  namespace: postgres
spec:
  type: ClusterIP
  ports:
  - name: postgres
    port: 5432
    targetPort: 5432
  selector:
    app: postgres
    role: standby
---
apiVersion: v1
kind: Service
metadata:
  name: postgres-headless
  namespace: postgres
spec:
  type: ClusterIP
  clusterIP: None
  ports:
  - name: postgres
    port: 5432
    targetPort: 5432
  selector:
    app: postgres
```

### 2.6. Secrets

```yaml
# k8s/postgres-secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: postgres-secrets
  namespace: postgres
type: Opaque
stringData:
  postgres-user: dict_admin
  postgres-password: <REPLACE_WITH_SECURE_PASSWORD>
  replicator-password: <REPLACE_WITH_SECURE_PASSWORD>
```

---

## 3. Database Schema

### 3.1. Database Initialization

```yaml
# k8s/postgres-init-scripts.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: postgres-init-scripts
  namespace: postgres
data:
  01-create-databases.sql: |
    -- Create databases
    CREATE DATABASE dict_core;
    CREATE DATABASE dict_connect;
    CREATE DATABASE temporal;

    -- Create users
    CREATE USER dict_core_user WITH ENCRYPTED PASSWORD 'CHANGE_ME';
    CREATE USER dict_connect_user WITH ENCRYPTED PASSWORD 'CHANGE_ME';
    CREATE USER temporal_user WITH ENCRYPTED PASSWORD 'CHANGE_ME';
    CREATE USER replicator WITH REPLICATION ENCRYPTED PASSWORD 'CHANGE_ME';

    -- Grant privileges
    GRANT ALL PRIVILEGES ON DATABASE dict_core TO dict_core_user;
    GRANT ALL PRIVILEGES ON DATABASE dict_connect TO dict_connect_user;
    GRANT ALL PRIVILEGES ON DATABASE temporal TO temporal_user;

  02-create-extensions.sql: |
    \c dict_core
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
    CREATE EXTENSION IF NOT EXISTS "pg_trgm";
    CREATE EXTENSION IF NOT EXISTS "pgcrypto";

    \c dict_connect
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
    CREATE EXTENSION IF NOT EXISTS "pg_trgm";
```

### 3.2. Core DICT Schema (Simplified)

Based on [DAT-001: Schema Database Core DICT](../../03_Dados/DAT-001_Schema_Database_Core_DICT.md)

```sql
-- Core DICT tables (see DAT-001 for full schema)
CREATE TABLE entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_type VARCHAR(10) NOT NULL,
    key_value VARCHAR(100) NOT NULL,
    account_ispb CHAR(8) NOT NULL,
    account_number VARCHAR(20) NOT NULL,
    account_type VARCHAR(20) NOT NULL,
    owner_name VARCHAR(200) NOT NULL,
    owner_document VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    UNIQUE(key_type, key_value)
);

CREATE INDEX idx_entries_key ON entries(key_type, key_value);
CREATE INDEX idx_entries_ispb ON entries(account_ispb);
CREATE INDEX idx_entries_status ON entries(status);

CREATE TABLE claims (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entry_id UUID NOT NULL REFERENCES entries(id),
    claimer_ispb CHAR(8) NOT NULL,
    owner_ispb CHAR(8) NOT NULL,
    status VARCHAR(30) NOT NULL,
    workflow_id VARCHAR(100),
    expires_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_claims_entry ON claims(entry_id);
CREATE INDEX idx_claims_status ON claims(status);
CREATE INDEX idx_claims_workflow ON claims(workflow_id);
```

---

## 4. Row-Level Security (RLS)

### 4.1. RLS Policies for Multi-Tenant Isolation

**Use Case**: Isolate data by ISPB (financial institution)

**Implementation**:

```sql
-- Enable RLS on entries table
ALTER TABLE entries ENABLE ROW LEVEL SECURITY;

-- Policy: Users can only see entries owned by their ISPB
CREATE POLICY entries_isolation_policy ON entries
    FOR SELECT
    USING (account_ispb = current_setting('app.current_ispb', true));

-- Policy: Users can only insert entries for their ISPB
CREATE POLICY entries_insert_policy ON entries
    FOR INSERT
    WITH CHECK (account_ispb = current_setting('app.current_ispb', true));

-- Policy: Users can only update entries owned by their ISPB
CREATE POLICY entries_update_policy ON entries
    FOR UPDATE
    USING (account_ispb = current_setting('app.current_ispb', true));

-- Policy: Users can only delete entries owned by their ISPB
CREATE POLICY entries_delete_policy ON entries
    FOR DELETE
    USING (account_ispb = current_setting('app.current_ispb', true));

-- Enable RLS on claims table
ALTER TABLE claims ENABLE ROW LEVEL SECURITY;

-- Policy: Users can see claims where they are owner or claimer
CREATE POLICY claims_isolation_policy ON claims
    FOR SELECT
    USING (
        owner_ispb = current_setting('app.current_ispb', true) OR
        claimer_ispb = current_setting('app.current_ispb', true)
    );
```

### 4.2. Application Integration

**Set ISPB context before queries**:

```go
// internal/infrastructure/postgres/client.go
package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (db *DB) SetISPBContext(ctx context.Context, ispb string) error {
	_, err := db.pool.Exec(ctx, fmt.Sprintf("SET app.current_ispb = '%s'", ispb))
	return err
}

// Example usage in handler
func (h *Handler) GetEntry(ctx context.Context, entryID string) (*Entry, error) {
	// Extract ISPB from JWT token or request context
	ispb := extractISPBFromContext(ctx)

	// Set RLS context
	if err := h.db.SetISPBContext(ctx, ispb); err != nil {
		return nil, err
	}

	// Query will automatically apply RLS policies
	return h.repo.GetByID(ctx, entryID)
}
```

### 4.3. Bypass RLS for Admin Users

```sql
-- Create admin role that bypasses RLS
CREATE ROLE dict_admin_role BYPASSRLS;

-- Grant to admin users
GRANT dict_admin_role TO dict_admin;
```

---

## 5. Table Partitioning

### 5.1. Partitioning Strategy

**Use Case**: Audit logs grow indefinitely, need efficient queries by date

**Partition Type**: Range partitioning by `created_at`

**Partition Interval**: Monthly

### 5.2. Implementation

```sql
-- Create partitioned audit_logs table
CREATE TABLE audit_logs (
    id BIGSERIAL,
    event_type VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    user_id UUID,
    ispb CHAR(8),
    action VARCHAR(20) NOT NULL,
    old_data JSONB,
    new_data JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (created_at);

-- Create partitions for 2025
CREATE TABLE audit_logs_2025_01 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE audit_logs_2025_02 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

CREATE TABLE audit_logs_2025_03 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');

-- ... (create partitions for remaining months)

-- Create default partition for future data
CREATE TABLE audit_logs_default PARTITION OF audit_logs DEFAULT;

-- Create indexes on each partition (automatically inherited)
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_ispb ON audit_logs(ispb);
CREATE INDEX idx_audit_logs_event ON audit_logs(event_type);
```

### 5.3. Automatic Partition Creation (pg_partman)

**Install Extension**:

```sql
CREATE EXTENSION pg_partman;

-- Configure automatic partition creation
SELECT partman.create_parent(
    p_parent_table := 'public.audit_logs',
    p_control := 'created_at',
    p_type := 'native',
    p_interval := '1 month',
    p_premake := 3  -- Create 3 months ahead
);

-- Schedule maintenance (run daily)
UPDATE partman.part_config
SET retention = '12 months',
    retention_keep_table = false
WHERE parent_table = 'public.audit_logs';
```

### 5.4. Partition Maintenance Cronjob

```yaml
# k8s/partition-maintenance-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-partition-maintenance
  namespace: postgres
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: maintenance
            image: postgres:16.2-alpine
            command:
            - psql
            - -h
            - postgres-primary
            - -U
            - dict_admin
            - -d
            - dict_core
            - -c
            - "CALL partman.run_maintenance_proc()"
            env:
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgres-secrets
                  key: postgres-password
          restartPolicy: OnFailure
```

---

## 6. Connection Pooling (PgBouncer)

### 6.1. PgBouncer ConfigMap

```yaml
# k8s/pgbouncer-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: pgbouncer-config
  namespace: postgres
data:
  pgbouncer.ini: |
    [databases]
    dict_core = host=postgres-primary port=5432 dbname=dict_core
    dict_connect = host=postgres-primary port=5432 dbname=dict_connect
    temporal = host=postgres-primary port=5432 dbname=temporal
    dict_core_readonly = host=postgres-standby port=5432 dbname=dict_core

    [pgbouncer]
    listen_addr = 0.0.0.0
    listen_port = 6432
    auth_type = md5
    auth_file = /etc/pgbouncer/userlist.txt
    admin_users = dict_admin
    pool_mode = transaction
    max_client_conn = 1000
    default_pool_size = 25
    min_pool_size = 10
    reserve_pool_size = 5
    reserve_pool_timeout = 3
    server_lifetime = 3600
    server_idle_timeout = 600
    log_connections = 1
    log_disconnections = 1
    log_pooler_errors = 1

  userlist.txt: |
    "dict_core_user" "md5<MD5_HASH_OF_PASSWORD>"
    "dict_connect_user" "md5<MD5_HASH_OF_PASSWORD>"
    "temporal_user" "md5<MD5_HASH_OF_PASSWORD>"
```

### 6.2. PgBouncer Deployment

```yaml
# k8s/pgbouncer-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pgbouncer
  namespace: postgres
spec:
  replicas: 3
  selector:
    matchLabels:
      app: pgbouncer
  template:
    metadata:
      labels:
        app: pgbouncer
    spec:
      containers:
      - name: pgbouncer
        image: edoburu/pgbouncer:1.21.0
        ports:
        - name: pgbouncer
          containerPort: 6432
        volumeMounts:
        - name: config
          mountPath: /etc/pgbouncer
        resources:
          requests:
            memory: "256Mi"
            cpu: "200m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          tcpSocket:
            port: 6432
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          tcpSocket:
            port: 6432
          initialDelaySeconds: 10
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: pgbouncer-config
---
apiVersion: v1
kind: Service
metadata:
  name: pgbouncer
  namespace: postgres
spec:
  type: ClusterIP
  ports:
  - name: pgbouncer
    port: 6432
    targetPort: 6432
  selector:
    app: pgbouncer
```

### 6.3. Application Connection String

**Connect via PgBouncer**:

```go
// config.yaml
database:
  host: pgbouncer.postgres.svc.cluster.local
  port: 6432
  database: dict_core
  user: dict_core_user
  password: ${DB_PASSWORD}
  pool_size: 50  # Application pool size
  max_idle_conns: 10
```

**Why PgBouncer?**
- Reduces PostgreSQL connection overhead
- Allows 1000+ client connections with only 25 DB connections
- Transaction pooling (connections released after each transaction)

---

## 7. Backup & Restore

### 7.1. Continuous WAL Archiving

**Configuration** (in `postgresql.conf`):

```conf
archive_mode = on
archive_command = 'wal-e wal-push %p'
archive_timeout = 300  # 5 minutes
```

**WAL-E Setup**:

```yaml
# k8s/wal-e-setup.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: wal-e-config
  namespace: postgres
data:
  AWS_REGION: us-east-1
  WALE_S3_PREFIX: s3://lbpay-backups/postgres/wal
  WALE_S3_ENDPOINT: https+path://s3.amazonaws.com
```

### 7.2. Full Backup (pg_dump)

```yaml
# k8s/postgres-backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-full-backup
  namespace: postgres
spec:
  schedule: "0 3 * * *"  # Daily at 3 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:16.2-alpine
            command:
            - /bin/sh
            - -c
            - |
              pg_dump -h postgres-primary -U dict_admin -Fc dict_core > /tmp/dict_core_$(date +%Y%m%d_%H%M%S).dump
              aws s3 cp /tmp/dict_core_$(date +%Y%m%d_%H%M%S).dump s3://lbpay-backups/postgres/dumps/
            env:
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgres-secrets
                  key: postgres-password
            - name: AWS_ACCESS_KEY_ID
              valueFrom:
                secretKeyRef:
                  name: aws-secrets
                  key: access-key-id
            - name: AWS_SECRET_ACCESS_KEY
              valueFrom:
                secretKeyRef:
                  name: aws-secrets
                  key: secret-access-key
          restartPolicy: OnFailure
```

### 7.3. Backup Retention Policy

| Backup Type | Frequency | Retention | Storage |
|-------------|-----------|-----------|---------|
| WAL Archives | Continuous | 7 days | S3 |
| Full Dump (pg_dump) | Daily | 30 days | S3 |
| Monthly Full Dump | Monthly | 12 months | S3 (Glacier) |
| Annual Full Dump | Yearly | 7 years | S3 (Deep Archive) |

### 7.4. Point-in-Time Recovery (PITR)

**Scenario**: Restore database to specific timestamp

**Steps**:

1. Stop PostgreSQL
2. Restore latest full backup (pg_dump or base backup)
3. Restore WAL archives up to target time
4. Set `recovery_target_time` in `postgresql.conf`
5. Start PostgreSQL in recovery mode

```bash
# recovery.conf (PostgreSQL 16 uses postgresql.auto.conf)
restore_command = 'wal-e wal-fetch "%f" "%p"'
recovery_target_time = '2025-10-25 14:30:00'
recovery_target_action = 'promote'
```

---

## 8. Monitoring & Observability

### 8.1. Prometheus Exporter

```yaml
# k8s/postgres-exporter.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres-exporter
  namespace: postgres
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres-exporter
  template:
    metadata:
      labels:
        app: postgres-exporter
    spec:
      containers:
      - name: postgres-exporter
        image: prometheuscommunity/postgres-exporter:v0.15.0
        ports:
        - name: metrics
          containerPort: 9187
        env:
        - name: DATA_SOURCE_NAME
          value: "postgresql://dict_admin:$(PGPASSWORD)@postgres-primary:5432/dict_core?sslmode=disable"
        - name: PGPASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secrets
              key: postgres-password
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
---
apiVersion: v1
kind: Service
metadata:
  name: postgres-exporter
  namespace: postgres
  labels:
    app: postgres-exporter
spec:
  type: ClusterIP
  ports:
  - name: metrics
    port: 9187
    targetPort: 9187
  selector:
    app: postgres-exporter
```

### 8.2. Key Metrics

```yaml
# Database connections
pg_stat_database_numbackends
pg_settings_max_connections

# Query performance
pg_stat_statements_total_time
pg_stat_statements_mean_time
pg_stat_statements_calls

# Replication lag
pg_replication_lag_seconds

# Table bloat
pg_table_bloat_bytes

# Cache hit ratio
pg_database_blks_hit / (pg_database_blks_hit + pg_database_blks_read)
```

### 8.3. Alerting Rules

```yaml
# prometheus/postgres-alerts.yaml
groups:
  - name: postgresql
    interval: 30s
    rules:
      - alert: PostgreSQLDown
        expr: up{job="postgres"} == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "PostgreSQL instance is down"

      - alert: PostgreSQLReplicationLag
        expr: pg_replication_lag_seconds > 60
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Replication lag is {{ $value }} seconds (> 60s)"

      - alert: PostgreSQLTooManyConnections
        expr: pg_stat_database_numbackends / pg_settings_max_connections > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Too many connections: {{ $value | humanizePercentage }}"

      - alert: PostgreSQLCacheHitRatioLow
        expr: |
          pg_database_blks_hit / (pg_database_blks_hit + pg_database_blks_read) < 0.9
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Cache hit ratio is low: {{ $value | humanizePercentage }}"
```

---

## 9. Performance Tuning

### 9.1. Memory Configuration

**Shared Buffers**: 2GB (25% of 8GB RAM)
**Effective Cache Size**: 6GB (75% of RAM)
**Work Mem**: 16MB (for sorting, hash joins)
**Maintenance Work Mem**: 512MB (for VACUUM, CREATE INDEX)

### 9.2. Query Optimization

**Enable pg_stat_statements**:

```sql
CREATE EXTENSION pg_stat_statements;

-- Find slowest queries
SELECT
    query,
    calls,
    total_exec_time,
    mean_exec_time,
    max_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;
```

**Create Indexes**:

```sql
-- Based on query patterns
CREATE INDEX CONCURRENTLY idx_entries_lookup ON entries(key_type, key_value, status);
CREATE INDEX CONCURRENTLY idx_claims_active ON claims(status, expires_at) WHERE status IN ('OPEN', 'WAITING_RESOLUTION');
```

### 9.3. Autovacuum Tuning

**Configuration**:

```conf
autovacuum = on
autovacuum_max_workers = 4
autovacuum_naptime = 30s
autovacuum_vacuum_scale_factor = 0.1
autovacuum_analyze_scale_factor = 0.05
```

**Manual VACUUM** (if needed):

```sql
VACUUM ANALYZE entries;
VACUUM ANALYZE claims;
```

---

## 10. High Availability

### 10.1. Failover Strategy

**Manual Failover**:

1. Promote standby to primary: `pg_ctl promote`
2. Update PgBouncer config to point to new primary
3. Reload PgBouncer: `RELOAD;`
4. Reconfigure old primary as new standby

**Automatic Failover** (Future: Patroni):

- Use Patroni for automatic leader election
- Integrates with etcd/Consul for distributed consensus
- Zero-downtime failover

### 10.2. Read Scaling

**Strategy**: Route read queries to standby

**PgBouncer Config**:

```ini
[databases]
dict_core_write = host=postgres-primary ...
dict_core_read = host=postgres-standby ...
```

**Application Code**:

```go
// Use separate connection for reads
readConn := pgx.Connect(ctx, readDSN)
writeConn := pgx.Connect(ctx, writeDSN)
```

---

## Rastreabilidade

### Requisitos Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RF-TSP-004-001 | Primary + Standby deployment | [DAT-001](../../03_Dados/DAT-001_Schema_Database_Core_DICT.md) | ✅ Especificado |
| RF-TSP-004-002 | RLS policies (ISPB isolation) | Security requirement | ✅ Especificado |
| RF-TSP-004-003 | Partitioning (audit logs) | Performance requirement | ✅ Especificado |
| RF-TSP-004-004 | Connection pooling (PgBouncer) | Scalability requirement | ✅ Especificado |
| RF-TSP-004-005 | Backup & PITR | DR requirement | ✅ Especificado |

### Requisitos Não-Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RNF-TSP-004-001 | HA: 99.9% availability | SLA requirement | ✅ Especificado |
| RNF-TSP-004-002 | RPO: < 5 minutes (WAL) | DR requirement | ✅ Especificado |
| RNF-TSP-004-003 | RTO: < 30 minutes (failover) | DR requirement | ✅ Especificado |
| RNF-TSP-004-004 | Backup retention: 30 days | Compliance | ✅ Especificado |

---

## Próximas Revisões

**Pendências**:
- [ ] Implementar Patroni para auto-failover
- [ ] Configurar logical replication para analytics
- [ ] Implementar query caching (pg_query_cache)
- [ ] Criar stored procedures para operações críticas
- [ ] Validar performance em carga real
- [ ] Implementar database sharding (se > 10M entries)

---

**Referências**:
- [DAT-001: Schema Database Core DICT](../../03_Dados/DAT-001_Schema_Database_Core_DICT.md)
- [DAT-002: Schema Database Connect](../../03_Dados/DAT-002_Schema_Database_Connect.md)
- [PostgreSQL 16 Documentation](https://www.postgresql.org/docs/16/)
- [PgBouncer Documentation](https://www.pgbouncer.org/config.html)
- [WAL-E Documentation](https://github.com/wal-e/wal-e)
