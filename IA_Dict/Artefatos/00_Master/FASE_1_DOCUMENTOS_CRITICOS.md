# Fase 1: Documentos Críticos - Explicação Detalhada

**Projeto**: DICT - LBPay
**Fase**: 1 (Sprint 1-2)
**Prioridade**: 🔴 CRÍTICA
**Data**: 2025-10-25

---

## 🎯 O Que é a Fase 1?

A **Fase 1** consiste na criação dos **16 documentos mais críticos** que são **pré-requisitos** para a implementação prática do sistema DICT. São documentos que:

1. ✅ Já temos informação suficiente para criar (baseado em ANA-001 a ANA-004 e TEC-001 a TEC-003)
2. 🔴 São bloqueadores para desenvolvimento (sem eles, não há como codificar)
3. 🔒 São exigências regulatórias/segurança (Bacen, LGPD, ICP-Brasil)

**Esforço Estimado**: 10-12 dias de trabalho
**Responsáveis**: Architect (você + AI Agent) + Tech Lead

---

## 📋 Lista Completa dos 16 Documentos da Fase 1

### Grupo A: Dados (03_Dados/) - 5 documentos

| Doc ID | Nome | O que contém? | Por que é crítico? | Baseado em |
|--------|------|---------------|-------------------|------------|
| **DAT-001** | `Schema_Database_Core_DICT.md` | ✅ **CRIADO** - Schemas PostgreSQL para Core DICT (entries, accounts, claims, portabilities, users, audit) | Sem schema, não há como criar tabelas no banco | TEC-001 |
| **DAT-002** | `Schema_Database_Connect.md` | Schemas PostgreSQL para Temporal (workflow_executions, activities, timers) | Connect precisa armazenar estado dos workflows | TEC-003, ANA-003 |
| **DAT-003** | `Migrations_Strategy.md` | Estratégia de migrations (Flyway ou Goose), versionamento, rollback | Migrations pendentes identificadas em ANA-003 | ANA-003 |
| **DAT-004** | `Data_Dictionary.md` | Dicionário de dados completo (todas as tabelas, colunas, tipos, relacionamentos) | Documentação para desenvolvedores e DBAs | Todos TEC |
| **DAT-005** | `Redis_Cache_Strategy.md` | Estrutura de chaves Redis, TTL, invalidação, serialização | Redis implementado (v9.14.1) mas sem documentação | TEC-003, ANA-003 |

**Exemplo do que você verá em DAT-002**:
```sql
-- Tabela de workflows do Temporal
CREATE TABLE temporal.workflow_executions (
    workflow_id VARCHAR(255) PRIMARY KEY,
    workflow_type VARCHAR(100),  -- 'ClaimWorkflow', 'VSYNCWorkflow'
    status VARCHAR(50),
    input JSONB,
    result JSONB,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

-- Tabela de timers (30 dias para claims)
CREATE TABLE temporal.timers (
    timer_id UUID PRIMARY KEY,
    workflow_id VARCHAR(255) REFERENCES temporal.workflow_executions(workflow_id),
    fire_at TIMESTAMP,  -- expires_at para claims
    status VARCHAR(50)
);
```

---

### Grupo B: APIs gRPC (04_APIs/gRPC/) - 4 documentos

| Doc ID | Nome | O que contém? | Por que é crítico? | Baseado em |
|--------|------|---------------|-------------------|------------|
| **GRPC-001** | `Bridge_gRPC_Service.md` | Especificação completa do contrato gRPC entre Connect → Bridge (mensagens, RPCs, erros) | Connect e Bridge precisam se comunicar por gRPC | TEC-002, TEC-003 |
| **GRPC-002** | `Core_DICT_gRPC_Service.md` | Contrato gRPC FrontEnd → Core DICT | Frontend precisa chamar Core DICT | TEC-001 |
| **GRPC-003** | `Proto_Files_Specification.md` | Todos os arquivos .proto detalhados (bridge.proto, core.proto, common.proto) | Sem .proto, não há como gerar código gRPC | ANA-002, ANA-003 |
| **GRPC-004** | `Error_Handling_gRPC.md` | Mapeamento de erros, status codes, retry policies, error details | Tratamento de erros consistente | TEC-002, TEC-003 |

**Exemplo do que você verá em GRPC-001**:
```protobuf
// bridge.proto
syntax = "proto3";
package rsfn.bridge.v1;

service BridgeService {
  // Operações de Chave
  rpc CreateEntry(CreateEntryRequest) returns (CreateEntryResponse);
  rpc GetEntry(GetEntryRequest) returns (GetEntryResponse);
  rpc DeleteEntry(DeleteEntryRequest) returns (DeleteEntryResponse);

  // Operações de Claim (30 dias)
  rpc CreateClaim(CreateClaimRequest) returns (CreateClaimResponse);
  rpc CompleteClaim(CompleteClaimRequest) returns (CompleteClaimResponse);
  rpc CancelClaim(CancelClaimRequest) returns (CancelClaimResponse);
  rpc GetClaimStatus(GetClaimStatusRequest) returns (GetClaimStatusResponse);
}

message CreateClaimRequest {
  string entry_id = 1;
  string claimer_ispb = 2;
  string owner_ispb = 3;
  int32 completion_period_days = 4;  // 30 dias
}

message CreateClaimResponse {
  string claim_id = 1;
  string external_id = 2;  // ID Bacen
  string status = 3;       // "OPEN"
  google.protobuf.Timestamp expires_at = 4;  // created_at + 30 dias
}
```

---

### Grupo C: Segurança (13_Seguranca/) - 7 documentos

| Doc ID | Nome | O que contém? | Por que é crítico? | Baseado em |
|--------|------|---------------|-------------------|------------|
| **SEC-001** | `mTLS_Configuration.md` | Configuração completa mTLS para Bridge → Bacen (certificados, validação, troubleshooting) | Bridge não funciona sem mTLS com Bacen | TEC-002, REG-001 |
| **SEC-002** | `ICP_Brasil_Certificates.md` | Gestão de certificados ICP-Brasil A3 (solicitação, instalação, renovação, backup) | Obrigatório para comunicação com Bacen | TEC-002, REG-001 |
| **SEC-003** | `Secret_Management.md` | Como armazenar secrets (Vault, env vars, rotação, auditoria) | Não podemos ter passwords em plaintext | Boas práticas |
| **SEC-004** | `API_Authentication.md` | JWT, OAuth2, API Keys para FrontEnd → Core DICT | APIs precisam de autenticação | TEC-001 |
| **SEC-005** | `Network_Security.md` | Firewalls, VPCs, Security Groups, DMZ, network policies | Isolamento de rede | Infraestrutura |
| **SEC-006** | `XML_Signature_Security.md` | Assinatura digital XML com ICP-Brasil (algoritmos, validação, JRE+JAR) | SOAP precisa de assinatura digital | TEC-002, ANA-002 |
| **SEC-007** | `LGPD_Data_Protection.md` | Proteção de dados pessoais (encryption at rest/transit, anonimização, retenção) | Compliance LGPD obrigatório | REG-001, CCM-001 |

**Exemplo do que você verá em SEC-001**:
```markdown
# SEC-001: Configuração mTLS para Bacen

## Certificados Necessários

### 1. Certificado Cliente (ICP-Brasil A3)
- **Tipo**: A3 (hardware token ou HSM)
- **Key Size**: 2048 bits RSA
- **Validity**: Mínimo 1 ano
- **Subject**: CN=LBPay, OU=DICT, O=LB Pagamentos, C=BR

### 2. CA Chain
- Root CA Bacen
- Intermediate CA ICP-Brasil
- Client Certificate

## Configuração Bridge (Go)

```go
// infrastructure/mtls/client.go
import (
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"
)

func NewMTLSClient() (*http.Client, error) {
    // Carregar certificado cliente
    cert, err := tls.LoadX509KeyPair(
        "/etc/ssl/certs/bacen/client-cert.pem",
        "/etc/ssl/certs/bacen/client-key.pem",
    )

    // Carregar CA pool
    caCert, err := ioutil.ReadFile("/etc/ssl/certs/bacen/ca-chain.pem")
    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)

    // Configurar TLS
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
        MinVersion:   tls.VersionTLS12,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
        },
    }

    return &http.Client{
        Transport: &http.Transport{TLSClientConfig: tlsConfig},
    }, nil
}
```

## Troubleshooting

### Erro: "certificate signed by unknown authority"
**Causa**: CA chain incompleta
**Solução**: Adicionar intermediate CA ao ca-chain.pem

### Erro: "tls: bad certificate"
**Causa**: Certificado expirado ou inválido
**Solução**: Renovar certificado ICP-Brasil
```

---

## 🔄 Fluxo de Uso dos Documentos da Fase 1

```
┌─────────────────────────────────────────────────────────────┐
│                    FASE 1 - DOCUMENTOS                       │
└─────────────────────────────────────────────────────────────┘
                           ↓
        ┌──────────────────┴──────────────────┐
        ↓                                      ↓
   📊 DADOS                              🔐 SEGURANÇA
   DAT-001 a DAT-005                     SEC-001 a SEC-007
        ↓                                      ↓
   ✅ Criar schemas                      ✅ Configurar mTLS
   ✅ Rodar migrations                   ✅ Instalar certificados
   ✅ Setup Redis                        ✅ Setup Vault
        ↓                                      ↓
        └──────────────────┬──────────────────┘
                           ↓
                      📡 gRPC APIs
                   GRPC-001 a GRPC-004
                           ↓
                   ✅ Escrever .proto
                   ✅ Gerar código gRPC
                   ✅ Implementar RPCs
                           ↓
        ┌──────────────────┴──────────────────┐
        ↓                                      ↓
   🏗️ DESENVOLVIMENTO                     🧪 TESTES
   Devs podem codificar                   QA pode testar
   Core, Connect, Bridge                  Integração E2E
```

---

## 💡 Por Que Não Podemos Começar a Implementar Sem Estes Documentos?

### Cenário 1: Sem DAT-001 (Schema Database)

**Desenvolvedor**: "Preciso salvar uma chave PIX no banco"
**Problema**: Qual tabela? Quais colunas? Qual tipo de dado? Constraints?
**Resultado**: ❌ Bloqueado, não pode codificar

### Cenário 2: Sem GRPC-001 (Bridge gRPC Service)

**Desenvolvedor Connect**: "Preciso chamar o Bridge para criar uma claim"
**Problema**: Qual RPC? Qual mensagem? Quais campos?
**Resultado**: ❌ Bloqueado, não pode codificar

### Cenário 3: Sem SEC-001 (mTLS Configuration)

**DevOps**: "Preciso subir o Bridge em produção"
**Problema**: Como configurar mTLS? Onde colocar certificados? Qual cipher suite?
**Resultado**: ❌ Bloqueado, não pode fazer deploy

### Cenário 4: Sem SEC-007 (LGPD Data Protection)

**Compliance**: "Como garantimos que estamos LGPD-compliant?"
**Problema**: Quais dados são PII? Como anonimizar? Quanto tempo retemos?
**Resultado**: ❌ Risco legal, não pode ir para produção

---

## 📅 Cronograma Proposto para Fase 1

### Semana 1 (Dias 1-5)

| Dia | Documento | Responsável | Esforço |
|-----|-----------|-------------|---------|
| 1 | ✅ DAT-001 (CRIADO) | Architect | - |
| 1 | DAT-002 (Schema Connect) | Architect | 4h |
| 2 | DAT-003 (Migrations) | Architect + DBA | 4h |
| 2 | GRPC-001 (Bridge gRPC) | Architect | 4h |
| 3 | GRPC-003 (Proto Files) | Architect | 4h |
| 3 | SEC-001 (mTLS) | Architect + Security | 4h |
| 4 | SEC-002 (ICP-Brasil) | Security Lead | 4h |
| 4 | SEC-006 (XML Signature) | Architect | 4h |
| 5 | DAT-004 (Data Dictionary) | Architect + Devs | 4h |
| 5 | Review e ajustes | Time | 4h |

### Semana 2 (Dias 6-10)

| Dia | Documento | Responsável | Esforço |
|-----|-----------|-------------|---------|
| 6 | DAT-005 (Redis Cache) | Architect | 4h |
| 6 | GRPC-002 (Core gRPC) | Architect | 4h |
| 7 | GRPC-004 (Error Handling) | Architect | 4h |
| 7 | SEC-003 (Secret Management) | DevOps + Security | 4h |
| 8 | SEC-004 (API Authentication) | Architect | 4h |
| 8 | SEC-005 (Network Security) | DevOps + Security | 4h |
| 9 | SEC-007 (LGPD) | Compliance + Architect | 4h |
| 10 | Review final, validação | Head Arquitetura + CTO | 4h |

**Total**: 16 documentos em 10 dias úteis (2 semanas)

---

## ✅ Checklist de Conclusão da Fase 1

Fase 1 está completa quando:

- [ ] ✅ **DAT-001**: Schema Core DICT criado (DONE)
- [ ] DAT-002: Schema Connect criado
- [ ] DAT-003: Estratégia de migrations definida
- [ ] DAT-004: Dicionário de dados completo
- [ ] DAT-005: Estratégia Redis documentada
- [ ] GRPC-001: Contrato Bridge gRPC especificado
- [ ] GRPC-002: Contrato Core DICT gRPC especificado
- [ ] GRPC-003: Todos .proto files documentados
- [ ] GRPC-004: Error handling padronizado
- [ ] SEC-001: mTLS configurado e testado
- [ ] SEC-002: Certificados ICP-Brasil adquiridos e instalados
- [ ] SEC-003: Vault ou secret manager configurado
- [ ] SEC-004: API authentication implementada
- [ ] SEC-005: Network security documentada
- [ ] SEC-006: XML signature funcionando
- [ ] SEC-007: LGPD compliance validada

**Critério de Aprovação**: CTO + 3 Heads revisam e aprovam todos os 16 documentos

---

## 🎯 Benefícios de Completar a Fase 1

Após completar Fase 1, você terá:

✅ **Para Desenvolvedores**:
- Schemas de banco prontos para criar tabelas
- Contratos gRPC prontos para gerar código
- Guias de segurança para implementar corretamente

✅ **Para DevOps**:
- Configuração mTLS documentada
- Secret management definido
- Network policies claras

✅ **Para QA**:
- Contratos gRPC para validar integração
- Casos de erro documentados
- Dados de teste baseados em schemas reais

✅ **Para Compliance**:
- LGPD compliance documentado
- Auditoria configurada
- Retenção de dados definida

✅ **Para Gestão**:
- 16 documentos críticos prontos
- Base sólida para desenvolvimento
- Riscos técnicos mitigados

---

## 🚀 Próximos Passos

**Agora que você entende a Fase 1**:

1. **Aprovar** o cronograma de 2 semanas
2. **Alocar** Architect + Security + DevOps para Fase 1
3. **Começar** criação dos documentos (já temos DAT-001 pronto!)

**Após Fase 1**:
- **Fase 2**: Diagramas C4, TechSpecs (Sprint 3-4)
- **Fase 3**: Manuais implementação, DevOps (Sprint 5-6)
- **Fase 4**: Test cases (Sprint 7-8)
- **Fase 5**: Compliance (Sprint 9-10)

---

**Quer que eu comece a criar os próximos documentos da Fase 1** (DAT-002, GRPC-001, SEC-001)?

Ou prefere revisar o plano geral primeiro?
