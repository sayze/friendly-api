// Package service implements the friend roster business logic: enforcing
// the roster size limit and orchestrating avatar image storage alongside
// friend persistence. Storage-agnostic - depends only on the domain ports.
package service

import (
	"context"
	"io"
	"strconv"

	"github.com/sayze/friendly-api/internal/friend/domain"
)

//go:generate moq -out mock_gen.go . Service

// MaxFriends is the maximum number of friends the roster will hold.
const MaxFriends = 100

// Service exposes the friend roster business logic that handlers call into.
type Service interface {
	// List returns every friend whose name contains search (case-insensitive
	// substring match). An empty search returns the full roster.
	List(ctx context.Context, search string) ([]*domain.Friend, error)

	// Get returns the friend with the given ID.
	Get(ctx context.Context, id int64) (*domain.Friend, error)

	// Create adds a new friend. image/filename may be nil/empty to create
	// without an avatar. Returns domain.ErrLimitExceeded if the roster is
	// already at MaxFriends.
	Create(ctx context.Context, name string, image io.Reader, filename string) (*domain.Friend, error)

	// Update applies a partial update to the friend with the given ID: an
	// empty name leaves it unchanged, and a nil image leaves the existing
	// avatar unchanged.
	Update(ctx context.Context, id int64, name string, image io.Reader, filename string) (*domain.Friend, error)

	// Delete removes the friend with the given ID, along with its stored
	// avatar image, if any.
	Delete(ctx context.Context, id int64) error
}

type service struct {
	repo   domain.FriendRepository
	images domain.ImageStore
}

// NewService builds a Service backed by the given repository and image
// store.
func NewService(repo domain.FriendRepository, images domain.ImageStore) Service {
	return &service{repo: repo, images: images}
}

func (s *service) List(ctx context.Context, search string) ([]*domain.Friend, error) {
	return s.repo.All(ctx, search)
}

func (s *service) Get(ctx context.Context, id int64) (*domain.Friend, error) {
	return s.repo.Get(ctx, id)
}

func (s *service) Create(ctx context.Context, name string, image io.Reader, filename string) (*domain.Friend, error) {
	count, err := s.repo.Count(ctx)
	if err != nil {
		return nil, err
	}
	if count >= MaxFriends {
		return nil, domain.ErrLimitExceeded
	}

	friend, err := s.repo.Create(ctx, name, "")
	if err != nil {
		return nil, err
	}

	if image == nil {
		return friend, nil
	}

	url, err := s.images.Upload(ctx, image, filename, strconv.FormatInt(friend.ID, 10))
	if err != nil {
		return nil, err
	}

	return s.repo.Update(ctx, friend.ID, "", url)
}

func (s *service) Update(ctx context.Context, id int64, name string, image io.Reader, filename string) (*domain.Friend, error) {
	var newImage string
	if image != nil {
		url, err := s.images.Upload(ctx, image, filename, strconv.FormatInt(id, 10))
		if err != nil {
			return nil, err
		}
		newImage = url
	}

	return s.repo.Update(ctx, id, name, newImage)
}

func (s *service) Delete(ctx context.Context, id int64) error {
	if err := s.images.Delete(ctx, strconv.FormatInt(id, 10)); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
