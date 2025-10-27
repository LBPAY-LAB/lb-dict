package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/repositories"
)

// PostgresAccountRepository implements AccountRepository using PostgreSQL
type PostgresAccountRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresAccountRepository creates a new account repository
func NewPostgresAccountRepository(pool *pgxpool.Pool) repositories.AccountRepository {
	return &PostgresAccountRepository{
		pool: pool,
	}
}

// FindByID finds an account by ID
func (r *PostgresAccountRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Account, error) {
	query := `
		SELECT
			id, participant_ispb, branch_code, account_number,
			account_type, account_status, holder_name,
			holder_document, holder_document_type,
			created_at, updated_at
		FROM core_dict.accounts
		WHERE id = $1 AND deleted_at IS NULL
		LIMIT 1
	`

	var account entities.Account
	var ownerType string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&account.ID,
		&account.ISPB,
		&account.Branch,
		&account.AccountNumber,
		&account.AccountType,
		&account.Status,
		&account.Owner.Name,
		&account.Owner.TaxID,
		&ownerType,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("account not found: %s", id)
		}
		return nil, fmt.Errorf("failed to find account: %w", err)
	}

	account.Owner.Type = entities.OwnerType(ownerType)
	return &account, nil
}

// FindByAccountNumber finds an account by ISPB + branch + account number
func (r *PostgresAccountRepository) FindByAccountNumber(ctx context.Context, ispb, branch, accountNumber string) (*entities.Account, error) {
	query := `
		SELECT
			id, participant_ispb, branch_code, account_number,
			account_type, account_status, holder_name,
			holder_document, holder_document_type,
			created_at, updated_at
		FROM core_dict.accounts
		WHERE participant_ispb = $1
			AND branch_code = $2
			AND account_number = $3
			AND deleted_at IS NULL
		LIMIT 1
	`

	var account entities.Account
	var ownerType string
	err := r.pool.QueryRow(ctx, query, ispb, branch, accountNumber).Scan(
		&account.ID,
		&account.ISPB,
		&account.Branch,
		&account.AccountNumber,
		&account.AccountType,
		&account.Status,
		&account.Owner.Name,
		&account.Owner.TaxID,
		&ownerType,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("account not found: %s/%s/%s", ispb, branch, accountNumber)
		}
		return nil, fmt.Errorf("failed to find account: %w", err)
	}

	account.Owner.Type = entities.OwnerType(ownerType)
	return &account, nil
}

// VerifyAccount verifies if an account is valid and active
func (r *PostgresAccountRepository) VerifyAccount(ctx context.Context, ispb, branch, accountNumber string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM core_dict.accounts
			WHERE participant_ispb = $1
				AND branch_code = $2
				AND account_number = $3
				AND account_status = 'ACTIVE'
				AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, ispb, branch, accountNumber).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to verify account: %w", err)
	}

	return exists, nil
}

// Create creates a new account
func (r *PostgresAccountRepository) Create(ctx context.Context, account *entities.Account) error {
	query := `
		INSERT INTO core_dict.accounts (
			id, participant_ispb, branch_code, account_number,
			account_type, account_status, holder_name,
			holder_document, holder_document_type,
			opened_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	documentType := getDocumentType(account.Owner.TaxID)

	_, err := r.pool.Exec(ctx, query,
		account.ID,
		account.ISPB,
		account.Branch,
		account.AccountNumber,
		account.AccountType,
		account.Status,
		account.Owner.Name,
		account.Owner.TaxID,
		documentType,
		account.OpenedAt,
		account.CreatedAt,
		account.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

// Update updates an existing account
func (r *PostgresAccountRepository) Update(ctx context.Context, account *entities.Account) error {
	query := `
		UPDATE core_dict.accounts
		SET holder_name = $2,
			account_status = $3,
			updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query,
		account.ID,
		account.Owner.Name,
		account.Status,
		account.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("account not found: %s", account.ID)
	}

	return nil
}

// Delete performs soft delete on an account
func (r *PostgresAccountRepository) Delete(ctx context.Context, accountID uuid.UUID) error {
	query := `
		UPDATE core_dict.accounts
		SET account_status = 'CLOSED',
			closed_at = $2,
			deleted_at = $2,
			updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.pool.Exec(ctx, query, accountID, now)

	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("account not found or already deleted: %s", accountID)
	}

	return nil
}

// FindByOwnerTaxID finds accounts by owner's CPF/CNPJ
func (r *PostgresAccountRepository) FindByOwnerTaxID(ctx context.Context, taxID string) ([]*entities.Account, error) {
	query := `
		SELECT
			id, participant_ispb, branch_code, account_number,
			account_type, account_status, holder_name,
			holder_document, holder_document_type,
			created_at, updated_at
		FROM core_dict.accounts
		WHERE holder_document = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, taxID)
	if err != nil {
		return nil, fmt.Errorf("failed to find accounts by tax ID: %w", err)
	}
	defer rows.Close()

	var accounts []*entities.Account
	for rows.Next() {
		var account entities.Account
		var ownerType string

		err := rows.Scan(
			&account.ID,
			&account.ISPB,
			&account.Branch,
			&account.AccountNumber,
			&account.AccountType,
			&account.Status,
			&account.Owner.Name,
			&account.Owner.TaxID,
			&ownerType,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}

		account.Owner.Type = entities.OwnerType(ownerType)
		accounts = append(accounts, &account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return accounts, nil
}

// FindByISPB lists accounts for a participant with pagination
func (r *PostgresAccountRepository) FindByISPB(ctx context.Context, ispb string, limit, offset int) ([]*entities.Account, error) {
	query := `
		SELECT
			id, participant_ispb, branch_code, account_number,
			account_type, account_status, holder_name,
			holder_document, holder_document_type,
			created_at, updated_at
		FROM core_dict.accounts
		WHERE participant_ispb = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, ispb, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find accounts by ISPB: %w", err)
	}
	defer rows.Close()

	var accounts []*entities.Account
	for rows.Next() {
		var account entities.Account
		var ownerType string

		err := rows.Scan(
			&account.ID,
			&account.ISPB,
			&account.Branch,
			&account.AccountNumber,
			&account.AccountType,
			&account.Status,
			&account.Owner.Name,
			&account.Owner.TaxID,
			&ownerType,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}

		account.Owner.Type = entities.OwnerType(ownerType)
		accounts = append(accounts, &account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return accounts, nil
}

// ExistsByAccountNumber checks if an account exists
func (r *PostgresAccountRepository) ExistsByAccountNumber(ctx context.Context, ispb, branch, accountNumber string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM core_dict.accounts
			WHERE participant_ispb = $1
				AND branch_code = $2
				AND account_number = $3
				AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, ispb, branch, accountNumber).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check account existence: %w", err)
	}

	return exists, nil
}

// List lists accounts with filters and pagination
func (r *PostgresAccountRepository) List(ctx context.Context, filters repositories.AccountFilters) ([]*entities.Account, error) {
	query := `
		SELECT
			id, participant_ispb, branch_code, account_number,
			account_type, account_status, holder_name,
			holder_document, holder_document_type,
			created_at, updated_at
		FROM core_dict.accounts
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	argPos := 1

	if filters.ISPB != nil {
		query += fmt.Sprintf(" AND participant_ispb = $%d", argPos)
		args = append(args, *filters.ISPB)
		argPos++
	}

	if filters.OwnerTaxID != nil {
		query += fmt.Sprintf(" AND holder_document = $%d", argPos)
		args = append(args, *filters.OwnerTaxID)
		argPos++
	}

	if filters.AccountType != nil {
		query += fmt.Sprintf(" AND account_type = $%d", argPos)
		args = append(args, *filters.AccountType)
		argPos++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND account_status = $%d", argPos)
		args = append(args, *filters.Status)
		argPos++
	}

	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filters.Limit)
		argPos++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filters.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*entities.Account
	for rows.Next() {
		var account entities.Account
		var ownerType string

		err := rows.Scan(
			&account.ID,
			&account.ISPB,
			&account.Branch,
			&account.AccountNumber,
			&account.AccountType,
			&account.Status,
			&account.Owner.Name,
			&account.Owner.TaxID,
			&ownerType,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}

		account.Owner.Type = entities.OwnerType(ownerType)
		accounts = append(accounts, &account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return accounts, nil
}

// Count counts total accounts with filters
func (r *PostgresAccountRepository) Count(ctx context.Context, filters repositories.AccountFilters) (int64, error) {
	query := `SELECT COUNT(*) FROM core_dict.accounts WHERE deleted_at IS NULL`

	args := []interface{}{}
	argPos := 1

	if filters.ISPB != nil {
		query += fmt.Sprintf(" AND participant_ispb = $%d", argPos)
		args = append(args, *filters.ISPB)
		argPos++
	}

	if filters.OwnerTaxID != nil {
		query += fmt.Sprintf(" AND holder_document = $%d", argPos)
		args = append(args, *filters.OwnerTaxID)
		argPos++
	}

	if filters.AccountType != nil {
		query += fmt.Sprintf(" AND account_type = $%d", argPos)
		args = append(args, *filters.AccountType)
		argPos++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND account_status = $%d", argPos)
		args = append(args, *filters.Status)
	}

	var count int64
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count accounts: %w", err)
	}

	return count, nil
}

// getDocumentType determines if the tax ID is CPF or CNPJ based on length
func getDocumentType(taxID string) string {
	if len(taxID) == 11 {
		return "CPF"
	}
	return "CNPJ"
}
