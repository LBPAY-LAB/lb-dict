


Faz-se necessário que a instituição tenha previamente concluído a preparação de que trata o art. 12 da Instrução Normativa BCB nº 508, de 30 de agosto de 2024, a saber:
a instituição não pode possuir pendências de portabilidade, de reivindicação, de posse ou de notificação de infração em ambiente de homologação;
registrar mil chaves Pix de um determinado tipo (exceto chave aleatória);
realizar, no mínimo, cinco transações, em ambiente de homologação, usando o participante virtual recebedor 99999004; e
estar apto a liquidar transações enviadas pelo virtual pagador 99999003.
Uma vez cumpridos esses requisitos, a instituição deve responder este e-mail informando:
o tipo de chave registrada;
o conteúdo do campo EndToEndId das cinco transações realizadas para o participante virtual recebedor 99999004; e
sugestão de data e horário para realização dos testes, que realizar-se-ão em dia útil e em horário comercial (todos os testes devem ser realizados dentro do espaço temporal de 1h). Salientamos que cabe ao Decem a definição da data e do horário definitivos para a execução dos testes.
Ressalte-se que o não envio dessas informações, a não conclusão ou a incorreção nos requisitos da preparação são fatores que impossibilitam o agendamento e, consequentemente, a execução dos testes.

Por ocasião da realização dos testes, o Banco Central do Brasil modificará e inserirá algumas chaves Pix manualmente, do tipo informado pela instituição, para serem utilizadas no teste de verificação de sincronismo, e criará reivindicações diversas ao longo da execução dos testes (Atenção: as claims serão geradas desde o primeiro minuto até o final do espaço temporal de 1h).

Por sua vez, a instituição deverá realizar os seguintes testes no espaço temporal agendado, nos termos do disposto na Instrução Normativa BCB nº 508, de 2024:
registro de chaves: registrar uma chave Pix de cada tipo;
consulta a chaves: consultar uma chave Pix de cada tipo (CPF, CNPJ, e-mail, número de telefone celular e chave aleatória);
verificação de sincronismo: realizar com sucesso uma verificação de sincronismo para o tipo de chave Pix registrada na etapa de preparação;
recebimento de reivindicações: realizar o recebimento de todas as portabilidades e reivindicações de posse geradas pelo Banco Central do Brasil, em que seja doador, em até um minuto;
fluxo de reivindicação: criar, confirmar, completar e cancelar pelo menos uma portabilidade ou reivindicação de posse, atuando como reivindicador;
fluxo de notificação de infração: criar, confirmar, completar e cancelar pelo menos uma notificação de infração; e
fluxo de solicitação de devolução:
                                        i.               criar pelo menos uma solicitação de devolução por motivo de falha operacional do prestador de serviços de pagamento do pagador;
                                      ii.               criar pelo menos uma solicitação de devolução por motivo de fundada suspeita de fraude;
                                   iii.               completar a solicitação de devolução por motivo de falha operacional, criada pelo PSP virtual pagador 99999003; e
                                    iv.               completar a solicitação de devolução por motivo de fundada suspeita de fraude, criada pelo PSP virtual pagador 99999003.



                                    Instrução Normativa BCB n° 508 de 30/8/2024


Versão vigente, atualizada em 6/1/2025
INSTRUÇÃO NORMATIVA BCB Nº 508, DE 30 DE AGOSTO DE 2024

Estabelece os procedimentos necessários para os testes formais de homologação no Diretório de Identificadores de Contas Transacionais (DICT), para a validação de QR Codes, para a validação da prestação de serviço de iniciação de transação de pagamento, para os testes de homologação para publicação de informações relativas ao serviço de saque e para os testes de homologação dos serviços do Pix Automático, no âmbito do Pix.

O Chefe do Departamento de Competição e de Estrutura do Mercado Financeiro (Decem), no uso das atribuições que lhe conferem o art. 23, inciso I, alínea “a”, e o art. 94, inciso IX, do Regimento Interno do Banco Central do Brasil, anexo à Resolução BCB nº 340, de 21 de setembro de 2023, e tendo em conta o disposto no art. 114 do Regulamento anexo à Resolução BCB nº 1, de 12 de agosto de 2020,

R E S O L V E :

Art. 1º  Esta Instrução Normativa estabelece, no âmbito do Pix, os procedimentos necessários para:

I - os testes formais de homologação no Diretório de Identificadores de Contas Transacionais (DICT);

II - a validação de QR Codes;

III - a validação da prestação de serviço de iniciação de transação de pagamento;

IV - os testes de homologação para publicação de informações relativas ao serviço de saque; e

V - os testes de homologação dos serviços do Pix Automático.

Art. 2º  Aplica-se o disposto nesta Instrução Normativa às instituições que, no âmbito do Pix, estão sujeitas:

I - à realização dos testes formais de homologação no DICT;

II - à realização dos testes formais de validação de QR Codes;

III - à realização dos testes formais de validação da prestação de serviço de iniciação de transação de pagamento;

IV - à publicação de informações relativas ao serviço de saque; ou

V - à realização dos testes formais de validação do Pix Automático.

CAPÍTULO I
DOS TESTES FORMAIS DE HOMOLOGAÇÃO NO DICT

Seção I
Dos Testes Formais de Homologação no DICT para as Instituições em Processo de Adesão ao Pix na Modalidade Liquidante Especial e para as Instituições Participantes ou em Processo de Adesão ao Pix nas Modalidades Provedor de Conta Transacional ou Instituição Usuária, ambas com Acesso Direto ao DICT

Subseção I
Dos Aspectos Gerais

Art. 3º  Os testes formais de homologação no DICT para as instituições em processo de adesão ao Pix na modalidade liquidante especial, bem como para aquelas em adesão nas modalidades provedor de conta transacional ou instituição usuária, ambas com acesso direto ao DICT, compreendem os testes de funcionalidades e o teste de capacidade.

Art. 3º  Os testes formais de homologação no DICT para as instituições em processo de adesão ao Pix na modalidade liquidante especial e na modalidade provedor de conta transacional com acesso direto ao DICT compreendem os testes de funcionalidades e o teste de capacidade. (Redação dada, a partir de 1º/1/2025, pela Instrução Normativa BCB nº 580, de 30/12/2024.)

Art. 3º-A  Os testes formais de homologação no DICT para as instituições em processo de adesão ao Pix na modalidade instituição usuária compreendem os testes de funcionalidades. (Incluído, a partir de 1º/1/2025, pela Instrução Normativa BCB nº 580, de 30/12/2024.)

Art. 4º  Os testes de funcionalidades são constituídos das seguintes etapas:

I - preparação da instituição;

II - preparação do Banco Central do Brasil;

III - execução dos testes; e

IV - conferência e resultado.

Art. 5º  O teste de capacidade é constituído das seguintes etapas:

I - execução dos testes; e

II - conferência e resultado.

Art. 6º  As instituições deverão obter prévia aprovação nos testes de funcionalidades a fim de submeterem-se ao teste de capacidade.

Art. 7º  As instituições em processo de adesão ao Pix na modalidade liquidante especial, bem como aquelas em adesão nas modalidades provedor de conta transacional ou instituição usuária, ambas com acesso direto ao DICT, deverão submeter-se aos testes de que trata o art. 3º, utilizando-se de ISPB próprio.

Art. 8º  Após a realização dos testes de que trata o art. 3º, as instituições em processo de adesão ao Pix na modalidade liquidante especial, bem como aquelas em adesão ou já participantes na modalidade provedor de conta transacional com acesso direto ao DICT e que pretendam prestar serviço de acesso a participante indireto, deverão submeter-se, ainda, aos testes de trata o art. 3º, utilizando-se de ISPB de participante indireto.

§ 1º  O participante indireto de que trata o caput deve estar devidamente cadastrado em ambiente de homologação do DICT a fim de que os testes pertinentes possam ser executados pela instituição que deseja prestar o serviço.

§ 2º  O Banco Central do Brasil não disponibiliza ISPB virtual para a realização dos testes de que trata o caput.

Art. 9º  A instituição que acessará o DICT de forma direta e que executará os testes deve solicitar agendamento prévio para tal.

Subseção II
Dos Aspectos Preliminares dos Testes de Funcionalidades

Art. 10.  O pedido de agendamento para a execução dos testes de funcionalidades de que trata o art. 9º deve ser apresentado ao Decem, pelo Protocolo Digital do Banco Central do Brasil (Protocolo Digital) em livre redação, indicando-se:

I - o ISPB e a razão social da instituição que acessará o DICT de forma direta e que executará os testes em ISPB próprio; ou

II - o ISPB e a razão social da instituição que acessa ou que acessará o DICT de forma direta, bem como o ISPB e a razão social de participante indireto, nos casos previstos no art. 8º.

Art. 11.  O Decem analisará o pedido de agendamento e, em resposta, fornecerá por e-mail, enviado aos contatos cadastrados para assuntos relacionados ao Pix, instruções direcionadas ao início da fase de preparação da instituição.

Subseção III
Da Preparação da Instituição para os Testes de Funcionalidades

Art. 12.  A preparação da instituição para os testes de funcionalidades constitui-se em:

I - registrar mil chaves Pix de um determinado tipo (exceto chave aleatória);

II - realizar, no mínimo, cinco transações, em ambiente de homologação, utilizando o participante virtual recebedor indicado pelo Decem em resposta à solicitação de agendamento dos testes, de que trata o art. 11;

III - observar as demais orientações apresentadas pelo Decem em resposta à solicitação de agendamento dos testes, de que trata o art. 11; e

IV - sugerir a data e o horário para a realização dos testes, que realizar-se-ão em dia útil e em horário comercial.

§ 1º  As informações sobre o tipo de chave de que trata o inciso I, o conteúdo do campo EndToEndId das transações de que trata o inciso II, bem como a sugestão de data e horário, de que trata o inciso IV, devem ser enviadas ao Decem por meio de envio de e-mail à caixa corporativa pix-operacional@bcb.gov.br.

§ 2º  Não havendo pendências por parte da instituição pleiteante, o Decem definirá a efetiva data de execução dos testes e a comunicará a respeito em e-mail enviado ao(s) contato(s) cadastrado(s) para assuntos relacionados ao Pix.

§ 3º  O disposto no inciso I do caput não se aplica às instituições em processo de adesão ao Pix na modalidade instituição usuária. (Incluído, a partir de 1º/1/2025, pela Instrução Normativa BCB nº 580, de 30/12/2024.)

Subseção IV
Da Preparação do Banco Central do Brasil para os Testes de Funcionalidades

Art. 13.  Preliminarmente ao início dos testes de funcionalidades, o Banco Central do Brasil modificará e inserirá algumas chaves Pix do tipo informado pela instituição, nos termos dispostos no § 1º do art. 12, para serem utilizadas no teste de verificação de sincronismo.

Parágrafo único.  O Banco Central do Brasil criará reivindicações diversas ao longo da execução dos testes de funcionalidades.

§ 1º  O Banco Central do Brasil criará reivindicações diversas ao longo da execução dos testes de funcionalidades. (Transformado em § 1º, a partir de 1º/1/2025, pela Instrução Normativa BCB nº 580, de 30/12/2024.)

§ 2º O disposto no caput não se aplica às instituições em processo de adesão ao Pix na modalidade instituição usuária. (Incluído, a partir de 1º/1/2025, pela Instrução Normativa BCB nº 580, de 30/12/2024.)

Subseção V
Da Execução dos Testes de Funcionalidades

Art. 14.  Todos os testes devem ser realizados dentro do prazo de uma hora, conforme horário determinado anteriormente pelo Banco Central do Brasil, nos termos dispostos no § 2º do art. 12.

Art. 15.  A instituição deve zelar para que, preliminarmente ao início dos testes, não haja pendências de portabilidade, de reivindicação de posse ou de notificação de infração em ambiente de homologação.

Art. 16.  Os seguintes testes devem ser realizados:

I - registro de chaves: registrar uma chave Pix de cada tipo (CPF, CNPJ, e-mail, número de telefone celular e chave aleatória);

II - consulta a chaves: consultar uma chave Pix de cada tipo (CPF, CNPJ, e-mail, número de telefone celular e chave aleatória);

III - verificação de sincronismo: realizar com sucesso uma verificação de sincronismo para o tipo de chave registrada na etapa de preparação, conforme disposto no inciso I do art. 12;

IV - recebimento de reivindicações: realizar o recebimento de todas as portabilidades e reivindicações de posse geradas pelo Banco Central do Brasil, em que a instituição em teste seja doador, em até um minuto após cada recebimento;

V - fluxo de reivindicação: atuando como reivindicador, criar, confirmar, completar e cancelar pelo menos uma portabilidade ou uma reivindicação de posse;

VI - fluxo de notificação de infração: criar, confirmar, completar e cancelar pelo menos uma notificação de infração; e

VII - fluxo de solicitação de devolução: criar e completar pelo menos uma solicitação de devolução por falha operacional do prestador de serviços de pagamento do pagador e uma por fundada suspeita de fraude.

§ 1º  No teste de que trata o inciso II, poderá haver, a critério do Banco Central do Brasil, solicitação de consulta a uma ou mais chaves específicas.

§ 2º  Na hipótese de que trata o § 1º, a solicitação de consulta poderá constar da confirmação de agendamento dos testes, de que trata o § 2º do art. 12, ou ser apresentada durante a execução dos testes, a critério do Banco Central do Brasil.

§ 3º  O disposto no inciso III do caput não se aplica às instituições em processo de adesão ao Pix na modalidade instituição usuária. (Incluído, a partir de 1º/1/2025, pela Instrução Normativa BCB nº 580, de 30/12/2024.)

Subseção VI
Da Conferência e do Resultado dos Testes de Funcionalidades

Art. 17.  Após o encerramento do horário previsto para a realização dos testes de funcionalidades, o Decem analisará o desempenho da instituição e a informará, por e-mail enviado aos contatos cadastrados para assuntos relacionados ao Pix, acerca de sua aprovação ou de sua não aprovação, indicando, nesse caso, os critérios inobservados pela instituição executante.

Art. 18.  A instituição que não obtiver aprovação na execução dos testes poderá submeter-se a até duas novas tentativas.

Parágrafo único.  A instituição que pretenda submeter-se a nova tentativa de execução dos testes de funcionalidades, nos termos do caput, deverá solicitar novo agendamento, nos termos do art. 10, e reiniciar o processo desde a etapa de preparação da instituição.

Art. 19.  A instituição aprovada nos testes de funcionalidades poderá pleitear agendamento para a execução do teste de capacidade.

Subseção VII
Dos Aspectos Preliminares, do Agendamento e das Orientações para o Teste de Capacidade

Art. 20.  O pedido de agendamento para o teste de capacidade de que trata o art. 19 deve ser apresentado ao Decem, pelo Protocolo Digital em livre redação, indicando-se:

I - o ISPB e a razão social da instituição que acessará o DICT de forma direta e que executará os testes em ISPB próprio;

II - o ISPB e a razão social da instituição que acessa ou que acessará o DICT de forma direta, bem como o ISPB e a razão social de participante indireto, nos casos previstos no art. 8º; e

III - sugestão de data e de horário para a realização do teste, que realizar-se-á em dia útil e em horário comercial.

Art. 21.  O Decem analisará o pedido e definirá a efetiva data de execução do teste.

Parágrafo único.  O Decem comunicará a instituição pleiteante, por e-mail enviado aos contatos cadastrados para assuntos relacionados ao Pix, acerca da data de que trata o caput, bem como fornecerá instruções direcionadas à execução do teste, dentre elas o intervalo de chaves a ser utilizado, nos termos do disposto no art. 23.

Subseção VIII
Da Execução do Teste de Capacidade

Art. 22.  O teste de capacidade deve ser realizado dentro do prazo de uma hora, conforme horário determinado anteriormente pelo Banco Central do Brasil, nos termos dispostos no caput do art. 21.

Art. 23.  O teste de capacidade consiste em:

I - consultar, no mínimo, mil chaves diferentes em um intervalo de sessenta segundos e receber resposta do DICT com sucesso, caso a instituição mantenha até um milhão de contas transacionais;

II - consultar, no mínimo, duas mil chaves diferentes em um intervalo de sessenta segundos e receber resposta do DICT com sucesso, caso a instituição mantenha entre um milhão e dez milhões de contas transacionais; ou

III - consultar, no mínimo, quatro mil chaves diferentes em um intervalo de sessenta segundos e receber resposta do DICT com sucesso, caso a instituição mantenha mais de dez milhões de contas transacionais.

Parágrafo único.  As consultas de que tratam os inciso I, II e III devem durar dez minutos e serem distribuídas de forma homogênea ao longo do tempo, com o total de operações, no mínimo, igual a:

I - dez mil, caso a instituição mantenha até um milhão de contas transacionais;

II - vinte mil, caso a instituição mantenha entre um milhão e dez milhões de contas transacionais; ou

III - quarenta mil, caso a instituição mantenha mais de dez milhões de contas transacionais.

Subseção IX
Da Conferência e do Resultado do Teste de Capacidade

Art. 24.  Após o encerramento do horário previsto para a realização do teste, o Decem analisará o desempenho da instituição e a informará, por e-mail enviado ao(s) contato(s) cadastrado(s) para assuntos relacionados ao Pix, acerca de sua aprovação ou de sua não aprovação, indicando, nesse caso, os critérios inobservados pela instituição executante.

Art. 25.  A instituição que não obtiver aprovação na execução dos testes poderá submeter-se a até duas novas tentativas.

Parágrafo único.  A instituição que pretenda submeter-se a nova tentativa de execução do teste, nos termos do caput, deverá solicitar novo agendamento, nos termos do art. 20.

Art. 26.  A instituição aprovada nos testes de funcionalidades e no teste de capacidade será considerada aprovada nos testes formais de homologação no DICT.

Seção II
Dos Testes Formais de Homologação no DICT para as Instituições em Processo de Adesão ao Pix na Modalidade Iniciador com Acesso Direto ao DICT

Subseção I
Dos Aspectos Gerais

Art. 27.  Os testes formais de homologação no DICT para as instituições em processo de adesão ao Pix na modalidade iniciador, com acesso direto ao DICT, compreendem a execução do teste de consulta de chaves e do teste de capacidade.

§ 1º  Os testes de que trata o caput são constituídos das seguintes etapas:

I - execução do teste; e

II - conferência e resultado.

§ 2º  Os testes de que trata o caput aplicam-se ainda às instituições já participantes do Pix na modalidade iniciador que desejam passar a acessar o DICT de forma direta.

Art. 28.  As instituições deverão obter prévia aprovação no teste de consulta de chave a fim de submeterem-se ao teste de capacidade.

Art. 29.  A instituição que executará os testes deve solicitar agendamento prévio para tal.

Subseção II
Dos Aspectos Preliminares do Teste de Consulta de Chave

Art. 30.  O pedido de agendamento para o teste de consulta de chave de que trata o art. 29 deve ser apresentado ao Decem, pelo Protocolo Digital em livre redação, indicando-se:

I - o ISPB e a razão social da instituição que executará o teste; e

II - sugestão de data e de horário para a realização do teste, que realizar-se-á em dia útil e em horário comercial.

Art. 31.  O Decem analisará o pedido de agendamento e, não havendo pendências por parte da instituição pleiteante, enviará comunicado, por e-mail enviado ao(s) contato(s) cadastrado(s) para assuntos relacionados ao Pix, acerca da efetiva data de execução do teste, bem como fornecerá instruções pertinentes.

Parágrafo único.  Dentre as instruções de que trata o caput, não serão fornecidas chaves para a execução do teste de consulta de chaves, ficando a cargo da instituição executante a obtenção das chaves a serem consultadas.

Subseção III
Da Execução do Teste de Consulta de Chave

Art. 32.  O teste de consulta de chave consiste na consulta de pelo menos uma chave Pix de cada tipo (CPF, CNPJ, e-mail, número de telefone celular e chave aleatória).

§ 1º  O teste deve ser realizado dentro do prazo de uma hora, conforme horário anteriormente determinado pelo Banco Central do Brasil, nos termos dispostos no caput do art. 31.

§ 2º  Poderá haver, a critério do Banco Central do Brasil, solicitação de consulta a uma ou mais chaves específicas.

§ 3º  Na hipótese de que trata o § 2º, a solicitação de consulta poderá constar da confirmação de agendamento dos testes, de que trata o caput do art. 31, ou ser apresentada durante a execução do teste, a critério do Banco Central do Brasil.

Subseção IV
Da Conferência e do Resultado do Teste de Consulta de Chave

Art. 33.  Após o encerramento do horário previsto para a realização do teste de consulta de chave, o Decem analisará o desempenho da instituição e a informará, por e-mail enviado ao(s) contato(s) cadastrado(s) para assuntos relacionados ao Pix, acerca de sua aprovação ou de sua não aprovação, indicando, nesse caso, os critérios inobservados pela instituição executante.

Art. 34.  A instituição que não obtiver aprovação na execução do teste poderá submeter-se a até duas novas tentativas.

Parágrafo único.  A instituição que pretenda submeter-se a nova tentativa de execução do teste, nos termos do caput, deverá solicitar novo agendamento, nos termos do art. 30.

Art. 35. A instituição aprovada no teste de consulta de chave poderá pleitear agendamento para a execução do teste de capacidade.

Subseção V
Dos Aspectos Preliminares, do Agendamento e das Orientações para o Teste de Capacidade

Art. 36.  O pedido de agendamento para o teste de capacidade de que trata o art. 35 deve ser apresentado ao Decem, pelo Protocolo Digital em livre redação, indicando-se o ISPB e a razão social da instituição que executará o teste.

Art. 37.  O Decem analisará o pedido e comunicará a instituição, por e-mail enviado ao(s) contato(s) cadastrado(s) para assuntos relacionados ao Pix, acerca da efetiva data de execução do teste, bem como fornecerá instruções pertinentes.

Subseção VI
Da Execução do Teste de Capacidade

Art. 38.  O teste de capacidade deve ser realizado dentro do prazo de uma hora, conforme prazo anteriormente determinado pelo Banco Central do Brasil, nos termos disposto no art. 37.

Art. 39.  O teste de capacidade consiste em consultar, no mínimo, mil chaves diferentes em um intervalo de sessenta segundos e receber resposta do DICT com sucesso.

Parágrafo único.  As consultas devem durar dez minutos e serem distribuídas de forma homogênea ao longo do tempo, com o total de operações, no mínimo, igual a dez mil.

Subseção VII
Da Conferência e do Resultado do Teste de Capacidade

Art. 40.  Após o encerramento do horário previsto para a realização do teste, o Decem analisará o desempenho da instituição, por e-mail enviado ao(s) contato(s) cadastrado(s) para assuntos relacionados ao Pix, e a informará acerca de sua aprovação ou de sua não aprovação, indicando, nesse caso, os critérios inobservados pela instituição executante.

Art. 41.  A instituição que não obtiver aprovação na execução do teste poderá submeter-se a até duas novas tentativas.

Parágrafo único.  A instituição que pretenda submeter-se a nova tentativa de execução dos testes, nos termos do caput, deverá solicitar novo agendamento, nos termos do art. 36.

Art. 42.  A instituição aprovada no teste de consulta de chave e no teste de capacidade será considerada aprovada nos testes formais de homologação no DICT.

Seção III
Das Disposições Gerais

Art. 43. Será considerada reprovada nos testes formais de homologação no DICT a instituição que não lograr sucesso na execução de teste agendado pela terceira vez consecutiva.

Parágrafo único.  O disposto no caput não se aplica aos testes realizados por instituição em processo de adesão ao Pix na modalidade provedor de conta transacional ou já participante nessa modalidade quando realizados com ISPB de participante indireto, nos termos previstos no art. 8º.

CAPÍTULO II
DOS TESTES FORMAIS DE VALIDAÇÃO DE QR CODES

Art. 44.  Os testes formais de validação de QR Codes compreendem aqueles relacionados:

I - à jornada do usuário pagador, que consiste na validação da leitura de QR Codes:

a) estáticos e dinâmicos, associados ao Pix Cobrança;

b) estáticos e dinâmicos, com finalidade de saque; e

c) dinâmicos, com finalidade de troco; e

II - à jornada do usuário recebedor, que consiste na validação da geração de QR Codes:

a) estáticos, associados ao Pix Cobrança, para pagamentos imediatos;

b) estáticos, com finalidade de saque;

c) dinâmicos, associados ao Pix Cobrança, para pagamentos imediatos;

d) dinâmicos, com finalidade de saque;

e) dinâmicos, com finalidade de troco; e

f) dinâmicos, associados ao Pix Cobrança, para pagamentos com vencimento.

Art. 45.  Os testes formais de validação de leitura de QR Codes, referentes à jornada do usuário pagador, visam o envio de Pix por meio de leitura de QR Code estático ou de QR Code dinâmico, e são obrigatórios:

I - para participantes do Pix na modalidade provedor de conta transacional, ou em processo de adesão nessa modalidade, que ofertem contas transacionais a usuários finais pessoas naturais;

II - para participantes do Pix na modalidade provedor de conta transacional, ou em processo de adesão nessa modalidade, que ofertem contas transacionais exclusivamente a usuários finais pessoas jurídicas e que, adicionalmente, disponibilizem, ou pretendam disponibilizar, para esse público, o envio de Pix por meio de leitura de QR Code estático ou de QR Code dinâmico; e

II - para participantes do Pix na modalidade instituição usuária, ou em processo de adesão nessa modalidade, e que, de forma facultativa, pretendam consumir a leitura de QR Codes.

Parágrafo único.  Os testes de que trata o caput compreendem a validação da leitura, no Pix Tester:

a) de QR Codes estáticos e dinâmicos, visando o envio de Pix associados ao Pix Cobrança, e o correto envio das respectivas ordens de pagamento;

b) de QR Codes estáticos e dinâmicos, associados ao Pix Cobrança com erros intencionais, bem como o correto tratamento das inconsistências encontradas;

c) de QR Codes estáticos e dinâmicos, visando o envio de Pix com finalidade de saque, bem como o correto envio das respectivas ordens de pagamento;

d) de QR Codes estáticos e dinâmicos, visando o envio de Pix com finalidade de saque com erros intencionais, bem como o correto tratamento das inconsistências encontradas;

e) de QR Codes dinâmicos, visando o envio de Pix com finalidade de troco, bem como o correto envio das respectivas ordens de pagamento; e

f) de QR Codes dinâmicos, visando o envio de Pix com finalidade de troco com erros intencionais, bem como o correto tratamento das inconsistências encontradas.

Art. 46.  Os testes formais de geração de QR Codes, referentes à jornada do usuário recebedor, visam o recebimento de Pix por meio de leitura de QR Code estático ou de QR Code dinâmico, gerados, no Pix Tester, por instituição participante do Pix nas modalidades provedor de conta transacional ou instituição usuária, ou em processo de adesão nessas modalidades, e compreendem a validação da geração:

I - de QR Codes estáticos, associados ao Pix Cobrança para pagamentos imediatos, obrigatórios para instituições:

a) participantes do Pix na modalidade provedor de conta transacional, ou em processo de adesão nessa modalidade, ofertantes de contas transacionais a usuários finais pessoas naturais;

b) participantes do Pix na modalidade provedor de conta transacional, ou em processo de adesão nessa modalidade, ofertantes de contas transacionais a usuários finais pessoas jurídicas e que, adicionalmente, optem por disponibilizar para esse público o recebimento de Pix por meio da geração de QR Code estático, associado ao Pix Cobrança para pagamentos imediatos;

c) participantes do Pix na modalidade provedor de conta transacional, ou em processo de adesão nessa modalidade, ofertante de contas transacionais a pessoas jurídicas e que, adicionalmente, optem por disponibilizar para esse público o recebimento de Pix por meio da geração de QR Code estático, com finalidade de saque; e

d) participantes do Pix na modalidade instituição usuária, ou em processo de adesão nessa modalidade, e que, de forma facultativa, pretendam consumir a geração de QR Codes estáticos, associados ao Pix Cobrança para pagamentos imediatos;

II - de QR Codes estáticos, com finalidade de saque, obrigatórios para instituições:

a) participantes do Pix na modalidade provedor de conta transacional, ou em processo de adesão nessa modalidade, ofertantes de contas transacionais a usuários finais pessoas jurídicas e que, adicionalmente, optem por disponibilizar para esse público o recebimento de Pix por meio da geração de QR Code estático, associado ao Pix Cobrança para pagamentos imediatos; e

b) participantes do Pix na modalidade provedor de conta transacional, ou em processo de adesão nessa modalidade, ofertantes de contas transacionais a usuários finais pessoas jurídicas e que, adicionalmente, optem por disponibilizar para esse público o recebimento de Pix por meio da geração de QR Code estático, com finalidade de saque;

III - de QR Codes dinâmicos, associados ao Pix Cobrança para pagamentos imediatos, obrigatórios para instituições:

a) participantes do Pix na modalidade provedor de conta transacional, ou em processo de adesão nesse modalidade, ofertantes de contas transacionais a usuários finais pessoas jurídicas e que, adicionalmente, optem por ofertar a API Pix; e

b) participantes do Pix na modalidade instituição usuária, ou em processo de adesão nessa modalidade, e que, de forma facultativa, pretendam consumir a geração de QR Codes dinâmicos associados ao Pix Cobrança para pagamentos imediatos;

IV - de QR Codes dinâmicos, com finalidade de saque, e de QR Codes dinâmicos, com finalidade de troco, obrigatórios para instituições participantes do Pix na modalidade provedor de conta transacional, ou em processo de adesão nesse modalidade, ofertantes de contas transacionais a usuários pessoas jurídicas, e que, adicionalmente, optem por ofertar a API Pix; e

V - de QR Codes dinâmicos, associados ao Pix Cobrança para pagamentos com vencimento, obrigatórios para instituições:

a) participantes do Pix na modalidade provedor de conta transacional, ou em processo de adesão nessa modalidade, ofertantes de contas transacionais a usuários finais pessoas jurídicas e que, adicionalmente, optem por disponibilizar a esse público o recebimento de Pix por meio da geração de QR Codes dinâmicos associados ao Pix Cobrança; e

b) participantes do Pix na modalidade instituição usuária, ou em processo de adesão nessa modalidade, e que, de forma facultativa, pretendam consumir a geração de QR Codes dinâmicos associados ao Pix Cobrança para pagamento com vencimento.

Art. 47.  Os participantes devem manter documentação que comprove a execução dos testes formais de validação de QR Codes para eventual análise por parte do Banco Central do Brasil pelo prazo de cinco anos.

CAPÍTULO III
DOS TESTES FORMAIS DE VALIDAÇÃO DA PRESTAÇÃO DE SERVIÇO DE INICIAÇÃO DE TRANSAÇÃO DE PAGAMENTO

Art. 48.  Os testes formais de validação da prestação de serviço de iniciação de transação de pagamento, no âmbito do Pix, compreendem aqueles relacionados:

I - à atuação, no âmbito do Pix, na figura do iniciador de transação de pagamento, nos termos da regulamentação específica do Open Finance; e

II - à atuação, no âmbito do Pix, na figura do detentor de contas, nos termos da regulamentação específica do Open Finance.

Art. 49.  Os testes destinados à figura do detentor de contas, no âmbito do Pix, consistem na verificação de sua aptidão para o envio de Pix após o recebimento de um pedido de iniciação de transação de pagamento enviado por instituição atuando na figura do iniciador de transação de pagamento.

Parágrafo único.  Os testes de que trata o caput compreendem a validação com sucesso, no Pix Tester, do envio de Pix mediante prévio recebimento de pedido de iniciação:

I - com chave Pix;

II - com inserção manual dos dados da conta transacional do usuário recebedor;

III - nos casos em que a instituição que está atuando na figura do iniciador de transação de pagamento detém todas as informações do usuário recebedor; e

IV - com leitura de QR Code.

Art. 50.  Os testes destinados à figura do iniciador de transação de pagamento, no âmbito do Pix, consistem na verificação de sua aptidão para a emissão de um pedido de iniciação de transação de pagamento com Pix para instituição atuando na figura do detentor de conta.

§ 1º  Os testes de que trata o caput compreendem a validação com sucesso, no Pix Tester, da emissão de um pedido de iniciação de transação de pagamento com Pix:

I - com chave Pix;

II - com inserção manual dos dados da conta transacional do usuário recebedor;

III - nos casos em que a instituição que está atuando na figura do iniciador de transação de pagamento detém todas as informações do usuário recebedor; e

IV - com leitura de QR Code.

§ 2º  Para a execução dos testes de que trata o caput, faz-se necessária a colaboração de instituição que na atue na figura do detentor de conta.

CAPÍTULO IV
DOS TESTES FORMAIS PARA PUBLICAÇÃO DE INFORMAÇÕES RELATIVAS AO SERVIÇO DE SAQUE

Art. 51.  Os testes formais para publicação de informações relativas ao serviço de saque, obrigatórios para os participantes do Pix na modalidade provedor de conta transacional, ou em processo de adesão nessa modalidade, que desejam facilitar serviço de saque, compreendem:

I - a atualização do seu Catálogo de Dados Abertos, em ambiente de homologação, de forma a incluir o conjunto de dados referente às informações relativas ao serviço de saque, conforme orientações disponíveis no endereço eletrônico  https://www.bcb.gov.br/estabilidadefinanceira/dadosabertossfn; e

II - a disponibilização do conjunto de dados referente às informações relativas ao serviço de saque, nos termos da Instrução Normativa BCB nº 313, de 24 de outubro de 2022, conforme especificações técnicas disponíveis no endereço eletrônico https://www.bcb.gov.br/estabilidadefinanceira/dadosabertossfn, no local informado em seu Catálogo de Dados Abertos.

Art. 52.  Para a aprovação nos testes de que trata este Capítulo, o Decem verificará se:

I - houve a devida inclusão do conjunto de dados no ambiente de homologação do Catálogo de Dados Abertos do participante, nos termos dispostos no inciso I do art. 51;

II - o conjunto de dados encontra-se disponível no local informado, nos termos dispostos no inciso II do art. 51; e

III - as informações disponibilizadas obedecem às especificações definidas na Instrução Normativa BCB nº 313, de 2022, e no endereço eletrônico https://www.bcb.gov.br/estabilidadefinanceira/dadosabertossfn, nos termos dispostos no inciso II do art. 51.

CAPÍTULO V
DOS TESTES FORMAIS DE VALIDAÇÃO DO PIX AUTOMÁTICO

Art. 53.  Os testes formais para oferta de serviços do Pix Automático compreendem a execução com sucesso dos cenários de simulação no Pix Tester.

§ 1º  Os testes no Pix Tester estão divididos em cenários para instituições que atuam na ponta recebedora, e na ponta pagadora, e para instituições não ofertantes de pagamentos via Pix Automático a usuários pessoas jurídicas.

§ 2º  A instituição deverá cumprir todos os cenários que se aplicam aos serviços do Pix Automático que ofertará.

§ 3º  Instituições que pediram dispensa da oferta de pagamentos via Pix Automático para usuários pessoas jurídicas deverão cumprir os testes destinados a rejeição de mensagens de Pix Automático encaminhadas indevidamente.

Art. 54.  Os testes formais de validação do Pix Automático são obrigatórios na ponta pagadora:

I - para participantes do Pix na modalidade provedor de conta transacional, ou em processo de adesão nessa modalidade, caso ofertantes de contas transacionais a usuários finais pessoas naturais;

II - para participantes do Pix na modalidade provedor de conta transacional, ou em processo de adesão nessa modalidade, caso ofertantes de contas transacionais a usuários finais pessoas jurídicas e optem por ofertar o pagamento com Pix Automático a esse público; e

III - para participantes do Pix na modalidade instituição usuária, ou em processo de adesão ao Pix nessa modalidade, caso optem por consumir o pagamento com Pix Automático.

Art. 55.  Os testes formais de validação do Pix Automático são obrigatórios na ponta recebedora:

I - para participantes do Pix na modalidade provedor de conta transacional, ou em processo de adesão nessa modalidade, caso exclusivamente ofertantes de contas transacionais a usuários pessoas jurídicas e optem por ofertar o recebimento de Pix Automático a esse público; e

II - para participantes do Pix na modalidade instituição usuária, ou em processo de adesão nessa modalidade, caso optem por consumir o recebimento com Pix Automático.

Art. 56.  Os testes formais de rejeição de mensagens para instituições não ofertantes são obrigatórios para os participantes provedores de conta transacional, caso ofertantes de contas transacionais a usuários pessoas jurídicas e optantes pela dispensa da oferta de pagamentos via Pix Automático a esse público.

CAPÍTULO VI
DAS DISPOSIÇÕES GERAIS

Art. 57.  O Decem pode, a seu exclusivo critério, determinar a realização de testes homologatórios complementares àqueles previstos nesta Instrução Normativa.

Art. 58  Fica revogada a Instrução Normativa BCB nº 290, publicada no Diário Oficial da União de 1º de agosto de 2022.

Art. 59.  Esta Instrução Normativa entra em vigor na data de sua publicação.

RICARDO TEIXEIRA LEITE MOURÃO

NOTA

O Decreto nº 10.411, de 30 de junho de 2020, prevê a obrigatoriedade da realização de análise de impacto regulatório (AIR) para a edição de atos normativos de interesse geral produzidos pelos órgãos e entidades da administração pública federal direta e indireta.

Todavia, consoante se definiu no parágrafo 8 do Voto 280/2021–BCB, de 10 de novembro de 2021, o Regulamento do Pix, inclusive os demais documentos que o integram ou que o detalham e o complementam, não se caracterizam como ato regulatório de força cogente, ostentando, em verdade, natureza eminentemente contratual. Assim, modificações promovidas no referido regulamento e nos demais documentos que o integram ou que o detalham e o complementam não se sujeitam à produção prévia de AIR.