# TST-004: Performance Tests

**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: QA Team
**Status**: ✅ Completo

---

## Sumário Executivo

Este documento apresenta os **test cases de performance** para o sistema DICT LBPay usando **k6** como ferramenta de teste de carga.

**Objetivo**: Validar que o sistema DICT atende aos requisitos de performance sob carga normal, estresse e endurance.

**Cobertura**:
- Load Test: 1000 TPS sustentado por 10 minutos
- Stress Test: Ramp-up até o ponto de quebra
- Endurance Test: 500 TPS por 4 horas
- Spike Test: Picos repentinos de tráfego
- Baseline Test: Performance sob carga normal

**Métricas Monitoradas**:
- Latência: p50, p95, p99
- Throughput: TPS (Transactions Per Second)
- Error Rate: % de erros HTTP
- Resource Usage: CPU, Memória, I/O

**Cenários Testados**:
- CreateEntry (POST /api/v1/keys)
- GetEntry (GET /api/v1/keys/{type}/{value})
- CreateClaim (POST /api/v1/claims)

**Referências**:
- [API-002: Core DICT REST API](../../04_APIs/REST/API-002_Core_DICT_REST_API.md)
- [TST-001: Test Cases CreateEntry](./TST-001_Test_Cases_CreateEntry.md)

---

## Test Environment Setup

### Environment Configuration
```yaml
Environment: performance
Core DICT API: https://dict-api-perf.lbpay.com.br
Database: PostgreSQL 16.4 (performance cluster with read replicas)
Redis Cache: Enabled (cluster mode)
Temporal: perf.temporal.lbpay.com.br
Bridge Mock: Bacen mock (fast mode)
Load Balancer: AWS ALB (2 instances minimum)
```

### Infrastructure
```yaml
API Instances:
  - Type: t3.2xlarge
  - Count: 2-10 (auto-scaling)
  - CPU: 8 vCPUs per instance
  - Memory: 32GB per instance

Database:
  - Type: db.r6g.2xlarge
  - Read Replicas: 2
  - Connection Pool: 100 connections

Redis:
  - Type: cache.r6g.xlarge
  - Nodes: 3 (cluster mode)

Load Generator:
  - k6 Cloud OR EC2 instances (c5.4xlarge)
  - Distributed load generation
```

### k6 Installation
```bash
# macOS
brew install k6

# Ubuntu/Debian
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Docker
docker pull grafana/k6:latest
```

### Test Data Preparation
```bash
# Generate test CPFs (10,000 unique valid CPFs)
node scripts/generate-test-cpfs.js > test-data/cpfs.json

# Generate test emails (10,000 unique emails)
node scripts/generate-test-emails.js > test-data/emails.json

# Generate test accounts (1,000 active accounts)
node scripts/generate-test-accounts.js > test-data/accounts.json
```

---

## Performance Test Strategy

### Test Levels

| Test Type | Load | Duration | Goal |
|-----------|------|----------|------|
| Baseline | 50 TPS | 5 min | Establish baseline metrics |
| Load Test | 1000 TPS | 10 min | Verify production capacity |
| Stress Test | 0→2000 TPS | 30 min | Find breaking point |
| Endurance | 500 TPS | 4 hours | Detect memory leaks |
| Spike Test | 50→2000→50 TPS | 15 min | Test elasticity |

### Success Criteria

```yaml
Latency:
  p50: < 150ms
  p95: < 500ms
  p99: < 1000ms

Throughput:
  Sustained: >= 1000 TPS
  Peak: >= 1500 TPS

Error Rate:
  HTTP Errors: < 0.1%
  Business Errors: < 1%

Resource Usage:
  CPU: < 70% average
  Memory: < 80% usage, stable
  Database Connections: < 80 active

Availability:
  Uptime: 99.99% during test
  No crashes or restarts
```

---

## PERF-001: Baseline Performance Test

**Priority**: P0 (Critical)
**Type**: Baseline Test
**Estimated Duration**: 5 minutes

### Test Objective
Establish baseline performance metrics under normal load (50 TPS) to serve as reference for other tests.

### Test Configuration
```javascript
// tests/performance/baseline.js
export const options = {
  stages: [
    { duration: '1m', target: 50 },   // Ramp-up to 50 TPS
    { duration: '3m', target: 50 },   // Sustain 50 TPS
    { duration: '1m', target: 0 },    // Ramp-down
  ],
  thresholds: {
    http_req_duration: ['p(50)<100', 'p(95)<200', 'p(99)<500'],
    http_req_failed: ['rate<0.001'],
  },
};
```

### Test Scenarios
```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');
const BASE_URL = __ENV.BASE_URL || 'https://dict-api-perf.lbpay.com.br';

export default function () {
  const scenarios = [
    { weight: 0.5, fn: createEntry },
    { weight: 0.3, fn: getEntry },
    { weight: 0.2, fn: createClaim },
  ];

  const scenario = weightedRandom(scenarios);
  scenario.fn();

  sleep(1);
}

function createEntry() {
  const url = `${BASE_URL}/api/v1/keys`;
  const payload = JSON.stringify({
    key_type: 'CPF',
    key_value: generateRandomCPF(),
    account: {
      ispb: '12345678',
      account_number: randomAccount(),
      branch: '0001',
      account_type: 'CACC'
    }
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${__ENV.ACCESS_TOKEN}`,
    },
    tags: { name: 'CreateEntry' },
  };

  const res = http.post(url, payload, params);

  const success = check(res, {
    'CreateEntry: status is 201': (r) => r.status === 201,
    'CreateEntry: has entry_id': (r) => JSON.parse(r.body).entry_id !== undefined,
    'CreateEntry: response time < 200ms': (r) => r.timings.duration < 200,
  });

  errorRate.add(!success);
}

function getEntry() {
  const cpf = selectExistingCPF();
  const url = `${BASE_URL}/api/v1/keys/CPF/${cpf}`;

  const params = {
    headers: {
      'Authorization': `Bearer ${__ENV.ACCESS_TOKEN}`,
    },
    tags: { name: 'GetEntry' },
  };

  const res = http.get(url, params);

  const success = check(res, {
    'GetEntry: status is 200': (r) => r.status === 200,
    'GetEntry: has entry data': (r) => JSON.parse(r.body).entry_id !== undefined,
    'GetEntry: response time < 100ms': (r) => r.timings.duration < 100,
  });

  errorRate.add(!success);
}

function createClaim() {
  const cpf = selectExistingCPF();
  const url = `${BASE_URL}/api/v1/claims`;
  const payload = JSON.stringify({
    key_type: 'CPF',
    key_value: cpf,
    claim_type: 'PORTABILITY',
    account: {
      ispb: '87654321',
      account_number: randomAccount(),
      branch: '0002',
      account_type: 'CACC'
    }
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${__ENV.ACCESS_TOKEN}`,
    },
    tags: { name: 'CreateClaim' },
  };

  const res = http.post(url, payload, params);

  const success = check(res, {
    'CreateClaim: status is 201': (r) => r.status === 201,
    'CreateClaim: has claim_id': (r) => JSON.parse(r.body).claim_id !== undefined,
  });

  errorRate.add(!success);
}
```

### Expected Results
```yaml
Latency:
  p50: 50-80ms
  p95: 100-150ms
  p99: 200-300ms

Throughput:
  Average: 50 TPS
  Total Requests: 15,000

Error Rate: < 0.1%

Resource Usage:
  CPU: 20-30%
  Memory: Stable, no growth
```

### Execution
```bash
# Set environment variables
export ACCESS_TOKEN=$(curl -X POST https://auth-perf.lbpay.com.br/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"perf.user@lbpay.com.br","password":"Perf@1234"}' | jq -r .access_token)

export BASE_URL=https://dict-api-perf.lbpay.com.br

# Run baseline test
k6 run tests/performance/baseline.js

# With detailed output
k6 run --out json=results/baseline.json tests/performance/baseline.js

# With InfluxDB/Grafana
k6 run --out influxdb=http://localhost:8086/k6 tests/performance/baseline.js
```

### Status
⬜ Not Run | 🟡 In Progress | ✅ Pass | ❌ Fail

---

## PERF-002: Load Test - 1000 TPS Sustained

**Priority**: P0 (Critical)
**Type**: Load Test
**Estimated Duration**: 17 minutes

### Test Objective
Verify system can handle production load of 1000 TPS sustained for 10 minutes with acceptable latency.

### Test Configuration
```javascript
// tests/performance/load-1000tps.js
export const options = {
  stages: [
    { duration: '5m', target: 1000 },  // Ramp-up to 1000 TPS
    { duration: '10m', target: 1000 }, // Sustain 1000 TPS
    { duration: '2m', target: 0 },     // Ramp-down
  ],
  thresholds: {
    // Latency thresholds
    http_req_duration: ['p(50)<150', 'p(95)<500', 'p(99)<1000'],

    // Error rate thresholds
    http_req_failed: ['rate<0.001'],  // < 0.1% errors

    // Per-scenario thresholds
    'http_req_duration{name:CreateEntry}': ['p(95)<500'],
    'http_req_duration{name:GetEntry}': ['p(95)<100'],
    'http_req_duration{name:CreateClaim}': ['p(95)<600'],
  },
  ext: {
    loadimpact: {
      distribution: {
        'amazon:us:ashburn': { loadZone: 'amazon:us:ashburn', percent: 50 },
        'amazon:br:sao paulo': { loadZone: 'amazon:br:sao paulo', percent: 50 },
      },
    },
  },
};
```

### BDD Format
```gherkin
Given the system is configured for production load
When 1000 TPS is sustained for 10 minutes
Then p50 latency is < 150ms
And p95 latency is < 500ms
And p99 latency is < 1000ms
And error rate is < 0.1%
And CPU usage is < 70%
And memory usage is stable
And no database connection pool exhaustion occurs
```

### Expected Results
```yaml
Total Requests: ~900,000 requests
Success Rate: > 99.9%

Latency by Scenario:
  CreateEntry:
    p50: < 120ms
    p95: < 450ms
    p99: < 900ms

  GetEntry:
    p50: < 30ms (cache hit)
    p95: < 80ms
    p99: < 150ms

  CreateClaim:
    p50: < 150ms
    p95: < 550ms
    p99: < 1000ms

Resource Usage:
  API CPU: 60-70%
  API Memory: 60-70%, stable
  Database CPU: 50-60%
  Database Connections: 60-80 active
  Redis Hit Rate: > 80%
```

### Execution
```bash
# Run load test
k6 run tests/performance/load-1000tps.js

# With cloud execution (distributed load)
k6 cloud tests/performance/load-1000tps.js

# With detailed metrics
k6 run --summary-export=results/load-1000tps-summary.json \
       --out json=results/load-1000tps.json \
       tests/performance/load-1000tps.js
```

### Monitoring During Test
```bash
# Monitor API metrics
watch -n 1 'curl -s http://dict-api-perf.lbpay.com.br/metrics | grep http_requests_total'

# Monitor database connections
psql -h db-perf.lbpay.com.br -U dict -c "SELECT count(*) FROM pg_stat_activity WHERE state='active';"

# Monitor Redis
redis-cli -h redis-perf.lbpay.com.br INFO stats | grep keyspace_hits
```

### Status
⬜ Not Run

---

## PERF-003: Stress Test - Find Breaking Point

**Priority**: P0 (Critical)
**Type**: Stress Test
**Estimated Duration**: 35 minutes

### Test Objective
Determine the maximum load the system can handle before degradation or failure.

### Test Configuration
```javascript
// tests/performance/stress-test.js
export const options = {
  stages: [
    { duration: '5m', target: 500 },   // Warm-up to 500 TPS
    { duration: '5m', target: 1000 },  // Ramp to 1000 TPS
    { duration: '5m', target: 1500 },  // Ramp to 1500 TPS
    { duration: '5m', target: 2000 },  // Ramp to 2000 TPS
    { duration: '5m', target: 2500 },  // Ramp to 2500 TPS (expect failure)
    { duration: '5m', target: 3000 },  // Ramp to 3000 TPS (breaking point)
    { duration: '5m', target: 0 },     // Ramp-down
  ],
  thresholds: {
    // No hard thresholds - we expect some to fail
    // This is exploratory to find limits
  },
};
```

### BDD Format
```gherkin
Given the system is under increasing load
When load is ramped from 500 TPS to 3000 TPS
Then the breaking point is identified
And the system degrades gracefully
And the system recovers when load decreases
And no data corruption occurs
```

### Metrics to Identify Breaking Point
```yaml
Breaking Point Indicators:
  - p95 latency exceeds 5 seconds
  - Error rate exceeds 5%
  - HTTP 503 Service Unavailable responses
  - Database connection pool exhausted
  - API instances crashing/restarting
  - Memory usage exceeding 90%

Graceful Degradation Check:
  - System returns proper error codes (429, 503)
  - Circuit breakers activate
  - No data corruption
  - System recovers when load decreases
```

### Expected Results
```yaml
Expected Breaking Point: 1800-2200 TPS

At Breaking Point:
  - p95 latency: 3-5 seconds
  - Error rate: 3-5%
  - CPU: 90-95%
  - Memory: 80-85%
  - Database connections: pool exhausted

Recovery:
  - System recovers within 2 minutes after load decrease
  - No manual intervention required
  - All data integrity maintained
```

### Execution
```bash
# Run stress test with detailed logging
k6 run --out json=results/stress-test.json tests/performance/stress-test.js

# Monitor system during test
./scripts/monitor-stress-test.sh
```

### Post-Test Verification
```sql
-- Verify no data corruption
SELECT COUNT(*) FROM dict.entries WHERE status = 'CORRUPTED';
-- Expected: 0

-- Check for orphaned entries
SELECT COUNT(*) FROM dict.entries WHERE created_at > NOW() - INTERVAL '1 hour'
  AND status = 'PENDING' AND updated_at < NOW() - INTERVAL '10 minutes';
-- Expected: 0 (all should be completed or failed)

-- Verify audit log consistency
SELECT COUNT(*) FROM dict.audit_logs WHERE created_at > NOW() - INTERVAL '1 hour';
-- Should match number of successful operations
```

### Status
⬜ Not Run

---

## PERF-004: Endurance Test - 4 Hours at 500 TPS

**Priority**: P1 (High)
**Type**: Soak Test / Endurance Test
**Estimated Duration**: 4 hours 10 minutes

### Test Objective
Detect memory leaks, resource exhaustion, and performance degradation over extended period.

### Test Configuration
```javascript
// tests/performance/endurance-4h.js
export const options = {
  stages: [
    { duration: '5m', target: 500 },   // Ramp-up
    { duration: '4h', target: 500 },   // Sustain for 4 hours
    { duration: '5m', target: 0 },     // Ramp-down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.001'],
  },
};
```

### BDD Format
```gherkin
Given the system runs under sustained load
When 500 TPS is maintained for 4 hours
Then memory usage remains stable
And no memory leaks are detected
And latency does not degrade over time
And error rate remains < 0.1%
And database connections remain stable
```

### Metrics to Monitor
```yaml
Memory Leak Detection:
  - API memory usage should be flat (±5% variance)
  - Database memory stable
  - Redis memory stable
  - No OOM kills

Resource Exhaustion:
  - Database connections stable (not growing)
  - File descriptors stable
  - Thread count stable
  - Disk space stable

Performance Degradation:
  - Compare latency: first hour vs. last hour
  - p95 latency variance < 10%
  - Throughput remains constant
```

### Expected Results
```yaml
Total Requests: ~7,200,000 requests
Total Duration: 4h 10m

Latency (consistent across all hours):
  p50: 100-120ms
  p95: 400-500ms
  p99: 800-1000ms

Memory Usage:
  Hour 1: 4.2GB ± 0.2GB
  Hour 2: 4.2GB ± 0.2GB
  Hour 3: 4.2GB ± 0.2GB
  Hour 4: 4.2GB ± 0.2GB
  Result: STABLE (no leak)

Database Connections:
  Consistently 60-70 active connections
  No growth over time

Error Rate: < 0.1% throughout
```

### Execution
```bash
# Run endurance test (run in screen/tmux for long duration)
screen -S endurance-test
k6 run --out json=results/endurance-4h.json tests/performance/endurance-4h.js

# Detach: Ctrl+A, D
# Reattach: screen -r endurance-test
```

### Monitoring Script
```bash
#!/bin/bash
# scripts/monitor-endurance.sh

while true; do
  TIMESTAMP=$(date +%Y-%m-%d_%H:%M:%S)

  # API memory
  API_MEM=$(kubectl top pod -l app=dict-api -n dict | awk 'NR>1 {print $3}')

  # Database connections
  DB_CONN=$(psql -h db-perf -U dict -t -c "SELECT count(*) FROM pg_stat_activity WHERE state='active';")

  # Log metrics
  echo "$TIMESTAMP,$API_MEM,$DB_CONN" >> results/endurance-metrics.csv

  sleep 60  # Log every minute
done
```

### Post-Test Analysis
```python
# scripts/analyze-endurance.py
import pandas as pd
import matplotlib.pyplot as plt

# Load metrics
df = pd.read_csv('results/endurance-metrics.csv',
                 names=['timestamp', 'memory', 'db_connections'])
df['timestamp'] = pd.to_datetime(df['timestamp'])

# Plot memory over time
plt.figure(figsize=(12, 6))
plt.plot(df['timestamp'], df['memory'])
plt.xlabel('Time')
plt.ylabel('Memory (MB)')
plt.title('API Memory Usage - 4 Hour Endurance Test')
plt.savefig('results/endurance-memory.png')

# Detect memory leak (linear regression slope)
from scipy import stats
slope, intercept, r_value, p_value, std_err = stats.linregress(
    range(len(df)), df['memory']
)

if abs(slope) < 0.01:
    print("✅ No memory leak detected (slope:", slope, ")")
else:
    print("❌ Possible memory leak (slope:", slope, ")")
```

### Status
⬜ Not Run

---

## PERF-005: Spike Test - Sudden Traffic Bursts

**Priority**: P1 (High)
**Type**: Spike Test
**Estimated Duration**: 15 minutes

### Test Objective
Verify system handles sudden traffic spikes and recovers gracefully.

### Test Configuration
```javascript
// tests/performance/spike-test.js
export const options = {
  stages: [
    { duration: '2m', target: 100 },    // Normal load
    { duration: '1m', target: 2000 },   // Spike to 2000 TPS
    { duration: '2m', target: 2000 },   // Sustain spike
    { duration: '1m', target: 100 },    // Drop back to normal
    { duration: '2m', target: 100 },    // Recover
    { duration: '1m', target: 2000 },   // Second spike
    { duration: '2m', target: 2000 },   // Sustain
    { duration: '1m', target: 100 },    // Drop
    { duration: '3m', target: 100 },    // Stabilize
  ],
  thresholds: {
    'http_req_duration{phase:normal}': ['p(95)<200'],
    'http_req_duration{phase:spike}': ['p(95)<2000'],
    'http_req_duration{phase:recovery}': ['p(95)<300'],
  },
};
```

### BDD Format
```gherkin
Given the system is at normal load (100 TPS)
When traffic suddenly spikes to 2000 TPS
Then the system scales automatically
And handles the spike with degraded but acceptable performance
When traffic returns to normal
Then the system recovers within 1 minute
And performance returns to baseline
```

### Expected Results
```yaml
Normal Phase (100 TPS):
  p95: < 200ms
  Error Rate: < 0.1%

Spike Phase (2000 TPS):
  p95: < 2000ms (degraded but functional)
  Error Rate: < 5%
  Auto-scaling triggered within 30 seconds

Recovery Phase (back to 100 TPS):
  Recovery time: < 60 seconds
  p95 returns to < 200ms
  Error rate returns to < 0.1%

Auto-Scaling Verification:
  - API instances scale from 2 → 6 during spike
  - API instances scale back to 2 after recovery
```

### Execution
```bash
# Run spike test
k6 run tests/performance/spike-test.js

# Monitor auto-scaling
watch -n 5 'kubectl get pods -n dict -l app=dict-api'
```

### Status
⬜ Not Run

---

## PERF-006: Mixed Workload - Real World Simulation

**Priority**: P1 (High)
**Type**: Load Test
**Estimated Duration**: 20 minutes

### Test Objective
Simulate realistic production workload with mixed operations and realistic think times.

### Test Configuration
```javascript
// tests/performance/mixed-workload.js
import { scenario } from 'k6/execution';

export const options = {
  scenarios: {
    createEntry: {
      executor: 'ramping-arrival-rate',
      startRate: 50,
      timeUnit: '1s',
      preAllocatedVUs: 200,
      maxVUs: 1000,
      stages: [
        { duration: '5m', target: 400 },   // Ramp to 400 TPS
        { duration: '10m', target: 400 },  // Sustain
        { duration: '5m', target: 0 },     // Ramp down
      ],
      exec: 'createEntry',
    },
    getEntry: {
      executor: 'ramping-arrival-rate',
      startRate: 100,
      timeUnit: '1s',
      preAllocatedVUs: 300,
      maxVUs: 1500,
      stages: [
        { duration: '5m', target: 800 },   // Ramp to 800 TPS
        { duration: '10m', target: 800 },  // Sustain
        { duration: '5m', target: 0 },
      ],
      exec: 'getEntry',
    },
    createClaim: {
      executor: 'ramping-arrival-rate',
      startRate: 20,
      timeUnit: '1s',
      preAllocatedVUs: 100,
      maxVUs: 500,
      stages: [
        { duration: '5m', target: 150 },
        { duration: '10m', target: 150 },
        { duration: '5m', target: 0 },
      ],
      exec: 'createClaim',
    },
    deleteClaim: {
      executor: 'ramping-arrival-rate',
      startRate: 10,
      timeUnit: '1s',
      preAllocatedVUs: 50,
      maxVUs: 200,
      stages: [
        { duration: '5m', target: 50 },
        { duration: '10m', target: 50 },
        { duration: '5m', target: 0 },
      ],
      exec: 'deleteClaim',
    },
  },
  thresholds: {
    'http_req_duration{scenario:createEntry}': ['p(95)<500'],
    'http_req_duration{scenario:getEntry}': ['p(95)<100'],
    'http_req_duration{scenario:createClaim}': ['p(95)<600'],
    'http_req_duration{scenario:deleteClaim}': ['p(95)<400'],
  },
};

export function createEntry() {
  // Implementation...
}

export function getEntry() {
  // Implementation...
}

export function createClaim() {
  // Implementation...
}

export function deleteClaim() {
  // Implementation...
}
```

### Expected Results
```yaml
Total Throughput: ~1400 TPS

Operation Distribution:
  - CreateEntry: 400 TPS (28%)
  - GetEntry: 800 TPS (57%)
  - CreateClaim: 150 TPS (11%)
  - DeleteClaim: 50 TPS (4%)

Latency by Operation:
  CreateEntry p95: < 500ms
  GetEntry p95: < 100ms
  CreateClaim p95: < 600ms
  DeleteClaim p95: < 400ms

Error Rate: < 0.1% across all operations
```

### Status
⬜ Not Run

---

## Performance Test Execution Checklist

### Pre-Test Checklist
- [ ] Performance environment deployed and stable
- [ ] Database tuned (indexes, connection pool, query optimization)
- [ ] Redis cache configured and warmed up
- [ ] Auto-scaling policies configured
- [ ] Monitoring dashboards ready (Grafana, CloudWatch)
- [ ] Test data generated and loaded
- [ ] Access tokens generated and valid
- [ ] Load generators provisioned (k6 Cloud or EC2 instances)
- [ ] Stakeholders notified of test schedule

### During Test
- [ ] Monitor system metrics in real-time
- [ ] Watch for error spikes
- [ ] Check auto-scaling behavior
- [ ] Monitor database performance
- [ ] Track memory usage trends
- [ ] Log any anomalies or incidents

### Post-Test Checklist
- [ ] Collect and archive all test results
- [ ] Generate performance reports
- [ ] Analyze latency percentiles
- [ ] Verify no data corruption
- [ ] Check for memory leaks
- [ ] Document breaking points
- [ ] Compare against SLAs
- [ ] Create performance improvement backlog items
- [ ] Share results with stakeholders

---

## Performance Analysis & Reporting

### k6 Results Analysis
```bash
# Generate summary report
k6 run --summary-export=summary.json tests/performance/load-1000tps.js

# Key metrics from summary
cat summary.json | jq '.metrics.http_req_duration'
cat summary.json | jq '.metrics.http_req_failed'

# Convert JSON to CSV for analysis
k6 convert --output csv results/load-1000tps.json > results/load-1000tps.csv
```

### Grafana Dashboard
```yaml
Dashboard Panels:
  - Request Rate (TPS)
  - Response Time Percentiles (p50, p95, p99)
  - Error Rate
  - Active VUs
  - HTTP Request Duration Trend
  - Database Connections
  - API CPU & Memory
  - Cache Hit Rate
```

### Performance Report Template
```markdown
# Performance Test Report - DICT LBPay

**Test Date**: 2025-10-25
**Test Type**: Load Test (1000 TPS)
**Duration**: 17 minutes
**Test ID**: PERF-002

## Summary
✅ PASS - System handled 1000 TPS with acceptable latency

## Key Metrics
| Metric | Target | Actual | Status |
|--------|--------|--------|--------|
| p50 Latency | < 150ms | 112ms | ✅ Pass |
| p95 Latency | < 500ms | 387ms | ✅ Pass |
| p99 Latency | < 1000ms | 823ms | ✅ Pass |
| Error Rate | < 0.1% | 0.03% | ✅ Pass |
| Throughput | 1000 TPS | 1003 TPS | ✅ Pass |

## Performance by Operation
- CreateEntry: p95 = 420ms
- GetEntry: p95 = 68ms
- CreateClaim: p95 = 512ms

## Resource Usage
- API CPU: 65% average
- API Memory: 68%, stable
- Database Connections: 72 active
- Redis Hit Rate: 87%

## Bottlenecks Identified
None

## Recommendations
- System is production-ready for 1000 TPS
- Consider scaling to 3 API instances for 1500+ TPS
```

---

## Continuous Performance Testing

### CI/CD Integration
```yaml
# .github/workflows/performance-test.yml
name: Performance Test

on:
  schedule:
    - cron: '0 2 * * 0'  # Weekly on Sunday 2am
  workflow_dispatch:

jobs:
  performance-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup k6
        run: |
          sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
          echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
          sudo apt-get update
          sudo apt-get install k6

      - name: Run Baseline Test
        run: |
          export ACCESS_TOKEN=${{ secrets.PERF_ACCESS_TOKEN }}
          k6 run --out json=baseline.json tests/performance/baseline.js

      - name: Check Performance Thresholds
        run: |
          ./scripts/check-performance-regression.sh baseline.json

      - name: Upload Results
        uses: actions/upload-artifact@v3
        with:
          name: performance-results
          path: baseline.json
```

### Performance Regression Detection
```python
# scripts/check-performance-regression.py
import json
import sys

def check_regression(current_file, baseline_file):
    with open(current_file) as f:
        current = json.load(f)

    with open(baseline_file) as f:
        baseline = json.load(f)

    current_p95 = current['metrics']['http_req_duration']['p(95)']
    baseline_p95 = baseline['metrics']['http_req_duration']['p(95)']

    regression_threshold = 0.2  # 20% regression

    if current_p95 > baseline_p95 * (1 + regression_threshold):
        print(f"❌ REGRESSION: p95 increased by {(current_p95/baseline_p95 - 1)*100:.1f}%")
        print(f"   Baseline: {baseline_p95:.2f}ms")
        print(f"   Current: {current_p95:.2f}ms")
        sys.exit(1)
    else:
        print(f"✅ No regression: p95 is {current_p95:.2f}ms (baseline: {baseline_p95:.2f}ms)")
        sys.exit(0)

if __name__ == '__main__':
    check_regression('baseline.json', 'baseline-reference.json')
```

---

## Glossary

- **TPS**: Transactions Per Second
- **p50/p95/p99**: Latency percentiles (50th, 95th, 99th)
- **VU**: Virtual User (k6 load generator concept)
- **RPS**: Requests Per Second (same as TPS)
- **Soak Test**: Extended duration test (endurance)
- **Spike Test**: Sudden load increase test
- **Stress Test**: Beyond-capacity test to find limits

---

**Última Revisão**: 2025-10-25
**Aprovado por**: QA Lead
**Próxima Execução**: Sprint performance testing cycle
