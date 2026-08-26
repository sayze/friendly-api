//go:build unit

package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sayze/friendly-api/internal/friend/infra/memory"
	"github.com/sayze/friendly-api/internal/friend/service"
)

type noopImages struct{}

func (noopImages) Upload(context.Context, io.Reader, string, string) (string, error) { return "", nil }
func (noopImages) Delete(context.Context, string) error                              { return nil }

func newTestServer() http.Handler {
	return New(service.NewService(memory.NewRepository(), noopImages{}))
}

func TestNew_HealthCheck(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"status": "ok"}`, rec.Body.String())
}

func TestNew_UnknownRoute(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestNew_CORS(t *testing.T) {
	srv := newTestServer()

	t.Run("preflight is answered", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/friend", nil)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Access-Control-Request-Method", http.MethodGet)
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("actual response carries the allow-origin header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://example.com")
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestNew_FriendRoutes(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/friend", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"status": "ok", "data": null}`, rec.Body.String())
}
