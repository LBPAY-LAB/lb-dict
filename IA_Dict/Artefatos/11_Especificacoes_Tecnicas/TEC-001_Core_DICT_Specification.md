# TEC-001: Especificação Técnica - Core DICT

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: Core DICT (Domain Service)
**Versão**: 1.0
**Data**: 2025-10-24
**Autor**: ARCHITECT (AI Agent - Technical Architect)
**Revisor**: [Aguardando]
**Aprovador**: Head de Arquitetura (Thiago Lima), CTO (José Luís Silva)

---

## Controle de Versão

| Versão | Data | Autor | Descrição das Mudanças |
|--------|------|-------|------------------------|
| 1.0 | 2025-10-24 | ARCHITECT | Versão inicial - Especificação técnica completa do Core DICT |

---

## Sumário Executivo

### Visão Geral

O **Core DICT** é o serviço central do sistema DICT da LBPay, responsável por:
- Implementar toda a **lógica de domínio** (regras de negócio PIX)
- Gerenciar **entidades de domínio** (Chaves PIX, Claims, Portabilidades)
- Expor **APIs gRPC** para clientes (LB-Connect)
- Persistir dados no **PostgreSQL**
- Publicar **eventos de domínio** no **Apache Pulsar**
- Validar **requisitos regulatórios** do Bacen

### Arquitetura

**Clean Architecture** (Domain-Driven Design):
```
┌─────────────────────────────────────────────────────────┐
│                   Interface Layer                        │
│  (gRPC Server, Interceptors, Validators)                │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                   Use Case Layer                         │
│  (RegisterKeyUseCase, DeleteKeyUseCase, etc.)           │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                   Domain Layer                           │
│  (Entities, Value Objects, Domain Services, Events)     │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                 Infrastructure Layer                     │
│  (PostgreSQL, Redis, Pulsar, External APIs)             │
└─────────────────────────────────────────────────────────┘
```

### Stack Tecnológica

| Componente | Tecnologia | Versão | Justificativa |
|------------|------------|--------|---------------|
| **Linguagem** | Go (Golang) | 1.22+ | Performance, concorrência, type-safe |
| **APIs** | gRPC + Protocol Buffers | v1.62+ | [ADR-003](../02_Arquitetura/ADR-003_Protocol_gRPC.md) |
| **Database** | PostgreSQL | 16+ | [ADR-005](../02_Arquitetura/ADR-005_Database_PostgreSQL.md) |
| **Cache** | Redis | 7+ | [ADR-004](../02_Arquitetura/ADR-004_Cache_Redis.md) |
| **Message Broker** | Apache Pulsar | 3.0+ | [ADR-001](../02_Arquitetura/ADR-001_Message_Broker_Apache_Pulsar.md) |
| **Observability** | Prometheus, Grafana, OpenTelemetry | Latest | Metrics, traces, logs |

---

## Índice

1. [Arquitetura de Software](#1-arquitetura-de-software)
2. [Domain Layer](#2-domain-layer)
3. [Use Case Layer](#3-use-case-layer)
4. [Interface Layer (gRPC APIs)](#4-interface-layer-grpc-apis)
5. [Infrastructure Layer](#5-infrastructure-layer)
6. [Database Schema](#6-database-schema)
7. [Eventos de Domínio](#7-eventos-de-domínio)
8. [Validações e Regras de Negócio](#8-validações-e-regras-de-negócio)
9. [Error Handling](#9-error-handling)
10. [Observabilidade](#10-observabilidade)
11. [Testing Strategy](#11-testing-strategy)
12. [Deployment](#12-deployment)
13. [Rastreabilidade](#13-rastreabilidade)

---

## 1. Arquitetura de Software

### 1.1 Clean Architecture (DDD)

**Estrutura de Diretórios**:

```
core-dict/
├── api/
│   └── proto/
│       └── dict/
│           └── v1/
│               ├── dict.proto              # Service definition
│               ├── dict.pb.go              # Generated
│               └── dict_grpc.pb.go         # Generated
│
├── cmd/
│   └── server/
│       └── main.go                         # Application entrypoint
│
├── internal/
│   ├── domain/                             # Domain Layer
│   │   ├── entity/
│   │   │   ├── key.go                      # DictKey entity
│   │   │   ├── claim.go                    # Claim entity
│   │   │   └── portability.go              # Portability entity
│   │   ├── valueobject/
│   │   │   ├── key_type.go                 # KeyType value object
│   │   │   ├── key_status.go               # KeyStatus value object
│   │   │   └── account.go                  # Account value object
│   │   ├── event/
│   │   │   ├── key_registered.go           # Domain event
│   │   │   ├── claim_received.go           # Domain event
│   │   │   └── event.go                    # Base event interface
│   │   ├── repository/
│   │   │   ├── key_repository.go           # Repository interface
│   │   │   ├── claim_repository.go         # Repository interface
│   │   │   └── portability_repository.go   # Repository interface
│   │   └── service/
│   │       ├── key_validator.go            # Domain service
│   │       └── ownership_validator.go      # Domain service
│   │
│   ├── usecase/                            # Use Case Layer
│   │   ├── register_key.go                 # RegisterKeyUseCase
│   │   ├── delete_key.go                   # DeleteKeyUseCase
│   │   ├── get_entry.go                    # GetEntryUseCase
│   │   ├── create_claim.go                 # CreateClaimUseCase
│   │   └── respond_claim.go                # RespondClaimUseCase
│   │
│   ├── interface/                          # Interface Layer
│   │   ├── grpc/
│   │   │   ├── server.go                   # gRPC server
│   │   │   ├── dict_handler.go             # DictService handler
│   │   │   ├── claim_handler.go            # ClaimService handler
│   │   │   └── interceptor/
│   │   │       ├── auth.go                 # JWT auth interceptor
│   │   │       ├── logging.go              # Logging interceptor
│   │   │       ├── metrics.go              # Metrics interceptor
│   │   │       └── ratelimit.go            # Rate limiting interceptor
│   │   └── dto/
│   │       └── mapper.go                   # Proto ↔ Domain mappers
│   │
│   └── infrastructure/                     # Infrastructure Layer
│       ├── persistence/
│       │   ├── postgres/
│       │   │   ├── key_repository_impl.go  # PostgreSQL implementation
│       │   │   ├── claim_repository_impl.go
│       │   │   └── transaction.go          # Transaction manager
│       │   └── redis/
│       │       ├── cache.go                # Redis cache
│       │       └── ratelimiter.go          # Redis rate limiter
│       ├── messaging/
│       │   └── pulsar/
│       │       ├── producer.go             # Pulsar event producer
│       │       └── consumer.go             # Pulsar event consumer
│       ├── external/
│       │   └── bacen/
│       │       └── client.go               # Bacen API client (mock)
│       └── config/
│           └── config.go                   # Configuration loader
│
├── db/
│   ├── migrations/                         # Database migrations (golang-migrate)
│   │   ├── 000001_create_dict_keys_table.up.sql
│   │   └── 000001_create_dict_keys_table.down.sql
│   └── queries/                            # SQL queries (SQLC)
│       ├── keys.sql
│       └── claims.sql
│
├── test/
│   ├── unit/                               # Unit tests
│   ├── integration/                        # Integration tests
│   └── e2e/                                # End-to-end tests
│
├── Makefile                                # Build automation
├── Dockerfile                              # Container image
├── go.mod                                  # Go dependencies
└── README.md                               # Project documentation
```

### 1.2 Dependency Flow

**Regra de Dependência**:
- **Domain Layer**: Não depende de nada (puro Go)
- **Use Case Layer**: Depende apenas do Domain Layer
- **Interface Layer**: Depende de Use Case e Domain
- **Infrastructure Layer**: Depende de Domain (implementa interfaces)

**Dependency Injection**:
```go
// cmd/server/main.go
func main() {
    // Infrastructure dependencies
    db := postgres.NewConnection()
    cache := redis.NewClient()
    eventBus := pulsar.NewProducer()

    // Repositories (infrastructure → domain interface)
    keyRepo := postgres.NewKeyRepository(db)
    claimRepo := postgres.NewClaimRepository(db)

    // Use cases (use case → domain interface)
    registerKeyUC := usecase.NewRegisterKeyUseCase(keyRepo, eventBus)
    deleteKeyUC := usecase.NewDeleteKeyUseCase(keyRepo, eventBus)

    // gRPC handlers (interface → use case)
    dictHandler := grpc.NewDictHandler(registerKeyUC, deleteKeyUC)

    // Start server
    server := grpc.NewServer(dictHandler)
    server.Start(":8080")
}
```

---

## 2. Domain Layer

### 2.1 Entidades (Entities)

#### Entity: `DictKey`

**Arquivo**: `internal/domain/entity/key.go`

```go
package entity

import (
    "errors"
    "time"

    "github.com/google/uuid"
    "github.com/lbpay/core-dict/internal/domain/event"
    "github.com/lbpay/core-dict/internal/domain/valueobject"
)

// DictKey representa uma chave PIX cadastrada no DICT
type DictKey struct {
    ID              uuid.UUID
    Type            valueobject.KeyType
    Value           string
    Status          valueobject.KeyStatus
    AccountID       uuid.UUID
    Account         valueobject.Account
    Owner           valueobject.Owner
    BacenEntryID    string  // ID retornado pelo Bacen
    Metadata        map[string]interface{}
    CreatedAt       time.Time
    UpdatedAt       time.Time
    DeletedAt       *time.Time  // Soft delete

    // Domain events (uncommitted)
    domainEvents []event.DomainEvent
}

// NewDictKey cria nova chave PIX (factory method)
func NewDictKey(
    keyType valueobject.KeyType,
    keyValue string,
    accountID uuid.UUID,
    account valueobject.Account,
    owner valueobject.Owner,
) (*DictKey, error) {
    // Validações de negócio
    if err := ValidateKeyFormat(keyType, keyValue); err != nil {
        return nil, err
    }

    key := &DictKey{
        ID:        uuid.New(),
        Type:      keyType,
        Value:     keyValue,
        Status:    valueobject.KeyStatusPending,
        AccountID: accountID,
        Account:   account,
        Owner:     owner,
        Metadata:  make(map[string]interface{}),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    // Adicionar evento de domínio
    key.AddDomainEvent(event.NewKeyRegisterRequested(key))

    return key, nil
}

// Activate marca chave como ativa (após confirmação Bacen)
func (k *DictKey) Activate(bacenEntryID string) error {
    if k.Status != valueobject.KeyStatusPending {
        return errors.New("key must be in PENDING status to activate")
    }

    k.Status = valueobject.KeyStatusActive
    k.BacenEntryID = bacenEntryID
    k.UpdatedAt = time.Now()

    // Evento de domínio
    k.AddDomainEvent(event.NewKeyRegistered(k))

    return nil
}

// Fail marca chave como falhada
func (k *DictKey) Fail(reason string) {
    k.Status = valueobject.KeyStatusFailed
    k.Metadata["failure_reason"] = reason
    k.UpdatedAt = time.Now()

    // Evento de domínio
    k.AddDomainEvent(event.NewKeyRegistrationFailed(k, reason))
}

// Delete realiza soft delete
func (k *DictKey) Delete() error {
    if k.Status != valueobject.KeyStatusActive {
        return errors.New("only active keys can be deleted")
    }

    now := time.Now()
    k.Status = valueobject.KeyStatusDeleted
    k.DeletedAt = &now
    k.UpdatedAt = now

    // Evento de domínio
    k.AddDomainEvent(event.NewKeyDeleted(k))

    return nil
}

// Domain Events Management
func (k *DictKey) AddDomainEvent(evt event.DomainEvent) {
    k.domainEvents = append(k.domainEvents, evt)
}

func (k *DictKey) DomainEvents() []event.DomainEvent {
    return k.domainEvents
}

func (k *DictKey) ClearDomainEvents() {
    k.domainEvents = nil
}
```

#### Entity: `Claim`

**Arquivo**: `internal/domain/entity/claim.go`

```go
package entity

import (
    "errors"
    "time"

    "github.com/google/uuid"
    "github.com/lbpay/core-dict/internal/domain/event"
    "github.com/lbpay/core-dict/internal/domain/valueobject"
)

type ClaimType string

const (
    ClaimTypeIncoming ClaimType = "INCOMING"  // Recebemos claim de outro PSP
    ClaimTypeOutgoing ClaimType = "OUTGOING"  // Enviamos claim para outro PSP
)

type ClaimStatus string

const (
    ClaimStatusPending      ClaimStatus = "PENDING"
    ClaimStatusConfirmed    ClaimStatus = "CONFIRMED"
    ClaimStatusCancelled    ClaimStatus = "CANCELLED"
    ClaimStatusAutoConfirmed ClaimStatus = "AUTO_CONFIRMED"
)

type Claim struct {
    ID              uuid.UUID
    KeyID           uuid.UUID
    Type            ClaimType
    Status          ClaimStatus
    ClaimerISPB     string  // ISPB do PSP que reivindica
    ClaimedISPB     string  // ISPB do PSP reivindicado
    BacenClaimID    string
    RequestedAt     time.Time
    DeadlineAt      time.Time  // 7 dias corridos
    ResolvedAt      *time.Time
    ResolutionReason string
    Metadata        map[string]interface{}
    CreatedAt       time.Time
    UpdatedAt       time.Time

    domainEvents []event.DomainEvent
}

// NewIncomingClaim cria claim recebida de outro PSP
func NewIncomingClaim(
    keyID uuid.UUID,
    claimerISPB string,
    claimedISPB string,
    bacenClaimID string,
) *Claim {
    now := time.Now()
    deadline := now.Add(7 * 24 * time.Hour)  // 7 dias corridos

    claim := &Claim{
        ID:           uuid.New(),
        KeyID:        keyID,
        Type:         ClaimTypeIncoming,
        Status:       ClaimStatusPending,
        ClaimerISPB:  claimerISPB,
        ClaimedISPB:  claimedISPB,
        BacenClaimID: bacenClaimID,
        RequestedAt:  now,
        DeadlineAt:   deadline,
        Metadata:     make(map[string]interface{}),
        CreatedAt:    now,
        UpdatedAt:    now,
    }

    // Evento de domínio
    claim.AddDomainEvent(event.NewClaimReceived(claim))

    return claim
}

// Accept aceita claim (usuário confirmou)
func (c *Claim) Accept() error {
    if c.Status != ClaimStatusPending {
        return errors.New("claim must be pending to accept")
    }

    now := time.Now()
    c.Status = ClaimStatusConfirmed
    c.ResolvedAt = &now
    c.UpdatedAt = now

    // Evento de domínio
    c.AddDomainEvent(event.NewClaimAccepted(c))

    return nil
}

// Reject rejeita claim (usuário recusou)
func (c *Claim) Reject(reason string) error {
    if c.Status != ClaimStatusPending {
        return errors.New("claim must be pending to reject")
    }

    now := time.Now()
    c.Status = ClaimStatusCancelled
    c.ResolutionReason = reason
    c.ResolvedAt = &now
    c.UpdatedAt = now

    // Evento de domínio
    c.AddDomainEvent(event.NewClaimRejected(c, reason))

    return nil
}

// AutoConfirm auto-confirma claim (7 dias sem resposta)
func (c *Claim) AutoConfirm() error {
    if c.Status != ClaimStatusPending {
        return errors.New("claim must be pending to auto-confirm")
    }

    if time.Now().Before(c.DeadlineAt) {
        return errors.New("deadline not reached yet")
    }

    now := time.Now()
    c.Status = ClaimStatusAutoConfirmed
    c.ResolutionReason = "No response within 7 days"
    c.ResolvedAt = &now
    c.UpdatedAt = now

    // Evento de domínio
    c.AddDomainEvent(event.NewClaimAutoConfirmed(c))

    return nil
}

func (c *Claim) AddDomainEvent(evt event.DomainEvent) {
    c.domainEvents = append(c.domainEvents, evt)
}

func (c *Claim) DomainEvents() []event.DomainEvent {
    return c.domainEvents
}

func (c *Claim) ClearDomainEvents() {
    c.domainEvents = nil
}
```

### 2.2 Value Objects

#### Value Object: `KeyType`

**Arquivo**: `internal/domain/valueobject/key_type.go`

```go
package valueobject

import "errors"

type KeyType string

const (
    KeyTypeCPF   KeyType = "CPF"
    KeyTypeCNPJ  KeyType = "CNPJ"
    KeyTypeEmail KeyType = "EMAIL"
    KeyTypePhone KeyType = "PHONE"
    KeyTypeEVP   KeyType = "EVP"
)

var validKeyTypes = map[KeyType]bool{
    KeyTypeCPF:   true,
    KeyTypeCNPJ:  true,
    KeyTypeEmail: true,
    KeyTypePhone: true,
    KeyTypeEVP:   true,
}

func NewKeyType(s string) (KeyType, error) {
    kt := KeyType(s)
    if !validKeyTypes[kt] {
        return "", errors.New("invalid key type")
    }
    return kt, nil
}

func (kt KeyType) String() string {
    return string(kt)
}

func (kt KeyType) IsValid() bool {
    return validKeyTypes[kt]
}
```

#### Value Object: `Account`

**Arquivo**: `internal/domain/valueobject/account.go`

```go
package valueobject

type AccountType string

const (
    AccountTypeCACC AccountType = "CACC"  // Checking account
    AccountTypeSVGS AccountType = "SVGS"  // Savings account
    AccountTypeSLRY AccountType = "SLRY"  // Salary account
)

type Account struct {
    ISPB          string
    Branch        string
    AccountNumber string
    AccountType   AccountType
}

func NewAccount(ispb, branch, accountNumber string, accountType AccountType) Account {
    return Account{
        ISPB:          ispb,
        Branch:        branch,
        AccountNumber: accountNumber,
        AccountType:   accountType,
    }
}
```

### 2.3 Domain Services

#### Domain Service: `KeyValidator`

**Arquivo**: `internal/domain/service/key_validator.go`

```go
package service

import (
    "errors"
    "regexp"
    "strconv"

    "github.com/lbpay/core-dict/internal/domain/valueobject"
)

type KeyValidator struct{}

func NewKeyValidator() *KeyValidator {
    return &KeyValidator{}
}

// ValidateFormat valida formato da chave conforme tipo
func (v *KeyValidator) ValidateFormat(keyType valueobject.KeyType, keyValue string) error {
    switch keyType {
    case valueobject.KeyTypeCPF:
        return v.validateCPFFormat(keyValue)
    case valueobject.KeyTypeCNPJ:
        return v.validateCNPJFormat(keyValue)
    case valueobject.KeyTypeEmail:
        return v.validateEmailFormat(keyValue)
    case valueobject.KeyTypePhone:
        return v.validatePhoneFormat(keyValue)
    case valueobject.KeyTypeEVP:
        return v.validateEVPFormat(keyValue)
    default:
        return errors.New("invalid key type")
    }
}

func (v *KeyValidator) validateCPFFormat(cpf string) error {
    // 1. Validar comprimento
    if len(cpf) != 11 {
        return errors.New("CPF must have 11 digits")
    }

    // 2. Validar apenas números
    if !regexp.MustCompile(`^\d{11}$`).MatchString(cpf) {
        return errors.New("CPF must contain only digits")
    }

    // 3. Rejeitar CPFs conhecidos inválidos
    if cpf == "00000000000" || cpf == "11111111111" || cpf == "22222222222" {
        return errors.New("invalid CPF (known invalid pattern)")
    }

    // 4. Validar dígitos verificadores
    if !v.validateCPFCheckDigits(cpf) {
        return errors.New("invalid CPF check digits")
    }

    return nil
}

func (v *KeyValidator) validateCPFCheckDigits(cpf string) bool {
    // Algoritmo de validação CPF (dígitos verificadores)
    // https://www.geradorcpf.com/algoritmo_do_cpf.htm

    // Primeiro dígito verificador
    sum := 0
    for i := 0; i < 9; i++ {
        digit, _ := strconv.Atoi(string(cpf[i]))
        sum += digit * (10 - i)
    }
    firstDigit := (sum * 10) % 11
    if firstDigit == 10 {
        firstDigit = 0
    }
    if firstDigit != int(cpf[9]-'0') {
        return false
    }

    // Segundo dígito verificador
    sum = 0
    for i := 0; i < 10; i++ {
        digit, _ := strconv.Atoi(string(cpf[i]))
        sum += digit * (11 - i)
    }
    secondDigit := (sum * 10) % 11
    if secondDigit == 10 {
        secondDigit = 0
    }
    if secondDigit != int(cpf[10]-'0') {
        return false
    }

    return true
}

func (v *KeyValidator) validateCNPJFormat(cnpj string) error {
    if len(cnpj) != 14 {
        return errors.New("CNPJ must have 14 digits")
    }

    if !regexp.MustCompile(`^\d{14}$`).MatchString(cnpj) {
        return errors.New("CNPJ must contain only digits")
    }

    // TODO: Validar dígitos verificadores CNPJ (similar ao CPF)

    return nil
}

func (v *KeyValidator) validateEmailFormat(email string) error {
    // RFC 5322 simplified
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(email) {
        return errors.New("invalid email format")
    }
    return nil
}

func (v *KeyValidator) validatePhoneFormat(phone string) error {
    // E.164 format: +5511999998888
    phoneRegex := regexp.MustCompile(`^\+55[1-9]{2}9?[0-9]{8}$`)
    if !phoneRegex.MatchString(phone) {
        return errors.New("invalid phone format (must be E.164: +5511999998888)")
    }
    return nil
}

func (v *KeyValidator) validateEVPFormat(evp string) error {
    // UUID v4
    uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
    if !uuidRegex.MatchString(evp) {
        return errors.New("invalid EVP format (must be UUID v4)")
    }
    return nil
}
```

### 2.4 Domain Events

#### Base Event

**Arquivo**: `internal/domain/event/event.go`

```go
package event

import (
    "time"

    "github.com/google/uuid"
)

type DomainEvent interface {
    EventID() string
    EventType() string
    AggregateID() string
    AggregateType() string
    OccurredAt() time.Time
    Payload() interface{}
}

type BaseDomainEvent struct {
    eventID       string
    eventType     string
    aggregateID   string
    aggregateType string
    occurredAt    time.Time
    payload       interface{}
}

func NewBaseDomainEvent(eventType, aggregateID, aggregateType string, payload interface{}) BaseDomainEvent {
    return BaseDomainEvent{
        eventID:       uuid.New().String(),
        eventType:     eventType,
        aggregateID:   aggregateID,
        aggregateType: aggregateType,
        occurredAt:    time.Now(),
        payload:       payload,
    }
}

func (e BaseDomainEvent) EventID() string       { return e.eventID }
func (e BaseDomainEvent) EventType() string     { return e.eventType }
func (e BaseDomainEvent) AggregateID() string   { return e.aggregateID }
func (e BaseDomainEvent) AggregateType() string { return e.aggregateType }
func (e BaseDomainEvent) OccurredAt() time.Time { return e.occurredAt }
func (e BaseDomainEvent) Payload() interface{}  { return e.payload }
```

#### Event: `KeyRegisterRequested`

**Arquivo**: `internal/domain/event/key_registered.go`

```go
package event

import "github.com/lbpay/core-dict/internal/domain/entity"

const (
    EventTypeKeyRegisterRequested   = "KeyRegisterRequested"
    EventTypeKeyRegistered          = "KeyRegistered"
    EventTypeKeyRegistrationFailed  = "KeyRegistrationFailed"
    EventTypeKeyDeleted             = "KeyDeleted"
)

type KeyRegisterRequestedPayload struct {
    KeyID     string
    KeyType   string
    KeyValue  string
    AccountID string
    ISPB      string
}

func NewKeyRegisterRequested(key *entity.DictKey) DomainEvent {
    payload := KeyRegisterRequestedPayload{
        KeyID:     key.ID.String(),
        KeyType:   string(key.Type),
        KeyValue:  key.Value,
        AccountID: key.AccountID.String(),
        ISPB:      key.Account.ISPB,
    }

    return NewBaseDomainEvent(
        EventTypeKeyRegisterRequested,
        key.ID.String(),
        "DictKey",
        payload,
    )
}

func NewKeyRegistered(key *entity.DictKey) DomainEvent {
    payload := KeyRegisterRequestedPayload{
        KeyID:     key.ID.String(),
        KeyType:   string(key.Type),
        KeyValue:  key.Value,
        AccountID: key.AccountID.String(),
        ISPB:      key.Account.ISPB,
    }

    return NewBaseDomainEvent(
        EventTypeKeyRegistered,
        key.ID.String(),
        "DictKey",
        payload,
    )
}
```

### 2.5 Repository Interfaces (Ports)

**Arquivo**: `internal/domain/repository/key_repository.go`

```go
package repository

import (
    "context"

    "github.com/google/uuid"
    "github.com/lbpay/core-dict/internal/domain/entity"
    "github.com/lbpay/core-dict/internal/domain/valueobject"
)

// KeyRepository interface (port para infrastructure)
type KeyRepository interface {
    // Create cria nova chave
    Create(ctx context.Context, key *entity.DictKey) error

    // FindByID busca chave por ID
    FindByID(ctx context.Context, keyID uuid.UUID) (*entity.DictKey, error)

    // FindByValue busca chave por valor (ex: CPF "12345678901")
    FindByValue(ctx context.Context, keyValue string) (*entity.DictKey, error)

    // FindByAccount lista chaves de uma conta
    FindByAccount(ctx context.Context, accountID uuid.UUID) ([]*entity.DictKey, error)

    // Update atualiza chave existente
    Update(ctx context.Context, key *entity.DictKey) error

    // CountByOwnerAndType conta chaves por titular e tipo (para limites)
    CountByOwnerAndType(ctx context.Context, ownerTaxID string, keyType valueobject.KeyType) (int, error)

    // Transaction management
    WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
```

---

## 3. Use Case Layer

### 3.1 Use Case: `RegisterKeyUseCase`

**Arquivo**: `internal/usecase/register_key.go`

```go
package usecase

import (
    "context"
    "errors"

    "github.com/google/uuid"
    "github.com/lbpay/core-dict/internal/domain/entity"
    "github.com/lbpay/core-dict/internal/domain/repository"
    "github.com/lbpay/core-dict/internal/domain/service"
    "github.com/lbpay/core-dict/internal/domain/valueobject"
    "github.com/lbpay/core-dict/internal/infrastructure/messaging"
)

type RegisterKeyInput struct {
    KeyType       valueobject.KeyType
    KeyValue      string
    AccountID     uuid.UUID
    Account       valueobject.Account
    Owner         valueobject.Owner
    RequestedBy   uuid.UUID  // User ID
}

type RegisterKeyOutput struct {
    KeyID  uuid.UUID
    Status valueobject.KeyStatus
}

type RegisterKeyUseCase struct {
    keyRepo      repository.KeyRepository
    keyValidator *service.KeyValidator
    eventBus     messaging.EventBus
}

func NewRegisterKeyUseCase(
    keyRepo repository.KeyRepository,
    keyValidator *service.KeyValidator,
    eventBus messaging.EventBus,
) *RegisterKeyUseCase {
    return &RegisterKeyUseCase{
        keyRepo:      keyRepo,
        keyValidator: keyValidator,
        eventBus:     eventBus,
    }
}

func (uc *RegisterKeyUseCase) Execute(ctx context.Context, input RegisterKeyInput) (*RegisterKeyOutput, error) {
    // 1. Validar formato da chave
    if err := uc.keyValidator.ValidateFormat(input.KeyType, input.KeyValue); err != nil {
        return nil, errors.New("invalid key format: " + err.Error())
    }

    // 2. Validar titularidade (CPF/CNPJ deve pertencer ao dono da conta)
    if err := uc.validateOwnership(input); err != nil {
        return nil, err
    }

    // 3. Validar unicidade local (chave já cadastrada neste PSP?)
    existingKey, err := uc.keyRepo.FindByValue(ctx, input.KeyValue)
    if err == nil && existingKey.Status == valueobject.KeyStatusActive {
        return nil, errors.New("key already registered in this PSP")
    }

    // 4. Validar limites (ex: máximo 5 chaves CPF por titular)
    if err := uc.validateLimits(ctx, input); err != nil {
        return nil, err
    }

    // 5. Criar entidade de domínio
    key, err := entity.NewDictKey(
        input.KeyType,
        input.KeyValue,
        input.AccountID,
        input.Account,
        input.Owner,
    )
    if err != nil {
        return nil, err
    }

    // 6. Persistir (transaction)
    err = uc.keyRepo.WithTransaction(ctx, func(txCtx context.Context) error {
        // 6.1. Salvar chave
        if err := uc.keyRepo.Create(txCtx, key); err != nil {
            return err
        }

        // 6.2. Publicar eventos de domínio
        for _, evt := range key.DomainEvents() {
            if err := uc.eventBus.Publish(txCtx, evt); err != nil {
                return err  // Rollback transaction
            }
        }

        return nil
    })

    if err != nil {
        return nil, err
    }

    // 7. Limpar eventos (já publicados)
    key.ClearDomainEvents()

    return &RegisterKeyOutput{
        KeyID:  key.ID,
        Status: key.Status,
    }, nil
}

func (uc *RegisterKeyUseCase) validateOwnership(input RegisterKeyInput) error {
    switch input.KeyType {
    case valueobject.KeyTypeCPF:
        if input.Owner.TaxID != input.KeyValue {
            return errors.New("CPF key must match account owner CPF")
        }
    case valueobject.KeyTypeCNPJ:
        if input.Owner.TaxID != input.KeyValue {
            return errors.New("CNPJ key must match account owner CNPJ")
        }
    case valueobject.KeyTypeEmail, valueobject.KeyTypePhone:
        // Email/Phone ownership validated via OTP (separate flow)
        // Aqui assume que OTP já foi validado
    case valueobject.KeyTypeEVP:
        // EVP não requer validação de titularidade (gerado aleatoriamente)
    }
    return nil
}

func (uc *RegisterKeyUseCase) validateLimits(ctx context.Context, input RegisterKeyInput) error {
    limits := map[valueobject.KeyType]int{
        valueobject.KeyTypeCPF:   5,
        valueobject.KeyTypeCNPJ:  1,
        valueobject.KeyTypeEmail: 20,
        valueobject.KeyTypePhone: 20,
        valueobject.KeyTypeEVP:   20,
    }

    limit, ok := limits[input.KeyType]
    if !ok {
        return errors.New("unknown key type")
    }

    count, err := uc.keyRepo.CountByOwnerAndType(ctx, input.Owner.TaxID, input.KeyType)
    if err != nil {
        return err
    }

    if count >= limit {
        return errors.New("key limit exceeded for this type")
    }

    return nil
}
```

### 3.2 Use Case: `DeleteKeyUseCase`

**Arquivo**: `internal/usecase/delete_key.go`

```go
package usecase

import (
    "context"
    "errors"

    "github.com/google/uuid"
    "github.com/lbpay/core-dict/internal/domain/repository"
    "github.com/lbpay/core-dict/internal/infrastructure/messaging"
)

type DeleteKeyInput struct {
    KeyID           uuid.UUID
    TwoFactorCode   string  // 2FA obrigatório
    RequestedBy     uuid.UUID
}

type DeleteKeyOutput struct {
    Success bool
}

type DeleteKeyUseCase struct {
    keyRepo  repository.KeyRepository
    eventBus messaging.EventBus
}

func NewDeleteKeyUseCase(
    keyRepo repository.KeyRepository,
    eventBus messaging.EventBus,
) *DeleteKeyUseCase {
    return &DeleteKeyUseCase{
        keyRepo:  keyRepo,
        eventBus: eventBus,
    }
}

func (uc *DeleteKeyUseCase) Execute(ctx context.Context, input DeleteKeyInput) (*DeleteKeyOutput, error) {
    // 1. Buscar chave
    key, err := uc.keyRepo.FindByID(ctx, input.KeyID)
    if err != nil {
        return nil, errors.New("key not found")
    }

    // 2. Validar 2FA
    if err := uc.validate2FA(input.TwoFactorCode); err != nil {
        return nil, errors.New("invalid 2FA code")
    }

    // 3. Deletar (soft delete)
    if err := key.Delete(); err != nil {
        return nil, err
    }

    // 4. Persistir (transaction)
    err = uc.keyRepo.WithTransaction(ctx, func(txCtx context.Context) error {
        // 4.1. Atualizar chave
        if err := uc.keyRepo.Update(txCtx, key); err != nil {
            return err
        }

        // 4.2. Publicar eventos
        for _, evt := range key.DomainEvents() {
            if err := uc.eventBus.Publish(txCtx, evt); err != nil {
                return err
            }
        }

        return nil
    })

    if err != nil {
        return nil, err
    }

    key.ClearDomainEvents()

    return &DeleteKeyOutput{Success: true}, nil
}

func (uc *DeleteKeyUseCase) validate2FA(code string) error {
    // TODO: Integrar com serviço de 2FA (TOTP, SMS, etc.)
    if code == "" {
        return errors.New("2FA code required")
    }
    return nil
}
```

---

## 4. Interface Layer (gRPC APIs)

### 4.1 Proto Definition

**Arquivo**: `api/proto/dict/v1/dict.proto`

```protobuf
syntax = "proto3";

package dict.v1;

option go_package = "github.com/lbpay/core-dict/api/proto/dict/v1";

import "google/protobuf/timestamp.proto";

service DictService {
  rpc RegisterKey(RegisterKeyRequest) returns (RegisterKeyResponse);
  rpc DeleteKey(DeleteKeyRequest) returns (DeleteKeyResponse);
  rpc GetEntry(GetEntryRequest) returns (GetEntryResponse);
  rpc ListKeys(ListKeysRequest) returns (stream Key);
}

message RegisterKeyRequest {
  KeyType key_type = 1;
  string key_value = 2;
  string account_id = 3;
  string otp = 4;  // Optional (email/phone)
}

message RegisterKeyResponse {
  string key_id = 1;
  KeyStatus status = 2;
  string error_code = 3;
  string error_message = 4;
  google.protobuf.Timestamp created_at = 5;
}

message DeleteKeyRequest {
  string key_id = 1;
  string two_factor_code = 2;
}

message DeleteKeyResponse {
  bool success = 1;
  string error_code = 2;
  string error_message = 3;
}

message GetEntryRequest {
  string key_value = 1;
}

message GetEntryResponse {
  Key key = 1;
  Account account = 2;
  Owner owner = 3;
}

message ListKeysRequest {
  string account_id = 1;
  int32 page_size = 2;
  string page_token = 3;
}

message Key {
  string key_id = 1;
  KeyType key_type = 2;
  string key_value = 3;
  KeyStatus status = 4;
  google.protobuf.Timestamp created_at = 5;
}

message Account {
  string ispb = 1;
  string branch = 2;
  string account_number = 3;
  AccountType account_type = 4;
}

message Owner {
  OwnerType owner_type = 1;
  string tax_id = 2;
  string name = 3;
}

enum KeyType {
  KEY_TYPE_UNSPECIFIED = 0;
  KEY_TYPE_CPF = 1;
  KEY_TYPE_CNPJ = 2;
  KEY_TYPE_EMAIL = 3;
  KEY_TYPE_PHONE = 4;
  KEY_TYPE_EVP = 5;
}

enum KeyStatus {
  KEY_STATUS_UNSPECIFIED = 0;
  KEY_STATUS_PENDING = 1;
  KEY_STATUS_ACTIVE = 2;
  KEY_STATUS_FAILED = 3;
  KEY_STATUS_DELETED = 4;
}

enum AccountType {
  ACCOUNT_TYPE_UNSPECIFIED = 0;
  ACCOUNT_TYPE_CACC = 1;
  ACCOUNT_TYPE_SVGS = 2;
  ACCOUNT_TYPE_SLRY = 3;
}

enum OwnerType {
  OWNER_TYPE_UNSPECIFIED = 0;
  OWNER_TYPE_NATURAL_PERSON = 1;
  OWNER_TYPE_LEGAL_ENTITY = 2;
}
```

### 4.2 gRPC Handler

**Arquivo**: `internal/interface/grpc/dict_handler.go`

```go
package grpc

import (
    "context"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/timestamppb"

    dictv1 "github.com/lbpay/core-dict/api/proto/dict/v1"
    "github.com/lbpay/core-dict/internal/usecase"
    "github.com/lbpay/core-dict/internal/interface/dto"
)

type DictHandler struct {
    dictv1.UnimplementedDictServiceServer
    registerKeyUC *usecase.RegisterKeyUseCase
    deleteKeyUC   *usecase.DeleteKeyUseCase
    // ... outros use cases
}

func NewDictHandler(
    registerKeyUC *usecase.RegisterKeyUseCase,
    deleteKeyUC *usecase.DeleteKeyUseCase,
) *DictHandler {
    return &DictHandler{
        registerKeyUC: registerKeyUC,
        deleteKeyUC:   deleteKeyUC,
    }
}

func (h *DictHandler) RegisterKey(ctx context.Context, req *dictv1.RegisterKeyRequest) (*dictv1.RegisterKeyResponse, error) {
    // 1. Validar request
    if req.KeyType == dictv1.KeyType_KEY_TYPE_UNSPECIFIED {
        return nil, status.Error(codes.InvalidArgument, "key_type is required")
    }

    // 2. Converter proto → domain (DTO mapper)
    input, err := dto.ToRegisterKeyInput(req)
    if err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "invalid input: %v", err)
    }

    // 3. Executar use case
    output, err := h.registerKeyUC.Execute(ctx, *input)
    if err != nil {
        // Mapear erro de domínio → gRPC status code
        return nil, mapErrorToStatus(err)
    }

    // 4. Converter domain → proto
    return &dictv1.RegisterKeyResponse{
        KeyId:     output.KeyID.String(),
        Status:    dto.ToProtoKeyStatus(output.Status),
        CreatedAt: timestamppb.Now(),
    }, nil
}

func (h *DictHandler) DeleteKey(ctx context.Context, req *dictv1.DeleteKeyRequest) (*dictv1.DeleteKeyResponse, error) {
    input, err := dto.ToDeleteKeyInput(req)
    if err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "invalid input: %v", err)
    }

    output, err := h.deleteKeyUC.Execute(ctx, *input)
    if err != nil {
        return nil, mapErrorToStatus(err)
    }

    return &dictv1.DeleteKeyResponse{
        Success: output.Success,
    }, nil
}

func mapErrorToStatus(err error) error {
    // TODO: Mapear erros específicos de domínio → gRPC codes
    // Ex: ErrKeyAlreadyExists → codes.AlreadyExists
    //     ErrInvalidFormat → codes.InvalidArgument
    //     ErrLimitExceeded → codes.ResourceExhausted
    return status.Error(codes.Internal, err.Error())
}
```

### 4.3 Interceptors

#### Auth Interceptor

**Arquivo**: `internal/interface/grpc/interceptor/auth.go`

```go
package interceptor

import (
    "context"
    "strings"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
)

type AuthInterceptor struct {
    // TODO: JWT validator
}

func NewAuthInterceptor() *AuthInterceptor {
    return &AuthInterceptor{}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
    return func(
        ctx context.Context,
        req interface{},
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (interface{}, error) {
        // 1. Extrair token do metadata
        md, ok := metadata.FromIncomingContext(ctx)
        if !ok {
            return nil, status.Error(codes.Unauthenticated, "missing metadata")
        }

        authHeader := md.Get("authorization")
        if len(authHeader) == 0 {
            return nil, status.Error(codes.Unauthenticated, "missing authorization header")
        }

        token := strings.TrimPrefix(authHeader[0], "Bearer ")

        // 2. Validar JWT
        claims, err := i.validateJWT(token)
        if err != nil {
            return nil, status.Error(codes.Unauthenticated, "invalid token")
        }

        // 3. Injetar claims no context
        ctx = context.WithValue(ctx, "user_id", claims.UserID)
        ctx = context.WithValue(ctx, "account_id", claims.AccountID)

        // 4. Continuar
        return handler(ctx, req)
    }
}

func (i *AuthInterceptor) validateJWT(token string) (*Claims, error) {
    // TODO: Implementar validação JWT (RS256)
    return &Claims{UserID: "user_123", AccountID: "acc_456"}, nil
}

type Claims struct {
    UserID    string
    AccountID string
}
```

---

**CONTINUA NO PRÓXIMO BLOCO...**

Este documento tem ~150KB total. Vou continuar criando as seções restantes (Infrastructure Layer, Database Schema, Observabilidade, Testing, Deployment) em seguida.

Posso continuar com o restante do TEC-001?