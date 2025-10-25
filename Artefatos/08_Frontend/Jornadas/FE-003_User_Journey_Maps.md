# FE-003: User Journey Maps - DICT Operations

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Documento**: FE-003 - User Journey Maps
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: FRONTEND (AI Agent - Frontend Specialist)
**Revisor**: [Aguardando]
**Aprovador**: Tech Lead, UX Designer

---

## Sumário Executivo

Este documento apresenta os **mapas de jornada do usuário** (User Journey Maps) para os principais fluxos do sistema DICT LBPay, identificando emoções, pain points e oportunidades de melhoria em cada etapa da experiência do usuário.

**Baseado em**:
- [FE-001: Component Specifications](../Componentes/FE-001_Component_Specifications.md)
- [FE-002: Wireframes DICT Operations](../Wireframes/FE-002_Wireframes_DICT_Operations.md)
- [API-002: Core DICT REST API](../../04_APIs/REST/API-002_Core_DICT_REST_API.md)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | FRONTEND | Versão inicial - User Journey Maps |

---

## Índice

1. [Metodologia](#1-metodologia)
2. [Personas](#2-personas)
3. [Journey 1: Criar Chave PIX](#3-journey-1-criar-chave-pix)
4. [Journey 2: Responder à Reivindicação](#4-journey-2-responder-à-reivindicação)
5. [Journey 3: Acompanhar Status da Reivindicação](#5-journey-3-acompanhar-status-da-reivindicação)
6. [Emotional Journey Graphs](#6-emotional-journey-graphs)
7. [Key Insights & Opportunities](#7-key-insights--opportunities)

---

## 1. Metodologia

### 1.1. Framework Utilizado

Este documento utiliza o framework de **User Journey Mapping** que inclui:

1. **Persona**: Quem é o usuário?
2. **Objetivo**: O que o usuário quer alcançar?
3. **Fases**: Etapas da jornada (Awareness, Consideration, Decision, Action, Retention)
4. **Ações**: O que o usuário faz em cada etapa?
5. **Pensamentos**: O que o usuário está pensando?
6. **Emoções**: Como o usuário se sente? (escala de 😊 a 😡)
7. **Pain Points**: Pontos de fricção e frustração
8. **Opportunities**: Oportunidades de melhoria

### 1.2. Emotional Scale

```
😊  Muito Satisfeito   (5/5)
🙂  Satisfeito         (4/5)
😐  Neutro             (3/5)
😕  Insatisfeito       (2/5)
😡  Muito Insatisfeito (1/5)
```

---

## 2. Personas

### 2.1. Persona 1: Maria Silva

**Perfil**:
- **Nome**: Maria Silva
- **Idade**: 35 anos
- **Ocupação**: Autônoma (Designer Gráfica)
- **Localização**: São Paulo, SP
- **Tecnologia**: Média familiaridade (usa smartphone diariamente, mas não é tech-savvy)
- **Comportamento**: Prefere transações rápidas, valoriza segurança, teme cometer erros

**Objetivo**: Criar uma chave PIX CPF para receber pagamentos de clientes de forma rápida e prática.

**Motivações**:
- Facilitar recebimentos de clientes
- Evitar erros ao passar dados bancários
- Aumentar profissionalismo

**Frustrações**:
- Processos bancários complexos
- Medo de cometer erros irreversíveis
- Falta de clareza sobre prazos e consequências

---

### 2.2. Persona 2: João Santos

**Perfil**:
- **Nome**: João Santos
- **Idade**: 42 anos
- **Ocupação**: Gerente de Vendas
- **Localização**: Rio de Janeiro, RJ
- **Tecnologia**: Alta familiaridade (early adopter, usa diversos apps financeiros)
- **Comportamento**: Quer controle total, lê termos e condições, valoriza transparência

**Objetivo**: Transferir sua chave PIX CPF de um banco para outro após trocar de instituição financeira.

**Motivações**:
- Consolidar contas em um único banco
- Manter o mesmo número/email/CPF para PIX
- Evitar informar mudanças para contatos

**Frustrações**:
- Processos longos e burocráticos
- Falta de visibilidade sobre o andamento
- Comunicação confusa entre bancos

---

### 2.3. Persona 3: Ana Costa

**Perfil**:
- **Nome**: Ana Costa
- **Idade**: 28 anos
- **Ocupação**: Desenvolvedora de Software
- **Localização**: Belo Horizonte, MG
- **Tecnologia**: Muito alta (tech expert, contribui para open source)
- **Comportamento**: Busca transparência técnica, lê documentação, valoriza APIs

**Objetivo**: Entender e gerenciar reivindicações de chaves PIX de forma proativa.

**Motivações**:
- Controle sobre seus dados
- Entendimento completo do processo
- Segurança e auditabilidade

**Frustrações**:
- Interfaces que escondem detalhes técnicos
- Falta de logs e histórico
- Processos "black box"

---

## 3. Journey 1: Criar Chave PIX

### 3.1. Persona: Maria Silva

**Objetivo**: Criar chave PIX CPF para receber pagamentos de clientes

**Contexto**: Maria acabou de abrir conta no LBPay e quer criar sua primeira chave PIX.

---

### Fase 1: Awareness (Conscientização)

**Ação**: Maria ouve de um cliente que ele prefere pagar via PIX.

| Aspecto | Descrição |
|---------|-----------|
| **Pensamento** | "O que é PIX? Como eu crio isso?" |
| **Emoção** | 😐 Neutro (3/5) - Curiosa, mas incerta |
| **Touchpoint** | Conversa com cliente, busca no Google |
| **Pain Points** | • Não sabe por onde começar<br>• Medo de ser complicado |
| **Opportunities** | • Tutorial/Onboarding no app<br>• FAQ sobre PIX<br>• Vídeo explicativo curto (30s) |

---

### Fase 2: Consideration (Consideração)

**Ação**: Maria faz login no app LBPay e procura onde criar chave PIX.

| Aspecto | Descrição |
|---------|-----------|
| **Pensamento** | "Cadê a opção de criar chave PIX? Será que é aqui?" |
| **Emoção** | 😕 Insatisfeito (2/5) - Confusa, navegação não é clara |
| **Touchpoint** | Dashboard do app, menu de navegação |
| **Pain Points** | • Menu não é intuitivo<br>• Botão "Criar Chave" não está visível<br>• Muitas opções no menu |
| **Opportunities** | • Botão de destaque no dashboard: "Criar Minha Primeira Chave PIX"<br>• Tooltip: "Ainda não tem chave PIX? Crie agora em 2 minutos"<br>• Search bar no menu |

---

### Fase 3: Decision (Decisão)

**Ação**: Maria encontra a opção "Minhas Chaves PIX" e clica em "+ Nova Chave".

| Aspecto | Descrição |
|---------|-----------|
| **Pensamento** | "Finalmente! Qual tipo de chave eu devo escolher?" |
| **Emoção** | 🙂 Satisfeito (4/5) - Aliviada por encontrar |
| **Touchpoint** | Tela "Criar Nova Chave PIX" |
| **Pain Points** | • Muitas opções de tipo (CPF, CNPJ, Email, Telefone, EVP)<br>• Não sabe qual escolher<br>• Medo de escolher errado |
| **Opportunities** | • Descrição clara de cada tipo de chave<br>• Recomendação: "Recomendado para você: CPF"<br>• Ícone de ajuda (?) com tooltip explicativo<br>• Comparação lado a lado: "CPF vs Email vs Telefone" |

---

### Fase 4: Action (Ação)

**Ação**: Maria seleciona "CPF", digita seu CPF, seleciona a conta e clica em "Criar Chave".

| Aspecto | Descrição |
|---------|-----------|
| **Pensamento** | "Será que vai dar certo? Digitei certo?" |
| **Emoção** | 😐 Neutro (3/5) - Ansiosa, esperando confirmação |
| **Touchpoint** | Formulário de criação, botão "Criar Chave" |
| **Pain Points** | • Validação de CPF não é instantânea<br>• Não há preview antes de criar<br>• Loading muito longo (3-5s)<br>• Mensagem de "período de 30 dias" assusta |
| **Opportunities** | • Validação em tempo real (formato de CPF)<br>• Preview: "Sua chave será: 123.456.789-00"<br>• Progress indicator durante criação<br>• Explicar melhor o período de 30 dias: "Durante 30 dias, outras pessoas podem reivindicar esta chave se ela já pertenceu a elas. Isso é uma proteção de segurança do Banco Central." |

---

### Fase 5: Retention (Retenção)

**Ação**: Maria recebe confirmação de que a chave foi criada com sucesso.

| Aspecto | Descrição |
|---------|-----------|
| **Pensamento** | "Oba! Consegui! Agora posso passar para meus clientes!" |
| **Emoção** | 😊 Muito Satisfeito (5/5) - Feliz, aliviada |
| **Touchpoint** | Tela de confirmação, Dashboard atualizado |
| **Pain Points** | • Não sabe o próximo passo<br>• Não sabe como compartilhar a chave |
| **Opportunities** | • Botão "Compartilhar minha chave PIX" (gera QR code + texto)<br>• Tutorial: "Agora você já pode receber pagamentos!"<br>• Notificação push: "Sua chave PIX está ativa e pronta para uso!"<br>• Gamification: Badge "Primeira chave PIX criada!" |

---

### 3.2. Emotional Journey Graph - Journey 1

```
Emoção
  5 😊  │                                         ●
       │
  4 🙂  │                           ●
       │
  3 😐  │     ●
       │                               ●
  2 😕  │           ●
       │
  1 😡  │
       └─────┬────────┬────────┬────────┬────────┬──> Fase
         Awareness  Consideration  Decision  Action  Retention
```

---

## 4. Journey 2: Responder à Reivindicação

### 4.1. Persona: João Santos

**Objetivo**: Entender e responder a uma reivindicação de chave PIX que recebeu.

**Contexto**: João recebeu uma notificação de que alguém está reivindicando sua chave PIX CPF. Ele precisa decidir se confirma ou cancela a reivindicação.

---

### Fase 1: Awareness (Conscientização)

**Ação**: João recebe notificação push: "Alguém reivindicou sua chave PIX CPF".

| Aspecto | Descrição |
|---------|-----------|
| **Pensamento** | "O que?! Quem está tentando pegar minha chave? É fraude?" |
| **Emoção** | 😡 Muito Insatisfeito (1/5) - Assustado, desconfiado |
| **Touchpoint** | Notificação push, Email |
| **Pain Points** | • Notificação alarmante sem contexto<br>• Medo de fraude<br>• Não sabe o que é reivindicação |
| **Opportunities** | • Notificação mais clara: "Reivindicação recebida - Você tem 30 dias para responder"<br>• Explicar contexto: "Isso acontece quando você cadastra uma chave que já foi sua em outro banco"<br>• Link direto para tutorial: "O que é reivindicação?" |

---

### Fase 2: Consideration (Consideração)

**Ação**: João abre o app e vai até "Reivindicações" para entender o que está acontecendo.

| Aspecto | Descrição |
|---------|-----------|
| **Pensamento** | "Deixa eu ver quem é e por que está reivindicando" |
| **Emoção** | 😕 Insatisfeito (2/5) - Ainda desconfiado, mas curioso |
| **Touchpoint** | Tela "Reivindicações", Card da reivindicação |
| **Pain Points** | • Informações técnicas (ISPB) não fazem sentido<br>• Não consegue ver quem é o reivindicante (apenas ISPB)<br>• Falta contexto histórico |
| **Opportunities** | • Traduzir ISPB para nome do banco: "Banco XYZ (ISPB: 12345678)"<br>• Mostrar histórico: "Esta chave foi cadastrada por você no Banco XYZ em 10/01/2024"<br>• Timeline visual: "Linha do tempo da sua chave PIX" |

---

### Fase 3: Decision (Decisão)

**Ação**: João lê os detalhes e decide que a reivindicação é legítima (ele realmente tinha essa chave no outro banco antes).

| Aspecto | Descrição |
|---------|-----------|
| **Pensamento** | "Ah, entendi. É do meu banco antigo. Posso confirmar." |
| **Emoção** | 🙂 Satisfeito (4/5) - Aliviado, confiante |
| **Touchpoint** | Tela "Detalhes da Reivindicação" |
| **Pain Points** | • Ainda não está 100% seguro<br>• Não sabe as consequências de confirmar |
| **Opportunities** | • Checklist: "Confirme se:"<br>  - ✅ Você reconhece o banco reivindicante<br>  - ✅ Você já teve conta neste banco<br>  - ✅ Você quer transferir esta chave<br>• Warning claro: "Ao confirmar, esta chave será transferida e você NÃO poderá desfazer" |

---

### Fase 4: Action (Ação)

**Ação**: João clica em "Responder", seleciona "Confirmar Transferência", escreve o motivo e confirma.

| Aspecto | Descrição |
|---------|-----------|
| **Pensamento** | "Vou escrever que estou trocando de banco. Pronto." |
| **Emoção** | 😐 Neutro (3/5) - Um pouco ansioso, esperando confirmação |
| **Touchpoint** | Formulário "Responder Reivindicação" |
| **Pain Points** | • Campo "Motivo" obrigatório parece burocrático<br>• Validação de "mínimo 10 caracteres" é frustrante<br>• Loading longo ao confirmar |
| **Opportunities** | • Explicar por que o motivo é necessário: "O Banco Central exige um motivo para fins de auditoria"<br>• Sugestões de motivo: "Troca de instituição financeira", "Consolidação de contas"<br>• Feedback visual imediato: "Confirmando..." com progress bar |

---

### Fase 5: Retention (Retenção)

**Ação**: João recebe confirmação de que a reivindicação foi respondida e a transferência será concluída em até 24 horas.

| Aspecto | Descrição |
|---------|-----------|
| **Pensamento** | "Ok, tudo certo. Vou receber confirmação quando terminar?" |
| **Emoção** | 🙂 Satisfeito (4/5) - Aliviado, mas quer acompanhar |
| **Touchpoint** | Tela de confirmação, Email de confirmação |
| **Pain Points** | • Não sabe como acompanhar o andamento<br>• "Até 24 horas" é vago<br>• Medo de não receber notificação quando terminar |
| **Opportunities** | • Status visual: "Etapa 1 de 3: Confirmação enviada ao Banco Central"<br>• Notificação push quando cada etapa for concluída<br>• Link para "Acompanhar status em tempo real"<br>• Email de resumo: "Sua reivindicação foi confirmada - Saiba o que acontece agora" |

---

### 4.2. Emotional Journey Graph - Journey 2

```
Emoção
  5 😊  │
       │
  4 🙂  │                     ●                       ●
       │
  3 😐  │                                 ●
       │
  2 😕  │           ●
       │
  1 😡  │     ●
       └─────┬────────┬────────┬────────┬────────┬──> Fase
         Awareness  Consideration  Decision  Action  Retention
```

---

## 5. Journey 3: Acompanhar Status da Reivindicação

### 5.1. Persona: Ana Costa

**Objetivo**: Acompanhar o status de uma reivindicação que ela iniciou para recuperar sua chave PIX de um banco anterior.

**Contexto**: Ana cadastrou uma chave PIX CPF no LBPay, mas essa chave já estava registrada em outro banco. Ela iniciou uma reivindicação há 15 dias e quer saber o status.

---

### Fase 1: Awareness (Conscientização)

**Ação**: Ana lembra que iniciou a reivindicação há 15 dias e decide verificar o status.

| Aspecto | Descrição |
|---------|-----------|
| **Pensamento** | "Já faz 15 dias. Cadê a resposta? Será que o outro banco vai responder?" |
| **Emoção** | 😐 Neutro (3/5) - Curiosa, mas levemente impaciente |
| **Touchpoint** | Lembrete mental, Notificação (ou falta dela) |
| **Pain Points** | • Não recebeu nenhuma notificação de progresso<br>• Esqueceu que havia iniciado a reivindicação<br>• Não sabe se precisa fazer algo |
| **Opportunities** | • Notificações proativas: "Sua reivindicação está há 15 dias em andamento"<br>• Digest semanal: "Resumo semanal das suas reivindicações"<br>• Widget no dashboard: "Reivindicações em andamento: 1 (15 dias restantes)" |

---

### Fase 2: Consideration (Consideração)

**Ação**: Ana abre o app e vai até "Reivindicações" para ver o status.

| Aspecto | Descrição |
|---------|-----------|
| **Pensamento** | "Onde está o status? O que significa 'Aguardando Resposta'?" |
| **Emoção** | 😕 Insatisfeito (2/5) - Frustrada pela falta de informação |
| **Touchpoint** | Tela "Reivindicações", Card da reivindicação |
| **Pain Points** | • Status "Aguardando Resposta" é genérico demais<br>• Não sabe se o outro banco foi notificado<br>• Não sabe quanto tempo falta |
| **Opportunities** | • Status detalhado: "Aguardando resposta do proprietário atual (Banco XYZ)"<br>• Progress bar: "15 de 30 dias (50%)"<br>• Última atualização: "Notificação enviada há 15 dias" |

---

### Fase 3: Decision (Decisão)

**Ação**: Ana clica em "Ver Detalhes" para ter mais informações.

| Aspecto | Descrição |
|---------|-----------|
| **Pensamento** | "Preciso de mais detalhes técnicos. Quero logs, datas, IDs." |
| **Emoção** | 😕 Insatisfeito (2/5) - Quer transparência técnica |
| **Touchpoint** | Tela "Detalhes da Reivindicação" |
| **Pain Points** | • Falta de logs detalhados<br>• IDs técnicos (UUID) não são exibidos<br>• Não há histórico de eventos |
| **Opportunities** | • Seção "Detalhes Técnicos" (toggle opcional):<br>  - Claim ID (UUID)<br>  - External ID (Bacen)<br>  - Timestamps de cada evento<br>• Timeline detalhada:<br>  - 10/10 14:30 - Reivindicação criada<br>  - 10/10 14:31 - Enviado ao Bacen<br>  - 10/10 14:35 - Banco proprietário notificado<br>  - 25/10 09:00 - Aguardando resposta (15 dias) |

---

### Fase 4: Action (Ação)

**Ação**: Ana decide aguardar mais alguns dias antes de tomar qualquer ação (cancelar ou entrar em contato).

| Aspecto | Descrição |
|---------|-----------|
| **Pensamento** | "Vou esperar até o dia 30. Se não responderem, vou cancelar." |
| **Emoção** | 😐 Neutro (3/5) - Paciente, mas monitorando |
| **Touchpoint** | Tela "Detalhes", Notificações futuras |
| **Pain Points** | • Não há opção de "lembrete"<br>• Não sabe o que acontece se ninguém responder em 30 dias |
| **Opportunities** | • Botão "Definir lembrete": "Me avise quando faltar X dias"<br>• FAQ inline: "O que acontece se ninguém responder em 30 dias?"<br>• Resposta: "A reivindicação será automaticamente cancelada e a chave permanecerá com o proprietário atual" |

---

### Fase 5: Retention (Retenção)

**Ação**: Ana recebe notificação no dia 28 (2 dias antes do prazo) informando que ainda não houve resposta.

| Aspecto | Descrição |
|---------|-----------|
| **Pensamento** | "Ainda não responderam. Vou cancelar e tentar outro método." |
| **Emoção** | 😕 Insatisfeito (2/5) - Frustrada, mas compreensiva |
| **Touchpoint** | Notificação push, Tela "Reivindicações" |
| **Pain Points** | • Processo de 30 dias parece longo demais<br>• Falta de comunicação do outro banco<br>• Não sabe se há alternativa |
| **Opportunities** | • Notificação proativa: "Sua reivindicação expira em 2 dias. Quer cancelar ou aguardar?"<br>• Sugestão: "Enquanto aguarda, você pode usar uma chave aleatória (EVP) temporária"<br>• Botão de suporte: "Falar com atendente" para casos urgentes |

---

### 5.2. Emotional Journey Graph - Journey 3

```
Emoção
  5 😊  │
       │
  4 🙂  │
       │
  3 😐  │     ●                               ●
       │
  2 😕  │           ●         ●                       ●
       │
  1 😡  │
       └─────┬────────┬────────┬────────┬────────┬──> Fase
         Awareness  Consideration  Decision  Action  Retention
```

---

## 6. Emotional Journey Graphs

### 6.1. Comparative Emotional Journey

```
Emoção
  5 😊  │                                         ● (Journey 1)
       │
  4 🙂  │                           ●             ● (Journey 2)
       │
  3 😐  │     ●     ●         ●         ●         ● (Journey 3)
       │
  2 😕  │           ●         ●   ●               ●
       │
  1 😡  │     ● (Journey 2)
       └─────┬────────┬────────┬────────┬────────┬──> Fase
         Awareness  Consideration  Decision  Action  Retention

Legenda:
● Journey 1: Criar Chave PIX (Maria - Autônoma)
● Journey 2: Responder Reivindicação (João - Gerente)
● Journey 3: Acompanhar Status (Ana - Dev)
```

### 6.2. Insights

**Journey 1 (Criar Chave PIX)**:
- Maior pico de satisfação (5/5) ao final
- Vale emocional na fase de Consideration (dificuldade de encontrar opção)
- Trajetória positiva após encontrar a funcionalidade

**Journey 2 (Responder Reivindicação)**:
- Maior vale emocional (1/5) no início (notificação alarmante)
- Recuperação gradual conforme entende o processo
- Termina satisfeito, mas não encantado

**Journey 3 (Acompanhar Status)**:
- Jornada mais plana (oscila entre 2/5 e 3/5)
- Nunca atinge satisfação alta
- Frustração persistente com falta de transparência

---

## 7. Key Insights & Opportunities

### 7.1. Pain Points Críticos

| Pain Point | Frequência | Severidade | Prioridade |
|------------|------------|------------|------------|
| Notificação de reivindicação alarmante | Alta (J2) | Crítica | 🔴 P0 |
| Falta de contexto sobre "30 dias" | Alta (J1, J3) | Alta | 🟠 P1 |
| Status genérico sem detalhes | Média (J3) | Média | 🟡 P2 |
| ISPB não traduzido para nome do banco | Alta (J2) | Média | 🟡 P2 |
| Navegação confusa no menu | Média (J1) | Média | 🟡 P2 |
| Loading longo sem feedback | Alta (J1, J2) | Baixa | 🟢 P3 |

### 7.2. Oportunidades de Impacto Rápido (Quick Wins)

1. **Melhorar notificação de reivindicação** (P0)
   - Antes: "Alguém reivindicou sua chave PIX CPF"
   - Depois: "Reivindicação recebida - Você tem 30 dias para responder. Saiba mais."

2. **Traduzir ISPB para nome do banco** (P1)
   - Adicionar mapping ISPB → Nome do banco
   - Exibir: "Banco XYZ (ISPB: 12345678)"

3. **Explicar período de 30 dias** (P1)
   - Adicionar tooltip inline: "Por que 30 dias?"
   - Resposta: "É uma proteção de segurança do Banco Central..."

4. **Adicionar progress bar visual** (P2)
   - Mostrar progresso: "15 de 30 dias (50%)"
   - Indicador visual de urgência nos últimos 5 dias

5. **Botão destaque no dashboard** (P2)
   - "Criar Minha Primeira Chave PIX" (se usuário não tem chaves)

### 7.3. Oportunidades de Médio Prazo

1. **Timeline visual detalhada**
   - Histórico completo de eventos de cada chave/reivindicação
   - Timestamps, ações, responsáveis

2. **Notificações proativas**
   - Digest semanal de reivindicações em andamento
   - Alertas quando faltam X dias para expiração

3. **Comparação de tipos de chave**
   - Tabela: "CPF vs Email vs Telefone vs EVP"
   - Recomendação inteligente baseada no perfil do usuário

4. **Detalhes técnicos opcionais**
   - Toggle "Modo avançado" para usuários tech-savvy
   - Exibir IDs, logs, timestamps

5. **Gamification**
   - Badges: "Primeira chave criada", "5 chaves ativas"
   - Progress bar: "Complete seu perfil PIX"

### 7.4. Oportunidades de Longo Prazo

1. **Assistente virtual / Chatbot**
   - Responder dúvidas em tempo real
   - Guiar usuário pelo processo de reivindicação

2. **Predição de comportamento**
   - Alertar usuário se reivindicação tem baixa chance de ser respondida
   - Sugerir ações alternativas

3. **Integração com calendário**
   - Adicionar lembretes de reivindicações ao Google Calendar

4. **Dashboard analytics**
   - Gráficos de uso: "Você tem 5 chaves, 3 estão ativas, 2 pendentes"
   - Comparação com média de usuários

---

## Rastreabilidade

### Requisitos de UX

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RF-UJ-001 | Journey Map - Criar Chave | FE-001, FE-002 | ✅ Especificado |
| RF-UJ-002 | Journey Map - Responder Reivindicação | FE-001, FE-002 | ✅ Especificado |
| RF-UJ-003 | Journey Map - Acompanhar Status | FE-001, FE-002 | ✅ Especificado |
| RF-UJ-004 | Pain Points identificados | User Research | ✅ Especificado |
| RF-UJ-005 | Opportunities documentadas | User Research | ✅ Especificado |

---

## Próximas Revisões

**Pendências**:
- [ ] Validar journey maps com usuários reais (testes de usabilidade)
- [ ] Adicionar journey maps para usuários PJ (CNPJ)
- [ ] Criar journey maps para casos de erro (falha de API, timeout)
- [ ] Mapear touchpoints em outros canais (SMS, email, push)
- [ ] Priorizar oportunidades com base em ROI

---

**Referências**:
- [FE-001: Component Specifications](../Componentes/FE-001_Component_Specifications.md)
- [FE-002: Wireframes DICT Operations](../Wireframes/FE-002_Wireframes_DICT_Operations.md)
- [API-002: Core DICT REST API](../../04_APIs/REST/API-002_Core_DICT_REST_API.md)
- [Nielsen Norman Group: Journey Mapping 101](https://www.nngroup.com/articles/journey-mapping-101/)
- [UX Booth: User Journey Mapping](https://www.uxbooth.com/articles/user-journey-mapping/)
