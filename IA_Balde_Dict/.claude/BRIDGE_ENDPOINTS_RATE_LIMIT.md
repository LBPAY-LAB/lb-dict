# Bridge gRPC Endpoints - Rate Limit Monitoring
**Data**: 2025-11-01
**Status**: ✅ Endpoints EXISTEM e estão PRONTOS para uso
**Repo Bridge**: `github.com/lb-conn/rsfn-connect-bacen-bridge`

---

## 📋 Resumo Executivo

O **Bridge** já possui endpoints gRPC implementados para consulta de políticas de rate limit do DICT BACEN. Estes endpoints podem ser **reutilizados diretamente** pelo sistema de monitoramento de rate limit.

### Status de Implementação

| Endpoint | Status | Proto Definition | Mapper | Pronto? |
|----------|--------|------------------|--------|---------|
| GetRateLimitPolicies | ✅ Implementado | ✅ Existe | ✅ XML→gRPC | ✅ SIM |
| GetRateLimitPolicy | ✅ Implementado | ✅ Existe | ✅ XML→gRPC | ✅ SIM |

**Conclusão**: ✅ **SEM BLOQUEADORES** - Integração pode começar imediatamente.

---

## 🔌 Endpoints Disponíveis

### 1. GetRateLimitPolicies (List All Policies)

**Descrição**: Consulta **todas as políticas** de rate limit do participante (PSP).

#### Proto Definition

```protobuf
// Location: proto/bacen/dict/v2/rate_limit_service.proto

service RateLimitService {
  // Lista todas as políticas de rate limit do participante
  rpc GetRateLimitPolicies(GetRateLimitPoliciesRequest)
      returns (GetRateLimitPoliciesResponse);
}

message GetRateLimitPoliciesRequest {
  // Vazio - retorna todas as políticas do participante autenticado
}

message GetRateLimitPoliciesResponse {
  // Timestamp da resposta do DICT (autoridade)
  google.protobuf.Timestamp response_time = 1;

  // Categoria do participante (A-H)
  string category = 2;

  // Lista de políticas
  repeated RateLimitPolicy policies = 3;
}

message RateLimitPolicy {
  // Nome da política (ex: "ENTRIES_WRITE")
  string name = 1;

  // Fichas disponíveis no momento da consulta
  int32 available_tokens = 2;

  // Capacidade máxima do balde
  int32 capacity = 3;

  // Quantidade de fichas repostas por período
  int32 refill_tokens = 4;

  // Período de reposição em segundos
  int32 refill_period_sec = 5;

  // Categoria específica da política (para políticas variáveis)
  // Opcional - nem todas as políticas têm categoria específica
  string policy_category = 6;
}
```

#### Exemplo de Uso (Go)

```go
package main

import (
    "context"
    "fmt"
    "log"

    pb "github.com/lb-conn/rsfn-connect-bacen-bridge/proto/bacen/dict/v2"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    // Conectar ao Bridge
    conn, err := grpc.Dial(
        "bridge.lb-conn.svc.cluster.local:50051",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewRateLimitServiceClient(conn)

    // Chamar GetRateLimitPolicies
    req := &pb.GetRateLimitPoliciesRequest{}

    resp, err := client.GetRateLimitPolicies(context.Background(), req)
    if err != nil {
        log.Fatalf("GetRateLimitPolicies failed: %v", err)
    }

    // Processar resposta
    fmt.Printf("Response Time: %v\n", resp.ResponseTime.AsTime())
    fmt.Printf("PSP Category: %s\n", resp.Category)
    fmt.Printf("Total Policies: %d\n", len(resp.Policies))

    for _, policy := range resp.Policies {
        utilizationPct := float64(policy.Capacity - policy.AvailableTokens) /
                          float64(policy.Capacity) * 100

        fmt.Printf("\nPolicy: %s\n", policy.Name)
        fmt.Printf("  Available: %d / %d fichas\n",
                   policy.AvailableTokens, policy.Capacity)
        fmt.Printf("  Utilization: %.2f%%\n", utilizationPct)
        fmt.Printf("  Refill: %d fichas every %ds\n",
                   policy.RefillTokens, policy.RefillPeriodSec)

        if policy.PolicyCategory != "" {
            fmt.Printf("  Category: %s\n", policy.PolicyCategory)
        }
    }
}
```

#### Mapeamento XML ↔ gRPC (Já Implementado no Bridge)

**Requisição DICT (XML)**:
```xml
GET /api/v2/rate-limit/policies
Authorization: Bearer {JWT_TOKEN}
```

**Resposta DICT (XML)**:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<GetPoliciesResponse>
  <ResponseTime>2025-11-01T10:30:00Z</ResponseTime>
  <Category>A</Category>
  <Policies>
    <Policy>
      <Name>ENTRIES_WRITE</Name>
      <AvailableTokens>35000</AvailableTokens>
      <Capacity>36000</Capacity>
      <RefillTokens>1200</RefillTokens>
      <RefillPeriodSec>60</RefillPeriodSec>
    </Policy>
    <Policy>
      <Name>ENTRIES_READ_PARTICIPANT_ANTISCAN</Name>
      <AvailableTokens>48500</AvailableTokens>
      <Capacity>50000</Capacity>
      <RefillTokens>25000</RefillTokens>
      <RefillPeriodSec>60</RefillPeriodSec>
      <Category>A</Category>
    </Policy>
    <!-- ... mais 22 políticas -->
  </Policies>
</GetPoliciesResponse>
```

**Mapper (Bridge)**:
```go
// Location: internal/mappers/ratelimit/policy_mapper.go

func MapXMLToGRPC(xmlResp *bacen.GetPoliciesResponse) *pb.GetRateLimitPoliciesResponse {
    resp := &pb.GetRateLimitPoliciesResponse{
        ResponseTime: timestamppb.New(xmlResp.ResponseTime),
        Category:     xmlResp.Category,
        Policies:     make([]*pb.RateLimitPolicy, 0, len(xmlResp.Policies)),
    }

    for _, p := range xmlResp.Policies {
        policy := &pb.RateLimitPolicy{
            Name:             p.Name,
            AvailableTokens:  int32(p.AvailableTokens),
            Capacity:         int32(p.Capacity),
            RefillTokens:     int32(p.RefillTokens),
            RefillPeriodSec:  int32(p.RefillPeriodSec),
            PolicyCategory:   p.Category, // Opcional
        }
        resp.Policies = append(resp.Policies, policy)
    }

    return resp
}
```

---

### 2. GetRateLimitPolicy (Get Single Policy)

**Descrição**: Consulta **uma política específica** de rate limit.

#### Proto Definition

```protobuf
service RateLimitService {
  // Consulta uma política específica de rate limit
  rpc GetRateLimitPolicy(GetRateLimitPolicyRequest)
      returns (GetRateLimitPolicyResponse);
}

message GetRateLimitPolicyRequest {
  // Nome da política (ex: "ENTRIES_WRITE")
  string policy_name = 1;
}

message GetRateLimitPolicyResponse {
  // Timestamp da resposta do DICT
  google.protobuf.Timestamp response_time = 1;

  // Categoria do participante (A-H)
  string category = 2;

  // Política consultada
  RateLimitPolicy policy = 3;
}
```

#### Exemplo de Uso (Go)

```go
func GetSpecificPolicy(client pb.RateLimitServiceClient, policyName string) error {
    req := &pb.GetRateLimitPolicyRequest{
        PolicyName: policyName,
    }

    resp, err := client.GetRateLimitPolicy(context.Background(), req)
    if err != nil {
        return fmt.Errorf("failed to get policy %s: %w", policyName, err)
    }

    policy := resp.Policy
    utilizationPct := float64(policy.Capacity - policy.AvailableTokens) /
                      float64(policy.Capacity) * 100

    fmt.Printf("Policy: %s\n", policy.Name)
    fmt.Printf("Response Time: %v\n", resp.ResponseTime.AsTime())
    fmt.Printf("PSP Category: %s\n", resp.Category)
    fmt.Printf("Available: %d / %d fichas (%.2f%% utilization)\n",
               policy.AvailableTokens, policy.Capacity, utilizationPct)

    // Calcular ETA para recuperação total
    if policy.AvailableTokens < policy.Capacity {
        tokensNeeded := policy.Capacity - policy.AvailableTokens
        periodsNeeded := (tokensNeeded + policy.RefillTokens - 1) / policy.RefillTokens
        secondsToRecovery := periodsNeeded * policy.RefillPeriodSec

        fmt.Printf("ETA to full recovery: %d seconds (~%d minutes)\n",
                   secondsToRecovery, secondsToRecovery/60)
    }

    return nil
}
```

#### Mapeamento XML ↔ gRPC

**Requisição DICT (XML)**:
```xml
GET /api/v2/rate-limit/policies/ENTRIES_WRITE
Authorization: Bearer {JWT_TOKEN}
```

**Resposta DICT (XML)**:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<GetPolicyResponse>
  <ResponseTime>2025-11-01T10:31:05Z</ResponseTime>
  <Category>A</Category>
  <Policy>
    <Name>ENTRIES_WRITE</Name>
    <AvailableTokens>34800</AvailableTokens>
    <Capacity>36000</Capacity>
    <RefillTokens>1200</RefillTokens>
    <RefillPeriodSec>60</RefillPeriodSec>
  </Policy>
</GetPolicyResponse>
```

---

## 🔐 Autenticação e Segurança

### mTLS Configuration

O Bridge já está configurado com **mTLS (mutual TLS)** para comunicação segura com o DICT BACEN.

#### Configuração (já implementada no Bridge)

```go
// Location: internal/grpc/server/config.go

type TLSConfig struct {
    // Certificado do cliente (PSP)
    ClientCertFile string
    ClientKeyFile  string

    // CA do DICT BACEN
    CAFile string

    // Server Name Indication
    ServerName string // "dict.pi.rsfn.net.br"
}
```

#### AWS Secrets Manager Integration

**Secrets armazenados**:

```json
{
  "SecretId": "lb-conn/dict/bridge/mtls",
  "SecretString": {
    "client_cert": "-----BEGIN CERTIFICATE-----\nMIIE...",
    "client_key": "-----BEGIN PRIVATE KEY-----\nMIIE...",
    "ca_cert": "-----BEGIN CERTIFICATE-----\nMIIE...",
    "server_name": "dict.pi.rsfn.net.br"
  }
}
```

**Como acessar secrets (exemplo)**:

```go
package secrets

import (
    "context"
    "encoding/json"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type BridgeTLSSecrets struct {
    ClientCert string `json:"client_cert"`
    ClientKey  string `json:"client_key"`
    CACert     string `json:"ca_cert"`
    ServerName string `json:"server_name"`
}

func LoadBridgeTLSSecrets(ctx context.Context) (*BridgeTLSSecrets, error) {
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return nil, err
    }

    client := secretsmanager.NewFromConfig(cfg)

    input := &secretsmanager.GetSecretValueInput{
        SecretId: aws.String("lb-conn/dict/bridge/mtls"),
    }

    result, err := client.GetSecretValue(ctx, input)
    if err != nil {
        return nil, err
    }

    var secrets BridgeTLSSecrets
    if err := json.Unmarshal([]byte(*result.SecretString), &secrets); err != nil {
        return nil, err
    }

    return &secrets, nil
}
```

---

## 🏗️ Integração com Temporal Activities

### Activity: GetPoliciesActivity

```go
// Location: apps/orchestration-worker/infrastructure/temporal/activities/ratelimit/get_policies_activity.go

package ratelimit

import (
    "context"
    "fmt"

    pb "github.com/lb-conn/rsfn-connect-bacen-bridge/proto/bacen/dict/v2"
    "go.temporal.io/sdk/activity"
)

type GetPoliciesActivity struct {
    bridgeClient pb.RateLimitServiceClient
}

func NewGetPoliciesActivity(client pb.RateLimitServiceClient) *GetPoliciesActivity {
    return &GetPoliciesActivity{
        bridgeClient: client,
    }
}

type GetPoliciesResult struct {
    ResponseTime   string                 `json:"response_time"`    // ISO8601
    Category       string                 `json:"category"`         // A-H
    Policies       []PolicySnapshot       `json:"policies"`
}

type PolicySnapshot struct {
    Name            string  `json:"name"`
    AvailableTokens int32   `json:"available_tokens"`
    Capacity        int32   `json:"capacity"`
    RefillTokens    int32   `json:"refill_tokens"`
    RefillPeriodSec int32   `json:"refill_period_sec"`
    UtilizationPct  float64 `json:"utilization_pct"`
    PolicyCategory  string  `json:"policy_category,omitempty"`
}

func (a *GetPoliciesActivity) Execute(ctx context.Context) (*GetPoliciesResult, error) {
    logger := activity.GetLogger(ctx)
    logger.Info("GetPoliciesActivity started")

    // Chamar Bridge gRPC
    req := &pb.GetRateLimitPoliciesRequest{}

    resp, err := a.bridgeClient.GetRateLimitPolicies(ctx, req)
    if err != nil {
        logger.Error("Bridge gRPC call failed", "error", err)
        return nil, fmt.Errorf("failed to get policies from bridge: %w", err)
    }

    // Converter para resultado
    result := &GetPoliciesResult{
        ResponseTime: resp.ResponseTime.AsTime().Format(time.RFC3339),
        Category:     resp.Category,
        Policies:     make([]PolicySnapshot, 0, len(resp.Policies)),
    }

    for _, p := range resp.Policies {
        utilization := float64(p.Capacity - p.AvailableTokens) /
                       float64(p.Capacity) * 100

        snapshot := PolicySnapshot{
            Name:            p.Name,
            AvailableTokens: p.AvailableTokens,
            Capacity:        p.Capacity,
            RefillTokens:    p.RefillTokens,
            RefillPeriodSec: p.RefillPeriodSec,
            UtilizationPct:  utilization,
            PolicyCategory:  p.PolicyCategory,
        }

        result.Policies = append(result.Policies, snapshot)

        logger.Info("Policy snapshot",
            "policy", p.Name,
            "available", p.AvailableTokens,
            "capacity", p.Capacity,
            "utilization", fmt.Sprintf("%.2f%%", utilization),
        )
    }

    logger.Info("GetPoliciesActivity completed",
        "category", result.Category,
        "total_policies", len(result.Policies),
    )

    return result, nil
}
```

---

## 📊 Error Handling

### Tipos de Erros (gRPC Status Codes)

```go
package ratelimit

import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// HandleBridgeError converte erros gRPC do Bridge em erros de domínio
func HandleBridgeError(err error) error {
    st, ok := status.FromError(err)
    if !ok {
        return fmt.Errorf("bridge unknown error: %w", err)
    }

    switch st.Code() {
    case codes.Unauthenticated:
        // mTLS certificate inválido ou expirado
        return &BridgeAuthError{
            Message: "bridge authentication failed - check mTLS certificates",
            Cause:   err,
        }

    case codes.PermissionDenied:
        // PSP não tem permissão para consultar rate limit
        return &BridgePermissionError{
            Message: "bridge permission denied - check PSP authorization",
            Cause:   err,
        }

    case codes.Unavailable:
        // Bridge ou DICT indisponível
        return &BridgeUnavailableError{
            Message: "bridge or DICT unavailable - retry later",
            Cause:   err,
            Retryable: true,
        }

    case codes.DeadlineExceeded:
        // Timeout na chamada
        return &BridgeTimeoutError{
            Message: "bridge request timeout",
            Cause:   err,
            Retryable: true,
        }

    case codes.NotFound:
        // Política não existe (apenas para GetRateLimitPolicy)
        return &PolicyNotFoundError{
            Message: "policy not found in DICT",
            Cause:   err,
        }

    case codes.Internal:
        // Erro interno do Bridge/DICT
        return &BridgeInternalError{
            Message: st.Message(),
            Cause:   err,
            Retryable: false,
        }

    default:
        return fmt.Errorf("bridge error: %s (code: %s)", st.Message(), st.Code())
    }
}
```

### Retry Policy (Temporal Activity)

```go
// Location: apps/orchestration-worker/infrastructure/temporal/workflows/ratelimit/monitor_policies_workflow.go

func (w *MonitorPoliciesWorkflow) Execute(ctx workflow.Context) error {
    // Retry policy para GetPoliciesActivity
    retryPolicy := &temporal.RetryPolicy{
        InitialInterval:    time.Second * 2,
        BackoffCoefficient: 2.0,
        MaximumInterval:    time.Minute * 1,
        MaximumAttempts:    5,

        // Não retryar erros de autenticação/permissão
        NonRetryableErrorTypes: []string{
            "BridgeAuthError",
            "BridgePermissionError",
        },
    }

    activityOptions := workflow.ActivityOptions{
        StartToCloseTimeout: time.Second * 30,
        RetryPolicy:         retryPolicy,
    }

    ctx = workflow.WithActivityOptions(ctx, activityOptions)

    // Executar activity
    var result *GetPoliciesResult
    err := workflow.ExecuteActivity(ctx, w.getPoliciesActivity.Execute).Get(ctx, &result)
    if err != nil {
        return fmt.Errorf("failed to get policies: %w", err)
    }

    // Continuar workflow...
    return nil
}
```

---

## 🧪 Testing com Mock Bridge

### Mock gRPC Server

```go
// Location: apps/orchestration-worker/tests/mocks/bridge/ratelimit_mock.go

package bridge

import (
    "context"

    pb "github.com/lb-conn/rsfn-connect-bacen-bridge/proto/bacen/dict/v2"
    "google.golang.org/grpc"
    "google.golang.org/protobuf/types/known/timestamppb"
)

type MockRateLimitServiceClient struct {
    // Dados mockados
    Policies []MockPolicy
    Category string
}

type MockPolicy struct {
    Name            string
    AvailableTokens int32
    Capacity        int32
    RefillTokens    int32
    RefillPeriodSec int32
}

func (m *MockRateLimitServiceClient) GetRateLimitPolicies(
    ctx context.Context,
    req *pb.GetRateLimitPoliciesRequest,
    opts ...grpc.CallOption,
) (*pb.GetRateLimitPoliciesResponse, error) {

    resp := &pb.GetRateLimitPoliciesResponse{
        ResponseTime: timestamppb.Now(),
        Category:     m.Category,
        Policies:     make([]*pb.RateLimitPolicy, 0, len(m.Policies)),
    }

    for _, p := range m.Policies {
        policy := &pb.RateLimitPolicy{
            Name:            p.Name,
            AvailableTokens: p.AvailableTokens,
            Capacity:        p.Capacity,
            RefillTokens:    p.RefillTokens,
            RefillPeriodSec: p.RefillPeriodSec,
        }
        resp.Policies = append(resp.Policies, policy)
    }

    return resp, nil
}

func (m *MockRateLimitServiceClient) GetRateLimitPolicy(
    ctx context.Context,
    req *pb.GetRateLimitPolicyRequest,
    opts ...grpc.CallOption,
) (*pb.GetRateLimitPolicyResponse, error) {

    for _, p := range m.Policies {
        if p.Name == req.PolicyName {
            return &pb.GetRateLimitPolicyResponse{
                ResponseTime: timestamppb.Now(),
                Category:     m.Category,
                Policy: &pb.RateLimitPolicy{
                    Name:            p.Name,
                    AvailableTokens: p.AvailableTokens,
                    Capacity:        p.Capacity,
                    RefillTokens:    p.RefillTokens,
                    RefillPeriodSec: p.RefillPeriodSec,
                },
            }, nil
        }
    }

    return nil, status.Errorf(codes.NotFound, "policy %s not found", req.PolicyName)
}
```

### Teste de Integração

```go
// Location: apps/orchestration-worker/tests/integration/ratelimit/get_policies_test.go

package ratelimit_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "your-module/tests/mocks/bridge"
    "your-module/infrastructure/temporal/activities/ratelimit"
)

func TestGetPoliciesActivity_Success(t *testing.T) {
    // Setup mock Bridge
    mockBridge := &bridge.MockRateLimitServiceClient{
        Category: "A",
        Policies: []bridge.MockPolicy{
            {
                Name:            "ENTRIES_WRITE",
                AvailableTokens: 35000,
                Capacity:        36000,
                RefillTokens:    1200,
                RefillPeriodSec: 60,
            },
            {
                Name:            "CLAIMS_WRITE",
                AvailableTokens: 30000,
                Capacity:        36000,
                RefillTokens:    1200,
                RefillPeriodSec: 60,
            },
        },
    }

    // Create activity
    activity := ratelimit.NewGetPoliciesActivity(mockBridge)

    // Execute
    ctx := context.Background()
    result, err := activity.Execute(ctx)

    // Assert
    require.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "A", result.Category)
    assert.Len(t, result.Policies, 2)

    // Verificar ENTRIES_WRITE
    entries := result.Policies[0]
    assert.Equal(t, "ENTRIES_WRITE", entries.Name)
    assert.Equal(t, int32(35000), entries.AvailableTokens)
    assert.Equal(t, int32(36000), entries.Capacity)
    assert.InDelta(t, 2.78, entries.UtilizationPct, 0.01) // ~2.78% utilizado

    // Verificar CLAIMS_WRITE
    claims := result.Policies[1]
    assert.Equal(t, "CLAIMS_WRITE", claims.Name)
    assert.InDelta(t, 16.67, claims.UtilizationPct, 0.01) // ~16.67% utilizado
}
```

---

## 📝 Checklist de Integração

- [x] ✅ Proto definitions existem no Bridge
- [x] ✅ Endpoints gRPC implementados
- [x] ✅ Mappers XML ↔ gRPC disponíveis
- [x] ✅ mTLS configuration pronta
- [x] ✅ AWS Secrets Manager definido como solução
- [ ] 🔄 Implementar GetPoliciesActivity no orchestration-worker
- [ ] 🔄 Implementar error handling e retry policies
- [ ] 🔄 Criar testes de integração com mock Bridge
- [ ] 🔄 Documentar secrets no AWS Secrets Manager
- [ ] 🔄 Configurar permissões IAM para acesso a secrets

---

## 🚀 Próximos Passos

### 1. Configurar AWS Secrets Manager (DevOps)

```bash
# Criar secret para mTLS certificates
aws secretsmanager create-secret \
  --name lb-conn/dict/bridge/mtls \
  --description "mTLS certificates for DICT Bridge communication" \
  --secret-string file://mtls-secrets.json

# Criar secret para Bridge endpoint
aws secretsmanager create-secret \
  --name lb-conn/dict/bridge/endpoint \
  --description "Bridge gRPC endpoint configuration" \
  --secret-string '{"host":"bridge.lb-conn.svc.cluster.local","port":"50051"}'
```

### 2. Implementar gRPC Client no Orchestration Worker

```go
// Location: apps/orchestration-worker/infrastructure/grpc/ratelimit/client.go

package ratelimit

import (
    "context"
    "fmt"

    pb "github.com/lb-conn/rsfn-connect-bacen-bridge/proto/bacen/dict/v2"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
)

type BridgeClient struct {
    conn   *grpc.ClientConn
    client pb.RateLimitServiceClient
}

func NewBridgeClient(ctx context.Context, tlsConfig *TLSConfig) (*BridgeClient, error) {
    // Carregar certificates do AWS Secrets Manager
    creds, err := credentials.NewClientTLSFromFile(
        tlsConfig.CAFile,
        tlsConfig.ServerName,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
    }

    // Conectar ao Bridge
    conn, err := grpc.DialContext(
        ctx,
        tlsConfig.Endpoint,
        grpc.WithTransportCredentials(creds),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to dial bridge: %w", err)
    }

    client := pb.NewRateLimitServiceClient(conn)

    return &BridgeClient{
        conn:   conn,
        client: client,
    }, nil
}

func (c *BridgeClient) Close() error {
    return c.conn.Close()
}

func (c *BridgeClient) GetPolicies(ctx context.Context) (*pb.GetRateLimitPoliciesResponse, error) {
    req := &pb.GetRateLimitPoliciesRequest{}
    return c.client.GetRateLimitPolicies(ctx, req)
}
```

### 3. Testar Integração End-to-End

```bash
# Rodar testes de integração
go test ./apps/orchestration-worker/tests/integration/ratelimit/... -v

# Verificar conexão com Bridge (em ambiente de dev)
go run ./cmd/test-bridge-connection/main.go
```

---

**Última Atualização**: 2025-11-01
**Responsável**: Tech Lead
**Status**: ✅ **DOCUMENTAÇÃO COMPLETA** - Pronto para implementação

**Conclusão**: Endpoints do Bridge estão **100% prontos**. Integração pode começar **IMEDIATAMENTE**.
