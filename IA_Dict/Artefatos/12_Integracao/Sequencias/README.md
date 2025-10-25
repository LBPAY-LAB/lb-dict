# Diagramas de Sequência

**Propósito**: Diagramas de sequência detalhados para operações críticas do sistema DICT

## 📋 Conteúdo

Esta pasta armazenará:

- **Sequence Diagrams**: Diagramas UML de sequência para operações DICT
- **Activity Diagrams**: Fluxos de atividades em workflows Temporal
- **Interaction Diagrams**: Interações complexas entre múltiplos componentes

## 📁 Estrutura Esperada

```
Sequencias/
├── Entry_Operations/
│   ├── SEQ_CreateEntry.md
│   ├── SEQ_GetEntry.md
│   ├── SEQ_DeleteEntry.md
│   └── SEQ_UpdateEntry.md
├── Claim_Operations/
│   ├── SEQ_CreateClaim.md
│   ├── SEQ_ClaimWorkflow_30_Days.md
│   ├── SEQ_CompleteClaim.md
│   └── SEQ_CancelClaim.md
├── Portability/
│   ├── SEQ_ConfirmPortability.md
│   └── SEQ_CancelPortability.md
└── Error_Handling/
    ├── SEQ_Retry_Logic.md
    └── SEQ_Rollback_Transaction.md
```

## 🎯 Exemplo: SEQ_CreateClaim (ClaimWorkflow 30 dias)

```mermaid
sequenceDiagram
    participant User as User (FrontEnd)
    participant Core as Core DICT
    participant Connect as RSFN Connect
    participant Temporal as Temporal Server
    participant Bridge as Bridge DICT
    participant Bacen as Bacen DICT
    participant DB as PostgreSQL
    participant Pulsar as Apache Pulsar

    User->>Core: POST /claims {entry_id, claimer_account}
    Core->>Connect: CreateClaim (gRPC)

    Connect->>Temporal: StartWorkflow(ClaimWorkflow)
    Temporal->>Temporal: Create Workflow Instance (30 days)

    Connect->>Bridge: CreateClaim (gRPC)
    Bridge->>Bridge: Sign SOAP XML (ICP-Brasil A3)
    Bridge->>Bacen: CreateClaim (SOAP/mTLS)

    Bacen-->>Bridge: ClaimCreated (external_id)
    Bridge-->>Connect: ClaimCreated (external_id)

    Connect->>DB: INSERT INTO claims (status='OPEN', expires_at=now()+30days)
    Connect->>Pulsar: Publish ClaimCreated event

    Connect->>Temporal: RecordActivity(ClaimCreated)
    Temporal->>Temporal: Schedule Timer (30 days)

    Connect-->>Core: ClaimCreated {claim_id, expires_at}
    Core-->>User: 201 Created

    Note over Temporal: Wait 30 days or external event

    alt Owner accepts claim (before 30 days)
        User->>Core: POST /claims/{id}/complete
        Core->>Connect: CompleteClaim (gRPC)
        Connect->>Temporal: Signal(ClaimAccepted)
        Temporal->>Temporal: Complete Workflow
        Connect->>Bridge: CompleteClaim (gRPC)
        Bridge->>Bacen: CompleteClaim (SOAP)
        Connect->>DB: UPDATE claims SET status='COMPLETED'
        Connect->>Pulsar: Publish ClaimCompleted event
    else 30 days timeout
        Temporal->>Temporal: Timer Expired
        Temporal->>Connect: ExecuteActivity(ExpireClaim)
        Connect->>DB: UPDATE claims SET status='EXPIRED'
        Connect->>Pulsar: Publish ClaimExpired event
    end
```

## 🔗 Padrões de Diagramas de Sequência

### 1. Operações Síncronas (gRPC)
```
Client->>Server: Request
Server-->>Client: Response
```

### 2. Operações Assíncronas (Pulsar)
```
Producer->>Pulsar: Publish Event
Pulsar->>Consumer: Consume Event
```

### 3. Workflows de Longa Duração (Temporal)
```
Client->>Temporal: StartWorkflow
Temporal->>Temporal: Schedule Timer (30 days)
Note over Temporal: Workflow Running
Temporal->>Worker: ExecuteActivity
```

## 📚 Referências

- [Fluxos](../Fluxos/)
- [TEC-003: ClaimWorkflow 30 dias](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [GRPC-001: Bridge gRPC Service](../../04_APIs/gRPC/GRPC-001_Bridge_gRPC_Service.md)
- [ArquiteturaDict_LBPAY.md](../../02_Arquitetura/ArquiteturaDict_LBPAY.md)

---

**Status**: 🔴 Pasta vazia (será preenchida na Fase 2)
**Fase de Preenchimento**: Fase 2 (durante design detalhado)
**Ferramenta**: Mermaid, PlantUML, IcePanel
