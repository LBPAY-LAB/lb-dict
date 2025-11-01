# DICT Rate Limit Monitoring System - Deployment Guide

**Version**: 1.0.0
**Date**: 2025-11-01
**Status**: Production Ready

---

## 📋 Table of Contents

1. [Prerequisites](#prerequisites)
2. [Database Setup](#database-setup)
3. [Environment Variables](#environment-variables)
4. [Deployment Steps](#deployment-steps)
5. [Verification](#verification)
6. [Monitoring](#monitoring)
7. [Troubleshooting](#troubleshooting)
8. [Rollback Procedures](#rollback-procedures)

---

## Prerequisites

### Required Infrastructure

- **PostgreSQL** 14+ (with partitioning support)
- **Temporal** Server 1.x
- **Apache Pulsar** cluster
- **Bridge gRPC** service (rsfn-connect-bacen-bridge)
- **AWS Secrets Manager** access
- **Prometheus** + AlertManager (for metrics)

### Required Permissions

- AWS IAM permissions:
  - `secretsmanager:GetSecretValue`
  - `secretsmanager:DescribeSecret`
- PostgreSQL:
  - CREATE TABLE, CREATE INDEX, CREATE FUNCTION
  - INSERT, UPDATE, DELETE, SELECT
- Temporal:
  - Workflow execution permissions
  - Activity registration permissions
- Pulsar:
  - Producer permissions on `rate-limit-alerts` topic

---

## Database Setup

### Step 1: Run Migrations

```bash
# Install Goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Set database connection
export DATABASE_URL="postgres://user:password@localhost:5432/connector_dict?sslmode=require"

# Run migrations
cd apps/orchestration-worker/infrastructure/database/migrations
goose postgres "$DATABASE_URL" up

# Verify migrations
goose postgres "$DATABASE_URL" status
```

### Step 2: Verify Tables Created

```sql
-- Connect to PostgreSQL
psql $DATABASE_URL

-- Verify tables exist
\dt dict_rate_limit*

-- Expected output:
-- dict_rate_limit_policies
-- dict_rate_limit_states (parent)
-- dict_rate_limit_states_YYYY_MM (13 partitions)
-- dict_rate_limit_alerts

-- Verify functions
\df auto_resolve_alerts
\df create_dict_rate_limit_state_partition
\df drop_old_partitions
\df perform_dict_rate_limit_maintenance

-- Verify views
\dv v_dict_rate_limit*
```

### Step 3: Initial Data Load (Optional)

```sql
-- No initial data needed
-- Policies will be loaded from DICT on first workflow run
```

---

## Environment Variables

### Required Variables

Create `.env` file or set in Kubernetes ConfigMap:

```bash
# AWS Secrets Manager
AWS_REGION=us-east-1
AWS_SECRET_BRIDGE_MTLS_ID=lb-conn/dict/bridge/mtls
AWS_SECRET_BRIDGE_ENDPOINT_ID=lb-conn/dict/bridge/endpoint
AWS_SECRET_DATABASE_CREDENTIALS_ID=lb-conn/dict/database/credentials

# Temporal
TEMPORAL_HOST=temporal.lb-conn.svc.cluster.local:7233
TEMPORAL_NAMESPACE=lb-conn
TEMPORAL_TASK_QUEUE=dict-workflows

# Pulsar
PULSAR_URL=pulsar://pulsar.lb-conn.svc.cluster.local:6650
PULSAR_TOPIC=persistent://lb-conn/dict/rate-limit-alerts

# Rate Limit Monitoring
RATE_LIMIT_CRON_SCHEDULE=*/5 * * * *  # Every 5 minutes
RATE_LIMIT_RETENTION_MONTHS=13

# Observability
LOG_LEVEL=info
METRICS_PORT=9090
HEALTH_CHECK_PORT=8080
```

### Optional Variables

```bash
# Performance tuning
DATABASE_MAX_CONNECTIONS=50
DATABASE_MIN_CONNECTIONS=10
DATABASE_CONNECTION_TIMEOUT=30s

# Temporal activity timeouts
TEMPORAL_ACTIVITY_START_TO_CLOSE_TIMEOUT=30s
TEMPORAL_ACTIVITY_SCHEDULE_TO_START_TIMEOUT=10s
TEMPORAL_ACTIVITY_RETRY_MAX_ATTEMPTS=5

# Pulsar producer settings
PULSAR_PRODUCER_COMPRESSION=LZ4
PULSAR_PRODUCER_BATCHING_MAX_MESSAGES=100
PULSAR_PRODUCER_BATCHING_MAX_DELAY=10ms
```

---

## Deployment Steps

### Step 1: Build Docker Image

```bash
# Build image
docker build -f apps/orchestration-worker/Dockerfile -t lb-conn/orchestration-worker:1.0.0 .

# Tag for registry
docker tag lb-conn/orchestration-worker:1.0.0 registry.lb-conn.com/orchestration-worker:1.0.0

# Push to registry
docker push registry.lb-conn.com/orchestration-worker:1.0.0
```

### Step 2: Deploy to Kubernetes

```bash
# Apply ConfigMap
kubectl apply -f k8s/orchestration-worker/configmap.yaml

# Apply Secrets (from AWS Secrets Manager)
kubectl apply -f k8s/orchestration-worker/secrets.yaml

# Deploy application
kubectl apply -f k8s/orchestration-worker/deployment.yaml

# Apply Service (for metrics)
kubectl apply -f k8s/orchestration-worker/service.yaml

# Apply ServiceMonitor (for Prometheus)
kubectl apply -f k8s/orchestration-worker/servicemonitor.yaml
```

### Step 3: Verify Deployment

```bash
# Check pod status
kubectl get pods -l app=orchestration-worker

# Check logs
kubectl logs -f deployment/orchestration-worker

# Expected log output:
# "Orchestration worker started"
# "Rate limit monitoring setup completed"
# "MonitorPoliciesWorkflow registered"
# "Cron workflow started: dict-rate-limit-monitor-cron"
```

### Step 4: Verify Temporal Workflow

```bash
# Using Temporal CLI
temporal workflow list --namespace lb-conn

# Expected output:
# dict-rate-limit-monitor-cron | RUNNING | CronSchedule: */5 * * * *

# Check workflow execution history
temporal workflow describe --workflow-id dict-rate-limit-monitor-cron
```

### Step 5: Verify Pulsar Topic

```bash
# Using Pulsar CLI
pulsar-admin topics stats persistent://lb-conn/dict/rate-limit-alerts

# Expected output:
# msgRateIn: 0 (initial, will increase when alerts are created)
# producerCount: 1
# subscriptionCount: 1 (core-dict consumer)
```

---

## Verification

### Health Checks

```bash
# Liveness probe
curl http://orchestration-worker:8080/health/live
# Expected: {"status":"ok"}

# Readiness probe
curl http://orchestration-worker:8080/health/ready
# Expected: {"status":"ok","checks":{"database":"ok","temporal":"ok","pulsar":"ok"}}

# Startup probe
curl http://orchestration-worker:8080/health/startup
# Expected: {"status":"ok"}
```

### Metrics Verification

```bash
# Check Prometheus metrics
curl http://orchestration-worker:9090/metrics | grep dict_rate_limit

# Expected metrics:
# dict_rate_limit_available_tokens{endpoint_id="ENTRIES_WRITE",psp_category="A"} 35000
# dict_rate_limit_capacity{endpoint_id="ENTRIES_WRITE",psp_category="A"} 36000
# dict_rate_limit_utilization_percent{endpoint_id="ENTRIES_WRITE",psp_category="A"} 2.78
# dict_rate_limit_alerts_created_total{endpoint_id="ENTRIES_WRITE",severity="WARNING",psp_category="A"} 0
```

### Database Verification

```sql
-- Check policies loaded
SELECT COUNT(*) FROM dict_rate_limit_policies;
-- Expected: ~24 policies (varies by PSP category)

-- Check states being collected
SELECT COUNT(*) FROM dict_rate_limit_states;
-- Expected: Increases every 5 minutes

-- Check latest states
SELECT * FROM v_dict_rate_limit_latest_states;
-- Expected: One row per endpoint with latest snapshot

-- Check alerts (if any)
SELECT * FROM v_dict_rate_limit_active_alerts;
-- Expected: Empty initially (unless thresholds breached)

-- Check partitions created
SELECT tablename FROM pg_tables
WHERE tablename LIKE 'dict_rate_limit_states_%'
ORDER BY tablename;
-- Expected: 13 partitions (current month + 12 previous)
```

---

## Monitoring

### Prometheus Alerts

Create alerts in AlertManager:

```yaml
groups:
  - name: dict_rate_limit
    interval: 30s
    rules:
      # CRITICAL: Tokens below 10%
      - alert: DICTRateLimitCritical
        expr: dict_rate_limit_utilization_percent > 90
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "DICT Rate Limit CRITICAL for {{ $labels.endpoint_id }}"
          description: "Utilization is {{ $value }}% (CRITICAL threshold: 90%)"

      # WARNING: Tokens below 20%
      - alert: DICTRateLimitWarning
        expr: dict_rate_limit_utilization_percent > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "DICT Rate Limit WARNING for {{ $labels.endpoint_id }}"
          description: "Utilization is {{ $value }}% (WARNING threshold: 80%)"

      # Exhaustion projection < 1 hour
      - alert: DICTRateLimitExhaustionSoon
        expr: dict_rate_limit_exhaustion_projection_seconds < 3600 and dict_rate_limit_exhaustion_projection_seconds > 0
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "DICT Rate Limit exhaustion predicted in <1h for {{ $labels.endpoint_id }}"
          description: "Tokens will exhaust in {{ $value }}s if current consumption continues"
```

### Grafana Dashboard

Import dashboard JSON (create at monitoring/grafana/dict-rate-limit-dashboard.json):

**Panels**:
1. Available Tokens (gauge per endpoint)
2. Utilization % (gauge per endpoint)
3. Consumption Rate (line graph)
4. Recovery ETA (bar chart)
5. Exhaustion Projection (bar chart)
6. Active Alerts (table)
7. Alerts Created/Resolved (counter)

---

## Troubleshooting

### Issue: Workflow not starting

**Symptoms**: No logs showing workflow execution

**Diagnosis**:
```bash
# Check Temporal connection
temporal workflow list --namespace lb-conn

# Check worker logs
kubectl logs -f deployment/orchestration-worker | grep "workflow registered"
```

**Solution**:
- Verify Temporal server is accessible
- Check TEMPORAL_HOST and TEMPORAL_NAMESPACE env vars
- Verify workflow registration in logs

### Issue: No policies being loaded

**Symptoms**: `dict_rate_limit_policies` table is empty

**Diagnosis**:
```bash
# Check Bridge connectivity
kubectl exec -it deployment/orchestration-worker -- curl http://bridge:50051/health

# Check activity logs
kubectl logs -f deployment/orchestration-worker | grep "GetPoliciesActivity"
```

**Solution**:
- Verify Bridge gRPC endpoint is accessible
- Check AWS Secrets Manager for mTLS certificates
- Verify PSP has permissions to query rate limits from DICT

### Issue: Metrics not updating

**Symptoms**: Prometheus metrics show stale data

**Diagnosis**:
```bash
# Check metrics exporter logs
kubectl logs -f deployment/orchestration-worker | grep "UpdateMetrics"

# Check database connectivity
kubectl exec -it deployment/orchestration-worker -- psql $DATABASE_URL -c "SELECT 1"
```

**Solution**:
- Verify database connection pool is healthy
- Check if states are being saved (query `dict_rate_limit_states`)
- Restart orchestration-worker pod if needed

### Issue: Alerts not being created

**Symptoms**: No alerts despite utilization >80%

**Diagnosis**:
```sql
-- Check if thresholds are actually breached
SELECT endpoint_id, available_tokens, capacity,
       100.0 - (available_tokens::DECIMAL / capacity * 100) AS utilization_pct
FROM dict_rate_limit_states
ORDER BY created_at DESC
LIMIT 20;
```

**Solution**:
- Verify AnalyzeThresholdsActivity is running
- Check CreateAlertsActivity logs for errors
- Verify alert deduplication logic (check if alerts already exist)

---

## Rollback Procedures

### Rollback Deployment

```bash
# Rollback to previous version
kubectl rollout undo deployment/orchestration-worker

# Verify rollback
kubectl rollout status deployment/orchestration-worker

# Check previous version is running
kubectl get pods -l app=orchestration-worker -o jsonpath='{.items[0].spec.containers[0].image}'
```

### Rollback Database Migrations

```bash
# Rollback one migration
goose postgres "$DATABASE_URL" down

# Rollback to specific version
goose postgres "$DATABASE_URL" down-to 003

# Verify current version
goose postgres "$DATABASE_URL" version
```

### Stop Cron Workflow

```bash
# Terminate workflow
temporal workflow terminate --workflow-id dict-rate-limit-monitor-cron --namespace lb-conn

# Verify termination
temporal workflow show --workflow-id dict-rate-limit-monitor-cron
# Expected: Status: TERMINATED
```

---

## Post-Deployment Checklist

- [ ] Database migrations applied successfully
- [ ] All 13 partitions created
- [ ] Orchestration worker pod running
- [ ] Temporal workflow started (cron: */5 * * * *)
- [ ] Policies loaded from DICT (count >0)
- [ ] States being collected every 5 minutes
- [ ] Prometheus metrics exposed on :9090
- [ ] Pulsar topic created and producer connected
- [ ] Grafana dashboard imported
- [ ] Prometheus alerts configured
- [ ] Health checks passing (liveness, readiness, startup)
- [ ] Logs showing no errors
- [ ] Bridge connectivity verified
- [ ] AWS Secrets Manager access verified

---

## Support

**Team**: Platform Engineering
**Slack**: #dict-rate-limit-monitoring
**Runbook**: https://wiki.lb-conn.com/dict/rate-limit-monitoring
**Oncall**: PagerDuty rotation "DICT Services"

---

**Last Updated**: 2025-11-01
**Document Version**: 1.0
**Author**: Tech Lead
