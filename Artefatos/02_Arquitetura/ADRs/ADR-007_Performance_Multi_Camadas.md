# ADR-003 - Estratégia de Performance Multi-Camadas

**Status**: ✅ Aprovado
**Data**: 2025-10-24
**Decisores**: José Luís Silva (CTO), NEXUS (Solution Architect)
**Contexto**: Fase de Especificação do Projeto DICT LBPay

---

## Contexto

### Requisito Crítico de Performance

**Declaração do CTO** (conforme DUVIDAS.md - DUV-012):

> "**NEW CRITICAL REQUIREMENT**: High volume of DICT queries expected **(dozens per second)**"
>
> "Emphasized **performance is critical**"

### Requisitos Não-Funcionais

| ID | Requisito | Valor | Prioridade |
|----|-----------|-------|------------|
| **RNF-001** | Latência P99 para consultas (GET /entries/{Key}) | < 1s | 🔴 Crítico |
| **RNF-002** | Throughput mínimo | **Dezenas de queries/segundo** | 🔴 Crítico |
| **RNF-003** | Cache hit ratio | > 70% | 🟠 Alto |
| **RNF-004** | Connection pool reutilização | > 90% | 🟠 Alto |

### Situação Atual (AS-IS)

**Sem estratégia de cache estruturada**:
```
┌──────────┐
│ Cliente  │
└────┬─────┘
     │ GET /entries/{Key}
     ↓
┌─────────────┐
│money-moving │
└────┬────────┘
     │ SEMPRE faz chamada ao DICT Bacen (sem cache!)
     ↓
┌──────────────┐     mTLS handshake: 100-300ms
│ DICT Bacen   │     Network RTT: 50-100ms
│              │     Processing: 50-150ms
└──────────────┘
     ↓
LATÊNCIA TOTAL: 200-550ms (P99) ❌ RUIM!
```

**Problemas identificados**:

1. ❌ **Sem cache**: Toda consulta vai ao DICT Bacen
2. ❌ **mTLS handshake repetido**: Conexões não reutilizadas (100-300ms por request)
3. ❌ **Alto custo de rede**: Latência RTT Brasil-Bacen (50-100ms)
4. ❌ **Rate limiting vulnerável**: Facilmente atinge 429 (balde esgota rápido)
5. ❌ **Throughput limitado**: Max 2-25k req/min (conforme categoria PSP)

**Exemplo real** (sem cache):
```
100 consultas/segundo = 6.000 consultas/minuto

Categoria E (LBPay?): 2.500/min de limite
❌ ESTOURA rate limiting em 25 segundos!

Resultado: HTTP 429 (Rate Limited) → indisponibilidade para usuários
```

---

## Decisão

Implementar **estratégia de performance multi-camadas** com **5 caches Redis especializados** + **connection pooling agressivo**.

### Arquitetura de Performance

```
┌──────────┐
│ Cliente  │
└────┬─────┘
     │ GET /entries/{Key}?taxIdNumber=123
     ↓
┌─────────────────────────────────────────────────────────────┐
│                    Core DICT Service                         │
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │         CAMADA 1: Response Cache (Redis 7001)      │    │
│  │  Key: cache-dict-response:{keyType}:{key}:{taxId} │    │
│  │  TTL: 5 minutos                                    │    │
│  │  Hit Rate esperado: 70-80%                         │    │
│  └────────────────────────┬───────────────────────────┘    │
│                           │ MISS                             │
│                           ↓                                  │
│  ┌────────────────────────────────────────────────────┐    │
│  │    CAMADA 2: Key Validation Cache (Redis 7003)     │    │
│  │  Key: cache-dict-key-validation:{key}             │    │
│  │  TTL: 10 minutos                                   │    │
│  │  Hit Rate esperado: 50-60%                         │    │
│  └────────────────────────┬───────────────────────────┘    │
│                           │ MISS                             │
│                           ↓                                  │
│  ┌────────────────────────────────────────────────────┐    │
│  │         CAMADA 3: Deduplication (Redis 7004)       │    │
│  │  Previne requisições duplicadas (RequestId)       │    │
│  │  TTL: 1 minuto                                     │    │
│  └────────────────────────┬───────────────────────────┘    │
│                           │ Not duplicate                    │
│                           ↓                                  │
│  ┌────────────────────────────────────────────────────┐    │
│  │      CAMADA 4: Rate Limit Local (Redis 7005)       │    │
│  │  Token Bucket local (previne 429 do Bacen)        │    │
│  │  Políticas: ENTRIES_READ_USER_ANTISCAN, etc       │    │
│  └────────────────────────┬───────────────────────────┘    │
│                           │ Tokens available                 │
│                           ↓                                  │
│  ┌────────────────────────────────────────────────────┐    │
│  │   CAMADA 5: Connection Pool (HTTP Keep-Alive)      │    │
│  │  Reutiliza conexões mTLS                           │    │
│  │  Evita handshake (100-300ms → 0ms)                 │    │
│  └────────────────────────┬───────────────────────────┘    │
└───────────────────────────┼──────────────────────────────────┘
                            │ Chamada real ao Bacen (apenas se MISS em todas as camadas)
                            ↓
                   ┌──────────────┐
                   │ DICT Bacen   │
                   │ (API v2.6.1) │
                   └──────────────┘

LATÊNCIA COM CACHE HIT (Camada 1): 5-20ms (P99) ✅ EXCELENTE!
LATÊNCIA COM CACHE MISS: 150-300ms (P99) ✅ BOM!
```

---

## Detalhamento das Camadas

### Camada 1: Response Cache (Redis Port 7001)

**Propósito**: Cache de respostas completas do DICT Bacen.

#### Configuração

```yaml
redis:
  host: cache-dict-response
  port: 7001
  db: 0
  maxRetries: 3
  poolSize: 100
  minIdleConns: 10
  maxConnAge: 30m
  poolTimeout: 4s
  idleTimeout: 5m
  idleCheckFrequency: 1m
```

#### Schema de Keys

```
Formato: cache-dict-response:{keyType}:{key}:{taxIdNumber}

Exemplos:
- cache-dict-response:PHONE:+5561988880000:11122233300
- cache-dict-response:CPF:12345678901:12345678901
- cache-dict-response:EMAIL:user@example.com:98765432100
```

**Motivo do esquema**: Diferentes usuários consultando mesma chave devem ter cache separado (requisito de segurança anti-scan).

#### Value (JSON)

```json
{
  "key": "+5561988880000",
  "keyType": "PHONE",
  "account": {
    "participant": "12345678",
    "branch": "0001",
    "accountNumber": "0007654321",
    "accountType": "CACC"
  },
  "owner": {
    "type": "NATURAL_PERSON",
    "taxIdNumber": "11122233300",
    "name": "João Silva"
  },
  "cachedAt": "2023-10-24T10:00:00Z"
}
```

#### TTL: 5 minutos

**Justificativa**:
- ✅ Balanceamento entre freshness e performance
- ✅ Dados de chaves PIX mudam pouco (atualizações são raras)
- ✅ 5min é suficiente para absorver múltiplas consultas ao mesmo vínculo

#### Invalidação

**Eventos que invalidam cache**:
1. ✅ **Atualização de vínculo** (updateEntry)
2. ✅ **Exclusão de vínculo** (deleteEntry)
3. ✅ **Claim completado** (chave mudou de PSP)
4. ✅ **Evento CID** (via Pulsar consumer)

```go
// Invalidação automática
func (uc *UpdatePixKeyUseCase) Execute(ctx context.Context, input UpdatePixKeyInput) error {
    // ... atualização no DICT Bacen e PostgreSQL

    // Invalidar cache
    cacheKey := fmt.Sprintf("cache-dict-response:%s:%s:*", input.KeyType, input.Key)
    uc.cache.DelPattern(ctx, cacheKey) // Remove todas as variações (wildcard no taxIdNumber)

    return nil
}
```

#### Hit Rate Esperado: 70-80%

**Cálculo**:
```
Cenário típico:
- 1000 chaves PIX consultadas
- 700-800 são consultas repetidas (mesmo key + taxIdNumber) dentro de 5min
- Hit rate = 70-80%

Benefício:
- Sem cache: 1000 chamadas ao DICT Bacen
- Com cache: 200-300 chamadas ao DICT Bacen
- Redução: 70-80% de chamadas
```

#### Implementação

```go
// pkg/infrastructure/cache/response_cache.go

type ResponseCache struct {
    client *redis.Client
}

func (c *ResponseCache) Get(ctx context.Context, keyType, key, taxIdNumber string) (*domain.PixKey, error) {
    cacheKey := fmt.Sprintf("cache-dict-response:%s:%s:%s", keyType, key, taxIdNumber)

    val, err := c.client.Get(ctx, cacheKey).Result()
    if err == redis.Nil {
        return nil, ErrCacheMiss
    }
    if err != nil {
        return nil, err
    }

    var pixKey domain.PixKey
    if err := json.Unmarshal([]byte(val), &pixKey); err != nil {
        return nil, err
    }

    return &pixKey, nil
}

func (c *ResponseCache) Set(ctx context.Context, pixKey *domain.PixKey) error {
    cacheKey := fmt.Sprintf("cache-dict-response:%s:%s:%s",
        pixKey.KeyType, pixKey.Key, pixKey.Owner.TaxIdNumber)

    data, err := json.Marshal(pixKey)
    if err != nil {
        return err
    }

    return c.client.Set(ctx, cacheKey, data, 5*time.Minute).Err()
}

func (c *ResponseCache) DelPattern(ctx context.Context, pattern string) error {
    iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
    for iter.Next(ctx) {
        if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
            return err
        }
    }
    return iter.Err()
}
```

---

### Camada 2: Key Validation Cache (Redis Port 7003)

**Propósito**: Cache de validações de chaves (formato, existência, CPF/CNPJ).

#### Schema de Keys

```
cache-dict-key-validation:{key}

Exemplos:
- cache-dict-key-validation:+5561988880000
- cache-dict-key-validation:12345678901
```

#### Value (JSON)

```json
{
  "key": "+5561988880000",
  "isValid": true,
  "keyType": "PHONE",
  "cpfCnpjValid": true,
  "validatedAt": "2023-10-24T10:00:00Z"
}
```

#### TTL: 10 minutos

**Justificativa**:
- ✅ Validações de formato são determin��sticas (não mudam)
- ✅ Validações CPF/CNPJ na Receita Federal mudam raramente
- ✅ 10min balanceia freshness e redução de chamadas à Receita Federal

#### Hit Rate Esperado: 50-60%

**Benefício**: Reduz chamadas à API da Receita Federal (que pode ter rate limiting próprio).

---

### Camada 3: Deduplication Cache (Redis Port 7004)

**Propósito**: Prevenir requisições duplicadas (idempotência).

#### Schema de Keys

```
cache-dict-dedup:{requestId}

Exemplo:
- cache-dict-dedup:a946d533-7f22-42a5-9a9b-e87cd55c0f4d
```

#### Value

```
SHA256(request body)
```

#### TTL: 1 minuto

**Justificativa**:
- ✅ RequestId deve ser único por participante
- ✅ 1min é suficiente para detectar duplicatas acidentais (ex: retry automático do cliente)
- ✅ TTL curto economiza memória

#### Uso

```go
func (uc *CreatePixKeyUseCase) Execute(ctx context.Context, input CreatePixKeyInput) error {
    // 1. Check deduplication
    bodyHash := sha256.Sum256([]byte(input.ToJSON()))
    cacheKey := fmt.Sprintf("cache-dict-dedup:%s", input.RequestId)

    cachedHash, err := uc.dedupCache.Get(ctx, cacheKey)
    if err == nil {
        if cachedHash == bodyHash {
            // Duplicate request with same RequestId and same body → idempotent
            return uc.repo.FindByRequestId(ctx, input.RequestId) // Return existing result
        } else {
            // Same RequestId but different body → ERROR
            return ErrRequestIdAlreadyUsed
        }
    }

    // 2. Process request...
    // 3. Save to dedup cache
    uc.dedupCache.Set(ctx, cacheKey, bodyHash, 1*time.Minute)

    return nil
}
```

---

### Camada 4: Rate Limit Local Cache (Redis Port 7005)

**Propósito**: Implementar **token bucket local** para prevenir HTTP 429 do DICT Bacen.

#### Por que Rate Limiting Local?

**Problema**:
```
Sem rate limiting local:
1. Cliente faz 100 req/seg ao Core DICT
2. Core DICT repassa todas ao DICT Bacen
3. DICT Bacen: balde esgota em segundos
4. HTTP 429 Rate Limited
5. ❌ Indisponibilidade para usuários
```

**Solução**:
```
Com rate limiting local:
1. Cliente faz 100 req/seg ao Core DICT
2. Core DICT verifica balde LOCAL (Redis 7005)
3. Se tokens disponíveis: repassa ao DICT Bacen
4. Se tokens esgotados: retorna 429 ANTES de chamar Bacen
5. ✅ Evita esgotar balde do Bacen
6. ✅ Fail fast (latência baixa mesmo em rate limit)
```

#### Schema de Keys

```
cache-dict-rate-limit:{policy}:{scope}

Exemplos:
- cache-dict-rate-limit:ENTRIES_READ_USER_ANTISCAN:PF:11122233300
- cache-dict-rate-limit:ENTRIES_READ_PARTICIPANT_ANTISCAN:PSP:12345678
- cache-dict-rate-limit:ENTRIES_WRITE:PSP:12345678
```

#### Value

```
Número inteiro: tokens disponíveis
```

#### Algoritmo: Token Bucket

```go
type TokenBucket struct {
    Capacity     int       // Tamanho do balde
    Tokens       int       // Tokens disponíveis
    RefillRate   int       // Taxa de reposição (tokens/min)
    LastRefill   time.Time // Último refill
}

func (tb *TokenBucket) TryAcquire(count int) bool {
    // 1. Calcular tokens a repor desde último refill
    now := time.Now()
    elapsed := now.Sub(tb.LastRefill).Minutes()
    tokensToAdd := int(elapsed * float64(tb.RefillRate))

    // 2. Repor tokens (max = Capacity)
    tb.Tokens = min(tb.Tokens + tokensToAdd, tb.Capacity)
    tb.LastRefill = now

    // 3. Tentar consumir
    if tb.Tokens >= count {
        tb.Tokens -= count
        return true // Sucesso
    }

    return false // Rate limited
}
```

#### Sincronização com Bacen

**Problema**: Rate limiting local pode desincronizar com balde real do Bacen.

**Solução**: Monitorar endpoint `/policies/{policy}` periodicamente.

```go
// Background worker (a cada 1min)
func (rl *RateLimiter) SyncWithBacen(ctx context.Context) {
    // 1. Consultar estado do balde no Bacen
    state, err := rl.dictClient.GetBucketState(ctx, "ENTRIES_WRITE")
    if err != nil {
        return
    }

    // 2. Atualizar balde local
    cacheKey := "cache-dict-rate-limit:ENTRIES_WRITE:PSP:12345678"
    rl.cache.Set(ctx, cacheKey, state.Tokens, 1*time.Minute)
}
```

#### Reposição Automática por Liquidação SPI

**Regra especial** (conforme OpenAPI):
- ✅ **Consulta seguida de pagamento** repõe fichas automaticamente
- ✅ Status 200 + ordem PIX enviada: +1 ficha (PF) ou +2 fichas (PJ)

```go
// Consumidor Pulsar: topic spi-liquidation
func (c *SPIConsumer) HandleLiquidationEvent(ctx context.Context, event SPILiquidationEvent) {
    // 1. Verificar se houve consulta DICT antes do PIX
    if event.DICTQueryMade {
        // 2. Repor fichas no balde local
        cacheKey := fmt.Sprintf("cache-dict-rate-limit:ENTRIES_READ_USER_ANTISCAN:%s:%s",
            event.UserType, event.TaxIdNumber)

        tokensToAdd := 1 // PF
        if event.UserType == "PJ" {
            tokensToAdd = 2
        }

        rl.cache.IncrBy(ctx, cacheKey, tokensToAdd)
    }
}
```

**Benefício**: Usuários que consultam e pagam não esgotam balde rapidamente.

---

### Camada 5: Connection Pooling (HTTP Keep-Alive)

**Propósito**: Reutilizar conexões mTLS para evitar handshake repetido (100-300ms).

#### Problema: mTLS Handshake Caro

```
Sem connection pooling:
┌──────────┐                           ┌──────────────┐
│Core DICT │                           │ DICT Bacen   │
└────┬─────┘                           └──────┬───────┘
     │ Request 1                              │
     ├──────────────────────────────────────>│
     │ TLS Handshake (6 RTTs)                │
     │ Client Cert validation                │
     │ Server Cert validation                │
     │ ⏱️  100-300ms                          │
     │                                        │
     │ HTTP GET /entries/{Key}               │
     │<───────────────────────────────────────│
     │ 200 OK (50ms)                          │
     │                                        │
     │ 🔴 FECHA CONEXÃO                       │
     │                                        │
     │ Request 2 (alguns segundos depois)    │
     ├──────────────────────────────────────>│
     │ TLS Handshake NOVAMENTE (6 RTTs)      │
     │ ⏱️  100-300ms (DESPERDIÇADO!)         │
     ...

Total por request: 150-350ms (handshake + processing)
```

```
Com connection pooling (keep-alive):
┌──────────┐                           ┌──────────────┐
│Core DICT │                           │ DICT Bacen   │
└────┬─────┘                           └──────┬───────┘
     │ Request 1                              │
     ├──────────────────────────────────────>│
     │ TLS Handshake (6 RTTs)                │
     │ ⏱️  100-300ms (apenas 1ª vez!)        │
     │                                        │
     │ HTTP GET /entries/{Key}               │
     │ Connection: keep-alive                │
     │<───────────────────────────────────────│
     │ 200 OK (50ms)                          │
     │ Keep-Alive: timeout=60                │
     │                                        │
     │ ✅ MANTÉM CONEXÃO ABERTA              │
     │                                        │
     │ Request 2 (alguns segundos depois)    │
     ├──────────────────────────────────────>│
     │ ⏱️  0ms handshake (reutiliza conexão!) │
     │ HTTP GET /entries/{Key}               │
     │<───────────────────────────────────────│
     │ 200 OK (50ms)                          │
     ...

Total por request (após 1ª): 50-100ms (apenas processing)
Redução: 60-85% de latência!
```

#### Configuração (Go)

```go
// pkg/infrastructure/clients/connect_dict_client.go

func NewConnectDICTClient(certFile, keyFile, caFile string) (*ConnectDICTClient, error) {
    // 1. Load mTLS certificates
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return nil, err
    }

    caCert, err := ioutil.ReadFile(caFile)
    if err != nil {
        return nil, err
    }
    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)

    // 2. Configure TLS
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
        MinVersion:   tls.VersionTLS12,
    }

    // 3. Configure HTTP Transport with connection pooling
    transport := &http.Transport{
        TLSClientConfig: tlsConfig,

        // Connection pooling settings
        MaxIdleConns:        100,  // Max connections total (todos os hosts)
        MaxIdleConnsPerHost: 10,   // Max connections por host (DICT Bacen)
        MaxConnsPerHost:     20,   // Max connections ativas por host

        IdleConnTimeout:     90 * time.Second, // Timeout para conexões idle
        DisableKeepAlives:   false,            // ✅ HABILITAR keep-alive
        DisableCompression:  false,            // ✅ HABILITAR compressão

        // Timeouts
        DialContext: (&net.Dialer{
            Timeout:   5 * time.Second,  // Timeout para dial
            KeepAlive: 30 * time.Second, // TCP keep-alive
        }).DialContext,
        TLSHandshakeTimeout:   10 * time.Second,
        ResponseHeaderTimeout: 5 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
    }

    // 4. Create HTTP client
    client := &http.Client{
        Transport: transport,
        Timeout:   10 * time.Second, // Timeout total do request
    }

    return &ConnectDICTClient{
        client:  client,
        baseURL: "https://dict.pi.rsfn.net.br:16422/api/v2",
    }, nil
}

func (c *ConnectDICTClient) GetEntry(ctx context.Context, key string) (*Entry, error) {
    url := fmt.Sprintf("%s/entries/%s", c.baseURL, key)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    // IMPORTANTE: Header para solicitar compressão
    req.Header.Set("Accept-Encoding", "gzip")

    resp, err := c.client.Do(req) // Reutiliza conexão do pool!
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // ... parse response
}
```

#### Monitoramento de Connection Pool

**Métricas Prometheus**:
```prometheus
# Conexões ativas
dict_http_connections_active{host="dict.pi.rsfn.net.br"}

# Conexões idle (disponíveis para reuso)
dict_http_connections_idle{host="dict.pi.rsfn.net.br"}

# Taxa de reuso (target: > 90%)
dict_http_connection_reuse_ratio
```

**Alerta**:
```yaml
- alert: LowConnectionReuseRatio
  expr: dict_http_connection_reuse_ratio < 0.9
  for: 10m
  severity: warning
  annotations:
    summary: "Connection reuse ratio < 90% for 10 minutes"
    description: "Possible issue with keep-alive or high connection churn"
```

---

## Compressão (Adicional)

**Header de Request**:
```
Accept-Encoding: gzip
```

**Benefício**:
- ✅ Reduz largura de banda em 60-80%
- ✅ Reduz latência de rede (menos bytes transmitidos)

**Exemplo**:
```
Resposta sem compressão: 5 KB
Resposta com gzip: 1-2 KB
Redução: 60-80%

Latência de rede (100 Mbps):
- Sem compressão: 5 KB * 8 / 100 Mbps = 0.4ms
- Com compressão: 1.5 KB * 8 / 100 Mbps = 0.12ms
Ganho: ~0.3ms (marginal, mas todo ganho conta!)
```

---

## Resultados Esperados

### Latências (P99)

| Cenário | Sem Otimizações | Com Otimizações | Melhoria |
|---------|-----------------|-----------------|----------|
| **Cache HIT (Camada 1)** | N/A | **5-20ms** | ✅ Excelente |
| **Cache MISS (primeira consulta)** | 200-550ms | **150-300ms** | 45% mais rápido |
| **Consulta recorrente (dentro de 5min)** | 200-550ms | **5-20ms** | **90-97% mais rápido** |

### Throughput

| Métrica | Sem Otimizações | Com Otimizações | Melhoria |
|---------|-----------------|-----------------|----------|
| **Req/seg suportados** | 10-40 (limitado por Bacen) | **100-500** | **10-50x** |
| **Cache hit ratio** | 0% | **70-80%** | N/A |
| **Chamadas ao Bacen reduzidas** | 100% | **20-30%** | **70-80% redução** |

### Exemplo Prático

**Cenário**: 100 consultas/segundo ao mesmo conjunto de 1000 chaves PIX.

```
SEM OTIMIZAÇÕES:
- 100 req/seg * 60 seg = 6.000 req/min
- Categoria E (LBPay?): limite 2.500/min
- ❌ RATE LIMITED em 25 segundos
- Indisponibilidade: 75% do tempo

COM OTIMIZAÇÕES (70% cache hit rate):
- 100 req/seg * 60 seg = 6.000 req/seg
- Cache HIT: 4.200 req/seg (não vão ao Bacen)
- Cache MISS: 1.800 req/seg (vão ao Bacen) → 30/seg → 1.800/min
- 1.800/min < 2.500/min (limite Categoria E)
- ✅ DENTRO DO LIMITE
- Disponibilidade: 100%
```

---

## Consequências

### ✅ Positivas

1. **Performance Crítica Atendida**:
   - ✅ Latência P99 < 1s (meta: 5-20ms com cache)
   - ✅ Throughput de dezenas/centenas de req/seg

2. **Redução de Custos**:
   - ✅ 70-80% menos chamadas ao DICT Bacen
   - ✅ Menor uso de rate limiting (evita 429)

3. **Melhor UX**:
   - ✅ Respostas instantâneas (5-20ms)
   - ✅ Menor latência percebida

4. **Resiliência**:
   - ✅ Se DICT Bacen estiver lento/indisponível, cache absorve carga
   - ✅ Degradação graciosa

5. **Observabilidade**:
   - ✅ Métricas de cache hit rate
   - ✅ Monitoramento de connection pool

### ⚠️ Negativas (e Mitigações)

#### 1. Complexidade de Infraestrutura

**Problema**: 5 instâncias Redis + lógica de invalidação.

**Mitigação**:
- ✅ **Redis Cluster**: Gerenciamento simplificado
- ✅ **Infrastructure as Code** (Terraform): Automação de provisionamento
- ✅ **Helm Charts**: Deploy padronizado no Kubernetes

#### 2. Dados Potencialmente Stale

**Problema**: Cache pode retornar dados desatualizados (até 5min).

**Mitigação**:
- ✅ **Invalidação ativa**: Eventos CID (Pulsar) invalidam cache imediatamente
- ✅ **TTL curto (5min)**: Balanceamento entre freshness e performance
- ✅ **Cache bypass**: Header `X-Cache-Control: no-cache` para forçar consulta ao Bacen

**Exemplo**:
```go
// Cliente pode forçar bypass de cache
req.Header.Set("X-Cache-Control", "no-cache")
```

#### 3. Uso de Memória

**Problema**: 5 Redis consumindo memória.

**Cálculo de Uso**:
```
Estimativa:
- 1 milhão de chaves PIX
- Cada entrada de cache: ~1 KB (JSON)
- Total: 1 GB de RAM (Response Cache)
- Outros caches: 200-300 MB cada
- Total: ~2.5 GB de RAM

Custo (AWS ElastiCache):
- cache.r6g.large (13.07 GB RAM): $0.201/hora
- 5 instâncias: $0.201 * 5 = $1.005/hora
- Mensal: ~$730/mês
```

**Mitigação**:
- ✅ **Eviction policy**: `allkeys-lru` (remove menos usados)
- ✅ **TTL agressivo**: Libera memória automaticamente
- ✅ **Monitoramento**: Alerta se uso > 80%

#### 4. Sincronização de Rate Limiting

**Problema**: Rate limiting local pode desincronizar com Bacen.

**Mitigação**:
- ✅ **Polling periódico**: `/policies/{policy}` a cada 1min
- ✅ **Ajuste conservador**: Usar 90% do limite do Bacen (margem de segurança)
- ✅ **Backoff em 429**: Se receber 429 do Bacen, ajustar balde local imediatamente

---

## Alternativas Consideradas

### Alternativa 1: Cache Único (sem multi-camadas)

**Prós**:
- ✅ Mais simples

**Contras**:
- ❌ Hit rate menor (~40-50%)
- ❌ Não previne rate limiting local
- ❌ Não previne duplicatas

**Decisão**: ❌ Rejeitada - Performance insuficiente.

---

### Alternativa 2: Cache em Memória (in-process)

**Prós**:
- ✅ Latência ultra-baixa (< 1ms)
- ✅ Sem dependência externa

**Contras**:
- ❌ Não compartilhado entre pods (cada pod tem cache próprio)
- ❌ Uso alto de RAM por pod
- ❌ Invalidação complexa (precisa broadcast)

**Decisão**: ❌ Rejeitada - Não escalável horizontalmente.

---

### Alternativa 3: Redis Único (1 instância para tudo)

**Prós**:
- ✅ Mais simples (menos instâncias)

**Contras**:
- ❌ Single point of failure
- ❌ Contenção (todos os acessos na mesma instância)
- ❌ Difícil tuning (TTLs e políticas diferentes)

**Decisão**: ❌ Rejeitada - 5 instâncias especializadas são mais performáticas.

---

## Decisão Final

✅ **APROVADA**: Implementar **estratégia de performance multi-camadas** com **5 caches Redis especializados** + **connection pooling agressivo**.

### Justificativa

1. ✅ **Única forma de atingir RNF-002** (dezenas de queries/segundo)
2. ✅ **Reduz 70-80% de chamadas** ao DICT Bacen
3. ✅ **Latência P99 < 20ms** (com cache hit)
4. ✅ **Previne HTTP 429** (rate limiting local)
5. ✅ **Observabilidade completa** (métricas de cada camada)
6. ✅ **Escalável horizontalmente** (Redis Cluster + Kubernetes HPA)

---

## Implementação

### Fase 1: Infraestrutura (Semana 1)

```bash
# 1. Provisionar 5 Redis (Terraform)
terraform apply -target=module.redis_cluster

# 2. Configurar Redis
kubectl apply -f k8s/redis/

# 3. Testar conectividade
make test-redis-connection
```

### Fase 2: Response Cache (Semana 2)

```bash
# Implementar Camada 1 (Response Cache)
# - pkg/infrastructure/cache/response_cache.go
# - Integrar em GetPixKeyUseCase
# - Testes unitários
```

### Fase 3: Camadas Restantes (Semanas 3-4)

```bash
# Implementar Camadas 2, 3, 4, 5
# - Key Validation Cache
# - Deduplication Cache
# - Rate Limit Local Cache
# - Connection Pooling
```

### Fase 4: Observabilidade (Semana 5)

```bash
# Métricas, dashboards, alertas
# - Prometheus metrics
# - Grafana dashboards
# - Alertmanager rules
```

### Fase 5: Testes de Performance (Semana 6)

```bash
# Load testing
k6 run tests/load/get_pixkey_load_test.js

# Validar:
# - Latência P99 < 20ms (cache hit)
# - Throughput > 100 req/seg
# - Cache hit rate > 70%
```

---

## Monitoramento

### Métricas Chave

```prometheus
# Cache performance
dict_cache_hit_ratio{cache="response"} > 0.7
dict_cache_latency_seconds{cache="response", quantile="0.99"} < 0.020

# Connection pool
dict_http_connection_reuse_ratio > 0.9
dict_http_connections_active < 50

# Rate limiting
dict_rate_limited_requests_total < 10/hour
```

### Dashboards Grafana

1. **Cache Performance Dashboard**
   - Hit rate por cache (5 gráficos)
   - Latência P50/P95/P99
   - Memory usage

2. **Connection Pool Dashboard**
   - Conexões ativas vs idle
   - Taxa de reuso
   - Handshake time

3. **Rate Limiting Dashboard**
   - Tokens disponíveis por política
   - Requisições rate limited
   - Sincronização com Bacen

---

## Referências

1. **Token Bucket Algorithm**
   https://en.wikipedia.org/wiki/Token_bucket

2. **HTTP Keep-Alive**
   https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Keep-Alive

3. **Redis Best Practices**
   https://redis.io/docs/management/optimization/

4. **Go HTTP Transport Tuning**
   https://golang.org/pkg/net/http/#Transport

5. **API-001** - Especificação de APIs DICT Bacen (Rate Limiting)
   [Artefatos/04_APIs/API-001_Especificacao_APIs_DICT_Bacen.md](../04_APIs/API-001_Especificacao_APIs_DICT_Bacen.md#52-todas-as-políticas-de-rate-limiting)

6. **DAS-001** - Arquitetura de Solução TO-BE (Performance)
   [Artefatos/02_Arquitetura/DAS-001_Arquitetura_Solucao_TO_BE.md](DAS-001_Arquitetura_Solucao_TO_BE.md#9-estratégia-de-performance)

---

**Documento criado por**: NEXUS (AGT-ARC-001) - Solution Architect
**Aprovado por**: José Luís Silva (CTO)
**Data de Aprovação**: 2025-10-24
**Status**: ✅ Aprovado
**Impacto**: 🔴 Crítico (performance é requisito essencial)
