package kafka_producer

import (
	"context"
	"fmt"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
	"github.com/segmentio/kafka-go"
)

// Send writes outbox events to Kafka in a single batch.
func (p *Producer) Send(ctx context.Context, events ...entity.OutboxEvent) error {
	ctx, span := tracer.Start(ctx, "kafka_producer Send")
	defer span.End()

	msgs := make([]kafka.Message, 0, len(events))
	for _, e := range events {
		msgs = append(msgs, kafka.Message{
			Key:   []byte(e.Key),
			Value: e.Payload,
		})
	}

	err := p.writer.WriteMessages(ctx, msgs...)
	if err != nil {
		return fmt.Errorf("p.writer.WriteMessages: %w", err)
	}

	return nil
}
