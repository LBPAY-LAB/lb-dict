# SEC-007: LGPD Data Protection

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Versão**: 1.0
**Data**: 2025-10-25
**Status**: ✅ Especificação Completa
**Responsável**: DPO (Data Protection Officer) + ARCHITECT + Compliance

---

## 📋 Resumo Executivo

Este documento especifica **todas as medidas técnicas e organizacionais** para conformidade com a LGPD (Lei Geral de Proteção de Dados) no sistema DICT, incluindo proteção de dados pessoais, direitos dos titulares, consentimento, e segurança da informação.

**Objetivo**: Garantir conformidade total com a Lei nº 13.709/2018 (LGPD) e evitar penalidades (até 2% do faturamento ou R$ 50 milhões por infração).

---

## 🎯 Contexto Regulatório

### Lei Geral de Proteção de Dados (LGPD)

**Lei nº 13.709/2018** - Em vigor desde setembro de 2020

**Autoridade Nacional**: ANPD (Autoridade Nacional de Proteção de Dados)

**Penalidades**:
- Advertência com prazo para adoção de medidas corretivas
- Multa simples: até 2% do faturamento (limitado a R$ 50 milhões por infração)
- Multa diária: até R$ 50 milhões
- Publicização da infração
- Bloqueio ou eliminação dos dados pessoais

---

## 📊 Dados Pessoais Tratados no DICT

### Inventário de Dados Pessoais

| Categoria | Dados | Finalidade | Base Legal | Retenção |
|-----------|-------|------------|------------|----------|
| **Identificação** | CPF, CNPJ | Chave PIX (entry key) | Execução de contrato | 5 anos após desativação |
| **Identificação** | Nome completo | Titular da conta | Execução de contrato | 5 anos após desativação |
| **Contato** | Telefone | Chave PIX | Execução de contrato | 5 anos após desativação |
| **Contato** | Email | Chave PIX | Execução de contrato | 5 anos após desativação |
| **Financeiros** | ISPB, agência, conta | Conta vinculada à chave | Execução de contrato | 5 anos após desativação |
| **Transacionais** | Histórico de claims | Auditoria e compliance | Obrigação legal (Bacen) | 5 anos |
| **Auditoria** | Logs de acesso | Segurança e compliance | Legítimo interesse | 1 ano |
| **Técnicos** | Endereço IP | Segurança (rate limiting) | Legítimo interesse | 90 dias |

### Dados NÃO Sensíveis

**Importante**: Segundo a LGPD, os dados tratados no DICT são considerados **dados pessoais comuns** (não sensíveis):

- CPF, CNPJ, nome, telefone, email ➝ **Dados pessoais comuns**
- Dados bancários (ISPB, agência, conta) ➝ **Dados pessoais comuns**

**Dados sensíveis** (Art. 5º, II LGPD) incluem: origem racial/étnica, convicção religiosa, opinião política, filiação sindical, dados genéticos, biométricos, saúde, orientação sexual ➝ **NÃO tratados no DICT**

---

## 🔐 Princípios da LGPD Aplicados

### 1. Finalidade (Art. 6º, I)

**Princípio**: Tratamento para propósitos legítimos, específicos e informados ao titular

**Aplicação**:
- ✅ Finalidade clara: "Cadastro de chaves PIX para transações financeiras"
- ✅ Não usar dados para marketing sem consentimento adicional
- ✅ Documentar finalidade em Política de Privacidade

---

### 2. Adequação (Art. 6º, II)

**Princípio**: Compatibilidade do tratamento com as finalidades informadas

**Aplicação**:
- ✅ Coletar apenas dados necessários para operação PIX
- ✅ Não coletar dados irrelevantes (ex: gênero, estado civil)

---

### 3. Necessidade (Art. 6º, III)

**Princípio**: Minimização de dados (coletar apenas o necessário)

**Aplicação**:
- ✅ Campos obrigatórios: CPF/CNPJ, nome, conta
- ✅ Campos opcionais: telefone, email (se não usados como chave)
- ❌ NÃO coletar: endereço completo, renda, profissão (desnecessários para PIX)

---

### 4. Transparência (Art. 6º, VI)

**Princípio**: Informações claras e acessíveis aos titulares

**Aplicação**:
- ✅ Política de Privacidade publicada (link visível no app)
- ✅ Avisos no momento da coleta ("Ao cadastrar chave PIX, você concorda...")
- ✅ Portal de Privacidade (titular pode consultar seus dados)

---

### 5. Segurança (Art. 6º, VII)

**Princípio**: Medidas técnicas e administrativas para proteger dados

**Aplicação** (detalhado na seção [Medidas de Segurança](#medidas-de-segurança)):
- ✅ Criptografia em trânsito (TLS 1.2+)
- ✅ Criptografia em repouso (PostgreSQL Transparent Data Encryption)
- ✅ Controle de acesso (RBAC - Role-Based Access Control)
- ✅ Logs de auditoria (quem acessou o quê, quando)
- ✅ Masking de dados sensíveis em logs

---

### 6. Prevenção (Art. 6º, VIII)

**Princípio**: Medidas para prevenir danos aos titulares

**Aplicação**:
- ✅ Privacy by Design (segurança desde o design)
- ✅ DPIA (Data Protection Impact Assessment) realizado
- ✅ Plano de Resposta a Incidentes (ver seção [Resposta a Incidentes](#resposta-a-incidentes))

---

### 7. Não Discriminação (Art. 6º, IX)

**Princípio**: Não usar dados para fins discriminatórios

**Aplicação**:
- ✅ Não usar dados para análise de crédito sem consentimento
- ✅ Não perfilar usuários sem base legal

---

### 8. Responsabilização e Prestação de Contas (Art. 6º, X)

**Princípio**: Demonstrar conformidade com LGPD

**Aplicação**:
- ✅ Documentação de processos (este documento + ROPA)
- ✅ DPO (Data Protection Officer) designado
- ✅ Auditorias periódicas (anuais)

---

## ⚖️ Bases Legais (Art. 7º LGPD)

### Base Legal Principal: Execução de Contrato (Art. 7º, V)

**Aplicação**:
- **Contrato**: Conta corrente do cliente com LBPay
- **Necessidade**: Cadastro de chave PIX é necessário para executar contrato (transações PIX)
- **Consentimento NÃO necessário** (base legal é o contrato, não consentimento)

### Bases Legais Secundárias

| Dados | Base Legal | Artigo LGPD |
|-------|------------|-------------|
| CPF, nome, conta (chave PIX) | Execução de contrato | Art. 7º, V |
| Logs de auditoria | Obrigação legal (Bacen exige auditoria) | Art. 7º, II |
| Logs de acesso (IP, timestamps) | Legítimo interesse (segurança) | Art. 7º, IX |
| Telefone/email (marketing) | Consentimento (se usado para fins diferentes) | Art. 7º, I |

---

## 👤 Direitos dos Titulares (Art. 18 LGPD)

### Direitos Garantidos pela LGPD

1. **Confirmação de tratamento** (Art. 18, I)
2. **Acesso aos dados** (Art. 18, II)
3. **Correção de dados** (Art. 18, III)
4. **Anonimização, bloqueio ou eliminação** (Art. 18, IV)
5. **Portabilidade** (Art. 18, V)
6. **Eliminação de dados tratados com consentimento** (Art. 18, VI)
7. **Informação sobre compartilhamento** (Art. 18, VII)
8. **Informação sobre não consentimento** (Art. 18, VIII)
9. **Revogação do consentimento** (Art. 18, IX)

---

### Implementação Técnica dos Direitos

#### 1. Acesso aos Dados (DSAR - Data Subject Access Request)

**API Endpoint** (especificação, NÃO implementar agora):
```http
GET /api/v1/privacy/my-data
Authorization: Bearer <user-token>
```

**Response**:
```json
{
  "request_id": "dsar-123456",
  "user_id": "user-550e8400",
  "data": {
    "personal_info": {
      "cpf": "123.456.789-00",
      "name": "João Silva",
      "email": "joao@example.com",
      "phone": "+5511987654321"
    },
    "dict_keys": [
      {
        "key_type": "CPF",
        "key_value": "12345678900",
        "status": "ACTIVE",
        "created_at": "2025-01-15T10:00:00Z",
        "account": {
          "ispb": "00000000",
          "branch": "0001",
          "account": "12345-6"
        }
      }
    ],
    "claims": [
      {
        "claim_id": "c123",
        "status": "COMPLETED",
        "created_at": "2025-02-01T10:00:00Z",
        "completed_at": "2025-02-10T10:00:00Z"
      }
    ]
  },
  "generated_at": "2025-10-25T10:00:00Z"
}
```

**SLA**: Responder em até **15 dias úteis** (Art. 18, §3º LGPD)

---

#### 2. Correção de Dados

**Processo**:
1. Cliente identifica dado incorreto (ex: nome errado)
2. Cliente solicita correção via app/suporte
3. Sistema valida identidade do cliente (autenticação)
4. Sistema atualiza dado no PostgreSQL
5. Sistema notifica Bacen via Bridge (UpdateEntry)
6. Sistema registra alteração em log de auditoria

**Importante**: Dados como CPF/CNPJ **NÃO podem ser corrigidos** (são imutáveis). Cliente deve deletar chave e criar nova.

---

#### 3. Exclusão de Dados (Direito ao Esquecimento)

**Regras**:
- ✅ Cliente pode solicitar exclusão de chave PIX a qualquer momento
- ⚠️ **Exceção**: Dados de auditoria (Bacen exige 5 anos) ➝ NÃO podem ser deletados
- ✅ Após 5 anos: hard delete completo (inclusive backups)

**Processo de Exclusão**:
```sql
-- 1. Soft delete (marcar como deletado)
UPDATE dict.entries
SET status = 'DELETED', deleted_at = NOW()
WHERE id = '550e8400-e29b-41d4-a716-446655440000';

-- 2. Após período de retenção (5 anos), hard delete
DELETE FROM dict.entries
WHERE deleted_at < NOW() - INTERVAL '5 years';

-- 3. Purge de backups antigos (automatizado)
-- Script rodado mensalmente para deletar backups > 5 anos
```

**SLA**: Exclusão imediata (soft delete), hard delete após período de retenção

---

#### 4. Portabilidade de Dados

**Formato**: JSON ou CSV (a escolha do titular)

**API Endpoint**:
```http
GET /api/v1/privacy/export?format=json
Authorization: Bearer <user-token>
```

**Response**: Arquivo JSON/CSV para download

**SLA**: Geração do arquivo em até **5 dias úteis**

---

#### 5. Revogação de Consentimento

**Aplicação**:
- Se consentimento foi usado para marketing (ex: email promocional usando chave email)
- Cliente pode revogar consentimento a qualquer momento
- Sistema deve parar imediatamente de usar dados para marketing

**Importante**: Revogação **NÃO afeta** dados necessários para contrato (chave PIX continua ativa)

---

## 🔐 Medidas de Segurança Técnicas

### 1. Criptografia

#### Em Trânsito (TLS 1.2+)
```yaml
# Todas as conexões HTTPS/gRPC/mTLS
- FrontEnd → Core DICT: TLS 1.2+
- Core → Connect: gRPC with TLS
- Connect → Bridge: gRPC with TLS
- Bridge → Bacen: mTLS (certificado ICP-Brasil A3)
```

#### Em Repouso
```yaml
# PostgreSQL
- Transparent Data Encryption (TDE)
- Encryption at rest habilitado (AES-256)

# Redis
- Encryption at rest (se suportado pelo provider)

# Backups
- Criptografados com AES-256
- Chaves armazenadas no Vault (HashiCorp ou AWS KMS)
```

---

### 2. Controle de Acesso (RBAC)

```yaml
# Roles
Roles:
  - name: admin
    permissions:
      - dict:entries:*
      - dict:claims:*
      - dict:users:*
      - audit:logs:read

  - name: support
    permissions:
      - dict:entries:read
      - dict:claims:read
      - audit:logs:read

  - name: user
    permissions:
      - dict:entries:read_own
      - dict:entries:create_own
      - dict:entries:delete_own
      - dict:claims:read_own

# Implementação (pseudocódigo)
func CanAccessEntry(user *User, entryID string) bool {
    entry, _ := repo.GetEntry(entryID)

    // Admin pode tudo
    if user.HasRole("admin") {
        return true
    }

    // Usuário só pode acessar suas próprias entries
    if user.ID == entry.UserID {
        return true
    }

    // Support pode ler (mas não modificar)
    if user.HasRole("support") && isReadOperation {
        return true
    }

    return false
}
```

---

### 3. Masking de Dados em Logs

```go
// Pseudocódigo (especificação)
type Entry struct {
    ID       string
    KeyType  string
    KeyValue string  // SENSITIVE
    Account  Account // SENSITIVE
}

// Masking function
func (e *Entry) MarshalJSON() ([]byte, error) {
    type Alias Entry
    return json.Marshal(&struct {
        KeyValue string `json:"key_value"`
        *Alias
    }{
        KeyValue: maskSensitiveData(e.KeyValue),
        Alias:    (*Alias)(e),
    })
}

func maskSensitiveData(value string) string {
    if len(value) <= 4 {
        return "***"
    }
    // Mostrar apenas últimos 4 dígitos
    return "***" + value[len(value)-4:]
}

// Logs
logger.Info("Entry created",
    zap.String("entry_id", entry.ID),
    zap.String("key_value", maskSensitiveData(entry.KeyValue)),  // ***8900
)
```

---

### 4. Auditoria (Audit Trail)

**Todas as operações CRUD devem ser auditadas**:

```sql
-- Tabela de auditoria (ver DAT-001)
CREATE TABLE audit.entry_events (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,  -- CREATE, UPDATE, DELETE, ACCESS
    entry_id UUID NOT NULL,
    user_id UUID NOT NULL,
    user_ip INET,
    operation VARCHAR(100),
    old_value JSONB,
    new_value JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    request_id VARCHAR(255)
);

-- Exemplo de log de auditoria
INSERT INTO audit.entry_events (event_type, entry_id, user_id, user_ip, operation, new_value)
VALUES (
    'CREATE',
    '550e8400-e29b-41d4-a716-446655440000',
    'user-123',
    '192.168.1.100',
    'CreateEntry',
    '{"key_type": "CPF", "key_value": "***8900", "account": "***45-6"}'::jsonb
);
```

**Retenção de Logs de Auditoria**: **5 anos** (exigência Bacen)

---

## 🚨 Resposta a Incidentes de Segurança

### Plano de Resposta a Incidentes (IRP)

#### Fase 1: Detecção (< 1 hora)

**Gatilhos**:
- Alertas de monitoramento (Prometheus, Grafana)
- Relatórios de usuários (suporte)
- Penetration tests / Security scans

**Ações**:
1. Confirmar incidente
2. Classificar gravidade (baixa, média, alta, crítica)
3. Notificar equipe de segurança

---

#### Fase 2: Contenção (< 4 horas)

**Ações**:
1. Isolar sistemas afetados (se necessário)
2. Bloquear acesso não autorizado
3. Preservar evidências (logs, snapshots)

---

#### Fase 3: Erradicação (< 24 horas)

**Ações**:
1. Identificar causa raiz
2. Remover vulnerabilidade
3. Aplicar patches de segurança

---

#### Fase 4: Recuperação (< 48 horas)

**Ações**:
1. Restaurar sistemas afetados
2. Validar integridade de dados
3. Monitorar atividade suspeita

---

#### Fase 5: Notificação (< 72 horas)

**Obrigação Legal (Art. 48 LGPD)**:
- ✅ Notificar **ANPD** em até **72 horas** após ter ciência do incidente
- ✅ Notificar **titulares afetados** (se incidente pode gerar risco/dano relevante)

**Template de Notificação**:
```
Assunto: Notificação de Incidente de Segurança - DICT LBPay

Prezado(a) [Nome],

Identificamos um incidente de segurança em [Data] que pode ter afetado seus dados pessoais no sistema DICT.

**Dados afetados**: [Listar: CPF, nome, telefone, etc.]
**Causa**: [Descrição breve da causa]
**Ações tomadas**: [Contenção, erradicação, recuperação]
**Recomendações para você**: [Ex: Trocar senha, monitorar conta bancária]

Para mais informações, entre em contato com nosso DPO:
Email: dpo@lbpay.com.br
Telefone: (11) 1234-5678

Atenciosamente,
LB Pagamentos - Equipe de Segurança
```

---

#### Fase 6: Lições Aprendidas (< 1 semana)

**Ações**:
1. Reunião post-mortem
2. Documentar lições aprendidas
3. Atualizar plano de resposta a incidentes
4. Implementar melhorias de segurança

---

## 📋 Checklist de Conformidade LGPD

### Organizacional

- [ ] **DPO nomeado** e registrado na ANPD
- [ ] **Política de Privacidade** publicada e acessível
- [ ] **ROPA** (Registro de Operações de Tratamento de Dados) mantido atualizado
- [ ] **DPIA** (Data Protection Impact Assessment) realizado
- [ ] **Contratos com processadores** de dados assinados (ex: AWS, Google Cloud)
- [ ] **Treinamento de funcionários** sobre LGPD realizado
- [ ] **Plano de Resposta a Incidentes** documentado e testado

### Técnico

- [ ] **Criptografia** em trânsito (TLS 1.2+) e em repouso (AES-256)
- [ ] **Controle de acesso** (RBAC) implementado
- [ ] **Auditoria** (logs de todas operações) implementada (5 anos retenção)
- [ ] **Masking de dados** sensíveis em logs
- [ ] **API de DSAR** (Data Subject Access Request) implementada
- [ ] **Exclusão de dados** (soft delete + hard delete após retenção)
- [ ] **Portabilidade de dados** (export JSON/CSV) implementada
- [ ] **Monitoramento de segurança** (alerts para atividades suspeitas)
- [ ] **Backups criptografados** e testados
- [ ] **Política de retenção** de dados implementada

### Processos

- [ ] **Processo de DSAR** documentado (SLA: 15 dias úteis)
- [ ] **Processo de exclusão de dados** documentado
- [ ] **Processo de correção de dados** documentado
- [ ] **Processo de portabilidade** documentado
- [ ] **Processo de resposta a incidentes** testado (simulação anual)
- [ ] **Processo de revisão de privacidade** (Privacy Review) em mudanças de sistema

---

## 📚 Referências

### Documentos Internos
- [SEC-001: mTLS Configuration](SEC-001_mTLS_Configuration.md) - Segurança em trânsito
- [SEC-002: ICP-Brasil Certificates](SEC-002_ICP_Brasil_Certificates.md) - Certificados
- [DAT-001: Schema Database Core DICT](../../03_Dados/DAT-001_Schema_Database_Core_DICT.md) - Audit trail
- [REG-001: Regulatory Compliance Bacen DICT](../../06_Regulatorio/REG-001_Regulatory_Compliance_Bacen_DICT.md)
- [Compliance](../../16_Compliance/)

### Legislação e Normas
- [Lei nº 13.709/2018 (LGPD)](http://www.planalto.gov.br/ccivil_03/_ato2015-2018/2018/lei/l13709.htm)
- [Guia de Boas Práticas LGPD - ANPD](https://www.gov.br/anpd/)
- [GDPR (Referência internacional)](https://gdpr-info.eu/)
- [ISO 27001](https://www.iso.org/isoiec-27001-information-security.html) - Segurança da informação
- [ISO 27701](https://www.iso.org/standard/71670.html) - Privacy Information Management

---

**Versão**: 1.0
**Status**: ✅ Especificação Completa (Aguardando implementação)
**Próxima Revisão**: Anual ou quando houver mudança na legislação

---

**IMPORTANTE**: Este é um documento de **especificação técnica e de compliance**. A implementação será feita pelos desenvolvedores e equipe de compliance em fase posterior, baseando-se neste documento.

**DPO**: Designar DPO (Data Protection Officer) é **obrigatório** para LBPay (instituição financeira tratando dados de alto volume).
