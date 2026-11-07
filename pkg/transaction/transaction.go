// Package transaction keeps a pgx transaction in context so that use cases can
// run several repository calls atomically without knowing about pgx.
package transaction

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type ctxKey struct{}

// Executor is implemented by both *pgxpool.Pool and pgx.Tx.
type Executor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Manager struct {
	pool *pgxpool.Pool
}

func NewManager(pool *pgxpool.Pool) *Manager {
	return &Manager{pool: pool}
}

// Do runs fn inside a transaction. The ctx passed to fn carries the transaction;
// repositories must obtain their executor via GetExecutor.
// If fn returns an error or panics, the transaction is rolled back.
// A nested Do joins the already running transaction.
func (m *Manager) Do(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	if _, ok := ctx.Value(ctxKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("pool.Begin: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			rollback(ctx, tx)

			panic(r)
		}
	}()

	err = fn(context.WithValue(ctx, ctxKey{}, tx))
	if err != nil {
		rollback(ctx, tx)

		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		rollback(ctx, tx)

		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

// GetExecutor returns the transaction from ctx, or the pool if there is none.
func GetExecutor(ctx context.Context, pool *pgxpool.Pool) Executor {
	if tx, ok := ctx.Value(ctxKey{}).(pgx.Tx); ok {
		return tx
	}

	return pool
}

func rollback(ctx context.Context, tx pgx.Tx) {
	// ctx may already be cancelled - rollback must still reach the database.
	err := tx.Rollback(context.WithoutCancel(ctx))
	if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		log.Error().Err(err).Msg("transaction.rollback")
	}
}
