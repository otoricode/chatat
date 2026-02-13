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

// -- Mock entity repository --

type mockEntityRepo struct {
	entities   map[uuid.UUID]*model.Entity
	links      map[uuid.UUID][]uuid.UUID // docID -> []entityID
	entityDocs map[uuid.UUID][]uuid.UUID // entityID -> []docID
}

func newMockEntityRepo() *mockEntityRepo {
	return &mockEntityRepo{
		entities:   make(map[uuid.UUID]*model.Entity),
		links:      make(map[uuid.UUID][]uuid.UUID),
		entityDocs: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (m *mockEntityRepo) Create(_ context.Context, input model.CreateEntityInput) (*model.Entity, error) {
	e := &model.Entity{
		ID:            uuid.New(),
		Name:          input.Name,
		Type:          input.Type,
		Fields:        input.Fields,
		OwnerID:       input.OwnerID,
		ContactUserID: input.ContactUserID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	m.entities[e.ID] = e
	return e, nil
}

func (m *mockEntityRepo) FindByID(_ context.Context, id uuid.UUID) (*model.Entity, error) {
	e, ok := m.entities[id]
	if !ok {
		return nil, apperror.NotFound("entity", id.String())
	}
	return e, nil
}

func (m *mockEntityRepo) Update(_ context.Context, id uuid.UUID, input model.UpdateEntityInput) (*model.Entity, error) {
	e, ok := m.entities[id]
	if !ok {
		return nil, apperror.NotFound("entity", id.String())
	}
	if input.Name != nil {
		e.Name = *input.Name
	}
	if input.Type != nil {
		e.Type = *input.Type
	}
	if input.Fields != nil {
		e.Fields = *input.Fields
	}
	e.UpdatedAt = time.Now()
	m.entities[id] = e
	return e, nil
}

func (m *mockEntityRepo) ListByOwner(_ context.Context, ownerID uuid.UUID) ([]*model.Entity, error) {
	var result []*model.Entity
	for _, e := range m.entities {
		if e.OwnerID == ownerID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *mockEntityRepo) ListByOwnerWithFilters(_ context.Context, ownerID uuid.UUID, entityType string, limit, offset int) ([]*model.EntityListItem, int, error) {
	var all []*model.EntityListItem
	for _, e := range m.entities {
		if e.OwnerID != ownerID {
			continue
		}
		if entityType != "" && e.Type != entityType {
			continue
		}
		all = append(all, &model.EntityListItem{Entity: *e, DocumentCount: 0})
	}

	total := len(all)
	if offset >= total {
		return []*model.EntityListItem{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (m *mockEntityRepo) Search(_ context.Context, ownerID uuid.UUID, query string) ([]*model.Entity, error) {
	var result []*model.Entity
	for _, e := range m.entities {
		if e.OwnerID == ownerID && (contains(e.Name, query) || contains(e.Type, query)) {
			result = append(result, e)
		}
	}
	return result, nil
}

func contains(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && (s == sub || indexStr(s, sub) >= 0)
}

func indexStr(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func (m *mockEntityRepo) ListTypes(_ context.Context, ownerID uuid.UUID) ([]string, error) {
	seen := map[string]bool{}
	var types []string
	for _, e := range m.entities {
		if e.OwnerID == ownerID && !seen[e.Type] {
			seen[e.Type] = true
			types = append(types, e.Type)
		}
	}
	return types, nil
}

func (m *mockEntityRepo) LinkToDocument(_ context.Context, docID, entityID uuid.UUID) error {
	// Check if already linked
	for _, eid := range m.links[docID] {
		if eid == entityID {
			return apperror.BadRequest("entity sudah ditautkan ke dokumen ini")
		}
	}
	m.links[docID] = append(m.links[docID], entityID)
	m.entityDocs[entityID] = append(m.entityDocs[entityID], docID)
	return nil
}

func (m *mockEntityRepo) UnlinkFromDocument(_ context.Context, docID, entityID uuid.UUID) error {
	eids := m.links[docID]
	for i, eid := range eids {
		if eid == entityID {
			m.links[docID] = append(eids[:i], eids[i+1:]...)
			break
		}
	}
	dids := m.entityDocs[entityID]
	for i, did := range dids {
		if did == docID {
			m.entityDocs[entityID] = append(dids[:i], dids[i+1:]...)
			break
		}
	}
	return nil
}

func (m *mockEntityRepo) ListByDocument(_ context.Context, docID uuid.UUID) ([]*model.Entity, error) {
	var result []*model.Entity
	for _, eid := range m.links[docID] {
		if e, ok := m.entities[eid]; ok {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *mockEntityRepo) ListDocumentsByEntity(_ context.Context, entityID uuid.UUID) ([]*model.Document, error) {
	var result []*model.Document
	for _, did := range m.entityDocs[entityID] {
		result = append(result, &model.Document{ID: did, Title: "Doc"})
	}
	return result, nil
}

func (m *mockEntityRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.entities, id)
	return nil
}

// -- Mock user repo for entity tests --

type mockEntityUserRepo struct {
	users map[uuid.UUID]*model.User
}

func newMockEntityUserRepo() *mockEntityUserRepo {
	return &mockEntityUserRepo{users: make(map[uuid.UUID]*model.User)}
}

func (m *mockEntityUserRepo) Create(_ context.Context, input model.CreateUserInput) (*model.User, error) {
	u := &model.User{ID: uuid.New(), Phone: input.Phone, Name: input.Name, CreatedAt: time.Now()}
	m.users[u.ID] = u
	return u, nil
}

func (m *mockEntityUserRepo) FindByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, apperror.NotFound("user", id.String())
	}
	return u, nil
}

func (m *mockEntityUserRepo) FindByPhone(_ context.Context, phone string) (*model.User, error) {
	for _, u := range m.users {
		if u.Phone == phone {
			return u, nil
		}
	}
	return nil, apperror.NotFound("user", phone)
}

func (m *mockEntityUserRepo) Update(_ context.Context, id uuid.UUID, _ model.UpdateUserInput) (*model.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, apperror.NotFound("user", id.String())
	}
	return u, nil
}

func (m *mockEntityUserRepo) Search(_ context.Context, _ string) ([]*model.User, error) {
	return nil, nil
}

func (m *mockEntityUserRepo) FindByPhones(_ context.Context, _ []string) ([]*model.User, error) {
	return nil, nil
}

func (m *mockEntityUserRepo) FindByPhoneHashes(_ context.Context, _ []string) ([]*model.User, error) {
	return nil, nil
}

func (m *mockEntityUserRepo) UpdatePhoneHash(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

func (m *mockEntityUserRepo) UpdateLastSeen(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockEntityUserRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

// -- Mock doc repo for entity tests --

type mockEntityDocRepo struct {
	docs map[uuid.UUID]*model.Document
}

func newMockEntityDocRepo() *mockEntityDocRepo {
	return &mockEntityDocRepo{docs: make(map[uuid.UUID]*model.Document)}
}

func (m *mockEntityDocRepo) FindByID(_ context.Context, id uuid.UUID) (*model.Document, error) {
	d, ok := m.docs[id]
	if !ok {
		return nil, apperror.NotFound("document", id.String())
	}
	return d, nil
}

func (m *mockEntityDocRepo) Create(_ context.Context, input model.CreateDocumentInput) (*model.Document, error) {
	d := &model.Document{ID: uuid.New(), Title: input.Title, OwnerID: input.OwnerID}
	m.docs[d.ID] = d
	return d, nil
}

func (m *mockEntityDocRepo) ListByChat(_ context.Context, _ uuid.UUID) ([]*model.Document, error) {
	return nil, nil
}
func (m *mockEntityDocRepo) ListByTopic(_ context.Context, _ uuid.UUID) ([]*model.Document, error) {
	return nil, nil
}
func (m *mockEntityDocRepo) ListByOwner(_ context.Context, _ uuid.UUID) ([]*model.Document, error) {
	return nil, nil
}
func (m *mockEntityDocRepo) ListByTag(_ context.Context, _ string) ([]*model.Document, error) {
	return nil, nil
}
func (m *mockEntityDocRepo) ListAccessible(_ context.Context, _ uuid.UUID) ([]*model.Document, error) {
	return nil, nil
}
func (m *mockEntityDocRepo) Update(_ context.Context, _ uuid.UUID, _ model.UpdateDocumentInput) (*model.Document, error) {
	return nil, nil
}
func (m *mockEntityDocRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockEntityDocRepo) Search(_ context.Context, _ uuid.UUID, _ string) ([]*model.Document, error) {
	return nil, nil
}
func (m *mockEntityDocRepo) AddCollaborator(_ context.Context, _, _ uuid.UUID, _ model.CollaboratorRole) error {
	return nil
}
func (m *mockEntityDocRepo) RemoveCollaborator(_ context.Context, _, _ uuid.UUID) error { return nil }
func (m *mockEntityDocRepo) ListCollaborators(_ context.Context, _ uuid.UUID) ([]*model.DocumentCollaborator, error) {
	return nil, nil
}
func (m *mockEntityDocRepo) UpdateCollaboratorRole(_ context.Context, _, _ uuid.UUID, _ model.CollaboratorRole) error {
	return nil
}
func (m *mockEntityDocRepo) AddSigner(_ context.Context, _, _ uuid.UUID) error { return nil }
func (m *mockEntityDocRepo) RemoveSigner(_ context.Context, _, _ uuid.UUID) error       { return nil }
func (m *mockEntityDocRepo) ListSigners(_ context.Context, _ uuid.UUID) ([]*model.DocumentSigner, error) {
	return nil, nil
}
func (m *mockEntityDocRepo) RecordSignature(_ context.Context, _, _ uuid.UUID, _ string) error {
	return nil
}
func (m *mockEntityDocRepo) Lock(_ context.Context, _ uuid.UUID, _ model.LockedByType) error {
	return nil
}
func (m *mockEntityDocRepo) Unlock(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockEntityDocRepo) AddTag(_ context.Context, _ uuid.UUID, _ string) error    { return nil }
func (m *mockEntityDocRepo) RemoveTag(_ context.Context, _ uuid.UUID, _ string) error { return nil }
func (m *mockEntityDocRepo) ListTags(_ context.Context, _ uuid.UUID) ([]string, error)   { return nil, nil }
func (m *mockEntityDocRepo) UpdateSignerStatus(_ context.Context, _, _ uuid.UUID, _ string) error {
	return nil
}

// -- Helper to create service --

func newTestEntityService() (EntityService, *mockEntityRepo, *mockEntityUserRepo, *mockEntityDocRepo) {
	entityRepo := newMockEntityRepo()
	userRepo := newMockEntityUserRepo()
	docRepo := newMockEntityDocRepo()
	svc := NewEntityService(entityRepo, userRepo, docRepo)
	return svc, entityRepo, userRepo, docRepo
}

// -- Tests --

func TestEntityService_Create(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _ := newTestEntityService()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		entity, err := svc.Create(ctx, userID, CreateEntityInput{
			Name:   "PT Otoritech",
			Type:   "Perusahaan",
			Fields: map[string]string{"alamat": "Jakarta"},
		})
		require.NoError(t, err)
		assert.Equal(t, "PT Otoritech", entity.Name)
		assert.Equal(t, "Perusahaan", entity.Type)
		assert.Equal(t, "Jakarta", entity.Fields["alamat"])
		assert.Equal(t, userID, entity.OwnerID)
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := svc.Create(ctx, userID, CreateEntityInput{
			Name: "",
			Type: "Orang",
		})
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "BAD_REQUEST", appErr.Code)
	})

	t.Run("name too long", func(t *testing.T) {
		longName := make([]byte, 101)
		for i := range longName {
			longName[i] = 'a'
		}
		_, err := svc.Create(ctx, userID, CreateEntityInput{
			Name: string(longName),
			Type: "Orang",
		})
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "BAD_REQUEST", appErr.Code)
	})

	t.Run("empty type", func(t *testing.T) {
		_, err := svc.Create(ctx, userID, CreateEntityInput{
			Name: "Test",
			Type: "",
		})
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "BAD_REQUEST", appErr.Code)
	})

	t.Run("nil fields defaults to empty map", func(t *testing.T) {
		entity, err := svc.Create(ctx, userID, CreateEntityInput{
			Name: "Test",
			Type: "Orang",
		})
		require.NoError(t, err)
		assert.NotNil(t, entity.Fields)
		assert.Empty(t, entity.Fields)
	})
}

func TestEntityService_GetByID(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _ := newTestEntityService()
	ownerID := uuid.New()
	otherID := uuid.New()

	entity, err := svc.Create(ctx, ownerID, CreateEntityInput{
		Name: "Test Entity",
		Type: "Orang",
	})
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		found, err := svc.GetByID(ctx, entity.ID, ownerID)
		require.NoError(t, err)
		assert.Equal(t, entity.ID, found.ID)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetByID(ctx, uuid.New(), ownerID)
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "NOT_FOUND", appErr.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		_, err := svc.GetByID(ctx, entity.ID, otherID)
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "FORBIDDEN", appErr.Code)
	})
}

func TestEntityService_Update(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _ := newTestEntityService()
	ownerID := uuid.New()
	otherID := uuid.New()

	entity, err := svc.Create(ctx, ownerID, CreateEntityInput{
		Name:   "Original",
		Type:   "Orang",
		Fields: map[string]string{"key": "val"},
	})
	require.NoError(t, err)

	t.Run("success - update name", func(t *testing.T) {
		newName := "Updated"
		updated, err := svc.Update(ctx, entity.ID, ownerID, UpdateEntityInput{
			Name: &newName,
		})
		require.NoError(t, err)
		assert.Equal(t, "Updated", updated.Name)
		assert.Equal(t, "Orang", updated.Type)
	})

	t.Run("success - update type", func(t *testing.T) {
		newType := "Perusahaan"
		updated, err := svc.Update(ctx, entity.ID, ownerID, UpdateEntityInput{
			Type: &newType,
		})
		require.NoError(t, err)
		assert.Equal(t, "Perusahaan", updated.Type)
	})

	t.Run("success - update fields", func(t *testing.T) {
		newFields := map[string]string{"key2": "val2"}
		updated, err := svc.Update(ctx, entity.ID, ownerID, UpdateEntityInput{
			Fields: &newFields,
		})
		require.NoError(t, err)
		assert.Equal(t, "val2", updated.Fields["key2"])
	})

	t.Run("forbidden", func(t *testing.T) {
		newName := "Hack"
		_, err := svc.Update(ctx, entity.ID, otherID, UpdateEntityInput{
			Name: &newName,
		})
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "FORBIDDEN", appErr.Code)
	})

	t.Run("empty name", func(t *testing.T) {
		empty := "  "
		_, err := svc.Update(ctx, entity.ID, ownerID, UpdateEntityInput{
			Name: &empty,
		})
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "BAD_REQUEST", appErr.Code)
	})

	t.Run("empty type", func(t *testing.T) {
		empty := ""
		_, err := svc.Update(ctx, entity.ID, ownerID, UpdateEntityInput{
			Type: &empty,
		})
		require.Error(t, err)
	})
}

func TestEntityService_Delete(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _ := newTestEntityService()
	ownerID := uuid.New()
	otherID := uuid.New()

	entity, err := svc.Create(ctx, ownerID, CreateEntityInput{
		Name: "To Delete",
		Type: "Orang",
	})
	require.NoError(t, err)

	t.Run("forbidden", func(t *testing.T) {
		err := svc.Delete(ctx, entity.ID, otherID)
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "FORBIDDEN", appErr.Code)
	})

	t.Run("success", func(t *testing.T) {
		err := svc.Delete(ctx, entity.ID, ownerID)
		require.NoError(t, err)
		// Verify deleted
		_, err = svc.GetByID(ctx, entity.ID, ownerID)
		require.Error(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.Delete(ctx, uuid.New(), ownerID)
		require.Error(t, err)
	})
}

func TestEntityService_Search(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _ := newTestEntityService()
	userID := uuid.New()

	_, err := svc.Create(ctx, userID, CreateEntityInput{Name: "Budi Santoso", Type: "Orang"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, userID, CreateEntityInput{Name: "PT Maju", Type: "Perusahaan"})
	require.NoError(t, err)

	t.Run("search by name", func(t *testing.T) {
		results, err := svc.Search(ctx, userID, "Budi")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Budi Santoso", results[0].Name)
	})

	t.Run("empty query returns empty", func(t *testing.T) {
		results, err := svc.Search(ctx, userID, "  ")
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestEntityService_ListTypes(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _ := newTestEntityService()
	userID := uuid.New()

	_, _ = svc.Create(ctx, userID, CreateEntityInput{Name: "A", Type: "Orang"})
	_, _ = svc.Create(ctx, userID, CreateEntityInput{Name: "B", Type: "Perusahaan"})
	_, _ = svc.Create(ctx, userID, CreateEntityInput{Name: "C", Type: "Orang"})

	types, err := svc.ListTypes(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, types, 2)
	assert.Contains(t, types, "Orang")
	assert.Contains(t, types, "Perusahaan")
}

func TestEntityService_LinkDocument(t *testing.T) {
	ctx := context.Background()
	svc, _, _, docRepo := newTestEntityService()
	userID := uuid.New()

	entity, err := svc.Create(ctx, userID, CreateEntityInput{Name: "E1", Type: "Orang"})
	require.NoError(t, err)

	doc := &model.Document{ID: uuid.New(), Title: "Doc1", OwnerID: userID}
	docRepo.docs[doc.ID] = doc

	t.Run("link success", func(t *testing.T) {
		err := svc.LinkToDocument(ctx, entity.ID, doc.ID, userID)
		require.NoError(t, err)
	})

	t.Run("get document entities", func(t *testing.T) {
		entities, err := svc.GetDocumentEntities(ctx, doc.ID)
		require.NoError(t, err)
		assert.Len(t, entities, 1)
		assert.Equal(t, entity.ID, entities[0].ID)
	})

	t.Run("get entity documents", func(t *testing.T) {
		docs, err := svc.GetEntityDocuments(ctx, entity.ID)
		require.NoError(t, err)
		assert.Len(t, docs, 1)
	})

	t.Run("unlink success", func(t *testing.T) {
		err := svc.UnlinkFromDocument(ctx, entity.ID, doc.ID, userID)
		require.NoError(t, err)

		entities, err := svc.GetDocumentEntities(ctx, doc.ID)
		require.NoError(t, err)
		assert.Empty(t, entities)
	})

	t.Run("link forbidden", func(t *testing.T) {
		otherID := uuid.New()
		err := svc.LinkToDocument(ctx, entity.ID, doc.ID, otherID)
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, "FORBIDDEN", appErr.Code)
	})
}

func TestEntityService_CreateFromContact(t *testing.T) {
	ctx := context.Background()
	svc, _, userRepo, _ := newTestEntityService()
	ownerID := uuid.New()

	contactUser := &model.User{
		ID:    uuid.New(),
		Phone: "+6281234567890",
		Name:  "Budi Santoso",
	}
	userRepo.users[contactUser.ID] = contactUser

	t.Run("success", func(t *testing.T) {
		entity, err := svc.CreateFromContact(ctx, contactUser.ID, ownerID)
		require.NoError(t, err)
		assert.Equal(t, "Budi Santoso", entity.Name)
		assert.Equal(t, "Orang", entity.Type)
		assert.Equal(t, "+6281234567890", entity.Fields["telepon"])
		assert.Equal(t, ownerID, entity.OwnerID)
		assert.NotNil(t, entity.ContactUserID)
		assert.Equal(t, contactUser.ID, *entity.ContactUserID)
	})

	t.Run("contact not found", func(t *testing.T) {
		_, err := svc.CreateFromContact(ctx, uuid.New(), ownerID)
		require.Error(t, err)
	})

	t.Run("contact without name uses phone", func(t *testing.T) {
		noNameUser := &model.User{ID: uuid.New(), Phone: "+628999", Name: ""}
		userRepo.users[noNameUser.ID] = noNameUser

		entity, err := svc.CreateFromContact(ctx, noNameUser.ID, ownerID)
		require.NoError(t, err)
		assert.Equal(t, "+628999", entity.Name)
	})
}

func TestEntityService_List(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _ := newTestEntityService()
	userID := uuid.New()

	// Create 5 entities
	for i := 0; i < 3; i++ {
		_, _ = svc.Create(ctx, userID, CreateEntityInput{Name: "Orang " + string(rune('A'+i)), Type: "Orang"})
	}
	for i := 0; i < 2; i++ {
		_, _ = svc.Create(ctx, userID, CreateEntityInput{Name: "Corp " + string(rune('A'+i)), Type: "Perusahaan"})
	}

	t.Run("list all", func(t *testing.T) {
		items, total, err := svc.List(ctx, userID, "", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, 5, total)
		assert.Len(t, items, 5)
	})

	t.Run("filter by type", func(t *testing.T) {
		items, total, err := svc.List(ctx, userID, "Orang", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, items, 3)
	})

	t.Run("pagination", func(t *testing.T) {
		items, total, err := svc.List(ctx, userID, "", 2, 0)
		require.NoError(t, err)
		assert.Equal(t, 5, total)
		assert.Len(t, items, 2)
	})

	t.Run("limit capped at 100", func(t *testing.T) {
		items, _, err := svc.List(ctx, userID, "", 200, 0)
		require.NoError(t, err)
		assert.Len(t, items, 5) // only 5 entities exist
	})

	t.Run("negative offset becomes 0", func(t *testing.T) {
		_, _, err := svc.List(ctx, userID, "", 20, -5)
		require.NoError(t, err)
	})
}
