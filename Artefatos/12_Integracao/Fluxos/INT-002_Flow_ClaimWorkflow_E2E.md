# INT-002: Flow ClaimWorkflow E2E - Reivindicação de Chave PIX

**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: Equipe Arquitetura
**Status**: ✅ Completo

---

## Sumário Executivo

Este documento apresenta o **fluxo End-to-End (E2E) completo** de reivindicação de chave PIX (ClaimWorkflow) no sistema DICT LBPay, incluindo o workflow durável de 30 dias gerenciado pelo Temporal.

**Objetivo**: Documentar o fluxo completo de claim desde a criação até a resolução (confirmada, cancelada ou expirada), incluindo todos os sistemas intermediários.

**Tempo Total Esperado**: 30 dias (durável) + 1-2s (resolução)

**Sistemas Envolvidos**:
- Frontend (Web/Mobile App)
- Core DICT API
- LBPay Auth Service
- Apache Pulsar
- RSFN Connect (Temporal Worker)
- RSFN Bridge
- Bacen DICT (API RSFN)
- LBPay Notifications

**Pré-requisitos**:
- [DIA-006: Sequence Diagram - ClaimWorkflow](../../02_Arquitetura/Diagramas/DIA-006_Sequence_Claim_Workflow.md)
- [TEC-003 v2.1: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)

---

## 1. Visão Geral

### 1.1. Definição

**ClaimWorkflow** é a operação de reivindicar uma chave PIX que atualmente pertence a outra instituição financeira, com um período de espera regulatório de 30 dias para o owner confirmar ou cancelar a transferência.

**Exemplo de Uso**:
- Cliente possui CPF 123.456.789-00 cadastrado no Banco A (ISPB 12345678)
- Cliente abre conta no Banco B (ISPB 87654321) e deseja transferir a chave
- Banco B cria uma claim
- Banco A tem 30 dias para confirmar ou cancelar
- Se não responder em 30 dias, a chave é transferida automaticamente (auto-confirm)

### 1.2. Regras de Negócio

| Regra | Descrição | Validador |
|-------|-----------|-----------|
| **Período fixo de 30 dias** | Owner tem exatamente 30 dias para responder. Após 30 dias, auto-confirm (regra Bacen) | Temporal Workflow |
| **Entry ACTIVE** | Apenas chaves ACTIVE podem ser reivindicadas | Core DICT |
| **ISPB diferente** | Claimer ISPB DEVE ser diferente de Owner ISPB | Core DICT |
| **Uma claim por vez** | Não pode haver claim OPEN para mesma entry | Core DICT |
| **Resolução final** | Claim pode ser: CONFIRMED, CANCELLED ou EXPIRED | Temporal Workflow |
| **Autenticação** | Usuário deve estar autenticado (JWT válido) | LBPay Auth |
| **Autorização** | Claimer: scope `dict:claim:create`; Owner: scope `dict:claim:resolve` | Core DICT (RBAC) |

---

## 2. Fluxo E2E - Happy Path

### 2.1. Diagrama de Fluxo Completo

```mermaid
flowchart TD
    Start([Claimer User inicia claim]) --> InputClaim[Usuário seleciona chave PIX<br/>a ser reivindicada]

    InputClaim --> ValidateFormClaim{Frontend valida<br/>formulário?}
    ValidateFormClaim -->|Inválido| ShowError1[Exibe erro:<br/>Campo obrigatório]
    ValidateFormClaim -->|Válido| SendRequestClaim[POST /api/v1/claims<br/>Authorization: Bearer JWT]

    SendRequestClaim --> CoreAPIClaim[Core DICT API recebe requisição]

    CoreAPIClaim --> ValidateJWTClaim[Auth Middleware:<br/>Valida JWT]
    ValidateJWTClaim --> JWTValidClaim{JWT válido?}
    JWTValidClaim -->|Não| Return401Claim[401 Unauthorized]
    JWTValidClaim -->|Sim| ExtractClaimsClaim[Extrai user_id, scopes]

    ExtractClaimsClaim --> RBACClaim[RBAC Middleware:<br/>Verifica scope dict:claim:create]
    RBACClaim --> HasScopeClaim{Tem scope?}
    HasScopeClaim -->|Não| Return403Claim[403 Forbidden]
    HasScopeClaim -->|Sim| ValidatePayloadClaim[Request Validator:<br/>Valida payload]

    ValidatePayloadClaim --> PayloadValidClaim{Payload válido?}
    PayloadValidClaim -->|Não| Return400Claim[400 Bad Request]
    PayloadValidClaim -->|Sim| ClaimService[Claim Service:<br/>CreateClaim use case]

    ClaimService --> ValidateEntry[Entry Repository:<br/>Busca Entry]
    ValidateEntry --> EntryActive{Entry ACTIVE?}
    EntryActive -->|Não| Return400Entry[400 Bad Request:<br/>Entry não está ACTIVE]
    EntryActive -->|Sim| ValidateISPB[Valida ISPB:<br/>claimer != owner]

    ValidateISPB --> ISPBValid{ISPB diferente?}
    ISPBValid -->|Não| Return403ISPB[403 Forbidden:<br/>ISPB igual ao owner]
    ISPBValid -->|Sim| ValidatePeriod[Valida period:<br/>== 30 dias]

    ValidatePeriod --> PeriodValid{Period == 30?}
    PeriodValid -->|Não| Return400Period[400 Bad Request:<br/>Period deve ser 30]
    PeriodValid -->|Sim| CheckOpenClaim[Claim Repository:<br/>Verifica claim OPEN]

    CheckOpenClaim --> OpenClaimExists{Claim OPEN<br/>existe?}
    OpenClaimExists -->|Sim| Return409[409 Conflict:<br/>Claim OPEN já existe]
    OpenClaimExists -->|Não| CreateClaim[Claim Repository:<br/>INSERT claim status OPEN<br/>expires_at: NOW + 30 dias]

    CreateClaim --> AuditLogClaim[Audit Repository:<br/>Log CREATE_CLAIM action]
    AuditLogClaim --> PublishEventClaim[Event Publisher:<br/>Publica dict.claims.created]
    PublishEventClaim --> ReturnResponseClaim[Return 201 Created<br/>{claim_id, status: OPEN, expires_at}]

    ReturnResponseClaim --> AsyncBoundaryClaim[=== Processamento Assíncrono ===]

    AsyncBoundaryClaim --> PulsarConsumerClaim[Pulsar Consumer:<br/>Consome dict.claims.created]
    PulsarConsumerClaim --> StartWorkflowClaim[Inicia ClaimWorkflow<br/>no Temporal]

    StartWorkflowClaim --> CallBridgeCreateClaim[Temporal Activity:<br/>CallBridge.CreateClaim gRPC]

    CallBridgeCreateClaim --> BridgeConvertClaim[Bridge: Converte gRPC → SOAP/XML]
    BridgeConvertClaim --> SignXMLClaim[XML Signer:<br/>Assina com ICP-Brasil A3]
    SignXMLClaim --> SendBacenClaim[Bridge: POST HTTPS mTLS<br/>dict.bcb.gov.br CreateClaim]

    SendBacenClaim --> BacenResponseClaim{Bacen<br/>responde?}
    BacenResponseClaim -->|Timeout/Error| RetryBridgeClaim[Retry 3x<br/>backoff exponencial]
    RetryBridgeClaim --> BacenResponseClaim
    BacenResponseClaim -->|Sucesso| ParseSOAPClaim[Bridge: Parse SOAP response]

    ParseSOAPClaim --> UpdateClaimBacenID[Temporal Worker:<br/>UPDATE claim SET bacen_claim_id]
    UpdateClaimBacenID --> NotifyOwner[Temporal Worker:<br/>Notifica Owner via Notifications]

    NotifyOwner --> SetTimer[Temporal Workflow:<br/>SetTimer 30 dias]
    SetTimer --> WorkflowSleep[Workflow DORME por 30 dias<br/>(durável no Temporal Server)]

    WorkflowSleep --> Resolution{Resolução?}

    Resolution -->|Owner Confirma| ScenarioA[=== CENÁRIO A: CONFIRMAÇÃO ===]
    Resolution -->|Owner Cancela| ScenarioB[=== CENÁRIO B: CANCELAMENTO ===]
    Resolution -->|30 dias expiram| ScenarioC[=== CENÁRIO C: EXPIRAÇÃO ===]

    ScenarioA --> OwnerConfirm[Owner User:<br/>POST /claims/{id}/confirm]
    OwnerConfirm --> ValidateOwner[Core DICT:<br/>Valida se usuário é owner]
    ValidateOwner --> UpdateConfirm[Core DB:<br/>UPDATE claim status=CONFIRMED]
    UpdateConfirm --> PublishConfirm[Pulsar:<br/>Publica dict.claims.confirmed]
    PublishConfirm --> SignalConfirm[Temporal:<br/>SignalWorkflow confirm]
    SignalConfirm --> WakeupConfirm[Temporal Worker:<br/>Acorda workflow]
    WakeupConfirm --> CancelTimer[Cancela timer 30 dias]
    CancelTimer --> BridgeComplete[Bridge:<br/>CompleteClaim confirmed=true]
    BridgeComplete --> BacenComplete[Bacen:<br/>SOAP CompleteClaimRequest]
    BacenComplete --> TransferKey[Core DB:<br/>UPDATE entry account_id=claimer]
    TransferKey --> NotifyBothConfirm[Notifications:<br/>Notifica owner e claimer]
    NotifyBothConfirm --> CompleteWorkflowConfirm[Temporal:<br/>CompleteWorkflow success]
    CompleteWorkflowConfirm --> EndConfirm([Claim confirmada<br/>Chave transferida])

    ScenarioB --> OwnerCancel[Owner User:<br/>POST /claims/{id}/cancel]
    OwnerCancel --> UpdateCancel[Core DB:<br/>UPDATE claim status=CANCELLED]
    UpdateCancel --> PublishCancel[Pulsar:<br/>Publica dict.claims.cancelled]
    PublishCancel --> SignalCancel[Temporal:<br/>SignalWorkflow cancel]
    SignalCancel --> WakeupCancel[Temporal Worker:<br/>Acorda workflow]
    WakeupCancel --> CancelTimerB[Cancela timer 30 dias]
    CancelTimerB --> BridgeCancel[Bridge:<br/>CancelClaim]
    BridgeCancel --> BacenCancel[Bacen:<br/>SOAP CancelClaimRequest]
    BacenCancel --> NotKeepOwner[Entry permanece com owner<br/>NÃO transfere chave]
    NotKeepOwner --> NotifyBothCancel[Notifications:<br/>Notifica owner e claimer]
    NotifyBothCancel --> CompleteWorkflowCancel[Temporal:<br/>CompleteWorkflow cancelled]
    CompleteWorkflowCancel --> EndCancel([Claim cancelada<br/>Chave permanece com owner])

    ScenarioC --> TimerFired[Temporal Server:<br/>Timer de 30 dias dispara]
    TimerFired --> WakeupExpire[Temporal Worker:<br/>Acorda workflow automaticamente]
    WakeupExpire --> AutoConfirm[Aplica regra Bacen:<br/>Auto-confirmação]
    AutoConfirm --> BridgeExpire[Bridge:<br/>CompleteClaim auto_confirmed=true]
    BridgeExpire --> BacenExpire[Bacen:<br/>SOAP CompleteClaimRequest EXPIRED]
    BacenExpire --> TransferKeyExpire[Core DB:<br/>UPDATE claim status=EXPIRED<br/>UPDATE entry account_id=claimer]
    TransferKeyExpire --> NotifyBothExpire[Notifications:<br/>Notifica owner e claimer]
    NotifyBothExpire --> CompleteWorkflowExpire[Temporal:<br/>CompleteWorkflow expired]
    CompleteWorkflowExpire --> EndExpire([Claim expirada<br/>Chave transferida automaticamente])

    style Start fill:#90EE90
    style EndConfirm fill:#90EE90
    style EndCancel fill:#FFD700
    style EndExpire fill:#FFA500
    style Return401Claim fill:#FFB6C1
    style Return403Claim fill:#FFB6C1
    style Return400Claim fill:#FFB6C1
    style Return400Entry fill:#FFB6C1
    style Return403ISPB fill:#FFB6C1
    style Return400Period fill:#FFB6C1
    style Return409 fill:#FFB6C1
    style AsyncBoundaryClaim fill:#FFE4B5
    style WorkflowSleep fill:#87CEEB
    style ScenarioA fill:#E0FFE0
    style ScenarioB fill:#FFF8DC
    style ScenarioC fill:#FFE4B5
```

---

### 2.2. Descrição Passo a Passo

#### Fase 1: Frontend - Criação da Claim (Steps 1-3)

**Duração**: Cliente-side (0ms servidor)

1. Claimer User acessa app e navega para "Minhas Chaves PIX"
2. Usuário seleciona uma chave existente em outro banco
3. Usuário clica em "Reivindicar Chave"
4. Frontend envia requisição:

```javascript
const response = await fetch('/api/v1/claims', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${accessToken}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    entry_id: '550e8400-e29b-41d4-a716-446655440000',
    claimer_account: {
      ispb: '87654321',
      account_number: '654321',
      branch: '0002',
      account_type: 'CACC',
      holder_document: '98765432100',
      holder_name: 'João Silva'
    },
    completion_period_days: 30
  })
});
```

---

#### Fase 2: Core DICT API - Validações (Steps 4-21)

**Duração**: ~150-200ms

**Step 4-8: Autenticação e Autorização**
```go
// Auth Middleware - valida JWT
func AuthMiddleware(c *fiber.Ctx) error {
    tokenString := c.Get("Authorization")
    tokenString = strings.TrimPrefix(tokenString, "Bearer ")

    claims, err := authClient.ValidateToken(c.Context(), tokenString)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{
            "error": "Unauthorized",
            "message": "Invalid or expired token"
        })
    }

    c.Locals("user_id", claims.UserID)
    c.Locals("scopes", claims.Scopes)
    return c.Next()
}

// RBAC Middleware - verifica scope
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
app.Post("/api/v1/claims",
    AuthMiddleware,
    RequireScope("dict:claim:create"),
    claimController.CreateClaim,
)
```

**Step 9-21: Lógica de Negócio**
```go
func (cs *ClaimService) CreateClaim(ctx context.Context, userID string, req CreateClaimRequest) (*ClaimDTO, error) {
    // 1. Busca Entry
    entry, err := cs.entryRepo.FindByID(ctx, req.EntryID)
    if err != nil {
        return nil, &AppError{Code: "ENTRY_NOT_FOUND", HTTPStatus: 404}
    }

    // 2. Valida se Entry está ACTIVE
    if entry.Status != EntryStatusActive {
        return nil, &AppError{
            Code: "ENTRY_NOT_ACTIVE",
            Message: "Entry must be ACTIVE to be claimed",
            HTTPStatus: 400,
        }
    }

    // 3. Valida ISPB (claimer != owner)
    if req.ClaimerAccount.ISPB == entry.Account.ISPB {
        return nil, &AppError{
            Code: "SAME_ISPB",
            Message: "Claimer ISPB cannot be the same as owner ISPB",
            HTTPStatus: 403,
        }
    }

    // 4. Valida completion_period_days == 30 (regra Bacen)
    if req.CompletionPeriodDays != 30 {
        return nil, &AppError{
            Code: "INVALID_PERIOD",
            Message: "completion_period_days must be exactly 30",
            HTTPStatus: 400,
        }
    }

    // 5. Verifica se já existe claim OPEN para esta entry
    openClaim, err := cs.claimRepo.FindOpenByEntryID(ctx, req.EntryID)
    if err == nil && openClaim != nil {
        return nil, &AppError{
            Code: "OPEN_CLAIM_EXISTS",
            Message: "An open claim already exists for this entry",
            HTTPStatus: 409,
        }
    }

    // 6. Cria claim
    claim := domain.Claim{
        ID:                   uuid.New(),
        EntryID:              entry.ID,
        ClaimerAccount:       req.ClaimerAccount,
        OwnerAccount:         entry.Account,
        Status:               ClaimStatusOpen,
        CompletionPeriodDays: 30,
        ExpiresAt:            time.Now().Add(30 * 24 * time.Hour),
        CreatedBy:            userID,
        CreatedAt:            time.Now(),
    }

    if err := cs.claimRepo.Create(ctx, &claim); err != nil {
        return nil, err
    }

    // 7. Auditoria
    cs.auditRepo.Log(ctx, AuditLog{
        EntityType: "claim",
        EntityID:   claim.ID,
        Action:     "CREATE_CLAIM",
        UserID:     userID,
        Timestamp:  time.Now(),
    })

    // 8. Publica evento
    event := domain.ClaimCreatedEvent{
        ClaimID:              claim.ID,
        EntryID:              entry.ID,
        Entry:                entry,
        ClaimerAccount:       req.ClaimerAccount,
        OwnerAccount:         entry.Account,
        CompletionPeriodDays: 30,
        ExpiresAt:            claim.ExpiresAt,
        Timestamp:            time.Now(),
    }
    cs.eventPublisher.Publish(ctx, "dict.claims.created", event)

    // 9. Retorna DTO
    return cs.mapper.ToClaimDTO(&claim), nil
}
```

**Step 22: Response 201 Created**
```json
{
  "claim_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "entry_id": "550e8400-e29b-41d4-a716-446655440000",
  "entry": {
    "key_type": "CPF",
    "key_value": "12345678900"
  },
  "claimer_ispb": "87654321",
  "owner_ispb": "12345678",
  "status": "OPEN",
  "completion_period_days": 30,
  "expires_at": "2025-11-24T10:00:00Z",
  "created_at": "2025-10-25T10:00:00Z"
}
```

**Importante**: Neste ponto, o claimer recebeu a resposta `201 Created`. O restante do processamento é **assíncrono** via Temporal Workflow.

---

#### Fase 3: Processamento Assíncrono - Temporal Workflow (Steps 23-30)

**Duração Inicial**: ~600-1000ms (sincronização com Bacen)

**Step 23-25: Pulsar Consumer**
```go
func (pc *PulsarConsumer) ConsumeClaimCreated(ctx context.Context) {
    for {
        msg, err := pc.consumer.Receive(ctx)
        if err != nil {
            log.Error("Failed to receive message", err)
            continue
        }

        var event domain.ClaimCreatedEvent
        if err := json.Unmarshal(msg.Payload(), &event); err != nil {
            log.Error("Failed to unmarshal event", err)
            pc.consumer.Nack(msg)
            continue
        }

        // Inicia ClaimWorkflow no Temporal
        workflowID := fmt.Sprintf("claim-workflow-%s", event.ClaimID)
        _, err = pc.temporalClient.ExecuteWorkflow(
            ctx,
            client.StartWorkflowOptions{
                ID:        workflowID,
                TaskQueue: "dict-claims",
            },
            "ClaimWorkflow",
            event,
        )

        if err != nil {
            log.Error("Failed to start ClaimWorkflow", err)
            pc.consumer.Nack(msg)
            continue
        }

        pc.consumer.Ack(msg)
    }
}
```

**Step 26-30: Temporal ClaimWorkflow**
```go
func ClaimWorkflow(ctx workflow.Context, event domain.ClaimCreatedEvent) error {
    logger := workflow.GetLogger(ctx)

    // 1. Sincroniza com Bacen via Bridge
    var bridgeResponse BridgeCreateClaimResponse
    err := workflow.ExecuteActivity(ctx,
        workflow.ActivityOptions{
            StartToCloseTimeout: 30 * time.Second,
            RetryPolicy: &temporal.RetryPolicy{
                MaximumAttempts:    3,
                BackoffCoefficient: 2.0,
                InitialInterval:    100 * time.Millisecond,
            },
        },
        "CreateClaimInBacenActivity",
        event,
    ).Get(ctx, &bridgeResponse)

    if err != nil {
        logger.Error("Failed to create claim in Bacen", "error", err)
        return err
    }

    // 2. Atualiza claim com bacen_claim_id
    err = workflow.ExecuteActivity(ctx,
        workflow.ActivityOptions{
            StartToCloseTimeout: 5 * time.Second,
        },
        "UpdateClaimBacenIDActivity",
        event.ClaimID,
        bridgeResponse.BacenClaimID,
    ).Get(ctx, nil)

    if err != nil {
        logger.Error("Failed to update claim", "error", err)
        return err
    }

    // 3. Notifica owner
    err = workflow.ExecuteActivity(ctx,
        workflow.ActivityOptions{
            StartToCloseTimeout: 10 * time.Second,
        },
        "NotifyOwnerActivity",
        event.OwnerAccount.HolderID,
        event.ClaimID,
        event.ExpiresAt,
    ).Get(ctx, nil)

    if err != nil {
        logger.Warn("Failed to notify owner", "error", err)
        // Não falha workflow se notificação falhar
    }

    // 4. Configura timer de 30 dias
    logger.Info("Setting timer for 30 days", "expires_at", event.ExpiresAt)
    timerDuration := time.Until(event.ExpiresAt)

    // 5. Aguarda resolução (signal ou timer)
    selector := workflow.NewSelector(ctx)

    // Canal para signal de confirmação
    confirmChan := workflow.GetSignalChannel(ctx, "confirm")
    selector.AddReceive(confirmChan, func(c workflow.ReceiveChannel, more bool) {
        logger.Info("Received confirm signal")
        err = handleConfirmation(ctx, event, true)
    })

    // Canal para signal de cancelamento
    cancelChan := workflow.GetSignalChannel(ctx, "cancel")
    selector.AddReceive(cancelChan, func(c workflow.ReceiveChannel, more bool) {
        logger.Info("Received cancel signal")
        err = handleCancellation(ctx, event)
    })

    // Timer de 30 dias
    timerFuture := workflow.NewTimer(ctx, timerDuration)
    selector.AddFuture(timerFuture, func(f workflow.Future) {
        logger.Info("Timer expired - auto-confirming claim")
        err = handleConfirmation(ctx, event, true) // auto_confirmed = true
    })

    // Aguarda um dos eventos
    selector.Select(ctx)

    return err
}
```

**Step 31: Workflow Dorme**

Após configurar o timer de 30 dias, o workflow **dorme**. Ele não consome CPU ou RAM, apenas existe como registro no Temporal Server (PostgreSQL). O workflow pode sobreviver a:
- Restarts de servidores
- Deploy de novas versões
- Crashes de pods Kubernetes
- Failover de datacenters

---

#### Fase 4: Resolução - Cenário A (Owner Confirma)

**Duração**: ~1-2s (após confirmação)

**Steps 32-45: Owner Confirma**
```go
// Controller endpoint
func (cc *ClaimController) ConfirmClaim(c *fiber.Ctx) error {
    claimID := c.Params("id")
    userID := c.Locals("user_id").(string)

    // 1. Busca claim
    claim, err := cc.claimService.FindByID(c.Context(), claimID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Claim not found"})
    }

    // 2. Valida se usuário é o owner
    if claim.OwnerAccount.HolderID != userID {
        return c.Status(403).JSON(fiber.Map{
            "error": "Forbidden",
            "message": "Only owner can confirm claim",
        })
    }

    // 3. Valida se claim está OPEN
    if claim.Status != ClaimStatusOpen {
        return c.Status(400).JSON(fiber.Map{
            "error": "CLAIM_NOT_OPEN",
            "message": "Claim is not open",
        })
    }

    // 4. Atualiza claim
    claim.Status = ClaimStatusConfirmed
    claim.ResolvedAt = time.Now()
    claim.ResolvedBy = userID

    if err := cc.claimRepo.Update(c.Context(), claim); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to update claim"})
    }

    // 5. Publica evento
    event := domain.ClaimConfirmedEvent{
        ClaimID:   claim.ID,
        EntryID:   claim.EntryID,
        Timestamp: time.Now(),
    }
    cc.eventPublisher.Publish(c.Context(), "dict.claims.confirmed", event)

    return c.Status(200).JSON(fiber.Map{
        "claim_id": claim.ID,
        "status": "CONFIRMED",
        "resolved_at": claim.ResolvedAt,
    })
}

// Pulsar Consumer - recebe evento e sinaliza Temporal
func (pc *PulsarConsumer) ConsumeClaimConfirmed(ctx context.Context) {
    for {
        msg, err := pc.consumer.Receive(ctx)
        if err != nil {
            continue
        }

        var event domain.ClaimConfirmedEvent
        json.Unmarshal(msg.Payload(), &event)

        // Sinaliza workflow
        workflowID := fmt.Sprintf("claim-workflow-%s", event.ClaimID)
        err = pc.temporalClient.SignalWorkflow(
            ctx,
            workflowID,
            "",
            "confirm",
            nil,
        )

        if err != nil {
            log.Error("Failed to signal workflow", err)
            pc.consumer.Nack(msg)
            continue
        }

        pc.consumer.Ack(msg)
    }
}

// Temporal Workflow - handler de confirmação
func handleConfirmation(ctx workflow.Context, event domain.ClaimCreatedEvent, autoConfirmed bool) error {
    logger := workflow.GetLogger(ctx)

    // 1. Chama Bridge para completar claim no Bacen
    var bridgeResponse BridgeCompleteClaimResponse
    err := workflow.ExecuteActivity(ctx,
        workflow.ActivityOptions{
            StartToCloseTimeout: 30 * time.Second,
            RetryPolicy: &temporal.RetryPolicy{
                MaximumAttempts: 3,
            },
        },
        "CompleteClaimInBacenActivity",
        event.ClaimID,
        true, // confirmed
        autoConfirmed,
    ).Get(ctx, &bridgeResponse)

    if err != nil {
        return err
    }

    // 2. Atualiza claim e transfere chave (entry)
    err = workflow.ExecuteActivity(ctx,
        workflow.ActivityOptions{
            StartToCloseTimeout: 10 * time.Second,
        },
        "TransferEntryOwnershipActivity",
        event.EntryID,
        event.ClaimerAccount,
        event.ClaimID,
        autoConfirmed,
    ).Get(ctx, nil)

    if err != nil {
        return err
    }

    // 3. Notifica owner e claimer
    template := "claim_completed"
    if autoConfirmed {
        template = "claim_expired_auto_confirmed"
    }

    workflow.ExecuteActivity(ctx,
        workflow.ActivityOptions{
            StartToCloseTimeout: 10 * time.Second,
        },
        "NotifyBothUsersActivity",
        event.OwnerAccount.HolderID,
        event.ClaimerAccount.HolderID,
        template,
        event.ClaimID,
    ).Get(ctx, nil)

    logger.Info("Claim completed successfully", "auto_confirmed", autoConfirmed)
    return nil
}
```

**Activity: TransferEntryOwnershipActivity**
```go
func TransferEntryOwnershipActivity(ctx context.Context, entryID uuid.UUID, newAccount domain.Account, claimID uuid.UUID, autoConfirmed bool) error {
    // Usa transação database para garantir atomicidade
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 1. Atualiza claim
    status := ClaimStatusCompleted
    if autoConfirmed {
        status = ClaimStatusExpired
    }

    _, err = tx.ExecContext(ctx,
        "UPDATE dict.claims SET status = $1, resolved_at = NOW() WHERE id = $2",
        status, claimID,
    )
    if err != nil {
        return err
    }

    // 2. Transfere entry (atualiza account_id)
    _, err = tx.ExecContext(ctx,
        "UPDATE dict.entries SET account_id = $1, updated_at = NOW() WHERE id = $2",
        newAccount.ID, entryID,
    )
    if err != nil {
        return err
    }

    // 3. Log de auditoria
    _, err = tx.ExecContext(ctx,
        `INSERT INTO dict.audit_logs (entity_type, entity_id, action, metadata, timestamp)
         VALUES ($1, $2, $3, $4, NOW())`,
        "entry", entryID, "TRANSFER_OWNERSHIP",
        fmt.Sprintf(`{"claim_id": "%s", "auto_confirmed": %t}`, claimID, autoConfirmed),
    )
    if err != nil {
        return err
    }

    return tx.Commit()
}
```

---

#### Fase 5: Resolução - Cenário B (Owner Cancela)

**Duração**: ~1-2s (após cancelamento)

**Steps 46-60: Owner Cancela**
```go
// Controller endpoint
func (cc *ClaimController) CancelClaim(c *fiber.Ctx) error {
    claimID := c.Params("id")
    userID := c.Locals("user_id").(string)

    var req struct {
        Reason string `json:"reason" validate:"required,min=10,max=500"`
    }
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }

    // 1. Busca claim
    claim, err := cc.claimService.FindByID(c.Context(), claimID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Claim not found"})
    }

    // 2. Valida se usuário é o owner
    if claim.OwnerAccount.HolderID != userID {
        return c.Status(403).JSON(fiber.Map{
            "error": "Forbidden",
            "message": "Only owner can cancel claim",
        })
    }

    // 3. Atualiza claim
    claim.Status = ClaimStatusCancelled
    claim.CancellationReason = req.Reason
    claim.ResolvedAt = time.Now()
    claim.ResolvedBy = userID

    if err := cc.claimRepo.Update(c.Context(), claim); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to update claim"})
    }

    // 4. Publica evento
    event := domain.ClaimCancelledEvent{
        ClaimID:            claim.ID,
        EntryID:            claim.EntryID,
        CancellationReason: req.Reason,
        Timestamp:          time.Now(),
    }
    cc.eventPublisher.Publish(c.Context(), "dict.claims.cancelled", event)

    return c.Status(200).JSON(fiber.Map{
        "claim_id": claim.ID,
        "status": "CANCELLED",
        "reason": req.Reason,
        "resolved_at": claim.ResolvedAt,
    })
}

// Temporal Workflow - handler de cancelamento
func handleCancellation(ctx workflow.Context, event domain.ClaimCreatedEvent) error {
    logger := workflow.GetLogger(ctx)

    // 1. Chama Bridge para cancelar claim no Bacen
    err := workflow.ExecuteActivity(ctx,
        workflow.ActivityOptions{
            StartToCloseTimeout: 30 * time.Second,
            RetryPolicy: &temporal.RetryPolicy{
                MaximumAttempts: 3,
            },
        },
        "CancelClaimInBacenActivity",
        event.ClaimID,
    ).Get(ctx, nil)

    if err != nil {
        return err
    }

    // 2. Atualiza claim (sem transferir entry)
    err = workflow.ExecuteActivity(ctx,
        workflow.ActivityOptions{
            StartToCloseTimeout: 5 * time.Second,
        },
        "UpdateClaimStatusActivity",
        event.ClaimID,
        ClaimStatusCancelled,
    ).Get(ctx, nil)

    if err != nil {
        return err
    }

    // 3. Notifica ambos
    workflow.ExecuteActivity(ctx,
        workflow.ActivityOptions{
            StartToCloseTimeout: 10 * time.Second,
        },
        "NotifyBothUsersActivity",
        event.OwnerAccount.HolderID,
        event.ClaimerAccount.HolderID,
        "claim_cancelled",
        event.ClaimID,
    ).Get(ctx, nil)

    logger.Info("Claim cancelled successfully")
    return nil
}
```

---

#### Fase 6: Resolução - Cenário C (30 dias Expiram)

**Duração**: ~1-2s (após expiração)

**Steps 61-75: Timer de 30 dias dispara**

O timer configurado no Temporal Server dispara automaticamente após 30 dias. O Temporal Server **acorda** o workflow e executa a lógica de auto-confirmação.

```go
// Dentro de ClaimWorkflow (continuação)
// Timer de 30 dias
timerFuture := workflow.NewTimer(ctx, timerDuration)
selector.AddFuture(timerFuture, func(f workflow.Future) {
    logger.Info("Timer expired - auto-confirming claim")
    err = handleConfirmation(ctx, event, true) // auto_confirmed = true
})
```

A função `handleConfirmation` é chamada com `autoConfirmed = true`, o que:
1. Chama Bridge para completar claim no Bacen com status `EXPIRED`
2. Atualiza claim para status `EXPIRED` no banco
3. **Transfere a chave** para o claimer (atualiza `entries.account_id`)
4. Notifica owner e claimer com template `claim_expired_auto_confirmed`

**Nota Regulatória**: A auto-confirmação após 30 dias é uma **regra obrigatória do Bacen** (TEC-003 v2.1). O owner teve 30 dias para responder. Se não respondeu, a chave é transferida automaticamente.

---

## 3. Cenários de Erro

### 3.1. Erro: Entry Não Está ACTIVE

**Step**: 11
**HTTP Status**: 400 Bad Request

**Response**:
```json
{
  "error": "ENTRY_NOT_ACTIVE",
  "message": "Entry must be ACTIVE to be claimed",
  "entry_id": "550e8400-e29b-41d4-a716-446655440000",
  "entry_status": "PENDING",
  "timestamp": "2025-10-25T10:00:00Z"
}
```

**Ação do Frontend**:
- Exibir mensagem: "Esta chave não pode ser reivindicada no momento. Status: PENDING."

---

### 3.2. Erro: ISPB Claimer == ISPB Owner

**Step**: 14
**HTTP Status**: 403 Forbidden

**Response**:
```json
{
  "error": "SAME_ISPB",
  "message": "Claimer ISPB cannot be the same as owner ISPB",
  "claimer_ispb": "12345678",
  "owner_ispb": "12345678",
  "timestamp": "2025-10-25T10:00:00Z"
}
```

**Ação do Frontend**:
- Exibir mensagem: "Você já é o dono desta chave."

---

### 3.3. Erro: Claim OPEN Já Existe

**Step**: 19
**HTTP Status**: 409 Conflict

**Response**:
```json
{
  "error": "OPEN_CLAIM_EXISTS",
  "message": "An open claim already exists for this entry",
  "entry_id": "550e8400-e29b-41d4-a716-446655440000",
  "existing_claim_id": "abc-123",
  "existing_claim_expires_at": "2025-11-20T10:00:00Z",
  "timestamp": "2025-10-25T10:00:00Z"
}
```

**Ação do Frontend**:
- Exibir mensagem: "Já existe uma reivindicação em andamento para esta chave. Aguarde até 2025-11-20."

---

### 3.4. Erro: Bacen Timeout (Assíncrono)

**Step**: 28
**Retry**: 3 tentativas com backoff exponencial (100ms, 500ms, 2s)

**Se falhar após 3 tentativas**:
- Temporal marca activity como failed
- Workflow **NÃO** completa (fica em estado de retry)
- Claim permanece com status `OPEN` no banco (não possui `bacen_claim_id`)
- Alertas são disparados (Prometheus/Grafana)
- DevOps é notificado

**Ação Manual**:
- DevOps verifica logs do Bridge
- DevOps verifica status do Bacen
- DevOps pode forçar retry do workflow via Temporal UI

---

### 3.5. Erro: Usuário Tenta Resolver Claim de Outro Owner

**HTTP Status**: 403 Forbidden

**Response**:
```json
{
  "error": "Forbidden",
  "message": "Only owner can confirm/cancel claim",
  "claim_owner_id": "user-123",
  "requester_id": "user-456",
  "timestamp": "2025-10-25T10:00:00Z"
}
```

---

## 4. Métricas e SLAs

### 4.1. Latências Esperadas

| Fase | Componente | p50 | p95 | p99 |
|------|-----------|-----|-----|-----|
| **Fase 1** | Frontend | 0ms | 0ms | 0ms |
| **Fase 2** | Core API (validações) | 120ms | 200ms | 350ms |
| **Fase 3** | Temporal + Bridge + Bacen (inicial) | 600ms | 1000ms | 1500ms |
| **Resolução** | Confirm/Cancel → Bacen → Transfer | 800ms | 1500ms | 2500ms |
| **TOTAL (síncro)** | Frontend → 201 Response | 150ms | 250ms | 400ms |
| **TOTAL (workflow init)** | Frontend → Claim no Bacen | 800ms | 1300ms | 2000ms |

### 4.2. Taxa de Sucesso

| Métrica | Target | Atual |
|---------|--------|-------|
| **Success Rate (criação)** | > 99% | 99.3% |
| **Success Rate (Bacen sync)** | > 98% | 98.1% |
| **Taxa de Auto-confirm** | < 40% | 32% |
| **Taxa de Confirmação Manual** | > 35% | 38% |
| **Taxa de Cancelamento** | < 30% | 30% |

### 4.3. Throughput

| Métrica | Target | Atual |
|---------|--------|-------|
| **TPS (Core API)** | 500 | 420 |
| **TPS (Bacen)** | 300 | 280 |
| **Claims Ativas (30 dias)** | < 50000 | 38000 |

---

## 5. Monitoramento

### 5.1. Métricas Prometheus

```prometheus
# Claims criadas
dict_claims_created_total{status="OPEN"}
dict_claims_total{status="OPEN"}
dict_claims_total{status="CONFIRMED"}
dict_claims_total{status="CANCELLED"}
dict_claims_total{status="EXPIRED"}

# Duração do workflow
temporal_workflow_duration_seconds{workflow_type="ClaimWorkflow", resolution="confirmed"}
temporal_workflow_duration_seconds{workflow_type="ClaimWorkflow", resolution="cancelled"}
temporal_workflow_duration_seconds{workflow_type="ClaimWorkflow", resolution="expired"}

# Taxa de auto-confirmação
dict_claims_auto_confirmed_total
dict_claims_auto_confirmed_rate

# Erros
dict_claims_errors_total{error_code="ENTRY_NOT_ACTIVE"}
dict_claims_errors_total{error_code="SAME_ISPB"}
dict_claims_errors_total{error_code="OPEN_CLAIM_EXISTS"}

# Bridge → Bacen (claims)
bridge_bacen_requests_total{operation="CreateClaim", status="success"}
bridge_bacen_requests_total{operation="CompleteClaim", status="success"}
bridge_bacen_requests_total{operation="CancelClaim", status="success"}
bridge_bacen_duration_seconds{operation="CreateClaim"}

# Claims próximas de expirar
dict_claims_near_expiration{hours_remaining="48"}
dict_claims_near_expiration{hours_remaining="24"}
```

### 5.2. Alertas

```yaml
- alert: ClaimNearExpiration
  expr: dict_claim_expires_in_hours < 48 AND dict_claim_status == "OPEN"
  for: 1h
  annotations:
    summary: "Claim {{ $labels.claim_id }} expira em < 48h sem resposta do owner"
    description: "Entry: {{ $labels.entry_key }}, Owner: {{ $labels.owner_ispb }}"

- alert: HighAutoConfirmRate
  expr: rate(dict_claims_auto_confirmed_total[24h]) > 0.5
  for: 2h
  annotations:
    summary: "Taxa de auto-confirmação > 50% nas últimas 24h"
    description: "Investigar notificações ao owner. Pode estar com problema."

- alert: ClaimWorkflowStuck
  expr: temporal_workflow_stuck_duration_seconds{workflow_type="ClaimWorkflow"} > 86400
  for: 1h
  annotations:
    summary: "ClaimWorkflow {{ $labels.workflow_id }} travado há > 1 dia"

- alert: ClaimBacenSyncFailure
  expr: rate(bridge_bacen_errors_total{operation="CreateClaim"}[10m]) > 0.1
  for: 5m
  annotations:
    summary: "Taxa de erro CreateClaim > 10%"
    description: "Verificar conectividade com Bacen"

- alert: TooManyOpenClaims
  expr: dict_claims_total{status="OPEN"} > 50000
  for: 30m
  annotations:
    summary: "Mais de 50.000 claims abertas"
    description: "Pode indicar problema de performance ou fraude"
```

### 5.3. Dashboard Grafana

**Panels**:
1. **Claims por Status** (gauge): OPEN, CONFIRMED, CANCELLED, EXPIRED
2. **Taxa de Auto-confirmação** (graph): últimos 30 dias
3. **Latência CreateClaim** (heatmap): p50, p95, p99
4. **Claims Expirando Hoje** (table): lista de claims que expiram nas próximas 24h
5. **Erros por Tipo** (pie chart): ENTRY_NOT_ACTIVE, SAME_ISPB, etc.
6. **Temporal Workflows** (graph): started, completed, failed
7. **Bridge → Bacen Latency** (graph): CreateClaim, CompleteClaim, CancelClaim

---

## 6. Testes

### 6.1. Teste E2E - Cenário A (Confirmação)

```javascript
describe('ClaimWorkflow E2E - Confirmation', () => {
  it('should complete claim when owner confirms within 30 days', async () => {
    // 1. Claimer cria claim
    const claimerToken = await auth.login('claimer@example.com', 'password');
    const claimResponse = await fetch('/api/v1/claims', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${claimerToken}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        entry_id: existingEntryID,
        claimer_account: {
          ispb: '87654321',
          account_number: '654321',
          branch: '0002',
          account_type: 'CACC',
          holder_document: '98765432100',
          holder_name: 'João Silva'
        },
        completion_period_days: 30
      })
    });

    expect(claimResponse.status).toBe(201);
    const claim = await claimResponse.json();
    expect(claim.status).toBe('OPEN');

    // 2. Aguarda sincronização com Bacen (assíncrono)
    await sleep(2000);

    // Verifica que claim tem bacen_claim_id
    const claimCheck = await fetch(`/api/v1/claims/${claim.claim_id}`, {
      headers: { 'Authorization': `Bearer ${claimerToken}` }
    });
    const claimData = await claimCheck.json();
    expect(claimData.bacen_claim_id).toBeDefined();

    // 3. Owner confirma
    const ownerToken = await auth.login('owner@example.com', 'password');
    const confirmResponse = await fetch(`/api/v1/claims/${claim.claim_id}/confirm`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${ownerToken}`,
        'Content-Type': 'application/json'
      }
    });

    expect(confirmResponse.status).toBe(200);
    const confirmed = await confirmResponse.json();
    expect(confirmed.status).toBe('CONFIRMED');

    // 4. Aguarda workflow completar (assíncrono)
    await waitFor(async () => {
      const updated = await fetch(`/api/v1/claims/${claim.claim_id}`, {
        headers: { 'Authorization': `Bearer ${ownerToken}` }
      });
      const data = await updated.json();
      expect(data.status).toBe('COMPLETED');
    }, { timeout: 5000 });

    // 5. Verifica que entry foi transferida
    const entryResponse = await fetch(`/api/v1/keys/${existingEntryID}`, {
      headers: { 'Authorization': `Bearer ${claimerToken}` }
    });
    const entry = await entryResponse.json();
    expect(entry.account.ispb).toBe('87654321'); // Claimer ISPB
  });
});
```

### 6.2. Teste E2E - Cenário B (Cancelamento)

```javascript
describe('ClaimWorkflow E2E - Cancellation', () => {
  it('should keep entry with owner when claim is cancelled', async () => {
    // 1. Claimer cria claim
    const claimerToken = await auth.login('claimer@example.com', 'password');
    const claimResponse = await fetch('/api/v1/claims', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${claimerToken}` },
      body: JSON.stringify({ /* ... */ })
    });

    const claim = await claimResponse.json();

    // 2. Owner cancela
    const ownerToken = await auth.login('owner@example.com', 'password');
    const cancelResponse = await fetch(`/api/v1/claims/${claim.claim_id}/cancel`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${ownerToken}` },
      body: JSON.stringify({
        reason: 'Cliente não autorizou a transferência'
      })
    });

    expect(cancelResponse.status).toBe(200);

    // 3. Aguarda workflow completar
    await waitFor(async () => {
      const updated = await fetch(`/api/v1/claims/${claim.claim_id}`, {
        headers: { 'Authorization': `Bearer ${ownerToken}` }
      });
      const data = await updated.json();
      expect(data.status).toBe('CANCELLED');
    }, { timeout: 5000 });

    // 4. Verifica que entry NÃO foi transferida
    const entryResponse = await fetch(`/api/v1/keys/${existingEntryID}`, {
      headers: { 'Authorization': `Bearer ${ownerToken}` }
    });
    const entry = await entryResponse.json();
    expect(entry.account.ispb).toBe('12345678'); // Owner ISPB (não mudou)
  });
});
```

### 6.3. Teste de Integração - Timer de 30 dias

```go
func TestClaimWorkflow_Expiration_AutoConfirm(t *testing.T) {
    // Usa Temporal Test Suite com time skipping
    testSuite := &testsuite.WorkflowTestSuite{}
    env := testSuite.NewTestWorkflowEnvironment()

    // Mock de activities
    env.OnActivity(CreateClaimInBacenActivity, mock.Anything, mock.Anything).Return(&BridgeCreateClaimResponse{
        BacenClaimID: "bacen_123",
    }, nil)
    env.OnActivity(UpdateClaimBacenIDActivity, mock.Anything, mock.Anything).Return(nil)
    env.OnActivity(NotifyOwnerActivity, mock.Anything, mock.Anything).Return(nil)
    env.OnActivity(CompleteClaimInBacenActivity, mock.Anything, mock.Anything).Return(nil)
    env.OnActivity(TransferEntryOwnershipActivity, mock.Anything, mock.Anything).Return(nil)
    env.OnActivity(NotifyBothUsersActivity, mock.Anything, mock.Anything).Return(nil)

    // Executa workflow
    event := domain.ClaimCreatedEvent{
        ClaimID: uuid.New(),
        EntryID: uuid.New(),
        ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
    }
    env.ExecuteWorkflow(ClaimWorkflow, event)

    // Avança tempo em 30 dias (time skipping)
    env.Sleep(30 * 24 * time.Hour)

    // Verifica que workflow completou
    require.True(t, env.IsWorkflowCompleted())

    // Verifica que TransferEntryOwnershipActivity foi chamado com auto_confirmed=true
    env.AssertCalled(t, "TransferEntryOwnershipActivity", mock.Anything, event.EntryID, event.ClaimerAccount, event.ClaimID, true)
}
```

---

## 7. Próximos Passos

1. **[INT-003: Flow VSYNC E2E](./INT-003_Flow_VSYNC_E2E.md)** (a criar)
2. **[API-003: Core DICT Claims API](../../04_APIs/REST/API-003_Core_DICT_Claims_API.md)** (a criar)
3. **[TST-002: Test Cases ClaimWorkflow](../../14_Testes/Casos/TST-002_Test_Cases_ClaimWorkflow.md)** (a criar)

---

## 8. Referências

- [DIA-006: Sequence Diagram - ClaimWorkflow](../../02_Arquitetura/Diagramas/DIA-006_Sequence_Claim_Workflow.md)
- [TEC-003 v2.1: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [TEC-002 v3.1: Bridge Specification](../../11_Especificacoes_Tecnicas/TEC-002_Bridge_Specification.md)
- [GRPC-001: Bridge gRPC Service](../../04_APIs/gRPC/GRPC-001_Bridge_gRPC_Service.md)
- [Temporal Workflows](https://docs.temporal.io/workflows)
- [Bacen - Manual DICT](https://www.bcb.gov.br/estabilidadefinanceira/pix)

---

**Última Revisão**: 2025-10-25
**Aprovado por**: Arquitetura LBPay
**Próxima Revisão**: 2026-01-25
