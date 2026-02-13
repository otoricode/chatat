package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/model"
)

func TestMessageService_SendMessage(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewMessageService(msgRepo, msgStatRepo, chatRepo, hub)

	// Create a chat with two members
	userA := uuid.New()
	userB := uuid.New()
	chat := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal, CreatedBy: userA}
	chatRepo.chats[chat.ID] = chat
	_ = chatRepo.AddMember(context.Background(), chat.ID, userA, model.MemberRoleAdmin)
	_ = chatRepo.AddMember(context.Background(), chat.ID, userB, model.MemberRoleMember)

	t.Run("send text message", func(t *testing.T) {
		msg, err := svc.SendMessage(context.Background(), SendMessageInput{
			ChatID:   chat.ID,
			SenderID: userA,
			Content:  "Hello",
			Type:     model.MessageTypeText,
		})
		require.NoError(t, err)
		assert.Equal(t, "Hello", msg.Content)
		assert.Equal(t, chat.ID, msg.ChatID)
		assert.Equal(t, userA, msg.SenderID)
	})

	t.Run("empty content rejected", func(t *testing.T) {
		_, err := svc.SendMessage(context.Background(), SendMessageInput{
			ChatID:   chat.ID,
			SenderID: userA,
			Content:  "",
			Type:     model.MessageTypeText,
		})
		require.Error(t, err)
	})

	t.Run("not a member rejected", func(t *testing.T) {
		_, err := svc.SendMessage(context.Background(), SendMessageInput{
			ChatID:   chat.ID,
			SenderID: uuid.New(),
			Content:  "Hello",
			Type:     model.MessageTypeText,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a member")
	})
}

func TestMessageService_GetMessages(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewMessageService(msgRepo, msgStatRepo, chatRepo, hub)

	chatID := uuid.New()
	userA := uuid.New()

	// Add some messages
	for i := 0; i < 5; i++ {
		_, _ = msgRepo.Create(context.Background(), model.CreateMessageInput{
			ChatID:   chatID,
			SenderID: userA,
			Content:  "msg",
			Type:     model.MessageTypeText,
		})
	}

	t.Run("get first page", func(t *testing.T) {
		page, err := svc.GetMessages(context.Background(), chatID, "", 3)
		require.NoError(t, err)
		assert.Len(t, page.Messages, 3)
		assert.True(t, page.HasMore)
		assert.NotEmpty(t, page.Cursor)
	})

	t.Run("get all messages", func(t *testing.T) {
		page, err := svc.GetMessages(context.Background(), chatID, "", 50)
		require.NoError(t, err)
		assert.Len(t, page.Messages, 5)
		assert.False(t, page.HasMore)
	})
}

func TestMessageService_ReplyMessage(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewMessageService(msgRepo, msgStatRepo, chatRepo, hub)

	userA := uuid.New()
	userB := uuid.New()
	chat := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal, CreatedBy: userA}
	chatRepo.chats[chat.ID] = chat
	_ = chatRepo.AddMember(context.Background(), chat.ID, userA, model.MemberRoleAdmin)
	_ = chatRepo.AddMember(context.Background(), chat.ID, userB, model.MemberRoleMember)

	// Send original
	original, err := svc.SendMessage(context.Background(), SendMessageInput{
		ChatID:   chat.ID,
		SenderID: userA,
		Content:  "Original",
		Type:     model.MessageTypeText,
	})
	require.NoError(t, err)

	t.Run("reply to message", func(t *testing.T) {
		reply, err := svc.SendMessage(context.Background(), SendMessageInput{
			ChatID:    chat.ID,
			SenderID:  userB,
			Content:   "Reply",
			Type:      model.MessageTypeText,
			ReplyToID: &original.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, &original.ID, reply.ReplyToID)
	})

	t.Run("reply to non-existent message", func(t *testing.T) {
		badID := uuid.New()
		_, err := svc.SendMessage(context.Background(), SendMessageInput{
			ChatID:    chat.ID,
			SenderID:  userA,
			Content:   "Reply",
			Type:      model.MessageTypeText,
			ReplyToID: &badID,
		})
		require.Error(t, err)
	})
}

func TestMessageService_ForwardMessage(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewMessageService(msgRepo, msgStatRepo, chatRepo, hub)

	userA := uuid.New()
	userB := uuid.New()
	userC := uuid.New()

	// Chat 1: A & B
	chat1 := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal, CreatedBy: userA}
	chatRepo.chats[chat1.ID] = chat1
	_ = chatRepo.AddMember(context.Background(), chat1.ID, userA, model.MemberRoleAdmin)
	_ = chatRepo.AddMember(context.Background(), chat1.ID, userB, model.MemberRoleMember)

	// Chat 2: A & C
	chat2 := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal, CreatedBy: userA}
	chatRepo.chats[chat2.ID] = chat2
	_ = chatRepo.AddMember(context.Background(), chat2.ID, userA, model.MemberRoleAdmin)
	_ = chatRepo.AddMember(context.Background(), chat2.ID, userC, model.MemberRoleMember)

	// Send message in chat1
	original, err := svc.SendMessage(context.Background(), SendMessageInput{
		ChatID:   chat1.ID,
		SenderID: userA,
		Content:  "Forward me",
		Type:     model.MessageTypeText,
	})
	require.NoError(t, err)

	t.Run("forward to another chat", func(t *testing.T) {
		fwd, err := svc.ForwardMessage(context.Background(), original.ID, userA, chat2.ID)
		require.NoError(t, err)
		assert.Equal(t, "Forward me", fwd.Content)
		assert.Equal(t, chat2.ID, fwd.ChatID)
		assert.NotNil(t, fwd.Metadata)
	})

	t.Run("forward not a member of target", func(t *testing.T) {
		_, err := svc.ForwardMessage(context.Background(), original.ID, userB, chat2.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a member")
	})
}

func TestMessageService_DeleteMessage(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewMessageService(msgRepo, msgStatRepo, chatRepo, hub)

	userA := uuid.New()
	userB := uuid.New()
	chat := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal, CreatedBy: userA}
	chatRepo.chats[chat.ID] = chat
	_ = chatRepo.AddMember(context.Background(), chat.ID, userA, model.MemberRoleAdmin)
	_ = chatRepo.AddMember(context.Background(), chat.ID, userB, model.MemberRoleMember)

	t.Run("delete for self", func(t *testing.T) {
		msg, _ := svc.SendMessage(context.Background(), SendMessageInput{
			ChatID: chat.ID, SenderID: userA, Content: "Hi", Type: model.MessageTypeText,
		})

		err := svc.DeleteMessage(context.Background(), msg.ID, userA, false)
		require.NoError(t, err)
	})

	t.Run("delete for all by sender", func(t *testing.T) {
		msg, _ := svc.SendMessage(context.Background(), SendMessageInput{
			ChatID: chat.ID, SenderID: userA, Content: "Hi", Type: model.MessageTypeText,
		})

		err := svc.DeleteMessage(context.Background(), msg.ID, userA, true)
		require.NoError(t, err)
		assert.True(t, msgRepo.messages[msg.ID].DeletedForAll)
	})

	t.Run("delete for all by non-sender rejected", func(t *testing.T) {
		msg, _ := svc.SendMessage(context.Background(), SendMessageInput{
			ChatID: chat.ID, SenderID: userA, Content: "Hi", Type: model.MessageTypeText,
		})

		err := svc.DeleteMessage(context.Background(), msg.ID, userB, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only the sender")
	})

	t.Run("delete for all expired", func(t *testing.T) {
		msg, _ := svc.SendMessage(context.Background(), SendMessageInput{
			ChatID: chat.ID, SenderID: userA, Content: "Hi", Type: model.MessageTypeText,
		})
		// Set created_at to 2 hours ago
		msgRepo.messages[msg.ID].CreatedAt = time.Now().Add(-2 * time.Hour)

		err := svc.DeleteMessage(context.Background(), msg.ID, userA, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "within 1 hour")
	})
}

func TestMessageService_MarkChatAsRead(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewMessageService(msgRepo, msgStatRepo, chatRepo, hub)

	chatID := uuid.New()
	userA := uuid.New()

	err := svc.MarkChatAsRead(context.Background(), chatID, userA)
	require.NoError(t, err)
}
