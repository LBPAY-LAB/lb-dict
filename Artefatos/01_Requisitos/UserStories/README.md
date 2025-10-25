# User Stories

**Propósito**: User Stories detalhadas para funcionalidades do sistema DICT

## 📋 Conteúdo

Esta pasta armazenará:

- **User Stories**: Histórias de usuário no formato "Como... Eu quero... Para..."
- **Acceptance Criteria**: Critérios de aceitação para cada story
- **Story Mapping**: Mapeamento de stories por épico/feature
- **Estimation**: Estimativas de story points (Planning Poker)

## 📁 Estrutura Esperada

```
UserStories/
├── Epics/
│   ├── EPIC-001_Claims_30_Dias/
│   │   ├── US-001_Criar_Claim.md
│   │   ├── US-002_Aceitar_Claim.md
│   │   ├── US-003_Rejeitar_Claim.md
│   │   └── US-004_Expirar_Claim_Automatico.md
│   ├── EPIC-002_Gerenciamento_Chaves/
│   │   ├── US-005_Cadastrar_Chave_CPF.md
│   │   ├── US-006_Cadastrar_Chave_Email.md
│   │   ├── US-007_Listar_Minhas_Chaves.md
│   │   └── US-008_Deletar_Chave.md
│   └── EPIC-003_Portabilidade/
│       ├── US-009_Iniciar_Portabilidade.md
│       └── US-010_Confirmar_Portabilidade.md
└── Story_Mapping.md
```

## 🎯 Template de User Story

```markdown
# US-001: Criar Claim de Chave DICT

## User Story

**Como** correntista do banco (reivindicador)
**Eu quero** reivindicar uma chave DICT que pertence a outra conta
**Para** transferir a chave para minha conta

## Acceptance Criteria

### Cenário 1: Claim criada com sucesso
**Dado** que sou correntista autenticado
**E** a chave CPF "12345678900" existe e pertence a outro usuário
**Quando** eu solicito claim dessa chave para minha conta
**Então** o sistema cria uma claim com status "OPEN"
**E** o sistema notifica o dono atual da chave
**E** o sistema exibe mensagem "Claim criada. O dono tem 30 dias para responder"
**E** a claim expira em exatamente 30 dias a partir de agora

### Cenário 2: Chave não existe
**Dado** que sou correntista autenticado
**Quando** eu solicito claim de chave que não existe
**Então** o sistema retorna erro 404 "Chave não encontrada"

### Cenário 3: Chave já é minha
**Dado** que sou correntista autenticado
**E** a chave já pertence à minha conta
**Quando** eu solicito claim dessa chave
**Então** o sistema retorna erro 409 "Você já é dono desta chave"

### Cenário 4: Claim já existe
**Dado** que sou correntista autenticado
**E** já existe uma claim ativa para esta chave
**Quando** eu solicito nova claim
**Então** o sistema retorna erro 409 "Já existe claim ativa para esta chave"

## Business Rules

- **BR-010**: Claim tem período fixo de 30 dias
- **BR-013**: Apenas 1 claim ativa por chave

## Technical Details

### API Endpoint
```
POST /api/v1/claims
```

### Request Body
```json
{
  "entry_id": "550e8400-e29b-41d4-a716-446655440000",
  "claimer_account": {
    "ispb": "00000000",
    "account_number": "12345-6",
    "branch_code": "0001"
  }
}
```

### Response (201 Created)
```json
{
  "claim_id": "c1234567-89ab-cdef-0123-456789abcdef",
  "status": "OPEN",
  "expires_at": "2025-11-24T10:00:00Z",
  "created_at": "2025-10-25T10:00:00Z"
}
```

## Estimation

- **Story Points**: 8
- **Complexity**: Médio-Alto
- **Dependencies**:
  - Entry já cadastrada no DICT
  - Temporal Workflow configurado
  - Bridge integrado com Bacen

## Definition of Done

- [ ] Código desenvolvido e commitado
- [ ] Testes unitários (cobertura > 80%)
- [ ] Testes de integração (E2E)
- [ ] Code review aprovado
- [ ] Documentação atualizada (Swagger)
- [ ] Deploy em staging
- [ ] QA validou em staging
- [ ] PO aprovou

## Related Documents

- [PROC-002: Reivindicação de Chave](../Processos/)
- [TEC-003 v2.1: ClaimWorkflow 30 dias](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [GRPC-001: Bridge gRPC Service](../../04_APIs/gRPC/GRPC-001_Bridge_gRPC_Service.md)
```

## 📊 Story Mapping

### Épico: Claims (30 dias)

| Backbone (Atividade) | Walking Skeleton | Release 1 | Release 2 |
|---------------------|------------------|-----------|-----------|
| **Criar Claim** | US-001 (MVP) | - | - |
| **Notificar Dono** | - | US-011 (Push) | US-012 (Email) |
| **Aceitar/Rejeitar** | US-002, US-003 | - | - |
| **Expiração Automática** | US-004 (MVP) | - | - |
| **Histórico Claims** | - | US-013 | US-014 (Filtros) |

### Priorização (MoSCoW)

- **Must Have** (MVP):
  - US-001: Criar Claim
  - US-002: Aceitar Claim
  - US-003: Rejeitar Claim
  - US-004: Expiração Automática (30 dias)

- **Should Have** (Release 1):
  - US-005: Cadastrar Chave CPF
  - US-006: Cadastrar Chave Email
  - US-007: Listar Minhas Chaves

- **Could Have** (Release 2):
  - US-008: Deletar Chave
  - US-009: Portabilidade
  - US-013: Histórico de Claims

- **Won't Have** (Futuro):
  - US-015: Claims em lote (Admin)
  - US-016: Relatórios de Claims

## 🎲 Estimativa de Story Points

### Escala Fibonacci
- **1 ponto**: Trivial (< 2 horas)
- **2 pontos**: Simples (2-4 horas)
- **3 pontos**: Médio (4-8 horas)
- **5 pontos**: Complexo (1-2 dias)
- **8 pontos**: Muito complexo (2-3 dias)
- **13 pontos**: Épico (quebrar em stories menores)

## 📚 Referências

- [Processos de Negócio](../Processos/)
- [Backlog](../../17_Gestao/Backlog/)
- [Sprints](../../17_Gestao/Sprints/)
- [Jornadas de Usuário](../../08_Frontend/Jornadas/)

---

**Status**: 🔴 Pasta vazia (será preenchida na Fase 2)
**Fase de Preenchimento**: Fase 2 (após especificações técnicas)
**Responsável**: Product Owner + Business Analyst
**Método**: Workshops de refinamento com stakeholders
