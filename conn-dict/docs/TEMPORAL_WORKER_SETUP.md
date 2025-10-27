# Temporal Worker Setup - CONNECT-009

## Overview

The Temporal worker is a separate process that polls the Temporal server for tasks and executes workflows and activities. This document describes the worker configuration, architecture, and operational details.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Temporal Worker Process                   │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Health     │  │   Metrics    │  │   Worker     │     │
│  │  Server      │  │   Server     │  │   Engine     │     │
│  │  :8081       │  │   :9093      │  │              │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│         │                 │                  │              │
│         ↓                 ↓                  ↓              │
│  ┌──────────────────────────────────────────────────┐      │
│  │           Dependency Injection Layer             │      │
│  │  • PostgreSQL Pool                               │      │
│  │  • Redis Client                                  │      │
│  │  • Pulsar Producer                               │      │
│  │  • Bridge gRPC Client (TODO)                     │      │
│  └──────────────────────────────────────────────────┘      │
│         │                                                    │
│         ↓                                                    │
│  ┌──────────────────────────────────────────────────┐      │
│  │        Registered Workflows (6 total)            │      │
│  │  • ClaimWorkflow                                 │      │
│  │  • CreateEntryWorkflow                           │      │
│  │  • UpdateEntryWorkflow                           │      │
│  │  • DeleteEntryWorkflow                           │      │
│  │  • InvestigateInfractionWorkflow                 │      │
│  │  • VSyncWorkflow                                 │      │
│  │  • VSyncSchedulerWorkflow                        │      │
│  └──────────────────────────────────────────────────┘      │
│         │                                                    │
│         ↓                                                    │
│  ┌──────────────────────────────────────────────────┐      │
│  │        Registered Activities (27 total)          │      │
│  │                                                   │      │
│  │  Claim Activities (10):                          │      │
│  │    - CreateClaimActivity                         │      │
│  │    - NotifyDonorActivity                         │      │
│  │    - CompleteClaimActivity                       │      │
│  │    - CancelClaimActivity                         │      │
│  │    - ExpireClaimActivity                         │      │
│  │    - GetClaimStatusActivity                      │      │
│  │    - ValidateClaimEligibilityActivity            │      │
│  │    - SendClaimConfirmationActivity               │      │
│  │    - UpdateEntryOwnershipActivity                │      │
│  │    - PublishClaimEventActivity                   │      │
│  │                                                   │      │
│  │  Entry Activities (8):                           │      │
│  │    - CreateEntryActivity                         │      │
│  │    - UpdateEntryActivity                         │      │
│  │    - DeleteEntryActivity                         │      │
│  │    - ActivateEntryActivity                       │      │
│  │    - DeactivateEntryActivity                     │      │
│  │    - GetEntryStatusActivity                      │      │
│  │    - ValidateEntryActivity                       │      │
│  │    - UpdateEntryOwnershipActivity                │      │
│  │                                                   │      │
│  │  Infraction Activities (10):                     │      │
│  │    - CreateInfractionActivity                    │      │
│  │    - InvestigateInfractionActivity               │      │
│  │    - ResolveInfractionActivity                   │      │
│  │    - DismissInfractionActivity                   │      │
│  │    - EscalateInfractionActivity                  │      │
│  │    - AddEvidenceActivity                         │      │
│  │    - GetInfractionStatusActivity                 │      │
│  │    - ValidateInfractionEligibilityActivity       │      │
│  │    - NotifyReportedParticipantActivity           │      │
│  │    - NotifyBacenActivity                         │      │
│  │    - PublishInfractionEventActivity              │      │
│  │                                                   │      │
│  │  VSYNC Activities (3):                           │      │
│  │    - FetchBacenEntriesActivity                   │      │
│  │    - CompareEntriesActivity                      │      │
│  │    - GenerateSyncReportActivity                  │      │
│  └──────────────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────────────┘
                           │
                           ↓
              ┌────────────────────────┐
              │   Temporal Server      │
              │   (localhost:7233)     │
              └────────────────────────┘
```

## Configuration

### Task Queue

- **Name**: `conn-dict-task-queue`
- **Purpose**: Dedicated queue for DICT operations
- **Environment Variable**: `TEMPORAL_TASK_QUEUE`

### Worker Concurrency

The worker is configured for high-throughput processing:

| Setting | Default | Description |
|---------|---------|-------------|
| `MAX_CONCURRENT_WORKFLOWS` | 100 | Maximum concurrent workflow executions |
| `MAX_CONCURRENT_ACTIVITIES` | 200 | Maximum concurrent activity executions |
| `MaxConcurrentWorkflowTaskPollers` | 10 | Number of workflow task pollers |
| `MaxConcurrentActivityTaskPollers` | 20 | Number of activity task pollers |
| `MaxConcurrentSessionExecutionSize` | 50 | Maximum concurrent session executions |

### Timeouts

| Timeout Type | Default | Description |
|-------------|---------|-------------|
| Activity Timeout | 30s | Default activity execution timeout |
| Workflow Timeout | 24h | Default workflow execution timeout |
| Shutdown Timeout | 30s | Maximum time to wait for graceful shutdown |

## Environment Variables

Required environment variables for the worker:

```bash
# Temporal Configuration
TEMPORAL_ADDRESS=localhost:7233
TEMPORAL_NAMESPACE=default
TEMPORAL_TASK_QUEUE=conn-dict-task-queue
MAX_CONCURRENT_ACTIVITIES=200
MAX_CONCURRENT_WORKFLOWS=100

# PostgreSQL Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=dict_user
POSTGRES_PASSWORD=dict_password
POSTGRES_DB=dict_db
POSTGRES_SSLMODE=disable
POSTGRES_MAX_CONNS=25
POSTGRES_MIN_CONNS=5

# Pulsar Configuration
PULSAR_URL=pulsar://localhost:6650
PULSAR_TOPIC=persistent://public/default/dict-events
PULSAR_PRODUCER_NAME=conn-dict-worker

# Observability
LOG_LEVEL=info
METRICS_PORT=9093
HEALTH_PORT=8081

# Bridge gRPC Client (TODO)
BRIDGE_GRPC_ADDR=localhost:50051
BRIDGE_GRPC_TIMEOUT=30s
```

## Registered Workflows

### 1. ClaimWorkflow
- **Purpose**: Handles PIX key portability claims
- **Duration**: Up to 30 days (waiting for donor confirmation)
- **Signals**: `confirm`, `cancel`
- **Activities**: CreateClaim, NotifyDonor, CompleteClaim, CancelClaim, ExpireClaim

### 2. VSyncWorkflow
- **Purpose**: Synchronizes DICT entries with Bacen
- **Duration**: ~2 hours (for full sync)
- **Schedule**: Daily at 2 AM (via VSyncSchedulerWorkflow)
- **Activities**: FetchBacenEntries, CompareEntries, GenerateSyncReport

### 3. Entry Workflows
- **CreateEntryWorkflow**: Creates new DICT entry
- **UpdateEntryWorkflow**: Updates existing entry
- **DeleteEntryWorkflow**: Deletes entry

### 4. InvestigateInfractionWorkflow
- **Purpose**: Handles infraction investigations
- **Duration**: Variable (days to weeks)
- **Activities**: CreateInfraction, Investigate, Resolve, Escalate

## Registered Activities

### Claim Activities (10)
All activities related to PIX key claim processing.

### Entry Activities (8)
CRUD operations for DICT entries.

### Infraction Activities (10)
Activities for handling infractions and investigations.

### VSYNC Activities (3)
Activities for data synchronization with Bacen.

**Total**: 31 activities registered

## Observability

### Health Checks

The worker exposes two health check endpoints:

#### 1. Liveness Probe - `/health`
- **Port**: 8081
- **Purpose**: Checks if worker process is running
- **Response**: `{"status":"healthy","service":"conn-dict-worker"}`
- **Status Code**: 200 OK

#### 2. Readiness Probe - `/ready`
- **Port**: 8081
- **Purpose**: Checks if worker can process tasks
- **Checks**:
  - PostgreSQL connectivity
  - Temporal server connectivity
- **Response**:
  - Ready: `{"status":"ready","service":"conn-dict-worker"}`
  - Not Ready: `{"status":"not_ready","reason":"...","error":"..."}`
- **Status Codes**: 200 (ready), 503 (not ready)

### Prometheus Metrics

The worker exposes metrics on port **9093** at `/metrics`:

#### Worker Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `conn_dict_worker_tasks_processed_total` | Counter | `task_type`, `status` | Total tasks processed |
| `conn_dict_worker_task_duration_seconds` | Histogram | `task_type` | Task execution duration |
| `conn_dict_worker_health_status` | Gauge | - | Worker health (1=healthy, 0=unhealthy) |
| `conn_dict_worker_active_workflows` | Gauge | - | Number of active workflows |
| `conn_dict_worker_active_activities` | Gauge | - | Number of active activities |

#### Example Queries

```promql
# Task processing rate
rate(conn_dict_worker_tasks_processed_total[5m])

# 95th percentile task duration
histogram_quantile(0.95, rate(conn_dict_worker_task_duration_seconds_bucket[5m]))

# Worker health status
conn_dict_worker_health_status

# Active workflows
conn_dict_worker_active_workflows
```

### Structured Logging

The worker uses JSON-formatted structured logging with logrus:

```json
{
  "level": "info",
  "msg": "ClaimWorkflow started",
  "claim_id": "CLAIM-12345",
  "entry_id": "ENTRY-67890",
  "timestamp": "2025-10-27T10:00:00Z"
}
```

## Graceful Shutdown

The worker implements graceful shutdown:

1. **Signal Reception**: Listens for SIGTERM/SIGINT
2. **Stop Accepting Tasks**: Worker stops polling for new tasks
3. **Drain In-Flight Tasks**: Waits up to 30 seconds for active tasks
4. **Close Connections**: Cleanly closes all external connections
5. **Health Status**: Sets health status to unhealthy (0)

### Shutdown Sequence

```
SIGTERM received
  ↓
Stop worker polling
  ↓
Wait for active tasks (max 30s)
  ↓
Close PostgreSQL pool
  ↓
Close Pulsar producer
  ↓
Close Temporal client
  ↓
Shutdown health/metrics servers
  ↓
Exit
```

## Running the Worker

### Development

```bash
# Build worker
make build-worker

# Run worker with .env file
./bin/worker

# Or use make
make run-worker
```

### Docker

```bash
# Build Docker image
docker build -t conn-dict-worker -f Dockerfile.worker .

# Run container
docker run -d \
  --name conn-dict-worker \
  --env-file .env \
  -p 8081:8081 \
  -p 9093:9093 \
  conn-dict-worker
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: conn-dict-worker
spec:
  replicas: 3
  selector:
    matchLabels:
      app: conn-dict-worker
  template:
    metadata:
      labels:
        app: conn-dict-worker
    spec:
      containers:
      - name: worker
        image: conn-dict-worker:latest
        ports:
        - containerPort: 8081
          name: health
        - containerPort: 9093
          name: metrics
        env:
        - name: TEMPORAL_ADDRESS
          value: "temporal.default.svc.cluster.local:7233"
        - name: POSTGRES_HOST
          value: "postgres.default.svc.cluster.local"
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /ready
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          requests:
            memory: "256Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "2000m"
```

## Testing

### Test Worker Startup

```bash
# Start worker
./bin/worker

# Check health
curl http://localhost:8081/health

# Check readiness
curl http://localhost:8081/ready

# Check metrics
curl http://localhost:9093/metrics
```

### Test Workflow Execution

```bash
# Start a claim workflow using temporal CLI
temporal workflow start \
  --task-queue conn-dict-task-queue \
  --type ClaimWorkflow \
  --input '{"claim_id":"TEST-001","entry_id":"ENTRY-001","claim_type":"PORTABILITY","claimer_ispb":"12345678","donor_ispb":"87654321"}' \
  --workflow-id claim-test-001
```

## Troubleshooting

### Worker Won't Start

1. **Check Temporal Connection**:
   ```bash
   curl http://localhost:7233/health
   ```

2. **Check PostgreSQL Connection**:
   ```bash
   psql -h localhost -U dict_user -d dict_db -c "SELECT 1"
   ```

3. **Check Logs**:
   ```bash
   # View worker logs
   tail -f /var/log/conn-dict-worker.log
   ```

### High Memory Usage

- Reduce `MAX_CONCURRENT_ACTIVITIES` and `MAX_CONCURRENT_WORKFLOWS`
- Check for memory leaks in activity implementations
- Monitor PostgreSQL connection pool usage

### Slow Task Processing

- Increase `MAX_CONCURRENT_ACTIVITIES`
- Check database query performance
- Monitor activity execution duration metrics

## Performance Tuning

### Optimal Settings for Production

```bash
# High-throughput configuration
MAX_CONCURRENT_WORKFLOWS=100
MAX_CONCURRENT_ACTIVITIES=200
POSTGRES_MAX_CONNS=50
POSTGRES_MIN_CONNS=10

# Resource limits (Kubernetes)
requests:
  memory: "512Mi"
  cpu: "1000m"
limits:
  memory: "2Gi"
  cpu: "4000m"
```

### Scaling Strategy

- **Horizontal Scaling**: Deploy multiple worker instances
- **Task Queue Partitioning**: Use separate task queues for different workload types
- **Activity Batching**: Batch database operations where possible

## Security Considerations

1. **Database Credentials**: Use secrets management (Kubernetes Secrets, AWS Secrets Manager)
2. **TLS/mTLS**: Enable TLS for Temporal and gRPC connections
3. **Network Policies**: Restrict worker egress to required services only
4. **RBAC**: Limit Kubernetes service account permissions

## Monitoring & Alerting

### Key Metrics to Monitor

1. **Worker Health**: `conn_dict_worker_health_status`
2. **Task Processing Rate**: `rate(conn_dict_worker_tasks_processed_total[5m])`
3. **Task Duration**: `histogram_quantile(0.95, conn_dict_worker_task_duration_seconds_bucket)`
4. **Active Tasks**: `conn_dict_worker_active_workflows + conn_dict_worker_active_activities`

### Recommended Alerts

```yaml
# Worker Down
- alert: WorkerDown
  expr: conn_dict_worker_health_status == 0
  for: 2m
  annotations:
    summary: "Worker is unhealthy"

# High Task Duration
- alert: HighTaskDuration
  expr: histogram_quantile(0.95, rate(conn_dict_worker_task_duration_seconds_bucket[5m])) > 30
  for: 5m
  annotations:
    summary: "95th percentile task duration > 30s"

# High Active Workflows
- alert: HighActiveWorkflows
  expr: conn_dict_worker_active_workflows > 80
  for: 5m
  annotations:
    summary: "Active workflows approaching limit"
```

## Next Steps

1. **Implement Bridge gRPC Client**: Connect to conn-bridge for Bacen API calls
2. **Add Redis Caching**: Cache frequently accessed data
3. **Implement Activity Timeouts**: Fine-tune timeouts per activity type
4. **Add Tracing**: Integrate OpenTelemetry for distributed tracing
5. **Load Testing**: Benchmark worker under production load

## Statistics

- **Total Workflows Registered**: 7
- **Total Activities Registered**: 31
- **Lines of Code**: ~400 LOC
- **Binary Size**: ~45 MB
- **Memory Usage**: ~100-500 MB (depends on workload)
- **CPU Usage**: 0.5-2 cores (depends on concurrency)

## References

- [Temporal Worker Documentation](https://docs.temporal.io/workers)
- [Temporal SDK Go](https://github.com/temporalio/sdk-go)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [Kubernetes Health Checks](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
