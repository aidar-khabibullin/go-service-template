package app

import (
	"context"
	"sync"

	"github.com/aidar-khabibullin/go-service-template/config"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/adapter/kafka_producer"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/adapter/postgres"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/adapter/redis"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/controller/http_router"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/controller/kafka_consumer"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/controller/worker"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/usecase"
	"github.com/aidar-khabibullin/go-service-template/pkg/transaction"
)

// AppleDomain wires the domain and starts its background goroutines.
// They stop when ctx is cancelled; wg lets the caller wait for them.
func AppleDomain(ctx context.Context, c config.Config, d Dependencies, wg *sync.WaitGroup) {
	appleUseCase := usecase.New(
		transaction.NewManager(d.Postgres.Pool),
		postgres.New(d.Postgres.Pool),
		kafka_producer.New(d.KafkaWriter.Writer),
		redis.New(d.Redis.Client),
	)

	http_router.AppleRouter(d.RouterHTTP, appleUseCase)

	wg.Go(func() {
		kafka_consumer.AppleConsumer(ctx, d.KafkaReader, appleUseCase)
	})

	wg.Go(func() {
		worker.OutboxWorker(ctx, appleUseCase, c.App.OutboxInterval)
	})
}
