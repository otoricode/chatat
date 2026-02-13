package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/pkg/apperror"
)

// BackupService defines operations for backup management.
type BackupService interface {
	ExportUserData(ctx context.Context, userID uuid.UUID) (*model.BackupBundle, error)
	ImportUserData(ctx context.Context, userID uuid.UUID, bundle *model.BackupBundle) error
	LogBackup(ctx context.Context, userID uuid.UUID, input model.LogBackupInput) (*model.BackupRecord, error)
	GetBackupHistory(ctx context.Context, userID uuid.UUID) ([]model.BackupRecord, error)
	GetLatestBackup(ctx context.Context, userID uuid.UUID) (*model.BackupRecord, error)
}

type backupService struct {
	backupRepo   repository.BackupRepository
	userRepo     repository.UserRepository
	chatRepo     repository.ChatRepository
	messageRepo  repository.MessageRepository
	contactRepo  repository.ContactRepository
	documentRepo repository.DocumentRepository
}

// NewBackupService creates a new BackupService.
func NewBackupService(
	backupRepo repository.BackupRepository,
	userRepo repository.UserRepository,
	chatRepo repository.ChatRepository,
	messageRepo repository.MessageRepository,
	contactRepo repository.ContactRepository,
	documentRepo repository.DocumentRepository,
) BackupService {
	return &backupService{
		backupRepo:   backupRepo,
		userRepo:     userRepo,
		chatRepo:     chatRepo,
		messageRepo:  messageRepo,
		contactRepo:  contactRepo,
		documentRepo: documentRepo,
	}
}

func (s *backupService) ExportUserData(ctx context.Context, userID uuid.UUID) (*model.BackupBundle, error) {
	// Get user profile
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("export user: %w", err))
	}
	if user == nil {
		return nil, apperror.NotFound("user", userID.String())
	}

	profile := &model.UserExport{
		Name:   user.Name,
		Avatar: user.Avatar,
		Status: user.Status,
	}

	// Get user chats
	chatItems, err := s.chatRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("export chats: %w", err))
	}

	var chatExports []model.ChatExport
	for _, item := range chatItems {
		chatExports = append(chatExports, model.ChatExport{
			ServerID:    item.Chat.ID.String(),
			Type:        string(item.Chat.Type),
			Name:        item.Chat.Name,
			Icon:        item.Chat.Icon,
			Description: item.Chat.Description,
			CreatedAt:   item.Chat.CreatedAt.Format(time.RFC3339),
		})
	}

	// Get messages (latest 1000 per chat)
	var messageExports []model.MessageExport
	for _, item := range chatItems {
		msgs, err := s.messageRepo.ListByChat(ctx, item.Chat.ID, nil, 1000)
		if err != nil {
			continue // skip failed chats
		}
		for _, msg := range msgs {
			if msg.IsDeleted {
				continue
			}
			messageExports = append(messageExports, model.MessageExport{
				ServerID:  msg.ID.String(),
				ChatID:    msg.ChatID.String(),
				SenderID:  msg.SenderID.String(),
				Content:   msg.Content,
				Type:      string(msg.Type),
				CreatedAt: msg.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	// Get contacts
	contacts, err := s.contactRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("export contacts: %w", err))
	}

	var contactExports []model.ContactExport
	for _, c := range contacts {
		contactExports = append(contactExports, model.ContactExport{
			UserID:  c.ContactUserID.String(),
			Name:    c.ContactName,
			AddedAt: c.SyncedAt.Format(time.RFC3339),
		})
	}

	// Get documents
	docs, err := s.documentRepo.ListByOwner(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("export documents: %w", err))
	}

	var docExports []model.DocumentExport
	for _, d := range docs {
		de := model.DocumentExport{
			ServerID: d.ID.String(),
			Title:    d.Title,
			Icon:     d.Icon,
		}
		if d.ChatID != nil {
			de.ChatID = d.ChatID.String()
		}
		if d.TopicID != nil {
			de.TopicID = d.TopicID.String()
		}
		docExports = append(docExports, de)
	}

	bundle := &model.BackupBundle{
		Version:   1,
		UserID:    userID.String(),
		CreatedAt: time.Now(),
		Data: model.BackupData{
			Profile:   profile,
			Chats:     chatExports,
			Messages:  messageExports,
			Contacts:  contactExports,
			Documents: docExports,
		},
	}

	return bundle, nil
}

func (s *backupService) ImportUserData(ctx context.Context, userID uuid.UUID, bundle *model.BackupBundle) error {
	if bundle == nil {
		return apperror.BadRequest("backup bundle is required")
	}
	if bundle.Version != 1 {
		return apperror.BadRequest("unsupported backup version")
	}

	// Update profile if present
	if bundle.Data.Profile != nil {
		name := bundle.Data.Profile.Name
		avatar := bundle.Data.Profile.Avatar
		status := bundle.Data.Profile.Status
		_, err := s.userRepo.Update(ctx, userID, model.UpdateUserInput{
			Name:   &name,
			Avatar: &avatar,
			Status: &status,
		})
		if err != nil {
			return apperror.Internal(fmt.Errorf("import profile: %w", err))
		}
	}

	// Import contacts (skip existing)
	for _, c := range bundle.Data.Contacts {
		contactUID, err := uuid.Parse(c.UserID)
		if err != nil {
			continue
		}
		// Upsert will skip if exists
		_ = s.contactRepo.Upsert(ctx, userID, contactUID, c.Name)
	}

	return nil
}

func (s *backupService) LogBackup(ctx context.Context, userID uuid.UUID, input model.LogBackupInput) (*model.BackupRecord, error) {
	if input.Platform == "" {
		return nil, apperror.BadRequest("platform is required")
	}
	if input.Platform != model.BackupPlatformGoogleDrive && input.Platform != model.BackupPlatformICloud {
		return nil, apperror.BadRequest("invalid platform")
	}

	rec, err := s.backupRepo.Create(ctx, userID, input)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("log backup: %w", err))
	}
	return rec, nil
}

func (s *backupService) GetBackupHistory(ctx context.Context, userID uuid.UUID) ([]model.BackupRecord, error) {
	records, err := s.backupRepo.FindByUserID(ctx, userID, 20)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("get backup history: %w", err))
	}
	if records == nil {
		records = []model.BackupRecord{}
	}
	return records, nil
}

func (s *backupService) GetLatestBackup(ctx context.Context, userID uuid.UUID) (*model.BackupRecord, error) {
	rec, err := s.backupRepo.FindLatestByUser(ctx, userID)
	if err != nil {
		return nil, nil // No backup found is not an error
	}
	return rec, nil
}
