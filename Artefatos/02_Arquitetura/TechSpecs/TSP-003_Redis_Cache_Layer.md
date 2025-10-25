# TSP-003: Redis Cache Layer - Technical Specification

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: Redis Cache Layer
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: BACKEND (AI Agent - Backend Specialist)
**Revisor**: [Aguardando]
**Aprovador**: Tech Lead, Head de Arquitetura

---

## Sumário Executivo

Este documento especifica a implementação do **Redis Cache Layer** (v9.14.1) para o projeto DICT LBPay, cobrindo deployment em Kubernetes (StatefulSet), 5 estratégias de cache (entries, claims, workflows, idempotency, rate limiting), persistência RDB, e estratégias de monitoramento.

**Baseado em**:
- [DAT-005: Redis Cache Strategy](../../03_Dados/DAT-005_Redis_Cache_Strategy.md)
- [TEC-003: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [ADR-005: Caching Strategy](../ADR-005_Caching_Strategy.md) (pendente)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | BACKEND | Versão inicial - Redis v9.14.1 specification |

---

## Índice

1. [Visão Geral](#1-visão-geral)
2. [Deployment Kubernetes](#2-deployment-kubernetes)
3. [Cache Strategies](#3-cache-strategies)
4. [Persistence Configuration](#4-persistence-configuration)
5. [High Availability](#5-high-availability)
6. [Monitoring & Observability](#6-monitoring--observability)
7. [Performance Tuning](#7-performance-tuning)
8. [Security](#8-security)
9. [Disaster Recovery](#9-disaster-recovery)

---

## 1. Visão Geral

### 1.1. Redis Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Redis Cluster (v7.2.4)                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Redis Master (Primary)                                     │ │
│  │  - Port: 6379                                              │ │
│  │  - StatefulSet Pod 0                                       │ │
│  │  - Handles writes + reads                                  │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           ↓ (replication)                         │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Redis Replica 1                                           │ │
│  │  - Port: 6379                                              │ │
│  │  - StatefulSet Pod 1                                       │ │
│  │  - Handles reads only                                      │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           ↓ (replication)                         │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Redis Replica 2                                           │ │
│  │  - Port: 6379                                              │ │
│  │  - StatefulSet Pod 2                                       │ │
│  │  - Handles reads only                                      │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│               Redis Sentinel (Failover Manager)                  │
│  - Monitors master health                                        │
│  - Auto-failover on master failure                               │
│  - Promotes replica to master                                    │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│                  DICT Connect Application                        │
│  - go-redis/v9 v9.14.1 client                                    │
│  - Connection pool (100 conns)                                   │
│  - 5 cache strategies                                            │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2. Key Features

| Feature | Value | Justification |
|---------|-------|---------------|
| **Redis Version** | 7.2.4 | Latest stable (as of 2025-10-25) |
| **Client Library** | go-redis/v9 v9.14.1 | Official Go client |
| **Deployment** | Kubernetes StatefulSet | Persistent storage, stable network IDs |
| **HA Mode** | Redis Sentinel | Automatic failover |
| **Replicas** | 3 (1 master + 2 replicas) | Read scalability, fault tolerance |
| **Persistence** | RDB + AOF | Durability + performance |
| **Memory** | 4Gi per pod | Hot cache for 1M+ entries |
| **Eviction Policy** | allkeys-lru | Least Recently Used eviction |

---

## 2. Deployment Kubernetes

### 2.1. Namespace

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: redis
  labels:
    name: redis
    environment: production
```

### 2.2. Redis ConfigMap

```yaml
# k8s/redis-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
  namespace: redis
data:
  redis.conf: |
    # Network
    bind 0.0.0.0
    port 6379
    tcp-backlog 511
    timeout 0
    tcp-keepalive 300

    # General
    daemonize no
    supervised no
    pidfile /var/run/redis.pid
    loglevel notice
    logfile ""

    # Persistence - RDB
    save 900 1       # Save if 1 key changed in 900s
    save 300 10      # Save if 10 keys changed in 300s
    save 60 10000    # Save if 10000 keys changed in 60s
    stop-writes-on-bgsave-error yes
    rdbcompression yes
    rdbchecksum yes
    dbfilename dump.rdb
    dir /data

    # Persistence - AOF
    appendonly yes
    appendfilename "appendonly.aof"
    appendfsync everysec
    no-appendfsync-on-rewrite no
    auto-aof-rewrite-percentage 100
    auto-aof-rewrite-min-size 64mb

    # Memory Management
    maxmemory 3gb
    maxmemory-policy allkeys-lru
    maxmemory-samples 5

    # Lazy freeing
    lazyfree-lazy-eviction yes
    lazyfree-lazy-expire yes
    lazyfree-lazy-server-del yes
    replica-lazy-flush yes

    # Threaded I/O
    io-threads 4
    io-threads-do-reads yes

    # Replication
    replica-read-only yes
    repl-diskless-sync yes
    repl-diskless-sync-delay 5
    repl-disable-tcp-nodelay no
    replica-priority 100

    # Security
    requirepass REDIS_PASSWORD_PLACEHOLDER
```

### 2.3. Redis StatefulSet

```yaml
# k8s/redis-statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis
  namespace: redis
spec:
  serviceName: redis-headless
  replicas: 3
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7.2.4-alpine
        ports:
        - name: redis
          containerPort: 6379
          protocol: TCP
        command:
        - redis-server
        - /etc/redis/redis.conf
        - --requirepass
        - $(REDIS_PASSWORD)
        env:
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-secrets
              key: redis-password
        volumeMounts:
        - name: config
          mountPath: /etc/redis
        - name: data
          mountPath: /data
        resources:
          requests:
            memory: "2Gi"
            cpu: "500m"
          limits:
            memory: "4Gi"
            cpu: "1000m"
        livenessProbe:
          exec:
            command:
            - redis-cli
            - --pass
            - $(REDIS_PASSWORD)
            - ping
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
        readinessProbe:
          exec:
            command:
            - redis-cli
            - --pass
            - $(REDIS_PASSWORD)
            - ping
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
      volumes:
      - name: config
        configMap:
          name: redis-config
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 20Gi
```

### 2.4. Redis Services

```yaml
# k8s/redis-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: redis
spec:
  type: ClusterIP
  ports:
  - name: redis
    port: 6379
    targetPort: 6379
    protocol: TCP
  selector:
    app: redis
---
apiVersion: v1
kind: Service
metadata:
  name: redis-headless
  namespace: redis
spec:
  type: ClusterIP
  clusterIP: None
  ports:
  - name: redis
    port: 6379
    targetPort: 6379
  selector:
    app: redis
---
apiVersion: v1
kind: Service
metadata:
  name: redis-read
  namespace: redis
  annotations:
    service.alpha.kubernetes.io/tolerate-unready-endpoints: "false"
spec:
  type: ClusterIP
  ports:
  - name: redis
    port: 6379
    targetPort: 6379
  selector:
    app: redis
```

### 2.5. Redis Sentinel

```yaml
# k8s/redis-sentinel.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis-sentinel
  namespace: redis
spec:
  replicas: 3
  selector:
    matchLabels:
      app: redis-sentinel
  template:
    metadata:
      labels:
        app: redis-sentinel
    spec:
      containers:
      - name: sentinel
        image: redis:7.2.4-alpine
        ports:
        - name: sentinel
          containerPort: 26379
        command:
        - redis-sentinel
        - /etc/redis/sentinel.conf
        volumeMounts:
        - name: sentinel-config
          mountPath: /etc/redis
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "200m"
      volumes:
      - name: sentinel-config
        configMap:
          name: redis-sentinel-config
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-sentinel-config
  namespace: redis
data:
  sentinel.conf: |
    port 26379
    sentinel monitor mymaster redis-0.redis-headless.redis.svc.cluster.local 6379 2
    sentinel down-after-milliseconds mymaster 5000
    sentinel parallel-syncs mymaster 1
    sentinel failover-timeout mymaster 10000
    sentinel auth-pass mymaster REDIS_PASSWORD_PLACEHOLDER
---
apiVersion: v1
kind: Service
metadata:
  name: redis-sentinel
  namespace: redis
spec:
  type: ClusterIP
  ports:
  - name: sentinel
    port: 26379
    targetPort: 26379
  selector:
    app: redis-sentinel
```

### 2.6. Secrets

```yaml
# k8s/redis-secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: redis-secrets
  namespace: redis
type: Opaque
stringData:
  redis-password: <REPLACE_WITH_SECURE_PASSWORD>
```

---

## 3. Cache Strategies

### 3.1. Entry Cache (DICT Keys)

**Purpose**: Reduce PostgreSQL load for frequently accessed entries

**Implementation**:

```go
// internal/cache/entry_cache.go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type EntryCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewEntryCache(client *redis.Client) *EntryCache {
	return &EntryCache{
		client: client,
		ttl:    5 * time.Minute,
	}
}

// Cache by ID
func (c *EntryCache) GetByID(ctx context.Context, entryID string) (*Entry, error) {
	key := fmt.Sprintf("dict:entry:id:%s", entryID)

	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		// Cache miss
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (c *EntryCache) SetByID(ctx context.Context, entry *Entry) error {
	key := fmt.Sprintf("dict:entry:id:%s", entry.ID)

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return c.client.SetEx(ctx, key, data, c.ttl).Err()
}

// Cache by key (key_type + key_value)
func (c *EntryCache) GetByKey(ctx context.Context, keyType, keyValue string) (*Entry, error) {
	key := fmt.Sprintf("dict:entry:key:%s:%s", keyType, keyValue)

	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (c *EntryCache) SetByKey(ctx context.Context, entry *Entry) error {
	key := fmt.Sprintf("dict:entry:key:%s:%s", entry.KeyType, entry.KeyValue)

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return c.client.SetEx(ctx, key, data, c.ttl).Err()
}

// Invalidate all entry caches
func (c *EntryCache) Invalidate(ctx context.Context, entryID, keyType, keyValue string) error {
	keys := []string{
		fmt.Sprintf("dict:entry:id:%s", entryID),
		fmt.Sprintf("dict:entry:key:%s:%s", keyType, keyValue),
	}

	return c.client.Del(ctx, keys...).Err()
}
```

**TTL**: 5 minutes
**Policy**: Write-through (update cache on CREATE/UPDATE)

---

### 3.2. Claim Cache

**Purpose**: Optimize claim status queries during 30-day workflows

**Implementation**:

```go
// internal/cache/claim_cache.go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type ClaimCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewClaimCache(client *redis.Client) *ClaimCache {
	return &ClaimCache{
		client: client,
		ttl:    1 * time.Minute,
	}
}

// Full claim cache
func (c *ClaimCache) Get(ctx context.Context, claimID string) (*Claim, error) {
	key := fmt.Sprintf("dict:claim:id:%s", claimID)

	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}

	var claim Claim
	if err := json.Unmarshal(data, &claim); err != nil {
		return nil, err
	}

	return &claim, nil
}

func (c *ClaimCache) Set(ctx context.Context, claim *Claim) error {
	key := fmt.Sprintf("dict:claim:id:%s", claim.ID)

	data, err := json.Marshal(claim)
	if err != nil {
		return err
	}

	return c.client.SetEx(ctx, key, data, c.ttl).Err()
}

// Status-only cache (lighter)
func (c *ClaimCache) GetStatus(ctx context.Context, claimID string) (string, error) {
	key := fmt.Sprintf("dict:claim:status:%s", claimID)
	return c.client.Get(ctx, key).Result()
}

func (c *ClaimCache) SetStatus(ctx context.Context, claimID, status string) error {
	key := fmt.Sprintf("dict:claim:status:%s", claimID)
	return c.client.SetEx(ctx, key, status, c.ttl).Err()
}

// Index by entry_id
func (c *ClaimCache) AddToEntryIndex(ctx context.Context, entryID, claimID string) error {
	key := fmt.Sprintf("dict:claim:index:entry:%s", entryID)
	return c.client.SAdd(ctx, key, claimID).Err()
}

func (c *ClaimCache) GetByEntryID(ctx context.Context, entryID string) ([]string, error) {
	key := fmt.Sprintf("dict:claim:index:entry:%s", entryID)
	return c.client.SMembers(ctx, key).Result()
}
```

**TTL**: 1 minute (status changes frequently)
**Policy**: Write-through on Temporal events

---

### 3.3. Workflow State Cache

**Purpose**: Track Temporal workflow state without querying Temporal

**Implementation**:

```go
// internal/cache/workflow_cache.go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type WorkflowCache struct {
	client *redis.Client
}

func NewWorkflowCache(client *redis.Client) *WorkflowCache {
	return &WorkflowCache{client: client}
}

// Workflow status (short TTL)
func (c *WorkflowCache) GetStatus(ctx context.Context, workflowID string) (string, error) {
	key := fmt.Sprintf("workflow:claim:%s:status", workflowID)
	return c.client.Get(ctx, key).Result()
}

func (c *WorkflowCache) SetStatus(ctx context.Context, workflowID, status string) error {
	key := fmt.Sprintf("workflow:claim:%s:status", workflowID)
	return c.client.SetEx(ctx, key, status, 30*time.Second).Err()
}

// Workflow metadata (long TTL for audit)
func (c *WorkflowCache) GetMetadata(ctx context.Context, workflowID string) (*WorkflowMetadata, error) {
	key := fmt.Sprintf("workflow:claim:%s:metadata", workflowID)

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var metadata WorkflowMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (c *WorkflowCache) SetMetadata(ctx context.Context, workflowID string, metadata *WorkflowMetadata) error {
	key := fmt.Sprintf("workflow:claim:%s:metadata", workflowID)

	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	return c.client.SetEx(ctx, key, data, 30*24*time.Hour).Err()
}

type WorkflowMetadata struct {
	WorkflowID      string    `json:"workflow_id"`
	WorkflowType    string    `json:"workflow_type"`
	StartedAt       time.Time `json:"started_at"`
	ClaimID         string    `json:"claim_id"`
	CurrentActivity string    `json:"current_activity"`
}
```

**TTL**: 30 seconds (status), 30 days (metadata)

---

### 3.4. Idempotency Cache

**Purpose**: Prevent duplicate Bridge/Bacen requests

**Implementation**:

```go
// internal/cache/idempotency_cache.go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type IdempotencyCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewIdempotencyCache(client *redis.Client) *IdempotencyCache {
	return &IdempotencyCache{
		client: client,
		ttl:    24 * time.Hour,
	}
}

func (c *IdempotencyCache) Get(ctx context.Context, requestID string) (*BridgeResponse, error) {
	key := fmt.Sprintf("bridge:idempotency:%s", requestID)

	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}

	var response BridgeResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *IdempotencyCache) Set(ctx context.Context, requestID string, response *BridgeResponse) error {
	key := fmt.Sprintf("bridge:idempotency:%s", requestID)

	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return c.client.SetEx(ctx, key, data, c.ttl).Err()
}

type BridgeResponse struct {
	RequestID    string    `json:"request_id"`
	Operation    string    `json:"operation"`
	ClaimID      string    `json:"claim_id"`
	ExternalID   string    `json:"external_id"`
	Status       string    `json:"status"`
	ResponseCode string    `json:"response_code"`
	Timestamp    time.Time `json:"timestamp"`
}
```

**TTL**: 24 hours
**Policy**: Write-once (only on first successful response)

---

### 3.5. Rate Limiting Cache

**Purpose**: Protect Bridge/Bacen from overload

**Implementation**:

```go
// internal/cache/ratelimit_cache.go
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimitCache struct {
	client *redis.Client
	limits map[string]int
}

func NewRateLimitCache(client *redis.Client) *RateLimitCache {
	return &RateLimitCache{
		client: client,
		limits: map[string]int{
			"CreateEntry": 100,  // per minute per ISPB
			"CreateClaim": 50,   // per minute per ISPB
			"GetEntry":    500,  // per minute per ISPB
		},
	}
}

func (c *RateLimitCache) CheckLimit(ctx context.Context, ispb, operation string) error {
	// 1-minute window
	window := time.Now().Format("2006-01-02-15:04")
	key := fmt.Sprintf("ratelimit:%s:%s:%s", ispb, operation, window)

	// Increment counter
	count, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return err
	}

	// Set TTL on first increment
	if count == 1 {
		c.client.Expire(ctx, key, 1*time.Minute)
	}

	// Check limit
	limit, exists := c.limits[operation]
	if !exists {
		limit = 100 // default
	}

	if count > int64(limit) {
		return fmt.Errorf("rate limit exceeded: %d/%d requests/min", count, limit)
	}

	return nil
}

func (c *RateLimitCache) GetCount(ctx context.Context, ispb, operation string) (int64, error) {
	window := time.Now().Format("2006-01-02-15:04")
	key := fmt.Sprintf("ratelimit:%s:%s:%s", ispb, operation, window)
	return c.client.Get(ctx, key).Int64()
}
```

**Limits**:
- CreateEntry: 100 req/min per ISPB
- CreateClaim: 50 req/min per ISPB
- GetEntry: 500 req/min per ISPB

---

## 4. Persistence Configuration

### 4.1. RDB (Redis Database Backup)

**Configuration** (in `redis.conf`):

```conf
# Save RDB snapshots
save 900 1       # Save if 1 key changed in 900s (15 min)
save 300 10      # Save if 10 keys changed in 300s (5 min)
save 60 10000    # Save if 10000 keys changed in 60s (1 min)

rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
dir /data
```

**Pros**:
- Fast restart (load RDB snapshot)
- Compact file size
- Periodic backups to S3

**Cons**:
- Data loss possible (up to last save point)

---

### 4.2. AOF (Append-Only File)

**Configuration**:

```conf
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb
```

**Pros**:
- Minimal data loss (1 second max)
- Durability

**Cons**:
- Larger file size
- Slower restart

---

### 4.3. Hybrid Persistence (RDB + AOF)

**Recommended for Production**:
- Use **both** RDB and AOF
- AOF for durability
- RDB for fast recovery
- Trade-off: disk space vs safety

---

## 5. High Availability

### 5.1. Redis Sentinel Configuration

**Sentinel Features**:
- Master health monitoring
- Automatic failover (promote replica to master)
- Client reconfiguration
- Notification system

**Quorum**: 2 sentinels (out of 3) must agree on master failure

**Failover Timeout**: 10 seconds

---

### 5.2. Replication Topology

```
Master (redis-0)
  ├─> Replica 1 (redis-1)
  └─> Replica 2 (redis-2)
```

**Replication Mode**: Asynchronous
**Diskless Sync**: Enabled (faster replication)

---

### 5.3. Client Configuration for HA

```go
// internal/infrastructure/redis/client.go
package redis

import (
	"github.com/redis/go-redis/v9"
)

func NewClient(cfg Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Host,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     100,
		MinIdleConns: 10,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

// For production with Sentinel
func NewFailoverClient(cfg Config) *redis.Client {
	return redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    "mymaster",
		SentinelAddrs: []string{
			"redis-sentinel-0.redis-sentinel:26379",
			"redis-sentinel-1.redis-sentinel:26379",
			"redis-sentinel-2.redis-sentinel:26379",
		},
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     100,
		MinIdleConns: 10,
		MaxRetries:   3,
	})
}
```

---

## 6. Monitoring & Observability

### 6.1. Prometheus Metrics

**Exported Metrics** (via redis_exporter):

```yaml
# Connection metrics
redis_connected_clients
redis_blocked_clients

# Memory metrics
redis_memory_used_bytes
redis_memory_max_bytes
redis_memory_fragmentation_ratio

# Persistence metrics
redis_rdb_last_save_timestamp_seconds
redis_rdb_changes_since_last_save
redis_aof_current_size_bytes

# Replication metrics
redis_connected_slaves
redis_replication_lag_bytes

# Command metrics
redis_commands_processed_total
redis_commands_duration_seconds_total

# Cache metrics
redis_keyspace_hits_total
redis_keyspace_misses_total
redis_evicted_keys_total
redis_expired_keys_total
```

### 6.2. ServiceMonitor

```yaml
# k8s/redis-servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: redis
  namespace: redis
spec:
  selector:
    matchLabels:
      app: redis
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
```

### 6.3. Redis Exporter Deployment

```yaml
# k8s/redis-exporter.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis-exporter
  namespace: redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis-exporter
  template:
    metadata:
      labels:
        app: redis-exporter
    spec:
      containers:
      - name: redis-exporter
        image: oliver006/redis_exporter:v1.55.0-alpine
        ports:
        - name: metrics
          containerPort: 9121
        env:
        - name: REDIS_ADDR
          value: "redis:6379"
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-secrets
              key: redis-password
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
---
apiVersion: v1
kind: Service
metadata:
  name: redis-exporter
  namespace: redis
  labels:
    app: redis-exporter
spec:
  type: ClusterIP
  ports:
  - name: metrics
    port: 9121
    targetPort: 9121
  selector:
    app: redis-exporter
```

### 6.4. Alerting Rules

```yaml
# prometheus/redis-alerts.yaml
groups:
  - name: redis
    interval: 30s
    rules:
      - alert: RedisDown
        expr: up{job="redis"} == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Redis instance is down"
          description: "Redis instance {{ $labels.instance }} has been down for 2 minutes"

      - alert: RedisCacheHitRateLow
        expr: |
          rate(redis_keyspace_hits_total[5m]) /
          (rate(redis_keyspace_hits_total[5m]) + rate(redis_keyspace_misses_total[5m])) < 0.7
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Redis cache hit rate is low"
          description: "Cache hit rate is {{ $value | humanizePercentage }} (< 70%)"

      - alert: RedisMemoryHigh
        expr: redis_memory_used_bytes / redis_memory_max_bytes > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Redis memory usage is high"
          description: "Memory usage is {{ $value | humanizePercentage }} (> 90%)"

      - alert: RedisReplicationLag
        expr: redis_replication_lag_bytes > 10485760
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Redis replication lag is high"
          description: "Replication lag is {{ $value | humanize }}B (> 10MB)"

      - alert: RedisTooManyConnections
        expr: redis_connected_clients > 5000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Redis has too many connections"
          description: "Connected clients: {{ $value }} (> 5000)"
```

---

## 7. Performance Tuning

### 7.1. Memory Optimization

**Eviction Policy**: `allkeys-lru`
- Evict least recently used keys when maxmemory reached
- Suitable for cache use case

**Memory Samples**: 5
- Number of samples for LRU algorithm
- Higher = more accurate but slower

**Lazy Freeing**: Enabled
- Free memory asynchronously (non-blocking)
- Prevents latency spikes on DEL commands

---

### 7.2. Threaded I/O

**Configuration**:
```conf
io-threads 4
io-threads-do-reads yes
```

**Effect**:
- Parallelize I/O operations
- Better CPU utilization on multi-core systems
- Recommended: `io-threads = CPU cores - 1` (max 8)

---

### 7.3. Connection Pooling

**Client Configuration**:
- **PoolSize**: 100 connections
- **MinIdleConns**: 10 idle connections
- **PoolTimeout**: 4 seconds (max wait for available conn)

**Trade-off**:
- Too small: Connection exhaustion, high latency
- Too large: Memory overhead on Redis server

---

### 7.4. Pipeline Commands

**Use Case**: Batch multiple commands to reduce RTT

**Example**:
```go
pipe := client.Pipeline()
pipe.Set(ctx, "key1", "value1", 0)
pipe.Set(ctx, "key2", "value2", 0)
pipe.Set(ctx, "key3", "value3", 0)
_, err := pipe.Exec(ctx)
```

**Benefit**: 3 commands in 1 RTT instead of 3 RTTs

---

## 8. Security

### 8.1. Authentication

**Redis Password**: Required
- Stored in Kubernetes Secret
- Rotate every 90 days
- Never commit to Git

**ACL (Access Control Lists)**: (Future enhancement)
- Restrict commands per client
- Read-only users for monitoring

---

### 8.2. Network Isolation

**VPC Private Subnet**: Redis NOT exposed to internet
**Security Group**: Only DICT pods can access Redis (port 6379)
**TLS**: Optional (enable for Redis in AWS ElastiCache)

---

### 8.3. Encryption

**At Rest**:
- Redis RDB/AOF encrypted on disk (Kubernetes encrypted PVC)
- Use `dm-crypt` or cloud provider encryption

**In Transit**:
- TLS 1.2+ for Redis connections (optional, if remote)
- Certificate-based authentication

---

## 9. Disaster Recovery

### 9.1. Backup Strategy

**RDB Snapshots**:
- **Frequency**: Every 5 minutes (if 10+ keys changed)
- **Retention**: 7 days local, 30 days in S3
- **Location**: `/data/dump.rdb`

**Backup to S3** (CronJob):

```yaml
# k8s/redis-backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: redis-backup
  namespace: redis
spec:
  schedule: "0 */6 * * *"  # Every 6 hours
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: amazon/aws-cli:2.13.0
            command:
            - /bin/sh
            - -c
            - |
              apk add redis
              redis-cli -h redis -a $REDIS_PASSWORD --rdb /tmp/dump.rdb
              aws s3 cp /tmp/dump.rdb s3://lbpay-backups/redis/dump_$(date +%Y%m%d_%H%M%S).rdb
            env:
            - name: REDIS_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: redis-secrets
                  key: redis-password
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

---

### 9.2. Recovery Procedures

**Scenario 1: Single Pod Failure**
- **Action**: Kubernetes auto-restarts pod
- **RTO**: < 2 minutes
- **RPO**: 0 (replication from master)

**Scenario 2: Master Failure**
- **Action**: Sentinel promotes replica to master
- **RTO**: < 10 seconds (failover-timeout)
- **RPO**: < 1 second (AOF everysec)

**Scenario 3: Complete Cluster Loss**
- **Action**: Restore from S3 backup
- **RTO**: 30 minutes
- **RPO**: Up to 6 hours (backup frequency)

**Recovery Steps**:
1. Deploy new Redis StatefulSet
2. Download latest RDB from S3
3. Copy RDB to `/data/dump.rdb` in Redis pod
4. Restart Redis (automatically loads RDB)
5. Verify data integrity
6. Reconfigure Sentinel

---

## Rastreabilidade

### Requisitos Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RF-TSP-003-001 | Cache de entries (by ID, by key) | [DAT-005](../../03_Dados/DAT-005_Redis_Cache_Strategy.md) | ✅ Especificado |
| RF-TSP-003-002 | Cache de claims (status, index) | [DAT-005](../../03_Dados/DAT-005_Redis_Cache_Strategy.md) | ✅ Especificado |
| RF-TSP-003-003 | Cache de workflow state | [DAT-005](../../03_Dados/DAT-005_Redis_Cache_Strategy.md) | ✅ Especificado |
| RF-TSP-003-004 | Cache de idempotência (Bridge) | [DAT-005](../../03_Dados/DAT-005_Redis_Cache_Strategy.md) | ✅ Especificado |
| RF-TSP-003-005 | Rate limiting (ISPB, operation) | [DAT-005](../../03_Dados/DAT-005_Redis_Cache_Strategy.md) | ✅ Especificado |

### Requisitos Não-Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RNF-TSP-003-001 | HA: 99.9% availability (Sentinel) | SLA requirement | ✅ Especificado |
| RNF-TSP-003-002 | Cache hit rate > 70% | Performance goal | ✅ Especificado |
| RNF-TSP-003-003 | Latency p95 < 5ms | Performance goal | ✅ Especificado |
| RNF-TSP-003-004 | Persistence (RDB + AOF) | Durability | ✅ Especificado |
| RNF-TSP-003-005 | Backup to S3 (every 6h) | DR requirement | ✅ Especificado |

---

## Próximas Revisões

**Pendências**:
- [ ] Implementar Redis Cluster (sharding) para > 10M keys
- [ ] Configurar TLS para conexões Redis
- [ ] Implementar ACLs (per-user permissions)
- [ ] Criar Grafana dashboards completos
- [ ] Validar cache hit rates em ambiente real
- [ ] Implementar cache warming strategy
- [ ] Configurar Redis Streams para event sourcing

---

**Referências**:
- [DAT-005: Redis Cache Strategy](../../03_Dados/DAT-005_Redis_Cache_Strategy.md)
- [TEC-003: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [Redis Documentation](https://redis.io/docs/)
- [go-redis v9 Documentation](https://redis.uptrace.dev/)
- [Redis Sentinel Documentation](https://redis.io/docs/management/sentinel/)
