package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/dto"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
	"github.com/rs/zerolog/log"
)

// GetApple implements cache-aside: Redis first, then Postgres, then refill the cache.
func (u *UseCase) GetApple(ctx context.Context, input dto.GetAppleInput) (dto.GetAppleOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase GetApple")
	defer span.End()

	apple, err := u.redis.GetApple(ctx, input.ID)
	switch {
	case err == nil:
		return toGetOutput(apple), nil

	case errors.Is(err, entity.ErrNotFound):
		// cache miss

	default:
		// Redis is unavailable - do not fail the request, go to the source of truth.
		log.Warn().Err(err).Msg("u.redis.GetApple")
	}

	apple, err = u.postgres.GetApple(ctx, input.ID)
	if err != nil {
		return dto.GetAppleOutput{}, fmt.Errorf("u.postgres.GetApple: %w", err)
	}

	err = u.redis.PutApple(ctx, apple)
	if err != nil {
		log.Warn().Err(err).Msg("u.redis.PutApple")
	}

	return toGetOutput(apple), nil
}

func toGetOutput(a entity.Apple) dto.GetAppleOutput {
	return dto.GetAppleOutput{
		ID:     a.ID,
		Name:   a.Name,
		Status: a.Status.String(),
	}
}
