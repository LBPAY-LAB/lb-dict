# Core DICT - gRPC Server

Servidor gRPC do Core DICT que expõe 15 RPCs para o Front-End gerenciar chaves PIX.

---

## 🚀 Quick Start (Mock Mode)

### 1. Configurar variáveis de ambiente

```bash
# Copiar exemplo
cp .env.example .env

# Editar .env (opcional, valores padrão já funcionam)
CORE_DICT_USE_MOCK_MODE=true
GRPC_PORT=9090
LOG_LEVEL=info
```

### 2. Rodar servidor

```bash
# A partir da raiz do projeto core-dict
go run cmd/grpc/main.go
```

**Output esperado**:
```json
{"time":"2025-10-27T...","level":"INFO","msg":"Starting Core DICT gRPC Server","port":"9090","mock_mode":true,"version":"1.0.0"}
{"time":"2025-10-27T...","level":"WARN","msg":"⚠️  MOCK MODE ENABLED - Using mock responses for all RPCs"}
{"time":"2025-10-27T...","level":"INFO","msg":"✅ CoreDictService registered (MOCK MODE)"}
{"time":"2025-10-27T...","level":"INFO","msg":"✅ Health Check service registered"}
{"time":"2025-10-27T...","level":"INFO","msg":"✅ gRPC Reflection enabled (for grpcurl)"}
{"time":"2025-10-27T...","level":"INFO","msg":"🚀 gRPC server listening","address":"[::]:9090"}
```

### 3. Testar com grpcurl

```bash
# Health Check
grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check

# Listar serviços disponíveis
grpcurl -plaintext localhost:9090 list

# Listar RPCs do CoreDictService
grpcurl -plaintext localhost:9090 list dict.core.v1.CoreDictService

# Criar chave PIX (mock)
grpcurl -plaintext -d '{
  "key_type": "KEY_TYPE_CPF",
  "key_value": "12345678900",
  "account_id": "acc-123"
}' localhost:9090 dict.core.v1.CoreDictService/CreateKey

# Listar chaves (mock)
grpcurl -plaintext -d '{
  "page_size": 20
}' localhost:9090 dict.core.v1.CoreDictService/ListKeys
```

---

## 📋 RPCs Disponíveis (15 total)

### 1️⃣ Directory (Vínculos DICT) - 4 RPCs
- `CreateKey`: Criar nova chave PIX
- `ListKeys`: Listar chaves do usuário
- `GetKey`: Ver detalhes de uma chave
- `DeleteKey`: Deletar chave

### 2️⃣ Claim (Reivindicação) - 6 RPCs
- `StartClaim`: Iniciar reivindicação
- `GetClaimStatus`: Ver status (com dias restantes)
- `ListIncomingClaims`: Claims recebidas
- `ListOutgoingClaims`: Claims enviadas
- `RespondToClaim`: Aceitar/Rejeitar claim
- `CancelClaim`: Cancelar claim

### 3️⃣ Portability (Portabilidade) - 3 RPCs
- `StartPortability`: Iniciar portabilidade
- `ConfirmPortability`: Confirmar portabilidade
- `CancelPortability`: Cancelar portabilidade

### 4️⃣ Directory Queries (Consultas) - 1 RPC
- `LookupKey`: Consultar chave de terceiro

### Health Check - 1 RPC
- `HealthCheck`: Status do serviço

**Documentação completa**: Ver `Artefatos/00_Master/VALIDACAO_INTERFACE_GRPC_FRONTEND.md`

---

## 🔄 Modes: Mock vs Real

### Mock Mode (Default) ✅
**Uso**: Front-End pode começar a integrar **HOJE**

**Configuração**:
```bash
CORE_DICT_USE_MOCK_MODE=true
```

**Comportamento**:
- ✅ Todas as validações funcionam (campos required, tipos, etc.)
- ✅ Retorna respostas mock realistas
- ✅ Não precisa de PostgreSQL, Redis, Pulsar
- ✅ Logs detalhados de cada RPC
- ⚠️ Não persiste dados (memória volátil)

**Quando usar**:
- Desenvolvimento Front-End
- Testes de integração iniciais
- Demos rápidas

---

### Real Mode (Em desenvolvimento) 🚧
**Uso**: Backend completo com lógica de negócio Bacen

**Configuração**:
```bash
CORE_DICT_USE_MOCK_MODE=false

# Configurar também:
DB_HOST=localhost
DB_PORT=5432
REDIS_HOST=localhost
REDIS_PORT=6379
PULSAR_URL=pulsar://localhost:6650
CONNECT_GRPC_URL=localhost:9092
```

**Comportamento**:
- ✅ Lógica de negócio completa (validações Bacen)
- ✅ Persistência PostgreSQL
- ✅ Cache Redis
- ✅ Eventos Pulsar
- ✅ Comunicação com RSFN via Connect
- ✅ Limites (5 CPF, 20 CNPJ)
- ✅ OTP para Email/Phone
- ✅ Claims de 30 dias

**Status**: ⏳ Em implementação (precisa ajustar mappers)

---

## 🐳 Docker (Futuro)

```bash
# Build
docker build -t core-dict-grpc -f cmd/grpc/Dockerfile .

# Run (mock mode)
docker run -p 9090:9090 -e CORE_DICT_USE_MOCK_MODE=true core-dict-grpc

# Run (real mode - com docker-compose para dependências)
docker-compose up core-dict-grpc
```

---

## 📊 Monitoring

### Logs
Logs estruturados em JSON (stdout):

```json
{
  "time": "2025-10-27T14:30:00Z",
  "level": "INFO",
  "msg": "gRPC request completed",
  "method": "/dict.core.v1.CoreDictService/CreateKey",
  "duration_ms": 5
}
```

### Metrics (Futuro)
Prometheus metrics em `:9091/metrics`:
- `grpc_requests_total`
- `grpc_request_duration_seconds`
- `grpc_errors_total`

### Health Check
```bash
# Via gRPC
grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check

# Response:
{
  "status": "SERVING"
}
```

---

## 🔧 Development

### Rebuild proto files
```bash
# Se modificou proto files em dict-contracts
cd ../dict-contracts
make generate

# Core-dict vai pegar as mudanças automaticamente via go.mod replace
```

### Hot reload (com air)
```bash
# Instalar air
go install github.com/cosmtrek/air@latest

# Rodar com hot reload
air -c cmd/grpc/.air.toml
```

---

## 🐛 Troubleshooting

### Erro: "bind: address already in use"
```bash
# Porta 9090 já está em uso
lsof -ti:9090 | xargs kill -9

# Ou mudar porta
GRPC_PORT=9091 go run cmd/grpc/main.go
```

### Erro: "package github.com/lbpay-lab/dict-contracts not found"
```bash
# Rebuild contracts
cd ../dict-contracts
make generate

# Atualizar deps
cd ../core-dict
go mod tidy
```

### Mock responses não realistas
Mock mode retorna dados fixos. Para testar dados realistas, use real mode.

---

## 📞 Próximos Passos

1. ✅ **Mock mode funcional** - Front-End pode começar
2. ⏳ **Ajustar mappers** - Alinhar com Commands/Queries
3. ⏳ **Implementar real mode** - Conectar com Application Layer
4. ⏳ **Docker compose** - PostgreSQL + Redis + Pulsar
5. ⏳ **Auth interceptor** - JWT validation
6. ⏳ **Rate limiting** - Token bucket
7. ⏳ **Metrics** - Prometheus
8. ⏳ **Tracing** - Jaeger

---

**Status**: ✅ **SERVIDOR FUNCIONAL EM MOCK MODE**
**Data**: 2025-10-27
**Versão**: 1.0.0
