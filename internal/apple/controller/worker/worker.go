package worker

import (
	"context"
	"time"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/usecase"
	"github.com/rs/zerolog/log"
)

// OutboxWorker periodically delivers outbox events to Kafka until ctx is cancelled.
func OutboxWorker(ctx context.Context, uc *usecase.UseCase, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Info().Msg("worker.OutboxWorker started")

	for {
		// Drain the outbox: keep sending while there are full batches.
		for {
			sent, err := uc.SendOutbox(ctx)
			if err != nil {
				if ctx.Err() == nil {
					log.Error().Err(err).Msg("worker.OutboxWorker: uc.SendOutbox")
				}

				break
			}

			if sent == 0 {
				break
			}
		}

		select {
		case <-ctx.Done():
			log.Info().Msg("worker.OutboxWorker stopped")

			return
		case <-ticker.C:
		}
	}
}
