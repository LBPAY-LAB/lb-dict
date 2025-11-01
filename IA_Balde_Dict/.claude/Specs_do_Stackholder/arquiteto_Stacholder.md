# Arquitetura Balde de Fichas

O mecanismo de limitação de requisições, ou *rate limit*, na API do DICT funciona para preservar a estabilidade do serviço e, especialmente, para prevenir ataques de varredura de dados. Ele é implementado utilizando o algoritmo **token bucket*- (balde de fichas).

![alt text](<images/Worker Rate Limit Component Diagram (Current).png>)

## Utilização no DICT

![alt text](<images/DICT Proxy Component Diagram (Current).png>)

## Como Funciona o Algoritmo Token Bucket

A lógica do *token bucket- pode ser entendida como um "balde" que armazena "fichas" para cada política de limitação.

1. **O Balde (Bucket)**: Cada política de limitação possui um balde com uma capacidade máxima específica ("Tamanho do balde").
2. **As Fichas (Tokens)**: O balde é reabastecido com fichas a uma taxa de reposição constante (por exemplo, 1200 fichas por minuto).
3. **Consumo de Fichas**: Cada requisição feita a uma operação da API consome (debita) uma ou mais fichas do balde correspondente.
4. **Violação do Limite**: Se um participante tentar fazer uma requisição quando o balde estiver vazio (sem fichas), a API rejeitará a solicitação e retornará um erro com o status **HTTP 429 (`RateLimited`)**. A resposta pode incluir um cabeçalho `Retry-After` com a previsão de tempo em segundos para que o saldo se torne positivo novamente.

## Políticas e Escopos

Existem diversas políticas de limitação, cada uma aplicada a operações específicas da API. Elas são definidas com base em um **escopo**, que pode ser:

- **PSP**: A limitação é aplicada ao participante como um todo.
- **USER**: A limitação é aplicada individualmente a cada usuário final (identificado pelo seu CPF ou CNPJ no cabeçalho `PI-PayerId` da requisição).

## Políticas de Limitação em Detalhe

A API do DICT define políticas específicas para diferentes grupos de operações. As mais complexas são as de consulta de chaves, projetadas para evitar varreduras de dados.

### 1. Políticas Anti-Varredura para Consulta de Vínculo (`getEntry`)

Estas são as políticas mais sofisticadas, pois recompensam o uso legítimo (consulta seguida de pagamento) e penalizam o uso indevido (muitas consultas sem resultado ou sem pagamento).

#### Escopo do Usuário Final (USER)

- Existem duas políticas, uma para chaves do tipo `EMAIL` e `PHONE` (`ENTRIES_READ_USER_ANTISCAN`) e outra para `CPF`, `CNPJ` e `EVP` (`ENTRIES_READ_USER_ANTISCAN_V2`).
- Os limites variam se o usuário final é Pessoa Física (PF) ou Pessoa Jurídica (PJ):
  - **PF**: Balde de **100 fichas**, com reposição de **2/minuto**.
  - **PJ**: Balde de **1.000 fichas**, com reposição de **20/minuto**.
- **Regras de contagem (consumo e crédito de fichas)**:
  - Consulta bem-sucedida (status 200): **subtrai 1 ficha**.
  - Consulta a uma chave inexistente (status 404): **subtrai 20 fichas*- (uma penalidade para evitar "adivinhação" de chaves).
  - Envio de uma ordem de pagamento (pacs.008) associada à consulta: *adiciona 1 ficha*- (para PF) ou *2 fichas*- (para PJ), repondo o balde.

#### Escopo do Participante (PSP)

- A política `ENTRIES_READ_PARTICIPANT_ANTISCAN` aplica-se a todos os tipos de chave.
- Os limites de taxa de reposição e tamanho do balde dependem da **categoria do participante (A a H)**, variando de 25.000/min para a categoria A até 2/min para a categoria H.

Regras de contagem:

- Consulta bem-sucedida (status 200): **subtrai 1 ficha**.
- Consulta a uma chave inexistente (status 404): **subtrai 3 fichas**.
- Envio de uma ordem de pagamento associada: **adiciona 1 ficha**.

### 2. Políticas para Outras Operações

Outras operações possuem políticas mais simples, geralmente baseadas apenas no participante (PSP) e subtraindo 1 ficha por requisição (que não resulte em erro 500).

Exemplos de limites para o PSP:

- **`ENTRIES_WRITE`*- (criar ou remover vínculo): 1200 requisições/minuto, com balde de 36.000.
- **`ENTRIES_UPDATE`*- (atualizar vínculo): 600 requisições/minuto, com balde de 600.
- **`SYNC_VERIFICATIONS_WRITE`*- (criar verificação de VSync): 10 requisições/minuto, com balde de 50.
- **`CIDS_EVENTS_LIST`*- (listar eventos de CID): 20 requisições/minuto, com balde de 100.
- **`KEYS_CHECK`*- (verificar existência de chaves): 70 requisições/minuto, com balde de 70.

### Como Monitorar os Limites

Os participantes podem e devem monitorar o estado de seus baldes para gerenciar melhor o fluxo de requisições e evitar o erro 429. O DICT disponibiliza endpoints específicos para isso:

- `listBucketStates`: Retorna uma lista de todas as políticas de limitação aplicáveis ao participante e o estado atual de cada balde.
- `getBucketState`: Retorna o estado atual de um balde para uma política específica informada.