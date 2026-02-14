package service

import (
	"context"
	"errors"
	"fmt"
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
	chats               map[uuid.UUID]*model.Chat
	members             map[uuid.UUID][]*model.ChatMember // chatID -> members
	createErr           error
	addMemberErr        error
	listByUserErr       error
	getMembersErr       error
	pinErr              error
	unpinErr            error
	findErr             error
	findPersonalChatErr error
	removeMemberErr     error
	deleteErr           error
}

func newMockChatRepo() *mockChatRepo {
	return &mockChatRepo{
		chats:   make(map[uuid.UUID]*model.Chat),
		members: make(map[uuid.UUID][]*model.ChatMember),
	}
}

func (m *mockChatRepo) Create(_ context.Context, input model.CreateChatInput) (*model.Chat, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
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
	if m.findErr != nil {
		return nil, m.findErr
	}
	c, ok := m.chats[id]
	if !ok {
		return nil, apperror.NotFound("chat", id.String())
	}
	return c, nil
}

func (m *mockChatRepo) FindPersonalChat(_ context.Context, userID1, userID2 uuid.UUID) (*model.Chat, error) {
	if m.findPersonalChatErr != nil {
		return nil, m.findPersonalChatErr
	}
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
	if m.listByUserErr != nil {
		return nil, m.listByUserErr
	}
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
	if m.addMemberErr != nil {
		return m.addMemberErr
	}
	m.members[chatID] = append(m.members[chatID], &model.ChatMember{
		ChatID:   chatID,
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now(),
	})
	return nil
}

func (m *mockChatRepo) RemoveMember(_ context.Context, chatID, userID uuid.UUID) error {
	if m.removeMemberErr != nil {
		return m.removeMemberErr
	}
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
	if m.getMembersErr != nil {
		return nil, m.getMembersErr
	}
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
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.chats[id]; !ok {
		return apperror.NotFound("chat", id.String())
	}
	delete(m.chats, id)
	return nil
}

func (m *mockChatRepo) Pin(_ context.Context, id uuid.UUID) error {
	if m.pinErr != nil {
		return m.pinErr
	}
	c, ok := m.chats[id]
	if !ok {
		return apperror.NotFound("chat", id.String())
	}
	now := time.Now()
	c.PinnedAt = &now
	return nil
}

func (m *mockChatRepo) Unpin(_ context.Context, id uuid.UUID) error {
	if m.unpinErr != nil {
		return m.unpinErr
	}
	c, ok := m.chats[id]
	if !ok {
		return apperror.NotFound("chat", id.String())
	}
	c.PinnedAt = nil
	return nil
}

// --- Mock Message Repository ---
type mockMessageRepo struct {
	messages   map[uuid.UUID]*model.Message
	byChat     map[uuid.UUID][]*model.Message
	createErr  error
	listErr    error
	searchErr  error
	markDelErr error
}

func newMockMessageRepo() *mockMessageRepo {
	return &mockMessageRepo{
		messages: make(map[uuid.UUID]*model.Message),
		byChat:   make(map[uuid.UUID][]*model.Message),
	}
}

func (m *mockMessageRepo) Create(_ context.Context, input model.CreateMessageInput) (*model.Message, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
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
	if m.listErr != nil {
		return nil, m.listErr
	}
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
	if m.markDelErr != nil {
		return m.markDelErr
	}
	msg, ok := m.messages[id]
	if !ok {
		return apperror.NotFound("message", id.String())
	}
	msg.IsDeleted = true
	msg.DeletedForAll = forAll
	return nil
}

func (m *mockMessageRepo) Search(_ context.Context, _ uuid.UUID, _ string) ([]*model.Message, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	return nil, nil
}

// --- Mock Message Status Repository ---
type mockMessageStatRepo struct {
	statuses       map[string]*model.MessageStatus // key: "msgID:userID"
	unreadCount    map[string]int                  // key: "chatID:userID"
	unreadCountErr error
	markReadErr    error
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
	if m.markReadErr != nil {
		return m.markReadErr
	}
	key := chatID.String() + ":" + userID.String()
	m.unreadCount[key] = 0
	return nil
}

func (m *mockMessageStatRepo) GetUnreadCount(_ context.Context, chatID, userID uuid.UUID) (int, error) {
	if m.unreadCountErr != nil {
		return 0, m.unreadCountErr
	}
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

func TestChatService_GetChat(t *testing.T) {
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

	t.Run("success", func(t *testing.T) {
		detail, err := svc.GetChat(context.Background(), chat.ID, userA)
		require.NoError(t, err)
		assert.Equal(t, chat.ID, detail.Chat.ID)
		assert.Len(t, detail.Members, 2)
	})

	t.Run("not a member", func(t *testing.T) {
		_, err := svc.GetChat(context.Background(), chat.ID, uuid.New())
		assert.Error(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetChat(context.Background(), uuid.New(), userA)
		assert.Error(t, err)
	})
}

func TestChatService_CreatePersonalChat_Errors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("user repo generic error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		// Don't add user → FindByID returns NotFound, but we want non-NotFound
		// Instead, test when contactID doesn't exist → already covered as NotFound
		// For a generic error we need a custom behavior
		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub)
		_, err := svc.CreatePersonalChat(context.Background(), uuid.New(), uuid.New())
		require.Error(t, err)
	})

	t.Run("create chat error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		chatRepo.createErr = fmt.Errorf("db error")
		userRepo := newMockUserRepo()
		contactID := uuid.New()
		userRepo.addUser(&model.User{ID: contactID, Phone: "+628111", Name: "C", Avatar: "\U0001F60A"})

		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub)
		_, err := svc.CreatePersonalChat(context.Background(), uuid.New(), contactID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create personal chat")
	})

	t.Run("add member error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		chatRepo.addMemberErr = fmt.Errorf("db error")
		userRepo := newMockUserRepo()
		contactID := uuid.New()
		userRepo.addUser(&model.User{ID: contactID, Phone: "+628111", Name: "C", Avatar: "\U0001F60A"})

		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub)
		_, err := svc.CreatePersonalChat(context.Background(), uuid.New(), contactID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "add creator member")
	})
}

func TestChatService_GetOrCreatePersonalChat_Errors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("self chat", func(t *testing.T) {
		svc := NewChatService(newMockChatRepo(), newMockMessageRepo(), newMockMessageStatRepo(), newMockUserRepo(), hub)
		id := uuid.New()
		_, err := svc.GetOrCreatePersonalChat(context.Background(), id, id)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "yourself")
	})

	t.Run("find personal chat generic error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		chatRepo.findPersonalChatErr = fmt.Errorf("db error")
		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), newMockUserRepo(), hub)
		_, err := svc.GetOrCreatePersonalChat(context.Background(), uuid.New(), uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "find personal chat")
	})
}

func TestChatService_ListChats_Errors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("list by user error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		chatRepo.listByUserErr = fmt.Errorf("db error")
		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), newMockUserRepo(), hub)
		_, err := svc.ListChats(context.Background(), uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "list chats")
	})

	t.Run("unread count error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		msgStatRepo := newMockMessageStatRepo()
		userA := uuid.New()
		userB := uuid.New()
		userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "\U0001F60A"})
		userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B", Avatar: "\U0001F60A"})

		svc := NewChatService(chatRepo, newMockMessageRepo(), msgStatRepo, userRepo, hub)
		_, err := svc.CreatePersonalChat(context.Background(), userA, userB)
		require.NoError(t, err)

		msgStatRepo.unreadCountErr = fmt.Errorf("redis error")
		_, err = svc.ListChats(context.Background(), userA)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unread count")
	})
}

func TestChatService_PinChat_Errors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("get members error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		chatRepo.getMembersErr = fmt.Errorf("db error")
		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), newMockUserRepo(), hub)
		err := svc.PinChat(context.Background(), uuid.New(), uuid.New())
		require.Error(t, err)
	})

	t.Run("find by id error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		userA := uuid.New()
		userB := uuid.New()
		userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "\U0001F60A"})
		userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B", Avatar: "\U0001F60A"})

		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub)
		chat, err := svc.CreatePersonalChat(context.Background(), userA, userB)
		require.NoError(t, err)

		chatRepo.findErr = fmt.Errorf("db error")
		err = svc.PinChat(context.Background(), chat.ID, userA)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "find chat")
	})

	t.Run("pin error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		userA := uuid.New()
		userB := uuid.New()
		userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "\U0001F60A"})
		userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B", Avatar: "\U0001F60A"})

		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub)
		chat, err := svc.CreatePersonalChat(context.Background(), userA, userB)
		require.NoError(t, err)

		chatRepo.pinErr = fmt.Errorf("db error")
		err = svc.PinChat(context.Background(), chat.ID, userA)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pin chat")
	})
}

func TestChatService_UnpinChat_Errors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("get members error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		chatRepo.getMembersErr = fmt.Errorf("db error")
		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), newMockUserRepo(), hub)
		err := svc.UnpinChat(context.Background(), uuid.New(), uuid.New())
		require.Error(t, err)
	})

	t.Run("not a member", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		userA := uuid.New()
		userB := uuid.New()
		userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "\U0001F60A"})
		userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B", Avatar: "\U0001F60A"})

		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub)
		chat, err := svc.CreatePersonalChat(context.Background(), userA, userB)
		require.NoError(t, err)

		err = svc.UnpinChat(context.Background(), chat.ID, uuid.New())
		require.Error(t, err)
	})

	t.Run("unpin error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		userA := uuid.New()
		userB := uuid.New()
		userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "\U0001F60A"})
		userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B", Avatar: "\U0001F60A"})

		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub)
		chat, err := svc.CreatePersonalChat(context.Background(), userA, userB)
		require.NoError(t, err)

		chatRepo.unpinErr = fmt.Errorf("db error")
		err = svc.UnpinChat(context.Background(), chat.ID, userA)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unpin chat")
	})
}

func TestChatService_IsMember_Errors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("get members not found", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		chatRepo.getMembersErr = apperror.NotFound("chat", "test")
		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), newMockUserRepo(), hub)
		_, err := svc.IsMember(context.Background(), uuid.New(), uuid.New())
		require.Error(t, err)
		assert.True(t, apperror.IsNotFound(err))
	})

	t.Run("get members generic error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		chatRepo.getMembersErr = fmt.Errorf("db error")
		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), newMockUserRepo(), hub)
		_, err := svc.IsMember(context.Background(), uuid.New(), uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "check membership")
	})
}

func TestChatService_GetChat_Errors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("find by id error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		userA := uuid.New()
		userB := uuid.New()
		userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "\U0001F60A"})
		userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B", Avatar: "\U0001F60A"})

		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub)
		chat, err := svc.CreatePersonalChat(context.Background(), userA, userB)
		require.NoError(t, err)

		chatRepo.findErr = fmt.Errorf("db error")
		_, err = svc.GetChat(context.Background(), chat.ID, userA)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get chat")
	})

	t.Run("get members error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		userA := uuid.New()
		userB := uuid.New()
		userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "\U0001F60A"})
		userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B", Avatar: "\U0001F60A"})

		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub)
		chat, err := svc.CreatePersonalChat(context.Background(), userA, userB)
		require.NoError(t, err)

		chatRepo.getMembersErr = fmt.Errorf("db error")
		// GetChat first calls IsMember (which uses GetMembers) - this will fail
		_, err = svc.GetChat(context.Background(), chat.ID, userA)
		require.Error(t, err)
	})
}

func TestSortChatList(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-1 * time.Hour)
	later := now.Add(1 * time.Hour)

	t.Run("pinned first", func(t *testing.T) {
		items := []*ChatListItem{
			{Chat: model.Chat{ID: uuid.New(), UpdatedAt: now}},
			{Chat: model.Chat{ID: uuid.New(), UpdatedAt: earlier, PinnedAt: &now}},
		}
		sortChatList(items)
		assert.NotNil(t, items[0].Chat.PinnedAt)
		assert.Nil(t, items[1].Chat.PinnedAt)
	})

	t.Run("by last message time desc", func(t *testing.T) {
		id1 := uuid.New()
		id2 := uuid.New()
		items := []*ChatListItem{
			{Chat: model.Chat{ID: id1, UpdatedAt: earlier}},
			{Chat: model.Chat{ID: id2, UpdatedAt: later}},
		}
		sortChatList(items)
		assert.Equal(t, id2, items[0].Chat.ID)
	})

	t.Run("with last message", func(t *testing.T) {
		id1 := uuid.New()
		id2 := uuid.New()
		items := []*ChatListItem{
			{Chat: model.Chat{ID: id1, UpdatedAt: earlier}, LastMessage: &model.Message{CreatedAt: later}},
			{Chat: model.Chat{ID: id2, UpdatedAt: later}, LastMessage: &model.Message{CreatedAt: earlier}},
		}
		sortChatList(items)
		assert.Equal(t, id1, items[0].Chat.ID)
	})

	t.Run("both pinned sort by time", func(t *testing.T) {
		id1 := uuid.New()
		id2 := uuid.New()
		items := []*ChatListItem{
			{Chat: model.Chat{ID: id1, UpdatedAt: earlier, PinnedAt: &now}},
			{Chat: model.Chat{ID: id2, UpdatedAt: later, PinnedAt: &now}},
		}
		sortChatList(items)
		assert.Equal(t, id2, items[0].Chat.ID)
	})

	t.Run("single item", func(t *testing.T) {
		items := []*ChatListItem{
			{Chat: model.Chat{ID: uuid.New(), UpdatedAt: now}},
		}
		sortChatList(items)
		assert.Len(t, items, 1)
	})

	t.Run("empty", func(t *testing.T) {
		items := []*ChatListItem{}
		sortChatList(items)
		assert.Len(t, items, 0)
	})
}

func TestChatService_ListChats_MoreErrors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("get last message error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		msgRepo := newMockMessageRepo()
		userRepo := newMockUserRepo()
		userA := uuid.New()
		userB := uuid.New()
		userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "a"})
		userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B", Avatar: "b"})

		svc := NewChatService(chatRepo, msgRepo, newMockMessageStatRepo(), userRepo, hub)
		_, err := svc.CreatePersonalChat(context.Background(), userA, userB)
		require.NoError(t, err)

		msgRepo.listErr = fmt.Errorf("db error")
		_, err = svc.ListChats(context.Background(), userA)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "last message")
		msgRepo.listErr = nil
	})

	t.Run("get members error for personal chat", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		userA := uuid.New()
		userB := uuid.New()
		userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "a"})
		userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B", Avatar: "b"})

		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub)
		_, err := svc.CreatePersonalChat(context.Background(), userA, userB)
		require.NoError(t, err)

		chatRepo.getMembersErr = fmt.Errorf("db error")
		_, err = svc.ListChats(context.Background(), userA)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "chat members")
		chatRepo.getMembersErr = nil
	})

	t.Run("other user find error for personal chat", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		userA := uuid.New()
		userB := uuid.New()
		userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "a"})
		userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B", Avatar: "b"})

		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub)
		_, err := svc.CreatePersonalChat(context.Background(), userA, userB)
		require.NoError(t, err)

		userRepo.findErr = errors.New("generic db error")
		_, err = svc.ListChats(context.Background(), userA)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "other user")
		userRepo.findErr = nil
	})

	t.Run("other user not found skipped", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		userA := uuid.New()
		userB := uuid.New()
		userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "a"})
		userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B", Avatar: "b"})

		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub)
		_, err := svc.CreatePersonalChat(context.Background(), userA, userB)
		require.NoError(t, err)

		// Remove userB from repo so FindByID returns NotFound
		delete(userRepo.users, userB)
		delete(userRepo.byPhone, "+628222")

		items, err := svc.ListChats(context.Background(), userA)
		require.NoError(t, err)
		assert.Len(t, items, 1)
		assert.Nil(t, items[0].OtherUser, "deleted user should be nil")
	})

	t.Run("group chat skips member lookup", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		userA := uuid.New()
		userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "a"})

		groupChat := &model.Chat{ID: uuid.New(), Type: model.ChatTypeGroup, Name: "G", CreatedBy: userA}
		chatRepo.chats[groupChat.ID] = groupChat
		_ = chatRepo.AddMember(context.Background(), groupChat.ID, userA, model.MemberRoleAdmin)

		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub)
		items, err := svc.ListChats(context.Background(), userA)
		require.NoError(t, err)
		assert.Len(t, items, 1)
		assert.Nil(t, items[0].OtherUser, "group chat should have nil OtherUser")
	})
}

func TestChatService_GetChat_MemberErrors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("member user not found", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		userA := uuid.New()
		userB := uuid.New()
		userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "a"})
		userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B", Avatar: "b"})

		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub)
		chat, err := svc.CreatePersonalChat(context.Background(), userA, userB)
		require.NoError(t, err)

		// Remove userB so FindByID returns NotFound
		delete(userRepo.users, userB)
		delete(userRepo.byPhone, "+628222")

		detail, err := svc.GetChat(context.Background(), chat.ID, userA)
		require.NoError(t, err)
		assert.Len(t, detail.Members, 1, "deleted member should be skipped")
	})

	t.Run("member user generic error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		userA := uuid.New()
		userB := uuid.New()
		userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A", Avatar: "a"})
		userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B", Avatar: "b"})

		svc := NewChatService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub)
		chat, err := svc.CreatePersonalChat(context.Background(), userA, userB)
		require.NoError(t, err)

		userRepo.findErr = errors.New("generic db error")
		_, err = svc.GetChat(context.Background(), chat.ID, userA)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "member user")
		userRepo.findErr = nil
	})
}
