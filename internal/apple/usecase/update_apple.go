package usecase

import (
	"context"
	"fmt"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/dto"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
	"github.com/rs/zerolog/log"
)

func (u *UseCase) UpdateApple(ctx context.Context, input dto.UpdateAppleInput) error {
	ctx, span := tracer.Start(ctx, "usecase UpdateApple")
	defer span.End()

	status, err := entity.ParseStatus(input.Status)
	if err != nil {
		return fmt.Errorf("entity.ParseStatus: %w", err)
	}

	err = u.tx.Do(ctx, func(ctx context.Context) error {
		apple, err := u.postgres.GetApple(ctx, input.ID)
		if err != nil {
			return fmt.Errorf("u.postgres.GetApple: %w", err)
		}

		err = apple.Update(input.Name, status)
		if err != nil {
			return fmt.Errorf("apple.Update: %w", err)
		}

		err = u.postgres.UpdateApple(ctx, apple)
		if err != nil {
			return fmt.Errorf("u.postgres.UpdateApple: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("u.tx.Do: %w", err)
	}

	// Invalidate the cache; the next GetApple refills it from Postgres.
	cacheCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cacheTimeout)
	defer cancel()

	err = u.redis.DeleteApple(cacheCtx, input.ID)
	if err != nil {
		log.Warn().Err(err).Str("id", input.ID.String()).Msg("u.redis.DeleteApple")
	}

	return nil
}
