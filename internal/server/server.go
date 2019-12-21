package server

import (
	validator2 "github.com/go-playground/validator/v10"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/sayze/foundu-taker-api/internal/entity"
	"github.com/sirupsen/logrus"
)

// Server describes api server implementation.
type Server struct {
	router      chi.Router
	server      *http.Server
	friendStore entity.FriendStore
}

var validate *validator2.Validate

// New creates new server instance.
func New(friend entity.FriendStore) (*Server, error) {
	r := chi.NewRouter()

	// Setup router middleware.
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// Configure CORS.
	r.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodHead, http.MethodOptions, http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	}).Handler)

	// Setup request logging.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			var requestID string

			if reqID := r.Context().Value(middleware.RequestIDKey); reqID != nil {
				requestID = reqID.(string)
			}

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			latency := time.Since(start)

			logrus.WithFields(logrus.Fields{
				"status":     ww.Status,
				"took":       latency,
				"remote":     r.RemoteAddr,
				"request":    r.RequestURI,
				"method":     r.Method,
				"request-id": requestID,
			}).Info("API Request")

		})
	})

	server := &Server{
		router:      r,
		friendStore: friend,
		server: &http.Server{
			Addr:    ":" + os.Getenv("PORT"),
			Handler: r,
		},
	}

	server.setupRoutes()

	validate = validator2.New()

	return server, nil
}

// ListenAndServe will listen for requests.
func (s *Server) ListenAndServe() {
	logrus.Infof("Server listening on port %s", s.server.Addr)
	logrus.Fatal(s.server.ListenAndServe())
}
