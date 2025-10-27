# README: Onde Implementar Workflows de Negócio?

**Data**: 2025-10-27
**Versão**: 1.0
**Status**: ✅ DECISÃO ARQUITETURAL VALIDADA

---

## 🎯 RESPOSTA RÁPIDA

### Pergunta
> "Workflows de negócio complexos (como Reivindicações) devem estar no Core-Dict ou Conn-Dict?"

### Resposta
**WORKFLOWS DE NEGÓCIO → CORE-DICT** ✅

---

## 📐 REGRA DE OURO

```
┌────────────────────────────────────────────────────┐
│  Se precisa de CONTEXTO DE NEGÓCIO → CORE-DICT    │
│  Se é INFRAESTRUTURA TÉCNICA → CONN-DICT          │
│  Se é ADAPTAÇÃO DE PROTOCOLO → CONN-BRIDGE        │
└────────────────────────────────────────────────────┘
```

---

## 🏢 CORE-DICT: O Que Vai Aqui?

### ✅ Workflows de Negócio Complexos
- **ClaimWorkflow** (7-30 dias)
- **PortabilityWorkflow**
- **Qualquer workflow que tome decisões de negócio**

### ✅ Validações de Domínio
- Limite de chaves (max 5 por conta)
- Validação de duplicata
- Validação de ownership
- Detecção de fraude

### ✅ Orquestração de Processos
- Integração com múltiplos domínios (Fraud, User, Notification, Account)
- Decisões baseadas em contexto (histórico, perfil)
- Transações multi-step

### ✅ Estado de Negócio
- Histórico completo (audit logs)
- Compliance Bacen
- Rastreabilidade

### Por quê?
- **Tem contexto de negócio**
- **Acessa múltiplos domínios**
- **Toma decisões complexas**
- **Mantém estado rico**

---

## 🔌 CONN-DICT: O Que Vai Aqui?

### ✅ Connection Pool Management
- Gerencia rate limiting Bacen (1000 TPS)
- Load balancing
- Health checks

### ✅ Retry Durável (Temporal)
- Retry técnico (não business logic)
- Backoff exponencial
- Circuit breaker integration

### ✅ Circuit Breaker
- Proteção contra falhas em cascata
- Fail fast
- Evita sobrecarregar Bacen

### ✅ Transformação de Protocolo
- gRPC ↔ Pulsar
- Core proto ↔ Bridge proto
- Event serialization/deserialization

### ✅ Pulsar Event Handling
- Consume events do Core
- Produce events de volta para Core
- Ack/Nack management

### Por quê?
- **Não tem contexto de negócio**
- **Infraestrutura técnica reutilizável**
- **Transparente para Core-Dict**
- **Não toma decisões de negócio**

---

## 🌉 CONN-BRIDGE: O Que Vai Aqui?

### ✅ SOAP/XML Transformation
- gRPC → SOAP/XML
- XML → gRPC
- SOAP envelope building/parsing

### ✅ mTLS/ICP-Brasil
- Certificados A3
- Handshake mTLS
- Isolamento de segurança

### ✅ Assinatura Digital
- XML Signer (Java integration)
- ICP-Brasil compliance

### ✅ Chamadas HTTPS para Bacen
- POST/GET/PUT/DELETE
- SOAP over HTTPS
- Bacen API integration

### Por quê?
- **Adaptação de protocolo**
- **Único que "fala" com Bacen**
- **Core e Connect não conhecem SOAP**
- **Isolamento de certificados**

---

## 📊 TABELA RESUMO

| Responsabilidade | Core-Dict | Conn-Dict | Conn-Bridge |
|------------------|-----------|-----------|-------------|
| **ClaimWorkflow (7-30 dias)** | ✅ | ❌ | ❌ |
| **PortabilityWorkflow** | ✅ | ❌ | ❌ |
| **Validações de Negócio** | ✅ | ❌ | ❌ |
| **Detecção de Fraude** | ✅ | ❌ | ❌ |
| **Integração Multi-Domínio** | ✅ | ❌ | ❌ |
| **Audit Logs** | ✅ | ❌ | ❌ |
| **Connection Pool** | ❌ | ✅ | ❌ |
| **Retry Técnico** | ❌ | ✅ | ❌ |
| **Circuit Breaker** | ❌ | ✅ | ❌ |
| **gRPC ↔ Pulsar** | ❌ | ✅ | ❌ |
| **Event Handling** | ❌ | ✅ | ❌ |
| **SOAP/XML Transform** | ❌ | ❌ | ✅ |
| **mTLS/ICP-Brasil** | ❌ | ❌ | ✅ |
| **XML Signer** | ❌ | ❌ | ✅ |
| **HTTPS para Bacen** | ❌ | ❌ | ✅ |

---

## 🔍 EXEMPLO: ClaimWorkflow

### ✅ CORRETO: ClaimWorkflow no CORE-DICT

```go
// CORE-DICT: Temporal Workflow com lógica de negócio
func ClaimWorkflow(ctx workflow.Context, cmd CreateClaimCommand) error {
    // 1. Validações de negócio
    if !isEligible(cmd) {
        return ErrNotEligible // Business rule
    }

    // 2. Anti-fraude (integração com outro domínio)
    fraudScore := workflow.ExecuteActivity(ctx, CheckFraudActivity, cmd)
    if fraudScore > 0.8 {
        return ErrSuspiciousFraud // Business decision
    }

    // 3. Criar claim (business state)
    claim := CreateClaim(cmd)
    SaveClaim(claim) // Core-Dict PostgreSQL

    // 4. Notificar proprietário (business process)
    workflow.ExecuteActivity(ctx, NotifyOwnerActivity, claim)

    // 5. Aguardar 7 dias (business rule Bacen)
    workflow.Sleep(ctx, 7*24*time.Hour)

    // 6. Decidir resultado (business logic)
    response := GetOwnerResponse(claim.ID)
    if response == nil {
        claim.Status = "AUTO-CONFIRMED" // Business rule
    }

    // 7. AQUI chama Conn-Dict (infraestrutura)
    if claim.Status == "CONFIRMED" {
        workflow.ExecuteActivity(ctx, CallConnectActivity, claim)
    }

    // 8. Audit log (compliance)
    workflow.ExecuteActivity(ctx, AuditLogActivity, claim)

    return nil
}

// CORE-DICT: Activity que chama Conn-Dict
func CallConnectActivity(ctx context.Context, req CallConnectRequest) error {
    // Simples: apenas chama Conn-Dict
    // Conn-Dict lida com connection pool, retry, circuit breaker
    return s.connectClient.CompleteClaim(ctx, protoReq)
}
```

**Por que está correto?**
- ✅ Lógica de negócio (validações, decisões) no Core-Dict
- ✅ Conn-Dict chamado APENAS para executar no Bacen
- ✅ Separação clara: Business (Core) vs Infrastructure (Connect)

---

### ❌ ERRADO: ClaimWorkflow no CONN-DICT

```go
// ❌ ERRADO: ClaimWorkflow no CONN-DICT
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

**Por que está errado?**
- ❌ Conn-Dict não tem contexto de negócio
- ❌ Conn-Dict não deveria validar regras Bacen
- ❌ Violação de SoC (Separation of Concerns)
- ❌ Dificulta testes e manutenção

---

## 🏗️ FLUXO COMPLETO

```
Frontend → Core-Dict REST API
  ↓
Core-Dict (Business Layer)
  ├─ Validações de negócio ✅
  ├─ ClaimWorkflow Temporal ✅
  │  ├─ CreateClaim (state)
  │  ├─ NotifyOwner (business process)
  │  ├─ WaitTimer(7 dias) (business rule)
  │  ├─ DecideClaim (business logic)
  │  └─ CallConnectActivity ← chama infra
  ↓
Conn-Dict (Integration Layer)
  ├─ Connection Pool ✅
  ├─ Circuit Breaker ✅
  ├─ Retry Durável ✅
  └─ gRPC call to Bridge
  ↓
Conn-Bridge (Protocol Adapter)
  ├─ Proto → XML ✅
  ├─ XML → SOAP ✅
  ├─ Assinar XML ✅
  └─ POST HTTPS + mTLS
  ↓
Bacen DICT API ✅
```

---

## 📚 DOCUMENTAÇÃO COMPLETA

### Leitura Obrigatória
- **[ANALISE_SEPARACAO_RESPONSABILIDADES.md](ANALISE_SEPARACAO_RESPONSABILIDADES.md)** - Análise completa (842 LOC)
- **[STATUS_GLOBAL_ARQUITETURA_FINALIZADO.md](STATUS_GLOBAL_ARQUITETURA_FINALIZADO.md)** - Status consolidado

### Guias de Implementação
- [CONN_DICT_API_REFERENCE.md](CONN_DICT_API_REFERENCE.md) - API reference completo
- [STATUS_FINAL_2025-10-27.md](STATUS_FINAL_2025-10-27.md) - Instruções de integração

### Análises Técnicas
- [ANALISE_SYNC_VS_ASYNC_OPERATIONS.md](ANALISE_SYNC_VS_ASYNC_OPERATIONS.md) - Temporal vs Pulsar
- [ESCOPO_BRIDGE_VALIDADO.md](ESCOPO_BRIDGE_VALIDADO.md) - Bridge scope

---

## ✅ CHECKLIST PARA CORE-DICT

Quando implementar no core-dict, seguir este checklist:

### Workflows de Negócio
- [ ] ClaimWorkflow no Core-Dict (não no Conn-Dict) ✅
- [ ] PortabilityWorkflow no Core-Dict ✅
- [ ] Validações de negócio no Core-Dict ✅
- [ ] Detecção de fraude no Core-Dict ✅
- [ ] Audit logs no Core-Dict ✅

### Chamadas para Conn-Dict
- [ ] Core chama Conn-Dict apenas para executar no Bacen ✅
- [ ] Core não conhece detalhes de connection pool ✅
- [ ] Core não conhece detalhes de retry técnico ✅
- [ ] Core não conhece detalhes de circuit breaker ✅

### Separação de Concerns
- [ ] Business logic isolada de infraestrutura ✅
- [ ] Infraestrutura transparente para Core ✅
- [ ] Testável (mock Conn-Dict para testar business logic) ✅

---

## 🎓 PRINCÍPIOS APLICADOS

1. **Domain-Driven Design (DDD)**
   - Bounded Contexts: Core (Business), Connect (Integration), Bridge (Adapter)

2. **Hexagonal Architecture (Ports & Adapters)**
   - Core como hexágono central
   - Connect e Bridge como adapters externos

3. **Separation of Concerns (SoC)**
   - Business logic ≠ Infrastructure ≠ Protocol Adaptation

---

## 📞 DÚVIDAS?

Consultar documentos:
- [ANALISE_SEPARACAO_RESPONSABILIDADES.md](ANALISE_SEPARACAO_RESPONSABILIDADES.md) - Detalhes completos
- [STATUS_GLOBAL_ARQUITETURA_FINALIZADO.md](STATUS_GLOBAL_ARQUITETURA_FINALIZADO.md) - Status global

Ou perguntar diretamente ao Project Manager.

---

**Última Atualização**: 2025-10-27 19:00 BRT
**Arquiteto**: Claude Sonnet 4.5
**Status**: ✅ DECISÃO VALIDADA
**Paradigma**: DDD + Hexagonal + SoC
