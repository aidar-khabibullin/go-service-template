package entity

import (
	"github.com/google/uuid"
)

// Apple is the single domain model. It is created through NewApple (validation)
// and restored from storage as a plain struct literal by adapters.
type Apple struct {
	ID     uuid.UUID
	Name   string
	Status Status
	Stuffs []string
}

// NewApple validates input and creates a new Apple with StatusNew.
func NewApple(name string) (Apple, error) {
	if name == "" {
		return Apple{}, ErrNameInvalid
	}

	return Apple{
		ID:     uuid.New(),
		Name:   name,
		Status: StatusNew,
	}, nil
}

// Update validates and applies new name and status.
func (a *Apple) Update(name string, status Status) error {
	if name == "" {
		return ErrNameInvalid
	}

	if status == StatusUnknown {
		return ErrStatusInvalid
	}

	a.Name = name
	a.Status = status

	return nil
}
