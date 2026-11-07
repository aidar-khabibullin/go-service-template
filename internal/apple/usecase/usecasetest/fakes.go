// Package usecasetest contains in-memory fakes of the use case dependencies for tests.
package usecasetest

import (
	"context"
	"errors"
	"sync"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/usecase"
	"github.com/google/uuid"
)

var ErrInjected = errors.New("injected error")

// Tx runs fn directly. Rolled counts calls where fn returned an error.
type Tx struct {
	Rolled int
}

func (t *Tx) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	err := fn(ctx)
	if err != nil {
		t.Rolled++
	}

	return err
}

type Postgres struct {
	mu sync.Mutex

	Apples     map[uuid.UUID]entity.Apple
	PineApples []entity.PineApple
	Outbox     []entity.OutboxEvent
	Sent       map[uuid.UUID]bool

	// Err, when not nil, is returned by the method named in the key.
	Err map[string]error
}

func NewPostgres() *Postgres {
	return &Postgres{
		Apples: map[uuid.UUID]entity.Apple{},
		Sent:   map[uuid.UUID]bool{},
		Err:    map[string]error{},
	}
}

func (p *Postgres) CreateApple(_ context.Context, a entity.Apple) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.Err["CreateApple"]; err != nil {
		return err
	}

	p.Apples[a.ID] = a

	return nil
}

func (p *Postgres) GetApple(_ context.Context, id uuid.UUID) (entity.Apple, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.Err["GetApple"]; err != nil {
		return entity.Apple{}, err
	}

	a, ok := p.Apples[id]
	if !ok {
		return entity.Apple{}, entity.ErrNotFound
	}

	return a, nil
}

func (p *Postgres) UpdateApple(_ context.Context, a entity.Apple) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.Err["UpdateApple"]; err != nil {
		return err
	}

	if _, ok := p.Apples[a.ID]; !ok {
		return entity.ErrNotFound
	}

	p.Apples[a.ID] = a

	return nil
}

func (p *Postgres) DeleteApple(_ context.Context, id uuid.UUID) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.Err["DeleteApple"]; err != nil {
		return err
	}

	if _, ok := p.Apples[id]; !ok {
		return entity.ErrNotFound
	}

	delete(p.Apples, id)

	return nil
}

func (p *Postgres) CreatePineApple(_ context.Context, pa entity.PineApple) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.Err["CreatePineApple"]; err != nil {
		return err
	}

	p.PineApples = append(p.PineApples, pa)

	return nil
}

func (p *Postgres) CreateOutboxEvent(_ context.Context, e entity.OutboxEvent) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.Err["CreateOutboxEvent"]; err != nil {
		return err
	}

	p.Outbox = append(p.Outbox, e)

	return nil
}

func (p *Postgres) GetOutboxEvents(_ context.Context, limit uint) ([]entity.OutboxEvent, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.Err["GetOutboxEvents"]; err != nil {
		return nil, err
	}

	var res []entity.OutboxEvent

	for _, e := range p.Outbox {
		if !p.Sent[e.ID] && uint(len(res)) < limit {
			res = append(res, e)
		}
	}

	return res, nil
}

func (p *Postgres) MarkOutboxEventsSent(_ context.Context, ids []uuid.UUID) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.Err["MarkOutboxEventsSent"]; err != nil {
		return err
	}

	for _, id := range ids {
		p.Sent[id] = true
	}

	return nil
}

type Kafka struct {
	Messages []entity.OutboxEvent
	Err      error
}

func (k *Kafka) Send(_ context.Context, events ...entity.OutboxEvent) error {
	if k.Err != nil {
		return k.Err
	}

	k.Messages = append(k.Messages, events...)

	return nil
}

type Redis struct {
	Cache map[uuid.UUID]entity.Apple

	// Err, when not nil, is returned by every method (cache is "down").
	Err error
}

func NewRedis() *Redis {
	return &Redis{Cache: map[uuid.UUID]entity.Apple{}}
}

func (r *Redis) GetApple(_ context.Context, id uuid.UUID) (entity.Apple, error) {
	if r.Err != nil {
		return entity.Apple{}, r.Err
	}

	a, ok := r.Cache[id]
	if !ok {
		return entity.Apple{}, entity.ErrNotFound
	}

	return a, nil
}

func (r *Redis) PutApple(_ context.Context, a entity.Apple) error {
	if r.Err != nil {
		return r.Err
	}

	r.Cache[a.ID] = a

	return nil
}

func (r *Redis) DeleteApple(_ context.Context, id uuid.UUID) error {
	if r.Err != nil {
		return r.Err
	}

	delete(r.Cache, id)

	return nil
}

// Env bundles a use case with its fakes.
type Env struct {
	UC       *usecase.UseCase
	Tx       *Tx
	Postgres *Postgres
	Kafka    *Kafka
	Redis    *Redis
}

func NewEnv() *Env {
	e := &Env{
		Tx:       &Tx{},
		Postgres: NewPostgres(),
		Kafka:    &Kafka{},
		Redis:    NewRedis(),
	}

	e.UC = usecase.New(e.Tx, e.Postgres, e.Kafka, e.Redis)

	return e
}
