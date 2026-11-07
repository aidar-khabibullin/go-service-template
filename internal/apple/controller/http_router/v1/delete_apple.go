package v1

import (
	"net/http"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/dto"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
)

func (h *Handlers) DeleteApple(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "http/v1 DeleteApple")
	defer span.End()

	id, err := idParam(r)
	if err != nil {
		badRequest("uuid.Parse", err, w, "uuid validate error")

		return
	}

	err = h.uc.DeleteApple(ctx, dto.DeleteAppleInput{ID: id})
	if err != nil {
		handleError(w, span, "uc.DeleteApple", err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
