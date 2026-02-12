package apperror_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/otoritech/chatat/pkg/apperror"
)

func TestRateLimited(t *testing.T) {
	err := apperror.RateLimited()
	assert.Equal(t, "RATE_LIMITED", err.Code)
	assert.Equal(t, http.StatusTooManyRequests, err.HTTPStatus)
}

func TestDocLocked(t *testing.T) {
	err := apperror.DocLocked()
	assert.Equal(t, "DOC_LOCKED", err.Code)
	assert.Equal(t, http.StatusLocked, err.HTTPStatus)
}

func TestInvalidOTP(t *testing.T) {
	err := apperror.InvalidOTP()
	assert.Equal(t, "INVALID_OTP", err.Code)
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
}
