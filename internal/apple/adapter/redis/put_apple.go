package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
	"github.com/google/uuid"
)

func key(id uuid.UUID) string {
	return "apple:" + id.String()
}

func (r *Redis) PutApple(ctx context.Context, a entity.Apple) error {
	ctx, span := tracer.Start(ctx, "redis PutApple")
	defer span.End()

	data, err := json.Marshal(toModel(a))
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	err = r.client.Set(ctx, key(a.ID), data, ttl).Err()
	if err != nil {
		return fmt.Errorf("r.client.Set: %w", err)
	}

	return nil
}
