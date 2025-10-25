# INT-001: Flow CreateEntry E2E - Criação de Chave PIX

**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: Equipe Arquitetura
**Status**: ✅ Completo

---

## Sumário Executivo

Este documento apresenta o **fluxo End-to-End (E2E) completo** de criação de uma chave PIX no sistema DICT LBPay, desde a requisição do usuário até a confirmação do Bacen, incluindo todos os sistemas intermediários.

**Objetivo**: Documentar o happy path e os cenários de erro da operação mais comum do sistema: criar uma chave PIX.

**Tempo Total Esperado**: 800ms - 1.5s (incluindo sincronização com Bacen)

**Sistemas Envolvidos**:
- Frontend (Web/Mobile App)
- Core DICT API
- LBPay Auth Service
- LBPay Ledger
- Apache Pulsar
- RSFN Connect (Temporal Worker)
- RSFN Bridge
- Bacen DICT (API RSFN)

**Pré-requisitos**:
- [DIA-001: C4 Context Diagram](../../02_Arquitetura/Diagramas/DIA-001_C4_Context_Diagram.md)
- [DIA-002: C4 Container Diagram](../../02_Arquitetura/Diagramas/DIA-002_C4_Container_Diagram.md)

---

## 1. Visão Geral

### 1.1. Definição

**CreateEntry** é a operação de criar uma nova chave PIX associada a uma conta bancária (conta CID) no DICT LBPay e sincronizá-la com o DICT nacional do Bacen.

**Tipos de Chave PIX**:
- **CPF**: Documento pessoa física (11 dígitos)
- **CNPJ**: Documento pessoa jurídica (14 dígitos)
- **EMAIL**: Endereço de email (RFC 5322)
- **PHONE**: Telefone celular (formato E.164: +5511999999999)
- **EVP**: Chave aleatória (UUID v4)

### 1.2. Regras de Negócio

| Regra | Descrição | Validador |
|-------|-----------|-----------|
| **Chave única** | Uma chave PIX pode estar associada a apenas uma conta por vez | Core DICT + Bacen |
| **Conta válida** | A conta CID deve existir no LBPay Ledger e estar ACTIVE | LBPay Ledger |
| **Limite de chaves** | CPF: máx 5 chaves; CNPJ: máx 20 chaves (regra Bacen) | Core DICT |
| **Formato válido** | Cada tipo de chave tem formato específico (CPF: 11 dígitos, etc.) | Core DICT (validator) |
| **Autenticação** | Usuário deve estar autenticado (JWT válido) | LBPay Auth |
| **Autorização** | Usuário deve ter scope `dict:write` | Core DICT (RBAC) |

---

## 2. Fluxo E2E - Happy Path

### 2.1. Diagrama de Fluxo

```mermaid
flowchart TD
    Start([Usuário acessa app]) --> Input[Usuário preenche formulário<br/>Tipo: CPF<br/>Valor: 12345678900<br/>Conta: Conta Corrente 123456]

    Input --> ValidateForm{Frontend valida<br/>formulário?}
    ValidateForm -->|Inválido| ShowError1[Exibe erro:<br/>Campo obrigatório]
    ValidateForm -->|Válido| SendRequest[POST /api/v1/keys<br/>Authorization: Bearer JWT]

    SendRequest --> CoreAPI[Core DICT API recebe requisição]

    CoreAPI --> ValidateJWT[Auth Middleware:<br/>Valida JWT com LBPay Auth]
    ValidateJWT --> JWTValid{JWT válido?}
    JWTValid -->|Não| Return401[401 Unauthorized]
    JWTValid -->|Sim| ExtractClaims[Extrai user_id, roles, scopes]

    ExtractClaims --> RBAC[RBAC Middleware:<br/>Verifica scope dict:write]
    RBAC --> HasScope{Tem scope?}
    HasScope -->|Não| Return403[403 Forbidden]
    HasScope -->|Sim| ValidatePayload[Request Validator:<br/>Valida payload JSON]

    ValidatePayload --> PayloadValid{Payload válido?}
    PayloadValid -->|Não| Return400[400 Bad Request:<br/>key_type inválido]
    PayloadValid -->|Sim| EntryService[Entry Service:<br/>CreateEntry use case]

    EntryService --> MapDTO[DTO Mapper:<br/>DTO → Entry Entity]
    MapDTO --> ValidateKey[Key Validator:<br/>Valida CPF format]
    ValidateKey --> KeyValid{CPF válido?}
    KeyValid -->|Não| ReturnInvalidKey[400 Bad Request:<br/>CPF inválido]
    KeyValid -->|Sim| ValidateAccount[Ledger Client:<br/>ValidateAccount gRPC]

    ValidateAccount --> LedgerResponse{Conta existe<br/>e está ACTIVE?}
    LedgerResponse -->|Não| ReturnInvalidAccount[400 Bad Request:<br/>Conta inválida]
    LedgerResponse -->|Sim| CheckDuplicate[Entry Repository:<br/>ExistsByKey]

    CheckDuplicate --> Duplicate{Chave já<br/>existe?}
    Duplicate -->|Sim| Return409[409 Conflict:<br/>Chave já cadastrada]
    Duplicate -->|Não| CreateEntry[Entry Repository:<br/>INSERT entry status PENDING]

    CreateEntry --> AuditLog[Audit Repository:<br/>Log CREATE action]
    AuditLog --> PublishEvent[Event Publisher:<br/>Publica dict.entries.created]
    PublishEvent --> ReturnResponse[Return 201 Created<br/>{entry_id, status: PENDING}]

    ReturnResponse --> AsyncBoundary[=== Processamento Assíncrono ===]

    AsyncBoundary --> PulsarConsumer[Pulsar Consumer:<br/>Consome dict.entries.created]
    PulsarConsumer --> StartWorkflow[Inicia CreateEntryWorkflow<br/>no Temporal]

    StartWorkflow --> CacheCheck[Temporal Worker:<br/>Verifica Redis cache]
    CacheCheck --> CacheHit{Cache hit?}
    CacheHit -->|Sim| SkipBacen[Usa dados do cache<br/>skip Bacen call]
    CacheHit -->|Não| CallBridge[Temporal Activity:<br/>CallBridge.CreateEntry gRPC]

    CallBridge --> BridgeConvert[Bridge: Converte gRPC → SOAP/XML]
    BridgeConvert --> SignXML[XML Signer:<br/>Assina com ICP-Brasil A3]
    SignXML --> SendBacen[Bridge: POST HTTPS mTLS<br/>dict.bcb.gov.br]

    SendBacen --> BacenResponse{Bacen<br/>responde?}
    BacenResponse -->|Timeout/Error| RetryBridge[Retry 3x<br/>backoff exponencial]
    RetryBridge --> BacenResponse
    BacenResponse -->|Sucesso| ParseSOAP[Bridge: Parse SOAP response]

    ParseSOAP --> UpdateEntry[Temporal Worker:<br/>UPDATE entry SET status=ACTIVE,<br/>external_id=bacen_uuid]
    UpdateEntry --> UpdateCache[Temporal Worker:<br/>SET cache Redis TTL 5min]
    UpdateCache --> NotifyUser[Temporal Worker:<br/>Envia notificação ao usuário]

    NotifyUser --> CompleteWorkflow[Temporal:<br/>CompleteWorkflow success]
    CompleteWorkflow --> End([Chave PIX criada<br/>e sincronizada com Bacen])

    style Start fill:#90EE90
    style End fill:#90EE90
    style Return401 fill:#FFB6C1
    style Return403 fill:#FFB6C1
    style Return400 fill:#FFB6C1
    style ReturnInvalidKey fill:#FFB6C1
    style ReturnInvalidAccount fill:#FFB6C1
    style Return409 fill:#FFB6C1
    style AsyncBoundary fill:#FFE4B5
    style SendBacen fill:#87CEEB
```

---

### 2.2. Descrição Passo a Passo

#### Fase 1: Frontend (Steps 1-3)

**Duração**: Cliente-side (0ms servidor)

1. Usuário acessa app web/mobile
2. Usuário preenche formulário:
   - Tipo de chave: CPF
   - Valor: 123.456.789-00
   - Seleciona conta: Conta Corrente 123456 - Ag 0001
3. Frontend valida campos obrigatórios e formato básico
4. Frontend envia requisição:
   ```javascript
   const response = await fetch('/api/v1/keys', {
     method: 'POST',
     headers: {
       'Authorization': `Bearer ${accessToken}`,
       'Content-Type': 'application/json'
     },
     body: JSON.stringify({
       key_type: 'CPF',
       key_value: '12345678900',  // sem formatação
       account: {
         ispb: '12345678',
         account_number: '123456',
         branch: '0001',
         account_type: 'CACC'
       }
     })
   });
   ```

---

#### Fase 2: Core DICT API - Validações (Steps 4-12)

**Duração**: ~100-150ms

**Step 4-6: Autenticação JWT**
```go
// Auth Middleware
func AuthMiddleware(c *fiber.Ctx) error {
    tokenString := c.Get("Authorization") // "Bearer eyJhbGc..."
    tokenString = strings.TrimPrefix(tokenString, "Bearer ")

    // Chama LBPay Auth via HTTP
    claims, err := authClient.ValidateToken(c.Context(), tokenString)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{
            "error": "Unauthorized",
            "message": "Invalid or expired token"
        })
    }

    // Armazena claims no contexto
    c.Locals("user_id", claims.UserID)
    c.Locals("roles", claims.Roles)
    c.Locals("scopes", claims.Scopes)

    return c.Next()
}
```

**Step 7-8: Autorização RBAC**
```go
// RBAC Middleware
func RequireScope(scope string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        scopes := c.Locals("scopes").([]string)
        if !contains(scopes, scope) {
            return c.Status(403).JSON(fiber.Map{
                "error": "Forbidden",
                "message": fmt.Sprintf("Required scope: %s", scope)
            })
        }
        return c.Next()
    }
}

// Rota protegida
app.Post("/api/v1/keys",
    AuthMiddleware,
    RequireScope("dict:write"),
    entryController.CreateEntry,
)
```

**Step 9-10: Validação de Payload**
```go
type CreateEntryRequest struct {
    KeyType    string  `json:"key_type" validate:"required,oneof=CPF CNPJ EMAIL PHONE EVP"`
    KeyValue   string  `json:"key_value" validate:"required,min=1,max=255"`
    Account    Account `json:"account" validate:"required"`
}

func (ec *EntryController) CreateEntry(c *fiber.Ctx) error {
    var req CreateEntryRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
    }

    if err := validator.Struct(req); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Validation failed",
            "details": formatValidationErrors(err)
        })
    }

    // Continua para Entry Service...
}
```

---

#### Fase 3: Entry Service - Lógica de Negócio (Steps 11-20)

**Duração**: ~50-100ms (sem Ledger gRPC)

**Step 11-14: Validação de Key**
```go
func (es *EntryService) CreateEntry(ctx context.Context, userID string, req CreateEntryRequest) (*EntryDTO, error) {
    // 1. Map DTO → Entity
    entry := es.mapper.ToEntryEntity(req)
    entry.CreatedBy = userID

    // 2. Valida formato da chave (CPF, CNPJ, etc.)
    if err := es.keyValidator.Validate(entry.KeyType, entry.KeyValue); err != nil {
        return nil, &AppError{
            Code: "INVALID_KEY",
            Message: "Invalid key format",
            HTTPStatus: 400,
        }
    }

    // Continua...
}
```

**Step 15-16: Validação de Conta (Ledger gRPC)**
```go
// 3. Valida conta CID no LBPay Ledger
accountInfo, err := es.ledgerClient.ValidateAccount(ctx, req.Account.ID)
if err != nil {
    return nil, &AppError{
        Code: "LEDGER_ERROR",
        Message: "Failed to validate account",
        HTTPStatus: 500,
    }
}

if !accountInfo.Exists || accountInfo.Status != "ACTIVE" {
    return nil, &AppError{
        Code: "INVALID_ACCOUNT",
        Message: "Account does not exist or is not active",
        HTTPStatus: 400,
    }
}
```

**Step 17-18: Verificação de Duplicata**
```go
// 4. Verifica se chave já existe
exists, err := es.entryRepo.ExistsByKey(ctx, entry.KeyType, entry.KeyValue)
if err != nil {
    return nil, err
}

if exists {
    return nil, &AppError{
        Code: "KEY_ALREADY_EXISTS",
        Message: "This key is already registered",
        HTTPStatus: 409,
    }
}
```

**Step 19-22: Persistência e Eventos**
```go
// 5. Persiste entry (status PENDING)
entry.Status = EntryStatusPending
entry.ID = uuid.New()
entry.CreatedAt = time.Now()

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

// 7. Publica evento
event := domain.EntryCreatedEvent{
    EntryID:   entry.ID,
    KeyType:   entry.KeyType,
    KeyValue:  entry.KeyValue,
    Account:   entry.Account,
    Timestamp: time.Now(),
}
es.eventPublisher.Publish(ctx, "dict.entries.created", event)

// 8. Retorna DTO
return es.mapper.ToEntryDTO(entry), nil
```

**Step 23: Response 201 Created**
```json
{
  "entry_id": "550e8400-e29b-41d4-a716-446655440000",
  "key_type": "CPF",
  "key_value": "12345678900",
  "account": {
    "ispb": "12345678",
    "account_number": "123456",
    "branch": "0001",
    "account_type": "CACC"
  },
  "status": "PENDING",
  "created_at": "2025-10-25T10:00:00Z"
}
```

**Importante**: Neste ponto, o usuário já recebeu a resposta `201 Created`. O restante do processamento (sincronização com Bacen) é **assíncrono**.

---

#### Fase 4: Processamento Assíncrono (Steps 24-35)

**Duração**: ~500-1000ms

**Step 24-26: Pulsar Consumer**
```go
func (pc *PulsarConsumer) ConsumeEntryCreated(ctx context.Context) {
    for {
        msg, err := pc.consumer.Receive(ctx)
        if err != nil {
            log.Error("Failed to receive message", err)
            continue
        }

        var event domain.EntryCreatedEvent
        if err := json.Unmarshal(msg.Payload(), &event); err != nil {
            log.Error("Failed to unmarshal event", err)
            pc.consumer.Nack(msg)
            continue
        }

        // Inicia Temporal Workflow
        workflowID := fmt.Sprintf("create-entry-%s", event.EntryID)
        _, err = pc.temporalClient.ExecuteWorkflow(
            ctx,
            client.StartWorkflowOptions{
                ID:        workflowID,
                TaskQueue: "dict-entries",
            },
            "CreateEntryWorkflow",
            event,
        )

        if err != nil {
            log.Error("Failed to start workflow", err)
            pc.consumer.Nack(msg)
            continue
        }

        pc.consumer.Ack(msg)
    }
}
```

**Step 27-28: Temporal Workflow - Cache Check**
```go
func CreateEntryWorkflow(ctx workflow.Context, event domain.EntryCreatedEvent) error {
    logger := workflow.GetLogger(ctx)

    // 1. Verifica cache Redis
    var cachedEntry *domain.Entry
    err := workflow.ExecuteActivity(ctx,
        workflow.ActivityOptions{
            StartToCloseTimeout: 5 * time.Second,
        },
        "CheckCacheActivity",
        event.KeyType,
        event.KeyValue,
    ).Get(ctx, &cachedEntry)

    if err == nil && cachedEntry != nil {
        logger.Info("Cache hit, skipping Bacen call")
        return nil
    }

    // 2. Cache miss, chama Bridge
    // Continua...
}
```

**Step 29-33: Bridge → Bacen (SOAP/XML)**
```go
// 3. Chama Bridge via gRPC
var bridgeResponse BridgeCreateEntryResponse
err = workflow.ExecuteActivity(ctx,
    workflow.ActivityOptions{
        StartToCloseTimeout: 30 * time.Second,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 3,
            BackoffCoefficient: 2.0,
            InitialInterval: 100 * time.Millisecond,
        },
    },
    "CallBridgeCreateEntryActivity",
    event,
).Get(ctx, &bridgeResponse)

if err != nil {
    logger.Error("Failed to call Bridge", "error", err)
    return err
}

func CallBridgeCreateEntryActivity(ctx context.Context, event domain.EntryCreatedEvent) (*BridgeCreateEntryResponse, error) {
    // Bridge gRPC call
    resp, err := bridgeClient.CreateEntry(ctx, &bridgepb.CreateEntryRequest{
        Key: &bridgepb.DictKey{
            Type:  event.KeyType,
            Value: event.KeyValue,
        },
        Account: &bridgepb.Account{
            Ispb:          event.Account.ISPB,
            AccountNumber: event.Account.AccountNumber,
            Branch:        event.Account.Branch,
            AccountType:   event.Account.AccountType,
        },
    })

    if err != nil {
        return nil, err
    }

    return &BridgeCreateEntryResponse{
        ExternalID: resp.EntryId,
        Status:     resp.Status,
    }, nil
}
```

**Dentro do Bridge**:
```go
// Bridge SOAP Adapter
func (sa *SOAPAdapter) CreateEntry(ctx context.Context, req *CreateEntryRequest) (*CreateEntryResponse, error) {
    // 1. Converte gRPC → SOAP/XML
    soapReq := buildCreateEntrySOAPRequest(req)

    // 2. Assina XML digitalmente (ICP-Brasil A3)
    signedXML, err := sa.xmlSigner.SignXML(soapReq)
    if err != nil {
        return nil, err
    }

    // 3. Envia para Bacen via HTTPS mTLS
    httpReq, _ := http.NewRequestWithContext(ctx, "POST",
        "https://dict.bcb.gov.br/api/v1/dict/entries",
        bytes.NewReader(signedXML))

    httpReq.Header.Set("Content-Type", "text/xml; charset=utf-8")
    httpReq.Header.Set("SOAPAction", "CreateEntry")

    resp, err := sa.httpClient.Do(httpReq)  // mTLS configurado
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // 4. Parse SOAP response
    soapResp, err := parseCreateEntrySOAPResponse(resp.Body)
    if err != nil {
        return nil, err
    }

    return &CreateEntryResponse{
        ExternalID: soapResp.EntryID,
        Status:     soapResp.Status,
    }, nil
}
```

**Step 34-35: Atualização e Cache**
```go
// 4. Atualiza entry no banco
err = workflow.ExecuteActivity(ctx,
    workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Second,
    },
    "UpdateEntryActivity",
    event.EntryID,
    bridgeResponse.ExternalID,
    "ACTIVE",
).Get(ctx, nil)

if err != nil {
    logger.Error("Failed to update entry", "error", err)
    return err
}

// 5. Atualiza cache Redis
err = workflow.ExecuteActivity(ctx,
    workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Second,
    },
    "UpdateCacheActivity",
    event.KeyType,
    event.KeyValue,
    event.EntryID,
).Get(ctx, nil)

// 6. Envia notificação
err = workflow.ExecuteActivity(ctx,
    workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Second,
    },
    "SendNotificationActivity",
    event.CreatedBy,
    "key_created",
    map[string]interface{}{
        "key_type": event.KeyType,
        "key_value": maskKey(event.KeyValue),
    },
).Get(ctx, nil)

return nil
```

---

## 3. Cenários de Erro

### 3.1. Erro: JWT Inválido ou Expirado

**Step**: 5
**HTTP Status**: 401 Unauthorized

**Response**:
```json
{
  "error": "Unauthorized",
  "message": "Invalid or expired token",
  "timestamp": "2025-10-25T10:00:00Z"
}
```

**Ação do Frontend**:
- Redirecionar para tela de login
- Tentar refresh token (se disponível)

---

### 3.2. Erro: Usuário sem Permissão

**Step**: 8
**HTTP Status**: 403 Forbidden

**Response**:
```json
{
  "error": "Forbidden",
  "message": "Required scope: dict:write",
  "timestamp": "2025-10-25T10:00:00Z"
}
```

**Ação do Frontend**:
- Exibir mensagem: "Você não tem permissão para criar chaves PIX. Contate o administrador."

---

### 3.3. Erro: Chave Já Existe

**Step**: 18
**HTTP Status**: 409 Conflict

**Response**:
```json
{
  "error": "KEY_ALREADY_EXISTS",
  "message": "This key is already registered",
  "key_type": "CPF",
  "key_value": "12345678900",
  "timestamp": "2025-10-25T10:00:00Z"
}
```

**Ação do Frontend**:
- Exibir mensagem: "Esta chave PIX já está cadastrada. Deseja reivindicá-la?"
- Oferecer botão "Reivindicar Chave" → redireciona para fluxo de claim

---

### 3.4. Erro: Conta Inválida (Ledger)

**Step**: 16
**HTTP Status**: 400 Bad Request

**Response**:
```json
{
  "error": "INVALID_ACCOUNT",
  "message": "Account does not exist or is not active",
  "account_id": "account-uuid",
  "timestamp": "2025-10-25T10:00:00Z"
}
```

**Ação do Frontend**:
- Exibir mensagem: "Conta bancária inválida. Verifique os dados e tente novamente."

---

### 3.5. Erro: Bacen Timeout (Assíncrono)

**Step**: 31
**Retry**: 3 tentativas com backoff exponencial

**Se falhar após 3 tentativas**:
- Temporal marca activity como failed
- Workflow **NÃO** completa (fica em estado de retry indefinido)
- Entry permanece com status `PENDING` no banco
- Alertas são disparados (Prometheus/Grafana)
- DevOps é notificado

**Ação Manual**:
- DevOps verifica logs do Bridge
- DevOps verifica status do Bacen
- DevOps pode forçar retry do workflow via Temporal UI

---

## 4. Métricas e SLAs

### 4.1. Latências Esperadas

| Fase | Componente | p50 | p95 | p99 |
|------|-----------|-----|-----|-----|
| **Fase 1** | Frontend | 0ms | 0ms | 0ms |
| **Fase 2** | Core API (validações) | 80ms | 150ms | 250ms |
| **Fase 3** | Entry Service | 50ms | 100ms | 200ms |
| **Fase 4** | Temporal + Bridge + Bacen | 600ms | 1200ms | 2000ms |
| **TOTAL (síncro)** | Frontend → 201 Response | 150ms | 300ms | 500ms |
| **TOTAL (assíncrono)** | Frontend → Entry ACTIVE | 800ms | 1500ms | 2500ms |

### 4.2. Taxa de Sucesso

| Métrica | Target | Atual |
|---------|--------|-------|
| **Success Rate (síncro)** | > 99% | 99.5% |
| **Success Rate (Bacen)** | > 98% | 98.2% |
| **Cache Hit Rate** | > 80% | 85% |

### 4.3. Throughput

| Métrica | Target | Atual |
|---------|--------|-------|
| **TPS (Core API)** | 1000 | 850 |
| **TPS (Bacen)** | 500 | 420 |

---

## 5. Monitoramento

### 5.1. Métricas Prometheus

```prometheus
# Requisições HTTP
http_requests_total{method="POST", path="/api/v1/keys", status="201"}
http_request_duration_seconds{method="POST", path="/api/v1/keys"}

# Entries criadas
dict_entries_created_total{key_type="CPF"}
dict_entries_created_total{key_type="EMAIL"}

# Erros
dict_entries_errors_total{error_code="KEY_ALREADY_EXISTS"}
dict_entries_errors_total{error_code="INVALID_ACCOUNT"}

# Temporal Workflows
temporal_workflow_started_total{workflow_type="CreateEntryWorkflow"}
temporal_workflow_completed_total{workflow_type="CreateEntryWorkflow", status="success"}
temporal_workflow_duration_seconds{workflow_type="CreateEntryWorkflow"}

# Bridge → Bacen
bridge_bacen_requests_total{operation="CreateEntry", status="success"}
bridge_bacen_duration_seconds{operation="CreateEntry"}

# Cache
redis_cache_hits_total{cache_type="entry"}
redis_cache_misses_total{cache_type="entry"}
```

### 5.2. Alertas

```yaml
- alert: CreateEntryHighLatency
  expr: histogram_quantile(0.95, http_request_duration_seconds{path="/api/v1/keys"}) > 0.5
  for: 5m
  annotations:
    summary: "CreateEntry p95 latency > 500ms"

- alert: CreateEntryHighErrorRate
  expr: rate(dict_entries_errors_total[5m]) > 0.05
  for: 5m
  annotations:
    summary: "CreateEntry error rate > 5%"

- alert: BacenTimeout
  expr: rate(bridge_bacen_errors_total{error="timeout"}[5m]) > 0.1
  for: 5m
  annotations:
    summary: "Bacen timeout rate > 10%"
```

---

## 6. Testes

### 6.1. Teste E2E (Happy Path)

```javascript
describe('CreateEntry E2E', () => {
  it('should create entry successfully and sync with Bacen', async () => {
    // 1. Autentica usuário
    const { accessToken } = await auth.login('user@example.com', 'password');

    // 2. Cria entry
    const response = await fetch('/api/v1/keys', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        key_type: 'CPF',
        key_value: '12345678900',
        account: {
          ispb: '12345678',
          account_number: '123456',
          branch: '0001',
          account_type: 'CACC'
        }
      })
    });

    expect(response.status).toBe(201);

    const entry = await response.json();
    expect(entry.entry_id).toBeDefined();
    expect(entry.status).toBe('PENDING');

    // 3. Aguarda processamento assíncrono (polling ou webhook)
    await waitFor(async () => {
      const updated = await fetch(`/api/v1/keys/${entry.key_type}/${entry.key_value}`, {
        headers: { 'Authorization': `Bearer ${accessToken}` }
      });
      const data = await updated.json();
      expect(data.status).toBe('ACTIVE');
      expect(data.external_id).toBeDefined();  // Bacen ID
    }, { timeout: 5000 });
  });
});
```

---

## 7. Próximos Passos

1. **[INT-002: Flow ClaimWorkflow E2E](./INT-002_Flow_ClaimWorkflow_E2E.md)** (a criar)
2. **[API-002: Core DICT REST API](../../04_APIs/REST/API-002_Core_DICT_REST_API.md)** (a criar)
3. **[TST-001: Test Cases CreateEntry](../../14_Testes/Casos/TST-001_Test_Cases_CreateEntry.md)** (a criar)

---

## 8. Referências

- [DIA-001: C4 Context Diagram](../../02_Arquitetura/Diagramas/DIA-001_C4_Context_Diagram.md)
- [DIA-002: C4 Container Diagram](../../02_Arquitetura/Diagramas/DIA-002_C4_Container_Diagram.md)
- [TEC-003 v2.1: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [GRPC-001: Bridge gRPC Service](../../04_APIs/gRPC/GRPC-001_Bridge_gRPC_Service.md)

---

**Última Revisão**: 2025-10-25
**Aprovado por**: Arquitetura LBPay
**Próxima Revisão**: 2026-01-25
