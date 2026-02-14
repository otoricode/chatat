package apperror_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/otoritech/chatat/pkg/apperror"
)

func TestNotFound(t *testing.T) {
	err := apperror.NotFound("user", "123")
	assert.Equal(t, "NOT_FOUND", err.Code)
	assert.Equal(t, http.StatusNotFound, err.HTTPStatus)
	assert.Contains(t, err.Message, "user")
	assert.Contains(t, err.Message, "123")
}

func TestUnauthorized(t *testing.T) {
	err := apperror.Unauthorized("token expired")
	assert.Equal(t, "UNAUTHORIZED", err.Code)
	assert.Equal(t, http.StatusUnauthorized, err.HTTPStatus)
	assert.Equal(t, "token expired", err.Message)
}

func TestForbidden(t *testing.T) {
	err := apperror.Forbidden("not admin")
	assert.Equal(t, "FORBIDDEN", err.Code)
	assert.Equal(t, http.StatusForbidden, err.HTTPStatus)
	assert.Equal(t, "not admin", err.Message)
}

func TestBadRequest(t *testing.T) {
	err := apperror.BadRequest("invalid input")
	assert.Equal(t, "BAD_REQUEST", err.Code)
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
	assert.Equal(t, "invalid input", err.Message)
}

func TestConflict(t *testing.T) {
	err := apperror.Conflict("already exists")
	assert.Equal(t, "CONFLICT", err.Code)
	assert.Equal(t, http.StatusConflict, err.HTTPStatus)
	assert.Equal(t, "already exists", err.Message)
}

func TestValidation(t *testing.T) {
	err := apperror.Validation("email", "must be valid")
	assert.Equal(t, "VALIDATION_ERROR", err.Code)
	assert.Equal(t, http.StatusUnprocessableEntity, err.HTTPStatus)
	assert.Contains(t, err.Message, "email")
	assert.Contains(t, err.Message, "must be valid")
}

func TestInternal(t *testing.T) {
	cause := fmt.Errorf("db connection failed")
	err := apperror.Internal(cause)
	assert.Equal(t, "INTERNAL_ERROR", err.Code)
	assert.Equal(t, http.StatusInternalServerError, err.HTTPStatus)
	assert.Equal(t, "an internal error occurred", err.Message)
}

func TestError_WithWrapped(t *testing.T) {
	cause := fmt.Errorf("db error")
	err := apperror.Internal(cause)
	assert.Contains(t, err.Error(), "db error")
}

func TestError_WithoutWrapped(t *testing.T) {
	err := apperror.BadRequest("bad")
	// BadRequest wraps ErrBadRequest, so Error() includes it
	assert.Contains(t, err.Error(), "bad")
}

func TestUnwrap(t *testing.T) {
	err := apperror.NotFound("user", "1")
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
}

func TestIsNotFound(t *testing.T) {
	err := apperror.NotFound("user", "1")
	assert.True(t, apperror.IsNotFound(err))
	assert.False(t, apperror.IsNotFound(fmt.Errorf("other")))
}

func TestIsConflict(t *testing.T) {
	err := apperror.Conflict("dup")
	assert.True(t, apperror.IsConflict(err))
	assert.False(t, apperror.IsConflict(fmt.Errorf("other")))
}

func TestIsForbidden(t *testing.T) {
	err := apperror.Forbidden("no")
	assert.True(t, apperror.IsForbidden(err))
	assert.False(t, apperror.IsForbidden(fmt.Errorf("other")))
}

func TestIsUnauthorized(t *testing.T) {
	err := apperror.Unauthorized("no")
	assert.True(t, apperror.IsUnauthorized(err))
	assert.False(t, apperror.IsUnauthorized(fmt.Errorf("other")))
}

func TestErrorsIs_SentinelCompatibility(t *testing.T) {
	err := apperror.NotFound("item", "x")
	assert.True(t, errors.Is(err, apperror.ErrNotFound))
	assert.False(t, errors.Is(err, apperror.ErrForbidden))

	err2 := apperror.Forbidden("deny")
	assert.True(t, errors.Is(err2, apperror.ErrForbidden))
}

func TestErrorsAs(t *testing.T) {
	err := apperror.BadRequest("bad input")
	var appErr *apperror.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, "BAD_REQUEST", appErr.Code)
}
