package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
)

// PostgresInfractionRepository implementa InfractionRepository usando PostgreSQL
type PostgresInfractionRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresInfractionRepository cria um novo PostgresInfractionRepository
func NewPostgresInfractionRepository(pool *pgxpool.Pool) *PostgresInfractionRepository {
	return &PostgresInfractionRepository{
		pool: pool,
	}
}

// Create cria uma nova infração
func (r *PostgresInfractionRepository) Create(ctx context.Context, infraction *entities.Infraction) error {
	query := `
		INSERT INTO infractions (
			id, entry_key, infraction_type, status,
			reporter_participant_ispb, reported_participant_ispb,
			bacen_infraction_id, description, evidence,
			resolution, resolved_at, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	_, err := r.pool.Exec(ctx, query,
		infraction.ID,
		infraction.EntryKey,
		infraction.Type,
		infraction.Status,
		infraction.ReporterParticipant.ISPB,
		infraction.ReportedParticipant.ISPB,
		infraction.BacenInfractionID,
		infraction.Description,
		infraction.Evidence,
		infraction.Resolution,
		infraction.ResolvedAt,
		infraction.Metadata,
		infraction.CreatedAt,
		infraction.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create infraction: %w", err)
	}

	return nil
}

// FindByID busca infração por ID
func (r *PostgresInfractionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Infraction, error) {
	// TODO: Implement full query when infractions table schema is finalized
	return nil, fmt.Errorf("infraction not found: %s", id)
}

// FindByEntryID busca infrações por entry ID
func (r *PostgresInfractionRepository) FindByEntryID(ctx context.Context, entryID uuid.UUID) ([]*entities.Infraction, error) {
	// TODO: Implement full query when infractions table schema is finalized
	return []*entities.Infraction{}, nil
}

// Update atualiza infração
func (r *PostgresInfractionRepository) Update(ctx context.Context, infraction *entities.Infraction) error {
	// TODO: Implement full update when infractions table schema is finalized
	return fmt.Errorf("infraction not found: %s", infraction.ID)
}

// List lista infrações com paginação
func (r *PostgresInfractionRepository) List(ctx context.Context, ispb string, limit, offset int) ([]*entities.Infraction, error) {
	// TODO: Implement full query when infractions table schema is finalized
	// For now, return empty list
	return []*entities.Infraction{}, nil
}

// CountByISPB conta infrações de um participante
func (r *PostgresInfractionRepository) CountByISPB(ctx context.Context, ispb string) (int64, error) {
	// TODO: Implement full query when infractions table schema is finalized
	// For now, return 0
	return 0, nil
}
