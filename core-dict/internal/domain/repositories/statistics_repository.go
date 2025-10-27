package repositories

import (
	"context"

	"github.com/lbpay-lab/core-dict/internal/domain/entities"
)

// StatisticsRepository interface para estatísticas agregadas
type StatisticsRepository interface {
	// GetStatistics retorna estatísticas agregadas do sistema
	GetStatistics(ctx context.Context) (*entities.Statistics, error)
}
