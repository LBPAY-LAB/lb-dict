# conn-dict - Contratos e Integração

**Versão**: 1.0
**Data**: 2025-10-27
**Status**: ✅ Pronto para integração com core-dict

---

## Contratos (dict-contracts v0.2.0)

Este serviço usa contratos formais definidos em `dict-contracts`.

### Importar Contratos

```go
import (
    commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
    connectv1 "github.com/lbpay-lab/dict-contracts/gen/proto/connect/v1"
    bridgev1 "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
)
```

### Setup (go.mod)

```bash
go mod edit -replace github.com/lbpay-lab/dict-contracts=../dict-contracts
go mod tidy
```

---

## APIs Disponíveis

### gRPC Server (Porta 9092)

**17 RPCs implementados**:
- **EntryService** (3 RPCs): GetEntry, GetEntryByKey, ListEntries
- **ClaimService** (5 RPCs): CreateClaim, ConfirmClaim, CancelClaim, GetClaim, ListClaims
- **InfractionService** (6 RPCs): CreateInfraction, InvestigateInfraction, ResolveInfraction, DismissInfraction, GetInfraction, ListInfractions
- **HealthCheck** (1 RPC): HealthCheck

### Pulsar Topics

**Input** (core-dict → conn-dict):
- `dict.entries.created` → EntryCreatedEvent
- `dict.entries.updated` → EntryUpdatedEvent
- `dict.entries.deleted.immediate` → EntryDeletedEvent

**Output** (conn-dict → core-dict):
- `dict.entries.status.changed` → EntryStatusChangedEvent
- `dict.claims.created` → ClaimCreatedEvent
- `dict.claims.completed` → ClaimCompletedEvent
- `dict.infractions.reported` → InfractionReportedEvent
- `dict.infractions.resolved` → InfractionResolvedEvent

---

## Executar

```bash
# Start dependencies
docker-compose up -d

# Run server
go run ./cmd/server

# Run worker (Temporal)
go run ./cmd/worker
```

**Portas**:
- gRPC Server: 9092
- Health Check: 8080
- Metrics: 9091

---

## Documentação

- [STATUS_FINAL_2025-10-27.md](../Artefatos/00_Master/STATUS_FINAL_2025-10-27.md) - Instruções para core-dict
- [CONN_DICT_API_REFERENCE.md](../Artefatos/00_Master/CONN_DICT_API_REFERENCE.md) - API reference completo
