# Arquitetura - RSFN Connect

Este documento descreve a arquitetura detalhada do RSFN Connect.

## Visao Geral

O RSFN Connect e composto por dois componentes principais:

1. **Connect API/Consumer**: Consome mensagens do Pulsar e expoe endpoints de saude
2. **Temporal Worker**: Executa workflows de longa duracao (Claims, VSYNC, OTP)

## Componentes

### 1. Connect API/Consumer

Responsabilidades:
- Consumir mensagens de `persistent://lb-conn/dict/rsfn-dict-req-out`
- Rotear mensagens para workflows apropriados
- Expor endpoints de health check (`:8081/health`)
- Expor metricas Prometheus (`:8081/metrics`)

### 2. Temporal Worker

Responsabilidades:
- Executar workflows registrados
- Executar activities (chamadas gRPC para Bridge, cache Redis, etc.)
- Gerenciar estado de workflows de longa duracao

## Workflows

### ClaimWorkflow (30 dias)

Gerencia o processo de reivindicacao de chave PIX:

```
[Inicio] -> [Criar Claim no Bacen via Bridge]
          -> [Aguardar 30 dias OU Signal de decisao]
          -> [Confirmar/Cancelar Claim]
          -> [Notificar Usuarios]
          -> [Fim]
```

### VSYNCWorkflow (Planejado)

Sincronizacao diaria de contas:

```
[Cron 00:00 BRT] -> [Buscar Contas do Core]
                 -> [Buscar Entradas do DICT]
                 -> [Comparar e Gerar Diff]
                 -> [Aplicar Sincronizacao]
                 -> [Fim]
```

### OTPWorkflow (Planejado)

Validacao de OTP para operacoes sensiveis.

## Integracao com Servicos Externos

### Bridge (gRPC)

- Endpoint: `bridge-grpc-svc:50051`
- Metodos:
  - `CreateEntry`
  - `CreateClaim`
  - `ConfirmClaim`
  - `CancelClaim`

### Pulsar (Mensageria)

- Consumer Topic: `persistent://lb-conn/dict/rsfn-dict-req-out`
- Producer Topic: `persistent://lb-conn/dict/rsfn-dict-res-out`
- Subscription: `connect-consumer-sub` (Shared)

### Redis (Cache)

- URL: `redis://redis:6379`
- Uso: Cache de consultas frequentes, dados temporarios

### PostgreSQL (Opcional)

- Database: `rsfn_connect`
- Uso: Persistencia de estado local (CID)

## Observabilidade

### Logs

- Formato: JSON estruturado
- Niveis: DEBUG, INFO, WARN, ERROR
- Contexto: traceID, workflowID, activityID

### Metricas

- `dict_claim_workflows_started_total`
- `dict_claim_workflows_completed_total{status}`
- `dict_entries_created_total`
- `dict_bridge_requests_total`

### Traces

- OpenTelemetry traces para todas as operacoes
- Propagacao de trace context para Bridge e dict.api

## Deployment

### Kubernetes

Componentes deployados:

1. `rsfn-connect-api` (Deployment, 3 replicas)
2. `rsfn-connect-worker` (Deployment, 2 replicas)
3. `temporal-server` (StatefulSet, 1 replica)
4. `pulsar` (StatefulSet, 1 replica)
5. `redis` (Deployment, 1 replica)

### Health Checks

- Liveness Probe: `GET /health`
- Readiness Probe: `GET /ready`

---

**Referencias**:
- [TEC-003: RSFN Connect Specification](../../Artefatos/11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [Temporal Documentation](https://docs.temporal.io/)
