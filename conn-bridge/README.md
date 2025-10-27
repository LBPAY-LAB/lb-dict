# RSFN Bridge - SOAP/mTLS Adapter

**Projeto**: DICT - Diretorio de Identificadores de Contas Transacionais (LBPay)
**Componente**: RSFN Bridge Service (Proxy/Adaptador para Bacen)
**Versao**: 1.0.0
**Repositorio**: github.com/lbpay-lab/conn-bridge

---

## Visao Geral

O **RSFN Bridge** e um **adaptador/proxy especializado** entre o **RSFN Connect** e o **Bacen (RSFN API)**. Sua responsabilidade e preparar e executar chamadas SOAP/XML com autenticacao mTLS para o Bacen.

### Missao do Bridge
> "Receber requisicoes -> Preparar XML SOAP -> Assinar digitalmente -> Executar chamada mTLS -> Retornar resposta"

### Responsabilidades

- Preparacao de XML SOAP (construir envelopes SOAP conformes com specs do Bacen)
- Assinatura Digital XML (assinar XMLs com certificado ICP-Brasil via JRE + JAR externo)
- Autenticacao mTLS (Mutual TLS com certificados ICP-Brasil)
- Execucao de Chamadas HTTP (POST para API REST do Bacen - SOAP over HTTPS)
- Parsing de Respostas (deserializar SOAP responses)
- Circuit Breaker (protecao contra falhas em cascata)
- Retry Simples (retry imediato para falhas temporarias de rede)

### Nao-Responsabilidades (Movidas para RSFN Connect)

- Orquestracao de Workflows (ClaimWorkflow, VSYNC, etc.)
- Logica de Negocio (validacoes complexas, regras de dominio)
- Gestao de Estado de Processos Longos (7 dias de claim)
- Retry com Temporal (durabilidade de workflows)

---

## Arquitetura

### Stack Tecnologico

**Go Service (Bridge)**:
- Go 1.24.5
- gRPC (sync operations)
- Pulsar (async operations)
- Circuit Breaker (sony/gobreaker)
- OpenTelemetry (observability)

**Java Service (XML Signer)**:
- Java 17
- Apache Santuario (XML Security)
- BouncyCastle (ICP-Brasil)

### Fluxo de Integracao

```
┌──────────────┐         ┌──────────────┐         ┌─────────┐
│              │         │              │         │         │
│ RSFN Connect │ ──────> │ RSFN Bridge  │ ──────> │  Bacen  │
│  (TEC-003)   │  gRPC/  │  (TEC-002)   │  SOAP/  │  RSFN   │
│              │  Pulsar │              │  mTLS   │  API    │
│              │ <────── │              │ <────── │         │
└──────────────┘         └──────────────┘         └─────────┘
   Orquestra              Adapta/Traduz           Processa
   Workflows              SOAP + mTLS             Requisicoes
```

### Dominos Funcionais

1. **Directory** (Vinculos DICT): Create, Get, Update, Delete Entry
2. **Claim** (Reivindicacao de Posse): Create, Get, List, Confirm, Complete, Cancel Claim
3. **Key** (Validacao de Chaves): CheckKeys
4. **Reconciliation** (CID e VSYNC): GetCidSetFile, CreateCidSetFile, GetEntryByCid
5. **Antifraud** (Marcacao de Fraude): CreateFraudMarker, CancelFraudMarker, GetStatistics
6. **Policies** (Politicas DICT): ListPolicies, GetPolicy
7. **Infraction Report** (Relatorios de Infracao): Create, Get, List, Acknowledge, Cancel, Close

**Total**: 51+ operacoes mapeadas para API RSFN do Bacen.

---

## Estrutura do Repositorio

```
conn-bridge/
├── cmd/
│   └── server/
│       └── main.go                     # Entrypoint Go
├── internal/
│   ├── domain/                         # Domain models
│   ├── application/                    # Use cases
│   │   ├── ports/                      # Interfaces
│   │   └── usecases/
│   │       ├── directory/
│   │       ├── claim/
│   │       └── antifraud/
│   ├── infrastructure/                 # Adaptadores externos
│   │   ├── bacen/                      # RSFN Client (SOAP + mTLS)
│   │   ├── signer/                     # XML Signer adapter
│   │   └── pulsar/                     # Pulsar client
│   └── handlers/                       # Interface layer
│       ├── grpc/                       # gRPC handlers
│       └── pulsar/                     # Pulsar handlers
├── xml-signer/                         # Java XML Signer
│   ├── pom.xml
│   ├── Dockerfile
│   └── src/main/java/
│       └── com/lbpay/xmlsigner/
│           └── XmlSigner.java
├── certs/                              # Certificados ICP-Brasil (gitignored)
├── config/
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── go.mod
├── .env.example
└── README.md
```

---

## Quick Start

### Pre-requisitos

- Go 1.24.5+
- Java 17+
- Docker & Docker Compose
- Certificados ICP-Brasil A3 (.pfx)

### Setup

1. **Clone o repositorio**:
```bash
git clone https://github.com/lbpay-lab/conn-bridge.git
cd conn-bridge
```

2. **Configure variaveis de ambiente**:
```bash
cp .env.example .env
# Edite .env com suas configuracoes
```

3. **Prepare certificados mTLS**:
```bash
# Extrair certificado do .pfx
openssl pkcs12 -in certificado-a3.pfx -clcerts -nokeys -out certs/cert.pem
openssl pkcs12 -in certificado-a3.pfx -nocerts -out certs/key.pem

# Baixar cadeia ICP-Brasil
wget http://acraiz.icpbrasil.gov.br/credenciadas/RAIZ/ICP-Brasilv10.crt -O certs/ac-raiz.crt
```

4. **Compilar XML Signer (Java)**:
```bash
make build-signer
```

5. **Rodar servicos**:
```bash
docker-compose up -d
```

6. **Testar gRPC**:
```bash
grpcurl -plaintext -d '{"key":"11122233344","type":"CPF"}' \
  localhost:9094 BridgeService/GetEntry
```

---

## Configuracao

### Variaveis de Ambiente

Veja `.env.example` para todas as configuracoes disponiveis.

**Principais**:
- `GRPC_PORT`: Porta do servidor gRPC (default: 9094)
- `HEALTH_PORT`: Porta do health check (default: 8082)
- `BACEN_URL`: URL da API Bacen (ex: https://rsfn.bcb.gov.br/dict)
- `XML_SIGNER_URL`: URL do servico Java XML Signer (ex: http://xml-signer:8080)
- `CERT_PATH`, `KEY_PATH`, `CA_PATH`: Caminhos dos certificados mTLS

### Certificados mTLS

O Bridge requer certificados **ICP-Brasil A3** para:
1. **Autenticacao mTLS** (comunicacao com Bacen)
2. **Assinatura XML** (assinatura digital de payloads SOAP)

Armazene certificados em `certs/` (diretorio ignorado pelo git).

---

## Desenvolvimento

### Build

```bash
# Build Go service
make build

# Build XML Signer (Java)
make build-signer

# Build Docker images
make docker-build
```

### Testes

```bash
# Testes unitarios
make test

# Testes de integracao
make test-integration

# Lint
make lint
```

### Executar localmente

```bash
# Rodar Go service
make run

# Rodar com Docker Compose
docker-compose up
```

---

## Observabilidade

### Metricas (Prometheus)

- `bridge_bacen_request_duration_seconds`: Duracao de requisicoes ao Bacen
- `bridge_circuit_breaker_state`: Estado do Circuit Breaker (0=closed, 1=open, 2=half-open)
- `bridge_grpc_requests_total`: Total de requisicoes gRPC
- `bridge_pulsar_messages_total`: Total de mensagens Pulsar

### Tracing (OpenTelemetry/Jaeger)

Spans importantes:
- `bridge.grpc.GetEntry`: Chamada gRPC recebida
- `bridge.signer.SignXML`: Assinatura de XML
- `bridge.bacen.SendSOAP`: Chamada SOAP ao Bacen
- `bridge.pulsar.Consume`: Consumo de mensagem Pulsar

### Logs

Logs estruturados em JSON (logrus) com niveis configurados por `LOG_LEVEL`.

---

## Deployment

### Kubernetes

```bash
# Deploy com Helm
helm install rsfn-bridge ./charts/rsfn-bridge \
  --set image.tag=v1.0.0 \
  --set secrets.certPath=/certs/cert.pem
```

### Docker Compose (Desenvolvimento)

```bash
docker-compose up -d
```

---

## Circuit Breaker

O Bridge implementa Circuit Breaker (sony/gobreaker) para protecao contra falhas em cascata:

- **CLOSED**: Operacao normal
- **OPEN**: Circuit aberto apos 5 falhas consecutivas (rejeita requisicoes por 30s)
- **HALF-OPEN**: Testando recuperacao (permite ate 3 requests)

---

## Dual Protocol Support

O Bridge suporta **AMBOS** protocolos simultaneamente:

### gRPC (Sync)
- Operacoes de baixa latencia
- Request/Response imediato
- Timeout 30s
- Porta: 9094

### Pulsar (Async)
- Operacoes de longa duracao
- Fire-and-forget
- Desacoplamento temporal
- Topics: `rsfn-dict-req-out` (in), `rsfn-dict-res-out` (out)

---

## Referencias

- **TEC-002**: Especificacao Tecnica - Bridge Specification
- **IMP-003**: Manual de Implementacao - RSFN Bridge
- **ICP-Brasil**: http://acraiz.icpbrasil.gov.br/
- **Apache Santuario**: https://santuario.apache.org/

---

## Licenca

Proprietario - LBPay Lab

---

## Contato

- **Squad**: Implementacao
- **Arquiteto**: Thiago Lima
- **CTO**: Jose Luis Silva
