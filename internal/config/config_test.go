//go:build unit

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tests := map[string]struct {
		addr         string
		cdnUploadURL string
		cdnAPIKey    string
		cdnAPISecret string
		want         Config
	}{
		"all vars set": {
			addr:         "127.0.0.1:9090",
			cdnUploadURL: "https://example.com/upload",
			cdnAPIKey:    "key",
			cdnAPISecret: "secret",
			want: Config{
				Addr: "127.0.0.1:9090",
				Cdn: Cdn{
					UploadURL: "https://example.com/upload",
					APIKey:    "key",
					APISecret: "secret",
				},
			},
		},
		"nothing set falls back to defaults": {
			want: Config{
				Addr: ":4040",
				Cdn: Cdn{
					UploadURL: "https://api.cloudinary.com/v1_1/sayze/image/upload",
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("ADDR", tt.addr)
			t.Setenv("CDN_UPLOAD_URL", tt.cdnUploadURL)
			t.Setenv("CDN_API_KEY", tt.cdnAPIKey)
			t.Setenv("CDN_API_SECRET", tt.cdnAPISecret)

			got := Load()

			assert.Equal(t, tt.want, got)
		})
	}
}
