// Package response provides HTTP JSON response helper functions.
package response

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/otoritech/chatat/pkg/apperror"
)

// ErrorBody is the JSON structure returned for error responses.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OK writes a 200 JSON response with the given data.
func OK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, data)
}

// Created writes a 201 JSON response with the given data.
func Created(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusCreated, data)
}

// NoContent writes a 204 response with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error writes a JSON error response using the AppError's HTTP status.
func Error(w http.ResponseWriter, err *apperror.AppError) {
	if err.HTTPStatus >= http.StatusInternalServerError {
		log.Error().
			Str("code", err.Code).
			Err(err.Err).
			Msg("internal error")
	}

	writeJSON(w, err.HTTPStatus, ErrorBody{
		Code:    err.Code,
		Message: err.Message,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("failed to encode json response")
	}
}
