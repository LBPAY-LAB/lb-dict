# ADR-004 - Bridge DICT Dedicado com Padrões Reutilizáveis

**Status**: ✅ Aprovado
**Data**: 2025-10-24
**Decisores**: José Luís Silva (CTO), NEXUS (Solution Architect)
**Contexto**: Fase de Especificação do Projeto DICT LBPay

---

## Contexto

### Situação Atual (AS-IS)

**Repositório existente**: `connector-dict` (parcialmente implementado)

**Problemas identificados**:

1. ❌ **Implementação incompleta**:
   - Apenas 3-4 endpoints implementados (de 28 totais da API DICT v2.6.1)
   - Falta mTLS configurado
   - Sem assinatura digital XML
   - Sem rate limiting local
   - Sem circuit breaker/retry logic

2. ❌ **Código não reutilizável**:
   - Lógica específica para DICT misturada com HTTP client
   - Sem abstração de mTLS (hardcoded para DICT)
   - Sem abstração de XML signing
   - Difícil usar para outros serviços Bacen (ex: SPI)

3. ❌ **Sem padrões do RSFN Bridge**:
   - `rsfn-connect-bacen-bridge` tem padrões estabelecidos
   - `connector-dict` não segue esses padrões
   - Duplicação de infraestrutura (mTLS, retry, etc.)

### Repositório RSFN Connect Bacen Bridge

**Repositório existente**: `rsfn-connect-bacen-bridge`

**Características**:
- ✅ Abstração de comunicação com Bacen (ISO 20022)
- ✅ mTLS configurável
- ✅ Retry logic com exponential backoff
- ✅ Circuit breaker
- ✅ Observabilidade (OpenTelemetry)
- ✅ Usado para SPI (pagamentos PIX)

**Padrão**: Biblioteca reutilizável de componentes de infraestrutura para comunicação com Bacen.

---

## Decisão

**Refatorar `connector-dict`** para se tornar **Connect DICT**: um bridge dedicado ao DICT Bacen que **reutiliza padrões** do `rsfn-connect-bacen-bridge`.

### Princípios Arquiteturais

1. **Separation of Concerns**: Bridge não tem lógica de negócio (apenas infraestrutura)
2. **Reusability**: Componentes reutilizáveis (mTLS, XML Signer, Retry Logic)
3. **Single Responsibility**: Cada módulo tem uma responsabilidade clara
4. **Interface Segregation**: APIs gRPC para comunicação com Core DICT
5. **Dependency Inversion**: Core DICT depende de interface, não de implementação

### Arquitetura Proposta

```
┌─────────────────────────────────────────────────────────────────────┐
│                      Core DICT (core-dict)                           │
│  • Business Logic (72 RFs)                                           │
│  • Domain, Application, Handlers                                     │
└──────────────────────────┬──────────────────────────────────────────┘
                           │ gRPC
                           ↓
┌─────────────────────────────────────────────────────────────────────┐
│             Connect DICT (rsfn-connect-bacen-bridge)                 │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────┐    │
│  │                    gRPC API Layer                           │    │
│  │  • DICTService (28 RPCs)                                    │    │
│  │  • Conversão gRPC ↔ REST                                    │    │
│  └────────────────────────────────────────────────────────────┘    │
│                           ↓                                          │
│  ┌────────────────────────────────────────────────────────────┐    │
│  │              REST Client Layer (28 APIs)                    │    │
│  │  • entries_api.go      (CRUD de vínculos)                   │    │
│  │  • keys_api.go         (Validação de chaves)                │    │
│  │  • claims_api.go       (Reivindicações)                     │    │
│  │  • reconciliation_api.go (VSync, CID)                       │    │
│  │  • infraction_reports_api.go                                │    │
│  │  • fraud_markers_api.go                                     │    │
│  │  • refunds_api.go                                           │    │
│  │  • statistics_api.go                                        │    │
│  │  • policies_api.go                                          │    │
│  └────────────────────────────────────────────────────────────┘    │
│                           ↓                                          │
│  ┌────────────────────────────────────────────────────────────┐    │
│  │           Shared Components (reutilizáveis)                 │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │    │
│  │  │ mTLS Setup  │  │ XML Signer  │  │ REST Client │        │    │
│  │  │ (P12 cert)  │  │ (SHA-256+   │  │ (pool)      │        │    │
│  │  │             │  │  RSA)       │  │             │        │    │
│  │  └─────────────┘  └─────────────┘  └─────────────┘        │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │    │
│  │  │ Rate Limiter│  │ Circuit     │  │ Retry Logic │        │    │
│  │  │ (local)     │  │ Breaker     │  │ (exp backoff)│       │    │
│  │  └─────────────┘  └─────────────┘  └─────────────┘        │    │
│  │  ┌─────────────┐  ┌─────────────┐                         │    │
│  │  │ Error       │  │ Observability│                        │    │
│  │  │ Handler     │  │ (OTel)      │                         │    │
│  │  └─────────────┘  └─────────────┘                         │    │
│  └────────────────────────────────────────────────────────────┘    │
└──────────────────────────┬───────────────────────────────────────────┘
                           │ mTLS/REST/XML Signed
                           ↓
                  ┌──────────────────┐
                  │   DICT Bacen     │
                  │   (API v2.6.1)   │
                  └──────────────────┘
```

---

## Estrutura do Repositório Connect DICT

```
rsfn-connect-bacen-bridge/
├── cmd/
│   └── dict-bridge/
│       └── main.go                     # Entrypoint (gRPC server)
│
├── pkg/
│   ├── dict/                           # DICT-SPECIFIC
│   │   ├── grpc/                       # gRPC Service (interface para Core DICT)
│   │   │   ├── server.go
│   │   │   ├── dict_service.go         # Implementa DICTService
│   │   │   └── protos/
│   │   │       └── dict.proto          # Definição gRPC
│   │   │
│   │   ├── rest/                       # REST Clients (28 APIs)
│   │   │   ├── entries_api.go          # POST /entries, GET /entries/{Key}, etc
│   │   │   ├── keys_api.go             # POST /keys/check
│   │   │   ├── claims_api.go           # POST /claims, GET /claims/{ClaimId}, etc
│   │   │   ├── reconciliation_api.go   # POST /sync-verifications, etc
│   │   │   ├── infraction_reports_api.go
│   │   │   ├── fraud_markers_api.go
│   │   │   ├── refunds_api.go
│   │   │   ├── statistics_api.go
│   │   │   └── policies_api.go
│   │   │
│   │   ├── models/                     # Structs XML (Request/Response)
│   │   │   ├── entry.go
│   │   │   ├── claim.go
│   │   │   ├── refund.go
│   │   │   └── common.go
│   │   │
│   │   └── config/
│   │       └── dict_config.go          # Configurações DICT (URLs, timeouts)
│   │
│   ├── shared/                         # SHARED COMPONENTS (reutilizáveis)
│   │   ├── mtls/
│   │   │   ├── mtls_config.go          # Configuração de mTLS (genérico)
│   │   │   └── cert_loader.go          # Carrega certificados P12
│   │   │
│   │   ├── xmlsigner/
│   │   │   ├── signer.go               # Assinatura XML (envelopada)
│   │   │   └── validator.go            # Validação de assinatura
│   │   │
│   │   ├── httpclient/
│   │   │   ├── client.go               # HTTP client com pool
│   │   │   ├── middleware/
│   │   │   │   ├── retry.go            # Retry com exponential backoff
│   │   │   │   ├── circuitbreaker.go   # Circuit breaker
│   │   │   │   ├── timeout.go          # Timeout wrapper
│   │   │   │   └── tracing.go          # OpenTelemetry
│   │   │   └── config.go
│   │   │
│   │   ├── ratelimiter/
│   │   │   ├── token_bucket.go         # Token bucket local
│   │   │   └── redis_backend.go        # Backend Redis
│   │   │
│   │   ├── errorhandler/
│   │   │   ├── rfc7807.go              # Parser de RFC 7807 (Problem Details)
│   │   │   └── dict_errors.go          # Mapeamento de erros DICT
│   │   │
│   │   └── observability/
│   │       ├── metrics.go              # Prometheus metrics
│   │       ├── tracer.go               # OpenTelemetry tracer
│   │       └── logger.go               # Structured logging
│   │
│   └── spi/                            # SPI-SPECIFIC (futuro, para SPI)
│       └── ... (estrutura similar ao dict/)
│
├── configs/
│   ├── dict-config.yaml                # Configuração DICT
│   └── spi-config.yaml                 # Configuração SPI (futuro)
│
├── scripts/
│   ├── generate-protos.sh              # Gera código Go from .proto
│   └── test-mtls.sh                    # Testa conectividade mTLS
│
├── tests/
│   ├── unit/
│   ├── integration/
│   └── mocks/
│       └── dict_simulator/             # Mock do DICT Bacen
│
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md
```

---

## Componentes Detalhados

### 1. gRPC API Layer (Interface para Core DICT)

**Motivo**: Core DICT não deve fazer chamadas HTTP diretamente. gRPC é mais eficiente e type-safe.

#### dict.proto

```protobuf
syntax = "proto3";

package lbpay.dict.v1;

option go_package = "github.com/lbpay/rsfn-connect-bacen-bridge/pkg/dict/grpc/protos;dictpb";

// DICTService define todas as operações do DICT Bacen
service DICTService {
  // Bloco 1 - CRUD de Vínculos
  rpc CreateEntry(CreateEntryRequest) returns (CreateEntryResponse);
  rpc GetEntry(GetEntryRequest) returns (GetEntryResponse);
  rpc UpdateEntry(UpdateEntryRequest) returns (UpdateEntryResponse);
  rpc DeleteEntry(DeleteEntryRequest) returns (DeleteEntryResponse);

  // Bloco 1 - Validação de Chaves
  rpc CheckKeys(CheckKeysRequest) returns (CheckKeysResponse);

  // Bloco 2 - Reivindicações
  rpc CreateClaim(CreateClaimRequest) returns (CreateClaimResponse);
  rpc GetClaim(GetClaimRequest) returns (GetClaimResponse);
  rpc ListClaims(ListClaimsRequest) returns (ListClaimsResponse);
  rpc AcknowledgeClaim(AcknowledgeClaimRequest) returns (AcknowledgeClaimResponse);
  rpc ConfirmClaim(ConfirmClaimRequest) returns (ConfirmClaimResponse);
  rpc CancelClaim(CancelClaimRequest) returns (CancelClaimResponse);
  rpc CompleteClaim(CompleteClaimRequest) returns (CompleteClaimResponse);

  // Bloco 5 - Reconciliação
  rpc CreateSyncVerification(CreateSyncVerificationRequest) returns (CreateSyncVerificationResponse);
  rpc CreateCidSetFile(CreateCidSetFileRequest) returns (CreateCidSetFileResponse);
  rpc GetCidSetFile(GetCidSetFileRequest) returns (GetCidSetFileResponse);
  rpc ListCidSetEvents(ListCidSetEventsRequest) returns (ListCidSetEventsResponse);
  rpc GetEntryByCid(GetEntryByCidRequest) returns (GetEntryByCidResponse);

  // Bloco 4 - Notificação de Infração
  rpc CreateInfractionReport(CreateInfractionReportRequest) returns (CreateInfractionReportResponse);
  rpc GetInfractionReport(GetInfractionReportRequest) returns (GetInfractionReportResponse);
  rpc ListInfractionReports(ListInfractionReportsRequest) returns (ListInfractionReportsResponse);
  rpc AcknowledgeInfractionReport(AcknowledgeInfractionReportRequest) returns (AcknowledgeInfractionReportResponse);
  rpc CancelInfractionReport(CancelInfractionReportRequest) returns (CancelInfractionReportResponse);
  rpc CloseInfractionReport(CloseInfractionReportRequest) returns (CloseInfractionReportResponse);

  // Bloco 4/5 - Antifraude
  rpc CreateFraudMarker(CreateFraudMarkerRequest) returns (CreateFraudMarkerResponse);
  rpc GetFraudMarker(GetFraudMarkerRequest) returns (GetFraudMarkerResponse);
  rpc ListFrauds(ListFraudsRequest) returns (ListFraudsResponse);
  rpc CancelFraudMarker(CancelFraudMarkerRequest) returns (CancelFraudMarkerResponse);
  rpc GetEntryStatistics(GetEntryStatisticsRequest) returns (GetEntryStatisticsResponse);
  rpc GetPersonStatistics(GetPersonStatisticsRequest) returns (GetPersonStatisticsResponse);

  // Bloco 4 - Solicitação de Devolução
  rpc CreateRefund(CreateRefundRequest) returns (CreateRefundResponse);
  rpc GetRefund(GetRefundRequest) returns (GetRefundResponse);
  rpc ListRefunds(ListRefundsRequest) returns (ListRefundsResponse);
  rpc CancelRefund(CancelRefundRequest) returns (CancelRefundResponse);
  rpc CloseRefund(CloseRefundRequest) returns (CloseRefundResponse);

  // Bloco 5 - Políticas de Limitação
  rpc GetBucketState(GetBucketStateRequest) returns (GetBucketStateResponse);
  rpc ListBucketStates(ListBucketStatesRequest) returns (ListBucketStatesResponse);
}

// Mensagens (exemplo para CreateEntry)
message CreateEntryRequest {
  string key = 1;
  string key_type = 2;
  Account account = 3;
  Owner owner = 4;
  string reason = 5;
  string request_id = 6;
}

message CreateEntryResponse {
  string key = 1;
  string cid = 2;
  string status = 3;
}

message Account {
  string participant = 1;
  string branch = 2;
  string account_number = 3;
  string account_type = 4;
  string opening_date = 5;
}

message Owner {
  string type = 1;
  string tax_id_number = 2;
  string name = 3;
  string trade_name = 4;
}

// ... (outros 27 RPCs com suas mensagens)
```

#### dict_service.go (Implementação gRPC)

```go
// pkg/dict/grpc/dict_service.go

type DICTServiceServer struct {
    dictpb.UnimplementedDICTServiceServer
    entriesAPI      *rest.EntriesAPI
    keysAPI         *rest.KeysAPI
    claimsAPI       *rest.ClaimsAPI
    // ... outros APIs
}

func (s *DICTServiceServer) CreateEntry(ctx context.Context, req *dictpb.CreateEntryRequest) (*dictpb.CreateEntryResponse, error) {
    // 1. Validação de input (opcional, Core DICT já valida)
    if req.Key == "" {
        return nil, status.Error(codes.InvalidArgument, "key is required")
    }

    // 2. Converter gRPC request → REST XML
    xmlReq := &rest.CreateEntryRequest{
        Entry: rest.Entry{
            Key:     req.Key,
            KeyType: req.KeyType,
            Account: rest.Account{
                Participant:   req.Account.Participant,
                Branch:        req.Account.Branch,
                AccountNumber: req.Account.AccountNumber,
                AccountType:   req.Account.AccountType,
                OpeningDate:   req.Account.OpeningDate,
            },
            Owner: rest.Owner{
                Type:         req.Owner.Type,
                TaxIdNumber:  req.Owner.TaxIdNumber,
                Name:         req.Owner.Name,
                TradeName:    req.Owner.TradeName,
            },
        },
        Reason:    req.Reason,
        RequestId: req.RequestId,
    }

    // 3. Chamar REST API (que lida com mTLS, XML signing, retry, etc.)
    xmlResp, err := s.entriesAPI.CreateEntry(ctx, xmlReq)
    if err != nil {
        // 4. Converter erro DICT → gRPC error
        return nil, handleDICTError(err)
    }

    // 5. Converter REST XML response → gRPC response
    return &dictpb.CreateEntryResponse{
        Key:    xmlResp.Entry.Key,
        Cid:    xmlResp.Cid,
        Status: "CREATED",
    }, nil
}

// ... (outros 27 métodos)
```

---

### 2. REST Client Layer (28 APIs)

**Responsabilidade**: Implementar 28 endpoints REST do DICT Bacen.

#### entries_api.go (Exemplo)

```go
// pkg/dict/rest/entries_api.go

type EntriesAPI struct {
    client     *httpclient.Client      // HTTP client com pool
    xmlSigner  *xmlsigner.Signer       // Assinatura XML
    baseURL    string                  // https://dict.pi.rsfn.net.br:16422/api/v2
    rateLimiter *ratelimiter.TokenBucket
}

func NewEntriesAPI(client *httpclient.Client, signer *xmlsigner.Signer, baseURL string, limiter *ratelimiter.TokenBucket) *EntriesAPI {
    return &EntriesAPI{
        client:      client,
        xmlSigner:   signer,
        baseURL:     baseURL,
        rateLimiter: limiter,
    }
}

func (api *EntriesAPI) CreateEntry(ctx context.Context, req *CreateEntryRequest) (*CreateEntryResponse, error) {
    // 1. Rate limiting local (prevenir 429 do Bacen)
    if !api.rateLimiter.TryAcquire("ENTRIES_WRITE", 1) {
        return nil, ErrRateLimitExceeded
    }

    // 2. Serializar para XML
    xmlBody, err := xml.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal XML: %w", err)
    }

    // 3. Assinar XML (envelopada)
    signedXML, err := api.xmlSigner.SignXML(xmlBody)
    if err != nil {
        return nil, fmt.Errorf("failed to sign XML: %w", err)
    }

    // 4. Criar HTTP request
    url := fmt.Sprintf("%s/entries/", api.baseURL)
    httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(signedXML))
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Content-Type", "application/xml")
    httpReq.Header.Set("Accept-Encoding", "gzip")

    // 5. Executar request (com retry, circuit breaker, tracing - via middleware)
    httpResp, err := api.client.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("HTTP request failed: %w", err)
    }
    defer httpResp.Body.Close()

    // 6. Ler response body
    respBody, err := io.ReadAll(httpResp.Body)
    if err != nil {
        return nil, err
    }

    // 7. Tratar erros HTTP
    if httpResp.StatusCode >= 400 {
        return nil, parseErrorResponse(httpResp.StatusCode, respBody)
    }

    // 8. Validar assinatura da resposta
    if err := api.xmlSigner.ValidateSignature(respBody); err != nil {
        return nil, fmt.Errorf("invalid signature in response: %w", err)
    }

    // 9. Deserializar XML response
    var resp CreateEntryResponse
    if err := xml.Unmarshal(respBody, &resp); err != nil {
        return nil, fmt.Errorf("failed to unmarshal XML response: %w", err)
    }

    return &resp, nil
}

func (api *EntriesAPI) GetEntry(ctx context.Context, key string, taxIdNumber string) (*GetEntryResponse, error) {
    // Similar implementation for GET
    // NOTE: GET não requer assinatura digital no request
}

func (api *EntriesAPI) UpdateEntry(ctx context.Context, key string, req *UpdateEntryRequest) (*UpdateEntryResponse, error) {
    // Similar implementation for POST (update)
}

func (api *EntriesAPI) DeleteEntry(ctx context.Context, key string, req *DeleteEntryRequest) error {
    // Similar implementation for POST (delete)
}
```

**Total**: 9 arquivos (1 por bloco funcional), ~50-100 linhas cada.

---

### 3. Shared Components (Reutilizáveis)

#### 3.1 mTLS Setup

```go
// pkg/shared/mtls/mtls_config.go

type Config struct {
    CertFile string // client.crt
    KeyFile  string // client.key
    CAFile   string // ca.crt
    P12File  string // client.p12 (alternative)
    P12Pass  string // P12 password
}

func NewTLSConfig(cfg Config) (*tls.Config, error) {
    var cert tls.Certificate
    var err error

    if cfg.P12File != "" {
        // Load from P12
        cert, err = loadP12Certificate(cfg.P12File, cfg.P12Pass)
    } else {
        // Load from PEM
        cert, err = tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
    }
    if err != nil {
        return nil, err
    }

    caCert, err := ioutil.ReadFile(cfg.CAFile)
    if err != nil {
        return nil, err
    }

    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)

    return &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
        MinVersion:   tls.VersionTLS12,
    }, nil
}

func loadP12Certificate(p12File, password string) (tls.Certificate, error) {
    p12Data, err := ioutil.ReadFile(p12File)
    if err != nil {
        return tls.Certificate{}, err
    }

    privateKey, cert, caCerts, err := pkcs12.DecodeChain(p12Data, password)
    if err != nil {
        return tls.Certificate{}, err
    }

    // ... convert to tls.Certificate
}
```

**Reutilizável**: Pode ser usado para SPI, outros serviços Bacen.

---

#### 3.2 XML Signer

```go
// pkg/shared/xmlsigner/signer.go

type Signer struct {
    privateKey *rsa.PrivateKey
    cert       *x509.Certificate
}

func NewSigner(certFile, keyFile string) (*Signer, error) {
    // Load certificate and private key
    // ...
}

func (s *Signer) SignXML(xml []byte) ([]byte, error) {
    // 1. Parse XML
    doc := etree.NewDocument()
    if err := doc.ReadFromBytes(xml); err != nil {
        return nil, err
    }

    // 2. Calculate digest (SHA-256)
    hash := sha256.Sum256(xml)

    // 3. Sign digest with RSA private key
    signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hash[:])
    if err != nil {
        return nil, err
    }

    // 4. Create <Signature> element (envelopada)
    sigElem := doc.Root().CreateElement("Signature")
    sigElem.CreateAttr("xmlns", "http://www.w3.org/2000/09/xmldsig#")

    signedInfo := sigElem.CreateElement("SignedInfo")
    signedInfo.CreateElement("CanonicalizationMethod").CreateAttr("Algorithm", "http://www.w3.org/TR/2001/REC-xml-c14n-20010315")
    signedInfo.CreateElement("SignatureMethod").CreateAttr("Algorithm", "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256")

    reference := signedInfo.CreateElement("Reference")
    reference.CreateAttr("URI", "")
    transforms := reference.CreateElement("Transforms")
    transforms.CreateElement("Transform").CreateAttr("Algorithm", "http://www.w3.org/2000/09/xmldsig#enveloped-signature")
    reference.CreateElement("DigestMethod").CreateAttr("Algorithm", "http://www.w3.org/2001/04/xmlenc#sha256")
    reference.CreateElement("DigestValue").SetText(base64.StdEncoding.EncodeToString(hash[:]))

    sigElem.CreateElement("SignatureValue").SetText(base64.StdEncoding.EncodeToString(signature))

    // 5. Return signed XML
    return doc.WriteToBytes()
}

func (s *Signer) ValidateSignature(xml []byte) error {
    // 1. Parse XML
    // 2. Extract <Signature>
    // 3. Calculate digest
    // 4. Verify signature with public key
    // ...
}
```

**Reutilizável**: Pode ser usado para qualquer API que exige XML Signature.

---

#### 3.3 HTTP Client with Middleware

```go
// pkg/shared/httpclient/client.go

type Client struct {
    httpClient *http.Client
    middlewares []Middleware
}

type Middleware func(http.RoundTripper) http.RoundTripper

func NewClient(tlsConfig *tls.Config, middlewares ...Middleware) *Client {
    transport := &http.Transport{
        TLSClientConfig:     tlsConfig,
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        DisableKeepAlives:   false,
    }

    // Wrap transport with middlewares (chain of responsibility)
    var rt http.RoundTripper = transport
    for i := len(middlewares) - 1; i >= 0; i-- {
        rt = middlewares[i](rt)
    }

    return &Client{
        httpClient: &http.Client{
            Transport: rt,
            Timeout:   10 * time.Second,
        },
    }
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
    return c.httpClient.Do(req)
}
```

**Middlewares**:

```go
// pkg/shared/httpclient/middleware/retry.go

func RetryMiddleware(maxRetries int, backoff time.Duration) Middleware {
    return func(next http.RoundTripper) http.RoundTripper {
        return &retryTransport{
            next:       next,
            maxRetries: maxRetries,
            backoff:    backoff,
        }
    }
}

type retryTransport struct {
    next       http.RoundTripper
    maxRetries int
    backoff    time.Duration
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    for i := 0; i < t.maxRetries; i++ {
        resp, err := t.next.RoundTrip(req)

        if err == nil && resp.StatusCode < 500 && resp.StatusCode != 429 {
            return resp, nil // Sucesso ou erro não retryable
        }

        if i < t.maxRetries-1 {
            time.Sleep(t.backoff * time.Duration(1<<i)) // Exponential backoff
        }
    }

    return t.next.RoundTrip(req) // Última tentativa
}
```

```go
// pkg/shared/httpclient/middleware/circuitbreaker.go

func CircuitBreakerMiddleware(threshold int, timeout time.Duration) Middleware {
    cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
        Name:        "dict-bacen",
        MaxRequests: 3,
        Timeout:     timeout,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= uint32(threshold) && failureRatio >= 0.5
        },
    })

    return func(next http.RoundTripper) http.RoundTripper {
        return &circuitBreakerTransport{next: next, cb: cb}
    }
}
```

```go
// pkg/shared/httpclient/middleware/tracing.go

func TracingMiddleware() Middleware {
    return func(next http.RoundTripper) http.RoundTripper {
        return otelhttp.NewTransport(next) // OpenTelemetry auto-instrumentation
    }
}
```

**Composição**:
```go
client := httpclient.NewClient(
    tlsConfig,
    middleware.TracingMiddleware(),
    middleware.RetryMiddleware(3, 1*time.Second),
    middleware.CircuitBreakerMiddleware(10, 30*time.Second),
    middleware.TimeoutMiddleware(5*time.Second),
)
```

**Reutilizável**: Mesma stack de middlewares para SPI, outros serviços.

---

#### 3.4 Rate Limiter (Token Bucket)

```go
// pkg/shared/ratelimiter/token_bucket.go

type TokenBucket struct {
    redis      *redis.Client
    policies   map[string]Policy
}

type Policy struct {
    Name       string
    Capacity   int
    RefillRate int // tokens/min
}

func (tb *TokenBucket) TryAcquire(policyName string, tokens int) bool {
    policy, ok := tb.policies[policyName]
    if !ok {
        return true // Policy não encontrada, allow
    }

    key := fmt.Sprintf("rate-limit:%s", policyName)

    // Lua script (atomic operation)
    script := `
        local key = KEYS[1]
        local capacity = tonumber(ARGV[1])
        local refillRate = tonumber(ARGV[2])
        local tokens = tonumber(ARGV[3])
        local now = tonumber(ARGV[4])

        local current = redis.call('HGETALL', key)
        local available = capacity
        local lastRefill = now

        if #current > 0 then
            available = tonumber(current[2])
            lastRefill = tonumber(current[4])
        end

        local elapsed = now - lastRefill
        local tokensToAdd = math.floor(elapsed / 60 * refillRate)
        available = math.min(available + tokensToAdd, capacity)

        if available >= tokens then
            available = available - tokens
            redis.call('HMSET', key, 'tokens', available, 'lastRefill', now)
            redis.call('EXPIRE', key, 120)
            return 1
        else
            return 0
        end
    `

    result, err := tb.redis.Eval(context.Background(), script, []string{key},
        policy.Capacity, policy.RefillRate, tokens, time.Now().Unix()).Int()

    return err == nil && result == 1
}
```

**Reutilizável**: Pode ser usado para qualquer API com rate limiting.

---

#### 3.5 Error Handler (RFC 7807)

```go
// pkg/shared/errorhandler/rfc7807.go

type ProblemDetails struct {
    Type     string `xml:"type"`
    Title    string `xml:"title"`
    Status   int    `xml:"status"`
    Detail   string `xml:"detail"`
    Instance string `xml:"instance,omitempty"`
}

func ParseError(statusCode int, body []byte) error {
    var problem ProblemDetails
    if err := xml.Unmarshal(body, &problem); err != nil {
        return fmt.Errorf("HTTP %d: %s", statusCode, string(body))
    }

    return &DICTError{
        Type:       problem.Type,
        Title:      problem.Title,
        StatusCode: problem.Status,
        Detail:     problem.Detail,
    }
}

type DICTError struct {
    Type       string
    Title      string
    StatusCode int
    Detail     string
}

func (e *DICTError) Error() string {
    return fmt.Sprintf("[%s] %s: %s", e.Type, e.Title, e.Detail)
}

func (e *DICTError) IsRetryable() bool {
    return e.StatusCode == 429 || e.StatusCode >= 500
}
```

**Reutilizável**: RFC 7807 é padrão, funciona para qualquer API.

---

### 4. Observabilidade

```go
// pkg/shared/observability/metrics.go

var (
    // HTTP request duration
    httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "dict_bridge_http_request_duration_seconds",
        Help:    "HTTP request duration in seconds",
        Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
    }, []string{"endpoint", "method", "status"})

    // HTTP request total
    httpRequestTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "dict_bridge_http_requests_total",
        Help: "Total number of HTTP requests",
    }, []string{"endpoint", "method", "status"})

    // Rate limited requests
    rateLimitedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "dict_bridge_rate_limited_total",
        Help: "Total number of rate limited requests",
    }, []string{"policy"})
)
```

**Reutilizável**: Mesmas métricas para SPI, outros serviços.

---

## Consequências

### ✅ Positivas

1. **Separation of Concerns**:
   - ✅ Core DICT: lógica de negócio (72 RFs)
   - ✅ Connect DICT: infraestrutura (mTLS, XML, HTTP)

2. **Reutilizabilidade**:
   - ✅ Componentes shared podem ser usados para SPI
   - ✅ mTLS setup genérico
   - ✅ XML Signer genérico
   - ✅ HTTP client com middlewares reutilizáveis

3. **Testabilidade**:
   - ✅ Core DICT testa lógica de negócio (mock do Connect DICT via gRPC)
   - ✅ Connect DICT testa infraestrutura (mock do DICT Bacen)

4. **Maintainability**:
   - ✅ Mudanças em infraestrutura (ex: migrar de REST para GraphQL) não afetam Core DICT
   - ✅ Mudanças em lógica de negócio não afetam Connect DICT

5. **Observabilidade**:
   - ✅ Métricas, traces, logs padronizados
   - ✅ Fácil debug (gRPC requests visíveis, HTTP requests visíveis)

6. **Performance**:
   - ✅ gRPC é mais eficiente que REST (binary protocol)
   - ✅ Connection pooling reutiliza conexões mTLS
   - ✅ Rate limiting local previne 429

### ⚠️ Negativas (e Mitigações)

#### 1. Overhead de gRPC

**Problema**: Adiciona camada gRPC entre Core DICT e HTTP client.

**Latência adicionada**: ~1-5ms (negligível comparado a 200-500ms de rede)

**Mitigação**:
- ✅ gRPC é local (mesmo cluster Kubernetes)
- ✅ Overhead compensado por benefícios (type-safety, streaming, etc.)

#### 2. Duplicação de Structs

**Problema**: gRPC messages duplicam structs XML.

**Mitigação**:
- ✅ **Geração automática**: protoc gera código Go from .proto
- ✅ **Mapping automático**: Usar biblioteca de mapping (ex: copier)

---

## Alternativas Consideradas

### Alternativa 1: Core DICT faz chamadas HTTP diretamente

**Prós**:
- ✅ Sem camada gRPC (mais simples)

**Contras**:
- ❌ Viola Clean Architecture (Domain depende de infraestrutura)
- ❌ Difícil testar (mock de HTTP client complexo)
- ❌ Acoplamento alto

**Decisão**: ❌ Rejeitada.

---

### Alternativa 2: REST API (em vez de gRPC)

**Prós**:
- ✅ Mais familiar

**Contras**:
- ❌ Menos eficiente (JSON > Protobuf)
- ❌ Sem type-safety compile-time
- ❌ Sem streaming (necessário para CID events)

**Decisão**: ❌ Rejeitada - gRPC é superior para comunicação interna.

---

## Decisão Final

✅ **APROVADA**: Refatorar `connector-dict` para **Connect DICT** com **padrões reutilizáveis** do `rsfn-connect-bacen-bridge`.

### Justificativa

1. ✅ **Separation of Concerns**: Core DICT (negócio) vs Connect DICT (infraestrutura)
2. ✅ **Reutilizabilidade**: Componentes shared para SPI, outros serviços Bacen
3. ✅ **28 APIs completas** (vs 3-4 atuais)
4. ✅ **mTLS + XML Signing + Retry + Circuit Breaker** completos
5. ✅ **Testabilidade máxima** (mock via gRPC)
6. ✅ **Observabilidade** (métricas, traces, logs)

---

## Implementação

### Fase 1: Estrutura Base (Semana 1)

```bash
# 1. Criar estrutura de pastas
mkdir -p pkg/dict/{grpc,rest,models,config}
mkdir -p pkg/shared/{mtls,xmlsigner,httpclient,ratelimiter,errorhandler,observability}

# 2. Definir .proto
# pkg/dict/grpc/protos/dict.proto

# 3. Gerar código Go
make generate-protos
```

### Fase 2: Shared Components (Semanas 2-3)

```bash
# Implementar componentes reutilizáveis (order matters):
# 1. mtls/mtls_config.go
# 2. xmlsigner/signer.go
# 3. httpclient/client.go + middlewares
# 4. ratelimiter/token_bucket.go
# 5. errorhandler/rfc7807.go
# 6. observability/metrics.go
```

### Fase 3: REST APIs (Semanas 4-7)

```bash
# Implementar 28 REST APIs (7 endpoints/semana):
# Semana 4: entries_api.go, keys_api.go
# Semana 5: claims_api.go, reconciliation_api.go
# Semana 6: infraction_reports_api.go, fraud_markers_api.go
# Semana 7: refunds_api.go, statistics_api.go, policies_api.go
```

### Fase 4: gRPC Service (Semana 8)

```bash
# Implementar gRPC service (wrapper dos REST APIs):
# pkg/dict/grpc/dict_service.go
```

### Fase 5: Testes (Semanas 9-10)

```bash
# Unit tests (shared components)
make test-unit

# Integration tests (com simulador DICT)
make test-integration

# E2E tests (ambiente Homologação Bacen)
make test-e2e
```

---

## Referências

1. **gRPC Go Tutorial**
   https://grpc.io/docs/languages/go/quickstart/

2. **XML Digital Signature**
   https://www.w3.org/TR/xmldsig-core/

3. **RFC 7807 - Problem Details**
   https://tools.ietf.org/html/rfc7807

4. **Token Bucket Algorithm**
   https://en.wikipedia.org/wiki/Token_bucket

5. **API-001** - Especificação de APIs DICT Bacen
   [Artefatos/04_APIs/API-001_Especificacao_APIs_DICT_Bacen.md](../04_APIs/API-001_Especificacao_APIs_DICT_Bacen.md)

6. **ADR-002** - Consolidação Core DICT
   [ADR-002_Consolidacao_Core_DICT.md](ADR-002_Consolidacao_Core_DICT.md)

7. **ADR-003** - Performance Multi-Camadas
   [ADR-003_Performance_Multi_Camadas.md](ADR-003_Performance_Multi_Camadas.md)

---

**Documento criado por**: NEXUS (AGT-ARC-001) - Solution Architect
**Aprovado por**: José Luís Silva (CTO)
**Data de Aprovação**: 2025-10-24
**Status**: ✅ Aprovado
**Impacto**: 🟠 Alto (refatoração de repositório existente)
