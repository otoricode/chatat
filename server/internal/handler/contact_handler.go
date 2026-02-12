package handler

import (
	"net/http"

	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/pkg/apperror"
	"github.com/otoritech/chatat/pkg/phone"
	"github.com/otoritech/chatat/pkg/response"
)

// ContactHandler handles contact endpoints.
type ContactHandler struct {
	contactService service.ContactService
}

// NewContactHandler creates a new contact handler.
func NewContactHandler(contactService service.ContactService) *ContactHandler {
	return &ContactHandler{contactService: contactService}
}

type syncContactsRequest struct {
	PhoneHashes []string `json:"phoneHashes"`
}

// Sync handles POST /api/v1/contacts/sync
func (h *ContactHandler) Sync(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	var req syncContactsRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	if len(req.PhoneHashes) == 0 {
		response.Error(w, apperror.BadRequest("phoneHashes is required"))
		return
	}

	matches, err := h.contactService.SyncContacts(r.Context(), userID, req.PhoneHashes)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, matches)
}

// List handles GET /api/v1/contacts
func (h *ContactHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	contacts, err := h.contactService.GetContacts(r.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, contacts)
}

// Search handles GET /api/v1/contacts/search?phone=+628xxx
func (h *ContactHandler) Search(w http.ResponseWriter, r *http.Request) {
	phoneParam := r.URL.Query().Get("phone")
	if phoneParam == "" {
		response.Error(w, apperror.BadRequest("phone query parameter is required"))
		return
	}

	normalized, err := phone.Normalize(phoneParam, "")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid phone number"))
		return
	}

	user, err := h.contactService.SearchByPhone(r.Context(), normalized)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, user)
}

// GetProfile handles GET /api/v1/contacts/:userId
func (h *ContactHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	contactID, err := GetPathUUID(r, "userId")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid userId"))
		return
	}

	contact, err := h.contactService.GetContactProfile(r.Context(), contactID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, contact)
}
