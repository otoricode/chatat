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

// DocumentService handles document business logic.
type DocumentService interface {
	Create(ctx context.Context, input CreateDocumentInput) (*DocumentFull, error)
	GetByID(ctx context.Context, docID, userID uuid.UUID) (*DocumentFull, error)
	ListByContext(ctx context.Context, contextType string, contextID uuid.UUID) ([]*DocumentListItem, error)
	ListAll(ctx context.Context, userID uuid.UUID) ([]*DocumentListItem, error)
	Update(ctx context.Context, docID uuid.UUID, userID uuid.UUID, input model.UpdateDocumentInput) (*model.Document, error)
	Delete(ctx context.Context, docID, userID uuid.UUID) error
	Duplicate(ctx context.Context, docID, userID uuid.UUID) (*DocumentFull, error)
	AddCollaborator(ctx context.Context, docID, ownerID, userID uuid.UUID, role model.CollaboratorRole) error
	RemoveCollaborator(ctx context.Context, docID, ownerID, userID uuid.UUID) error
	UpdateCollaboratorRole(ctx context.Context, docID, ownerID, userID uuid.UUID, role model.CollaboratorRole) error
	AddTag(ctx context.Context, docID uuid.UUID, tag string) error
	RemoveTag(ctx context.Context, docID uuid.UUID, tag string) error
	GetHistory(ctx context.Context, docID uuid.UUID) ([]*model.DocumentHistory, error)
	LockDocument(ctx context.Context, docID, userID uuid.UUID, mode model.LockedByType) error
	UnlockDocument(ctx context.Context, docID, userID uuid.UUID) error
	AddSigner(ctx context.Context, docID, ownerID, signerID uuid.UUID) error
	RemoveSigner(ctx context.Context, docID, ownerID, signerID uuid.UUID) error
	SignDocument(ctx context.Context, docID, userID uuid.UUID, name string) (*model.Document, error)
	ListSigners(ctx context.Context, docID uuid.UUID) ([]*model.DocumentSigner, error)
}

// CreateDocumentInput holds data for creating a new document.
type CreateDocumentInput struct {
	Title        string     `json:"title"`
	Icon         string     `json:"icon"`
	Cover        string     `json:"cover"`
	OwnerID      uuid.UUID  `json:"ownerId"`
	ChatID       *uuid.UUID `json:"chatId"`
	TopicID      *uuid.UUID `json:"topicId"`
	IsStandalone bool       `json:"isStandalone"`
	TemplateID   string     `json:"templateId"`
}

// DocumentFull contains a document with all related data.
type DocumentFull struct {
	Document      model.Document              `json:"document"`
	Blocks        []*model.Block              `json:"blocks"`
	Collaborators []*DocumentCollaboratorInfo `json:"collaborators"`
	Signers       []*model.DocumentSigner     `json:"signers"`
	Tags          []string                    `json:"tags"`
	History       []*model.DocumentHistory    `json:"history"`
}

// DocumentCollaboratorInfo extends collaborator with user info.
type DocumentCollaboratorInfo struct {
	UserID  uuid.UUID              `json:"userId"`
	Name    string                 `json:"name"`
	Avatar  string                 `json:"avatar"`
	Role    model.CollaboratorRole `json:"role"`
	AddedAt time.Time              `json:"addedAt"`
}

// DocumentListItem is a summary for list views.
type DocumentListItem struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Icon        string    `json:"icon"`
	Locked      bool      `json:"locked"`
	RequireSigs bool      `json:"requireSigs"`
	OwnerID     uuid.UUID `json:"ownerId"`
	ContextType string    `json:"contextType"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type documentService struct {
	docRepo     repository.DocumentRepository
	blockRepo   repository.BlockRepository
	historyRepo repository.DocumentHistoryRepository
	userRepo    repository.UserRepository
	templateSvc TemplateService
}

// NewDocumentService creates a new document service.
func NewDocumentService(
	docRepo repository.DocumentRepository,
	blockRepo repository.BlockRepository,
	historyRepo repository.DocumentHistoryRepository,
	userRepo repository.UserRepository,
	templateSvc TemplateService,
) DocumentService {
	return &documentService{
		docRepo:     docRepo,
		blockRepo:   blockRepo,
		historyRepo: historyRepo,
		userRepo:    userRepo,
		templateSvc: templateSvc,
	}
}

func (s *documentService) Create(ctx context.Context, input CreateDocumentInput) (*DocumentFull, error) {
	if input.Title == "" {
		input.Title = "Dokumen Tanpa Judul"
	}
	if input.Icon == "" {
		input.Icon = "\U0001F4C4" // 📄
	}

	modelInput := model.CreateDocumentInput{
		Title:        input.Title,
		Icon:         input.Icon,
		OwnerID:      input.OwnerID,
		ChatID:       input.ChatID,
		TopicID:      input.TopicID,
		IsStandalone: input.IsStandalone,
	}

	// Default to standalone if no context
	if input.ChatID == nil && input.TopicID == nil {
		modelInput.IsStandalone = true
	}

	doc, err := s.docRepo.Create(ctx, modelInput)
	if err != nil {
		return nil, fmt.Errorf("create document: %w", err)
	}

	// Apply template blocks if specified
	var blocks []*model.Block
	if input.TemplateID != "" {
		templateBlocks := s.templateSvc.GetTemplateBlocks(input.TemplateID)
		for i, tb := range templateBlocks {
			blockInput := model.CreateBlockInput{
				DocumentID: doc.ID,
				Type:       model.BlockType(tb.Type),
				Content:    tb.Content,
				SortOrder:  i,
			}
			if tb.Rows != nil {
				blockInput.Rows = tb.Rows
			}
			if tb.Columns != nil {
				blockInput.Columns = tb.Columns
			}
			if tb.Emoji != "" {
				blockInput.Emoji = tb.Emoji
			}
			if tb.Color != "" {
				blockInput.Color = tb.Color
			}

			block, blockErr := s.blockRepo.Create(ctx, blockInput)
			if blockErr != nil {
				return nil, fmt.Errorf("create template block: %w", blockErr)
			}
			blocks = append(blocks, block)
		}
	}

	// Log history
	_ = s.historyRepo.Create(ctx, doc.ID, input.OwnerID, "created", "Dokumen dibuat")

	return &DocumentFull{
		Document:      *doc,
		Blocks:        blocks,
		Collaborators: []*DocumentCollaboratorInfo{},
		Signers:       []*model.DocumentSigner{},
		Tags:          []string{},
		History:       []*model.DocumentHistory{},
	}, nil
}

func (s *documentService) GetByID(ctx context.Context, docID, userID uuid.UUID) (*DocumentFull, error) {
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return nil, err
	}

	// Check access: owner, collaborator, or context member
	if doc.OwnerID != userID {
		collabs, _ := s.docRepo.ListCollaborators(ctx, docID)
		hasAccess := false
		for _, c := range collabs {
			if c.UserID == userID {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			return nil, apperror.Forbidden("anda tidak memiliki akses ke dokumen ini")
		}
	}

	blocks, _ := s.blockRepo.ListByDocument(ctx, docID)
	collabs, _ := s.docRepo.ListCollaborators(ctx, docID)
	signers, _ := s.docRepo.ListSigners(ctx, docID)
	tags, _ := s.docRepo.ListTags(ctx, docID)
	history, _ := s.historyRepo.ListByDocument(ctx, docID)

	// Enrich collaborators with user info
	collaboratorInfos := make([]*DocumentCollaboratorInfo, 0, len(collabs))
	for _, c := range collabs {
		user, userErr := s.userRepo.FindByID(ctx, c.UserID)
		name := "Unknown"
		avatar := "\U0001F464"
		if userErr == nil && user != nil {
			name = user.Name
			avatar = user.Avatar
		}
		collaboratorInfos = append(collaboratorInfos, &DocumentCollaboratorInfo{
			UserID:  c.UserID,
			Name:    name,
			Avatar:  avatar,
			Role:    c.Role,
			AddedAt: c.AddedAt,
		})
	}

	if blocks == nil {
		blocks = []*model.Block{}
	}
	if tags == nil {
		tags = []string{}
	}
	if signers == nil {
		signers = []*model.DocumentSigner{}
	}
	if history == nil {
		history = []*model.DocumentHistory{}
	}

	return &DocumentFull{
		Document:      *doc,
		Blocks:        blocks,
		Collaborators: collaboratorInfos,
		Signers:       signers,
		Tags:          tags,
		History:       history,
	}, nil
}

func (s *documentService) ListByContext(ctx context.Context, contextType string, contextID uuid.UUID) ([]*DocumentListItem, error) {
	var docs []*model.Document
	var err error

	switch contextType {
	case "chat":
		docs, err = s.docRepo.ListByChat(ctx, contextID)
	case "topic":
		docs, err = s.docRepo.ListByTopic(ctx, contextID)
	default:
		return nil, apperror.BadRequest("tipe konteks tidak valid: " + contextType)
	}

	if err != nil {
		return nil, err
	}

	return s.toListItems(docs, contextType), nil
}

func (s *documentService) ListAll(ctx context.Context, userID uuid.UUID) ([]*DocumentListItem, error) {
	docs, err := s.docRepo.ListAccessible(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.toListItems(docs, ""), nil
}

func (s *documentService) Update(ctx context.Context, docID uuid.UUID, userID uuid.UUID, input model.UpdateDocumentInput) (*model.Document, error) {
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return nil, err
	}

	if doc.Locked {
		return nil, apperror.Forbidden("dokumen terkunci, tidak dapat diubah")
	}

	// Only owner or editor can update
	if doc.OwnerID != userID {
		role, accessErr := s.getCollaboratorRole(ctx, docID, userID)
		if accessErr != nil || role != model.CollaboratorRoleEditor {
			return nil, apperror.Forbidden("anda tidak memiliki izin untuk mengubah dokumen ini")
		}
	}

	updated, err := s.docRepo.Update(ctx, docID, input)
	if err != nil {
		return nil, err
	}

	_ = s.historyRepo.Create(ctx, docID, userID, "updated", "Dokumen diperbarui")

	return updated, nil
}

func (s *documentService) Delete(ctx context.Context, docID, userID uuid.UUID) error {
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return err
	}

	if doc.OwnerID != userID {
		return apperror.Forbidden("hanya pemilik yang dapat menghapus dokumen")
	}

	if doc.Locked {
		return apperror.Forbidden("dokumen terkunci, tidak dapat dihapus")
	}

	return s.docRepo.Delete(ctx, docID)
}

func (s *documentService) Duplicate(ctx context.Context, docID, userID uuid.UUID) (*DocumentFull, error) {
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return nil, err
	}

	// Create copy
	newDoc, err := s.docRepo.Create(ctx, model.CreateDocumentInput{
		Title:        doc.Title + " (Salinan)",
		Icon:         doc.Icon,
		OwnerID:      userID,
		ChatID:       doc.ChatID,
		TopicID:      doc.TopicID,
		IsStandalone: doc.IsStandalone,
	})
	if err != nil {
		return nil, fmt.Errorf("duplicate document: %w", err)
	}

	// Copy blocks
	blocks, _ := s.blockRepo.ListByDocument(ctx, docID)
	var newBlocks []*model.Block
	for _, b := range blocks {
		newBlock, blockErr := s.blockRepo.Create(ctx, model.CreateBlockInput{
			DocumentID: newDoc.ID,
			Type:       b.Type,
			Content:    b.Content,
			Checked:    b.Checked,
			Rows:       b.Rows,
			Columns:    b.Columns,
			Language:   b.Language,
			Emoji:      b.Emoji,
			Color:      b.Color,
			SortOrder:  b.SortOrder,
		})
		if blockErr == nil {
			newBlocks = append(newBlocks, newBlock)
		}
	}

	_ = s.historyRepo.Create(ctx, newDoc.ID, userID, "created", "Duplikasi dari dokumen lain")

	return &DocumentFull{
		Document:      *newDoc,
		Blocks:        newBlocks,
		Collaborators: []*DocumentCollaboratorInfo{},
		Signers:       []*model.DocumentSigner{},
		Tags:          []string{},
		History:       []*model.DocumentHistory{},
	}, nil
}

func (s *documentService) AddCollaborator(ctx context.Context, docID, ownerID, userID uuid.UUID, role model.CollaboratorRole) error {
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return err
	}
	if doc.OwnerID != ownerID {
		return apperror.Forbidden("hanya pemilik yang dapat mengelola kolaborator")
	}
	if userID == ownerID {
		return apperror.BadRequest("pemilik tidak perlu ditambahkan sebagai kolaborator")
	}

	if err := s.docRepo.AddCollaborator(ctx, docID, userID, role); err != nil {
		return err
	}

	_ = s.historyRepo.Create(ctx, docID, ownerID, "collaborator_added", "Kolaborator ditambahkan")
	return nil
}

func (s *documentService) RemoveCollaborator(ctx context.Context, docID, ownerID, userID uuid.UUID) error {
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return err
	}
	if doc.OwnerID != ownerID {
		return apperror.Forbidden("hanya pemilik yang dapat mengelola kolaborator")
	}

	if err := s.docRepo.RemoveCollaborator(ctx, docID, userID); err != nil {
		return err
	}

	_ = s.historyRepo.Create(ctx, docID, ownerID, "collaborator_removed", "Kolaborator dihapus")
	return nil
}

func (s *documentService) UpdateCollaboratorRole(ctx context.Context, docID, ownerID, userID uuid.UUID, role model.CollaboratorRole) error {
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return err
	}
	if doc.OwnerID != ownerID {
		return apperror.Forbidden("hanya pemilik yang dapat mengelola kolaborator")
	}

	return s.docRepo.UpdateCollaboratorRole(ctx, docID, userID, role)
}

func (s *documentService) AddTag(ctx context.Context, docID uuid.UUID, tag string) error {
	if tag == "" {
		return apperror.BadRequest("tag tidak boleh kosong")
	}
	return s.docRepo.AddTag(ctx, docID, tag)
}

func (s *documentService) RemoveTag(ctx context.Context, docID uuid.UUID, tag string) error {
	return s.docRepo.RemoveTag(ctx, docID, tag)
}

func (s *documentService) GetHistory(ctx context.Context, docID uuid.UUID) ([]*model.DocumentHistory, error) {
	return s.historyRepo.ListByDocument(ctx, docID)
}

func (s *documentService) LockDocument(ctx context.Context, docID, userID uuid.UUID, mode model.LockedByType) error {
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return err
	}

	if doc.OwnerID != userID {
		return apperror.Forbidden("hanya pemilik yang dapat mengunci dokumen")
	}

	if doc.Locked {
		return apperror.BadRequest("dokumen sudah terkunci")
	}

	if mode == model.LockedBySignatures {
		// Check that there are signers configured
		signers, _ := s.docRepo.ListSigners(ctx, docID)
		if len(signers) == 0 {
			return apperror.BadRequest("tambahkan penandatangan sebelum mengunci dengan tanda tangan")
		}
	}

	if err := s.docRepo.Lock(ctx, docID, mode); err != nil {
		return err
	}

	action := "locked_manual"
	details := "Dokumen dikunci secara manual"
	if mode == model.LockedBySignatures {
		action = "locked_signatures"
		details = "Dokumen dikunci, menunggu tanda tangan"
	}

	_ = s.historyRepo.Create(ctx, docID, userID, action, details)
	return nil
}

func (s *documentService) UnlockDocument(ctx context.Context, docID, userID uuid.UUID) error {
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return err
	}

	if doc.OwnerID != userID {
		return apperror.Forbidden("hanya pemilik yang dapat membuka kunci dokumen")
	}

	if !doc.Locked {
		return apperror.BadRequest("dokumen tidak terkunci")
	}

	if doc.LockedBy != nil && *doc.LockedBy == model.LockedBySignatures {
		// Check if any signatures have been recorded
		signers, _ := s.docRepo.ListSigners(ctx, docID)
		for _, signer := range signers {
			if signer.SignedAt != nil {
				return apperror.Forbidden("dokumen yang sudah ditandatangani tidak dapat dibuka kuncinya")
			}
		}
	}

	if err := s.docRepo.Unlock(ctx, docID); err != nil {
		return err
	}

	_ = s.historyRepo.Create(ctx, docID, userID, "unlocked", "Kunci dokumen dibuka")
	return nil
}

func (s *documentService) AddSigner(ctx context.Context, docID, ownerID, signerID uuid.UUID) error {
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return err
	}

	if doc.OwnerID != ownerID {
		return apperror.Forbidden("hanya pemilik yang dapat mengelola penandatangan")
	}

	if doc.Locked {
		return apperror.Forbidden("tidak dapat menambah penandatangan pada dokumen yang terkunci")
	}

	if err := s.docRepo.AddSigner(ctx, docID, signerID); err != nil {
		return err
	}

	_ = s.historyRepo.Create(ctx, docID, ownerID, "signer_added", "Penandatangan ditambahkan")
	return nil
}

func (s *documentService) RemoveSigner(ctx context.Context, docID, ownerID, signerID uuid.UUID) error {
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return err
	}

	if doc.OwnerID != ownerID {
		return apperror.Forbidden("hanya pemilik yang dapat mengelola penandatangan")
	}

	if doc.Locked {
		return apperror.Forbidden("tidak dapat menghapus penandatangan dari dokumen yang terkunci")
	}

	if err := s.docRepo.RemoveSigner(ctx, docID, signerID); err != nil {
		return err
	}

	_ = s.historyRepo.Create(ctx, docID, ownerID, "signer_removed", "Penandatangan dihapus")
	return nil
}

func (s *documentService) SignDocument(ctx context.Context, docID, userID uuid.UUID, name string) (*model.Document, error) {
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return nil, err
	}

	if !doc.Locked || doc.LockedBy == nil || *doc.LockedBy != model.LockedBySignatures {
		return nil, apperror.BadRequest("dokumen tidak dalam mode tanda tangan")
	}

	// Verify user is a signer
	signers, err := s.docRepo.ListSigners(ctx, docID)
	if err != nil {
		return nil, err
	}

	isSigner := false
	for _, signer := range signers {
		if signer.UserID == userID {
			if signer.SignedAt != nil {
				return nil, apperror.BadRequest("anda sudah menandatangani dokumen ini")
			}
			isSigner = true
			break
		}
	}

	if !isSigner {
		return nil, apperror.Forbidden("anda bukan penandatangan dokumen ini")
	}

	if name == "" {
		// Get user name as default
		user, userErr := s.userRepo.FindByID(ctx, userID)
		if userErr == nil && user != nil {
			name = user.Name
		}
	}

	if err := s.docRepo.RecordSignature(ctx, docID, userID, name); err != nil {
		return nil, err
	}

	_ = s.historyRepo.Create(ctx, docID, userID, "signed", fmt.Sprintf("Ditandatangani oleh %s", name))

	// Re-fetch the document to return updated state
	updatedDoc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return nil, err
	}

	return updatedDoc, nil
}

func (s *documentService) ListSigners(ctx context.Context, docID uuid.UUID) ([]*model.DocumentSigner, error) {
	return s.docRepo.ListSigners(ctx, docID)
}

// Helper methods

func (s *documentService) getCollaboratorRole(ctx context.Context, docID, userID uuid.UUID) (model.CollaboratorRole, error) {
	collabs, err := s.docRepo.ListCollaborators(ctx, docID)
	if err != nil {
		return "", err
	}
	for _, c := range collabs {
		if c.UserID == userID {
			return c.Role, nil
		}
	}
	return "", apperror.Forbidden("bukan kolaborator")
}

func (s *documentService) toListItems(docs []*model.Document, contextType string) []*DocumentListItem {
	items := make([]*DocumentListItem, 0, len(docs))
	for _, doc := range docs {
		ct := contextType
		if ct == "" {
			if doc.ChatID != nil {
				ct = "chat"
			} else if doc.TopicID != nil {
				ct = "topic"
			} else {
				ct = "standalone"
			}
		}

		items = append(items, &DocumentListItem{
			ID:          doc.ID,
			Title:       doc.Title,
			Icon:        doc.Icon,
			Locked:      doc.Locked,
			RequireSigs: doc.RequireSigs,
			OwnerID:     doc.OwnerID,
			ContextType: ct,
			UpdatedAt:   doc.UpdatedAt,
		})
	}
	return items
}
