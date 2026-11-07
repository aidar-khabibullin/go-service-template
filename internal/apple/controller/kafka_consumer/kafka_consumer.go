package kafka_consumer

import (
	"context"
	"errors"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/usecase"
	"github.com/aidar-khabibullin/go-service-template/pkg/kafka_reader"
	"github.com/aidar-khabibullin/go-service-template/pkg/metrics"
	"github.com/rs/zerolog/log"
)

// AppleConsumer reads messages until ctx is cancelled or the reader is closed.
func AppleConsumer(ctx context.Context, reader *kafka_reader.Reader, _ *usecase.UseCase) {
	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Error().Err(err).Msg("kafka_consumer.AppleConsumer: reader.FetchMessage")
			}

			return
		}

		metrics.KafkaConsumed(m.Topic)

		// UseCase call here

		log.Info().
			Str("topic", m.Topic).
			Int("partition", m.Partition).
			Int64("offset", m.Offset).
			Str("key", string(m.Key)).
			Str("value", string(m.Value)).
			Msg("kafka_consumer.AppleConsumer: reader.FetchMessage")

		if err = reader.CommitMessages(ctx, m); err != nil {
			log.Error().Err(err).Msg("kafka_consumer.AppleConsumer: reader.CommitMessages")
		}
	}
}
