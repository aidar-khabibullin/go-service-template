package postgres

import (
	"context"
	"fmt"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
	"github.com/doug-martin/goqu/v9"
)

func (p *Postgres) CreatePineApple(ctx context.Context, pa entity.PineApple) (err error) {
	ctx, span := tracer.Start(ctx, "postgres CreatePineApple")
	defer span.End()

	sql, args, err := dialect.Insert(tablePineApple).
		Rows(goqu.Record{
			"id":       pa.ID,
			"apple_id": pa.AppleID,
			colName:    pa.Name,
			colStatus:  pa.Status.String(),
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
