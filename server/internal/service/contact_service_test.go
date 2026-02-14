package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/internal/ws"
)

type mockContactRepo struct {
	contacts     map[uuid.UUID][]repository.UserContact
	findByUIDErr error
}

func newMockContactRepo() *mockContactRepo {
	return &mockContactRepo{
		contacts: make(map[uuid.UUID][]repository.UserContact),
	}
}

func (m *mockContactRepo) Upsert(_ context.Context, userID, contactUserID uuid.UUID, contactName string) error {
	m.contacts[userID] = append(m.contacts[userID], repository.UserContact{
		UserID:        userID,
		ContactUserID: contactUserID,
		ContactName:   contactName,
	})
	return nil
}

func (m *mockContactRepo) UpsertBatch(_ context.Context, userID uuid.UUID, contacts []repository.ContactUpsertInput) error {
	for _, c := range contacts {
		m.contacts[userID] = append(m.contacts[userID], repository.UserContact{
			UserID:        userID,
			ContactUserID: c.ContactUserID,
			ContactName:   c.ContactName,
		})
	}
	return nil
}

func (m *mockContactRepo) FindByUserID(_ context.Context, userID uuid.UUID) ([]repository.UserContact, error) {
	if m.findByUIDErr != nil {
		return nil, m.findByUIDErr
	}
	return m.contacts[userID], nil
}

func (m *mockContactRepo) Delete(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (m *mockContactRepo) DeleteAllByUserID(_ context.Context, userID uuid.UUID) error {
	delete(m.contacts, userID)
	return nil
}

func (m *mockContactRepo) FindContactsOf(_ context.Context, contactUserID uuid.UUID) ([]uuid.UUID, error) {
	var result []uuid.UUID
	for ownerID, contacts := range m.contacts {
		for _, c := range contacts {
			if c.ContactUserID == contactUserID {
				result = append(result, ownerID)
			}
		}
	}
	return result, nil
}

func TestContactService_SyncContacts(t *testing.T) {
	userRepo := newMockUserRepo()
	contactRepo := newMockContactRepo()
	hub := ws.NewHub()
	svc := NewContactService(userRepo, contactRepo, hub)

	myID := uuid.New()
	userRepo.addUser(&model.User{ID: myID, Phone: "+6281111111111", PhoneHash: hashPhone("+6281111111111")})

	// Add some other users
	user2 := &model.User{ID: uuid.New(), Phone: "+6282222222222", Name: "Alice", Avatar: "\U0001F60A", PhoneHash: hashPhone("+6282222222222")}
	user3 := &model.User{ID: uuid.New(), Phone: "+6283333333333", Name: "Bob", Avatar: "\U0001F60E", PhoneHash: hashPhone("+6283333333333")}
	userRepo.addUser(user2)
	userRepo.addUser(user3)

	t.Run("matches found", func(t *testing.T) {
		hashes := []string{
			hashPhone("+6282222222222"),
			hashPhone("+6283333333333"),
			hashPhone("+6284444444444"), // non-existent
		}
		matches, err := svc.SyncContacts(context.Background(), myID, hashes)
		require.NoError(t, err)
		assert.Len(t, matches, 2)
	})

	t.Run("no matches", func(t *testing.T) {
		hashes := []string{hashPhone("+6289999999999")}
		matches, err := svc.SyncContacts(context.Background(), myID, hashes)
		require.NoError(t, err)
		assert.Empty(t, matches)
	})

	t.Run("empty hashes", func(t *testing.T) {
		matches, err := svc.SyncContacts(context.Background(), myID, []string{})
		require.NoError(t, err)
		assert.Empty(t, matches)
	})

	t.Run("skips self", func(t *testing.T) {
		hashes := []string{hashPhone("+6281111111111")} // own phone hash
		matches, err := svc.SyncContacts(context.Background(), myID, hashes)
		require.NoError(t, err)
		assert.Empty(t, matches)
	})

	t.Run("contacts cached", func(t *testing.T) {
		cached := contactRepo.contacts[myID]
		assert.True(t, len(cached) >= 2)
	})
}

func TestContactService_GetContacts(t *testing.T) {
	userRepo := newMockUserRepo()
	contactRepo := newMockContactRepo()
	hub := ws.NewHub()
	svc := NewContactService(userRepo, contactRepo, hub)

	myID := uuid.New()
	user2 := &model.User{ID: uuid.New(), Phone: "+6282222222222", Name: "Zara", Avatar: "\U0001F60A"}
	user3 := &model.User{ID: uuid.New(), Phone: "+6283333333333", Name: "Alice", Avatar: "\U0001F60E"}
	userRepo.addUser(user2)
	userRepo.addUser(user3)

	// Pre-populate contacts
	contactRepo.contacts[myID] = []repository.UserContact{
		{UserID: myID, ContactUserID: user2.ID, ContactName: "Zara"},
		{UserID: myID, ContactUserID: user3.ID, ContactName: "Alice"},
	}

	t.Run("returns sorted contacts", func(t *testing.T) {
		contacts, err := svc.GetContacts(context.Background(), myID)
		require.NoError(t, err)
		assert.Len(t, contacts, 2)
		// Sorted alphabetically (both offline): Alice before Zara
		assert.Equal(t, "Alice", contacts[0].Name)
		assert.Equal(t, "Zara", contacts[1].Name)
	})

	t.Run("empty contacts", func(t *testing.T) {
		contacts, err := svc.GetContacts(context.Background(), uuid.New())
		require.NoError(t, err)
		assert.Empty(t, contacts)
	})
}

func TestContactService_SearchByPhone(t *testing.T) {
	userRepo := newMockUserRepo()
	contactRepo := newMockContactRepo()
	hub := ws.NewHub()
	svc := NewContactService(userRepo, contactRepo, hub)

	userRepo.addUser(&model.User{ID: uuid.New(), Phone: "+6281234567890", Name: "Found"})

	t.Run("found", func(t *testing.T) {
		user, err := svc.SearchByPhone(context.Background(), "+6281234567890")
		require.NoError(t, err)
		assert.Equal(t, "Found", user.Name)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.SearchByPhone(context.Background(), "+6289999999999")
		require.Error(t, err)
	})
}

func TestContactService_GetContactProfile(t *testing.T) {
	userRepo := newMockUserRepo()
	contactRepo := newMockContactRepo()
	hub := ws.NewHub()
	svc := NewContactService(userRepo, contactRepo, hub)

	user := &model.User{ID: uuid.New(), Phone: "+6281234567890", Name: "Profile", Avatar: "\U0001F60A"}
	userRepo.addUser(user)

	t.Run("success", func(t *testing.T) {
		info, err := svc.GetContactProfile(context.Background(), user.ID)
		require.NoError(t, err)
		assert.Equal(t, "Profile", info.Name)
		assert.False(t, info.IsOnline)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetContactProfile(context.Background(), uuid.New())
		require.Error(t, err)
	})
}

func TestSortContacts(t *testing.T) {
	contacts := []ContactInfo{
		{Name: "Zara", IsOnline: false},
		{Name: "Bob", IsOnline: true},
		{Name: "Alice", IsOnline: false},
		{Name: "Charlie", IsOnline: true},
	}

	sortContacts(contacts)

	// Online first (alphabetical): Bob, Charlie
	// Then offline (alphabetical): Alice, Zara
	assert.Equal(t, "Bob", contacts[0].Name)
	assert.Equal(t, "Charlie", contacts[1].Name)
	assert.Equal(t, "Alice", contacts[2].Name)
	assert.Equal(t, "Zara", contacts[3].Name)
}

// --- Error-Path Tests ---

func TestContactService_GetContacts_Errors(t *testing.T) {
	hub := ws.NewHub()

	t.Run("find by user error", func(t *testing.T) {
		userRepo := newMockUserRepo()
		contactRepo := newMockContactRepo()
		contactRepo.findByUIDErr = errors.New("db error")
		svc := NewContactService(userRepo, contactRepo, hub)
		_, err := svc.GetContacts(context.Background(), uuid.New())
		require.Error(t, err)
	})

	t.Run("deleted contact skipped", func(t *testing.T) {
		userRepo := newMockUserRepo()
		contactRepo := newMockContactRepo()
		svc := NewContactService(userRepo, contactRepo, hub)

		myID := uuid.New()
		deletedUserID := uuid.New() // not in userRepo = "deleted"
		goodUser := &model.User{ID: uuid.New(), Phone: "+628111", Name: "Good"}
		userRepo.addUser(goodUser)

		contactRepo.contacts[myID] = []repository.UserContact{
			{UserID: myID, ContactUserID: deletedUserID, ContactName: "Deleted"},
			{UserID: myID, ContactUserID: goodUser.ID, ContactName: "Good"},
		}

		contacts, err := svc.GetContacts(context.Background(), myID)
		require.NoError(t, err)
		assert.Len(t, contacts, 1) // deleted one skipped
		assert.Equal(t, "Good", contacts[0].Name)
	})

	t.Run("find user generic error", func(t *testing.T) {
		userRepo := newMockUserRepo()
		contactRepo := newMockContactRepo()
		svc := NewContactService(userRepo, contactRepo, hub)

		myID := uuid.New()
		badUser := uuid.New()
		userRepo.addUser(&model.User{ID: badUser, Phone: "+628222", Name: "Bad"})
		userRepo.findErr = errors.New("generic db error")

		contactRepo.contacts[myID] = []repository.UserContact{
			{UserID: myID, ContactUserID: badUser, ContactName: "Bad"},
		}

		_, err := svc.GetContacts(context.Background(), myID)
		require.Error(t, err)
		userRepo.findErr = nil
	})
}
