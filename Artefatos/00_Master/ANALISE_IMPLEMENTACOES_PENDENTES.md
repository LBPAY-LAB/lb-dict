# 🔍 ANÁLISE DE IMPLEMENTAÇÕES PENDENTES - CORE-DICT

**Data**: 2025-10-27
**Projeto**: DICT LBPay - Core-Dict
**Versão**: 1.0.0
**Tipo**: Análise Técnica
**Baseado em**: Observação do usuário sobre TODOs no código

---

## 🎯 OBJETIVO

Analisar implementações **PARCIAIS** ou **PREPARADAS mas não funcionais** no Core-Dict que podem impactar o fluxo completo de transação PIX.

---

## 📊 SUMÁRIO EXECUTIVO

### Status das Implementações

```
╔════════════════════════════════════════════════════════════╗
║  IMPLEMENTAÇÃO COMPLETA:     ✅ 85%                        ║
║  PREPARADA MAS PENDENTE:     🟡 10%                        ║
║  NÃO IMPLEMENTADA:           🔴 5%                         ║
╚════════════════════════════════════════════════════════════╝
```

### TODOs Identificados

| # | Componente | Status | Impacto | Prioridade |
|---|------------|--------|---------|------------|
| **1** | GetEntryQueryHandler - Fallback RSFN | 🟡 Preparado (TODO linha 75-78) | MÉDIO | P2 |
| **2** | ConnectServiceAdapter - VerifyAccount | 🟡 Otimista (TODO linha 35) | ALTO | P1 |
| **3** | LookupKey - Rate Limiting | 🔴 Não implementado | ALTO | P1 |

---

## 📑 ÍNDICE

1. [TODO 1: Fallback RSFN Connect](#todo-1-fallback-rsfn-connect)
2. [TODO 2: VerifyAccount Real Implementation](#todo-2-verifyaccount-real-implementation)
3. [TODO 3: Rate Limiting no LookupKey](#todo-3-rate-limiting-no-lookupkey)
4. [Fluxo PIX Completo - Validação](#fluxo-pix-completo---validação)
5. [Matriz de Impacto](#matriz-de-impacto)
6. [Plano de Implementação](#plano-de-implementação)

---

## 🟡 TODO 1: Fallback RSFN Connect

### Localização

**Arquivo**: `core-dict/internal/application/queries/get_entry_query.go`
**Linhas**: 72-79

```go
// 3. Database miss - try Connect service as fallback (optional)
// This allows querying RSFN DICT directly for keys owned by other ISPBs
if h.connectClient != nil {
    // TODO: Connect client would need a method to query by key value
    // For now, just return the database error
    // Future: rsfnEntry, err := h.connectClient.GetEntryByKey(ctx, query.KeyValue)
    // if err == nil { return mapRSFNEntryToDomain(rsfnEntry), nil }
}
```

### Análise

**O que está implementado**:
- ✅ Estrutura do handler
- ✅ Cache-Aside pattern (Redis)
- ✅ Query PostgreSQL local
- ✅ Preparação para fallback (if h.connectClient != nil)

**O que está PENDENTE**:
- ❌ Método `ConnectClient.GetEntryByKey()` não existe
- ❌ Mapeamento `mapRSFNEntryToDomain()` não implementado
- ❌ Integração com RSFN Bacen para chaves de outros ISPBs

### Impacto

**MÉDIO** - Afeta cenário específico:

**Cenário Afetado**:
```
Usuário LBPay quer enviar PIX para chave CPF 12345678900
→ Chave NÃO está no PostgreSQL do LBPay (pertence a outro banco)
→ Sistema retorna "key not found"
→ ❌ DEVERIA consultar RSFN Bacen (via Connect) e encontrar a chave
```

**Cenário NÃO Afetado**:
```
Usuário LBPay quer enviar PIX para chave CPF 98765432100
→ Chave ESTÁ no PostgreSQL do LBPay
→ Sistema encontra localmente
→ ✅ Funciona normalmente
```

### Workaround Atual

Para chaves de outros ISPBs, o **FrontEnd** ou **Core de Pagamentos** pode:
1. Chamar diretamente o **Conn-Dict** (via gRPC)
2. Conn-Dict consulta RSFN Bacen
3. Retorna informações da chave

**Problema**: Core-Dict não é self-sufficient para lookup de chaves externas.

### Prioridade

**P2 - Média**

**Por quê P2?**
- Core-Dict funciona para chaves locais (caso principal)
- Workaround via Conn-Dict existe
- Não bloqueia produção
- Melhoria de UX (latência menor se resolver internamente)

---

## 🟡 TODO 2: VerifyAccount Real Implementation

### Localização

**Arquivo**: `core-dict/internal/infrastructure/adapters/connect_service_adapter.go`
**Linhas**: 21-38

```go
// VerifyAccount implements ConnectService.VerifyAccount
// For now, it returns true (optimistic) since ConnectClient doesn't have this method yet
func (a *ConnectServiceAdapter) VerifyAccount(ctx context.Context, ispb, branch, accountNumber string) (bool, error) {
	// If client is nil, assume account is valid (optimistic verification)
	if a.client == nil {
		return true, nil
	}

	// Check if Connect service is reachable
	if err := a.client.HealthCheck(ctx); err != nil {
		// If Connect is down, still allow operation (degraded mode)
		return true, nil
	}

	// TODO: When ConnectClient has VerifyAccount method, use it here
	// For now, assume account is valid if Connect is reachable
	return true, nil
}
```

### Análise

**O que está implementado**:
- ✅ Interface `ConnectService.VerifyAccount()`
- ✅ Adapter pattern para bridge de interfaces
- ✅ Graceful degradation (assume válido se Connect indisponível)
- ✅ HealthCheck do Connect

**O que está PENDENTE**:
- ❌ Validação REAL de conta via RSFN Bacen
- ❌ Método `ConnectClient.VerifyAccount()` não existe
- ❌ Integração com Bridge → Bacen SOAP API

**Comportamento Atual**:
```go
// SEMPRE retorna true (otimista)
verifyAccount("12345678", "0001", "123456") → true ✅
verifyAccount("99999999", "9999", "999999") → true ✅ (conta inválida, mas aceita!)
```

### Impacto

**ALTO** - Afeta validação de negócio crítica:

**Cenário Afetado**:
```
Usuário cria chave PIX com conta INVÁLIDA
→ Core-Dict chama VerifyAccount()
→ Retorna true (otimista) ❌
→ Chave é CRIADA com conta inválida
→ Transações PIX para essa chave vão FALHAR
```

**Validação que DEVERIA acontecer**:
1. Core-Dict → Conn-Dict → Bridge → Bacen SOAP
2. Bacen verifica se conta existe no ISPB
3. Bacen retorna true/false
4. Se false → CreateKey falha com erro claro

### Workaround Atual

**Nenhum workaround real** - o sistema está aceitando contas inválidas.

**Mitigação**:
- FrontEnd valida formato (ISPB 8 dígitos, etc.)
- Testes E2E usam apenas contas válidas
- Em produção, **primeira transação PIX vai falhar** e usuário descobre que conta está errada

**Risco**: Má experiência do usuário + poluição do banco com chaves inválidas.

### Prioridade

**P1 - Alta**

**Por quê P1?**
- Validação de negócio crítica
- Bacen pode reprovar sistema em homologação
- Má experiência do usuário (chave criada mas não funciona)
- Sem workaround viável

---

## 🔴 TODO 3: Rate Limiting no LookupKey

### Localização

**Arquivo**: `core-dict/internal/infrastructure/grpc/core_dict_service_handler.go`
**Método**: `LookupKey()` (linhas 871-932)

**Endpoint**: Público (sem autenticação)

### Análise

**O que está implementado**:
- ✅ Método LookupKey funcional
- ✅ Validação de formato
- ✅ Query handler (cache + database)
- ✅ Resposta com dados públicos apenas

**O que está PENDENTE**:
- ❌ **NENHUM rate limiting** no endpoint
- ❌ Proteção contra abuso/scraping
- ❌ Rate limit por IP (TODO do GAP 2)
- ❌ Rate limit por ISPB

**Vulnerabilidade Atual**:
```bash
# Atacante pode fazer scraping do banco inteiro:
for key in {00000000000..99999999999}; do
    grpcurl -d '{"key": {"key_value": "'$key'"}}' localhost:9090 LookupKey
done

# Sem rate limiting → 100.000 requests/segundo
# Todo o banco de chaves exposto em minutos
```

### Impacto

**ALTO** - Segurança e Performance:

**Cenário de Ataque**:
1. Atacante descobre endpoint público `LookupKey`
2. Faz 1.000.000 requests testando CPFs sequenciais
3. Extrai todas as chaves PIX do banco
4. Vazamento de dados pessoais (chave + nome + conta)

**Impacto Regulatório**:
- ❌ Não conforme LGPD (dados pessoais expostos)
- ❌ Não conforme Bacen (falta proteção anti-abuse)
- ❌ Risco de multa ANPD (até 2% faturamento)

### Workaround Atual

**Nenhum** - Endpoint está completamente desprotegido.

**Possíveis mitigações**:
1. Kubernetes Ingress rate limiting (layer 7)
2. Cloudflare rate limiting (se existir)
3. WAF (Web Application Firewall)

**Problema**: Proteção externa não é suficiente - deve ter rate limiting no código.

### Prioridade

**P1 - Alta (CRÍTICA para produção)**

**Por quê P1?**
- Endpoint PÚBLICO sem proteção
- Risco de segurança ALTO (scraping de dados)
- Não conforme LGPD + Bacen
- Blocker para produção
- Já identificado como **GAP 2** no plano de conformidade

---

## 📊 FLUXO PIX COMPLETO - VALIDAÇÃO

### Fluxo Atual (Com TODOs)

```
╔════════════════════════════════════════════════════════════════╗
║  FLUXO: Enviar R$ 100 para chave CPF 12345678900              ║
╠════════════════════════════════════════════════════════════════╣
║                                                                 ║
║  PASSO 1: Validar Formato                                      ║
║  ├─ KeyValidatorService.ValidateFormat()                       ║
║  ├─ ✅ IMPLEMENTADO (linha 47-62)                              ║
║  └─ Result: CPF válido                                         ║
║                                                                 ║
║  PASSO 2: Verificar Existência (Lookup)                        ║
║  ├─ GetEntryQueryHandler.Handle()                              ║
║  ├─ ✅ Cache → ✅ PostgreSQL                                   ║
║  ├─ 🟡 TODO: Fallback RSFN (linha 75-78)                       ║
║  └─ Result: Chave encontrada (se local)                        ║
║                                                                 ║
║  PASSO 3: Verificar Status                                     ║
║  ├─ entry.Status == "ACTIVE"?                                  ║
║  ├─ ✅ IMPLEMENTADO                                            ║
║  └─ Result: Chave ativa                                        ║
║                                                                 ║
║  PASSO 4: Verificar Conta Destino                              ║
║  ├─ VerifyAccountQueryHandler.Handle()                         ║
║  ├─ 🟡 TODO: ConnectServiceAdapter retorna sempre true         ║
║  ├─ Deveria: Core → Conn → Bridge → Bacen SOAP                ║
║  └─ Result: Conta "válida" (otimista, não verificada)          ║
║                                                                 ║
║  PASSO 5: Processar Transação PIX                              ║
║  ├─ Core de Pagamentos (fora do Core-Dict)                     ║
║  └─ ✅ Se chegou aqui, dados estão válidos                     ║
║                                                                 ║
╚════════════════════════════════════════════════════════════════╝
```

### O que Funciona

✅ **Chaves Locais (mesmo ISPB)**:
- Validação de formato
- Lookup local (cache + database)
- Verificação de status
- Fluxo completo OK

✅ **Mock Mode**:
- Tudo funciona (dados mockados)
- Testes E2E passam

### O que NÃO Funciona Completamente

🟡 **Chaves de Outros ISPBs**:
- Lookup local falha
- Fallback RSFN não implementado
- Retorna "key not found" (deveria buscar no Bacen)

🟡 **Validação de Conta**:
- Sempre retorna true (otimista)
- Não valida se conta realmente existe no ISPB
- Risco de criar chaves com contas inválidas

🔴 **Rate Limiting**:
- Nenhuma proteção em endpoint público LookupKey
- Risco de scraping e abuso

---

## 📊 MATRIZ DE IMPACTO

| TODO | Componente | Impacto Produção | Impacto Homologação | Impacto UX | Conformidade Bacen | Prioridade |
|------|------------|------------------|---------------------|------------|-------------------|------------|
| **TODO 1** | Fallback RSFN | 🟡 Médio (workaround existe) | 🟡 Médio | 🟡 Médio (latência) | ✅ OK | **P2** |
| **TODO 2** | VerifyAccount | 🔴 Alto (chaves inválidas) | 🔴 Alto (pode reprovar) | 🔴 Alto (chaves não funcionam) | ❌ Não conforme | **P1** |
| **TODO 3** | Rate Limiting | 🔴 CRÍTICO (segurança) | 🔴 CRÍTICO | 🟡 Médio | ❌ Não conforme LGPD | **P1** |

### Legenda
- 🔴 Alto: Bloqueia produção ou causa problemas graves
- 🟡 Médio: Funciona mas com limitações
- ✅ Baixo: Não afeta

---

## 📋 PLANO DE IMPLEMENTAÇÃO

### Sprint +1 (Semana 1-2): TODOs Críticos

**TODO 3: Rate Limiting no LookupKey** (incluso no GAP 2)
- Esforço: 2 dias (já planejado)
- Benefício: Endpoint público protegido
- Status: Planejado no [PLANO_IMPLEMENTACAO_GAPS_CONFORMIDADE.md](PLANO_IMPLEMENTACAO_GAPS_CONFORMIDADE.md)

**TODO 2: VerifyAccount Real Implementation**
- Esforço: 3 dias
- Dependência: Conn-Dict precisa ter método `VerifyAccount()`
- Benefício: Validação real de contas

**Tarefas**:
```go
// 1. Criar método no ConnectClient (core-dict)
type ConnectClient interface {
    GetEntryByKey(ctx context.Context, keyValue string) (*Entry, error)
    VerifyAccount(ctx context.Context, ispb, branch, accountNumber string) (bool, error)  // [NOVO]
    HealthCheck(ctx context.Context) error
}

// 2. Implementar gRPC client stub (core-dict)
func (c *ConnectGRPCClient) VerifyAccount(ctx context.Context, ispb, branch, accountNumber string) (bool, error) {
    req := &connectv1.VerifyAccountRequest{
        Ispb:          ispb,
        Branch:        branch,
        AccountNumber: accountNumber,
    }
    resp, err := c.client.VerifyAccount(ctx, req)
    if err != nil {
        return false, err
    }
    return resp.Valid, nil
}

// 3. Modificar ConnectServiceAdapter para usar método real
func (a *ConnectServiceAdapter) VerifyAccount(ctx context.Context, ispb, branch, accountNumber string) (bool, error) {
    if a.client == nil {
        return true, nil  // Fallback
    }

    // [MODIFICAR] Chamar método real ao invés de HealthCheck
    valid, err := a.client.VerifyAccount(ctx, ispb, branch, accountNumber)
    if err != nil {
        // Se Connect falhou, modo degradado (assume válido)
        return true, nil
    }

    return valid, nil
}
```

### Sprint +2 (Semana 3-4): TODOs de Melhoria

**TODO 1: Fallback RSFN Connect**
- Esforço: 2 dias
- Dependência: Conn-Dict precisa ter método `GetEntryByKey()`
- Benefício: Lookup de chaves de outros ISPBs

**Tarefas**:
```go
// 1. Criar método no ConnectClient (core-dict)
type ConnectClient interface {
    GetEntryByKey(ctx context.Context, keyValue string) (*Entry, error)  // [NOVO]
    VerifyAccount(ctx context.Context, ispb, branch, accountNumber string) (bool, error)
    HealthCheck(ctx context.Context) error
}

// 2. Implementar gRPC client stub (core-dict)
func (c *ConnectGRPCClient) GetEntryByKey(ctx context.Context, keyValue string) (*Entry, error) {
    req := &connectv1.GetEntryByKeyRequest{
        KeyValue: keyValue,
    }
    resp, err := c.client.GetEntryByKey(ctx, req)
    if err != nil {
        return nil, err
    }
    return mapProtoToDomain(resp.Entry), nil
}

// 3. Modificar GetEntryQueryHandler para usar fallback
func (h *GetEntryQueryHandler) Handle(ctx context.Context, query GetEntryQuery) (*entities.Entry, error) {
    // ... cache + database ...

    // [MODIFICAR] Fallback RSFN
    if h.connectClient != nil {
        rsfnEntry, err := h.connectClient.GetEntryByKey(ctx, query.KeyValue)
        if err == nil {
            // Cache RSFN result (shorter TTL: 1 minute)
            _ = h.cache.Set(ctx, cacheKey, rsfnEntry, 1*time.Minute)
            return rsfnEntry, nil
        }
    }

    return nil, fmt.Errorf("entry not found")
}
```

---

## ✅ CHECKLIST DE VALIDAÇÃO

### Para Homologação Bacen

- [x] Validação de formato de chaves (✅ Implementado)
- [ ] **Lookup de chaves de outros ISPBs** (🟡 TODO 1 - P2)
- [ ] **Validação REAL de contas** (🟡 TODO 2 - **P1 CRÍTICO**)
- [ ] **Rate limiting em endpoints públicos** (🔴 TODO 3 - **P1 CRÍTICO**)
- [x] Verificação de status de chaves (✅ Implementado)
- [x] Cache para performance (✅ Implementado)

### Para Produção

- [ ] **TODO 2**: VerifyAccount real (BLOCKER)
- [ ] **TODO 3**: Rate Limiting no LookupKey (BLOCKER)
- [ ] TODO 1: Fallback RSFN (nice-to-have)

---

## 📞 RECOMENDAÇÕES

### Ação Imediata (Antes de Homologação)

1. ✅ **Implementar TODO 3** (Rate Limiting) - GAP 2 já planejado
2. ✅ **Implementar TODO 2** (VerifyAccount) - Adicionar ao Sprint +1

### Ação Médio Prazo (Antes de Produção)

3. ✅ **Implementar TODO 1** (Fallback RSFN) - Sprint +2

### Priorização Sugerida

```
Sprint +1 (Semanas 1-2):
├─ GAP 2: Rate Limiting por IP (2 dias) ← inclui TODO 3
├─ GAP 3: Circuit Breaker (3 dias)
└─ TODO 2: VerifyAccount Real (3 dias) ← ADICIONAR

Sprint +2 (Semanas 3-4):
├─ GAP 1: OTP Validation (5 dias)
└─ TODO 1: Fallback RSFN (2 dias) ← ADICIONAR
```

---

## 📚 REFERÊNCIAS

### Código Analisado

- [key_validator_service.go](../../core-dict/internal/application/services/key_validator_service.go)
- [get_entry_query.go](../../core-dict/internal/application/queries/get_entry_query.go)
- [verify_account_query.go](../../core-dict/internal/application/queries/verify_account_query.go)
- [connect_service_adapter.go](../../core-dict/internal/infrastructure/adapters/connect_service_adapter.go)
- [core_dict_service_handler.go](../../core-dict/internal/infrastructure/grpc/core_dict_service_handler.go)

### Documentos Relacionados

- [PLANO_IMPLEMENTACAO_GAPS_CONFORMIDADE.md](PLANO_IMPLEMENTACAO_GAPS_CONFORMIDADE.md) - GAP 2 (Rate Limiting)
- [RELATORIO_CONFORMIDADE_BACEN_CORE_DICT.md](RELATORIO_CONFORMIDADE_BACEN_CORE_DICT.md) - 95% conformidade
- [SESSAO_2025-10-27_CONTEXTO_COMPLETO.md](SESSAO_2025-10-27_CONTEXTO_COMPLETO.md) - Contexto da sessão

---

**Última Atualização**: 2025-10-27
**Versão**: 1.0.0
**Status**: ✅ **ANÁLISE COMPLETA**

---

**FIM DA ANÁLISE**
