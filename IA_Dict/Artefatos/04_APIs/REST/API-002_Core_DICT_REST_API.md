# API-002: Core DICT REST API Specification

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Serviço**: Core DICT REST API
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: BACKEND (AI Agent - Backend Specialist)
**Revisor**: [Aguardando]
**Aprovador**: Tech Lead, Head de Arquitetura

---

## Sumário Executivo

Este documento especifica a **API REST** do **Core DICT**, expondo endpoints HTTP para gerenciamento de chaves PIX, reivindicações e portabilidades. A API segue os princípios RESTful e está documentada em formato OpenAPI 3.0.

**Baseado em**:
- [TEC-001: Core DICT Specification](../../11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md)
- [GRPC-001: Bridge gRPC Service](../gRPC/GRPC-001_Bridge_gRPC_Service.md)
- [TEC-003: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | BACKEND | Versão inicial - Core DICT REST API |

---

## Índice

1. [Visão Geral](#1-visão-geral)
2. [Authentication & Authorization](#2-authentication--authorization)
3. [Entry Management Endpoints](#3-entry-management-endpoints)
4. [Claim Management Endpoints](#4-claim-management-endpoints)
5. [Portability Management Endpoints](#5-portability-management-endpoints)
6. [Error Handling](#6-error-handling)
7. [Rate Limiting](#7-rate-limiting)
8. [OpenAPI Specification](#8-openapi-specification)

---

## 1. Visão Geral

### 1.1. Base URL

```
Production:  https://api.lbpay.com.br/dict/v1
Staging:     https://api-stg.lbpay.com.br/dict/v1
Development: http://localhost:8080/api/v1
```

### 1.2. Características

| Aspecto | Valor |
|---------|-------|
| **Protocol** | HTTPS (TLS 1.2+) |
| **Content-Type** | application/json |
| **Authentication** | JWT Bearer Token |
| **Authorization** | RBAC (Role-Based Access Control) |
| **API Version** | v1 (URL-based versioning) |
| **Rate Limit** | 1000 req/min per user |
| **Timeout** | 30s |

### 1.3. HTTP Methods

| Method | Usage | Idempotent |
|--------|-------|------------|
| `GET` | Retrieve resources | ✅ Yes |
| `POST` | Create resources | ❌ No |
| `PUT` | Replace resources | ✅ Yes |
| `PATCH` | Partial update | ❌ No |
| `DELETE` | Remove resources | ✅ Yes |

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
  "sub": "user-uuid",
  "email": "user@example.com",
  "roles": ["DICT_USER", "DICT_ADMIN"],
  "scopes": ["dict:read", "dict:write", "dict:admin"],
  "iss": "lbpay-auth",
  "exp": 1698350400,
  "iat": 1698264000
}
```

### 2.2. Authorization (RBAC)

**Roles**:

| Role | Permissions | Scopes |
|------|-------------|--------|
| `DICT_USER` | Read own entries, create claims | `dict:read`, `dict:write` |
| `DICT_ADMIN` | Full access | `dict:read`, `dict:write`, `dict:admin` |
| `DICT_AUDITOR` | Read-only access to all | `dict:read` |

**Endpoint Permissions**:

| Endpoint | Required Scope | Required Role |
|----------|---------------|---------------|
| `GET /api/v1/keys` | `dict:read` | `DICT_USER` |
| `POST /api/v1/keys` | `dict:write` | `DICT_USER` |
| `DELETE /api/v1/keys/:id` | `dict:write` | `DICT_USER` (owner only) |
| `POST /api/v1/claims` | `dict:write` | `DICT_USER` |
| `GET /api/v1/claims` | `dict:read` | `DICT_USER` |
| `POST /api/v1/portabilities` | `dict:write` | `DICT_USER` |

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
    "message": "Insufficient permissions. Required scope: dict:write",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

---

## 3. Entry Management Endpoints

### 3.1. Create Entry (PIX Key)

**Endpoint**: `POST /api/v1/keys`

**Description**: Create a new PIX key entry in the DICT.

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:write`

**Request Body**:
```json
{
  "key_type": "CPF",
  "key_value": "12345678901",
  "account": {
    "account_number": "12345678",
    "branch_code": "0001",
    "account_type": "CACC",
    "holder_document": "12345678901",
    "holder_name": "João Silva",
    "participant_ispb": "12345678"
  },
  "idempotency_key": "uuid-v4"
}
```

**Request Schema**:
```json
{
  "key_type": {
    "type": "string",
    "enum": ["CPF", "CNPJ", "EMAIL", "PHONE", "EVP"],
    "required": true
  },
  "key_value": {
    "type": "string",
    "required": true,
    "description": "CPF: 11 digits, CNPJ: 14 digits, EMAIL: valid email, PHONE: +5511999999999, EVP: UUID"
  },
  "account": {
    "type": "object",
    "required": true
  },
  "idempotency_key": {
    "type": "string",
    "format": "uuid",
    "required": true
  }
}
```

**Response 201 Created**:
```json
{
  "entry_id": "uuid-v4",
  "external_id": "bacen-dict-id",
  "key_type": "CPF",
  "key_value": "12345678901",
  "status": "ACTIVE",
  "created_at": "2025-10-25T10:30:00Z",
  "updated_at": "2025-10-25T10:30:00Z"
}
```

**Error Responses**:

**400 Bad Request** (Invalid payload):
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid CPF format",
    "details": [
      {
        "field": "key_value",
        "message": "CPF must have 11 digits"
      }
    ],
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

**409 Conflict** (Key already exists):
```json
{
  "error": {
    "code": "KEY_ALREADY_EXISTS",
    "message": "PIX key already registered",
    "existing_entry_id": "uuid-v4",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

**422 Unprocessable Entity** (Business rule violation):
```json
{
  "error": {
    "code": "MAX_KEYS_EXCEEDED",
    "message": "User has reached maximum number of PIX keys (20)",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

**500 Internal Server Error** (Bacen unavailable):
```json
{
  "error": {
    "code": "EXTERNAL_SERVICE_ERROR",
    "message": "Bacen DICT is temporarily unavailable. Please try again later.",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

---

### 3.2. Get Entry (Query PIX Key)

**Endpoint**: `GET /api/v1/keys/:key`

**Description**: Retrieve a PIX key entry by key value.

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:read`

**Path Parameters**:
- `key` (string): PIX key value (CPF, CNPJ, email, phone, EVP)

**Query Parameters**:
- `key_type` (string, optional): Filter by key type (CPF, CNPJ, EMAIL, PHONE, EVP)

**Example Request**:
```http
GET /api/v1/keys/12345678901?key_type=CPF
Authorization: Bearer <jwt-token>
```

**Response 200 OK**:
```json
{
  "entry_id": "uuid-v4",
  "external_id": "bacen-dict-id",
  "key_type": "CPF",
  "key_value": "12345678901",
  "account": {
    "account_number": "12345678",
    "branch_code": "0001",
    "account_type": "CACC",
    "holder_document": "12345678901",
    "holder_name": "João Silva",
    "participant_ispb": "12345678"
  },
  "status": "ACTIVE",
  "created_at": "2025-10-25T10:30:00Z",
  "updated_at": "2025-10-25T10:30:00Z"
}
```

**Error Responses**:

**404 Not Found**:
```json
{
  "error": {
    "code": "KEY_NOT_FOUND",
    "message": "PIX key not found",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

---

### 3.3. List User Entries

**Endpoint**: `GET /api/v1/keys`

**Description**: List all PIX keys for the authenticated user.

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:read`

**Query Parameters**:
- `page` (integer, default: 1): Page number
- `limit` (integer, default: 20, max: 100): Results per page
- `status` (string, optional): Filter by status (ACTIVE, PENDING, DELETED)

**Example Request**:
```http
GET /api/v1/keys?page=1&limit=20&status=ACTIVE
Authorization: Bearer <jwt-token>
```

**Response 200 OK**:
```json
{
  "entries": [
    {
      "entry_id": "uuid-v4-1",
      "key_type": "CPF",
      "key_value": "12345678901",
      "status": "ACTIVE",
      "created_at": "2025-10-25T10:30:00Z"
    },
    {
      "entry_id": "uuid-v4-2",
      "key_type": "EMAIL",
      "key_value": "user@example.com",
      "status": "ACTIVE",
      "created_at": "2025-10-24T09:15:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 3,
    "total_entries": 45,
    "per_page": 20
  }
}
```

---

### 3.4. Delete Entry (Remove PIX Key)

**Endpoint**: `DELETE /api/v1/keys/:id`

**Description**: Delete a PIX key (soft delete).

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:write` + Entry owner

**Path Parameters**:
- `id` (uuid): Entry ID

**Example Request**:
```http
DELETE /api/v1/keys/uuid-v4
Authorization: Bearer <jwt-token>
```

**Response 204 No Content**:
```
(Empty body)
```

**Error Responses**:

**403 Forbidden** (Not owner):
```json
{
  "error": {
    "code": "NOT_ENTRY_OWNER",
    "message": "You can only delete your own PIX keys",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

**404 Not Found**:
```json
{
  "error": {
    "code": "ENTRY_NOT_FOUND",
    "message": "Entry not found",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

**409 Conflict** (Pending claim):
```json
{
  "error": {
    "code": "CLAIM_PENDING",
    "message": "Cannot delete entry with pending claim",
    "claim_id": "uuid-v4",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

---

## 4. Claim Management Endpoints

### 4.1. Create Claim (Initiate 30-day process)

**Endpoint**: `POST /api/v1/claims`

**Description**: Initiate a claim for a PIX key ownership. Starts a 30-day resolution period.

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:write`

**Request Body**:
```json
{
  "entry_key": "12345678901",
  "key_type": "CPF",
  "claimer_account": {
    "account_number": "87654321",
    "branch_code": "0002",
    "account_type": "CACC",
    "holder_document": "12345678901",
    "holder_name": "João Silva",
    "participant_ispb": "87654321"
  },
  "claim_reason": "OWNERSHIP",
  "idempotency_key": "uuid-v4"
}
```

**Request Schema**:
```json
{
  "entry_key": {
    "type": "string",
    "required": true
  },
  "key_type": {
    "type": "string",
    "enum": ["CPF", "CNPJ", "EMAIL", "PHONE", "EVP"],
    "required": true
  },
  "claimer_account": {
    "type": "object",
    "required": true
  },
  "claim_reason": {
    "type": "string",
    "enum": ["OWNERSHIP", "FRAUD"],
    "required": true
  },
  "idempotency_key": {
    "type": "string",
    "format": "uuid",
    "required": true
  }
}
```

**Response 201 Created**:
```json
{
  "claim_id": "uuid-v4",
  "external_id": "bacen-claim-id",
  "entry_key": "12345678901",
  "status": "OPEN",
  "completion_period_days": 30,
  "created_at": "2025-10-25T10:30:00Z",
  "expires_at": "2025-11-24T10:30:00Z",
  "claimer_ispb": "87654321",
  "owner_ispb": "12345678"
}
```

**Error Responses**:

**400 Bad Request** (Invalid claim):
```json
{
  "error": {
    "code": "INVALID_CLAIM",
    "message": "Cannot claim your own PIX key",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

**409 Conflict** (Claim already exists):
```json
{
  "error": {
    "code": "CLAIM_ALREADY_EXISTS",
    "message": "An active claim already exists for this PIX key",
    "existing_claim_id": "uuid-v4",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

---

### 4.2. Get Claim Status

**Endpoint**: `GET /api/v1/claims/:id`

**Description**: Retrieve claim status and details.

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:read`

**Path Parameters**:
- `id` (uuid): Claim ID

**Example Request**:
```http
GET /api/v1/claims/uuid-v4
Authorization: Bearer <jwt-token>
```

**Response 200 OK**:
```json
{
  "claim_id": "uuid-v4",
  "external_id": "bacen-claim-id",
  "entry_key": "12345678901",
  "status": "WAITING_RESOLUTION",
  "completion_period_days": 30,
  "created_at": "2025-10-25T10:30:00Z",
  "expires_at": "2025-11-24T10:30:00Z",
  "days_remaining": 25,
  "claimer_ispb": "87654321",
  "owner_ispb": "12345678",
  "resolution": null
}
```

**Status Values**:
- `OPEN`: Claim initiated
- `WAITING_RESOLUTION`: Waiting for owner response
- `CONFIRMED`: Owner confirmed
- `CANCELLED`: Claim cancelled
- `COMPLETED`: Ownership transferred
- `EXPIRED`: 30-day period expired

**Error Responses**:

**404 Not Found**:
```json
{
  "error": {
    "code": "CLAIM_NOT_FOUND",
    "message": "Claim not found",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

---

### 4.3. List User Claims

**Endpoint**: `GET /api/v1/claims`

**Description**: List all claims for the authenticated user.

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:read`

**Query Parameters**:
- `page` (integer, default: 1)
- `limit` (integer, default: 20, max: 100)
- `status` (string, optional): Filter by status
- `role` (string, optional): Filter by role (`CLAIMER` or `OWNER`)

**Example Request**:
```http
GET /api/v1/claims?page=1&limit=20&status=OPEN&role=CLAIMER
Authorization: Bearer <jwt-token>
```

**Response 200 OK**:
```json
{
  "claims": [
    {
      "claim_id": "uuid-v4-1",
      "entry_key": "12345678901",
      "status": "OPEN",
      "role": "CLAIMER",
      "days_remaining": 25,
      "created_at": "2025-10-25T10:30:00Z",
      "expires_at": "2025-11-24T10:30:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 1,
    "total_claims": 1,
    "per_page": 20
  }
}
```

---

### 4.4. Confirm Claim (Owner accepts)

**Endpoint**: `PUT /api/v1/claims/:id/confirm`

**Description**: Confirm claim (owner accepts ownership transfer).

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:write` + Claim owner

**Path Parameters**:
- `id` (uuid): Claim ID

**Request Body**:
```json
{
  "confirmation_reason": "Transferring to new account"
}
```

**Example Request**:
```http
PUT /api/v1/claims/uuid-v4/confirm
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "confirmation_reason": "Transferring to new account"
}
```

**Response 200 OK**:
```json
{
  "claim_id": "uuid-v4",
  "status": "CONFIRMED",
  "confirmed_at": "2025-10-25T10:35:00Z",
  "message": "Claim confirmed. Ownership will be transferred within 24 hours."
}
```

**Error Responses**:

**403 Forbidden** (Not claim owner):
```json
{
  "error": {
    "code": "NOT_CLAIM_OWNER",
    "message": "Only the current PIX key owner can confirm claims",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

**409 Conflict** (Already resolved):
```json
{
  "error": {
    "code": "CLAIM_ALREADY_RESOLVED",
    "message": "Claim has already been resolved",
    "current_status": "COMPLETED",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

---

### 4.5. Cancel Claim

**Endpoint**: `DELETE /api/v1/claims/:id`

**Description**: Cancel a claim (claimer or owner can cancel).

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:write` + (Claimer OR Owner)

**Path Parameters**:
- `id` (uuid): Claim ID

**Query Parameters**:
- `reason` (string, optional): Cancellation reason

**Example Request**:
```http
DELETE /api/v1/claims/uuid-v4?reason=User+requested+cancellation
Authorization: Bearer <jwt-token>
```

**Response 200 OK**:
```json
{
  "claim_id": "uuid-v4",
  "status": "CANCELLED",
  "cancelled_at": "2025-10-25T10:40:00Z",
  "cancellation_reason": "User requested cancellation"
}
```

---

## 5. Portability Management Endpoints

### 5.1. Initiate Portability

**Endpoint**: `POST /api/v1/portabilities`

**Description**: Initiate portability of PIX key to another financial institution.

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:write`

**Request Body**:
```json
{
  "entry_key": "12345678901",
  "key_type": "CPF",
  "new_account": {
    "account_number": "11111111",
    "branch_code": "0003",
    "account_type": "CACC",
    "holder_document": "12345678901",
    "holder_name": "João Silva",
    "participant_ispb": "11111111"
  },
  "idempotency_key": "uuid-v4"
}
```

**Response 201 Created**:
```json
{
  "portability_id": "uuid-v4",
  "external_id": "bacen-portability-id",
  "entry_key": "12345678901",
  "status": "PENDING",
  "from_ispb": "12345678",
  "to_ispb": "11111111",
  "created_at": "2025-10-25T10:30:00Z"
}
```

**Error Responses**:

**400 Bad Request** (Invalid portability):
```json
{
  "error": {
    "code": "INVALID_PORTABILITY",
    "message": "Cannot port PIX key to the same institution",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

**409 Conflict** (Portability already in progress):
```json
{
  "error": {
    "code": "PORTABILITY_IN_PROGRESS",
    "message": "A portability process is already in progress for this PIX key",
    "existing_portability_id": "uuid-v4",
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

---

### 5.2. Get Portability Status

**Endpoint**: `GET /api/v1/portabilities/:id`

**Description**: Retrieve portability status.

**Authentication**: Required (JWT)

**Authorization**: Scope: `dict:read`

**Path Parameters**:
- `id` (uuid): Portability ID

**Example Request**:
```http
GET /api/v1/portabilities/uuid-v4
Authorization: Bearer <jwt-token>
```

**Response 200 OK**:
```json
{
  "portability_id": "uuid-v4",
  "external_id": "bacen-portability-id",
  "entry_key": "12345678901",
  "status": "COMPLETED",
  "from_ispb": "12345678",
  "to_ispb": "11111111",
  "created_at": "2025-10-25T10:30:00Z",
  "completed_at": "2025-10-25T12:00:00Z"
}
```

**Status Values**:
- `PENDING`: Portability initiated
- `CONFIRMED`: Confirmed by both institutions
- `COMPLETED`: Portability completed
- `CANCELLED`: Portability cancelled
- `REJECTED`: Portability rejected

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
    "path": "/api/v1/keys",
    "method": "POST"
  }
}
```

### 6.2. HTTP Status Codes

| Status Code | Meaning | When to Use |
|-------------|---------|-------------|
| `200 OK` | Success (GET, PUT) | Resource retrieved or updated |
| `201 Created` | Success (POST) | Resource created |
| `204 No Content` | Success (DELETE) | Resource deleted |
| `400 Bad Request` | Invalid request payload | Validation errors |
| `401 Unauthorized` | Missing/invalid auth | No JWT token |
| `403 Forbidden` | Insufficient permissions | Wrong scope or role |
| `404 Not Found` | Resource not found | Entry/Claim not found |
| `409 Conflict` | Resource conflict | Key/Claim already exists |
| `422 Unprocessable Entity` | Business rule violation | Max keys exceeded |
| `429 Too Many Requests` | Rate limit exceeded | > 1000 req/min |
| `500 Internal Server Error` | Server error | Unexpected error |
| `502 Bad Gateway` | Upstream error | Bacen DICT unavailable |
| `503 Service Unavailable` | Service down | Maintenance mode |

### 6.3. Error Codes

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Invalid request payload |
| `UNAUTHORIZED` | 401 | Missing or invalid JWT |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `KEY_NOT_FOUND` | 404 | PIX key not found |
| `ENTRY_NOT_FOUND` | 404 | Entry not found |
| `CLAIM_NOT_FOUND` | 404 | Claim not found |
| `KEY_ALREADY_EXISTS` | 409 | PIX key already registered |
| `CLAIM_ALREADY_EXISTS` | 409 | Claim already exists |
| `PORTABILITY_IN_PROGRESS` | 409 | Portability already in progress |
| `MAX_KEYS_EXCEEDED` | 422 | User has too many keys |
| `INVALID_CLAIM` | 400 | Cannot claim own key |
| `NOT_ENTRY_OWNER` | 403 | Not the entry owner |
| `NOT_CLAIM_OWNER` | 403 | Not the claim owner |
| `CLAIM_PENDING` | 409 | Cannot delete entry with pending claim |
| `CLAIM_ALREADY_RESOLVED` | 409 | Claim already resolved |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `EXTERNAL_SERVICE_ERROR` | 502 | Bacen DICT error |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## 7. Rate Limiting

### 7.1. Rate Limit Policy

**Default Limits**:
- **Per User**: 1000 requests/minute
- **Per IP**: 5000 requests/minute
- **Per Endpoint**: 500 requests/minute

**Response Headers**:
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1698264060
```

### 7.2. Rate Limit Exceeded Response

**429 Too Many Requests**:
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Try again in 60 seconds.",
    "retry_after": 60,
    "timestamp": "2025-10-25T10:30:00Z",
    "trace_id": "abc123"
  }
}
```

**Retry-After Header**:
```http
Retry-After: 60
```

---

## 8. OpenAPI Specification

### 8.1. OpenAPI 3.0 Definition

```yaml
openapi: 3.0.3
info:
  title: Core DICT REST API
  description: API for managing PIX keys (DICT entries), claims, and portabilities
  version: 1.0.0
  contact:
    name: LBPay Tech Team
    email: tech@lbpay.com.br

servers:
  - url: https://api.lbpay.com.br/dict/v1
    description: Production
  - url: https://api-stg.lbpay.com.br/dict/v1
    description: Staging
  - url: http://localhost:8080/api/v1
    description: Development

security:
  - BearerAuth: []

components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

  schemas:
    Entry:
      type: object
      required:
        - entry_id
        - key_type
        - key_value
        - status
      properties:
        entry_id:
          type: string
          format: uuid
        external_id:
          type: string
          description: Bacen DICT ID
        key_type:
          type: string
          enum: [CPF, CNPJ, EMAIL, PHONE, EVP]
        key_value:
          type: string
        account:
          $ref: '#/components/schemas/Account'
        status:
          type: string
          enum: [ACTIVE, PENDING, DELETED, CLAIM_PENDING]
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time

    Account:
      type: object
      required:
        - account_number
        - branch_code
        - account_type
        - holder_document
        - holder_name
        - participant_ispb
      properties:
        account_number:
          type: string
        branch_code:
          type: string
        account_type:
          type: string
          enum: [CACC, SVGS, SLRY, TRAN]
        holder_document:
          type: string
        holder_name:
          type: string
        participant_ispb:
          type: string

    Claim:
      type: object
      required:
        - claim_id
        - entry_key
        - status
      properties:
        claim_id:
          type: string
          format: uuid
        external_id:
          type: string
        entry_key:
          type: string
        status:
          type: string
          enum: [OPEN, WAITING_RESOLUTION, CONFIRMED, CANCELLED, COMPLETED, EXPIRED]
        completion_period_days:
          type: integer
          default: 30
        created_at:
          type: string
          format: date-time
        expires_at:
          type: string
          format: date-time
        days_remaining:
          type: integer
        claimer_ispb:
          type: string
        owner_ispb:
          type: string

    Portability:
      type: object
      required:
        - portability_id
        - entry_key
        - status
      properties:
        portability_id:
          type: string
          format: uuid
        external_id:
          type: string
        entry_key:
          type: string
        status:
          type: string
          enum: [PENDING, CONFIRMED, COMPLETED, CANCELLED, REJECTED]
        from_ispb:
          type: string
        to_ispb:
          type: string
        created_at:
          type: string
          format: date-time
        completed_at:
          type: string
          format: date-time

    Error:
      type: object
      required:
        - code
        - message
        - timestamp
      properties:
        code:
          type: string
        message:
          type: string
        details:
          type: array
          items:
            type: object
        timestamp:
          type: string
          format: date-time
        trace_id:
          type: string
        path:
          type: string
        method:
          type: string

paths:
  /keys:
    post:
      summary: Create PIX Key
      operationId: createEntry
      tags:
        - Entry Management
      security:
        - BearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - key_type
                - key_value
                - account
                - idempotency_key
              properties:
                key_type:
                  type: string
                  enum: [CPF, CNPJ, EMAIL, PHONE, EVP]
                key_value:
                  type: string
                account:
                  $ref: '#/components/schemas/Account'
                idempotency_key:
                  type: string
                  format: uuid
      responses:
        '201':
          description: Entry created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Entry'
        '400':
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '401':
          description: Unauthorized
        '409':
          description: Key already exists
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

    get:
      summary: List User Entries
      operationId: listEntries
      tags:
        - Entry Management
      security:
        - BearerAuth: []
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            default: 1
        - name: limit
          in: query
          schema:
            type: integer
            default: 20
            maximum: 100
        - name: status
          in: query
          schema:
            type: string
            enum: [ACTIVE, PENDING, DELETED]
      responses:
        '200':
          description: List of entries
          content:
            application/json:
              schema:
                type: object
                properties:
                  entries:
                    type: array
                    items:
                      $ref: '#/components/schemas/Entry'
                  pagination:
                    type: object
                    properties:
                      current_page:
                        type: integer
                      total_pages:
                        type: integer
                      total_entries:
                        type: integer
                      per_page:
                        type: integer

  /keys/{key}:
    get:
      summary: Get Entry by Key
      operationId: getEntry
      tags:
        - Entry Management
      security:
        - BearerAuth: []
      parameters:
        - name: key
          in: path
          required: true
          schema:
            type: string
        - name: key_type
          in: query
          schema:
            type: string
            enum: [CPF, CNPJ, EMAIL, PHONE, EVP]
      responses:
        '200':
          description: Entry found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Entry'
        '404':
          description: Entry not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /keys/{id}:
    delete:
      summary: Delete Entry
      operationId: deleteEntry
      tags:
        - Entry Management
      security:
        - BearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '204':
          description: Entry deleted successfully
        '403':
          description: Forbidden (not owner)
        '404':
          description: Entry not found
        '409':
          description: Conflict (claim pending)

  /claims:
    post:
      summary: Create Claim
      operationId: createClaim
      tags:
        - Claim Management
      security:
        - BearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - entry_key
                - key_type
                - claimer_account
                - claim_reason
                - idempotency_key
              properties:
                entry_key:
                  type: string
                key_type:
                  type: string
                  enum: [CPF, CNPJ, EMAIL, PHONE, EVP]
                claimer_account:
                  $ref: '#/components/schemas/Account'
                claim_reason:
                  type: string
                  enum: [OWNERSHIP, FRAUD]
                idempotency_key:
                  type: string
                  format: uuid
      responses:
        '201':
          description: Claim created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Claim'
        '400':
          description: Bad request
        '409':
          description: Claim already exists

    get:
      summary: List User Claims
      operationId: listClaims
      tags:
        - Claim Management
      security:
        - BearerAuth: []
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            default: 1
        - name: limit
          in: query
          schema:
            type: integer
            default: 20
            maximum: 100
        - name: status
          in: query
          schema:
            type: string
            enum: [OPEN, WAITING_RESOLUTION, CONFIRMED, CANCELLED, COMPLETED, EXPIRED]
        - name: role
          in: query
          schema:
            type: string
            enum: [CLAIMER, OWNER]
      responses:
        '200':
          description: List of claims
          content:
            application/json:
              schema:
                type: object
                properties:
                  claims:
                    type: array
                    items:
                      $ref: '#/components/schemas/Claim'
                  pagination:
                    type: object

  /claims/{id}:
    get:
      summary: Get Claim by ID
      operationId: getClaim
      tags:
        - Claim Management
      security:
        - BearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Claim found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Claim'
        '404':
          description: Claim not found

    delete:
      summary: Cancel Claim
      operationId: cancelClaim
      tags:
        - Claim Management
      security:
        - BearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
        - name: reason
          in: query
          schema:
            type: string
      responses:
        '200':
          description: Claim cancelled
          content:
            application/json:
              schema:
                type: object
                properties:
                  claim_id:
                    type: string
                    format: uuid
                  status:
                    type: string
                  cancelled_at:
                    type: string
                    format: date-time
                  cancellation_reason:
                    type: string

  /claims/{id}/confirm:
    put:
      summary: Confirm Claim (Owner accepts)
      operationId: confirmClaim
      tags:
        - Claim Management
      security:
        - BearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                confirmation_reason:
                  type: string
      responses:
        '200':
          description: Claim confirmed
          content:
            application/json:
              schema:
                type: object
                properties:
                  claim_id:
                    type: string
                    format: uuid
                  status:
                    type: string
                  confirmed_at:
                    type: string
                    format: date-time
                  message:
                    type: string

  /portabilities:
    post:
      summary: Initiate Portability
      operationId: initiatePortability
      tags:
        - Portability Management
      security:
        - BearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - entry_key
                - key_type
                - new_account
                - idempotency_key
              properties:
                entry_key:
                  type: string
                key_type:
                  type: string
                  enum: [CPF, CNPJ, EMAIL, PHONE, EVP]
                new_account:
                  $ref: '#/components/schemas/Account'
                idempotency_key:
                  type: string
                  format: uuid
      responses:
        '201':
          description: Portability initiated
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Portability'

  /portabilities/{id}:
    get:
      summary: Get Portability Status
      operationId: getPortability
      tags:
        - Portability Management
      security:
        - BearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Portability found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Portability'
```

---

## Rastreabilidade

### Requisitos Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RF-API-001 | Criar chave PIX via REST | [TEC-001](../../11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md) | ✅ Especificado |
| RF-API-002 | Consultar chave PIX via REST | [TEC-001](../../11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md) | ✅ Especificado |
| RF-API-003 | Deletar chave PIX via REST | [TEC-001](../../11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md) | ✅ Especificado |
| RF-API-004 | Criar reivindicação (30 dias) via REST | [TEC-003](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md) | ✅ Especificado |
| RF-API-005 | Confirmar/Cancelar reivindicação via REST | [TEC-003](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md) | ✅ Especificado |
| RF-API-006 | Iniciar portabilidade via REST | [TEC-001](../../11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md) | ✅ Especificado |

### Requisitos Não-Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RNF-API-001 | Autenticação JWT | SEC-001 (pendente) | ✅ Especificado |
| RNF-API-002 | Autorização RBAC | SEC-001 (pendente) | ✅ Especificado |
| RNF-API-003 | Rate limiting (1000 req/min) | ADR-005 (pendente) | ✅ Especificado |
| RNF-API-004 | OpenAPI 3.0 documentation | Best Practices | ✅ Especificado |
| RNF-API-005 | Idempotência em POST | ADR-005 (pendente) | ✅ Especificado |

---

## Próximas Revisões

**Pendências**:
- [ ] Implementar autenticação OAuth 2.0 (além de JWT)
- [ ] Definir webhooks para notificações assíncronas
- [ ] Adicionar endpoints para VSYNC (quando implementado)
- [ ] Validar rate limits em ambiente real
- [ ] Adicionar suporte para API versioning via headers
- [ ] Implementar HATEOAS (links para recursos relacionados)

---

**Referências**:
- [TEC-001: Core DICT Specification](../../11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md)
- [TEC-003: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [GRPC-001: Bridge gRPC Service](../gRPC/GRPC-001_Bridge_gRPC_Service.md)
- [OpenAPI 3.0 Specification](https://swagger.io/specification/)
- [REST API Best Practices](https://restfulapi.net/)
