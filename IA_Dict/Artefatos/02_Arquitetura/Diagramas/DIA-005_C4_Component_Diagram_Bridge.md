# DIA-005: C4 Component Diagram - RSFN Bridge

**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: Equipe Arquitetura
**Status**: ✅ Completo

---

## Sumário Executivo

Este documento apresenta o **C4 Component Diagram** (nível 3) do **RSFN Bridge**, detalhando os componentes internos responsáveis pela conversão gRPC → SOAP, assinatura digital XML, e comunicação mTLS com Bacen.

**Objetivo**: Mostrar a estrutura interna do Bridge gRPC API, SOAP Adapter, XML Signer, e como os componentes interagem para adaptar protocolos modernos (gRPC) para o legado do Bacen (SOAP/XML).

**Pré-requisitos**:
- [DIA-001: C4 Context Diagram](./DIA-001_C4_Context_Diagram.md)
- [DIA-002: C4 Container Diagram](./DIA-002_C4_Container_Diagram.md)
- [TEC-002 v3.1: Bridge Specification](../../11_Especificacoes_Tecnicas/TEC-002_Bridge_Specification.md)

---

## 1. Clean Architecture no RSFN Bridge

O RSFN Bridge segue **Adapter Pattern** com foco em conversão de protocolos:

```
┌─────────────────────────────────────────────────────┐
│  API Layer (gRPC)                                    │  ← gRPC Server, Handlers
│  - Bridge gRPC Service, Request/Response Validators │
├─────────────────────────────────────────────────────┤
│  Application Layer (Protocol Conversion)            │  ← gRPC ↔ SOAP conversion
│  - SOAP Adapters, XML Marshaller/Unmarshaller       │
├─────────────────────────────────────────────────────┤
│  Domain Layer (Business Rules)                      │  ← Validation, Circuit Breaker
│  - Bacen Request Validators, Retry Policies         │
├─────────────────────────────────────────────────────┤
│  Infrastructure Layer (External Interfaces)         │  ← HTTP Client, XML Signer, Cache
│  - Bacen HTTP Client (mTLS), XML Signer (Java JNI)  │
└─────────────────────────────────────────────────────┘
```

**Diferença do Core DICT e Connect**:
- ✅ Foco em **conversão de protocolos** (gRPC → SOAP)
- ✅ **Assinatura digital XML** com ICP-Brasil A3
- ✅ **mTLS** com certificados do Bacen
- ✅ **Circuit Breaker** para proteger contra falhas do Bacen

---

## 2. C4 Component Diagram - RSFN Bridge

### 2.1. Diagrama

```mermaid
C4Component
  title Component Diagram - RSFN Bridge (Adapter Pattern)

  Container_Boundary(rsfn_bridge, "RSFN Bridge") {

    Component_Boundary(api_layer, "API Layer (gRPC)") {
      Component(bridge_grpc_server, "Bridge gRPC Server", "Go + gRPC", "Expõe API gRPC para Connect")
      Component(entry_handler, "Entry Handler", "Go", "Handlers: CreateEntry, GetEntry, DeleteEntry")
      Component(claim_handler, "Claim Handler", "Go", "Handlers: CreateClaim, CompleteClaim, CancelClaim")
      Component(portability_handler, "Portability Handler", "Go", "Handlers: ConfirmPortability")
      Component(health_handler, "Health Handler", "Go", "Health checks e readiness")
      Component(request_validator, "Request Validator", "Go", "Valida payloads gRPC (protobuf)")
    }

    Component_Boundary(application_layer, "Application Layer (Protocol Conversion)") {
      Component(soap_adapter, "SOAP Adapter", "Go", "Converte gRPC ↔ SOAP/XML")
      Component(xml_marshaller, "XML Marshaller", "Go + encoding/xml", "Serializa Go structs → XML")
      Component(xml_unmarshaller, "XML Unmarshaller", "Go + encoding/xml", "Deserializa XML → Go structs")
      Component(grpc_to_soap_mapper, "gRPC to SOAP Mapper", "Go", "Mapeia campos gRPC → SOAP")
      Component(soap_to_grpc_mapper, "SOAP to gRPC Mapper", "Go", "Mapeia campos SOAP → gRPC")
      Component(soap_envelope_builder, "SOAP Envelope Builder", "Go", "Constrói SOAP Envelope com headers")
    }

    Component_Boundary(domain_layer, "Domain Layer (Business Rules)") {
      Component(bacen_request_validator, "Bacen Request Validator", "Go", "Valida requests antes de enviar ao Bacen")
      Component(bacen_response_validator, "Bacen Response Validator", "Go", "Valida respostas do Bacen (SOAP faults)")
      Component(circuit_breaker, "Circuit Breaker", "Go + sony/gobreaker", "Protege contra falhas do Bacen")
      Component(retry_policy, "Retry Policy", "Go", "Retry com backoff exponencial (3x)")
      Component(timeout_manager, "Timeout Manager", "Go", "Timeout de 30s por operação")
    }

    Component_Boundary(infrastructure_layer, "Infrastructure Layer") {
      Component(bacen_http_client, "Bacen HTTP Client", "Go + net/http", "Cliente HTTP com mTLS")
      Component(mtls_config, "mTLS Config", "Go + crypto/tls", "Certificados ICP-Brasil A3")
      Component(xml_signer, "XML Signer", "Java 17 + JRE", "Assinatura digital XML (ICP-Brasil)")
      Component(xml_signature_validator, "XML Signature Validator", "Java 17", "Valida assinaturas XML do Bacen")
      Component(jni_bridge, "JNI Bridge", "Go + cgo", "Interface Go ↔ Java (para XML Signer)")
      Component(redis_cache, "Redis Cache Client", "Go + go-redis", "Cache de respostas do Bacen")
      Component(circuit_breaker_store, "Circuit Breaker Store", "Redis", "Persiste estado do circuit breaker")
    }
  }

  Container(temporal_worker, "Temporal Worker", "Go + Temporal SDK", "Chama Bridge via gRPC")
  ContainerDb(bacen_dict, "Bacen DICT", "SOAP/XML API", "API oficial do Banco Central")
  ContainerDb(bridge_cache, "Bridge Cache", "Redis", "Cache + circuit breaker state")
  System_Ext(icp_brasil_ca, "ICP-Brasil CA", "Autoridade certificadora")

  Rel(temporal_worker, bridge_grpc_server, "gRPC mTLS", "CreateEntry, CreateClaim, etc")
  Rel(bridge_grpc_server, request_validator, "Valida payload")

  Rel(bridge_grpc_server, entry_handler, "Route: CreateEntry")
  Rel(entry_handler, soap_adapter, "Convert to SOAP")

  Rel(soap_adapter, grpc_to_soap_mapper, "Map fields")
  Rel(grpc_to_soap_mapper, soap_envelope_builder, "Build envelope")
  Rel(soap_envelope_builder, xml_marshaller, "Serialize to XML")

  Rel(xml_marshaller, xml_signer, "Sign XML", "JNI call")
  Rel(xml_signer, jni_bridge, "Java ↔ Go")

  Rel(xml_marshaller, bacen_request_validator, "Validate request")
  Rel(bacen_request_validator, circuit_breaker, "Check circuit state")

  Rel(circuit_breaker, retry_policy, "Retry logic")
  Rel(retry_policy, timeout_manager, "Apply timeout")

  Rel(timeout_manager, bacen_http_client, "Send SOAP request")
  Rel(bacen_http_client, mtls_config, "Load certificates")
  Rel(bacen_http_client, bacen_dict, "HTTPS mTLS + SOAP", "POST /api/v1/dict/entries")

  Rel(bacen_dict, bacen_http_client, "SOAP Response")
  Rel(bacen_http_client, xml_signature_validator, "Validate signature")
  Rel(xml_signature_validator, xml_unmarshaller, "Deserialize XML")

  Rel(xml_unmarshaller, soap_to_grpc_mapper, "Map to gRPC")
  Rel(soap_to_grpc_mapper, bacen_response_validator, "Validate response")

  Rel(bacen_response_validator, redis_cache, "Cache response")
  Rel(redis_cache, bridge_cache, "SET key EX 30")

  Rel(soap_adapter, entry_handler, "Return gRPC response")
  Rel(entry_handler, bridge_grpc_server, "Send response")
  Rel(bridge_grpc_server, temporal_worker, "gRPC Response")

  Rel(circuit_breaker, circuit_breaker_store, "Persist state")
  Rel(circuit_breaker_store, bridge_cache, "Redis")

  Rel(mtls_config, icp_brasil_ca, "Validate chain")

  UpdateLayoutConfig($c4ShapeInRow="3", $c4BoundaryInRow="1")
```

---

### 2.2. Versão PlantUML (Alternativa)

```plantuml
@startuml
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Component.puml

LAYOUT_WITH_LEGEND()

title Component Diagram - RSFN Bridge

Container_Boundary(bridge, "RSFN Bridge") {

  ' API Layer
  Component(grpc_server, "Bridge gRPC Server", "Go", "gRPC API")
  Component(entry_handler, "Entry Handler", "Go", "CreateEntry, etc")
  Component(claim_handler, "Claim Handler", "Go", "CreateClaim, etc")
  Component(validator, "Request Validator", "Go", "Validate protobuf")

  ' Application Layer
  Component(soap_adapter, "SOAP Adapter", "Go", "gRPC ↔ SOAP")
  Component(xml_marshaller, "XML Marshaller", "Go", "Serialize XML")
  Component(xml_unmarshaller, "XML Unmarshaller", "Go", "Deserialize XML")
  Component(grpc_mapper, "gRPC Mapper", "Go", "Map fields")

  ' Domain Layer
  Component(req_validator, "Request Validator", "Go", "Validate Bacen request")
  Component(resp_validator, "Response Validator", "Go", "Validate Bacen response")
  Component(circuit_breaker, "Circuit Breaker", "Go", "Protect from failures")
  Component(retry, "Retry Policy", "Go", "Backoff exponential")

  ' Infrastructure Layer
  Component(http_client, "Bacen HTTP Client", "Go", "mTLS")
  Component(mtls, "mTLS Config", "Go", "ICP-Brasil A3")
  Component(xml_signer, "XML Signer", "Java 17", "Digital signature")
  Component(jni, "JNI Bridge", "cgo", "Go ↔ Java")
  Component(redis, "Redis Client", "Go", "Cache")
}

Container(temporal, "Temporal Worker", "Go")
ContainerDb(bacen, "Bacen DICT", "SOAP")
ContainerDb(cache, "Redis", "Cache")

Rel(temporal, grpc_server, "gRPC")
Rel(grpc_server, validator, "Validate")
Rel(grpc_server, entry_handler, "Route")
Rel(entry_handler, soap_adapter, "Convert")

Rel(soap_adapter, grpc_mapper, "Map")
Rel(grpc_mapper, xml_marshaller, "Serialize")
Rel(xml_marshaller, xml_signer, "Sign (JNI)")
Rel(xml_signer, jni, "cgo")

Rel(xml_marshaller, req_validator, "Validate")
Rel(req_validator, circuit_breaker, "Check")
Rel(circuit_breaker, retry, "Retry")
Rel(retry, http_client, "Send")
Rel(http_client, mtls, "Load certs")
Rel(http_client, bacen, "HTTPS mTLS")

Rel(bacen, http_client, "Response")
Rel(http_client, xml_unmarshaller, "Deserialize")
Rel(xml_unmarshaller, grpc_mapper, "Map")
Rel(grpc_mapper, resp_validator, "Validate")
Rel(resp_validator, redis, "Cache")
Rel(redis, cache, "SET")

Rel(soap_adapter, entry_handler, "Return")
Rel(entry_handler, grpc_server, "Response")
Rel(grpc_server, temporal, "gRPC")

@enduml
```

---

## 3. Componentes por Camada

### 3.1. API Layer (gRPC)

#### Bridge gRPC Server
- **Responsabilidade**: Expor API gRPC para Connect
- **Tecnologia**: Go + gRPC
- **Porta**: 9091 (gRPC mTLS)
- **Service Definition**:
  ```protobuf
  service BridgeService {
      // Entries
      rpc CreateEntry(CreateEntryRequest) returns (CreateEntryResponse);
      rpc GetEntry(GetEntryRequest) returns (GetEntryResponse);
      rpc DeleteEntry(DeleteEntryRequest) returns (DeleteEntryResponse);

      // Claims
      rpc CreateClaim(CreateClaimRequest) returns (CreateClaimResponse);
      rpc CompleteClaim(CompleteClaimRequest) returns (CompleteClaimResponse);
      rpc CancelClaim(CancelClaimRequest) returns (CancelClaimResponse);

      // Portabilities
      rpc ConfirmPortability(ConfirmPortabilityRequest) returns (ConfirmPortabilityResponse);

      // Health
      rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
  }
  ```
- **mTLS**: Certificados internos (Connect ↔ Bridge)
- **Localização**: `internal/api/grpc/bridge_server.go`

#### Entry Handler
- **Responsabilidade**: Handlers para operações de entries
- **Implementação**:
  ```go
  type EntryHandler struct {
      soapAdapter *SOAPAdapter
      validator   *RequestValidator
  }

  func (eh *EntryHandler) CreateEntry(ctx context.Context, req *pb.CreateEntryRequest) (*pb.CreateEntryResponse, error) {
      // 1. Valida payload gRPC
      if err := eh.validator.ValidateCreateEntry(req); err != nil {
          return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
      }

      // 2. Converte gRPC → SOAP
      soapResp, err := eh.soapAdapter.CreateEntry(ctx, req)
      if err != nil {
          return nil, status.Errorf(codes.Internal, "soap adapter failed: %v", err)
      }

      // 3. Retorna resposta gRPC
      return &pb.CreateEntryResponse{
          BacenEntryId: soapResp.EntryID,
          Status:       soapResp.Status,
      }, nil
  }

  func (eh *EntryHandler) GetEntry(ctx context.Context, req *pb.GetEntryRequest) (*pb.GetEntryResponse, error) {
      // Similar flow
  }

  func (eh *EntryHandler) DeleteEntry(ctx context.Context, req *pb.DeleteEntryRequest) (*pb.DeleteEntryResponse, error) {
      // Similar flow
  }
  ```
- **Localização**: `internal/api/grpc/handlers/entry_handler.go`

#### Claim Handler
- **Responsabilidade**: Handlers para operações de claims
- **Métodos**: `CreateClaim`, `CompleteClaim`, `CancelClaim`
- **Localização**: `internal/api/grpc/handlers/claim_handler.go`

#### Request Validator
- **Responsabilidade**: Validar payloads gRPC antes de processar
- **Implementação**:
  ```go
  type RequestValidator struct{}

  func (rv *RequestValidator) ValidateCreateEntry(req *pb.CreateEntryRequest) error {
      if req.KeyType == "" {
          return errors.New("key_type is required")
      }
      if req.KeyValue == "" {
          return errors.New("key_value is required")
      }
      if req.Account == nil {
          return errors.New("account is required")
      }
      if req.Account.Ispb == "" || len(req.Account.Ispb) != 8 {
          return errors.New("ispb must be 8 digits")
      }
      return nil
  }
  ```
- **Localização**: `internal/api/grpc/validators/request_validator.go`

---

### 3.2. Application Layer (Protocol Conversion)

#### SOAP Adapter
- **Responsabilidade**: Orquestrar conversão gRPC → SOAP → gRPC
- **Fluxo**:
  ```go
  type SOAPAdapter struct {
      grpcToSOAPMapper      *GRPCToSOAPMapper
      soapEnvelopeBuilder   *SOAPEnvelopeBuilder
      xmlMarshaller         *XMLMarshaller
      xmlSigner             *XMLSigner
      bacenHTTPClient       *BacenHTTPClient
      xmlUnmarshaller       *XMLUnmarshaller
      soapToGRPCMapper      *SOAPToGRPCMapper
      bacenRequestValidator *BacenRequestValidator
      circuitBreaker        *CircuitBreaker
  }

  func (sa *SOAPAdapter) CreateEntry(ctx context.Context, req *pb.CreateEntryRequest) (*SOAPCreateEntryResponse, error) {
      // 1. Map gRPC → SOAP struct
      soapReq := sa.grpcToSOAPMapper.MapCreateEntry(req)

      // 2. Build SOAP Envelope
      envelope := sa.soapEnvelopeBuilder.BuildCreateEntryEnvelope(soapReq)

      // 3. Marshal to XML
      xmlBytes, err := sa.xmlMarshaller.Marshal(envelope)
      if err != nil {
          return nil, fmt.Errorf("failed to marshal xml: %w", err)
      }

      // 4. Sign XML (ICP-Brasil A3)
      signedXML, err := sa.xmlSigner.Sign(xmlBytes)
      if err != nil {
          return nil, fmt.Errorf("failed to sign xml: %w", err)
      }

      // 5. Validate request
      if err := sa.bacenRequestValidator.Validate(signedXML); err != nil {
          return nil, fmt.Errorf("invalid bacen request: %w", err)
      }

      // 6. Send to Bacen (with circuit breaker + retry)
      var responseXML []byte
      err = sa.circuitBreaker.Execute(func() error {
          responseXML, err = sa.bacenHTTPClient.Post(ctx, "/api/v1/dict/entries", signedXML)
          return err
      })
      if err != nil {
          return nil, fmt.Errorf("bacen request failed: %w", err)
      }

      // 7. Unmarshal response XML
      var soapResp SOAPCreateEntryResponse
      err = sa.xmlUnmarshaller.Unmarshal(responseXML, &soapResp)
      if err != nil {
          return nil, fmt.Errorf("failed to unmarshal response: %w", err)
      }

      // 8. Map SOAP → gRPC
      return &soapResp, nil
  }
  ```
- **Localização**: `internal/application/adapters/soap_adapter.go`

#### XML Marshaller
- **Responsabilidade**: Serializar Go structs → XML
- **Tecnologia**: Go `encoding/xml`
- **Implementação**:
  ```go
  type XMLMarshaller struct{}

  func (xm *XMLMarshaller) Marshal(v interface{}) ([]byte, error) {
      return xml.MarshalIndent(v, "", "  ")
  }
  ```
- **Localização**: `internal/application/xml/marshaller.go`

#### XML Unmarshaller
- **Responsabilidade**: Deserializar XML → Go structs
- **Implementação**:
  ```go
  type XMLUnmarshaller struct{}

  func (xu *XMLUnmarshaller) Unmarshal(data []byte, v interface{}) error {
      return xml.Unmarshal(data, v)
  }
  ```
- **Localização**: `internal/application/xml/unmarshaller.go`

#### gRPC to SOAP Mapper
- **Responsabilidade**: Mapear campos gRPC → SOAP
- **Exemplo**:
  ```go
  type GRPCToSOAPMapper struct{}

  func (mapper *GRPCToSOAPMapper) MapCreateEntry(req *pb.CreateEntryRequest) *SOAPCreateEntryRequest {
      return &SOAPCreateEntryRequest{
          Key: SOAPKey{
              Type:  req.KeyType,
              Value: req.KeyValue,
          },
          Account: SOAPAccount{
              ISPB:          req.Account.Ispb,
              AccountNumber: req.Account.AccountNumber,
              Branch:        req.Account.Branch,
              AccountType:   req.Account.AccountType,
          },
      }
  }
  ```
- **Localização**: `internal/application/mappers/grpc_to_soap_mapper.go`

#### SOAP to gRPC Mapper
- **Responsabilidade**: Mapear campos SOAP → gRPC
- **Exemplo**:
  ```go
  type SOAPToGRPCMapper struct{}

  func (mapper *SOAPToGRPCMapper) MapCreateEntryResponse(soapResp *SOAPCreateEntryResponse) *pb.CreateEntryResponse {
      return &pb.CreateEntryResponse{
          BacenEntryId: soapResp.EntryID,
          Status:       soapResp.Status,
      }
  }
  ```
- **Localização**: `internal/application/mappers/soap_to_grpc_mapper.go`

#### SOAP Envelope Builder
- **Responsabilidade**: Construir SOAP Envelope com headers
- **Estrutura**:
  ```go
  type SOAPEnvelopeBuilder struct{}

  func (seb *SOAPEnvelopeBuilder) BuildCreateEntryEnvelope(req *SOAPCreateEntryRequest) *SOAPEnvelope {
      return &SOAPEnvelope{
          XMLName: xml.Name{Space: "http://schemas.xmlsoap.org/soap/envelope/", Local: "Envelope"},
          Header: &SOAPHeader{
              Authentication: &Authentication{
                  Certificate: "ICP-Brasil A3 Certificate",
              },
          },
          Body: &SOAPBody{
              CreateEntryRequest: req,
          },
      }
  }
  ```
- **Localização**: `internal/application/soap/envelope_builder.go`

---

### 3.3. Domain Layer (Business Rules)

#### Bacen Request Validator
- **Responsabilidade**: Validar requests antes de enviar ao Bacen
- **Regras**:
  ```go
  type BacenRequestValidator struct{}

  func (brv *BacenRequestValidator) Validate(xmlBytes []byte) error {
      // 1. Verifica se XML está bem formado
      var envelope SOAPEnvelope
      err := xml.Unmarshal(xmlBytes, &envelope)
      if err != nil {
          return fmt.Errorf("invalid xml: %w", err)
      }

      // 2. Verifica se assinatura digital está presente
      if !strings.Contains(string(xmlBytes), "<ds:Signature") {
          return errors.New("xml signature missing")
      }

      // 3. Valida campos obrigatórios
      if envelope.Body.CreateEntryRequest != nil {
          req := envelope.Body.CreateEntryRequest
          if req.Key.Type == "" || req.Key.Value == "" {
              return errors.New("key type/value required")
          }
          if req.Account.ISPB == "" || len(req.Account.ISPB) != 8 {
              return errors.New("ispb must be 8 digits")
          }
      }

      return nil
  }
  ```
- **Localização**: `internal/domain/validators/bacen_request_validator.go`

#### Bacen Response Validator
- **Responsabilidade**: Validar respostas do Bacen (SOAP faults)
- **Implementação**:
  ```go
  type BacenResponseValidator struct{}

  func (brv *BacenResponseValidator) Validate(xmlBytes []byte) error {
      // 1. Verifica se é SOAP Fault
      if strings.Contains(string(xmlBytes), "<soap:Fault>") {
          var fault SOAPFault
          xml.Unmarshal(xmlBytes, &fault)
          return fmt.Errorf("bacen soap fault: %s - %s", fault.FaultCode, fault.FaultString)
      }

      // 2. Verifica assinatura digital do Bacen
      // (delega para XML Signature Validator)

      return nil
  }
  ```
- **Localização**: `internal/domain/validators/bacen_response_validator.go`

#### Circuit Breaker
- **Responsabilidade**: Proteger contra falhas do Bacen
- **Tecnologia**: `sony/gobreaker`
- **Configuração**:
  ```go
  type CircuitBreaker struct {
      breaker *gobreaker.CircuitBreaker
  }

  func NewCircuitBreaker() *CircuitBreaker {
      settings := gobreaker.Settings{
          Name:        "bacen-http-client",
          MaxRequests: 3,
          Interval:    10 * time.Second,
          Timeout:     60 * time.Second,
          ReadyToTrip: func(counts gobreaker.Counts) bool {
              failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
              return counts.Requests >= 5 && failureRatio >= 0.6
          },
      }

      return &CircuitBreaker{
          breaker: gobreaker.NewCircuitBreaker(settings),
      }
  }

  func (cb *CircuitBreaker) Execute(fn func() error) error {
      _, err := cb.breaker.Execute(func() (interface{}, error) {
          err := fn()
          return nil, err
      })
      return err
  }
  ```
- **Estados**: Closed (normal) → Open (falhas) → Half-Open (teste) → Closed
- **Localização**: `internal/domain/circuit_breaker/circuit_breaker.go`

#### Retry Policy
- **Responsabilidade**: Retry com backoff exponencial
- **Configuração**:
  ```go
  type RetryPolicy struct {
      maxAttempts int
      initialDelay time.Duration
      maxDelay time.Duration
      multiplier float64
  }

  func NewRetryPolicy() *RetryPolicy {
      return &RetryPolicy{
          maxAttempts:  3,
          initialDelay: 100 * time.Millisecond,
          maxDelay:     2 * time.Second,
          multiplier:   2.0,
      }
  }

  func (rp *RetryPolicy) Execute(ctx context.Context, fn func() error) error {
      var err error
      delay := rp.initialDelay

      for attempt := 0; attempt < rp.maxAttempts; attempt++ {
          err = fn()
          if err == nil {
              return nil
          }

          if attempt < rp.maxAttempts-1 {
              time.Sleep(delay)
              delay = time.Duration(float64(delay) * rp.multiplier)
              if delay > rp.maxDelay {
                  delay = rp.maxDelay
              }
          }
      }

      return fmt.Errorf("retry exhausted after %d attempts: %w", rp.maxAttempts, err)
  }
  ```
- **Localização**: `internal/domain/retry/retry_policy.go`

---

### 3.4. Infrastructure Layer

#### Bacen HTTP Client
- **Responsabilidade**: Cliente HTTP com mTLS para comunicar com Bacen
- **Tecnologia**: Go `net/http` + `crypto/tls`
- **Implementação**:
  ```go
  type BacenHTTPClient struct {
      client  *http.Client
      baseURL string
  }

  func NewBacenHTTPClient(baseURL string, mtlsConfig *tls.Config) *BacenHTTPClient {
      client := &http.Client{
          Transport: &http.Transport{
              TLSClientConfig: mtlsConfig,
          },
          Timeout: 30 * time.Second,
      }

      return &BacenHTTPClient{
          client:  client,
          baseURL: baseURL,
      }
  }

  func (bhc *BacenHTTPClient) Post(ctx context.Context, path string, xmlBody []byte) ([]byte, error) {
      url := bhc.baseURL + path

      req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(xmlBody))
      if err != nil {
          return nil, err
      }

      req.Header.Set("Content-Type", "text/xml; charset=utf-8")
      req.Header.Set("SOAPAction", "")

      resp, err := bhc.client.Do(req)
      if err != nil {
          return nil, fmt.Errorf("http request failed: %w", err)
      }
      defer resp.Body.Close()

      if resp.StatusCode != 200 {
          return nil, fmt.Errorf("bacen returned status %d", resp.StatusCode)
      }

      body, err := ioutil.ReadAll(resp.Body)
      if err != nil {
          return nil, fmt.Errorf("failed to read response body: %w", err)
      }

      return body, nil
  }
  ```
- **Timeout**: 30s por operação
- **Localização**: `internal/infrastructure/http/bacen_http_client.go`

#### mTLS Config
- **Responsabilidade**: Configurar mTLS com certificados ICP-Brasil A3
- **Implementação**:
  ```go
  type MTLSConfig struct {
      tlsConfig *tls.Config
  }

  func NewMTLSConfig(certFile, keyFile, caFile string) (*MTLSConfig, error) {
      // Load client certificate (ICP-Brasil A3)
      cert, err := tls.LoadX509KeyPair(certFile, keyFile)
      if err != nil {
          return nil, fmt.Errorf("failed to load client cert: %w", err)
      }

      // Load CA certificate (Bacen CA)
      caCert, err := ioutil.ReadFile(caFile)
      if err != nil {
          return nil, fmt.Errorf("failed to load ca cert: %w", err)
      }

      caCertPool := x509.NewCertPool()
      caCertPool.AppendCertsFromPEM(caCert)

      tlsConfig := &tls.Config{
          Certificates: []tls.Certificate{cert},
          RootCAs:      caCertPool,
          MinVersion:   tls.VersionTLS12,
      }

      return &MTLSConfig{tlsConfig: tlsConfig}, nil
  }

  func (mc *MTLSConfig) GetTLSConfig() *tls.Config {
      return mc.tlsConfig
  }
  ```
- **Certificados**:
  - Cliente: ICP-Brasil A3 (hardware token ou Cloud HSM)
  - CA: Bacen CA (cadeia de confiança)
- **Localização**: `internal/infrastructure/mtls/mtls_config.go`

#### XML Signer (Java)
- **Responsabilidade**: Assinar digitalmente XML com ICP-Brasil A3
- **Tecnologia**: Java 17 + XMLDSig (JSR 105)
- **Implementação Java**:
  ```java
  public class XMLSigner {
      private KeyStore keyStore;
      private char[] keyPassword;

      public XMLSigner(String keystorePath, String keystorePassword, String keyPassword) throws Exception {
          this.keyStore = KeyStore.getInstance("PKCS12");
          this.keyStore.load(new FileInputStream(keystorePath), keystorePassword.toCharArray());
          this.keyPassword = keyPassword.toCharArray();
      }

      public String signXML(String xmlContent) throws Exception {
          // 1. Parse XML
          DocumentBuilderFactory dbf = DocumentBuilderFactory.newInstance();
          dbf.setNamespaceAware(true);
          Document doc = dbf.newDocumentBuilder().parse(new ByteArrayInputStream(xmlContent.getBytes()));

          // 2. Get private key
          PrivateKey privateKey = (PrivateKey) keyStore.getKey("alias", keyPassword);
          X509Certificate cert = (X509Certificate) keyStore.getCertificate("alias");

          // 3. Create signature
          XMLSignatureFactory fac = XMLSignatureFactory.getInstance("DOM");
          Reference ref = fac.newReference(
              "#Body",
              fac.newDigestMethod(DigestMethod.SHA256, null),
              Collections.singletonList(fac.newTransform(Transform.ENVELOPED, (TransformParameterSpec) null)),
              null, null
          );

          SignedInfo si = fac.newSignedInfo(
              fac.newCanonicalizationMethod(CanonicalizationMethod.INCLUSIVE, (C14NMethodParameterSpec) null),
              fac.newSignatureMethod(SignatureMethod.RSA_SHA256, null),
              Collections.singletonList(ref)
          );

          KeyInfoFactory kif = fac.getKeyInfoFactory();
          KeyInfo ki = kif.newKeyInfo(Collections.singletonList(kif.newX509Data(Collections.singletonList(cert))));

          DOMSignContext dsc = new DOMSignContext(privateKey, doc.getDocumentElement());
          XMLSignature signature = fac.newXMLSignature(si, ki);
          signature.sign(dsc);

          // 4. Convert back to string
          TransformerFactory tf = TransformerFactory.newInstance();
          Transformer trans = tf.newTransformer();
          StringWriter sw = new StringWriter();
          trans.transform(new DOMSource(doc), new StreamResult(sw));

          return sw.toString();
      }
  }
  ```
- **Comunicação com Go**: JNI (Java Native Interface) via cgo
- **Localização**: `java/src/main/java/com/lbpay/dict/XMLSigner.java`

#### JNI Bridge (Go ↔ Java)
- **Responsabilidade**: Interface Go ↔ Java para chamar XML Signer
- **Implementação Go**:
  ```go
  /*
  #cgo LDFLAGS: -L./java/build -lxmlsigner
  #include <stdlib.h>
  #include "xml_signer.h"
  */
  import "C"
  import "unsafe"

  type XMLSigner struct{}

  func (xs *XMLSigner) Sign(xmlBytes []byte) ([]byte, error) {
      cXML := C.CString(string(xmlBytes))
      defer C.free(unsafe.Pointer(cXML))

      // Call Java method via JNI
      cSigned := C.signXML(cXML)
      if cSigned == nil {
          return nil, errors.New("xml signing failed")
      }
      defer C.free(unsafe.Pointer(cSigned))

      signedXML := C.GoString(cSigned)
      return []byte(signedXML), nil
  }
  ```
- **Build**: Requer JRE no runtime
- **Localização**: `internal/infrastructure/xml/xml_signer.go`

#### Redis Cache Client
- **Responsabilidade**: Cache de respostas do Bacen
- **TTL**: 30 segundos
- **Uso**:
  ```go
  type RedisCacheClient struct {
      client *redis.Client
  }

  func (rcc *RedisCacheClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
      return rcc.client.Set(ctx, key, value, ttl).Err()
  }

  func (rcc *RedisCacheClient) Get(ctx context.Context, key string) ([]byte, error) {
      val, err := rcc.client.Get(ctx, key).Bytes()
      if err == redis.Nil {
          return nil, nil  // Cache miss
      }
      return val, err
  }
  ```
- **Keys**: `bridge:response:bacen:{entry_id}`
- **Localização**: `internal/infrastructure/cache/redis_cache_client.go`

---

## 4. Estrutura de Diretórios

```
rsfn-bridge/
├── cmd/
│   └── api/
│       └── main.go                      # Bridge gRPC Server
├── internal/
│   ├── api/                             # API Layer
│   │   └── grpc/
│   │       ├── bridge_server.go
│   │       ├── handlers/
│   │       │   ├── entry_handler.go
│   │       │   ├── claim_handler.go
│   │       │   └── portability_handler.go
│   │       └── validators/
│   │           └── request_validator.go
│   ├── application/                     # Application Layer
│   │   ├── adapters/
│   │   │   └── soap_adapter.go
│   │   ├── xml/
│   │   │   ├── marshaller.go
│   │   │   └── unmarshaller.go
│   │   ├── mappers/
│   │   │   ├── grpc_to_soap_mapper.go
│   │   │   └── soap_to_grpc_mapper.go
│   │   └── soap/
│   │       └── envelope_builder.go
│   ├── domain/                          # Domain Layer
│   │   ├── validators/
│   │   │   ├── bacen_request_validator.go
│   │   │   └── bacen_response_validator.go
│   │   ├── circuit_breaker/
│   │   │   └── circuit_breaker.go
│   │   └── retry/
│   │       └── retry_policy.go
│   └── infrastructure/                  # Infrastructure Layer
│       ├── http/
│       │   └── bacen_http_client.go
│       ├── mtls/
│       │   └── mtls_config.go
│       ├── xml/
│       │   └── xml_signer.go            # JNI bridge
│       └── cache/
│           └── redis_cache_client.go
├── java/                                # Java XML Signer
│   ├── src/main/java/com/lbpay/dict/
│   │   ├── XMLSigner.java
│   │   └── XMLSignatureValidator.java
│   ├── pom.xml
│   └── build.sh
├── proto/
│   └── bridge.proto                     # Bridge gRPC API
├── go.mod
└── go.sum
```

---

## 5. Fluxo de Requisição Completo

### Exemplo: gRPC CreateEntry → SOAP → Bacen

```
1. Temporal Worker
   └→ gRPC: CreateEntry(key_type, key_value, account)
   ↓
2. Bridge gRPC Server
   └→ Request Validator (valida payload gRPC)
   └→ Entry Handler
   ↓
3. SOAP Adapter
   ├→ gRPC to SOAP Mapper (mapeia campos)
   ├→ SOAP Envelope Builder (constrói envelope)
   └→ XML Marshaller (serializa Go → XML)
   ↓
4. XML Signer (Java via JNI)
   └→ Assina XML com ICP-Brasil A3 (RSA-SHA256)
   ↓
5. Bacen Request Validator
   └→ Valida XML assinado
   ↓
6. Circuit Breaker
   └→ Verifica estado (Closed, Open, Half-Open)
       ├→ Se OPEN → Retorna erro imediatamente
       └→ Se CLOSED/HALF-OPEN → Continue
   ↓
7. Retry Policy
   └→ Tentativa 1
   ↓
8. Timeout Manager
   └→ Aplica timeout de 30s
   ↓
9. Bacen HTTP Client (mTLS)
   ├→ Load certificados ICP-Brasil A3
   └→ POST https://dict.bcb.gov.br/api/v1/dict/entries
       (HTTPS mTLS + SOAP/XML)
   ↓
10. Bacen DICT
    └→ Processa SOAP request
    └→ Retorna SOAP response (XML assinado)
   ↓
11. XML Signature Validator (Java)
    └→ Valida assinatura digital do Bacen
   ↓
12. XML Unmarshaller
    └→ Deserializa XML → Go struct
   ↓
13. SOAP to gRPC Mapper
    └→ Mapeia SOAP → gRPC response
   ↓
14. Bacen Response Validator
    └→ Valida response (verifica SOAP faults)
   ↓
15. Redis Cache Client
    └→ SET bridge:response:bacen:{entry_id} EX 30
   ↓
16. Entry Handler
    └→ Retorna gRPC response
   ↓
17. Temporal Worker
    └→ Recebe {bacen_entry_id, status}
```

**Duração Total**: 500-800ms (incluindo Bacen)

---

## 6. Testes por Camada

### 6.1. API Layer - Integration Tests

```go
func TestEntryHandler_CreateEntry_Success(t *testing.T) {
    // Mock SOAP Adapter
    mockSOAPAdapter := new(MockSOAPAdapter)
    handler := &EntryHandler{soapAdapter: mockSOAPAdapter}

    mockSOAPAdapter.On("CreateEntry", mock.Anything, mock.Anything).
        Return(&SOAPCreateEntryResponse{EntryID: "bacen_123", Status: "ACTIVE"}, nil)

    req := &pb.CreateEntryRequest{
        KeyType:  "CPF",
        KeyValue: "12345678900",
        Account: &pb.Account{
            Ispb:          "12345678",
            AccountNumber: "123456",
            Branch:        "0001",
            AccountType:   "CACC",
        },
    }

    resp, err := handler.CreateEntry(context.Background(), req)

    assert.NoError(t, err)
    assert.Equal(t, "bacen_123", resp.BacenEntryId)
    mockSOAPAdapter.AssertExpectations(t)
}
```

### 6.2. Application Layer - Unit Tests

```go
func TestSOAPAdapter_CreateEntry_Success(t *testing.T) {
    // Mock all dependencies
    mockBacenClient := new(MockBacenHTTPClient)
    mockXMLSigner := new(MockXMLSigner)

    adapter := &SOAPAdapter{
        bacenHTTPClient: mockBacenClient,
        xmlSigner:       mockXMLSigner,
    }

    mockXMLSigner.On("Sign", mock.Anything).Return([]byte("<signed>...</signed>"), nil)
    mockBacenClient.On("Post", mock.Anything, mock.Anything, mock.Anything).
        Return([]byte("<soap:Envelope>...</soap:Envelope>"), nil)

    resp, err := adapter.CreateEntry(context.Background(), &pb.CreateEntryRequest{...})

    assert.NoError(t, err)
    assert.NotNil(t, resp)
}
```

### 6.3. Infrastructure Layer - Integration Tests

```go
func TestBacenHTTPClient_Post_Success(t *testing.T) {
    // Start mock SOAP server
    server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/xml")
        w.WriteHeader(200)
        w.Write([]byte(`<soap:Envelope>...</soap:Envelope>`))
    }))
    defer server.Close()

    client := &BacenHTTPClient{client: server.Client(), baseURL: server.URL}

    resp, err := client.Post(context.Background(), "/api/v1/dict/entries", []byte("<soap>...</soap>"))

    assert.NoError(t, err)
    assert.Contains(t, string(resp), "soap:Envelope")
}
```

---

## 7. Próximos Passos

1. **[DIA-007: Sequence Diagram - CreateEntry Flow](./DIA-007_Sequence_CreateEntry.md)** (a criar)
   - Sequência completa CreateEntry (Frontend → Core → Ledger → Pulsar → Connect → Bridge → Bacen)

2. **[SEC-006: XML Signature Security](../../13_Seguranca/SEC-006_XML_Signature_Security.md)** (a criar)
   - Especificação de assinatura digital XML

3. **[IMP-003: Manual Implementação Bridge](../../09_Implementacao/IMP-003_Manual_Implementacao_Bridge.md)** (a criar)
   - Guia de implementação passo a passo

---

## 8. Checklist de Validação

- [ ] SOAP Adapter converte gRPC → SOAP corretamente?
- [ ] XML Signer assina com ICP-Brasil A3?
- [ ] mTLS está configurado com certificados corretos?
- [ ] Circuit Breaker protege contra falhas do Bacen?
- [ ] Retry Policy usa backoff exponencial?
- [ ] Timeout de 30s está aplicado?
- [ ] Cache Redis armazena respostas por 30s?
- [ ] Validações de request/response estão implementadas?

---

## 9. Referências

### Documentos Internos
- [DIA-002: C4 Container Diagram](./DIA-002_C4_Container_Diagram.md)
- [TEC-002 v3.1: Bridge Specification](../../11_Especificacoes_Tecnicas/TEC-002_Bridge_Specification.md)
- [GRPC-001: Bridge gRPC Service](../../04_APIs/gRPC/GRPC-001_Bridge_gRPC_Service.md)
- [SEC-002: ICP-Brasil Certificates](../../13_Seguranca/SEC-002_ICP_Brasil_Certificates.md)

### Documentos Externos
- [SOAP 1.2 Specification](https://www.w3.org/TR/soap12/)
- [XML Digital Signature (XMLDSig)](https://www.w3.org/TR/xmldsig-core/)
- [ICP-Brasil Technical Standards](https://www.iti.gov.br/legislacao)
- [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html)

---

**Última Revisão**: 2025-10-25
**Aprovado por**: Arquitetura LBPay
**Próxima Revisão**: 2026-01-25 (trimestral)
