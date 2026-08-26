//go:build unit

package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sayze/friendly-api/internal/friend/domain"
)

func TestRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository()

	first, err := repo.Create(ctx, "Adam Smith", "")
	require.NoError(t, err)
	assert.Equal(t, &domain.Friend{ID: 1, Name: "Adam Smith"}, first)

	second, err := repo.Create(ctx, "Nolan Andrew", "fake23")
	require.NoError(t, err)
	assert.Equal(t, &domain.Friend{ID: 2, Name: "Nolan Andrew", Image: "fake23"}, second)
}

func TestRepository_Create_ReusesFreedIDsSafely(t *testing.T) {
	// Regression test: IDs must never be derived from the current slice
	// length, since deleting an earlier friend would shrink the length and
	// cause a freshly created friend to collide with an existing ID.
	ctx := context.Background()
	repo := NewRepository()

	_, err := repo.Create(ctx, "Adam Smith", "")
	require.NoError(t, err)
	second, err := repo.Create(ctx, "Nolan Andrew", "")
	require.NoError(t, err)
	_, err = repo.Create(ctx, "Russel Evans", "")
	require.NoError(t, err)

	require.NoError(t, repo.Delete(ctx, second.ID))

	fourth, err := repo.Create(ctx, "Priya Patel", "")
	require.NoError(t, err)
	assert.Equal(t, int64(4), fourth.ID)
}

func TestRepository_Get(t *testing.T) {
	ctx := context.Background()
	repo := seedRepo(t)

	got, err := repo.Get(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, "Nolan Andrew", got.Name)

	_, err = repo.Get(ctx, 999)
	assert.ErrorIs(t, err, domain.ErrFriendNotFound)
}

func TestRepository_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("updates name and image", func(t *testing.T) {
		repo := seedRepo(t)

		got, err := repo.Update(ctx, 1, "New Name", "new-image")
		require.NoError(t, err)
		assert.Equal(t, "New Name", got.Name)
		assert.Equal(t, "new-image", got.Image)
	})

	t.Run("empty fields leave existing values unchanged", func(t *testing.T) {
		repo := seedRepo(t)

		got, err := repo.Update(ctx, 1, "", "")
		require.NoError(t, err)
		assert.Equal(t, "Adam Smith", got.Name)
		assert.Equal(t, "fake1", got.Image)
	})

	t.Run("unknown id is not found", func(t *testing.T) {
		repo := seedRepo(t)

		_, err := repo.Update(ctx, 999, "New Name", "")
		assert.ErrorIs(t, err, domain.ErrFriendNotFound)
	})
}

func TestRepository_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("removes an existing friend", func(t *testing.T) {
		repo := seedRepo(t)

		require.NoError(t, repo.Delete(ctx, 2))

		_, err := repo.Get(ctx, 2)
		assert.ErrorIs(t, err, domain.ErrFriendNotFound)

		count, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("unknown id is not found", func(t *testing.T) {
		repo := seedRepo(t)

		err := repo.Delete(ctx, 999)
		assert.ErrorIs(t, err, domain.ErrFriendNotFound)
	})
}

func TestRepository_All(t *testing.T) {
	ctx := context.Background()
	repo := seedRepo(t)

	all, err := repo.All(ctx, "")
	require.NoError(t, err)
	assert.Len(t, all, 3)

	filtered, err := repo.All(ctx, "adam")
	require.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "Adam Smith", filtered[0].Name)
}

func TestRepository_Count(t *testing.T) {
	ctx := context.Background()
	repo := seedRepo(t)

	count, err := repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

// seedRepo builds a repository preloaded with three friends: (1, Adam
// Smith, fake1), (2, Nolan Andrew, fake23), (3, Russel Evans, "").
func seedRepo(t *testing.T) domain.FriendRepository {
	t.Helper()

	ctx := context.Background()
	repo := NewRepository()

	for _, f := range []struct{ name, image string }{
		{"Adam Smith", "fake1"},
		{"Nolan Andrew", "fake23"},
		{"Russel Evans", ""},
	} {
		if _, err := repo.Create(ctx, f.name, f.image); err != nil {
			t.Fatalf("seed create: %v", err)
		}
	}

	return repo
}
