package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/internal/ws"
)

// statusMockContactRepo implements ContactRepository for status tests.
type statusMockContactRepo struct {
	observers map[uuid.UUID][]uuid.UUID
}

func (m *statusMockContactRepo) Upsert(_ context.Context, _, _ uuid.UUID, _ string) error {
	return nil
}

func (m *statusMockContactRepo) UpsertBatch(_ context.Context, _ uuid.UUID, _ []repository.ContactUpsertInput) error {
	return nil
}

func (m *statusMockContactRepo) FindByUserID(_ context.Context, _ uuid.UUID) ([]repository.UserContact, error) {
	return nil, nil
}

func (m *statusMockContactRepo) FindContactsOf(_ context.Context, contactUserID uuid.UUID) ([]uuid.UUID, error) {
	return m.observers[contactUserID], nil
}

func (m *statusMockContactRepo) Delete(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (m *statusMockContactRepo) DeleteAllByUserID(_ context.Context, _ uuid.UUID) error {
	return nil
}

// statusMockUserRepo implements UserRepository for status tests.
type statusMockUserRepo struct {
	user        *model.User
	lastSeenSet bool
}

func (m *statusMockUserRepo) Create(_ context.Context, _ model.CreateUserInput) (*model.User, error) {
	return m.user, nil
}

func (m *statusMockUserRepo) FindByID(_ context.Context, _ uuid.UUID) (*model.User, error) {
	return m.user, nil
}

func (m *statusMockUserRepo) FindByPhone(_ context.Context, _ string) (*model.User, error) {
	return m.user, nil
}

func (m *statusMockUserRepo) FindByPhones(_ context.Context, _ []string) ([]*model.User, error) {
	return nil, nil
}

func (m *statusMockUserRepo) FindByPhoneHashes(_ context.Context, _ []string) ([]*model.User, error) {
	return nil, nil
}

func (m *statusMockUserRepo) Update(_ context.Context, _ uuid.UUID, _ model.UpdateUserInput) (*model.User, error) {
	return m.user, nil
}

func (m *statusMockUserRepo) UpdateLastSeen(_ context.Context, _ uuid.UUID) error {
	m.lastSeenSet = true
	return nil
}

func (m *statusMockUserRepo) UpdatePhoneHash(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

func (m *statusMockUserRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

func TestHub_ConnectBroadcast(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Shutdown()

	userA := uuid.New()
	userB := uuid.New()

	// B has A as a contact (B observes A)
	contactRepo := &statusMockContactRepo{
		observers: map[uuid.UUID][]uuid.UUID{
			userA: {userB},
		},
	}

	userRepo := &statusMockUserRepo{
		user: &model.User{ID: userA, Name: "A"},
	}

	// Register observer B first so they can receive messages
	clientB := &ws.Client{
		UserID: userB,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}
	hub.RegisterClient(clientB)
	time.Sleep(20 * time.Millisecond)

	// Create StatusNotifier (sets hub callbacks)
	_ = service.NewStatusNotifier(hub, contactRepo, userRepo, nil)

	// A connects → B should be notified
	clientA := &ws.Client{
		UserID: userA,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}
	hub.RegisterClient(clientA)
	time.Sleep(100 * time.Millisecond)

	select {
	case msg := <-clientB.Send:
		assert.Contains(t, string(msg), "online_status")
		assert.Contains(t, string(msg), userA.String())
	case <-time.After(500 * time.Millisecond):
		t.Fatal("observer B should have received online_status")
	}
}

func TestHub_DisconnectDebounce(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Shutdown()

	userA := uuid.New()
	userB := uuid.New()

	contactRepo := &statusMockContactRepo{
		observers: map[uuid.UUID][]uuid.UUID{
			userA: {userB},
		},
	}

	userRepo := &statusMockUserRepo{
		user: &model.User{ID: userA, Name: "A"},
	}

	clientB := &ws.Client{
		UserID: userB,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}
	hub.RegisterClient(clientB)
	time.Sleep(20 * time.Millisecond)

	_ = service.NewStatusNotifier(hub, contactRepo, userRepo, nil)

	clientA := &ws.Client{
		UserID: userA,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}
	hub.RegisterClient(clientA)
	time.Sleep(100 * time.Millisecond)

	// Drain the connect event
	select {
	case <-clientB.Send:
	case <-time.After(500 * time.Millisecond):
	}

	// A disconnects
	hub.UnregisterClient(clientA)

	// Within debounce window (5s), no offline broadcast yet
	select {
	case <-clientB.Send:
		// If we get a message within 200ms, the debounce is not working
		t.Fatal("should not receive offline_status immediately (debounce)")
	case <-time.After(200 * time.Millisecond):
		// Good, no immediate broadcast
	}
}

func TestHub_ReconnectCancelsOffline(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Shutdown()

	userA := uuid.New()
	userB := uuid.New()

	contactRepo := &statusMockContactRepo{
		observers: map[uuid.UUID][]uuid.UUID{
			userA: {userB},
		},
	}

	userRepo := &statusMockUserRepo{
		user: &model.User{ID: userA, Name: "A"},
	}

	clientB := &ws.Client{
		UserID: userB,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}
	hub.RegisterClient(clientB)
	time.Sleep(20 * time.Millisecond)

	_ = service.NewStatusNotifier(hub, contactRepo, userRepo, nil)

	clientA := &ws.Client{
		UserID: userA,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}
	hub.RegisterClient(clientA)
	time.Sleep(100 * time.Millisecond)

	// Drain connect event
	select {
	case <-clientB.Send:
	case <-time.After(500 * time.Millisecond):
	}

	// Disconnect A
	hub.UnregisterClient(clientA)
	time.Sleep(50 * time.Millisecond)

	// Reconnect A quickly (before 5s debounce)
	clientA2 := &ws.Client{
		UserID: userA,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}
	hub.RegisterClient(clientA2)
	time.Sleep(100 * time.Millisecond)

	// Drain the reconnect "online" event
	select {
	case <-clientB.Send:
	case <-time.After(500 * time.Millisecond):
	}

	// No offline broadcast should arrive
	select {
	case msg := <-clientB.Send:
		assert.NotContains(t, string(msg), "\"isOnline\":false", "should not get offline_status after quick reconnect")
	case <-time.After(500 * time.Millisecond):
		// Good, no offline broadcast
	}
}
