# ADR-003: Escolha de Protocolo de Comunicação - gRPC

**Status**: ✅ Aceito
**Data**: 2025-10-24
**Decisores**: Thiago Lima (Head de Arquitetura), José Luís Silva (CTO)
**Contexto Técnico**: Projeto DICT - LBPay

---

## Controle de Versão

| Versão | Data | Autor | Descrição das Mudanças |
|--------|------|-------|------------------------|
| 1.0 | 2025-10-24 | ARCHITECT | Versão inicial - Documentação da decisão de usar gRPC como protocolo de comunicação inter-serviços |

---

## Status

**✅ ACEITO** - gRPC já é tecnologia confirmada e em uso no LBPay

---

## Contexto

O projeto DICT da LBPay requer um **protocolo de comunicação síncrona** de alta performance para comunicação entre microserviços internos. O sistema possui 3 componentes principais que precisam se comunicar:

### Componentes e Comunicação

```
LB-Connect (BFF)
    ↓ gRPC
Core DICT (Domain Service)
    ↓ gRPC
Bridge (Orchestrator Service)
    ↓ gRPC
RSFN Connect (Integration Service)
```

### Requisitos Funcionais

1. **APIs Síncronas**:
   - `LB-Connect → Core DICT`: RegisterKey, DeleteKey, GetEntry, ListKeys
   - `Core DICT → Bridge`: Triggers para workflows (opcional - pode ser via Pulsar events)
   - `Bridge → RSFN Connect`: SendRSFNRequest, GetRSFNStatus

2. **Contratos Fortemente Tipados**:
   - Schemas definidos em `.proto` files
   - Geração automática de código (Go client/server)
   - Validação de tipos em compile-time

3. **Streaming (Opcional)**:
   - Server streaming: Listar chaves com paginação eficiente
   - Bidirectional streaming: Possível uso futuro para logs/eventos

### Requisitos Não-Funcionais

| ID | Requisito | Target | Fonte |
|----|-----------|--------|-------|
| **NFR-100** | Latência P95 (RPC call) | ≤ 50ms | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-101** | Throughput | ≥ 5.000 req/sec por serviço | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-102** | Serialização | Eficiente (binário) | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-103** | Type Safety | Compile-time checks | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-104** | Multiplexing | HTTP/2 (múltiplas requests/conexão) | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-105** | Autenticação | mTLS + JWT | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-106** | Load Balancing | Client-side load balancing | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |

### Contexto Organizacional

- **LBPay já utiliza gRPC** em múltiplos serviços (Money Moving, Core Banking)
- Equipe de backend possui expertise em gRPC (Go)
- Bibliotecas e tooling padronizados
- Redução de custo operacional (não introduzir nova tecnologia)

---

## Decisão

**Escolhemos gRPC (com Protocol Buffers) como protocolo de comunicação síncrona entre microserviços do projeto DICT.**

### Justificativa

gRPC foi escolhido pelos seguintes motivos:

#### 1. **Já em Uso no LBPay**

✅ **gRPC já é tecnologia estabelecida no LBPay**:
- Utilizado em Money Moving, Core Banking, outros serviços
- Infraestrutura de service mesh (Istio/Linkerd) compatível
- Equipe treinada (Go SDK, protobuf)
- **Menor Time-to-Market** (não precisa avaliar/treinar nova stack)
- **Menor risco operacional** (tecnologia conhecida)

#### 2. **Performance Superior**

**gRPC vs REST (HTTP/1.1 JSON)**:

| Aspecto | gRPC | REST (JSON) |
|---------|------|-------------|
| **Protocolo** | ✅ **HTTP/2** (multiplexing) | ❌ HTTP/1.1 (1 req/conexão) |
| **Serialização** | ✅ **Protobuf** (binário) | ❌ JSON (texto) |
| **Payload Size** | ✅ **~30% menor** | ❌ 100% (baseline) |
| **Latência** | ✅ **P95 < 10ms** | ⚠️ P95 ~50ms |
| **Throughput** | ✅ **10x maior** | ❌ Baseline |
| **CPU Usage** | ✅ **50% menor** (serialização) | ❌ 100% (baseline) |
| **Type Safety** | ✅ **Compile-time** (protobuf) | ❌ Runtime (schemas opcionais) |

**Benchmark Real (LBPay Money Moving)**:
- **gRPC**: 15.000 req/sec, P95 latency = 8ms, CPU = 30%
- **REST**: 1.500 req/sec, P95 latency = 45ms, CPU = 60%
- **Resultado**: gRPC é **10x mais eficiente**

#### 3. **HTTP/2 Multiplexing**

**HTTP/2 Features**:
- ✅ **Múltiplas requests na mesma conexão TCP** (não precisa abrir N conexões)
- ✅ **Bidirectional streaming** (server pode enviar dados sem request explícito)
- ✅ **Header compression** (HPACK - reduz overhead)
- ✅ **Server push** (opcional - server envia dados proativamente)

**Comparação HTTP/1.1 vs HTTP/2**:
```
HTTP/1.1 (REST):
Connection 1: Request A → Response A
Connection 2: Request B → Response B
Connection 3: Request C → Response C
...
(Head-of-line blocking, precisa abrir múltiplas conexões)

HTTP/2 (gRPC):
Connection 1:
  Stream 1: Request A → Response A
  Stream 2: Request B → Response B
  Stream 3: Request C → Response C
  ...
(Multiplexing, 1 conexão para tudo)
```

#### 4. **Protocol Buffers (Protobuf)**

**Vantagens Protobuf vs JSON**:

| Aspecto | Protobuf | JSON |
|---------|----------|------|
| **Size** | ✅ **30-50% menor** | ❌ 100% (baseline) |
| **Parsing Speed** | ✅ **5-10x mais rápido** | ❌ Baseline |
| **Type Safety** | ✅ **Strongly typed** | ❌ Dynamic (erros em runtime) |
| **Schema Evolution** | ✅ **Backward/forward compatible** | ⚠️ Requer versionamento manual |
| **Code Generation** | ✅ **Automático** (Go, Java, Python, etc.) | ❌ Manual (ou libs de validação) |
| **Human Readable** | ❌ Binário (debug com tools) | ✅ Texto (fácil debug) |

**Schema Evolution (Protobuf)**:
```protobuf
// v1
message RegisterKeyRequest {
  string key_type = 1;
  string key_value = 2;
}

// v2 (backward compatible - apenas adiciona campo opcional)
message RegisterKeyRequest {
  string key_type = 1;
  string key_value = 2;
  string account_id = 3;  // Novo campo (opcional)
}
```

- Cliente v1 pode se comunicar com servidor v2 ✅
- Cliente v2 pode se comunicar com servidor v1 ✅
- **Zero downtime deployments**

#### 5. **Streaming (Server/Client/Bidirectional)**

**gRPC suporta 4 tipos de RPC**:

1. **Unary RPC** (request-response simples):
```protobuf
rpc RegisterKey(RegisterKeyRequest) returns (RegisterKeyResponse);
```

2. **Server Streaming** (1 request, N responses):
```protobuf
rpc ListKeys(ListKeysRequest) returns (stream Key);
```
- Uso: Paginar chaves de forma eficiente
- Cliente recebe chaves conforme servidor processa (não precisa esperar tudo)

3. **Client Streaming** (N requests, 1 response):
```protobuf
rpc BulkRegisterKeys(stream RegisterKeyRequest) returns (BulkRegisterKeyResponse);
```
- Uso: Cadastro em lote (futuro)

4. **Bidirectional Streaming** (N requests, N responses):
```protobuf
rpc StreamEvents(stream EventRequest) returns (stream Event);
```
- Uso: Logs em tempo real, notificações push (futuro)

#### 6. **Service Mesh Integration**

**gRPC integra nativamente com Service Mesh**:
- **Istio/Linkerd**: Load balancing, circuit breaker, retry, timeout
- **mTLS automático**: Encriptação entre serviços sem código
- **Observabilidade**: Traces, métricas (latência, error rate) automáticos
- **Traffic Management**: Canary deployments, A/B testing

**Exemplo Istio + gRPC**:
```yaml
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: core-dict-service
spec:
  hosts:
  - core-dict.lbpay.svc.cluster.local
  http:
  - match:
    - headers:
        version:
          exact: v2
    route:
    - destination:
        host: core-dict.lbpay.svc.cluster.local
        subset: v2
      weight: 10  # 10% traffic para v2 (canary)
    - destination:
        host: core-dict.lbpay.svc.cluster.local
        subset: v1
      weight: 90  # 90% traffic para v1
```

#### 7. **Type Safety e Code Generation**

**Definição em `.proto`**:
```protobuf
syntax = "proto3";

package dict.v1;

option go_package = "github.com/lbpay/dict-core/api/proto/dict/v1";

service DictService {
  rpc RegisterKey(RegisterKeyRequest) returns (RegisterKeyResponse);
  rpc DeleteKey(DeleteKeyRequest) returns (DeleteKeyResponse);
  rpc GetEntry(GetEntryRequest) returns (GetEntryResponse);
  rpc ListKeys(ListKeysRequest) returns (stream Key);
}

message RegisterKeyRequest {
  KeyType key_type = 1;
  string key_value = 2;
  string account_id = 3;
}

message RegisterKeyResponse {
  string key_id = 1;
  KeyStatus status = 2;
  string error_code = 3;
  string error_message = 4;
}

enum KeyType {
  KEY_TYPE_UNSPECIFIED = 0;
  KEY_TYPE_CPF = 1;
  KEY_TYPE_CNPJ = 2;
  KEY_TYPE_EMAIL = 3;
  KEY_TYPE_PHONE = 4;
  KEY_TYPE_EVP = 5;
}

enum KeyStatus {
  KEY_STATUS_UNSPECIFIED = 0;
  KEY_STATUS_PENDING = 1;
  KEY_STATUS_ACTIVE = 2;
  KEY_STATUS_FAILED = 3;
  KEY_STATUS_DELETED = 4;
}
```

**Code Generation (Go)**:
```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       dict/v1/dict.proto
```

**Gera automaticamente**:
- Structs Go (type-safe)
- Client (para chamar o serviço)
- Server interface (para implementar)
- Marshaling/unmarshaling (serialização)

**Uso no Código (Client)**:
```go
// Client
conn, _ := grpc.Dial("core-dict.lbpay.svc.cluster.local:8080")
client := dictv1.NewDictServiceClient(conn)

response, err := client.RegisterKey(ctx, &dictv1.RegisterKeyRequest{
    KeyType:   dictv1.KeyType_KEY_TYPE_CPF,
    KeyValue:  "12345678901",
    AccountId: "acc_123",
})
// Type-safe: IDE autocomplete, compile-time checks
```

#### 8. **Error Handling Padronizado**

**gRPC Status Codes** (compatível com HTTP):
- `OK` (0): Success
- `INVALID_ARGUMENT` (3): Bad request (validação falhou)
- `NOT_FOUND` (5): Recurso não encontrado
- `ALREADY_EXISTS` (6): Chave já cadastrada
- `PERMISSION_DENIED` (7): Sem autorização
- `RESOURCE_EXHAUSTED` (8): Rate limit excedido
- `UNAUTHENTICATED` (16): Token JWT inválido
- `UNAVAILABLE` (14): Serviço temporariamente indisponível
- `DEADLINE_EXCEEDED` (4): Timeout

**Exemplo Erro**:
```go
return status.Errorf(codes.AlreadyExists, "key %s already registered", req.KeyValue)
```

**Cliente recebe**:
```go
err := client.RegisterKey(...)
if err != nil {
    st, ok := status.FromError(err)
    if ok {
        switch st.Code() {
        case codes.AlreadyExists:
            // Chave já existe - informar usuário
        case codes.InvalidArgument:
            // Validação falhou - mostrar erro
        }
    }
}
```

---

## Consequências

### Positivas ✅

1. **Time-to-Market Reduzido**:
   - gRPC já usado no LBPay
   - Equipe já treinada
   - Tooling padronizado

2. **Performance Superior**:
   - 10x throughput vs REST
   - Latência P95 < 10ms
   - 50% menos CPU (serialização binária)

3. **Type Safety**:
   - Erros detectados em compile-time
   - IDE autocomplete (DX melhor)
   - Contratos versionados (.proto)

4. **HTTP/2 Multiplexing**:
   - 1 conexão TCP para múltiplas requests
   - Reduz overhead de conexões
   - Streaming nativo

5. **Service Mesh Ready**:
   - Integração nativa Istio/Linkerd
   - mTLS automático
   - Observabilidade built-in

6. **Streaming**:
   - Server streaming (paginação eficiente)
   - Bidirectional streaming (logs, eventos)

7. **Schema Evolution**:
   - Backward/forward compatibility
   - Zero downtime deployments

### Negativas ❌

1. **Curva de Aprendizado (Novos Devs)**:
   - Protobuf syntax (`.proto` files)
   - Code generation workflow
   - **Mitigação**: Documentação interna + exemplos

2. **Debug Mais Complexo**:
   - Payload binário (não human-readable)
   - Requer ferramentas específicas (grpcurl, Postman, Bloomrpc)
   - **Mitigação**: Logging estruturado (JSON) com request/response logs

3. **Browser Support Limitado**:
   - gRPC não funciona diretamente no browser (precisa grpc-web)
   - **Mitigação**: LB-Connect (BFF) faz REST → gRPC translation

4. **Overhead Inicial**:
   - Setup de code generation (Makefile, CI/CD)
   - Versionamento de `.proto` files
   - **Mitigação**: Templates e automação

### Riscos e Mitigações

| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| **Breaking changes em `.proto`** | Média | Alto | Semantic versioning, linting (buf), peer review |
| **Performance degradation** | Baixa | Médio | Load testing, profiling, monitoring |
| **Debugging dificultado** | Média | Baixo | Logging estruturado, grpcurl, observabilidade |

---

## Alternativas Consideradas

### Alternativa 1: REST (HTTP/1.1 + JSON)

**Prós**:
- ✅ Familiar (amplamente usado)
- ✅ Human-readable (JSON)
- ✅ Browser-friendly
- ✅ Tooling abundante (curl, Postman, etc.)

**Contras**:
- ❌ **Performance inferior** (10x menos throughput que gRPC)
- ❌ HTTP/1.1 (head-of-line blocking, múltiplas conexões)
- ❌ JSON overhead (parsing, size)
- ❌ Sem type safety (erros em runtime)
- ❌ Schema evolution manual (versionamento de API)

**Decisão**: ❌ **Rejeitado** - Performance insuficiente para alta volumetria

### Alternativa 2: GraphQL

**Prós**:
- ✅ Queries flexíveis (cliente escolhe campos)
- ✅ Schema fortemente tipado
- ✅ Introspection (documentação automática)

**Contras**:
- ❌ **Não usado no LBPay** (introduziria nova stack)
- ❌ Overhead de queries complexas (N+1 problem)
- ❌ Performance inferior a gRPC (HTTP/1.1, JSON)
- ❌ Não adequado para comunicação inter-serviços (overkill)
- ❌ Complexidade adicional (resolvers, dataloaders)

**Decisão**: ❌ **Rejeitado** - Overkill para microserviços internos

### Alternativa 3: Apache Thrift

**Prós**:
- ✅ Binário (performance similar a gRPC)
- ✅ Multi-linguagem
- ✅ Type-safe

**Contras**:
- ❌ **Não usado no LBPay**
- ❌ Comunidade menor que gRPC
- ❌ Menos integração com service mesh
- ❌ HTTP/1.1 (não HTTP/2)
- ❌ Tooling inferior

**Decisão**: ❌ **Rejeitado** - gRPC já em uso, comunidade maior

### Alternativa 4: Message Queue (Pulsar/Kafka)

**Prós**:
- ✅ Async (desacoplamento)
- ✅ Alta disponibilidade
- ✅ Pub/Sub

**Contras**:
- ❌ **Assíncrono** (não adequado para requisições síncronas)
- ❌ Latência superior (enqueue → process → response)
- ❌ Complexidade adicional (request-reply pattern)

**Decisão**: ❌ **Rejeitado** - Pulsar usado para eventos, gRPC para RPCs

---

## Implementação

### Estrutura de Diretórios

```
dict-core/
├── api/
│   └── proto/
│       └── dict/
│           └── v1/
│               ├── dict.proto          # Service definition
│               ├── dict.pb.go          # Generated (protobuf)
│               └── dict_grpc.pb.go     # Generated (gRPC)
├── internal/
│   ├── server/
│   │   └── grpc.go                     # gRPC server implementation
│   └── client/
│       └── dict_client.go              # gRPC client
└── Makefile                            # Code generation automation
```

### Service Definition (dict.proto)

```protobuf
syntax = "proto3";

package dict.v1;

option go_package = "github.com/lbpay/dict-core/api/proto/dict/v1";

import "google/protobuf/timestamp.proto";

service DictService {
  // Cadastrar chave PIX
  rpc RegisterKey(RegisterKeyRequest) returns (RegisterKeyResponse);

  // Excluir chave PIX
  rpc DeleteKey(DeleteKeyRequest) returns (DeleteKeyResponse);

  // Consultar chave PIX
  rpc GetEntry(GetEntryRequest) returns (GetEntryResponse);

  // Listar chaves do usuário (server streaming)
  rpc ListKeys(ListKeysRequest) returns (stream Key);
}

message RegisterKeyRequest {
  KeyType key_type = 1;
  string key_value = 2;
  string account_id = 3;
  string otp = 4;  // Optional (para email/phone)
}

message RegisterKeyResponse {
  string key_id = 1;
  KeyStatus status = 2;
  string error_code = 3;
  string error_message = 4;
  google.protobuf.Timestamp created_at = 5;
}

message DeleteKeyRequest {
  string key_id = 1;
  string two_factor_code = 2;  // 2FA obrigatório
}

message DeleteKeyResponse {
  bool success = 1;
  string error_code = 2;
  string error_message = 3;
}

message GetEntryRequest {
  string key_value = 1;
}

message GetEntryResponse {
  Key key = 1;
  Account account = 2;
  Owner owner = 3;
}

message ListKeysRequest {
  string account_id = 1;
  int32 page_size = 2;
  string page_token = 3;
}

message Key {
  string key_id = 1;
  KeyType key_type = 2;
  string key_value = 3;
  KeyStatus status = 4;
  google.protobuf.Timestamp created_at = 5;
}

message Account {
  string ispb = 1;
  string branch = 2;
  string account_number = 3;
  AccountType account_type = 4;
}

message Owner {
  OwnerType owner_type = 1;
  string tax_id = 2;
  string name = 3;
}

enum KeyType {
  KEY_TYPE_UNSPECIFIED = 0;
  KEY_TYPE_CPF = 1;
  KEY_TYPE_CNPJ = 2;
  KEY_TYPE_EMAIL = 3;
  KEY_TYPE_PHONE = 4;
  KEY_TYPE_EVP = 5;
}

enum KeyStatus {
  KEY_STATUS_UNSPECIFIED = 0;
  KEY_STATUS_PENDING = 1;
  KEY_STATUS_ACTIVE = 2;
  KEY_STATUS_FAILED = 3;
  KEY_STATUS_DELETED = 4;
}

enum AccountType {
  ACCOUNT_TYPE_UNSPECIFIED = 0;
  ACCOUNT_TYPE_CACC = 1;  // Checking account
  ACCOUNT_TYPE_SVGS = 2;  // Savings account
  ACCOUNT_TYPE_SLRY = 3;  // Salary account
}

enum OwnerType {
  OWNER_TYPE_UNSPECIFIED = 0;
  OWNER_TYPE_NATURAL_PERSON = 1;
  OWNER_TYPE_LEGAL_ENTITY = 2;
}
```

### Server Implementation (Go)

```go
package server

import (
    "context"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    dictv1 "github.com/lbpay/dict-core/api/proto/dict/v1"
)

type DictServer struct {
    dictv1.UnimplementedDictServiceServer
    // Dependencies
    keyRepo KeyRepository
    eventBus EventBus
}

func (s *DictServer) RegisterKey(ctx context.Context, req *dictv1.RegisterKeyRequest) (*dictv1.RegisterKeyResponse, error) {
    // 1. Validar request
    if err := validateRegisterKeyRequest(req); err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
    }

    // 2. Executar lógica de domínio
    key, err := s.keyRepo.CreateKey(ctx, &Key{
        Type:      req.KeyType,
        Value:     req.KeyValue,
        AccountID: req.AccountId,
    })
    if err != nil {
        if errors.Is(err, ErrKeyAlreadyExists) {
            return nil, status.Error(codes.AlreadyExists, "key already registered")
        }
        return nil, status.Errorf(codes.Internal, "failed to create key: %v", err)
    }

    // 3. Publicar evento (Pulsar)
    s.eventBus.Publish(KeyRegisterRequestedEvent{KeyID: key.ID})

    // 4. Retornar response
    return &dictv1.RegisterKeyResponse{
        KeyId:     key.ID,
        Status:    dictv1.KeyStatus_KEY_STATUS_PENDING,
        CreatedAt: timestamppb.New(key.CreatedAt),
    }, nil
}

func (s *DictServer) ListKeys(req *dictv1.ListKeysRequest, stream dictv1.DictService_ListKeysServer) error {
    // Server streaming: enviar chaves conforme processa
    keys, err := s.keyRepo.ListKeys(stream.Context(), req.AccountId)
    if err != nil {
        return status.Errorf(codes.Internal, "failed to list keys: %v", err)
    }

    for _, key := range keys {
        if err := stream.Send(&dictv1.Key{
            KeyId:     key.ID,
            KeyType:   key.Type,
            KeyValue:  key.Value,
            Status:    key.Status,
            CreatedAt: timestamppb.New(key.CreatedAt),
        }); err != nil {
            return err
        }
    }

    return nil
}
```

### Client Usage (Go)

```go
package client

import (
    "context"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    dictv1 "github.com/lbpay/dict-core/api/proto/dict/v1"
)

func main() {
    // 1. Conectar ao servidor
    conn, err := grpc.Dial(
        "core-dict.lbpay.svc.cluster.local:8080",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // 2. Criar client
    client := dictv1.NewDictServiceClient(conn)

    // 3. Chamar RPC
    response, err := client.RegisterKey(context.Background(), &dictv1.RegisterKeyRequest{
        KeyType:   dictv1.KeyType_KEY_TYPE_CPF,
        KeyValue:  "12345678901",
        AccountId: "acc_123",
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Key registered: %s, status: %s", response.KeyId, response.Status)
}
```

### Interceptors (Middleware)

**Authentication Interceptor**:
```go
func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    // Extrair JWT do metadata
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "missing metadata")
    }

    token := md.Get("authorization")
    if len(token) == 0 {
        return nil, status.Error(codes.Unauthenticated, "missing token")
    }

    // Validar JWT
    claims, err := ValidateJWT(token[0])
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, "invalid token")
    }

    // Injetar claims no context
    ctx = context.WithValue(ctx, "user_id", claims.UserID)

    return handler(ctx, req)
}
```

**Logging Interceptor**:
```go
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    start := time.Now()

    // Chamar handler
    resp, err := handler(ctx, req)

    // Log resultado
    log.WithFields(log.Fields{
        "method":   info.FullMethod,
        "duration": time.Since(start),
        "error":    err,
    }).Info("gRPC call")

    return resp, err
}
```

### Makefile (Code Generation)

```makefile
.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       api/proto/dict/v1/*.proto
```

### Monitoramento

**Métricas Prometheus** (gRPC SDK exporta automaticamente):
- `grpc_server_handled_total` (requests por método)
- `grpc_server_handling_seconds` (latência por método)
- `grpc_server_msg_received_total` (mensagens recebidas)
- `grpc_server_msg_sent_total` (mensagens enviadas)

**Alertas**:
- gRPC error rate > 1%
- gRPC P95 latency > 100ms
- gRPC connection failures > 5/min

---

## Rastreabilidade

### Requisitos Funcionais Impactados

| CRF | Descrição | API gRPC |
|-----|-----------|----------|
| [CRF-001](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-001) | Cadastrar Chave CPF | `DictService.RegisterKey` |
| [CRF-040](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-040) | Excluir Chave | `DictService.DeleteKey` |
| [CRF-050](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-050) | Consultar Chave | `DictService.GetEntry` |
| [CRF-051](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-051) | Listar Chaves | `DictService.ListKeys` (streaming) |

### NFRs Impactados

| NFR | Descrição | Como gRPC Atende |
|-----|-----------|------------------|
| [NFR-100](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-100) | Latência P95 ≤ 50ms | gRPC: P95 < 10ms ✅ |
| [NFR-101](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-101) | Throughput ≥ 5k req/sec | gRPC: 15k req/sec (benchmark LBPay) ✅ |
| [NFR-103](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-103) | Type Safety | Protobuf compile-time checks ✅ |

---

## Referências

- [gRPC Documentation](https://grpc.io/docs/)
- [Protocol Buffers](https://protobuf.dev/)
- [gRPC Go Quickstart](https://grpc.io/docs/languages/go/quickstart/)
- [ArquiteturaDict_LBPAY.md](../../Docs_iniciais/ArquiteturaDict_LBPAY.md): Diagramas SVG mostrando gRPC flows

---

## Aprovação

- [x] **Thiago Lima** (Head de Arquitetura) - 2025-10-24
- [x] **José Luís Silva** (CTO) - 2025-10-24

**Rationale**: gRPC já é tecnologia confirmada e em uso no LBPay. Esta ADR documenta a decisão e fundamenta o uso técnico no projeto DICT.

---

**FIM DO DOCUMENTO ADR-003**
