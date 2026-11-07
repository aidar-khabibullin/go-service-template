package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/dto"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (u *UseCase) CreateApple(ctx context.Context, input dto.CreateAppleInput) (dto.CreateAppleOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase CreateApple")
	defer span.End()

	var output dto.CreateAppleOutput

	a, err := entity.NewApple(input.Name)
	if err != nil {
		return output, fmt.Errorf("entity.NewApple: %w", err)
	}

	payload, err := json.Marshal(entity.CreateEvent{ID: a.ID, Name: a.Name})
	if err != nil {
		return output, fmt.Errorf("json.Marshal: %w", err)
	}

	// The apple and its event are stored atomically (transactional outbox);
	// the outbox worker delivers the event to Kafka.
	err = u.tx.Do(ctx, func(ctx context.Context) error {
		err := u.postgres.CreateApple(ctx, a)
		if err != nil {
			return fmt.Errorf("u.postgres.CreateApple: %w", err)
		}

		err = u.postgres.CreateOutboxEvent(ctx, entity.OutboxEvent{
			ID:      uuid.New(),
			Key:     a.ID.String(),
			Payload: payload,
		})
		if err != nil {
			return fmt.Errorf("u.postgres.CreateOutboxEvent: %w", err)
		}

		return nil
	})
	if err != nil {
		return output, fmt.Errorf("u.tx.Do: %w", err)
	}

	// Cache is best-effort: Postgres is the source of truth, GetApple falls back to it.
	cacheCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cacheTimeout)
	defer cancel()

	err = u.redis.PutApple(cacheCtx, a)
	if err != nil {
		log.Warn().Err(err).Str("id", a.ID.String()).Msg("u.redis.PutApple")
	}

	output.ID = a.ID

	return output, nil
}
