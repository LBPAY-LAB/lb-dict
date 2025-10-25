# FE-004: State Management Architecture

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Documento**: FE-004 - State Management Architecture
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: FRONTEND (AI Agent - Frontend Specialist)
**Revisor**: [Aguardando]
**Aprovador**: Tech Lead, Head de Arquitetura

---

## Sumário Executivo

Este documento especifica a **arquitetura de gerenciamento de estado** do frontend DICT LBPay, incluindo state global (Zustand), state local (React hooks), integração com API (React Query), estratégias de cache e otimizações de performance.

**Baseado em**:
- [FE-001: Component Specifications](./FE-001_Component_Specifications.md)
- [API-002: Core DICT REST API](../../04_APIs/REST/API-002_Core_DICT_REST_API.md)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | FRONTEND | Versão inicial - State Management Architecture |

---

## Índice

1. [Visão Geral](#1-visão-geral)
2. [Global State (Zustand)](#2-global-state-zustand)
3. [Server State (React Query)](#3-server-state-react-query)
4. [Local State (React Hooks)](#4-local-state-react-hooks)
5. [Caching Strategy](#5-caching-strategy)
6. [Optimistic Updates](#6-optimistic-updates)
7. [Error Handling](#7-error-handling)
8. [Performance Optimizations](#8-performance-optimizations)

---

## 1. Visão Geral

### 1.1. State Management Stack

```
┌─────────────────────────────────────────────────────────────┐
│                    Frontend State Architecture               │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────────┐  ┌─────────────────────┐           │
│  │  Global State       │  │  Server State       │           │
│  │  (Zustand)          │  │  (React Query)      │           │
│  │                     │  │                     │           │
│  │  • User Auth        │  │  • Keys List        │           │
│  │  • Current Account  │  │  • Claims List      │           │
│  │  • Notifications    │  │  • Portabilities    │           │
│  │  • UI Preferences   │  │  • Account Details  │           │
│  └─────────────────────┘  └─────────────────────┘           │
│            ↓                        ↓                        │
│  ┌──────────────────────────────────────────────┐           │
│  │           Local Component State               │           │
│  │           (React useState/useReducer)         │           │
│  │                                               │           │
│  │  • Form state (React Hook Form)               │           │
│  │  • UI state (modals, tooltips, dropdowns)     │           │
│  │  • Validation errors                          │           │
│  └──────────────────────────────────────────────┘           │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 1.2. State Categories

| Category | Technology | Purpose | Examples |
|----------|-----------|---------|----------|
| **Global State** | Zustand | App-wide state shared across components | User auth, current account, theme |
| **Server State** | React Query | Remote data fetched from API, cached | Keys, claims, portabilities |
| **Local State** | React Hooks | Component-specific state | Form inputs, modal open/close |
| **Form State** | React Hook Form | Form-specific state with validation | CreateKeyForm, ClaimResponseForm |

### 1.3. Design Principles

1. **Separation of Concerns**: Global vs Server vs Local state
2. **Single Source of Truth**: Each piece of state has one owner
3. **Immutability**: State updates are immutable (Immer for Zustand)
4. **Optimistic Updates**: Update UI before API response
5. **Cache Invalidation**: Smart cache invalidation strategies
6. **Type Safety**: Full TypeScript coverage

---

## 2. Global State (Zustand)

### 2.1. Store Structure

**File**: `store/index.ts`

```typescript
import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';

// Types
interface User {
  id: string;
  email: string;
  name: string;
  roles: string[];
  scopes: string[];
}

interface Account {
  id: string;
  account_number: string;
  branch_code: string;
  account_type: 'CACC' | 'SVGS' | 'SLRY' | 'TRAN';
  holder_name: string;
  holder_document: string;
  participant_ispb: string;
}

interface Notification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message: string;
  timestamp: Date;
  read: boolean;
}

interface AppState {
  // Auth State
  user: User | null;
  accessToken: string | null;
  isAuthenticated: boolean;

  // Account State
  currentAccount: Account | null;
  availableAccounts: Account[];

  // Notifications State
  notifications: Notification[];
  unreadNotificationsCount: number;

  // UI State
  theme: 'light' | 'dark';
  sidebarOpen: boolean;

  // Actions
  setUser: (user: User | null) => void;
  setAccessToken: (token: string | null) => void;
  login: (user: User, token: string) => void;
  logout: () => void;

  setCurrentAccount: (account: Account) => void;
  setAvailableAccounts: (accounts: Account[]) => void;

  addNotification: (notification: Omit<Notification, 'id' | 'timestamp' | 'read'>) => void;
  markNotificationAsRead: (id: string) => void;
  clearNotifications: () => void;

  setTheme: (theme: 'light' | 'dark') => void;
  toggleSidebar: () => void;
}

// Store
export const useAppStore = create<AppState>()(
  devtools(
    persist(
      immer((set) => ({
        // Initial State
        user: null,
        accessToken: null,
        isAuthenticated: false,
        currentAccount: null,
        availableAccounts: [],
        notifications: [],
        unreadNotificationsCount: 0,
        theme: 'light',
        sidebarOpen: true,

        // Auth Actions
        setUser: (user) =>
          set((state) => {
            state.user = user;
            state.isAuthenticated = user !== null;
          }),

        setAccessToken: (token) =>
          set((state) => {
            state.accessToken = token;
          }),

        login: (user, token) =>
          set((state) => {
            state.user = user;
            state.accessToken = token;
            state.isAuthenticated = true;
          }),

        logout: () =>
          set((state) => {
            state.user = null;
            state.accessToken = null;
            state.isAuthenticated = false;
            state.currentAccount = null;
            state.availableAccounts = [];
            state.notifications = [];
            state.unreadNotificationsCount = 0;
          }),

        // Account Actions
        setCurrentAccount: (account) =>
          set((state) => {
            state.currentAccount = account;
          }),

        setAvailableAccounts: (accounts) =>
          set((state) => {
            state.availableAccounts = accounts;
            if (accounts.length > 0 && !state.currentAccount) {
              state.currentAccount = accounts[0];
            }
          }),

        // Notification Actions
        addNotification: (notification) =>
          set((state) => {
            const newNotification: Notification = {
              ...notification,
              id: crypto.randomUUID(),
              timestamp: new Date(),
              read: false,
            };
            state.notifications.unshift(newNotification);
            state.unreadNotificationsCount += 1;
          }),

        markNotificationAsRead: (id) =>
          set((state) => {
            const notification = state.notifications.find((n) => n.id === id);
            if (notification && !notification.read) {
              notification.read = true;
              state.unreadNotificationsCount -= 1;
            }
          }),

        clearNotifications: () =>
          set((state) => {
            state.notifications = [];
            state.unreadNotificationsCount = 0;
          }),

        // UI Actions
        setTheme: (theme) =>
          set((state) => {
            state.theme = theme;
          }),

        toggleSidebar: () =>
          set((state) => {
            state.sidebarOpen = !state.sidebarOpen;
          }),
      })),
      {
        name: 'lbpay-dict-storage',
        partialize: (state) => ({
          // Only persist these fields
          user: state.user,
          accessToken: state.accessToken,
          isAuthenticated: state.isAuthenticated,
          theme: state.theme,
          sidebarOpen: state.sidebarOpen,
        }),
      }
    )
  )
);
```

### 2.2. Using Global State

```typescript
// In a component
import { useAppStore } from '@/store';

function Header() {
  const user = useAppStore((state) => state.user);
  const logout = useAppStore((state) => state.logout);
  const notificationsCount = useAppStore((state) => state.unreadNotificationsCount);

  return (
    <header>
      <span>Welcome, {user?.name}</span>
      <NotificationBell count={notificationsCount} />
      <button onClick={logout}>Logout</button>
    </header>
  );
}
```

---

## 3. Server State (React Query)

### 3.1. Query Client Setup

**File**: `lib/queryClient.ts`

```typescript
import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Stale time: 5 minutes
      staleTime: 5 * 60 * 1000,

      // Cache time: 10 minutes
      gcTime: 10 * 60 * 1000,

      // Retry failed requests 3 times
      retry: 3,

      // Retry delay: exponential backoff
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),

      // Refetch on window focus
      refetchOnWindowFocus: true,

      // Refetch on reconnect
      refetchOnReconnect: true,
    },
    mutations: {
      // Retry failed mutations once
      retry: 1,
    },
  },
});
```

### 3.2. API Client

**File**: `lib/apiClient.ts`

```typescript
import axios, { AxiosError, AxiosRequestConfig } from 'axios';
import { useAppStore } from '@/store';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 30000, // 30 seconds
});

// Request interceptor: Add JWT token
apiClient.interceptors.request.use(
  (config) => {
    const accessToken = useAppStore.getState().accessToken;
    if (accessToken) {
      config.headers.Authorization = `Bearer ${accessToken}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor: Handle errors globally
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    // 401: Unauthorized - logout user
    if (error.response?.status === 401) {
      useAppStore.getState().logout();
      window.location.href = '/login';
    }

    // 429: Rate limit exceeded
    if (error.response?.status === 429) {
      const retryAfter = error.response.headers['retry-after'];
      console.error(`Rate limit exceeded. Retry after ${retryAfter} seconds.`);
    }

    return Promise.reject(error);
  }
);

// Generic request function
export async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const response = await apiClient.request<T>(config);
  return response.data;
}
```

### 3.3. Keys Management Hooks

**File**: `hooks/useKeyManagement.ts`

```typescript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { request } from '@/lib/apiClient';
import { Entry, ListEntriesResponse, CreateKeyRequest } from '@/types/dict';

// Query Keys
export const keysQueryKeys = {
  all: ['keys'] as const,
  lists: () => [...keysQueryKeys.all, 'list'] as const,
  list: (filters: { page?: number; limit?: number; status?: string }) =>
    [...keysQueryKeys.lists(), filters] as const,
  details: () => [...keysQueryKeys.all, 'detail'] as const,
  detail: (id: string) => [...keysQueryKeys.details(), id] as const,
};

// Hooks
export function useKeyManagement() {
  const queryClient = useQueryClient();

  // List Keys
  const useListKeys = (params: { page?: number; limit?: number; status?: string }) =>
    useQuery({
      queryKey: keysQueryKeys.list(params),
      queryFn: () =>
        request<ListEntriesResponse>({
          url: '/keys',
          method: 'GET',
          params,
        }),
    });

  // Get Key Details
  const useKeyDetails = (keyId: string) =>
    useQuery({
      queryKey: keysQueryKeys.detail(keyId),
      queryFn: () =>
        request<Entry>({
          url: `/keys/${keyId}`,
          method: 'GET',
        }),
      enabled: !!keyId, // Only run if keyId is provided
    });

  // Create Key
  const createKey = useMutation({
    mutationFn: (data: CreateKeyRequest) =>
      request<Entry>({
        url: '/keys',
        method: 'POST',
        data,
      }),
    onSuccess: () => {
      // Invalidate keys list to refetch
      queryClient.invalidateQueries({ queryKey: keysQueryKeys.lists() });
    },
  });

  // Delete Key
  const deleteKey = useMutation({
    mutationFn: (keyId: string) =>
      request<void>({
        url: `/keys/${keyId}`,
        method: 'DELETE',
      }),
    onSuccess: (_, keyId) => {
      // Remove from cache
      queryClient.removeQueries({ queryKey: keysQueryKeys.detail(keyId) });
      // Invalidate keys list
      queryClient.invalidateQueries({ queryKey: keysQueryKeys.lists() });
    },
  });

  return {
    useListKeys,
    useKeyDetails,
    createKey,
    deleteKey,
  };
}
```

### 3.4. Claims Management Hooks

**File**: `hooks/useClaimManagement.ts`

```typescript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { request } from '@/lib/apiClient';
import { Claim, ListClaimsResponse, CreateClaimRequest } from '@/types/dict';

// Query Keys
export const claimsQueryKeys = {
  all: ['claims'] as const,
  lists: () => [...claimsQueryKeys.all, 'list'] as const,
  list: (filters: { page?: number; limit?: number; status?: string; role?: string }) =>
    [...claimsQueryKeys.lists(), filters] as const,
  details: () => [...claimsQueryKeys.all, 'detail'] as const,
  detail: (id: string) => [...claimsQueryKeys.details(), id] as const,
};

// Hooks
export function useClaimManagement() {
  const queryClient = useQueryClient();

  // List Claims
  const useListClaims = (params: {
    page?: number;
    limit?: number;
    status?: string;
    role?: string;
  }) =>
    useQuery({
      queryKey: claimsQueryKeys.list(params),
      queryFn: () =>
        request<ListClaimsResponse>({
          url: '/claims',
          method: 'GET',
          params,
        }),
    });

  // Get Claim Details
  const useClaimDetails = (claimId: string) =>
    useQuery({
      queryKey: claimsQueryKeys.detail(claimId),
      queryFn: () =>
        request<Claim>({
          url: `/claims/${claimId}`,
          method: 'GET',
        }),
      enabled: !!claimId,
      // Refetch every 30 seconds (to track countdown)
      refetchInterval: 30000,
    });

  // Create Claim
  const createClaim = useMutation({
    mutationFn: (data: CreateClaimRequest) =>
      request<Claim>({
        url: '/claims',
        method: 'POST',
        data,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: claimsQueryKeys.lists() });
    },
  });

  // Confirm Claim
  const confirmClaim = useMutation({
    mutationFn: (data: { claim_id: string; confirmation_reason: string }) =>
      request<Claim>({
        url: `/claims/${data.claim_id}/confirm`,
        method: 'PUT',
        data: { confirmation_reason: data.confirmation_reason },
      }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: claimsQueryKeys.detail(variables.claim_id) });
      queryClient.invalidateQueries({ queryKey: claimsQueryKeys.lists() });
    },
  });

  // Cancel Claim
  const cancelClaim = useMutation({
    mutationFn: (data: { claim_id: string; reason?: string }) =>
      request<Claim>({
        url: `/claims/${data.claim_id}`,
        method: 'DELETE',
        params: { reason: data.reason },
      }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: claimsQueryKeys.detail(variables.claim_id) });
      queryClient.invalidateQueries({ queryKey: claimsQueryKeys.lists() });
    },
  });

  return {
    useListClaims,
    useClaimDetails,
    createClaim,
    confirmClaim,
    cancelClaim,
  };
}
```

---

## 4. Local State (React Hooks)

### 4.1. Form State with React Hook Form

```typescript
// Example: CreateKeyForm
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { createKeySchema } from '@/lib/validations/dict';

function CreateKeyForm() {
  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isSubmitting },
    reset,
  } = useForm({
    resolver: zodResolver(createKeySchema),
    defaultValues: {
      key_type: 'CPF',
      key_value: '',
      account_id: '',
    },
  });

  const onSubmit = async (data) => {
    // Handle submission
  };

  return <form onSubmit={handleSubmit(onSubmit)}>...</form>;
}
```

### 4.2. UI State with useState

```typescript
// Example: Modal state
function KeyDetailsPage() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedKeyId, setSelectedKeyId] = useState<string | null>(null);

  const handleViewDetails = (keyId: string) => {
    setSelectedKeyId(keyId);
    setIsModalOpen(true);
  };

  return (
    <>
      <KeyList onViewDetails={handleViewDetails} />
      {isModalOpen && (
        <Modal onClose={() => setIsModalOpen(false)}>
          <KeyDetailsCard keyId={selectedKeyId} />
        </Modal>
      )}
    </>
  );
}
```

### 4.3. Complex State with useReducer

```typescript
// Example: Multi-step form wizard
type WizardStep = 'select-type' | 'enter-value' | 'confirm';

interface WizardState {
  currentStep: WizardStep;
  data: Partial<CreateKeyFormData>;
  canGoBack: boolean;
  canGoNext: boolean;
}

type WizardAction =
  | { type: 'NEXT_STEP'; payload?: Partial<CreateKeyFormData> }
  | { type: 'PREVIOUS_STEP' }
  | { type: 'RESET' };

function wizardReducer(state: WizardState, action: WizardAction): WizardState {
  switch (action.type) {
    case 'NEXT_STEP':
      // Logic to move to next step
      return { ...state, currentStep: getNextStep(state.currentStep), data: { ...state.data, ...action.payload } };
    case 'PREVIOUS_STEP':
      return { ...state, currentStep: getPreviousStep(state.currentStep) };
    case 'RESET':
      return initialState;
    default:
      return state;
  }
}

function CreateKeyWizard() {
  const [state, dispatch] = useReducer(wizardReducer, initialState);

  return <div>...</div>;
}
```

---

## 5. Caching Strategy

### 5.1. Cache Configuration

| Resource | Stale Time | GC Time | Refetch on Focus | Refetch Interval |
|----------|------------|---------|------------------|------------------|
| Keys List | 5 min | 10 min | ✅ Yes | ❌ No |
| Key Details | 5 min | 10 min | ✅ Yes | ❌ No |
| Claims List | 1 min | 5 min | ✅ Yes | ❌ No |
| Claim Details | 30 sec | 5 min | ✅ Yes | ✅ 30 sec (countdown) |
| Portabilities | 5 min | 10 min | ✅ Yes | ❌ No |
| User Profile | 10 min | 30 min | ❌ No | ❌ No |

### 5.2. Cache Invalidation Rules

```typescript
// When to invalidate cache
const cacheInvalidationRules = {
  // After creating a key
  'keys.create': ['keys.list'],

  // After deleting a key
  'keys.delete': ['keys.list', 'keys.detail:{id}'],

  // After creating a claim
  'claims.create': ['claims.list', 'keys.detail:{entry_key}'],

  // After confirming/cancelling a claim
  'claims.respond': ['claims.detail:{id}', 'claims.list', 'keys.list'],

  // After logging out
  logout: ['*'], // Invalidate all
};
```

### 5.3. Prefetching

```typescript
// Prefetch key details on hover
function KeyCard({ entry }: { entry: Entry }) {
  const queryClient = useQueryClient();

  const handleMouseEnter = () => {
    queryClient.prefetchQuery({
      queryKey: keysQueryKeys.detail(entry.entry_id),
      queryFn: () =>
        request<Entry>({
          url: `/keys/${entry.entry_id}`,
          method: 'GET',
        }),
    });
  };

  return <div onMouseEnter={handleMouseEnter}>...</div>;
}
```

---

## 6. Optimistic Updates

### 6.1. Optimistic Delete Key

```typescript
const deleteKey = useMutation({
  mutationFn: (keyId: string) =>
    request<void>({
      url: `/keys/${keyId}`,
      method: 'DELETE',
    }),

  // Optimistic update
  onMutate: async (keyId) => {
    // Cancel outgoing refetches
    await queryClient.cancelQueries({ queryKey: keysQueryKeys.lists() });

    // Snapshot previous value
    const previousKeys = queryClient.getQueryData<ListEntriesResponse>(
      keysQueryKeys.list({ page: 1, limit: 20 })
    );

    // Optimistically update cache
    queryClient.setQueryData<ListEntriesResponse>(
      keysQueryKeys.list({ page: 1, limit: 20 }),
      (old) => {
        if (!old) return old;
        return {
          ...old,
          entries: old.entries.filter((entry) => entry.entry_id !== keyId),
        };
      }
    );

    // Return context with snapshot
    return { previousKeys };
  },

  // If mutation fails, rollback
  onError: (err, keyId, context) => {
    queryClient.setQueryData(
      keysQueryKeys.list({ page: 1, limit: 20 }),
      context?.previousKeys
    );
  },

  // Always refetch after mutation (success or error)
  onSettled: () => {
    queryClient.invalidateQueries({ queryKey: keysQueryKeys.lists() });
  },
});
```

### 6.2. Optimistic Create Claim

```typescript
const createClaim = useMutation({
  mutationFn: (data: CreateClaimRequest) =>
    request<Claim>({
      url: '/claims',
      method: 'POST',
      data,
    }),

  onMutate: async (newClaim) => {
    await queryClient.cancelQueries({ queryKey: claimsQueryKeys.lists() });

    const previousClaims = queryClient.getQueryData<ListClaimsResponse>(
      claimsQueryKeys.list({ page: 1, limit: 20 })
    );

    // Optimistically add new claim
    queryClient.setQueryData<ListClaimsResponse>(
      claimsQueryKeys.list({ page: 1, limit: 20 }),
      (old) => {
        if (!old) return old;
        const optimisticClaim: Claim = {
          claim_id: 'temp-' + Date.now(),
          entry_key: newClaim.entry_key,
          status: 'OPEN',
          completion_period_days: 30,
          created_at: new Date().toISOString(),
          expires_at: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
          days_remaining: 30,
          claimer_ispb: newClaim.claimer_account.participant_ispb,
          owner_ispb: '', // Unknown at this point
        };
        return {
          ...old,
          claims: [optimisticClaim, ...old.claims],
        };
      }
    );

    return { previousClaims };
  },

  onError: (err, newClaim, context) => {
    queryClient.setQueryData(
      claimsQueryKeys.list({ page: 1, limit: 20 }),
      context?.previousClaims
    );
  },

  onSettled: () => {
    queryClient.invalidateQueries({ queryKey: claimsQueryKeys.lists() });
  },
});
```

---

## 7. Error Handling

### 7.1. Error Boundaries

```typescript
// components/ErrorBoundary.tsx
import React, { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    // Send to error tracking service (Sentry, etc.)
  }

  render() {
    if (this.state.hasError) {
      return (
        this.props.fallback || (
          <div className="p-6 bg-red-50 border border-red-200 rounded-lg">
            <h2 className="text-lg font-semibold text-red-800">Algo deu errado</h2>
            <p className="text-red-600">
              Ocorreu um erro inesperado. Por favor, recarregue a página.
            </p>
            <button
              onClick={() => window.location.reload()}
              className="mt-4 px-4 py-2 bg-red-600 text-white rounded"
            >
              Recarregar Página
            </button>
          </div>
        )
      );
    }

    return this.props.children;
  }
}
```

### 7.2. Query Error Handling

```typescript
function KeyList() {
  const { data, error, isLoading, isError } = useListKeys({ page: 1, limit: 20 });

  if (isError) {
    return (
      <ErrorState
        title="Erro ao carregar chaves PIX"
        message={error?.message || 'Tente novamente mais tarde'}
        retry={() => queryClient.invalidateQueries({ queryKey: keysQueryKeys.lists() })}
      />
    );
  }

  // Render content
}
```

### 7.3. Global Error Handling with Toast

```typescript
// hooks/useApiError.ts
import { useToast } from '@/hooks/useToast';
import { AxiosError } from 'axios';

export function useApiError() {
  const { toast } = useToast();

  const handleApiError = (error: unknown) => {
    if (error instanceof AxiosError) {
      const errorCode = error.response?.data?.error?.code;
      const errorMessage = error.response?.data?.error?.message;

      // Map error codes to user-friendly messages
      const messages: Record<string, string> = {
        KEY_ALREADY_EXISTS: 'Esta chave PIX já está cadastrada',
        MAX_KEYS_EXCEEDED: 'Você atingiu o limite máximo de chaves PIX (20)',
        VALIDATION_ERROR: 'Dados inválidos. Verifique e tente novamente',
        RATE_LIMIT_EXCEEDED: 'Muitas requisições. Aguarde alguns segundos',
      };

      toast({
        title: 'Erro',
        description: messages[errorCode] || errorMessage || 'Erro desconhecido',
        variant: 'error',
      });
    } else {
      toast({
        title: 'Erro',
        description: 'Ocorreu um erro inesperado',
        variant: 'error',
      });
    }
  };

  return { handleApiError };
}
```

---

## 8. Performance Optimizations

### 8.1. React.memo for Expensive Components

```typescript
import React, { memo } from 'react';

export const KeyCard = memo<KeyCardProps>(
  ({ entry, onViewDetails, onDelete }) => {
    // Component implementation
  },
  (prevProps, nextProps) => {
    // Only re-render if entry changes
    return prevProps.entry.entry_id === nextProps.entry.entry_id;
  }
);
```

### 8.2. useMemo for Expensive Computations

```typescript
function ClaimsList({ claims }: { claims: Claim[] }) {
  const expiringSoonClaims = useMemo(() => {
    return claims.filter((claim) => {
      const daysRemaining = claim.days_remaining ?? 0;
      return daysRemaining > 0 && daysRemaining <= 5;
    });
  }, [claims]);

  return (
    <div>
      {expiringSoonClaims.length > 0 && (
        <Alert>Você tem {expiringSoonClaims.length} reivindicações expirando em breve!</Alert>
      )}
      {/* Render claims */}
    </div>
  );
}
```

### 8.3. useCallback for Event Handlers

```typescript
function KeyList() {
  const [selectedKeyId, setSelectedKeyId] = useState<string | null>(null);

  const handleViewDetails = useCallback((keyId: string) => {
    setSelectedKeyId(keyId);
  }, []);

  return (
    <div>
      {keys.map((key) => (
        <KeyCard key={key.entry_id} entry={key} onViewDetails={handleViewDetails} />
      ))}
    </div>
  );
}
```

### 8.4. Lazy Loading Components

```typescript
import { lazy, Suspense } from 'react';

// Lazy load heavy components
const ClaimResponseForm = lazy(() => import('@/components/features/claims/ClaimResponseForm'));

function ClaimDetailsPage() {
  return (
    <Suspense fallback={<Skeleton className="h-64 w-full" />}>
      <ClaimResponseForm claim={claim} />
    </Suspense>
  );
}
```

### 8.5. Virtual Scrolling for Long Lists

```typescript
import { useVirtualizer } from '@tanstack/react-virtual';

function KeyListVirtualized({ keys }: { keys: Entry[] }) {
  const parentRef = React.useRef<HTMLDivElement>(null);

  const virtualizer = useVirtualizer({
    count: keys.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 80, // Estimated height of each item
  });

  return (
    <div ref={parentRef} style={{ height: '600px', overflow: 'auto' }}>
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          width: '100%',
          position: 'relative',
        }}
      >
        {virtualizer.getVirtualItems().map((virtualItem) => (
          <div
            key={virtualItem.key}
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%',
              height: `${virtualItem.size}px`,
              transform: `translateY(${virtualItem.start}px)`,
            }}
          >
            <KeyCard entry={keys[virtualItem.index]} />
          </div>
        ))}
      </div>
    </div>
  );
}
```

---

## Rastreabilidade

### Requisitos de State Management

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RF-SM-001 | Global state com Zustand | FE-001 | ✅ Especificado |
| RF-SM-002 | Server state com React Query | FE-001 | ✅ Especificado |
| RF-SM-003 | Caching strategy | Best Practices | ✅ Especificado |
| RF-SM-004 | Optimistic updates | Best Practices | ✅ Especificado |
| RF-SM-005 | Error handling | Best Practices | ✅ Especificado |
| RF-SM-006 | Performance optimizations | Best Practices | ✅ Especificado |

---

## Próximas Revisões

**Pendências**:
- [ ] Implementar offline-first strategy (Service Worker + IndexedDB)
- [ ] Adicionar telemetria para monitorar performance de queries
- [ ] Implementar retry logic customizado por tipo de erro
- [ ] Adicionar background sync para mutations failed
- [ ] Implementar WebSockets para real-time updates (claims countdown)

---

**Referências**:
- [FE-001: Component Specifications](./FE-001_Component_Specifications.md)
- [API-002: Core DICT REST API](../../04_APIs/REST/API-002_Core_DICT_REST_API.md)
- [Zustand Documentation](https://docs.pmnd.rs/zustand/getting-started/introduction)
- [React Query Documentation](https://tanstack.com/query/latest)
- [React Hook Form Documentation](https://react-hook-form.com/)
