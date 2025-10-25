# Jornadas de Usuário

**Propósito**: Mapeamento de jornadas completas de usuário no sistema DICT

## 📋 Conteúdo

Esta pasta armazenará:

- **User Journeys**: Jornadas end-to-end de usuários
- **User Personas**: Perfis de usuários típicos (correntista, empresa, etc.)
- **Pain Points**: Pontos de dor identificados em cada jornada
- **Opportunity Maps**: Oportunidades de melhoria da experiência

## 📁 Estrutura Esperada

```
Jornadas/
├── Personas/
│   ├── Persona_Correntista_PF.md
│   ├── Persona_Empresa_PJ.md
│   └── Persona_Admin_Banco.md
├── Journeys/
│   ├── Journey_CreateEntry_First_Time.md
│   ├── Journey_Claim_Ownership.md
│   ├── Journey_Portability.md
│   └── Journey_Delete_Entry.md
└── Pain_Points/
    └── Analysis_Pain_Points.md
```

## 🎯 Exemplo: Journey - Criar Primeira Chave DICT

```markdown
# Jornada: Criar Primeira Chave DICT (Pessoa Física)

## Persona
**Nome**: João Silva
**Idade**: 35 anos
**Ocupação**: Freelancer
**Objetivo**: Receber pagamentos via PIX usando CPF

## Jornada

### 1. Descoberta (Awareness)
**Canal**: App do banco
**Ação**: João vê banner "Cadastre sua chave PIX"
**Pensamento**: "O que é PIX? Vale a pena?"

### 2. Consideração
**Canal**: FAQ do banco
**Ação**: João lê sobre benefícios do PIX
**Pensamento**: "Parece mais fácil do que passar dados da conta"

### 3. Cadastro
**Canal**: App do banco (tela de criar chave)
**Ação**: João clica em "Cadastrar chave PIX"

**Passos**:
1. Escolhe tipo de chave: CPF
2. Sistema preenche CPF automaticamente (já autenticado)
3. Seleciona conta corrente
4. Confirma dados
5. Aguarda confirmação

**Pain Points**:
- ❌ Não entende diferença entre tipos de chave
- ❌ Não sabe se pode ter múltiplas chaves
- ❌ Incerteza sobre tempo de processamento

**Soluções UX**:
- ✅ Tooltip explicando cada tipo de chave
- ✅ Mensagem "Você pode cadastrar até 5 chaves"
- ✅ Feedback "Processamento em até 10 segundos"

### 4. Confirmação
**Canal**: Notificação push + email
**Ação**: João recebe confirmação de chave criada
**Pensamento**: "Pronto! Agora posso receber pagamentos"

### 5. Uso (First Transaction)
**Canal**: Enviar chave para pagador
**Ação**: João compartilha CPF com cliente
**Pensamento**: "Muito mais fácil do que passar agência e conta"

### 6. Retenção
**Canal**: App (ver minhas chaves)
**Ação**: João verifica chaves cadastradas
**Potencial**: Cadastrar mais chaves (email, telefone)

## Métricas de Sucesso

- **Time to First Key**: < 2 minutos (meta)
- **Completion Rate**: > 90%
- **User Satisfaction**: NPS > 70
- **Support Tickets**: < 5% dos cadastros
```

## 🗺️ Customer Journey Map

```
Etapa:     [Descoberta] → [Consideração] → [Cadastro] → [Confirmação] → [Uso]
Emoção:         🤔      →       😊       →     😐     →      😃       →  🎉
Touchpoint:     App     →       FAQ      →     Form   →   Push Notif  →  App
Pain Point:     ?       →        -       →  Confusão →       -        →   -
Oportunidade:  Banner   →    Educação    → Tooltips  →   Onboarding   → Upsell
```

## 📊 Personas

### Persona 1: Correntista Pessoa Física
- **Objetivo**: Receber pagamentos de clientes/amigos via PIX
- **Tech-savvy**: Médio
- **Frequência de uso**: Semanal

### Persona 2: Empresa (PJ)
- **Objetivo**: Receber pagamentos de clientes B2C
- **Tech-savvy**: Alto (tem contador/admin)
- **Frequência de uso**: Diária
- **Necessidades**: Múltiplas chaves, relatórios, API

### Persona 3: Admin do Banco
- **Objetivo**: Gerenciar chaves DICT dos clientes, resolver claims
- **Tech-savvy**: Alto
- **Frequência de uso**: Diária
- **Necessidades**: Dashboard, busca avançada, auditoria

## 📚 Referências

- [Wireframes](../Wireframes/)
- [User Stories](../../01_Requisitos/UserStories/)
- [Componentes](../Componentes/)

---

**Status**: 🔴 Pasta vazia (será preenchida na Fase 3)
**Fase de Preenchimento**: Fase 3 (discovery UX)
**Responsável**: UX Designer + Product Owner
**Método**: Entrevistas com usuários, análise de concorrentes
