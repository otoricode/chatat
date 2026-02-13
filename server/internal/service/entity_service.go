package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/pkg/apperror"
)

// EntityService defines operations for entity management.
type EntityService interface {
	Create(ctx context.Context, userID uuid.UUID, input CreateEntityInput) (*model.Entity, error)
	GetByID(ctx context.Context, entityID, userID uuid.UUID) (*model.Entity, error)
	List(ctx context.Context, userID uuid.UUID, entityType string, limit, offset int) ([]*model.EntityListItem, int, error)
	Update(ctx context.Context, entityID, userID uuid.UUID, input UpdateEntityInput) (*model.Entity, error)
	Delete(ctx context.Context, entityID, userID uuid.UUID) error
	Search(ctx context.Context, userID uuid.UUID, query string) ([]*model.Entity, error)
	ListTypes(ctx context.Context, userID uuid.UUID) ([]string, error)

	// Linking
	LinkToDocument(ctx context.Context, entityID, docID, userID uuid.UUID) error
	UnlinkFromDocument(ctx context.Context, entityID, docID, userID uuid.UUID) error
	GetDocumentEntities(ctx context.Context, docID uuid.UUID) ([]*model.Entity, error)
	GetEntityDocuments(ctx context.Context, entityID uuid.UUID) ([]*model.Document, error)

	// Contact-as-entity
	CreateFromContact(ctx context.Context, contactUserID, userID uuid.UUID) (*model.Entity, error)
}

// CreateEntityInput holds data for creating an entity via the service.
type CreateEntityInput struct {
	Name   string            `json:"name"`
	Type   string            `json:"type"`
	Fields map[string]string `json:"fields"`
}

// UpdateEntityInput holds data for updating an entity via the service.
type UpdateEntityInput struct {
	Name   *string            `json:"name,omitempty"`
	Type   *string            `json:"type,omitempty"`
	Fields *map[string]string `json:"fields,omitempty"`
}

type entityService struct {
	entityRepo repository.EntityRepository
	userRepo   repository.UserRepository
	docRepo    repository.DocumentRepository
}

// NewEntityService creates a new entity service.
func NewEntityService(entityRepo repository.EntityRepository, userRepo repository.UserRepository, docRepo repository.DocumentRepository) EntityService {
	return &entityService{
		entityRepo: entityRepo,
		userRepo:   userRepo,
		docRepo:    docRepo,
	}
}

func (s *entityService) Create(ctx context.Context, userID uuid.UUID, input CreateEntityInput) (*model.Entity, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, apperror.BadRequest("nama entity wajib diisi")
	}
	if len(name) > 100 {
		return nil, apperror.BadRequest("nama entity maksimal 100 karakter")
	}

	entityType := strings.TrimSpace(input.Type)
	if entityType == "" {
		return nil, apperror.BadRequest("tipe entity wajib diisi")
	}

	fields := input.Fields
	if fields == nil {
		fields = make(map[string]string)
	}

	entity, err := s.entityRepo.Create(ctx, model.CreateEntityInput{
		Name:    name,
		Type:    entityType,
		Fields:  fields,
		OwnerID: userID,
	})
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *entityService) GetByID(ctx context.Context, entityID, userID uuid.UUID) (*model.Entity, error) {
	entity, err := s.entityRepo.FindByID(ctx, entityID)
	if err != nil {
		return nil, err
	}

	if entity.OwnerID != userID {
		return nil, apperror.Forbidden("tidak memiliki akses ke entity ini")
	}

	return entity, nil
}

func (s *entityService) List(ctx context.Context, userID uuid.UUID, entityType string, limit, offset int) ([]*model.EntityListItem, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.entityRepo.ListByOwnerWithFilters(ctx, userID, entityType, limit, offset)
}

func (s *entityService) Update(ctx context.Context, entityID, userID uuid.UUID, input UpdateEntityInput) (*model.Entity, error) {
	entity, err := s.entityRepo.FindByID(ctx, entityID)
	if err != nil {
		return nil, err
	}

	if entity.OwnerID != userID {
		return nil, apperror.Forbidden("hanya pemilik yang dapat mengubah entity")
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, apperror.BadRequest("nama entity tidak boleh kosong")
		}
		if len(name) > 100 {
			return nil, apperror.BadRequest("nama entity maksimal 100 karakter")
		}
		input.Name = &name
	}

	if input.Type != nil {
		t := strings.TrimSpace(*input.Type)
		if t == "" {
			return nil, apperror.BadRequest("tipe entity tidak boleh kosong")
		}
		input.Type = &t
	}

	return s.entityRepo.Update(ctx, entityID, model.UpdateEntityInput{
		Name:   input.Name,
		Type:   input.Type,
		Fields: input.Fields,
	})
}

func (s *entityService) Delete(ctx context.Context, entityID, userID uuid.UUID) error {
	entity, err := s.entityRepo.FindByID(ctx, entityID)
	if err != nil {
		return err
	}

	if entity.OwnerID != userID {
		return apperror.Forbidden("hanya pemilik yang dapat menghapus entity")
	}

	return s.entityRepo.Delete(ctx, entityID)
}

func (s *entityService) Search(ctx context.Context, userID uuid.UUID, query string) ([]*model.Entity, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return []*model.Entity{}, nil
	}

	return s.entityRepo.Search(ctx, userID, q)
}

func (s *entityService) ListTypes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.entityRepo.ListTypes(ctx, userID)
}

func (s *entityService) LinkToDocument(ctx context.Context, entityID, docID, userID uuid.UUID) error {
	// Verify entity belongs to user
	entity, err := s.entityRepo.FindByID(ctx, entityID)
	if err != nil {
		return err
	}
	if entity.OwnerID != userID {
		return apperror.Forbidden("tidak memiliki akses ke entity ini")
	}

	// Verify document exists
	_, err = s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return err
	}

	return s.entityRepo.LinkToDocument(ctx, docID, entityID)
}

func (s *entityService) UnlinkFromDocument(ctx context.Context, entityID, docID, userID uuid.UUID) error {
	entity, err := s.entityRepo.FindByID(ctx, entityID)
	if err != nil {
		return err
	}
	if entity.OwnerID != userID {
		return apperror.Forbidden("tidak memiliki akses ke entity ini")
	}

	return s.entityRepo.UnlinkFromDocument(ctx, docID, entityID)
}

func (s *entityService) GetDocumentEntities(ctx context.Context, docID uuid.UUID) ([]*model.Entity, error) {
	return s.entityRepo.ListByDocument(ctx, docID)
}

func (s *entityService) GetEntityDocuments(ctx context.Context, entityID uuid.UUID) ([]*model.Document, error) {
	return s.entityRepo.ListDocumentsByEntity(ctx, entityID)
}

func (s *entityService) CreateFromContact(ctx context.Context, contactUserID, userID uuid.UUID) (*model.Entity, error) {
	// Get the contact user to populate entity name
	contactUser, err := s.userRepo.FindByID(ctx, contactUserID)
	if err != nil {
		return nil, apperror.NotFound("kontak", contactUserID.String())
	}

	name := contactUser.Name
	if name == "" {
		name = contactUser.Phone
	}

	fields := make(map[string]string)
	if contactUser.Phone != "" {
		fields["telepon"] = contactUser.Phone
	}

	entity, err := s.entityRepo.Create(ctx, model.CreateEntityInput{
		Name:          name,
		Type:          "Orang",
		Fields:        fields,
		OwnerID:       userID,
		ContactUserID: &contactUserID,
	})
	if err != nil {
		return nil, err
	}

	return entity, nil
}
