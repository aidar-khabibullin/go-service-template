package usecase

import (
	"context"
	"time"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/google/uuid"
)

// Transactor runs fn in a transaction; repositories called with the passed ctx join it.
type Transactor interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type Postgres interface {
	CreateApple(ctx context.Context, a entity.Apple) error
	GetApple(ctx context.Context, id uuid.UUID) (entity.Apple, error)
	UpdateApple(ctx context.Context, a entity.Apple) error
	DeleteApple(ctx context.Context, id uuid.UUID) error

	CreatePineApple(ctx context.Context, p entity.PineApple) error

	CreateOutboxEvent(ctx context.Context, e entity.OutboxEvent) error
	GetOutboxEvents(ctx context.Context, limit uint) ([]entity.OutboxEvent, error)
	MarkOutboxEventsSent(ctx context.Context, ids []uuid.UUID) error
}

type Kafka interface {
	Send(ctx context.Context, events ...entity.OutboxEvent) error
}

type Redis interface {
	GetApple(ctx context.Context, id uuid.UUID) (entity.Apple, error)
	PutApple(ctx context.Context, a entity.Apple) error
	DeleteApple(ctx context.Context, id uuid.UUID) error
}

type UseCase struct {
	tx       Transactor
	postgres Postgres
	kafka    Kafka
	redis    Redis

	outboxBatch uint
}

const defaultOutboxBatch = 100

func New(tx Transactor, p Postgres, k Kafka, r Redis) *UseCase {
	return &UseCase{
		tx:          tx,
		postgres:    p,
		kafka:       k,
		redis:       r,
		outboxBatch: defaultOutboxBatch,
	}
}

// cacheTimeout bounds best-effort cache operations done after the main work.
const cacheTimeout = 2 * time.Second
