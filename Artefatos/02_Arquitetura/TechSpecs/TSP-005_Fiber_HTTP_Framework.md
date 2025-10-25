# TSP-005: Fiber HTTP Framework - Technical Specification

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: Fiber HTTP Framework (v3)
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: BACKEND (AI Agent - Backend Specialist)
**Revisor**: [Aguardando]
**Aprovador**: Tech Lead, Head de Arquitetura

---

## Sumário Executivo

Este documento especifica a implementação do **Fiber HTTP Framework v3** para o projeto DICT LBPay, cobrindo configuração do servidor, middleware chain (logging, auth, CORS, rate limiting, recovery), error handling, request validation, e performance tuning.

**Baseado em**:
- [TEC-003: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [API-001: Core DICT REST API](../../04_APIs/REST/API-001_Core_DICT_REST_API.md)
- [ADR-002: HTTP Framework Selection](../ADR-002_HTTP_Framework_Selection.md) (pendente)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | BACKEND | Versão inicial - Fiber v3 specification |

---

## Índice

1. [Visão Geral](#1-visão-geral)
2. [Server Configuration](#2-server-configuration)
3. [Middleware Chain](#3-middleware-chain)
4. [Error Handling](#4-error-handling)
5. [Request Validation](#5-request-validation)
6. [Response Formatting](#6-response-formatting)
7. [Performance Tuning](#7-performance-tuning)
8. [Security](#8-security)
9. [Monitoring & Observability](#9-monitoring--observability)

---

## 1. Visão Geral

### 1.1. Fiber Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Fiber v3 HTTP Server                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  fasthttp Server (underlying engine)                        │ │
│  │  - Port: 8080 (HTTP)                                       │ │
│  │  - Zero-allocation routing                                 │ │
│  │  - High throughput (1M+ req/s)                             │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           ↓                                       │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Middleware Chain                                          │ │
│  │  1. Logger (request/response logging)                      │ │
│  │  2. Recovery (panic recovery)                              │ │
│  │  3. CORS (cross-origin resource sharing)                   │ │
│  │  4. RequestID (X-Request-ID header)                        │ │
│  │  5. RateLimiter (per-IP rate limiting)                     │ │
│  │  6. Authentication (JWT validation)                        │ │
│  │  7. Authorization (RBAC/ISPB check)                        │ │
│  │  8. Compression (gzip/brotli)                              │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           ↓                                       │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Router                                                     │ │
│  │  - /api/v1/entries (CRUD operations)                       │ │
│  │  - /api/v1/claims (claim management)                       │ │
│  │  - /health (health check)                                  │ │
│  │  - /metrics (Prometheus metrics)                           │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           ↓                                       │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Handlers (Business Logic)                                 │ │
│  │  - EntryHandler (CreateEntry, GetEntry, etc.)              │ │
│  │  - ClaimHandler (CreateClaim, CancelClaim, etc.)           │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2. Key Features

| Feature | Value | Justification |
|---------|-------|---------------|
| **Fiber Version** | v3.0.0-beta.3 | Latest (Express-like API) |
| **Underlying Engine** | fasthttp | 10x faster than net/http |
| **Routing** | Zero-allocation | Low GC pressure |
| **Middleware** | Composable chain | Modular, reusable |
| **Validation** | go-playground/validator | Industry standard |
| **Serialization** | encoding/json | Native Go |
| **Compression** | gzip + brotli | Reduce bandwidth |
| **Max Body Size** | 10MB | Prevent DoS |
| **Read Timeout** | 10s | Prevent slow clients |
| **Write Timeout** | 10s | Prevent slow responses |

---

## 2. Server Configuration

### 2.1. Fiber App Initialization

```go
// cmd/api/main.go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/lb-conn/rsfn-connect/internal/config"
	"github.com/lb-conn/rsfn-connect/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:               "DICT Core API v1.0",
		ServerHeader:          "DICT",
		StrictRouting:         true,
		CaseSensitive:         true,
		DisableStartupMessage: false,
		EnablePrintRoutes:     true,
		BodyLimit:             10 * 1024 * 1024, // 10MB
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		IdleTimeout:           120 * time.Second,
		ReadBufferSize:        8192,
		WriteBufferSize:       8192,
		CompressedFileSuffix:  ".gz",
		Concurrency:           256 * 1024, // 256K concurrent connections
		ErrorHandler:          server.ErrorHandler,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
	})

	// Setup server (middleware, routes, etc.)
	server.Setup(app, cfg)

	// Start server in goroutine
	go func() {
		if err := app.Listen(cfg.Server.Address); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	log.Printf("Server started on %s", cfg.Server.Address)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
```

### 2.2. Configuration Structure

```go
// internal/config/config.go
package config

import (
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Address         string        // ":8080"
	ReadTimeout     time.Duration // 10s
	WriteTimeout    time.Duration // 10s
	ShutdownTimeout time.Duration // 30s
	Environment     string        // "production", "staging", "development"
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration // 24h
	Issuer     string        // "dict-api"
}

func Load() (*Config, error) {
	// Load from environment variables, config file, etc.
	// Implementation details omitted
}
```

---

## 3. Middleware Chain

### 3.1. Middleware Order

**Execution Order** (top to bottom):

1. **Logger**: Log all requests/responses
2. **Recovery**: Recover from panics
3. **CORS**: Handle CORS headers
4. **RequestID**: Generate unique request ID
5. **RateLimiter**: Enforce rate limits
6. **Authentication**: Validate JWT token
7. **Authorization**: Check ISPB/role permissions
8. **Compression**: Compress responses

### 3.2. Logger Middleware

```go
// internal/middleware/logger.go
package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func Logger() fiber.Handler {
	return logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${latency} ${method} ${path} | ${ip} | ${ua}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "America/Sao_Paulo",
		Output:     os.Stdout,
		Next: func(c fiber.Ctx) bool {
			// Skip logging for health checks
			return c.Path() == "/health" || c.Path() == "/metrics"
		},
	})
}
```

### 3.3. Recovery Middleware

```go
// internal/middleware/recovery.go
package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"log"
)

func Recovery() fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c fiber.Ctx, e interface{}) {
			log.Printf("PANIC: %v\n", e)
			log.Printf("Request: %s %s", c.Method(), c.Path())
		},
	})
}
```

### 3.4. CORS Middleware

```go
// internal/middleware/cors.go
package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func CORS() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins: []string{
			"https://dict.lbpay.com.br",
			"https://admin.lbpay.com.br",
		},
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Request-ID",
		},
		ExposeHeaders: []string{
			"X-Request-ID",
			"X-Total-Count",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	})
}
```

### 3.5. RequestID Middleware

```go
// internal/middleware/requestid.go
package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func RequestID() fiber.Handler {
	return requestid.New(requestid.Config{
		Header: "X-Request-ID",
		Generator: func() string {
			return uuid.New().String()
		},
	})
}
```

### 3.6. Rate Limiter Middleware

```go
// internal/middleware/ratelimiter.go
package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/storage/redis/v3"
)

func RateLimiter(redisClient *redis.Storage) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100,              // 100 requests
		Expiration: 1 * time.Minute,  // per minute
		KeyGenerator: func(c fiber.Ctx) string {
			// Rate limit by IP
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
				"message": "You have sent too many requests. Please try again later.",
			})
		},
		Storage: redisClient, // Use Redis for distributed rate limiting
	})
}
```

### 3.7. Authentication Middleware

```go
// internal/middleware/auth.go
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lb-conn/rsfn-connect/internal/config"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	ISPB   string `json:"ispb"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func Authentication(cfg *config.JWTConfig) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		// Bearer <token>
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		tokenString := parts[1]

		// Parse and validate JWT
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Extract claims
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		// Store claims in context
		c.Locals("user_id", claims.UserID)
		c.Locals("ispb", claims.ISPB)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}
```

### 3.8. Authorization Middleware

```go
// internal/middleware/authz.go
package middleware

import (
	"github.com/gofiber/fiber/v3"
)

func RequireISPB() fiber.Handler {
	return func(c fiber.Ctx) error {
		ispb, ok := c.Locals("ispb").(string)
		if !ok || ispb == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "ISPB not found in token",
			})
		}
		return c.Next()
	}
}

func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Role not found in token",
			})
		}

		// Check if user has allowed role
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
		})
	}
}
```

### 3.9. Compression Middleware

```go
// internal/middleware/compression.go
package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
)

func Compression() fiber.Handler {
	return compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // Trade-off: speed vs compression ratio
	})
}
```

---

## 4. Error Handling

### 4.1. Custom Error Handler

```go
// internal/server/errors.go
package server

import (
	"github.com/gofiber/fiber/v3"
	"github.com/lb-conn/rsfn-connect/internal/domain/errors"
	"log"
)

func ErrorHandler(c fiber.Ctx, err error) error {
	// Log error
	log.Printf("ERROR: %v | Request: %s %s", err, c.Method(), c.Path())

	// Default 500 status
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"
	errorType := "INTERNAL_ERROR"

	// Check for custom domain errors
	switch e := err.(type) {
	case *errors.ValidationError:
		code = fiber.StatusBadRequest
		message = e.Message
		errorType = "VALIDATION_ERROR"
	case *errors.NotFoundError:
		code = fiber.StatusNotFound
		message = e.Message
		errorType = "NOT_FOUND"
	case *errors.ConflictError:
		code = fiber.StatusConflict
		message = e.Message
		errorType = "CONFLICT"
	case *errors.UnauthorizedError:
		code = fiber.StatusUnauthorized
		message = e.Message
		errorType = "UNAUTHORIZED"
	case *errors.ForbiddenError:
		code = fiber.StatusForbidden
		message = e.Message
		errorType = "FORBIDDEN"
	case *errors.BridgeError:
		code = fiber.StatusBadGateway
		message = "Bridge communication error"
		errorType = "BRIDGE_ERROR"
	case *fiber.Error:
		code = e.Code
		message = e.Message
		errorType = "HTTP_ERROR"
	}

	// Return JSON error response
	return c.Status(code).JSON(fiber.Map{
		"error":      errorType,
		"message":    message,
		"request_id": c.Locals("requestid"),
		"timestamp":  time.Now().Format(time.RFC3339),
	})
}
```

### 4.2. Domain Errors

```go
// internal/domain/errors/errors.go
package errors

import "fmt"

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("Validation error on field '%s': %s", e.Field, e.Message)
}

type NotFoundError struct {
	Entity string
	ID     string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID '%s' not found", e.Entity, e.ID)
}

type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	return e.Message
}

type UnauthorizedError struct {
	Message string
}

func (e *UnauthorizedError) Error() string {
	return e.Message
}

type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string {
	return e.Message
}

type BridgeError struct {
	Operation string
	Err       error
}

func (e *BridgeError) Error() string {
	return fmt.Sprintf("Bridge error during %s: %v", e.Operation, e.Err)
}
```

---

## 5. Request Validation

### 5.1. Validator Setup

```go
// internal/validator/validator.go
package validator

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom validators
	validate.RegisterValidation("cpf", validateCPF)
	validate.RegisterValidation("cnpj", validateCNPJ)
	validate.RegisterValidation("ispb", validateISPB)
	validate.RegisterValidation("dict_key_type", validateKeyType)
}

func Validate(data interface{}) error {
	return validate.Struct(data)
}

// Custom validators
func validateCPF(fl validator.FieldLevel) bool {
	cpf := fl.Field().String()
	// CPF validation logic (11 digits + checksum)
	return len(cpf) == 11 && isValidCPF(cpf)
}

func validateCNPJ(fl validator.FieldLevel) bool {
	cnpj := fl.Field().String()
	// CNPJ validation logic (14 digits + checksum)
	return len(cnpj) == 14 && isValidCNPJ(cnpj)
}

func validateISPB(fl validator.FieldLevel) bool {
	ispb := fl.Field().String()
	// ISPB: 8 digits
	return len(ispb) == 8 && isNumeric(ispb)
}

func validateKeyType(fl validator.FieldLevel) bool {
	keyType := fl.Field().String()
	validTypes := []string{"CPF", "CNPJ", "EMAIL", "PHONE", "EVP"}
	for _, valid := range validTypes {
		if keyType == valid {
			return true
		}
	}
	return false
}
```

### 5.2. Request DTOs

```go
// internal/dto/entry.go
package dto

type CreateEntryRequest struct {
	KeyType       string `json:"key_type" validate:"required,dict_key_type"`
	KeyValue      string `json:"key_value" validate:"required,min=1,max=100"`
	AccountISPB   string `json:"account_ispb" validate:"required,ispb"`
	AccountNumber string `json:"account_number" validate:"required,max=20"`
	AccountType   string `json:"account_type" validate:"required,oneof=CHECKING SAVINGS PAYMENT"`
	OwnerName     string `json:"owner_name" validate:"required,min=3,max=200"`
	OwnerDocument string `json:"owner_document" validate:"required"`
}

func (r *CreateEntryRequest) Validate() error {
	return validator.Validate(r)
}
```

### 5.3. Handler with Validation

```go
// internal/handlers/entry_handler.go
package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/lb-conn/rsfn-connect/internal/dto"
	"github.com/lb-conn/rsfn-connect/internal/domain/errors"
)

type EntryHandler struct {
	service EntryService
}

func (h *EntryHandler) CreateEntry(c fiber.Ctx) error {
	var req dto.CreateEntryRequest

	// Parse body
	if err := c.Bind().JSON(&req); err != nil {
		return &errors.ValidationError{
			Field:   "body",
			Message: "Invalid JSON body",
		}
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return &errors.ValidationError{
			Field:   "body",
			Message: err.Error(),
		}
	}

	// Extract ISPB from JWT
	ispb := c.Locals("ispb").(string)

	// Call service
	entry, err := h.service.CreateEntry(c.Context(), &req, ispb)
	if err != nil {
		return err
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(entry)
}
```

---

## 6. Response Formatting

### 6.1. Success Response

```go
// internal/dto/response.go
package dto

type SuccessResponse struct {
	Data      interface{} `json:"data"`
	Message   string      `json:"message,omitempty"`
	RequestID string      `json:"request_id"`
	Timestamp string      `json:"timestamp"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	TotalCount int         `json:"total_count"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	RequestID  string      `json:"request_id"`
	Timestamp  string      `json:"timestamp"`
}
```

### 6.2. Error Response (Already defined in Error Handler)

```json
{
  "error": "VALIDATION_ERROR",
  "message": "Invalid CPF format",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-10-25T10:30:00Z"
}
```

---

## 7. Performance Tuning

### 7.1. Connection Pooling

**Fiber Config**:

```go
fiber.Config{
	Concurrency: 256 * 1024, // 256K concurrent connections
	ReadBufferSize: 8192,    // 8KB read buffer
	WriteBufferSize: 8192,   // 8KB write buffer
}
```

### 7.2. JSON Encoding Optimization

**Use sonic (faster JSON encoder)**:

```go
import "github.com/bytedance/sonic"

fiber.Config{
	JSONEncoder: sonic.Marshal,
	JSONDecoder: sonic.Unmarshal,
}
```

**Benchmark**: sonic is 2-3x faster than encoding/json

### 7.3. Preforking (Multi-process)

**Enable for production**:

```go
app := fiber.New(fiber.Config{
	Prefork: true, // Use SO_REUSEPORT (Linux only)
})
```

**Effect**: Spawn multiple server processes (= CPU cores)

### 7.4. HTTP/2 Support

**Enable HTTP/2**:

```go
// Use TLS with HTTP/2
app.ListenTLS(":443", "./cert.pem", "./key.pem")
```

---

## 8. Security

### 8.1. Helmet Middleware (Security Headers)

```go
// internal/middleware/helmet.go
package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/helmet"
)

func Helmet() fiber.Handler {
	return helmet.New(helmet.Config{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,
		ContentSecurityPolicy: "default-src 'self'",
		ReferrerPolicy:        "no-referrer",
	})
}
```

### 8.2. CSRF Protection

```go
// internal/middleware/csrf.go
package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
)

func CSRF() fiber.Handler {
	return csrf.New(csrf.Config{
		KeyLookup:      "header:X-CSRF-Token",
		CookieName:     "csrf_token",
		CookieSameSite: "Strict",
		Expiration:     1 * time.Hour,
	})
}
```

### 8.3. Input Sanitization

**Prevent SQL injection, XSS**:

```go
import "github.com/microcosm-cc/bluemonday"

func sanitizeInput(input string) string {
	p := bluemonday.StrictPolicy()
	return p.Sanitize(input)
}
```

---

## 9. Monitoring & Observability

### 9.1. Prometheus Metrics

```go
// internal/middleware/metrics.go
package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/monitor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func Metrics() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := c.Response().StatusCode()

		httpRequestsTotal.WithLabelValues(
			c.Method(),
			c.Path(),
			fmt.Sprintf("%d", status),
		).Inc()

		httpRequestDuration.WithLabelValues(
			c.Method(),
			c.Path(),
		).Observe(duration)

		return err
	}
}
```

### 9.2. Health Check Endpoint

```go
// internal/handlers/health.go
package handlers

import (
	"github.com/gofiber/fiber/v3"
)

func HealthCheck(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "healthy",
		"version": "1.0.0",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
```

### 9.3. Metrics Endpoint

```go
// cmd/api/routes.go
import (
	"github.com/gofiber/contrib/fiberprometheus/v2"
)

func setupRoutes(app *fiber.App) {
	// Prometheus metrics
	prometheus := fiberprometheus.New("dict-api")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	// Health check
	app.Get("/health", handlers.HealthCheck)

	// API routes
	api := app.Group("/api/v1")
	// ... (register handlers)
}
```

---

## Rastreabilidade

### Requisitos Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RF-TSP-005-001 | Middleware chain (logger, auth, etc.) | Best Practices | ✅ Especificado |
| RF-TSP-005-002 | JWT authentication | Security requirement | ✅ Especificado |
| RF-TSP-005-003 | Request validation (go-playground) | Data integrity | ✅ Especificado |
| RF-TSP-005-004 | Error handling (domain errors) | API consistency | ✅ Especificado |
| RF-TSP-005-005 | Rate limiting (per-IP) | DoS protection | ✅ Especificado |

### Requisitos Não-Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RNF-TSP-005-001 | Performance: < 50ms p95 latency | Performance goal | ✅ Especificado |
| RNF-TSP-005-002 | Throughput: > 10K req/s | Performance goal | ✅ Especificado |
| RNF-TSP-005-003 | Security headers (Helmet) | Security requirement | ✅ Especificado |
| RNF-TSP-005-004 | Prometheus metrics | Observability | ✅ Especificado |

---

## Próximas Revisões

**Pendências**:
- [ ] Implementar OpenAPI/Swagger documentation
- [ ] Configurar distributed tracing (OpenTelemetry)
- [ ] Implementar circuit breaker (gobreaker)
- [ ] Adicionar API versioning strategy
- [ ] Criar integration tests (testify)
- [ ] Implementar GraphQL endpoint (optional)

---

**Referências**:
- [Fiber v3 Documentation](https://docs.gofiber.io/)
- [go-playground/validator](https://github.com/go-playground/validator)
- [Prometheus Go Client](https://prometheus.io/docs/guides/go-application/)
- [OWASP API Security](https://owasp.org/www-project-api-security/)
