// Package handler implements the HTTP transport layer: decoding requests,
// invoking the friend service, and encoding responses. Contains no business
// logic or storage concerns.
package handler

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	validator "github.com/go-playground/validator/v10"

	"github.com/sayze/friendly-api/internal/friend/domain"
	"github.com/sayze/friendly-api/internal/friend/service"
)

var validate = validator.New()

const maxUploadBytes = 10 << 20 // 10MB

// FriendHandler serves the friend roster endpoints, backed by a
// service.Service.
type FriendHandler struct {
	svc service.Service
}

// NewFriendHandler builds a FriendHandler backed by the given service.
func NewFriendHandler(svc service.Service) *FriendHandler {
	return &FriendHandler{svc: svc}
}

type createFriendRequest struct {
	Name string `validate:"required,min=2,max=20"`
}

type updateFriendRequest struct {
	ID   int64  `validate:"required"`
	Name string `validate:"required,min=2,max=20"`
}

// HandleHealth handles GET /.
func HandleHealth(w http.ResponseWriter, _ *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// List handles GET /friend, returning friends whose name matches the
// optional "search" query parameter.
func (h *FriendHandler) List(w http.ResponseWriter, r *http.Request) {
	friends, err := h.svc.List(r.Context(), r.URL.Query().Get("search"))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Internal error.", err)
		return
	}

	respondData(w, friends)
}

// Get handles GET /friend/{id}.
func (h *FriendHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request.", err)
		return
	}

	friend, err := h.svc.Get(r.Context(), id)
	switch {
	case errors.Is(err, domain.ErrFriendNotFound):
		respondNotFound(w, notFoundMessage(id))
	case err != nil:
		respondError(w, http.StatusInternalServerError, "Internal error.", err)
	default:
		respondData(w, friend)
	}
}

// Create handles POST /friend. Expects a multipart form with a "name"
// field and an optional "image" file.
func (h *FriendHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request.", err)
		return
	}

	name := r.FormValue("name")
	if err := validate.Struct(createFriendRequest{Name: name}); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request.", err)
		return
	}

	file, filename, err := formImage(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request.", err)
		return
	}
	if file != nil {
		defer file.Close()
	}

	friend, err := h.svc.Create(r.Context(), name, file, filename)
	switch {
	case errors.Is(err, domain.ErrLimitExceeded):
		respondError(w, http.StatusBadRequest, "server error", err)
	case err != nil:
		respondError(w, http.StatusInternalServerError, "Internal error.", err)
	default:
		respondData(w, friend)
	}
}

// Update handles PATCH /friend. Expects a multipart form with "id", "name"
// fields and an optional "image" file.
func (h *FriendHandler) Update(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request.", err)
		return
	}

	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request.", err)
		return
	}

	name := r.FormValue("name")
	if err := validate.Struct(updateFriendRequest{ID: id, Name: name}); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request.", err)
		return
	}

	file, filename, err := formImage(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request.", err)
		return
	}
	if file != nil {
		defer file.Close()
	}

	friend, err := h.svc.Update(r.Context(), id, name, file, filename)
	switch {
	case errors.Is(err, domain.ErrFriendNotFound):
		respondNotFound(w, notFoundMessage(id))
	case err != nil:
		respondError(w, http.StatusInternalServerError, "Internal error.", err)
	default:
		respondData(w, friend)
	}
}

// Delete handles DELETE /friend/{id}.
func (h *FriendHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request.", err)
		return
	}

	err = h.svc.Delete(r.Context(), id)
	switch {
	case errors.Is(err, domain.ErrFriendNotFound):
		respondNotFound(w, notFoundMessage(id))
	case err != nil:
		respondError(w, http.StatusInternalServerError, "Internal error.", err)
	default:
		respondData(w, "Friend removed successfully")
	}
}

// parseID extracts and validates the "id" URL parameter.
func parseID(r *http.Request) (int64, error) {
	raw := chi.URLParam(r, "id")
	if err := validate.Var(raw, "required,numeric"); err != nil {
		return 0, errors.New("invalid friend id provided")
	}
	return strconv.ParseInt(raw, 10, 64)
}

// formImage extracts the optional "image" file from a parsed multipart
// form. Returns a nil file and empty filename if no image was provided.
func formImage(r *http.Request) (multipart.File, string, error) {
	file, header, err := r.FormFile("image")
	switch {
	case err == nil:
		return file, header.Filename, nil
	case errors.Is(err, http.ErrMissingFile):
		return nil, "", nil
	default:
		return nil, "", err
	}
}

func notFoundMessage(id int64) string {
	return "Could not find friend with id " + strconv.FormatInt(id, 10)
}

func respondJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func respondData(w http.ResponseWriter, data any) {
	respondJSON(w, http.StatusOK, map[string]any{"status": "ok", "data": data})
}

func respondNotFound(w http.ResponseWriter, message string) {
	respondJSON(w, http.StatusNotFound, map[string]any{"status": "resource not found", "data": message})
}

func respondError(w http.ResponseWriter, status int, statusText string, err error) {
	respondJSON(w, status, map[string]any{"status": statusText, "error": err.Error()})
}
