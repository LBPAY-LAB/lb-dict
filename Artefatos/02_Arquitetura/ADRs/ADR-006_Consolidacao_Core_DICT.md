# ADR-002 - Consolidação de Lógica DICT em Repositório Único

**Status**: ✅ Aprovado
**Data**: 2025-10-24
**Decisores**: José Luís Silva (CTO), NEXUS (Solution Architect)
**Contexto**: Fase de Especificação do Projeto DICT LBPay

---

## Contexto

### Situação Atual (AS-IS)

A lógica de negócio do DICT está atualmente **dispersa em múltiplos repositórios**:

1. **money-moving** (repositório principal):
   - Lógica de criação de chaves PIX
   - Consultas a chaves
   - Parte da validação de contas
   - **Problema**: Mistura lógica DICT com lógica de movimentação financeira

2. **orchestration-go**:
   - Orquestração de fluxos PIX
   - Algumas validações de chaves
   - **Problema**: Duplicação de lógica com money-moving

3. **operation**:
   - Operações administrativas
   - Algumas consultas DICT
   - **Problema**: Acoplamento desnecessário

4. **connector-dict** (parcialmente implementado):
   - Cliente REST para DICT Bacen
   - Apenas 3-4 endpoints implementados (de 28 totais)
   - **Problema**: Implementação incompleta e inconsistente

### Problemas Identificados

#### 1. Dispersão e Duplicação
```
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│  money-moving   │      │ orchestration-go│      │   operation     │
│                 │      │                 │      │                 │
│ • CreatePixKey  │      │ • CreatePixKey  │      │ • GetPixKey     │
│ • GetPixKey     │      │   (duplicado!)  │      │   (duplicado!)  │
│ • ValidateKey   │      │ • ValidateKey   │      │                 │
│                 │      │   (duplicado!)  │      │                 │
└─────────────────┘      └─────────────────┘      └─────────────────┘
         ↓                        ↓                        ↓
    ❌ Duplicação de código (3 implementações da mesma lógica)
    ❌ Inconsistências (cada repo valida de forma diferente)
    ❌ Difícil manutenção (bug fix precisa ser replicado em 3 lugares)
```

#### 2. Baixa Cobertura de Requisitos

**72 Requisitos Funcionais** mapeados no CRF-001:
- ✅ **6 RFs implementados** (8.4%)
- ❌ **66 RFs NÃO implementados** (91.6%)

**Gap crítico**:
- Bloco 2 (Reivindicação/Portabilidade): **0% implementado**
- Bloco 4 (Devolução/Infração): **0% implementado**
- Bloco 5 (Segurança): **23.1% implementado**

#### 3. Ausência de Arquitetura Clara

Não há separação clara entre:
- **Lógica de negócio** (regras DICT)
- **Infraestrutura** (comunicação com Bacen)
- **Aplicação** (use cases)

Exemplo atual em `money-moving`:
```go
// BAD: Lógica de negócio misturada com chamadas HTTP
func CreatePixKey(key string) error {
    // Validação inline (deveria estar no Domain)
    if len(key) < 5 {
        return errors.New("invalid key")
    }

    // Chamada HTTP inline (deveria estar na Infrastructure)
    resp, err := http.Post("https://dict.bacen...", body)
    if err != nil {
        return err
    }

    // Persistência inline (deveria estar na Infrastructure)
    db.Exec("INSERT INTO pixkeys...")

    return nil
}
```

#### 4. Difícil Evolução

Para implementar novos RFs:
- ❌ Precisa modificar 3 repositórios diferentes
- ❌ Risco de quebrar funcionalidades existentes
- ❌ Testes distribuídos e incompletos
- ❌ Deploy de 3 serviços coordenado

---

## Decisão

**Criar um único repositório `core-dict`** que concentra **TODA** a lógica de negócio DICT, seguindo **Clean Architecture**.

### Estrutura do Novo Repositório

```
core-dict/
├── cmd/
│   └── server/
│       └── main.go                 # Entrypoint
├── pkg/
│   ├── domain/                     # CAMADA DE DOMÍNIO
│   │   ├── aggregates/
│   │   │   ├── pixkey.go           # Aggregate Root
│   │   │   ├── claim.go
│   │   │   ├── refund.go
│   │   │   └── infraction.go
│   │   ├── entities/
│   │   │   ├── account.go
│   │   │   └── owner.go
│   │   ├── valueobjects/
│   │   │   ├── key.go              # CPF, CNPJ, EMAIL, PHONE, EVP
│   │   │   ├── cid.go              # Content Identifier
│   │   │   └── vsync.go            # VSync (XOR de CIDs)
│   │   ├── validators/
│   │   │   ├── key_validator.go
│   │   │   ├── limit_validator.go  # 5/20 keys per account
│   │   │   └── owner_validator.go
│   │   ├── services/
│   │   │   ├── cid_calculator.go   # CID = HMAC-SHA256(...)
│   │   │   └── vsync_calculator.go # VSync = XOR(cid1, cid2, ...)
│   │   └── repositories/
│   │       ├── pixkey_repository.go    # Interface
│   │       └── claim_repository.go     # Interface
│   │
│   ├── application/                # CAMADA DE APLICAÇÃO
│   │   └── usecases/
│   │       ├── pixkey/
│   │       │   ├── create_pixkey.go    # RF-BLO1-001
│   │       │   ├── get_pixkey.go       # RF-BLO1-008
│   │       │   ├── update_pixkey.go    # RF-BLO1-009
│   │       │   ├── delete_pixkey.go    # RF-BLO1-003
│   │       │   └── validate_pixkey.go  # RF-BLO1-011
│   │       ├── claim/
│   │       │   ├── create_claim.go     # RF-BLO2-001
│   │       │   ├── confirm_claim.go    # RF-BLO2-010
│   │       │   └── complete_claim.go   # RF-BLO2-003
│   │       ├── refund/
│   │       │   ├── create_refund.go    # RF-BLO4-001
│   │       │   └── close_refund.go     # RF-BLO4-001
│   │       └── validation/
│   │           ├── validate_possession.go  # RF-BLO3-001
│   │           └── validate_cpf_cnpj.go    # RF-BLO3-002
│   │
│   ├── handlers/                   # CAMADA DE HANDLERS
│   │   ├── rest/
│   │   │   ├── router.go
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go
│   │   │   │   ├── logging.go
│   │   │   │   └── tracing.go
│   │   │   └── v1/
│   │   │       ├── pixkey_handler.go
│   │   │       ├── claim_handler.go
│   │   │       └── refund_handler.go
│   │   ├── grpc/
│   │   │   ├── server.go
│   │   │   └── services/
│   │   │       └── pixkey_service.go
│   │   └── pulsar/
│   │       └── consumers/
│   │           ├── cid_events.go
│   │           └── spi_events.go
│   │
│   └── infrastructure/             # CAMADA DE INFRAESTRUTURA
│       ├── persistence/
│       │   ├── postgres/
│       │   │   ├── pixkey_repo_impl.go    # Implementação
│       │   │   └── claim_repo_impl.go
│       │   └── migrations/
│       ├── cache/
│       │   ├── response_cache.go       # Redis 7001
│       │   ├── account_cache.go        # Redis 7002
│       │   └── ratelimit_cache.go      # Redis 7005
│       ├── clients/
│       │   ├── connect_dict_client.go  # gRPC → Connect DICT
│       │   ├── validator_client.go     # HTTP → Receita Federal
│       │   └── temporal_client.go
│       └── messaging/
│           └── pulsar_publisher.go
│
├── workflows/                      # TEMPORAL WORKFLOWS
│   ├── claim_workflow.go           # 7-day claim process
│   ├── validation_workflow.go      # SMS/Email validation
│   └── reconciliation_workflow.go  # VSync periodic check
│
├── configs/
│   ├── config.yaml
│   └── config.go
│
├── scripts/
│   ├── migrations/
│   └── seeds/
│
├── tests/
│   ├── unit/
│   ├── integration/
│   └── e2e/
│
├── docs/
│   ├── architecture.md
│   └── api/
│
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md
```

### Princípios da Clean Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        HANDLERS                              │
│  (REST API, gRPC, Pulsar Consumers)                         │
│  • Recebe requests                                           │
│  • Valida inputs                                             │
│  • Chama Use Cases                                           │
│  • Retorna responses                                         │
└──────────────────────┬──────────────────────────────────────┘
                       │ Depende ↓
┌──────────────────────┴──────────────────────────────────────┐
│                     APPLICATION                              │
│  (Use Cases)                                                 │
│  • Orquestra fluxo de negócio                                │
│  • Coordena Domain + Infrastructure                          │
│  • Sem lógica de negócio (delega ao Domain)                  │
└──────────────────────┬──────────────────────────────────────┘
                       │ Depende ↓
┌──────────────────────┴──────────────────────────────────────┐
│                       DOMAIN                                 │
│  (Business Logic - 72 RFs)                                   │
│  • Aggregates (PixKey, Claim, Refund)                        │
│  • Entities (Account, Owner)                                 │
│  • Value Objects (Key, CID, VSync)                           │
│  • Validators (regras de negócio)                            │
│  • Domain Services (CID calc, VSync calc)                    │
│  • Repository Interfaces (NÃO implementação!)                │
└──────────────────────▲──────────────────────────────────────┘
                       │ É usado por ↑
┌──────────────────────┴──────────────────────────────────────┐
│                  INFRASTRUCTURE                              │
│  (Detalhes técnicos)                                         │
│  • Repository Implementations (PostgreSQL)                   │
│  • Cache Implementations (Redis)                             │
│  • External Clients (Connect DICT, Validator, Temporal)      │
│  • Messaging (Pulsar)                                        │
└─────────────────────────────────────────────────────────────┘

REGRA DE OURO: Dependências apontam SEMPRE para dentro (Domain não depende de nada)
```

---

## Consequências

### ✅ Positivas

#### 1. Código Limpo e Manutenível
```go
// GOOD: Separação clara de responsabilidades

// Domain Layer (Business Logic)
type PixKey struct {
    Key     Key
    Account Account
    Owner   Owner
    CID     CID
}

func (pk *PixKey) Validate() error {
    // Business rules ONLY
    if pk.Account.KeyCount() >= pk.Owner.MaxKeys() {
        return ErrLimitExceeded
    }
    return nil
}

// Application Layer (Use Case)
type CreatePixKeyUseCase struct {
    repo       domain.PixKeyRepository   // Interface
    dictClient infrastructure.DICTClient // Interface
    cache      infrastructure.Cache      // Interface
}

func (uc *CreatePixKeyUseCase) Execute(ctx context.Context, input CreatePixKeyInput) error {
    // 0. Validate ownership (Manual Bacen Subseção 2.1) - PRÉ-REQUISITO
    // ⚠️ ATENÇÃO: Esta validação DEVE ter sido feita ANTES de chamar este Use Case
    // Para chaves PHONE/EMAIL: código SMS/e-mail validado (timeout: 30 min)
    // Para chaves CPF/CNPJ: titularidade da conta já valida posse
    // Para chaves EVP: gerada pelo DICT, não requer validação prévia
    if input.KeyType == domain.PHONE || input.KeyType == domain.EMAIL {
        if !uc.ownershipValidator.IsValidated(ctx, input.Key, input.ValidationToken) {
            return domain.ErrOwnershipNotValidated // Usuário deve receber novo código
        }
    }

    // 1. Validate business rules (Domain)
    pixKey := domain.NewPixKey(input.Key, input.Account, input.Owner)
    if err := pixKey.Validate(); err != nil {
        return err
    }

    // 2. Call DICT Bacen (Infrastructure)
    if err := uc.dictClient.CreateEntry(ctx, pixKey); err != nil {
        return err
    }

    // 3. Persist (Infrastructure)
    if err := uc.repo.Save(ctx, pixKey); err != nil {
        return err
    }

    // 4. Update cache (Infrastructure)
    uc.cache.Set(ctx, pixKey.Key, pixKey)

    // 5. Clear ownership validation token (security)
    if input.KeyType == domain.PHONE || input.KeyType == domain.EMAIL {
        uc.ownershipValidator.ClearToken(ctx, input.Key)
    }

    return nil
}
```

#### 2. Testabilidade Máxima

```go
// Teste unitário (Domain - sem dependências externas)
func TestPixKey_Validate_LimitExceeded(t *testing.T) {
    owner := domain.Owner{Type: PF, MaxKeys: 5}
    account := domain.Account{KeyCount: 5} // Já tem 5 chaves

    pixKey := domain.NewPixKey("+5561988880000", account, owner)

    err := pixKey.Validate()

    assert.Error(t, err)
    assert.Equal(t, domain.ErrLimitExceeded, err)
}

// Teste de integração (Use Case com mocks)
func TestCreatePixKeyUseCase_Success(t *testing.T) {
    // Arrange
    mockRepo := new(MockPixKeyRepository)
    mockDICTClient := new(MockDICTClient)
    mockCache := new(MockCache)

    mockDICTClient.On("CreateEntry", mock.Anything, mock.Anything).Return(nil)
    mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

    uc := NewCreatePixKeyUseCase(mockRepo, mockDICTClient, mockCache)

    // Act
    err := uc.Execute(ctx, CreatePixKeyInput{...})

    // Assert
    assert.NoError(t, err)
    mockDICTClient.AssertExpectations(t)
    mockRepo.AssertExpectations(t)
}
```

**Cobertura de testes esperada**: 80%+ (vs 20-30% atual)

#### 3. Única Fonte de Verdade

```
ANTES (AS-IS):
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ money-moving │  │orchestration │  │  operation   │
│              │  │              │  │              │
│ CreatePixKey │  │ CreatePixKey │  │              │
│ (versão 1)   │  │ (versão 2)   │  │              │
└──────────────┘  └──────────────┘  └──────────────┘
   ❌ Qual é a implementação correta???

DEPOIS (TO-BE):
                 ┌──────────────┐
                 │  core-dict   │
                 │              │
                 │ CreatePixKey │
                 │ (ÚNICA VERSÃO)
                 └──────────────┘
                   ✅ Single Source of Truth
```

#### 4. Evolução Facilitada

Para implementar **novo RF** (ex: RF-BLO2-007 - Portabilidade):

**ANTES** (AS-IS):
1. ❌ Modificar `money-moving/pkg/dict/portability.go`
2. ❌ Modificar `orchestration-go/workflows/claim.go`
3. ❌ Adicionar endpoint em `connector-dict/...` (se existir)
4. ❌ Atualizar 3 conjuntos de testes
5. ❌ Deploy coordenado de 3 serviços
6. ⏱️ **Tempo estimado**: 5-7 dias

**DEPOIS** (TO-BE):
1. ✅ Criar `domain/aggregates/claim.go` (business logic)
2. ✅ Criar `application/usecases/claim/create_portability.go`
3. ✅ Adicionar handler em `handlers/rest/v1/claim_handler.go`
4. ✅ Testar (unit + integration)
5. ✅ Deploy de 1 serviço
6. ⏱️ **Tempo estimado**: 2-3 dias

**Ganho**: 50-60% redução de tempo de desenvolvimento.

#### 5. Independência de Frameworks

Domain não depende de:
- ❌ Gin/Chi (REST framework)
- ❌ gRPC
- ❌ GORM (ORM)
- ❌ PostgreSQL
- ❌ Redis

**Benefício**: Podemos trocar qualquer framework/lib sem mexer no Domain.

Exemplo: Migrar de PostgreSQL para CockroachDB:
```go
// ANTES: Código acoplado a PostgreSQL
func CreatePixKey(key string) error {
    db, _ := sql.Open("postgres", "...")
    db.Exec("INSERT INTO pixkeys...")  // ❌ Acoplado!
}

// DEPOIS: Código desacoplado via Repository Interface
// Domain Layer (não muda!)
type PixKeyRepository interface {
    Save(ctx context.Context, key *PixKey) error
}

// Infrastructure Layer (implementação PostgreSQL)
type PostgresPixKeyRepository struct { ... }

// Para migrar para CockroachDB:
// 1. Criar nova implementação: CockroachDBPixKeyRepository
// 2. Trocar no DI container
// 3. Domain NÃO precisa mudar!
```

#### 6. Alinhamento com Regulamentação Bacen

Os **72 RFs** do CRF-001 mapeiam diretamente para:
- **Aggregates**: PixKey, Claim, Refund, InfractionReport
- **Use Cases**: 1 use case por RF

Exemplo:
- **RF-BLO1-001**: `usecases/pixkey/create_pixkey.go`
- **RF-BLO2-007**: `usecases/claim/create_portability.go`
- **RF-BLO4-001**: `usecases/refund/create_refund.go`

**Rastreabilidade completa**: Requisito → Use Case → Testes

#### 7. Serviço de Validação de Posse (Ownership Validator)

**Contexto**: Manual Operacional DICT Bacen - **Subseção 2.1** (Validação da posse da chave)

**Responsabilidade**: Validar posse de chaves tipo PHONE (celular) e EMAIL antes do registro no DICT.

**Interface** (Domain Layer):
```go
// Domain Service Interface
type OwnershipValidator interface {
    // Gerar e enviar código de validação
    SendValidationCode(ctx context.Context, keyType KeyType, key string, recipient string) (token string, err error)

    // Verificar se código foi validado (usuário inseriu corretamente)
    IsValidated(ctx context.Context, key string, token string) bool

    // Limpar token após uso (segurança)
    ClearToken(ctx context.Context, key string) error
}

// Value Object para Token de Validação
type ValidationToken struct {
    Key         string
    Token       string    // Código de 6 dígitos (ex: 123456)
    ExpiresAt   time.Time // Timeout: 30 minutos (configurável)
    ValidatedAt *time.Time
}
```

**Implementação** (Infrastructure Layer):
```go
type RedisOwnershipValidator struct {
    cache        Cache              // Redis para armazenar tokens (TTL 30 min)
    smsGateway   SMSGateway         // Gateway para envio de SMS
    emailGateway EmailGateway       // Gateway para envio de e-mail
    tokenTTL     time.Duration      // Configurável (default: 30 minutos)
}

func (v *RedisOwnershipValidator) SendValidationCode(ctx context.Context, keyType KeyType, key string, recipient string) (string, error) {
    // 1. Gerar código aleatório de 6 dígitos
    code := generateRandomCode(6) // Ex: "123456"

    // 2. Gerar token único (hash)
    token := sha256(key + code + salt)

    // 3. Armazenar no Redis com TTL de 30 min
    validationToken := ValidationToken{
        Key:       key,
        Token:     token,
        ExpiresAt: time.Now().Add(v.tokenTTL), // 30 min
    }
    v.cache.Set(ctx, "ownership:"+key, validationToken, v.tokenTTL)

    // 4. Enviar código via SMS ou E-mail
    if keyType == PHONE {
        return token, v.smsGateway.Send(ctx, recipient, fmt.Sprintf("Seu código PIX: %s", code))
    } else {
        return token, v.emailGateway.Send(ctx, recipient, "Código PIX", fmt.Sprintf("Seu código: %s", code))
    }
}

func (v *RedisOwnershipValidator) IsValidated(ctx context.Context, key string, token string) bool {
    // 1. Buscar token no Redis
    var vt ValidationToken
    exists, _ := v.cache.Get(ctx, "ownership:"+key, &vt)
    if !exists {
        return false // Token expirou (30 min) ou não existe
    }

    // 2. Verificar se token confere
    if vt.Token != token {
        return false // Token inválido
    }

    // 3. Verificar se foi validado (usuário inseriu código corretamente)
    return vt.ValidatedAt != nil
}
```

**Fluxo Completo** (Subseção 2.1 do Manual Bacen):
```
1. Portal/App → CreatePixKeyUseCase.Execute()
   ↓
2. Use Case verifica se keyType == PHONE ou EMAIL
   ↓
3. Se sim → ownershipValidator.SendValidationCode()
   • Gera código de 6 dígitos
   • Envia SMS/e-mail
   • Armazena token no Redis (TTL 30 min)
   • Retorna token para frontend
   ↓
4. Frontend mostra tela "Insira o código recebido"
   ↓
5. Usuário insere código → Chama endpoint /validate-ownership
   • Valida código
   • Marca token.ValidatedAt = now()
   ↓
6. Frontend chama novamente CreatePixKeyUseCase.Execute() com token validado
   ↓
7. Use Case verifica ownershipValidator.IsValidated() = true
   ↓
8. Prossegue com registro no DICT Bacen
   ↓
9. Limpa token (ownershipValidator.ClearToken())
```

**Configuração** (timeout ajustável):
```yaml
# config/ownership_validation.yaml
ownership_validation:
  token_ttl: 30m              # Timeout padrão: 30 minutos (Manual Bacen)
  code_length: 6              # Código de 6 dígitos
  max_retry_attempts: 3       # Máximo 3 tentativas de reenvio
  retry_cooldown: 1m          # 1 minuto entre reenvios
```

**Casos de Exceção**:
- **Chaves CPF/CNPJ**: Posse validada pela titularidade da conta → `OwnershipValidator` não é chamado
- **Chaves EVP (aleatória)**: Gerada pelo DICT → `OwnershipValidator` não é chamado
- **Timeout expirado** (30 min): Redis remove token automaticamente → Usuário deve solicitar novo código

---

### ⚠️ Negativas (e Mitigações)

#### 1. Migração de Código Existente

**Problema**: Precisa migrar código dos 3 repos atuais para `core-dict`.

**Estimativa de Esforço**:
- Análise de código: 2 semanas
- Refatoração: 4 semanas
- Testes: 2 semanas
- **Total**: 8 semanas (160 horas)

**Mitigação**:
- ✅ **Estratégia incremental**: Migrar bloco por bloco (começar por Bloco 1)
- ✅ **Manter compatibilidade**: Repos antigos continuam funcionando durante migração
- ✅ **Feature flags**: Alternar entre implementação antiga/nova gradualmente

**Roadmap de Migração**:
```
Semana 1-2:  Criar estrutura core-dict (esqueleto Clean Architecture)
Semana 3-4:  Migrar Bloco 1 - CRUD (13 RFs)
Semana 5-6:  Migrar validações + reconciliação
Semana 7-8:  Testes E2E + homologação
Semana 9+:   Deprecar código antigo gradualmente
```

#### 2. Curva de Aprendizado (Clean Architecture)

**Problema**: Time pode não estar familiarizado com Clean Architecture.

**Mitigação**:
- ✅ **Treinamento**: Workshop de 2 dias sobre Clean Architecture (Robert C. Martin)
- ✅ **Documentação**: README detalhado com exemplos práticos
- ✅ **Pair Programming**: Desenvolvedores seniores mentoram juniores
- ✅ **Code Review**: Revisão rigorosa para garantir aderência aos princípios

**Materiais de Treinamento**:
- [Clean Architecture - Uncle Bob](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Clean Architecture Example](https://github.com/bxcodec/go-clean-arch)
- Documentação interna: `docs/architecture.md`

#### 3. Overhead Inicial de Código

**Problema**: Clean Architecture cria mais arquivos/camadas.

**Exemplo**:
```
ANTES (AS-IS): 1 arquivo, 100 linhas
money-moving/pkg/dict/pixkey.go

DEPOIS (TO-BE): 6 arquivos, 300 linhas
core-dict/pkg/domain/aggregates/pixkey.go         (50 linhas)
core-dict/pkg/domain/validators/key_validator.go  (30 linhas)
core-dict/pkg/application/usecases/pixkey/create_pixkey.go (60 linhas)
core-dict/pkg/handlers/rest/v1/pixkey_handler.go  (50 linhas)
core-dict/pkg/infrastructure/persistence/postgres/pixkey_repo_impl.go (80 linhas)
core-dict/tests/unit/pixkey_test.go               (30 linhas)
```

**Mitigação**:
- ✅ **Gerador de código**: Script para gerar estrutura base de novos RFs
  ```bash
  $ make generate-usecase name=CreateClaim block=2 rf=RF-BLO2-001
  # Cria automaticamente:
  # - domain/aggregates/claim.go
  # - application/usecases/claim/create_claim.go
  # - handlers/rest/v1/claim_handler.go
  # - tests/unit/create_claim_test.go
  ```
- ✅ **Templates**: Templates prontos para cada camada
- ✅ **Benefício a longo prazo**: Overhead inicial compensa com manutenibilidade

---

## Alternativas Consideradas

### Alternativa 1: Manter Status Quo (Múltiplos Repos)

**Prós**:
- ✅ Sem esforço de migração
- ✅ Time já conhece código atual

**Contras**:
- ❌ **91.6% dos RFs não implementados** (66 de 72)
- ❌ Duplicação de código
- ❌ Inconsistências
- ❌ Difícil evolução
- ❌ Baixa testabilidade
- ❌ Alto risco de bugs

**Decisão**: ❌ **Rejeitada** - Inviável para implementar 72 RFs.

---

### Alternativa 2: Refatorar Repos Existentes (sem novo repo)

**Prós**:
- ✅ Sem criação de novo repo
- ✅ Menos migração de código

**Contras**:
- ❌ Ainda mantém dispersão (3 repos)
- ❌ Difícil aplicar Clean Architecture em código legado
- ❌ Acoplamento existente (money-moving mistura DICT + movimentação financeira)

**Decisão**: ❌ **Rejeitada** - Não resolve problema de dispersão.

---

### Alternativa 3: Monorepo (todos os serviços em 1 repo)

**Prós**:
- ✅ Versionamento único
- ✅ Compartilhamento de código facilitado

**Contras**:
- ❌ Não é padrão LBPay
- ❌ CI/CD complexo (deploy seletivo)
- ❌ Alto impacto em toda organização

**Decisão**: ❌ **Rejeitada** - Mudança muito disruptiva.

---

## Decisão Final

✅ **APROVADA**: Criar repositório **`core-dict`** único com **Clean Architecture**.

### Justificativa

1. ✅ **Única forma de implementar 72 RFs** de forma sustentável
2. ✅ **Elimina duplicação** e inconsistências
3. ✅ **Testabilidade máxima** (80%+ cobertura)
4. ✅ **Manutenibilidade** a longo prazo
5. ✅ **Alinhamento com regulamentação** Bacen
6. ✅ **Independência de frameworks**
7. ✅ **Evolução facilitada** (novos RFs em 50-60% menos tempo)

---

## Implementação

### Fase 1: Criação do Repositório (Semanas 1-2)

```bash
# 1. Criar repositório
git init core-dict
cd core-dict

# 2. Criar estrutura base
make scaffold-clean-architecture

# 3. Configurar CI/CD
cp .github/workflows/ci.yml.template .github/workflows/ci.yml

# 4. Configurar Docker
docker-compose up -d postgres redis pulsar

# 5. Primeira migração
make migrate-up
```

### Fase 2: Migração de Código (Semanas 3-8)

**Prioridade**: Bloco 1 (CRUD) - **Must Have**

1. Migrar `CreatePixKey` (money-moving → core-dict)
2. Migrar `GetPixKey` (money-moving → core-dict)
3. Migrar `DeletePixKey` (money-moving → core-dict)
4. Implementar RFs faltantes do Bloco 1 (7 novos RFs)

### Fase 3: Deprecação Gradual (Semanas 9+)

```go
// Em money-moving (código legado)
func CreatePixKey(key string) error {
    if featureflags.IsEnabled("use-core-dict") {
        // Nova implementação (chama core-dict via gRPC)
        return coreDictClient.CreatePixKey(ctx, key)
    } else {
        // Implementação antiga (deprecated)
        return createPixKeyLegacy(key)
    }
}
```

**Timeline de deprecação**:
- Semana 9: Feature flag 10% (canary)
- Semana 10: Feature flag 50%
- Semana 11: Feature flag 100%
- Semana 12: Remover código legado

---

## Referências

1. **Clean Architecture** - Robert C. Martin
   https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html

2. **Domain-Driven Design** - Eric Evans
   https://www.domainlanguage.com/ddd/

3. **Go Clean Architecture Example** - bxcodec
   https://github.com/bxcodec/go-clean-arch

4. **CRF-001** - Checklist de Requisitos Funcionais
   [Artefatos/05_Requisitos/CRF-001_Checklist_Requisitos_Funcionais.md](../05_Requisitos/CRF-001_Checklist_Requisitos_Funcionais.md)

5. **DAS-001** - Arquitetura de Solução TO-BE
   [Artefatos/02_Arquitetura/DAS-001_Arquitetura_Solucao_TO_BE.md](DAS-001_Arquitetura_Solucao_TO_BE.md)

---

**Documento criado por**: NEXUS (AGT-ARC-001) - Solution Architect
**Aprovado por**: José Luís Silva (CTO)
**Data de Aprovação**: 2025-10-24
**Status**: ✅ Aprovado
**Impacto**: 🔴 Alto (mudança estrutural)

---

## Assinaturas

**Solution Architect**: NEXUS (AGT-ARC-001)
**CTO**: José Luís Silva
**Data**: 2025-10-24
