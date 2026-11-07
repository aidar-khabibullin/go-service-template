package config

import (
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/aidar-khabibullin/go-service-template/pkg/http_server"
	"github.com/aidar-khabibullin/go-service-template/pkg/kafka_reader"
	"github.com/aidar-khabibullin/go-service-template/pkg/kafka_writer"
	"github.com/aidar-khabibullin/go-service-template/pkg/logger"
	"github.com/aidar-khabibullin/go-service-template/pkg/metrics"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel"
	"github.com/aidar-khabibullin/go-service-template/pkg/postgres"
	"github.com/aidar-khabibullin/go-service-template/pkg/redis"
	"github.com/aidar-khabibullin/go-service-template/pkg/sentry"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type App struct {
	Name    string `envconfig:"APP_NAME"    required:"true"`
	Version string `envconfig:"APP_VERSION" required:"true"`

	OutboxInterval time.Duration `envconfig:"OUTBOX_INTERVAL" default:"1s"`
}

type Config struct {
	App         App
	HTTP        http_server.Config
	Metrics     metrics.Config
	Postgres    postgres.Config
	KafkaReader kafka_reader.Config
	KafkaWriter kafka_writer.Config
	Redis       redis.Config
	Logger      logger.Config
	Sentry      sentry.Config
	OTEL        otel.Config
}

func New() (Config, error) {
	var config Config

	// .env is optional: in containers/CI variables come from the environment.
	err := godotenv.Load(".env")
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return config, fmt.Errorf("godotenv.Load: %w", err)
	}

	err = envconfig.Process("", &config)
	if err != nil {
		return config, fmt.Errorf("envconfig.Process: %w", err)
	}

	return config, nil
}
