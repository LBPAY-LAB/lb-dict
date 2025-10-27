# Core DICT gRPC - Quick Start Guide

**Status**: ✅ Mock Mode 100% funcional e pronto para testes Front-End

---

## 🚀 Inicio Rápido (5 minutos)

### 1. Compilar Servidor

```bash
cd /Users/jose.silva.lb/LBPay/IA_Dict/core-dict

# Compilar
go build -o bin/core-dict-grpc ./cmd/grpc/

# Verificar binário criado
ls -lh bin/core-dict-grpc
# Esperado: ~25 MB
```

### 2. Iniciar em Mock Mode

```bash
# Modo Mock (sem infraestrutura)
CORE_DICT_USE_MOCK_MODE=true \
GRPC_PORT=9090 \
LOG_LEVEL=info \
./bin/core-dict-grpc
```

**Logs Esperados**:
```json
{"level":"INFO","msg":"Starting Core DICT gRPC Server","port":"9090","mock_mode":true,"version":"1.0.0"}
{"level":"WARN","msg":"⚠️  MOCK MODE ENABLED - Using mock responses for all RPCs"}
{"level":"INFO","msg":"✅ CoreDictService registered (MOCK MODE)"}
{"level":"INFO","msg":"✅ Health Check service registered"}
{"level":"INFO","msg":"✅ gRPC Reflection enabled (for grpcurl)"}
{"level":"INFO","msg":"🚀 gRPC server listening","address":"[::]:9090"}
```

**Server pronto!** Ctrl+C para parar.

---

## 🧪 Testar com grpcurl

### Instalar grpcurl (se necessário)

```bash
# macOS
brew install grpcurl

# Linux
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Verificar instalação
grpcurl --version
```

### Listar Serviços Disponíveis

```bash
grpcurl -plaintext localhost:9090 list
```

**Output esperado**:
```
dict.core.v1.CoreDictService
grpc.health.v1.Health
grpc.reflection.v1alpha.ServerReflection
```

### Listar Métodos do CoreDictService

```bash
grpcurl -plaintext localhost:9090 list dict.core.v1.CoreDictService
```

**Output esperado (15 métodos)**:
```
dict.core.v1.CoreDictService.CancelClaim
dict.core.v1.CoreDictService.CancelPortability
dict.core.v1.CoreDictService.ConfirmPortability
dict.core.v1.CoreDictService.CreateKey
dict.core.v1.CoreDictService.DeleteKey
dict.core.v1.CoreDictService.GetClaimStatus
dict.core.v1.CoreDictService.GetKey
dict.core.v1.CoreDictService.HealthCheck
dict.core.v1.CoreDictService.ListIncomingClaims
dict.core.v1.CoreDictService.ListKeys
dict.core.v1.CoreDictService.ListOutgoingClaims
dict.core.v1.CoreDictService.LookupKey
dict.core.v1.CoreDictService.RespondToClaim
dict.core.v1.CoreDictService.StartClaim
dict.core.v1.CoreDictService.StartPortability
```

---

## 📞 Exemplos de Chamadas (Mock Mode)

### 1. Health Check

```bash
grpcurl -plaintext localhost:9090 dict.core.v1.CoreDictService/HealthCheck
```

**Response**:
```json
{
  "status": "HEALTH_STATUS_HEALTHY",
  "connectReachable": true,
  "checkedAt": "2025-10-27T14:24:48Z"
}
```

### 2. CreateKey - CPF

```bash
grpcurl -plaintext \
  -d '{
    "key_type": "KEY_TYPE_CPF",
    "key_value": "12345678900",
    "account_id": "550e8400-e29b-41d4-a716-446655440000"
  }' \
  localhost:9090 dict.core.v1.CoreDictService/CreateKey
```

**Response**:
```json
{
  "keyId": "mock-key-1730039080",
  "key": {
    "keyType": "KEY_TYPE_CPF",
    "keyValue": "12345678900"
  },
  "status": "ENTRY_STATUS_ACTIVE",
  "createdAt": "2025-10-27T14:24:40Z"
}
```

### 3. CreateKey - Email

```bash
grpcurl -plaintext \
  -d '{
    "key_type": "KEY_TYPE_EMAIL",
    "key_value": "usuario@example.com",
    "account_id": "550e8400-e29b-41d4-a716-446655440000"
  }' \
  localhost:9090 dict.core.v1.CoreDictService/CreateKey
```

### 4. CreateKey - Telefone

```bash
grpcurl -plaintext \
  -d '{
    "key_type": "KEY_TYPE_PHONE",
    "key_value": "+5511999999999",
    "account_id": "550e8400-e29b-41d4-a716-446655440000"
  }' \
  localhost:9090 dict.core.v1.CoreDictService/CreateKey
```

### 5. CreateKey - CNPJ

```bash
grpcurl -plaintext \
  -d '{
    "key_type": "KEY_TYPE_CNPJ",
    "key_value": "12345678000195",
    "account_id": "550e8400-e29b-41d4-a716-446655440000"
  }' \
  localhost:9090 dict.core.v1.CoreDictService/CreateKey
```

### 6. CreateKey - EVP (chave aleatória)

```bash
grpcurl -plaintext \
  -d '{
    "key_type": "KEY_TYPE_EVP",
    "account_id": "550e8400-e29b-41d4-a716-446655440000"
  }' \
  localhost:9090 dict.core.v1.CoreDictService/CreateKey
```

**Note**: `key_value` é opcional para EVP, será gerado automaticamente.

### 7. ListKeys - Com Paginação

```bash
grpcurl -plaintext \
  -d '{
    "page_size": 10,
    "page_token": ""
  }' \
  localhost:9090 dict.core.v1.CoreDictService/ListKeys
```

**Response**:
```json
{
  "keys": [
    {
      "keyId": "key-1",
      "key": {
        "keyType": "KEY_TYPE_CPF",
        "keyValue": "12345678900"
      },
      "status": "ENTRY_STATUS_ACTIVE",
      "accountId": "mock-account-id",
      "createdAt": "2025-10-27T14:24:48Z",
      "updatedAt": "2025-10-27T14:24:48Z"
    }
  ],
  "nextPageToken": "",
  "totalCount": 1
}
```

### 8. ListKeys - Com Filtro por Status

```bash
grpcurl -plaintext \
  -d '{
    "page_size": 20,
    "status": "ENTRY_STATUS_ACTIVE"
  }' \
  localhost:9090 dict.core.v1.CoreDictService/ListKeys
```

### 9. GetKey - Por key_id

```bash
grpcurl -plaintext \
  -d '{
    "key_id": "550e8400-e29b-41d4-a716-446655440000"
  }' \
  localhost:9090 dict.core.v1.CoreDictService/GetKey
```

### 10. GetKey - Por key (tipo + valor)

```bash
grpcurl -plaintext \
  -d '{
    "key": {
      "key_type": "KEY_TYPE_CPF",
      "key_value": "12345678900"
    }
  }' \
  localhost:9090 dict.core.v1.CoreDictService/GetKey
```

**Response**:
```json
{
  "keyId": "mock-key-123",
  "key": {
    "keyType": "KEY_TYPE_CPF",
    "keyValue": "12345678900"
  },
  "account": {
    "ispb": "12345678",
    "branch": "0001",
    "accountNumber": "123456",
    "accountType": "ACCOUNT_TYPE_CHECKING"
  },
  "status": "ENTRY_STATUS_ACTIVE",
  "createdAt": "2025-10-27T14:24:48Z",
  "updatedAt": "2025-10-27T14:24:48Z"
}
```

### 11. DeleteKey

```bash
grpcurl -plaintext \
  -d '{
    "key_id": "550e8400-e29b-41d4-a716-446655440000"
  }' \
  localhost:9090 dict.core.v1.CoreDictService/DeleteKey
```

**Response**:
```json
{
  "deleted": true,
  "deletedAt": "2025-10-27T14:24:48Z"
}
```

### 12. StartClaim - Reivindicar Chave (30 dias)

```bash
grpcurl -plaintext \
  -d '{
    "key": {
      "key_type": "KEY_TYPE_EMAIL",
      "key_value": "usuario@example.com"
    },
    "account_id": "550e8400-e29b-41d4-a716-446655440001"
  }' \
  localhost:9090 dict.core.v1.CoreDictService/StartClaim
```

**Response**:
```json
{
  "claimId": "mock-claim-1730039088",
  "entryId": "mock-entry-123",
  "status": "CLAIM_STATUS_OPEN",
  "expiresAt": "2025-11-26T14:24:48Z",
  "createdAt": "2025-10-27T14:24:48Z",
  "message": "Claim criada com sucesso. O dono atual tem 30 dias para responder."
}
```

### 13. GetClaimStatus

```bash
grpcurl -plaintext \
  -d '{
    "claim_id": "mock-claim-1730039088"
  }' \
  localhost:9090 dict.core.v1.CoreDictService/GetClaimStatus
```

**Response**:
```json
{
  "claimId": "mock-claim-1730039088",
  "entryId": "mock-entry-123",
  "key": {
    "keyType": "KEY_TYPE_EMAIL",
    "keyValue": "usuario@example.com"
  },
  "status": "CLAIM_STATUS_OPEN",
  "claimerIspb": "12345678",
  "ownerIspb": "87654321",
  "createdAt": "2025-10-27T14:24:48Z",
  "expiresAt": "2025-11-26T14:24:48Z",
  "daysRemaining": 30
}
```

### 14. ListIncomingClaims - Claims recebidas

```bash
grpcurl -plaintext \
  -d '{
    "page_size": 10,
    "status": "CLAIM_STATUS_OPEN"
  }' \
  localhost:9090 dict.core.v1.CoreDictService/ListIncomingClaims
```

### 15. ListOutgoingClaims - Claims enviadas

```bash
grpcurl -plaintext \
  -d '{
    "page_size": 10
  }' \
  localhost:9090 dict.core.v1.CoreDictService/ListOutgoingClaims
```

### 16. RespondToClaim - Aceitar Claim

```bash
grpcurl -plaintext \
  -d '{
    "claim_id": "mock-claim-123",
    "response": "CLAIM_RESPONSE_ACCEPT"
  }' \
  localhost:9090 dict.core.v1.CoreDictService/RespondToClaim
```

**Response**:
```json
{
  "claimId": "mock-claim-123",
  "newStatus": "CLAIM_STATUS_CONFIRMED",
  "respondedAt": "2025-10-27T14:24:48Z",
  "message": "Claim aceita com sucesso. A chave será transferida."
}
```

### 17. RespondToClaim - Rejeitar Claim

```bash
grpcurl -plaintext \
  -d '{
    "claim_id": "mock-claim-123",
    "response": "CLAIM_RESPONSE_REJECT",
    "reason": "Chave em uso ativo"
  }' \
  localhost:9090 dict.core.v1.CoreDictService/RespondToClaim
```

**Response**:
```json
{
  "claimId": "mock-claim-123",
  "newStatus": "CLAIM_STATUS_CANCELLED",
  "respondedAt": "2025-10-27T14:24:48Z",
  "message": "Claim rejeitada."
}
```

### 18. CancelClaim - Cancelar Claim Enviada

```bash
grpcurl -plaintext \
  -d '{
    "claim_id": "mock-claim-123",
    "reason": "Mudou de ideia"
  }' \
  localhost:9090 dict.core.v1.CoreDictService/CancelClaim
```

### 19. StartPortability - Iniciar Portabilidade

```bash
grpcurl -plaintext \
  -d '{
    "key_id": "550e8400-e29b-41d4-a716-446655440000",
    "new_account_id": "550e8400-e29b-41d4-a716-446655440002"
  }' \
  localhost:9090 dict.core.v1.CoreDictService/StartPortability
```

**Response**:
```json
{
  "portabilityId": "mock-port-1730039088",
  "keyId": "550e8400-e29b-41d4-a716-446655440000",
  "newAccount": {
    "ispb": "12345678",
    "branch": "0002",
    "accountNumber": "654321",
    "accountType": "ACCOUNT_TYPE_CHECKING"
  },
  "startedAt": "2025-10-27T14:24:48Z",
  "message": "Portabilidade iniciada. Aguarde confirmação."
}
```

### 20. ConfirmPortability

```bash
grpcurl -plaintext \
  -d '{
    "portability_id": "mock-port-1730039088"
  }' \
  localhost:9090 dict.core.v1.CoreDictService/ConfirmPortability
```

### 21. CancelPortability

```bash
grpcurl -plaintext \
  -d '{
    "portability_id": "mock-port-1730039088",
    "reason": "Dados incorretos"
  }' \
  localhost:9090 dict.core.v1.CoreDictService/CancelPortability
```

### 22. LookupKey - Consultar Chave de Terceiro

```bash
grpcurl -plaintext \
  -d '{
    "key": {
      "key_type": "KEY_TYPE_PHONE",
      "key_value": "+5511999999999"
    }
  }' \
  localhost:9090 dict.core.v1.CoreDictService/LookupKey
```

**Response**:
```json
{
  "key": {
    "keyType": "KEY_TYPE_PHONE",
    "keyValue": "+5511999999999"
  },
  "account": {
    "ispb": "12345678",
    "branch": "0001",
    "accountNumber": "123456",
    "accountType": "ACCOUNT_TYPE_CHECKING"
  },
  "accountHolderName": "Nome do Titular",
  "status": "ENTRY_STATUS_ACTIVE"
}
```

---

## 🔍 Inspecionar Mensagens Proto

### Ver estrutura de uma Request

```bash
grpcurl -plaintext localhost:9090 describe dict.core.v1.CreateKeyRequest
```

### Ver estrutura de uma Response

```bash
grpcurl -plaintext localhost:9090 describe dict.core.v1.CreateKeyResponse
```

### Ver todos os enums de KeyType

```bash
grpcurl -plaintext localhost:9090 describe dict.common.v1.KeyType
```

---

## 🐛 Troubleshooting

### Server não inicia

**Erro**: `bind: address already in use`

**Solução**:
```bash
# Matar processo na porta 9090
lsof -ti:9090 | xargs kill -9

# Ou usar outra porta
GRPC_PORT=9091 ./bin/core-dict-grpc
```

### grpcurl não encontra serviço

**Erro**: `Failed to dial target host`

**Verificar**:
```bash
# Server está rodando?
ps aux | grep core-dict-grpc

# Porta correta?
lsof -i :9090
```

### Reflection não funciona

**Solução**:
```bash
# Servidor já tem reflection habilitado
# Verificar logs de startup para confirmar:
# "✅ gRPC Reflection enabled (for grpcurl)"
```

---

## 🎯 Casos de Teste Recomendados

### Happy Path

1. CreateKey (CPF) → Sucesso
2. GetKey (por key_id) → Retorna chave
3. ListKeys → Lista com 1 key
4. StartClaim → Cria claim
5. GetClaimStatus → Status OPEN
6. RespondToClaim (Accept) → Claim CONFIRMED
7. StartPortability → Inicia
8. ConfirmPortability → Confirma
9. LookupKey → Retorna dados públicos
10. DeleteKey → Sucesso

### Edge Cases

1. CreateKey sem key_type → Error InvalidArgument
2. CreateKey key_value vazio → Error InvalidArgument
3. GetKey com key_id inválido → Mock retorna sucesso (por enquanto)
4. RespondToClaim com UNSPECIFIED → Error InvalidArgument

---

## 📚 Recursos Adicionais

### Documentação Proto

- [dict-contracts/proto/core_dict.proto](../dict-contracts/proto/core_dict.proto)
- [dict-contracts/proto/common.proto](../dict-contracts/proto/common.proto)

### Documentação Detalhada

- [VALIDACAO_INTERFACE_GRPC_FRONTEND.md](../Artefatos/00_Master/VALIDACAO_INTERFACE_GRPC_FRONTEND.md)
- [REAL_MODE_STATUS_FINAL.md](../Artefatos/00_Master/REAL_MODE_STATUS_FINAL.md)

### Logs do Servidor

```bash
# Ver logs em tempo real
tail -f /tmp/grpc-server.log

# Ver apenas requests gRPC
tail -f /tmp/grpc-server.log | grep "gRPC request"
```

---

## ✅ Checklist Front-End

- [ ] Servidor Mock Mode rodando
- [ ] grpcurl instalado e testado
- [ ] Testado CreateKey (5 tipos: CPF, CNPJ, Email, Phone, EVP)
- [ ] Testado ListKeys com paginação
- [ ] Testado GetKey (por ID e por key)
- [ ] Testado DeleteKey
- [ ] Testado StartClaim
- [ ] Testado GetClaimStatus
- [ ] Testado RespondToClaim (Accept + Reject)
- [ ] Testado CancelClaim
- [ ] Testado StartPortability
- [ ] Testado ConfirmPortability
- [ ] Testado LookupKey
- [ ] Testado HealthCheck

**Quando todos checados**: Front-End pronto para começar integração!

---

**Última Atualização**: 2025-10-27
**Status**: ✅ Pronto para uso
**Suporte**: Backend Team

