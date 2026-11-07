package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (p *Postgres) CreateOutboxEvent(ctx context.Context, e entity.OutboxEvent) error {
	ctx, span := tracer.Start(ctx, "postgres CreateOutboxEvent")
	defer span.End()

	sql, args, err := dialect.Insert(tableOutbox).
		Rows(goqu.Record{
			"id":      e.ID,
			"key":     e.Key,
			"payload": e.Payload,
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

// GetOutboxEvents returns up to limit unsent events and locks them until the end of
// the surrounding transaction (FOR UPDATE SKIP LOCKED), so several workers do not
// send the same event. Must be called inside a transaction.
func (p *Postgres) GetOutboxEvents(ctx context.Context, limit uint) ([]entity.OutboxEvent, error) {
	ctx, span := tracer.Start(ctx, "postgres GetOutboxEvents")
	defer span.End()

	sql, args, err := dialect.From(tableOutbox).
		Select("id", "key", "payload").
		Where(goqu.C("sent_at").IsNull()).
		Order(goqu.C("created_at").Asc()).
		Limit(limit).
		ForUpdate(goqu.SkipLocked).
		Prepared(true).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("dataset.ToSQL: %w", err)
	}

	rows, err := p.executor(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("executor.Query: %w", err)
	}

	events, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (entity.OutboxEvent, error) {
		var e entity.OutboxEvent

		err := row.Scan(&e.ID, &e.Key, &e.Payload)
		if err != nil {
			return e, fmt.Errorf("row.Scan: %w", err)
		}

		return e, nil
	})
	if err != nil {
		return nil, fmt.Errorf("pgx.CollectRows: %w", err)
	}

	return events, nil
}

func (p *Postgres) MarkOutboxEventsSent(ctx context.Context, ids []uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "postgres MarkOutboxEventsSent")
	defer span.End()

	if len(ids) == 0 {
		return nil
	}

	sql, args, err := dialect.Update(tableOutbox).
		Set(goqu.Record{"sent_at": time.Now()}).
		Where(goqu.C("id").In(ids)).
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
