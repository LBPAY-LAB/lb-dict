# DIA-003: C4 Component Diagram - Core DICT

**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: Equipe Arquitetura
**Status**: ✅ Completo

---

## Sumário Executivo

Este documento apresenta o **C4 Component Diagram** (nível 3) do **Core DICT**, detalhando os componentes internos da aplicação Core DICT API e como se organizam segundo a **Clean Architecture**.

**Objetivo**: Mostrar a estrutura interna do Core DICT API, separação de responsabilidades por camada, e como os componentes interagem entre si e com sistemas externos.

**Pré-requisitos**:
- [DIA-001: C4 Context Diagram](./DIA-001_C4_Context_Diagram.md)
- [DIA-002: C4 Container Diagram](./DIA-002_C4_Container_Diagram.md)

---

## 1. Clean Architecture - 4 Camadas

O Core DICT segue **Clean Architecture** (Uncle Bob), com 4 camadas concêntricas:

```
┌─────────────────────────────────────────────────────┐
│  API Layer (Presentation)                           │  ← HTTP/gRPC handlers
│  - Controllers, Middleware, Validators              │
├─────────────────────────────────────────────────────┤
│  Application Layer (Use Cases)                      │  ← Business logic orchestration
│  - Services, DTOs, Mappers                          │
├─────────────────────────────────────────────────────┤
│  Domain Layer (Entities + Business Rules)           │  ← Core business logic
│  - Entities, Value Objects, Domain Services         │
├─────────────────────────────────────────────────────┤
│  Infrastructure Layer (External Interfaces)         │  ← DB, Messaging, External APIs
│  - Repositories, HTTP Clients, Pulsar Producers     │
└─────────────────────────────────────────────────────┘
```

**Regras de Dependência**:
- ✅ Camadas externas dependem de camadas internas
- ✅ Camadas internas NÃO conhecem camadas externas
- ✅ Domain Layer é o centro (zero dependências externas)
- ❌ Domain Layer NUNCA importa Infrastructure

**Referência**: [ADR-001: Clean Architecture](../ADRs/ADR-001_Clean_Architecture.md)

---

## 2. C4 Component Diagram - Core DICT API

### 2.1. Diagrama

```mermaid
C4Component
  title Component Diagram - Core DICT API (Clean Architecture)

  Container_Boundary(core_api, "Core DICT API") {

    Component_Boundary(api_layer, "API Layer (Presentation)") {
      Component(http_router, "HTTP Router", "Fiber v3", "Roteamento de requisições HTTP")
      Component(auth_middleware, "Auth Middleware", "Go", "Valida JWT, extrai claims")
      Component(rbac_middleware, "RBAC Middleware", "Go", "Verifica permissões por role")
      Component(entry_controller, "Entry Controller", "Go", "Handlers para /api/v1/keys/*")
      Component(claim_controller, "Claim Controller", "Go", "Handlers para /api/v1/claims/*")
      Component(portability_controller, "Portability Controller", "Go", "Handlers para /api/v1/portabilities/*")
      Component(health_controller, "Health Controller", "Go", "Health checks e readiness")
      Component(request_validator, "Request Validator", "Go + validator.v10", "Valida payloads JSON")
    }

    Component_Boundary(application_layer, "Application Layer (Use Cases)") {
      Component(entry_service, "Entry Service", "Go", "Casos de uso: CreateEntry, GetEntry, DeleteEntry")
      Component(claim_service, "Claim Service", "Go", "Casos de uso: CreateClaim, ConfirmClaim, CancelClaim")
      Component(portability_service, "Portability Service", "Go", "Casos de uso: StartPortability, ConfirmPortability")
      Component(dto_mapper, "DTO Mapper", "Go", "Converte Entity ↔ DTO")
      Component(event_publisher, "Event Publisher", "Go", "Publica eventos no Pulsar")
    }

    Component_Boundary(domain_layer, "Domain Layer (Entities + Rules)") {
      Component(entry_entity, "Entry Entity", "Go", "Chave PIX (key_type, key_value, account)")
      Component(claim_entity, "Claim Entity", "Go", "Reivindicação (30 dias)")
      Component(account_entity, "Account Entity", "Go", "Conta CID (ISPB, agência, conta)")
      Component(key_validator, "Key Validator", "Go", "Valida CPF, CNPJ, Email, Phone, EVP")
      Component(claim_rules, "Claim Business Rules", "Go", "Regras: 30 dias, ISPB diferente, etc.")
      Component(domain_events, "Domain Events", "Go", "EntryCreated, ClaimCreated, etc.")
    }

    Component_Boundary(infrastructure_layer, "Infrastructure Layer") {
      Component(entry_repository, "Entry Repository", "Go + pgx", "CRUD entries no PostgreSQL")
      Component(claim_repository, "Claim Repository", "Go + pgx", "CRUD claims no PostgreSQL")
      Component(audit_repository, "Audit Repository", "Go + pgx", "Insere audit logs")
      Component(pulsar_producer, "Pulsar Producer", "Go + pulsar-client-go", "Publica eventos dict.entries.*, dict.claims.*")
      Component(ledger_client, "Ledger gRPC Client", "Go + gRPC", "Valida contas CID no LBPay Ledger")
      Component(auth_client, "Auth HTTP Client", "Go + resty", "Valida JWT no LBPay Auth")
    }
  }

  ContainerDb(core_db, "Core Database", "PostgreSQL 16", "Schemas dict + audit")
  ContainerQueue(pulsar, "Apache Pulsar", "Pulsar v0.16.0", "Event streaming")
  System_Ext(ledger, "LBPay Ledger", "Validação de contas")
  System_Ext(auth, "LBPay Auth", "Validação de JWT")

  Person(user, "Usuário Final", "Cliente")

  Rel(user, http_router, "HTTPS REST + JWT", "POST /api/v1/keys")

  Rel(http_router, auth_middleware, "Middleware chain")
  Rel(auth_middleware, auth_client, "Valida token", "HTTPS")
  Rel(auth_middleware, rbac_middleware, "Next middleware")

  Rel(rbac_middleware, entry_controller, "Route to controller")

  Rel(entry_controller, request_validator, "Valida payload")
  Rel(entry_controller, entry_service, "CreateEntry(dto)")

  Rel(entry_service, dto_mapper, "DTO → Entity")
  Rel(entry_service, key_validator, "Valida key_value")
  Rel(entry_service, ledger_client, "Valida conta", "gRPC")
  Rel(entry_service, entry_entity, "Cria entity")
  Rel(entry_service, entry_repository, "Save(entry)")
  Rel(entry_service, event_publisher, "Publica EntryCreated")

  Rel(entry_repository, core_db, "SQL INSERT", "pgx")
  Rel(audit_repository, core_db, "SQL INSERT audit", "pgx")
  Rel(event_publisher, pulsar_producer, "Send event")
  Rel(pulsar_producer, pulsar, "Publish", "Pulsar Protocol")

  Rel(claim_controller, claim_service, "CreateClaim(dto)")
  Rel(claim_service, claim_entity, "Cria claim")
  Rel(claim_service, claim_rules, "Valida regras de negócio")
  Rel(claim_service, claim_repository, "Save(claim)")

  Rel(health_controller, core_db, "Ping DB", "pgx")
  Rel(health_controller, pulsar, "Ping Pulsar", "Pulsar Client")

  UpdateLayoutConfig($c4ShapeInRow="3", $c4BoundaryInRow="1")
```

---

### 2.2. Versão PlantUML (Alternativa)

```plantuml
@startuml
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Component.puml

LAYOUT_WITH_LEGEND()

title Component Diagram - Core DICT API

Container_Boundary(core_api, "Core DICT API") {

  ' API Layer
  Component(router, "HTTP Router", "Fiber v3")
  Component(auth_mw, "Auth Middleware", "Go")
  Component(entry_ctrl, "Entry Controller", "Go")
  Component(claim_ctrl, "Claim Controller", "Go")
  Component(validator, "Request Validator", "Go")

  ' Application Layer
  Component(entry_svc, "Entry Service", "Go")
  Component(claim_svc, "Claim Service", "Go")
  Component(mapper, "DTO Mapper", "Go")
  Component(event_pub, "Event Publisher", "Go")

  ' Domain Layer
  Component(entry_entity, "Entry Entity", "Go")
  Component(claim_entity, "Claim Entity", "Go")
  Component(key_validator, "Key Validator", "Go")
  Component(claim_rules, "Claim Rules", "Go")

  ' Infrastructure Layer
  Component(entry_repo, "Entry Repository", "Go + pgx")
  Component(claim_repo, "Claim Repository", "Go + pgx")
  Component(pulsar_prod, "Pulsar Producer", "Go")
  Component(ledger_client, "Ledger Client", "gRPC")
  Component(auth_client, "Auth Client", "HTTP")
}

ContainerDb(db, "Core DB", "PostgreSQL 16")
ContainerQueue(pulsar, "Pulsar", "v0.16.0")
System_Ext(ledger, "LBPay Ledger")
System_Ext(auth, "LBPay Auth")

Person(user, "Usuário")

Rel(user, router, "HTTPS")
Rel(router, auth_mw, "Middleware")
Rel(auth_mw, auth_client, "Valida JWT")
Rel(auth_mw, entry_ctrl, "Route")

Rel(entry_ctrl, validator, "Valida")
Rel(entry_ctrl, entry_svc, "CreateEntry")

Rel(entry_svc, mapper, "DTO → Entity")
Rel(entry_svc, key_validator, "Valida key")
Rel(entry_svc, ledger_client, "Valida conta")
Rel(entry_svc, entry_entity, "Cria")
Rel(entry_svc, entry_repo, "Save")
Rel(entry_svc, event_pub, "Publica evento")

Rel(entry_repo, db, "SQL")
Rel(pulsar_prod, pulsar, "Publish")
Rel(event_pub, pulsar_prod, "Send")

Rel(claim_ctrl, claim_svc, "CreateClaim")
Rel(claim_svc, claim_entity, "Cria")
Rel(claim_svc, claim_rules, "Valida regras")
Rel(claim_svc, claim_repo, "Save")

@enduml
```

---

## 3. Componentes por Camada

### 3.1. API Layer (Presentation)

#### HTTP Router
- **Responsabilidade**: Roteamento de requisições HTTP
- **Tecnologia**: Fiber v3
- **Rotas Principais**:
  ```go
  // Entries
  POST   /api/v1/keys
  GET    /api/v1/keys/:key_type/:key_value
  DELETE /api/v1/keys/:key_type/:key_value
  GET    /api/v1/keys  // List user's keys

  // Claims
  POST   /api/v1/claims
  GET    /api/v1/claims/:claim_id
  POST   /api/v1/claims/:claim_id/confirm
  POST   /api/v1/claims/:claim_id/cancel
  GET    /api/v1/claims  // List claims (owner or claimer)

  // Portabilities
  POST   /api/v1/portabilities
  GET    /api/v1/portabilities/:portability_id
  POST   /api/v1/portabilities/:portability_id/confirm

  // Health
  GET    /health
  GET    /ready
  ```
- **Middleware Chain**: Logger → Auth → RBAC → Rate Limit → Controller
- **Localização**: `internal/api/http/router.go`

#### Auth Middleware
- **Responsabilidade**: Validar JWT, extrair claims (user_id, roles, scopes)
- **Tecnologia**: Go + jwt-go
- **Fluxo**:
  ```go
  func AuthMiddleware(c *fiber.Ctx) error {
      token := c.Get("Authorization") // "Bearer <jwt>"
      claims, err := authClient.ValidateToken(token)
      if err != nil {
          return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
      }
      c.Locals("user_id", claims.UserID)
      c.Locals("roles", claims.Roles)
      return c.Next()
  }
  ```
- **Dependência**: Auth HTTP Client (Infrastructure)
- **Localização**: `internal/api/http/middleware/auth.go`

#### RBAC Middleware
- **Responsabilidade**: Verificar permissões por role e scope
- **Roles**: `user`, `admin`, `support`, `auditor`
- **Scopes**: `dict:read`, `dict:write`, `dict:admin`
- **Exemplo**:
  ```go
  func RequireScope(scope string) fiber.Handler {
      return func(c *fiber.Ctx) error {
          scopes := c.Locals("scopes").([]string)
          if !contains(scopes, scope) {
              return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
          }
          return c.Next()
      }
  }

  // Usage
  app.Post("/api/v1/keys", RequireScope("dict:write"), entryController.CreateEntry)
  ```
- **Localização**: `internal/api/http/middleware/rbac.go`

#### Entry Controller
- **Responsabilidade**: Handlers para operações de chaves PIX
- **Métodos**:
  ```go
  type EntryController struct {
      entryService *application.EntryService
      validator    *validator.Validate
  }

  func (ec *EntryController) CreateEntry(c *fiber.Ctx) error {
      var req CreateEntryRequest
      if err := c.BodyParser(&req); err != nil {
          return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
      }

      if err := ec.validator.Struct(req); err != nil {
          return c.Status(400).JSON(fiber.Map{"error": err.Error()})
      }

      userID := c.Locals("user_id").(string)
      entry, err := ec.entryService.CreateEntry(c.Context(), userID, req)
      if err != nil {
          return handleError(c, err)
      }

      return c.Status(201).JSON(entry)
  }

  func (ec *EntryController) GetEntry(c *fiber.Ctx) error { ... }
  func (ec *EntryController) DeleteEntry(c *fiber.Ctx) error { ... }
  func (ec *EntryController) ListEntries(c *fiber.Ctx) error { ... }
  ```
- **Localização**: `internal/api/http/controllers/entry_controller.go`

#### Claim Controller
- **Responsabilidade**: Handlers para operações de reivindicações
- **Métodos**: `CreateClaim`, `GetClaim`, `ConfirmClaim`, `CancelClaim`, `ListClaims`
- **Validações Específicas**:
  - `completion_period_days` DEVE ser 30 (TEC-003 v2.1)
  - ISPB claimer DEVE ser diferente de ISPB owner
- **Localização**: `internal/api/http/controllers/claim_controller.go`

#### Request Validator
- **Responsabilidade**: Validar payloads JSON com tags de validação
- **Tecnologia**: `go-playground/validator.v10`
- **Exemplo**:
  ```go
  type CreateEntryRequest struct {
      KeyType    string  `json:"key_type" validate:"required,oneof=CPF CNPJ EMAIL PHONE EVP"`
      KeyValue   string  `json:"key_value" validate:"required,min=1,max=255"`
      Account    Account `json:"account" validate:"required"`
  }

  type Account struct {
      ISPB          string `json:"ispb" validate:"required,len=8,numeric"`
      AccountNumber string `json:"account_number" validate:"required"`
      Branch        string `json:"branch" validate:"required"`
      AccountType   string `json:"account_type" validate:"required,oneof=CACC SVGS TRAN"`
  }
  ```
- **Custom Validators**: CPF, CNPJ, Email (RFC 5322), Phone (E.164)
- **Localização**: `internal/api/http/validators/request_validator.go`

---

### 3.2. Application Layer (Use Cases)

#### Entry Service
- **Responsabilidade**: Orquestrar casos de uso de chaves PIX
- **Casos de Uso**:
  ```go
  type EntryService struct {
      entryRepo      domain.EntryRepository
      auditRepo      domain.AuditRepository
      ledgerClient   infrastructure.LedgerClient
      eventPublisher infrastructure.EventPublisher
      keyValidator   domain.KeyValidator
      mapper         DTOMapper
  }

  func (es *EntryService) CreateEntry(ctx context.Context, userID string, req CreateEntryRequest) (*EntryDTO, error) {
      // 1. Mapear DTO → Entity
      entry := es.mapper.ToEntryEntity(req)
      entry.CreatedBy = userID

      // 2. Validar key_value (CPF, CNPJ, etc.)
      if err := es.keyValidator.Validate(entry.KeyType, entry.KeyValue); err != nil {
          return nil, ErrInvalidKey
      }

      // 3. Validar conta CID no Ledger
      account, err := es.ledgerClient.ValidateAccount(ctx, req.Account.ID)
      if err != nil || account.Status != "ACTIVE" {
          return nil, ErrInvalidAccount
      }

      // 4. Verificar se chave já existe
      exists, err := es.entryRepo.ExistsByKey(ctx, entry.KeyType, entry.KeyValue)
      if exists {
          return nil, ErrKeyAlreadyExists
      }

      // 5. Persistir entry (status PENDING)
      entry.Status = "PENDING"
      if err := es.entryRepo.Create(ctx, entry); err != nil {
          return nil, err
      }

      // 6. Auditoria
      es.auditRepo.Log(ctx, AuditLog{
          EntityType: "entry",
          EntityID:   entry.ID,
          Action:     "CREATE",
          UserID:     userID,
          Timestamp:  time.Now(),
      })

      // 7. Publicar evento
      event := domain.EntryCreatedEvent{
          EntryID:   entry.ID,
          KeyType:   entry.KeyType,
          KeyValue:  entry.KeyValue,
          Account:   entry.Account,
          Timestamp: time.Now(),
      }
      es.eventPublisher.Publish(ctx, "dict.entries.created", event)

      // 8. Mapear Entity → DTO
      return es.mapper.ToEntryDTO(entry), nil
  }

  func (es *EntryService) GetEntry(ctx context.Context, keyType, keyValue string) (*EntryDTO, error) { ... }
  func (es *EntryService) DeleteEntry(ctx context.Context, userID, keyType, keyValue string) error { ... }
  func (es *EntryService) ListUserEntries(ctx context.Context, userID string) ([]*EntryDTO, error) { ... }
  ```
- **Localização**: `internal/application/services/entry_service.go`

#### Claim Service
- **Responsabilidade**: Orquestrar casos de uso de reivindicações
- **Casos de Uso**: `CreateClaim`, `ConfirmClaim`, `CancelClaim`, `GetClaim`, `ListClaims`
- **Regras de Negócio Aplicadas**:
  - `completion_period_days` fixo em 30 (TEC-003 v2.1)
  - ISPB claimer ≠ ISPB owner
  - Entry deve existir e estar ACTIVE
  - Não pode haver claim OPEN para mesma entry
- **Localização**: `internal/application/services/claim_service.go`

#### DTO Mapper
- **Responsabilidade**: Converter Entity ↔ DTO (Data Transfer Object)
- **Por que necessário**: Desacoplar Domain Layer de API Layer
- **Exemplo**:
  ```go
  type DTOMapper struct{}

  func (m *DTOMapper) ToEntryEntity(dto CreateEntryRequest) *domain.Entry {
      return &domain.Entry{
          ID:       uuid.New(),
          KeyType:  dto.KeyType,
          KeyValue: dto.KeyValue,
          Account:  m.ToAccountEntity(dto.Account),
      }
  }

  func (m *DTOMapper) ToEntryDTO(entity *domain.Entry) *EntryDTO {
      return &EntryDTO{
          ID:        entity.ID.String(),
          KeyType:   entity.KeyType,
          KeyValue:  entity.KeyValue,
          Account:   m.ToAccountDTO(entity.Account),
          Status:    entity.Status,
          CreatedAt: entity.CreatedAt,
      }
  }
  ```
- **Localização**: `internal/application/mappers/dto_mapper.go`

#### Event Publisher
- **Responsabilidade**: Interface para publicar eventos de domínio
- **Interface**:
  ```go
  type EventPublisher interface {
      Publish(ctx context.Context, topic string, event interface{}) error
      PublishBatch(ctx context.Context, topic string, events []interface{}) error
  }
  ```
- **Implementação**: Pulsar Producer (Infrastructure Layer)
- **Localização**: `internal/application/ports/event_publisher.go` (interface)

---

### 3.3. Domain Layer (Entities + Business Rules)

#### Entry Entity
- **Responsabilidade**: Representar chave PIX como entidade de domínio
- **Estrutura**:
  ```go
  type Entry struct {
      ID         uuid.UUID
      KeyType    KeyType  // CPF, CNPJ, EMAIL, PHONE, EVP
      KeyValue   string
      Account    Account
      Status     EntryStatus  // PENDING, ACTIVE, DELETED
      ExternalID string       // ID retornado pelo Bacen
      CreatedBy  string       // user_id
      CreatedAt  time.Time
      UpdatedAt  time.Time
      DeletedAt  *time.Time   // Soft delete
  }

  type KeyType string
  const (
      KeyTypeCPF   KeyType = "CPF"
      KeyTypeCNPJ  KeyType = "CNPJ"
      KeyTypeEMAIL KeyType = "EMAIL"
      KeyTypePHONE KeyType = "PHONE"
      KeyTypeEVP   KeyType = "EVP"
  )

  type EntryStatus string
  const (
      EntryStatusPending EntryStatus = "PENDING"
      EntryStatusActive  EntryStatus = "ACTIVE"
      EntryStatusDeleted EntryStatus = "DELETED"
  )
  ```
- **Métodos de Negócio**:
  ```go
  func (e *Entry) Activate(externalID string) error {
      if e.Status != EntryStatusPending {
          return ErrEntryNotPending
      }
      e.Status = EntryStatusActive
      e.ExternalID = externalID
      e.UpdatedAt = time.Now()
      return nil
  }

  func (e *Entry) MarkAsDeleted() error {
      if e.Status != EntryStatusActive {
          return ErrEntryNotActive
      }
      now := time.Now()
      e.Status = EntryStatusDeleted
      e.DeletedAt = &now
      e.UpdatedAt = now
      return nil
  }

  func (e *Entry) IsOwnedBy(ispb string) bool {
      return e.Account.ISPB == ispb
  }
  ```
- **Localização**: `internal/domain/entities/entry.go`

#### Claim Entity
- **Responsabilidade**: Representar reivindicação de chave PIX
- **Estrutura**:
  ```go
  type Claim struct {
      ID                   uuid.UUID
      WorkflowID           string        // Temporal workflow ID
      Entry                *Entry
      ClaimerAccount       Account
      OwnerAccount         Account
      CompletionPeriodDays int           // SEMPRE 30 (TEC-003 v2.1)
      ExpiresAt            time.Time     // CreatedAt + 30 dias
      Status               ClaimStatus
      Resolution           *ClaimResolution  // Confirmado, Cancelado, Expirado
      CreatedAt            time.Time
      UpdatedAt            time.Time
  }

  type ClaimStatus string
  const (
      ClaimStatusOpen              ClaimStatus = "OPEN"
      ClaimStatusWaitingResolution ClaimStatus = "WAITING_RESOLUTION"
      ClaimStatusConfirmed         ClaimStatus = "CONFIRMED"
      ClaimStatusCancelled         ClaimStatus = "CANCELLED"
      ClaimStatusCompleted         ClaimStatus = "COMPLETED"
      ClaimStatusExpired           ClaimStatus = "EXPIRED"
  )
  ```
- **Métodos de Negócio**:
  ```go
  func NewClaim(entry *Entry, claimerAccount Account) (*Claim, error) {
      if entry.Account.ISPB == claimerAccount.ISPB {
          return nil, ErrSameISPB
      }

      return &Claim{
          ID:                   uuid.New(),
          Entry:                entry,
          ClaimerAccount:       claimerAccount,
          OwnerAccount:         entry.Account,
          CompletionPeriodDays: 30,  // Fixo (TEC-003 v2.1)
          ExpiresAt:            time.Now().AddDate(0, 0, 30),
          Status:               ClaimStatusOpen,
          CreatedAt:            time.Now(),
      }, nil
  }

  func (c *Claim) Confirm() error {
      if c.Status != ClaimStatusOpen && c.Status != ClaimStatusWaitingResolution {
          return ErrInvalidClaimStatus
      }
      c.Status = ClaimStatusConfirmed
      c.Resolution = &ClaimResolution{Type: "CONFIRMED", Timestamp: time.Now()}
      c.UpdatedAt = time.Now()
      return nil
  }

  func (c *Claim) Cancel(reason string) error {
      if c.Status != ClaimStatusOpen {
          return ErrInvalidClaimStatus
      }
      c.Status = ClaimStatusCancelled
      c.Resolution = &ClaimResolution{Type: "CANCELLED", Reason: reason, Timestamp: time.Now()}
      c.UpdatedAt = time.Now()
      return nil
  }

  func (c *Claim) Expire() error {
      if time.Now().After(c.ExpiresAt) {
          c.Status = ClaimStatusExpired
          c.Resolution = &ClaimResolution{Type: "EXPIRED", Timestamp: time.Now()}
          c.UpdatedAt = time.Now()
          return nil
      }
      return ErrClaimNotExpired
  }
  ```
- **Localização**: `internal/domain/entities/claim.go`

#### Account Entity
- **Responsabilidade**: Representar conta CID
- **Estrutura**:
  ```go
  type Account struct {
      ID            uuid.UUID
      ISPB          string   // 8 dígitos
      AccountNumber string
      Branch        string
      AccountType   AccountType  // CACC, SVGS, TRAN
      HolderName    string
      HolderDocument string   // CPF ou CNPJ
      Status        AccountStatus
  }

  type AccountType string
  const (
      AccountTypeCACC AccountType = "CACC"  // Conta Corrente
      AccountTypeSVGS AccountType = "SVGS"  // Conta Poupança
      AccountTypeTRAN AccountType = "TRAN"  // Conta Pagamento
  )
  ```
- **Localização**: `internal/domain/entities/account.go`

#### Key Validator (Domain Service)
- **Responsabilidade**: Validar key_value por tipo de chave
- **Métodos**:
  ```go
  type KeyValidator struct{}

  func (kv *KeyValidator) Validate(keyType KeyType, keyValue string) error {
      switch keyType {
      case KeyTypeCPF:
          return kv.validateCPF(keyValue)
      case KeyTypeCNPJ:
          return kv.validateCNPJ(keyValue)
      case KeyTypeEMAIL:
          return kv.validateEmail(keyValue)
      case KeyTypePHONE:
          return kv.validatePhone(keyValue)
      case KeyTypeEVP:
          return kv.validateEVP(keyValue)
      default:
          return ErrInvalidKeyType
      }
  }

  func (kv *KeyValidator) validateCPF(cpf string) error {
      // Remove formatação (., -)
      cpf = strings.ReplaceAll(cpf, ".", "")
      cpf = strings.ReplaceAll(cpf, "-", "")

      if len(cpf) != 11 {
          return ErrInvalidCPF
      }

      // Verifica dígitos verificadores
      if !kv.validateCPFCheckDigits(cpf) {
          return ErrInvalidCPF
      }

      return nil
  }

  func (kv *KeyValidator) validateEmail(email string) error {
      // RFC 5322
      regex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
      if !regex.MatchString(email) {
          return ErrInvalidEmail
      }
      return nil
  }

  func (kv *KeyValidator) validatePhone(phone string) error {
      // E.164: +5511999999999
      regex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
      if !regex.MatchString(phone) {
          return ErrInvalidPhone
      }
      return nil
  }

  func (kv *KeyValidator) validateEVP(evp string) error {
      // UUID v4
      _, err := uuid.Parse(evp)
      return err
  }
  ```
- **Localização**: `internal/domain/services/key_validator.go`

#### Claim Business Rules (Domain Service)
- **Responsabilidade**: Validar regras de negócio de reivindicações
- **Regras**:
  ```go
  type ClaimRules struct{}

  func (cr *ClaimRules) ValidateClaimCreation(entry *Entry, claimerISPB string) error {
      // 1. Entry deve estar ACTIVE
      if entry.Status != EntryStatusActive {
          return ErrEntryNotActive
      }

      // 2. ISPB claimer deve ser diferente de ISPB owner
      if entry.Account.ISPB == claimerISPB {
          return ErrSameISPB
      }

      // 3. Não pode haver claim OPEN para mesma entry
      // (verificado no repository antes de chamar esta regra)

      return nil
  }

  func (cr *ClaimRules) ValidateCompletionPeriod(days int) error {
      // TEC-003 v2.1: completion_period_days DEVE ser 30
      if days != 30 {
          return ErrInvalidCompletionPeriod
      }
      return nil
  }

  func (cr *ClaimRules) CanConfirmClaim(claim *Claim, userISPB string) error {
      // Apenas owner pode confirmar
      if claim.OwnerAccount.ISPB != userISPB {
          return ErrUnauthorized
      }

      // Claim deve estar OPEN ou WAITING_RESOLUTION
      if claim.Status != ClaimStatusOpen && claim.Status != ClaimStatusWaitingResolution {
          return ErrInvalidClaimStatus
      }

      return nil
  }
  ```
- **Localização**: `internal/domain/services/claim_rules.go`

#### Domain Events
- **Responsabilidade**: Representar eventos de domínio
- **Eventos**:
  ```go
  type EntryCreatedEvent struct {
      EntryID   uuid.UUID
      KeyType   KeyType
      KeyValue  string
      Account   Account
      Timestamp time.Time
  }

  type ClaimCreatedEvent struct {
      ClaimID        uuid.UUID
      EntryID        uuid.UUID
      ClaimerAccount Account
      OwnerAccount   Account
      ExpiresAt      time.Time
      Timestamp      time.Time
  }

  type ClaimConfirmedEvent struct {
      ClaimID   uuid.UUID
      Timestamp time.Time
  }

  type ClaimExpiredEvent struct {
      ClaimID   uuid.UUID
      Timestamp time.Time
  }
  ```
- **Localização**: `internal/domain/events/events.go`

---

### 3.4. Infrastructure Layer

#### Entry Repository
- **Responsabilidade**: Persistir e recuperar entries do PostgreSQL
- **Interface (Domain)**:
  ```go
  type EntryRepository interface {
      Create(ctx context.Context, entry *Entry) error
      FindByID(ctx context.Context, id uuid.UUID) (*Entry, error)
      FindByKey(ctx context.Context, keyType KeyType, keyValue string) (*Entry, error)
      ExistsByKey(ctx context.Context, keyType KeyType, keyValue string) (bool, error)
      Update(ctx context.Context, entry *Entry) error
      Delete(ctx context.Context, id uuid.UUID) error  // Soft delete
      ListByUserID(ctx context.Context, userID string) ([]*Entry, error)
  }
  ```
- **Implementação (Infrastructure)**:
  ```go
  type PostgresEntryRepository struct {
      pool *pgxpool.Pool
  }

  func (r *PostgresEntryRepository) Create(ctx context.Context, entry *Entry) error {
      query := `
          INSERT INTO dict.entries (id, key_type, key_value, account_id, status, created_by, created_at)
          VALUES ($1, $2, $3, $4, $5, $6, $7)
      `
      _, err := r.pool.Exec(ctx, query,
          entry.ID,
          entry.KeyType,
          entry.KeyValue,
          entry.Account.ID,
          entry.Status,
          entry.CreatedBy,
          entry.CreatedAt,
      )
      return err
  }

  func (r *PostgresEntryRepository) FindByKey(ctx context.Context, keyType KeyType, keyValue string) (*Entry, error) {
      query := `
          SELECT e.id, e.key_type, e.key_value, e.status, e.external_id, e.created_at,
                 a.id, a.ispb, a.account_number, a.branch, a.account_type
          FROM dict.entries e
          JOIN dict.accounts a ON e.account_id = a.id
          WHERE e.key_type = $1 AND e.key_value = $2 AND e.deleted_at IS NULL
      `
      var entry Entry
      var account Account
      err := r.pool.QueryRow(ctx, query, keyType, keyValue).Scan(
          &entry.ID, &entry.KeyType, &entry.KeyValue, &entry.Status, &entry.ExternalID, &entry.CreatedAt,
          &account.ID, &account.ISPB, &account.AccountNumber, &account.Branch, &account.AccountType,
      )
      if err == pgx.ErrNoRows {
          return nil, ErrEntryNotFound
      }
      entry.Account = account
      return &entry, err
  }
  ```
- **Localização**: `internal/infrastructure/repositories/postgres_entry_repository.go`

#### Claim Repository
- **Interface**: `ClaimRepository` (similar a EntryRepository)
- **Métodos**: `Create`, `FindByID`, `Update`, `ListByEntryID`, `ListByISPB`
- **Localização**: `internal/infrastructure/repositories/postgres_claim_repository.go`

#### Audit Repository
- **Responsabilidade**: Inserir logs de auditoria
- **Interface**:
  ```go
  type AuditRepository interface {
      Log(ctx context.Context, log AuditLog) error
  }

  type AuditLog struct {
      ID         uuid.UUID
      EntityType string  // "entry", "claim", "portability"
      EntityID   uuid.UUID
      Action     string  // "CREATE", "UPDATE", "DELETE"
      UserID     string
      Changes    map[string]interface{}  // JSON
      Timestamp  time.Time
  }
  ```
- **Implementação**:
  ```go
  func (r *PostgresAuditRepository) Log(ctx context.Context, log AuditLog) error {
      query := `
          INSERT INTO audit.entry_events (id, entity_type, entity_id, action, user_id, changes, timestamp)
          VALUES ($1, $2, $3, $4, $5, $6, $7)
      `
      changesJSON, _ := json.Marshal(log.Changes)
      _, err := r.pool.Exec(ctx, query,
          uuid.New(),
          log.EntityType,
          log.EntityID,
          log.Action,
          log.UserID,
          changesJSON,
          log.Timestamp,
      )
      return err
  }
  ```
- **Localização**: `internal/infrastructure/repositories/postgres_audit_repository.go`

#### Pulsar Producer
- **Responsabilidade**: Publicar eventos no Apache Pulsar
- **Implementação**:
  ```go
  type PulsarEventPublisher struct {
      client   pulsar.Client
      producer pulsar.Producer
  }

  func NewPulsarEventPublisher(pulsarURL string) (*PulsarEventPublisher, error) {
      client, err := pulsar.NewClient(pulsar.ClientOptions{
          URL: pulsarURL,
      })
      if err != nil {
          return nil, err
      }

      producer, err := client.CreateProducer(pulsar.ProducerOptions{
          Topic: "dict.entries.created",  // Configurável por topic
      })
      if err != nil {
          return nil, err
      }

      return &PulsarEventPublisher{client: client, producer: producer}, nil
  }

  func (p *PulsarEventPublisher) Publish(ctx context.Context, topic string, event interface{}) error {
      payload, err := json.Marshal(event)
      if err != nil {
          return err
      }

      _, err = p.producer.Send(ctx, &pulsar.ProducerMessage{
          Payload: payload,
          Key:     generateKey(event),  // Partitioning por entry_id
      })
      return err
  }
  ```
- **Localização**: `internal/infrastructure/messaging/pulsar_producer.go`

#### Ledger gRPC Client
- **Responsabilidade**: Validar contas CID no LBPay Ledger
- **Interface**:
  ```go
  type LedgerClient interface {
      ValidateAccount(ctx context.Context, accountID string) (*AccountInfo, error)
  }

  type AccountInfo struct {
      Exists         bool
      Status         string  // "ACTIVE", "BLOCKED", "CLOSED"
      HolderName     string
      HolderDocument string
  }
  ```
- **Implementação**:
  ```go
  type GRPCLedgerClient struct {
      client ledgerpb.LedgerServiceClient
  }

  func (c *GRPCLedgerClient) ValidateAccount(ctx context.Context, accountID string) (*AccountInfo, error) {
      req := &ledgerpb.ValidateAccountRequest{AccountId: accountID}
      resp, err := c.client.ValidateAccount(ctx, req)
      if err != nil {
          return nil, err
      }

      return &AccountInfo{
          Exists:         resp.Exists,
          Status:         resp.Status,
          HolderName:     resp.HolderName,
          HolderDocument: resp.HolderDocument,
      }, nil
  }
  ```
- **Localização**: `internal/infrastructure/clients/ledger_grpc_client.go`

#### Auth HTTP Client
- **Responsabilidade**: Validar JWT no LBPay Auth
- **Interface**:
  ```go
  type AuthClient interface {
      ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
  }

  type TokenClaims struct {
      UserID string
      Roles  []string
      Scopes []string
  }
  ```
- **Implementação**:
  ```go
  type HTTPAuthClient struct {
      client  *resty.Client
      authURL string
  }

  func (c *HTTPAuthClient) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
      var resp struct {
          UserID string   `json:"user_id"`
          Roles  []string `json:"roles"`
          Scopes []string `json:"scopes"`
      }

      _, err := c.client.R().
          SetContext(ctx).
          SetHeader("Authorization", token).
          SetResult(&resp).
          Get(c.authURL + "/auth/validate")

      if err != nil {
          return nil, err
      }

      return &TokenClaims{
          UserID: resp.UserID,
          Roles:  resp.Roles,
          Scopes: resp.Scopes,
      }, nil
  }
  ```
- **Localização**: `internal/infrastructure/clients/auth_http_client.go`

---

## 4. Estrutura de Diretórios

```
core-dict/
├── cmd/
│   └── api/
│       └── main.go                      # Entrypoint
├── internal/
│   ├── api/                             # API Layer
│   │   └── http/
│   │       ├── router.go
│   │       ├── middleware/
│   │       │   ├── auth.go
│   │       │   ├── rbac.go
│   │       │   └── logger.go
│   │       ├── controllers/
│   │       │   ├── entry_controller.go
│   │       │   ├── claim_controller.go
│   │       │   └── health_controller.go
│   │       └── validators/
│   │           └── request_validator.go
│   ├── application/                     # Application Layer
│   │   ├── services/
│   │   │   ├── entry_service.go
│   │   │   └── claim_service.go
│   │   ├── mappers/
│   │   │   └── dto_mapper.go
│   │   └── ports/
│   │       └── event_publisher.go       # Interface
│   ├── domain/                          # Domain Layer
│   │   ├── entities/
│   │   │   ├── entry.go
│   │   │   ├── claim.go
│   │   │   └── account.go
│   │   ├── services/
│   │   │   ├── key_validator.go
│   │   │   └── claim_rules.go
│   │   ├── events/
│   │   │   └── events.go
│   │   └── repositories/                # Interfaces
│   │       ├── entry_repository.go
│   │       └── claim_repository.go
│   └── infrastructure/                  # Infrastructure Layer
│       ├── repositories/
│       │   ├── postgres_entry_repository.go
│       │   ├── postgres_claim_repository.go
│       │   └── postgres_audit_repository.go
│       ├── messaging/
│       │   └── pulsar_producer.go
│       └── clients/
│           ├── ledger_grpc_client.go
│           └── auth_http_client.go
├── go.mod
└── go.sum
```

---

## 5. Fluxo de Requisição Completo

### Exemplo: POST /api/v1/keys (CreateEntry)

```
1. HTTP Request
   ↓
2. HTTP Router (Fiber)
   ↓
3. Auth Middleware
   ├→ Auth HTTP Client → LBPay Auth (valida JWT)
   └→ Extrai user_id, roles, scopes
   ↓
4. RBAC Middleware
   └→ Verifica scope "dict:write"
   ↓
5. Entry Controller
   ├→ Request Validator (valida payload JSON)
   └→ Entry Service.CreateEntry(dto)
       ↓
6. Entry Service (Application Layer)
   ├→ DTO Mapper (DTO → Entity)
   ├→ Key Validator (valida CPF/CNPJ/etc.)
   ├→ Ledger gRPC Client (valida conta CID)
   ├→ Entry Repository.ExistsByKey() (verifica duplicata)
   ├→ Entry Entity.New() (cria entity com regras de domínio)
   ├→ Entry Repository.Create() (persiste no PostgreSQL)
   ├→ Audit Repository.Log() (auditoria)
   ├→ Event Publisher.Publish("dict.entries.created", event)
   │   └→ Pulsar Producer → Apache Pulsar
   └→ DTO Mapper (Entity → DTO)
   ↓
7. Entry Controller
   └→ HTTP Response 201 Created
```

**Duração**: ~150ms (sem Bacen, que é assíncrono via Pulsar)

---

## 6. Testes por Camada

### 6.1. API Layer - Integration Tests

```go
func TestCreateEntry_Success(t *testing.T) {
    app := setupTestApp()

    reqBody := `{
        "key_type": "CPF",
        "key_value": "12345678900",
        "account": {
            "ispb": "12345678",
            "account_number": "123456",
            "branch": "0001",
            "account_type": "CACC"
        }
    }`

    req := httptest.NewRequest("POST", "/api/v1/keys", strings.NewReader(reqBody))
    req.Header.Set("Authorization", "Bearer valid_jwt_token")
    req.Header.Set("Content-Type", "application/json")

    resp, _ := app.Test(req)

    assert.Equal(t, 201, resp.StatusCode)
}
```

### 6.2. Application Layer - Unit Tests

```go
func TestEntryService_CreateEntry_Success(t *testing.T) {
    // Arrange
    mockEntryRepo := new(MockEntryRepository)
    mockLedgerClient := new(MockLedgerClient)
    mockEventPublisher := new(MockEventPublisher)

    service := NewEntryService(mockEntryRepo, mockLedgerClient, mockEventPublisher)

    mockLedgerClient.On("ValidateAccount", mock.Anything, "account_id").
        Return(&AccountInfo{Exists: true, Status: "ACTIVE"}, nil)
    mockEntryRepo.On("ExistsByKey", mock.Anything, KeyTypeCPF, "12345678900").
        Return(false, nil)
    mockEntryRepo.On("Create", mock.Anything, mock.Anything).
        Return(nil)
    mockEventPublisher.On("Publish", mock.Anything, "dict.entries.created", mock.Anything).
        Return(nil)

    // Act
    entry, err := service.CreateEntry(context.Background(), "user_id", CreateEntryRequest{...})

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, entry)
    mockEntryRepo.AssertExpectations(t)
}
```

### 6.3. Domain Layer - Unit Tests

```go
func TestKeyValidator_ValidateCPF_Invalid(t *testing.T) {
    validator := NewKeyValidator()

    err := validator.Validate(KeyTypeCPF, "00000000000")  // CPF inválido

    assert.Error(t, err)
    assert.Equal(t, ErrInvalidCPF, err)
}

func TestClaim_Confirm_Success(t *testing.T) {
    claim := &Claim{Status: ClaimStatusOpen}

    err := claim.Confirm()

    assert.NoError(t, err)
    assert.Equal(t, ClaimStatusConfirmed, claim.Status)
}
```

### 6.4. Infrastructure Layer - Integration Tests

```go
func TestPostgresEntryRepository_Create_Success(t *testing.T) {
    // Setup test database
    pool := setupTestDatabase(t)
    repo := NewPostgresEntryRepository(pool)

    entry := &Entry{
        ID:       uuid.New(),
        KeyType:  KeyTypeCPF,
        KeyValue: "12345678900",
        Account:  testAccount,
        Status:   EntryStatusPending,
    }

    err := repo.Create(context.Background(), entry)

    assert.NoError(t, err)

    // Verify
    found, _ := repo.FindByID(context.Background(), entry.ID)
    assert.Equal(t, entry.KeyValue, found.KeyValue)
}
```

---

## 7. Próximos Passos

1. **[DIA-004: C4 Component Diagram - RSFN Connect](./DIA-004_C4_Component_Diagram_Connect.md)** (a criar)
   - Componentes do Temporal Worker e Pulsar Consumer

2. **[DIA-005: C4 Component Diagram - RSFN Bridge](./DIA-005_C4_Component_Diagram_Bridge.md)** (a criar)
   - Componentes do Bridge (SOAP Adapter, XML Signer)

3. **[API-002: Core DICT REST API](../../04_APIs/REST/API-002_Core_DICT_REST_API.md)** (a criar)
   - Especificação completa da API REST

4. **[IMP-001: Manual Implementação Core DICT](../../09_Implementacao/IMP-001_Manual_Implementacao_Core_DICT.md)** (a criar)
   - Guia de implementação passo a passo

---

## 8. Checklist de Validação

- [ ] Clean Architecture está clara (4 camadas)?
- [ ] Regras de dependência estão respeitadas (camadas externas → internas)?
- [ ] Componentes têm responsabilidades únicas (SRP)?
- [ ] Domain Layer não tem dependências externas?
- [ ] Interfaces de repositories estão no Domain Layer?
- [ ] Implementações de repositories estão no Infrastructure Layer?
- [ ] DTOs são usados para desacoplar API de Domain?
- [ ] Eventos de domínio são publicados corretamente?
- [ ] Validações de negócio estão no Domain Layer?
- [ ] Validações de entrada estão no API Layer?

---

## 9. Referências

### Documentos Internos
- [DIA-001: C4 Context Diagram](./DIA-001_C4_Context_Diagram.md)
- [DIA-002: C4 Container Diagram](./DIA-002_C4_Container_Diagram.md)
- [ADR-001: Clean Architecture](../ADRs/ADR-001_Clean_Architecture.md)
- [TEC-001: IcePanel Architecture](../../11_Especificacoes_Tecnicas/TEC-001_IcePanel_Architecture_and_Decisions.md)
- [DAT-001: Core Database Schema](../../03_Dados/DAT-001_Schema_Database_Core_DICT.md)

### Documentos Externos
- [Clean Architecture - Uncle Bob](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [C4 Model - Component Diagram](https://c4model.com/#ComponentDiagram)
- [Domain-Driven Design - Eric Evans](https://domainlanguage.com/ddd/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

---

**Última Revisão**: 2025-10-25
**Aprovado por**: Arquitetura LBPay
**Próxima Revisão**: 2026-01-25 (trimestral)
