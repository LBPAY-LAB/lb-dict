# CMP-002: LGPD Compliance Checklist

**Projeto**: DICT - Diretorio de Identificadores de Contas Transacionais (LBPay)
**Versao**: 1.0
**Data**: 2025-10-25
**Status**: Checklist Operacional
**Responsavel**: DPO (Data Protection Officer) + Compliance Team

---

## 1. Resumo Executivo

Este documento apresenta **checklist completo de conformidade com a LGPD** (Lei Geral de Protecao de Dados - Lei 13.709/2018) para o sistema DICT da LBPay.

**Objetivo**: Garantir conformidade total com todos os artigos aplicaveis da LGPD e evitar penalidades (ate 2% do faturamento ou R$ 50 milhoes por infracao).

---

## 2. Fundamentos da LGPD

### 2.1 Escopo da LGPD no DICT

**Dados Pessoais Tratados**:
- CPF, CNPJ (chaves PIX)
- Nome completo
- Telefone celular
- Email
- Dados bancarios (ISPB, agencia, conta)
- Historico de claims e portabilidades

**Base Legal Principal**: **Execucao de Contrato** (Art. 7º, V LGPD)
- Contrato: Conta corrente do cliente com LBPay
- Necessidade: Cadastro de chave PIX e necessario para transacoes PIX

**Bases Legais Secundarias**:
- Logs de auditoria: **Obrigacao Legal** (Art. 7º, II) - Bacen exige retencao de 5 anos
- Logs de acesso (IP, timestamps): **Legitimo Interesse** (Art. 7º, IX) - Seguranca
- Marketing (se aplicavel): **Consentimento** (Art. 7º, I)

---

## 3. Checklist: 10 Principios da LGPD (Art. 6º)

### Principio 1: Finalidade (Art. 6º, I)

**Requisito**: Tratamento para propositos legitimos, especificos e informados ao titular.

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Finalidade documentada em Politica de Privacidade | [ ] | Link para Politica de Privacidade publicada | DPO + Legal |
| Finalidade: "Cadastro de chaves PIX para transacoes financeiras" | [ ] | Texto na Politica de Privacidade | DPO |
| Dados NAO sao usados para marketing sem consentimento adicional | [ ] | Verificacao de processos internos | Compliance |
| Aviso de Privacidade exibido ao cadastrar chave PIX | [ ] | Screenshot do app/web | Product Team |

**Prazo de Implementacao**: P0 (antes do lancamento)

---

### Principio 2: Adequacao (Art. 6º, II)

**Requisito**: Compatibilidade do tratamento com as finalidades informadas.

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Coleta apenas dados necessarios para operacao PIX | [ ] | Analise de campos no cadastro de chave | Tech Lead |
| Dados irrelevantes NAO sao coletados (genero, estado civil, etc.) | [ ] | Analise de schema de banco de dados | Architect |
| Alteracoes de finalidade sao comunicadas ao titular | [ ] | Processo de change management | DPO + Product |

**Prazo de Implementacao**: P0 (antes do lancamento)

---

### Principio 3: Necessidade (Art. 6º, III)

**Requisito**: Minimizacao de dados (coletar apenas o necessario).

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Campos obrigatorios limitados a: CPF/CNPJ, nome, conta | [ ] | Validacao de formularios | Tech Lead |
| Campos opcionais claramente identificados | [ ] | Screenshot do app | UX Team |
| Dados desnecessarios NAO sao coletados (ex: renda, profissao) | [ ] | Analise de schema de banco de dados | Architect |

**Prazo de Implementacao**: P0 (antes do lancamento)

---

### Principio 4: Livre Acesso (Art. 6º, IV)

**Requisito**: Garantir ao titular consulta facilitada e gratuita sobre seus dados.

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| API de DSAR (Data Subject Access Request) implementada | [ ] | Endpoint GET /api/v1/privacy/my-data funcionando | Backend Team |
| Usuario pode consultar seus dados pelo app/web | [ ] | Screenshot da funcionalidade | Product Team |
| Consulta e gratuita (sem custo ao usuario) | [ ] | Validacao de politica de precos | Legal + Finance |
| SLA de resposta: 15 dias uteis (Art. 18, §3º) | [ ] | Processo de DSAR documentado | DPO + Support |

**Prazo de Implementacao**: P1 (ate 3 meses apos lancamento)

---

### Principio 5: Qualidade dos Dados (Art. 6º, V)

**Requisito**: Garantir exatidao, clareza, relevancia e atualizacao dos dados.

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Validacao de CPF/CNPJ (algoritmo de digito verificador) | [ ] | Codigo de validacao implementado | Backend Team |
| Validacao de email (formato RFC 5322) | [ ] | Codigo de validacao implementado | Backend Team |
| Validacao de telefone (formato E.164) | [ ] | Codigo de validacao implementado | Backend Team |
| Usuario pode corrigir dados incorretos | [ ] | Funcionalidade de edicao no app | Product Team |
| Processo de correcao documentado | [ ] | Documento de processo | DPO + Support |

**Prazo de Implementacao**: P0 (antes do lancamento)

---

### Principio 6: Transparencia (Art. 6º, VI)

**Requisito**: Informacoes claras e acessiveis aos titulares.

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Politica de Privacidade publicada e acessivel | [ ] | Link visivel no app/web | Legal + Product |
| Politica em linguagem clara e simples | [ ] | Revisao por especialista em Plain Language | Legal + UX |
| Aviso no momento da coleta ("Ao cadastrar chave PIX, voce concorda...") | [ ] | Screenshot do aviso | Product Team |
| Portal de Privacidade disponivel (titular consulta seus dados) | [ ] | URL do portal funcional | Product Team |
| Informacao sobre compartilhamento de dados (se aplicavel) | [ ] | Secao na Politica de Privacidade | Legal + DPO |

**Prazo de Implementacao**: P0 (antes do lancamento)

---

### Principio 7: Seguranca (Art. 6º, VII)

**Requisito**: Medidas tecnicas e administrativas para proteger dados.

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Criptografia em transito (TLS 1.2+) | [ ] | Configuracao de servidores HTTPS | Infra Team |
| Criptografia em repouso (PostgreSQL TDE) | [ ] | Configuracao de banco de dados | DBA + Security |
| Controle de acesso (RBAC) implementado | [ ] | Codigo de RBAC + testes | Backend Team |
| Logs de auditoria (quem acessou o que, quando) | [ ] | Tabelas audit.* no PostgreSQL | Backend + DBA |
| Masking de dados sensiveis em logs | [ ] | Codigo de masking + validacao | Backend Team |
| Backups criptografados | [ ] | Configuracao de backup (AES-256) | Infra + DBA |

**Prazo de Implementacao**: P0 (antes do lancamento)

---

### Principio 8: Prevencao (Art. 6º, VIII)

**Requisito**: Medidas para prevenir danos aos titulares.

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Privacy by Design implementado (seguranca desde o design) | [ ] | Analise de arquitetura | Architect + Security |
| DPIA (Data Protection Impact Assessment) realizado | [ ] | Documento CMP-005 (DPIA) completo | DPO + Security |
| Plano de Resposta a Incidentes documentado | [ ] | Documento de IRP (Incident Response Plan) | Security Team |
| Simulacao de incidente de seguranca (anual) | [ ] | Relatorio de simulacao | Security Team |

**Prazo de Implementacao**: P1 (ate 6 meses apos lancamento)

---

### Principio 9: Nao Discriminacao (Art. 6º, IX)

**Requisito**: Nao usar dados para fins discriminatorios.

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Dados NAO sao usados para analise de credito sem consentimento | [ ] | Validacao de processos de negocio | Compliance + Legal |
| Dados NAO sao usados para perfilagem discriminatoria | [ ] | Validacao de algoritmos (se aplicavel) | Data Science + Legal |
| Politica contra discriminacao documentada | [ ] | Documento de politica interna | Legal + HR |

**Prazo de Implementacao**: P0 (antes do lancamento)

---

### Principio 10: Responsabilizacao e Prestacao de Contas (Art. 6º, X)

**Requisito**: Demonstrar conformidade com LGPD.

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| DPO (Data Protection Officer) nomeado | [ ] | Portaria de nomeacao + registro ANPD | CEO + DPO |
| ROPA (Registro de Operacoes de Tratamento de Dados) mantido | [ ] | Documento ROPA atualizado | DPO |
| Auditorias periodicas (anuais) agendadas | [ ] | Calendario de auditorias | DPO + Compliance |
| Documentacao de processos de privacidade | [ ] | Biblioteca de documentos (CMP-001 a CMP-005) | DPO + Architect |
| Treinamento de funcionarios sobre LGPD (anual) | [ ] | Certificados de treinamento | HR + DPO |

**Prazo de Implementacao**: P0 (antes do lancamento)

---

## 4. Checklist: 9 Direitos dos Titulares (Art. 18)

### Direito 1: Confirmacao de Tratamento (Art. 18, I)

**Requisito**: Confirmar se a LBPay trata dados do titular.

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Processo de confirmacao documentado | [ ] | Fluxo de atendimento | DPO + Support |
| Resposta em ate 15 dias uteis | [ ] | SLA monitorado | Support Team |
| Canal de atendimento disponivel (email, formulario web) | [ ] | Email dpo@lbpay.com.br ativo | IT + DPO |

**Prazo de Implementacao**: P1 (ate 3 meses apos lancamento)

---

### Direito 2: Acesso aos Dados (Art. 18, II)

**Requisito**: Fornecer copia dos dados pessoais do titular.

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| API de DSAR implementada (GET /api/v1/privacy/my-data) | [ ] | Endpoint funcionando | Backend Team |
| Dados retornados em formato legivel (JSON ou PDF) | [ ] | Teste de endpoint | Backend Team |
| SLA de resposta: 15 dias uteis | [ ] | Processo de DSAR | DPO + Support |

**Prazo de Implementacao**: P1 (ate 3 meses apos lancamento)

---

### Direito 3: Correcao de Dados (Art. 18, III)

**Requisito**: Permitir correcao de dados incorretos.

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Funcionalidade de edicao de dados no app/web | [ ] | Screenshot da funcionalidade | Product Team |
| Validacao de identidade antes de correcao | [ ] | Codigo de autenticacao | Backend Team |
| Correcao refletida no DICT Bacen (UpdateEntry) | [ ] | Integracao com Bridge | Backend Team |
| Log de auditoria de correcoes | [ ] | Tabela audit.entry_events | Backend Team |

**Prazo de Implementacao**: P1 (ate 3 meses apos lancamento)

---

### Direito 4: Anonimizacao, Bloqueio ou Eliminacao (Art. 18, IV)

**Requisito**: Permitir anonimizacao, bloqueio ou eliminacao de dados desnecessarios ou excessivos.

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Processo de exclusao de chave PIX implementado | [ ] | Funcionalidade de exclusao no app | Product Team |
| Soft delete (status = DELETED) implementado | [ ] | Codigo de soft delete | Backend Team |
| Hard delete apos periodo de retencao (5 anos) | [ ] | Script de purge automatizado | DBA + Backend |
| EXCECAO: Logs de auditoria (Bacen exige 5 anos) NAO deletados | [ ] | Validacao de retencao | DPO + Compliance |

**Prazo de Implementacao**: P0 (antes do lancamento)

---

### Direito 5: Portabilidade (Art. 18, V)

**Requisito**: Fornecer dados em formato estruturado e interoperavel.

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Export de dados em JSON ou CSV | [ ] | Endpoint GET /api/v1/privacy/export | Backend Team |
| Formato conforme especificacao (RFC, OpenAPI) | [ ] | Validacao de formato | Backend Team |
| SLA de geracao do arquivo: 5 dias uteis | [ ] | Processo de export | DPO + Backend |

**Prazo de Implementacao**: P1 (ate 6 meses apos lancamento)

---

### Direito 6: Eliminacao de Dados (Art. 18, VI)

**Requisito**: Excluir dados tratados com consentimento.

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Funcionalidade de exclusao disponivel no app | [ ] | Screenshot | Product Team |
| Exclusao imediata (soft delete) | [ ] | Codigo de exclusao | Backend Team |
| Hard delete apos periodo de retencao | [ ] | Script de purge | DBA + Backend |
| NOTA: Nao aplicavel a dados de contrato (base legal: execucao de contrato) | [ ] | Validacao legal | Legal + DPO |

**Prazo de Implementacao**: P1 (ate 3 meses apos lancamento)

---

### Direito 7: Informacao sobre Compartilhamento (Art. 18, VII)

**Requisito**: Informar com quem os dados sao compartilhados.

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Lista de compartilhamentos em Politica de Privacidade | [ ] | Secao "Compartilhamento" na Politica | Legal + DPO |
| Compartilhamentos: Bacen (DICT), AWS (infraestrutura) | [ ] | Texto na Politica | Legal + DPO |
| Contratos de processamento com terceiros (AWS, etc.) | [ ] | Contratos assinados | Legal + Procurement |

**Prazo de Implementacao**: P0 (antes do lancamento)

---

### Direito 8: Informacao sobre Nao Consentimento (Art. 18, VIII)

**Requisito**: Informar sobre consequencias de nao fornecer consentimento.

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| NOTA: Nao aplicavel (base legal: execucao de contrato, nao consentimento) | [x] | N/A | N/A |
| Se marketing: informar que marketing e opcional | [ ] | Texto em formulario de opt-in | Product + Legal |

**Prazo de Implementacao**: N/A

---

### Direito 9: Revogacao de Consentimento (Art. 18, IX)

**Requisito**: Permitir revogacao de consentimento a qualquer momento.

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| NOTA: Nao aplicavel (base legal: execucao de contrato, nao consentimento) | [x] | N/A | N/A |
| Se marketing: botao de opt-out disponivel | [ ] | Funcionalidade de opt-out | Product Team |
| Revogacao nao afeta dados de contrato (chave PIX continua ativa) | [ ] | Validacao de regra de negocio | Product + Legal |

**Prazo de Implementacao**: N/A

---

## 5. Checklist: Obrigacoes Organizacionais

### 5.1 DPO (Data Protection Officer) - Art. 41

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| DPO nomeado (nome completo, CPF, contato) | [ ] | Portaria de nomeacao | CEO |
| DPO registrado na ANPD | [ ] | Comprovante de registro ANPD | DPO |
| Contato do DPO publicado (email, telefone) | [ ] | Texto na Politica de Privacidade | DPO + Legal |
| DPO tem autonomia e recursos adequados | [ ] | Validacao organizacional | CEO + HR |

**Prazo de Implementacao**: P0 (antes do lancamento)

---

### 5.2 ROPA (Registro de Operacoes de Tratamento de Dados)

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| ROPA documentado (planilha ou sistema) | [ ] | Arquivo ROPA atualizado | DPO |
| ROPA inclui: Finalidade, Base Legal, Categorias de Dados, Retencao | [ ] | Validacao de conteudo | DPO + Legal |
| ROPA atualizado semestralmente | [ ] | Processo de atualizacao | DPO |

**Prazo de Implementacao**: P0 (antes do lancamento)

---

### 5.3 DPIA (Data Protection Impact Assessment) - Art. 38

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| DPIA realizado (documento CMP-005) | [ ] | Documento CMP-005 completo | DPO + Security |
| DPIA identifica PII processado | [ ] | Secao 3 do CMP-005 | DPO |
| DPIA avalia riscos | [ ] | Secao 4 do CMP-005 | DPO + Security |
| DPIA propoe medidas de mitigacao | [ ] | Secao 5 do CMP-005 | DPO + Security |
| DPIA revisado anualmente | [ ] | Processo de revisao | DPO |

**Prazo de Implementacao**: P1 (ate 6 meses apos lancamento)

---

### 5.4 Politica de Privacidade

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Politica de Privacidade redigida | [ ] | Documento de Politica | Legal + DPO |
| Politica publicada e acessivel (link visivel) | [ ] | URL da Politica | Product Team |
| Politica inclui: Finalidade, Base Legal, Retencao, Direitos, Contato DPO | [ ] | Validacao de conteudo | Legal + DPO |
| Politica em linguagem clara | [ ] | Revisao por especialista | Legal + UX |
| Versao data e historico de alteracoes mantido | [ ] | Controle de versao | Legal + DPO |

**Prazo de Implementacao**: P0 (antes do lancamento)

---

### 5.5 Contratos de Processamento (Art. 39)

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Contrato com AWS (infraestrutura) inclui clausulas LGPD | [ ] | Contrato assinado | Legal + Procurement |
| Contrato com Bacen (DICT) - nao aplicavel (orgao regulador) | [x] | N/A | N/A |
| Contratos incluem: Finalidade, Prazo, Medidas de Seguranca, Responsabilidades | [ ] | Validacao de clausulas | Legal |

**Prazo de Implementacao**: P0 (antes do lancamento)

---

### 5.6 Treinamento de Funcionarios

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Treinamento LGPD obrigatorio para todos os funcionarios | [ ] | Plataforma de treinamento | HR + DPO |
| Frequencia: anual | [ ] | Calendario de treinamentos | HR |
| Certificados de conclusao mantidos | [ ] | Registros de treinamento | HR |
| Treinamento especifico para equipe de TI e Seguranca | [ ] | Conteudo avancado | DPO + Security |

**Prazo de Implementacao**: P1 (ate 3 meses apos lancamento)

---

## 6. Checklist: Resposta a Incidentes (Art. 48)

### 6.1 Plano de Resposta a Incidentes

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Plano de Resposta a Incidentes (IRP) documentado | [ ] | Documento de IRP | Security Team |
| Equipe de resposta a incidentes definida | [ ] | Lista de membros + contatos | Security Team |
| Runbook de incidentes criado (passo a passo) | [ ] | Runbook documentado | Security Team |
| Simulacao de incidente realizada (anualmente) | [ ] | Relatorio de simulacao | Security Team |

**Prazo de Implementacao**: P1 (ate 6 meses apos lancamento)

---

### 6.2 Notificacao de Incidentes

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Processo de notificacao a ANPD documentado | [ ] | Fluxo de notificacao | DPO + Legal |
| SLA de notificacao: 72 horas apos ciencia do incidente | [ ] | Processo de SLA | DPO |
| Template de notificacao a ANPD preparado | [ ] | Documento de template | DPO + Legal |
| Template de notificacao a titulares afetados preparado | [ ] | Documento de template | DPO + Legal |
| Canal de comunicacao com ANPD estabelecido | [ ] | Email/portal ANPD configurado | DPO |

**Prazo de Implementacao**: P1 (ate 6 meses apos lancamento)

---

## 7. Checklist Tecnico

### 7.1 Seguranca Tecnica

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| TLS 1.2+ em todas as conexoes | [ ] | Configuracao de servidores | Infra Team |
| Criptografia em repouso (PostgreSQL TDE) | [ ] | Configuracao de banco | DBA + Security |
| Criptografia de backups (AES-256) | [ ] | Configuracao de backup | Infra + DBA |
| RBAC (Role-Based Access Control) implementado | [ ] | Codigo de RBAC | Backend Team |
| MFA (Multi-Factor Authentication) disponivel | [ ] | Funcionalidade de MFA | Backend + Product |
| Rate limiting implementado | [ ] | Configuracao de API Gateway | Infra Team |

**Prazo de Implementacao**: P0 (antes do lancamento)

---

### 7.2 Auditoria e Logging

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Logs de auditoria implementados (CMP-001) | [ ] | Tabelas audit.* no PostgreSQL | Backend + DBA |
| Masking de dados sensiveis em logs | [ ] | Codigo de masking | Backend Team |
| Retencao de logs: 5 anos (Bacen) | [ ] | Politica de retencao implementada | DBA + Infra |
| Archival para S3 Glacier configurado | [ ] | Pipeline de archival | Infra Team |
| Dashboards de auditoria disponiveis (Grafana) | [ ] | Dashboards criados | DevOps Team |

**Prazo de Implementacao**: P0 (antes do lancamento)

---

### 7.3 Gestao de Dados

| Item | Status | Evidencia | Responsavel |
|------|--------|-----------|-------------|
| Soft delete implementado (status = DELETED) | [ ] | Codigo de soft delete | Backend Team |
| Hard delete apos retencao (5 anos) | [ ] | Script de purge | DBA + Backend |
| Politica de retencao documentada (CMP-004) | [ ] | Documento CMP-004 completo | DPO + DBA |
| Processo de exclusao de backups antigos | [ ] | Script de purge de backups | Infra + DBA |

**Prazo de Implementacao**: P0 (antes do lancamento)

---

## 8. Resumo de Prioridades

### P0 - Critico (Antes do Lancamento)

- DPO nomeado e registrado na ANPD
- Politica de Privacidade publicada
- Criptografia (TLS 1.2+, TDE, backups)
- RBAC implementado
- Logs de auditoria com masking
- Soft delete e hard delete implementados
- Validacao de dados (CPF, CNPJ, email, telefone)

### P1 - Alto (Ate 3 meses apos lancamento)

- API de DSAR (acesso aos dados)
- Funcionalidade de correcao de dados
- Funcionalidade de exclusao de dados
- ROPA mantido e atualizado
- Treinamento LGPD para funcionarios

### P2 - Medio (Ate 6 meses apos lancamento)

- DPIA completo (CMP-005)
- Plano de Resposta a Incidentes
- Portabilidade de dados (export JSON/CSV)
- Simulacao de incidente de seguranca

---

## 9. Responsaveis e Prazos

| Responsavel | Total de Itens | Prazo Maximo |
|-------------|----------------|--------------|
| **DPO** | 25 | P0 a P2 |
| **Legal** | 15 | P0 |
| **Backend Team** | 18 | P0 a P1 |
| **Product Team** | 12 | P0 a P1 |
| **Security Team** | 10 | P0 a P2 |
| **Infra/DBA** | 8 | P0 |
| **HR** | 3 | P1 |

---

## 10. Referencias

### Internas

- [SEC-007: LGPD Data Protection](../13_Seguranca/SEC-007_LGPD_Data_Protection.md)
- [CMP-001: Audit Logs Specification](CMP-001_Audit_Logs_Specification.md)
- [CMP-004: Data Retention Policy](CMP-004_Data_Retention_Policy.md)
- [CMP-005: Privacy Impact Assessment](CMP-005_Privacy_Impact_Assessment.md)

### Externas

- Lei 13.709/2018 (LGPD): http://www.planalto.gov.br/ccivil_03/_ato2015-2018/2018/lei/l13709.htm
- Guia de Boas Praticas LGPD - ANPD: https://www.gov.br/anpd/
- ISO 27701:2019 (Privacy Information Management)

---

**Versao**: 1.0
**Status**: Checklist Operacional
**Proxima Revisao**: Trimestral
