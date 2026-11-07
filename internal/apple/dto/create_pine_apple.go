package dto

import (
	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/google/uuid"
)

type CreatePineAppleOutput struct {
	ID uuid.UUID `json:"id"`
}

type CreatePineAppleInput struct {
	Name string `json:"name"`
}

func (i *CreatePineAppleInput) Validate() error {
	if i.Name == "" {
		return entity.ErrNameInvalid
	}

	return nil
}
