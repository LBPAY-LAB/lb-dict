# SPECS-DATABASE.md - Database Schema & Repositories

## 📋 Overview

Especificação completa do schema PostgreSQL, migrations, repositories e queries para o sistema de monitoramento de Rate Limit do DICT BACEN.

**Responsável**: DB & Domain Engineer
**Versão**: 1.0.0
**Última Atualização**: 2025-10-31

---

## 🎯 Objetivos

1. **Armazenar configurações** de políticas de rate limit conforme BACEN Cap. 19
2. **Registrar histórico** de estados dos baldes (time-series data)
3. **Manter log de alertas** disparados (audit trail)
4. **Performance otimizada** para queries de análise temporal
5. **Escalabilidade** via partitioning (suportar anos de histórico)

---

## 🗄️ Database Schema

### Diagrama ER

```
┌─────────────────────────────────┐
│  dict_rate_limit_policies       │
│  (Configuração)                 │
├─────────────────────────────────┤
│ PK id                           │
│ UK policy_name                  │
│    description                  │
│    capacity_max                 │
│    refill_tokens                │
│    refill_period_sec            │
│    warning_threshold_pct        │
│    critical_threshold_pct       │
│    enabled                      │
│    created_at                   │
│    updated_at                   │
└────────┬────────────────────────┘
         │
         │ 1:N
         │
┌────────▼────────────────────────┐
│  dict_rate_limit_states         │
│  (Histórico - Particionado)     │
├─────────────────────────────────┤
│ PK id                           │
│ FK policy_name                  │
│    available_tokens             │
│    capacity                     │
│    refill_tokens                │
│    refill_period_sec            │
│    utilization_pct              │
│    category (A-H)               │
│    checked_at                   │
│    created_at                   │
└────────┬────────────────────────┘
         │
         │ 1:N
         │
┌────────▼────────────────────────┐
│  dict_rate_limit_alerts         │
│  (Log de Alertas)               │
├─────────────────────────────────┤
│ PK id                           │
│ FK policy_name                  │
│    severity (WARNING/CRITICAL)  │
│    available_tokens             │
│    capacity                     │
│    utilization_pct              │
│    message                      │
│    resolved                     │
│    resolved_at                  │
│    resolved_by                  │
│    created_at                   │
└─────────────────────────────────┘
```

---

## 📝 Tabela 1: `dict_rate_limit_policies`

### Descrição
Armazena a configuração de todas as políticas de rate limit conforme BACEN Manual Cap. 19. Seed inicial com 24 políticas padrão.

### DDL

```sql
CREATE TABLE dict_rate_limit_policies (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_name             VARCHAR(100) NOT NULL UNIQUE,
    description             TEXT,
    capacity_max            INTEGER NOT NULL CHECK (capacity_max > 0),
    refill_tokens           INTEGER NOT NULL CHECK (refill_tokens > 0),
    refill_period_sec       INTEGER NOT NULL CHECK (refill_period_sec > 0),
    warning_threshold_pct   DECIMAL(5,2) NOT NULL DEFAULT 25.00
        CHECK (warning_threshold_pct > 0 AND warning_threshold_pct <= 100),
    critical_threshold_pct  DECIMAL(5,2) NOT NULL DEFAULT 10.00
        CHECK (critical_threshold_pct > 0 AND critical_threshold_pct <= 100),
    enabled                 BOOLEAN NOT NULL DEFAULT true,
    created_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT valid_thresholds CHECK (critical_threshold_pct < warning_threshold_pct)
);

-- Indexes
CREATE INDEX idx_policies_enabled ON dict_rate_limit_policies(enabled)
    WHERE enabled = true;

CREATE INDEX idx_policies_name_enabled ON dict_rate_limit_policies(policy_name)
    WHERE enabled = true;

-- Comments
COMMENT ON TABLE dict_rate_limit_policies IS
    'Configuração de políticas de rate limit do DICT BACEN (Cap. 19)';
COMMENT ON COLUMN dict_rate_limit_policies.policy_name IS
    'Nome da política conforme BACEN (ex: ENTRIES_WRITE, CLAIMS_READ)';
COMMENT ON COLUMN dict_rate_limit_policies.capacity_max IS
    'Capacidade máxima do balde (tokens) conforme categoria do PSP';
COMMENT ON COLUMN dict_rate_limit_policies.refill_tokens IS
    'Quantidade de tokens repostos por período';
COMMENT ON COLUMN dict_rate_limit_policies.refill_period_sec IS
    'Período de reposição em segundos (geralmente 60s = 1min)';
COMMENT ON COLUMN dict_rate_limit_policies.warning_threshold_pct IS
    'Threshold para alerta WARNING (% restante do balde, default 25%)';
COMMENT ON COLUMN dict_rate_limit_policies.critical_threshold_pct IS
    'Threshold para alerta CRITICAL (% restante do balde, default 10%)';
```

### Seed Data (24 políticas BACEN)

```sql
INSERT INTO dict_rate_limit_policies (policy_name, description, capacity_max, refill_tokens, refill_period_sec) VALUES
-- Entries (Vínculos)
('ENTRIES_WRITE',            'Criar/deletar vínculo',                           36000, 1200, 60),
('ENTRIES_UPDATE',           'Atualizar vínculo',                                 600,  600, 60),
('ENTRIES_READ_PARTICIPANT_ANTISCAN', 'Consultar vínculo (PSP)',               50000, 25000, 60), -- Categoria A (max)
('ENTRIES_STATISTICS_READ',  'Consultar estatísticas de vínculos',             36000, 12000, 60),

-- Claims (Reivindicações)
('CLAIMS_READ',              'Consultar reivindicação',                         18000,  600, 60),
('CLAIMS_WRITE',             'Criar/atualizar reivindicação',                   36000, 1200, 60),
('CLAIMS_LIST_WITH_ROLE',    'Listar reivindicações (com filtro doador/reiv)',    200,   40, 60),
('CLAIMS_LIST_WITHOUT_ROLE', 'Listar reivindicações (sem filtro)',                 50,   10, 60),

-- Sync/CIDs
('SYNC_VERIFICATIONS_WRITE', 'Criar verificação de VSync',                         50,   10, 60),
('CIDS_FILES_WRITE',         'Criar arquivo de CID (async)',                      200,   40, 86400), -- 40/dia
('CIDS_FILES_READ',          'Consultar arquivo de CID',                           50,   10, 60),
('CIDS_EVENTS_LIST',         'Listar eventos de CID',                             100,   20, 60),
('CIDS_ENTRIES_READ',        'Consultar vínculo por CID',                       36000, 1200, 60),

-- Infraction Reports (Infrações)
('INFRACTION_REPORTS_READ',  'Consultar infração',                              18000,  600, 60),
('INFRACTION_REPORTS_WRITE', 'Criar/atualizar infração',                        36000, 1200, 60),
('INFRACTION_REPORTS_LIST_WITH_ROLE',    'Listar infrações (com filtro)',        200,   40, 60),
('INFRACTION_REPORTS_LIST_WITHOUT_ROLE', 'Listar infrações (sem filtro)',         50,   10, 60),

-- Misc
('KEYS_CHECK',               'Verificar existência de chaves',                     70,   70, 60),

-- Refunds (Devoluções)
('REFUNDS_READ',             'Consultar devolução',                             36000, 1200, 60),
('REFUNDS_WRITE',            'Criar/atualizar devolução',                       72000, 2400, 60),
('REFUND_LIST_WITH_ROLE',    'Listar devoluções (com filtro)',                    200,   40, 60),
('REFUND_LIST_WITHOUT_ROLE', 'Listar devoluções (sem filtro)',                     50,   10, 60),

-- Fraud Markers (Marcadores de Fraude)
('FRAUD_MARKERS_READ',       'Consultar marcador de fraude',                    18000,  600, 60),
('FRAUD_MARKERS_WRITE',      'Criar/cancelar marcador de fraude',               36000, 1200, 60),
('FRAUD_MARKERS_LIST',       'Listar marcadores de fraude',                     18000,  600, 60),

-- Statistics & Policies
('PERSONS_STATISTICS_READ',  'Consultar estatísticas de pessoa',                36000, 12000, 60),
('POLICIES_READ',            'Consultar política de limitação (esta API)',        200,   60, 60),
('POLICIES_LIST',            'Listar políticas de limitação (esta API)',           20,    6, 60);
```

### Queries Típicas

```sql
-- Buscar política por nome
SELECT * FROM dict_rate_limit_policies
WHERE policy_name = $1 AND enabled = true;

-- Listar todas políticas ativas
SELECT policy_name, capacity_max, refill_tokens, warning_threshold_pct, critical_threshold_pct
FROM dict_rate_limit_policies
WHERE enabled = true
ORDER BY policy_name;

-- Buscar políticas com thresholds customizados
SELECT * FROM dict_rate_limit_policies
WHERE warning_threshold_pct != 25.00 OR critical_threshold_pct != 10.00;
```

---

## 📝 Tabela 2: `dict_rate_limit_states`

### Descrição
Histórico time-series de estados dos baldes. **Particionado por mês** para performance e gerenciamento de retenção de dados.

### DDL

```sql
CREATE TABLE dict_rate_limit_states (
    id                  BIGSERIAL,
    policy_name         VARCHAR(100) NOT NULL,
    available_tokens    INTEGER NOT NULL CHECK (available_tokens >= 0),
    capacity            INTEGER NOT NULL CHECK (capacity > 0),
    refill_tokens       INTEGER NOT NULL CHECK (refill_tokens > 0),
    refill_period_sec   INTEGER NOT NULL CHECK (refill_period_sec > 0),
    utilization_pct     DECIMAL(5,2) NOT NULL CHECK (utilization_pct >= 0 AND utilization_pct <= 100),
    category            VARCHAR(1) CHECK (category IN ('A', 'B', 'C', 'D', 'E', 'F', 'G', 'H')),
    checked_at          TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id, checked_at),
    FOREIGN KEY (policy_name) REFERENCES dict_rate_limit_policies(policy_name) ON DELETE CASCADE
) PARTITION BY RANGE (checked_at);

-- Indexes (aplicados a cada partição)
CREATE INDEX idx_states_policy_checked ON dict_rate_limit_states(policy_name, checked_at DESC);
CREATE INDEX idx_states_utilization_high ON dict_rate_limit_states(utilization_pct)
    WHERE utilization_pct > 75.0;
CREATE INDEX idx_states_checked_at ON dict_rate_limit_states(checked_at DESC);

-- Comments
COMMENT ON TABLE dict_rate_limit_states IS
    'Histórico de estados dos baldes (time-series, particionado por mês)';
COMMENT ON COLUMN dict_rate_limit_states.available_tokens IS
    'Tokens disponíveis no momento da consulta ao DICT';
COMMENT ON COLUMN dict_rate_limit_states.capacity IS
    'Capacidade total do balde no momento (pode variar por categoria)';
COMMENT ON COLUMN dict_rate_limit_states.utilization_pct IS
    'Percentual de utilização: (1 - available/capacity) * 100';
COMMENT ON COLUMN dict_rate_limit_states.category IS
    'Categoria do PSP (A-H) - apenas para ENTRIES_READ_PARTICIPANT_ANTISCAN';
COMMENT ON COLUMN dict_rate_limit_states.checked_at IS
    'Timestamp da consulta ao DICT BACEN (ResponseTime)';
```

### Partitions (criar via migration script)

```sql
-- Script para criar partições mensais automaticamente
-- Executar via migration ou cron job

-- Partição atual (Novembro 2025)
CREATE TABLE IF NOT EXISTS dict_rate_limit_states_2025_11
    PARTITION OF dict_rate_limit_states
    FOR VALUES FROM ('2025-11-01 00:00:00+00') TO ('2025-12-01 00:00:00+00');

-- Partição próximo mês (Dezembro 2025)
CREATE TABLE IF NOT EXISTS dict_rate_limit_states_2025_12
    PARTITION OF dict_rate_limit_states
    FOR VALUES FROM ('2025-12-01 00:00:00+00') TO ('2026-01-01 00:00:00+00');

-- Partição mês seguinte (Janeiro 2026)
CREATE TABLE IF NOT EXISTS dict_rate_limit_states_2026_01
    PARTITION OF dict_rate_limit_states
    FOR VALUES FROM ('2026-01-01 00:00:00+00') TO ('2026-02-01 00:00:00+00');

-- Script para criar partição automaticamente (executar mensalmente)
CREATE OR REPLACE FUNCTION create_monthly_partition()
RETURNS void AS $$
DECLARE
    start_date DATE;
    end_date DATE;
    partition_name TEXT;
BEGIN
    -- Calcular próximo mês
    start_date := DATE_TRUNC('month', CURRENT_DATE + INTERVAL '2 months');
    end_date := start_date + INTERVAL '1 month';
    partition_name := 'dict_rate_limit_states_' || TO_CHAR(start_date, 'YYYY_MM');

    -- Criar partição se não existir
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF dict_rate_limit_states ' ||
        'FOR VALUES FROM (%L) TO (%L)',
        partition_name,
        start_date,
        end_date
    );

    RAISE NOTICE 'Partição % criada para período % a %', partition_name, start_date, end_date;
END;
$$ LANGUAGE plpgsql;
```

### Data Retention Policy

```sql
-- Reter dados por 13 meses (1 ano + mês corrente)
-- Executar mensalmente via cron job

CREATE OR REPLACE FUNCTION drop_old_partitions()
RETURNS void AS $$
DECLARE
    partition_record RECORD;
    retention_date DATE;
BEGIN
    retention_date := DATE_TRUNC('month', CURRENT_DATE - INTERVAL '13 months');

    FOR partition_record IN
        SELECT tablename
        FROM pg_tables
        WHERE schemaname = 'public'
        AND tablename LIKE 'dict_rate_limit_states_20%'
        AND tablename < 'dict_rate_limit_states_' || TO_CHAR(retention_date, 'YYYY_MM')
    LOOP
        EXECUTE format('DROP TABLE IF EXISTS %I CASCADE', partition_record.tablename);
        RAISE NOTICE 'Partição % removida (dados > 13 meses)', partition_record.tablename;
    END LOOP;
END;
$$ LANGUAGE plpgsql;
```

### Queries Típicas

```sql
-- Último estado de uma política
SELECT * FROM dict_rate_limit_states
WHERE policy_name = $1
ORDER BY checked_at DESC
LIMIT 1;

-- Estados das últimas 24 horas
SELECT policy_name, available_tokens, capacity, utilization_pct, checked_at
FROM dict_rate_limit_states
WHERE checked_at >= NOW() - INTERVAL '24 hours'
ORDER BY policy_name, checked_at DESC;

-- Análise temporal: utilização média por dia (últimos 7 dias)
SELECT
    policy_name,
    DATE(checked_at) AS date,
    AVG(utilization_pct) AS avg_utilization,
    MAX(utilization_pct) AS max_utilization,
    MIN(available_tokens) AS min_available
FROM dict_rate_limit_states
WHERE checked_at >= CURRENT_DATE - INTERVAL '7 days'
GROUP BY policy_name, DATE(checked_at)
ORDER BY policy_name, date DESC;

-- Identificar picos de utilização (>90%) nas últimas 24h
SELECT policy_name, available_tokens, capacity, utilization_pct, checked_at
FROM dict_rate_limit_states
WHERE utilization_pct > 90.0
AND checked_at >= NOW() - INTERVAL '24 hours'
ORDER BY utilization_pct DESC, checked_at DESC;
```

---

## 📝 Tabela 3: `dict_rate_limit_alerts`

### Descrição
Log de audit trail de todos os alertas disparados quando thresholds são atingidos.

### DDL

```sql
-- Enum para severity
CREATE TYPE alert_severity AS ENUM ('WARNING', 'CRITICAL');

CREATE TABLE dict_rate_limit_alerts (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_name         VARCHAR(100) NOT NULL,
    severity            alert_severity NOT NULL,
    available_tokens    INTEGER NOT NULL CHECK (available_tokens >= 0),
    capacity            INTEGER NOT NULL CHECK (capacity > 0),
    utilization_pct     DECIMAL(5,2) NOT NULL CHECK (utilization_pct >= 0 AND utilization_pct <= 100),
    message             TEXT NOT NULL,
    resolved            BOOLEAN NOT NULL DEFAULT false,
    resolved_at         TIMESTAMP WITH TIME ZONE,
    resolved_by         VARCHAR(255),
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (policy_name) REFERENCES dict_rate_limit_policies(policy_name) ON DELETE CASCADE,

    CONSTRAINT resolved_validation CHECK (
        (resolved = false AND resolved_at IS NULL AND resolved_by IS NULL) OR
        (resolved = true AND resolved_at IS NOT NULL)
    )
);

-- Indexes
CREATE INDEX idx_alerts_policy_severity_created ON dict_rate_limit_alerts(policy_name, severity, created_at DESC);
CREATE INDEX idx_alerts_unresolved ON dict_rate_limit_alerts(created_at DESC)
    WHERE resolved = false;
CREATE INDEX idx_alerts_severity_created ON dict_rate_limit_alerts(severity, created_at DESC);
CREATE INDEX idx_alerts_created_at ON dict_rate_limit_alerts(created_at DESC);

-- Comments
COMMENT ON TABLE dict_rate_limit_alerts IS
    'Log de alertas de rate limit (WARNING >75%, CRITICAL >90%)';
COMMENT ON COLUMN dict_rate_limit_alerts.message IS
    'Mensagem descritiva do alerta para operadores (ex: "ENTRIES_WRITE em 92% de utilização")';
COMMENT ON COLUMN dict_rate_limit_alerts.resolved IS
    'Indica se o alerta foi resolvido (balde reabastecido ou ação tomada)';
COMMENT ON COLUMN dict_rate_limit_alerts.resolved_by IS
    'Identificador do sistema/usuário que resolveu (ex: "auto-refill", "operator:john")';
```

### Queries Típicas

```sql
-- Alertas ativos (não resolvidos)
SELECT id, policy_name, severity, utilization_pct, message, created_at,
       CURRENT_TIMESTAMP - created_at AS age
FROM dict_rate_limit_alerts
WHERE resolved = false
ORDER BY severity DESC, created_at DESC;

-- Contagem de alertas por política (últimos 7 dias)
SELECT policy_name, severity, COUNT(*) AS alert_count
FROM dict_rate_limit_alerts
WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'
GROUP BY policy_name, severity
ORDER BY alert_count DESC;

-- Resolver alerta (marcar como resolvido)
UPDATE dict_rate_limit_alerts
SET resolved = true, resolved_at = CURRENT_TIMESTAMP, resolved_by = $2
WHERE id = $1;

-- Histórico de alertas de uma política
SELECT severity, utilization_pct, message, created_at, resolved, resolved_at
FROM dict_rate_limit_alerts
WHERE policy_name = $1
ORDER BY created_at DESC
LIMIT 50;
```

---

## 📊 Views Úteis

### View 1: Latest States (Último estado de cada política)

```sql
CREATE OR REPLACE VIEW v_latest_rate_limit_states AS
SELECT DISTINCT ON (policy_name)
    policy_name,
    available_tokens,
    capacity,
    utilization_pct,
    category,
    checked_at,
    ROUND((available_tokens::DECIMAL / capacity) * 100, 2) AS availability_pct
FROM dict_rate_limit_states
ORDER BY policy_name, checked_at DESC;

COMMENT ON VIEW v_latest_rate_limit_states IS
    'Último estado de cada política (consulta otimizada)';
```

### View 2: Active Alerts (Alertas não resolvidos)

```sql
CREATE OR REPLACE VIEW v_active_rate_limit_alerts AS
SELECT
    policy_name,
    severity,
    available_tokens,
    capacity,
    utilization_pct,
    message,
    created_at,
    EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - created_at)) / 3600 AS age_hours
FROM dict_rate_limit_alerts
WHERE resolved = false
ORDER BY severity DESC, created_at DESC;

COMMENT ON VIEW v_active_rate_limit_alerts IS
    'Alertas ativos com idade em horas';
```

### View 3: Policy Health Summary (Dashboard resumo)

```sql
CREATE OR REPLACE VIEW v_rate_limit_policy_health AS
SELECT
    p.policy_name,
    p.description,
    p.capacity_max AS config_capacity,
    p.warning_threshold_pct,
    p.critical_threshold_pct,
    s.available_tokens,
    s.capacity AS current_capacity,
    s.utilization_pct,
    s.category,
    s.checked_at AS last_checked,
    CASE
        WHEN s.utilization_pct >= (100 - p.critical_threshold_pct) THEN 'CRITICAL'
        WHEN s.utilization_pct >= (100 - p.warning_threshold_pct) THEN 'WARNING'
        WHEN s.utilization_pct >= 50.0 THEN 'MODERATE'
        ELSE 'HEALTHY'
    END AS health_status,
    COUNT(a.id) FILTER (WHERE a.resolved = false) AS active_alerts_count
FROM dict_rate_limit_policies p
LEFT JOIN LATERAL (
    SELECT * FROM dict_rate_limit_states
    WHERE policy_name = p.policy_name
    ORDER BY checked_at DESC
    LIMIT 1
) s ON true
LEFT JOIN dict_rate_limit_alerts a ON a.policy_name = p.policy_name AND a.resolved = false
WHERE p.enabled = true
GROUP BY
    p.policy_name, p.description, p.capacity_max, p.warning_threshold_pct, p.critical_threshold_pct,
    s.available_tokens, s.capacity, s.utilization_pct, s.category, s.checked_at
ORDER BY
    CASE health_status
        WHEN 'CRITICAL' THEN 1
        WHEN 'WARNING' THEN 2
        WHEN 'MODERATE' THEN 3
        ELSE 4
    END,
    utilization_pct DESC NULLS LAST;

COMMENT ON VIEW v_rate_limit_policy_health IS
    'Dashboard de saúde de todas as políticas (usado em Grafana)';
```

---

## 🔧 Triggers & Functions

### Auto-update `updated_at`

```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_dict_rate_limit_policies_updated_at
    BEFORE UPDATE ON dict_rate_limit_policies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

### Auto-resolve alerts (quando balde reabastece)

```sql
-- Trigger para resolver alertas automaticamente quando balde volta a nível saudável
CREATE OR REPLACE FUNCTION auto_resolve_alerts()
RETURNS TRIGGER AS $$
BEGIN
    -- Se utilização voltou abaixo de WARNING threshold, resolver alertas
    IF NEW.utilization_pct < (100 - (
        SELECT warning_threshold_pct
        FROM dict_rate_limit_policies
        WHERE policy_name = NEW.policy_name
    )) THEN
        UPDATE dict_rate_limit_alerts
        SET resolved = true,
            resolved_at = CURRENT_TIMESTAMP,
            resolved_by = 'auto-refill'
        WHERE policy_name = NEW.policy_name
        AND resolved = false;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER auto_resolve_alerts_trigger
    AFTER INSERT ON dict_rate_limit_states
    FOR EACH ROW
    EXECUTE FUNCTION auto_resolve_alerts();
```

---

## 🧪 Testing Queries

### Test Data Generation

```sql
-- Inserir estados de teste (simular monitoramento por 24h)
DO $$
DECLARE
    ts TIMESTAMP WITH TIME ZONE;
    i INT;
BEGIN
    FOR i IN 0..287 LOOP  -- 288 pontos (24h * 12 checks/hora = 5min interval)
        ts := CURRENT_TIMESTAMP - (i * INTERVAL '5 minutes');

        INSERT INTO dict_rate_limit_states (policy_name, available_tokens, capacity, refill_tokens, refill_period_sec, utilization_pct, checked_at)
        VALUES
            ('ENTRIES_WRITE', 36000 - (i * 100), 36000, 1200, 60, ROUND((i * 100.0 / 36000.0) * 100, 2), ts),
            ('CLAIMS_WRITE', 36000 - (i * 150), 36000, 1200, 60, ROUND((i * 150.0 / 36000.0) * 100, 2), ts);
    END LOOP;
END;
$$;

-- Inserir alertas de teste
INSERT INTO dict_rate_limit_alerts (policy_name, severity, available_tokens, capacity, utilization_pct, message)
VALUES
    ('ENTRIES_WRITE', 'WARNING', 8000, 36000, 77.78, 'ENTRIES_WRITE em 77.78% de utilização'),
    ('CLAIMS_WRITE', 'CRITICAL', 2000, 36000, 94.44, 'CLAIMS_WRITE em nível crítico (5.56% restante)');
```

### Performance Tests

```sql
-- Explain analyze para query de últimos estados
EXPLAIN ANALYZE
SELECT * FROM v_latest_rate_limit_states;

-- Explain analyze para query de histórico (24h)
EXPLAIN ANALYZE
SELECT * FROM dict_rate_limit_states
WHERE checked_at >= NOW() - INTERVAL '24 hours'
ORDER BY policy_name, checked_at DESC;
```

---

## 📊 Performance Benchmarks

### Expected Query Performance

| Query | Rows | Expected Time | Index Used |
|-------|------|---------------|------------|
| Get latest state (1 policy) | 1 | <5ms | idx_states_policy_checked |
| Get all latest states (24 policies) | 24 | <10ms | idx_states_policy_checked |
| Get history 24h (1 policy) | 288 | <20ms | idx_states_policy_checked + partition |
| Get active alerts | ~10 | <5ms | idx_alerts_unresolved |
| Health summary (v_rate_limit_policy_health) | 24 | <50ms | Combined indexes |

---

## 🚀 Migration Scripts

### Up Migration (001_create_rate_limit_tables.up.sql)

```sql
-- Location: infrastructure/database/migrations/001_create_rate_limit_tables.up.sql

BEGIN;

-- 1. Create policies table
CREATE TABLE dict_rate_limit_policies (
    -- DDL completo conforme especificado acima
);

-- 2. Create states table (partitioned)
CREATE TABLE dict_rate_limit_states (
    -- DDL completo conforme especificado acima
) PARTITION BY RANGE (checked_at);

-- 3. Create alerts table
CREATE TYPE alert_severity AS ENUM ('WARNING', 'CRITICAL');
CREATE TABLE dict_rate_limit_alerts (
    -- DDL completo conforme especificado acima
);

-- 4. Create indexes (aplicado em todas as tabelas)
-- Indexes conforme especificado acima

-- 5. Create triggers
-- Triggers conforme especificado acima

-- 6. Create views
-- Views conforme especificado acima

-- 7. Create initial partitions (3 meses)
SELECT create_monthly_partition(); -- Executa 3x

-- 8. Seed initial data
INSERT INTO dict_rate_limit_policies (...) VALUES (...);  -- 24 políticas

COMMIT;
```

### Down Migration (001_create_rate_limit_tables.down.sql)

```sql
-- Location: infrastructure/database/migrations/001_create_rate_limit_tables.down.sql

BEGIN;

-- Drop views
DROP VIEW IF EXISTS v_rate_limit_policy_health CASCADE;
DROP VIEW IF EXISTS v_active_rate_limit_alerts CASCADE;
DROP VIEW IF EXISTS v_latest_rate_limit_states CASCADE;

-- Drop functions
DROP FUNCTION IF EXISTS auto_resolve_alerts() CASCADE;
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
DROP FUNCTION IF EXISTS create_monthly_partition() CASCADE;
DROP FUNCTION IF EXISTS drop_old_partitions() CASCADE;

-- Drop tables (cascades to partitions)
DROP TABLE IF EXISTS dict_rate_limit_alerts CASCADE;
DROP TABLE IF EXISTS dict_rate_limit_states CASCADE;
DROP TABLE IF EXISTS dict_rate_limit_policies CASCADE;

-- Drop enum
DROP TYPE IF EXISTS alert_severity CASCADE;

COMMIT;
```

---

## 🔗 Referências

- [SPECS-API.md](./SPECS-API.md) - Endpoints que consultam estas tabelas
- [SPECS-WORKFLOWS.md](./SPECS-WORKFLOWS.md) - Workflows que populam estas tabelas
- [SPECS-TESTING.md](./SPECS-TESTING.md) - Testes de integração com PostgreSQL

**Última Atualização**: 2025-10-31
**Versão**: 1.0.0
**Próximo**: [SPECS-API.md](./SPECS-API.md)
