# GRPC-003: Proto Files Specification

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Versão**: 1.0
**Data**: 2025-10-25
**Status**: ✅ Especificação Completa
**Responsável**: ARCHITECT (AI Agent - Technical Architect)

---

## 📋 Resumo Executivo

Este documento especifica **todos os arquivos Protocol Buffers (.proto)** necessários para a comunicação gRPC entre os componentes do sistema DICT:

1. **Connect ↔ Bridge** (gRPC)
2. **FrontEnd ↔ Core DICT** (gRPC - futuro)
3. **Common types** (compartilhados)

**Objetivo**: Fornecer especificação completa para desenvolvedores gerarem código gRPC em Go usando `protoc` e `protoc-gen-go-grpc`.

---

## 🎯 Arquitetura de Proto Files

```
proto/
├── common/
│   └── v1/
│       ├── types.proto          # Tipos comuns (Account, Address, etc.)
│       ├── errors.proto         # Error details padronizados
│       └── timestamps.proto     # Re-export google.protobuf.Timestamp
├── bridge/
│   └── v1/
│       ├── bridge.proto         # BridgeService (Connect → Bridge)
│       ├── messages.proto       # Request/Response messages
│       └── enums.proto          # Enums específicos do Bridge
├── core/
│   └── v1/
│       ├── core.proto           # CoreDictService (FrontEnd → Core)
│       ├── messages.proto       # Request/Response messages
│       └── enums.proto          # Enums específicos do Core
└── buf.yaml                     # Buf configuration (linting, breaking changes)
```

---

## 📄 Proto File 1: `common/v1/types.proto`

### Propósito
Tipos comuns compartilhados entre Bridge e Core DICT (Account, Address, Key)

### Conteúdo Completo
```protobuf
syntax = "proto3";

package rsfn.common.v1;

option go_package = "github.com/lbpay/dict/proto/common/v1;commonv1";

import "google/protobuf/timestamp.proto";

// ====================================================================
// ACCOUNT - Representa uma conta bancária no SPB
// ====================================================================
message Account {
  // ISPB da instituição (8 dígitos)
  string ispb = 1;

  // Tipo de conta
  AccountType account_type = 2;

  // Número da conta (sem dígito verificador)
  string account_number = 3;

  // Dígito verificador da conta
  string account_check_digit = 4;

  // Agência (branch code)
  string branch_code = 5;

  // Nome do titular da conta
  string account_holder_name = 6;

  // CPF/CNPJ do titular
  string account_holder_document = 7;

  // Tipo de documento (CPF ou CNPJ)
  DocumentType document_type = 8;
}

// ====================================================================
// DICT KEY - Representa uma chave DICT (PIX)
// ====================================================================
message DictKey {
  // Tipo de chave
  KeyType key_type = 1;

  // Valor da chave (CPF, phone, email, EVP)
  string key_value = 2;
}

// ====================================================================
// ADDRESS - Endereço do titular da conta
// ====================================================================
message Address {
  string street = 1;
  string number = 2;
  string complement = 3;
  string neighborhood = 4;
  string city = 5;
  string state = 6;  // UF (2 letras)
  string postal_code = 7;  // CEP
  string country = 8;  // BR
}

// ====================================================================
// ENUMS - Tipos comuns
// ====================================================================

// Tipo de conta bancária
enum AccountType {
  ACCOUNT_TYPE_UNSPECIFIED = 0;
  ACCOUNT_TYPE_CHECKING = 1;    // Conta Corrente (CACC)
  ACCOUNT_TYPE_SAVINGS = 2;     // Conta Poupança (SVGS)
  ACCOUNT_TYPE_PAYMENT = 3;     // Conta Pagamento (TRAN)
  ACCOUNT_TYPE_SALARY = 4;      // Conta Salário (SLRY)
}

// Tipo de chave DICT
enum KeyType {
  KEY_TYPE_UNSPECIFIED = 0;
  KEY_TYPE_CPF = 1;
  KEY_TYPE_CNPJ = 2;
  KEY_TYPE_PHONE = 3;
  KEY_TYPE_EMAIL = 4;
  KEY_TYPE_EVP = 5;  // Chave aleatória (UUID)
}

// Tipo de documento (CPF ou CNPJ)
enum DocumentType {
  DOCUMENT_TYPE_UNSPECIFIED = 0;
  DOCUMENT_TYPE_CPF = 1;
  DOCUMENT_TYPE_CNPJ = 2;
}

// Status de uma entry DICT
enum EntryStatus {
  ENTRY_STATUS_UNSPECIFIED = 0;
  ENTRY_STATUS_ACTIVE = 1;
  ENTRY_STATUS_PORTABILITY_PENDING = 2;
  ENTRY_STATUS_PORTABILITY_CONFIRMED = 3;
  ENTRY_STATUS_CLAIM_PENDING = 4;
  ENTRY_STATUS_DELETED = 5;
}

// Status de uma claim (reivindicação)
enum ClaimStatus {
  CLAIM_STATUS_UNSPECIFIED = 0;
  CLAIM_STATUS_OPEN = 1;
  CLAIM_STATUS_WAITING_RESOLUTION = 2;
  CLAIM_STATUS_CONFIRMED = 3;
  CLAIM_STATUS_CANCELLED = 4;
  CLAIM_STATUS_COMPLETED = 5;
  CLAIM_STATUS_EXPIRED = 6;
}

// Tipo de operação (para auditoria)
enum OperationType {
  OPERATION_TYPE_UNSPECIFIED = 0;
  OPERATION_TYPE_CREATE = 1;
  OPERATION_TYPE_UPDATE = 2;
  OPERATION_TYPE_DELETE = 3;
  OPERATION_TYPE_CLAIM = 4;
  OPERATION_TYPE_PORTABILITY = 5;
}
```

---

## 📄 Proto File 2: `common/v1/errors.proto`

### Propósito
Estruturas de erro padronizadas (seguindo Google API Design Guide)

### Conteúdo Completo
```protobuf
syntax = "proto3";

package rsfn.common.v1;

option go_package = "github.com/lbpay/dict/proto/common/v1;commonv1";

// ====================================================================
// ERROR DETAILS - Informações adicionais sobre erros
// ====================================================================

// Detalhes de erro de validação
message ValidationError {
  // Campo que falhou na validação
  string field = 1;

  // Descrição do erro
  string description = 2;

  // Constraint violada (e.g., "required", "max_length", "pattern")
  string constraint = 3;

  // Valor fornecido (para debugging)
  string provided_value = 4;
}

// Detalhes de erro de negócio
message BusinessError {
  // Código de erro de negócio (e.g., "CLAIM_ALREADY_EXISTS")
  string error_code = 1;

  // Mensagem legível
  string message = 2;

  // Contexto adicional
  map<string, string> context = 3;
}

// Detalhes de erro de infraestrutura
message InfrastructureError {
  // Componente que falhou (e.g., "PostgreSQL", "Redis", "Bridge")
  string component = 1;

  // Tipo de erro (e.g., "CONNECTION_TIMEOUT", "UNAVAILABLE")
  string error_type = 2;

  // Mensagem original do erro
  string original_message = 3;

  // Retry possível?
  bool retriable = 4;
}

// Detalhes de erro do Bacen (via Bridge)
message BacenError {
  // Código de erro do Bacen (SOAP fault code)
  string bacen_code = 1;

  // Mensagem do Bacen
  string bacen_message = 2;

  // ID da requisição ao Bacen
  string bacen_request_id = 3;

  // Timestamp do erro
  google.protobuf.Timestamp occurred_at = 4;
}

// ====================================================================
// ERROR RESPONSE - Wrapper para todos os erros
// ====================================================================
message ErrorResponse {
  // Código de status gRPC (mapeado de grpc.Code)
  int32 grpc_code = 1;

  // Mensagem principal
  string message = 2;

  // Detalhes específicos (usar oneof para tipo seguro)
  oneof details {
    ValidationError validation = 3;
    BusinessError business = 4;
    InfrastructureError infrastructure = 5;
    BacenError bacen = 6;
  }

  // Request ID para rastreamento (correlação com logs)
  string request_id = 7;

  // Timestamp do erro
  google.protobuf.Timestamp timestamp = 8;
}
```

---

## 📄 Proto File 3: `bridge/v1/bridge.proto`

### Propósito
Definição do serviço gRPC BridgeService (Connect → Bridge)

### Conteúdo Completo
```protobuf
syntax = "proto3";

package rsfn.bridge.v1;

option go_package = "github.com/lbpay/dict/proto/bridge/v1;bridgev1";

import "common/v1/types.proto";
import "common/v1/errors.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/empty.proto";

// ====================================================================
// BRIDGE SERVICE - Comunicação Connect → Bridge → Bacen DICT
// ====================================================================
service BridgeService {
  // ========== Operações de Entry (Chave DICT) ==========

  // Criar nova chave DICT
  rpc CreateEntry(CreateEntryRequest) returns (CreateEntryResponse);

  // Buscar chave DICT existente
  rpc GetEntry(GetEntryRequest) returns (GetEntryResponse);

  // Deletar chave DICT
  rpc DeleteEntry(DeleteEntryRequest) returns (DeleteEntryResponse);

  // Atualizar dados da conta vinculada à chave
  rpc UpdateEntry(UpdateEntryRequest) returns (UpdateEntryResponse);

  // ========== Operações de Claim (Reivindicação - 30 dias) ==========

  // Criar nova claim (portabilidade/ownership)
  rpc CreateClaim(CreateClaimRequest) returns (CreateClaimResponse);

  // Buscar status de claim
  rpc GetClaim(GetClaimRequest) returns (GetClaimResponse);

  // Completar claim (confirmação pelo dono)
  rpc CompleteClaim(CompleteClaimRequest) returns (CompleteClaimResponse);

  // Cancelar claim (rejeição ou timeout)
  rpc CancelClaim(CancelClaimRequest) returns (CancelClaimResponse);

  // ========== Operações de Portabilidade ==========

  // Confirmar portabilidade de conta
  rpc ConfirmPortability(ConfirmPortabilityRequest) returns (ConfirmPortabilityResponse);

  // Cancelar portabilidade
  rpc CancelPortability(CancelPortabilityRequest) returns (CancelPortabilityResponse);

  // ========== Health Check ==========

  // Health check do Bridge (verifica conectividade com Bacen)
  rpc HealthCheck(google.protobuf.Empty) returns (HealthCheckResponse);
}

// ====================================================================
// ENTRY OPERATIONS - Messages
// ====================================================================

message CreateEntryRequest {
  // Chave DICT a ser criada
  rsfn.common.v1.DictKey key = 1;

  // Conta a ser vinculada
  rsfn.common.v1.Account account = 2;

  // Idempotency key (para retry safety)
  string idempotency_key = 3;

  // Request ID (rastreamento)
  string request_id = 4;
}

message CreateEntryResponse {
  // ID da entry criada (UUID)
  string entry_id = 1;

  // ID externo do Bacen
  string external_id = 2;

  // Status da entry
  rsfn.common.v1.EntryStatus status = 3;

  // Timestamp de criação
  google.protobuf.Timestamp created_at = 4;
}

message GetEntryRequest {
  // Buscar por chave
  rsfn.common.v1.DictKey key = 1;

  // Request ID
  string request_id = 2;
}

message GetEntryResponse {
  // ID da entry
  string entry_id = 1;

  // Chave DICT
  rsfn.common.v1.DictKey key = 2;

  // Conta vinculada
  rsfn.common.v1.Account account = 3;

  // Status
  rsfn.common.v1.EntryStatus status = 4;

  // Timestamps
  google.protobuf.Timestamp created_at = 5;
  google.protobuf.Timestamp updated_at = 6;
}

message DeleteEntryRequest {
  // ID da entry a deletar
  string entry_id = 1;

  // Chave DICT (alternativa ao entry_id)
  rsfn.common.v1.DictKey key = 2;

  // Idempotency key
  string idempotency_key = 3;

  // Request ID
  string request_id = 4;
}

message DeleteEntryResponse {
  // Confirmação de deleção
  bool deleted = 1;

  // Timestamp da deleção
  google.protobuf.Timestamp deleted_at = 2;
}

message UpdateEntryRequest {
  // ID da entry a atualizar
  string entry_id = 1;

  // Nova conta (atualização parcial de campos)
  rsfn.common.v1.Account new_account = 2;

  // Idempotency key
  string idempotency_key = 3;

  // Request ID
  string request_id = 4;
}

message UpdateEntryResponse {
  // ID da entry atualizada
  string entry_id = 1;

  // Nova conta
  rsfn.common.v1.Account account = 2;

  // Timestamp da atualização
  google.protobuf.Timestamp updated_at = 3;
}

// ====================================================================
// CLAIM OPERATIONS - Messages (30 dias)
// ====================================================================

message CreateClaimRequest {
  // ID da entry a ser reivindicada
  string entry_id = 1;

  // ISPB do reivindicador
  string claimer_ispb = 2;

  // ISPB do dono atual
  string owner_ispb = 3;

  // Conta do reivindicador
  rsfn.common.v1.Account claimer_account = 4;

  // Período de conclusão (deve ser 30 dias - TEC-003 v2.1)
  int32 completion_period_days = 5;

  // Idempotency key
  string idempotency_key = 6;

  // Request ID
  string request_id = 7;
}

message CreateClaimResponse {
  // ID da claim criada (UUID)
  string claim_id = 1;

  // ID externo do Bacen
  string external_id = 2;

  // Status inicial (sempre "OPEN")
  rsfn.common.v1.ClaimStatus status = 3;

  // Data de expiração (created_at + 30 dias)
  google.protobuf.Timestamp expires_at = 4;

  // Timestamp de criação
  google.protobuf.Timestamp created_at = 5;
}

message GetClaimRequest {
  // ID da claim
  string claim_id = 1;

  // ID externo do Bacen (alternativa)
  string external_id = 2;

  // Request ID
  string request_id = 3;
}

message GetClaimResponse {
  // ID da claim
  string claim_id = 1;

  // ID da entry
  string entry_id = 2;

  // Status atual
  rsfn.common.v1.ClaimStatus status = 3;

  // ISPBs
  string claimer_ispb = 4;
  string owner_ispb = 5;

  // Timestamps
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp expires_at = 7;
  google.protobuf.Timestamp completed_at = 8;  // Null se não completa
}

message CompleteClaimRequest {
  // ID da claim a completar
  string claim_id = 1;

  // Confirmação explícita (safety check)
  bool confirmed = 2;

  // Idempotency key
  string idempotency_key = 3;

  // Request ID
  string request_id = 4;
}

message CompleteClaimResponse {
  // ID da claim
  string claim_id = 1;

  // Novo status (COMPLETED)
  rsfn.common.v1.ClaimStatus status = 2;

  // Timestamp de conclusão
  google.protobuf.Timestamp completed_at = 3;
}

message CancelClaimRequest {
  // ID da claim a cancelar
  string claim_id = 1;

  // Razão do cancelamento
  string reason = 2;

  // Idempotency key
  string idempotency_key = 3;

  // Request ID
  string request_id = 4;
}

message CancelClaimResponse {
  // ID da claim
  string claim_id = 1;

  // Novo status (CANCELLED)
  rsfn.common.v1.ClaimStatus status = 2;

  // Timestamp de cancelamento
  google.protobuf.Timestamp cancelled_at = 3;
}

// ====================================================================
// PORTABILITY OPERATIONS - Messages
// ====================================================================

message ConfirmPortabilityRequest {
  // ID da entry em portabilidade
  string entry_id = 1;

  // Nova conta de destino
  rsfn.common.v1.Account new_account = 2;

  // Idempotency key
  string idempotency_key = 3;

  // Request ID
  string request_id = 4;
}

message ConfirmPortabilityResponse {
  // ID da entry
  string entry_id = 1;

  // Novo status (ACTIVE)
  rsfn.common.v1.EntryStatus status = 2;

  // Nova conta
  rsfn.common.v1.Account account = 3;

  // Timestamp de confirmação
  google.protobuf.Timestamp confirmed_at = 4;
}

message CancelPortabilityRequest {
  // ID da entry
  string entry_id = 1;

  // Razão do cancelamento
  string reason = 2;

  // Idempotency key
  string idempotency_key = 3;

  // Request ID
  string request_id = 4;
}

message CancelPortabilityResponse {
  // ID da entry
  string entry_id = 1;

  // Status revertido (ACTIVE)
  rsfn.common.v1.EntryStatus status = 2;

  // Timestamp de cancelamento
  google.protobuf.Timestamp cancelled_at = 3;
}

// ====================================================================
// HEALTH CHECK
// ====================================================================

message HealthCheckResponse {
  // Status geral do Bridge
  HealthStatus status = 1;

  // Conectividade com Bacen DICT
  bool bacen_reachable = 2;

  // Certificado mTLS válido
  bool mtls_certificate_valid = 3;

  // Dias até expiração do certificado
  int32 certificate_days_until_expiry = 4;

  // Timestamp do health check
  google.protobuf.Timestamp checked_at = 5;
}

enum HealthStatus {
  HEALTH_STATUS_UNSPECIFIED = 0;
  HEALTH_STATUS_HEALTHY = 1;
  HEALTH_STATUS_DEGRADED = 2;
  HEALTH_STATUS_UNHEALTHY = 3;
}
```

---

## 📄 Proto File 4: `core/v1/core.proto` (Futuro)

### Propósito
Definição do serviço gRPC CoreDictService (FrontEnd → Core DICT)

### Conteúdo Completo
```protobuf
syntax = "proto3";

package rsfn.core.v1;

option go_package = "github.com/lbpay/dict/proto/core/v1;corev1";

import "common/v1/types.proto";
import "common/v1/errors.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/empty.proto";

// ====================================================================
// CORE DICT SERVICE - FrontEnd → Core DICT
// ====================================================================
service CoreDictService {
  // ========== User Operations ==========

  // Criar nova chave DICT para usuário
  rpc CreateKey(CreateKeyRequest) returns (CreateKeyResponse);

  // Listar chaves do usuário
  rpc ListKeys(ListKeysRequest) returns (ListKeysResponse);

  // Deletar chave DICT
  rpc DeleteKey(DeleteKeyRequest) returns (DeleteKeyResponse);

  // ========== Claim Operations ==========

  // Iniciar reivindicação de chave
  rpc StartClaim(StartClaimRequest) returns (StartClaimResponse);

  // Verificar status de claim
  rpc GetClaimStatus(GetClaimStatusRequest) returns (GetClaimStatusResponse);

  // Responder a claim (aceitar/rejeitar)
  rpc RespondToClaim(RespondToClaimRequest) returns (RespondToClaimResponse);

  // ========== Portability Operations ==========

  // Iniciar portabilidade de conta
  rpc StartPortability(StartPortabilityRequest) returns (StartPortabilityResponse);

  // Confirmar portabilidade
  rpc ConfirmPortability(ConfirmPortabilityRequest) returns (ConfirmPortabilityResponse);

  // ========== Query Operations ==========

  // Consultar chave DICT de terceiros (para transações PIX)
  rpc LookupKey(LookupKeyRequest) returns (LookupKeyResponse);
}

// Mensagens omitidas nesta versão (criar quando Core DICT for implementado)
// Placeholder messages para compilação:

message CreateKeyRequest {
  rsfn.common.v1.DictKey key = 1;
  string user_id = 2;
  rsfn.common.v1.Account account = 3;
}

message CreateKeyResponse {
  string key_id = 1;
  rsfn.common.v1.EntryStatus status = 2;
  google.protobuf.Timestamp created_at = 3;
}

// ... (outros messages serão especificados quando Core DICT for implementado)
```

---

## 🛠️ Configuração do Buf (Linting e Breaking Changes)

### Arquivo: `buf.yaml`

```yaml
version: v1
name: buf.build/lbpay/dict-protos
deps:
  - buf.build/googleapis/googleapis  # Para google.protobuf.Timestamp, Empty

lint:
  use:
    - DEFAULT
    - COMMENTS
    - FILE_LOWER_SNAKE_CASE
  except:
    - UNARY_RPC  # Permitir unary RPCs

breaking:
  use:
    - FILE
  except:
    - FIELD_SAME_LABEL  # Permitir optional → required (em desenvolvimento)
```

---

## 🔧 Code Generation

### Comando para Gerar Código Go

```bash
# Instalar ferramentas
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Gerar código (executar na raiz do projeto)
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  --proto_path=proto \
  proto/common/v1/*.proto \
  proto/bridge/v1/*.proto \
  proto/core/v1/*.proto
```

### Saída Esperada
```
proto/
├── common/v1/
│   ├── types.pb.go
│   └── errors.pb.go
├── bridge/v1/
│   ├── bridge.pb.go
│   └── bridge_grpc.pb.go
└── core/v1/
    ├── core.pb.go
    └── core_grpc.pb.go
```

---

## 📋 Checklist de Implementação

Para desenvolvedores:

- [ ] Criar estrutura de diretórios `proto/`
- [ ] Criar todos os arquivos .proto conforme especificado
- [ ] Instalar `protoc` e plugins Go
- [ ] Configurar `buf.yaml`
- [ ] Executar code generation
- [ ] Adicionar arquivos gerados ao `.gitignore` (só versionar .proto)
- [ ] Criar Makefile com target `make proto-gen`
- [ ] Validar que imports estão corretos
- [ ] Testar compilação dos arquivos gerados
- [ ] Criar testes unitários para serialização/deserialização

---

## 📚 Referências

### Documentos Internos
- [GRPC-001: Bridge gRPC Service](GRPC-001_Bridge_gRPC_Service.md) - Contrato Connect ↔ Bridge
- [TEC-002 v3.1: Bridge Specification](../../11_Especificacoes_Tecnicas/TEC-002_Bridge_Specification.md) - Dual Protocol Support
- [TEC-003 v2.1: Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md) - ClaimWorkflow 30 dias
- [DAT-001: Schema Database Core DICT](../../03_Dados/DAT-001_Schema_Database_Core_DICT.md) - Estrutura de dados

### Documentação Externa
- [Protocol Buffers Language Guide](https://protobuf.dev/programming-guides/proto3/)
- [gRPC Go Quick Start](https://grpc.io/docs/languages/go/quickstart/)
- [Google API Design Guide](https://cloud.google.com/apis/design)
- [Buf Documentation](https://buf.build/docs/introduction)

---

**Versão**: 1.0
**Status**: ✅ Especificação Completa (Aguardando implementação)
**Próxima Revisão**: Após code generation (validar tipos gerados)

---

**IMPORTANTE**: Este é um documento de **especificação técnica**. Os desenvolvedores usarão estes arquivos .proto para gerar código Go automaticamente usando `protoc`.
