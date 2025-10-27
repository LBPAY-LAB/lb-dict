package repositories

import (
	"context"
)

// HealthRepository interface para health checks
type HealthRepository interface {
	// CheckDatabase verifica conectividade com PostgreSQL
	CheckDatabase(ctx context.Context) error

	// CheckRedis verifica conectividade com Redis
	CheckRedis(ctx context.Context) error

	// CheckPulsar verifica conectividade com Pulsar
	CheckPulsar(ctx context.Context) error
}
