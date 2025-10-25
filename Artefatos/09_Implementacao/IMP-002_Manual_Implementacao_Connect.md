# IMP-002: Manual de Implementação - RSFN Connect

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: RSFN Connect (Orchestration Service with Temporal Workflows)
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: BACKEND (AI Agent - Backend Developer)

---

## Sumário Executivo

Este manual fornece instruções passo-a-passo para configurar, implementar e executar o **RSFN Connect**, o módulo orquestrador que gerencia workflows de longa duração (Claims de 30 dias) usando Temporal.

**Baseado em**:
- [TEC-003: RSFN Connect Specification](../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)

---

## Índice

1. [Pré-requisitos](#1-pré-requisitos)
2. [Setup do Repositório](#2-setup-do-repositório)
3. [Setup do Temporal Server](#3-setup-do-temporal-server)
4. [Configuração da Aplicação](#4-configuração-da-aplicação)
5. [Implementação de Workflows](#5-implementação-de-workflows)
6. [Setup de Workers](#6-setup-de-workers)
7. [Integração Pulsar](#7-integração-pulsar)
8. [Integração Redis](#8-integração-redis)
9. [Execução Local](#9-execução-local)
10. [Execução de Testes](#10-execução-de-testes)
11. [Checklist de Implementação](#11-checklist-de-implementação)

---

## 1. Pré-requisitos

### 1.1. Software Necessário

| Software | Versão Mínima | Propósito |
|----------|---------------|-----------|
| **Go** | 1.24+ | Linguagem de programação |
| **Temporal Server** | 1.22+ | Workflow orchestration |
| **Temporal CLI** | 0.10+ | Gerenciamento de workflows |
| **PostgreSQL** | 15+ | Banco de dados (opcional para CID local) |
| **Apache Pulsar** | 3.0+ | Mensageria |
| **Redis** | 7.0+ | Cache |
| **Docker** | 20.10+ | Containerização |
| **Docker Compose** | 2.0+ | Orquestração local |

### 1.2. Variáveis de Ambiente

Criar arquivo `.env` na raiz do projeto:

```bash
# Application Configuration
APP_ENV=development
APP_NAME=rsfn-connect
API_PORT=8080
WORKER_PORT=9090

# Temporal Configuration
TEMPORAL_HOST=localhost:7233
TEMPORAL_NAMESPACE=dict
TEMPORAL_TASK_QUEUE=dict-task-queue

# Pulsar Configuration
PULSAR_URL=pulsar://localhost:6650
PULSAR_API_KEY=
PULSAR_TOPIC_REQ_IN=persistent://lb-conn/dict/rsfn-dict-req-out
PULSAR_TOPIC_RES_OUT=persistent://lb-conn/dict/rsfn-dict-res-out
PULSAR_SUBSCRIPTION=connect-consumer-sub

# Redis Configuration
REDIS_URL=redis://localhost:6379
REDIS_DB=0
REDIS_POOL_SIZE=10

# Bridge gRPC Client
BRIDGE_GRPC_ADDR=localhost:50051
BRIDGE_GRPC_TIMEOUT=30s

# PostgreSQL (opcional - para CID local)
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=rsfn_connect
POSTGRES_USER=connect
POSTGRES_PASSWORD=secure_password_here

# OpenTelemetry
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
ENABLE_TRACING=true

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json
```

---

## 2. Setup do Repositório

### 2.1. Criar Estrutura Multi-App

```bash
# Criar diretório do projeto
mkdir -p connector-dict
cd connector-dict

# Criar estrutura multi-app
mkdir -p apps/dict
mkdir -p apps/orchestration-worker
mkdir -p apps/shared

# Inicializar módulos Go independentes
cd apps/dict
go mod init github.com/lbpay/connector-dict/apps/dict

cd ../orchestration-worker
go mod init github.com/lbpay/connector-dict/apps/orchestration-worker

cd ../shared
go mod init github.com/lbpay/connector-dict/apps/shared
```

### 2.2. Estrutura de Diretórios

```
connector-dict/
├── apps/
│   ├── dict/                              # API REST (Fiber + Huma)
│   │   ├── main.go
│   │   ├── setup/
│   │   │   └── setup.go
│   │   ├── handlers/
│   │   │   ├── rest/
│   │   │   │   ├── entry_handler.go
│   │   │   │   └── claim_handler.go
│   │   │   └── pulsar/
│   │   │       └── dict_consumer.go
│   │   ├── services/
│   │   │   ├── entry_service.go
│   │   │   └── claim_service.go
│   │   └── go.mod
│   │
│   ├── orchestration-worker/              # Temporal Workers
│   │   ├── cmd/worker/
│   │   │   └── main.go
│   │   ├── workflows/
│   │   │   └── claims/
│   │   │       ├── create_workflow.go
│   │   │       ├── monitor_status_workflow.go
│   │   │       ├── expire_completion_period_workflow.go
│   │   │       ├── complete_workflow.go
│   │   │       └── cancel_workflow.go
│   │   ├── activities/
│   │   │   ├── claims/
│   │   │   │   ├── create_activity.go
│   │   │   │   ├── complete_activity.go
│   │   │   │   ├── cancel_activity.go
│   │   │   │   └── get_claim_activity.go
│   │   │   ├── cache/
│   │   │   │   └── cache_activity.go
│   │   │   └── events/
│   │   │       ├── core_events_activity.go
│   │   │       └── dict_events_activity.go
│   │   ├── setup/
│   │   │   └── setup.go
│   │   └── go.mod
│   │
│   └── shared/                            # Infraestrutura compartilhada
│       ├── config/
│       │   └── config.go
│       ├── grpc/
│       │   └── bridge_client.go
│       ├── pulsar/
│       │   ├── consumer.go
│       │   └── producer.go
│       ├── redis/
│       │   └── cache.go
│       ├── temporal/
│       │   └── client.go
│       ├── observability/
│       │   ├── logger.go
│       │   └── tracing.go
│       └── go.mod
│
├── db/
│   └── migrations/                        # Migrations (opcional)
│
├── docker/
│   ├── Dockerfile.dict
│   ├── Dockerfile.worker
│   └── docker-compose.yaml
│
├── .env
└── README.md
```

### 2.3. Instalar Dependências

**Para `apps/dict`**:

```bash
cd apps/dict

# Fiber + Huma (API REST + OpenAPI)
go get github.com/gofiber/fiber/v2
go get github.com/danielgtaylor/huma/v2

# Pulsar Client
go get github.com/apache/pulsar-client-go/pulsar

# Configuration
go get github.com/spf13/viper

# Logging
go get github.com/sirupsen/logrus
```

**Para `apps/orchestration-worker`**:

```bash
cd apps/orchestration-worker

# Temporal SDK
go get go.temporal.io/sdk

# Temporal API
go get go.temporal.io/api
```

**Para `apps/shared`**:

```bash
cd apps/shared

# gRPC
go get google.golang.org/grpc
go get google.golang.org/protobuf

# Redis
go get github.com/redis/go-redis/v9

# Pulsar
go get github.com/apache/pulsar-client-go/pulsar

# OpenTelemetry
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc

# UUID
go get github.com/google/uuid
```

---

## 3. Setup do Temporal Server

### 3.1. Iniciar Temporal via Docker Compose

**Arquivo**: `docker/docker-compose.yaml`

```yaml
version: '3.8'

services:
  temporal:
    image: temporalio/auto-setup:latest
    container_name: temporal
    depends_on:
      - postgresql
      - elasticsearch
    environment:
      - DB=postgresql
      - DB_PORT=5432
      - POSTGRES_USER=temporal
      - POSTGRES_PWD=temporal
      - POSTGRES_SEEDS=postgresql
      - DYNAMIC_CONFIG_FILE_PATH=config/dynamicconfig/development-sql.yaml
      - ENABLE_ES=true
      - ES_SEEDS=elasticsearch
      - ES_VERSION=v7
    ports:
      - "7233:7233"
    volumes:
      - ./dynamicconfig:/etc/temporal/config/dynamicconfig

  temporal-ui:
    image: temporalio/ui:latest
    container_name: temporal-ui
    depends_on:
      - temporal
    environment:
      - TEMPORAL_ADDRESS=temporal:7233
      - TEMPORAL_CORS_ORIGINS=http://localhost:3000
    ports:
      - "8088:8080"

  postgresql:
    image: postgres:15-alpine
    container_name: temporal-postgres
    environment:
      POSTGRES_USER: temporal
      POSTGRES_PASSWORD: temporal
    ports:
      - "5433:5432"
    volumes:
      - temporal_postgres_data:/var/lib/postgresql/data

  elasticsearch:
    image: elasticsearch:7.17.10
    container_name: temporal-elasticsearch
    environment:
      - discovery.type=single-node
      - ES_JAVA_OPTS=-Xms256m -Xmx256m
    ports:
      - "9200:9200"
    volumes:
      - temporal_es_data:/usr/share/elasticsearch/data

  pulsar:
    image: apachepulsar/pulsar:latest
    container_name: pulsar
    command: bin/pulsar standalone
    ports:
      - "6650:6650"
      - "8080:8080"
    volumes:
      - pulsar_data:/pulsar/data

  redis:
    image: redis:7-alpine
    container_name: redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  temporal_postgres_data:
  temporal_es_data:
  pulsar_data:
  redis_data:
```

```bash
# Iniciar serviços
cd docker
docker-compose up -d

# Verificar se Temporal está rodando
docker-compose ps

# Acessar Temporal UI
open http://localhost:8088
```

### 3.2. Criar Namespace no Temporal

```bash
# Instalar Temporal CLI
brew install temporal  # macOS
# ou
go install go.temporal.io/sdk/cmd/temporal@latest

# Criar namespace "dict"
temporal operator namespace create dict

# Verificar namespaces
temporal operator namespace list
```

---

## 4. Configuração da Aplicação

### 4.1. Configuração Compartilhada

**Arquivo**: `apps/shared/config/config.go` (pseudocode)

```go
package config

import (
    "github.com/spf13/viper"
)

type Config struct {
    App      AppConfig
    Temporal TemporalConfig
    Pulsar   PulsarConfig
    Redis    RedisConfig
    Bridge   BridgeConfig
}

type AppConfig struct {
    Env  string
    Name string
    Port string
}

type TemporalConfig struct {
    Host      string
    Namespace string
    TaskQueue string
}

type PulsarConfig struct {
    URL          string
    TopicReqIn   string
    TopicResOut  string
    Subscription string
}

type RedisConfig struct {
    URL      string
    DB       int
    PoolSize int
}

type BridgeConfig struct {
    GRPCAddr string
    Timeout  string
}

func LoadConfig() (*Config, error) {
    viper.SetConfigFile(".env")
    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    config := &Config{
        App: AppConfig{
            Env:  viper.GetString("APP_ENV"),
            Name: viper.GetString("APP_NAME"),
            Port: viper.GetString("API_PORT"),
        },
        Temporal: TemporalConfig{
            Host:      viper.GetString("TEMPORAL_HOST"),
            Namespace: viper.GetString("TEMPORAL_NAMESPACE"),
            TaskQueue: viper.GetString("TEMPORAL_TASK_QUEUE"),
        },
        Pulsar: PulsarConfig{
            URL:          viper.GetString("PULSAR_URL"),
            TopicReqIn:   viper.GetString("PULSAR_TOPIC_REQ_IN"),
            TopicResOut:  viper.GetString("PULSAR_TOPIC_RES_OUT"),
            Subscription: viper.GetString("PULSAR_SUBSCRIPTION"),
        },
        Redis: RedisConfig{
            URL:      viper.GetString("REDIS_URL"),
            DB:       viper.GetInt("REDIS_DB"),
            PoolSize: viper.GetInt("REDIS_POOL_SIZE"),
        },
        Bridge: BridgeConfig{
            GRPCAddr: viper.GetString("BRIDGE_GRPC_ADDR"),
            Timeout:  viper.GetString("BRIDGE_GRPC_TIMEOUT"),
        },
    }

    return config, nil
}
```

---

## 5. Implementação de Workflows

### 5.1. ClaimWorkflow (30 dias)

**Arquivo**: `apps/orchestration-worker/workflows/claims/create_workflow.go` (pseudocode)

```go
package claims

import (
    "time"
    "go.temporal.io/sdk/workflow"
)

type ClaimWorkflowInput struct {
    ClaimID     string
    EntryKey    string
    ClaimerISPB string
    OwnerISPB   string
}

func CreateClaimWorkflow(ctx workflow.Context, input ClaimWorkflowInput) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("CreateClaimWorkflow iniciado", "claimID", input.ClaimID)

    // Activity Options
    activityOptions := workflow.ActivityOptions{
        StartToCloseTimeout: 30 * time.Second,
        RetryPolicy: &workflow.RetryPolicy{
            InitialInterval:    1 * time.Second,
            BackoffCoefficient: 2.0,
            MaximumInterval:    30 * time.Second,
            MaximumAttempts:    3,
            NonRetriableErrorTypes: []string{"ValidationError"},
        },
    }
    ctx = workflow.WithActivityOptions(ctx, activityOptions)

    // 1. Criar reivindicação no Bacen (via Bridge)
    var claimResponse CreateClaimResponse
    err := workflow.ExecuteActivity(ctx, CreateClaimGRPCActivity, input).Get(ctx, &claimResponse)
    if err != nil {
        logger.Error("Falha ao criar claim no Bacen", "error", err)
        return err
    }

    logger.Info("Claim criado no Bacen", "externalID", claimResponse.ExternalID)

    // 2. Aguardar decisão ou timeout de 30 dias
    signalChannel := workflow.GetSignalChannel(ctx, "claim-decision")
    selector := workflow.NewSelector(ctx)

    var confirmed bool
    selector.AddReceive(signalChannel, func(c workflow.ReceiveChannel, more bool) {
        c.Receive(ctx, &confirmed)
        logger.Info("Decisão recebida via signal", "confirmed", confirmed)
    })

    // Timer de 30 dias
    timer := workflow.NewTimer(ctx, 30*24*time.Hour)
    selector.AddFuture(timer, func(f workflow.Future) {
        logger.Warn("Timeout de 30 dias atingido - cancelando claim")
        confirmed = false
    })

    selector.Select(ctx)

    // 3. Confirmar ou Cancelar claim
    if confirmed {
        err = workflow.ExecuteActivity(ctx, CompleteClaimGRPCActivity, input.ClaimID).Get(ctx, nil)
        if err != nil {
            return err
        }
        logger.Info("Claim confirmado", "claimID", input.ClaimID)
    } else {
        err = workflow.ExecuteActivity(ctx, CancelClaimGRPCActivity, input.ClaimID).Get(ctx, nil)
        if err != nil {
            return err
        }
        logger.Info("Claim cancelado", "claimID", input.ClaimID)
    }

    // 4. Publicar evento de conclusão
    err = workflow.ExecuteActivity(ctx, DictEventsPublishActivity, input.ClaimID, confirmed).Get(ctx, nil)
    if err != nil {
        logger.Warn("Falha ao publicar evento", "error", err)
    }

    return nil
}
```

### 5.2. Activities de Claims

**Arquivo**: `apps/orchestration-worker/activities/claims/create_activity.go` (pseudocode)

```go
package claims

import (
    "context"
    "github.com/lbpay/connector-dict/apps/shared/grpc"
)

type CreateClaimGRPCActivity struct {
    bridgeClient *grpc.BridgeClient
}

func NewCreateClaimGRPCActivity(bridgeClient *grpc.BridgeClient) *CreateClaimGRPCActivity {
    return &CreateClaimGRPCActivity{bridgeClient: bridgeClient}
}

func (a *CreateClaimGRPCActivity) Execute(ctx context.Context, input ClaimWorkflowInput) (*CreateClaimResponse, error) {
    // Chamar Bridge via gRPC
    response, err := a.bridgeClient.CreateClaim(ctx, &grpc.CreateClaimRequest{
        ClaimID:     input.ClaimID,
        EntryKey:    input.EntryKey,
        ClaimerISPB: input.ClaimerISPB,
        OwnerISPB:   input.OwnerISPB,
    })
    if err != nil {
        return nil, err
    }

    return &CreateClaimResponse{
        ExternalID: response.ExternalID,
        Status:     response.Status,
    }, nil
}
```

---

## 6. Setup de Workers

### 6.1. Worker Principal

**Arquivo**: `apps/orchestration-worker/cmd/worker/main.go` (pseudocode)

```go
package main

import (
    "log"
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
    "github.com/lbpay/connector-dict/apps/orchestration-worker/workflows/claims"
    "github.com/lbpay/connector-dict/apps/orchestration-worker/activities/claims"
    "github.com/lbpay/connector-dict/apps/shared/config"
)

func main() {
    // Carregar configuração
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Erro ao carregar config: %v", err)
    }

    // Criar Temporal client
    c, err := client.Dial(client.Options{
        HostPort:  cfg.Temporal.Host,
        Namespace: cfg.Temporal.Namespace,
    })
    if err != nil {
        log.Fatalf("Unable to create Temporal client: %v", err)
    }
    defer c.Close()

    // Criar Worker
    w := worker.New(c, cfg.Temporal.TaskQueue, worker.Options{})

    // Registrar Workflows
    w.RegisterWorkflow(claims.CreateClaimWorkflow)
    w.RegisterWorkflow(claims.MonitorStatusWorkflow)
    w.RegisterWorkflow(claims.ExpireCompletionPeriodWorkflow)
    w.RegisterWorkflow(claims.CompleteClaimWorkflow)
    w.RegisterWorkflow(claims.CancelClaimWorkflow)

    // Registrar Activities
    bridgeClient := grpc.NewBridgeClient(cfg.Bridge.GRPCAddr)
    w.RegisterActivity(claims.NewCreateClaimGRPCActivity(bridgeClient))
    w.RegisterActivity(claims.NewCompleteClaimGRPCActivity(bridgeClient))
    w.RegisterActivity(claims.NewCancelClaimGRPCActivity(bridgeClient))

    // Iniciar Worker
    log.Println("Temporal Worker iniciado no TaskQueue:", cfg.Temporal.TaskQueue)
    err = w.Run(worker.InterruptCh())
    if err != nil {
        log.Fatalf("Unable to start Worker: %v", err)
    }
}
```

---

## 7. Integração Pulsar

### 7.1. Pulsar Consumer

**Arquivo**: `apps/shared/pulsar/consumer.go` (pseudocode)

```go
package pulsar

import (
    "context"
    "encoding/json"
    "github.com/apache/pulsar-client-go/pulsar"
    "github.com/sirupsen/logrus"
)

type Consumer struct {
    client   pulsar.Client
    consumer pulsar.Consumer
    logger   *logrus.Logger
}

func NewConsumer(url, topic, subscription string, logger *logrus.Logger) (*Consumer, error) {
    client, err := pulsar.NewClient(pulsar.ClientOptions{
        URL: url,
    })
    if err != nil {
        return nil, err
    }

    consumer, err := client.Subscribe(pulsar.ConsumerOptions{
        Topic:            topic,
        SubscriptionName: subscription,
        Type:             pulsar.Shared,
    })
    if err != nil {
        return nil, err
    }

    return &Consumer{
        client:   client,
        consumer: consumer,
        logger:   logger,
    }, nil
}

func (c *Consumer) Start(ctx context.Context, handler func(context.Context, []byte) error) {
    for {
        select {
        case <-ctx.Done():
            c.consumer.Close()
            c.client.Close()
            return
        default:
            msg, err := c.consumer.Receive(ctx)
            if err != nil {
                c.logger.Errorf("Erro ao receber mensagem: %v", err)
                continue
            }

            if err := handler(ctx, msg.Payload()); err != nil {
                c.logger.Errorf("Erro ao processar mensagem: %v", err)
                c.consumer.Nack(msg)
                continue
            }

            c.consumer.Ack(msg)
        }
    }
}
```

---

## 8. Integração Redis

### 8.1. Redis Cache

**Arquivo**: `apps/shared/redis/cache.go` (pseudocode)

```go
package redis

import (
    "context"
    "time"
    "github.com/redis/go-redis/v9"
)

type Cache struct {
    client *redis.Client
}

func NewCache(url string, db int) (*Cache, error) {
    client := redis.NewClient(&redis.Options{
        Addr: url,
        DB:   db,
    })

    // Ping para verificar conexão
    if err := client.Ping(context.Background()).Err(); err != nil {
        return nil, err
    }

    return &Cache{client: client}, nil
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
    return c.client.Get(ctx, key).Result()
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *Cache) Delete(ctx context.Context, key string) error {
    return c.client.Del(ctx, key).Err()
}
```

---

## 9. Execução Local

### 9.1. Iniciar Dependências

```bash
# Iniciar Temporal, Pulsar, Redis
cd docker
docker-compose up -d

# Verificar serviços
docker-compose ps
```

### 9.2. Executar Worker

```bash
cd apps/orchestration-worker
go run cmd/worker/main.go
```

### 9.3. Executar API (dict)

```bash
cd apps/dict
go run main.go
```

### 9.4. Testar ClaimWorkflow

```bash
# Usando Temporal CLI
temporal workflow start \
  --task-queue dict-task-queue \
  --type CreateClaimWorkflow \
  --input '{"ClaimID":"claim-123","EntryKey":"12345678901","ClaimerISPB":"87654321","OwnerISPB":"12345678"}'

# Enviar signal de confirmação
temporal workflow signal \
  --workflow-id claim-123 \
  --name claim-decision \
  --input true
```

---

## 10. Execução de Testes

### 10.1. Testes de Workflows

```bash
cd apps/orchestration-worker
go test ./workflows/... -v
```

### 10.2. Testes de Activities

```bash
go test ./activities/... -v
```

---

## 11. Checklist de Implementação

### 11.1. Temporal Setup

- [ ] Temporal Server rodando (Docker Compose)
- [ ] Namespace `dict` criado
- [ ] Temporal UI acessível (http://localhost:8088)
- [ ] TaskQueue `dict-task-queue` configurado

### 11.2. ClaimWorkflow (30 dias)

- [ ] CreateClaimWorkflow implementado
- [ ] MonitorStatusWorkflow implementado
- [ ] ExpireCompletionPeriodWorkflow implementado (30 dias)
- [ ] CompleteClaimWorkflow implementado
- [ ] CancelClaimWorkflow implementado
- [ ] Activities de Claims implementadas:
  - [ ] CreateClaimGRPCActivity
  - [ ] CompleteClaimGRPCActivity
  - [ ] CancelClaimGRPCActivity
  - [ ] GetClaimGRPCActivity

### 11.3. Pulsar Integration

- [ ] Pulsar Server rodando
- [ ] Consumer configurado (rsfn-dict-req-out)
- [ ] Producer configurado (rsfn-dict-res-out)
- [ ] Handler de mensagens implementado

### 11.4. Redis Integration

- [ ] Redis Server rodando
- [ ] Cache client configurado
- [ ] CacheActivity implementada

### 11.5. Bridge Integration

- [ ] gRPC client para Bridge implementado
- [ ] Endpoints implementados:
  - [ ] CreateClaim
  - [ ] CompleteClaim
  - [ ] CancelClaim

### 11.6. Testing

- [ ] Testes unitários de workflows (>70% coverage)
- [ ] Testes de activities
- [ ] Testes de integração com Temporal

---

## Próximos Passos

1. Implementar VSYNCWorkflow (sincronização diária)
2. Implementar OTPWorkflow (validação OTP)
3. Adicionar database migrations (PostgreSQL)
4. Configurar observabilidade (OpenTelemetry)
5. Configurar CI/CD

---

**Referências**:
- [TEC-003: RSFN Connect Specification](../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [Temporal Documentation](https://docs.temporal.io/)
