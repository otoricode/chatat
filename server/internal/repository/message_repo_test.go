package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/internal/testutil"
	"github.com/otoritech/chatat/pkg/apperror"
)

func setupMessageTest(t *testing.T) (repository.ChatRepository, repository.MessageRepository, *model.User, *model.Chat) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutil.CleanTables(t, testPool)

	chatRepo := repository.NewChatRepository(testPool)
	msgRepo := repository.NewMessageRepository(testPool)
	user := createTestUser(t, "+62100", "MsgUser")

	chat, err := chatRepo.Create(context.Background(), model.CreateChatInput{
		Type:      model.ChatTypeGroup,
		Name:      "Msg Chat",
		CreatedBy: user.ID,
	})
	require.NoError(t, err)
	err = chatRepo.AddMember(context.Background(), chat.ID, user.ID, model.MemberRoleMember)
	require.NoError(t, err)

	return chatRepo, msgRepo, user, chat
}

func TestMessageRepository_Create(t *testing.T) {
	_, msgRepo, user, chat := setupMessageTest(t)
	ctx := context.Background()

	msg, err := msgRepo.Create(ctx, model.CreateMessageInput{
		ChatID:   chat.ID,
		SenderID: user.ID,
		Content:  "hello world",
		Type:     model.MessageTypeText,
	})
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, msg.ID)
	assert.Equal(t, "hello world", msg.Content)
}

func TestMessageRepository_Reply(t *testing.T) {
	_, msgRepo, user, chat := setupMessageTest(t)
	ctx := context.Background()

	original, err := msgRepo.Create(ctx, model.CreateMessageInput{
		ChatID:   chat.ID,
		SenderID: user.ID,
		Content:  "original",
		Type:     model.MessageTypeText,
	})
	require.NoError(t, err)

	reply, err := msgRepo.Create(ctx, model.CreateMessageInput{
		ChatID:    chat.ID,
		SenderID:  user.ID,
		Content:   "reply",
		Type:      model.MessageTypeText,
		ReplyToID: &original.ID,
	})
	require.NoError(t, err)
	assert.NotNil(t, reply.ReplyToID)
	assert.Equal(t, original.ID, *reply.ReplyToID)
}

func TestMessageRepository_CursorPagination(t *testing.T) {
	_, msgRepo, user, chat := setupMessageTest(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_, err := msgRepo.Create(ctx, model.CreateMessageInput{
			ChatID:   chat.ID,
			SenderID: user.ID,
			Content:  "msg",
			Type:     model.MessageTypeText,
		})
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	msgs, err := msgRepo.ListByChat(ctx, chat.ID, nil, 3)
	require.NoError(t, err)
	assert.Len(t, msgs, 3)

	cursor := msgs[len(msgs)-1].CreatedAt
	msgs2, err := msgRepo.ListByChat(ctx, chat.ID, &cursor, 3)
	require.NoError(t, err)
	assert.Len(t, msgs2, 2)
}

func TestMessageRepository_SoftDelete(t *testing.T) {
	_, msgRepo, user, chat := setupMessageTest(t)
	ctx := context.Background()

	msg, err := msgRepo.Create(ctx, model.CreateMessageInput{
		ChatID:   chat.ID,
		SenderID: user.ID,
		Content:  "delete me",
		Type:     model.MessageTypeText,
	})
	require.NoError(t, err)

	err = msgRepo.MarkAsDeleted(ctx, msg.ID, true)
	require.NoError(t, err)

	found, err := msgRepo.FindByID(ctx, msg.ID)
	require.NoError(t, err)
	assert.True(t, found.IsDeleted)
	assert.True(t, found.DeletedForAll)
}

func TestMessageRepository_Search(t *testing.T) {
	_, msgRepo, user, chat := setupMessageTest(t)
	ctx := context.Background()

	_, err := msgRepo.Create(ctx, model.CreateMessageInput{
		ChatID:   chat.ID,
		SenderID: user.ID,
		Content:  "important meeting tomorrow",
		Type:     model.MessageTypeText,
	})
	require.NoError(t, err)

	_, err = msgRepo.Create(ctx, model.CreateMessageInput{
		ChatID:   chat.ID,
		SenderID: user.ID,
		Content:  "hello there",
		Type:     model.MessageTypeText,
	})
	require.NoError(t, err)

	results, err := msgRepo.Search(ctx, chat.ID, "meeting")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Content, "meeting")
}

func TestMessageRepository_FindByID_NotFound(t *testing.T) {
	_, msgRepo, _, _ := setupMessageTest(t)
	ctx := context.Background()

	_, err := msgRepo.FindByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.True(t, apperror.IsNotFound(err))
}
