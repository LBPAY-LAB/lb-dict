package repositories

import (
	"context"

	"github.com/lbpay-lab/core-dict/internal/domain/entities"
)

// InfractionRepository interface para operações com infrações
type InfractionRepository interface {
	// List lista infrações com paginação
	List(ctx context.Context, ispb string, limit, offset int) ([]*entities.Infraction, error)

	// CountByISPB conta infrações de um participante
	CountByISPB(ctx context.Context, ispb string) (int64, error)
}
