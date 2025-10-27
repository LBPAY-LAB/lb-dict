# Backend Bridge Agent

**Role**: Backend Developer - RSFN Bridge
**Repo**: `conn-bridge/`
**Stack**: Go 1.24.5, Java 17 (XML Signer)

## 🎯 Responsabilidade

Implementar RSFN Bridge como adapter SOAP/REST com assinatura XML e mTLS.

## 📋 Tarefas

### gRPC Server
- Implementar BridgeService (GRPC-001)
- Receber chamadas do Connect
- Retornar respostas Bacen

### SOAP/REST Adapter
- Converter gRPC → SOAP envelope
- Gerar XML conforme spec Bacen
- Chamar XML Signer (Java)
- Fazer REST call com mTLS para Bacen

### XML Signer (Reutilizar Código Existente)
- **IMPORTANTE**: Copiar código de assinatura XML dos repos existentes (ver ANA-002)
- Java 17 + ICP-Brasil A3
- Endpoint /sign-xml
- Validação de certificado

### mTLS Configuration
- Certificados ICP-Brasil A3
- TLS 1.2+
- Cipher suites seguros (SEC-001)

## 🔗 Referências

- [TEC-002 v3.1](../../../../Artefatos/11_Especificacoes_Tecnicas/TEC-002_Bridge_Specification.md)
- [IMP-003](../../../../Artefatos/09_Implementacao/IMP-003_Manual_Implementacao_Bridge.md)
- [GRPC-001](../../../../Artefatos/04_APIs/gRPC/GRPC-001_Bridge_gRPC_Service.md)
- [SEC-001](../../../../Artefatos/13_Seguranca/SEC-001_mTLS_Configuration.md)
- [SEC-006](../../../../Artefatos/13_Seguranca/SEC-006_XML_Signature_Security.md)

## 💡 Reaproveitamento

Consultar repos existentes (Backlog DICT.csv) via MCP:
- Código de assinatura XML
- Configuração mTLS
- SDK Bacen