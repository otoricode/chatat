package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/pkg/apperror"
)

// SearchResults holds combined search results across all types.
type SearchResults struct {
	Messages  []*model.MessageSearchRow  `json:"messages"`
	Documents []*model.DocumentSearchRow `json:"documents"`
	Contacts  []*model.User              `json:"contacts"`
	Entities  []*model.Entity            `json:"entities"`
}

// SearchOpts holds pagination options for search.
type SearchOpts struct {
	Offset int
	Limit  int
}

// SearchService defines operations for searching across the application.
type SearchService interface {
	SearchAll(ctx context.Context, userID uuid.UUID, query string, limit int) (*SearchResults, error)
	SearchMessages(ctx context.Context, userID uuid.UUID, query string, opts SearchOpts) ([]*model.MessageSearchRow, error)
	SearchDocuments(ctx context.Context, userID uuid.UUID, query string, opts SearchOpts) ([]*model.DocumentSearchRow, error)
	SearchContacts(ctx context.Context, userID uuid.UUID, query string) ([]*model.User, error)
	SearchEntities(ctx context.Context, userID uuid.UUID, query string) ([]*model.Entity, error)
	SearchInChat(ctx context.Context, chatID, userID uuid.UUID, query string, opts SearchOpts) ([]*model.MessageSearchRow, error)
}

type searchService struct {
	searchRepo repository.SearchRepository
	chatRepo   repository.ChatRepository
}

// NewSearchService creates a new search service.
func NewSearchService(
	searchRepo repository.SearchRepository,
	chatRepo repository.ChatRepository,
) SearchService {
	return &searchService{
		searchRepo: searchRepo,
		chatRepo:   chatRepo,
	}
}

func (s *searchService) SearchAll(ctx context.Context, userID uuid.UUID, query string, limit int) (*SearchResults, error) {
	if len(query) < 2 {
		return nil, apperror.BadRequest("kata kunci pencarian minimal 2 karakter")
	}
	if limit <= 0 || limit > 5 {
		limit = 3
	}

	messages, err := s.searchRepo.SearchMessages(ctx, userID, query, 0, limit)
	if err != nil {
		return nil, fmt.Errorf("search messages: %w", err)
	}

	documents, err := s.searchRepo.SearchDocuments(ctx, userID, query, 0, limit)
	if err != nil {
		return nil, fmt.Errorf("search documents: %w", err)
	}

	contacts, err := s.searchRepo.SearchContacts(ctx, userID, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search contacts: %w", err)
	}

	entities, err := s.searchRepo.SearchEntities(ctx, userID, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search entities: %w", err)
	}

	return &SearchResults{
		Messages:  messages,
		Documents: documents,
		Contacts:  contacts,
		Entities:  entities,
	}, nil
}

func (s *searchService) SearchMessages(ctx context.Context, userID uuid.UUID, query string, opts SearchOpts) ([]*model.MessageSearchRow, error) {
	if len(query) < 2 {
		return nil, apperror.BadRequest("kata kunci pencarian minimal 2 karakter")
	}
	if opts.Limit <= 0 || opts.Limit > 100 {
		opts.Limit = 20
	}
	return s.searchRepo.SearchMessages(ctx, userID, query, opts.Offset, opts.Limit)
}

func (s *searchService) SearchDocuments(ctx context.Context, userID uuid.UUID, query string, opts SearchOpts) ([]*model.DocumentSearchRow, error) {
	if len(query) < 2 {
		return nil, apperror.BadRequest("kata kunci pencarian minimal 2 karakter")
	}
	if opts.Limit <= 0 || opts.Limit > 100 {
		opts.Limit = 20
	}
	return s.searchRepo.SearchDocuments(ctx, userID, query, opts.Offset, opts.Limit)
}

func (s *searchService) SearchContacts(ctx context.Context, userID uuid.UUID, query string) ([]*model.User, error) {
	if len(query) < 2 {
		return nil, apperror.BadRequest("kata kunci pencarian minimal 2 karakter")
	}
	return s.searchRepo.SearchContacts(ctx, userID, query, 50)
}

func (s *searchService) SearchEntities(ctx context.Context, userID uuid.UUID, query string) ([]*model.Entity, error) {
	if len(query) < 2 {
		return nil, apperror.BadRequest("kata kunci pencarian minimal 2 karakter")
	}
	return s.searchRepo.SearchEntities(ctx, userID, query, 50)
}

func (s *searchService) SearchInChat(ctx context.Context, chatID, userID uuid.UUID, query string, opts SearchOpts) ([]*model.MessageSearchRow, error) {
	if len(query) < 2 {
		return nil, apperror.BadRequest("kata kunci pencarian minimal 2 karakter")
	}

	// Verify user is member of the chat
	members, err := s.chatRepo.GetMembers(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get chat members: %w", err)
	}
	isMember := false
	for _, m := range members {
		if m.UserID == userID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, apperror.Forbidden("Anda bukan anggota chat ini")
	}

	if opts.Limit <= 0 || opts.Limit > 100 {
		opts.Limit = 20
	}
	return s.searchRepo.SearchMessagesInChat(ctx, chatID, query, opts.Offset, opts.Limit)
}
