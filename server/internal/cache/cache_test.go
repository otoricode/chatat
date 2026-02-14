package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/model"
)

func setupTestCache(t *testing.T) (*Service, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svc := NewService(client)
	return svc, mr
}

// --- User Cache Tests ---

func TestService_UserCache(t *testing.T) {
	svc, _ := setupTestCache(t)
	ctx := context.Background()
	userID := uuid.New()

	t.Run("miss returns nil", func(t *testing.T) {
		user, err := svc.GetUser(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, user)
	})

	t.Run("set and get", func(t *testing.T) {
		u := &model.User{
			ID:    userID,
			Phone: "+6281234567890",
			Name:  "Test User",
		}
		svc.SetUser(ctx, u)

		got, err := svc.GetUser(ctx, userID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, userID, got.ID)
		assert.Equal(t, "+6281234567890", got.Phone)
		assert.Equal(t, "Test User", got.Name)
	})

	t.Run("invalidate removes from cache", func(t *testing.T) {
		u := &model.User{ID: userID, Phone: "+6281234567890", Name: "Test"}
		svc.SetUser(ctx, u)
		svc.InvalidateUser(ctx, userID)

		got, err := svc.GetUser(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestService_UserCache_TTL(t *testing.T) {
	svc, mr := setupTestCache(t)
	ctx := context.Background()
	userID := uuid.New()
	u := &model.User{ID: userID, Phone: "+6281234567890", Name: "TTL User"}

	svc.SetUser(ctx, u)

	got, _ := svc.GetUser(ctx, userID)
	require.NotNil(t, got)

	mr.FastForward(UserTTL + time.Second)

	got, _ = svc.GetUser(ctx, userID)
	assert.Nil(t, got)
}

// --- Online Status Tests ---

func TestService_OnlineStatus(t *testing.T) {
	svc, _ := setupTestCache(t)
	ctx := context.Background()
	userID := uuid.New()

	t.Run("not online by default", func(t *testing.T) {
		assert.False(t, svc.IsOnline(ctx, userID))
	})

	t.Run("set online", func(t *testing.T) {
		svc.SetOnline(ctx, userID)
		assert.True(t, svc.IsOnline(ctx, userID))
	})

	t.Run("set offline", func(t *testing.T) {
		svc.SetOnline(ctx, userID)
		svc.SetOffline(ctx, userID)
		assert.False(t, svc.IsOnline(ctx, userID))
	})
}

func TestService_OnlineStatus_TTL(t *testing.T) {
	svc, mr := setupTestCache(t)
	ctx := context.Background()
	userID := uuid.New()

	svc.SetOnline(ctx, userID)
	assert.True(t, svc.IsOnline(ctx, userID))

	mr.FastForward(OnlineUserTTL + time.Second)
	assert.False(t, svc.IsOnline(ctx, userID))
}

// --- Chat List Cache Tests ---

func TestService_ChatListCache(t *testing.T) {
	svc, _ := setupTestCache(t)
	ctx := context.Background()
	userID := uuid.New()

	t.Run("miss returns nil", func(t *testing.T) {
		data, err := svc.GetChatList(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, data)
	})

	t.Run("set and get", func(t *testing.T) {
		payload := []byte(`[{"id":"chat-1"}]`)
		svc.SetChatList(ctx, userID, payload)

		got, err := svc.GetChatList(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, payload, got)
	})

	t.Run("invalidate removes from cache", func(t *testing.T) {
		svc.SetChatList(ctx, userID, []byte(`[{"id":"chat-1"}]`))
		svc.InvalidateChatList(ctx, userID)

		got, err := svc.GetChatList(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestService_ChatListCache_TTL(t *testing.T) {
	svc, mr := setupTestCache(t)
	ctx := context.Background()
	userID := uuid.New()

	svc.SetChatList(ctx, userID, []byte(`[{"id":"chat-1"}]`))

	got, _ := svc.GetChatList(ctx, userID)
	require.NotNil(t, got)

	mr.FastForward(ChatListTTL + time.Second)

	got, _ = svc.GetChatList(ctx, userID)
	assert.Nil(t, got)
}

// --- Generic Helpers Tests ---

func TestService_GenericHelpers(t *testing.T) {
	svc, _ := setupTestCache(t)
	ctx := context.Background()

	t.Run("get miss returns nil", func(t *testing.T) {
		data, err := svc.Get(ctx, "nonexistent")
		assert.NoError(t, err)
		assert.Nil(t, data)
	})

	t.Run("set and get", func(t *testing.T) {
		err := svc.Set(ctx, "test:key", []byte("hello"), time.Minute)
		require.NoError(t, err)

		data, err := svc.Get(ctx, "test:key")
		require.NoError(t, err)
		assert.Equal(t, []byte("hello"), data)
	})

	t.Run("del removes key", func(t *testing.T) {
		err := svc.Set(ctx, "test:del", []byte("bye"), time.Minute)
		require.NoError(t, err)

		err = svc.Del(ctx, "test:del")
		require.NoError(t, err)

		data, err := svc.Get(ctx, "test:del")
		assert.NoError(t, err)
		assert.Nil(t, data)
	})
}

// --- Key Format Tests ---

func TestKeyFormats(t *testing.T) {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	assert.Equal(t, "cache:user:550e8400-e29b-41d4-a716-446655440000", userKey(id))
	assert.Equal(t, "cache:online:550e8400-e29b-41d4-a716-446655440000", onlineKey(id))
	assert.Equal(t, "cache:chatlist:550e8400-e29b-41d4-a716-446655440000", chatListKey(id))
}

// --- Unmarshal Error Test ---

func TestService_GetUser_InvalidJSON(t *testing.T) {
	svc, mr := setupTestCache(t)
	ctx := context.Background()
	userID := uuid.New()

	// Manually set invalid JSON in redis
	require.NoError(t, mr.Set(userKey(userID), "not-valid-json"))

	user, err := svc.GetUser(ctx, userID)
	assert.Nil(t, user)
	assert.NoError(t, err) // returns nil,nil for unmarshal error
}
