# ADR-004: Escolha de Cache e Rate Limiting - Redis

**Status**: ✅ Aceito
**Data**: 2025-10-24
**Decisores**: Thiago Lima (Head de Arquitetura), José Luís Silva (CTO)
**Contexto Técnico**: Projeto DICT - LBPay

---

## Controle de Versão

| Versão | Data | Autor | Descrição das Mudanças |
|--------|------|-------|------------------------|
| 1.0 | 2025-10-24 | ARCHITECT | Versão inicial - Documentação da decisão de usar Redis como cache e rate limiting |

---

## Status

**✅ ACEITO** - Redis já é tecnologia confirmada e em uso no LBPay

---

## Contexto

O projeto DICT da LBPay requer um **sistema de cache in-memory** e **rate limiting** para otimizar performance e proteger o sistema de sobrecarga. Os principais casos de uso são:

### Requisitos Funcionais

1. **Cache de Consultas DICT**:
   - Cachear respostas de `GetEntry` (consulta de chave PIX)
   - TTL: 5 minutos (dados podem estar desatualizados, mas aceitável para performance)
   - Hit rate target: ≥ 80%

2. **Cache de Validações**:
   - Cachear limites de chaves por CPF/CNPJ (ex: "CPF 123 tem 3 chaves cadastradas")
   - TTL: 5 minutos
   - Invalidação quando chave é criada/excluída

3. **Rate Limiting**:
   - Limitar requisições por ISPB (100 req/min para cadastro, 500 req/min para consulta)
   - Limitar requisições por usuário (10 req/min por endpoint)
   - Prevenir abuse/DoS

4. **Session Storage (OTP)**:
   - Armazenar OTP (email/telefone) com TTL (10 min para email, 5 min para SMS)
   - One-time use (delete após validação)

5. **Distributed Locks**:
   - Lock para operações críticas (ex: claim response - apenas 1 worker processa)
   - TTL automático (evitar deadlocks)

### Requisitos Não-Funcionais

| ID | Requisito | Target | Fonte |
|----|-----------|--------|-------|
| **NFR-110** | Latência cache (read) | ≤ 1ms (P95) | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-111** | Throughput | ≥ 50.000 ops/sec | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-112** | Disponibilidade | ≥ 99.99% | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-113** | Cache hit rate | ≥ 80% (consultas DICT) | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-114** | Eviction policy | LRU (Least Recently Used) | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-115** | Persistence | Optional (RDB snapshots) | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |

### Contexto Organizacional

- **LBPay já utiliza Redis** em múltiplos serviços (session storage, cache, rate limiting)
- Equipe de backend possui expertise em Redis
- Infraestrutura Redis provisionada (Redis Cluster)
- Redução de custo operacional (não introduzir nova tecnologia)

---

## Decisão

**Escolhemos Redis como solução de cache in-memory, rate limiting, e session storage para o projeto DICT.**

### Justificativa

Redis foi escolhido pelos seguintes motivos:

#### 1. **Já em Uso no LBPay**

✅ **Redis já é tecnologia estabelecida no LBPay**:
- Utilizado em Money Moving, Core Banking, Auth Service
- Redis Cluster provisionado e operacional
- Equipe treinada e experiente
- **Menor Time-to-Market** (não precisa provisionar nova stack)
- **Menor risco operacional** (tecnologia conhecida)

#### 2. **Performance Excepcional**

**Redis Benchmarks**:
- **Latência**: P95 < 1ms (local network)
- **Throughput**: 100.000+ ops/sec (single instance)
- **Throughput**: 1.000.000+ ops/sec (Redis Cluster)
- **In-memory**: Nenhum I/O de disco (exceto persistence)

**Comparação com Alternativas**:

| Aspecto | Redis | Memcached | DynamoDB | PostgreSQL (cache queries) |
|---------|-------|-----------|----------|----------------------------|
| **Latência (read)** | ✅ **< 1ms** | ✅ < 1ms | ⚠️ ~10ms | ❌ ~50ms |
| **Throughput** | ✅ **1M ops/sec** (cluster) | ⚠️ 500k ops/sec | ⚠️ Variable | ❌ 10k ops/sec |
| **Data Structures** | ✅ **Rich** (strings, hashes, lists, sets, sorted sets) | ❌ Key-value only | ⚠️ Limited | ❌ N/A |
| **TTL** | ✅ **Per-key** | ✅ Per-key | ✅ Per-item | ❌ Manual |
| **Atomic Operations** | ✅ **INCR, DECR, etc.** | ⚠️ Limited | ⚠️ Limited | ⚠️ Transações SQL |
| **Pub/Sub** | ✅ **Built-in** | ❌ No | ❌ No | ✅ LISTEN/NOTIFY |
| **Distributed Locks** | ✅ **Redlock** | ❌ No | ⚠️ Manual | ⚠️ Advisory locks |
| **Persistence** | ✅ **RDB, AOF** | ❌ No | ✅ Managed | ✅ Nativo |
| **High Availability** | ✅ **Redis Sentinel/Cluster** | ⚠️ Manual | ✅ Managed | ✅ Replicação |

#### 3. **Data Structures Ricas**

**Redis não é apenas key-value**:

1. **Strings** (cache simples):
```redis
SET dict:entry:12345678901 '{"name":"José","ispb":"99999999"}' EX 300  # TTL 5 min
GET dict:entry:12345678901
```

2. **Hashes** (objetos estruturados):
```redis
HSET dict:key:abc123 key_type CPF key_value 12345678901 status ACTIVE
HGETALL dict:key:abc123
```

3. **Sets** (unicidade, contagem):
```redis
SADD keys:cpf:12345678901 key_abc123 key_def456 key_ghi789
SCARD keys:cpf:12345678901  # Retorna: 3 (contagem de chaves)
```

4. **Sorted Sets** (ranking, leaderboards):
```redis
ZADD rate_limit:user:123 <timestamp> req_1
ZCOUNT rate_limit:user:123 <now - 60s> <now>  # Requests último minuto
```

5. **Lists** (filas, logs):
```redis
LPUSH audit:logs '{"event":"KeyRegistered","timestamp":"..."}'
LRANGE audit:logs 0 99  # Últimos 100 logs
```

#### 4. **Rate Limiting Eficiente**

**Sliding Window Algorithm (Redis)**:

```go
func CheckRateLimit(userID string, limit int, window time.Duration) (bool, error) {
    key := fmt.Sprintf("ratelimit:user:%s", userID)
    now := time.Now().UnixNano()
    windowStart := now - window.Nanoseconds()

    // 1. Remover requests antigas (fora da janela)
    rdb.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

    // 2. Contar requests na janela atual
    count, err := rdb.ZCount(ctx, key, fmt.Sprintf("%d", windowStart), "+inf").Result()
    if err != nil {
        return false, err
    }

    if count >= int64(limit) {
        return false, nil  // Rate limit excedido
    }

    // 3. Adicionar nova request
    rdb.ZAdd(ctx, key, redis.Z{
        Score:  float64(now),
        Member: fmt.Sprintf("req_%d", now),
    })

    // 4. Definir TTL (cleanup automático)
    rdb.Expire(ctx, key, window)

    return true, nil
}
```

**Vantagens**:
- ✅ **Precisão**: Sliding window (não fixed window)
- ✅ **Performance**: O(log N) (sorted set)
- ✅ **Cleanup automático**: TTL remove requests antigas
- ✅ **Distributed**: Funciona em ambiente multi-instância

#### 5. **Cache de Consultas DICT**

**Pattern: Cache-Aside (Lazy Loading)**:

```go
func GetEntry(keyValue string) (*DictEntry, error) {
    cacheKey := fmt.Sprintf("dict:entry:%s", keyValue)

    // 1. Tentar buscar no cache
    cached, err := rdb.Get(ctx, cacheKey).Result()
    if err == nil {
        // Cache hit
        var entry DictEntry
        json.Unmarshal([]byte(cached), &entry)
        return &entry, nil
    }

    // 2. Cache miss: buscar no DICT Bacen
    entry, err := GetEntryFromBacen(keyValue)
    if err != nil {
        return nil, err
    }

    // 3. Armazenar no cache (TTL 5 min)
    data, _ := json.Marshal(entry)
    rdb.Set(ctx, cacheKey, data, 5*time.Minute)

    return entry, nil
}
```

**Cache Invalidation** (quando chave é modificada):
```go
func InvalidateCacheOnKeyChange(keyValue string) {
    cacheKey := fmt.Sprintf("dict:entry:%s", keyValue)
    rdb.Del(ctx, cacheKey)
}
```

**Métricas**:
- Cache hit rate: `hits / (hits + misses)` ≥ 80%
- Latência: P95 < 1ms (cache hit), P95 < 300ms (cache miss + Bacen call)

#### 6. **Session Storage (OTP)**

**OTP Email (TTL 10 min)**:
```go
func StoreOTP(email string, otp string) error {
    key := fmt.Sprintf("otp:email:%s", email)
    return rdb.Set(ctx, key, otp, 10*time.Minute).Err()
}

func ValidateOTP(email string, otp string) (bool, error) {
    key := fmt.Sprintf("otp:email:%s", email)

    // 1. Buscar OTP armazenado
    storedOTP, err := rdb.Get(ctx, key).Result()
    if err == redis.Nil {
        return false, ErrOTPExpired
    }
    if err != nil {
        return false, err
    }

    // 2. Comparar OTP
    if storedOTP != otp {
        return false, nil
    }

    // 3. Deletar OTP (one-time use)
    rdb.Del(ctx, key)

    return true, nil
}
```

**Vantagens**:
- ✅ **TTL automático**: OTP expira após 10 min (não precisa cleanup manual)
- ✅ **One-time use**: Delete após validação (segurança)
- ✅ **Performance**: < 1ms (in-memory)

#### 7. **Distributed Locks (Redlock)**

**Uso**: Garantir que apenas 1 worker processa claim response (evitar double-processing)

```go
import "github.com/go-redsync/redsync/v4"

func ProcessClaimResponse(claimID string) error {
    // 1. Adquirir lock
    lockKey := fmt.Sprintf("lock:claim:%s", claimID)
    lock := redsync.NewMutex(lockKey, redsync.WithExpiry(30*time.Second))

    if err := lock.Lock(); err != nil {
        return ErrLockAcquisitionFailed
    }
    defer lock.Unlock()

    // 2. Processar claim (apenas 1 worker executa)
    return processClaim(claimID)
}
```

**Vantagens**:
- ✅ **Distributed**: Funciona em ambiente multi-instância
- ✅ **TTL automático**: Lock expira após 30s (evita deadlocks)
- ✅ **Safety**: Redlock algorithm (consensus-based)

#### 8. **High Availability (Redis Cluster)**

**Topologia Redis para Projeto DICT**:

```
Redis Cluster (LBPay Production)
│
├── Master 1 (shards: 0-5461)
│   └── Replica 1a
│   └── Replica 1b
│
├── Master 2 (shards: 5462-10922)
│   └── Replica 2a
│   └── Replica 2b
│
└── Master 3 (shards: 10923-16383)
    └── Replica 3a
    └── Replica 3b
```

**Características**:
- ✅ **Sharding automático**: Dados distribuídos entre 3 masters
- ✅ **Replicação**: 2 replicas por master (read scaling)
- ✅ **Failover automático**: Se master falha, replica é promovida
- ✅ **Disponibilidade**: 99.99% (tolerância a falhas)

#### 9. **Persistence (Opcional)**

**RDB Snapshots**:
- Snapshot periódico (ex: a cada 5 min se houver mudanças)
- Útil para recuperação após crash (não perde tudo)
- **Tradeoff**: Performance vs Durabilidade

**Para DICT**:
- **Cache**: Não precisa persistence (dados reconstruídos do PostgreSQL/Bacen)
- **Rate Limiting**: Não precisa persistence (reset após restart é aceitável)
- **OTP**: Não precisa persistence (TTL curto)
- **Conclusão**: **Desabilitar persistence** para máxima performance

---

## Consequências

### Positivas ✅

1. **Time-to-Market Reduzido**:
   - Redis já usado no LBPay
   - Equipe já treinada
   - Infraestrutura provisionada

2. **Performance Excepcional**:
   - Latência P95 < 1ms
   - Throughput: 1M ops/sec (cluster)
   - In-memory (zero disk I/O)

3. **Data Structures Ricas**:
   - Strings, hashes, sets, sorted sets, lists
   - Permite implementar rate limiting, cache, OTP, locks eficientemente

4. **Rate Limiting Built-in**:
   - Sliding window algorithm (sorted sets)
   - Distributed (funciona em múltiplas instâncias)

5. **High Availability**:
   - Redis Cluster (sharding + replicação)
   - Failover automático
   - 99.99% disponibilidade

6. **TTL Automático**:
   - Cleanup automático de dados expirados
   - OTP, rate limiting, cache invalidation

7. **Distributed Locks**:
   - Redlock algorithm
   - Previne race conditions em ambiente distribuído

### Negativas ❌

1. **Volatilidade** (sem persistence):
   - Dados perdidos em caso de crash (se persistence desabilitada)
   - **Mitigação**: Dados não-críticos (cache, rate limiting) - aceitável perder
   - **Mitigação**: Dados críticos (chaves PIX) persistidos no PostgreSQL

2. **Memory Constraints**:
   - Redis in-memory = limitado pela RAM
   - **Mitigação**: Eviction policy (LRU), monitoring de memória

3. **Single-threaded** (por instância):
   - CPU-bound operations bloqueiam outras requests
   - **Mitigação**: Redis Cluster (múltiplos masters)

4. **Debugging Complexo**:
   - Dados efêmeros (expiram)
   - **Mitigação**: Logging, monitoring (Redis Insights, Prometheus)

### Riscos e Mitigações

| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| **Memory exhaustion** | Média | Alto | Eviction policy (LRU), monitoring, alertas |
| **Cache stampede** | Média | Médio | Cache warming, probabilistic early expiration |
| **Redis cluster failure** | Baixa | Alto | Multi-master, replicas, failover automático |
| **Rate limiting bypass** | Baixa | Médio | Distribuited rate limiting, múltiplas camadas |

---

## Alternativas Consideradas

### Alternativa 1: Memcached

**Prós**:
- ✅ Performance similar (< 1ms latency)
- ✅ Simples (key-value puro)
- ✅ Multi-threaded (usa múltiplos cores)

**Contras**:
- ❌ **Não usado no LBPay**
- ❌ **Apenas key-value** (sem sorted sets, hashes, lists)
- ❌ **Sem TTL por key** (TTL global)
- ❌ **Sem distributed locks**
- ❌ **Sem persistence**
- ❌ **Sem Pub/Sub**
- ❌ Rate limiting complexo (não tem sorted sets)

**Decisão**: ❌ **Rejeitado** - Data structures limitadas, Redis já em uso

### Alternativa 2: DynamoDB (AWS)

**Prós**:
- ✅ Managed service (zero ops)
- ✅ Escalabilidade automática
- ✅ Persistence built-in

**Contras**:
- ❌ **Vendor lock-in** (AWS-only)
- ❌ **Latência superior** (~10ms vs 1ms Redis)
- ❌ **Custos variáveis** (pay-per-request)
- ❌ **Não usado no LBPay** (cache)
- ❌ **Rate limiting manual** (não tem sorted sets)

**Decisão**: ❌ **Rejeitado** - Lock-in, latência, custos

### Alternativa 3: Hazelcast

**Prós**:
- ✅ In-memory (performance similar)
- ✅ Distributed locks built-in
- ✅ Cache replication

**Contras**:
- ❌ **Não usado no LBPay**
- ❌ **Java-centric** (Go support limitado)
- ❌ **Complexidade operacional**
- ❌ **Overhead de configuração**

**Decisão**: ❌ **Rejeitado** - Redis já em uso, Java-centric

### Alternativa 4: PostgreSQL (Cache Queries)

**Prós**:
- ✅ Já usado (banco principal)
- ✅ Persistence built-in
- ✅ ACID

**Contras**:
- ❌ **Latência alta** (~50ms vs 1ms Redis)
- ❌ **Throughput limitado** (10k ops/sec vs 1M Redis)
- ❌ **Disk I/O** (não in-memory)
- ❌ **Rate limiting complexo** (queries SQL pesadas)

**Decisão**: ❌ **Rejeitado** - Performance insuficiente para cache

---

## Implementação

### Use Cases Redis no Projeto DICT

| Use Case | Data Structure | TTL | Exemplo Key |
|----------|----------------|-----|-------------|
| **Cache consultas DICT** | String (JSON) | 5 min | `dict:entry:12345678901` |
| **Cache limites chaves** | String (int) | 5 min | `key_count:cpf:12345678901` |
| **Rate limiting (ISPB)** | Sorted Set | 1 min | `ratelimit:ispb:99999999:register` |
| **Rate limiting (user)** | Sorted Set | 1 min | `ratelimit:user:user_123:register` |
| **OTP Email** | String | 10 min | `otp:email:user@example.com` |
| **OTP SMS** | String | 5 min | `otp:phone:+5511999998888` |
| **Distributed Lock** | String | 30 sec | `lock:claim:claim_abc123` |
| **Session (JWT blacklist)** | Set | Variable | `jwt:blacklist` |

### Exemplo 1: Cache de Consulta DICT

```go
package cache

import (
    "context"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"
)

type DictCache struct {
    rdb *redis.Client
}

func NewDictCache(rdb *redis.Client) *DictCache {
    return &DictCache{rdb: rdb}
}

func (c *DictCache) GetEntry(ctx context.Context, keyValue string) (*DictEntry, error) {
    cacheKey := fmt.Sprintf("dict:entry:%s", keyValue)

    // 1. Tentar cache
    cached, err := c.rdb.Get(ctx, cacheKey).Result()
    if err == nil {
        // Cache hit
        var entry DictEntry
        if err := json.Unmarshal([]byte(cached), &entry); err != nil {
            return nil, err
        }
        return &entry, nil
    }

    // 2. Cache miss: buscar do Bacen
    entry, err := c.fetchFromBacen(ctx, keyValue)
    if err != nil {
        return nil, err
    }

    // 3. Armazenar no cache (TTL 5 min)
    data, _ := json.Marshal(entry)
    c.rdb.Set(ctx, cacheKey, data, 5*time.Minute)

    return entry, nil
}

func (c *DictCache) InvalidateEntry(ctx context.Context, keyValue string) error {
    cacheKey := fmt.Sprintf("dict:entry:%s", keyValue)
    return c.rdb.Del(ctx, cacheKey).Err()
}
```

### Exemplo 2: Rate Limiting (Sliding Window)

```go
package ratelimit

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type RateLimiter struct {
    rdb *redis.Client
}

func NewRateLimiter(rdb *redis.Client) *RateLimiter {
    return &RateLimiter{rdb: rdb}
}

func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
    now := time.Now().UnixNano()
    windowStart := now - window.Nanoseconds()

    pipe := rl.rdb.Pipeline()

    // 1. Remover requests antigas
    pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

    // 2. Contar requests na janela
    countCmd := pipe.ZCount(ctx, key, fmt.Sprintf("%d", windowStart), "+inf")

    // 3. Adicionar nova request
    pipe.ZAdd(ctx, key, redis.Z{
        Score:  float64(now),
        Member: fmt.Sprintf("req_%d", now),
    })

    // 4. Definir TTL
    pipe.Expire(ctx, key, window)

    _, err := pipe.Exec(ctx)
    if err != nil {
        return false, err
    }

    count := countCmd.Val()
    return count < int64(limit), nil
}

// Uso
func CheckUserRateLimit(userID string) error {
    limiter := NewRateLimiter(redisClient)
    key := fmt.Sprintf("ratelimit:user:%s:register", userID)

    allowed, err := limiter.Allow(ctx, key, 10, 1*time.Minute)
    if err != nil {
        return err
    }

    if !allowed {
        return ErrRateLimitExceeded
    }

    return nil
}
```

### Exemplo 3: OTP Storage

```go
package otp

import (
    "context"
    "crypto/rand"
    "fmt"
    "math/big"
    "time"

    "github.com/redis/go-redis/v9"
)

type OTPService struct {
    rdb *redis.Client
}

func NewOTPService(rdb *redis.Client) *OTPService {
    return &OTPService{rdb: rdb}
}

func (s *OTPService) GenerateAndStoreOTP(ctx context.Context, email string) (string, error) {
    // Gerar OTP (6 dígitos)
    otp := generateOTP(6)

    // Armazenar no Redis (TTL 10 min)
    key := fmt.Sprintf("otp:email:%s", email)
    if err := s.rdb.Set(ctx, key, otp, 10*time.Minute).Err(); err != nil {
        return "", err
    }

    return otp, nil
}

func (s *OTPService) ValidateOTP(ctx context.Context, email string, otp string) (bool, error) {
    key := fmt.Sprintf("otp:email:%s", email)

    // Buscar OTP armazenado
    storedOTP, err := s.rdb.Get(ctx, key).Result()
    if err == redis.Nil {
        return false, ErrOTPExpired
    }
    if err != nil {
        return false, err
    }

    // Comparar OTP
    if storedOTP != otp {
        return false, nil
    }

    // Deletar OTP (one-time use)
    s.rdb.Del(ctx, key)

    return true, nil
}

func generateOTP(length int) string {
    digits := "0123456789"
    otp := make([]byte, length)
    for i := range otp {
        num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
        otp[i] = digits[num.Int64()]
    }
    return string(otp)
}
```

### Exemplo 4: Distributed Lock (Redlock)

```go
package lock

import (
    "time"

    "github.com/go-redsync/redsync/v4"
    "github.com/go-redsync/redsync/v4/redis/goredis/v9"
    goredislib "github.com/redis/go-redis/v9"
)

type LockService struct {
    rs *redsync.Redsync
}

func NewLockService(rdb *goredislib.Client) *LockService {
    pool := goredis.NewPool(rdb)
    rs := redsync.New(pool)
    return &LockService{rs: rs}
}

func (s *LockService) WithLock(lockKey string, ttl time.Duration, fn func() error) error {
    // Criar mutex
    mutex := s.rs.NewMutex(lockKey, redsync.WithExpiry(ttl))

    // Adquirir lock
    if err := mutex.Lock(); err != nil {
        return fmt.Errorf("failed to acquire lock: %w", err)
    }
    defer mutex.Unlock()

    // Executar função protegida
    return fn()
}

// Uso
func ProcessClaimResponse(claimID string) error {
    lockSvc := NewLockService(redisClient)
    lockKey := fmt.Sprintf("lock:claim:%s", claimID)

    return lockSvc.WithLock(lockKey, 30*time.Second, func() error {
        // Apenas 1 worker executa isso
        return processClaim(claimID)
    })
}
```

### Configuração Redis Client (Go)

```go
package config

import (
    "github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:         "redis-cluster.lbpay.svc.cluster.local:6379",
        Password:     "",  // No password (internal network)
        DB:           0,   // Default DB
        PoolSize:     100, // Connection pool
        MinIdleConns: 10,
        MaxRetries:   3,
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
    })
}

// Redis Cluster (para high availability)
func NewRedisClusterClient() *redis.ClusterClient {
    return redis.NewClusterClient(&redis.ClusterOptions{
        Addrs: []string{
            "redis-master1.lbpay.svc.cluster.local:6379",
            "redis-master2.lbpay.svc.cluster.local:6379",
            "redis-master3.lbpay.svc.cluster.local:6379",
        },
        PoolSize:     100,
        MinIdleConns: 10,
        MaxRetries:   3,
    })
}
```

### Monitoramento

**Métricas Prometheus**:
- `redis_connected_clients` (conexões ativas)
- `redis_used_memory_bytes` (uso de memória)
- `redis_evicted_keys_total` (keys evicted por LRU)
- `redis_keyspace_hits_total` (cache hits)
- `redis_keyspace_misses_total` (cache misses)
- `redis_commands_processed_total` (ops/sec)

**Métricas Customizadas** (via Redis SDK):
```go
// Cache hit rate
hit_rate = hits / (hits + misses)

// Latência por operação
redis_operation_latency_ms{operation="GET"} histogram
```

**Alertas**:
- Memory usage > 80%
- Cache hit rate < 70%
- Evicted keys > 1000/min
- Connection failures > 10/min

---

## Rastreabilidade

### Requisitos Funcionais Impactados

| CRF | Descrição | Uso Redis |
|-----|-----------|-----------|
| [CRF-050](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-050) | Consultar Chave DICT | Cache de respostas (TTL 5 min) |
| [CRF-012](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-012) | Validar Limites Chaves | Cache de contagem (TTL 5 min) |
| [CRF-003](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-003) | Cadastro Email (OTP) | Armazenar OTP (TTL 10 min) |
| [CRF-004](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-004) | Cadastro Telefone (OTP) | Armazenar OTP (TTL 5 min) |
| [CRF-100](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-100) | Rate Limiting | Sliding window (sorted sets) |

### NFRs Impactados

| NFR | Descrição | Como Redis Atende |
|-----|-----------|-------------------|
| [NFR-110](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-110) | Latência < 1ms | Redis: P95 < 1ms ✅ |
| [NFR-111](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-111) | Throughput ≥ 50k ops/sec | Redis Cluster: 1M ops/sec ✅ |
| [NFR-113](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-113) | Cache hit rate ≥ 80% | Monitoring + tuning TTL ✅ |

---

## Referências

- [Redis Documentation](https://redis.io/docs/)
- [Redis Go Client](https://github.com/redis/go-redis)
- [Redlock Algorithm](https://redis.io/docs/manual/patterns/distributed-locks/)
- [ArquiteturaDict_LBPAY.md](../../Docs_iniciais/ArquiteturaDict_LBPAY.md): Diagramas SVG mostrando Redis

---

## Aprovação

- [x] **Thiago Lima** (Head de Arquitetura) - 2025-10-24
- [x] **José Luís Silva** (CTO) - 2025-10-24

**Rationale**: Redis já é tecnologia confirmada e em uso no LBPay. Esta ADR documenta a decisão e fundamenta o uso técnico no projeto DICT.

---

**FIM DO DOCUMENTO ADR-004**
