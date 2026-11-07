package kafka_writer

import (
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

type Config struct {
	Addr  []string `envconfig:"KAFKA_WRITER_ADDR" required:"true"`
	Topic string   `envconfig:"KAFKA_WRITER_TOPIC" required:"true"`
}

type Writer struct {
	*kafka.Writer
}

func New(c Config) (*Writer, error) {
	w := &kafka.Writer{
		Addr:     kafka.TCP(c.Addr...),
		Topic:    c.Topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &Writer{Writer: w}, nil
}

func (w *Writer) Close() {
	err := w.Writer.Close()
	if err != nil {
		log.Error().Err(err).Msg("kafka_writer.Close")
	}

	log.Info().Msg("Kafka writer closed")
}
