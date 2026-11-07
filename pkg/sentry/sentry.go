package sentry

import (
	"github.com/rs/zerolog/log"
)

type Config struct {
	DSN string `envconfig:"SENTRY_DSN"`
}

func Init(c Config) error {
	if c.DSN == "" {
		log.Info().Msg("Sentry is disabled")

		return nil
	}

	log.Info().Msg("Sentry initialized")

	return nil
}

func Close() {
}
