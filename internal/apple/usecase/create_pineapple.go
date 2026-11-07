package usecase

import (
	"context"
	"fmt"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/dto"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
)

// CreatePineApple creates an apple and its pineapple atomically.
func (u *UseCase) CreatePineApple(ctx context.Context, input dto.CreatePineAppleInput) (dto.CreatePineAppleOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase CreatePineApple")
	defer span.End()

	var output dto.CreatePineAppleOutput

	apple, err := entity.NewApple(input.Name)
	if err != nil {
		return output, fmt.Errorf("entity.NewApple: %w", err)
	}

	pineApple, err := entity.NewPineApple(apple.ID, input.Name)
	if err != nil {
		return output, fmt.Errorf("entity.NewPineApple: %w", err)
	}

	err = u.tx.Do(ctx, func(ctx context.Context) error {
		err := u.postgres.CreateApple(ctx, apple)
		if err != nil {
			return fmt.Errorf("u.postgres.CreateApple: %w", err)
		}

		err = u.postgres.CreatePineApple(ctx, pineApple)
		if err != nil {
			return fmt.Errorf("u.postgres.CreatePineApple: %w", err)
		}

		return nil
	})
	if err != nil {
		return output, fmt.Errorf("u.tx.Do: %w", err)
	}

	output.ID = pineApple.ID

	return output, nil
}
