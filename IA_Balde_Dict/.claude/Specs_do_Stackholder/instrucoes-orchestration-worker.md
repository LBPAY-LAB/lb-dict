# Instruções para Criação de Novos Workflows no Orchestration Worker

## Contexto Geral

O **Orchestration Worker** (`apps/orchestration-worker`) é uma aplicação baseada em **Temporal** que consome mensagens do **Pulsar** e orquestra workflows complexos. A aplicação segue padrões de **Clean Architecture** e implementa processos de longa duração com retry automático, monitoramento e cache distribuído.

---

## Arquitetura em Camadas

### 1. **Handlers (Pulsar)** - `handlers/pulsar/`

Consome mensagens do Pulsar e delega para use cases.

### 2. **Application (Use Cases)** - `application/usecases/`

Contém regras de negócio e orquestra a execução de workflows Temporal.

### 3. **Application Ports** - `application/ports/`

Define interfaces (contratos) para serviços externos (cache, publishers, workflows).

### 4. **Infrastructure Temporal** - `infrastructure/temporal/`

Implementação de workflows, activities e services do Temporal.

### 5. **Infrastructure Pulsar/gRPC** - `infrastructure/`

Publishers Pulsar, clientes gRPC e adaptadores.

### 6. **Setup** - `setup/`

Injeção de dependências e inicialização de processos (Temporal, Pulsar, gRPC, Redis).

---

## Tipos de Workflows

### 📌 **1. Workflows de Ação (Create, Update, Delete)**

Executam operações síncronas via gRPC e publicam eventos de sucesso/falha.

**Fluxo padrão:**

1. Executar activity gRPC (criar/atualizar/deletar recurso)
2. Gravar resposta no cache Redis (sucesso ou erro)
3. Publicar evento no CoreEvents (notificação interna)
4. Publicar evento no DictEvents (notificação externa)
5. Iniciar workflows de monitoramento (child workflows) se aplicável

**Exemplo de referência:** `CreateClaimWorkflow`, `CancelClaimWorkflow`, `CompleteClaimWorkflow`

### 📌 **2. Workflows de Monitoramento (Polling)**

Fazem polling periódico de status até atingir condição final ou deadline.

**Fluxo padrão:**

1. Loop de polling com `workflow.Sleep()` e intervalos configuráveis
2. Executar GetActivity para verificar status
3. Verificar condições de saída (status final, deadline, etc.)
4. Solicitar Continue-As-New a cada N iterações (evitar histórico grande)
5. Publicar eventos quando condição final for atingida

**Exemplo de referência:** `MonitorClaimStatusWorkflow`, `ExpireCompletionPeriodEndWorkflow`

---

## Checklist de Implementação

### ✅ **1. Handlers Pulsar (handlers/pulsar/<resource>/)**

#### **<resource>\_handler.go**

```go
package <resource>

import (
    "github.com/lb-conn/connector-dict/apps/orchestration-worker/application/usecases/<resource>"
    "github.com/lb-conn/connector-dict/shared/infrastructure/observability/interfaces"
)

type Handler struct {
    <resource>App    *<resource>.Application
    obsProvider interfaces.Provider
}

func NewHandler(<resource>App *<resource>.Application, obsProvider interfaces.Provider) *Handler {
    return &Handler{
        <resource>App:    <resource>App,
        obsProvider: obsProvider,
    }
}
```

#### **create\_<resource>\_handler.go**

```go
package <resource>

import (
    "context"
    "github.com/lb-conn/libutils/pubsub"
    "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg"
    <resource>sdk "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/bacen/<resource>"
)

func (h *Handler) CreateHandler(ctx context.Context, message pubsub.Message) error {
    logger := h.obsProvider.Logger()

    // Parse message properties (correlation ID, action, etc.)
    props, err := pkg.ParseMessageProperties(message.Properties)
    if err != nil {
        return err
    }

    // Decode message payload
    var request <resource>sdk.Create<Resource>Request
    if err := message.Decode(&request); err != nil {
        logger.Error(ctx, "failed to decode message", err)
        return err
    }

    // Delegate to application use case
    return h.<resource>App.Create<Resource>(ctx, props.CorrelationID, &request)
}
```

**⚠️ Importante:**

- Sempre parsear `MessageProperties` para obter `CorrelationID`
- Sempre logar erros de decode
- Delegar lógica para camada de application

---

### ✅ **2. Application Layer (application/usecases/<resource>/)**

#### **application.go**

```go
package <resource>

import (
    "github.com/lb-conn/connector-dict/apps/orchestration-worker/application/ports"
    "github.com/lb-conn/connector-dict/shared/infrastructure/observability/interfaces"
)

type Application struct {
    obsProvider     interfaces.Provider
    <resource>Service ports.<Resource>Service
}

func NewApplication(<resource>Service ports.<Resource>Service, obsProvider interfaces.Provider) *Application {
    return &Application{
        <resource>Service: <resource>Service,
        obsProvider:     obsProvider,
    }
}
```

#### **create\_<resource>.go**

```go
package <resource>

import (
    "context"
    "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/bacen/<resource>"
)

func (a *Application) Create<Resource>(ctx context.Context, requestID string, request *<resource>.Create<Resource>Request) error {
    return a.<resource>Service.Create<Resource>(ctx, requestID, request)
}
```

**⚠️ Importante:**

- Camada de application é fina: apenas delega para o service (Temporal)
- `requestID` é usado como Workflow ID (idempotência)

---

### ✅ **3. Application Ports (application/ports/<resource>.go)**

```go
package ports

import (
    "context"
    "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/bacen/<resource>"
)

type <Resource>Service interface {
    Create<Resource>(ctx context.Context, requestID string, request *<resource>.Create<Resource>Request) error
    Update<Resource>(ctx context.Context, requestID string, request *<resource>.Update<Resource>Request) error
    Delete<Resource>(ctx context.Context, requestID string, request *<resource>.Delete<Resource>Request) error
}
```

**⚠️ Importante:**

- Ports definem contratos, não implementações
- Implementação concreta fica em `infrastructure/temporal/services/`

---

### ✅ **4. Temporal Workflows (infrastructure/temporal/workflows/<resource>/)**

#### **create_workflow.go**

```go
package <resource>s

import (
    "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/activities"
    <resource>Activities "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/activities/<resource>s"
    "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/workflows"
    "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg"
    pkg<Resource> "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/bacen/<resource>"
    "go.temporal.io/sdk/workflow"
)

type Create<Resource>WorkflowInput struct {
    Request *pkg<Resource>.Create<Resource>Request
    Hash    string
}

func Create<Resource>Workflow(ctx workflow.Context, input Create<Resource>WorkflowInput) error {
    logger := workflow.GetLogger(ctx)

    // 1. Execute gRPC Activity
    bacenResp, err := executeCreate<Resource>Activity(ctx, input)
    if err != nil {
        return err
    }

    // 2. Cache response (success)
    if err := workflows.ExecuteCacheActivity(ctx, input.Hash, bacenResp, false, nil); err != nil {
        logger.Error("CacheActivity failed", "error", err)
        return err
    }

    // 3. Publish to CoreEvents
    if err := workflows.ExecuteCoreEventsPublishActivity(ctx, input.Hash, pkg.ActionCreate<Resource>, bacenResp); err != nil {
        logger.Error("CoreEventsPublishActivity failed", "error", err)
        return err
    }

    // 4. Publish to DictEvents
    if err := workflows.ExecuteDictEventsPublishActivity(ctx, input.Hash, pkg.ActionCreate<Resource>, bacenResp); err != nil {
        logger.Error("DictEventsPublishActivity failed", "error", err)
        return err
    }

    // 5. Start monitoring workflows (if applicable)
    // startMonitor<Resource>Workflow(ctx, bacenResp)

    return nil
}

func executeCreate<Resource>Activity(ctx workflow.Context, input Create<Resource>WorkflowInput) (*pkg<Resource>.Create<Resource>Response, error) {
    ctx = workflow.WithActivityOptions(ctx, activities.GRPCOptions)

    var bacenResp pkg<Resource>.Create<Resource>Response

    err := workflow.ExecuteActivity(ctx, <resource>Activities.Create<Resource>ActivityName, input.Request).Get(ctx, &bacenResp)
    if err != nil {
        workflow.GetLogger(ctx).Error("Create<Resource>Activity failed", "error", err)

        // Notify failure to core
        if notifyErr := workflows.NotifyFailure(ctx, input.Hash, pkg.ActionCreate<Resource>, err); notifyErr != nil {
            workflow.GetLogger(ctx).Error("Failed to notify creation failure", "error", notifyErr)
        }

        return nil, err
    }

    return &bacenResp, nil
}
```

**⚠️ Importante:**

- Sempre usar `workflow.WithActivityOptions()` para configurar timeouts/retry
- Sempre usar `workflows.NotifyFailure()` em caso de erro
- Sempre gravar resposta no cache antes de publicar eventos
- Usar helpers: `ExecuteCacheActivity`, `ExecuteCoreEventsPublishActivity`, `ExecuteDictEventsPublishActivity`

---

#### **monitor\_<resource>\_workflow.go** (Workflow de Monitoramento)

```go
package <resource>s

import (
    "errors"
    "time"
    "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/workflows"
    "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg"
    "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/bacen"
    <resource> "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/bacen/<resource>"
    "go.temporal.io/sdk/workflow"
)

const (
    pollInterval       = 30 * time.Second
    maxMonitorDuration = 5 * time.Minute
)

var (
    workflowMonitor<Resource>Name = "Monitor<Resource>Workflow"
    errRequestContinueAsNew       = errors.New("request-continue-as-new")
)

func Monitor<Resource>Workflow(ctx workflow.Context, input *<resource>.Create<Resource>Response) error {
    logger := workflow.GetLogger(ctx)

    result, err := loopMonitor<Resource>(ctx, input)

    // Continue-As-New para evitar histórico grande
    if errors.Is(err, errRequestContinueAsNew) {
        return workflow.NewContinueAsNewError(ctx, Monitor<Resource>Workflow, input)
    }

    if err != nil {
        logger.Error("erro durante polling", "erro", err)
        return err
    }

    if result == nil {
        return nil
    }

    // Publicar eventos de status final
    action := determineAction(result.Status)

    if err := workflows.ExecuteCoreEventsPublishActivity(ctx, input.ID, action, result); err != nil {
        logger.Error("CoreEventsPublishActivity failed", "error", err)
        return err
    }

    if err := workflows.ExecuteDictEventsPublishActivity(ctx, input.ID, action, result); err != nil {
        logger.Error("DictEventsPublishActivity failed", "error", err)
        return err
    }

    return nil
}

func loopMonitor<Resource>(ctx workflow.Context, input *<resource>.Create<Resource>Response) (*<resource>.Get<Resource>Response, error) {
    logger := workflow.GetLogger(ctx)

    for range int(maxMonitorDuration / pollInterval) {
        // Get<Resource>Activity
        getResp, err := executeGet<Resource>Activity(ctx, input.Participant, input.ID)
        if err != nil {
            return nil, err
        }

        logger.Info("Status atual", "id", getResp.ID, "status", getResp.Status)

        // Verificar condições de saída
        if isFinalStatus(getResp.Status) {
            return getResp, nil
        }

        // Aguardar próximo poll
        if err := workflow.Sleep(ctx, pollInterval); err != nil {
            return nil, err
        }
    }

    // Solicitar Continue-As-New após maxMonitorDuration
    return nil, errRequestContinueAsNew
}

func executeGet<Resource>Activity(ctx workflow.Context, participant, id string) (*<resource>.Get<Resource>Response, error) {
    ctx = workflow.WithActivityOptions(ctx, activities.GRPCOptions)

    var getResp <resource>.Get<Resource>Response

    getReq := <resource>.Get<Resource>Request{
        Headers: <resource>.Get<Resource>RequestHeaders{
            PIRequestingParticipant: participant,
        },
        ID: id,
    }

    err := workflow.ExecuteActivity(ctx, <resource>Activities.Get<Resource>ActivityName, &getReq).Get(ctx, &getResp)
    if err != nil {
        workflow.GetLogger(ctx).Error("Get<Resource>Activity failed", "error", err)
        return nil, err
    }

    return &getResp, nil
}

func isFinalStatus(status string) bool {
    return status == bacen.<Resource>StatusCompleted ||
           status == bacen.<Resource>StatusCancelled
}
```

**⚠️ Importante - Workflows de Monitoramento:**

- Usar `workflow.Sleep()` para polling periódico
- Implementar Continue-As-New a cada N iterações ou após maxDuration
- Sempre verificar condições de saída (status final, deadline, etc.)
- Para monitoramento com deadline específico, calcular `remaining = deadline.Sub(workflow.Now(ctx))`
- Usar margem de segurança: `wait = min(pollInterval, remaining + margin)`

---

#### **expire\_<resource>\_workflow.go** (Workflow de Expiração com Deadline)

```go
package <resource>s

import (
    "errors"
    "time"
    "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/workflows"
    "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg"
    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"
)

var (
    workflowExpire<Resource>Name = "Expire<Resource>Workflow"
    postDeadlineMargin          = 2 * time.Second
)

func Expire<Resource>Workflow(ctx workflow.Context, input *<resource>.Create<Resource>Response) error {
    logger := workflow.GetLogger(ctx)

    // Validar deadline
    deadline, derr := computeDeadline(ctx, input.ExpirationDate)
    if derr != nil {
        logger.Error("ExpirationDate ausente", "id", input.ID, "erro", derr)
        return derr
    }

    // Polling até deadline ou status final
    final, err := pollUntilFinalOrDeadline(ctx, input.ID, input.Participant, deadline)

    if errors.Is(err, errRequestContinueAsNew) {
        logger.Info("Continue-As-New solicitado após 10 polls")
        return workflow.NewContinueAsNewError(ctx, Expire<Resource>Workflow, input)
    }

    if err != nil {
        logger.Error("erro durante polling", "erro", err)
        return err
    }

    if final {
        logger.Info("Recurso alcançou status final — finalizando workflow.")
        return nil
    }

    // Deadline atingido — executar ação automática (ex: cancelamento)
    logger.Info("Deadline atingido — executando ação automática.")

    result, err := executeCancel<Resource>Activity(ctx, &<resource>.Cancel<Resource>Request{
        ID:          input.ID,
        Participant: input.Participant,
    })
    if err != nil {
        return err
    }

    // Publicar eventos
    if err := workflows.ExecuteCoreEventsPublishActivity(ctx, input.ID, pkg.ActionCancel<Resource>, result); err != nil {
        logger.Error("CoreEventsPublishActivity failed", "error", err)
        return err
    }

    if err := workflows.ExecuteDictEventsPublishActivity(ctx, input.ID, pkg.ActionCancel<Resource>, result); err != nil {
        logger.Error("DictEventsPublishActivity failed", "error", err)
        return err
    }

    return nil
}

func pollUntilFinalOrDeadline(ctx workflow.Context, id, participant string, deadline time.Time) (bool, error) {
    logger := workflow.GetLogger(ctx)
    polls := 0

    for {
        polls++

        getResp, err := executeGet<Resource>Activity(ctx, participant, id)
        if err != nil {
            return false, err
        }

        logger.Info("Verificação de status", "id", id, "status", getResp.Status, "polls", polls)

        if isFinalStatus(getResp.Status) {
            return true, nil
        }

        now := workflow.Now(ctx)
        remaining := deadline.Sub(now)

        if remaining <= 0 {
            return false, nil
        }

        // Calcular próximo sleep: min(24h, remaining + margin)
        wait := 24 * time.Hour
        target := remaining + postDeadlineMargin

        if target < wait {
            wait = target
        }

        if wait < time.Second {
            wait = time.Second
        }

        logger.Info("Aguardando próxima verificação", "id", id, "remaining", remaining.String(), "nextWait", wait.String())

        if err := workflow.Sleep(ctx, wait); err != nil {
            return false, err
        }

        // Continue-As-New a cada 10 polls
        if polls%10 == 0 {
            return false, errRequestContinueAsNew
        }
    }
}

func computeDeadline(ctx workflow.Context, expirationDate *time.Time) (time.Time, error) {
    if expirationDate == nil {
        return time.Time{}, temporal.NewNonRetryableApplicationError(
            "missing ExpirationDate",
            "MissingExpirationDate",
            nil,
        )
    }
    return *expirationDate, nil
}
```

**⚠️ Importante - Workflows de Expiração:**

- Sempre validar deadline; retornar `NonRetryableApplicationError` se ausente
- Calcular `remaining = deadline.Sub(workflow.Now(ctx))` a cada iteração
- Usar `wait = min(maxPollInterval, remaining + margin)` com clamp mínimo de 1s
- Solicitar Continue-As-New a cada N polls (ex: 10)
- Executar ação automática (cancelamento, etc.) quando deadline for atingido

---

#### **shared.go** (Helpers compartilhados)

```go
package <resource>s

import (
    "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/activities"
    <resource>Activities "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/activities/<resource>s"
    pkg<Resource> "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/bacen/<resource>"
    "go.temporal.io/sdk/workflow"
)

func executeGet<Resource>Activity(ctx workflow.Context, participant, id string) (*pkg<Resource>.Get<Resource>Response, error) {
    ctx = workflow.WithActivityOptions(ctx, activities.GRPCOptions)

    var getResp pkg<Resource>.Get<Resource>Response

    getReq := pkg<Resource>.Get<Resource>Request{
        Headers: pkg<Resource>.Get<Resource>RequestHeaders{
            PIRequestingParticipant: participant,
        },
        ID: id,
    }

    err := workflow.ExecuteActivity(ctx, <resource>Activities.Get<Resource>ActivityName, &getReq).Get(ctx, &getResp)
    if err != nil {
        workflow.GetLogger(ctx).Error("Get<Resource>Activity failed", "error", err)
        return nil, err
    }

    return &getResp, nil
}

func executeCancel<Resource>Activity(ctx workflow.Context, request *pkg<Resource>.Cancel<Resource>Request) (*pkg<Resource>.Cancel<Resource>Response, error) {
    ctx = workflow.WithActivityOptions(ctx, activities.GRPCOptions)

    var result pkg<Resource>.Cancel<Resource>Response

    err := workflow.ExecuteActivity(ctx, <resource>Activities.Cancel<Resource>ActivityName, request).Get(ctx, &result)
    if err != nil {
        workflow.GetLogger(ctx).Error("Cancel<Resource>Activity failed", "error", err)
        return nil, err
    }

    return &result, nil
}
```

---

#### **Iniciar Child Workflows**

```go
import (
    "fmt"
    "go.temporal.io/api/enums/v1"
    "go.temporal.io/sdk/workflow"
)

func startMonitor<Resource>Workflow(ctx workflow.Context, bacenResp *pkg<Resource>.Create<Resource>Response) {
    ctxMonitor := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
        ParentClosePolicy: enums.PARENT_CLOSE_POLICY_ABANDON,
        WorkflowID:        fmt.Sprintf("%s_%s", workflowMonitor<Resource>Name, bacenResp.ID),
    })

    childWorkflow := workflow.ExecuteChildWorkflow(ctxMonitor, Monitor<Resource>Workflow, bacenResp)

    var execution workflow.Execution
    _ = childWorkflow.GetChildWorkflowExecution().Get(ctx, &execution)
}
```

**⚠️ Importante:**

- Usar `ParentClosePolicy: ABANDON` para workflows de monitoramento (não cancelar se parent terminar)
- Usar Workflow ID único baseado no recurso: `fmt.Sprintf("%s_%s", workflowName, resourceID)`

---

### ✅ **5. Temporal Activities (infrastructure/temporal/activities/<resource>s/)**

#### **<resource>\_activity.go**

```go
package <resource>s

import (
    "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/grpc"
)

type Activity struct {
    grpcGateway *grpc.Gateway
}

func NewActivity(grpcGateway *grpc.Gateway) *Activity {
    return &Activity{grpcGateway: grpcGateway}
}
```

#### **create_activity.go**

```go
package <resource>s

import (
    "context"
    "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/utils"
    "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/bacen/<resource>"
    "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/mappers"
    <resource>Mapper "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/mappers/<resource>"
    "go.temporal.io/sdk/temporal"
)

const Create<Resource>ActivityName = "Create<Resource>GRPCActivity"

func (a *Activity) Create<Resource>Activity(ctx context.Context, req *<resource>.Create<Resource>Request) (*<resource>.Create<Resource>Response, error) {
    grpcResp, err := a.grpcGateway.<Resource>sClient.Create<Resource>(ctx, req)
    if err != nil {
        // Classify error type
        if utils.IsNonRetryableError(err) {
            // Business logic error - don't retry
            return nil, temporal.NewNonRetryableApplicationError(
                "<resource> creation failed due to invalid request",
                "InvalidRequest",
                mappers.GrpcErrorToBacenProblem(err),
            )
        }
        // Transient error - allow retry
        return nil, mappers.GrpcErrorToBacenProblem(err)
    }

    return <resource>Mapper.MapGrpcCreate<Resource>ResponseToBacen(grpcResp), nil
}
```

**⚠️ Importante:**

- Sempre classificar erros: `NonRetryableApplicationError` vs retryable
- Usar `utils.IsNonRetryableError()` para determinar tipo de erro
- Sempre mapear erros gRPC para `bacen.Problem` usando `mappers.GrpcErrorToBacenProblem()`
- Sempre mapear responses usando mappers do SDK compartilhado

---

### ✅ **6. Temporal Service (infrastructure/temporal/services/<resource>\_service.go)**

```go
package services

import (
    "context"
    "github.com/lb-conn/connector-dict/apps/orchestration-worker/application/ports"
    "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/workflows/<resource>s"
    "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/bacen/<resource>"
    "go.temporal.io/api/enums/v1"
    "go.temporal.io/sdk/client"
)

type <Resource>Service struct {
    temporal  client.Client
    taskQueue string
}

var _ ports.<Resource>Service = (*<Resource>Service)(nil)

func New<Resource>Service(temporal client.Client, taskQueue string) <Resource>Service {
    return <Resource>Service{
        temporal:  temporal,
        taskQueue: taskQueue,
    }
}

func (t <Resource>Service) Create<Resource>(ctx context.Context, requestID string, request *<resource>.Create<Resource>Request) error {
    input := <resource>s.Create<Resource>WorkflowInput{
        Request: request,
        Hash:    requestID,
    }

    _, err := t.temporal.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
        ID:                    requestID,
        TaskQueue:             t.taskQueue,
        WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
    }, <resource>s.Create<Resource>Workflow, input)

    return err
}
```

**⚠️ Importante:**

- Usar `requestID` como Workflow ID (garante idempotência)
- Usar `WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY` (permite retry apenas de workflows falhados)
- Implementar interface definida em `application/ports/`

---

### ✅ **7. Setup (setup/)**

#### **Adicionar tópicos ao config.go**

```go
type Config struct {
    // ... campos existentes ...
    PulsarTopicCreate<Resource>  string
    PulsarTopicUpdate<Resource>  string
    PulsarTopicDelete<Resource>  string
}

func NewConfigurationFromEnv() Config {
    // ... código existente ...

    viper.SetDefault("PULSAR_TOPIC_CREATE_<RESOURCE>", "persistent://lb-conn/dict/orchestration-worker-create-<resource>")
    viper.SetDefault("PULSAR_TOPIC_UPDATE_<RESOURCE>", "persistent://lb-conn/dict/orchestration-worker-update-<resource>")
    viper.SetDefault("PULSAR_TOPIC_DELETE_<RESOURCE>", "persistent://lb-conn/dict/orchestration-worker-delete-<resource>")

    return Config{
        // ... campos existentes ...
        PulsarTopicCreate<Resource>: viper.GetString("PULSAR_TOPIC_CREATE_<RESOURCE>"),
        PulsarTopicUpdate<Resource>: viper.GetString("PULSAR_TOPIC_UPDATE_<RESOURCE>"),
        PulsarTopicDelete<Resource>: viper.GetString("PULSAR_TOPIC_DELETE_<RESOURCE>"),
    }
}
```

#### **Registrar workflows e activities no temporal.go**

```go
func NewTemporal(cfg *Config, obsProvider interfaces.Provider, taskQueue string, grpcGateway *grpc.Gateway, publishers PulsarPublishers, cache ports.Cache) (*TemporalProcess, error) {
    // ... código existente ...

    // Workflows
    w.RegisterWorkflow(<resource>s.Create<Resource>Workflow)
    w.RegisterWorkflow(<resource>s.Update<Resource>Workflow)
    w.RegisterWorkflow(<resource>s.Delete<Resource>Workflow)
    w.RegisterWorkflow(<resource>s.Monitor<Resource>Workflow)
    w.RegisterWorkflow(<resource>s.Expire<Resource>Workflow)

    // Activities
    <resource>Activities := activities<Resource>s.NewActivity(grpcGateway)

    w.RegisterActivityWithOptions(<resource>Activities.Create<Resource>Activity, activity.RegisterOptions{
        Name: activities<Resource>s.Create<Resource>ActivityName,
    })
    w.RegisterActivityWithOptions(<resource>Activities.Get<Resource>Activity, activity.RegisterOptions{
        Name: activities<Resource>s.Get<Resource>ActivityName,
    })
    // ... registrar outras activities ...

    return &TemporalProcess{...}, nil
}
```

#### **Adicionar consumer Pulsar no pulsar.go**

```go
type PulsarHandlers struct {
    // ... handlers existentes ...
    <resource>Handler *<resource>.Handler
}

func NewPulsarConsumer(pulsarClient pulsar.Client, cfg *Config, tracer trace.Tracer, propagator propagation.TextMapPropagator, handlers PulsarHandlers) (pubsub.PulsarConsumer, error) {
    consumer, err := pulsarClient.Subscribe(pulsar.ConsumerOptions{
        Topics: []string{
            // ... tópicos existentes ...
            cfg.PulsarTopicCreate<Resource>,
            cfg.PulsarTopicUpdate<Resource>,
            cfg.PulsarTopicDelete<Resource>,
        },
        SubscriptionName: fmt.Sprintf("%s-subscription", cfg.ServiceName),
        Type:             pulsar.Shared,
    })
    if err != nil {
        return pubsub.PulsarConsumer{}, err
    }

    process := pubsub.NewPulsarConsumer(consumer, tracer, propagator)

    // ... registros existentes ...
    process.OnMessage(cfg.PulsarTopicCreate<Resource>, handlers.<resource>Handler.CreateHandler)
    process.OnMessage(cfg.PulsarTopicUpdate<Resource>, handlers.<resource>Handler.UpdateHandler)
    process.OnMessage(cfg.PulsarTopicDelete<Resource>, handlers.<resource>Handler.DeleteHandler)

    return process, nil
}
```

#### **Adicionar injeção de dependências no setup.go**

```go
func NewSetup(ctx context.Context) (*Setup, error) {
    // ... código existente ...

    // Services
    <resource>Service := services.New<Resource>Service(temporalProcess.Client, taskQueue)

    // Applications
    <resource>App := <resource>App.NewApplication(<resource>Service, obs)

    // Handlers
    <resource>Handler := <resource>Hndl.NewHandler(<resource>App, obs)

    // Pulsar Consumer
    pulsarProcess, err := NewPulsarConsumer(pulsarClient, &cfg, otelTracer, tracingProcess.propagator, PulsarHandlers{
        // ... handlers existentes ...
        <resource>Handler: <resource>Handler,
    })

    return &Setup{...}, nil
}
```

---

## Helpers Compartilhados

### ✅ **Cache Activity** (`infrastructure/temporal/activities/cache/`)

```go
// Usar helper para gravar no cache
workflows.ExecuteCacheActivity(ctx, key, value, isError, ttl)
```

### ✅ **Events Activities** (`infrastructure/temporal/activities/events/`)

```go
// Publicar no CoreEvents
workflows.ExecuteCoreEventsPublishActivity(ctx, correlationID, action, payload)

// Publicar no DictEvents
workflows.ExecuteDictEventsPublishActivity(ctx, correlationID, action, payload)
```

### ✅ **NotifyFailure** (Notificar erros)

```go
// Notificar falha (cache + CoreEvents)
workflows.NotifyFailure(ctx, hash, action, err)
```

---

## Activity Options Padrão

```go
// infrastructure/temporal/activities/activity_retry_options.go

var GRPCOptions = workflow.ActivityOptions{
    StartToCloseTimeout: 2 * time.Minute,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:    1 * time.Second,
        BackoffCoefficient: 2.0,
        MaximumInterval:    30 * time.Second,
        MaximumAttempts:    3,
    },
}

var CacheOptions = workflow.ActivityOptions{
    StartToCloseTimeout: 30 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:    500 * time.Millisecond,
        BackoffCoefficient: 2.0,
        MaximumInterval:    10 * time.Second,
        MaximumAttempts:    3,
    },
}

var PublishEventOptions = workflow.ActivityOptions{
    StartToCloseTimeout: 1 * time.Minute,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:    1 * time.Second,
        BackoffCoefficient: 2.0,
        MaximumInterval:    30 * time.Second,
        MaximumAttempts:    5,
    },
}
```

**⚠️ Importante:**

- Sempre usar `workflow.WithActivityOptions()` antes de executar activities
- `GRPCOptions` para chamadas gRPC (2min timeout, 3 retries)
- `CacheOptions` para cache (30s timeout, 3 retries)
- `PublishEventOptions` para eventos (1min timeout, 5 retries)

---

## Variáveis de Ambiente

Adicionar ao `.env`:

```bash
# <Resource> Topics
PULSAR_TOPIC_CREATE_<RESOURCE>=persistent://lb-conn/dict/orchestration-worker-create-<resource>
PULSAR_TOPIC_UPDATE_<RESOURCE>=persistent://lb-conn/dict/orchestration-worker-update-<resource>
PULSAR_TOPIC_DELETE_<RESOURCE>=persistent://lb-conn/dict/orchestration-worker-delete-<resource>
```

---

## Padrões de Erro

### ✅ **Classificação de Erros**

```go
import "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/utils"

if utils.IsNonRetryableError(err) {
    // Erro de negócio (400, 404, etc.) - não tentar novamente
    return nil, temporal.NewNonRetryableApplicationError(
        "operation failed due to invalid request",
        "InvalidRequest",
        mappers.GrpcErrorToBacenProblem(err),
    )
}

// Erro transitório (500, timeout, etc.) - retry automático
return nil, mappers.GrpcErrorToBacenProblem(err)
```

### ✅ **Conversão de Erros**

```go
import "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/mappers"

// Converter erro gRPC para bacen.Problem
bacenErr := mappers.GrpcErrorToBacenProblem(err)

// Converter erro genérico para bacen.Problem
bacenErr := utils.ConvertErrToBacenProblem(err)
```

---

## Checklist Final

Antes de concluir, verificar:

- [ ] Handler Pulsar criado em `handlers/pulsar/<resource>/`
- [ ] Application use case criado em `application/usecases/<resource>/`
- [ ] Port (interface) definido em `application/ports/<resource>.go`
- [ ] Workflows criados em `infrastructure/temporal/workflows/<resource>s/`
- [ ] Activities criadas em `infrastructure/temporal/activities/<resource>s/`
- [ ] Service Temporal criado em `infrastructure/temporal/services/<resource>_service.go`
- [ ] Workflows registrados em `setup/temporal.go`
- [ ] Activities registrados em `setup/temporal.go`
- [ ] Tópicos Pulsar adicionados em `setup/config.go`
- [ ] Consumer Pulsar registrado em `setup/pulsar.go`
- [ ] Injeção de dependências em `setup/setup.go`
- [ ] Variáveis de ambiente adicionadas
- [ ] Workflows de ação executam: gRPC → Cache → CoreEvents → DictEvents
- [ ] Workflows de monitoramento implementam Continue-As-New
- [ ] Activities classificam erros corretamente (retryable vs non-retryable)
- [ ] Helpers compartilhados utilizados (`ExecuteCacheActivity`, etc.)
- [ ] Child workflows usam `ParentClosePolicy: ABANDON`
- [ ] Workflow IDs baseados em `requestID` (idempotência)
- [ ] Logs adicionados em pontos importantes

---

## Exemplo Completo: Claim (Referência)

Use o recurso `Claim` como exemplo de referência completo:

### **Workflows de Ação:**

- `CreateClaimWorkflow` — gRPC → Cache → Events → Child Workflows
- `CancelClaimWorkflow` — gRPC → Cache → Events
- `CompleteClaimWorkflow` — gRPC → Cache → Events

### **Workflows de Monitoramento:**

- `MonitorClaimStatusWorkflow` — Polling 30s até status final + Continue-As-New
- `ExpireCompletionPeriodEndWorkflow` — Polling diário até deadline + Continue-As-New + Cancelamento automático

### **Activities:**

- `CreateClaimActivity` — gRPC call com error classification
- `GetClaimActivity` — gRPC call para polling
- `CancelClaimActivity` — gRPC call para cancelamento

### **Setup:**

- `setup/temporal.go` — Registro de workflows e activities
- `setup/pulsar.go` — Consumer e producers
- `setup/setup.go` — Injeção de dependências

---

## Notas Importantes

1. **Temporal Workflows são determinísticos:** Não use `time.Now()`, use `workflow.Now(ctx)`
2. **Continue-As-New:** Essencial para workflows de longa duração (evita histórico gigante)
3. **ParentClosePolicy:** Use `ABANDON` para child workflows que devem continuar independentemente
4. **Workflow ID:** Sempre usar `requestID` como ID (garante idempotência)
5. **Error Classification:** Separar erros de negócio (non-retryable) de erros transitórios (retryable)
6. **Cache primeiro:** Sempre gravar resposta no cache antes de publicar eventos
7. **Helpers:** Sempre usar helpers compartilhados (`ExecuteCacheActivity`, `ExecuteCoreEventsPublishActivity`, etc.)
8. **Logs:** Sempre usar `workflow.GetLogger(ctx)` em workflows
9. **Activity Options:** Sempre usar options apropriados (`GRPCOptions`, `CacheOptions`, etc.)
10. **Testes:** Seguir padrões em `tests/unit/infrastructure/temporal/` para cada workflow/activity

---

## Fim das Instruções

Seguindo este guia, você conseguirá criar novos workflows e activities consistentes com a arquitetura e padrões estabelecidos no Orchestration Worker.
