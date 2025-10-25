# IMP-001: Manual de Implementação - Core DICT

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: Core DICT (API REST + Database)
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: BACKEND (AI Agent - Backend Developer)

---

## Sumário Executivo

Este manual fornece instruções passo-a-passo para configurar, implementar e executar o **Core DICT**, o módulo central que gerencia chaves PIX, contas CID, e expõe APIs REST para FrontEnd e BackOffice.

**Baseado em**:
- [DAT-001: Schema Database Core DICT](../03_Dados/DAT-001_Schema_Database_Core_DICT.md)

---

## Índice

1. [Pré-requisitos](#1-pré-requisitos)
2. [Setup do Repositório](#2-setup-do-repositório)
3. [Setup do Banco de Dados](#3-setup-do-banco-de-dados)
4. [Configuração da Aplicação](#4-configuração-da-aplicação)
5. [Execução Local](#5-execução-local)
6. [Execução de Testes](#6-execução-de-testes)
7. [Checklist de Implementação](#7-checklist-de-implementação)

---

## 1. Pré-requisitos

### 1.1. Software Necessário

| Software | Versão Mínima | Propósito |
|----------|---------------|-----------|
| **Go** | 1.22+ | Linguagem de programação |
| **PostgreSQL** | 15+ | Banco de dados principal |
| **Docker** | 20.10+ | Containerização (opcional) |
| **Docker Compose** | 2.0+ | Orquestração local (opcional) |
| **Make** | 3.81+ | Automação de build |
| **Git** | 2.30+ | Controle de versão |

### 1.2. Variáveis de Ambiente

Criar arquivo `.env` na raiz do projeto:

```bash
# Database Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=lbpay_core_dict
POSTGRES_USER=dict_app
POSTGRES_PASSWORD=secure_password_here
POSTGRES_SSLMODE=disable

# Application Configuration
APP_ENV=development
APP_PORT=8080
APP_LOG_LEVEL=debug

# ISPB do Participante (LBPay)
PARTICIPANT_ISPB=12345678

# JWT Secret (para autenticação)
JWT_SECRET=your_jwt_secret_here
JWT_EXPIRATION=24h

# Pulsar Configuration (integração com Connect)
PULSAR_URL=pulsar://localhost:6650
PULSAR_TOPIC_REQ_OUT=persistent://lb-conn/dict/rsfn-dict-req-out
PULSAR_TOPIC_RES_IN=persistent://lb-conn/dict/rsfn-dict-res-in

# Redis Configuration (cache)
REDIS_URL=redis://localhost:6379
REDIS_DB=0
```

---

## 2. Setup do Repositório

### 2.1. Criar Estrutura do Projeto

```bash
# Criar diretório do projeto
mkdir -p core-dict
cd core-dict

# Inicializar módulo Go
go mod init github.com/lbpay/core-dict

# Criar estrutura de diretórios
mkdir -p cmd/api
mkdir -p internal/{domain,application,infrastructure,ports}
mkdir -p internal/infrastructure/{database,fiber,pulsar,redis}
mkdir -p db/migrations
mkdir -p config
mkdir -p pkg/{logger,validator}
```

### 2.2. Estrutura de Diretórios

```
core-dict/
├── cmd/
│   └── api/
│       └── main.go                    # Entrypoint da aplicação
├── internal/
│   ├── domain/
│   │   ├── entry.go                   # Domain model: Entry
│   │   ├── account.go                 # Domain model: Account
│   │   ├── claim.go                   # Domain model: Claim
│   │   └── errors.go                  # Domain errors
│   ├── application/
│   │   ├── entry/
│   │   │   ├── create_entry.go        # UseCase: Create Entry
│   │   │   ├── get_entry.go           # UseCase: Get Entry
│   │   │   └── list_entries.go        # UseCase: List Entries
│   │   └── claim/
│   │       ├── create_claim.go        # UseCase: Create Claim
│   │       └── confirm_claim.go       # UseCase: Confirm Claim
│   ├── infrastructure/
│   │   ├── database/
│   │   │   ├── postgres.go            # PostgreSQL connection
│   │   │   ├── entry_repository.go    # Entry repository implementation
│   │   │   └── account_repository.go  # Account repository implementation
│   │   ├── fiber/
│   │   │   ├── server.go              # Fiber HTTP server
│   │   │   ├── routes.go              # Route definitions
│   │   │   └── handlers/
│   │   │       ├── entry_handler.go   # Entry HTTP handlers
│   │   │       └── claim_handler.go   # Claim HTTP handlers
│   │   └── pulsar/
│   │       └── producer.go            # Pulsar producer (to Connect)
│   └── ports/
│       ├── repositories.go            # Repository interfaces
│       └── logger.go                  # Logger interface
├── db/
│   └── migrations/
│       ├── 001_create_schema.sql
│       ├── 002_create_tables.sql
│       └── 003_create_indexes.sql
├── config/
│   └── config.yaml                    # Application config
├── pkg/
│   ├── logger/
│   │   └── logger.go                  # Logger implementation
│   └── validator/
│       └── validator.go               # Input validation
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yaml
├── .env                               # Environment variables
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### 2.3. Instalar Dependências

```bash
# Fiber Web Framework
go get github.com/gofiber/fiber/v2

# PostgreSQL Driver
go get github.com/lib/pq
go get gorm.io/gorm
go get gorm.io/driver/postgres

# Configuration
go get github.com/spf13/viper

# Logging
go get github.com/sirupsen/logrus

# Validation
go get github.com/go-playground/validator/v10

# UUID
go get github.com/google/uuid

# Pulsar Client
go get github.com/apache/pulsar-client-go/pulsar

# Redis Client
go get github.com/redis/go-redis/v9

# JWT
go get github.com/golang-jwt/jwt/v5

# Migration Tool (Goose)
go install github.com/pressly/goose/v3/cmd/goose@latest
```

---

## 3. Setup do Banco de Dados

### 3.1. Criar Database e Usuário

```bash
# Conectar ao PostgreSQL como superuser
psql -U postgres

# Dentro do psql:
CREATE DATABASE lbpay_core_dict;

-- Criar schemas
\c lbpay_core_dict

CREATE SCHEMA IF NOT EXISTS dict;
CREATE SCHEMA IF NOT EXISTS audit;
CREATE SCHEMA IF NOT EXISTS config;

-- Criar role de aplicação
CREATE ROLE dict_app WITH LOGIN PASSWORD 'secure_password_here';
GRANT CONNECT ON DATABASE lbpay_core_dict TO dict_app;
GRANT USAGE ON SCHEMA dict, audit, config TO dict_app;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA dict TO dict_app;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA dict TO dict_app;

-- Criar extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

\q
```

### 3.2. Criar Migrations

**Arquivo**: `db/migrations/001_create_schema.sql`

```sql
-- +goose Up
CREATE SCHEMA IF NOT EXISTS dict;
CREATE SCHEMA IF NOT EXISTS audit;
CREATE SCHEMA IF NOT EXISTS config;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- +goose Down
DROP SCHEMA IF EXISTS dict CASCADE;
DROP SCHEMA IF EXISTS audit CASCADE;
DROP SCHEMA IF EXISTS config CASCADE;
```

**Arquivo**: `db/migrations/002_create_tables.sql`

```sql
-- +goose Up

-- Tabela: Users
CREATE TABLE dict.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(50) CHECK (role IN ('ADMIN', 'OPERATOR', 'VIEWER', 'AUDITOR')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE
);

-- Tabela: Accounts
CREATE TABLE dict.accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id VARCHAR(100) UNIQUE,
    account_number VARCHAR(20) NOT NULL,
    branch_code VARCHAR(10) NOT NULL,
    account_type VARCHAR(20) NOT NULL CHECK (
        account_type IN ('CACC', 'SVGS', 'SLRY', 'TRAN')
    ),
    account_status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (
        account_status IN ('ACTIVE', 'BLOCKED', 'CLOSED', 'PENDING_CLOSURE')
    ),
    holder_document VARCHAR(14) NOT NULL,
    holder_document_type VARCHAR(10) NOT NULL CHECK (
        holder_document_type IN ('CPF', 'CNPJ')
    ),
    holder_name VARCHAR(255) NOT NULL,
    participant_ispb VARCHAR(8) NOT NULL,
    opened_at TIMESTAMP WITH TIME ZONE NOT NULL,
    closed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by UUID REFERENCES dict.users(id),
    updated_by UUID REFERENCES dict.users(id),
    CONSTRAINT unique_account UNIQUE (participant_ispb, branch_code, account_number, deleted_at)
        WHERE deleted_at IS NULL
);

-- Tabela: Claims
CREATE TABLE dict.claims (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id VARCHAR(100) UNIQUE,
    workflow_id VARCHAR(255),
    entry_id UUID REFERENCES dict.entries(id) ON DELETE RESTRICT,
    claim_type VARCHAR(50) NOT NULL CHECK (claim_type IN ('OWNERSHIP', 'PORTABILITY')),
    claimer_ispb VARCHAR(8) NOT NULL,
    claimer_account_id UUID REFERENCES dict.accounts(id),
    owner_ispb VARCHAR(8) NOT NULL,
    owner_account_id UUID NOT NULL REFERENCES dict.accounts(id),
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN' CHECK (
        status IN ('OPEN', 'WAITING_RESOLUTION', 'CONFIRMED', 'CANCELLED', 'COMPLETED', 'EXPIRED')
    ),
    completion_period_days INT NOT NULL DEFAULT 30,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    resolution_type VARCHAR(50) CHECK (
        resolution_type IN ('APPROVED', 'REJECTED', 'TIMEOUT', 'CANCELLED')
    ),
    resolution_reason TEXT,
    resolution_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by UUID REFERENCES dict.users(id),
    updated_by UUID REFERENCES dict.users(id),
    CHECK (expires_at > created_at),
    CHECK (completion_period_days > 0)
);

-- Tabela: Entries (deve ser criada após accounts e claims para foreign keys)
CREATE TABLE dict.entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id VARCHAR(100) UNIQUE,
    key_type VARCHAR(20) NOT NULL CHECK (
        key_type IN ('CPF', 'CNPJ', 'EMAIL', 'PHONE', 'EVP')
    ),
    key_value VARCHAR(255) NOT NULL,
    key_hash VARCHAR(64) NOT NULL,
    account_id UUID NOT NULL REFERENCES dict.accounts(id) ON DELETE RESTRICT,
    participant_ispb VARCHAR(8) NOT NULL,
    participant_branch VARCHAR(10),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (
        status IN ('PENDING', 'ACTIVE', 'PORTABILITY_REQUESTED',
                 'OWNERSHIP_CONFIRMED', 'DELETED', 'CLAIM_PENDING')
    ),
    ownership_type VARCHAR(20) NOT NULL CHECK (
        ownership_type IN ('NATURAL_PERSON', 'LEGAL_ENTITY')
    ),
    claim_id UUID REFERENCES dict.claims(id) ON DELETE SET NULL,
    portability_id UUID REFERENCES dict.portabilities(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    last_sync_at TIMESTAMP WITH TIME ZONE,
    sync_status VARCHAR(20) CHECK (
        sync_status IN ('SYNCED', 'PENDING_SYNC', 'SYNC_ERROR', 'NOT_SYNCED')
    ),
    sync_error_message TEXT,
    created_by UUID REFERENCES dict.users(id),
    updated_by UUID REFERENCES dict.users(id),
    CONSTRAINT unique_active_key UNIQUE (key_type, key_value, deleted_at)
        WHERE deleted_at IS NULL,
    CONSTRAINT unique_key_hash UNIQUE (key_hash)
);

-- Tabela: Portabilities
CREATE TABLE dict.portabilities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id VARCHAR(100) UNIQUE,
    workflow_id VARCHAR(255),
    entry_id UUID NOT NULL REFERENCES dict.entries(id) ON DELETE RESTRICT,
    origin_ispb VARCHAR(8) NOT NULL,
    origin_account_id UUID NOT NULL REFERENCES dict.accounts(id),
    destination_ispb VARCHAR(8) NOT NULL,
    destination_account_id UUID NOT NULL REFERENCES dict.accounts(id),
    status VARCHAR(50) NOT NULL DEFAULT 'INITIATED' CHECK (
        status IN ('INITIATED', 'PENDING_APPROVAL', 'APPROVED',
                 'REJECTED', 'COMPLETED', 'CANCELLED', 'FAILED')
    ),
    initiated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    requires_otp BOOLEAN NOT NULL DEFAULT TRUE,
    otp_validated_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES dict.users(id)
);

-- Tabela: Audit Events
CREATE TABLE audit.entry_events (
    id BIGSERIAL PRIMARY KEY,
    event_id UUID NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_subtype VARCHAR(100),
    old_values JSONB,
    new_values JSONB,
    diff JSONB,
    user_id UUID REFERENCES dict.users(id),
    ip_address INET,
    user_agent TEXT,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB
);

-- +goose Down
DROP TABLE IF EXISTS audit.entry_events CASCADE;
DROP TABLE IF EXISTS dict.portabilities CASCADE;
DROP TABLE IF EXISTS dict.entries CASCADE;
DROP TABLE IF EXISTS dict.claims CASCADE;
DROP TABLE IF EXISTS dict.accounts CASCADE;
DROP TABLE IF EXISTS dict.users CASCADE;
```

**Arquivo**: `db/migrations/003_create_indexes.sql`

```sql
-- +goose Up

-- Índices para Entries
CREATE INDEX idx_entries_key_type_value ON dict.entries (key_type, key_value) WHERE deleted_at IS NULL;
CREATE INDEX idx_entries_key_hash ON dict.entries (key_hash);
CREATE INDEX idx_entries_account_id ON dict.entries (account_id);
CREATE INDEX idx_entries_status ON dict.entries (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_entries_sync_status ON dict.entries (sync_status) WHERE sync_status != 'SYNCED';

-- Índices para Accounts
CREATE INDEX idx_accounts_holder_document ON dict.accounts (holder_document) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_participant ON dict.accounts (participant_ispb, account_status);
CREATE INDEX idx_accounts_holder_name_trgm ON dict.accounts USING gin (holder_name gin_trgm_ops);

-- Índices para Claims
CREATE INDEX idx_claims_status ON dict.claims (status) WHERE status IN ('OPEN', 'WAITING_RESOLUTION');
CREATE INDEX idx_claims_expires_at ON dict.claims (expires_at) WHERE status = 'OPEN';
CREATE INDEX idx_claims_workflow_id ON dict.claims (workflow_id);

-- Índices para Portabilities
CREATE INDEX idx_portabilities_status ON dict.portabilities (status);
CREATE INDEX idx_portabilities_workflow_id ON dict.portabilities (workflow_id);

-- Índices para Audit
CREATE INDEX idx_entry_events_occurred_at ON audit.entry_events (occurred_at DESC);
CREATE INDEX idx_entry_events_entity ON audit.entry_events (entity_type, entity_id);
CREATE INDEX idx_entry_events_user ON audit.entry_events (user_id, occurred_at DESC);

-- +goose Down
DROP INDEX IF EXISTS dict.idx_entries_key_type_value;
DROP INDEX IF EXISTS dict.idx_entries_key_hash;
DROP INDEX IF EXISTS dict.idx_entries_account_id;
DROP INDEX IF EXISTS dict.idx_entries_status;
DROP INDEX IF EXISTS dict.idx_entries_sync_status;
DROP INDEX IF EXISTS dict.idx_accounts_holder_document;
DROP INDEX IF EXISTS dict.idx_accounts_participant;
DROP INDEX IF EXISTS dict.idx_accounts_holder_name_trgm;
DROP INDEX IF EXISTS dict.idx_claims_status;
DROP INDEX IF EXISTS dict.idx_claims_expires_at;
DROP INDEX IF EXISTS dict.idx_claims_workflow_id;
DROP INDEX IF EXISTS dict.idx_portabilities_status;
DROP INDEX IF EXISTS dict.idx_portabilities_workflow_id;
DROP INDEX IF EXISTS audit.idx_entry_events_occurred_at;
DROP INDEX IF EXISTS audit.idx_entry_events_entity;
DROP INDEX IF EXISTS audit.idx_entry_events_user;
```

### 3.3. Executar Migrations

```bash
# Definir string de conexão
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="host=localhost port=5432 user=dict_app password=secure_password_here dbname=lbpay_core_dict sslmode=disable"

# Executar migrations
cd db/migrations
goose up

# Verificar status
goose status

# Rollback (se necessário)
goose down
```

---

## 4. Configuração da Aplicação

### 4.1. Arquivo de Configuração

**Arquivo**: `config/config.yaml`

```yaml
app:
  name: "Core DICT API"
  version: "1.0.0"
  env: "development"
  port: 8080

database:
  host: "localhost"
  port: 5432
  database: "lbpay_core_dict"
  user: "dict_app"
  password: "secure_password_here"
  sslmode: "disable"
  max_connections: 100
  max_idle_connections: 10
  connection_max_lifetime: 3600

participant:
  ispb: "12345678"
  name: "LBPay"

pulsar:
  url: "pulsar://localhost:6650"
  topics:
    request_out: "persistent://lb-conn/dict/rsfn-dict-req-out"
    response_in: "persistent://lb-conn/dict/rsfn-dict-res-in"
  consumer:
    subscription: "core-dict-sub"

redis:
  url: "redis://localhost:6379"
  db: 0
  pool_size: 10

jwt:
  secret: "your_jwt_secret_here"
  expiration: "24h"

logging:
  level: "debug"
  format: "json"
```

### 4.2. Setup do Fiber Server

**Arquivo**: `cmd/api/main.go` (pseudocode)

```go
package main

import (
    "log"
    "github.com/gofiber/fiber/v2"
    "github.com/lbpay/core-dict/internal/infrastructure/fiber"
    "github.com/lbpay/core-dict/internal/infrastructure/database"
    "github.com/spf13/viper"
)

func main() {
    // Carregar configuração
    viper.SetConfigFile("config/config.yaml")
    if err := viper.ReadInConfig(); err != nil {
        log.Fatalf("Erro ao ler configuração: %v", err)
    }

    // Conectar ao PostgreSQL
    db, err := database.NewPostgresConnection()
    if err != nil {
        log.Fatalf("Erro ao conectar ao banco: %v", err)
    }
    defer db.Close()

    // Criar Fiber app
    app := fiber.New(fiber.Config{
        AppName: viper.GetString("app.name"),
    })

    // Setup de rotas
    fiberServer := fiber.NewServer(app, db)
    fiberServer.SetupRoutes()

    // Iniciar servidor
    port := viper.GetString("app.port")
    log.Printf("Servidor iniciado na porta %s", port)
    if err := app.Listen(":" + port); err != nil {
        log.Fatalf("Erro ao iniciar servidor: %v", err)
    }
}
```

### 4.3. Setup de Rotas

**Arquivo**: `internal/infrastructure/fiber/routes.go` (pseudocode)

```go
package fiber

import (
    "github.com/gofiber/fiber/v2"
    "github.com/lbpay/core-dict/internal/infrastructure/fiber/handlers"
)

func (s *Server) SetupRoutes() {
    api := s.app.Group("/api/v1")

    // Health check
    api.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "ok"})
    })

    // Entries
    entries := api.Group("/entries")
    entries.Post("/", s.entryHandler.CreateEntry)
    entries.Get("/:key", s.entryHandler.GetEntry)
    entries.Get("/", s.entryHandler.ListEntries)
    entries.Delete("/:key", s.entryHandler.DeleteEntry)

    // Claims
    claims := api.Group("/claims")
    claims.Post("/", s.claimHandler.CreateClaim)
    claims.Get("/:id", s.claimHandler.GetClaim)
    claims.Put("/:id/confirm", s.claimHandler.ConfirmClaim)
    claims.Put("/:id/cancel", s.claimHandler.CancelClaim)

    // Accounts
    accounts := api.Group("/accounts")
    accounts.Post("/", s.accountHandler.CreateAccount)
    accounts.Get("/:id", s.accountHandler.GetAccount)
}
```

---

## 5. Execução Local

### 5.1. Iniciar Dependências (Docker Compose)

**Arquivo**: `docker/docker-compose.yaml`

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: lbpay_core_dict
      POSTGRES_USER: dict_app
      POSTGRES_PASSWORD: secure_password_here
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  pulsar:
    image: apachepulsar/pulsar:latest
    command: bin/pulsar standalone
    ports:
      - "6650:6650"
      - "8080:8080"

volumes:
  postgres_data:
```

```bash
# Iniciar dependências
cd docker
docker-compose up -d

# Verificar se serviços estão rodando
docker-compose ps
```

### 5.2. Executar Aplicação

```bash
# Na raiz do projeto
go run cmd/api/main.go

# Ou com hot-reload (usando air)
go install github.com/cosmtrek/air@latest
air
```

### 5.3. Testar Endpoints

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Criar entrada (exemplo)
curl -X POST http://localhost:8080/api/v1/entries \
  -H "Content-Type: application/json" \
  -d '{
    "key_type": "CPF",
    "key_value": "12345678901",
    "account_id": "550e8400-e29b-41d4-a716-446655440000",
    "participant_ispb": "12345678"
  }'

# Buscar entrada
curl http://localhost:8080/api/v1/entries/12345678901
```

---

## 6. Execução de Testes

### 6.1. Testes Unitários

```bash
# Executar todos os testes
go test ./...

# Executar testes com coverage
go test -cover ./...

# Gerar relatório de coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 6.2. Testes de Integração

```bash
# Executar apenas testes de integração (com tag)
go test -tags=integration ./...

# Exemplo de teste de integração
go test -v ./internal/infrastructure/database -run TestEntryRepository
```

### 6.3. Exemplo de Teste Unitário

**Arquivo**: `internal/domain/entry_test.go` (pseudocode)

```go
package domain_test

import (
    "testing"
    "github.com/lbpay/core-dict/internal/domain"
)

func TestEntryValidation(t *testing.T) {
    tests := []struct {
        name    string
        entry   *domain.Entry
        wantErr bool
    }{
        {
            name: "Valid CPF entry",
            entry: &domain.Entry{
                KeyType:  "CPF",
                KeyValue: "12345678901",
                AccountID: "550e8400-e29b-41d4-a716-446655440000",
            },
            wantErr: false,
        },
        {
            name: "Invalid CPF length",
            entry: &domain.Entry{
                KeyType:  "CPF",
                KeyValue: "123", // CPF deve ter 11 dígitos
                AccountID: "550e8400-e29b-41d4-a716-446655440000",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.entry.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Entry.Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## 7. Checklist de Implementação

### 7.1. Database Setup

- [ ] PostgreSQL 15+ instalado
- [ ] Database `lbpay_core_dict` criada
- [ ] Schemas `dict`, `audit`, `config` criados
- [ ] Extensions (`uuid-ossp`, `pg_trgm`, `pgcrypto`) habilitadas
- [ ] Role `dict_app` criada com permissões adequadas
- [ ] Migrations executadas com sucesso
- [ ] Índices criados corretamente

### 7.2. Application Setup

- [ ] Go 1.22+ instalado
- [ ] Dependências instaladas (`go mod tidy`)
- [ ] Arquivo `.env` configurado
- [ ] Arquivo `config/config.yaml` configurado
- [ ] Estrutura de diretórios criada
- [ ] Fiber server configurado
- [ ] Rotas definidas

### 7.3. Domain Layer

- [ ] Models criados: `Entry`, `Account`, `Claim`, `Portability`
- [ ] Validações de negócio implementadas
- [ ] Domain errors definidos

### 7.4. Application Layer

- [ ] Use cases implementados:
  - [ ] `CreateEntryUseCase`
  - [ ] `GetEntryUseCase`
  - [ ] `ListEntriesUseCase`
  - [ ] `DeleteEntryUseCase`
  - [ ] `CreateClaimUseCase`
  - [ ] `ConfirmClaimUseCase`

### 7.5. Infrastructure Layer

- [ ] PostgreSQL connection configurada
- [ ] Repositories implementados:
  - [ ] `EntryRepository`
  - [ ] `AccountRepository`
  - [ ] `ClaimRepository`
- [ ] Fiber handlers implementados
- [ ] Pulsar producer configurado
- [ ] Redis client configurado

### 7.6. Testing

- [ ] Testes unitários escritos (>70% coverage)
- [ ] Testes de integração escritos
- [ ] Testes de API (Postman/Insomnia collection)

### 7.7. Deployment

- [ ] Dockerfile criado
- [ ] Docker Compose configurado
- [ ] Variáveis de ambiente documentadas
- [ ] README atualizado

---

## Próximos Passos

Após completar este manual:

1. Implementar integração com RSFN Connect (via Pulsar)
2. Adicionar autenticação JWT
3. Implementar observabilidade (OpenTelemetry)
4. Adicionar rate limiting
5. Configurar CI/CD

---

**Referências**:
- [DAT-001: Schema Database Core DICT](../03_Dados/DAT-001_Schema_Database_Core_DICT.md)
- [TEC-001: Core DICT Specification](../11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md)
