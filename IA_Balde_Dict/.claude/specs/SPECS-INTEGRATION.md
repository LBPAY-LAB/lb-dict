# SPECS-INTEGRATION.md - Integration Specification (Bridge gRPC & Pulsar Events)

**Projeto**: DICT Rate Limit Monitoring System
**Componentes**: Bridge gRPC Client + Pulsar Event Producer
**Status**: ✅ ESPECIFICAÇÃO COMPLETA - Production-Ready

---

## 🎯 Objetivo

Especificação técnica completa das **integrações externas** do sistema:

1. **Bridge gRPC Client**: Comunicação com DICT BACEN via Bridge
2. **Pulsar Event Producer**: Publicação de alertas para Core-Dict
3. **Proto Definitions**: Contratos gRPC
4. **Event Schemas**: Schemas de eventos Pulsar
5. **Error Handling**: Tratamento de erros de rede e timeout

---

## 📋 Tabela de Conteúdos

- [1. Arquitetura de Integração](#1-arquitetura-de-integração)
- [2. Bridge gRPC Client](#2-bridge-grpc-client)
- [3. Proto Definitions](#3-proto-definitions)
- [4. Pulsar Event Producer](#4-pulsar-event-producer)
- [5. Event Schemas](#5-event-schemas)
- [6. Error Handling & Retry](#6-error-handling--retry)
- [7. mTLS Configuration](#7-mtls-configuration)
- [8. Integration Testing](#8-integration-testing)
- [9. Production Checklist](#9-production-checklist)

---

## 1. Arquitetura de Integração

### Integration Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                    DICT RATE LIMIT MONITORING                        │
│                (apps/orchestration-worker)                           │
└─────────────────────────────────────────────────────────────────────┘
                              │
                ┌─────────────┴─────────────┐
                │                           │
                ▼                           ▼
┌───────────────────────────┐   ┌───────────────────────────┐
│   BRIDGE gRPC CLIENT      │   │   PULSAR PRODUCER         │
│   (Outbound)              │   │   (Outbound)              │
│                           │   │                           │
│   - ListPolicies()        │   │   - Topic: core-events    │
│   - GetPolicy(name)       │   │   - Schema: JSON          │
│   - Timeout: 5s           │   │   - Async publish         │
│   - Retry: 3x             │   │   - At-least-once         │
│   - mTLS enabled          │   │                           │
└───────────────────────────┘   └───────────────────────────┘
                │                           │
                │ gRPC (HTTP/2)             │ Pulsar Binary Protocol
                ▼                           ▼
┌───────────────────────────┐   ┌───────────────────────────┐
│   RSFN-CONNECT-BACEN-     │   │   APACHE PULSAR           │
│   BRIDGE                  │   │   (Message Broker)        │
│                           │   │                           │
│   - Port: 9090            │   │   - Port: 6650            │
│   - mTLS required         │   │   - Tenant: lb-conn       │
│   - Endpoints:            │   │   - Namespace: dict       │
│     /ratelimit.           │   │   - Topic: core-events    │
│     RateLimitService/     │   │                           │
│     ListPolicies          │   │   Consumer: Core-Dict     │
└───────────────────────────┘   └───────────────────────────┘
                │                           │
                │ HTTPS REST API             │
                ▼                           ▼
┌───────────────────────────┐   ┌───────────────────────────┐
│   DICT BACEN              │   │   CORE-DICT               │
│   (OpenAPI 3.1)           │   │   (Consumer)              │
│                           │   │                           │
│   - POST /api/v1/policies │   │   - Action:               │
│   - Response: 24 policies │   │     ActionRateLimitAlert  │
└───────────────────────────┘   └───────────────────────────┘
```

### Data Flow

```
1. Temporal Activity → Bridge gRPC Client
   ├─ Request: ListPoliciesRequest{}
   └─ Response: ListPoliciesResponse{policies: [...]}

2. Bridge → DICT BACEN REST API
   ├─ Request: POST /api/v1/policies/list
   └─ Response: JSON array of 24 policies

3. Temporal Activity → Analyze balances
   ├─ Detect: Policies exceeding WARNING/CRITICAL thresholds
   └─ Generate: AlertEvent[]

4. Temporal Activity → Pulsar Producer
   ├─ Event: ActionRateLimitAlert
   ├─ Topic: persistent://lb-conn/dict/core-events
   └─ Payload: {policy, severity, utilization, message}

5. Pulsar → Core-Dict Consumer
   ├─ Consume: ActionRateLimitAlert event
   └─ Action: Log/Dashboard/Notification
```

---

## 2. Bridge gRPC Client

### Client Interface

```go
// Location: apps/orchestration-worker/infrastructure/grpc/bridge/rate_limit_client.go
package bridge

import (
	"context"
	"time"
)

// PolicyState representa o estado de uma política retornado pelo Bridge
type PolicyState struct {
	PolicyName           string
	Category             string
	Capacity             int
	RefillTokens         int
	RefillPeriodSec      int
	AvailableTokens      int
	WarningThresholdPct  float64
	CriticalThresholdPct float64
	CheckedAt            time.Time
}

// RateLimitClient define a interface do cliente gRPC
type RateLimitClient interface {
	// ListPolicies retorna todas as políticas do DICT
	ListPolicies(ctx context.Context) (*ListPoliciesResponse, error)

	// GetPolicy retorna uma política específica
	GetPolicy(ctx context.Context, policyName string) (*PolicyState, error)

	// Close fecha a conexão gRPC
	Close() error
}

// ListPoliciesResponse representa a resposta de listagem
type ListPoliciesResponse struct {
	Policies  []PolicyState
	CheckedAt time.Time
}
```

### Client Implementation

```go
// Location: apps/orchestration-worker/infrastructure/grpc/bridge/rate_limit_client_impl.go
package bridge

import (
	"context"
	"fmt"
	"time"

	"github.com/lb-conn/connector-dict/shared/logger"
	pb "github.com/lb-conn/rsfn-connect-bacen-bridge/proto/ratelimit"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

// rateLimitClient implementa RateLimitClient
type rateLimitClient struct {
	conn       *grpc.ClientConn
	grpcClient pb.RateLimitServiceClient
	logger     logger.Logger
	tracer     trace.Tracer
	timeout    time.Duration
}

// Config representa a configuração do cliente gRPC
type Config struct {
	// Bridge address (host:port)
	Address string

	// TLS configuration
	TLSCertFile string
	TLSKeyFile  string
	TLSCAFile   string

	// Timeouts
	Timeout         time.Duration
	ConnectTimeout  time.Duration
	KeepAliveTime   time.Duration
	KeepAliveTimeout time.Duration

	// Retry
	MaxRetries int
}

// NewRateLimitClient cria uma nova instância do cliente
func NewRateLimitClient(cfg Config, logger logger.Logger) (RateLimitClient, error) {
	// Load TLS credentials
	creds, err := credentials.NewClientTLSFromFile(cfg.TLSCertFile, "")
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
	}

	// gRPC connection options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
		grpc.WithTimeout(cfg.ConnectTimeout),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                cfg.KeepAliveTime,
			Timeout:             cfg.KeepAliveTimeout,
			PermitWithoutStream: true,
		}),
	}

	// Dial
	conn, err := grpc.Dial(cfg.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial bridge: %w", err)
	}

	logger.InfoContext(context.Background(), "Bridge gRPC client connected",
		"address", cfg.Address,
	)

	return &rateLimitClient{
		conn:       conn,
		grpcClient: pb.NewRateLimitServiceClient(conn),
		logger:     logger,
		tracer:     otel.Tracer("dict.grpc.bridge.ratelimit"),
		timeout:    cfg.Timeout,
	}, nil
}

// ListPolicies implementa RateLimitClient
func (c *rateLimitClient) ListPolicies(ctx context.Context) (*ListPoliciesResponse, error) {
	ctx, span := c.tracer.Start(ctx, "BridgeClient.ListPolicies")
	defer span.End()

	// Add timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	c.logger.InfoContext(ctx, "calling Bridge.ListPolicies")

	// Call gRPC
	resp, err := c.grpcClient.ListPolicies(ctx, &pb.ListPoliciesRequest{})
	if err != nil {
		span.RecordError(err)

		// Convert gRPC error to domain error
		st := status.Convert(err)
		c.logger.ErrorContext(ctx, "Bridge.ListPolicies failed",
			"code", st.Code(),
			"message", st.Message(),
		)

		return nil, c.convertGRPCError(err)
	}

	// Convert proto to domain
	policies := make([]PolicyState, 0, len(resp.Policies))
	for _, p := range resp.Policies {
		policies = append(policies, PolicyState{
			PolicyName:           p.PolicyName,
			Category:             p.Category,
			Capacity:             int(p.Capacity),
			RefillTokens:         int(p.RefillTokens),
			RefillPeriodSec:      int(p.RefillPeriodSec),
			AvailableTokens:      int(p.AvailableTokens),
			WarningThresholdPct:  p.WarningThresholdPct,
			CriticalThresholdPct: p.CriticalThresholdPct,
			CheckedAt:            p.CheckedAt.AsTime(),
		})
	}

	c.logger.InfoContext(ctx, "Bridge.ListPolicies succeeded",
		"count", len(policies),
	)

	return &ListPoliciesResponse{
		Policies:  policies,
		CheckedAt: resp.CheckedAt.AsTime(),
	}, nil
}

// GetPolicy implementa RateLimitClient
func (c *rateLimitClient) GetPolicy(ctx context.Context, policyName string) (*PolicyState, error) {
	ctx, span := c.tracer.Start(ctx, "BridgeClient.GetPolicy")
	defer span.End()

	// Add timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	c.logger.InfoContext(ctx, "calling Bridge.GetPolicy",
		"policy", policyName,
	)

	// Call gRPC
	resp, err := c.grpcClient.GetPolicy(ctx, &pb.GetPolicyRequest{
		PolicyName: policyName,
	})
	if err != nil {
		span.RecordError(err)

		st := status.Convert(err)
		c.logger.ErrorContext(ctx, "Bridge.GetPolicy failed",
			"code", st.Code(),
			"message", st.Message(),
			"policy", policyName,
		)

		return nil, c.convertGRPCError(err)
	}

	// Convert proto to domain
	policy := &PolicyState{
		PolicyName:           resp.PolicyName,
		Category:             resp.Category,
		Capacity:             int(resp.Capacity),
		RefillTokens:         int(resp.RefillTokens),
		RefillPeriodSec:      int(resp.RefillPeriodSec),
		AvailableTokens:      int(resp.AvailableTokens),
		WarningThresholdPct:  resp.WarningThresholdPct,
		CriticalThresholdPct: resp.CriticalThresholdPct,
		CheckedAt:            resp.CheckedAt.AsTime(),
	}

	c.logger.InfoContext(ctx, "Bridge.GetPolicy succeeded",
		"policy", policyName,
	)

	return policy, nil
}

// Close implementa RateLimitClient
func (c *rateLimitClient) Close() error {
	return c.conn.Close()
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// convertGRPCError converte erro gRPC para erro de domínio
func (c *rateLimitClient) convertGRPCError(err error) error {
	st := status.Convert(err)

	switch st.Code() {
	case codes.Unavailable:
		return ErrBridgeUnavailable

	case codes.DeadlineExceeded:
		return ErrBridgeTimeout

	case codes.NotFound:
		return ErrPolicyNotFound

	case codes.InvalidArgument:
		return ErrInvalidArgument

	case codes.Unauthenticated:
		return ErrAuthenticationFailed

	default:
		return fmt.Errorf("bridge grpc error: %w", err)
	}
}
```

### Error Types

```go
// Location: apps/orchestration-worker/infrastructure/grpc/bridge/errors.go
package bridge

import "errors"

var (
	// Retryable errors
	ErrBridgeUnavailable = errors.New("bridge service unavailable")
	ErrBridgeTimeout     = errors.New("bridge request timeout")

	// Non-retryable errors
	ErrPolicyNotFound        = errors.New("policy not found")
	ErrInvalidArgument       = errors.New("invalid argument")
	ErrAuthenticationFailed  = errors.New("authentication failed")
)

// IsRetryable determina se um erro é retryable
func IsRetryable(err error) bool {
	return errors.Is(err, ErrBridgeUnavailable) ||
		errors.Is(err, ErrBridgeTimeout)
}
```

---

## 3. Proto Definitions

### RateLimitService Proto

```protobuf
// Location: shared/proto/ratelimit/rate_limit_service.proto
syntax = "proto3";

package ratelimit;

option go_package = "github.com/lb-conn/rsfn-connect-bacen-bridge/proto/ratelimit;ratelimit";

import "google/protobuf/timestamp.proto";

// RateLimitService fornece acesso aos dados de rate limit do DICT
service RateLimitService {
  // ListPolicies retorna todas as políticas de rate limit
  rpc ListPolicies(ListPoliciesRequest) returns (ListPoliciesResponse);

  // GetPolicy retorna uma política específica
  rpc GetPolicy(GetPolicyRequest) returns (PolicyResponse);
}

// ============================================================================
// REQUEST/RESPONSE MESSAGES
// ============================================================================

// ListPoliciesRequest requisita lista de todas as políticas
message ListPoliciesRequest {
  // Empty - retorna todas as 24 políticas
}

// ListPoliciesResponse contém lista de políticas
message ListPoliciesResponse {
  repeated Policy policies = 1;
  google.protobuf.Timestamp checked_at = 2;
}

// GetPolicyRequest requisita uma política específica
message GetPolicyRequest {
  string policy_name = 1;
}

// PolicyResponse contém dados de uma política
message PolicyResponse {
  Policy policy = 1;
}

// ============================================================================
// DATA MESSAGES
// ============================================================================

// Policy representa uma política de rate limit do DICT
message Policy {
  // Nome da política (ex: ENTRIES_CREATE)
  string policy_name = 1;

  // Categoria BACEN (A, B, C, D, E, F, G, H)
  string category = 2;

  // Capacidade máxima do balde
  int32 capacity = 3;

  // Tokens adicionados por refill
  int32 refill_tokens = 4;

  // Período de refill (segundos)
  int32 refill_period_sec = 5;

  // Tokens disponíveis no momento
  int32 available_tokens = 6;

  // Threshold de WARNING (% restante)
  double warning_threshold_pct = 7;

  // Threshold de CRITICAL (% restante)
  double critical_threshold_pct = 8;

  // Timestamp da consulta
  google.protobuf.Timestamp checked_at = 9;
}
```

### Proto Generation

```bash
# Location: scripts/generate_protos.sh
#!/bin/bash

# Generate Go code from proto files
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  shared/proto/ratelimit/*.proto

echo "Proto files generated successfully"
```

---

## 4. Pulsar Event Producer

### Producer Interface

```go
// Location: apps/orchestration-worker/infrastructure/pulsar/producer.go
package pulsar

import (
	"context"

	pulsarClient "github.com/apache/pulsar-client-go/pulsar"
)

// EventProducer define a interface do producer Pulsar
type EventProducer interface {
	// PublishAlert publica um alerta de rate limit
	PublishAlert(ctx context.Context, alert AlertEvent) error

	// Close fecha o producer
	Close() error
}

// AlertEvent representa um evento de alerta
type AlertEvent struct {
	PolicyName      string    `json:"policy_name"`
	Category        string    `json:"category"`
	Severity        string    `json:"severity"`
	AvailableTokens int       `json:"available_tokens"`
	CapacityMax     int       `json:"capacity_max"`
	UtilizationPct  float64   `json:"utilization_pct"`
	Message         string    `json:"message"`
	DetectedAt      time.Time `json:"detected_at"`
}
```

### Producer Implementation

```go
// Location: apps/orchestration-worker/infrastructure/pulsar/producer_impl.go
package pulsar

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pulsarClient "github.com/apache/pulsar-client-go/pulsar"
	"github.com/lb-conn/connector-dict/shared/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// eventProducer implementa EventProducer
type eventProducer struct {
	producer pulsarClient.Producer
	logger   logger.Logger
	tracer   trace.Tracer
}

// ProducerConfig representa a configuração do producer
type ProducerConfig struct {
	// Pulsar broker URL
	BrokerURL string

	// Topic (persistent://lb-conn/dict/core-events)
	Topic string

	// Producer name
	ProducerName string

	// Send timeout
	SendTimeout time.Duration

	// Batching
	BatchingEnabled   bool
	BatchingMaxDelay  time.Duration
	BatchingMaxSize   uint

	// Compression
	CompressionType pulsarClient.CompressionType
}

// NewEventProducer cria uma nova instância do producer
func NewEventProducer(cfg ProducerConfig, logger logger.Logger) (EventProducer, error) {
	// Create Pulsar client
	client, err := pulsarClient.NewClient(pulsarClient.ClientOptions{
		URL:               cfg.BrokerURL,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 10 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create pulsar client: %w", err)
	}

	// Create producer
	producer, err := client.CreateProducer(pulsarClient.ProducerOptions{
		Topic:                   cfg.Topic,
		Name:                    cfg.ProducerName,
		SendTimeout:             cfg.SendTimeout,
		DisableBatching:         !cfg.BatchingEnabled,
		BatchingMaxPublishDelay: cfg.BatchingMaxDelay,
		BatchingMaxMessages:     cfg.BatchingMaxSize,
		CompressionType:         cfg.CompressionType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	logger.InfoContext(context.Background(), "Pulsar producer created",
		"topic", cfg.Topic,
		"broker", cfg.BrokerURL,
	)

	return &eventProducer{
		producer: producer,
		logger:   logger,
		tracer:   otel.Tracer("dict.pulsar.producer"),
	}, nil
}

// PublishAlert implementa EventProducer
func (p *eventProducer) PublishAlert(ctx context.Context, alert AlertEvent) error {
	ctx, span := p.tracer.Start(ctx, "PulsarProducer.PublishAlert")
	defer span.End()

	p.logger.InfoContext(ctx, "publishing alert to Pulsar",
		"policy", alert.PolicyName,
		"severity", alert.Severity,
	)

	// Build Pulsar event payload
	payload := map[string]interface{}{
		"action": "ActionRateLimitAlert",
		"data":   alert,
		"metadata": map[string]string{
			"source":     "dict-rate-limit-monitoring",
			"version":    "1.0.0",
			"created_at": time.Now().UTC().Format(time.RFC3339),
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(payload)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Publish message
	_, err = p.producer.Send(ctx, &pulsarClient.ProducerMessage{
		Payload: data,
		Key:     alert.PolicyName,
		Properties: map[string]string{
			"action":   "ActionRateLimitAlert",
			"severity": alert.Severity,
		},
	})

	if err != nil {
		span.RecordError(err)
		p.logger.ErrorContext(ctx, "failed to publish alert",
			"error", err,
			"policy", alert.PolicyName,
		)
		return fmt.Errorf("pulsar publish failed: %w", err)
	}

	p.logger.InfoContext(ctx, "alert published successfully",
		"policy", alert.PolicyName,
		"severity", alert.Severity,
	)

	return nil
}

// Close implementa EventProducer
func (p *eventProducer) Close() error {
	p.producer.Close()
	return nil
}
```

---

## 5. Event Schemas

### ActionRateLimitAlert Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "ActionRateLimitAlert",
  "description": "Evento de alerta de rate limit do DICT BACEN",
  "type": "object",
  "required": ["action", "data", "metadata"],
  "properties": {
    "action": {
      "type": "string",
      "const": "ActionRateLimitAlert",
      "description": "Identificador da ação"
    },
    "data": {
      "type": "object",
      "required": [
        "policy_name",
        "category",
        "severity",
        "available_tokens",
        "capacity_max",
        "utilization_pct",
        "message",
        "detected_at"
      ],
      "properties": {
        "policy_name": {
          "type": "string",
          "description": "Nome da política (ex: ENTRIES_CREATE)",
          "example": "ENTRIES_CREATE"
        },
        "category": {
          "type": "string",
          "enum": ["A", "B", "C", "D", "E", "F", "G", "H"],
          "description": "Categoria BACEN",
          "example": "A"
        },
        "severity": {
          "type": "string",
          "enum": ["WARNING", "CRITICAL"],
          "description": "Severidade do alerta",
          "example": "CRITICAL"
        },
        "available_tokens": {
          "type": "integer",
          "minimum": 0,
          "description": "Tokens disponíveis no momento",
          "example": 30
        },
        "capacity_max": {
          "type": "integer",
          "minimum": 1,
          "description": "Capacidade máxima do balde",
          "example": 300
        },
        "utilization_pct": {
          "type": "number",
          "minimum": 0,
          "maximum": 100,
          "description": "Utilização percentual",
          "example": 90.0
        },
        "message": {
          "type": "string",
          "description": "Mensagem descritiva do alerta",
          "example": "URGENT: Policy ENTRIES_CREATE (Category A) is at 90.00% utilization (CRITICAL threshold exceeded)"
        },
        "detected_at": {
          "type": "string",
          "format": "date-time",
          "description": "Timestamp de detecção (ISO 8601)",
          "example": "2025-10-31T10:30:00Z"
        }
      }
    },
    "metadata": {
      "type": "object",
      "properties": {
        "source": {
          "type": "string",
          "const": "dict-rate-limit-monitoring",
          "description": "Sistema de origem"
        },
        "version": {
          "type": "string",
          "description": "Versão do schema",
          "example": "1.0.0"
        },
        "created_at": {
          "type": "string",
          "format": "date-time",
          "description": "Timestamp de criação do evento"
        }
      }
    }
  }
}
```

### Example Event

```json
{
  "action": "ActionRateLimitAlert",
  "data": {
    "policy_name": "ENTRIES_CREATE",
    "category": "A",
    "severity": "CRITICAL",
    "available_tokens": 30,
    "capacity_max": 300,
    "utilization_pct": 90.0,
    "message": "URGENT: Policy ENTRIES_CREATE (Category A) is at 90.00% utilization (CRITICAL threshold exceeded)",
    "detected_at": "2025-10-31T10:30:00Z"
  },
  "metadata": {
    "source": "dict-rate-limit-monitoring",
    "version": "1.0.0",
    "created_at": "2025-10-31T10:30:00Z"
  }
}
```

---

## 6. Error Handling & Retry

### gRPC Retry Configuration

```go
// Location: apps/orchestration-worker/infrastructure/grpc/bridge/retry.go
package bridge

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// DefaultRetryConfig retorna configuração padrão de retry para gRPC
func DefaultRetryConfig() grpc.ServiceConfig {
	return grpc.ServiceConfig{
		MethodConfig: []grpc.MethodConfig{
			{
				Name: []grpc.MethodName{
					{Service: "ratelimit.RateLimitService"},
				},
				RetryPolicy: &grpc.RetryPolicy{
					MaxAttempts:          3,
					InitialBackoff:       1 * time.Second,
					MaxBackoff:           30 * time.Second,
					BackoffMultiplier:    2.0,
					RetryableStatusCodes: []codes.Code{
						codes.Unavailable,
						codes.DeadlineExceeded,
					},
				},
			},
		},
	}
}
```

### Pulsar Retry Configuration

```go
// Location: apps/orchestration-worker/infrastructure/pulsar/retry.go
package pulsar

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// RetryPublishAlert tenta publicar alerta com retry exponencial
func RetryPublishAlert(
	ctx context.Context,
	producer EventProducer,
	alert AlertEvent,
	maxRetries int,
) error {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 30 * time.Second
	b.MaxInterval = 10 * time.Second

	operation := func() error {
		return producer.PublishAlert(ctx, alert)
	}

	return backoff.Retry(operation, backoff.WithMaxRetries(b, uint64(maxRetries)))
}
```

---

## 7. mTLS Configuration

### TLS Certificate Setup

```go
// Location: apps/orchestration-worker/infrastructure/grpc/tls/config.go
package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

// LoadTLSCredentials carrega certificados mTLS
func LoadTLSCredentials(certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
	// Load client cert/key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert: %w", err)
	}

	// Load CA cert
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert: %w", err)
	}

	// Create cert pool
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to add CA cert to pool")
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
		MinVersion:   tls.VersionTLS13,
	}

	return credentials.NewTLS(tlsConfig), nil
}
```

### Environment Configuration

```bash
# Location: .env.example (Bridge gRPC)
BRIDGE_GRPC_ADDRESS=bridge.lbpay.com:9090
BRIDGE_TLS_CERT_FILE=/certs/client.crt
BRIDGE_TLS_KEY_FILE=/certs/client.key
BRIDGE_TLS_CA_FILE=/certs/ca.crt
BRIDGE_GRPC_TIMEOUT=5s
BRIDGE_CONNECT_TIMEOUT=10s
BRIDGE_KEEPALIVE_TIME=30s
BRIDGE_KEEPALIVE_TIMEOUT=10s

# Pulsar Configuration
PULSAR_BROKER_URL=pulsar://pulsar.lbpay.com:6650
PULSAR_TOPIC=persistent://lb-conn/dict/core-events
PULSAR_PRODUCER_NAME=dict-rate-limit-monitoring
PULSAR_SEND_TIMEOUT=30s
PULSAR_BATCHING_ENABLED=true
PULSAR_BATCHING_MAX_DELAY=100ms
PULSAR_BATCHING_MAX_SIZE=100
PULSAR_COMPRESSION_TYPE=ZSTD
```

---

## 8. Integration Testing

### gRPC Mock Server

```go
// Location: apps/orchestration-worker/infrastructure/grpc/bridge/mock_server_test.go
package bridge_test

import (
	"context"
	"net"
	"testing"

	pb "github.com/lb-conn/rsfn-connect-bacen-bridge/proto/ratelimit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MockRateLimitServer implementa pb.RateLimitServiceServer
type MockRateLimitServer struct {
	pb.UnimplementedRateLimitServiceServer
}

func (m *MockRateLimitServer) ListPolicies(
	ctx context.Context,
	req *pb.ListPoliciesRequest,
) (*pb.ListPoliciesResponse, error) {
	return &pb.ListPoliciesResponse{
		Policies: []*pb.Policy{
			{
				PolicyName:           "ENTRIES_CREATE",
				Category:             "A",
				Capacity:             300,
				RefillTokens:         5,
				RefillPeriodSec:      60,
				AvailableTokens:      150,
				WarningThresholdPct:  25.0,
				CriticalThresholdPct: 10.0,
				CheckedAt:            timestamppb.Now(),
			},
		},
		CheckedAt: timestamppb.Now(),
	}, nil
}

// StartMockServer inicia servidor gRPC mock
func StartMockServer(t *testing.T) (*grpc.Server, *bufconn.Listener) {
	lis := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()
	pb.RegisterRateLimitServiceServer(server, &MockRateLimitServer{})

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Errorf("Server exited with error: %v", err)
		}
	}()

	return server, lis
}

func TestRateLimitClient_ListPolicies(t *testing.T) {
	server, lis := StartMockServer(t)
	defer server.Stop()

	// Connect to mock server
	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithInsecure(),
	)
	require.NoError(t, err)
	defer conn.Close()

	// Create client
	client := pb.NewRateLimitServiceClient(conn)

	// Test ListPolicies
	resp, err := client.ListPolicies(context.Background(), &pb.ListPoliciesRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Policies, 1)
	assert.Equal(t, "ENTRIES_CREATE", resp.Policies[0].PolicyName)
}
```

### Pulsar Integration Test

```go
// Location: apps/orchestration-worker/infrastructure/pulsar/producer_test.go
package pulsar_test

import (
	"context"
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestEventProducer_PublishAlert(t *testing.T) {
	// Start Pulsar container
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "apachepulsar/pulsar:2.11.0",
		ExposedPorts: []string{"6650/tcp", "8080/tcp"},
		WaitingFor:   wait.ForHTTP("/admin/v2/tenants").WithPort("8080"),
		Cmd:          []string{"bin/pulsar", "standalone"},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer container.Terminate(ctx)

	// Get broker URL
	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "6650")
	brokerURL := fmt.Sprintf("pulsar://%s:%s", host, port.Port())

	// Create producer
	producer, err := NewEventProducer(ProducerConfig{
		BrokerURL:    brokerURL,
		Topic:        "test-topic",
		ProducerName: "test-producer",
		SendTimeout:  10 * time.Second,
	}, nil)
	require.NoError(t, err)
	defer producer.Close()

	// Publish alert
	alert := AlertEvent{
		PolicyName:      "ENTRIES_CREATE",
		Severity:        "CRITICAL",
		AvailableTokens: 30,
		CapacityMax:     300,
		UtilizationPct:  90.0,
		Message:         "Test alert",
		DetectedAt:      time.Now(),
	}

	err = producer.PublishAlert(ctx, alert)
	assert.NoError(t, err)
}
```

---

## 9. Production Checklist

### Pre-Deployment

- [ ] Bridge gRPC endpoints validated (ListPolicies, GetPolicy)
- [ ] mTLS certificates configured and tested
- [ ] Pulsar topic created: `persistent://lb-conn/dict/core-events`
- [ ] Core-Dict consumer confirmed (consuming `core-events`)
- [ ] gRPC retry policies validated
- [ ] Timeout configurations tested (5s for Bridge)
- [ ] Event schema validated against JSON Schema
- [ ] Integration tests passing (>90% coverage)
- [ ] Load testing completed (1000 req/s)

### Post-Deployment

- [ ] Monitor Bridge gRPC latency (<100ms p99)
- [ ] Verify Pulsar publish success rate (>99%)
- [ ] Check Core-Dict event consumption
- [ ] Validate mTLS certificate expiry alerts
- [ ] Monitor gRPC connection pool health
- [ ] Verify Pulsar batching effectiveness
- [ ] Check compression ratio (ZSTD)
- [ ] Runbook created for Bridge outages

### Monitoring Metrics

```yaml
Bridge gRPC Metrics:
  - grpc_client_requests_total
  - grpc_client_request_duration_seconds
  - grpc_client_errors_total
  - grpc_client_connection_status

Pulsar Producer Metrics:
  - pulsar_producer_messages_sent_total
  - pulsar_producer_send_duration_seconds
  - pulsar_producer_errors_total
  - pulsar_producer_batching_size
```

---

**Última Atualização**: 2025-10-31
**Versão**: 1.0.0
**Status**: ✅ ESPECIFICAÇÃO COMPLETA - Production-Ready
