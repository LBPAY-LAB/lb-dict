# Tech Specs (Especificações Técnicas Detalhadas)

**Propósito**: Especificações técnicas detalhadas de componentes e subsistemas

## 📋 Conteúdo

Esta pasta armazenará:

- **Component Specs**: Especificações técnicas de componentes individuais
- **API Contracts**: Contratos de API detalhados
- **Interface Specifications**: Especificações de interfaces entre componentes
- **Technology Deep Dives**: Análises técnicas profundas de tecnologias usadas

## 📁 Estrutura Esperada

```
TechSpecs/
├── Components/
│   ├── COMP-001_Connect_API_Server.md
│   ├── COMP-002_Connect_Orchestration_Worker.md
│   ├── COMP-003_Bridge_SOAP_Adapter.md
│   ├── COMP-004_Bridge_gRPC_Server.md
│   └── COMP-005_Core_DICT_API.md
├── Interfaces/
│   ├── INT-001_Connect_Bridge_gRPC.md
│   ├── INT-002_Core_Connect_gRPC.md
│   └── INT-003_Bridge_Bacen_SOAP.md
├── Technology_Choices/
│   ├── TECH-001_Why_Temporal_For_Workflows.md
│   ├── TECH-002_Why_Pulsar_Over_Kafka.md
│   ├── TECH-003_Why_gRPC_For_Internal_APIs.md
│   └── TECH-004_Why_Goose_For_Migrations.md
└── Performance/
    ├── PERF-001_Latency_Requirements.md
    ├── PERF-002_Throughput_Targets.md
    └── PERF-003_Scalability_Strategy.md
```

## 🎯 Template: Component Spec

```markdown
# COMP-001: Connect API Server

## Visão Geral

**Propósito**: Servidor gRPC que expõe APIs para Core DICT e recebe operações DICT

**Linguagem**: Go 1.22
**Framework**: gRPC (google.golang.org/grpc)
**Port**: 8080 (gRPC), 9090 (HTTP metrics)

## Responsabilidades

1. **API gRPC**:
   - Expor RPCs para Core DICT (CreateEntry, CreateClaim, etc.)
   - Validar requests (input validation)
   - Transformar modelos gRPC → Domain

2. **Orquestração**:
   - Iniciar workflows Temporal (ClaimWorkflow)
   - Chamar Bridge via gRPC
   - Coordenar transações distribuídas

3. **Persistência**:
   - Salvar entries, claims no PostgreSQL
   - Atualizar status de workflows
   - Manter histórico de auditoria

4. **Cache**:
   - Cache de entries no Redis (5min TTL)
   - Idempotency keys (24h TTL)
   - Rate limiting counters

5. **Eventos**:
   - Publicar eventos no Pulsar (EntryCreated, ClaimCreated)
   - Event sourcing para auditoria

## Arquitetura Interna (Clean Architecture)

### Camada API (Handlers gRPC)
```go
type ConnectServiceServer struct {
    createEntryUseCase *usecases.CreateEntryUseCase
    createClaimUseCase *usecases.CreateClaimUseCase
}

func (s *ConnectServiceServer) CreateEntry(ctx context.Context, req *pb.CreateEntryRequest) (*pb.CreateEntryResponse, error) {
    // 1. Validate request
    // 2. Transform proto → domain
    // 3. Call use case
    // 4. Transform domain → proto
    // 5. Return response
}
```

### Camada Domain (Entities, Use Cases)
```go
type CreateEntryUseCase struct {
    entryRepo      repository.EntryRepository
    bridgeClient   bridge.BridgeClient
    eventPublisher events.Publisher
}

func (uc *CreateEntryUseCase) Execute(ctx context.Context, cmd CreateEntryCommand) (*Entry, error) {
    // Business logic aqui
}
```

### Camada Infrastructure (Repositories, Clients)
```go
type PostgresEntryRepository struct {
    db *sql.DB
}

func (r *PostgresEntryRepository) Create(ctx context.Context, entry *domain.Entry) error {
    // SQL INSERT
}
```

## Dependências

| Dependência | Versão | Propósito |
|-------------|--------|-----------|
| `google.golang.org/grpc` | v1.64.0 | gRPC server/client |
| `github.com/lib/pq` | v1.10.9 | PostgreSQL driver |
| `github.com/redis/go-redis/v9` | v9.14.1 | Redis client |
| `github.com/apache/pulsar-client-go` | v0.16.0 | Pulsar producer |
| `go.temporal.io/sdk` | v1.36.0 | Temporal client |
| `go.uber.org/zap` | v1.27.0 | Structured logging |
| `github.com/prometheus/client_golang` | v1.19.0 | Metrics |

## Configuração

```yaml
# config.yaml
server:
  grpc_port: 8080
  http_port: 9090

database:
  host: postgres.dict.svc.cluster.local
  port: 5432
  database: dict
  user: dict_user
  max_connections: 100

redis:
  host: redis.dict.svc.cluster.local
  port: 6379
  pool_size: 100

bridge:
  grpc_address: bridge.dict.svc.cluster.local:8081
  timeout: 5s
  retry_attempts: 3

temporal:
  host: temporal.temporal.svc.cluster.local
  port: 7233
  namespace: dict

pulsar:
  brokers:
    - pulsar://pulsar.pulsar.svc.cluster.local:6650
  topic_prefix: dict
```

## Métricas (Prometheus)

```go
var (
    requestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "connect_grpc_requests_total",
            Help: "Total gRPC requests",
        },
        []string{"method", "status"},
    )

    requestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "connect_grpc_request_duration_seconds",
            Help:    "gRPC request duration",
            Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
        },
        []string{"method"},
    )
)
```

## Health Checks

### Liveness Probe
```bash
grpc_health_probe -addr=:8080
```

### Readiness Probe
```bash
# Verifica:
# - PostgreSQL conectado
# - Redis conectado
# - Bridge acessível
# - Temporal conectado
```

## Testes

- **Unit Tests**: 85% coverage
- **Integration Tests**: PostgreSQL + Redis (Testcontainers)
- **E2E Tests**: Connect + Bridge + Mock Bacen

## Performance Targets

- **Latency p95**: < 200ms (CreateEntry)
- **Latency p99**: < 500ms
- **Throughput**: 1000 req/s
- **Concurrent connections**: 10,000

## Referências

- [TEC-003 v2.1: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [GRPC-001: Bridge gRPC Service](../../04_APIs/gRPC/GRPC-001_Bridge_gRPC_Service.md)
- [DAT-002: Schema Database Connect](../../03_Dados/DAT-002_Schema_Database_Connect.md)
```

## 📊 Tech Choice Documents

### TECH-001: Why Temporal for Workflows?

**Decisão**: Usar Temporal para ClaimWorkflow (30 dias)

**Razões**:
- ✅ Workflows duráveis (30 dias sem perder estado)
- ✅ Built-in retry e compensação (Saga pattern)
- ✅ Visibilidade (Temporal UI mostra workflows em execução)
- ✅ Testabilidade (mocking de activities)
- ✅ Escalabilidade (workflows distribuídos)

**Alternativas Consideradas**:
- ❌ Cron jobs: Não escala, difícil debugar workflows longos
- ❌ Celery: Não suporta workflows duráveis de 30 dias
- ❌ AWS Step Functions: Vendor lock-in, custo por execução

**Referências**: [ADR-005: Temporal Workflows](../ADRs/ADR-005_Temporal_Workflows.md)

---

### TECH-002: Why Pulsar over Kafka?

**Decisão**: Usar Apache Pulsar para event streaming

**Razões**:
- ✅ Multi-tenancy nativo (namespaces)
- ✅ Geo-replication built-in
- ✅ Separação storage/compute (BookKeeper)
- ✅ Schema registry built-in
- ✅ Performance superior (leitura paralela)

**Alternativas Consideradas**:
- ❌ Kafka: Complexo para multi-tenancy, requer Kafka Streams
- ❌ RabbitMQ: Não otimizado para streaming (mais para queues)

**Referências**: [TEC-003 v2.1: Apache Pulsar](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)

## 📚 Referências

- [Especificações Técnicas TEC-001, TEC-002, TEC-003](../../11_Especificacoes_Tecnicas/)
- [ADRs](../ADRs/)
- [Diagramas de Arquitetura](../Diagramas/)

---

**Status**: 🔴 Pasta vazia (será preenchida na Fase 2)
**Fase de Preenchimento**: Fase 2 (detalhamento técnico)
**Responsável**: Tech Lead + Arquitetos
