# DAT-004: Data Dictionary

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Versão**: 1.0
**Data**: 2025-10-25
**Status**: ✅ Especificação Completa
**Responsável**: ARCHITECT + Data Team

---

## 📋 Resumo Executivo

Este documento fornece o **dicionário de dados completo** do sistema DICT, descrevendo todas as tabelas, colunas, tipos de dados, constraints, relacionamentos e regras de negócio associadas aos dados.

**Objetivo**: Servir como referência única (single source of truth) para a estrutura de dados do DICT, facilitando desenvolvimento, manutenção, e onboarding de novos membros da equipe.

---

## 🗄️ Database: Core DICT (PostgreSQL)

### Schema: `dict`

---

## 📊 Tabela: `dict.entries`

**Descrição**: Armazena chaves DICT (PIX) vinculadas a contas bancárias

| Coluna | Tipo | Nullable | Default | Descrição | Constraints |
|--------|------|----------|---------|-----------|-------------|
| `id` | UUID | NOT NULL | `gen_random_uuid()` | Identificador único da entry | PRIMARY KEY |
| `key_type` | VARCHAR(20) | NOT NULL | - | Tipo de chave DICT | CHECK IN ('CPF', 'CNPJ', 'PHONE', 'EMAIL', 'EVP') |
| `key_value` | VARCHAR(255) | NOT NULL | - | Valor da chave (CPF, telefone, etc.) | UNIQUE (key_type, key_value) |
| `account_ispb` | CHAR(8) | NOT NULL | - | ISPB da instituição financeira | CHECK (length = 8) |
| `account_type` | VARCHAR(20) | NOT NULL | - | Tipo de conta | CHECK IN ('CHECKING', 'SAVINGS', 'PAYMENT', 'SALARY') |
| `account_number` | VARCHAR(20) | NOT NULL | - | Número da conta | - |
| `account_check_digit` | VARCHAR(2) | NULL | - | Dígito verificador | - |
| `branch_code` | VARCHAR(4) | NOT NULL | - | Código da agência | - |
| `account_holder_name` | VARCHAR(255) | NOT NULL | - | Nome do titular | - |
| `account_holder_document` | VARCHAR(14) | NOT NULL | - | CPF/CNPJ do titular | CHECK (validação de formato) |
| `status` | VARCHAR(50) | NOT NULL | 'ACTIVE' | Status da entry | CHECK IN ('ACTIVE', 'PORTABILITY_PENDING', 'PORTABILITY_CONFIRMED', 'CLAIM_PENDING', 'DELETED') |
| `external_id` | VARCHAR(255) | NULL | - | ID da entry no Bacen DICT | - |
| `created_at` | TIMESTAMP WITH TIME ZONE | NOT NULL | `NOW()` | Data/hora de criação | - |
| `updated_at` | TIMESTAMP WITH TIME ZONE | NOT NULL | `NOW()` | Data/hora última atualização | - |
| `deleted_at` | TIMESTAMP WITH TIME ZONE | NULL | - | Data/hora de soft delete | - |
| `created_by` | UUID | NULL | - | Usuário que criou | FK → users(id) |
| `updated_by` | UUID | NULL | - | Usuário que atualizou | FK → users(id) |

**Indexes**:
- `idx_entries_key_type_value` - UNIQUE (key_type, key_value) WHERE deleted_at IS NULL
- `idx_entries_account` - (account_ispb, account_number, branch_code)
- `idx_entries_status` - (status) WHERE status != 'DELETED'
- `idx_entries_created_at` - (created_at DESC)
- `idx_entries_external_id` - (external_id) WHERE external_id IS NOT NULL

**Triggers**:
- `trg_entries_updated_at` - Atualiza `updated_at` automaticamente
- `trg_entries_audit` - Grava evento em `audit.entry_events`

**Regras de Negócio**:
- `key_type` + `key_value` devem ser únicos (soft delete)
- CPF/CNPJ deve ser validado (algoritmo de dígito verificador)
- Telefone formato: +55 (11 dígitos)
- Email formato: RFC 5322
- EVP: UUID v4 gerado automaticamente

---

## 📊 Tabela: `dict.accounts`

**Descrição**: Contas bancárias dos clientes (usuários do sistema)

| Coluna | Tipo | Nullable | Default | Descrição |
|--------|------|----------|---------|-----------|
| `id` | UUID | NOT NULL | `gen_random_uuid()` | PRIMARY KEY |
| `user_id` | UUID | NOT NULL | - | FK → users(id) |
| `ispb` | CHAR(8) | NOT NULL | - | ISPB da instituição |
| `account_type` | VARCHAR(20) | NOT NULL | - | Tipo de conta |
| `account_number` | VARCHAR(20) | NOT NULL | - | Número da conta |
| `account_check_digit` | VARCHAR(2) | NULL | - | Dígito verificador |
| `branch_code` | VARCHAR(4) | NOT NULL | - | Código da agência |
| `is_default` | BOOLEAN | NOT NULL | FALSE | Conta padrão do usuário |
| `status` | VARCHAR(20) | NOT NULL | 'ACTIVE' | Status ('ACTIVE', 'INACTIVE', 'BLOCKED') |
| `created_at` | TIMESTAMP WITH TIME ZONE | NOT NULL | `NOW()` | Data de criação |
| `updated_at` | TIMESTAMP WITH TIME ZONE | NOT NULL | `NOW()` | Última atualização |

**Indexes**:
- `idx_accounts_user_id` - (user_id)
- `idx_accounts_ispb_account` - UNIQUE (ispb, account_number, branch_code, user_id)

**Constraints**:
- UNIQUE (user_id, is_default) WHERE is_default = TRUE (apenas 1 conta default por usuário)

---

## 📊 Tabela: `dict.claims`

**Descrição**: Reivindicações de chaves DICT (período de 30 dias)

| Coluna | Tipo | Nullable | Default | Descrição |
|--------|------|----------|---------|-----------|
| `id` | UUID | NOT NULL | `gen_random_uuid()` | PRIMARY KEY |
| `workflow_id` | VARCHAR(255) | NULL | - | Temporal Workflow ID |
| `entry_id` | UUID | NOT NULL | - | FK → entries(id) |
| `claimer_ispb` | CHAR(8) | NOT NULL | - | ISPB do reivindicador |
| `claimer_account_number` | VARCHAR(20) | NOT NULL | - | Conta do reivindicador |
| `claimer_branch_code` | VARCHAR(4) | NOT NULL | - | Agência do reivindicador |
| `owner_ispb` | CHAR(8) | NOT NULL | - | ISPB do dono atual |
| `completion_period_days` | INT | NOT NULL | 30 | Período de conclusão (sempre 30) |
| `status` | VARCHAR(50) | NOT NULL | 'OPEN' | Status da claim |
| `expires_at` | TIMESTAMP WITH TIME ZONE | NOT NULL | - | Data de expiração (created_at + 30 dias) |
| `completed_at` | TIMESTAMP WITH TIME ZONE | NULL | - | Data de conclusão |
| `external_id` | VARCHAR(255) | NULL | - | ID da claim no Bacen |
| `created_at` | TIMESTAMP WITH TIME ZONE | NOT NULL | `NOW()` | Data de criação |
| `updated_at` | TIMESTAMP WITH TIME ZONE | NOT NULL | `NOW()` | Última atualização |
| `created_by` | UUID | NULL | - | FK → users(id) |

**Indexes**:
- `idx_claims_entry_id` - (entry_id)
- `idx_claims_status` - (status)
- `idx_claims_expires_at` - (expires_at) WHERE status IN ('OPEN', 'WAITING_RESOLUTION')
- `idx_claims_workflow_id` - (workflow_id) WHERE workflow_id IS NOT NULL

**Constraints**:
- CHECK (completion_period_days = 30)
- CHECK (expires_at = created_at + INTERVAL '30 days')
- CHECK (status IN ('OPEN', 'WAITING_RESOLUTION', 'CONFIRMED', 'CANCELLED', 'COMPLETED', 'EXPIRED'))
- UNIQUE (entry_id) WHERE status IN ('OPEN', 'WAITING_RESOLUTION') (apenas 1 claim ativa por entry)

**Triggers**:
- `trg_claims_set_expires_at` - Define expires_at automaticamente (created_at + 30 dias)
- `trg_claims_audit` - Grava evento em audit log

**Regras de Negócio**:
- Período de conclusão é SEMPRE 30 dias (TEC-003 v2.1)
- Apenas 1 claim ativa por entry
- Após 30 dias sem resposta, status muda para 'EXPIRED' automaticamente (Temporal Workflow)

---

## 📊 Tabela: `dict.portabilities`

**Descrição**: Histórico de portabilidades de conta

| Coluna | Tipo | Nullable | Default | Descrição |
|--------|------|----------|---------|-----------|
| `id` | UUID | NOT NULL | `gen_random_uuid()` | PRIMARY KEY |
| `entry_id` | UUID | NOT NULL | - | FK → entries(id) |
| `old_account_ispb` | CHAR(8) | NOT NULL | - | ISPB da conta antiga |
| `old_account_number` | VARCHAR(20) | NOT NULL | - | Número da conta antiga |
| `old_branch_code` | VARCHAR(4) | NOT NULL | - | Agência antiga |
| `new_account_ispb` | CHAR(8) | NOT NULL | - | ISPB da nova conta |
| `new_account_number` | VARCHAR(20) | NOT NULL | - | Número da nova conta |
| `new_branch_code` | VARCHAR(4) | NOT NULL | - | Agência nova |
| `status` | VARCHAR(20) | NOT NULL | 'PENDING' | Status ('PENDING', 'CONFIRMED', 'CANCELLED') |
| `confirmed_at` | TIMESTAMP WITH TIME ZONE | NULL | - | Data de confirmação |
| `external_id` | VARCHAR(255) | NULL | - | ID da portabilidade no Bacen |
| `created_at` | TIMESTAMP WITH TIME ZONE | NOT NULL | `NOW()` | Data de criação |

**Indexes**:
- `idx_portabilities_entry_id` - (entry_id)
- `idx_portabilities_status` - (status)

---

## 📊 Tabela: `dict.users`

**Descrição**: Usuários do sistema DICT (titulares de contas)

| Coluna | Tipo | Nullable | Default | Descrição |
|--------|------|----------|---------|-----------|
| `id` | UUID | NOT NULL | `gen_random_uuid()` | PRIMARY KEY |
| `email` | VARCHAR(255) | NOT NULL | - | Email (login) |
| `name` | VARCHAR(255) | NOT NULL | - | Nome completo |
| `document_type` | VARCHAR(10) | NOT NULL | - | 'CPF' ou 'CNPJ' |
| `document_number` | VARCHAR(14) | NOT NULL | - | CPF/CNPJ (sem formatação) |
| `phone` | VARCHAR(20) | NULL | - | Telefone (formato +5511999999999) |
| `status` | VARCHAR(20) | NOT NULL | 'ACTIVE' | 'ACTIVE', 'INACTIVE', 'BLOCKED' |
| `email_verified` | BOOLEAN | NOT NULL | FALSE | Email verificado |
| `phone_verified` | BOOLEAN | NOT NULL | FALSE | Telefone verificado |
| `created_at` | TIMESTAMP WITH TIME ZONE | NOT NULL | `NOW()` | Data de criação |
| `updated_at` | TIMESTAMP WITH TIME ZONE | NOT NULL | `NOW()` | Última atualização |
| `last_login_at` | TIMESTAMP WITH TIME ZONE | NULL | - | Último login |

**Indexes**:
- `idx_users_email` - UNIQUE (email)
- `idx_users_document` - UNIQUE (document_type, document_number)

**Segurança**:
- Senhas NÃO são armazenadas nesta tabela (autenticação via OAuth/JWT)
- Document_number: indexado mas masked em logs (LGPD)

---

## 📊 Tabela: `audit.entry_events`

**Descrição**: Log de auditoria de operações em entries (LGPD + Bacen compliance)

| Coluna | Tipo | Nullable | Default | Descrição |
|--------|------|----------|---------|-----------|
| `id` | BIGSERIAL | NOT NULL | - | PRIMARY KEY |
| `event_type` | VARCHAR(50) | NOT NULL | - | 'CREATE', 'UPDATE', 'DELETE', 'ACCESS' |
| `entry_id` | UUID | NOT NULL | - | FK → entries(id) |
| `user_id` | UUID | NULL | - | FK → users(id) |
| `user_ip` | INET | NULL | - | Endereço IP do usuário |
| `operation` | VARCHAR(100) | NULL | - | Nome da operação (CreateEntry, UpdateEntry) |
| `old_value` | JSONB | NULL | - | Valor antigo (para UPDATE) |
| `new_value` | JSONB | NULL | - | Valor novo |
| `timestamp` | TIMESTAMP WITH TIME ZONE | NOT NULL | `NOW()` | Timestamp do evento |
| `request_id` | VARCHAR(255) | NULL | - | Request ID (correlação) |

**Indexes**:
- `idx_audit_entry_id` - (entry_id)
- `idx_audit_timestamp` - (timestamp DESC)
- `idx_audit_user_id` - (user_id)
- `idx_audit_event_type` - (event_type)

**Particionamento**:
- Particionar por mês (performance em histórico longo)

**Retenção**:
- 5 anos (exigência Bacen)

---

## 🗄️ Database: Connect (PostgreSQL)

### Schema: `connect`

---

## 📊 Tabela: `connect.claim_workflows`

**Descrição**: Metadata de workflows Temporal para claims

| Coluna | Tipo | Nullable | Default | Descrição |
|--------|------|----------|---------|-----------|
| `id` | UUID | NOT NULL | `gen_random_uuid()` | PRIMARY KEY |
| `workflow_id` | VARCHAR(255) | NOT NULL | - | Temporal Workflow ID |
| `workflow_type` | VARCHAR(100) | NOT NULL | 'ClaimWorkflow' | Tipo de workflow |
| `claim_id` | UUID | NOT NULL | - | FK → dict.claims(id) |
| `status` | VARCHAR(50) | NOT NULL | 'RUNNING' | 'RUNNING', 'COMPLETED', 'FAILED', 'TIMED_OUT' |
| `started_at` | TIMESTAMP WITH TIME ZONE | NOT NULL | `NOW()` | Início do workflow |
| `completed_at` | TIMESTAMP WITH TIME ZONE | NULL | - | Conclusão do workflow |
| `current_activity` | VARCHAR(255) | NULL | - | Activity atual em execução |
| `error_message` | TEXT | NULL | - | Mensagem de erro (se failed) |

**Indexes**:
- `idx_claim_workflows_workflow_id` - UNIQUE (workflow_id)
- `idx_claim_workflows_claim_id` - (claim_id)
- `idx_claim_workflows_status` - (status)

---

## 📊 Enums e Tipos Customizados

### KeyType (Enum)
```sql
CREATE TYPE key_type AS ENUM (
    'CPF',
    'CNPJ',
    'PHONE',
    'EMAIL',
    'EVP'
);
```

### EntryStatus (Enum)
```sql
CREATE TYPE entry_status AS ENUM (
    'ACTIVE',
    'PORTABILITY_PENDING',
    'PORTABILITY_CONFIRMED',
    'CLAIM_PENDING',
    'DELETED'
);
```

### ClaimStatus (Enum)
```sql
CREATE TYPE claim_status AS ENUM (
    'OPEN',
    'WAITING_RESOLUTION',
    'CONFIRMED',
    'CANCELLED',
    'COMPLETED',
    'EXPIRED'
);
```

### AccountType (Enum)
```sql
CREATE TYPE account_type AS ENUM (
    'CHECKING',   -- Conta Corrente
    'SAVINGS',    -- Poupança
    'PAYMENT',    -- Pagamento
    'SALARY'      -- Salário
);
```

---

## 🔗 Relacionamentos (ER Diagram)

```
users (1) ──────< (N) accounts
  │
  │
  └──────< (N) entries
              │
              │
              ├──────< (N) claims
              │
              └──────< (N) portabilities

entries (1) ──────< (N) audit.entry_events
```

---

## 📏 Validações de Dados

### CPF Validation (Pseudocódigo)
```sql
-- Função PostgreSQL para validar CPF
CREATE OR REPLACE FUNCTION is_valid_cpf(cpf VARCHAR(11)) RETURNS BOOLEAN AS $$
DECLARE
    sum1 INT := 0;
    sum2 INT := 0;
    digit1 INT;
    digit2 INT;
BEGIN
    -- Validações básicas
    IF LENGTH(cpf) != 11 THEN
        RETURN FALSE;
    END IF;

    -- CPFs inválidos conhecidos (111.111.111-11, etc.)
    IF cpf IN ('00000000000', '11111111111', '22222222222', '33333333333',
               '44444444444', '55555555555', '66666666666', '77777777777',
               '88888888888', '99999999999') THEN
        RETURN FALSE;
    END IF;

    -- Calcular primeiro dígito verificador
    FOR i IN 1..9 LOOP
        sum1 := sum1 + SUBSTRING(cpf, i, 1)::INT * (11 - i);
    END LOOP;

    digit1 := 11 - (sum1 % 11);
    IF digit1 >= 10 THEN
        digit1 := 0;
    END IF;

    -- Calcular segundo dígito verificador
    FOR i IN 1..10 LOOP
        sum2 := sum2 + SUBSTRING(cpf, i, 1)::INT * (12 - i);
    END LOOP;

    digit2 := 11 - (sum2 % 11);
    IF digit2 >= 10 THEN
        digit2 := 0;
    END IF;

    -- Validar
    RETURN SUBSTRING(cpf, 10, 1)::INT = digit1
       AND SUBSTRING(cpf, 11, 1)::INT = digit2;
END;
$$ LANGUAGE plpgsql;

-- Constraint
ALTER TABLE dict.entries ADD CONSTRAINT chk_cpf_valid
    CHECK (key_type != 'CPF' OR is_valid_cpf(key_value));
```

### Telefone Validation
```sql
-- Formato: +5511999999999 (13 caracteres)
ALTER TABLE dict.entries ADD CONSTRAINT chk_phone_format
    CHECK (
        key_type != 'PHONE'
        OR (key_value ~ '^\+55[1-9]{2}9[0-9]{8}$')
    );
```

### Email Validation
```sql
-- RFC 5322 simplificado
ALTER TABLE dict.entries ADD CONSTRAINT chk_email_format
    CHECK (
        key_type != 'EMAIL'
        OR (key_value ~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$')
    );
```

---

## 📊 Tamanhos Estimados

### Estimativa de Crescimento (5 anos)

| Tabela | Registros/Ano | Total 5 Anos | Tamanho/Registro | Tamanho Total |
|--------|---------------|--------------|------------------|---------------|
| `entries` | 1.000.000 | 5.000.000 | 1 KB | ~5 GB |
| `claims` | 100.000 | 500.000 | 0.5 KB | ~250 MB |
| `portabilities` | 50.000 | 250.000 | 0.3 KB | ~75 MB |
| `users` | 500.000 | 2.500.000 | 0.5 KB | ~1.25 GB |
| `audit.entry_events` | 5.000.000 | 25.000.000 | 1 KB | ~25 GB |
| **TOTAL** | - | - | - | **~32 GB** |

**Provisionamento Recomendado**: 100 GB (margem de 3x)

---

## 📚 Referências

### Documentos Internos
- [DAT-001: Schema Database Core DICT](DAT-001_Schema_Database_Core_DICT.md)
- [DAT-002: Schema Database Connect](DAT-002_Schema_Database_Connect.md)
- [DAT-003: Migrations Strategy](DAT-003_Migrations_Strategy.md)
- [SEC-007: LGPD Data Protection](../../13_Seguranca/SEC-007_LGPD_Data_Protection.md)

### Documentação Externa
- [PostgreSQL Data Types](https://www.postgresql.org/docs/16/datatype.html)
- [PostgreSQL Constraints](https://www.postgresql.org/docs/16/ddl-constraints.html)
- [PostgreSQL Indexes](https://www.postgresql.org/docs/16/indexes.html)

---

**Versão**: 1.0
**Status**: ✅ Especificação Completa
**Próxima Revisão**: Após implementação do schema

---

**IMPORTANTE**: Este dicionário de dados é baseado nos schemas especificados em DAT-001 e DAT-002. Qualquer mudança no schema deve ser refletida neste documento.
