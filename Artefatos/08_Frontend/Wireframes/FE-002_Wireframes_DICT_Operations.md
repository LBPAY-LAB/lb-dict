# FE-002: Wireframes - DICT Operations

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Documento**: FE-002 - Wireframes DICT Operations
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: FRONTEND (AI Agent - Frontend Specialist)
**Revisor**: [Aguardando]
**Aprovador**: Tech Lead, UX Designer

---

## Sumário Executivo

Este documento apresenta os **wireframes** das telas do sistema DICT LBPay, incluindo layouts desktop (1920x1080) e mobile (375x667), fluxos de navegação e especificações visuais. Os wireframes seguem os princípios de UX/UI e acessibilidade WCAG 2.1 AA.

**Baseado em**:
- [API-002: Core DICT REST API](../../04_APIs/REST/API-002_Core_DICT_REST_API.md)
- [FE-001: Component Specifications](../Componentes/FE-001_Component_Specifications.md)
- [DIA-006: Sequence Claim Workflow](../../02_Arquitetura/Diagramas/DIA-006_Sequence_Claim_Workflow.md)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | FRONTEND | Versão inicial - Wireframes DICT Operations |

---

## Índice

1. [Design System](#1-design-system)
2. [Navigation Flow](#2-navigation-flow)
3. [Dashboard Screen](#3-dashboard-screen)
4. [Create Key Screen](#4-create-key-screen)
5. [Key Details Screen](#5-key-details-screen)
6. [Claims List Screen](#6-claims-list-screen)
7. [Claim Details Screen](#7-claim-details-screen)
8. [Respond to Claim Screen](#8-respond-to-claim-screen)

---

## 1. Design System

### 1.1. Colors

```
Primary:    #2563EB (Blue 600)
Secondary:  #64748B (Slate 500)
Success:    #10B981 (Green 500)
Warning:    #F59E0B (Amber 500)
Error:      #EF4444 (Red 500)
Neutral:    #F1F5F9 (Slate 100)
Text:       #0F172A (Slate 900)
```

### 1.2. Typography

```
Headings:   Inter, sans-serif (700 weight)
Body:       Inter, sans-serif (400 weight)
Mono:       JetBrains Mono (code/numbers)

H1: 32px (2rem)
H2: 24px (1.5rem)
H3: 20px (1.25rem)
Body: 16px (1rem)
Small: 14px (0.875rem)
```

### 1.3. Spacing

```
xs:  4px   (0.25rem)
sm:  8px   (0.5rem)
md:  16px  (1rem)
lg:  24px  (1.5rem)
xl:  32px  (2rem)
2xl: 48px  (3rem)
```

### 1.4. Icons

```
Library: Lucide React / Heroicons
Size: 20px (default), 24px (large)
Stroke: 2px
```

---

## 2. Navigation Flow

### 2.1. Main Navigation Structure

```
┌─────────────────────────────────────────┐
│ Home                                     │
├─────────────────────────────────────────┤
│ ├─ Dashboard                             │
│ ├─ Minhas Chaves PIX                     │
│ │   ├─ Listar Chaves                     │
│ │   ├─ Criar Nova Chave                  │
│ │   └─ Detalhes da Chave                 │
│ ├─ Reivindicações                        │
│ │   ├─ Listar Reivindicações             │
│ │   ├─ Detalhes da Reivindicação         │
│ │   └─ Responder Reivindicação           │
│ ├─ Portabilidade                         │
│ │   ├─ Listar Portabilidades             │
│ │   ├─ Iniciar Portabilidade             │
│ │   └─ Detalhes da Portabilidade         │
│ └─ Configurações                         │
└─────────────────────────────────────────┘
```

### 2.2. User Flow Diagram

```
[Login] → [Dashboard] → [Minhas Chaves]
                     ↓
              [Criar Chave] → [Confirmação] → [Dashboard]
                     ↓
              [Reivindicações] → [Detalhes] → [Responder] → [Confirmação]
                     ↓
              [Portabilidade] → [Iniciar] → [Confirmação]
```

---

## 3. Dashboard Screen

### 3.1. Desktop View (1920x1080)

```
┌────────────────────────────────────────────────────────────────────────────┐
│ ┌────────┐                                      [Perfil ▼] [Notif 🔔]      │
│ │ LBPay  │  Dashboard   Chaves   Reivindicações   Portabilidade           │
│ └────────┘                                                                 │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │  Bem-vindo, João Silva                                                │ │
│  │  Aqui está o resumo das suas chaves PIX                               │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  ┌────────────────────┐  ┌────────────────────┐  ┌────────────────────┐  │
│  │  Chaves Ativas      │  │  Reivindicações    │  │  Portabilidades    │  │
│  │                     │  │                     │  │                     │  │
│  │       5             │  │       2             │  │       0             │  │
│  │                     │  │                     │  │                     │  │
│  │  [Ver Todas]        │  │  [Ver Todas]        │  │  [Ver Todas]        │  │
│  └────────────────────┘  └────────────────────┘  └────────────────────┘  │
│                                                                             │
│  Minhas Chaves PIX                                      [+ Nova Chave]     │
│  ────────────────────────────────────────────────────────────────────────  │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │  CPF: 123.456.789-00                                   ✅ ATIVA       │ │
│  │  Conta: Corrente - 0001/12345678                                      │ │
│  │  Criado: 25/10/2025                              [Ver] [Deletar]      │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │  Email: joao.silva@email.com                           ✅ ATIVA       │ │
│  │  Conta: Corrente - 0001/12345678                                      │ │
│  │  Criado: 20/10/2025                              [Ver] [Deletar]      │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │  Telefone: +55 11 99999-9999                           ✅ ATIVA       │ │
│  │  Conta: Poupança - 0001/87654321                                      │ │
│  │  Criado: 15/10/2025                              [Ver] [Deletar]      │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  [Anterior]                              Página 1 de 2            [Próxima]│
│                                                                             │
└────────────────────────────────────────────────────────────────────────────┘
```

### 3.2. Mobile View (375x667)

```
┌───────────────────────┐
│ ☰  Dashboard    🔔    │
├───────────────────────┤
│                       │
│ Bem-vindo,            │
│ João Silva            │
│                       │
│ ┌───────────────────┐ │
│ │ Chaves Ativas     │ │
│ │       5           │ │
│ └───────────────────┘ │
│                       │
│ ┌───────────────────┐ │
│ │ Reivindicações    │ │
│ │       2           │ │
│ └───────────────────┘ │
│                       │
│ ┌───────────────────┐ │
│ │ Portabilidades    │ │
│ │       0           │ │
│ └───────────────────┘ │
│                       │
│ Minhas Chaves PIX     │
│ [+ Nova Chave]        │
│ ───────────────────   │
│                       │
│ ┌───────────────────┐ │
│ │ CPF               │ │
│ │ 123.456.789-00    │ │
│ │ ✅ ATIVA          │ │
│ │ [Ver] [Deletar]   │ │
│ └───────────────────┘ │
│                       │
│ ┌───────────────────┐ │
│ │ Email             │ │
│ │ joao@email.com    │ │
│ │ ✅ ATIVA          │ │
│ │ [Ver] [Deletar]   │ │
│ └───────────────────┘ │
│                       │
│ [<]  1 de 2  [>]      │
│                       │
└───────────────────────┘
```

---

## 4. Create Key Screen

### 4.1. Desktop View (1920x1080)

```
┌────────────────────────────────────────────────────────────────────────────┐
│ ┌────────┐                                      [Perfil ▼] [Notif 🔔]      │
│ │ LBPay  │  Dashboard > Chaves > Criar Nova Chave                          │
│ └────────┘                                                                 │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │  Criar Nova Chave PIX                                                 │ │
│  │  ────────────────────────────────────────────────────────────────────│ │
│  │                                                                       │ │
│  │  Tipo de Chave *                                                      │ │
│  │  ┌────────────────────────────────────────────────────────────────┐  │ │
│  │  │ CPF                                                           ▼│  │ │
│  │  └────────────────────────────────────────────────────────────────┘  │ │
│  │  └─ CPF, CNPJ, Email, Telefone, Chave Aleatória (EVP)               │ │
│  │                                                                       │ │
│  │  Valor da Chave *                                                     │ │
│  │  ┌────────────────────────────────────────────────────────────────┐  │ │
│  │  │ 123.456.789-00                                                  │  │ │
│  │  └────────────────────────────────────────────────────────────────┘  │ │
│  │  └─ Digite seu CPF (apenas números)                                  │ │
│  │                                                                       │ │
│  │  Conta *                                                              │ │
│  │  ┌────────────────────────────────────────────────────────────────┐  │ │
│  │  │ Corrente - 0001/12345678                                      ▼│  │ │
│  │  └────────────────────────────────────────────────────────────────┘  │ │
│  │  └─ Selecione a conta para vincular a chave PIX                      │ │
│  │                                                                       │ │
│  │  ┌──────────────────────────────────────────────────────────────┐    │ │
│  │  │ ℹ️  Após criar a chave, ela entrará em período de              │    │ │
│  │  │    reivindicação de 30 dias. Durante este período, outras      │    │ │
│  │  │    pessoas poderão reivindicar a posse da chave.               │    │ │
│  │  └──────────────────────────────────────────────────────────────┘    │ │
│  │                                                                       │ │
│  │                                             [Cancelar]  [Criar Chave] │ │
│  │                                                                       │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└────────────────────────────────────────────────────────────────────────────┘
```

### 4.2. Mobile View (375x667)

```
┌───────────────────────┐
│ ← Criar Chave         │
├───────────────────────┤
│                       │
│ Tipo de Chave *       │
│ ┌───────────────────┐ │
│ │ CPF             ▼ │ │
│ └───────────────────┘ │
│                       │
│ Valor da Chave *      │
│ ┌───────────────────┐ │
│ │ 123.456.789-00    │ │
│ └───────────────────┘ │
│ Digite seu CPF        │
│                       │
│ Conta *               │
│ ┌───────────────────┐ │
│ │ Corrente - 0001 ▼ │ │
│ └───────────────────┘ │
│                       │
│ ┌───────────────────┐ │
│ │ ℹ️  Após criar a   │ │
│ │ chave, ela entrará│ │
│ │ em período de     │ │
│ │ reivindicação de  │ │
│ │ 30 dias.          │ │
│ └───────────────────┘ │
│                       │
│ [Cancelar]            │
│ [Criar Chave]         │
│                       │
└───────────────────────┘
```

---

## 5. Key Details Screen

### 5.1. Desktop View (1920x1080)

```
┌────────────────────────────────────────────────────────────────────────────┐
│ ┌────────┐                                      [Perfil ▼] [Notif 🔔]      │
│ │ LBPay  │  Dashboard > Chaves > Detalhes                                  │
│ └────────┘                                                                 │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │  Detalhes da Chave PIX                                 ✅ ATIVA       │ │
│  │  ────────────────────────────────────────────────────────────────────│ │
│  │                                                                       │ │
│  │  Tipo de Chave                                                        │ │
│  │  CPF                                                                  │ │
│  │                                                                       │ │
│  │  Valor da Chave                                                       │ │
│  │  123.456.789-00                                                       │ │
│  │                                                                       │ │
│  │  Conta Vinculada                                                      │ │
│  │  Corrente - Agência 0001 / Conta 12345678                            │ │
│  │                                                                       │ │
│  │  Titular                                                              │ │
│  │  João Silva                                                           │ │
│  │  CPF: 123.456.789-00                                                  │ │
│  │                                                                       │ │
│  │  Instituição Financeira                                               │ │
│  │  LBPay (ISPB: 12345678)                                               │ │
│  │                                                                       │ │
│  │  Data de Criação                                                      │ │
│  │  25/10/2025 às 14:30                                                  │ │
│  │                                                                       │ │
│  │  Última Atualização                                                   │ │
│  │  25/10/2025 às 14:30                                                  │ │
│  │                                                                       │ │
│  │  ID Interno                                                           │ │
│  │  uuid-v4-entry-id                                                     │ │
│  │                                                                       │ │
│  │  ID Bacen                                                             │ │
│  │  bacen-dict-external-id                                               │ │
│  │                                                                       │ │
│  │                                             [Voltar]  [Deletar Chave] │ │
│  │                                                                       │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└────────────────────────────────────────────────────────────────────────────┘
```

### 5.2. Mobile View (375x667)

```
┌───────────────────────┐
│ ← Detalhes da Chave   │
├───────────────────────┤
│                       │
│ ✅ ATIVA              │
│                       │
│ Tipo de Chave         │
│ CPF                   │
│                       │
│ Valor da Chave        │
│ 123.456.789-00        │
│                       │
│ Conta Vinculada       │
│ Corrente              │
│ 0001/12345678         │
│                       │
│ Titular               │
│ João Silva            │
│ 123.456.789-00        │
│                       │
│ Instituição           │
│ LBPay (12345678)      │
│                       │
│ Criado em             │
│ 25/10/2025 14:30      │
│                       │
│ Atualizado em         │
│ 25/10/2025 14:30      │
│                       │
│ [Voltar]              │
│ [Deletar Chave]       │
│                       │
└───────────────────────┘
```

---

## 6. Claims List Screen

### 6.1. Desktop View (1920x1080)

```
┌────────────────────────────────────────────────────────────────────────────┐
│ ┌────────┐                                      [Perfil ▼] [Notif 🔔]      │
│ │ LBPay  │  Dashboard > Reivindicações                                      │
│ └────────┘                                                                 │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Reivindicações                                                             │
│  ────────────────────────────────────────────────────────────────────────  │
│                                                                             │
│  Filtros:  [Todas ▼]  [Status: Todos ▼]  [Meu Papel: Todos ▼]             │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │  Chave: 123.456.789-00                         ⚠️  AGUARDANDO RESPOSTA│ │
│  │  Você está sendo reivindicado                                         │ │
│  │  ────────────────────────────────────────────────────────────────────│ │
│  │  Criado: há 5 dias                                                    │ │
│  │  Expira em: 25 dias                                                   │ │
│  │  ⚠️  Atenção! Esta reivindicação expira em 25 dias.                  │ │
│  │                                             [Ver Detalhes] [Responder]│ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │  Chave: joao.silva@email.com                   ℹ️  ABERTO              │ │
│  │  Você está reivindicando                                              │ │
│  │  ────────────────────────────────────────────────────────────────────│ │
│  │  Criado: há 10 dias                                                   │ │
│  │  Expira em: 20 dias                                                   │ │
│  │                                             [Ver Detalhes] [Cancelar] │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  [Anterior]                              Página 1 de 1            [Próxima]│
│                                                                             │
└────────────────────────────────────────────────────────────────────────────┘
```

### 6.2. Mobile View (375x667)

```
┌───────────────────────┐
│ ☰  Reivindicações 🔔  │
├───────────────────────┤
│                       │
│ Filtros:              │
│ [Todas ▼] [Status ▼]  │
│                       │
│ ┌───────────────────┐ │
│ │ CPF               │ │
│ │ 123.456.789-00    │ │
│ │ ⚠️  AGUARDANDO    │ │
│ │ Você está sendo   │ │
│ │ reivindicado      │ │
│ │ ─────────────────│ │
│ │ Criado: há 5 dias │ │
│ │ Expira: 25 dias   │ │
│ │                   │ │
│ │ ⚠️  Expira em 25   │ │
│ │ dias              │ │
│ │                   │ │
│ │ [Ver] [Responder] │ │
│ └───────────────────┘ │
│                       │
│ ┌───────────────────┐ │
│ │ Email             │ │
│ │ joao@email.com    │ │
│ │ ℹ️  ABERTO         │ │
│ │ Você está         │ │
│ │ reivindicando     │ │
│ │ ─────────────────│ │
│ │ Criado: há 10 dias│ │
│ │ Expira: 20 dias   │ │
│ │                   │ │
│ │ [Ver] [Cancelar]  │ │
│ └───────────────────┘ │
│                       │
│ [<]  1 de 1  [>]      │
│                       │
└───────────────────────┘
```

---

## 7. Claim Details Screen

### 7.1. Desktop View (1920x1080)

```
┌────────────────────────────────────────────────────────────────────────────┐
│ ┌────────┐                                      [Perfil ▼] [Notif 🔔]      │
│ │ LBPay  │  Dashboard > Reivindicações > Detalhes                          │
│ └────────┘                                                                 │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │  Detalhes da Reivindicação                     ⚠️  AGUARDANDO RESPOSTA│ │
│  │  ────────────────────────────────────────────────────────────────────│ │
│  │                                                                       │ │
│  │  Chave PIX                                                            │ │
│  │  CPF: 123.456.789-00                                                  │ │
│  │                                                                       │ │
│  │  Status                                                               │ │
│  │  Aguardando Resposta do Proprietário                                  │ │
│  │                                                                       │ │
│  │  Período de Resolução                                                 │ │
│  │  30 dias                                                              │ │
│  │                                                                       │ │
│  │  Data de Criação                                                      │ │
│  │  20/10/2025 às 10:00                                                  │ │
│  │                                                                       │ │
│  │  Data de Expiração                                                    │ │
│  │  19/11/2025 às 10:00                                                  │ │
│  │                                                                       │ │
│  │  Dias Restantes                                                       │ │
│  │  25 dias                                                              │ │
│  │                                                                       │ │
│  │  ┌──────────────────────────────────────────────────────────────┐    │ │
│  │  │ ⏱️  Linha do Tempo                                             │    │ │
│  │  │ ────────────────────────────────────────────────────────────│    │ │
│  │  │                                                              │    │ │
│  │  │ 20/10/2025  ●  Reivindicação criada                          │    │ │
│  │  │                                                              │    │ │
│  │  │ 20/10/2025  ●  Notificação enviada ao proprietário           │    │ │
│  │  │                                                              │    │ │
│  │  │ ...         ○  Aguardando resposta                           │    │ │
│  │  │                                                              │    │ │
│  │  │ 19/11/2025  ○  Prazo final (30 dias)                         │    │ │
│  │  │                                                              │    │ │
│  │  └──────────────────────────────────────────────────────────────┘    │ │
│  │                                                                       │ │
│  │  Proprietário Atual                                                   │ │
│  │  LBPay (ISPB: 12345678)                                               │ │
│  │                                                                       │ │
│  │  Reivindicante                                                        │ │
│  │  Outro Banco (ISPB: 87654321)                                         │ │
│  │                                                                       │ │
│  │  ID Interno                                                           │ │
│  │  uuid-v4-claim-id                                                     │ │
│  │                                                                       │ │
│  │  ID Bacen                                                             │ │
│  │  bacen-claim-external-id                                              │ │
│  │                                                                       │ │
│  │                                                   [Voltar] [Responder]│ │
│  │                                                                       │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└────────────────────────────────────────────────────────────────────────────┘
```

### 7.2. Mobile View (375x667)

```
┌───────────────────────┐
│ ← Detalhes            │
├───────────────────────┤
│                       │
│ ⚠️  AGUARDANDO        │
│                       │
│ Chave PIX             │
│ CPF: 123.456.789-00   │
│                       │
│ Status                │
│ Aguardando Resposta   │
│                       │
│ Período               │
│ 30 dias               │
│                       │
│ Criado em             │
│ 20/10/2025 10:00      │
│                       │
│ Expira em             │
│ 19/11/2025 10:00      │
│                       │
│ Dias Restantes        │
│ 25 dias               │
│                       │
│ ┌───────────────────┐ │
│ │ ⏱️  Linha do Tempo │ │
│ │ ─────────────────│ │
│ │ 20/10 ● Criada    │ │
│ │ 20/10 ● Notificado│ │
│ │ ...   ○ Aguardando│ │
│ │ 19/11 ○ Expira    │ │
│ └───────────────────┘ │
│                       │
│ Proprietário Atual    │
│ LBPay (12345678)      │
│                       │
│ Reivindicante         │
│ Outro (87654321)      │
│                       │
│ [Voltar]              │
│ [Responder]           │
│                       │
└───────────────────────┘
```

---

## 8. Respond to Claim Screen

### 8.1. Desktop View (1920x1080)

```
┌────────────────────────────────────────────────────────────────────────────┐
│ ┌────────┐                                      [Perfil ▼] [Notif 🔔]      │
│ │ LBPay  │  Dashboard > Reivindicações > Responder                         │
│ └────────┘                                                                 │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │  Responder Reivindicação                                              │ │
│  │  ────────────────────────────────────────────────────────────────────│ │
│  │                                                                       │ │
│  │  Chave PIX: CPF 123.456.789-00                                        │ │
│  │                                                                       │ │
│  │  ┌──────────────────────────────────────────────────────────────┐    │ │
│  │  │ ℹ️  Você tem 25 dias para responder esta reivindicação.       │    │ │
│  │  │    Após esse período, ela será cancelada automaticamente.     │    │ │
│  │  └──────────────────────────────────────────────────────────────┘    │ │
│  │                                                                       │ │
│  │  Como deseja responder?                                               │ │
│  │                                                                       │ │
│  │  ┌────────────────────────────────────────────────────────────────┐  │ │
│  │  │  ✅  Confirmar Transferência                                    │  │ │
│  │  │                                                                 │  │ │
│  │  │  Ao confirmar, você autoriza a transferência desta chave PIX    │  │ │
│  │  │  para a conta do reivindicante. Esta ação não pode ser          │  │ │
│  │  │  desfeita.                                                       │  │ │
│  │  │                                                                 │  │ │
│  │  │                                         [Confirmar Transferência]│  │ │
│  │  └────────────────────────────────────────────────────────────────┘  │ │
│  │                                                                       │ │
│  │  ┌────────────────────────────────────────────────────────────────┐  │ │
│  │  │  ❌  Recusar Reivindicação                                      │  │ │
│  │  │                                                                 │  │ │
│  │  │  Ao recusar, a reivindicação será cancelada e você manterá      │  │ │
│  │  │  a posse da chave PIX.                                          │  │ │
│  │  │                                                                 │  │ │
│  │  │                                         [Recusar Reivindicação]  │  │ │
│  │  └────────────────────────────────────────────────────────────────┘  │ │
│  │                                                                       │ │
│  │                                                            [Voltar]   │ │
│  │                                                                       │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└────────────────────────────────────────────────────────────────────────────┘

--- AFTER SELECTING "CONFIRMAR" ---

┌────────────────────────────────────────────────────────────────────────────┐
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │  Confirmar Transferência                                              │ │
│  │  ────────────────────────────────────────────────────────────────────│ │
│  │                                                                       │ │
│  │  Chave PIX: CPF 123.456.789-00                                        │ │
│  │                                                                       │ │
│  │  ┌──────────────────────────────────────────────────────────────┐    │ │
│  │  │ ⚠️  ATENÇÃO! Ao confirmar, você está autorizando a            │    │ │
│  │  │    transferência desta chave PIX para outra conta.            │    │ │
│  │  │    Esta ação NÃO PODE SER DESFEITA.                           │    │ │
│  │  └──────────────────────────────────────────────────────────────┘    │ │
│  │                                                                       │ │
│  │  Motivo *                                                             │ │
│  │  ┌────────────────────────────────────────────────────────────────┐  │ │
│  │  │                                                                 │  │ │
│  │  │                                                                 │  │ │
│  │  │                                                                 │  │ │
│  │  │                                                                 │  │ │
│  │  └────────────────────────────────────────────────────────────────┘  │ │
│  │  └─ Por que você está confirmando esta reivindicação?                │ │
│  │     (mínimo 10 caracteres)                                            │ │
│  │                                                                       │ │
│  │                                             [Voltar]  [Confirmar]     │ │
│  │                                                                       │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────────────────┘
```

### 8.2. Mobile View (375x667)

```
┌───────────────────────┐
│ ← Responder           │
├───────────────────────┤
│                       │
│ Chave PIX             │
│ CPF: 123.456.789-00   │
│                       │
│ ┌───────────────────┐ │
│ │ ℹ️  Você tem 25    │ │
│ │ dias para responder│ │
│ └───────────────────┘ │
│                       │
│ Como deseja responder?│
│                       │
│ ┌───────────────────┐ │
│ │ ✅  Confirmar     │ │
│ │ Transferência     │ │
│ │                   │ │
│ │ Ao confirmar, você│ │
│ │ autoriza a        │ │
│ │ transferência.    │ │
│ │                   │ │
│ │ [Confirmar]       │ │
│ └───────────────────┘ │
│                       │
│ ┌───────────────────┐ │
│ │ ❌  Recusar       │ │
│ │ Reivindicação     │ │
│ │                   │ │
│ │ Ao recusar, você  │ │
│ │ manterá a posse   │ │
│ │ da chave.         │ │
│ │                   │ │
│ │ [Recusar]         │ │
│ └───────────────────┘ │
│                       │
│ [Voltar]              │
│                       │
└───────────────────────┘

--- AFTER SELECTING ---

┌───────────────────────┐
│ ← Confirmar           │
├───────────────────────┤
│                       │
│ Chave PIX             │
│ CPF: 123.456.789-00   │
│                       │
│ ┌───────────────────┐ │
│ │ ⚠️  ATENÇÃO!       │ │
│ │ Esta ação NÃO PODE│ │
│ │ ser desfeita.     │ │
│ └───────────────────┘ │
│                       │
│ Motivo *              │
│ ┌───────────────────┐ │
│ │                   │ │
│ │                   │ │
│ │                   │ │
│ │                   │ │
│ └───────────────────┘ │
│ Mín: 10 caracteres    │
│                       │
│ [Voltar]              │
│ [Confirmar]           │
│                       │
└───────────────────────┘
```

---

## Rastreabilidade

### Requisitos de UX/UI

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RF-WF-001 | Dashboard wireframe | FE-001 | ✅ Especificado |
| RF-WF-002 | Create Key wireframe | FE-001 | ✅ Especificado |
| RF-WF-003 | Key Details wireframe | FE-001 | ✅ Especificado |
| RF-WF-004 | Claims List wireframe | FE-001, DIA-006 | ✅ Especificado |
| RF-WF-005 | Claim Details wireframe | FE-001 | ✅ Especificado |
| RF-WF-006 | Respond to Claim wireframe | FE-001, DIA-006 | ✅ Especificado |

---

## Próximas Revisões

**Pendências**:
- [ ] Criar protótipos interativos no Figma
- [ ] Validar acessibilidade WCAG 2.1 AA
- [ ] Realizar testes de usabilidade com usuários reais
- [ ] Adicionar dark mode variants
- [ ] Criar wireframes para flows de erro (offline, timeout)
- [ ] Adicionar animações e micro-interactions

---

**Referências**:
- [FE-001: Component Specifications](../Componentes/FE-001_Component_Specifications.md)
- [API-002: Core DICT REST API](../../04_APIs/REST/API-002_Core_DICT_REST_API.md)
- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
- [Material Design](https://material.io/design)
