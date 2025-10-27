# Bridge Claim + Portability Implementation

**Data**: 2025-10-27
**Versão**: 1.0
**Status**: ✅ IMPLEMENTADO E COMPILANDO

---

## 📋 Resumo Executivo

Implementação completa das operações de **Claim** (reivindicação 30 dias) e **Portability** no **conn-bridge**, incluindo:
- 7 handlers gRPC (4 Claim + 3 Portability)
- Conversores XML bidirecionais (proto ↔ XML)
- Validação de requests
- Estrutura preparada para integração futura com XML Signer e SOAP Client

---

## 📁 Arquivos Criados/Modificados

### 1. Claim Handlers
**Arquivo**: `/Users/jose.silva.lb/LBPay/IA_Dict/conn-bridge/internal/grpc/claim_handlers.go`

**LOC**: 285 linhas

**RPCs Implementados**:
1. `CreateClaim` - Criar reivindicação (30 dias)
2. `GetClaim` - Consultar status de reivindicação
3. `CompleteClaim` - Completar reivindicação (transferir ownership)
4. `CancelClaim` - Cancelar reivindicação

**Validações Implementadas**:
- `validateCreateClaimRequest`: Valida entry_id, key_type, key_value, claimer_ispb, owner_ispb, claimer_account, completion_period_days=30
- `validateGetClaimRequest`: Valida claim_id OU external_id (oneof)
- `validateCompleteClaimRequest`: Valida claim_id OU external_id
- `validateCancelClaimRequest`: Valida claim_id OU external_id + cancellation_reason

---

### 2. Portability Handlers
**Arquivo**: `/Users/jose.silva.lb/LBPay/IA_Dict/conn-bridge/internal/grpc/portability_handlers.go`

**LOC**: 202 linhas

**RPCs Implementados**:
1. `InitiatePortability` - Iniciar portabilidade de chave
2. `ConfirmPortability` - Confirmar e completar portabilidade
3. `CancelPortability` - Cancelar portabilidade (reverter)

**Validações Implementadas**:
- `validateInitiatePortabilityRequest`: Valida entry_id, key (type + value), new_account (ispb, account_number, branch_code)
- `validateConfirmPortabilityRequest`: Valida entry_id, portability_id, new_account (ispb, account_number)
- `validateCancelPortabilityRequest`: Valida entry_id, portability_id, reason

---

### 3. XML Converters (Adicionados)
**Arquivo**: `/Users/jose.silva.lb/LBPay/IA_Dict/conn-bridge/internal/xml/converter.go`

**LOC Adicionadas**: ~230 linhas

**Funções Adicionadas**:

#### Claim Converters:
- `GetClaimRequestToXML(req *pb.GetClaimRequest) ([]byte, error)`
- `GetClaimResponseFromXML(xmlData []byte) (*pb.GetClaimResponse, error)`

#### Portability Converters:
- `InitiatePortabilityRequestToXML(req *pb.InitiatePortabilityRequest) ([]byte, error)`
- `InitiatePortabilityResponseFromXML(xmlData []byte) (*pb.InitiatePortabilityResponse, error)`
- `ConfirmPortabilityRequestToXML(req *pb.ConfirmPortabilityRequest) ([]byte, error)`
- `ConfirmPortabilityResponseFromXML(xmlData []byte) (*pb.ConfirmPortabilityResponse, error)`
- `CancelPortabilityRequestToXML(req *pb.CancelPortabilityRequest) ([]byte, error)`
- `CancelPortabilityResponseFromXML(xmlData []byte) (*pb.CancelPortabilityResponse, error)`

**Reutilização**:
- Funções de converter existentes (CreateClaimRequestToXML, CompleteClaimRequestToXML, CancelClaimRequestToXML) já existiam no código

---

## 🔄 Fluxo de Execução (Exemplo: CreateClaim)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         CreateClaim Flow                                 │
└─────────────────────────────────────────────────────────────────────────┘

1. RECEBER gRPC Request (Connect → Bridge)
   ┌─────────────────────────────────────────────────────┐
   │ CreateClaimRequest {                                 │
   │   entry_id: "entry-550e8400"                         │
   │   key_type: CPF                                      │
   │   key_value: "12345678900"                           │
   │   claimer_ispb: "12345678"                           │
   │   owner_ispb: "87654321"                             │
   │   claimer_account: { ... }                           │
   │   completion_period_days: 30                         │
   │ }                                                     │
   └─────────────────────────────────────────────────────┘

2. VALIDAR Request (claim_handlers.go)
   ✅ entry_id presente
   ✅ key_type != UNSPECIFIED
   ✅ key_value não vazio
   ✅ claimer_ispb presente
   ✅ owner_ispb presente
   ✅ claimer_account != nil
   ✅ completion_period_days == 30

3. CONVERTER Proto → XML (xml/converter.go)
   ┌─────────────────────────────────────────────────────┐
   │ <?xml version="1.0" encoding="UTF-8"?>               │
   │ <CreateClaimRequest>                                 │
   │   <Claim>                                            │
   │     <Type>OWNERSHIP</Type>                           │
   │     <Key>12345678900</Key>                           │
   │     <KeyType>CPF</KeyType>                           │
   │     <ClaimerAccount>                                 │
   │       <Participant>12345678</Participant>            │
   │       <Branch>0001</Branch>                          │
   │       <AccountNumber>123456</AccountNumber>          │
   │     </ClaimerAccount>                                │
   │     <CompletionPeriodEnd>2025-11-26T12:00:00Z</...  │
   │   </Claim>                                           │
   │ </CreateClaimRequest>                                │
   └─────────────────────────────────────────────────────┘

4. ASSINAR XML (XML Signer - TODO)
   ❌ NÃO IMPLEMENTADO (DEV MODE)
   ⚠️ Placeholder: xmlData usado sem assinatura

5. ENVIAR SOAP/mTLS para Bacen (SOAP Client - TODO)
   ❌ NÃO IMPLEMENTADO (DEV MODE)
   ⚠️ Placeholder: response fake

6. PARSEAR XML Response (xml/converter.go - TODO na integração real)
   ┌─────────────────────────────────────────────────────┐
   │ <?xml version="1.0" encoding="UTF-8"?>               │
   │ <CreateClaimResponse>                                │
   │   <Claim>                                            │
   │     <ClaimId>claim-bacen-12345</ClaimId>             │
   │     <Status>OPEN</Status>                            │
   │     <CompletionPeriodEnd>2025-11-26T12:00:00Z</...  │
   │   </Claim>                                           │
   │   <CorrelationId>tx-550e8400</CorrelationId>         │
   │ </CreateClaimResponse>                               │
   └─────────────────────────────────────────────────────┘

7. RETORNAR gRPC Response (Bridge → Connect)
   ┌─────────────────────────────────────────────────────┐
   │ CreateClaimResponse {                                │
   │   claim_id: "claim-1730037840000000000"              │
   │   external_id: "bacen-claim-1730037840000000000"     │
   │   status: CLAIM_STATUS_OPEN                          │
   │   completion_period_days: 30                         │
   │   expires_at: "2025-11-26T12:00:00Z"                 │
   │   created_at: "2025-10-27T12:00:00Z"                 │
   │   bacen_claim_id: "bacen-claim-id-1730037840000..." │
   │ }                                                     │
   └─────────────────────────────────────────────────────┘
```

---

## 📊 Mapeamento Proto ↔ XML

### CreateClaim

**Proto → XML**:
```go
// Proto
CreateClaimRequest {
  entry_id: "entry-550e8400"
  key_type: CPF
  key_value: "12345678900"
  claimer_ispb: "12345678"
  owner_ispb: "87654321"
  claimer_account: Account { ... }
  completion_period_days: 30
}

// XML
<CreateClaimRequest>
  <Claim>
    <Key>12345678900</Key>
    <KeyType>CPF</KeyType>
    <ClaimerAccount>...</ClaimerAccount>
    <!-- Outros campos -->
  </Claim>
</CreateClaimRequest>
```

**XML → Proto**:
```go
// XML Response
<CreateClaimResponse>
  <Claim>
    <ClaimId>bacen-claim-123</ClaimId>
    <Status>OPEN</Status>
  </Claim>
  <CorrelationId>tx-550e8400</CorrelationId>
</CreateClaimResponse>

// Proto Response
CreateClaimResponse {
  claim_id: "bacen-claim-123"
  external_id: "tx-550e8400"
  status: CLAIM_STATUS_OPEN
  completion_period_days: 30
  ...
}
```

---

### InitiatePortability

**Proto → XML**:
```go
// Proto
InitiatePortabilityRequest {
  entry_id: "entry-550e8400"
  key: DictKey { key_type: CPF, key_value: "12345678900" }
  new_account: Account { ispb: "99999999", ... }
}

// XML
<InitiatePortabilityRequest>
  <EntryId>entry-550e8400</EntryId>
  <Key>
    <Type>CPF</Type>
    <Value>12345678900</Value>
  </Key>
  <NewAccount>
    <Participant>99999999</Participant>
    <Branch>0001</Branch>
    <AccountNumber>654321</AccountNumber>
    <AccountType>CHECKING</AccountType>
  </NewAccount>
  <IdempotencyKey>idem-550e8400</IdempotencyKey>
  <RequestId>req-1730037840</RequestId>
</InitiatePortabilityRequest>
```

**XML → Proto**:
```go
// XML Response
<InitiatePortabilityResponse>
  <PortabilityId>port-bacen-456</PortabilityId>
  <EntryId>entry-550e8400</EntryId>
  <Status>PORTABILITY_PENDING</Status>
  <CorrelationId>tx-portability-550e8400</CorrelationId>
</InitiatePortabilityResponse>

// Proto Response
InitiatePortabilityResponse {
  portability_id: "port-bacen-456"
  entry_id: "entry-550e8400"
  status: ENTRY_STATUS_PORTABILITY_PENDING
  bacen_transaction_id: "tx-portability-550e8400"
  initiated_at: "2025-10-27T12:00:00Z"
}
```

---

## 🧪 Exemplos de XML SOAP (Bacen)

### CreateClaim Request (Bacen Format)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
                   xmlns:dict="https://www.bcb.gov.br/pi/dict/v1">
  <soapenv:Header>
    <wsse:Security xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
      <!-- XML Signature (ICP-Brasil A3) -->
      <ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
        <ds:SignedInfo>...</ds:SignedInfo>
        <ds:SignatureValue>...</ds:SignatureValue>
        <ds:KeyInfo>...</ds:KeyInfo>
      </ds:Signature>
    </wsse:Security>
  </soapenv:Header>

  <soapenv:Body>
    <dict:CreateClaimRequest>
      <dict:Claim>
        <dict:Type>PORTABILITY</dict:Type>
        <dict:Key>12345678900</dict:Key>
        <dict:KeyType>CPF</dict:KeyType>
        <dict:Status>OPEN</dict:Status>
        <dict:DonorParticipant>87654321</dict:DonorParticipant>
        <dict:ClaimerAccount>
          <dict:Participant>12345678</dict:Participant>
          <dict:Branch>0001</dict:Branch>
          <dict:AccountNumber>123456</dict:AccountNumber>
          <dict:AccountType>CACC</dict:AccountType>
        </dict:ClaimerAccount>
        <dict:Claimer>
          <dict:Type>NATURAL_PERSON</dict:Type>
          <dict:TaxIdNumber>12345678900</dict:TaxIdNumber>
          <dict:Name>João da Silva</dict:Name>
        </dict:Claimer>
        <dict:CompletionPeriodEnd>2025-11-26T23:59:59Z</dict:CompletionPeriodEnd>
        <dict:ResolutionPeriodEnd>2025-11-03T23:59:59Z</dict:ResolutionPeriodEnd>
      </dict:Claim>
    </dict:CreateClaimRequest>
  </soapenv:Body>
</soapenv:Envelope>
```

### CreateClaim Response (Bacen Format)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
                   xmlns:dict="https://www.bcb.gov.br/pi/dict/v1">
  <soapenv:Body>
    <dict:CreateClaimResponse>
      <dict:ResponseTime>2025-10-27T12:00:00.000Z</dict:ResponseTime>
      <dict:CorrelationId>bacen-tx-550e8400-abc123</dict:CorrelationId>
      <dict:Claim>
        <dict:ClaimId>claim-bacen-98765</dict:ClaimId>
        <dict:Type>PORTABILITY</dict:Type>
        <dict:Key>12345678900</dict:Key>
        <dict:KeyType>CPF</dict:KeyType>
        <dict:Status>OPEN</dict:Status>
        <dict:DonorParticipant>87654321</dict:DonorParticipant>
        <dict:ClaimerAccount>...</dict:ClaimerAccount>
        <dict:Claimer>...</dict:Claimer>
        <dict:CompletionPeriodEnd>2025-11-26T23:59:59Z</dict:CompletionPeriodEnd>
        <dict:ResolutionPeriodEnd>2025-11-03T23:59:59Z</dict:ResolutionPeriodEnd>
        <dict:CreationTime>2025-10-27T12:00:00.000Z</dict:CreationTime>
        <dict:LastModified>2025-10-27T12:00:00.000Z</dict:LastModified>
      </dict:Claim>
    </dict:CreateClaimResponse>
  </soapenv:Body>
</soapenv:Envelope>
```

---

## ✅ Status de Compilação

```bash
$ cd /Users/jose.silva.lb/LBPay/IA_Dict/conn-bridge
$ go build -o /tmp/conn-bridge ./cmd/bridge/main.go
✅ SUCESSO

$ ls -lh /tmp/conn-bridge
-rwxr-xr-x  1 jose.silva.lb  staff    31M Oct 27 12:04 /tmp/conn-bridge
```

**Binary**:
- Tamanho: 31 MB
- Tipo: Mach-O 64-bit executable arm64
- Status: Compilado com sucesso

---

## 📝 TODOs Restantes (Integração Futura)

### 1. XML Signer Integration
**Localização**: `claim_handlers.go` e `portability_handlers.go`

**Código Atual (Placeholder)**:
```go
// Step 2: Sign XML with ICP-Brasil A3
_ = xmlData // signedXML will be used when XML signer is integrated
s.logger.Warn("XML signing not yet implemented - using unsigned XML (DEV MODE)")
```

**Ação Necessária**:
- Integrar com Java XML Signer service (HTTP call ou JAR execution)
- Substituir `_ = xmlData` por `signedXML := xmlSigner.Sign(xmlData)`
- Passar `signedXML` para SOAP Client

**Referência**: `/Users/jose.silva.lb/LBPay/IA_Dict/conn-bridge/xml-signer/` (Java 17 + ICP-Brasil A3)

---

### 2. SOAP Client Integration
**Localização**: `claim_handlers.go` e `portability_handlers.go`

**Código Atual (Placeholder)**:
```go
// Step 3: Send signed XML to Bacen via SOAP/mTLS
s.logger.Info("SOAP call to Bacen not yet implemented - returning placeholder (DEV MODE)")
```

**Ação Necessária**:
- Reutilizar `HTTPClient` existente: `/Users/jose.silva.lb/LBPay/IA_Dict/conn-bridge/internal/infrastructure/bacen/http_client.go`
- Adicionar métodos específicos para Claim e Portability:
  - `CreateClaim(ctx, entry) (*XMLCreateClaimResponse, error)`
  - `GetClaim(ctx, claimId) (*XMLGetClaimResponse, error)`
  - `CompleteClaim(ctx, claimId) (*XMLCompleteClaimResponse, error)`
  - `CancelClaim(ctx, claimId) (*XMLCancelClaimResponse, error)`
  - `InitiatePortability(ctx, req) (*XMLInitiatePortabilityResponse, error)`
  - `ConfirmPortability(ctx, req) (*XMLConfirmPortabilityResponse, error)`
  - `CancelPortability(ctx, req) (*XMLCancelPortabilityResponse, error)`

**Configuração mTLS** (já implementada):
```go
tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{clientCert}, // ICP-Brasil A3
    RootCAs:      bacenCertPool,
    MinVersion:   tls.VersionTLS12,
}
```

---

### 3. Response Parsing (XML → Proto)
**Localização**: `xml/converter.go`

**Ação Necessária**:
- Testar parsers com responses reais do Bacen
- Ajustar mapeamento de campos se necessário
- Adicionar tratamento de erros SOAP Faults

**Exemplo SOAP Fault** (erro Bacen):
```xml
<soapenv:Fault>
  <faultcode>dict:ClaimAlreadyExists</faultcode>
  <faultstring>Claim already exists for this key</faultstring>
  <detail>
    <dict:ErrorDetail>
      <dict:ErrorCode>CLAIM_ALREADY_EXISTS</dict:ErrorCode>
      <dict:Message>A claim for CPF 12345678900 is already OPEN</dict:Message>
    </dict:ErrorDetail>
  </detail>
</soapenv:Fault>
```

---

## 📊 Métricas de Implementação

| Métrica | Valor |
|---------|-------|
| **Arquivos Criados** | 2 (claim_handlers.go, portability_handlers.go) |
| **Arquivos Modificados** | 1 (xml/converter.go) |
| **Total LOC Adicionadas** | ~720 linhas |
| **RPCs Implementados** | 7 (4 Claim + 3 Portability) |
| **Validações** | 7 funções de validação |
| **XML Converters** | 8 converters (request + response) |
| **Tempo de Compilação** | ~10s |
| **Tamanho do Binary** | 31 MB |

---

## 🎯 Próximos Passos

### Fase 1: Integração XML Signer (P0)
1. Criar gRPC client ou HTTP client para XML Signer service
2. Implementar método `Sign(xmlData []byte) ([]byte, error)`
3. Integrar nos handlers (substituir placeholders)
4. Testar assinatura com certificado ICP-Brasil A3 válido

### Fase 2: Integração SOAP Client (P0)
1. Adicionar métodos Claim/Portability ao HTTPClient
2. Configurar endpoints Bacen corretos (homologação vs produção)
3. Implementar retry logic com Circuit Breaker
4. Testar conectividade mTLS com Bacen

### Fase 3: Testes E2E (P1)
1. Criar mocks de Bacen DICT API
2. Escrever testes de integração Connect → Bridge → Mock Bacen
3. Validar cenários de sucesso e erro
4. Testar ClaimWorkflow (30 dias) no Temporal

### Fase 4: Homologação Bacen (P1)
1. Executar testes em ambiente de homologação Bacen
2. Validar conformidade com Manual DICT v3.1
3. Obter aprovação do Bacen
4. Deploy em produção

---

## 📞 Referências

- **Especificação Bridge**: `/Users/jose.silva.lb/LBPay/IA_Dict/Artefatos/00_Master/ESCOPO_BRIDGE_VALIDADO.md`
- **Manual DICT Bacen**: `/Users/jose.silva.lb/LBPay/IA_Dict/Artefatos/REG-001_Manual_Bacen_DICT_v3.1.md`
- **Proto Contracts**: `/Users/jose.silva.lb/LBPay/IA_Dict/dict-contracts/proto/bridge.proto`
- **SOAP Client**: `/Users/jose.silva.lb/LBPay/IA_Dict/conn-bridge/internal/infrastructure/bacen/http_client.go`
- **XML Structs**: `/Users/jose.silva.lb/LBPay/IA_Dict/conn-bridge/internal/xml/structs.go`

---

**Última Atualização**: 2025-10-27 12:05 BRT
**Autor**: Agente 2 (Backend Specialist)
**Status**: ✅ IMPLEMENTAÇÃO COMPLETA E COMPILANDO
**Próximo Milestone**: Integração XML Signer + SOAP Client
