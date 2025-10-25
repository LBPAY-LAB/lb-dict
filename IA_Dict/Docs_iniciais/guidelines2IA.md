Fatos
    Somos o LB PAY, uma instituição de pagamento licenciada pelo Banco Central do Brasil e participante direto do PIX.
    Implementamos a nossa própria solução de Core Banking (Contas de Pagamento) e de integrações com o Bacen via RSFN: SPI (PIX) e agora precisamos de implementar a nossa solução DICT (Chaves PIX)
    Este projeto tem como objetivo organizar e munir a squad claude code de todos os documentos necessários para a concepção, implementação do projeto e suporte para a homologação (funcionalidades necessários para despoletar o que for necessário funcionalmente para homologar)
    Resumindo, Para podermos operar precisamos de homologar no DICT  (Bacen)
    O objetivo deste projeto é a implementação do módulo dict visando a homologação primeiro e entrar em produção depois.

Nesse sentido, identificamos os seguintes macro passos do projeto IA_Dict:

1º Abstrair todos os requisitos funcionais com base no manual do Bacen (manual_Operacional_DICT_Bacen.md) e interface Rest API com o Dict (OpenAPI_Dict_Bacen)
   - Terá que ser gerada uma checklista total de todos os requisitos funcionais abstraidos do manual operacional do dict Bacen, gerando um documento de checklist de requisitos.
   - erá que ser gerada uma lista de funcionalidades de front-end;
   - Terá que ser gerada uma lista das funções que a interface do módulo core dict terá que disponibilizar para atender o front-end e de acordo os vários tipos de requisitos de sincronismo;
   - Terá que ser gerada toda a logica de negócio interna que irá representar o módulo core dict (logica de negócio conforme os requisitos do bacen)
   - Terá ser geradas a lista de todas as chamadas que terão que existir ao connect e bridge.
   - Terá que gerar um plano de todas as tarefas e actividades sequênciais e com dependências;
   - Terá que criar uma squad do projeto: Lista dos vários tipos de agentes necessários e indexálos ao plano de actividades. Todos agentes terão que ter um nome de coódigo e todas as tarefas terão que ter um codigo;
   - Todas as listas seja do que for: funcionalidades, requisito funcionais, processos de negócios, diagramas de processos... tudo deverá ser numerado para que se possam indexar entre si. Um requisito de negócio-> Processo de negócio que será implementado no core dict, -> Front-End -> Interfaces.... Desta forma os agentes de implementação conseguiram criar a visão end2end da implementação (desde o front-end) até a integração com o Bacen e a volta.
   -> Torna-se assim fundamental que um dos agentes seja um project-manager e outro scrum master, permitindo assim a gestão total de todas as cards de back log de desenvolvimento que foram criadas

2º Abstrair a arquitetura de implementação, integração com os módulos internos e com o Dict (Bacen) com base nas definições de arquitetura que foram desenhadas pela area de arquitetura do LBPAY, e que constam no documento ArquiteturaDict_LBPAY que foi resultado de uma exportação icePanel (Diagrama C4)
    - As implementações já efectudas deverão constituir para a IA um padrão de desenvolvimento no entanto precisam de evoluir de modo a permitir que os módulos Connect Dict e Bridge Dict sejam completamente abstratos.
3º Inferir e documentar todos os fluxos end2end: Cliente (Fronte-End)->Core Banking (modulo Dict) -> Connect (Interface interna) - > Bridge (interface para o Bacen)
4º Abstrair toda a logica de negócio que terá que ser implementada no moódulo do Core Dict e as interfaces grpc que terão que atender o Front-End e chamar a rest API do Bacen (via connect->bridge) conforme será observado no documento de arquitetura. Fundamental entender quais os fluxos que serão sincronos e os que serão assicronos. o BAckLog(plano Dict).csv já mapea uma grande parte desses requisitos de sincronismo.
5º Entender que repositórios já existentes, que já implementam parte da arquitetura implementada, mas precisa de ser evoluida e acrescentada para que sejam 100% compliances com o banco central que pode ser encontrado no documento Backlog(Plano DICT).csv. Aqui podemos observar o que já está implementado e em que repositórios.
6º Inferir toda a stack tecnologica, com base nos repositórios e no documento de arquitetura mencionado no 2º ponto.
7º 

Grandes objetivos e prioridades:
-> Os modulos de Connect Dict e Bridge Dict deverão ser implementados de forma 100% abstracta permitindo ser usados como "trilhos" para interfacear com o Bacen qualquer que seja o tipo de interfaces que tenha que ser chamada na rest api do Bacen.
-> Nesse sentido a arquitetura do Bridge e de Connect terão que evoluir e implicará a reimplementação desses repos.

Para podermos operar precisamos de homologar no DICT  (Bacen)
O objetivo deste projeto é a implementação do módulo dict visando a homologação antes de entrar em produção

Documentos que constam na pasta: Docs_inicias:
manual_Operacional_DICT_Bacen.md : Requisitos funcionais do Bacen (objeto da homologação)
OpenAPI_Dict_Bacen
ArquiteturaDict_LBPAY
Backlog(Plano DICT).csv
Requisitos_Homologação_Dict.md

Por fim, antes de se iniciar a implementação do projeto, o objetivo deste projeto é produzir de forma organizar todos os artefactos necessários para que os agentes fiquem completamente autonomos. Esses artefactos terão que ser inicialmente aprovados pelo CTO do projeto, o head de arquitetura e o head de produto e o head de engenharia, conforme já referenciado neste texto anteriormente.

Todos os critérios de aceitação deverão ficar muito bem definidos de modo a permitir que os agentes responsaveis pelos testes possam testar, validar e homologar a solução. Obviamente com critérios aseados nos requisitos de homologação que constam na documentação oficial do banco central (Requisitos_Homologação_Dict.md).

Como primeira tarefa o Claude Code deverá criar a .Claude folder, os agentes que irão montar este projeto de especificação e todo o planeamento da proópria especificação a que chamaremos Squad_de_Arquitetura.

O projeto só avança para a implementação quantos todos os artefactos considerados necessários tiverem sido criados e aprovados. Fundamental: esses artefactos terão que ser orientados a agentes Claude Code porque o projeto será 100% immplementado por agentes: desde designers, programadores, devops, dbas, tudo o que for necessário.

Por ultimo, via MCP, o projeto irá implementar branches em repositórios ja existem, dando origem a PR (pull requests). Claro um dos primeiros requisitos será a clonagem de todos os repositórios já existentes que se enquadrem neste projeto LBPAY Dict e os novos que terão que ser criados.

A metodologia de gestão de todas essas implementações serão definidas pelo project-manager deste projeto que definirá por sua vez as metodologias de gestão da SQUAD de desenvolvimento que este projecto terá definir.

Por fim, teremos assim duas grandes fases:
1. Escopo deste projeto. Ideialização/especificação detalhada de tudo o que precisamos para implementar a solução e homoloogar no dict;
2. Implementação, escopo do projeto que este projecto irºa definir,

Quaisquer duvidas devem ser colocadas num documento de duvidas para que possamos ir respndendo e atualizando o contexto.

O gestor deste projeto (agente claude code) tem comno missão de manter atualizado todos os documentos que este projeto tem que gerar, bem cmo a gestao de todo o backlog.
