# DICT Rate Limit Monitoring - Quick Start Guide

**Version**: 1.0.0 | **Status**: ✅ Production Ready | **Date**: 2025-11-01

---

## 🚀 Quick Start (5 Minutes)

### Prerequisites Check

```bash
# Verify you have all required tools
psql --version        # PostgreSQL 14+
temporal --version    # Temporal CLI
go version            # Go 1.21+
```

### 1. Database Setup (2 minutes)

```bash
# Navigate to migrations directory
cd apps/orchestration-worker/infrastructure/database/migrations

# Run migrations
export DATABASE_URL="postgres://user:pass@localhost:5432/dict_db?sslmode=disable"
goose postgres "$DATABASE_URL" up

# Verify tables created
psql $DATABASE_URL -c "\dt dict_rate_limit*"
# Expected output: 4 tables (policies, states, alerts, states_YYYY_MM partitions)
```

### 2. Environment Configuration (1 minute)

```bash
# Copy example config
cp .env.example .env

# Edit .env with your values
nano .env
```

**Required variables**:
```bash
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/dict_db?sslmode=disable

# Temporal
TEMPORAL_HOST=localhost:7233

# Bridge gRPC
BRIDGE_GRPC_ADDR=bridge-service:50051
BRIDGE_MTLS_CERT_SECRET=aws:secretsmanager:dict-bridge-mtls-cert
BRIDGE_MTLS_KEY_SECRET=aws:secretsmanager:dict-bridge-mtls-key

# Pulsar
PULSAR_URL=pulsar://localhost:6650
PULSAR_RATE_LIMIT_TOPIC=persistent://lb-conn/dict/rate-limit-alerts

# Prometheus
PROMETHEUS_PORT=9090

# Rate Limit Monitoring
DICT_RATE_LIMIT_ENABLED=true
DICT_RATE_LIMIT_CRON_SCHEDULE="*/5 * * * *"
DICT_RATE_LIMIT_WARNING_THRESHOLD=20
DICT_RATE_LIMIT_CRITICAL_THRESHOLD=10
DICT_RATE_LIMIT_RETENTION_MONTHS=13
```

### 3. Build and Run (2 minutes)

```bash
# Build
cd apps/orchestration-worker
go build -o bin/orchestration-worker ./cmd/orchestration-worker

# Run
./bin/orchestration-worker
```

**Expected logs**:
```
INFO  Starting orchestration worker
INFO  Temporal client connected
INFO  Registered workflow: MonitorPoliciesWorkflow
INFO  Registered 7 activities
INFO  Started cron workflow: dict-rate-limit-monitor-cron
INFO  Prometheus metrics exposed on :9090
INFO  Worker started successfully
```

---

## ✅ Verification (2 Minutes)

### Check Workflow Running

```bash
temporal workflow list --query 'WorkflowId="dict-rate-limit-monitor-cron"'
```

**Expected output**:
```
WORKFLOW ID                        RUN ID                             TYPE                      START TIME           EXECUTION TIME
dict-rate-limit-monitor-cron       abc123...                          MonitorPoliciesWorkflow   2025-11-01T10:00:00  Running
```

### Check Policies Loaded

```sql
psql $DATABASE_URL -c "SELECT endpoint_id, capacity, psp_category FROM dict_rate_limit_policies LIMIT 5;"
```

**Expected output**:
```
    endpoint_id     | capacity | psp_category
--------------------+----------+--------------
 ENTRIES_WRITE      |    36000 | A
 ENTRIES_READ       |    72000 | A
 CLAIMS_WRITE       |    18000 | A
 ...
(24 rows)
```

### Check Latest States

```sql
psql $DATABASE_URL -c "SELECT * FROM v_dict_rate_limit_latest_states LIMIT 3;"
```

**Expected output**:
```
    endpoint_id     | available_tokens | capacity | utilization_percent |   response_timestamp
--------------------+------------------+----------+---------------------+------------------------
 ENTRIES_WRITE      |            30000 |    36000 |               16.67 | 2025-11-01 10:05:23+00
 ENTRIES_READ       |            65000 |    72000 |                9.72 | 2025-11-01 10:05:23+00
 ...
```

### Check Prometheus Metrics

```bash
curl -s http://localhost:9090/metrics | grep dict_rate_limit | head -10
```

**Expected output**:
```
dict_rate_limit_available_tokens{endpoint_id="ENTRIES_WRITE",psp_category="A"} 30000
dict_rate_limit_capacity{endpoint_id="ENTRIES_WRITE",psp_category="A"} 36000
dict_rate_limit_utilization_percent{endpoint_id="ENTRIES_WRITE",psp_category="A"} 16.67
...
```

---

## 🔍 Monitoring

### Grafana Dashboard

Import the pre-built dashboard:

```bash
# Import dashboard JSON
curl -X POST http://localhost:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @deployment/grafana-dashboard.json
```

**Dashboard URL**: http://localhost:3000/d/dict-rate-limit

**Panels included**:
- Token availability per endpoint
- Utilization trends (last 24h)
- Consumption rate
- Recovery ETA
- Exhaustion projection
- Alert history

### Prometheus Alerts

```bash
# Import alert rules
cp deployment/prometheus-alerts.yml /etc/prometheus/alerts/dict-rate-limit.yml

# Reload Prometheus
curl -X POST http://localhost:9090/-/reload
```

**Alerts configured**:
- `DICTRateLimitCritical` - Fires when utilization >90%
- `DICTRateLimitWarning` - Fires when utilization >80%
- `DICTRateLimitExhaustionSoon` - Fires when exhaustion <1h

---

## 🐛 Troubleshooting

### Problem: Workflow not running

**Solution**:
```bash
# Check Temporal server status
temporal server status

# Restart worker
pkill orchestration-worker
./bin/orchestration-worker
```

### Problem: No policies loaded

**Solution**:
```bash
# Check Bridge connectivity
curl -k https://bridge-service:50051/health

# Check logs
tail -f logs/orchestration-worker.log | grep GetPoliciesActivity

# Force reload
psql $DATABASE_URL -c "DELETE FROM dict_rate_limit_policies;"
# Wait for next cron run (max 5 minutes)
```

### Problem: States not updating

**Solution**:
```sql
-- Check latest state timestamp
SELECT MAX(response_timestamp) FROM dict_rate_limit_states;
-- Should be within last 5 minutes

-- Check workflow history
temporal workflow describe -w dict-rate-limit-monitor-cron
```

### Problem: Alerts not resolving

**Solution**:
```sql
-- Manually trigger auto-resolve
SELECT auto_resolve_alerts('ENTRIES_WRITE', 35000, 36000);

-- Check function is working
SELECT proname, prosrc FROM pg_proc WHERE proname = 'auto_resolve_alerts';
```

---

## 📚 Next Steps

1. **Configure AlertManager**: Set up Slack/PagerDuty notifications
2. **Customize Thresholds**: Adjust WARNING/CRITICAL thresholds per endpoint
3. **Tune Cron Schedule**: Change from 5min to desired frequency
4. **Set up Backup**: Configure backup for `dict_rate_limit_*` tables
5. **Review Logs**: Set up log aggregation (ELK/Loki)

---

## 🆘 Support

**Need help?**

1. Check [DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md) for detailed instructions
2. Check [PROJECT_COMPLETE.md](PROJECT_COMPLETE.md) for architecture details
3. Review [.claude/config.json](.claude/config.json) for configuration reference

**Common Questions**:

Q: How do I change the monitoring frequency?
A: Update `DICT_RATE_LIMIT_CRON_SCHEDULE` in `.env` and restart

Q: How do I adjust WARNING/CRITICAL thresholds?
A: Update `DICT_RATE_LIMIT_WARNING_THRESHOLD` and `DICT_RATE_LIMIT_CRITICAL_THRESHOLD` in `.env`

Q: How do I test with mock data?
A: See `domain/ratelimit/*_test.go` for examples

---

## ✅ Success Checklist

After following this guide, you should have:

- [x] Database migrations applied (4 tables + 13 partitions)
- [x] Temporal cron workflow running every 5 minutes
- [x] 24+ policies loaded from DICT
- [x] States table receiving snapshots every 5 minutes
- [x] Prometheus metrics exposed on :9090
- [x] Grafana dashboard showing real-time data
- [x] Alerts configured in Prometheus

---

**Status**: ✅ Ready for Production
**Deployment Time**: ~10 minutes
**Last Updated**: 2025-11-01
