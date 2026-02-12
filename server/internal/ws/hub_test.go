package ws_test

import (
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
