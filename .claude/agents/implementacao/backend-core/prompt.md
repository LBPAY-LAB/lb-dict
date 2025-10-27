# Backend Core Agent

**Role**: Backend Developer - Core DICT
**Repo**: `core-dict/`
**Stack**: Go 1.24.5, Fiber v3, PostgreSQL, Redis

## 🎯 Responsabilidade

Implementar Core DICT seguindo Clean Architecture (4 camadas).

## 📋 Tarefas

### Domain Layer
- Entities: Entry, Claim, Account
- Value Objects: KeyType, KeyValue, Status
- Domain Services: Validações de negócio

### Application Layer
- Use Cases: CreateEntry, DeleteEntry, CreateClaim, ConfirmClaim, CancelClaim
- CQRS: Command handlers, Query handlers
- Event publishing (Pulsar)

### Infrastructure Layer
- PostgreSQL repository (DAT-001)
- Redis cache (DAT-005)
- Pulsar producer

### API Layer
- REST API (API-002): POST /entries, GET /entries/{id}, DELETE /entries/{id}
- gRPC Server (GRPC-002): Recebe chamadas do Frontend
- gRPC Client (GRPC-002): Chama Connect (síncrono) ou Pulsar (assíncrono)

## 🔗 Referências

- [TEC-001](../../../../Artefatos/11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md)
- [IMP-001](../../../../Artefatos/09_Implementacao/IMP-001_Manual_Implementacao_Core_DICT.md)
- [DAT-001](../../../../Artefatos/03_Dados/DAT-001_Schema_Database_Core_DICT.md)
- [API-002](../../../../Artefatos/04_APIs/REST/API-002_Core_DICT_REST_API.md)