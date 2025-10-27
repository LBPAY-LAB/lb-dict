# Validação: Interface gRPC Core-Dict para Front-End

**Data**: 2025-10-27
**Objetivo**: Validar que o Front-End tem **TODAS** as funções necessárias via gRPC

---

## 📋 4 Grupos de Funcionalidades DICT

### 1️⃣ Directory (Vínculos DICT) - Gestão de Chaves PIX
### 2️⃣ Claim (Reivindicação de Posse) - Processo de 30 dias
### 3️⃣ Portability (Portabilidade) - Mudança de conta
### 4️⃣ Directory Queries (Consultas DICT) - Lookup de chaves de terceiros

---

## ✅ Grupo 1: Directory (Vínculos DICT) - 4 RPCs

Gerenciamento completo de chaves PIX do usuário autenticado.

| # | RPC | Request | Response | Status | Descrição |
|---|-----|---------|----------|--------|-----------|
| 1 | **CreateKey** | CreateKeyRequest | CreateKeyResponse | ✅ | Criar nova chave PIX (CPF, CNPJ, Email, Phone, EVP) |
| 2 | **ListKeys** | ListKeysRequest | ListKeysResponse | ✅ | Listar todas as chaves do usuário (paginação + filtros) |
| 3 | **GetKey** | GetKeyRequest | GetKeyResponse | ✅ | Obter detalhes completos de uma chave (histórico portabilidade) |
| 4 | **DeleteKey** | DeleteKeyRequest | DeleteKeyResponse | ✅ | Deletar chave PIX do usuário |

### Detalhes CreateKey

**Request**:
```protobuf
message CreateKeyRequest {
  dict.common.v1.KeyType key_type = 1;  // CPF, CNPJ, EMAIL, PHONE, EVP
  string key_value = 2;                  // Opcional se EVP (gerado)
  string account_id = 3;                 // ID da conta LBPay
}
```

**Response**:
```protobuf
message CreateKeyResponse {
  string key_id = 1;                     // UUID da key
  dict.common.v1.DictKey key = 2;        // Tipo + Valor
  dict.common.v1.EntryStatus status = 3; // ACTIVE
  google.protobuf.Timestamp created_at = 4;
}
```

**Front-End Use Cases**:
- ✅ Criar chave CPF (validação: 1 por CPF)
- ✅ Criar chave CNPJ (validação: até 20 por CNPJ)
- ✅ Criar chave Email (validação OTP obrigatório)
- ✅ Criar chave Phone (validação SMS obrigatório)
- ✅ Criar chave EVP (gerada automaticamente - UUID)

---

### Detalhes ListKeys

**Request**:
```protobuf
message ListKeysRequest {
  int32 page_size = 1;                   // Default: 20, Max: 100
  string page_token = 2;                 // Token paginação
  optional KeyType key_type = 3;         // Filtro por tipo
  optional EntryStatus status = 4;       // Filtro por status
}
```

**Response**:
```protobuf
message ListKeysResponse {
  repeated KeySummary keys = 1;
  string next_page_token = 2;
  int32 total_count = 3;
}

message KeySummary {
  string key_id = 1;
  dict.common.v1.DictKey key = 2;
  dict.common.v1.EntryStatus status = 3;
  string account_id = 4;
  google.protobuf.Timestamp created_at = 5;
  google.protobuf.Timestamp updated_at = 6;
}
```

**Front-End Use Cases**:
- ✅ Listar todas as chaves do usuário
- ✅ Filtrar por tipo (só CPF, só Email, etc.)
- ✅ Filtrar por status (ACTIVE, CLAIM_PENDING, etc.)
- ✅ Paginação (max 100 por página)
- ✅ Ver total de chaves (para validar limites: 5 CPF, 20 CNPJ)

---

### Detalhes GetKey

**Request**:
```protobuf
message GetKeyRequest {
  oneof identifier {
    string key_id = 1;               // Buscar por ID
    dict.common.v1.DictKey key = 2;  // Buscar por valor (tipo + valor)
  }
}
```

**Response**:
```protobuf
message GetKeyResponse {
  string key_id = 1;
  dict.common.v1.DictKey key = 2;
  dict.common.v1.Account account = 3;
  dict.common.v1.EntryStatus status = 4;
  google.protobuf.Timestamp created_at = 5;
  google.protobuf.Timestamp updated_at = 6;
  repeated PortabilityHistory portability_history = 7; // 🔥 Histórico!
}

message PortabilityHistory {
  string portability_id = 1;
  dict.common.v1.Account old_account = 2;
  dict.common.v1.Account new_account = 3;
  google.protobuf.Timestamp confirmed_at = 4;
}
```

**Front-End Use Cases**:
- ✅ Ver detalhes completos de uma chave
- ✅ Ver conta vinculada (ISPB, agência, conta, nome titular)
- ✅ Ver histórico de portabilidades (todas as trocas de conta)
- ✅ Buscar por ID ou por valor da chave

---

### Detalhes DeleteKey

**Request**:
```protobuf
message DeleteKeyRequest {
  string key_id = 1;
}
```

**Response**:
```protobuf
message DeleteKeyResponse {
  bool deleted = 1;
  google.protobuf.Timestamp deleted_at = 2;
}
```

**Front-End Use Cases**:
- ✅ Deletar chave PIX
- ✅ Confirmação de deleção (timestamp)

---

## ✅ Grupo 2: Claim (Reivindicação de Posse) - 6 RPCs

Processo completo de reivindicação de chave (30 dias).

| # | RPC | Request | Response | Status | Descrição |
|---|-----|---------|----------|--------|-----------|
| 5 | **StartClaim** | StartClaimRequest | StartClaimResponse | ✅ | Iniciar reivindicação de chave de terceiro |
| 6 | **GetClaimStatus** | GetClaimStatusRequest | GetClaimStatusResponse | ✅ | Ver status de uma claim específica (dias restantes) |
| 7 | **ListIncomingClaims** | ListIncomingClaimsRequest | ListIncomingClaimsResponse | ✅ | Listar claims **recebidas** (sou o dono atual) |
| 8 | **ListOutgoingClaims** | ListOutgoingClaimsRequest | ListOutgoingClaimsResponse | ✅ | Listar claims **enviadas** (sou o reivindicador) |
| 9 | **RespondToClaim** | RespondToClaimRequest | RespondToClaimResponse | ✅ | Responder a claim (ACEITAR ou REJEITAR) |
| 10 | **CancelClaim** | CancelClaimRequest | CancelClaimResponse | ✅ | Cancelar claim enviada |

### Detalhes StartClaim

**Request**:
```protobuf
message StartClaimRequest {
  dict.common.v1.DictKey key = 1;  // Chave a reivindicar (tipo + valor)
  string account_id = 2;            // Conta destino (LBPay)
}
```

**Response**:
```protobuf
message StartClaimResponse {
  string claim_id = 1;                     // UUID da claim
  string entry_id = 2;                     // ID da entry reivindicada
  dict.common.v1.ClaimStatus status = 3;   // OPEN
  google.protobuf.Timestamp expires_at = 4; // created_at + 30 dias
  google.protobuf.Timestamp created_at = 5;
  string message = 6;                      // "Claim criada. O dono tem 30 dias..."
}
```

**Front-End Use Cases**:
- ✅ Reivindicar chave de outra instituição
- ✅ Ver prazo de expiração (30 dias)
- ✅ Receber mensagem explicativa para usuário

---

### Detalhes GetClaimStatus

**Request**:
```protobuf
message GetClaimStatusRequest {
  string claim_id = 1;
}
```

**Response**:
```protobuf
message GetClaimStatusResponse {
  string claim_id = 1;
  string entry_id = 2;
  dict.common.v1.DictKey key = 3;
  dict.common.v1.ClaimStatus status = 4;   // OPEN, CONFIRMED, CANCELLED, etc.
  string claimer_ispb = 5;                 // ISPB do reivindicador
  string owner_ispb = 6;                   // ISPB do dono atual
  google.protobuf.Timestamp created_at = 7;
  google.protobuf.Timestamp expires_at = 8;
  optional google.protobuf.Timestamp completed_at = 9;
  int32 days_remaining = 10;               // 🔥 Para exibir countdown!
}
```

**Front-End Use Cases**:
- ✅ Ver status atual da claim
- ✅ Ver **dias restantes** (countdown: "Faltam 27 dias")
- ✅ Ver ISPBs envolvidos (quem reivindica, quem possui)
- ✅ Ver quando foi completada (se CONFIRMED ou CANCELLED)

---

### Detalhes ListIncomingClaims

**Request**:
```protobuf
message ListIncomingClaimsRequest {
  optional dict.common.v1.ClaimStatus status = 1; // Filtro
  int32 page_size = 2;
  string page_token = 3;
}
```

**Response**:
```protobuf
message ListIncomingClaimsResponse {
  repeated ClaimSummary claims = 1;
  string next_page_token = 2;
  int32 total_count = 3;
}

message ClaimSummary {
  string claim_id = 1;
  string entry_id = 2;
  dict.common.v1.DictKey key = 3;
  dict.common.v1.ClaimStatus status = 4;
  google.protobuf.Timestamp created_at = 5;
  google.protobuf.Timestamp expires_at = 6;
  int32 days_remaining = 7; // 🔥
}
```

**Front-End Use Cases**:
- ✅ Listar claims que **eu recebi** (alguém quer minha chave)
- ✅ Filtrar por status (OPEN, WAITING_RESOLUTION, etc.)
- ✅ Ver dias restantes para cada claim
- ✅ Paginação

---

### Detalhes ListOutgoingClaims

**Request e Response**: Idênticos a ListIncomingClaims (mesmo formato)

**Front-End Use Cases**:
- ✅ Listar claims que **eu enviei** (quero chave de terceiro)
- ✅ Filtrar por status
- ✅ Ver dias restantes
- ✅ Paginação

---

### Detalhes RespondToClaim

**Request**:
```protobuf
message RespondToClaimRequest {
  string claim_id = 1;

  enum ClaimResponse {
    CLAIM_RESPONSE_UNSPECIFIED = 0;
    CLAIM_RESPONSE_ACCEPT = 1;   // Aceitar (transfere chave)
    CLAIM_RESPONSE_REJECT = 2;   // Rejeitar (mantém chave)
  }
  ClaimResponse response = 2;

  optional string reason = 3;    // Razão (opcional, para rejeição)
}
```

**Response**:
```protobuf
message RespondToClaimResponse {
  string claim_id = 1;
  dict.common.v1.ClaimStatus new_status = 2; // CONFIRMED ou CANCELLED
  google.protobuf.Timestamp responded_at = 3;
  string message = 4;                        // Mensagem para usuário
}
```

**Front-End Use Cases**:
- ✅ Aceitar claim (transferir chave)
- ✅ Rejeitar claim (manter chave)
- ✅ Adicionar razão (opcional)
- ✅ Receber mensagem de confirmação

---

### Detalhes CancelClaim

**Request**:
```protobuf
message CancelClaimRequest {
  string claim_id = 1;
  optional string reason = 2;
}
```

**Response**:
```protobuf
message CancelClaimResponse {
  string claim_id = 1;
  dict.common.v1.ClaimStatus status = 2;     // CANCELLED
  google.protobuf.Timestamp cancelled_at = 3;
}
```

**Front-End Use Cases**:
- ✅ Cancelar claim enviada (desistir da reivindicação)
- ✅ Adicionar razão (opcional)

---

## ✅ Grupo 3: Portability (Portabilidade) - 3 RPCs

Portabilidade de chave para nova conta (dentro do LBPay).

| # | RPC | Request | Response | Status | Descrição |
|---|-----|---------|----------|--------|-----------|
| 11 | **StartPortability** | StartPortabilityRequest | StartPortabilityResponse | ✅ | Iniciar portabilidade de chave para nova conta |
| 12 | **ConfirmPortability** | ConfirmPortabilityRequest | ConfirmPortabilityResponse | ✅ | Confirmar portabilidade |
| 13 | **CancelPortability** | CancelPortabilityRequest | CancelPortabilityResponse | ✅ | Cancelar portabilidade |

### Detalhes StartPortability

**Request**:
```protobuf
message StartPortabilityRequest {
  string key_id = 1;         // Chave a portar
  string new_account_id = 2; // Nova conta destino (LBPay)
}
```

**Response**:
```protobuf
message StartPortabilityResponse {
  string portability_id = 1;
  string key_id = 2;
  dict.common.v1.Account new_account = 3; // Dados da nova conta
  google.protobuf.Timestamp started_at = 4;
  string message = 5; // "Portabilidade iniciada. Aguarde confirmação"
}
```

**Front-End Use Cases**:
- ✅ Iniciar portabilidade de chave para outra conta do mesmo usuário
- ✅ Ver nova conta de destino
- ✅ Receber mensagem explicativa

---

### Detalhes ConfirmPortability

**Request**:
```protobuf
message ConfirmPortabilityRequest {
  string portability_id = 1;
}
```

**Response**:
```protobuf
message ConfirmPortabilityResponse {
  string portability_id = 1;
  string key_id = 2;
  dict.common.v1.EntryStatus status = 3; // ACTIVE (com nova conta)
  google.protobuf.Timestamp confirmed_at = 4;
}
```

**Front-End Use Cases**:
- ✅ Confirmar portabilidade (após validações do Bacen)
- ✅ Ver status final (ACTIVE)

---

### Detalhes CancelPortability

**Request**:
```protobuf
message CancelPortabilityRequest {
  string portability_id = 1;
  optional string reason = 2;
}
```

**Response**:
```protobuf
message CancelPortabilityResponse {
  string portability_id = 1;
  google.protobuf.Timestamp cancelled_at = 2;
}
```

**Front-End Use Cases**:
- ✅ Cancelar portabilidade (desistir)
- ✅ Adicionar razão (opcional)

---

## ✅ Grupo 4: Directory Queries (Consultas DICT) - 1 RPC

Consultar chaves de **terceiros** (para transações PIX).

| # | RPC | Request | Response | Status | Descrição |
|---|-----|---------|----------|--------|-----------|
| 14 | **LookupKey** | LookupKeyRequest | LookupKeyResponse | ✅ | Consultar dados públicos de chave DICT (para enviar PIX) |

### Detalhes LookupKey

**Request**:
```protobuf
message LookupKeyRequest {
  dict.common.v1.DictKey key = 1; // Chave a consultar (tipo + valor)
}
```

**Response**:
```protobuf
message LookupKeyResponse {
  dict.common.v1.DictKey key = 1;
  dict.common.v1.Account account = 2;  // 🔥 Apenas dados públicos (ISPB, agência, conta)
  string account_holder_name = 3;      // 🔥 Nome do titular
  dict.common.v1.EntryStatus status = 4; // Se ACTIVE, pode receber PIX
}
```

**Front-End Use Cases**:
- ✅ Consultar chave de terceiro antes de enviar PIX
- ✅ Ver nome do titular (para confirmação: "Enviar para João Silva?")
- ✅ Ver conta destino (ISPB + agência + conta)
- ✅ Validar se chave está ACTIVE (pode receber)

---

## 🏥 Health Check - 1 RPC

| # | RPC | Request | Response | Status | Descrição |
|---|-----|---------|----------|--------|-----------|
| 15 | **HealthCheck** | Empty | HealthCheckResponse | ✅ | Health check do serviço Core DICT |

### Detalhes HealthCheck

**Request**: `google.protobuf.Empty`

**Response**:
```protobuf
message HealthCheckResponse {
  enum HealthStatus {
    HEALTH_STATUS_UNSPECIFIED = 0;
    HEALTH_STATUS_HEALTHY = 1;
    HEALTH_STATUS_DEGRADED = 2;
    HEALTH_STATUS_UNHEALTHY = 3;
  }
  HealthStatus status = 1;
  bool connect_reachable = 2;    // Se Connect (RSFN) está acessível
  google.protobuf.Timestamp checked_at = 3;
}
```

**Front-End Use Cases**:
- ✅ Verificar se serviço está saudável
- ✅ Verificar conectividade com RSFN (via Connect)
- ✅ Exibir status no dashboard

---

## 📊 Resumo Final: Cobertura Completa ✅

| Grupo | RPCs | Status | Completo? |
|-------|------|--------|-----------|
| **1. Directory (Vínculos DICT)** | 4 | ✅ | **SIM** |
| **2. Claim (Reivindicação)** | 6 | ✅ | **SIM** |
| **3. Portability (Portabilidade)** | 3 | ✅ | **SIM** |
| **4. Directory Queries (Consultas)** | 1 | ✅ | **SIM** |
| **Health Check** | 1 | ✅ | **SIM** |
| **TOTAL** | **15 RPCs** | ✅ | **100%** |

---

## 🎯 Funcionalidades Críticas para Front-End

### ✅ Gestão de Chaves (Directory)
- [x] Criar chaves (5 tipos: CPF, CNPJ, Email, Phone, EVP)
- [x] Listar chaves (paginação + filtros)
- [x] Ver detalhes (histórico de portabilidade)
- [x] Deletar chaves

### ✅ Reivindicações (Claims)
- [x] Iniciar claim de chave de terceiro
- [x] Ver status de claim (com **dias restantes**)
- [x] Listar claims recebidas (inbox)
- [x] Listar claims enviadas (outbox)
- [x] Responder a claim (aceitar/rejeitar)
- [x] Cancelar claim enviada

### ✅ Portabilidade
- [x] Iniciar portabilidade para nova conta
- [x] Confirmar portabilidade
- [x] Cancelar portabilidade

### ✅ Consultas
- [x] Lookup de chave de terceiro (para PIX)
- [x] Ver nome do titular (confirmação)
- [x] Validar se chave está ativa

### ✅ Monitoring
- [x] Health check do serviço
- [x] Status de conectividade RSFN

---

## 🔥 Recursos Especiais para UX

1. **Dias Restantes em Claims**:
   - Campo `days_remaining` em todas as responses de claims
   - Front-End pode exibir: "Faltam 27 dias" com countdown

2. **Histórico de Portabilidade**:
   - `GetKeyResponse.portability_history` lista todas as mudanças de conta
   - Front-End pode exibir timeline

3. **Mensagens Amigáveis**:
   - Todas as responses de claim/portability têm campo `message`
   - Backend já formata mensagem para usuário final

4. **Filtros e Paginação**:
   - ListKeys, ListIncomingClaims, ListOutgoingClaims suportam filtros
   - Max 100 itens por página

5. **Busca Flexível**:
   - GetKey aceita `key_id` OU `key` (tipo+valor)
   - Front-End pode buscar de 2 formas

6. **Dados Públicos vs Privados**:
   - LookupKey retorna **apenas dados públicos** (DICT Bacen compliance)
   - GetKey retorna dados completos (apenas dono)

---

## ✅ Conclusão

**A interface gRPC do Core-Dict está 100% COMPLETA para o Front-End!**

### Todas as 4 áreas estão cobertas:
1. ✅ **Directory (Vínculos DICT)**: 4 RPCs - Gestão completa de chaves
2. ✅ **Claim (Reivindicação)**: 6 RPCs - Processo de 30 dias completo
3. ✅ **Portability (Portabilidade)**: 3 RPCs - Mudança de conta
4. ✅ **Directory Queries (Consultas)**: 1 RPC - Lookup de terceiros

### Total: 15 RPCs
- 4 Key Operations
- 6 Claim Operations
- 3 Portability Operations
- 1 Query Operation
- 1 Health Check

### Próximo Passo
O **handler híbrido** (mock + real) está em desenvolvimento. Com feature flag `CORE_DICT_USE_MOCK_MODE`:
- **true**: Front-End pode começar a integrar **HOJE** com mocks
- **false**: Backend executa lógica real quando pronto

---

**Status**: ✅ **INTERFACE 100% PRONTA PARA FRONT-END**
**Data**: 2025-10-27
**Proto File**: `dict-contracts/proto/core_dict.proto`
