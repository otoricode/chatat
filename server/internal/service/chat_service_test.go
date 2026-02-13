package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/ws"
	"github.com/otoritech/chatat/pkg/apperror"
)

// --- Mock Chat Repository ---
type mockChatRepo struct {
	chats   map[uuid.UUID]*model.Chat
	members map[uuid.UUID][]*model.ChatMember // chatID -> members
}

func newMockChatRepo() *mockChatRepo {
	return &mockChatRepo{
		chats:   make(map[uuid.UUID]*model.Chat),
		members: make(map[uuid.UUID][]*model.ChatMember),
	}
}

func (m *mockChatRepo) Create(_ context.Context, input model.CreateChatInput) (*model.Chat, error) {
	chat := &model.Chat{
		ID:        uuid.New(),
		Type:      input.Type,
		Name:      input.Name,
		Icon:      input.Icon,
		CreatedBy: input.CreatedBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.chats[chat.ID] = chat
	return chat, nil
}

func (m *mockChatRepo) FindByID(_ context.Context, id uuid.UUID) (*model.Chat, error) {
	c, ok := m.chats[id]
	if !ok {
		return nil, apperror.NotFound("chat", id.String())
	}
	return c, nil
}

func (m *mockChatRepo) FindPersonalChat(_ context.Context, userID1, userID2 uuid.UUID) (*model.Chat, error) {
	for chatID, chat := range m.chats {
		if chat.Type != model.ChatTypePersonal {
			continue
		}
		members := m.members[chatID]
		hasU1, hasU2 := false, false
		for _, mem := range members {
			if mem.UserID == userID1 {
				hasU1 = true
			}
			if mem.UserID == userID2 {
				hasU2 = true
			}
		}
		if hasU1 && hasU2 {
			return chat, nil
		}
	}
	return nil, apperror.NotFound("personal chat", userID1.String()+":"+userID2.String())
}

func (m *mockChatRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]*model.ChatWithLastMessage, error) {
	var result []*model.ChatWithLastMessage
	for chatID, chat := range m.chats {
		for _, mem := range m.members[chatID] {
			if mem.UserID == userID {
				result = append(result, &model.ChatWithLastMessage{Chat: *chat})
				break
			}
		}
	}
	return result, nil
}

func (m *mockChatRepo) AddMember(_ context.Context, chatID, userID uuid.UUID, role model.MemberRole) error {
	m.members[chatID] = append(m.members[chatID], &model.ChatMember{
		ChatID:   chatID,
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now(),
	})
	return nil
}

func (m *mockChatRepo) RemoveMember(_ context.Context, chatID, userID uuid.UUID) error {
	members := m.members[chatID]
	for i, mem := range members {
		if mem.UserID == userID {
			m.members[chatID] = append(members[:i], members[i+1:]...)
			return nil
		}
	}
	return apperror.NotFound("chat member", userID.String())
}

func (m *mockChatRepo) GetMembers(_ context.Context, chatID uuid.UUID) ([]*model.ChatMember, error) {
	return m.members[chatID], nil
}

func (m *mockChatRepo) Update(_ context.Context, id uuid.UUID, input model.UpdateChatInput) (*model.Chat, error) {
	c, ok := m.chats[id]
	if !ok {
		return nil, apperror.NotFound("chat", id.String())
	}
	if input.Name != nil {
		c.Name = *input.Name
	}
	return c, nil
}

func (m *mockChatRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := m.chats[id]; !ok {
		return apperror.NotFound("chat", id.String())
	}
	delete(m.chats, id)
	return nil
}

func (m *mockChatRepo) Pin(_ context.Context, id uuid.UUID) error {
	c, ok := m.chats[id]
	if !ok {
		return apperror.NotFound("chat", id.String())
	}
	now := time.Now()
	c.PinnedAt = &now
	return nil
}

func (m *mockChatRepo) Unpin(_ context.Context, id uuid.UUID) error {
	c, ok := m.chats[id]
	if !ok {
		return apperror.NotFound("chat", id.String())
	}
	c.PinnedAt = nil
	return nil
}

// --- Mock Message Repository ---
type mockMessageRepo struct {
	messages map[uuid.UUID]*model.Message
	byChat   map[uuid.UUID][]*model.Message
}

func newMockMessageRepo() *mockMessageRepo {
	return &mockMessageRepo{
		messages: make(map[uuid.UUID]*model.Message),
		byChat:   make(map[uuid.UUID][]*model.Message),
	}
}

func (m *mockMessageRepo) Create(_ context.Context, input model.CreateMessageInput) (*model.Message, error) {
	msgType := input.Type
	if msgType == "" {
		msgType = model.MessageTypeText
	}
	msg := &model.Message{
		ID:        uuid.New(),
		ChatID:    input.ChatID,
		SenderID:  input.SenderID,
		Content:   input.Content,
		ReplyToID: input.ReplyToID,
		Type:      msgType,
		Metadata:  input.Metadata,
		CreatedAt: time.Now(),
	}
	m.messages[msg.ID] = msg
	m.byChat[input.ChatID] = append([]*model.Message{msg}, m.byChat[input.ChatID]...)
	return msg, nil
}

func (m *mockMessageRepo) FindByID(_ context.Context, id uuid.UUID) (*model.Message, error) {
	msg, ok := m.messages[id]
	if !ok {
		return nil, apperror.NotFound("message", id.String())
	}
	return msg, nil
}

func (m *mockMessageRepo) ListByChat(_ context.Context, chatID uuid.UUID, cursor *time.Time, limit int) ([]*model.Message, error) {
	msgs := m.byChat[chatID]
	if limit <= 0 {
		limit = 50
	}
	var result []*model.Message
	for _, msg := range msgs {
		if cursor != nil && !msg.CreatedAt.Before(*cursor) {
			continue
		}
		result = append(result, msg)
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func (m *mockMessageRepo) MarkAsDeleted(_ context.Context, id uuid.UUID, forAll bool) error {
	msg, ok := m.messages[id]
	if !ok {
		return apperror.NotFound("message", id.String())
	}
	msg.IsDeleted = true
	msg.DeletedForAll = forAll
	return nil
}

func (m *mockMessageRepo) Search(_ context.Context, _ uuid.UUID, _ string) ([]*model.Message, error) {
	return nil, nil
}

// --- Mock Message Status Repository ---
type mockMessageStatRepo struct {
	statuses    map[string]*model.MessageStatus // key: "msgID:userID"
	unreadCount map[string]int                  // key: "chatID:userID"
}

func newMockMessageStatRepo() *mockMessageStatRepo {
	return &mockMessageStatRepo{
		statuses:    make(map[string]*model.MessageStatus),
		unreadCount: make(map[string]int),
	}
}

func (m *mockMessageStatRepo) Create(_ context.Context, messageID, userID uuid.UUID, status model.DeliveryStatus) error {
	key := messageID.String() + ":" + userID.String()
	m.statuses[key] = &model.MessageStatus{
		MessageID: messageID,
		UserID:    userID,
		Status:    status,
	}
	return nil
}

func (m *mockMessageStatRepo) UpdateStatus(_ context.Context, messageID, userID uuid.UUID, status model.DeliveryStatus) error {
	key := messageID.String() + ":" + userID.String()
	if s, ok := m.statuses[key]; ok {
		s.Status = status
	}
	return nil
}

func (m *mockMessageStatRepo) GetStatus(_ context.Context, messageID uuid.UUID) ([]*model.MessageStatus, error) {
	var result []*model.MessageStatus
	for _, s := range m.statuses {
		if s.MessageID == messageID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *mockMessageStatRepo) MarkChatAsRead(_ context.Context, chatID, userID uuid.UUID) error {
	key := chatID.String() + ":" + userID.String()
	m.unreadCount[key] = 0
	return nil
}

func (m *mockMessageStatRepo) GetUnreadCount(_ context.Context, chatID, userID uuid.UUID) (int, error) {
	key := chatID.String() + ":" + userID.String()
	return m.unreadCount[key], nil
}

// --- Helper to create test hub ---
func newTestHub() *ws.Hub {
	hub := ws.NewHub()
	go hub.Run()
	return hub
}

// ==================== ChatService Tests ====================

func TestChatService_CreatePersonalChat(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewChatService(chatRepo, msgRepo, msgStatRepo, userRepo, hub)

	userA := uuid.New()
	userB := uuid.New()
	userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "\U0001F60A"})
	userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B", Avatar: "\U0001F60A"})

	t.Run("success", func(t *testing.T) {
		chat, err := svc.CreatePersonalChat(context.Background(), userA, userB)
		require.NoError(t, err)
		assert.Equal(t, model.ChatTypePersonal, chat.Type)
		assert.Equal(t, userA, chat.CreatedBy)
	})

	t.Run("cannot chat with yourself", func(t *testing.T) {
		_, err := svc.CreatePersonalChat(context.Background(), userA, userA)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "yourself")
	})

	t.Run("contact not found", func(t *testing.T) {
		_, err := svc.CreatePersonalChat(context.Background(), userA, uuid.New())
		require.Error(t, err)
		assert.True(t, apperror.IsNotFound(err))
	})
}

func TestChatService_GetOrCreatePersonalChat(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewChatService(chatRepo, msgRepo, msgStatRepo, userRepo, hub)

	userA := uuid.New()
	userB := uuid.New()
	userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "\U0001F60A"})
	userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B", Avatar: "\U0001F60A"})

	t.Run("create new", func(t *testing.T) {
		chat1, err := svc.GetOrCreatePersonalChat(context.Background(), userA, userB)
		require.NoError(t, err)
		assert.NotNil(t, chat1)

		// Call again → should return same chat
		chat2, err := svc.GetOrCreatePersonalChat(context.Background(), userA, userB)
		require.NoError(t, err)
		assert.Equal(t, chat1.ID, chat2.ID)
	})
}

func TestChatService_ListChats(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewChatService(chatRepo, msgRepo, msgStatRepo, userRepo, hub)

	userA := uuid.New()
	userB := uuid.New()
	userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "\U0001F60A"})
	userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B", Avatar: "\U0001F60A"})

	// Create a chat
	chat, err := svc.CreatePersonalChat(context.Background(), userA, userB)
	require.NoError(t, err)

	t.Run("list for user A", func(t *testing.T) {
		items, err := svc.ListChats(context.Background(), userA)
		require.NoError(t, err)
		assert.Len(t, items, 1)
		assert.Equal(t, chat.ID, items[0].Chat.ID)
		assert.NotNil(t, items[0].OtherUser)
		assert.Equal(t, "B", items[0].OtherUser.Name)
	})

	t.Run("empty for unknown user", func(t *testing.T) {
		items, err := svc.ListChats(context.Background(), uuid.New())
		require.NoError(t, err)
		assert.Len(t, items, 0)
	})
}

func TestChatService_PinUnpin(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewChatService(chatRepo, msgRepo, msgStatRepo, userRepo, hub)

	userA := uuid.New()
	userB := uuid.New()
	userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "\U0001F60A"})
	userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B", Avatar: "\U0001F60A"})

	chat, err := svc.CreatePersonalChat(context.Background(), userA, userB)
	require.NoError(t, err)

	t.Run("pin chat", func(t *testing.T) {
		err := svc.PinChat(context.Background(), chat.ID, userA)
		require.NoError(t, err)
		assert.NotNil(t, chatRepo.chats[chat.ID].PinnedAt)
	})

	t.Run("unpin chat", func(t *testing.T) {
		err := svc.UnpinChat(context.Background(), chat.ID, userA)
		require.NoError(t, err)
		assert.Nil(t, chatRepo.chats[chat.ID].PinnedAt)
	})

	t.Run("pin not a member", func(t *testing.T) {
		err := svc.PinChat(context.Background(), chat.ID, uuid.New())
		require.Error(t, err)
	})
}

func TestChatService_IsMember(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewChatService(chatRepo, msgRepo, msgStatRepo, userRepo, hub)

	userA := uuid.New()
	userB := uuid.New()
	userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "\U0001F60A"})
	userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B", Avatar: "\U0001F60A"})

	chat, err := svc.CreatePersonalChat(context.Background(), userA, userB)
	require.NoError(t, err)

	t.Run("is member", func(t *testing.T) {
		ok, err := svc.IsMember(context.Background(), chat.ID, userA)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("not a member", func(t *testing.T) {
		ok, err := svc.IsMember(context.Background(), chat.ID, uuid.New())
		require.NoError(t, err)
		assert.False(t, ok)
	})
}
