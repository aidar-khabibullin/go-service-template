package dto

import "github.com/google/uuid"

type GetAppleOutput struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Status string    `json:"status"`
}

type GetAppleInput struct {
	ID uuid.UUID
}
