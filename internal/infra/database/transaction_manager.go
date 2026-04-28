package database

import (
	"context"
	"database/sql"
	"fmt"
)

type txContextKey struct{}

type sqlExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type transactionManager struct {
	db *sql.DB
}

func NewTransactionManager(db *sql.DB) *transactionManager {
	return &transactionManager{db: db}
}

func (m *transactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("transactionManager.WithinTransaction: begin: %w", err)
	}

	txCtx := context.WithValue(ctx, txContextKey{}, tx)
	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transactionManager.WithinTransaction: rollback: %v (original: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("transactionManager.WithinTransaction: commit: %w", err)
	}
	return nil
}

func executorFromContext(ctx context.Context, db *sql.DB) sqlExecutor {
	tx, ok := ctx.Value(txContextKey{}).(*sql.Tx)
	if ok && tx != nil {
		return tx
	}
	return db
}
