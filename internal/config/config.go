// Package config loads friendly-api's configuration from environment
// variables.
package config

import "os"

// Config holds all environment-derived configuration.
type Config struct {
	// Addr is the address the HTTP server listens on, e.g. ":4040".
	Addr string

	Cdn Cdn
}

// Cdn holds the Cloudinary account settings used to store friend avatar
// images.
type Cdn struct {
	UploadURL string
	APIKey    string
	APISecret string
}

// Load reads configuration from the environment, applying defaults for
// anything unset.
func Load() Config {
	return Config{
		Addr: getEnv("ADDR", ":4040"),
		Cdn: Cdn{
			UploadURL: getEnv("CDN_UPLOAD_URL", "https://api.cloudinary.com/v1_1/sayze/image/upload"),
			APIKey:    os.Getenv("CDN_API_KEY"),
			APISecret: os.Getenv("CDN_API_SECRET"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
