package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/otoritech/chatat/internal/handler"
)

func TestParseOffsetPagination(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/items", nil)
		offset, limit := handler.ParseOffsetPagination(r)
		assert.Equal(t, 0, offset)
		assert.Equal(t, 20, limit)
	})

	t.Run("custom values", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/items?offset=10&limit=50", nil)
		offset, limit := handler.ParseOffsetPagination(r)
		assert.Equal(t, 10, offset)
		assert.Equal(t, 50, limit)
	})

	t.Run("invalid offset uses default", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/items?offset=abc", nil)
		offset, limit := handler.ParseOffsetPagination(r)
		assert.Equal(t, 0, offset)
		assert.Equal(t, 20, limit)
	})

	t.Run("negative offset uses default", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/items?offset=-5", nil)
		offset, limit := handler.ParseOffsetPagination(r)
		assert.Equal(t, 0, offset)
		assert.Equal(t, 20, limit)
	})

	t.Run("limit exceeding max uses default", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/items?limit=200", nil)
		offset, limit := handler.ParseOffsetPagination(r)
		assert.Equal(t, 0, offset)
		assert.Equal(t, 20, limit)
	})

	t.Run("zero limit uses default", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/items?limit=0", nil)
		offset, limit := handler.ParseOffsetPagination(r)
		assert.Equal(t, 0, offset)
		assert.Equal(t, 20, limit)
	})
}
