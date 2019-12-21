// Package memory implements an in memory storage interface for all the entities.
package memory

import (
	"github.com/sayze/foundu-taker-api/internal/entity"
	"time"
)

// Client contains properties for a memory storage client.
type Client struct {
	id      int64
	friends []*entity.Friend
}

// New creates a new in memory client.
func New() (*Client, error) {
	return &Client{id: time.Now().Unix()}, nil
}
