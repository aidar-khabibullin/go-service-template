package usecase

import (
	"context"
	"fmt"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
	"github.com/google/uuid"
)

// SendOutbox delivers one batch of unsent outbox events to Kafka and returns how many were sent.
// Delivery is at-least-once: if the commit fails after a successful send, the batch is sent again,
// so consumers must be idempotent.
func (u *UseCase) SendOutbox(ctx context.Context) (int, error) {
	ctx, span := tracer.Start(ctx, "usecase SendOutbox")
	defer span.End()

	var sent int

	err := u.tx.Do(ctx, func(ctx context.Context) error {
		events, err := u.postgres.GetOutboxEvents(ctx, u.outboxBatch)
		if err != nil {
			return fmt.Errorf("u.postgres.GetOutboxEvents: %w", err)
		}

		if len(events) == 0 {
			return nil
		}

		err = u.kafka.Send(ctx, events...)
		if err != nil {
			return fmt.Errorf("u.kafka.Send: %w", err)
		}

		err = u.postgres.MarkOutboxEventsSent(ctx, eventIDs(events))
		if err != nil {
			return fmt.Errorf("u.postgres.MarkOutboxEventsSent: %w", err)
		}

		sent = len(events)

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("u.tx.Do: %w", err)
	}

	return sent, nil
}

func eventIDs(events []entity.OutboxEvent) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(events))
	for _, e := range events {
		ids = append(ids, e.ID)
	}

	return ids
}
