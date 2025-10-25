# User Story: [Título Conciso]
**ID**: UST-XXX
**Data de Criação**: [YYYY-MM-DD]
**Criado por**: ORACLE (AGT-BA-001)
**Status**: [Draft | Review | Approved | Implemented | Tested]
**Prioridade**: [Alta | Média | Baixa]

## Informações de Rastreabilidade
- **Requisito Funcional**: RF-XXX
- **Bloco Funcional**: [Bloco 1-6]
- **Manual Bacen**: Seção [X.Y.Z]
- **Homologação**: [Sim | Não]
- **Sprint**: [número ou TBD]

## User Story

**Como** [tipo de usuário/persona]
**Quero** [realizar uma ação/ter uma funcionalidade]
**Para** [alcançar um benefício/objetivo]

## Contexto de Negócio
[Explicação detalhada do contexto de negócio e por que esta funcionalidade é necessária. Incluir requisitos do Bacen se aplicável.]

## Personas Envolvidas
- **Usuário Final**: [descrição]
- **Operador LBPay**: [se aplicável]
- **Sistema**: [sistemas envolvidos]

## Regras de Negócio
- **RN-XXX-1**: [Regra de negócio 1]
- **RN-XXX-2**: [Regra de negócio 2]
- **RN-XXX-3**: [Regra de negócio 3]

## Validações
- **VAL-XXX-1**: [Validação 1 - campo, tipo, formato]
- **VAL-XXX-2**: [Validação 2]
- **VAL-XXX-3**: [Validação 3]

## Critérios de Aceitação

### Cenário 1: [Nome do Cenário - Caminho Feliz]
**Dado que** [contexto/pré-condições]
**Quando** [ação do usuário]
**Então** [resultado esperado]
**E** [resultado adicional se aplicável]

### Cenário 2: [Nome do Cenário - Validação de Erro]
**Dado que** [contexto]
**Quando** [ação que causa erro]
**Então** [comportamento esperado de erro]
**E** [mensagem/código de erro específico]

### Cenário 3: [Nome do Cenário - Caso Especial]
[continuar conforme necessário...]

## Fluxo de Interface (Frontend)
1. [Passo 1 da interação do usuário]
2. [Passo 2]
3. [Passo 3]
4. [Resultado final]

**Wireframe**: [link para WFR-XXX se existir]

## Fluxo Técnico (Backend)

### Comunicação
- **Tipo**: [Síncrona | Assíncrona]
- **Protocolo**: [gRPC | REST | Event]

### Componentes Envolvidos
1. **Frontend** → Chama API Core DICT
2. **Core DICT** → [processamento/validações]
3. **Connect DICT** → [transformação/roteamento]
4. **Bridge DICT** → Interface com Bacen
5. **DICT Bacen** → [processamento Bacen]

### APIs Relacionadas
- **Frontend → Core**: [EAI-XXX] `gRPC Service.Method`
- **Core → Connect**: [EIF-XXX] `gRPC Service.Method`
- **Bridge → Bacen**: [CAB-XXX] `POST /api/v1/entries`

**Diagrama de Sequência**: [link para DSQ-XXX]

## Dados Necessários

### Input (Request)
```json
{
  "campo1": "valor exemplo",
  "campo2": "valor exemplo",
  "campo3": {
    "subcampo": "valor"
  }
}
```

### Output (Response - Sucesso)
```json
{
  "id": "uuid",
  "status": "CREATED",
  "data": {
    "campo": "valor"
  }
}
```

### Output (Response - Erro)
```json
{
  "error": {
    "code": "ERR_XXX",
    "message": "Mensagem descritiva",
    "details": []
  }
}
```

## Modelo de Dados
- **Entidades Afetadas**: [Entity1, Entity2]
- **Eventos Gerados**: [EventXxxCreated, EventYyyUpdated]
- **Modelo de Dados**: [link para MDC/MDL/MDF-XXX]

## Requisitos Não-Funcionais

### Performance
- **Latência Máxima**: [XXms]
- **Throughput**: [YY req/s]

### Segurança
- **Autenticação**: [JWT | OAuth2 | mTLS]
- **Autorização**: [roles/permissions necessários]
- **Dados Sensíveis**: [PII? Sim/Não - quais campos]
- **Auditoria**: [Sim | Não]

### Resiliência
- **Retry Policy**: [sim/não - quantas tentativas]
- **Timeout**: [XXs]
- **Circuit Breaker**: [sim/não]
- **Idempotência**: [sim/não - como garantir]

## Dependências

### Funcionalidades Dependentes
- **Depende de**: [UST-YYY, UST-ZZZ]
- **Pré-requisitos**: [lista de pré-requisitos]

### Dependências Técnicas
- **Serviços**: [serviços externos necessários]
- **Dados**: [dados que devem existir previamente]
- **Configurações**: [configurações necessárias]

## Estimativa
- **Story Points**: [pontos - Fibonacci: 1,2,3,5,8,13,21]
- **Complexidade**: [Baixa | Média | Alta]
- **Estimativa de Horas**: [XX-YY horas]

## Casos de Teste Relacionados
- [CTS-XXX-1]: [Nome do caso de teste 1]
- [CTS-XXX-2]: [Nome do caso de teste 2]
- [CTS-XXX-3]: [Nome do caso de teste 3]

## Documentação Bacen
- **Manual Operacional**: Seção [X.Y.Z], Página [N]
- **OpenAPI Endpoint**: `[METHOD] /api/v1/path`
- **Requisitos Homologação**: [HOM-XXX]

## Notas de Implementação
[Dicas, considerações especiais, edge cases, observações técnicas importantes para os desenvolvedores]

## Riscos Identificados
| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| [Risco 1] | [Alta/Média/Baixa] | [Alto/Médio/Baixo] | [Como mitigar] |

## Checklist de Conclusão
- [ ] User story revisada e aprovada
- [ ] Especificação técnica criada (ETS-XXX)
- [ ] Wireframes criados (se aplicável)
- [ ] Casos de teste escritos
- [ ] Implementação concluída
- [ ] Testes unitários passando
- [ ] Testes de integração passando
- [ ] Code review aprovado
- [ ] Testes de aceitação passando
- [ ] Documentação atualizada
- [ ] Deploy em ambiente de teste
- [ ] Validação com stakeholders

## Histórico de Alterações
| Data | Versão | Autor | Mudanças |
|------|--------|-------|----------|
| [YYYY-MM-DD] | 1.0 | ORACLE | Criação inicial |

## Aprovações
- [ ] **Business Analyst (ORACLE)**: [data]
- [ ] **Product Owner**: [data]
- [ ] **Solution Architect (NEXUS)**: [data - aprovação técnica]
- [ ] **Tech Lead**: [data]
