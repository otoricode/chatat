package response

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/otoritech/chatat/pkg/apperror"
)

// SuccessResponse wraps a successful API response.
type SuccessResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data,omitempty"`
}

// ErrorBody represents the error portion of an API response.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse wraps an error API response.
type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   ErrorBody `json:"error"`
}

// PaginationMeta holds pagination metadata.
type PaginationMeta struct {
	Cursor  string `json:"cursor,omitempty"`
	HasMore bool   `json:"hasMore"`
	Total   int    `json:"total,omitempty"`
}

// PaginatedResponse wraps a paginated API response.
type PaginatedResponse struct {
	Success bool           `json:"success"`
	Data    any            `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}

// OK sends a 200 success response.
func OK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, SuccessResponse{Success: true, Data: data})
}

// Created sends a 201 success response.
func Created(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusCreated, SuccessResponse{Success: true, Data: data})
}

// NoContent sends a 204 response with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Paginated sends a paginated response.
func Paginated(w http.ResponseWriter, data any, meta PaginationMeta) {
	writeJSON(w, http.StatusOK, PaginatedResponse{Success: true, Data: data, Meta: meta})
}

// Error sends an error response based on AppError.
func Error(w http.ResponseWriter, err *apperror.AppError) {
	if err.HTTPStatus >= http.StatusInternalServerError {
		log.Error().Str("code", err.Code).Err(err.Err).Msg("internal error")
	}
	writeJSON(w, err.HTTPStatus, ErrorResponse{
		Success: false,
		Error:   ErrorBody{Code: err.Code, Message: err.Message},
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("failed to encode json response")
	}
}
