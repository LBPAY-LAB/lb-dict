package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
)

// AccountRepository define as operações de persistência para Account
type AccountRepository interface {
	// Create cria uma nova conta
	Create(ctx context.Context, account *entities.Account) error

	// Update atualiza uma conta existente
	Update(ctx context.Context, account *entities.Account) error

	// Delete realiza soft delete de uma conta
	Delete(ctx context.Context, accountID uuid.UUID) error

	// FindByID busca conta por ID
	FindByID(ctx context.Context, accountID uuid.UUID) (*entities.Account, error)

	// FindByAccountNumber busca conta por número, agência e ISPB
	FindByAccountNumber(ctx context.Context, ispb, branch, accountNumber string) (*entities.Account, error)

	// FindByOwnerTaxID busca contas por CPF/CNPJ do titular
	FindByOwnerTaxID(ctx context.Context, taxID string) ([]*entities.Account, error)

	// FindByISPB lista contas de um participante
	FindByISPB(ctx context.Context, ispb string, limit, offset int) ([]*entities.Account, error)

	// ExistsByAccountNumber verifica se conta existe
	ExistsByAccountNumber(ctx context.Context, ispb, branch, accountNumber string) (bool, error)

	// List lista contas com paginação e filtros
	List(ctx context.Context, filters AccountFilters) ([]*entities.Account, error)

	// Count conta total de contas
	Count(ctx context.Context, filters AccountFilters) (int64, error)
}

// AccountFilters define filtros para listagem de contas
type AccountFilters struct {
	ISPB         *string
	OwnerTaxID   *string
	AccountType  *entities.AccountType
	Status       *string
	Limit        int
	Offset       int
}
