Este projeto será criado tendo como base o repo github.com/lb-conn/connector-dict que precisamos clonar para dentro desta pasta e em seguinda criar um branch com o nome balde_dict. Todos os desenvolvimentos desta squad neste projeto terão que ser nesta branch.

Precisa de entender o escopo do projeto em primeiro lugar baseado na RF_dict_Bacen.md e considerar que o escopo será implementado como uma nova funcionalidade do Connector-Dict.

Criar a Squad de agentes técnicos mais indicada para este projeto, escolhendo os agentes, os plugins e seguindo as melhores praticas do manual agents/claude-code-agent-squad-guide.md.

Teremos que criar um claude.md e um specs.md com toda a especificação técnica e funcional do projeto e como o projeto será implementado.

Tópicos de contexto importantes:

Deverão ser considerados como referência para a especificação de arquitetura, funcional e técnica os seguintes RF_Dict_Bacen.md, instruções_app_dict.md e arquitecto_Stacholder.md.


Abaixo do .claude:
O documento de claude.md deve conter todos os requisitos do projeto incluindo as responsabilidades dos agentes, planeamentos toda a organização do projeto.

O documento specs.md deverá conter toda a especificação de arquitetura e stack de arquitetura, tendo como referência a análise ao documento 

github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/bacen/
github.com/lb-conn/connector-dict/apps/dict/application/

Crie um documento de duvidas.md para que possamos esclarecer as duvidas antes de iniciarmos a implementação.

Fundamental a criação da branch balde_dict após o clone dentro da raiz deste projeto do repo github.com/lb-conn/connector-dict

A App solução deste escopo que será implementada não irá integrar com o bacen directamente mas apenas através da APP Bridge. Para isso é importante ler atentamente os três arquivos .md abaixo da folder specs_do_Stackholder.

Espero que consiga abstrair das imagens de arquitetura (duas abaixo da folder .claude/images) requisitos de arquitetura especificos. 

Muito importante, considere todas as definições de agents, commands e plugins abaixo da folder .claude conforme a relevância que tiver para cada parte do contexto do projeto.

Por ultimo, think harder and deply, de forma o mais meticolosamente possivel os dois documentos de arquitetura e de especificação do projeto. Deverá ficar super bem definido qual o papel que cada sub agente terá na implementação deste projeto. Deverá encontrar nos documentos do arquitecto stacholder e nos três documentos de instruções os elementos para definição da stack de tecnologia mantendo a stack do repor connectior-dict.