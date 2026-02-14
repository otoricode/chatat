package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/pkg/apperror"
)

// --- Mock Topic Repository ---
type mockTopicRepo struct {
	topics  map[uuid.UUID]*model.Topic
	members map[uuid.UUID][]*model.TopicMember // topicID -> members

	getMembersErr error
}

func newMockTopicRepo() *mockTopicRepo {
	return &mockTopicRepo{
		topics:  make(map[uuid.UUID]*model.Topic),
		members: make(map[uuid.UUID][]*model.TopicMember),
	}
}

func (m *mockTopicRepo) Create(_ context.Context, input model.CreateTopicInput) (*model.Topic, error) {
	topic := &model.Topic{
		ID:          uuid.New(),
		Name:        input.Name,
		Icon:        input.Icon,
		Description: input.Description,
		ParentType:  input.ParentType,
		ParentID:    input.ParentID,
		CreatedBy:   input.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if topic.Icon == "" {
		topic.Icon = "💬"
	}
	m.topics[topic.ID] = topic
	return topic, nil
}

func (m *mockTopicRepo) FindByID(_ context.Context, id uuid.UUID) (*model.Topic, error) {
	t, ok := m.topics[id]
	if !ok {
		return nil, apperror.NotFound("topic", id.String())
	}
	return t, nil
}

func (m *mockTopicRepo) ListByParent(_ context.Context, parentID uuid.UUID) ([]*model.Topic, error) {
	var result []*model.Topic
	for _, t := range m.topics {
		if t.ParentID == parentID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTopicRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]*model.Topic, error) {
	var result []*model.Topic
	for topicID, t := range m.topics {
		for _, mem := range m.members[topicID] {
			if mem.UserID == userID {
				result = append(result, t)
				break
			}
		}
	}
	return result, nil
}

func (m *mockTopicRepo) AddMember(_ context.Context, topicID, userID uuid.UUID, role model.MemberRole) error {
	// Check for duplicate
	for _, mem := range m.members[topicID] {
		if mem.UserID == userID {
			return nil // ON CONFLICT DO NOTHING
		}
	}
	m.members[topicID] = append(m.members[topicID], &model.TopicMember{
		TopicID:  topicID,
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now(),
	})
	return nil
}

func (m *mockTopicRepo) RemoveMember(_ context.Context, topicID, userID uuid.UUID) error {
	members := m.members[topicID]
	for i, mem := range members {
		if mem.UserID == userID {
			m.members[topicID] = append(members[:i], members[i+1:]...)
			return nil
		}
	}
	return apperror.NotFound("topic member", userID.String())
}

func (m *mockTopicRepo) GetMembers(_ context.Context, topicID uuid.UUID) ([]*model.TopicMember, error) {
	if m.getMembersErr != nil {
		return nil, m.getMembersErr
	}
	return m.members[topicID], nil
}

func (m *mockTopicRepo) Update(_ context.Context, id uuid.UUID, input model.UpdateTopicInput) (*model.Topic, error) {
	t, ok := m.topics[id]
	if !ok {
		return nil, apperror.NotFound("topic", id.String())
	}
	if input.Name != nil {
		t.Name = *input.Name
	}
	if input.Icon != nil {
		t.Icon = *input.Icon
	}
	if input.Description != nil {
		t.Description = *input.Description
	}
	t.UpdatedAt = time.Now()
	return t, nil
}

func (m *mockTopicRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := m.topics[id]; !ok {
		return apperror.NotFound("topic", id.String())
	}
	delete(m.topics, id)
	delete(m.members, id)
	return nil
}

// --- Mock Topic Message Repository ---
type mockTopicMsgRepo struct {
	messages map[uuid.UUID]*model.TopicMessage
	byTopic  map[uuid.UUID][]*model.TopicMessage

	createErr      error
	findErr        error
	listErr        error
	markDeletedErr error
}

func newMockTopicMsgRepo() *mockTopicMsgRepo {
	return &mockTopicMsgRepo{
		messages: make(map[uuid.UUID]*model.TopicMessage),
		byTopic:  make(map[uuid.UUID][]*model.TopicMessage),
	}
}

func (m *mockTopicMsgRepo) Create(_ context.Context, input model.CreateTopicMessageInput) (*model.TopicMessage, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	msgType := input.Type
	if msgType == "" {
		msgType = model.MessageTypeText
	}
	msg := &model.TopicMessage{
		ID:        uuid.New(),
		TopicID:   input.TopicID,
		SenderID:  input.SenderID,
		Content:   input.Content,
		ReplyToID: input.ReplyToID,
		Type:      msgType,
		CreatedAt: time.Now(),
	}
	m.messages[msg.ID] = msg
	m.byTopic[input.TopicID] = append([]*model.TopicMessage{msg}, m.byTopic[input.TopicID]...)
	return msg, nil
}

func (m *mockTopicMsgRepo) FindByID(_ context.Context, id uuid.UUID) (*model.TopicMessage, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	msg, ok := m.messages[id]
	if !ok {
		return nil, apperror.NotFound("topic message", id.String())
	}
	return msg, nil
}

func (m *mockTopicMsgRepo) ListByTopic(_ context.Context, topicID uuid.UUID, cursor *time.Time, limit int) ([]*model.TopicMessage, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	msgs := m.byTopic[topicID]
	if limit <= 0 {
		limit = 50
	}
	var result []*model.TopicMessage
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

func (m *mockTopicMsgRepo) MarkAsDeleted(_ context.Context, id uuid.UUID, forAll bool) error {
	if m.markDeletedErr != nil {
		return m.markDeletedErr
	}
	msg, ok := m.messages[id]
	if !ok {
		return apperror.NotFound("topic message", id.String())
	}
	msg.IsDeleted = true
	msg.DeletedForAll = forAll
	return nil
}

// --- Topic Service Tests ---

func setupTopicService() (TopicService, *mockTopicRepo, *mockTopicMsgRepo, *mockChatRepo, *mockUserRepo) {
	topicRepo := newMockTopicRepo()
	topicMsgRepo := newMockTopicMsgRepo()
	chatRepo := newMockChatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewTopicService(topicRepo, topicMsgRepo, chatRepo, userRepo, hub)
	return svc, topicRepo, topicMsgRepo, chatRepo, userRepo
}

func TestTopicService_CreateFromPersonalChat(t *testing.T) {
	svc, _, _, chatRepo, userRepo := setupTopicService()

	userA := uuid.New()
	userB := uuid.New()
	userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "Andi"})
	userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "Budi"})

	// Create personal chat
	chat, _ := chatRepo.Create(context.Background(), model.CreateChatInput{
		Type:      model.ChatTypePersonal,
		CreatedBy: userA,
	})
	_ = chatRepo.AddMember(context.Background(), chat.ID, userA, model.MemberRoleAdmin)
	_ = chatRepo.AddMember(context.Background(), chat.ID, userB, model.MemberRoleMember)

	// Create topic from personal chat
	topic, err := svc.CreateTopic(context.Background(), userA, CreateTopicInput{
		Name:     "Pembagian Lahan",
		Icon:     "🌾",
		ParentID: chat.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, "Pembagian Lahan", topic.Name)
	assert.Equal(t, "🌾", topic.Icon)
	assert.Equal(t, model.ChatTypePersonal, topic.ParentType)
}

func TestTopicService_CreateFromGroupChat(t *testing.T) {
	svc, topicRepo, _, chatRepo, userRepo := setupTopicService()

	creator := uuid.New()
	memberA := uuid.New()
	memberB := uuid.New()
	userRepo.addUser(&model.User{ID: creator, Phone: "+628111", Name: "Creator"})
	userRepo.addUser(&model.User{ID: memberA, Phone: "+628222", Name: "MemberA"})
	userRepo.addUser(&model.User{ID: memberB, Phone: "+628333", Name: "MemberB"})

	// Create group chat
	chat, _ := chatRepo.Create(context.Background(), model.CreateChatInput{
		Type:      model.ChatTypeGroup,
		Name:      "Tim Proyek",
		CreatedBy: creator,
	})
	_ = chatRepo.AddMember(context.Background(), chat.ID, creator, model.MemberRoleAdmin)
	_ = chatRepo.AddMember(context.Background(), chat.ID, memberA, model.MemberRoleMember)
	_ = chatRepo.AddMember(context.Background(), chat.ID, memberB, model.MemberRoleMember)

	t.Run("with subset members", func(t *testing.T) {
		topic, err := svc.CreateTopic(context.Background(), creator, CreateTopicInput{
			Name:      "Desain UI",
			Icon:      "🎨",
			ParentID:  chat.ID,
			MemberIDs: []uuid.UUID{creator, memberA},
		})
		require.NoError(t, err)
		assert.Equal(t, "Desain UI", topic.Name)
		assert.Equal(t, model.ChatTypeGroup, topic.ParentType)

		// Check members
		members, _ := topicRepo.GetMembers(context.Background(), topic.ID)
		assert.Len(t, members, 2) // creator + memberA
	})

	t.Run("with all members (no memberIDs)", func(t *testing.T) {
		topic, err := svc.CreateTopic(context.Background(), creator, CreateTopicInput{
			Name:     "General",
			Icon:     "📢",
			ParentID: chat.ID,
		})
		require.NoError(t, err)

		members, _ := topicRepo.GetMembers(context.Background(), topic.ID)
		assert.Len(t, members, 3) // all group members
	})
}

func TestTopicService_CreateInvalidMember(t *testing.T) {
	svc, _, _, chatRepo, userRepo := setupTopicService()

	creator := uuid.New()
	memberA := uuid.New()
	outsider := uuid.New()
	userRepo.addUser(&model.User{ID: creator, Phone: "+628111", Name: "Creator"})
	userRepo.addUser(&model.User{ID: memberA, Phone: "+628222", Name: "MemberA"})
	userRepo.addUser(&model.User{ID: outsider, Phone: "+628444", Name: "Outsider"})

	chat, _ := chatRepo.Create(context.Background(), model.CreateChatInput{
		Type:      model.ChatTypeGroup,
		Name:      "Tim",
		CreatedBy: creator,
	})
	_ = chatRepo.AddMember(context.Background(), chat.ID, creator, model.MemberRoleAdmin)
	_ = chatRepo.AddMember(context.Background(), chat.ID, memberA, model.MemberRoleMember)

	// Outsider not in parent chat
	_, err := svc.CreateTopic(context.Background(), creator, CreateTopicInput{
		Name:      "Test",
		Icon:      "📌",
		ParentID:  chat.ID,
		MemberIDs: []uuid.UUID{creator, outsider},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not in parent chat")
}

func TestTopicService_CreateValidation(t *testing.T) {
	svc, _, _, chatRepo, userRepo := setupTopicService()

	creator := uuid.New()
	userRepo.addUser(&model.User{ID: creator, Phone: "+628111", Name: "Creator"})

	chat, _ := chatRepo.Create(context.Background(), model.CreateChatInput{
		Type:      model.ChatTypePersonal,
		CreatedBy: creator,
	})
	_ = chatRepo.AddMember(context.Background(), chat.ID, creator, model.MemberRoleAdmin)

	t.Run("missing name", func(t *testing.T) {
		_, err := svc.CreateTopic(context.Background(), creator, CreateTopicInput{
			Icon:     "📌",
			ParentID: chat.ID,
		})
		require.Error(t, err)
	})

	t.Run("missing icon", func(t *testing.T) {
		_, err := svc.CreateTopic(context.Background(), creator, CreateTopicInput{
			Name:     "Test",
			ParentID: chat.ID,
		})
		require.Error(t, err)
	})

	t.Run("non-member creator", func(t *testing.T) {
		nonMember := uuid.New()
		_, err := svc.CreateTopic(context.Background(), nonMember, CreateTopicInput{
			Name:     "Test",
			Icon:     "📌",
			ParentID: chat.ID,
		})
		require.Error(t, err)
	})
}

func TestTopicService_AddRemoveMember(t *testing.T) {
	svc, topicRepo, _, chatRepo, userRepo := setupTopicService()

	admin := uuid.New()
	memberA := uuid.New()
	memberB := uuid.New()
	userRepo.addUser(&model.User{ID: admin, Phone: "+628111", Name: "Admin"})
	userRepo.addUser(&model.User{ID: memberA, Phone: "+628222", Name: "MemberA"})
	userRepo.addUser(&model.User{ID: memberB, Phone: "+628333", Name: "MemberB"})

	// Create group and topic
	chat, _ := chatRepo.Create(context.Background(), model.CreateChatInput{
		Type:      model.ChatTypeGroup,
		Name:      "Tim",
		CreatedBy: admin,
	})
	_ = chatRepo.AddMember(context.Background(), chat.ID, admin, model.MemberRoleAdmin)
	_ = chatRepo.AddMember(context.Background(), chat.ID, memberA, model.MemberRoleMember)
	_ = chatRepo.AddMember(context.Background(), chat.ID, memberB, model.MemberRoleMember)

	topic, _ := svc.CreateTopic(context.Background(), admin, CreateTopicInput{
		Name:      "Test",
		Icon:      "📌",
		ParentID:  chat.ID,
		MemberIDs: []uuid.UUID{admin},
	})

	t.Run("add member", func(t *testing.T) {
		err := svc.AddMember(context.Background(), topic.ID, memberA, admin)
		require.NoError(t, err)

		members, _ := topicRepo.GetMembers(context.Background(), topic.ID)
		assert.Len(t, members, 2)
	})

	t.Run("add member not in parent", func(t *testing.T) {
		outsider := uuid.New()
		err := svc.AddMember(context.Background(), topic.ID, outsider, admin)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a member of the parent chat")
	})

	t.Run("non-admin cannot add", func(t *testing.T) {
		err := svc.AddMember(context.Background(), topic.ID, memberB, memberA)
		require.Error(t, err)
	})

	t.Run("remove member", func(t *testing.T) {
		err := svc.RemoveMember(context.Background(), topic.ID, memberA, admin)
		require.NoError(t, err)

		members, _ := topicRepo.GetMembers(context.Background(), topic.ID)
		assert.Len(t, members, 1)
	})
}

func TestTopicService_ListByChat(t *testing.T) {
	svc, _, _, chatRepo, userRepo := setupTopicService()

	user := uuid.New()
	userRepo.addUser(&model.User{ID: user, Phone: "+628111", Name: "User"})

	chat, _ := chatRepo.Create(context.Background(), model.CreateChatInput{
		Type:      model.ChatTypePersonal,
		CreatedBy: user,
	})
	_ = chatRepo.AddMember(context.Background(), chat.ID, user, model.MemberRoleAdmin)

	// Create 2 topics
	_, _ = svc.CreateTopic(context.Background(), user, CreateTopicInput{
		Name: "Topic A", Icon: "📌", ParentID: chat.ID,
	})
	_, _ = svc.CreateTopic(context.Background(), user, CreateTopicInput{
		Name: "Topic B", Icon: "📎", ParentID: chat.ID,
	})

	topics, err := svc.ListByChat(context.Background(), chat.ID, user)
	require.NoError(t, err)
	assert.Len(t, topics, 2)
}

func TestTopicService_DeleteTopic(t *testing.T) {
	svc, topicRepo, _, chatRepo, userRepo := setupTopicService()

	admin := uuid.New()
	member := uuid.New()
	userRepo.addUser(&model.User{ID: admin, Phone: "+628111", Name: "Admin"})
	userRepo.addUser(&model.User{ID: member, Phone: "+628222", Name: "Member"})

	chat, _ := chatRepo.Create(context.Background(), model.CreateChatInput{
		Type:      model.ChatTypeGroup,
		Name:      "Group",
		CreatedBy: admin,
	})
	_ = chatRepo.AddMember(context.Background(), chat.ID, admin, model.MemberRoleAdmin)
	_ = chatRepo.AddMember(context.Background(), chat.ID, member, model.MemberRoleMember)

	topic, _ := svc.CreateTopic(context.Background(), admin, CreateTopicInput{
		Name: "To Delete", Icon: "🗑️", ParentID: chat.ID,
	})

	t.Run("non-admin cannot delete", func(t *testing.T) {
		err := svc.DeleteTopic(context.Background(), topic.ID, member)
		require.Error(t, err)
	})

	t.Run("admin can delete", func(t *testing.T) {
		err := svc.DeleteTopic(context.Background(), topic.ID, admin)
		require.NoError(t, err)

		_, err = topicRepo.FindByID(context.Background(), topic.ID)
		require.Error(t, err)
		assert.True(t, apperror.IsNotFound(err))
	})
}

// --- Topic Message Service Tests ---

func TestTopicMessageService_SendMessage(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicMsgRepo := newMockTopicMsgRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewTopicMessageService(topicMsgRepo, topicRepo, hub)

	user := uuid.New()
	topicID := uuid.New()
	topicRepo.topics[topicID] = &model.Topic{ID: topicID, Name: "Test"}
	topicRepo.members[topicID] = []*model.TopicMember{
		{TopicID: topicID, UserID: user, Role: model.MemberRoleAdmin},
	}

	t.Run("success", func(t *testing.T) {
		msg, err := svc.SendMessage(context.Background(), SendTopicMessageInput{
			TopicID:  topicID,
			SenderID: user,
			Content:  "Hello topic",
		})
		require.NoError(t, err)
		assert.Equal(t, "Hello topic", msg.Content)
		assert.Equal(t, topicID, msg.TopicID)
	})

	t.Run("non-member cannot send", func(t *testing.T) {
		outsider := uuid.New()
		_, err := svc.SendMessage(context.Background(), SendTopicMessageInput{
			TopicID:  topicID,
			SenderID: outsider,
			Content:  "Hello",
		})
		require.Error(t, err)
	})

	t.Run("empty content", func(t *testing.T) {
		_, err := svc.SendMessage(context.Background(), SendTopicMessageInput{
			TopicID:  topicID,
			SenderID: user,
			Content:  "",
		})
		require.Error(t, err)
	})
}

func TestTopicMessageService_GetMessages(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicMsgRepo := newMockTopicMsgRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewTopicMessageService(topicMsgRepo, topicRepo, hub)

	user := uuid.New()
	topicID := uuid.New()
	topicRepo.topics[topicID] = &model.Topic{ID: topicID, Name: "Test"}
	topicRepo.members[topicID] = []*model.TopicMember{
		{TopicID: topicID, UserID: user, Role: model.MemberRoleAdmin},
	}

	// Send 3 messages
	for i := 0; i < 3; i++ {
		_, _ = svc.SendMessage(context.Background(), SendTopicMessageInput{
			TopicID:  topicID,
			SenderID: user,
			Content:  "msg",
		})
		time.Sleep(time.Millisecond) // ensure different timestamps
	}

	page, err := svc.GetMessages(context.Background(), topicID, "", 10)
	require.NoError(t, err)
	assert.Len(t, page.Messages, 3)
	assert.False(t, page.HasMore)
}

func TestTopicMessageService_DeleteMessage(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicMsgRepo := newMockTopicMsgRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewTopicMessageService(topicMsgRepo, topicRepo, hub)

	user := uuid.New()
	other := uuid.New()
	topicID := uuid.New()
	topicRepo.topics[topicID] = &model.Topic{ID: topicID, Name: "Test"}
	topicRepo.members[topicID] = []*model.TopicMember{
		{TopicID: topicID, UserID: user, Role: model.MemberRoleAdmin},
		{TopicID: topicID, UserID: other, Role: model.MemberRoleMember},
	}

	msg, _ := svc.SendMessage(context.Background(), SendTopicMessageInput{
		TopicID:  topicID,
		SenderID: user,
		Content:  "to delete",
	})

	t.Run("other user cannot delete", func(t *testing.T) {
		err := svc.DeleteMessage(context.Background(), msg.ID, other, true)
		require.Error(t, err)
	})

	t.Run("sender can delete", func(t *testing.T) {
		err := svc.DeleteMessage(context.Background(), msg.ID, user, true)
		require.NoError(t, err)
		assert.True(t, topicMsgRepo.messages[msg.ID].IsDeleted)
	})
}

func TestTopicService_GetTopic(t *testing.T) {
	svc, _, _, chatRepo, userRepo := setupTopicService()

	userA := uuid.New()
	userB := uuid.New()
	userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A"})
	userRepo.addUser(&model.User{ID: userB, Phone: "+628222", Name: "B"})

	// Create a parent chat
	chatID := uuid.New()
	chatRepo.chats[chatID] = &model.Chat{ID: chatID, Type: model.ChatTypeGroup, Name: "Group"}
	_ = chatRepo.AddMember(context.Background(), chatID, userA, model.MemberRoleAdmin)
	_ = chatRepo.AddMember(context.Background(), chatID, userB, model.MemberRoleMember)

	// Create topic
	topic, err := svc.CreateTopic(context.Background(), userA, CreateTopicInput{
		Name:     "Test Topic",
		Icon:     "\U0001F4AC",
		ParentID: chatID,
	})
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		detail, err := svc.GetTopic(context.Background(), topic.ID, userA)
		require.NoError(t, err)
		assert.Equal(t, topic.ID, detail.Topic.ID)
		assert.NotEmpty(t, detail.Members)
	})

	t.Run("non member forbidden", func(t *testing.T) {
		outsider := uuid.New()
		_, err := svc.GetTopic(context.Background(), topic.ID, outsider)
		assert.Error(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetTopic(context.Background(), uuid.New(), userA)
		assert.Error(t, err)
	})
}

func TestTopicService_ListByUser(t *testing.T) {
	svc, _, _, chatRepo, userRepo := setupTopicService()

	userA := uuid.New()
	userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A"})

	chatID := uuid.New()
	chatRepo.chats[chatID] = &model.Chat{ID: chatID, Type: model.ChatTypeGroup, Name: "Group"}
	_ = chatRepo.AddMember(context.Background(), chatID, userA, model.MemberRoleAdmin)

	_, err := svc.CreateTopic(context.Background(), userA, CreateTopicInput{
		Name:     "Topic 1",
		Icon:     "\U0001F4AC",
		ParentID: chatID,
	})
	require.NoError(t, err)

	items, err := svc.ListByUser(context.Background(), userA)
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestTopicService_UpdateTopic(t *testing.T) {
	svc, topicRepo, _, chatRepo, userRepo := setupTopicService()

	userA := uuid.New()
	userRepo.addUser(&model.User{ID: userA, Phone: "+628111", Name: "A"})

	chatID := uuid.New()
	chatRepo.chats[chatID] = &model.Chat{ID: chatID, Type: model.ChatTypeGroup, Name: "Group"}
	_ = chatRepo.AddMember(context.Background(), chatID, userA, model.MemberRoleAdmin)

	topic, err := svc.CreateTopic(context.Background(), userA, CreateTopicInput{
		Name:     "Old Name",
		Icon:     "\U0001F4AC",
		ParentID: chatID,
	})
	require.NoError(t, err)

	newName := "New Name"
	updated, err := svc.UpdateTopic(context.Background(), topic.ID, userA, UpdateTopicInput{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)

	// Verify in repo
	assert.Equal(t, "New Name", topicRepo.topics[topic.ID].Name)

	t.Run("non admin cannot update", func(t *testing.T) {
		outsider := uuid.New()
		_, err := svc.UpdateTopic(context.Background(), topic.ID, outsider, UpdateTopicInput{
			Name: &newName,
		})
		assert.Error(t, err)
	})
}

// --- Topic Message Error-Path Tests ---

func TestTopicMessageService_SendMessage_Errors(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicMsgRepo := newMockTopicMsgRepo()
	svc := NewTopicMessageService(topicMsgRepo, topicRepo, nil)

	topicID := uuid.New()
	user := uuid.New()
	topicRepo.topics[topicID] = &model.Topic{ID: topicID, Name: "T"}
	topicRepo.members[topicID] = []*model.TopicMember{
		{TopicID: topicID, UserID: user, Role: model.MemberRoleAdmin},
	}

	t.Run("default type is text", func(t *testing.T) {
		msg, err := svc.SendMessage(context.Background(), SendTopicMessageInput{
			TopicID: topicID, SenderID: user, Content: "Hi",
		})
		require.NoError(t, err)
		assert.Equal(t, model.MessageTypeText, msg.Type)
	})

	t.Run("reply to message in different topic", func(t *testing.T) {
		otherMsg := &model.TopicMessage{ID: uuid.New(), TopicID: uuid.New()}
		topicMsgRepo.messages[otherMsg.ID] = otherMsg
		replyID := otherMsg.ID
		_, err := svc.SendMessage(context.Background(), SendTopicMessageInput{
			TopicID: topicID, SenderID: user, Content: "Reply", Type: model.MessageTypeText,
			ReplyToID: &replyID,
		})
		require.Error(t, err)
	})

	t.Run("reply find generic error", func(t *testing.T) {
		topicMsgRepo.findErr = errors.New("db error")
		badID := uuid.New()
		_, err := svc.SendMessage(context.Background(), SendTopicMessageInput{
			TopicID: topicID, SenderID: user, Content: "Reply", Type: model.MessageTypeText,
			ReplyToID: &badID,
		})
		require.Error(t, err)
		topicMsgRepo.findErr = nil
	})

	t.Run("create error", func(t *testing.T) {
		topicMsgRepo.createErr = errors.New("db insert failed")
		_, err := svc.SendMessage(context.Background(), SendTopicMessageInput{
			TopicID: topicID, SenderID: user, Content: "Hello", Type: model.MessageTypeText,
		})
		require.Error(t, err)
		topicMsgRepo.createErr = nil
	})

	t.Run("get members error", func(t *testing.T) {
		topicRepo.getMembersErr = errors.New("db error")
		_, err := svc.SendMessage(context.Background(), SendTopicMessageInput{
			TopicID: topicID, SenderID: user, Content: "Hello", Type: model.MessageTypeText,
		})
		require.Error(t, err)
		topicRepo.getMembersErr = nil
	})
}

func TestTopicMessageService_GetMessages_Errors(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicMsgRepo := newMockTopicMsgRepo()
	svc := NewTopicMessageService(topicMsgRepo, topicRepo, nil)

	topicID := uuid.New()

	t.Run("invalid cursor", func(t *testing.T) {
		_, err := svc.GetMessages(context.Background(), topicID, "bad-cursor", 10)
		require.Error(t, err)
	})

	t.Run("list error", func(t *testing.T) {
		topicMsgRepo.listErr = errors.New("db error")
		_, err := svc.GetMessages(context.Background(), topicID, "", 10)
		require.Error(t, err)
		topicMsgRepo.listErr = nil
	})

	t.Run("default limit when zero", func(t *testing.T) {
		page, err := svc.GetMessages(context.Background(), topicID, "", 0)
		require.NoError(t, err)
		assert.NotNil(t, page)
	})

	t.Run("with valid cursor", func(t *testing.T) {
		cursor := time.Now().Add(time.Hour).Format(time.RFC3339Nano)
		page, err := svc.GetMessages(context.Background(), topicID, cursor, 10)
		require.NoError(t, err)
		assert.NotNil(t, page)
	})
}

func TestTopicMessageService_DeleteMessage_Errors(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicMsgRepo := newMockTopicMsgRepo()
	svc := NewTopicMessageService(topicMsgRepo, topicRepo, nil)

	topicID := uuid.New()
	sender := uuid.New()

	msg := &model.TopicMessage{ID: uuid.New(), TopicID: topicID, SenderID: sender}
	topicMsgRepo.messages[msg.ID] = msg

	t.Run("mark deleted error", func(t *testing.T) {
		topicMsgRepo.markDeletedErr = errors.New("db error")
		err := svc.DeleteMessage(context.Background(), msg.ID, sender, false)
		require.Error(t, err)
		topicMsgRepo.markDeletedErr = nil
	})

	t.Run("find error", func(t *testing.T) {
		topicMsgRepo.findErr = errors.New("db error")
		err := svc.DeleteMessage(context.Background(), msg.ID, sender, false)
		require.Error(t, err)
		topicMsgRepo.findErr = nil
	})

	t.Run("delete for all no hub", func(t *testing.T) {
		err := svc.DeleteMessage(context.Background(), msg.ID, sender, true)
		require.NoError(t, err)
	})
}

func TestTopicService_ListByChat_Errors(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicMsgRepo := newMockTopicMsgRepo()
	chatRepo := newMockChatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewTopicService(topicRepo, topicMsgRepo, chatRepo, userRepo, hub)

	t.Run("not a member of parent chat", func(t *testing.T) {
		chatID := uuid.New()
		creator := uuid.New()
		userRepo.addUser(&model.User{ID: creator, Phone: "+1", Name: "C"})

		chatRepo.chats[chatID] = &model.Chat{ID: chatID, Type: model.ChatTypeGroup}
		_ = chatRepo.AddMember(context.Background(), chatID, creator, model.MemberRoleAdmin)

		_, err := svc.ListByChat(context.Background(), chatID, uuid.New())
		require.Error(t, err)
	})

	t.Run("get parent members error", func(t *testing.T) {
		chatRepo.getMembersErr = errors.New("db error")
		_, err := svc.ListByChat(context.Background(), uuid.New(), uuid.New())
		require.Error(t, err)
		chatRepo.getMembersErr = nil
	})
}

func TestTopicService_GetTopic_Errors(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicMsgRepo := newMockTopicMsgRepo()
	chatRepo := newMockChatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewTopicService(topicRepo, topicMsgRepo, chatRepo, userRepo, hub)

	t.Run("topic not found", func(t *testing.T) {
		_, err := svc.GetTopic(context.Background(), uuid.New(), uuid.New())
		require.Error(t, err)
	})

	t.Run("not a member", func(t *testing.T) {
		creator := uuid.New()
		userRepo.addUser(&model.User{ID: creator, Phone: "+1", Name: "C"})

		chatID := uuid.New()
		chatRepo.chats[chatID] = &model.Chat{ID: chatID, Type: model.ChatTypeGroup}
		_ = chatRepo.AddMember(context.Background(), chatID, creator, model.MemberRoleAdmin)

		topic, err := svc.CreateTopic(context.Background(), creator, CreateTopicInput{
			Name:      "TestTopic",
			Icon:      "T",
			ParentID:  chatID,
			MemberIDs: []uuid.UUID{},
		})
		require.NoError(t, err)

		_, err = svc.GetTopic(context.Background(), topic.ID, uuid.New())
		require.Error(t, err)
	})

	t.Run("get members error", func(t *testing.T) {
		creator := uuid.New()
		userRepo.addUser(&model.User{ID: creator, Phone: "+2", Name: "C2"})

		chatID := uuid.New()
		chatRepo.chats[chatID] = &model.Chat{ID: chatID, Type: model.ChatTypeGroup}
		_ = chatRepo.AddMember(context.Background(), chatID, creator, model.MemberRoleAdmin)

		topic, err := svc.CreateTopic(context.Background(), creator, CreateTopicInput{
			Name:      "TestTopic2",
			Icon:      "T",
			ParentID:  chatID,
			MemberIDs: []uuid.UUID{},
		})
		require.NoError(t, err)

		topicRepo.getMembersErr = errors.New("db error")
		_, err = svc.GetTopic(context.Background(), topic.ID, creator)
		require.Error(t, err)
		topicRepo.getMembersErr = nil
	})
}

func TestTopicService_UpdateTopic_Errors(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicMsgRepo := newMockTopicMsgRepo()
	chatRepo := newMockChatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewTopicService(topicRepo, topicMsgRepo, chatRepo, userRepo, hub)

	t.Run("not admin", func(t *testing.T) {
		creator := uuid.New()
		other := uuid.New()
		userRepo.addUser(&model.User{ID: creator, Phone: "+1", Name: "C"})
		userRepo.addUser(&model.User{ID: other, Phone: "+2", Name: "O"})

		chatID := uuid.New()
		chatRepo.chats[chatID] = &model.Chat{ID: chatID, Type: model.ChatTypeGroup}
		_ = chatRepo.AddMember(context.Background(), chatID, creator, model.MemberRoleAdmin)
		_ = chatRepo.AddMember(context.Background(), chatID, other, model.MemberRoleMember)

		topic, err := svc.CreateTopic(context.Background(), creator, CreateTopicInput{
			Name:      "TopicForUpdate",
			Icon:      "U",
			ParentID:  chatID,
			MemberIDs: []uuid.UUID{other},
		})
		require.NoError(t, err)

		newName := "Updated"
		_, err = svc.UpdateTopic(context.Background(), topic.ID, other, UpdateTopicInput{Name: &newName})
		require.Error(t, err)
	})
}

func TestTopicService_AddMember_Errors(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicMsgRepo := newMockTopicMsgRepo()
	chatRepo := newMockChatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewTopicService(topicRepo, topicMsgRepo, chatRepo, userRepo, hub)

	creator := uuid.New()
	other := uuid.New()
	outsider := uuid.New()
	userRepo.addUser(&model.User{ID: creator, Phone: "+1", Name: "C"})
	userRepo.addUser(&model.User{ID: other, Phone: "+2", Name: "O"})
	userRepo.addUser(&model.User{ID: outsider, Phone: "+3", Name: "X"})

	chatID := uuid.New()
	chatRepo.chats[chatID] = &model.Chat{ID: chatID, Type: model.ChatTypeGroup}
	_ = chatRepo.AddMember(context.Background(), chatID, creator, model.MemberRoleAdmin)
	_ = chatRepo.AddMember(context.Background(), chatID, other, model.MemberRoleMember)

	topic, _ := svc.CreateTopic(context.Background(), creator, CreateTopicInput{
		Name:      "TopicForAdd",
		Icon:      "A",
		ParentID:  chatID,
		MemberIDs: []uuid.UUID{},
	})

	t.Run("user not in parent chat", func(t *testing.T) {
		err := svc.AddMember(context.Background(), topic.ID, outsider, creator)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parent chat")
	})
}

func TestTopicService_RemoveMember_Errors(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicMsgRepo := newMockTopicMsgRepo()
	chatRepo := newMockChatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewTopicService(topicRepo, topicMsgRepo, chatRepo, userRepo, hub)

	creator := uuid.New()
	userRepo.addUser(&model.User{ID: creator, Phone: "+1", Name: "C"})

	chatID := uuid.New()
	chatRepo.chats[chatID] = &model.Chat{ID: chatID, Type: model.ChatTypeGroup}
	_ = chatRepo.AddMember(context.Background(), chatID, creator, model.MemberRoleAdmin)

	topic, _ := svc.CreateTopic(context.Background(), creator, CreateTopicInput{
		Name:      "TopicForRemove",
		Icon:      "R",
		ParentID:  chatID,
		MemberIDs: []uuid.UUID{},
	})

	t.Run("remove self", func(t *testing.T) {
		err := svc.RemoveMember(context.Background(), topic.ID, creator, creator)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "leave")
	})
}

func TestTopicService_DeleteTopic_Errors(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicMsgRepo := newMockTopicMsgRepo()
	chatRepo := newMockChatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewTopicService(topicRepo, topicMsgRepo, chatRepo, userRepo, hub)

	t.Run("not admin", func(t *testing.T) {
		creator := uuid.New()
		other := uuid.New()
		userRepo.addUser(&model.User{ID: creator, Phone: "+1", Name: "C"})
		userRepo.addUser(&model.User{ID: other, Phone: "+2", Name: "O"})

		chatID := uuid.New()
		chatRepo.chats[chatID] = &model.Chat{ID: chatID, Type: model.ChatTypeGroup}
		_ = chatRepo.AddMember(context.Background(), chatID, creator, model.MemberRoleAdmin)
		_ = chatRepo.AddMember(context.Background(), chatID, other, model.MemberRoleMember)

		topic, err := svc.CreateTopic(context.Background(), creator, CreateTopicInput{
			Name:      "TopicForDelete",
			Icon:      "D",
			ParentID:  chatID,
			MemberIDs: []uuid.UUID{other},
		})
		require.NoError(t, err)

		err = svc.DeleteTopic(context.Background(), topic.ID, other)
		require.Error(t, err)
	})
}
