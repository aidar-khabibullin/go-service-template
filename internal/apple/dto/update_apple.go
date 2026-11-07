package dto

import (
	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/google/uuid"
)

type UpdateAppleInput struct {
	ID     uuid.UUID `json:"-"`
	Name   string    `json:"name"`
	Status string    `json:"status"`
}

func (i *UpdateAppleInput) Validate() error {
	if i.Name == "" {
		return entity.ErrNameInvalid
	}

	_, err := entity.ParseStatus(i.Status)
	if err != nil || i.Status == "unknown" {
		return entity.ErrStatusInvalid
	}

	return nil
}
