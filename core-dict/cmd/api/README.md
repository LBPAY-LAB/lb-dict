# Core DICT - REST API Server (Futuro)

**Status**: 📦 **RESERVADO PARA IMPLEMENTAÇÃO FUTURA**

---

## 🎯 Decisão Arquitetural

Esta pasta está **reservada** para um possível servidor REST API no futuro, mas **NÃO será implementado agora**.

### Por que não implementar REST agora?

1. ✅ **gRPC já implementado** em `cmd/grpc/` - Funcional e pronto
2. ✅ **Melhor performance** - gRPC usa HTTP/2 + Protocol Buffers (binary)
3. ✅ **Tipagem forte** - Proto files garantem contratos
4. ✅ **Menos complexidade** - 1 servidor é mais simples que 2
5. ✅ **Padrão moderno** - Microserviços usam gRPC internamente

---

## 🚀 Arquitetura Atual

```
Front-End (React/Next.js)
    ↓
    gRPC-Web (HTTP/2)
    ↓
Core-Dict gRPC Server (:9090)
    ↓
Application Layer (CQRS)
    ↓
PostgreSQL + Redis + Pulsar
```

**Porta**: 9090 (gRPC)
**Protocolo**: gRPC (Protocol Buffers)

---

## 📋 Quando Implementar REST?

### Cenários para adicionar REST:

1. **APIs Públicas**: Parceiros externos precisam integrar via REST
2. **Webhooks**: Notificações HTTP para sistemas legados
3. **Compatibilidade**: Front-End não pode usar gRPC-Web
4. **Documentação**: Swagger/OpenAPI necessário para parceiros

### Como seria a arquitetura:

```
Front-End ou Parceiros
    ↓
    REST/JSON (:8080)
    ↓
REST Gateway (cmd/api)
    ↓ (gRPC interno)
    ↓
Core-Dict gRPC Server (:9090)
```

**Função do REST Gateway**:
- Receber requests HTTP/JSON
- Converter JSON → Proto (gRPC)
- Chamar servidor gRPC interno
- Converter Proto → JSON
- Retornar response HTTP

---

## 🛠️ Tecnologias Previstas

Se implementado no futuro:

- **Framework**: Fiber v3 (Go)
- **Validação**: go-playground/validator
- **Docs**: Swagger/OpenAPI
- **Auth**: JWT middleware
- **Rate Limiting**: Token bucket
- **CORS**: Configurável

### Endpoints REST (mapeamento)

```
POST   /v1/keys                → CreateKey
GET    /v1/keys                → ListKeys
GET    /v1/keys/:id            → GetKey
DELETE /v1/keys/:id            → DeleteKey

POST   /v1/claims              → StartClaim
GET    /v1/claims/:id          → GetClaimStatus
GET    /v1/claims/incoming     → ListIncomingClaims
GET    /v1/claims/outgoing     → ListOutgoingClaims
POST   /v1/claims/:id/respond  → RespondToClaim
DELETE /v1/claims/:id          → CancelClaim

POST   /v1/portability         → StartPortability
POST   /v1/portability/:id/confirm → ConfirmPortability
DELETE /v1/portability/:id     → CancelPortability

GET    /v1/lookup              → LookupKey
GET    /health                 → HealthCheck
```

---

## 📊 Comparação: gRPC vs REST

| Aspecto | gRPC (atual) | REST (futuro) |
|---------|--------------|---------------|
| Performance | ⚡ HTTP/2, Binary | 🐢 HTTP/1.1, JSON |
| Tipagem | ✅ Proto files | ⚠️ Manual |
| Streaming | ✅ Bidirecional | ❌ Não |
| Browser | ⚠️ grpc-web lib | ✅ fetch nativo |
| Debugging | grpcurl | curl, Postman |
| Docs | Proto comments | Swagger/OpenAPI |
| Cache | ⚠️ Difícil | ✅ HTTP cache |

---

## 🚀 Front-End: Como usar gRPC

### 1. Instalar dependências

```bash
npm install grpc-web
npm install google-protobuf
```

### 2. Gerar client TypeScript

```bash
# Usar proto files de dict-contracts/proto/core_dict.proto
protoc --js_out=import_style=commonjs:. \
       --grpc-web_out=import_style=typescript,mode=grpcwebtext:. \
       dict-contracts/proto/core_dict.proto
```

### 3. Usar no código

```typescript
import { CoreDictServiceClient } from './generated/core_dict_grpc_web_pb';
import { CreateKeyRequest } from './generated/core_dict_pb';

const client = new CoreDictServiceClient('http://localhost:9090');

const request = new CreateKeyRequest();
request.setKeyType(KeyType.KEY_TYPE_CPF);
request.setKeyValue('12345678900');
request.setAccountId('acc-123');

client.createKey(request, {}, (err, response) => {
  if (err) {
    console.error('Error:', err);
  } else {
    console.log('Key created:', response.getKeyId());
  }
});
```

---

## 📞 Estimativa de Implementação REST

Se decidir implementar no futuro:

| Tarefa | Tempo |
|--------|-------|
| Setup Fiber v3 server | 30min |
| 15 REST handlers | 3h |
| Middleware (auth, CORS, rate limit) | 1h |
| Swagger/OpenAPI docs | 1h |
| Tests | 1h |
| **TOTAL** | **6-7h** |

---

## ✅ Conclusão

**Decisão**: Não implementar REST agora

**Motivo**:
- gRPC já está pronto e é superior tecnicamente
- Front-End pode usar gRPC-Web
- Menos complexidade = mais rápido

**Futuro**:
- Se necessário, implementar REST gateway aqui
- Estimativa: 6-7h de trabalho
- Não bloqueia Front-End atual

---

**Data Decisão**: 2025-10-27
**Status**: 📦 Reservado (não implementado)
**Prioridade**: Baixa (apenas se necessário)
