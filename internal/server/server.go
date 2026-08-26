// Package server wires up the chi router: middleware and route
// registration. Contains no business logic.
package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"github.com/sayze/friendly-api/internal/friend/service"
	"github.com/sayze/friendly-api/internal/handler"
)

// requestsPerMinute caps each client to this many requests per minute.
const requestsPerMinute = 50

// New builds the chi router for the friend roster API, backed by svc.
func New(svc service.Service) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromHeader("CF-Connecting-IP"))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(httprate.LimitByIP(requestsPerMinute, time.Minute))

	fh := handler.NewFriendHandler(svc)

	r.Get("/", handler.HandleHealth)
	r.Get("/friend", fh.List)
	r.Get("/friend/{id}", fh.Get)
	r.Post("/friend", fh.Create)
	r.Patch("/friend", fh.Update)
	r.Delete("/friend/{id}", fh.Delete)

	return r
}
