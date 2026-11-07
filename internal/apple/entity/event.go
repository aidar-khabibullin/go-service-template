package entity

import "github.com/google/uuid"

type CreateEvent struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// OutboxEvent is a message stored in the same transaction as the business data
// and delivered to Kafka later by the outbox worker.
type OutboxEvent struct {
	ID      uuid.UUID
	Key     string
	Payload []byte
}
