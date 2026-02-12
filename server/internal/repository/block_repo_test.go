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
)

func setupBlockTest(t *testing.T) (repository.BlockRepository, *model.Document) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutil.CleanTables(t, testPool)

	user := createTestUser(t, "+62400", "BlockUser")
	docRepo := repository.NewDocumentRepository(testPool)
	doc, err := docRepo.Create(context.Background(), model.CreateDocumentInput{
		Title: "Block Doc", Icon: "B", OwnerID: user.ID, IsStandalone: true,
	})
	require.NoError(t, err)

	return repository.NewBlockRepository(testPool), doc
}

func TestBlockRepository_CRUD(t *testing.T) {
	repo, doc := setupBlockTest(t)
	ctx := context.Background()

	block, err := repo.Create(ctx, model.CreateBlockInput{
		DocumentID: doc.ID,
		Type:       model.BlockTypeParagraph,
		Content:    "Hello world",
		SortOrder:  0,
	})
	require.NoError(t, err)
	assert.Equal(t, model.BlockTypeParagraph, block.Type)

	blocks, err := repo.ListByDocument(ctx, doc.ID)
	require.NoError(t, err)
	assert.Len(t, blocks, 1)

	newContent := "Updated content"
	updated, err := repo.Update(ctx, block.ID, model.UpdateBlockInput{
		Content: &newContent,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated content", updated.Content)

	err = repo.Delete(ctx, block.ID)
	require.NoError(t, err)

	blocks, err = repo.ListByDocument(ctx, doc.ID)
	require.NoError(t, err)
	assert.Len(t, blocks, 0)
}

func TestBlockRepository_Reorder(t *testing.T) {
	repo, doc := setupBlockTest(t)
	ctx := context.Background()

	b1, err := repo.Create(ctx, model.CreateBlockInput{
		DocumentID: doc.ID, Type: model.BlockTypeParagraph,
		Content: "First", SortOrder: 0,
	})
	require.NoError(t, err)

	b2, err := repo.Create(ctx, model.CreateBlockInput{
		DocumentID: doc.ID, Type: model.BlockTypeParagraph,
		Content: "Second", SortOrder: 1,
	})
	require.NoError(t, err)

	b3, err := repo.Create(ctx, model.CreateBlockInput{
		DocumentID: doc.ID, Type: model.BlockTypeParagraph,
		Content: "Third", SortOrder: 2,
	})
	require.NoError(t, err)

	err = repo.Reorder(ctx, doc.ID, []uuid.UUID{b3.ID, b1.ID, b2.ID})
	require.NoError(t, err)

	blocks, err := repo.ListByDocument(ctx, doc.ID)
	require.NoError(t, err)
	assert.Len(t, blocks, 3)
	assert.Equal(t, b3.ID, blocks[0].ID)
	assert.Equal(t, b1.ID, blocks[1].ID)
	assert.Equal(t, b2.ID, blocks[2].ID)
}
