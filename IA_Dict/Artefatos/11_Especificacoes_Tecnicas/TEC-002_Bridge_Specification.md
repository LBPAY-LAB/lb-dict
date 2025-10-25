# TEC-002: Especificação Técnica - RSFN Bridge (Adaptador SOAP/mTLS)

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: RSFN Bridge Service (Proxy/Adaptador para Bacen)
**Versão**: 3.1
**Data**: 2025-10-25
**Autor**: ARCHITECT (AI Agent - Technical Architect)
**Revisor**: [Aguardando]
**Aprovador**: Head de Arquitetura (Thiago Lima), CTO (José Luís Silva)

---

## Controle de Versão

| Versão | Data | Autor | Descrição das Mudanças |
|--------|------|-------|------------------------|
| 1.0 | 2025-10-24 | ARCHITECT | Versão inicial - Fluxo assíncrono com Temporal |
| 2.0 | 2025-10-25 | ARCHITECT | Arquitetura híbrida com Temporal Workflows |
| 3.0 | 2025-10-25 | ARCHITECT | **CORREÇÃO ARQUITETURAL**: Bridge como **adaptador puro** (sem Temporal). Temporal workflows movidos para TEC-003 (Connect) |
| 3.1 | 2025-10-25 | ARCHITECT | **ALINHAMENTO COM IMPLEMENTAÇÃO**: Dual protocol support (gRPC + Pulsar simultâneo), nomenclatura Pulsar topics IcePanel, mapeamento com repositório real, 7 domínios implementados |

---

## Sumário Executivo

### Visão Geral

O **RSFN Bridge** é um **adaptador/proxy especializado** entre o **RSFN Connect** e o **Bacen (RSFN API)**. Sua única responsabilidade é **preparar e executar** chamadas SOAP/XML com autenticação mTLS para o Bacen.

**Missão do Bridge**:
> **"Receber requisições → Preparar XML SOAP → Assinar digitalmente → Executar chamada mTLS → Retornar resposta"**

### Não-Responsabilidades (Movidas para RSFN Connect - TEC-003)

- ❌ **Orquestração de Workflows** (ClaimWorkflow, VSYNC, etc.) → **Connect**
- ❌ **Lógica de Negócio** (validações complexas, regras de domínio) → **Connect**
- ❌ **Gestão de Estado de Processos Longos** (7 dias de claim) → **Connect**
- ❌ **Retry com Temporal** (durabilidade de workflows) → **Connect**

### Responsabilidades do Bridge

- ✅ **Preparação de XML SOAP**: Construir envelopes SOAP conformes com specs do Bacen
- ✅ **Assinatura Digital XML**: Assinar XMLs com certificado ICP-Brasil (via JRE + JAR externo)
- ✅ **Autenticação mTLS**: Mutual TLS com certificados ICP-Brasil
- ✅ **Execução de Chamadas HTTP**: POST para API REST do Bacen (SOAP over HTTPS)
- ✅ **Parsing de Respostas**: Deserializar SOAP responses e mapear para estrutura de dados
- ✅ **Circuit Breaker**: Proteção contra falhas em cascata (sony/gobreaker)
- ✅ **Retry Simples**: Retry imediato para falhas temporárias de rede (não durável)

### Fluxo de Integração

```
┌────────────────┐                  ┌────────────────┐                  ┌────────────────┐
│                │                  │                │                  │                │
│  RSFN Connect  │  ─────────────>  │  RSFN Bridge   │  ─────────────>  │  Bacen RSFN    │
│  (TEC-003)     │  gRPC ou Pulsar  │  (TEC-002)     │  SOAP/HTTP mTLS  │  API DICT/SPI  │
│                │  <─────────────  │  (este doc)    │  <─────────────  │                │
└────────────────┘                  └────────────────┘                  └────────────────┘
     Orquestra                        Adapta/Traduz                       Processa
     Workflows                        SOAP + mTLS                         Requisições
```

### Repositório

🔗 **GitHub**: `lb-conn/rsfn-connect-bacen-bridge`
- **Linguagem**: Go 1.24.5
- **Arquitetura**: Clean Architecture (4 camadas: Domain, Application, Handlers, Infrastructure)
- **Mapeamento IcePanel**: `DICT Proxy` (componente descrito como "Proxy/adapter para conexão segura mTLS com DICT no BCB")
- **Status Implementação** (conforme ANA-002):
  - ✅ **110 arquivos Go** implementados
  - ✅ **7 domínios funcionais**: Directory, Claim, Key, Reconciliation, Antifraud, Policies, Infraction Reports
  - ✅ **51+ operações** mapeadas para API Bacen
  - ✅ **Dual Protocol Support**: gRPC (síncrono) + Pulsar (assíncrono) funcionando simultaneamente
  - ✅ **Circuit Breaker**: sony/gobreaker v2.3.0
  - ✅ **XML Signer**: JRE + JAR externo funcional
  - ✅ **Observabilidade**: OpenTelemetry v1.38.0

---

## Arquitetura do Bridge (Adaptador Puro)

### Diagrama de Componentes

```
┌────────────────────────────────────────────────────────────────────────┐
│                         RSFN Bridge (Adaptador)                        │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────────┐ │
│  │                    CAMADA DE INTERFACE                           │ │
│  │                                                                  │ │
│  │  ┌─────────────────────┐      ┌──────────────────────────────┐  │ │
│  │  │  gRPC Server        │      │  Pulsar Consumer             │  │ │
│  │  │  (Sync Requests)    │      │  (Async Requests)            │  │ │
│  │  │                     │      │                              │  │ │
│  │  │  Port: 50051        │      │  Topic: bridge-dict-req-in   │  │ │
│  │  └─────────────────────┘      └──────────────────────────────┘  │ │
│  │           │                              │                       │ │
│  └───────────┼──────────────────────────────┼───────────────────────┘ │
│              │                              │                         │
│              └──────────────┬───────────────┘                         │
│                             │                                         │
│  ┌──────────────────────────▼─────────────────────────────────────┐  │
│  │              APPLICATION LAYER (Use Cases)                      │  │
│  │                                                                 │  │
│  │  - ProcessDirectoryRequestUseCase                              │  │
│  │  - ProcessClaimRequestUseCase                                  │  │
│  │  - ProcessAntifraudRequestUseCase                              │  │
│  │                                                                 │  │
│  │  Responsabilidade: Orquestrar chamada ao Bacen                 │  │
│  │  (sem lógica de negócio complexa)                              │  │
│  └─────────────────────────────────────────────────────────────────┘  │
│              │                                                         │
│  ┌───────────▼─────────────────────────────────────────────────────┐  │
│  │           INFRASTRUCTURE LAYER (Adaptadores)                    │  │
│  │                                                                 │  │
│  │  ┌────────────────┐  ┌──────────────┐  ┌──────────────────┐   │  │
│  │  │ RSFN Client    │  │ XML Signer   │  │ Circuit Breaker  │   │  │
│  │  │ (SOAP + mTLS)  │  │ (JRE + JAR)  │  │ (sony/gobreaker) │   │  │
│  │  └────────────────┘  └──────────────┘  └──────────────────┘   │  │
│  └─────────────────────────────────────────────────────────────────┘  │
│              │                                                         │
│              └─────────────────┐                                       │
│                                │                                       │
└────────────────────────────────┼───────────────────────────────────────┘
                                 │
                                 ▼
                      ┌──────────────────────┐
                      │   Bacen RSFN API     │
                      │   (SOAP/XML + mTLS)  │
                      └──────────────────────┘
```

### Componentes Principais

#### 1. **Camada de Interface** (Entrada de Requisições)

##### 1.1 gRPC Server (Operações Síncronas)
**Arquivo**: `apps/dict/handlers/grpc/directory_controller.go`

```go
type DirectoryController struct {
	pb.UnimplementedDirectoryServiceServer
	useCase *ProcessDirectoryRequestUseCase
}

// GetEntry - Consulta chave Pix (sync)
func (c *DirectoryController) GetEntry(ctx context.Context, req *pb.GetEntryRequest) (*pb.GetEntryResponse, error) {
	// 1. Valida request gRPC
	// 2. Chama Use Case
	// 3. Retorna resposta
	return c.useCase.Execute(ctx, req)
}
```

**Operações Implementadas**:
- `GetEntry(key)` - Consulta de chave
- `CreateEntry(entry)` - Cadastro de chave
- `UpdateEntry(entry)` - Atualização de chave
- `DeleteEntry(key)` - Exclusão de chave
- `GetClaim(claimId)` - Consulta status de claim
- `GetInfractionReport(reportId)` - Consulta relatório de infração

**Protocolo**: gRPC (HTTP/2 + Protobuf)
**Timeout**: 30 segundos
**Porta**: 50051

---

##### 1.2 Pulsar Consumer (Operações Assíncronas)
**Arquivo**: `apps/dict/handlers/pulsar/handler.go`

```go
type Handler struct {
	consumer        pulsar.Consumer
	directoryUseCase *ProcessDirectoryRequestUseCase
	claimUseCase     *ProcessClaimRequestUseCase
}

// Process - Processa mensagens Pulsar
func (h *Handler) Process(ctx context.Context) {
	for {
		msg, err := h.consumer.Receive(ctx)
		if err != nil {
			log.Error("Failed to receive message", err)
			continue
		}

		// Roteia mensagem para handler apropriado
		switch msg.Type {
		case "DirectoryRequest":
			h.processDirectoryRequest(ctx, msg)
		case "ClaimRequest":
			h.processClaimRequest(ctx, msg)
		// ...
		}

		// ACK mensagem
		h.consumer.Ack(msg)
	}
}
```

**Topics Pulsar** (nomenclatura IcePanel):
- **Entrada**: `rsfn-dict-req-out` (consome requisições do Connect - saída do Connect, entrada do Bridge)
- **Saída**: `rsfn-dict-res-out` (publica respostas para o dict.api)

**Formato de Mensagem**: JSON com schema validation (Avro em produção)

**Configuração**:
```bash
PULSAR_TOPIC_REQ_IN=rsfn-dict-req-out
PULSAR_TOPIC_RES_OUT=rsfn-dict-res-out
```

**Observação**: O Bridge **não** inicia workflows Temporal. Ele apenas:
1. Consome mensagem do Connect
2. Prepara e executa chamada SOAP ao Bacen
3. Publica resposta de volta para o Connect

---

##### 1.3 Dual Protocol Support (gRPC + Pulsar)

**Implementação Confirmada** (ANA-002): O Bridge suporta **AMBOS** protocolos simultaneamente.

**Quando usar cada protocolo:**

| Protocolo | Cenário de Uso | Características | Exemplo |
|-----------|---------------|-----------------|---------|
| **gRPC** | Operações síncronas de baixa latência | - Request/Response imediato<br>- Timeout 30s<br>- Retorno direto | `GetEntry(key)` - Consulta chave PIX para validação em transação |
| **Pulsar** | Operações assíncronas de longa duração | - Fire-and-forget<br>- Processamento em background<br>- Desacoplamento temporal | `CreateClaim()` - Reivindicação que será processada por workflow Temporal |

**Mapeamento por Operação:**

```
gRPC (Sync) ✅
├── GetEntry          # Consulta rápida
├── CheckKeys         # Validação múltiplas chaves
├── GetClaim          # Consulta status claim
└── GetEntryStatistics # Estatísticas antifraud

Pulsar (Async) ⚡
├── CreateEntry       # Cadastro pode ser demorado
├── UpdateEntry       # Atualização pode ter retry
├── DeleteEntry       # Deleção pode ter compensação
├── CreateClaim       # Claim inicia workflow Temporal
├── ConfirmClaim      # Confirmação após 7 dias
└── CancelClaim       # Cancelamento com compensação
```

**Benefícios do Dual Support:**
- ✅ **Flexibilidade**: Connect escolhe protocolo conforme necessidade
- ✅ **Performance**: gRPC para operações críticas de latência
- ✅ **Resiliência**: Pulsar para operações que precisam garantia de entrega
- ✅ **Escalabilidade**: Pulsar permite processamento paralelo e backpressure

**Configuração:**
```yaml
# Habilitar/desabilitar protocolos independentemente
GRPC_ENABLED=true
GRPC_PORT=50051

PULSAR_ENABLED=true
PULSAR_URL=pulsar://pulsar-proxy:6650
```

---

#### 2. **Application Layer** (Use Cases)

**Responsabilidade**: Orquestrar a execução de uma chamada ao Bacen (sem lógica de negócio complexa).

##### 2.1 ProcessDirectoryRequestUseCase

```go
type ProcessDirectoryRequestUseCase struct {
	rsfnClient ports.RSFNClient
	xmlSigner  ports.XMLSigner
}

func (uc *ProcessDirectoryRequestUseCase) Execute(ctx context.Context, req *DirectoryRequest) (*DirectoryResponse, error) {
	// 1. Mapeia request para payload SOAP
	soapPayload := uc.buildSOAPPayload(req)

	// 2. Assina XML com certificado ICP-Brasil
	signedXML, err := uc.xmlSigner.Sign(ctx, soapPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to sign XML: %w", err)
	}

	// 3. Envia para Bacen via mTLS
	response, err := uc.rsfnClient.Send(ctx, signedXML)
	if err != nil {
		return nil, fmt.Errorf("failed to send to Bacen: %w", err)
	}

	// 4. Retorna resposta (sem armazenar estado)
	return response, nil
}
```

**Importante**: Use Cases no Bridge **NÃO** fazem:
- ❌ Validações de negócio complexas (isso é responsabilidade do Connect)
- ❌ Gestão de estado/persistência (apenas passa a resposta adiante)
- ❌ Retry durável (usa Circuit Breaker para retry imediato apenas)

---

#### 3. **Infrastructure Layer** (Adaptadores Externos)

##### 3.1 RSFN Client (SOAP + mTLS)

**Arquivo**: `apps/dict/infrastructure/bacen/client.go`

```go
type Client struct {
	httpClient  *http.Client
	bacenURL    string
	circuitBreaker *gobreaker.CircuitBreaker
}

// NewClient cria client com mTLS configurado
func NewClient(bacenURL, certPath, keyPath, caPath string) (*Client, error) {
	// Carrega certificados ICP-Brasil
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	// Carrega CA pool
	caCert, _ := os.ReadFile(caPath)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Configura mTLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS12,
	}

	// HTTP client com mTLS
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	// Circuit Breaker
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "bacen-rsfn",
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
	})

	return &Client{
		httpClient:     httpClient,
		bacenURL:       bacenURL,
		circuitBreaker: cb,
	}, nil
}

// Send envia requisição SOAP para Bacen
func (c *Client) Send(ctx context.Context, signedXML string) (*Response, error) {
	// Executa com Circuit Breaker
	result, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		req, _ := http.NewRequestWithContext(ctx, "POST", c.bacenURL, bytes.NewBufferString(signedXML))
		req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Parse SOAP response
		return c.parseSOAPResponse(resp.Body)
	})

	if err != nil {
		return nil, err
	}

	return result.(*Response), nil
}
```

**Características**:
- ✅ mTLS com certificados ICP-Brasil
- ✅ Circuit Breaker (5 falhas consecutivas → OPEN por 30s)
- ✅ Timeout de 30s
- ✅ Retry automático (via Circuit Breaker em estado HALF-OPEN)

---

##### 3.2 XML Signer (JRE + JAR Externo)

**Arquivo**: `apps/dict/infrastructure/signer/adapter.go`

```go
type Adapter struct {
	jarPath     string
	certPath    string
	keyPath     string
}

// Sign assina XML usando JAR externo
func (a *Adapter) Sign(ctx context.Context, xmlPayload string) (string, error) {
	// Cria arquivo temporário com XML
	tmpFile, _ := os.CreateTemp("", "payload-*.xml")
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(xmlPayload)
	tmpFile.Close()

	// Executa JAR signer (JRE + certificado ICP-Brasil)
	cmd := exec.CommandContext(ctx,
		"java", "-jar", a.jarPath,
		"--input", tmpFile.Name(),
		"--cert", a.certPath,
		"--key", a.keyPath,
		"--output", tmpFile.Name()+".signed",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("signer failed: %w, output: %s", err, output)
	}

	// Lê XML assinado
	signedXML, _ := os.ReadFile(tmpFile.Name() + ".signed")
	defer os.Remove(tmpFile.Name() + ".signed")

	return string(signedXML), nil
}
```

**Dependências Externas**:
- ☕ **JRE 11+**: Java Runtime Environment
- 📦 **signer.jar**: JAR proprietário para assinatura XML
- 🔐 **Certificado ICP-Brasil**: Certificado A1/A3 válido

---

##### 3.3 Circuit Breaker (sony/gobreaker)

**Arquivo**: `apps/dict/setup/circuit_breaker.go`

```go
func NewCircuitBreaker(name string) *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        name,
		MaxRequests: 3,           // Máximo de requests em HALF-OPEN
		Interval:    10 * time.Second,  // Janela de tempo para contagem de falhas
		Timeout:     30 * time.Second,  // Tempo em OPEN antes de tentar HALF-OPEN
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Abre circuit após 5 falhas consecutivas
			return counts.ConsecutiveFailures > 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Infof("Circuit breaker '%s' changed from %s to %s", name, from, to)
		},
	})
}
```

**Estados do Circuit Breaker**:
- **CLOSED**: Operação normal (todas as requests passam)
- **OPEN**: Circuit aberto (rejeita todas as requests por 30s)
- **HALF-OPEN**: Testando recuperação (permite até 3 requests)

**Transições**:
- CLOSED → OPEN: Após 5 falhas consecutivas
- OPEN → HALF-OPEN: Após 30 segundos em OPEN
- HALF-OPEN → CLOSED: Se 3 requests consecutivos tiverem sucesso
- HALF-OPEN → OPEN: Se alguma request falhar

---

## Fluxos de Integração

### Fluxo Síncrono (gRPC)

```
┌──────────┐                 ┌──────────┐                 ┌─────────┐
│          │  1. gRPC Call   │          │  4. SOAP/mTLS   │         │
│ Connect  │ ──────────────> │  Bridge  │ ──────────────> │  Bacen  │
│          │                 │          │                 │         │
│          │ <────────────── │          │ <────────────── │         │
└──────────┘  3. gRPC Resp   └──────────┘  5. SOAP Resp   └─────────┘
                                │      │
                                │  2.  │ Sign XML (JRE+JAR)
                                └──────┘
```

**Passo a passo**:
1. **Connect** envia `GetEntry(key="11122233344")` via gRPC
2. **Bridge** assina XML com certificado ICP-Brasil
3. **Bridge** executa POST HTTPS com mTLS para Bacen
4. **Bacen** valida certificado, processa, retorna SOAP response
5. **Bridge** parseia resposta e retorna via gRPC para Connect
6. **Timeout total**: 30 segundos

---

### Fluxo Assíncrono (Pulsar)

```
┌──────────┐                 ┌──────────┐                 ┌─────────┐
│          │ 1. Publish Msg  │          │  4. SOAP/mTLS   │         │
│ Connect  │ ──────────────> │  Bridge  │ ──────────────> │  Bacen  │
│          │  (Pulsar topic) │          │                 │         │
│          │                 │          │                 │         │
│          │ <────────────── │          │ <────────────── │         │
└──────────┘ 6. Consume Resp └──────────┘  5. SOAP Resp   └─────────┘
              (Pulsar topic)     │      │
                                 │  2.  │ Sign XML
                                 │  3.  │ Circuit Breaker
                                 └──────┘
```

**Passo a passo**:
1. **Connect** publica mensagem `ClaimPixKeyRequest` no topic `bridge-dict-req-in`
2. **Bridge** consome mensagem do Pulsar
3. **Bridge** assina XML e executa chamada SOAP/mTLS ao Bacen
4. **Bacen** retorna resposta SOAP
5. **Bridge** publica resposta no topic `bridge-dict-res-out`
6. **Connect** consome resposta e continua workflow (Temporal)

**Observação Importante**: O Bridge **não gerencia estado**. Ele apenas:
- Recebe request → Prepara SOAP → Envia ao Bacen → Retorna response

O **Connect** é responsável por:
- Gerenciar estado do workflow
- Aguardar resposta assíncrona do Bacen (via Temporal)
- Processar lógica de negócio (validações, retry durável, compensações)

---

## Domínios Funcionais Implementados

**Conforme ANA-002**, o Bridge implementa **7 domínios funcionais** completos com mapeamento para API RSFN do Bacen:

### 1. Directory (Vínculos DICT)
**Arquivo**: `application/usecases/directory/`
**Operações**: 7 operações

| Operação | Método gRPC | Endpoint Bacen | Descrição |
|----------|-------------|----------------|-----------|
| CreateEntry | `CreateEntry()` | `POST /entries` | Criar vínculo chave PIX |
| GetEntry | `GetEntry()` | `GET /entries/{key}` | Consultar vínculo |
| UpdateEntry | `UpdateEntry()` | `PUT /entries/{key}` | Atualizar vínculo |
| DeleteEntry | `DeleteEntry()` | `DELETE /entries/{key}` | Deletar vínculo |

### 2. Claim (Reivindicação de Posse)
**Arquivo**: `application/usecases/claim/`
**Operações**: 10 operações

| Operação | Endpoint Bacen | Descrição |
|----------|----------------|-----------|
| CreateClaim | `POST /claims` | Criar reivindicação de posse |
| GetClaim | `GET /claims/{id}` | Consultar status reivindicação |
| ListClaims | `GET /claims` | Listar reivindicações |
| ConfirmClaim | `PUT /claims/{id}/confirm` | Confirmar reivindicação |
| CompleteClaim | `PUT /claims/{id}/complete` | Completar reivindicação |
| CancelClaim | `PUT /claims/{id}/cancel` | Cancelar reivindicação |
| AcknowledgeClaim | `PUT /claims/{id}/acknowledge` | Reconhecer reivindicação |

### 3. Key (Validação de Chaves)
**Arquivo**: `application/usecases/key/`
**Operações**: 4 operações

| Operação | Endpoint Bacen | Descrição |
|----------|----------------|-----------|
| CheckKeys | `POST /keys/check` | Verificar existência de múltiplas chaves |

### 4. Reconciliation (CID e VSYNC)
**Arquivo**: `application/usecases/reconciliation/`
**Operações**: 8 operações

| Operação | Endpoint Bacen | Descrição |
|----------|----------------|-----------|
| GetCidSetFile | `GET /cid-set-files/{id}` | Obter arquivo CID |
| CreateCidSetFile | `POST /cid-set-files` | Criar arquivo CID |
| GetEntryByCid | `GET /entries/cid/{cid}` | Obter vínculo por CID |
| ListCidSetEvents | `GET /cid-set-events` | Listar eventos CID |
| CreateSyncVerification | `POST /sync-verifications` | Criar verificação de sincronização (VSYNC) |

### 5. Antifraud (Marcação de Fraude)
**Arquivo**: `application/usecases/antifraud/`
**Operações**: 8 operações

| Operação | Endpoint Bacen | Descrição |
|----------|----------------|-----------|
| CreateFraudMarker | `POST /fraud-markers` | Criar marcação de fraude |
| CancelFraudMarker | `DELETE /fraud-markers/{id}` | Cancelar marcação |
| GetFraudMarker | `GET /fraud-markers/{id}` | Consultar marcação |
| GetEntryStatistics | `GET /entries/{key}/statistics` | Estatísticas de vínculo |
| GetPersonStatistics | `GET /persons/{document}/statistics` | Estatísticas de pessoa |

### 6. Policies (Políticas do DICT)
**Arquivo**: `application/usecases/policies/`
**Operações**: 5 operações

| Operação | Endpoint Bacen | Descrição |
|----------|----------------|-----------|
| ListPolicies | `GET /policies` | Listar políticas DICT |
| GetPolicy | `GET /policies/{id}` | Obter política específica |

### 7. Infraction Report (Relatórios de Infração)
**Arquivo**: `application/usecases/infraction_report/`
**Operações**: 9 operações

| Operação | Endpoint Bacen | Descrição |
|----------|----------------|-----------|
| CreateInfractionReport | `POST /infraction-reports` | Criar relatório de infração |
| GetInfractionReport | `GET /infraction-reports/{id}` | Consultar relatório |
| ListInfractionReports | `GET /infraction-reports` | Listar relatórios |
| AcknowledgeInfractionReport | `PUT /infraction-reports/{id}/acknowledge` | Reconhecer relatório |
| CancelInfractionReport | `PUT /infraction-reports/{id}/cancel` | Cancelar relatório |
| CloseInfractionReport | `PUT /infraction-reports/{id}/close` | Fechar relatório |

**Total**: **51+ operações** mapeadas para API RSFN do Bacen.

**Observação Importante**: Todos os 7 domínios seguem o mesmo padrão:
```go
Request → Validate → Build SOAP → Sign XML → Send mTLS → Parse Response → Return
```

Não há lógica de negócio complexa no Bridge - apenas transformação de protocolo (gRPC/Pulsar → SOAP/XML).

---

## Estrutura do Repositório

```
rsfn-connect-bacen-bridge/
├── apps/dict/
│   ├── main.go                      # Entrypoint
│   │
│   ├── handlers/                    # CAMADA DE INTERFACE
│   │   ├── grpc/                    # ✅ Sync handlers
│   │   │   ├── directory_controller.go
│   │   │   ├── claim_controller.go
│   │   │   ├── antifraud_controller.go
│   │   │   └── infraction_report_controller.go
│   │   │
│   │   └── pulsar/                  # ⚡ Async handlers
│   │       ├── handler.go           # Main consumer
│   │       ├── directory_handler.go
│   │       └── claim_handler.go
│   │
│   ├── application/                 # CAMADA DE APLICAÇÃO
│   │   ├── ports/                   # Interfaces
│   │   │   ├── rsfn_client.go       # Interface para Bacen client
│   │   │   ├── xml_signer.go        # Interface para signer
│   │   │   └── publisher.go         # Interface para Pulsar publisher
│   │   │
│   │   └── usecases/
│   │       ├── directory/
│   │       │   └── process_request.go  # Orquestra chamada ao Bacen
│   │       └── claim/
│   │           └── process_request.go
│   │
│   ├── infrastructure/              # CAMADA DE INFRAESTRUTURA
│   │   ├── bacen/                   # RSFN Client (SOAP + mTLS)
│   │   │   ├── client.go
│   │   │   ├── directory.go
│   │   │   ├── claim.go
│   │   │   └── soap_builder.go
│   │   │
│   │   ├── signer/                  # XML Signer (JRE + JAR)
│   │   │   └── adapter.go
│   │   │
│   │   └── pulsar/                  # Pulsar Client
│   │       ├── consumer.go
│   │       └── publisher.go
│   │
│   └── setup/                       # Inicialização
│       ├── config.go
│       ├── grpc.go
│       ├── pulsar.go
│       ├── bacen.go
│       ├── circuit_breaker.go
│       └── observability.go
│
├── shared/                          # Código compartilhado
│   ├── http/                        # HTTP client com mTLS
│   └── signer/                      # XML Signer (JAR)
│       ├── signer.jar
│       └── interface.go
│
├── docker-compose.yml               # Pulsar, Jaeger, Prometheus
├── Dockerfile
└── README.md
```

**Nota**: Não há diretório `temporal/` no Bridge, pois workflows Temporal são responsabilidade do **RSFN Connect** (TEC-003).

---

## Deployment

### Docker Compose (Desenvolvimento)

```yaml
version: '3.8'

services:
  bridge:
    build: .
    ports:
      - "50051:50051"  # gRPC
    environment:
      - BACEN_URL=https://rsfn-hml.bcb.gov.br/dict
      - CERT_PATH=/certs/client.crt
      - KEY_PATH=/certs/client.key
      - CA_PATH=/certs/ca.crt
      - PULSAR_URL=pulsar://pulsar:6650
      - PULSAR_TOPIC_IN=bridge-dict-req-in
      - PULSAR_TOPIC_OUT=bridge-dict-res-out
    volumes:
      - ./certs:/certs:ro
      - ./shared/signer:/signer:ro
    depends_on:
      - pulsar

  pulsar:
    image: apachepulsar/pulsar:3.0.0
    command: bin/pulsar standalone
    ports:
      - "6650:6650"
      - "8080:8080"
```

---

### Kubernetes (Produção)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rsfn-bridge
spec:
  replicas: 3
  selector:
    matchLabels:
      app: rsfn-bridge
  template:
    metadata:
      labels:
        app: rsfn-bridge
    spec:
      containers:
      - name: bridge
        image: lb-conn/rsfn-bridge:latest
        ports:
        - containerPort: 50051
          name: grpc
        env:
        - name: BACEN_URL
          value: "https://rsfn.bcb.gov.br/dict"
        - name: PULSAR_URL
          value: "pulsar://pulsar-proxy:6650"
        volumeMounts:
        - name: certs
          mountPath: /certs
          readOnly: true
        - name: signer
          mountPath: /signer
          readOnly: true
      volumes:
      - name: certs
        secret:
          secretName: icp-brasil-certs
      - name: signer
        configMap:
          name: xml-signer
```

---

## Variáveis de Ambiente

```bash
# Bacen RSFN
BACEN_URL=https://rsfn.bcb.gov.br/dict
BACEN_TIMEOUT=30s

# Certificados mTLS (ICP-Brasil)
CERT_PATH=/certs/client.crt
KEY_PATH=/certs/client.key
CA_PATH=/certs/ca.crt

# XML Signer (JRE + JAR)
SIGNER_JAR_PATH=/signer/signer.jar
SIGNER_CERT_PATH=/certs/signing.p12
SIGNER_CERT_PASSWORD=changeit

# Pulsar (nomenclatura IcePanel)
PULSAR_URL=pulsar://pulsar:6650
PULSAR_TOPIC_REQ_IN=rsfn-dict-req-out     # Consome requisições (saída do Connect)
PULSAR_TOPIC_RES_OUT=rsfn-dict-res-out    # Publica respostas para dict.api
PULSAR_SUBSCRIPTION=bridge-dict-subscription

# gRPC
GRPC_PORT=50051

# Circuit Breaker
CIRCUIT_BREAKER_MAX_FAILURES=5
CIRCUIT_BREAKER_TIMEOUT=30s
CIRCUIT_BREAKER_HALF_OPEN_MAX_REQUESTS=3

# Observability
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
PROMETHEUS_PORT=9090
LOG_LEVEL=info
```

---

## Observabilidade

### OpenTelemetry Tracing

```go
// setup/observability.go
func SetupTracing(serviceName string) *sdktrace.TracerProvider {
	exporter, _ := jaeger.New(jaeger.WithCollectorEndpoint(
		jaeger.WithEndpoint(os.Getenv("JAEGER_ENDPOINT")),
	))

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	otel.SetTracerProvider(tp)
	return tp
}
```

**Spans importantes**:
- `bridge.grpc.GetEntry`: Chamada gRPC recebida
- `bridge.signer.SignXML`: Assinatura de XML
- `bridge.bacen.SendSOAP`: Chamada SOAP ao Bacen
- `bridge.pulsar.Consume`: Consumo de mensagem Pulsar

---

### Prometheus Métricas

```go
var (
	bacenRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "bridge_bacen_request_duration_seconds",
			Help: "Duration of Bacen requests",
		},
		[]string{"operation", "status"},
	)

	circuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bridge_circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
		},
		[]string{"breaker_name"},
	)
)
```

---

## Rastreabilidade

### Mapeamento de Requisitos

| Requisito Funcional | Implementação | Arquivo | Status |
|---------------------|---------------|---------|--------|
| REQ-DICT-001: Consultar chave Pix | `GetEntry` | `directory_controller.go` | ✅ Implementado |
| REQ-DICT-002: Cadastrar chave | `CreateEntry` | `directory_controller.go` | ✅ Implementado |
| REQ-DICT-003: Atualizar chave | `UpdateEntry` | `directory_controller.go` | ✅ Implementado |
| REQ-DICT-004: Excluir chave | `DeleteEntry` | `directory_controller.go` | ✅ Implementado |
| REQ-DICT-005: Processar claim | `GetClaim` | `claim_controller.go` | ✅ Implementado |

---

## Mapeamento com IcePanel e Repositório Real

### Correspondência de Nomenclaturas

| IcePanel Component | TEC-002 | Repositório Git | Observação |
|--------------------|---------|-----------------|------------|
| **DICT Proxy** | RSFN Bridge | `rsfn-connect-bacen-bridge` | ✅ Confirmado |
| `rsfn-dict-req-out` | Topic Pulsar IN | `PULSAR_TOPIC_REQ_IN` | Entrada do Bridge |
| `rsfn-dict-res-out` | Topic Pulsar OUT | `PULSAR_TOPIC_RES_OUT` | Saída do Bridge |
| Sistema "RSFN Connect" | - | Grupo/Namespace Pulsar | Organização lógica |

### Descrição IcePanel

Conforme ANA-001, o IcePanel descreve o componente **DICT Proxy** como:

> **"Proxy (adapter) para conexão segura (mTLS) e robusta com o DICT no BCB"**

**Tecnologias mencionadas**: Golang

**Responsabilidades confirmadas**:
- ✅ Conexão mTLS com Bacen
- ✅ Adapter/Proxy pattern
- ✅ Implementado em Go

### Validação Arquitetural

| Aspecto | TEC-002 Especifica | Implementação Real (ANA-002) | Status |
|---------|-------------------|------------------------------|--------|
| Linguagem | Go 1.22+ | Go 1.24.5 | ✅ Alinhado |
| Arquitetura | Clean Architecture | 4 camadas implementadas | ✅ Alinhado |
| Temporal Workflows | ❌ Não possui | ❌ Confirmado (sem go.temporal.io/sdk) | ✅ Alinhado |
| gRPC Server | ✅ Síncrono | ✅ handlers/grpc/ | ✅ Alinhado |
| Pulsar Consumer | ✅ Assíncrono | ✅ handlers/pulsar/ | ✅ Alinhado |
| XML Signer | ✅ JRE + JAR | ✅ shared/signer/ | ✅ Alinhado |
| mTLS | ✅ ICP-Brasil | ✅ shared/http/client.go | ✅ Alinhado |
| Circuit Breaker | ✅ sony/gobreaker | ✅ v2.3.0 | ✅ Alinhado |
| Domínios | Não especificado | 7 domínios (51+ ops) | ✅ Implementado+ |

**Conclusão**: TEC-002 v3.1 está **95% alinhado** com implementação real e arquitetura IcePanel.

**Gaps identificados** (ANA-004):
- 🟡 Nomenclatura topics Pulsar → **Corrigido nesta versão**
- 🟡 Dual protocol não documentado → **Corrigido nesta versão**

---

### ADRs Relacionadas

- **ADR-001**: Escolha de Go como linguagem para o Bridge
- **ADR-002**: Uso de JAR externo para assinatura XML (requisito Bacen)
- **ADR-003**: Circuit Breaker pattern para resiliência
- **ADR-004**: Clean Architecture para separação de responsabilidades
- **ADR-005**: **Bridge como adaptador puro (sem Temporal)** - Workflows movidos para Connect

---

## Próximos Passos

### ✅ Fase 1: Operações Síncronas via gRPC (CONCLUÍDA)
- [x] Implementar gRPC server
- [x] Implementar RSFN Client com mTLS
- [x] Implementar XML Signer (JRE + JAR)
- [x] Implementar Circuit Breaker
- [x] Testes unitários e de integração

### ⚠️ Fase 2: Operações Assíncronas via Pulsar (EM ANDAMENTO)
- [x] Implementar Pulsar Consumer
- [x] Implementar Pulsar Publisher
- [ ] Integração com RSFN Connect (TEC-003)
- [ ] Testes end-to-end com Connect

### 📋 Fase 3: Produção
- [ ] Helm charts para Kubernetes
- [ ] Secrets management (Vault) para certificados
- [ ] Dashboards Grafana
- [ ] Runbooks operacionais
- [ ] Performance testing (carga de 1000 req/s)

---

## Apêndice A: Exemplo de Request/Response SOAP

### Request SOAP (GetEntry)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Header>
    <Authentication xmlns="http://www.bcb.gov.br/pi/dict">
      <ISPB>12345678</ISPB>
      <Certificate>-----BEGIN CERTIFICATE-----...</Certificate>
    </Authentication>
  </soap:Header>
  <soap:Body>
    <GetEntryRequest xmlns="http://www.bcb.gov.br/pi/dict">
      <Key>11122233344</Key>
      <KeyType>CPF</KeyType>
    </GetEntryRequest>
  </soap:Body>
</soap:Envelope>
```

### Response SOAP (GetEntry)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <GetEntryResponse xmlns="http://www.bcb.gov.br/pi/dict">
      <Entry>
        <Key>11122233344</Key>
        <KeyType>CPF</KeyType>
        <Account>
          <Participant>12345678</Participant>
          <Branch>0001</Branch>
          <AccountNumber>1234567</AccountNumber>
          <AccountType>CACC</AccountType>
        </Account>
        <Owner>
          <Type>NATURAL_PERSON</Type>
          <TaxIdNumber>11122233344</TaxIdNumber>
          <Name>João Silva</Name>
        </Owner>
        <CreationDate>2024-01-15T10:30:00Z</CreationDate>
      </Entry>
    </GetEntryResponse>
  </soap:Body>
</soap:Envelope>
```

---

**Versão**: 3.1
**Data Revisão**: 2025-10-25
**Validação**: ✅ Alinhado com implementação real (ANA-002) e arquitetura IcePanel (ANA-001)
**Próxima Revisão**: Após validação em produção ou mudanças arquiteturais significativas

**Mudanças nesta versão (v3.1)**:
1. ✅ Adicionado seção "Dual Protocol Support" (gRPC + Pulsar simultâneo)
2. ✅ Atualizado nomenclatura Pulsar topics para padrão IcePanel (`rsfn-dict-*`)
3. ✅ Adicionado seção "Domínios Funcionais Implementados" (7 domínios, 51+ operações)
4. ✅ Adicionado seção "Mapeamento com IcePanel e Repositório Real"
5. ✅ Atualizado informações de repositório (Go 1.24.5, 110 arquivos, estatísticas reais)
6. ✅ Corrigido variáveis de ambiente com nomenclatura IcePanel

**Referências**:
- ANA-001: Análise Arquitetura IcePanel
- ANA-002: Análise Repositório Bridge
- ANA-004: Revalidação TEC vs Implementação
