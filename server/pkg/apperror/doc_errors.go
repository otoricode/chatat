package apperror

import "net/http"

// Additional domain-specific error constructors.

// RateLimited creates a 429 Too Many Requests error.
func RateLimited() *AppError {
	return &AppError{
		Code:       "RATE_LIMITED",
		Message:    "too many requests, please try again later",
		HTTPStatus: http.StatusTooManyRequests,
		Err:        ErrBadRequest,
	}
}

// DocLocked creates a 423 Locked error for locked documents.
func DocLocked() *AppError {
	return &AppError{
		Code:       "DOC_LOCKED",
		Message:    "document is locked and cannot be modified",
		HTTPStatus: http.StatusLocked,
		Err:        ErrForbidden,
	}
}

// InvalidOTP creates a bad request error for invalid OTP.
func InvalidOTP() *AppError {
	return &AppError{
		Code:       "INVALID_OTP",
		Message:    "OTP is invalid or expired",
		HTTPStatus: http.StatusBadRequest,
		Err:        ErrBadRequest,
	}
}
