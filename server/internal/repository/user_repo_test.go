package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/internal/testutil"
	"github.com/otoritech/chatat/pkg/apperror"
)

func setupUserRepo(t *testing.T) repository.UserRepository {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutil.CleanTables(t, testPool)
	return repository.NewUserRepository(testPool)
}

func TestUserRepository_Create(t *testing.T) {
	repo := setupUserRepo(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user, err := repo.Create(ctx, model.CreateUserInput{
			Phone: "+6281234567890",
			Name:  "Test User",
		})
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, user.ID)
		assert.Equal(t, "+6281234567890", user.Phone)
		assert.Equal(t, "Test User", user.Name)
		assert.NotEmpty(t, user.Avatar)
	})

	t.Run("duplicate phone", func(t *testing.T) {
		testutil.CleanTables(t, testPool)
		_, err := repo.Create(ctx, model.CreateUserInput{
			Phone: "+6281234567891",
			Name:  "User 1",
		})
		require.NoError(t, err)

		_, err = repo.Create(ctx, model.CreateUserInput{
			Phone: "+6281234567891",
			Name:  "User 2",
		})
		assert.Error(t, err)
		assert.True(t, apperror.IsConflict(err))
	})
}

func TestUserRepository_FindByID(t *testing.T) {
	repo := setupUserRepo(t)
	ctx := context.Background()

	t.Run("exists", func(t *testing.T) {
		created, err := repo.Create(ctx, model.CreateUserInput{
			Phone: "+6281234567892",
			Name:  "Find Me",
		})
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.Phone, found.Phone)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.FindByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.True(t, apperror.IsNotFound(err))
	})
}

func TestUserRepository_FindByPhone(t *testing.T) {
	repo := setupUserRepo(t)
	ctx := context.Background()

	_, err := repo.Create(ctx, model.CreateUserInput{
		Phone: "+6281234567893",
		Name:  "Phone User",
	})
	require.NoError(t, err)

	found, err := repo.FindByPhone(ctx, "+6281234567893")
	require.NoError(t, err)
	assert.Equal(t, "Phone User", found.Name)

	_, err = repo.FindByPhone(ctx, "+6200000000000")
	assert.Error(t, err)
	assert.True(t, apperror.IsNotFound(err))
}

func TestUserRepository_FindByPhones(t *testing.T) {
	repo := setupUserRepo(t)
	ctx := context.Background()

	_, err := repo.Create(ctx, model.CreateUserInput{Phone: "+62111", Name: "A"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, model.CreateUserInput{Phone: "+62222", Name: "B"})
	require.NoError(t, err)

	t.Run("batch find", func(t *testing.T) {
		users, err := repo.FindByPhones(ctx, []string{"+62111", "+62222", "+62999"})
		require.NoError(t, err)
		assert.Len(t, users, 2)
	})

	t.Run("empty input", func(t *testing.T) {
		users, err := repo.FindByPhones(ctx, []string{})
		require.NoError(t, err)
		assert.Len(t, users, 0)
	})
}

func TestUserRepository_Update(t *testing.T) {
	repo := setupUserRepo(t)
	ctx := context.Background()

	user, err := repo.Create(ctx, model.CreateUserInput{
		Phone: "+62333",
		Name:  "Original",
	})
	require.NoError(t, err)

	newName := "Updated"
	updated, err := repo.Update(ctx, user.ID, model.UpdateUserInput{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)

	_, err = repo.Update(ctx, uuid.New(), model.UpdateUserInput{Name: &newName})
	assert.True(t, apperror.IsNotFound(err))
}

func TestUserRepository_Delete(t *testing.T) {
	repo := setupUserRepo(t)
	ctx := context.Background()

	user, err := repo.Create(ctx, model.CreateUserInput{
		Phone: "+62444",
		Name:  "Delete Me",
	})
	require.NoError(t, err)

	err = repo.Delete(ctx, user.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, user.ID)
	assert.True(t, apperror.IsNotFound(err))

	err = repo.Delete(ctx, uuid.New())
	assert.True(t, apperror.IsNotFound(err))
}

func TestUserRepository_UpdateLastSeen(t *testing.T) {
	repo := setupUserRepo(t)
	ctx := context.Background()

	user, err := repo.Create(ctx, model.CreateUserInput{
		Phone: "+62555",
		Name:  "Last Seen",
	})
	require.NoError(t, err)

	err = repo.UpdateLastSeen(ctx, user.ID)
	require.NoError(t, err)

	err = repo.UpdateLastSeen(ctx, uuid.New())
	assert.True(t, apperror.IsNotFound(err))
}
