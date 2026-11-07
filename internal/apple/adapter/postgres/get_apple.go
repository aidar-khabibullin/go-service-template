package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (p *Postgres) GetApple(ctx context.Context, id uuid.UUID) (entity.Apple, error) {
	ctx, span := tracer.Start(ctx, "postgres GetApple")
	defer span.End()

	var apple entity.Apple

	sql, args, err := dialect.From(tableApple).
		Select("id", colName, colStatus).
		Where(goqu.C("id").Eq(id), goqu.C("deleted_at").IsNull()).
		Prepared(true).
		ToSQL()
	if err != nil {
		return apple, fmt.Errorf("dataset.ToSQL: %w", err)
	}

	var status string

	err = p.executor(ctx).QueryRow(ctx, sql, args...).Scan(&apple.ID, &apple.Name, &status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apple, entity.ErrNotFound
		}

		return apple, fmt.Errorf("row.Scan: %w", err)
	}

	apple.Status = entity.NewStatus(status)

	return apple, nil
}
