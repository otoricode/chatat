package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/pkg/apperror"
	"github.com/otoritech/chatat/pkg/response"
)

func TestOK(t *testing.T) {
	rec := httptest.NewRecorder()
	response.OK(rec, map[string]string{"key": "value"})

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body response.SuccessResponse
	err := json.NewDecoder(rec.Body).Decode(&body)
	require.NoError(t, err)
	assert.True(t, body.Success)
}

func TestCreated(t *testing.T) {
	rec := httptest.NewRecorder()
	response.Created(rec, map[string]string{"id": "123"})

	assert.Equal(t, http.StatusCreated, rec.Code)

	var body response.SuccessResponse
	err := json.NewDecoder(rec.Body).Decode(&body)
	require.NoError(t, err)
	assert.True(t, body.Success)
}

func TestNoContent(t *testing.T) {
	rec := httptest.NewRecorder()
	response.NoContent(rec)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestPaginated(t *testing.T) {
	rec := httptest.NewRecorder()
	meta := response.PaginationMeta{
		Cursor:  "next-cursor",
		HasMore: true,
		Total:   100,
	}
	response.Paginated(rec, []string{"item1", "item2"}, meta)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body response.PaginatedResponse
	err := json.NewDecoder(rec.Body).Decode(&body)
	require.NoError(t, err)
	assert.True(t, body.Success)
	assert.Equal(t, "next-cursor", body.Meta.Cursor)
	assert.True(t, body.Meta.HasMore)
	assert.Equal(t, 100, body.Meta.Total)
}

func TestError_ClientError(t *testing.T) {
	rec := httptest.NewRecorder()
	appErr := apperror.BadRequest("invalid input")
	response.Error(rec, appErr)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var body response.ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&body)
	require.NoError(t, err)
	assert.False(t, body.Success)
	assert.Equal(t, "BAD_REQUEST", body.Error.Code)
}

func TestError_ServerError(t *testing.T) {
	rec := httptest.NewRecorder()
	appErr := apperror.Internal(nil)
	response.Error(rec, appErr)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
