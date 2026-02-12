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

func setupDocumentTest(t *testing.T) (repository.DocumentRepository, *model.User, *model.Chat) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutil.CleanTables(t, testPool)

	user := createTestUser(t, "+62300", "DocUser")
	chatRepo := repository.NewChatRepository(testPool)
	chat, err := chatRepo.Create(context.Background(), model.CreateChatInput{
		Type:      model.ChatTypeGroup,
		Name:      "Doc Chat",
		CreatedBy: user.ID,
	})
	require.NoError(t, err)

	return repository.NewDocumentRepository(testPool), user, chat
}

func TestDocumentRepository_Create(t *testing.T) {
	repo, user, chat := setupDocumentTest(t)
	ctx := context.Background()

	t.Run("in chat", func(t *testing.T) {
		doc, err := repo.Create(ctx, model.CreateDocumentInput{
			Title:   "Chat Document",
			Icon:    "D",
			OwnerID: user.ID,
			ChatID:  &chat.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, "Chat Document", doc.Title)
		assert.Equal(t, &chat.ID, doc.ChatID)
	})

	t.Run("standalone", func(t *testing.T) {
		testutil.CleanTables(t, testPool)
		u := createTestUser(t, "+62301", "DocUser2")
		doc, err := repo.Create(ctx, model.CreateDocumentInput{
			Title:        "Standalone Doc",
			Icon:         "S",
			OwnerID:      u.ID,
			IsStandalone: true,
		})
		require.NoError(t, err)
		assert.True(t, doc.IsStandalone)
		assert.Nil(t, doc.ChatID)
	})
}

func TestDocumentRepository_Collaborators(t *testing.T) {
	repo, user, chat := setupDocumentTest(t)
	ctx := context.Background()

	doc, err := repo.Create(ctx, model.CreateDocumentInput{
		Title: "Collab Doc", Icon: "C", OwnerID: user.ID, ChatID: &chat.ID,
	})
	require.NoError(t, err)

	collab := createTestUser(t, "+62302", "Collaborator")
	err = repo.AddCollaborator(ctx, doc.ID, collab.ID, model.CollaboratorRoleEditor)
	require.NoError(t, err)

	err = repo.RemoveCollaborator(ctx, doc.ID, collab.ID)
	require.NoError(t, err)
}

func TestDocumentRepository_Signers(t *testing.T) {
	repo, user, chat := setupDocumentTest(t)
	ctx := context.Background()

	doc, err := repo.Create(ctx, model.CreateDocumentInput{
		Title: "Sign Doc", Icon: "S", OwnerID: user.ID, ChatID: &chat.ID,
	})
	require.NoError(t, err)

	signer := createTestUser(t, "+62303", "Signer")
	err = repo.AddSigner(ctx, doc.ID, signer.ID)
	require.NoError(t, err)

	err = repo.RecordSignature(ctx, doc.ID, signer.ID, "Signed Name")
	require.NoError(t, err)
}

func TestDocumentRepository_Lock(t *testing.T) {
	repo, user, chat := setupDocumentTest(t)
	ctx := context.Background()

	doc, err := repo.Create(ctx, model.CreateDocumentInput{
		Title: "Lock Doc", Icon: "L", OwnerID: user.ID, ChatID: &chat.ID,
	})
	require.NoError(t, err)

	err = repo.Lock(ctx, doc.ID, model.LockedByManual)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, doc.ID)
	require.NoError(t, err)
	assert.True(t, found.Locked)
}

func TestDocumentRepository_Tags(t *testing.T) {
	repo, user, chat := setupDocumentTest(t)
	ctx := context.Background()

	doc, err := repo.Create(ctx, model.CreateDocumentInput{
		Title: "Tag Doc", Icon: "T", OwnerID: user.ID, ChatID: &chat.ID,
	})
	require.NoError(t, err)

	err = repo.AddTag(ctx, doc.ID, "important")
	require.NoError(t, err)

	docs, err := repo.ListByTag(ctx, "important")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(docs), 1)

	err = repo.RemoveTag(ctx, doc.ID, "important")
	require.NoError(t, err)
}

func TestDocumentRepository_Update(t *testing.T) {
	repo, user, chat := setupDocumentTest(t)
	ctx := context.Background()

	doc, err := repo.Create(ctx, model.CreateDocumentInput{
		Title: "Old Title", Icon: "U", OwnerID: user.ID, ChatID: &chat.ID,
	})
	require.NoError(t, err)

	newTitle := "New Title"
	updated, err := repo.Update(ctx, doc.ID, model.UpdateDocumentInput{
		Title: &newTitle,
	})
	require.NoError(t, err)
	assert.Equal(t, "New Title", updated.Title)
}

func TestDocumentRepository_CascadeDelete(t *testing.T) {
	repo, user, chat := setupDocumentTest(t)
	blockRepo := repository.NewBlockRepository(testPool)
	ctx := context.Background()

	doc, err := repo.Create(ctx, model.CreateDocumentInput{
		Title: "Delete Doc", Icon: "X", OwnerID: user.ID, ChatID: &chat.ID,
	})
	require.NoError(t, err)

	_, err = blockRepo.Create(ctx, model.CreateBlockInput{
		DocumentID: doc.ID,
		Type:       model.BlockTypeParagraph,
		Content:    "test block",
		SortOrder:  0,
	})
	require.NoError(t, err)

	err = repo.Delete(ctx, doc.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, doc.ID)
	assert.True(t, apperror.IsNotFound(err))
}

func TestDocumentRepository_History(t *testing.T) {
	docRepo, user, chat := setupDocumentTest(t)
	histRepo := repository.NewDocumentHistoryRepository(testPool)
	ctx := context.Background()

	doc, err := docRepo.Create(ctx, model.CreateDocumentInput{
		Title: "History Doc", Icon: "H", OwnerID: user.ID, ChatID: &chat.ID,
	})
	require.NoError(t, err)

	err = histRepo.Create(ctx, doc.ID, user.ID, "created", "{\"note\":\"initial version\"}")
	require.NoError(t, err)

	history, err := histRepo.ListByDocument(ctx, doc.ID)
	require.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, "created", history[0].Action)
}

func TestDocumentRepository_ListByOwner(t *testing.T) {
	repo, user, _ := setupDocumentTest(t)
	ctx := context.Background()

	_, err := repo.Create(ctx, model.CreateDocumentInput{
		Title: "My Doc", Icon: "M", OwnerID: user.ID, IsStandalone: true,
	})
	require.NoError(t, err)

	docs, err := repo.ListByOwner(ctx, user.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(docs), 1)
}

func TestDocumentRepository_FindByID_NotFound(t *testing.T) {
	repo, _, _ := setupDocumentTest(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, uuid.New())
	assert.True(t, apperror.IsNotFound(err))
}
