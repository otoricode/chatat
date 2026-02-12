package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/handler"
)

func TestDecodeJSON_ValidBody(t *testing.T) {
	type testInput struct {
		Name string `json:"name"`
	}

	body := bytes.NewBufferString(`{"name":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/json")

	var input testInput
	err := handler.DecodeJSON(req, &input)
	require.NoError(t, err)
	assert.Equal(t, "test", input.Name)
}

func TestDecodeJSON_NilBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Body = nil

	var input map[string]any
	err := handler.DecodeJSON(req, &input)
	assert.Error(t, err)
}

func TestDecodeJSON_InvalidJSON(t *testing.T) {
	body := bytes.NewBufferString("not-json")
	req := httptest.NewRequest(http.MethodPost, "/test", body)

	var input map[string]any
	err := handler.DecodeJSON(req, &input)
	assert.Error(t, err)
}

func TestParsePagination_Defaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	cursor, limit := handler.ParsePagination(req)
	assert.Equal(t, "", cursor)
	assert.Equal(t, 20, limit)
}

func TestParsePagination_CustomValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test?cursor=abc&limit=50", nil)
	cursor, limit := handler.ParsePagination(req)
	assert.Equal(t, "abc", cursor)
	assert.Equal(t, 50, limit)
}

func TestParsePagination_InvalidLimit(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test?limit=abc", nil)
	_, limit := handler.ParsePagination(req)
	assert.Equal(t, 20, limit)
}

func TestParsePagination_ExceedsMax(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test?limit=200", nil)
	_, limit := handler.ParsePagination(req)
	assert.Equal(t, 20, limit) // should use default since 200 > 100
}

func TestGetUserID_NotAuthenticated(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	_, err := handler.GetUserID(req)
	assert.Error(t, err)
}
