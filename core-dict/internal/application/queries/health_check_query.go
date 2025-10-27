package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/repositories"
)

// HealthCheckQuery representa a query para health check completo
type HealthCheckQuery struct {
	// Pode adicionar flags para checks específicos no futuro
}

// HealthCheckQueryHandler lida com a query HealthCheck
type HealthCheckQueryHandler struct {
	healthRepo    repositories.HealthRepository
	connectClient interface{ HealthCheck(ctx context.Context) error } // NEW: Connect health check
	startTime     time.Time                                            // Para calcular uptime
}

// NewHealthCheckQueryHandler cria um novo handler para HealthCheck
func NewHealthCheckQueryHandler(
	healthRepo repositories.HealthRepository,
	connectClient interface{ HealthCheck(ctx context.Context) error }, // Optional: can be nil
) *HealthCheckQueryHandler {
	return &HealthCheckQueryHandler{
		healthRepo:    healthRepo,
		connectClient: connectClient,
		startTime:     time.Now(),
	}
}

// Handle executa o health check completo
// IMPORTANTE: Health checks NÃO usam cache (precisam ser real-time)
func (h *HealthCheckQueryHandler) Handle(ctx context.Context, query HealthCheckQuery) (*entities.HealthStatus, error) {
	health := &entities.HealthStatus{
		Status:       "healthy",
		Version:      "v0.1.0", // TODO: Pegar de variável de ambiente
		Uptime:       int64(time.Since(h.startTime).Seconds()),
		Dependencies: make(map[string]interface{}),
		Timestamp:    time.Now(),
	}

	// 1. Check PostgreSQL
	dbStatus := "healthy"
	dbStart := time.Now()
	if err := h.healthRepo.CheckDatabase(ctx); err != nil {
		dbStatus = "unhealthy"
		health.Status = "degraded"
		health.Dependencies["database_error"] = err.Error()
	}
	dbLatency := time.Since(dbStart).Milliseconds()
	health.DatabaseStatus = dbStatus
	health.Dependencies["database"] = map[string]interface{}{
		"status":      dbStatus,
		"latency_ms":  dbLatency,
		"checked_at":  time.Now().Unix(),
	}

	// 2. Check Redis
	redisStatus := "healthy"
	redisStart := time.Now()
	if err := h.healthRepo.CheckRedis(ctx); err != nil {
		redisStatus = "unhealthy"
		health.Status = "degraded"
		health.Dependencies["redis_error"] = err.Error()
	}
	redisLatency := time.Since(redisStart).Milliseconds()
	health.RedisStatus = redisStatus
	health.Dependencies["redis"] = map[string]interface{}{
		"status":     redisStatus,
		"latency_ms": redisLatency,
		"checked_at": time.Now().Unix(),
	}

	// 3. Check Pulsar
	pulsarStatus := "healthy"
	pulsarStart := time.Now()
	if err := h.healthRepo.CheckPulsar(ctx); err != nil {
		pulsarStatus = "unhealthy"
		health.Status = "degraded"
		health.Dependencies["pulsar_error"] = err.Error()
	}
	pulsarLatency := time.Since(pulsarStart).Milliseconds()
	health.PulsarStatus = pulsarStatus
	health.Dependencies["pulsar"] = map[string]interface{}{
		"status":     pulsarStatus,
		"latency_ms": pulsarLatency,
		"checked_at": time.Now().Unix(),
	}

	// 4. Check Connect service (gRPC to conn-dict)
	connectStatus := "healthy"
	connectStart := time.Now()
	if h.connectClient != nil {
		if err := h.connectClient.HealthCheck(ctx); err != nil {
			connectStatus = "unhealthy"
			health.Status = "degraded"
			health.Dependencies["connect_error"] = err.Error()
		}
	} else {
		connectStatus = "not_configured"
	}
	connectLatency := time.Since(connectStart).Milliseconds()
	health.Dependencies["connect"] = map[string]interface{}{
		"status":     connectStatus,
		"latency_ms": connectLatency,
		"checked_at": time.Now().Unix(),
	}

	// 5. Overall status
	if health.DatabaseStatus == "unhealthy" {
		// PostgreSQL is critical - mark as unhealthy
		health.Status = "unhealthy"
	} else if health.RedisStatus == "unhealthy" || health.PulsarStatus == "unhealthy" || connectStatus == "unhealthy" {
		// Redis, Pulsar, or Connect down - mark as degraded
		health.Status = "degraded"
	}

	return health, nil
}

// QuickCheck executa um health check rápido (apenas database)
func (h *HealthCheckQueryHandler) QuickCheck(ctx context.Context) error {
	if err := h.healthRepo.CheckDatabase(ctx); err != nil {
		return fmt.Errorf("database unhealthy: %w", err)
	}
	return nil
}
