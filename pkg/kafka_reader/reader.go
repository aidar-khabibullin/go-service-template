package kafka_reader

import (
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

type Config struct {
	Addr  []string `envconfig:"KAFKA_READER_ADDR" required:"true"`
	Group string   `envconfig:"KAFKA_READER_GROUP" required:"true"`
	Topic string   `envconfig:"KAFKA_READER_TOPIC" required:"true"`
}

type Reader struct {
	*kafka.Reader
}

func New(c Config) (*Reader, error) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.Addr,
		GroupID:  c.Group,
		Topic:    c.Topic,
		MaxBytes: 10e6, // 10MB
	})

	return &Reader{Reader: r}, nil
}

func (r *Reader) Close() {
	err := r.Reader.Close()
	if err != nil {
		log.Error().Err(err).Msg("kafka_reader.Close")
	}

	log.Info().Msg("Kafka reader closed")
}
