package domain

import (
	"context"
	"errors"
	"io"
)

//go:generate moq -out mock_gen.go . FriendRepository ImageStore

// ErrFriendNotFound is returned when an operation references a friend ID
// that isn't in the roster.
var ErrFriendNotFound = errors.New("friend not found")

// ErrLimitExceeded is returned when creating a friend would exceed the
// maximum roster size.
var ErrLimitExceeded = errors.New("maximum friends exceeded")

// Friend is a person tracked in the roster, with an optional avatar image.
type Friend struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

// FriendRepository defines the storage contract for friends.
type FriendRepository interface {
	// All returns every friend whose name contains search (case-insensitive).
	// An empty search returns the full roster.
	All(ctx context.Context, search string) ([]*Friend, error)

	// Get returns the friend with the given ID, or ErrFriendNotFound.
	Get(ctx context.Context, id int64) (*Friend, error)

	// Create adds a new friend with the given name and image, assigning it a
	// new ID.
	Create(ctx context.Context, name, image string) (*Friend, error)

	// Update applies a partial update to the friend with the given ID: an
	// empty name or image leaves that field unchanged. Returns
	// ErrFriendNotFound if no friend has the given ID.
	Update(ctx context.Context, id int64, name, image string) (*Friend, error)

	// Delete removes the friend with the given ID, or returns
	// ErrFriendNotFound if none exists.
	Delete(ctx context.Context, id int64) error

	// Count returns the number of friends currently stored.
	Count(ctx context.Context) (int, error)
}

// ImageStore defines the contract for storing and removing a friend's
// avatar image in external image hosting.
type ImageStore interface {
	// Upload stores img (read from filename) under id and returns its public
	// URL.
	Upload(ctx context.Context, img io.Reader, filename, id string) (url string, err error)

	// Delete removes the image stored under id, if any.
	Delete(ctx context.Context, id string) error
}
