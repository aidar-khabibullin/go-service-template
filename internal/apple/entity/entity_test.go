package entity_test

import (
	"errors"
	"testing"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/google/uuid"
)

func TestNewApple(t *testing.T) {
	t.Parallel()

	a, err := entity.NewApple("antonovka")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.ID == uuid.Nil || a.Name != "antonovka" || a.Status != entity.StatusNew {
		t.Fatalf("unexpected apple: %+v", a)
	}

	_, err = entity.NewApple("")
	if !errors.Is(err, entity.ErrNameInvalid) {
		t.Fatalf("want ErrNameInvalid, got %v", err)
	}
}

func TestApple_Update(t *testing.T) {
	t.Parallel()

	a, _ := entity.NewApple("a")

	if err := a.Update("b", entity.StatusSuccess); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.Name != "b" || a.Status != entity.StatusSuccess {
		t.Fatalf("not updated: %+v", a)
	}

	if err := a.Update("", entity.StatusError); !errors.Is(err, entity.ErrNameInvalid) {
		t.Fatalf("want ErrNameInvalid, got %v", err)
	}

	if err := a.Update("c", entity.StatusUnknown); !errors.Is(err, entity.ErrStatusInvalid) {
		t.Fatalf("want ErrStatusInvalid, got %v", err)
	}

	if a.Name != "b" {
		t.Fatal("failed update must not change the entity")
	}
}

func TestStatus(t *testing.T) {
	t.Parallel()

	for _, s := range []entity.Status{entity.StatusNew, entity.StatusSuccess, entity.StatusError, entity.StatusUnknown} {
		if got := entity.NewStatus(s.String()); got != s {
			t.Errorf("round trip %q: got %v", s.String(), got)
		}
	}

	if entity.NewStatus("garbage") != entity.StatusUnknown {
		t.Error("garbage must be unknown")
	}

	if _, err := entity.ParseStatus("garbage"); !errors.Is(err, entity.ErrStatusInvalid) {
		t.Errorf("want ErrStatusInvalid, got %v", err)
	}

	if st, err := entity.ParseStatus("success"); err != nil || st != entity.StatusSuccess {
		t.Errorf("got %v, %v", st, err)
	}
}

func TestNewPineApple(t *testing.T) {
	t.Parallel()

	appleID := uuid.New()

	p, err := entity.NewPineApple(appleID, "p")
	if err != nil || p.AppleID != appleID || p.ID == uuid.Nil || p.Status != entity.StatusNew {
		t.Fatalf("unexpected: %+v, %v", p, err)
	}

	if _, err = entity.NewPineApple(appleID, ""); !errors.Is(err, entity.ErrNameInvalid) {
		t.Errorf("want ErrNameInvalid, got %v", err)
	}

	if _, err = entity.NewPineApple(uuid.Nil, "p"); !errors.Is(err, entity.ErrUUIDInvalid) {
		t.Errorf("want ErrUUIDInvalid, got %v", err)
	}
}
