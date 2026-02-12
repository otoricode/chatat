package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/internal/testutil"
)

func setupTopicTest(t *testing.T) (repository.TopicRepository, *model.User, *model.Chat) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutil.CleanTables(t, testPool)

	user := createTestUser(t, "+62200", "TopicUser")
	chatRepo := repository.NewChatRepository(testPool)
	chat, err := chatRepo.Create(context.Background(), model.CreateChatInput{
		Type:      model.ChatTypeGroup,
		Name:      "Topic Parent",
		CreatedBy: user.ID,
	})
	require.NoError(t, err)

	return repository.NewTopicRepository(testPool), user, chat
}

func TestTopicRepository_Create(t *testing.T) {
	repo, user, chat := setupTopicTest(t)
	ctx := context.Background()

	topic, err := repo.Create(ctx, model.CreateTopicInput{
		Name:       "Design Discussion",
		Icon:       "D",
		ParentType: model.ChatTypeGroup,
		ParentID:   chat.ID,
		CreatedBy:  user.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, "Design Discussion", topic.Name)
	assert.Equal(t, chat.ID, topic.ParentID)
}

func TestTopicRepository_ListByParent(t *testing.T) {
	repo, user, chat := setupTopicTest(t)
	ctx := context.Background()

	_, err := repo.Create(ctx, model.CreateTopicInput{
		Name: "Topic1", ParentType: model.ChatTypeGroup,
		ParentID: chat.ID, CreatedBy: user.ID,
	})
	require.NoError(t, err)

	_, err = repo.Create(ctx, model.CreateTopicInput{
		Name: "Topic2", ParentType: model.ChatTypeGroup,
		ParentID: chat.ID, CreatedBy: user.ID,
	})
	require.NoError(t, err)

	topics, err := repo.ListByParent(ctx, chat.ID)
	require.NoError(t, err)
	assert.Len(t, topics, 2)
}

func TestTopicRepository_Members(t *testing.T) {
	repo, user, chat := setupTopicTest(t)
	ctx := context.Background()

	topic, err := repo.Create(ctx, model.CreateTopicInput{
		Name: "Members Topic", ParentType: model.ChatTypeGroup,
		ParentID: chat.ID, CreatedBy: user.ID,
	})
	require.NoError(t, err)

	err = repo.AddMember(ctx, topic.ID, user.ID, model.MemberRoleAdmin)
	require.NoError(t, err)

	members, err := repo.GetMembers(ctx, topic.ID)
	require.NoError(t, err)
	assert.Len(t, members, 1)

	err = repo.RemoveMember(ctx, topic.ID, user.ID)
	require.NoError(t, err)

	members, err = repo.GetMembers(ctx, topic.ID)
	require.NoError(t, err)
	assert.Len(t, members, 0)
}

func TestTopicRepository_CascadeDelete(t *testing.T) {
	repo, user, chat := setupTopicTest(t)
	msgRepo := repository.NewTopicMessageRepository(testPool)
	ctx := context.Background()

	topic, err := repo.Create(ctx, model.CreateTopicInput{
		Name: "Cascade Topic", ParentType: model.ChatTypeGroup,
		ParentID: chat.ID, CreatedBy: user.ID,
	})
	require.NoError(t, err)

	_, err = msgRepo.Create(ctx, model.CreateTopicMessageInput{
		TopicID:  topic.ID,
		SenderID: user.ID,
		Content:  "topic msg",
		Type:     model.MessageTypeText,
	})
	require.NoError(t, err)

	err = repo.Delete(ctx, topic.ID)
	require.NoError(t, err)
}
