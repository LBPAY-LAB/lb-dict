# DICT Contracts - Protocol Buffers

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Versão**: v0.1.0
**Status**: Initial Release
**Data**: 2025-10-26

---

## Visão Geral

Este repositório contém as definições de **Protocol Buffers (proto3)** para todos os serviços gRPC do sistema DICT. Os contratos aqui definidos garantem a comunicação consistente entre os componentes do sistema.

### Componentes do Sistema DICT

```
[FrontEnd Web/Mobile]
        │
        │ gRPC (CoreDictService)
        ▼
[Core DICT API]
        │
        │ gRPC (interno)
        ▼
[RSFN Connect]
        │
        │ gRPC (BridgeService)
        ▼
[Bridge DICT]
        │
        │ SOAP/mTLS
        ▼
[Bacen DICT]
```

---

## Estrutura do Repositório

```
dict-contracts/
├── proto/
│   ├── common.proto           # Tipos compartilhados (Account, DictKey, Status, Error)
│   ├── core_dict.proto        # CoreDictService (FrontEnd ↔ Core DICT)
│   └── bridge.proto           # BridgeService (Connect ↔ Bridge)
├── gen/                       # Código gerado (gitignored)
├── buf/                       # Configuração Buf
├── buf.yaml                   # Configuração de linting e breaking changes
└── README.md                  # Este arquivo
```

---

## Arquivos Proto

### 1. `proto/common.proto`

Tipos compartilhados entre todos os serviços:

**Enums:**
- `KeyType`: Tipos de chave PIX (CPF, CNPJ, EMAIL, PHONE, EVP)
- `AccountType`: Tipos de conta bancária (CHECKING, SAVINGS, PAYMENT, SALARY)
- `EntryStatus`: Status de uma chave DICT
- `ClaimStatus`: Status de reivindicação (30 dias)
- `DocumentType`: Tipo de documento (CPF ou CNPJ)

**Messages:**
- `Account`: Dados de conta bancária (ISPB, agência, conta, titular)
- `DictKey`: Chave DICT (tipo + valor)
- `ValidationError`: Erros de validação de campos
- `BusinessError`: Erros de regra de negócio
- `InfrastructureError`: Erros de infraestrutura
- `BacenError`: Erros específicos do Bacen
- `ErrorResponse`: Wrapper para todos os tipos de erro

---

### 2. `proto/core_dict.proto`

Serviço **CoreDictService** (FrontEnd → Core DICT):

**RPCs:**

**Operações de Chave:**
- `CreateKey`: Criar nova chave PIX
- `ListKeys`: Listar chaves do usuário
- `GetKey`: Obter detalhes de uma chave
- `DeleteKey`: Deletar chave PIX

**Operações de Claim (30 dias):**
- `StartClaim`: Iniciar reivindicação de chave
- `GetClaimStatus`: Verificar status de claim
- `ListIncomingClaims`: Listar claims recebidas
- `ListOutgoingClaims`: Listar claims enviadas
- `RespondToClaim`: Aceitar ou rejeitar claim
- `CancelClaim`: Cancelar claim

**Operações de Portabilidade:**
- `StartPortability`: Iniciar portabilidade de conta
- `ConfirmPortability`: Confirmar portabilidade
- `CancelPortability`: Cancelar portabilidade

**Consultas:**
- `LookupKey`: Consultar chave DICT de terceiros

**Health Check:**
- `HealthCheck`: Verificar saúde do serviço

---

### 3. `proto/bridge.proto`

Serviço **BridgeService** (Connect → Bridge → Bacen):

**RPCs:**

**Operações de Entry:**
- `CreateEntry`: Criar chave no Bacen
- `GetEntry`: Buscar chave no Bacen
- `DeleteEntry`: Deletar chave no Bacen
- `UpdateEntry`: Atualizar dados da conta

**Operações de Claim (30 dias):**
- `CreateClaim`: Criar claim no Bacen
- `GetClaim`: Buscar status de claim
- `CompleteClaim`: Completar claim (aprovação)
- `CancelClaim`: Cancelar claim

**Operações de Portabilidade:**
- `InitiatePortability`: Iniciar portabilidade
- `ConfirmPortability`: Confirmar portabilidade
- `CancelPortability`: Cancelar portabilidade

**Consultas de Diretório:**
- `GetDirectory`: Consultar diretório completo
- `SearchEntries`: Buscar chaves por critérios

**Health Check:**
- `HealthCheck`: Verificar conectividade com Bacen

---

## Geração de Código

### Pré-requisitos

```bash
# Instalar protoc
brew install protobuf  # macOS
# ou
apt-get install protobuf-compiler  # Linux

# Instalar plugins Go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Adicionar ao PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Gerar Código Go

```bash
# Executar na raiz do repositório
make proto-gen

# Ou manualmente:
protoc \
  --go_out=gen \
  --go_opt=paths=source_relative \
  --go-grpc_out=gen \
  --go-grpc_opt=paths=source_relative \
  --proto_path=proto \
  proto/*.proto
```

### Saída Esperada

```
gen/
├── common.pb.go
├── core_dict.pb.go
├── core_dict_grpc.pb.go
├── bridge.pb.go
└── bridge_grpc.pb.go
```

---

## Validação com Buf

### Instalar Buf

```bash
# macOS
brew install bufbuild/buf/buf

# Linux
BIN="/usr/local/bin" && \
VERSION="1.28.1" && \
curl -sSL \
  "https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m)" \
  -o "${BIN}/buf" && \
chmod +x "${BIN}/buf"
```

### Executar Linting

```bash
# Lint dos arquivos proto
buf lint

# Breaking change detection
buf breaking --against '.git#branch=main'
```

---

## Uso em Projetos Go

### Importar Contratos

```go
import (
    commonv1 "github.com/lbpay/dict-contracts/proto/common/v1"
    corev1 "github.com/lbpay/dict-contracts/proto/core/v1"
    bridgev1 "github.com/lbpay/dict-contracts/proto/bridge/v1"
)
```

### Exemplo: Cliente gRPC

```go
// Conectar ao Core DICT
conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

client := corev1.NewCoreDictServiceClient(conn)

// Criar chave PIX
resp, err := client.CreateKey(context.Background(), &corev1.CreateKeyRequest{
    KeyType:   commonv1.KeyType_KEY_TYPE_CPF,
    KeyValue:  "12345678900",
    AccountId: "acc-550e8400",
})
```

### Exemplo: Servidor gRPC

```go
// Implementar serviço
type coreDictServer struct {
    corev1.UnimplementedCoreDictServiceServer
}

func (s *coreDictServer) CreateKey(
    ctx context.Context,
    req *corev1.CreateKeyRequest,
) (*corev1.CreateKeyResponse, error) {
    // Implementação...
    return &corev1.CreateKeyResponse{
        KeyId:  "key-550e8400",
        Status: commonv1.EntryStatus_ENTRY_STATUS_ACTIVE,
    }, nil
}

// Registrar servidor
lis, _ := net.Listen("tcp", ":50051")
grpcServer := grpc.NewServer()
corev1.RegisterCoreDictServiceServer(grpcServer, &coreDictServer{})
grpcServer.Serve(lis)
```

---

## Convenções e Boas Práticas

### Nomenclatura

- **Packages**: `dict.{service}.v1` (ex: `dict.bridge.v1`)
- **Services**: `{Service}Service` (ex: `BridgeService`)
- **Messages**: PascalCase (ex: `CreateKeyRequest`)
- **Fields**: snake_case (ex: `account_holder_name`)
- **Enums**: UPPER_SNAKE_CASE (ex: `KEY_TYPE_CPF`)

### Versionamento

- Sempre incluir versão no package (`v1`, `v2`, etc.)
- Breaking changes requerem nova versão
- Manter retrocompatibilidade quando possível

### Idempotência

Todas as operações de escrita incluem:
- `idempotency_key`: Para retry safety
- `request_id`: Para rastreamento de logs

### Error Handling

Usar `ErrorResponse` para erros estruturados:

```protobuf
message ErrorResponse {
  int32 grpc_code = 1;
  string message = 2;
  oneof details {
    ValidationError validation = 3;
    BusinessError business = 4;
    InfrastructureError infrastructure = 5;
    BacenError bacen = 6;
  }
  string request_id = 7;
  google.protobuf.Timestamp timestamp = 8;
}
```

---

## Documentação de Referência

### Documentos Internos

- **GRPC-001**: Bridge gRPC Service Specification
- **GRPC-002**: Core DICT gRPC Service Specification
- **GRPC-003**: Proto Files Specification
- **TEC-002**: Bridge Specification
- **TEC-003**: RSFN Connect Specification

### Documentação Externa

- [Protocol Buffers Guide](https://protobuf.dev/programming-guides/proto3/)
- [gRPC Go Quick Start](https://grpc.io/docs/languages/go/quickstart/)
- [Google API Design Guide](https://cloud.google.com/apis/design)
- [Buf Documentation](https://buf.build/docs/introduction)

---

## Makefile

Comandos úteis:

```makefile
# Gerar código
make proto-gen

# Lint
make proto-lint

# Breaking changes
make proto-breaking

# Limpar código gerado
make proto-clean

# Instalar dependências
make proto-deps
```

---

## Controle de Versão

Este projeto segue [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

**Versão Atual**: `v0.1.0`

Para histórico completo de mudanças, consulte [CHANGELOG.md](CHANGELOG.md).

### Histórico de Releases

| Versão  | Data       | Autor          | Descrição                          |
|---------|------------|----------------|------------------------------------|
| v0.1.0  | 2025-10-26 | api-specialist | Initial release - 29 gRPC methods  |

### Instalação

```bash
# Via Go modules
go get github.com/lbpay/dict-contracts@v0.1.0
```

---

## Suporte

**Squad**: Implementação
**Responsável**: api-specialist
**Contato**: [Squad de Implementação]

---

## Licença

Proprietário - LBPay
Uso interno apenas
