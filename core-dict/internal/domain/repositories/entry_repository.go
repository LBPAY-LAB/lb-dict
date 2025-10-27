package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
)

// EntryRepository interface para operações CRUD de chaves PIX
type EntryRepository interface {
	// Create cria uma nova entrada de chave PIX
	Create(ctx context.Context, entry *entities.Entry) error

	// Update atualiza uma entrada existente
	Update(ctx context.Context, entry *entities.Entry) error

	// Delete realiza soft delete de uma entrada
	Delete(ctx context.Context, entryID uuid.UUID) error

	// UpdateStatus atualiza apenas o status de uma entrada
	UpdateStatus(ctx context.Context, entryID uuid.UUID, status entities.KeyStatus) error

	// FindByKey busca uma chave PIX por seu valor
	FindByKey(ctx context.Context, keyValue string) (*entities.Entry, error)

	// FindByID busca uma chave PIX por seu ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Entry, error)

	// List lista chaves PIX com paginação
	List(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*entities.Entry, error)

	// CountByAccount conta chaves de uma conta
	CountByAccount(ctx context.Context, accountID uuid.UUID) (int64, error)

	// CountByOwnerAndType conta chaves de um titular por tipo
	// Usado para validar limites (max 5 CPF, 20 CNPJ, etc.)
	CountByOwnerAndType(ctx context.Context, ownerTaxID string, keyType entities.KeyType) (int, error)
}

// Note: Other repository interfaces (AccountRepository, ClaimRepository, etc.)
// are defined in their respective files
