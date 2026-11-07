package render

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// JSON writes output as a JSON response with status 200.
func JSON(w http.ResponseWriter, output any) {
	b, err := json.Marshal(output)
	if err != nil {
		log.Error().Err(err).Msg("render.JSON json.Marshal")
		http.Error(w, "marshal json error", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	_, err = w.Write(b)
	if err != nil {
		// Headers are already sent, the only thing left is to log.
		log.Error().Err(err).Msg("render.JSON w.Write")
	}
}
