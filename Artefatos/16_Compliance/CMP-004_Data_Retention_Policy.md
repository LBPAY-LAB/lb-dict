# CMP-004: Data Retention Policy

**Projeto**: DICT - Diretorio de Identificadores de Contas Transacionais (LBPay)
**Versao**: 1.0
**Data**: 2025-10-25
**Status**: Politica Operacional
**Responsavel**: DPO + DBA + Compliance Team

---

## 1. Resumo Executivo

Este documento define a **Politica de Retencao de Dados** para todos os tipos de dados tratados no sistema DICT da LBPay, em conformidade com:

- **Banco Central**: Retencao de 5 anos para dados de auditoria (Circular 3.682/2013)
- **LGPD**: Retencao limitada ao necessario para a finalidade (Art. 15, Lei 13.709/2018)
- **Obrigacoes Contratuais**: Retencao necessaria para cumprimento de contratos

**Objetivo**: Estabelecer periodos de retencao claros, procedimentos de exclusao e responsabilidades.

---

## 2. Principios da Politica de Retencao

### 2.1 Principios LGPD

- **Minimizacao de Dados**: Reter dados apenas pelo tempo necessario (LGPD Art. 6o, III)
- **Qualidade dos Dados**: Manter dados atualizados e relevantes (LGPD Art. 6o, V)
- **Seguranca**: Proteger dados durante toda a retencao (LGPD Art. 6o, VII)
- **Transparencia**: Informar titulares sobre periodos de retencao (LGPD Art. 6o, VI)

### 2.2 Principios Regulatorios Bacen

- **Auditoria**: Retencao de 5 anos para logs de auditoria (Circular 3.682/2013)
- **Rastreabilidade**: Comprovar operacoes em caso de auditoria ou investigacao
- **Disponibilidade**: Dados disponiveis para consulta rapida (online) ou arquivados (offline)

---

## 3. Categorias de Dados e Periodos de Retencao

### 3.1 Dados Operacionais (DICT)

#### 3.1.1 Chaves PIX Ativas

| Dado | Periodo de Retencao | Justificativa | Base Legal |
|------|---------------------|---------------|------------|
| **Chave PIX ativa** | Enquanto ativa | Necessario para transacoes PIX | Execucao de contrato (LGPD Art. 7o, V) |
| **Dados da conta (ISPB, agencia, conta)** | Enquanto chave ativa | Vinculo necessario para transacoes | Execucao de contrato |
| **Nome do titular** | Enquanto chave ativa | Identificacao do titular | Execucao de contrato |

**Armazenamento**: PostgreSQL (tabela `dict.entries`)

**Exclusao**: Quando usuario solicita exclusao de chave PIX OU encerramento de conta.

---

#### 3.1.2 Chaves PIX Deletadas (Soft Delete)

| Dado | Periodo de Retencao | Justificativa | Base Legal |
|------|---------------------|---------------|------------|
| **Chave PIX deletada** | 5 anos apos exclusao | Auditoria e compliance Bacen | Obrigacao legal (Circular 3.682) |
| **Historico de alteracoes** | 5 anos apos exclusao | Rastreabilidade | Obrigacao legal |

**Armazenamento**: PostgreSQL (tabela `dict.entries`, campo `status = 'DELETED'`, `deleted_at`)

**Exclusao**: Hard delete apos 5 anos (script automatizado mensal).

**Procedimento de Hard Delete**:
```sql
-- Executar mensalmente (via cron job)
DELETE FROM dict.entries
WHERE status = 'DELETED'
  AND deleted_at < NOW() - INTERVAL '5 years';
```

---

#### 3.1.3 Claims (Reivindicacoes e Portabilidades)

| Dado | Periodo de Retencao | Justificativa | Base Legal |
|------|---------------------|---------------|------------|
| **Claim (qualquer status)** | 5 anos apos criacao | Auditoria e compliance Bacen | Obrigacao legal |
| **Historico de transicoes de status** | 5 anos apos criacao | Rastreabilidade | Obrigacao legal |

**Armazenamento**: PostgreSQL (tabelas `dict.claims`, `dict.claim_events`)

**Exclusao**: Hard delete apos 5 anos.

---

### 3.2 Dados de Auditoria

#### 3.2.1 Logs de Auditoria (DICT Operations)

| Dado | Periodo Online (PostgreSQL) | Periodo Arquivo (S3 Glacier) | Total | Base Legal |
|------|----------------------------|------------------------------|-------|------------|
| **Entry Events** | 1 ano | 4 anos | **5 anos** | Circular 3.682 + LGPD Art. 37 |
| **Claim Events** | 1 ano | 4 anos | **5 anos** | Circular 3.682 |
| **Portability Events** | 1 ano | 4 anos | **5 anos** | Circular 3.682 |

**Armazenamento**:
- Online: PostgreSQL (schema `audit`, tabelas `entry_events`, `claim_events`, etc.)
- Arquivo: AWS S3 Glacier (bucket `lbpay-audit-archive`)

**Procedimento de Archival**:
1. Exportar logs de 1 ano atras para JSON (mensalmente)
2. Upload para S3 Glacier
3. Deletar do PostgreSQL apos confirmacao de upload
4. Hard delete do S3 Glacier apos 5 anos (via S3 Lifecycle Policy)

**Script de Archival** (ver CMP-001, Secao 6.2):
```bash
# Executar mensalmente (dia 1 de cada mes)
/scripts/audit-archival.sh
```

---

#### 3.2.2 Logs de Usuario (Login, Logout, Alteracoes)

| Dado | Periodo Online | Periodo Arquivo | Total |
|------|----------------|-----------------|-------|
| **User Events** | 6 meses | 6 meses | **1 ano** |

**Armazenamento**: PostgreSQL (`audit.user_events`) + S3 Glacier

**Exclusao**: Hard delete apos 1 ano.

---

#### 3.2.3 Logs de Seguranca (Tentativas de Acesso, Anomalias)

| Dado | Periodo Online | Periodo Arquivo | Total |
|------|----------------|-----------------|-------|
| **Security Events** | 1 ano | 1 ano | **2 anos** |

**Armazenamento**: PostgreSQL (`audit.security_events`) + S3 Glacier

**Exclusao**: Hard delete apos 2 anos.

---

### 3.3 Dados de Sistema (Logs Tecnicos)

| Dado | Periodo de Retencao | Justificativa | Armazenamento |
|------|---------------------|---------------|---------------|
| **Logs de Aplicacao (stdout/stderr)** | 90 dias | Debug e troubleshooting | CloudWatch Logs (AWS) |
| **Metricas (Prometheus)** | 90 dias | Monitoramento | Prometheus TSDB |
| **Traces (OpenTelemetry)** | 30 dias | Performance analysis | Jaeger/Tempo |

**Exclusao**: Automatica (configuracao de retencao no CloudWatch, Prometheus, Jaeger).

---

### 3.4 Dados de Usuario (Conta LBPay)

| Dado | Periodo de Retencao | Justificativa | Base Legal |
|------|---------------------|---------------|------------|
| **Dados de cadastro (nome, CPF, email)** | Enquanto conta ativa | Relacionamento contratual | Execucao de contrato |
| **Dados de cadastro (apos encerramento de conta)** | 5 anos apos encerramento | Obrigacoes fiscais/legais | Obrigacao legal |
| **Historico de transacoes** | 5 anos | Auditoria | Obrigacao legal |

**Armazenamento**: PostgreSQL (tabela `users`)

**Exclusao**: Hard delete apos 5 anos do encerramento da conta.

---

### 3.5 Backups

| Tipo de Backup | Periodo de Retencao | Armazenamento | Criptografia |
|----------------|---------------------|---------------|--------------|
| **Backup Diario (PostgreSQL)** | 30 dias | AWS S3 (Standard) | AES-256 |
| **Backup Semanal** | 3 meses | AWS S3 (Standard-IA) | AES-256 |
| **Backup Mensal** | 5 anos | AWS S3 (Glacier Deep Archive) | AES-256 |

**Exclusao**: Automatica (via S3 Lifecycle Policy).

**S3 Lifecycle Policy**:
```json
{
  "Rules": [
    {
      "Id": "BackupRetention",
      "Status": "Enabled",
      "Prefix": "backups/daily/",
      "Expiration": { "Days": 30 }
    },
    {
      "Id": "WeeklyBackupRetention",
      "Status": "Enabled",
      "Prefix": "backups/weekly/",
      "Transitions": [
        { "Days": 7, "StorageClass": "STANDARD_IA" }
      ],
      "Expiration": { "Days": 90 }
    },
    {
      "Id": "MonthlyBackupRetention",
      "Status": "Enabled",
      "Prefix": "backups/monthly/",
      "Transitions": [
        { "Days": 30, "StorageClass": "GLACIER_DEEP_ARCHIVE" }
      ],
      "Expiration": { "Days": 1825 }
    }
  ]
}
```

---

## 4. Procedimentos de Exclusao de Dados

### 4.1 Soft Delete (Exclusao Logica)

**Quando Aplicavel**: Chaves PIX deletadas (mas ainda dentro do periodo de retencao de 5 anos)

**Procedimento**:
1. Marcar registro com `status = 'DELETED'`
2. Registrar `deleted_at = NOW()`
3. Manter dados intactos (nao modificar)
4. Nao retornar em consultas normais (filtrar WHERE status != 'DELETED')

**SQL**:
```sql
UPDATE dict.entries
SET status = 'DELETED',
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = '550e8400-e29b-41d4-a716-446655440000';
```

---

### 4.2 Hard Delete (Exclusao Fisica)

**Quando Aplicavel**: Dados alem do periodo de retencao

**Procedimento**:
1. Validar que periodo de retencao foi cumprido
2. Deletar registros do banco de dados
3. Purgar backups antigos (se aplicavel)
4. Registrar exclusao em log de auditoria

**SQL**:
```sql
-- Deletar chaves PIX apos 5 anos de soft delete
DELETE FROM dict.entries
WHERE status = 'DELETED'
  AND deleted_at < NOW() - INTERVAL '5 years';

-- Deletar claims apos 5 anos
DELETE FROM dict.claims
WHERE created_at < NOW() - INTERVAL '5 years';

-- Deletar logs de auditoria apos 5 anos (apos archival para S3)
DELETE FROM audit.entry_events
WHERE timestamp < NOW() - INTERVAL '5 years';
```

**Automacao**: Script executado mensalmente (cron job, dia 1 de cada mes).

---

### 4.3 Purge de Backups Antigos

**Procedimento**:
1. AWS S3 Lifecycle Policy deleta automaticamente backups apos periodo de retencao
2. Validacao mensal: verificar que Lifecycle Policy esta ativa

**Validacao**:
```bash
# Verificar Lifecycle Policy do bucket de backups
aws s3api get-bucket-lifecycle-configuration --bucket lbpay-backups

# Listar backups com mais de 5 anos (nao devem existir)
aws s3 ls s3://lbpay-backups/monthly/ --recursive | awk '{if ($1 < "2020-10-25") print $0}'
```

---

## 5. Exclusao de Dados por Solicitacao do Titular (LGPD)

### 5.1 Direito ao Esquecimento (LGPD Art. 18, VI)

**Processo**:
1. Usuario solicita exclusao de seus dados via app/suporte
2. Validar identidade do usuario (autenticacao)
3. Verificar se exclusao e permitida (nao violar obrigacoes legais)
4. Executar soft delete imediatamente
5. Hard delete apos periodo de retencao (5 anos)

**Restricoes**:
- Logs de auditoria (Bacen exige 5 anos) NAO podem ser deletados antes do prazo
- Dados necessarios para obrigacoes legais NAO podem ser deletados

**Comunicacao ao Usuario**:
```
Prezado [Nome],

Sua solicitacao de exclusao de dados foi recebida e processada.

- Chaves PIX: Deletadas imediatamente (soft delete)
- Dados pessoais: Deletados imediatamente (soft delete)
- Logs de auditoria: Mantidos por 5 anos conforme exigencia do Banco Central

Apos 5 anos, todos os seus dados serao permanentemente deletados de nossos sistemas.

Atenciosamente,
LB Pagamentos - Equipe de Privacidade
```

---

## 6. Responsabilidades

### 6.1 Matriz de Responsabilidades

| Responsabilidade | Responsavel | Frequencia |
|------------------|-------------|------------|
| **Implementacao de soft delete** | Backend Team | Continuo |
| **Script de hard delete (mensal)** | DBA + DevOps | Mensal (dia 1) |
| **Archival de logs para S3** | DBA + DevOps | Mensal (dia 1) |
| **Validacao de S3 Lifecycle Policy** | DevOps Team | Mensal |
| **Auditoria de retencao (compliance)** | DPO + Compliance | Trimestral |
| **Revisao da Politica de Retencao** | DPO + Legal | Anual |
| **Atendimento a solicitacoes de exclusao (LGPD)** | DPO + Support Team | Sob demanda |

---

### 6.2 Checklist Mensal (DBA + DevOps)

- [ ] Executar script de hard delete (dia 1)
- [ ] Executar archival de logs para S3 (dia 1)
- [ ] Validar upload bem-sucedido para S3 Glacier
- [ ] Validar S3 Lifecycle Policy ativa
- [ ] Gerar relatorio de retencao (quantidades de dados por categoria)
- [ ] Revisar anomalias (dados fora do periodo esperado)

---

## 7. Monitoramento e Alertas

### 7.1 Metricas de Retencao

| Metrica | Threshold | Alerta |
|---------|-----------|--------|
| **Dados DELETED com > 5 anos** | > 0 | WARN - Executar hard delete |
| **Logs de auditoria com > 5 anos (PostgreSQL)** | > 0 | WARN - Executar archival |
| **Backups mensais com > 5 anos** | > 0 | WARN - Verificar S3 Lifecycle |
| **Tamanho da tabela audit.entry_events** | > 100 GB | INFO - Considerar particoes adicionais |

---

### 7.2 Dashboards (Grafana)

**Dashboard: Data Retention Compliance**

- Total de chaves DELETED (aguardando hard delete)
- Idade media de chaves DELETED
- Total de logs de auditoria por ano
- Tamanho de backups por tipo (diario, semanal, mensal)
- Status de archival (ultimo archival bem-sucedido)

---

## 8. Excecoes e Casos Especiais

### 8.1 Ordem Judicial

**Cenario**: Ordem judicial solicita retencao de dados alem do periodo normal.

**Procedimento**:
1. Receber ordem judicial via canais oficiais (Bacen, justica)
2. Marcar dados com flag `legal_hold = true`
3. NAO deletar dados com legal hold (mesmo apos periodo de retencao)
4. Deletar somente apos autorizacao judicial

**SQL**:
```sql
-- Marcar dados com legal hold
UPDATE dict.entries
SET legal_hold = true,
    legal_hold_reason = 'Ordem Judicial #12345/2025',
    legal_hold_date = NOW()
WHERE id IN (...);

-- Modificar script de hard delete para respeitar legal hold
DELETE FROM dict.entries
WHERE status = 'DELETED'
  AND deleted_at < NOW() - INTERVAL '5 years'
  AND (legal_hold IS NULL OR legal_hold = false);
```

---

### 8.2 Investigacao de Fraude/Seguranca

**Cenario**: Investigacao de fraude exige retencao prolongada.

**Procedimento**:
1. Security Team marca dados com `security_hold = true`
2. Dados retidos ate conclusao da investigacao
3. Apos conclusao: remover flag e seguir retencao normal

---

## 9. Conformidade e Auditorias

### 9.1 Evidencias de Conformidade

Para demonstrar conformidade em auditorias:

1. **Documentacao**: Esta politica (CMP-004)
2. **Logs de Execucao**: Logs de scripts de hard delete e archival
3. **Relatorios Mensais**: Relatorio de retencao (quantidades, datas)
4. **S3 Lifecycle Policies**: Configuracao exportada
5. **Scripts**: Scripts de hard delete e archival (versionados no Git)

---

### 9.2 Auditorias Periodicas

**Frequencia**: Trimestral (interna) + Anual (externa)

**Itens Auditados**:
- Cumprimento de periodos de retencao
- Exclusao correta de dados alem do periodo
- Funcionamento de scripts automatizados
- S3 Lifecycle Policies ativas
- Atendimento a solicitacoes de exclusao (LGPD)

---

## 10. Tabela Resumida: Periodos de Retencao

| Categoria de Dado | Retencao Online | Retencao Arquivo | Total | Base Legal |
|-------------------|-----------------|------------------|-------|------------|
| **Chave PIX ativa** | Enquanto ativa | - | Indeterminado | Execucao de contrato |
| **Chave PIX deletada** | 5 anos | - | 5 anos | Circular 3.682 |
| **Claims** | 5 anos | - | 5 anos | Circular 3.682 |
| **Audit Logs (DICT)** | 1 ano | 4 anos | 5 anos | Circular 3.682 + LGPD |
| **User Logs** | 6 meses | 6 meses | 1 ano | LGPD Art. 37 |
| **Security Logs** | 1 ano | 1 ano | 2 anos | LGPD + ISO 27001 |
| **System Logs** | 90 dias | - | 90 dias | Operacional |
| **Backups Diarios** | 30 dias | - | 30 dias | Contingencia |
| **Backups Mensais** | - | 5 anos | 5 anos | Circular 3.682 |
| **Dados de Usuario (conta ativa)** | Enquanto ativa | - | Indeterminado | Execucao de contrato |
| **Dados de Usuario (conta encerrada)** | 5 anos | - | 5 anos | Obrigacao legal |

---

## 11. Referencias

### Internas

- [CMP-001: Audit Logs Specification](CMP-001_Audit_Logs_Specification.md)
- [CMP-002: LGPD Compliance Checklist](CMP-002_LGPD_Compliance_Checklist.md)
- [SEC-007: LGPD Data Protection](../13_Seguranca/SEC-007_LGPD_Data_Protection.md)
- [REG-001: Requisitos Regulatorios Bacen](../06_Regulatorio/REG-001_Requisitos_Regulatorios_Bacen.md)

### Externas

- Lei 13.709/2018 (LGPD) - Art. 15, Art. 16, Art. 18
- Circular 3.682/2013 (Bacen) - Auditoria de sistemas
- ISO 27001:2013 - A.12.4.2 (Protection of log information)

---

**Versao**: 1.0
**Status**: Politica Operacional
**Proxima Revisao**: Anual ou apos mudanca regulatoria
