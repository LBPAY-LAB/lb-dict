# CMP-003: Bacen Regulatory Compliance

**Projeto**: DICT - Diretorio de Identificadores de Contas Transacionais (LBPay)
**Versao**: 1.0
**Data**: 2025-10-25
**Status**: Checklist Regulatorio Bacen
**Responsavel**: Compliance Team + Head de Produto + CTO

---

## 1. Resumo Executivo

Este documento apresenta **checklist completo de conformidade com requisitos regulatorios do Banco Central** para participacao no DICT/SPI (Sistema de Pagamentos Instantaneos).

**Base Legal**:
- Manual Operacional DICT v8 (Banco Central do Brasil)
- Instrucao Normativa BCB no 508/2024 (Homologacao DICT)
- Resolucao BCB no 1/2020 (Regulamento PIX)
- Circular 3.682/2013 (Auditoria de sistemas)

**Objetivo**: Garantir 100% de conformidade com exigencias do Bacen para certificacao e operacao do DICT.

---

## 2. Requisitos de Adesao ao DICT

### 2.1 Certificacao Digital ICP-Brasil

**Requisito**: REG-001 (Manual Operacional DICT, Secao "Interface de Comunicacao")

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| Certificado digital ICP-Brasil e-CNPJ valido | [ ] | Certificado instalado | Infra Team | P0 |
| Certificado instalado para mTLS (autenticacao mutua) | [ ] | Configuracao mTLS validada | Security Team | P0 |
| Certificado emitido por AC credenciada ICP-Brasil | [ ] | Cadeia de certificacao valida | Security Team | P0 |
| Rotacao automatica antes do vencimento (90 dias) | [ ] | Script de renovacao configurado | Infra Team | P0 |
| Suporte a CRL (Certificate Revocation List) | [ ] | Validacao de revogacao ativa | Security Team | P0 |

**Penalidade**: Impossibilidade de comunicacao com DICT. Bloqueio de acesso ao ambiente de homologacao e producao.

**Prazo de Implementacao**: P0 (critico - antes da homologacao)

---

### 2.2 Conectividade RSFN

**Requisito**: REG-002 (Manual Operacional DICT, Secao 11 "Interface de Comunicacao")

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| Conectividade fisica/logica com RSFN estabelecida | [ ] | Contrato com provedor RSFN | Infra Team | P0 |
| Protocolos SOAP/XML sobre HTTPS configurados | [ ] | Testes de conectividade | RSFN Connect Team | P0 |
| Latencia de rede <= 50ms (percentil 95) ate endpoints Bacen | [ ] | Relatorio de latencia | Infra + DevOps | P0 |
| Redundancia de links (minimo 2 links independentes) | [ ] | Configuracao de redundancia | Infra Team | P0 |

**Penalidade**: Impossibilidade de participar do DICT. Falha na homologacao.

**Prazo de Implementacao**: P0 (critico - antes da homologacao)

---

### 2.3 Cadastramento no DICT

**Requisito**: REG-003 (IN BCB 508/2024, Art. 7o)

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| ISPB da LBPay cadastrado no ambiente de homologacao DICT | [ ] | Confirmacao do Bacen | Head de Produto | P0 |
| ISPB da LBPay cadastrado no ambiente de producao DICT | [ ] | Confirmacao do Bacen (apos homologacao) | Head de Produto | P1 |
| Dados institucionais completos e atualizados no cadastro Bacen | [ ] | Validacao de cadastro | Compliance | P0 |
| Contatos tecnicos e operacionais registrados | [ ] | Lista de contatos enviada ao Bacen | Head de Produto | P0 |

**Penalidade**: Rejeicao de todas as operacoes no DICT. Impossibilidade de registrar chaves PIX.

**Prazo de Implementacao**: P0 (critico - antes da homologacao)

---

## 3. Requisitos de Homologacao (IN BCB 508/2024)

### 3.1 Preparacao para Homologacao

**Requisito**: REG-004, REG-005, REG-006 (IN BCB 508/2024, Art. 12)

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| 1.000 chaves PIX registradas (mesmo tipo: CPF, CNPJ, Email ou Telefone) | [ ] | Relatorio de chaves cadastradas | Backend Team | P0 |
| Tipo de chave informado ao Bacen via email | [ ] | Email enviado para pix-operacional@bcb.gov.br | Head de Produto | P0 |
| 5 transacoes PIX realizadas com participante virtual recebedor 99999004 | [ ] | EndToEndIds das transacoes | Payments Team | P0 |
| EndToEndIds informados ao Bacen via email | [ ] | Email enviado | Head de Produto | P0 |
| Aptidao para receber transacoes do virtual pagador 99999003 | [ ] | Testes de recebimento | Payments Team | P0 |

**Penalidade**: Impossibilidade de agendar testes de homologacao. Atraso no cronograma de certificacao.

**Prazo de Implementacao**: P0 (critico - antes da homologacao)

---

### 3.2 Agendamento de Testes

**Requisito**: REG-008, REG-009, REG-010 (IN BCB 508/2024, Art. 9o, 10o, 11o, 12o)

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| Pedido de agendamento via Protocolo Digital do Bacen | [ ] | Numero de protocolo | Head de Produto | P0 |
| Email ao DECEM com: Tipo de chave, EndToEndIds, Sugestao de data/horario | [ ] | Email enviado | Head de Produto | P0 |
| Confirmacao de agendamento recebida do DECEM | [ ] | Email de confirmacao | Head de Produto | P0 |
| Equipe tecnica mobilizada para a data agendada | [ ] | Lista de participantes | CTO + Head de Produto | P0 |

**Penalidade**: Impossibilidade de realizar homologacao. Atraso no cronograma.

**Prazo de Implementacao**: P0 (critico - antes da homologacao)

---

### 3.3 Execucao dos Testes (Janela de 1 Hora)

**Requisito**: REG-011 a REG-018 (IN BCB 508/2024, Art. 14, 16)

#### Teste 1: Registro de Chaves PIX (Todos os Tipos)

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| Registrar 1 chave CPF | [ ] | Entry ID + timestamp | Backend Team | P0 |
| Registrar 1 chave CNPJ | [ ] | Entry ID + timestamp | Backend Team | P0 |
| Registrar 1 chave Email | [ ] | Entry ID + timestamp | Backend Team | P0 |
| Registrar 1 chave Telefone | [ ] | Entry ID + timestamp | Backend Team | P0 |
| Registrar 1 chave Aleatoria (EVP) | [ ] | Entry ID + timestamp | Backend Team | P0 |
| Tempo total <= 10 minutos | [ ] | Log de execucao | DevOps Team | P0 |

**Penalidade**: Reprovacao na homologacao.

---

#### Teste 2: Consulta a Chaves PIX (Todos os Tipos)

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| Consultar 1 chave CPF | [ ] | Dados retornados corretos | Backend Team | P0 |
| Consultar 1 chave CNPJ | [ ] | Dados retornados corretos | Backend Team | P0 |
| Consultar 1 chave Email | [ ] | Dados retornados corretos | Backend Team | P0 |
| Consultar 1 chave Telefone | [ ] | Dados retornados corretos | Backend Team | P0 |
| Consultar 1 chave EVP | [ ] | Dados retornados corretos | Backend Team | P0 |
| Consultas adicionais (se solicitadas pelo Bacen) | [ ] | Resultados corretos | Backend Team | P0 |
| Tempo total <= 5 minutos | [ ] | Log de execucao | DevOps Team | P0 |

**Penalidade**: Reprovacao na homologacao.

---

#### Teste 3: Verificacao de Sincronismo (VSYNC)

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| VSYNC executado para tipo de chave da preparacao (1.000 chaves) | [ ] | Relatorio de VSYNC | Backend + Bridge Team | P0 |
| Identificacao de chaves modificadas/inseridas pelo Bacen | [ ] | Lista de diferencas | Backend Team | P0 |
| Sincronizacao completa das diferencas | [ ] | Confirmacao de sincronizacao | Backend Team | P0 |
| Tempo total <= 15 minutos | [ ] | Log de execucao | DevOps Team | P0 |

**Penalidade**: Reprovacao na homologacao. Impossibilidade de garantir consistencia DICT <-> Base Local.

---

#### Teste 4: Recebimento de Reivindicacoes (< 1 minuto)

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| TODAS as claims criadas pelo Bacen recebidas | [ ] | Lista de claims recebidas | Bridge Team | P0 |
| Recebimento confirmado em <= 60 segundos para CADA claim | [ ] | Timestamps de recebimento | Bridge Team | P0 |
| Sistema notifica usuario final (se aplicavel) | [ ] | Logs de notificacao | Backend Team | P0 |
| Claims processadas e persistidas corretamente | [ ] | Validacao no banco de dados | Backend Team | P0 |

**Penalidade**: Reprovacao na homologacao. SLA violado -> penalidades operacionais pos-producao.

**CRITICO**: Este e o teste mais critico da homologacao.

---

#### Teste 5: Fluxo de Reivindicacao Completo (Reivindicador)

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| Criar 1 portabilidade OU 1 reivindicacao de posse | [ ] | Claim ID | Bridge Team | P0 |
| Confirmar a reivindicacao criada | [ ] | Timestamp de confirmacao | Bridge Team | P0 |
| Completar a reivindicacao confirmada | [ ] | Timestamp de completacao | Bridge Team | P0 |
| Cancelar 1 reivindicacao | [ ] | Timestamp de cancelamento | Bridge Team | P0 |
| Tempo total <= 15 minutos | [ ] | Log de execucao | DevOps Team | P0 |

**Penalidade**: Reprovacao na homologacao. Impossibilidade de oferecer portabilidade aos clientes.

---

#### Teste 6: Fluxo de Notificacao de Infracao Completo

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| Criar 1 notificacao de infracao | [ ] | Infraction ID | Backend + Bridge | P0 |
| Confirmar a notificacao criada | [ ] | Timestamp de confirmacao | Bridge Team | P0 |
| Completar a notificacao confirmada | [ ] | Timestamp de completacao | Bridge Team | P0 |
| Cancelar 1 notificacao | [ ] | Timestamp de cancelamento | Bridge Team | P0 |
| Tempo total <= 10 minutos | [ ] | Log de execucao | DevOps Team | P0 |

**Penalidade**: Reprovacao na homologacao. Impossibilidade de reportar/receber notificacoes de fraude.

---

#### Teste 7: Fluxo de Solicitacao de Devolucao (2 Motivos)

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| Criar 1 devolucao por falha operacional | [ ] | Refund ID | Payments Team | P0 |
| Criar 1 devolucao por fundada suspeita de fraude | [ ] | Refund ID | Payments Team | P0 |
| Completar devolucao por falha operacional (do virtual 99999003) | [ ] | Timestamp de completacao | Payments Team | P0 |
| Completar devolucao por fraude (do virtual 99999003) | [ ] | Timestamp de completacao | Payments Team | P0 |
| Tempo total <= 10 minutos | [ ] | Log de execucao | DevOps Team | P0 |

**Penalidade**: Reprovacao na homologacao. Impossibilidade de gerenciar devolucoes PIX.

---

### 3.4 Ausencia de Pendencias Pre-Homologacao

**Requisito**: REG-007 (IN BCB 508/2024, Art. 15)

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| Zero portabilidades pendentes (status REQUESTED, CONFIRMED) | [ ] | Query ao banco de dados | Backend Team | P0 |
| Zero reivindicacoes pendentes (status WAITING_RESOLUTION, CONFIRMED) | [ ] | Query ao banco de dados | Backend Team | P0 |
| Zero notificacoes de infracao pendentes | [ ] | Query ao banco de dados | Backend Team | P0 |
| Verificacao automatizada antes do inicio dos testes | [ ] | Script de pre-check | DevOps Team | P0 |

**Penalidade**: Falha nos testes de homologacao. Impossibilidade de iniciar janela de 1 hora de testes.

---

## 4. Requisitos de Disponibilidade (SLA)

### 4.1 SLA de Disponibilidade

**Requisito**: Manual Operacional DICT, Secao "Disponibilidade"

| Metrica | Requisito Bacen | Status | Evidencia | Responsavel |
|---------|-----------------|--------|-----------|-------------|
| **Disponibilidade Mensal** | >= 99.9% | [ ] | Relatorio de uptime (Grafana) | DevOps Team |
| **Downtime Maximo Mensal** | <= 43.2 minutos/mes | [ ] | Relatorio de downtime | DevOps Team |
| **Downtime Maximo por Incidente** | <= 30 minutos | [ ] | Logs de incidentes | DevOps Team |
| **RTO (Recovery Time Objective)** | <= 15 minutos | [ ] | Testes de recuperacao | DevOps + Infra |
| **RPO (Recovery Point Objective)** | <= 5 minutos | [ ] | Testes de backup/restore | DBA + Infra |

**Penalidade**: Multa proporcional ao tempo de indisponibilidade. Suspensao temporaria em casos graves.

**Prazo de Implementacao**: P0 (critico - antes da producao)

---

### 4.2 Monitoramento e Alertas

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| Monitoramento 24/7 implementado (Prometheus + Grafana) | [ ] | Dashboards ativos | DevOps Team | P0 |
| Alertas de indisponibilidade configurados | [ ] | Configuracao AlertManager | DevOps Team | P0 |
| Escalation automatica para on-call team | [ ] | Integracao PagerDuty/OpsGenie | DevOps Team | P0 |
| Postmortem obrigatorio para incidentes > 5 minutos | [ ] | Processo de postmortem | CTO + DevOps | P1 |

**Prazo de Implementacao**: P0 (critico - antes da producao)

---

## 5. Requisitos de Seguranca

### 5.1 Autenticacao e Autorizacao

**Requisito**: Manual Operacional DICT, Secao "Seguranca"

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| mTLS (autenticacao mutua) com DICT Bacen | [ ] | Configuracao mTLS validada | Security Team | P0 |
| Certificado ICP-Brasil A3 (hardware) | [ ] | Certificado instalado | Security Team | P0 |
| Validacao de certificado do servidor Bacen | [ ] | Teste de validacao | Security Team | P0 |
| Revogacao de certificados (CRL/OCSP) | [ ] | Configuracao de revogacao | Security Team | P0 |

**Penalidade**: Suspensao imediata do acesso ao DICT.

**Prazo de Implementacao**: P0 (critico - antes da homologacao)

---

### 5.2 Criptografia

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| TLS 1.2+ em todas as conexoes externas | [ ] | Configuracao de servidores | Infra Team | P0 |
| TLS 1.3 preferencial (se suportado pelo Bacen) | [ ] | Configuracao de servidores | Infra Team | P1 |
| Criptografia em repouso (PostgreSQL TDE) | [ ] | Configuracao de banco de dados | DBA + Security | P0 |
| Criptografia de backups (AES-256) | [ ] | Configuracao de backup | Infra + DBA | P0 |

**Prazo de Implementacao**: P0 (critico - antes da producao)

---

## 6. Requisitos de Auditoria (Circular 3.682/2013)

### 6.1 Logs de Auditoria

**Requisito**: Circular 3.682/2013, Art. 4o

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| Logs de auditoria de TODAS as operacoes DICT | [ ] | Tabelas audit.* no PostgreSQL | Backend + DBA | P0 |
| Retencao de logs: 5 anos | [ ] | Politica de retencao implementada | DBA + Compliance | P0 |
| Logs incluem: Timestamp, User ID, IP, Operacao, Resultado | [ ] | Validacao de schema de logs | Backend Team | P0 |
| Masking de dados sensiveis em logs | [ ] | Codigo de masking validado | Backend Team | P0 |
| Archival para S3 Glacier (apos 1 ano) | [ ] | Pipeline de archival configurado | Infra Team | P0 |

**Penalidade**: Multa por nao conformidade. Impossibilidade de comprovar operacoes em auditoria.

**Prazo de Implementacao**: P0 (critico - antes da producao)

---

### 6.2 Rastreabilidade

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| Correlation IDs em todas as operacoes | [ ] | Validacao de logs | Backend Team | P0 |
| Request IDs do Bacen armazenados | [ ] | Campo bacen_request_id em logs | Backend Team | P0 |
| Rastreamento end-to-end (Core -> Bridge -> Connect -> Bacen) | [ ] | Dashboard de rastreamento | DevOps Team | P0 |

**Prazo de Implementacao**: P0 (critico - antes da producao)

---

## 7. Requisitos de Validacao de Dados

### 7.1 Validacao de Chaves PIX

**Requisito**: REG-021 a REG-027 (Manual Operacional DICT)

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| Validacao de CPF (11 digitos + algoritmo de digito verificador) | [ ] | Codigo de validacao + testes | Backend Team | P0 |
| Validacao de CNPJ (14 digitos + algoritmo de digito verificador) | [ ] | Codigo de validacao + testes | Backend Team | P0 |
| Validacao de Email (formato RFC 5322, <= 77 caracteres) | [ ] | Codigo de validacao + testes | Backend Team | P0 |
| Validacao de Telefone (formato E.164, +5511XXXXXXXXX) | [ ] | Codigo de validacao + testes | Backend Team | P0 |
| Validacao de EVP (UUID v4, RFC 4122) | [ ] | Codigo de validacao + testes | Backend Team | P0 |

**Penalidade**: Rejeicao pelo DICT Bacen. Dados inconsistentes.

**Prazo de Implementacao**: P0 (critico - antes da homologacao)

---

### 7.2 Validacao de Limites

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| Limite de 5 chaves por conta (CPF) | [ ] | Codigo de validacao + testes | Backend Team | P0 |
| Limite de 20 chaves por conta (CNPJ) | [ ] | Codigo de validacao + testes | Backend Team | P0 |
| Rejeicao de tentativa de exceder limite | [ ] | Teste de validacao | Backend Team | P0 |

**Penalidade**: Rejeicao pelo DICT Bacen (HTTP 400 - Limit Exceeded).

**Prazo de Implementacao**: P0 (critico - antes da homologacao)

---

## 8. Requisitos de Retencao de Dados

### 8.1 Politica de Retencao

**Requisito**: Manual Operacional DICT + Circular 3.682/2013

| Tipo de Dado | Retencao Minima | Status | Evidencia | Responsavel |
|--------------|-----------------|--------|-----------|-------------|
| **Logs de Auditoria** | 5 anos | [ ] | Politica implementada (CMP-004) | DBA + Compliance |
| **Dados de Chaves PIX (ativas)** | Enquanto ativa | [ ] | Politica implementada | DBA |
| **Dados de Chaves PIX (deletadas)** | 5 anos apos exclusao | [ ] | Soft delete + hard delete | Backend + DBA |
| **Dados de Claims** | 5 anos | [ ] | Politica de retencao | DBA |
| **Backups** | 5 anos | [ ] | Retencao de backups | Infra + DBA |

**Penalidade**: Violacao de obrigacao legal. Impossibilidade de comprovar operacoes em auditoria.

**Prazo de Implementacao**: P0 (critico - antes da producao)

---

## 9. Requisitos de Contingencia e Recuperacao

### 9.1 Plano de Contingencia

**Requisito**: Manual Operacional DICT, Secao "Contingencia"

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| Plano de Contingencia documentado | [ ] | Documento de Contingency Plan | CTO + DevOps | P0 |
| Backup automatico diario | [ ] | Configuracao de backup | DBA + Infra | P0 |
| Teste de restore de backup (mensal) | [ ] | Relatorios de teste | DBA | P1 |
| Replica de banco de dados (standby) | [ ] | Configuracao de replica | DBA + Infra | P0 |
| Failover automatico (RTO <= 15 minutos) | [ ] | Teste de failover | DevOps + Infra | P0 |
| Multi-AZ deployment (AWS) | [ ] | Configuracao de infraestrutura | Infra Team | P0 |

**Penalidade**: Indisponibilidade prolongada. Violacao de SLA.

**Prazo de Implementacao**: P0 (critico - antes da producao)

---

### 9.2 Disaster Recovery

| Item | Status | Evidencia | Responsavel | Prazo |
|------|--------|-----------|-------------|-------|
| Disaster Recovery Plan documentado | [ ] | Documento de DR Plan | CTO + DevOps | P1 |
| Backup em regiao AWS diferente (cross-region) | [ ] | Configuracao de backup | Infra Team | P1 |
| Simulacao de disaster recovery (anual) | [ ] | Relatorio de simulacao | CTO + DevOps | P2 |

**Prazo de Implementacao**: P1 (alta prioridade)

---

## 10. Checklist Resumido: Criterios de Go-Live

### 10.1 Criterios Obrigatorios (Go/No-Go)

| Criterio | Status | Bloqueante | Responsavel |
|----------|--------|------------|-------------|
| Certificado ICP-Brasil valido e instalado | [ ] | SIM | Security Team |
| Conectividade RSFN estabelecida | [ ] | SIM | Infra Team |
| ISPB cadastrado no DICT homologacao | [ ] | SIM | Head de Produto |
| Homologacao Bacen APROVADA (todos os 7 testes) | [ ] | SIM | CTO + All Teams |
| SLA de disponibilidade >= 99.9% | [ ] | SIM | DevOps Team |
| Logs de auditoria com retencao de 5 anos | [ ] | SIM | Backend + DBA |
| Criptografia em transito e repouso | [ ] | SIM | Security + Infra |
| Backup automatico configurado | [ ] | SIM | DBA + Infra |
| Monitoramento 24/7 ativo | [ ] | SIM | DevOps Team |
| Plano de Contingencia documentado e testado | [ ] | SIM | CTO + DevOps |

**REGRA**: Todos os criterios DEVEM estar marcados como [x] para Go-Live em producao.

---

## 11. Resumo de Prioridades

### P0 - Critico (Antes da Homologacao)

- Certificacao ICP-Brasil
- Conectividade RSFN
- ISPB cadastrado
- Preparacao para homologacao (1.000 chaves + 5 transacoes)
- Todos os 7 testes de homologacao prontos
- Validacao de chaves PIX
- Validacao de limites
- mTLS configurado
- Logs de auditoria com retencao de 5 anos
- Criptografia (TLS 1.2+, TDE, backups)

### P1 - Alto (Antes da Producao)

- SLA >= 99.9%
- Monitoramento 24/7
- Backup e restore testado
- Failover automatico
- Disaster Recovery Plan

### P2 - Medio (Apos Producao)

- Simulacao de disaster recovery
- Otimizacoes de performance

---

## 12. Contatos Bacen

| Assunto | Contato |
|---------|---------|
| **Email Operacional PIX** | pix-operacional@bcb.gov.br |
| **Protocolo Digital** | https://www3.bcb.gov.br/protocolo |
| **Departamento** | DECEM (Departamento de Competicao e de Estrutura do Mercado Financeiro) |

---

## 13. Referencias

### Internas

- [REG-001: Requisitos Regulatorios Bacen](../06_Regulatorio/REG-001_Requisitos_Regulatorios_Bacen.md)
- [CMP-001: Audit Logs Specification](CMP-001_Audit_Logs_Specification.md)
- [CMP-004: Data Retention Policy](CMP-004_Data_Retention_Policy.md)
- [SEC-001: mTLS Configuration](../13_Seguranca/SEC-001_mTLS_Configuration.md)
- [SEC-002: ICP-Brasil Certificates](../13_Seguranca/SEC-002_ICP_Brasil_Certificates.md)

### Externas

- Manual Operacional DICT v8 (Banco Central do Brasil)
- Instrucao Normativa BCB no 508/2024 (atualizada pela IN BCB 580/2024)
- Resolucao BCB no 1/2020 (Regulamento PIX)
- Circular 3.682/2013 (Auditoria de sistemas)

---

**Versao**: 1.0
**Status**: Checklist Regulatorio Bacen
**Proxima Revisao**: Trimestral ou apos mudanca regulatoria
