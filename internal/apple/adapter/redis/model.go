package redis

import (
	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/google/uuid"
)

// appleModel is the cache representation of entity.Apple.
// It is separate from the domain entity so the storage format can change independently.
type appleModel struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Status string    `json:"status"`
	Stuffs []string  `json:"stuffs,omitempty"`
}

func toModel(a entity.Apple) appleModel {
	return appleModel{ID: a.ID, Name: a.Name, Status: a.Status.String(), Stuffs: a.Stuffs}
}

func (m appleModel) toEntity() entity.Apple {
	return entity.Apple{ID: m.ID, Name: m.Name, Status: entity.NewStatus(m.Status), Stuffs: m.Stuffs}
}
