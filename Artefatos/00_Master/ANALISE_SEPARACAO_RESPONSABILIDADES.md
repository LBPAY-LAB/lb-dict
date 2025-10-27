# Análise: Separação de Responsabilidades Core-Dict vs Conn-Dict
**Data**: 2025-10-27 18:30 BRT
**Autor**: Claude Sonnet 4.5 (Architect Analysis)
**Versão**: 1.0
**Status**: Análise Arquitetural Crítica

---

## 🎯 PERGUNTA ESSENCIAL

> **"Workflows de negócio complexos (como Reivindicações) devem estar no Core-Dict ou Conn-Dict?"**

### Resposta Direta

**✅ WORKFLOWS DE NEGÓCIO → CORE-DICT**
**✅ INTEGRAÇÃO TÉCNICA (connection pool, retry, circuit breaker) → CONN-DICT**

**Por quê?** Vou detalhar cada camada arquitetural abaixo.

---

## 📐 PRINCÍPIOS ARQUITETURAIS (Fundamento da Decisão)

### 1. **Separation of Concerns** (SoC)

```
┌─────────────────────────────────────────────────────────────┐
│                    CAMADAS DO SISTEMA                        │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  🏢 CORE-DICT (Business Layer)                              │
│     ├─ Lógica de Negócio PIX/DICT                           │
│     ├─ Regras Regulatórias Bacen                            │
│     ├─ Workflows Complexos (Claim, Portability)             │
│     ├─ Validações de Domínio                                │
│     └─ Orquestração de Processos                            │
│                                                               │
│  🔌 CONN-DICT (Integration Layer)                           │
│     ├─ Adaptação gRPC/Pulsar                                │
│     ├─ Connection Pool Management                            │
│     ├─ Retry Durável (Temporal)                             │
│     ├─ Circuit Breaker                                       │
│     └─ Transformação de Protocolos                          │
│                                                               │
│  🌉 CONN-BRIDGE (Protocol Adapter)                          │
│     ├─ SOAP/XML Transformation                              │
│     ├─ mTLS/ICP-Brasil                                       │
│     ├─ Assinatura Digital                                    │
│     └─ Bacen DICT API Calls                                 │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

### 2. **Domain-Driven Design** (DDD)

**Bounded Contexts**:
- **Core-Dict**: Contexto de Domínio PIX (Business)
- **Conn-Dict**: Contexto de Integração (Technical)
- **Conn-Bridge**: Contexto de Adaptação de Protocolo (Technical)

### 3. **Hexagonal Architecture** (Ports & Adapters)

```
Core-Dict (Hexágono Central)
    ↓ Port: gRPC/Pulsar
Conn-Dict (Adapter Externo)
    ↓ Port: gRPC
Conn-Bridge (Adapter Externo)
    ↓ Port: SOAP/HTTPS
Bacen DICT API
```

---

## 🏢 CORE-DICT: Lógica de Negócio (Business Layer)

### RESPONSABILIDADES CORE-DICT

#### 1. **Workflows de Negócio Complexos** ✅ CORE-DICT

##### Exemplo: ClaimWorkflow (Reivindicação de Chave)

**Complexidade de Negócio**:
```
┌─────────────────────────────────────────────────────────────┐
│           ClaimWorkflow (30 dias durável)                    │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  1. User A reivindica chave de User B                        │
│     ↓                                                         │
│  2. CORE valida: User A é dono original?                     │
│     ├─ Consulta histórico transacional                       │
│     ├─ Valida documentos                                     │
│     └─ Verifica se chave foi roubada/clonada                 │
│     ↓                                                         │
│  3. CORE cria Claim (status: OPEN)                           │
│     ↓                                                         │
│  4. CORE notifica User B (proprietário atual)                │
│     ├─ Email                                                  │
│     ├─ Push notification                                     │
│     └─ SMS                                                    │
│     ↓                                                         │
│  5. CORE aguarda 7 dias (Bacen rule)                         │
│     ├─ Timer Temporal (durável)                              │
│     └─ Se User B não responder → auto-accept                 │
│     ↓                                                         │
│  6. CORE decide (business rule):                             │
│     ├─ User B confirmou? → CONFIRMED                         │
│     ├─ User B rejeitou? → DENIED                             │
│     ├─ Timeout (7 dias)? → AUTO-CONFIRMED                    │
│     └─ Evidências de fraude? → DENIED                        │
│     ↓                                                         │
│  7. CORE executa ação:                                        │
│     ├─ CONFIRMED: Transfer chave para User A                 │
│     ├─ DENIED: Mantém chave com User B                       │
│     └─ Atualiza status em PostgreSQL                         │
│     ↓                                                         │
│  8. CORE notifica resultado a ambos usuários                 │
│     ↓                                                         │
│  9. CORE registra audit log (compliance Bacen)               │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

**Por que no CORE-DICT?**
- ✅ **Regras de negócio complexas** (validação de propriedade, histórico, fraude)
- ✅ **Decisões baseadas em contexto** (histórico transacional, perfil usuário)
- ✅ **Orquestração multi-step** (notificar → aguardar → decidir → executar)
- ✅ **Compliance Bacen** (audit logs, rastreabilidade, reports regulatórios)
- ✅ **Integração com outros domínios** (Fraud Detection, User Profile, Transaction History)

**Por que NÃO no CONN-DICT?**
- ❌ Conn-Dict não tem contexto de negócio
- ❌ Conn-Dict não conhece regras Bacen complexas
- ❌ Conn-Dict não acessa outros domínios (Fraud, User, etc)
- ❌ Separação de concerns violada

---

#### 2. **Validações de Domínio** ✅ CORE-DICT

**Exemplos**:

```go
// CORE-DICT: Validação de regra de negócio
func (s *EntryService) CreateEntry(cmd CreateEntryCommand) error {
    // 1. Validação de limite (regra Bacen: max 5 chaves por conta)
    count, _ := s.entryRepo.CountByAccountID(cmd.AccountID)
    if count >= 5 {
        return ErrMaxKeysExceeded // Business rule
    }

    // 2. Validação de duplicata (regra Bacen: chave única por participante)
    exists, _ := s.entryRepo.ExistsByKey(cmd.KeyValue)
    if exists {
        return ErrKeyAlreadyExists // Business rule
    }

    // 3. Validação de conta ativa (regra de negócio)
    account, _ := s.accountService.GetAccount(cmd.AccountID)
    if account.Status != "ACTIVE" {
        return ErrAccountInactive // Business rule
    }

    // 4. Validação de ownership (regra de negócio)
    if cmd.KeyType == CPF && cmd.KeyValue != account.OwnerCPF {
        return ErrKeyOwnershipMismatch // Business rule
    }

    // 5. Validação anti-fraude (regra de negócio)
    fraudScore, _ := s.fraudService.CheckFraud(cmd)
    if fraudScore > 0.8 {
        return ErrSuspiciousFraud // Business rule
    }

    // Após TODAS as validações de negócio → chama Conn-Dict
    return s.connectClient.CreateEntry(ctx, protoReq)
}
```

**Por que no CORE-DICT?**
- ✅ Validações dependem de contexto de negócio
- ✅ Integração com múltiplos serviços (Account, Fraud, User)
- ✅ Regras Bacen complexas (limite, ownership, duplicata)

---

#### 3. **Orquestração de Processos** ✅ CORE-DICT

**Exemplo: Portabilidade de Chave**

```go
// CORE-DICT: Orquestração complexa
func (s *PortabilityService) InitiatePortability(req InitiatePortabilityRequest) error {
    // 1. Validar elegibilidade (business rule)
    if !s.isEligibleForPortability(req.Key) {
        return ErrNotEligible
    }

    // 2. Verificar histórico de portabilidades (business rule: max 2 por ano)
    count, _ := s.portabilityRepo.CountLastYear(req.Key)
    if count >= 2 {
        return ErrMaxPortabilityExceeded
    }

    // 3. Consultar instituição destino (validação)
    destinationBank, _ := s.bankService.GetBank(req.DestinationISPB)
    if !destinationBank.AcceptsPortability {
        return ErrDestinationNotSupported
    }

    // 4. Criar portability request (state)
    portability := s.portabilityRepo.Create(req)

    // 5. Notificar banco origem (business process)
    s.notificationService.NotifyOriginBank(portability)

    // 6. Chamar Conn-Dict para executar no Bacen (integration)
    err := s.connectClient.InitiatePortability(ctx, protoReq)

    // 7. Atualizar status baseado em resposta (state management)
    if err != nil {
        portability.Status = "FAILED"
        portability.FailureReason = err.Error()
    } else {
        portability.Status = "PENDING_APPROVAL"
    }
    s.portabilityRepo.Update(portability)

    // 8. Registrar audit log (compliance)
    s.auditService.LogPortabilityInitiated(portability)

    return nil
}
```

**Por que no CORE-DICT?**
- ✅ Múltiplos passos com lógica de negócio
- ✅ Validações complexas (elegibilidade, histórico, limites)
- ✅ Integração com múltiplos serviços
- ✅ State management (PostgreSQL próprio)
- ✅ Audit logs e compliance

---

#### 4. **Gestão de Estado de Negócio** ✅ CORE-DICT

```go
// CORE-DICT: Estado de negócio rico
type Entry struct {
    EntryID           string
    KeyValue          string
    KeyType           KeyType
    Status            EntryStatus // Business state

    // Business context
    AccountID         string
    OwnerName         string
    OwnerCPF          string

    // Business history
    CreatedBy         string
    CreatedAt         time.Time
    UpdatedAt         time.Time
    DeletedAt         *time.Time

    // Business tracking
    ClaimHistory      []Claim
    PortabilityHistory []Portability
    AuditLogs         []AuditLog

    // Business rules
    MaxClaims         int // Regra: max 3 claims por chave
    IsLocked          bool // Regra: lock durante portabilidade
}
```

**Por que no CORE-DICT?**
- ✅ Estado rico com contexto de negócio
- ✅ Histórico para decisões futuras
- ✅ Audit trail para compliance

---

## 🔌 CONN-DICT: Camada de Integração (Integration Layer)

### RESPONSABILIDADES CONN-DICT

#### 1. **Connection Pool Management** ✅ CONN-DICT

```go
// CONN-DICT: Gerenciamento técnico de conexões
type BridgeClientPool struct {
    clients       []*BridgeClient
    maxConnections int
    currentIndex  int
    mu            sync.Mutex
}

func (p *BridgeClientPool) GetClient() *BridgeClient {
    p.mu.Lock()
    defer p.mu.Unlock()

    // Round-robin load balancing
    client := p.clients[p.currentIndex]
    p.currentIndex = (p.currentIndex + 1) % len(p.clients)

    return client
}

func (p *BridgeClientPool) HealthCheck() {
    // Remove unhealthy clients
    // Add new clients if needed
    // Rebalance load
}
```

**Por que no CONN-DICT?**
- ✅ Concern técnico de infraestrutura
- ✅ Não tem contexto de negócio
- ✅ Reutilizável para qualquer tipo de request
- ✅ Core-Dict não deve saber de connection pools

---

#### 2. **Retry Durável com Temporal** ✅ CONN-DICT

```go
// CONN-DICT: Retry técnico (não business logic)
func (w *BridgeCallActivity) Execute(ctx context.Context, req BridgeRequest) error {
    // Retry com backoff exponencial
    retryPolicy := &temporal.RetryPolicy{
        InitialInterval:    1 * time.Second,
        BackoffCoefficient: 2.0,
        MaximumInterval:    10 * time.Second,
        MaximumAttempts:    5,
    }

    // Se Bridge retornar erro HTTP (503, 500, 429) → retry
    // Se Bridge retornar erro de negócio (404, 400) → NÃO retry
    return w.bridgeClient.Call(ctx, req)
}
```

**Por que no CONN-DICT?**
- ✅ Retry é concern técnico de infraestrutura
- ✅ Não depende de regras de negócio
- ✅ Core-Dict não deve gerenciar retry HTTP

**MAS ATENÇÃO**: Workflow de NEGÓCIO (ClaimWorkflow) fica no CORE!

---

#### 3. **Circuit Breaker** ✅ CONN-DICT

```go
// CONN-DICT: Proteção contra falhas em cascata
type CircuitBreaker struct {
    breaker *gobreaker.CircuitBreaker
}

func (cb *CircuitBreaker) CallBridge(ctx context.Context, req BridgeRequest) error {
    // Se Bridge estiver falhando muito → OPEN circuit
    // Evita sobrecarregar Bacen
    // Retorna erro rápido para Core-Dict
    result, err := cb.breaker.Execute(func() (interface{}, error) {
        return cb.bridgeClient.Call(ctx, req)
    })

    if err != nil {
        // Core-Dict recebe erro e decide o que fazer (business decision)
        return err
    }

    return result
}
```

**Por que no CONN-DICT?**
- ✅ Proteção de infraestrutura
- ✅ Não depende de lógica de negócio
- ✅ Reutilizável para todos os tipos de request

---

#### 4. **Transformação de Protocolo** ✅ CONN-DICT

```go
// CONN-DICT: Adaptação técnica de protocolos
func (s *ConnectService) CreateEntry(ctx context.Context, req *corepb.CreateEntryRequest) (*corepb.CreateEntryResponse, error) {
    // 1. Converte Core proto → Connect proto
    connectReq := s.convertCoreToConnect(req)

    // 2. Envia para Bridge via gRPC
    bridgeResp, err := s.bridgeClient.CreateEntry(ctx, connectReq)

    // 3. Converte Bridge proto → Core proto
    coreResp := s.convertBridgeToCore(bridgeResp)

    return coreResp, nil
}
```

**Por que no CONN-DICT?**
- ✅ Transformação técnica de mensagens
- ✅ Core-Dict não deve conhecer detalhes de proto do Bridge
- ✅ Separação de concerns (Core não sabe de Bridge)

---

#### 5. **Pulsar Event Handling** ✅ CONN-DICT

```go
// CONN-DICT: Consumir eventos assíncronos do Core
func (c *PulsarConsumer) HandleEntryCreatedEvent(event EntryCreatedEvent) {
    // 1. Deserializa evento Pulsar
    req := s.convertEventToRequest(event)

    // 2. Chama Bridge (síncrono)
    err := s.bridgeClient.CreateEntry(ctx, req)

    // 3. Publica resultado de volta para Core (Pulsar)
    if err != nil {
        s.pulsarProducer.Publish("dict.entry.created.failed", FailedEvent{...})
    } else {
        s.pulsarProducer.Publish("dict.entry.created.success", SuccessEvent{...})
    }

    // 4. Ack ou Nack (Pulsar concern)
    if err != nil {
        return err // Nack → retry
    }
    return nil // Ack
}
```

**Por que no CONN-DICT?**
- ✅ Infraestrutura de mensageria
- ✅ Core-Dict não deve conhecer detalhes de Pulsar
- ✅ Conn-Dict é o "adaptador" entre Core e Bacen

---

## 🌉 CONN-BRIDGE: Adaptador de Protocolo (Protocol Adapter)

### RESPONSABILIDADES CONN-BRIDGE

#### 1. **SOAP/XML Transformation** ✅ CONN-BRIDGE

```go
// CONN-BRIDGE: Transformação SOAP/XML (concern técnico)
func (b *BridgeService) CreateEntry(ctx context.Context, req *bridgepb.CreateEntryRequest) (*bridgepb.CreateEntryResponse, error) {
    // 1. Proto → XML
    xmlReq := b.converter.ProtoToXML(req)

    // 2. XML → SOAP envelope
    soapEnvelope := b.soapBuilder.BuildEnvelope(xmlReq)

    // 3. Assinar XML (ICP-Brasil A3)
    signedSOAP := b.xmlSigner.Sign(soapEnvelope)

    // 4. POST HTTPS + mTLS para Bacen
    httpResp, err := b.httpClient.Post(bacenURL, signedSOAP)

    // 5. Parse SOAP response
    xmlResp := b.soapParser.ParseResponse(httpResp)

    // 6. XML → Proto
    protoResp := b.converter.XMLToProto(xmlResp)

    return protoResp, nil
}
```

**Por que no CONN-BRIDGE?**
- ✅ Transformação técnica de protocolo (gRPC ↔ SOAP)
- ✅ Core e Connect não devem conhecer SOAP/XML
- ✅ Bridge é o único que "fala" com Bacen

---

## 🔍 ANÁLISE: Onde Ficam os Workflows?

### ❌ ERRADO: Workflow de Negócio no CONN-DICT

```go
// ❌ ERRADO: ClaimWorkflow no CONN-DICT
// Conn-Dict NÃO deve ter lógica de negócio!

func (c *ConnectService) CreateClaim(ctx context.Context, req *CreateClaimRequest) error {
    // ❌ ERRO: Validação de negócio no Conn-Dict
    if req.ClaimerISPB == req.OwnerISPB {
        return ErrSameParticipant // Business rule → deveria estar no Core!
    }

    // ❌ ERRO: Consulta de histórico de negócio
    claimHistory := c.claimRepo.GetHistory(req.Key)
    if len(claimHistory) >= 3 {
        return ErrMaxClaimsExceeded // Business rule → deveria estar no Core!
    }

    // ❌ ERRO: Temporal Workflow de NEGÓCIO no Conn-Dict
    workflowID := "claim-" + req.Key
    c.temporalClient.ExecuteWorkflow(ctx, workflowID, ClaimWorkflow, req)
    // ^ ERRADO! ClaimWorkflow tem lógica de negócio → deveria estar no Core!
}
```

**Por que está ERRADO?**
- ❌ Conn-Dict não tem contexto de negócio
- ❌ Conn-Dict não deveria validar regras Bacen
- ❌ Conn-Dict não deveria consultar histórico de claims
- ❌ Violação de SoC (Separation of Concerns)
- ❌ Dificulta testes (business logic misturada com infra)
- ❌ Dificulta manutenção (mudança de regra requer mudança em Conn-Dict)

---

### ✅ CORRETO: Workflow de Negócio no CORE-DICT

```go
// ✅ CORRETO: ClaimWorkflow no CORE-DICT

// CORE-DICT: Domain Service
func (s *ClaimService) CreateClaim(ctx context.Context, cmd CreateClaimCommand) error {
    // 1. Validações de negócio (CORE)
    if cmd.ClaimerISPB == cmd.OwnerISPB {
        return ErrSameParticipant
    }

    // 2. Consulta histórico (CORE)
    claimHistory := s.claimRepo.GetHistory(cmd.Key)
    if len(claimHistory) >= 3 {
        return ErrMaxClaimsExceeded
    }

    // 3. Validação anti-fraude (CORE - integração com outro domínio)
    fraudScore, _ := s.fraudService.CheckFraud(cmd)
    if fraudScore > 0.8 {
        return ErrSuspiciousFraud
    }

    // 4. Iniciar Temporal Workflow DE NEGÓCIO (CORE)
    workflowID := "claim-" + cmd.Key
    err := s.temporalClient.ExecuteWorkflow(ctx, workflowID, ClaimWorkflow, cmd)

    return err
}

// CORE-DICT: Temporal Workflow (lógica de negócio)
func ClaimWorkflow(ctx workflow.Context, cmd CreateClaimCommand) error {
    // 1. Criar claim (business state)
    claim := CreateClaim(cmd)
    SaveClaim(claim) // Core PostgreSQL

    // 2. Notificar proprietário atual (business process)
    workflow.ExecuteActivity(ctx, NotifyOwnerActivity, claim)

    // 3. Aguardar 7 dias (business rule Bacen)
    workflow.Sleep(ctx, 7*24*time.Hour)

    // 4. Verificar se owner respondeu (business logic)
    response := GetOwnerResponse(claim.ID)

    // 5. Decidir resultado (business rule)
    if response == nil {
        // Auto-accept após 7 dias (business rule)
        claim.Status = "AUTO-CONFIRMED"
    } else if response.Accepted {
        claim.Status = "CONFIRMED"
    } else {
        claim.Status = "DENIED"
    }

    // 6. Executar ação no Bacen (infraestrutura - delega para Conn-Dict)
    if claim.Status == "CONFIRMED" || claim.Status == "AUTO-CONFIRMED" {
        // AQUI sim, chama Conn-Dict (infra layer)
        workflow.ExecuteActivity(ctx, CallConnectActivity, CallConnectRequest{
            Method: "CompleteClaim",
            ClaimID: claim.ID,
        })
    }

    // 7. Atualizar estado final (business state)
    SaveClaim(claim) // Core PostgreSQL

    // 8. Audit log (compliance)
    workflow.ExecuteActivity(ctx, AuditLogActivity, claim)

    return nil
}

// CORE-DICT: Activity que chama Conn-Dict (ponte infra)
func CallConnectActivity(ctx context.Context, req CallConnectRequest) error {
    // Esta Activity é simples: apenas chama Conn-Dict
    // Conn-Dict lida com connection pool, retry, circuit breaker
    return s.connectClient.CompleteClaim(ctx, protoReq)
}
```

**Por que está CORRETO?**
- ✅ Lógica de negócio (validações, decisões) no Core-Dict
- ✅ Conn-Dict é chamado APENAS para executar no Bacen (infra)
- ✅ Separação clara: Business (Core) vs Infrastructure (Connect)
- ✅ Testável: Mock Conn-Dict para testar business logic
- ✅ Manutenível: Mudança de regra → apenas Core-Dict

---

## 📊 TABELA RESUMO: Onde Fica O Quê?

| Responsabilidade | Core-Dict | Conn-Dict | Conn-Bridge | Justificativa |
|------------------|-----------|-----------|-------------|---------------|
| **Validações de Negócio** | ✅ | ❌ | ❌ | Core tem contexto de domínio |
| **Regras Bacen Complexas** | ✅ | ❌ | ❌ | Core implementa compliance |
| **Workflows Complexos (Claim, Portability)** | ✅ | ❌ | ❌ | Core orquestra processos de negócio |
| **Decisões Baseadas em Contexto** | ✅ | ❌ | ❌ | Core tem histórico, perfil, fraud detection |
| **State Management (Business)** | ✅ | ❌ | ❌ | Core mantém estado rico de negócio |
| **Integração com Outros Domínios** | ✅ | ❌ | ❌ | Core orquestra (Fraud, User, Account) |
| **Audit Logs (Compliance)** | ✅ | ❌ | ❌ | Core responsável por compliance |
| **Connection Pool Management** | ❌ | ✅ | ❌ | Connect gerencia infra técnica |
| **Retry Durável (Temporal)** | ❌ | ✅ | ❌ | Connect faz retry técnico |
| **Circuit Breaker** | ❌ | ✅ | ❌ | Connect protege infra |
| **Transformação de Protocolo (gRPC/Pulsar)** | ❌ | ✅ | ❌ | Connect adapta protocolos |
| **Pulsar Event Handling** | ❌ | ✅ | ❌ | Connect consome/produz eventos |
| **Balanceamento de Carga** | ❌ | ✅ | ❌ | Connect distribui requests |
| **SOAP/XML Transformation** | ❌ | ❌ | ✅ | Bridge adapta SOAP |
| **mTLS/ICP-Brasil** | ❌ | ❌ | ✅ | Bridge lida com Bacen |
| **Assinatura Digital XML** | ❌ | ❌ | ✅ | Bridge assina com A3 |
| **Chamada HTTPS para Bacen** | ❌ | ❌ | ✅ | Bridge executa HTTP |

---

## 🎯 DECISÕES ARQUITETURAIS CRÍTICAS

### Decisão 1: ClaimWorkflow no Core-Dict ✅

**Contexto**: Claim é processo de negócio de 7-30 dias com decisões complexas.

**Decisão**: ClaimWorkflow Temporal vive no **Core-Dict**.

**Razões**:
1. **Lógica de Negócio Complexa**:
   - Validação de propriedade (histórico transacional)
   - Detecção de fraude (integração Fraud Detection Service)
   - Decisões baseadas em contexto (perfil usuário, histórico)

2. **Orquestração Multi-Serviço**:
   - Core-Dict orquestra: Notification Service, Fraud Service, User Service
   - Conn-Dict não tem acesso a esses serviços

3. **State Management Rico**:
   - Claim tem estado complexo (history, logs, attachments)
   - Core-Dict PostgreSQL mantém esse estado

4. **Compliance Bacen**:
   - Core-Dict gera audit logs regulatórios
   - Core-Dict rastreia compliance

**Alternativa Rejeitada**: ClaimWorkflow no Conn-Dict
- ❌ Conn-Dict não tem contexto de negócio
- ❌ Conn-Dict não acessa outros domínios
- ❌ Violaria Separation of Concerns

---

### Decisão 2: BridgeCallActivity no Conn-Dict ✅

**Contexto**: Retry técnico de chamadas HTTP ao Bridge.

**Decisão**: BridgeCallActivity Temporal vive no **Conn-Dict**.

**Razões**:
1. **Concern Técnico de Infraestrutura**:
   - Retry não depende de lógica de negócio
   - É transparente para Core-Dict

2. **Reutilizável**:
   - Mesma lógica de retry para todas as chamadas
   - Core-Dict não precisa saber de retry HTTP

3. **Connection Pool**:
   - Conn-Dict gerencia pool de conexões ao Bridge
   - Retry deve usar o mesmo pool

**Alternativa Rejeitada**: BridgeCallActivity no Core-Dict
- ❌ Core-Dict não deveria gerenciar retry HTTP
- ❌ Violaria Separation of Concerns

---

### Decisão 3: Balde de Conexões Bacen no Conn-Dict ✅

**Contexto**: Bacen tem rate limit (1000 TPS). Precisa gerenciar pool de conexões.

**Decisão**: Connection Pool Management no **Conn-Dict**.

**Razões**:
1. **Infraestrutura Técnica**:
   - Rate limiting é concern técnico, não de negócio
   - Core-Dict não deve saber de TPS limits

2. **Transparência**:
   - Core-Dict chama ConnectService normalmente
   - Conn-Dict gerencia pool internamente

3. **Resiliência**:
   - Conn-Dict implementa circuit breaker
   - Conn-Dict faz retry com backoff
   - Core-Dict recebe erro ou sucesso

**Implementação**:
```go
// CONN-DICT: Connection Pool
type BridgeConnectionPool struct {
    connections []*grpc.ClientConn
    maxConn     int // Bacen rate limit: 1000 TPS
    semaphore   chan struct{}
}

func (p *BridgeConnectionPool) AcquireConnection() (*grpc.ClientConn, error) {
    // Wait for available slot (rate limiting)
    select {
    case p.semaphore <- struct{}{}:
        return p.getHealthyConnection(), nil
    case <-time.After(5 * time.Second):
        return nil, ErrConnectionPoolExhausted
    }
}

func (p *BridgeConnectionPool) ReleaseConnection() {
    <-p.semaphore
}
```

**Alternativa Rejeitada**: Connection Pool no Core-Dict
- ❌ Core-Dict não deveria gerenciar rate limiting
- ❌ Violaria Separation of Concerns

---

## 📐 ARQUITETURA FINAL VALIDADA

### Fluxo Completo: CreateClaim

```
┌─────────────────────────────────────────────────────────────┐
│                    FLUXO COMPLETO                            │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  1. Frontend → Core-Dict REST API                            │
│     POST /claims                                             │
│     ↓                                                         │
│  2. CORE-DICT (Business Layer)                               │
│     ├─ Validações de negócio ✅                              │
│     │  ├─ Limite de claims (max 3)                           │
│     │  ├─ Validação de propriedade                           │
│     │  └─ Anti-fraude (FraudService)                         │
│     │                                                         │
│     ├─ Iniciar ClaimWorkflow Temporal ✅                     │
│     │  ├─ CreateClaim (state)                                │
│     │  ├─ NotifyOwner (business process)                     │
│     │  ├─ WaitTimer(7 dias) (business rule)                  │
│     │  ├─ DecideClaim (business logic)                       │
│     │  └─ CallConnectActivity ← AQUI chama infra             │
│     │     ↓                                                   │
│  3. CONN-DICT (Integration Layer)                            │
│     ├─ Connection Pool (acquire) ✅                          │
│     ├─ Circuit Breaker (check) ✅                            │
│     ├─ Transformação proto (Core → Bridge) ✅                │
│     ├─ gRPC call to Bridge ✅                                │
│     └─ Retry durável (se falhar) ✅                          │
│        ↓                                                      │
│  4. CONN-BRIDGE (Protocol Adapter)                           │
│     ├─ Proto → XML ✅                                         │
│     ├─ XML → SOAP envelope ✅                                │
│     ├─ Assinar XML (ICP-Brasil A3) ✅                        │
│     ├─ POST HTTPS + mTLS para Bacen ✅                       │
│     └─ Parse SOAP response ✅                                │
│        ↓                                                      │
│  5. Bacen DICT API                                           │
│     ✅ Claim criado no DICT                                  │
│        ↓                                                      │
│  6. Response volta (Bridge → Connect → Core → Frontend)     │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

**Separação Clara**:
- **Core-Dict**: Business logic, workflows complexos, decisões
- **Conn-Dict**: Connection management, retry, circuit breaker, protocol adaptation
- **Conn-Bridge**: SOAP/XML transformation, mTLS, Bacen API calls

---

## 🎓 CONCLUSÃO

### Resposta à Pergunta Original

> **"Workflows de negócio complexos (como Reivindicações) devem estar no Core-Dict ou Conn-Dict?"**

### ✅ RESPOSTA: CORE-DICT

**Workflows de negócio (ClaimWorkflow, PortabilityWorkflow) → CORE-DICT**
- Lógica de negócio complexa
- Decisões baseadas em contexto
- Orquestração multi-serviço
- Validações de domínio
- Compliance Bacen

**Infraestrutura técnica (Connection Pool, Retry, Circuit Breaker) → CONN-DICT**
- Gerenciamento de conexões
- Rate limiting Bacen
- Retry durável técnico
- Circuit breaker
- Transformação de protocolos

---

### Princípio Arquitetural Fundamental

```
┌─────────────────────────────────────────────────────────────┐
│             REGRA DE OURO (Golden Rule)                      │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  "Se a lógica precisa de CONTEXTO DE NEGÓCIO para decidir,  │
│   ela pertence ao CORE-DICT."                                │
│                                                               │
│  "Se a lógica é INFRAESTRUTURA TÉCNICA reutilizável,        │
│   ela pertence ao CONN-DICT."                                │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

---

**Última Atualização**: 2025-10-27 18:30 BRT
**Arquiteto**: Claude Sonnet 4.5
**Status**: ✅ ARQUITETURA VALIDADA E APROVADA
**Conformidade**: DDD, Hexagonal Architecture, Separation of Concerns
