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
	"github.com/otoritech/chatat/internal/repository"
)

// -- Mock search repository --

type mockSearchRepo struct {
	messages  []*model.MessageSearchRow
	documents []*model.DocumentSearchRow
	contacts  []*model.User
	entities  []*model.Entity

	searchMsgErr  error
	searchDocErr  error
	searchConErr  error
	searchEntErr  error
}

func newMockSearchRepo() *mockSearchRepo {
	return &mockSearchRepo{
		messages:  []*model.MessageSearchRow{},
		documents: []*model.DocumentSearchRow{},
		contacts:  []*model.User{},
		entities:  []*model.Entity{},
	}
}

func (m *mockSearchRepo) SearchMessages(_ context.Context, _ uuid.UUID, _ string, offset, limit int) ([]*model.MessageSearchRow, error) {
	if m.searchMsgErr != nil {
		return nil, m.searchMsgErr
	}
	end := offset + limit
	if end > len(m.messages) {
		end = len(m.messages)
	}
	if offset >= len(m.messages) {
		return []*model.MessageSearchRow{}, nil
	}
	return m.messages[offset:end], nil
}

func (m *mockSearchRepo) SearchMessagesInChat(_ context.Context, chatID uuid.UUID, _ string, offset, limit int) ([]*model.MessageSearchRow, error) {
	var filtered []*model.MessageSearchRow
	for _, msg := range m.messages {
		if msg.ChatID == chatID {
			filtered = append(filtered, msg)
		}
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	if offset >= len(filtered) {
		return []*model.MessageSearchRow{}, nil
	}
	return filtered[offset:end], nil
}

func (m *mockSearchRepo) SearchDocuments(_ context.Context, _ uuid.UUID, _ string, offset, limit int) ([]*model.DocumentSearchRow, error) {
	if m.searchDocErr != nil {
		return nil, m.searchDocErr
	}
	end := offset + limit
	if end > len(m.documents) {
		end = len(m.documents)
	}
	if offset >= len(m.documents) {
		return []*model.DocumentSearchRow{}, nil
	}
	return m.documents[offset:end], nil
}

func (m *mockSearchRepo) SearchContacts(_ context.Context, _ uuid.UUID, _ string, limit int) ([]*model.User, error) {
	if m.searchConErr != nil {
		return nil, m.searchConErr
	}
	end := limit
	if end > len(m.contacts) {
		end = len(m.contacts)
	}
	return m.contacts[:end], nil
}

func (m *mockSearchRepo) SearchEntities(_ context.Context, _ uuid.UUID, _ string, limit int) ([]*model.Entity, error) {
	if m.searchEntErr != nil {
		return nil, m.searchEntErr
	}
	end := limit
	if end > len(m.entities) {
		end = len(m.entities)
	}
	return m.entities[:end], nil
}

// -- Tests --

func TestSearchService_SearchAll(t *testing.T) {
	searchRepo := newMockSearchRepo()
	chatRepo := newMockNotifChatRepo() // reuse from notification tests
	svc := NewSearchService(searchRepo, chatRepo)

	userID := uuid.New()
	chatID := uuid.New()

	// Seed data
	searchRepo.messages = []*model.MessageSearchRow{
		{ID: uuid.New(), ChatID: chatID, SenderID: uuid.New(), Content: "Halo semuanya", Type: model.MessageTypeText, CreatedAt: time.Now(), ChatName: "Keluarga", SenderName: "Ahmad", Highlight: "Halo <mark>semuanya</mark>"},
		{ID: uuid.New(), ChatID: chatID, SenderID: uuid.New(), Content: "Selamat pagi", Type: model.MessageTypeText, CreatedAt: time.Now(), ChatName: "Kantor", SenderName: "Budi", Highlight: "Selamat <mark>pagi</mark>"},
	}
	searchRepo.documents = []*model.DocumentSearchRow{
		{ID: uuid.New(), Title: "Notulen Rapat", Icon: "memo", OwnerID: userID, Locked: false, UpdatedAt: time.Now(), Highlight: "<mark>Notulen</mark> Rapat"},
	}
	searchRepo.contacts = []*model.User{
		{ID: uuid.New(), Name: "Ahmad", Phone: "+6281234567890"},
	}
	searchRepo.entities = []*model.Entity{
		{ID: uuid.New(), Name: "PT Contoh", Type: "organization", OwnerID: userID, Fields: map[string]string{}},
	}

	t.Run("returns mixed results limited to 3", func(t *testing.T) {
		results, err := svc.SearchAll(context.Background(), userID, "test", 3)
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.Len(t, results.Messages, 2) // only 2 available
		assert.Len(t, results.Documents, 1)
		assert.Len(t, results.Contacts, 1)
		assert.Len(t, results.Entities, 1)
	})

	t.Run("short query rejected", func(t *testing.T) {
		_, err := svc.SearchAll(context.Background(), userID, "a", 3)
		require.Error(t, err)
	})
}

func TestSearchService_SearchMessages(t *testing.T) {
	searchRepo := newMockSearchRepo()
	chatRepo := newMockNotifChatRepo()
	svc := NewSearchService(searchRepo, chatRepo)

	userID := uuid.New()

	// Seed 5 messages
	for i := 0; i < 5; i++ {
		searchRepo.messages = append(searchRepo.messages, &model.MessageSearchRow{
			ID: uuid.New(), ChatID: uuid.New(), SenderID: uuid.New(),
			Content: "Test message", Type: model.MessageTypeText,
			CreatedAt: time.Now(), ChatName: "Chat", SenderName: "User",
			Highlight: "Test <mark>message</mark>",
		})
	}

	t.Run("returns paginated results", func(t *testing.T) {
		results, err := svc.SearchMessages(context.Background(), userID, "test", SearchOpts{Offset: 0, Limit: 3})
		require.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("offset pagination", func(t *testing.T) {
		results, err := svc.SearchMessages(context.Background(), userID, "test", SearchOpts{Offset: 3, Limit: 10})
		require.NoError(t, err)
		assert.Len(t, results, 2) // 5 total, offset 3 = 2 remaining
	})

	t.Run("default limit applied", func(t *testing.T) {
		results, err := svc.SearchMessages(context.Background(), userID, "test", SearchOpts{})
		require.NoError(t, err)
		assert.Len(t, results, 5) // default limit 20, but only 5 available
	})

	t.Run("short query rejected", func(t *testing.T) {
		_, err := svc.SearchMessages(context.Background(), userID, "x", SearchOpts{})
		require.Error(t, err)
	})
}

func TestSearchService_SearchDocuments(t *testing.T) {
	searchRepo := newMockSearchRepo()
	chatRepo := newMockNotifChatRepo()
	svc := NewSearchService(searchRepo, chatRepo)

	userID := uuid.New()
	searchRepo.documents = []*model.DocumentSearchRow{
		{ID: uuid.New(), Title: "Notulen", Icon: "memo", OwnerID: userID, UpdatedAt: time.Now(), Highlight: "<mark>Notulen</mark>"},
	}

	t.Run("returns results", func(t *testing.T) {
		results, err := svc.SearchDocuments(context.Background(), userID, "notulen", SearchOpts{Limit: 20})
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}

func TestSearchService_SearchContacts(t *testing.T) {
	searchRepo := newMockSearchRepo()
	chatRepo := newMockNotifChatRepo()
	svc := NewSearchService(searchRepo, chatRepo)

	userID := uuid.New()
	searchRepo.contacts = []*model.User{
		{ID: uuid.New(), Name: "Ahmad", Phone: "+6281234567890"},
		{ID: uuid.New(), Name: "Budi", Phone: "+6281234567891"},
	}

	t.Run("returns contacts", func(t *testing.T) {
		results, err := svc.SearchContacts(context.Background(), userID, "ahmad")
		require.NoError(t, err)
		assert.Len(t, results, 2) // mock returns all regardless of query
	})
}

func TestSearchService_SearchEntities(t *testing.T) {
	searchRepo := newMockSearchRepo()
	chatRepo := newMockNotifChatRepo()
	svc := NewSearchService(searchRepo, chatRepo)

	userID := uuid.New()
	searchRepo.entities = []*model.Entity{
		{ID: uuid.New(), Name: "PT Contoh", Type: "organization", OwnerID: userID, Fields: map[string]string{}},
	}

	t.Run("returns entities", func(t *testing.T) {
		results, err := svc.SearchEntities(context.Background(), userID, "contoh")
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}

func TestSearchService_SearchInChat(t *testing.T) {
	searchRepo := newMockSearchRepo()
	chatRepo := newMockNotifChatRepo()
	svc := NewSearchService(searchRepo, chatRepo)

	userID := uuid.New()
	chatID := uuid.New()

	// Setup chat membership
	chatRepo.members[chatID] = []*model.ChatMember{
		{UserID: userID},
	}

	// Seed messages in different chats
	searchRepo.messages = []*model.MessageSearchRow{
		{ID: uuid.New(), ChatID: chatID, Content: "Hello", Type: model.MessageTypeText, CreatedAt: time.Now(), Highlight: "<mark>Hello</mark>"},
		{ID: uuid.New(), ChatID: chatID, Content: "World", Type: model.MessageTypeText, CreatedAt: time.Now(), Highlight: "<mark>World</mark>"},
		{ID: uuid.New(), ChatID: uuid.New(), Content: "Other chat", Type: model.MessageTypeText, CreatedAt: time.Now(), Highlight: "<mark>Other</mark>"},
	}

	t.Run("returns only chat messages", func(t *testing.T) {
		results, err := svc.SearchInChat(context.Background(), chatID, userID, "hello", SearchOpts{Limit: 20})
		require.NoError(t, err)
		assert.Len(t, results, 2) // only 2 messages in chatID
	})

	t.Run("non-member rejected", func(t *testing.T) {
		nonMember := uuid.New()
		_, err := svc.SearchInChat(context.Background(), chatID, nonMember, "hello", SearchOpts{Limit: 20})
		require.Error(t, err)
	})

	t.Run("short query rejected", func(t *testing.T) {
		_, err := svc.SearchInChat(context.Background(), chatID, userID, "h", SearchOpts{})
		require.Error(t, err)
	})
}

func TestBuildTSQuery(t *testing.T) {
	t.Run("single word", func(t *testing.T) {
		q := repository.BuildTSQuery("halo")
		assert.Equal(t, "'halo':*", q)
	})

	t.Run("multiple words", func(t *testing.T) {
		q := repository.BuildTSQuery("halo dunia")
		assert.Equal(t, "'halo':* & 'dunia':*", q)
	})

	t.Run("empty query", func(t *testing.T) {
		q := repository.BuildTSQuery("")
		assert.Equal(t, "", q)
	})

	t.Run("whitespace only", func(t *testing.T) {
		q := repository.BuildTSQuery("   ")
		assert.Equal(t, "", q)
	})
}

// --- Additional Coverage Tests ---

func TestSearchService_SearchDocuments_Errors(t *testing.T) {
	searchRepo := newMockSearchRepo()
	chatRepo := newMockNotifChatRepo()
	svc := NewSearchService(searchRepo, chatRepo)
	userID := uuid.New()

	searchRepo.documents = []*model.DocumentSearchRow{
		{ID: uuid.New(), Title: "Doc", Icon: "memo", OwnerID: userID, UpdatedAt: time.Now(), Highlight: "<mark>Doc</mark>"},
	}

	t.Run("short query rejected", func(t *testing.T) {
		_, err := svc.SearchDocuments(context.Background(), userID, "a", SearchOpts{Limit: 20})
		require.Error(t, err)
	})

	t.Run("default limit applied", func(t *testing.T) {
		results, err := svc.SearchDocuments(context.Background(), userID, "doc", SearchOpts{})
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}

func TestSearchService_SearchContacts_ShortQuery(t *testing.T) {
	searchRepo := newMockSearchRepo()
	chatRepo := newMockNotifChatRepo()
	svc := NewSearchService(searchRepo, chatRepo)

	_, err := svc.SearchContacts(context.Background(), uuid.New(), "x")
	require.Error(t, err)
}

func TestSearchService_SearchEntities_ShortQuery(t *testing.T) {
	searchRepo := newMockSearchRepo()
	chatRepo := newMockNotifChatRepo()
	svc := NewSearchService(searchRepo, chatRepo)

	_, err := svc.SearchEntities(context.Background(), uuid.New(), "x")
	require.Error(t, err)
}

func TestSearchService_SearchAll_LimitClamping(t *testing.T) {
	searchRepo := newMockSearchRepo()
	chatRepo := newMockNotifChatRepo()
	svc := NewSearchService(searchRepo, chatRepo)
	userID := uuid.New()

	t.Run("limit zero defaults to 3", func(t *testing.T) {
		results, err := svc.SearchAll(context.Background(), userID, "test", 0)
		require.NoError(t, err)
		assert.NotNil(t, results)
	})

	t.Run("limit over 5 defaults to 3", func(t *testing.T) {
		results, err := svc.SearchAll(context.Background(), userID, "test", 10)
		require.NoError(t, err)
		assert.NotNil(t, results)
	})
}

func TestSearchService_SearchAll_ErrorPaths(t *testing.T) {
	userID := uuid.New()

	t.Run("search messages error", func(t *testing.T) {
		searchRepo := newMockSearchRepo()
		chatRepo := newMockNotifChatRepo()
		svc := NewSearchService(searchRepo, chatRepo)
		searchRepo.searchMsgErr = errors.New("db error")
		_, err := svc.SearchAll(context.Background(), userID, "test", 3)
		require.Error(t, err)
	})

	t.Run("search documents error", func(t *testing.T) {
		searchRepo := newMockSearchRepo()
		chatRepo := newMockNotifChatRepo()
		svc := NewSearchService(searchRepo, chatRepo)
		searchRepo.searchDocErr = errors.New("db error")
		_, err := svc.SearchAll(context.Background(), userID, "test", 3)
		require.Error(t, err)
	})

	t.Run("search contacts error", func(t *testing.T) {
		searchRepo := newMockSearchRepo()
		chatRepo := newMockNotifChatRepo()
		svc := NewSearchService(searchRepo, chatRepo)
		searchRepo.searchConErr = errors.New("db error")
		_, err := svc.SearchAll(context.Background(), userID, "test", 3)
		require.Error(t, err)
	})

	t.Run("search entities error", func(t *testing.T) {
		searchRepo := newMockSearchRepo()
		chatRepo := newMockNotifChatRepo()
		svc := NewSearchService(searchRepo, chatRepo)
		searchRepo.searchEntErr = errors.New("db error")
		_, err := svc.SearchAll(context.Background(), userID, "test", 3)
		require.Error(t, err)
	})
}

func TestSearchService_SearchInChat_DefaultLimit(t *testing.T) {
	searchRepo := newMockSearchRepo()
	chatRepo := newMockNotifChatRepo()
	svc := NewSearchService(searchRepo, chatRepo)

	userID := uuid.New()
	chatID := uuid.New()
	chatRepo.members[chatID] = []*model.ChatMember{{UserID: userID}}

	results, err := svc.SearchInChat(context.Background(), chatID, userID, "hello", SearchOpts{})
	require.NoError(t, err)
	assert.NotNil(t, results)
}
