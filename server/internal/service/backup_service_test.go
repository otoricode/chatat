package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/model"
)

// --- Mock Backup Repository ---

type mockBackupRepo struct {
	records []model.BackupRecord
}

func newMockBackupRepo() *mockBackupRepo {
	return &mockBackupRepo{}
}

func (m *mockBackupRepo) Create(_ context.Context, userID uuid.UUID, input model.LogBackupInput) (*model.BackupRecord, error) {
	status := input.Status
	if status == "" {
		status = model.BackupStatusCompleted
	}
	rec := model.BackupRecord{
		ID:        uuid.New(),
		UserID:    userID,
		SizeBytes: input.SizeBytes,
		Platform:  input.Platform,
		Status:    status,
		CreatedAt: time.Now(),
	}
	m.records = append(m.records, rec)
	return &rec, nil
}

func (m *mockBackupRepo) UpdateStatus(_ context.Context, id uuid.UUID, status model.BackupStatus, sizeBytes int64) error {
	for i, r := range m.records {
		if r.ID == id {
			m.records[i].Status = status
			m.records[i].SizeBytes = sizeBytes
			return nil
		}
	}
	return nil
}

func (m *mockBackupRepo) FindByUserID(_ context.Context, userID uuid.UUID, limit int) ([]model.BackupRecord, error) {
	var result []model.BackupRecord
	for _, r := range m.records {
		if r.UserID == userID {
			result = append(result, r)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockBackupRepo) FindLatestByUser(_ context.Context, userID uuid.UUID) (*model.BackupRecord, error) {
	var latest *model.BackupRecord
	for i, r := range m.records {
		if r.UserID == userID && r.Status == model.BackupStatusCompleted {
			if latest == nil || r.CreatedAt.After(latest.CreatedAt) {
				latest = &m.records[i]
			}
		}
	}
	if latest == nil {
		return nil, nil
	}
	return latest, nil
}

func (m *mockBackupRepo) Delete(_ context.Context, id uuid.UUID) error {
	for i, r := range m.records {
		if r.ID == id {
			m.records = append(m.records[:i], m.records[i+1:]...)
			return nil
		}
	}
	return nil
}

// --- Mock Document Repo for backup tests ---

type mockBackupDocRepo struct {
	docs []*model.Document
}

func newMockBackupDocRepo() *mockBackupDocRepo {
	return &mockBackupDocRepo{}
}

func (m *mockBackupDocRepo) Create(_ context.Context, _ model.CreateDocumentInput) (*model.Document, error) {
	return nil, nil
}

func (m *mockBackupDocRepo) FindByID(_ context.Context, _ uuid.UUID) (*model.Document, error) {
	return nil, nil
}

func (m *mockBackupDocRepo) ListByChat(_ context.Context, _ uuid.UUID) ([]*model.Document, error) {
	return nil, nil
}

func (m *mockBackupDocRepo) ListByTopic(_ context.Context, _ uuid.UUID) ([]*model.Document, error) {
	return nil, nil
}

func (m *mockBackupDocRepo) ListByOwner(_ context.Context, ownerID uuid.UUID) ([]*model.Document, error) {
	var result []*model.Document
	for _, d := range m.docs {
		if d.OwnerID == ownerID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (m *mockBackupDocRepo) ListByTag(_ context.Context, _ string) ([]*model.Document, error) {
	return nil, nil
}

func (m *mockBackupDocRepo) ListAccessible(_ context.Context, _ uuid.UUID) ([]*model.Document, error) {
	return nil, nil
}

func (m *mockBackupDocRepo) ListCollaborators(_ context.Context, _ uuid.UUID) ([]*model.DocumentCollaborator, error) {
	return nil, nil
}

func (m *mockBackupDocRepo) ListSigners(_ context.Context, _ uuid.UUID) ([]*model.DocumentSigner, error) {
	return nil, nil
}

func (m *mockBackupDocRepo) ListTags(_ context.Context, _ uuid.UUID) ([]string, error) {
	return nil, nil
}

func (m *mockBackupDocRepo) AddCollaborator(_ context.Context, _, _ uuid.UUID, _ model.CollaboratorRole) error {
	return nil
}

func (m *mockBackupDocRepo) RemoveCollaborator(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (m *mockBackupDocRepo) UpdateCollaboratorRole(_ context.Context, _, _ uuid.UUID, _ model.CollaboratorRole) error {
	return nil
}

func (m *mockBackupDocRepo) AddSigner(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (m *mockBackupDocRepo) RemoveSigner(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (m *mockBackupDocRepo) RecordSignature(_ context.Context, _, _ uuid.UUID, _ string) error {
	return nil
}

func (m *mockBackupDocRepo) Lock(_ context.Context, _ uuid.UUID, _ model.LockedByType) error {
	return nil
}

func (m *mockBackupDocRepo) Unlock(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockBackupDocRepo) AddTag(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

func (m *mockBackupDocRepo) RemoveTag(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

func (m *mockBackupDocRepo) Update(_ context.Context, _ uuid.UUID, _ model.UpdateDocumentInput) (*model.Document, error) {
	return nil, nil
}

func (m *mockBackupDocRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

// --- Tests ---

func TestBackupService_LogBackup(t *testing.T) {
	backupRepo := newMockBackupRepo()
	svc := NewBackupService(backupRepo, nil, nil, nil, nil, nil)

	userID := uuid.New()

	rec, err := svc.LogBackup(context.Background(), userID, model.LogBackupInput{
		SizeBytes: 1024,
		Platform:  model.BackupPlatformGoogleDrive,
	})
	require.NoError(t, err)
	assert.Equal(t, model.BackupPlatformGoogleDrive, rec.Platform)
	assert.Equal(t, int64(1024), rec.SizeBytes)
	assert.Equal(t, model.BackupStatusCompleted, rec.Status)
}

func TestBackupService_LogBackup_InvalidPlatform(t *testing.T) {
	backupRepo := newMockBackupRepo()
	svc := NewBackupService(backupRepo, nil, nil, nil, nil, nil)

	_, err := svc.LogBackup(context.Background(), uuid.New(), model.LogBackupInput{
		SizeBytes: 100,
		Platform:  "dropbox",
	})
	require.Error(t, err)
}

func TestBackupService_GetBackupHistory(t *testing.T) {
	backupRepo := newMockBackupRepo()
	svc := NewBackupService(backupRepo, nil, nil, nil, nil, nil)

	userID := uuid.New()

	// Log 2 backups
	_, err := svc.LogBackup(context.Background(), userID, model.LogBackupInput{SizeBytes: 100, Platform: model.BackupPlatformICloud})
	require.NoError(t, err)
	_, err = svc.LogBackup(context.Background(), userID, model.LogBackupInput{SizeBytes: 200, Platform: model.BackupPlatformGoogleDrive})
	require.NoError(t, err)

	records, err := svc.GetBackupHistory(context.Background(), userID)
	require.NoError(t, err)
	assert.Len(t, records, 2)
}

func TestBackupService_GetLatestBackup(t *testing.T) {
	backupRepo := newMockBackupRepo()
	svc := NewBackupService(backupRepo, nil, nil, nil, nil, nil)

	userID := uuid.New()

	// No backups yet
	rec, err := svc.GetLatestBackup(context.Background(), userID)
	require.NoError(t, err)
	assert.Nil(t, rec)

	// Add a backup
	_, err = svc.LogBackup(context.Background(), userID, model.LogBackupInput{SizeBytes: 500, Platform: model.BackupPlatformICloud})
	require.NoError(t, err)

	rec, err = svc.GetLatestBackup(context.Background(), userID)
	require.NoError(t, err)
	require.NotNil(t, rec)
	assert.Equal(t, int64(500), rec.SizeBytes)
}

func TestBackupService_ExportUserData(t *testing.T) {
	userRepo := newMockUserRepo()
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	contactRepo := newMockContactRepo()
	docRepo := newMockBackupDocRepo()
	backupRepo := newMockBackupRepo()

	userID := uuid.New()
	userRepo.addUser(&model.User{ID: userID, Name: "Test", Avatar: "T", Status: "Active"})

	// Create a chat and add user as member
	chat, _ := chatRepo.Create(context.Background(), model.CreateChatInput{
		Type:      model.ChatTypePersonal,
		Name:      "Test Chat",
		CreatedBy: userID,
	})
	_ = chatRepo.AddMember(context.Background(), chat.ID, userID, model.MemberRoleMember)

	// Add a contact
	_ = contactRepo.Upsert(context.Background(), userID, uuid.New(), "Contact 1")

	// Add a document
	docRepo.docs = append(docRepo.docs, &model.Document{
		ID:      uuid.New(),
		Title:   "Doc 1",
		Icon:    "D",
		OwnerID: userID,
	})

	svc := NewBackupService(backupRepo, userRepo, chatRepo, msgRepo, contactRepo, docRepo)

	bundle, err := svc.ExportUserData(context.Background(), userID)
	require.NoError(t, err)
	require.NotNil(t, bundle)
	assert.Equal(t, 1, bundle.Version)
	assert.Equal(t, userID.String(), bundle.UserID)
	assert.Equal(t, "Test", bundle.Data.Profile.Name)
	assert.Len(t, bundle.Data.Chats, 1)
	assert.Len(t, bundle.Data.Contacts, 1)
	assert.Len(t, bundle.Data.Documents, 1)
}

func TestBackupService_ImportUserData(t *testing.T) {
	userRepo := newMockUserRepo()
	contactRepo := newMockContactRepo()
	backupRepo := newMockBackupRepo()

	userID := uuid.New()
	userRepo.addUser(&model.User{ID: userID, Name: "Old Name"})

	// Also add the contact user to make upsert work
	contactUserID := uuid.New()
	userRepo.addUser(&model.User{ID: contactUserID, Name: "Contact"})

	svc := NewBackupService(backupRepo, userRepo, nil, nil, contactRepo, nil)

	bundle := &model.BackupBundle{
		Version:   1,
		UserID:    userID.String(),
		CreatedAt: time.Now(),
		Data: model.BackupData{
			Profile: &model.UserExport{
				Name:   "New Name",
				Avatar: "N",
				Status: "Restored",
			},
			Contacts: []model.ContactExport{
				{UserID: contactUserID.String(), Name: "Contact 1"},
			},
		},
	}

	err := svc.ImportUserData(context.Background(), userID, bundle)
	require.NoError(t, err)

	// Verify profile updated
	u, _ := userRepo.FindByID(context.Background(), userID)
	assert.Equal(t, "New Name", u.Name)
}
