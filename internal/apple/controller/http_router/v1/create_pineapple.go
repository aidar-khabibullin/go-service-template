package v1

import (
	"net/http"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/dto"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
	"github.com/aidar-khabibullin/go-service-template/pkg/render"
)

func (h *Handlers) CreatePineApple(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "http/v1 CreatePineApple")
	defer span.End()

	var input dto.CreatePineAppleInput

	err := decodeJSON(w, r, &input)
	if err != nil {
		badRequest("decodeJSON", err, w, "json error")

		return
	}

	err = input.Validate()
	if err != nil {
		badRequest("input.Validate", err, w, "validate error")

		return
	}

	output, err := h.uc.CreatePineApple(ctx, input)
	if err != nil {
		handleError(w, span, "uc.CreatePineApple", err)

		return
	}

	render.JSON(w, output)
}
