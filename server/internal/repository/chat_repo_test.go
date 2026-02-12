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

func createTestUser(t *testing.T, phone, name string) *model.User {
	t.Helper()
	repo := repository.NewUserRepository(testPool)
	user, err := repo.Create(context.Background(), model.CreateUserInput{
		Phone: phone,
		Name:  name,
	})
	require.NoError(t, err)
	return user
}

func setupChatRepo(t *testing.T) repository.ChatRepository {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutil.CleanTables(t, testPool)
	return repository.NewChatRepository(testPool)
}

func TestChatRepository_CreatePersonal(t *testing.T) {
	repo := setupChatRepo(t)
	ctx := context.Background()
	user := createTestUser(t, "+62001", "User1")

	chat, err := repo.Create(ctx, model.CreateChatInput{
		Type:      model.ChatTypePersonal,
		CreatedBy: user.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, model.ChatTypePersonal, chat.Type)
	assert.NotEqual(t, uuid.Nil, chat.ID)
}

func TestChatRepository_CreateGroup(t *testing.T) {
	repo := setupChatRepo(t)
	ctx := context.Background()
	user := createTestUser(t, "+62002", "User2")

	chat, err := repo.Create(ctx, model.CreateChatInput{
		Type:      model.ChatTypeGroup,
		Name:      "Test Group",
		Icon:      "G",
		CreatedBy: user.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, model.ChatTypeGroup, chat.Type)
	assert.Equal(t, "Test Group", chat.Name)
}

func TestChatRepository_AddRemoveMember(t *testing.T) {
	repo := setupChatRepo(t)
	ctx := context.Background()
	user1 := createTestUser(t, "+62003", "User3")
	user2 := createTestUser(t, "+62004", "User4")

	chat, err := repo.Create(ctx, model.CreateChatInput{
		Type:      model.ChatTypeGroup,
		Name:      "Members Group",
		CreatedBy: user1.ID,
	})
	require.NoError(t, err)

	err = repo.AddMember(ctx, chat.ID, user1.ID, model.MemberRoleAdmin)
	require.NoError(t, err)
	err = repo.AddMember(ctx, chat.ID, user2.ID, model.MemberRoleMember)
	require.NoError(t, err)

	members, err := repo.GetMembers(ctx, chat.ID)
	require.NoError(t, err)
	assert.Len(t, members, 2)

	err = repo.RemoveMember(ctx, chat.ID, user2.ID)
	require.NoError(t, err)

	members, err = repo.GetMembers(ctx, chat.ID)
	require.NoError(t, err)
	assert.Len(t, members, 1)
}

func TestChatRepository_FindPersonalChat(t *testing.T) {
	repo := setupChatRepo(t)
	ctx := context.Background()
	user1 := createTestUser(t, "+62005", "User5")
	user2 := createTestUser(t, "+62006", "User6")

	chat, err := repo.Create(ctx, model.CreateChatInput{
		Type:      model.ChatTypePersonal,
		CreatedBy: user1.ID,
	})
	require.NoError(t, err)
	err = repo.AddMember(ctx, chat.ID, user1.ID, model.MemberRoleMember)
	require.NoError(t, err)
	err = repo.AddMember(ctx, chat.ID, user2.ID, model.MemberRoleMember)
	require.NoError(t, err)

	found, err := repo.FindPersonalChat(ctx, user1.ID, user2.ID)
	require.NoError(t, err)
	assert.Equal(t, chat.ID, found.ID)
}

func TestChatRepository_ListByUser(t *testing.T) {
	repo := setupChatRepo(t)
	ctx := context.Background()
	user := createTestUser(t, "+62007", "User7")

	chat, err := repo.Create(ctx, model.CreateChatInput{
		Type:      model.ChatTypeGroup,
		Name:      "My Chat",
		CreatedBy: user.ID,
	})
	require.NoError(t, err)
	err = repo.AddMember(ctx, chat.ID, user.ID, model.MemberRoleAdmin)
	require.NoError(t, err)

	list, err := repo.ListByUser(ctx, user.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 1)
}

func TestChatRepository_Update(t *testing.T) {
	repo := setupChatRepo(t)
	ctx := context.Background()
	user := createTestUser(t, "+62008", "User8")

	chat, err := repo.Create(ctx, model.CreateChatInput{
		Type:      model.ChatTypeGroup,
		Name:      "Old Name",
		CreatedBy: user.ID,
	})
	require.NoError(t, err)

	newName := "New Name"
	updated, err := repo.Update(ctx, chat.ID, model.UpdateChatInput{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)
}

func TestChatRepository_CascadeDelete(t *testing.T) {
	chatRepo := setupChatRepo(t)
	msgRepo := repository.NewMessageRepository(testPool)
	ctx := context.Background()
	user := createTestUser(t, "+62009", "User9")

	chat, err := chatRepo.Create(ctx, model.CreateChatInput{
		Type:      model.ChatTypeGroup,
		Name:      "Delete Me",
		CreatedBy: user.ID,
	})
	require.NoError(t, err)
	err = chatRepo.AddMember(ctx, chat.ID, user.ID, model.MemberRoleMember)
	require.NoError(t, err)

	_, err = msgRepo.Create(ctx, model.CreateMessageInput{
		ChatID:   chat.ID,
		SenderID: user.ID,
		Content:  "hello",
		Type:     model.MessageTypeText,
	})
	require.NoError(t, err)

	err = chatRepo.Delete(ctx, chat.ID)
	require.NoError(t, err)

	_, err = chatRepo.FindByID(ctx, chat.ID)
	assert.True(t, apperror.IsNotFound(err))
}
