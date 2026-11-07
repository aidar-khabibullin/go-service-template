package v1

import (
	"net/http"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/dto"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
	"github.com/aidar-khabibullin/go-service-template/pkg/render"
)

func (h *Handlers) GetApple(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "http/v1 GetApple")
	defer span.End()

	id, err := idParam(r)
	if err != nil {
		badRequest("uuid.Parse", err, w, "uuid validate error")

		return
	}

	output, err := h.uc.GetApple(ctx, dto.GetAppleInput{ID: id})
	if err != nil {
		handleError(w, span, "uc.GetApple", err)

		return
	}

	render.JSON(w, output)
}
