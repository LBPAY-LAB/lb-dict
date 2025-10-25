# DIA-004: C4 Component Diagram - RSFN Connect

**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: Equipe Arquitetura
**Status**: ✅ Completo

---

## Sumário Executivo

Este documento apresenta o **C4 Component Diagram** (nível 3) do **RSFN Connect**, detalhando os componentes internos do Temporal Worker e Pulsar Consumer, e como se organizam segundo a **Clean Architecture**.

**Objetivo**: Mostrar a estrutura interna dos containers Connect API, Temporal Worker e Pulsar Consumer, separação de responsabilidades por camada, e como os componentes interagem para orquestrar workflows duráveis.

**Pré-requisitos**:
- [DIA-001: C4 Context Diagram](./DIA-001_C4_Context_Diagram.md)
- [DIA-002: C4 Container Diagram](./DIA-002_C4_Container_Diagram.md)
- [TEC-003 v2.1: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)

---

## 1. Clean Architecture no RSFN Connect

O RSFN Connect segue **Clean Architecture** com foco em workflows duráveis:

```
┌─────────────────────────────────────────────────────┐
│  Workflow Layer (Orchestration)                     │  ← Temporal Workflows, Activities
│  - ClaimWorkflow, PortabilityWorkflow, VSyncWorkflow│
├─────────────────────────────────────────────────────┤
│  Application Layer (Use Cases)                      │  ← Workflow orchestration logic
│  - Workflow Services, Activity Implementations      │
├─────────────────────────────────────────────────────┤
│  Domain Layer (Business Rules)                      │  ← Claim/Entry rules, State machines
│  - Workflow State, Business Validators              │
├─────────────────────────────────────────────────────┤
│  Infrastructure Layer (External Interfaces)         │  ← Bridge, Pulsar, Redis, Database
│  - gRPC Clients, Pulsar Consumers, Repositories     │
└─────────────────────────────────────────────────────┘
```

**Diferença do Core DICT**:
- ✅ Temporal Workflows são a camada de orquestração (substituem controllers)
- ✅ Activities são as operações externas (chamadas ao Bridge, notificações)
- ✅ Domain Layer contém state machines e regras de transição de workflow

---

## 2. C4 Component Diagram - RSFN Connect

### 2.1. Diagrama

```mermaid
C4Component
  title Component Diagram - RSFN Connect (Clean Architecture)

  Container_Boundary(rsfn_connect, "RSFN Connect") {

    Component_Boundary(workflow_layer, "Workflow Layer (Temporal)") {
      Component(claim_workflow, "ClaimWorkflow", "Go + Temporal SDK", "Workflow durável de 30 dias para reivindicações")
      Component(portability_workflow, "PortabilityWorkflow", "Go + Temporal SDK", "Workflow de portabilidade de conta")
      Component(vsync_workflow, "VSyncWorkflow", "Go + Temporal SDK", "Sincronização diária com Bacen")
      Component(create_entry_workflow, "CreateEntryWorkflow", "Go + Temporal SDK", "Workflow de criação de chave PIX")
      Component(workflow_interceptor, "Workflow Interceptor", "Go", "Logging, tracing, metrics para workflows")
    }

    Component_Boundary(application_layer, "Application Layer (Activities)") {
      Component(bridge_activity, "Bridge Activities", "Go", "Activities: CreateEntry, CreateClaim, CompleteClaim, CancelClaim")
      Component(notification_activity, "Notification Activities", "Go", "Activities: NotifyOwner, NotifyBoth, SendEmail, SendSMS")
      Component(database_activity, "Database Activities", "Go", "Activities: UpdateClaimStatus, UpdateEntryStatus, LogAudit")
      Component(cache_activity, "Cache Activities", "Go", "Activities: CacheEntry, InvalidateCache, GetFromCache")
      Component(activity_options, "Activity Options", "Go", "Retry policies, timeouts, heartbeats")
    }

    Component_Boundary(consumer_layer, "Event Consumer Layer") {
      Component(pulsar_consumer, "Pulsar Consumer", "Go + pulsar-client-go", "Consome eventos dict.entries.*, dict.claims.*")
      Component(event_handler, "Event Handler", "Go", "Mapeia eventos para workflows (StartWorkflow)")
      Component(consumer_config, "Consumer Config", "Go", "Subscription, ack timeout, redelivery policy")
      Component(idempotency_checker, "Idempotency Checker", "Go", "Garante processamento único de eventos")
    }

    Component_Boundary(api_layer, "API Layer (gRPC)") {
      Component(connect_grpc_server, "Connect gRPC Server", "Go + gRPC", "API interna para BackOffice")
      Component(workflow_query_service, "Workflow Query Service", "Go", "Consulta status de workflows")
      Component(workflow_signal_service, "Workflow Signal Service", "Go", "Envia signals para workflows (confirm, cancel)")
      Component(health_check_service, "Health Check Service", "Go", "Health e readiness checks")
    }

    Component_Boundary(domain_layer, "Domain Layer") {
      Component(workflow_state, "Workflow State", "Go", "State machine de workflows (OPEN, WAITING, COMPLETED)")
      Component(claim_validator, "Claim Validator", "Go", "Valida regras de claim (30 dias, ISPB diferente)")
      Component(entry_validator, "Entry Validator", "Go", "Valida regras de entry (key unique, account valid)")
      Component(workflow_events, "Workflow Events", "Go", "WorkflowStarted, ActivityCompleted, WorkflowCompleted")
    }

    Component_Boundary(infrastructure_layer, "Infrastructure Layer") {
      Component(bridge_grpc_client, "Bridge gRPC Client", "Go + gRPC", "Cliente gRPC para chamar Bridge")
      Component(notification_http_client, "Notification HTTP Client", "Go + resty", "Cliente HTTP para LBPay Notifications")
      Component(pulsar_producer, "Pulsar Producer", "Go + pulsar-client-go", "Publica eventos de resposta")
      Component(redis_client, "Redis Client", "Go + go-redis", "Cache e idempotency keys")
      Component(postgres_repository, "Postgres Repository", "Go + pgx", "Persiste workflow state e history")
      Component(temporal_client, "Temporal Client", "Go + Temporal SDK", "Interage com Temporal Server")
    }
  }

  ContainerDb(temporal_server, "Temporal Server", "Temporal v1.36.0", "Orquestrador de workflows")
  ContainerDb(connect_db, "Connect Database", "PostgreSQL 16", "Workflow state, history")
  ContainerDb(redis_cache, "Redis Cache", "Redis v9.14.1", "Cache, idempotency")
  ContainerQueue(pulsar, "Apache Pulsar", "Pulsar v0.16.0", "Event streaming")
  Container(bridge_api, "Bridge gRPC API", "Go", "Comunica com Bacen")
  System_Ext(notifications, "LBPay Notifications", "Email/SMS")
  System_Ext(bacen, "Bacen DICT", "SOAP/XML")

  Rel(pulsar, pulsar_consumer, "Entrega eventos", "Pulsar Protocol")
  Rel(pulsar_consumer, event_handler, "Parse event")
  Rel(event_handler, idempotency_checker, "Check duplicates")
  Rel(idempotency_checker, redis_client, "Get/Set idempotency key")

  Rel(event_handler, temporal_client, "StartWorkflow")
  Rel(temporal_client, temporal_server, "ExecuteWorkflow", "gRPC")

  Rel(temporal_server, claim_workflow, "Execute")
  Rel(claim_workflow, workflow_interceptor, "Intercepted")
  Rel(claim_workflow, bridge_activity, "Execute activity")
  Rel(claim_workflow, notification_activity, "Execute activity")
  Rel(claim_workflow, database_activity, "Execute activity")

  Rel(bridge_activity, bridge_grpc_client, "Call Bridge")
  Rel(bridge_grpc_client, bridge_api, "gRPC mTLS")

  Rel(notification_activity, notification_http_client, "Send notification")
  Rel(notification_http_client, notifications, "HTTPS REST")

  Rel(database_activity, postgres_repository, "Persist state")
  Rel(postgres_repository, connect_db, "SQL")

  Rel(cache_activity, redis_client, "Cache operations")
  Rel(redis_client, redis_cache, "Redis Protocol")

  Rel(connect_grpc_server, workflow_query_service, "Query workflow status")
  Rel(workflow_query_service, temporal_client, "QueryWorkflow")

  Rel(connect_grpc_server, workflow_signal_service, "Send signal")
  Rel(workflow_signal_service, temporal_client, "SignalWorkflow")

  Rel(workflow_signal_service, claim_workflow, "Signal: confirm/cancel")

  UpdateLayoutConfig($c4ShapeInRow="3", $c4BoundaryInRow="1")
```

---

### 2.2. Versão PlantUML (Alternativa)

```plantuml
@startuml
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Component.puml

LAYOUT_WITH_LEGEND()

title Component Diagram - RSFN Connect

Container_Boundary(connect, "RSFN Connect") {

  ' Workflow Layer
  Component(claim_wf, "ClaimWorkflow", "Temporal", "30 dias workflow")
  Component(portability_wf, "PortabilityWorkflow", "Temporal", "Portability workflow")
  Component(vsync_wf, "VSyncWorkflow", "Temporal", "Daily sync")

  ' Application Layer
  Component(bridge_act, "Bridge Activities", "Go", "CreateEntry, CreateClaim, etc")
  Component(notif_act, "Notification Activities", "Go", "Notify users")
  Component(db_act, "Database Activities", "Go", "Update state")

  ' Event Consumer
  Component(consumer, "Pulsar Consumer", "Go", "Consume events")
  Component(handler, "Event Handler", "Go", "Start workflows")
  Component(idempotency, "Idempotency Checker", "Go", "Deduplicate")

  ' API Layer
  Component(grpc_server, "Connect gRPC Server", "Go", "Internal API")
  Component(query_svc, "Query Service", "Go", "Query workflows")
  Component(signal_svc, "Signal Service", "Go", "Send signals")

  ' Infrastructure
  Component(bridge_client, "Bridge Client", "gRPC", "Call Bridge")
  Component(notif_client, "Notification Client", "HTTP", "Send emails")
  Component(redis_client, "Redis Client", "Go", "Cache")
  Component(pg_repo, "Postgres Repo", "pgx", "Persist state")
  Component(temporal_client, "Temporal Client", "Go", "Temporal SDK")
}

ContainerDb(temporal, "Temporal Server", "v1.36.0")
ContainerDb(db, "Connect DB", "PostgreSQL 16")
ContainerDb(redis, "Redis", "v9.14.1")
ContainerQueue(pulsar, "Pulsar", "v0.16.0")
Container(bridge, "Bridge API", "Go")
System_Ext(notif, "Notifications")

Rel(pulsar, consumer, "Events")
Rel(consumer, handler, "Parse")
Rel(handler, idempotency, "Check")
Rel(idempotency, redis_client, "Get/Set")
Rel(handler, temporal_client, "StartWorkflow")
Rel(temporal_client, temporal, "gRPC")

Rel(temporal, claim_wf, "Execute")
Rel(claim_wf, bridge_act, "Activity")
Rel(claim_wf, notif_act, "Activity")
Rel(claim_wf, db_act, "Activity")

Rel(bridge_act, bridge_client, "Call")
Rel(bridge_client, bridge, "gRPC")

Rel(notif_act, notif_client, "Call")
Rel(notif_client, notif, "HTTP")

Rel(db_act, pg_repo, "Persist")
Rel(pg_repo, db, "SQL")

Rel(grpc_server, query_svc, "Query")
Rel(query_svc, temporal_client, "QueryWorkflow")

Rel(grpc_server, signal_svc, "Signal")
Rel(signal_svc, temporal_client, "SignalWorkflow")

@enduml
```

---

## 3. Componentes por Camada

### 3.1. Workflow Layer (Temporal)

#### ClaimWorkflow
- **Responsabilidade**: Orquestrar reivindicação de chave PIX por 30 dias
- **Tecnologia**: Go + Temporal SDK
- **Estrutura**:
  ```go
  type ClaimWorkflow struct {
      claimID          string
      entryID          string
      claimerAccount   Account
      ownerAccount     Account
      expiresAt        time.Time
      status           ClaimStatus
  }

  func ClaimWorkflow(ctx workflow.Context, params ClaimWorkflowParams) error {
      logger := workflow.GetLogger(ctx)
      logger.Info("Starting ClaimWorkflow", "claim_id", params.ClaimID)

      // Activity Options
      activityOptions := workflow.ActivityOptions{
          StartToCloseTimeout: 30 * time.Second,
          RetryPolicy: &temporal.RetryPolicy{
              InitialInterval:    time.Second,
              BackoffCoefficient: 2.0,
              MaximumInterval:    time.Minute,
              MaximumAttempts:    3,
          },
      }
      ctx = workflow.WithActivityOptions(ctx, activityOptions)

      // Step 1: Create Claim in Bacen
      var bacenClaimID string
      err := workflow.ExecuteActivity(ctx, CreateClaimActivity, params).Get(ctx, &bacenClaimID)
      if err != nil {
          return fmt.Errorf("failed to create claim in bacen: %w", err)
      }

      // Step 2: Update claim with bacen_claim_id
      err = workflow.ExecuteActivity(ctx, UpdateClaimStatusActivity, params.ClaimID, bacenClaimID).Get(ctx, nil)
      if err != nil {
          return fmt.Errorf("failed to update claim status: %w", err)
      }

      // Step 3: Notify Owner
      err = workflow.ExecuteActivity(ctx, NotifyOwnerActivity, params.OwnerAccount, params.ClaimID).Get(ctx, nil)
      if err != nil {
          logger.Warn("Failed to notify owner", "error", err)
          // Não falha workflow se notificação falhar
      }

      // Step 4: Wait for 30 days OR signal (confirm/cancel)
      selector := workflow.NewSelector(ctx)

      // Timer de 30 dias
      timer := workflow.NewTimer(ctx, 30*24*time.Hour)
      selector.AddFuture(timer, func(f workflow.Future) {
          logger.Info("30 days timer expired, auto-confirming claim")
          // Auto-confirm
          err := workflow.ExecuteActivity(ctx, CompleteClaimActivity, params.ClaimID, true, true).Get(ctx, nil)
          if err != nil {
              logger.Error("Failed to auto-confirm claim", "error", err)
          }
      })

      // Signal: confirm
      confirmChannel := workflow.GetSignalChannel(ctx, "confirm")
      selector.AddReceive(confirmChannel, func(c workflow.ReceiveChannel, more bool) {
          logger.Info("Received confirm signal")
          timer.Cancel()
          err := workflow.ExecuteActivity(ctx, CompleteClaimActivity, params.ClaimID, true, false).Get(ctx, nil)
          if err != nil {
              logger.Error("Failed to confirm claim", "error", err)
          }
      })

      // Signal: cancel
      cancelChannel := workflow.GetSignalChannel(ctx, "cancel")
      selector.AddReceive(cancelChannel, func(c workflow.ReceiveChannel, more bool) {
          var reason string
          c.Receive(ctx, &reason)
          logger.Info("Received cancel signal", "reason", reason)
          timer.Cancel()
          err := workflow.ExecuteActivity(ctx, CancelClaimActivity, params.ClaimID, reason).Get(ctx, nil)
          if err != nil {
              logger.Error("Failed to cancel claim", "error", err)
          }
      })

      // Wait for first event (timer or signal)
      selector.Select(ctx)

      // Step 5: Notify both parties
      err = workflow.ExecuteActivity(ctx, NotifyBothActivity, params.OwnerAccount, params.ClaimerAccount, params.ClaimID).Get(ctx, nil)
      if err != nil {
          logger.Warn("Failed to notify parties", "error", err)
      }

      logger.Info("ClaimWorkflow completed", "claim_id", params.ClaimID)
      return nil
  }
  ```
- **Duração**: Até 30 dias
- **Signals**: `confirm`, `cancel`
- **Activities**: CreateClaim, CompleteClaim, CancelClaim, NotifyOwner, NotifyBoth, UpdateClaimStatus
- **Localização**: `internal/workflows/claim_workflow.go`

#### PortabilityWorkflow
- **Responsabilidade**: Orquestrar portabilidade de conta (change account)
- **Estrutura**: Similar ao ClaimWorkflow, mas sem timer de 30 dias
- **Activities**: CreatePortability, ConfirmPortability, UpdateEntryAccount
- **Localização**: `internal/workflows/portability_workflow.go`

#### VSyncWorkflow
- **Responsabilidade**: Sincronização diária com Bacen (validar consistência de entries)
- **Estrutura**:
  ```go
  func VSyncWorkflow(ctx workflow.Context) error {
      // Cron schedule: daily at 2 AM
      err := workflow.Sleep(ctx, workflow.GetInfo(ctx).WorkflowExecutionTimeout)

      // Step 1: Fetch all entries from local DB
      var localEntries []Entry
      err := workflow.ExecuteActivity(ctx, FetchLocalEntriesActivity).Get(ctx, &localEntries)

      // Step 2: Fetch all entries from Bacen
      var bacenEntries []Entry
      err = workflow.ExecuteActivity(ctx, FetchBacenEntriesActivity).Get(ctx, &bacenEntries)

      // Step 3: Compare and reconcile
      var reconcileResults ReconcileResults
      err = workflow.ExecuteActivity(ctx, ReconcileEntriesActivity, localEntries, bacenEntries).Get(ctx, &reconcileResults)

      // Step 4: Log discrepancies
      if len(reconcileResults.Discrepancies) > 0 {
          workflow.GetLogger(ctx).Warn("Found discrepancies", "count", len(reconcileResults.Discrepancies))
      }

      return nil
  }
  ```
- **Schedule**: Diário às 2 AM
- **Localização**: `internal/workflows/vsync_workflow.go`

#### CreateEntryWorkflow
- **Responsabilidade**: Criar chave PIX no Bacen (após evento `dict.entries.created`)
- **Estrutura**:
  ```go
  func CreateEntryWorkflow(ctx workflow.Context, params CreateEntryParams) error {
      // Step 1: Create entry in Bacen
      var bacenEntryID string
      err := workflow.ExecuteActivity(ctx, CreateEntryActivity, params).Get(ctx, &bacenEntryID)
      if err != nil {
          return fmt.Errorf("failed to create entry in bacen: %w", err)
      }

      // Step 2: Update entry status to ACTIVE
      err = workflow.ExecuteActivity(ctx, UpdateEntryStatusActivity, params.EntryID, "ACTIVE", bacenEntryID).Get(ctx, nil)
      if err != nil {
          return fmt.Errorf("failed to update entry status: %w", err)
      }

      // Step 3: Cache entry
      err = workflow.ExecuteActivity(ctx, CacheEntryActivity, params.EntryID).Get(ctx, nil)
      if err != nil {
          workflow.GetLogger(ctx).Warn("Failed to cache entry", "error", err)
      }

      // Step 4: Notify user
      err = workflow.ExecuteActivity(ctx, NotifyUserActivity, params.UserID, "entry_created").Get(ctx, nil)
      if err != nil {
          workflow.GetLogger(ctx).Warn("Failed to notify user", "error", err)
      }

      return nil
  }
  ```
- **Duração**: ~1-2s
- **Localização**: `internal/workflows/create_entry_workflow.go`

#### Workflow Interceptor
- **Responsabilidade**: Interceptar workflows para logging, tracing, metrics
- **Implementação**:
  ```go
  type WorkflowInterceptor struct{}

  func (w *WorkflowInterceptor) InterceptWorkflow(ctx workflow.Context, next workflow.WorkflowInboundInterceptor) error {
      workflowInfo := workflow.GetInfo(ctx)
      logger := workflow.GetLogger(ctx)

      startTime := workflow.Now(ctx)
      logger.Info("Workflow started", "workflow_type", workflowInfo.WorkflowType.Name, "workflow_id", workflowInfo.WorkflowExecution.ID)

      // Propagate tracing context
      ctx = workflow.WithValue(ctx, "trace_id", workflowInfo.WorkflowExecution.ID)

      err := next.Execute(ctx)

      duration := workflow.Now(ctx).Sub(startTime)
      logger.Info("Workflow completed", "duration", duration, "error", err)

      // Emit metric
      workflow.GetMetricsHandler(ctx).Counter("workflow_completed_total").Inc(1)
      workflow.GetMetricsHandler(ctx).Timer("workflow_duration_seconds").Record(duration)

      return err
  }
  ```
- **Localização**: `internal/workflows/interceptors/workflow_interceptor.go`

---

### 3.2. Application Layer (Activities)

#### Bridge Activities
- **Responsabilidade**: Activities que chamam Bridge gRPC API
- **Activities**:
  ```go
  type BridgeActivities struct {
      bridgeClient bridge.BridgeServiceClient
  }

  func (a *BridgeActivities) CreateEntryActivity(ctx context.Context, params CreateEntryParams) (string, error) {
      ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
      defer cancel()

      resp, err := a.bridgeClient.CreateEntry(ctx, &bridge.CreateEntryRequest{
          KeyType:  params.KeyType,
          KeyValue: params.KeyValue,
          Account:  params.Account,
      })
      if err != nil {
          return "", fmt.Errorf("bridge.CreateEntry failed: %w", err)
      }

      return resp.BacenEntryId, nil
  }

  func (a *BridgeActivities) CreateClaimActivity(ctx context.Context, params ClaimParams) (string, error) {
      ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
      defer cancel()

      resp, err := a.bridgeClient.CreateClaim(ctx, &bridge.CreateClaimRequest{
          ClaimId:        params.ClaimID,
          EntryId:        params.EntryID,
          ClaimerAccount: params.ClaimerAccount,
      })
      if err != nil {
          return "", fmt.Errorf("bridge.CreateClaim failed: %w", err)
      }

      return resp.BacenClaimId, nil
  }

  func (a *BridgeActivities) CompleteClaimActivity(ctx context.Context, claimID string, confirmed, autoConfirmed bool) error {
      ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
      defer cancel()

      _, err := a.bridgeClient.CompleteClaim(ctx, &bridge.CompleteClaimRequest{
          ClaimId:       claimID,
          Confirmed:     confirmed,
          AutoConfirmed: autoConfirmed,
      })
      if err != nil {
          return fmt.Errorf("bridge.CompleteClaim failed: %w", err)
      }

      return nil
  }

  func (a *BridgeActivities) CancelClaimActivity(ctx context.Context, claimID, reason string) error {
      ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
      defer cancel()

      _, err := a.bridgeClient.CancelClaim(ctx, &bridge.CancelClaimRequest{
          ClaimId: claimID,
          Reason:  reason,
      })
      if err != nil {
          return fmt.Errorf("bridge.CancelClaim failed: %w", err)
      }

      return nil
  }
  ```
- **Localização**: `internal/activities/bridge_activities.go`

#### Notification Activities
- **Responsabilidade**: Activities que enviam notificações
- **Activities**:
  ```go
  type NotificationActivities struct {
      notificationClient *resty.Client
      notificationURL    string
  }

  func (a *NotificationActivities) NotifyOwnerActivity(ctx context.Context, ownerAccount Account, claimID string) error {
      ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
      defer cancel()

      _, err := a.notificationClient.R().
          SetContext(ctx).
          SetBody(map[string]interface{}{
              "user_id":  ownerAccount.HolderID,
              "template": "claim_received",
              "data": map[string]string{
                  "claim_id": claimID,
                  "link":     fmt.Sprintf("https://app.lbpay.com/claims/%s", claimID),
              },
          }).
          Post(a.notificationURL + "/notifications/send")

      if err != nil {
          return fmt.Errorf("failed to notify owner: %w", err)
      }

      return nil
  }

  func (a *NotificationActivities) NotifyBothActivity(ctx context.Context, ownerAccount, claimerAccount Account, claimID string) error {
      // Notify owner
      err := a.NotifyOwnerActivity(ctx, ownerAccount, claimID)
      if err != nil {
          return err
      }

      // Notify claimer
      ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
      defer cancel()

      _, err = a.notificationClient.R().
          SetContext(ctx).
          SetBody(map[string]interface{}{
              "user_id":  claimerAccount.HolderID,
              "template": "claim_completed",
              "data": map[string]string{
                  "claim_id": claimID,
              },
          }).
          Post(a.notificationURL + "/notifications/send")

      if err != nil {
          return fmt.Errorf("failed to notify claimer: %w", err)
      }

      return nil
  }
  ```
- **Localização**: `internal/activities/notification_activities.go`

#### Database Activities
- **Responsabilidade**: Activities que persistem state no Connect Database
- **Activities**:
  ```go
  type DatabaseActivities struct {
      repo *PostgresRepository
  }

  func (a *DatabaseActivities) UpdateClaimStatusActivity(ctx context.Context, claimID, bacenClaimID string) error {
      return a.repo.UpdateClaimStatus(ctx, claimID, bacenClaimID)
  }

  func (a *DatabaseActivities) UpdateEntryStatusActivity(ctx context.Context, entryID, status, bacenEntryID string) error {
      return a.repo.UpdateEntryStatus(ctx, entryID, status, bacenEntryID)
  }

  func (a *DatabaseActivities) LogAuditActivity(ctx context.Context, log AuditLog) error {
      return a.repo.LogAudit(ctx, log)
  }
  ```
- **Localização**: `internal/activities/database_activities.go`

#### Cache Activities
- **Responsabilidade**: Activities de cache (Redis)
- **Activities**:
  ```go
  type CacheActivities struct {
      redisClient *redis.Client
  }

  func (a *CacheActivities) CacheEntryActivity(ctx context.Context, entryID string) error {
      // Fetch entry from DB
      entry, err := a.repo.GetEntry(ctx, entryID)
      if err != nil {
          return err
      }

      // Cache for 5 minutes
      entryJSON, _ := json.Marshal(entry)
      return a.redisClient.Set(ctx, fmt.Sprintf("dict:entry:%s", entryID), entryJSON, 5*time.Minute).Err()
  }

  func (a *CacheActivities) InvalidateCacheActivity(ctx context.Context, entryID string) error {
      return a.redisClient.Del(ctx, fmt.Sprintf("dict:entry:%s", entryID)).Err()
  }
  ```
- **Localização**: `internal/activities/cache_activities.go`

#### Activity Options
- **Responsabilidade**: Configurar retry policies, timeouts, heartbeats
- **Configuração**:
  ```go
  var DefaultActivityOptions = workflow.ActivityOptions{
      StartToCloseTimeout: 30 * time.Second,
      HeartbeatTimeout:    10 * time.Second,
      RetryPolicy: &temporal.RetryPolicy{
          InitialInterval:        time.Second,
          BackoffCoefficient:     2.0,
          MaximumInterval:        time.Minute,
          MaximumAttempts:        3,
          NonRetryableErrorTypes: []string{"InvalidArgumentError"},
      },
  }

  var BridgeActivityOptions = workflow.ActivityOptions{
      StartToCloseTimeout: 30 * time.Second,
      RetryPolicy: &temporal.RetryPolicy{
          InitialInterval:    time.Second,
          BackoffCoefficient: 2.0,
          MaximumInterval:    time.Minute,
          MaximumAttempts:    3,
      },
  }

  var NotificationActivityOptions = workflow.ActivityOptions{
      StartToCloseTimeout: 10 * time.Second,
      RetryPolicy: &temporal.RetryPolicy{
          InitialInterval:        time.Second,
          BackoffCoefficient:     2.0,
          MaximumInterval:        30 * time.Second,
          MaximumAttempts:        5,
          NonRetryableErrorTypes: []string{},
      },
  }
  ```
- **Localização**: `internal/workflows/activity_options.go`

---

### 3.3. Event Consumer Layer

#### Pulsar Consumer
- **Responsabilidade**: Consumir eventos do Apache Pulsar
- **Tecnologia**: Go + pulsar-client-go
- **Implementação**:
  ```go
  type PulsarConsumer struct {
      client   pulsar.Client
      consumer pulsar.Consumer
      handler  EventHandler
  }

  func NewPulsarConsumer(pulsarURL string, handler EventHandler) (*PulsarConsumer, error) {
      client, err := pulsar.NewClient(pulsar.ClientOptions{
          URL: pulsarURL,
      })
      if err != nil {
          return nil, err
      }

      consumer, err := client.Subscribe(pulsar.ConsumerOptions{
          Topics:           []string{"dict.entries.created", "dict.claims.created", "dict.portabilities.created"},
          SubscriptionName: "rsfn-connect-subscription",
          Type:             pulsar.Shared,
          AckTimeout:       30 * time.Second,
      })
      if err != nil {
          return nil, err
      }

      return &PulsarConsumer{client: client, consumer: consumer, handler: handler}, nil
  }

  func (pc *PulsarConsumer) Start(ctx context.Context) error {
      for {
          select {
          case <-ctx.Done():
              return ctx.Err()
          default:
              msg, err := pc.consumer.Receive(ctx)
              if err != nil {
                  log.Error("Failed to receive message", "error", err)
                  continue
              }

              // Process message
              err = pc.handler.Handle(ctx, msg)
              if err != nil {
                  log.Error("Failed to handle message", "error", err)
                  pc.consumer.Nack(msg)
              } else {
                  pc.consumer.Ack(msg)
              }
          }
      }
  }
  ```
- **Topics**: `dict.entries.created`, `dict.claims.created`, `dict.portabilities.created`
- **Subscription**: `rsfn-connect-subscription` (shared)
- **Localização**: `internal/consumers/pulsar_consumer.go`

#### Event Handler
- **Responsabilidade**: Mapear eventos para workflows
- **Implementação**:
  ```go
  type EventHandler struct {
      temporalClient client.Client
      idempotencyChecker IdempotencyChecker
  }

  func (eh *EventHandler) Handle(ctx context.Context, msg pulsar.Message) error {
      // Parse event
      var event map[string]interface{}
      err := json.Unmarshal(msg.Payload(), &event)
      if err != nil {
          return fmt.Errorf("failed to parse event: %w", err)
      }

      // Check idempotency
      messageID := msg.ID().String()
      isDuplicate, err := eh.idempotencyChecker.IsDuplicate(ctx, messageID)
      if err != nil {
          return err
      }
      if isDuplicate {
          log.Info("Duplicate message, skipping", "message_id", messageID)
          return nil
      }

      // Map event to workflow
      topic := msg.Topic()
      switch topic {
      case "dict.entries.created":
          return eh.startCreateEntryWorkflow(ctx, event)
      case "dict.claims.created":
          return eh.startClaimWorkflow(ctx, event)
      case "dict.portabilities.created":
          return eh.startPortabilityWorkflow(ctx, event)
      default:
          return fmt.Errorf("unknown topic: %s", topic)
      }
  }

  func (eh *EventHandler) startClaimWorkflow(ctx context.Context, event map[string]interface{}) error {
      claimID := event["claim_id"].(string)

      workflowOptions := client.StartWorkflowOptions{
          ID:        fmt.Sprintf("claim-workflow-%s", claimID),
          TaskQueue: "rsfn-connect-task-queue",
      }

      params := ClaimWorkflowParams{
          ClaimID:        claimID,
          EntryID:        event["entry_id"].(string),
          ClaimerAccount: parseAccount(event["claimer_account"]),
          OwnerAccount:   parseAccount(event["owner_account"]),
          ExpiresAt:      parseTime(event["expires_at"]),
      }

      _, err := eh.temporalClient.ExecuteWorkflow(ctx, workflowOptions, ClaimWorkflow, params)
      if err != nil {
          return fmt.Errorf("failed to start ClaimWorkflow: %w", err)
      }

      return nil
  }
  ```
- **Localização**: `internal/consumers/event_handler.go`

#### Idempotency Checker
- **Responsabilidade**: Garantir processamento único de eventos (evitar duplicatas)
- **Implementação**:
  ```go
  type IdempotencyChecker struct {
      redisClient *redis.Client
  }

  func (ic *IdempotencyChecker) IsDuplicate(ctx context.Context, messageID string) (bool, error) {
      key := fmt.Sprintf("idempotency:message:%s", messageID)

      // Try to set key with NX (only if not exists)
      set, err := ic.redisClient.SetNX(ctx, key, "1", 24*time.Hour).Result()
      if err != nil {
          return false, err
      }

      // If set = false, key already exists (duplicate)
      return !set, nil
  }
  ```
- **TTL**: 24 horas
- **Localização**: `internal/consumers/idempotency_checker.go`

---

### 3.4. API Layer (gRPC)

#### Connect gRPC Server
- **Responsabilidade**: API interna para BackOffice
- **Porta**: 9090
- **RPCs**:
  ```protobuf
  service ConnectService {
      rpc GetWorkflowStatus(GetWorkflowStatusRequest) returns (GetWorkflowStatusResponse);
      rpc QueryClaim(QueryClaimRequest) returns (QueryClaimResponse);
      rpc SignalConfirmClaim(SignalConfirmClaimRequest) returns (SignalConfirmClaimResponse);
      rpc SignalCancelClaim(SignalCancelClaimRequest) returns (SignalCancelClaimResponse);
      rpc ListActiveWorkflows(ListActiveWorkflowsRequest) returns (ListActiveWorkflowsResponse);
      rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
  }
  ```
- **Localização**: `internal/api/grpc/connect_server.go`

#### Workflow Query Service
- **Responsabilidade**: Consultar status de workflows
- **Implementação**:
  ```go
  type WorkflowQueryService struct {
      temporalClient client.Client
  }

  func (wqs *WorkflowQueryService) GetWorkflowStatus(ctx context.Context, req *pb.GetWorkflowStatusRequest) (*pb.GetWorkflowStatusResponse, error) {
      workflowID := req.WorkflowId

      // Query workflow
      resp, err := wqs.temporalClient.QueryWorkflow(ctx, workflowID, "", "status")
      if err != nil {
          return nil, status.Errorf(codes.NotFound, "workflow not found: %v", err)
      }

      var workflowStatus string
      err = resp.Get(&workflowStatus)
      if err != nil {
          return nil, status.Errorf(codes.Internal, "failed to get status: %v", err)
      }

      return &pb.GetWorkflowStatusResponse{
          WorkflowId: workflowID,
          Status:     workflowStatus,
      }, nil
  }
  ```
- **Localização**: `internal/api/grpc/workflow_query_service.go`

#### Workflow Signal Service
- **Responsabilidade**: Enviar signals para workflows (confirm, cancel)
- **Implementação**:
  ```go
  type WorkflowSignalService struct {
      temporalClient client.Client
  }

  func (wss *WorkflowSignalService) SignalConfirmClaim(ctx context.Context, req *pb.SignalConfirmClaimRequest) (*pb.SignalConfirmClaimResponse, error) {
      claimID := req.ClaimId
      workflowID := fmt.Sprintf("claim-workflow-%s", claimID)

      // Send signal "confirm"
      err := wss.temporalClient.SignalWorkflow(ctx, workflowID, "", "confirm", nil)
      if err != nil {
          return nil, status.Errorf(codes.NotFound, "workflow not found: %v", err)
      }

      return &pb.SignalConfirmClaimResponse{
          Success: true,
      }, nil
  }

  func (wss *WorkflowSignalService) SignalCancelClaim(ctx context.Context, req *pb.SignalCancelClaimRequest) (*pb.SignalCancelClaimResponse, error) {
      claimID := req.ClaimId
      reason := req.Reason
      workflowID := fmt.Sprintf("claim-workflow-%s", claimID)

      // Send signal "cancel" with reason
      err := wss.temporalClient.SignalWorkflow(ctx, workflowID, "", "cancel", reason)
      if err != nil {
          return nil, status.Errorf(codes.NotFound, "workflow not found: %v", err)
      }

      return &pb.SignalCancelClaimResponse{
          Success: true,
      }, nil
  }
  ```
- **Localização**: `internal/api/grpc/workflow_signal_service.go`

---

### 3.5. Domain Layer

#### Workflow State
- **Responsabilidade**: State machine de workflows
- **Estados**:
  ```go
  type WorkflowState string

  const (
      WorkflowStateOpen              WorkflowState = "OPEN"
      WorkflowStateWaitingResolution WorkflowState = "WAITING_RESOLUTION"
      WorkflowStateCompleted         WorkflowState = "COMPLETED"
      WorkflowStateCancelled         WorkflowState = "CANCELLED"
      WorkflowStateExpired           WorkflowState = "EXPIRED"
      WorkflowStateFailed            WorkflowState = "FAILED"
  )

  type WorkflowStateMachine struct {
      currentState WorkflowState
  }

  func (wsm *WorkflowStateMachine) Transition(event string) error {
      switch wsm.currentState {
      case WorkflowStateOpen:
          if event == "confirm" {
              wsm.currentState = WorkflowStateCompleted
              return nil
          }
          if event == "cancel" {
              wsm.currentState = WorkflowStateCancelled
              return nil
          }
          if event == "expire" {
              wsm.currentState = WorkflowStateExpired
              return nil
          }
          return fmt.Errorf("invalid transition from OPEN: %s", event)
      default:
          return fmt.Errorf("invalid state: %s", wsm.currentState)
      }
  }
  ```
- **Localização**: `internal/domain/workflow_state.go`

---

### 3.6. Infrastructure Layer

#### Bridge gRPC Client
- **Responsabilidade**: Cliente gRPC para chamar Bridge
- **Implementação**:
  ```go
  type BridgeGRPCClient struct {
      conn   *grpc.ClientConn
      client bridge.BridgeServiceClient
  }

  func NewBridgeGRPCClient(bridgeURL string, creds credentials.TransportCredentials) (*BridgeGRPCClient, error) {
      conn, err := grpc.Dial(bridgeURL, grpc.WithTransportCredentials(creds))
      if err != nil {
          return nil, err
      }

      client := bridge.NewBridgeServiceClient(conn)
      return &BridgeGRPCClient{conn: conn, client: client}, nil
  }

  func (bgc *BridgeGRPCClient) CreateEntry(ctx context.Context, req *bridge.CreateEntryRequest) (*bridge.CreateEntryResponse, error) {
      return bgc.client.CreateEntry(ctx, req)
  }

  func (bgc *BridgeGRPCClient) CreateClaim(ctx context.Context, req *bridge.CreateClaimRequest) (*bridge.CreateClaimResponse, error) {
      return bgc.client.CreateClaim(ctx, req)
  }
  ```
- **Localização**: `internal/infrastructure/clients/bridge_grpc_client.go`

#### Redis Client
- **Responsabilidade**: Cache e idempotency keys
- **Uso**: Cache de entries, claims, idempotency keys
- **Localização**: `internal/infrastructure/cache/redis_client.go`

---

## 4. Estrutura de Diretórios

```
rsfn-connect/
├── cmd/
│   ├── worker/
│   │   └── main.go                      # Temporal Worker
│   └── api/
│       └── main.go                      # Connect gRPC Server
├── internal/
│   ├── workflows/                       # Workflow Layer
│   │   ├── claim_workflow.go
│   │   ├── portability_workflow.go
│   │   ├── vsync_workflow.go
│   │   ├── create_entry_workflow.go
│   │   ├── activity_options.go
│   │   └── interceptors/
│   │       └── workflow_interceptor.go
│   ├── activities/                      # Application Layer
│   │   ├── bridge_activities.go
│   │   ├── notification_activities.go
│   │   ├── database_activities.go
│   │   └── cache_activities.go
│   ├── consumers/                       # Event Consumer Layer
│   │   ├── pulsar_consumer.go
│   │   ├── event_handler.go
│   │   ├── idempotency_checker.go
│   │   └── consumer_config.go
│   ├── api/                             # API Layer
│   │   └── grpc/
│   │       ├── connect_server.go
│   │       ├── workflow_query_service.go
│   │       ├── workflow_signal_service.go
│   │       └── health_check_service.go
│   ├── domain/                          # Domain Layer
│   │   ├── workflow_state.go
│   │   ├── claim_validator.go
│   │   └── entry_validator.go
│   └── infrastructure/                  # Infrastructure Layer
│       ├── clients/
│       │   ├── bridge_grpc_client.go
│       │   ├── notification_http_client.go
│       │   └── temporal_client.go
│       ├── cache/
│       │   └── redis_client.go
│       └── repositories/
│           └── postgres_repository.go
├── proto/
│   └── connect.proto                    # Connect gRPC API
├── go.mod
└── go.sum
```

---

## 5. Fluxo de Requisição Completo

### Exemplo: Evento `dict.claims.created` → ClaimWorkflow

```
1. Apache Pulsar
   └→ Publica evento dict.claims.created
   ↓
2. Pulsar Consumer
   └→ Consome evento
   ↓
3. Event Handler
   ├→ Parse evento (JSON → struct)
   └→ Idempotency Checker
       └→ Redis: Check duplicate (GET idempotency:message:{id})
           ├→ Se EXISTS → Skip (já processado)
           └→ Se NOT EXISTS → Continue
   ↓
4. Event Handler
   └→ Temporal Client.ExecuteWorkflow(ClaimWorkflow, params)
   ↓
5. Temporal Server
   └→ Persiste workflow state no Connect Database
   ↓
6. ClaimWorkflow (Step 1)
   └→ Execute Activity: CreateClaimActivity
       └→ Bridge gRPC Client
           └→ Bridge API → Bacen DICT (SOAP)
   ↓
7. ClaimWorkflow (Step 2)
   └→ Execute Activity: UpdateClaimStatusActivity
       └→ Postgres Repository
           └→ UPDATE connect.claim_workflows SET bacen_claim_id = '...'
   ↓
8. ClaimWorkflow (Step 3)
   └→ Execute Activity: NotifyOwnerActivity
       └→ Notification HTTP Client
           └→ LBPay Notifications (HTTPS)
   ↓
9. ClaimWorkflow (Step 4)
   └→ SetTimer(30 dias)
       └→ Temporal Server persiste timer
       └→ Workflow DORME
   ↓
10. ... 30 dias depois OU Signal recebido ...
   ↓
11. Temporal Server
    └→ TimerFired OU SignalReceived
        └→ Acorda ClaimWorkflow
   ↓
12. ClaimWorkflow (Step 5)
    └→ Execute Activity: CompleteClaimActivity
        └→ Bridge gRPC Client → Bacen
   ↓
13. ClaimWorkflow (Step 6)
    └→ Execute Activity: NotifyBothActivity
        └→ Notification HTTP Client
   ↓
14. Temporal Server
    └→ CompleteWorkflow(success)
```

**Duração Total**: Até 30 dias

---

## 6. Testes por Camada

### 6.1. Workflow Layer - Unit Tests

```go
func TestClaimWorkflow_Success(t *testing.T) {
    testSuite := &testsuite.WorkflowTestSuite{}
    env := testSuite.NewTestWorkflowEnvironment()

    // Mock activities
    env.OnActivity(CreateClaimActivity, mock.Anything, mock.Anything).Return("bacen_claim_123", nil)
    env.OnActivity(UpdateClaimStatusActivity, mock.Anything, mock.Anything, mock.Anything).Return(nil)
    env.OnActivity(NotifyOwnerActivity, mock.Anything, mock.Anything, mock.Anything).Return(nil)
    env.OnActivity(CompleteClaimActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
    env.OnActivity(NotifyBothActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

    // Start workflow
    env.ExecuteWorkflow(ClaimWorkflow, ClaimWorkflowParams{
        ClaimID: "claim_123",
        EntryID: "entry_456",
    })

    // Simulate timer firing (30 days)
    env.RegisterDelayedCallback(func() {
        env.SignalWorkflow("confirm", nil)
    }, 30*24*time.Hour)

    // Assert workflow completed
    require.True(t, env.IsWorkflowCompleted())
    require.NoError(t, env.GetWorkflowError())
}
```

### 6.2. Activities - Unit Tests

```go
func TestCreateClaimActivity_Success(t *testing.T) {
    mockBridgeClient := new(MockBridgeClient)
    activities := &BridgeActivities{bridgeClient: mockBridgeClient}

    mockBridgeClient.On("CreateClaim", mock.Anything, mock.Anything).
        Return(&bridge.CreateClaimResponse{BacenClaimId: "bacen_123"}, nil)

    bacenClaimID, err := activities.CreateClaimActivity(context.Background(), ClaimParams{
        ClaimID: "claim_123",
        EntryID: "entry_456",
    })

    assert.NoError(t, err)
    assert.Equal(t, "bacen_123", bacenClaimID)
    mockBridgeClient.AssertExpectations(t)
}
```

---

## 7. Próximos Passos

1. **[DIA-005: C4 Component Diagram - RSFN Bridge](./DIA-005_C4_Component_Diagram_Bridge.md)** (a criar)
   - Componentes do Bridge (SOAP Adapter, XML Signer)

2. **[TSP-001: Temporal Workflow Engine](../../11_Especificacoes_Tecnicas/TSP-001_Temporal_Workflow_Engine.md)** (a criar)
   - Especificação técnica detalhada do Temporal

3. **[IMP-002: Manual Implementação Connect](../../09_Implementacao/IMP-002_Manual_Implementacao_Connect.md)** (a criar)
   - Guia de implementação passo a passo

---

## 8. Checklist de Validação

- [ ] Clean Architecture está clara (4 camadas)?
- [ ] Workflows são duráveis e sobrevivem a restarts?
- [ ] Activities têm retry policies configuradas?
- [ ] Idempotency checker previne duplicatas?
- [ ] Signals (confirm/cancel) estão implementados?
- [ ] Timer de 30 dias está configurado?
- [ ] Pulsar Consumer usa acknowledgment?
- [ ] gRPC API expõe query e signal services?

---

## 9. Referências

### Documentos Internos
- [DIA-002: C4 Container Diagram](./DIA-002_C4_Container_Diagram.md)
- [DIA-006: Sequence Diagram - ClaimWorkflow](./DIA-006_Sequence_Claim_Workflow.md)
- [TEC-003 v2.1: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [GRPC-001: Bridge gRPC Service](../../04_APIs/gRPC/GRPC-001_Bridge_gRPC_Service.md)

### Documentos Externos
- [Temporal Workflows](https://docs.temporal.io/workflows)
- [Temporal Activities](https://docs.temporal.io/activities)
- [Temporal Signals](https://docs.temporal.io/dev-guide/go/features#signals)
- [Apache Pulsar](https://pulsar.apache.org/docs/concepts-messaging/)

---

**Última Revisão**: 2025-10-25
**Aprovado por**: Arquitetura LBPay
**Próxima Revisão**: 2026-01-25 (trimestral)
