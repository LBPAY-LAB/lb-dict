# ADR-005: Escolha de Database - PostgreSQL

**Status**: ✅ Aceito
**Data**: 2025-10-24
**Decisores**: Thiago Lima (Head de Arquitetura), José Luís Silva (CTO)
**Contexto Técnico**: Projeto DICT - LBPay

---

## Controle de Versão

| Versão | Data | Autor | Descrição das Mudanças |
|--------|------|-------|------------------------|
| 1.0 | 2025-10-24 | ARCHITECT | Versão inicial - Documentação da decisão de usar PostgreSQL como banco de dados principal |

---

## Status

**✅ ACEITO** - PostgreSQL já é tecnologia confirmada e em uso no LBPay

---

## Contexto

O projeto DICT da LBPay requer um **banco de dados relacional** robusto para armazenar dados críticos com garantias ACID (Atomicity, Consistency, Isolation, Durability). O sistema precisa gerenciar:

### Requisitos Funcionais

1. **Entidades Principais**:
   - **Chaves PIX** (`dict_keys`): Armazenar chaves cadastradas (CPF, CNPJ, Email, Telefone, EVP)
   - **Claims** (`claims`): Reivindicações recebidas/enviadas
   - **Portabilidades** (`portabilities`): Solicitações de portabilidade
   - **Auditoria** (`audit_logs`): Logs de todas as operações (5 anos de retenção)
   - **VSYNC** (`vsync_reports`): Relatórios de sincronização

2. **Relacionamentos**:
   - 1 Conta → N Chaves PIX
   - 1 Chave PIX → N Claims (histórico)
   - 1 Chave PIX → N Portabilidades (histórico)

3. **Queries Complexas**:
   - Listar chaves por conta (com filtros: tipo, status)
   - Buscar claim por chave + status
   - Relatórios de auditoria (filtrar por período, tipo de operação, usuário)

4. **Transações ACID**:
   - Criar chave + publicar evento (atomicidade)
   - Atualizar status claim + notificar usuário (consistência)

### Requisitos Não-Funcionais

| ID | Requisito | Target | Fonte |
|----|-----------|--------|-------|
| **NFR-120** | Durabilidade | 100% (ACID compliance) | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-121** | Latência (write) | ≤ 10ms (P95) | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-122** | Latência (read) | ≤ 5ms (P95) | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-123** | Throughput | ≥ 10.000 writes/sec | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-124** | Disponibilidade | ≥ 99.99% | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-125** | Backup/Recovery | RPO ≤ 5 min, RTO ≤ 1 hora | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-126** | Auditoria | Retenção 5 anos (Lei 12.865/2013) | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |

### Contexto Organizacional

- **LBPay já utiliza PostgreSQL** em múltiplos serviços (Core Banking, Money Moving, Auth)
- Equipe de backend possui expertise em PostgreSQL
- DBAs especializados em tuning e operação
- Infraestrutura provisionada (PostgreSQL clusters)
- Redução de custo operacional (não introduzir nova tecnologia)

---

## Decisão

**Escolhemos PostgreSQL como banco de dados relacional principal para o projeto DICT.**

### Justificativa

PostgreSQL foi escolhido pelos seguintes motivos:

#### 1. **Já em Uso no LBPay**

✅ **PostgreSQL já é tecnologia estabelecida no LBPay**:
- Utilizado em Core Banking, Money Moving, Auth Service, etc.
- Clusters PostgreSQL provisionados e operacionais
- Equipe treinada (DBAs, backend devs)
- **Menor Time-to-Market** (não precisa provisionar nova stack)
- **Menor risco operacional** (tecnologia conhecida)
- **Consistência tecnológica** (mesma stack entre projetos)

#### 2. **ACID Compliance**

**PostgreSQL garante ACID**:

| Propriedade | Descrição | Importância para DICT |
|-------------|-----------|----------------------|
| **Atomicity** | Transação completa ou nada | Criar chave + publicar evento (tudo ou nada) ✅ |
| **Consistency** | Dados sempre em estado válido | Status de chave consistente (não PENDING e ACTIVE simultaneamente) ✅ |
| **Isolation** | Transações concorrentes não interferem | Múltiplos workers atualizando claims (serializable isolation) ✅ |
| **Durability** | Dados persistidos em disco (não voláteis) | Chaves PIX nunca perdidas (compliance Bacen) ✅ |

**Exemplo Transação**:
```sql
BEGIN;
    -- 1. Criar chave
    INSERT INTO dict_keys (key_id, key_type, key_value, status)
    VALUES ('key_123', 'CPF', '12345678901', 'PENDING');

    -- 2. Registrar auditoria
    INSERT INTO audit_logs (event_type, key_id, user_id, timestamp)
    VALUES ('KEY_REGISTER_REQUESTED', 'key_123', 'user_456', NOW());

    -- Se qualquer operação falhar, ROLLBACK automático
COMMIT;
```

#### 3. **Funcionalidades Avançadas**

**PostgreSQL vs Outros Bancos Relacionais**:

| Feature | PostgreSQL | MySQL | SQL Server | Oracle |
|---------|------------|-------|------------|--------|
| **ACID Compliance** | ✅ **Full** | ⚠️ Depende de engine (InnoDB ok) | ✅ Full | ✅ Full |
| **JSON Support** | ✅ **JSONB** (indexável) | ⚠️ JSON (não indexável) | ✅ JSON | ⚠️ Limited |
| **Full Text Search** | ✅ **Built-in** (tsvector) | ⚠️ Limited | ✅ Built-in | ✅ Built-in |
| **Window Functions** | ✅ **Advanced** | ⚠️ Basic | ✅ Advanced | ✅ Advanced |
| **Partial Indexes** | ✅ **Sim** | ❌ Não | ❌ Não | ❌ Não |
| **CTE (Common Table Expressions)** | ✅ **Recursive** | ⚠️ Non-recursive | ✅ Recursive | ✅ Recursive |
| **Foreign Data Wrappers** | ✅ **Sim** | ❌ Não | ❌ Não | ⚠️ DB Links |
| **Extensibilidade** | ✅ **Extensions** (PostGIS, pg_cron, etc.) | ❌ Limited | ❌ Limited | ⚠️ Packages |
| **Open Source** | ✅ **Sim** | ✅ Sim | ❌ Não | ❌ Não |
| **Custo** | ✅ **Gratuito** | ✅ Gratuito | ❌ Licenciamento | ❌ Licenciamento caro |

#### 4. **JSONB para Flexibilidade**

**Uso de JSONB no DICT**:

Armazenar metadados variáveis sem alterar schema:

```sql
CREATE TABLE dict_keys (
    key_id UUID PRIMARY KEY,
    key_type VARCHAR(10) NOT NULL,
    key_value VARCHAR(77) NOT NULL,
    status VARCHAR(20) NOT NULL,
    account_id UUID NOT NULL,
    metadata JSONB,  -- Metadados flexíveis
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Exemplo: armazenar dados específicos por tipo de chave
INSERT INTO dict_keys (key_id, key_type, key_value, status, account_id, metadata)
VALUES (
    'key_123',
    'EMAIL',
    'user@example.com',
    'ACTIVE',
    'acc_456',
    '{"otp_validated_at": "2025-10-24T10:30:00Z", "email_provider": "gmail.com"}'::jsonb
);

-- Query com JSONB (indexável)
SELECT * FROM dict_keys
WHERE metadata->>'email_provider' = 'gmail.com';
```

**Vantagens JSONB**:
- ✅ **Flexibilidade**: Adicionar campos sem ALTER TABLE
- ✅ **Indexável**: GIN indexes para queries rápidas
- ✅ **Operadores**: `->`, `->>`, `@>`, `?`, `?|`, `?&`
- ✅ **Performance**: Binário (não texto como JSON)

#### 5. **Partial Indexes (Performance)**

**Índices Parciais para Queries Específicas**:

```sql
-- Índice apenas para chaves ACTIVE (ignora DELETED, PENDING)
CREATE INDEX idx_dict_keys_active
ON dict_keys (key_value)
WHERE status = 'ACTIVE';

-- Índice para claims pendentes (ignora CONFIRMED, CANCELLED)
CREATE INDEX idx_claims_pending
ON claims (key_id, created_at)
WHERE status = 'PENDING';
```

**Vantagens**:
- ✅ **Menor tamanho**: Índice apenas subset dos dados (economiza espaço/memória)
- ✅ **Queries rápidas**: Menos dados a escanear
- ✅ **Único no PostgreSQL**: MySQL/SQL Server não suportam

#### 6. **High Availability (Replicação)**

**Topologia PostgreSQL para Projeto DICT**:

```
PostgreSQL Cluster (LBPay Production)
│
├── Primary (Master)
│   ├── Write operations
│   └── Synchronous replication to Replica 1
│
├── Replica 1 (Sync)
│   ├── Read operations (load balancing)
│   └── Failover candidate (automatic promotion)
│
├── Replica 2 (Async)
│   ├── Read operations (load balancing)
│   └── Backup candidate
│
└── Backup (WAL archiving to S3)
    └── Point-in-time recovery (PITR)
```

**Características**:
- ✅ **Synchronous replication**: Dados confirmados em ≥ 2 nós antes de commit (zero data loss)
- ✅ **Failover automático**: Patroni/Stolon promove replica se primary falha
- ✅ **Read scaling**: Replicas atendem queries read-only (load balancing)
- ✅ **Backup contínuo**: WAL (Write-Ahead Log) arquivado em S3 (PITR)

#### 7. **Partitioning (Auditoria)**

**Particionar `audit_logs` por tempo** (retenção 5 anos):

```sql
-- Tabela particionada por mês
CREATE TABLE audit_logs (
    log_id BIGSERIAL,
    event_type VARCHAR(50) NOT NULL,
    key_id UUID,
    user_id UUID,
    timestamp TIMESTAMPTZ NOT NULL,
    payload JSONB
) PARTITION BY RANGE (timestamp);

-- Partições por mês
CREATE TABLE audit_logs_2025_01 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE audit_logs_2025_02 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

-- ... (criar partições automaticamente via pg_cron)

-- Query (PostgreSQL escolhe partição automaticamente)
SELECT * FROM audit_logs
WHERE timestamp >= '2025-01-15' AND timestamp < '2025-01-20';
```

**Vantagens**:
- ✅ **Performance**: Query escaneia apenas partições relevantes (não tabela inteira)
- ✅ **Manutenção**: Drop partições antigas (rápido, sem bloqueio)
- ✅ **Backup**: Backup partições antigas para S3 (tiered storage)

#### 8. **Full-Text Search (Auditoria/Logs)**

**Buscar eventos de auditoria por texto**:

```sql
-- Adicionar coluna tsvector (search vector)
ALTER TABLE audit_logs ADD COLUMN search_vector tsvector;

-- Popular search_vector (trigger automático)
CREATE TRIGGER audit_logs_search_update
BEFORE INSERT OR UPDATE ON audit_logs
FOR EACH ROW EXECUTE FUNCTION
tsvector_update_trigger(search_vector, 'pg_catalog.portuguese', payload);

-- Criar índice GIN (full-text search)
CREATE INDEX idx_audit_logs_search ON audit_logs USING GIN(search_vector);

-- Query full-text search
SELECT * FROM audit_logs
WHERE search_vector @@ to_tsquery('portuguese', 'cadastro & chave');
```

**Vantagens**:
- ✅ **Built-in**: Não precisa Elasticsearch (simples)
- ✅ **Performance**: GIN index (busca em ms)
- ✅ **Idiomas**: Suporte português (stemming, stop words)

#### 9. **Extensions Úteis**

**PostgreSQL Extensions para DICT**:

| Extension | Uso | Benefício |
|-----------|-----|-----------|
| **uuid-ossp** | Gerar UUIDs (v4) | IDs únicos para chaves, claims, etc. |
| **pg_trgm** | Fuzzy matching (trigrams) | Buscar chaves similar (typo tolerance) |
| **pgcrypto** | Criptografia | Encriptar dados sensíveis (PII) |
| **pg_stat_statements** | Análise de queries | Identificar queries lentas (tuning) |
| **pg_cron** | Cron jobs no DB | Particionar audit_logs automaticamente |

**Exemplo `uuid-ossp`**:
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Gerar UUID v4
INSERT INTO dict_keys (key_id, ...)
VALUES (uuid_generate_v4(), ...);
```

---

## Consequências

### Positivas ✅

1. **Time-to-Market Reduzido**:
   - PostgreSQL já usado no LBPay
   - Equipe já treinada (DBAs, devs)
   - Infraestrutura provisionada

2. **ACID Compliance**:
   - Durabilidade 100% (dados nunca perdidos)
   - Transações atômicas (tudo ou nada)
   - Isolamento (queries concorrentes seguras)

3. **Performance**:
   - Latência: P95 < 10ms (writes), < 5ms (reads)
   - Throughput: 10k+ writes/sec (single instance)
   - Partial indexes, JSONB indexes

4. **Funcionalidades Avançadas**:
   - JSONB (flexibilidade)
   - Partial indexes (performance)
   - Full-text search (auditoria)
   - Partitioning (retenção 5 anos)
   - Extensions (uuid, trigrams, cron)

5. **High Availability**:
   - Synchronous replication (zero data loss)
   - Failover automático (Patroni)
   - Read scaling (replicas)

6. **Open Source**:
   - Gratuito (zero custo de licenciamento)
   - Comunidade ativa
   - Extensibilidade

### Negativas ❌

1. **Complexidade Operacional**:
   - Tuning (vacuum, autovacuum, indexes)
   - Monitoramento (pg_stat_statements, logs)
   - **Mitigação**: DBAs especializados no LBPay

2. **Vertical Scaling Limits**:
   - Single primary (writes não distribuídos)
   - **Mitigação**: Sharding futuro (se necessário), cache Redis

3. **Vacuum Overhead**:
   - MVCC (Multi-Version Concurrency Control) gera dead tuples
   - Vacuum periódico necessário (pode causar I/O spike)
   - **Mitigação**: Autovacuum tuning, vacuum durante off-peak

4. **Locks em Migrações**:
   - ALTER TABLE em tabelas grandes = lock (downtime)
   - **Mitigação**: Zero-downtime migrations (adicionar coluna nullable, backfill assíncrono)

### Riscos e Mitigações

| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| **Primary failure** | Baixa | Alto | Synchronous replication, failover automático (Patroni) |
| **Performance degradation** | Média | Médio | Monitoring (pg_stat_statements), tuning, cache Redis |
| **Disk space exhaustion** | Média | Alto | Monitoring, alertas, partitioning (drop partições antigas) |
| **Slow queries** | Média | Médio | Indexes, EXPLAIN ANALYZE, query optimization |

---

## Alternativas Consideradas

### Alternativa 1: MySQL

**Prós**:
- ✅ Open-source (gratuito)
- ✅ Performance boa (InnoDB)
- ✅ Amplamente usado

**Contras**:
- ❌ **Não usado no LBPay** (introduziria nova stack)
- ❌ **JSONB limitado** (JSON não indexável eficientemente)
- ❌ **Sem partial indexes**
- ❌ **Full-text search limitado**
- ❌ **Funcionalidades inferiores** (window functions, CTEs recursivos)

**Decisão**: ❌ **Rejeitado** - PostgreSQL já em uso, funcionalidades superiores

### Alternativa 2: MongoDB (NoSQL)

**Prós**:
- ✅ Schema-less (flexibilidade)
- ✅ JSON nativo
- ✅ Horizontal scaling (sharding built-in)

**Contras**:
- ❌ **Não usado no LBPay** (dados relacionais)
- ❌ **Sem ACID multi-document** (limitado até v4.0)
- ❌ **Joins complexos** (não adequado para dados relacionais)
- ❌ **Queries relacionais difíceis** (ex: listar chaves com claims)
- ❌ **Overhead de aprendizado** (agregação pipelines)

**Decisão**: ❌ **Rejeitado** - Dados DICT são relacionais (conta ← chave → claim)

### Alternativa 3: DynamoDB (AWS NoSQL)

**Prós**:
- ✅ Managed service (zero ops)
- ✅ Escalabilidade automática
- ✅ Performance previsível

**Contras**:
- ❌ **Vendor lock-in** (AWS-only)
- ❌ **Não usado no LBPay**
- ❌ **Sem transações multi-table** (limitado)
- ❌ **Queries relacionais complexas** (não suporta joins)
- ❌ **Custos variáveis** (pay-per-request)

**Decisão**: ❌ **Rejeitado** - Lock-in, dados relacionais, PostgreSQL já em uso

### Alternativa 4: CockroachDB (Distributed SQL)

**Prós**:
- ✅ PostgreSQL-compatible
- ✅ Distributed (horizontal scaling)
- ✅ Multi-region (geo-replication)

**Contras**:
- ❌ **Não usado no LBPay**
- ❌ **Complexidade operacional** (cluster Raft consensus)
- ❌ **Overhead de aprendizado** (tuning distribuído)
- ❌ **Custos** (Enterprise features)
- ❌ **Overkill** (DICT não precisa multi-region)

**Decisão**: ❌ **Rejeitado** - Overkill, PostgreSQL suficiente

---

## Implementação

### Schema PostgreSQL

#### Tabela: `dict_keys`

```sql
CREATE TABLE dict_keys (
    key_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_type VARCHAR(10) NOT NULL CHECK (key_type IN ('CPF', 'CNPJ', 'EMAIL', 'PHONE', 'EVP')),
    key_value VARCHAR(77) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('PENDING', 'ACTIVE', 'FAILED', 'DELETED')),
    account_id UUID NOT NULL,
    ispb VARCHAR(8) NOT NULL,
    branch VARCHAR(4) NOT NULL,
    account_number VARCHAR(20) NOT NULL,
    account_type VARCHAR(4) NOT NULL,
    owner_type VARCHAR(20) NOT NULL,
    owner_tax_id VARCHAR(14) NOT NULL,
    owner_name VARCHAR(255) NOT NULL,
    bacen_entry_id VARCHAR(50),  -- ID retornado pelo Bacen
    metadata JSONB,  -- Metadados flexíveis
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ  -- Soft delete
);

-- Índices
CREATE UNIQUE INDEX idx_dict_keys_value_active
ON dict_keys (key_value)
WHERE status = 'ACTIVE' AND deleted_at IS NULL;

CREATE INDEX idx_dict_keys_account
ON dict_keys (account_id, status);

CREATE INDEX idx_dict_keys_owner
ON dict_keys (owner_tax_id, key_type);

CREATE INDEX idx_dict_keys_metadata
ON dict_keys USING GIN (metadata);
```

#### Tabela: `claims`

```sql
CREATE TABLE claims (
    claim_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_id UUID NOT NULL REFERENCES dict_keys(key_id),
    claim_type VARCHAR(10) NOT NULL CHECK (claim_type IN ('INCOMING', 'OUTGOING')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('PENDING', 'CONFIRMED', 'CANCELLED', 'AUTO_CONFIRMED')),
    claimer_ispb VARCHAR(8),  -- ISPB do PSP que reivindica (incoming)
    claimed_ispb VARCHAR(8),  -- ISPB do PSP reivindicado (outgoing)
    bacen_claim_id VARCHAR(50),
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deadline_at TIMESTAMPTZ NOT NULL,  -- 7 dias corridos
    resolved_at TIMESTAMPTZ,
    resolution_reason VARCHAR(50),
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Índices
CREATE INDEX idx_claims_key_status
ON claims (key_id, status);

CREATE INDEX idx_claims_pending
ON claims (deadline_at)
WHERE status = 'PENDING';
```

#### Tabela: `portabilities`

```sql
CREATE TABLE portabilities (
    portability_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_id UUID NOT NULL REFERENCES dict_keys(key_id),
    portability_type VARCHAR(10) NOT NULL CHECK (portability_type IN ('INCOMING', 'OUTGOING')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('PENDING', 'CONFIRMED', 'CANCELLED')),
    source_ispb VARCHAR(8),
    target_ispb VARCHAR(8),
    bacen_portability_id VARCHAR(50),
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deadline_at TIMESTAMPTZ NOT NULL,  -- 7 dias corridos
    resolved_at TIMESTAMPTZ,
    resolution_reason VARCHAR(50),
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Índices
CREATE INDEX idx_portabilities_key_status
ON portabilities (key_id, status);
```

#### Tabela: `audit_logs` (Particionada)

```sql
CREATE TABLE audit_logs (
    log_id BIGSERIAL,
    event_type VARCHAR(50) NOT NULL,
    entity_type VARCHAR(20),  -- KEY, CLAIM, PORTABILITY
    entity_id UUID,
    user_id UUID,
    correlation_id VARCHAR(50),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payload JSONB,
    search_vector tsvector,  -- Full-text search
    PRIMARY KEY (log_id, timestamp)
) PARTITION BY RANGE (timestamp);

-- Criar partições (via pg_cron ou script)
CREATE TABLE audit_logs_2025_01 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

-- Índices
CREATE INDEX idx_audit_logs_entity
ON audit_logs (entity_type, entity_id, timestamp);

CREATE INDEX idx_audit_logs_search
ON audit_logs USING GIN(search_vector);
```

#### Tabela: `vsync_reports`

```sql
CREATE TABLE vsync_reports (
    vsync_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    local_hash VARCHAR(32) NOT NULL,  -- MD5 hash
    bacen_hash VARCHAR(32) NOT NULL,
    match BOOLEAN NOT NULL,
    discrepancies JSONB,  -- Se não match: {missing: [...], extra: [...]}
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vsync_executed_at ON vsync_reports (executed_at DESC);
```

### Migrations (Zero-Downtime)

**Ferramenta**: `golang-migrate`

```bash
# Criar migration
migrate create -ext sql -dir db/migrations -seq create_dict_keys_table

# Aplicar migrations
migrate -path db/migrations -database "postgresql://localhost:5432/dict?sslmode=disable" up

# Rollback
migrate -path db/migrations -database "postgresql://localhost:5432/dict?sslmode=disable" down 1
```

**Exemplo Migration**:
```sql
-- 0001_create_dict_keys_table.up.sql
CREATE TABLE dict_keys (...);
CREATE INDEX ...;

-- 0001_create_dict_keys_table.down.sql
DROP TABLE dict_keys CASCADE;
```

### Connection Pool (Go)

```go
package database

import (
    "database/sql"
    _ "github.com/lib/pq"
)

func NewPostgresDB() (*sql.DB, error) {
    connStr := "host=postgres.lbpay.svc.cluster.local port=5432 user=dict_user password=*** dbname=dict_db sslmode=require"

    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }

    // Connection pool settings
    db.SetMaxOpenConns(100)      // Max conexões abertas
    db.SetMaxIdleConns(10)       // Max conexões idle
    db.SetConnMaxLifetime(time.Hour) // Reciclar conexões após 1 hora

    return db, nil
}
```

### Queries (SQLC Code Generation)

**Ferramenta**: `sqlc` (gera código Go type-safe a partir de SQL)

```yaml
# sqlc.yaml
version: "2"
sql:
  - schema: "db/schema.sql"
    queries: "db/queries"
    engine: "postgresql"
    gen:
      go:
        package: "db"
        out: "internal/db"
```

**Query SQL**:
```sql
-- db/queries/keys.sql

-- name: GetKeyByValue :one
SELECT * FROM dict_keys
WHERE key_value = $1 AND status = 'ACTIVE' AND deleted_at IS NULL
LIMIT 1;

-- name: ListKeysByAccount :many
SELECT * FROM dict_keys
WHERE account_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CreateKey :one
INSERT INTO dict_keys (key_type, key_value, status, account_id, ...)
VALUES ($1, $2, $3, $4, ...)
RETURNING *;

-- name: UpdateKeyStatus :exec
UPDATE dict_keys
SET status = $2, updated_at = NOW()
WHERE key_id = $1;
```

**Código Gerado** (type-safe):
```go
// internal/db/keys.sql.go (gerado por sqlc)
func (q *Queries) GetKeyByValue(ctx context.Context, keyValue string) (DictKey, error) {
    // Implementação gerada automaticamente
}

// Uso
key, err := queries.GetKeyByValue(ctx, "12345678901")
```

### Monitoramento

**Métricas Prometheus**:
- `pg_up` (database disponível)
- `pg_stat_database_*` (conexões, transações, commits, rollbacks)
- `pg_stat_statements_*` (queries lentas, execuções)
- `pg_replication_lag` (lag das replicas)

**Queries Lentas** (pg_stat_statements):
```sql
SELECT query, mean_exec_time, calls
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;
```

**Alertas**:
- Connection pool exhaustion (max_connections - active < 10)
- Slow queries (P95 > 100ms)
- Replication lag > 5 seconds
- Disk usage > 80%

---

## Rastreabilidade

### Requisitos Funcionais Impactados

| CRF | Descrição | Tabela PostgreSQL |
|-----|-----------|-------------------|
| [CRF-001](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-001) | Cadastrar Chave | `dict_keys` |
| [CRF-020](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-020) | Solicitar Claim | `claims` |
| [CRF-030](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-030) | Solicitar Portabilidade | `portabilities` |
| [CRF-080](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-080) | Auditoria | `audit_logs` (particionada) |
| [CRF-060](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-060) | VSYNC | `vsync_reports` |

### NFRs Impactados

| NFR | Descrição | Como PostgreSQL Atende |
|-----|-----------|------------------------|
| [NFR-120](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-120) | Durabilidade 100% | ACID compliance, WAL, replicação ✅ |
| [NFR-121](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-121) | Latência write ≤ 10ms | P95 < 10ms (benchmark) ✅ |
| [NFR-124](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-124) | Disponibilidade 99.99% | Replicação, failover automático ✅ |
| [NFR-126](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-126) | Retenção 5 anos | Partitioning, S3 archive ✅ |

---

## Referências

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [PostgreSQL Performance Tuning](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [SQLC (Code Generation)](https://docs.sqlc.dev/)
- [ArquiteturaDict_LBPAY.md](../../Docs_iniciais/ArquiteturaDict_LBPAY.md): Diagramas SVG mostrando PostgreSQL

---

## Aprovação

- [x] **Thiago Lima** (Head de Arquitetura) - 2025-10-24
- [x] **José Luís Silva** (CTO) - 2025-10-24

**Rationale**: PostgreSQL já é tecnologia confirmada e em uso no LBPay. Esta ADR documenta a decisão e fundamenta o uso técnico no projeto DICT.

---

**FIM DO DOCUMENTO ADR-005**
