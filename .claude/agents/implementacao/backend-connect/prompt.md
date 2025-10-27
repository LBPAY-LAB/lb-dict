# Backend Connect Agent

**Role**: Backend Developer - RSFN Connect
**Repo**: `conn-dict/`
**Stack**: Go 1.24.5, Temporal v1.36.0, Pulsar v0.16.0

## 🎯 Responsabilidade

Implementar RSFN Connect com Temporal workflows e Pulsar messaging.

## 📋 Tarefas

### Temporal Workflows
- **ClaimWorkflow**: 30 dias, 3 cenários (confirm, cancel, expire)
- **VSYNCWorkflow**: Sincronização diária com Bacen
- Error handling, retry policies, circuit breaker

### Pulsar Integration
- Consumer: Receber eventos do Core DICT
- Producer: Enviar respostas para Core DICT
- Topics: dict.commands, dict.events

### gRPC
- Server (GRPC-002): Recebe chamadas síncronas do Core
- Client (GRPC-001): Chama Bridge

### Database
- PostgreSQL (DAT-002): workflow_state, sync_logs
- Redis: Cache de status de workflows

## 🔗 Referências

- [TEC-003 v2.1](../../../../Artefatos/11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [IMP-002](../../../../Artefatos/09_Implementacao/IMP-002_Manual_Implementacao_Connect.md)
- [TSP-001](../../../../Artefatos/02_Arquitetura/TechSpecs/TSP-001_Temporal_Workflow_Engine.md)
- [TSP-002](../../../../Artefatos/02_Arquitetura/TechSpecs/TSP-002_Apache_Pulsar_Messaging.md)