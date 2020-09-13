package http

import (
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/go-chi/render"
	validator2 "github.com/go-playground/validator/v10"
	"github.com/sayze/friendly-api/entity"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	FriendService entity.FriendService
	router        chi.Router
	server        *http.Server
}

var validate *validator2.Validate

// NewHandler creates new server instance.
func NewHandler(service entity.FriendService) (*Handler, error) {
	r := chi.NewRouter()

	// Setup router middleware.
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(httprate.Limit(50, 1*time.Minute))

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
				"status":     ww.Status(),
				"took":       latency,
				"remote":     r.RemoteAddr,
				"request":    r.RequestURI,
				"method":     r.Method,
				"request-id": requestID,
			}).Info("API Request")

		})
	})

	handler := &Handler{
		router:        r,
		FriendService: service,
	}

	handler.setupRoutes()

	validate = validator2.New()

	return handler, nil
}

func (h *Handler) setupRoutes() {
	h.router.Get("/friend", h.HandleGetFriend)
	h.router.Get("/friend/{id}", h.HandleGetFriend)
	h.router.Post("/friend", h.HandleCreateFriend)
	h.router.Delete("/friend/{id}", h.HandleDeleteFriend)
	h.router.Patch("/friend", h.HandleUpdateFriend)
}

// ListenAndServe will listen for requests.
func (h *Handler) ListenAndServe(host string, port string) {
	h.server = &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: h.router,
	}

	logrus.Infof("Server started on %s", h.server.Addr)
	logrus.Fatal(h.server.ListenAndServe())
}
