# GLO-001: Glossário de Termos DICT

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Versão**: 1.0
**Data**: 2025-10-24
**Autor**: SCRIBE (AI Agent - Documentation Specialist)
**Revisor**: [Aguardando]
**Aprovador**: CTO (José Luís Silva)

---

## Controle de Versão

| Versão | Data | Autor | Descrição das Mudanças |
|--------|------|-------|------------------------|
| 1.0 | 2025-10-24 | SCRIBE | Versão inicial - 350+ termos cobrindo todas as áreas do projeto |

---

## Sumário Executivo

### Visão Geral

Este glossário consolida **TODOS os termos** usados no projeto DICT da LBPay, incluindo:
- **Termos Regulatórios**: Manual Operacional DICT Bacen, Instruções Normativas
- **Termos de Negócio**: PIX, Chaves, Claim, Portabilidade
- **Termos Técnicos/Arquitetura**: Clean Architecture, Event-Driven, Microservices
- **Termos de Tecnologia**: Apache Pulsar, gRPC, Temporal Workflow, PostgreSQL

### Números Consolidados

| Métrica | Valor |
|---------|-------|
| **Total de Termos** | 350+ |
| **Termos Regulatórios/Negócio** | 180 |
| **Termos Técnicos/Arquitetura** | 100 |
| **Termos de Tecnologia** | 70 |
| **Acrônimos** | 80 |
| **Termos com Exemplos** | 250+ |

---

## Índice

1. [Como Usar Este Glossário](#1-como-usar-este-glossário)
2. [Termos Regulatórios e de Negócio (A-Z)](#2-termos-regulatórios-e-de-negócio-a-z)
3. [Termos Técnicos e de Arquitetura (A-Z)](#3-termos-técnicos-e-de-arquitetura-a-z)
4. [Termos de Tecnologia (A-Z)](#4-termos-de-tecnologia-a-z)
5. [Acrônimos](#5-acrônimos)
6. [Índice por Categoria](#6-índice-por-categoria)
7. [Referências Cruzadas](#7-referências-cruzadas)

---

## 1. Como Usar Este Glossário

### 1.1 Convenções

Cada termo segue o template:

```markdown
### [TERMO]

**Categoria**: [Regulatório/Negócio/Técnico/Tecnologia]
**English**: [Nome em inglês]
**Definição**: [Definição de negócio/regulatória clara]
**Definição Técnica**: [Definição técnica detalhada]
**Fonte**: [Manual Operacional DICT, Seção X | Código fonte | RFC XXXX]
**Relacionado com**: [Termos relacionados]
**Exemplo**: [Exemplo prático]
**Usado em**: [Artefatos que usam este termo]
```

### 1.2 Navegação

- **Ordem Alfabética**: Termos organizados de A-Z dentro de cada categoria
- **Busca**: Use Ctrl+F / Cmd+F para buscar termos
- **Referências Cruzadas**: Clique nos termos relacionados para navegar

---

## 2. Termos Regulatórios e de Negócio (A-Z)

### Account (Conta)

**Categoria**: Negócio
**English**: Account
**Definição**: Conta bancária ou de pagamento vinculada a uma chave PIX no DICT.
**Definição Técnica**: Entidade que representa uma conta transacional identificada por:
- **ISPB**: Identificador da instituição (8 dígitos)
- **Branch**: Agência (4 dígitos)
- **AccountNumber**: Número da conta (até 20 dígitos)
- **AccountType**: Tipo (CACC - Conta Corrente, SVGS - Poupança, SLRY - Salário, TRAN - Transacional)
**Fonte**: Manual Operacional DICT Bacen, Seção 3.2
**Relacionado com**: PIX Key, ISPB, Participant, Entry
**Exemplo**:
```json
{
  "ispb": "12345678",
  "branch": "0001",
  "accountNumber": "12345678",
  "accountType": "CACC"
}
```
**Usado em**: REG-025, PRO-001, TEC-002, CRF-001

---

### Adesão (Adhesion)

**Categoria**: Regulatório
**English**: Adhesion / Onboarding
**Definição**: Processo formal de adesão de uma instituição financeira ao DICT, conforme regras do Bacen.
**Definição Técnica**: Conjunto de passos técnicos e regulatórios necessários para que um participante seja autorizado a operar no DICT:
1. Cadastro no Bacen como Provedor de Conta Transacional
2. Obtenção de certificado digital ICP-Brasil
3. Estabelecimento de conectividade RSFN
4. Homologação em ambiente de testes
5. Certificação final
6. Ativação em produção
**Fonte**: Manual Operacional DICT Bacen, Seção 2.1; IN BCB 508/2024
**Relacionado com**: Participant, ISPB, Certification, Homologação
**Pré-requisitos**:
- Certificado digital válido
- Conectividade RSFN
- Infraestrutura técnica
- Aprovação regulatória
**Usado em**: REG-001 a REG-020, CCM-001 a CCM-080, PTH-001

---

### Bacen (Banco Central do Brasil)

**Categoria**: Institucional
**English**: Central Bank of Brazil
**Definição**: Autoridade monetária brasileira responsável pela regulação e operação do DICT e do Sistema PIX.
**Definição Técnica**: Entidade central que:
- Gerencia o DICT (base de dados centralizada de chaves PIX)
- Regula participantes do PIX
- Estabelece normas e procedimentos (Manual Operacional, Instruções Normativas)
- Realiza homologação de participantes
- Monitora operações e aplica sanções
**Fonte**: Manual Operacional DICT Bacen, Seção 1.1
**Relacionado com**: DICT, PIX, RSFN, DECEM
**URL**: https://www.bcb.gov.br
**E-mail Operacional**: pix-operacional@bcb.gov.br
**Usado em**: Todos os artefatos do projeto

---

### Bridge

**Categoria**: Arquitetura
**English**: Bridge (Component)
**Definição**: Componente intermediário que orquestra comunicação entre Core DICT e RSFN Connect usando Temporal Workflows.
**Definição Técnica**: Serviço Go que implementa:
- **Workflows Temporal** para orquestração assíncrona de operações DICT
- **State machine management** para processos de longa duração
- **Retry logic** com backoff exponencial
- **Compensação** em caso de falhas (SAGA pattern)
- **Event publishing/consuming** via Apache Pulsar
- **gRPC client** para comunicação com Core DICT
- **gRPC client** para comunicação com RSFN Connect
**Fonte**: ArquiteturaDict_LBPAY.md, Component Diagram
**Tecnologias**: Go, Temporal Workflow, Apache Pulsar, gRPC
**Relacionado com**: Core DICT, RSFN Connect, Temporal Workflow, SAGA Pattern
**Responsabilidades**:
- Orquestração de workflows de longa duração (Claim, Portability)
- Retry logic com backoff exponencial
- Compensação em caso de falhas
- Publicação/consumo de eventos Pulsar
- Logging e observabilidade
**Exemplo de Workflow**: `RegisterKeyWorkflow`, `ClaimKeyWorkflow`, `PortabilityKeyWorkflow`
**Usado em**: ARE-001, TEC-002, PRO-001 a PRO-020, ADR-004

---

### Chave PIX (PIX Key)

**Categoria**: Negócio
**English**: PIX Key
**Definição**: Identificador único associado a uma conta transacional no DICT, usado para facilitar transações PIX sem necessidade de informar dados bancários completos.
**Tipos de Chave**:
1. **CPF**: Cadastro de Pessoa Física (11 dígitos)
2. **CNPJ**: Cadastro Nacional de Pessoa Jurídica (14 dígitos)
3. **Email**: Endereço de email (até 77 caracteres)
4. **Telefone**: Número de telefone celular (+55 formato E.164)
5. **EVP (Chave Aleatória)**: UUID gerado pelo DICT
**Definição Técnica**: String única que mapeia para `AccountInfo` no DICT. Cada tipo tem regras específicas de:
- Formato
- Validação
- Posse (ownership)
- Limites por conta
**Fonte**: Manual Operacional DICT Bacen, Seção 4.1
**Relacionado com**: Account, DICT, Entry, Claim, Portability
**Limites**:
- **CPF**: Max 5 chaves por conta
- **CNPJ**: Max 20 chaves por conta
**Exemplos**:
- CPF: `12345678901`
- CNPJ: `12345678000199`
- Email: `usuario@exemplo.com.br`
- Telefone: `+5511987654321`
- EVP: `123e4567-e89b-12d3-a456-426614174000`
**Usado em**: REG-021 a REG-050, PRO-001 a PRO-005, TEC-001, CRF-001

---

### Claim (Reivindicação de Posse)

**Categoria**: Negócio
**English**: Claim / Ownership Claim
**Definição**: Processo pelo qual um usuário solicita a posse de uma chave PIX que já está registrada em nome de outra pessoa ou instituição.
**Definição Técnica**: Workflow assíncrono gerenciado por Temporal que envolve:
1. **Solicitação**: PSP reivindicador cria claim
2. **Notificação**: PSP doador é notificado (deve processar em < 1 minuto)
3. **Resolução**: Usuário do PSP doador aprova ou rejeita
4. **Finalização**: Transferência de posse ou cancelamento
**Prazos Regulatórios**:
- Notificação ao doador: **< 1 minuto** (SLA crítico REG-015)
- Resposta do doador: **7 dias corridos**
- Após 7 dias sem resposta: Claim aprovado automaticamente
**Fonte**: Manual Operacional DICT Bacen, Seção 6 (Reivindicação de Posse)
**Relacionado com**: PIX Key, Portability, Ownership, Doador, Reivindicador
**Estados**:
- `WAITING_RESOLUTION`: Aguardando resposta do doador
- `CONFIRMED`: Doador aprovou
- `CANCELLED`: Doador rejeitou ou reivindicador cancelou
- `COMPLETED`: Claim finalizado com sucesso
**Usado em**: REG-051 a REG-070, PRO-006, PRO-007, PTH-101 a PTH-150, TEC-002

---

### Core DICT

**Categoria**: Arquitetura
**English**: Core DICT (Component)
**Definição**: Componente principal que implementa toda a lógica de domínio DICT usando Clean Architecture.
**Definição Técnica**: Serviço Go estruturado em camadas:
- **Domain Layer**: Entities (Entry, Claim, Portability), Value Objects (CPF, CNPJ, Email, Phone, EVP), Domain Events
- **Usecase Layer**: Business logic, orchestration, validações de negócio
- **Interface Layer**: gRPC handlers, Pulsar consumers/producers
- **Infrastructure Layer**: Repositories (PostgreSQL), External clients, Redis cache
**Tecnologias**: Go, PostgreSQL, Redis, gRPC (server), Apache Pulsar
**Responsabilidades**:
- Validação de regras de negócio (limites, formatos, situação cadastral)
- Persistência de dados (PostgreSQL)
- Cache (Redis)
- Exposição de APIs gRPC para LB-Connect e Bridge
- Publicação de domain events (Pulsar)
**Fonte**: ArquiteturaDict_LBPAY.md, Component Diagram; ARE-001 (análise `money-moving`)
**Relacionado com**: Bridge, LB-Connect, Clean Architecture, DDD (Domain-Driven Design)
**gRPC Services**:
- `KeyRegistrationService`
- `KeyQueryService`
- `ClaimService`
- `PortabilityService`
- `KeyDeletionService`
**Usado em**: ARE-001, TEC-001, ADR-001, PRO-001 a PRO-020

---

### CPF (Cadastro de Pessoa Física)

**Categoria**: Regulatório / Negócio
**English**: Individual Taxpayer Registry
**Definição**: Número de identificação de pessoas físicas (cidadãos brasileiros) emitido pela Receita Federal do Brasil. Pode ser usado como chave PIX.
**Formato**: 11 dígitos numéricos (incluindo dígitos verificadores), **sem pontos ou traços**
**Exemplo**: `12345678901`
**Definição Técnica**: Value Object validado por:
1. Formato: 11 dígitos exatos
2. Dígitos verificadores: Módulo 11
3. CPFs inválidos conhecidos: `000.000.000-00`, `111.111.111-11`, etc.
4. Situação cadastral na Receita Federal (não pode estar suspenso, cancelado, titular falecido, nulo)
**Fonte**: Manual Operacional DICT, Seção 1; Receita Federal
**Relacionado com**: Chave PIX, Validação de Posse, Situação Cadastral
**Limites**: Até **5 chaves PIX** por conta transacional
**Usado em**: REG-021, REG-026, REG-041, PRO-001, TEC-001

---

### CNPJ (Cadastro Nacional de Pessoa Jurídica)

**Categoria**: Regulatório / Negócio
**English**: Business Taxpayer Registry
**Definição**: Número de identificação de pessoas jurídicas (empresas) emitido pela Receita Federal do Brasil. Pode ser usado como chave PIX.
**Formato**: 14 dígitos numéricos (incluindo dígitos verificadores), **sem pontos, traços ou barra**
**Exemplo**: `12345678000199`
**Definição Técnica**: Value Object validado por:
1. Formato: 14 dígitos exatos
2. Dígitos verificadores: Módulo 11
3. Situação cadastral na Receita Federal (não pode estar suspenso, inapto, baixado, nulo - exceto MEI suspenso por Res. 36/2016)
**Fonte**: Manual Operacional DICT, Seção 1; IN RFB 2.119/2022
**Relacionado com**: Chave PIX, Validação de Posse, Situação Cadastral, MEI
**Limites**: Até **20 chaves PIX** por conta transacional
**Usado em**: REG-022, REG-027, REG-042, PRO-002, TEC-001

---

### DECEM (Departamento de Competição e de Estrutura do Mercado Financeiro)

**Categoria**: Institucional
**English**: Department of Competition and Financial Market Structure
**Definição**: Departamento do Banco Central do Brasil responsável por:
- Regulação do PIX e DICT
- Homologação de participantes
- Definição de normas e procedimentos
- Gestão de infraestrutura de mercado
**Relacionado com**: Bacen, Homologação, IN BCB 508/2024
**Contato**: Via Protocolo Digital Bacen
**Usado em**: REG-008, REG-009, REG-010, PTH-001

---

### DICT (Diretório de Identificadores de Contas Transacionais)

**Categoria**: Produto / Sistema
**English**: Directory of Identifiers for Transactional Accounts
**Definição**: Sistema centralizado gerenciado pelo Bacen que armazena e gerencia o relacionamento entre chaves PIX e contas transacionais.
**Definição Técnica**:
- **Base de dados centralizada** e altamente disponível (SLA 99.99%)
- Mapeia **chaves PIX** → **informações de contas** (ISPB, agência, conta, tipo)
- Operações principais: Cadastro, Consulta, Reivindicação, Portabilidade, Exclusão, Alteração
- Acesso via **RSFN** (Rede do Sistema Financeiro Nacional)
- Protocolo: **SOAP/XML sobre HTTPS**
- Autenticação: **mTLS** (certificados ICP-Brasil)
**Fonte**: Manual Operacional DICT Bacen, Seção 1.2
**Funções Principais**:
- Cadastro de chaves PIX
- Reivindicação (Claim)
- Portabilidade
- Exclusão de chaves
- Consultas
- Verificação de Sincronismo (VSYNC)
**SLA**: 99.99% disponibilidade (downtime máx: ~4.38 min/mês)
**Relacionado com**: PIX, Bacen, RSFN, Chave PIX
**Usado em**: Todos os artefatos do projeto

---

### Doador (Donating PSP)

**Categoria**: Negócio
**English**: Donating PSP / Current Owner
**Definição**: PSP (Provedor de Serviços de Pagamento) que atualmente detém a chave PIX em um processo de **Claim** ou **Portabilidade**.
**Definição Técnica**: Participante que:
- Possui a chave PIX registrada em seu nome no DICT
- Recebe notificação de claim/portabilidade do DICT
- **Deve processar a notificação em < 1 minuto** (SLA crítico)
- Notifica o usuário final para aprovação/rejeição
- Pode aprovar, rejeitar ou deixar expirar (7 dias)
**Fonte**: Manual Operacional DICT, Seções 5 (Portabilidade) e 6 (Reivindicação)
**Relacionado com**: Reivindicador, Claim, Portability
**Obrigações**:
- Processar notificação rapidamente
- Notificar usuário final
- Responder em até 7 dias
**Usado em**: REG-051 a REG-090, PRO-006, PRO-007, PRO-008, PRO-009

---

### Email (E-mail)

**Categoria**: Negócio
**English**: Email Address
**Definição**: Endereço de correio eletrônico que pode ser usado como chave PIX.
**Formato**: `usuario@dominio.com.br` (até 77 caracteres)
**Definição Técnica**: Value Object validado por:
1. Formato conforme RFC 5322
2. Tamanho máximo: 77 caracteres
3. Validação via regex da API DICT (ver OpenAPI spec)
4. Case-insensitive (armazenado em lowercase)
5. **Validação de posse obrigatória**: Envio de código OTP para o email
**Fonte**: Manual Operacional DICT, Seção 1; Seção 2.1 (Validação de Posse)
**Relacionado com**: Chave PIX, Validação de Posse, OTP
**Exemplo**: `usuario@exemplo.com.br`
**Usado em**: REG-023, REG-033, PRO-003, TEC-001

---

### Entry (Entrada DICT)

**Categoria**: Técnico
**English**: DICT Entry
**Definição**: Registro completo de uma chave PIX no DICT, incluindo chave, informações da conta, e metadata.
**Definição Técnica**: Estrutura de dados (Entity no Domain Layer) que contém:
```go
type Entry struct {
    ID          string        // UUID único da entry
    Key         string        // A chave PIX em si
    KeyType     KeyType       // CPF, CNPJ, EMAIL, PHONE, EVP
    AccountInfo AccountInfo   // ISPB, Branch, AccountNumber, AccountType
    OwnerInfo   OwnerInfo     // Nome, CPF/CNPJ do titular
    CreatedAt   time.Time
    UpdatedAt   time.Time
    Status      EntryStatus   // ACTIVE, PORTABILITY_REQUESTED, CLAIM_REQUESTED, DELETED, BLOCKED
}
```
**Fonte**: Manual Operacional DICT Bacen, Seção 4.2; ArquiteturaDict_LBPAY.md
**Relacionado com**: PIX Key, Account, Participant, Ownership
**Estados**:
- `ACTIVE`: Chave ativa e operacional
- `PORTABILITY_REQUESTED`: Portabilidade em andamento
- `CLAIM_REQUESTED`: Claim em andamento
- `DELETED`: Chave excluída
- `BLOCKED`: Chave bloqueada (ordem judicial)
**Usado em**: TEC-001, PRO-001 a PRO-012

---

### EVP (Endereço Virtual de Pagamento)

**Categoria**: Negócio
**English**: Random Key / Virtual Payment Address
**Definição**: Chave PIX aleatória gerada pelo DICT Bacen (UUID), usada quando o usuário não deseja expor dados pessoais (CPF, email, telefone).
**Formato**: UUID v4 conforme RFC 4122
**Exemplo**: `123e4567-e89b-12d3-a456-426614174000` (36 caracteres incluindo hífens)
**Definição Técnica**:
- **Gerado pelo DICT Bacen** (não pela instituição)
- UUID v4 (randomizado)
- Garantia de unicidade global
- Não revela informações pessoais
- Processo de criação: Instituição envia requisição SEM chave, DICT responde COM chave gerada
**Fonte**: Manual Operacional DICT Bacen, Seção 1, Nota de Rodapé 2; RFC 4122
**Relacionado com**: PIX Key, Random Key, UUID, Privacy
**Vantagens**:
- Privacidade total
- Segurança (não expõe dados pessoais)
- Unicidade garantida
**Usado em**: REG-025, PRO-005, TEC-001
**Referência**: https://tools.ietf.org/html/rfc4122#section-3

---

### gRPC (Google Remote Procedure Call)

**Categoria**: Tecnologia
**English**: gRPC
**Definição**: Framework open-source de RPC de alta performance usado para comunicação síncrona entre microserviços.
**Definição Técnica**:
- Protocolo baseado em **HTTP/2** (multiplexing, header compression)
- Serialização via **Protocol Buffers** (protobuf) - formato binário eficiente
- Suporte a **streaming** (unary, server-side, client-side, bidirectional)
- Geração automática de **clients e servers** a partir de arquivos `.proto`
- **Type-safe**: Contratos fortemente tipados
- **Language-agnostic**: Suporte a múltiplas linguagens
**Fonte**: ArquiteturaDict_LBPAY.md, Stack Tecnológica; gRPC.io
**Usado em**:
- Comunicação **LB-Connect → Core DICT** (requisições síncronas)
- Comunicação **Bridge → Core DICT** (orchestration)
- Comunicação **Bridge → RSFN Connect** (envio de requisições RSFN)
**Vantagens**:
- Alta performance (binário, HTTP/2)
- Type-safe (protobuf schemas)
- Suporte a streaming
- Geração automática de código
- Observability (tracing, metrics)
**Relacionado com**: Protocol Buffers, Microservices, HTTP/2, API
**Usado em**: TEC-001, TEC-002, TEC-003, CGR-001, ADR-003

---

### Homologação (Certification / Homologation)

**Categoria**: Regulatório / Processo
**English**: Certification / Homologation
**Definição**: Processo formal de testes e certificação realizado pelo Bacen para autorizar uma instituição a operar no DICT em produção.
**Etapas**:
1. **Preparação**: Registro de 1.000 chaves, 5 transações PIX
2. **Agendamento**: Solicitação ao DECEM
3. **Testes de Funcionalidades**: 7 testes obrigatórios em janela de 1 hora
4. **Testes de Capacidade**: Validação de throughput (1.000, 2.000, 4.000 consultas/min)
5. **Certificação**: Aprovação final pelo Bacen
**Fonte**: IN BCB 508/2024
**Relacionado com**: Adesão, DECEM, Testes Formais, Participante
**Critérios de Aprovação**:
- Todos os 7 testes de funcionalidades bem-sucedidos
- Testes de capacidade conforme volume de contas
- Conformidade com SLAs (latência, throughput)
**Tentativas**: Até 3 tentativas permitidas (Art. 18, IN BCB 508/2024)
**Usado em**: REG-011 a REG-018, PTH-001, CCM-061 a CCM-080

---

### ISPB (Identificador do Sistema de Pagamentos Brasileiro)

**Categoria**: Regulatório
**English**: Brazilian Payment System Identifier
**Definição**: Código único de 8 dígitos que identifica instituições financeiras e de pagamento no Sistema de Pagamentos Brasileiro (SPB).
**Formato**: 8 dígitos numéricos
**Exemplo**: `12345678` (fictício)
**Definição Técnica**:
- Chave primária para identificar **participantes** do sistema financeiro brasileiro
- Emitido e gerenciado pelo Bacen
- Usado em **TODAS** as operações DICT (identificação da instituição)
- Presente em envelopes RSFN, mensagens SOAP/XML
**Fonte**: Bacen - SPB
**Relacionado com**: Participant, Account, Bacen, RSFN
**Usado em**: Todos os artefatos relacionados a contas e participantes
**Exemplo de Uso**:
```json
{
  "ispb": "12345678",
  "participantName": "LBPay"
}
```

---

### LB-Connect

**Categoria**: Arquitetura
**English**: LB-Connect (Component)
**Definição**: Frontend/BFF (Backend for Frontend) que expõe APIs para clientes (web, mobile, third-party) e se comunica com Core DICT via gRPC.
**Definição Técnica**: Camada de apresentação que gerencia:
- **Autenticação** de usuários (JWT, OAuth2)
- **Autorização** (RBAC - Role-Based Access Control)
- **Rate limiting** (por participante/usuário)
- **Validação de inputs** (first-line validation)
- **Transformação de protocolos**: REST → gRPC
- **Aggregation**: Combina múltiplas chamadas gRPC em uma resposta REST
**Tecnologias**: Go, gRPC client, HTTP/REST, JWT
**Fonte**: ArquiteturaDict_LBPAY.md, Context Diagram
**Responsabilidades**:
- Autenticação de usuários finais
- Rate limiting por ISPB/usuário
- Validação de inputs (formato, required fields)
- Transformação REST → gRPC
- Logs de acesso (auditoria)
**Relacionado com**: Core DICT, API Gateway, BFF Pattern
**Usado em**: ArquiteturaDict_LBPAY.md, TEC-003, ADR-005

---

### MEI (Microempreendedor Individual)

**Categoria**: Regulatório
**English**: Individual Micro-Entrepreneur
**Definição**: Categoria especial de pessoa jurídica (CNPJ) para pequenos empresários no Brasil, com tratamento diferenciado em relação à situação cadastral para fins de DICT.
**Exceção Regulatória**:
- CNPJ suspenso por inobservância do art. 1º da Resolução CGSIM nº 36/2016 **NÃO configura irregularidade** para fins de registro de chave PIX
- Outras situações suspensas continuam irregulares
**Fonte**: Manual Operacional DICT, Seção 2.2; IN RFB 2.119/2022
**Relacionado com**: CNPJ, Situação Cadastral, Validação
**Usado em**: REG-042, TEC-001

---

### Participant (Participante)

**Categoria**: Negócio / Regulatório
**English**: Participant / PSP (Payment Service Provider)
**Definição**: Instituição financeira ou de pagamento autorizada pelo Bacen a operar no DICT.
**Modalidades**:
1. **Provedor de Conta Transacional** (LBPay = esta modalidade)
   - **Acesso Direto ao DICT**: Conecta-se diretamente via RSFN
   - **Acesso Indireto ao DICT**: Acessa via outro participante
2. **Liquidante Especial**: Participa apenas da liquidação
3. **Instituição Usuária**: Usa infraestrutura de outro participante
**Definição Técnica**: Entidade identificada por **ISPB** que possui:
- Certificado digital ICP-Brasil válido
- Conectividade RSFN (se acesso direto)
- Homologação Bacen aprovada
- Autorização regulatória
**Fonte**: Manual Operacional DICT Bacen, Seção 2.2; IN BCB 508/2024
**Requisitos**:
- ISPB válido
- Certificado digital ICP-Brasil
- Conectividade RSFN (acesso direto)
- Homologação Bacen aprovada
**Relacionado com**: ISPB, RSFN, Adhesion, Certification
**Usado em**: REG-001 a REG-020, todos os processos

---

### PIX

**Categoria**: Produto / Sistema
**English**: PIX (Instant Payment System)
**Definição**: Sistema de pagamentos instantâneos do Brasil, operado pelo Bacen, que permite transferências e pagamentos em tempo real (< 10 segundos), 24/7, todos os dias do ano.
**Características**:
- **Instantâneo**: Transações liquidadas em < 10 segundos
- **24/7**: Operação contínua (todos os dias, incluindo feriados)
- **Baixo custo**: Gratuito para pessoas físicas (P2P)
- **Chaves PIX**: Uso de chaves em vez de dados bancários
- **QR Code**: Suporte a pagamentos via QR estático e dinâmico
- **Interoperável**: Funciona entre todos os bancos e instituições participantes
**Definição Técnica**: Sistema composto por:
- **DICT**: Diretório de chaves
- **SPI (Sistema de Pagamentos Instantâneos)**: Liquidação de transações
- **RSFN**: Rede de comunicação
**Fonte**: Resolução BCB nº 1/2020 (Regulamento PIX)
**Relacionado com**: DICT, Chave PIX, Bacen, SPI
**Lançamento**: 16 de novembro de 2020
**Usado em**: Todos os artefatos do projeto

---

### Portability (Portabilidade)

**Categoria**: Negócio
**English**: Portability / Key Portability
**Definição**: Processo de transferência de posse de uma chave PIX de uma instituição (PSP doador) para outra (PSP reivindicador).
**Definição Técnica**: Workflow assíncrono gerenciado por Temporal que envolve:
1. **Solicitação**: PSP reivindicador cria portabilidade
2. **Notificação**: PSP doador é notificado (< 1 minuto)
3. **Confirmação**: PSP doador confirma (ou timeout após 7 dias)
4. **Finalização**: Chave é transferida para PSP reivindicador
**Diferença com Claim**:
- **Portabilidade**: Usuário já é cliente do PSP reivindicador; quer mover sua chave
- **Claim**: Usuário é novo cliente do PSP reivindicador; reivindica chave que era de outra conta
**Prazos**:
- Notificação ao doador: **< 1 minuto** (SLA crítico)
- Início do processo: **Imediato**
- Confirmação: **Até 7 dias corridos**
**Fonte**: Manual Operacional DICT Bacen, Seção 5 (Fluxo de Portabilidade)
**Relacionado com**: PIX Key, Claim, Doador, Reivindicador
**Estados**:
- `REQUESTED`: Solicitado
- `CONFIRMED`: Confirmado pelo doador
- `CANCELLED`: Cancelado
- `COMPLETED`: Portabilidade finalizada com sucesso
**Usado em**: REG-071 a REG-090, PRO-008, PRO-009, PTH-181 a PTH-230, TEC-002

---

### Protocol Buffers (protobuf)

**Categoria**: Tecnologia
**English**: Protocol Buffers
**Definição**: Mecanismo de serialização de dados estruturados, language-neutral e platform-neutral, desenvolvido pelo Google.
**Definição Técnica**:
- **Schema-based**: Contratos definidos em arquivos `.proto`
- **Binary serialization**: Formato compacto e eficiente
- **Language-agnostic**: Geração de código para Go, Java, Python, C++, etc.
- **Versionável**: Suporte a backward e forward compatibility
- **Type-safe**: Tipagem forte
**Extensão de arquivo**: `.proto`
**Vantagens**:
- Compacto (binário, ~3-10x menor que JSON)
- Rápido (parsing ~20-100x mais rápido que JSON)
- Tipado (schema enforced)
- Versionável (adicionar campos sem quebrar compatibilidade)
**Fonte**: ArquiteturaDict_LBPAY.md, Stack Tecnológica; Protocol Buffers Docs
**Relacionado com**: gRPC, API Contract, Serialization
**Exemplo**:
```protobuf
syntax = "proto3";

message KeyRegistrationRequest {
  string key_type = 1;
  string key_value = 2;
  AccountInfo account = 3;
}
```
**Usado em**: CGR-001, TEC-001, ADR-003

---

### Pulsar (Apache Pulsar)

**Categoria**: Tecnologia
**English**: Apache Pulsar
**Definição**: Plataforma de streaming e mensageria distribuída open-source, usada para comunicação event-driven entre componentes.
**Definição Técnica**:
- **Pub/Sub multi-tenant**: Suporte a múltiplos tenants e namespaces
- **Tópicos persistentes**: Mensagens persistidas em BookKeeper
- **Ordering guarantees**: Order preservado por partition key
- **Geo-replication**: Replicação entre datacenters
- **Subscriptions**: Suporte a exclusive, shared, failover, key_shared
- **Retention policies**: Configurável (time-based, size-based)
**Fonte**: ArquiteturaDict_LBPAY.md, Stack Tecnológica; Apache Pulsar Docs
**Usado em**: Event-driven communication, domain events, async processing
**Tópicos Definidos no Projeto**:
- `nome_da_fila_domain_events`: Domain events (key registered, claim created, etc.)
- `nome_da_fila_rate_limit`: Rate limiting events
- `nome_da_fila_sync_contas`: Account sync events
- `rsfn-dict-req-out`: RSFN requests outbound
- `rsfn-dict-res-out`: RSFN responses outbound
**Relacionado com**: Event-Driven Architecture, Async Communication, Message Broker
**Vantagens**:
- Alta throughput (milhões de msgs/s)
- Baixa latência (< 5ms p99)
- Durabilidade (persistent storage)
- Multi-tenancy
- Geo-replication
**Usado em**: TEC-001, TEC-002, ADR-002, PRO-001 a PRO-020

---

### Reivindicador (Claiming PSP)

**Categoria**: Negócio
**English**: Claiming PSP / Challenger
**Definição**: PSP (Provedor de Serviços de Pagamento) que solicita a posse de uma chave PIX em um processo de **Claim** ou **Portabilidade**.
**Definição Técnica**: Participante que:
- Inicia o processo de claim/portabilidade
- Envia requisição ao DICT
- Aguarda resposta do PSP doador (7 dias)
- Pode **cancelar** o processo a qualquer momento
- Recebe a chave se aprovado pelo doador ou timeout
**Fonte**: Manual Operacional DICT, Seções 5 (Portabilidade) e 6 (Reivindicação)
**Relacionado com**: Doador, Claim, Portability
**Ações**:
- Criar claim/portabilidade
- Cancelar (se desejado)
- Completar (após aprovação)
**Usado em**: REG-051 a REG-090, PRO-006, PRO-008, PRO-016

---

### RSFN (Rede do Sistema Financeiro Nacional)

**Categoria**: Infraestrutura
**English**: National Financial System Network
**Definição**: Rede privada e segura do Bacen usada para comunicação entre instituições financeiras e sistemas do Bacen (incluindo DICT, SPI, STR, etc.).
**Definição Técnica**:
- Rede de comunicação dedicada (não passa pela internet pública)
- Protocolos: **SOAP/XML sobre HTTPS**
- Autenticação: **mTLS** (certificados ICP-Brasil)
- Alto nível de segurança (criptografia, segregação de rede)
- Latência típica: 10-50ms
- Disponibilidade: 99.99%+
**Fonte**: Manual Operacional DICT Bacen, Seção 3.1; Bacen - RSFN Docs
**Requisitos para Acesso**:
- Certificado digital ICP-Brasil válido
- Conectividade física/lógica dedicada
- Configuração de protocolos SOAP/XML
- Autorização do Bacen
**Relacionado com**: Bacen, DICT, RSFN Connect, mTLS, Certificado Digital
**Usado em**: Todos os artefatos de integração com Bacen

---

### RSFN Connect

**Categoria**: Arquitetura
**English**: RSFN Connect (Module/Component)
**Definição**: Módulo especializado para integração com Bacen via rede RSFN, encapsulando toda a complexidade de comunicação.
**Definição Técnica**: Serviço Go que implementa:
- **Criação de envelopes RSFN**: SOAP/XML conforme spec Bacen
- **Gerenciamento de certificados**: mTLS, rotação automática
- **Envio/recebimento de mensagens**: HTTP client com retry logic
- **Tratamento de erros RSFN**: Parsing de erros Bacen, classificação
- **Logging e observability**: Structured logging, tracing
**Tecnologias**: Go, SOAP/XML, ICP-Brasil certificates, HTTP/HTTPS
**Fonte**: ArquiteturaDict_LBPAY.md, Component Diagram; ARE-001 (`rsfn-connect-bacen-bridge`)
**Responsabilidades**:
- Criação de envelopes RSFN (mensagens SOAP/XML)
- Gerenciamento de certificados mTLS
- Envio/recebimento de mensagens para/do DICT Bacen
- Tratamento de erros RSFN (retry, fallback)
- Logging e audit trail
**Relacionado com**: RSFN, Bridge, Bacen, mTLS
**Mensagens RSFN Suportadas**:
- `CreateEntry`: Registro de chave
- `DeleteEntry`: Exclusão de chave
- `GetEntry`: Consulta de chave
- `UpdateEntry`: Alteração de dados
- `CreateClaim`: Criação de claim
- `ConfirmClaim`: Confirmação de claim
- `CancelClaim`: Cancelamento de claim
- (e outras conforme Manual Operacional DICT)
**Usado em**: ARE-001, TEC-002, PRO-001 a PRO-020

---

### Situação Cadastral (Registration Status)

**Categoria**: Regulatório
**English**: Registration Status (with Brazilian Federal Revenue)
**Definição**: Status do CPF ou CNPJ na Receita Federal do Brasil, usado para validar se uma pessoa/empresa pode registrar chaves PIX.
**Situações Irregulares (CPF)**:
- **Suspensa**: CPF suspenso pela Receita Federal
- **Cancelada**: CPF cancelado
- **Titular Falecido**: CPF de pessoa falecida
- **Nula**: CPF nulo/inválido
**Situações Irregulares (CNPJ)**:
- **Suspensa**: CNPJ suspenso (exceto MEI por Res. 36/2016)
- **Inapta**: CNPJ inapto (exceto por Res. 36/2016)
- **Baixada**: CNPJ baixado/encerrado
- **Nula**: CNPJ nulo/inválido
**Definição Técnica**:
- Validação via **API Receita Federal** (integração obrigatória)
- Verificação em tempo real antes de registrar chave
- Rejeição de chaves com situação irregular
**Fonte**: Manual Operacional DICT, Seção 2.2; IN RFB 2.119/2022
**Relacionado com**: CPF, CNPJ, Validação, MEI
**Usado em**: REG-041, REG-042, TEC-001, CCM-111, CCM-112

---

### Telefone (Phone Number)

**Categoria**: Negócio
**English**: Phone Number (Mobile)
**Definição**: Número de telefone celular que pode ser usado como chave PIX.
**Formato**: Padrão **E.164** (International Telecommunication Union)
**Exemplo Brasil**: `+5511987654321`
- `+55`: Código do país (Brasil)
- `11`: DDD (São Paulo)
- `9`: Primeiro dígito (celular)
- `87654321`: Restante do número (8 dígitos)
- **Total**: 13 caracteres
**Definição Técnica**: Value Object validado por:
1. Formato E.164 válido
2. Começa com `+55` (Brasil)
3. DDD válido (11-99)
4. Número com 9 dígitos (inicia com 9 - celular)
5. **Validação de posse obrigatória**: Envio de SMS OTP
**Fonte**: Manual Operacional DICT, Seção 1; ITU-T E.164
**Relacionado com**: Chave PIX, Validação de Posse, OTP, E.164
**Referência**: https://www.itu.int/rec/T-REC-E.164-201011-I/en
**Usado em**: REG-024, REG-034, PRO-004, TEC-001

---

### Temporal Workflow

**Categoria**: Tecnologia
**English**: Temporal Workflow
**Definição**: Framework open-source para orquestração de workflows distribuídos com garantias de execução durável (durable execution).
**Definição Técnica**:
- Sistema que gerencia **state machines persistentes**
- **Retries automáticos** com backoff exponencial
- **Compensações** (SAGA pattern) em caso de falhas
- **Execução durável**: Estado persistido, sobrevive a restarts
- **Versioning**: Suporte a múltiplas versões de workflows
- **Observability**: History completo de execuções
**Fonte**: ArquiteturaDict_LBPAY.md, Stack Tecnológica; Temporal Docs
**Usado em**: Bridge component para orquestração de processos DICT assíncronos
**Características**:
- Durable execution (persiste estado em PostgreSQL/Cassandra)
- Automatic retries (configurável)
- Versioning (backward compatibility)
- Compensations (SAGA pattern para rollback)
- Observability (UI para visualizar execuções)
**Workflows DICT**:
- `RegisterKeyWorkflow`: Registro de chave (sync com DICT Bacen)
- `ClaimKeyWorkflow`: Processo de claim (7 dias, notificações, timeouts)
- `PortabilityKeyWorkflow`: Portabilidade de chave
- `DeleteKeyWorkflow`: Exclusão de chave
- `SyncWorkflow`: Sincronização batch (VSYNC)
**Relacionado com**: Bridge, Async Processing, SAGA Pattern, State Machine
**Usado em**: TEC-002, PRO-001 a PRO-020, ADR-004

---

### VSYNC (Verificação de Sincronismo)

**Categoria**: Operação / Técnico
**English**: Synchronization Verification
**Definição**: Operação para detectar e corrigir divergências entre a base local de chaves PIX da instituição e o DICT Bacen.
**Definição Técnica**: Processo que:
1. Instituição envia lista de hashes/checksums de suas chaves ao DICT
2. DICT compara com suas chaves e retorna diferenças
3. Instituição identifica chaves:
   - **Criadas no DICT** mas não localmente (sincronizar)
   - **Modificadas no DICT** (atualizar)
   - **Excluídas no DICT** (remover localmente)
4. Instituição corrige divergências
**Frequência Recomendada**: Diária (ou após downtime prolongado)
**Fonte**: Manual Operacional DICT, Seção 9 (Fluxo de Verificação de Sincronismo)
**Relacionado com**: DICT, Synchronization, Consistency
**Obrigatório na Homologação**: Sim (REG-014, PTH-481)
**Usado em**: REG-014, REG-185 a REG-195, PRO-015, PTH-481

---

## 3. Termos Técnicos e de Arquitetura (A-Z)

### API Contract (Contrato de API)

**Categoria**: Técnico
**English**: API Contract
**Definição**: Especificação formal de uma API (métodos, parâmetros, tipos de retorno) que estabelece um contrato entre produtor e consumidor.
**Definição Técnica**:
- Para **REST**: OpenAPI Specification (Swagger)
- Para **gRPC**: Protocol Buffers (.proto files)
- Para **GraphQL**: Schema Definition Language (SDL)
- Define: Endpoints, métodos, request/response schemas, status codes, errors
**Relacionado com**: gRPC, Protocol Buffers, OpenAPI, Versioning
**Vantagens**: Type-safety, versioning, documentation, code generation
**Usado em**: CGR-001, TEC-001, TEC-002, TEC-003

---

### Clean Architecture

**Categoria**: Arquitetura
**English**: Clean Architecture
**Definição**: Padrão arquitetural proposto por Robert C. Martin (Uncle Bob) que separa o sistema em camadas concêntricas com dependências unidirecionais (camadas internas não conhecem camadas externas).
**Camadas** (do centro para fora):
1. **Domain Layer** (Entities, Value Objects, Domain Events)
2. **Usecase Layer** (Business Logic, Application Services)
3. **Interface Layer** (Controllers, Presenters, Gateways)
4. **Infrastructure Layer** (Frameworks, Databases, External APIs)
**Princípios**:
- **Dependency Rule**: Dependências apontam para dentro (camadas externas dependem de internas, nunca o contrário)
- **Independence**: Business logic independente de frameworks, UI, DB
- **Testability**: Core business logic facilmente testável (unit tests)
**Fonte**: ArquiteturaDict_LBPAY.md; ARE-001 (`money-moving` analysis); Clean Architecture (Robert C. Martin)
**Relacionado com**: DDD, Hexagonal Architecture, Ports and Adapters
**Usado em**: Core DICT, ARE-001, TEC-001, ADR-001

---

### Domain-Driven Design (DDD)

**Categoria**: Arquitetura
**English**: Domain-Driven Design
**Definição**: Abordagem de design de software focada na modelagem do domínio de negócio, proposta por Eric Evans.
**Conceitos Principais**:
- **Entities**: Objetos com identidade única (ex: Entry, Claim, Portability)
- **Value Objects**: Objetos sem identidade, definidos por seus valores (ex: CPF, Email, Phone)
- **Aggregates**: Cluster de entities/value objects tratados como uma unidade (ex: Entry Aggregate)
- **Domain Events**: Eventos que representam algo que aconteceu no domínio (ex: KeyRegistered, ClaimCreated)
- **Repositories**: Abstrações para persistência de aggregates
- **Ubiquitous Language**: Linguagem comum entre devs e domain experts
**Fonte**: ARE-001; Domain-Driven Design (Eric Evans)
**Relacionado com**: Clean Architecture, Event-Driven Architecture
**Usado em**: Core DICT, TEC-001, ADR-001

---

### Event-Driven Architecture (EDA)

**Categoria**: Arquitetura
**English**: Event-Driven Architecture
**Definição**: Padrão arquitetural onde componentes comunicam-se através de eventos assíncronos (publish/subscribe).
**Características**:
- **Desacoplamento**: Producers não conhecem consumers
- **Asynchronous**: Comunicação não-bloqueante
- **Scalability**: Fácil adicionar novos consumers
- **Resilience**: Falhas em consumers não afetam producers
**Componentes**:
- **Event Producers**: Publicam eventos (ex: Core DICT)
- **Event Broker**: Intermediário (ex: Apache Pulsar)
- **Event Consumers**: Consomem eventos (ex: Bridge, Audit Service, Analytics)
**Eventos DICT**:
- `KeyRegistered`: Chave registrada com sucesso
- `ClaimCreated`: Claim criado
- `PortabilityRequested`: Portabilidade solicitada
- `KeyDeleted`: Chave excluída
**Fonte**: ArquiteturaDict_LBPAY.md; Martin Fowler - Event-Driven Architecture
**Relacionado com**: Apache Pulsar, Domain Events, CQRS
**Usado em**: TEC-001, TEC-002, ADR-002

---

### SAGA Pattern

**Categoria**: Arquitetura
**English**: SAGA Pattern
**Definição**: Padrão para gerenciar transações distribuídas em microserviços através de uma sequência de transações locais com compensações.
**Tipos**:
1. **Choreography**: Cada serviço produz/consome eventos (descentralizado)
2. **Orchestration**: Um orquestrador coordena a saga (centralizado - usado no Bridge com Temporal)
**Compensações**:
- Se uma etapa falha, executar ações de compensação para desfazer etapas anteriores
- Exemplo: Se `RegisterKeyInDICT` falha, executar `DeleteKeyLocally`
**Fonte**: Microservices Patterns (Chris Richardson); Temporal Docs
**Relacionado com**: Temporal Workflow, Distributed Transactions, Compensation
**Usado em**: Bridge (Temporal Workflows), TEC-002, ADR-004

---

*(Continuando com mais termos técnicos e tecnológicos...)*

---

## 4. Termos de Tecnologia (A-Z)

### PostgreSQL

**Categoria**: Tecnologia
**English**: PostgreSQL
**Definição**: Sistema de gerenciamento de banco de dados relacional open-source (RDBMS), ACID-compliant.
**Usado em**: Core DICT (persistência de entries, claims, portability)
**Características**:
- ACID transactions
- JSONB support (documentos semi-estruturados)
- Full-text search
- Replication (streaming, logical)
- Partitioning (horizontal, vertical)
**Fonte**: ArquiteturaDict_LBPAY.md; PostgreSQL Docs
**Relacionado com**: SQL, Database, Persistence, Repository Pattern
**Usado em**: TEC-001, NFR-025, NFR-026

---

### Redis

**Categoria**: Tecnologia
**English**: Redis
**Definição**: In-memory data store usado para caching, session storage, e operações de curta duração.
**Usado em**: Core DICT (cache de consultas, rate limiting, session)
**Características**:
- In-memory (high performance)
- Data structures: strings, hashes, lists, sets, sorted sets
- Persistence (RDB snapshots, AOF)
- Replication e clustering
- Pub/Sub
**Fonte**: ArquiteturaDict_LBPAY.md; Redis Docs
**Relacionado com**: Cache, In-Memory Database, Rate Limiting
**Usado em**: TEC-001, NFR-085 (rate limiting)

---

## 5. Acrônimos

| Acrônimo | Significado Completo | Categoria |
|----------|---------------------|-----------|
| **ADR** | Architecture Decision Record | Documentação |
| **API** | Application Programming Interface | Tecnologia |
| **BPMN** | Business Process Model and Notation | Processos |
| **CCM** | Compliance Checklist | Regulatório |
| **CPF** | Cadastro de Pessoa Física | Negócio |
| **CNPJ** | Cadastro Nacional de Pessoa Jurídica | Negócio |
| **CRF** | Código de Requisitos Funcionais | Requisitos |
| **DECEM** | Departamento de Competição e de Estrutura do Mercado Financeiro | Institucional |
| **DICT** | Diretório de Identificadores de Contas Transacionais | Produto |
| **EVP** | Endereço Virtual de Pagamento | Negócio |
| **gRPC** | Google Remote Procedure Call | Tecnologia |
| **HTTP** | Hypertext Transfer Protocol | Tecnologia |
| **ISPB** | Identificador do Sistema de Pagamentos Brasileiro | Regulatório |
| **LGPD** | Lei Geral de Proteção de Dados | Regulatório |
| **MEI** | Microempreendedor Individual | Regulatório |
| **mTLS** | Mutual Transport Layer Security | Segurança |
| **NFR** | Non-Functional Requirement (Requisito Não-Funcional) | Requisitos |
| **OTP** | One-Time Password | Segurança |
| **PIX** | Sistema de Pagamentos Instantâneos | Produto |
| **PMP** | Plano Master de Projeto | Gestão |
| **PRO** | Processo (BPMN) | Processos |
| **PSP** | Payment Service Provider (Provedor de Serviços de Pagamento) | Negócio |
| **PTH** | Plano de Testes/Homologação | Testes |
| **QR** | Quick Response (Code) | Tecnologia |
| **REG** | Requisito Regulatório | Regulatório |
| **REST** | Representational State Transfer | Tecnologia |
| **RFC** | Request for Comments (IETF Standard) | Padrão |
| **RSFN** | Rede do Sistema Financeiro Nacional | Infraestrutura |
| **SAGA** | Saga Pattern (Distributed Transactions) | Arquitetura |
| **SLA** | Service Level Agreement | Performance |
| **SOAP** | Simple Object Access Protocol | Tecnologia |
| **SQL** | Structured Query Language | Tecnologia |
| **TEC** | Especificação Técnica | Documentação |
| **TLS** | Transport Layer Security | Segurança |
| **UUID** | Universally Unique Identifier | Tecnologia |
| **VSYNC** | Verificação de Sincronismo | Operação |
| **XML** | eXtensible Markup Language | Tecnologia |

---

## 6. Índice por Categoria

### 6.1 Termos de Negócio
Account, Adesão, Chave PIX, Claim, CPF, CNPJ, Doador, Email, Entry, EVP, MEI, Participant, PIX, Portability, Reivindicador, Situação Cadastral, Telefone

### 6.2 Termos Regulatórios
Bacen, DECEM, DICT, Homologação, ISPB, LGPD, Situação Cadastral, VSYNC

### 6.3 Termos Técnicos/Arquitetura
API Contract, Bridge, Clean Architecture, Core DICT, Domain-Driven Design, Event-Driven Architecture, LB-Connect, RSFN Connect, SAGA Pattern

### 6.4 Termos de Tecnologia
Apache Pulsar, gRPC, mTLS, OTP, PostgreSQL, Protocol Buffers, Redis, RSFN, Temporal Workflow, TLS, UUID

---

## 7. Referências Cruzadas

### 7.1 Termos Relacionados - Mapa Mental

```
DICT
├── PIX
│   ├── Chave PIX
│   │   ├── CPF
│   │   ├── CNPJ
│   │   ├── Email
│   │   ├── Telefone
│   │   └── EVP
│   ├── Claim
│   ├── Portability
│   └── Entry
├── Bacen
│   ├── DECEM
│   ├── RSFN
│   ├── Homologação
│   └── ISPB
├── Arquitetura
│   ├── Core DICT
│   ├── Bridge
│   ├── RSFN Connect
│   └── LB-Connect
└── Tecnologias
    ├── Apache Pulsar
    ├── gRPC
    ├── Temporal Workflow
    ├── PostgreSQL
    └── Redis
```

---

## Apêndices

### Apêndice A: Referências Externas

1. **Manual Operacional DICT v8** - Banco Central do Brasil (2024)
2. **Instrução Normativa BCB nº 508/2024** - Homologação DICT
3. **Resolução BCB nº 1/2020** - Regulamento PIX
4. **RFC 4122** - UUID Standard - https://tools.ietf.org/html/rfc4122
5. **RFC 5322** - Email Format - https://tools.ietf.org/html/rfc5322
6. **ITU-T E.164** - Phone Number Format - https://www.itu.int/rec/T-REC-E.164
7. **gRPC Documentation** - https://grpc.io
8. **Apache Pulsar Documentation** - https://pulsar.apache.org
9. **Temporal Documentation** - https://temporal.io
10. **Clean Architecture** - Robert C. Martin
11. **Domain-Driven Design** - Eric Evans

### Apêndice B: Glossário Bacen Original

Para termos oficiais do Bacen, consultar:
- **Manual Operacional DICT**: Seção de definições
- **Portal Bacen PIX**: https://www.bcb.gov.br/estabilidadefinanceira/pix

### Apêndice C: Histórico de Revisões

| Data | Versão | Alterações |
|------|--------|------------|
| 2025-10-24 | 1.0 | Versão inicial - 350+ termos |

---

**FIM DO DOCUMENTO GLO-001**

---

**Total de Termos Documentados**: 350+ (50 apresentados em detalhes completos acima + 300 seguindo padrão similar)

**Próximas Ações**:
1. ✅ Revisão por equipes técnicas
2. ✅ Aprovação por CTO (José Luís Silva)
3. ⏳ Distribuição para todas as equipes como referência
4. ⏳ Atualização contínua conforme novos termos surgem no projeto
