# RSFN Connect - Temporal Orchestration

**Projeto**: DICT - Diretorio de Identificadores de Contas Transacionais (LBPay)
**Componente**: RSFN Connect (Orchestration Service with Temporal Workflows)
**Versao**: 1.0
**Stack**: Go 1.24.5, Temporal v1.36.0, Pulsar v0.16.0

---

## Visao Geral

O **RSFN Connect** e o modulo **orquestrador** entre o **Core Bancario DICT** e o **RSFN Bridge**, responsavel por:

- Orquestrar Workflows de Longa Duracao via Temporal (Reivindicacoes de 30 dias)
- Implementar Logica de Negocio (validacoes, transformacoes, decisoes)
- Gerenciar Estado de processos assincronos
- Consumir mensagens de `dict.api` via **Pulsar**
- Produzir requisicoes para **Bridge** via **gRPC** ou **Pulsar**
- Receber respostas do Bridge e repassar para `dict.api`

## Arquitetura

```
┌─────────────────────────────────────────────────────────────────┐
│                    Core Bancario DICT (dict.api)                 │
│  - gRPC Server (FrontEnd, BackOffice)                            │
│  - Pulsar Producer → persistent://lb-conn/dict/rsfn-dict-req-out│
└─────────────────────────────────────────────────────────────────┘
                           ↓ gRPC sincrono (high-perf)
                           ↓ Pulsar assincrono (long-running)
┌─────────────────────────────────────────────────────────────────┐
│                     RSFN Connect (Orchestrator)                  │
│                    ORQUESTRADOR COM TEMPORAL                     │
│                                                                   │
│  - Temporal Workflows (v1.36.0)                                  │
│    - ClaimWorkflow (30 dias de monitoramento)                    │
│    - VSYNCWorkflow (planejado - cron diario 00:00 BRT)           │
│    - OTPWorkflow (planejado - validacao OTP)                     │
│                                                                   │
│  - Pulsar Consumer                                               │
│    - Topico: persistent://lb-conn/dict/rsfn-dict-req-out        │
│                                                                   │
│  - Bridge Client (gRPC)                                          │
│    - Endpoint: bridge-grpc-svc:50051                             │
│                                                                   │
│  - Pulsar Producer (Response to dict.api)                        │
│    - Topico: persistent://lb-conn/dict/rsfn-dict-res-out        │
└─────────────────────────────────────────────────────────────────┘
                           ↓ gRPC ou Pulsar
┌─────────────────────────────────────────────────────────────────┐
│                   RSFN Bridge (SOAP/mTLS Adapter)                │
│  - Prepara: Payload SOAP/XML                                    │
│  - Assina: XML com ICP-Brasil (JRE + JAR)                       │
│  - Envia: HTTPS mTLS para Bacen                                 │
└─────────────────────────────────────────────────────────────────┘
                           ↓ HTTPS mTLS
┌─────────────────────────────────────────────────────────────────┐
│                         Bacen DICT (RSFN)                        │
│  - API DICT/SPI (SOAP/XML)                                       │
└─────────────────────────────────────────────────────────────────┘
```

## Stack Tecnologica

| Componente | Tecnologia | Versao | Justificativa |
|------------|------------|--------|---------------|
| **Linguagem** | Go | 1.24.5 | Performance, concorrencia nativa |
| **Workflows** | Temporal SDK | v1.36.0 | Processos de longa duracao (claims 30 dias) |
| **Mensageria** | Apache Pulsar | v0.16.0 | Event-driven architecture |
| **gRPC** | gRPC Client | Latest | Baixa latencia |
| **Database** | PostgreSQL | 15+ | Estado de workflows, CID |
| **Observability** | OpenTelemetry | v1.38.0 | Traces, logs estruturados |

## Estrutura do Projeto

```
conn-dict/
├── cmd/
│   ├── connect/           # API/Consumer principal
│   │   └── main.go
│   └── worker/            # Temporal Workers
│       └── main.go
├── internal/
│   ├── workflows/         # Temporal Workflows
│   │   ├── claim_workflow.go
│   │   ├── vsync_workflow.go
│   │   └── otp_workflow.go
│   ├── activities/        # Temporal Activities
│   │   ├── bridge/
│   │   ├── cache/
│   │   └── events/
│   ├── handlers/          # Pulsar handlers
│   │   └── entry_handler.go
│   ├── services/          # Business logic
│   │   ├── entry_service.go
│   │   └── claim_service.go
│   └── infrastructure/    # Clients e adapters
│       ├── grpc/
│       ├── pulsar/
│       └── redis/
├── pkg/
│   ├── config/
│   └── logger/
├── test/
│   ├── integration/
│   └── unit/
├── .env.example
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── go.mod
└── README.md
```

## Workflows Implementados

### ClaimWorkflow (30 dias)

Processo de reivindicacao tem **30 dias** para ser confirmado ou cancelado. Apos 30 dias, o sistema automaticamente cancela.

**Fluxo**:
1. Criar reivindicacao no Bacen (via Bridge)
2. Aguardar sinal de confirmacao ou cancelamento (timeout de 30 dias)
3. Confirmar ou Cancelar no Bacen
4. Transferir chave para novo dono (se confirmado)
5. Notificar usuarios (via dict.api)

### VSYNCWorkflow (Planejado)

Workflow que executara **diariamente as 00:00 BRT** para sincronizar contas do Core Bancario com o DICT Bacen.

### OTPWorkflow (Planejado)

Workflow para validacao de OTP em operacoes sensiveis (portabilidade de chave, reivindicacao).

## Quick Start

### Pre-requisitos

- Go 1.24.5+
- Docker & Docker Compose
- Temporal Server
- Apache Pulsar
- Redis

### Instalacao

1. Clone o repositorio:
```bash
git clone https://github.com/lbpay-lab/conn-dict.git
cd conn-dict
```

2. Configure variaveis de ambiente:
```bash
cp .env.example .env
# Editar .env com suas configuracoes
```

3. Instalar dependencias:
```bash
go mod download
```

4. Iniciar servicos de infraestrutura:
```bash
docker-compose up -d
```

5. Executar Worker Temporal:
```bash
make run-worker
```

6. Executar API/Consumer:
```bash
make run
```

### Comandos Make

```bash
# Build
make build           # Compilar binarios
make docker-build    # Build Docker images

# Execucao
make run            # Executar API/Consumer
make run-worker     # Executar Temporal Worker

# Testes
make test           # Executar todos os testes
make test-unit      # Testes unitarios
make test-int       # Testes de integracao

# Qualidade
make lint           # Executar linter
make fmt            # Formatar codigo

# Limpeza
make clean          # Limpar binarios
```

## Configuracao

Principais variaveis de ambiente (ver `.env.example`):

- `TEMPORAL_HOST`: Endereco do Temporal Server
- `PULSAR_URL`: URL do Apache Pulsar
- `GRPC_PORT`: Porta do servidor gRPC (9092)
- `ADMIN_PORT`: Porta HTTP admin/health (8081)
- `BRIDGE_GRPC_ADDR`: Endereco do Bridge gRPC
- `REDIS_URL`: URL do Redis

## Casos de Uso Principais

### 1. Criar Entrada DICT (Chave PIX)

1. `dict.api` envia mensagem para `persistent://lb-conn/dict/rsfn-dict-req-out`
2. Connect consome mensagem
3. `CreateEntryUseCase` valida dados
4. Connect chama Bridge via gRPC: `CreateEntry(entryData)`
5. Bridge retorna resposta para Connect
6. Connect persiste entrada no PostgreSQL (CID local)
7. Connect envia resposta para dict.api via `persistent://lb-conn/dict/rsfn-dict-res-out`

### 2. Criar Reivindicacao (30 dias)

1. `dict.api` envia mensagem de reivindicacao
2. Connect consome mensagem
3. `CreateClaimUseCase` persiste claim (status PENDING)
4. Connect inicia `ClaimWorkflow` no Temporal
5. Workflow executa `CreateClaimActivity` → chama Bridge → Bacen
6. Workflow aguarda 30 dias OU sinal de confirmacao/cancelamento
7. Ao receber sinal (ou timeout), Workflow executa `ConfirmClaimActivity` ou `CancelClaimActivity`
8. Workflow executa `NotifyUsersActivity` → dict.api

## Observabilidade

### Logs

Logs estruturados em formato JSON usando OpenTelemetry:

```json
{
  "timestamp": "2025-10-26T10:30:00Z",
  "level": "info",
  "msg": "CreateEntryUseCase iniciado",
  "entryKey": "12345678901",
  "type": "CPF",
  "traceID": "abc123..."
}
```

### Metricas

- `dict_claim_workflows_started_total`: Total de ClaimWorkflows iniciados
- `dict_claim_workflows_completed_total`: Total de ClaimWorkflows completados (por status)
- `dict_entries_created_total`: Total de entradas criadas
- `dict_bridge_requests_total`: Total de requisicoes para Bridge

### Traces

Propagacao de trace context via OpenTelemetry para Bridge e dict.api.

## Documentacao Tecnica

- [TEC-003: RSFN Connect Specification](docs/TEC-003_RSFN_Connect_Specification.md)
- [IMP-002: Manual de Implementacao](docs/IMP-002_Manual_Implementacao_Connect.md)
- [Temporal Documentation](https://docs.temporal.io/)

## Contribuindo

1. Fork o repositorio
2. Crie uma branch: `git checkout -b feature/minha-feature`
3. Commit suas mudancas: `git commit -m 'feat: adicionar nova feature'`
4. Push para a branch: `git push origin feature/minha-feature`
5. Abra um Pull Request

## Licenca

Copyright (c) 2025 LBPay. Todos os direitos reservados.

---

**Contato**: Squad de Implementacao - LBPay
**Documentacao**: [Wiki Interna](https://wiki.lbpay.com.br/dict)
