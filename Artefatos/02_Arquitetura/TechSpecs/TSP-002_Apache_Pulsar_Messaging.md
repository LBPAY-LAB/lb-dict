# TSP-002: Apache Pulsar Messaging - Technical Specification

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: Apache Pulsar Messaging Platform
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: BACKEND (AI Agent - Backend Specialist)
**Revisor**: [Aguardando]
**Aprovador**: Tech Lead, Head de Arquitetura

---

## Sumário Executivo

Este documento especifica a implementação do **Apache Pulsar** (v0.16.0 Go Client) para o projeto DICT LBPay, cobrindo deployment de brokers e bookies em Kubernetes, topologia de tópicos, estratégias de particionamento, políticas de retenção, e padrões de producer/consumer.

**Baseado em**:
- [TEC-001: Core DICT Specification](../../11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md)
- [TEC-003: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [ADR-002: Event-Driven Architecture](../ADR-002_Event_Driven_Architecture.md) (pendente)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | BACKEND | Versão inicial - Apache Pulsar specification |

---

## Índice

1. [Visão Geral](#1-visão-geral)
2. [Deployment Kubernetes](#2-deployment-kubernetes)
3. [Topic Hierarchy](#3-topic-hierarchy)
4. [Partitioning Strategy](#4-partitioning-strategy)
5. [Retention Policies](#5-retention-policies)
6. [Producer Patterns](#6-producer-patterns)
7. [Consumer Patterns](#7-consumer-patterns)
8. [Schema Registry](#8-schema-registry)
9. [Monitoring & Observability](#9-monitoring--observability)
10. [High Availability](#10-high-availability)

---

## 1. Visão Geral

### 1.1. Pulsar Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Apache Pulsar Cluster                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Brokers (Stateless - 3 replicas)                          │ │
│  │  - Port: 6650 (Binary Protocol)                            │ │
│  │  - Port: 8080 (HTTP Admin)                                 │ │
│  │  - Port: 8081 (Metrics)                                    │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           ↓                                       │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  BookKeeper (Stateful - 3 bookies)                         │ │
│  │  - Persistent storage for messages                         │ │
│  │  - Quorum-based writes (Write Quorum: 3, Ack Quorum: 2)   │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           ↓                                       │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  ZooKeeper (Coordination - 3 replicas)                     │ │
│  │  - Metadata storage                                         │ │
│  │  - Cluster coordination                                     │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2. Key Features

| Feature | Value | Justification |
|---------|-------|---------------|
| **Go Client Version** | v0.16.0 | Latest stable for Go |
| **Brokers** | 3 replicas (stateless) | High availability |
| **Bookies** | 3 replicas (stateful) | Persistent storage |
| **ZooKeeper** | 3 replicas | Quorum for metadata |
| **Replication Factor** | 3 | Durability guarantee |
| **Namespace** | `lb-conn/dict` | Logical isolation |
| **Retention** | 7 days (default) | Balance storage/compliance |
| **Max Message Size** | 5 MB | Large payloads support |
| **Throughput** | 1M msg/sec | High performance |

---

## 2. Deployment Kubernetes

### 2.1. Namespace

```yaml
# k8s/pulsar-namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: pulsar
  labels:
    name: pulsar
    environment: production
```

### 2.2. ZooKeeper StatefulSet

```yaml
# k8s/zookeeper.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: zookeeper
  namespace: pulsar
spec:
  serviceName: zookeeper-headless
  replicas: 3
  selector:
    matchLabels:
      app: zookeeper
  template:
    metadata:
      labels:
        app: zookeeper
    spec:
      containers:
      - name: zookeeper
        image: zookeeper:3.8.3
        ports:
        - containerPort: 2181
          name: client
        - containerPort: 2888
          name: follower
        - containerPort: 3888
          name: election
        env:
        - name: ZOO_MY_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: ZOO_SERVERS
          value: "server.1=zookeeper-0.zookeeper-headless:2888:3888;2181 server.2=zookeeper-1.zookeeper-headless:2888:3888;2181 server.3=zookeeper-2.zookeeper-headless:2888:3888;2181"
        volumeMounts:
        - name: data
          mountPath: /data
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 10Gi
---
apiVersion: v1
kind: Service
metadata:
  name: zookeeper-headless
  namespace: pulsar
spec:
  type: ClusterIP
  clusterIP: None
  ports:
  - port: 2181
    name: client
  - port: 2888
    name: follower
  - port: 3888
    name: election
  selector:
    app: zookeeper
---
apiVersion: v1
kind: Service
metadata:
  name: zookeeper
  namespace: pulsar
spec:
  type: ClusterIP
  ports:
  - port: 2181
    name: client
  selector:
    app: zookeeper
```

### 2.3. BookKeeper StatefulSet

```yaml
# k8s/bookkeeper.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: bookkeeper
  namespace: pulsar
spec:
  serviceName: bookkeeper-headless
  replicas: 3
  selector:
    matchLabels:
      app: bookkeeper
  template:
    metadata:
      labels:
        app: bookkeeper
    spec:
      initContainers:
      - name: init-metadata
        image: apachepulsar/pulsar:3.2.0
        command:
        - /bin/bash
        - -c
        - |
          bin/bookkeeper shell metaformat -nonInteractive -force
        env:
        - name: BOOKIE_ZK_SERVERS
          value: "zookeeper:2181"
      containers:
      - name: bookkeeper
        image: apachepulsar/pulsar:3.2.0
        command:
        - /bin/bash
        - -c
        - |
          bin/bookkeeper bookie
        ports:
        - containerPort: 3181
          name: bookie
        env:
        - name: BOOKIE_ZK_SERVERS
          value: "zookeeper:2181"
        - name: BOOKIE_JOURNAL_DIR
          value: "/pulsar/data/bookkeeper/journal"
        - name: BOOKIE_LEDGERS_DIR
          value: "/pulsar/data/bookkeeper/ledgers"
        volumeMounts:
        - name: journal
          mountPath: /pulsar/data/bookkeeper/journal
        - name: ledgers
          mountPath: /pulsar/data/bookkeeper/ledgers
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
  volumeClaimTemplates:
  - metadata:
      name: journal
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "fast-ssd"
      resources:
        requests:
          storage: 20Gi
  - metadata:
      name: ledgers
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 50Gi
---
apiVersion: v1
kind: Service
metadata:
  name: bookkeeper-headless
  namespace: pulsar
spec:
  type: ClusterIP
  clusterIP: None
  ports:
  - port: 3181
    name: bookie
  selector:
    app: bookkeeper
```

### 2.4. Pulsar Broker Deployment

```yaml
# k8s/pulsar-broker.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pulsar-broker
  namespace: pulsar
spec:
  replicas: 3
  selector:
    matchLabels:
      app: pulsar-broker
  template:
    metadata:
      labels:
        app: pulsar-broker
    spec:
      containers:
      - name: broker
        image: apachepulsar/pulsar:3.2.0
        command:
        - /bin/bash
        - -c
        - |
          bin/pulsar broker
        ports:
        - containerPort: 6650
          name: pulsar
        - containerPort: 8080
          name: http
        - containerPort: 8081
          name: metrics
        env:
        - name: PULSAR_MEM
          value: "-Xms2g -Xmx2g"
        - name: zookeeperServers
          value: "zookeeper:2181"
        - name: configurationStoreServers
          value: "zookeeper:2181"
        - name: clusterName
          value: "lbpay-pulsar"
        - name: managedLedgerDefaultEnsembleSize
          value: "3"
        - name: managedLedgerDefaultWriteQuorum
          value: "3"
        - name: managedLedgerDefaultAckQuorum
          value: "2"
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
        livenessProbe:
          httpGet:
            path: /metrics
            port: 8080
          initialDelaySeconds: 60
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /metrics
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: pulsar-broker
  namespace: pulsar
spec:
  type: ClusterIP
  ports:
  - port: 6650
    name: pulsar
    targetPort: 6650
  - port: 8080
    name: http
    targetPort: 8080
  - port: 8081
    name: metrics
    targetPort: 8081
  selector:
    app: pulsar-broker
```

### 2.5. Pulsar Admin Deployment

```yaml
# k8s/pulsar-admin.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pulsar-admin
  namespace: pulsar
spec:
  replicas: 1
  selector:
    matchLabels:
      app: pulsar-admin
  template:
    metadata:
      labels:
        app: pulsar-admin
    spec:
      containers:
      - name: admin
        image: apachepulsar/pulsar-manager:v0.3.0
        ports:
        - containerPort: 9527
          name: admin-ui
        env:
        - name: SPRING_CONFIGURATION_FILE
          value: "/pulsar-manager/pulsar-manager/application.properties"
---
apiVersion: v1
kind: Service
metadata:
  name: pulsar-admin
  namespace: pulsar
spec:
  type: ClusterIP
  ports:
  - port: 9527
    targetPort: 9527
  selector:
    app: pulsar-admin
```

---

## 3. Topic Hierarchy

### 3.1. Topic Naming Convention

```
persistent://{tenant}/{namespace}/{topic}

tenant: lb-conn (LBPay Connect)
namespace: dict (DICT system)
topic: {domain}.{entity}.{event_type}
```

### 3.2. DICT Topics

```yaml
# Tenant: lb-conn
# Namespace: dict

topics:
  # Entry (PIX Key) Events
  - name: persistent://lb-conn/dict/dict.entries.created
    description: PIX key created
    schema: EntryCreatedEvent
    partitions: 10

  - name: persistent://lb-conn/dict/dict.entries.updated
    description: PIX key updated
    schema: EntryUpdatedEvent
    partitions: 10

  - name: persistent://lb-conn/dict/dict.entries.deleted
    description: PIX key deleted
    schema: EntryDeletedEvent
    partitions: 10

  # Claim (Reivindicação) Events
  - name: persistent://lb-conn/dict/dict.claims.created
    description: Claim initiated (30-day period)
    schema: ClaimCreatedEvent
    partitions: 5

  - name: persistent://lb-conn/dict/dict.claims.status_updated
    description: Claim status changed
    schema: ClaimStatusUpdatedEvent
    partitions: 5

  - name: persistent://lb-conn/dict/dict.claims.completed
    description: Claim completed (ownership transferred)
    schema: ClaimCompletedEvent
    partitions: 5

  - name: persistent://lb-conn/dict/dict.claims.cancelled
    description: Claim cancelled
    schema: ClaimCancelledEvent
    partitions: 5

  - name: persistent://lb-conn/dict/dict.claims.expired
    description: Claim expired (30 days passed)
    schema: ClaimExpiredEvent
    partitions: 5

  # Portability Events
  - name: persistent://lb-conn/dict/dict.portabilities.initiated
    description: Portability process started
    schema: PortabilityInitiatedEvent
    partitions: 5

  - name: persistent://lb-conn/dict/dict.portabilities.completed
    description: Portability completed
    schema: PortabilityCompletedEvent
    partitions: 5

  # Request/Response Topics (Connect <-> Bridge)
  - name: persistent://lb-conn/dict/rsfn-dict-req-out
    description: Requests from dict.api to Connect
    schema: DictRequest
    partitions: 20

  - name: persistent://lb-conn/dict/rsfn-dict-res-out
    description: Responses from Connect to dict.api
    schema: DictResponse
    partitions: 20

  # Audit Trail
  - name: persistent://lb-conn/dict/dict.audit.events
    description: Audit trail for all DICT operations
    schema: AuditEvent
    partitions: 10
```

### 3.3. Create Topics (CLI)

```bash
# Create namespace
bin/pulsar-admin namespaces create lb-conn/dict

# Create entry topics
bin/pulsar-admin topics create-partitioned-topic \
  persistent://lb-conn/dict/dict.entries.created \
  --partitions 10

bin/pulsar-admin topics create-partitioned-topic \
  persistent://lb-conn/dict/dict.entries.updated \
  --partitions 10

bin/pulsar-admin topics create-partitioned-topic \
  persistent://lb-conn/dict/dict.entries.deleted \
  --partitions 10

# Create claim topics
bin/pulsar-admin topics create-partitioned-topic \
  persistent://lb-conn/dict/dict.claims.created \
  --partitions 5

bin/pulsar-admin topics create-partitioned-topic \
  persistent://lb-conn/dict/dict.claims.status_updated \
  --partitions 5

bin/pulsar-admin topics create-partitioned-topic \
  persistent://lb-conn/dict/dict.claims.completed \
  --partitions 5

bin/pulsar-admin topics create-partitioned-topic \
  persistent://lb-conn/dict/dict.claims.cancelled \
  --partitions 5

bin/pulsar-admin topics create-partitioned-topic \
  persistent://lb-conn/dict/dict.claims.expired \
  --partitions 5

# Create request/response topics
bin/pulsar-admin topics create-partitioned-topic \
  persistent://lb-conn/dict/rsfn-dict-req-out \
  --partitions 20

bin/pulsar-admin topics create-partitioned-topic \
  persistent://lb-conn/dict/rsfn-dict-res-out \
  --partitions 20
```

---

## 4. Partitioning Strategy

### 4.1. Partition Key Strategy

**Entry Topics**: Partition by `key_value` (CPF, CNPJ, email, phone, EVP)
```go
producer.Send(ctx, &pulsar.ProducerMessage{
    Key:     entry.KeyValue,  // "12345678901"
    Payload: entryJSON,
})
```

**Claim Topics**: Partition by `claim_id`
```go
producer.Send(ctx, &pulsar.ProducerMessage{
    Key:     claim.ClaimID,  // UUID
    Payload: claimJSON,
})
```

**Request/Response Topics**: Partition by `correlation_id`
```go
producer.Send(ctx, &pulsar.ProducerMessage{
    Key:     request.CorrelationID,  // UUID
    Payload: requestJSON,
})
```

### 4.2. Partition Count Justification

| Topic Pattern | Partitions | Justification |
|---------------|------------|---------------|
| `dict.entries.*` | 10 | High volume (~10k entries/sec) |
| `dict.claims.*` | 5 | Medium volume (~100 claims/sec) |
| `dict.portabilities.*` | 5 | Low volume (~10 portabilities/sec) |
| `rsfn-dict-req-out` | 20 | Very high volume (~50k req/sec) |
| `rsfn-dict-res-out` | 20 | Very high volume (~50k res/sec) |
| `dict.audit.*` | 10 | Audit all operations |

---

## 5. Retention Policies

### 5.1. Default Retention

```yaml
# Default: 7 days
retention:
  time: 168h  # 7 days
  size: 100GB  # Per topic
```

### 5.2. Per-Topic Retention

```bash
# Entry topics: 30 days (compliance)
bin/pulsar-admin namespaces set-retention lb-conn/dict \
  --time 720h \
  --size 500GB

# Claim topics: 90 days (legal requirement - 30 days + 60 buffer)
bin/pulsar-admin topics set-retention \
  persistent://lb-conn/dict/dict.claims.created \
  --time 2160h \
  --size 100GB

# Audit topics: 365 days (compliance)
bin/pulsar-admin topics set-retention \
  persistent://lb-conn/dict/dict.audit.events \
  --time 8760h \
  --size 1TB

# Request/Response topics: 7 days (operational)
bin/pulsar-admin topics set-retention \
  persistent://lb-conn/dict/rsfn-dict-req-out \
  --time 168h \
  --size 200GB
```

### 5.3. Compaction (Claims)

Enable compaction for claim topics to keep only latest state:

```bash
bin/pulsar-admin topics set-compaction-threshold \
  persistent://lb-conn/dict/dict.claims.status_updated \
  --threshold 100M

bin/pulsar-admin topics compact \
  persistent://lb-conn/dict/dict.claims.status_updated
```

---

## 6. Producer Patterns

### 6.1. Go Producer Configuration

```go
// internal/infrastructure/pulsar/producer.go
package pulsar

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

type ProducerConfig struct {
	PulsarURL string
	Topic     string
}

func NewProducer(cfg ProducerConfig) (pulsar.Producer, error) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               cfg.PulsarURL,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pulsar client: %w", err)
	}

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic:           cfg.Topic,
		Name:            "dict-producer",
		CompressionType: pulsar.LZ4,
		BatchingMaxMessages: 1000,
		BatchingMaxPublishDelay: 10 * time.Millisecond,
		MaxPendingMessages: 10000,
		SendTimeout: 30 * time.Second,
		Properties: map[string]string{
			"service": "dict-api",
			"version": "1.0",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return producer, nil
}
```

### 6.2. Send Message with Idempotency

```go
// Send EntryCreatedEvent
func (p *Producer) SendEntryCreatedEvent(ctx context.Context, entry *Entry) error {
	eventJSON, err := json.Marshal(EntryCreatedEvent{
		EntryID:   entry.ID,
		KeyType:   entry.KeyType,
		KeyValue:  entry.KeyValue,
		CreatedAt: entry.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	_, err = p.producer.Send(ctx, &pulsar.ProducerMessage{
		Key:     entry.KeyValue,  // Partition key
		Payload: eventJSON,
		Properties: map[string]string{
			"event_type":      "EntryCreated",
			"correlation_id":  entry.CorrelationID,
			"idempotency_key": entry.IdempotencyKey,
		},
		EventTime: entry.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
```

---

## 7. Consumer Patterns

### 7.1. Go Consumer Configuration

```go
// internal/infrastructure/pulsar/consumer.go
package pulsar

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

type ConsumerConfig struct {
	PulsarURL        string
	Topic            string
	SubscriptionName string
	SubscriptionType pulsar.SubscriptionType
}

func NewConsumer(cfg ConsumerConfig) (pulsar.Consumer, error) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               cfg.PulsarURL,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pulsar client: %w", err)
	}

	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            cfg.Topic,
		SubscriptionName: cfg.SubscriptionName,
		Type:             cfg.SubscriptionType,  // Shared, Exclusive, Failover, Key_Shared
		ReceiverQueueSize: 1000,
		NackRedeliveryDelay: 1 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return consumer, nil
}
```

### 7.2. Consume Messages (Connect Worker)

```go
// apps/dict/handlers/pulsar/consumer.go
package pulsar

import (
	"context"
	"encoding/json"
	"log"

	"github.com/apache/pulsar-client-go/pulsar"
)

type DictConsumer struct {
	consumer pulsar.Consumer
	handler  *EntryHandler
}

func (c *DictConsumer) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer stopping...")
			c.consumer.Close()
			return
		default:
			msg, err := c.consumer.Receive(ctx)
			if err != nil {
				log.Printf("Error receiving message: %v\n", err)
				continue
			}

			// Process message
			go c.processMessage(ctx, msg)
		}
	}
}

func (c *DictConsumer) processMessage(ctx context.Context, msg pulsar.Message) {
	var request DictRequest
	if err := json.Unmarshal(msg.Payload(), &request); err != nil {
		log.Printf("Error unmarshaling message: %v\n", err)
		c.consumer.Nack(msg)
		return
	}

	// Route to appropriate handler
	switch request.Type {
	case "CREATE_ENTRY":
		err := c.handler.HandleCreateEntry(ctx, request)
		if err != nil {
			log.Printf("Error processing CREATE_ENTRY: %v\n", err)
			c.consumer.Nack(msg)
			return
		}
	case "CREATE_CLAIM":
		err := c.handler.HandleCreateClaim(ctx, request)
		if err != nil {
			log.Printf("Error processing CREATE_CLAIM: %v\n", err)
			c.consumer.Nack(msg)
			return
		}
	default:
		log.Printf("Unknown message type: %s\n", request.Type)
		c.consumer.Nack(msg)
		return
	}

	// Ack only if processing succeeded
	c.consumer.Ack(msg)
}
```

### 7.3. Subscription Types

| Type | Use Case | Guarantees |
|------|----------|------------|
| `Shared` | Multiple consumers, load balancing | At-least-once delivery |
| `Exclusive` | Single consumer, ordered processing | Exactly-once (with dedup) |
| `Failover` | Active-passive HA | Ordered processing with failover |
| `Key_Shared` | Ordered per key, parallel processing | Ordered per partition key |

**DICT Usage**:
- `rsfn-dict-req-out`: **Shared** (Connect workers load balance)
- `dict.entries.*`: **Key_Shared** (ordered per key_value)
- `dict.claims.*`: **Key_Shared** (ordered per claim_id)

---

## 8. Schema Registry

### 8.1. Enable Schema Registry

```bash
# Enable schema validation
bin/pulsar-admin namespaces set-schema-validation-enforced lb-conn/dict
```

### 8.2. EntryCreatedEvent Schema

```json
{
  "type": "record",
  "name": "EntryCreatedEvent",
  "namespace": "com.lbpay.dict.events",
  "fields": [
    {"name": "entry_id", "type": "string"},
    {"name": "key_type", "type": {"type": "enum", "name": "KeyType", "symbols": ["CPF", "CNPJ", "EMAIL", "PHONE", "EVP"]}},
    {"name": "key_value", "type": "string"},
    {"name": "account_number", "type": "string"},
    {"name": "participant_ispb", "type": "string"},
    {"name": "created_at", "type": "long", "logicalType": "timestamp-millis"}
  ]
}
```

### 8.3. Upload Schema

```bash
bin/pulsar-admin schemas upload \
  persistent://lb-conn/dict/dict.entries.created \
  --filename schemas/entry_created_event.avsc
```

---

## 9. Monitoring & Observability

### 9.1. Prometheus Metrics

**Exposed Metrics** (Port 8081):

```yaml
# Producer metrics
pulsar_producer_num_msg_sent
pulsar_producer_num_bytes_sent
pulsar_producer_send_latency_ms

# Consumer metrics
pulsar_consumer_num_msgs_received
pulsar_consumer_num_bytes_received
pulsar_consumer_receive_queue_size
pulsar_consumer_num_acks_sent

# Topic metrics
pulsar_topics_count
pulsar_subscriptions_count
pulsar_msg_backlog
pulsar_storage_size
```

### 9.2. Grafana Dashboard

**Key Panels**:
- Message throughput (msg/sec)
- Topic backlog depth
- Consumer lag
- Producer latency (p50, p95, p99)
- Storage usage per topic

### 9.3. Alerting Rules

```yaml
# prometheus/pulsar-alerts.yaml
groups:
  - name: pulsar
    interval: 30s
    rules:
      - alert: PulsarHighMessageBacklog
        expr: pulsar_msg_backlog > 100000
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High message backlog in topic {{ $labels.topic }}"
          description: "Message backlog is {{ $value }} messages"

      - alert: PulsarBrokerDown
        expr: up{job="pulsar-broker"} == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Pulsar broker {{ $labels.instance }} is down"

      - alert: PulsarConsumerLag
        expr: pulsar_consumer_receive_queue_size > 5000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Consumer lag detected on {{ $labels.subscription }}"
```

---

## 10. High Availability

### 10.1. Replication

**Geo-Replication** (future):

```bash
# Create geo-replicated namespace
bin/pulsar-admin namespaces create lb-conn/dict-geo \
  --clusters lbpay-sp,lbpay-rj

# Set replication
bin/pulsar-admin namespaces set-clusters lb-conn/dict-geo \
  --clusters lbpay-sp,lbpay-rj
```

### 10.2. Disaster Recovery

**Backup Strategy**:

```yaml
# k8s/pulsar-backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: pulsar-bookkeeper-backup
  namespace: pulsar
spec:
  schedule: "0 3 * * *"  # Daily at 3 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: apachepulsar/pulsar:3.2.0
            command:
            - /bin/bash
            - -c
            - |
              bin/bookkeeper shell listledgers | \
              xargs -I {} bin/bookkeeper shell readledger {} > /backups/ledgers_$(date +%Y%m%d).txt
              aws s3 cp /backups/ledgers_$(date +%Y%m%d).txt s3://lbpay-backups/pulsar/
            volumeMounts:
            - name: backups
              mountPath: /backups
          volumes:
          - name: backups
            emptyDir: {}
          restartPolicy: OnFailure
```

---

## Rastreabilidade

### Requisitos Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RF-TSP-001 | Produzir eventos de entrada (CREATE, UPDATE, DELETE) | [TEC-001](../../11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md) | ✅ Especificado |
| RF-TSP-002 | Produzir eventos de claims (CREATE, COMPLETE, CANCEL, EXPIRE) | [TEC-003](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md) | ✅ Especificado |
| RF-TSP-003 | Consumir requisições de dict.api | [TEC-003](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md) | ✅ Especificado |
| RF-TSP-004 | Publicar audit trail de todas operações | Compliance | ✅ Especificado |

### Requisitos Não-Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RNF-TSP-001 | Throughput: 1M msg/sec | Performance | ✅ Especificado |
| RNF-TSP-002 | Durabilidade: Replication Factor 3 | HA | ✅ Especificado |
| RNF-TSP-003 | Retenção: 7-365 dias (por tópico) | Compliance | ✅ Especificado |
| RNF-TSP-004 | Latência p99 < 100ms | Performance | ✅ Especificado |

---

## Próximas Revisões

**Pendências**:
- [ ] Implementar geo-replication (multi-region)
- [ ] Configurar schema evolution policies
- [ ] Implementar dead-letter queue (DLQ) para mensagens falhadas
- [ ] Validar compression (LZ4 vs Zstd)
- [ ] Criar Grafana dashboards completos

---

**Referências**:
- [Apache Pulsar Documentation](https://pulsar.apache.org/docs/)
- [Pulsar Go Client v0.16.0](https://github.com/apache/pulsar-client-go)
- [TEC-003: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
