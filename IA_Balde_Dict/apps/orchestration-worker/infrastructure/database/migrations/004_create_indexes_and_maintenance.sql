-- +goose Up
-- +goose StatementBegin
-- Additional indexes and maintenance functions for dict_rate_limit_* tables

-- =============================================================================
-- ADDITIONAL INDEXES (performance optimization)
-- =============================================================================

-- Composite index for latest state queries (most common query pattern)
CREATE INDEX IF NOT EXISTS idx_states_endpoint_latest
    ON dict_rate_limit_states(endpoint_id, response_timestamp DESC);

-- Index for threshold analysis queries (finding endpoints near limits)
CREATE INDEX IF NOT EXISTS idx_states_low_tokens
    ON dict_rate_limit_states(available_tokens, capacity, created_at DESC)
    WHERE available_tokens <= capacity * 0.2; -- Only index states below 20%

-- =============================================================================
-- PARTITION MAINTENANCE FUNCTIONS
-- =============================================================================

-- Function to automatically create next month's partition
CREATE OR REPLACE FUNCTION create_next_month_partition()
RETURNS VOID AS $$
DECLARE
    next_month DATE;
BEGIN
    next_month := DATE_TRUNC('month', NOW() AT TIME ZONE 'UTC') + INTERVAL '1 month';
    PERFORM create_dict_rate_limit_state_partition(next_month);

    RAISE NOTICE 'Created partition for %', TO_CHAR(next_month, 'YYYY-MM');
END;
$$ LANGUAGE plpgsql;

-- Function to drop partitions older than 13 months (retention policy)
CREATE OR REPLACE FUNCTION drop_old_partitions()
RETURNS TABLE(dropped_partition TEXT, partition_date DATE) AS $$
DECLARE
    cutoff_date DATE;
    partition_record RECORD;
    partition_name TEXT;
    partition_start DATE;
BEGIN
    -- Calculate cutoff date (13 months ago)
    cutoff_date := DATE_TRUNC('month', NOW() AT TIME ZONE 'UTC') - INTERVAL '13 months';

    -- Find and drop old partitions
    FOR partition_record IN
        SELECT
            c.relname AS partition_name,
            pg_get_expr(c.relpartbound, c.oid) AS partition_bound
        FROM pg_class c
        JOIN pg_inherits i ON i.inhrelid = c.oid
        JOIN pg_class p ON p.oid = i.inhparent
        WHERE p.relname = 'dict_rate_limit_states'
          AND c.relkind = 'r'
          AND c.relname LIKE 'dict_rate_limit_states_%'
    LOOP
        -- Extract partition date from name (format: dict_rate_limit_states_YYYY_MM)
        partition_start := TO_DATE(
            SUBSTRING(partition_record.partition_name FROM '\d{4}_\d{2}$'),
            'YYYY_MM'
        );

        IF partition_start < cutoff_date THEN
            EXECUTE FORMAT('DROP TABLE IF EXISTS %I', partition_record.partition_name);

            dropped_partition := partition_record.partition_name;
            partition_date := partition_start;
            RETURN NEXT;

            RAISE NOTICE 'Dropped old partition: % (date: %)', partition_record.partition_name, partition_start;
        END IF;
    END LOOP;

    RETURN;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- ALERT RESOLUTION FUNCTIONS
-- =============================================================================

-- Function to auto-resolve alerts when tokens recover above threshold
CREATE OR REPLACE FUNCTION auto_resolve_alerts(
    p_endpoint_id VARCHAR(100),
    p_available_tokens INTEGER,
    p_capacity INTEGER
)
RETURNS INTEGER AS $$
DECLARE
    resolved_count INTEGER := 0;
    utilization_percent DECIMAL(5,2);
BEGIN
    utilization_percent := 100.0 - ((p_available_tokens::DECIMAL / p_capacity) * 100);

    -- Resolve WARNING alerts if utilization < 80%
    IF utilization_percent < 80 THEN
        UPDATE dict_rate_limit_alerts
        SET resolved = TRUE,
            resolved_at = NOW() AT TIME ZONE 'UTC',
            resolution_notes = FORMAT(
                'Auto-resolved: tokens recovered to %s/%s (%.2f%% utilization)',
                p_available_tokens,
                p_capacity,
                utilization_percent
            )
        WHERE endpoint_id = p_endpoint_id
          AND severity = 'WARNING'
          AND NOT resolved;

        GET DIAGNOSTICS resolved_count = ROW_COUNT;
    END IF;

    -- Resolve CRITICAL alerts if utilization < 90%
    IF utilization_percent < 90 THEN
        UPDATE dict_rate_limit_alerts
        SET resolved = TRUE,
            resolved_at = NOW() AT TIME ZONE 'UTC',
            resolution_notes = FORMAT(
                'Auto-resolved: tokens recovered to %s/%s (%.2f%% utilization)',
                p_available_tokens,
                p_capacity,
                utilization_percent
            )
        WHERE endpoint_id = p_endpoint_id
          AND severity = 'CRITICAL'
          AND NOT resolved;

        GET DIAGNOSTICS resolved_count = resolved_count + ROW_COUNT;
    END IF;

    RETURN resolved_count;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- STATISTICS AND MONITORING VIEWS
-- =============================================================================

-- View: Latest state for all endpoints (most common query)
CREATE OR REPLACE VIEW v_dict_rate_limit_latest_states AS
SELECT DISTINCT ON (s.endpoint_id)
    s.endpoint_id,
    p.endpoint_path,
    p.http_method,
    s.available_tokens,
    s.capacity,
    s.refill_tokens,
    s.refill_period_sec,
    s.psp_category,
    s.consumption_rate_per_minute,
    s.recovery_eta_seconds,
    s.exhaustion_projection_seconds,
    s.error_404_rate,
    100.0 - ((s.available_tokens::DECIMAL / s.capacity) * 100) AS utilization_percent,
    CASE
        WHEN s.available_tokens <= s.capacity * 0.1 THEN 'CRITICAL'
        WHEN s.available_tokens <= s.capacity * 0.2 THEN 'WARNING'
        ELSE 'OK'
    END AS health_status,
    s.response_timestamp,
    s.created_at
FROM dict_rate_limit_states s
JOIN dict_rate_limit_policies p ON p.endpoint_id = s.endpoint_id
ORDER BY s.endpoint_id, s.created_at DESC;

COMMENT ON VIEW v_dict_rate_limit_latest_states IS 'Latest state snapshot for each monitored endpoint with health status';

-- View: Active (unresolved) alerts
CREATE OR REPLACE VIEW v_dict_rate_limit_active_alerts AS
SELECT
    a.id,
    a.endpoint_id,
    p.endpoint_path,
    p.http_method,
    a.severity,
    a.available_tokens,
    a.capacity,
    a.utilization_percent,
    a.consumption_rate_per_minute,
    a.recovery_eta_seconds,
    a.psp_category,
    a.message,
    a.created_at,
    EXTRACT(EPOCH FROM (NOW() AT TIME ZONE 'UTC' - a.created_at))::INTEGER AS age_seconds
FROM dict_rate_limit_alerts a
JOIN dict_rate_limit_policies p ON p.endpoint_id = a.endpoint_id
WHERE NOT a.resolved
ORDER BY a.severity DESC, a.created_at DESC;

COMMENT ON VIEW v_dict_rate_limit_active_alerts IS 'Currently active (unresolved) rate limit alerts';

-- =============================================================================
-- SCHEDULED MAINTENANCE (to be called by cron or Temporal workflow)
-- =============================================================================

-- Function to perform all maintenance tasks
CREATE OR REPLACE FUNCTION perform_dict_rate_limit_maintenance()
RETURNS TABLE(
    task VARCHAR(50),
    status VARCHAR(20),
    details TEXT
) AS $$
BEGIN
    -- Task 1: Create next month partition
    BEGIN
        PERFORM create_next_month_partition();
        task := 'create_next_partition';
        status := 'SUCCESS';
        details := 'Next month partition created';
        RETURN NEXT;
    EXCEPTION WHEN OTHERS THEN
        task := 'create_next_partition';
        status := 'ERROR';
        details := SQLERRM;
        RETURN NEXT;
    END;

    -- Task 2: Drop old partitions
    BEGIN
        DECLARE
            drop_count INTEGER := 0;
        BEGIN
            SELECT COUNT(*) INTO drop_count FROM drop_old_partitions();
            task := 'drop_old_partitions';
            status := 'SUCCESS';
            details := FORMAT('%s partitions dropped', drop_count);
            RETURN NEXT;
        END;
    EXCEPTION WHEN OTHERS THEN
        task := 'drop_old_partitions';
        status := 'ERROR';
        details := SQLERRM;
        RETURN NEXT;
    END;

    -- Task 3: Vacuum and analyze tables
    BEGIN
        EXECUTE 'VACUUM ANALYZE dict_rate_limit_policies';
        EXECUTE 'VACUUM ANALYZE dict_rate_limit_states';
        EXECUTE 'VACUUM ANALYZE dict_rate_limit_alerts';
        task := 'vacuum_analyze';
        status := 'SUCCESS';
        details := 'All tables vacuumed and analyzed';
        RETURN NEXT;
    EXCEPTION WHEN OTHERS THEN
        task := 'vacuum_analyze';
        status := 'ERROR';
        details := SQLERRM;
        RETURN NEXT;
    END;

    RETURN;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS v_dict_rate_limit_active_alerts;
DROP VIEW IF EXISTS v_dict_rate_limit_latest_states;
DROP FUNCTION IF EXISTS perform_dict_rate_limit_maintenance();
DROP FUNCTION IF EXISTS auto_resolve_alerts(VARCHAR, INTEGER, INTEGER);
DROP FUNCTION IF EXISTS drop_old_partitions();
DROP FUNCTION IF EXISTS create_next_month_partition();
-- +goose StatementEnd
