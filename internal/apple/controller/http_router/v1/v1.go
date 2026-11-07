package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/usecase"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"

	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
)

// maxBodySize limits request bodies to protect against memory exhaustion.
const maxBodySize = 1 << 20 // 1 MiB

type Handlers struct {
	uc *usecase.UseCase
}

func New(uc *usecase.UseCase) *Handlers {
	return &Handlers{
		uc: uc,
	}
}

// decodeJSON reads a size-limited JSON body and rejects unknown fields.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodySize))
	dec.DisallowUnknownFields()

	err := dec.Decode(v)
	if err != nil {
		return fmt.Errorf("decode json: %w", err)
	}

	return nil
}

// idParam parses the {id} URL parameter.
func idParam(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, fmt.Errorf("uuid.Parse: %w", err)
	}

	return id, nil
}

// handleError maps a use case error to an HTTP response. It is the single place
// where domain errors become status codes.
func handleError(w http.ResponseWriter, span trace.Span, op string, err error) {
	switch {
	case errors.Is(err, entity.ErrNotFound):
		log.Warn().Err(err).Msg(op + ": not found")
		http.Error(w, "not found", http.StatusNotFound)

	case errors.Is(err, entity.ErrUUIDInvalid),
		errors.Is(err, entity.ErrStatusInvalid),
		errors.Is(err, entity.ErrNameInvalid):
		log.Warn().Err(err).Msg(op + ": validate error")
		http.Error(w, "validate error", http.StatusBadRequest)

	default:
		log.Error().Err(err).Msg(op + ": internal error")
		http.Error(w, "internal error", http.StatusInternalServerError)

		tracer.SetStatus(span, err)
		span.RecordError(err)
	}
}

func badRequest(op string, err error, w http.ResponseWriter, msg string) {
	log.Warn().Err(err).Msg(op)
	http.Error(w, msg, http.StatusBadRequest)
}
