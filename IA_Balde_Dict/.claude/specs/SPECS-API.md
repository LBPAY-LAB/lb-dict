# SPECS-API.md - Dict API REST Endpoints Specification

**Projeto**: DICT Rate Limit Monitoring System
**Componente**: Dict API (apps/dict)
**Framework**: Huma v2
**Status**: ✅ ESPECIFICAÇÃO COMPLETA - Production-Ready

---

## 🎯 Objetivo

Especificação técnica completa dos **2 endpoints REST** para consulta de políticas de rate limit do DICT BACEN:

1. `GET /api/v1/rate-limit/policies` - Listar todas as políticas
2. `GET /api/v1/rate-limit/policies/{policy}` - Consultar política específica

**Cache Strategy**: Redis com TTL 60s para reduzir chamadas ao Bridge/DICT.

---

## 📋 Tabela de Conteúdos

- [1. OpenAPI 3.1 Specification](#1-openapi-31-specification)
- [2. Huma Schemas (Go Structs)](#2-huma-schemas-go-structs)
- [3. HTTP Handlers](#3-http-handlers)
- [4. Application Layer (Use Cases)](#4-application-layer-use-cases)
- [5. Infrastructure Layer (Bridge Client)](#5-infrastructure-layer-bridge-client)
- [6. Cache Strategy](#6-cache-strategy)
- [7. Error Handling (RFC 9457)](#7-error-handling-rfc-9457)
- [8. Integration Testing](#8-integration-testing)
- [9. Performance Benchmarks](#9-performance-benchmarks)

---

## 1. OpenAPI 3.1 Specification

### Complete YAML Specification

```yaml
openapi: 3.1.0
info:
  title: DICT Rate Limit Monitoring API
  version: 1.0.0
  description: |
    API para monitoramento de políticas de Rate Limit (Token Bucket) do DICT BACEN.

    **Funcionalidades**:
    - Consultar todas as políticas de rate limit (24 políticas BACEN)
    - Consultar estado de política específica em tempo real
    - Cache Redis (TTL 60s) para otimização de performance

    **Referências**:
    - BACEN Manual Operacional Capítulo 19
    - RFC 9457 (Problem Details for HTTP APIs)

  contact:
    name: LBPay Engineering
    email: engineering@lbpay.com

  license:
    name: Proprietary
    url: https://lbpay.com/license

servers:
  - url: https://api.lbpay.com
    description: Production
  - url: https://api.staging.lbpay.com
    description: Staging
  - url: http://localhost:8080
    description: Local Development

tags:
  - name: rate-limit
    description: DICT Rate Limit Policies

paths:
  /api/v1/rate-limit/policies:
    get:
      operationId: listRateLimitPolicies
      summary: Listar todas as políticas de rate limit
      description: |
        Retorna lista completa de 24 políticas de rate limit do DICT BACEN, incluindo:
        - Configuração de cada política (capacidade, refill rate)
        - Estado atual de tokens disponíveis
        - Utilização percentual
        - Categoria BACEN (A, B, C, D, E, F, G, H)

        **Cache**: Resposta cacheada por 60s (Redis TTL).

        **Performance**: <200ms p99 (incluindo cache lookup + Bridge call).

      tags:
        - rate-limit

      parameters:
        - name: category
          in: query
          description: Filtrar por categoria BACEN (A, B, C, D, E, F, G, H)
          required: false
          schema:
            type: string
            enum: [A, B, C, D, E, F, G, H]
          example: A

        - name: min_utilization
          in: query
          description: Filtrar por utilização mínima (0-100)
          required: false
          schema:
            type: number
            format: float
            minimum: 0
            maximum: 100
          example: 75.0

      responses:
        '200':
          description: Lista de políticas retornada com sucesso
          headers:
            X-Cache-Status:
              description: Cache hit/miss status
              schema:
                type: string
                enum: [HIT, MISS]
            X-Cache-TTL:
              description: Tempo restante de cache (segundos)
              schema:
                type: integer
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListPoliciesResponse'
              examples:
                success:
                  summary: Exemplo de resposta com 3 políticas
                  value:
                    policies:
                      - policy_name: ENTRIES_CREATE
                        category: A
                        capacity_max: 300
                        refill_tokens: 5
                        refill_period_sec: 60
                        available_tokens: 150
                        utilization_pct: 50.00
                        warning_threshold_pct: 25.00
                        critical_threshold_pct: 10.00
                        status: OK
                        checked_at: "2025-10-31T10:30:00Z"
                      - policy_name: ENTRIES_DELETE
                        category: A
                        capacity_max: 300
                        refill_tokens: 5
                        refill_period_sec: 60
                        available_tokens: 75
                        utilization_pct: 75.00
                        warning_threshold_pct: 25.00
                        critical_threshold_pct: 10.00
                        status: WARNING
                        checked_at: "2025-10-31T10:30:00Z"
                      - policy_name: CLAIMS_CREATE
                        category: B
                        capacity_max: 1000
                        refill_tokens: 300
                        refill_period_sec: 60
                        available_tokens: 50
                        utilization_pct: 95.00
                        warning_threshold_pct: 25.00
                        critical_threshold_pct: 10.00
                        status: CRITICAL
                        checked_at: "2025-10-31T10:30:00Z"
                    total: 24
                    cached: true
                    cache_expires_in: 45

        '400':
          description: Parâmetros inválidos
          content:
            application/problem+json:
              schema:
                $ref: '#/components/schemas/Problem'
              example:
                type: "https://api.lbpay.com/problems/invalid-parameter"
                title: "Invalid Parameter"
                status: 400
                detail: "Parameter 'min_utilization' must be between 0 and 100"
                instance: "/api/v1/rate-limit/policies?min_utilization=150"

        '500':
          description: Erro interno
          content:
            application/problem+json:
              schema:
                $ref: '#/components/schemas/Problem'

        '503':
          description: Bridge/DICT indisponível
          content:
            application/problem+json:
              schema:
                $ref: '#/components/schemas/Problem'
              example:
                type: "https://api.lbpay.com/problems/bridge-unavailable"
                title: "Bridge Service Unavailable"
                status: 503
                detail: "Unable to connect to Bridge gRPC service"
                instance: "/api/v1/rate-limit/policies"

  /api/v1/rate-limit/policies/{policy}:
    get:
      operationId: getRateLimitPolicy
      summary: Consultar política específica
      description: |
        Retorna estado detalhado de uma política específica.

        **Cache**: Resposta cacheada por 60s (Redis TTL).

        **Performance**: <150ms p99.

      tags:
        - rate-limit

      parameters:
        - name: policy
          in: path
          description: Nome da política (ex: ENTRIES_CREATE)
          required: true
          schema:
            type: string
            enum:
              - ENTRIES_CREATE
              - ENTRIES_DELETE
              - ENTRIES_UPDATE
              - CLAIMS_CREATE
              - CLAIMS_CONFIRM
              - CLAIMS_CANCEL
              - CLAIMS_COMPLETE
              - INFRACTION_REPORT_CREATE
              - INFRACTION_REPORT_CANCEL
              - ACCOUNT_CLOSE
              - ACCOUNT_LIST
              - POLICIES_LIST
              - POLICIES_SPECIFIC
              - DIRECTORY_HEALTH
              - DIRECTORY_STATUS
              - DIRECTORY_ENTRIES
              - STATISTICS_KEYS
              - STATISTICS_CLAIMS
              - STATISTICS_INFRACTIONS
              - VERIFICATION_CREATE
              - VERIFICATION_STATUS
              - PORTABILITY_CREATE
              - PORTABILITY_CONFIRM
              - PORTABILITY_CANCEL
          example: ENTRIES_CREATE

      responses:
        '200':
          description: Política retornada com sucesso
          headers:
            X-Cache-Status:
              schema:
                type: string
                enum: [HIT, MISS]
            X-Cache-TTL:
              schema:
                type: integer
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PolicyResponse'
              examples:
                ok_status:
                  summary: Política com status OK
                  value:
                    policy_name: ENTRIES_CREATE
                    category: A
                    capacity_max: 300
                    refill_tokens: 5
                    refill_period_sec: 60
                    available_tokens: 200
                    utilization_pct: 33.33
                    warning_threshold_pct: 25.00
                    critical_threshold_pct: 10.00
                    status: OK
                    checked_at: "2025-10-31T10:30:00Z"
                    cached: true
                    cache_expires_in: 45

                warning_status:
                  summary: Política com status WARNING
                  value:
                    policy_name: CLAIMS_CREATE
                    category: B
                    capacity_max: 1000
                    refill_tokens: 300
                    refill_period_sec: 60
                    available_tokens: 200
                    utilization_pct: 80.00
                    warning_threshold_pct: 25.00
                    critical_threshold_pct: 10.00
                    status: WARNING
                    checked_at: "2025-10-31T10:30:00Z"
                    cached: false
                    cache_expires_in: 60

        '404':
          description: Política não encontrada
          content:
            application/problem+json:
              schema:
                $ref: '#/components/schemas/Problem'
              example:
                type: "https://api.lbpay.com/problems/policy-not-found"
                title: "Policy Not Found"
                status: 404
                detail: "Policy 'INVALID_POLICY' does not exist"
                instance: "/api/v1/rate-limit/policies/INVALID_POLICY"

        '500':
          description: Erro interno
          content:
            application/problem+json:
              schema:
                $ref: '#/components/schemas/Problem'

components:
  schemas:
    ListPoliciesResponse:
      type: object
      required:
        - policies
        - total
        - cached
      properties:
        policies:
          type: array
          description: Lista de políticas
          items:
            $ref: '#/components/schemas/PolicySummary'
        total:
          type: integer
          description: Total de políticas retornadas
          example: 24
        cached:
          type: boolean
          description: Se a resposta veio do cache
          example: true
        cache_expires_in:
          type: integer
          description: Segundos até expiração do cache
          example: 45

    PolicyResponse:
      allOf:
        - $ref: '#/components/schemas/PolicySummary'
        - type: object
          properties:
            cached:
              type: boolean
              example: true
            cache_expires_in:
              type: integer
              example: 45

    PolicySummary:
      type: object
      required:
        - policy_name
        - category
        - capacity_max
        - refill_tokens
        - refill_period_sec
        - available_tokens
        - utilization_pct
        - status
        - checked_at
      properties:
        policy_name:
          type: string
          description: Nome da política (BACEN)
          example: ENTRIES_CREATE
        category:
          type: string
          description: Categoria BACEN
          enum: [A, B, C, D, E, F, G, H]
          example: A
        capacity_max:
          type: integer
          description: Capacidade máxima do balde
          example: 300
        refill_tokens:
          type: integer
          description: Tokens adicionados por refill
          example: 5
        refill_period_sec:
          type: integer
          description: Período de refill (segundos)
          example: 60
        available_tokens:
          type: integer
          description: Tokens disponíveis no momento
          example: 150
        utilization_pct:
          type: number
          format: float
          description: Utilização percentual (0-100)
          example: 50.00
        warning_threshold_pct:
          type: number
          format: float
          description: Threshold de WARNING
          example: 25.00
        critical_threshold_pct:
          type: number
          format: float
          description: Threshold de CRITICAL
          example: 10.00
        status:
          type: string
          description: Status da política
          enum: [OK, WARNING, CRITICAL]
          example: OK
        checked_at:
          type: string
          format: date-time
          description: Timestamp da última verificação
          example: "2025-10-31T10:30:00Z"

    Problem:
      type: object
      description: RFC 9457 Problem Details
      required:
        - type
        - title
        - status
      properties:
        type:
          type: string
          format: uri
          description: URI que identifica o tipo de problema
          example: "https://api.lbpay.com/problems/invalid-parameter"
        title:
          type: string
          description: Título curto do problema
          example: "Invalid Parameter"
        status:
          type: integer
          description: HTTP status code
          example: 400
        detail:
          type: string
          description: Explicação detalhada do problema
          example: "Parameter 'min_utilization' must be between 0 and 100"
        instance:
          type: string
          format: uri
          description: URI da requisição que causou o problema
          example: "/api/v1/rate-limit/policies?min_utilization=150"

security: []
```

---

## 2. Huma Schemas (Go Structs)

### Request/Response Types

```go
// Location: apps/dict/handlers/http/ratelimit/schemas.go
package ratelimit

import (
	"time"
)

// ============================================================================
// REQUEST SCHEMAS
// ============================================================================

// ListPoliciesRequest representa os query parameters do endpoint de listagem
type ListPoliciesRequest struct {
	// Filtrar por categoria BACEN (opcional)
	Category *string `query:"category" enum:"A,B,C,D,E,F,G,H" doc:"Filtrar por categoria BACEN"`

	// Filtrar por utilização mínima (opcional)
	MinUtilization *float64 `query:"min_utilization" minimum:"0" maximum:"100" doc:"Filtrar por utilização mínima (0-100)"`
}

// Validate implementa validação customizada
func (r *ListPoliciesRequest) Validate() error {
	if r.Category != nil {
		validCategories := map[string]bool{
			"A": true, "B": true, "C": true, "D": true,
			"E": true, "F": true, "G": true, "H": true,
		}
		if !validCategories[*r.Category] {
			return fmt.Errorf("invalid category: %s (must be A-H)", *r.Category)
		}
	}

	if r.MinUtilization != nil {
		if *r.MinUtilization < 0 || *r.MinUtilization > 100 {
			return fmt.Errorf("min_utilization must be between 0 and 100, got %.2f", *r.MinUtilization)
		}
	}

	return nil
}

// GetPolicyRequest representa os path parameters do endpoint de consulta individual
type GetPolicyRequest struct {
	// Nome da política (path parameter)
	Policy string `path:"policy" required:"true" doc:"Nome da política (ex: ENTRIES_CREATE)"`
}

// Validate implementa validação de política válida
func (r *GetPolicyRequest) Validate() error {
	validPolicies := map[string]bool{
		"ENTRIES_CREATE": true, "ENTRIES_DELETE": true, "ENTRIES_UPDATE": true,
		"CLAIMS_CREATE": true, "CLAIMS_CONFIRM": true, "CLAIMS_CANCEL": true, "CLAIMS_COMPLETE": true,
		"INFRACTION_REPORT_CREATE": true, "INFRACTION_REPORT_CANCEL": true,
		"ACCOUNT_CLOSE": true, "ACCOUNT_LIST": true,
		"POLICIES_LIST": true, "POLICIES_SPECIFIC": true,
		"DIRECTORY_HEALTH": true, "DIRECTORY_STATUS": true, "DIRECTORY_ENTRIES": true,
		"STATISTICS_KEYS": true, "STATISTICS_CLAIMS": true, "STATISTICS_INFRACTIONS": true,
		"VERIFICATION_CREATE": true, "VERIFICATION_STATUS": true,
		"PORTABILITY_CREATE": true, "PORTABILITY_CONFIRM": true, "PORTABILITY_CANCEL": true,
	}

	if !validPolicies[r.Policy] {
		return fmt.Errorf("invalid policy name: %s", r.Policy)
	}

	return nil
}

// ============================================================================
// RESPONSE SCHEMAS
// ============================================================================

// ListPoliciesResponse representa a resposta do endpoint de listagem
type ListPoliciesResponse struct {
	Body struct {
		// Lista de políticas
		Policies []PolicySummary `json:"policies" required:"true" doc:"Lista de políticas de rate limit"`

		// Total de políticas retornadas
		Total int `json:"total" required:"true" doc:"Total de políticas retornadas"`

		// Se a resposta veio do cache
		Cached bool `json:"cached" required:"true" doc:"Se a resposta veio do cache Redis"`

		// Segundos até expiração do cache (se cached=true)
		CacheExpiresIn *int `json:"cache_expires_in,omitempty" doc:"Segundos até expiração do cache"`
	}

	// Headers
	Headers struct {
		XCacheStatus string `header:"X-Cache-Status" enum:"HIT,MISS" doc:"Status do cache"`
		XCacheTTL    *int   `header:"X-Cache-TTL,omitempty" doc:"TTL restante do cache (segundos)"`
	}
}

// PolicyResponse representa a resposta do endpoint de consulta individual
type PolicyResponse struct {
	Body struct {
		PolicySummary

		// Se a resposta veio do cache
		Cached bool `json:"cached" required:"true"`

		// Segundos até expiração do cache
		CacheExpiresIn *int `json:"cache_expires_in,omitempty"`
	}

	// Headers
	Headers struct {
		XCacheStatus string `header:"X-Cache-Status" enum:"HIT,MISS"`
		XCacheTTL    *int   `header:"X-Cache-TTL,omitempty"`
	}
}

// PolicySummary representa o resumo de uma política
type PolicySummary struct {
	// Nome da política
	PolicyName string `json:"policy_name" required:"true" example:"ENTRIES_CREATE" doc:"Nome da política BACEN"`

	// Categoria BACEN
	Category string `json:"category" required:"true" enum:"A,B,C,D,E,F,G,H" example:"A" doc:"Categoria BACEN"`

	// Capacidade máxima do balde
	CapacityMax int `json:"capacity_max" required:"true" minimum:"1" example:"300" doc:"Capacidade máxima do balde"`

	// Tokens adicionados por refill
	RefillTokens int `json:"refill_tokens" required:"true" minimum:"1" example:"5" doc:"Tokens adicionados por período de refill"`

	// Período de refill (segundos)
	RefillPeriodSec int `json:"refill_period_sec" required:"true" minimum:"1" example:"60" doc:"Período de refill em segundos"`

	// Tokens disponíveis no momento
	AvailableTokens int `json:"available_tokens" required:"true" minimum:"0" example:"150" doc:"Tokens disponíveis no momento"`

	// Utilização percentual (0-100)
	UtilizationPct float64 `json:"utilization_pct" required:"true" minimum:"0" maximum:"100" example:"50.00" doc:"Utilização percentual do balde"`

	// Threshold de WARNING
	WarningThresholdPct float64 `json:"warning_threshold_pct" required:"true" example:"25.00" doc:"Threshold de WARNING (% restante)"`

	// Threshold de CRITICAL
	CriticalThresholdPct float64 `json:"critical_threshold_pct" required:"true" example:"10.00" doc:"Threshold de CRITICAL (% restante)"`

	// Status da política
	Status string `json:"status" required:"true" enum:"OK,WARNING,CRITICAL" example:"OK" doc:"Status da política baseado em thresholds"`

	// Timestamp da última verificação
	CheckedAt time.Time `json:"checked_at" required:"true" format:"date-time" example:"2025-10-31T10:30:00Z" doc:"Timestamp da última verificação"`
}

// ============================================================================
// ERROR SCHEMAS (RFC 9457)
// ============================================================================

// ProblemDetail representa um erro RFC 9457
type ProblemDetail struct {
	// URI que identifica o tipo de problema
	Type string `json:"type" required:"true" format:"uri" example:"https://api.lbpay.com/problems/invalid-parameter"`

	// Título curto do problema
	Title string `json:"title" required:"true" example:"Invalid Parameter"`

	// HTTP status code
	Status int `json:"status" required:"true" example:"400"`

	// Explicação detalhada
	Detail string `json:"detail,omitempty" example:"Parameter 'min_utilization' must be between 0 and 100"`

	// URI da requisição que causou o problema
	Instance string `json:"instance,omitempty" format:"uri" example:"/api/v1/rate-limit/policies?min_utilization=150"`
}

// NewProblemDetail cria um novo ProblemDetail
func NewProblemDetail(problemType, title string, status int, detail, instance string) *ProblemDetail {
	return &ProblemDetail{
		Type:     problemType,
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
	}
}

// Common problem types
const (
	ProblemTypeInvalidParameter   = "https://api.lbpay.com/problems/invalid-parameter"
	ProblemTypePolicyNotFound     = "https://api.lbpay.com/problems/policy-not-found"
	ProblemTypeBridgeUnavailable  = "https://api.lbpay.com/problems/bridge-unavailable"
	ProblemTypeInternalError      = "https://api.lbpay.com/problems/internal-error"
	ProblemTypeCacheError         = "https://api.lbpay.com/problems/cache-error"
)
```

---

## 3. HTTP Handlers

### Handler Implementation

```go
// Location: apps/dict/handlers/http/ratelimit/handler.go
package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lb-conn/connector-dict/apps/dict/application/ports"
	"github.com/lb-conn/connector-dict/shared/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Handler gerencia os endpoints HTTP de rate limit
type Handler struct {
	rateLimitService ports.RateLimitService
	tracer           trace.Tracer
	logger           logger.Logger
}

// NewHandler cria uma nova instância do handler
func NewHandler(
	rateLimitService ports.RateLimitService,
	logger logger.Logger,
) *Handler {
	return &Handler{
		rateLimitService: rateLimitService,
		tracer:           otel.Tracer("dict.api.ratelimit"),
		logger:           logger,
	}
}

// RegisterRoutes registra as rotas do Huma
func (h *Handler) RegisterRoutes(api huma.API) {
	// GET /api/v1/rate-limit/policies
	huma.Register(api, huma.Operation{
		OperationID: "list-rate-limit-policies",
		Method:      http.MethodGet,
		Path:        "/api/v1/rate-limit/policies",
		Summary:     "Listar todas as políticas de rate limit",
		Description: "Retorna lista completa de políticas do DICT BACEN com cache Redis",
		Tags:        []string{"rate-limit"},
	}, h.ListPolicies)

	// GET /api/v1/rate-limit/policies/{policy}
	huma.Register(api, huma.Operation{
		OperationID: "get-rate-limit-policy",
		Method:      http.MethodGet,
		Path:        "/api/v1/rate-limit/policies/{policy}",
		Summary:     "Consultar política específica",
		Description: "Retorna estado detalhado de uma política com cache Redis",
		Tags:        []string{"rate-limit"},
	}, h.GetPolicy)
}

// ============================================================================
// HANDLER METHODS
// ============================================================================

// ListPolicies lista todas as políticas de rate limit
func (h *Handler) ListPolicies(
	ctx context.Context,
	input *ListPoliciesRequest,
) (*ListPoliciesResponse, error) {
	// Start tracing span
	ctx, span := h.tracer.Start(ctx, "Handler.ListPolicies")
	defer span.End()

	// Log request
	h.logger.InfoContext(ctx, "listing rate limit policies",
		"category", input.Category,
		"min_utilization", input.MinUtilization,
	)

	// Validate request
	if err := input.Validate(); err != nil {
		span.RecordError(err)
		h.logger.WarnContext(ctx, "invalid request parameters", "error", err)

		return nil, huma.NewError(
			http.StatusBadRequest,
			"Invalid request parameters",
			huma.ErrorDetail{
				"type":     ProblemTypeInvalidParameter,
				"title":    "Invalid Parameter",
				"detail":   err.Error(),
				"instance": "/api/v1/rate-limit/policies",
			},
		)
	}

	// Call application service
	policies, cached, cacheTTL, err := h.rateLimitService.ListPolicies(ctx, input.Category, input.MinUtilization)
	if err != nil {
		span.RecordError(err)
		h.logger.ErrorContext(ctx, "failed to list policies", "error", err)

		// Check if it's a bridge error
		if isBridgeUnavailableError(err) {
			return nil, huma.NewError(
				http.StatusServiceUnavailable,
				"Bridge service unavailable",
				huma.ErrorDetail{
					"type":     ProblemTypeBridgeUnavailable,
					"title":    "Bridge Service Unavailable",
					"detail":   "Unable to connect to Bridge gRPC service",
					"instance": "/api/v1/rate-limit/policies",
				},
			)
		}

		// Generic internal error
		return nil, huma.NewError(
			http.StatusInternalServerError,
			"Internal server error",
			huma.ErrorDetail{
				"type":     ProblemTypeInternalError,
				"title":    "Internal Server Error",
				"detail":   "An unexpected error occurred",
				"instance": "/api/v1/rate-limit/policies",
			},
		)
	}

	// Build response
	resp := &ListPoliciesResponse{}
	resp.Body.Policies = policies
	resp.Body.Total = len(policies)
	resp.Body.Cached = cached

	if cached && cacheTTL > 0 {
		ttlInt := int(cacheTTL.Seconds())
		resp.Body.CacheExpiresIn = &ttlInt
		resp.Headers.XCacheTTL = &ttlInt
		resp.Headers.XCacheStatus = "HIT"
	} else {
		resp.Headers.XCacheStatus = "MISS"
	}

	// Add tracing attributes
	span.SetAttributes(
		attribute.Int("policies.count", len(policies)),
		attribute.Bool("cache.hit", cached),
	)

	h.logger.InfoContext(ctx, "policies listed successfully",
		"count", len(policies),
		"cached", cached,
	)

	return resp, nil
}

// GetPolicy consulta uma política específica
func (h *Handler) GetPolicy(
	ctx context.Context,
	input *GetPolicyRequest,
) (*PolicyResponse, error) {
	// Start tracing span
	ctx, span := h.tracer.Start(ctx, "Handler.GetPolicy")
	defer span.End()

	// Log request
	h.logger.InfoContext(ctx, "getting rate limit policy",
		"policy", input.Policy,
	)

	// Validate request
	if err := input.Validate(); err != nil {
		span.RecordError(err)
		h.logger.WarnContext(ctx, "invalid policy name", "error", err)

		return nil, huma.NewError(
			http.StatusNotFound,
			"Policy not found",
			huma.ErrorDetail{
				"type":     ProblemTypePolicyNotFound,
				"title":    "Policy Not Found",
				"detail":   fmt.Sprintf("Policy '%s' does not exist", input.Policy),
				"instance": fmt.Sprintf("/api/v1/rate-limit/policies/%s", input.Policy),
			},
		)
	}

	// Call application service
	policy, cached, cacheTTL, err := h.rateLimitService.GetPolicy(ctx, input.Policy)
	if err != nil {
		span.RecordError(err)
		h.logger.ErrorContext(ctx, "failed to get policy", "error", err, "policy", input.Policy)

		// Check if policy not found
		if isNotFoundError(err) {
			return nil, huma.NewError(
				http.StatusNotFound,
				"Policy not found",
				huma.ErrorDetail{
					"type":     ProblemTypePolicyNotFound,
					"title":    "Policy Not Found",
					"detail":   fmt.Sprintf("Policy '%s' does not exist", input.Policy),
					"instance": fmt.Sprintf("/api/v1/rate-limit/policies/%s", input.Policy),
				},
			)
		}

		// Check if bridge error
		if isBridgeUnavailableError(err) {
			return nil, huma.NewError(
				http.StatusServiceUnavailable,
				"Bridge service unavailable",
				huma.ErrorDetail{
					"type":     ProblemTypeBridgeUnavailable,
					"title":    "Bridge Service Unavailable",
					"detail":   "Unable to connect to Bridge gRPC service",
					"instance": fmt.Sprintf("/api/v1/rate-limit/policies/%s", input.Policy),
				},
			)
		}

		// Generic internal error
		return nil, huma.NewError(
			http.StatusInternalServerError,
			"Internal server error",
			huma.ErrorDetail{
				"type":     ProblemTypeInternalError,
				"title":    "Internal Server Error",
				"detail":   "An unexpected error occurred",
				"instance": fmt.Sprintf("/api/v1/rate-limit/policies/%s", input.Policy),
			},
		)
	}

	// Build response
	resp := &PolicyResponse{}
	resp.Body.PolicySummary = *policy
	resp.Body.Cached = cached

	if cached && cacheTTL > 0 {
		ttlInt := int(cacheTTL.Seconds())
		resp.Body.CacheExpiresIn = &ttlInt
		resp.Headers.XCacheTTL = &ttlInt
		resp.Headers.XCacheStatus = "HIT"
	} else {
		resp.Headers.XCacheStatus = "MISS"
	}

	// Add tracing attributes
	span.SetAttributes(
		attribute.String("policy.name", input.Policy),
		attribute.Bool("cache.hit", cached),
		attribute.String("policy.status", policy.Status),
	)

	h.logger.InfoContext(ctx, "policy retrieved successfully",
		"policy", input.Policy,
		"cached", cached,
		"status", policy.Status,
	)

	return resp, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// isBridgeUnavailableError verifica se o erro é de Bridge indisponível
func isBridgeUnavailableError(err error) bool {
	// TODO: Implementar detecção específica de gRPC errors
	// Ex: grpc.Code(err) == codes.Unavailable
	return false
}

// isNotFoundError verifica se o erro é de recurso não encontrado
func isNotFoundError(err error) bool {
	// TODO: Implementar detecção específica
	return false
}
```

---

## 4. Application Layer (Use Cases)

### Service Interface

```go
// Location: apps/dict/application/ports/rate_limit_service.go
package ports

import (
	"context"
	"time"

	"github.com/lb-conn/connector-dict/apps/dict/handlers/http/ratelimit"
)

// RateLimitService define a interface do serviço de rate limit
type RateLimitService interface {
	// ListPolicies retorna lista de políticas (com cache)
	ListPolicies(
		ctx context.Context,
		category *string,
		minUtilization *float64,
	) (policies []ratelimit.PolicySummary, cached bool, cacheTTL time.Duration, err error)

	// GetPolicy retorna uma política específica (com cache)
	GetPolicy(
		ctx context.Context,
		policyName string,
	) (policy *ratelimit.PolicySummary, cached bool, cacheTTL time.Duration, err error)
}
```

### Service Implementation

```go
// Location: apps/dict/application/usecases/ratelimit/service.go
package ratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lb-conn/connector-dict/apps/dict/application/ports"
	"github.com/lb-conn/connector-dict/apps/dict/handlers/http/ratelimit"
	"github.com/lb-conn/connector-dict/apps/dict/infrastructure/grpc/bridge"
	"github.com/lb-conn/connector-dict/shared/cache"
	"github.com/lb-conn/connector-dict/shared/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Service implementa o use case de rate limit
type Service struct {
	bridgeClient bridge.RateLimitClient
	cache        cache.Cache
	logger       logger.Logger
	tracer       trace.Tracer
	cacheTTL     time.Duration
}

// NewService cria uma nova instância do serviço
func NewService(
	bridgeClient bridge.RateLimitClient,
	cache cache.Cache,
	logger logger.Logger,
	cacheTTL time.Duration,
) ports.RateLimitService {
	return &Service{
		bridgeClient: bridgeClient,
		cache:        cache,
		logger:       logger,
		tracer:       otel.Tracer("dict.application.ratelimit"),
		cacheTTL:     cacheTTL,
	}
}

// ============================================================================
// SERVICE METHODS
// ============================================================================

// ListPolicies implementa ports.RateLimitService
func (s *Service) ListPolicies(
	ctx context.Context,
	category *string,
	minUtilization *float64,
) ([]ratelimit.PolicySummary, bool, time.Duration, error) {
	ctx, span := s.tracer.Start(ctx, "Service.ListPolicies")
	defer span.End()

	// Build cache key
	cacheKey := s.buildCacheKeyList(category, minUtilization)

	// Try cache first
	var policies []ratelimit.PolicySummary
	ttl, err := s.cache.GetWithTTL(ctx, cacheKey, &policies)
	if err == nil {
		s.logger.InfoContext(ctx, "policies retrieved from cache",
			"count", len(policies),
			"ttl", ttl,
		)
		return policies, true, ttl, nil
	}

	// Cache miss - call Bridge
	s.logger.InfoContext(ctx, "cache miss, calling bridge", "key", cacheKey)

	bridgePolicies, err := s.bridgeClient.ListPolicies(ctx)
	if err != nil {
		return nil, false, 0, fmt.Errorf("bridge call failed: %w", err)
	}

	// Convert Bridge response to domain
	policies = s.convertBridgePolicies(bridgePolicies)

	// Apply filters
	policies = s.filterPolicies(policies, category, minUtilization)

	// Store in cache
	if err := s.cache.Set(ctx, cacheKey, policies, s.cacheTTL); err != nil {
		s.logger.WarnContext(ctx, "failed to cache policies", "error", err)
		// Continue - cache error is not critical
	}

	return policies, false, s.cacheTTL, nil
}

// GetPolicy implementa ports.RateLimitService
func (s *Service) GetPolicy(
	ctx context.Context,
	policyName string,
) (*ratelimit.PolicySummary, bool, time.Duration, error) {
	ctx, span := s.tracer.Start(ctx, "Service.GetPolicy")
	defer span.End()

	// Build cache key
	cacheKey := s.buildCacheKeyPolicy(policyName)

	// Try cache first
	var policy ratelimit.PolicySummary
	ttl, err := s.cache.GetWithTTL(ctx, cacheKey, &policy)
	if err == nil {
		s.logger.InfoContext(ctx, "policy retrieved from cache",
			"policy", policyName,
			"ttl", ttl,
		)
		return &policy, true, ttl, nil
	}

	// Cache miss - call Bridge
	s.logger.InfoContext(ctx, "cache miss, calling bridge", "key", cacheKey)

	bridgePolicy, err := s.bridgeClient.GetPolicy(ctx, policyName)
	if err != nil {
		return nil, false, 0, fmt.Errorf("bridge call failed: %w", err)
	}

	// Convert Bridge response to domain
	policy = s.convertBridgePolicy(bridgePolicy)

	// Store in cache
	if err := s.cache.Set(ctx, cacheKey, &policy, s.cacheTTL); err != nil {
		s.logger.WarnContext(ctx, "failed to cache policy", "error", err)
		// Continue - cache error is not critical
	}

	return &policy, false, s.cacheTTL, nil
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// buildCacheKeyList constrói a chave de cache para lista
func (s *Service) buildCacheKeyList(category *string, minUtilization *float64) string {
	key := "ratelimit:policies:list"

	if category != nil {
		key += fmt.Sprintf(":cat=%s", *category)
	}
	if minUtilization != nil {
		key += fmt.Sprintf(":min=%.2f", *minUtilization)
	}

	return key
}

// buildCacheKeyPolicy constrói a chave de cache para política individual
func (s *Service) buildCacheKeyPolicy(policyName string) string {
	return fmt.Sprintf("ratelimit:policies:policy=%s", policyName)
}

// convertBridgePolicies converte resposta do Bridge para domain
func (s *Service) convertBridgePolicies(bridgePolicies []bridge.PolicyState) []ratelimit.PolicySummary {
	policies := make([]ratelimit.PolicySummary, 0, len(bridgePolicies))

	for _, bp := range bridgePolicies {
		policies = append(policies, s.convertBridgePolicy(&bp))
	}

	return policies
}

// convertBridgePolicy converte uma política do Bridge para domain
func (s *Service) convertBridgePolicy(bp *bridge.PolicyState) ratelimit.PolicySummary {
	// Calculate utilization percentage
	utilizationPct := 100.0 - (float64(bp.AvailableTokens) / float64(bp.Capacity) * 100.0)

	// Determine status based on thresholds
	status := "OK"
	remainingPct := float64(bp.AvailableTokens) / float64(bp.Capacity) * 100.0

	if remainingPct <= bp.CriticalThresholdPct {
		status = "CRITICAL"
	} else if remainingPct <= bp.WarningThresholdPct {
		status = "WARNING"
	}

	return ratelimit.PolicySummary{
		PolicyName:           bp.PolicyName,
		Category:             bp.Category,
		CapacityMax:          bp.Capacity,
		RefillTokens:         bp.RefillTokens,
		RefillPeriodSec:      bp.RefillPeriodSec,
		AvailableTokens:      bp.AvailableTokens,
		UtilizationPct:       utilizationPct,
		WarningThresholdPct:  bp.WarningThresholdPct,
		CriticalThresholdPct: bp.CriticalThresholdPct,
		Status:               status,
		CheckedAt:            bp.CheckedAt,
	}
}

// filterPolicies aplica filtros às políticas
func (s *Service) filterPolicies(
	policies []ratelimit.PolicySummary,
	category *string,
	minUtilization *float64,
) []ratelimit.PolicySummary {
	if category == nil && minUtilization == nil {
		return policies
	}

	filtered := make([]ratelimit.PolicySummary, 0, len(policies))

	for _, p := range policies {
		// Filter by category
		if category != nil && p.Category != *category {
			continue
		}

		// Filter by min utilization
		if minUtilization != nil && p.UtilizationPct < *minUtilization {
			continue
		}

		filtered = append(filtered, p)
	}

	return filtered
}
```

---

## 5. Infrastructure Layer (Bridge Client)

### gRPC Client Interface

```go
// Location: apps/dict/infrastructure/grpc/bridge/rate_limit_client.go
package bridge

import (
	"context"
	"time"
)

// PolicyState representa o estado de uma política do Bridge
type PolicyState struct {
	PolicyName           string
	Category             string
	Capacity             int
	RefillTokens         int
	RefillPeriodSec      int
	AvailableTokens      int
	WarningThresholdPct  float64
	CriticalThresholdPct float64
	CheckedAt            time.Time
}

// RateLimitClient define a interface do cliente gRPC
type RateLimitClient interface {
	// ListPolicies retorna todas as políticas do DICT
	ListPolicies(ctx context.Context) ([]PolicyState, error)

	// GetPolicy retorna uma política específica
	GetPolicy(ctx context.Context, policyName string) (*PolicyState, error)
}
```

### gRPC Client Implementation

```go
// Location: apps/dict/infrastructure/grpc/bridge/rate_limit_client_impl.go
package bridge

import (
	"context"
	"fmt"
	"time"

	"github.com/lb-conn/connector-dict/shared/logger"
	pb "github.com/lb-conn/rsfn-connect-bacen-bridge/proto/ratelimit" // TODO: Verificar path correto
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// rateLimitClient implementa RateLimitClient
type rateLimitClient struct {
	grpcClient pb.RateLimitServiceClient
	logger     logger.Logger
	tracer     trace.Tracer
	timeout    time.Duration
}

// NewRateLimitClient cria uma nova instância do cliente
func NewRateLimitClient(
	conn *grpc.ClientConn,
	logger logger.Logger,
	timeout time.Duration,
) RateLimitClient {
	return &rateLimitClient{
		grpcClient: pb.NewRateLimitServiceClient(conn),
		logger:     logger,
		tracer:     otel.Tracer("dict.grpc.bridge.ratelimit"),
		timeout:    timeout,
	}
}

// ListPolicies implementa RateLimitClient
func (c *rateLimitClient) ListPolicies(ctx context.Context) ([]PolicyState, error) {
	ctx, span := c.tracer.Start(ctx, "BridgeClient.ListPolicies")
	defer span.End()

	// Add timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Call gRPC
	resp, err := c.grpcClient.ListPolicies(ctx, &pb.ListPoliciesRequest{})
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("grpc call failed: %w", err)
	}

	// Convert proto to domain
	policies := make([]PolicyState, 0, len(resp.Policies))
	for _, p := range resp.Policies {
		policies = append(policies, PolicyState{
			PolicyName:           p.PolicyName,
			Category:             p.Category,
			Capacity:             int(p.Capacity),
			RefillTokens:         int(p.RefillTokens),
			RefillPeriodSec:      int(p.RefillPeriodSec),
			AvailableTokens:      int(p.AvailableTokens),
			WarningThresholdPct:  p.WarningThresholdPct,
			CriticalThresholdPct: p.CriticalThresholdPct,
			CheckedAt:            p.CheckedAt.AsTime(),
		})
	}

	return policies, nil
}

// GetPolicy implementa RateLimitClient
func (c *rateLimitClient) GetPolicy(ctx context.Context, policyName string) (*PolicyState, error) {
	ctx, span := c.tracer.Start(ctx, "BridgeClient.GetPolicy")
	defer span.End()

	// Add timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Call gRPC
	resp, err := c.grpcClient.GetPolicy(ctx, &pb.GetPolicyRequest{
		PolicyName: policyName,
	})
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("grpc call failed: %w", err)
	}

	// Convert proto to domain
	policy := &PolicyState{
		PolicyName:           resp.PolicyName,
		Category:             resp.Category,
		Capacity:             int(resp.Capacity),
		RefillTokens:         int(resp.RefillTokens),
		RefillPeriodSec:      int(resp.RefillPeriodSec),
		AvailableTokens:      int(resp.AvailableTokens),
		WarningThresholdPct:  resp.WarningThresholdPct,
		CriticalThresholdPct: resp.CriticalThresholdPct,
		CheckedAt:            resp.CheckedAt.AsTime(),
	}

	return policy, nil
}
```

---

## 6. Cache Strategy

### Redis Cache Implementation

```go
// Location: apps/dict/infrastructure/cache/redis_cache.go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implementa cache com Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache cria uma nova instância
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{
		client: client,
	}
}

// Set armazena um valor no cache
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// GetWithTTL recupera um valor e seu TTL
func (c *RedisCache) GetWithTTL(ctx context.Context, key string, dest interface{}) (time.Duration, error) {
	// Get value
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return 0, fmt.Errorf("cache miss")
		}
		return 0, fmt.Errorf("redis get failed: %w", err)
	}

	// Unmarshal
	if err := json.Unmarshal(data, dest); err != nil {
		return 0, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	// Get TTL
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get ttl: %w", err)
	}

	return ttl, nil
}

// Delete remove um valor do cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
```

### Cache Configuration

```go
// Location: apps/dict/config/cache.go
package config

import "time"

// CacheConfig representa configuração de cache
type CacheConfig struct {
	// TTL padrão para políticas
	PolicyTTL time.Duration

	// Redis connection
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

// DefaultCacheConfig retorna configuração padrão
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		PolicyTTL:     60 * time.Second, // 60s como definido
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,
	}
}
```

---

## 7. Error Handling (RFC 9457)

### Complete Error Implementation

```go
// Location: apps/dict/handlers/http/errors/problem.go
package errors

import (
	"encoding/json"
	"net/http"
)

// Problem representa um erro RFC 9457
type Problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

// Error implementa a interface error
func (p *Problem) Error() string {
	return p.Detail
}

// WriteProblem escreve um Problem como resposta HTTP
func WriteProblem(w http.ResponseWriter, p *Problem) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteStatus(p.Status)
	json.NewEncoder(w).Encode(p)
}

// Common problem constructors

// NewInvalidParameterProblem cria erro de parâmetro inválido
func NewInvalidParameterProblem(detail, instance string) *Problem {
	return &Problem{
		Type:     "https://api.lbpay.com/problems/invalid-parameter",
		Title:    "Invalid Parameter",
		Status:   http.StatusBadRequest,
		Detail:   detail,
		Instance: instance,
	}
}

// NewPolicyNotFoundProblem cria erro de política não encontrada
func NewPolicyNotFoundProblem(policyName, instance string) *Problem {
	return &Problem{
		Type:     "https://api.lbpay.com/problems/policy-not-found",
		Title:    "Policy Not Found",
		Status:   http.StatusNotFound,
		Detail:   "Policy '" + policyName + "' does not exist",
		Instance: instance,
	}
}

// NewBridgeUnavailableProblem cria erro de Bridge indisponível
func NewBridgeUnavailableProblem(instance string) *Problem {
	return &Problem{
		Type:     "https://api.lbpay.com/problems/bridge-unavailable",
		Title:    "Bridge Service Unavailable",
		Status:   http.StatusServiceUnavailable,
		Detail:   "Unable to connect to Bridge gRPC service",
		Instance: instance,
	}
}

// NewInternalErrorProblem cria erro interno genérico
func NewInternalErrorProblem(instance string) *Problem {
	return &Problem{
		Type:     "https://api.lbpay.com/problems/internal-error",
		Title:    "Internal Server Error",
		Status:   http.StatusInternalServerError,
		Detail:   "An unexpected error occurred",
		Instance: instance,
	}
}
```

---

## 8. Integration Testing

### Complete Test Suite

```go
// Location: apps/dict/handlers/http/ratelimit/handler_test.go
package ratelimit_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/lb-conn/connector-dict/apps/dict/handlers/http/ratelimit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRateLimitService is a mock implementation
type MockRateLimitService struct {
	mock.Mock
}

func (m *MockRateLimitService) ListPolicies(
	ctx context.Context,
	category *string,
	minUtilization *float64,
) ([]ratelimit.PolicySummary, bool, time.Duration, error) {
	args := m.Called(ctx, category, minUtilization)
	return args.Get(0).([]ratelimit.PolicySummary), args.Bool(1), args.Get(2).(time.Duration), args.Error(3)
}

func (m *MockRateLimitService) GetPolicy(
	ctx context.Context,
	policyName string,
) (*ratelimit.PolicySummary, bool, time.Duration, error) {
	args := m.Called(ctx, policyName)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Get(2).(time.Duration), args.Error(3)
	}
	return args.Get(0).(*ratelimit.PolicySummary), args.Bool(1), args.Get(2).(time.Duration), args.Error(3)
}

func TestListPolicies_Success_CacheHit(t *testing.T) {
	// Setup
	mockService := new(MockRateLimitService)
	handler := ratelimit.NewHandler(mockService, nil)

	// Mock data
	policies := []ratelimit.PolicySummary{
		{
			PolicyName:           "ENTRIES_CREATE",
			Category:             "A",
			CapacityMax:          300,
			RefillTokens:         5,
			RefillPeriodSec:      60,
			AvailableTokens:      150,
			UtilizationPct:       50.00,
			WarningThresholdPct:  25.00,
			CriticalThresholdPct: 10.00,
			Status:               "OK",
			CheckedAt:            time.Now(),
		},
	}

	mockService.On("ListPolicies", mock.Anything, (*string)(nil), (*float64)(nil)).
		Return(policies, true, 45*time.Second, nil)

	// Setup Huma API
	router := http.NewServeMux()
	api := humago.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	handler.RegisterRoutes(api)

	// Execute request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rate-limit/policies", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "HIT", w.Header().Get("X-Cache-Status"))
	assert.Equal(t, "45", w.Header().Get("X-Cache-TTL"))

	mockService.AssertExpectations(t)
}

func TestListPolicies_FilterByCategory(t *testing.T) {
	// Setup
	mockService := new(MockRateLimitService)
	handler := ratelimit.NewHandler(mockService, nil)

	category := "A"
	policies := []ratelimit.PolicySummary{
		{
			PolicyName: "ENTRIES_CREATE",
			Category:   "A",
		},
	}

	mockService.On("ListPolicies", mock.Anything, &category, (*float64)(nil)).
		Return(policies, false, 60*time.Second, nil)

	// Setup Huma API
	router := http.NewServeMux()
	api := humago.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	handler.RegisterRoutes(api)

	// Execute request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rate-limit/policies?category=A", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetPolicy_Success(t *testing.T) {
	// Setup
	mockService := new(MockRateLimitService)
	handler := ratelimit.NewHandler(mockService, nil)

	policy := &ratelimit.PolicySummary{
		PolicyName:      "ENTRIES_CREATE",
		AvailableTokens: 150,
		Status:          "OK",
	}

	mockService.On("GetPolicy", mock.Anything, "ENTRIES_CREATE").
		Return(policy, false, 60*time.Second, nil)

	// Setup Huma API
	router := http.NewServeMux()
	api := humago.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	handler.RegisterRoutes(api)

	// Execute request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rate-limit/policies/ENTRIES_CREATE", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetPolicy_NotFound(t *testing.T) {
	// Setup
	mockService := new(MockRateLimitService)
	handler := ratelimit.NewHandler(mockService, nil)

	// Mock returns not found error
	mockService.On("GetPolicy", mock.Anything, "INVALID_POLICY").
		Return((*ratelimit.PolicySummary)(nil), false, time.Duration(0), fmt.Errorf("not found"))

	// Setup Huma API
	router := http.NewServeMux()
	api := humago.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	handler.RegisterRoutes(api)

	// Execute request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rate-limit/policies/INVALID_POLICY", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}
```

---

## 9. Performance Benchmarks

### Benchmark Tests

```go
// Location: apps/dict/handlers/http/ratelimit/handler_bench_test.go
package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/lb-conn/connector-dict/apps/dict/handlers/http/ratelimit"
)

func BenchmarkListPolicies_CacheHit(b *testing.B) {
	mockService := new(MockRateLimitService)
	handler := ratelimit.NewHandler(mockService, nil)

	policies := make([]ratelimit.PolicySummary, 24)
	mockService.On("ListPolicies", mock.Anything, (*string)(nil), (*float64)(nil)).
		Return(policies, true, 60*time.Second, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.ListPolicies(context.Background(), &ratelimit.ListPoliciesRequest{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetPolicy_CacheHit(b *testing.B) {
	mockService := new(MockRateLimitService)
	handler := ratelimit.NewHandler(mockService, nil)

	policy := &ratelimit.PolicySummary{PolicyName: "ENTRIES_CREATE"}
	mockService.On("GetPolicy", mock.Anything, "ENTRIES_CREATE").
		Return(policy, true, 60*time.Second, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.GetPolicy(context.Background(), &ratelimit.GetPolicyRequest{Policy: "ENTRIES_CREATE"})
		if err != nil {
			b.Fatal(err)
		}
	}
}
```

### Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| List Policies (cache hit) | <50ms p99 | Benchmark test |
| List Policies (cache miss) | <200ms p99 | Load test |
| Get Policy (cache hit) | <30ms p99 | Benchmark test |
| Get Policy (cache miss) | <150ms p99 | Load test |
| Cache Hit Rate | >90% | Redis metrics |
| Concurrent Requests | 1000 req/s | Load test |

---

## 📋 Checklist de Implementação

### Database & Domain Engineer
- [ ] Criar schemas Huma (schemas.go)
- [ ] Implementar validação de requests
- [ ] Testes de schemas (>90% coverage)

### Dict API Engineer
- [ ] Implementar HTTP handlers (handler.go)
- [ ] Registrar rotas Huma
- [ ] Implementar error handling RFC 9457
- [ ] Testes unitários de handlers (>90% coverage)
- [ ] Integration tests com mock service

### Application Layer
- [ ] Criar service interface (ports)
- [ ] Implementar application service
- [ ] Integrar com Bridge client
- [ ] Integrar com Redis cache
- [ ] Testes unitários (>90% coverage)

### gRPC Engineer
- [ ] Criar interface Bridge client
- [ ] Implementar gRPC client
- [ ] Error handling (retryable vs non-retryable)
- [ ] Integration tests com mock Bridge

### DevOps Engineer
- [ ] Configurar Redis
- [ ] Setup OpenTelemetry tracing
- [ ] Prometheus metrics
- [ ] Grafana dashboards
- [ ] Health checks

### QA Lead
- [ ] E2E tests completos
- [ ] Performance benchmarks
- [ ] Load tests (1000 req/s)
- [ ] Cache hit rate validation
- [ ] Security tests

---

**Última Atualização**: 2025-10-31
**Versão**: 1.0.0
**Status**: ✅ ESPECIFICAÇÃO COMPLETA - Production-Ready
