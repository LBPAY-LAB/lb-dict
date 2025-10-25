# Índice Geral - Projeto DICT LBPay

**Data**: 2025-10-24
**Status**: Setup Completo - Aguardando Kickoff
**Fase**: Fase 1 - Especificação e Planejamento

---

## 📋 Documentos de Início (LEITURA OBRIGATÓRIA)

### Para Executivos e Stakeholders
1. 📊 **[RESUMO_EXECUTIVO.md](./RESUMO_EXECUTIVO.md)** ⭐ COMECE AQUI
   - Visão geral em 5 minutos
   - O que precisa ser aprovado
   - Compromissos necessários

2. 📘 **[KICKOFF.md](./KICKOFF.md)**
   - Documento completo de kickoff (20 páginas)
   - Detalhes de escopo, cronograma, riscos
   - Cerimônias e processos

3. 📊 **[Plano Master](../11_Gestao/PMP-001_Plano_Master_Projeto.md)**
   - Plano completo do projeto
   - Fases, entregas, critérios de sucesso
   - Governança e aprovações

### Para Entender a Abordagem
4. 👥 **[SQUAD_ARCHITECTURE.md](../SQUAD_ARCHITECTURE.md)**
   - Definição completa dos 14 agentes
   - Responsabilidades e artefatos
   - Matriz RACI

5. ❓ **[DUVIDAS.md](./DUVIDAS.md)**
   - 10 dúvidas técnicas críticas
   - Requer resposta dos stakeholders
   - Atualizado continuamente

### Documento Principal do Projeto
6. 📖 **[README.md](../../README.md)**
   - Página principal do projeto
   - Links para toda documentação
   - Como usar comandos Claude Code

---

## 📁 Estrutura de Artefatos

### 00_Master/ - Documentos Master
Você está aqui! Documentos centrais do projeto.

```
00_Master/
├── INDICE_GERAL.md          # Este documento
├── RESUMO_EXECUTIVO.md      # Resumo para stakeholders ⭐
├── KICKOFF.md               # Documento de kickoff
├── DUVIDAS.md               # Dúvidas técnicas
└── SQUAD_ARCHITECTURE.md    # Squad de Arquitetura
```

### 01_Requisitos/ - Requisitos Funcionais
Requisitos extraídos do Manual Bacen, user stories, processos de negócio.

```
01_Requisitos/
├── CRF-001_Checklist_Requisitos.md        # Checklist master
├── MTR-001_Matriz_Rastreabilidade.md      # Matriz de rastreabilidade
├── RNE-001_Regras_Negocio.md              # Regras de negócio
├── UserStories/
│   ├── UST-001_Criar_Chave.md
│   └── ...
└── Processos/
    ├── MPN-001_CRUD_Chaves.md
    └── ...
```

### 02_Arquitetura/ - Arquitetura de Solução
Arquitetura, ADRs, especificações técnicas, diagramas.

```
02_Arquitetura/
├── DAS-001_Arquitetura_Solucao.md         # Doc principal arquitetura
├── MIG-001_Mapa_Integracoes.md            # Mapa de integrações
├── ADRs/
│   ├── ADR-001_Abstracao_Bridge.md
│   └── ...
├── TechSpecs/
│   ├── ETS-001_Core_DICT.md
│   └── ...
└── Diagramas/
    ├── C4-001_Contexto.mmd
    └── ...
```

### 03_Dados/ - Modelos de Dados
Modelagem de dados, eventos, cache.

```
03_Dados/
├── MDC-001_Modelo_Conceitual.md           # Modelo conceitual
├── MDL-001_Modelo_Logico.md               # Modelo lógico
├── MDF-001_Modelo_Fisico.md               # Modelo físico
├── SEV-001_Eventos_Dominio.md             # Eventos de domínio
└── ECA-001_Estrategia_Cache.md            # Estratégia de cache
```

### 04_APIs/ - Especificações de APIs
Catálogo de APIs, contratos gRPC, OpenAPI.

```
04_APIs/
├── CAB-001_Catalogo_APIs_Bacen.md         # APIs do Bacen
├── EAI-001_APIs_Core_DICT.md              # APIs internas
├── MSA-001_Matriz_Sync_Async.md           # Matriz sync/async
└── gRPC/
    ├── CGR-001_CreateKey.md
    └── ...
```

### 05_Frontend/ - Especificações de Frontend
Funcionalidades, jornadas, componentes, wireframes.

```
05_Frontend/
├── LFF-001_Lista_Funcionalidades.md       # Lista de funcionalidades
├── MFB-001_Matriz_Frontend_Backend.md     # Matriz FE-BE
├── Jornadas/
│   ├── MJU-001_Cadastro_Chave.md
│   └── ...
├── Componentes/
│   ├── ECO-001_ChaveForm.md
│   └── ...
└── Wireframes/
    ├── WFR-001_Dashboard.md
    └── ...
```

### 12_Integracao/ - Integração E2E
Connect, Bridge, fluxos end-to-end, resiliência.

```
12_Integracao/
├── ECD-001_Especificacao_Connect.md       # Spec Connect DICT
├── EBD-001_Especificacao_Bridge.md        # Spec Bridge DICT
├── PDR-001_Padroes_Resiliencia.md         # Padrões de resiliência
├── Fluxos/
│   ├── MFE-001_Criar_Chave_E2E.md
│   └── ...
└── Sequencias/
    ├── DSQ-001_CreateKey.mmd
    └── ...
```

### 13_Seguranca/ - Segurança
Análise de segurança, controles, políticas.

```
13_Seguranca/
├── ASG-001_Analise_Seguranca.md           # Análise geral
├── RSG-001_Requisitos_Seguranca.md        # Requisitos
├── MCS-001_Matriz_Controles.md            # Controles de segurança
├── PRL-001_Politica_Rate_Limiting.md      # Rate limiting
└── PPF-001_Prevencao_Fraudes.md           # Prevenção fraudes
```

### 08_Testes/ - Testes
Estratégia de testes, casos de teste, homologação.

```
08_Testes/
├── EST-001_Estrategia_Testes.md           # Estratégia geral
├── PTH-001_Plano_Homologacao.md           # Plano homologação Bacen
├── MCO-001_Matriz_Cobertura.md            # Matriz de cobertura
├── ETA-001_Testes_Automatizados.md        # Testes automatizados
└── Casos/
    ├── CTS-001_Criar_Chave.md
    └── ...
```

### 09_DevOps/ - DevOps
CI/CD, ambientes, pipelines, monitoramento.

```
09_DevOps/
├── ECD-001_Estrategia_CICD.md             # Estratégia CI/CD
├── EAM-001_Ambientes.md                   # Ambientes
├── EMO-001_Monitoramento.md               # Monitoramento
├── GWF-001_Git_Workflow.md                # Git workflow
└── Pipelines/
    ├── PPL-001_Build.yaml
    └── ...
```

### 10_Compliance/ - Compliance e Homologação
Checklist de homologação, conformidade, gaps.

```
10_Compliance/
├── CHO-001_Checklist_Homologacao.md       # Checklist homologação
├── MCF-001_Matriz_Conformidade.md         # Matriz de conformidade
├── AGA-001_Analise_Gaps.md                # Análise de gaps
├── RRE-001_Requisitos_Regulatorios.md     # Requisitos regulatórios
└── PAU-001_Plano_Auditoria.md             # Plano de auditoria
```

### 11_Gestao/ - Gestão de Projeto
Planos, status reports, backlog, sprints.

```
11_Gestao/
├── PMP-001_Plano_Master_Projeto.md        # Plano master
├── CRN-001_Cronograma.md                  # Cronograma
├── MRK-001_Matriz_Riscos.md               # Riscos
├── Status_Reports/
│   ├── RST-20251024.md
│   └── ...
├── Backlog/
│   ├── BKL-001_Master.md
│   └── ...
├── Sprints/
│   ├── SPL-001_Sprint01.md
│   └── ...
├── Retrospectivas/
│   ├── RET-001_Sprint01.md
│   └── ...
└── Checklists/
    ├── CHA-001_Artefatos.md
    └── ...
```

### 99_Templates/ - Templates
Templates reutilizáveis para criação de artefatos.

```
99_Templates/
├── TPL-UserStory.md                       # Template User Story
├── TPL-ADR.md                             # Template ADR
├── TPL-TechSpec.md                        # Template Spec Técnica
└── TPL-CasoTeste.md                       # Template Caso Teste
```

---

## 📚 Documentação de Input (Docs_iniciais/)

Documentação fornecida pelo Bacen e LBPay:

```
Docs_iniciais/
├── manual_Operacional_DICT_Bacen.md       # Manual oficial Bacen
├── OpenAPI_Dict_Bacen.json                # API spec Bacen
├── Requisitos_Homologação_Dict.md         # Requisitos homologação
├── ArquiteturaDict_LBPAY.md               # Arquitetura LBPay (C4)
├── Backlog(Plano DICT).csv                # Backlog inicial
└── guidelines2IA.md                       # Guidelines do projeto
```

---

## 🛠️ Comandos Claude Code (.claude/)

Comandos customizados para interagir com os agentes:

```
.claude/
├── commands/
│   ├── pm-status.md                       # Status do projeto
│   ├── arch-analysis.md                   # Análise de arquitetura
│   ├── req-check.md                       # Checklist de requisitos
│   ├── tech-spec.md                       # Gerar spec técnica
│   ├── gen-docs.md                        # Gerar documentação
│   └── update-checklist.md                # Atualizar checklists
└── README.md                              # Sobre Claude Code
```

**Comandos disponíveis**:
- `/pm-status` - Status geral do projeto
- `/arch-analysis` - Análise de arquitetura
- `/req-check` - Verificar requisitos
- `/tech-spec` - Gerar especificação técnica
- `/gen-docs` - Gerar/consolidar documentação
- `/update-checklist` - Atualizar checklists

---

## 🎯 Navegação Rápida

### Por Papel/Stakeholder

**CTO**:
1. [RESUMO_EXECUTIVO.md](./RESUMO_EXECUTIVO.md) ⭐
2. [PMP-001 - Plano Master](../11_Gestao/PMP-001_Plano_Master_Projeto.md)
3. [DAS-001 - Arquitetura](../02_Arquitetura/) (quando criado)
4. [ADRs - Decisões Arquiteturais](../02_Arquitetura/ADRs/) (quando criados)

**Head de Arquitetura**:
1. [SQUAD_ARCHITECTURE.md](../SQUAD_ARCHITECTURE.md)
2. [DUVIDAS.md](./DUVIDAS.md) - Dúvidas arquiteturais
3. [DAS-001 - Arquitetura](../02_Arquitetura/) (quando criado)
4. [ADRs](../02_Arquitetura/ADRs/) (quando criados)
5. [Specs Técnicas](../02_Arquitetura/TechSpecs/) (quando criadas)

**Head de Produto**:
1. [RESUMO_EXECUTIVO.md](./RESUMO_EXECUTIVO.md)
2. [CRF-001 - Requisitos](../01_Requisitos/) (quando criado)
3. [User Stories](../01_Requisitos/UserStories/) (quando criadas)
4. [LFF-001 - Frontend](../05_Frontend/) (quando criado)

**Head de Engenharia**:
1. [AST-001 - Stack Tecnológica](../02_Arquitetura/) (quando criado)
2. [ECD-001 - CI/CD](../09_DevOps/) (quando criado)
3. [EST-001 - Estratégia de Testes](../08_Testes/) (quando criado)
4. [DUVIDAS.md](./DUVIDAS.md) - Dúvidas técnicas

### Por Fase do Projeto

**Agora (Setup)**:
- ✅ [RESUMO_EXECUTIVO.md](./RESUMO_EXECUTIVO.md)
- ✅ [KICKOFF.md](./KICKOFF.md)
- ✅ [PMP-001](../11_Gestao/PMP-001_Plano_Master_Projeto.md)
- ✅ [SQUAD_ARCHITECTURE.md](../SQUAD_ARCHITECTURE.md)
- ✅ [DUVIDAS.md](./DUVIDAS.md)

**Sprint 1-2 (Semanas 1-2)**:
- [ ] CRF-001 - Checklist Requisitos
- [ ] CAB-001 - Catálogo APIs Bacen
- [ ] AST-001 - Análise Stack
- [ ] ARE-XXX - Análise Repositórios

**Sprint 3-4 (Semanas 3-4)**:
- [ ] DAS-001 - Arquitetura Solução
- [ ] MDC/MDL/MDF-001 - Modelos de Dados
- [ ] ADR-XXX - ADRs
- [ ] MIG-001 - Mapa Integrações

**Sprint 5-6 (Semanas 5-6)**:
- [ ] UST-XXX - User Stories
- [ ] EAI/CGR-XXX - Specs APIs
- [ ] LFF-001 - Frontend
- [ ] EST-001 - Testes

**Sprint 7-8 (Semanas 7-8)**:
- [ ] BKL-001 - Backlog
- [ ] IMD-001 - Índice Master
- [ ] PAP-XXX - Pacotes Aprovação
- [ ] Squad Fase 2

---

## 📊 Status Atual

**Data**: 2025-10-24
**Fase**: Fase 1 - Especificação e Planejamento
**Sprint**: Pré-Sprint (Setup)
**Status**: ✅ Setup Completo - Aguardando Kickoff

### Progresso Geral
- **Fase 1**: 0% (aguardando kickoff)
- **Setup**: 100% ✅
- **Artefatos Criados**: 20 documentos de setup
- **Artefatos Pendentes**: 100+ artefatos de especificação

### Próximos Marcos
1. ⏳ Aprovação de Kickoff
2. ⏳ Sprint 1 Planning
3. ⏳ Início de Sprint 1

---

## 🔍 Como Encontrar Informações

### Por Tipo de Informação

**Requisitos Funcionais**:
→ `/01_Requisitos/CRF-001_Checklist_Requisitos.md`

**Arquitetura**:
→ `/02_Arquitetura/DAS-001_Arquitetura_Solucao.md`

**APIs**:
→ `/04_APIs/` - CAB-001 (Bacen), EAI-XXX (internas)

**Decisões Arquiteturais**:
→ `/02_Arquitetura/ADRs/ADR-XXX_*.md`

**User Stories**:
→ `/01_Requisitos/UserStories/UST-XXX_*.md`

**Testes**:
→ `/08_Testes/` - EST-001 (estratégia), CTS-XXX (casos)

**Homologação**:
→ `/10_Compliance/CHO-001_Checklist_Homologacao.md`

**Status do Projeto**:
→ `/11_Gestao/Status_Reports/RST-YYYYMMDD.md`

**Dúvidas**:
→ `/00_Master/DUVIDAS.md`

---

## 📞 Suporte e Contatos

**Project Manager**: PHOENIX (AGT-PM-001)
**Scrum Master**: CATALYST (AGT-SM-001)

**Para dúvidas técnicas**: Adicionar em [DUVIDAS.md](./DUVIDAS.md)
**Para questões de projeto**: Contatar PHOENIX

---

## 📝 Convenções

### Nomenclatura de Artefatos
- **CRF**: Checklist Requisitos Funcionais
- **UST**: User Story
- **DAS**: Documento Arquitetura Solução
- **ADR**: Architecture Decision Record
- **ETS**: Especificação Técnica
- **MDC/MDL/MDF**: Modelo de Dados (Conceitual/Lógico/Físico)
- **CAB**: Catálogo APIs Bacen
- **EAI**: Especificação APIs Internas
- **CGR**: Contrato gRPC
- **E muitos outros...**

### Status de Artefatos
- ⬜ **Not Started**: Não iniciado
- 🟡 **Draft**: Em elaboração
- 🟠 **Review**: Em revisão
- 🔵 **Approved**: Aprovado
- 🟢 **Final**: Finalizado

---

## 🎓 Glossário Rápido

**DICT**: Diretório de Identificadores de Contas Transacionais
**PIX**: Sistema de Pagamentos Instantâneos do Brasil
**Bacen**: Banco Central do Brasil
**PSP**: Participante do Sistema PIX
**ADR**: Architecture Decision Record
**E2E**: End-to-End (ponta a ponta)
**gRPC**: Google Remote Procedure Call
**mTLS**: Mutual TLS (autenticação mútua)

---

**Última Atualização**: 2025-10-24
**Versão**: 1.0
**Responsável**: PHOENIX (AGT-PM-001)
