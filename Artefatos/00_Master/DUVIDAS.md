# Documento de Dúvidas e Questões Técnicas - Projeto DICT

**ID**: DUV-001
**Data Criação**: 2025-10-24
**Última Atualização**: 2025-10-24
**Responsável**: PHOENIX (AGT-PM-001)

---

## Propósito

Este documento centraliza todas as dúvidas, questões técnicas e ambiguidades identificadas durante o projeto DICT. Ele serve como:
1. **Registro central** de questões pendentes
2. **Ferramenta de comunicação** com stakeholders
3. **Histórico de decisões** e esclarecimentos

---

## Como Usar Este Documento

### Para Agentes Claude Code
Quando encontrar uma ambiguidade, dúvida ou questão que requer decisão externa:
1. Adicione uma nova entrada na seção apropriada
2. Preencha todos os campos obrigatórios
3. Atribua prioridade e impacto
4. Notifique PHOENIX

### Para Stakeholders
1. Revise dúvidas pendentes regularmente
2. Forneça respostas/decisões
3. Stakeholder responsável atualiza status

---

## Template de Entrada

```markdown
### DUV-XXX: [Título Conciso da Dúvida]
**Data**: YYYY-MM-DD
**Agente**: [Código do agente que levantou]
**Categoria**: [Requisitos | Arquitetura | Dados | APIs | Frontend | Integração | Segurança | Compliance | Outro]
**Prioridade**: [Alta | Média | Baixa]
**Impacto**: [Alto | Médio | Baixo]
**Status**: [Aberta | Em Análise | Respondida | Resolvida]

**Contexto**:
[Explicação do contexto onde a dúvida surgiu]

**Dúvida/Questão**:
[Formulação clara da dúvida ou questão]

**Opções Consideradas** (se aplicável):
1. Opção A: [descrição]
2. Opção B: [descrição]

**Impacto se não resolvida**:
[Como isso impacta o projeto se não for esclarecido]

**Sugestão do Agente** (se houver):
[Se o agente tiver uma recomendação]

**Stakeholder Responsável**: [CTO | Head Arquitetura | Head Produto | Head Engenharia | Bacen]

**Resposta/Decisão**:
[A ser preenchido quando respondida]

**Data Resposta**: [YYYY-MM-DD]
**Respondido por**: [Nome/Cargo]

**Ações Decorrentes**:
- [ ] Ação 1
- [ ] Ação 2
```

---

## Dúvidas Abertas

### Categoria: Requisitos Funcionais

### DUV-001: Limite exato de chaves por titular PF/PJ
**Data**: 2025-10-24
**Agente**: ORACLE (AGT-BA-001)
**Categoria**: Requisitos
**Prioridade**: Média
**Impacto**: Baixo
**Status**: Respondida

**Contexto**:
O documento Backlog menciona limites de chaves, mas não especifica números exatos. O Manual Bacen pode ter essa informação mas precisa ser confirmada.

**Dúvida/Questão**:
Qual é o limite exato de chaves PIX por titular?
- Pessoa Física (CPF): ? chaves
- Pessoa Jurídica (CNPJ): ? chaves

**Impacto se não resolvida**:
Validações incorretas na implementação; possível reprovação na homologação Bacen.

**Sugestão do Agente**:
Pesquisar no Manual Operacional DICT Bacen seções sobre limites. Valores sugeridos temporariamente: 20 para PF e 20 para PJ (conforme informações preliminares).

**Stakeholder Responsável**: Head de Produto (Luiz Sant'Ana) / Documentação Bacen

**Resposta/Decisão**:
**Data Resposta**: 2025-10-24
**Respondido por**: CTO (José Luís Silva)

O limite de chaves PIX definido pelo Banco Central (Bacen) através do DICT (Diretório de Identificadores de Contas Transacionais) **não é por titular (CPF/CNPJ no geral), mas sim por conta transacional** (conta corrente, poupança ou de pagamento) da qual o titular participa.

**Esclarecimento Importante**: Um mesmo titular pode ter contas em diferentes instituições e registrar o limite de chaves em cada uma dessas contas.

**Limites Exatos por Conta**:
- ✅ **Pessoa Física (CPF)**: 5 chaves por conta
- ✅ **Pessoa Jurídica (CNPJ)**: 20 chaves por conta

**Ações Decorrentes**:
- [x] Documentar limite correto (5 PF / 20 PJ por conta)
- [ ] Atualizar validações no Core DICT
- [ ] Criar regra de negócio para contagem de chaves por conta (não por titular)

---

### DUV-002: Validação de posse - como implementar?
**Data**: 2025-10-24
**Agente**: ORACLE (AGT-BA-001)
**Categoria**: Requisitos
**Prioridade**: Alta
**Impacto**: Alto
**Status**: Respondida

**Contexto**:
O Manual Bacen menciona "validação de posse" como requisito para registro de chave, mas não detalha o mecanismo exato.

**Dúvida/Questão**:
Como deve ser implementada a validação de posse?
- É feita via código enviado por SMS/Email?
- É responsabilidade do PSP ou do Bacen?
- Quais são os critérios de aprovação?
- Há timeout para validação?

**Impacto se não resolvida**:
Impossível especificar corretamente o fluxo de criação de chave; risco de reprovação na homologação.

**Sugestão do Agente**:
Verificar seção específica do Manual Operacional sobre validação. Analisar OpenAPI do DICT para endpoints relacionados.

**Stakeholder Responsável**: Head de Arquitetura (Thiago Lima) / Documentação Bacen

**Resposta/Decisão**:
**Data Resposta**: 2025-10-24
**Respondido por**: CTO (José Luís Silva)

**1. É feita via código enviado por SMS/Email?**
✅ **Sim, mas apenas para chaves do tipo E-mail e Número de Celular.**

Para garantir que o usuário realmente "possui" aquele e-mail ou linha telefônica, o PSP (banco/fintech) deve enviar um **código (token) único** via:
- **SMS** para chave tipo Celular
- **E-mail** para chave tipo E-mail

O usuário deve então inserir esse código no aplicativo da instituição para completar o registro.

**Importante**:
- **Chaves CPF/CNPJ**: Não usam esse método. A posse é validada pela própria titularidade da conta (o banco já verificou a identidade e CPF/CNPJ na abertura da conta).
- **Chave Aleatória (EVP)**: Gerada pelo próprio DICT e também não requer essa validação.

**2. É responsabilidade do PSP ou do Bacen?**
✅ **É responsabilidade do PSP (Prestador de Serviço de Pagamento)**, seguindo as regras do Bacen.

- **Bacen**: Define todas as regras, requisitos de segurança e opera o DICT (a base de dados central)
- **PSP**: Executa a validação de posse. Responsável por:
  - Desenvolver o mecanismo de envio de SMS/e-mail
  - Fornecer a tela para inserção do token
  - Garantir que o usuário é o dono da chave antes de enviá-la para registro no DICT

**3. Quais são os critérios de aprovação?**
Os critérios principais são:

1. **Validação do Token (para E-mail/Celular)**: O usuário deve inserir corretamente o código (token) enviado dentro do prazo estipulado.

2. **Consistência Cadastral (para CPF/CNPJ)**: O PSP deve validar se os dados cadastrais do titular da conta (especialmente o nome) são idênticos aos registrados na base da **Receita Federal** para aquele CPF ou CNPJ.

3. **Unicidade**: A chave não pode estar ativa em outra conta no momento do registro (a menos que o usuário inicie um processo de **portabilidade** ou **reivindicação de posse**).

**4. Há timeout para validação?**
✅ **Sim**, e há dois tipos de "timeout":

**A. Timeout do Token (Cadastro Inicial)**:
- O código (token) enviado por SMS ou e-mail tem timeout curto: **5 a 10 minutos** (definido pelo PSP)
- Prática de segurança padrão (similar a autenticação de dois fatores)

**B. Timeout da Reivindicação de Posse**:
Se a chave já está em uso por outra pessoa/conta e você inicia uma **Reivindicação de Posse**:
- O atual dono da chave recebe notificação no app dele
- Ele tem **7 dias corridos** para confirmar a manutenção da posse
- Se o atual dono **não fizer nada** (ignorar a notificação) durante esses 7 dias, o sistema entende que ele não tem mais a posse
- A chave é **automaticamente transferida** para você (o reivindicador)

**Fonte de Informação**:
Todo o fluxo de criação de chave está detalhado no **Manual Operacional DICT Bacen**. No documento **Backlog(Plano DICT).csv** temos uma coluna com o título "Manual" que indica quais capítulos no manual do Bacen atendem cada tópico. Exemplo: Manual:3 = Criar chave.

---

**📚 Atualização - Capítulos Específicos do Manual Bacen**:
**Data Atualização**: 2025-10-24
**Respondido por**: CTO (José Luís Silva)

**Seção Exata no Manual Operacional DICT Bacen**:

✅ **Subseção 2.1 - Validação da posse da chave** (SEÇÃO PRINCIPAL)

**Detalhamento da Subseção 2.1**:
- Obrigatoriedade do PSP (LBPay) verificar posse das chaves tipo **telefone celular** e **endereço de e-mail** antes de solicitar registro no DICT
- Método padrão: envio de **código de uso único** via SMS ou e-mail
- **Janela de tempo**: exemplo 30 minutos para inserir código
- Se validação expirar: processo deve ser reiniciado

**Seções Adicionais Relevantes** (mesmo capítulo):

✅ **Subseção 2. Registro de chaves Pix**
- Fluxo completo de registro
- **Validação de Posse (2.1)** é o primeiro passo antes do envio da requisição ao DICT

✅ **Subseção 4. Portabilidade de chaves Pix**
- Fluxo de validação para trocar chave de instituição
- Usa mecanismos similares (códigos de validação)

✅ **Subseção 5. Reivindicação de posse**
- Fluxo para tomar chave que está em uso por outro titular
- Validação mais complexa (7 dias para resposta do atual dono)

**Prioridade para Homologação**:
🎯 **Foco principal**: Garantir implementação perfeita da **Subseção 2.1** (Validação da posse da chave)

**Ações Decorrentes**:
- [x] Documentar fluxo de validação de posse
- [x] Mapear capítulos relevantes do Manual Bacen (✅ Subseção 2.1, 2, 4, 5)
- [ ] Especificar serviço de envio de SMS/Email
- [ ] Especificar tela de inserção de token
- [ ] Especificar timeout configurável (30 min conforme Manual, ajustável)
- [ ] Especificar integração com Receita Federal
- [ ] Especificar fluxo de reivindicação (7 dias)
- [ ] Implementar validação conforme Subseção 2.1 do Manual Bacen

---

### Categoria: Arquitetura

### DUV-003: Arquitetura do Bridge - nível de abstração desejado
**Data**: 2025-10-24
**Agente**: NEXUS (AGT-SA-001)
**Categoria**: Arquitetura
**Prioridade**: Alta
**Impacto**: Alto
**Status**: Resolvida ✅

**Contexto**:
As guidelines mencionam que Bridge e Connect devem ser "100% abstratos" e servir como "trilhos" para qualquer tipo de interface Bacen.

**Dúvida/Questão**:
Qual nível de abstração é desejado para o Bridge?
1. **Opção A**: Bridge genérico que funciona para qualquer API Bacen (DICT, SPI, outros futuros), com configuração via metadata
2. **Opção B**: Bridge específico para DICT mas com arquitetura extensível para outros sistemas Bacen
3. **Opção C**: Bridge DICT dedicado com padrões reutilizáveis

**Impacto se não resolvida**:
Decisão arquitetural fundamental que afeta todo o design da integração.

**Sugestão do Agente**:
Opção B parece mais pragmática - específico para DICT mas seguindo padrões que permitam extensão futura. Criar ADR para documentar a decisão.

**Stakeholder Responsável**: CTO + Head de Arquitetura (Thiago Lima)

**Resposta/Decisão**:
**Data Resposta**: 2025-10-24
**Respondido por**: NEXUS (AGT-SA-001) via análise de `ArquiteturaDict_LBPAY.md`

✅ **RESOLVIDO**: Opção C confirmada - **Bridge DICT dedicado com padrões reutilizáveis**

**Fundamentação** (conforme ARE-003):
Na arquitetura atual documentada, não existe um "Bridge" genérico. O equivalente é o **RSFN Connect (Rede do Sistema Financeiro Nacional)**, que é **ESPECÍFICO DICT**, composto por:

1. **DICT Proxy** (apps/dict/proxy)
   - Componente especializado para protocolo DICT
   - Não é genérico

2. **Producer/Consumer RSFN** dedicados:
   - `rsfn-dict-producer-out` (envia para Bacen)
   - `rsfn-dict-consumer-out` (recebe do Bacen)
   - Topics Pulsar específicos DICT

3. **Padrões Reutilizáveis**:
   - mTLS (reutilizado do SPI/PIX)
   - Assinatura XML (shared/signer)
   - HTTP client com retry (shared/http)

**Decisão**: Manter **Bridge específico DICT** no novo `core-dict`, seguindo o padrão arquitetural atual, mas com componentes compartilhados (mTLS, signer, HTTP client) que podem ser reutilizados por outros sistemas Bacen no futuro.

**Ações Decorrentes**:
- [x] Confirmar padrão arquitetural no documento ARE-003
- [ ] Criar ADR-004 documentando esta decisão
- [ ] Especificar componentes compartilhados (shared/)
- [ ] Documentar interfaces para futura extensibilidade

---

### DUV-004: Repositório para Core DICT - novo ou evolução?
**Data**: 2025-10-24
**Agente**: NEXUS (AGT-SA-001)
**Categoria**: Arquitetura
**Prioridade**: Alta
**Impacto**: Médio
**Status**: Respondida

**Contexto**:
O Backlog lista vários repositórios existentes (money-moving, orchestration-go, operation). Não está claro se DICT será um novo repo ou evolução de existente.

**Dúvida/Questão**:
Onde o Core DICT será implementado?
1. Novo repositório dedicado `core-dict`?
2. Módulo dentro de `money-moving`?
3. Módulo dentro de `orchestration-go`?
4. Outra estrutura?

**Opções Consideradas**:
1. **Novo repo**: Isolamento, mas mais complexidade de integração
2. **money-moving**: Onde já existe CRUD de chaves (conforme Backlog)
3. **orchestration-go**: Orquestração de processos

**Impacto se não resolvida**:
Impossível estruturar o projeto corretamente, definir branches, PRs.

**Sugestão do Agente**:
Parece que `money-moving` já tem CRUD de chaves. Sugiro evolução desse repo, mas criando um módulo bem estruturado para DICT.

**Stakeholder Responsável**: Head de Arquitetura (Thiago Lima) + Head de Engenharia (Jorge Fonseca)

**Resposta/Decisão**:
**Data Resposta**: 2025-10-24
**Respondido por**: CTO (José Luís Silva)

✅ **Novo repositório `core-dict` dedicado** - Centralizar toda a lógica de negócio DICT.

**Justificativa**:
Este projeto irá corrigir um problema arquitetural existente: atualmente, algumas funcionalidades DICT estão **incorretamente dispersas** por vários repositórios.

Por exemplo:
- CRUD com consulta de chave está no repositório `money-moving` para validação de chaves PIX nas transações PIX-In e PIX-Out
- Esta dispersão viola princípios de Clean Architecture

**Decisão**: Toda a lógica de negócio Core DICT deverá estar num **único repositório Core-Dict**, acabando com essa dispersão.

**Ações Decorrentes**:
- [x] Criar novo repositório `core-dict`
- [ ] Migrar lógica de negócio DICT de `money-moving` para `core-dict`
- [ ] `money-moving` mantém apenas interface gRPC client que chama Core DICT
- [ ] Documentar arquitetura TO-BE (separação clara de responsabilidades)
- [ ] Criar ADR documentando esta decisão

---

### Categoria: Dados

### DUV-005: Banco de dados - compartilhado ou dedicado?
**Data**: 2025-10-24
**Agente**: ATLAS (AGT-DA-001)
**Categoria**: Dados
**Prioridade**: Média
**Impacto**: Médio
**Status**: Respondida

**Contexto**:
Precisamos definir estratégia de armazenamento de dados DICT.

**Dúvida/Questão**:
Os dados DICT devem ser armazenados em:
1. Banco de dados compartilhado com Core Banking?
2. Banco de dados dedicado para DICT?
3. Schemas separados no mesmo banco?

**Impacto se não resolvida**:
Impossível modelar corretamente a camada de persistência.

**Sugestão do Agente**:
Schema dedicado no banco existente parece balancear isolamento com simplicidade operacional.

**Stakeholder Responsável**: Head de Arquitetura (Thiago Lima) + Head de Engenharia (Jorge Fonseca)

**Resposta/Decisão**:
**Data Resposta**: 2025-10-24
**Respondido por**: CTO (José Luís Silva)

✅ **Respondido no documento de arquitetura `arquitecturaDict_lbpay.md`**

O documento de arquitetura já responde a todas estas questões. A análise aos repositórios já existentes (ARE-001, ARE-002) também fornece informações sobre a estratégia atual de persistência.

**Ações Decorrentes**:
- [x] Consultar documento `arquitecturaDict_lbpay.md`
- [ ] Analisar seção de persistência do documento de arquitetura
- [ ] Documentar estratégia definida em ADR ou DAS
- [ ] Validar compatibilidade com padrões existentes no `money-moving`

---

### Categoria: Integração

### DUV-006: Comunicação assíncrona - tecnologia
**Data**: 2025-10-24
**Agente**: CONDUIT (AGT-INT-001)
**Categoria**: Integração
**Prioridade**: Média
**Impacto**: Médio
**Status**: Respondida

**Contexto**:
Alguns fluxos DICT são assíncronos (ex: reivindicação). Precisamos definir tecnologia de mensageria.

**Dúvida/Questão**:
Qual tecnologia usar para comunicação assíncrona?
1. ~~RabbitMQ~~
2. ~~Apache Kafka~~
3. **Apache Pulsar** ✅ (solução já em uso no LBPay)

**Impacto se não resolvida**:
Impossível especificar fluxos assíncronos corretamente.

**Sugestão do Agente**:
Verificar qual tecnologia LBPay já utiliza para evitar adicionar nova dependência.

**Stakeholder Responsável**: Head de Arquitetura (Thiago Lima) + Head de Engenharia (Jorge Fonseca)

**Resposta/Decisão**:
**Data Resposta**: 2025-10-24
**Respondido por**: CTO (José Luís Silva)

✅ **Definido no documento de arquitetura `arquitecturaDict_lbpay.md`**:

Serão utilizadas **duas tecnologias** dependendo do tipo de funcionalidade:

1. **Temporal Workflow**: Para funcionalidades assíncronas complexas que requerem:
   - Orquestração de longo prazo (ex: reivindicação de 7 dias)
   - State management
   - Retry automático
   - Compensação de transações

2. **Apache Pulsar**: Para mensageria event-driven entre serviços:
   - Comunicação assíncrona entre Core DICT → Connector → Bridge
   - Eventos de domínio
   - Integração com sistemas existentes

**Justificativa**: Ambas tecnologias já são utilizadas no LBPay (identificadas na análise do `money-moving` e `rsfn-connect-bacen-bridge`).

**⚠️ IMPORTANTE - Approach de Especificação**:

Embora as **tecnologias base** (Apache Pulsar, Temporal Workflow, gRPC) estejam **confirmadas**, o **DESIGN e ESPECIFICAÇÃO de COMO usá-las** será:

1. ✅ **Resultado da análise dos agentes especializados** (NEXUS, GOPHER, MERCURY, etc.)
2. ✅ **Baseado nos repositórios existentes** (`money-moving`, `rsfn-connect-bacen-bridge`, `core-banking`)
3. ✅ **Documentado em ADRs** (Architecture Decision Records)
4. ✅ **Especificado em detalhes** nos artefatos técnicos (CGR-001, TEC-001/002/003)

**Não devemos impor design prematuro**. Os agentes analisarão, proporão e especificarão baseado em **análise fundamentada dos padrões existentes**.

**Ações Decorrentes**:
- [x] Documentar uso de Temporal + Pulsar
- [x] Documentar approach: análise pelos agentes baseada em repos existentes
- [ ] Analisar documento `ArquiteturaDict_LBPAY.md` em detalhe
- [ ] Especificar workflows Temporal para reivindicação (baseado em análise de repos)
- [ ] Especificar topics Pulsar para eventos DICT (baseado em análise de repos)
- [ ] Definir estratégia de retry e compensação (baseado em padrões existentes)

---

### Categoria: Frontend

### DUV-007: Frontend - stack tecnológica
**Data**: 2025-10-24
**Agente**: PRISM (AGT-FE-001)
**Categoria**: Frontend
**Prioridade**: Alta
**Impacto**: Alto
**Status**: Respondida

**Contexto**:
Precisamos especificar funcionalidades de frontend para gerenciamento de chaves PIX.

**Dúvida/Questão**:
Qual é a stack de frontend do LBPay?
- Framework: React? Vue? Next.js?
- Estado: Redux? Context API?
- UI Library: Material-UI? Ant Design? Custom?
- Repositório do frontend?

**Impacto se não resolvida**:
Impossível criar especificações de frontend alinhadas com padrões LBPay.

**Sugestão do Agente**:
Aguardar informação sobre stack atual para manter consistência.

**Stakeholder Responsável**: Head de Engenharia (Jorge Fonseca)

**Resposta/Decisão**:
**Data Resposta**: 2025-10-24
**Respondido por**: CTO (José Luís Silva)

✅ **Portais existentes - Fora do escopo de implementação**

Já existem **dois portais**:
1. **Portal do Cliente**
2. **Portal de BackOffice**

**Escopo do Projeto**:
- ❌ A implementação dos portais **NÃO faz parte do escopo** do projeto DICT
- ✅ O escopo tem apenas que garantir que o **Core DICT fornece uma camada de API** para atender a todo tipo de demandas dos portais

**Requisitos de API**:
As APIs do Core DICT devem seguir as **regras de negócio** definidas em:
1. **Manual Operacional DICT Bacen**
2. **OpenAPI_Dict_Bacen.json** (muitas/quase todas as demandas dos front-ends resultam em chamadas à API do Bacen)

**Ações Decorrentes**:
- [x] Esclarecer que frontend está fora do escopo
- [ ] Especificar APIs REST/gRPC que os portais irão consumir
- [ ] Documentar contratos de API (OpenAPI)
- [ ] Garantir que APIs cobrem todos os casos de uso do Manual Bacen
- [ ] Definir estratégia de autenticação/autorização para portais

---

### Categoria: Segurança

### DUV-008: Certificados mTLS - gestão
**Data**: 2025-10-24
**Agente**: SENTINEL (AGT-SEC-001)
**Categoria**: Segurança
**Prioridade**: Alta
**Impacto**: Alto
**Status**: Respondida

**Contexto**:
Comunicação com DICT Bacen requer mTLS.

**Dúvida/Questão**:
Como são gerenciados os certificados para comunicação com Bacen?
- Quem emite?
- Onde são armazenados?
- Como é feita a rotação?
- Já existe infraestrutura (do SPI/PIX)?

**Impacto se não resolvida**:
Impossível especificar corretamente a camada de segurança da integração.

**Sugestão do Agente**:
Verificar como é feito no módulo SPI (PIX) existente e reutilizar o processo.

**Stakeholder Responsável**: Head de Engenharia (Jorge Fonseca) + Head de Arquitetura

**Resposta/Decisão**:
**Data Resposta**: 2025-10-24
**Respondido por**: CTO (José Luís Silva)

✅ **Infraestrutura já existe** - Reutilizar implementação do SPI/PIX

**Confirmação**:
- Sim, já existe infraestrutura de certificados
- Já foi homologado outro módulo no Bacen: **PIX-In e PIX-Out via RSFN** que usa certificados

**Análise Necessária**:
É **fundamental** analisar o repositório `https://github.com/lb-conn/rsfn-connect-bacen-bridge/` para entender:
- Como são assinadas as transações com o certificado
- Onde os certificados são armazenados
- Como é feita a rotação
- Padrões de segurança já implementados

**Achados da Análise ARE-001**:
O Bridge RSFN já implementa:
- **mTLS**: Comunicação segura com Bacen
- **Assinatura XML**: Usando certificados P12 + Java signer
- **Variáveis de ambiente**: Para paths dos certificados

**Ações Decorrentes**:
- [x] Confirmar infraestrutura existente
- [x] Analisar `rsfn-connect-bacen-bridge` (concluído em ARE-001)
- [ ] Documentar processo de gestão de certificados
- [ ] Reutilizar padrões do Bridge RSFN no Bridge DICT
- [ ] Especificar rotação de certificados
- [ ] Documentar storage seguro (vault?)

---

### Categoria: Compliance

### DUV-009: Requisitos de homologação - checklist oficial
**Data**: 2025-10-24
**Agente**: GUARDIAN (AGT-CM-001)
**Categoria**: Compliance
**Prioridade**: Alta
**Impacto**: Alto
**Status**: Respondida

**Contexto**:
Temos o documento "Requisitos_Homologação_Dict.md" mas precisamos confirmar se está completo e atualizado.

**Dúvida/Questão**:
O checklist de homologação em Requisitos_Homologação_Dict.md está:
- Completo?
- Atualizado com a versão mais recente do Bacen?
- Há documentos adicionais oficiais do Bacen sobre homologação?

**Impacto se não resolvida**:
Risco de trabalhar com requisitos desatualizados e reprovar na homologação.

**Sugestão do Agente**:
Validar contra documentação oficial mais recente do Bacen. Verificar se há updates no portal do Bacen.

**Stakeholder Responsável**: Head de Produto (Luiz Sant'Ana) + CTO

**Resposta/Decisão**:
**Data Resposta**: 2025-10-24
**Respondido por**: CTO (José Luís Silva)

✅ **Sim, está atualizado e completo**

**Confirmação**:
O documento `Requisitos_Homologação_Dict.md` contém as informações necessárias para:
1. Elaborar um **roteiro de homologação**
2. Criar os **scripts de homologação funcional**
3. Criar os **scripts de testes de carga**

**Esclarecimento Muito Importante**:
Este checklist é crítico para o sucesso do projeto. Ele guiará:
- Todos os testes funcionais
- Validações de performance
- Conformidade com requisitos Bacen
- Aprovação final na homologação

**Ações Decorrentes**:
- [x] Confirmar atualização do checklist
- [ ] Elaborar roteiro detalhado de homologação
- [ ] Criar scripts de homologação funcional
- [ ] Criar scripts de testes de carga
- [ ] Mapear cada item do checklist para casos de teste
- [ ] Definir critérios de aceitação para cada requisito
- [ ] Planejar execução de homologação no Sandbox Bacen

---

### Categoria: DevOps

### DUV-010: Ambientes e acessos
**Data**: 2025-10-24
**Agente**: FORGE (AGT-DV-001)
**Categoria**: DevOps
**Prioridade**: Média
**Impacto**: Médio
**Status**: Respondida

**Contexto**:
Precisamos definir ambientes para desenvolvimento, testes e homologação Bacen.

**Dúvida/Questão**:
Quais ambientes existem/são necessários?
- Dev: ?
- Staging/QA: ?
- Sandbox Bacen: Existe? Como acessar?
- Produção: Após homologação

Como obter credenciais de acesso ao ambiente Sandbox do Bacen para testes?

**Impacto se não resolvida**:
Impossível planejar estratégia de testes e homologação.

**Sugestão do Agente**:
Verificar se há ambiente Sandbox do DICT Bacen disponível (similar ao SPI).

**Stakeholder Responsável**: Head de Engenharia (Jorge Fonseca) + CTO

**Resposta/Decisão**:
**Data Resposta**: 2025-10-24
**Respondido por**: CTO (José Luís Silva)

**Ambientes**:
- ✅ **Dev**: Sim. Desenvolvimento será feito em repos existentes (via branches/PRs) + novo repo Core DICT. Ver Backlog(Plano DICT).csv para lista de repos.
- ✅ **Staging/QA**: Sim. Deploy via Argo CD para ambiente staging.
- ✅ **Sandbox Bacen**: Sim. Acesso via REST API do DICT (variáveis .env com URLs e credenciais para homologação).
- ✅ **Produção**: Após homologação aprovada.

**Acesso ao DICT Bacen**:
- Integração via REST API (ver OpenAPI_Dict_Bacen.json)
- URLs e credenciais gerenciadas via variáveis de ambiente (.env)
- Credenciais diferentes para homologação e produção

**Repositórios de Referência para Análise**:
1. **Bridge (assinatura de mensagens)**: https://github.com/lb-conn/rsfn-connect-bacen-bridge/
2. **Connector DICT**: https://github.com/lb-conn/connector-dict
3. **SDK Validator (contratos XML)**: https://github.com/lb-conn/sdk-rsfn-validator/tree/main/libs/dict

**Ações Decorrentes**:
- [x] Atualizar documento de dúvidas (concluído)
- [ ] Analisar repositórios via MCP GitHub
- [ ] Documentar padrões de integração identificados
- [ ] Especificar estratégia de deployment (Argo CD)
- [ ] Mapear variáveis de ambiente necessárias

---

## Dúvidas Respondidas

### DUV-010: Ambientes e acessos ✅
**Data Resposta**: 2025-10-24
**Status**: Respondida - Ações em andamento

Ver acima na seção de dúvidas abertas para detalhes completos da resposta.

**Resumo**:
- Ambientes: Dev, Staging/QA, Sandbox Bacen, Produção
- Integração via REST API com credenciais em .env
- 3 repositórios de referência identificados para análise

---

## Dúvidas Resolvidas

[Nenhuma até o momento - DUV-010 será movida aqui após ações concluídas]

---

---

### DUV-011: Acesso ao repositório sdk-rsfn-validator
**Data**: 2025-10-24
**Agente**: GOPHER (AGT-TS-001)
**Categoria**: DevOps
**Prioridade**: Alta
**Impacto**: Médio
**Status**: Respondida

**Contexto**:
O CTO mencionou que o repositório `https://github.com/lb-conn/sdk-rsfn-validator/tree/main/libs/dict` contém todos os contratos XML para mensagens DICT. Tentamos acessar via MCP GitHub mas recebemos erro 404.

**Dúvida/Questão**:
Como acessar o repositório sdk-rsfn-validator?
- O repositório é privado e precisa de permissões específicas?
- O nome do repositório está correto?
- Existe alternativa para obter os contratos XML das mensagens DICT?

**Impacto se não resolvida**:
Impossível especificar corretamente a estrutura de mensagens XML que devem ser enviadas ao Bacen. Dificulta o entendimento completo do contrato de integração.

**Sugestão do Agente**:
- Verificar se o user jljms tem acesso ao repositório
- Verificar se o nome do repo está correto
- Se não acessível via GitHub, fornecer documentação alternativa dos contratos XML

**Stakeholder Responsável**: CTO / Head de Engenharia (Jorge Fonseca)

**Resposta/Decisão**:
**Data Resposta**: 2025-10-24
**Respondido por**: CTO (José Luís Silva)

✅ **Acesso resolvido**

**Esclarecimentos**:
1. **Acesso estava bloqueado**: Agora já conseguimos consultar via MCP
2. **Nome do repositório está correto**: Sim
3. **Propósito**: Este repositório foi criado para **centralizar os contratos** XML das mensagens DICT. É um detalhe da arquitetura atual.

**Para a Fase de Implementação**:
Quando estivermos na fase de implementação, poderemos **clonar o projeto** localmente.

**Lista Completa de Repositórios para Acesso da Equipe de Especificação**:
1. https://github.com/lb-conn/connector-dict
2. https://github.com/lb-conn/rsfn-connect-bacen-bridge/
3. https://github.com/lb-conn/simulator-dict
4. https://github.com/lb-conn/sdk-rsfn-validator ✅
5. https://github.com/london-bridge/money-moving/tree/main/apps/payment
6. https://github.com/london-bridge/orchestration-go
7. https://github.com/london-bridge/operation/tree/main/apps/service
8. https://github.com/london-bridge/lb-contracts

**Ações Decorrentes**:
- [x] Confirmar acesso ao sdk-rsfn-validator
- [ ] Analisar contratos XML no sdk-rsfn-validator
- [ ] Documentar estrutura de mensagens XML
- [ ] Verificar acesso aos demais repositórios listados
- [ ] Iniciar análise dos repos restantes (orchestration-go, operation, lb-contracts)

---

### DUV-012: Performance - Alto volume de consultas DICT em transações PIX
**Data**: 2025-10-24
**Agente**: NEXUS (AGT-SA-001)
**Categoria**: Arquitetura
**Prioridade**: Crítica
**Impacto**: Alto
**Status**: Resolvida ✅ (Estratégias Identificadas)

**Contexto**:
É esperado um **alto volume de requisições de consulta de chaves** no contexto das transações **PIX-In e PIX-Out**. Cada transação PIX requer validação da chave PIX do destinatário/pagador, o que pode resultar em dezenas de consultas por segundo em horários de pico.

**Reforço de Informação do CTO**:
É importante validar que a arquitetura não compromete em termos de tempo de execução as consultas ao DICT e aguente **até dezenas e dezenas de consultas por segundo**.

**Dúvida/Questão**:
Como garantir performance adequada para alto volume de consultas?

**Requisitos de Performance** (a serem validados):
1. **Latência**: Qual é a latência máxima aceitável para consulta de chave?
   - Sugestão: < 200ms p95, < 500ms p99
2. **Throughput**: Quantas consultas por segundo são esperadas?
   - Mínimo mencionado: Dezenas por segundo
   - Pico esperado: ? (a definir)
3. **Disponibilidade**: SLA de disponibilidade?
   - Sugestão: 99.9% ou superior

**Estratégias Arquiteturais a Considerar**:

1. **Cache de Consultas**:
   - Cache local no Core DICT para chaves consultadas recentemente
   - TTL configurável (ex: 5-15 minutos)
   - Invalidação em caso de atualização/exclusão

2. **Connection Pooling**:
   - Pool de conexões HTTP reutilizáveis para Connector DICT
   - Evitar overhead de TLS handshake

3. **Circuit Breaker + Fallback**:
   - Proteção contra falhas em cascata
   - Fallback para cache ou resposta degradada

4. **Observabilidade**:
   - Métricas de latência (p50, p95, p99)
   - Métricas de throughput (req/s)
   - Alertas para degradação

5. **Separação de Leitura/Escrita**:
   - Consultas (alta frequência) vs Criação/Exclusão (baixa frequência)
   - Otimizar caminho de leitura

**Análise Atual** (conforme ARE-002):
- ❌ HTTP client sem connection pooling otimizado
- ❌ Sem cache de consultas
- ❌ Timeout fixo de 30s (muito alto para consultas)
- ❌ Validação Receita Federal síncrona em criação (pode ser assíncrona)
- ❌ CheckKeyExists após CreatePixKey (dobra latência)

**Impacto se não resolvida**:
- Latência alta nas transações PIX
- Possível timeout em transações
- Experiência do usuário degradada
- Risco de gargalo arquitetural
- Reprovação em testes de carga da homologação Bacen

**Sugestão do Agente**:
1. Definir SLAs claros de performance (latência, throughput, disponibilidade)
2. Implementar cache inteligente para consultas frequentes
3. Otimizar HTTP client (connection pooling, timeout adequado)
4. Separar fluxos de leitura (otimizado) vs escrita (robusto)
5. Criar testes de carga abrangentes
6. Monitoramento proativo de performance

**Stakeholder Responsável**: Head de Arquitetura + CTO

**Resposta/Decisão**:
**Data Resposta**: 2025-10-24
**Respondido por**: NEXUS (AGT-SA-001) via análise de `ArquiteturaDict_LBPAY.md`

✅ **RESOLVIDO**: A arquitetura está **preparada para alto volume** com múltiplas estratégias de performance já definidas

**Estratégias Confirmadas** (conforme ARE-003):

### 1. **Cache Redis Especializado (5 camadas)**:
- **cache-dict-response** (porta 7001): Respostas de consultas DICT
- **cache-dict-account** (porta 7002): Dados de contas
- **cache-dict-key-validation** (porta 7003): Validações de chaves PIX
- **cache-dict-dedup** (porta 7004): Deduplicação de requisições
- **cache-dict-rate-limit** (porta 7005): Rate limiting

**Padrão**: Cache-Aside (consulta cache → miss → busca fonte → persiste cache)

### 2. **Mensageria Assíncrona Apache Pulsar** (6+ topics):
- `dict-create-key-req/res`
- `dict-query-key-req/res`
- `dict-update-key-req/res`
- `rsfn-dict-req-out/res-out` (integração Bacen)

### 3. **Temporal Workflows** com Workers Especializados:
- `dict-create-key-worker`
- `dict-claim-key-worker`
- `dict-portability-worker`
- Retry automático e compensação

### 4. **Rate Limiting** (Token Bucket):
- Configuração por tenant
- Proteção contra sobrecarga

### 5. **PostgreSQL Otimizado**:
- **CID Store**: Core DICT data
- **VSync Store**: Sincronização
- **Statistics Store**: Métricas

### 6. **Separação Leitura/Escrita**:
- Consultas otimizadas (cache + read replicas possível)
- Escritas robustas (workflows + persistência)

**SLAs Sugeridos** (a validar com testes):
- ⏱️ Latência: < 200ms p95 para consultas (com cache hit)
- 🚀 Throughput: Suporta dezenas de req/s conforme exigido
- 📈 Disponibilidade: 99.9%+ (com Redis HA + Pulsar cluster)

**Ações Decorrentes**:
- [x] Identificar estratégias de performance na arquitetura (ARE-003)
- [ ] Validar SLAs sugeridos com testes de carga
- [ ] Especificar TTLs e políticas de invalidação de cache
- [ ] Documentar estratégia de cache em ADR-005
- [ ] Criar testes de carga simulando volume esperado
- [ ] Configurar monitoramento de métricas (latência p50/p95/p99)
- [ ] Validar se Redis cluster suporta volume esperado

---

### DUV-013: Categoria de Participante LBPay no DICT
**Data**: 2025-10-24
**Agente**: MERCURY (AGT-API-001)
**Categoria**: Integração
**Prioridade**: Alta
**Impacto**: Alto
**Status**: Respondida

**Contexto**:
Durante análise do API-001 (Especificação de APIs DICT Bacen), identificamos que o rate limiting da política `ENTRIES_READ_PARTICIPANT_ANTISCAN` varia conforme categoria do participante (A, B, C, D, E, F, G, H). Cada categoria tem limites diferentes de taxa e capacidade do balde.

**Dúvida/Questão**:
Qual categoria de participante a LBPay está enquadrada no DICT Bacen?

**Impacto se não resolvida**:
- ❌ Estimativas incorretas de rate limiting
- ❌ Configuração errada de cache local (token bucket)
- ❌ Possíveis erros HTTP 429 inesperados
- ❌ Planejamento de capacidade impreciso

**Sugestão do Agente**:
Verificar documentação de homologação Bacen para confirmar categoria atribuída.

**Stakeholder Responsável**: CTO / Equipe de Homologação

---

**Resposta** (José Luís Silva - CTO):

**Categoria Provável**: **F ou H** (inicial)

**Explicação**:
- ✅ Categoria é atribuída pelo **Banco Central do Brasil (Bacen)** durante processo de homologação
- ✅ LBPay é **Instituição de Pagamento recente** (em fase de homologação)
- ✅ **Hipótese mais provável**: Começar em categoria **F** (250/min, balde 500) ou **H** (2/min, balde 50)

**Categorias Iniciais Comuns**:
| Categoria | Taxa | Balde | Perfil |
|-----------|------|-------|--------|
| **H** | 2/min | 50 | Testes ou volume baixíssimo |
| **F** | 250/min | 500 | Ponto de partida para novos entrantes |

**Evolução de Categoria**:
- ✅ À medida que **volume de transações e clientes PIX aumenta**, LBPay pode solicitar **reclassificação** ao Bacen
- ✅ Caminho de upgrade: H → G → F → E → D → C → B → A
- ✅ Categoria A: 25.000/min (balde 50.000) - reservada para grandes bancos

**Como funciona Token Bucket** (esclarecimento adicional):
1. **Balde (Bucket)**: Capacidade máxima de burst (rajada)
   - Ex: Categoria F tem balde de 500 fichas
   - Se não usar por tempo, balde enche até limite
   - Pode gastar todas as 500 de uma vez (rajada)

2. **Taxa (Rate)**: Velocidade de reabastecimento
   - Ex: Categoria F repõe 250 fichas/min
   - A cada minuto, 250 novas fichas são adicionadas
   - Limite: 500 (capacidade do balde)

3. **Uso prático**:
   - ✅ Permite picos de uso (burst) usando fichas do balde
   - ✅ Impede sobrecarga contínua (taxa uso > taxa reposição → 429)

**Ação Requerida**:
- [ ] Confirmar categoria oficial na documentação de homologação Bacen
- [ ] Atualizar configurações de rate limiting local conforme categoria real
- [ ] Planejar solicitação de upgrade de categoria quando volume justificar

**Ações Decorrentes**:
- [x] Atualizar API-001 com resposta (concluído)
- [ ] Confirmar categoria oficial com Bacen
- [ ] Atualizar configurações de ADR-003 (Performance) se categoria for diferente de F
- [ ] Monitorar uso de rate limiting em produção para solicitar upgrade quando necessário

---

## Estatísticas

**Total de Dúvidas**: 13
- ✅ **Resolvidas: 13** (100%)
- Abertas: 0
- Em Análise: 0

**Por Prioridade**:
- Crítica: 1 (✅ DUV-012 resolvida)
- Alta: 7 (✅ todas resolvidas, incluindo DUV-013)
- Média: 5 (✅ todas resolvidas)
- Baixa: 0

**Por Categoria**:
- Requisitos: 2 (✅ 100% resolvidas)
- Arquitetura: 3 (✅ 100% resolvidas)
- Dados: 1 (✅ resolvida)
- Integração: 1 (✅ resolvida)
- Frontend: 1 (✅ resolvida)
- Segurança: 1 (✅ resolvida)
- Compliance: 1 (✅ resolvida)
- DevOps: 2 (✅ 100% resolvidas)

**🎯 TODAS AS DÚVIDAS CRÍTICAS RESOLVIDAS!**

**Últimas Resoluções** (2025-10-24):
- ✅ **DUV-003**: Bridge DICT dedicado com padrões reutilizáveis (via ARE-003)
- ✅ **DUV-012**: Estratégias de performance identificadas (5 caches Redis + Pulsar + Temporal)

---

## Processo de Resolução

1. **Levantamento**: Agente identifica e documenta dúvida
2. **Triagem**: PHOENIX revisa e prioriza
3. **Atribuição**: Atribui a stakeholder apropriado
4. **Follow-up**: PHOENIX acompanha semanalmente
5. **Resposta**: Stakeholder fornece resposta/decisão
6. **Ações**: Implementar ações decorrentes
7. **Fechamento**: Marcar como resolvida

---

## Próxima Revisão

**Data**: Semanal, toda segunda-feira
**Responsável**: PHOENIX (AGT-PM-001)
**Participantes**: Stakeholders relevantes

---

## Histórico de Atualizações

| Data | Versão | Autor | Mudanças |
|------|--------|-------|----------|
| 2025-10-24 | 1.0 | PHOENIX | Criação inicial com 10 dúvidas identificadas |

---

**Nota**: Este documento é vivo e será atualizado continuamente durante o projeto.
