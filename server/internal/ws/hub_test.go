package ws_test

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/otoritech/chatat/internal/ws"
)

// mockConn is not needed since we test Hub without actual websocket connections.
// We create clients with nil Conn for hub-only logic tests.

func startHub(t *testing.T) *ws.Hub {
	t.Helper()
	hub := ws.NewHub()
	go hub.Run()
	t.Cleanup(func() { hub.Shutdown() })
	return hub
}

func TestHub_RegisterAndIsOnline(t *testing.T) {
	hub := startHub(t)
	userID := uuid.New()

	client := &ws.Client{
		UserID: userID,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}

	hub.RegisterClient(client)
	time.Sleep(10 * time.Millisecond)

	assert.True(t, hub.IsOnline(userID))
	assert.False(t, hub.IsOnline(uuid.New()))
}

func TestHub_UnregisterClient(t *testing.T) {
	hub := startHub(t)
	userID := uuid.New()

	client := &ws.Client{
		UserID: userID,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}

	hub.RegisterClient(client)
	time.Sleep(10 * time.Millisecond)
	assert.True(t, hub.IsOnline(userID))

	hub.UnregisterClient(client)
	time.Sleep(10 * time.Millisecond)
	assert.False(t, hub.IsOnline(userID))
}

func TestHub_JoinAndLeaveRoom(t *testing.T) {
	hub := startHub(t)
	userID := uuid.New()

	client := &ws.Client{
		UserID: userID,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}

	hub.RegisterClient(client)
	time.Sleep(10 * time.Millisecond)

	hub.JoinRoom(client, "chat:123")
	members := hub.GetRoomMembers("chat:123")
	assert.Len(t, members, 1)
	assert.Equal(t, userID, members[0])

	hub.LeaveRoom(client, "chat:123")
	members = hub.GetRoomMembers("chat:123")
	assert.Nil(t, members)
}

func TestHub_SendToRoom(t *testing.T) {
	hub := startHub(t)

	user1 := uuid.New()
	user2 := uuid.New()

	c1 := &ws.Client{UserID: user1, Send: make(chan []byte, 256), Hub: hub}
	c2 := &ws.Client{UserID: user2, Send: make(chan []byte, 256), Hub: hub}

	hub.RegisterClient(c1)
	hub.RegisterClient(c2)
	time.Sleep(10 * time.Millisecond)

	hub.JoinRoom(c1, "chat:abc")
	hub.JoinRoom(c2, "chat:abc")

	// Send to room excluding user1
	hub.SendToRoom("chat:abc", []byte("hello"), user1)
	time.Sleep(10 * time.Millisecond)

	// user2 should receive
	select {
	case msg := <-c2.Send:
		assert.Equal(t, "hello", string(msg))
	case <-time.After(100 * time.Millisecond):
		t.Fatal("user2 should have received message")
	}

	// user1 should NOT receive (excluded)
	select {
	case <-c1.Send:
		t.Fatal("user1 should not have received message")
	case <-time.After(50 * time.Millisecond):
		// expected
	}
}

func TestHub_SendToUser(t *testing.T) {
	hub := startHub(t)
	userID := uuid.New()

	client := &ws.Client{
		UserID: userID,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}

	hub.RegisterClient(client)
	time.Sleep(10 * time.Millisecond)

	hub.SendToUser(userID, []byte("direct"))

	select {
	case msg := <-client.Send:
		assert.Equal(t, "direct", string(msg))
	case <-time.After(100 * time.Millisecond):
		t.Fatal("should have received direct message")
	}
}

func TestHub_GetOnlineUsers(t *testing.T) {
	hub := startHub(t)

	user1 := uuid.New()
	user2 := uuid.New()
	user3 := uuid.New()

	c1 := &ws.Client{UserID: user1, Send: make(chan []byte, 256), Hub: hub}
	c2 := &ws.Client{UserID: user2, Send: make(chan []byte, 256), Hub: hub}

	hub.RegisterClient(c1)
	hub.RegisterClient(c2)
	time.Sleep(10 * time.Millisecond)

	online := hub.GetOnlineUsers([]uuid.UUID{user1, user2, user3})
	assert.Len(t, online, 2)
}

func TestHub_SetEventCallbacks(t *testing.T) {
	hub := ws.NewHub()

	var connectedUser, disconnectedUser uuid.UUID
	var mu sync.Mutex
	hub.SetEventCallbacks(
		func(id uuid.UUID) {
			mu.Lock()
			connectedUser = id
			mu.Unlock()
		},
		func(id uuid.UUID) {
			mu.Lock()
			disconnectedUser = id
			mu.Unlock()
		},
	)

	go hub.Run()
	t.Cleanup(func() { hub.Shutdown() })

	userID := uuid.New()
	client := &ws.Client{
		UserID: userID,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}

	hub.RegisterClient(client)
	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return connectedUser == userID
	}, time.Second, 10*time.Millisecond)

	hub.UnregisterClient(client)
	// onDisconnect is debounced by 5 seconds in hub.go
	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return disconnectedUser == userID
	}, 7*time.Second, 100*time.Millisecond)
}

func TestHub_Shutdown(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()

	userID := uuid.New()
	client := &ws.Client{
		UserID: userID,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}

	hub.RegisterClient(client)
	time.Sleep(10 * time.Millisecond)
	assert.True(t, hub.IsOnline(userID))

	hub.Shutdown()
	time.Sleep(10 * time.Millisecond)
	// After shutdown, hub should not process new operations
}

func TestHub_SendToUser_Offline(t *testing.T) {
	hub := startHub(t)
	// SendToUser for an offline user should not panic
	hub.SendToUser(uuid.New(), []byte("nobody"))
}

func TestHub_SendToUser_BufferFull(t *testing.T) {
	hub := startHub(t)
	userID := uuid.New()

	// Create client with a tiny buffer
	client := &ws.Client{
		UserID: userID,
		Send:   make(chan []byte, 1),
		Hub:    hub,
	}

	hub.RegisterClient(client)
	time.Sleep(10 * time.Millisecond)

	// Fill the buffer
	client.Send <- []byte("fill")

	// This should hit the default branch (buffer full) and not block
	hub.SendToUser(userID, []byte("overflow"))
	time.Sleep(10 * time.Millisecond)
}

func TestHub_SendToRoom_BufferFull(t *testing.T) {
	hub := startHub(t)
	userID := uuid.New()

	client := &ws.Client{
		UserID: userID,
		Send:   make(chan []byte, 1),
		Hub:    hub,
	}

	hub.RegisterClient(client)
	time.Sleep(10 * time.Millisecond)

	hub.JoinRoom(client, "room:full")

	// Fill the buffer
	client.Send <- []byte("fill")

	// Broadcast should hit buffer-full default branch
	hub.SendToRoom("room:full", []byte("overflow"), uuid.Nil)
	time.Sleep(20 * time.Millisecond)
}

func TestHub_Reconnect_CancelsTimer(t *testing.T) {
	hub := ws.NewHub()

	var mu sync.Mutex
	disconnected := false
	hub.SetEventCallbacks(
		func(id uuid.UUID) {},
		func(id uuid.UUID) {
			mu.Lock()
			disconnected = true
			mu.Unlock()
		},
	)

	go hub.Run()
	t.Cleanup(func() { hub.Shutdown() })

	userID := uuid.New()
	client := &ws.Client{
		UserID: userID,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}

	// Register then unregister (starts disconnect timer)
	hub.RegisterClient(client)
	time.Sleep(10 * time.Millisecond)
	hub.UnregisterClient(client)
	time.Sleep(10 * time.Millisecond)

	// Quickly re-register (should cancel the disconnect timer)
	client2 := &ws.Client{
		UserID: userID,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}
	hub.RegisterClient(client2)
	time.Sleep(10 * time.Millisecond)

	assert.True(t, hub.IsOnline(userID))

	// Wait long enough for the debounce timer to have fired if it wasn't cancelled
	time.Sleep(6 * time.Second)

	mu.Lock()
	wasDisconnected := disconnected
	mu.Unlock()

	assert.False(t, wasDisconnected, "disconnect callback should not have been called after reconnect")
}

func TestHub_Unregister_RoomCleanup(t *testing.T) {
	hub := startHub(t)
	userID := uuid.New()

	client := &ws.Client{
		UserID: userID,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}

	hub.RegisterClient(client)
	time.Sleep(10 * time.Millisecond)

	hub.JoinRoom(client, "room:cleanup")
	assert.Len(t, hub.GetRoomMembers("room:cleanup"), 1)

	hub.UnregisterClient(client)
	time.Sleep(10 * time.Millisecond)

	// After unregister, room should be empty/removed
	assert.Nil(t, hub.GetRoomMembers("room:cleanup"))
}

func TestHub_SendToRoom_NonexistentRoom(t *testing.T) {
	hub := startHub(t)
	// Should not panic
	hub.SendToRoom("nonexistent", []byte("data"), uuid.Nil)
	time.Sleep(10 * time.Millisecond)
}

func TestHub_LeaveRoom_LastMember(t *testing.T) {
	hub := startHub(t)
	userID := uuid.New()

	client := &ws.Client{
		UserID: userID,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}

	hub.RegisterClient(client)
	time.Sleep(10 * time.Millisecond)

	hub.JoinRoom(client, "room:leave")
	assert.Len(t, hub.GetRoomMembers("room:leave"), 1)

	hub.LeaveRoom(client, "room:leave")
	assert.Nil(t, hub.GetRoomMembers("room:leave"))

	// LeaveRoom on nonexistent room should not panic
	hub.LeaveRoom(client, "nonexistent")
}
