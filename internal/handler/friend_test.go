//go:build unit

package handler

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sayze/friendly-api/internal/friend/infra/memory"
	"github.com/sayze/friendly-api/internal/friend/service"
)

// noopImages is a domain.ImageStore that returns a deterministic URL and
// never fails, for handler tests that don't exercise upload behavior.
type noopImages struct{}

func (noopImages) Upload(_ context.Context, _ io.Reader, _, id string) (string, error) {
	return "https://cdn.example.com/" + id, nil
}

func (noopImages) Delete(context.Context, string) error { return nil }

func newTestHandler(t *testing.T) *FriendHandler {
	t.Helper()
	svc := service.NewService(memory.NewRepository(), noopImages{})
	return NewFriendHandler(svc)
}

func withURLParam(r *http.Request, key, value string) *http.Request {
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
}

func multipartBody(t *testing.T, fields map[string]string, withImage bool) (io.Reader, string) {
	t.Helper()

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	for k, v := range fields {
		require.NoError(t, w.WriteField(k, v))
	}
	if withImage {
		fw, err := w.CreateFormFile("image", "avatar.png")
		require.NoError(t, err)
		_, err = fw.Write([]byte("image-bytes"))
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())

	return body, w.FormDataContentType()
}

func TestHandleHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	HandleHealth(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"status": "ok"}`, rec.Body.String())
}

func TestFriendHandler_Create(t *testing.T) {
	t.Run("creates a friend without an image", func(t *testing.T) {
		h := newTestHandler(t)
		body, contentType := multipartBody(t, map[string]string{"name": "Alice"}, false)

		req := httptest.NewRequest(http.MethodPost, "/friend", body)
		req.Header.Set("Content-Type", contentType)
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"status": "ok", "data": {"id": 1, "name": "Alice", "image": ""}}`, rec.Body.String())
	})

	t.Run("creates a friend with an image", func(t *testing.T) {
		h := newTestHandler(t)
		body, contentType := multipartBody(t, map[string]string{"name": "Alice"}, true)

		req := httptest.NewRequest(http.MethodPost, "/friend", body)
		req.Header.Set("Content-Type", contentType)
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"status": "ok", "data": {"id": 1, "name": "Alice", "image": "https://cdn.example.com/1"}}`, rec.Body.String())
	})

	t.Run("rejects a name that's too short", func(t *testing.T) {
		h := newTestHandler(t)
		body, contentType := multipartBody(t, map[string]string{"name": "A"}, false)

		req := httptest.NewRequest(http.MethodPost, "/friend", body)
		req.Header.Set("Content-Type", contentType)
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("refuses once the roster limit is reached", func(t *testing.T) {
		repo := memory.NewRepository()
		ctx := context.Background()
		for i := 0; i < service.MaxFriends; i++ {
			_, err := repo.Create(ctx, "Friend", "")
			require.NoError(t, err)
		}
		h := NewFriendHandler(service.NewService(repo, noopImages{}))
		body, contentType := multipartBody(t, map[string]string{"name": "One Too Many"}, false)

		req := httptest.NewRequest(http.MethodPost, "/friend", body)
		req.Header.Set("Content-Type", contentType)
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestFriendHandler_Get(t *testing.T) {
	h := newTestHandler(t)
	body, contentType := multipartBody(t, map[string]string{"name": "Alice"}, false)
	req := httptest.NewRequest(http.MethodPost, "/friend", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	t.Run("found", func(t *testing.T) {
		req := withURLParam(httptest.NewRequest(http.MethodGet, "/friend/1", nil), "id", "1")
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		req := withURLParam(httptest.NewRequest(http.MethodGet, "/friend/999", nil), "id", "999")
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		req := withURLParam(httptest.NewRequest(http.MethodGet, "/friend/abc", nil), "id", "abc")
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestFriendHandler_Update(t *testing.T) {
	h := newTestHandler(t)
	createBody, createContentType := multipartBody(t, map[string]string{"name": "Alice"}, false)
	createReq := httptest.NewRequest(http.MethodPost, "/friend", createBody)
	createReq.Header.Set("Content-Type", createContentType)
	createRec := httptest.NewRecorder()
	h.Create(createRec, createReq)
	require.Equal(t, http.StatusOK, createRec.Code)

	t.Run("updates name", func(t *testing.T) {
		body, contentType := multipartBody(t, map[string]string{"id": "1", "name": "Alice Smith"}, false)
		req := httptest.NewRequest(http.MethodPatch, "/friend", body)
		req.Header.Set("Content-Type", contentType)
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"status": "ok", "data": {"id": 1, "name": "Alice Smith", "image": ""}}`, rec.Body.String())
	})

	t.Run("unknown id is not found", func(t *testing.T) {
		body, contentType := multipartBody(t, map[string]string{"id": "999", "name": "Nobody"}, false)
		req := httptest.NewRequest(http.MethodPatch, "/friend", body)
		req.Header.Set("Content-Type", contentType)
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestFriendHandler_Delete(t *testing.T) {
	h := newTestHandler(t)
	body, contentType := multipartBody(t, map[string]string{"name": "Alice"}, false)
	req := httptest.NewRequest(http.MethodPost, "/friend", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	t.Run("removes an existing friend", func(t *testing.T) {
		req := withURLParam(httptest.NewRequest(http.MethodDelete, "/friend/1", nil), "id", "1")
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("unknown id is not found", func(t *testing.T) {
		req := withURLParam(httptest.NewRequest(http.MethodDelete, "/friend/999", nil), "id", "999")
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestFriendHandler_List(t *testing.T) {
	h := newTestHandler(t)
	for _, name := range []string{"Adam Smith", "Nolan Andrew"} {
		body, contentType := multipartBody(t, map[string]string{"name": name}, false)
		req := httptest.NewRequest(http.MethodPost, "/friend", body)
		req.Header.Set("Content-Type", contentType)
		rec := httptest.NewRecorder()
		h.Create(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/friend?search=adam", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Adam Smith")
	assert.NotContains(t, rec.Body.String(), "Nolan Andrew")
}

func TestRespondJSON(t *testing.T) {
	rec := httptest.NewRecorder()

	respondJSON(rec, http.StatusTeapot, map[string]string{"foo": "bar"})

	assert.Equal(t, http.StatusTeapot, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"foo": "bar"}`, rec.Body.String())
}
