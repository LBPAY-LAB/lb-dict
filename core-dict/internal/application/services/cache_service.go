package services

import (
	"context"
	"time"
)

// CacheService interface para operações de cache Redis
type CacheService interface {
	// Get recupera um valor do cache
	Get(ctx context.Context, key string) (interface{}, error)

	// Set armazena um valor no cache com TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete remove um valor do cache
	Delete(ctx context.Context, key string) error

	// Exists verifica se uma chave existe no cache
	Exists(ctx context.Context, key string) (bool, error)

	// Invalidate invalida múltiplas chaves (pattern matching)
	Invalidate(ctx context.Context, pattern string) error
}

// ConnectService interface para chamadas ao Connect service
type ConnectService interface {
	// VerifyAccount verifica uma conta no RSFN via Connect
	VerifyAccount(ctx context.Context, ispb, branch, accountNumber string) (bool, error)
}

// ConnectClient interface for optional Connect service integration
type ConnectClient interface {
	// GetEntryByKey retrieves entry from RSFN by key value
	GetEntryByKey(ctx context.Context, keyValue string) (interface{}, error)
	// HealthCheck checks if Connect service is reachable
	HealthCheck(ctx context.Context) error
}
