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

func setupEntityTest(t *testing.T) (repository.EntityRepository, *model.User) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutil.CleanTables(t, testPool)

	user := createTestUser(t, "+62500", "EntityUser")
	return repository.NewEntityRepository(testPool), user
}

func TestEntityRepository_Create(t *testing.T) {
	repo, user := setupEntityTest(t)
	ctx := context.Background()

	t.Run("regular entity", func(t *testing.T) {
		entity, err := repo.Create(ctx, model.CreateEntityInput{
			Name:    "Company ABC",
			Type:    "company",
			OwnerID: user.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, "Company ABC", entity.Name)
	})

	t.Run("contact entity", func(t *testing.T) {
		contact := createTestUser(t, "+62501", "Contact")
		entity, err := repo.Create(ctx, model.CreateEntityInput{
			Name:          "My Contact",
			Type:          "contact",
			OwnerID:       user.ID,
			ContactUserID: &contact.ID,
		})
		require.NoError(t, err)
		assert.NotNil(t, entity.ContactUserID)
	})
}

func TestEntityRepository_LinkUnlinkDocument(t *testing.T) {
	repo, user := setupEntityTest(t)
	ctx := context.Background()

	entity, err := repo.Create(ctx, model.CreateEntityInput{
		Name: "Link Entity", Type: "company", OwnerID: user.ID,
	})
	require.NoError(t, err)

	docRepo := repository.NewDocumentRepository(testPool)
	doc, err := docRepo.Create(ctx, model.CreateDocumentInput{
		Title: "Link Doc", Icon: "L", OwnerID: user.ID, IsStandalone: true,
	})
	require.NoError(t, err)

	err = repo.LinkToDocument(ctx, doc.ID, entity.ID)
	require.NoError(t, err)

	entities, err := repo.ListByDocument(ctx, doc.ID)
	require.NoError(t, err)
	assert.Len(t, entities, 1)

	docs, err := repo.ListDocumentsByEntity(ctx, entity.ID)
	require.NoError(t, err)
	assert.Len(t, docs, 1)

	err = repo.UnlinkFromDocument(ctx, doc.ID, entity.ID)
	require.NoError(t, err)

	entities, err = repo.ListByDocument(ctx, doc.ID)
	require.NoError(t, err)
	assert.Len(t, entities, 0)
}

func TestEntityRepository_Search(t *testing.T) {
	repo, user := setupEntityTest(t)
	ctx := context.Background()

	_, err := repo.Create(ctx, model.CreateEntityInput{
		Name: "Acme Corporation", Type: "company", OwnerID: user.ID,
	})
	require.NoError(t, err)

	_, err = repo.Create(ctx, model.CreateEntityInput{
		Name: "Beta Inc", Type: "company", OwnerID: user.ID,
	})
	require.NoError(t, err)

	results, err := repo.Search(ctx, user.ID, "acme")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Acme Corporation", results[0].Name)
}

func TestEntityRepository_Delete(t *testing.T) {
	repo, user := setupEntityTest(t)
	ctx := context.Background()

	entity, err := repo.Create(ctx, model.CreateEntityInput{
		Name: "Delete Me", Type: "other", OwnerID: user.ID,
	})
	require.NoError(t, err)

	err = repo.Delete(ctx, entity.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, entity.ID)
	assert.True(t, apperror.IsNotFound(err))

	err = repo.Delete(ctx, uuid.New())
	assert.True(t, apperror.IsNotFound(err))
}

func TestEntityRepository_ListByOwner(t *testing.T) {
	repo, user := setupEntityTest(t)
	ctx := context.Background()

	_, err := repo.Create(ctx, model.CreateEntityInput{
		Name: "Entity 1", Type: "person", OwnerID: user.ID,
	})
	require.NoError(t, err)

	entities, err := repo.ListByOwner(ctx, user.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(entities), 1)
}
