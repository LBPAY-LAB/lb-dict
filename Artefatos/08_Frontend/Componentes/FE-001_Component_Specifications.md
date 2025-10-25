# FE-001: Component Specifications - DICT Frontend

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Documento**: FE-001 - Component Specifications
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: FRONTEND (AI Agent - Frontend Specialist)
**Revisor**: [Aguardando]
**Aprovador**: Tech Lead, Head de Arquitetura

---

## Sumário Executivo

Este documento especifica os **componentes React** do frontend DICT LBPay, incluindo interfaces TypeScript, props, state, validação com Zod, e integração com state management (Zustand). Os componentes seguem princípios de Clean Code, SOLID, e Atomic Design.

**Baseado em**:
- [API-002: Core DICT REST API](../../04_APIs/REST/API-002_Core_DICT_REST_API.md)
- [DIA-001: C4 Context Diagram](../../02_Arquitetura/Diagramas/DIA-001_C4_Context_Diagram.md)
- [DIA-006: Sequence Claim Workflow](../../02_Arquitetura/Diagramas/DIA-006_Sequence_Claim_Workflow.md)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | FRONTEND | Versão inicial - Component Specifications |

---

## Índice

1. [Visão Geral](#1-visão-geral)
2. [Component Architecture](#2-component-architecture)
3. [Key Management Components](#3-key-management-components)
4. [Claim Management Components](#4-claim-management-components)
5. [Portability Components](#5-portability-components)
6. [Shared Components](#6-shared-components)
7. [Type Definitions](#7-type-definitions)
8. [Validation Schemas](#8-validation-schemas)

---

## 1. Visão Geral

### 1.1. Technology Stack

| Tecnologia | Versão | Propósito |
|------------|--------|-----------|
| **React** | 18.3+ | UI Framework |
| **TypeScript** | 5.3+ | Type Safety |
| **Zustand** | 4.5+ | State Management |
| **React Hook Form** | 7.50+ | Form Management |
| **Zod** | 3.22+ | Schema Validation |
| **React Query** | 5.0+ | API Integration |
| **TailwindCSS** | 3.4+ | Styling |
| **Radix UI** | 1.0+ | Headless Components |

### 1.2. Component Categories

```
components/
├── features/
│   ├── keys/
│   │   ├── KeyList.tsx
│   │   ├── CreateKeyForm.tsx
│   │   ├── KeyDetailsCard.tsx
│   │   └── DeleteKeyDialog.tsx
│   ├── claims/
│   │   ├── ClaimCard.tsx
│   │   ├── ClaimsList.tsx
│   │   ├── ClaimResponseForm.tsx
│   │   └── ClaimDetailsModal.tsx
│   └── portability/
│       ├── PortabilityCard.tsx
│       ├── InitiatePortabilityForm.tsx
│       └── PortabilityStatusBadge.tsx
├── shared/
│   ├── ui/
│   │   ├── Button.tsx
│   │   ├── Input.tsx
│   │   ├── Select.tsx
│   │   ├── Card.tsx
│   │   ├── Badge.tsx
│   │   ├── Dialog.tsx
│   │   └── Toast.tsx
│   ├── forms/
│   │   ├── FormField.tsx
│   │   ├── FormError.tsx
│   │   └── FormSelect.tsx
│   └── layout/
│       ├── DashboardLayout.tsx
│       ├── Header.tsx
│       └── Sidebar.tsx
└── hooks/
    ├── useKeyManagement.ts
    ├── useClaimManagement.ts
    ├── usePortability.ts
    └── useAuth.ts
```

---

## 2. Component Architecture

### 2.1. Design Principles

1. **Atomic Design**: Atoms → Molecules → Organisms → Templates → Pages
2. **Component Composition**: Prefer composition over inheritance
3. **Single Responsibility**: Each component has one clear purpose
4. **Type Safety**: Full TypeScript coverage, no `any` types
5. **Accessibility**: WCAG 2.1 AA compliance
6. **Performance**: React.memo, useMemo, useCallback for optimization

### 2.2. Component Structure Template

```typescript
// components/features/example/ExampleComponent.tsx

import React from 'react';
import { z } from 'zod';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';

// 1. Type Definitions
interface ExampleComponentProps {
  // Props here
}

// 2. Validation Schema
const exampleSchema = z.object({
  // Schema here
});

type ExampleFormData = z.infer<typeof exampleSchema>;

// 3. Component
export const ExampleComponent: React.FC<ExampleComponentProps> = ({
  // Destructure props
}) => {
  // 4. Hooks
  const form = useForm<ExampleFormData>({
    resolver: zodResolver(exampleSchema),
  });

  // 5. Event Handlers
  const handleSubmit = (data: ExampleFormData) => {
    // Handle submission
  };

  // 6. Render
  return (
    <div>
      {/* JSX here */}
    </div>
  );
};
```

---

## 3. Key Management Components

### 3.1. KeyList Component

**Purpose**: Display list of user's PIX keys with filtering and actions.

**File**: `components/features/keys/KeyList.tsx`

```typescript
import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { KeyCard } from './KeyCard';
import { Button } from '@/components/shared/ui/Button';
import { Select } from '@/components/shared/ui/Select';
import { Skeleton } from '@/components/shared/ui/Skeleton';
import { useKeyManagement } from '@/hooks/useKeyManagement';
import { KeyType, KeyStatus, Entry } from '@/types/dict';

// Props Interface
interface KeyListProps {
  onCreateKey: () => void;
  onViewDetails: (entryId: string) => void;
  onDeleteKey: (entryId: string) => void;
}

// Component
export const KeyList: React.FC<KeyListProps> = ({
  onCreateKey,
  onViewDetails,
  onDeleteKey,
}) => {
  // State
  const [statusFilter, setStatusFilter] = useState<KeyStatus | 'ALL'>('ALL');
  const [page, setPage] = useState(1);
  const [limit] = useState(20);

  // API Hook
  const { useListKeys } = useKeyManagement();
  const { data, isLoading, error } = useListKeys({
    page,
    limit,
    status: statusFilter !== 'ALL' ? statusFilter : undefined,
  });

  // Handlers
  const handleFilterChange = (value: string) => {
    setStatusFilter(value as KeyStatus | 'ALL');
    setPage(1); // Reset to first page
  };

  const handlePageChange = (newPage: number) => {
    setPage(newPage);
  };

  // Loading State
  if (isLoading) {
    return (
      <div className="space-y-4">
        {[...Array(3)].map((_, i) => (
          <Skeleton key={i} className="h-24 w-full" />
        ))}
      </div>
    );
  }

  // Error State
  if (error) {
    return (
      <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
        <p className="text-red-800">
          Erro ao carregar chaves PIX. Tente novamente.
        </p>
      </div>
    );
  }

  // Empty State
  if (!data?.entries || data.entries.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 mb-4">Nenhuma chave PIX cadastrada.</p>
        <Button onClick={onCreateKey}>Criar Primeira Chave</Button>
      </div>
    );
  }

  // Render
  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">Minhas Chaves PIX</h2>
        <Button onClick={onCreateKey}>+ Nova Chave</Button>
      </div>

      {/* Filters */}
      <div className="flex gap-4">
        <Select
          value={statusFilter}
          onValueChange={handleFilterChange}
          options={[
            { value: 'ALL', label: 'Todas' },
            { value: 'ACTIVE', label: 'Ativas' },
            { value: 'PENDING', label: 'Pendentes' },
            { value: 'DELETED', label: 'Deletadas' },
          ]}
        />
      </div>

      {/* Key List */}
      <div className="grid gap-4">
        {data.entries.map((entry) => (
          <KeyCard
            key={entry.entry_id}
            entry={entry}
            onViewDetails={() => onViewDetails(entry.entry_id)}
            onDelete={() => onDeleteKey(entry.entry_id)}
          />
        ))}
      </div>

      {/* Pagination */}
      {data.pagination && data.pagination.total_pages > 1 && (
        <div className="flex justify-center gap-2">
          <Button
            variant="outline"
            disabled={page === 1}
            onClick={() => handlePageChange(page - 1)}
          >
            Anterior
          </Button>
          <span className="px-4 py-2">
            Página {page} de {data.pagination.total_pages}
          </span>
          <Button
            variant="outline"
            disabled={page === data.pagination.total_pages}
            onClick={() => handlePageChange(page + 1)}
          >
            Próxima
          </Button>
        </div>
      )}
    </div>
  );
};
```

---

### 3.2. CreateKeyForm Component

**Purpose**: Form to create new PIX key.

**File**: `components/features/keys/CreateKeyForm.tsx`

```typescript
import React from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/shared/ui/Button';
import { FormField } from '@/components/shared/forms/FormField';
import { Select } from '@/components/shared/ui/Select';
import { Input } from '@/components/shared/ui/Input';
import { useKeyManagement } from '@/hooks/useKeyManagement';
import { useToast } from '@/hooks/useToast';
import { Account, KeyType } from '@/types/dict';

// Validation Schema
const createKeySchema = z.object({
  key_type: z.enum(['CPF', 'CNPJ', 'EMAIL', 'PHONE', 'EVP'], {
    required_error: 'Selecione o tipo de chave',
  }),
  key_value: z
    .string()
    .min(1, 'Campo obrigatório')
    .refine(
      (val, ctx) => {
        const keyType = ctx.path[0] as KeyType;

        // CPF validation (11 digits)
        if (keyType === 'CPF') {
          return /^\d{11}$/.test(val.replace(/\D/g, ''));
        }

        // CNPJ validation (14 digits)
        if (keyType === 'CNPJ') {
          return /^\d{14}$/.test(val.replace(/\D/g, ''));
        }

        // EMAIL validation
        if (keyType === 'EMAIL') {
          return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(val);
        }

        // PHONE validation (+5511999999999)
        if (keyType === 'PHONE') {
          return /^\+55\d{11}$/.test(val.replace(/\D/g, '+55'));
        }

        // EVP (generated automatically)
        if (keyType === 'EVP') {
          return true; // No manual input required
        }

        return true;
      },
      {
        message: 'Formato inválido para o tipo de chave selecionado',
      }
    ),
  account_id: z.string().uuid('Selecione uma conta'),
});

type CreateKeyFormData = z.infer<typeof createKeySchema>;

// Props Interface
interface CreateKeyFormProps {
  accounts: Account[];
  onSuccess: () => void;
  onCancel: () => void;
}

// Component
export const CreateKeyForm: React.FC<CreateKeyFormProps> = ({
  accounts,
  onSuccess,
  onCancel,
}) => {
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { createKey } = useKeyManagement();

  // Form Setup
  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<CreateKeyFormData>({
    resolver: zodResolver(createKeySchema),
    defaultValues: {
      key_type: 'CPF',
    },
  });

  const selectedKeyType = watch('key_type');

  // Mutation
  const mutation = useMutation({
    mutationFn: createKey,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['keys'] });
      toast({
        title: 'Chave PIX criada com sucesso!',
        variant: 'success',
      });
      onSuccess();
    },
    onError: (error: any) => {
      toast({
        title: 'Erro ao criar chave PIX',
        description: error.message || 'Tente novamente mais tarde',
        variant: 'error',
      });
    },
  });

  // Handlers
  const onSubmit = async (data: CreateKeyFormData) => {
    const selectedAccount = accounts.find((acc) => acc.id === data.account_id);

    if (!selectedAccount) {
      toast({
        title: 'Conta inválida',
        variant: 'error',
      });
      return;
    }

    await mutation.mutateAsync({
      key_type: data.key_type,
      key_value: data.key_type === 'EVP' ? '' : data.key_value, // EVP generated by backend
      account: {
        account_number: selectedAccount.account_number,
        branch_code: selectedAccount.branch_code,
        account_type: selectedAccount.account_type,
        holder_document: selectedAccount.holder_document,
        holder_name: selectedAccount.holder_name,
        participant_ispb: selectedAccount.participant_ispb,
      },
      idempotency_key: crypto.randomUUID(),
    });
  };

  // Helper Functions
  const getKeyValuePlaceholder = () => {
    switch (selectedKeyType) {
      case 'CPF':
        return '123.456.789-00';
      case 'CNPJ':
        return '12.345.678/0001-00';
      case 'EMAIL':
        return 'seu@email.com';
      case 'PHONE':
        return '+55 11 99999-9999';
      case 'EVP':
        return 'Gerado automaticamente';
      default:
        return '';
    }
  };

  const getKeyValueHelp = () => {
    switch (selectedKeyType) {
      case 'CPF':
        return 'Digite seu CPF (apenas números)';
      case 'CNPJ':
        return 'Digite seu CNPJ (apenas números)';
      case 'EMAIL':
        return 'Digite um email válido';
      case 'PHONE':
        return 'Digite seu telefone com DDD (+55)';
      case 'EVP':
        return 'Uma chave aleatória será gerada automaticamente';
      default:
        return '';
    }
  };

  // Render
  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Key Type */}
      <FormField
        label="Tipo de Chave"
        error={errors.key_type?.message}
        required
      >
        <Select
          {...register('key_type')}
          options={[
            { value: 'CPF', label: 'CPF' },
            { value: 'CNPJ', label: 'CNPJ' },
            { value: 'EMAIL', label: 'Email' },
            { value: 'PHONE', label: 'Telefone' },
            { value: 'EVP', label: 'Chave Aleatória (EVP)' },
          ]}
        />
      </FormField>

      {/* Key Value (hidden for EVP) */}
      {selectedKeyType !== 'EVP' && (
        <FormField
          label="Valor da Chave"
          error={errors.key_value?.message}
          help={getKeyValueHelp()}
          required
        >
          <Input
            {...register('key_value')}
            type={selectedKeyType === 'EMAIL' ? 'email' : 'text'}
            placeholder={getKeyValuePlaceholder()}
          />
        </FormField>
      )}

      {/* Account Selection */}
      <FormField
        label="Conta"
        error={errors.account_id?.message}
        help="Selecione a conta para vincular a chave PIX"
        required
      >
        <Select
          {...register('account_id')}
          options={accounts.map((acc) => ({
            value: acc.id,
            label: `${acc.account_type} - ${acc.branch_code}/${acc.account_number}`,
          }))}
        />
      </FormField>

      {/* Actions */}
      <div className="flex gap-4 justify-end">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={isSubmitting}
        >
          Cancelar
        </Button>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Criando...' : 'Criar Chave'}
        </Button>
      </div>
    </form>
  );
};
```

---

## 4. Claim Management Components

### 4.1. ClaimCard Component

**Purpose**: Display claim information with status and actions.

**File**: `components/features/claims/ClaimCard.tsx`

```typescript
import React from 'react';
import { Card } from '@/components/shared/ui/Card';
import { Badge } from '@/components/shared/ui/Badge';
import { Button } from '@/components/shared/ui/Button';
import { Claim, ClaimStatus } from '@/types/dict';
import { formatDistance } from 'date-fns';
import { ptBR } from 'date-fns/locale';

// Props Interface
interface ClaimCardProps {
  claim: Claim;
  role: 'CLAIMER' | 'OWNER';
  onViewDetails: () => void;
  onRespond?: () => void; // Only for OWNER
  onCancel?: () => void;
}

// Helper Functions
const getStatusBadgeVariant = (status: ClaimStatus) => {
  switch (status) {
    case 'OPEN':
      return 'info';
    case 'WAITING_RESOLUTION':
      return 'warning';
    case 'CONFIRMED':
      return 'success';
    case 'CANCELLED':
      return 'neutral';
    case 'COMPLETED':
      return 'success';
    case 'EXPIRED':
      return 'error';
    default:
      return 'neutral';
  }
};

const getStatusLabel = (status: ClaimStatus) => {
  const labels: Record<ClaimStatus, string> = {
    OPEN: 'Aberto',
    WAITING_RESOLUTION: 'Aguardando Resposta',
    CONFIRMED: 'Confirmado',
    CANCELLED: 'Cancelado',
    COMPLETED: 'Concluído',
    EXPIRED: 'Expirado',
  };
  return labels[status];
};

// Component
export const ClaimCard: React.FC<ClaimCardProps> = ({
  claim,
  role,
  onViewDetails,
  onRespond,
  onCancel,
}) => {
  const daysRemaining = claim.days_remaining ?? 0;
  const isExpiringSoon = daysRemaining > 0 && daysRemaining <= 5;
  const canRespond =
    role === 'OWNER' &&
    (claim.status === 'OPEN' || claim.status === 'WAITING_RESOLUTION');

  return (
    <Card className="p-6 hover:shadow-lg transition-shadow">
      {/* Header */}
      <div className="flex justify-between items-start mb-4">
        <div>
          <h3 className="text-lg font-semibold mb-1">
            Chave: {claim.entry_key}
          </h3>
          <p className="text-sm text-gray-500">
            {role === 'CLAIMER' ? 'Você está reivindicando' : 'Reivindicação recebida'}
          </p>
        </div>
        <Badge variant={getStatusBadgeVariant(claim.status)}>
          {getStatusLabel(claim.status)}
        </Badge>
      </div>

      {/* Details */}
      <div className="space-y-2 mb-4">
        <div className="flex justify-between text-sm">
          <span className="text-gray-600">Criado em:</span>
          <span className="font-medium">
            {formatDistance(new Date(claim.created_at), new Date(), {
              addSuffix: true,
              locale: ptBR,
            })}
          </span>
        </div>

        {claim.expires_at && daysRemaining > 0 && (
          <div className="flex justify-between text-sm">
            <span className="text-gray-600">Expira em:</span>
            <span className={`font-medium ${isExpiringSoon ? 'text-red-600' : ''}`}>
              {daysRemaining} {daysRemaining === 1 ? 'dia' : 'dias'}
            </span>
          </div>
        )}
      </div>

      {/* Warning for expiring claims */}
      {isExpiringSoon && role === 'OWNER' && (
        <div className="mb-4 p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
          <p className="text-sm text-yellow-800">
            Atenção! Esta reivindicação expira em {daysRemaining} {daysRemaining === 1 ? 'dia' : 'dias'}.
          </p>
        </div>
      )}

      {/* Actions */}
      <div className="flex gap-3">
        <Button variant="outline" size="sm" onClick={onViewDetails}>
          Ver Detalhes
        </Button>

        {canRespond && onRespond && (
          <Button size="sm" onClick={onRespond}>
            Responder
          </Button>
        )}

        {(claim.status === 'OPEN' || claim.status === 'WAITING_RESOLUTION') && onCancel && (
          <Button variant="destructive" size="sm" onClick={onCancel}>
            Cancelar
          </Button>
        )}
      </div>
    </Card>
  );
};
```

---

### 4.2. ClaimResponseForm Component

**Purpose**: Form for owner to confirm or cancel claim.

**File**: `components/features/claims/ClaimResponseForm.tsx`

```typescript
import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/shared/ui/Button';
import { FormField } from '@/components/shared/forms/FormField';
import { Textarea } from '@/components/shared/ui/Textarea';
import { useClaimManagement } from '@/hooks/useClaimManagement';
import { useToast } from '@/hooks/useToast';
import { Claim } from '@/types/dict';

// Validation Schema
const claimResponseSchema = z.object({
  action: z.enum(['CONFIRM', 'CANCEL']),
  reason: z
    .string()
    .min(10, 'Digite pelo menos 10 caracteres')
    .max(500, 'Máximo de 500 caracteres'),
});

type ClaimResponseFormData = z.infer<typeof claimResponseSchema>;

// Props Interface
interface ClaimResponseFormProps {
  claim: Claim;
  onSuccess: () => void;
  onCancel: () => void;
}

// Component
export const ClaimResponseForm: React.FC<ClaimResponseFormProps> = ({
  claim,
  onSuccess,
  onCancel,
}) => {
  const [selectedAction, setSelectedAction] = useState<'CONFIRM' | 'CANCEL' | null>(null);
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { confirmClaim, cancelClaim } = useClaimManagement();

  // Form Setup
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ClaimResponseFormData>({
    resolver: zodResolver(claimResponseSchema),
  });

  // Mutations
  const confirmMutation = useMutation({
    mutationFn: confirmClaim,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['claims'] });
      toast({
        title: 'Reivindicação confirmada!',
        description: 'A transferência será concluída em até 24 horas.',
        variant: 'success',
      });
      onSuccess();
    },
    onError: (error: any) => {
      toast({
        title: 'Erro ao confirmar reivindicação',
        description: error.message,
        variant: 'error',
      });
    },
  });

  const cancelMutation = useMutation({
    mutationFn: cancelClaim,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['claims'] });
      toast({
        title: 'Reivindicação cancelada',
        variant: 'success',
      });
      onSuccess();
    },
    onError: (error: any) => {
      toast({
        title: 'Erro ao cancelar reivindicação',
        description: error.message,
        variant: 'error',
      });
    },
  });

  // Handlers
  const onSubmit = async (data: ClaimResponseFormData) => {
    if (selectedAction === 'CONFIRM') {
      await confirmMutation.mutateAsync({
        claim_id: claim.claim_id,
        confirmation_reason: data.reason,
      });
    } else if (selectedAction === 'CANCEL') {
      await cancelMutation.mutateAsync({
        claim_id: claim.claim_id,
        reason: data.reason,
      });
    }
  };

  const handleActionSelect = (action: 'CONFIRM' | 'CANCEL') => {
    setSelectedAction(action);
  };

  // Render Action Selection
  if (!selectedAction) {
    return (
      <div className="space-y-6">
        <div>
          <h3 className="text-xl font-semibold mb-2">Responder Reivindicação</h3>
          <p className="text-gray-600">
            Chave PIX: <strong>{claim.entry_key}</strong>
          </p>
        </div>

        <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <p className="text-sm text-blue-800">
            Você tem {claim.days_remaining ?? 0} dias para responder esta reivindicação.
            Após esse período, ela será cancelada automaticamente.
          </p>
        </div>

        <div className="space-y-3">
          <Button
            variant="default"
            className="w-full"
            onClick={() => handleActionSelect('CONFIRM')}
          >
            Confirmar Transferência
          </Button>
          <Button
            variant="outline"
            className="w-full"
            onClick={() => handleActionSelect('CANCEL')}
          >
            Recusar Reivindicação
          </Button>
          <Button variant="ghost" className="w-full" onClick={onCancel}>
            Voltar
          </Button>
        </div>
      </div>
    );
  }

  // Render Response Form
  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      <div>
        <h3 className="text-xl font-semibold mb-2">
          {selectedAction === 'CONFIRM' ? 'Confirmar' : 'Recusar'} Reivindicação
        </h3>
        <p className="text-gray-600">
          Chave PIX: <strong>{claim.entry_key}</strong>
        </p>
      </div>

      {selectedAction === 'CONFIRM' && (
        <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
          <p className="text-sm text-yellow-800 font-medium">
            Atenção! Ao confirmar, você está autorizando a transferência desta chave PIX
            para outra conta. Esta ação não pode ser desfeita.
          </p>
        </div>
      )}

      {/* Reason Field */}
      <FormField
        label="Motivo"
        error={errors.reason?.message}
        help={`Por que você está ${selectedAction === 'CONFIRM' ? 'confirmando' : 'recusando'} esta reivindicação?`}
        required
      >
        <Textarea
          {...register('reason')}
          rows={4}
          placeholder={
            selectedAction === 'CONFIRM'
              ? 'Ex: Transferindo conta para nova instituição financeira'
              : 'Ex: Não reconheço esta solicitação'
          }
        />
      </FormField>

      {/* Actions */}
      <div className="flex gap-4 justify-end">
        <Button
          type="button"
          variant="outline"
          onClick={() => setSelectedAction(null)}
          disabled={isSubmitting}
        >
          Voltar
        </Button>
        <Button
          type="submit"
          variant={selectedAction === 'CONFIRM' ? 'default' : 'destructive'}
          disabled={isSubmitting}
        >
          {isSubmitting
            ? 'Processando...'
            : selectedAction === 'CONFIRM'
            ? 'Confirmar Transferência'
            : 'Recusar Reivindicação'}
        </Button>
      </div>
    </form>
  );
};
```

---

## 5. Portability Components

### 5.1. PortabilityCard Component

**Purpose**: Display portability status and information.

**File**: `components/features/portability/PortabilityCard.tsx`

```typescript
import React from 'react';
import { Card } from '@/components/shared/ui/Card';
import { Badge } from '@/components/shared/ui/Badge';
import { Button } from '@/components/shared/ui/Button';
import { Portability, PortabilityStatus } from '@/types/dict';
import { formatDistance } from 'date-fns';
import { ptBR } from 'date-fns/locale';

// Props Interface
interface PortabilityCardProps {
  portability: Portability;
  onViewDetails: () => void;
}

// Helper Functions
const getStatusBadgeVariant = (status: PortabilityStatus) => {
  switch (status) {
    case 'PENDING':
      return 'warning';
    case 'CONFIRMED':
      return 'info';
    case 'COMPLETED':
      return 'success';
    case 'CANCELLED':
      return 'neutral';
    case 'REJECTED':
      return 'error';
    default:
      return 'neutral';
  }
};

const getStatusLabel = (status: PortabilityStatus) => {
  const labels: Record<PortabilityStatus, string> = {
    PENDING: 'Pendente',
    CONFIRMED: 'Confirmado',
    COMPLETED: 'Concluído',
    CANCELLED: 'Cancelado',
    REJECTED: 'Rejeitado',
  };
  return labels[status];
};

// Component
export const PortabilityCard: React.FC<PortabilityCardProps> = ({
  portability,
  onViewDetails,
}) => {
  return (
    <Card className="p-6 hover:shadow-lg transition-shadow">
      {/* Header */}
      <div className="flex justify-between items-start mb-4">
        <div>
          <h3 className="text-lg font-semibold mb-1">
            Chave: {portability.entry_key}
          </h3>
          <p className="text-sm text-gray-500">Portabilidade de chave PIX</p>
        </div>
        <Badge variant={getStatusBadgeVariant(portability.status)}>
          {getStatusLabel(portability.status)}
        </Badge>
      </div>

      {/* Details */}
      <div className="space-y-2 mb-4">
        <div className="flex justify-between text-sm">
          <span className="text-gray-600">De:</span>
          <span className="font-medium">ISPB {portability.from_ispb}</span>
        </div>
        <div className="flex justify-between text-sm">
          <span className="text-gray-600">Para:</span>
          <span className="font-medium">ISPB {portability.to_ispb}</span>
        </div>
        <div className="flex justify-between text-sm">
          <span className="text-gray-600">Iniciado em:</span>
          <span className="font-medium">
            {formatDistance(new Date(portability.created_at), new Date(), {
              addSuffix: true,
              locale: ptBR,
            })}
          </span>
        </div>
        {portability.completed_at && (
          <div className="flex justify-between text-sm">
            <span className="text-gray-600">Concluído em:</span>
            <span className="font-medium">
              {formatDistance(new Date(portability.completed_at), new Date(), {
                addSuffix: true,
                locale: ptBR,
              })}
            </span>
          </div>
        )}
      </div>

      {/* Actions */}
      <div className="flex gap-3">
        <Button variant="outline" size="sm" onClick={onViewDetails}>
          Ver Detalhes
        </Button>
      </div>
    </Card>
  );
};
```

---

## 6. Shared Components

### 6.1. Button Component

**File**: `components/shared/ui/Button.tsx`

```typescript
import React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const buttonVariants = cva(
  'inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:opacity-50 disabled:pointer-events-none',
  {
    variants: {
      variant: {
        default: 'bg-blue-600 text-white hover:bg-blue-700',
        destructive: 'bg-red-600 text-white hover:bg-red-700',
        outline: 'border border-gray-300 bg-white hover:bg-gray-50',
        ghost: 'hover:bg-gray-100',
      },
      size: {
        sm: 'h-9 px-3',
        md: 'h-10 px-4',
        lg: 'h-11 px-8',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'md',
    },
  }
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={cn(buttonVariants({ variant, size, className }))}
        {...props}
      />
    );
  }
);

Button.displayName = 'Button';
```

---

## 7. Type Definitions

### 7.1. DICT Types

**File**: `types/dict.ts`

```typescript
// Key Types
export type KeyType = 'CPF' | 'CNPJ' | 'EMAIL' | 'PHONE' | 'EVP';
export type KeyStatus = 'ACTIVE' | 'PENDING' | 'DELETED' | 'CLAIM_PENDING';
export type AccountType = 'CACC' | 'SVGS' | 'SLRY' | 'TRAN';

// Claim Types
export type ClaimStatus =
  | 'OPEN'
  | 'WAITING_RESOLUTION'
  | 'CONFIRMED'
  | 'CANCELLED'
  | 'COMPLETED'
  | 'EXPIRED';

export type ClaimReason = 'OWNERSHIP' | 'FRAUD';

// Portability Types
export type PortabilityStatus =
  | 'PENDING'
  | 'CONFIRMED'
  | 'COMPLETED'
  | 'CANCELLED'
  | 'REJECTED';

// Interfaces
export interface Account {
  id: string;
  account_number: string;
  branch_code: string;
  account_type: AccountType;
  holder_document: string;
  holder_name: string;
  participant_ispb: string;
}

export interface Entry {
  entry_id: string;
  external_id?: string;
  key_type: KeyType;
  key_value: string;
  account?: Account;
  status: KeyStatus;
  created_at: string;
  updated_at: string;
}

export interface Claim {
  claim_id: string;
  external_id?: string;
  entry_key: string;
  status: ClaimStatus;
  completion_period_days: number;
  created_at: string;
  expires_at?: string;
  days_remaining?: number;
  claimer_ispb: string;
  owner_ispb: string;
}

export interface Portability {
  portability_id: string;
  external_id?: string;
  entry_key: string;
  status: PortabilityStatus;
  from_ispb: string;
  to_ispb: string;
  created_at: string;
  completed_at?: string;
}

// API Response Types
export interface ListEntriesResponse {
  entries: Entry[];
  pagination: {
    current_page: number;
    total_pages: number;
    total_entries: number;
    per_page: number;
  };
}

export interface ListClaimsResponse {
  claims: Claim[];
  pagination: {
    current_page: number;
    total_pages: number;
    total_claims: number;
    per_page: number;
  };
}
```

---

## 8. Validation Schemas

### 8.1. Zod Schemas

**File**: `lib/validations/dict.ts`

```typescript
import { z } from 'zod';

// CPF Validation
const cpfSchema = z
  .string()
  .refine(
    (val) => {
      const digits = val.replace(/\D/g, '');
      return digits.length === 11;
    },
    { message: 'CPF deve ter 11 dígitos' }
  )
  .refine(
    (val) => {
      const digits = val.replace(/\D/g, '');
      // Check for repeated digits
      if (/^(\d)\1{10}$/.test(digits)) return false;

      // Validate check digits
      let sum = 0;
      for (let i = 0; i < 9; i++) {
        sum += parseInt(digits[i]) * (10 - i);
      }
      let checkDigit = 11 - (sum % 11);
      if (checkDigit >= 10) checkDigit = 0;
      if (checkDigit !== parseInt(digits[9])) return false;

      sum = 0;
      for (let i = 0; i < 10; i++) {
        sum += parseInt(digits[i]) * (11 - i);
      }
      checkDigit = 11 - (sum % 11);
      if (checkDigit >= 10) checkDigit = 0;
      return checkDigit === parseInt(digits[10]);
    },
    { message: 'CPF inválido' }
  );

// CNPJ Validation
const cnpjSchema = z
  .string()
  .refine(
    (val) => {
      const digits = val.replace(/\D/g, '');
      return digits.length === 14;
    },
    { message: 'CNPJ deve ter 14 dígitos' }
  );

// Phone Validation
const phoneSchema = z
  .string()
  .refine(
    (val) => {
      const digits = val.replace(/\D/g, '');
      return /^55\d{11}$/.test(digits);
    },
    { message: 'Telefone deve estar no formato +55 11 99999-9999' }
  );

// Export schemas
export const dictSchemas = {
  cpf: cpfSchema,
  cnpj: cnpjSchema,
  email: z.string().email('Email inválido'),
  phone: phoneSchema,
};
```

---

## Rastreabilidade

### Requisitos Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RF-FE-001 | KeyList component | API-002 | ✅ Especificado |
| RF-FE-002 | CreateKeyForm component | API-002 | ✅ Especificado |
| RF-FE-003 | ClaimCard component | API-002, DIA-006 | ✅ Especificado |
| RF-FE-004 | ClaimResponseForm component | API-002 | ✅ Especificado |
| RF-FE-005 | PortabilityCard component | API-002 | ✅ Especificado |

---

## Próximas Revisões

**Pendências**:
- [ ] Implementar testes unitários (Jest + React Testing Library)
- [ ] Adicionar acessibilidade (ARIA labels, keyboard navigation)
- [ ] Implementar i18n (internacionalização)
- [ ] Adicionar Storybook para documentação de componentes
- [ ] Implementar lazy loading de componentes
- [ ] Adicionar testes E2E (Playwright/Cypress)

---

**Referências**:
- [API-002: Core DICT REST API](../../04_APIs/REST/API-002_Core_DICT_REST_API.md)
- [React 18 Documentation](https://react.dev/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Zod Documentation](https://zod.dev/)
- [React Hook Form](https://react-hook-form.com/)
