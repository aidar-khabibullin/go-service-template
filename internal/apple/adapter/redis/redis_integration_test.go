//go:build integration

package redis_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/adapter/internal/itest"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/adapter/redis"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/google/uuid"
)

func TestRedis(t *testing.T) {
	ctx := context.Background()
	r := redis.New(itest.Redis(t))

	a, _ := entity.NewApple("fuji")
	a.Status = entity.StatusSuccess

	if _, err := r.GetApple(ctx, a.ID); !errors.Is(err, entity.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}

	if err := r.PutApple(ctx, a); err != nil {
		t.Fatalf("PutApple: %v", err)
	}

	got, err := r.GetApple(ctx, a.ID)
	if err != nil || got.ID != a.ID || got.Name != "fuji" || got.Status != entity.StatusSuccess {
		t.Fatalf("GetApple: %+v, %v", got, err)
	}

	if err = r.DeleteApple(ctx, a.ID); err != nil {
		t.Fatalf("DeleteApple: %v", err)
	}

	if _, err = r.GetApple(ctx, a.ID); !errors.Is(err, entity.ErrNotFound) {
		t.Fatalf("want ErrNotFound after delete, got %v", err)
	}

	if err = r.DeleteApple(ctx, uuid.New()); err != nil {
		t.Fatalf("deleting a missing key must not fail: %v", err)
	}
}
