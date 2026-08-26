//go:build unit

package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sayze/friendly-api/internal/friend/domain"
)

// fakeRepo is a minimal, hand-rolled domain.FriendRepository double letting
// tests observe calls and force errors without a mocking framework.
type fakeRepo struct {
	allFn    func(ctx context.Context, search string) ([]*domain.Friend, error)
	getFn    func(ctx context.Context, id int64) (*domain.Friend, error)
	createFn func(ctx context.Context, name, image string) (*domain.Friend, error)
	updateFn func(ctx context.Context, id int64, name, image string) (*domain.Friend, error)
	deleteFn func(ctx context.Context, id int64) error
	countFn  func(ctx context.Context) (int, error)
}

func (f *fakeRepo) All(ctx context.Context, search string) ([]*domain.Friend, error) {
	return f.allFn(ctx, search)
}

func (f *fakeRepo) Get(ctx context.Context, id int64) (*domain.Friend, error) {
	return f.getFn(ctx, id)
}

func (f *fakeRepo) Create(ctx context.Context, name, image string) (*domain.Friend, error) {
	return f.createFn(ctx, name, image)
}

func (f *fakeRepo) Update(ctx context.Context, id int64, name, image string) (*domain.Friend, error) {
	return f.updateFn(ctx, id, name, image)
}

func (f *fakeRepo) Delete(ctx context.Context, id int64) error {
	return f.deleteFn(ctx, id)
}

func (f *fakeRepo) Count(ctx context.Context) (int, error) {
	return f.countFn(ctx)
}

// fakeImages is a minimal, hand-rolled domain.ImageStore double.
type fakeImages struct {
	uploadFn func(ctx context.Context, img io.Reader, filename, id string) (string, error)
	deleteFn func(ctx context.Context, id string) error
}

func (f *fakeImages) Upload(ctx context.Context, img io.Reader, filename, id string) (string, error) {
	return f.uploadFn(ctx, img, filename, id)
}

func (f *fakeImages) Delete(ctx context.Context, id string) error {
	return f.deleteFn(ctx, id)
}

func TestService_Create(t *testing.T) {
	t.Run("creates without an image", func(t *testing.T) {
		repo := &fakeRepo{
			countFn: func(context.Context) (int, error) { return 0, nil },
			createFn: func(_ context.Context, name, image string) (*domain.Friend, error) {
				assert.Equal(t, "Alice", name)
				assert.Empty(t, image)
				return &domain.Friend{ID: 1, Name: name}, nil
			},
		}
		images := &fakeImages{
			uploadFn: func(context.Context, io.Reader, string, string) (string, error) {
				t.Fatal("Upload should not be called without an image")
				return "", nil
			},
		}
		svc := NewService(repo, images)

		got, err := svc.Create(context.Background(), "Alice", nil, "")

		require.NoError(t, err)
		assert.Equal(t, &domain.Friend{ID: 1, Name: "Alice"}, got)
	})

	t.Run("uploads image and persists the resulting URL", func(t *testing.T) {
		var uploadedID string
		repo := &fakeRepo{
			countFn: func(context.Context) (int, error) { return 0, nil },
			createFn: func(_ context.Context, name, image string) (*domain.Friend, error) {
				return &domain.Friend{ID: 7, Name: name}, nil
			},
			updateFn: func(_ context.Context, id int64, name, image string) (*domain.Friend, error) {
				assert.Equal(t, int64(7), id)
				assert.Empty(t, name)
				assert.Equal(t, "https://cdn.example.com/7", image)
				return &domain.Friend{ID: 7, Name: "Alice", Image: image}, nil
			},
		}
		images := &fakeImages{
			uploadFn: func(_ context.Context, img io.Reader, filename, id string) (string, error) {
				uploadedID = id
				assert.Equal(t, "avatar.png", filename)
				b, _ := io.ReadAll(img)
				assert.Equal(t, "image-bytes", string(b))
				return "https://cdn.example.com/7", nil
			},
		}
		svc := NewService(repo, images)

		got, err := svc.Create(context.Background(), "Alice", strings.NewReader("image-bytes"), "avatar.png")

		require.NoError(t, err)
		assert.Equal(t, "7", uploadedID)
		assert.Equal(t, "https://cdn.example.com/7", got.Image)
	})

	t.Run("refuses to exceed the roster limit", func(t *testing.T) {
		repo := &fakeRepo{
			countFn: func(context.Context) (int, error) { return MaxFriends, nil },
			createFn: func(context.Context, string, string) (*domain.Friend, error) {
				t.Fatal("Create should not be called once the limit is reached")
				return nil, nil
			},
		}
		svc := NewService(repo, &fakeImages{})

		_, err := svc.Create(context.Background(), "Alice", nil, "")

		assert.ErrorIs(t, err, domain.ErrLimitExceeded)
	})

	t.Run("propagates upload errors without persisting a friend", func(t *testing.T) {
		wantErr := errors.New("upload failed")
		repo := &fakeRepo{
			countFn: func(context.Context) (int, error) { return 0, nil },
			createFn: func(_ context.Context, name, image string) (*domain.Friend, error) {
				return &domain.Friend{ID: 1, Name: name}, nil
			},
			updateFn: func(context.Context, int64, string, string) (*domain.Friend, error) {
				t.Fatal("Update should not be called when upload fails")
				return nil, nil
			},
		}
		images := &fakeImages{
			uploadFn: func(context.Context, io.Reader, string, string) (string, error) {
				return "", wantErr
			},
		}
		svc := NewService(repo, images)

		_, err := svc.Create(context.Background(), "Alice", strings.NewReader("x"), "a.png")

		assert.ErrorIs(t, err, wantErr)
	})
}

func TestService_Update(t *testing.T) {
	t.Run("without an image, image field is left empty", func(t *testing.T) {
		repo := &fakeRepo{
			updateFn: func(_ context.Context, id int64, name, image string) (*domain.Friend, error) {
				assert.Equal(t, int64(1), id)
				assert.Equal(t, "New Name", name)
				assert.Empty(t, image)
				return &domain.Friend{ID: id, Name: name}, nil
			},
		}
		svc := NewService(repo, &fakeImages{})

		_, err := svc.Update(context.Background(), 1, "New Name", nil, "")

		require.NoError(t, err)
	})

	t.Run("uploads and applies a new image", func(t *testing.T) {
		repo := &fakeRepo{
			updateFn: func(_ context.Context, id int64, name, image string) (*domain.Friend, error) {
				assert.Equal(t, "https://cdn.example.com/1", image)
				return &domain.Friend{ID: id, Name: name, Image: image}, nil
			},
		}
		images := &fakeImages{
			uploadFn: func(context.Context, io.Reader, string, string) (string, error) {
				return "https://cdn.example.com/1", nil
			},
		}
		svc := NewService(repo, images)

		got, err := svc.Update(context.Background(), 1, "", strings.NewReader("x"), "a.png")

		require.NoError(t, err)
		assert.Equal(t, "https://cdn.example.com/1", got.Image)
	})

	t.Run("propagates not found", func(t *testing.T) {
		repo := &fakeRepo{
			updateFn: func(context.Context, int64, string, string) (*domain.Friend, error) {
				return nil, domain.ErrFriendNotFound
			},
		}
		svc := NewService(repo, &fakeImages{})

		_, err := svc.Update(context.Background(), 999, "New Name", nil, "")

		assert.ErrorIs(t, err, domain.ErrFriendNotFound)
	})
}

func TestService_Delete(t *testing.T) {
	t.Run("deletes the image before the record", func(t *testing.T) {
		var order []string
		repo := &fakeRepo{
			deleteFn: func(_ context.Context, id int64) error {
				order = append(order, "repo")
				assert.Equal(t, int64(1), id)
				return nil
			},
		}
		images := &fakeImages{
			deleteFn: func(_ context.Context, id string) error {
				order = append(order, "image")
				assert.Equal(t, "1", id)
				return nil
			},
		}
		svc := NewService(repo, images)

		err := svc.Delete(context.Background(), 1)

		require.NoError(t, err)
		assert.Equal(t, []string{"image", "repo"}, order)
	})

	t.Run("does not delete the record if image deletion fails", func(t *testing.T) {
		wantErr := errors.New("image delete failed")
		repo := &fakeRepo{
			deleteFn: func(context.Context, int64) error {
				t.Fatal("repo.Delete should not be called when image deletion fails")
				return nil
			},
		}
		images := &fakeImages{
			deleteFn: func(context.Context, string) error {
				return wantErr
			},
		}
		svc := NewService(repo, images)

		err := svc.Delete(context.Background(), 1)

		assert.ErrorIs(t, err, wantErr)
	})

	t.Run("propagates not found from the repository", func(t *testing.T) {
		repo := &fakeRepo{
			deleteFn: func(context.Context, int64) error {
				return domain.ErrFriendNotFound
			},
		}
		images := &fakeImages{
			deleteFn: func(context.Context, string) error { return nil },
		}
		svc := NewService(repo, images)

		err := svc.Delete(context.Background(), 999)

		assert.ErrorIs(t, err, domain.ErrFriendNotFound)
	})
}

func TestService_List(t *testing.T) {
	repo := &fakeRepo{
		allFn: func(_ context.Context, search string) ([]*domain.Friend, error) {
			assert.Equal(t, "ali", search)
			return []*domain.Friend{{ID: 1, Name: "Alice"}}, nil
		},
	}
	svc := NewService(repo, &fakeImages{})

	got, err := svc.List(context.Background(), "ali")

	require.NoError(t, err)
	assert.Equal(t, []*domain.Friend{{ID: 1, Name: "Alice"}}, got)
}

func TestService_Get(t *testing.T) {
	repo := &fakeRepo{
		getFn: func(_ context.Context, id int64) (*domain.Friend, error) {
			if id == 1 {
				return &domain.Friend{ID: 1, Name: "Alice"}, nil
			}
			return nil, domain.ErrFriendNotFound
		},
	}
	svc := NewService(repo, &fakeImages{})

	got, err := svc.Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Alice", got.Name)

	_, err = svc.Get(context.Background(), 999)
	assert.ErrorIs(t, err, domain.ErrFriendNotFound)
}
