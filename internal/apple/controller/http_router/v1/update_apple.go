package v1

import (
	"net/http"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/dto"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
)

func (h *Handlers) UpdateApple(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "http/v1 UpdateApple")
	defer span.End()

	id, err := idParam(r)
	if err != nil {
		badRequest("uuid.Parse", err, w, "uuid validate error")

		return
	}

	input := dto.UpdateAppleInput{ID: id}

	err = decodeJSON(w, r, &input)
	if err != nil {
		badRequest("decodeJSON", err, w, "json error")

		return
	}

	err = input.Validate()
	if err != nil {
		badRequest("input.Validate", err, w, "validate error")

		return
	}

	err = h.uc.UpdateApple(ctx, input)
	if err != nil {
		handleError(w, span, "uc.UpdateApple", err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
