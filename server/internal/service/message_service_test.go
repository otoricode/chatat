package service

import (
	"context"
	"fmt"
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

	svc := NewMessageService(msgRepo, msgStatRepo, chatRepo, nil, hub, nil)

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

	svc := NewMessageService(msgRepo, msgStatRepo, chatRepo, nil, hub, nil)

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

	svc := NewMessageService(msgRepo, msgStatRepo, chatRepo, nil, hub, nil)

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

	svc := NewMessageService(msgRepo, msgStatRepo, chatRepo, nil, hub, nil)

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

	svc := NewMessageService(msgRepo, msgStatRepo, chatRepo, nil, hub, nil)

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

	svc := NewMessageService(msgRepo, msgStatRepo, chatRepo, nil, hub, nil)

	chatID := uuid.New()
	userA := uuid.New()

	err := svc.MarkChatAsRead(context.Background(), chatID, userA)
	require.NoError(t, err)
}

func TestMessageService_SearchMessages(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewMessageService(msgRepo, msgStatRepo, chatRepo, nil, hub, nil)

	chatID := uuid.New()

	t.Run("success", func(t *testing.T) {
		msgs, err := svc.SearchMessages(context.Background(), chatID, "hello")
		require.NoError(t, err)
		// mock repo returns nil, just check no error
		assert.Empty(t, msgs)
	})

	t.Run("empty query rejected", func(t *testing.T) {
		_, err := svc.SearchMessages(context.Background(), chatID, "")
		assert.Error(t, err)
	})
}

func TestMessageService_SendMessage_Errors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("get members error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		chatRepo.getMembersErr = fmt.Errorf("db error")
		svc := NewMessageService(newMockMessageRepo(), newMockMessageStatRepo(), chatRepo, nil, hub, nil)
		_, err := svc.SendMessage(context.Background(), SendMessageInput{
			ChatID: uuid.New(), SenderID: uuid.New(), Content: "Hi", Type: model.MessageTypeText,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get chat members")
	})

	t.Run("create message error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		msgRepo := newMockMessageRepo()
		userA := uuid.New()
		chat := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal, CreatedBy: userA}
		chatRepo.chats[chat.ID] = chat
		_ = chatRepo.AddMember(context.Background(), chat.ID, userA, model.MemberRoleAdmin)

		msgRepo.createErr = fmt.Errorf("db error")
		svc := NewMessageService(msgRepo, newMockMessageStatRepo(), chatRepo, nil, hub, nil)
		_, err := svc.SendMessage(context.Background(), SendMessageInput{
			ChatID: chat.ID, SenderID: userA, Content: "Hi", Type: model.MessageTypeText,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create message")
	})

	t.Run("reply in different chat", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		msgRepo := newMockMessageRepo()
		userA := uuid.New()

		chat1 := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal, CreatedBy: userA}
		chatRepo.chats[chat1.ID] = chat1
		_ = chatRepo.AddMember(context.Background(), chat1.ID, userA, model.MemberRoleAdmin)

		chat2 := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal, CreatedBy: userA}
		chatRepo.chats[chat2.ID] = chat2
		_ = chatRepo.AddMember(context.Background(), chat2.ID, userA, model.MemberRoleAdmin)

		svc := NewMessageService(msgRepo, newMockMessageStatRepo(), chatRepo, nil, hub, nil)
		original, err := svc.SendMessage(context.Background(), SendMessageInput{
			ChatID: chat1.ID, SenderID: userA, Content: "Original", Type: model.MessageTypeText,
		})
		require.NoError(t, err)

		_, err = svc.SendMessage(context.Background(), SendMessageInput{
			ChatID: chat2.ID, SenderID: userA, Content: "Reply", Type: model.MessageTypeText,
			ReplyToID: &original.ID,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "same chat")
	})

	t.Run("default type", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userA := uuid.New()
		chat := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal, CreatedBy: userA}
		chatRepo.chats[chat.ID] = chat
		_ = chatRepo.AddMember(context.Background(), chat.ID, userA, model.MemberRoleAdmin)

		svc := NewMessageService(newMockMessageRepo(), newMockMessageStatRepo(), chatRepo, nil, hub, nil)
		msg, err := svc.SendMessage(context.Background(), SendMessageInput{
			ChatID: chat.ID, SenderID: userA, Content: "Hi",
		})
		require.NoError(t, err)
		assert.Equal(t, model.MessageTypeText, msg.Type)
	})
}

func TestMessageService_GetMessages_Errors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("invalid cursor", func(t *testing.T) {
		svc := NewMessageService(newMockMessageRepo(), newMockMessageStatRepo(), newMockChatRepo(), nil, hub, nil)
		_, err := svc.GetMessages(context.Background(), uuid.New(), "not-a-time", 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid cursor")
	})

	t.Run("list error", func(t *testing.T) {
		msgRepo := newMockMessageRepo()
		msgRepo.listErr = fmt.Errorf("db error")
		svc := NewMessageService(msgRepo, newMockMessageStatRepo(), newMockChatRepo(), nil, hub, nil)
		_, err := svc.GetMessages(context.Background(), uuid.New(), "", 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "list messages")
	})

	t.Run("default limit", func(t *testing.T) {
		svc := NewMessageService(newMockMessageRepo(), newMockMessageStatRepo(), newMockChatRepo(), nil, hub, nil)
		page, err := svc.GetMessages(context.Background(), uuid.New(), "", 0)
		require.NoError(t, err)
		assert.Empty(t, page.Messages)
	})

	t.Run("with valid cursor", func(t *testing.T) {
		svc := NewMessageService(newMockMessageRepo(), newMockMessageStatRepo(), newMockChatRepo(), nil, hub, nil)
		cursor := time.Now().Format(time.RFC3339Nano)
		page, err := svc.GetMessages(context.Background(), uuid.New(), cursor, 10)
		require.NoError(t, err)
		assert.Empty(t, page.Messages)
	})
}

func TestMessageService_ForwardMessage_Errors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("original not found", func(t *testing.T) {
		svc := NewMessageService(newMockMessageRepo(), newMockMessageStatRepo(), newMockChatRepo(), nil, hub, nil)
		_, err := svc.ForwardMessage(context.Background(), uuid.New(), uuid.New(), uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "find original message")
	})

	t.Run("get target members error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		msgRepo := newMockMessageRepo()
		userA := uuid.New()

		chat := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal, CreatedBy: userA}
		chatRepo.chats[chat.ID] = chat
		_ = chatRepo.AddMember(context.Background(), chat.ID, userA, model.MemberRoleAdmin)

		svc := NewMessageService(msgRepo, newMockMessageStatRepo(), chatRepo, nil, hub, nil)
		msg, _ := svc.SendMessage(context.Background(), SendMessageInput{
			ChatID: chat.ID, SenderID: userA, Content: "Hi", Type: model.MessageTypeText,
		})

		chatRepo.getMembersErr = fmt.Errorf("db error")
		_, err := svc.ForwardMessage(context.Background(), msg.ID, userA, uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get target chat members")
	})
}

func TestMessageService_DeleteMessage_Errors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("message not found", func(t *testing.T) {
		svc := NewMessageService(newMockMessageRepo(), newMockMessageStatRepo(), newMockChatRepo(), nil, hub, nil)
		err := svc.DeleteMessage(context.Background(), uuid.New(), uuid.New(), false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "find message")
	})
}

func TestMessageService_SearchMessages_Error(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	msgRepo := newMockMessageRepo()
	msgRepo.searchErr = fmt.Errorf("db error")
	svc := NewMessageService(msgRepo, newMockMessageStatRepo(), newMockChatRepo(), nil, hub, nil)
	_, err := svc.SearchMessages(context.Background(), uuid.New(), "hello")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "search messages")
}

func TestMessageService_MarkChatAsRead_Error(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	msgStatRepo := newMockMessageStatRepo()
	msgStatRepo.markReadErr = fmt.Errorf("db error")
	svc := NewMessageService(newMockMessageRepo(), msgStatRepo, newMockChatRepo(), nil, hub, nil)
	err := svc.MarkChatAsRead(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mark chat as read")
}

func TestMessageService_SendMessage_WithNotifSvc(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()
	notif := &mockNotifSvc{}

	t.Run("personal chat push notification", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		msgRepo := newMockMessageRepo()
		userRepo := newMockUserRepo()
		userA := uuid.New()
		userB := uuid.New()
		userRepo.addUser(&model.User{ID: userA, Name: "Alice"})
		userRepo.addUser(&model.User{ID: userB, Name: "Bob"})
		chat := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal, CreatedBy: userA}
		chatRepo.chats[chat.ID] = chat
		_ = chatRepo.AddMember(context.Background(), chat.ID, userA, model.MemberRoleAdmin)
		_ = chatRepo.AddMember(context.Background(), chat.ID, userB, model.MemberRoleMember)

		svc := NewMessageService(msgRepo, newMockMessageStatRepo(), chatRepo, userRepo, hub, notif)
		msg, err := svc.SendMessage(context.Background(), SendMessageInput{
			ChatID: chat.ID, SenderID: userA, Content: "Hi Bob", Type: model.MessageTypeText,
		})
		require.NoError(t, err)
		assert.Equal(t, "Hi Bob", msg.Content)
		time.Sleep(50 * time.Millisecond) // allow goroutine to finish
	})

	t.Run("group chat push notification", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		msgRepo := newMockMessageRepo()
		userRepo := newMockUserRepo()
		userA := uuid.New()
		userRepo.addUser(&model.User{ID: userA, Name: "Alice"})
		chat := &model.Chat{ID: uuid.New(), Type: model.ChatTypeGroup, Name: "Team", CreatedBy: userA}
		chatRepo.chats[chat.ID] = chat
		_ = chatRepo.AddMember(context.Background(), chat.ID, userA, model.MemberRoleAdmin)

		svc := NewMessageService(msgRepo, newMockMessageStatRepo(), chatRepo, userRepo, hub, notif)
		msg, err := svc.SendMessage(context.Background(), SendMessageInput{
			ChatID: chat.ID, SenderID: userA, Content: "Hello team", Type: model.MessageTypeText,
		})
		require.NoError(t, err)
		assert.Equal(t, "Hello team", msg.Content)
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("push with sender name empty", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		msgRepo := newMockMessageRepo()
		userRepo := newMockUserRepo()
		userA := uuid.New()
		userRepo.addUser(&model.User{ID: userA, Name: ""}) // empty name
		chat := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal, CreatedBy: userA}
		chatRepo.chats[chat.ID] = chat
		_ = chatRepo.AddMember(context.Background(), chat.ID, userA, model.MemberRoleAdmin)

		svc := NewMessageService(msgRepo, newMockMessageStatRepo(), chatRepo, userRepo, hub, notif)
		msg, err := svc.SendMessage(context.Background(), SendMessageInput{
			ChatID: chat.ID, SenderID: userA, Content: "anon msg",
		})
		require.NoError(t, err)
		assert.Equal(t, "anon msg", msg.Content)
		time.Sleep(50 * time.Millisecond)
	})
}

func TestMessageService_ForwardMessage_CreateError(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	userA := uuid.New()

	chat1 := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal, CreatedBy: userA}
	chatRepo.chats[chat1.ID] = chat1
	_ = chatRepo.AddMember(context.Background(), chat1.ID, userA, model.MemberRoleAdmin)

	chat2 := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal, CreatedBy: userA}
	chatRepo.chats[chat2.ID] = chat2
	_ = chatRepo.AddMember(context.Background(), chat2.ID, userA, model.MemberRoleAdmin)

	svc := NewMessageService(msgRepo, newMockMessageStatRepo(), chatRepo, nil, hub, nil)
	original, err := svc.SendMessage(context.Background(), SendMessageInput{
		ChatID: chat1.ID, SenderID: userA, Content: "Fwd me", Type: model.MessageTypeText,
	})
	require.NoError(t, err)

	msgRepo.createErr = fmt.Errorf("db error")
	_, err = svc.ForwardMessage(context.Background(), original.ID, userA, chat2.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create forwarded message")
	msgRepo.createErr = nil
}

func TestMessageService_DeleteMessage_MarkError(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	userA := uuid.New()

	chat := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal, CreatedBy: userA}
	chatRepo.chats[chat.ID] = chat
	_ = chatRepo.AddMember(context.Background(), chat.ID, userA, model.MemberRoleAdmin)

	svc := NewMessageService(msgRepo, newMockMessageStatRepo(), chatRepo, nil, hub, nil)
	msg, err := svc.SendMessage(context.Background(), SendMessageInput{
		ChatID: chat.ID, SenderID: userA, Content: "Hi", Type: model.MessageTypeText,
	})
	require.NoError(t, err)

	msgRepo.markDelErr = fmt.Errorf("db error")
	err = svc.DeleteMessage(context.Background(), msg.ID, userA, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mark as deleted")
	msgRepo.markDelErr = nil
}
