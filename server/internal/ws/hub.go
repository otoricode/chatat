package ws

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// disconnectDebounce is the time to wait before broadcasting offline status.
const disconnectDebounce = 5 * time.Second

// Hub maintains the set of active clients and broadcasts messages to rooms.
type Hub struct {
	clients          map[uuid.UUID]*Client
	rooms            map[string]map[uuid.UUID]*Client
	register         chan *Client
	unregister       chan *Client
	broadcast        chan *BroadcastMessage
	done             chan struct{}
	mu               sync.RWMutex
	onConnect        func(userID uuid.UUID)
	onDisconnect     func(userID uuid.UUID)
	disconnectTimers map[uuid.UUID]*time.Timer
}

// BroadcastMessage represents a message to be broadcast to a room.
type BroadcastMessage struct {
	Room    string
	Data    []byte
	Exclude uuid.UUID
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		clients:          make(map[uuid.UUID]*Client),
		rooms:            make(map[string]map[uuid.UUID]*Client),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		broadcast:        make(chan *BroadcastMessage, 256),
		done:             make(chan struct{}),
		disconnectTimers: make(map[uuid.UUID]*time.Timer),
	}
}

// SetEventCallbacks sets the callbacks for client connect/disconnect events.
// The callbacks are invoked asynchronously from the Hub's event loop.
func (h *Hub) SetEventCallbacks(onConnect, onDisconnect func(uuid.UUID)) {
	h.onConnect = onConnect
	h.onDisconnect = onDisconnect
}

// Run starts the hub event loop. Should be called in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case <-h.done:
			return

		case client := <-h.register:
			h.mu.Lock()
			// Cancel pending disconnect timer if user reconnects quickly
			if timer, ok := h.disconnectTimers[client.UserID]; ok {
				timer.Stop()
				delete(h.disconnectTimers, client.UserID)
				log.Debug().Str("user_id", client.UserID.String()).Msg("reconnect: cancelled offline broadcast")
			}
			h.clients[client.UserID] = client
			h.mu.Unlock()
			log.Debug().Str("user_id", client.UserID.String()).Msg("client registered")

			if h.onConnect != nil {
				go h.onConnect(client.UserID)
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)

				// Remove from all rooms
				for roomID, members := range h.rooms {
					delete(members, client.UserID)
					if len(members) == 0 {
						delete(h.rooms, roomID)
					}
				}

				// Debounce disconnect: wait before broadcasting offline
				if h.onDisconnect != nil {
					userID := client.UserID
					h.disconnectTimers[userID] = time.AfterFunc(disconnectDebounce, func() {
						h.mu.Lock()
						delete(h.disconnectTimers, userID)
						h.mu.Unlock()

						// Only broadcast offline if user is still disconnected
						if !h.IsOnline(userID) {
							h.onDisconnect(userID)
						}
					})
				}
			}
			h.mu.Unlock()
			log.Debug().Str("user_id", client.UserID.String()).Msg("client unregistered")

		case msg := <-h.broadcast:
			h.mu.RLock()
			if members, ok := h.rooms[msg.Room]; ok {
				for userID, client := range members {
					if userID == msg.Exclude {
						continue
					}
					select {
					case client.Send <- msg.Data:
					default:
						// Client buffer full, skip
						log.Warn().Str("user_id", userID.String()).Msg("client send buffer full, skipping")
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Shutdown stops the hub event loop.
func (h *Hub) Shutdown() {
	close(h.done)
}

// RegisterClient registers a client with the hub.
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient unregisters a client from the hub.
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// JoinRoom adds a client to a room.
func (h *Hub) JoinRoom(client *Client, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[roomID]; !ok {
		h.rooms[roomID] = make(map[uuid.UUID]*Client)
	}
	h.rooms[roomID][client.UserID] = client
	log.Debug().Str("user_id", client.UserID.String()).Str("room", roomID).Msg("client joined room")
}

// LeaveRoom removes a client from a room.
func (h *Hub) LeaveRoom(client *Client, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if members, ok := h.rooms[roomID]; ok {
		delete(members, client.UserID)
		if len(members) == 0 {
			delete(h.rooms, roomID)
		}
	}
}

// SendToUser sends data directly to a specific user.
func (h *Hub) SendToUser(userID uuid.UUID, data []byte) {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()

	if ok {
		select {
		case client.Send <- data:
		default:
			log.Warn().Str("user_id", userID.String()).Msg("send to user: buffer full")
		}
	}
}

// SendToRoom broadcasts data to all clients in a room, optionally excluding one.
func (h *Hub) SendToRoom(roomID string, data []byte, excludeUserID uuid.UUID) {
	h.broadcast <- &BroadcastMessage{
		Room:    roomID,
		Data:    data,
		Exclude: excludeUserID,
	}
}

// IsOnline checks if a user is currently connected.
func (h *Hub) IsOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

// GetOnlineUsers returns the list of online user IDs from the given set.
func (h *Hub) GetOnlineUsers(userIDs []uuid.UUID) []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	online := make([]uuid.UUID, 0)
	for _, id := range userIDs {
		if _, ok := h.clients[id]; ok {
			online = append(online, id)
		}
	}
	return online
}

// GetRoomMembers returns the user IDs of all clients in a room.
func (h *Hub) GetRoomMembers(roomID string) []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	members, ok := h.rooms[roomID]
	if !ok {
		return nil
	}

	result := make([]uuid.UUID, 0, len(members))
	for userID := range members {
		result = append(result, userID)
	}
	return result
}
