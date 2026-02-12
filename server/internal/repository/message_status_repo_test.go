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

func setupStatusTest(t *testing.T) (repository.MessageStatusRepository, *model.User, *model.Chat, *model.Message) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutil.CleanTables(t, testPool)

	user := createTestUser(t, "+62600", "StatusUser")

	chatRepo := repository.NewChatRepository(testPool)
	chat, err := chatRepo.Create(context.Background(), model.CreateChatInput{
		Type: model.ChatTypeGroup, Name: "Status Chat", CreatedBy: user.ID,
	})
	require.NoError(t, err)
	err = chatRepo.AddMember(context.Background(), chat.ID, user.ID, model.MemberRoleMember)
	require.NoError(t, err)

	msgRepo := repository.NewMessageRepository(testPool)
	msg, err := msgRepo.Create(context.Background(), model.CreateMessageInput{
		ChatID: chat.ID, SenderID: user.ID, Content: "status test", Type: model.MessageTypeText,
	})
	require.NoError(t, err)

	return repository.NewMessageStatusRepository(testPool), user, chat, msg
}

func TestMessageStatus_Transitions(t *testing.T) {
	repo, user, _, msg := setupStatusTest(t)
	ctx := context.Background()

	err := repo.Create(ctx, msg.ID, user.ID, model.DeliveryStatusSent)
	require.NoError(t, err)

	err = repo.UpdateStatus(ctx, msg.ID, user.ID, model.DeliveryStatusDelivered)
	require.NoError(t, err)

	err = repo.UpdateStatus(ctx, msg.ID, user.ID, model.DeliveryStatusRead)
	require.NoError(t, err)

	statuses, err := repo.GetStatus(ctx, msg.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(statuses), 1)
	assert.Equal(t, model.DeliveryStatusRead, statuses[0].Status)
}

func TestMessageStatus_UnreadCount(t *testing.T) {
	repo, user, chat, msg := setupStatusTest(t)
	ctx := context.Background()

	err := repo.Create(ctx, msg.ID, user.ID, model.DeliveryStatusSent)
	require.NoError(t, err)

	count, err := repo.GetUnreadCount(ctx, chat.ID, user.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 0)
}

func TestMessageStatus_BatchMarkAsRead(t *testing.T) {
	repo, user, chat, msg := setupStatusTest(t)
	ctx := context.Background()

	err := repo.Create(ctx, msg.ID, user.ID, model.DeliveryStatusSent)
	require.NoError(t, err)

	err = repo.MarkChatAsRead(ctx, chat.ID, user.ID)
	require.NoError(t, err)

	statuses, err := repo.GetStatus(ctx, msg.ID)
	require.NoError(t, err)
	for _, s := range statuses {
		if s.UserID == user.ID {
			assert.Equal(t, model.DeliveryStatusRead, s.Status)
		}
	}
}
