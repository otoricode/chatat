package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/model"
)

// -- Mock push sender --

type mockPushSender struct {
	sentSingle []struct {
		Token string
		Notif model.Notification
	}
	sentMulti []struct {
		Tokens []string
		Notif  model.Notification
	}
	shouldError bool
}

func newMockPushSender() *mockPushSender {
	return &mockPushSender{}
}

func (m *mockPushSender) Send(_ context.Context, token string, notif model.Notification) error {
	if m.shouldError {
		return assert.AnError
	}
	m.sentSingle = append(m.sentSingle, struct {
		Token string
		Notif model.Notification
	}{Token: token, Notif: notif})
	return nil
}

func (m *mockPushSender) SendMulti(_ context.Context, tokens []string, notif model.Notification) error {
	if m.shouldError {
		return assert.AnError
	}
	m.sentMulti = append(m.sentMulti, struct {
		Tokens []string
		Notif  model.Notification
	}{Tokens: tokens, Notif: notif})
	return nil
}

// -- Mock device token repo --

type mockDeviceTokenRepo struct {
	tokens      map[uuid.UUID][]*model.DeviceToken
	findErr     error
	findUsrsErr error
}

func newMockDeviceTokenRepo() *mockDeviceTokenRepo {
	return &mockDeviceTokenRepo{tokens: make(map[uuid.UUID][]*model.DeviceToken)}
}

func (m *mockDeviceTokenRepo) Upsert(_ context.Context, userID uuid.UUID, token, platform string) (*model.DeviceToken, error) {
	// Remove existing with same token for this user
	existing := m.tokens[userID]
	for i, t := range existing {
		if t.Token == token {
			existing[i].Platform = platform
			return existing[i], nil
		}
	}
	dt := &model.DeviceToken{
		ID:       uuid.New(),
		UserID:   userID,
		Token:    token,
		Platform: platform,
	}
	m.tokens[userID] = append(m.tokens[userID], dt)
	return dt, nil
}

func (m *mockDeviceTokenRepo) Delete(_ context.Context, userID uuid.UUID, token string) error {
	existing := m.tokens[userID]
	for i, t := range existing {
		if t.Token == token {
			m.tokens[userID] = append(existing[:i], existing[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockDeviceTokenRepo) DeleteByToken(_ context.Context, token string) error {
	for uid, tokens := range m.tokens {
		for i, t := range tokens {
			if t.Token == token {
				m.tokens[uid] = append(tokens[:i], tokens[i+1:]...)
				return nil
			}
		}
	}
	return nil
}

func (m *mockDeviceTokenRepo) FindByUser(_ context.Context, userID uuid.UUID) ([]*model.DeviceToken, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	tokens := m.tokens[userID]
	if tokens == nil {
		return []*model.DeviceToken{}, nil
	}
	return tokens, nil
}

func (m *mockDeviceTokenRepo) FindByUsers(_ context.Context, userIDs []uuid.UUID) ([]*model.DeviceToken, error) {
	if m.findUsrsErr != nil {
		return nil, m.findUsrsErr
	}
	var result []*model.DeviceToken
	for _, uid := range userIDs {
		result = append(result, m.tokens[uid]...)
	}
	if result == nil {
		result = []*model.DeviceToken{}
	}
	return result, nil
}

func (m *mockDeviceTokenRepo) DeleteStale(_ context.Context, _ int) (int64, error) {
	return 0, nil
}

// -- Mock chat repo for notification tests --

type mockNotifChatRepo struct {
	members    map[uuid.UUID][]*model.ChatMember
	getMembErr error
}

func newMockNotifChatRepo() *mockNotifChatRepo {
	return &mockNotifChatRepo{members: make(map[uuid.UUID][]*model.ChatMember)}
}

func (m *mockNotifChatRepo) GetMembers(_ context.Context, chatID uuid.UUID) ([]*model.ChatMember, error) {
	if m.getMembErr != nil {
		return nil, m.getMembErr
	}
	return m.members[chatID], nil
}

// Implement remaining ChatRepository methods as stubs
func (m *mockNotifChatRepo) Create(_ context.Context, _ model.CreateChatInput) (*model.Chat, error) {
	return nil, nil
}
func (m *mockNotifChatRepo) FindByID(_ context.Context, _ uuid.UUID) (*model.Chat, error) {
	return nil, nil
}
func (m *mockNotifChatRepo) FindPersonalChat(_ context.Context, _, _ uuid.UUID) (*model.Chat, error) {
	return nil, nil
}
func (m *mockNotifChatRepo) ListByUser(_ context.Context, _ uuid.UUID) ([]*model.ChatWithLastMessage, error) {
	return nil, nil
}
func (m *mockNotifChatRepo) AddMember(_ context.Context, _, _ uuid.UUID, _ model.MemberRole) error {
	return nil
}
func (m *mockNotifChatRepo) RemoveMember(_ context.Context, _, _ uuid.UUID) error { return nil }
func (m *mockNotifChatRepo) Update(_ context.Context, _ uuid.UUID, _ model.UpdateChatInput) (*model.Chat, error) {
	return nil, nil
}
func (m *mockNotifChatRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockNotifChatRepo) Pin(_ context.Context, _ uuid.UUID) error    { return nil }
func (m *mockNotifChatRepo) Unpin(_ context.Context, _ uuid.UUID) error  { return nil }

// -- Tests --

func TestNotificationService_RegisterDevice(t *testing.T) {
	ctx := context.Background()
	deviceRepo := newMockDeviceTokenRepo()
	chatRepo := newMockNotifChatRepo()
	sender := newMockPushSender()
	svc := NewNotificationService(deviceRepo, chatRepo, sender)

	userID := uuid.New()

	t.Run("success ios", func(t *testing.T) {
		err := svc.RegisterDevice(ctx, userID, "token-ios-123", "ios")
		require.NoError(t, err)
		tokens, _ := deviceRepo.FindByUser(ctx, userID)
		assert.Len(t, tokens, 1)
		assert.Equal(t, "ios", tokens[0].Platform)
	})

	t.Run("success android", func(t *testing.T) {
		err := svc.RegisterDevice(ctx, userID, "token-android-456", "android")
		require.NoError(t, err)
		tokens, _ := deviceRepo.FindByUser(ctx, userID)
		assert.Len(t, tokens, 2)
	})

	t.Run("empty token", func(t *testing.T) {
		err := svc.RegisterDevice(ctx, userID, "", "ios")
		require.Error(t, err)
	})

	t.Run("invalid platform", func(t *testing.T) {
		err := svc.RegisterDevice(ctx, userID, "token-web", "web")
		require.Error(t, err)
	})

	t.Run("upsert same token updates platform", func(t *testing.T) {
		err := svc.RegisterDevice(ctx, userID, "token-ios-123", "android")
		require.NoError(t, err)
		tokens, _ := deviceRepo.FindByUser(ctx, userID)
		// Should still have 2 tokens, first one updated to android
		found := false
		for _, tok := range tokens {
			if tok.Token == "token-ios-123" {
				assert.Equal(t, "android", tok.Platform)
				found = true
			}
		}
		assert.True(t, found)
	})
}

func TestNotificationService_UnregisterDevice(t *testing.T) {
	ctx := context.Background()
	deviceRepo := newMockDeviceTokenRepo()
	chatRepo := newMockNotifChatRepo()
	sender := newMockPushSender()
	svc := NewNotificationService(deviceRepo, chatRepo, sender)

	userID := uuid.New()
	_ = svc.RegisterDevice(ctx, userID, "token-to-remove", "ios")

	t.Run("success", func(t *testing.T) {
		err := svc.UnregisterDevice(ctx, userID, "token-to-remove")
		require.NoError(t, err)
		tokens, _ := deviceRepo.FindByUser(ctx, userID)
		assert.Len(t, tokens, 0)
	})
}

func TestNotificationService_SendToUser(t *testing.T) {
	ctx := context.Background()
	deviceRepo := newMockDeviceTokenRepo()
	chatRepo := newMockNotifChatRepo()
	sender := newMockPushSender()
	svc := NewNotificationService(deviceRepo, chatRepo, sender)

	userID := uuid.New()
	_ = svc.RegisterDevice(ctx, userID, "token-1", "ios")
	_ = svc.RegisterDevice(ctx, userID, "token-2", "android")

	notif := model.Notification{
		Type:  model.NotifTypeMessage,
		Title: "Ahmad",
		Body:  "Ahmad: Halo, apa kabar",
	}

	t.Run("sends to all user tokens", func(t *testing.T) {
		err := svc.SendToUser(ctx, userID, notif)
		require.NoError(t, err)
		assert.Len(t, sender.sentMulti, 1)
		assert.Len(t, sender.sentMulti[0].Tokens, 2)
	})

	t.Run("no tokens - no error", func(t *testing.T) {
		noTokenUser := uuid.New()
		err := svc.SendToUser(ctx, noTokenUser, notif)
		require.NoError(t, err)
	})
}

func TestNotificationService_SendToChat(t *testing.T) {
	ctx := context.Background()
	deviceRepo := newMockDeviceTokenRepo()
	chatRepo := newMockNotifChatRepo()
	sender := newMockPushSender()
	svc := NewNotificationService(deviceRepo, chatRepo, sender)

	senderID := uuid.New()
	member1 := uuid.New()
	member2 := uuid.New()
	chatID := uuid.New()

	// Register devices for members
	_ = svc.RegisterDevice(ctx, member1, "member1-token", "ios")
	_ = svc.RegisterDevice(ctx, member2, "member2-token", "android")
	_ = svc.RegisterDevice(ctx, senderID, "sender-token", "ios")

	// Add members to chat
	chatRepo.members[chatID] = []*model.ChatMember{
		{UserID: senderID},
		{UserID: member1},
		{UserID: member2},
	}

	notif := model.Notification{
		Type:  model.NotifTypeMessage,
		Title: "Test",
		Body:  "Hello",
	}

	t.Run("excludes sender", func(t *testing.T) {
		err := svc.SendToChat(ctx, chatID, senderID, notif)
		require.NoError(t, err)
		require.Len(t, sender.sentMulti, 1)
		// Should have tokens for member1 and member2 only (2 tokens)
		assert.Len(t, sender.sentMulti[0].Tokens, 2)
		for _, token := range sender.sentMulti[0].Tokens {
			assert.NotEqual(t, "sender-token", token)
		}
	})
}

func TestNotificationService_SendToUsers(t *testing.T) {
	ctx := context.Background()
	deviceRepo := newMockDeviceTokenRepo()
	chatRepo := newMockNotifChatRepo()
	sender := newMockPushSender()
	svc := NewNotificationService(deviceRepo, chatRepo, sender)

	user1 := uuid.New()
	user2 := uuid.New()
	_ = svc.RegisterDevice(ctx, user1, "u1-token", "ios")
	_ = svc.RegisterDevice(ctx, user2, "u2-token", "android")

	notif := model.Notification{
		Type:  model.NotifTypeGroupInvite,
		Title: "Undangan",
		Body:  "Anda diundang ke grup",
	}

	t.Run("sends to multiple users", func(t *testing.T) {
		err := svc.SendToUsers(ctx, []uuid.UUID{user1, user2}, notif)
		require.NoError(t, err)
		assert.Len(t, sender.sentMulti, 1)
		assert.Len(t, sender.sentMulti[0].Tokens, 2)
	})

	t.Run("empty user list - no error", func(t *testing.T) {
		err := svc.SendToUsers(ctx, []uuid.UUID{}, notif)
		require.NoError(t, err)
	})
}

func TestBuildNotifications(t *testing.T) {
	chatID := uuid.New()
	topicID := uuid.New()
	docID := uuid.New()

	t.Run("message notif", func(t *testing.T) {
		n := BuildMessageNotif("Ahmad", "Halo, apa kabar? Saya ingin menanyakan tentang proyek baru", chatID, "personal")
		assert.Equal(t, model.NotifTypeMessage, n.Type)
		assert.Equal(t, "Ahmad", n.Title)
		assert.Contains(t, n.Body, "Ahmad:")
		assert.Equal(t, chatID.String(), n.Data["chatId"])
		assert.Equal(t, "high", n.Priority)
		// Body should be truncated
		assert.LessOrEqual(t, len([]rune(n.Body)), 70) // name + ": " + 50 chars + "..."
	})

	t.Run("group message notif", func(t *testing.T) {
		n := BuildGroupMessageNotif("Keluarga", "Ahmad", "Besok kumpul ya", chatID)
		assert.Equal(t, model.NotifTypeGroupMessage, n.Type)
		assert.Equal(t, "Keluarga", n.Title)
		assert.Contains(t, n.Body, "Ahmad:")
		assert.Equal(t, chatID.String(), n.Data["chatId"])
	})

	t.Run("topic message notif", func(t *testing.T) {
		n := BuildTopicMessageNotif("Keuangan", "Budi", "Sudah bayar", topicID)
		assert.Equal(t, model.NotifTypeTopicMessage, n.Type)
		assert.Equal(t, "Keuangan", n.Title)
		assert.Contains(t, n.Body, "Budi:")
		assert.Equal(t, topicID.String(), n.Data["topicId"])
	})

	t.Run("signature request notif", func(t *testing.T) {
		n := BuildSignatureRequestNotif("Ahmad", "Notulen Rapat", docID)
		assert.Equal(t, model.NotifTypeSignatureRequest, n.Type)
		assert.Contains(t, n.Body, "Ahmad")
		assert.Contains(t, n.Body, "Notulen Rapat")
		assert.Equal(t, docID.String(), n.Data["documentId"])
	})

	t.Run("document locked notif", func(t *testing.T) {
		n := BuildDocLockedNotif("Ahmad", "Notulen", docID)
		assert.Equal(t, model.NotifTypeDocumentLocked, n.Type)
		assert.Contains(t, n.Body, "Notulen")
		assert.Contains(t, n.Body, "Ahmad")
		assert.Equal(t, docID.String(), n.Data["documentId"])
	})

	t.Run("group invite notif", func(t *testing.T) {
		n := BuildGroupInviteNotif("Ahmad", "Keluarga Besar", chatID)
		assert.Equal(t, model.NotifTypeGroupInvite, n.Type)
		assert.Contains(t, n.Body, "Ahmad")
		assert.Contains(t, n.Body, "Keluarga Besar")
		assert.Equal(t, chatID.String(), n.Data["chatId"])
	})

	t.Run("truncate long text", func(t *testing.T) {
		longMsg := "Ini adalah pesan yang sangat panjang sekali yang melebihi lima puluh karakter dan harus dipotong"
		n := BuildMessageNotif("X", longMsg, chatID, "personal")
		// Should have "..." at the end
		assert.Contains(t, n.Body, "...")
	})
}

func TestBuildMessageNotif_ChatType(t *testing.T) {
	chatID := uuid.New()

	t.Run("personal chat type", func(t *testing.T) {
		n := BuildMessageNotif("Ahmad", "Halo", chatID, "personal")
		assert.Equal(t, model.NotifTypeMessage, n.Type)
	})

	t.Run("group chat type", func(t *testing.T) {
		n := BuildMessageNotif("Ahmad", "Halo", chatID, "group")
		assert.Equal(t, model.NotifTypeGroupMessage, n.Type)
	})
}

// --- Error-Path Tests ---

func TestNotificationService_SendToUser_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("find tokens error", func(t *testing.T) {
		deviceRepo := newMockDeviceTokenRepo()
		deviceRepo.findErr = errors.New("db error")
		svc := NewNotificationService(deviceRepo, newMockNotifChatRepo(), newMockPushSender())
		err := svc.SendToUser(ctx, uuid.New(), model.Notification{Title: "T"})
		require.Error(t, err)
	})

	t.Run("send multi error", func(t *testing.T) {
		deviceRepo := newMockDeviceTokenRepo()
		sender := newMockPushSender()
		sender.shouldError = true
		svc := NewNotificationService(deviceRepo, newMockNotifChatRepo(), sender)
		userID := uuid.New()
		_ = svc.RegisterDevice(ctx, userID, "tok1", "ios")
		err := svc.SendToUser(ctx, userID, model.Notification{Title: "T"})
		require.Error(t, err)
	})
}

func TestNotificationService_SendToUsers_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("find tokens error", func(t *testing.T) {
		deviceRepo := newMockDeviceTokenRepo()
		deviceRepo.findUsrsErr = errors.New("db error")
		svc := NewNotificationService(deviceRepo, newMockNotifChatRepo(), newMockPushSender())
		err := svc.SendToUsers(ctx, []uuid.UUID{uuid.New()}, model.Notification{Title: "T"})
		require.Error(t, err)
	})

	t.Run("send multi error", func(t *testing.T) {
		deviceRepo := newMockDeviceTokenRepo()
		sender := newMockPushSender()
		sender.shouldError = true
		svc := NewNotificationService(deviceRepo, newMockNotifChatRepo(), sender)
		userID := uuid.New()
		_ = svc.RegisterDevice(ctx, userID, "tok1", "ios")
		err := svc.SendToUsers(ctx, []uuid.UUID{userID}, model.Notification{Title: "T"})
		require.Error(t, err)
	})

	t.Run("no tokens found for users", func(t *testing.T) {
		deviceRepo := newMockDeviceTokenRepo()
		svc := NewNotificationService(deviceRepo, newMockNotifChatRepo(), newMockPushSender())
		err := svc.SendToUsers(ctx, []uuid.UUID{uuid.New()}, model.Notification{Title: "T"})
		require.NoError(t, err) // empty tokens = no error
	})
}

func TestNotificationService_SendToChat_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("get members error", func(t *testing.T) {
		chatRepo := newMockNotifChatRepo()
		chatRepo.getMembErr = errors.New("db error")
		svc := NewNotificationService(newMockDeviceTokenRepo(), chatRepo, newMockPushSender())
		err := svc.SendToChat(ctx, uuid.New(), uuid.New(), model.Notification{Title: "T"})
		require.Error(t, err)
	})

	t.Run("all members excluded", func(t *testing.T) {
		chatRepo := newMockNotifChatRepo()
		senderID := uuid.New()
		chatID := uuid.New()
		chatRepo.members[chatID] = []*model.ChatMember{{UserID: senderID}}
		svc := NewNotificationService(newMockDeviceTokenRepo(), chatRepo, newMockPushSender())
		err := svc.SendToChat(ctx, chatID, senderID, model.Notification{Title: "T"})
		require.NoError(t, err) // empty userIDs = no error
	})
}
