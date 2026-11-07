package postgres

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
)

func (p *Postgres) CreateApple(ctx context.Context, a entity.Apple) (err error) {
	ctx, span := tracer.Start(ctx, "postgres CreateApple")
	defer span.End()

	sql, args, err := dialect.Insert(tableApple).
		Rows(goqu.Record{
			"id":      a.ID,
			colName:   a.Name,
			colStatus: a.Status.String(),
		}).
		Prepared(true).
		ToSQL()
	if err != nil {
		return fmt.Errorf("dataset.ToSQL: %w", err)
	}

	_, err = p.executor(ctx).Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executor.Exec: %w", err)
	}

	return nil
}
