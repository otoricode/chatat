// Package apperror provides structured application error types
// with HTTP status codes and user-friendly messages.
package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// Sentinel errors for error checking with errors.Is.
var (
	ErrNotFound      = errors.New("not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrConflict      = errors.New("conflict")
	ErrBadRequest    = errors.New("bad request")
	ErrInternal      = errors.New("internal error")
	ErrAlreadyExists = errors.New("already exists")
)

// AppError wraps errors with HTTP status and user-facing message.
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NotFound creates a 404 error for a missing entity.
func NotFound(entity, id string) *AppError {
	return &AppError{
		Code:       "NOT_FOUND",
		Message:    fmt.Sprintf("%s '%s' not found", entity, id),
		HTTPStatus: http.StatusNotFound,
		Err:        ErrNotFound,
	}
}

// Unauthorized creates a 401 error.
func Unauthorized(msg string) *AppError {
	return &AppError{
		Code:       "UNAUTHORIZED",
		Message:    msg,
		HTTPStatus: http.StatusUnauthorized,
		Err:        ErrUnauthorized,
	}
}

// Forbidden creates a 403 error.
func Forbidden(msg string) *AppError {
	return &AppError{
		Code:       "FORBIDDEN",
		Message:    msg,
		HTTPStatus: http.StatusForbidden,
		Err:        ErrForbidden,
	}
}

// BadRequest creates a 400 error.
func BadRequest(msg string) *AppError {
	return &AppError{
		Code:       "BAD_REQUEST",
		Message:    msg,
		HTTPStatus: http.StatusBadRequest,
		Err:        ErrBadRequest,
	}
}

// Conflict creates a 409 error.
func Conflict(msg string) *AppError {
	return &AppError{
		Code:       "CONFLICT",
		Message:    msg,
		HTTPStatus: http.StatusConflict,
		Err:        ErrConflict,
	}
}

// Validation creates a 422 validation error.
func Validation(field, msg string) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    fmt.Sprintf("validation error on '%s': %s", field, msg),
		HTTPStatus: http.StatusUnprocessableEntity,
		Err:        ErrBadRequest,
	}
}

// Internal creates a 500 error. The original error is logged but not exposed to clients.
func Internal(err error) *AppError {
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    "an internal error occurred",
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
	}
}

// IsNotFound checks whether an error is a not-found error.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsConflict checks whether an error is a conflict error.
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsForbidden checks whether an error is a forbidden error.
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsUnauthorized checks whether an error is an unauthorized error.
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}
