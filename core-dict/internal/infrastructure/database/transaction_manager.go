package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TransactionManager handles database transactions
type TransactionManager struct {
	pool *pgxpool.Pool
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{
		pool: pool,
	}
}

// TransactionFunc is a function that executes within a transaction
type TransactionFunc func(tx pgx.Tx) error

// WithTransaction executes a function within a transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, the transaction is committed
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn TransactionFunc) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			// Log rollback error (in production, use proper logger)
			fmt.Printf("failed to rollback transaction: %v\n", err)
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithTransactionContext executes a function within a transaction
// and returns a context with the transaction
func (tm *TransactionManager) WithTransactionContext(ctx context.Context, fn func(context.Context) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			fmt.Printf("failed to rollback transaction: %v\n", err)
		}
	}()

	// Create context with transaction
	txCtx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(txCtx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// txKey is used as a key for storing transaction in context
type txKey struct{}

// GetTx retrieves a transaction from context
func GetTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

// Savepoint creates a savepoint within a transaction
func (tm *TransactionManager) Savepoint(ctx context.Context, tx pgx.Tx, name string) error {
	_, err := tx.Exec(ctx, fmt.Sprintf("SAVEPOINT %s", name))
	return err
}

// RollbackToSavepoint rolls back to a savepoint
func (tm *TransactionManager) RollbackToSavepoint(ctx context.Context, tx pgx.Tx, name string) error {
	_, err := tx.Exec(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", name))
	return err
}

// ReleaseSavepoint releases a savepoint
func (tm *TransactionManager) ReleaseSavepoint(ctx context.Context, tx pgx.Tx, name string) error {
	_, err := tx.Exec(ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", name))
	return err
}
