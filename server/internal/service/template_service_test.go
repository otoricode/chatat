package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplateService_GetTemplates(t *testing.T) {
	svc := NewTemplateService()
	templates := svc.GetTemplates()
	assert.NotEmpty(t, templates, "should have built-in templates")
}

func TestTemplateService_GetTemplate(t *testing.T) {
	svc := NewTemplateService()

	t.Run("existing template", func(t *testing.T) {
		tpl := svc.GetTemplate("kosong")
		assert.NotNil(t, tpl)
		assert.Equal(t, "kosong", tpl.ID)
	})

	t.Run("non-existing template", func(t *testing.T) {
		tpl := svc.GetTemplate("nonexistent")
		assert.Nil(t, tpl)
	})
}

func TestTemplateService_GetTemplateBlocks(t *testing.T) {
	svc := NewTemplateService()

	t.Run("existing returns blocks", func(t *testing.T) {
		blocks := svc.GetTemplateBlocks("kosong")
		assert.NotNil(t, blocks)
		assert.NotEmpty(t, blocks)
	})

	t.Run("non-existing returns nil", func(t *testing.T) {
		blocks := svc.GetTemplateBlocks("nonexistent")
		assert.Nil(t, blocks)
	})

	t.Run("notulen-rapat has blocks", func(t *testing.T) {
		blocks := svc.GetTemplateBlocks("notulen-rapat")
		assert.NotEmpty(t, blocks)
	})

	t.Run("inventaris template has rows and columns", func(t *testing.T) {
		blocks := svc.GetTemplateBlocks("inventaris-aset")
		if assert.NotEmpty(t, blocks) {
			// Find block with table type
			found := false
			for _, b := range blocks {
				if b.Type == "table" {
					found = true
					assert.NotNil(t, b.Rows)
					assert.NotNil(t, b.Columns)
				}
			}
			assert.True(t, found, "should have a table block")
		}
	})
}

func TestMustJSON(t *testing.T) {
	// mustJSON should work with valid values
	result := mustJSON(map[string]string{"key": "value"})
	assert.NotNil(t, result)
	assert.Contains(t, string(result), "key")
}
