package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/pkg/apperror"
)

// UserService defines operations for user profile management.
type UserService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, input model.UpdateUserInput) (*model.User, error)
	SetupProfile(ctx context.Context, userID uuid.UUID, name, avatar string) (*model.User, error)
	UpdateLastSeen(ctx context.Context, userID uuid.UUID) error
	DeleteAccount(ctx context.Context, userID uuid.UUID) error
}

type userService struct {
	userRepo      repository.UserRepository
	lastSeenMu    sync.Mutex
	lastSeenCache map[uuid.UUID]time.Time
	debounceDur   time.Duration
}

// NewUserService creates a new UserService.
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo:      userRepo,
		lastSeenCache: make(map[uuid.UUID]time.Time),
		debounceDur:   30 * time.Second,
	}
}

func (s *userService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	return user, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID uuid.UUID, input model.UpdateUserInput) (*model.User, error) {
	if err := validateUpdateInput(input); err != nil {
		return nil, err
	}

	user, err := s.userRepo.Update(ctx, userID, input)
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return user, nil
}

func (s *userService) SetupProfile(ctx context.Context, userID uuid.UUID, name, avatar string) (*model.User, error) {
	if name == "" {
		return nil, apperror.Validation("name", "name is required")
	}
	if len(name) > 100 {
		return nil, apperror.Validation("name", "name must be at most 100 characters")
	}
	if avatar == "" {
		avatar = "\U0001F464" // 👤
	}
	if !isValidEmoji(avatar) {
		return nil, apperror.Validation("avatar", "avatar must be a valid emoji")
	}

	namePtr := &name
	avatarPtr := &avatar
	user, err := s.userRepo.Update(ctx, userID, model.UpdateUserInput{
		Name:   namePtr,
		Avatar: avatarPtr,
	})
	if err != nil {
		return nil, fmt.Errorf("setup profile: %w", err)
	}

	// Compute and store phone hash for contact sync
	if user.Phone != "" && user.PhoneHash == "" {
		hash := hashPhone(user.Phone)
		if err := s.userRepo.UpdatePhoneHash(ctx, userID, hash); err != nil {
			log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to update phone hash")
		}
		user.PhoneHash = hash
	}

	return user, nil
}

func (s *userService) UpdateLastSeen(ctx context.Context, userID uuid.UUID) error {
	s.lastSeenMu.Lock()
	last, ok := s.lastSeenCache[userID]
	if ok && time.Since(last) < s.debounceDur {
		s.lastSeenMu.Unlock()
		return nil
	}
	s.lastSeenCache[userID] = time.Now()
	s.lastSeenMu.Unlock()

	if err := s.userRepo.UpdateLastSeen(ctx, userID); err != nil {
		return fmt.Errorf("update last seen: %w", err)
	}
	return nil
}

func (s *userService) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
}

// validateUpdateInput validates the UpdateUserInput fields.
func validateUpdateInput(input model.UpdateUserInput) error {
	if input.Name != nil {
		if *input.Name == "" {
			return apperror.Validation("name", "name cannot be empty")
		}
		if len(*input.Name) > 100 {
			return apperror.Validation("name", "name must be at most 100 characters")
		}
	}
	if input.Avatar != nil && *input.Avatar != "" {
		if !isValidEmoji(*input.Avatar) {
			return apperror.Validation("avatar", "avatar must be a valid emoji")
		}
	}
	if input.Status != nil && len(*input.Status) > 200 {
		return apperror.Validation("status", "status must be at most 200 characters")
	}
	if input.PrivacySettings != nil {
		if err := validatePrivacySettings(*input.PrivacySettings); err != nil {
			return err
		}
	}
	return nil
}

// validVisibility values for privacy settings.
var validVisibility = map[string]bool{
	"everyone": true,
	"contacts": true,
	"nobody":   true,
}

// validatePrivacySettings validates privacy setting values.
func validatePrivacySettings(ps model.PrivacySettings) error {
	if !validVisibility[ps.LastSeenVisibility] {
		return apperror.Validation("lastSeenVisibility", "must be everyone, contacts, or nobody")
	}
	if !validVisibility[ps.OnlineVisibility] {
		return apperror.Validation("onlineVisibility", "must be everyone, contacts, or nobody")
	}
	if !validVisibility[ps.ProfilePhotoVisibility] {
		return apperror.Validation("profilePhotoVisibility", "must be everyone, contacts, or nobody")
	}
	return nil
}

// isValidEmoji checks if the string is a valid emoji (1-4 runes, non-ASCII).
func isValidEmoji(s string) bool {
	if s == "" {
		return false
	}
	count := utf8.RuneCountInString(s)
	// Emojis are typically 1-4 runes (some are compound with ZWJ etc)
	if count < 1 || count > 10 {
		return false
	}
	// At least the first rune should be non-ASCII (emoji range)
	r, _ := utf8.DecodeRuneInString(s)
	return r > 127
}

// hashPhone returns the SHA-256 hex digest of a phone number.
func hashPhone(phone string) string {
	h := sha256.Sum256([]byte(phone))
	return hex.EncodeToString(h[:])
}
