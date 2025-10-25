# Plano Master do Projeto - DICT LBPay

**ID**: PMP-001
**Versão**: 1.0
**Data**: 2025-10-24
**PM**: PHOENIX (AGT-PM-001)
**Status**: Draft - Aguardando Kickoff

---

## 1. Visão Geral do Projeto

### 1.1 Contexto
O LBPay é uma instituição de pagamento licenciada pelo Banco Central do Brasil e participante direto do PIX. Implementamos nossa própria solução de Core Banking (Contas de Pagamento) e integrações com o Bacen via RSFN (SPI - PIX). Este projeto visa implementar a solução DICT (Diretório de Identificadores de Contas Transacionais) para possibilitar a operação completa do PIX.

### 1.2 Objetivo
Implementar o módulo DICT visando:
1. **Homologação no DICT Bacen** (objetivo primário)
2. **Entrada em produção** (após homologação)
3. **Operação completa do PIX** com gerenciamento de chaves

### 1.3 Justificativa
Para operarmos o PIX completamente, precisamos:
- Homologar no DICT (Bacen)
- Gerenciar chaves PIX dos nossos clientes
- Integrar com o diretório centralizado do Bacen
- Atender todos os requisitos regulatórios

---

## 2. Escopo do Projeto

### 2.1 Fases do Projeto

#### Fase 1: Especificação e Planejamento (Atual)
**Objetivo**: Criar todos os artefatos necessários para implementação autônoma por agentes Claude Code

**Entregas**:
- [ ] Checklist completo de requisitos funcionais (CRF-001)
- [ ] Arquitetura de solução detalhada (DAS-001)
- [ ] Modelo de dados completo (MDC/MDL/MDF-001)
- [ ] Especificações de todas as APIs (EAI/CGR-XXX)
- [ ] Especificações de frontend (LFF-001)
- [ ] Especificações de integração E2E (MFE-XXX)
- [ ] Análise de segurança (ASG-001)
- [ ] Estratégia de testes (EST-001)
- [ ] Plano de homologação (PTH-001)
- [ ] Backlog de desenvolvimento (BKL-001)
- [ ] Squad de desenvolvimento definida

**Duração Estimada**: 8 semanas
**Squad**: Squad de Arquitetura (14 agentes)

#### Fase 2: Implementação
**Objetivo**: Implementar a solução completa baseada nos artefatos da Fase 1

**Entregas**:
- [ ] Core DICT implementado
- [ ] Connect DICT (abstraído) implementado
- [ ] Bridge DICT (abstraído) implementado
- [ ] Frontend implementado
- [ ] Testes automatizados (unit, integration, e2e)
- [ ] CI/CD pipelines
- [ ] Documentação técnica

**Duração Estimada**: A definir após Fase 1
**Squad**: Squad de Desenvolvimento (a definir)

### 2.2 Escopo Funcional

#### Bloco 1: CRUD de Chaves PIX
- Criar chave (todos os tipos: CPF, CNPJ, Email, Telefone, Aleatória)
- Consultar chave
- Alterar dados da chave
- Excluir chave (vários motivos)
- Validar chave

#### Bloco 2: Reivindicação e Portabilidade
- Criar reivindicação (reivindicador)
- Cancelar reivindicação
- Confirmar reivindicação
- Consultar/Listar reivindicações
- Portabilidade de chave

#### Bloco 3: Validações
- Validação de posse
- Validação cadastral (Receita Federal)
- Validação de nomes vinculados

#### Bloco 4: Devolução e Infração
- Solicitar devolução (falha operacional)
- Solicitar devolução (suspeita de fraude)
- Notificação de infração
- Cancelamento de devolução

#### Bloco 5: Segurança e Infraestrutura
- Prevenção a ataques de leitura
- Rate limiting
- Cache de chaves
- Comunicação segura (mTLS)
- Monitoramento

#### Bloco 6: Recuperação de Valores
- Fluxo interativo de recuperação
- Fluxo automatizado de recuperação
- Rastreamento de valores
- Bloqueio/Desbloqueio de recursos

### 2.3 Escopo Técnico

**Repositórios Envolvidos**:
1. **Core DICT** (novo ou evolution de existing)
   - Repos: money-moving, orchestration-go, operation, lb-contracts
2. **Connect DICT** (evolution)
   - Repo: connector-dict
3. **Bridge DICT** (evolution)
   - Repo: rsfn-connect-bacen-bridge
4. **Simulador DICT** (evolution)
   - Repo: simulator-dict

**Stack Tecnológica**:
- **Backend**: Golang
- **Comunicação Interna**: gRPC
- **Comunicação Bacen**: REST (via Bridge)
- **Banco de Dados**: PostgreSQL
- **Cache**: Redis
- **Message Broker**: RabbitMQ/Kafka
- **Observabilidade**: Prometheus, Grafana, Jaeger
- **CI/CD**: GitHub Actions
- **Frontend**: [A definir - provavelmente React/Next.js]

### 2.4 Fora do Escopo (Fase 1)
- Implementação de código
- Testes em ambiente Bacen
- Deploy em produção
- Treinamento de usuários finais

---

## 3. Stakeholders

### 3.1 Stakeholders Executivos
| Nome/Papel | Responsabilidade | Nível de Envolvimento |
|------------|------------------|----------------------|
| CTO | Aprovação final de especificações e arquitetura | Alto - Aprovações |
| Head de Arquitetura | Aprovação de decisões arquiteturais | Alto - Reviews semanais |
| Head de Produto | Aprovação de requisitos funcionais | Alto - Reviews semanais |
| Head de Engenharia | Aprovação de stack e implementação | Médio - Reviews quinzenais |

### 3.2 Squad de Arquitetura (Fase 1)
Ver documento: [SQUAD_ARCHITECTURE.md](../SQUAD_ARCHITECTURE.md)

- AGT-PM-001 (PHOENIX) - Project Manager
- AGT-SM-001 (CATALYST) - Scrum Master
- AGT-BA-001 (ORACLE) - Business Analyst
- AGT-SA-001 (NEXUS) - Solution Architect
- AGT-DA-001 (ATLAS) - Data Architect
- AGT-API-001 (MERCURY) - API Specialist
- AGT-FE-001 (PRISM) - Frontend Architect
- AGT-INT-001 (CONDUIT) - Integration Architect
- AGT-SEC-001 (SENTINEL) - Security Architect
- AGT-QA-001 (VALIDATOR) - QA Architect
- AGT-DV-001 (FORGE) - DevOps Architect
- AGT-TS-001 (GOPHER) - Tech Specialist Golang
- AGT-DOC-001 (SCRIBE) - Technical Writer
- AGT-CM-001 (GUARDIAN) - Compliance Manager

### 3.3 Comunicação com Stakeholders

**Reuniões Regulares**:
- **Daily Standup**: Diário, Squad interna (15min)
- **Sprint Planning**: Semanal (1h)
- **Sprint Review**: Semanal com stakeholders (1h)
- **Retrospectiva**: Semanal, Squad interna (45min)
- **Status Report**: Semanal para executivos (assíncrono)

---

## 4. Cronograma (Fase 1 - 8 semanas)

### Sprint 1-2: Análise e Descoberta
**Objetivos**:
- Análise completa de documentação Bacen
- Análise de código existente (repos)
- Identificação de requisitos funcionais
- Identificação de gaps

**Entregas**:
- CRF-001: Checklist de Requisitos Funcionais
- MTR-001: Matriz de Rastreabilidade
- AST-001: Análise de Stack Tecnológica
- ARE-XXX: Análise de Repositórios existentes
- CAB-001: Catálogo de APIs Bacen

**Agentes Ativos**: ORACLE, MERCURY, GUARDIAN, GOPHER

### Sprint 3-4: Design e Arquitetura
**Objetivos**:
- Design de arquitetura de solução
- Modelagem de dados
- Especificação de integrações
- Decisões arquiteturais (ADRs)

**Entregas**:
- DAS-001: Arquitetura de Solução
- MDC/MDL/MDF-001: Modelos de Dados
- ADR-XXX: Architecture Decision Records
- MIG-001: Mapa de Integrações
- ECD/EBD-001: Especificações Connect/Bridge

**Agentes Ativos**: NEXUS, ATLAS, CONDUIT, SENTINEL

### Sprint 5-6: Especificação Detalhada
**Objetivos**:
- Especificação de frontend
- Especificação de APIs internas
- User stories detalhadas
- Estratégia de testes

**Entregas**:
- LFF-001: Lista de Funcionalidades Frontend
- UST-XXX: User Stories completas
- EAI/CGR-XXX: Especificações de APIs
- EST-001: Estratégia de Testes
- PTH-001: Plano de Homologação

**Agentes Ativos**: PRISM, ORACLE, MERCURY, VALIDATOR

### Sprint 7-8: Consolidação e Planejamento
**Objetivos**:
- Consolidação de toda documentação
- Criação de backlog de desenvolvimento
- Preparação para aprovações
- Definição de Squad de Desenvolvimento

**Entregas**:
- BKL-001: Backlog Master de Desenvolvimento
- IMD-001: Índice Master de Documentação
- PAP-XXX: Pacotes de Aprovação
- Squad de Desenvolvimento definida
- Plano de Fase 2

**Agentes Ativos**: PHOENIX, CATALYST, SCRIBE

---

## 5. Riscos e Mitigações

### 5.1 Matriz de Riscos

| ID | Risco | Probabilidade | Impacto | Mitigação |
|----|-------|---------------|---------|-----------|
| R-001 | Documentação Bacen incompleta ou ambígua | Média | Alto | Criar documento de dúvidas; consultar Bacen se necessário |
| R-002 | Requisitos de homologação mudarem | Baixa | Alto | Monitorar comunicados Bacen; manter arquitetura flexível |
| R-003 | Código existente não documentado suficientemente | Alta | Médio | Análise profunda de código; engenharia reversa se necessário |
| R-004 | Atraso nas aprovações de stakeholders | Média | Médio | Preparar pacotes de aprovação claros; follow-ups semanais |
| R-005 | Complexidade subestimada | Média | Alto | Revisões frequentes; ajustar estimativas conforme necessário |
| R-006 | Falta de expertise em domínio DICT | Baixa | Médio | Documentação Bacen é suficiente; consultar especialistas |
| R-007 | Mudanças em repos existentes durante Fase 1 | Média | Médio | Snapshots de código; comunicação com outras equipes |

### 5.2 Plano de Contingência

**Se atrasos ocorrerem**:
1. Priorizar artefatos críticos (requisitos, arquitetura)
2. Paralelizar trabalhos quando possível
3. Reduzir escopo de artefatos não-críticos
4. Comunicar transparentemente com stakeholders

**Se requisitos mudarem**:
1. Impact assessment rápido
2. Atualizar artefatos afetados
3. Re-priorizar backlog
4. Comunicar mudanças

---

## 6. Critérios de Sucesso (Fase 1)

### 6.1 Critérios Obrigatórios
- [ ] 100% dos requisitos funcionais do Bacen catalogados
- [ ] Arquitetura de solução aprovada por Head de Arquitetura
- [ ] Modelo de dados completo e aprovado
- [ ] Todas as APIs especificadas (internas e Bacen)
- [ ] Backlog de desenvolvimento criado e priorizado
- [ ] Plano de homologação completo
- [ ] Todos os artefatos revisados e aprovados

### 6.2 Critérios de Qualidade
- [ ] Rastreabilidade completa (requisito → implementação)
- [ ] Todos os documentos indexados e cross-referenced
- [ ] Especificações claras o suficiente para implementação autônoma
- [ ] Critérios de aceitação mensuráveis para cada funcionalidade
- [ ] 0 ambiguidades críticas não resolvidas

### 6.3 Métricas de Sucesso
- **Completude**: > 95% de artefatos planejados criados
- **Qualidade**: > 90% de aprovação nas revisões
- **Clareza**: < 10 dúvidas críticas não resolvidas
- **Tempo**: Dentro de 8 semanas (±1 semana aceitável)

---

## 7. Dependências

### 7.1 Dependências Externas
- Acesso aos repositórios GitHub do LBPay
- Documentação Bacen (já disponível)
- Aprovações de stakeholders (CTO, Heads)

### 7.2 Dependências Internas
- Squad de Arquitetura disponível e operacional
- Ferramentas Claude Code funcionando
- MCP GitHub configurado

### 7.3 Dependências Técnicas
- Acesso aos repositórios para análise
- Documentação existente dos repos
- Ambiente de desenvolvimento (se necessário)

---

## 8. Comunicação e Reporting

### 8.1 Relatórios Regulares

**Status Report Semanal** (RST-YYYYMMDD):
- Progresso geral (%)
- Artefatos completados vs planejados
- Riscos e impedimentos
- Próximos passos
- Decisões necessárias

**Daily Summary** (DLY-YYYYMMDD):
- Trabalho de ontem
- Trabalho de hoje
- Bloqueadores

**Sprint Review Report** (SRV-XXX):
- Artefatos entregues
- Demo de artefatos-chave
- Feedback de stakeholders
- Ajustes necessários

### 8.2 Canais de Comunicação
- **Documentos**: `/Artefatos/11_Gestao/`
- **Status Reports**: Semanal para stakeholders
- **Dúvidas**: `/Artefatos/00_Master/DUVIDAS.md`
- **Decisões**: ADRs em `/Artefatos/02_Arquitetura/ADRs/`

---

## 9. Governança

### 9.1 Processo de Aprovação de Artefatos

1. **Criação**: Agente responsável cria artefato
2. **Peer Review**: Revisão por outros agentes (RACI)
3. **Consolidação**: SCRIBE revisa formato e clareza
4. **Validação PM**: PHOENIX valida completude
5. **Aprovação Stakeholder**: Submissão para aprovador apropriado
6. **Finalização**: Artefato marcado como "Approved"

### 9.2 Controle de Mudanças

**Mudanças em Artefatos Aprovados**:
1. Criar proposta de mudança
2. Impact assessment
3. Aprovação de PM
4. Atualização de artefato
5. Re-aprovação de stakeholder (se necessário)
6. Atualização de versão

### 9.3 Gestão de Backlog

**CATALYST (Scrum Master)** responsável por:
- Manter backlog priorizado
- Garantir que user stories estão "ready"
- Facilitar sprint planning
- Remover impedimentos

**PHOENIX (PM)** responsável por:
- Priorização estratégica
- Alinhamento com stakeholders
- Gestão de escopo
- Decisões de trade-off

---

## 10. Qualidade e Padrões

### 10.1 Padrões de Documentação
- Todos os documentos em Markdown
- Diagramas em Mermaid ou PlantUML
- Nomenclatura padronizada (códigos de artefatos)
- Templates obrigatórios para documentos principais
- Cross-referencing entre documentos

### 10.2 Definition of Done (DoD) - Artefatos

Um artefato está "Done" quando:
- [ ] Criado seguindo template apropriado
- [ ] Revisado por peers (conforme RACI)
- [ ] Sem inconsistências ou ambiguidades críticas
- [ ] Cross-references atualizados
- [ ] Indexado no IMD-001
- [ ] Aprovado por stakeholder apropriado (se aplicável)
- [ ] Versionado corretamente

### 10.3 Definition of Ready (DoR) - User Stories

Uma user story está "Ready" quando:
- [ ] Escrita no formato padrão (Como/Quero/Para)
- [ ] Critérios de aceitação claros e testáveis
- [ ] Regras de negócio identificadas
- [ ] Dependências mapeadas
- [ ] Estimada (story points)
- [ ] Priorizada
- [ ] Revisada e aprovada

---

## 11. Lições Aprendidas e Melhorias Contínuas

### 11.1 Retrospectivas
Ao final de cada sprint:
- O que funcionou bem?
- O que pode melhorar?
- Ações de melhoria
- Follow-up de ações anteriores

### 11.2 Ajustes de Processo
- Processos podem ser ajustados baseado em retrospectivas
- Mudanças de processo requerem consenso da Squad
- Documentar mudanças de processo

---

## 12. Transição para Fase 2

### 12.1 Critérios para Iniciar Fase 2
- [ ] Todos os artefatos da Fase 1 aprovados
- [ ] Squad de Desenvolvimento definida
- [ ] Repositórios preparados (branches, acessos)
- [ ] Ambiente de desenvolvimento pronto
- [ ] Backlog de implementação priorizado
- [ ] Go/No-Go decision de stakeholders

### 12.2 Handover
- Sessão de handover com Squad de Desenvolvimento
- Walkthrough de toda a documentação
- Q&A session
- Definição de canais de suporte

---

## 13. Próximos Passos Imediatos

1. **Kickoff Meeting** (data a definir)
   - Apresentação deste plano
   - Alinhamento com stakeholders
   - Definição de prazos finais
   - Aprovação para iniciar

2. **Configuração de Ambiente**
   - Estrutura de diretórios criada ✅
   - Templates criados ✅
   - Comandos Claude Code criados ✅
   - Acesso a repositórios (pendente)

3. **Sprint 1 Planning**
   - Definir tarefas detalhadas para Sprint 1
   - Atribuir agentes a tarefas
   - Definir entregáveis da sprint
   - Estabelecer cadência de dailies

---

## 14. Aprovações

- [ ] **PHOENIX (PM)**: [data]
- [ ] **CTO**: [data]
- [ ] **Head de Arquitetura**: [data]
- [ ] **Head de Produto**: [data]
- [ ] **Head de Engenharia**: [data]

---

## 15. Anexos

### A. Estrutura de Diretórios
Ver: `/Artefatos/` estrutura completa

### B. Squad de Arquitetura
Ver: [SQUAD_ARCHITECTURE.md](../SQUAD_ARCHITECTURE.md)

### C. Templates
Ver: `/Artefatos/99_Templates/`

### D. Comandos Claude Code
Ver: `/.claude/commands/`

---

**Histórico de Versões**:
| Data | Versão | Autor | Mudanças |
|------|--------|-------|----------|
| 2025-10-24 | 1.0 | PHOENIX | Criação inicial do plano master |

---

**Próxima Revisão**: Após Sprint 2 (ou conforme necessário)
