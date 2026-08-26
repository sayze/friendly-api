//go:build unit

package cloudinary

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSign(t *testing.T) {
	got, err := sign("1700000000", "42", "shh")
	require.NoError(t, err)
	assert.NotEmpty(t, got)

	// Signing must be deterministic for the same inputs.
	again, err := sign("1700000000", "42", "shh")
	require.NoError(t, err)
	assert.Equal(t, got, again)

	// A different secret must produce a different signature.
	other, err := sign("1700000000", "42", "different")
	require.NoError(t, err)
	assert.NotEqual(t, got, other)
}

func TestTrimExt(t *testing.T) {
	tests := map[string]string{
		"avatar.png":    "avatar",
		"avatar.tar.gz": "avatar.tar",
		"no-extension":  "no-extension",
		".hidden":       "",
	}

	for filename, want := range tests {
		assert.Equal(t, want, trimExt(filename))
	}
}

func TestImageStore_Upload(t *testing.T) {
	var gotMethod, gotPublicID, gotAPIKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		require.NoError(t, r.ParseMultipartForm(1<<20))
		gotPublicID = r.FormValue("public_id")
		gotAPIKey = r.FormValue("api_key")

		file, _, err := r.FormFile("file")
		require.NoError(t, err)
		defer file.Close()

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"secure_url": "https://cdn.example.com/42.png"}`))
	}))
	defer srv.Close()

	store := NewImageStore(Config{UploadURL: srv.URL, APIKey: "key", APISecret: "secret"})
	store.(*imageStore).now = func() time.Time { return time.Unix(1700000000, 0) }

	url, err := store.Upload(context.Background(), strings.NewReader("image-bytes"), "avatar.png", "42")

	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/42.png", url)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "42", gotPublicID)
	assert.Equal(t, "key", gotAPIKey)
}

func TestImageStore_Upload_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "invalid signature"}`))
	}))
	defer srv.Close()

	store := NewImageStore(Config{UploadURL: srv.URL, APIKey: "key", APISecret: "secret"})

	_, err := store.Upload(context.Background(), strings.NewReader("x"), "a.png", "1")

	assert.Error(t, err)
}

func TestImageStore_Delete(t *testing.T) {
	var gotPublicID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseMultipartForm(1<<20))
		gotPublicID = r.FormValue("public_id")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	store := NewImageStore(Config{UploadURL: srv.URL, APIKey: "key", APISecret: "secret"})

	err := store.Delete(context.Background(), "42")

	require.NoError(t, err)
	assert.Equal(t, "42", gotPublicID)
}

func TestImageStore_Delete_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	store := NewImageStore(Config{UploadURL: srv.URL, APIKey: "key", APISecret: "secret"})

	err := store.Delete(context.Background(), "42")

	assert.Error(t, err)
}
