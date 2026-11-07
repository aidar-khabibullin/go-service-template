package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func (r *Redis) GetApple(ctx context.Context, id uuid.UUID) (entity.Apple, error) {
	ctx, span := tracer.Start(ctx, "redis GetApple")
	defer span.End()

	data, err := r.client.Get(ctx, key(id)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return entity.Apple{}, entity.ErrNotFound
		}

		return entity.Apple{}, fmt.Errorf("r.client.Get: %w", err)
	}

	var m appleModel

	err = json.Unmarshal(data, &m)
	if err != nil {
		return entity.Apple{}, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return m.toEntity(), nil
}
