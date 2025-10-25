# Componentes Frontend

**Propósito**: Especificação de componentes React para interface do usuário DICT

## 📋 Conteúdo

Esta pasta armazenará:

- **Component Specs**: Especificações detalhadas de componentes React
- **Component Library**: Biblioteca de componentes reutilizáveis (Design System)
- **API Integration**: Como componentes interagem com Core DICT API
- **State Management**: Estratégia de gerenciamento de estado (Redux, Zustand, etc.)

## 📁 Estrutura Esperada

```
Componentes/
├── Atomicos/
│   ├── Button.md
│   ├── Input.md
│   ├── Select.md
│   └── Card.md
├── Compostos/
│   ├── EntryForm.md
│   ├── ClaimList.md
│   ├── AccountSelector.md
│   └── KeyTypeSelector.md
├── Paginas/
│   ├── CreateEntryPage.md
│   ├── ClaimStatusPage.md
│   └── PortabilityPage.md
└── State/
    ├── Redux_Store_Design.md
    └── API_Client_Hooks.md
```

## 🎯 Exemplo: EntryForm Component

```markdown
# EntryForm Component

**Propósito**: Formulário para criação de nova chave DICT

## Props

```typescript
interface EntryFormProps {
  onSubmit: (entry: CreateEntryRequest) => Promise<void>;
  onCancel: () => void;
  initialValues?: Partial<CreateEntryRequest>;
  disabled?: boolean;
}
```

## State

```typescript
interface EntryFormState {
  keyType: KeyType;  // CPF, CNPJ, PHONE, EMAIL, EVP
  keyValue: string;
  account: {
    ispb: string;
    accountNumber: string;
    accountCheckDigit: string;
    branchCode: string;
  };
  errors: Record<string, string>;
  isSubmitting: boolean;
}
```

## Validation Rules

- **keyType**: Required, must be one of [CPF, CNPJ, PHONE, EMAIL, EVP]
- **keyValue**:
  - CPF: 11 digits, valid CPF algorithm
  - CNPJ: 14 digits, valid CNPJ algorithm
  - PHONE: +55 format, 11 digits
  - EMAIL: Valid email format
  - EVP: UUID v4
- **account.ispb**: 8 digits
- **account.accountNumber**: Max 20 chars

## API Integration

```typescript
const handleSubmit = async (values: EntryFormState) => {
  try {
    const response = await apiClient.post('/api/v1/entries', {
      key: {
        keyType: values.keyType,
        keyValue: values.keyValue,
      },
      account: values.account,
    });

    toast.success('Chave DICT criada com sucesso!');
    navigate('/entries');
  } catch (error) {
    if (error.status === 409) {
      setErrors({ keyValue: 'Esta chave já existe' });
    } else {
      toast.error('Erro ao criar chave DICT');
    }
  }
};
```

## Dependencies

- `react-hook-form` (form management)
- `zod` (validation)
- `@tanstack/react-query` (API calls)
```

## 🔗 Design System

Componentes devem seguir Design System da LBPay:

- **Cores**: Paleta corporativa
- **Tipografia**: Fonte padrão (Inter, Roboto, etc.)
- **Espaçamento**: Grid 8px
- **Acessibilidade**: WCAG 2.1 AA

## 📚 Referências

- [Wireframes](../Wireframes/)
- [Jornadas de Usuário](../Jornadas/)
- [Core DICT API](../../04_APIs/REST/)
- [Design System LBPay](link-interno)

---

**Status**: 🔴 Pasta vazia (será preenchida na Fase 3+)
**Fase de Preenchimento**: Fase 3 (após Backend pronto)
**Stack**: React, TypeScript, TailwindCSS, React Query
