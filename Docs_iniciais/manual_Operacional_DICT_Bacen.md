# Manual Operacional

# do Diretório de

# Identificadores de

# Contas Transacionais

# (DICT)

## Versão 8.


## Sumário



- 1 CHAVES PIX
   - 1.1 CHAVES BLOQUEADAS POR ORDEM JUDICIAL
- 2 VALIDAÇÃO DAS CHAVES PIX
   - 2.1 VALIDAÇÃO DA POSSE DA CHAVE
   - 2.2 VALIDAÇÃO DA SITUAÇÃO CADASTRAL DO USUÁRIO NA RECEITA FEDERAL
   - 2.3 VALIDAÇÃO DOS NOMES VINCULADOS ÀS CHAVES PIX
- 3 FLUXO DE REGISTRO DE CHAVE
   - 3.1 FLUXO DE REGISTRO DE CHAVE POR SOLICITAÇÃO DO USUÁRIO FINAL (PARTICIPANTES DO PIX COM ACESSO DIRETO AO DICT)
   - DICT) 3.2 FLUXO DE REGISTRO DE CHAVE POR SOLICITAÇÃO DO USUÁRIO FINAL (PARTICIPANTES DO PIX COM ACESSO INDIRETO AO
   - 3.3 FLUXO DE REGISTRO DE CHAVE INICIADO PELO PARTICIPANTE (PARTICIPANTES DO PIX COM ACESSO DIRETO AO DICT)
   - 3.4 FLUXO DE REGISTRO DE CHAVE INICIADO PELO PARTICIPANTE (PARTICIPANTES DO PIX COM ACESSO INDIRETO AO DICT)
- 4 FLUXO DE EXCLUSÃO DE CHAVE
   - 4.1 EXCLUSÃO DE CHAVE POR INCOMPATIBILIDADE DE DADOS COM A RECEITA FEDERAL
   - DICT) 4.2 FLUXO DE EXCLUSÃO DE CHAVE POR SOLICITAÇÃO DO USUÁRIO FINAL (PARTICIPANTES DO PIX COM ACESSO DIRETO AO
   - DICT) 4.3 FLUXO DE EXCLUSÃO DE CHAVE POR SOLICITAÇÃO DO USUÁRIO FINAL (PARTICIPANTES DO PIX COM ACESSO INDIRETO AO
   - 4.4 FLUXO DE EXCLUSÃO DE CHAVE INICIADO PELO PARTICIPANTE (PARTICIPANTES DO PIX COM ACESSO DIRETO AO DICT)
   - 4.5 FLUXO DE EXCLUSÃO DE CHAVE INICIADO PELO PARTICIPANTE (PARTICIPANTES DO PIX COM ACESSO INDIRETO AO DICT)..
- 5 FLUXO DE PORTABILIDADE DE CHAVE
   - 5.1 FLUXO DE PORTABILIDADE PARA O PSP REIVINDICADOR COM ACESSO DIRETO AO DICT
   - 5.2 FLUXO DE PORTABILIDADE PARA O PSP REIVINDICADOR COM ACESSO INDIRETO AO DICT
   - 5.3 FLUXO DE PORTABILIDADE PARA O PSP DOADOR COM ACESSO DIRETO AO DICT
   - 5.4 FLUXO DE PORTABILIDADE PARA O PSP DOADOR COM ACESSO INDIRETO AO DICT
- 6 FLUXO DE REIVINDICAÇÃO DE POSSE DE CHAVE
   - 6.1 FLUXO DE REIVINDICAÇÃO DE POSSE PARA O PSP REIVINDICADOR COM ACESSO DIRETO AO DICT
   - 6.2 FLUXO DE REIVINDICAÇÃO DE POSSE PARA O PSP REIVINDICADOR COM ACESSO INDIRETO AO DICT
   - 6.3 FLUXO DE REIVINDICAÇÃO DE POSSE PARA O PSP DOADOR COM ACESSO DIRETO AO DICT
   - 6.4 FLUXO DE REIVINDICAÇÃO DE POSSE PARA O PSP DOADOR COM ACESSO INDIRETO AO DICT
- 7 FLUXO DE ALTERAÇÃO DOS DADOS VINCULADOS À CHAVE
   - 7.1 FLUXO DE ALTERAÇÃO DOS DADOS VINCULADOS À CHAVE (PARTICIPANTES DO PIX COM ACESSO DIRETO AO DICT)
   - 7.2 FLUXO DE ALTERAÇÃO DOS DADOS VINCULADOS À CHAVE (PARTICIPANTES DO PIX COM ACESSO INDIRETO AO DICT)
   - 7.3 ALTERAÇÃO DOS DADOS VINCULADOS À CHAVE PARA CORREÇÃO DE INCONSISTÊNCIAS
- 8 FLUXO DE CONSULTA DE CHAVE
   - 8.1 DADOS DE CHAVE PERMITIDOS NA EXIBIÇÃO AO USUÁRIO
   - 8.2 FLUXO DE CONSULTA PARA O PARTICIPANTE DO PIX COM ACESSO DIRETO AO DICT...................................................
   - 8.3 FLUXO DE CONSULTA PARA O PARTICIPANTE DO PIX COM ACESSO INDIRETO AO DICT
   - DE PAGAMENTO, COM ACESSO DIRETO AO DICT 8.4 FLUXO DE CONSULTA PARA O PARTICIPANTE DO PIX QUE ATUA COMO PRESTADOR DE SERVIÇO DE INICIAÇÃO DE TRANSAÇÃO
   - DE PAGAMENTO, COM ACESSO INDIRETO AO DICT 8.5 FLUXO DE CONSULTA PARA O PARTICIPANTE DO PIX QUE ATUA COMO PRESTADOR DE SERVIÇO DE INICIAÇÃO DE TRANSAÇÃO
- 9 FLUXO DE VERIFICAÇÃO DE SINCRONISMO
   - 9.1 VERIFICAÇÃO DE VSYNC (PARTICIPANTE DO PIX COM ACESSO DIRETO AO DICT)
   - 9.2 VERIFICAÇÃO DE VSYNC (PARTICIPANTE DO PIX COM ACESSO INDIRETO AO DICT)
   - 9.3 LISTA DE CIDS
      - 9.3.1 Participante do Pix com acesso direto
      - 9.3.2 Participante do Pix com acesso indireto
- 10 FLUXO DE NOTIFICAÇÃO DE INFRAÇÃO
   - 10.1 NOTIFICAÇÃO DE INFRAÇÃO PARA SOLICITAÇÃO DE DEVOLUÇÃO OU PARA CANCELAMENTO DE DEVOLUÇÃO
      - com acesso direto ao DICT) 10.1.1 Fluxo de notificação de infração para abertura de solicitação de devolução (participantes do Pix
      - com acesso indireto ao DICT) 10.1.2 Fluxo de notificação de infração para abertura de solicitação de devolução (participantes do Pix
      - acesso direto ao DICT 10.1.3 Fluxo de notificação de infração para cancelamento de devolução entre participantes do Pix com
      - acesso indireto ao DICT 10.1.4 Fluxo de notificação de infração para cancelamento de devolução entre participantes do Pix com
   - 10.2 NOTIFICAÇÃO DE INFRAÇÃO PARA MARCAÇÃO DE FRAUDE TRANSACIONAL
      - com acesso direto ao DICT 10.2.1 Fluxo de notificação de infração para marcação de fraude transacional entre participantes do Pix
      - com acesso indireto ao DICT 10.2.2 Fluxo de notificação de infração para marcação de fraude transacional entre participantes do Pix
- 11 INTERFACE DE COMUNICAÇÃO
- 12 CACHE DE CHAVES CONSULTADAS
- 13 MECANISMOS DE PREVENÇÃO A ATAQUES DE LEITURA
   - 13.1 MECANISMOS ADOTADOS PELO DICT
   - 13.2 MECANISMOS QUE DEVEM SER ADOTADOS PELOS PARTICIPANTES DO PIX
      - 13.2.1 Verificação de autenticidade do usuário solicitante da consulta
      - 13.2.2 Estabelecimento de política interna de limitação de consultas
      - 13.2.3 Monitoramento qualitativo e permanente das consultas
      - 13.2.4 Plano de ação para tratamento de casos suspeitos
      - 13.2.5 Restrição dos dados da chave exibidos ao usuário que faz a consulta
- 14 LIMITAÇÃO DE REQUISIÇÕES À API DO DICT
- 15 FLUXO DE VERIFICAÇÃO DE CHAVES PIX REGISTRADAS
   - 15.1 FLUXO DE VERIFICAÇÃO DE CHAVES PIX REGISTRADAS PARA O PARTICIPANTE DO PIX COM ACESSO DIRETO AO DICT
   - 15.2 FLUXO DE VERIFICAÇÃO DE CHAVES PIX REGISTRADAS PARA O PARTICIPANTE DO PIX COM ACESSO INDIRETO AO DICT
- 16 CACHE DE EXISTÊNCIA DE CHAVE PIX
- 17 FLUXO DE SOLICITAÇÃO DE DEVOLUÇÃO
   - 17.1 SOLICITAÇÃO DE DEVOLUÇÃO POR FALHA OPERACIONAL
      - Pix com acesso direto ao DICT) 1.1.1. Fluxo de solicitação de devolução por “falha operacional do PSP do pagador” (participantes do
      - Pix com acesso indireto ao DICT)................................................................................................................... 1.1.2. Fluxo de solicitação de devolução por “falha operacional do PSP do pagador” (participantes do
   - AO DICT) 17.2 FLUXO DE SOLICITAÇÃO DE DEVOLUÇÃO POR “FUNDADA SUSPEITA DE FRAUDE” (PARTICIPANTES DO PIX COM ACESSO DIRETO
   - INDIRETO AO DICT) 17.3 FLUXO DE SOLICITAÇÃO DE DEVOLUÇÃO POR “FUNDADA SUSPEITA DE FRAUDE” (PARTICIPANTES DO PIX COM ACESSO
   - 17.4 FLUXO DE CANCELAMENTO DE DEVOLUÇÃO.....................................................................................................
   - AO PIX AUTOMÁTICO 17.5 FLUXO DE SOLICITAÇÃO DE DEVOLUÇÃO POR ERRO DO PSP DO PAGADOR NO ENVIO DE ORDEM DE PAGAMENTO REFERENTE
      - referente ao Pix Automático (participantes do Pix com acesso direto ao DICT) 1.1.3. Fluxo de solicitação de devolução por erro do PSP do pagador no envio de ordem de pagamento
      - referente ao Pix Automático (participantes do Pix com acesso indireto ao DICT) 1.1.4. Fluxo de solicitação de devolução por erro do PSP do pagador no envio de ordem de pagamento
- 18 FLUXO DE CONSULTA A INFORMAÇÕES DE SEGURANÇA
   - 18.1 FLUXO DE CONSULTA A INFORMAÇÕES DE SEGURANÇA PARA O PARTICIPANTE DO PIX COM ACESSO DIRETO AO DICT
   - 18.2 FLUXO DE CONSULTA A INFORMAÇÕES DE SEGURANÇA PARA O PARTICIPANTE DO PIX COM ACESSO INDIRETO AO DICT
- 19 CONSULTA DE BALDES
- 20 FLUXO DE RECUPERAÇÃO DE VALORES
   - 20.1 REGRAS GERAIS
      - 20.1.1 Instauração no fluxo interativo
      - 20.1.2 Rastreamento no fluxo interativo
      - 20.1.3 Priorização no fluxo interativo
      - 20.1.4 Solicitação de bloqueio no fluxo interativo
      - 20.1.5 Instauração no fluxo automatizado
      - 20.1.6 Etapa de análise
      - 20.1.7 Etapa de devolução
      - 20.1.8 Desbloqueio dos recursos
      - 20.1.9 Recuperação de valores para transações liquidadas nos sistemas dos participantes
   - 20.2 FLUXO DE INSTAURAÇÃO E SOLICITAÇÃO DE BLOQUEIO NO FLUXO INTERATIVO
   - 20.3 FLUXO DE INSTAURAÇÃO E SOLICITAÇÃO DE BLOQUEIO NO FLUXO AUTOMATIZADO
   - 20.4 FLUXO DE ANÁLISE
   - 20.5 FLUXO DE DEVOLUÇÃO
- 21 NOTIFICAÇÕES DE EVENTOS
- 22 HISTÓRICO DE REVISÃO


## 1 CHAVES PIX

As chaves Pix serão armazenadas no DICT no formato indicado no quadro abaixo:

```
Tipo Formato Descrição^
Número de telefone
celular
```
```
+XXXXXXXXXXXXX O telefone celular usará o padrão E.164^1.
```
```
Endereço de e-mail xxxxxxxx@xxxxxxx.xxx(.xx)
```
```
O e-mail terá tamanho máximo de 77
caracteres e será validado de acordo com
expressão regular definida na especificação da
API do DICT.
```
##### CPF XXXXXXXXXXX

```
O CPF será utilizado com 11 números,
incluindo os dígitos verificadores. Deverá ser
informado sem pontos ou traços.
```
##### CNPJ XXXXXXXXXXXXXX

```
O CNPJ será utilizado com 14 números,
incluindo os dígitos verificadores. Deverá ser
informado sem pontos ou traços.
```
```
Chave aleatória^2 XXXXXXXX-XXXX-XXXX-
XXXX-XXXXXXXXXXXX
```
```
UUID gerado pelo DICT, conforme formato
especificado na RFC4122^3.
```
O usuário final com número de inscrição no CPF pode vincular até cinco chaves Pix para cada conta
transacional da qual for titular. O usuário final com número de inscrição no CNPJ pode vincular até vinte
chaves Pix para cada conta transacional da qual for titular. O limite de chaves é aplicado para cada conta
transacional, independentemente da quantidade de titulares da conta.

### 1.1 CHAVES BLOQUEADAS POR ORDEM JUDICIAL

Quando há um pedido de bloqueio para uma chave, por ordem judicial, o DICT retorna a mensagem de
erro de chave bloqueada ( _EntryBlocked_ ), com o código HTTP 400, nas operações de consulta, alteração
de dados, exclusão, portabilidade e reivindicação de posse. Como parte do processo, o participante do
Pix, ao receber o pedido via ofício do Banco Central, deve realizar o mesmo bloqueio em suas bases
internas, para que as consultas a essa chave, em uma transação interna, retornem a informação de
bloqueio, sem a exibição das informações permitidas da chave.

## 2 VALIDAÇÃO DAS CHAVES PIX

### 2.1 VALIDAÇÃO DA POSSE DA CHAVE

Para validar o número no CPF ou no CNPJ, o participante do Pix deve, pelo menos, verificar se o número
informado corresponde ao número utilizado para a abertura da conta.

(^1) https://www.itu.int/rec/T-REC-E.164- 201011 - I/en.
(^2) A chave aleatória é uma sequência alfanumérica que não possui qualquer significado, a não ser o de servir como
uma chave Pix. A chave aleatória não é uma chave de determinação livre. Trata-se de uma chave constituída por
uma sequência de _bits_ gerados aleatoriamente pelo DICT.
(^3) https://tools.ietf.org/html/rfc4122#section-3.


Para validar o número de telefone celular ou o endereço de e-mail, o participante do Pix deve, pelo
menos, enviar um código para o número de telefone celular ou para o endereço de e-mail informado e
solicitar a inclusão desse código em algum canal de atendimento disponibilizado por ele, o qual deve ser
confirmado por meio de algum mecanismo de autenticação digital, de livre escolha do participante do
Pix.

### 2.2 VALIDAÇÃO DA SITUAÇÃO CADASTRAL DO USUÁRIO NA RECEITA FEDERAL

O participante do Pix deve observar as situações cadastrais de seus usuários finais que constam do
Cadastro de Pessoas Físicas, CPF, e do Cadastro Nacional da Pessoa Jurídica, CNPJ, conforme o caso, para
atendimento dos critérios de registro, alteração, portabilidade, reivindicação de posse e exclusão de
chaves, conforme disposto no Regulamento do Pix.

No caso de pessoas naturais, são consideradas **irregulares** no DICT as situações cadastrais:

- suspensa;
- cancelada;
- titular falecido;
- nula.

No caso de pessoas jurídicas, são consideradas **irregulares** as situações:

- suspensa;
- inapta, exceto na hipótese prevista no inciso I do art. 38 da Instrução Normativa RFB nº 2.119,
    de 6 de dezembro de 2022;
- baixada;
- nula.

Nos casos de CNPJ atribuído a Microempreendedor Individual (MEI), a situação “suspensa” não
configurará irregularidade quando for decorrente da inobservância do disposto no art. 1º da Resolução
nº 36, de 2 de maio de 2016, do Comitê para Gestão da Rede Nacional para Simplificação do Registro e
da Legalização de Empresas e Negócios - CGSIM.

### 2.3 VALIDAÇÃO DOS NOMES VINCULADOS ÀS CHAVES PIX

As informações vinculadas às chaves Pix armazenadas no DICT devem estar conforme o Cadastro de
Pessoas Físicas (CPF) ou o Cadastro Nacional de Pessoas Jurídicas (CNPJ) da Receita Federal, dependendo
do tipo de pessoa proprietária da chave. Para essa conformidade, devem ser observadas as regras dessa
seção.

Para pessoas físicas, deverá ser usado o nome civil no campo _Name_ , podendo o nome social ser utilizado
somente se registrado no CPF. Para pessoa jurídica, deverá ser usada a razão social no campo _Name_ , e
o nome fantasia poderá ser preenchido no campo _TradeName,_ mas somente quando este constar na
inscrição da empresa no CNPJ. No caso de Microempreendedores Individuais, MEI, não é permitido o
preenchimento do campo _TradeName_ , que deve ficar vazio.

Tanto para pessoas físicas quanto jurídicas, todas as palavras que compõem o nome do usuário no CPF,
ou no CNPJ, devem ser escritas no campo adequado da chave. Além disso, os nomes vinculados à chave


não podem conter palavras que não existam no nome original. Cada palavra deve ser grafada de forma
idêntica à constante no registro da Receita Federal, ressalvadas as seguintes diferenças:

- Permite-se o uso de diacríticos nas palavras dos nomes vinculados à chave Pix, ainda que as
    bases da Receita Federal não os utilizem, porém, restritos à lista: Ã, Õ, Á, É, Í, Ó, Ú, À, È, Ì, Ò, Ù,
    Â, Ê, Î, Ô, Û, Ä, Ë, Ï, Ö, Ü, Ç, Ñ, Å;
- A troca dos seguintes caracteres por espaço, ou vice-versa: ponto (.), vírgula (,), hífen (-),
    apóstrofo (');
- A troca do símbolo &, “e comercial” ou ampersand, pela letra E, ou vice-versa;
- Não é necessário diferenciar as letras maiúsculas das minúsculas.

No caso de pessoas físicas, é permitido abreviar palavras do nome, seja nome civil ou nome social, desde
que seguidas as regras abaixo:

- Uma abreviação é a troca de uma palavra por sua primeira letra. Não é permitido trocar uma
    palavra por duas ou mais letras, como ocorreria em FCO para FRANCISCO, ou JR para JUNIOR;
- O primeiro nome, tomado como a primeira palavra do nome registrado no CPF, deve aparecer
    por extenso, ainda que seja parte de um nome composto e que os outros nomes permaneçam
    por extenso;
- O último sobrenome, tomado como a última palavra do nome registrado no CPF, deve
    aparecer por extenso, ainda que seja um agnome como “Filho”, “Junior”, “Sobrinho”,
    “Primeiro”, “Segundo” etc.;
- Ao abreviar o nome do usuário, não é permitida a omissão de nenhuma palavra, nem mesmo
    preposições como “da”, “de”, ou “do”, por exemplo.

No caso de pessoas jurídicas, a razão social deve ser escrita por extenso, assim como o nome fantasia,
quando este constar da inscrição no CNPJ. Isso inclui os termos indicativos de sua natureza jurídica,
como LTDA, S/A, ME, EIRELI, etc. Todas as palavras devem ser grafadas por extenso, sem nenhuma
abreviação, e exatamente como constam no CNPJ, ressalvadas as diferenças admitidas já listadas e a
troca da barra (/) por espaço, ou vice-versa.

Tanto para pessoas físicas quanto jurídicas, as palavras, abreviadas ou não, devem ser separadas por um
único espaço, mesmo nas situações em que tenham sido usadas as regras sobre troca de caracteres por
espaço desta seção. Por exemplo, ao se omitir um ponto que é seguido de espaço, não é necessário
escrever dois espaços. Para uma empresa hipotética com razão social “M. Almeida Reparos M. E.”, ela
poderia ser escrita como “M Almeida Reparos M E”, com apenas um espaço entre cada palavra. Também
não é necessário escrever espaços após o final da última palavra.

**Exemplos para auxílio na interpretação das regras:**

Diferenças permitidas de grafia:

- Maria, MARIA, maria
- João, Joao
- Conceição, CONCEICAO
- D’Alembert, D Alembert
- Sant’Anna, Sant Anna
- Ben-Hur, Ben Hur
- LAVA-RÁPIDO, LAVA RAPIDO


- Comes & Bebes, Comes e Bebes
- Empresa ABCD S/A, Empresa ABCD S A, Empresa ABCD S.A.

Os pares abaixo são considerados diferentes, ou não conformes, segundo as regras dessa seção:

- Sousa, Souza
- Luis, Luiz
- Sant’Anna, Santanna
- Empresa S/A, Empresa SA
- Empresa S.A., Empresa as

Seguem alguns exemplos de abreviação de nomes de pessoa física. Partindo do nome Fulano Beltrano
Sicrano de Tal, são aceitáveis Fulano B. S. d. Tal, Fulano Beltrano S. de Tal, Fulano B. Sicrano de Tal, e
não são aceitáveis F. Beltrano S. de Tal, Fulano Beltrano de Tal, Fulano Beltrano Sicrano Tal. Para nomes
compostos, caso a pessoa se chame, por exemplo, Maria Cecília, o nome “Maria” não poderá ser
abreviado, ainda que se pretenda escrever “M. Cecília”. O uso de ponto (.) para denotar as abreviações
é facultativo, pela forma como são tratados.

## 3 FLUXO DE REGISTRO DE CHAVE

### 3.1 FLUXO DE REGISTRO DE CHAVE POR SOLICITAÇÃO DO USUÁRIO FINAL (PARTICIPANTES DO PIX COM ACESSO DIRETO AO DICT)

```
# Camada Tipo Descrição
1 Usuário final Ação Usuário final acessa seu canal de atendimento.
```

```
2 Usuário final Ação
```
```
Usuário final informa qual chave deseja cadastrar entre os
cinco tipos possíveis: CPF, CNPJ, e-mail, número de
telefone celular e chave aleatória. Usuário final solicita
registro da chave.
```
```
3 Usuário final Comunicação
```
```
Usuário final encaminha sua solicitação e os dados
necessários a seu prestador de serviços de pagamento
(PSP).
```
```
4
```
```
PSP do usuário
final
```
```
Comunicação
```
```
PSP do usuário final recebe a solicitação de registro de
chave no DICT.
```
```
5 PSP do usuário
final
```
```
Ação
```
```
PSP faz a validação da posse da chave pelo usuário final e
confirma seus dados e situação cadastral na Receita
Federal Antes de encaminhar a solicitação de registro ao
DICT, o PSP deve verificar se a chave já está registrada em
sua base interna. Caso esteja, não é necessário o envio de
solicitação ao DICT. PSP segue diretamente para a etapa 15.
```
```
6
```
```
PSP do usuário
final Mensagem^
```
```
Solicitação de registro de chave é encaminhada ao DICT
através da mensagem “Diretório / Criar Vínculo”.
```
```
7 DICT Mensagem
```
```
DICT recebe mensagem com a solicitação de registro de
chave.
```
```
8 DICT Ação
```
```
DICT realiza verificações de conformidade:
i) instituição que solicitou o registro, ou seja, o PSP com
acesso direto, deve possuir autorização para realizar o
registro de chaves Pix; e
ii) PSP vinculado à chave deve ser o mesmo PSP do usuário
final que solicitou o registro da chave.
```
```
9 DICT Ação
```
```
DICT verifica se a chave Pix solicitada já está registrada.
No caso da chave aleatória, DICT gera o número
hexadecimal que servirá como endereço.
```
10 DICT Ação

```
Caso a chave solicitada não esteja registrada no DICT,
efetua-se seu cadastro, vinculando a chave aos dados da
conta transacional.
```
11 DICT Ação

```
Caso a chave solicitada já esteja registrada no DICT, é criada
mensagem indicando a existência de registro para aquela
chave, com os dados da chave existente.
```
12 DICT Mensagem

```
DICT envia mensagem de resposta ao PSP do usuário final,
informando o sucesso do registro ou a existência de chave
já registrada.
```
13

```
PSP do usuário
final
```
```
Mensagem PSP do usuário final recebe mensagem de resposta do DICT.
```
14 PSP do usuário
final

```
Ação
```
```
No caso de sucesso no registro, o PSP do usuário final
atualiza sua base de dados interna com a chave e os dados
da conta vinculados a ela.
```
##### 15

```
PSP do usuário
final
```
```
Comunicação
```
```
PSP do usuário final pode encaminhar três comunicações
para o usuário final:
i) caso a chave tenha sido registrada com sucesso, é
enviada comunicação confirmando o registro da
chave;
```

```
ii) caso a chave não tenha sido registrada e a chave
esteja de posse do mesmo CPF/CNPJ que solicitou
seu registro, é enviada comunicação indicando o
PSP da chave já registrada e perguntando se o
usuário deseja efetuar portabilidade da chave; ou
iii) caso a chave não tenha sido registrada e a chave
esteja de posse de CPF/CNPJ diferente daquele que
solicitou seu registro, é enviada comunicação
indicando o PSP da chave já registrada e
perguntando se o usuário deseja efetuar
reivindicação de posse da chave.
```
16 Usuário final Comunicação

```
Usuário final recebe comunicação do seu PSP. Caso a chave
tenha sido registrada com sucesso, o fluxo é encerrado.
```
17 Usuário final Comunicação Usuário final envia comunicação ao seu PSP, caso o usuário
decida solicitar portabilidade ou reivindicação de posse.

##### 18

```
PSP do usuário
final Comunicação^
```
```
PSP do usuário final recebe comunicação do usuário.
Caso o usuário escolha solicitar a portabilidade da chave, o
fluxo de portabilidade é iniciado.
Caso o usuário escolha solicitar a reivindicação da chave, o
fluxo de reivindicação de posse é iniciado.
```
### DICT) 3.2 FLUXO DE REGISTRO DE CHAVE POR SOLICITAÇÃO DO USUÁRIO FINAL (PARTICIPANTES DO PIX COM ACESSO INDIRETO AO

### com acesso indireto ao DICT)


**# Camada Tipo Descrição**
1 Usuário final Ação Usuário final acessa seu canal de atendimento.

2 Usuário final Ação

```
Usuário final informa qual chave deseja cadastrar entre os
cinco tipos possíveis: CPF, CNPJ, e-mail, número de
telefone celular e chave aleatória. Usuário final solicita
registro da chave.
```
3 Usuário final Comunicação

```
Usuário final encaminha sua solicitação e os dados
necessários a seu prestador de serviços de pagamento
(PSP).
```
4

```
PSP do usuário
final
```
```
Comunicação
```
```
PSP do usuário final recebe a solicitação de registro de
chave no DICT.
```
##### 5

```
PSP do usuário
final
```
```
Ação
```
```
PSP faz a validação da posse da chave pelo usuário final e
confirma seus dados e situação cadastral na Receita
Federal. Antes de encaminhar a solicitação de registro ao
DICT, o PSP deve verificar se a chave já está registrada em
sua base interna. Caso esteja, não é necessário o envio de
solicitação ao DICT. PSP segue diretamente para a etapa 19.
```
6 PSP do usuário
final

```
Comunicação Solicitação de registro de chave é encaminhada a PSP com
acesso direto ao DICT.
```

##### 7

```
PSP com acesso
direto ao DICT
```
```
Comunicação
```
```
PSP com acesso direto ao DICT recebe comunicação com a
solicitação de registro de chave.
```
##### 8

```
PSP com acesso
direto ao DICT
```
```
Mensagem
```
```
PSP com acesso direto encaminha a solicitação de registro
de chave ao DICT através da mensagem “Diretório / Criar
Vínculo”.
```
```
9 DICT Mensagem DICT recebe mensagem com a solicitação de registro de
chave.
```
10 DICT Ação

```
DICT realiza verificações de conformidade:
i) instituição que solicitou o registro, ou seja, o PSP com
acesso direto, deve possuir autorização para realizar o
registro de chaves Pix. O DICT deve, inclusive, verificar se o
PSP com acesso direto ao DICT tem autorização para
registrar chaves para o PSP sem acesso direto;
ii) PSP vinculado à chave deve ser o mesmo PSP do usuário
final que solicitou o registro da chave; e
iii) PSP com acesso direto ao DICT tem permissão para
efetuar registros em nome do PSP do usuário.
```
11 DICT Ação

```
DICT verifica se a chave Pix solicitada já está registrada.
No caso da chave aleatória, DICT gera o número
hexadecimal que servirá como endereço.
```
12 DICT Ação

```
Caso a chave solicitada já esteja registrada no DICT, é criada
mensagem indicando a existência de registro para aquela
chave, com os dados da chave existente.
```
13 DICT Ação

```
Caso a chave solicitada já esteja registrada no DICT, é criada
mensagem indicando registro para aquela chave, com os
dados da chave existente.
```
14 DICT Mensagem

```
DICT envia mensagem de resposta ao PSP com acesso
direto, informando o sucesso do registro ou a existência de
chave já registrada.
```
15

```
PSP com acesso
direto ao DICT Mensagem^
```
```
PSP com acesso direto ao DICT recebe mensagem de
resposta.
```
##### 16

```
PSP com acesso
direto ao DICT Comunicação^
```
```
PSP com acesso direto ao DICT envia comunicação ao PSP
do usuário final, informando o registro ou a existência de
chave já registrada.
```
17

```
PSP do usuário
final Comunicação^ PSP do usuário final recebe comunicação.^
```
##### 18

```
PSP do usuário
final
```
```
Ação
```
```
No caso de sucesso no registro, o PSP do usuário final
atualiza sua base de dados interna com a chave e os dados
da conta vinculados a ela.
```
19 PSP do usuário
final

```
Comunicação
```
```
PSP do usuário final pode encaminhar três comunicações
para o usuário final:
i) caso a chave tenha sido registrada com sucesso, é
enviada comunicação confirmando o registro da
chave;
ii) caso a chave não tenha sido registrada e a chave
esteja de posse do mesmo CPF/CNPJ que solicitou
seu registro, é enviada comunicação indicando o
```

```
PSP da chave já registrada e perguntando se o
usuário deseja efetuar portabilidade da chave; ou
iii) caso a chave não tenha sido registrada e a chave
esteja de posse de CPF/CNPJ diferente daquele que
solicitou seu registro, é enviada comunicação
indicando o PSP da chave já registrada e
perguntando se o usuário deseja efetuar
reivindicação de posse da chave.
```
20 Usuário final Comunicação

```
Usuário final recebe comunicação do seu PSP. Caso a chave
tenha sido registrada com sucesso, o fluxo é encerrado.
```
21 Usuário final Comunicação

```
Usuário final envia comunicação ao seu PSP, caso o usuário
decida solicitar portabilidade ou reivindicação de posse.
```
##### 22

```
PSP do usuário
final
```
```
Comunicação
```
```
PSP do usuário final recebe comunicação do usuário.
Caso o usuário escolha solicitar a portabilidade da chave, o
fluxo de portabilidade é iniciado.
Caso o usuário escolha solicitar a reivindicação da chave, o
fluxo de reivindicação de posse é iniciado.
```

### 3.3 FLUXO DE REGISTRO DE CHAVE INICIADO PELO PARTICIPANTE (PARTICIPANTES DO PIX COM ACESSO DIRETO AO DICT)

```
# Camada Tipo Descrição
```
##### 1

```
PSP do usuário
final
```
```
Ação
```
```
Após processo de verificação de sincronismo, PSP identifica
chave registrada corretamente em sua base, mas que, por
falha operacional, não está registrada no DICT.
```
```
2
```
```
PSP do usuário
final
```
```
Mensagem
```
```
Solicitação de registro de chave é encaminhada ao DICT
através da mensagem “Diretório / Criar Vínculo”.
```
```
3 DICT Mensagem DICT recebe mensagem com a solicitação de registro^ de
chave.
```
```
4 DICT Ação
```
```
DICT realiza verificações de conformidade:
i) instituição que solicitou o registro, ou seja, o PSP com
acesso direto, deve possuir autorização para realizar o
registro de chaves Pix; e
ii) PSP vinculado à chave deve ser o mesmo PSP do usuário
final que solicitou o registro da chave.
```
```
5 DICT Ação
```
```
DICT verifica se a chave de endereçamento solicitada já
está registrada.
```
```
6 DICT Ação
```
```
Caso a chave solicitada não esteja registrada no DICT,
efetua-se seu cadastro, vinculando a chave aos dados da
conta transacional.
```

```
7 DICT Ação
```
```
Caso a chave solicitada já esteja registrada no DICT, é criada
mensagem indicando a existência de registro para aquela
chave, com os dados da chave existente.
```
```
8 DICT Mensagem
```
```
DICT envia mensagem de resposta ao PSP do usuário final,
informando o sucesso do registro ou a existência de chave
já registrada.
```
```
9
```
```
PSP do usuário
final Mensagem^ PSP do usuário final recebe mensagem de resposta do DICT.^
```
### 3.4 FLUXO DE REGISTRO DE CHAVE INICIADO PELO PARTICIPANTE (PARTICIPANTES DO PIX COM ACESSO INDIRETO AO DICT)

```
# Camada Tipo Descrição
```
##### 1

```
PSP do usuário
final Ação^
```
```
Após processo de verificação de sincronismo, PSP identifica
chave registrada corretamente em sua base, mas que, por
falha operacional, não está registrada no DICT.
```
```
2
```
```
PSP do usuário
final Comunicação^
```
```
Solicitação de registro de chave no DICT é encaminhada ao
PSP com acesso direto ao DICT.
```

##### 3

```
PSP com acesso
direto ao DICT
```
```
Comunicação
```
```
PSP com acesso direto ao DICT recebe comunicação com a
solicitação de registro de chave.
```
```
4 PSP com acesso
direto ao DICT
```
```
Mensagem
```
```
PSP com acesso direto encaminha solicitação de registro de
chave ao DICT através da mensagem “Diretório / Criar
Vínculo”.
```
```
5 DICT Mensagem DICT recebe mensagem com a solicitação de registro^ de
chave.
```
```
6 DICT Ação
```
```
DICT realiza verificações de conformidade:
i) instituição que solicitou o registro, ou seja, o PSP com
acesso direto, deve possuir autorização para realizar o
registro de chaves Pix;
ii) PSP vinculado à chave deve ser o mesmo PSP do usuário
final que solicitou o registro da chave; e
iii) PSP com acesso direto ao DICT tem permissão para
efetuar registros em nome do PSP do usuário.
```
```
7 DICT Ação
```
```
DICT verifica se a chave de endereçamento solicitada já
está registrada.
```
```
8 DICT Ação
```
```
Caso a chave solicitada não esteja registrada no DICT,
efetua-se seu cadastro, vinculando a chave aos dados da
conta transacional.
```
```
9 DICT Ação
```
```
Caso a chave solicitada já esteja registrada no DICT, é criada
mensagem indicando a existência de registro para aquela
chave, com os dados da chave existente.
```
10 DICT Mensagem

```
DICT envia mensagem de resposta ao PSP com acesso
direto, informando o sucesso do registro ou a existência de
chave já registrada.
```
11

```
PSP com acesso
direto ao DICT
```
```
Mensagem
```
```
PSP com acesso direto ao DICT recebe mensagem de
resposta.
```
##### 12

```
PSP com acesso
direto ao DICT
```
```
Comunicação
```
```
PSP com acesso direto ao DICT envia comunicação ao PSP
do usuário final, informando o registro ou a existência de
chave já registrada.
```
13

```
PSP do usuário
final
```
```
Comunicação PSP do usuário final recebe comunicação.
```

## 4 FLUXO DE EXCLUSÃO DE CHAVE

### 4.1 EXCLUSÃO DE CHAVE POR INCOMPATIBILIDADE DE DADOS COM A RECEITA FEDERAL

No caso de exclusão de chave por incompatibilidade de dados com o cadastro da Receita Federal, como
desconformidade de nome ou situação irregular de CPF ou CNPJ, o PSP deverá utilizar os seguintes
códigos no campo “ _Reason”_ do _endpoint_ “Remover Vínculo”:

- “FRAUD”: caso o PSP identifique que a divergência de nome ou a irregularidade do CPF/CNPJ
    está relacionada com a prática de fraude;
- “RFB_VALIDATION”: nos demais casos.

O usuário deverá ser comunicado sobre a exclusão da chave imediatamente após sua efetivação,
conforme fluxos das seções 4.4 e 4.5 deste manual, informando o motivo da exclusão. Por exemplo, em
caso de incompatibilidade de nome, o participante deverá dar detalhes que permitam ao usuário
identificar o problema, ou em caso de situação irregular, especificar a situação que causou a exclusão.

### DICT) 4.2 FLUXO DE EXCLUSÃO DE CHAVE POR SOLICITAÇÃO DO USUÁRIO FINAL (PARTICIPANTES DO PIX COM ACESSO DIRETO AO

### com acesso direto ao DICT)

```
# Camada Tipo Descrição
1 Usuário final Ação Usuário final acessa seu canal de atendimento.
2 Usuário final Ação Usuário final solicita a exclusão de chave Pix.
3 Usuário final Comunicação Usuário final encaminha sua solicitação a seu PSP.
```

##### 4

```
PSP do usuário
final
```
```
Comunicação
```
```
PSP do usuário final recebe a solicitação de exclusão de
chave no DICT.
```
```
5 PSP do usuário
final
```
```
Ação
```
```
PSP verifica se a chave está registrada em sua base interna e
se o usuário que está requisitando a exclusão é o mesmo
usuário que está vinculado à chave.
```
```
6
```
```
PSP do usuário
final
```
```
Mensagem
```
```
Solicitação de exclusão de chave é encaminhada ao DICT
através da mensagem “Diretório / Remover Vínculo”.
```
```
7 DICT Mensagem DICT recebe mensagem com a solicitação de exclusão de
chave.
```
```
8 DICT Ação
```
```
DICT realiza verificações de conformidade:
i) chave está registrada no DICT;
ii) instituição que solicitou a exclusão deve ser a mesma que
efetuou o registro; e
iii) chave deve pertencer ao usuário que solicitou a exclusão.
9 DICT Ação DICT exclui a chave Pix de seu banco de dados.
```
10 DICT Mensagem

```
DICT envia mensagem de confirmação da exclusão ao PSP do
usuário final.
```
11

```
PSP do usuário
final Mensagem^
```
```
PSP do usuário final recebe comunicação informando a
exclusão da chave.
```
12

```
PSP do usuário
final
```
```
Ação
```
```
PSP do usuário final atualiza sua base de dados interna,
excluindo a chave.
```
13 PSP do usuário
final

```
Comunicação PSP do usuário final envia confirmação de exclusão da chave.
```
14 Usuário final Comunicação

```
Usuário final recebe confirmação de exclusão da chave
solicitada.
```
### DICT) 4.3 FLUXO DE EXCLUSÃO DE CHAVE POR SOLICITAÇÃO DO USUÁRIO FINAL (PARTICIPANTES DO PIX COM ACESSO INDIRETO AO

### com acesso indireto ao DICT)


**# Camada Tipo Descrição**
1 Usuário final Ação Usuário final acessa seu canal de atendimento.
2 Usuário final Ação Usuário final solicita a exclusão de chave Pix.
3 Usuário final Comunicação Usuário final encaminha sua solicitação a seu PSP.

4 PSP do usuário final Comunicação

```
PSP do usuário final recebe a solicitação de exclusão de
chave no DICT.
```
5 PSP do usuário final Ação

```
PSP verifica se a chave está registrada em sua base interna
e se o usuário que está requisitando a exclusão é o mesmo
usuário que está vinculado à chave.
```
6 PSP do usuário final Comunicação

```
Solicitação de exclusão de chave é encaminhada ao PSP
com acesso direto ao DICT.
```
7 PSP com acesso
direto ao DICT

```
Comunicação PSP com acesso direto ao DICT recebe comunicação com a
solicitação de exclusão de chave.
```
8 PSP com acesso
direto ao DICT

```
Mensagem
```
```
PSP com acesso direto encaminha solicitação de exclusão
de chave através da mensagem “Diretório / Remover
Vínculo”.
```
9 DICT Mensagem

```
DICT recebe mensagem com a solicitação de exclusão de
chave.
```

```
10 DICT Ação
```
```
DICT realiza verificações de conformidade:
i) chave está registrada no DICT;
ii) instituição que solicitou a exclusão deve ser a mesma que
efetuou o registro; e
iii) chave deve pertencer ao usuário que solicitou a
exclusão.
11 DICT Ação DICT exclui a chave Pix de seu banco de dados.
```
```
12 DICT Mensagem
```
```
DICT envia mensagem de confirmação da exclusão ao PSP
com acesso direto.
```
```
13
```
```
PSP com acesso
direto ao DICT Mensagem^
```
```
PSP com acesso direto ao DICT recebe resposta da exclusão
da chave.
```
```
14
```
```
PSP com acesso
direto ao DICT
```
```
Comunicação
```
```
PSP com acesso direto ao DICT encaminha comunicação ao
PSP do usuário final, informando a exclusão da chave.
```
```
15 PSP do usuário final Comunicação PSP do usuário final recebe comunicação informando a
exclusão da chave.
```
```
16 PSP do usuário final Ação
```
```
PSP do usuário final atualiza sua base de dados interna,
excluindo a chave.
```
```
17 PSP do usuário final Comunicação
```
```
PSP do usuário final envia confirmação de exclusão da
chave.
```
```
18 Usuário final Comunicação Usuário final recebe confirmação de exclusão da chave
solicitada.
```
### 4.4 FLUXO DE EXCLUSÃO DE CHAVE INICIADO PELO PARTICIPANTE (PARTICIPANTES DO PIX COM ACESSO DIRETO AO DICT)


```
# Camada Tipo Descrição
```
```
1
```
```
PSP do usuário
final Ação^ Inicia processo de exclusão.^
```
```
2
```
```
PSP do usuário
final
```
```
Mensagem
```
```
Solicitação de exclusão de chave é enviada ao DICT através
da mensagem “Diretório / Remover Vínculo”.
```
```
3 DICT Mensagem
```
```
DICT recebe mensagem com a solicitação de exclusão de
chave.
```
```
4 DICT Ação
```
```
DICT realiza verificações de conformidade:
i) chave está registrada no DICT; e
ii) instituição que solicitou a exclusão deve ser a mesma que
efetuou o registro.
5 DICT Ação DICT exclui a chave Pix de seu banco de dados.
```
```
6 DICT Mensagem
```
```
DICT envia mensagem de confirmação da exclusão ao PSP do
usuário final.
```
```
7
```
```
PSP do usuário
final
```
```
Mensagem
```
```
PSP do usuário final recebe comunicação informando a
exclusão da chave.
```
```
8 PSP do usuário
final
```
```
Ação PSP do usuário final atualiza sua base de dados interna,
excluindo a chave.
```
```
9
```
PSP do usuário
final Comunicação^ PSP do usuário final envia confirmação de exclusão da chave.^
10 Usuário final Comunicação Usuário final recebe confirmação de exclusão da chave.


### 4.5 FLUXO DE EXCLUSÃO DE CHAVE INICIADO PELO PARTICIPANTE (PARTICIPANTES DO PIX COM ACESSO INDIRETO AO DICT)..

```
# Camada Tipo Descrição
1 PSP do usuário final Ação Inicia processo de exclusão.
```
```
2 PSP do usuário final Comunicação
```
```
Solicitação de exclusão de chave é encaminhada ao PSP
com acesso direto ao DICT.
```
```
3
```
```
PSP com acesso
direto ao DICT
```
```
Comunicação
```
```
PSP com acesso direto ao DICT recebe comunicação com a
solicitação de exclusão de chave.
```
##### 4

```
PSP com acesso
direto ao DICT
```
```
Mensagem
```
```
PSP com acesso direto encaminha solicitação de exclusão
de chave através da mensagem “Diretório / Remover
Vínculo”.
```
```
5 DICT Mensagem DICT recebe mensagem com a solicitação de exclusão de
chave.
```

```
6 DICT Ação
```
```
DICT realiza verificações de conformidade:
i) chave está registrada no DICT; e
ii) instituição que solicitou a exclusão deve ser a mesma que
efetuou o registro.
7 DICT Ação DICT exclui a chave Pix de seu banco de dados.
```
```
8 DICT Mensagem DICT envia mensagem de confirmação da exclusão ao PSP
com acesso direto.
```
```
9
```
```
PSP com acesso
direto ao DICT Mensagem^
```
```
PSP com acesso direto ao DICT recebe resposta da exclusão
da chave.
```
```
10
```
```
PSP com acesso
direto ao DICT
```
```
Comunicação
```
```
PSP com acesso direto ao DICT encaminha comunicação ao
PSP do usuário final, informando a exclusão da chave.
```
```
11 PSP do usuário final Comunicação PSP do usuário final recebe comunicação informando a
exclusão da chave.
```
```
12 PSP do usuário final Ação
```
```
PSP do usuário final atualiza sua base de dados interna,
excluindo a chave.
```
```
13 PSP do usuário final Comunicação
```
```
PSP do usuário final envia confirmação de exclusão da
chave.
14 Usuário final Comunicação Usuário final recebe confirmação de exclusão da chave.
```
## 5 FLUXO DE PORTABILIDADE DE CHAVE

No âmbito do processo de portabilidade, PSP doador é o participante do Pix ao qual a chave está
originalmente vinculada.

Durante o processo de portabilidade, a partir do momento em que o DICT coloca o status do pedido em
“Aberto” até o momento em que o DICT altera o status para “Completo”, a chave Pix não está passível
de pedidos de registro e de exclusão^4. Enquanto o status do pedido estiver como “Aguardando
Resolução”, as consultas continuarão retornando normalmente a identificação da conta transacional
originalmente vinculada à chave. Também durante esse período, tanto o PSP reivindicador quanto o PSP
doador poderão cancelar o pedido, em caso de suspeita de fraude ou em caso de solicitação de
cancelamento pelo usuário do PSP reivindicador. O pedido poderá ainda ser cancelado pelo PSP
reivindicador caso seu status esteja como “Aberto”.

O período de resolução de um processo de portabilidade é de sete dias. Caso o usuário doador não
cancele ou confirme a portabilidade até o final desse período, o PSP doador deve necessariamente
cancelar o pedido. O cancelamento do processo deve ser feito imediatamente após o fim do período de
resolução. O processo continuará com o status de “Aguardando Resolução” até o recebimento dessa
informação pelo DICT.

O PSP reivindicador também pode cancelar uma portabilidade que esteja com status “Confirmado”. Essa
possibilidade é necessária em caso de algum erro no processo de portabilidade.

O DICT disponibiliza um serviço que permite consultar e filtrar os processos de portabilidade. PSPs com
acesso direto ao DICT deverão consultar pelo menos uma vez por minuto esse serviço a fim de identificar

(^4) É permitido ao PSP doador atualizar dados da conta vinculados à chave enquanto o status da requisição estiver
“Aberto” ou “Aguardando Resolução”.


mudanças de status das portabilidades de interesse do PSP (além daquelas de interesse dos PSPs com
acesso indireto para os quais ele presta serviço), tanto como reivindicador quanto como doador.

### 5.1 FLUXO DE PORTABILIDADE PARA O PSP REIVINDICADOR COM ACESSO DIRETO AO DICT

```
# Camada Tipo Descrição
```
```
1 PSP reivindicador Ação
```
```
O processo é iniciado:
(i) após a detecção de que já existe registro para a chave
solicitada e depois do pedido do usuário; ou
(ii) diretamente pelo usuário final, a partir de
funcionalidade específica ofertada por seu PSP (nesse caso,
o usuário deve passar por processo de validação ativa da
chave).
```
```
Antes de enviar o pedido ao DICT, o PSP deve confirmar os
dados e situação cadastral do usuário final na Receita
Federal. Caso a situação seja irregular ou haja divergência,
o PSP não deve iniciar o processo e ir direto para a etapa
13.
```
```
2 PSP reivindicador Mensagem
```
```
O PSP reivindicador envia pedido de portabilidade ao DICT
através da mensagem “Reinvindicação / Criar
Reivindicação” indicando o tipo “Portabilidade”.
```
```
3 DICT Mensagem DICT recebe mensagem com a solicitação de portabilidade
de chave.
```

```
4 DICT Ação
```
```
DICT cria um pedido de portabilidade com status “Aberto”,
com os dados da chave e do PSP reivindicador, e inicia a
contagem do período de resolução, aguardando resposta
do PSP doador.
```
```
5 DICT Ação
```
```
Uma vez recebida, do PSP doador, a confirmação ou o
cancelamento da portabilidade, DICT atualiza o status do
pedido para "Confirmado" ou "Cancelado", conforme o
caso.
Caso o status seja "Cancelado", o DICT altera o status do
pedido para “Completo”, finalizando o processo.
Caso o status seja “Confirmado”, o DICT bloqueia a chave
até o recebimento da confirmação pelo PSP reivindicador.
O bloqueio significa que as consultas à essa chave no DICT
retornarão mensagem de erro.
```
```
6 PSP reivindicador Ação
```
```
PSP reivindicador consulta periodicamente o DICT até
identificar mudança no status da solicitação de
portabilidade, através da mensagem “Reivindicação / Listar
Reivindicações” ou “Reivindicação / Consultar
Reivindicação”.
Ao identificar mudança do status para "Cancelado", PSP
reivindicador prossegue diretamente para a etapa 12.
Ao identificar mudança do status para "Confirmado", PSP
reivindicador prossegue para a etapa 6.
```
```
7 PSP reivindicador Mensagem
```
```
PSP reivindicador solicita conclusão da portabilidade da
chave no DICT, através da mensagem “Reivindicação /
Concluir Reivindicação”.
```
```
8 DICT Mensagem DICT recebe a solicitação de conclusão da portabilidade da
chave.
```
```
9 DICT Ação
```
```
DICT atualiza os dados vinculados à chave e altera o status
da solicitação de portabilidade para “Completo”.
```
10 DICT Mensagem

```
DICT envia mensagem de confirmação de atualização dos
dados vinculados à chave ao PSP reivindicador.
```
11 PSP reivindicador Mensagem

```
PSP reivindicador recebe confirmação de atualização dos
dados vinculados à chave.
```
12 PSP reivindicador Ação PSP reivindicador atualiza sua base de dados interna.

13 PSP reivindicador Comunicação

```
PSP reivindicador informa o usuário final sobre o
cancelamento ou a confirmação do pedido de
portabilidade da chave.
```
14 Usuário final Comunicação Usuário final recebe a informação sobre o cancelamento ou
a confirmação do pedido de portabilidade da chave.


### 5.2 FLUXO DE PORTABILIDADE PARA O PSP REIVINDICADOR COM ACESSO INDIRETO AO DICT

```
# Camada Tipo Descrição
```
```
1 PSP reivindicador Ação
```
```
O processo é iniciado:
(i) após a detecção de que já existe registro para a chave
solicitada e depois do pedido do usuário; ou
(ii) diretamente pelo usuário final, a partir de
funcionalidade específica ofertada por seu PSP (nesse caso,
o usuário deve passar por processo de validação ativa da
chave).
```
```
Antes de enviar o pedido ao DICT, o PSP deve confirmar os
dados e situação cadastral do usuário final na Receita
Federal. Caso a situação seja irregular ou haja divergência,
o PSP não deve iniciar o processo e ir direto para a etapa
21.
```

```
2 PSP reivindicador Comunicação
```
```
O PSP reivindicador envia pedido de portabilidade ao PSP
com acesso direto ao DICT.
```
```
3 PSP com acesso
direto ao DICT
```
```
Comunicação PSP com acesso direto ao DICT recebe pedido de
portabilidade.
```
```
4 PSP com acesso
direto ao DICT
```
```
Mensagem
```
```
PSP com acesso direto ao DICT encaminha pedido de
portabilidade através da mensagem “Reivindicação / Criar
Reivindicação” do tipo “Portabilidade”.
```
```
5 DICT Mensagem DICT recebe mensagem com a solicitação de portabilidade
de chave.
```
```
6 DICT Ação
```
```
DICT cria um pedido de portabilidade com status “Aberto”,
com os dados da chave e do PSP reivindicador, e inicia a
contagem do período de resolução, aguardando resposta
do PSP doador.
```
```
7 DICT Ação
```
```
Uma vez recebida, do PSP doador, a confirmação ou o
cancelamento da portabilidade, DICT atualiza o status do
pedido para "Confirmado" ou "Cancelado", conforme o
caso.
Caso o status seja "Cancelado", o DICT altera o status do
pedido para “Completo”, finalizando o processo.
Caso o status seja “Confirmado”, o DICT bloqueia a chave
até o recebimento da confirmação pelo PSP reivindicador.
```
##### 8

```
PSP com acesso
direto ao DICT
```
```
Ação
```
```
PSP com acesso direto consulta periodicamente o DICT até
identificar mudança no status da solicitação de
portabilidade, através da mensagem “Reivindicação / Listar
Reivindicações” ou “Reivindicação / Consultar
Reivindicação”.
```
```
9 PSP com acesso
direto ao DICT
```
```
Comunicação
```
```
PSP com acesso direto envia comunicação ao PSP
reivindicador, identificando se o status do pedido foi
alterado para “Cancelado” ou para “Confirmado”.
```
10 PSP reivindicador Comunicação

```
PSP reivindicador recebe comunicação do PSP com acesso
direto informando mudança no status do pedido.
Caso o status seja "Cancelado", prossegue-se diretamente
para a etapa 21.
```
11 PSP reivindicador Comunicação

```
Caso o status do pedido seja "Confirmado", o PSP
reivindicador solicita conclusão da portabilidade da chave
no DICT.
```
12 PSP com acesso
direto ao DICT

```
Comunicação PSP com acesso direto recebe a solicitação de conclusão da
portabilidade da chave.
```
13 PSP com acesso
direto ao DICT

```
Mensagem
```
```
PSP com acesso direto encaminha solicitação de conclusão
da portabilidade da chave ao DICT através da mensagem
“Reivindicação / Concluir Reivindicação”.
```
14 DICT Mensagem DICT recebe a solicitação de conclusão da portabilidade da
chave.

15 DICT Ação DICT atualiza os dados vinculados à chave e altera o status
da solicitação de portabilidade para “Completo”.


```
16 DICT Mensagem
```
```
DICT envia mensagem de confirmação de atualização dos
dados vinculados à chave ao PSP com acesso direto.
```
```
17 PSP com acesso
direto ao DICT
```
```
Mensagem PSP com acesso direto recebe confirmação de atualização
da chave.
```
```
18
```
```
PSP com acesso
direto ao DICT Comunicação^
```
```
PSP com acesso direto encaminha confirmação de
atualização da chave.
```
```
19 PSP reivindicador Comunicação
```
```
PSP reivindicador recebe confirmação de atualização dos
dados vinculados à chave.
```
```
20 PSP reivindicador Ação PSP reivindicador atualiza sua base de dados interna.
```
```
21 PSP reivindicador Comunicação
```
```
PSP reivindicador informa o usuário final sobre o
cancelamento ou a confirmação do pedido de
portabilidade da chave.
```
```
22 Usuário final Comunicação
```
```
Usuário final recebe a informação sobre o cancelamento ou
a confirmação do pedido de portabilidade da chave.
```
### 5.3 FLUXO DE PORTABILIDADE PARA O PSP DOADOR COM ACESSO DIRETO AO DICT


**# Camada Tipo Descrição**

1 PSP doador Ação

```
PSP doador consulta periodicamente no DICT a lista de
portabilidades com status “Aberto” que se referem a
chaves registradas por ele, através da mensagem
“Reivindicação / Listar Reivindicações”.
```
2 PSP doador

```
Mensagem/
Comunicação
```
```
Ao identificar portabilidade com status “Aberto”, PSP
doador envia mensagem “Reivindicação / Receber
Reivindicação” ao DICT e notifica o usuário final, solicitando
confirmação ou cancelamento da portabilidade.
```
3 DICT Mensagem DICT recebe mensagem do PSP doador.

4 DICT Ação

```
DICT muda status do pedido para “Aguardando Resolução”
e fica aguardando o recebimento de informação sobre o
processo de portabilidade.
```
5 Usuário final Comunicação Usuário final recebe notificação do PSP doador.


```
6 Usuário final Ação
```
```
Usuário final pode cancelar ou confirmar a portabilidade.
O usuário tem até sete dias para isso. Para cancelar a
portabilidade dentro desse período, o usuário deve fazer
validação ativa da chave. Caso o usuário não cancele nem
confirme a portabilidade nesse período, o PSP doador deve
necessariamente cancelar o pedido, sem a necessidade de
envio de resposta do usuário, prosseguindo diretamente
para a etapa 12. Nesse caso, enquanto o PSP doador não
cancelar o pedido, a chave permanecerá com o status
“Aguardando Resolução”, em que ela está bloqueada para
alteração, mas continua ativa para consultas.
```
```
7 Usuário final Comunicação Usuário final envia comunicação ao PSP doador.
```
```
8 PSP doador Comunicação PSP doador recebe comunicação do usuário.
```
```
9 PSP doador Ação
```
```
Caso o usuário responda solicitando o cancelamento da
portabilidade, o PSP avança para a etapa 12.
Caso o usuário responda solicitando a confirmação da
portabilidade, o PSP doador remove a chave de sua base
interna.
```
```
10 PSP doador Comunicação
```
```
Caso a chave tenha sido excluída da base interna, o PSP
doador comunica o usuário final sobre a exclusão da chave.
```
```
11 Usuário final Comunicação Usuário final recebe a informação de exclusão da chave.
```
```
12 PSP doador Mensagem
```
```
PSP doador informa o DICT sobre a conclusão do processo,
informando o cancelamento, através da mensagem
“Reivindicação / Cancelar Reivindicação”, ou a confirmação
da portabilidade, através da mensagem “Reivindicação /
Confirmar Reivindicação”, conforme o caso.
```
```
13 DICT Mensagem
```
```
DICT recebe a informação de cancelamento ou de
confirmação da portabilidade e continua o processo (etapa
4 do fluxo de portabilidade do PSP reivindicador com
acesso direto ao DICT).
```
### 5.4 FLUXO DE PORTABILIDADE PARA O PSP DOADOR COM ACESSO INDIRETO AO DICT


**# Camada Tipo Descrição**

##### 1

```
PSP com acesso
direto ao DICT
```
```
Ação
```
```
PSP com acesso direto consulta periodicamente no DICT a
lista de portabilidades com status “Aberto” que se referem
a chaves registradas pelo PSP doador ao qual ele presta
serviços, através da mensagem “Reivindicação / Listar
Reivindicações”.
```
##### 2

```
PSP com acesso
direto ao DICT
```
```
Mensagem/
Comunicação
```
```
Ao identificar portabilidade com status “Aberto”, PSP com
acesso direto envia mensagem “Reivindicação / Receber
Reivindicação” ao DICT e comunicação ao PSP doador.
```
3 DICT Mensagem DICT recebe mensagem do PSP com acesso direto.


```
4 DICT Ação
```
```
DICT muda status do pedido para “Aguardando Resolução”
e fica aguardando o recebimento de informação sobre o
processo de portabilidade.
```
```
5 PSP doador Comunicação PSP doador recebe comunicação do PSP com acesso direto.
```
```
6 PSP doador Comunicação
```
```
PSP doador notifica o usuário final, solicitando confirmação
ou cancelamento da portabilidade.
```
```
7 Usuário final Comunicação Usuário final recebe notificação do PSP doador.
```
```
8 Usuário final Ação
```
```
Usuário final pode cancelar ou confirmar a portabilidade.
O usuário tem até sete dias para isso. Para cancelar a
portabilidade dentro desse período, o usuário deve fazer
validação ativa da chave. Caso o usuário não cancele nem
confirme a portabilidade nesse período, o PSP doador deve
necessariamente cancelar o pedido, sem a necessidade de
envio de resposta do usuário, prosseguindo diretamente
para a etapa 14. Nesse caso, enquanto o PSP doador não
cancelar o pedido, a chave permanecerá com o status
“Aguardando Resolução”, em que ela está bloqueada para
alteração, mas continua ativa para consultas.
```
```
9 Usuário final Comunicação Usuário final envia comunicação ao PSP doador.
```
10 PSP doador Comunicação PSP doador recebe comunicação do usuário.

11 PSP doador Ação

```
Caso o usuário responda solicitando o cancelamento da
portabilidade, o PSP avança para a etapa 14.
Caso o usuário responda solicitando a confirmação da
portabilidade, o PSP doador remove a chave de sua base
interna.
```
12 PSP doador Comunicação Caso a chave tenha sido excluída da base interna, o PSP
doador comunica o usuário final sobre a exclusão da chave.

13 Usuário final Comunicação Usuário final recebe a informação de exclusão da chave.

14 PSP doador Comunicação

```
PSP doador informa o PSP com acesso direto sobre a
conclusão do processo, informando o cancelamento ou a
confirmação da portabilidade, conforme o caso.
```
15

```
PSP com acesso
direto ao DICT Comunicação^ PSP com acesso direto recebe informação.^
```
##### 16

```
PSP com acesso
direto ao DICT
```
```
Mensagem
```
```
PSP com acesso direto encaminha informação ao DICT
sobre cancelamento ou confirmação, através das
mensagens “Reivindicação / Cancelar Reivindicação” ou
“Reivindicação / Confirmar Reivindicação”,
respectivamente, conforme o caso.
```

```
17 DICT Mensagem
```
```
DICT recebe a informação de cancelamento ou de
confirmação da portabilidade e continua o processo (etapa
6 do fluxo de portabilidade do PSP reivindicador com
acesso indireto ao DICT).
```
## 6 FLUXO DE REIVINDICAÇÃO DE POSSE DE CHAVE

Assim como no processo de portabilidade, no âmbito do processo de reivindicação de posse, PSP doador
é o participante do Pix ao qual a chave está originalmente vinculada.

Durante o processo de reivindicação de posse, a partir do momento em que o DICT coloca o status do
pedido em “Aberto” até o momento em que o DICT altera o status para “Completo”, a chave Pix não
está passível de pedidos de registro e de exclusão^5. Durante os primeiros sete dias, enquanto o status
do pedido estiver como “Aguardando Resolução”, as consultas continuarão retornando normalmente a
identificação da conta transacional originalmente vinculada à chave. Após o sétimo dia, se não houver
indício de fraude ou anuência do usuário, o PSP doador deve confirmar a reivindicação. A chave será
desassociada do PSP doador e as consultas à chave retornarão a mensagem de “chave inexistente”.
Durante o período de resolução, tanto o PSP reivindicador quanto o PSP doador poderão cancelar o
pedido, em caso de suspeita de fraude ou em caso de solicitação de cancelamento pelo usuário do PSP
reivindicador. O pedido poderá ainda ser cancelado pelo PSP reivindicador caso seu status esteja como
“Aberto”.

O período de resolução de um processo de reivindicação de posse é de sete dias. Caso o usuário doador
não se manifeste dentro desse período de resolução, o PSP doador deve necessariamente confirmar a
reivindicação no DICT. Além do período de resolução, existe um período de encerramento, também de
sete dias, que se inicia imediatamente após o fim do período de resolução. Ao longo do período de
encerramento, o usuário doador ainda pode validar a posse da chave objeto da reivindicação,
cancelando a reivindicação. Nesse caso, o PSP do doador deve cancelar o processo, por indício de fraude.
Se, ao final do período de encerramento, o PSP doador não tiver cancelado a reinvindicação, o PSP
reivindicador deve solicitar a validação de posse ao seu usuário e, após essa validação, completar o
pedido no DICT. Caso o usuário reivindicador não faça a validação de posse até o trigésimo dia após o
início do processo de reivindicação, o PSP reivindicador deve necessariamente cancelar a solicitação no
DICT, para que a chave em disputa possa ser liberada.

O DICT disponibiliza um serviço que permite consultar e filtrar os processos de reivindicação de posse.
PSPs com acesso direto ao DICT deverão consultar pelo menos uma vez por minuto esse serviço a fim
de identificar mudanças de status das reivindicações de posse de interesse do PSP (além daquelas de
interesse dos PSPs com acesso indireto para os quais ele presta serviço), tanto como reivindicador
quanto como doador.

### 6.1 FLUXO DE REIVINDICAÇÃO DE POSSE PARA O PSP REIVINDICADOR COM ACESSO DIRETO AO DICT

(^5) É permitido ao PSP doador atualizar dados da conta vinculados à chave enquanto o status da requisição estiver
“Aberto” ou “Aguardando Resolução”.


**# Camada Tipo Descrição**

1 PSP reivindicador Ação

```
O processo é iniciado após a detecção de que já existe
registro para a chave solicitada e depois do pedido do
usuário.
Antes de enviar o pedido ao DICT, o PSP deve confirmar os
dados e situação cadastral do usuário final na Receita
Federal. Caso a situação seja irregular ou haja divergência,
o PSP não deve iniciar o processo e ir direto para a etapa
14.
```
2 PSP reivindicador Mensagem

```
O PSP reivindicador envia pedido de reivindicação de posse
ao DICT através da mensagem “Reivindicação / Criar
Reivindicação” do tipo “Posse”.
```
3 DICT Mensagem

```
DICT recebe mensagem com a solicitação de reivindicação
de posse de chave.
```
4 DICT Ação

```
DICT cria um pedido de reivindicação de posse com status
“Aberto”, com os dados da chave e do PSP reivindicador, e
inicia a contagem do período de resolução, aguardando
resposta do PSP doador.
```
5 DICT Ação

```
Uma vez recebida, do PSP doador, a confirmação ou o
cancelamento da reivindicação de posse, DICT atualiza o
status do pedido para "Confirmado" ou "Cancelado",
conforme o caso.
```

```
Caso o status seja "Cancelado", o DICT altera o status do
pedido para “Completo”, finalizando o processo.
Caso o status seja “Confirmado”, o DICT bloqueia a chave
até o recebimento da confirmação pelo PSP reivindicador.
O bloqueio significa que as consultas à essa chave no DICT
retornarão mensagem de erro.
```
6 PSP reivindicador Ação

```
PSP reivindicador consulta periodicamente o DICT até
identificar mudança no status da solicitação de
reivindicação de posse, através das mensagens
“Reivindicação / Listar Reivindicações” ou “Reivindicação /
Consultar Reivindicação”.
A solicitação pode ter três status: (i) “Cancelado”, com
motivo “Fraude”, caso o usuário que registrou a chave
originalmente tenha feito a validação ativa da posse da
chave até o 7º dia; (ii) “Confirmado”, com motivo “A pedido
do usuário”, caso o usuário que registrou a chave
originalmente tenha confirmado a reivindicação até o 7º
dia; ou (iii) “Confirmado”, com motivo “Padrão”, caso o
usuário que registrou a chave originalmente não tenha se
manifestado até o 7º dia.
No caso (i), PSP reivindicador prossegue diretamente para
a etapa 14.
No caso (ii), PSP reivindicador prossegue para a etapa 6. No
caso (iii), PSP reivindicador aguarda o fim do período de
encerramento. Caso o usuário que registrou a chave
originalmente tenha feito a validação ativa da posse da
chave entre o 8º e o 14º dia, a solicitação passará para o
status “Cancelado”, com motivo “Fraude”. Nesse caso, o
fluxo segue diretamente para a etapa 14. Caso o status da
solicitação continue como “Confirmado”, com motivo
“Padrão”, ao final do 14º dia, o PSP reivindicador segue o
fluxo a partir da etapa 7.
```
7 PSP reivindicador Ação

```
Caso o status do pedido seja “Confirmado”, o PSP
reivindicador faz a validação da posse da chave com o
usuário final.
```
8 PSP reivindicador Mensagem

```
Caso a validação de posse tenha sido efetuada com
sucesso, o PSP reivindicador solicita a conclusão do
processo no DICT, através da mensagem “Reivindicação /
Concluir Reivindicação”.
Caso a validação de posse não tenha sido efetuada até o
trigésimo dia após a abertura da reivindicação, o PSP
reivindicador solicita o cancelamento do processo no DICT,
através da mensagem “Reivindicação / Cancelar
Reivindicação”, com motivo “Padrão”. Nesse caso, DICT
altera o status da solicitação de reivindicação de posse para
“Cancelada” e o fluxo é encerrado, com a chave objeto da
reivindicação excluída do DICT e disponível para um novo
registro.
```

```
9 DICT Mensagem DICT recebe a solicitação de conclusão da reivindicação.
```
```
10 DICT Ação DICT atualiza os dados vinculados à chave e altera o status
da solicitação de reivindicação de posse para “Completo”.
```
```
11 DICT Mensagem
```
```
DICT envia mensagem de confirmação de atualização dos
dados vinculados à chave ao PSP reivindicador.
```
```
12 PSP reivindicador Mensagem
```
```
PSP reivindicador recebe confirmação de atualização dos
dados vinculados à chave.
```
```
13 PSP reivindicador Ação PSP reivindicador atualiza sua base de dados interna.
```
```
14 PSP reivindicador Comunicação
```
```
PSP reivindicador informa o usuário final sobre o
cancelamento ou a confirmação do pedido de reivindicação
de posse da chave.
```
```
15 Usuário final Comunicação
```
```
Usuário final recebe a informação sobre o cancelamento ou
a confirmação do pedido de reivindicação de posse da
chave.
```
### 6.2 FLUXO DE REIVINDICAÇÃO DE POSSE PARA O PSP REIVINDICADOR COM ACESSO INDIRETO AO DICT


**# Camada Tipo Descrição**

1 PSP reivindicador Ação

```
O processo é iniciado após a detecção de que já existe
registro para a chave solicitada e depois do pedido do
usuário.
Antes de enviar o pedido ao DICT, o PSP deve confirmar os
dados e situação cadastral do usuário final na Receita
Federal. Caso a situação seja irregular ou haja divergência,
o PSP não deve iniciar o processo e ir direto para a etapa
22.
```
2 PSP reivindicador Comunicação O PSP reivindicador envia pedido de reivindicação de posse
ao PSP com acesso direto ao DICT.

3

```
PSP com acesso
direto ao DICT Comunicação^
```
```
PSP com acesso direto ao DICT recebe pedido de
reivindicação de posse.
```

##### 4

```
PSP com acesso
direto ao DICT
```
```
Mensagem
```
```
PSP com acesso direto ao DICT encaminha pedido de
reivindicação de posse através da mensagem
“Reivindicação / Criar Reivindicação”.
```
```
5 DICT Mensagem
```
```
DICT recebe mensagem com a solicitação de reivindicação
de posse.
```
```
6 DICT Ação
```
```
DICT cria um pedido de reivindicação de posse com status
“Aberto”, com os dados da chave e do PSP reivindicador, e
inicia a contagem do período de resolução, aguardando
resposta do PSP doador.
```
```
7 DICT Ação
```
```
Uma vez recebida, do PSP doador, a confirmação ou o
cancelamento da reivindicação de posse, DICT atualiza o
status do pedido para "Confirmado" ou "Cancelado",
conforme o caso.
Caso o status seja "Cancelado", o DICT altera o status do
pedido para “Completo”, finalizando o processo.
Caso o status seja “Confirmado”, o DICT bloqueia a chave
até o recebimento da confirmação pelo PSP reivindicador.
O bloqueio significa que as consultas à essa chave no DICT
retornarão mensagem de erro.
```
##### 8

```
PSP com acesso
direto ao DICT
```
```
Ação
```
```
PSP com acesso direto consulta periodicamente o DICT até
identificar mudança no status da solicitação de
reivindicação de posse, através das mensagens
“Reivindicação / Listar Reivindicações” ou “Reivindicação /
Consultar Reivindicação”.
A solicitação pode ter três status: (i) “Cancelado”, com
motivo “Fraude”, caso o usuário que registrou a chave
originalmente tenha feito a validação ativa da posse da
chave até o 7º dia; (ii) “Confirmado”, com motivo “A pedido
do usuário”, caso o usuário que registrou a chave
originalmente tenha confirmado a reivindicação até o 7º
dia; ou (iii) “Confirmado”, com motivo “Padrão”, caso o
usuário que registrou a chave originalmente não tenha se
manifestado até o 7º dia.
Em qualquer um dos casos, PSP com acesso direto
prossegue para a etapa 8. No caso (iii), o PSP com aceso
direto deve esperar até o fim do período de encerramento
(catorze dias após o início do processo), caso a solicitação
não seja cancelada pelo PSP doador nesse período, para
poder seguir para a etapa 9.
```
##### 9

```
PSP com acesso
direto ao DICT
```
```
Comunicação
```
```
PSP com acesso direto envia comunicação ao PSP
reivindicador, identificando se o status do pedido foi
alterado para “Cancelado” ou para “Confirmado”.
```
10 PSP reivindicador Comunicação

```
PSP reivindicador recebe comunicação do PSP com acesso
direto informando mudança no status do pedido.
Caso o status seja "Cancelado", prossegue-se diretamente
à etapa 22.
```

11 PSP reivindicador Ação

```
Caso o status do pedido seja “Confirmado”, o PSP
reivindicador faz a validação da posse da chave com o
usuário final.
```
12 PSP reivindicador Comunicação

```
Caso a validação de posse tenha sido efetuada com
sucesso, o PSP reivindicador solicita conclusão da
reivindicação no DICT.
Caso a validação de posse não tenha sido efetuada até o
trigésimo dia após a abertura da reivindicação, o PSP
reivindicador solicita o cancelamento do processo no DICT.
```
13

```
PSP com acesso
direto ao DICT
```
```
Comunicação PSP com acesso direto recebe a solicitação.
```
##### 14

```
PSP com acesso
direto ao DICT
```
```
Mensagem
```
```
Caso a validação de posse tenha sido efetuada com
sucesso, PSP com acesso direto encaminha solicitação de
conclusão da reivindicação no DICT através da mensagem
“Reivindicação / Concluir Reivindicação”.
Caso a validação de posse não tenha sido efetuada até o
trigésimo dia após a abertura da reivindicação, o PSP com
acesso direto encaminha solicitação de cancelamento da
reivindicação no DICT através da mensagem “Reivindicação
/ Cancelar Reivindicação”, com motivo “Padrão”. Nesse
caso, DICT altera o status da solicitação de reivindicação de
posse para “Cancelada” e o fluxo é encerrado, com a chave
objeto da reivindicação excluída do DICT e disponível para
um novo registro.
```
15 DICT Mensagem DICT recebe a solicitação de conclusão da reivindicação.

16 DICT Ação

```
DICT atualiza os dados vinculados à chave e altera o status
da solicitação de reivindicação de posse para “Completo”.
```
17 DICT Mensagem

```
DICT envia mensagem de confirmação de atualização dos
dados vinculados à chave ao PSP com acesso direto.
```
18

```
PSP com acesso
direto ao DICT
```
```
Mensagem
```
```
PSP com acesso direto recebe confirmação de atualização
da chave.
```
19 PSP com acesso
direto ao DICT

```
Comunicação PSP com acesso direto encaminha confirmação de
atualização da chave.
```
20 PSP reivindicador Comunicação

```
PSP reivindicador recebe confirmação de atualização dos
dados vinculados à chave.
```
21 PSP reivindicador Ação PSP reivindicador atualiza sua base de dados interna.

22 PSP reivindicador Comunicação

```
PSP reivindicador informa o usuário final sobre o
cancelamento ou a confirmação do pedido de reivindicação
de posse da chave.
```
23 Usuário final Comunicação

```
Usuário final recebe a informação sobre o cancelamento ou
a confirmação do pedido de reivindicação de posse da
chave.
```

### 6.3 FLUXO DE REIVINDICAÇÃO DE POSSE PARA O PSP DOADOR COM ACESSO DIRETO AO DICT

```
# Camada Tipo Descrição
```
```
1 PSP doador Ação
```
```
PSP doador consulta periodicamente no DICT a lista de
reivindicações de posse com status “Aberto” que se
referem a chaves registradas por ele, através da
mensagem “Reivindicação / Listar Reivindicação”.
```
```
2 PSP doador
```
```
Mensagem/
Comunicação
```
```
Ao identificar reivindicação de posse com status
“Aberto”, PSP doador envia mensagem “Reivindicação /
Receber Reivindicação” ao DICT e notifica o usuário final,
solicitando confirmação da reivindicação ou validação da
posse da chave.
Caso o PSP doador receba comunicação do usuário até o
7º dia, ele avança para a etapa 8.
Caso o PSP do doador não receba comunicação do
usuário até o 7º dia, o PSP doador prossegue diretamente
para a etapa 11.
```

```
3 DICT Mensagem DICT recebe mensagem do PSP doador.
```
```
4 DICT Ação
```
```
DICT muda status do pedido para “Aguardando
Resolução” e fica aguardando o recebimento de
informação sobre o processo de reivindicação de posse.
```
##### 5

```
Usuário final que
registrou a chave
originalmente
```
```
Comunicação Usuário recebe notificação do PSP doador.
```
##### 6

```
Usuário final que
registrou a chave
originalmente
```
```
Ação
```
```
Usuário pode validar a posse da chave ou confirmar a
reivindicação de posse.
O usuário tem até sete dias para isso.
Caso o usuário se manifeste até o 7º dia, ele avança para
a etapa 7.
Caso o usuário não se manifeste até o 7º dia, ele avança
diretamente para a etapa 15.
```
##### 7

```
Usuário final que
registrou a chave
originalmente
```
```
Comunicação
```
```
Usuário envia comunicação ao PSP doador.
Caso a chave tenha sido validada, o processo é finalizado
para ele. O usuário permanece com a posse da chave.
Caso o processo de reivindicação tenha sido confirmado,
o processo é finalizado para ele. O usuário perde a posse
da chave.
```
```
8 PSP doador Comunicação
```
```
PSP doador recebe comunicação do usuário.
Caso o usuário tenha confirmado a reivindicação de
posse, PSP doador avança para etapa 9.
Caso o usuário faça a validação ativa da posse da chave,
o PSP avança diretamente para a etapa 10.
```
```
9 PSP doador Ação PSP doador exclui a chave de sua base interna.
```
10 PSP doador Mensagem

```
PSP doador envia mensagem ao DICT.
Caso o usuário tenha confirmado a reivindicação, PSP
envia mensagem “Reivindicação / Confirmar
Reivindicação”, com motivo “A pedido do usuário”.
Caso o usuário tenha validado a posse da chave, PSP
envia mensagem “Reivindicação / Cancelar
Reivindicação”, por motivo de “Fraude”.
Em qualquer dos casos, o fluxo vai diretamente para a
etapa 13.
```
11 PSP doador Ação PSP doador exclui a chave de sua base interna.

12 PSP doador Comunicação/^
Mensagem

```
PSP doador comunica o usuário sobre a exclusão da
chave e sobre a abertura de período adicional de sete
dias para que o usuário ainda possa validar a posse da
chave, cancelando o processo de reivindicação.
```

```
PSP doador também envia mensagem ao DICT
informando a exclusão da chave de sua base interna
através da mensagem “Reivindicação / Confirmar
Reivindicação”, com motivo “Padrão”.
```
13 DICT Mensagem DICT recebe mensagem do PSP doador.

14 DICT Ação

```
Caso o usuário que registrou a chave originalmente tenha
se manifestado até o 7º dia, DICT continua o processo
(etapa 4 do fluxo de reivindicação de posse do PSP
reivindicador com acesso direto ao DICT).
Caso DICT receba a mensagem de confirmação da
reivindicação com motivo “Padrão”, DICT bloqueia a
chave e aguarda o fim do período de encerramento
(entre o 8º e o 14º dia). Durante esse período, DICT pode
receber mensagem do PSP doador, com o cancelamento
da reivindicação.
Caso receba a mensagem, DICT vai para etapa 20.
Caso não receba, DICT continua o processo (etapa 4 do
fluxo de reivindicação de posse do PSP reivindicador com
acesso direto ao DICT).
O bloqueio significa que o DICT retornará a mensagem
“não registrado”, caso a chave em questão seja objeto de
alguma consulta.
```
##### 15

```
Usuário final que
registrou a chave
originalmente
```
```
Comunicação
```
```
Usuário recebe notificação do PSP doador.
Caso o usuário cancele a reivindicação entre o 8º e o 14º
dia, ele vai para a etapa 1 6.
Caso o usuário não cancele a reivindicação entre o 8º e o
14º dia, o fluxo é encerrado para ele.
```
##### 16

```
Usuário final que
registrou a chave
originalmente
```
```
Ação
```
```
Usuário pode validar a posse da chave em até sete dias.
Caso a chave tenha sido validada, o processo é finalizado
para ele. Como a chave já terá sido excluída da base
interna e do DICT, o usuário que registrou a chave
originalmente deve registrar novamente sua chave, se
assim desejar.
```
##### 17

```
Usuário final que
registrou a chave
originalmente
```
```
Comunicação
```
```
Usuário envia comunicação ao PSP doador, cancelando o
processo de reivindicação de posse.
```
18 PSP doador Comunicação PSP doador recebe comunicação do usuário.

19 PSP doador Mensagem

```
PSP doador envia mensagem ao DICT.
Caso o usuário tenha validado a posse da chave, PSP
envia mensagem “Reivindicação / Cancelar
Reivindicação”, por motivo de “Fraude”.
```

```
20 DICT Mensagem
```
```
DICT recebe a informação de cancelamento da
reivindicação de posse e continua o processo (etapa 4 do
fluxo de reivindicação de posse do PSP reivindicador com
acesso direto ao DICT).
```
### 6.4 FLUXO DE REIVINDICAÇÃO DE POSSE PARA O PSP DOADOR COM ACESSO INDIRETO AO DICT

```
# Camada Tipo Descrição
```
##### 1

```
PSP com acesso
direto ao DICT
```
```
Ação
```
```
PSP com acesso direto consulta periodicamente no DICT
a lista de reivindicações de posse com status “Aberto”
que se referem a chaves registradas pelo PSP doador ao
```

```
qual ele presta serviços, através da mensagem
“Reivindicação / Listar Reivindicação”.
```
##### 2

```
PSP com acesso
direto ao DICT
```
```
Mensagem/
Comunicação
```
```
Ao identificar reivindicação de posse com status
“Aberto”, PSP com acesso direto envia mensagem
“Reivindicação / Receber Reivindicação” ao DICT e envia
comunicação ao PSP doador.
3 PSP doador Comunicação PSP doador recebe comunicação.
```
```
4 PSP doador Comunicação
```
```
PSP doador notifica o usuário final, solicitando
confirmação da reivindicação ou validação da posse da
chave.
Caso o PSP doador receba comunicação do usuário até o
7º dia, ele avança para a etapa 10.
Caso o PSP do doador não receba comunicação do
usuário até o 7º dia, o PSP doador prossegue diretamente
para a etapa 1 5.
```
```
5 DICT Mensagem DICT recebe mensagem do PSP com acesso direto.
```
```
6 DICT Ação
```
```
DICT muda status do pedido para “Aguardando
Resolução” e fica aguardando o recebimento de
informação sobre o processo de reivindicação de posse.
```
##### 7

```
Usuário final que
registrou a chave
originalmente
```
```
Comunicação Usuário recebe notificação do PSP doador.
```
##### 8

```
Usuário final que
registrou a chave
originalmente
```
```
Ação
```
```
Usuário pode validar a posse da chave ou confirmar a
reivindicação de posse.
O usuário tem até sete dias para isso.
Caso o usuário se manifeste até o 7º dia, ele avança para
a etapa 9.
Caso o usuário não se manifeste até o 7º dia, ele avança
diretamente para a etapa 2 1.
```
##### 9

```
Usuário final que
registrou a chave
originalmente
```
```
Comunicação
```
```
Usuário envia comunicação ao PSP doador.
Caso a chave tenha sido validada, o processo é finalizado
para ele. O usuário permanece com a posse da chave.
Caso o processo de reivindicação tenha sido confirmado,
o processo é finalizado para ele. O usuário perde a posse
da chave.
```
10 PSP doador Comunicação

```
PSP doador recebe comunicação do usuário.
Caso o usuário tenha confirmado a reivindicação de
posse, PSP doador avança para etapa 11.
Caso o usuário faça a validação ativa da posse da chave,
o PSP avança diretamente para a etapa 1 2.
```
11 PSP doador Ação PSP doador exclui a chave de sua base interna.


12 PSP doador Comunicação PSP doador envia comunicação para PSP com acesso
direto.

##### 13

```
PSP com acesso
direto ao DICT Comunicação^ PSP com acesso direto recebe comunicação.^
```
##### 14

```
PSP com acesso
direto ao DICT
Mensagem
```
```
PSP com acesso direto envia mensagem ao DICT.
Caso o usuário tenha confirmado a reivindicação, PSP
envia mensagem “Reivindicação / Confirmar
Reivindicação”, com motivo “A pedido do usuário”.
Caso o usuário tenha validado a posse da chave, PSP
envia mensagem “Reivindicação / Cancelar
Reivindicação”, por motivo de “Fraude”.
Em qualquer dos casos, o fluxo vai diretamente para a
etapa 19.
```
15 PSP doador Ação PSP doador exclui a chave de sua base interna.

16 PSP doador

```
Comunicação
```
```
PSP doador comunica o usuário sobre a exclusão da
chave e sobre a abertura de período adicional de sete
dias para que o usuário ainda possa validar a posse da
chave, cancelando o processo de reivindicação.
PSP doador também envia comunicação ao PSP com
acesso direto, informando a exclusão da chave de sua
base interna.
```
##### 17

```
PSP com acesso
direto ao DICT Comunicação^ PSP com acesso direto recebe comunicação.^
```
18 PSP com acesso
direto ao DICT

```
PSP com acesso direto envia mensagem ao DICT
informando a exclusão da chave da base interna do PSP
doador, através da mensagem “Reivindicação /
Confirmar Reivindicação”, com motivo “Padrão”.
```
19 DICT Mensagem DICT recebe mensagem do PSP com acesso direto.

20 DICT Ação

```
Caso o usuário que registrou a chave originalmente tenha
se manifestado até o 7º dia, DICT continua o processo
(etapa 4 do fluxo de reivindicação de posse do PSP
reivindicador com acesso direto ao DICT).
Caso DICT receba a mensagem de confirmação da
reivindicação com motivo “Padrão”, DICT bloqueia a
chave e aguarda o fim do período de encerramento
(entre o 8º e o 14º dia). Durante esse período, DICT pode
receber mensagem com o cancelamento da
reivindicação.
Caso receba a mensagem, DICT vai para etapa 28.
```

```
Caso não receba, DICT continua o processo (etapa 4 do
fluxo de reivindicação de posse do PSP reivindicador com
acesso direto ao DICT).
O bloqueio significa que o DICT retornará a mensagem
“não registrado”, caso a chave em questão seja objeto de
alguma consulta.
```
##### 21

```
Usuário final que
registrou a chave
originalmente
```
```
Comunicação
```
```
Usuário recebe notificação do PSP doador.
Caso o usuário cancele a reivindicação entre o 8º e o 14º
dia, ele vai para a etapa 22.
Caso o usuário não cancele a reivindicação entre o 8º e o
14º dia, o fluxo é encerrado para ele.
```
##### 22

```
Usuário final que
registrou a chave
originalmente
```
```
Ação
```
```
Usuário pode validar a posse da chave em até sete dias.
Caso a chave tenha sido validada, o processo é finalizado
para ele. Como a chave já terá sido excluída da base
interna e do DICT, o usuário que registrou a chave
originalmente deve registrar novamente sua chave, se
assim desejar.
```
##### 23

```
Usuário final que
registrou a chave
originalmente
```
```
Comunicação
```
```
Usuário envia comunicação ao PSP doador, cancelando o
processo de reivindicação de posse.
```
```
24 PSP doador Comunicação PSP doador recebe comunicação do usuário.
```
```
25 PSP doador Comunicação PSP doador envia comunicação ao PSP com acesso direto.
```
##### 26

```
PSP com acesso
direto ao DICT Comunicação^ PSP com acesso direto recebe comunicação.^
```
```
27 PSP com acesso
direto ao DICT
```
```
Mensagem
```
```
PSP com acesso direto envia mensagem ao DICT.
Caso o usuário tenha validado a posse da chave, PSP
envia mensagem “Reivindicação / Cancelar
Reivindicação”, por motivo de “Fraude”.
```
```
28 DICT Mensagem
```
```
DICT recebe a informação de cancelamento da
reivindicação de posse e continua o processo (etapa 6 do
fluxo de reivindicação de posse do PSP reivindicador com
acesso direto ao DICT).
```
## 7 FLUXO DE ALTERAÇÃO DOS DADOS VINCULADOS À CHAVE

### 7.1 FLUXO DE ALTERAÇÃO DOS DADOS VINCULADOS À CHAVE (PARTICIPANTES DO PIX COM ACESSO DIRETO AO DICT)



```
# Camada Tipo Descrição
```
```
1 PSP reivindicador Ação
```
```
Processo pode ser iniciado pelo usuário ou pelo próprio
PSP.
```
```
2 PSP reivindicador Ação
```
```
Caso a alteração seja de nome completo, nome empresarial
ou nome de fantasia, PSP faz a validação dos dados
alterados com o cadastro do usuário na Receita Federal.
```
```
3 PSP reivindicador Mensagem
```
```
PSP reivindicador envia pedido de alteração dos dados
(nome completo, nome empresarial, nome de fantasia,
agência ou agência e conta) vinculados à chave, através da
mensagem “Diretório / Atualizar Vínculo”.
```
```
4 DICT Mensagem
```
```
DICT recebe o pedido de alteração dos dados encaminhado
pelo PSP reivindicador.
```
```
5 DICT Ação
```
```
DICT verifica se a chave está cadastrada para o PSP e para
o usuário envolvidos. Caso o resultado seja negativo, o
pedido é rejeitado e segue-se à etapa 6.
```
```
6 DICT Ação
```
```
Caso o DICT verifique que a chave é do PSP e do usuário
envolvidos, os dados vinculados à chave são alterados.
```
```
7 DICT Mensagem
```
```
DICT informa a confirmação ou a rejeição do pedido de
alteração dos dados.
```
```
8 PSP reivindicador Mensagem
```
```
PSP reivindicador recebe informação de confirmação ou de
rejeição do DICT.
```
```
9 PSP reivindicador Ação
```
```
Após receber a informação de alteração dos dados
vinculados à chave, o PSP reivindicador atualiza sua base
interna com os novos dados da chave.
```
```
10 PSP reivindicador Comunicação
```
```
PSP reivindicador informa o usuário final sobre o
cancelamento ou a confirmação do pedido de alteração dos
dados.
```
```
11 Usuário final Comunicação
```
```
Usuário final recebe a informação de confirmação ou de
cancelamento da alteração dos dados.
```
### 7.2 FLUXO DE ALTERAÇÃO DOS DADOS VINCULADOS À CHAVE (PARTICIPANTES DO PIX COM ACESSO INDIRETO AO DICT)



**# Camada Tipo Descrição**

1 PSP reivindicador Ação

```
Processo pode ser iniciado pelo usuário ou pelo próprio
PSP.
```
2 PSP reivindicador Ação

```
Caso a alteração seja de nome completo, nome empresarial
ou nome de fantasia, PSP faz a validação cadastral do
usuário na Receita Federal.
```
3 PSP reivindicador Comunicação

```
PSP reivindicador envia pedido de alteração dos dados
(nome completo, nome empresarial, nome de fantasia,
agência ou agência e conta) vinculados à chave.
```
4

```
PSP com acesso
direto ao DICT Comunicação^
```
```
PSP com acesso direto ao DICT recebe pedido de alteração
dos dados.
```
##### 5

```
PSP com acesso
direto ao DICT
```
```
Mensagem
```
```
PSP com acesso direto encaminha pedido de alteração dos
dados ao DICT através da mensagem “Diretório / Atualizar
Vínculo”.
```

```
6 DICT Mensagem DICT recebe o pedido de alteração dos dados.
```
```
7 DICT Ação
```
```
DICT verifica se o PSP com acesso direto pode enviar
solicitações para o PSP reivindicador.
```
```
8 DICT Ação
```
```
DICT verifica se a chave está cadastrada para o PSP
reivindicador e para o usuário envolvidos. Caso o resultado
seja negativo, o pedido é rejeitado e segue-se à etapa 9.
```
```
9 DICT Ação
```
```
Caso o DICT verifique que a chave é do PSP reivindicador e
do usuário envolvidos, os dados vinculados à chave são
alterados.
```
```
10 DICT Mensagem DICT informa a confirmação ou a rejeição do pedido de
alteração dos dados.
```
##### 11

```
PSP com acesso
direto ao DICT
```
```
Mensagem
```
```
PSP com acesso direto ao DICT recebe informação de
confirmação ou de rejeição do pedido de alteração dos
dados.
```
##### 12

```
PSP com acesso
direto ao DICT Comunicação^
```
```
PSP com acesso direto ao DICT encaminha informação de
confirmação ou de rejeição do pedido de alteração dos
dados.
```
```
13 PSP reivindicador Comunicação
```
```
PSP reivindicador recebe informação de confirmação ou de
rejeição.
```
```
14 PSP reivindicador Ação
```
```
Após receber a informação de alteração dos dados
vinculados à chave, o PSP reivindicador atualiza sua base
interna com os novos dados da chave.
```
```
15 PSP reivindicador Comunicação
```
```
PSP reivindicador informa o usuário final sobre o
cancelamento ou a confirmação do pedido de alteração dos
dados.
```
```
16 Usuário final Comunicação
```
```
Usuário final recebe a informação de confirmação ou de
cancelamento da alteração dos dados.
```
### 7.3 ALTERAÇÃO DOS DADOS VINCULADOS À CHAVE PARA CORREÇÃO DE INCONSISTÊNCIAS

Na hipótese de alteração dos dados vinculados à chave Pix independentemente de solicitação do
usuário, o participante deve incluir na comunicação feita ao usuário os motivos da alteração.

## 8 FLUXO DE CONSULTA DE CHAVE

O DICT retornará as seguintes informações após a consulta de uma chave (sempre que a chave estiver
registrada):

- chave consultada;
- tipo da chave consultada;
- identificação (ISPB) do participante que registrou a chave consultada;
- número da agência vinculada à chave consultada (se houver);
- número da conta vinculada à chave consultada;
- tipo da conta vinculada à chave consultada;
- data de abertura da conta vinculada à chave consultada;


- natureza jurídica do usuário que registrou a chave consultada;
- identificação (CPF ou CNPJ) do usuário que registrou a chave consultada;
- nome completo do usuário que registrou a chave consultada, podendo, a critério do usuário
    final, ser o nome civil ou o nome social, conforme registrados no CPF, caso a chave consultada
    esteja vinculada a um usuário pessoa física;
- nome empresarial do usuário que registrou a chave consultada, conforme registrado no CNPJ,
    caso a chave consultada esteja vinculada a um usuário pessoa jurídica;
- título do estabelecimento (nome de fantasia) do usuário que registrou a chave consultada, se
    registrado no CNPJ, caso a chave consultada esteja vinculada a um usuário pessoa jurídica;
- data de registro da chave consultada no DICT;
- data de registro da chave consultada no participante atual vinculado à chave;
- data de abertura da portabilidade ou da reivindicação de posse da data consultada, caso
    alguma dessas operações esteja aberta no momento da consulta da chave.

Além dos dados acima, o participante poderá informar, através do parâmetro opcional _includeStatistics,_
se as informações de segurança da chave deverão ser inseridas na resposta do DICT. Caso o parâmetro
seja omitido, apenas os dados cadastrais vinculados à chave serão retornados.

O detalhamento das informações de segurança está disponível na seção 1 8.

### 8.1 DADOS DE CHAVE PERMITIDOS NA EXIBIÇÃO AO USUÁRIO

Das informações retornadas pelo DICT na consulta de chave, apenas algumas podem ser disponibilizadas
ao usuário que faz a consulta, seja por meio do aplicativo, do _internet banking_ , dos sistemas internos ou
da API do PSP:

- nome completo ou nome empresarial^6 do usuário que registrou a chave consultada;
- título do estabelecimento (nome fantasia) do usuário que registrou a chave consultada, caso
    esteja vinculada a um usuário pessoa jurídica;
- CPF mascarado (exemplo: ***.777.888-**) ou CNPJ do usuário que registrou a chave
    consultada;
- chave consultada;
- nome do PSP do recebedor (opcional).

### 8.2 FLUXO DE CONSULTA PARA O PARTICIPANTE DO PIX COM ACESSO DIRETO AO DICT...................................................

(^6) No caso de aplicativo de usuário pessoa natural, conforme determina o manual de Requisitos Mínimos para a
Experiência do Usuário, o nome empresarial somente pode ser disponibilizado caso o nome de fantasia não
esteja registrado na chave Pix


**# Camada Tipo Descrição**
1 Usuário pagador Ação Usuário pagador acessa seu canal de atendimento.

2 Usuário pagador Ação

Usuário pagador insere, manualmente ou através da leitura
de um QR Code, os dados da chave, que podem
corresponder a qualquer um dos cinco tipos existentes, do
usuário recebedor.
3 Usuário pagador Comunicação Usuário pagador encaminha solicitação a seu PSP.

4 PSP do usuário
pagador

```
Comunicação PSP do usuário pagador recebe solicitação.
```
##### 5

```
PSP do usuário
pagador Ação^ PSP verifica se a chave está cadastrada em sua base interna.^
```
##### 6

```
PSP do usuário
pagador
```
```
Ação
```
```
Caso a chave esteja cadastrada em sua base interna, o PSP
cria mensagem de resposta. Após, segue-se direto para a
etapa 15.
```
7 PSP do usuário
pagador

```
Mensagem
```
```
Caso a chave não esteja cadastrada em sua base interna, o
PSP do usuário pagador envia mensagem de consulta ao
DICT, na qual deve informar, além da chave Pix, o
identificador único da transação que será utilizado na
mensagem de liquidação. Esse identificador único
corresponde ao campo EndToEndId , que é um campo
obrigatório da mensagem de liquidação PACS.008. Ele
```

```
deverá ser gerado pelo PSP e transmitido durante a
operação de consulta e novamente na PACS.008. Outra
informação que deverá ser enviada é o identificador do
usuário pagador (seu CPF ou CNPJ). A mensagem de
consulta é “Diretório / Consultar Vínculo”.
8 DICT Mensagem DICT recebe mensagem com solicitação de consulta.
```
```
9 DICT Ação
```
```
DICT verifica se a instituição é autorizada a realizar
consultas.
```
```
10 DICT Ação
```
```
DICT verifica se a chave está registrada.
Caso o PSP do pagador seja o mesmo PSP vinculado à
chave, o DICT irá rejeitar a consulta e retornar mensagem
de erro.
```
```
11 DICT Ação
```
```
Caso a chave esteja registrada, DICT cria mensagem de
resposta, que deve conter todos os dados vinculados à
chave.
```
```
12 DICT Ação
```
```
Caso a chave não esteja registrada, DICT cria mensagem
informando a inexistência da chave consultada.
```
```
13 DICT Mensagem DICT envia mensagem de resposta à consulta.
```
```
14 PSP do usuário
pagador
```
```
Mensagem PSP do usuário pagador^ recebe mensagem com a resposta
à consulta.
```
##### 15

```
PSP do usuário
pagador
```
```
Comunicação
```
```
PSP do usuário pagador encaminha resposta à consulta
para o usuário final.
O número de inscrição no CPF, caso presente na resposta,
deve ser disponibilizado no formato ***.111.111-**
(substitui os números do início e do fim por asteriscos).
```
```
16 Usuário pagador Comunicação
```
```
Usuário pagador recebe mensagem de resposta à sua
consulta.
```
### 8.3 FLUXO DE CONSULTA PARA O PARTICIPANTE DO PIX COM ACESSO INDIRETO AO DICT


**# Camada Tipo Descrição**
1 Usuário pagador Ação Usuário pagador acessa seu canal de atendimento.

2 Usuário pagador Ação

Usuário pagador insere, manualmente ou através da leitura
de um QR Code, os dados da chave, que podem
corresponder a qualquer um dos cinco tipos existentes, do
usuário recebedor.
3 Usuário pagador Comunicação Usuário pagador encaminha solicitação a seu PSP.

4

```
PSP do usuário
pagador Comunicação^ PSP do usuário pagador recebe solicitação.^
```
5

```
PSP do usuário
pagador
```
```
Ação PSP verifica se a chave está cadastrada em sua base interna.
```
##### 6

```
PSP do usuário
pagador
```
```
Ação
```
```
Caso a chave esteja cadastrada em sua base interna, o PSP
cria mensagem de resposta. Após, segue-se direto para a
etapa 21.
```

```
7 PSP do usuário
pagador
```
```
Comunicação
```
```
Caso a chave não esteja cadastrada em sua base interna, o
PSP do usuário pagador envia comunicação ao PSP com
acesso direto, solicitando consulta ao DICT.
```
```
8
```
```
PSP com acesso
direto ao DICT Comunicação^
```
```
PSP com acesso direto recebe comunicação com a
solicitação de consulta.
```
```
9
```
```
PSP com acesso
direto ao DICT
```
```
Ação
```
```
PSP com acesso direto ao DICT verifica se a chave está
cadastrada em sua base interna.
```
10 PSP com acesso
direto ao DICT

```
Ação Caso a chave esteja cadastrada em sua base interna, segue-
se direto para a etapa 19.
```
##### 11

```
PSP com acesso
direto ao DICT
```
```
Mensagem
```
Caso a chave não esteja cadastrada em sua base interna,
PSP com acesso direto ao DICT encaminha mensagem com
solicitação de consulta ao DICT. A mensagem deve conter,
além da chave Pix, o identificador único da transação que
será utilizado na mensagem de liquidação. Esse
identificador único corresponde ao campo _EndToEndId_ ,
que é um campo obrigatório da mensagem de liquidação
PACS.008. Ele deverá ser gerado pelo PSP e transmitido
durante a operação de consulta e novamente na PACS.008.
Outra informação que deverá ser enviada é o identificador
do usuário pagador (seu CPF ou CNPJ). A mensagem de
consulta é “Diretório / Consultar Vínculo”.
12 DICT Mensagem DICT recebe mensagem com solicitação de consulta.

13 DICT Ação

```
DICT verifica se a instituição é autorizada a realizar
consultas. O DICT deve também verificar se o PSP com
acesso direto tem autorização para consultar chaves para o
PSP com acesso indireto.
```
14 DICT Ação

```
DICT verifica se a chave está registrada.
Caso o PSP do pagador seja o mesmo PSP vinculado à
chave, o DICT irá rejeitar a consulta e retornar mensagem
de erro.
```
15 DICT Ação

```
Caso a chave esteja registrada, DICT cria mensagem de
resposta, que deve conter todos os dados vinculados à
chave.
```
16 DICT Ação Caso a chave não esteja registrada, DICT cria mensagem
informando a inexistência da chave consultada.

17 DICT Mensagem DICT envia mensagem de resposta à consulta.

##### 18

```
PSP com acesso
direto ao DICT
```
```
Mensagem PSP com acesso direto ao DICT recebe resposta da consulta.
```
##### 19

```
PSP com acesso
direto ao DICT
```
```
Comunicação
```
```
PSP com acesso direto ao DICT comunica ao PSP do usuário
pagador a resposta da consulta.
```
20

```
PSP do usuário
pagador
```
```
Comunicação
```
```
PSP do usuário pagador recebe comunicação sobre a
resposta à consulta.
```
21

```
PSP do usuário
pagador
```
```
Comunicação
```
```
PSP do usuário pagador encaminha resposta à consulta
para o usuário final.
```

```
O número de inscrição no CPF, caso presente na resposta,
deve ser disponibilizado no formato ***.111.111-**
(substitui os números do início e do fim por asteriscos).
```
22 Usuário pagador Comunicação

```
Usuário pagador recebe mensagem de resposta à sua
consulta.
```
### DE PAGAMENTO, COM ACESSO DIRETO AO DICT 8.4 FLUXO DE CONSULTA PARA O PARTICIPANTE DO PIX QUE ATUA COMO PRESTADOR DE SERVIÇO DE INICIAÇÃO DE TRANSAÇÃO

### serviço de iniciação de transação de pagamento, com acesso direto ao DICT

```
# Camada Tipo Descrição
```
```
1 Usuário pagador Ação
```
```
Usuário pagador acessa interface do participante que
presta serviço de iniciação de transação de pagamento.
```
```
2 Usuário pagador Ação
```
```
Usuário pagador insere, manualmente ou através da leitura
de um QR Code, os dados da chave, que podem
corresponder a qualquer um dos cinco tipos existentes, do
usuário recebedor.
3 Usuário pagador Comunicação Usuário pagador encaminha solicitação.
```
```
4
```
```
Prestador de serviço
de iniciação Comunicação^ Prestador de serviço de iniciação recebe solicitação.^
```
##### 5

```
Prestador de serviço
de iniciação
```
```
Mensagem
```
```
Prestador de serviço de iniciação envia mensagem de
consulta ao DICT, na qual deve informar a chave Pix do
usuário recebedor. Além da chave Pix, o prestador de
serviço de iniciação deve informar o identificador único da
transação que será utilizado na mensagem de liquidação.
Esse identificador único corresponde ao campo
```

```
EndToEndId , que é um campo obrigatório da mensagem de
liquidação PACS.008. Ele deverá ser gerado pelo prestador
de serviço de iniciação, enviado ao DICT na operação de
consulta e transmitido pelo PSP do pagador na PACS.008
enviada ao SPI. Outra informação que deverá ser enviada é
o identificador do usuário pagador (seu CPF ou CNPJ). A
mensagem de consulta é “Diretório / Consultar Vínculo”.
Participantes iniciadores deverão se identificar, na
mensagem de consulta, por meio dos oito primeiros dígitos
de seu CNPJ (campo PI-RequestingParticipant ).
6 DICT Mensagem DICT recebe mensagem com solicitação de consulta.
```
```
7 DICT Ação
```
```
DICT verifica se a instituição é autorizada a realizar
consultas.
```
```
8 DICT Ação DICT verifica se a chave está registrada.
```
```
9 DICT Ação
```
```
Caso a chave esteja registrada, DICT cria mensagem de
resposta, que deve conter todos os dados vinculados à
chave.
```
10 DICT Ação Caso a chave não esteja registrada, DICT cria mensagem
informando a inexistência da chave consultada.

11 DICT Mensagem DICT envia mensagem de resposta à consulta.

##### 12

```
Prestador de serviço
de iniciação Mensagem^
```
```
Prestador de serviço de iniciação recebe mensagem com a
resposta à consulta.
```
##### 13

```
Prestador de serviço
de iniciação
```
```
Comunicação
```
```
Prestador de serviço de iniciação encaminha resposta à
consulta para o usuário final.
O número de inscrição no CPF, caso presente na resposta,
deve ser disponibilizado no formato ***.111.111-**
(substitui os números do início e do fim por asteriscos).
```
14 Usuário pagador Comunicação

```
Usuário pagador recebe mensagem de resposta à sua
consulta.
```
### DE PAGAMENTO, COM ACESSO INDIRETO AO DICT 8.5 FLUXO DE CONSULTA PARA O PARTICIPANTE DO PIX QUE ATUA COMO PRESTADOR DE SERVIÇO DE INICIAÇÃO DE TRANSAÇÃO

### serviço de iniciação de transação de pagamento, com acesso indireto ao DICT


**# Camada Tipo Descrição**

1 Usuário pagador Ação

```
Usuário pagador acessa interface do participante que
presta serviço de iniciação de transação de pagamento.
```
2 Usuário pagador Ação

Usuário pagador insere, manualmente ou através da leitura
de um QR Code, os dados da chave, que podem
corresponder a qualquer um dos cinco tipos existentes, do
usuário recebedor.
3 Usuário pagador Comunicação Usuário pagador encaminha solicitação.

4

```
Prestador de serviço
de iniciação Comunicação^ Prestador de serviço de iniciação recebe solicitação.^
```
5 Prestador de serviço
de iniciação

```
Comunicação
```
```
Prestador de serviço de iniciação envia comunicação ao PSP
com acesso direto, solicitando consulta ao DICT. Além da
chave Pix, o prestador de serviço de iniciação deve informar
o identificador único da transação que será utilizado na
mensagem de liquidação. Esse identificador único
corresponde ao campo EndToEndId , que é um campo
obrigatório da mensagem de liquidação PACS.008. Ele
deverá ser gerado pelo prestador de serviço de iniciação,
```

```
enviado ao DICT pelo PSP com acesso direto na operação
de consulta e transmitido pelo PSP do pagador na PACS.008
enviada ao SPI. Outra informação que deverá ser enviada é
o identificador do usuário pagador (seu CPF ou CNPJ).
```
##### 6

```
PSP com acesso
direto ao DICT Comunicação^
```
```
PSP com acesso direto recebe comunicação com a
solicitação de consulta.
```
##### 7

```
PSP com acesso
direto ao DICT
```
```
Ação
```
```
PSP com acesso direto ao DICT verifica se a chave está
cadastrada em sua base interna.
```
##### 8

```
PSP com acesso
direto ao DICT
```
```
Ação
```
```
Caso a chave esteja cadastrada em sua base interna, segue-
se direto para a etapa 1 7.
```
```
9 PSP com acesso
direto ao DICT
```
```
Mensagem
```
Caso a chave não esteja cadastrada em sua base interna,
PSP com acesso direto ao DICT encaminha mensagem com
solicitação de consulta. A mensagem de consulta é
“Diretório / Consultar Vínculo”.
Participantes iniciadores deverão ser identificados, na
mensagem de consulta, por meio dos oito primeiros dígitos
de seu CNPJ (campo _PI-RequestingParticipant_ ).
10 DICT Mensagem DICT recebe mensagem com solicitação de consulta.

11 DICT Ação

```
DICT verifica se a instituição é autorizada a realizar
consultas. O DICT deve também verificar se o PSP com
acesso direto tem autorização para consultar chaves para o
prestador de serviço de iniciação.
```
12 DICT Ação DICT verifica se a chave está registrada.

13 DICT Ação

```
Caso a chave esteja registrada, DICT cria mensagem de
resposta, que deve conter todos os dados vinculados à
chave.
```
14 DICT Ação

```
Caso a chave não esteja registrada, DICT cria mensagem
informando a inexistência da chave consultada.
```
15 DICT Mensagem DICT envia mensagem de resposta à consulta.

##### 16

```
PSP com acesso
direto ao DICT
```
```
Mensagem
```
```
PSP com acesso direto recebe mensagem com a resposta à
consulta.
```
17

```
PSP com acesso
direto ao DICT
```
```
Comunicação
```
```
PSP com acesso direto comunica ao prestador de serviço de
iniciação a resposta da consulta.
```
18

```
Prestador de serviço
de iniciação Comunicação^
```
```
Prestador de serviço de iniciação recebe comunicação
sobre a resposta à consulta.
```
##### 19

```
Prestador de serviço
de iniciação
```
```
Comunicação
```
```
Prestador de serviço de iniciação encaminha resposta à
consulta para o usuário final.
O número de inscrição no CPF, caso presente na resposta,
deve ser disponibilizado no formato ***.111.111-**
(substitui os números do início e do fim por asteriscos).
```
20 Usuário pagador Comunicação

```
Usuário pagador recebe mensagem de resposta à sua
consulta.
```

## 9 FLUXO DE VERIFICAÇÃO DE SINCRONISMO

O processo de verificação de sincronismo apoia-se em dois conceitos: identificador de conteúdo
( _content identifier_ ou CID) e verificador de sincronismo ( _VSync_ ).

CID é um número único de 256 bits gerado a partir de um _hash_ do estado da chave (chave junto com as
informações vinculadas a ela). Sua regra de criação está especificada na API do DICT.

Usando a regra de formação, o CID é calculado de forma independente do lado do participante do Pix e
do DICT. O participante do Pix deverá armazenar esse número na sua base de dados interna,
preferencialmente de forma indexada, pois o processo de verificação de sincronismo será baseado nele.
A utilização do CID permite: (i) a comparação simplificada entre dois registros, uma vez que é mais
simples comparar dois números do que duas chaves com diversos atributos; (ii) a possibilidade de
verificação interna da integridade de cada registro de chave, permitindo identificar uma mudança feita
diretamente sobre um atributo sem a atualização do CID; e (iii) a possibilidade de gerar o _VSync_ de forma
otimizada.

O _VSync_ é o resultado da aplicação de uma função _XOR_ (‘ou’ exclusivo) num conjunto de CIDs. Como os
CIDs são únicos, aleatórios e suficientemente grandes, a operação de _XOR_ entre CIDs mantém a chance
de colisão dos próprios CIDs (infinitesimalmente pequena). Assim, é possível garantir que, se o _VSync_
gerado pelo DICT e pelo participante do Pix forem o mesmo, o conjunto de CIDs é o mesmo e ambos
possuem as mesmas chaves com os mesmos atributos. Cada participante deverá manter cinco _VSyncs_ ,
um para cada conjunto de CIDs por tipo de chave (CPF, CNPJ, número de telefone celular, e-mail e chave
aleatória).

O PSP pode verificar o sincronismo entre a sua base e o DICT utilizando as seguintes funcionalidades: (i)
a verificação de sincronismo do _VSync_ ; (ii) o log de alterações de chaves do participante; e (iii) o arquivo
de CIDs registrados no DICT.

Para a verificação de sincronismo do _VSync_ , o PSP deverá informar os _VSyncs_ de cada tipo de chave. O
DICT responderá se o conjunto está síncrono ou não. Caso a verificação de sincronismo falhe, o PSP
deverá tomar ações para correção das divergências. Após as correções, o PSP deve enviar novamente o
_VSync_.

O processo de verificação de sincronismo pode ser mais simples se for realizado fora da janela
obrigatória de disponibilização de registro, de exclusão, de alteração, de portabilidade e de reivindicação
de posse. Assim, o participante pode suspender temporariamente essas funcionalidades, realizar o
cálculo do _VSync_ e atuar para corrigir eventuais divergências, sem se preocupar com alterações
concorrentes no conjunto de CIDs.

O participante do Pix também pode acompanhar o log de alteração de chaves. O log mostra os CIDs que
estão sendo registrados e excluídos do DICT, a partir das operações realizadas pelo participante. Assim,
é recomendado ao participante construir um mecanismo de acompanhamento do log, garantindo que
sua base esteja continuamente síncrona com o DICT.

Por fim, caso seja necessária uma reconciliação completa, é possível fazer a verificação individual de
chaves do participante, solicitando um arquivo com os CIDs registrados no DICT. Esse arquivo contém


os CIDs de todas as chaves ativas, sem nenhuma informação adicional, visando simplificar o
armazenamento e a distribuição desse arquivo, do ponto de vista de segurança. Assim, o participante
deve buscar os CIDs armazenados em sua base interna e verificar se existe a mesma entrada listada no
arquivo.

A partir do batimento do arquivo com a base interna do participante, poderá haver divergências de dois
tipos: (i) CIDs que existem somente no arquivo (e não na base do participante); e (ii) CIDs que existem
somente no participante (e não no arquivo).

É recomendado aos participantes manter, do seu lado, o registro de todos os CIDs já gerados, associados
à chave envolvida na operação. Isso permitirá identificar, nos casos de divergências, qual chave a
originou. Caso não localize o CID em sua base, o participante ainda poderá realizar uma consulta a partir
do CID e receber os dados vinculados à chave como resposta.

Qualquer modificação necessária em decorrência do processo de verificação de sincronismo deve ser
realizada:

- para chaves registradas somente no DICT, e não no participante, a chave deve ser incluída na
    base do participante com os mesmos atributos registrado no DICT;
- para chaves registradas somente no participante, e não no DICT, a chave deve ser removida da
    base do participante; e
- para chaves registradas em ambas as bases, mas com valores divergentes, deve haver correção
    dos valores na base do participante, exatamente como consta na base do DICT.

O DICT disponibiliza o motivo “reconciliação” para quando for necessária a atualização de informações
por erros nos processos do participante, por exemplo ao inserir os dados da conta incorretamente. As
outras alterações devem, sempre que possível, usar os motivos relacionados ao fato que gerou a perda
de sincronia.

É importante ressaltar que, ao término das correções, o participante deve gerar novamente o _VSync_ e
enviá-lo ao DICT para confirmar o sincronismo. Caso as bases ainda estejam divergentes, o participante
deve reiniciar o processo de reconciliação.

### 9.1 VERIFICAÇÃO DE VSYNC (PARTICIPANTE DO PIX COM ACESSO DIRETO AO DICT)


```
# Camada Tipo Descrição
```
```
1
```
```
PSP com acesso
direto ao DICT
```
```
Ação PSP acessa canal de comunicação com o DICT.
```
##### 2

```
PSP com acesso
direto ao DICT Mensagem^
```
```
PSP envia ao DICT os verificadores de sincronismo de cada
tipo de chave registrada em sua base interna através de
mensagens “Reconciliação / Verificar Sincronismo”.
```
```
3 DICT Mensagem DICT recebe mensagem com os verificadores de
sincronismo do PSP com acesso direto.
```
```
4 DICT Ação
```
```
DICT identifica os verificadores de sincronismo gerados
para cada tipo de chave do PSP. Os verificadores do DICT
são então comparados com os verificadores do PSP.
```
```
5 DICT Mensagem
```
```
DICT envia resposta de confere (OK) ou não confere (NOK)
por tipo de chave ao PSP com acesso direto.
```
```
6 PSP com acesso
direto ao DICT
```
```
Mensagem PSP com acesso direto recebe resposta de conferência por
tipo de chave.
```
### 9.2 VERIFICAÇÃO DE VSYNC (PARTICIPANTE DO PIX COM ACESSO INDIRETO AO DICT)


**# Camada Tipo Descrição**

1

```
PSP com acesso
indireto ao DICT Ação^
```
```
PSP sem acesso direto acessa canal de comunicação com o
PSP com acesso direto ao DICT.
```
##### 2

```
PSP com acesso
indireto ao DICT Comunicação^
```
```
PSP sem acesso direto envia ao PSP com acesso direto os
verificadores de sincronismo de cada tipo de chave
registrada em sua base interna.
```
3 PSP com acesso
direto ao DICT

```
Comunicação PSP com acesso direto ao DICT recebe os verificadores de
sincronismo do PSP sem acesso direto.
```
4

```
PSP com acesso
direto ao DICT
```
```
Mensagem
```
```
PSP com acesso direto encaminha mensagem ao DICT com
os verificadores de sincronismo do PSP com acesso indireto
```

```
através de mensagens “Reconciliação / Verificar
Sincronismo”.
```
```
5 DICT Mensagem
```
```
DICT recebe mensagem com os verificadores de
sincronismo do PSP com acesso indireto.
```
```
6 DICT Ação
```
```
DICT verifica se o PSP com acesso direto tem autorização
para encaminhar solicitações para o PSP com acesso
indireto.
```
```
7 DICT Ação
```
```
DICT identifica os verificadores de sincronismo gerados
para cada tipo de chave do PSP com acesso indireto. Os
verificadores do DICT são então comparados com os
verificadores desse PSP.
```
```
8 DICT Mensagem DICT envia resposta de confere (OK) ou não confere (NOK)
por tipo de chave ao PSP com acesso direto.
```
```
9
```
```
PSP com acesso
direto ao DICT Mensagem^
```
```
PSP com acesso direto recebe resposta de conferência do
DICT.
```
```
10
```
```
PSP com acesso
direto ao DICT Comunicação^
```
```
PSP com acesso direto encaminha resposta de conferência
ao PSP com acesso indireto.
```
```
11
```
```
PSP com acesso
indireto ao DICT Comunicação^
```
```
PSP com acesso indireto recebe resposta de conferência
por tipo de chave.
```
### 9.3 LISTA DE CIDS

Para obtenção da lista de CIDs, é necessário que o participante faça o pedido para um tipo específico de
chave. Assim, sob demanda, o DICT gera a lista de CIDs de determinado tipo que estejam registrados
pelo participante solicitante.

A lista de CIDS é gerada de forma assíncrona. O participante com acesso direto faz a solicitação e o DICT
inicia o processo de geração, retornando um ID ao participante. O participante deve consultar o DICT
periodicamente para verificar o status da geração da lista. Assim que o processo é concluído, o DICT
altera o status da solicitação e disponibiliza o nome do arquivo e a url para realização do _download_. O
_download_ deverá ser feito via conexão HTTPS em um serviço provido no ambiente da Rede do Sistema
Financeiro Nacional (RSFN).

#### 9.3.1 Participante do Pix com acesso direto


```
# Camada Tipo Descrição
```
```
1 PSP com acesso
direto ao DICT
```
```
Ação PSP acessa canal de comunicação com o DICT.
```
```
2 PSP com acesso
direto ao DICT
```
```
Mensagem PSP solicita lista de suas chaves, por^ tipo específico,^ através
da mensagem “Reconciliação / Criar Arquivo de CIDs”.
```
```
3 DICT Mensagem DICT recebe mensagem com a solicitação.
```
```
4 DICT Ação
```
```
DICT gera lista com as chaves de tipo específico do PSP com
acesso direto.
```
```
5 DICT Ação
```
```
DICT altera o status da solicitação e disponibiliza o nome do
arquivo gerado e a url para realização do download.
```
##### 6

```
PSP com acesso
direto ao DICT
```
```
Ação
```
```
PSP com acesso direto identifica mudança no status e nome
do arquivo com lista de CIDs através da mensagem
“Reconciliação / Consultar Arquivo de CIDs”. PSP faz o
download por meio da url informada.
```
#### 9.3.2 Participante do Pix com acesso indireto


**# Camada Tipo Descrição**

1

```
PSP com acesso
indireto ao DICT Ação^
```
```
PSP com acesso indireto acessa canal de comunicação com
o PSP com acesso direto ao DICT.
```
##### 2

```
PSP com acesso
indireto ao DICT Comunicação^
```
```
PSP com acesso indireto solicita lista de suas chaves, por
tipo específico.
```
3 PSP com acesso
direto ao DICT

```
Comunicação PSP com acesso direto recebe solicitação^ do PSP com^
acesso indireto.
```
4 PSP com acesso
direto ao DICT

```
Mensagem
```
```
PSP com acesso direto encaminha solicitação para o DICT
através da mensagem “Reconciliação / Criar Arquivo de
CIDs”.
```
5 DICT Mensagem DICT recebe mensagem com a solicitação.


```
6 DICT Ação
```
```
DICT verifica se o PSP com acesso direto tem autorização
para encaminhar solicitações para o PSP com acesso
indireto.
```
```
7 DICT Ação
```
```
DICT gera lista com as chaves de tipo específico do PSP com
acesso indireto.
```
```
8 DICT Ação
```
```
DICT altera o status da solicitação e disponibiliza o nome do
arquivo gerado e a url para realização do download.
```
##### 9

```
PSP com acesso
direto ao DICT
```
```
Ação
```
```
PSP com acesso direto identifica mudança no status e nome
do arquivo com lista de CIDs através da mensagem
“Reconciliação / Consultar Arquivo de CIDs”. PSP faz o
download por meio da url informada.
```
```
10 PSP com acesso
direto ao DICT
```
```
Comunicação PSP com acesso direto encaminha lista de chaves de tipo
específico ao PSP com acesso indireto.
```
```
11 PSP com acesso
indireto ao DICT
```
```
Comunicação PSP com acesso indireto recebe lista de chaves de tipo
específico.
```
## 10 FLUXO DE NOTIFICAÇÃO DE INFRAÇÃO

A notificação de infração deve ser feita exclusivamente para transações Pix com fundada suspeita de
fraude^7. Ela não pode ser feita para transações realizadas no âmbito de outros arranjos de pagamento.

Na API do DICT, a notificação de infração pode ser criada por meio de dois _endpoints_ distintos:
“notificação de infração” e “marcação de fraude transacional”. No _endpoint_ “notificação de infração”,
devem ser reportadas as infrações para solicitação de devolução ou para cancelamento de devolução.
No _endpoint_ “marcação de fraude transacional” devem ser notificadas as infrações que têm como
objetivo único e exclusivo marcar CPFs, CNPJs e chaves de usuários envolvidos em fraudes relacionadas
a transações Pix e que sejam clientes do próprio participante que criar a marcação.

Com a evolução do Mecanismo Especial de Devolução, os participantes devem substituir a abertura da
notificação de infração associada à solicitação de devolução, ou o cancelamento de uma devolução, pela
instauração de uma Recuperação de Valores. Para mais informações sobre este ponto, consultar a seção
20 - Fluxo de Recuperação de Valores. A notificação de infração associada à solicitação de devolução
continuará existindo, mas ela será aberta pelo DICT durante o processo da Recuperação de Valores.

### 10.1 NOTIFICAÇÃO DE INFRAÇÃO PARA SOLICITAÇÃO DE DEVOLUÇÃO OU PARA CANCELAMENTO DE DEVOLUÇÃO

(^7) Para fins de notificação de infração, considera-se fraude qualquer transação Pix que (i) tenha sido iniciada ou
autorizada, inclusive no âmbito do Pix Automático, pelo usuário pagador em decorrência de golpe/estelionato
(incluindo-se golpes de engenharia social); (ii) tenha sido iniciada sem que o usuário pagador tenha autorizado a
transação por meio de autenticação digital; (iii) tenha sido iniciada por terceiro que teve acesso ao instrumento
de iniciação da transação e conseguiu autenticar e autorizar a transação, mas que não é reconhecida pelo usuário
pagador; ou (iv) tenha sido iniciada pelo usuário pagador mediante coerção ou extorsão.


Uma notificação de infração para solicitação de devolução ou para cancelamento de devolução pode ter
os seguintes estados no DICT:

- aberta: a notificação de infração foi aberta pelo PSP;
- recebida: a notificação de infração foi recebida pelo PSP;
- analisada/concluída/fechada: a notificação de infração foi analisada pelo PSP que a recebeu e o
    resultado dessa análise está disponível; ou
- cancelada: o PSP que abriu a notificação pode cancelar a notificação a qualquer tempo, inclusive
    após ela já ter sido analisada.

Uma notificação de infração para solicitação de devolução pode ser aberta apenas pelo PSP do pagador
e deve ser fechada pelo PSP do recebedor. Esse tipo de notificação deve ser usado nos casos em que o
PSP do pagador deseja recuperar os valores de uma transação Pix liquidada no SPI com fundada suspeita
de fraude. No momento da abertura de uma notificação desse tipo, deve-se necessariamente identificar
a transação a que se refere a notificação (por meio do seu _EndToEndID_ ). No contexto de uma
Recuperação de Valores, as notificações de infração serão abertas pelo DICT durante as etapas de
instauração e/ou bloqueio.

Uma notificação de infração para cancelamento de devolução pode ser aberta apenas pelo PSP do
recebedor da transação original e deve ser fechada pelo PSP do pagador da transação original. Esse tipo
de notificação deve ser usado nos casos em que os recursos financeiros de uma determinada transação
já tenham sido devolvidos, mas o PSP do recebedor da transação original deseja cancelar essa devolução
devido à suspeita de fraude recair sobre o usuário pagador. Ele pode ser utilizado apenas para
transações de devolução (pacs.004) liquidadas no SPI, sendo que a transação deve necessariamente ser
identificada (por meio do seu _RtrId_ ) no momento da abertura da notificação. A notificação de infração
para cancelamento de devolução não se aplica a devoluções por falha operacional do PSP do pagador.
Cancelamentos de devolução por fraude também poderão ser realizados por meio da instauração de
uma Recuperação de Valores.

Ao abrir uma notificação de infração para solicitação de devolução ou para cancelamento de devolução
no DICT, o PSP do pagador (ou o PSP com acesso direto ao DICT, ou o PSP do recebedor, conforme o
caso) deve informar os seguintes campos:

```
Campo Obrigatório ou
Facultativo
```
```
Descrição
```
```
Motivo para abertura da
notificação ( Reason )
```
```
Obrigatório
```
```
Domínios:
```
- solicitação de devolução
    ( _refund_request_ )
- cancelamento de devolução
    ( _refund_cancelled_ )

```
EndToEndID ou RtrId
( TransactionId ) Obrigatório^
```
```
Identificador da transação original
( EndToEndID ) ou identificador da
transação de devolução ( RtrId ), em
caso de cancelamento de devolução.
```
```
Causa da fraude
( SituationType )
```
```
Obrigatório
```
```
Domínios:
```
- golpe/estelionato ( _scam_ )
- transação não autorizada
    ( _account_takeover_ )


- crime de coerção ( _coercion_ )
- acesso e autorização fraudulenta
    ( _fraudulent_access_ )
- outros ( _other_ )

```
E-mail de contato
( Email )
```
```
Obrigatório quando a
notificação for aberta
pelo participante
```
```
E-mail de contato da área responsável
do participante que abre a notificação,
para obtenção de esclarecimentos
sobre a notificação de infração.
```
```
Telefone de contato
( Phone )
```
```
Obrigatório quando a
notificação for aberta
pelo participante
```
```
Telefone de contato da área
responsável do participante que abre a
notificação, para obtenção de
esclarecimentos sobre a notificação de
infração.
```
```
Comentários ( ReportDetails )
```
```
Obrigatório, caso
SituationType = “ other ”
```
```
Campo de texto livre, com informações
que possam auxiliar o outro PSP
envolvido na transação na análise da
notificação
```
Os domínios do campo “ _SituationType_ ” devem ser usados conforme explicado na nota de rodapé 6:

- _scam_ : transação que tenha sido iniciada pelo usuário pagador em decorrência de
    golpe/estelionato (golpes de engenharia social, de falso intermediário, de falso vendedor, etc.);
- _account_takeover_ : transação que tenha sido iniciada sem que o usuário pagador tenha
    autorizado a transação por meio de autenticação digital (transação iniciada por terceiro sem o
    uso de autenticação digital);
- _coercion_ : transação que tenha sido iniciada pelo usuário pagador mediante coerção ou extorsão
    (sequestro relâmpago, sequestro falso de conhecido etc.);
- _fraudulent_access_ : transação que tenha sido iniciada por terceiro que teve acesso à interface
    de iniciação da transação e conseguiu autenticar e autorizar a transação, mas que não é
    reconhecida pelo usuário pagador (golpes de engenharia social, como o do falso funcionário,
    que obtêm a senha do usuário pagador e iniciam transações autenticadas com senha obtida
    dessa forma); e
- _other_ : qualquer caso que não se encaixe de forma apropriada nos quatro casos anteriores.

Uma notificação de infração pode ser aberta e ser aceita inclusive nos casos em que o usuário recebedor
atua como intermediário de pagamentos^8 e o destinatário final da transação não é necessariamente
esse intermediário. Como o intermediário está identificado como usuário recebedor, ele deve ser
responsabilizado por eventuais fraudes e golpes cometidos pelo destinatário final da transação, que não

(^8) Define-se como intermediário de pagamentos o agente pessoa jurídica que estabelece uma relação jurídica com
um usuário final para a prestação ou uso de serviços que envolvam cobranças, pagamentos, recebimentos e outras
atividades similares por meio do Pix; e que possui conta em seu próprio nome, em um participante do Pix, para
receber recursos de um pagador em favor do destinatário final. Inclui-se, no conceito de intermediário de
pagamentos, _marketplaces_ nacionais e internacionais, agentes de coleta (ou agentes de cobrança ou agentes de
arrecadação), intermediadores de pagamentos, dentre outros. Os intermediários, em geral, detêm contas gráficas
(ou contas bolsão) dos seus clientes, que são os destinatários finais das transações Pix, e fazem a gestão desses
recursos e o posterior repasse para eles.


está explícito. Isso é válido para qualquer uma das causas da fraude constante no campo _SituationType_ ,
explicado acima.

Uma notificação de infração para cancelamento de devolução deverá ser sempre preenchida com
_SituationType_ = “ _other_ ”. Sempre que o campo “ _SituationType_ ” for preenchido com “ _other_ ”, o campo
“ _ReportDetails_ ” deverá ser preenchido com a explicação do caso em questão.

Os campos _“Email”_ e _“Phone”_ devem ser preenchidos sempre que possível. O objetivo é facilitar o
contato entre os PSPs para troca de informações e esclarecimentos sobre a suspeita de fraude em
questão.

Ao enviar mensagem para o DICT para fechar uma notificação de infração para solicitação de devolução
ou para cancelamento de devolução, o PSP do recebedor (ou o PSP com acesso direto ao DICT, ou o PSP
do pagador, conforme o caso) deve informar os seguintes campos:

```
Campo
```
```
Obrigatório ou
Facultativo Descrição^
```
```
Resultado da análise
( AnalysisResult )
```
```
Obrigatório
```
```
Possíveis resultados:
```
- aceita ( _agreed_ ); ou
- rejeitada ( _disagreed_ )

```
Tipo da fraude
( FraudType )
```
```
Obrigatório, caso
AnalysisResult = “ agreed ”
```
```
Domínios:
```
- falsidade ideológica, ou seja, o fraudador
    abriu a conta usada para aplicar a fraude
    usando documentos de outra pessoa
    ( _application_fraud_ )
- conta-laranja, ou seja, a conta usada para
    receber recursos de fraude foi aberta de
    forma legítima ( _mule_account_ )
- a conta usada para receber recursos de
    fraude estava no nome do próprio fraudador
    ( _scammer_account_ )
- outro ( _other_ )

```
Comentários
( AnalysisDetails )
```
```
Obrigatório, caso
FraudType = “ other ”
```
```
Campo de texto livre, com informações sobre a
análise da notificação de infração, incluindo
motivos para eventual rejeição da notificação.
```
Uma notificação de infração para solicitação de devolução que tenha sido confirmada gera
automaticamente uma marcação de fraude para o usuário recebedor da transação. Uma notificação de
infração para cancelamento de devolução que tenha sido confirmada gera automaticamente uma
marcação de fraude para o usuário pagador da transação original. Uma notificação de infração fechada
só pode ser cancelada pelo participante que a abriu. Se a notificação de infração fechada for cancelada,
a marcação de fraude gerada por ela também é cancelada automaticamente. O participante que
analisou a notificação de infração, caso volte atrás em sua decisão, poderá cancelar somente a marcação
de fraude, através do _endpoint CancelFraudMarker_.


Se o PSP do pagador cancelar uma notificação de infração para solicitação de devolução depois de ter
aberto a solicitação de devolução, assume-se que não houve fraude e, portanto, o PSP do pagador
deverá cancelar também a solicitação de devolução, se ela ainda estiver aberta. Se a solicitação de
devolução estiver fechada e tiver havido devolução, o PSP do pagador deverá devolver os recursos para
o PSP do recebedor através de uma nova transação Pix. Se o usuário pagador tiver agido de má fé, o PSP
do pagador deverá abrir uma notificação de infração para marcação de fraude transacional contra ele.

No fluxo de notificação de infração para abertura de solicitação de devolução, o PSP do pagador deve
abrir a notificação de infração no DICT imediatamente após a reclamação do usuário pagador (desde
que a transação tenha sido realizada nos últimos oitenta dias corridos). Após abrir a notificação, o PSP
do pagador deve fazer análise de consistência da reclamação do usuário pagador, para ter a certeza de
que o caso está dentro do escopo do Mecanismo Especial de Devolução de que trata a seção II do
capítulo XI do Regulamento do Pix. A análise do mérito da reclamação do usuário pagador também deve
ser feita concomitantemente ao período de sete dias disponibilizado para a análise do PSP do recebedor.
Caso, durante o período de análise, o PSP do pagador identifique que o pedido não deveria ter sido
aberto, por qualquer motivo, ele deve cancelar a notificação de infração. O cancelamento deve ser feito
mesmo que a notificação já tenha sido analisada e fechada pelo PSP do recebedor. Caso o PSP do
pagador deseje alterar ou complementar as informações da notificação de infração, poderá fazer isso
através do endpoint _updateInfractionReport_ desde que ela esteja nos estados "aberta" ou "recebida".

Após o PSP do recebedor ter analisado e concluído a notificação de infração, o PSP do pagador tem até
72 horas para iniciar a solicitação de devolução, caso o PSP do recebedor tenha aceitado a notificação.
Durante esse período, o PSP do pagador pode continuar fazendo a análise do mérito da abertura da
notificação. Caso o PSP do pagador identifique que o pedido não deveria ter sido aberto, por qualquer
motivo, ele deve cancelar a notificação de infração e não iniciar a solicitação de devolução, mesmo que
o PSP do recebedor tenha aceitado a notificação.

Sempre que uma notificação de infração for cancelada pelo PSP do pagador, o PSP do recebedor deve
desbloquear imediatamente os recursos bloqueados na conta do usuário recebedor e informá-lo.

#### com acesso direto ao DICT) 10.1.1 Fluxo de notificação de infração para abertura de solicitação de devolução (participantes do Pix

```
com acesso direto ao DICT)
```

```
# Camada Tipo Descrição
1 Usuário pagador Ação Usuário pagador identifica o problema.
```
```
2 Usuário pagador Ação
```
```
Usuário pagador acessa canal de atendimento e solicita
devolução do valor de uma determinada transação.
O valor solicitado para devolução pode ser igual ou menor
que o valor da transação original.
3 Usuário pagador Comunicação Usuário pagador envia solicitação de devolução.
4 PSP do pagador Comunicação PSP do pagador recebe solicitação de devolução.
```
```
5 PSP do pagador Ação
```
```
PSP do pagador recupera os dados da transação original e
verifica se a transação foi realizada nos últimos oitenta dias
corridos.
Caso a transação tenha sido realizada nos últimos oitenta
dias corridos, PSP do pagador segue para a etapa 8.
Caso a transação tenha sido realizada há mais de oitenta
dias corridos, PSP do pagador segue para a etapa 6.
```
```
6 PSP do pagador Comunicação
```
```
PSP do pagador informa usuário sobre rejeição da
solicitação.
7 Usuário pagador Comunicação Usuário pagador recebe informação e o fluxo é encerrado.
```
```
8 PSP do pagador Mensagem
```
PSP do pagador abre notificação de infração no DICT
através da mensagem “Notificação de Infração / Criar
Notificação de Infração”, com motivo “Solicitação de
devolução”.
9 DICT Mensagem DICT recebe mensagem com notificação de infração.
10 DICT Ação DICT disponibiliza notificação, com status “Aberta”.

11 PSP do recebedor Ação PSP consulta periodicamente no DICT a lista de
notificações de infração com estado “Aberta” e com


```
motivo “Vinculada a abertura de solicitação de
devolução”, através da mensagem “Notificação de
Infração / Listar Notificações de Infração”.
Assim que identificar uma notificação de infração com
estado “Aberta”, o PSP deve mudar seu estado para
“Recebida” através da mensagem “Notificação de Infração
/ Receber Notificação de Infração”.
```
12 PSP do recebedor Ação

```
PSP do recebedor bloqueia imediatamente o montante
total da transação original na conta do usuário recebedor.
Caso o montante disponível na conta do usuário
recebedor seja menor do que o valor da transação
original, o PSP do recebedor bloqueia o montante total
disponível.
Caso o PSP do pagador não inicie a solicitação de
devolução 72 horas após a conclusão do processo de
notificação de infração, o PSP do recebedor deverá
desbloquear os recursos.
```
13 PSP do recebedor Ação

```
PSP do recebedor analisa a notificação de infração. Ele
tem sete dias corridos para informar ao DICT o resultado
da análise.
Caso a notificação seja aceita, o fluxo segue para a etapa
15.
Caso a notificação não seja aceita, o fluxo segue para a
etapa 14.
```
14 PSP do recebedor Ação

```
PSP do recebedor desbloqueia recursos que haviam sido
bloqueados na conta do usuário recebedor.
```
15 PSP do recebedor Mensagem

PSP do recebedor envia mensagem para o DICT, alterando
o estado da notificação de infração para “Concluída”
através da mensagem “Notificação de Infração / Fechar
Notificação de Infração”.
16 DICT Mensagem DICT recebe mensagem.

17 DICT Ação DICT disponibiliza notificação de infração, com status
“Concluída”.

18 PSP do pagador Ação

```
PSP consulta periodicamente no DICT a lista de
notificações de infração com estado “Concluída” e com
motivo “Vinculada a abertura de solicitação de
devolução”, através das mensagens “Notificação de
Infração / Listar Notificações de Infração” ou “Notificação
de Infração / Consultar Notificação de Infração”.
```
#### com acesso indireto ao DICT) 10.1.2 Fluxo de notificação de infração para abertura de solicitação de devolução (participantes do Pix

```
com acesso indireto ao DICT)
```

**# Camada Tipo Descrição**
1 Usuário pagador Ação Usuário pagador identifica o problema.

2 Usuário pagador Ação

Usuário pagador acessa canal de atendimento e solicita
devolução do valor de uma determinada transação.
O valor solicitado para devolução pode ser igual ou menor
que o valor da transação original.
3 Usuário pagador Comunicação Usuário pagador envia solicitação de devolução.
4 PSP do pagador Comunicação PSP do pagador recebe solicitação de devolução.

5 PSP do pagador Ação

```
PSP do pagador recupera os dados da transação original e
verifica se a transação foi realizada nos últimos oitenta dias
corridos.
Caso a transação tenha sido realizada nos últimos oitenta
dias corridos, PSP do pagador segue para a etapa 8.
```

```
Caso a transação tenha sido realizada há mais de oitenta
dias corridos, PSP do pagador segue para a etapa 6.
```
```
6 PSP do pagador Comunicação PSP do pagador informa usuário sobre rejeição da
solicitação.
7 Usuário pagador Comunicação Usuário pagador recebe informação e o fluxo é encerrado.
```
```
8 PSP do pagador Comunicação
```
```
PSP do pagador envia solicitação de abertura de notificação
de infração para solicitação de devolução para PSP com
acesso direto (responsável ou liquidante do PSP do
pagador).
```
##### 9

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do pagador)
```
```
Comunicação PSP com acesso direto recebe comunicação.
```
##### 10

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do pagador)
```
```
Mensagem
```
```
PSP com acesso direto abre notificação de infração no DICT
através da mensagem “Notificação de Infração / Criar
Notificação de Infração”, com motivo “Solicitação de
devolução”.
```
11 DICT Mensagem DICT recebe mensagem com notificação de infração.

12 DICT Ação DICT disponibiliza solicitação, com status “Aberta”.

##### 13

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do recebedor)
```
```
Ação
```
```
PSP com acesso direto consulta periodicamente no DICT a
lista de notificações de infração com estado “Aberta” e com
motivo “Solicitação de devolução”, através da mensagem
“Notificação de Infração / Listar Notificações de Infração”.
Assim que identificar uma notificação de infração com
estado “Aberta”, o PSP com acesso direto deve mudar seu
estado para “Recebida” através da mensagem “Notificação
de Infração / Receber Notificação de Infração”.
```
##### 14

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do recebedor)
```
```
Comunicação PSP com acesso direto envia comunicação ao PSP do
recebedor.
```
15 PSP do recebedor Comunicação

```
PSP do recebedor recebe a solicitação do PSP com acesso
direto ao DICT.
```
16 PSP do recebedor Ação

```
PSP do recebedor bloqueia imediatamente o montante
total da transação original na conta do usuário recebedor.
Caso o montante disponível na conta do usuário
recebedor seja menor do que o valor da transação
original, o PSP do recebedor bloqueia o montante total
disponível.
Caso o PSP do pagador não inicie a solicitação de devolução
72 horas após a conclusão do processo de notificação de
```

```
infração, o PSP do recebedor deverá desbloquear os
recursos.
```
17 PSP do recebedor Ação

```
PSP do recebedor analisa a notificação de infração. Ele
tem sete dias corridos para informar ao PSP com acesso
direto o resultado da análise.
Caso a notificação seja aceita, o fluxo segue para a etapa
19.
Caso a notificação não seja aceita, o fluxo segue para a
etapa 18.
```
18 PSP do recebedor Ação

```
PSP do recebedor desbloqueia recursos que haviam sido
bloqueados na conta do usuário recebedor.
```
19 PSP do recebedor Comunicação

```
PSP do recebedor envia comunicação ao PSP com acesso
direto.
```
##### 20

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do recebedor)
```
```
Comunicação PSP com acesso direto recebe comunicação.
```
##### 21

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do recebedor)
```
```
Mensagem
```
```
PSP com acesso direto ao DICT envia mensagem para o
DICT, alterando o estado da notificação de infração para
“Concluída” através da mensagem “Notificação de Infração
/ Fechar Notificação de Infração”.
```
22 DICT Mensagem DICT recebe mensagem.

23 DICT Ação

```
DICT disponibiliza notificação de infração, com status
“Concluída”.
```
##### 24

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do pagador)
```
```
Ação
```
```
PSP com acesso direto consulta periodicamente no DICT a
lista de notificações de infração com estado “Concluída” e
com motivo “Solicitação de devolução”, através das
mensagens “Notificação de Infração / Listar Notificações
de Infração” ou “Notificação de Infração / Consultar
Notificação de Infração”.
```
##### 25

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do pagador)
```
```
Comunicação
```
```
PSP com acesso direto envia comunicação com o resultado
da notificação de infração.
```
26 PSP do pagador Comunicação PSP do pagador recebe comunicação.

#### acesso direto ao DICT 10.1.3 Fluxo de notificação de infração para cancelamento de devolução entre participantes do Pix com

```
acesso direto ao DICT
```

**# Camada Tipo Descrição**
1 Usuário recebedor Ação Usuário recebedor identifica o problema.

2 Usuário recebedor Ação

```
Usuário recebedor acessa canal de atendimento e solicita
cancelamento de uma determinada devolução.
```
3 Usuário recebedor Comunicação Usuário recebedor envia solicitação de cancelamento de
devolução.

4 PSP do recebedor Comunicação

```
PSP do recebedor recebe solicitação de cancelamento de
devolução.
```
5 PSP do recebedor Ação

```
PSP do recebedor recupera os dados da transação de
devolução e verifica se ela foi realizada nos últimos trinta
dias corridos.
Caso a transação de devolução tenha sido realizada nos
últimos trinta dias corridos, PSP do recebedor segue para a
etapa 8.
Caso a transação de devolução tenha sido realizada há mais
de trinta dias corridos, PSP do recebedor segue para a
etapa 6.
```
6 PSP do recebedor Comunicação

```
PSP do recebedor informa usuário sobre rejeição da
solicitação.
```
7 Usuário recebedor Comunicação

```
Usuário recebedor recebe informação e o fluxo é
encerrado.
```
8 PSP do recebedor Mensagem PSP do recebedor abre notificação de infração no DICT
através da mensagem “Notificação de Infração / Criar


Notificação de Infração”, do tipo “Cancelamento de
devolução”.
9 DICT Mensagem DICT recebe mensagem com notificação de infração.
10 DICT Ação DICT disponibiliza notificação, com status “Aberta”.

11 PSP do pagador Ação

```
PSP do pagador consulta periodicamente no DICT a lista de
notificações de infração com estado “Aberta” e com tipo
“Cancelamento de devolução”, através da mensagem
“Notificação de Infração / Listar Notificações de Infração”.
Assim que identificar uma notificação de infração com
estado “Aberta”, o PSP deve mudar seu estado para
“Recebida” através da mensagem “Notificação de Infração
/ Receber Notificação de Infração”.
```
12 PSP do pagador Ação

```
PSP do pagador bloqueia imediatamente o montante total
da transação de devolução na conta do usuário pagador.
Caso o montante disponível na conta do usuário pagador
seja menor do que o valor da transação de devolução, o PSP
do pagador bloqueia o montante total disponível.
Caso o PSP do recebedor não inicie a solicitação de
devolução 72 horas após a conclusão do processo de
notificação de infração, o PSP do pagador deverá
desbloquear os recursos.
```
13 PSP do pagador Ação

```
PSP do pagador analisa a notificação de infração. Ele tem
sete dias corridos para informar ao DICT o resultado da
análise.
Caso a notificação seja aceita, o fluxo segue para a etapa
15.
Caso a notificação não seja aceita, o fluxo segue para a
etapa 14.
```
14 PSP do pagador Ação

```
PSP do pagador desbloqueia recursos que haviam sido
bloqueados na conta do usuário pagador.
```
15 PSP do pagador Mensagem

PSP do pagador envia mensagem para o DICT, alterando o
estado da notificação de infração para “Concluída” através
da mensagem “Notificação de Infração / Fechar Notificação
de Infração”.
16 DICT Mensagem DICT recebe mensagem.

17 DICT Ação

```
DICT disponibiliza notificação de infração, com status
“Concluída”.
```
18 PSP do recebedor Ação

```
PSP do recebedor consulta periodicamente no DICT a lista
de notificações de infração com estado “Concluída” e com
tipo “Cancelamento de devolução”, através das mensagens
“Notificação de Infração / Listar Notificações de Infração”
ou “Notificação de Infração / Consultar Notificação de
Infração”.
```
#### acesso indireto ao DICT 10.1.4 Fluxo de notificação de infração para cancelamento de devolução entre participantes do Pix com

```
acesso indireto ao DICT
```

**# Camada Tipo Descrição**
1 Usuário recebedor Ação Usuário recebedor identifica o problema.

2 Usuário recebedor Ação

```
Usuário recebedor acessa canal de atendimento e solicita
cancelamento de uma determinada devolução.
```
3 Usuário recebedor Comunicação Usuário recebedor envia solicitação de cancelamento de
devolução.

4 PSP do recebedor Comunicação

```
PSP do recebedor recebe solicitação de cancelamento de
devolução.
```
5 PSP do recebedor Ação

```
PSP do recebedor recupera os dados da transação de
devolução e verifica se ela foi realizada nos últimos trinta
dias corridos.
```

```
Caso a transação de devolução tenha sido realizada nos
últimos trinta dias corridos, PSP do recebedor segue para
a etapa 8.
Caso a transação de devolução tenha sido realizada há
mais de trinta dias corridos, PSP do recebedor segue para
a etapa 6.
```
```
6 PSP do recebedor Comunicação PSP do recebedor informa usuário sobre rejeição da
solicitação.
```
```
7 Usuário recebedor Comunicação
```
```
Usuário recebedor recebe informação e o fluxo é
encerrado.
```
```
8 PSP do recebedor Comunicação
```
```
PSP do recebedor envia solicitação de abertura de
notificação de infração do tipo “Cancelamento de
devolução” para PSP com acesso direto (liquidante do PSP
do recebedor).
```
##### 9

```
PSP com acesso
direto ao DICT
(liquidante do PSP
do recebedor)
```
```
Comunicação PSP com acesso direto recebe comunicação.
```
##### 10

```
PSP com acesso
direto ao DICT
(liquidante do PSP
do recebedor)
```
```
Mensagem
```
PSP com acesso direto abre notificação de infração no
DICT através da mensagem “Notificação de Infração / Criar
Notificação de Infração”, do tipo “Cancelamento de
devolução”.
11 DICT Mensagem DICT recebe mensagem com notificação de infração.
12 DICT Ação DICT disponibiliza notificação, com status “Aberta”.

##### 13

```
PSP com acesso
direto ao DICT
(liquidante do PSP
do pagador)
```
```
Ação
```
```
PSP com acesso direto consulta periodicamente no DICT a
lista de notificações de infração com estado “Aberta” e
com tipo “Cancelamento de devolução”, através da
mensagem “Notificação de Infração / Listar Notificações
de Infração”. Assim que identificar uma notificação de
infração com estado “Aberta”, o PSP com acesso direto
deve mudar seu estado para “Recebida” através da
mensagem “Notificação de Infração / Receber Notificação
de Infração”.
```
##### 14

```
PSP com acesso
direto ao DICT
(liquidante do PSP
do pagador)
```
```
Comunicação PSP com acesso direto envia comunicação ao PSP do
pagador.
```
15 PSP do pagador Comunicação

```
PSP do pagador recebe a solicitação do PSP com acesso
direto ao DICT.
```
16 PSP do pagador Ação

```
PSP do pagador bloqueia imediatamente o montante total
da transação de devolução na conta do usuário pagador.
Caso o montante disponível na conta do usuário pagador
seja menor do que o valor da transação de devolução, o
PSP do pagador bloqueia o montante total disponível.
Caso o PSP do recebedor não inicie a solicitação de
devolução 72 horas após a conclusão do processo de
notificação de infração, o PSP do pagador deverá
desbloquear os recursos.
```

```
17 PSP do pagador Ação
```
```
PSP do pagador analisa a notificação de infração. Ele tem
sete dias corridos para informar ao DICT o resultado da
análise.
Caso a notificação seja aceita, o fluxo segue para a etapa
19.
Caso a notificação não seja aceita, o fluxo segue para a
etapa 18.
```
```
18 PSP do pagador Ação
```
```
PSP do pagador desbloqueia recursos que haviam sido
bloqueados na conta do usuário pagador.
```
```
19 PSP do pagador Comunicação
```
```
PSP do pagador envia comunicação ao PSP com acesso
direto.
```
##### 20

```
PSP com acesso
direto ao DICT
(liquidante do PSP
do pagador)
```
```
Comunicação PSP com acesso direto recebe comunicação.
```
##### 21

```
PSP com acesso
direto ao DICT
(liquidante do PSP
do pagador)
```
```
Mensagem
```
```
PSP com acesso direto ao DICT envia mensagem para o
DICT, alterando o estado da notificação de infração para
“Concluída” através da mensagem “Notificação de
Infração / Fechar Notificação de Infração”.
22 DICT Mensagem DICT recebe mensagem.
```
```
23 DICT Ação
```
```
DICT disponibiliza notificação de infração, com status
“Concluída”.
```
##### 24

```
PSP com acesso
direto ao DICT
(liquidante do PSP
do recebedor)
```
```
Ação
```
```
PSP com acesso direto consulta periodicamente no DICT a
lista de notificações de infração com estado “Concluída” e
com tipo “Cancelamento de devolução”, através das
mensagens “Notificação de Infração / Listar Notificações
de Infração” ou “Notificação de Infração / Consultar
Notificação de Infração”
```
##### 25

```
PSP com acesso
direto ao DICT
(liquidante do PSP
do recebedor)
```
```
Comunicação
```
```
PSP com acesso direto envia comunicação com o
resultado da notificação de infração.
```
```
26 PSP do recebedor Comunicação PSP do recebedor recebe comunicação.
```
### 10.2 NOTIFICAÇÃO DE INFRAÇÃO PARA MARCAÇÃO DE FRAUDE TRANSACIONAL

A notificação de infração para marcação de fraude transacional deve ser usada nos casos em que o PSP
deseja apenas marcar o CPF ou o CNPJ de um usuário que seja seu cliente e que esteja envolvido em
algum episódio de fraude relacionado a uma transação Pix específica. Ela precisa apenas ser criada pelo
participante do Pix através do _endpoint CreateFraudMarker_. Além disso, ela pode, a qualquer momento,
ser cancelada^9 pelo participante que a criou.

(^9) Uma marcação de fraude que tenha sido gerada automaticamente através de uma notificação de infração para
solicitação de devolução ou para cancelamento de devolução pode ser cancelada pelo PSP que fechou a
notificação, o qual possui relacionamento com o usuário que recebeu a marcação.


A marcação de fraude transacional pode ser criada por qualquer PSP envolvido em uma determinada
transação Pix. O PSP que cria a marcação deve ter certeza de que seu cliente se trata de um fraudador.
Sempre que a chave Pix do fraudador for conhecida, ela deve ser informada, para que a chave também
receba a marcação de fraude. Exemplos, não-exaustivos, de casos em que a marcação de fraude
transacional pode ser usada:

- fundada suspeita de fraude de uma transação Pix liquidada fora do SPI (entre clientes de um
    mesmo participante ou entre clientes de diferentes participantes que possuem um mesmo
    participante liquidante);
- fundada suspeita de fraude de uma transação Pix que foi rejeitada, tanto pelo PSP do pagador
    quando pelo PSP do recebedor;
- após um bloqueio cautelar, nos casos em que os recursos da transação original foram devolvidos
    pelo PSP do recebedor – nesse caso, quem abre a notificação é o PSP do recebedor;
- após a identificação de recursos oriundos de uma fraude pelo PSP do recebedor, havendo
    devolução dos recursos ou não, sem ter recebido uma notificação de infração para solicitação
    de devolução ou sem ter feito o bloqueio cautelar – nesse caso, quem abre a notificação é o PSP
    do recebedor;
- fundada suspeita de fraude de uma transação Pix liquidada no SPI, quando o PSP do pagador
    deseja apenas marcar o seu próprio usuário, não havendo a intenção de reaver os recursos.

Ao criar uma marcação de fraude transacional no DICT, o PSP deve informar os seguintes campos:

```
Campo Obrigatório ou Facultativo Descrição
```
```
Usuário ( TaxIdNumber ) Obrigatório CPF ou CNPJ do usuário com fundada
suspeita de fraude
```
```
Chave ( Key ) Facultativo
```
```
Chave Pix do usuário com fundada
suspeita de fraude (deve ser informada
sempre que for conhecida)
```
```
Tipo da fraude
( FraudType )
```
```
Obrigatório
```
```
Domínios:
```
- falsidade ideológica, ou seja, o
    fraudador abriu a conta usada para
    aplicar a fraude usando documentos
    de outra pessoa ( _application_fraud_ )
- conta-laranja, ou seja, a conta usada
    para receber recursos de fraude foi
    aberta de forma legítima
    ( _mule_account_ )
- a conta usada para receber recursos
    de fraude estava no nome do próprio
    fraudador ( _scammer_account_ )
- outro ( _other_ )

#### com acesso direto ao DICT 10.2.1 Fluxo de notificação de infração para marcação de fraude transacional entre participantes do Pix

```
Pix com acesso direto ao DICT
```

**# Camada Tipo Descrição**
1 PSP Ação PSP identifica problema.

2 PSP Mensagem

```
PSP cria marcação de fraude transacional no DICT através
da mensagem “Marcação de Fraude / Criar Marcação de
Fraude Transacional”.
```
3 DICT Mensagem

```
DICT recebe mensagem com marcação de fraude
transacional.
```
#### com acesso indireto ao DICT 10.2.2 Fluxo de notificação de infração para marcação de fraude transacional entre participantes do Pix

```
Pix com acesso indireto ao DICT
```

```
# Camada Tipo Descrição
1 PSP Ação PSP identifica problema.
```
```
2 PSP Comunicação
```
```
PSP envia marcação de fraude transacional para seu
liquidante no SPI.
```
##### 3

```
PSP com acesso
direto ao DICT
(liquidante no SPI
do PSP)
```
```
Comunicação
```
```
PSP com acesso direto recebe marcação de fraude
transacional.
```
##### 4

```
PSP com acesso
direto ao DICT
(liquidante no SPI
do PSP)
```
```
Mensagem
```
```
PSP com acesso direto cria marcação de fraude
transacional no DICT através da mensagem “Marcação de
Fraude / Criar Marcação de Fraude Transacional”.
```
```
5 DICT Mensagem
```
```
DICT recebe mensagem com marcação de fraude
transacional.
```
## 11 INTERFACE DE COMUNICAÇÃO

A interface do DICT proverá todas as funcionalidades para manutenção e obtenção de dados da conta
transacional vinculada a uma chave. Detalhes da implementação estão disponíveis na seção sobre o
DICT no Manual das Interfaces de Comunicação.


## 12 CACHE DE CHAVES CONSULTADAS

Para diminuir a carga de requisições ao DICT, as consultas a chaves vinculadas a usuários pessoa jurídica
podem ter suas respostas cacheadas no próprio participante, devendo-se seguir o prazo máximo de 180
segundos definido no header _Cache-Control_. Recomenda-se que o PSP não realize novas consultas no
DICT para vínculos salvos no cache, que estejam dentro do prazo de validade, sob o risco de ativar os
mecanismos de prevenção de _denial of service_ e ataque de leitura.

## 13 MECANISMOS DE PREVENÇÃO A ATAQUES DE LEITURA

Ataques de leitura são situações em que um usuário (ou um invasor utilizando a estrutura do
participante) realiza capturas de vínculos existentes no DICT para propósitos distintos de realização de
pagamentos. Os ataques de leitura podem ocorrer de forma rápida e intensa, em período de poucas
horas, ou de forma lenta, ao longo de dias ou meses.

A proteção eficaz dos dados constantes da base do DICT contra ataques de leitura exige uma atuação
conjunta entre o Banco Central do Brasil, como mantenedor da base de dados, e os participantes, com
acesso direto e indireto, como requisitantes de consultas e responsáveis por suas bases de dados
internas.

Nesta seção são abordados os mecanismos adotados pelo DICT, bem como os mecanismos mínimos que
devem ser adotados pelos participantes, conforme previsto no capítulo XIII, Seção V, do Regulamento
Pix.

### 13.1 MECANISMOS ADOTADOS PELO DICT

O DICT possui limitação de consultas de chave Pix baseada em dois indicadores, aplicáveis aos
participantes provedores de conta transacional:

```
a. consultas sem ordem de pagamento: quantidade de consultas que não resultaram em ordem
de pagamento; e
b. consultas inválidas: quantidade de consultas realizadas para chaves não registradas no DICT.
```
Para contabilização desses indicadores, o participante sempre deverá informar, no header _PI-PayerId_ , o
identificador do usuário pagador que está originando a consulta, o qual deverá ser o mesmo utilizado
na ordem de pagamento^10. Todas as consultas de um usuário final a partir de um determinado
participante devem possuir o mesmo identificador. Assim, o DICT irá controlar as consultas tanto para o
usuário final, quanto para o participante.

O mecanismo de limitação utiliza o algoritmo de _Token Bucket_^11 (balde de fichas ou balde de símbolos),
com as seguintes definições, aplicáveis à operação _getEntry_ :

(^10) O campo _PayerId_ deve ser preenchido com o CPF ou o CNPJ do usuário que está originando a consulta e que será
o responsável pelo pagamento da transação Pix, utilizando o formato numérico, sem pontos, traços ou barras, com
11 dígitos para CPF e 14 dígitos para CNPJ.
(^11) Tanenbaum, Andrew S. REDES DE COMPUTADORES. Rio de Janeiro: Editora Elsevier, 2003, p.428.


```
Usuário final
pessoa natural
```
```
Tamanho máximo do balde para consultas de telefone celular ou e-mail: 100
fichas
Tamanho máximo do balde para consultas de CPF, CNPJ ou chave aleatória:
100 fichas
Decréscimo: 1 ficha por consulta válida de qualquer chave e 20 fichas por
consulta inválida de qualquer chave, por balde
Acréscimo: 1 ficha por consulta de qualquer chave após o recebimento da
ordem de pagamento pelo SPI, por balde
Incremento temporal: 2 fichas a cada minuto em cada balde
Usuário final
pessoa jurídica
```
```
Tamanho máximo do balde para consultas de telefone celular ou e-mail:
1. 000 fichas
Tamanho máximo do balde para consultas de CPF, CNPJ ou chave aleatória:
1 .000 fichas
Decréscimo: 1 ficha por consulta válida de qualquer chave e 20 fichas por
consulta inválida de qualquer chave, por balde
Acréscimo: 2 fichas por consulta de qualquer chave após o recebimento da
ordem de pagamento pelo SPI, por balde
Incremento temporal: 20 fichas a cada minuto em cada balde
```
```
Obs.: Excepcionalmente, a critério do Banco Central do Brasil, os parâmetros
de um usuário podem ser alterados.
Participante Tamanho máximo do balde e incremento temporal: conforme categoria em
que o participante se enquadrar, conforme tabela abaixo^12 :
```
```
Categoria Tamanho Incremento / min
A^50 .000^25 .000^
B^40 .000^20 .0^00
C^30 .000^15 .000^
D^16 .000^ 8.000^
E^5.^000 2.500^
F^500 250
G^250 25
H^50 2
```
```
Decréscimo: 1 ficha por consulta válida de qualquer chave e 3 fichas por
consulta inválida de qualquer chave
Acréscimo: 1 ficha por consulta de qualquer chave após o recebimento da
ordem de pagamento pelo SPI
```
(^12) O Banco Central do Brasil enquadrará cada participante em uma das categorias. O participante poderá solicitar,
de forma devidamente fundamentada em dados históricos (e não em projeções), com a aprovação do diretor de
segurança cibernética do participante responsável (a que se refere a seção III do Capítulo VII do Regulamento do
Pix), quando houver, a alteração para uma categoria superior. Em casos de suspeita de ataque de leitura ou de
problema operacional, o participante poderá ter sua categoria alterada pelo Banco Central do Brasil, que
comunicará a mudança ao PSP.


Ao enviar uma requisição de consulta ao DICT, quando o balde já estiver vazio, o usuário ou participante
receberá um retorno de erro _RATE LIMITING_ (código HTTP 429), ficando impossibilitado de realizar novas
consultas, até que haja reposição de fichas. Por exemplo, caso um usuário pessoa natural, cujo balde
tenha apenas 5 fichas, após realizar uma requisição de consulta, receba um retorno de chave não
registrada no DICT ( _NOT FOUND_ – código HTTP 404), a qual possui penalidade de 20 fichas, seu balde
ficará com um saldo devedor de 15 fichas. Portanto, somente será possível a esse usuário realizar novas
consultas com sucesso após decorridos 8 minutos (reposição temporal de 2 fichas/minuto).

Nas transações iniciadas por meio do serviço de iniciação de transação de pagamento, o crédito de ficha
será realizado no balde do participante que está prestando serviço de iniciação e no balde do
participante pagador (detentor de conta transacional), caso este tenha feito uma consulta de chave
utilizando o mesmo código _EndToEndID_ gerado pelo iniciador.

O mecanismo de limitação da operação _getEntryStatistics_ possui os mesmos parâmetros de tamanho
máximo do balde e de incremento temporal aplicados à operação _getEntry_ para participantes, os quais
variam de acordo com a categoria em que o participante seja classificado.

### 13.2 MECANISMOS QUE DEVEM SER ADOTADOS PELOS PARTICIPANTES DO PIX

Os participantes têm papel fundamental na proteção dos dados do DICT contra ataques de leitura e
devem atuar de forma preventiva, colaborativa e complementar em relação aos mecanismos adotados
pelo Banco Central do Brasil, mencionados na subseção 13.1.

Para atingir esse objetivo e complementar a proteção realizada pelo Banco Central do Brasil, os
participantes, com acesso direto ou indireto ao DICT, devem manter em seus sistemas ao menos os
mecanismos de proteção discriminados nesta subseção.

#### 13.2.1 Verificação de autenticidade do usuário solicitante da consulta

Os participantes devem adotar mecanismos eficazes de validação dos seus clientes aptos a enviar ordens
de requisição ao DICT, de modo a não permitir o envio de consultas e o recebimento de informações da
base por pessoas não autorizadas. Os participantes devem garantir que os usuários que estão originando
a consulta, identificados por meio de seu CPF ou de seu CNPJ no header _PI-PayerId_ , são, de fato, clientes
que estão usando o serviço de consulta para o envio de um Pix.

Para isso, os participantes devem implementar práticas robustas de validação de seus clientes e de seus
dispositivos (por exemplo, positivação cadastral do cliente no momento da criação da conta, validação
do dispositivo, embarque de token no dispositivo, reconhecimento facial ou biométrico, MFA na
efetivação da transação, etc.). O participante pode, a seu critério, implementar as práticas dadas como
exemplo ou quaisquer outros controles que tornem a validação da autenticidade do usuário mais
robusta. A consulta às chaves Pix deve estar restrita a ambiente logado, cujo acesso deve se dar por
meio de autenticação, no mínimo, por login e por senha, além de o ambiente estar devidamente
autorizado para disponibilizar tal ação.

#### 13.2.2 Estabelecimento de política interna de limitação de consultas

Cada participante do Pix deve estabelecer uma política interna própria de limitação de consultas igual
ou mais restritiva do que a aplicada pelo mecanismo de _Token Bucket_ do DICT, de modo a prover uma


camada de proteção anterior, a qual deve limitar o envio excessivo de requisições e evitar o estouro do
balde dos seus usuários e do próprio participante, com consequente retorno de erro _RATE LIMITING_
(código HTTP 429).

O participante deve implementar um controle de requisições a nível de aplicação e/ou de infraestrutura
para não repassar ao DICT consultas de chaves quando o número de requisições do _bucket_ do usuário
ou do paticipante for atingido, prevenindo, assim, o retorno do erro _RATE LIMITING_. Por exemplo, se o
limite de consultas de um usuário é 20, o participante não deve enviar ao DICT a vigésima primeira
consulta. Caso essa limitação seja mais restritiva que a do DICT, o participante deve monitorar essas
consultas, para garantir que as requisições que possam exceder os limites internos não sejam enviadas
ao DICT.

#### 13.2.3 Monitoramento qualitativo e permanente das consultas

Os participantes devem implementar e manter mecanismos de acompanhamento das requisições feitas
ao DICT por solicitação de seus usuários/clientes. O monitoramento deve considerar períodos curtos
(horas) e longos (dias/meses), e ter como objetivo a identificação e o bloqueio de casos suspeitos de
ataque de leitura aos dados do DICT, utilizando como principais critérios as consultas excessivas que não
resultem em ordens de pagamento e as consultas excessivas de chaves que não estejam registradas no
DICT (erro _NOT FOUND_ – código HTTP 404).

Os participantes devem realizar o acompanhamento da relação entre o volume de consultas ao DICT
(VCD) e o envio de ordens ao SPI (EOS) numa escala de minutos ou, no máximo, poucas horas, assim
como em períodos mais longos de dias ou meses. Dado que alguns fatores afetam naturalmente o não
envio de ordens após uma consulta, é esperado que a relação VDC/EOS do participante seja ligeiramente
superior a 1. Contudo, valores muito superiores podem indicar comportamentos de ataque de leitura^13.
Cabe aos participantes analisar a sua jornada do usuário e os seus dados históricos de forma a
determinar a sua razão usual e os valores a partir dos quais uma situação anômala (ataque de leitura,
finalidade não permitida ou falha nos sistemas internos) pode ser caracterizada. Em caso de obtenção
de valores superiores, o participante deverá verificar eventuais falhas em seus sistemas, ou, se
necessário, adotar políticas mais restritivas em seus controles de limitação de requisições.

Os participantes devem, adicionalmente, monitorar a relação entre consultas de chaves não registradas
no DICT ( _NOT FOUND_ – código HTTP 404) e chaves encontradas ( _FOUND_ – código HTTP 200). Da mesma
forma como ocorre na razão VDC/EOS, os participantes devem estabelecer parâmetros de normalidade
para a relação _NOT FOUND/(NOT FOUND+FOUND)_^14 , tanto para o PSP como um todo como para cada
um dos seus usuários. Caso esses valores sejam superados, o participante deve verificar se o
comportamento anormal é justificado. Se verificada fundada suspeita e elevado indício de uso indevido
do DICT, o participante deve bloquear imediatamente o(s) usuário(s) em questão. Caso seja necessário
para determinado(s) usuário(s) realizar pagamentos em grande volume ou em lote, sugere-se ao

(^13) Tipicamente, valores próximos de 2 ou maiores que 2 são indicativos de anomalias, mas esse limiar pode variar
entre participantes, dependendo das características da jornada do usuário, de aspectos comportamentais do seu
público-alvo, do volume de agendamentos e da atuação do participante como detentor de conta em transações
iniciadas por meio de serviço de iniciação.
(^14) A relação _NOT FOUND/(NOT FOUND+FOUND)_ também é afetada pela jornada e por aspectos comportamentais.
Usualmente, percentuais superiores a 7% para o participante e superiores a 20 % para usuários que tenham
realizado no mínimo 100 consultas nas últimas 24 horas são bons indicativos de comportamentos suspeitos.


participante o uso da funcionalidade _checkKeys_ do DICT (de uso exclusivo do PSP, não podendo ser
disponibilizada, mesmo que de forma indireta, a seus usuários) para confirmar a existência de chaves,
previamente ao efetivo envio de requisições, de modo a evitar altas quantidades de consultas inválidas
( _NOT FOUND_ – código HTTP 404). Essa prática visa evitar a ocorrência de situações que possam indicar
suspeitas de ataque de leitura, bem como prevenir o esvaziamento dos baldes do usuário e do
participante.

O participante pode definir, a seu critério, o intervalo de tempo em que executará o monitoramento dos
dois indicadores e os parâmetros que serão utilizados para desencadear as ações de controle do possível
ataque.

Além dos controles relativos às relações VDC/EOS e _NOT FOUND_ /( _NOT FOUND_ + _FOUND_ ), citados nos
parágrafos anteriores, os participantes devem criar todos os controles adicionais que se façam
necessários para garantia da segurança, considerando os aspectos específicos do seu domínio de
negócio, da sua arquitetura tecnológica e do comportamento do seu público-alvo.

#### 13.2.4 Plano de ação para tratamento de casos suspeitos

Os participantes devem adotar plano de ação para tratamento de casos identificados como suspeitos de
ataque de leitura. As medidas previstas no plano de ação devem visar a imediata cessação do
comportamento anômalo, a preservação dos dados pessoais vinculados às chaves Pix e a
implementação de ações que evitem a sua reincidência.

Em caso de confirmação de ataque de leitura, o participante deve informar imediatamente o Banco
Central do Brasil sobre o caso, por meio do endereço eletrônico pix-operacional@bcb.gov.br,
descrevendo as medidas emergenciais adotadas para a contenção do problema e para a prevenção de
reincidência. O relato de incidente ao Banco Central do Brasil não dispensa o participante de tomar
outras ações previstas em lei, em especial relativas à Lei Geral de Proteção de Dados Pessoais (LGPD).

#### 13.2.5 Restrição dos dados da chave exibidos ao usuário que faz a consulta

Conforme descrito na seção 8.1 desse Manual, apenas algumas informações podem ser retornadas ao
usuário que esteja realizando a consulta (seja por meio do aplicativo, do _internet banking_ , dos sistemas
internos ou da API do PSP): nome completo ou nome empresarial do usuário que registrou a chave
consultada; título do estabelecimento (nome fantasia) do usuário que registrou a chave consultada, caso
esteja vinculada a um usuário pessoa jurídica; CPF mascarado (exemplo: ***.777.888-**) ou CNPJ do
usuário que registrou a chave consultada; chave consultada e nome do PSP do recebedor (opcional).

A disponibilização dessas informações deve ser feita com base nos requisitos de segurança descritos na
seção 6 do Manual de Segurança do Pix^15. As demais informações retornadas pelo DICT devem ser de
uso exclusivo do PSP do pagador, não podendo, em hipótese alguma, ser exibidas ao usuário que faz a
consulta.

(^15) O Manual de Segurança do Pix está disponível em
https://www.bcb.gov.br/content/estabilidadefinanceira/cedsfn/Manual_de_Seguranca_PIX.pdf


## 14 LIMITAÇÃO DE REQUISIÇÕES À API DO DICT

Cada operação da API está associada a uma política de limitação, conforme tabela abaixo^16. Na tabela,
o balde significa a margem que cada participante tem para eventualmente, durante um curto espaço de
tempo, exceder o limite estabelecido.

```
Política de rate limit Operações da API
```
```
Incremento do
balde
```
```
Tamanho máximo
do balde
entries.read getEntry Ver seção 13.1 Ver seção 13.1
```
```
entries.write createEntry, deleteEntry
```
```
800 requisições
/min
```
```
1. 200 requisições
```
```
entries.update updateEntry 6 00 requisições
/min
```
```
600 requisições
```
```
claims.read getClaim
```
```
6 00 requisições
/min
```
```
18. 000 requisições
```
```
claims.write
```
```
createClaim,
acknowledgeClaim,
cancelClaim, confirmClaim,
completeClaim
```
```
1.2 00 requisições
/min 36.^000 requisições^
```
```
claims.list-with-role-
filter
```
```
listClaims 40 requisições /min 200 requisições
```
```
claims.list-without-
role-filter listClaims^10 requisições /min^50 requisições^
sync-
verifications.write
```
```
createSyncVerification 10 requisições /min 50 requisições
```
```
cids-files.write createCidSetFile 40 requisições /dia 200 requisições
cids-files.read getCidSetFile 10 requisições /min 50 requisições
cids-events.list listCidSetEvents 40 requisições /min 200 requisições
```
```
cids-entries.read getEntryByCid 1.2^00 requisições
/min
```
```
36. 000 requisições
```
```
infraction-
reports.read getInfractionReport^
```
```
600 requisições
/min 18.^000 requisições^
```
```
infraction-
reports.write
```
```
createInfractionReport,
acknowledgeInfractionReport,
cancelInfractionReport,
closeInfractionReport
```
```
1.2 00 requisições
/min
```
```
36. 000 requisições
```
```
infraction-reports.list-
with-role-filter
```
```
listInfractionReports 40 requisições /min 200 requisições
```
```
infraction-reports.list-
without-role-filter listInfractionReports^ 10 requisições /min^ 50 requisições^
```
```
fraud-markers.write createFraudMarker
```
```
1.200 requisições
/min
```
```
36.000 requisições
```
(^16) A política de limitação da consulta de chaves (entries.read), cuja operação da API é getEntry, está detalhada na
seção 13.


```
keys.check^17 checkKeys 70 requisições/min 70 requisições^18
```
```
keys.check para
usuários finais^19 -^
```
```
300 chaves/dia (ou
mais restritivo, a
critério do PSP)
```
```
1.000 chaves (ou
mais restritivo, a
critério do PSP)
```
```
refunds.read getRefund
```
```
1.200 requisições
/min
```
```
36.000 requisições
```
```
refunds.write
```
```
createRefund, cancelRefund,
closeRefund
```
```
2.400 requisições
/min
```
```
72.000 requisições
```
```
refund_list_with_role listRefunds 40 requisições/min 200 requisições
refund_list_without
role
```
```
listRefunds 10 requisições/min 50 requisições
```
```
statistics_person_read getPersonStatistics
```
##### 12.000

```
requisições/min
```
```
36.000 requisições
```
```
statistics_entry_read getEntryStatistics
```
```
Mesmo incremento
do getEntry, por
categoria do
participante (ver
seção 13.1)
```
```
Mesmo tamanho do
getEntry, por
categoria do
participante (ver
seção 13.1)
policies_read getBucketState 60 requisições/min 200 requisições
policies_list listBucketStates 6 requisições/min 20 requisições
```
## 15 FLUXO DE VERIFICAÇÃO DE CHAVES PIX REGISTRADAS

### 15.1 FLUXO DE VERIFICAÇÃO DE CHAVES PIX REGISTRADAS PARA O PARTICIPANTE DO PIX COM ACESSO DIRETO AO DICT

(^17) Os mesmos limites são aplicados para o participante iniciador. Para a verificação de chaves Pix registradas, o
participante iniciador consulta o mesmo _endpoint_ que o participante provedor de conta transacional.
(^18) Podem ser consultadas no máximo 200 chaves em cada requisição.
(^19) O DICT não realiza o controle de verificação de chaves Pix registradas por usuário final. Os valores da tabela são
recomendações para que cada participante aplique internamente em relação a seus usuários.


```
# Camada Tipo Descrição
1 PSP Mensagem PSP envia lista de chaves Pix para o DICT.
2 DICT Mensagem DICT recebe lista do PSP.
3 DICT Ação DICT verifica se as chaves estão registradas.
```
```
4 DICT Mensagem
```
```
DICT envia mensagem, identificando se cada uma das
chaves da lista está registrada ou não.
5 PSP Mensagem PSP recebe mensagem.
```
```
6 PSP Ação PSP atualiza cache de existência de chave Pix, com base na
lista enviada pelo DICT.
```
### 15.2 FLUXO DE VERIFICAÇÃO DE CHAVES PIX REGISTRADAS PARA O PARTICIPANTE DO PIX COM ACESSO INDIRETO AO DICT


**# Camada Tipo Descrição**

1

```
PSP com acesso
indireto Comunicação^ PSP envia lista de chaves Pix para o PSP com acesso direto.^
```
2

```
PSP com acesso
direto
```
```
Comunicação PSP com acesso direto recebe lista.
```
3 PSP^ com acesso
direto

```
Mensagem PSP com acesso direto envia lista de chaves Pix para o DICT.
```
4 DICT Mensagem DICT recebe lista do PSP.
5 DICT Ação DICT verifica se as chaves estão registradas.

6 DICT Mensagem

```
DICT envia mensagem, identificando se cada uma das
chaves da lista está registrada ou não.
```
7

```
PSP com acesso
direto
```
```
Mensagem PSP com acesso direto recebe mensagem.
```
8 PSP com acesso
direto

```
Comunicação PSP com acesso direto envia lista para PSP com acesso
indireto.
```

##### 9

```
PSP com acesso
indireto
```
```
Comunicação PSP com acesso indireto recebe lista.
```
##### 10

```
PSP com acesso
indireto Ação^
```
```
PSP com acesso indireto atualiza cache de existência de
chave Pix, com base na lista enviada pelo DICT.
```
## 16 CACHE DE EXISTÊNCIA DE CHAVE PIX

Para permitir filtrar, em um conjunto de elementos passíveis de serem chaves Pix, quais estão
registrados no DICT, o PSP pode construir um _cache_ de existência de chave Pix. Nesse _cache_ , o PSP pode
registrar a situação da chave (existente ou inexistente), inclusive a situação de chaves registradas por
outros PSPs. Todos os tipos de chave podem ser consultados.

O _cache_ não deve servir de base para um serviço disponibilizado aos usuários, ele deve ser consumido
apenas pelo participante, para auxiliá-lo, por exemplo, no controle de balde de consultas ou para realizar
pagamentos em grande volume ou em lotes.

O _cache_ de existência de chave Pix pode ser alimentado a partir das seguintes fontes:
a. consultas de chave realizadas para fins de iniciação de um Pix, nos casos em que o PSP é o
próprio participante requerente ou nos casos em que ele está atuando como liquidante de outro
participante;
b. processos de portabilidade e de reivindicação de posse, nos casos em que o PSP é uma das
partes do processo ou nos casos em que ele está atuando como liquidante de outro participante
em uma das partes do processo;
c. a própria base de dados interna do PSP; e
d. verificação de chaves registradas no DICT ( _checkKeys_ - de uso exclusivo do PSP).

Registros no _cache_ feitos por meio das fontes “a” e “b” podem ser mantidos no _cache_ por até trinta dias.

Caso o registro no _cache_ seja feito por meio da fonte “d”, o registro deve seguir as diretivas contidas no
header _Cache-Control_ , para que ele seja válido.

O _cache_ pode armazenar apenas as seguintes informações:

- o _hash_ da chave (SHA-256, ou algoritmo mais robusto);
- a data de atualização da chave no _cache_ ; e
- a data de atualização da chave no DICT.

Portanto, não é permitido que o _cache_ armazene a própria chave em claro, os dados do portador da
chave e os dados da conta vinculada à chave.

## 17 FLUXO DE SOLICITAÇÃO DE DEVOLUÇÃO

Uma solicitação de devolução pode ter os seguintes estados no DICT:

- aberta: a devolução foi solicitada pelo PSP do usuário pagador;
- analisada/concluída/fechada: a solicitação de devolução foi analisada pelo PSP do usuário
    recebedor e o resultado dessa análise está disponível; ou
- cancelada: enquanto a solicitação não estiver no estado “analisada”, o PSP do pagador pode
    cancelar a solicitação.


Note-se que, ao contrário dos fluxos de portabilidade, de reivindicação de posse e de notificação de
infração, não é necessário o recebimento de uma solicitação de devolução.

A solicitação de devolução deve ser criada pelo PSP do usuário pagador nos casos de:

- falha operacional do PSP do pagador^20 ;
- fundada suspeita de fraude; ou
- erro do PSP do pagador no envio de uma ordem de pagamento referente ao Pix Automático^21 ;

Ela pode ocorrer por iniciativa do usuário pagador ou por iniciativa do próprio PSP^22. No caso de pedido
de cancelamento de devolução, a solicitação deve ser criada pelo PSP do recebedor.

Ao abrir uma solicitação de devolução no DICT, o PSP do pagador (ou o PSP com acesso direto ao DICT,
conforme o caso) deve informar os seguintes campos:

```
Campo Obrigatório ou Facultativo Descrição
```
```
EndToEndID ou RtrId
( TransactionId ) Obrigatório^
```
```
Identificador da transação original ou
identificador da transação de devolução, em
caso de cancelamento de devolução
```
```
Motivo
( RefundReason )
```
```
Obrigatório
```
```
Motivo pelo qual a transação deve ser
devolvida:
```
- falha operacional do PSP do pagador
    ( _operational_flaw_ )
- fundada suspeita de fraude ( _fraud_ )
- cancelamento de devolução
    ( _refund_cancelled_ )
- erro do PSP do pagador no envio de uma
    ordem de pagamento referente ao Pix
    Automático ( _pix_automatico_ )

```
Valor para devolução
( RefundAmount ) Obrigatório^ Valor a ser devolvido^
```
```
Comentários
( RefundDetails )
```
```
Facultativo / Obrigatório
```
```
Campo de texto livre, com informações que
possam auxiliar o PSP do recebedor na
análise da solicitação, incluindo,
eventualmente, canais de comunicação
entre as partes e formas de acesso a
documentos adicionais.
Obrigatório quando o
RefundReason=operational_flaw.
```
(^20) Não inclui erro operacional em transação de Pix Automático.
(^21) Possíveis erros são: ausência de autorização válida, inconsistência entre a instrução de pagamento e a
autorização ou erro operacional do PSP do pagador (por exemplo, envio de ordem de pagamento após o
agendamento ter sido cancelado).
(^22) Conforme previsto no Regulamento do Pix, em seu capítulo XI, a devolução pode ser iniciada pelo usuário
recebedor ou pelo Mecanismo Especial de Devolução, mesmo sem a solicitação de devolução pelo usuário pagador
ou pelo PSP do pagador.


Ao enviar mensagem para o DICT para fechar uma solicitação de devolução, o PSP do recebedor (ou o
PSP com acesso direto ao DICT, conforme o caso) deve informar os seguintes campos:

```
Campo Obrigatório ou
Facultativo
```
```
Descrição
```
```
Resultado da análise
( RefundAnalysisResult )
```
```
Obrigatório
```
```
Possíveis resultados:
```
- aceita total ( _totally_accepted_ )
- aceita parcial ( _partially_accepted_ )
- rejeitada ( _rejected_ )

```
Motivo para rejeição
( RefundRejectionReason )
```
```
Obrigatório, caso
resultado da análise for
igual a “rejeitada”
```
```
Domínios:
```
- falta de saldo na conta do cliente
    ( _no_balance_ )
- relacionamento com cliente encerrado
    ( _account_closure_ )
- solicitação inválida ( _invalid_request_ )
    (somente quando
    _RefundReason=operational_flaw_ )
- motivo genérico, que não se encaixe na
    descrição dos outros dois domínios
    ( _other_ )

```
Identificador da
transação de devolução
(RefundTransactionId)
```
```
Obrigatório, caso
resultado da análise for
igual a “aceita total” ou
“aceita parcial”
```
```
Identificador da transação na PACS.004 ou
na PACS.008 (caso o motivo da solicitação
de devolução seja pix_automatico ou
refund_cancelled )
```
```
Comentários
( RefundAnalysisDetails ) Facultativo^ / Obrigatório^
```
```
Campo de texto livre, com informações
sobre a análise da solicitação de devolução.
Obrigatório quando
RefundAnalysisResult = reject ,
RefundRejectionReason= invalid_request e
RefundReason=operational_flaw.
```
O resultado da análise deve ser identificado como “aceita total” nos casos em que a devolução dos
recursos corresponder exatamente ao valor da transação original. Caso a devolução dos recursos for em
montante inferior ao valor da transação original, o resultado da análise deve ser identificado como
“aceita parcial”. Esse caso pode ocorrer caso não haja recursos suficientes na conta do usuário
recebedor no momento da efetivação da devolução.

Caso tenha havido uma devolução parcial ou a solicitação tenha sido rejeitada (desde que a conta
transacional não tenha sido encerrada, pelo usuário ou pelo próprio PSP), o PSP do recebedor deve
monitorar a conta do usuário recebedor para que possa realizar outras devoluções parciais até o


atingimento do valor solicitado pelo PSP do pagador, sempre que houver recursos na conta^23. O
monitoramento e as devoluções parciais devem ocorrer dentro do prazo de 90 dias, contados a partir
da transação original. O PSP do recebedor poderá encerrar o monitoramento da conta do usuário
recebedor caso a devolução financeira tenha sido rejeitada pelo PSP do pagador ou caso a notificação
de infração para solicitação de devolução tenha sido cancelada pelo PSP do pagador. O PSP do pagador
também deve executar esse monitoramento em caso de devolução parcial ou de rejeição (desde que a
conta transacional não tenha sido encerrada, pelo usuário ou pelo próprio PSP), após uma solicitação
de cancelamento de devolução. Nos casos relacionados a transações de Pix Automático em que houver
erro do PSP do pagador no envio da ordem de pagamento, não há necessidade de monitoramento pelo
PSP do recebedor em caso de devolução parcial ou de rejeição da solicitação.

O resultado da análise deve ser identificado como “rejeitada” nos casos em que não há devolução de
recursos financeiros (ou seja, não houve a liquidação de uma pacs.004 ou de uma pacs.008, no caso de
solicitação de devolução com motivo “ _pix_automatico_ ” ou cancelamento de devolução) ou quando a
solicitação de devolução por falha operacional foi aberta indevidamente (ver seção Rejeição de
solicitação de devolução por solicitação inválida). O PSP pode rejeitar uma solicitação de devolução,
mesmo após ter aceitado a notificação de infração que a antecede^24 , caso (i) não haja recursos na conta
do cliente, (ii) o relacionamento com o cliente tenha sido encerrado, ou (iii) motivo genérico (que não
se encaixe na descrição dos outros dois domínios). Nos casos de rejeição, apesar de não haver devolução
de recursos financeiros, assegura-se que a chave, o CPF/CNPJ e a conta tenham sido devidamente
marcadas e associadas a transações fraudulentas.

No caso de uma solicitação de devolução aberta no contexto de uma Recuperação de Valores, o campo
_RefundAccount_ conterá os dados da conta para a qual os recursos deverão ser devolvidos.
Adicionalmente, o campo _MonitorAccount_ indicará se a conta do usuário recebedor deverá ser
monitorada em situações de devolução parcial ou de rejeição por ausência de saldo.

### 17.1 SOLICITAÇÃO DE DEVOLUÇÃO POR FALHA OPERACIONAL

No caso específico de solicitação de devolução por falha operacional, o PSP do recebedor deverá analisar
a solicitação de devolução e deverá rejeitá-la utilizando o motivo de rejeição “ _invalid_request”_ caso não
se trate de falha operacional.

Conforme disposto no Regulamento do Pix, não são considerados como falha operacional, para fins de
devolução, os casos em que a transação Pix foi devidamente iniciada pelo usuário pagador e o valor
indicado na iniciação da transação foi corretamente creditado na conta transacional do usuário

(^23) Caso a conta do usuário recebedor seja creditada durante o período de análise da notificação de infração ou
após o seu fechamento, no período que antecede a abertura da solicitação de devolução, os novos recursos devem
ser bloqueados e somados ao montante inicialmente bloqueado, até o valor total da transação original. Após a
abertura da solicitação de devolução, o novo valor a ser monitorado pelo PSP do recebedor deverá ser o valor
solicitado pelo PSP do pagador. A devolução deve ocorrer somente após o fechamento da notificação de infração
e a abertura da solicitação de devolução.
(^24) As solicitações de devolução relacionadas aos casos de falha operacional do PSP do pagador e aos casos de
erro do PSP do pagador envolvendo transações de Pix Automático não requerem a criação prévia de notificações
de infração.


recebedor. Em geral, considera-se falha operacional do PSP do pagador erros que causem dano
financeiro ao usuário pagador, como por exemplo: duplicidade da transação, transação realizada em
valor diferente do instruído pelo usuário ou envio de recursos sem que a transação tenha sido
confirmada pelo pagador.

Alguns exemplos de situações nas quais a solicitação de devolução por falha operacional não é aplicável:

- PSP não debitou a conta do usuário pagador para a efetivação do Pix;
- PSP debitou a conta do usuário pagador e não realizou o acerto contábil;
- Não exibir no extrato do usuário pagador o débito realizado;
- O usuário pagador indicou conta destino/chave Pix incorreta;
- O usuário pagador realizou duas ordens, quando gostaria de ter executado apenas uma;
- Qualquer caso de fraude ou golpe (estes casos devem seguir o fluxo completo do MED, com a
    abertura da notificação de infração, e o motivo da solicitação de devolução deve ser _fraud_ ou
    _refund_cancelled_ ).

Para que o PSP do recebedor possa analisar corretamente o caso, é recomendável que o PSP do pagador
forneça todas as informações possíveis na abertura da solicitação. Assim, sugere-se a inclusão das
seguintes informações no campo Comentários ( _details_ ):

- Telefone de contato da área responsável;
- E-mail de contato da área responsável;
- Falha ocorrida:
    o Duplicidade
    o Desacordo de valor
    o Transação não iniciada
    o Outros
- Breve relato do ocorrido.

Da mesma forma, em caso de rejeição da solicitação de devolução com o motivo “ _invalid_request”_ , o
PSP Recebedor deve incluir no campo Comentários _(RefundAnalysisDetails_ ) as hipóteses de não
cabimento da solicitação, conforme exemplificado acima. Caso o PSP Pagador não tenha enviado
informações suficientes para a verificação da falha, o PSP Recebedor deve indicar isso no campo
Comentários.

#### Pix com acesso direto ao DICT) 1.1.1. Fluxo de solicitação de devolução por “falha operacional do PSP do pagador” (participantes do

```
com acesso direto ao DICT)
```

**# Camada Tipo Descrição**

1 PSP do pagador Ação

```
PSP do pagador identifica problema e recupera os dados da
transação original.
```
2 PSP do pagador Mensagem

```
Caso a transação tenha sido realizada nos últimos noventa
dias, PSP do pagador envia solicitação de devolução para o
DICT através da mensagem “Devolução / Criar uma
solicitação de devolução”, com o valor a ser devolvido, o
motivo para a solicitação (“falha operacional”) e as
informações complementares (no campo “Comentários”)
para permitir a confirmação da falha operacional pelo PSP
do recebedor. O valor solicitado para devolução pode ser
igual ou menor que o valor da transação original.
```
3 DICT Mensagem DICT recebe mensagem com solicitação de devolução.

4 DICT Ação DICT disponibiliza solicitação, com status “Aberta”.

5 PSP do recebedor Ação

```
PSP do recebedor consulta periodicamente no DICT a lista
de solicitações de devolução com estado “Aberta” que se
referem a transações recebidas por seus clientes.
Assim que identificar uma solicitação de devolução com
estado “Aberta”, o PSP do recebedor avalia se é uma
situação de falha operacional. Se sim, ele verifica se
```

```
existem recursos disponíveis na conta do usuário
recebedor.
Caso existam recursos, PSP do recebedor segue para etapa
6.
Caso não se trate de falha operacional ou não existam
recursos, o fluxo segue diretamente para a etapa 9.
```
```
6 PSP do recebedor Ação
```
```
PSP do recebedor imediatamente inicia ordem de
devolução (envio da pacs.004). Após essa ação, o fluxo
segue concomitantemente para as etapas 7 e 9.
```
```
7 PSP do recebedor Comunicação
```
```
PSP do recebedor envia comunicação para seu usuário,
informando sobre a devolução.
```
```
8 Usuário recebedor Comunicação Usuário recebedor recebe comunicação do seu PSP.
```
```
9 PSP do recebedor Mensagem
```
```
PSP do recebedor envia mensagem para o DICT, alterando
o estado da solicitação para “Concluída”, com resultado
“Aceita total”, “Aceita parcial” ou “Rejeitada”, conforme o
caso.
Caso o resultado seja “Rejeitada”, o motivo para rejeição
deve ser identificado.
```
10 DICT Mensagem DICT recebe mensagem.

11 DICT Ação DICT disponibiliza solicitação, com status “Concluída”.

12 PSP do pagador Ação

```
PSP do pagador consulta periodicamente no DICT a lista de
solicitações de devolução com estado “Concluída” que se
referem a transações iniciadas por seus clientes.
```
#### Pix com acesso indireto ao DICT)................................................................................................................... 1.1.2. Fluxo de solicitação de devolução por “falha operacional do PSP do pagador” (participantes do

```
com acesso indireto ao DICT)
```

**# Camada Tipo Descrição**

1 PSP do pagador Ação PSP do pagador identifica problema e recupera os dados da
transação original.


```
2 PSP do pagador Comunicação
```
```
Caso a transação tenha sido realizada nos últimos noventa
dias, PSP do pagador envia solicitação de devolução para
PSP com acesso direto (responsável ou liquidante do PSP
do pagador), com o valor a ser devolvido, o motivo para a
solicitação (“falha operacional”) e as informações
complementares (no campo “Comentários”) para permitir
a confirmação da falha operacional pelo PSP do recebedor.
O valor solicitado para devolução pode ser igual ou menor
que o valor da transação original.
```
##### 3

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do pagador)
```
```
Comunicação
```
```
PSP com acesso direto recebe comunicação com solicitação
de devolução.
```
##### 4

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do pagador)
```
```
Mensagem
```
```
PSP com acesso direto envia solicitação de devolução ao
DICT através da mensagem “Devolução / Criar uma
solicitação de devolução”, com o valor a ser devolvido e o
motivo para a solicitação (“falha operacional”).
```
```
5 DICT Mensagem DICT recebe mensagem com solicitação de devolução.
```
```
6 DICT Ação DICT disponibiliza solicitação, com status “Aberta”.
```
##### 7

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do recebedor)
```
```
Ação
```
```
PSP com acesso direto (responsável ou liquidante do PSP
do recebedor) consulta periodicamente no DICT a lista de
solicitações de devolução com estado “Aberta” que se
referem a transações recebidas pelos clientes do PSP do
recebedor.
```
##### 8

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do recebedor)
```
```
Comunicação
```
```
Assim que identificar uma solicitação de devolução com
estado “Aberta”, o PSP com acesso direto comunica o PSP
do recebedor sobre a solicitação.
```
```
9 PSP do recebedor Comunicação PSP do recebedor recebe a solicitação de devolução.
```
10 PSP do recebedor Ação

```
PSP do recebedor avalia se é uma situação de falha
operacional. Se sim, ele consulta o saldo disponível.
Caso existam recursos na conta de seu cliente, PSP do
recebedor aceita a solicitação e segue para a etapa 11.
Caso não se trate de falha operacional ou não existam
recursos na conta de seu cliente, PSP do recebedor rejeita
a solicitação e segue diretamente para a etapa 13.
```
11 PSP do recebedor Comunicação

```
PSP do recebedor envia comunicação para seu usuário,
informando sobre a devolução.
```
12 Usuário recebedor Comunicação Usuário recebedor recebe comunicação do seu PSP.


13 PSP do recebedor Comunicação

```
PSP do recebedor envia comunicação ao PSP com acesso
direto ao DICT.
```
##### 14

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do recebedor)
```
```
Comunicação
```
```
PSP com acesso direto recebe comunicação sobre a
devolução e verifica se a devolução foi aceita.
```
##### 15

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do recebedor)
```
```
Mensagem
```
```
Caso a solicitação seja aceita, inicia a ordem de devolução
(envio da pacs.004) e envia mensagem para o DICT,
alterando o estado da solicitação para “Concluída”, com
resultado “Aceita total” ou “Aceita parcial”, a depender do
caso, após o processamento da pacs.004 ser finalizado com
sucesso.
Caso a solicitação seja rejeitada, envia mensagem para o
DICT, alterando o estado da solicitação para “Concluída”,
com resultado “Rejeitada” e identificando o motivo para
rejeição.
```
16 DICT Mensagem DICT recebe mensagem.

17 DICT Ação DICT disponibiliza solicitação, com status “Concluída”.

##### 18

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do pagador)
```
```
Ação
```
```
PSP com acesso direto consulta periodicamente no DICT a
lista de solicitações de devolução com estado “Concluída”
que se referem a transações iniciadas pelo PSP do pagador.
```
##### 19

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do pagador)
```
```
Comunicação PSP com acesso direto envia comunicação com o resultado
da solicitação de devolução.
```
20 PSP do pagador Comunicação

```
PSP do pagador recebe comunicação com o resultado da
solicitação de devolução.
```

### AO DICT) 17.2 FLUXO DE SOLICITAÇÃO DE DEVOLUÇÃO POR “FUNDADA SUSPEITA DE FRAUDE” (PARTICIPANTES DO PIX COM ACESSO DIRETO

### (participantes do Pix com acesso direto ao DICT)

**# Camada Tipo Descrição**

1 PSP do pagador Mensagem

```
Fluxo inicia após a conclusão do fluxo de notificação de
infração disposto na seção 10.3. O PSP do pagador tem até
72 horas para iniciar a solicitação de devolução após a
conclusão do processo de notificação de infração.
Caso a notificação de infração tenha sido aceita pelo PSP do
recebedor, PSP do pagador envia solicitação de devolução
para o DICT através da mensagem “Devolução / Criar uma
solicitação de devolução”, com o valor a ser devolvido e o
motivo para a solicitação (“fundada suspeita de fraude”). O
valor solicitado para devolução pode ser igual ou menor
que o valor da transação original.
```
2 DICT Mensagem DICT recebe mensagem com solicitação de devolução.

3 DICT Ação DICT disponibiliza solicitação, com status “Aberta”.

4 PSP do recebedor Ação

```
PSP do recebedor consulta periodicamente no DICT a lista
de solicitações de devolução com estado “Aberta” que se
referem a transações recebidas por seus clientes.
```

```
Assim que identificar uma solicitação de devolução com
estado “Aberta”, o PSP do recebedor verifica se existem
recursos disponíveis na conta do usuário recebedor.
Caso existam recursos, PSP do recebedor segue para etapa
5.
Caso não existam recursos, o fluxo segue diretamente para
a etapa 8.
```
```
5 PSP do recebedor Ação
```
```
PSP do recebedor imediatamente inicia ordem de
devolução (envio da pacs.004). Após essa ação, o fluxo
segue concomitantemente para as etapas 6 e 8.
```
```
6 PSP do recebedor Comunicação PSP do recebedor envia comunicação para seu usuário,
informando sobre a devolução.
```
```
7 Usuário recebedor Comunicação Usuário recebedor recebe comunicação do seu PSP.
```
```
8 PSP do recebedor Mensagem
```
```
PSP do recebedor envia mensagem para o DICT, alterando
o estado da solicitação para “Concluída”, com resultado
“Aceita total”, “Aceita parcial” ou “Rejeitada”, conforme o
caso.
Caso o resultado seja “Rejeitada”, o motivo para rejeição
deve ser identificado.
```
```
9 DICT Mensagem DICT recebe mensagem.
```
10 DICT Ação DICT disponibiliza solicitação, com status “Concluída”.

11 PSP do pagador Ação

```
PSP do pagador consulta periodicamente no DICT a lista de
solicitações de devolução com estado “Concluída” que se
referem a transações iniciadas por seus clientes.
```

### INDIRETO AO DICT) 17.3 FLUXO DE SOLICITAÇÃO DE DEVOLUÇÃO POR “FUNDADA SUSPEITA DE FRAUDE” (PARTICIPANTES DO PIX COM ACESSO

### (participantes do Pix com acesso indireto ao DICT)

**# Camada Tipo Descrição**

1 PSP do pagador Comunicação

```
Fluxo inicia após a conclusão do fluxo de notificação de
infração disposto na seção 10.4. O PSP do pagador tem até
72 horas para iniciar a solicitação de devolução após a
conclusão do processo de notificação de infração.
Caso a notificação de infração tenha sido aceita pelo PSP do
recebedor, PSP do pagador envia solicitação de devolução
para o PSP com acesso direto.
```

##### 2

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do pagador)
```
```
Comunicação PSP com acesso direto recebe comunicação.
```
##### 3

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do pagador)
```
```
Mensagem
```
```
PSP com acesso direto envia solicitação de devolução para
o DICT através da mensagem “Devolução / Criar uma
solicitação de devolução”, com o valor a ser devolvido e o
motivo para a solicitação (“fundada suspeita de fraude”). O
valor solicitado para devolução pode ser igual ou menor
que o valor da transação original.
4 DICT Mensagem DICT recebe mensagem com notificação de devolução.
```
```
5 DICT Ação DICT disponibiliza solicitação, com status “Aberta”.
```
##### 6

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do recebedor)
```
```
Ação
```
```
PSP com acesso direto consulta periodicamente no DICT a
lista de solicitações de devolução com estado “Aberta” que
se referem a transações recebidas por clientes do PSP do
recebedor.
```
##### 7

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do recebedor)
```
```
Comunicação
```
```
Assim que identificar uma solicitação de devolução com
estado “Aberta”, o PSP acesso direto comunica o PSP do
recebedor sobre a solicitação.
```
```
8 PSP do recebedor Comunicação
```
```
PSP do recebedor recebe a solicitação de devolução.
Caso existam recursos na conta de seu cliente, PSP do
recebedor aceita a solicitação e segue para a etapa 9.
Caso não existam recursos na conta de seu cliente, PSP do
recebedor rejeita a solicitação e segue diretamente para a
etapa 11.
```
```
9 PSP do recebedor Comunicação
```
```
PSP do recebedor envia comunicação para seu usuário,
informando sobre a devolução.
```
10 Usuário recebedor Comunicação Usuário recebedor recebe comunicação do seu PSP.

11 PSP do recebedor Comunicação

```
PSP do recebedor envia comunicação ao PSP com acesso
direto, informando sobre a devolução.
```
##### 12

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do recebedor)
```
```
Comunicação
```
```
PSP com acesso direto recebe comunicação sobre
devolução e verifica se a devolução foi aceita.
Caso a solicitação seja aceita, inicia a ordem de devolução
(envio da pacs.004) e segue para a etapa 13.
Caso a solicitação seja rejeitada, segue diretamente para a
etapa 13.
```
##### 13

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do recebedor)
```
```
Mensagem
```
```
PSP com acesso direto envia mensagem para o DICT,
alterando o estado da solicitação para “Concluída”, com
resultado “Aceita total”, “Aceita parcial” ou “Rejeitada”,
conforme o caso.
Caso o resultado seja “Rejeitada”, o motivo para rejeição
deve ser identificado.
```

```
14 DICT Mensagem DICT recebe mensagem.
```
```
15 DICT Ação DICT disponibiliza solicitação, com status “Concluída”.
```
##### 16

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do pagador)
```
```
Ação
```
```
PSP com acesso direto consulta periodicamente no DICT a
lista de solicitações de devolução com estado “Concluída”
que se referem a transações iniciadas por clientes do PSP
do pagador.
```
##### 17

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do pagador)
```
```
Comunicação PSP com acesso direto envia comunicação com o resultado
da solicitação de devolução.
```
```
18 PSP do pagador Comunicação
```
```
PSP do pagador recebe comunicação com o resultado da
solicitação de devolução.
```
### 17.4 FLUXO DE CANCELAMENTO DE DEVOLUÇÃO.....................................................................................................

Caso o usuário recebedor solicite ao seu PSP o cancelamento da devolução, o PSP do recebedor:

- caso tenha acesso direto ao DICT, inicia o fluxo de notificação de infração para cancelamento de
    devolução, detalhado na seção 10.1.3; ou
- caso tenha acesso indireto ao DICT, inicia o fluxo de notificação de infração para cancelamento
    de devolução, detalhado na seção 10.1.4.

Em ambos os casos, a notificação de infração deve ser aberta com motivo “cancelamento de devolução”
e a transação deve ser identificada ( _TransactionId_ ) por meio do identificador único ( _RtrId_ ) da pacs.004
que efetivou a devolução original.

Caso o PSP do pagador aceite a notificação de infração, o PSP do recebedor:

- caso tenha acesso direto ao DICT, inicia o fluxo de solicitação de devolução por “fundada
    suspeita de fraude”, detalhado na seção 1 7. 2 ; ou
- caso tenha acesso indireto ao DICT, inicia o fluxo de solicitação de devolução por “fundada
    suspeita de fraude”, detalhado na seção 1 7. 325.

Em ambos os casos, a solicitação de devolução deve ser aberta com motivo “cancelamento de
devolução” e a transação deve ser identificada ( _TransactionId_ ) por meio do identificador único ( _RtrId_ )
da pacs.004 que efetivou a devolução original.

Caso a solicitação de devolução seja aceita pelo PSP do pagador, ele deve iniciar uma nova transação
para cancelar a devolução. Ou seja, ele deve enviar uma pacs.008, uma vez que não é possível efetuar
uma devolução da devolução original por meio de uma pacs.004. Assim, para fechar a solicitação de
devolução, o PSP do pagador deve identificar o _EndToEndId_ dessa nova transação no campo

(^25) Em ambos os fluxos, o PSP do recebedor irá assumir o papel do PSP do pagador e o PSP do pagador irá assumir
o papel do PSP do recebedor.


“ _RefundTransactionId_ ”. Além disso, ele deve cancelar, imediatamente, a notificação de infração para
solicitação de devolução que ele criou para solicitar a devolução da transação original.

### AO PIX AUTOMÁTICO 17.5 FLUXO DE SOLICITAÇÃO DE DEVOLUÇÃO POR ERRO DO PSP DO PAGADOR NO ENVIO DE ORDEM DE PAGAMENTO REFERENTE

### ordem de pagamento referente ao Pix Automático

#### referente ao Pix Automático (participantes do Pix com acesso direto ao DICT) 1.1.3. Fluxo de solicitação de devolução por erro do PSP do pagador no envio de ordem de pagamento

PSP do pagador deverá, primeiramente, devolver o valor total da transação para o seu cliente utilizando
recursos próprios. Posteriormente, o PSP do pagador inicia o fluxo de solicitação de devolução no DICT
com o motivo “pix_automatico” para tentar ser ressarcido junto ao PSP do recebedor.

Como o usuário pagador já recebeu os recursos em sua conta, a devolução pelo PSP do recebedor, em
caso de saldo disponível na conta do usuário recebedor, deve ser realizada através de uma PACS.008,
com o campo finalidadeDaTransacao preenchido com “REFU”, e identificando como usuário recebedor
dessa transação o PSP do pagador da transação original. Para isso, o PSP do recebedor deverá
preencher na PACS.008 o número do CNPJ do PSP do pagador no campo de identificação do usuário
recebedor. O PSP do recebedor pode obter o CNPJ de cada participante do Pix na lista de participantes
do Pix publicada no sítio do BC^26.

#### referente ao Pix Automático (participantes do Pix com acesso indireto ao DICT) 1.1.4. Fluxo de solicitação de devolução por erro do PSP do pagador no envio de ordem de pagamento

```
referente ao Pix Automático (participantes do Pix com acesso direto ao DICT)
```
(^26) Disponível em https://www.bcb.gov.br/estabilidadefinanceira/participantespix


**# Camada Tipo Descrição**

1 PSP do pagador Ação

```
Fluxo inicia após reclamação do usuário pagador ou
identificação pelo próprio PSP do pagador do envio
indevido de uma transação de Pix Automático.
```

```
2 PSP do pagador Ação
```
```
PSP do pagador verifica se houve um erro seu no envio
da ordem de pagamento, seja por inconsistência entre a
instrução de pagamento enviada pelo PSP do recebedor
e os parâmetros da autorização concedida pelo usuário
pagador, por não haver uma autorização vigente
concedida pelo usuário pagador ou por um erro
operacional. Caso tenha havido erro, o fluxo segue
concomitantemente para as etapas 3 e 6.
```
```
3 PSP do pagador Ação
```
```
PSP do pagador devolve os recursos totais para seu
cliente, usando recursos próprios.
```
```
4 PSP do pagador Comunicação PSP do pagador envia comunicação ao seu usuário,
informando sobre a devolução.
5 Usuário pagador Comunicação Usuário pagador recebe comunicação do seu PSP.
```
```
6 PSP do pagador Mensagem
```
```
PSP do pagador envia solicitação de devolução para o
DICT através da mensagem “Devolução / Criar uma
solicitação de devolução”, com o valor a ser devolvido e
o motivo “pix_automatico”.
7 DICT Mensagem DICT recebe mensagem com solicitação de devolução.
```
```
8 DICT Ação DICT disponibiliza solicitação, com status “Aberta”.
```
```
9 PSP do recebedor Ação
```
```
PSP do recebedor consulta periodicamente no DICT a
lista de solicitações de devolução com estado “Aberta”
que se referem a transações recebidas por seus clientes.
Assim que identificar uma solicitação de devolução com
estado “Aberta”, o PSP do recebedor verifica se existem
recursos disponíveis na conta do usuário recebedor.
Caso existam recursos, PSP do recebedor segue para
etapa 10.
Caso não existam recursos, o fluxo segue diretamente
para a etapa 13.
```
10 PSP do recebedor Mensagem

```
PSP do recebedor inicia a devolução através de uma
pacs.008. Após essa ação, o fluxo segue
concomitantemente para as etapas 11 e 13.
```
11 PSP do recebedor Comunicação

```
PSP do recebedor envia comunicação para seu usuário,
informando sobre a devolução.
```
12 Usuário recebedor Comunicação Usuário recebedor recebe comunicação do seu PSP.

13 PSP do recebedor Mensagem

```
PSP do recebedor envia mensagem para o DICT,
alterando o estado da solicitação para “Concluída”, com
resultado “Aceita total”, “Aceita parcial” ou “Rejeitada”,
conforme o caso.
Caso o resultado seja “Rejeitada”, o motivo para rejeição
deve ser identificado.
```
14 DICT Mensagem DICT recebe mensagem.


```
15 DICT Ação DICT disponibiliza solicitação, com status “Concluída”.
```
```
16 PSP do pagador Ação
```
```
PSP do pagador consulta periodicamente no DICT a lista
de solicitações de devolução com estado “Concluída”
que se referem a transações iniciadas por seus clientes.
```
_1.1.4. Fluxo de solicitação de devolução por erro do PSP do pagador no envio de ordem de pagamento
referente ao Pix Automático (participantes do Pix com acesso indireto ao DICT)_


1 PSP do pagador Ação

```
Fluxo inicia após reclamação do usuário pagador ou
identificação pelo próprio PSP do pagador do envio
indevido de uma transação de Pix Automático.
```

```
2 PSP do pagador Ação
```
```
PSP do pagador verifica se houve um erro seu no envio
da ordem de pagamento, seja por inconsistência entre a
instrução de pagamento enviada pelo PSP do recebedor
e os parâmetros da autorização concedida pelo usuário
pagador, por não haver uma autorização vigente
concedida pelo usuário pagador ou por um erro
operacional. Caso tenha havido erro, o fluxo segue
concomitantemente para as etapas 3 e 6.
```
```
3 PSP do pagador Ação
```
```
PSP do pagador devolve os recursos totais para seu
cliente, usando recursos próprios.
```
```
4 PSP do pagador Comunicação PSP do pagador envia comunicação ao seu usuário,
informando sobre a devolução.
5 Usuário pagador Comunicação Usuário pagador recebe comunicação do seu PSP.
```
```
6 PSP do pagador Comunicação
```
```
PSP do pagador envia solicitação de devolução para o
PSP com acesso direto
```
##### 7

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do pagador)
```
```
Comunicação PSP com acesso direto recebe comunicação.
```
##### 8

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do pagador)
```
```
Mensagem
```
```
PSP com acesso direto envia solicitação de devolução
para o DICT através da mensagem “Devolução / Criar
uma solicitação de devolução”, com o valor a ser
devolvido e o motivo “pix_automatico”.
```
```
9 DICT Mensagem DICT recebe mensagem com solicitação de devolução.
```
10 DICT Ação DICT disponibiliza solicitação, com status “Aberta”.

##### 11

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do recebedor)
```
```
Ação
```
```
PSP com acesso direto consulta periodicamente no DICT
a lista de solicitações de devolução com estado “Aberta”
que se referem a transações recebidas por clientes do
PSP do recebedor.
```
##### 12

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do recebedor)
```
```
Comunicação
```
```
Assim que identificar uma solicitação de devolução com
estado “Aberta”, o PSP com acesso direto comunica o
PSP do recebedor sobre a solicitação.
```
13 PSP do recebedor Comunicação PSP do recebedor recebe a solicitação de devolução.

14 PSP do recebedor Ação PSP do recebedor verifica se existem recursos
disponíveis na conta do usuário recebedor.

15 PSP do recebedor Comunicação

```
PSP do recebedor comunica o PSP com acesso direto
sobre o resultado da solicitação de devolução.
```
##### 16

```
PSP com acesso
direto ao DICT
(responsável ou
```
```
Comunicação PSP com acesso direto recebe comunicação.
```

```
liquidante do PSP
do recebedor)
```
##### 17

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do recebedor)
```
```
Ação
```
```
PSP com acesso direto verifica o resultado da solicitação
de devolução informado pelo PSP do recebedor.
Caso a solicitação tenha sido aceita, segue para a etapa
18.
Caso a solicitação tenha sido rejeitada, segue
diretamente para a etapa 23.
```
##### 18

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do recebedor)
```
```
Mensagem
```
```
PSP com acesso direto inicia a devolução através de uma
pacs.008. Após essa ação, o fluxo segue
concomitantemente para as etapas 1 9 e 2 3.
```
##### 19

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do recebedor)
```
```
Comunicação PSP com acesso direto comunica PSP do recebedor sobre
a efetivação da devolução.
```
20 PSP do recebedor Comunicação PSP do recebedor recebe comunicação.

21 PSP do recebedor Comunicação

```
PSP do recebedor envia comunicação para seu usuário,
informando sobre a devolução.
```
22 Usuário recebedor Comunicação Usuário recebedor recebe comunicação do seu PSP.

##### 23

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do recebedor)
```
```
Mensagem
```
```
PSP do recebedor envia mensagem para o DICT,
alterando o estado da solicitação para “Concluída”, com
resultado “Aceita total”, “Aceita parcial” ou “Rejeitada”,
conforme o caso.
Caso o resultado seja “Rejeitada”, o motivo para rejeição
deve ser identificado.
```
24 DICT Mensagem DICT recebe mensagem.

25 DICT Ação DICT disponibiliza solicitação, com status “Concluída”.

##### 26

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do pagador)
```
```
Ação
```
```
PSP com acesso direto consulta periodicamente no DICT
a lista de solicitações de devolução com estado
“Concluída” que se referem a transações iniciadas por
clientes do PSP do pagador.
```
##### 27

```
PSP com acesso
direto ao DICT
(responsável ou
liquidante do PSP
do pagador)
```
```
Comunicação
```
```
PSP com acesso direto envia comunicação com o
resultado da solicitação de devolução.
```
28 PSP do pagador Comunicação PSP do pagador recebe comunicação com o resultado da
solicitação de devolução.


## 18 FLUXO DE CONSULTA A INFORMAÇÕES DE SEGURANÇA

A consulta a informações de segurança deve ser feita com o propósito de alimentar os mecanismos de
análise de fraude dos participantes. Ela pode ser utilizada inclusive em processos que não estejam
diretamente relacionados ao Pix.

O acesso ao _endpoint_ deve ser feito exclusivamente por iniciativa do próprio participante, sendo vedada
a disponibilização da funcionalidade para os usuários finais. Na API do DICT, o _endpoint_ é chamado de
“Estatísticas” ( _statistics_ ). A consulta de estatísticas pode estar vinculada a uma chave Pix
( _getEntryStatistics_ ) ou a um usuário ( _getPersonStatistics_ ).

As informações de segurança são atualizadas com um atraso máximo de 12 horas, contadas a partir da
ocorrência do último evento, cuja data é indicada pelo campo _watermark_.

A consulta a informações de segurança vinculadas a CPF ou CNPJ ( _getPersonStatistics_ ) pode ser realizada
a critério de cada participante.

A resposta do DICT conterá as seguintes informações:
a. quantidade de liquidações como recebedor no SPI (S _ettlements_ );
b. notificações de infração confirmadas e marcações de fraude transacional criadas, com fraude
do tipo “falsidade ideológica” ( _ApplicationFrauds_ );
c. notificações de infração confirmadas e marcações de fraude transacional criadas, com fraude
do tipo “conta laranja” ( _MuleAccounts_ );
d. notificações de infração confirmadas e marcações de fraude transacional criadas, com fraude
do tipo “conta fraudador” ( _ScammerAccounts_ );
e. notificações de infração confirmadas e marcações de fraude transacional criadas, com fraude
do tipo “outros” ( _OtherFrauds_ );
f. notificações de infração confirmadas sem identificação de tipo de fraude ( _UnknownFrauds_ )^27 ;
g. valor total das notificações confirmadas ( _TotalFraudTransactionAmount_ )^28 ;
h. quantidade de participantes distintos que confirmaram pelo menos uma notificação contra a
chave ou o CPF/CNPJ consultado ( _DistinctFraudReporters_ );
i. quantidade de notificações ainda não fechadas no momento da consulta ( _OpenReports_ );
j. quantidade de participantes distintos que possuem notificações ainda não fechadas no
momento da consulta ( _OpenReportsDistinctReporters_ );
k. notificações de infração rejeitadas totais, incluindo as notificações sem identificação do tipo de
fraude ( _RejectedReports_ ); e
l. quantidade de contas do CPF ou do CNPJ vinculadas a chaves Pix (quando um usuário for
consultado); ou quantidade de contas distintas às quais a chave esteve associada (quando uma
chave for consultada) ( _RegisteredAccounts_ ).

As informações dispostas nas letras “a” a “h” e “k” serão retornadas para os seguintes prazos:

(^27) Notificações abertas nas versões anteriores à versão 2.0 da API do DICT.
(^28) Não entram no cálculo, a nível de usuário, as notificações de infração com fraude do tipo “falsidade
ideológica”.


```
a. últimos 90 dias;
b. últimos 12 meses (sem considerar o mês no qual a consulta é realizada); e
c. últimos 60 meses (sem considerar o mês no qual a consulta é realizada).
```
Quando uma chave for consultada ( _getEntryStatistics_ ), as informações dispostas nas letras “a” a “l” serão
retornadas tanto para a chave consultada como para o CPF/CNPJ vinculado à chave consultada.

Quando um CPF ou um CNPJ for consultado ( _getPersonStatistics_ ), as informações dispostas nas letras
“a” a “l” serão retornadas apenas para o CPF/CNPJ consultado.

Para fins de contagem de prazos das notificações confirmadas, considera-se como data inicial a data em
que a notificação foi aceita.

Qualquer ação que possa resultar em modificações nas chaves (exclusão, alteração, portabilidade ou
reivindicação de posse) não remove nem invalida as informações de notificação de infração.

As informações de notificações de infração por chave e por usuário são independentes entre si. Assim,
caso uma chave seja excluída e, posteriormente, registrada, as informações das notificações de infração
serão herdadas pelo novo registro, desde que o conjunto de informações relativas à chave e ao usuário
sejam exatamente iguais ao conjunto de informações vinculadas à chave excluída.

As informações de segurança disponibilizadas pelo DICT devem ser utilizadas por cada participante do
Pix, a seu critério, em seus processos internos de gerenciamento de risco.

### 18.1 FLUXO DE CONSULTA A INFORMAÇÕES DE SEGURANÇA PARA O PARTICIPANTE DO PIX COM ACESSO DIRETO AO DICT

```
# Camada Tipo Descrição
```

```
1 PSP Ação
```
```
PSP identifica necessidade de consultar informações de
segurança
```
```
2 PSP Mensagem
```
```
PSP envia mensagem de consulta às informações de
segurança ao DICT, identificando a chave Pix ou o CPF/CNPJ
que se quer consultar.
3 DICT Mensagem DICT recebe mensagem com solicitação de consulta.
```
```
4 DICT Ação
```
```
DICT verifica se a instituição é autorizada a realizar
consultas.
```
```
5 DICT Ação DICT verifica se existem informações para a chave Pix ou
para o CPF/CNPJ consultado.
```
```
6 DICT Ação
```
```
Caso existam informações, DICT cria mensagem de
resposta, que deve conter as informações de segurança
vinculadas à chave Pix ou ao CPF/CNPJ consultado.
```
```
7 DICT Ação
```
```
Caso não existam informações, DICT cria mensagem de erro
informando a inexistência de informações para a chave Pix
ou para o CPF/CNPJ consultado.
```
```
8 DICT Mensagem DICT envia mensagem de resposta à consulta.
```
```
9 PSP Mensagem
```
```
PSP recebe mensagem com a resposta à consulta. Fim do
processo.
```
### 18.2 FLUXO DE CONSULTA A INFORMAÇÕES DE SEGURANÇA PARA O PARTICIPANTE DO PIX COM ACESSO INDIRETO AO DICT


**# Camada Tipo Descrição**

1

```
PSP com acesso
indireto
```
```
Ação
```
```
PSP com acesso indireto identifica necessidade de
consultar informações de segurança
```
##### 2

```
PSP com acesso
indireto
```
```
Comunicação
```
```
PSP com acesso indireto envia comunicação ao PSP com
acesso direto solicitando consulta às informações de
segurança, identificando a chave Pix ou o CPF/CNPJ que se
quer consultar.
```
3

```
PSP com acesso
direto Comunicação^ PSP com acesso direto recebe comunicação.^
```
4 PSP^ com acesso
direto

```
Mensagem
```
```
PSP com acesso direto envia mensagem de consulta às
informações de segurança ao DICT, identificando a chave
Pix ou o CPF/CNPJ que se quer consultar.
```
5 DICT Mensagem DICT recebe mensagem com solicitação de consulta.

6 DICT Ação

```
DICT verifica se a instituição é autorizada a realizar
consultas.
```
7 DICT Ação

```
DICT verifica se existem informações para a chave Pix ou
para o CPF/CNPJ consultado.
```
8 DICT Ação

```
Caso existam informações, DICT cria mensagem de
resposta, que deve conter as informações de segurança
vinculadas à chave Pix ou ao CPF/CNPJ consultado.
```

```
9 DICT Ação
```
```
Caso não existam informações, DICT cria mensagem de erro
informando a inexistência de informações para a chave Pix
ou para o CPF/CNPJ consultado.
```
```
10 DICT Mensagem DICT envia mensagem de resposta à consulta.
```
##### 11

```
PSP com acesso
direto
```
```
Mensagem
```
```
PSP com acesso direto recebe mensagem com a resposta à
consulta.
```
```
12 PSP com acesso
direto
```
```
Comunicação PSP com acesso direto envia comunicação ao PSP com
acesso indireto com a resposta à consulta.
```
```
13
```
```
PSP com acesso
indireto Comunicação^
```
```
PSP com acesso indireto recebe comunicação com a
resposta à consulta. Fim do processo.
```
## 19 CONSULTA DE BALDES

O DICT possui um _endpoint_ que permite aos participantes consultarem os seus limites de requisição à
API do DICT e o tamanho dos seus baldes no momento da consulta. Esse _endpoint_ tem como objetivo
permitir que os participantes façam uma melhor gestão sobre seus baldes, no intuito de evitar seu
esvaziamento. Os participantes podem realizar dois tipos de requisição:
a. listagem de políticas de limitação ( _policies_list_ ): possibilita a obtenção de uma lista contendo,
para cada política de requisições, a quantidade de fichas disponíveis no balde no momento da
consulta, a capacidade máxima do balde, a quantidade de reposição periódica de fichas, o
período de reposição de fichas e a categoria em que o participante está enquadrado (para o
balde de consulta de chaves); e
b. consulta de política de limitação ( _policies_read_ ): permite ao participante consultar, para uma
política de requisições individual, a quantidade de fichas disponíveis no balde no momento da
consulta, a capacidade máxima do balde, a quantidade de reposição periódica de fichas, o
período de reposição de fichas e a categoria em que o participante está enquadrado (para o
balde de consulta de chaves).

Na consulta de política de limitação, o participante deve informar como _input_ o nome da política de
requisições.


## 20 FLUXO DE RECUPERAÇÃO DE VALORES

A Recuperação de Valores é um aprimoramento do Mecanismo Especial de Devolução para casos
relacionados a fraudes e golpes, inclusive nas situações de cancelamento de devolução motivado por
fundada suspeita de fraude do usuário pagador^29. Esse processo permite rastrear e bloquear recursos
com suspeita de fraude não apenas na conta de destino da transação original realizada pela vítima, mas
também em transações subsequentes, aumentando as chances de recuperação. Isso é viabilizado por
meio do bloqueio de saldo em contas sob suspeita, análise pelos participantes envolvidos e posterior
devolução dos valores bloqueados diretamente para a conta da vítima, caso se conclua que a transação
foi de fato fraudulenta.

A instauração da Recuperação de Valores deve ser feita pelo prestador de serviço de pagamento do
usuário pagador sempre que houver identificação de conduta supostamente fraudulenta ou
recebimento de uma reclamação por parte do cliente. O participante que inicia o processo é
denominado **participante recuperador** , e a transação que dá origem à solicitação é chamada de
**transação raiz** , sempre tendo um cliente seu como pagador.

A abertura da Recuperação de Valores deve ocorrer de forma imediata, logo após a identificação da
suspeita de fraude ou o recebimento da reclamação por parte do usuário. A análise do mérito da
solicitação deve ser realizada após a instauração do processo, com o objetivo de agilizar o bloqueio dos
recursos nas contas sob suspeita.

Os campos de uma Recuperação de Valores são os seguintes:

```
Campo Obrigatoriedade Descrição
Tipo de fluxo ( FlowType ) Obrigatório Domínio:
```
- _Interactive_ : fluxo interativo
- _Automated_ : fluxo automatizado
ID da transação raiz
( _RootTransactionId_ )

```
Obrigatório Identificador da transação original
( EndToEndID ) ou identificador da
transação de devolução ( RtrId ), em caso
de cancelamento de devolução.
Causa da fraude
( SituationType )
```
```
Obrigatório Domínios:
```
- Scam: golpe/estelionato
- _Account_takeover_ : transação não
    autorizada
- _Coercion_ : crime de coerção
- _Fraudulent_access_ : acesso e
    autorização fraudulenta
- _Other_ : outros
Informação de contato
( _ContactInformation_ )

```
Obrigatório Campos E-mail ( Email ) e telefone
( Phone ) de contato da área responsável
do participante que abre a notificação,
```
(^29) A Recuperação de Valores para o cancelamento de devolução deverá ser feita da mesma forma que a
contestação de outra transação Pix qualquer. Na abertura da Recuperação de Valores o DICT não fará distinção
entre uma transação Pix normal ou uma transação de devolução.


```
para obtenção de esclarecimentos sobre
a Recuperação de Valores. Ambos são
obrigatórios.
Comentários ( ReportDetails ) Obrigatório, caso
SituationType = “OTHER”
```
```
Campo de texto livre, com informações
que possam auxiliar a análise da
notificação pelo outro PSP envolvido na
transação.
Parâmetros do Grafo de
Rastreamento
( TrackingGraphParameters )
```
```
Devem ser indicados na
instauração quando
FlowType =
”AUTOMATED” , e para
iniciar a etapa de
rastreamento quando
FlowType =
”INTERACTIVE”
```
```
Subcampos:
```
- _MinTransactionAmount_ : Valor
    mínimo da transação para ser
    incluída no grafo.
- _MaxTransactions_ : Número máximo
    de transações.
- _HopWindow_ : Janela máxima de
    tempo entre transação que entra em
    uma conta e transação que sai.
- _MaxHops_ : Profundidade máxima do
    grafo.

A profundidade do grafo de rastreamento é dada pela sua quantidade de camadas. A transação raiz é
considerada a primeira camada. Todas as transações realizadas pelo recebedor da transação raiz
compõem a segunda camada; as transações feitas pelos recebedores da segunda camada formam a
terceira camada, e assim sucessivamente.

As notificações de infração abertas no contexto da Recuperação de Valores não conterão dados de
contato. Os PSPs recebedores das notificações de infração poderão consultar as informações de contato
da área responsável do PSP pagador da transação raiz (participante recuperador) diretamente na
Recuperação de Valores.


### 20.1 REGRAS GERAIS

Uma Recuperação de Valores possui as seguintes etapas:

```
I - instauração : abertura do procedimento de Recuperação de Valores, solicitada pelo participante
recuperador;
II - rastreamento : seleção de um conjunto ordenado de transações a partir da transação raiz
contestada, denominado grafo de rastreamento;
III - priorização : seleção de um conjunto ordenado de transações do grafo de rastreamento que
representam o caminho mais provável de dispersão dos recursos oriundo da fraude ou golpe;
IV - bloqueio : abertura das notificações de infração pelo DICT e recebimento pelos prestadores de
serviço de pagamento das transações selecionadas na etapa de priorização, com imediato
bloqueio dos recursos solicitados;
V - análise : avaliação, pelos participantes notificados, quanto à procedência da suspeita de fraude;
VI - devolução : envio de pedidos de devolução aos participantes que identificaram contas de seus
clientes envolvidos na pulverização de recursos, além da efetiva devolução, em caso de
disponibilidade de saldo.
```
Além disso, o processo de Recuperação de Valores possui dois fluxos: interativo e automatizado. No
fluxo automatizado, as etapas de rastreamento, priorização e bloqueio ocorrem em sequência e sem
intervenção do participante após a instauração do processo. As etapas de análise e devolução ocorrem
de forma idêntica tanto no fluxo interativo quanto no automatizado. O processo completo é descrito
nas seções a seguir.

#### 20.1.1 Instauração no fluxo interativo

O participante recuperador inicia o processo através do _endpoint CreateFundsRecovery_ , informando que
deseja usar o fluxo de recuperação interativo. Nesse fluxo, o DICT envia imediatamente uma ordem de
bloqueio de fundos ao participante do usuário recebedor da transação raiz, por meio de uma notificação
de infração.

#### 20.1.2 Rastreamento no fluxo interativo

No próximo passo do fluxo interativo, chamado **etapa de rastreamento** , o participante deverá solicitar
um grafo de rastreamento ao DICT, através do _endpoint CreateTrackingGraph_ , informando os
parâmetros do grafo de rastreamento ( _TrackingGraphParameters_ ).

Quando houver um grafo que atenda aos parâmetros informados, o DICT responderá com um **grafo de
rastreamento**. Caso contrário, a resposta será vazia. Independentemente da resposta, o participante
recuperador pode solicitar novos rastreamentos utilizando parâmetros diferentes, tantas vezes quantas
forem necessárias.

#### 20.1.3 Priorização no fluxo interativo

Utilizando a resposta do DICT, o participante recuperador deverá definir uma **priorização de caminhos
de bloqueio** usando um algoritmo próprio, na **etapa de priorização**. O participante pode repetir a etapa
de rastreamento com parâmetros diferentes, se for necessário definir priorizações diferentes.


A priorização de caminhos de bloqueio é uma lista de transações, que deve atender pelo menos aos
seguintes critérios:
1) A transação raiz está incluída;
2) Todas as transações fazem parte do grafo de rastreamento;
3) Aciclicidade, ou não existência de ciclos^30 ;
4) Conectividade: todas as transações incluídas partem de contas que são recebedoras de pelo
menos uma transação incluída;

#### 20.1.4 Solicitação de bloqueio no fluxo interativo

Para iniciar a **etapa de solicitação de bloqueio** , o participante recuperador deve informar ao DICT a lista
de transações priorizadas por meio do _endpoint BlockFundsRecovery_. Esta lista de transações estará
sujeita a crítica pelo DICT com base nos critérios definidos na etapa anterior, de priorização. Falhar em
cumpri-los leva a um erro do tipo _FundsRecoveryPrioritizationInvalid_.

Havendo sucesso na priorização, prossegue-se à etapa de bloqueio, em que o DICT disponibilizará
notificações de infração aos PSPs recebedores das transações priorizadas, chamados neste processo de
**participantes recebedores.** Ao receberem as notificações, eles deverão bloquear imediatamente os
recursos correspondentes nas contas de seus usuários, no montante do valor informado no campo
_InfractionAmount_ , até o limite do saldo disponível. Caso não haja um valor preenchido, deve ser
considerado o valor total da transação. Eles também deverão efetuar novos bloqueios caso entrem
novos recursos e o valor bloqueado ainda seja inferior à soma dos valores a serem bloqueados
( _InfractionAmount_ ) das notificações de infração recebidas no processo de recuperação.

#### 20.1.5 Instauração no fluxo automatizado

Caso o participante recuperador opte pelo fluxo de recuperação automatizado por meio do _endpoint
CreateFundsRecovery_ , ele deverá informar também os parâmetros no _endpoint
TrackingGraphParameters_. O DICT utilizará esses parâmetros para construir um grafo de rastreamento,
sobre o qual será realizada a priorização dos caminhos de bloqueio, e disponibilizar as notificações de
infração, o que culminará na etapa de análise.

Os conceitos e restrições aplicáveis a todas as etapas são os mesmos do fluxo interativo. No entanto, no
fluxo automatizado, as etapas até o bloqueio são executadas sequencialmente pelo DICT, sem
necessidade de intervenção adicional do participante recuperador após a instauração do processo.

No fluxo automatizado, a notificação de infração da transação raiz é enviada conjuntamente com as
demais, se houver.

#### 20.1.6 Etapa de análise

(^30) Um grafo é acíclico quando não tem voltas ou caminhos que formam um círculo. Ou seja, ao seguir o caminho
do dinheiro — de quem enviou para quem — nunca se retorna ao ponto de origem passando por outros
intermediários.


Durante a **etapa de análise** , cada participante recebedor analisa as notificações de infração e avalia se
a suspeita de participação de seu cliente na dispersão dos recursos oriundos de fraude se confirma.
Mesmo não havendo saldo na conta, se o participante recebedor concluir que a transação fez parte do
esquema de dispersão do dinheiro da fraude, ele deve aceitar a notificação para permitir que as
transações posteriores sejam objeto da etapa de devolução subsequente.

A aceitação de uma notificação gera uma marcação de fraude ao usuário recebedor da transação
subjacente, independentemente de a devolução ser efetuada ou não. A etapa de análise se encerra
quando todos os participantes recebedores concluírem as respectivas análises das notificações de
infração recebidas. O participante recuperador também deve realizar a análise de mérito da reclamação
do usuário pagador e cancelar a Recuperação de Valores caso entenda de que não se trata de um caso
de fraude ou golpe.

Durante a etapa de análise, o participante recuperador deve consultar periodicamente o _endpoint
ListEventNotifications_ para verificar o encerramento dessa etapa, indicado pelo evento
_FUNDS_RECOVERY_ANALYSED_ da entidade _FUNDS_RECOVERY_. Todos os participantes que tiverem
aceitado as notificações de infração deverão aguardar o recebimento da solicitação de devolução, o
encerramento ou o cancelamento da Recuperação de Valores para desbloquear os recursos nas contas
de seus usuários.

#### 20.1.7 Etapa de devolução

Assim que a etapa de análise tiver sido cumprida, o participante recuperador tem até 72 horas para
iniciar a **etapa de devolução,** por meio do _endpoint RefundFundsRecovery_. Após iniciada, a etapa de
devolução não pode ser interrompida e a Recuperação de Valores não poderá mais ser cancelada.
Portanto, o participante recuperador deve analisar se a situação relatada pelo seu cliente se enquadra
no MED antes de dar início à devolução.

Nessa etapa, o DICT constrói o **grafo de devolução** com base nas transações cujas notificações de
infração foram aceitas e que estejam conectadas à transação raiz por meio de um caminho contínuo de
notificações também aceitas.

As solicitações de devolução são realizadas de forma sequencial, sendo cada uma enviada após a
conclusão da anterior, na ordem estabelecida pelo DICT. O valor de cada solicitação será,
cumulativamente, menor ou igual, a:
1) o valor a ser bloqueado ( _InfractionAmount_ ) da notificação de infração associada à transação
objeto da devolução;
2) o valor em aberto, dado pelo valor da transação raiz subtraído do valor total das devoluções
efetivamente realizadas até aquele momento no processo de Recuperação de Valores.

Todas as devoluções deverão ser efetivadas por meio de transações Pix em favor do usuário pagador da
transação raiz, até o limite do valor bloqueado. A devolução da transação raiz deve ser feita por meio
de uma mensagem pacs.004, exceto quando se tratar de um cancelamento de devolução, quando
deverá ser usada uma mensagem pacs.008 preenchida com os dados dos usuários pagador e recebedor
da transação raiz. Para as demais transações identificadas na etapa de rastreio (transações
subsequentes à transação raiz), os participantes recebedores devem debitar a conta de seus usuários e
enviar uma transação Pix, por meio de uma mensagem pacs.008, na qual deve figurar como pagador o


próprio PSP que efetua a devolução. Além disso, o campo finalidadeDaTransacao deve ser preenchido
com o valor “IPRT”.

Ao fechar a solicitação de devolução, o participante recebedor deve informar ao DICT o valor
efetivamente devolvido. Caso o PSP do recebedor da transação raiz seja o único participante a ter
recebido uma notificação de infração no processo de Recuperação de Valores, ele deverá realizar o
monitoramento da conta para a realização de novas devoluções, no caso de rejeição por ausência de
saldo ou de devolução parcial do valor solicitado, conforme disposto no Regulamento do Pix. O DICT
indicará, por meio do campo _MonitorAccount_ da solicitação de devolução, se o participante recebedor
deverá ou não realizar o monitoramento da conta.

#### 20.1.8 Desbloqueio dos recursos

O desbloqueio dos recursos pelo participante recebedor deve ocorrer apenas nos seguintes momentos:

- após rejeitar a notificação de infração;
- após concluir a solicitação de devolução; ou
- caso a Recuperação de Valores seja encerrada.

#### 20.1.9 Recuperação de valores para transações liquidadas nos sistemas dos participantes

É possível instaurar uma Recuperação de Valores para uma transação Pix liquidada fora do SPI, desde
que ela atenda aos parâmetros que permitem a geração do grafo de rastreamento. Caso a transação
seja identificada, o processo ocorre normalmente.

As operações liquidadas fora do SPI que estejam fora desses limites não são reconhecidas pelo processo
de Recuperação de Valores. Nesses casos, o participante deve conduzir a solicitação de devolução por
meio de um procedimento que não envolva o DICT, mas que ainda esteja em conformidade com as
regras do MED. No caso de uma transação dentro de um mesmo participante, ele deverá se encarregar
de bloquear o saldo, analisar a ocorrência de golpe ou fraude e processar a devolução conforme a sua
conclusão. No caso de transações entre participantes distintos, mas sob um mesmo liquidante, a
intermediação da devolução, seguindo as diretrizes estabelecidas pelo MED, ficará a cargo do
participante liquidante.


### 20.2 FLUXO DE INSTAURAÇÃO E SOLICITAÇÃO DE BLOQUEIO NO FLUXO INTERATIVO

Para fins de clareza, esse fluxo ilustra somente o caso para participantes diretos. No caso de
participantes que acessam o DICT de forma indireta, a comunicação é intermediada por um participante
direto, conforme acordado entre as partes.


```
# Camada Tipo Descrição
```
```
1 Usuário pagador Ação
```
```
Início do processo. Usuário pagador acessa canal de
atendimento e solicita abertura do MED referente a uma
determinada transação por motivo de fraude ou golpe.
O valor solicitado para devolução pode ser igual ou
menor que o valor da transação raiz.
2 Usuário pagador Comunicação Usuário pagador envia solicitação de abertura do MED.
3 PSP recuperador Comunicação PSP recuperador recebe solicitação de abertura do MED.
4 PSP recuperador Ação PSP recuperador recupera os dados da transação raiz.
```
```
5 PSP recuperador Ação
```
```
PSP recuperador verifica se a transação foi realizada nos
últimos oitenta dias corridos anteriores.
Se a transação tiver ocorrido há mais de oitenta dias
corridos, o PSP recuperador deve seguir para a etapa 6.
Caso tenha sido realizada nos últimos oitenta dias
corridos, deve seguir para a etapa 9.
6 PSP recuperador Ação PSP recuperador rejeita solicitação de devolução.
```
```
7 PSP recuperador Comunicação
```
```
PSP recuperador envia notificação, informando ao
usuário pagador sobre a rejeição da solicitação de
devolução.
```
```
8 Usuário pagador Comunicação
```
```
Usuário pagador recebe a notificação acerca da rejeição
da solicitação de devolução. Fim do processo.
```
```
9 PSP recuperador Mensagem
```
```
PSP recuperador envia mensagem solicitando a criação
de Recuperação de Valores no DICT através da
mensagem “Recuperação de Valores / Criar Recuperação
de Valores”, indicando a opção Fluxo Interativo.
Caso a transação tenha sido liquidada fora do SPI, como
no caso entre usuários com conta no mesmo participante
ou em participantes de um mesmo liquidante, e não se
enquadrar nos limites operacionais do grafo de
rastreamento, o PSP recuperador deverá resolver a
devolução internamente.
```
10 DICT Mensagem DICT recebe mensagem solicitando a criação de
Recuperação de Valores.

11 DICT Ação

```
DICT verifica se a transação consta no SPI ou no grafo de
rastreamento, cria a Recuperação de Valores com status
“CREATED”, e disponibiliza notificação de infração ao PSP
do recebedor da Transação Raiz, com status “Aberta”.
```
12 DICT Comunicação DICT comunica a criação da Recuperação de Valores ao
PSP recuperador.

13 PSP recuperador Comunicação

```
PSP recuperador recebe comunicação sobre a criação da
Recuperação de Valores.
```
14 PSP recuperador Mensagem

```
PSP recuperador envia ao DICT os parâmetros do grafo de
rastreamento através da mensagem “Recuperação de
Valores / Grafo de Rastreamento”.
```

15 DICT Comunicação

```
DICT recebe os parâmetros do grafo de rastreamento
enviados pelo PSP recuperador.
```
16 DICT Ação

```
DICT cria grafo de rastreamento a partir dos parâmetros
enviados pelo PSP recuperador.
```
17 DICT Comunicação

```
DICT disponibiliza o grafo de rastreamento com base nos
parâmetros informados.
```
18 PSP recuperador Comunicação PSP recuperador recebe grafo de rastreamento.

19 PSP recuperador Ação

```
PSP recuperador avalia grafo recebido. Caso queira
avançar com o grafo retornado, seguir para a etapa 20.
Se decidir pela geração de novo grafo de rastreamento,
deve retornar à etapa 14.
```
20 PSP recuperador Comunicação

```
PSP recuperador envia mensagem ao DICT com a
priorização de caminhos de bloqueio e solicitação de
bloqueio aos PSPs envolvidos, através da mensagem
“Recuperação de Valores / Solicitar Bloqueio”.
```
21 DICT Comunicação

```
DICT recebe mensagem com a priorização de caminhos
de bloqueio e solicitação de bloqueio aos PSPs
envolvidos.
```
22 DICT Ação

```
DICT avalia se priorização obedece aos critérios da
priorização de caminhos de bloqueio. Em caso negativo,
seguir para etapa 23. Se a priorização atender aos
critérios estabelecidos, seguir para etapa 25.
```
23 DICT Comunicação DICT envia comunicação ao PSP recuperador, informando
que a priorização não atende aos critérios do MED.

24 PSP recuperador Comunicação

```
PSP recuperador recebe a comunicação de que a
priorização não atende aos critérios do MED. Caso deseje
refazer o processo, o PSP recuperador deve retornar à
etapa 19, indicando novos parâmetros do grafo de
rastreamento, ou retornar à etapa 20, indicando nova
priorização de caminhos de bloqueio.
```
25 DICT Ação

```
DICT disponibiliza notificação de infração aos PSPs
recebedores das transações na priorização de caminhos
de bloqueio calculada pelo PSP recuperador e atualiza o
status da Recuperação de Valores para
WAITING_ANALYSIS.
```
26 DICT Comunicação

```
DICT envia mensagem ao PSP recuperador, indicando o
sucesso da transação.
```
27 PSP recuperador Comunicação

```
PSP recuperador recebe a mensagem de sucesso da
transação e todos os PSPs recebedores envolvidos no
rastreamento prosseguem para a etapa de análise. Fim
do processo.
```

### 20.3 FLUXO DE INSTAURAÇÃO E SOLICITAÇÃO DE BLOQUEIO NO FLUXO AUTOMATIZADO

Para fins de clareza, esse fluxo ilustra somente o caso para participantes diretos. No caso de
participantes que acessam o DICT de forma indireta, a comunicação com o DICT é intermediada por um
participante direto, conforme acordado entre as partes.


**# Camada Tipo Descrição**

1 Usuário pagador Ação

```
Início do processo. Usuário pagador acessa
canal de atendimento e solicita abertura do
MED referente a uma determinada transação
por motivo de fraude ou golpe.
O valor solicitado para devolução pode ser
igual ou menor que o valor da transação raiz.
```
2 Usuário pagador Comunicação

```
Usuário pagador envia solicitação de abertura
do MED.
```
3 PSP recuperador Comunicação PSP recuperador recebe solicitação de
abertura do MED.

4 PSP recuperador Ação

```
PSP recuperador recupera os dados da
transação raiz.
```
5 PSP recuperador Decisão

```
PSP recuperador verifica se a transação raiz foi
realizada nos últimos oitenta dias corridos
anteriores.
Se a transação tiver ocorrido há mais de
oitenta dias corridos, o PSP recuperador deve
seguir para a etapa 6.
Caso tenha sido realizada nos últimos oitenta
dias corridos, deve seguir para a etapa 9.
```
6 PSP recuperador Ação PSP recuperador rejeita solicitação de
devolução.

7 PSP recuperador Comunicação

```
PSP recuperador envia notificação, informando
ao usuário pagador sobre a rejeição da
solicitação de devolução.
```
8 Usuário pagador Comunicação

```
Usuário pagador recebe a notificação acerca da
rejeição da solicitação de devolução. Fim do
processo.
```
9 PSP recuperador Comunicação

```
Caso a transação tenha sido liquidada fora do
SPI, como no caso entre usuários com conta no
mesmo participante ou em participantes de
um mesmo liquidante e não se enquadrar nos
limites operacionais do GRAF, o PSP
recuperador deverá resolver a devolução
internamente, conforme explicado na
seção20.1.9, sobre Recuperação de Valores
para transações liquidadas nos sistemas dos
participantes.
```
```
Nos outros casos, o PSP recuperador envia
comunicação solicitando a criação de
Recuperação de Valores no DICT através da
mensagem “Recuperação de Valores / Criar
Recuperação de Valores”, indicando a opção
```

```
Fluxo Automatizado e informando os
Parâmetros do Grafo de Rastreamento.
```
10 DICT Comunicação

```
DICT recebe mensagem solicitando a criação
de Recuperação de Valores.
```
11 DICT Ação

```
DICT cria Recuperação de Valores com Status
“CREATED” e constrói grafo de rastreamento e
calcula a Priorização de Caminhos de Bloqueio.
```
12 DICT Ação

```
DICT disponibiliza notificações de Infração aos
PSPs recebedores das transações na
Priorização de Caminhos de Bloqueio
calculada, com status "Aberta" e atualiza o
Status da Recuperação de Valores para
“WAITING_ANALYSIS”.
```
13 DICT Mensagem

```
DICT envia mensagem ao PSP recuperador,
indicando o sucesso da transação.
```
14 PSP recuperador Comunicação

```
PSP recuperador recebe a mensagem de
sucesso da transação e todos prosseguem para
a etapa de análise. Fim do processo.
```

### 20.4 FLUXO DE ANÁLISE

Para fins de clareza, esse fluxo ilustra somente o caso para participantes diretos. No caso de
participantes que acessam o DICT de forma indireta, a comunicação com o DICT é intermediada por um
participante direto, conforme acordado entre as partes.

```
# Camada Tipo Descrição
```
```
1 DICT Ação
```
```
Início do processo. DICT disponibiliza notificações de
infração aos PSPs recebedores das transações na
priorização de caminhos de bloqueio calculada, com
status "Aberta".
```
```
2 PSP do recebedor Ação
```
```
PSP do recebedor consulta periodicamente no DICT
a lista de notificações de infração com estado
“Aberta” e com motivo “Solicitação de devolução”,
através da mensagem “Notificação de Infração /
Listar Notificações de Infração”.
```
```
3 PSP do recebedor Ação
```
```
PSP do recebedor identifica notificação de infração
com estado “Aberta” e altera seu estado para
“Recebida” através da mensagem “Notificação de
Infração / Receber Notificação de Infração”.
```

4 PSP do recebedor Ação

```
PSP do recebedor bloqueia imediatamente o
montante total da transação na conta do usuário
recebedor.
Caso o montante disponível na conta do usuário
recebedor seja menor do que o valor da transação
raiz, o PSP do recebedor bloqueia o montante total
disponível.
```
5 PSP do recebedor Decisão

```
PSP do recebedor analisa a notificação de infração.
Ele tem sete dias corridos para concluir a análise.
Caso a notificação não seja aceita, o fluxo segue para
a etapa 6.
Caso a notificação seja aceita, o fluxo segue para a
etapa 7.
```
6 PSP do recebedor Ação

```
PSP do recebedor desbloqueia recursos que haviam
sido bloqueados na conta do usuário recebedor.
```
7 PSP recebedor Mensagem

```
PSP do recebedor envia mensagem ao DICT,
alterando o estado da notificação de infração para
“Concluída” através da mensagem “Notificação de
Infração / Fechar Notificação de Infração”.
```
8 DICT Mensagem

```
DICT recebe mensagem, solicitando alteração do
status para “Concluída”.
```
9 DICT Ação

```
DICT disponibiliza notificação de infração, com
status “Concluída” e motivo “Solicitação de
devolução”. Fim do processo.
```

### 20.5 FLUXO DE DEVOLUÇÃO

Para fins de clareza, esse fluxo ilustra somente o caso para participantes diretos. No caso de
participantes que acessam o DICT de forma indireta, a comunicação com o DICT é intermediada por um
participante direto, conforme acordado entre as partes.

Da mesma forma, para fins de clareza na visualização do fluxo, o processo de devolução foi dividido em
duas etapas: etapa de devolução da transação raiz e etapa de devolução das transações seguintes.
Assim, os dois fluxos são complementares e o segundo se inicia imediatamente após a conclusão do
primeiro.


**_Etapa de devolução da transação raiz:_**


```
# Camada Tipo Descrição
```
```
1 PSP recuperador Ação
```
```
Início do processo. PSP recuperador identifica a
ocorrência do evento FUNDS_RECOVERY_ANALYSED
em consulta periódica ao endpoint de eventos do
DICT, referente à Recuperação de Valores
instaurada.
```
```
2 PSP recuperador Mensagem
```
```
PSP recuperador envia mensagem “Recuperação de
Valores/ Solicitar Devolução” ao DICT.
```
```
3 DICT Mensagem DICT recebe mensagem para iniciar devolução na
Recuperação de Valores.
```
```
4 DICT Ação
```
```
DICT identifica, dentre os PSPs que aceitaram as
notificações de infração, quais formam o maior grafo
conexo a partir da transação raiz, i.e., qual a maior
cadeia de transações com notificações de infração
aceitas que formam um caminho contínuo desde a
transação raiz.
```
```
5 DICT Decisão
```
```
DICT da transação raiz verifica se o grafo contém
alguma transação.
Se não contiver nenhuma transação, segue para a
etapa 6.
Se esse grafo contiver pelo menos uma transação,
segue para a etapa 10.
```
```
6 DICT Ação
```
```
DICT disponibiliza no endpoint de eventos a
conclusão da Recuperação de Valores.
```
```
7 PSP recuperador Ação
```
```
PSP recuperador identifica a ocorrência do evento
FUNDS_RECOVERY_COMPLETED em consulta
periódica ao endpoint de eventos do DICT, referente
à Recuperação de Valores instaurada.
```
```
8 PSP recuperador Mensagem PSP recuperador notifica o usuário pagador sobre a
finalização do processo de Recuperação de Valores.
```
```
9 Usuário pagador Comunicação
```
```
Usuário pagador recebe notificação sobre a
finalização do processo de Recuperação de Valores.
Fim do processo.
```
10 DICT Ação

```
DICT disponibiliza Solicitação de Devolução ao PSP
recebedor da transação raiz.
```
##### 11

```
PSP do recebedor
da transação raiz Ação^
```
```
PSP do recebedor da transação raiz identifica
solicitação de devolução com estado “Aberta” em
consulta periódica ao DICT, referente a transação
recebida pelo usuário recebedor da transação raiz.
```
12 PSP do recebedor
da transação raiz

```
Decisão
```
```
PSP do recebedor da transação raiz verifica se
existem recursos disponíveis na conta do usuário
recebedor.
Caso existam recursos, PSP do recebedor segue para
etapa 13.
```

```
Se não existirem recursos disponíveis na conta, o
fluxo segue diretamente para a etapa 16.
```
##### 13

```
PSP do recebedor
da transação raiz
```
```
Ação
```
```
PSP do recebedor da transação raiz envia ordem de
devolução (envio da pacs.004) imediatamente.
No caso de a transação raiz já ser uma PACS.004,
deve ser usada uma PACS.008.
```
14 PSP do recebedor
da transação raiz

```
Comunicação
```
```
PSP do recebedor da transação raiz envia
comunicação para o usuário recebedor da transação
raiz, informando sobre a devolução.
```
15 Usuário^ recebedor
da transação raiz

```
Comunicação Usuário recebedor da transação raiz recebe
comunicação do seu PSP. Fim do processo.
```
16 PSP do recebedor
da transação raiz

```
Mensagem
```
```
PSP do recebedor da transação raiz envia mensagem
para o DICT, alterando o estado da solicitação para
“Concluída”, com resultado “Aceita total”, “Aceita
parcial” ou “Rejeitada”, conforme o caso.
Caso o resultado seja “Rejeitada”, o motivo para
rejeição deve ser identificado.
```
17 DICT Mensagem DICT recebe mensagem de conclusão da devolução.

18 DICT Ação

```
DICT atualiza o saldo devolvido na Recuperação de
Valores. Fim do processo.
```

**_Etapa de devolução das transações seguintes:_**


```
# Camada Tipo Descrição
```
```
1 DICT Ação
```
```
DICT verifica se há saldo remanescente da
Recuperação de Valores e recebedores na
sequência, seguindo a priorização definida pelo
DICT. Em caso afirmativo, o fluxo segue para o passo
```
2. Se não houver saldo remanescente ou
recebedores na sequência, segue para o passo 11.

```
2 DICT Mensagem
```
```
DICT disponibiliza Solicitação de Devolução ao PSP
recebedor da transação seguinte.
```
##### 3

```
PSP do recebedor
da transação
seguinte
```
```
Ação
```
```
PSP do recebedor da transação seguinte identifica
solicitação de devolução com estado “Aberta” em
consulta periódica ao DICT, referente a transação
recebida pelo usuário recebedor da transação
seguinte.
```
##### 4

```
PSP do recebedor
da transação
seguinte
```
```
Decisão
```
```
PSP do recebedor da transação seguinte verifica se
existem recursos disponíveis na conta do usuário
recebedor.
Caso existam recursos, PSP do recebedor segue
para etapa 5.
Se não existirem recursos disponíveis na conta, o
fluxo segue diretamente para a etapa 8.
```
##### 5

```
PSP do recebedor
da transação
seguinte
```
```
Ação
```
```
PSP do recebedor da transação seguinte
imediatamente debita os recursos bloqueados da
conta de seu usuário e realiza devolução em favor do
usuário pagador da transação raiz, por meio de
mensagem pacs.008 em que o próprio PSP figura
como pagador (e não seu usuário).
```
##### 6

```
PSP do recebedor
da transação
seguinte
```
```
Comunicação
```
```
PSP do recebedor da transação seguinte envia
comunicação para seu usuário, informando sobre a
devolução.
```
##### 7

```
Usuário recebedor
da transação
seguinte
```
```
Comunicação
```
```
Usuário recebedor da transação seguinte recebe
comunicação do PSP do recebedor da transação
seguinte. Fim do processo.
```
##### 8

```
PSP do recebedor
da transação
seguinte
```
```
Mensagem
```
```
PSP do recebedor da transação seguinte envia
mensagem para o DICT, alterando o estado da
solicitação para “Concluída”, com resultado “Aceita
total”, “Aceita parcial” ou “Rejeitada”, conforme o
caso.
Caso o resultado seja “Rejeitada”, o motivo para
rejeição deve ser identificado.
9 DICT Mensagem DICT recebe mensagem de conclusão da devolução.
```
10 DICT Ação

```
DICT atualiza o saldo devolvido na Recuperação de
Valores e retorna para o passo 1.
```

11 DICT Ação

```
DICT disponibiliza no endpoint de eventos a
conclusão da Recuperação de Valores.
```
12 PSP recuperador Ação

```
PSP recuperador identifica a ocorrência do evento
FUNDS_RECOVERY_COMPLETED em consulta
periódica ao endpoint de eventos do DICT, referente
à Recuperação de Valores instaurada.
```
13 PSP recuperador Mensagem

```
PSP recuperador notifica o usuário pagador sobre a
finalização do processo de Recuperação de Valores.
```
14 Usuário pagador Comunicação

```
Usuário pagador recebe notificação sobre a
finalização do processo de Recuperação de Valores.
Fim do processo.
```

## 21 NOTIFICAÇÕES DE EVENTOS

O _endpoint_ de eventos é uma forma centralizada de consultar a ocorrência de eventos no DICT que
necessitem da ação de um PSP. Inicialmente, somente os eventos da Recuperação de Valores serão
consultados por esse canal.

Os parâmetros de entrada são os seguintes:

```
Campo Obrigatoriedade Descrição
ISPB do participante
( Participant )
```
```
Obrigatório ISPB do participante direto ou indireto
interessado
Cursor ( Cursor ) Opcional Usado para acessar uma paginação
específica do histórico de notificações
do participante.
```
O participante enviará uma requisição ao _endpoint ListEventNotifications_ , informando seu ISPB no
campo _Participant_. A resposta, descrita na API do DICT, informará a notificação mais antiga. Um valor
“ _true_ ” no campo _HasMoreElements_ indica que há outras notificações, e o campo _NextCursor_ fornece o
identificador da próxima. Para obtê-la, o PSP deve enviar a requisição preenchendo também o campo
_Cursor_ , com o identificador recebido.

Cada notificação tem os seguintes campos, que vêm sempre todos preenchidos:

```
Campo Descrição
Id da notificação ( Id ) Identificador daquele evento específico no DICT
Evento sendo notificado ( Event ) Domínios:
```
- FUNDS_RECOVERY_ANALYSED: indica que a etapa de
    análise da Recuperação de Valores, referente ao _EntityId_
    informado, foi concluída
- FUNDS_RECOVERY_COMPLETED, indica que o processo de
    Recuperação de Valores, referente ao _EntityId_ informado,
    foi finalizado.
Tipo de evento que foi notificado
( _EntityType_ )

```
Domínio:
```
- FUNDS_RECOVERY, quando se tratar de eventos
    relacionados ao processo de Recuperação de Valores
Id do tipo do evento ( _EntityId_ ) Identificador da entidade descrita em _EntityType_ , por
exemplo o _FundRecoveryId_ no caso de uma Recuperação de
Valores.
Momento da criação do evento pelo
DICT ( _Timestamp_ )

```
Apresentada no formato ISO8601.
```

## 22 HISTÓRICO DE REVISÃO

```
Data Versão Descrição das alterações
11/8/2020 1.0
10/9/2020 1.1 Seção 6: Ajustes na redação do fluxo de reivindicação, para deixar mais
claro seu funcionamento:
```
- caso o usuário doador não se manifeste dentro do período de
    resolução, o PSP doador deve necessariamente confirmar a
    reivindicação no DICT;
- no período de encerramento, o usuário doador pode somente validar
    a posse da chave, cancelando o processo. A confirmação não é
    possível durante esse período; e
- previsão de que o PSP reivindicador deve cancelar o processo de
    reivindicação no DICT caso seu usuário não faça a validação ativa da
    chave até o trigésimo dia após o início do processo de reivindicação.

```
Seção 6.1: ajustes nas etapas 5 e 7, para deixar a redação mais clara e
para incorporar os ajustes feitos ne seção 6.
```
```
Seção 6.2: ajustes nas etapas 7, 11, 12 e 13, para deixar a redação mais
clara e para incorporar os ajustes feitos ne seção 6.
```
```
Seção 6.3:
```
- ajuste no fluxo; e
- ajustes nas etapas 4, 12, 14, 15, 16, 17, 19 e 20, para deixar a redação
    mais clara e para incorporar os ajustes feitos ne seção 6.

```
Seção 6.4:
```
- ajuste no fluxo; e
- ajustes nas etapas 6, 16, 18, 20, 21, 22, 23, 27 e 28, para deixar a
    redação mais clara e para incorporar os ajustes feitos ne seção 6.

```
Seção 9: ajuste para deixar claro que a verificação de sincronismo não
precisa ser realizada diariamente. Ela precisa ser realizada em
intervalos máximos de 36 horas, conforme Manual de Tempos do Pix.
```
```
Seção 10: ajuste para prever que a notificação de infração pode ser
cancelada a qualquer tempo.
```
```
Seção 10.1: ajuste na nomenclatura das mensagens enviadas para o
DICT.
```
```
Seção 10.2: ajuste na nomenclatura das mensagens enviadas para o
DICT.
```
```
13/11/2020 2.0 Estrutura: inserção da seção 15 “Limitação de requisições à API do
DICT”.
```

```
Seção 5: inserção de nota de rodapé para explicitar que é possível
atualizar dados da conta vinculados à chave enquanto o status da
requisição de portabilidade estiver “Aberto” ou “Aguardando
Resolução”.
```
```
Seção 6: inserção de nota de rodapé para explicitar que é possível
atualizar dados da conta vinculados à chave enquanto o status da
requisição de reivindicação de posse estiver “Aberto” ou “Aguardando
Resolução”.
```
```
Seção 9: inserção de texto para detalhar como deve ser o processo de
correção de chave divergente após uma verificação de sincronismo.
```
```
Seção 10: retirada do campo “Motivo” no processo de abertura de uma
notificação de infração.
```
```
Seção 14: retirada das informações, que o DICT armazena, relativas a
transações com suspeita de infração à regulação de prevenção à
lavagem de dinheiro e ao financiamento do terrorismo.
```
17/11/2020 2.1 Seção 9: orientação para que eventuais divergências encontradas entre
a base interna e o DICT, após processo de verificação de sincronismo,
sejam corrigidas na base interna.

```
18/3/2021 3.0 Estrutura: inserção das seções 16 “Fluxo de verificação de chaves Pix
registradas” e 17 “Cache de existência de chave Pix”.
```
```
Seção 7.1: ajuste no fluxo e na tabela de passo a passo, para prever
possibilidade de alteração do nome do usuário vinculado à chave Pix.
```
```
Seção 7.2: ajuste no fluxo e na tabela de passo a passo, para prever
possibilidade de alteração do nome do usuário vinculado à chave Pix.
```
```
Seção 15: ajuste de forma e de texto na tabela que detalha a política de
rate limit , com a inclusão dos limites para o keys.read.
```
```
8/6/2021 4.0 Estrutura: inserção da seção 18 “Fluxo de solicitação de devolução”.
```
```
Estrutura: inserção das subseções 10.3 “Fluxo de notificação de infração
para abertura de solicitação de devolução (participantes do Pix com
acesso direto ao DICT)” e 10.4 “Fluxo de notificação de infração para
abertura de solicitação de devolução (participantes do Pix com acesso
indireto ao DICT)”.
```
```
Seção 5: previsão de possibilidade de cancelamento de uma
portabilidade com status “Confirmado” pelo PSP reivindicador.
```

```
Seção 10: inserção do campo “Motivo” e detalhamento dos campos na
abertura de uma notificação de infração; detalhamento dos campos no
fechamento de uma notificação de infração; e detalhamento do
funcionamento do fluxo de notificação de infração para abertura de
solicitação de devolução.
```
```
Seção 10.1: alteração do nome da seção para “Fluxo de notificação de
infração entre participantes do Pix com acesso direto ao DICT, por
motivo ‘fraude’”.
```
```
Seção 10.1: prazo máximo para abertura de notificação de infração
passa a ser oitenta dias corridos (etapa 2).
```
```
Seção 10.1: prazo máximo para análise de uma notificação de infração
passa a ser sete dias (etapa 7).
```
```
Seção 10.2: alteração do nome da seção para “Fluxo de notificação de
infração entre participantes do Pix com acesso indireto ao DICT, por
motivo ‘fraude’”.
```
```
Seção 10.2: prazo máximo para abertura de notificação de infração
passa a ser oitenta dias corridos (etapa 2).
```
```
Seção 10.2: prazo máximo para análise de uma notificação de infração
passa a ser sete dias (etapa 14).
```
```
Seção 13: tamanho máximo do balde do usuário final passa a ser 1.000
fichas, com incremento temporal de 2 fichas a cada minuto; tamanho
máximo do balde do participante passa a ser 20.000 fichas, com
incremento temporal de 6.000 fichas a cada minuto; e inserção de texto
para dar flexibilidade ao Banco Central do Brasil na gestão dos baldes.
```
```
Seção 15: ajustes na tabela com os limites de requisições à API do DICT.
```
29/6/2021 4.1 Seção 15: incorporação de novos limites de requisição à API do DICT.

22/7/2021 4.2 Estrutura: inserção das subseções 8.3 “Fluxo de consulta para o
participante do Pix que atua como prestador de serviço de iniciação de
transação de pagamento, com acesso direto ao DICT” e 8.4 “Fluxo de
consulta para o participante do Pix que atua como prestador de serviço
de iniciação de transação de pagamento, com acesso indireto ao DICT”.

```
Seção 10: inclusão das notas de rodapé 6 e 7, para deixar clara a data a
partir da qual os prazos relacionados à notificação de infração
começarão a valer.
```

```
Seção 13: inserção dos mecanismos de prevenção a ataques de leitura
para os participantes que prestem serviço de iniciação de transação de
pagamento.
```
Seção 15: inclusão da nota de rodapé 10, para deixar claro que os
mesmos limites para a verificação de chaves Pix registradas são
aplicáveis aos participantes iniciadores.
24/8/2021 4.3 Seção 10: inserção de nota de rodapé para deixar claro que a
notificação de infração para abertura de solicitação de devolução estará
disponível somente a partir de 16 de novembro de 2021, nos termos da
Resolução BCB nº 103.

```
Seção 13: ajustes nos mecanismos de prevenção a ataques de leitura do
DICT. As consultas sem liquidação para todos os tipos de chave passam
a consumir fichas nos baldes, tanto para os usuários finais quanto para
os participantes. Para isso, foi criado um novo balde, com 1.000 fichas,
para as chaves CPF, CNPJ e aleatória para os usuários finais; e foi
aumentado o incremento temporal de fichas para os participantes.
```
```
Seção 13: inserção de nota de rodapé para explicar as novas regras de
formação do campo PayerId.
```
```
Seção 15: criação de um balde específico para o endpoint updateEntry.
Com isso, o balde do createEntry e do deleteEntry foi diminuído.
```
```
Seção 18: ajuste em nota de rodapé, para deixar claro que a solicitação
de devolução estará disponível somente a partir de 16 de novembro de
2021, nos termos da Resolução BCB nº 103.
```
21/9/2021 4.4 Seção 10: inserção de texto para deixar mais claro o funcionamento da
funcionalidade.

```
Seção 10.3: inserção de notas de rodapé para deixar claro os casos em
que o Mecanismo Especial de Devolução não pode ser acionado.
```
```
Seção 10.4: inserção de notas de rodapé para deixar claro os casos em
que o Mecanismo Especial de Devolução não pode ser acionado.
```
```
Seção 13: ajustes nos mecanismos de prevenção a ataques de leitura do
DICT. Separação de baldes de consultas de usuários PF e PJ, com
definição de parâmetros diferenciados, através da identificação do tipo
de pessoa pelo campo PayerID. Além disso, os baldes de consultas de
participantes passam a ter categorias com parâmetros diferenciados de
tamanho e incremento, de forma a se adequar às necessidades de cada
participante.
```
```
Seção 13: ajuste na nota de rodapé 13, para deixar claro o formato a ser
usado no campo PayerId.
```

```
Seção 15: remoção da política geral entries.read e inclusão de nota de
rodapé, para explicar que essa política está sendo tratada com mais
detalhes na seção 13. Além disso, os parâmetros da política
update.entries foram reduzidos.
```
```
Seção 18: inserção de texto para deixar mais claro o funcionamento da
funcionalidade.
```
```
Seção 18.2: ajuste na etapa 15, para deixar o texto mais claro.
```
```
Seção 18.3: ajuste no fluxo, para consertar a etapa 4, que estava
identificando um estado de forma equivocada.
```
```
Seção 18.4: ajuste no fluxo, para consertar a etapa 6, que estava
identificando um estado de forma equivocada.
```
```
Seção 18.5: inserção de notas de rodapé, para deixar mais claro o
funcionamento da funcionalidade.
```
3/11/2021 5.0 Estrutura: inserção da seção 19 “Consulta a informações vinculadas às
chaves Pix para fins de segurança do Pix”.

```
Seção 8.1: alteração na etapa 7 do fluxo, para alterar a forma de
identificação do usuário pagador na consulta (usuário pagador deve ser
identificado por meio de seu CPF/CNPJ, e não mais por meio de um
identificador pseudonimizado.
```
```
Seção 8.1: alteração na etapa 9 do fluxo, para alterar a forma de
identificação do usuário pagador na consulta (usuário pagador deve ser
identificado por meio de seu CPF/CNPJ, e não mais por meio de um
identificador pseudonimizado).
```
```
Seção 10: ajustes no texto para prever os novos campos que permitirão
a notificação de infração para transações liquidadas fora do SPI e para
transações rejeitadas.
```
```
Seção 10.1: ajustes para explicar como deve ser a interpretação do
fluxo nos casos de transações liquidadas fora do SPI e de transações
rejeitadas.
```
```
Seção 10.2: ajustes para explicar como deve ser a interpretação do
fluxo nos casos de transações liquidadas fora do SPI e de transações
rejeitadas.
```
```
Seção 13: alteração na forma de identificação do usuário pagador na
consulta (usuário pagador deve ser identificado por meio de seu
CPF/CNPJ, e não mais por meio de um identificador pseudonimizado).
```

```
Seção 14: alteração nas informações para fins de segurança que são
retornadas pelo DICT sempre que uma chave é consultada.
```
19/11/2021 5.1 Seção 14: as informações para fins de segurança referentes a 3 dias
continuarão, provisoriamente, sendo apresentadas sempre que uma
chave é consultada.

```
Seção 15: inserção da limitação de requisições ao endpoint
“ statistics_read ”.
Seção 18: inserção de novo domínio no campo
“ RefundRejectionReason ”.
```
```
Seção 19: previsão de que informações sobre transações rejeitadas que
sofreram notificação de infração também serão retornadas na consulta
a informações vinculadas às chaves Pix.
```
```
12/1/2022 5.2 Seção 10: ajuste no texto para prever que, em transações “INTERNAL”
em que o PSP do pagador e o PSP do recebedor possuem um mesmo
liquidante, quem fecha a notificação, concordando ou discordando, é a
contraparte que não abriu a notificação.
```
```
Seção 16.1: ajuste no fluxo e na tabela de passo a passo em decorrência
da possibilidade de verificação de registro de todos os tipos de chaves
Pix.
```
```
Seção 16.2: ajuste no fluxo e na tabela de passo a passo em decorrência
da possibilidade de verificação de registro de todos os tipos de chaves
Pix.
```
```
Seção 17: ajustes no texto em decorrência da possibilidade de
verificação de registro de todos os tipos de chaves Pix.
```
```
11/2/2022 5.3 Seção 13: alteração no modo de recomposição de fichas dos baldes de
consulta do DICT, que passam a ser repostas após o recebimento da
ordem de pagamento pelo SPI na PACS.008, e não mais após uma
liquidação.
```
```
Seção 15: inclusão da informação em nota de rodapé da quantidade
máxima de 200 (duzentas) chaves passíveis de serem verificadas por
cada requisição da operação checkKeys.
```
```
1/9/2022 5.4 Seção 8.3: ajuste na etapa 5, para prever que o EndToEndId de uma
transação deve ser gerado pelo prestador de serviço de iniciação.
```
```
Seção 8.4: ajuste na etapa 5, para prever que o EndToEndId de uma
transação deve ser gerado pelo prestador de serviço de iniciação.
```

```
Seção 13: participantes que prestam serviço de iniciação devem passar
a usar o mesmo endpoint para consulta de chaves que os participantes
provedores de conta transacional. Como consequência, as regras de
limites e de decréscimo e de acréscimo de fichas passam a ser as
mesmas para todos os participantes.
```
```
Seção 15: aumento do incremento do balde e do tamanho máximo do
balde para a transação statistics_read.
```
3/10/2022 6.0 Estrutura: inserção da seção 20 “Consulta de baldes”.

```
Seção 10.3: ajuste na tabela de passo a passo (passo 12), para deixar
claro que não há especificação de valor a ser devolvido em uma
notificação de infração.
```
```
Seção 10.4: ajuste na tabela de passo a passo (passo 16), para deixar
claro que não há especificação de valor a ser devolvido em uma
notificação de infração.
```
```
Seção 13: (i) aumento de 10 para 20 no decréscimo de fichas por
consulta inválida de qualquer chave, para usuários pessoa natural e
pessoa jurídica; (ii) aumento de 8.000 para 12.000 e de 5.000 para
8.000 no incremento de fichas por minuto dos baldes das categorias A e
B, respectivamente; e (iii) ajuste no tamanho máximo do balde para
usuários finais pessoa natural e pessoa jurídica.
```
```
Seção 15: (i) alteração no nome da política de rate limit de keys.read
para keys.check; e (ii) inclusão de novas políticas de rate limit.
```
```
2/1/2023 6.1 Seção 10: ajustes no texto para enfatizar que o PSP do pagador deve
abrir a notificação de infração no DICT imediatamente após a
reclamação do usuário pagador.
```
```
Seção 10.3: ajuste no fluxo e na tabela de passo a passo para remover a
condição em que o PSP não pode acionar o MED.
```
```
Seção 10.4: ajuste no fluxo e na tabela de passo a passo para remover a
condição em que o PSP não pode acionar o MED.
```
```
Seção 18: (i) ajuste no texto para esclarecer que a solicitação do
cancelamento de devolução deve ser criada pelo PSP do recebedor; (ii)
inclusão do detalhamento sobre o monitoramento a ser realizado pelo
PSP em caso de devoluções parciais; e (iii) remoção da condição em que
o PSP não pode acionar o MED.
```
```
Seção 18.1: ajuste no fluxo e na tabela de passo a passo para remover a
condição em que o PSP não pode acionar o MED.
```

```
Seção 18.2: ajuste no fluxo e na tabela de passo a passo para remover a
condição em que o PSP não pode acionar o MED.
```
```
Seção 18.3: ajuste no fluxo e na tabela de passo a passo para remover a
condição em que o PSP não pode acionar o MED.
```
```
Seção 18.4: ajuste no fluxo e na tabela de passo a passo para remover a
condição em que o PSP não pode acionar o MED.
```
5/11/2023 7.0 Estrutura: exclusão da seção 14 “Informações vinculadas às chaves para
fins de segurança” e renumeração das seções posteriores

```
Seção 8: inclusão das informações retornadas pelo DICT quando uma
chave é consultada.
```
```
Seção 10: reestruturação da seção, com criação de duas subseções:
uma para detalhar a notificação de infração para solicitação de
devolução ou para cancelamento de devolução; e outra para detalhar a
notificação de infração para marcação de fraude transacional. O
detalhamento da funcionalidade foi atualizado para incluir novas
informações de segurança a serem compartilhadas com os
participantes. As subseções 10.1 e 10.2 da versão anterior foram
transformadas em subseções 10.2.1 e 10.2.2, respectivamente, com
ajustes no fluxo. As subseções 10.3 e 10.4 da versão anterior foram
transformadas em subseções 10.1.1 e 10.1.2, respectivamente. Foram
criadas, ainda, duas subseções, 10.1.3 e 10.1.4, para detalhar,
respectivamente, o fluxo de notificação de infração do tipo
“cancelamento de devolução” entre participantes com acesso direto ao
DICT e o fluxo de notificação de infração do tipo “cancelamento de
devolução” entre participantes com acesso indireto ao DICT
```
```
Seção 13: criação de duas subseções: 13.1 Mecanismos adotados pelo
DICT (que manteve o texto da versão anterior, com a atualização da
política de crédito de ficha em transações envolvendo prestadores de
serviço de iniciação e o detalhamento da política de limitação para a
nova operação getEntryStatistics ) e 13.2 Mecanismos que devem ser
adotados pelos participantes do Pix.
```
```
Seção 14 (corresponde à seção 15 da versão anterior): ajuste na política
de limite de requisições da operação getPersonStatistics e criação da
política de limite de requisição para as novas operações
getEntryStatistics e createFraudMarker.
```
```
Seção 17 (corresponde à seção 18 da versão anterior): ajuste no texto
para deixar claro que a conta deve ser monitorada em caso de
devolução parcial ou de rejeição da solicitação de devolução, desde que
a conta transacional não tenha sido encerrada, pelo usuário ou pelo
próprio PSP.
```

```
Seção 18 (corresponde à seção 19 da versão anterior): reestruturação
completa da seção, inclusive de seu título, para refletir as novas
informações de segurança que serão retornadas pelo DICT quando um
CPF, um CNPJ ou uma chave é consultada no endpoint statistics.
```
```
Seção 18.1 (corresponde à seção 19.1 da versão anterior): ajuste no
título e no fluxo.
```
```
Seção 18.2 (corresponde à seção 19.2 da versão anterior): ajuste no
título e no fluxo.
```
1/12/2023 7.1 Seção 13.1: Inclusão de duas novas categorias de baldes para
participantes no mecanismo de prevenção a ataque de leitura do DICT e
ajustes nos parâmetros de tamanho máximo e incremento temporal
dos baldes.
Seção 17.5: Inserção de determinação para que o PSP do pagador, caso
aceite a notificação de infração para cancelamento de devolução,
cancele imediatamente a notificação de infração para solicitação de
devolução que ele criou para solicitar a devolução da transação original.

```
2/5/2024 7.2 Seção 5: Ajuste no texto para informar que uma portabilidade pode ser
cancelada pelo PSP reivindicador enquanto o status do pedido for
“Aberto”.
```
```
Seção 6: Ajuste no texto para informar que uma reivindicação de posse
pode ser cancelada pelo PSP reivindicador enquanto o status do pedido
for “Aberto”.
```
```
Seção 10.1: Inserção de explicação sobre notificação de infração contra
usuário recebedor que atua como intermediário de pagamentos.
```
```
Seção 10.1: Inserção de determinação para que o PSP do pagador
cancele a solicitação de devolução aberta caso ele tenha cancelado a
notificação de infração que deu origem a ela. Se tiver havido
devolução, o PSP do pagador deverá devolver os recursos para o PSP do
recebedor através de uma nova transação Pix e abrir uma notificação
de infração para marcação de fraude contra seu usuário se concluir que
ele agiu de má fé.
```
```
Seção 13.1: Aumento da taxa de reposição por consulta de qualquer
chave após o recebimento da ordem de pagamento pelo SPI para 2
fichas para o balde de usuário PJ e aumento do incremento temporal
para 20 fichas a cada minuto em cada balde de usuário PJ.
```
```
Seção 13.1: Inserção da informação de que, excepcionalmente, a
critério do Banco Central do Brasil, os parâmetros de balde de um
usuário PJ podem ser alterados.
```

```
Seção 13.1: Inclusão de trecho na nota de rodapé para deixar claro que
solicitações de aumento de categoria de balde devem estar
devidamente fundamentadas em dados históricos, e não em projeções
futuras.
```
```
Seção 17: Inclusão de trecho para permitir que o PSP do recebedor
encerre o monitoramento da conta do usuário recebedor caso a
notificação de infração para solicitação de devolução seja cancelada
pelo PSP do pagador.
```
02/09/2024 7.3 Seção 1: criação da subseção 1.1 para orientações sobre chaves
bloqueadas por ordem judicial

```
Seção 8: reestruturação da seção, com criação de uma subseção 8.1
para detalhar quais informações devem ser exibidas ao usuário na
consulta de chave. As subseções 8.1, 8.2, 8.3 e 8.4 da versão antiga
foram transformadas 8.2, 8.3, 8.4 e 8.5 respectivamente
Seção 8.2 (antiga 8.1): inserção de texto, na tabela do fluxo, para deixar
claro que os dados da chave podem ser informados de forma manual
ou via leitura de QR Code.
```
```
Seção 8.3 (antiga 8.2): inserção de texto, na tabela do fluxo, para deixar
claro que os dados da chave podem ser informados de forma manual
ou via leitura de QR Code. Inserção de mais uma etapa para o PSP com
acesso direto ao DICT, para que verifique se a chave está cadastrada em
sua base interna.
```
```
Seção 8.4 (antiga 8.3): inserção de texto, na tabela do fluxo, para deixar
claro que os dados da chave podem ser informados de forma manual
ou via leitura de QR Code.
```
```
Seção 8.5 (antiga 8.4): inserção de texto, na tabela do fluxo, para deixar
claro que os dados da chave podem ser informados de forma manual
ou via leitura de QR Code. Inserção de mais uma etapa para o PSP com
acesso direto ao DICT, para que verifique se a chave está cadastrada em
sua base interna.
```
```
Seção 10: Inclusão de golpes de engenharia social e exclusão de texto
que restringia as possibilidades de enquadramento como fraude.
```
```
Seção 10.1: Exclusão de texto que restringia as possibilidades de
enquadramento como fraude no domínio “ scam ”.
```
```
Seção 13: criação da subseção 13.2.5 com as restrições dos dados da
chave exibidos ao usuário que faz a consulta.
```

```
Seção 13.1: Inclusão de texto explicativo do funcionamento do balde
quando estiver com poucas fichas (menos fichas que a penalização de
uma consulta com retorno de chave inválida).
```
```
Seção 13.2.3: Correção da relação entre chaves existentes e não
existentes para fins de monitoramento (NOT FOUND/(NOT
FOUND+FOUND)). Na nota de rodapé, a correção da mesma relação e
alteração do parâmetro de 30% para 20% no monitoramento de
usuários. Inclusão de texto para deixar claro que o endpoint checkKeys
do DICT é de uso exclusivo do PSP, não devendo ser disponibilizado,
mesmo que indiretamente, aos usuários.
```
```
Seção 16: Inclusão de texto para reforçar que o cache de existência de
chave Pix não deve servir de base para um serviço disponibilizado ao
usuário. Inclusão de texto para deixar claro que o endpoint checkKeys
do DICT é de uso exclusivo do PSP.
```
```
Seção 18: correções dos textos das respostas do DICT para alinhamento
com a terminologia da API do DICT.
```
16/06/2025 Seção 10: Inclusão, na nota de rodapé, de texto para incluir a
autorização do Pix automático nas possibilidades de fraude.

```
Seção 17: Inclusão do erro do PSP do pagador no envio de uma ordem
de pagamento referente ao Pix Automático como um dos casos de
possibilidade de abertura de solicitação de devolução criada pelo PSP
do pagador. Inclusão no quadro no campo de Motivo a descrição: erro
do PSP do pagador no envio de uma ordem de pagamento referente ao
Pix Automático (pix_automatico). Inclusão de nota de rodapé para
explicação de que erro operacional em transação de Pix Automático
não está incluído no motivo falha operacional do PSP do pagador.
Inclusão de nota de rodapé para explicação do que pode ser
considerado como erro do PSP do pagador no envio de uma ordem de
pagamento referente ao Pix Automático. Inclusão de texto, no quadro
dos campos para fechamento de uma solicitação de devolução, de que
se o motivo da solicitação de devolução for pix_automatico o
Identificador da transação de devolução ( RefundTransactionId ) deve ser
informado em uma pacs.008. Inclusão do texto “Nos casos relacionados
a transações de Pix Automático em que houver erro do PSP do pagador
no envio da ordem de pagamento, não há necessidade de
monitoramento pelo PSP do recebedor em caso de devolução parcial
ou de rejeição da solicitação”. Inclusão de nota de rodapé explicando
que as solicitações de devolução relacionadas aos casos de falha
operacional do PSP do pagador e aos casos envolvendo transações de
Pix Automático não requerem a criação prévia de notificações de
infração. Criação da subseção 17.6 para detalhamento do fluxo de
solicitação de devolução por erro do PSP do pagador no envio de ordem
de pagamento referente ao Pix Automático. Criação das subseções
```

```
17.6.1 com o fluxo de solicitação de devolução por erro do PSP do
pagador no envio de ordem de pagamento referente ao Pix Automático
para participantes do Pix com acesso direto ao DICT e 17.6.2 para os
participantes do Pix com acesso indireto ao DICT.
```
```
09/12/2024 7.4 Seção 1.1: inserção de texto para deixar claro que, caso uma chave
bloqueada judicialmente seja consultada em uma transação interna, o
PSP deve retornar a informação de bloqueio ao usuário, sem a exibição
das informações permitidas da chave.
```
```
Seção 2: validação de posse vira subseção 2.1, e adição de subseções
detalhando como a situação cadastral do usuário da Receita impacta a
criação e exclusão de chaves Pix.
```
```
Seção 4: criação da subseção “4.1. Exclusão de chave por
incompatibilidade de dados com a Receita Federal”, com a orientação
do código a ser usado na exclusão de chaves nessas situações. Os fluxos
anteriores 4.1, 4.2, 4.3 e 4.4 foram renumerados para 4.2, 4.3, 4.4 e 4.5,
respectivamente.
```
```
Seção 10.1: Inclusão de informações de contato (e-mail e telefone) do
PSP que abre a notificação de infração.
```
```
Seção 12: Aumento do prazo máximo do cache de chaves consultadas
para 180 segundos.
```
```
Seção 17: criação da subseção “17.1 – Solicitação de devolução por
falha operacional”, com orientações para a abertura e análise deste
tipo de solicitação de devolução. Os fluxos anteriores 17.1 e 17.2 foram
renumerados para 17.1.1 e 17.1.2. As subseções 17.3, 17.4, 17.5 e 17.6
da versão anterior foram transformadas em 17.2, 17.3, 17.4 e 17.5,
respectivamente.
```
```
01/04/2025 Seções 3.1, 3.2, 5.1, 5.2, 6.1, 6.2, 7.1 e 7.2: inclusão de etapa de validação
dos dados e situação cadastral do usuário na Receita Federal.
```
```
Seção 17: obrigatoriedade de preenchimento do campo RefundDetails
para pedido de devolução por falha operacional, e do campo
RefundAnalysisDetails nos casos de rejeição de pedido de devolução por
falha operacional.
```
19 / 03 /2025 7.5 Seção 2.2: detalhamentos feitos no texto da seção para incluir os
processos de alteração, portabilidade e reivindicação das chaves Pix e
para detalhar as situações cadastrais consideradas irregulares.

```
Seção 2.3: previamente "2.3. Prazo para regularização do cadastro na
Receita Federal", foi transformada em "2. 3. Validação dos nomes
vinculados às chaves Pix".
```

```
Seção 4.1: esclarecimentos sobre a prestação de informações ao usuário.
```
```
Seções 7.1 e 7.2: alteração do texto do passo 1 do diagrama para
englobar qualquer mudança nas informações vinculadas a chave por
iniciativa do PSP.
```
```
Inclusão de seção 7.3 com esclarecimentos sobre a prestação de
informações ao usuário.
```
```
Seção 12: alteração de obrigação para recomendação em relação à
utilização do cache interno para consultas de uma mesma chave pelo
mesmo participante dentro do prazo de validade.
```
```
Seção 13.2.2: alteração do termo “equivalente” por “igual” no que se
refere à política de limitação de consultas dos participantes em relação
à política de token bucket do DICT.
```
01/07/2025 Seções 3.1, 3.2, 7.1 e 7.2: alteração da data de entrada em vigor da etapa
de validação dos dados e situação cadastral do usuário na Receita
Federal durante o registro e alteração de chaves.

01/10/2025 Seções 5.1, 5.2, 6.1 e 6.2: alteração da data de entrada em vigor da etapa
de validação dos dados e situação cadastral do usuário na Receita
Federal durante a portabilidade e a reivindicação de posse de chave.

28 /0 8 /2025 8.0 Seção 6: melhoria de redação.

```
Seção 8: esclarecimento de que o nome social deve constar do CPF para
poder ser cadastrado em uma chave Pix de pessoa física.
```
```
Seção 9: esclarecimento sobre o motivo “reconciliação” para atualização
de informações.
```
```
Seção 13.1: esclarecimento de que o payer-ID informado na consulta ao
DICT deve ser o mesmo identificador do usuário pagador da transação
Pix relacionada e inclusão da necessidade de aprovação do diretor de
segurança cibernética do participante responsável (quando houver) para
pedidos de aumento de categoria de balde de consultas do DICT.
```
```
Seção 13.2.3: inclusão da necessidade de monitoramento em períodos
mais longos e da finalidade não permitida como uma situação anômala.
```
```
Seção 17: correção do parágrafo sobre o resultado da análise de uma
solicitação de devolução para considerar a liquidação, e não a emissão,
de uma pacs.004 ou de uma pacs.008, nos casos de Pix Automático ou
cancelamento de devolução.
```

Seção 17.1: esclarecimento sobre o envio de informações pelo PSP que
abre a solicitação de devolução por falha operacional.
01/10/2025 Seção 10.1: esclarecimento de que é permitido ao PSP do pagador editar
a notificação de infração enquanto ela está nos estados “aberta” ou
“recebida”.
23 / 11 / 2025 Aprimoramentos das regras e funcionalidades relacionadas ao MED:

- Alterações nas seções 10 “Notificação de Infração”; 10.1
    “Notificação de infração para solicitação de devolução ou para
    cancelamento de devolução”; e 17 “Fluxo de solicitação de
    devolução”.
- Criação da Seção 20: “Fluxo de Recuperação de Valores”
- Criação da Seção 21: “Notificações de Eventos”


