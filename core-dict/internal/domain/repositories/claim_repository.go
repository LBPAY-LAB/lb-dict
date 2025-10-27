package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
)

// ClaimRepository define as operações de persistência para Claim
type ClaimRepository interface {
	// Create cria uma nova reivindicação
	Create(ctx context.Context, claim *entities.Claim) error

	// Update atualiza uma reivindicação existente
	Update(ctx context.Context, claim *entities.Claim) error

	// Delete realiza soft delete de uma reivindicação
	Delete(ctx context.Context, claimID uuid.UUID) error

	// FindByID busca reivindicação por ID
	FindByID(ctx context.Context, claimID uuid.UUID) (*entities.Claim, error)

	// FindByEntryKey busca reivindicações por chave PIX
	FindByEntryKey(ctx context.Context, entryKey string) ([]*entities.Claim, error)

	// FindByStatus lista reivindicações por status
	FindByStatus(ctx context.Context, status valueobjects.ClaimStatus, limit, offset int) ([]*entities.Claim, error)

	// FindByParticipant lista reivindicações de um participante
	FindByParticipant(ctx context.Context, ispb string, limit, offset int) ([]*entities.Claim, error)

	// FindExpired busca reivindicações expiradas (expires_at < now)
	FindExpired(ctx context.Context, limit int) ([]*entities.Claim, error)

	// FindByWorkflowID busca reivindicação por ID do workflow Temporal
	FindByWorkflowID(ctx context.Context, workflowID string) (*entities.Claim, error)

	// FindPendingResolution busca reivindicações aguardando resolução
	FindPendingResolution(ctx context.Context, limit int) ([]*entities.Claim, error)

	// ExistsActiveClaim verifica se existe claim ativo para uma chave
	ExistsActiveClaim(ctx context.Context, entryKey string) (bool, error)

	// List lista reivindicações com paginação e filtros
	List(ctx context.Context, filters ClaimFilters) ([]*entities.Claim, error)

	// Count conta total de reivindicações
	Count(ctx context.Context, filters ClaimFilters) (int64, error)
}

// ClaimFilters define filtros para listagem de reivindicações
type ClaimFilters struct {
	EntryKey         *string
	ClaimType        *valueobjects.ClaimType
	Status           *valueobjects.ClaimStatus
	ClaimerISPB      *string
	DonorISPB        *string
	ExpiresAfter     *time.Time
	ExpiresBefore    *time.Time
	Limit            int
	Offset           int
}
