package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/pkg/apperror"
)

// -- Mock implementations --

type mockDocumentRepo struct {
	docs          map[uuid.UUID]*model.Document
	collaborators map[uuid.UUID][]*model.DocumentCollaborator
	signers       map[uuid.UUID][]*model.DocumentSigner
	tags          map[uuid.UUID][]string
}

func newMockDocumentRepo() *mockDocumentRepo {
	return &mockDocumentRepo{
		docs:          make(map[uuid.UUID]*model.Document),
		collaborators: make(map[uuid.UUID][]*model.DocumentCollaborator),
		signers:       make(map[uuid.UUID][]*model.DocumentSigner),
		tags:          make(map[uuid.UUID][]string),
	}
}

func (m *mockDocumentRepo) Create(_ context.Context, input model.CreateDocumentInput) (*model.Document, error) {
	doc := &model.Document{
		ID:           uuid.New(),
		Title:        input.Title,
		Icon:         input.Icon,
		OwnerID:      input.OwnerID,
		ChatID:       input.ChatID,
		TopicID:      input.TopicID,
		IsStandalone: input.IsStandalone,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	m.docs[doc.ID] = doc
	return doc, nil
}

func (m *mockDocumentRepo) FindByID(_ context.Context, id uuid.UUID) (*model.Document, error) {
	doc, ok := m.docs[id]
	if !ok {
		return nil, apperror.NotFound("document", id.String())
	}
	return doc, nil
}

func (m *mockDocumentRepo) ListByChat(_ context.Context, chatID uuid.UUID) ([]*model.Document, error) {
	var result []*model.Document
	for _, doc := range m.docs {
		if doc.ChatID != nil && *doc.ChatID == chatID {
			result = append(result, doc)
		}
	}
	return result, nil
}

func (m *mockDocumentRepo) ListByTopic(_ context.Context, topicID uuid.UUID) ([]*model.Document, error) {
	var result []*model.Document
	for _, doc := range m.docs {
		if doc.TopicID != nil && *doc.TopicID == topicID {
			result = append(result, doc)
		}
	}
	return result, nil
}

func (m *mockDocumentRepo) ListByOwner(_ context.Context, ownerID uuid.UUID) ([]*model.Document, error) {
	var result []*model.Document
	for _, doc := range m.docs {
		if doc.OwnerID == ownerID {
			result = append(result, doc)
		}
	}
	return result, nil
}

func (m *mockDocumentRepo) ListByTag(_ context.Context, _ string) ([]*model.Document, error) {
	return nil, nil
}

func (m *mockDocumentRepo) ListAccessible(_ context.Context, userID uuid.UUID) ([]*model.Document, error) {
	var result []*model.Document
	for _, doc := range m.docs {
		if doc.OwnerID == userID {
			result = append(result, doc)
			continue
		}
		for _, c := range m.collaborators[doc.ID] {
			if c.UserID == userID {
				result = append(result, doc)
				break
			}
		}
	}
	return result, nil
}

func (m *mockDocumentRepo) ListCollaborators(_ context.Context, docID uuid.UUID) ([]*model.DocumentCollaborator, error) {
	return m.collaborators[docID], nil
}

func (m *mockDocumentRepo) ListSigners(_ context.Context, docID uuid.UUID) ([]*model.DocumentSigner, error) {
	return m.signers[docID], nil
}

func (m *mockDocumentRepo) ListTags(_ context.Context, docID uuid.UUID) ([]string, error) {
	return m.tags[docID], nil
}

func (m *mockDocumentRepo) AddCollaborator(_ context.Context, docID, userID uuid.UUID, role model.CollaboratorRole) error {
	m.collaborators[docID] = append(m.collaborators[docID], &model.DocumentCollaborator{
		DocumentID: docID,
		UserID:     userID,
		Role:       role,
		AddedAt:    time.Now(),
	})
	return nil
}

func (m *mockDocumentRepo) RemoveCollaborator(_ context.Context, docID, userID uuid.UUID) error {
	collabs := m.collaborators[docID]
	for i, c := range collabs {
		if c.UserID == userID {
			m.collaborators[docID] = append(collabs[:i], collabs[i+1:]...)
			return nil
		}
	}
	return apperror.NotFound("collaborator", userID.String())
}

func (m *mockDocumentRepo) UpdateCollaboratorRole(_ context.Context, docID, userID uuid.UUID, role model.CollaboratorRole) error {
	for _, c := range m.collaborators[docID] {
		if c.UserID == userID {
			c.Role = role
			return nil
		}
	}
	return apperror.NotFound("collaborator", userID.String())
}

func (m *mockDocumentRepo) AddSigner(_ context.Context, docID, userID uuid.UUID) error {
	m.signers[docID] = append(m.signers[docID], &model.DocumentSigner{
		DocumentID: docID,
		UserID:     userID,
	})
	return nil
}
func (m *mockDocumentRepo) RemoveSigner(_ context.Context, docID, userID uuid.UUID) error {
	signers := m.signers[docID]
	for i, s := range signers {
		if s.UserID == userID {
			m.signers[docID] = append(signers[:i], signers[i+1:]...)
			return nil
		}
	}
	return apperror.NotFound("signer", userID.String())
}
func (m *mockDocumentRepo) RecordSignature(_ context.Context, docID, userID uuid.UUID, name string) error {
	for _, s := range m.signers[docID] {
		if s.UserID == userID {
			now := time.Now()
			s.SignedAt = &now
			s.SignerName = name
			return nil
		}
	}
	return apperror.NotFound("signer", userID.String())
}
func (m *mockDocumentRepo) Lock(_ context.Context, docID uuid.UUID, lockedBy model.LockedByType) error {
	doc, ok := m.docs[docID]
	if !ok {
		return apperror.NotFound("document", docID.String())
	}
	doc.Locked = true
	now := time.Now()
	doc.LockedAt = &now
	doc.LockedBy = &lockedBy
	return nil
}
func (m *mockDocumentRepo) Unlock(_ context.Context, docID uuid.UUID) error {
	doc, ok := m.docs[docID]
	if !ok {
		return apperror.NotFound("document", docID.String())
	}
	doc.Locked = false
	doc.LockedAt = nil
	doc.LockedBy = nil
	return nil
}

func (m *mockDocumentRepo) AddTag(_ context.Context, docID uuid.UUID, tag string) error {
	m.tags[docID] = append(m.tags[docID], tag)
	return nil
}

func (m *mockDocumentRepo) RemoveTag(_ context.Context, docID uuid.UUID, tag string) error {
	tags := m.tags[docID]
	for i, t := range tags {
		if t == tag {
			m.tags[docID] = append(tags[:i], tags[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockDocumentRepo) Update(_ context.Context, id uuid.UUID, input model.UpdateDocumentInput) (*model.Document, error) {
	doc, ok := m.docs[id]
	if !ok {
		return nil, apperror.NotFound("document", id.String())
	}
	if input.Title != nil {
		doc.Title = *input.Title
	}
	if input.Icon != nil {
		doc.Icon = *input.Icon
	}
	doc.UpdatedAt = time.Now()
	return doc, nil
}

func (m *mockDocumentRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := m.docs[id]; !ok {
		return apperror.NotFound("document", id.String())
	}
	delete(m.docs, id)
	return nil
}

type mockBlockRepo struct {
	blocks     map[uuid.UUID]*model.Block
	listErr    error
	updateErr  error
	deleteErr  error
	reorderErr error
	createErr  error
}

func newMockBlockRepo() *mockBlockRepo {
	return &mockBlockRepo{blocks: make(map[uuid.UUID]*model.Block)}
}

func (m *mockBlockRepo) Create(_ context.Context, input model.CreateBlockInput) (*model.Block, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	b := &model.Block{
		ID:         uuid.New(),
		DocumentID: input.DocumentID,
		Type:       input.Type,
		Content:    input.Content,
		SortOrder:  input.SortOrder,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	m.blocks[b.ID] = b
	return b, nil
}

func (m *mockBlockRepo) FindByID(_ context.Context, id uuid.UUID) (*model.Block, error) {
	b, ok := m.blocks[id]
	if !ok {
		return nil, apperror.NotFound("block", id.String())
	}
	return b, nil
}

func (m *mockBlockRepo) ListByDocument(_ context.Context, docID uuid.UUID) ([]*model.Block, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*model.Block
	for _, b := range m.blocks {
		if b.DocumentID == docID {
			result = append(result, b)
		}
	}
	return result, nil
}

func (m *mockBlockRepo) Update(_ context.Context, id uuid.UUID, input model.UpdateBlockInput) (*model.Block, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	b, ok := m.blocks[id]
	if !ok {
		return nil, apperror.NotFound("block", id.String())
	}
	if input.Content != nil {
		b.Content = *input.Content
	}
	b.UpdatedAt = time.Now()
	return b, nil
}

func (m *mockBlockRepo) Reorder(_ context.Context, _ uuid.UUID, _ []uuid.UUID) error {
	if m.reorderErr != nil {
		return m.reorderErr
	}
	return nil
}

func (m *mockBlockRepo) Delete(_ context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.blocks[id]; !ok {
		return apperror.NotFound("block", id.String())
	}
	delete(m.blocks, id)
	return nil
}

type mockDocHistoryRepo struct {
	entries []*model.DocumentHistory
}

func (m *mockDocHistoryRepo) Create(_ context.Context, docID, userID uuid.UUID, action, details string) error {
	m.entries = append(m.entries, &model.DocumentHistory{
		ID:         uuid.New(),
		DocumentID: docID,
		UserID:     userID,
		Action:     action,
		Details:    details,
		CreatedAt:  time.Now(),
	})
	return nil
}

func (m *mockDocHistoryRepo) ListByDocument(_ context.Context, docID uuid.UUID) ([]*model.DocumentHistory, error) {
	var result []*model.DocumentHistory
	for _, h := range m.entries {
		if h.DocumentID == docID {
			result = append(result, h)
		}
	}
	return result, nil
}

type docTestUserRepo struct {
	users map[uuid.UUID]*model.User
}

func (m *docTestUserRepo) Create(_ context.Context, _ model.CreateUserInput) (*model.User, error) {
	return nil, nil
}
func (m *docTestUserRepo) FindByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, apperror.NotFound("user", id.String())
	}
	return u, nil
}
func (m *docTestUserRepo) FindByPhone(_ context.Context, _ string) (*model.User, error) {
	return nil, nil
}
func (m *docTestUserRepo) FindByPhones(_ context.Context, _ []string) ([]*model.User, error) {
	return nil, nil
}
func (m *docTestUserRepo) Update(_ context.Context, _ uuid.UUID, _ model.UpdateUserInput) (*model.User, error) {
	return nil, nil
}
func (m *docTestUserRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }
func (m *docTestUserRepo) FindByPhoneHashes(_ context.Context, _ []string) ([]*model.User, error) {
	return nil, nil
}
func (m *docTestUserRepo) UpdatePhoneHash(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (m *docTestUserRepo) UpdateLastSeen(_ context.Context, _ uuid.UUID) error {
	return nil
}

// -- Tests --

func TestDocumentService_Create(t *testing.T) {
	docRepo := newMockDocumentRepo()
	blockRepo := newMockBlockRepo()
	historyRepo := &mockDocHistoryRepo{}
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	templateSvc := NewTemplateService()

	svc := NewDocumentService(docRepo, blockRepo, historyRepo, userRepo, templateSvc, nil)
	ctx := context.Background()
	ownerID := uuid.New()

	t.Run("create standalone doc", func(t *testing.T) {
		result, err := svc.Create(ctx, CreateDocumentInput{
			Title:        "Test Doc",
			OwnerID:      ownerID,
			IsStandalone: true,
		})
		require.NoError(t, err)
		assert.Equal(t, "Test Doc", result.Document.Title)
		assert.True(t, result.Document.IsStandalone)
		assert.Equal(t, ownerID, result.Document.OwnerID)
	})

	t.Run("create with default title", func(t *testing.T) {
		result, err := svc.Create(ctx, CreateDocumentInput{
			OwnerID: ownerID,
		})
		require.NoError(t, err)
		assert.Equal(t, "Dokumen Tanpa Judul", result.Document.Title)
	})

	t.Run("create with template", func(t *testing.T) {
		result, err := svc.Create(ctx, CreateDocumentInput{
			Title:      "Rapat",
			OwnerID:    ownerID,
			TemplateID: "notulen-rapat",
		})
		require.NoError(t, err)
		assert.True(t, len(result.Blocks) > 0)
	})

	t.Run("create in chat context", func(t *testing.T) {
		chatID := uuid.New()
		result, err := svc.Create(ctx, CreateDocumentInput{
			Title:   "Chat Doc",
			OwnerID: ownerID,
			ChatID:  &chatID,
		})
		require.NoError(t, err)
		assert.NotNil(t, result.Document.ChatID)
		assert.Equal(t, chatID, *result.Document.ChatID)
	})

	t.Run("history logged on create", func(t *testing.T) {
		result, err := svc.Create(ctx, CreateDocumentInput{
			Title:   "History Doc",
			OwnerID: ownerID,
		})
		require.NoError(t, err)

		history, _ := historyRepo.ListByDocument(ctx, result.Document.ID)
		assert.True(t, len(history) > 0)
		assert.Equal(t, "created", history[0].Action)
	})
}

func TestDocumentService_GetByID(t *testing.T) {
	docRepo := newMockDocumentRepo()
	blockRepo := newMockBlockRepo()
	historyRepo := &mockDocHistoryRepo{}
	ownerID := uuid.New()
	userRepo := &docTestUserRepo{users: map[uuid.UUID]*model.User{
		ownerID: {ID: ownerID, Name: "Owner", Avatar: "O"},
	}}
	templateSvc := NewTemplateService()
	svc := NewDocumentService(docRepo, blockRepo, historyRepo, userRepo, templateSvc, nil)
	ctx := context.Background()

	t.Run("owner can access", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "Private", OwnerID: ownerID})
		result, err := svc.GetByID(ctx, doc.Document.ID, ownerID)
		require.NoError(t, err)
		assert.Equal(t, "Private", result.Document.Title)
	})

	t.Run("non-owner without collab cannot access", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "Private2", OwnerID: ownerID})
		otherUser := uuid.New()
		_, err := svc.GetByID(ctx, doc.Document.ID, otherUser)
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "FORBIDDEN", appErr.Code)
	})

	t.Run("collaborator can access", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "Collab", OwnerID: ownerID})
		collabUser := uuid.New()
		_ = docRepo.AddCollaborator(ctx, doc.Document.ID, collabUser, model.CollaboratorRoleViewer)

		result, err := svc.GetByID(ctx, doc.Document.ID, collabUser)
		require.NoError(t, err)
		assert.Equal(t, "Collab", result.Document.Title)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetByID(ctx, uuid.New(), ownerID)
		require.Error(t, err)
	})
}

func TestDocumentService_Update(t *testing.T) {
	docRepo := newMockDocumentRepo()
	blockRepo := newMockBlockRepo()
	historyRepo := &mockDocHistoryRepo{}
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	templateSvc := NewTemplateService()
	svc := NewDocumentService(docRepo, blockRepo, historyRepo, userRepo, templateSvc, nil)
	ctx := context.Background()
	ownerID := uuid.New()

	t.Run("owner can update", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "Original", OwnerID: ownerID})
		newTitle := "Updated"
		result, err := svc.Update(ctx, doc.Document.ID, ownerID, model.UpdateDocumentInput{Title: &newTitle})
		require.NoError(t, err)
		assert.Equal(t, "Updated", result.Title)
	})

	t.Run("locked doc cannot be updated", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "Locked", OwnerID: ownerID})
		docRepo.docs[doc.Document.ID].Locked = true

		newTitle := "Try Update"
		_, err := svc.Update(ctx, doc.Document.ID, ownerID, model.UpdateDocumentInput{Title: &newTitle})
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "FORBIDDEN", appErr.Code)
	})

	t.Run("non-owner/non-editor cannot update", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "Mine", OwnerID: ownerID})
		otherUser := uuid.New()

		newTitle := "Hacked"
		_, err := svc.Update(ctx, doc.Document.ID, otherUser, model.UpdateDocumentInput{Title: &newTitle})
		require.Error(t, err)
	})
}

func TestDocumentService_Delete(t *testing.T) {
	docRepo := newMockDocumentRepo()
	blockRepo := newMockBlockRepo()
	historyRepo := &mockDocHistoryRepo{}
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	templateSvc := NewTemplateService()
	svc := NewDocumentService(docRepo, blockRepo, historyRepo, userRepo, templateSvc, nil)
	ctx := context.Background()
	ownerID := uuid.New()

	t.Run("owner can delete", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "DeleteMe", OwnerID: ownerID})
		err := svc.Delete(ctx, doc.Document.ID, ownerID)
		require.NoError(t, err)
	})

	t.Run("non-owner cannot delete", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "CantDelete", OwnerID: ownerID})
		err := svc.Delete(ctx, doc.Document.ID, uuid.New())
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "FORBIDDEN", appErr.Code)
	})

	t.Run("locked doc cannot be deleted", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "LockedDoc", OwnerID: ownerID})
		docRepo.docs[doc.Document.ID].Locked = true

		err := svc.Delete(ctx, doc.Document.ID, ownerID)
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "FORBIDDEN", appErr.Code)
	})
}

func TestDocumentService_Duplicate(t *testing.T) {
	docRepo := newMockDocumentRepo()
	blockRepo := newMockBlockRepo()
	historyRepo := &mockDocHistoryRepo{}
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	templateSvc := NewTemplateService()
	svc := NewDocumentService(docRepo, blockRepo, historyRepo, userRepo, templateSvc, nil)
	ctx := context.Background()
	ownerID := uuid.New()

	t.Run("duplicate doc", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "Original", OwnerID: ownerID})
		dup, err := svc.Duplicate(ctx, doc.Document.ID, ownerID)
		require.NoError(t, err)
		assert.Equal(t, "Original (Salinan)", dup.Document.Title)
		assert.NotEqual(t, doc.Document.ID, dup.Document.ID)
	})
}

func TestDocumentService_Collaborators(t *testing.T) {
	docRepo := newMockDocumentRepo()
	blockRepo := newMockBlockRepo()
	historyRepo := &mockDocHistoryRepo{}
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	templateSvc := NewTemplateService()
	svc := NewDocumentService(docRepo, blockRepo, historyRepo, userRepo, templateSvc, nil)
	ctx := context.Background()
	ownerID := uuid.New()
	collabID := uuid.New()

	t.Run("owner can add collaborator", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "Collab", OwnerID: ownerID})
		err := svc.AddCollaborator(ctx, doc.Document.ID, ownerID, collabID, model.CollaboratorRoleEditor)
		require.NoError(t, err)
	})

	t.Run("non-owner cannot add collaborator", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "Collab2", OwnerID: ownerID})
		err := svc.AddCollaborator(ctx, doc.Document.ID, uuid.New(), collabID, model.CollaboratorRoleEditor)
		require.Error(t, err)
	})

	t.Run("owner cannot add self as collaborator", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "Self", OwnerID: ownerID})
		err := svc.AddCollaborator(ctx, doc.Document.ID, ownerID, ownerID, model.CollaboratorRoleEditor)
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "BAD_REQUEST", appErr.Code)
	})

	t.Run("owner can remove collaborator", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "Collab3", OwnerID: ownerID})
		_ = svc.AddCollaborator(ctx, doc.Document.ID, ownerID, collabID, model.CollaboratorRoleEditor)
		err := svc.RemoveCollaborator(ctx, doc.Document.ID, ownerID, collabID)
		require.NoError(t, err)
	})

	t.Run("owner can update collaborator role", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "RoleUpdate", OwnerID: ownerID})
		_ = svc.AddCollaborator(ctx, doc.Document.ID, ownerID, collabID, model.CollaboratorRoleEditor)
		err := svc.UpdateCollaboratorRole(ctx, doc.Document.ID, ownerID, collabID, model.CollaboratorRoleViewer)
		require.NoError(t, err)
	})
}

func TestDocumentService_Tags(t *testing.T) {
	docRepo := newMockDocumentRepo()
	blockRepo := newMockBlockRepo()
	historyRepo := &mockDocHistoryRepo{}
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	templateSvc := NewTemplateService()
	svc := NewDocumentService(docRepo, blockRepo, historyRepo, userRepo, templateSvc, nil)
	ctx := context.Background()
	ownerID := uuid.New()

	t.Run("add tag", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "Tagged", OwnerID: ownerID})
		err := svc.AddTag(ctx, doc.Document.ID, "penting")
		require.NoError(t, err)
	})

	t.Run("empty tag rejected", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "NoTag", OwnerID: ownerID})
		err := svc.AddTag(ctx, doc.Document.ID, "")
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "BAD_REQUEST", appErr.Code)
	})

	t.Run("remove tag", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "Tagged2", OwnerID: ownerID})
		_ = svc.AddTag(ctx, doc.Document.ID, "hapus")
		err := svc.RemoveTag(ctx, doc.Document.ID, "hapus")
		require.NoError(t, err)
	})
}

func TestDocumentService_ListAll(t *testing.T) {
	docRepo := newMockDocumentRepo()
	blockRepo := newMockBlockRepo()
	historyRepo := &mockDocHistoryRepo{}
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	templateSvc := NewTemplateService()
	svc := NewDocumentService(docRepo, blockRepo, historyRepo, userRepo, templateSvc, nil)
	ctx := context.Background()

	ownerID := uuid.New()
	_, _ = svc.Create(ctx, CreateDocumentInput{Title: "Doc1", OwnerID: ownerID})
	_, _ = svc.Create(ctx, CreateDocumentInput{Title: "Doc2", OwnerID: ownerID})
	_, _ = svc.Create(ctx, CreateDocumentInput{Title: "OtherDoc", OwnerID: uuid.New()})

	items, err := svc.ListAll(ctx, ownerID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(items))
}

func TestDocumentService_ListByContext(t *testing.T) {
	docRepo := newMockDocumentRepo()
	blockRepo := newMockBlockRepo()
	historyRepo := &mockDocHistoryRepo{}
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	templateSvc := NewTemplateService()
	svc := NewDocumentService(docRepo, blockRepo, historyRepo, userRepo, templateSvc, nil)
	ctx := context.Background()
	ownerID := uuid.New()

	t.Run("invalid context type", func(t *testing.T) {
		_, err := svc.ListByContext(ctx, "invalid", uuid.New())
		require.Error(t, err)
	})

	t.Run("list by chat", func(t *testing.T) {
		chatID := uuid.New()
		_, _ = svc.Create(ctx, CreateDocumentInput{Title: "ChatDoc", OwnerID: ownerID, ChatID: &chatID})

		items, err := svc.ListByContext(ctx, "chat", chatID)
		require.NoError(t, err)
		assert.True(t, len(items) > 0)
	})
}

func TestDocumentService_LockUnlock(t *testing.T) {
	docRepo := newMockDocumentRepo()
	blockRepo := newMockBlockRepo()
	historyRepo := &mockDocHistoryRepo{}
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	templateSvc := NewTemplateService()
	svc := NewDocumentService(docRepo, blockRepo, historyRepo, userRepo, templateSvc, nil)
	ctx := context.Background()
	ownerID := uuid.New()

	t.Run("owner can lock manually", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "LockMe", OwnerID: ownerID})
		err := svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedByManual)
		require.NoError(t, err)

		fetched, _ := docRepo.FindByID(ctx, doc.Document.ID)
		assert.True(t, fetched.Locked)
		assert.NotNil(t, fetched.LockedBy)
		assert.Equal(t, model.LockedByManual, *fetched.LockedBy)
	})

	t.Run("non-owner cannot lock", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "NoLock", OwnerID: ownerID})
		err := svc.LockDocument(ctx, doc.Document.ID, uuid.New(), model.LockedByManual)
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "FORBIDDEN", appErr.Code)
	})

	t.Run("cannot lock already locked doc", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "AlreadyLocked", OwnerID: ownerID})
		_ = svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedByManual)
		err := svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedByManual)
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "BAD_REQUEST", appErr.Code)
	})

	t.Run("owner can unlock manual lock", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "UnlockMe", OwnerID: ownerID})
		_ = svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedByManual)
		err := svc.UnlockDocument(ctx, doc.Document.ID, ownerID)
		require.NoError(t, err)

		fetched, _ := docRepo.FindByID(ctx, doc.Document.ID)
		assert.False(t, fetched.Locked)
	})

	t.Run("cannot unlock doc that is not locked", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "NotLocked", OwnerID: ownerID})
		err := svc.UnlockDocument(ctx, doc.Document.ID, ownerID)
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "BAD_REQUEST", appErr.Code)
	})

	t.Run("lock with signatures requires signers", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "NoSigners", OwnerID: ownerID})
		err := svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedBySignatures)
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "BAD_REQUEST", appErr.Code)
	})
}

func TestDocumentService_Signers(t *testing.T) {
	docRepo := newMockDocumentRepo()
	blockRepo := newMockBlockRepo()
	historyRepo := &mockDocHistoryRepo{}
	ownerID := uuid.New()
	signerID := uuid.New()
	userRepo := &docTestUserRepo{users: map[uuid.UUID]*model.User{
		ownerID:  {ID: ownerID, Name: "Owner"},
		signerID: {ID: signerID, Name: "Signer"},
	}}
	templateSvc := NewTemplateService()
	svc := NewDocumentService(docRepo, blockRepo, historyRepo, userRepo, templateSvc, nil)
	ctx := context.Background()

	t.Run("owner can add signer", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "SignDoc", OwnerID: ownerID})
		err := svc.AddSigner(ctx, doc.Document.ID, ownerID, signerID)
		require.NoError(t, err)

		signers, _ := svc.ListSigners(ctx, doc.Document.ID)
		assert.Equal(t, 1, len(signers))
	})

	t.Run("non-owner cannot add signer", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "SignDoc2", OwnerID: ownerID})
		err := svc.AddSigner(ctx, doc.Document.ID, uuid.New(), signerID)
		require.Error(t, err)
	})

	t.Run("owner can remove signer", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "SignDoc3", OwnerID: ownerID})
		_ = svc.AddSigner(ctx, doc.Document.ID, ownerID, signerID)
		err := svc.RemoveSigner(ctx, doc.Document.ID, ownerID, signerID)
		require.NoError(t, err)
	})

	t.Run("cannot add signer to locked doc", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "LockedSign", OwnerID: ownerID})
		_ = svc.AddSigner(ctx, doc.Document.ID, ownerID, signerID)
		_ = svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedBySignatures)

		newSigner := uuid.New()
		err := svc.AddSigner(ctx, doc.Document.ID, ownerID, newSigner)
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "FORBIDDEN", appErr.Code)
	})

	t.Run("signer can sign document", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "ToSign", OwnerID: ownerID})
		_ = svc.AddSigner(ctx, doc.Document.ID, ownerID, signerID)
		_ = svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedBySignatures)

		result, err := svc.SignDocument(ctx, doc.Document.ID, signerID, "Test Signer")
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("non-signer cannot sign", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "CantSign", OwnerID: ownerID})
		_ = svc.AddSigner(ctx, doc.Document.ID, ownerID, signerID)
		_ = svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedBySignatures)

		_, err := svc.SignDocument(ctx, doc.Document.ID, uuid.New(), "Nobody")
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "FORBIDDEN", appErr.Code)
	})

	t.Run("cannot sign unlocked doc", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "NotSigMode", OwnerID: ownerID})
		_, err := svc.SignDocument(ctx, doc.Document.ID, signerID, "Test")
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "BAD_REQUEST", appErr.Code)
	})

	t.Run("cannot sign twice", func(t *testing.T) {
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "SignTwice", OwnerID: ownerID})
		_ = svc.AddSigner(ctx, doc.Document.ID, ownerID, signerID)
		_ = svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedBySignatures)

		_, _ = svc.SignDocument(ctx, doc.Document.ID, signerID, "First")
		_, err := svc.SignDocument(ctx, doc.Document.ID, signerID, "Second")
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "BAD_REQUEST", appErr.Code)
	})
}

func TestDocumentService_GetHistory(t *testing.T) {
	docRepo := newMockDocumentRepo()
	blockRepo := newMockBlockRepo()
	historyRepo := &mockDocHistoryRepo{}
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	templateSvc := NewTemplateService()

	svc := NewDocumentService(docRepo, blockRepo, historyRepo, userRepo, templateSvc, nil)
	ctx := context.Background()
	ownerID := uuid.New()

	t.Run("returns history entries", func(t *testing.T) {
		result, err := svc.Create(ctx, CreateDocumentInput{Title: "HistDoc", OwnerID: ownerID})
		require.NoError(t, err)

		history, err := svc.GetHistory(ctx, result.Document.ID)
		require.NoError(t, err)
		assert.True(t, len(history) > 0)
		assert.Equal(t, "created", history[0].Action)
	})

	t.Run("no history for unknown doc", func(t *testing.T) {
		history, err := svc.GetHistory(ctx, uuid.New())
		require.NoError(t, err)
		assert.Empty(t, history)
	})
}

func TestDocumentService_RemoveTag(t *testing.T) {
	docRepo := newMockDocumentRepo()
	blockRepo := newMockBlockRepo()
	historyRepo := &mockDocHistoryRepo{}
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	templateSvc := NewTemplateService()

	svc := NewDocumentService(docRepo, blockRepo, historyRepo, userRepo, templateSvc, nil)
	ctx := context.Background()
	ownerID := uuid.New()

	result, err := svc.Create(ctx, CreateDocumentInput{Title: "TagDoc", OwnerID: ownerID})
	require.NoError(t, err)

	_ = svc.AddTag(ctx, result.Document.ID, "important")

	t.Run("remove existing tag", func(t *testing.T) {
		err := svc.RemoveTag(ctx, result.Document.ID, "important")
		assert.NoError(t, err)
	})
}

// --- Additional error-path and coverage tests ---

type mockNotifSvc struct{}

func (m *mockNotifSvc) SendToUser(_ context.Context, _ uuid.UUID, _ model.Notification) error {
	return nil
}
func (m *mockNotifSvc) SendToUsers(_ context.Context, _ []uuid.UUID, _ model.Notification) error {
	return nil
}
func (m *mockNotifSvc) SendToChat(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ model.Notification) error {
	return nil
}
func (m *mockNotifSvc) RegisterDevice(_ context.Context, _ uuid.UUID, _, _ string) error { return nil }
func (m *mockNotifSvc) RemoveDevice(_ context.Context, _ uuid.UUID, _ string) error      { return nil }
func (m *mockNotifSvc) UnregisterDevice(_ context.Context, _ uuid.UUID, _ string) error  { return nil }

func TestDocumentService_LockDocument_Errors(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	collabID := uuid.New()
	userRepo := &docTestUserRepo{users: map[uuid.UUID]*model.User{
		ownerID:  {ID: ownerID, Name: "Owner"},
		collabID: {ID: collabID, Name: "Collab"},
	}}

	t.Run("not found", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		err := svc.LockDocument(ctx, uuid.New(), ownerID, model.LockedByManual)
		require.Error(t, err)
	})

	t.Run("lock with signatures mode", func(t *testing.T) {
		docRepo := newMockDocumentRepo()
		svc := NewDocumentService(docRepo, newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), &mockNotifSvc{})

		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "SigLock", OwnerID: ownerID})
		_ = svc.AddSigner(ctx, doc.Document.ID, ownerID, collabID)
		err := svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedBySignatures)
		require.NoError(t, err)

		fetched, _ := docRepo.FindByID(ctx, doc.Document.ID)
		assert.True(t, fetched.Locked)
	})

	t.Run("lock with notif and collaborators", func(t *testing.T) {
		docRepo := newMockDocumentRepo()
		svc := NewDocumentService(docRepo, newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), &mockNotifSvc{})

		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "NotifLock", OwnerID: ownerID})
		_ = svc.AddCollaborator(ctx, doc.Document.ID, ownerID, collabID, model.CollaboratorRoleEditor)
		err := svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedByManual)
		require.NoError(t, err)
	})
}

func TestDocumentService_UnlockDocument_Errors(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	signerID := uuid.New()
	userRepo := &docTestUserRepo{users: map[uuid.UUID]*model.User{
		ownerID:  {ID: ownerID, Name: "Owner"},
		signerID: {ID: signerID, Name: "Signer"},
	}}

	t.Run("not found", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		err := svc.UnlockDocument(ctx, uuid.New(), ownerID)
		require.Error(t, err)
	})

	t.Run("non-owner cannot unlock", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "Locked", OwnerID: ownerID})
		_ = svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedByManual)
		err := svc.UnlockDocument(ctx, doc.Document.ID, uuid.New())
		require.Error(t, err)
	})

	t.Run("cannot unlock signed doc", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "Signed", OwnerID: ownerID})
		_ = svc.AddSigner(ctx, doc.Document.ID, ownerID, signerID)
		_ = svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedBySignatures)
		_, _ = svc.SignDocument(ctx, doc.Document.ID, signerID, "Signer")

		err := svc.UnlockDocument(ctx, doc.Document.ID, ownerID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sudah ditandatangani")
	})

	t.Run("can unlock sig-locked unsigned doc", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "SigNoSign", OwnerID: ownerID})
		_ = svc.AddSigner(ctx, doc.Document.ID, ownerID, signerID)
		_ = svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedBySignatures)

		err := svc.UnlockDocument(ctx, doc.Document.ID, ownerID)
		require.NoError(t, err)
	})
}

func TestDocumentService_AddSigner_Errors(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	signerID := uuid.New()
	userRepo := &docTestUserRepo{users: map[uuid.UUID]*model.User{
		ownerID:  {ID: ownerID, Name: "Owner"},
		signerID: {ID: signerID, Name: "Signer"},
	}}

	t.Run("not found", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		err := svc.AddSigner(ctx, uuid.New(), ownerID, signerID)
		require.Error(t, err)
	})

	t.Run("add signer with notif", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), &mockNotifSvc{})
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "NotifSign", OwnerID: ownerID})
		err := svc.AddSigner(ctx, doc.Document.ID, ownerID, signerID)
		require.NoError(t, err)
	})
}

func TestDocumentService_RemoveSigner_Errors(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	signerID := uuid.New()
	userRepo := &docTestUserRepo{users: map[uuid.UUID]*model.User{
		ownerID:  {ID: ownerID, Name: "Owner"},
		signerID: {ID: signerID, Name: "Signer"},
	}}

	t.Run("not found doc", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		err := svc.RemoveSigner(ctx, uuid.New(), ownerID, signerID)
		require.Error(t, err)
	})

	t.Run("non-owner cannot remove signer", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "RS", OwnerID: ownerID})
		_ = svc.AddSigner(ctx, doc.Document.ID, ownerID, signerID)
		err := svc.RemoveSigner(ctx, doc.Document.ID, uuid.New(), signerID)
		require.Error(t, err)
	})

	t.Run("cannot remove from locked doc", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "LockedRS", OwnerID: ownerID})
		_ = svc.AddSigner(ctx, doc.Document.ID, ownerID, signerID)
		_ = svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedBySignatures)
		err := svc.RemoveSigner(ctx, doc.Document.ID, ownerID, signerID)
		require.Error(t, err)
	})
}

func TestDocumentService_RemoveCollaborator_Errors(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	collabID := uuid.New()
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}

	t.Run("non-owner cannot remove", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "RC", OwnerID: ownerID})
		_ = svc.AddCollaborator(ctx, doc.Document.ID, ownerID, collabID, model.CollaboratorRoleEditor)
		err := svc.RemoveCollaborator(ctx, doc.Document.ID, uuid.New(), collabID)
		require.Error(t, err)
	})

	t.Run("not found doc", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		err := svc.RemoveCollaborator(ctx, uuid.New(), ownerID, collabID)
		require.Error(t, err)
	})
}

func TestDocumentService_UpdateCollaboratorRole_Errors(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	collabID := uuid.New()
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}

	t.Run("non-owner cannot update role", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "UCR", OwnerID: ownerID})
		_ = svc.AddCollaborator(ctx, doc.Document.ID, ownerID, collabID, model.CollaboratorRoleEditor)
		err := svc.UpdateCollaboratorRole(ctx, doc.Document.ID, uuid.New(), collabID, model.CollaboratorRoleViewer)
		require.Error(t, err)
	})

	t.Run("not found doc", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		err := svc.UpdateCollaboratorRole(ctx, uuid.New(), ownerID, collabID, model.CollaboratorRoleViewer)
		require.Error(t, err)
	})
}

func TestDocumentService_SignDocument_Errors(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	signerID := uuid.New()
	userRepo := &docTestUserRepo{users: map[uuid.UUID]*model.User{
		ownerID:  {ID: ownerID, Name: "Owner"},
		signerID: {ID: signerID, Name: "SignerUser"},
	}}

	t.Run("not found doc", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		_, err := svc.SignDocument(ctx, uuid.New(), signerID, "Test")
		require.Error(t, err)
	})

	t.Run("sign with empty name uses user name", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "EmptyName", OwnerID: ownerID})
		_ = svc.AddSigner(ctx, doc.Document.ID, ownerID, signerID)
		_ = svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedBySignatures)

		result, err := svc.SignDocument(ctx, doc.Document.ID, signerID, "")
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestDocumentService_Duplicate_WithBlocks(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}

	t.Run("duplicate with blocks", func(t *testing.T) {
		docRepo := newMockDocumentRepo()
		blockRepo := newMockBlockRepo()
		svc := NewDocumentService(docRepo, blockRepo, &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)

		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "Orig", OwnerID: ownerID, TemplateID: "notulen-rapat"})
		dup, err := svc.Duplicate(ctx, doc.Document.ID, ownerID)
		require.NoError(t, err)
		assert.True(t, len(dup.Blocks) > 0)
	})

	t.Run("not found", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		_, err := svc.Duplicate(ctx, uuid.New(), ownerID)
		require.Error(t, err)
	})
}

func TestDocumentService_ListByContext_Topic(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}

	svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
	topicID := uuid.New()
	_, _ = svc.Create(ctx, CreateDocumentInput{Title: "TopicDoc", OwnerID: ownerID, TopicID: &topicID})

	items, err := svc.ListByContext(ctx, "topic", topicID)
	require.NoError(t, err)
	assert.True(t, len(items) > 0)
	assert.Equal(t, "topic", items[0].ContextType)
}

func TestDocumentService_toListItems_ContextTypes(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}

	svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)

	// Standalone doc
	standalone, _ := svc.Create(ctx, CreateDocumentInput{Title: "Standalone", OwnerID: ownerID, IsStandalone: true})

	// Topic doc
	topicID := uuid.New()
	topicDoc, _ := svc.Create(ctx, CreateDocumentInput{Title: "TopicD", OwnerID: ownerID, TopicID: &topicID})

	items, err := svc.ListAll(ctx, ownerID)
	require.NoError(t, err)

	for _, item := range items {
		if item.ID == standalone.Document.ID {
			assert.Equal(t, "standalone", item.ContextType)
		}
		if item.ID == topicDoc.Document.ID {
			assert.Equal(t, "topic", item.ContextType)
		}
	}
}

func TestDocumentService_Update_EditorCollaborator(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	editorID := uuid.New()
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	docRepo := newMockDocumentRepo()

	svc := NewDocumentService(docRepo, newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
	doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "EditorTest", OwnerID: ownerID})
	_ = svc.AddCollaborator(ctx, doc.Document.ID, ownerID, editorID, model.CollaboratorRoleEditor)

	newTitle := "EditorUpdated"
	result, err := svc.Update(ctx, doc.Document.ID, editorID, model.UpdateDocumentInput{Title: &newTitle})
	require.NoError(t, err)
	assert.Equal(t, "EditorUpdated", result.Title)
}

func TestDocumentService_Create_Errors(t *testing.T) {
	ctx := context.Background()
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	ownerID := uuid.New()

	t.Run("default title and icon", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, err := svc.Create(ctx, CreateDocumentInput{OwnerID: ownerID})
		require.NoError(t, err)
		assert.Equal(t, "Dokumen Tanpa Judul", doc.Document.Title)
		assert.NotEmpty(t, doc.Document.Icon)
		assert.True(t, doc.Document.IsStandalone)
	})

	t.Run("with template having rows and columns", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, err := svc.Create(ctx, CreateDocumentInput{
			OwnerID:    ownerID,
			TemplateID: "inventaris-aset",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, doc.Blocks)
	})

	t.Run("with template having emoji and color", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		// Use notulen-rapat or absensi which may have callout blocks
		doc, err := svc.Create(ctx, CreateDocumentInput{
			OwnerID:    ownerID,
			TemplateID: "notulen-rapat",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, doc.Blocks)
	})
}

func TestDocumentService_ListAll_Error(t *testing.T) {
	ctx := context.Background()
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	docRepo := newMockDocumentRepo()

	svc := NewDocumentService(docRepo, newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
	ownerID := uuid.New()

	// Add doc owned by user
	_, err := svc.Create(ctx, CreateDocumentInput{Title: "MyDoc", OwnerID: ownerID})
	require.NoError(t, err)

	items, err := svc.ListAll(ctx, ownerID)
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestDocumentService_Update_LockedDoc(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	docRepo := newMockDocumentRepo()

	svc := NewDocumentService(docRepo, newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
	doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "LockedDoc", OwnerID: ownerID})

	// Lock manually
	docRepo.docs[doc.Document.ID].Locked = true

	newTitle := "Attempted"
	_, err := svc.Update(ctx, doc.Document.ID, ownerID, model.UpdateDocumentInput{Title: &newTitle})
	require.Error(t, err)
}

func TestDocumentService_Update_NonOwnerNonEditor(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	viewerID := uuid.New()
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	docRepo := newMockDocumentRepo()

	svc := NewDocumentService(docRepo, newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
	doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "ViewerDoc", OwnerID: ownerID})
	_ = svc.AddCollaborator(ctx, doc.Document.ID, ownerID, viewerID, model.CollaboratorRoleViewer)

	newTitle := "Attempted"
	_, err := svc.Update(ctx, doc.Document.ID, viewerID, model.UpdateDocumentInput{Title: &newTitle})
	require.Error(t, err)
}

func TestDocumentService_Delete_LockedDoc(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	docRepo := newMockDocumentRepo()

	svc := NewDocumentService(docRepo, newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
	doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "LockedDoc", OwnerID: ownerID})
	docRepo.docs[doc.Document.ID].Locked = true

	err := svc.Delete(ctx, doc.Document.ID, ownerID)
	require.Error(t, err)
}

func TestDocumentService_Delete_NonOwner(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	otherID := uuid.New()
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	docRepo := newMockDocumentRepo()

	svc := NewDocumentService(docRepo, newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
	doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "NotMine", OwnerID: ownerID})

	err := svc.Delete(ctx, doc.Document.ID, otherID)
	require.Error(t, err)
}

func TestDocumentService_SignDocument_MoreErrors(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	signerID := uuid.New()
	nonSignerID := uuid.New()
	userRepo := &docTestUserRepo{users: map[uuid.UUID]*model.User{
		ownerID:     {ID: ownerID, Name: "Owner"},
		signerID:    {ID: signerID, Name: "Signer"},
		nonSignerID: {ID: nonSignerID, Name: "Other"},
	}}

	t.Run("not in signature mode", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "NoSigMode", OwnerID: ownerID})
		// Lock manually (not in signature mode)
		_, err := svc.SignDocument(ctx, doc.Document.ID, signerID, "Test")
		require.Error(t, err)
	})

	t.Run("not a signer", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "NotSigner", OwnerID: ownerID})
		_ = svc.AddSigner(ctx, doc.Document.ID, ownerID, signerID)
		_ = svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedBySignatures)

		_, err := svc.SignDocument(ctx, doc.Document.ID, nonSignerID, "Test")
		require.Error(t, err)
	})

	t.Run("already signed", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "AlreadySigned", OwnerID: ownerID})
		_ = svc.AddSigner(ctx, doc.Document.ID, ownerID, signerID)
		_ = svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedBySignatures)

		_, err := svc.SignDocument(ctx, doc.Document.ID, signerID, "Test")
		require.NoError(t, err)

		// Try to sign again
		_, err = svc.SignDocument(ctx, doc.Document.ID, signerID, "Test")
		require.Error(t, err)
	})
}

func TestDocumentService_ListByContext_Errors(t *testing.T) {
	ctx := context.Background()
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)

	t.Run("invalid context type", func(t *testing.T) {
		_, err := svc.ListByContext(ctx, "invalid", uuid.New())
		require.Error(t, err)
	})

	t.Run("chat context", func(t *testing.T) {
		ownerID := uuid.New()
		chatID := uuid.New()
		_, _ = svc.Create(ctx, CreateDocumentInput{Title: "ChatDoc", OwnerID: ownerID, ChatID: &chatID})

		items, err := svc.ListByContext(ctx, "chat", chatID)
		require.NoError(t, err)
		assert.True(t, len(items) > 0)
		assert.Equal(t, "chat", items[0].ContextType)
	})
}

func TestDocumentService_AddCollaborator_OwnerAsCollaborator(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}

	svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
	doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "SelfCollab", OwnerID: ownerID})

	err := svc.AddCollaborator(ctx, doc.Document.ID, ownerID, ownerID, model.CollaboratorRoleEditor)
	require.Error(t, err)
}

func TestDocumentService_UnlockDocument_MoreErrors(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}

	t.Run("not locked", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "NotLocked", OwnerID: ownerID})

		err := svc.UnlockDocument(ctx, doc.Document.ID, ownerID)
		require.Error(t, err)
	})

	t.Run("non-owner cannot unlock", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "NotOwner", OwnerID: ownerID})
		_ = svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedByManual)

		err := svc.UnlockDocument(ctx, doc.Document.ID, uuid.New())
		require.Error(t, err)
	})

	t.Run("cannot unlock signed doc", func(t *testing.T) {
		signerID := uuid.New()
		ur := &docTestUserRepo{users: map[uuid.UUID]*model.User{
			ownerID:  {ID: ownerID, Name: "Owner"},
			signerID: {ID: signerID, Name: "Signer"},
		}}
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, ur, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "SignedDoc", OwnerID: ownerID})
		_ = svc.AddSigner(ctx, doc.Document.ID, ownerID, signerID)
		_ = svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedBySignatures)
		_, _ = svc.SignDocument(ctx, doc.Document.ID, signerID, "Signer")

		err := svc.UnlockDocument(ctx, doc.Document.ID, ownerID)
		require.Error(t, err)
	})
}

func TestDocumentService_LockDocument_MoreErrors(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}

	t.Run("non-owner cannot lock", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "NotOwner", OwnerID: ownerID})

		err := svc.LockDocument(ctx, doc.Document.ID, uuid.New(), model.LockedByManual)
		require.Error(t, err)
	})

	t.Run("already locked", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "AlreadyLocked", OwnerID: ownerID})
		_ = svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedByManual)

		err := svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedByManual)
		require.Error(t, err)
	})

	t.Run("lock signatures without signers", func(t *testing.T) {
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "NoSigners", OwnerID: ownerID})

		err := svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedBySignatures)
		require.Error(t, err)
	})

	t.Run("lock with notification service", func(t *testing.T) {
		collabID := uuid.New()
		ur := &docTestUserRepo{users: map[uuid.UUID]*model.User{
			ownerID:  {ID: ownerID, Name: "Owner"},
			collabID: {ID: collabID, Name: "Collab"},
		}}
		notif := &mockNotifSvc{}
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, ur, NewTemplateService(), notif)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "WithNotif", OwnerID: ownerID})
		_ = svc.AddCollaborator(ctx, doc.Document.ID, ownerID, collabID, model.CollaboratorRoleEditor)

		err := svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedByManual)
		require.NoError(t, err)
	})
}

func TestDocumentService_AddSigner_MoreErrors(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	signerID := uuid.New()

	t.Run("non-owner cannot add signer", func(t *testing.T) {
		userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "NonOwner", OwnerID: ownerID})

		err := svc.AddSigner(ctx, doc.Document.ID, uuid.New(), signerID)
		require.Error(t, err)
	})

	t.Run("cannot add to locked doc", func(t *testing.T) {
		ur := &docTestUserRepo{users: map[uuid.UUID]*model.User{
			ownerID:  {ID: ownerID, Name: "Owner"},
			signerID: {ID: signerID, Name: "Signer"},
		}}
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, ur, NewTemplateService(), nil)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "Locked", OwnerID: ownerID})
		_ = svc.AddSigner(ctx, doc.Document.ID, ownerID, signerID)
		_ = svc.LockDocument(ctx, doc.Document.ID, ownerID, model.LockedBySignatures)

		newSigner := uuid.New()
		err := svc.AddSigner(ctx, doc.Document.ID, ownerID, newSigner)
		require.Error(t, err)
	})

	t.Run("with notification service", func(t *testing.T) {
		ur := &docTestUserRepo{users: map[uuid.UUID]*model.User{
			ownerID:  {ID: ownerID, Name: "Owner"},
			signerID: {ID: signerID, Name: "Signer"},
		}}
		notif := &mockNotifSvc{}
		svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, ur, NewTemplateService(), notif)
		doc, _ := svc.Create(ctx, CreateDocumentInput{Title: "WithNotif", OwnerID: ownerID})

		err := svc.AddSigner(ctx, doc.Document.ID, ownerID, signerID)
		require.NoError(t, err)
	})
}

func TestDocumentService_Tag_Errors(t *testing.T) {
	ctx := context.Background()
	userRepo := &docTestUserRepo{users: make(map[uuid.UUID]*model.User)}
	svc := NewDocumentService(newMockDocumentRepo(), newMockBlockRepo(), &mockDocHistoryRepo{}, userRepo, NewTemplateService(), nil)

	t.Run("empty tag", func(t *testing.T) {
		err := svc.AddTag(ctx, uuid.New(), "")
		require.Error(t, err)
	})
}
