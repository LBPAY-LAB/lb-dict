# ADRs - Architecture Decision Records

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Última Atualização**: 2025-10-25

---

## Índice de ADRs

| ADR | Título | Data | Status | Decisão |
|-----|--------|------|--------|---------|
| [ADR-001](ADR-001_Message_Broker_Apache_Pulsar.md) | Message Broker - Apache Pulsar | 2025-10-24 | ✅ Aprovado | Apache Pulsar para mensageria assíncrona |
| [ADR-002](ADR-002_Orchestrator_Temporal_Workflow.md) | Orchestrator - Temporal Workflows | 2025-10-24 | ✅ Aprovado | Temporal para workflows de longa duração |
| [ADR-003](ADR-003_Protocol_gRPC.md) | Protocol - gRPC | 2025-10-24 | ✅ Aprovado | gRPC para comunicação síncrona de alta performance |
| [ADR-004](ADR-004_Cache_Redis.md) | Cache - Redis | 2025-10-24 | ✅ Aprovado | Redis para cache distribuído |
| [ADR-005](ADR-005_Database_PostgreSQL.md) | Database - PostgreSQL | 2025-10-24 | ✅ Aprovado | PostgreSQL para persistência de dados |
| [ADR-006](ADR-006_Consolidacao_Core_DICT.md) | Consolidação Core DICT | 2025-10-24 | ✅ Aprovado | Centralização da lógica DICT em módulo único |
| [ADR-007](ADR-007_Performance_Multi_Camadas.md) | Performance Multi-Camadas | 2025-10-24 | ✅ Aprovado | Otimizações de performance em camadas |
| [ADR-008](ADR-008_Bridge_DICT_Dedicado.md) | Bridge DICT Dedicado | 2025-10-24 | ✅ Aprovado | Bridge dedicado para comunicação com Bacen |

---

## Categorias

### Infraestrutura e Comunicação
- **ADR-001**: Apache Pulsar (Mensageria)
- **ADR-003**: gRPC (Protocolo síncrono)
- **ADR-008**: Bridge DICT (Adapter Bacen)

### Dados e Estado
- **ADR-004**: Redis (Cache)
- **ADR-005**: PostgreSQL (Database)

### Arquitetura e Orquestração
- **ADR-002**: Temporal Workflows (Orquestração)
- **ADR-006**: Consolidação Core DICT (Centralização)
- **ADR-007**: Performance Multi-Camadas (Otimização)

---

## Status das Decisões

- ✅ **Aprovado**: 8 ADRs
- 🟡 **Em Revisão**: 0 ADRs
- 🔴 **Rejeitado**: 0 ADRs
- ⏳ **Proposto**: 0 ADRs

---

## Processo de ADR

### Como Criar um Novo ADR

1. Copiar template: [TPL-ADR.md](../../99_Templates/TPL-ADR.md)
2. Numerar sequencialmente (próximo: ADR-009)
3. Preencher todas as seções
4. Submeter para revisão (Head de Arquitetura)
5. Após aprovação, mover para esta pasta

### Estrutura de um ADR

```markdown
# ADR-XXX: [Título da Decisão]

## Status
[Proposto | Em Revisão | Aprovado | Rejeitado | Obsoleto]

## Contexto
[Por que precisamos tomar esta decisão?]

## Decisão
[O que decidimos fazer?]

## Consequências
[Quais são os impactos (positivos e negativos)?]

## Alternativas Consideradas
[Quais outras opções foram avaliadas?]

## Referências
[Links para docs, discussions, etc.]
```

---

## Mapeamento com Especificações Técnicas

| ADR | TEC Document | Seção |
|-----|--------------|-------|
| ADR-001 (Pulsar) | TEC-003 v2.1 | Stack Tecnológica, Pulsar Consumer/Producer |
| ADR-002 (Temporal) | TEC-003 v2.1 | Temporal Workflows (ClaimWorkflow 30 dias) |
| ADR-003 (gRPC) | TEC-002 v3.1, TEC-003 v2.1 | Dual Protocol Support, Bridge Client |
| ADR-004 (Redis) | TEC-003 v2.1 | Stack Tecnológica (Redis v9.14.1) |
| ADR-005 (PostgreSQL) | TEC-001, DAT-001 | State Management, Database Schemas |
| ADR-006 (Core DICT) | TEC-001 | Arquitetura, Consolidação |
| ADR-008 (Bridge) | TEC-002 v3.1 | Bridge Architecture, SOAP/mTLS Adapter |

---

## Próximos ADRs Planejados

| ID | Título | Prioridade | Responsável |
|----|--------|-----------|-------------|
| ADR-009 | Observability - OpenTelemetry | 🟡 Média | Architect |
| ADR-010 | Secret Management - Vault | 🔴 Alta | Security Lead |
| ADR-011 | CI/CD Pipeline Strategy | 🟡 Média | DevOps Lead |
| ADR-012 | Testing Strategy (Unit/Integration/E2E) | 🟡 Média | QA Lead |
| ADR-013 | LGPD Data Protection | 🔴 Alta | Compliance + Architect |
| ADR-014 | Multi-Tenancy Strategy (ISPB isolation) | 🟢 Baixa | Architect |

---

**Responsável por ADRs**: Head de Arquitetura (Thiago Lima)
**Revisores**: CTO (José Luís Silva), Tech Leads
