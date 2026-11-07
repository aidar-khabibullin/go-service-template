package usecase

import (
	"context"
	"fmt"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/dto"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
	"github.com/rs/zerolog/log"
)

func (u *UseCase) DeleteApple(ctx context.Context, input dto.DeleteAppleInput) error {
	ctx, span := tracer.Start(ctx, "usecase DeleteApple")
	defer span.End()

	err := u.postgres.DeleteApple(ctx, input.ID)
	if err != nil {
		return fmt.Errorf("u.postgres.DeleteApple: %w", err)
	}

	cacheCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cacheTimeout)
	defer cancel()

	err = u.redis.DeleteApple(cacheCtx, input.ID)
	if err != nil {
		log.Warn().Err(err).Str("id", input.ID.String()).Msg("u.redis.DeleteApple")
	}

	return nil
}
