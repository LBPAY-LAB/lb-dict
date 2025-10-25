# Compliance

**Propósito**: Documentos de conformidade regulatória e auditoria para o sistema DICT

## 📋 Conteúdo

Esta pasta armazenará:

- **LGPD Compliance**: Documentos de conformidade com LGPD (Lei Geral de Proteção de Dados)
- **Bacen Compliance**: Conformidade com regulamentações do Banco Central
- **Audit Trails**: Trilhas de auditoria e logs de compliance
- **Data Retention Policies**: Políticas de retenção e descarte de dados
- **Privacy Impact Assessment (PIA)**: Avaliações de impacto à privacidade
- **Compliance Reports**: Relatórios de compliance para auditores

## 📁 Estrutura Esperada

```
Compliance/
├── LGPD/
│   ├── LGPD_Compliance_Checklist.md
│   ├── Data_Processing_Record.md
│   ├── Privacy_Policy.md
│   └── Data_Subject_Rights.md
├── Bacen/
│   ├── Resolucao_BCB_4985_Compliance.md
│   ├── Circular_DICT_Compliance.md
│   └── Audit_Requirements.md
├── Audits/
│   ├── Audit_Trail_Design.md
│   ├── Audit_Report_2025_Q1.md
│   └── Audit_Report_2025_Q2.md
├── Data_Retention/
│   ├── Retention_Policy.md
│   └── Data_Deletion_Procedures.md
└── PIA/
    ├── PIA_DICT_System.md
    └── Risk_Assessment.md
```

## 🎯 Principais Áreas de Compliance

### 1. LGPD (Lei Geral de Proteção de Dados)

**Requisitos Principais**:
- ✅ Consentimento explícito para coleta de dados pessoais
- ✅ Direito de acesso, correção, exclusão de dados (DSAR)
- ✅ Minimização de dados (coletar apenas o necessário)
- ✅ Segurança da informação (criptografia, controle de acesso)
- ✅ Notificação de incidentes de segurança (72 horas)
- ✅ DPO (Data Protection Officer) designado

**Dados Pessoais no DICT**:
- CPF, CNPJ (identificadores)
- Nome do titular da conta
- Telefone, email (chaves PIX)
- Dados bancários (ISPB, agência, conta)

**Base Legal**: Execução de contrato (abertura de conta corrente)

### 2. Bacen Compliance

**Regulamentação**: Resolução BCB nº 4.985/2021

**Requisitos Principais**:
- ✅ Certificado digital ICP-Brasil A3 (SEC-002)
- ✅ mTLS para comunicação com DICT Bacen (SEC-001)
- ✅ Assinatura digital XML em mensagens SOAP (SEC-006)
- ✅ Auditoria de todas as operações DICT
- ✅ Disponibilidade 99.9% (SLA)
- ✅ Tempo de resposta < 2 segundos (p95)

### 3. Audit Trail (Trilha de Auditoria)

**O Que Auditar**:
- Todas as operações CRUD em entries (create, update, delete)
- Todas as claims criadas, aceitas, rejeitadas
- Todas as portabilidades de conta
- Acessos administrativos ao sistema
- Mudanças em configurações de segurança
- Falhas de autenticação/autorização

**Formato de Log de Auditoria**:
```json
{
  "timestamp": "2025-10-25T10:00:00Z",
  "event_type": "ENTRY_CREATED",
  "user_id": "user-123",
  "user_ip": "192.168.1.100",
  "resource_type": "entry",
  "resource_id": "entry-550e8400",
  "action": "CREATE",
  "details": {
    "key_type": "CPF",
    "key_value": "***678900",  // Masked
    "account_ispb": "00000000"
  },
  "result": "SUCCESS",
  "request_id": "req-abc123"
}
```

**Retenção de Logs**: 5 anos (exigência Bacen)

### 4. Data Retention Policy

**Períodos de Retenção**:

| Tipo de Dado | Retenção | Base Legal |
|--------------|----------|------------|
| Entries ativas | Enquanto conta ativa | Contrato |
| Entries deletadas | 5 anos | Bacen |
| Claims | 5 anos após conclusão | Bacen |
| Logs de auditoria | 5 anos | Bacen |
| Logs de acesso | 1 ano | LGPD |
| Backups | 30 dias | Operacional |

**Procedimento de Exclusão**:
1. Soft delete (marcar como deletado)
2. Após período de retenção, hard delete (GDPR-compliant)
3. Exclusão de backups após 5 anos

### 5. Data Subject Rights (Direitos do Titular)

**LGPD - Direitos do Titular**:
- **Acesso**: Solicitar cópia de seus dados pessoais
- **Retificação**: Corrigir dados incorretos
- **Exclusão**: Solicitar exclusão de dados (direito ao esquecimento)
- **Portabilidade**: Exportar dados em formato estruturado
- **Oposição**: Opor-se ao tratamento de dados

**SLA para DSAR (Data Subject Access Request)**:
- Resposta inicial: 5 dias úteis
- Conclusão: 15 dias úteis (LGPD)

## 📊 Compliance Checklist

### LGPD Compliance

- [ ] Política de Privacidade publicada e acessível
- [ ] Consentimento coletado antes de processar dados
- [ ] DPO (Data Protection Officer) nomeado
- [ ] Registro de atividades de tratamento mantido
- [ ] Avaliação de impacto à privacidade (PIA) realizada
- [ ] Contratos com processadores de dados assinados
- [ ] Processo de notificação de incidentes implementado
- [ ] Treinamento de funcionários sobre LGPD realizado
- [ ] Processo de DSAR (Data Subject Access Request) implementado
- [ ] Criptografia de dados sensíveis implementada

### Bacen Compliance

- [ ] Certificado ICP-Brasil A3 adquirido e instalado
- [ ] mTLS configurado para comunicação com Bacen
- [ ] Assinatura digital XML implementada
- [ ] Logs de auditoria implementados (5 anos retenção)
- [ ] Monitoramento de disponibilidade (SLA 99.9%)
- [ ] Testes de disaster recovery realizados
- [ ] Documentação técnica completa
- [ ] Homologação com Bacen (ambiente staging)

## 📚 Referências

### Documentos Internos
- [SEC-007: LGPD Data Protection](../13_Seguranca/SEC-007_LGPD_Data_Protection.md) - Proteção de dados pessoais
- [SEC-001: mTLS Configuration](../13_Seguranca/SEC-001_mTLS_Configuration.md) - Segurança na comunicação
- [SEC-002: ICP-Brasil Certificates](../13_Seguranca/SEC-002_ICP_Brasil_Certificates.md) - Certificados digitais
- [DAT-001: Schema Database Core DICT](../03_Dados/DAT-001_Schema_Database_Core_DICT.md) - Audit trail (audit.entry_events)
- [REG-001: Regulatory Compliance Bacen DICT](../06_Regulatorio/REG-001_Regulatory_Compliance_Bacen_DICT.md)

### Legislação e Normas
- [Lei nº 13.709/2018 (LGPD)](http://www.planalto.gov.br/ccivil_03/_ato2015-2018/2018/lei/l13709.htm)
- [Resolução BCB nº 4.985/2021](https://www.bcb.gov.br/estabilidadefinanceira/exibenormativo?tipo=Resolu%C3%A7%C3%A3o%20BCB&numero=4985)
- [Circular DICT Bacen](https://www.bcb.gov.br/estabilidadefinanceira/pix)
- [GDPR (Referência internacional)](https://gdpr-info.eu/)

---

**Status**: 🔴 Pasta vazia (será preenchida na Fase 2+)
**Fase de Preenchimento**: Fase 2 (paralelo ao desenvolvimento)
**Responsável**: Compliance Officer + DPO + Legal
**Revisão**: Anual ou quando houver mudança regulatória
