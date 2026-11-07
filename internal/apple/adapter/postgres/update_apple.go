package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
	"github.com/doug-martin/goqu/v9"
)

func (p *Postgres) UpdateApple(ctx context.Context, a entity.Apple) error {
	ctx, span := tracer.Start(ctx, "postgres UpdateApple")
	defer span.End()

	sql, args, err := dialect.Update(tableApple).
		Set(goqu.Record{
			colName:      a.Name,
			colStatus:    a.Status.String(),
			"updated_at": time.Now(),
		}).
		Where(goqu.C("id").Eq(a.ID), goqu.C("deleted_at").IsNull()).
		Prepared(true).
		ToSQL()
	if err != nil {
		return fmt.Errorf("dataset.ToSQL: %w", err)
	}

	tag, err := p.executor(ctx).Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executor.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return entity.ErrNotFound
	}

	return nil
}
