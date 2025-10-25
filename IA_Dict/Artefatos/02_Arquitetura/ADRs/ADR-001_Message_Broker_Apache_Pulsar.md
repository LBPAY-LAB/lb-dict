# ADR-001: Escolha de Message Broker - Apache Pulsar

**Status**: ✅ Aceito
**Data**: 2025-10-24
**Decisores**: Thiago Lima (Head de Arquitetura), José Luís Silva (CTO)
**Contexto Técnico**: Projeto DICT - LBPay

---

## Controle de Versão

| Versão | Data | Autor | Descrição das Mudanças |
|--------|------|-------|------------------------|
| 1.0 | 2025-10-24 | ARCHITECT | Versão inicial - Documentação da decisão de usar Apache Pulsar como message broker |

---

## Status

**✅ ACEITO** - Apache Pulsar já é tecnologia confirmada e em uso no LBPay

---

## Contexto

O projeto DICT da LBPay requer uma **plataforma de streaming de mensagens** para implementar uma arquitetura **Event-Driven** (EDA - Event-Driven Architecture). O sistema precisa de:

### Requisitos Funcionais

1. **Pub/Sub de Eventos de Domínio**:
   - Publicar eventos quando chaves PIX são cadastradas, excluídas, reivindicadas, etc.
   - Consumir eventos para orquestração assíncrona (Bridge Temporal Workflows)
   - Suportar múltiplos consumers independentes (ex: auditoria, analytics, notificações)

2. **Integração RSFN Connect**:
   - Filas para requests RSFN outbound: `rsfn-dict-req-out`
   - Filas para responses RSFN inbound: `rsfn-dict-res-out`
   - Garantir ordem de mensagens (FIFO) por chave PIX

3. **Event Sourcing (futuro)**:
   - Retenção de eventos de domínio para reconstrução de estado
   - Suportar replay de eventos para debugging/auditoria

### Requisitos Não-Funcionais

| ID | Requisito | Target | Fonte |
|----|-----------|--------|-------|
| **NFR-080** | Latência P95 (publish) | ≤ 10ms | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-081** | Throughput | ≥ 10.000 msgs/sec | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-082** | Durabilidade | ≥ 99.999% (zero data loss) | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-083** | Disponibilidade | ≥ 99.99% | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-084** | Retenção de mensagens | Configurável (7 dias a indefinido) | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-085** | Multi-tenancy | Suporte a namespaces isolados | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |
| **NFR-086** | Geo-replicação | Suporte a replicação entre datacenters | [NFR-001](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md) |

### Contexto Organizacional

- **LBPay já utiliza Apache Pulsar** em outros sistemas de produção
- Equipe de infraestrutura possui expertise em operação de clusters Pulsar
- Redução de custo operacional (não introduzir nova tecnologia)
- Consistência tecnológica entre projetos LBPay

---

## Decisão

**Escolhemos Apache Pulsar como plataforma de streaming/messaging para o projeto DICT.**

### Justificativa

Apache Pulsar foi escolhido pelos seguintes motivos:

#### 1. **Já em Uso no LBPay**

✅ **Apache Pulsar já é tecnologia estabelecida no LBPay**:
- Utilizado no projeto Money Moving (transações PIX)
- Infraestrutura provisionada e operacional
- Equipe treinada e experiente
- **Menor Time-to-Market** (não precisa provisionar nova stack)
- **Menor risco operacional** (tecnologia conhecida)

#### 2. **Arquitetura Superior para Event-Driven Systems**

**Apache Pulsar vs Apache Kafka**:

| Aspecto | Apache Pulsar | Apache Kafka |
|---------|---------------|--------------|
| **Arquitetura de Armazenamento** | ✅ **Separada** (BookKeeper) - escalabilidade independente | ❌ Acoplada (brokers = storage) |
| **Multi-tenancy** | ✅ **Nativo** (namespaces, tenants) | ❌ Limitado (tópicos compartilham recursos) |
| **Geo-replicação** | ✅ **Nativa** (replicação assíncrona multi-cluster) | ❌ Requer ferramentas externas (MirrorMaker) |
| **Retenção de Mensagens** | ✅ **Flexível** (time-based, size-based, indefinido) | ⚠️ Limitada (configs complexas) |
| **Latência** | ✅ **P99 < 10ms** | ⚠️ P99 ~ 20-30ms |
| **Throughput** | ✅ **Milhões msgs/sec** | ✅ Milhões msgs/sec |
| **Garantias de Entrega** | ✅ **Exactly-once**, at-least-once, at-most-once | ⚠️ At-least-once (exactly-once limitado) |
| **Ordenação** | ✅ **Por key** (partitioning) | ✅ Por partição |
| **Schema Registry** | ✅ **Built-in** (Avro, Protobuf, JSON) | ❌ Requer componente externo (Confluent) |
| **Consumo de Recursos** | ✅ **Eficiente** (compactação, tiered storage) | ⚠️ Alto uso de disco (log-based) |

**Apache Pulsar vs RabbitMQ**:

| Aspecto | Apache Pulsar | RabbitMQ |
|---------|---------------|----------|
| **Modelo** | ✅ **Pub/Sub + Queues** (híbrido) | ⚠️ Foco em queues (AMQP) |
| **Throughput** | ✅ **Milhões msgs/sec** | ❌ Milhares msgs/sec |
| **Persistência** | ✅ **Durable** (BookKeeper) | ⚠️ Opcional (impacta performance) |
| **Retenção** | ✅ **Indefinida** (event sourcing) | ❌ Mensagens consumidas = deletadas |
| **Escalabilidade** | ✅ **Horizontal** (add brokers/bookies) | ⚠️ Vertical (clustering complexo) |
| **Casos de Uso** | ✅ **Event streaming**, Pub/Sub, Queues | ⚠️ Task queues, RPC, workflows simples |

#### 3. **Event-Driven Architecture (EDA) Requirements**

**Apache Pulsar é ideal para EDA** devido a:

- **Event Sourcing**: Retenção indefinida de eventos de domínio
- **CQRS**: Suporte a múltiplos consumers lendo o mesmo stream
- **Replay de Eventos**: Consumidores podem re-processar eventos (debugging, auditoria)
- **Subscription Models**:
  - **Exclusive**: 1 consumer por subscription (ordem garantida)
  - **Shared**: Múltiplos consumers (load balancing)
  - **Failover**: 1 consumer ativo + N standby
  - **Key_Shared**: Múltiplos consumers, mas ordem por key

#### 4. **Funcionalidades Avançadas**

**Pulsar Functions** (serverless event processing):
- Processar eventos sem criar consumers standalone
- Transformações, filtragens, enriquecimento de dados
- Deploy simplificado (no need for Kubernetes pods)

**Tiered Storage**:
- Armazenar mensagens antigas em S3/GCS (custo reduzido)
- Hot data em BookKeeper, cold data em object storage
- Acesso transparente (consumer não percebe mudança)

**Schema Evolution**:
- Versionamento de schemas (Avro, Protobuf)
- Validação automática na publicação/consumo
- Backward/forward compatibility checks

#### 5. **Performance e Confiabilidade**

**Benchmarks Apache Pulsar**:
- **Throughput**: 3 milhões msgs/sec (cluster 3 brokers)
- **Latência P95**: < 5ms (publish)
- **Latência P99**: < 10ms (publish)
- **Durabilidade**: 99.9999% (replicação síncrona)
- **Disponibilidade**: 99.99% (multi-broker, multi-bookie)

**Casos de Uso Reais**:
- **Yahoo**: 100+ clusters, trillions msgs/day
- **Splunk**: Ingestão de logs em tempo real
- **Tencent**: Billing e payments (similar ao nosso caso)

---

## Consequências

### Positivas ✅

1. **Time-to-Market Reduzido**:
   - Infraestrutura Pulsar já provisionada no LBPay
   - Equipe já treinada (curva de aprendizado zero)
   - Não precisa approval de nova tecnologia

2. **Consistência Tecnológica**:
   - Mesma stack do Money Moving (sinergia)
   - Redução de complexidade operacional
   - Reutilização de bibliotecas e padrões

3. **Performance Superior**:
   - Latência P99 < 10ms (melhor que Kafka e RabbitMQ)
   - Throughput: milhões msgs/sec
   - Zero data loss (replicação síncrona BookKeeper)

4. **Escalabilidade**:
   - Separação de compute (brokers) e storage (bookies)
   - Escalabilidade horizontal independente
   - Suporte a geo-replicação nativa

5. **Event Sourcing Ready**:
   - Retenção indefinida de eventos
   - Replay de eventos para debugging/auditoria
   - CQRS pattern facilitado

6. **Multi-tenancy**:
   - Namespaces isolados por projeto
   - Quotas configuráveis por tenant
   - Segurança: ACLs granulares

7. **Custos Otimizados**:
   - Tiered storage (S3) para dados antigos
   - Compactação de mensagens
   - Deduplicação nativa

### Negativas ❌

1. **Complexidade Operacional**:
   - Mais componentes que RabbitMQ (brokers, bookies, ZooKeeper)
   - Requer expertise para tuning e troubleshooting
   - **Mitigação**: Equipe LBPay já possui expertise

2. **Curva de Aprendizado (Novos Devs)**:
   - Conceitos avançados: namespaces, tenants, subscriptions
   - Diferente de RabbitMQ/Kafka (híbrido)
   - **Mitigação**: Documentação interna + treinamentos

3. **Overhead de Configuração**:
   - Schema registry setup
   - Namespace/tenant creation
   - ACLs configuration
   - **Mitigação**: Terraform/IaC para automação

4. **Debugging Distribuído**:
   - Rastreamento de mensagens entre brokers/bookies
   - Requer tooling adequado (Pulsar Manager, Prometheus)
   - **Mitigação**: Observabilidade desde o início (logs, métricas, tracing)

### Riscos e Mitigações

| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| **Falha de cluster Pulsar** | Baixa | Alto | Multi-broker deployment, monitoramento 24/7, runbooks |
| **Performance degradation** | Média | Médio | Capacity planning, load testing, auto-scaling |
| **Erros de configuração** | Média | Alto | IaC (Terraform), peer review de configs, staging env |
| **Perda de mensagens** | Muito Baixa | Crítico | Replicação síncrona (ack quorum), backups regulares |

---

## Alternativas Consideradas

### Alternativa 1: Apache Kafka

**Prós**:
- ✅ Amplamente adotado (ecossistema maduro)
- ✅ Performance comprovada (milhões msgs/sec)
- ✅ Grande comunidade e documentação

**Contras**:
- ❌ **Não usado no LBPay** (introduziria nova stack)
- ❌ Arquitetura acoplada (brokers = storage)
- ❌ Geo-replicação requer MirrorMaker (complexidade)
- ❌ Multi-tenancy limitado
- ❌ Schema registry externo (Confluent)
- ❌ Latência superior (~20-30ms P99)

**Decisão**: ❌ **Rejeitado** - Pulsar já em uso, arquitetura superior

### Alternativa 2: RabbitMQ

**Prós**:
- ✅ Simples de operar (menos componentes)
- ✅ AMQP protocol (padrão)
- ✅ Bom para task queues

**Contras**:
- ❌ **Não usado no LBPay para streaming**
- ❌ Throughput limitado (milhares msgs/sec, não milhões)
- ❌ Retenção de mensagens não é foco (mensagens consumidas = deletadas)
- ❌ Escalabilidade vertical (clustering complexo)
- ❌ Não adequado para event sourcing
- ❌ Latência superior em high throughput

**Decisão**: ❌ **Rejeitado** - Inadequado para event streaming/EDA

### Alternativa 3: AWS SNS/SQS

**Prós**:
- ✅ Managed service (zero ops)
- ✅ Integração nativa AWS
- ✅ Escalabilidade automática

**Contras**:
- ❌ **Vendor lock-in** (AWS-only)
- ❌ Custos variáveis (milhões msgs/mês = caro)
- ❌ Latência superior (~50-100ms)
- ❌ Retenção limitada (14 dias SQS)
- ❌ Não suporta event sourcing (retenção indefinida)
- ❌ Funcionalidades limitadas vs Pulsar

**Decisão**: ❌ **Rejeitado** - Lock-in, custos, funcionalidades limitadas

### Alternativa 4: NATS JetStream

**Prós**:
- ✅ Extremamente rápido (latência < 1ms)
- ✅ Simples (escrito em Go)
- ✅ Leve (low resource usage)

**Contras**:
- ❌ **Não usado no LBPay**
- ❌ Comunidade menor que Kafka/Pulsar
- ❌ Funcionalidades limitadas (vs Pulsar/Kafka)
- ❌ Menos maduro para casos enterprise complexos

**Decisão**: ❌ **Rejeitado** - Não usado no LBPay, menos maduro

---

## Implementação

### Topologia Pulsar para Projeto DICT

```
Pulsar Cluster (LBPay Production)
│
├── Tenant: lbpay
│   └── Namespace: dict
│       ├── Topic: dict_domain_events (persistent)
│       │   ├── Subscription: bridge-orchestrator (exclusive)
│       │   ├── Subscription: audit-consumer (shared)
│       │   └── Subscription: analytics-consumer (shared)
│       │
│       ├── Topic: rsfn-dict-req-out (persistent)
│       │   └── Subscription: rsfn-connect-consumer (exclusive)
│       │
│       ├── Topic: rsfn-dict-res-in (persistent)
│       │   └── Subscription: bridge-response-handler (exclusive)
│       │
│       └── Topic: dict_notifications (persistent)
│           ├── Subscription: push-notification-service (shared)
│           ├── Subscription: email-service (shared)
│           └── Subscription: sms-service (shared)
```

### Configuração de Tópicos

#### Topic: `dict_domain_events`

**Propósito**: Eventos de domínio principais (KeyRegistered, ClaimReceived, etc.)

**Configuração**:
```yaml
topic: persistent://lbpay/dict/dict_domain_events
schema: AVRO (auto-update enabled)
retention:
  time: -1  # Indefinido (event sourcing)
  size: 1TB
partitions: 8
replication: 3
deduplication: enabled
ttl: -1  # Sem expiração
compaction: disabled  # Manter todos os eventos
```

**Subscriptions**:
- `bridge-orchestrator`: Exclusive (ordem garantida)
- `audit-consumer`: Shared (load balancing)
- `analytics-consumer`: Shared (load balancing)

#### Topic: `rsfn-dict-req-out`

**Propósito**: Requests RSFN outbound (CreateEntry, DeleteEntry, etc.)

**Configuração**:
```yaml
topic: persistent://lbpay/dict/rsfn-dict-req-out
schema: AVRO
retention:
  time: 7d  # 7 dias (tempo de retry máximo)
  size: 100GB
partitions: 4
replication: 3
deduplication: enabled
message_ttl: 7d
compaction: disabled
```

**Subscriptions**:
- `rsfn-connect-consumer`: Exclusive (ordem por chave garantida via key_shared)

#### Topic: `rsfn-dict-res-in`

**Propósito**: Responses RSFN inbound (callbacks Bacen)

**Configuração**:
```yaml
topic: persistent://lbpay/dict/rsfn-dict-res-in
schema: AVRO
retention:
  time: 7d
  size: 100GB
partitions: 4
replication: 3
deduplication: enabled
message_ttl: 7d
```

**Subscriptions**:
- `bridge-response-handler`: Exclusive

#### Topic: `dict_notifications`

**Propósito**: Eventos para notificações de usuários

**Configuração**:
```yaml
topic: persistent://lbpay/dict/dict_notifications
schema: AVRO
retention:
  time: 3d  # 3 dias (notificações não precisam retenção longa)
  size: 50GB
partitions: 8
replication: 3
deduplication: enabled
message_ttl: 3d
```

**Subscriptions**:
- `push-notification-service`: Shared
- `email-service`: Shared
- `sms-service`: Shared

### Schema Avro para Eventos

**Exemplo: `KeyRegistered` event**:

```json
{
  "type": "record",
  "name": "KeyRegistered",
  "namespace": "com.lbpay.dict.events",
  "doc": "Evento publicado quando uma chave PIX é cadastrada com sucesso no DICT Bacen",
  "fields": [
    {
      "name": "event_id",
      "type": "string",
      "doc": "UUID do evento"
    },
    {
      "name": "event_type",
      "type": "string",
      "default": "KeyRegistered"
    },
    {
      "name": "timestamp",
      "type": "string",
      "doc": "ISO 8601 timestamp"
    },
    {
      "name": "aggregate_id",
      "type": "string",
      "doc": "ID da chave PIX"
    },
    {
      "name": "aggregate_type",
      "type": "string",
      "default": "DictKey"
    },
    {
      "name": "version",
      "type": "int",
      "doc": "Versão do evento"
    },
    {
      "name": "payload",
      "type": {
        "type": "record",
        "name": "KeyRegisteredPayload",
        "fields": [
          {"name": "key_id", "type": "string"},
          {"name": "key_type", "type": "string"},
          {"name": "key_value", "type": "string"},
          {"name": "account_id", "type": "string"},
          {"name": "ispb", "type": "string"},
          {"name": "bacen_entry_id", "type": "string"},
          {"name": "creation_date", "type": "string"}
        ]
      }
    },
    {
      "name": "metadata",
      "type": {
        "type": "record",
        "name": "EventMetadata",
        "fields": [
          {"name": "correlation_id", "type": "string"},
          {"name": "causation_id", "type": "string"},
          {"name": "user_id", "type": ["null", "string"], "default": null}
        ]
      }
    }
  ]
}
```

### Client Libraries

**Go (Linguagem do Projeto)**:

```go
import (
    "github.com/apache/pulsar-client-go/pulsar"
)

// Producer example
func CreateProducer(client pulsar.Client) (pulsar.Producer, error) {
    return client.CreateProducer(pulsar.ProducerOptions{
        Topic:           "persistent://lbpay/dict/dict_domain_events",
        Schema:          pulsar.NewAvroSchema(schemaDefinition, nil),
        CompressionType: pulsar.ZSTD,  // Compressão eficiente
        BatchingMaxMessages: 1000,
        BatchingMaxPublishDelay: 10 * time.Millisecond,
    })
}

// Consumer example
func CreateConsumer(client pulsar.Client) (pulsar.Consumer, error) {
    return client.Subscribe(pulsar.ConsumerOptions{
        Topic:            "persistent://lbpay/dict/dict_domain_events",
        SubscriptionName: "bridge-orchestrator",
        Type:             pulsar.Exclusive,  // Ordem garantida
        Schema:           pulsar.NewAvroSchema(schemaDefinition, nil),
    })
}
```

### Segurança

**Autenticação**:
- **TLS**: Enabled (mTLS entre clients e brokers)
- **Auth Plugin**: JWT tokens

**Autorização (ACLs)**:
```yaml
# Namespace: lbpay/dict
acls:
  - role: dict-core-service
    permissions:
      - produce: dict_domain_events
      - consume: dict_domain_events

  - role: bridge-service
    permissions:
      - produce: rsfn-dict-req-out
      - consume: dict_domain_events, rsfn-dict-res-in

  - role: rsfn-connect-service
    permissions:
      - consume: rsfn-dict-req-out
      - produce: rsfn-dict-res-in

  - role: audit-service
    permissions:
      - consume: dict_domain_events
```

### Monitoramento

**Métricas Prometheus**:
- `pulsar_producer_send_latency_ms` (P50, P95, P99)
- `pulsar_consumer_receive_latency_ms`
- `pulsar_producer_msg_rate_in`
- `pulsar_consumer_msg_rate_out`
- `pulsar_storage_size` (por topic)
- `pulsar_backlog_size` (mensagens pendentes por subscription)

**Alertas**:
- Backlog > 10.000 mensagens (consumer lag)
- Latência P95 > 50ms (performance degradation)
- Storage size > 80% quota (capacity)
- Producer failures > 1% (connectivity issues)

**Dashboards Grafana**:
- Topic throughput (msgs/sec)
- Consumer lag (por subscription)
- Latency percentiles (P50, P95, P99)
- Storage usage (por namespace)

---

## Rastreabilidade

### Requisitos Funcionais Impactados

| CRF | Descrição | Impacto |
|-----|-----------|---------|
| [CRF-001](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-001) | Cadastrar Chave CPF | Evento `KeyRegisterRequested` publicado no Pulsar |
| [CRF-020](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-020) | Solicitar Claim | Evento `ClaimRequested` publicado no Pulsar |
| [CRF-040](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-040) | Excluir Chave | Evento `KeyDeletionRequested` publicado no Pulsar |
| [CRF-080](../05_Requisitos/CRF-001_Requisitos_Funcionais.md#crf-080) | Auditoria | Consumer `audit-service` consome todos os eventos |

### NFRs Impactados

| NFR | Descrição | Como Pulsar Atende |
|-----|-----------|-------------------|
| [NFR-080](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-080) | Latência P95 ≤ 10ms | Pulsar: P95 < 5ms (benchmark) ✅ |
| [NFR-081](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-081) | Throughput ≥ 10k msgs/sec | Pulsar: 3M msgs/sec ✅ |
| [NFR-082](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-082) | Durabilidade 99.999% | BookKeeper replicação síncrona ✅ |
| [NFR-083](../05_Requisitos/NFR-001_Requisitos_Nao_Funcionais.md#nfr-083) | Disponibilidade 99.99% | Multi-broker, multi-bookie ✅ |

### Processos BPMN Impactados

| PRO | Descrição | Integração Pulsar |
|-----|-----------|-------------------|
| [PRO-001](../04_Processos/PRO-001_Processos_BPMN.md#pro-001) | Cadastro CPF | Core DICT → Pulsar → Bridge |
| [PRO-006](../04_Processos/PRO-001_Processos_BPMN.md#pro-006) | Reivindicação | Core DICT → Pulsar → Bridge |
| [PRO-015](../04_Processos/PRO-001_Processos_BPMN.md#pro-015) | VSYNC | Core DICT → Pulsar → Audit |

---

## Referências

### Documentação Técnica

- [Apache Pulsar Documentation](https://pulsar.apache.org/docs/en/)
- [Pulsar Go Client](https://github.com/apache/pulsar-client-go)
- [Pulsar Schema Registry](https://pulsar.apache.org/docs/en/schema-get-started/)
- [BookKeeper (Storage Layer)](https://bookkeeper.apache.org/)

### Benchmarks e Case Studies

- [Yahoo: Pulsar at Scale](https://yahooeng.tumblr.com/post/150078336821/open-sourcing-pulsar-pub-sub-messaging-at-scale)
- [Splunk: Why We Chose Pulsar](https://www.splunk.com/en_us/blog/it/why-splunk-chose-pulsar.html)
- [Pulsar Performance Benchmarks](https://pulsar.apache.org/blog/2020/10/09/pulsar-2-6-0-benchmark/)

### Arquitetura LBPay

- [ArquiteturaDict_LBPAY.md](../../Docs_iniciais/ArquiteturaDict_LBPAY.md): Diagramas SVG mostrando tópicos Pulsar
- [ARE-001](./ARE-001_Analise_Repositorios_Existentes.md): Análise do uso de Pulsar em repos existentes

---

## Aprovação

- [x] **Thiago Lima** (Head de Arquitetura) - 2025-10-24
- [x] **José Luís Silva** (CTO) - 2025-10-24

**Rationale**: Apache Pulsar já é tecnologia confirmada e em uso no LBPay. Esta ADR documenta a decisão e fundamenta o uso técnico no projeto DICT.

---

**FIM DO DOCUMENTO ADR-001**
