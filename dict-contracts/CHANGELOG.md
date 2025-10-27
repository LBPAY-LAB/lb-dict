# Changelog

All notable changes to the DICT Contracts project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.3.0] - 2025-10-27

### Changed

#### proto/common.proto - Enums Validados com Implementações

- **EntryStatus**: Renumerados e expandidos com base em core-dict validado
  - `ENTRY_STATUS_PENDING` (1): Pendente (aguardando confirmação)
  - `ENTRY_STATUS_ACTIVE` (2): Ativa (pode receber PIX)
  - `ENTRY_STATUS_BLOCKED` (3): Bloqueada (não pode receber PIX)
  - `ENTRY_STATUS_PORTABILITY_PENDING` (4): Mantido
  - `ENTRY_STATUS_PORTABILITY_CONFIRMED` (5): Mantido
  - `ENTRY_STATUS_CLAIM_PENDING` (6): Mantido
  - `ENTRY_STATUS_DELETED` (7): Deletada (soft delete)

- **ClaimStatus**: Adicionado novo status validado
  - `CLAIM_STATUS_AUTO_CONFIRMED` (7): Auto-confirmada após timeout (30 dias)
  - Validado contra implementação conn-dict e core-dict

### Technical Details

**Breaking Changes**: Sim (renumeração de EntryStatus)
- Números dos status de Entry foram reordenados para acomodar novos status
- Requer atualização de código que usa comparação numérica direta
- Código que usa enums pelo nome não é afetado

**Compatibility**:
- conn-dict: ✅ Validado contra implementação existente
- conn-bridge: ✅ Compatible
- core-dict: ✅ Validado contra entities

**Migration Guide**:
1. Atualizar dict-contracts para v0.3.0
2. Executar `make proto-gen` em cada repo
3. Recompilar todos os repos (conn-dict, conn-bridge, core-dict)

---

## [v0.2.0] - 2025-10-27

### Added

#### New Proto Files - ConnectService (Core DICT ↔ Connect)

- **proto/conn_dict/v1/connect_service.proto**: ConnectService for Core DICT → Connect communication
  - 17 RPC methods for complete Connect integration
  - Entry operations (read-only): GetEntry, GetEntryByKey, ListEntries
  - Claim operations: CreateClaim, ConfirmClaim, CancelClaim, GetClaim, ListClaims
  - Infraction operations: CreateInfraction, InvestigateInfraction, ResolveInfraction, DismissInfraction, GetInfraction, ListInfractions
  - Health check: HealthCheck with component-level status

- **proto/conn_dict/v1/events.proto**: Pulsar event schemas for async communication
  - Input events (Core DICT → Connect): EntryCreatedEvent, EntryUpdatedEvent, EntryDeletedEvent
  - Output events (Connect → Core DICT): EntryStatusChangedEvent, ClaimCreatedEvent, ClaimCompletedEvent, InfractionReportedEvent, InfractionResolvedEvent

#### Code Generation
- Generated Go code: 5,837 lines (connect_service.pb.go, connect_service_grpc.pb.go, events.pb.go)

### Technical Specifications

**New gRPC Methods**: 17 (total now: 46)
**New Pulsar Event Types**: 8
**New Package**: dict.connect.v1

---

## [v0.1.0] - 2025-10-26

### Added
- common.proto, core_dict.proto, bridge.proto
- 29 gRPC methods (15 CoreDictService + 14 BridgeService)
- Go module setup with code generation

---

[v0.3.0]: https://github.com/lbpay/dict-contracts/releases/tag/v0.3.0
[v0.2.0]: https://github.com/lbpay/dict-contracts/releases/tag/v0.2.0
[v0.1.0]: https://github.com/lbpay/dict-contracts/releases/tag/v0.1.0
