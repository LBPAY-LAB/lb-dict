# Quick Start Guide - RSFN Connect

Guia rapido para iniciar o desenvolvimento do RSFN Connect.

## Pre-requisitos

Certifique-se de ter instalado:

- [Go 1.24.5+](https://go.dev/dl/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Make](https://www.gnu.org/software/make/)
- [Git](https://git-scm.com/)

## Passo 1: Clone o Repositorio

```bash
git clone https://github.com/lbpay-lab/conn-dict.git
cd conn-dict
```

## Passo 2: Configure as Variaveis de Ambiente

```bash
# Copiar exemplo de configuracao
cp .env.example .env

# Editar configuracoes (se necessario)
vim .env
```

Configuracoes minimas necessarias:
- `TEMPORAL_HOST`: Deixar como `localhost:7233` para desenvolvimento local
- `PULSAR_URL`: Deixar como `pulsar://localhost:6650`
- `REDIS_URL`: Deixar como `redis://localhost:6379`
- `BRIDGE_GRPC_ADDR`: Configurar endpoint do Bridge (se disponivel)

## Passo 3: Iniciar Servicos de Infraestrutura

```bash
# Iniciar Temporal, Pulsar, Redis via Docker Compose
make docker-compose-up

# Aguardar servicos iniciarem (1-2 minutos)
# Verificar status
docker-compose ps
```

Servicos que devem estar rodando:
- `conn-dict-temporal`: Temporal Server (porta 7233)
- `conn-dict-temporal-ui`: Temporal UI (porta 8088)
- `conn-dict-pulsar`: Apache Pulsar (portas 6650, 8080)
- `conn-dict-redis`: Redis (porta 6379)
- `conn-dict-postgres`: PostgreSQL (porta 5434)

## Passo 4: Instalar Dependencias Go

```bash
# Baixar dependencias
make install

# Verificar que tudo foi baixado
go mod verify
```

## Passo 5: Executar Testes

```bash
# Executar todos os testes
make test

# Ou apenas testes unitarios
make test-unit
```

## Passo 6: Executar a Aplicacao Localmente

Voce precisa de **dois terminais**:

### Terminal 1: Temporal Worker

```bash
# Executar Worker (processa workflows)
make run-worker
```

Voce deve ver:
```
INFO  Temporal Worker iniciado no TaskQueue: dict-task-queue
```

### Terminal 2: Connect API/Consumer

```bash
# Executar API/Consumer (consome Pulsar)
make run
```

Voce deve ver:
```
INFO  Connect API iniciado na porta :8081
INFO  Pulsar Consumer conectado ao topico: persistent://lb-conn/dict/rsfn-dict-req-out
```

## Passo 7: Verificar Health

```bash
# Verificar health check
make health-check

# Ou manualmente
curl http://localhost:8081/health
```

Resposta esperada:
```json
{
  "status": "healthy",
  "temporal": "connected",
  "pulsar": "connected",
  "redis": "connected"
}
```

## Passo 8: Acessar Temporal UI

Abrir no navegador:
```
http://localhost:8088
```

Ou via comando:
```bash
make temporal-ui
```

No Temporal UI voce pode:
- Ver workflows em execucao
- Monitorar task queues
- Debugar workflows
- Ver historico de execucoes

## Desenvolvimento

### Executar Linter

```bash
make lint
```

### Formatar Codigo

```bash
make fmt
```

### Build para Producao

```bash
# Build binarios
make build

# Build imagens Docker
make docker-build
```

## Testar Workflow Manualmente

### Usando Temporal CLI

```bash
# Iniciar ClaimWorkflow
temporal workflow start \
  --task-queue dict-task-queue \
  --type CreateClaimWorkflow \
  --workflow-id claim-test-001 \
  --input '{
    "ClaimID": "claim-001",
    "EntryKey": "12345678901",
    "ClaimerISPB": "87654321",
    "OwnerISPB": "12345678"
  }'

# Ver status do workflow
temporal workflow describe --workflow-id claim-test-001

# Enviar signal de confirmacao
temporal workflow signal \
  --workflow-id claim-test-001 \
  --name claim-decision \
  --input true

# Ver historico
temporal workflow show --workflow-id claim-test-001
```

## Parar Servicos

```bash
# Parar aplicacao: Ctrl+C nos terminais

# Parar servicos Docker
make docker-compose-down

# Limpar volumes (CUIDADO: apaga dados)
docker-compose down -v
```

## Troubleshooting

### Temporal nao conecta

```bash
# Verificar se container esta rodando
docker-compose ps temporal

# Ver logs
docker-compose logs temporal

# Reiniciar
docker-compose restart temporal
```

### Pulsar nao conecta

```bash
# Verificar se container esta rodando
docker-compose ps pulsar

# Ver logs
docker-compose logs pulsar

# Verificar healthcheck
curl http://localhost:8080/admin/v2/brokers/health
```

### Redis nao conecta

```bash
# Verificar se container esta rodando
docker-compose ps redis

# Testar conexao
docker exec -it conn-dict-redis redis-cli ping
```

## Proximos Passos

1. Ler [docs/ARCHITECTURE.md](./ARCHITECTURE.md)
2. Ler [TEC-003: RSFN Connect Specification](../../Artefatos/11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
3. Implementar primeiro workflow (ClaimWorkflow)
4. Configurar integracao com Bridge
5. Adicionar testes de integracao

## Links Uteis

- [Temporal Documentation](https://docs.temporal.io/)
- [Go Temporal SDK](https://pkg.go.dev/go.temporal.io/sdk)
- [Apache Pulsar Go Client](https://pulsar.apache.org/docs/client-libraries-go/)
- [Redis Go Client](https://redis.uptrace.dev/)

---

**Suporte**: Squad de Implementacao - LBPay
