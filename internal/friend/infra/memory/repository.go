// Package memory implements an in-memory domain.FriendRepository. State
// does not survive a process restart; used for local dev.
package memory

import (
	"context"
	"strings"
	"sync"

	"github.com/sayze/friendly-api/internal/friend/domain"
)

// repo is an in-memory implementation of domain.FriendRepository, backed by
// an insertion-ordered slice so All() results are stable. Safe for
// concurrent use.
type repo struct {
	mu      sync.Mutex
	friends []*domain.Friend
	nextID  int64
}

// NewRepository builds an empty in-memory friend repository.
func NewRepository() domain.FriendRepository {
	return &repo{}
}

func (r *repo) All(_ context.Context, search string) ([]*domain.Friend, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	searchLower := strings.ToLower(search)

	var out []*domain.Friend
	for _, f := range r.friends {
		if searchLower == "" || strings.Contains(strings.ToLower(f.Name), searchLower) {
			cp := *f
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *repo) Get(_ context.Context, id int64) (*domain.Friend, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	f := r.find(id)
	if f == nil {
		return nil, domain.ErrFriendNotFound
	}

	cp := *f
	return &cp, nil
}

func (r *repo) Create(_ context.Context, name, image string) (*domain.Friend, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextID++
	f := &domain.Friend{ID: r.nextID, Name: name, Image: image}
	r.friends = append(r.friends, f)

	cp := *f
	return &cp, nil
}

func (r *repo) Update(_ context.Context, id int64, name, image string) (*domain.Friend, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	f := r.find(id)
	if f == nil {
		return nil, domain.ErrFriendNotFound
	}

	if name != "" {
		f.Name = name
	}
	if image != "" {
		f.Image = image
	}

	cp := *f
	return &cp, nil
}

func (r *repo) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, f := range r.friends {
		if f.ID == id {
			r.friends = append(r.friends[:i], r.friends[i+1:]...)
			return nil
		}
	}
	return domain.ErrFriendNotFound
}

func (r *repo) Count(_ context.Context) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.friends), nil
}

// find returns the friend with the given ID, or nil. Callers must hold r.mu.
func (r *repo) find(id int64) *domain.Friend {
	for _, f := range r.friends {
		if f.ID == id {
			return f
		}
	}
	return nil
}
