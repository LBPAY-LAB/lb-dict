# Core DICT - Business Logic

**Projeto**: DICT - Diretorio de Identificadores de Contas Transacionais (LBPay)
**Componente**: Core DICT (Domain Service)
**Versao**: 1.0
**Stack**: Go 1.24.5, Fiber v3, PostgreSQL 16, Redis, Apache Pulsar

---

## Visao Geral

O **Core DICT** e o servico central do sistema DICT da LBPay, responsavel por:

- Implementar toda a **logica de dominio** (regras de negocio PIX)
- Gerenciar **entidades de dominio** (Chaves PIX, Claims, Portabilidades)
- Expor **APIs gRPC** para clientes (LB-Connect)
- Persistir dados no **PostgreSQL**
- Publicar **eventos de dominio** no **Apache Pulsar**
- Validar **requisitos regulatorios** do Bacen

---

## Arquitetura

**Clean Architecture** (Domain-Driven Design):

```
┌─────────────────────────────────────────────────────────┐
│                   Interface Layer                        │
│  (gRPC Server, Interceptors, Validators)                │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                   Use Case Layer                         │
│  (RegisterKeyUseCase, DeleteKeyUseCase, etc.)           │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                   Domain Layer                           │
│  (Entities, Value Objects, Domain Services, Events)     │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                 Infrastructure Layer                     │
│  (PostgreSQL, Redis, Pulsar, External APIs)             │
└─────────────────────────────────────────────────────────┘
```

---

## Stack Tecnologica

| Componente | Tecnologia | Versao | Justificativa |
|------------|------------|--------|---------------|
| **Linguagem** | Go (Golang) | 1.24.5 | Performance, concorrencia, type-safe |
| **Framework** | Fiber | v3 | HTTP framework de alta performance |
| **APIs** | gRPC + Protocol Buffers | v1.62+ | Comunicacao eficiente entre servicos |
| **Database** | PostgreSQL | 16+ | Banco de dados relacional robusto |
| **Cache** | Redis | v9.14.1 | Cache distribuido de alta performance |
| **Message Broker** | Apache Pulsar | v0.16.0 | Sistema de mensageria distribuido |
| **Observability** | Prometheus, Grafana, OpenTelemetry | Latest | Metricas, traces, logs |

---

## Estrutura de Diretorios

```
core-dict/
├── api/
│   └── proto/                          # Definicoes gRPC/Protobuf
│       └── dict/v1/
├── cmd/
│   └── server/
│       └── main.go                     # Entrypoint da aplicacao
├── internal/
│   ├── domain/                         # Domain Layer (Clean Architecture)
│   │   ├── entity/                     # Entidades de dominio
│   │   ├── valueobject/                # Value Objects
│   │   ├── event/                      # Eventos de dominio
│   │   ├── repository/                 # Interfaces de repositorio
│   │   └── service/                    # Servicos de dominio
│   ├── usecase/                        # Use Case Layer
│   │   ├── register_key.go
│   │   ├── delete_key.go
│   │   └── ...
│   ├── interface/                      # Interface Layer
│   │   ├── grpc/                       # gRPC handlers
│   │   └── dto/                        # DTOs e mappers
│   └── infrastructure/                 # Infrastructure Layer
│       ├── persistence/
│       │   ├── postgres/
│       │   └── redis/
│       ├── messaging/
│       │   └── pulsar/
│       └── config/
├── migrations/                         # Database migrations
├── test/
│   ├── unit/
│   ├── integration/
│   └── e2e/
├── pkg/                                # Pacotes reutilizaveis
├── .env.example                        # Exemplo de variaveis de ambiente
├── Dockerfile                          # Container image
├── docker-compose.yml                  # Orquestracao local
├── Makefile                            # Automacao de build
├── go.mod                              # Dependencias Go
└── README.md                           # Este arquivo
```

---

## Requisitos

### Software

- **Go**: 1.24.5+
- **PostgreSQL**: 16+
- **Redis**: 7+
- **Apache Pulsar**: 3.0+
- **Docker**: 20.10+ (opcional, para desenvolvimento local)
- **Make**: 3.81+

### Dependencias Go

Principais dependencias (ver `go.mod` para lista completa):

- `github.com/gofiber/fiber/v3` - Framework HTTP
- `google.golang.org/grpc` - gRPC server/client
- `gorm.io/gorm` - ORM para PostgreSQL
- `github.com/redis/go-redis/v9` - Cliente Redis
- `github.com/apache/pulsar-client-go/pulsar` - Cliente Pulsar
- `github.com/google/uuid` - Geracao de UUIDs
- `github.com/golang-jwt/jwt/v5` - Autenticacao JWT

---

## Configuracao

### 1. Clonar Repositorio

```bash
git clone https://github.com/lbpay-lab/core-dict.git
cd core-dict
```

### 2. Configurar Variaveis de Ambiente

Copie `.env.example` para `.env` e ajuste os valores:

```bash
cp .env.example .env
```

Edite `.env` com suas configuracoes locais.

### 3. Subir Dependencias com Docker Compose

```bash
docker-compose up -d
```

Isso iniciara:
- PostgreSQL (porta 5432)
- Redis (porta 6379)
- Apache Pulsar (porta 6650)

### 4. Executar Migrations

```bash
make migrate-up
```

### 5. Iniciar Aplicacao

```bash
# Desenvolvimento (com hot-reload)
make run

# Producao
make build
./bin/core-dict
```

---

## Comandos Disponiveis (Makefile)

```bash
make build         # Compila a aplicacao
make test          # Executa testes unitarios
make lint          # Executa linter (golangci-lint)
make run           # Executa aplicacao em modo desenvolvimento
make docker-build  # Constroi imagem Docker
make migrate-up    # Aplica migrations
make migrate-down  # Reverte migrations
```

---

## APIs

### gRPC Service

O Core DICT expoe um servico gRPC `DictService` com os seguintes metodos:

- `RegisterKey` - Registra uma chave PIX
- `DeleteKey` - Remove uma chave PIX
- `GetEntry` - Consulta uma entrada DICT
- `ListKeys` - Lista chaves de uma conta
- `CreateClaim` - Cria um claim de portabilidade/posse
- `RespondClaim` - Responde a um claim

**Porta padrao**: 9090

### REST API (Fiber)

APIs REST para integracao com FrontEnd/BackOffice.

**Porta padrao**: 8080

Documentacao Swagger disponivel em: `http://localhost:8080/swagger`

---

## Eventos de Dominio

O Core DICT publica eventos de dominio no Apache Pulsar:

| Evento | Topico | Descricao |
|--------|--------|-----------|
| `KeyRegisterRequested` | `persistent://dict/events/key-register-requested` | Nova chave solicitada |
| `KeyRegistered` | `persistent://dict/events/key-registered` | Chave confirmada pelo Bacen |
| `KeyDeleted` | `persistent://dict/events/key-deleted` | Chave removida |
| `ClaimReceived` | `persistent://dict/events/claim-received` | Claim recebido de outro PSP |
| `ClaimAccepted` | `persistent://dict/events/claim-accepted` | Claim aceito |

---

## Validacoes de Negocio (PIX)

### Limites de Chaves por Titular

- **CPF**: Maximo 5 chaves por titular
- **CNPJ**: Maximo 20 chaves por titular
- **Email**: Maximo 20 chaves por titular
- **Telefone**: Maximo 20 chaves por titular
- **EVP (Chave Aleatoria)**: Maximo 20 chaves por titular

### Validacoes de Formato

- **CPF**: 11 digitos numericos, com validacao de digito verificador
- **CNPJ**: 14 digitos numericos, com validacao de digito verificador
- **Email**: RFC 5322 compliant
- **Telefone**: Formato E.164 (+5511999998888)
- **EVP**: UUID v4

### Regras de Claim

- **Prazo de resposta**: 7 dias corridos
- **Auto-confirmacao**: Se nao houver resposta em 7 dias, claim e auto-confirmado
- **Validacao de titularidade**: CPF/CNPJ da chave deve pertencer ao titular da conta

---

## Testes

### Testes Unitarios

```bash
make test
```

### Testes de Integracao

```bash
make test-integration
```

### Cobertura de Testes

```bash
make test-coverage
```

Objetivo: **>80% de cobertura**

---

## Observabilidade

### Metricas (Prometheus)

Metricas expostas em: `http://localhost:8080/metrics`

Principais metricas:
- `dict_keys_total` - Total de chaves cadastradas
- `dict_claims_total` - Total de claims
- `dict_api_requests_total` - Total de requisicoes HTTP
- `dict_grpc_requests_total` - Total de requisicoes gRPC

### Traces (OpenTelemetry)

Traces enviados para Jaeger/Zipkin (configuravel via env vars).

### Logs (Structured Logging)

Logs estruturados em formato JSON para facil integracao com ELK/Splunk.

---

## Deployment

### Docker

```bash
# Build da imagem
make docker-build

# Executar container
docker run -p 8080:8080 -p 9090:9090 \
  --env-file .env \
  lbpay/core-dict:latest
```

### Kubernetes

Manifests disponiveis em: `/k8s/`

```bash
kubectl apply -f k8s/
```

---

## Seguranca

### Autenticacao

- **JWT** para APIs REST
- **mTLS** para gRPC (producao)

### Autorizacao

- Role-based access control (RBAC)
- Scopes: `dict:read`, `dict:write`, `dict:admin`

### Criptografia

- Dados em transito: TLS 1.3
- Dados em repouso: PostgreSQL encryption at rest

---

## Contribuindo

1. Fork o repositorio
2. Crie uma branch (`git checkout -b feature/nova-feature`)
3. Commit suas mudancas (`git commit -m 'Add nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

---

## Documentacao Adicional

- **Especificacao Tecnica**: [TEC-001_Core_DICT_Specification.md](docs/TEC-001_Core_DICT_Specification.md)
- **Manual de Implementacao**: [IMP-001_Manual_Implementacao_Core_DICT.md](docs/IMP-001_Manual_Implementacao_Core_DICT.md)
- **Schema de Banco de Dados**: [DAT-001_Schema_Database_Core_DICT.md](docs/DAT-001_Schema_Database_Core_DICT.md)
- **API Reference**: [API.md](docs/API.md)

---

## Licenca

Proprietary - LBPay (c) 2025

---

## Contato

- **Time**: Squad de Implementacao
- **Tech Lead**: Backend-Core
- **Slack**: #squad-implementacao
- **Email**: dev@lbpay.com.br
