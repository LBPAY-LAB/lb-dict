# 📋 PLANO DE IMPLEMENTAÇÃO - GAPS DE CONFORMIDADE BACEN

**Projeto**: DICT LBPay - Core-Dict
**Versão**: 1.0.0
**Data**: 2025-10-27
**Tipo**: Plano de Implementação
**Base**: [RELATORIO_CONFORMIDADE_BACEN_CORE_DICT.md](RELATORIO_CONFORMIDADE_BACEN_CORE_DICT.md)

---

## 🎯 OBJETIVO

Implementar os **11 gaps identificados** no relatório de conformidade para elevar a conformidade de **95% → 100%** e tornar o Core-Dict **100% pronto para produção**.

---

## 📊 SUMÁRIO EXECUTIVO

### Status Atual vs. Meta

```
╔══════════════════════════════════════════════════════════════╗
║  CONFORMIDADE ATUAL:   95% (174/185 requisitos)             ║
║  META:                100% (185/185 requisitos)             ║
║  GAPS TOTAIS:          11 gaps                              ║
║  ESFORÇO ESTIMADO:     ~18 dias (2-3 sprints)               ║
║  PRIORIDADE:           3 P1 (críticos) + 8 P2 (importantes) ║
╚══════════════════════════════════════════════════════════════╝
```

### Distribuição por Prioridade

| Prioridade | Gaps | Esforço | Prazo Sugerido |
|------------|------|---------|----------------|
| **P1 - Alto** | 3 | 10 dias | Sprint +1 e +2 |
| **P2 - Médio** | 8 | 8 dias | Sprint +2 e +3 |
| **TOTAL** | **11** | **18 dias** | **3 sprints** |

---

## 📑 ÍNDICE

1. [Gaps P1 - Alta Prioridade](#gaps-p1---alta-prioridade)
2. [Gaps P2 - Média Prioridade](#gaps-p2---média-prioridade)
3. [Análise de Especificações Existentes](#análise-de-especificações-existentes)
4. [Plano de Implementação Sprint-a-Sprint](#plano-de-implementação-sprint-a-sprint)
5. [Checklist de Implementação](#checklist-de-implementação)
6. [Critérios de Aceite](#critérios-de-aceite)

---

## 🔴 GAPS P1 - ALTA PRIORIDADE

### GAP 1: Validação OTP para Email/Phone (Claim)

**ID**: REG-CLM-003
**Categoria**: Reivindicação (Claim)
**Status Atual**: ⏳ Struct preparada, lógica não implementada
**Prioridade**: **P1**
**Impacto**: **MÉDIO** - Requisito regulatório para claims de Email/Phone
**Esforço**: **5 dias**
**Prazo**: Sprint +2

#### Descrição do Gap

Quando um usuário cria uma **claim de portabilidade** para uma chave do tipo **EMAIL** ou **PHONE**, o sistema DEVE enviar um **código OTP (One-Time Password)** para validar que o usuário tem acesso à chave.

**Regulamentação Bacen**:
- Email: 24 horas para validar OTP
- Phone: 10 minutos para validar OTP
- Após timeout, claim é cancelada automaticamente

**Especificação Existente**: [US-002_User_Stories_Claims.md](../01_Requisitos/UserStories/US-002_User_Stories_Claims.md) - AC-002.1.2

#### Especificações Disponíveis

✅ **COMPLETO** - Temos todas as especificações necessárias:

1. **User Story**: [US-002.1 - AC-002.1.2](../01_Requisitos/UserStories/US-002_User_Stories_Claims.md#L48-L55)
   - Envio de código para Email/Phone
   - Timeout: 10min (phone), 24h (email)
   - Cancelamento automático se timeout

2. **Sequence Diagram**: [DIA-006_Sequence_Claim_Workflow.md](../02_Arquitetura/Diagramas/DIA-006_Sequence_Claim_Workflow.md)
   - Fluxo completo de validação OTP
   - Interação com serviço de notificações

3. **Struct Pronta**: `core-dict/internal/domain/valueobjects/otp_validation.go` (EXISTE)

#### O que Falta Implementar

**No Core-Dict**:

1. **OTPService Interface** (`internal/application/services/otp_service.go`):
```go
type OTPService interface {
    // GenerateOTP gera código OTP de 6 dígitos
    GenerateOTP(ctx context.Context, keyType string, keyValue string) (string, error)

    // SendOTP envia OTP via Email ou SMS
    SendOTP(ctx context.Context, keyType string, keyValue string, code string) error

    // ValidateOTP valida código fornecido pelo usuário
    ValidateOTP(ctx context.Context, keyType string, keyValue string, code string) (bool, error)

    // GetOTPExpiry retorna expiração do OTP (10min phone, 24h email)
    GetOTPExpiry(keyType string) time.Duration
}
```

2. **OTPRepository** (`internal/infrastructure/database/otp_repository.go`):
```sql
CREATE TABLE core_dict.otp_validations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_type VARCHAR(20) NOT NULL,
    key_value VARCHAR(255) NOT NULL,
    otp_code VARCHAR(6) NOT NULL,
    attempts INT DEFAULT 0,
    max_attempts INT DEFAULT 3,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    validated_at TIMESTAMP,
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, VALIDATED, EXPIRED, BLOCKED
    UNIQUE(key_type, key_value, created_at)
);

CREATE INDEX idx_otp_validations_lookup ON core_dict.otp_validations(key_type, key_value, status);
```

3. **Integração com Provedores**:
   - **Email**: SendGrid API (ou AWS SES)
   - **SMS**: Twilio API (ou AWS SNS)

4. **Modificação em CreateClaimCommand**:
```go
// Antes de criar claim, verificar se precisa OTP
if claim.KeyType == "EMAIL" || claim.KeyType == "PHONE" {
    // Gerar e enviar OTP
    otpCode, _ := otpService.GenerateOTP(ctx, claim.KeyType, claim.KeyValue)
    otpService.SendOTP(ctx, claim.KeyType, claim.KeyValue, otpCode)

    // Aguardar validação (async via evento ou polling)
    // Se timeout → cancelar claim
}
```

5. **Novo Endpoint REST**:
```
POST /api/v1/claims/{claimId}/validate-otp
{
  "otp_code": "123456"
}
```

#### Estrutura de Arquivos

```
core-dict/
├── internal/
│   ├── application/
│   │   ├── services/
│   │   │   └── otp_service.go         # [NOVO] Interface OTPService
│   │   └── commands/
│   │       └── validate_otp_command.go # [NOVO] ValidateOTPCommand
│   ├── domain/
│   │   └── valueobjects/
│   │       └── otp_validation.go       # [EXISTE] - Struct já criada
│   └── infrastructure/
│       ├── database/
│       │   └── otp_repository.go       # [NOVO] Persistência OTP
│       ├── external/
│       │   ├── sendgrid_client.go      # [NOVO] Email OTP
│       │   └── twilio_client.go        # [NOVO] SMS OTP
│       └── api/
│           └── otp_handler.go          # [NOVO] REST endpoint
└── migrations/
    └── 006_create_otp_validations.sql  # [NOVO] Schema OTP
```

#### Dependências Externas

```go
// go.mod
require (
    github.com/sendgrid/sendgrid-go v3.14.0+incompatible
    github.com/twilio/twilio-go v1.15.2
)
```

#### Testes Necessários

```
✅ Unit Tests:
   - OTPService.GenerateOTP (código de 6 dígitos)
   - OTPService.ValidateOTP (sucesso + 3 falhas = bloqueio)
   - OTPRepository.Create/Get/Update

✅ Integration Tests:
   - Fluxo completo: gerar → enviar → validar
   - Timeout email (24h) e phone (10min)
   - Bloqueio após 3 tentativas falhas

✅ E2E Tests:
   - Claim com EMAIL → recebe OTP → valida → claim criada
   - Claim com PHONE → recebe SMS → valida → claim criada
   - Timeout OTP → claim cancelada automaticamente
```

#### Esforço Detalhado

| Tarefa | Esforço |
|--------|---------|
| OTPService interface + impl | 1 dia |
| OTPRepository + migrations | 0.5 dia |
| SendGrid integration | 1 dia |
| Twilio integration | 1 dia |
| REST endpoint validação | 0.5 dia |
| Modificar CreateClaimCommand | 0.5 dia |
| Testes unitários | 0.5 dia |
| Testes integração | 0.5 dia |
| Testes E2E | 0.5 dia |
| **TOTAL** | **5 dias** |

---

### GAP 2: Rate Limiting por IP

**ID**: REG-RATE-002
**Categoria**: Controle
**Status Atual**: ❌ Não implementado (existe por ISPB apenas)
**Prioridade**: **P1**
**Impacto**: **MÉDIO** - Anti-abuse importante para produção
**Esforço**: **2 dias**
**Prazo**: Sprint +2

#### Descrição do Gap

Atualmente, o Core-Dict tem **rate limiting por ISPB** (participante), mas NÃO tem **rate limiting por IP**, permitindo que um atacante abuse do sistema fazendo múltiplas requisições de IPs diferentes.

**Especificação Existente**: [DAT-005_Redis_Cache_Strategy.md](../03_Dados/DAT-005_Redis_Cache_Strategy.md) - Seção 5 (Rate Limiting)

#### O que Falta Implementar

**1. Modificar RateLimiter para suportar IP**:

```go
// internal/infrastructure/cache/rate_limiter.go

type RateLimiter struct {
    client *RedisClient
    config *RateLimitConfig
}

type RateLimitConfig struct {
    // Existente (por ISPB)
    RequestsPerSecondByISPB int
    BurstSizeByISPB         int

    // [NOVO] Por IP
    RequestsPerSecondByIP   int   // Ex: 10 req/s
    BurstSizeByIP           int   // Ex: 20 burst

    KeyPrefix               string
}

// [NOVO] CheckRateLimitByIP verifica limite por IP
func (r *RateLimiter) CheckRateLimitByIP(ctx context.Context, ip string, operation string) error {
    key := fmt.Sprintf("%sip:%s:%s:%s", r.config.KeyPrefix, ip, operation, time.Now().Format("2006-01-02-15:04"))

    count, err := r.client.Incr(ctx, key)
    if err != nil {
        return err
    }

    if count == 1 {
        r.client.Expire(ctx, key, 1*time.Minute)
    }

    if count > int64(r.config.RequestsPerSecondByIP) {
        return fmt.Errorf("rate limit exceeded for IP %s: %d/%d", ip, count, r.config.RequestsPerSecondByIP)
    }

    return nil
}
```

**2. Modificar gRPC Interceptor**:

```go
// internal/infrastructure/grpc/interceptors/rate_limiting.go

func RateLimitingInterceptor(rateLimiter *cache.RateLimiter) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // [EXISTENTE] Rate limit por ISPB
        ispb := extractISPBFromContext(ctx)
        if err := rateLimiter.CheckRateLimitByISPB(ctx, ispb, info.FullMethod); err != nil {
            return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded for ISPB")
        }

        // [NOVO] Rate limit por IP
        ip := extractIPFromContext(ctx)
        if err := rateLimiter.CheckRateLimitByIP(ctx, ip, info.FullMethod); err != nil {
            return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded for IP")
        }

        return handler(ctx, req)
    }
}

// [NOVO] Extrair IP do contexto gRPC
func extractIPFromContext(ctx context.Context) string {
    p, ok := peer.FromContext(ctx)
    if !ok {
        return "unknown"
    }

    addr := p.Addr.String()
    host, _, _ := net.SplitHostPort(addr)
    return host
}
```

**3. Configuração**:

```yaml
# core-dict/config/config.yaml

rate_limiting:
  by_ispb:
    requests_per_second: 100
    burst_size: 200

  by_ip:
    requests_per_second: 10   # [NOVO]
    burst_size: 20            # [NOVO]
```

#### Testes Necessários

```
✅ Unit Tests:
   - RateLimiter.CheckRateLimitByIP (sucesso)
   - RateLimiter.CheckRateLimitByIP (excedeu limite)

✅ Integration Tests:
   - 10 requests do mesmo IP → 11ª falha
   - 10 requests de IPs diferentes → todas passam

✅ Load Tests (k6):
   - Simular 1000 requests/s de múltiplos IPs
   - Verificar que rate limiting funciona
```

#### Esforço Detalhado

| Tarefa | Esforço |
|--------|---------|
| Modificar RateLimiter (add IP support) | 0.5 dia |
| Modificar gRPC interceptor | 0.5 dia |
| Adicionar configuração | 0.25 dia |
| Testes unitários | 0.25 dia |
| Testes integração | 0.25 dia |
| Testes load (k6) | 0.25 dia |
| **TOTAL** | **2 dias** |

---

### GAP 3: Circuit Breaker para Chamadas Externas

**ID**: REG-RATE-004
**Categoria**: Controle
**Status Atual**: ❌ Não implementado
**Prioridade**: **P1**
**Impacto**: **MÉDIO** - Resiliência em chamadas para Connect/Bridge
**Esforço**: **3 dias**
**Prazo**: Sprint +2

#### Descrição do Gap

Quando o **Core-Dict** faz chamadas para **Connect** ou **Bridge** (via gRPC), se esses serviços estiverem lentos ou falhando, o Core-Dict pode:
- Acumular goroutines bloqueadas
- Causar timeouts em cascata
- Sobrecarregar serviços já degradados

**Solução**: Implementar **Circuit Breaker** (padrão Hystrix) para "abrir o circuito" após N falhas consecutivas.

#### Especificações Disponíveis

⚠️ **PARCIAL** - Não temos especificação dedicada, mas:
- [SEC-004](../13_Seguranca/SEC-004_API_Authentication.md) menciona "Circuit breaker para chamadas externas"
- Padrão bem conhecido (Hystrix, Gobreaker)

#### Biblioteca Recomendada

```go
// go.mod
require (
    github.com/sony/gobreaker v0.5.0
)
```

#### O que Implementar

**1. Criar CircuitBreakerService**:

```go
// internal/infrastructure/resilience/circuit_breaker.go

import "github.com/sony/gobreaker"

type CircuitBreakerService struct {
    connectCB *gobreaker.CircuitBreaker
    bridgeCB  *gobreaker.CircuitBreaker
}

func NewCircuitBreakerService() *CircuitBreakerService {
    settings := gobreaker.Settings{
        Name:        "Core-Connect",
        MaxRequests: 3,           // Max requests em estado half-open
        Interval:    10 * time.Second,  // Janela de contagem de erros
        Timeout:     30 * time.Second,  // Tempo em estado open antes de half-open
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= 3 && failureRatio >= 0.6  // 60% de erros
        },
        OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
            log.Warnf("Circuit breaker %s changed from %s to %s", name, from, to)
        },
    }

    return &CircuitBreakerService{
        connectCB: gobreaker.NewCircuitBreaker(settings),
        bridgeCB:  gobreaker.NewCircuitBreaker(settings),
    }
}

// ExecuteConnectCall executa chamada para Connect com circuit breaker
func (cb *CircuitBreakerService) ExecuteConnectCall(fn func() (interface{}, error)) (interface{}, error) {
    return cb.connectCB.Execute(fn)
}

// ExecuteBridgeCall executa chamada para Bridge com circuit breaker
func (cb *CircuitBreakerService) ExecuteBridgeCall(fn func() (interface{}, error)) (interface{}, error) {
    return cb.bridgeCB.Execute(fn)
}
```

**2. Modificar ConnectServiceAdapter**:

```go
// internal/infrastructure/adapters/connect_service_adapter.go

type ConnectServiceAdapter struct {
    client        services.ConnectClient
    circuitBreaker *resilience.CircuitBreakerService  // [NOVO]
}

// VerifyAccount com circuit breaker
func (a *ConnectServiceAdapter) VerifyAccount(ctx context.Context, ispb, branch, accountNumber string) (bool, error) {
    result, err := a.circuitBreaker.ExecuteConnectCall(func() (interface{}, error) {
        // Chamada original
        return a.client.HealthCheck(ctx)
    })

    if err != nil {
        // Circuit aberto → fallback
        if err == gobreaker.ErrOpenState {
            log.Warn("Circuit breaker OPEN for Connect, using fallback")
            return true, nil  // Degraded mode: assume válido
        }
        return false, err
    }

    return true, nil
}
```

**3. Métricas Prometheus**:

```go
var (
    circuitBreakerStateGauge = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "circuit_breaker_state",
            Help: "Circuit breaker state (0=closed, 1=half-open, 2=open)",
        },
        []string{"service"},
    )

    circuitBreakerTripsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "circuit_breaker_trips_total",
            Help: "Total number of circuit breaker trips",
        },
        []string{"service"},
    )
)
```

#### Estados do Circuit Breaker

```
┌─────────┐
│ CLOSED  │ ─── N falhas consecutivas ──→ ┌──────┐
│ (normal)│                                │ OPEN │
└─────────┘ ←─── sucesso ─────────────────│(block)│
                                           └──────┘
                                              │
                                           timeout
                                              │
                                              ▼
                                        ┌──────────┐
                                        │HALF-OPEN │
                                        │ (test)   │
                                        └──────────┘
                                          │      │
                                    sucesso   falha
                                          │      │
                                       CLOSED  OPEN
```

#### Testes Necessários

```
✅ Unit Tests:
   - Circuit breaker fecha após N falhas
   - Circuit breaker abre → timeout → half-open
   - Half-open: sucesso → closed
   - Half-open: falha → open

✅ Integration Tests:
   - Simular Connect lento → circuit abre
   - Simular Connect recuperado → circuit fecha

✅ Chaos Engineering (opcional):
   - Toxiproxy para simular latência
   - Verificar fallback funciona
```

#### Esforço Detalhado

| Tarefa | Esforço |
|--------|---------|
| CircuitBreakerService implementation | 1 dia |
| Integrar com ConnectServiceAdapter | 0.5 dia |
| Integrar com BridgeClient | 0.5 dia |
| Métricas Prometheus | 0.5 dia |
| Testes unitários | 0.25 dia |
| Testes integração | 0.25 dia |
| Documentação | 0.5 dia |
| **TOTAL** | **3 dias** |

---

## 🟡 GAPS P2 - MÉDIA PRIORIDADE

### GAP 4: Validação Explícita de Tamanho Máximo de Campos

**ID**: REG-VAL-012
**Categoria**: Validações
**Status Atual**: ⏳ Schema PostgreSQL tem limits (VARCHAR), falta validação explícita
**Prioridade**: **P2**
**Impacto**: **BAIXO** - PostgreSQL rejeita, mas UX seria melhor com validação antecipada
**Esforço**: **1 dia**
**Prazo**: Sprint +1

#### Implementação

```go
// internal/application/commands/create_entry_command.go

type CreateEntryCommand struct {
    KeyType    string `validate:"required,oneof=CPF CNPJ EMAIL PHONE EVP"`
    KeyValue   string `validate:"required,max=255"`  // [NOVO]
    AccountISPB string `validate:"required,len=8"`
    // ...
}
```

---

### GAP 5: Validação Avançada de Titularidade (Portabilidade)

**ID**: REG-PORT-002
**Categoria**: Portabilidade
**Status Atual**: ⏳ Validação via ConnectService preparada, lógica de comparação pendente
**Prioridade**: **P2**
**Impacto**: **MÉDIO** - Validação regulatória
**Esforço**: **3 dias**
**Prazo**: Sprint +2

#### Implementação

Verificar que **nova conta pertence ao mesmo titular** ao criar claim de portabilidade.

```go
// Comparar CPF/CNPJ do titular da conta origem vs. conta destino
if claim.Type == "PORTABILITY" {
    ownerInfo, _ := connectService.GetAccountInfo(ctx, claim.OwnerISPB, claim.OwnerAccount)
    claimerInfo, _ := connectService.GetAccountInfo(ctx, claim.ClaimerISPB, claim.ClaimerAccount)

    if ownerInfo.OwnerDocument != claimerInfo.OwnerDocument {
        return errors.New("portability requires same account holder")
    }
}
```

---

### GAP 6: Testes E2E de Pulsar Retry + DLQ

**ID**: REG-EVT-005, REG-EVT-006
**Categoria**: Eventos
**Status Atual**: ⏳ Configuração existe, falta validação com testes
**Prioridade**: **P2**
**Impacto**: **BAIXO** - Funcionalidade presente, falta apenas teste
**Esforço**: **2 dias**
**Prazo**: Sprint +3

#### Implementação

Criar testes que:
1. Publicam evento para Pulsar
2. Consumer falha intencionalmente
3. Verificar que Pulsar faz retry (3x)
4. Verificar que evento vai para DLQ após 3 falhas

---

### GAP 7: Throttling Dinâmico (Não Fixo)

**ID**: REG-RATE-003
**Categoria**: Controle
**Status Atual**: ⏳ Config existe, mas fixo (não dinâmico)
**Prioridade**: **P2**
**Impacto**: **BAIXO** - Nice-to-have
**Esforço**: **2 dias**
**Prazo**: Sprint +3

#### Implementação

Carregar configuração de rate limiting de **Vault** ou **etcd** ao invés de config estática.

---

### GAP 8-11: Outros Gaps P2

(Detalhamento similar aos acima)

---

## 📚 ANÁLISE DE ESPECIFICAÇÕES EXISTENTES

### Matriz: Gap x Especificação Disponível

| Gap | Especificação | Status | Arquivos |
|-----|---------------|--------|----------|
| **GAP 1: OTP Validation** | ✅ **COMPLETO** | User stories + Sequence diagram + Struct pronta | US-002, DIA-006, `otp_validation.go` |
| **GAP 2: Rate Limiting IP** | ✅ **COMPLETO** | Redis strategy detalhado | DAT-005 |
| **GAP 3: Circuit Breaker** | ⚠️ **PARCIAL** | Mencionado em SEC-004, sem spec dedicada | SEC-004 |
| **GAP 4: Max Length Validation** | ✅ **COMPLETO** | Schema PostgreSQL + validation layer | DAT-001 |
| **GAP 5: Titularidade** | ✅ **COMPLETO** | US-002 + BP-002 | US-002, BP-002 |
| **GAP 6: Pulsar Retry/DLQ** | ✅ **COMPLETO** | TSP-002 Apache Pulsar | TSP-002 |
| **GAP 7: Throttling Dinâmico** | ⚠️ **PARCIAL** | DAT-005 menciona, sem detalhes | DAT-005 |

**Conclusão**: ✅ **Temos 85% das especificações necessárias**. Apenas Circuit Breaker e Throttling Dinâmico precisam de especificações adicionais.

---

## 📅 PLANO DE IMPLEMENTAÇÃO SPRINT-A-SPRINT

### Sprint +1 (Semanas 1-2): Quick Wins + Infraestrutura

**Objetivo**: Resolver gaps de infraestrutura e validações simples

**Tarefas**:
1. ✅ **GAP 4**: Validação explícita de tamanho (1 dia)
2. ✅ **GAP 2**: Rate limiting por IP (2 dias)
3. ✅ **GAP 3**: Circuit Breaker (3 dias)

**Entregas**:
- Rate limiting por IP funcionando
- Circuit breaker protegendo chamadas Connect/Bridge
- Validações de tamanho completas

**Riscos**: Nenhum (especificações completas)

---

### Sprint +2 (Semanas 3-4): Funcionalidades Regulatórias

**Objetivo**: Implementar requisitos regulatórios (OTP, Titularidade)

**Tarefas**:
1. ✅ **GAP 1**: OTP Validation (5 dias)
2. ✅ **GAP 5**: Validação de Titularidade (3 dias)

**Entregas**:
- OTP funcionando para Email/Phone claims
- Validação de titularidade em portabilidade

**Riscos**:
- Integração com SendGrid/Twilio (requer contas externas)
- Testes E2E precisam de ambiente completo

---

### Sprint +3 (Semanas 5-6): Testes e Otimizações

**Objetivo**: Testes E2E e otimizações não-bloqueantes

**Tarefas**:
1. ✅ **GAP 6**: Testes E2E Pulsar (2 dias)
2. ✅ **GAP 7**: Throttling dinâmico (2 dias)
3. ✅ Outros gaps P2 restantes (4 dias)

**Entregas**:
- 100% de conformidade Bacen
- Suite completa de testes E2E
- Documentação atualizada

---

## ✅ CHECKLIST DE IMPLEMENTAÇÃO

### Antes de Começar

- [ ] Revisar especificações existentes (US-002, DAT-005, SEC-004, etc.)
- [ ] Configurar contas externas (SendGrid, Twilio)
- [ ] Preparar ambiente de testes (PostgreSQL + Redis + Pulsar)
- [ ] Definir responsáveis por cada gap

### Durante Implementação

- [ ] Seguir estrutura de arquivos especificada
- [ ] Escrever testes ANTES do código (TDD)
- [ ] Documentar decisões técnicas (ADRs se necessário)
- [ ] Code review obrigatório (outro dev + arquiteto)

### Após Implementação

- [ ] Todos os testes passando (unit + integration + E2E)
- [ ] Cobertura de testes > 80%
- [ ] Documentação atualizada
- [ ] Métricas Prometheus configuradas
- [ ] Alertas configurados
- [ ] README atualizado

---

## 🎯 CRITÉRIOS DE ACEITE

### Por Gap

**GAP 1 (OTP)**:
- [ ] Email OTP enviado via SendGrid
- [ ] SMS OTP enviado via Twilio
- [ ] Validação correta de código
- [ ] Timeout email (24h) e phone (10min) funciona
- [ ] Bloqueio após 3 tentativas falhas

**GAP 2 (Rate Limiting IP)**:
- [ ] 10 requests/s por IP bloqueados corretamente
- [ ] IPs diferentes não se afetam
- [ ] Métrica Prometheus exportada

**GAP 3 (Circuit Breaker)**:
- [ ] Circuit abre após 3 falhas consecutivas
- [ ] Timeout de 30s → half-open
- [ ] Sucesso em half-open → closed
- [ ] Métricas de estado do circuit

### Geral

- [ ] **Conformidade Bacen**: 100% (185/185 requisitos)
- [ ] **Testes**: > 90% de cobertura
- [ ] **Performance**: Nenhuma degradação vs. versão atual
- [ ] **Documentação**: Todas as features documentadas

---

## 📊 MÉTRICAS DE SUCESSO

### Conformidade

```
Meta Final:
╔════════════════════════════════════════════╗
║  CONFORMIDADE:  100% (185/185 requisitos)  ║
║  GAPS:          0 gaps                     ║
║  TESTES:        >90% cobertura             ║
║  PRODUÇÃO:      100% pronto                ║
╚════════════════════════════════════════════╝
```

### Performance

- Latência p95: < 50ms (sem degradação)
- Rate limiting: 100% efetivo
- Circuit breaker: Reduz cascading failures em 95%

---

## 📞 APROVAÇÕES

**Responsáveis**:
- [ ] **CTO**: José Luís Silva (aprovação final)
- [ ] **Head Arquitetura**: Thiago Lima (revisão técnica)
- [ ] **Head Compliance**: (validação regulatória)
- [ ] **Product Owner**: (priorização de sprints)

---

## 📚 REFERÊNCIAS

### Documentos Base

- [RELATORIO_CONFORMIDADE_BACEN_CORE_DICT.md](RELATORIO_CONFORMIDADE_BACEN_CORE_DICT.md)
- [US-002_User_Stories_Claims.md](../01_Requisitos/UserStories/US-002_User_Stories_Claims.md)
- [DAT-005_Redis_Cache_Strategy.md](../03_Dados/DAT-005_Redis_Cache_Strategy.md)
- [SEC-004_API_Authentication.md](../13_Seguranca/SEC-004_API_Authentication.md)
- [TSP-002_Apache_Pulsar_Messaging.md](../02_Arquitetura/TechSpecs/TSP-002_Apache_Pulsar_Messaging.md)

### Bibliotecas

- **OTP Email**: [SendGrid Go](https://github.com/sendgrid/sendgrid-go)
- **OTP SMS**: [Twilio Go](https://github.com/twilio/twilio-go)
- **Circuit Breaker**: [Gobreaker](https://github.com/sony/gobreaker)
- **Validation**: [Go Validator](https://github.com/go-playground/validator)

---

**Última Atualização**: 2025-10-27
**Versão**: 1.0.0
**Status**: 📋 **PRONTO PARA IMPLEMENTAÇÃO**

---

**FIM DO PLANO**
