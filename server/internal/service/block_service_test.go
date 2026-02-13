package service

import (
	"context"
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
}

func TestBlockService_GetBlocks(t *testing.T) {
	svc, docRepo, _ := newTestBlockService()
	ctx := context.Background()
	ownerID := uuid.New()
	doc := createTestDoc(docRepo, ownerID)

	_, _ = svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{Type: model.BlockTypeParagraph, Content: "A"})
	_, _ = svc.AddBlock(ctx, doc.ID, ownerID, AddBlockInput{Type: model.BlockTypeParagraph, Content: "B"})

	blocks, err := svc.GetBlocks(ctx, doc.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(blocks))
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
