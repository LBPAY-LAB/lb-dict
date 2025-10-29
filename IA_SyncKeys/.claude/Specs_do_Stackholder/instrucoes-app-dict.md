# Instruções para Criação de Novos Endpoints na Dict API

## Contexto Geral

A **Dict API** (`apps/dict`) é uma aplicação REST construída com **Huma** (framework HTTP) que segue padrões de **Clean Architecture** e **Hexagonal Architecture**. A aplicação lida com operações síncronas (GET) via gRPC e operações assíncronas (POST/PUT/DELETE) via Pulsar com cache Redis.

---

## Arquitetura em Camadas

### 1. **Handlers (HTTP)** - `handlers/http/`

Responsável por receber requisições REST e validar schemas.

### 2. **Application** - `application/`

Contém regras de negócio, consultas ao cache, publicação de eventos Pulsar e chamadas gRPC.

### 3. **Domain** - `domain/`

Entidades de domínio, erros customizados e lógica de negócio pura.

### 4. **Infrastructure** - `infrastructure/`

Implementações concretas de publishers (Pulsar), cache (Redis), clientes gRPC e hasher.

### 5. **Setup** - `setup/`

Injeção de dependências e inicialização da aplicação.

---

## Padrões de Implementação

### 📌 **Operações Assíncronas (POST, PUT, DELETE)**

**Fluxo:**

1. Gerar hash determinístico (`requestID`) do payload usando `domain.Fingerprint()`
2. Consultar cache Redis para verificar se já existe resposta
3. Se houver erro no cache (`GetCachedWithError`), retornar erro
4. Se houver resposta no cache, retornar imediatamente (idempotência)
5. Se não houver resposta, publicar evento no Pulsar
6. Retornar `requestID` e resposta vazia (aceita para processamento assíncrono)

**Exemplo de referência:** `CreateClaim`, `ConfirmClaim`, `CancelClaim`, `CompleteClaim`

### 📌 **Operações Síncronas (GET)**

**Fluxo:**

1. Validar schema de entrada
2. Chamar cliente gRPC (bridge) diretamente
3. Retornar resposta mapeada

**Exemplo de referência:** `GetClaim`, `ListClaims`

---

## Checklist de Implementação

### ✅ **1. Schemas (handlers/http/schemas/)**

Criar dois schemas por endpoint:

#### **Request Schema:**

```go
package <resource>

import (
    pkg<Resource> "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/bacen/<resource>"
)

type <Action><Resource>RequestSchema struct {
    Body pkg<Resource>.<Action><Resource>Request `json:"body" doc:"Corpo da requisição"`
    // Adicionar headers/path params conforme necessário
}
```

#### **Response Schema:**

```go
type <Action><Resource>ResponseSchema struct {
    Body <Action><Resource>Body `json:"body" doc:"Descrição da resposta"`
}

type <Action><Resource>Body struct {
    RequestID string `json:"request_id,omitempty" doc:"request_id determinístico para idempotência"`
    Response  *pkg<Resource>.<Action><Resource>Response `json:"response,omitempty" doc:"Corpo completo da resposta"`
}

func MapTo<Action><Resource>ResponseSchema(reqID string, resp *pkg<Resource>.<Action><Resource>Response) *<Action><Resource>ResponseSchema {
    responseSchema := <Action><Resource>ResponseSchema{
        Body: <Action><Resource>Body{
            RequestID: reqID,
            Response:  resp,
        },
    }

    if resp != nil {
        responseSchema.Status = 200
    }

    return &responseSchema
}
```

**⚠️ Importante:**

- Utilizar tipos do SDK compartilhado: `github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/bacen/<resource>`
- Para GET requests, não precisa de `RequestID` no response
- Adicionar validações via tags: `validate:"required,uuid"`, `header:"PI-RequestingParticipant"`, etc.

---

### ✅ **2. Controller (handlers/http/<resource>/)**

#### **controller.go**

```go
package <resource>

import (
    "net/http"
    "github.com/danielgtaylor/huma/v2"
    "github.com/lb-conn/connector-dict/apps/dict/application/<resource>"
)

type Controller struct {
    <resource>App *<resource>.Application
}

func (c *Controller) RegisterRoutes(api huma.API) {
    // GET - operação síncrona
    huma.Register(api, huma.Operation{
        Method:        http.MethodGet,
        Path:          "/<resources>/{id}",
        DefaultStatus: http.StatusOK,
        Tags:          []string{"<Resource>"},
        Summary:       "Descrição curta",
        Description:   "Descrição detalhada",
        OperationID:   "get-<resource>",
    }, c.Get<Resource>Handler)

    // POST - operação assíncrona
    huma.Register(api, huma.Operation{
        Method:        http.MethodPost,
        Path:          "/<resources>",
        DefaultStatus: http.StatusAccepted, // 202 para async
        Tags:          []string{"<Resource>"},
        Summary:       "Descrição curta",
        Description:   "Descrição detalhada",
        OperationID:   "create-<resource>",
    }, c.Create<Resource>Handler)
}

func NewController(<resource>App *<resource>.Application) *Controller {
    return &Controller{<resource>App: <resource>App}
}
```

#### **handlers individuais (ex: create\_<resource>.go)**

**Para operações assíncronas (POST/PUT/DELETE):**

```go
package <resource>

import (
    "context"
    "github.com/lb-conn/connector-dict/apps/dict/domain"
    "github.com/lb-conn/connector-dict/apps/dict/handlers/http/adapters"
    internal "github.com/lb-conn/connector-dict/apps/dict/handlers/http/schemas/<resource>"
)

func (c *Controller) Create<Resource>Handler(ctx context.Context, req *internal.Create<Resource>RequestSchema) (*internal.Create<Resource>ResponseSchema, error) {
    if err := req.Body.Validate(); err != nil {
        return nil, domain.NewRFC9457Error400(domain.ErrBadRequest.Title, err.Error())
    }

    resp, hash, err := c.<resource>App.Create<Resource>(ctx, req.Body)
    if err != nil {
        return nil, adapters.ConvertDomainError(err)
    }

    return internal.MapToCreate<Resource>ResponseSchema(hash, resp), nil
}
```

**Para operações síncronas (GET):**

```go
func (c *Controller) Get<Resource>Handler(ctx context.Context, req *internal.Get<Resource>Request) (*internal.Get<Resource>ResponseSchema, error) {
    payload := pkg<Resource>.Get<Resource>Request{
        ID: req.ID,
        Headers: pkg<Resource>.Get<Resource>RequestHeaders{
            PIRequestingParticipant: req.PIRequestingParticipant,
        },
    }

    if err := payload.Validate(); err != nil {
        return nil, domain.NewRFC9457Error400(domain.ErrBadRequest.Title, err.Error())
    }

    <resource>, err := c.<resource>App.Get<Resource>(ctx, &payload)
    if err != nil {
        return nil, adapters.ConvertDomainError(err)
    }

    resp := internal.MapToGet<Resource>ResponseSchema(*<resource>)
    return &resp, nil
}
```

**⚠️ Importante:**

- Sempre validar schema com `req.Body.Validate()`
- Sempre converter erros com `adapters.ConvertDomainError(err)`
- Para operações assíncronas, retornar `StatusAccepted` (202)
- Para operações síncronas, retornar `StatusOK` (200)

---

### ✅ **3. Application Layer (application/<resource>/)**

#### **interface.go**

```go
package <resource>

import (
    "context"
    pkg<Resource> "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/bacen/<resource>"
)

type <Resource>Service interface {
    Create<Resource>(ctx context.Context, input pkg<Resource>.Create<Resource>Request) (*pkg<Resource>.Create<Resource>Response, error)
    // Adicionar outros métodos conforme necessário
}

// Interface para cliente gRPC (operações síncronas)
type Client<Resource> interface {
    Get<Resource>(ctx context.Context, req *pkg<Resource>.Get<Resource>Request) (*pkg<Resource>.Get<Resource>Response, error)
    List<Resource>s(ctx context.Context, req *pkg<Resource>.List<Resource>sRequest) (*pkg<Resource>.List<Resource>sResponse, error)
}
```

#### **application.go**

```go
package <resource>

import (
    "context"
    "github.com/lb-conn/connector-dict/apps/dict/application/cache"
    "github.com/lb-conn/connector-dict/apps/dict/application/publisher"
    "github.com/lb-conn/connector-dict/apps/dict/domain"
    observability "github.com/lb-conn/connector-dict/shared/infrastructure/observability/interfaces"
    "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg"
    "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/bacen/<resource>"
)

type Application struct {
    observer        observability.Provider
    <resource>      <Resource>Service
    cache           cache.Cache
    client<Resource> Client<Resource>
    createPublisher publisher.Publisher
    updatePublisher publisher.Publisher
    deletePublisher publisher.Publisher
}

func NewApplication(
    observer observability.Provider,
    <resource> <Resource>Service,
    cache cache.Cache,
    client<Resource> Client<Resource>,
    createPublisher publisher.Publisher,
    updatePublisher publisher.Publisher,
    deletePublisher publisher.Publisher,
) *Application {
    return &Application{
        observer:        observer,
        <resource>:      <resource>,
        cache:           cache,
        client<Resource>: client<Resource>,
        createPublisher: createPublisher,
        updatePublisher: updatePublisher,
        deletePublisher: deletePublisher,
    }
}
```

**Para operações assíncronas:**

```go
func (app *Application) Create<Resource>(ctx context.Context, request <resource>.Create<Resource>Request) (*<resource>.Create<Resource>Response, string, error) {
    const op = string(pkg.ActionCreate<Resource>)

    logger := app.observer.Logger()
    logger.InfoWithOperation(ctx, op, "starting Create<Resource>")

    // 1. Gerar hash determinístico (requestID)
    requestID, err := domain.Fingerprint(op, request)
    if err != nil {
        logger.ErrorWithOperation(ctx, op, "failed to compute requestID", err)
        return nil, "", err
    }

    // 2. Consultar cache (sucesso OU erro)
    var resp *<resource>.Create<Resource>Response
    if err := app.cache.GetCachedWithError(ctx, op, requestID, &resp); err != nil {
        return nil, requestID, err
    }

    // 3. Se já existe resposta no cache, retornar (idempotência)
    if resp != nil {
        return resp, requestID, nil
    }

    // 4. Publicar evento no Pulsar
    props := pkg.MessageProperties{
        Action:        pkg.ActionCreate<Resource>,
        CorrelationID: requestID,
    }

    if err := app.createPublisher.Publish(ctx, props, request); err != nil {
        logger.ErrorWithOperation(ctx, op, "failed to publish Create<Resource> event", err,
            observability.String("request_id", requestID),
        )
        return nil, "", err
    }

    // 5. Retornar resposta vazia (aceito para processamento assíncrono)
    return resp, requestID, nil
}
```

**Para operações síncronas (GET):**

```go
func (app *Application) Get<Resource>(ctx context.Context, get *<resource>.Get<Resource>Request) (*<resource>.Get<Resource>Response, error) {
    return app.client<Resource>.Get<Resource>(ctx, get)
}
```

**⚠️ Importante:**

- Sempre usar `domain.Fingerprint(op, request)` para gerar requestID
- Sempre chamar `app.cache.GetCachedWithError()` antes de publicar
- Sempre logar operações importantes com `logger.InfoWithOperation` e `logger.ErrorWithOperation`
- Operações GET sempre delegam para o cliente gRPC

---

### ✅ **4. Infrastructure (infrastructure/)**

#### **gRPC Client (infrastructure/grpc/<resource>/client.go)**

```go
package <resource>

import (
    "context"
    "errors"
    "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/bacen/<resource>"
    pb "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/grpc/<resource>"
    <resource>Mappers "github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/mappers/<resource>"
)

var (
    mapGet<Resource>Request   = <resource>Mappers.MapBacenGet<Resource>RequestToGrpc
    mapGet<Resource>Response  = <resource>Mappers.MapGrpcGet<Resource>ResponseToBacen
)

type <Resource>GRPCClient struct {
    Client pb.<Resource>ServiceClient
}

func (c *<Resource>GRPCClient) Get<Resource>(ctx context.Context, req *<resource>.Get<Resource>Request) (*<resource>.Get<Resource>Response, error) {
    if c == nil || c.Client == nil {
        return nil, errors.New("nil Client")
    }

    pbReq, err := mapGet<Resource>Request(req)
    if err != nil {
        return nil, err
    }

    pbResp, err := c.Client.Get<Resource>(ctx, pbReq)
    if err != nil {
        return nil, err
    }

    return mapGet<Resource>Response(pbResp), nil
}
```

**⚠️ Importante:**

- Utilizar mappers do SDK compartilhado: `github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/mappers/<resource>`
- Sempre validar cliente não-nulo antes de usar

---

### ✅ **5. Setup (setup/)**

#### **Adicionar Publishers ao config.go**

```go
// Configuration
type Configuration struct {
    // ... campos existentes ...
    PulsarTopicCreate<Resource> string // Topic for create <resource> messages
    PulsarTopicUpdate<Resource> string // Topic for update <resource> messages
    PulsarTopicDelete<Resource> string // Topic for delete <resource> messages
}

// NewConfigurationFromEnv
func NewConfigurationFromEnv() Configuration {
    // ... código existente ...

    viper.SetDefault("PULSAR_TOPIC_CREATE_<RESOURCE>", "persistent://lb-conn/dict/orchestration-worker-create-<resource>")
    viper.SetDefault("PULSAR_TOPIC_UPDATE_<RESOURCE>", "persistent://lb-conn/dict/orchestration-worker-update-<resource>")
    viper.SetDefault("PULSAR_TOPIC_DELETE_<RESOURCE>", "persistent://lb-conn/dict/orchestration-worker-delete-<resource>")

    return Configuration{
        // ... campos existentes ...
        PulsarTopicCreate<Resource>: viper.GetString("PULSAR_TOPIC_CREATE_<RESOURCE>"),
        PulsarTopicUpdate<Resource>: viper.GetString("PULSAR_TOPIC_UPDATE_<RESOURCE>"),
        PulsarTopicDelete<Resource>: viper.GetString("PULSAR_TOPIC_DELETE_<RESOURCE>"),
    }
}
```

#### **Adicionar injeção de dependências no setup.go**

```go
// Setup struct
type Setup struct {
    // ... campos existentes ...
    <resource>App  *<resource>.Application
    <resource>Ctrl *<resource>ctrl.Controller
}

// NewSetup
func NewSetup() (*Setup, error) {
    // ... código existente ...

    // Publishers para <resource>
    create<Resource>Publisher := infraPulsar.NewPulsarPublisher(
        setup.publisher,
        config.PulsarTopicCreate<Resource>,
    )

    update<Resource>Publisher := infraPulsar.NewPulsarPublisher(
        setup.publisher,
        config.PulsarTopicUpdate<Resource>,
    )

    delete<Resource>Publisher := infraPulsar.NewPulsarPublisher(
        setup.publisher,
        config.PulsarTopicDelete<Resource>,
    )

    // Application
    setup.<resource>App = <resource>.NewApplication(
        setup.observabilityProvider,
        &dictService.<Resource>,
        setup.redisCache.cache,
        &setup.grpcGateway.Gateway.<Resource>Client,
        create<Resource>Publisher,
        update<Resource>Publisher,
        delete<Resource>Publisher,
    )

    // Controller
    setup.<resource>Ctrl = <resource>ctrl.NewController(setup.<resource>App)

    return setup, nil
}

// RegisterRoutes
func (s Setup) RegisterRoutes() {
    // ... registros existentes ...
    s.<resource>Ctrl.RegisterRoutes(s.humaProcess)
}
```

**⚠️ Importante:**

- Criar um publisher por ação (create, update, delete, etc.)
- Injetar cliente gRPC do gateway para operações GET
- Registrar rotas no método `RegisterRoutes()`

---

### ✅ **6. Domain (domain/)**

Adicionar erros específicos se necessário:

```go
var (
    ErrInvalid<Resource> = &RFC9457Error{
        Status: 400,
        Title:  "Invalid<Resource>",
        Detail: "Descrição do erro específico do recurso.",
    }
)
```

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

## Padrões de Retorno de Erro

### ✅ **Converter Erros de Domínio**

Sempre usar `adapters.ConvertDomainError(err)` nos handlers para converter erros de domínio em respostas HTTP seguindo RFC 9457.

### ✅ **Erros Suportados**

- `DomainError` (interface com `GetStatus()`, `GetTitle()`, `GetDetail()`)
- `bacen.Problem` (erros do SDK compartilhado)
- Erros genéricos (convertidos para 500)

---

## Checklist Final

Antes de concluir, verificar:

- [ ] Schemas criados em `handlers/http/schemas/<resource>/`
- [ ] Controller criado em `handlers/http/<resource>/`
- [ ] Todos os handlers registrados no controller
- [ ] Application layer criada em `application/<resource>/`
- [ ] Interfaces definidas em `application/<resource>/interface.go`
- [ ] Cliente gRPC criado em `infrastructure/grpc/<resource>/` (se necessário)
- [ ] Publishers configurados no `setup/config.go`
- [ ] Injeção de dependências no `setup/setup.go`
- [ ] Rotas registradas no `RegisterRoutes()`
- [ ] Variáveis de ambiente adicionadas
- [ ] Operações assíncronas usando cache + Pulsar
- [ ] Operações síncronas usando gRPC
- [ ] Validação de schemas implementada
- [ ] Conversão de erros aplicada
- [ ] Logs adicionados nas operações importantes

---

## Exemplo Completo: Claim (Referência)

Use o recurso `Claim` como exemplo de referência completo:

- **Schemas:** `handlers/http/schemas/claim/`
- **Controller:** `handlers/http/claim/controller.go`
- **Application:** `application/claim/application.go`
- **gRPC Client:** `infrastructure/grpc/claim/client.go`
- **Setup:** Veja injeção em `setup/setup.go`

**Operações assíncronas:** `CreateClaim`, `ConfirmClaim`, `CancelClaim`, `CompleteClaim`  
**Operações síncronas:** `GetClaim`, `ListClaims`

---

## Notas Importantes

1. **SDK Compartilhado:** Todos os tipos de request/response vêm de `github.com/lb-conn/sdk-rsfn-validator/libs/dict`
2. **Idempotência:** Garantida via hash determinístico (`domain.Fingerprint`) e consulta ao cache
3. **Assincronicidade:** POST/PUT/DELETE publicam no Pulsar e retornam 202 Accepted
4. **Sincronicidade:** GET delega para cliente gRPC e retorna 200 OK
5. **Observabilidade:** Sempre logar operações importantes com contexto
6. **Validação:** Sempre validar schemas antes de processar
7. **Erros:** Sempre converter com `adapters.ConvertDomainError`
8. **Testes:** Seguir padrões em `tests/unit/` para cada camada
