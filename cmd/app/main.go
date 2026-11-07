package main

import (
	"context"
	"os"

	"github.com/aidar-khabibullin/go-service-template/config"
	"github.com/aidar-khabibullin/go-service-template/internal/app"
	"github.com/aidar-khabibullin/go-service-template/pkg/logger"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel"
	"github.com/aidar-khabibullin/go-service-template/pkg/sentry"
	"github.com/rs/zerolog/log"
	_ "go.uber.org/automaxprocs"
)

func main() {
	// os.Exit does not run defers, so all work lives in run().
	os.Exit(run())
}

func run() int {
	ctx := context.Background()

	c, err := config.New()
	if err != nil {
		log.Error().Err(err).Msg("config.New")

		return 1
	}

	logger.Init(c.Logger)

	err = sentry.Init(c.Sentry)
	if err != nil {
		log.Error().Err(err).Msg("sentry.Init")
	}

	defer sentry.Close()

	err = otel.Init(ctx, c.OTEL)
	if err != nil {
		log.Error().Err(err).Msg("otel.Init")
	}

	defer otel.Close()

	err = app.Run(ctx, c)
	if err != nil {
		log.Error().Err(err).Msg("app.Run")

		return 1
	}

	log.Info().Msg("App stopped!")

	return 0
}
