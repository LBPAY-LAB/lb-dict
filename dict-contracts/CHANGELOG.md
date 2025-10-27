# Changelog

All notable changes to the DICT Contracts project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[v0.2.0]: https://github.com/lbpay/dict-contracts/releases/tag/v0.2.0
[v0.1.0]: https://github.com/lbpay/dict-contracts/releases/tag/v0.1.0
