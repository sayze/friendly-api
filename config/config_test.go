package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfiguration(t *testing.T) {
	config := NewConfiguration()
	assert.Equal(t, config, &Configuration{&Http{
		Host: "0.0.0.0",
		Port: "4040",
	}, &Cdn{
		UploadUrl: "https://api.cloudinary.com/v1_1/sayze",
		ApiKey:    "",
		ApiSecret: "",
	}})
}
