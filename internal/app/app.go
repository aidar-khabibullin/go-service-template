package app

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/aidar-khabibullin/go-service-template/config"
	"github.com/aidar-khabibullin/go-service-template/pkg/http_server"
	"github.com/aidar-khabibullin/go-service-template/pkg/kafka_reader"
	"github.com/aidar-khabibullin/go-service-template/pkg/kafka_writer"
	"github.com/aidar-khabibullin/go-service-template/pkg/postgres"
	"github.com/aidar-khabibullin/go-service-template/pkg/redis"
	"github.com/aidar-khabibullin/go-service-template/pkg/router"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Dependencies struct {
	// Adapters
	Postgres    *postgres.Pool
	KafkaWriter *kafka_writer.Writer
	Redis       *redis.Client

	// Controllers
	RouterHTTP  *chi.Mux
	KafkaReader *kafka_reader.Reader
}

func Run(ctx context.Context, c config.Config) (err error) {
	// ctx is cancelled on SIGINT/SIGTERM and stops all background goroutines.
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	var deps Dependencies

	// Adapters
	deps.Postgres, err = postgres.New(ctx, c.Postgres)
	if err != nil {
		return fmt.Errorf("postgres.New: %w", err)
	}
	defer deps.Postgres.Close()

	deps.KafkaWriter, err = kafka_writer.New(c.KafkaWriter)
	if err != nil {
		return fmt.Errorf("kafka_writer.New: %w", err)
	}
	defer deps.KafkaWriter.Close()

	deps.Redis, err = redis.New(c.Redis)
	if err != nil {
		return fmt.Errorf("redis.New: %w", err)
	}
	defer deps.Redis.Close()

	// Controllers
	deps.RouterHTTP = router.New(
		router.Check{Name: "postgres", Fn: deps.Postgres.Ping},
		router.Check{Name: "redis", Fn: func(ctx context.Context) error { return deps.Redis.Ping(ctx).Err() }},
	)

	deps.KafkaReader, err = kafka_reader.New(c.KafkaReader)
	if err != nil {
		return fmt.Errorf("kafka_reader.New: %w", err)
	}
	defer deps.KafkaReader.Close()

	// Domains
	var wg sync.WaitGroup

	AppleDomain(ctx, c, deps, &wg)

	httpServer := http_server.New(deps.RouterHTTP, c.HTTP.Port)

	log.Info().Msg("App started!")

	select {
	case <-ctx.Done():
		log.Info().Msg("App got stop signal")
	case err = <-httpServer.Notify():
		log.Error().Err(err).Msg("App got notify: httpServer.Notify")
	}

	log.Info().Msg("App is stopping...")

	// Graceful shutdown: stop accepting requests and finish in-flight ones,
	// then stop workers/consumers; adapters are closed last by the defers above.
	httpServer.Close(context.WithoutCancel(ctx))
	stop()
	wg.Wait()

	return nil
}
