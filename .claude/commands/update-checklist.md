---
description: Atualizar checklists de progresso e conformidade do projeto
---

# Comando: Update Checklist

Modo colaborativo: **PHOENIX** (PM) + **GUARDIAN** (Compliance) + **CATALYST** (SM)

## Tipos de Checklist

### 1. Checklist de Requisitos Funcionais
Atualizar progresso de implementação de requisitos:
```
/Artefatos/01_Requisitos/CRF-001_Checklist_Requisitos.md
```

### 2. Checklist de Homologação
Atualizar conformidade com requisitos Bacen:
```
/Artefatos/10_Compliance/CHO-001_Checklist_Homologacao.md
```

### 3. Checklist de Artefatos
Atualizar status de criação de artefatos:
```
/Artefatos/11_Gestao/Checklists/CHA-001_Artefatos.md
```

### 4. Checklist de Testes
Atualizar execução de casos de teste:
```
/Artefatos/08_Testes/MCO-001_Matriz_Cobertura.md
```

## Tarefas

### Para Checklist de Requisitos (PHOENIX + ORACLE)
1. Ler CRF-001_Checklist_Requisitos.md
2. Para cada requisito (RF-XXX):
   - Verificar se há artefatos relacionados
   - Verificar se há código implementado
   - Verificar se há testes
   - Atualizar status: [Not Started | In Progress | Completed | Blocked]
3. Calcular percentual de completude por bloco
4. Identificar requisitos bloqueados
5. Atualizar data da última revisão

### Para Checklist de Homologação (GUARDIAN)
1. Ler Requisitos_Homologação_Dict.md (Bacen)
2. Ler CHO-001_Checklist_Homologacao.md
3. Para cada requisito de homologação:
   - Verificar implementação
   - Verificar evidências
   - Verificar testes de conformidade
   - Atualizar status
4. Identificar gaps críticos para homologação
5. Gerar relatório de prontidão para homologação

### Para Checklist de Artefatos (SCRIBE + PHOENIX)
1. Varrer estrutura /Artefatos/
2. Listar todos os artefatos esperados vs criados
3. Para cada artefato:
   - Status: [Pending | Draft | Review | Approved | Final]
   - Responsável (agente)
   - Data última atualização
   - Aprovador
   - Status de aprovação
4. Identificar artefatos faltantes
5. Calcular progresso geral

### Para Checklist de Testes (VALIDATOR)
1. Ler MCO-001_Matriz_Cobertura.md
2. Para cada caso de teste:
   - Status: [Not Run | Running | Passed | Failed | Blocked]
   - Cobertura de código
   - Cobertura de requisitos
3. Calcular métricas:
   - % testes executados
   - % testes passed
   - % cobertura de código
   - % cobertura de requisitos

## Template: Checklist de Requisitos

```markdown
# Checklist de Requisitos Funcionais
**ID**: CRF-001
**Última Atualização**: [data]
**Responsável**: ORACLE (AGT-BA-001)
**Status Geral**: X% completo

## Resumo Executivo
- **Total de Requisitos**: X
- **Completos**: Y (Z%)
- **Em Progresso**: A (B%)
- **Bloqueados**: C (D%)
- **Não Iniciados**: E (F%)

## Bloco 1: CRUD de Chaves
**Progresso**: X/Y (Z%)

### RF-001: Criar Chave
- **Status**: ✅ Completed | 🟡 In Progress | ⬜ Not Started | 🔴 Blocked
- **Prioridade**: Alta
- **Homologação**: Sim
- **Artefatos**:
  - [x] UST-001: User Story criada
  - [x] ETS-001: Especificação técnica
  - [ ] Implementação
  - [ ] Testes unitários
  - [ ] Testes de integração
  - [ ] Documentação
- **Responsável Impl**: [TBD]
- **Bloqueadores**: [se houver]
- **Observações**: [notas]

### RF-002: Consultar Chave
[similar...]

## Bloco 2: Reivindicação
[similar...]

## Riscos Identificados
| ID | Risco | Impacto | Mitigação |
|----|-------|---------|-----------|

## Próximas Ações
1. [Ação prioritária 1]
2. [Ação prioritária 2]
```

## Template: Checklist de Homologação

```markdown
# Checklist de Homologação DICT Bacen
**ID**: CHO-001
**Última Atualização**: [data]
**Responsável**: GUARDIAN (AGT-CM-001)
**Prontidão para Homologação**: X%

## Status Geral
- **Requisitos Obrigatórios**: X/Y (Z%)
- **Requisitos Opcionais**: A/B (C%)
- **Evidências Coletadas**: D/E (F%)
- **Gaps Críticos**: G

## 1. Funcionalidades Obrigatórias

### 1.1 Gestão de Chaves
- [ ] **HOM-001**: Criar chave (todos os tipos)
  - Status: [Implementado | Em Teste | Pendente]
  - Evidência: [link/documento]
  - Teste Bacen: [Passou | Falhou | Não Executado]
  - Observações: [notas]

- [ ] **HOM-002**: Consultar chave
  - [similar...]

### 1.2 Reivindicação
[similar...]

## 2. Requisitos de Segurança
- [ ] **HOM-SEC-001**: Prevenção ataques de leitura
- [ ] **HOM-SEC-002**: Rate limiting
- [ ] **HOM-SEC-003**: Autenticação mTLS
[continuar...]

## 3. Requisitos de Resiliência
[similar...]

## 4. Documentação Obrigatória
- [ ] Manual de integração
- [ ] Matriz de responsabilidades
- [ ] Plano de continuidade
[continuar...]

## 5. Testes de Certificação
| ID Teste | Descrição | Status | Resultado | Evidência |
|----------|-----------|--------|-----------|-----------|

## Gaps Críticos
| ID | Gap | Impacto | Plano de Ação | Prazo |
|----|-----|---------|---------------|-------|

## Prontidão por Categoria
- Funcionalidades: X%
- Segurança: Y%
- Resiliência: Z%
- Documentação: A%

## Recomendação
[Pronto para homologação | Necessita ações | Não pronto]

## Próximas Ações para Homologação
1. [Ação crítica 1]
2. [Ação crítica 2]
```

## Outputs
1. Checklists atualizados com status corrente
2. Relatório de progresso
3. Lista de bloqueadores
4. Ações prioritárias

## Métricas a Calcular
- Percentual de completude por bloco
- Velocity (requisitos/sprint)
- Tempo médio por requisito
- Taxa de bloqueio
- Prontidão para homologação

## Alertas
Gerar alertas se:
- Mais de 3 requisitos bloqueados
- Progresso < 10% por semana
- Gaps críticos de homologação identificados
- Atraso em artefatos críticos
