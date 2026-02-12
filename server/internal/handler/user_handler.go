package handler

import (
	"net/http"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/pkg/apperror"
	"github.com/otoritech/chatat/pkg/response"
)

// UserHandler handles user profile endpoints.
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type setupProfileRequest struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type updateProfileRequest struct {
	Name   *string `json:"name"`
	Avatar *string `json:"avatar"`
	Status *string `json:"status"`
}

// GetMe handles GET /api/v1/users/me
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	user, err := h.userService.GetProfile(r.Context(), userID)
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

// UpdateMe handles PUT /api/v1/users/me
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	var req updateProfileRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	// At least one field must be provided
	if req.Name == nil && req.Avatar == nil && req.Status == nil {
		response.Error(w, apperror.BadRequest("at least one field must be provided"))
		return
	}

	user, err := h.userService.UpdateProfile(r.Context(), userID, model.UpdateUserInput{
		Name:   req.Name,
		Avatar: req.Avatar,
		Status: req.Status,
	})
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

// SetupProfile handles POST /api/v1/users/me/setup
func (h *UserHandler) SetupProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	var req setupProfileRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	user, err := h.userService.SetupProfile(r.Context(), userID, req.Name, req.Avatar)
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

// DeleteAccount handles DELETE /api/v1/users/me
func (h *UserHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	if err := h.userService.DeleteAccount(r.Context(), userID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(err))
		return
	}

	response.NoContent(w)
}
