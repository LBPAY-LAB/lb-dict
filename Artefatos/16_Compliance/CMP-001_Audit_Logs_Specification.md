# CMP-001: Audit Logs Specification

**Projeto**: DICT - Diretorio de Identificadores de Contas Transacionais (LBPay)
**Versao**: 1.0
**Data**: 2025-10-25
**Status**: Especificacao Completa
**Responsavel**: Security Team + Compliance + DPO

---

## 1. Resumo Executivo

Este documento especifica **todos os requisitos de auditoria e logging** para o sistema DICT da LBPay, em conformidade com:
- **Bacen**: Retencao de logs de auditoria por **5 anos** (Manual Operacional DICT, Requisito REG-171 a REG-180)
- **LGPD**: Rastreabilidade de operacoes sobre dados pessoais (Art. 37, Lei 13.709/2018)
- **ISO 27001**: Controles de seguranca da informacao

---

## 2. Objetivos dos Audit Logs

### 2.1 Conformidade Regulatoria

- Atender exigencias do Banco Central (retencao 5 anos)
- Demonstrar conformidade com LGPD (prestacao de contas)
- Suportar auditorias externas (Bacen, ANPD, auditorias independentes)

### 2.2 Seguranca e Deteccao de Fraudes

- Detectar acessos nao autorizados
- Identificar padroes de comportamento suspeitos
- Rastrear tentativas de violacao de seguranca
- Suportar investigacoes de incidentes

### 2.3 Rastreabilidade Operacional

- Rastrear todas as operacoes DICT (criacao, alteracao, exclusao de chaves)
- Correlacionar eventos entre microservicos (Core, Bridge, Connect)
- Debug e troubleshooting de problemas em producao

---

## 3. Categorias de Eventos Auditaveis

### 3.1 Operacoes DICT (Criticidade: ALTA)

Todas as operacoes do DICT DEVEM ser auditadas:

| Operacao | Event Type | Dados a Logar |
|----------|------------|---------------|
| **Cadastro de Chave** | `ENTRY_CREATED` | entry_id, key_type, key_value (masked), account, user_id, request_id |
| **Alteracao de Chave** | `ENTRY_UPDATED` | entry_id, old_value, new_value, user_id, request_id |
| **Exclusao de Chave** | `ENTRY_DELETED` | entry_id, key_value (masked), deletion_reason, user_id, request_id |
| **Consulta de Chave** | `ENTRY_QUERIED` | entry_id OR key_value (masked), user_id, request_id, query_type |
| **Claim Criado** | `CLAIM_CREATED` | claim_id, entry_id, claim_type, claimer_ispb, donor_ispb, user_id |
| **Claim Confirmado** | `CLAIM_CONFIRMED` | claim_id, confirmed_by, timestamp |
| **Claim Completado** | `CLAIM_COMPLETED` | claim_id, completed_by, timestamp |
| **Claim Cancelado** | `CLAIM_CANCELLED` | claim_id, cancellation_reason, cancelled_by |
| **Portabilidade Iniciada** | `PORTABILITY_STARTED` | portability_id, entry_id, target_ispb, user_id |
| **Portabilidade Completada** | `PORTABILITY_COMPLETED` | portability_id, completion_timestamp |

### 3.2 Operacoes de Usuario (Criticidade: MEDIA)

| Operacao | Event Type | Dados a Logar |
|----------|------------|---------------|
| **Login bem-sucedido** | `USER_LOGIN_SUCCESS` | user_id, ip_address, user_agent, timestamp |
| **Login falhado** | `USER_LOGIN_FAILED` | username (nao user_id), ip_address, failure_reason, timestamp |
| **Logout** | `USER_LOGOUT` | user_id, session_id, timestamp |
| **Alteracao de Senha** | `USER_PASSWORD_CHANGED` | user_id, ip_address, timestamp |
| **MFA Ativado** | `USER_MFA_ENABLED` | user_id, mfa_method, timestamp |
| **MFA Desativado** | `USER_MFA_DISABLED` | user_id, disabled_by, timestamp |
| **Criacao de Usuario** | `USER_CREATED` | user_id, created_by, role, timestamp |
| **Alteracao de Permissoes** | `USER_PERMISSIONS_CHANGED` | user_id, old_permissions, new_permissions, changed_by |

### 3.3 Acesso a Dados Pessoais (Criticidade: ALTA - LGPD)

| Operacao | Event Type | Dados a Logar |
|----------|------------|---------------|
| **DSAR (Data Subject Access Request)** | `DSAR_REQUESTED` | user_id, request_id, requested_at, requester_ip |
| **Exportacao de Dados** | `DATA_EXPORT` | user_id, export_format, exported_by, timestamp |
| **Exclusao de Dados (LGPD)** | `DATA_DELETION_LGPD` | user_id, deletion_type, deleted_by, timestamp |
| **Correcao de Dados** | `DATA_CORRECTION` | user_id, field_corrected, old_value, new_value, corrected_by |
| **Acesso a Dados Sensiveis** | `SENSITIVE_DATA_ACCESS` | user_id, data_type, accessed_by, purpose, timestamp |

### 3.4 Eventos de Seguranca (Criticidade: CRITICA)

| Operacao | Event Type | Dados a Logar |
|----------|------------|---------------|
| **Tentativa de Acesso Nao Autorizado** | `UNAUTHORIZED_ACCESS_ATTEMPT` | user_id, resource, ip_address, timestamp |
| **Rate Limit Excedido** | `RATE_LIMIT_EXCEEDED` | user_id, endpoint, ip_address, timestamp, count |
| **Certificado mTLS Invalido** | `MTLS_CERT_INVALID` | cert_subject, cert_issuer, validation_error, timestamp |
| **Falha de Autenticacao RSFN** | `RSFN_AUTH_FAILED` | ispb, error_code, timestamp |
| **Deteccao de Anomalia** | `SECURITY_ANOMALY_DETECTED` | anomaly_type, user_id, ip_address, details, timestamp |

### 3.5 Eventos de Sistema (Criticidade: MEDIA)

| Operacao | Event Type | Dados a Logar |
|----------|------------|---------------|
| **Startup de Servico** | `SERVICE_STARTED` | service_name, version, node_id, timestamp |
| **Shutdown de Servico** | `SERVICE_STOPPED` | service_name, shutdown_reason, timestamp |
| **Erro de Integracao RSFN** | `RSFN_INTEGRATION_ERROR` | operation, error_code, error_message, request_id |
| **Erro de Banco de Dados** | `DATABASE_ERROR` | operation, error_code, error_message (sanitized), timestamp |
| **Workflow Falhado** | `WORKFLOW_FAILED` | workflow_id, workflow_type, failure_reason, timestamp |

---

## 4. Formato de Log: JSON Estruturado

### 4.1 Estrutura Padrao

Todos os logs de auditoria DEVEM seguir este formato JSON:

```json
{
  "version": "1.0",
  "timestamp": "2025-10-25T14:30:00.123Z",
  "event_type": "ENTRY_CREATED",
  "severity": "INFO",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
  "request_id": "req-123456",
  "trace_id": "trace-789012",
  "service": {
    "name": "core-dict",
    "version": "1.0.0",
    "instance_id": "core-dict-pod-1",
    "environment": "production"
  },
  "actor": {
    "user_id": "user-550e8400",
    "username": "joao.silva@lbpay.com.br",
    "role": "user",
    "ip_address": "192.168.1.100",
    "user_agent": "Mozilla/5.0 (compatible; LBPay Mobile App/2.3.0)"
  },
  "resource": {
    "type": "dict_entry",
    "id": "entry-123456",
    "owner_id": "user-550e8400"
  },
  "action": {
    "type": "CREATE",
    "status": "SUCCESS",
    "http_method": "POST",
    "endpoint": "/api/v1/entries",
    "http_status": 201
  },
  "data": {
    "key_type": "CPF",
    "key_value": "***8900",
    "account": {
      "ispb": "00000000",
      "branch": "0001",
      "account": "***45-6"
    }
  },
  "metadata": {
    "duration_ms": 120,
    "bacen_request_id": "bacen-req-789012",
    "workflow_id": "wf-RegisterKey-123456"
  }
}
```

### 4.2 Campos Obrigatorios

| Campo | Tipo | Descricao | Obrigatorio |
|-------|------|-----------|-------------|
| `version` | string | Versao do schema de log (ex: "1.0") | Sim |
| `timestamp` | ISO 8601 | Timestamp do evento (UTC) | Sim |
| `event_type` | string | Tipo de evento (enum) | Sim |
| `severity` | string | Nivel de severidade (DEBUG, INFO, WARN, ERROR, CRITICAL) | Sim |
| `correlation_id` | UUID | ID unico para rastrear evento end-to-end | Sim |
| `request_id` | string | ID da requisicao HTTP/gRPC | Sim (quando aplicavel) |
| `trace_id` | string | ID de trace distribuido (OpenTelemetry) | Sim |
| `service.name` | string | Nome do servico (core-dict, bridge, connect) | Sim |
| `service.version` | string | Versao do servico | Sim |
| `service.instance_id` | string | ID da instancia (pod ID) | Sim |
| `service.environment` | string | Ambiente (dev, staging, production) | Sim |
| `actor.user_id` | UUID | ID do usuario (se aplicavel) | Condicional |
| `actor.ip_address` | string | Endereco IP do ator | Sim |
| `resource.type` | string | Tipo de recurso (dict_entry, claim, user) | Sim |
| `resource.id` | string | ID do recurso | Sim |
| `action.type` | string | Tipo de acao (CREATE, READ, UPDATE, DELETE) | Sim |
| `action.status` | string | Status da acao (SUCCESS, FAILURE, PARTIAL) | Sim |
| `data` | object | Dados especificos do evento (com masking) | Nao |
| `metadata` | object | Metadados adicionais | Nao |

---

## 5. Masking de Dados Sensiveis

### 5.1 Regras de Masking

Dados pessoais DEVEM ser mascarados em logs conforme tabela:

| Tipo de Dado | Regra de Masking | Exemplo |
|--------------|------------------|---------|
| **CPF** | Mostrar apenas ultimos 4 digitos | `12345678900` → `***8900` |
| **CNPJ** | Mostrar apenas ultimos 4 digitos | `12345678000190` → `***0190` |
| **Email** | Mostrar 1º caractere + dominio | `joao.silva@example.com` → `j***@example.com` |
| **Telefone** | Mostrar apenas ultimos 4 digitos | `+5511987654321` → `***4321` |
| **Nome Completo** | Mostrar apenas primeiro nome | `Joao Silva Santos` → `Joao ***` |
| **Conta Bancaria** | Mostrar apenas ultimos 2 digitos | `123456-7` → `***56-7` |
| **Senha** | NUNCA logar | - |
| **Token** | NUNCA logar | - |

### 5.2 Implementacao (Pseudocodigo Go)

```go
package audit

import (
    "regexp"
    "strings"
)

func MaskCPF(cpf string) string {
    if len(cpf) != 11 {
        return "***"
    }
    return "***" + cpf[7:]
}

func MaskEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return "***"
    }
    return string(parts[0][0]) + "***@" + parts[1]
}

func MaskPhoneNumber(phone string) string {
    // E.164: +5511987654321
    if len(phone) < 4 {
        return "***"
    }
    return "***" + phone[len(phone)-4:]
}

func MaskAccountNumber(account string) string {
    // Ex: 123456-7
    re := regexp.MustCompile(`(\d+)-(\d)`)
    return re.ReplaceAllString(account, "***$2")
}
```

---

## 6. Armazenamento de Logs

### 6.1 PostgreSQL - Audit Schema

#### Tabela: `audit.entry_events`

```sql
CREATE SCHEMA IF NOT EXISTS audit;

CREATE TABLE audit.entry_events (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    correlation_id UUID NOT NULL,
    request_id VARCHAR(255),
    trace_id VARCHAR(255),

    -- Service
    service_name VARCHAR(50) NOT NULL,
    service_version VARCHAR(20) NOT NULL,
    service_instance_id VARCHAR(100) NOT NULL,
    service_environment VARCHAR(20) NOT NULL,

    -- Actor
    actor_user_id UUID,
    actor_username VARCHAR(255),
    actor_role VARCHAR(50),
    actor_ip_address INET NOT NULL,
    actor_user_agent TEXT,

    -- Resource
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    resource_owner_id UUID,

    -- Action
    action_type VARCHAR(20) NOT NULL,
    action_status VARCHAR(20) NOT NULL,
    action_http_method VARCHAR(10),
    action_endpoint VARCHAR(255),
    action_http_status SMALLINT,

    -- Data (JSON)
    data JSONB,

    -- Metadata (JSON)
    metadata JSONB,

    -- Indexes
    CONSTRAINT entry_events_event_type_check CHECK (event_type IN (
        'ENTRY_CREATED', 'ENTRY_UPDATED', 'ENTRY_DELETED', 'ENTRY_QUERIED',
        'CLAIM_CREATED', 'CLAIM_CONFIRMED', 'CLAIM_COMPLETED', 'CLAIM_CANCELLED',
        'PORTABILITY_STARTED', 'PORTABILITY_COMPLETED'
    )),
    CONSTRAINT entry_events_severity_check CHECK (severity IN ('DEBUG', 'INFO', 'WARN', 'ERROR', 'CRITICAL')),
    CONSTRAINT entry_events_action_type_check CHECK (action_type IN ('CREATE', 'READ', 'UPDATE', 'DELETE', 'EXECUTE'))
);

-- Indexes para performance de queries
CREATE INDEX idx_entry_events_timestamp ON audit.entry_events (timestamp DESC);
CREATE INDEX idx_entry_events_correlation_id ON audit.entry_events (correlation_id);
CREATE INDEX idx_entry_events_user_id ON audit.entry_events (actor_user_id);
CREATE INDEX idx_entry_events_resource_id ON audit.entry_events (resource_id);
CREATE INDEX idx_entry_events_event_type ON audit.entry_events (event_type);
CREATE INDEX idx_entry_events_ip_address ON audit.entry_events (actor_ip_address);

-- Particionamento por mes (para performance em retencao de 5 anos)
CREATE TABLE audit.entry_events_y2025m10 PARTITION OF audit.entry_events
    FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
-- Criar particoes automaticamente (via script mensal)
```

#### Tabela: `audit.user_events`

```sql
CREATE TABLE audit.user_events (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    correlation_id UUID NOT NULL,

    user_id UUID,
    username VARCHAR(255),
    ip_address INET NOT NULL,
    user_agent TEXT,

    action_type VARCHAR(20) NOT NULL,
    action_status VARCHAR(20) NOT NULL,

    data JSONB,
    metadata JSONB,

    CONSTRAINT user_events_event_type_check CHECK (event_type IN (
        'USER_LOGIN_SUCCESS', 'USER_LOGIN_FAILED', 'USER_LOGOUT',
        'USER_PASSWORD_CHANGED', 'USER_MFA_ENABLED', 'USER_MFA_DISABLED',
        'USER_CREATED', 'USER_PERMISSIONS_CHANGED'
    ))
);

CREATE INDEX idx_user_events_timestamp ON audit.user_events (timestamp DESC);
CREATE INDEX idx_user_events_user_id ON audit.user_events (user_id);
CREATE INDEX idx_user_events_ip_address ON audit.user_events (ip_address);
```

#### Tabela: `audit.security_events`

```sql
CREATE TABLE audit.security_events (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'WARN',
    correlation_id UUID NOT NULL,

    user_id UUID,
    ip_address INET NOT NULL,
    resource_type VARCHAR(50),
    resource_id VARCHAR(255),

    anomaly_type VARCHAR(50),
    details JSONB,

    CONSTRAINT security_events_event_type_check CHECK (event_type IN (
        'UNAUTHORIZED_ACCESS_ATTEMPT', 'RATE_LIMIT_EXCEEDED',
        'MTLS_CERT_INVALID', 'RSFN_AUTH_FAILED', 'SECURITY_ANOMALY_DETECTED'
    ))
);

CREATE INDEX idx_security_events_timestamp ON audit.security_events (timestamp DESC);
CREATE INDEX idx_security_events_ip_address ON audit.security_events (ip_address);
CREATE INDEX idx_security_events_severity ON audit.security_events (severity);
```

### 6.2 Retencao de Dados

#### Politica de Retencao

| Categoria | Retencao Online (PostgreSQL) | Retencao Arquivo (S3 Glacier) | Total |
|-----------|------------------------------|-------------------------------|-------|
| **Audit Logs (DICT)** | 1 ano | 4 anos | **5 anos** (Bacen) |
| **User Logs** | 6 meses | 6 meses | **1 ano** |
| **Security Logs** | 1 ano | 1 ano | **2 anos** |
| **System Logs** | 90 dias | - | **90 dias** |

#### Script de Archival (Executar mensalmente)

```sql
-- 1. Exportar logs de 1 ano atras para arquivo JSON (via pg_dump)
COPY (
    SELECT row_to_json(t)
    FROM audit.entry_events t
    WHERE timestamp >= '2024-10-01' AND timestamp < '2024-11-01'
) TO '/backup/audit/entry_events_2024_10.json';

-- 2. Upload para S3 Glacier (via AWS CLI)
-- aws s3 cp /backup/audit/entry_events_2024_10.json s3://lbpay-audit-archive/2024/10/ --storage-class GLACIER

-- 3. Deletar logs do PostgreSQL (apos confirmacao de upload)
DELETE FROM audit.entry_events
WHERE timestamp >= '2024-10-01' AND timestamp < '2024-11-01';

-- 4. VACUUM para liberar espaco
VACUUM FULL audit.entry_events;
```

### 6.3 S3 Archival (AWS S3 Glacier)

#### Estrutura de Pastas S3

```
s3://lbpay-audit-archive/
├── entry_events/
│   ├── 2025/
│   │   ├── 01/entry_events_2025_01.json.gz
│   │   ├── 02/entry_events_2025_02.json.gz
│   │   └── ...
│   ├── 2026/
│   └── ...
├── user_events/
│   ├── 2025/
│   └── ...
└── security_events/
    ├── 2025/
    └── ...
```

#### Lifecycle Policy S3

```json
{
  "Rules": [
    {
      "Id": "ArchiveAuditLogs",
      "Status": "Enabled",
      "Prefix": "entry_events/",
      "Transitions": [
        {
          "Days": 0,
          "StorageClass": "GLACIER"
        }
      ],
      "Expiration": {
        "Days": 1825
      }
    }
  ]
}
```

---

## 7. Consulta e Analise de Logs

### 7.1 Queries Comuns

#### 7.1.1 Rastrear Operacoes de um Usuario

```sql
SELECT
    timestamp,
    event_type,
    action_type,
    resource_type,
    resource_id,
    action_status
FROM audit.entry_events
WHERE actor_user_id = 'user-550e8400'
ORDER BY timestamp DESC
LIMIT 100;
```

#### 7.1.2 Detectar Tentativas de Acesso Nao Autorizadas

```sql
SELECT
    timestamp,
    actor_ip_address,
    resource_id,
    action_endpoint,
    COUNT(*) as attempt_count
FROM audit.security_events
WHERE event_type = 'UNAUTHORIZED_ACCESS_ATTEMPT'
  AND timestamp >= NOW() - INTERVAL '24 hours'
GROUP BY timestamp, actor_ip_address, resource_id, action_endpoint
HAVING COUNT(*) > 5
ORDER BY attempt_count DESC;
```

#### 7.1.3 Rastrear Workflow Completo (Correlation ID)

```sql
SELECT
    timestamp,
    service_name,
    event_type,
    action_status
FROM audit.entry_events
WHERE correlation_id = '550e8400-e29b-41d4-a716-446655440000'
ORDER BY timestamp ASC;
```

### 7.2 Dashboards (Grafana)

#### Dashboard 1: Audit Overview

- Total de eventos por tipo (ultimas 24h)
- Taxa de sucesso vs falha
- Top 10 usuarios mais ativos
- Top 10 IPs mais ativos

#### Dashboard 2: Security Monitoring

- Tentativas de acesso nao autorizado (timeline)
- Rate limit violations (por IP)
- Certificados mTLS invalidos
- Anomalias detectadas

#### Dashboard 3: Compliance LGPD

- Requests DSAR (por dia)
- Exportacoes de dados
- Exclusoes de dados (LGPD)
- Correcoes de dados

---

## 8. Alertas

### 8.1 Alertas Criticos

| Alerta | Condicao | Acao |
|--------|----------|------|
| **Multiplas tentativas de login falhadas** | > 5 falhas do mesmo IP em 5 min | Bloquear IP temporariamente |
| **Acesso nao autorizado a dados sensiveis** | Qualquer tentativa | Notificar Security Team imediatamente |
| **Certificado mTLS invalido** | Qualquer tentativa | Alertar Infra Team |
| **Anomalia de seguranca detectada** | Qualquer deteccao | Notificar Security + DPO |

### 8.2 Implementacao (Prometheus AlertManager)

```yaml
groups:
  - name: audit_alerts
    rules:
      - alert: MultipleFailedLogins
        expr: |
          sum(rate(audit_user_events{event_type="USER_LOGIN_FAILED"}[5m])) by (actor_ip_address) > 5
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Multiple failed login attempts from IP {{ $labels.actor_ip_address }}"

      - alert: UnauthorizedAccessAttempt
        expr: |
          sum(rate(audit_security_events{event_type="UNAUTHORIZED_ACCESS_ATTEMPT"}[1m])) > 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Unauthorized access attempt detected"
```

---

## 9. Conformidade LGPD

### 9.1 Logs de Acesso a Dados Pessoais

Conforme LGPD Art. 37, a LBPay DEVE manter logs de:

- Quem acessou dados pessoais (user_id)
- Quando acessou (timestamp)
- Qual dado foi acessado (resource_id)
- Finalidade do acesso (metadata.purpose)

### 9.2 Direitos dos Titulares - Rastreabilidade

Para atender direitos do titular (Art. 18 LGPD), os logs DEVEM permitir:

- Listar todas as operacoes sobre dados de um usuario
- Identificar quem acessou os dados do usuario
- Comprovar exclusao de dados (quando solicitado)

---

## 10. Testes e Validacao

### 10.1 Checklist de Validacao

- [ ] Todos os eventos criticos sao auditados
- [ ] Masking de dados sensiveis funciona corretamente
- [ ] Logs seguem formato JSON padrao
- [ ] Correlation IDs permitem rastreamento end-to-end
- [ ] Retencao de 5 anos implementada (PostgreSQL + S3)
- [ ] Queries de auditoria executam em < 5 segundos
- [ ] Alertas de seguranca funcionam corretamente
- [ ] Dashboards de compliance disponiveis

---

## 11. Referencias

### Internas

- [SEC-007: LGPD Data Protection](../13_Seguranca/SEC-007_LGPD_Data_Protection.md)
- [REG-001: Regulatory Compliance Bacen](../06_Regulatorio/REG-001_Requisitos_Regulatorios_Bacen.md)
- [DAT-001: Schema Database Core DICT](../03_Dados/DAT-001_Schema_Database_Core_DICT.md)

### Externas

- Lei 13.709/2018 (LGPD) - Art. 37, Art. 48
- Manual Operacional DICT v8 - Secao "Auditoria"
- ISO 27001:2013 - A.12.4.1 (Event logging)

---

**Versao**: 1.0
**Status**: Especificacao Completa
**Proxima Revisao**: Anual ou apos auditoria externa
