# CRF-001: Checklist de Requisitos Funcionais DICT

**Versão**: 1.0  
**Data**: 2025-10-24  
**Autor**: ORACLE (Business Analyst - AGT-BA-001)  
**Status**: Rascunho - Consolidação em Progresso  
**Fonte**: Manual Operacional DICT Bacen v8 + Backlog Plano DICT (73 funcionalidades)

---

## Resumo Executivo

### Números Consolidados
- **Total de Requisitos Funcionais**: 72 (mapeados do Backlog CSV)
- **Blocos Funcionais**: 6 blocos principais
- **Capítulos Manual**: 20 capítulos operacionais
- **Status Geral**: 13.9% implementado (apenas DICT e Bridge em desenvolvimento)

### Distribuição por Bloco
| Bloco | Descrição | QTD RFs | Prioridade | Status |
|-------|-----------|---------|-----------|--------|
| 1 | CRUD de Chaves | 13 | Must Have | Parcial |
| 2 | Reivindicação | 14 | Should Have | Não iniciado |
| 3 | Validação | 3 | Must Have | Não iniciado |
| 4 | Devolução/Infração | 6 | Should Have | Não iniciado |
| 5 | Segurança | 13 | Should Have | Não iniciado |
| 6 | Recuperação de Valores | 13 | Nice to Have | Não iniciado |
| **TOTAL** | | **62** | | |

**Outros**: 10 funcionalidades de suporte técnico (comunicação, repositórios, etc.)

---

## BLOCO 1: CRUD DE CHAVES (Cap 1-4, 7-8)

Requisitos essenciais para operações básicas de criar, ler, atualizar e deletar chaves PIX no DICT.

### RF-BLO1-001: Registrar chave por solicitação do usuário final (Acesso Direto)
- **Manual**: Cap 3.1 - Fluxo de Registro de Chave por Solicitação do Usuário Final (Acesso Direto)
  - **Subseção 2.1**: Validação da posse da chave (pré-requisito)
  - **Subseção 2**: Registro de chaves Pix (fluxo completo)
- **Backlog**: Linha 3
- **Descrição**: O usuário final solicita registro de chave PIX (CPF, CNPJ, email, telefone ou aleatória) através de seu PSP com acesso direto ao DICT. O PSP valida posse, verifica compatibilidade com Receita Federal e envia ao DICT. O DICT registra a chave se não houver conflito.
- **Prioridade**: Must Have
- **Complexidade**: Média
- **Tipo de Acesso**: Direto
- **Status**: Parcialmente Implementado
- **Critérios de Aceitação**:
  - **Validação de Posse (Subseção 2.1 do Manual Bacen)**:
    - ✅ **Chaves tipo telefone celular**: PSP envia código único via SMS
    - ✅ **Chaves tipo e-mail**: PSP envia código único via e-mail
    - ✅ **Timeout**: Usuário tem 30 minutos para inserir código (configurável)
    - ✅ **Expiração**: Se timeout expirar, processo deve ser reiniciado
    - ✅ **Chaves CPF/CNPJ**: Posse validada pela titularidade da conta (sem código)
    - ✅ **Chaves aleatórias (EVP)**: Gerada pelo DICT, não requer validação de posse
  - PSP deve confirmar dados e situação cadastral na Receita Federal
  - DICT deve verificar autorização do PSP para registrar
  - DICT deve verificar se chave já está registrada
  - Em caso de chave aleatória, DICT gera UUID (RFC4122)
  - Sistema retorna sucesso ou erro com dados de conflito
  - Casos de conflito: notificar usuário sobre portabilidade/reivindicação

### RF-BLO1-002: Registrar chave por solicitação do usuário final (Acesso Indireto)
- **Manual**: Cap 3.2 - Fluxo de Registro de Chave por Solicitação do Usuário Final (Acesso Indireto)
  - **Subseção 2.1**: Validação da posse da chave (pré-requisito)
  - **Subseção 2**: Registro de chaves Pix (fluxo completo)
- **Backlog**: Linha 3 (variante)
- **Descrição**: Mesmo fluxo que RF-BLO1-001, mas por intermediação de PSP com acesso indireto ao DICT. PSP indireto comunica com PSP direto que efetua operação no DICT.
- **Prioridade**: Must Have
- **Complexidade**: Alta
- **Tipo de Acesso**: Indireto
- **Status**: Não Iniciado
- **Critérios de Aceitação**:
  - **Validação de Posse**: Mesmos critérios do RF-BLO1-001 (Subseção 2.1)
  - Fluxo de comunicação entre dois PSPs funcional
  - PSP direto deve verificar autorização de delegação
  - PSP direto deve validar permissão de registrar em nome de terceiro
  - Retorno de erros e sucessos para ambos PSPs

### RF-BLO1-003: Registrar chave iniciado pelo participante (Acesso Direto)
- **Manual**: Cap 3.3 - Fluxo de Registro de Chave Iniciado pelo Participante (Acesso Direto)
  - **Subseção 2.1**: Validação da posse da chave (pré-requisito)
- **Backlog**: Linha 3 (variante operacional)
- **Descrição**: PSP com acesso direto detecta chave registrada internamente mas não no DICT (sincronismo). Envia registro direto ao DICT sem solicitação do usuário.
- **Prioridade**: Must Have
- **Complexidade**: Média
- **Status**: Não Iniciado
- **Critérios de Aceitação**:
  - **Validação de Posse**: Chave já deve ter sido validada anteriormente (Subseção 2.1)
  - PSP deve processar verificação de sincronismo
  - PSP envia mensagem "Diretório / Criar Vínculo"
  - DICT verifica conformidade (autorização e vinculação)
  - DICT registra chave ou retorna conflito existente

### RF-BLO1-004: Registrar chave iniciado pelo participante (Acesso Indireto)
- **Manual**: Cap 3.4 - Fluxo de Registro de Chave Iniciado pelo Participante (Acesso Indireto)
  - **Subseção 2.1**: Validação da posse da chave (pré-requisito)
- **Backlog**: Linha 3 (variante)
- **Descrição**: Mesmo que RF-BLO1-003 mas via intermediação de PSP indireto.
- **Prioridade**: Must Have
- **Complexidade**: Alta
- **Status**: Não Iniciado
- **Critérios de Aceitação**:
  - **Validação de Posse**: Mesmos critérios do RF-BLO1-003 (Subseção 2.1)

### RF-BLO1-005: Excluir chave por incompatibilidade com Receita Federal
- **Manual**: Cap 4.1 - Exclusão de Chave por Incompatibilidade de Dados com a Receita Federal
- **Backlog**: Linha 4
- **Descrição**: PSP identifica divergências em nome ou irregularidade em CPF/CNPJ (suspenso, cancelado, falecido, nulo). Exclui chave do DICT usando código "FRAUD" ou "RFB_VALIDATION" no campo Reason. Usuário é notificado imediatamente.
- **Prioridade**: Must Have
- **Complexidade**: Alta
- **Status**: Não Iniciado
- **Critérios de Aceitação**:
  - PSP detecta incompatibilidade via Receita Federal
  - PSP envia exclusão com campo Reason preenchido ("FRAUD" ou "RFB_VALIDATION")
  - DICT exclui chave de seu banco de dados
  - Usuário recebe notificação imediata com motivo detalhado
  - Classificar como fraude ou validação RFB

### RF-BLO1-006: Excluir chave por solicitação do usuário final (Acesso Direto)
- **Manual**: Cap 4.2 - Fluxo de Exclusão de Chave por Solicitação do Usuário Final (Acesso Direto)
- **Backlog**: Linha 5
- **Descrição**: Usuário final solicita exclusão de chave PIX através de seu PSP com acesso direto ao DICT. PSP valida autorização e envia ao DICT. DICT exclui e confirma operação.
- **Prioridade**: Must Have
- **Complexidade**: Média
- **Status**: Parcialmente Implementado
- **Critérios de Aceitação**:
  - Usuário acessa canal de atendimento e solicita exclusão
  - PSP valida se chave está registrada e se usuário é titular
  - PSP envia mensagem "Diretório / Remover Vínculo"
  - DICT verifica: (i) chave registrada, (ii) PSP autorizado, (iii) usuário titular
  - DICT exclui chave de banco de dados
  - PSP recebe confirmação e atualiza base interna
  - Confirmação enviada ao usuário final

### RF-BLO1-007: Excluir chave por solicitação do usuário final (Acesso Indireto)
- **Manual**: Cap 4.3 - Fluxo de Exclusão de Chave por Solicitação do Usuário Final (Acesso Indireto)
- **Backlog**: Linha 5 (variante)
- **Descrição**: Mesmo que RF-BLO1-006 mas via intermediação de PSP indireto.
- **Prioridade**: Must Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO1-008: Excluir chave iniciado pelo participante (Acesso Direto)
- **Manual**: Cap 4.4 - Fluxo de Exclusão de Chave Iniciado pelo Participante (Acesso Direto)
- **Backlog**: Linha 6
- **Descrição**: PSP com acesso direto identifica chave que deve ser excluída (encerramento, sincronismo, fraude) e inicia exclusão direto no DICT sem solicitação do usuário.
- **Prioridade**: Must Have
- **Complexidade**: Média
- **Status**: Não Iniciado
- **Critérios de Aceitação**:
  - PSP inicia processo de exclusão
  - PSP envia "Diretório / Remover Vínculo"
  - DICT verifica: (i) chave registrada, (ii) PSP autorizado
  - DICT exclui e envia confirmação
  - PSP atualiza base e notifica usuário

### RF-BLO1-009: Excluir chave iniciado pelo participante (Acesso Indireto)
- **Manual**: Cap 4.5 - Fluxo de Exclusão de Chave Iniciado pelo Participante (Acesso Indireto)
- **Backlog**: Linha 7 (variante)
- **Descrição**: Mesmo que RF-BLO1-008 mas via intermediação de PSP indireto.
- **Prioridade**: Must Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO1-010: Alterar dados vinculados à chave (Acesso Direto)
- **Manual**: Cap 7.1 - Fluxo de Alteração dos Dados Vinculados à Chave (Acesso Direto)
- **Backlog**: Linha 11
- **Descrição**: Usuário final solicita alteração dos dados da conta transacional vinculada à chave (titular, ISPB, agência, conta). PSP com acesso direto valida e envia ao DICT.
- **Prioridade**: Must Have
- **Complexidade**: Alta
- **Status**: Não Iniciado
- **Critérios de Aceitação**:
  - Usuário acessa canal e solicita alteração
  - PSP valida dados e situação cadastral na Receita Federal
  - PSP envia mensagem de alteração ao DICT
  - DICT valida conformidade (autorização PSP e vínculo com chave)
  - DICT atualiza dados vinculados à chave
  - PSP recebe confirmação e atualiza base interna

### RF-BLO1-011: Alterar dados vinculados à chave (Acesso Indireto)
- **Manual**: Cap 7.2 - Fluxo de Alteração dos Dados Vinculados à Chave (Acesso Indireto)
- **Backlog**: Linha 11 (variante)
- **Descrição**: Mesmo que RF-BLO1-010 mas via intermediação de PSP indireto.
- **Prioridade**: Must Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO1-012: Alterar dados para correção de inconsistências
- **Manual**: Cap 7.3 - Alteração dos Dados Vinculados à Chave para Correção de Inconsistências
- **Backlog**: Linha 12
- **Descrição**: PSP utiliza fluxo especial para corrigir dados inconsistentes ou divergências de nome detectadas após registro (sem solicitação do usuário final).
- **Prioridade**: Must Have
- **Complexidade**: Média
- **Status**: Não Iniciado
- **Critérios de Aceitação**:
  - PSP detecta inconsistência
  - PSP envia correção ao DICT com justificativa
  - DICT valida e aplica correção
  - Alteração é registrada em auditoria

### RF-BLO1-013: Consultar dados da chave para participantes PIX
- **Manual**: Cap 8.2 e 8.3 - Fluxo de Consulta de Chave
- **Backlog**: Linha 10
- **Descrição**: Participante do PIX consulta dados da chave registrada no DICT (para ambos acesso direto e indireto). DICT retorna dados permitidos (dados de identificação, mas não dados bancários completos).
- **Prioridade**: Must Have
- **Complexidade**: Média
- **Status**: Parcialmente Implementado
- **Critérios de Aceitação**:
  - Mensagem "Diretório / Consultar Vínculo"
  - DICT valida autorização do consultante
  - Retorna dados de chave permitidos (cap 8.1)
  - Restringir exposição de dados sensíveis
  - Chaves bloqueadas retornam erro "EntryBlocked"
  - Registrar log de consultas (para monitoramento de ataques)

---

## BLOCO 2: REIVINDICAÇÃO DE POSSE (Cap 5-6)

Requisitos para fluxos de reivindicação de posse (quando usuário acredita que chave é sua) e portabilidade de chave.

### RF-BLO2-001: Portabilidade de chave para PSP reivindicador (Acesso Direto)
- **Manual**: Cap 5.1 - Fluxo de Portabilidade para PSP Reivindicador com Acesso Direto
- **Backlog**: Linha 27
- **Descrição**: Usuário solicita portabilidade de chave que já possui em outro PSP (doador). PSP reivindicador com acesso direto envia pedido ao DICT. Pedido fica "Aberto" por 7 dias. PSP doador confirma ou cancela. Em caso de confirmação, chave é transferida.
- **Prioridade**: Should Have
- **Complexidade**: Muito Alta
- **Status**: Não Iniciado
- **Critérios de Aceitação**:
  - Pedido criado com status "Aberto"
  - Período de resolução de 7 dias iniciado
  - PSP doador recebe notificação
  - PSP doador pode confirmar ou cancelar
  - PSP reivindicador pode cancelar se status "Aberto"
  - PSP reivindicador consulta periodicamente (min 1x/min)
  - Ao confirmar: chave bloqueada até conclusão
  - Ao concluir: chave atualizada para reivindicador
  - Consultas retornam conta doadora até "Aguardando Resolução"
  - Status final: "Completo"

### RF-BLO2-002: Portabilidade de chave para PSP reivindicador (Acesso Indireto)
- **Manual**: Cap 5.2 - Fluxo de Portabilidade para PSP Reivindicador com Acesso Indireto
- **Backlog**: Linha 27 (variante)
- **Descrição**: Mesmo que RF-BLO2-001 mas PSP reivindicador tem acesso indireto ao DICT.
- **Prioridade**: Should Have
- **Complexidade**: Muito Alta
- **Status**: Não Iniciado

### RF-BLO2-003: Portabilidade de chave para PSP doador (Acesso Direto)
- **Manual**: Cap 5.3 - Fluxo de Portabilidade para PSP Doador com Acesso Direto
- **Backlog**: Linha 28
- **Descrição**: PSP doador recebe pedido de portabilidade do DICT. Deve resolver em 7 dias. Pode confirmar (aceita portabilidade) ou cancelar (nega). Após 7 dias, deve cancelar se ainda aberto.
- **Prioridade**: Should Have
- **Complexidade**: Alta
- **Status**: Não Iniciado
- **Critérios de Aceitação**:
  - PSP doador consulta periodicamente (min 1x/min)
  - PSP doador pode confirmar portabilidade
  - PSP doador pode cancelar portabilidade
  - Cancelamento automático após 7 dias se não resolvido
  - PSP doador pode cancelar status "Confirmado" se erro
  - Notificação de usuário doador
  - Permitir atualização de dados enquanto status "Aberto" ou "Aguardando"

### RF-BLO2-004: Portabilidade de chave para PSP doador (Acesso Indireto)
- **Manual**: Cap 5.4 - Fluxo de Portabilidade para PSP Doador com Acesso Indireto
- **Backlog**: Linha 28 (variante)
- **Descrição**: Mesmo que RF-BLO2-003 mas PSP doador tem acesso indireto ao DICT.
- **Prioridade**: Should Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO2-005: Reivindicar chave - PSP reivindicador cria reivindicação
- **Manual**: Cap 6.1 - Fluxo de Reivindicação de Posse para PSP Reivindicador (Acesso Direto)
- **Backlog**: Linha 16
- **Descrição**: Usuário acredita que chave está em outro PSP (doador) mas é sua. PSP reivindicador com acesso direto cria reivindicação de posse no DICT. Aguarda resposta do PSP doador por 7 dias. Se confirmado, chave é transferida.
- **Prioridade**: Should Have
- **Complexidade**: Muito Alta
- **Status**: Não Iniciado
- **Critérios de Aceitação**:
  - Mensagem "Reivindicação / Criar Reivindicação" com tipo "Reivindicação"
  - DICT cria reivindicação com status "Aberto"
  - Período de resolução de 7 dias
  - PSP reivindicador pode cancelar
  - PSP reivindicador consulta periodicamente
  - PSP doador notificado para responder

### RF-BLO2-006: Reivindicar chave - PSP reivindicador consulta
- **Manual**: Cap 6.1
- **Backlog**: Linha 21
- **Descrição**: PSP reivindicador consulta status de reivindicações existentes para monitorar andamento.
- **Prioridade**: Should Have
- **Complexidade**: Baixa
- **Status**: Não Iniciado
- **Critérios de Aceitação**:
  - Mensagem "Reivindicação / Consultar Reivindicação"
  - Retornar dados da reivindicação (status, datas, PSP doador)
  - Filtros por chave, período, status

### RF-BLO2-007: Reivindicar chave - PSP reivindicador lista reivindicações
- **Manual**: Cap 6.1
- **Backlog**: Linha 20
- **Descrição**: PSP reivindicador lista todas suas reivindicações pendentes/concluídas.
- **Prioridade**: Should Have
- **Complexidade**: Baixa
- **Status**: Não Iniciado

### RF-BLO2-008: Reivindicar chave - PSP reivindicador monitora resposta
- **Manual**: Cap 6.1
- **Backlog**: Linha 19
- **Descrição**: PSP reivindicador recebe (monitora) respostas do PSP doador sobre reivindicação. Status pode mudar para "Confirmado" ou "Cancelado".
- **Prioridade**: Should Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO2-009: Reivindicar chave - PSP reivindicador conclui reivindicação
- **Manual**: Cap 6.1
- **Backlog**: Linha 18
- **Descrição**: Após confirmação do PSP doador, reivindicador conclui o processo. Chave é transferida para o reivindicador.
- **Prioridade**: Should Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO2-010: Reivindicar chave - PSP reivindicador cancela reivindicação
- **Manual**: Cap 6.1
- **Backlog**: Linha 17
- **Descrição**: PSP reivindicador pode cancelar reivindicação aberta ou em resposta.
- **Prioridade**: Should Have
- **Complexidade**: Média
- **Status**: Não Iniciado

### RF-BLO2-011: Reivindicar chave - PSP doador recebe reivindicação
- **Manual**: Cap 6.3
- **Backlog**: Linha 22
- **Descrição**: PSP doador é notificado sobre reivindicação de posse. Precisa avaliar em até 7 dias se aceita ou nega.
- **Prioridade**: Should Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO2-012: Reivindicar chave - PSP doador confirma reivindicação
- **Manual**: Cap 6.3
- **Backlog**: Linha 23
- **Descrição**: PSP doador confirma que chave é do reivindicador e concorda com transferência.
- **Prioridade**: Should Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO2-013: Reivindicar chave - PSP doador cancela reivindicação
- **Manual**: Cap 6.3
- **Backlog**: Linha 24
- **Descrição**: PSP doador nega reivindicação e mantém chave.
- **Prioridade**: Should Have
- **Complexidade**: Média
- **Status**: Não Iniciado

### RF-BLO2-014: Reivindicar chave - PSP doador consulta/lista reivindicações
- **Manual**: Cap 6.3
- **Backlog**: Linhas 25, 26
- **Descrição**: PSP doador consulta detalhes ou lista reivindicações recebidas.
- **Prioridade**: Should Have
- **Complexidade**: Baixa
- **Status**: Não Iniciado

---

## BLOCO 3: VALIDAÇÃO DE CHAVES (Cap 1-2)

Requisitos de validação de posse, compatibilidade com Receita Federal e regras de nome.

### RF-BLO3-001: Validar posse da chave PIX
- **Manual**: Cap 2.1 - Validação da Posse da Chave
- **Backlog**: Linha 31
- **Descrição**: PSP deve validar que usuário é realmente proprietário da chave. Para CPF/CNPJ: verificar se número corresponde à conta aberta. Para telefone/email: enviar código de validação e requerer confirmação via autenticação.
- **Prioridade**: Must Have
- **Complexidade**: Alta
- **Status**: Não Iniciado
- **Critérios de Aceitação**:
  - CPF/CNPJ: validação contra registro de abertura da conta
  - Telefone/Email: código enviado e confirmado via autenticação
  - Registro da validação em auditoria
  - Falha de validação impede registro/alteração

### RF-BLO3-002: Validar situação cadastral na Receita Federal
- **Manual**: Cap 2.2 - Validação da Situação Cadastral do Usuário na Receita Federal
- **Backlog**: Linha 32
- **Descrição**: PSP verifica se CPF ou CNPJ do usuário tem situação irregular na Receita Federal. Situações irregulares para PF: suspensa, cancelada, titular falecido, nula. Para PJ: suspensa (exceto MEI em certos casos), inapta (exceto art. 38), baixada, nula. Usuário em situação irregular NÃO pode registrar/alterar/portabilizar/reivindicar/excluir chaves.
- **Prioridade**: Must Have
- **Complexidade**: Alta
- **Status**: Não Iniciado
- **Critérios de Aceitação**:
  - Integração com dados da Receita Federal (CPF/CNPJ)
  - Validação antes de qualquer operação de chave
  - Bloqueio de operações para situações irregulares
  - Exceção para MEI suspenso (art. 1º Resolução 36/2016 CGSIM)
  - Log de validações

### RF-BLO3-003: Validar nomes vinculados à chave
- **Manual**: Cap 2.3 - Validação dos Nomes Vinculados às Chaves PIX
- **Backlog**: Linha 33
- **Descrição**: Valida que Name (nome civil/razão social) e TradeName (nome social/nome fantasia) estão conforme Receita Federal. Permite variações: diacríticos, troca de caracteres (., -, ,, '), uso de "&" vs "E", maiúsculas/minúsculas, abreviações (com regras estritas para PF). Não permite omissão de palavras.
- **Prioridade**: Must Have
- **Complexidade**: Muito Alta
- **Status**: Não Iniciado
- **Critérios de Aceitação**:
  - PF: nome civil (obrigatório por extenso em 1º e último nome), nome social (se registrado no CPF)
  - PJ: razão social (obrigatória por extenso, incluindo termos jurídicos), nome fantasia (se constar no CNPJ)
  - MEI: sem TradeName
  - Regras de diacríticos: Ã, Õ, Á, É, Í, Ó, Ú, À, È, Ì, Ò, Ù, Â, Ê, Î, Ô, Û, Ä, Ë, Ï, Ö, Ü, Ç, Ñ, Å permitidos
  - Troca permitida: ., ,, -, ' por espaço
  - & vs E aceitável
  - Maiúsculas/minúsculas flexíveis
  - PF: abreviação apenas de nomes intermediários (1º e último por extenso)
  - Nenhuma palavra pode ser omitida (preposições incluídas)
  - Múltiplos espaços reduzidos a um
  - Validar contra expressão regular ou algoritmo de conformidade
  - Rejeitar nomes não conformes com detalhamento

---

## BLOCO 4: DEVOLUÇÃO E INFRAÇÃO (Cap 10, 17)

Requisitos para fluxos de devolução de valores (por fraude ou falha operacional) e notificação de infrações.

### RF-BLO4-001: Solicitar devolução por falha operacional do PSP do pagador
- **Manual**: Cap 17.1 - Solicitação de Devolução por Falha Operacional
- **Backlog**: Linha 36
- **Descrição**: Participante PIX (com acesso direto ao DICT) solicita devolução de valor porque houve falha operacional no PSP do pagador. Fluxo cria solicitação no DICT que é processada conforme regras de devolução.
- **Prioridade**: Should Have
- **Complexidade**: Muito Alta
- **Status**: Não Iniciado
- **Critérios de Aceitação**:
  - Mensagem de solicitação de devolução
  - Justificativa: "falha operacional do PSP do pagador"
  - DICT registra solicitação
  - Status inicial: "Aberto"
  - Notificação ao PSP do pagador
  - Processamento conforme cap 17

### RF-BLO4-002: Solicitar devolução por fundada suspeita de fraude
- **Manual**: Cap 17.2 - Solicitação de Devolução por Fundada Suspeita de Fraude
- **Backlog**: Linha 37
- **Descrição**: Participante PIX solicita devolução porque suspeita que transação foi fraudulenta. Fluxo análogo ao RF-BLO4-001 mas com justificativa de fraude.
- **Prioridade**: Should Have
- **Complexidade**: Muito Alta
- **Status**: Não Iniciado

### RF-BLO4-003: Cancelar solicitação de devolução
- **Manual**: Cap 17.4 - Fluxo de Cancelamento de Devolução
- **Backlog**: Linha 38
- **Descrição**: Participante PIX cancela solicitação de devolução já aberta. Operação permitida se status ainda estiver aberto.
- **Prioridade**: Should Have
- **Complexidade**: Média
- **Status**: Não Iniciado

### RF-BLO4-004: Notificação de infração para abertura de devolução (Acesso Direto)
- **Manual**: Cap 10.1.1 - Notificação de Infração para Solicitação de Devolução (Acesso Direto)
- **Backlog**: Linha 39
- **Descrição**: DICT notifica participante com acesso direto que recebeu denúncia de infração e solicita abertura de devolução. Participante deve processar em prazo.
- **Prioridade**: Should Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO4-005: Notificação de infração para cancelamento de devolução (Acesso Direto)
- **Manual**: Cap 10.1.3 - Notificação de Infração para Cancelamento de Devolução (Acesso Direto)
- **Backlog**: Linha 40
- **Descrição**: DICT notifica que devolução anterior deve ser cancelada. Participante deve processar o cancelamento.
- **Prioridade**: Should Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO4-006: Notificação de infração para marcação de fraude transacional
- **Manual**: Cap 10.2 - Notificação de Infração para Marcação de Fraude Transacional
- **Backlog**: Linha 41
- **Descrição**: DICT notifica que transação foi marcada como fraudulenta. Participante deve registrar para fins de inteligência.
- **Prioridade**: Should Have
- **Complexidade**: Média
- **Status**: Não Iniciado

---

## BLOCO 5: SEGURANÇA E INFRAESTRUTURA (Cap 9, 11-19)

Requisitos de mecanismos de segurança, comunicação, cache e prevenção a ataques.

### RF-BLO5-001: Verificação de sincronismo (VSYNC) - Acesso Direto
- **Manual**: Cap 9.1 - Verificação de VSYNC (Acesso Direto ao DICT)
- **Backlog**: Linha 56
- **Descrição**: PSP com acesso direto realiza verificação periódica de sincronismo entre sua base de dados interna e o DICT. Identifica chaves não sincronizadas e dispara registros/exclusões automáticas.
- **Prioridade**: Should Have
- **Complexidade**: Muito Alta
- **Status**: Não Iniciado
- **Critérios de Aceitação**:
  - Mensagem "Diretório / Verificar Sincronismo"
  - DICT retorna lista de chaves registradas para o PSP
  - PSP compara com base interna
  - PSP identifica diferenças: (i) chaves no DICT que não estão internamente, (ii) chaves internas não no DICT
  - PSP dispara registros para chaves faltando no DICT
  - PSP dispara exclusões para chaves não mais desejadas
  - Periodicidade: conforme necessidade (recomendado frequente)

### RF-BLO5-002: Lista de CIDs (Customer IDs)
- **Manual**: Cap 9.3 - Lista de CIDs
- **Backlog**: Linha 57
- **Descrição**: DICT retorna lista de identificadores de cliente (CIDs) do PSP, útil para sincronismo. Implementação diferencia acesso direto e indireto.
- **Prioridade**: Should Have
- **Complexidade**: Média
- **Status**: Não Iniciado

### RF-BLO5-003: Interface de comunicação
- **Manual**: Cap 11 - Interface de Comunicação
- **Backlog**: Linha 52
- **Descrição**: Infraestrutura de comunicação entre PSPs e DICT. Define padrões de mensagem, protocolos, autenticação, autorização e tratamento de erros.
- **Prioridade**: Must Have
- **Complexidade**: Muito Alta
- **Status**: Não Iniciado
- **Critérios de Aceitação**:
  - Protocolo de comunicação definido (REST, gRPC, etc)
  - Autenticação de PSPs
  - Autorização granular por operação
  - Tratamento de erros padronizado
  - Resiliência (retry, timeout)
  - Logging e auditoria
  - Documentação de API

### RF-BLO5-004: Cache de chaves consultadas
- **Manual**: Cap 12 - Cache de Chaves Consultadas
- **Backlog**: Linha 54
- **Descrição**: PSP implementa cache de chaves que já consultou no DICT para evitar requisições repetidas e melhorar performance. Deve respeitar regras de expiração.
- **Prioridade**: Nice to Have
- **Complexidade**: Média
- **Status**: Não Iniciado

### RF-BLO5-005: Mecanismo de prevenção a ataque de leitura - Verificação de autenticidade
- **Manual**: Cap 13.2.1 - Verificação de Autenticidade do Usuário Solicitante da Consulta
- **Backlog**: Linha 44
- **Descrição**: PSP implementa verificação de autenticidade: só usuários autenticados podem consultar chaves. Previne ataques de leitura não autorizada.
- **Prioridade**: Should Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO5-006: Mecanismo de prevenção a ataque de leitura - Política de limitação
- **Manual**: Cap 13.2.2 - Estabelecimento de Política Interna de Limitação de Consultas
- **Backlog**: Linha 45
- **Descrição**: PSP estabelece limite de consultas por usuário/período (ex: máximo N consultas por minuto). Bloqueia usuários que excedem.
- **Prioridade**: Should Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO5-007: Mecanismo de prevenção a ataque de leitura - Monitoramento
- **Manual**: Cap 13.2.3 - Monitoramento Qualitativo e Permanente das Consultas
- **Backlog**: Linha 46
- **Descrição**: PSP monitora padrões de consulta para detectar comportamentos anormais (muitas consultas a chaves aleatórias, consultas fora do expediente, IPs suspeitos).
- **Prioridade**: Should Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO5-008: Mecanismo de prevenção a ataque de leitura - Plano de ação
- **Manual**: Cap 13.2.4 - Plano de Ação para Tratamento de Casos Suspeitos
- **Backlog**: Linha 47
- **Descrição**: PSP tem protocolo para responder a suspicion de ataques de leitura. Ações: bloqueio de usuário, notificação, investigação.
- **Prioridade**: Should Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO5-009: Mecanismo de prevenção a ataque de leitura - Restrição de dados
- **Manual**: Cap 13.2.5 - Restrição dos Dados da Chave Exibidos ao Usuário que Faz a Consulta
- **Backlog**: Linha 48
- **Descrição**: PSP restringe dados exibidos ao usuário que consulta chave. Exemplo: não mostrar ISPB/agência/conta completa para usuários não autorizados.
- **Prioridade**: Should Have
- **Complexidade**: Média
- **Status**: Não Iniciado

### RF-BLO5-010: Limitação de requisições à API do DICT
- **Manual**: Cap 14 - Limitação de Requisições à API do DICT
- **Backlog**: Linha 49
- **Descrição**: DICT implementa rate limiting: máximo N requisições por segundo/minuto por PSP. Previne DDoS e abuso.
- **Prioridade**: Should Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO5-011: Verificação de chaves PIX registradas
- **Manual**: Cap 15 - Fluxo de Verificação de Chaves PIX Registradas
- **Backlog**: Linha 51
- **Descrição**: Participante PIX consulta se chave(s) específica(s) estão registradas no DICT (sem obter dados completos). Útil para validação rápida.
- **Prioridade**: Should Have
- **Complexidade**: Média
- **Status**: Não Iniciado

### RF-BLO5-012: Cache de existência de chave PIX
- **Manual**: Cap 16 - Cache de Existência de Chave PIX
- **Backlog**: Linha 53
- **Descrição**: PSP implementa cache de existência de chaves (booleano: existe ou não). Melhora performance sem expor dados. Respeita expiração.
- **Prioridade**: Nice to Have
- **Complexidade**: Média
- **Status**: Não Iniciado

### RF-BLO5-013: Consulta a informações de segurança
- **Manual**: Cap 18 - Fluxo de Consulta a Informações de Segurança
- **Backlog**: Linha 55
- **Descrição**: Participante PIX consulta informações de segurança sobre uma chave (ex: se foi marcada como fraude, status de bloqueio judicial). Acesso restrito a PSPs autorizados.
- **Prioridade**: Should Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

---

## BLOCO 6: RECUPERAÇÃO DE VALORES (Cap 20)

Requisitos para fluxos de recuperação de valores em transações com fraude ou falha operacional.

### RF-BLO6-001: Instauração no fluxo interativo - Início manual
- **Manual**: Cap 20.1.1 - Instauração no Fluxo Interativo
- **Backlog**: Linha 60
- **Descrição**: Operador/sistema inicia manualmente processo de recuperação de valores em fluxo interativo. Define chave, valor, motivo e conta destino.
- **Prioridade**: Nice to Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO6-002: Rastreamento no fluxo interativo
- **Manual**: Cap 20.1.2 - Rastreamento no Fluxo Interativo
- **Backlog**: Linha 61
- **Descrição**: Sistema rastreia movimentação dos valores durante processo de recuperação. Registra cada etapa, transferências, bloqueios.
- **Prioridade**: Nice to Have
- **Complexidade**: Muito Alta
- **Status**: Não Iniciado

### RF-BLO6-003: Priorização no fluxo interativo
- **Manual**: Cap 20.1.3 - Priorização no Fluxo Interativo
- **Backlog**: Linha 62
- **Descrição**: Sistema permite definir prioridade de processos de recuperação (crítico, alto, médio, baixo). Afeta ordem de processamento.
- **Prioridade**: Nice to Have
- **Complexidade**: Média
- **Status**: Não Iniciado

### RF-BLO6-004: Solicitação de bloqueio no fluxo interativo
- **Manual**: Cap 20.1.4 - Solicitação de Bloqueio no Fluxo Interativo
- **Backlog**: Linha 63
- **Descrição**: Sistema solicita bloqueio de valores suspeitos em contas específicas. Bloqueio impede movimentação até desbloqueio.
- **Prioridade**: Nice to Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO6-005: Instauração no fluxo automatizado
- **Manual**: Cap 20.1.5 - Instauração no Fluxo Automatizado
- **Backlog**: Linha 64
- **Descrição**: Sistema inicia automaticamente processo de recuperação baseado em regras pré-configuradas. Exemplo: transação marcada como fraude = recuperação automática.
- **Prioridade**: Nice to Have
- **Complexidade**: Muito Alta
- **Status**: Não Iniciado

### RF-BLO6-006: Etapa de análise
- **Manual**: Cap 20.1.6 - Etapa de Análise
- **Backlog**: Linha 65
- **Descrição**: Especialista analisa solicitação de recuperação: verifica dados, justificativas, conformidade. Aprova ou rejeita.
- **Prioridade**: Nice to Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO6-007: Etapa de devolução
- **Manual**: Cap 20.1.7 - Etapa de Devolução
- **Backlog**: Linha 66
- **Descrição**: Sistema transfere valores recuperados para conta legítima do proprietário.
- **Prioridade**: Nice to Have
- **Complexidade**: Muito Alta
- **Status**: Não Iniciado

### RF-BLO6-008: Desbloqueio dos recursos
- **Manual**: Cap 20.1.8 - Desbloqueio dos Recursos
- **Backlog**: Linha 67
- **Descrição**: Após análise e se não há fraude confirmada, sistema desbloquia valores suspeitos. Libera acesso ao titular.
- **Prioridade**: Nice to Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO6-009: Recuperação para transações já liquidadas
- **Manual**: Cap 20.1.9 - Recuperação de Valores para Transações Liquidadas nos Sistemas dos Participantes
- **Backlog**: Linha 68
- **Descrição**: Fluxo especial para recuperar valores que já foram processados e liquidados no sistema do participante. Requer coordenação entre sistemas.
- **Prioridade**: Nice to Have
- **Complexidade**: Muito Alta
- **Status**: Não Iniciado

### RF-BLO6-010: Fluxo interativo - Instauração e Bloqueio
- **Manual**: Cap 20.2 - Fluxo de Instauração e Solicitação de Bloqueio no Fluxo Interativo
- **Backlog**: Linha 69
- **Descrição**: Fluxo completo de instauração manual e solicitação de bloqueio em fluxo interativo.
- **Prioridade**: Nice to Have
- **Complexidade**: Muito Alta
- **Status**: Não Iniciado

### RF-BLO6-011: Fluxo automatizado - Instauração e Bloqueio
- **Manual**: Cap 20.3 - Fluxo de Instauração e Solicitação de Bloqueio no Fluxo Automatizado
- **Backlog**: Linha 70
- **Descrição**: Fluxo completo de instauração automática e solicitação de bloqueio em fluxo automatizado.
- **Prioridade**: Nice to Have
- **Complexidade**: Muito Alta
- **Status**: Não Iniciado

### RF-BLO6-012: Fluxo de análise
- **Manual**: Cap 20.4 - Fluxo de Análise
- **Backlog**: Linha 71
- **Descrição**: Fluxo de análise de solicitação de recuperação por especialista.
- **Prioridade**: Nice to Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-BLO6-013: Fluxo de devolução
- **Manual**: Cap 20.5 - Fluxo de Devolução
- **Backlog**: Linha 72
- **Descrição**: Fluxo de devolução de valores após aprovação em análise.
- **Prioridade**: Nice to Have
- **Complexidade**: Muito Alta
- **Status**: Não Iniciado

---

## Características Funcionais Transversais

### RF-TRANS-001: Consulta de Baldes (Buckets)
- **Manual**: Cap 19 - Consulta de Baldes
- **Backlog**: Linha 50
- **Descrição**: DICT disponibiliza consulta de "baldes" (agrupamentos de chaves por características). Usado para análise de tendências e segurança.
- **Prioridade**: Nice to Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-TRANS-002: Notificações de Eventos
- **Manual**: Cap 21 - Notificações de Eventos
- **Backlog**: (Implícito em vários fluxos)
- **Descrição**: Sistema envia notificações para PSPs sobre eventos: mudança de status, novas reivindicações, infrações, etc. Pode ser síncronas ou assíncronas.
- **Prioridade**: Must Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

### RF-TRANS-003: Bloqueio judicial de chave
- **Manual**: Cap 1.1 - Chaves Bloqueadas por Ordem Judicial
- **Backlog**: Linha 9
- **Descrição**: DICT recebe ordem judicial para bloquear chave. Retorna erro "EntryBlocked" em operações de consulta, alteração, exclusão, portabilidade e reivindicação. PSP também bloqueia internamente.
- **Prioridade**: Must Have
- **Complexidade**: Média
- **Status**: Não Iniciado

### RF-TRANS-004: Auditoria e Logging
- **Manual**: (Implícito)
- **Backlog**: (Implícito)
- **Descrição**: Sistema registra todas operações (criação, leitura, atualização, exclusão de chaves) com timestamp, usuário/PSP, IP, resultado. Rastreável por período, chave, usuário.
- **Prioridade**: Must Have
- **Complexidade**: Alta
- **Status**: Não Iniciado

---

## Matriz de Rastreabilidade

### RF-Bloco1 (CRUD)
| ID | Descrição Resumida | Manual Cap | Backlog L | Status |
|----|--------------------|------------|-----------|--------|
| RF-BLO1-001 | Registrar chave - Acesso Direto | 3.1 | 3 | Parcial |
| RF-BLO1-002 | Registrar chave - Acesso Indireto | 3.2 | 3 | Não Iniciado |
| RF-BLO1-003 | Registrar chave - Operador Direto | 3.3 | 3 | Não Iniciado |
| RF-BLO1-004 | Registrar chave - Operador Indireto | 3.4 | 3 | Não Iniciado |
| RF-BLO1-005 | Excluir chave - RFB Incompatibilidade | 4.1 | 4 | Não Iniciado |
| RF-BLO1-006 | Excluir chave - Usuário Direto | 4.2 | 5 | Parcial |
| RF-BLO1-007 | Excluir chave - Usuário Indireto | 4.3 | 5 | Não Iniciado |
| RF-BLO1-008 | Excluir chave - Operador Direto | 4.4 | 6 | Não Iniciado |
| RF-BLO1-009 | Excluir chave - Operador Indireto | 4.5 | 7 | Não Iniciado |
| RF-BLO1-010 | Alterar dados chave - Direto | 7.1 | 11 | Não Iniciado |
| RF-BLO1-011 | Alterar dados chave - Indireto | 7.2 | 11 | Não Iniciado |
| RF-BLO1-012 | Alterar dados - Correção | 7.3 | 12 | Não Iniciado |
| RF-BLO1-013 | Consultar chave | 8.2-8.3 | 10 | Parcial |

### RF-Bloco2 (Reivindicação)
| ID | Descrição Resumida | Manual Cap | Backlog L | Status |
|----|--------------------|------------|-----------|--------|
| RF-BLO2-001 | Portabilidade - Reivindicador Direto | 5.1 | 27 | Não Iniciado |
| RF-BLO2-002 | Portabilidade - Reivindicador Indireto | 5.2 | 27 | Não Iniciado |
| RF-BLO2-003 | Portabilidade - Doador Direto | 5.3 | 28 | Não Iniciado |
| RF-BLO2-004 | Portabilidade - Doador Indireto | 5.4 | 28 | Não Iniciado |
| RF-BLO2-005 | Reivindicar - Criação | 6.1 | 16 | Não Iniciado |
| RF-BLO2-006 | Reivindicar - Consultar | 6.1 | 21 | Não Iniciado |
| RF-BLO2-007 | Reivindicar - Listar | 6.1 | 20 | Não Iniciado |
| RF-BLO2-008 | Reivindicar - Monitorar | 6.1 | 19 | Não Iniciado |
| RF-BLO2-009 | Reivindicar - Concluir | 6.1 | 18 | Não Iniciado |
| RF-BLO2-010 | Reivindicar - Cancelar | 6.1 | 17 | Não Iniciado |
| RF-BLO2-011 | Reivindicar - Doador Recebe | 6.3 | 22 | Não Iniciado |
| RF-BLO2-012 | Reivindicar - Doador Confirma | 6.3 | 23 | Não Iniciado |
| RF-BLO2-013 | Reivindicar - Doador Cancela | 6.3 | 24 | Não Iniciado |
| RF-BLO2-014 | Reivindicar - Doador Consulta/Lista | 6.3 | 25-26 | Não Iniciado |

### RF-Bloco3 (Validação)
| ID | Descrição Resumida | Manual Cap | Backlog L | Status |
|----|--------------------|------------|-----------|--------|
| RF-BLO3-001 | Validar posse da chave | 2.1 | 31 | Não Iniciado |
| RF-BLO3-002 | Validar situação RFB | 2.2 | 32 | Não Iniciado |
| RF-BLO3-003 | Validar nomes | 2.3 | 33 | Não Iniciado |

### RF-Bloco4 (Devolução/Infração)
| ID | Descrição Resumida | Manual Cap | Backlog L | Status |
|----|--------------------|------------|-----------|--------|
| RF-BLO4-001 | Devolução - Falha Operacional | 17.1 | 36 | Não Iniciado |
| RF-BLO4-002 | Devolução - Fraude Suspeita | 17.2 | 37 | Não Iniciado |
| RF-BLO4-003 | Devolução - Cancelamento | 17.4 | 38 | Não Iniciado |
| RF-BLO4-004 | Infração - Abertura Devolução Direto | 10.1.1 | 39 | Não Iniciado |
| RF-BLO4-005 | Infração - Cancelamento Direto | 10.1.3 | 40 | Não Iniciado |
| RF-BLO4-006 | Infração - Fraude Transacional | 10.2 | 41 | Não Iniciado |

### RF-Bloco5 (Segurança)
| ID | Descrição Resumida | Manual Cap | Backlog L | Status |
|----|--------------------|------------|-----------|--------|
| RF-BLO5-001 | Verificação VSYNC Direto | 9.1 | 56 | Não Iniciado |
| RF-BLO5-002 | Lista de CIDs | 9.3 | 57 | Não Iniciado |
| RF-BLO5-003 | Interface Comunicação | 11 | 52 | Não Iniciado |
| RF-BLO5-004 | Cache Chaves Consultadas | 12 | 54 | Não Iniciado |
| RF-BLO5-005 | Prevenção Leitura - Autenticidade | 13.2.1 | 44 | Não Iniciado |
| RF-BLO5-006 | Prevenção Leitura - Limitação | 13.2.2 | 45 | Não Iniciado |
| RF-BLO5-007 | Prevenção Leitura - Monitoramento | 13.2.3 | 46 | Não Iniciado |
| RF-BLO5-008 | Prevenção Leitura - Plano Ação | 13.2.4 | 47 | Não Iniciado |
| RF-BLO5-009 | Prevenção Leitura - Restrição Dados | 13.2.5 | 48 | Não Iniciado |
| RF-BLO5-010 | Rate Limiting API DICT | 14 | 49 | Não Iniciado |
| RF-BLO5-011 | Verificação Chaves Registradas | 15 | 51 | Não Iniciado |
| RF-BLO5-012 | Cache Existência Chave | 16 | 53 | Não Iniciado |
| RF-BLO5-013 | Consulta Informações Segurança | 18 | 55 | Não Iniciado |

### RF-Bloco6 (Recuperação)
| ID | Descrição Resumida | Manual Cap | Backlog L | Status |
|----|--------------------|------------|-----------|--------|
| RF-BLO6-001 | Recuperação - Instauração Manual | 20.1.1 | 60 | Não Iniciado |
| RF-BLO6-002 | Recuperação - Rastreamento | 20.1.2 | 61 | Não Iniciado |
| RF-BLO6-003 | Recuperação - Priorização | 20.1.3 | 62 | Não Iniciado |
| RF-BLO6-004 | Recuperação - Bloqueio | 20.1.4 | 63 | Não Iniciado |
| RF-BLO6-005 | Recuperação - Instauração Auto | 20.1.5 | 64 | Não Iniciado |
| RF-BLO6-006 | Recuperação - Análise | 20.1.6 | 65 | Não Iniciado |
| RF-BLO6-007 | Recuperação - Devolução | 20.1.7 | 66 | Não Iniciado |
| RF-BLO6-008 | Recuperação - Desbloqueio | 20.1.8 | 67 | Não Iniciado |
| RF-BLO6-009 | Recuperação - Transações Liquidadas | 20.1.9 | 68 | Não Iniciado |
| RF-BLO6-010 | Recuperação - Fluxo Interativo | 20.2 | 69 | Não Iniciado |
| RF-BLO6-011 | Recuperação - Fluxo Auto | 20.3 | 70 | Não Iniciado |
| RF-BLO6-012 | Recuperação - Fluxo Análise | 20.4 | 71 | Não Iniciado |
| RF-BLO6-013 | Recuperação - Fluxo Devolução | 20.5 | 72 | Não Iniciado |

---

## Top 10 Requisitos Críticos (Priorizados)

### Categoria: Must Have + Alta Criticidade

1. **RF-BLO1-001/002**: Registrar chave PIX
   - Funcionalidade básica, bloqueador para todo o sistema
   - Depende de: RF-BLO3-001, RF-BLO3-002, RF-BLO5-003
   - Impacto: Crítico

2. **RF-BLO1-006/007**: Excluir chave por solicitação do usuário
   - Direito fundamental do usuário, requisito regulatório
   - Complexidade: Média
   - Impacto: Crítico

3. **RF-BLO3-002**: Validar situação cadastral na Receita Federal
   - Bloqueador: todas operações dependem desta validação
   - Integração externa crítica
   - Impacto: Crítico

4. **RF-BLO3-001**: Validar posse da chave
   - Segurança fundamental, requerido antes de qualquer operação
   - Complexidade: Alta
   - Impacto: Crítico

5. **RF-BLO3-003**: Validar nomes vinculados à chave
   - Conformidade com Receita Federal, requerido
   - Complexidade: Muito Alta (regras complexas)
   - Impacto: Alto

6. **RF-BLO1-010/011**: Alterar dados vinculados à chave
   - Funcionalidade essencial para correção de erros
   - Prioridade operacional alta
   - Impacto: Alto

7. **RF-BLO2-001/003**: Fluxo de Portabilidade
   - Requisito regulatório PIX (direito do usuário)
   - Muito complexo (coordenação entre PSPs, timeouts de 7 dias)
   - Impacto: Crítico

8. **RF-BLO5-003**: Interface de Comunicação
   - Bloqueador: infraestrutura necessária para todas operações
   - Autenticação, autorização, mensageria
   - Impacto: Crítico

9. **RF-BLO2-005/012**: Fluxo de Reivindicação de Posse
   - Requisito regulatório PIX (direito do usuário)
   - Complexidade muito alta, coordenação entre PSPs
   - Impacto: Crítico

10. **RF-TRANS-003/004**: Bloqueio judicial e Auditoria
    - Conformidade legal e regulatória
    - Suporte a investigações e auditoria
    - Impacto: Alto

---

## Gaps vs Implementação Atual

### Baseado em ARE-002 (Status Atual)

**Componentes em Desenvolvimento:**
- DICT: Alguns requisitos (estimated 5-10% do Bloco 1)
- Bridge: Alguns requisitos (estimated 5-10% do Bloco 1)

**Componentes Não Iniciados:**
- Core: Nenhum requisito de DICT (pronto para iniciar)
- Comunicação: Interface ainda não estabelecida
- Arquivos/Repo: Não aplicável

### Gaps Principais Identificados

#### Gap 1: Blocos 2-6 Não Iniciados
- **Reivindicação (14 RFs)**: Nenhum desenvolvimento iniciado
- **Devolução/Infração (6 RFs)**: Nenhum desenvolvimento iniciado
- **Recuperação (13 RFs)**: Nenhum desenvolvimento iniciado
- **Impacto**: 33 RFs críticos sem implementação

#### Gap 2: Validações Incompletas
- **RF-BLO3-001/002/003**: Validação de posse, RFB, nomes não implementadas
- **Impacto**: Risco de chaves inválidas/fraudulentas no DICT

#### Gap 3: Segurança Básica Ausente
- **RF-BLO5-003**: Interface de comunicação sem autenticação/autorização
- **RF-BLO5-001**: Sincronismo não implementado
- **RF-TRANS-004**: Auditoria/logging não implementado
- **Impacto**: Acesso não controlado, impossibilidade de rastreamento

#### Gap 4: Fluxos Multi-PSP Não Iniciados
- **Acesso Indireto**: Todas variantes de acesso indireto (RF-BLO1-002/004/007/011, etc.)
- **Impacto**: Exclusão de 50% dos participantes PIX (PSPs sem acesso direto)

#### Gap 5: Transações Jurídicas/Fraudulentas
- **Bloqueio Judicial**: RF-TRANS-003 não implementado
- **Devolução por Fraude**: RF-BLO4-002 não implementado
- **Notificação Infração**: RF-BLO4-004/005/006 não implementadas
- **Impacto**: Impossibilidade de responder a fraudes/ordens judiciais

#### Gap 6: Inteligência Operacional
- **Sincronismo (RF-BLO5-001)**: PSPs cegos para estado do DICT
- **Cache/Otimização (RF-BLO5-004/012)**: Sem otimização de performance
- **Impacto**: Operacional ineficiente, possíveis inconsistências

---

## Recomendações de Priorização

### Fase 1 (Máximo Impacto): RFs Bloqueadores
**Timeline**: 4-6 semanas  
**Entregas**: 7 RFs

1. RF-BLO5-003 - Interface de Comunicação (bloqueador absoluto)
2. RF-BLO3-002 - Validar RFB (bloqueador regulatório)
3. RF-BLO1-001 - Registrar chave direto (funcionalidade básica)
4. RF-TRANS-004 - Auditoria/Logging (suporte a todas operações)
5. RF-TRANS-003 - Bloqueio judicial (conformidade legal)
6. RF-BLO3-001 - Validar posse (segurança)
7. RF-BLO3-003 - Validar nomes (conformidade RFB)

### Fase 2 (Completar CRUD): RFs Essenciais
**Timeline**: 6-8 semanas  
**Entregas**: 13 RFs (todos Bloco 1)

1. RF-BLO1-002-004 - Registrar chave (variantes)
2. RF-BLO1-005-009 - Excluir chave (todas variantes)
3. RF-BLO1-010-012 - Alterar dados chave
4. RF-BLO1-013 - Consultar chave

### Fase 3 (Requisitos Regulatórios): Reivindicação + Portabilidade
**Timeline**: 8-12 semanas  
**Entregas**: 14 RFs (Bloco 2)

Foco: RF-BLO2-001/003/005/012 (casos de sucesso primários)

### Fase 4+ (Remaining)
- Bloco 4: Devolução/Infração (6 RFs) - 4 semanas
- Bloco 5: Segurança Avançada (13 RFs) - 6 semanas
- Bloco 6: Recuperação Valores (13 RFs) - 8 semanas

---

## Conclusões

Este documento estabelece base sólida para rastreabilidade de requisitos funcionais DICT. Com 72 RFs mapeados, é possível:

- Planejar sprints com precisão
- Rastrear progresso contra requisitos regulatórios
- Identificar dependências e bloqueadores
- Validar completude de implementação

**Total de Trabalho Estimado**: ~1000-1200 horas de desenvolvimento (6-8 meses para equipe de 2-3 devs)

**Status Crítico**: 86% dos RFs ainda não iniciados - ação imediata recomendada para Fase 1.

---

**Documento Gerado**: 2025-10-24  
**Próxima Revisão Recomendada**: 2025-11-07 (após conclusão Fase 1)  
**Responsável**: ORACLE (AGT-BA-001)

