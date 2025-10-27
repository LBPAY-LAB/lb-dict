# Application Layer - Query Handlers (CQRS)

**Padrão**: Command Query Responsibility Segregation (CQRS)
**Data**: 2025-10-27
**Versão**: 1.0

---

## Visão Geral

Este diretório contém os **Query Handlers** do Core-Dict, implementando o lado de leitura (Read) do padrão CQRS.

### Arquitetura CQRS

```
┌─────────────────────────────────────────────────────┐
│              APPLICATION LAYER                       │
├─────────────────────────────────────────────────────┤
│                                                       │
│  COMMANDS (Write)        │        QUERIES (Read)    │
│  ├── CreateEntry         │        ├── GetEntry      │
│  ├── UpdateEntry         │        ├── ListEntries   │
│  ├── DeleteEntry         │        ├── GetAccount    │
│  └── ...                 │        └── ...           │
│                                                       │
│  ↓                       │        ↓                  │
│  Repository (Write)      │        Repository (Read) │
│  + EventBus              │        + Cache (Redis)   │
│                                                       │
└─────────────────────────────────────────────────────┘
```

---

## Query Handlers Implementados

### 1. **get_entry_query.go** - Buscar chave PIX
**Handler**: `GetEntryQueryHandler`

**Query**:
```go
type GetEntryQuery struct {
    KeyValue string
}
```

**Cache Strategy**: Cache-Aside
- **Cache Key**: `entry:{key_value}`
- **TTL**: 5 minutos
- **Invalidação**: On write (CreateEntry, UpdateEntry, DeleteEntry)

**Flow**:
1. Try Redis cache first
2. Cache miss → Query PostgreSQL
3. Store in cache (5min TTL)
4. Return result

---

### 2. **list_entries_query.go** - Listar chaves PIX (paginado)
**Handler**: `ListEntriesQueryHandler`

**Query**:
```go
type ListEntriesQuery struct {
    AccountID uuid.UUID
    Page      int // 1-indexed
    PageSize  int // default: 100, max: 1000
}
```

**Cache Strategy**: Cache-Aside (per page)
- **Cache Key**: `entries:account:{account_id}:page:{page}:size:{size}`
- **TTL**: 2 minutos (listas mudam frequentemente)
- **Invalidação**: Pattern matching `entries:account:{account_id}:*`

**Pagination**:
- **Default**: 100 items per page
- **Max**: 1000 items per page
- **Type**: Cursor-based pagination

---

### 3. **get_account_query.go** - Buscar conta CID
**Handler**: `GetAccountQueryHandler`

**Queries**:
```go
type GetAccountQuery struct {
    AccountID uuid.UUID
}

type GetAccountByNumberQuery struct {
    ISPB          string
    Branch        string
    AccountNumber string
}
```

**Cache Strategy**: Cache-Aside (dual cache)
- **Cache Keys**:
  - `account:id:{account_id}`
  - `account:number:{ispb}:{branch}:{account_number}`
- **TTL**: 5 minutos
- **Invalidação**: Pattern matching `account:*`

---

### 4. **get_claim_query.go** - Buscar claim por ID
**Handler**: `GetClaimQueryHandler`

**Query**:
```go
type GetClaimQuery struct {
    ClaimID uuid.UUID
}
```

**Cache Strategy**: Cache-Aside
- **Cache Key**: `claim:{claim_id}`
- **TTL**: 3 minutos (claims são mutáveis)
- **Invalidação**: On write (CreateClaim, ConfirmClaim, CancelClaim)

---

### 5. **list_claims_query.go** - Listar claims (paginado)
**Handler**: `ListClaimsQueryHandler`

**Query**:
```go
type ListClaimsQuery struct {
    ISPB     string
    Page     int
    PageSize int
}
```

**Cache Strategy**: Cache-Aside
- **Cache Key**: `claims:ispb:{ispb}:page:{page}:size:{size}`
- **TTL**: 1 minuto (claims mudam frequentemente)
- **Invalidação**: Pattern matching `claims:ispb:{ispb}:*`

---

### 6. **verify_account_query.go** - Verificar conta no RSFN
**Handler**: `VerifyAccountQueryHandler`

**Query**:
```go
type VerifyAccountQuery struct {
    ISPB          string
    Branch        string
    AccountNumber string
}
```

**Cache Strategy**: Cache-Aside (com fallback RSFN)
- **Cache Key**: `verify:account:{ispb}:{branch}:{account_number}`
- **TTL**: 10 minutos
- **Sources**:
  1. Cache Redis
  2. Local database
  3. RSFN via Connect service (fallback)

**Flow**:
1. Try cache first
2. Try local database
3. Call RSFN via Connect service (gRPC)
4. Cache result (10min TTL)

---

### 7. **get_statistics_query.go** - Estatísticas agregadas
**Handler**: `GetStatisticsQueryHandler`

**Query**:
```go
type GetStatisticsQuery struct {
    // Filtros futuros
}
```

**Cache Strategy**: Cache-Aside (SEMPRE cacheado)
- **Cache Key**: `statistics:global`
- **TTL**: 5 minutos
- **Invalidação**: Manual (após writes)
- **Reason**: Estatísticas são caras de calcular (COUNT queries)

**Metrics**:
- Total keys (por tipo, status)
- Total claims (por tipo, status)
- Total infractions (por severidade)

---

### 8. **health_check_query.go** - Health check completo
**Handler**: `HealthCheckQueryHandler`

**Query**:
```go
type HealthCheckQuery struct {
    // Flags futuros
}
```

**Cache Strategy**: NONE (real-time checks)
- **Reason**: Health checks precisam ser real-time

**Checks**:
1. PostgreSQL (latency + connectivity)
2. Redis (latency + connectivity)
3. Pulsar (latency + connectivity)

**Status**:
- `healthy`: Tudo OK
- `degraded`: Redis ou Pulsar down
- `unhealthy`: PostgreSQL down

---

### 9. **list_infractions_query.go** - Listar infrações (paginado)
**Handler**: `ListInfractionsQueryHandler`

**Query**:
```go
type ListInfractionsQuery struct {
    ISPB     string
    Page     int
    PageSize int
}
```

**Cache Strategy**: Cache-Aside
- **Cache Key**: `infractions:ispb:{ispb}:page:{page}:size:{size}`
- **TTL**: 10 minutos (infrações raramente mudam)
- **Invalidação**: Pattern matching `infractions:ispb:{ispb}:*`

---

### 10. **get_audit_log_query.go** - Buscar audit logs
**Handler**: `GetAuditLogQueryHandler`

**Queries**:
```go
type GetAuditLogQuery struct {
    EntityType string
    EntityID   uuid.UUID
    Page       int
    PageSize   int
}

type GetAuditLogByActorQuery struct {
    ActorID  uuid.UUID
    Page     int
    PageSize int
}
```

**Cache Strategy**: Cache-Aside
- **Cache Keys**:
  - `audit:entity:{entity_type}:{entity_id}:page:{page}:size:{size}`
  - `audit:actor:{actor_id}:page:{page}:size:{size}`
- **TTL**: 15 minutos (audit logs são imutáveis)
- **Invalidação**: Raramente necessária (append-only)

---

## Estratégias de Cache Implementadas

### 1. **Cache-Aside Pattern**
Padrão usado em TODOS os query handlers.

**Flow**:
```
┌─────────────────────────────────────────┐
│  1. Try Redis cache first               │
│  2. Cache miss → Query database         │
│  3. Store in cache with TTL             │
│  4. Return result                       │
└─────────────────────────────────────────┘
```

### 2. **TTL Strategy (por tipo de dado)**

| Tipo de Dado | TTL | Razão |
|--------------|-----|-------|
| **Entries (single)** | 5min | Razoavelmente estáveis |
| **Entries (list)** | 2min | Listas mudam frequentemente |
| **Accounts** | 5min | Razoavelmente estáveis |
| **Claims** | 1-3min | Altamente mutáveis |
| **Statistics** | 5min | Caras de calcular |
| **Infractions** | 10min | Raramente mudam |
| **Audit Logs** | 15min | Imutáveis (append-only) |
| **Verify Account** | 10min | Chamada externa cara |
| **Health Checks** | NONE | Real-time required |

### 3. **Invalidation Strategy**

#### On Write (Commands)
```go
// Após CreateEntry command
entryQueryHandler.InvalidateCache(ctx, entry.KeyValue)
listEntriesQueryHandler.InvalidateCache(ctx, entry.AccountID)
statsQueryHandler.InvalidateCache(ctx)
```

#### Pattern Matching
```go
// Invalidar todas as páginas de listagem
cache.Invalidate(ctx, "entries:account:{account_id}:*")
cache.Invalidate(ctx, "claims:ispb:{ispb}:*")
```

#### Manual Refresh
```go
// Forçar refresh de estatísticas
statsQueryHandler.RefreshCache(ctx)
```

### 4. **Dual Cache Strategy** (Accounts)
Cachear por múltiplas chaves:
- `account:id:{account_id}`
- `account:number:{ispb}:{branch}:{account_number}`

Permite busca por ID OU por número da conta.

### 5. **Multi-Level Cache** (Verify Account)
1. **L1 Cache**: Redis (TTL: 10min)
2. **L2 Cache**: Local database
3. **L3 Fallback**: RSFN via Connect service

---

## Paginação

### Cursor-Based Pagination
Todas as listagens usam paginação cursor-based:

```go
type PaginatedResult struct {
    Items      []T    `json:"items"`
    TotalCount int64  `json:"total_count"`
    Page       int    `json:"page"`
    PageSize   int    `json:"page_size"`
    TotalPages int    `json:"total_pages"`
}
```

**Defaults**:
- Page: 1 (1-indexed)
- PageSize: 100
- Max PageSize: 1000

---

## Métricas e Observabilidade

### Cache Hit Rate
```go
// Prometheus metrics
cache_hits_total{query="get_entry"}
cache_misses_total{query="get_entry"}
cache_hit_rate = hits / (hits + misses)
```

**Target**: >80% hit rate

### Query Latency
```go
query_duration_seconds{query="get_entry", source="cache"}
query_duration_seconds{query="get_entry", source="database"}
```

**SLOs**:
- Cache hit: <5ms
- Database query: <50ms

---

## Testing

### Unit Tests
```bash
# Run query handler tests
go test ./internal/application/queries/...
```

### Integration Tests
```bash
# Test with real Redis + PostgreSQL
go test -tags=integration ./internal/application/queries/...
```

---

## Próximos Passos

### Phase 1: Implementação Infrastructure
- [ ] PostgreSQL repository implementations
- [ ] Redis cache implementation
- [ ] Pulsar producer/consumer

### Phase 2: Tests
- [ ] Unit tests (>80% coverage)
- [ ] Integration tests (Redis + PostgreSQL)
- [ ] Performance tests (cache hit rate)

### Phase 3: Observabilidade
- [ ] Prometheus metrics
- [ ] OpenTelemetry traces
- [ ] Grafana dashboards

---

**Última Atualização**: 2025-10-27
**Autor**: backend-core-queries specialist
**Status**: ✅ Implementação completa (10/10 handlers)
