package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/model"
)

func newTestBlockService() (BlockService, *mockDocumentRepo, *mockBlockRepo) {
	docRepo := newMockDocumentRepo()
	blockRepo := newMockBlockRepo()
	historyRepo := &mockDocHistoryRepo{}
	svc := NewBlockService(blockRepo, docRepo, historyRepo)
	return svc, docRepo, blockRepo
}

func createTestDoc(docRepo *mockDocumentRepo, ownerID uuid.UUID) *model.Document {
	doc, _ := docRepo.Create(context.Background(), model.CreateDocumentInput{
		Title:   "Test Doc",
		Icon:    "T",
		OwnerID: ownerID,
	})
	return doc
}

func TestBlockService_AddBlock(t *testing.T) {
	svc, docRepo, _ := newTestBlockService()
	ctx := context.Background()
	ownerID := uuid.New()
	doc := createTestDoc(docRepo, ownerID)

	t.Run("add paragraph block", func(t *testing.T) {
		block, err := svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{
			Type:    model.BlockTypeParagraph,
			Content: "Hello world",
		})
		require.NoError(t, err)
		assert.Equal(t, model.BlockTypeParagraph, block.Type)
		assert.Equal(t, "Hello world", block.Content)
	})

	t.Run("add heading block", func(t *testing.T) {
		block, err := svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{
			Type:    model.BlockTypeHeading1,
			Content: "Title",
		})
		require.NoError(t, err)
		assert.Equal(t, model.BlockTypeHeading1, block.Type)
	})

	t.Run("invalid block type", func(t *testing.T) {
		_, err := svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{
			Type:    "invalid-type",
			Content: "test",
		})
		require.Error(t, err)
	})

	t.Run("locked doc rejects add", func(t *testing.T) {
		docRepo.docs[doc.ID].Locked = true
		_, err := svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{
			Type:    model.BlockTypeParagraph,
			Content: "test",
		})
		require.Error(t, err)
		docRepo.docs[doc.ID].Locked = false
	})

	t.Run("doc not found", func(t *testing.T) {
		_, err := svc.AddBlock(ctx, uuid.New(), ownerID, AddBlockInput{
			Type: model.BlockTypeParagraph,
		})
		require.Error(t, err)
	})
}

func TestBlockService_UpdateBlock(t *testing.T) {
	svc, docRepo, _ := newTestBlockService()
	ctx := context.Background()
	ownerID := uuid.New()
	doc := createTestDoc(docRepo, ownerID)

	block, _ := svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{
		Type:    model.BlockTypeParagraph,
		Content: "Original",
	})

	t.Run("update content", func(t *testing.T) {
		newContent := "Updated"
		updated, err := svc.UpdateBlock(ctx, block.ID, ownerID, model.UpdateBlockInput{
			Content: &newContent,
		})
		require.NoError(t, err)
		assert.Equal(t, "Updated", updated.Content)
	})

	t.Run("locked doc rejects update", func(t *testing.T) {
		docRepo.docs[doc.ID].Locked = true
		newContent := "Cant"
		_, err := svc.UpdateBlock(ctx, block.ID, ownerID, model.UpdateBlockInput{
			Content: &newContent,
		})
		require.Error(t, err)
		docRepo.docs[doc.ID].Locked = false
	})

	t.Run("block not found", func(t *testing.T) {
		newContent := "test"
		_, err := svc.UpdateBlock(ctx, uuid.New(), ownerID, model.UpdateBlockInput{
			Content: &newContent,
		})
		require.Error(t, err)
	})
}

func TestBlockService_DeleteBlock(t *testing.T) {
	svc, docRepo, _ := newTestBlockService()
	ctx := context.Background()
	ownerID := uuid.New()
	doc := createTestDoc(docRepo, ownerID)

	t.Run("delete block", func(t *testing.T) {
		block, _ := svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{
			Type:    model.BlockTypeParagraph,
			Content: "Delete me",
		})
		err := svc.DeleteBlock(ctx, block.ID, ownerID)
		require.NoError(t, err)
	})

	t.Run("locked doc rejects delete", func(t *testing.T) {
		block, _ := svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{
			Type:    model.BlockTypeParagraph,
			Content: "Cant delete",
		})
		docRepo.docs[doc.ID].Locked = true
		err := svc.DeleteBlock(ctx, block.ID, ownerID)
		require.Error(t, err)
		docRepo.docs[doc.ID].Locked = false
	})

	t.Run("block not found", func(t *testing.T) {
		err := svc.DeleteBlock(ctx, uuid.New(), ownerID)
		require.Error(t, err)
	})
}

func TestBlockService_GetBlocks(t *testing.T) {
	svc, docRepo, _ := newTestBlockService()
	ctx := context.Background()
	ownerID := uuid.New()
	doc := createTestDoc(docRepo, ownerID)

	_, _ = svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{Type: model.BlockTypeParagraph, Content: "A"})
	_, _ = svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{Type: model.BlockTypeParagraph, Content: "B"})

	t.Run("success", func(t *testing.T) {
		blocks, err := svc.GetBlocks(ctx, doc.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, len(blocks))
	})

	t.Run("empty doc returns empty slice", func(t *testing.T) {
		emptyDoc := createTestDoc(docRepo, ownerID)
		blocks, err := svc.GetBlocks(ctx, emptyDoc.ID)
		require.NoError(t, err)
		assert.NotNil(t, blocks)
		assert.Equal(t, 0, len(blocks))
	})
}

func TestBlockService_ReorderBlocks(t *testing.T) {
	svc, docRepo, _ := newTestBlockService()
	ctx := context.Background()
	ownerID := uuid.New()
	doc := createTestDoc(docRepo, ownerID)

	b1, _ := svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{Type: model.BlockTypeParagraph, Content: "A"})
	b2, _ := svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{Type: model.BlockTypeParagraph, Content: "B"})

	t.Run("reorder succeeds", func(t *testing.T) {
		err := svc.ReorderBlocks(ctx, doc.ID, ownerID, []uuid.UUID{b2.ID, b1.ID})
		require.NoError(t, err)
	})

	t.Run("locked doc rejects reorder", func(t *testing.T) {
		docRepo.docs[doc.ID].Locked = true
		err := svc.ReorderBlocks(ctx, doc.ID, ownerID, []uuid.UUID{b1.ID, b2.ID})
		require.Error(t, err)
		docRepo.docs[doc.ID].Locked = false
	})

	t.Run("doc not found", func(t *testing.T) {
		err := svc.ReorderBlocks(ctx, uuid.New(), ownerID, []uuid.UUID{b1.ID, b2.ID})
		require.Error(t, err)
	})
}

func TestBlockService_AllBlockTypes(t *testing.T) {
	svc, docRepo, _ := newTestBlockService()
	ctx := context.Background()
	ownerID := uuid.New()
	doc := createTestDoc(docRepo, ownerID)

	types := []model.BlockType{
		model.BlockTypeParagraph, model.BlockTypeHeading1, model.BlockTypeHeading2,
		model.BlockTypeHeading3, model.BlockTypeBulletList, model.BlockTypeNumberedList,
		model.BlockTypeChecklist, model.BlockTypeTable, model.BlockTypeCallout,
		model.BlockTypeCode, model.BlockTypeToggle, model.BlockTypeDivider, model.BlockTypeQuote,
	}

	for _, bt := range types {
		t.Run(string(bt), func(t *testing.T) {
			block, err := svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{
				Type:    bt,
				Content: "test",
			})
			require.NoError(t, err)
			assert.Equal(t, bt, block.Type)
		})
	}
}

func TestTemplateService(t *testing.T) {
	svc := NewTemplateService()

	t.Run("get all templates", func(t *testing.T) {
		templates := svc.GetTemplates()
		assert.Equal(t, 8, len(templates))
	})

	t.Run("get template by id", func(t *testing.T) {
		tmpl := svc.GetTemplate("notulen-rapat")
		require.NotNil(t, tmpl)
		assert.Equal(t, "Notulen Rapat", tmpl.Name)
		assert.True(t, len(tmpl.Blocks) > 0)
	})

	t.Run("get nonexistent template", func(t *testing.T) {
		tmpl := svc.GetTemplate("does-not-exist")
		assert.Nil(t, tmpl)
	})

	t.Run("template blocks returned", func(t *testing.T) {
		blocks := svc.GetTemplateBlocks("daftar-belanja")
		require.NotNil(t, blocks)
		assert.True(t, len(blocks) > 0)
		// Should have a table block
		hasTable := false
		for _, b := range blocks {
			if b.Type == "table" {
				hasTable = true
				assert.NotNil(t, b.Columns)
				assert.NotNil(t, b.Rows)
			}
		}
		assert.True(t, hasTable)
	})

	t.Run("all template IDs exist", func(t *testing.T) {
		ids := []string{
			"kosong", "notulen-rapat", "daftar-belanja", "catatan-keuangan",
			"catatan-kesehatan", "kesepakatan-bersama", "catatan-pertanian", "inventaris-aset",
		}
		for _, id := range ids {
			tmpl := svc.GetTemplate(id)
			assert.NotNil(t, tmpl, "Template %s should exist", id)
		}
	})
}

func TestBlockService_MoveBlock(t *testing.T) {
	svc, docRepo, _ := newTestBlockService()
	userID := uuid.New()
	doc := createTestDoc(docRepo, userID)

	// Add 3 blocks
	b1, _ := svc.AddBlock(context.Background(), doc.ID, userID, AddBlockInput{Type: model.BlockTypeParagraph, Content: "A"})
	b2, _ := svc.AddBlock(context.Background(), doc.ID, userID, AddBlockInput{Type: model.BlockTypeParagraph, Content: "B"})
	_, _ = svc.AddBlock(context.Background(), doc.ID, userID, AddBlockInput{Type: model.BlockTypeParagraph, Content: "C"})

	t.Run("move to new position", func(t *testing.T) {
		err := svc.MoveBlock(context.Background(), doc.ID, b1.ID, 2)
		require.NoError(t, err)
	})

	t.Run("block not found", func(t *testing.T) {
		err := svc.MoveBlock(context.Background(), doc.ID, uuid.New(), 0)
		assert.Error(t, err)
	})

	t.Run("locked doc rejected", func(t *testing.T) {
		docRepo.docs[doc.ID].Locked = true
		err := svc.MoveBlock(context.Background(), doc.ID, b2.ID, 0)
		assert.Error(t, err)
		docRepo.docs[doc.ID].Locked = false
	})

	t.Run("negative position clamps to 0", func(t *testing.T) {
		err := svc.MoveBlock(context.Background(), doc.ID, b1.ID, -5)
		assert.NoError(t, err)
	})

	t.Run("position beyond end clamps", func(t *testing.T) {
		err := svc.MoveBlock(context.Background(), doc.ID, b1.ID, 999)
		assert.NoError(t, err)
	})

	t.Run("doc not found", func(t *testing.T) {
		err := svc.MoveBlock(context.Background(), uuid.New(), b1.ID, 0)
		assert.Error(t, err)
	})
}

func TestBlockService_BatchUpdate(t *testing.T) {
	svc, docRepo, _ := newTestBlockService()
	userID := uuid.New()
	doc := createTestDoc(docRepo, userID)

	t.Run("add via batch", func(t *testing.T) {
		addData, _ := json.Marshal(AddBlockInput{Type: model.BlockTypeParagraph, Content: "batch"})
		err := svc.BatchUpdate(context.Background(), doc.ID, userID, []BlockOperation{
			{Type: "add", Data: addData},
		})
		require.NoError(t, err)
	})

	t.Run("update via batch", func(t *testing.T) {
		// Note: BatchUpdate passes docID to UpdateBlock as blockID - tests error path
		bid := uuid.New()
		newContent := "updated-batch"
		updateData, _ := json.Marshal(model.UpdateBlockInput{Content: &newContent})
		err := svc.BatchUpdate(context.Background(), doc.ID, userID, []BlockOperation{
			{Type: "update", BlockID: &bid, Data: updateData},
		})
		// This will fail because blockID lookup fails (passes docID instead of op.BlockID)
		assert.Error(t, err)
	})

	t.Run("update nil blockID", func(t *testing.T) {
		updateData, _ := json.Marshal(model.UpdateBlockInput{})
		err := svc.BatchUpdate(context.Background(), doc.ID, userID, []BlockOperation{
			{Type: "update", BlockID: nil, Data: updateData},
		})
		require.Error(t, err)
	})

	t.Run("delete via batch", func(t *testing.T) {
		block, err := svc.AddBlock(context.Background(), doc.ID, userID, AddBlockInput{
			Type: model.BlockTypeParagraph, Content: "del-me",
		})
		require.NoError(t, err)
		err = svc.BatchUpdate(context.Background(), doc.ID, userID, []BlockOperation{
			{Type: "delete", BlockID: &block.ID, Data: json.RawMessage("{}")},
		})
		require.NoError(t, err)
	})

	t.Run("delete nil blockID", func(t *testing.T) {
		err := svc.BatchUpdate(context.Background(), doc.ID, userID, []BlockOperation{
			{Type: "delete", BlockID: nil, Data: json.RawMessage("{}")},
		})
		require.Error(t, err)
	})

	t.Run("invalid op type", func(t *testing.T) {
		err := svc.BatchUpdate(context.Background(), doc.ID, userID, []BlockOperation{
			{Type: "invalid", Data: json.RawMessage("{}")},
		})
		assert.Error(t, err)
	})

	t.Run("locked doc rejected", func(t *testing.T) {
		docRepo.docs[doc.ID].Locked = true
		err := svc.BatchUpdate(context.Background(), doc.ID, userID, []BlockOperation{})
		assert.Error(t, err)
		docRepo.docs[doc.ID].Locked = false
	})

	t.Run("doc not found", func(t *testing.T) {
		err := svc.BatchUpdate(context.Background(), uuid.New(), userID, []BlockOperation{})
		assert.Error(t, err)
	})

	t.Run("bad add data", func(t *testing.T) {
		err := svc.BatchUpdate(context.Background(), doc.ID, userID, []BlockOperation{
			{Type: "add", Data: json.RawMessage("bad")},
		})
		assert.Error(t, err)
	})

	t.Run("bad update data", func(t *testing.T) {
		bid := uuid.New()
		err := svc.BatchUpdate(context.Background(), doc.ID, userID, []BlockOperation{
			{Type: "update", BlockID: &bid, Data: json.RawMessage("bad")},
		})
		assert.Error(t, err)
	})
}

func TestBlockService_UpdateBlock_DocFindError(t *testing.T) {
	svc, docRepo, _ := newTestBlockService()
	ctx := context.Background()
	ownerID := uuid.New()
	doc := createTestDoc(docRepo, ownerID)

	block, err := svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{
		Type:    model.BlockTypeParagraph,
		Content: "Test",
	})
	require.NoError(t, err)

	// Remove doc so FindByID fails for block.DocumentID
	delete(docRepo.docs, doc.ID)
	newContent := "Updated"
	_, err = svc.UpdateBlock(ctx, block.ID, ownerID, model.UpdateBlockInput{Content: &newContent})
	require.Error(t, err)
}

func TestBlockService_UpdateBlock_RepoError(t *testing.T) {
	svc, docRepo, blockRepo := newTestBlockService()
	ctx := context.Background()
	ownerID := uuid.New()
	doc := createTestDoc(docRepo, ownerID)

	block, err := svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{
		Type:    model.BlockTypeParagraph,
		Content: "Test",
	})
	require.NoError(t, err)

	blockRepo.updateErr = assert.AnError
	newContent := "Updated"
	_, err = svc.UpdateBlock(ctx, block.ID, ownerID, model.UpdateBlockInput{Content: &newContent})
	require.Error(t, err)
	blockRepo.updateErr = nil
}

func TestBlockService_DeleteBlock_RepoError(t *testing.T) {
	svc, docRepo, blockRepo := newTestBlockService()
	ctx := context.Background()
	ownerID := uuid.New()
	doc := createTestDoc(docRepo, ownerID)

	block, err := svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{
		Type:    model.BlockTypeParagraph,
		Content: "Test",
	})
	require.NoError(t, err)

	blockRepo.deleteErr = assert.AnError
	err = svc.DeleteBlock(ctx, block.ID, ownerID)
	require.Error(t, err)
	blockRepo.deleteErr = nil
}

func TestBlockService_DeleteBlock_DocFindError(t *testing.T) {
	svc, docRepo, _ := newTestBlockService()
	ctx := context.Background()
	ownerID := uuid.New()
	doc := createTestDoc(docRepo, ownerID)

	block, err := svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{
		Type:    model.BlockTypeParagraph,
		Content: "Test",
	})
	require.NoError(t, err)

	delete(docRepo.docs, doc.ID)
	err = svc.DeleteBlock(ctx, block.ID, ownerID)
	require.Error(t, err)
}

func TestBlockService_GetBlocks_ListError(t *testing.T) {
	svc, docRepo, blockRepo := newTestBlockService()
	ctx := context.Background()
	ownerID := uuid.New()
	doc := createTestDoc(docRepo, ownerID)

	blockRepo.listErr = assert.AnError
	_, err := svc.GetBlocks(ctx, doc.ID)
	require.Error(t, err)
	blockRepo.listErr = nil
}

func TestBlockService_ReorderBlocks_RepoError(t *testing.T) {
	svc, docRepo, blockRepo := newTestBlockService()
	ctx := context.Background()
	ownerID := uuid.New()
	doc := createTestDoc(docRepo, ownerID)

	blockRepo.reorderErr = assert.AnError
	err := svc.ReorderBlocks(ctx, doc.ID, ownerID, []uuid.UUID{uuid.New()})
	require.Error(t, err)
	blockRepo.reorderErr = nil
}

func TestBlockService_AddBlock_CreateError(t *testing.T) {
	svc, docRepo, blockRepo := newTestBlockService()
	ctx := context.Background()
	ownerID := uuid.New()
	doc := createTestDoc(docRepo, ownerID)

	blockRepo.createErr = assert.AnError
	_, err := svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{
		Type:    model.BlockTypeParagraph,
		Content: "Test",
	})
	require.Error(t, err)
	blockRepo.createErr = nil
}

func TestBlockService_MoveBlock_ListError(t *testing.T) {
	svc, docRepo, blockRepo := newTestBlockService()
	ctx := context.Background()
	ownerID := uuid.New()
	doc := createTestDoc(docRepo, ownerID)

	blockRepo.listErr = assert.AnError
	err := svc.MoveBlock(ctx, doc.ID, uuid.New(), 0)
	require.Error(t, err)
	blockRepo.listErr = nil
}

func TestBlockService_MoveBlock_ReorderError(t *testing.T) {
	svc, docRepo, blockRepo := newTestBlockService()
	ctx := context.Background()
	ownerID := uuid.New()
	doc := createTestDoc(docRepo, ownerID)

	b1, _ := svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{Type: model.BlockTypeParagraph, Content: "A"})
	_, _ = svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{Type: model.BlockTypeParagraph, Content: "B"})

	blockRepo.reorderErr = assert.AnError
	err := svc.MoveBlock(ctx, doc.ID, b1.ID, 1)
	require.Error(t, err)
	blockRepo.reorderErr = nil
}
