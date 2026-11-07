package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/dto"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/usecase/usecasetest"
	"github.com/google/uuid"
)

func TestCreateApple(t *testing.T) {
	t.Parallel()

	env := usecasetest.NewEnv()

	out, err := env.UC.CreateApple(context.Background(), dto.CreateAppleInput{Name: "fuji"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := env.Postgres.Apples[out.ID]; !ok {
		t.Error("apple is not stored in postgres")
	}

	if _, ok := env.Redis.Cache[out.ID]; !ok {
		t.Error("apple is not cached")
	}

	if len(env.Postgres.Outbox) != 1 {
		t.Fatalf("want 1 outbox event, got %d", len(env.Postgres.Outbox))
	}

	var ev entity.CreateEvent
	if err = json.Unmarshal(env.Postgres.Outbox[0].Payload, &ev); err != nil || ev.ID != out.ID || ev.Name != "fuji" {
		t.Errorf("bad outbox payload: %+v, %v", ev, err)
	}

	if len(env.Kafka.Messages) != 0 {
		t.Error("kafka must be written by the outbox worker, not synchronously")
	}
}

func TestCreateApple_InvalidName(t *testing.T) {
	t.Parallel()

	env := usecasetest.NewEnv()

	_, err := env.UC.CreateApple(context.Background(), dto.CreateAppleInput{})
	if !errors.Is(err, entity.ErrNameInvalid) {
		t.Fatalf("want ErrNameInvalid, got %v", err)
	}
}

func TestCreateApple_OutboxFailureFailsRequest(t *testing.T) {
	t.Parallel()

	env := usecasetest.NewEnv()
	env.Postgres.Err["CreateOutboxEvent"] = usecasetest.ErrInjected

	_, err := env.UC.CreateApple(context.Background(), dto.CreateAppleInput{Name: "fuji"})
	if !errors.Is(err, usecasetest.ErrInjected) {
		t.Fatalf("want injected error, got %v", err)
	}

	if env.Tx.Rolled != 1 {
		t.Error("transaction must be rolled back")
	}

	if len(env.Redis.Cache) != 0 {
		t.Error("cache must not be filled when the transaction failed")
	}
}

func TestCreateApple_RedisFailureIsNotFatal(t *testing.T) {
	t.Parallel()

	env := usecasetest.NewEnv()
	env.Redis.Err = usecasetest.ErrInjected

	out, err := env.UC.CreateApple(context.Background(), dto.CreateAppleInput{Name: "fuji"})
	if err != nil {
		t.Fatalf("redis failure must not fail the request: %v", err)
	}

	if _, ok := env.Postgres.Apples[out.ID]; !ok {
		t.Error("apple must be stored")
	}
}

func TestGetApple(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	apple := entity.Apple{ID: uuid.New(), Name: "gala", Status: entity.StatusSuccess}

	t.Run("cache hit", func(t *testing.T) {
		t.Parallel()

		env := usecasetest.NewEnv()
		env.Redis.Cache[apple.ID] = apple

		out, err := env.UC.GetApple(ctx, dto.GetAppleInput{ID: apple.ID})
		if err != nil || out.Name != "gala" || out.Status != "success" {
			t.Fatalf("got %+v, %v", out, err)
		}
	})

	t.Run("cache miss refills cache from postgres", func(t *testing.T) {
		t.Parallel()

		env := usecasetest.NewEnv()
		env.Postgres.Apples[apple.ID] = apple

		out, err := env.UC.GetApple(ctx, dto.GetAppleInput{ID: apple.ID})
		if err != nil || out.Name != "gala" {
			t.Fatalf("got %+v, %v", out, err)
		}

		if _, ok := env.Redis.Cache[apple.ID]; !ok {
			t.Error("cache must be refilled")
		}
	})

	t.Run("redis down falls back to postgres", func(t *testing.T) {
		t.Parallel()

		env := usecasetest.NewEnv()
		env.Redis.Err = usecasetest.ErrInjected
		env.Postgres.Apples[apple.ID] = apple

		out, err := env.UC.GetApple(ctx, dto.GetAppleInput{ID: apple.ID})
		if err != nil || out.Name != "gala" {
			t.Fatalf("got %+v, %v", out, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		env := usecasetest.NewEnv()

		_, err := env.UC.GetApple(ctx, dto.GetAppleInput{ID: uuid.New()})
		if !errors.Is(err, entity.ErrNotFound) {
			t.Fatalf("want ErrNotFound, got %v", err)
		}
	})
}

func TestUpdateApple(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	apple := entity.Apple{ID: uuid.New(), Name: "old", Status: entity.StatusNew}

	t.Run("updates and invalidates cache", func(t *testing.T) {
		t.Parallel()

		env := usecasetest.NewEnv()
		env.Postgres.Apples[apple.ID] = apple
		env.Redis.Cache[apple.ID] = apple

		err := env.UC.UpdateApple(ctx, dto.UpdateAppleInput{ID: apple.ID, Name: "new", Status: "success"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got := env.Postgres.Apples[apple.ID]
		if got.Name != "new" || got.Status != entity.StatusSuccess {
			t.Errorf("not updated: %+v", got)
		}

		if _, ok := env.Redis.Cache[apple.ID]; ok {
			t.Error("cache must be invalidated")
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		env := usecasetest.NewEnv()

		err := env.UC.UpdateApple(ctx, dto.UpdateAppleInput{ID: uuid.New(), Name: "n", Status: "success"})
		if !errors.Is(err, entity.ErrNotFound) {
			t.Fatalf("want ErrNotFound, got %v", err)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		t.Parallel()

		env := usecasetest.NewEnv()
		env.Postgres.Apples[apple.ID] = apple

		err := env.UC.UpdateApple(ctx, dto.UpdateAppleInput{ID: apple.ID, Name: "n", Status: "bogus"})
		if !errors.Is(err, entity.ErrStatusInvalid) {
			t.Fatalf("want ErrStatusInvalid, got %v", err)
		}
	})
}

func TestDeleteApple(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	apple := entity.Apple{ID: uuid.New(), Name: "x"}

	env := usecasetest.NewEnv()
	env.Postgres.Apples[apple.ID] = apple
	env.Redis.Cache[apple.ID] = apple

	if err := env.UC.DeleteApple(ctx, dto.DeleteAppleInput{ID: apple.ID}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(env.Postgres.Apples) != 0 || len(env.Redis.Cache) != 0 {
		t.Error("apple must be removed from postgres and cache")
	}

	err := env.UC.DeleteApple(ctx, dto.DeleteAppleInput{ID: apple.ID})
	if !errors.Is(err, entity.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestCreatePineApple(t *testing.T) {
	t.Parallel()

	env := usecasetest.NewEnv()

	out, err := env.UC.CreatePineApple(context.Background(), dto.CreatePineAppleInput{Name: "pine"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(env.Postgres.Apples) != 1 || len(env.Postgres.PineApples) != 1 {
		t.Fatal("apple and pineapple must both be stored")
	}

	pa := env.Postgres.PineApples[0]
	if pa.ID != out.ID {
		t.Error("output must contain the pineapple id")
	}

	if _, ok := env.Postgres.Apples[pa.AppleID]; !ok {
		t.Error("pineapple must reference the stored apple")
	}
}

func TestCreatePineApple_Rollback(t *testing.T) {
	t.Parallel()

	env := usecasetest.NewEnv()
	env.Postgres.Err["CreatePineApple"] = usecasetest.ErrInjected

	_, err := env.UC.CreatePineApple(context.Background(), dto.CreatePineAppleInput{Name: "pine"})
	if !errors.Is(err, usecasetest.ErrInjected) {
		t.Fatalf("want injected error, got %v", err)
	}

	if env.Tx.Rolled != 1 {
		t.Error("transaction must be rolled back")
	}
}

func TestSendOutbox(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	env := usecasetest.NewEnv()

	for range 3 {
		if _, err := env.UC.CreateApple(ctx, dto.CreateAppleInput{Name: "a"}); err != nil {
			t.Fatal(err)
		}
	}

	n, err := env.UC.SendOutbox(ctx)
	if err != nil || n != 3 {
		t.Fatalf("got %d, %v", n, err)
	}

	if len(env.Kafka.Messages) != 3 {
		t.Errorf("want 3 kafka messages, got %d", len(env.Kafka.Messages))
	}

	n, err = env.UC.SendOutbox(ctx)
	if err != nil || n != 0 {
		t.Fatalf("second run must send nothing, got %d, %v", n, err)
	}
}

func TestSendOutbox_KafkaFailureKeepsEventsUnsent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	env := usecasetest.NewEnv()

	if _, err := env.UC.CreateApple(ctx, dto.CreateAppleInput{Name: "a"}); err != nil {
		t.Fatal(err)
	}

	env.Kafka.Err = usecasetest.ErrInjected

	if _, err := env.UC.SendOutbox(ctx); !errors.Is(err, usecasetest.ErrInjected) {
		t.Fatalf("want injected error, got %v", err)
	}

	if len(env.Postgres.Sent) != 0 {
		t.Error("events must stay unsent")
	}

	env.Kafka.Err = nil

	n, err := env.UC.SendOutbox(ctx)
	if err != nil || n != 1 {
		t.Fatalf("retry: got %d, %v", n, err)
	}
}
