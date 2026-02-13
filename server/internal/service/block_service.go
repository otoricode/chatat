package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/pkg/apperror"
)

// BlockService handles block business logic.
type BlockService interface {
	AddBlock(ctx context.Context, docID, userID uuid.UUID, input AddBlockInput) (*model.Block, error)
	UpdateBlock(ctx context.Context, blockID, userID uuid.UUID, input model.UpdateBlockInput) (*model.Block, error)
	DeleteBlock(ctx context.Context, blockID, userID uuid.UUID) error
	MoveBlock(ctx context.Context, docID uuid.UUID, blockID uuid.UUID, newPosition int) error
	GetBlocks(ctx context.Context, docID uuid.UUID) ([]*model.Block, error)
	ReorderBlocks(ctx context.Context, docID, userID uuid.UUID, blockIDs []uuid.UUID) error
	BatchUpdate(ctx context.Context, docID, userID uuid.UUID, operations []BlockOperation) error
}

// AddBlockInput holds data for adding a block.
type AddBlockInput struct {
	Type          model.BlockType `json:"type"`
	Content       string          `json:"content"`
	Position      int             `json:"position"`
	Checked       *bool           `json:"checked"`
	Rows          json.RawMessage `json:"rows"`
	Columns       json.RawMessage `json:"columns"`
	Language      string          `json:"language"`
	Emoji         string          `json:"emoji"`
	Color         string          `json:"color"`
	ParentBlockID *uuid.UUID      `json:"parentBlockId"`
}

// BlockOperation represents a batch operation on blocks.
type BlockOperation struct {
	Type    string          `json:"type"` // "add", "update", "delete"
	BlockID *uuid.UUID      `json:"blockId"`
	Data    json.RawMessage `json:"data"`
}

type blockService struct {
	blockRepo   repository.BlockRepository
	docRepo     repository.DocumentRepository
	historyRepo repository.DocumentHistoryRepository
}

// NewBlockService creates a new block service.
func NewBlockService(
	blockRepo repository.BlockRepository,
	docRepo repository.DocumentRepository,
	historyRepo repository.DocumentHistoryRepository,
) BlockService {
	return &blockService{
		blockRepo:   blockRepo,
		docRepo:     docRepo,
		historyRepo: historyRepo,
	}
}

func (s *blockService) AddBlock(ctx context.Context, docID, userID uuid.UUID, input AddBlockInput) (*model.Block, error) {
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return nil, err
	}
	if doc.Locked {
		return nil, apperror.Forbidden("dokumen terkunci, tidak dapat menambah blok")
	}

	if err := validateBlockType(input.Type); err != nil {
		return nil, err
	}

	block, err := s.blockRepo.Create(ctx, model.CreateBlockInput{
		DocumentID:    docID,
		Type:          input.Type,
		Content:       input.Content,
		Checked:       input.Checked,
		Rows:          input.Rows,
		Columns:       input.Columns,
		Language:      input.Language,
		Emoji:         input.Emoji,
		Color:         input.Color,
		SortOrder:     input.Position,
		ParentBlockID: input.ParentBlockID,
	})
	if err != nil {
		return nil, fmt.Errorf("add block: %w", err)
	}

	_ = s.historyRepo.Create(ctx, docID, userID, "block_added", "Blok ditambahkan")
	return block, nil
}

func (s *blockService) UpdateBlock(ctx context.Context, blockID, userID uuid.UUID, input model.UpdateBlockInput) (*model.Block, error) {
	block, err := s.blockRepo.FindByID(ctx, blockID)
	if err != nil {
		return nil, err
	}

	doc, err := s.docRepo.FindByID(ctx, block.DocumentID)
	if err != nil {
		return nil, err
	}
	if doc.Locked {
		return nil, apperror.Forbidden("dokumen terkunci, tidak dapat mengubah blok")
	}

	updated, err := s.blockRepo.Update(ctx, blockID, input)
	if err != nil {
		return nil, err
	}

	_ = s.historyRepo.Create(ctx, doc.ID, userID, "block_updated", "Blok diperbarui")
	return updated, nil
}

func (s *blockService) DeleteBlock(ctx context.Context, blockID, userID uuid.UUID) error {
	block, err := s.blockRepo.FindByID(ctx, blockID)
	if err != nil {
		return err
	}

	doc, err := s.docRepo.FindByID(ctx, block.DocumentID)
	if err != nil {
		return err
	}
	if doc.Locked {
		return apperror.Forbidden("dokumen terkunci, tidak dapat menghapus blok")
	}

	if err := s.blockRepo.Delete(ctx, blockID); err != nil {
		return err
	}

	_ = s.historyRepo.Create(ctx, doc.ID, userID, "block_deleted", "Blok dihapus")
	return nil
}

func (s *blockService) MoveBlock(ctx context.Context, docID uuid.UUID, blockID uuid.UUID, newPosition int) error {
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return err
	}
	if doc.Locked {
		return apperror.Forbidden("dokumen terkunci, tidak dapat memindahkan blok")
	}

	blocks, err := s.blockRepo.ListByDocument(ctx, docID)
	if err != nil {
		return err
	}

	// Build reordered list
	var movedBlock *model.Block
	remaining := make([]*model.Block, 0, len(blocks))
	for _, b := range blocks {
		if b.ID == blockID {
			movedBlock = b
		} else {
			remaining = append(remaining, b)
		}
	}
	if movedBlock == nil {
		return apperror.NotFound("block", blockID.String())
	}

	if newPosition < 0 {
		newPosition = 0
	}
	if newPosition > len(remaining) {
		newPosition = len(remaining)
	}

	// Insert at new position
	reordered := make([]uuid.UUID, 0, len(blocks))
	for i := 0; i < len(remaining)+1; i++ {
		if i == newPosition {
			reordered = append(reordered, movedBlock.ID)
		}
		if i < len(remaining) {
			reordered = append(reordered, remaining[i].ID)
		}
	}

	return s.blockRepo.Reorder(ctx, docID, reordered)
}

func (s *blockService) GetBlocks(ctx context.Context, docID uuid.UUID) ([]*model.Block, error) {
	blocks, err := s.blockRepo.ListByDocument(ctx, docID)
	if err != nil {
		return nil, err
	}
	if blocks == nil {
		blocks = []*model.Block{}
	}
	return blocks, nil
}

func (s *blockService) ReorderBlocks(ctx context.Context, docID, userID uuid.UUID, blockIDs []uuid.UUID) error {
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return err
	}
	if doc.Locked {
		return apperror.Forbidden("dokumen terkunci, tidak dapat mengurutkan ulang")
	}

	if err := s.blockRepo.Reorder(ctx, docID, blockIDs); err != nil {
		return err
	}

	_ = s.historyRepo.Create(ctx, docID, userID, "blocks_reordered", "Urutan blok diubah")
	return nil
}

func (s *blockService) BatchUpdate(ctx context.Context, docID, userID uuid.UUID, operations []BlockOperation) error {
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		return err
	}
	if doc.Locked {
		return apperror.Forbidden("dokumen terkunci, tidak dapat melakukan batch update")
	}

	for _, op := range operations {
		switch op.Type {
		case "add":
			var input AddBlockInput
			if err := json.Unmarshal(op.Data, &input); err != nil {
				return apperror.BadRequest("data blok 'add' tidak valid")
			}
			_, addErr := s.AddBlock(ctx, docID, userID, input)
			if addErr != nil {
				return addErr
			}
		case "update":
			if op.BlockID == nil {
				return apperror.BadRequest("blockId diperlukan untuk operasi 'update'")
			}
			var input model.UpdateBlockInput
			if err := json.Unmarshal(op.Data, &input); err != nil {
				return apperror.BadRequest("data blok 'update' tidak valid")
			}
			_, updateErr := s.UpdateBlock(ctx, docID, userID, input)
			if updateErr != nil {
				return updateErr
			}
		case "delete":
			if op.BlockID == nil {
				return apperror.BadRequest("blockId diperlukan untuk operasi 'delete'")
			}
			if err := s.DeleteBlock(ctx, *op.BlockID, userID); err != nil {
				return err
			}
		default:
			return apperror.BadRequest("tipe operasi tidak valid: " + op.Type)
		}
	}

	return nil
}

// validateBlockType checks if a block type is valid.
func validateBlockType(bt model.BlockType) error {
	switch bt {
	case model.BlockTypeParagraph,
		model.BlockTypeHeading1,
		model.BlockTypeHeading2,
		model.BlockTypeHeading3,
		model.BlockTypeBulletList,
		model.BlockTypeNumberedList,
		model.BlockTypeChecklist,
		model.BlockTypeTable,
		model.BlockTypeCallout,
		model.BlockTypeCode,
		model.BlockTypeToggle,
		model.BlockTypeDivider,
		model.BlockTypeQuote:
		return nil
	default:
		return apperror.BadRequest("tipe blok tidak valid: " + string(bt))
	}
}
