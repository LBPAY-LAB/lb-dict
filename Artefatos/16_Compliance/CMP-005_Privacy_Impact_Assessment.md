# CMP-005: Privacy Impact Assessment (DPIA)

**Projeto**: DICT - Diretorio de Identificadores de Contas Transacionais (LBPay)
**Versao**: 1.0
**Data**: 2025-10-25
**Status**: Relatorio de Impacto a Protecao de Dados
**Responsavel**: DPO (Data Protection Officer) + Security Team + Architect

---

## 1. Resumo Executivo

Este documento apresenta a **Avaliacao de Impacto a Protecao de Dados** (DPIA - Data Protection Impact Assessment, ou RIPD - Relatorio de Impacto a Protecao de Dados Pessoais) do sistema DICT da LBPay, conforme exigido pela LGPD Art. 38.

**Objetivo**: Identificar e mitigar riscos ao tratamento de dados pessoais no sistema DICT.

**Conclusao**: O sistema DICT trata dados pessoais de **alto volume** e **alta sensibilidade** (dados financeiros). A LBPay implementou **medidas tecnicas e organizacionais robustas** para mitigar riscos. O risco residual e considerado **BAIXO** apos implementacao completa das medidas de seguranca.

---

## 2. Contexto do Tratamento de Dados

### 2.1 Descricao do Sistema DICT

**Sistema**: Diretorio de Identificadores de Contas Transacionais (DICT) da LBPay

**Finalidade**: Cadastro, gerenciamento e consulta de chaves PIX para transacoes financeiras instantaneas.

**Usuarios**: Clientes pessoa fisica (CPF) e pessoa juridica (CNPJ) da LBPay.

**Volume Estimado**:
- 100.000+ usuarios (ano 1)
- 500.000+ chaves PIX (ano 1)
- 10.000.000+ operacoes/ano

---

### 2.2 Base Legal (LGPD)

**Base Legal Principal**: **Execucao de Contrato** (LGPD Art. 7o, V)

- **Contrato**: Conta corrente/pagamento do cliente com LBPay
- **Necessidade**: Cadastro de chave PIX e necessario para realizar transacoes PIX

**Bases Legais Secundarias**:
- Logs de auditoria: **Obrigacao Legal** (Art. 7o, II) - Bacen exige retencao de 5 anos
- Logs de acesso (IP, timestamps): **Legitimo Interesse** (Art. 7o, IX) - Seguranca
- Marketing (se aplicavel): **Consentimento** (Art. 7o, I)

---

## 3. Inventario de Dados Pessoais Tratados (PII)

### 3.1 Dados Pessoais Identificaveis (PII)

| Categoria | Dados Coletados | Finalidade | Origem | Volume Estimado |
|-----------|-----------------|------------|--------|-----------------|
| **Identificacao** | CPF (11 digitos) | Chave PIX | Usuario final | 300.000+ |
| **Identificacao** | CNPJ (14 digitos) | Chave PIX | Usuario final | 50.000+ |
| **Identificacao** | Nome completo | Titular da conta | Usuario final | 350.000+ |
| **Contato** | Telefone celular (E.164) | Chave PIX | Usuario final | 200.000+ |
| **Contato** | Email | Chave PIX | Usuario final | 150.000+ |
| **Financeiros** | ISPB (8 digitos) | Conta vinculada a chave | Sistema interno LBPay | 350.000+ |
| **Financeiros** | Agencia bancaria | Conta vinculada a chave | Sistema interno LBPay | 350.000+ |
| **Financeiros** | Numero de conta | Conta vinculada a chave | Sistema interno LBPay | 350.000+ |
| **Transacionais** | Historico de claims | Auditoria e compliance | Sistema DICT | 10.000+ |
| **Transacionais** | Historico de portabilidades | Auditoria e compliance | Sistema DICT | 5.000+ |
| **Auditoria** | Logs de acesso (usuario, IP, timestamp) | Seguranca e compliance | Sistema DICT | 50.000.000+ eventos |
| **Tecnicos** | Endereco IP | Rate limiting, seguranca | Sistema DICT | 10.000.000+ |
| **Tecnicos** | User Agent | Deteccao de fraudes | Sistema DICT | 10.000.000+ |

**Total de Titulares**: 350.000+ (estimativa ano 1)

**Total de Registros de Dados Pessoais**: 500.000+ chaves PIX

---

### 3.2 Dados Sensiveis (LGPD Art. 5o, II)

**Nota Importante**: O sistema DICT **NAO trata dados sensiveis** conforme definicao da LGPD.

**Dados sensiveis (NAO tratados no DICT)**:
- Origem racial ou etnica
- Conviccao religiosa
- Opiniao politica
- Filiacao sindical
- Dados geneticos
- Dados biometricos
- Dados de saude
- Dados sobre orientacao sexual

**Dados financeiros (CPF, CNPJ, conta bancaria)** sao considerados **dados pessoais comuns** (nao sensiveis) pela LGPD.

---

## 4. Mapeamento de Fluxo de Dados

### 4.1 Diagrama de Fluxo de Dados

```
[Usuario Final]
    |
    | (1) Cadastro de Chave PIX (HTTPS)
    v
[LBPay App/Web]
    |
    | (2) API REST (TLS 1.2+)
    v
[Core DICT]
    |
    | (3) Validacao + Persistencia
    v
[PostgreSQL] (criptografia em repouso)
    |
    | (4) Envio ao DICT Bacen (via Bridge + Connect)
    v
[Bridge] (workflow assincrono)
    |
    | (5) gRPC (TLS 1.2+)
    v
[RSFN Connect]
    |
    | (6) SOAP/XML (mTLS, ICP-Brasil)
    v
[DICT Bacen] (Banco Central do Brasil)
```

---

### 4.2 Locais de Armazenamento

| Sistema | Tipo de Dado | Localizacao | Criptografia |
|---------|--------------|-------------|--------------|
| **PostgreSQL (Core)** | Chaves PIX, usuarios, claims | AWS RDS (us-east-1) | TDE (Transparent Data Encryption) |
| **Redis (Cache)** | Cache de consultas | AWS ElastiCache | Encryption at rest |
| **S3 (Backups)** | Backups de banco de dados | AWS S3 (us-east-1) | AES-256 |
| **S3 Glacier (Archival)** | Logs de auditoria antigos (> 1 ano) | AWS S3 Glacier (us-east-1) | AES-256 |
| **CloudWatch Logs** | Logs de aplicacao | AWS CloudWatch | Criptografia AWS |
| **Prometheus** | Metricas | Servidor interno | N/A (nao armazena PII) |

**Observacao**: Todos os dados estao armazenados em **regiao AWS us-east-1** (Virginia, EUA). Conforme LGPD Art. 33, transferencia internacional de dados para EUA e permitida se houver **clausulas contratuais adequadas** (AWS Data Processing Addendum).

---

### 4.3 Compartilhamento de Dados

| Destinatario | Dados Compartilhados | Finalidade | Base Legal | Contrato |
|--------------|----------------------|------------|------------|----------|
| **Banco Central (DICT)** | Chaves PIX, CPF/CNPJ, nome, conta | Registro no DICT centralizado | Obrigacao legal | N/A (orgao regulador) |
| **AWS (Infraestrutura)** | Todos os dados (armazenamento) | Hosting, infraestrutura | Execucao de contrato | AWS DPA assinado |

**Nota**: LBPay **NAO compartilha dados com terceiros** para marketing ou outras finalidades nao relacionadas ao DICT.

---

## 5. Avaliacao de Riscos

### 5.1 Metodologia de Avaliacao

**Escala de Impacto**:
- **BAIXO**: Impacto minimo ao titular (ex: inconveniencia temporaria)
- **MEDIO**: Impacto moderado (ex: constrangimento, perda financeira pequena)
- **ALTO**: Impacto significativo (ex: perda financeira grande, discriminacao)
- **CRITICO**: Impacto severo (ex: risco a vida, danos irreversiveis)

**Escala de Probabilidade**:
- **RARA**: < 5% de chance de ocorrencia
- **BAIXA**: 5% - 25%
- **MEDIA**: 25% - 50%
- **ALTA**: > 50%

**Risco = Impacto x Probabilidade**

---

### 5.2 Riscos Identificados

#### Risco 1: Vazamento de Dados Pessoais (Data Breach)

**Descricao**: Acesso nao autorizado a base de dados (PostgreSQL) resulta em vazamento de chaves PIX, CPF, nomes, contas bancarias.

**Impacto**: **ALTO**
- Exposicao de dados financeiros sensiveis
- Risco de fraude financeira
- Dano reputacional a LBPay
- Penalidades LGPD (ate R$ 50 milhoes)

**Probabilidade SEM Controles**: **MEDIA** (25-50%)
- Ataques de hackers a instituicoes financeiras sao comuns
- Dados financeiros sao alvo de alto valor

**Controles Implementados**:
1. **Criptografia em repouso** (PostgreSQL TDE)
2. **Criptografia em transito** (TLS 1.2+)
3. **Controle de acesso** (RBAC - apenas usuarios autorizados)
4. **Auditoria** (logs de todos os acessos)
5. **Firewall de banco de dados** (AWS Security Groups)
6. **Masking de dados em logs**
7. **Monitoramento 24/7** (alertas de atividades suspeitas)
8. **Penetration Testing** (anual)

**Probabilidade COM Controles**: **RARA** (< 5%)

**Risco Residual**: **BAIXO** (Impacto ALTO x Probabilidade RARA)

**Responsavel**: Security Team + DBA

---

#### Risco 2: Acesso Interno Nao Autorizado (Insider Threat)

**Descricao**: Funcionario da LBPay acessa dados pessoais de clientes sem autorizacao ou para fins nao relacionados ao trabalho.

**Impacto**: **MEDIO**
- Violacao de privacidade
- Perda de confianca do cliente
- Penalidades LGPD

**Probabilidade SEM Controles**: **BAIXA** (5-25%)

**Controles Implementados**:
1. **RBAC** (acesso minimo necessario - Principle of Least Privilege)
2. **Logs de auditoria** (rastreamento de quem acessou o que)
3. **Alertas de acesso suspeito** (ex: acesso a grande volume de dados)
4. **Treinamento LGPD** (anual para todos os funcionarios)
5. **Background check** (verificacao de antecedentes em contratacao)
6. **NDAs assinados** (confidencialidade)

**Probabilidade COM Controles**: **RARA** (< 5%)

**Risco Residual**: **BAIXO** (Impacto MEDIO x Probabilidade RARA)

**Responsavel**: HR + Security Team

---

#### Risco 3: Perda de Dados (Data Loss)

**Descricao**: Falha de hardware, desastre natural ou erro humano resulta em perda de dados pessoais.

**Impacto**: **ALTO**
- Perda de chaves PIX dos clientes
- Impossibilidade de realizar transacoes PIX
- Violacao de obrigacao contratual

**Probabilidade SEM Controles**: **MEDIA** (25-50%)

**Controles Implementados**:
1. **Backups automaticos diarios** (PostgreSQL)
2. **Backups em multiplas regioes AWS** (cross-region)
3. **Testes de restore** (mensal)
4. **Replica de banco de dados** (standby)
5. **Multi-AZ deployment** (AWS)
6. **Disaster Recovery Plan** documentado

**Probabilidade COM Controles**: **RARA** (< 5%)

**Risco Residual**: **BAIXO** (Impacto ALTO x Probabilidade RARA)

**Responsavel**: DBA + Infra Team

---

#### Risco 4: Nao Conformidade com LGPD (Direitos dos Titulares)

**Descricao**: LBPay nao atende solicitacoes de direitos dos titulares (acesso, correcao, exclusao) dentro do prazo legal (15 dias uteis).

**Impacto**: **MEDIO**
- Penalidades ANPD
- Dano reputacional

**Probabilidade SEM Controles**: **MEDIA** (25-50%)

**Controles Implementados**:
1. **API de DSAR** (Data Subject Access Request)
2. **Funcionalidade de correcao de dados** no app
3. **Funcionalidade de exclusao de chave PIX** no app
4. **Processo de atendimento documentado** (SLA de 15 dias)
5. **Treinamento de equipe de suporte**
6. **DPO nomeado** (responsavel por conformidade)

**Probabilidade COM Controles**: **RARA** (< 5%)

**Risco Residual**: **BAIXO** (Impacto MEDIO x Probabilidade RARA)

**Responsavel**: DPO + Support Team

---

#### Risco 5: Incidente de Seguranca Nao Detectado/Notificado

**Descricao**: Incidente de seguranca (ex: vazamento de dados) ocorre mas nao e detectado em tempo habil ou nao e notificado a ANPD em 72 horas.

**Impacto**: **ALTO**
- Penalidades LGPD agravadas (por nao notificacao)
- Dano reputacional maior
- Impossibilidade de mitigar danos aos titulares

**Probabilidade SEM Controles**: **MEDIA** (25-50%)

**Controles Implementados**:
1. **Monitoramento 24/7** (Prometheus, Grafana)
2. **Alertas de anomalias de seguranca**
3. **Plano de Resposta a Incidentes** (IRP) documentado
4. **Equipe de resposta a incidentes** (Security Team)
5. **Template de notificacao a ANPD** preparado
6. **Simulacao de incidente** (anual)

**Probabilidade COM Controles**: **RARA** (< 5%)

**Risco Residual**: **BAIXO** (Impacto ALTO x Probabilidade RARA)

**Responsavel**: Security Team + DPO

---

#### Risco 6: Uso Indevido de Dados para Perfilamento/Discriminacao

**Descricao**: Dados de chaves PIX sao usados para perfilamento de clientes ou discriminacao (ex: negar servico com base em dados pessoais).

**Impacto**: **ALTO**
- Violacao do Principio de Nao Discriminacao (LGPD Art. 6o, IX)
- Penalidades LGPD
- Dano reputacional severo

**Probabilidade SEM Controles**: **BAIXA** (5-25%)

**Controles Implementados**:
1. **Politica contra discriminacao** documentada
2. **Base legal clara** (execucao de contrato - nao ha perfilamento)
3. **Treinamento LGPD** (conscientizacao sobre nao discriminacao)
4. **Auditoria de algoritmos** (se aplicavel - nao ha algoritmos de perfilamento no DICT)

**Probabilidade COM Controles**: **RARA** (< 5%)

**Risco Residual**: **BAIXO** (Impacto ALTO x Probabilidade RARA)

**Responsavel**: DPO + Legal + Compliance

---

### 5.3 Matriz de Riscos (Resumo)

| Risco | Impacto | Probabilidade (SEM Controles) | Probabilidade (COM Controles) | Risco Residual |
|-------|---------|-------------------------------|-------------------------------|----------------|
| **1. Vazamento de Dados** | ALTO | MEDIA | RARA | BAIXO |
| **2. Acesso Interno Nao Autorizado** | MEDIO | BAIXA | RARA | BAIXO |
| **3. Perda de Dados** | ALTO | MEDIA | RARA | BAIXO |
| **4. Nao Conformidade LGPD** | MEDIO | MEDIA | RARA | BAIXO |
| **5. Incidente Nao Detectado** | ALTO | MEDIA | RARA | BAIXO |
| **6. Uso Indevido de Dados** | ALTO | BAIXA | RARA | BAIXO |

**Conclusao**: Todos os riscos residuais sao **BAIXO** apos implementacao completa dos controles.

---

## 6. Medidas de Mitigacao Implementadas

### 6.1 Medidas Tecnicas

| Medida | Status | Descricao | Responsavel |
|--------|--------|-----------|-------------|
| **Criptografia em Transito (TLS 1.2+)** | [x] | Todas as conexoes HTTPS/gRPC/mTLS | Infra Team |
| **Criptografia em Repouso (TDE)** | [x] | PostgreSQL Transparent Data Encryption | DBA + Security |
| **Criptografia de Backups (AES-256)** | [x] | Backups S3 criptografados | Infra + DBA |
| **RBAC (Role-Based Access Control)** | [x] | Controle de acesso por role (admin, user, support) | Backend Team |
| **Masking de Dados em Logs** | [x] | Dados sensiveis mascarados (ex: CPF -> ***8900) | Backend Team |
| **Auditoria (Audit Trail)** | [x] | Logs de todas as operacoes CRUD | Backend + DBA |
| **Retencao de Logs (5 anos)** | [x] | Conforme Bacen (Circular 3.682) | DBA + Compliance |
| **Monitoramento 24/7** | [x] | Prometheus + Grafana + AlertManager | DevOps Team |
| **Firewall de Banco de Dados** | [x] | AWS Security Groups | Infra Team |
| **Rate Limiting** | [x] | Protecao contra brute force | Backend + Infra |
| **MFA (Multi-Factor Authentication)** | [ ] | Para usuarios admin e support | Backend Team |
| **Penetration Testing** | [ ] | Anual | Security Team |

---

### 6.2 Medidas Organizacionais

| Medida | Status | Descricao | Responsavel |
|--------|--------|-----------|-------------|
| **DPO Nomeado** | [x] | Data Protection Officer designado | CEO + DPO |
| **Politica de Privacidade** | [x] | Publicada e acessivel | Legal + DPO |
| **ROPA (Registro de Operacoes de Tratamento)** | [x] | Mantido e atualizado | DPO |
| **DPIA (Este Documento)** | [x] | Avaliacao de impacto realizada | DPO + Security |
| **Treinamento LGPD** | [ ] | Anual para todos os funcionarios | HR + DPO |
| **Plano de Resposta a Incidentes (IRP)** | [ ] | Documentado e testado | Security Team |
| **Contratos com Processadores (AWS)** | [x] | AWS DPA assinado | Legal + Procurement |
| **Background Check (Funcionarios)** | [ ] | Verificacao de antecedentes | HR |
| **NDAs Assinados** | [x] | Confidencialidade | HR + Legal |

---

### 6.3 Medidas de Conformidade

| Medida | Status | Descricao | Responsavel |
|--------|--------|-----------|-------------|
| **API de DSAR** | [ ] | Acesso aos dados pelo titular | Backend Team |
| **Funcionalidade de Correcao** | [ ] | Titular pode corrigir dados | Backend + Product |
| **Funcionalidade de Exclusao** | [x] | Titular pode deletar chave PIX | Backend + Product |
| **Portabilidade (Export JSON/CSV)** | [ ] | Titular pode exportar dados | Backend Team |
| **Processo de Atendimento DSAR** | [ ] | SLA de 15 dias uteis | DPO + Support |

---

## 7. Plano de Acao (Medidas Pendentes)

### 7.1 Prioridade Alta (Implementar ate 3 meses apos lancamento)

| ID | Medida | Prazo | Responsavel | Status |
|----|--------|-------|-------------|--------|
| PA-01 | Implementar MFA para admin/support | T+3 meses | Backend Team | [ ] |
| PA-02 | Implementar API de DSAR | T+3 meses | Backend Team | [ ] |
| PA-03 | Implementar funcionalidade de correcao de dados | T+3 meses | Backend + Product | [ ] |
| PA-04 | Treinamento LGPD para todos os funcionarios | T+3 meses | HR + DPO | [ ] |
| PA-05 | Processo de atendimento DSAR documentado | T+3 meses | DPO + Support | [ ] |

---

### 7.2 Prioridade Media (Implementar ate 6 meses apos lancamento)

| ID | Medida | Prazo | Responsavel | Status |
|----|--------|-------|-------------|--------|
| PM-01 | Implementar portabilidade de dados (export) | T+6 meses | Backend Team | [ ] |
| PM-02 | Plano de Resposta a Incidentes documentado | T+6 meses | Security Team | [ ] |
| PM-03 | Background check para novos funcionarios | T+6 meses | HR | [ ] |
| PM-04 | Penetration Testing (primeiro teste) | T+6 meses | Security Team | [ ] |

---

### 7.3 Prioridade Baixa (Implementar ate 12 meses apos lancamento)

| ID | Medida | Prazo | Responsavel | Status |
|----|--------|-------|-------------|--------|
| PB-01 | Simulacao de incidente de seguranca | T+12 meses | Security Team | [ ] |
| PB-02 | Auditoria externa de conformidade LGPD | T+12 meses | DPO + Compliance | [ ] |

---

## 8. Consulta a Partes Interessadas

### 8.1 Stakeholders Consultados

Durante a elaboracao deste DPIA, foram consultados:

| Stakeholder | Papel | Contribuicao |
|-------------|-------|--------------|
| **DPO (Data Protection Officer)** | Responsavel por privacidade | Avaliacao de conformidade LGPD |
| **Security Team** | Responsavel por seguranca | Avaliacao de riscos tecnicos |
| **Architect (Thiago Lima)** | Arquiteto de Sistemas | Mapeamento de fluxo de dados |
| **Legal Team** | Juridico | Validacao de base legal |
| **Head de Produto (Luiz Sant'Ana)** | Produto | Validacao de finalidades |
| **CTO (Jose Luis Silva)** | Tecnologia | Aprovacao de medidas tecnicas |
| **DBA** | Banco de Dados | Medidas de seguranca de dados |

---

### 8.2 Consulta aos Titulares (Usuarios Finais)

**Mecanismo de Consulta**: Pesquisa de privacidade (opcional, via app)

**Pergunta**: "Voce se sente confortavel com o nivel de protecao de seus dados pessoais no sistema DICT da LBPay?"

**Resultado** (se aplicavel):
- [Aguardando pesquisa pos-lancamento]

---

## 9. Revisao e Atualizacao do DPIA

### 9.1 Gatilhos para Revisao

Este DPIA DEVE ser revisado quando:

1. **Mudanca significativa no tratamento de dados** (ex: nova finalidade, novos dados coletados)
2. **Incidente de seguranca** que afete dados pessoais
3. **Mudanca na legislacao** (LGPD, Bacen)
4. **Anualmente** (revisao periodica)
5. **Auditoria externa** identifica novos riscos

---

### 9.2 Historico de Revisoes

| Versao | Data | Autor | Mudancas |
|--------|------|-------|----------|
| 1.0 | 2025-10-25 | DPO + Security Team | Versao inicial |

---

## 10. Conclusao

### 10.1 Resumo da Avaliacao

O sistema DICT da LBPay trata **dados pessoais de alto volume** (350.000+ titulares) e **dados financeiros sensiveis** (chaves PIX, contas bancarias). No entanto, apos implementacao de **medidas tecnicas e organizacionais robustas**, o **risco residual e BAIXO** para todos os riscos identificados.

**Principais Medidas de Mitigacao**:
- Criptografia em transito e repouso
- RBAC (controle de acesso)
- Auditoria completa (logs de 5 anos)
- Monitoramento 24/7
- DPO nomeado
- Politica de Privacidade publicada
- Processos de atendimento a direitos dos titulares

---

### 10.2 Recomendacao

**Recomendacao**: **APROVADO** para processamento de dados pessoais, condicionado a:

1. Implementacao completa de medidas tecnicas (P0 - antes do lancamento)
2. Implementacao de medidas organizacionais prioritarias (P1 - ate 3 meses apos lancamento)
3. Monitoramento continuo de riscos
4. Revisao anual deste DPIA

---

### 10.3 Aprovacao

| Papel | Nome | Assinatura | Data |
|-------|------|------------|------|
| **DPO** | [Nome do DPO] | [Aguardando] | [Aguardando] |
| **CTO** | Jose Luis Silva | [Aguardando] | [Aguardando] |
| **Head de Produto** | Luiz Sant'Ana | [Aguardando] | [Aguardando] |
| **Legal** | [Nome do Legal] | [Aguardando] | [Aguardando] |

---

## 11. Referencias

### Internas

- [SEC-007: LGPD Data Protection](../13_Seguranca/SEC-007_LGPD_Data_Protection.md)
- [CMP-001: Audit Logs Specification](CMP-001_Audit_Logs_Specification.md)
- [CMP-002: LGPD Compliance Checklist](CMP-002_LGPD_Compliance_Checklist.md)
- [CMP-004: Data Retention Policy](CMP-004_Data_Retention_Policy.md)
- [REG-001: Requisitos Regulatorios Bacen](../06_Regulatorio/REG-001_Requisitos_Regulatorios_Bacen.md)

### Externas

- Lei 13.709/2018 (LGPD) - Art. 38 (DPIA obrigatorio)
- GDPR Art. 35 (Data Protection Impact Assessment)
- ISO 29134:2017 (Privacy Impact Assessment Guidelines)
- Guia de DPIA - ANPD: https://www.gov.br/anpd/

---

**Versao**: 1.0
**Status**: Relatorio de Impacto a Protecao de Dados
**Proxima Revisao**: Anual (2026-10-25) ou apos gatilhos de revisao
