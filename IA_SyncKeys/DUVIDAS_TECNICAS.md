# Documento de Dúvidas Técnicas - Projeto DICT CID/VSync Sync

**Data**: 2025-10-28
**Status**: 🟡 AGUARDANDO ESCLARECIMENTOS
**Responsável**: Tech Lead

---

## ✅ Confirmações

Todos os 4 tópicos fornecidos estão **100% alinhados** com o documento RF_Dict_Bacen.md (Capítulo 9 do Manual DICT BACEN):

1. ✅ Geração de CIDs ao criar/alterar chaves no connector-dict
2. ✅ Armazenamento em tabela PostgreSQL local
3. ✅ VSync periódico via Bridge para DICT BACEN
4. ✅ Reconciliação e atualização do Core-Dict em caso de dessincronização

---

## ❓ Dúvidas Técnicas (Requerem Esclarecimento)

### 1. Eventos do Connector-Dict

**Dúvida**: O connector-dict já emite eventos Pulsar quando cria/altera chaves com sucesso?

**Contexto**: O documento menciona:
> "O connect-dict gera um evento sempre que criar uma chave dict com sucesso ou alterar uma chave dict com sucesso."

**Informações Necessárias**:
- [ ] Topic Pulsar atual para criação de chaves: `_____________`
- [ ] Topic Pulsar atual para alteração de chaves: `_____________`
- [ ] Schema dos eventos atuais (JSON ou Avro): `_____________`
- [ ] Estes eventos já contêm todos os dados necessários para gerar o CID?

**Ação Requerida**:
- Verificar código do connector-dict para identificar pontos de emissão de eventos
- Analisar se precisamos criar novos eventos ou apenas consumir existentes

---

### 2. Schema da Tabela PostgreSQL de CIDs

**Dúvida**: Qual deve ser o schema exato da tabela de CIDs?

**Contexto**: O documento técnico não especifica o schema completo da tabela.

**Proposta Inicial** (sujeita a validação):

```sql
CREATE TABLE dict_cids (
    id                  BIGSERIAL PRIMARY KEY,
    cid                 VARCHAR(64) NOT NULL UNIQUE,          -- SHA-256 hex (64 chars)
    key_type            VARCHAR(10) NOT NULL,                  -- CPF|CNPJ|PHONE|EMAIL|EVP
    key_value           VARCHAR(255) NOT NULL,                 -- Valor da chave

    -- Dados completos para regenerar CID (conforme algoritmo)
    ispb                VARCHAR(8) NOT NULL,
    branch              VARCHAR(10),
    account_number      VARCHAR(20) NOT NULL,
    account_type        VARCHAR(4) NOT NULL,                   -- CACC|SVGS|TRAN
    account_opened_at   TIMESTAMP NOT NULL,

    owner_type          VARCHAR(2) NOT NULL,                   -- PF|PJ
    owner_tax_id        VARCHAR(14) NOT NULL,                  -- CPF/CNPJ
    owner_name          VARCHAR(255) NOT NULL,
    owner_trade_name    VARCHAR(255),

    registered_at       TIMESTAMP NOT NULL,
    participant_registered_at TIMESTAMP NOT NULL,
    request_id          UUID NOT NULL,

    algorithm_version   VARCHAR(10) NOT NULL DEFAULT '1.0',

    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP NOT NULL DEFAULT NOW(),

    INDEX idx_key_type (key_type),
    INDEX idx_key_value (key_value),
    INDEX idx_owner_tax_id (owner_tax_id),
    INDEX idx_created_at (created_at)
);

-- Tabela para VSyncs calculados
CREATE TABLE dict_vsyncs (
    id                  SERIAL PRIMARY KEY,
    key_type            VARCHAR(10) NOT NULL UNIQUE,
    vsync_value         VARCHAR(64) NOT NULL,                  -- XOR cumulativo (hex)
    total_keys          BIGINT NOT NULL DEFAULT 0,
    last_calculated_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    last_verified_at    TIMESTAMP,
    synchronized        BOOLEAN NOT NULL DEFAULT TRUE,

    INDEX idx_key_type (key_type)
);

-- Tabela de auditoria de verificações
CREATE TABLE dict_sync_verifications (
    id                  BIGSERIAL PRIMARY KEY,
    key_type            VARCHAR(10) NOT NULL,
    vsync_local         VARCHAR(64) NOT NULL,
    vsync_dict          VARCHAR(64),
    synchronized        BOOLEAN NOT NULL,
    total_keys_local    BIGINT NOT NULL,
    verified_at         TIMESTAMP NOT NULL DEFAULT NOW(),

    INDEX idx_key_type (key_type),
    INDEX idx_verified_at (verified_at)
);
```

**Questões**:
- [ ] Este schema atende todas as necessidades?
- [ ] Falta algum campo importante?
- [ ] Precisamos de particionamento por key_type?
- [ ] Qual a volumetria esperada (para definir estratégia de índices)?

---

### 3. Integração com Core-Dict

**Dúvida**: Como exatamente o Core-Dict deve ser informado em caso de dessincronização?

**Contexto**: O documento menciona:
> "Teremos que informar o Core_Dict, fornecendo o arquivo que teremos que solicitar ao Dict-Bacen"

**Informações Necessárias**:
- [ ] O Core-Dict possui API REST para receber arquivo de CIDs?
- [ ] Endpoint: `_____________`
- [ ] Formato esperado: CSV | JSON | XML | Outro `_____________`
- [ ] O Core-Dict processa sincronamente ou asyncronamente?
- [ ] Existe callback de confirmação?

**Alternativas a Considerar**:
1. **API REST**: POST /api/v1/sync/cids-reconciliation (arquivo no body)
2. **Evento Pulsar**: Publicar evento com referência ao arquivo
3. **S3 + Notificação**: Upload do arquivo no S3 e notificar Core-Dict
4. **gRPC Stream**: Stream de CIDs via gRPC

**Ação Requerida**:
- Verificar documentação do Core-Dict
- Confirmar método preferencial de integração

---

### 4. Periodicidade da Verificação VSync

**Dúvida**: Qual deve ser a frequência das verificações de sincronismo?

**Contexto**: O documento técnico sugere verificação periódica mas não especifica.

**Opções**:
- [ ] A cada 1 hora (verificação frequente)
- [ ] A cada 6 horas (balanceado)
- [ ] A cada 24 horas (diária, horário específico)
- [ ] Configurable via variável de ambiente

**Considerações**:
- Manual BACEN não especifica periodicidade mínima/máxima
- Cada verificação consome recursos (API DICT via Bridge)
- Quanto mais frequente, menor janela de dessincronização
- Precisamos balancear entre custo e detecção rápida

**Proposta**:
- **Produção**: 1x por dia (03:00 AM) - janela de baixo movimento
- **Configurable**: Via variável `VSYNC_CHECK_CRON` (ex: `0 3 * * *`)

---

### 5. Temporal Worker ou Serviço Standalone?

**Dúvida**: O serviço de VSync deve ser implementado como?

**Contexto**: O documento menciona:
> "Idealmente um atividade + worker temporal"

**Opções**:

#### Opção A: Temporal Workflow + Activity
```
✅ Vantagens:
- Retry automático em caso de falhas
- Histórico completo de execuções
- Integração com stack existente
- Observability built-in

❌ Desvantagens:
- Complexidade adicional
- Dependência do Temporal
```

#### Opção B: Cronjob Kubernetes + Go Service
```
✅ Vantagens:
- Mais simples e direto
- Independente do Temporal
- Fácil de testar localmente

❌ Desvantagens:
- Retry manual
- Menos observability
- Sem histórico de execuções
```

**Recomendação**: **Opção A (Temporal)** devido a:
- Projeto já usa Temporal extensivamente
- Necessidade de retry robusto (API BACEN pode falhar)
- Observability é crítica para operações regulatórias
- Temporal guarda histórico completo (auditoria)

---

### 6. Tratamento de Dessincronização

**Dúvida**: O que fazer quando detectamos dessincronização?

**Contexto**: Seção 9.3 descreve reconciliação mas não especifica SLA.

**Cenário**: VSync local ≠ VSync DICT

**Fluxo Proposto**:
1. ⚠️ Alertar equipe (PagerDuty / Slack)
2. 📋 Solicitar lista completa de CIDs ao DICT (Seção 9.3)
3. 🔍 Identificar divergências (local vs DICT)
4. 🔄 Iniciar reconciliação automática ou manual?

**Questões**:
- [ ] Reconciliação automática ou requer aprovação manual?
- [ ] Há limite de divergências para reconciliação automática? (ex: máx 100 CIDs)
- [ ] Se >1000 divergências, processo manual obrigatório?
- [ ] Durante reconciliação, bloquear criação/alteração de chaves?

---

### 7. Variáveis de Ambiente

**Dúvida**: Quais são as credenciais PostgreSQL e configurações necessárias?

**Informações Necessárias**:

```bash
# PostgreSQL
POSTGRES_HOST=_____________
POSTGRES_PORT=_____________
POSTGRES_DB=_____________
POSTGRES_USER=_____________
POSTGRES_PASSWORD=_____________
POSTGRES_SSL_MODE=_____________

# Bridge gRPC
BRIDGE_GRPC_HOST=_____________
BRIDGE_GRPC_PORT=_____________
BRIDGE_MTLS_CERT_PATH=_____________
BRIDGE_MTLS_KEY_PATH=_____________
BRIDGE_MTLS_CA_PATH=_____________

# VSync Scheduler
VSYNC_CHECK_CRON="0 3 * * *"  # 03:00 AM diário
VSYNC_TIMEOUT_SECONDS=300      # 5 minutos

# ISPB Participante
ISPB_PARTICIPANTE=_____________

# Pulsar (eventos de chaves)
PULSAR_KEY_CREATED_TOPIC=_____________
PULSAR_KEY_UPDATED_TOPIC=_____________

# Core-Dict Integration
CORE_DICT_API_URL=_____________
CORE_DICT_API_TOKEN=_____________
```

---

### 8. Normalização de Dados

**Dúvida**: O connector-dict já normaliza dados conforme algoritmo CID?

**Contexto**: O algoritmo de CID requer normalização estrita:
- CPF: sem máscara, 11 dígitos
- CNPJ: sem máscara, 14 dígitos
- Telefone: formato E.164 (+5511999999999)
- Email: lowercase
- EVP: lowercase sem hífens

**Questões**:
- [ ] O connector-dict já armazena dados normalizados?
- [ ] Precisamos normalizar antes de calcular CID?
- [ ] Onde aplicar normalização: no consumer do evento ou no CID generator?

---

### 9. Testes com DICT Sandbox

**Dúvida**: Existe ambiente sandbox do DICT BACEN para testes?

**Informações Necessárias**:
- [ ] URL do DICT Sandbox: `_____________`
- [ ] Credenciais de teste: `_____________`
- [ ] Limitações do sandbox (rate limits, features disponíveis)
- [ ] Dados de teste disponíveis

---

### 10. Rollback e Disaster Recovery

**Dúvida**: Como proceder se reconciliação corromper dados?

**Cenários**:
- Reconciliação remove CIDs que existem
- Reconciliação adiciona CIDs inexistentes
- Falha durante aplicação de 50% das alterações

**Questões**:
- [ ] Backup automático antes de reconciliação?
- [ ] Procedimento de rollback documentado?
- [ ] Quem pode autorizar rollback?

---

## 📋 Próximos Passos

### Antes de Iniciar Implementação

1. **Responder todas as 10 dúvidas acima**
2. **Validar schema PostgreSQL** com DBA
3. **Confirmar integração Core-Dict** com time Core-Dict
4. **Validar acesso Bridge** com time Infrastructure
5. **Definir SLAs** de sincronização e reconciliação
6. **Aprovar documento de specs técnicas**

### Após Esclarecimentos

1. Atualizar `Specs.md` com respostas
2. Criar `Claude.md` com squad e roadmap
3. Iniciar implementação

---

## 📞 Stakeholders a Consultar

| Dúvida | Stakeholder | Contato |
|--------|-------------|---------|
| 1, 8 | Time Connector-Dict | ___________ |
| 2, 9 | DBA / Data Team | ___________ |
| 3 | Time Core-Dict | ___________ |
| 4, 6, 10 | Product Owner / Tech Lead | ___________ |
| 5, 7 | DevOps / Infrastructure | ___________ |

---

**Status**: 🟡 Aguardando respostas antes de prosseguir com implementação
**Última atualização**: 2025-10-28
