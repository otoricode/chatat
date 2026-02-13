package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/otoritech/chatat/internal/middleware"
	"github.com/otoritech/chatat/pkg/apperror"
)

// DecodeJSON decodes a JSON request body into the given value.
func DecodeJSON(r *http.Request, v any) error {
	if r.Body == nil {
		return apperror.BadRequest("request body is required")
	}
	defer func() { _ = r.Body.Close() }()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		return apperror.BadRequest("invalid request body: " + err.Error())
	}
	return nil
}

// GetUserID extracts the authenticated user ID from request context.
func GetUserID(r *http.Request) (uuid.UUID, error) {
	id, ok := middleware.GetUserID(r.Context())
	if !ok {
		return uuid.Nil, apperror.Unauthorized("user not authenticated")
	}
	return id, nil
}

// GetPathParam extracts a URL path parameter.
func GetPathParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

// GetPathUUID extracts a UUID path parameter.
func GetPathUUID(r *http.Request, key string) (uuid.UUID, error) {
	param := chi.URLParam(r, key)
	id, err := uuid.Parse(param)
	if err != nil {
		return uuid.Nil, apperror.BadRequest("invalid " + key + " format")
	}
	return id, nil
}

// ParsePagination extracts cursor and limit from query parameters.
func ParsePagination(r *http.Request) (cursor string, limit int) {
	cursor = r.URL.Query().Get("cursor")
	limit = 20 // default

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	return cursor, limit
}

// ParseOffsetPagination extracts offset and limit from query parameters.
// Default limit is 20, max 100. Offset defaults to 0.
func ParseOffsetPagination(r *http.Request) (offset, limit int) {
	limit = 20

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	return offset, limit
}
