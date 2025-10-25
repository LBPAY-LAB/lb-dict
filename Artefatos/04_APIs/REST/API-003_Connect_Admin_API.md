# API-003: Connect Admin API Specification

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Serviço**: Connect Admin API
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: BACKEND (AI Agent - Backend Specialist)
**Revisor**: [Aguardando]
**Aprovador**: Tech Lead, Head de Arquitetura

---

## Sumário Executivo

Este documento especifica a **Connect Admin API**, uma API REST para operações administrativas do RSFN Connect. Permite gerenciar workflows Temporal, monitorar tarefas, visualizar status de sincronização e executar operações de manutenção.

**Baseado em**:
- [TEC-003: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [API-002: Core DICT REST API](./API-002_Core_DICT_REST_API.md)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | BACKEND | Versão inicial - Connect Admin API |

---

## Índice

1. [Visão Geral](#1-visão-geral)
2. [Authentication & Authorization](#2-authentication--authorization)
3. [Workflow Management Endpoints](#3-workflow-management-endpoints)
4. [Monitoring Endpoints](#4-monitoring-endpoints)
5. [Maintenance Endpoints](#5-maintenance-endpoints)
6. [Error Handling](#6-error-handling)

---

## 1. Visão Geral

### 1.1. Base URL

```
Production:  https://api.lbpay.com.br/connect/admin/v1
Staging:     https://api-stg.lbpay.com.br/connect/admin/v1
Development: http://localhost:8081/admin/v1
```

### 1.2. Características

| Aspecto | Valor |
|---------|-------|
| **Protocol** | HTTPS (TLS 1.2+) |
| **Content-Type** | application/json |
| **Authentication** | JWT Bearer Token |
| **Authorization** | RBAC (Admin Role Required) |
| **API Version** | v1 (URL-based versioning) |
| **Rate Limit** | 500 req/min per admin |
| **Timeout** | 60s |

### 1.3. HTTP Methods

| Method | Usage | Idempotent |
|--------|-------|------------|
| `GET` | Retrieve resources | ✅ Yes |
| `POST` | Create/Execute operations | ❌ No |
| `DELETE` | Cancel operations | ✅ Yes |

---

## 2. Authentication & Authorization

### 2.1. Authentication (JWT)

**Header**:
```http
Authorization: Bearer <jwt-token>
```

**JWT Payload**:
```json
{
  "sub": "admin-user-uuid",
  "email": "admin@lbpay.com.br",
  "roles": ["DICT_ADMIN"],
  "scopes": ["dict:admin", "dict:read", "dict:write"],
  "iss": "lbpay-auth",
  "exp": 1698350400,
  "iat": 1698264000
}
```

### 2.2. Authorization (RBAC)

**Required Role**: `DICT_ADMIN`

**Endpoint Permissions**:

| Endpoint | Required Scope | Required Role |
|----------|---------------|---------------|
| `GET /admin/v1/workflows` | `dict:admin` | `DICT_ADMIN` |
| `POST /admin/v1/workflows` | `dict:admin` | `DICT_ADMIN` |
| `GET /admin/v1/workflows/:id` | `dict:admin` | `DICT_ADMIN` |
| `POST /admin/v1/workflows/:id/cancel` | `dict:admin` | `DICT_ADMIN` |
| `GET /admin/v1/health` | `dict:read` | `DICT_ADMIN` |

### 2.3. Error Responses (Auth)

**401 Unauthorized**:
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Missing or invalid JWT token",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

**403 Forbidden**:
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "Admin role required. Current roles: [DICT_USER]",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

---

## 3. Workflow Management Endpoints

### 3.1. List Workflows

**Endpoint**: `GET /admin/v1/workflows`

**Description**: List all Temporal workflows with filtering and pagination.

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:admin`, Role: `DICT_ADMIN`

**Query Parameters**:
- `page` (integer, default: 1): Page number
- `limit` (integer, default: 50, max: 100): Results per page
- `workflow_type` (string, optional): Filter by workflow type
  - Values: `ENTRY_CREATE`, `ENTRY_DELETE`, `CLAIM_CREATE`, `CLAIM_CONFIRM`, `PORTABILITY_CREATE`
- `status` (string, optional): Filter by status
  - Values: `RUNNING`, `COMPLETED`, `FAILED`, `CANCELLED`, `TIMED_OUT`
- `start_time_from` (string, optional): Filter by start time (ISO 8601)
- `start_time_to` (string, optional): Filter by start time (ISO 8601)

**Example Request**:
```http
GET /admin/v1/workflows?page=1&limit=20&workflow_type=ENTRY_CREATE&status=RUNNING
Authorization: Bearer <jwt-token>
```

**Response 200 OK**:
```json
{
  "workflows": [
    {
      "workflow_id": "entry-create-uuid-123",
      "workflow_type": "ENTRY_CREATE",
      "status": "RUNNING",
      "start_time": "2025-10-25T10:30:00Z",
      "close_time": null,
      "execution_time_ms": 5000,
      "input": {
        "entry_id": "uuid-v4",
        "key_type": "CPF",
        "key_value": "12345678901"
      },
      "current_activity": "SyncToBacenDICT",
      "history_length": 15,
      "retry_count": 0
    },
    {
      "workflow_id": "claim-create-uuid-456",
      "workflow_type": "CLAIM_CREATE",
      "status": "COMPLETED",
      "start_time": "2025-10-25T09:15:00Z",
      "close_time": "2025-10-25T09:17:00Z",
      "execution_time_ms": 120000,
      "input": {
        "claim_id": "uuid-v5",
        "entry_key": "user@example.com"
      },
      "current_activity": null,
      "history_length": 42,
      "retry_count": 1
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 5,
    "total_workflows": 94,
    "per_page": 20
  }
}
```

**Error Responses**:

**400 Bad Request** (Invalid filters):
```json
{
  "error": {
    "code": "INVALID_QUERY_PARAMS",
    "message": "Invalid workflow_type: INVALID_TYPE",
    "details": [
      {
        "field": "workflow_type",
        "message": "Must be one of: ENTRY_CREATE, ENTRY_DELETE, CLAIM_CREATE, CLAIM_CONFIRM, PORTABILITY_CREATE"
      }
    ],
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

---

### 3.2. Get Workflow Details

**Endpoint**: `GET /admin/v1/workflows/:id`

**Description**: Retrieve detailed information about a specific workflow.

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:admin`, Role: `DICT_ADMIN`

**Path Parameters**:
- `id` (string): Workflow ID

**Query Parameters**:
- `include_history` (boolean, default: false): Include full workflow history
- `include_events` (boolean, default: false): Include workflow events

**Example Request**:
```http
GET /admin/v1/workflows/entry-create-uuid-123?include_history=true
Authorization: Bearer <jwt-token>
```

**Response 200 OK**:
```json
{
  "workflow_id": "entry-create-uuid-123",
  "workflow_type": "ENTRY_CREATE",
  "status": "RUNNING",
  "start_time": "2025-10-25T10:30:00Z",
  "close_time": null,
  "execution_time_ms": 5000,
  "input": {
    "entry_id": "uuid-v4",
    "key_type": "CPF",
    "key_value": "12345678901",
    "account": {
      "account_number": "12345678",
      "branch_code": "0001",
      "participant_ispb": "12345678"
    }
  },
  "output": null,
  "current_activity": "SyncToBacenDICT",
  "pending_activities": [
    {
      "activity_id": "activity-uuid-1",
      "activity_type": "SyncToBacenDICT",
      "scheduled_time": "2025-10-25T10:30:05Z",
      "attempt": 1,
      "max_attempts": 3
    }
  ],
  "history_length": 15,
  "retry_count": 0,
  "last_failure": null,
  "stack_trace": null,
  "history": [
    {
      "event_id": 1,
      "event_type": "WorkflowExecutionStarted",
      "timestamp": "2025-10-25T10:30:00Z",
      "attributes": {
        "workflow_type": "ENTRY_CREATE",
        "task_queue": "dict-task-queue"
      }
    },
    {
      "event_id": 2,
      "event_type": "ActivityTaskScheduled",
      "timestamp": "2025-10-25T10:30:01Z",
      "attributes": {
        "activity_type": "ValidateEntry",
        "activity_id": "activity-uuid-0"
      }
    }
  ]
}
```

**Error Responses**:

**404 Not Found**:
```json
{
  "error": {
    "code": "WORKFLOW_NOT_FOUND",
    "message": "Workflow not found: entry-create-uuid-999",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

---

### 3.3. Retry Failed Workflow

**Endpoint**: `POST /admin/v1/workflows`

**Description**: Manually retry a failed workflow or create a new workflow execution.

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:admin`, Role: `DICT_ADMIN`

**Request Body**:
```json
{
  "workflow_type": "ENTRY_CREATE",
  "input": {
    "entry_id": "uuid-v4",
    "key_type": "CPF",
    "key_value": "12345678901",
    "account_id": "account-uuid"
  },
  "workflow_id": "custom-workflow-id-123",
  "task_queue": "dict-task-queue",
  "execution_timeout_seconds": 300,
  "retry_policy": {
    "max_attempts": 3,
    "initial_interval_seconds": 1,
    "backoff_coefficient": 2.0
  }
}
```

**Request Schema**:
```json
{
  "workflow_type": {
    "type": "string",
    "enum": ["ENTRY_CREATE", "ENTRY_DELETE", "CLAIM_CREATE", "CLAIM_CONFIRM", "PORTABILITY_CREATE"],
    "required": true
  },
  "input": {
    "type": "object",
    "required": true,
    "description": "Workflow-specific input parameters"
  },
  "workflow_id": {
    "type": "string",
    "required": false,
    "description": "Custom workflow ID (auto-generated if not provided)"
  },
  "task_queue": {
    "type": "string",
    "required": false,
    "default": "dict-task-queue"
  },
  "execution_timeout_seconds": {
    "type": "integer",
    "required": false,
    "default": 300
  }
}
```

**Response 201 Created**:
```json
{
  "workflow_id": "custom-workflow-id-123",
  "workflow_type": "ENTRY_CREATE",
  "status": "RUNNING",
  "start_time": "2025-10-25T10:35:00Z",
  "task_queue": "dict-task-queue",
  "message": "Workflow started successfully"
}
```

**Error Responses**:

**409 Conflict** (Workflow already exists):
```json
{
  "error": {
    "code": "WORKFLOW_ALREADY_EXISTS",
    "message": "Workflow with ID 'custom-workflow-id-123' already exists",
    "existing_workflow_status": "RUNNING",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

---

### 3.4. Cancel Workflow

**Endpoint**: `POST /admin/v1/workflows/:id/cancel`

**Description**: Cancel a running workflow.

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:admin`, Role: `DICT_ADMIN`

**Path Parameters**:
- `id` (string): Workflow ID

**Request Body**:
```json
{
  "reason": "Manual cancellation by admin due to duplicate request"
}
```

**Example Request**:
```http
POST /admin/v1/workflows/entry-create-uuid-123/cancel
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "reason": "Manual cancellation by admin due to duplicate request"
}
```

**Response 200 OK**:
```json
{
  "workflow_id": "entry-create-uuid-123",
  "status": "CANCELLED",
  "cancelled_at": "2025-10-25T10:40:00Z",
  "cancellation_reason": "Manual cancellation by admin due to duplicate request",
  "message": "Workflow cancelled successfully"
}
```

**Error Responses**:

**404 Not Found**:
```json
{
  "error": {
    "code": "WORKFLOW_NOT_FOUND",
    "message": "Workflow not found: entry-create-uuid-999",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

**409 Conflict** (Already completed):
```json
{
  "error": {
    "code": "WORKFLOW_ALREADY_COMPLETED",
    "message": "Cannot cancel workflow: already completed",
    "current_status": "COMPLETED",
    "completed_at": "2025-10-25T09:17:00Z",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

---

## 4. Monitoring Endpoints

### 4.1. Get System Health

**Endpoint**: `GET /admin/v1/health`

**Description**: Retrieve system health status and metrics.

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:read`, Role: `DICT_ADMIN`

**Example Request**:
```http
GET /admin/v1/health
Authorization: Bearer <jwt-token>
```

**Response 200 OK**:
```json
{
  "status": "healthy",
  "timestamp": "2025-10-25T10:30:00Z",
  "services": {
    "temporal": {
      "status": "up",
      "response_time_ms": 12,
      "worker_count": 5,
      "task_queue_pollers": 10
    },
    "pulsar": {
      "status": "up",
      "response_time_ms": 8,
      "connected_consumers": 3,
      "connected_producers": 2,
      "pending_messages": 0
    },
    "postgres": {
      "status": "up",
      "response_time_ms": 5,
      "active_connections": 15,
      "max_connections": 100,
      "connection_utilization": "15%"
    },
    "redis": {
      "status": "up",
      "response_time_ms": 2,
      "memory_used_mb": 128,
      "memory_max_mb": 512,
      "memory_utilization": "25%"
    },
    "bacen_dict": {
      "status": "up",
      "response_time_ms": 45,
      "last_sync": "2025-10-25T10:29:00Z",
      "sync_status": "OK"
    }
  },
  "metrics": {
    "workflows_running": 12,
    "workflows_completed_last_hour": 234,
    "workflows_failed_last_hour": 2,
    "average_execution_time_ms": 3500,
    "p95_execution_time_ms": 8000,
    "p99_execution_time_ms": 15000
  }
}
```

**Service Status Values**:
- `up`: Service is healthy
- `degraded`: Service is running but with issues
- `down`: Service is unavailable

---

### 4.2. Get Workflow Metrics

**Endpoint**: `GET /admin/v1/metrics/workflows`

**Description**: Retrieve aggregated workflow metrics.

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:admin`, Role: `DICT_ADMIN`

**Query Parameters**:
- `time_range` (string, optional): Time range for metrics
  - Values: `1h`, `6h`, `24h`, `7d`, `30d` (default: `24h`)
- `workflow_type` (string, optional): Filter by workflow type
- `group_by` (string, optional): Group by field
  - Values: `workflow_type`, `status`, `hour`, `day`

**Example Request**:
```http
GET /admin/v1/metrics/workflows?time_range=24h&group_by=workflow_type
Authorization: Bearer <jwt-token>
```

**Response 200 OK**:
```json
{
  "time_range": "24h",
  "start_time": "2025-10-24T10:30:00Z",
  "end_time": "2025-10-25T10:30:00Z",
  "metrics": [
    {
      "workflow_type": "ENTRY_CREATE",
      "total_executions": 1543,
      "successful": 1520,
      "failed": 15,
      "cancelled": 5,
      "timed_out": 3,
      "success_rate": 98.5,
      "average_duration_ms": 3200,
      "p50_duration_ms": 2800,
      "p95_duration_ms": 7500,
      "p99_duration_ms": 12000,
      "total_retries": 42
    },
    {
      "workflow_type": "CLAIM_CREATE",
      "total_executions": 87,
      "successful": 85,
      "failed": 2,
      "cancelled": 0,
      "timed_out": 0,
      "success_rate": 97.7,
      "average_duration_ms": 5600,
      "p50_duration_ms": 5200,
      "p95_duration_ms": 9800,
      "p99_duration_ms": 15000,
      "total_retries": 8
    }
  ]
}
```

---

### 4.3. Get Queue Statistics

**Endpoint**: `GET /admin/v1/metrics/queues`

**Description**: Retrieve Pulsar queue statistics.

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:admin`, Role: `DICT_ADMIN`

**Example Request**:
```http
GET /admin/v1/metrics/queues
Authorization: Bearer <jwt-token>
```

**Response 200 OK**:
```json
{
  "queues": [
    {
      "topic": "persistent://lb-conn/dict/rsfn-dict-req-out",
      "type": "producer",
      "pending_messages": 5,
      "backlog_size_bytes": 2048,
      "publish_rate_msgs_per_sec": 15.3,
      "publish_throughput_bytes_per_sec": 12800,
      "average_message_size_bytes": 836
    },
    {
      "topic": "persistent://lb-conn/dict/rsfn-dict-res-in",
      "type": "consumer",
      "subscription": "connect-subscription",
      "pending_messages": 0,
      "backlog_size_bytes": 0,
      "consume_rate_msgs_per_sec": 14.8,
      "consume_throughput_bytes_per_sec": 11500,
      "unacked_messages": 2
    },
    {
      "topic": "persistent://lb-conn/dict/rsfn-dict-dlq",
      "type": "dead-letter",
      "pending_messages": 3,
      "backlog_size_bytes": 1536,
      "oldest_message_age_seconds": 3600
    }
  ],
  "timestamp": "2025-10-25T10:30:00Z"
}
```

---

## 5. Maintenance Endpoints

### 5.1. Reprocess Dead Letter Queue

**Endpoint**: `POST /admin/v1/maintenance/reprocess-dlq`

**Description**: Reprocess messages from the Dead Letter Queue.

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:admin`, Role: `DICT_ADMIN`

**Request Body**:
```json
{
  "topic": "persistent://lb-conn/dict/rsfn-dict-dlq",
  "max_messages": 10,
  "filter": {
    "message_age_hours_min": 1,
    "message_age_hours_max": 24
  }
}
```

**Response 200 OK**:
```json
{
  "reprocessed_count": 7,
  "successful_count": 6,
  "failed_count": 1,
  "messages": [
    {
      "message_id": "msg-123",
      "status": "SUCCESS",
      "reprocessed_at": "2025-10-25T10:35:00Z"
    },
    {
      "message_id": "msg-456",
      "status": "FAILED",
      "error": "Validation error: Invalid CPF format",
      "reprocessed_at": "2025-10-25T10:35:01Z"
    }
  ]
}
```

---

### 5.2. Clear Cache

**Endpoint**: `POST /admin/v1/maintenance/clear-cache`

**Description**: Clear Redis cache for specific keys or patterns.

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:admin`, Role: `DICT_ADMIN`

**Request Body**:
```json
{
  "pattern": "entry:*",
  "confirm": true
}
```

**Response 200 OK**:
```json
{
  "keys_deleted": 1543,
  "pattern": "entry:*",
  "timestamp": "2025-10-25T10:40:00Z",
  "message": "Cache cleared successfully"
}
```

---

## 6. Error Handling

### 6.1. Standard Error Response

All error responses follow this format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": [...],
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123",
    "path": "/admin/v1/workflows",
    "method": "GET"
  }
}
```

### 6.2. HTTP Status Codes

| Status Code | Meaning | When to Use |
|-------------|---------|-------------|
| `200 OK` | Success | Successful GET/POST/DELETE |
| `201 Created` | Success | Workflow created |
| `400 Bad Request` | Invalid request | Validation errors |
| `401 Unauthorized` | Missing/invalid auth | No JWT token |
| `403 Forbidden` | Insufficient permissions | Not admin role |
| `404 Not Found` | Resource not found | Workflow not found |
| `409 Conflict` | Resource conflict | Workflow already exists |
| `500 Internal Server Error` | Server error | Unexpected error |
| `502 Bad Gateway` | Upstream error | Temporal unavailable |
| `503 Service Unavailable` | Service down | Maintenance mode |

### 6.3. Error Codes

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Missing or invalid JWT |
| `FORBIDDEN` | 403 | Admin role required |
| `INVALID_QUERY_PARAMS` | 400 | Invalid query parameters |
| `WORKFLOW_NOT_FOUND` | 404 | Workflow not found |
| `WORKFLOW_ALREADY_EXISTS` | 409 | Workflow already exists |
| `WORKFLOW_ALREADY_COMPLETED` | 409 | Cannot cancel completed workflow |
| `TEMPORAL_CONNECTION_ERROR` | 502 | Cannot connect to Temporal |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Rastreabilidade

### Requisitos Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RF-ADMIN-001 | Listar workflows Temporal | [TEC-003](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md) | ✅ Especificado |
| RF-ADMIN-002 | Obter detalhes de workflow | [TEC-003](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md) | ✅ Especificado |
| RF-ADMIN-003 | Cancelar workflow em execução | [TEC-003](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md) | ✅ Especificado |
| RF-ADMIN-004 | Reprocessar Dead Letter Queue | [TEC-003](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md) | ✅ Especificado |

### Requisitos Não-Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RNF-ADMIN-001 | Apenas admin pode acessar | SEC-001 (pendente) | ✅ Especificado |
| RNF-ADMIN-002 | Rate limiting (500 req/min) | ADR-005 (pendente) | ✅ Especificado |
| RNF-ADMIN-003 | Timeout de 60s | Best Practices | ✅ Especificado |

---

## Próximas Revisões

**Pendências**:
- [ ] Adicionar endpoint para visualizar logs de workflow
- [ ] Implementar filtros avançados para workflows
- [ ] Adicionar suporte para export de métricas (CSV, JSON)
- [ ] Implementar webhooks para alertas de workflows falhados
- [ ] Adicionar endpoint para forçar retry de activity específica

---

**Referências**:
- [TEC-003: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [API-002: Core DICT REST API](./API-002_Core_DICT_REST_API.md)
- [Temporal Documentation](https://docs.temporal.io/)
