package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
)

// InfractionRepository interface para operações com infrações
type InfractionRepository interface {
	// Create cria uma nova infração
	Create(ctx context.Context, infraction *entities.Infraction) error

	// FindByID busca infração por ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Infraction, error)

	// FindByEntryID busca infrações por entry ID
	FindByEntryID(ctx context.Context, entryID uuid.UUID) ([]*entities.Infraction, error)

	// Update atualiza infração
	Update(ctx context.Context, infraction *entities.Infraction) error

	// List lista infrações com paginação
	List(ctx context.Context, ispb string, limit, offset int) ([]*entities.Infraction, error)

	// CountByISPB conta infrações de um participante
	CountByISPB(ctx context.Context, ispb string) (int64, error)
}
