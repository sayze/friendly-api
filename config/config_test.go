package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewConfiguration(t *testing.T) {
	config := NewConfiguration()
	assert.Equal(t, config, &Configuration{&Http{
		Host: "localhost",
		Port: "4040",
	}, &Cdn{
		Host: "localhost",
		Port: "6060",
		Base: "friendly",
	}})
}
