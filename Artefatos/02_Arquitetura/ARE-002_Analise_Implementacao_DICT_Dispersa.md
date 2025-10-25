# ARE-002: Análise da Implementação DICT Dispersa no Money-Moving

**ID**: ARE-002
**Data de Criação**: 2025-10-24
**Criado por**: NEXUS (AGT-SA-001)
**Status**: Concluído
**Relacionado a**: ARE-001 (Análise de Repositórios Existentes)

---

## 1. Resumo Executivo

Esta análise documenta a **implementação atual da lógica de negócio DICT** que está **incorretamente dispersa** no repositório `money-moving`, especificamente no app `payment`. O objetivo é identificar todas as funcionalidades DICT implementadas, entender como são executadas as chamadas ao Connect DICT, e mapear o que precisa ser consolidado no novo repositório **Core DICT**.

### Problema Identificado

**Dispersão Arquitetural**: A lógica de negócio DICT está espalhada em múltiplos repositórios:
- **money-moving/apps/payment**: Contém CRUD de chaves PIX + validações + serviço gRPC
- **connector-dict**: Contém API REST + transformação de dados
- **rsfn-connect-bacen-bridge**: Contém integração com Bacen

Esta dispersão viola os princípios de Clean Architecture e Single Responsibility Principle.

### Objetivo do Projeto DICT LBPay

Consolidar **toda a lógica de negócio DICT** em um único repositório **Core DICT**, mantendo apenas:
- **Payment App**: Interface gRPC client que chama Core DICT
- **Core DICT**: Toda regra de negócio, validações, orquestração (novo repo)
- **Connect DICT**: Transformação e roteamento
- **Bridge DICT**: Integração com Bacen

---

## 2. Arquitetura Atual (AS-IS)

### 2.1 Componentes Identificados

```
┌─────────────────────────────────────────────────────────────┐
│                     money-moving/apps/payment               │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │           gRPC Service (dict.go)                     │  │
│  │  - DictSvc (pb.PixKeyServiceServer)                  │  │
│  │  - CreateKey, DeleteKey, ConsultKey, ListKeys       │  │
│  └──────────────┬───────────────────────────────────────┘  │
│                 │                                           │
│  ┌──────────────▼───────────────────────────────────────┐  │
│  │         Business Logic (pixKey.go)                   │  │
│  │  - PixKey Service (IPixKey interface)               │  │
│  │  - Create, Delete, GetByKey, ExistsKey              │  │
│  │  - Validations: formato, limites, status            │  │
│  └──────────────┬───────────────────────────────────────┘  │
│                 │                                           │
│  ┌──────────────▼───────────────────────────────────────┐  │
│  │      HTTP Client (workflow-dict/client.go)          │  │
│  │  - DictClient (implementa IPixKeyClient)            │  │
│  │  - CreatePixKey, DeletePixKey, CheckKeyExists       │  │
│  │  - ConsultPixKey                                    │  │
│  └──────────────┬───────────────────────────────────────┘  │
│                 │                                           │
└─────────────────┼───────────────────────────────────────────┘
                  │ HTTP/REST
                  │
        ┌─────────▼─────────┐
        │  Connector DICT   │
        │   (Port 8082)     │
        │  REST API         │
        └─────────┬─────────┘
                  │ Apache Pulsar
                  │
        ┌─────────▼─────────┐
        │    Bridge DICT    │
        │   mTLS + XML      │
        │   Signing         │
        └─────────┬─────────┘
                  │ HTTPS
                  │
        ┌─────────▼─────────┐
        │   DICT Bacen      │
        └───────────────────┘
```

### 2.2 Fluxo de Chamadas (Exemplo: CreateKey)

1. **Cliente** → gRPC `DictSvc.CreateKey()` em [apps/payment/services/pix-key/dict.go:36](apps/payment/services/pix-key/dict.go#L36)
2. **DictSvc** → Valida formato da chave via `ValidateKeyFormatValue()`
3. **DictSvc** → Monta `CreateEntryRequest` com dados da chave
4. **DictSvc** → Valida request via `dict.ValidateCreateEntryRequest()`
5. **DictSvc** → Valida conta na Receita Federal via `FederalRevenueService`
6. **DictSvc** → Chama `PixKeyClient.CreatePixKey()` (HTTP client)
7. **DictClient** → POST HTTP para `connector-dict/entries` em [apps/payment/shared/workflow-dict/client.go:111](apps/payment/shared/workflow-dict/client.go#L111)
8. **Connector DICT** → Processa e envia para Bridge via Pulsar
9. **Bridge DICT** → Assina XML e envia para Bacen via mTLS
10. **DictSvc** → Verifica criação via `CheckKeyExists()`
11. **DictSvc** → Persiste no banco local via `PixKeyService.Create()`
12. **Response** → Retorna `CreateKeyResponse` ao cliente

---

## 3. Análise Detalhada dos Componentes

### 3.1 gRPC Service Layer ([dict.go](apps/payment/services/pix-key/dict.go))

**Localização**: `apps/payment/services/pix-key/dict.go`

**Responsabilidades**:
- Implementa interface gRPC `pb.PixKeyServiceServer`
- Orquestra fluxo de criação/exclusão/consulta de chaves
- Validações de negócio
- Integração com Receita Federal
- Persistência local

**Métodos Implementados**:

#### CreateKey ([dict.go:36](apps/payment/services/pix-key/dict.go#L36))
```go
func (d *DictSvc) CreateKey(ctx context.Context, req *pb.CreateKeyRequest)
    (*pb.CreateKeyResponse, error)
```

**Fluxo**:
1. Valida formato da chave PIX (CPF, CNPJ, Email, Phone, EVP)
2. Monta `CreateEntryRequest` com Account + Owner + Reason
3. Valida request completo
4. Valida regularidade da conta na Receita Federal
5. Cria chave no DICT via HTTP client
6. Verifica existência da chave criada (CheckKeyExists)
7. Persiste no banco de dados local
8. Registra no histórico

**Validações**:
- Formato da chave (tipo vs valor)
- Campos obrigatórios
- Regularidade fiscal (CPF/CNPJ na Receita Federal)
- Limites e constraints

#### DeleteKey ([dict.go:73](apps/payment/services/pix-key/dict.go#L73))
```go
func (d *DictSvc) DeleteKey(ctx context.Context, req *pb.DeleteKeyRequest)
    (*pb.DeleteKeyResponse, error)
```

**Fluxo**:
1. Valida campos obrigatórios (PixKey, ISPB)
2. Monta `DeleteEntryRequest` com Participant + Reason
3. Valida request
4. Verifica se chave existe no banco local
5. Deleta no DICT via HTTP client
6. Remove do banco de dados local
7. Registra exclusão no histórico

#### ConsultKey ([dict.go:105](apps/payment/services/pix-key/dict.go#L105))
```go
func (d *DictSvc) ConsultKey(ctx context.Context, req *pb.ConsultKeyRequest)
    (*pb.ConsultKeyResponse, error)
```

**Fluxo**:
1. Monta `ConsultEntryRequest` com PixKey + ISPB + PayerID + EndToEndID
2. Valida request
3. Garante que chave não existe no banco local (para consultas externas)
4. Consulta no DICT via HTTP client
5. Retorna dados da chave (Account + Owner)

**Observação Importante**: O método `EnsureKeyDoesNotExist` verifica se a chave já existe no banco local. Isso é usado para **consultas de chaves externas** (não pertencentes ao PSP).

#### ListKeys ([dict.go:138](apps/payment/services/pix-key/dict.go#L138))
```go
func (d *DictSvc) ListKeys(ctx context.Context, req *pb.ListKeysRequest)
    (*pb.ListKeysResponse, error)
```

**Fluxo**:
1. Busca chaves por AccountID no banco local
2. Retorna lista de chaves (tipo + valor)

**Observação**: Este método **não** consulta o DICT, apenas o banco local.

---

### 3.2 Business Logic Layer ([pixKey.go](apps/payment/services/pix-key/pixKey.go))

**Localização**: `apps/payment/services/pix-key/pixKey.go`

**Responsabilidades**:
- CRUD de chaves PIX no banco de dados local
- Validações de status e limites
- Gerenciamento de histórico

**Interface IPixKey**:
```go
type IPixKey interface {
    Create(ctx, *pixKeyModel.PixKeyRequestModel) (*pixKeyModel.PixKeyModel, error)
    DeleteByKey(ctx, *pixKeyModel.PixKeyRequestModel) error
    GetByKey(ctx, pixKeyValue string) (*pixKeyModel.PixKeyModel, error)
    GetStatusByKey(ctx, pixKeyValue string) (*string, error)
    CountByAccountId(ctx, accountID, operationID string) (*int64, error)
    ExistsKey(ctx, pixKeyValue string) error
    EnsureKeyDoesNotExist(ctx, pixKeyValue string) error
    GetPixKeysByAccountId(ctx, accountID string) ([]pixKeyModel.PixKeyModel, error)
}
```

**Regras de Negócio Implementadas**:

1. **Limite de Chaves**: Constante `LIMIT_PIX_KEY = 77` caracteres
2. **Status da Chave**: `PixKeyStatusActive = "ATIVADO"`
3. **Validações**:
   - Chave não pode estar vazia
   - Chave não pode exceder 77 caracteres
   - Chave deve estar ativa para operações
   - Account ID é obrigatório

4. **Histórico**: Todas as operações são registradas em tabela de histórico:
   - `historyStore.SaveCreation()` - Criação
   - `historyStore.SaveExclusion()` - Exclusão

**Métodos Críticos**:

#### ExistsKey ([pixKey.go:112](apps/payment/services/pix-key/pixKey.go#L112))
Valida se chave existe **e está ativa** no banco local.

```go
func (p *PixKey) ExistsKey(ctx context.Context, pixKeyValue string) error {
    keyState, err := p.GetStatusByKey(ctx, pixKeyValue)
    if err != nil {
        return status.Error(codes.NotFound, "pix key not found")
    }
    if *keyState != pixKeyModel.PixKeyStatusActive {
        return status.Error(codes.FailedPrecondition, "pix key is not active")
    }
    return nil
}
```

#### EnsureKeyDoesNotExist ([pixKey.go:151](apps/payment/services/pix-key/pixKey.go#L151))
Garante que chave **não existe** no banco local (usado para consultas externas).

```go
func (p *PixKey) EnsureKeyDoesNotExist(ctx context.Context, pixKeyValue string) error {
    _, err := p.store.GetByKey(ctx, pixKeyValue)
    if err != nil {
        if errors.Is(err, pixKeyStore.ErrPixKeyNotFound) {
            return nil // OK - chave não existe
        }
        return status.Error(codes.Internal, "error checking pix key existence")
    }
    // Chave foi encontrada - erro!
    return status.Error(codes.AlreadyExists,
        fmt.Sprintf("key found in the internal PSP database: %s", pixKeyValue))
}
```

---

### 3.3 HTTP Client Layer ([workflow-dict/client.go](apps/payment/shared/workflow-dict/client.go))

**Localização**: `apps/payment/shared/workflow-dict/client.go`

**Responsabilidades**:
- Implementa interface `IPixKeyClient`
- Comunicação HTTP/REST com Connector DICT
- Tratamento de erros HTTP
- Timeout e retry (implícito via http.Client)

**Estrutura DictClient**:
```go
type DictClient struct {
    baseURL    string
    httpClient *http.Client
}

func NewDictClient(baseURL string) *DictClient {
    return &DictClient{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}
```

**Métodos Implementados**:

#### CreatePixKey ([client.go:111](apps/payment/shared/workflow-dict/client.go#L111))
```go
func (c *DictClient) CreatePixKey(ctx context.Context, req *dict.CreateEntryRequest)
    (*dict.CreateEntryResponse, error)
```

**Request**:
- **Method**: `POST`
- **Endpoint**: `/entries`
- **Body**: JSON com `CreateEntryRequest`
- **Headers**: `Content-Type: application/json`

**Response**:
- **Success**: 200/201 → `CreateEntryResponse` com Entry criada
- **Error**: 4xx/5xx → Parse de `ErrorResponse` ou mensagem genérica

#### CheckKeyExists ([client.go:132](apps/payment/shared/workflow-dict/client.go#L132))
```go
func (c *DictClient) CheckKeyExists(ctx context.Context, req *dict.CheckKeysRequest)
    (*dict.CheckKeysResponse, error)
```

**Request**:
- **Method**: `POST`
- **Endpoint**: `/keys/check`
- **Body**: JSON com array de chaves (max 100)

**Response**:
- Array de `KeyCheck` com `has_entry` boolean

#### DeletePixKey ([client.go:153](apps/payment/shared/workflow-dict/client.go#L153))
```go
func (c *DictClient) DeletePixKey(ctx context.Context, key string,
    req *dict.DeleteEntryRequest) error
```

**Request**:
- **Method**: `POST` (não é DELETE!)
- **Endpoint**: `/entries/{key}/delete`
- **Body**: JSON com `DeleteEntryRequest` (Participant + Reason)

**Response**:
- **Success**: 200/204 → sem body
- **Error**: 4xx/5xx → Parse de `ErrorResponse`

#### ConsultPixKey ([client.go:168](apps/payment/shared/workflow-dict/client.go#L168))
```go
func (c *DictClient) ConsultPixKey(ctx context.Context, req *dict.ConsultEntryRequest)
    (*dict.ConsultEntryResponse, error)
```

**Request**:
- **Method**: `GET`
- **Endpoint**: `/entries/{pix_key}`
- **Headers**:
  - `PI-RequestingParticipant`: ISPB do solicitante
  - `PI-PayerId`: ID do pagador
  - `PI-EndToEndId`: ID end-to-end da transação

**Response**:
- **Success**: 200 → `ConsultEntryResponse` com Entry completa
- **Error**: 404 → "Entrada associada à chave fornecida não existe"

**Tratamento de Erros**:
```go
var errorMessages = map[int]string{
    http.StatusNotFound:           "Entrada associada à chave fornecida não existe",
    http.StatusBadRequest:         "Requisição inválida",
    http.StatusForbidden:          "Participante não está autorizado a acessar este recurso",
    http.StatusServiceUnavailable: "Serviço indisponível no momento",
}
```

**Observações**:
1. **Timeout**: Fixo em 30 segundos
2. **Retry**: Não implementado (deveria estar no Core DICT)
3. **Circuit Breaker**: Não implementado
4. **Idempotência**: Depende do Connector DICT

---

### 3.4 Data Models ([model/dict/dict.go](apps/payment/model/dict/dict.go))

**Localização**: `apps/payment/model/dict/dict.go`

**Estruturas de Dados**:

#### CreateEntryRequest
```go
type CreateEntryRequest struct {
    RequestID string  `json:"request_id" validate:"required"`
    Key       string  `json:"key" validate:"required,min=10,max=77"`
    KeyType   string  `json:"key_type" validate:"required,oneof=CPF CNPJ PHONE EMAIL EVP"`
    Account   Account `json:"account" validate:"required"`
    Owner     Owner   `json:"owner" validate:"required"`
    Reason    string  `json:"reason" validate:"required,oneof=USER_REQUESTED BRANCH_TRANSFER RECONCILIATION RFB_VALIDATION"`
}
```

**Validações**:
- `RequestID`: Obrigatório (UUID)
- `Key`: 10-77 caracteres
- `KeyType`: Enum (CPF, CNPJ, PHONE, EMAIL, EVP)
- `Reason`: Enum (USER_REQUESTED, BRANCH_TRANSFER, RECONCILIATION, RFB_VALIDATION)

#### Account
```go
type Account struct {
    Participant   string    `json:"participant" validate:"required,len=8"`      // ISPB
    Branch        string    `json:"branch" validate:"required,len=4"`
    AccountNumber string    `json:"account_number" validate:"required,min=1,max=10"`
    AccountType   string    `json:"account_type" validate:"required,oneof=CACC SAVINGS"`
    OpeningDate   time.Time `json:"opening_date" validate:"required"`
}
```

**Validações**:
- `Participant`: Exatamente 8 dígitos (ISPB)
- `Branch`: Exatamente 4 dígitos
- `AccountNumber`: 1-10 caracteres
- `AccountType`: Enum (CACC, SAVINGS)

#### Owner
```go
type Owner struct {
    Type        string `json:"type" validate:"required,oneof=NATURAL_PERSON LEGAL_PERSON"`
    TaxIDNumber string `json:"tax_id_number" validate:"required,min=11,max=14"`
    Name        string `json:"name" validate:"required,min=3,max=100"`
    TradeName   string `json:"trade_name,omitempty" validate:"omitempty,min=3,max=100"`
}
```

**Validações**:
- `Type`: Enum (NATURAL_PERSON, LEGAL_PERSON)
- `TaxIDNumber`: 11-14 caracteres (CPF=11, CNPJ=14)
- `Name`: 3-100 caracteres
- `TradeName`: Opcional, 3-100 caracteres se presente

#### DeleteEntryRequest
```go
type DeleteEntryRequest struct {
    Participant string `json:"participant" validate:"required,len=8"`
    Reason      string `json:"reason" validate:"required,oneof=USER_REQUESTED ACCOUNT_CLOSURE RECONCILIATION FRAUD RFB_VALIDATION"`
}
```

#### ConsultEntryRequest
```go
type ConsultEntryRequest struct {
    PixKey     string `json:"pix_key" validate:"required"`
    ISPB       string `json:"ispb" validate:"required"`
    PayerID    string `json:"payer_id" validate:"required"`
    EndToEndID string `json:"end_to_end_id" validate:"required"`
}
```

**Observação**: Todos os campos são obrigatórios para consultas (requisito Bacen).

#### CheckKeysRequest
```go
type CheckKeysRequest struct {
    Keys []string `json:"keys" validate:"required,min=1,max=100"`
}
```

**Validação**: Máximo 100 chaves por request (batch check).

#### ErrorResponse (RFC 9457)
```go
type ErrorResponse struct {
    Schema string        `json:"$schema,omitempty"`
    Title  string        `json:"title"`
    Status int           `json:"status"`
    Detail string        `json:"detail"`
    Errors []ErrorDetail `json:"errors,omitempty"`
}
```

---

### 3.5 Validation Service ([validate_dict_key_service.go](apps/payment/internal/service/validate_dict_key_service.go))

**Localização**: `apps/payment/internal/service/validate_dict_key_service.go`

**Responsabilidade**: Validar chaves PIX para transações PIX-In/PIX-Out.

**Estrutura**:
```go
type ValidateDictKey struct {
    PixKeyService services.IPixKey
}
```

**Método Principal**:
```go
func (s *ValidateDictKey) ValidatePixKey(ctx context.Context,
    req *pb.ValidatePixKeyRequest) (*pb.ValidatePixKeyResponse, error)
```

**Fluxo**:
1. Valida formato da chave via `shared.IsPixKey()`
2. Busca chave no banco local via `PixKeyService.GetByKey()`
3. Retorna status: `VALID` ou `INVALID`
4. Se válida, retorna também `AccountId`

**Casos de Uso**:
- Validação de chave PIX em transações PIX-In
- Validação de chave PIX em transações PIX-Out
- Verificação rápida sem consulta ao DICT

**Observação Crítica**: Este serviço valida apenas chaves **locais** (do próprio PSP). Não consulta o DICT para chaves externas.

---

## 4. Funcionalidades DICT Implementadas

### 4.1 Bloco 1: Gerenciamento de Chaves (CRUD)

| Funcionalidade | Status | Implementado Em | Observações |
|----------------|--------|-----------------|-------------|
| **Criação de Chave** | ✅ Implementado | `DictSvc.CreateKey()` | Inclui validação Receita Federal |
| **Exclusão de Chave** | ✅ Implementado | `DictSvc.DeleteKey()` | Inclui registro de histórico |
| **Consulta de Chave** | ✅ Implementado | `DictSvc.ConsultKey()` | Com headers PI-* |
| **Listagem de Chaves** | ✅ Implementado | `DictSvc.ListKeys()` | Apenas banco local |
| **Verificação em Lote** | ✅ Implementado | `CheckKeyExists()` | Max 100 chaves |
| **Validação Rápida** | ✅ Implementado | `ValidateDictKey` | Para transações |

### 4.2 Blocos 2-6: Funcionalidades Ausentes

| Bloco | Funcionalidade | Status |
|-------|----------------|--------|
| **Bloco 2** | Reivindicação (Claim) | ❌ Não implementado |
| **Bloco 3** | Portabilidade | ❌ Não implementado |
| **Bloco 4** | Devolução | ❌ Não implementado |
| **Bloco 5** | Segurança Avançada | ⚠️ Parcial (mTLS no Bridge) |
| **Bloco 6** | Recuperação de Dados | ❌ Não implementado |

---

## 5. Integrações Identificadas

### 5.1 Integração com Receita Federal

**Localização**: `apps/payment/services/federal-revenue/`

**Uso**: Validação de regularidade fiscal (CPF/CNPJ) antes de criar chave PIX.

**Chamada**: `FederalRevenueService.ValidateAccountRegularity()`

**Observação**: Esta validação é um **requisito Bacen** para criação de chaves.

### 5.2 Integração com Banco de Dados Local

**Localização**: `apps/payment/store/pixKeyStore/`

**Tabelas**:
- `pix_keys`: Armazena chaves ativas
- `pix_keys_history`: Auditoria de operações

**Stores**:
- `Store`: CRUD de chaves
- `HistoryStore`: Registro de histórico

**Métodos**:
```go
type Store interface {
    Save(ctx, *pixKeyModel.PixKeyModel) (*pixKeyModel.PixKeyModel, error)
    DeleteByKey(ctx, key string) (*string, *string, error)
    GetByKey(ctx, key string) (*pixKeyModel.PixKeyModel, error)
    GetStatusByKey(ctx, key string) (string, error)
    CountByAccountId(ctx, accountID, operationID string) (int64, error)
    GetPixKeysByAccountId(ctx, accountID string) ([]pixKeyModel.PixKeyModel, error)
}
```

**Observação**: O Core DICT precisa decidir se mantém este banco local ou usa apenas o DICT Bacen como source of truth.

### 5.3 Integração com Connector DICT

**Protocolo**: HTTP/REST
**Base URL**: Configurável via env var
**Timeout**: 30 segundos
**Endpoints**:

| Método | Endpoint | Operação |
|--------|----------|----------|
| POST | `/entries` | Criar chave |
| POST | `/keys/check` | Verificar chaves (batch) |
| POST | `/entries/{key}/delete` | Deletar chave |
| GET | `/entries/{key}` | Consultar chave |

**Observação**: O verbo DELETE não é usado (POST para `/delete`).

### 5.4 Integração com lb-contracts

**Localização**: `github.com/london-bridge/lb-contracts/go/money-moving/payment/dict`

**Uso**: Contratos gRPC/Protobuf para comunicação com clientes.

**Services**:
- `PixKeyService`: Serviço gRPC principal
- `ValidatePixKeyService`: Serviço de validação rápida

**Observação**: Estes contratos precisam ser evoluídos para o novo Core DICT.

---

## 6. Gaps e Limitações Identificadas

### 6.1 Arquiteturais

1. **Violação de Clean Architecture**:
   - Lógica de negócio DICT dispersa em múltiplos apps/repos
   - Payment app não deveria ter regras de negócio DICT
   - Falta camada de domínio isolada

2. **Falta de Resiliência**:
   - ❌ Sem circuit breaker
   - ❌ Sem retry policy
   - ❌ Sem fallback strategies
   - ⚠️ Timeout fixo (30s) sem configuração

3. **Falta de Observabilidade**:
   - ⚠️ Logging básico (logrus)
   - ❌ Sem distributed tracing
   - ❌ Sem métricas de negócio
   - ❌ Sem correlationID entre serviços

4. **Falta de Testes**:
   - Arquivos `*_test.go` existem mas não foram analisados
   - Necessário avaliar cobertura e qualidade

### 6.2 Funcionais

1. **Blocos DICT Ausentes**:
   - ❌ Reivindicação (Claim) - Bloco 2
   - ❌ Portabilidade - Bloco 3
   - ❌ Devolução - Bloco 4
   - ❌ Recuperação - Bloco 6

2. **Validações Incompletas**:
   - Falta validação de tipos de chave vs tipo de conta
   - Falta validação de limites de chaves por conta
   - Falta validação de blacklist/whitelist

3. **Idempotência**:
   - Não está claro como é garantida idempotência
   - Falta correlation ID entre camadas
   - Falta deduplicação de requests

4. **Eventos e Auditoria**:
   - ⚠️ Histórico em banco local (não é event sourcing)
   - ❌ Sem eventos de domínio
   - ❌ Sem integração com sistema de auditoria central

### 6.3 Técnicos

1. **Configuração**:
   - Hardcoded timeout (30s)
   - Base URL do connector via configuração (não documentado)
   - Falta service discovery

2. **Segurança**:
   - Não há autenticação entre payment → connector
   - Falta autorização baseada em roles
   - Dados sensíveis (CPF/CNPJ) sem masking em logs

3. **Performance**:
   - Validação Receita Federal síncrona (pode ser lenta)
   - CheckKeyExists é chamado após CreatePixKey (dobra latência)
   - Sem cache de chaves consultadas

---

## 7. Modelo de Dados Atual

### 7.1 Entidades Identificadas

#### PixKeyModel (Banco Local)
```go
type PixKeyModel struct {
    ID            string
    AccountID     string
    TypeKey       string // CPF, CNPJ, PHONE, EMAIL, EVP
    ValueKey      string // Valor da chave
    Status        string // ATIVADO, INATIVO, etc.
    OperationID   string
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

#### Entry (DICT Bacen)
```go
type Entry struct {
    Key     string
    KeyType string
    Account Account
    Owner   Owner
}
```

**Diferenças**:
- Modelo local tem `AccountID` (interno LBPay)
- Modelo DICT tem `Account` completo (ISPB + Branch + Number)
- Modelo local tem `Status` e `OperationID`
- Falta mapeamento explícito entre os dois

### 7.2 Relacionamentos

```
Account (LBPay) 1 ──── N PixKey
              │
              │ (foreign key: account_id)
              │
PixKey 1 ──── N PixKeyHistory
```

**Observação**: Não há relação explícita com entidades de transação (PIX-In/PIX-Out).

---

## 8. Padrões de Código Identificados

### 8.1 Clean Architecture (Parcial)

**Camadas Identificadas**:
1. **Handlers (gRPC)**: `dict.go` - DictSvc
2. **Application Services**: `pixKey.go` - PixKey service
3. **Infrastructure**: `client.go` - DictClient, `store/` - Database
4. **Domain Models**: `model/dict/`, `model/pixKeyModel/`

**Problemas**:
- Falta camada de **Domain** isolada (regras de negócio)
- Falta **Use Cases** explícitos
- Infraestrutura (HTTP client) usada diretamente em Application

### 8.2 Dependency Injection

**Padrão Usado**: Constructor injection

```go
func NewDictSvc(db *sql.DB, dict IPixKeyClient) *DictSvc {
    return &DictSvc{
        PixKeyService:         NewPixKeyService(db),
        FederalRevenueService: federalRevenue.NewFederalRevenue(),
        PixKeyClient:          dict,
    }
}
```

**Interfaces**:
- `IPixKey`: Business logic de chaves
- `IPixKeyClient`: Cliente HTTP para connector
- `IFederalRevenue`: Validação Receita Federal

**Observação**: Padrão correto, facilita testes e desacoplamento.

### 8.3 Error Handling

**Padrão**: gRPC status codes

```go
if err != nil {
    return nil, status.Error(codes.InvalidArgument, err.Error())
}
```

**Códigos Usados**:
- `codes.InvalidArgument`: Validação de campos
- `codes.NotFound`: Entidade não encontrada
- `codes.Internal`: Erros internos
- `codes.AlreadyExists`: Chave já existe
- `codes.FailedPrecondition`: Estado inválido

**Observação**: Falta padronização de mensagens de erro (i18n).

### 8.4 Validações

**Biblioteca**: `github.com/go-playground/validator/v10`

**Estratégia**:
1. Validações de formato em structs (tags `validate`)
2. Validações de negócio em métodos específicos
3. Funções helper (`ValidatePixKeyRequest`, etc.)

**Exemplo**:
```go
err := validate.Struct(req)
if err != nil {
    // Parse validator.ValidationErrors
    // Retornar erro específico por campo
}
```

---

## 9. Proposta de Migração para Core DICT

### 9.1 O que deve permanecer em Payment App

```go
// payment/internal/clients/dict_client.go
type DictClient interface {
    CreateKey(ctx, *pb.CreateKeyRequest) (*pb.CreateKeyResponse, error)
    DeleteKey(ctx, *pb.DeleteKeyRequest) (*pb.DeleteKeyResponse, error)
    ConsultKey(ctx, *pb.ConsultKeyRequest) (*pb.ConsultKeyResponse, error)
    ValidateKey(ctx, keyValue string) (bool, error)
}

// gRPC client para Core DICT - APENAS chamada remota
```

**Responsabilidade**: Interface thin client que chama Core DICT via gRPC.

### 9.2 O que deve migrar para Core DICT (novo repo)

**Tudo de Business Logic**:
- ✅ Validações de formato e regras de negócio
- ✅ Integração com Receita Federal
- ✅ Orquestração de fluxo (criar → verificar → persistir)
- ✅ Gerenciamento de status
- ✅ Histórico e auditoria
- ✅ Cliente HTTP para Connector DICT
- ✅ Modelos de dados DICT
- ✅ Blocos 2-6 (novos)

**Estrutura Proposta** (Clean Architecture):
```
core-dict/
├── cmd/
│   └── server/          # gRPC server entrypoint
├── internal/
│   ├── domain/          # Entidades, Value Objects, Domain Services
│   │   ├── pixkey/
│   │   ├── claim/
│   │   └── portability/
│   ├── application/     # Use Cases, DTOs
│   │   ├── usecases/
│   │   └── ports/       # Interfaces para infra
│   ├── infrastructure/  # Implementações concretas
│   │   ├── grpc/        # gRPC handlers
│   │   ├── http/        # HTTP client para Connector
│   │   ├── database/    # Repositories
│   │   └── external/    # Receita Federal, etc.
│   └── shared/          # Utils, constants
└── api/
    └── proto/           # gRPC contracts
```

### 9.3 O que deve permanecer em Connector DICT

- ✅ API REST endpoints
- ✅ Transformação de dados (REST → Pulsar)
- ✅ Roteamento de mensagens
- ✅ Validações básicas de schema

**Observação**: Connector não tem regras de negócio.

### 9.4 O que deve permanecer em Bridge DICT

- ✅ Assinatura XML
- ✅ mTLS com Bacen
- ✅ Consumo Pulsar
- ✅ Envio HTTP para Bacen

---

## 10. Dependências Externas Identificadas

### 10.1 Bibliotecas Golang

| Biblioteca | Versão | Uso |
|------------|--------|-----|
| `google.golang.org/grpc` | - | gRPC server/client |
| `github.com/go-playground/validator/v10` | - | Validações |
| `github.com/sirupsen/logrus` | - | Logging |
| `database/sql` | stdlib | Database access |
| `github.com/google/uuid` | - | Request IDs |

### 10.2 Contratos (lb-contracts)

| Contrato | Localização |
|----------|-------------|
| `PixKeyService` | `go/money-moving/payment/dict` |
| `ValidatePixKeyService` | `go/money-moving/payment/validatekeypix` |

**Observação**: Estes contratos são específicos do payment app e não devem ser reusados no Core DICT.

### 10.3 Serviços Externos

| Serviço | Integração | Protocolo |
|---------|------------|-----------|
| **Connector DICT** | HTTP client | REST/JSON |
| **Receita Federal** | FederalRevenueService | ? (não detalhado) |
| **Database** | sql.DB | SQL (Postgres?) |

---

## 11. Métricas e SLAs (Ausentes)

### 11.1 Performance

**Não documentado**:
- Latência máxima esperada
- Throughput mínimo
- Percentis (p50, p95, p99)

**Timeout Atual**: 30s fixo no HTTP client

### 11.2 Disponibilidade

**Não documentado**:
- SLA de disponibilidade
- RPO/RTO
- Estratégia de failover

### 11.3 Observabilidade

**Implementado**:
- ⚠️ Logging com logrus (não estruturado)

**Ausente**:
- ❌ Distributed tracing
- ❌ Métricas Prometheus
- ❌ Dashboards
- ❌ Alertas

---

## 12. Riscos da Migração

### 12.1 Riscos Técnicos

| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| **Perda de funcionalidade** | Média | Alto | Testes E2E + Checklist de funcionalidades |
| **Quebra de contratos gRPC** | Baixa | Alto | Versionamento de APIs + Backward compatibility |
| **Regressão de validações** | Média | Alto | Suite de testes de validação + Property-based testing |
| **Perda de dados históricos** | Baixa | Médio | Migração incremental + Backup |

### 12.2 Riscos de Negócio

| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| **Downtime em produção** | Baixa | Crítico | Deploy incremental + Feature flags |
| **Inconsistência com Bacen** | Baixa | Crítico | Validação extensiva em sandbox |
| **Impacto em transações PIX** | Média | Alto | Canary deployment + Rollback plan |

---

## 13. Checklist de Funcionalidades para Migração

### 13.1 Bloco 1: CRUD de Chaves

- [ ] **CreateKey**
  - [ ] Validação de formato
  - [ ] Validação Receita Federal
  - [ ] Chamada para Connector DICT
  - [ ] Verificação de criação (CheckKeyExists)
  - [ ] Persistência local
  - [ ] Registro de histórico
  - [ ] Testes unitários
  - [ ] Testes de integração

- [ ] **DeleteKey**
  - [ ] Validação de ownership (ISPB)
  - [ ] Verificação de existência
  - [ ] Chamada para Connector DICT
  - [ ] Remoção local
  - [ ] Registro de histórico
  - [ ] Testes

- [ ] **ConsultKey**
  - [ ] Headers PI-* obrigatórios
  - [ ] Verificação de não-ownership
  - [ ] Chamada para Connector DICT
  - [ ] Tratamento de erros (404, 403, etc.)
  - [ ] Testes

- [ ] **ListKeys**
  - [ ] Busca por AccountID
  - [ ] Paginação (se necessário)
  - [ ] Testes

- [ ] **CheckKeyExists** (batch)
  - [ ] Limite de 100 chaves
  - [ ] Chamada para Connector DICT
  - [ ] Testes

- [ ] **ValidateKey** (transações)
  - [ ] Validação de formato
  - [ ] Busca local
  - [ ] Retorno rápido (< 100ms)
  - [ ] Testes

### 13.2 Infraestrutura

- [ ] **HTTP Client**
  - [ ] Configuração de timeout
  - [ ] Retry policy
  - [ ] Circuit breaker
  - [ ] Connection pooling
  - [ ] Testes de resiliência

- [ ] **Database**
  - [ ] Schema de chaves
  - [ ] Schema de histórico
  - [ ] Indexes
  - [ ] Migrations
  - [ ] Testes

- [ ] **gRPC Server**
  - [ ] Implementação de services
  - [ ] Interceptors (auth, logging, tracing)
  - [ ] Health checks
  - [ ] Graceful shutdown
  - [ ] Testes

---

## 14. Próximos Passos

### 14.1 Análises Adicionais Necessárias

1. **Analisar testes existentes**:
   - `dict_test.go`
   - `client_test.go`
   - Avaliar cobertura e qualidade

2. **Analisar integração com Receita Federal**:
   - `apps/payment/services/federal-revenue/`
   - Protocolo e contratos
   - Performance e SLA

3. **Analisar database schema**:
   - Tabelas `pix_keys` e `pix_keys_history`
   - Indexes e performance
   - Estratégia de backup

4. **Analisar logs e observabilidade**:
   - Padrões de logging
   - Correlation IDs
   - Integração com sistemas de monitoramento

5. **Analisar configurações**:
   - Variáveis de ambiente
   - Feature flags
   - Secrets management

### 14.2 Documentos a Criar

1. **ADR-003**: Consolidação de Core DICT em Repositório Único
2. **ADR-004**: Estratégia de Migração Incremental
3. **DAS-002**: Arquitetura TO-BE do Core DICT
4. **MDC-001**: Modelo de Dados Unificado DICT
5. **ETS-001**: Especificação Técnica do Core DICT gRPC Service
6. **PLM-001**: Plano de Migração e Rollout

### 14.3 Artefatos de Implementação

1. **RFC-001**: Proposta de Arquitetura Core DICT
2. **POC-001**: Prova de Conceito - Core DICT MVP
3. **TST-001**: Plano de Testes de Migração
4. **DOC-001**: Guia de Migração para Desenvolvedores

---

## 15. Conclusões

### 15.1 Principais Achados

1. **Lógica DICT está dispersa** no app payment, violando princípios de Clean Architecture e SRP.

2. **Funcionalidades básicas (CRUD) estão implementadas**, mas blocos 2-6 estão ausentes.

3. **Padrões corretos são usados** (DI, interfaces, validações), facilitando a extração.

4. **Faltam aspectos críticos** de resiliência, observabilidade e testes.

5. **HTTP client é simples demais** para produção (sem retry, circuit breaker, etc.).

### 15.2 Recomendações Prioritárias

1. **Criar Core DICT como novo repositório** seguindo Clean Architecture rigorosa.

2. **Implementar resiliência** (retry, circuit breaker, timeout configurável).

3. **Adicionar observabilidade completa** (tracing, métricas, structured logging).

4. **Implementar blocos 2-6** do Manual DICT Bacen.

5. **Criar suite de testes abrangente** (unit, integration, E2E, contract).

6. **Migração incremental** com feature flags e canary deployment.

### 15.3 Validação com Stakeholders

- [ ] **CTO**: Aprovação da estratégia de migração
- [ ] **Arquiteto de Soluções**: Revisão da arquitetura TO-BE
- [ ] **Tech Leads**: Validação de esforço e riscos
- [ ] **QA**: Definição de estratégia de testes
- [ ] **DevOps**: Plano de deployment e rollback

---

## 16. Referências

### 16.1 Código Analisado

- [money-moving/apps/payment/services/pix-key/dict.go](https://github.com/london-bridge/money-moving/blob/main/apps/payment/services/pix-key/dict.go)
- [money-moving/apps/payment/services/pix-key/pixKey.go](https://github.com/london-bridge/money-moving/blob/main/apps/payment/services/pix-key/pixKey.go)
- [money-moving/apps/payment/services/pix-key/pixKeyClient.go](https://github.com/london-bridge/money-moving/blob/main/apps/payment/services/pix-key/pixKeyClient.go)
- [money-moving/apps/payment/shared/workflow-dict/client.go](https://github.com/london-bridge/money-moving/blob/main/apps/payment/shared/workflow-dict/client.go)
- [money-moving/apps/payment/model/dict/dict.go](https://github.com/london-bridge/money-moving/blob/main/apps/payment/model/dict/dict.go)
- [money-moving/apps/payment/internal/service/validate_dict_key_service.go](https://github.com/london-bridge/money-moving/blob/main/apps/payment/internal/service/validate_dict_key_service.go)

### 16.2 Documentos Relacionados

- [ARE-001: Análise de Repositórios Existentes](./ARE-001_Analise_Repositorios_Existentes.md)
- [DUVIDAS.md](../00_Master/DUVIDAS.md) - DUV-010, DUV-011
- [PMP-001: Plano Master do Projeto](../11_Gestao/PMP-001_Plano_Master_Projeto.md)

### 16.3 Repositórios Relacionados

- [connector-dict](https://github.com/lb-conn/connector-dict)
- [rsfn-connect-bacen-bridge](https://github.com/lb-conn/rsfn-connect-bacen-bridge)
- [money-moving](https://github.com/london-bridge/money-moving)
- [lb-contracts](https://github.com/london-bridge/lb-contracts)

---

## Histórico de Revisões

| Data | Versão | Autor | Mudanças |
|------|--------|-------|----------|
| 2025-10-24 | 1.0 | NEXUS (AGT-SA-001) | Criação inicial do documento |

---

**Documento produzido por**: NEXUS (Solution Architect - AGT-SA-001)
**Revisado por**: (pendente)
**Aprovado por**: (pendente)
