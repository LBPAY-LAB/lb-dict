# Escopo RSFN Bridge - VALIDADO
**Data**: 2025-10-27 16:00 BRT
**Status**: ✅ VALIDADO com base em TEC-002 v3.1 + GRPC-001 + REG-001
**Versão**: 1.0

---

## 🎯 MISSÃO DO BRIDGE (Escopo Exato)

> **"Receber requisições gRPC/Pulsar → Preparar XML SOAP → Assinar digitalmente → Executar chamada SOAP/HTTP + mTLS → Retornar resposta"**

---

## ✅ RESPONSABILIDADES DO BRIDGE (O Que Faz)

### 1. **Interpretação de Função**
✅ Bridge **INTERPRETA** qual operação está sendo solicitada (CreateEntry, GetEntry, CreateClaim, etc.)
- Recebe request gRPC ou Pulsar
- Identifica qual endpoint Bacen chamar
- Roteia para o handler correto

### 2. **Transformação de Protocolo**
✅ Bridge **TRANSFORMA** gRPC/Pulsar → SOAP/XML
- Converte proto messages → XML structs
- Constrói envelope SOAP correto
- Valida schema XML

### 3. **Assinatura Digital XML**
✅ Bridge **ASSINA** XML com certificado ICP-Brasil A3
- Executa JAR Java (xml-signer) ou HTTP call para serviço Java
- Usa certificado ICP-Brasil A3 válido
- Garante integridade e autenticidade

### 4. **Execução de Chamada SOAP/HTTPS + mTLS**
✅ Bridge **EXECUTA** chamada **SÍNCRONA** ao Bacen
- POST HTTPS para endpoint Bacen
- mTLS (Mutual TLS) com certificados ICP-Brasil
- Header: `Content-Type: application/soap+xml; charset=utf-8`
- Timeout: 30s

### 5. **Parsing de Resposta**
✅ Bridge **PARSEIA** SOAP response do Bacen
- Deserializa XML → struct Go
- Converte struct Go → proto message
- Retorna para Connect via gRPC ou Pulsar

### 6. **Circuit Breaker**
✅ Bridge **PROTEGE** contra falhas em cascata
- sony/gobreaker
- OPEN após 5 falhas consecutivas
- HALF-OPEN após 30s
- Retry imediato (não durável)

---

## ❌ NÃO-RESPONSABILIDADES DO BRIDGE (O Que NÃO Faz)

### 1. ❌ **Lógica de Negócio**
Bridge **NÃO** valida regras de negócio complexas
- Validação de dados: responsabilidade do Connect
- Regras DICT: responsabilidade do Connect
- Exemplo: Bridge NÃO valida se CPF é válido, apenas envia para Bacen

### 2. ❌ **Gestão de Estado**
Bridge **NÃO** armazena estado
- Não tem banco de dados próprio
- Não persiste nada
- Não mantém sessões

### 3. ❌ **Orquestração de Workflows**
Bridge **NÃO** usa Temporal
- ClaimWorkflow (30 dias): responsabilidade do Connect
- VSYNC: responsabilidade do Connect
- Retry durável: responsabilidade do Connect

### 4. ❌ **Retry Durável**
Bridge **NÃO** faz retry durável
- Apenas retry imediato via Circuit Breaker
- Retry com persistência: responsabilidade do Connect (Temporal)

---

## 🌐 API DO BACEN DICT

### Protocolo
- **SOAP 1.2** over HTTPS
- **mTLS** (Mutual TLS com ICP-Brasil)
- **Formato**: XML assinado digitalmente

### Endpoints (51+ Operações)

#### **1. Directory (Vínculos DICT)** - 4 Operações PRINCIPAIS

| Operação | Endpoint Bacen | Método | Descrição |
|----------|----------------|--------|-----------|
| CreateEntry | `POST /dict/api/v1/entries` | POST | Criar chave PIX |
| GetEntry | `GET /dict/api/v1/entries/{key}` | GET | Consultar chave PIX |
| UpdateEntry | `PUT /dict/api/v1/entries/{key}` | PUT | Atualizar dados da conta |
| DeleteEntry | `DELETE /dict/api/v1/entries/{key}` | DELETE | Deletar chave PIX |

**Formato Request** (CreateEntry exemplo):
```xml
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Header>
    <wsse:Security>
      <!-- Signature XML com ICP-Brasil -->
    </wsse:Security>
  </soap:Header>
  <soap:Body>
    <CreateEntryRequest xmlns="http://www.bcb.gov.br/dict/api/v1">
      <Key>
        <Type>CPF</Type>
        <Value>12345678900</Value>
      </Key>
      <Account>
        <ISPB>12345678</ISPB>
        <AccountType>CACC</AccountType>
        <AccountNumber>123456</AccountNumber>
        <Branch>0001</Branch>
        <AccountHolder>João Silva</AccountHolder>
      </Account>
    </CreateEntryRequest>
  </soap:Body>
</soap:Envelope>
```

---

#### **2. Claim (Reivindicação de Posse)** - 4 Operações PRINCIPAIS

| Operação | Endpoint Bacen | Método | Descrição |
|----------|----------------|--------|-----------|
| CreateClaim | `POST /dict/api/v1/claims` | POST | Criar reivindicação (30 dias) |
| GetClaim | `GET /dict/api/v1/claims/{id}` | GET | Consultar status reivindicação |
| CompleteClaim | `PUT /dict/api/v1/claims/{id}/complete` | PUT | Completar reivindicação |
| CancelClaim | `PUT /dict/api/v1/claims/{id}/cancel` | PUT | Cancelar reivindicação |

---

#### **3. Portability (Portabilidade)** - 3 Operações

| Operação | Endpoint Bacen | Método | Descrição |
|----------|----------------|--------|-----------|
| InitiatePortability | `POST /dict/api/v1/portability` | POST | Iniciar portabilidade |
| ConfirmPortability | `PUT /dict/api/v1/portability/{id}/confirm` | PUT | Confirmar portabilidade |
| CancelPortability | `PUT /dict/api/v1/portability/{id}/cancel` | PUT | Cancelar portabilidade |

---

#### **4. Directory Queries (Consultas DICT)** - 2 Operações

| Operação | Endpoint Bacen | Método | Descrição |
|----------|----------------|--------|-----------|
| GetDirectory | `GET /dict/api/v1/directory` | GET | Consultar diretório completo |
| SearchEntries | `GET /dict/api/v1/entries/search` | GET | Buscar chaves por critérios |

---

#### **5. Health Check** - 1 Operação

| Operação | Endpoint Bacen | Método | Descrição |
|----------|----------------|--------|-----------|
| HealthCheck | `GET /dict/api/v1/health` | GET | Verificar disponibilidade API |

---

### Autenticação mTLS

**Certificados ICP-Brasil A3**:
- Certificado client (LBPay)
- Certificado CA (Bacen)
- Validação mútua (client ↔ server)

**Configuração TLS**:
```go
tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{clientCert},
    RootCAs:      bacenCertPool,
    MinVersion:   tls.VersionTLS12,
}
```

---

## 🔄 FLUXO COMPLETO (FrontEnd → Core → Connect → Bridge → Bacen)

```
┌──────────────┐         ┌──────────────┐         ┌──────────────┐         ┌──────────────┐         ┌─────────┐
│              │         │              │         │              │         │              │         │         │
│  FrontEnd    │  gRPC   │  Core DICT   │  gRPC/  │ Conn-Dict    │  gRPC   │  Conn-Bridge │  SOAP/  │  Bacen  │
│  (Cliente)   │ ──────> │  (TEC-001)   │  Pulsar │  (TEC-003)   │ ──────> │  (TEC-002)   │  mTLS   │  DICT   │
│              │         │              │ ──────> │              │         │              │ ──────> │  API    │
│              │         │              │         │              │         │              │         │         │
└──────────────┘         └──────────────┘         └──────────────┘         └──────────────┘         └─────────┘
   User Action           Valida Negócio          Orquestra                Adapta Protocolo         Processa
                         Persiste DB             Workflows                SOAP + mTLS              Autoriza
                                                 Temporal
```

### Exemplo: Criar Chave PIX

**1. FrontEnd → Core DICT** (gRPC):
```protobuf
CreateKeyRequest {
  key_type: CPF
  key_value: "12345678900"
  account_id: "acc-550e8400"
}
```

**2. Core DICT → Connect** (Pulsar Topic: `dict.entries.created`):
```json
{
  "entry_id": "entry-550e8400",
  "key_type": "CPF",
  "key_value": "12345678900",
  "account": {...}
}
```

**3. Connect → Bridge** (gRPC):
```protobuf
CreateEntryRequest {
  key: {type: CPF, value: "12345678900"}
  account: {ispb: "12345678", ...}
  idempotency_key: "idem-550e8400"
}
```

**4. Bridge → Bacen** (SOAP/mTLS):
```xml
<soap:Envelope>
  <soap:Header>
    <wsse:Security>
      <!-- XML Signature ICP-Brasil -->
    </wsse:Security>
  </soap:Header>
  <soap:Body>
    <CreateEntryRequest>
      <Key><Type>CPF</Type><Value>12345678900</Value></Key>
      <Account>...</Account>
    </CreateEntryRequest>
  </soap:Body>
</soap:Envelope>
```

**5. Bacen → Bridge** (SOAP Response):
```xml
<CreateEntryResponse>
  <EntryId>bacen-entry-id</EntryId>
  <Status>ACTIVE</Status>
  <TransactionId>tx-550e8400</TransactionId>
</CreateEntryResponse>
```

**6. Bridge → Connect** (gRPC):
```protobuf
CreateEntryResponse {
  entry_id: "entry-550e8400"
  external_id: "bacen-entry-id"
  status: ACTIVE
  bacen_transaction_id: "tx-550e8400"
}
```

**7. Connect → Core** (Pulsar Topic: `dict.entries.status.changed`):
```json
{
  "entry_id": "entry-550e8400",
  "old_status": "PENDING",
  "new_status": "ACTIVE"
}
```

---

## 🔧 IMPLEMENTAÇÃO DO BRIDGE

### Estrutura de Arquivos (Foco nos Essenciais)

```
conn-bridge/
├── cmd/
│   └── server/
│       └── main.go                     # Entrypoint
│
├── internal/
│   ├── grpc/
│   │   ├── server.go                   # gRPC Server setup
│   │   ├── entry_handlers.go           # 4 RPCs Entry (CreateEntry, GetEntry, UpdateEntry, DeleteEntry)
│   │   ├── claim_handlers.go           # 4 RPCs Claim (CreateClaim, GetClaim, CompleteClaim, CancelClaim)
│   │   ├── portability_handlers.go     # 3 RPCs Portability
│   │   ├── directory_handlers.go       # 2 RPCs Directory
│   │   └── health_handler.go           # 1 RPC HealthCheck
│   │
│   ├── infrastructure/
│   │   ├── bacen/
│   │   │   ├── soap_client.go          # HTTP client com mTLS
│   │   │   ├── soap_builder.go         # Constrói envelopes SOAP
│   │   │   ├── soap_parser.go          # Parseia SOAP responses
│   │   │   └── circuit_breaker.go      # Circuit Breaker
│   │   │
│   │   └── signer/
│   │       └── xml_signer_client.go    # Chama Java service
│   │
│   └── xml/
│       ├── structs.go                  # Structs XML (Go)
│       └── converter.go                # proto ↔ XML
│
└── xml-signer/                         # ✅ Java Service (JÁ PRONTO)
    └── src/main/java/...               # Assinatura ICP-Brasil A3
```

---

### Checklist de Implementação

#### **Fase 1: Entry Operations** (P0 - Prioridade Máxima)
- [ ] Implementar `CreateEntry` handler completo
  - [ ] Converter proto → XML
  - [ ] Chamar XML Signer
  - [ ] Executar SOAP/mTLS call
  - [ ] Parsear response
- [ ] Implementar `GetEntry` handler
- [ ] Implementar `UpdateEntry` handler
- [ ] Implementar `DeleteEntry` handler
- [ ] Testar E2E: Connect → Bridge → Bacen (mock)

#### **Fase 2: Claim Operations** (P1)
- [ ] Implementar `CreateClaim` handler
- [ ] Implementar `GetClaim` handler
- [ ] Implementar `CompleteClaim` handler
- [ ] Implementar `CancelClaim` handler

#### **Fase 3: Portability + Directory + Health** (P2)
- [ ] Implementar 3 Portability handlers
- [ ] Implementar 2 Directory handlers
- [ ] Implementar HealthCheck handler

#### **Fase 4: Infraestrutura**
- [ ] SOAP Client completo com mTLS
- [ ] XML Signer integration (Go → Java)
- [ ] Circuit Breaker production-ready
- [ ] Testes compilando e passando

---

## 🎓 VALIDAÇÃO DO ESCOPO

### ✅ **Confirmações**

1. ✅ **Bridge é adaptador puro** - Sem lógica de negócio
2. ✅ **Bridge interpreta função** - Sabe qual endpoint Bacen chamar para cada operação
3. ✅ **Bridge faz chamadas síncronas** - SOAP/HTTPS com mTLS
4. ✅ **Bridge assina XML** - ICP-Brasil A3 via Java service
5. ✅ **API Bacen é SOAP/XML** - Não é REST puro, é SOAP over HTTPS
6. ✅ **14 RPCs no proto bridge.proto** - Alinhado com especificação
7. ✅ **XML Signer já existe** - Java 17 + ICP-Brasil pronto

### ⚠️ **Correções de Entendimento**

1. ❌ **API Bacen NÃO é REST pura** → ✅ É **SOAP over HTTPS**
2. ❌ **Bridge NÃO chama "REST API"** → ✅ Chama **SOAP API com mTLS**
3. ✅ **Mas é HTTP POST** para endpoint REST-like (`/dict/api/v1/entries`)
4. ✅ **Payload é XML SOAP**, não JSON

---

## 📞 PRÓXIMOS PASSOS

### Decisão Validada
✅ **Implementar conn-bridge AGORA** com escopo 100% claro

### Estratégia de Implementação
**3 Agentes em Paralelo**:

1. **Agente 1**: Entry Operations + SOAP Client (8h)
2. **Agente 2**: Claim + Portability Operations (14h)
3. **Agente 3**: Directory + Health + Tests (9h)

**Tempo Total**: ~17-18h com paralelismo

---

## ✅ VALIDAÇÃO FINAL

**Escopo do Bridge está 100% CLARO**:
- ✅ Responsabilidades definidas
- ✅ API Bacen documentada (SOAP/mTLS)
- ✅ Fluxo E2E mapeado
- ✅ Arquivos a implementar listados
- ✅ Checklist de implementação pronto

**PODE INICIAR IMPLEMENTAÇÃO AGORA!**

---

**Última Atualização**: 2025-10-27 16:00 BRT
**Status**: ✅ **VALIDADO E PRONTO PARA IMPLEMENTAÇÃO**
**Aprovação**: Aguardando confirmação do usuário
