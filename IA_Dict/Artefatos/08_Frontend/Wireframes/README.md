# Wireframes

**Propósito**: Wireframes de baixa/média fidelidade para interfaces do usuário DICT

## 📋 Conteúdo

Esta pasta armazenará:

- **Low-Fidelity Wireframes**: Esboços iniciais (papel, Balsamiq)
- **High-Fidelity Wireframes**: Wireframes detalhados (Figma, Adobe XD)
- **Responsive Designs**: Versões mobile, tablet, desktop
- **User Flows**: Fluxos de navegação entre telas

## 📁 Estrutura Esperada

```
Wireframes/
├── Low_Fidelity/
│   ├── WF_001_CreateEntry.png
│   ├── WF_002_ClaimStatus.png
│   └── WF_003_Portability.png
├── High_Fidelity/
│   ├── HF_001_CreateEntry_Desktop.fig
│   ├── HF_001_CreateEntry_Mobile.fig
│   ├── HF_002_ClaimStatus_Desktop.fig
│   └── HF_002_ClaimStatus_Mobile.fig
└── User_Flows/
    ├── Flow_CreateEntry.png
    └── Flow_ClaimWorkflow.png
```

## 🎯 Telas Principais

### 1. Create Entry (Criar Chave DICT)
- **Campos**:
  - Tipo de chave (CPF, CNPJ, Phone, Email, EVP)
  - Valor da chave
  - Dados da conta (ISPB, agência, conta)
- **Ações**:
  - Validar em tempo real
  - Submit (criar chave)
  - Cancelar

### 2. Claim Status (Status de Reivindicação)
- **Informações**:
  - ID da claim
  - Status (OPEN, WAITING_RESOLUTION, CONFIRMED, etc.)
  - Data de expiração (30 dias)
  - Progresso visual (barra ou timeline)
- **Ações**:
  - Aceitar claim (se dono)
  - Rejeitar claim
  - Ver detalhes

### 3. Portability (Portabilidade de Conta)
- **Formulário**:
  - Chave DICT
  - Nova conta de destino
  - Confirmação de portabilidade
- **Ações**:
  - Confirmar portabilidade
  - Cancelar

### 4. Entry List (Lista de Chaves)
- **Tabela**:
  - Tipo de chave
  - Valor da chave
  - Status
  - Ações (editar, deletar, ver detalhes)
- **Filtros**:
  - Por tipo de chave
  - Por status
  - Busca por valor

## 🔗 Ferramentas

- **Balsamiq**: Low-fidelity wireframes rápidos
- **Figma**: High-fidelity wireframes e protótipos
- **Adobe XD**: Alternativa ao Figma
- **Miro**: Colaboração em wireframes

## 📐 Breakpoints Responsivos

```css
/* Mobile */
@media (max-width: 640px) { ... }

/* Tablet */
@media (min-width: 641px) and (max-width: 1024px) { ... }

/* Desktop */
@media (min-width: 1025px) { ... }
```

## 📚 Referências

- [Jornadas de Usuário](../Jornadas/)
- [Componentes](../Componentes/)
- [Requisitos](../../01_Requisitos/)
- [Design System LBPay](link-interno)

---

**Status**: 🔴 Pasta vazia (será preenchida na Fase 3)
**Fase de Preenchimento**: Fase 3 (design UX/UI)
**Responsável**: UX Designer + Product Owner
