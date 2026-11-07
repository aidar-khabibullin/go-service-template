package redis

import (
	"context"
	"fmt"

	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
	"github.com/google/uuid"
)

func (r *Redis) DeleteApple(ctx context.Context, id uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "redis DeleteApple")
	defer span.End()

	err := r.client.Del(ctx, key(id)).Err()
	if err != nil {
		return fmt.Errorf("r.client.Del: %w", err)
	}

	return nil
}
