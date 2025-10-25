# API-004: Complete OpenAPI 3.0 Specifications

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Serviço**: Complete OpenAPI Specification
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: BACKEND (AI Agent - Backend Specialist)
**Revisor**: [Aguardando]
**Aprovador**: Tech Lead, Head de Arquitetura

---

## Sumário Executivo

Este documento fornece a especificação completa em **OpenAPI 3.0** de todas as APIs do projeto DICT, combinando:
- **API-002**: Core DICT REST API (Entry/Claim/Portability Management)
- **API-003**: Connect Admin API (Workflow Management)

A especificação pode ser importada em ferramentas como Swagger UI, Postman, Insomnia, ou usada para geração de código cliente/servidor.

**Baseado em**:
- [API-002: Core DICT REST API](./API-002_Core_DICT_REST_API.md)
- [API-003: Connect Admin API](./API-003_Connect_Admin_API.md)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | BACKEND | Versão inicial - Complete OpenAPI 3.0 |

---

## Índice

1. [Visão Geral](#1-visão-geral)
2. [OpenAPI 3.0 Specification (YAML)](#2-openapi-30-specification-yaml)
3. [Usage Instructions](#3-usage-instructions)

---

## 1. Visão Geral

### 1.1. Servers

| Environment | Base URL | Description |
|-------------|----------|-------------|
| **Production** | `https://api.lbpay.com.br` | Production environment |
| **Staging** | `https://api-stg.lbpay.com.br` | Staging/QA environment |
| **Development** | `http://localhost:8080` | Local development |

### 1.2. APIs Incluídas

| API | Path Prefix | Description |
|-----|-------------|-------------|
| **Core DICT** | `/dict/v1` | Entry, Claim, Portability management |
| **Connect Admin** | `/connect/admin/v1` | Workflow management and monitoring |

---

## 2. OpenAPI 3.0 Specification (YAML)

```yaml
openapi: 3.0.3
info:
  title: LBPay DICT - Complete API Specification
  description: |
    Complete API specification for LBPay DICT project, including:
    - Core DICT REST API: PIX key management, claims, portabilities
    - Connect Admin API: Temporal workflow management and monitoring

    ## Authentication
    All endpoints require JWT Bearer token authentication with appropriate roles and scopes.

    ## Rate Limiting
    - Core DICT API: 1000 requests/minute per user
    - Connect Admin API: 500 requests/minute per admin

    ## Support
    For support, contact: tech@lbpay.com.br
  version: 1.0.0
  contact:
    name: LBPay Tech Team
    email: tech@lbpay.com.br
    url: https://lbpay.com.br
  license:
    name: Proprietary
    url: https://lbpay.com.br/license

servers:
  - url: https://api.lbpay.com.br
    description: Production
  - url: https://api-stg.lbpay.com.br
    description: Staging
  - url: http://localhost:8080
    description: Development

tags:
  - name: Entry Management
    description: PIX key entry operations (Core DICT)
  - name: Claim Management
    description: Claim operations for ownership disputes (Core DICT)
  - name: Portability Management
    description: PIX key portability between institutions (Core DICT)
  - name: Workflow Management
    description: Temporal workflow operations (Connect Admin)
  - name: Monitoring
    description: System health and metrics (Connect Admin)
  - name: Maintenance
    description: Maintenance operations (Connect Admin)

security:
  - BearerAuth: []

components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      description: |
        JWT Bearer token authentication. Include the token in the Authorization header:
        `Authorization: Bearer <token>`

        Token payload should contain:
        - `sub`: User UUID
        - `roles`: Array of roles (DICT_USER, DICT_ADMIN, etc.)
        - `scopes`: Array of scopes (dict:read, dict:write, dict:admin)

  schemas:
    # ==================== Core DICT Schemas ====================

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
          description: Account number
          example: "12345678"
        branch_code:
          type: string
          description: Branch code
          example: "0001"
        account_type:
          type: string
          description: Account type
          enum: [CACC, SVGS, SLRY, TRAN]
          example: CACC
        holder_document:
          type: string
          description: Account holder CPF or CNPJ
          example: "12345678901"
        holder_name:
          type: string
          description: Account holder name
          example: "João Silva"
        participant_ispb:
          type: string
          description: Financial institution ISPB
          pattern: '^\d{8}$'
          example: "12345678"

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
          description: Internal entry UUID
          example: "550e8400-e29b-41d4-a716-446655440000"
        external_id:
          type: string
          description: Bacen DICT external ID
          example: "bacen-dict-id-123"
        key_type:
          type: string
          description: PIX key type
          enum: [CPF, CNPJ, EMAIL, PHONE, EVP]
          example: CPF
        key_value:
          type: string
          description: PIX key value
          example: "12345678901"
        account:
          $ref: '#/components/schemas/Account'
        status:
          type: string
          description: Entry status
          enum: [ACTIVE, PENDING, DELETED, CLAIM_PENDING]
          example: ACTIVE
        created_at:
          type: string
          format: date-time
          description: Creation timestamp
          example: "2025-10-25T10:30:00Z"
        updated_at:
          type: string
          format: date-time
          description: Last update timestamp
          example: "2025-10-25T10:30:00Z"

    CreateEntryRequest:
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
          description: PIX key type
          example: CPF
        key_value:
          type: string
          description: |
            PIX key value:
            - CPF: 11 digits
            - CNPJ: 14 digits
            - EMAIL: valid email
            - PHONE: +5511999999999
            - EVP: UUID
          example: "12345678901"
        account:
          $ref: '#/components/schemas/Account'
        idempotency_key:
          type: string
          format: uuid
          description: Idempotency key for duplicate prevention
          example: "550e8400-e29b-41d4-a716-446655440001"

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
          description: Internal claim UUID
          example: "660e8400-e29b-41d4-a716-446655440000"
        external_id:
          type: string
          description: Bacen DICT claim external ID
          example: "bacen-claim-id-456"
        entry_key:
          type: string
          description: PIX key being claimed
          example: "12345678901"
        status:
          type: string
          description: Claim status
          enum: [OPEN, WAITING_RESOLUTION, CONFIRMED, CANCELLED, COMPLETED, EXPIRED]
          example: OPEN
        completion_period_days:
          type: integer
          description: Resolution period in days
          default: 30
          example: 30
        created_at:
          type: string
          format: date-time
          description: Creation timestamp
          example: "2025-10-25T10:30:00Z"
        expires_at:
          type: string
          format: date-time
          description: Expiration timestamp
          example: "2025-11-24T10:30:00Z"
        days_remaining:
          type: integer
          description: Days remaining until expiration
          example: 25
        claimer_ispb:
          type: string
          description: Claimer institution ISPB
          example: "87654321"
        owner_ispb:
          type: string
          description: Owner institution ISPB
          example: "12345678"

    CreateClaimRequest:
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
          description: PIX key to claim
          example: "12345678901"
        key_type:
          type: string
          enum: [CPF, CNPJ, EMAIL, PHONE, EVP]
          description: PIX key type
          example: CPF
        claimer_account:
          $ref: '#/components/schemas/Account'
        claim_reason:
          type: string
          enum: [OWNERSHIP, FRAUD]
          description: Reason for claiming
          example: OWNERSHIP
        idempotency_key:
          type: string
          format: uuid
          description: Idempotency key
          example: "770e8400-e29b-41d4-a716-446655440000"

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
          description: Internal portability UUID
          example: "880e8400-e29b-41d4-a716-446655440000"
        external_id:
          type: string
          description: Bacen DICT portability external ID
          example: "bacen-portability-id-789"
        entry_key:
          type: string
          description: PIX key being ported
          example: "12345678901"
        status:
          type: string
          description: Portability status
          enum: [PENDING, CONFIRMED, COMPLETED, CANCELLED, REJECTED]
          example: PENDING
        from_ispb:
          type: string
          description: Source institution ISPB
          example: "12345678"
        to_ispb:
          type: string
          description: Destination institution ISPB
          example: "11111111"
        created_at:
          type: string
          format: date-time
          description: Creation timestamp
          example: "2025-10-25T10:30:00Z"
        completed_at:
          type: string
          format: date-time
          description: Completion timestamp
          nullable: true
          example: "2025-10-25T12:00:00Z"

    CreatePortabilityRequest:
      type: object
      required:
        - entry_key
        - key_type
        - new_account
        - idempotency_key
      properties:
        entry_key:
          type: string
          description: PIX key to port
          example: "12345678901"
        key_type:
          type: string
          enum: [CPF, CNPJ, EMAIL, PHONE, EVP]
          description: PIX key type
          example: CPF
        new_account:
          $ref: '#/components/schemas/Account'
        idempotency_key:
          type: string
          format: uuid
          description: Idempotency key
          example: "990e8400-e29b-41d4-a716-446655440000"

    Pagination:
      type: object
      properties:
        current_page:
          type: integer
          description: Current page number
          example: 1
        total_pages:
          type: integer
          description: Total number of pages
          example: 5
        total_entries:
          type: integer
          description: Total number of entries
          example: 94
        per_page:
          type: integer
          description: Results per page
          example: 20

    # ==================== Connect Admin Schemas ====================

    Workflow:
      type: object
      properties:
        workflow_id:
          type: string
          description: Workflow ID
          example: "entry-create-uuid-123"
        workflow_type:
          type: string
          description: Workflow type
          enum: [ENTRY_CREATE, ENTRY_DELETE, CLAIM_CREATE, CLAIM_CONFIRM, PORTABILITY_CREATE]
          example: ENTRY_CREATE
        status:
          type: string
          description: Workflow status
          enum: [RUNNING, COMPLETED, FAILED, CANCELLED, TIMED_OUT]
          example: RUNNING
        start_time:
          type: string
          format: date-time
          description: Workflow start time
          example: "2025-10-25T10:30:00Z"
        close_time:
          type: string
          format: date-time
          nullable: true
          description: Workflow close time
          example: null
        execution_time_ms:
          type: integer
          description: Execution time in milliseconds
          example: 5000
        input:
          type: object
          description: Workflow input parameters
          additionalProperties: true
        current_activity:
          type: string
          nullable: true
          description: Current running activity
          example: "SyncToBacenDICT"
        history_length:
          type: integer
          description: Number of events in workflow history
          example: 15
        retry_count:
          type: integer
          description: Number of retries
          example: 0

    WorkflowDetails:
      allOf:
        - $ref: '#/components/schemas/Workflow'
        - type: object
          properties:
            output:
              type: object
              nullable: true
              description: Workflow output
              additionalProperties: true
            pending_activities:
              type: array
              description: List of pending activities
              items:
                type: object
                properties:
                  activity_id:
                    type: string
                  activity_type:
                    type: string
                  scheduled_time:
                    type: string
                    format: date-time
                  attempt:
                    type: integer
                  max_attempts:
                    type: integer
            history:
              type: array
              description: Workflow execution history
              items:
                type: object
                properties:
                  event_id:
                    type: integer
                  event_type:
                    type: string
                  timestamp:
                    type: string
                    format: date-time
                  attributes:
                    type: object
                    additionalProperties: true

    CreateWorkflowRequest:
      type: object
      required:
        - workflow_type
        - input
      properties:
        workflow_type:
          type: string
          enum: [ENTRY_CREATE, ENTRY_DELETE, CLAIM_CREATE, CLAIM_CONFIRM, PORTABILITY_CREATE]
          description: Workflow type
          example: ENTRY_CREATE
        input:
          type: object
          description: Workflow-specific input parameters
          additionalProperties: true
        workflow_id:
          type: string
          description: Custom workflow ID (auto-generated if not provided)
          example: "custom-workflow-id-123"
        task_queue:
          type: string
          description: Task queue name
          default: "dict-task-queue"
          example: "dict-task-queue"
        execution_timeout_seconds:
          type: integer
          description: Execution timeout in seconds
          default: 300
          example: 300

    HealthStatus:
      type: object
      properties:
        status:
          type: string
          enum: [healthy, degraded, unhealthy]
          description: Overall system health
          example: healthy
        timestamp:
          type: string
          format: date-time
          description: Health check timestamp
          example: "2025-10-25T10:30:00Z"
        services:
          type: object
          description: Individual service statuses
          additionalProperties:
            type: object
            properties:
              status:
                type: string
                enum: [up, degraded, down]
              response_time_ms:
                type: integer
        metrics:
          type: object
          description: System metrics
          additionalProperties:
            type: number

    # ==================== Error Schemas ====================

    Error:
      type: object
      required:
        - code
        - message
        - timestamp
      properties:
        code:
          type: string
          description: Error code
          example: "VALIDATION_ERROR"
        message:
          type: string
          description: Human-readable error message
          example: "Invalid CPF format"
        details:
          type: array
          description: Additional error details
          items:
            type: object
            properties:
              field:
                type: string
              message:
                type: string
        timestamp:
          type: string
          format: date-time
          description: Error timestamp
          example: "2025-10-25T10:30:00Z"
        trace_id:
          type: string
          description: Request trace ID for debugging
          example: "abc123"
        path:
          type: string
          description: Request path
          example: "/dict/v1/keys"
        method:
          type: string
          description: HTTP method
          example: "POST"

  responses:
    BadRequest:
      description: Bad Request - Invalid input
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            error:
              code: "VALIDATION_ERROR"
              message: "Invalid CPF format"
              details:
                - field: "key_value"
                  message: "CPF must have 11 digits"
              timestamp: "2025-10-25T10:30:00Z"
              trace_id: "abc123"

    Unauthorized:
      description: Unauthorized - Missing or invalid JWT token
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            error:
              code: "UNAUTHORIZED"
              message: "Missing or invalid JWT token"
              timestamp: "2025-10-25T10:30:00Z"
              trace_id: "abc123"

    Forbidden:
      description: Forbidden - Insufficient permissions
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            error:
              code: "FORBIDDEN"
              message: "Insufficient permissions. Required scope: dict:write"
              timestamp: "2025-10-25T10:30:00Z"
              trace_id: "abc123"

    NotFound:
      description: Not Found - Resource not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            error:
              code: "KEY_NOT_FOUND"
              message: "PIX key not found"
              timestamp: "2025-10-25T10:30:00Z"
              trace_id: "abc123"

    Conflict:
      description: Conflict - Resource already exists or conflict state
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            error:
              code: "KEY_ALREADY_EXISTS"
              message: "PIX key already registered"
              timestamp: "2025-10-25T10:30:00Z"
              trace_id: "abc123"

    InternalServerError:
      description: Internal Server Error
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            error:
              code: "INTERNAL_ERROR"
              message: "An unexpected error occurred"
              timestamp: "2025-10-25T10:30:00Z"
              trace_id: "abc123"

  parameters:
    PageParam:
      name: page
      in: query
      description: Page number for pagination
      schema:
        type: integer
        default: 1
        minimum: 1
      example: 1

    LimitParam:
      name: limit
      in: query
      description: Number of results per page
      schema:
        type: integer
        default: 20
        minimum: 1
        maximum: 100
      example: 20

paths:
  # ==================== Core DICT API ====================

  /dict/v1/keys:
    post:
      summary: Create PIX Key Entry
      description: Create a new PIX key entry in the DICT
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
              $ref: '#/components/schemas/CreateEntryRequest'
      responses:
        '201':
          description: Entry created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Entry'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '409':
          $ref: '#/components/responses/Conflict'
        '500':
          $ref: '#/components/responses/InternalServerError'

    get:
      summary: List User Entries
      description: List all PIX keys for the authenticated user
      operationId: listEntries
      tags:
        - Entry Management
      security:
        - BearerAuth: []
      parameters:
        - $ref: '#/components/parameters/PageParam'
        - $ref: '#/components/parameters/LimitParam'
        - name: status
          in: query
          description: Filter by status
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
                    $ref: '#/components/schemas/Pagination'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '500':
          $ref: '#/components/responses/InternalServerError'

  /dict/v1/keys/{key}:
    get:
      summary: Get Entry by Key
      description: Retrieve a PIX key entry by key value
      operationId: getEntry
      tags:
        - Entry Management
      security:
        - BearerAuth: []
      parameters:
        - name: key
          in: path
          required: true
          description: PIX key value
          schema:
            type: string
          example: "12345678901"
        - name: key_type
          in: query
          description: PIX key type
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
        '401':
          $ref: '#/components/responses/Unauthorized'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/InternalServerError'

  /dict/v1/keys/{id}:
    delete:
      summary: Delete Entry
      description: Delete a PIX key entry (soft delete)
      operationId: deleteEntry
      tags:
        - Entry Management
      security:
        - BearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          description: Entry UUID
          schema:
            type: string
            format: uuid
          example: "550e8400-e29b-41d4-a716-446655440000"
      responses:
        '204':
          description: Entry deleted successfully
        '401':
          $ref: '#/components/responses/Unauthorized'
        '403':
          $ref: '#/components/responses/Forbidden'
        '404':
          $ref: '#/components/responses/NotFound'
        '409':
          $ref: '#/components/responses/Conflict'
        '500':
          $ref: '#/components/responses/InternalServerError'

  /dict/v1/claims:
    post:
      summary: Create Claim
      description: Initiate a claim for PIX key ownership
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
              $ref: '#/components/schemas/CreateClaimRequest'
      responses:
        '201':
          description: Claim created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Claim'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '409':
          $ref: '#/components/responses/Conflict'
        '500':
          $ref: '#/components/responses/InternalServerError'

    get:
      summary: List User Claims
      description: List all claims for the authenticated user
      operationId: listClaims
      tags:
        - Claim Management
      security:
        - BearerAuth: []
      parameters:
        - $ref: '#/components/parameters/PageParam'
        - $ref: '#/components/parameters/LimitParam'
        - name: status
          in: query
          description: Filter by status
          schema:
            type: string
            enum: [OPEN, WAITING_RESOLUTION, CONFIRMED, CANCELLED, COMPLETED, EXPIRED]
        - name: role
          in: query
          description: Filter by role
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
                    $ref: '#/components/schemas/Pagination'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '500':
          $ref: '#/components/responses/InternalServerError'

  /dict/v1/claims/{id}:
    get:
      summary: Get Claim by ID
      description: Retrieve claim details by ID
      operationId: getClaim
      tags:
        - Claim Management
      security:
        - BearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          description: Claim UUID
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
        '401':
          $ref: '#/components/responses/Unauthorized'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/InternalServerError'

    delete:
      summary: Cancel Claim
      description: Cancel a claim
      operationId: cancelClaim
      tags:
        - Claim Management
      security:
        - BearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          description: Claim UUID
          schema:
            type: string
            format: uuid
        - name: reason
          in: query
          description: Cancellation reason
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
        '401':
          $ref: '#/components/responses/Unauthorized'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/InternalServerError'

  /dict/v1/claims/{id}/confirm:
    put:
      summary: Confirm Claim
      description: Confirm claim (owner accepts ownership transfer)
      operationId: confirmClaim
      tags:
        - Claim Management
      security:
        - BearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          description: Claim UUID
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
        '401':
          $ref: '#/components/responses/Unauthorized'
        '403':
          $ref: '#/components/responses/Forbidden'
        '404':
          $ref: '#/components/responses/NotFound'
        '409':
          $ref: '#/components/responses/Conflict'
        '500':
          $ref: '#/components/responses/InternalServerError'

  /dict/v1/portabilities:
    post:
      summary: Initiate Portability
      description: Initiate portability of PIX key to another institution
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
              $ref: '#/components/schemas/CreatePortabilityRequest'
      responses:
        '201':
          description: Portability initiated
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Portability'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '409':
          $ref: '#/components/responses/Conflict'
        '500':
          $ref: '#/components/responses/InternalServerError'

  /dict/v1/portabilities/{id}:
    get:
      summary: Get Portability Status
      description: Retrieve portability status
      operationId: getPortability
      tags:
        - Portability Management
      security:
        - BearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          description: Portability UUID
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
        '401':
          $ref: '#/components/responses/Unauthorized'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/InternalServerError'

  # ==================== Connect Admin API ====================

  /connect/admin/v1/workflows:
    get:
      summary: List Workflows
      description: List all Temporal workflows with filtering
      operationId: listWorkflows
      tags:
        - Workflow Management
      security:
        - BearerAuth: []
      parameters:
        - $ref: '#/components/parameters/PageParam'
        - name: limit
          in: query
          description: Number of results per page
          schema:
            type: integer
            default: 50
            minimum: 1
            maximum: 100
        - name: workflow_type
          in: query
          description: Filter by workflow type
          schema:
            type: string
            enum: [ENTRY_CREATE, ENTRY_DELETE, CLAIM_CREATE, CLAIM_CONFIRM, PORTABILITY_CREATE]
        - name: status
          in: query
          description: Filter by status
          schema:
            type: string
            enum: [RUNNING, COMPLETED, FAILED, CANCELLED, TIMED_OUT]
      responses:
        '200':
          description: List of workflows
          content:
            application/json:
              schema:
                type: object
                properties:
                  workflows:
                    type: array
                    items:
                      $ref: '#/components/schemas/Workflow'
                  pagination:
                    $ref: '#/components/schemas/Pagination'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '403':
          $ref: '#/components/responses/Forbidden'
        '500':
          $ref: '#/components/responses/InternalServerError'

    post:
      summary: Create/Retry Workflow
      description: Manually create or retry a workflow
      operationId: createWorkflow
      tags:
        - Workflow Management
      security:
        - BearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateWorkflowRequest'
      responses:
        '201':
          description: Workflow created successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  workflow_id:
                    type: string
                  workflow_type:
                    type: string
                  status:
                    type: string
                  start_time:
                    type: string
                    format: date-time
                  message:
                    type: string
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '403':
          $ref: '#/components/responses/Forbidden'
        '409':
          $ref: '#/components/responses/Conflict'
        '500':
          $ref: '#/components/responses/InternalServerError'

  /connect/admin/v1/workflows/{id}:
    get:
      summary: Get Workflow Details
      description: Retrieve detailed information about a workflow
      operationId: getWorkflow
      tags:
        - Workflow Management
      security:
        - BearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          description: Workflow ID
          schema:
            type: string
        - name: include_history
          in: query
          description: Include full workflow history
          schema:
            type: boolean
            default: false
      responses:
        '200':
          description: Workflow details
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/WorkflowDetails'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '403':
          $ref: '#/components/responses/Forbidden'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/InternalServerError'

  /connect/admin/v1/workflows/{id}/cancel:
    post:
      summary: Cancel Workflow
      description: Cancel a running workflow
      operationId: cancelWorkflow
      tags:
        - Workflow Management
      security:
        - BearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          description: Workflow ID
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                reason:
                  type: string
                  description: Cancellation reason
      responses:
        '200':
          description: Workflow cancelled
          content:
            application/json:
              schema:
                type: object
                properties:
                  workflow_id:
                    type: string
                  status:
                    type: string
                  cancelled_at:
                    type: string
                    format: date-time
                  cancellation_reason:
                    type: string
                  message:
                    type: string
        '401':
          $ref: '#/components/responses/Unauthorized'
        '403':
          $ref: '#/components/responses/Forbidden'
        '404':
          $ref: '#/components/responses/NotFound'
        '409':
          $ref: '#/components/responses/Conflict'
        '500':
          $ref: '#/components/responses/InternalServerError'

  /connect/admin/v1/health:
    get:
      summary: Get System Health
      description: Retrieve system health status and metrics
      operationId: getHealth
      tags:
        - Monitoring
      security:
        - BearerAuth: []
      responses:
        '200':
          description: System health status
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HealthStatus'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '500':
          $ref: '#/components/responses/InternalServerError'
```

---

## 3. Usage Instructions

### 3.1. Import to Swagger UI

1. Copy the YAML specification above
2. Navigate to [Swagger Editor](https://editor.swagger.io/)
3. Paste the YAML content
4. View interactive documentation

### 3.2. Import to Postman

1. Open Postman
2. Click **Import** > **Raw text**
3. Paste the YAML specification
4. Click **Continue** > **Import**
5. Collection will be created with all endpoints

### 3.3. Generate Client SDK

**Go Client**:
```bash
# Using openapi-generator
openapi-generator generate \
  -i openapi.yaml \
  -g go \
  -o ./sdk/go
```

**TypeScript Client**:
```bash
openapi-generator generate \
  -i openapi.yaml \
  -g typescript-axios \
  -o ./sdk/typescript
```

**Python Client**:
```bash
openapi-generator generate \
  -i openapi.yaml \
  -g python \
  -o ./sdk/python
```

### 3.4. Validate Specification

```bash
# Using openapi-validator
npx @openapitools/openapi-generator-cli validate -i openapi.yaml

# Using Spectral
npx @stoplight/spectral-cli lint openapi.yaml
```

---

## Rastreabilidade

### Requisitos Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RF-OAS-001 | OpenAPI 3.0 specification for Core DICT | [API-002](./API-002_Core_DICT_REST_API.md) | ✅ Especificado |
| RF-OAS-002 | OpenAPI 3.0 specification for Connect Admin | [API-003](./API-003_Connect_Admin_API.md) | ✅ Especificado |
| RF-OAS-003 | Complete schemas and examples | Best Practices | ✅ Especificado |

---

**Referências**:
- [API-002: Core DICT REST API](./API-002_Core_DICT_REST_API.md)
- [API-003: Connect Admin API](./API-003_Connect_Admin_API.md)
- [OpenAPI 3.0 Specification](https://swagger.io/specification/)
- [OpenAPI Generator](https://openapi-generator.tech/)
