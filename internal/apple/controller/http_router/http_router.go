package http_router

import (
	ver1 "github.com/aidar-khabibullin/go-service-template/internal/apple/controller/http_router/v1"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/riandyrn/otelchi"
)

func AppleRouter(r *chi.Mux, uc *usecase.UseCase) {
	r.Route("/api/apple", func(r chi.Router) {
		r.Use(otelchi.Middleware("Apple", otelchi.WithChiRoutes(r)))

		v1 := ver1.New(uc)

		r.Route("/v1", func(r chi.Router) {
			r.Post("/create_apple", v1.CreateApple)
			r.Post("/create_pineapple", v1.CreatePineApple)
			r.Get("/get_apple/{id}", v1.GetApple)
			r.Put("/update_apple/{id}", v1.UpdateApple)
			r.Delete("/delete_apple/{id}", v1.DeleteApple)
		})
	})
}
