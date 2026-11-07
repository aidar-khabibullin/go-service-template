package v1

import (
	"net/http"

	"github.com/aidar-khabibullin/go-service-template/internal/apple/dto"
	"github.com/aidar-khabibullin/go-service-template/pkg/otel/tracer"
	"github.com/aidar-khabibullin/go-service-template/pkg/render"
)

func (h *Handlers) CreateApple(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "http/v1 CreateApple")
	defer span.End()

	var input dto.CreateAppleInput

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

	output, err := h.uc.CreateApple(ctx, input)
	if err != nil {
		handleError(w, span, "uc.CreateApple", err)

		return
	}

	render.JSON(w, output)
}
