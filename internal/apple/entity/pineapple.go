package entity

import (
	"github.com/google/uuid"
)

// PineApple belongs to an Apple.
type PineApple struct {
	ID      uuid.UUID
	AppleID uuid.UUID
	Name    string
	Status  Status
}

// NewPineApple validates and creates a new PineApple for the given apple.
func NewPineApple(appleID uuid.UUID, name string) (PineApple, error) {
	if name == "" {
		return PineApple{}, ErrNameInvalid
	}

	if appleID == uuid.Nil {
		return PineApple{}, ErrUUIDInvalid
	}

	return PineApple{
		ID:      uuid.New(),
		AppleID: appleID,
		Name:    name,
		Status:  StatusNew,
	}, nil
}
