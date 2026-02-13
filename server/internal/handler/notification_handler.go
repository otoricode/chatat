package handler

import (
	"encoding/json"
	"net/http"

	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/pkg/apperror"
	"github.com/otoritech/chatat/pkg/response"
)

// NotificationHandler handles push notification HTTP endpoints.
type NotificationHandler struct {
	service service.NotificationService
}

// NewNotificationHandler creates a new notification handler.
func NewNotificationHandler(svc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: svc}
}

type registerDeviceRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

type unregisterDeviceRequest struct {
	Token string `json:"token"`
}

// RegisterDevice handles POST /notifications/devices
func (h *NotificationHandler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	var req registerDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperror.BadRequest("body request tidak valid"))
		return
	}

	if req.Token == "" {
		response.Error(w, apperror.BadRequest("token wajib diisi"))
		return
	}
	if req.Platform != "ios" && req.Platform != "android" {
		response.Error(w, apperror.BadRequest("platform harus ios atau android"))
		return
	}

	if err := h.service.RegisterDevice(r.Context(), userID, req.Token, req.Platform); err != nil {
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, map[string]string{"message": "device berhasil didaftarkan"})
}

// UnregisterDevice handles DELETE /notifications/devices
func (h *NotificationHandler) UnregisterDevice(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	var req unregisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperror.BadRequest("body request tidak valid"))
		return
	}

	if req.Token == "" {
		response.Error(w, apperror.BadRequest("token wajib diisi"))
		return
	}

	if err := h.service.UnregisterDevice(r.Context(), userID, req.Token); err != nil {
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, map[string]string{"message": "device berhasil dihapus"})
}
