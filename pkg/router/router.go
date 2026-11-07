package router

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

const readyTimeout = 2 * time.Second

// Check is a named dependency probe used by the readiness endpoint.
type Check struct {
	Name string
	Fn   func(ctx context.Context) error
}

func New(readiness ...Check) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Get("/live", live)
	r.Get("/ready", ready(readiness))

	return r
}

// live reports that the process is up. It must not check dependencies.
func live(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// ready reports whether the app can serve traffic: every dependency must answer.
func ready(checks []Check) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readyTimeout)
		defer cancel()

		for _, c := range checks {
			err := c.Fn(ctx)
			if err != nil {
				log.Error().Err(err).Str("dependency", c.Name).Msg("router: readiness check failed")
				http.Error(w, c.Name+" unavailable", http.StatusServiceUnavailable)

				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
