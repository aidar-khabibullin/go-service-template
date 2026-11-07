package postgres

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres" // registers the "postgres" dialect ($1 placeholders)
	"github.com/aidar-khabibullin/go-service-template/pkg/transaction"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	tableApple     = "apple"
	tablePineApple = "pine_apple"
	tableOutbox    = "outbox"

	colName   = "name"
	colStatus = "status"
)

// dialect builds SQL with $N placeholders; values are always passed as arguments.
var dialect = goqu.Dialect("postgres")

type Postgres struct {
	pool *pgxpool.Pool
}

func New(p *pgxpool.Pool) *Postgres {
	return &Postgres{
		pool: p,
	}
}

// executor returns the transaction stored in ctx or the pool.
func (p *Postgres) executor(ctx context.Context) transaction.Executor {
	return transaction.GetExecutor(ctx, p.pool)
}
