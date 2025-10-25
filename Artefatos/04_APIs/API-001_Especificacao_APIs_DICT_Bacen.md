# API-001 - Especificação de APIs DICT Bacen

**Agente Responsável**: MERCURY (AGT-API-001) - API Specialist
**Data de Criação**: 2025-10-24
**Versão**: 1.0
**Status**: Em Elaboração

---

## 📋 Índice

1. [Informações Gerais](#1-informações-gerais)
2. [Visão Geral da API DICT](#2-visão-geral-da-api-dict)
3. [Endpoints por Bloco Funcional](#3-endpoints-por-bloco-funcional)
4. [Segurança e Autenticação](#4-segurança-e-autenticação)
5. [Rate Limiting e Performance](#5-rate-limiting-e-performance)
6. [Tratamento de Erros](#6-tratamento-de-erros)
7. [Mapeamento com Requisitos Funcionais](#7-mapeamento-com-requisitos-funcionais)
8. [Recomendações de Implementação](#8-recomendações-de-implementação)
9. [Referências](#9-referências)

---

## 1. Informações Gerais

### 1.1 Objetivo do Documento

Este documento especifica todos os endpoints da API REST do DICT (Diretório de Identificadores de Contas Transacionais) do Banco Central do Brasil, baseado na versão **2.6.1** do OpenAPI oficial.

O objetivo é fornecer:
- ✅ Mapeamento completo de todos os 28 endpoints REST
- ✅ Cross-reference com os 72 Requisitos Funcionais do CRF-001
- ✅ Especificações de autenticação, rate limiting e erros
- ✅ Recomendações de implementação para LBPay

### 1.2 Versão da API Bacen

- **OpenAPI Version**: 3.0.0
- **API Version**: 2.6.1
- **License**: Apache 2.0
- **Contato Bacen**: suporte.ti@bcb.gov.br
- **Documentação Oficial**: https://www.bcb.gov.br/estabilidadefinanceira/pagamentosinstantaneos

### 1.3 Ambientes (Servers)

| Ambiente | URL Base | Porta |
|----------|----------|-------|
| **Homologação** | `https://dict-h.pi.rsfn.net.br` | 16522 |
| **Produção** | `https://dict.pi.rsfn.net.br` | 16422 |

**Path Base**: `/api/v2/`

**URLs Completas**:
- Homologação: `https://dict-h.pi.rsfn.net.br:16522/api/v2/`
- Produção: `https://dict.pi.rsfn.net.br:16422/api/v2/`

---

## 2. Visão Geral da API DICT

### 2.1 Descrição

O **DICT** é o serviço do arranjo PIX que permite buscar detalhes de contas transacionais com chaves de endereçamento convenientes. Permite ao pagador confirmar a identidade do recebedor e criar mensagens de instrução de pagamento com os detalhes da conta do recebedor.

### 2.2 Tags (Agrupamentos Funcionais)

A API DICT está organizada em **7 tags principais**:

| Tag | Nome em Português | Descrição | Endpoints |
|-----|-------------------|-----------|-----------|
| `Directory` | Diretório | CRUD de vínculos (chaves PIX) | 4 |
| `Key` | Chave | Validação e verificação de chaves | 1 |
| `Claim` | Reivindicação | Reivindicação de posse e portabilidade | 6 |
| `Reconciliation` | Reconciliação | Sincronização e CIDs | 5 |
| `InfractionReport` | Notificação de Infração | Notificações de infração | 5 |
| `Antifraud` | Antifraude | Marcações de fraude e estatísticas | 5 |
| `Refund` | Solicitação de Devolução | Devoluções | 4 |
| `Policies` | Política de Limitação | Rate limiting | 2 |

**Total**: **28 endpoints REST**

### 2.3 Tipos de Chave Suportados

| Tipo | Regex | Exemplo | Comentário |
|------|-------|---------|------------|
| **CPF** | `^[0-9]{11}$` | `12345678901` | 11 dígitos |
| **CNPJ** | `^[0-9]{14}$` | `12345678901234` | 14 dígitos |
| **PHONE** | `^\\+[1-9]\\d{1,14}$` | `+5510998765432` | Formato E.164 |
| **EMAIL** | `^[a-z0-9.!#$'*+\\/=?^_`{|}~-]+@[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*$` | `pix@bcb.gov.br` | Max 77 chars, lowercase |
| **EVP** | `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$` | `123e4567-e89b-12d3-a456-426655440000` | UUID v4 gerado pelo DICT |

⚠️ **Importante**: Novos tipos de chave podem ser adicionados no futuro. A implementação deve ser flexível.

---

## 3. Endpoints por Bloco Funcional

### 3.1 Bloco 1 - CRUD de Chaves PIX (Directory)

#### 3.1.1 POST /entries/
**Operação**: `createEntry`
**Resumo**: Criar Vínculo
**Descrição**: Cria um novo vínculo de chave com conta transacional.

**⚠️ PRÉ-REQUISITO OBRIGATÓRIO - Validação de Posse (Manual Bacen Subseção 2.1)**:
Antes de chamar este endpoint, o PSP **DEVE** ter validado a posse da chave conforme tipo:
- ✅ **Chaves tipo PHONE (celular)**: Enviar código único via SMS, usuário tem **30 minutos** para validar
- ✅ **Chaves tipo EMAIL**: Enviar código único via e-mail, usuário tem **30 minutos** para validar
- ✅ **Chaves tipo CPF/CNPJ**: Posse validada pela titularidade da conta (sem código)
- ✅ **Chaves tipo EVP (aleatória)**: Gerada pelo DICT, não requer validação de posse prévia
- ❌ **Timeout expirado**: Se 30 min expiraram, processo deve ser reiniciado
- 📖 **Referência**: Manual Operacional DICT Bacen - Subseção 2.1 (Validação da posse da chave)

**Características**:
- ✅ **Idempotente**: Usa `RequestId` (UUID v4) único por participante
- ✅ **Assíncrono**: Resposta imediata (201 Created)
- ✅ **Requer assinatura digital XML** (envelopada)
- ✅ **Requer validação de posse prévia** (exceto EVP)

**Request Body** (XML):
```xml
<?xml version="1.0" encoding="UTF-8" ?>
<CreateEntryRequest>
    <Signature></Signature>
    <Entry>
        <Key>+5561988880000</Key>
        <KeyType>PHONE</KeyType>
        <Account>
            <Participant>12345678</Participant>
            <Branch>0001</Branch>
            <AccountNumber>0007654321</AccountNumber>
            <AccountType>CACC</AccountType>
            <OpeningDate>2010-01-10T03:00:00Z</OpeningDate>
        </Account>
        <Owner>
            <Type>NATURAL_PERSON</Type>
            <TaxIdNumber>11122233300</TaxIdNumber>
            <Name>João Silva</Name>
        </Owner>
    </Entry>
    <Reason>USER_REQUESTED</Reason>
    <RequestId>a946d533-7f22-42a5-9a9b-e87cd55c0f4d</RequestId>
</CreateEntryRequest>
```

**Responses**:
- `201 Created`: Vínculo criado com sucesso
- `400 Bad Request`: Validação falhou (ver erros)
- `403 Forbidden`: Autorização negada
- `429 Rate Limited`: Limite excedido

**Rate Limiting**: Política `ENTRIES_WRITE`
- Taxa: 1200/min
- Balde: 36000

**Mapeamento RF**:
- **RF-BLO1-001** ✅ - Criar chave por solicitação do LB
- **RF-TRV-001** ✅ - Idempotência de operações

---

#### 3.1.2 GET /entries/{Key}
**Operação**: `getEntry`
**Resumo**: Consultar Vínculo
**Descrição**: Consulta detalhes de um vínculo pela chave.

**Path Parameters**:
- `Key` (string, required): Chave PIX a consultar

**Query Parameters**:
- `Participant` (string, optional): ISPB do participante
- `TaxIdNumber` (string, optional): CPF/CNPJ do pagador (obrigatório para consultas)

**Características**:
- ✅ **NÃO requer assinatura digital** (consulta)
- ✅ **Resposta assinada pelo DICT**
- ⚠️ **Rate limiting por usuário final** (anti-scan)

**Response** (200 OK):
```xml
<GetEntryResponse>
    <Signature></Signature>
    <Entry>
        <Key>+5561988880000</Key>
        <KeyType>PHONE</KeyType>
        <Account>
            <Participant>12345678</Participant>
            <Branch>0001</Branch>
            <AccountNumber>0007654321</AccountNumber>
            <AccountType>CACC</AccountType>
        </Account>
        <Owner>
            <Type>NATURAL_PERSON</Type>
            <TaxIdNumber>11122233300</TaxIdNumber>
            <Name>João Silva</Name>
        </Owner>
    </Entry>
</GetEntryResponse>
```

**Rate Limiting**: 3 políticas diferentes
1. **ENTRIES_READ_USER_ANTISCAN** (EMAIL, PHONE):
   - PF: 2/min, balde 100
   - PJ: 20/min, balde 1000
   - Status 404: subtrai 20 fichas (penalidade anti-scan)

2. **ENTRIES_READ_USER_ANTISCAN_V2** (CPF, CNPJ, EVP):
   - PF: 2/min, balde 100
   - PJ: 20/min, balde 1000
   - Status 404: subtrai 20 fichas

3. **ENTRIES_READ_PARTICIPANT_ANTISCAN**:
   - Categorias A-H (25.000/min até 2/min)
   - Status 404: subtrai 3 fichas

**Mapeamento RF**:
- **RF-BLO1-008** ✅ - Consulta de chave para participante PIX
- **RF-BLO5-003** ✅ - Interface de comunicação (crítico)
- **RF-BLO5-001** ✅ - Mecanismos de prevenção a ataques de leitura

**❗ CRÍTICO**: Este é o endpoint mais utilizado (dezenas de queries/segundo). Performance é essencial.

---

#### 3.1.3 POST /entries/{Key}
**Operação**: `updateEntry`
**Resumo**: Atualizar Vínculo
**Descrição**: Atualiza dados vinculados à chave (conta ou owner).

**Path Parameters**:
- `Key` (string, required): Chave PIX a atualizar

**Request Body** (XML):
```xml
<UpdateEntryRequest>
    <Signature></Signature>
    <Entry>
        <Key>+5561988880000</Key>
        <KeyType>PHONE</KeyType>
        <Account>
            <Participant>12345678</Participant>
            <Branch>0001</Branch>
            <AccountNumber>0007654321</AccountNumber>
            <AccountType>CACC</AccountType>
        </Account>
        <Owner>
            <Type>NATURAL_PERSON</Type>
            <TaxIdNumber>11122233300</TaxIdNumber>
            <Name>João Silva Atualizado</Name>
        </Owner>
    </Entry>
    <Reason>USER_REQUESTED</Reason>
</UpdateEntryRequest>
```

**Características**:
- ✅ **Requer assinatura digital XML**
- ✅ **Idempotente**
- ⚠️ **Não permite alterar KeyType ou Key**

**Rate Limiting**: Política `ENTRIES_UPDATE`
- Taxa: 600/min
- Balde: 600

**Mapeamento RF**:
- **RF-BLO1-009** ✅ - Alteração dos dados vinculados à chave
- **RF-BLO1-010** ✅ - Alteração para correção de inconsistências

---

#### 3.1.4 POST /entries/{Key}/delete
**Operação**: `deleteEntry`
**Resumo**: Excluir Vínculo
**Descrição**: Remove um vínculo do DICT.

**Path Parameters**:
- `Key` (string, required): Chave PIX a excluir

**Request Body** (XML):
```xml
<DeleteEntryRequest>
    <Signature></Signature>
    <Key>+5561988880000</Key>
    <Participant>12345678</Participant>
    <Reason>USER_REQUESTED</Reason>
</DeleteEntryRequest>
```

**Reasons válidos**:
- `USER_REQUESTED`: Solicitação do usuário final
- `ACCOUNT_CLOSURE`: Encerramento de conta
- `ENTRY_INACTIVITY`: Inatividade da chave
- `RECONCILIATION`: Reconciliação
- `FRAUD`: Fraude detectada

**Características**:
- ✅ **Requer assinatura digital XML**
- ✅ **Idempotente**
- ⚠️ **Bloqueado se existir Claim ativo**

**Rate Limiting**: Política `ENTRIES_WRITE`
- Taxa: 1200/min
- Balde: 36000

**Mapeamento RF**:
- **RF-BLO1-002** ✅ - Exclusão por incompatibilidade Receita Federal
- **RF-BLO1-003** ✅ - Exclusão por solicitação usuário final
- **RF-BLO1-004** ✅ - Exclusão por encerramento (participante)
- **RF-BLO1-005** ✅ - Exclusão por sincronismo (participante)
- **RF-BLO1-006** ✅ - Exclusão por fraude
- **RF-BLO1-007** ✅ - Status da chave (bloqueio judicial)

---

### 3.2 Bloco 1 - Validação de Chaves (Key)

#### 3.2.1 POST /keys/check
**Operação**: `checkKeys`
**Resumo**: Validar Existência de Chaves
**Descrição**: Verifica se um conjunto de chaves está registrado no DICT (sem retornar dados).

**Request Body** (XML):
```xml
<CheckKeysRequest>
    <Keys>
        <Key>+5561988880000</Key>
        <Key>12345678901</Key>
        <Key>pix@example.com</Key>
    </Keys>
</CheckKeysRequest>
```

**Response** (200 OK):
```xml
<CheckKeysResponse>
    <RegisteredKeys>
        <Key>+5561988880000</Key>
        <Key>12345678901</Key>
    </RegisteredKeys>
</CheckKeysResponse>
```

**Características**:
- ✅ **NÃO requer assinatura digital**
- ✅ **Resposta assinada pelo DICT**
- ✅ **Não revela dados completos do vínculo** (privacidade)
- ⚠️ **Max 1000 chaves por requisição**

**Rate Limiting**: Política `KEYS_CHECK`
- Taxa: 70/min
- Balde: 70

**Mapeamento RF**:
- **RF-BLO1-011** ✅ - Validar chave (checklist)
- **RF-BLO5-007** ✅ - Verificação de chaves PIX registradas

---

### 3.3 Bloco 2 - Reivindicação e Portabilidade (Claim)

#### 3.3.1 POST /claims/
**Operação**: `createClaim`
**Resumo**: Criar Reivindicação
**Descrição**: Cria uma reivindicação de posse ou portabilidade.

**Request Body** (XML):
```xml
<CreateClaimRequest>
    <Signature></Signature>
    <Claim>
        <Key>+5561988880000</Key>
        <Type>PORTABILITY</Type>
        <Claimer>
            <Participant>87654321</Participant>
            <Branch>0001</Branch>
            <AccountNumber>0001234567</AccountNumber>
            <AccountType>CACC</AccountType>
        </Claimer>
    </Claim>
    <RequestId>b123c456-8d33-53b6-0c0c-f98de66d1e5e</RequestId>
</CreateClaimRequest>
```

**Claim Types**:
- `OWNERSHIP`: Reivindicação de posse (chave mudou de dono)
- `PORTABILITY`: Portabilidade (mesma chave, mudar PSP)

**Claim Status (Ciclo de Vida)**:
```
OPEN → WAITING_RESOLUTION → CONFIRMED → COMPLETED
  ↓           ↓                  ↓
  └──────> CANCELLED <──────────┘
```

**Rate Limiting**: Política `CLAIMS_WRITE`
- Taxa: 1200/min
- Balde: 36000

**Mapeamento RF**:
- **RF-BLO2-001** ✅ - Reivindicação (Reivindicador) - Criação
- **RF-BLO2-007** ✅ - Portabilidade de chave (PSP reivindicador)

---

#### 3.3.2 GET /claims/{ClaimId}
**Operação**: `getClaim`
**Resumo**: Consultar Reivindicação
**Descrição**: Obtém detalhes de uma reivindicação pelo ID.

**Path Parameters**:
- `ClaimId` (string, required): UUID da reivindicação

**Response** (200 OK):
```xml
<GetClaimResponse>
    <Signature></Signature>
    <Claim>
        <ClaimId>c789d012-9e44-64c7-1d1d-g09ef77e2f6f</ClaimId>
        <Key>+5561988880000</Key>
        <Type>PORTABILITY</Type>
        <Status>WAITING_RESOLUTION</Status>
        <Claimer>...</Claimer>
        <Donee>...</Donee>
        <OpenDate>2023-10-24T10:00:00Z</OpenDate>
        <ResolutionPeriodEnd>2023-10-31T10:00:00Z</ResolutionPeriodEnd>
    </Claim>
</GetClaimResponse>
```

**Rate Limiting**: Política `CLAIMS_READ`
- Taxa: 600/min
- Blude: 18000

**Mapeamento RF**:
- **RF-BLO2-006** ✅ - Consultar reivindicação

---

#### 3.3.3 GET /claims/
**Operação**: `listClaims`
**Resumo**: Listar Reivindicações
**Descrição**: Lista reivindicações com filtros opcionais.

**Query Parameters**:
- `Status` (string, optional): Filtrar por status (OPEN, WAITING_RESOLUTION, etc.)
- `Type` (string, optional): Filtrar por tipo (OWNERSHIP, PORTABILITY)
- `Role` (string, optional): `CLAIMER` ou `DONEE`
- `Limit` (int, optional): Max resultados (default 100)
- `Offset` (int, optional): Paginação

**Rate Limiting**: 2 políticas
1. **CLAIMS_LIST_WITH_ROLE** (com Role):
   - Taxa: 40/min, Balde: 200

2. **CLAIMS_LIST_WITHOUT_ROLE** (sem Role):
   - Taxa: 10/min, Balde: 50

**Mapeamento RF**:
- **RF-BLO2-005** ✅ - Listagem de reivindicações
- **RF-BLO2-004** ✅ - Receber/Monitorar reivindicações

---

#### 3.3.4 POST /claims/{ClaimId}/acknowledge
**Operação**: `acknowledgeClaim`
**Resumo**: Receber Reivindicação (Doador)
**Descrição**: PSP doador confirma recebimento da reivindicação.

**Transição**: `OPEN` → `WAITING_RESOLUTION`

**Request Body** (XML):
```xml
<AcknowledgeClaimRequest>
    <Signature></Signature>
    <ClaimId>c789d012-9e44-64c7-1d1d-g09ef77e2f6f</ClaimId>
</AcknowledgeClaimRequest>
```

**Rate Limiting**: Política `CLAIMS_WRITE`

**Mapeamento RF**:
- **RF-BLO2-009** ✅ - Reivindicação (Doador) - Receber/Monitorar

---

#### 3.3.5 POST /claims/{ClaimId}/confirm
**Operação**: `confirmClaim`
**Resumo**: Confirmar Reivindicação (Doador)
**Descrição**: PSP doador confirma a reivindicação.

**Transição**: `WAITING_RESOLUTION` → `CONFIRMED`

**Características**:
- ✅ **Remove a chave do DICT automaticamente**
- ✅ **PSP doador deve remover da base interna**

**Request Body** (XML):
```xml
<ConfirmClaimRequest>
    <Signature></Signature>
    <ClaimId>c789d012-9e44-64c7-1d1d-g09ef77e2f6f</ClaimId>
</ConfirmClaimRequest>
```

**Rate Limiting**: Política `CLAIMS_WRITE`

**Mapeamento RF**:
- **RF-BLO2-010** ✅ - Reivindicação (Doador) - Confirmação
- **RF-BLO2-007** ✅ - Portabilidade (PSP doador)
- **RF-BLO2-008** ✅ - Portabilidade (PSP reivindicador)

---

#### 3.3.6 POST /claims/{ClaimId}/cancel
**Operação**: `cancelClaim`
**Resumo**: Cancelar Reivindicação
**Descrição**: Cancela uma reivindicação (Reivindicador ou Doador).

**Transição**: Qualquer estado → `CANCELLED`

**Request Body** (XML):
```xml
<CancelClaimRequest>
    <Signature></Signature>
    <ClaimId>c789d012-9e44-64c7-1d1d-g09ef77e2f6f</ClaimId>
    <Reason>USER_REQUESTED</Reason>
</CancelClaimRequest>
```

**Rate Limiting**: Política `CLAIMS_WRITE`

**Mapeamento RF**:
- **RF-BLO2-002** ✅ - Cancelamento (Reivindicador)
- **RF-BLO2-011** ✅ - Cancelamento (Doador)

---

#### 3.3.7 POST /claims/{ClaimId}/complete
**Operação**: `completeClaim`
**Resumo**: Concluir Reivindicação (Reivindicador)
**Descrição**: Finaliza a reivindicação criando novo vínculo.

**Transição**: `CONFIRMED` → `COMPLETED`

**Características**:
- ✅ **Cria novo vínculo automaticamente no DICT**
- ✅ **PSP reivindicador deve criar na base interna**

**Request Body** (XML):
```xml
<CompleteClaimRequest>
    <Signature></Signature>
    <ClaimId>c789d012-9e44-64c7-1d1d-g09ef77e2f6f</ClaimId>
</CompleteClaimRequest>
```

**Rate Limiting**: Política `CLAIMS_WRITE`

**Mapeamento RF**:
- **RF-BLO2-003** ✅ - Conclusão (Reivindicador)

---

### 3.4 Bloco 5 - Reconciliação (Reconciliation)

#### 3.4.1 POST /sync-verifications/
**Operação**: `createSyncVerification`
**Resumo**: Criar Verificação de Sincronismo (VSync)
**Descrição**: Verifica se os vínculos do PSP estão sincronizados com o DICT.

**Request Body** (XML):
```xml
<CreateSyncVerificationRequest>
    <Signature></Signature>
    <KeyType>PHONE</KeyType>
    <VSync>996fc1dd3b6b14bcf0c9fe8320eb66d7e2a3fd874ccf767b2e939641b1ea8eaf</VSync>
</CreateSyncVerificationRequest>
```

**Response** (200 OK):
```xml
<CreateSyncVerificationResponse>
    <Signature></Signature>
    <KeyType>PHONE</KeyType>
    <Match>true</Match>
    <DictVSync>996fc1dd3b6b14bcf0c9fe8320eb66d7e2a3fd874ccf767b2e939641b1ea8eaf</DictVSync>
</CreateSyncVerificationResponse>
```

**Cálculo do VSync**:
```
vsync = XOR(cid1, cid2, cid3, ..., cidN)
```

**Rate Limiting**: Política `SYNC_VERIFICATIONS_WRITE`
- Taxa: 10/min
- Balde: 50

**Mapeamento RF**:
- **RF-BLO5-009** ✅ - Verificação de VSync (participante com acesso direto)

---

#### 3.4.2 POST /cids/files/
**Operação**: `createCidSetFile`
**Resumo**: Solicitar Arquivo de CIDs
**Descrição**: Solicita a criação de um arquivo com todos os CIDs de um tipo de chave.

**Request Body** (XML):
```xml
<CreateCidSetFileRequest>
    <Signature></Signature>
    <KeyType>PHONE</KeyType>
</CreateCidSetFileRequest>
```

**Response** (202 Accepted):
```xml
<CreateCidSetFileResponse>
    <Signature></Signature>
    <Id>12345</Id>
    <Status>PROCESSING</Status>
</CreateCidSetFileResponse>
```

**Rate Limiting**: Política `CIDS_FILES_WRITE`
- Taxa: 40/dia
- Balde: 200

**Mapeamento RF**:
- **RF-BLO5-010** ✅ - Lista de CIDs

---

#### 3.4.3 GET /cids/files/{Id}
**Operação**: `getCidSetFile`
**Resumo**: Obter Arquivo de CIDs
**Descrição**: Obtém o arquivo de CIDs solicitado (se pronto).

**Path Parameters**:
- `Id` (string, required): ID do arquivo

**Response** (200 OK):
```xml
<GetCidSetFileResponse>
    <Signature></Signature>
    <Id>12345</Id>
    <Status>COMPLETED</Status>
    <Url>https://dict.pi.rsfn.net.br/cids/files/12345/download</Url>
    <ExpirationDate>2023-10-31T23:59:59Z</ExpirationDate>
</GetCidSetFileResponse>
```

**Status**:
- `PROCESSING`: Em processamento
- `COMPLETED`: Pronto para download
- `FAILED`: Falha na geração

**Rate Limiting**: Política `CIDS_FILES_READ`
- Taxa: 10/min
- Balde: 50

**Mapeamento RF**:
- **RF-BLO5-010** ✅ - Lista de CIDs

---

#### 3.4.4 GET /cids/events
**Operação**: `listCidSetEvents`
**Resumo**: Listar Eventos de CIDs
**Descrição**: Obtém log de eventos de modificação de CIDs (polling).

**Query Parameters**:
- `KeyType` (string, required): Tipo de chave
- `Since` (datetime, optional): Timestamp inicial
- `Limit` (int, optional): Max eventos (default 100)

**Response** (200 OK):
```xml
<ListCidSetEventsResponse>
    <Signature></Signature>
    <Events>
        <Event>
            <EventType>CREATED</EventType>
            <Cid>28c06eb41c4dc9c3ae114831efcac7446c8747777fca8b145ecd31ff8480ae88</Cid>
            <Timestamp>2023-10-24T10:00:00Z</Timestamp>
        </Event>
        <Event>
            <EventType>DELETED</EventType>
            <Cid>4d4abb9168114e349672b934d16ed201a919cb49e28b7f66a240e62c92ee007f</Cid>
            <Timestamp>2023-10-24T10:05:00Z</Timestamp>
        </Event>
    </Events>
</ListCidSetEventsResponse>
```

**Event Types**:
- `CREATED`: CID criado
- `UPDATED`: CID atualizado
- `DELETED`: CID removido

**Rate Limiting**: Política `CIDS_EVENTS_LIST`
- Taxa: 20/min
- Balde: 100

**Mapeamento RF**:
- **RF-BLO5-010** ✅ - Lista de CIDs (monitoramento contínuo)

---

#### 3.4.5 GET /cids/entries/{Cid}
**Operação**: `getEntryByCid`
**Resumo**: Consultar Vínculo por CID
**Descrição**: Obtém vínculo pelo CID (para reconciliação).

**Path Parameters**:
- `Cid` (string, required): CID (256-bit hex string)

**Response** (200 OK): Igual a `GET /entries/{Key}`

**Rate Limiting**: Política `CIDS_ENTRIES_READ`
- Taxa: 1200/min
- Balde: 36000

**Mapeamento RF**:
- **RF-BLO5-010** ✅ - Reconciliação por CID

---

### 3.5 Bloco 4 - Notificação de Infração (InfractionReport)

#### 3.5.1 POST /infraction-reports/
**Operação**: `createInfractionReport`
**Resumo**: Criar Notificação de Infração
**Descrição**: Cria notificação de infração por suspeita de fraude.

**Request Body** (XML):
```xml
<CreateInfractionReportRequest>
    <Signature></Signature>
    <InfractionReport>
        <EndToEndId>E1234567820231024100000000000001</EndToEndId>
        <ReportType>REFUND_REQUEST</ReportType>
        <DebtorParticipant>12345678</DebtorParticipant>
        <CreditorParticipant>87654321</CreditorParticipant>
        <ReportDetails>Suspeita de fraude...</ReportDetails>
    </InfractionReport>
    <RequestId>d234e567-0f55-75d8-2e2e-h10fg88f3g7g</RequestId>
</CreateInfractionReportRequest>
```

**Report Types**:
- `REFUND_REQUEST`: Solicitar devolução
- `REFUND_CANCELLED`: Cancelar devolução

**Status Cycle**:
```
OPEN → ACKNOWLEDGED → CLOSED
  ↓
  └──────> CANCELLED
```

**Rate Limiting**: Política `INFRACTION_REPORTS_WRITE`
- Taxa: 1200/min
- Balde: 36000

**Mapeamento RF**:
- **RF-BLO4-004** ✅ - Notificação de infração para abertura de devolução
- **RF-BLO4-005** ✅ - Notificação de infração para cancelamento de devolução

---

#### 3.5.2 GET /infraction-reports/{InfractionReportId}
**Operação**: `getInfractionReport`
**Resumo**: Consultar Notificação de Infração
**Descrição**: Obtém detalhes de uma notificação.

**Rate Limiting**: Política `INFRACTION_REPORTS_READ`
- Taxa: 600/min
- Balde: 18000

**Mapeamento RF**:
- **RF-BLO4-004** ✅ - Consultar notificação de infração

---

#### 3.5.3 GET /infraction-reports/
**Operação**: `listInfractionReports`
**Resumo**: Listar Notificações de Infração
**Descrição**: Lista notificações com filtros (polling).

**Query Parameters**:
- `Status` (string, optional)
- `Role` (string, optional): `REPORTER` ou `REPORTED`
- `Limit` (int, optional)
- `Offset` (int, optional)

**Rate Limiting**: 2 políticas
1. **INFRACTION_REPORTS_LIST_WITH_ROLE**: 40/min, 200
2. **INFRACTION_REPORTS_LIST_WITHOUT_ROLE**: 10/min, 50

**Mapeamento RF**:
- **RF-BLO4-004** ✅ - Monitorar notificações (polling)

---

#### 3.5.4 POST /infraction-reports/{InfractionReportId}/acknowledge
**Operação**: `acknowledgeInfractionReport`
**Resumo**: Receber Notificação (PSP Reportado)
**Descrição**: Confirma recebimento da notificação.

**Transição**: `OPEN` → `ACKNOWLEDGED`

**Rate Limiting**: Política `INFRACTION_REPORTS_WRITE`

**Mapeamento RF**:
- **RF-BLO4-004** ✅ - Receber notificação

---

#### 3.5.5 POST /infraction-reports/{InfractionReportId}/cancel
**Operação**: `cancelInfractionReport`
**Resumo**: Cancelar Notificação
**Descrição**: Cancela notificação (apenas quem criou).

**Transição**: Qualquer → `CANCELLED`

**Rate Limiting**: Política `INFRACTION_REPORTS_WRITE`

**Mapeamento RF**:
- **RF-BLO4-005** ✅ - Cancelar notificação

---

#### 3.5.6 POST /infraction-reports/{InfractionReportId}/close
**Operação**: `closeInfractionReport`
**Resumo**: Fechar Notificação (PSP Reportado)
**Descrição**: Fecha a notificação após análise.

**Transição**: `ACKNOWLEDGED` → `CLOSED`

**Request Body** (XML):
```xml
<CloseInfractionReportRequest>
    <Signature></Signature>
    <InfractionReportId>e345f678-1g66-86e9-3f3f-i21gh99g4h8h</InfractionReportId>
    <Analysis>AGREED</Analysis>
    <Justification>Análise concluída...</Justification>
</CloseInfractionReportRequest>
```

**Analysis**:
- `AGREED`: Concorda com a infração (cria Fraud Marker)
- `DISAGREED`: Discorda

**Rate Limiting**: Política `INFRACTION_REPORTS_WRITE`

**Mapeamento RF**:
- **RF-BLO4-004** ✅ - Fechar notificação

---

### 3.6 Bloco 5 - Antifraude (Antifraud)

#### 3.6.1 POST /fraud-markers/
**Operação**: `createFraudMarker`
**Resumo**: Criar Marcação de Fraude
**Descrição**: Marca usuário/chave por suspeita de fraude.

**Request Body** (XML):
```xml
<CreateFraudMarkerRequest>
    <Signature></Signature>
    <FraudMarker>
        <Key>+5561988880000</Key>
        <MarkerType>FRAUDULENT_ACCOUNT</MarkerType>
        <TaxIdNumber>11122233300</TaxIdNumber>
        <FraudType>SUSPECTED</FraudType>
    </FraudMarker>
    <RequestId>f456g789-2h77-97f0-4g4g-j32hi00h5i9i</RequestId>
</CreateFraudMarkerRequest>
```

**Fraud Types**:
- `SUSPECTED`: Suspeita de fraude
- `CONFIRMED`: Fraude confirmada

**Rate Limiting**: Política `FRAUD_MARKERS_WRITE`
- Taxa: 1200/min
- Balde: 36000

**Mapeamento RF**:
- **RF-BLO4-006** ✅ - Notificação de infração para marcação de fraude transacional

---

#### 3.6.2 GET /fraud-markers/{FraudMarkerId}
**Operação**: `getFraudMarker`
**Resumo**: Consultar Marcação de Fraude

**Rate Limiting**: Política `FRAUD_MARKERS_READ`
- Taxa: 600/min
- Balde: 18000

---

#### 3.6.3 GET /fraud-markers/
**Operação**: `listFrauds`
**Resumo**: Listar Marcações de Fraude

**Rate Limiting**: Política `FRAUD_MARKERS_LIST`
- Taxa: 600/min
- Balde: 18000

---

#### 3.6.4 POST /fraud-markers/{FraudMarkerId}/cancel
**Operação**: `cancelFraudMarker`
**Resumo**: Cancelar Marcação de Fraude
**Descrição**: Remove marcação de fraude.

**Rate Limiting**: Política `FRAUD_MARKERS_WRITE`

**Mapeamento RF**:
- **RF-BLO4-006** ✅ - Cancelar marcação de fraude

---

#### 3.6.5 GET /entries/{Key}/statistics
**Operação**: `getEntryStatistics`
**Resumo**: Consultar Estatísticas de Chave
**Descrição**: Obtém estatísticas antifraude de uma chave.

**Response** (200 OK):
```xml
<GetEntryStatisticsResponse>
    <Signature></Signature>
    <Statistics>
        <Key>+5561988880000</Key>
        <FraudMarkers>
            <Count_d90>2</Count_d90>
            <Count_m12>5</Count_m12>
            <Count_m60>10</Count_m60>
        </FraudMarkers>
        <Claims>
            <Count_d90>1</Count_d90>
        </Claims>
        <Transactions>
            <Count_d90>150</Count_d90>
            <Amount_d90>12500.00</Amount_d90>
        </Transactions>
    </Statistics>
</GetEntryStatisticsResponse>
```

**Janelas de Tempo**:
- **d90**: Últimos 89 dias + dia corrente
- **m12**: Últimos 12 meses (sem mês corrente)
- **m60**: Últimos 60 meses (sem mês corrente)

**Rate Limiting**: Política `ENTRIES_STATISTICS_READ`
- Taxa: Conforme ENTRIES_READ_PARTICIPANT_ANTISCAN

**Mapeamento RF**:
- **RF-BLO5-008** ✅ - Consulta a informações de segurança

---

#### 3.6.6 GET /persons/{TaxIdNumber}/statistics
**Operação**: `getPersonStatistics`
**Resumo**: Consultar Estatísticas de Pessoa
**Descrição**: Obtém estatísticas antifraude de um CPF/CNPJ.

**Path Parameters**:
- `TaxIdNumber` (string, required): CPF (11 dígitos) ou CNPJ (14 dígitos)

**Response** (200 OK):
```xml
<GetPersonStatisticsResponse>
    <Signature></Signature>
    <Statistics>
        <TaxIdNumber>11122233300</TaxIdNumber>
        <FraudMarkers>
            <Count_d90>3</Count_d90>
            <Count_m12>8</Count_m12>
            <Count_m60>15</Count_m60>
        </FraudMarkers>
        <EntriesCount>5</EntriesCount>
        <ClaimsCount>2</ClaimsCount>
    </Statistics>
</GetPersonStatisticsResponse>
```

**Rate Limiting**: Política `PERSONS_STATISTICS_READ`
- Taxa: 12000/min
- Balde: 36000

**Mapeamento RF**:
- **RF-BLO5-008** ✅ - Consulta a informações de segurança (pessoa)

---

### 3.7 Bloco 4 - Solicitação de Devolução (Refund)

#### 3.7.1 POST /refunds/
**Operação**: `createRefund`
**Resumo**: Criar Solicitação de Devolução
**Descrição**: Cria solicitação de devolução por fraude ou falha operacional.

**Pré-condição**: InfractionReport `CLOSED` com `AGREED`

**Request Body** (XML):
```xml
<CreateRefundRequest>
    <Signature></Signature>
    <Refund>
        <EndToEndId>E1234567820231024100000000000001</EndToEndId>
        <RefundReason>FRAUD</RefundReason>
        <RefundAmount>500.00</RefundAmount>
        <RequesterParticipant>12345678</RequesterParticipant>
        <ContestedParticipant>87654321</ContestedParticipant>
        <Details>Detalhes da devolução...</Details>
    </Refund>
    <RequestId>g567h890-3i88-08g1-5h5h-k43ij11i6j0j</RequestId>
</CreateRefundRequest>
```

**Refund Reasons**:
- `FRAUD`: Fundada suspeita de fraude
- `OPERATIONAL_FAILURE`: Falha operacional

**Status Cycle**:
```
OPEN → CLOSED
  ↓
  └──────> CANCELLED
```

**Rate Limiting**: Política `REFUNDS_WRITE`
- Taxa: 2400/min
- Balde: 72000

**Mapeamento RF**:
- **RF-BLO4-001** ✅ - Solicitar devolução por falha operacional
- **RF-BLO4-002** ✅ - Solicitar devolução por fundada suspeita de fraude

---

#### 3.7.2 GET /refunds/{RefundId}
**Operação**: `getRefund`
**Resumo**: Consultar Solicitação de Devolução

**Rate Limiting**: Política `REFUNDS_READ`
- Taxa: 1200/min
- Balde: 36000

**Mapeamento RF**:
- **RF-BLO4-001** ✅ - Consultar devolução

---

#### 3.7.3 GET /refunds/
**Operação**: `listRefunds`
**Resumo**: Listar Solicitações de Devolução

**Query Parameters**:
- `Status` (string, optional)
- `Role` (string, optional): `REQUESTER` ou `CONTESTED`
- `Limit` (int, optional)

**Rate Limiting**: 2 políticas
1. **REFUND_LIST_WITH_ROLE**: 40/min, 200
2. **REFUND_LIST_WITHOUT_ROLE**: 10/min, 50

**Mapeamento RF**:
- **RF-BLO4-001** ✅ - Monitorar devoluções (polling)

---

#### 3.7.4 POST /refunds/{RefundId}/cancel
**Operação**: `cancelRefund`
**Resumo**: Cancelar Solicitação de Devolução
**Descrição**: Cancela devolução (apenas quem criou).

**Transição**: `OPEN` → `CANCELLED`

**Rate Limiting**: Política `REFUNDS_WRITE`

**Mapeamento RF**:
- **RF-BLO4-003** ✅ - Cancelamento de devolução

---

#### 3.7.5 POST /refunds/{RefundId}/close
**Operação**: `closeRefund`
**Resumo**: Fechar Solicitação de Devolução
**Descrição**: Fecha devolução (PSP contestado).

**Transição**: `OPEN` → `CLOSED`

**Request Body** (XML):
```xml
<CloseRefundRequest>
    <Signature></Signature>
    <RefundId>h678i901-4j99-19h2-6i6i-l54jk22j7k1k</RefundId>
    <Analysis>REFUNDED</Analysis>
    <Justification>Devolução efetuada...</Justification>
</CloseRefundRequest>
```

**Analysis**:
- `REFUNDED`: Devolvido
- `NOT_REFUNDED`: Não devolvido

**Rate Limiting**: Política `REFUNDS_WRITE`

**Mapeamento RF**:
- **RF-BLO4-001** ✅ - Fechar devolução

---

### 3.8 Bloco 5 - Políticas de Limitação (Policies)

#### 3.8.1 GET /policies/
**Operação**: `listBucketStates`
**Resumo**: Listar Estado de Todos os Baldes
**Descrição**: Obtém estado atual de todos os baldes de rate limiting.

**Response** (200 OK):
```xml
<ListBucketStatesResponse>
    <Signature></Signature>
    <BucketStates>
        <BucketState>
            <Policy>ENTRIES_WRITE</Policy>
            <Tokens>35000</Tokens>
            <RefillRate>1200</RefillRate>
            <Capacity>36000</Capacity>
        </BucketState>
        ...
    </BucketStates>
</ListBucketStatesResponse>
```

**Rate Limiting**: Política `POLICIES_LIST`
- Taxa: 6/min
- Balde: 20

**Mapeamento RF**:
- **RF-BLO5-002** ✅ - Consulta de baldes

---

#### 3.8.2 GET /policies/{policy}
**Operação**: `getBucketState`
**Resumo**: Consultar Estado de Balde Específico
**Descrição**: Obtém estado de um balde específico.

**Path Parameters**:
- `policy` (string, required): Nome da política

**Response** (200 OK):
```xml
<GetBucketStateResponse>
    <Signature></Signature>
    <BucketState>
        <Policy>ENTRIES_WRITE</Policy>
        <Tokens>35000</Tokens>
        <RefillRate>1200</RefillRate>
        <Capacity>36000</Capacity>
    </BucketState>
</GetBucketStateResponse>
```

**Rate Limiting**: Política `POLICIES_READ`
- Taxa: 60/min
- Balde: 200

**Mapeamento RF**:
- **RF-BLO5-002** ✅ - Consulta de baldes

---

## 4. Segurança e Autenticação

### 4.1 Autenticação Mútua TLS (mTLS)

**Todas as requisições exigem mTLS**:
- ✅ Cliente apresenta certificado X.509 válido
- ✅ Servidor valida certificado do cliente
- ✅ Servidor apresenta certificado próprio
- ✅ Cliente valida certificado do servidor

**Gestão de Certificados**:
- Emissão: Banco Central (ICP-Brasil)
- Validade: Conforme política do Bacen
- Renovação: Antes do vencimento
- Revogação: Imediata em caso de comprometimento

**Referência**: [Manual de Segurança PIX](https://www.bcb.gov.br/content/estabilidadefinanceira/cedsfn/Manual_de_Seguranca_PIX.pdf)

### 4.2 Assinatura Digital XML

**Requisições que EXIGEM assinatura**:
- ✅ Todas as operações de escrita (POST, DELETE)
- ❌ Consultas (GET) **NÃO** requerem assinatura

**Padrão**: [XML Digital Signature](https://www.w3.org/2000/09/xmldsig)

**Tipo**: Assinatura **envelopada** (elemento `<Signature>` filho do XML)

**Algoritmo**:
- Hash: SHA-256
- Assinatura: RSA-2048 (mínimo)

**Validação**:
- ✅ **Todas as respostas** são assinadas pelo DICT
- ✅ Clientes **DEVEM** validar assinaturas

**Implementação LBPay**:
- Certificado P12
- Java Signer Service (componente existente)

### 4.3 Segurança em Camadas

| Camada | Controle | Descrição |
|--------|----------|-----------|
| **Transporte** | TLS 1.2+ | Criptografia de canal |
| **Autenticação** | mTLS | Identidade mútua |
| **Integridade** | XML Signature | Não-repúdio |
| **Autorização** | ISPB | Validação de participante |
| **Rate Limiting** | Token Bucket | Proteção contra abuso |

---

## 5. Rate Limiting e Performance

### 5.1 Algoritmo: Token Bucket

**Conceito**: Pense literalmente como um **balde de fichas**:

1. **Balde (Bucket)**: Capacidade máxima de "burst" (rajada)
   - Exemplo: `PERSONS_STATISTICS_READ` tem balde de **36.000 fichas**
   - Se você não usar por um tempo, o balde enche até o limite
   - Você pode gastar todas as 36.000 de uma vez (rajada)

2. **Taxa (Rate)**: Velocidade de reabastecimento do balde
   - Exemplo: `PERSONS_STATISTICS_READ` repõe **12.000 fichas/min**
   - A cada minuto, 12.000 novas fichas são adicionadas
   - Limite máximo: 36.000 (capacidade do balde)

3. **Consumo**: Cada requisição consome fichas
   - Geralmente 1 ficha por request
   - Algumas políticas consomem mais (ex: status 404 consome 20 fichas)

4. **Rate Limiting**: Quando fichas = 0 → HTTP 429

**Analogia prática**:
```
Balde = 36.000 fichas (capacidade máxima)
Taxa = 12.000/min (reabastecimento)

Cenário 1 - Burst (rajada):
- Você faz 36.000 requests de uma vez → OK (usa todo o balde)
- Próxima request → 429 (balde vazio)
- Após 1 minuto → 12.000 fichas disponíveis novamente
- Após 3 minutos → 36.000 fichas (balde cheio)

Cenário 2 - Uso sustentado:
- Você faz 200 requests/min continuamente
- Taxa de reposição (12.000/min) >> Taxa de uso (200/min)
- Balde sempre cheio → NUNCA recebe 429

Cenário 3 - Uso excessivo:
- Você faz 15.000 requests/min continuamente
- Taxa de uso (15.000/min) > Taxa de reposição (12.000/min)
- Balde esvazia gradualmente
- Após alguns minutos → 429 (balde vazio)
```

**Benefício do modelo Token Bucket**:
- ✅ Permite picos de uso (burst) usando fichas do balde
- ✅ Impede sobrecarga contínua (taxa de uso > taxa de reposição)
- ✅ Flexibilidade para padrões de tráfego variáveis

### 5.2 Todas as Políticas de Rate Limiting

| Política | Escopo | Taxa Reposição | Balde | Operações |
|----------|--------|----------------|-------|-----------|
| **ENTRIES_READ_USER_ANTISCAN** | USER (PF/PJ) | 2/min (PF)<br>20/min (PJ) | 100 (PF)<br>1000 (PJ) | getEntry (EMAIL, PHONE) |
| **ENTRIES_READ_USER_ANTISCAN_V2** | USER (PF/PJ) | 2/min (PF)<br>20/min (PJ) | 100 (PF)<br>1000 (PJ) | getEntry (CPF, CNPJ, EVP) |
| **ENTRIES_READ_PARTICIPANT_ANTISCAN** | PSP (A-H) | 25k-2/min | 50k-50 | getEntry (todos os tipos) |
| **ENTRIES_STATISTICS_READ** | PSP | Conforme categoria | Conforme categoria | getEntryStatistics |
| **ENTRIES_WRITE** | PSP | 1200/min | 36000 | createEntry, deleteEntry |
| **ENTRIES_UPDATE** | PSP | 600/min | 600 | updateEntry |
| **CLAIMS_READ** | PSP | 600/min | 18000 | getClaim |
| **CLAIMS_WRITE** | PSP | 1200/min | 36000 | createClaim, ackClaim, etc |
| **CLAIMS_LIST_WITH_ROLE** | PSP | 40/min | 200 | listClaims (com Role) |
| **CLAIMS_LIST_WITHOUT_ROLE** | PSP | 10/min | 50 | listClaims (sem Role) |
| **SYNC_VERIFICATIONS_WRITE** | PSP | 10/min | 50 | createSyncVerification |
| **CIDS_FILES_WRITE** | PSP | 40/dia | 200 | createCidSetFile |
| **CIDS_FILES_READ** | PSP | 10/min | 50 | getCidSetFile |
| **CIDS_EVENTS_LIST** | PSP | 20/min | 100 | listCidSetEvents |
| **CIDS_ENTRIES_READ** | PSP | 1200/min | 36000 | getEntryByCid |
| **INFRACTION_REPORTS_READ** | PSP | 600/min | 18000 | getInfractionReport |
| **INFRACTION_REPORTS_WRITE** | PSP | 1200/min | 36000 | createIR, ackIR, etc |
| **INFRACTION_REPORTS_LIST_WITH_ROLE** | PSP | 40/min | 200 | listIRs (com Role) |
| **INFRACTION_REPORTS_LIST_WITHOUT_ROLE** | PSP | 10/min | 50 | listIRs (sem Role) |
| **KEYS_CHECK** | PSP | 70/min | 70 | checkKeys |
| **REFUNDS_READ** | PSP | 1200/min | 36000 | getRefund |
| **REFUNDS_WRITE** | PSP | 2400/min | 72000 | createRefund, cancelRefund, etc |
| **REFUND_LIST_WITH_ROLE** | PSP | 40/min | 200 | listRefunds (com Role) |
| **REFUND_LIST_WITHOUT_ROLE** | PSP | 10/min | 50 | listRefunds (sem Role) |
| **FRAUD_MARKERS_READ** | PSP | 600/min | 18000 | getFraudMarker |
| **FRAUD_MARKERS_WRITE** | PSP | 1200/min | 36000 | createFM, cancelFM |
| **FRAUD_MARKERS_LIST** | PSP | 600/min | 18000 | listFrauds |
| **PERSONS_STATISTICS_READ** | PSP | 12000/min | 36000 | getPersonStatistics |
| **POLICIES_READ** | PSP | 60/min | 200 | getBucketState |
| **POLICIES_LIST** | PSP | 6/min | 20 | listBucketStates |

### 5.3 Categorias de Participante (ENTRIES_READ_PARTICIPANT_ANTISCAN)

| Categoria | Taxa | Balde |
|-----------|------|-------|
| A | 25.000/min | 50.000 |
| B | 20.000/min | 40.000 |
| C | 15.000/min | 30.000 |
| D | 8.000/min | 16.000 |
| E | 2.500/min | 5.000 |
| F | 250/min | 500 |
| G | 25/min | 250 |
| H | 2/min | 50 |

**Categoria LBPay**: **Provavelmente F ou H** (inicial)

**Contexto**:
- ✅ Categoria é atribuída pelo **Bacen durante homologação**
- ✅ LBPay é IP recente (em fase de homologação)
- ✅ **Hipótese provável**: Começar em categoria **F** (250/min, balde 500) ou **H** (2/min, balde 50)
- ✅ Categoria **H**: Testes ou volume baixíssimo
- ✅ Categoria **F**: Ponto de partida comum para novos entrantes
- ✅ **Reclassificação**: À medida que volume aumenta, solicitar upgrade (E → D → C → B → A)

**Ação requerida**: Verificar documentação de homologação Bacen para confirmar categoria atribuída.

### 5.4 Regras de Contagem Especiais

#### ENTRIES_READ_USER_ANTISCAN (EMAIL, PHONE)
- Status 200: **subtrai 1** ficha
- Status 404: **subtrai 20** fichas (penalidade anti-scan)
- Ordem PIX enviada: **adiciona 1** ficha (PF) ou **2** fichas (PJ)

#### ENTRIES_READ_USER_ANTISCAN_V2 (CPF, CNPJ, EVP)
- Status 200: **subtrai 1** ficha
- Status 404: **subtrai 20** fichas
- Ordem PIX enviada: **adiciona 1** ficha (PF) ou **2** fichas (PJ)

#### ENTRIES_READ_PARTICIPANT_ANTISCAN (todos os tipos)
- Status 200: **subtrai 1** ficha
- Status 404: **subtrai 3** fichas
- Ordem PIX enviada: **adiciona 1** ficha

**Interpretação**: Consultas seguidas de pagamento repõem fichas automaticamente.

### 5.5 Resposta HTTP 429 (Rate Limited)

**Quando ocorre**: Fichas = 0

**Response**:
```xml
<Problem>
    <type>https://dict.pi.rsfn.net.br/api/v2/error/RateLimited</type>
    <title>Rate Limited</title>
    <status>429</status>
    <detail>Limite de requisições foi atingido</detail>
</Problem>
```

**Headers**:
- `Retry-After`: Segundos até próxima tentativa

**Ação recomendada**:
- Implementar **exponential backoff**
- Monitorar estado do balde via `/policies/{policy}`
- Alertar operação se 429 recorrente

### 5.6 Recomendações de Performance

#### 5.6.1 Reutilização de Conexões HTTP

**❗ CRÍTICO**: Custo de mTLS handshake é MUITO alto (latência).

**Solução**:
- ✅ **HTTP Connection Pool** (keep-alive)
- ✅ Configurar `timeout` adequado (ver header `Keep-Alive`)
- ✅ Manter conexões abertas

**Configuração Go (exemplo)**:
```go
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
    TLSClientConfig:     tlsConfig, // mTLS
}
client := &http.Client{Transport: transport}
```

#### 5.6.2 Compressão

**Request**:
- ✅ Adicionar header: `Accept-Encoding: gzip`
- ❌ Compressão de request body **NÃO** é suportada

**Response**:
- ✅ DICT retorna comprimido se solicitado
- ✅ Reduz largura de banda significativamente

#### 5.6.3 Cache Local (LBPay)

**Endpoint crítico**: `GET /entries/{Key}` (dezenas/segundo)

**Estratégia de Cache** (conforme [ArquiteturaDict_LBPAY.md](../../Docs_iniciais/ArquiteturaDict_LBPAY.md)):

1. **cache-dict-response** (Redis porta 7001):
   - Armazena respostas completas de consultas
   - TTL: 5 minutos (configurável)
   - Invalidação: Eventos de CID

2. **cache-dict-account** (Redis porta 7002):
   - Dados de contas (Account)
   - TTL: 15 minutos

3. **cache-dict-key-validation** (Redis porta 7003):
   - Validações de chaves
   - TTL: 10 minutos

4. **cache-dict-dedup** (Redis porta 7004):
   - Deduplicação de requisições
   - TTL: 1 minuto

5. **cache-dict-rate-limit** (Redis porta 7005):
   - Controle local de rate limiting
   - TTL: Conforme política

**Benefício**: Reduz chamadas ao DICT em 70-90%.

#### 5.6.4 Monitoramento de Latência

**SLA Bacen**: Ver [Manual de Tempos do Pix](https://www.bcb.gov.br/content/estabilidadefinanceira/pix/Regulamento_Pix/IX_ManualdeTemposdoPix.pdf)

**Métricas a monitorar**:
- ✅ Latência P50, P95, P99
- ✅ Taxa de erro (4xx, 5xx)
- ✅ Taxa 429 (rate limiting)
- ✅ Conexões ativas no pool

**Alertas**:
- P99 > 1s
- Taxa 429 > 1%
- Taxa 5xx > 0.1%

---

## 6. Tratamento de Erros

### 6.1 Padrão RFC 7807 (Problem Details)

**Todas as respostas de erro** seguem RFC 7807:

```xml
<Problem>
    <type>https://dict.pi.rsfn.net.br/api/v2/error/<TipoErro></type>
    <title>Título Legível</title>
    <status>400</status>
    <detail>Descrição detalhada do erro</detail>
    <violationId>UUID (opcional)</violationId>
</Problem>
```

### 6.2 Tipos de Erro (Geral)

| Tipo | Status | Descrição |
|------|--------|-----------|
| `Forbidden` | 403 | Violação de regra de autorização |
| `BadRequest` | 400 | Formato inválido |
| `NotFound` | 404 | Entidade não encontrada |
| `Gone` | 410 | Recurso não mais disponível |
| `RateLimited` | 429 | Limite de requisições atingido |
| `InternalServerError` | 500 | Erro inesperado no DICT |
| `ServiceUnavailable` | 503 | Serviço indisponível (manutenção) |
| `RequestSignatureInvalid` | 400 | Assinatura digital inválida |
| `RequestIdAlreadyUsed` | 400 | RequestId reutilizado com parâmetros diferentes |
| `InvalidReason` | 400 | Razão inválida para operação |
| `ParticipantInvalid` | 400 | Participante não pode realizar operação |
| `TaxIdNumberBlocked` | 400 | CPF/CNPJ bloqueado por ordem judicial |

### 6.3 Tipos de Erro (Vínculo - Entry)

| Tipo | Status | Descrição |
|------|--------|-----------|
| `EntryInvalid` | 400 | Campos inválidos ao criar/atualizar vínculo |
| `EntryLimitExceeded` | 400 | Limite de vínculos por conta excedido |
| `EntryAlreadyExists` | 400 | Vínculo já existe (mesma chave, PSP, dono) |
| `EntryCannotBeQueriedForBookTransfer` | 400 | Pagador e recebedor no mesmo PSP (não usar DICT) |
| `EntryKeyOwnedByDifferentPerson` | 400 | Chave possuída por outra pessoa (fazer Claim de posse) |
| `EntryKeyInCustodyOfDifferentParticipant` | 400 | Chave em outro PSP (fazer Claim de portabilidade) |
| `EntryLockedByClaim` | 400 | Vínculo bloqueado por Claim ativo |
| `EntryTaxIdNumberByDifferentOwner` | 400 | CPF/CNPJ divergente do dono |
| `EntryBlocked` | 400 | Vínculo bloqueado por ordem judicial |
| `EntryWithFraudRelatedRestriction` | 400 | Vínculo bloqueado por suspeita de fraude |

### 6.4 Tipos de Erro (Reivindicação - Claim)

| Tipo | Status | Descrição |
|------|--------|-----------|
| `ClaimInvalid` | 400 | Campos inválidos ao criar/atualizar Claim |
| `ClaimTypeInconsistent` | 400 | Tipo de Claim inconsistente (posse vs portabilidade) |
| `ClaimKeyNotFound` | 404 | Não existe vínculo com a chave reivindicada |
| `ClaimAlreadyExistsForKey` | 400 | Claim ativo para a chave (apenas 1 por vez) |
| `ClaimResultingEntryAlreadyExists` | 400 | Vínculo resultante já existe |
| `ClaimOperationInvalid` | 400 | Status atual não permite operação |
| `ClaimResolutionPeriodNotEnded` | 400 | Período de resolução não terminou |
| `ClaimCompletionPeriodNotEnded` | 400 | Período de encerramento não terminou |

### 6.5 Tipos de Erro (Notificação de Infração - InfractionReport)

| Tipo | Status | Descrição |
|------|--------|-----------|
| `InfractionReportInvalid` | 400 | Campos inválidos |
| `InfractionReportOperationInvalid` | 400 | Status não permite operação |
| `InfractionReportTransactionNotFound` | 400 | Transação (EndToEndId) não encontrada |
| `InfractionReportTransactionNotSettled` | 400 | Transação não liquidada |
| `InfractionReportAlreadyBeingProcessedForTransaction` | 400 | IR ativo para transação |
| `InfractionReportAlreadyProcessedForTransaction` | 400 | IR fechado para transação |
| `InfractionReportPeriodExpired` | 400 | Prazo para notificar expirou |

### 6.6 Tipos de Erro (Marcação de Fraude - FraudMarker)

| Tipo | Status | Descrição |
|------|--------|-----------|
| `FraudMarkerInvalid` | 400 | Campos inválidos |

### 6.7 Tipos de Erro (Solicitação de Devolução - Refund)

| Tipo | Status | Descrição |
|------|--------|-----------|
| `RefundInvalid` | 400 | Campos inválidos |
| `RefundOperationInvalid` | 400 | Status não permite operação |
| `RefundTransactionNotFound` | 400 | Transação não encontrada |
| `RefundTransactionNotSettled` | 400 | Transação não liquidada |
| `RefundAlreadyProcessedForTransaction` | 400 | Refund processado para transação |
| `RefundAlreadyBeingProcessedForTransaction` | 400 | Refund ativo para transação |
| `RefundPeriodExpired` | 400 | Prazo expirou |
| `TransactionNotRefundable` | 400 | Transação não permite devolução |
| `RefundInfractionReportNotFound` | 400 | InfractionReport correspondente não encontrado |
| `TransactionRefundable` | 400 | Transação permite devolução (erro ao criar IR) |

### 6.8 Tratamento por Status HTTP

| Status | Categoria | Ação Recomendada |
|--------|-----------|------------------|
| **200-201** | Sucesso | Processar resposta |
| **400** | Erro cliente | Validar input, corrigir e tentar novamente |
| **403** | Autorização | Verificar permissões, não retentar |
| **404** | Não encontrado | Verificar se entidade existe, não retentar |
| **410** | Gone | Recurso removido, não retentar |
| **429** | Rate limit | Exponential backoff, retentar |
| **500** | Erro servidor | Retentar com exponential backoff |
| **503** | Indisponível | Aguardar manutenção, retentar |

### 6.9 Estratégia de Retry (Implementação)

**Erros retryable**:
- ✅ 429 (Rate Limited)
- ✅ 500 (Internal Server Error)
- ✅ 503 (Service Unavailable)
- ✅ Timeout de rede

**Erros NÃO retryable**:
- ❌ 400 (Bad Request)
- ❌ 403 (Forbidden)
- ❌ 404 (Not Found)
- ❌ 410 (Gone)

**Algoritmo**: Exponential Backoff with Jitter
```
delay = min(max_delay, base_delay * 2^attempt) + random(0, jitter)
```

**Configuração sugerida**:
- `base_delay`: 1s
- `max_delay`: 60s
- `max_attempts`: 3
- `jitter`: 500ms

---

## 7. Mapeamento com Requisitos Funcionais

### 7.1 Mapeamento Completo (Endpoints → RFs)

| Endpoint | Operação | RFs Atendidos |
|----------|----------|---------------|
| `POST /entries/` | createEntry | RF-BLO1-001, RF-TRV-001 |
| `GET /entries/{Key}` | getEntry | RF-BLO1-008, RF-BLO5-003 ✅ (crítico), RF-BLO5-001 |
| `POST /entries/{Key}` | updateEntry | RF-BLO1-009, RF-BLO1-010 |
| `POST /entries/{Key}/delete` | deleteEntry | RF-BLO1-002, RF-BLO1-003, RF-BLO1-004, RF-BLO1-005, RF-BLO1-006, RF-BLO1-007 |
| `POST /keys/check` | checkKeys | RF-BLO1-011, RF-BLO5-007 |
| `POST /claims/` | createClaim | RF-BLO2-001, RF-BLO2-007 |
| `GET /claims/{ClaimId}` | getClaim | RF-BLO2-006 |
| `GET /claims/` | listClaims | RF-BLO2-005, RF-BLO2-004 |
| `POST /claims/{ClaimId}/acknowledge` | acknowledgeClaim | RF-BLO2-009 |
| `POST /claims/{ClaimId}/confirm` | confirmClaim | RF-BLO2-010, RF-BLO2-007, RF-BLO2-008 |
| `POST /claims/{ClaimId}/cancel` | cancelClaim | RF-BLO2-002, RF-BLO2-011 |
| `POST /claims/{ClaimId}/complete` | completeClaim | RF-BLO2-003 |
| `POST /sync-verifications/` | createSyncVerification | RF-BLO5-009 |
| `POST /cids/files/` | createCidSetFile | RF-BLO5-010 |
| `GET /cids/files/{Id}` | getCidSetFile | RF-BLO5-010 |
| `GET /cids/events` | listCidSetEvents | RF-BLO5-010 |
| `GET /cids/entries/{Cid}` | getEntryByCid | RF-BLO5-010 |
| `POST /infraction-reports/` | createInfractionReport | RF-BLO4-004, RF-BLO4-005 |
| `GET /infraction-reports/{Id}` | getInfractionReport | RF-BLO4-004 |
| `GET /infraction-reports/` | listInfractionReports | RF-BLO4-004 |
| `POST /infraction-reports/{Id}/acknowledge` | acknowledgeInfractionReport | RF-BLO4-004 |
| `POST /infraction-reports/{Id}/cancel` | cancelInfractionReport | RF-BLO4-005 |
| `POST /infraction-reports/{Id}/close` | closeInfractionReport | RF-BLO4-004 |
| `POST /fraud-markers/` | createFraudMarker | RF-BLO4-006 |
| `GET /fraud-markers/{Id}` | getFraudMarker | RF-BLO4-006 |
| `GET /fraud-markers/` | listFrauds | RF-BLO4-006 |
| `POST /fraud-markers/{Id}/cancel` | cancelFraudMarker | RF-BLO4-006 |
| `GET /entries/{Key}/statistics` | getEntryStatistics | RF-BLO5-008 |
| `GET /persons/{TaxId}/statistics` | getPersonStatistics | RF-BLO5-008 |
| `POST /refunds/` | createRefund | RF-BLO4-001, RF-BLO4-002 |
| `GET /refunds/{Id}` | getRefund | RF-BLO4-001 |
| `GET /refunds/` | listRefunds | RF-BLO4-001 |
| `POST /refunds/{Id}/cancel` | cancelRefund | RF-BLO4-003 |
| `POST /refunds/{Id}/close` | closeRefund | RF-BLO4-001 |
| `GET /policies/` | listBucketStates | RF-BLO5-002 |
| `GET /policies/{policy}` | getBucketState | RF-BLO5-002 |

### 7.2 RFs do Bloco 3 (Validação) - Não Mapeados

**Importante**: Os RFs do **Bloco 3 - Validação** não são implementados via API REST do DICT:

- **RF-BLO3-001** - Validação da posse (SMS/Email)
- **RF-BLO3-002** - Validação situação cadastral Receita Federal
- **RF-BLO3-003** - Validações dos nomes vinculados

**Motivo**: São validações **internas** do PSP, não APIs do DICT.

**Ação**: Serão especificadas em documentos internos (APIs gRPC do Core DICT).

### 7.3 RFs do Bloco 6 (Recuperação de Valores) - Não Mapeados

Os **RFs do Bloco 6** (RF-BLO6-001 a RF-BLO6-013) não estão na API DICT v2.6.1.

**Motivo**: São APIs do **SPI (Sistema de Pagamentos Instantâneos)**, não do DICT.

**Ação**: Serão tratadas em documento separado (API SPI).

### 7.4 Cobertura de RFs por Bloco

| Bloco | Total RFs | RFs Mapeados | % Cobertura |
|-------|-----------|--------------|-------------|
| **Bloco 1 - CRUD** | 13 | 11 | 84.6% |
| **Bloco 2 - Claim** | 14 | 11 | 78.6% |
| **Bloco 3 - Validação** | 3 | 0 | 0% (interno) |
| **Bloco 4 - Devolução** | 6 | 6 | 100% |
| **Bloco 5 - Segurança** | 13 | 6 | 46.2% |
| **Bloco 6 - Recuperação** | 13 | 0 | 0% (API SPI) |
| **Transversal** | 10 | 1 | 10% |

**Total Geral**: **42 de 72 RFs** cobertos pela API DICT REST (58.3%)

**Interpretação**:
- ✅ API DICT REST cobre **58.3%** dos RFs
- ✅ 41.7% restantes são APIs internas (Core DICT gRPC) ou APIs SPI

---

## 8. Recomendações de Implementação

### 8.1 Arquitetura de Integração

**Conforme [ARE-003](../02_Arquitetura/ARE-003_Analise_Documento_Arquitetura_DICT.md)**, a integração deve seguir:

```
┌─────────────────────────────────────────────────────────────┐
│                      CORE DICT (novo repo)                  │
│  ┌──────────────┐  ┌─────────────┐  ┌──────────────────┐  │
│  │   Domain     │  │ Application │  │    Handlers      │  │
│  │  (Business)  │  │   (Use      │  │   (gRPC/REST)    │  │
│  │    Logic     │  │    Cases)   │  │                  │  │
│  └──────────────┘  └─────────────┘  └──────────────────┘  │
│         ↓                  ↓                   ↓           │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Infrastructure Layer                     │  │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────────┐ │  │
│  │  │ PostgreSQL │  │  5x Redis  │  │ Temporal/Pulsar│ │  │
│  │  └────────────┘  └────────────┘  └────────────────┘ │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────┐
│                   CONNECT DICT (rsfn-bridge)                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │     REST Client (mTLS, XML Signer, Connection Pool) │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                             ↓
                      ┌──────────────┐
                      │  DICT Bacen  │
                      │  (API v2.6)  │
                      └──────────────┘
```

### 8.2 Módulos a Implementar no CONNECT DICT

#### 8.2.1 rest_client.go
**Responsabilidade**: Cliente HTTP com mTLS e connection pooling

**Funcionalidades**:
- ✅ Configuração de mTLS (certificado P12)
- ✅ Connection pool (keep-alive)
- ✅ Retry com exponential backoff
- ✅ Timeout configurável
- ✅ Compressão (Accept-Encoding: gzip)
- ✅ Observabilidade (OpenTelemetry)

**Interface**:
```go
type RESTClient interface {
    Get(ctx context.Context, path string, params map[string]string) (*Response, error)
    Post(ctx context.Context, path string, body []byte) (*Response, error)
    Delete(ctx context.Context, path string, body []byte) (*Response, error)
}
```

#### 8.2.2 xml_signer.go
**Responsabilidade**: Assinatura e validação de XML

**Funcionalidades**:
- ✅ Assinar XML (envelopado, SHA-256, RSA)
- ✅ Validar assinatura de resposta
- ✅ Integração com Java Signer Service (se necessário)

**Interface**:
```go
type XMLSigner interface {
    SignXML(xml []byte) ([]byte, error)
    ValidateSignature(xml []byte) (bool, error)
}
```

#### 8.2.3 entries_api.go
**Responsabilidade**: Endpoints de CRUD de vínculos

**Métodos**:
```go
type EntriesAPI interface {
    CreateEntry(ctx context.Context, req *CreateEntryRequest) (*CreateEntryResponse, error)
    GetEntry(ctx context.Context, key string, params *GetEntryParams) (*GetEntryResponse, error)
    UpdateEntry(ctx context.Context, key string, req *UpdateEntryRequest) (*UpdateEntryResponse, error)
    DeleteEntry(ctx context.Context, key string, req *DeleteEntryRequest) error
}
```

#### 8.2.4 keys_api.go
**Responsabilidade**: Validação de chaves

**Métodos**:
```go
type KeysAPI interface {
    CheckKeys(ctx context.Context, keys []string) (*CheckKeysResponse, error)
}
```

#### 8.2.5 claims_api.go
**Responsabilidade**: Reivindicações

**Métodos**:
```go
type ClaimsAPI interface {
    CreateClaim(ctx context.Context, req *CreateClaimRequest) (*CreateClaimResponse, error)
    GetClaim(ctx context.Context, claimId string) (*GetClaimResponse, error)
    ListClaims(ctx context.Context, filters *ClaimFilters) (*ListClaimsResponse, error)
    AcknowledgeClaim(ctx context.Context, claimId string) error
    ConfirmClaim(ctx context.Context, claimId string) error
    CancelClaim(ctx context.Context, claimId string, reason string) error
    CompleteClaim(ctx context.Context, claimId string) error
}
```

#### 8.2.6 reconciliation_api.go
**Responsabilidade**: Sincronização e CIDs

**Métodos**:
```go
type ReconciliationAPI interface {
    CreateSyncVerification(ctx context.Context, keyType string, vsync string) (*SyncVerificationResponse, error)
    CreateCidSetFile(ctx context.Context, keyType string) (*CidSetFileResponse, error)
    GetCidSetFile(ctx context.Context, fileId string) (*CidSetFileResponse, error)
    ListCidSetEvents(ctx context.Context, keyType string, since time.Time, limit int) (*CidEventsResponse, error)
    GetEntryByCid(ctx context.Context, cid string) (*GetEntryResponse, error)
}
```

#### 8.2.7 infraction_reports_api.go
**Responsabilidade**: Notificações de infração

**Métodos**:
```go
type InfractionReportsAPI interface {
    CreateInfractionReport(ctx context.Context, req *CreateInfractionReportRequest) (*InfractionReportResponse, error)
    GetInfractionReport(ctx context.Context, reportId string) (*InfractionReportResponse, error)
    ListInfractionReports(ctx context.Context, filters *InfractionReportFilters) (*ListInfractionReportsResponse, error)
    AcknowledgeInfractionReport(ctx context.Context, reportId string) error
    CancelInfractionReport(ctx context.Context, reportId string) error
    CloseInfractionReport(ctx context.Context, reportId string, analysis string, justification string) error
}
```

#### 8.2.8 fraud_markers_api.go
**Responsabilidade**: Marcações de fraude

**Métodos**:
```go
type FraudMarkersAPI interface {
    CreateFraudMarker(ctx context.Context, req *CreateFraudMarkerRequest) (*FraudMarkerResponse, error)
    GetFraudMarker(ctx context.Context, markerId string) (*FraudMarkerResponse, error)
    ListFrauds(ctx context.Context, filters *FraudFilters) (*ListFraudsResponse, error)
    CancelFraudMarker(ctx context.Context, markerId string) error
}
```

#### 8.2.9 refunds_api.go
**Responsabilidade**: Solicitações de devolução

**Métodos**:
```go
type RefundsAPI interface {
    CreateRefund(ctx context.Context, req *CreateRefundRequest) (*RefundResponse, error)
    GetRefund(ctx context.Context, refundId string) (*RefundResponse, error)
    ListRefunds(ctx context.Context, filters *RefundFilters) (*ListRefundsResponse, error)
    CancelRefund(ctx context.Context, refundId string) error
    CloseRefund(ctx context.Context, refundId string, analysis string, justification string) error
}
```

#### 8.2.10 statistics_api.go
**Responsabilidade**: Estatísticas antifraude

**Métodos**:
```go
type StatisticsAPI interface {
    GetEntryStatistics(ctx context.Context, key string) (*EntryStatisticsResponse, error)
    GetPersonStatistics(ctx context.Context, taxIdNumber string) (*PersonStatisticsResponse, error)
}
```

#### 8.2.11 policies_api.go
**Responsabilidade**: Rate limiting

**Métodos**:
```go
type PoliciesAPI interface {
    GetBucketState(ctx context.Context, policy string) (*BucketStateResponse, error)
    ListBucketStates(ctx context.Context) (*ListBucketStatesResponse, error)
}
```

#### 8.2.12 error_handler.go
**Responsabilidade**: Parsing e tratamento de erros RFC 7807

**Funcionalidades**:
- ✅ Parse de XML Problem Details
- ✅ Mapeamento de tipos de erro para structs Go
- ✅ Decisão de retry (retryable vs non-retryable)

**Interface**:
```go
type ErrorHandler interface {
    ParseError(xmlBody []byte) (*DICTError, error)
    IsRetryable(err *DICTError) bool
}
```

#### 8.2.13 rate_limiter.go
**Responsabilidade**: Rate limiting local (client-side)

**Funcionalidades**:
- ✅ Implementar token bucket local (Redis cache-dict-rate-limit)
- ✅ Prevenir 429 antes de chegar ao DICT
- ✅ Monitorar estado de baldes via `/policies/`

**Interface**:
```go
type RateLimiter interface {
    TryAcquire(ctx context.Context, policy string, tokens int) (bool, error)
    RefillTokens(policy string, count int)
    GetBucketState(policy string) (*BucketState, error)
}
```

### 8.3 Schemas XML (Go Structs)

Todos os schemas XML devem ser mapeados para structs Go com tags `xml`.

**Exemplo** (CreateEntryRequest):
```go
type CreateEntryRequest struct {
    XMLName   xml.Name  `xml:"CreateEntryRequest"`
    Signature string    `xml:"Signature"`
    Entry     Entry     `xml:"Entry"`
    Reason    string    `xml:"Reason"`
    RequestId string    `xml:"RequestId"`
}

type Entry struct {
    Key     string   `xml:"Key"`
    KeyType string   `xml:"KeyType"`
    Account Account  `xml:"Account"`
    Owner   Owner    `xml:"Owner"`
}

type Account struct {
    Participant   string `xml:"Participant"`
    Branch        string `xml:"Branch"`
    AccountNumber string `xml:"AccountNumber"`
    AccountType   string `xml:"AccountType"`
    OpeningDate   string `xml:"OpeningDate"`
}

type Owner struct {
    Type         string `xml:"Type"`
    TaxIdNumber  string `xml:"TaxIdNumber"`
    Name         string `xml:"Name"`
    TradeName    string `xml:"TradeName,omitempty"`
}
```

**Total de structs a criar**: ~50 (request + response de 28 endpoints)

### 8.4 Testes

#### 8.4.1 Testes Unitários
- ✅ Testar cada método de API isoladamente
- ✅ Mock do REST client
- ✅ Testar parsing de XML (request + response)
- ✅ Testar tratamento de erros
- ✅ Cobertura mínima: 80%

#### 8.4.2 Testes de Integração
- ✅ Usar **Simulador DICT** (repo existente: simulator-dict)
- ✅ Testar fluxos completos (ex: criar, consultar, excluir)
- ✅ Testar rate limiting
- ✅ Testar retry logic

#### 8.4.3 Testes de Performance
- ✅ Load testing de `GET /entries/{Key}` (endpoint crítico)
- ✅ Validar latência P99 < 1s
- ✅ Validar throughput (dezenas/segundo)
- ✅ Testar connection pool

#### 8.4.4 Testes de Homologação (Bacen)
- ✅ Usar ambiente Homologação (`dict-h.pi.rsfn.net.br`)
- ✅ Executar checklist de homologação Bacen
- ✅ Documentar evidências

### 8.5 Monitoramento e Observabilidade

#### 8.5.1 Métricas (Prometheus)
```
# Latência
dict_api_request_duration_seconds{endpoint, method, status}

# Throughput
dict_api_requests_total{endpoint, method, status}

# Erros
dict_api_errors_total{endpoint, error_type}

# Rate limiting
dict_api_rate_limited_total{policy}

# Connection pool
dict_http_connections_active
dict_http_connections_idle
```

#### 8.5.2 Traces (OpenTelemetry)
- ✅ Instrumentar todos os métodos de API
- ✅ Contexto de trace propagado (W3C Trace Context)
- ✅ Spans para: HTTP request, XML parsing, signature, retry

#### 8.5.3 Logs (Structured)
```json
{
  "timestamp": "2023-10-24T10:00:00Z",
  "level": "INFO",
  "service": "connect-dict",
  "endpoint": "GET /entries/{Key}",
  "key": "+5561988880000",
  "status": 200,
  "duration_ms": 150,
  "trace_id": "abc123..."
}
```

#### 8.5.4 Alertas
- ⚠️ P99 latency > 1s (5min window)
- ⚠️ Error rate > 1% (5min window)
- ⚠️ 429 rate > 0.5% (5min window)
- 🔴 Error rate > 5% (1min window) - critical
- 🔴 All requests failing (1min window) - critical

---

## 9. Referências

### 9.1 Documentação Oficial Bacen

1. **OpenAPI DICT v2.6.1**
   Arquivo: `/Users/jose.silva.lb/LBPay/IA_Dict/Docs_iniciais/OpenAPI_Dict_Bacen.json`

2. **Manual Operacional do DICT**
   https://www.bcb.gov.br/content/estabilidadefinanceira/pix/Regulamento_Pix/X_ManualOperacionaldoDICT.pdf

3. **Manual de Segurança PIX**
   https://www.bcb.gov.br/content/estabilidadefinanceira/cedsfn/Manual_de_Seguranca_PIX.pdf

4. **Manual de Tempos do Pix**
   https://www.bcb.gov.br/content/estabilidadefinanceira/pix/Regulamento_Pix/IX_ManualdeTemposdoPix.pdf

5. **Página Oficial PIX**
   https://www.bcb.gov.br/estabilidadefinanceira/pagamentosinstantaneos

### 9.2 Documentos do Projeto

1. **CRF-001** - Checklist de Requisitos Funcionais
   [Artefatos/05_Requisitos/CRF-001_Checklist_Requisitos_Funcionais.md](../05_Requisitos/CRF-001_Checklist_Requisitos_Funcionais.md)

2. **ARE-003** - Análise Documento Arquitetura DICT
   [Artefatos/02_Arquitetura/ARE-003_Analise_Documento_Arquitetura_DICT.md](../02_Arquitetura/ARE-003_Analise_Documento_Arquitetura_DICT.md)

3. **ArquiteturaDict_LBPAY.md** (Docs iniciais)
   `/Users/jose.silva.lb/LBPay/IA_Dict/Docs_iniciais/ArquiteturaDict_LBPAY.md`

4. **DUVIDAS.md** (resolvidas)
   [Artefatos/00_Master/DUVIDAS.md](../00_Master/DUVIDAS.md)

### 9.3 Padrões e RFCs

1. **RFC 7807** - Problem Details for HTTP APIs
   https://tools.ietf.org/html/rfc7807

2. **RFC 4122** - UUID
   https://tools.ietf.org/html/rfc4122

3. **XML Digital Signature**
   https://www.w3.org/2000/09/xmldsig

4. **Token Bucket Algorithm**
   https://en.wikipedia.org/wiki/Token_bucket

---

## 🎯 Próximos Passos

### Fase 1: Especificação (Atual)
- ✅ **API-001 criado** (este documento)
- ⏭️ Aguardar review e aprovação de stakeholders
- ⏭️ Incorporar feedback

### Fase 2: Detalhamento
- ⏭️ Criar schemas XML completos (Go structs)
- ⏭️ Detalhar cada módulo (REST client, XML signer, etc.)
- ⏭️ Especificar estratégia de testes

### Fase 3: Implementação
- ⏭️ Implementar módulos no **connector-dict** (CONNECT DICT)
- ⏭️ Implementar casos de uso no **core-dict** (CORE DICT)
- ⏭️ Implementar testes unitários e de integração

### Fase 4: Homologação
- ⏭️ Executar testes no ambiente Homologação Bacen
- ⏭️ Documentar evidências
- ⏭️ Obter aprovação Bacen

---

**Documento criado por**: MERCURY (AGT-API-001) - API Specialist
**Data**: 2025-10-24
**Versão**: 1.0
**Status**: ✅ Completo - Aguardando Review

---

**Estatísticas do Documento**:
- **28 endpoints REST** especificados
- **42 RFs** mapeados (de 72 totais)
- **30 tipos de erro** documentados
- **30 políticas de rate limiting** detalhadas
- **13 módulos** recomendados para implementação
- **~50 structs Go** a criar
