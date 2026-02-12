package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/internal/ws"
	"github.com/otoritech/chatat/pkg/apperror"
)

// ContactMatch represents a matched contact from phone hash sync.
type ContactMatch struct {
	PhoneHash string    `json:"phoneHash"`
	UserID    uuid.UUID `json:"userId"`
	Name      string    `json:"name"`
	Avatar    string    `json:"avatar"`
	Status    string    `json:"status"`
	LastSeen  time.Time `json:"lastSeen"`
}

// ContactInfo represents a contact with online status.
type ContactInfo struct {
	UserID   uuid.UUID `json:"userId"`
	Phone    string    `json:"phone"`
	Name     string    `json:"name"`
	Avatar   string    `json:"avatar"`
	Status   string    `json:"status"`
	IsOnline bool      `json:"isOnline"`
	LastSeen time.Time `json:"lastSeen"`
}

// ContactService defines operations for contact management.
type ContactService interface {
	SyncContacts(ctx context.Context, userID uuid.UUID, phoneHashes []string) ([]ContactMatch, error)
	GetContacts(ctx context.Context, userID uuid.UUID) ([]ContactInfo, error)
	SearchByPhone(ctx context.Context, phone string) (*model.User, error)
	GetContactProfile(ctx context.Context, contactUserID uuid.UUID) (*ContactInfo, error)
}

type contactService struct {
	userRepo    repository.UserRepository
	contactRepo repository.ContactRepository
	hub         *ws.Hub
}

// NewContactService creates a new ContactService.
func NewContactService(
	userRepo repository.UserRepository,
	contactRepo repository.ContactRepository,
	hub *ws.Hub,
) ContactService {
	return &contactService{
		userRepo:    userRepo,
		contactRepo: contactRepo,
		hub:         hub,
	}
}

func (s *contactService) SyncContacts(ctx context.Context, userID uuid.UUID, phoneHashes []string) ([]ContactMatch, error) {
	if len(phoneHashes) == 0 {
		return []ContactMatch{}, nil
	}
	if len(phoneHashes) > 5000 {
		return nil, apperror.BadRequest("too many phone hashes, maximum 5000")
	}

	// Find users matching the provided phone hashes
	users, err := s.userRepo.FindByPhoneHashes(ctx, phoneHashes)
	if err != nil {
		return nil, fmt.Errorf("find by phone hashes: %w", err)
	}

	// Build hash→user map for quick lookup
	hashMap := make(map[string]*model.User, len(users))
	for _, u := range users {
		if u.PhoneHash != "" {
			hashMap[u.PhoneHash] = u
		}
	}

	// Build matches and cache contacts
	var matches []ContactMatch
	var contactInputs []repository.ContactUpsertInput

	for _, hash := range phoneHashes {
		u, ok := hashMap[hash]
		if !ok {
			continue
		}
		// Don't match yourself
		if u.ID == userID {
			continue
		}
		matches = append(matches, ContactMatch{
			PhoneHash: hash,
			UserID:    u.ID,
			Name:      u.Name,
			Avatar:    u.Avatar,
			Status:    u.Status,
			LastSeen:  u.LastSeen,
		})
		contactInputs = append(contactInputs, repository.ContactUpsertInput{
			ContactUserID: u.ID,
			ContactName:   u.Name,
		})
	}

	// Cache matched contacts
	if len(contactInputs) > 0 {
		if err := s.contactRepo.UpsertBatch(ctx, userID, contactInputs); err != nil {
			return nil, fmt.Errorf("cache contacts: %w", err)
		}
	}

	if matches == nil {
		matches = []ContactMatch{}
	}

	return matches, nil
}

func (s *contactService) GetContacts(ctx context.Context, userID uuid.UUID) ([]ContactInfo, error) {
	cached, err := s.contactRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get cached contacts: %w", err)
	}

	if len(cached) == 0 {
		return []ContactInfo{}, nil
	}

	// Fetch latest user profiles
	contacts := make([]ContactInfo, 0, len(cached))
	for _, c := range cached {
		user, err := s.userRepo.FindByID(ctx, c.ContactUserID)
		if err != nil {
			// Skip contacts whose accounts were deleted
			if apperror.IsNotFound(err) {
				continue
			}
			return nil, fmt.Errorf("get contact user: %w", err)
		}

		contacts = append(contacts, ContactInfo{
			UserID:   user.ID,
			Phone:    user.Phone,
			Name:     user.Name,
			Avatar:   user.Avatar,
			Status:   user.Status,
			IsOnline: s.hub.IsOnline(user.ID),
			LastSeen: user.LastSeen,
		})
	}

	// Sort: online first, then alphabetically by name
	sortContacts(contacts)

	return contacts, nil
}

func (s *contactService) SearchByPhone(ctx context.Context, phone string) (*model.User, error) {
	user, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("search by phone: %w", err)
	}
	return user, nil
}

func (s *contactService) GetContactProfile(ctx context.Context, contactUserID uuid.UUID) (*ContactInfo, error) {
	user, err := s.userRepo.FindByID(ctx, contactUserID)
	if err != nil {
		return nil, fmt.Errorf("get contact profile: %w", err)
	}

	return &ContactInfo{
		UserID:   user.ID,
		Phone:    user.Phone,
		Name:     user.Name,
		Avatar:   user.Avatar,
		Status:   user.Status,
		IsOnline: s.hub.IsOnline(user.ID),
		LastSeen: user.LastSeen,
	}, nil
}

// sortContacts sorts contacts: online first, then alphabetically by name.
func sortContacts(contacts []ContactInfo) {
	for i := 1; i < len(contacts); i++ {
		for j := i; j > 0; j-- {
			if shouldSwap(contacts[j-1], contacts[j]) {
				contacts[j-1], contacts[j] = contacts[j], contacts[j-1]
			} else {
				break
			}
		}
	}
}

func shouldSwap(a, b ContactInfo) bool {
	// Online users first
	if a.IsOnline != b.IsOnline {
		return !a.IsOnline && b.IsOnline
	}
	// Then alphabetically by name
	return a.Name > b.Name
}
