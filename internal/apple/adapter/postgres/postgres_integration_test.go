//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/adapter/internal/itest"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/adapter/postgres"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/aidar-khabibullin/go-service-template/pkg/transaction"
	"github.com/google/uuid"
)

const migrations = "../../../../migration/postgres/apple"

var errBoom = errors.New("boom")

func TestAppleCRUD(t *testing.T) {
	ctx := context.Background()
	pg := postgres.New(itest.Postgres(t, migrations))

	a, _ := entity.NewApple("fuji")

	if err := pg.CreateApple(ctx, a); err != nil {
		t.Fatalf("CreateApple: %v", err)
	}

	got, err := pg.GetApple(ctx, a.ID)
	if err != nil || got.Name != "fuji" || got.Status != entity.StatusNew {
		t.Fatalf("GetApple: %+v, %v", got, err)
	}

	// SQL injection attempt must be stored as plain data, not executed.
	evil := entity.Apple{ID: uuid.New(), Name: "x'); DROP TABLE apple;--", Status: entity.StatusNew}
	if err = pg.CreateApple(ctx, evil); err != nil {
		t.Fatalf("CreateApple evil: %v", err)
	}

	if got, err = pg.GetApple(ctx, evil.ID); err != nil || got.Name != evil.Name {
		t.Fatalf("evil apple not stored verbatim: %+v, %v", got, err)
	}

	a.Name, a.Status = "gala", entity.StatusSuccess
	if err = pg.UpdateApple(ctx, a); err != nil {
		t.Fatalf("UpdateApple: %v", err)
	}

	if got, _ = pg.GetApple(ctx, a.ID); got.Name != "gala" || got.Status != entity.StatusSuccess {
		t.Fatalf("not updated: %+v", got)
	}

	if err = pg.DeleteApple(ctx, a.ID); err != nil {
		t.Fatalf("DeleteApple: %v", err)
	}

	if _, err = pg.GetApple(ctx, a.ID); !errors.Is(err, entity.ErrNotFound) {
		t.Fatalf("deleted apple must be not found, got %v", err)
	}

	if err = pg.DeleteApple(ctx, a.ID); !errors.Is(err, entity.ErrNotFound) {
		t.Fatalf("second delete: want ErrNotFound, got %v", err)
	}

	if err = pg.UpdateApple(ctx, entity.Apple{ID: uuid.New(), Name: "n", Status: entity.StatusNew}); !errors.Is(err, entity.ErrNotFound) {
		t.Fatalf("update missing: want ErrNotFound, got %v", err)
	}
}

func TestTransaction(t *testing.T) {
	ctx := context.Background()
	pool := itest.Postgres(t, migrations)
	pg := postgres.New(pool)
	tx := transaction.NewManager(pool)

	t.Run("rollback", func(t *testing.T) {
		a, _ := entity.NewApple("rolled-back")

		err := tx.Do(ctx, func(ctx context.Context) error {
			if err := pg.CreateApple(ctx, a); err != nil {
				return err
			}

			return errBoom
		})
		if !errors.Is(err, errBoom) {
			t.Fatalf("want errBoom, got %v", err)
		}

		if _, err = pg.GetApple(ctx, a.ID); !errors.Is(err, entity.ErrNotFound) {
			t.Fatalf("apple must be rolled back, got %v", err)
		}
	})

	t.Run("commit", func(t *testing.T) {
		a, _ := entity.NewApple("committed")
		p, _ := entity.NewPineApple(a.ID, "pine")

		err := tx.Do(ctx, func(ctx context.Context) error {
			if err := pg.CreateApple(ctx, a); err != nil {
				return err
			}

			return pg.CreatePineApple(ctx, p)
		})
		if err != nil {
			t.Fatalf("Do: %v", err)
		}

		if _, err = pg.GetApple(ctx, a.ID); err != nil {
			t.Fatalf("apple must be committed: %v", err)
		}
	})

	t.Run("panic rolls back", func(t *testing.T) {
		a, _ := entity.NewApple("panicked")

		func() {
			defer func() { _ = recover() }()

			_ = tx.Do(ctx, func(ctx context.Context) error {
				_ = pg.CreateApple(ctx, a)

				panic("oops")
			})
		}()

		if _, err := pg.GetApple(ctx, a.ID); !errors.Is(err, entity.ErrNotFound) {
			t.Fatalf("apple must be rolled back after panic, got %v", err)
		}
	})
}

func TestOutbox(t *testing.T) {
	ctx := context.Background()
	pool := itest.Postgres(t, migrations)
	pg := postgres.New(pool)
	tx := transaction.NewManager(pool)

	for i := range 3 {
		e := entity.OutboxEvent{ID: uuid.New(), Key: "k", Payload: []byte{byte(i)}}
		if err := pg.CreateOutboxEvent(ctx, e); err != nil {
			t.Fatalf("CreateOutboxEvent: %v", err)
		}
	}

	var sent int

	err := tx.Do(ctx, func(ctx context.Context) error {
		events, err := pg.GetOutboxEvents(ctx, 2)
		if err != nil {
			return err
		}

		sent = len(events)

		ids := make([]uuid.UUID, 0, len(events))
		for _, e := range events {
			ids = append(ids, e.ID)
		}

		return pg.MarkOutboxEventsSent(ctx, ids)
	})
	if err != nil || sent != 2 {
		t.Fatalf("first batch: sent=%d err=%v", sent, err)
	}

	err = tx.Do(ctx, func(ctx context.Context) error {
		events, err := pg.GetOutboxEvents(ctx, 10)
		sent = len(events)

		return err
	})
	if err != nil || sent != 1 {
		t.Fatalf("second batch: want 1 left, sent=%d err=%v", sent, err)
	}
}
