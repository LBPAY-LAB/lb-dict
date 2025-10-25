# TST-003: Test Cases - Bridge mTLS Communication

**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: QA Team
**Status**: ✅ Completo

---

## Sumário Executivo

Este documento apresenta os **test cases completos** para a comunicação **mTLS (Mutual TLS)** entre o **RSFN Bridge** e o **Bacen DICT**, incluindo validação de certificados ICP-Brasil, assinatura XML e comunicação SOAP.

**Objetivo**: Garantir que a comunicação segura entre LBPay e Bacen está corretamente implementada conforme especificações do Banco Central.

**Cobertura**:
- Certificate Validation (P0)
- mTLS Handshake (P0)
- XML Signature (P0)
- SOAP Request/Response (P0)
- Negative Tests: Invalid cert, Expired cert, Untrusted CA (P0)
- Certificate Rotation (P1)
- Performance Tests (P1)

**Referências**:
- [TEC-002: Bridge Specification](../../11_Especificacoes_Tecnicas/TEC-002_Bridge_Specification.md)
- [SEC-006: XML Signature Security](../../13_Seguranca/SEC-006_XML_Signature_Security.md)
- [Bacen DICT Manual](https://www.bcb.gov.br/estabilidadefinanceira/pix)

---

## Test Environment Setup

### Prerequisites
```yaml
Environment: test
Bridge URL: bridge-test.lbpay.com.br:50051 (gRPC)
Bacen URL: dict-test.bcb.gov.br:443 (HTTPS mTLS)
Certificate Store: /etc/bridge/certs/
CA Certificate: ICP-Brasil v10 (test)
Client Certificate: LBPay PIX Certificate (ICP-Brasil A3)
HSM: Software HSM (test mode)
```

### Test Certificates
```yaml
Valid Certificate (Test):
  CN: LBPay LTDA - PIX DICT
  Issuer: AC Certisign RFB G5
  Serial: 0x1234567890ABCDEF
  Valid From: 2025-01-01
  Valid To: 2027-01-01
  Key Usage: Digital Signature, Key Encipherment
  Extended Key Usage: Client Authentication
  ICP-Brasil: True (Test Chain)

Expired Certificate:
  CN: LBPay LTDA - PIX DICT
  Valid From: 2023-01-01
  Valid To: 2025-01-01 (expired)
  Status: EXPIRED

Invalid Certificate:
  CN: Unknown Company
  Issuer: Self-Signed
  ICP-Brasil: False
  Status: INVALID

Untrusted CA Certificate:
  Issuer: Unknown CA (not in ICP-Brasil chain)
  Status: UNTRUSTED
```

### Test Environment Variables
```bash
export BRIDGE_CERT_PATH=/etc/bridge/certs/client-cert.pem
export BRIDGE_KEY_PATH=/etc/bridge/certs/client-key.pem
export BRIDGE_CA_PATH=/etc/bridge/certs/ca-cert.pem
export BACEN_URL=https://dict-test.bcb.gov.br
export HSM_ENABLED=false
export XML_SIGNATURE_ENABLED=true
```

---

## P0 Test Cases (Certificate Validation)

### TC-003-001: Certificate Validation - Valid ICP-Brasil A3 Certificate

**Priority**: P0 (Critical)
**Type**: Security Test
**Estimated Duration**: 100-200ms

#### Preconditions
- Bridge has valid ICP-Brasil A3 certificate
- Certificate is not expired
- Certificate is issued by trusted CA (ICP-Brasil chain)
- Private key is accessible

#### BDD Format (Given/When/Then)
```gherkin
Given the Bridge has valid ICP-Brasil A3 certificate
And the certificate is issued by AC Certisign RFB G5
And the certificate is not expired (valid until 2027-01-01)
And the private key matches the public key
When the Bridge starts up
Then the certificate is loaded successfully
And the certificate is validated against ICP-Brasil root CA
And the certificate chain is verified
And the certificate key usage is validated
And the Bridge is ready to accept requests
```

#### Steps
1. Prepare valid certificate:
   ```bash
   # Copy valid test certificate
   cp /test-certs/valid-cert.pem /etc/bridge/certs/client-cert.pem
   cp /test-certs/valid-key.pem /etc/bridge/certs/client-key.pem
   cp /test-certs/ca-cert.pem /etc/bridge/certs/ca-cert.pem
   ```

2. Start Bridge service:
   ```bash
   docker-compose up -d bridge
   ```

3. Verify Bridge started successfully:
   ```bash
   docker logs bridge | grep "Certificate loaded successfully"
   ```
   Expected output:
   ```
   [INFO] Certificate loaded successfully
   [INFO] Certificate CN: LBPay LTDA - PIX DICT
   [INFO] Certificate Serial: 0x1234567890ABCDEF
   [INFO] Certificate Valid From: 2025-01-01
   [INFO] Certificate Valid To: 2027-01-01
   [INFO] Certificate Issuer: AC Certisign RFB G5
   [INFO] Certificate chain validated against ICP-Brasil root CA
   [INFO] Bridge gRPC server listening on :50051
   ```

4. Verify certificate details via gRPC:
   ```bash
   grpcurl -plaintext -d '{}' \
     localhost:50051 bridge.BridgeService/GetCertificateInfo
   ```
   Expected response:
   ```json
   {
     "certificate": {
       "common_name": "LBPay LTDA - PIX DICT",
       "serial_number": "0x1234567890ABCDEF",
       "issuer": "AC Certisign RFB G5",
       "valid_from": "2025-01-01T00:00:00Z",
       "valid_to": "2027-01-01T00:00:00Z",
       "key_usage": ["Digital Signature", "Key Encipherment"],
       "extended_key_usage": ["Client Authentication"],
       "icp_brasil_validated": true
     },
     "status": "VALID"
   }
   ```

#### Expected Result
- ✅ Certificate loaded successfully
- ✅ Certificate validated against ICP-Brasil root CA
- ✅ Certificate chain verified (3 levels: Root → Intermediate → Leaf)
- ✅ Key usage validated (Digital Signature + Client Auth)
- ✅ Bridge ready to accept requests
- ✅ No errors in logs

#### Actual Result
[To be filled during execution]

#### Status
⬜ Not Run | 🟡 In Progress | ✅ Pass | ❌ Fail | 🚫 Blocked

---

### TC-003-002: Certificate Validation - Expired Certificate

**Priority**: P0 (Critical)
**Type**: Negative Security Test

#### Preconditions
- Bridge has expired ICP-Brasil certificate (valid_to < NOW)

#### BDD Format
```gherkin
Given the Bridge has expired certificate (valid until 2024-01-01)
When the Bridge attempts to start
Then the certificate validation fails
And the Bridge logs error "Certificate expired"
And the Bridge refuses to start
And an alert is triggered to DevOps
```

#### Steps
1. Prepare expired certificate:
   ```bash
   cp /test-certs/expired-cert.pem /etc/bridge/certs/client-cert.pem
   cp /test-certs/expired-key.pem /etc/bridge/certs/client-key.pem
   ```

2. Attempt to start Bridge:
   ```bash
   docker-compose up -d bridge
   ```

3. Verify Bridge failed to start:
   ```bash
   docker logs bridge | grep "Certificate expired"
   ```
   Expected error:
   ```
   [ERROR] Certificate validation failed
   [ERROR] Certificate expired on 2024-01-01
   [ERROR] Current date: 2025-10-25
   [ERROR] Bridge cannot start with expired certificate
   [FATAL] Exiting with code 1
   ```

4. Verify Bridge is not running:
   ```bash
   docker ps | grep bridge
   ```
   Expected: No bridge container running

5. Verify alert triggered:
   ```bash
   # Check Prometheus alert manager
   curl http://alertmanager:9093/api/v2/alerts | jq '.[] | select(.labels.alertname == "BridgeCertificateExpired")'
   ```

#### Expected Result
- ✅ Bridge refuses to start
- ✅ Clear error message: "Certificate expired"
- ✅ Alert triggered to DevOps
- ✅ No mTLS connection attempted

#### Status
⬜ Not Run

---

### TC-003-003: Certificate Validation - Invalid Certificate (Not ICP-Brasil)

**Priority**: P0 (Critical)
**Type**: Negative Security Test

#### Preconditions
- Bridge has certificate NOT issued by ICP-Brasil CA

#### BDD Format
```gherkin
Given the Bridge has certificate issued by unknown CA
And the certificate is not in ICP-Brasil chain
When the Bridge attempts to start
Then the certificate validation fails
And the error is "Certificate not issued by ICP-Brasil CA"
And the Bridge refuses to start
```

#### Expected Result
- ✅ Bridge refuses to start
- ✅ Error: Certificate not ICP-Brasil
- ✅ Bridge logs CA chain mismatch

#### Status
⬜ Not Run

---

### TC-003-004: Certificate Validation - Missing Private Key

**Priority**: P0 (Critical)
**Type**: Negative Security Test

#### BDD Format
```gherkin
Given the Bridge has valid certificate
But the private key file is missing or inaccessible
When the Bridge attempts to start
Then the certificate loading fails
And the error is "Private key not found or inaccessible"
And the Bridge refuses to start
```

#### Expected Result
- ✅ Bridge refuses to start
- ✅ Error: Private key missing
- ✅ File permission error logged

#### Status
⬜ Not Run

---

### TC-003-005: Certificate Validation - Key Pair Mismatch

**Priority**: P0 (Critical)
**Type**: Negative Security Test

#### BDD Format
```gherkin
Given the Bridge has valid certificate
But the private key does not match the public key in certificate
When the Bridge attempts to start
Then the key pair validation fails
And the error is "Private key does not match certificate public key"
And the Bridge refuses to start
```

#### Expected Result
- ✅ Bridge refuses to start
- ✅ Error: Key pair mismatch
- ✅ Cryptographic validation error logged

#### Status
⬜ Not Run

---

## P0 Test Cases (mTLS Handshake)

### TC-003-006: mTLS Handshake - Successful Connection to Bacen

**Priority**: P0 (Critical)
**Type**: Integration Test
**Estimated Duration**: 200-300ms

#### Preconditions
- Bridge has valid ICP-Brasil certificate
- Bacen test environment is available
- Bridge can reach Bacen URL (network connectivity)

#### BDD Format
```gherkin
Given the Bridge has valid ICP-Brasil certificate
And the Bacen test environment is available
And the Bridge trusts Bacen's certificate (ICP-Brasil root CA)
When the Bridge attempts to connect to Bacen via mTLS
Then the TLS handshake succeeds
And client certificate is presented to Bacen
And Bacen validates client certificate
And Bacen presents its certificate
And Bridge validates Bacen certificate
And secure mTLS connection is established
And the connection is ready for SOAP requests
```

#### Steps
1. Ensure Bridge is running with valid certificate (use TC-003-001)

2. Call Bridge gRPC endpoint to trigger Bacen connection:
   ```bash
   grpcurl -plaintext -d '{
     "key": {
       "type": "CPF",
       "value": "12345678900"
     },
     "account": {
       "ispb": "12345678",
       "account_number": "123456",
       "branch": "0001",
       "account_type": "CACC"
     }
   }' localhost:50051 bridge.BridgeService/CreateEntry
   ```

3. Verify mTLS handshake in logs:
   ```bash
   docker logs bridge | grep "mTLS handshake"
   ```
   Expected output:
   ```
   [INFO] Initiating mTLS connection to dict-test.bcb.gov.br:443
   [INFO] Presenting client certificate: CN=LBPay LTDA - PIX DICT
   [INFO] Client certificate sent
   [INFO] Server certificate received: CN=DICT Bacen Test
   [INFO] Server certificate validated against ICP-Brasil root CA
   [INFO] mTLS handshake successful
   [INFO] Cipher Suite: TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
   [INFO] TLS Version: TLS 1.3
   [INFO] Connection established
   ```

4. Verify connection in network logs (tcpdump):
   ```bash
   tcpdump -i any -n host dict-test.bcb.gov.br and port 443
   ```
   Expected: TLS handshake packets (ClientHello, ServerHello, Certificate, CertificateVerify, Finished)

5. Verify Bacen response (SOAP success):
   Expected gRPC response:
   ```json
   {
     "entry_id": "bacen-uuid-123",
     "status": "ACTIVE"
   }
   ```

#### Expected Result
- ✅ mTLS handshake successful
- ✅ TLS 1.2 or 1.3 used
- ✅ Strong cipher suite (AES-256-GCM)
- ✅ Both certificates validated
- ✅ Secure connection established
- ✅ SOAP request/response successful

#### Status
⬜ Not Run

---

### TC-003-007: mTLS Handshake - Bacen Rejects Invalid Client Certificate

**Priority**: P0 (Critical)
**Type**: Negative Security Test

#### BDD Format
```gherkin
Given the Bridge has invalid/expired certificate
When the Bridge attempts to connect to Bacen via mTLS
Then Bacen rejects the client certificate
And the mTLS handshake fails
And the error is "Client certificate validation failed"
And no SOAP request is sent
```

#### Expected Result
- ✅ Bacen rejects connection
- ✅ TLS alert: "certificate_unknown" or "certificate_expired"
- ✅ No SOAP request sent
- ✅ Bridge logs error

#### Status
⬜ Not Run

---

### TC-003-008: mTLS Handshake - Bridge Rejects Untrusted Bacen Certificate

**Priority**: P0 (Critical)
**Type**: Negative Security Test

#### BDD Format
```gherkin
Given Bacen presents certificate NOT issued by ICP-Brasil CA
When the Bridge validates Bacen's certificate
Then the validation fails
And the mTLS handshake is aborted
And the error is "Server certificate not trusted"
And no SOAP request is sent
```

#### Expected Result
- ✅ Bridge rejects connection
- ✅ Error: Untrusted server certificate
- ✅ Connection aborted
- ✅ Security alert triggered

#### Status
⬜ Not Run

---

## P0 Test Cases (XML Signature)

### TC-003-009: XML Signature - Valid Signature with ICP-Brasil

**Priority**: P0 (Critical)
**Type**: Security Test
**Estimated Duration**: 50-100ms

#### Preconditions
- Bridge has valid ICP-Brasil certificate
- XML signer is configured
- SOAP request is ready to be signed

#### BDD Format
```gherkin
Given the Bridge has valid ICP-Brasil certificate
And a SOAP CreateEntry request is prepared
When the Bridge signs the XML with ICP-Brasil certificate
Then the XML signature is created successfully
And the signature uses RSA-SHA256 algorithm
And the signature contains certificate info (X509Data)
And the signature is valid according to XML-DSig spec
And Bacen can verify the signature
```

#### Steps
1. Prepare SOAP request (unsigned):
   ```xml
   <soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
                     xmlns:dict="http://www.bcb.gov.br/dict/v1">
     <soapenv:Body wsu:Id="Body">
       <dict:CreateEntryRequest>
         <dict:Key>
           <dict:Type>CPF</dict:Type>
           <dict:Value>12345678900</dict:Value>
         </dict:Key>
       </dict:CreateEntryRequest>
     </soapenv:Body>
   </soapenv:Envelope>
   ```

2. Call Bridge to sign XML:
   ```bash
   grpcurl -plaintext -d '{
     "key": { "type": "CPF", "value": "12345678900" },
     "account": { ... }
   }' localhost:50051 bridge.BridgeService/CreateEntry
   ```

3. Capture signed XML from logs or network:
   ```bash
   docker logs bridge | grep "Signed XML"
   ```

4. Verify signed XML structure:
   Expected:
   ```xml
   <soapenv:Envelope ...>
     <soapenv:Body wsu:Id="Body">
       ...
     </soapenv:Body>
     <ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
       <ds:SignedInfo>
         <ds:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
         <ds:SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"/>
         <ds:Reference URI="#Body">
           <ds:Transforms>
             <ds:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
           </ds:Transforms>
           <ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"/>
           <ds:DigestValue>...</ds:DigestValue>
         </ds:Reference>
       </ds:SignedInfo>
       <ds:SignatureValue>...</ds:SignatureValue>
       <ds:KeyInfo>
         <ds:X509Data>
           <ds:X509Certificate>MIIDxz...</ds:X509Certificate>
         </ds:X509Data>
       </ds:KeyInfo>
     </ds:Signature>
   </soapenv:Envelope>
   ```

5. Verify signature components:
   - Canonicalization: Exclusive XML Canonicalization
   - Signature Algorithm: RSA-SHA256
   - Digest Algorithm: SHA256
   - X509Certificate: Present and valid

6. Verify signature programmatically:
   ```bash
   # Use xmlsec1 tool to verify
   xmlsec1 --verify signed-request.xml
   ```
   Expected: "Signature is OK"

#### Expected Result
- ✅ XML signed successfully
- ✅ Signature algorithm: RSA-SHA256
- ✅ Certificate embedded in signature
- ✅ Signature valid (verified by xmlsec1)
- ✅ Bacen accepts signature

#### Status
⬜ Not Run

---

### TC-003-010: XML Signature - Signature Verification by Bacen

**Priority**: P0 (Critical)
**Type**: Integration Test

#### BDD Format
```gherkin
Given the Bridge sends signed SOAP request to Bacen
When Bacen receives the request
Then Bacen verifies the XML signature
And Bacen validates the certificate in signature
And Bacen validates signature algorithm (RSA-SHA256)
And Bacen accepts the request
And Bacen returns SOAP response
```

#### Expected Result
- ✅ Bacen accepts signed request
- ✅ No signature validation errors
- ✅ SOAP response returned successfully

#### Status
⬜ Not Run

---

### TC-003-011: XML Signature - Invalid Signature (Tampered XML)

**Priority**: P0 (Critical)
**Type**: Negative Security Test

#### BDD Format
```gherkin
Given the Bridge signs a SOAP request
And the XML is modified after signing (e.g., CPF changed)
When the tampered XML is sent to Bacen
Then Bacen rejects the request
And the error is "Invalid XML signature"
And no operation is performed
```

#### Expected Result
- ✅ Bacen rejects tampered request
- ✅ Error: Invalid signature
- ✅ Security alert triggered

#### Status
⬜ Not Run

---

### TC-003-012: XML Signature - Signature with Expired Certificate

**Priority**: P0 (Critical)
**Type**: Negative Security Test

#### BDD Format
```gherkin
Given the Bridge signs XML with expired certificate
When the signed request is sent to Bacen
Then Bacen rejects the signature
And the error is "Certificate in signature expired"
```

#### Expected Result
- ✅ Bacen rejects signature
- ✅ Error: Certificate expired
- ✅ No operation performed

#### Status
⬜ Not Run

---

## P0 Test Cases (SOAP Request/Response)

### TC-003-013: SOAP Request - CreateEntry Successful

**Priority**: P0 (Critical)
**Type**: Integration Test
**Estimated Duration**: 500-800ms

#### BDD Format
```gherkin
Given the Bridge has valid certificate and connection to Bacen
When the Bridge sends SOAP CreateEntry request
Then Bacen returns SOAP CreateEntry response
And the response status is 200 OK
And the response contains bacen_entry_id
And the response is valid SOAP XML
```

#### Steps
1. Send gRPC request to Bridge:
   ```bash
   grpcurl -plaintext -d '{
     "key": { "type": "CPF", "value": "12345678900" },
     "account": {
       "ispb": "12345678",
       "account_number": "123456",
       "branch": "0001",
       "account_type": "CACC"
     }
   }' localhost:50051 bridge.BridgeService/CreateEntry
   ```

2. Verify SOAP request sent (from logs):
   ```xml
   <soapenv:Envelope ...>
     <soapenv:Header>
       <dict:Authentication>
         <dict:Certificate>ICP-Brasil A3 Certificate</dict:Certificate>
       </dict:Authentication>
     </soapenv:Header>
     <soapenv:Body>
       <dict:CreateEntryRequest>
         <dict:Key>
           <dict:Type>CPF</dict:Type>
           <dict:Value>12345678900</dict:Value>
         </dict:Key>
         <dict:Account>
           <dict:ISPB>12345678</dict:ISPB>
           <dict:AccountNumber>123456</dict:AccountNumber>
           <dict:Branch>0001</dict:Branch>
           <dict:AccountType>CACC</dict:AccountType>
         </dict:Account>
       </dict:CreateEntryRequest>
     </soapenv:Body>
     <ds:Signature>...</ds:Signature>
   </soapenv:Envelope>
   ```

3. Verify SOAP response received:
   Expected:
   ```xml
   <soapenv:Envelope ...>
     <soapenv:Body>
       <dict:CreateEntryResponse>
         <dict:EntryId>bacen-uuid-123456</dict:EntryId>
         <dict:Status>ACTIVE</dict:Status>
         <dict:CreatedAt>2025-10-25T10:00:00Z</dict:CreatedAt>
       </dict:CreateEntryResponse>
     </soapenv:Body>
   </soapenv:Envelope>
   ```

4. Verify gRPC response:
   Expected:
   ```json
   {
     "entry_id": "bacen-uuid-123456",
     "status": "ACTIVE"
   }
   ```

#### Expected Result
- ✅ SOAP request sent successfully
- ✅ SOAP response received (200 OK)
- ✅ Response contains entry_id
- ✅ Total time < 800ms

#### Status
⬜ Not Run

---

### TC-003-014: SOAP Response - Bacen Returns SOAP Fault

**Priority**: P0 (Critical)
**Type**: Negative Integration Test

#### BDD Format
```gherkin
Given the Bridge sends SOAP request with invalid data (e.g., invalid CPF)
When Bacen validates the request
Then Bacen returns SOAP Fault
And the fault code is "dict:InvalidKey"
And the fault string contains error description
And the Bridge parses the SOAP Fault
And the Bridge returns gRPC error to caller
```

#### Expected Result
- ✅ SOAP Fault parsed correctly
- ✅ gRPC error returned with fault details
- ✅ Error logged in Bridge

#### Status
⬜ Not Run

---

### TC-003-015: SOAP Request - Timeout Handling

**Priority**: P0 (Critical)
**Type**: Error Handling Test

#### BDD Format
```gherkin
Given Bacen is configured to delay response > 30 seconds
When the Bridge sends SOAP request
Then the request times out after 30 seconds
And the Bridge returns error "Bacen timeout"
And the Temporal workflow retries the request
```

#### Expected Result
- ✅ Request times out after 30s
- ✅ Error: Bacen timeout
- ✅ Retry triggered (3 attempts)

#### Status
⬜ Not Run

---

## P1 Test Cases (Advanced Scenarios)

### TC-003-016: Certificate Rotation - Hot Reload New Certificate

**Priority**: P1 (High)
**Type**: Operational Test

#### BDD Format
```gherkin
Given the Bridge is running with certificate A
And certificate A will expire in 30 days
When a new certificate B is deployed
And the Bridge receives SIGHUP signal (reload config)
Then the Bridge loads certificate B
And validates certificate B
And switches to certificate B for new connections
And existing connections complete with certificate A
And no downtime occurs
```

#### Expected Result
- ✅ Certificate rotated without downtime
- ✅ New connections use new certificate
- ✅ Old connections gracefully complete

#### Status
⬜ Not Run

---

### TC-003-017: Performance - mTLS Connection Pool

**Priority**: P1 (High)
**Type**: Performance Test

#### BDD Format
```gherkin
Given the Bridge maintains connection pool of 10 connections
When 100 concurrent requests arrive
Then connections are reused from pool
And connection establishment overhead is minimized
And p95 latency is < 100ms per request
```

#### Expected Result
- ✅ Connection pool working
- ✅ Connection reuse > 90%
- ✅ p95 latency < 100ms

#### Status
⬜ Not Run

---

### TC-003-018: Performance - XML Signature Caching

**Priority**: P1 (High)
**Type**: Performance Test

#### BDD Format
```gherkin
Given XML signature creation takes 50ms
When the same request is sent twice
Then the second request reuses cached signature components
And signature creation time is < 10ms
```

#### Expected Result
- ✅ Signature caching working
- ✅ 80% latency reduction on cache hit

#### Status
⬜ Not Run

---

### TC-003-019: Security - Certificate Pinning

**Priority**: P1 (High)
**Type**: Security Test

#### BDD Format
```gherkin
Given the Bridge has pinned Bacen's certificate fingerprint
When Bacen presents different certificate (even if valid ICP-Brasil)
Then the Bridge rejects the connection
And the error is "Certificate pinning failed"
And a security alert is triggered
```

#### Expected Result
- ✅ Certificate pinning enforced
- ✅ Unexpected certificate rejected
- ✅ Security alert triggered

#### Status
⬜ Not Run

---

### TC-003-020: Security - TLS Version Enforcement

**Priority**: P1 (High)
**Type**: Security Test

#### BDD Format
```gherkin
Given the Bridge enforces TLS 1.2+
When a connection attempt is made with TLS 1.0 or 1.1
Then the connection is rejected
And the error is "TLS version too old"
```

#### Expected Result
- ✅ TLS 1.0/1.1 rejected
- ✅ Only TLS 1.2+ allowed
- ✅ Security best practices enforced

#### Status
⬜ Not Run

---

## Test Execution Summary

### Coverage Matrix

| Category | P0 Tests | P1 Tests | Total |
|----------|----------|----------|-------|
| Certificate Validation | 5 | 1 | 6 |
| mTLS Handshake | 3 | 1 | 4 |
| XML Signature | 4 | 1 | 5 |
| SOAP Request/Response | 3 | 1 | 4 |
| Security Tests | 0 | 2 | 2 |
| **Total** | **15** | **6** | **21** |

### Priority Distribution
- P0 (Critical): 15 tests (71%)
- P1 (High): 6 tests (29%)

### Test Types
- Security Tests: 10 tests
- Integration Tests: 6 tests
- Negative Tests: 8 tests
- Performance Tests: 3 tests
- Operational Tests: 1 test

---

## Automation Scripts

### Go Integration Test Example
```go
// bridge_mtls_test.go
package bridge_test

import (
    "context"
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    pb "lbpay.com.br/bridge/proto"
)

func TestMTLSConnection(t *testing.T) {
    // TC-003-006: mTLS Handshake - Successful Connection

    // Load client certificate
    cert, err := tls.LoadX509KeyPair(
        "/etc/bridge/certs/client-cert.pem",
        "/etc/bridge/certs/client-key.pem",
    )
    require.NoError(t, err)

    // Load CA certificate
    caCert, err := ioutil.ReadFile("/etc/bridge/certs/ca-cert.pem")
    require.NoError(t, err)

    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)

    // Configure mTLS
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
        MinVersion:   tls.VersionTLS12,
    }

    // Create Bridge client
    conn, err := grpc.Dial(
        "localhost:50051",
        grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
    )
    require.NoError(t, err)
    defer conn.Close()

    client := pb.NewBridgeServiceClient(conn)

    // Send CreateEntry request
    req := &pb.CreateEntryRequest{
        Key: &pb.DictKey{
            Type:  "CPF",
            Value: "12345678900",
        },
        Account: &pb.Account{
            Ispb:          "12345678",
            AccountNumber: "123456",
            Branch:        "0001",
            AccountType:   "CACC",
        },
    }

    resp, err := client.CreateEntry(context.Background(), req)
    require.NoError(t, err)

    // Verify response
    assert.NotEmpty(t, resp.EntryId)
    assert.Equal(t, "ACTIVE", resp.Status)
}

func TestExpiredCertificate(t *testing.T) {
    // TC-003-002: Expired Certificate

    // Load expired certificate
    cert, err := tls.LoadX509KeyPair(
        "/test-certs/expired-cert.pem",
        "/test-certs/expired-key.pem",
    )
    require.NoError(t, err)

    // Attempt to connect (should fail)
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        // ...
    }

    _, err = grpc.Dial(
        "localhost:50051",
        grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
    )

    // Verify error
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "certificate expired")
}

func TestXMLSignature(t *testing.T) {
    // TC-003-009: XML Signature - Valid Signature

    signer := NewXMLSigner("/etc/bridge/certs/client-cert.pem",
                            "/etc/bridge/certs/client-key.pem")

    xmlUnsigned := `<soapenv:Envelope ...>...</soapenv:Envelope>`

    // Sign XML
    xmlSigned, err := signer.Sign(xmlUnsigned)
    require.NoError(t, err)

    // Verify signature present
    assert.Contains(t, xmlSigned, "<ds:Signature")
    assert.Contains(t, xmlSigned, "<ds:SignatureValue>")
    assert.Contains(t, xmlSigned, "<ds:X509Certificate>")

    // Verify signature valid
    valid, err := signer.Verify(xmlSigned)
    require.NoError(t, err)
    assert.True(t, valid)
}
```

### Shell Script for Certificate Testing
```bash
#!/bin/bash
# test-certificate.sh

# TC-003-001: Test valid certificate

echo "Testing certificate validation..."

# Check certificate validity
openssl x509 -in /etc/bridge/certs/client-cert.pem -noout -text

# Verify certificate not expired
not_after=$(openssl x509 -in /etc/bridge/certs/client-cert.pem -noout -enddate | cut -d= -f2)
current_date=$(date)

echo "Certificate expires: $not_after"
echo "Current date: $current_date"

# Verify certificate chain
openssl verify -CAfile /etc/bridge/certs/ca-cert.pem \
  /etc/bridge/certs/client-cert.pem

if [ $? -eq 0 ]; then
  echo "✅ Certificate validation PASSED"
else
  echo "❌ Certificate validation FAILED"
  exit 1
fi

# Test mTLS connection
curl --cert /etc/bridge/certs/client-cert.pem \
     --key /etc/bridge/certs/client-key.pem \
     --cacert /etc/bridge/certs/ca-cert.pem \
     https://dict-test.bcb.gov.br/health

if [ $? -eq 0 ]; then
  echo "✅ mTLS connection PASSED"
else
  echo "❌ mTLS connection FAILED"
  exit 1
fi
```

---

## Entry/Exit Criteria

### Entry Criteria
- ✅ Bridge deployed to test environment
- ✅ Valid ICP-Brasil test certificates available
- ✅ Bacen test environment accessible
- ✅ Test certificates (valid, expired, invalid) prepared
- ✅ Network connectivity verified

### Exit Criteria
- ✅ All P0 tests passed (15/15)
- ✅ 90%+ P1 tests passed (5/6 minimum)
- ✅ No P0 security defects open
- ✅ Certificate validation working
- ✅ mTLS handshake working
- ✅ XML signature working

---

## Security Checklist

### Certificate Security
- ✅ Certificate is ICP-Brasil A3
- ✅ Certificate issued by trusted CA
- ✅ Certificate not expired
- ✅ Private key stored securely (HSM or encrypted)
- ✅ Certificate rotation process defined
- ✅ Certificate expiration monitoring enabled

### mTLS Security
- ✅ TLS 1.2+ enforced
- ✅ Strong cipher suites only
- ✅ Client certificate validation enabled
- ✅ Server certificate validation enabled
- ✅ Certificate pinning implemented
- ✅ No fallback to non-mTLS

### XML Signature Security
- ✅ RSA-SHA256 or stronger algorithm
- ✅ Signature covers entire body
- ✅ Canonicalization prevents tampering
- ✅ Certificate embedded in signature
- ✅ Signature validation on receive

---

## Monitoring During Tests

### Metrics
```prometheus
# Certificate validity
bridge_certificate_expiry_days

# mTLS connections
bridge_mtls_connections_total{status="success"}
bridge_mtls_connections_total{status="failed"}

# mTLS handshake duration
bridge_mtls_handshake_duration_seconds

# XML signature operations
bridge_xml_signature_duration_seconds{operation="sign"}
bridge_xml_signature_duration_seconds{operation="verify"}

# SOAP requests
bridge_soap_requests_total{operation="CreateEntry", status="success"}
bridge_soap_request_duration_seconds{operation="CreateEntry"}
```

### Alerts
```yaml
- alert: BridgeCertificateExpiringSoon
  expr: bridge_certificate_expiry_days < 30
  annotations:
    summary: "Bridge certificate expires in < 30 days"

- alert: BridgeMTLSHandshakeFailed
  expr: rate(bridge_mtls_connections_total{status="failed"}[5m]) > 0.05
  annotations:
    summary: "mTLS handshake failure rate > 5%"

- alert: BridgeXMLSignatureSlow
  expr: histogram_quantile(0.95, bridge_xml_signature_duration_seconds) > 0.1
  annotations:
    summary: "XML signature p95 latency > 100ms"
```

---

## Glossary

- **mTLS**: Mutual TLS - Both client and server authenticate with certificates
- **ICP-Brasil**: Brazilian Public Key Infrastructure
- **A3 Certificate**: Hardware-protected certificate (HSM or smart card)
- **XML-DSig**: XML Digital Signature standard
- **SOAP**: Simple Object Access Protocol
- **X509**: Standard for public key certificates
- **HSM**: Hardware Security Module
- **Certificate Pinning**: Hardcoding expected certificate fingerprint

---

**Última Revisão**: 2025-10-25
**Aprovado por**: QA Lead + Security Team
**Próxima Revisão**: Sprint review
