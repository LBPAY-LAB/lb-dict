# Processos de Negócio

**Propósito**: Documentação de processos de negócio relacionados ao DICT

## 📋 Conteúdo

Esta pasta armazenará:

- **Business Process Models**: Modelos BPMN de processos de negócio
- **Process Flows**: Fluxos de processo detalhados
- **Business Rules**: Regras de negócio formalizadas
- **SLAs**: Service Level Agreements para processos

## 📁 Estrutura Esperada

```
Processos/
├── BPMN/
│   ├── PROC_001_Cadastro_Chave_DICT.bpmn
│   ├── PROC_002_Reivindicacao_Chave.bpmn
│   ├── PROC_003_Portabilidade_Conta.bpmn
│   └── PROC_004_Exclusao_Chave.bpmn
├── Business_Rules/
│   ├── BR_001_Validacao_Chave_PIX.md
│   ├── BR_002_Claim_30_Dias.md
│   └── BR_003_Limite_Chaves_Por_Usuario.md
└── SLAs/
    ├── SLA_Cadastro_Chave.md
    └── SLA_Resolucao_Claim.md
```

## 🎯 Principais Processos

### PROC-001: Cadastro de Chave DICT

**Descrição**: Processo de cadastro de nova chave PIX no DICT

**Atores**:
- Cliente (titular da conta)
- Sistema DICT (Core + Connect + Bridge)
- Bacen DICT (sistema externo)

**Fluxo Principal**:
1. Cliente acessa app do banco
2. Cliente seleciona "Cadastrar chave PIX"
3. Cliente escolhe tipo de chave (CPF, CNPJ, Phone, Email, EVP)
4. Sistema valida dados do cliente
5. Sistema valida disponibilidade da chave no Bacen
6. Sistema cria entry no DICT local
7. Sistema envia CreateEntry ao Bacen via Bridge
8. Bacen confirma criação
9. Sistema notifica cliente (push + email)

**Fluxos Alternativos**:
- **FA1**: Chave já existe no Bacen → Rejeitar com mensagem "Chave já cadastrada"
- **FA2**: Limite de chaves excedido (5 chaves) → Rejeitar
- **FA3**: Erro de comunicação com Bacen → Retry (3x) → Se falhar, notificar cliente

**Regras de Negócio**:
- **BR-001**: CPF deve ser válido (dígitos verificadores)
- **BR-002**: Telefone deve estar no formato +55 (11 dígitos)
- **BR-003**: Email deve ser válido (formato RFC 5322)
- **BR-004**: Usuário pode ter no máximo 5 chaves DICT
- **BR-005**: Chave EVP é gerada automaticamente (UUID v4)

**SLA**:
- Tempo de processamento: < 10 segundos (p95)
- Disponibilidade: 99.9%

---

### PROC-002: Reivindicação de Chave (Claim)

**Descrição**: Processo de reivindicação de chave já cadastrada por outro usuário

**Período**: 30 dias (TEC-003 v2.1)

**Atores**:
- Reivindicador (claimer)
- Dono atual da chave (owner)
- Sistema DICT (Connect + Temporal Workflows)
- Bacen DICT

**Fluxo Principal**:
1. Reivindicador solicita claim de chave existente
2. Sistema valida que chave existe e pertence a outro usuário
3. Sistema cria Claim no Bacen via Bridge
4. Sistema inicia ClaimWorkflow no Temporal (30 dias)
5. Sistema notifica dono atual da chave
6. Dono tem 30 dias para aceitar ou rejeitar claim
7. **Se dono aceitar**: Chave transferida para reivindicador
8. **Se dono rejeitar**: Claim cancelada
9. **Se 30 dias expirar**: Chave automaticamente transferida

**Regras de Negócio**:
- **BR-010**: Claim tem período fixo de 30 dias
- **BR-011**: Dono pode aceitar ou rejeitar claim a qualquer momento
- **BR-012**: Se dono não responder em 30 dias, claim é automaticamente aceita
- **BR-013**: Apenas 1 claim ativa por chave

**SLA**:
- Notificação ao dono: < 1 minuto após criação
- Processamento de aceite/rejeição: < 5 segundos

---

### PROC-003: Portabilidade de Conta

**Descrição**: Processo de mudança de conta vinculada a uma chave DICT

**Atores**:
- Cliente (titular da chave)
- Sistema DICT
- Bacen DICT

**Fluxo Principal**:
1. Cliente solicita portabilidade de chave para nova conta
2. Sistema valida que cliente é dono da chave
3. Sistema valida nova conta
4. Sistema envia ConfirmPortability ao Bacen
5. Bacen confirma portabilidade
6. Sistema atualiza entry com nova conta
7. Sistema notifica cliente

**Regras de Negócio**:
- **BR-020**: Apenas dono da chave pode solicitar portabilidade
- **BR-021**: Nova conta deve ser válida (ISPB + agência + conta)
- **BR-022**: Portabilidade é imediata (não tem período de espera)

**SLA**:
- Tempo de processamento: < 10 segundos

## 📊 BPMN Diagram Example

```
[Cliente] → [Solicita Cadastro] → [Sistema Valida] → [Bacen Confirma] → [Notifica Cliente]
                                        ↓
                                   [Chave Existe?]
                                        ↓
                                   [Rejeitar]
```

## 📚 Referências

- [User Stories](../UserStories/)
- [TEC-003: ClaimWorkflow 30 dias](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [Fluxos de Integração](../../12_Integracao/Fluxos/)
- [Diagramas de Sequência](../../12_Integracao/Sequencias/)

---

**Status**: 🔴 Pasta vazia (será preenchida na Fase 2)
**Fase de Preenchimento**: Fase 2 (modelagem de processos)
**Ferramenta**: Camunda Modeler, Bizagi, Draw.io
**Responsável**: Business Analyst + Product Owner
