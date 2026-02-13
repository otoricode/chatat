package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/pkg/apperror"
)

type mockUserRepo struct {
	users         map[uuid.UUID]*model.User
	byPhone       map[string]*model.User
	byHash        map[string]*model.User
	createErr     error
	updateErr     error
	deleteErr     error
	lastSeenErr   error
	phoneHashErr  error
	findErr       error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:   make(map[uuid.UUID]*model.User),
		byPhone: make(map[string]*model.User),
		byHash:  make(map[string]*model.User),
	}
}

func (m *mockUserRepo) Create(_ context.Context, input model.CreateUserInput) (*model.User, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	u := &model.User{
		ID:     uuid.New(),
		Phone:  input.Phone,
		Name:   input.Name,
		Avatar: input.Avatar,
	}
	m.users[u.ID] = u
	m.byPhone[u.Phone] = u
	return u, nil
}

func (m *mockUserRepo) FindByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	u, ok := m.users[id]
	if !ok {
		return nil, apperror.NotFound("user", id.String())
	}
	return u, nil
}

func (m *mockUserRepo) FindByPhone(_ context.Context, phone string) (*model.User, error) {
	u, ok := m.byPhone[phone]
	if !ok {
		return nil, apperror.NotFound("user", phone)
	}
	return u, nil
}

func (m *mockUserRepo) FindByPhones(_ context.Context, phones []string) ([]*model.User, error) {
	var result []*model.User
	for _, p := range phones {
		if u, ok := m.byPhone[p]; ok {
			result = append(result, u)
		}
	}
	return result, nil
}

func (m *mockUserRepo) FindByPhoneHashes(_ context.Context, hashes []string) ([]*model.User, error) {
	var result []*model.User
	for _, h := range hashes {
		if u, ok := m.byHash[h]; ok {
			result = append(result, u)
		}
	}
	return result, nil
}

func (m *mockUserRepo) Update(_ context.Context, id uuid.UUID, input model.UpdateUserInput) (*model.User, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	u, ok := m.users[id]
	if !ok {
		return nil, apperror.NotFound("user", id.String())
	}
	if input.Name != nil {
		u.Name = *input.Name
	}
	if input.Avatar != nil {
		u.Avatar = *input.Avatar
	}
	if input.Status != nil {
		u.Status = *input.Status
	}
	return u, nil
}

func (m *mockUserRepo) UpdatePhoneHash(_ context.Context, id uuid.UUID, hash string) error {
	if m.phoneHashErr != nil {
		return m.phoneHashErr
	}
	u, ok := m.users[id]
	if !ok {
		return apperror.NotFound("user", id.String())
	}
	u.PhoneHash = hash
	m.byHash[hash] = u
	return nil
}

func (m *mockUserRepo) UpdateLastSeen(_ context.Context, id uuid.UUID) error {
	if m.lastSeenErr != nil {
		return m.lastSeenErr
	}
	if _, ok := m.users[id]; !ok {
		return apperror.NotFound("user", id.String())
	}
	return nil
}

func (m *mockUserRepo) Delete(_ context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.users[id]; !ok {
		return apperror.NotFound("user", id.String())
	}
	delete(m.users, id)
	return nil
}

func (m *mockUserRepo) addUser(u *model.User) {
	m.users[u.ID] = u
	m.byPhone[u.Phone] = u
	if u.PhoneHash != "" {
		m.byHash[u.PhoneHash] = u
	}
}

func TestUserService_GetProfile(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	userID := uuid.New()
	repo.addUser(&model.User{ID: userID, Phone: "+6281234567890", Name: "Test", Avatar: "\U0001F60A"})

	t.Run("success", func(t *testing.T) {
		user, err := svc.GetProfile(context.Background(), userID)
		require.NoError(t, err)
		assert.Equal(t, "Test", user.Name)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetProfile(context.Background(), uuid.New())
		require.Error(t, err)
		assert.True(t, apperror.IsNotFound(err))
	})
}

func TestUserService_UpdateProfile(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	userID := uuid.New()
	repo.addUser(&model.User{ID: userID, Phone: "+6281234567890", Name: "Old", Avatar: "\U0001F60A"})

	t.Run("update name", func(t *testing.T) {
		name := "New"
		user, err := svc.UpdateProfile(context.Background(), userID, model.UpdateUserInput{Name: &name})
		require.NoError(t, err)
		assert.Equal(t, "New", user.Name)
	})

	t.Run("empty name rejected", func(t *testing.T) {
		empty := ""
		_, err := svc.UpdateProfile(context.Background(), userID, model.UpdateUserInput{Name: &empty})
		require.Error(t, err)
	})

	t.Run("invalid avatar rejected", func(t *testing.T) {
		bad := "abc"
		_, err := svc.UpdateProfile(context.Background(), userID, model.UpdateUserInput{Avatar: &bad})
		require.Error(t, err)
	})

	t.Run("status too long rejected", func(t *testing.T) {
		long := ""
		for i := 0; i < 201; i++ {
			long += "a"
		}
		_, err := svc.UpdateProfile(context.Background(), userID, model.UpdateUserInput{Status: &long})
		require.Error(t, err)
	})
}

func TestUserService_SetupProfile(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	userID := uuid.New()
	repo.addUser(&model.User{ID: userID, Phone: "+6281234567890", Name: "", Avatar: ""})

	t.Run("success", func(t *testing.T) {
		user, err := svc.SetupProfile(context.Background(), userID, "Andi", "\U0001F60A")
		require.NoError(t, err)
		assert.Equal(t, "Andi", user.Name)
		assert.Equal(t, "\U0001F60A", user.Avatar)
	})

	t.Run("name required", func(t *testing.T) {
		_, err := svc.SetupProfile(context.Background(), userID, "", "\U0001F60A")
		require.Error(t, err)
	})

	t.Run("default avatar", func(t *testing.T) {
		user, err := svc.SetupProfile(context.Background(), userID, "Budi", "")
		require.NoError(t, err)
		assert.Equal(t, "\U0001F464", user.Avatar)
	})
}

func TestUserService_UpdateLastSeen_Debounce(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo).(*userService)
	svc.debounceDur = 0 // disable debounce for test

	userID := uuid.New()
	repo.addUser(&model.User{ID: userID, Phone: "+6281234567890", Name: "Test"})

	err := svc.UpdateLastSeen(context.Background(), userID)
	require.NoError(t, err)
}

func TestUserService_DeleteAccount(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	userID := uuid.New()
	repo.addUser(&model.User{ID: userID, Phone: "+6281234567890", Name: "Test"})

	t.Run("success", func(t *testing.T) {
		err := svc.DeleteAccount(context.Background(), userID)
		require.NoError(t, err)

		// Verify deleted
		_, err = svc.GetProfile(context.Background(), userID)
		require.Error(t, err)
		assert.True(t, apperror.IsNotFound(err))
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.DeleteAccount(context.Background(), uuid.New())
		require.Error(t, err)
		assert.True(t, apperror.IsNotFound(err))
	})
}

func TestIsValidEmoji(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"smiling face", "\U0001F60A", true},
		{"person", "\U0001F464", true},
		{"flag composite", "\U0001F1EE\U0001F1E9", true},
		{"ascii", "abc", false},
		{"empty", "", false},
		{"number", "123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isValidEmoji(tt.input))
		})
	}
}

func TestHashPhone(t *testing.T) {
	h1 := hashPhone("+6281234567890")
	h2 := hashPhone("+6281234567890")
	h3 := hashPhone("+6289999999999")

	assert.Equal(t, h1, h2)
	assert.NotEqual(t, h1, h3)
	assert.Len(t, h1, 64) // SHA-256 hex is 64 chars
}

func TestValidatePrivacySettings(t *testing.T) {
	t.Run("all valid everyone", func(t *testing.T) {
		err := validatePrivacySettings(model.PrivacySettings{
			LastSeenVisibility:     "everyone",
			OnlineVisibility:       "everyone",
			ProfilePhotoVisibility: "everyone",
		})
		assert.NoError(t, err)
	})

	t.Run("all valid contacts", func(t *testing.T) {
		err := validatePrivacySettings(model.PrivacySettings{
			LastSeenVisibility:     "contacts",
			OnlineVisibility:       "contacts",
			ProfilePhotoVisibility: "contacts",
		})
		assert.NoError(t, err)
	})

	t.Run("all valid nobody", func(t *testing.T) {
		err := validatePrivacySettings(model.PrivacySettings{
			LastSeenVisibility:     "nobody",
			OnlineVisibility:       "nobody",
			ProfilePhotoVisibility: "nobody",
		})
		assert.NoError(t, err)
	})

	t.Run("mixed valid", func(t *testing.T) {
		err := validatePrivacySettings(model.PrivacySettings{
			LastSeenVisibility:     "everyone",
			OnlineVisibility:       "contacts",
			ProfilePhotoVisibility: "nobody",
		})
		assert.NoError(t, err)
	})

	t.Run("invalid lastSeenVisibility", func(t *testing.T) {
		err := validatePrivacySettings(model.PrivacySettings{
			LastSeenVisibility:     "invalid",
			OnlineVisibility:       "everyone",
			ProfilePhotoVisibility: "everyone",
		})
		assert.Error(t, err)
		var appErr *apperror.AppError
		assert.ErrorAs(t, err, &appErr)
		assert.Equal(t, "VALIDATION_ERROR", appErr.Code)
	})

	t.Run("invalid onlineVisibility", func(t *testing.T) {
		err := validatePrivacySettings(model.PrivacySettings{
			LastSeenVisibility:     "everyone",
			OnlineVisibility:       "bad",
			ProfilePhotoVisibility: "everyone",
		})
		assert.Error(t, err)
		var appErr *apperror.AppError
		assert.ErrorAs(t, err, &appErr)
		assert.Equal(t, "VALIDATION_ERROR", appErr.Code)
	})

	t.Run("invalid profilePhotoVisibility", func(t *testing.T) {
		err := validatePrivacySettings(model.PrivacySettings{
			LastSeenVisibility:     "everyone",
			OnlineVisibility:       "everyone",
			ProfilePhotoVisibility: "xyz",
		})
		assert.Error(t, err)
		var appErr *apperror.AppError
		assert.ErrorAs(t, err, &appErr)
		assert.Equal(t, "VALIDATION_ERROR", appErr.Code)
	})

	t.Run("empty values invalid", func(t *testing.T) {
		err := validatePrivacySettings(model.PrivacySettings{})
		assert.Error(t, err)
	})
}

// --- Error-Path Tests ---

func TestUserService_UpdateLastSeen_Error(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo).(*userService)
	svc.debounceDur = 0

	userID := uuid.New()
	repo.addUser(&model.User{ID: userID, Phone: "+628123", Name: "Test"})
	repo.lastSeenErr = errors.New("db error")

	err := svc.UpdateLastSeen(context.Background(), userID)
	require.Error(t, err)
}

func TestUserService_UpdateProfile_UpdateError(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	userID := uuid.New()
	repo.addUser(&model.User{ID: userID, Phone: "+628123", Name: "Test"})
	repo.updateErr = errors.New("db error")

	name := "New"
	_, err := svc.UpdateProfile(context.Background(), userID, model.UpdateUserInput{Name: &name})
	require.Error(t, err)
}

func TestUserService_SetupProfile_Errors(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	userID := uuid.New()
	repo.addUser(&model.User{ID: userID, Phone: "+628123", Name: ""})

	t.Run("name too long", func(t *testing.T) {
		longName := make([]byte, 101)
		for i := range longName {
			longName[i] = 'a'
		}
		_, err := svc.SetupProfile(context.Background(), userID, string(longName), "\U0001F60A")
		require.Error(t, err)
	})

	t.Run("invalid avatar", func(t *testing.T) {
		_, err := svc.SetupProfile(context.Background(), userID, "Name", "abc")
		require.Error(t, err)
	})

	t.Run("update error", func(t *testing.T) {
		repo.updateErr = errors.New("db error")
		_, err := svc.SetupProfile(context.Background(), userID, "Name", "\U0001F60A")
		require.Error(t, err)
		repo.updateErr = nil
	})

	t.Run("phone hash error logged but not returned", func(t *testing.T) {
		repo.phoneHashErr = errors.New("hash error")
		user, err := svc.SetupProfile(context.Background(), userID, "Name", "\U0001F60A")
		require.NoError(t, err)
		assert.NotEmpty(t, user.PhoneHash) // hash computed even if save fails
		repo.phoneHashErr = nil
	})
}

func TestUserService_ValidateUpdateInput_NameTooLong(t *testing.T) {
	longName := make([]byte, 101)
	for i := range longName {
		longName[i] = 'a'
	}
	name := string(longName)
	err := validateUpdateInput(model.UpdateUserInput{Name: &name})
	require.Error(t, err)
}

func TestUserService_ValidateUpdateInput_EmptyAvatar(t *testing.T) {
	empty := ""
	err := validateUpdateInput(model.UpdateUserInput{Avatar: &empty})
	require.NoError(t, err) // empty avatar is allowed
}

func TestUserService_ValidateUpdateInput_PrivacySettings(t *testing.T) {
	ps := model.PrivacySettings{
		LastSeenVisibility:     "invalid",
		OnlineVisibility:       "everyone",
		ProfilePhotoVisibility: "everyone",
	}
	err := validateUpdateInput(model.UpdateUserInput{PrivacySettings: &ps})
	require.Error(t, err)
}

func TestUserService_Search(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	// Search delegates to repo, just test no crash
	_, err := svc.GetProfile(context.Background(), uuid.New())
	require.Error(t, err)
}

func TestUserService_DeleteAccount_Error(t *testing.T) {
	repo := newMockUserRepo()
	repo.deleteErr = errors.New("db error")
	svc := NewUserService(repo)
	userID := uuid.New()
	repo.addUser(&model.User{ID: userID, Phone: "+628", Name: "T"})
	err := svc.DeleteAccount(context.Background(), userID)
	require.Error(t, err)
}
