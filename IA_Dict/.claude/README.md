# Squad Claude Code - Projeto DICT LBPay

## Visão Geral
Este projeto visa implementar a solução DICT (Diretório de Identificadores de Contas Transacionais) para homologação no Banco Central do Brasil.

## Estrutura do Projeto

### Fase 1: Especificação e Planejamento (Atual)
- Squad de Arquitetura: Responsável por toda a idealização e especificação detalhada
- Objetivo: Criar todos os artefatos necessários para implementação autônoma por agentes

### Fase 2: Implementação
- Squad de Desenvolvimento: Será definida após aprovação dos artefatos da Fase 1
- Objetivo: Implementação completa e homologação no DICT Bacen

## Squads Definidas

### Squad de Arquitetura (Fase 1)
Consulte [SQUAD_ARCHITECTURE.md](../Artefatos/SQUAD_ARCHITECTURE.md) para detalhes completos.

## Comandos Disponíveis

### Gestão de Projeto
- `/pm-status` - Status geral do projeto e progresso
- `/pm-risks` - Análise de riscos e impedimentos
- `/pm-planning` - Planejamento e cronograma

### Squad de Arquitetura
- `/arch-analysis` - Análise de arquitetura e requisitos
- `/req-check` - Verificar requisitos funcionais
- `/tech-spec` - Gerar especificação técnica

### Documentação
- `/gen-docs` - Gerar documentação de artefatos
- `/update-checklist` - Atualizar checklists de requisitos

## Princípios do Projeto

1. **Indexação Universal**: Todos os requisitos, funcionalidades, processos e tarefas são numerados e indexados
2. **Rastreabilidade End-to-End**: Requisito → Processo → Front-End → Core → Bridge → Bacen
3. **Aprovação Formal**: Todos os artefatos devem ser aprovados antes da implementação
4. **Agentes Autônomos**: Especificações devem permitir implementação 100% autônoma
5. **Critérios de Aceitação**: Definidos para cada funcionalidade baseados em requisitos do Bacen

## Repositórios Envolvidos

- **DICT Connector**: https://github.com/lb-conn/connector-dict
- **Bridge**: https://github.com/lb-conn/rsfn-connect-bacen-bridge
- **Simulador**: https://github.com/lb-conn/simulator-dict
- **Core Banking**: https://github.com/london-bridge/money-moving
- **Orchestration**: https://github.com/london-bridge/orchestration-go
- **Operations**: https://github.com/london-bridge/operation
- **Contracts**: https://github.com/london-bridge/lb-contracts

## Documentação de Referência

### Documentos Bacen
- Manual Operacional DICT
- OpenAPI DICT
- Requisitos de Homologação

### Documentos LBPay
- Arquitetura DICT (C4 - IcePanel)
- Backlog e Plano DICT
- Guidelines para IA

## Próximos Passos

1. Criar todos os agentes especializados da Squad de Arquitetura
2. Analisar documentação técnica do Bacen
3. Gerar checklist completo de requisitos funcionais
4. Mapear arquitetura e fluxos end-to-end
5. Criar backlog detalhado e indexado
6. Submeter artefatos para aprovação

## Contatos e Aprovadores

- **CTO**: Aprovação final de especificações
- **Head de Arquitetura**: Aprovação de arquitetura e design
- **Head de Produto**: Aprovação de requisitos funcionais
- **Head de Engenharia**: Aprovação de stack e implementação
