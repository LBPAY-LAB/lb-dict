# Bridge Entry Operations - Implementação Completa

**Data**: 2025-10-27
**Status**: ✅ IMPLEMENTADO E COMPILANDO
**Versão**: 1.0

---

## 📋 Resumo da Implementação

Implementação completa das **4 operações Entry** do conn-bridge (RSFN Bridge) com integração SOAP/mTLS para API Bacen DICT.

### Arquivos Implementados

1. **SOAP Client** (`internal/infrastructure/bacen/soap_client.go`) - 450 linhas
2. **XML Signer Client** (`internal/infrastructure/signer/xml_signer_client.go`) - 200 linhas
3. **Entry Handlers** (`internal/grpc/entry_handlers.go`) - 360 linhas (reescrito)
4. **Server Update** (`internal/grpc/server.go`) - Atualizado com dependências

**Total**: ~1010 linhas de código Go implementadas

---

## 🎯 Operações Implementadas

### 1. CreateEntry
**Endpoint Bacen**: `POST /api/v1/dict/entries`

**Fluxo**:
```
gRPC Request → Validate → XML Conversion → ICP-Brasil A3 Signature
→ SOAP Envelope → mTLS POST → Bacen Response → Parse → gRPC Response
```

**Validações**:
- Key (tipo e valor obrigatórios)
- Account (ISPB e AccountNumber obrigatórios)
- RequestId obrigatório

**Exemplo XML Gerado**:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<CreateEntryRequest>
  <Entry>
    <Key>12345678900</Key>
    <KeyType>CPF</KeyType>
    <Account>
      <Participant>12345678</Participant>
      <Branch>0001</Branch>
      <AccountNumber>123456</AccountNumber>
      <AccountType>CHECKING</AccountType>
    </Account>
  </Entry>
  <RequestId>req-1234567890</RequestId>
</CreateEntryRequest>
```

---

### 2. GetEntry
**Endpoint Bacen**: `POST /api/v1/dict/entries` (com query params)

**Fluxo**:
```
gRPC Request → Validate (oneof) → XML Conversion → Sign → SOAP
→ mTLS POST → Parse → gRPC Response
```

**Identificadores Suportados (oneof)**:
- `entry_id` - ID interno LBPay
- `external_id` - ID Bacen
- `key_query` - Busca por chave (KeyType + KeyValue)

**Exemplo XML Gerado**:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<GetEntryRequest>
  <Key>12345678900</Key>
  <KeyType>CPF</KeyType>
  <RequestId>req-1234567890</RequestId>
</GetEntryRequest>
```

---

### 3. UpdateEntry
**Endpoint Bacen**: `PUT /api/v1/dict/entries`

**Fluxo**:
```
gRPC Request → Validate → XML Conversion → Sign → SOAP
→ mTLS PUT → Parse → gRPC Response
```

**Permite Atualizar**:
- Conta transacional (NewAccount)
- Mantém a chave PIX inalterada

**Exemplo XML Gerado**:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<UpdateEntryRequest>
  <Key>12345678900</Key>
  <KeyType>CPF</KeyType>
  <NewAccount>
    <Participant>87654321</Participant>
    <Branch>0002</Branch>
    <AccountNumber>654321</AccountNumber>
    <AccountType>SAVINGS</AccountType>
  </NewAccount>
  <RequestId>req-1234567890</RequestId>
</UpdateEntryRequest>
```

---

### 4. DeleteEntry
**Endpoint Bacen**: `DELETE /api/v1/dict/entries`

**Fluxo**:
```
gRPC Request → Validate → XML Conversion → Sign → SOAP
→ mTLS DELETE → Parse → gRPC Response (with timestamp)
```

**Exemplo XML Gerado**:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<DeleteEntryRequest>
  <Key>12345678900</Key>
  <KeyType>CPF</KeyType>
  <RequestId>req-1234567890</RequestId>
</DeleteEntryRequest>
```

---

## 🔐 SOAP Client (soap_client.go)

### Características

**mTLS Configuration**:
- TLS 1.2+ (MinVersion)
- ICP-Brasil A3 certificates
- Client certificate + key
- CA certificate for server verification
- Dev mode com `InsecureSkipVerify` para testes

**Circuit Breaker**:
- Biblioteca: `github.com/sony/gobreaker`
- Nome: `BacenSOAPClient`
- MaxRequests: 3 (half-open)
- Interval: 10s
- Timeout: 30s
- ReadyToTrip: 5 requests com 60% falhas

**Connection Pooling**:
- MaxIdleConns: 20
- MaxIdleConnsPerHost: 20
- MaxConnsPerHost: 20
- IdleConnTimeout: 90s
- TLSHandshakeTimeout: 10s

**Timeouts**:
- Connection: 30s
- Request: 60s (padrão)
- KeepAlive: 30s

### Estruturas SOAP

**SOAPEnvelope**:
```go
type SOAPEnvelope struct {
    XMLName xml.Name    `xml:"soap:Envelope"`
    SoapNS  string      `xml:"xmlns:soap,attr"`
    DictNS  string      `xml:"xmlns:dict,attr"`
    Header  *SOAPHeader `xml:"soap:Header,omitempty"`
    Body    SOAPBody    `xml:"soap:Body"`
}
```

**SOAP Namespaces**:
- SOAP Envelope: `http://www.w3.org/2003/05/soap-envelope`
- WS-Security: `http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd`
- Bacen DICT: `http://www.bcb.gov.br/dict/api/v1`

### Headers HTTP

```
Content-Type: application/soap+xml; charset=utf-8
Accept: application/soap+xml
User-Agent: LBPay-DICT-Bridge/1.0
X-Correlation-ID: <from context>
```

### Métodos Principais

**BuildSOAPEnvelope**: Constrói envelope SOAP com body XML e assinatura
**SendSOAPRequest**: Envia request via circuit breaker
**ParseSOAPResponse**: Extrai body XML do envelope SOAP
**checkSOAPFault**: Detecta e parseia SOAP Faults

### SOAP Fault Handling

```go
type SOAPFault struct {
    Code   string `xml:"Code>Value"`
    Reason string `xml:"Reason>Text"`
    Detail string `xml:"Detail,omitempty"`
}
```

Erros SOAP são convertidos em `fmt.Errorf` com código, motivo e detalhe.

---

## 🔏 XML Signer Client (xml_signer_client.go)

### Características

**Serviço Java**:
- URL padrão: `http://localhost:8081`
- Endpoint: `POST /sign`
- Content-Type: `application/json`
- Timeout: 30s

### Request/Response Format

**SignRequest**:
```json
{
  "xml": "<CreateEntryRequest>...</CreateEntryRequest>"
}
```

**SignResponse**:
```json
{
  "signedXml": "<CreateEntryRequest>...<Signature>...</Signature></CreateEntryRequest>",
  "signature": "<Signature>...</Signature>",
  "error": ""
}
```

**ErrorResponse**:
```json
{
  "error": "SIGNING_FAILED",
  "message": "Certificate expired",
  "code": "CERT_EXPIRED"
}
```

### Métodos

**SignXML**: Assina XML e retorna XML completo assinado
**SignXMLAndGetSignature**: Assina e retorna XML + elemento Signature separado (TODO)
**HealthCheck**: Verifica disponibilidade do serviço Java

### Error Handling

- HTTP 4xx/5xx → `ErrorResponse` parseado
- Timeout → Context error
- JSON inválido → Parse error

---

## 🔄 Fluxo Completo (CreateEntry)

```
┌─────────────┐
│  Connect    │
│  (gRPC)     │
└──────┬──────┘
       │ CreateEntryRequest
       │ {key, account, request_id}
       ▼
┌─────────────────────────────────────────────────────────────┐
│                    BRIDGE (Entry Handler)                    │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  STEP 1: Validate Request                                    │
│  ✓ Key não nulo, tipo especificado, valor preenchido        │
│  ✓ Account não nulo, ISPB + AccountNumber preenchidos       │
│  ✓ RequestId preenchido                                      │
│                                                               │
│  STEP 2: Convert Proto → XML                                 │
│  xml.CreateEntryRequestToXML(req)                            │
│  → XMLCreateEntryRequest (300 bytes)                         │
│                                                               │
│  STEP 3: Sign XML (ICP-Brasil A3)                            │
│  xmlSigner.SignXML(ctx, xmlData)                             │
│  → HTTP POST http://localhost:8081/sign                      │
│  → Java service assina com certificado A3                    │
│  → Signed XML (450 bytes)                                    │
│                                                               │
│  STEP 4: Build SOAP Envelope                                 │
│  soapClient.BuildSOAPEnvelope(signedXML, "")                 │
│  → SOAP 1.2 envelope com namespaces (550 bytes)              │
│                                                               │
│  STEP 5: Send SOAP via mTLS                                  │
│  soapClient.SendSOAPRequest(ctx, endpoint, envelope)         │
│  → Circuit breaker wrapping                                  │
│  → HTTPS POST com mTLS (TLS 1.2+)                            │
│  → Timeout 30s                                               │
│                                                               │
│  STEP 6: Parse SOAP Response                                 │
│  soapClient.ParseSOAPResponse(soapResponse)                  │
│  → Extrai body XML do envelope                               │
│  → Verifica SOAP Faults                                      │
│                                                               │
│  STEP 7: Convert XML → Proto Response                        │
│  xml.CreateEntryResponseFromXML(bodyXML)                     │
│  → CreateEntryResponse proto                                 │
│                                                               │
└─────────────────────────────────────────────────────────────┘
       │ CreateEntryResponse
       │ {entry_id, external_id, status}
       ▼
┌─────────────┐
│  Connect    │
│  (returns)  │
└─────────────┘
```

**Tempo Total Estimado**:
- XML Conversion: 1-2ms
- XML Signing: 50-100ms (Java service)
- SOAP Build: 1ms
- mTLS Round-trip: 100-300ms (Bacen)
- Parse: 2-5ms
- **TOTAL**: ~150-410ms por operação

---

## 📊 Logs de Exemplo (CreateEntry)

```
INFO  CreateEntry called requestId=req-123 keyType=CPF keyValue=12****00
DEBUG Converted gRPC request to XML xmlSize=320
DEBUG XML signed successfully with ICP-Brasil A3 signedXMLSize=485
DEBUG Built SOAP envelope soapEnvelopeSize=612
INFO  Sending SOAP/HTTPS request to Bacen method=POST url=https://api.bacen.br/api/v1/dict/entries
INFO  Received SOAP response from Bacen statusCode=200 bodySize=580
DEBUG Received SOAP response from Bacen responseSize=580
INFO  CreateEntry completed successfully entryId=entry-550e8400 externalId=bacen-123 status=ACTIVE
```

---

## 🧪 Testes Necessários

### Unit Tests

**SOAP Client**:
- [ ] BuildSOAPEnvelope com body válido
- [ ] BuildSOAPEnvelope com signature
- [ ] ParseSOAPResponse extrai body corretamente
- [ ] checkSOAPFault detecta faults
- [ ] Circuit breaker abre após falhas
- [ ] mTLS handshake com certificados válidos

**XML Signer Client**:
- [ ] SignXML retorna XML assinado
- [ ] SignXML trata erros HTTP 4xx/5xx
- [ ] HealthCheck retorna OK
- [ ] Timeout funciona corretamente

**Entry Handlers**:
- [ ] CreateEntry valida request
- [ ] CreateEntry converte proto → XML
- [ ] GetEntry com entry_id
- [ ] GetEntry com key_query
- [ ] UpdateEntry valida new_account
- [ ] DeleteEntry seta DeletedAt timestamp

### Integration Tests

- [ ] CreateEntry E2E com mock Bacen
- [ ] GetEntry E2E com mock Bacen
- [ ] UpdateEntry E2E com mock Bacen
- [ ] DeleteEntry E2E com mock Bacen
- [ ] Circuit breaker abre após 5 falhas
- [ ] XML Signer integração com Java service

### E2E Tests (Com Bacen Sandbox)

- [ ] CreateEntry → Bacen Sandbox
- [ ] GetEntry → Bacen Sandbox
- [ ] UpdateEntry → Bacen Sandbox
- [ ] DeleteEntry → Bacen Sandbox
- [ ] mTLS handshake com certificado ICP-Brasil válido
- [ ] SOAP Fault handling

---

## 🔧 Configuração

### Environment Variables

```bash
# SOAP Client (Bacen API)
BACEN_BASE_URL=https://api.bacen.br
BACEN_TIMEOUT=60s
BACEN_CERT_PATH=/certs/lbpay-client.crt
BACEN_KEY_PATH=/certs/lbpay-client.key
BACEN_CA_PATH=/certs/bacen-ca.crt
BACEN_DEV_MODE=false

# XML Signer
XML_SIGNER_URL=http://localhost:8081
XML_SIGNER_TIMEOUT=30s

# gRPC Server
GRPC_PORT=9094
```

### Docker Compose

```yaml
services:
  bridge:
    build: ./conn-bridge
    ports:
      - "9094:9094"
    environment:
      - BACEN_BASE_URL=${BACEN_BASE_URL}
      - XML_SIGNER_URL=http://xml-signer:8081
    volumes:
      - ./certs:/certs:ro
    depends_on:
      - xml-signer

  xml-signer:
    build: ./conn-bridge/xml-signer
    ports:
      - "8081:8081"
    environment:
      - KEYSTORE_PATH=/keystore/lbpay-a3.p12
      - KEYSTORE_PASSWORD=${KEYSTORE_PASSWORD}
    volumes:
      - ./keystore:/keystore:ro
```

---

## 📈 Métricas e Observabilidade

### Prometheus Metrics (TODO)

```
# Requests
bridge_entry_requests_total{operation="CreateEntry",status="success"} 1250
bridge_entry_requests_total{operation="CreateEntry",status="error"} 15

# Latency
bridge_entry_request_duration_seconds{operation="CreateEntry",quantile="0.5"} 0.215
bridge_entry_request_duration_seconds{operation="CreateEntry",quantile="0.95"} 0.380
bridge_entry_request_duration_seconds{operation="CreateEntry",quantile="0.99"} 0.520

# Circuit Breaker
bridge_circuit_breaker_state{name="BacenSOAPClient"} 0  # 0=closed, 1=open, 2=half-open
bridge_circuit_breaker_failures_total{name="BacenSOAPClient"} 5
bridge_circuit_breaker_successes_total{name="BacenSOAPClient"} 1245

# XML Signer
bridge_xml_signer_requests_total{status="success"} 1265
bridge_xml_signer_duration_seconds{quantile="0.95"} 0.085
```

### Logs Estruturados (Logrus)

```json
{
  "level": "info",
  "msg": "CreateEntry completed successfully",
  "requestId": "req-1234567890",
  "entryId": "entry-550e8400",
  "externalId": "bacen-abc123",
  "status": "ACTIVE",
  "duration_ms": 215,
  "timestamp": "2025-10-27T10:30:45Z"
}
```

---

## 🚀 Próximos Passos

### Imediato (Sprint 1)

1. ✅ **SOAP Client** - Completo
2. ✅ **XML Signer Client** - Completo
3. ✅ **Entry Handlers** - Completo
4. ✅ **Compilação** - Sucesso
5. 🔲 **Unit Tests** - Criar
6. 🔲 **Integration Tests** - Criar

### Sprint 2

1. 🔲 **Claim Handlers** - Integrar SOAP + XML Signer
2. 🔲 **Portability Handlers** - Integrar SOAP + XML Signer
3. 🔲 **Metrics** - Implementar Prometheus
4. 🔲 **E2E Tests** - Com mock Bacen

### Sprint 3

1. 🔲 **Bacen Sandbox Tests** - Com certificado real
2. 🔲 **Performance Tests** - 1000 TPS
3. 🔲 **Stress Tests** - Circuit breaker validation
4. 🔲 **Production Ready** - Homologação Bacen

---

## 📚 Referências

### Especificações

- **TEC-002**: Bridge Architecture (v3.1)
- **GRPC-001**: Bridge gRPC API Specification
- **REG-001**: Regulatory Requirements (Bacen)
- **ESCOPO_BRIDGE_VALIDADO.md**: Validated scope

### Bibliotecas

- `github.com/sony/gobreaker`: Circuit breaker
- `google.golang.org/grpc`: gRPC server
- `github.com/sirupsen/logrus`: Structured logging
- `encoding/xml`: XML marshaling/unmarshaling

### Bacen DICT API

- Manual Técnico DICT v3.1
- SOAP 1.2 Specification
- WS-Security 1.0
- ICP-Brasil A3 PKI

---

## ✅ Validação Final

### Compilação

```bash
cd /Users/jose.silva.lb/LBPay/IA_Dict/conn-bridge
go build ./...
```

**Resultado**: ✅ Compilando sem erros

### Arquivos Criados/Modificados

```
conn-bridge/
├── internal/
│   ├── grpc/
│   │   ├── entry_handlers.go       [REESCRITO - 360 linhas]
│   │   └── server.go               [MODIFICADO - +interfaces]
│   ├── infrastructure/
│   │   ├── bacen/
│   │   │   └── soap_client.go      [NOVO - 450 linhas]
│   │   └── signer/
│   │       └── xml_signer_client.go [NOVO - 200 linhas]
│   └── xml/
│       ├── structs.go              [EXISTENTE]
│       └── converter.go            [EXISTENTE - bugfix]
```

### Status das Operações

| Operação     | Status | LOC | Tests |
|--------------|--------|-----|-------|
| CreateEntry  | ✅ OK  | 70  | 🔲    |
| GetEntry     | ✅ OK  | 70  | 🔲    |
| UpdateEntry  | ✅ OK  | 70  | 🔲    |
| DeleteEntry  | ✅ OK  | 75  | 🔲    |
| SOAP Client  | ✅ OK  | 450 | 🔲    |
| XML Signer   | ✅ OK  | 200 | 🔲    |

---

**Última Atualização**: 2025-10-27
**Status**: ✅ IMPLEMENTAÇÃO COMPLETA E COMPILANDO
**Próximo Marco**: Unit Tests + Integration Tests
