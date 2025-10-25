# DAT-005: Redis Cache Strategy

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Versão**: 1.0
**Data**: 2025-10-25
**Status**: ✅ Especificação Completa
**Responsável**: ARCHITECT (AI Agent - Technical Architect)

---

## 📋 Resumo Executivo

Este documento especifica a **estratégia completa de cache com Redis** para o sistema DICT, incluindo estrutura de chaves, TTL (Time-To-Live), políticas de invalidação, serialização de dados e monitoramento.

**Objetivo**: Documentar o uso do Redis (v9.14.1) já implementado no Connect para otimização de performance e redução de carga no PostgreSQL e nas chamadas ao Bridge/Bacen.

---

## 🎯 Objetivos do Cache

### 1. Performance
- **Reduzir latência** de consultas repetidas (GET entry by key)
- **Evitar roundtrips** desnecessários ao Bridge/Bacen
- **Otimizar workflows** do Temporal (status de claims, idempotência)

### 2. Resiliência
- **Idempotência** de requests ao Bridge (evitar duplicação de claims)
- **Cache de fallback** quando Bridge está temporariamente indisponível
- **Rate limiting** para proteger APIs downstream

### 3. Escalabilidade
- **Offload do PostgreSQL** para queries de leitura frequentes
- **Session storage** para workflows de longa duração (30 dias)
- **Distributed cache** compartilhado entre múltiplas instâncias do Connect

---

## 🔑 Estrutura de Chaves Redis

### Convenção de Nomenclatura

```
{namespace}:{entity}:{identifier}:{attribute}
```

**Regras**:
- Usar `:` como separador
- Sempre prefixar com namespace (`dict`, `workflow`, `bridge`)
- Incluir versão quando necessário (`v1`, `v2`)
- Usar lowercase e snake_case

---

## 📊 Tipos de Cache Implementados

### 1. Cache de Entries (Chaves DICT)

**Propósito**: Reduzir consultas ao PostgreSQL para entries frequentemente acessadas

#### Estrutura
```redis
# Cache por ID
dict:entry:id:{entry_id} → JSON
TTL: 5 minutos

# Cache por chave (key_type + key_value)
dict:entry:key:{key_type}:{key_value} → JSON
TTL: 5 minutos

# Exemplo:
dict:entry:id:550e8400-e29b-41d4-a716-446655440000 → {
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "key_type": "CPF",
  "key_value": "12345678900",
  "account_ispb": "00000000",
  "account_number": "12345-6",
  "status": "ACTIVE",
  "created_at": "2025-10-25T10:00:00Z"
}

dict:entry:key:CPF:12345678900 → {same JSON}
```

#### Políticas
- **Write-through**: Atualizar cache imediatamente após CREATE/UPDATE no PostgreSQL
- **TTL**: 5 minutos (balanceamento entre freshness e performance)
- **Invalidação**: DELETE imediata ao deletar entry ou mudar status

---

### 2. Cache de Claims (Reivindicações)

**Propósito**: Otimizar consultas de status de claims (workflows de 30 dias)

#### Estrutura
```redis
# Cache de claim completo
dict:claim:id:{claim_id} → JSON
TTL: 1 minuto (status muda frequentemente)

# Cache de status apenas (mais leve)
dict:claim:status:{claim_id} → "OPEN"|"WAITING_RESOLUTION"|"CONFIRMED"|"CANCELLED"|"COMPLETED"|"EXPIRED"
TTL: 1 minuto

# Índice por entry_id (para listar claims de uma entry)
dict:claim:index:entry:{entry_id} → SET de claim_ids
TTL: 5 minutos

# Exemplo:
dict:claim:id:c1234567-89ab-cdef-0123-456789abcdef → {
  "id": "c1234567-89ab-cdef-0123-456789abcdef",
  "workflow_id": "claim-workflow-123",
  "entry_id": "550e8400-e29b-41d4-a716-446655440000",
  "claimer_ispb": "11111111",
  "owner_ispb": "00000000",
  "status": "OPEN",
  "expires_at": "2025-11-24T10:00:00Z",
  "completion_period_days": 30
}
```

#### Políticas
- **Write-through**: Atualizar cache ao receber eventos do Temporal
- **TTL Curto**: 1 minuto (status de claim muda com ações do usuário)
- **Invalidação**: DELETE ao completar/cancelar claim

---

### 3. Cache de Workflow State (Temporal)

**Propósito**: Rastrear estado de workflows em execução sem consultar Temporal

#### Estrutura
```redis
# Status de workflow
workflow:claim:{workflow_id}:status → "RUNNING"|"COMPLETED"|"FAILED"|"TIMED_OUT"
TTL: 30 segundos

# Metadata de workflow (para debugging)
workflow:claim:{workflow_id}:metadata → JSON
TTL: 30 dias (manter histórico completo)

# Exemplo:
workflow:claim:claim-workflow-123:status → "RUNNING"

workflow:claim:claim-workflow-123:metadata → {
  "workflow_id": "claim-workflow-123",
  "workflow_type": "ClaimWorkflow",
  "started_at": "2025-10-25T10:00:00Z",
  "claim_id": "c1234567-89ab-cdef-0123-456789abcdef",
  "current_activity": "WaitForOwnerResponse"
}
```

#### Políticas
- **Write-through**: Atualizar ao iniciar/completar activities
- **TTL Curto para status**: 30 segundos (pode consultar Temporal se expirar)
- **TTL Longo para metadata**: 30 dias (auditoria completa)

---

### 4. Cache de Idempotência (Bridge Requests)

**Propósito**: Evitar duplicação de requests ao Bridge/Bacen (CreateClaim, CreateEntry)

#### Estrutura
```redis
# Idempotency key para requests
bridge:idempotency:{request_id} → JSON (response do Bridge)
TTL: 24 horas

# Exemplo:
bridge:idempotency:req-20251025-100000-abc123 → {
  "request_id": "req-20251025-100000-abc123",
  "operation": "CreateClaim",
  "claim_id": "c1234567-89ab-cdef-0123-456789abcdef",
  "external_id": "BACEN-CLAIM-987654",
  "status": "SUCCESS",
  "response_code": "00",
  "timestamp": "2025-10-25T10:00:00Z"
}
```

#### Políticas
- **Write-once**: Escrever apenas na primeira execução bem-sucedida
- **TTL**: 24 horas (tempo suficiente para retries transientes)
- **Read-before-write**: Sempre verificar cache antes de chamar Bridge

#### Algoritmo de Idempotência
```go
// Pseudocódigo (especificação, NÃO implementar agora)
func CreateClaimIdempotent(ctx context.Context, req *CreateClaimRequest) (*CreateClaimResponse, error) {
    // 1. Gerar idempotency key
    idempotencyKey := fmt.Sprintf("bridge:idempotency:%s", req.RequestID)

    // 2. Tentar buscar do cache
    cachedResponse, err := redis.Get(ctx, idempotencyKey)
    if err == nil {
        // Cache hit - retornar resposta anterior
        return cachedResponse, nil
    }

    // 3. Cache miss - chamar Bridge
    response, err := bridgeClient.CreateClaim(ctx, req)
    if err != nil {
        return nil, err
    }

    // 4. Salvar no cache (TTL 24h)
    redis.SetEx(ctx, idempotencyKey, response, 24*time.Hour)

    return response, nil
}
```

---

### 5. Cache de Rate Limiting

**Propósito**: Proteger APIs downstream (Bridge, Bacen) de sobrecarga

#### Estrutura
```redis
# Contador de requests por ISPB
ratelimit:{ispb}:{operation}:{window} → contador INT
TTL: Duração da janela (1 minuto, 1 hora)

# Exemplo:
ratelimit:00000000:CreateEntry:2025-10-25-10:00 → 42
TTL: 1 minuto

ratelimit:00000000:CreateClaim:2025-10-25-10 → 158
TTL: 1 hora
```

#### Limites Recomendados
- **CreateEntry**: 100 req/min por ISPB, 1000 req/hour por ISPB
- **CreateClaim**: 50 req/min por ISPB, 500 req/hour por ISPB
- **GetEntry**: 500 req/min por ISPB (mais permissivo, leitura)

#### Algoritmo
```go
// Pseudocódigo (especificação)
func CheckRateLimit(ctx context.Context, ispb string, operation string) error {
    // Janela de 1 minuto
    key := fmt.Sprintf("ratelimit:%s:%s:%s", ispb, operation, time.Now().Format("2006-01-02-15:04"))

    // Incrementar contador
    count, err := redis.Incr(ctx, key)
    if err != nil {
        return err
    }

    // Definir TTL na primeira vez
    if count == 1 {
        redis.Expire(ctx, key, 1*time.Minute)
    }

    // Verificar limite
    limit := rateLimits[operation]
    if count > limit {
        return fmt.Errorf("rate limit exceeded: %d/%d", count, limit)
    }

    return nil
}
```

---

## ⚙️ Configuração do Redis Client

### Dependências (TEC-003 v2.1)
- **Biblioteca**: `github.com/redis/go-redis/v9` v9.14.1
- **Go Version**: 1.22+
- **Deployment**: Redis Standalone ou Cluster (produção)

### Connection Pool
```go
// Especificação de configuração (NÃO implementar agora)
type RedisConfig struct {
    Host            string        // "localhost:6379"
    Password        string        // Carregar de Vault/env
    DB              int           // 0 (default)
    PoolSize        int           // 100 (conexões simultâneas)
    MinIdleConns    int           // 10 (manter idle)
    MaxRetries      int           // 3
    DialTimeout     time.Duration // 5s
    ReadTimeout     time.Duration // 3s
    WriteTimeout    time.Duration // 3s
    PoolTimeout     time.Duration // 4s
}
```

### High Availability (Produção)
- **Modo**: Redis Sentinel (failover automático)
- **Replicas**: 3 nós (1 master, 2 replicas)
- **Persistence**: RDB snapshot (cada 5 minutos) + AOF (append-only file)

---

## 🔄 Serialização de Dados

### Formato: JSON
**Por quê?**
- ✅ Human-readable (debugging facilitado)
- ✅ Schema flexível (adicionar campos sem quebrar cache)
- ✅ Compatível com Go `encoding/json`

**Alternativas Consideradas**:
- ❌ Protocol Buffers: Mais rápido, mas dificulta debugging
- ❌ MessagePack: Mais compacto, mas menos suportado

### Compressão (Opcional)
Para valores grandes (> 10KB):
- **Algoritmo**: Gzip
- **Threshold**: 10KB
- **Trade-off**: CPU vs memória Redis

```go
// Pseudocódigo para serialização
func Serialize(v interface{}) ([]byte, error) {
    data, err := json.Marshal(v)
    if err != nil {
        return nil, err
    }

    // Comprimir se > 10KB
    if len(data) > 10*1024 {
        var buf bytes.Buffer
        gzWriter := gzip.NewWriter(&buf)
        gzWriter.Write(data)
        gzWriter.Close()
        return buf.Bytes(), nil
    }

    return data, nil
}
```

---

## 🗑️ Políticas de Invalidação

### 1. Time-based (TTL)
- **Padrão**: Todos os caches têm TTL
- **Evita**: Dados obsoletos ficarem indefinidamente

### 2. Event-based (Manual)
- **Trigger**: Eventos de mudança de estado
- **Exemplo**: Entry deletada → DELETE `dict:entry:*` relacionados

### 3. Cache Stampede Prevention
**Problema**: Múltiplos requests simultâneos ao expirar cache sobrecarregam DB/Bridge

**Solução**: Probabilistic Early Expiration
```go
// Pseudocódigo
func GetWithProbabilisticRefresh(ctx context.Context, key string, ttl time.Duration) (interface{}, error) {
    value, expiry, err := redis.GetWithExpiry(ctx, key)
    if err != nil {
        // Cache miss - buscar do DB
        return fetchFromDB(ctx, key)
    }

    // Calcular probabilidade de refresh antecipado
    delta := time.Until(expiry)
    beta := 1.0 // Tuning parameter
    prob := -beta * math.Log(rand.Float64())

    if delta < time.Duration(prob*float64(ttl)) {
        // Refresh antecipado
        go refreshCache(ctx, key)
    }

    return value, nil
}
```

---

## 📈 Monitoramento e Métricas

### Métricas Essenciais

#### 1. Cache Hit Rate
```
cache_hit_rate = cache_hits / (cache_hits + cache_misses)
```
**Target**: > 80% para `dict:entry:*`

#### 2. Latência
- **p50**: < 2ms
- **p95**: < 5ms
- **p99**: < 10ms

#### 3. Memory Usage
- **Monitorar**: `INFO memory` do Redis
- **Alert**: Uso > 80% da memória configurada

#### 4. Eviction Rate
- **Monitorar**: `evicted_keys` (INFO stats)
- **Alert**: > 100 evictions/min (sinal de memória insuficiente)

### Dashboards (Grafana/Prometheus)
```prometheus
# Cache hit rate
rate(redis_keyspace_hits_total[5m]) /
  (rate(redis_keyspace_hits_total[5m]) + rate(redis_keyspace_misses_total[5m]))

# Latência p95
histogram_quantile(0.95, redis_command_duration_seconds_bucket)

# Memory usage
redis_memory_used_bytes / redis_memory_max_bytes
```

---

## 🚨 Troubleshooting

### Problema 1: Cache Hit Rate Baixo (< 50%)
**Causas**:
- TTL muito curto
- Padrão de acesso não repetitivo
- Chaves mal projetadas

**Soluções**:
1. Aumentar TTL gradualmente (5min → 10min)
2. Analisar logs de access patterns
3. Revisar estrutura de chaves

---

### Problema 2: Memória Redis Cheia (Evictions)
**Causas**:
- Muitos dados cacheados
- TTLs muito longos
- Memory leak (chaves sem TTL)

**Soluções**:
1. Verificar chaves sem TTL: `redis-cli KEYS * --scan | xargs -L 1 redis-cli TTL`
2. Aumentar memória Redis (scale up)
3. Reduzir TTLs de caches menos críticos

---

### Problema 3: Latência Alta (p95 > 10ms)
**Causas**:
- Network latency (Redis em servidor remoto)
- CPU Redis saturada
- Slow commands (`KEYS *`, `SCAN` sem MATCH)

**Soluções**:
1. Usar `SCAN` com MATCH ao invés de `KEYS`
2. Monitorar slow log: `SLOWLOG GET 10`
3. Considerar Redis Cluster para distribuir carga

---

## 🔐 Segurança

### 1. Autenticação
- **Redis Password**: Obrigatório em produção
- **Carregar de**: Vault ou Kubernetes Secret
- **Rotação**: A cada 90 dias

### 2. Network Isolation
- **VPC privada**: Redis NÃO exposto à internet
- **Firewall**: Apenas pods do Connect podem acessar (Security Group)

### 3. Encryption
- **At rest**: Redis RDB/AOF criptografados (AES-256)
- **In transit**: TLS 1.2+ para conexões Redis (opcional, se Redis remoto)

---

## 📋 Checklist de Implementação

Para desenvolvedores que forem implementar esta especificação:

- [ ] Instalar Redis (local: Docker, prod: AWS ElastiCache/GCP MemoryStore)
- [ ] Configurar connection pool com parâmetros especificados
- [ ] Implementar helpers de serialização/deserialização JSON
- [ ] Criar funções para cada tipo de cache (entries, claims, workflows, idempotency)
- [ ] Implementar rate limiting com Redis INCR
- [ ] Configurar TTLs conforme especificado
- [ ] Adicionar logging de cache hits/misses
- [ ] Configurar Prometheus metrics
- [ ] Criar alertas para hit rate < 70%, evictions > 100/min
- [ ] Testar failover (Redis Sentinel)
- [ ] Documentar runbook de troubleshooting

---

## 📚 Referências

### Documentos Internos
- [TEC-003 v2.1: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md) - Stack tecnológica (Redis v9.14.1)
- [ANA-003: Análise Repo Connect](../../00_Analises/ANA-003_Analise_Repo_Connect.md) - Redis implementado mas sem docs
- [DAT-001: Schema Database Core DICT](DAT-001_Schema_Database_Core_DICT.md) - Estrutura de entries e claims
- [DAT-002: Schema Database Connect](DAT-002_Schema_Database_Connect.md) - Workflow metadata
- [GRPC-001: Bridge gRPC Service](../../04_APIs/gRPC/GRPC-001_Bridge_gRPC_Service.md) - Operações a cachear

### Documentação Externa
- [Redis Documentation](https://redis.io/docs/)
- [go-redis v9 Documentation](https://redis.uptrace.dev/)
- [Cache Stampede Prevention](https://en.wikipedia.org/wiki/Cache_stampede)
- [Redis Best Practices](https://redis.io/docs/manual/patterns/)

---

**Versão**: 1.0
**Status**: ✅ Especificação Completa (Aguardando implementação)
**Próxima Revisão**: Após implementação (validar métricas reais)

---

**IMPORTANTE**: Este é um documento de **especificação técnica**. A implementação será feita pelos desenvolvedores em fase posterior, baseando-se neste documento.
