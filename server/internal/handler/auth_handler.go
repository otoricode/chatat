package handler

import (
	"net/http"
	"time"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/pkg/apperror"
	"github.com/otoritech/chatat/pkg/phone"
	"github.com/otoritech/chatat/pkg/response"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	otpService     service.OTPService
	reverseOTP     service.ReverseOTPService
	tokenService   service.TokenService
	sessionService service.SessionService
	userRepo       repository.UserRepository
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(
	otpService service.OTPService,
	reverseOTP service.ReverseOTPService,
	tokenService service.TokenService,
	sessionService service.SessionService,
	userRepo repository.UserRepository,
) *AuthHandler {
	return &AuthHandler{
		otpService:     otpService,
		reverseOTP:     reverseOTP,
		tokenService:   tokenService,
		sessionService: sessionService,
		userRepo:       userRepo,
	}
}

type sendOTPRequest struct {
	Phone string `json:"phone"`
}

type sendOTPResponse struct {
	ExpiresIn int `json:"expiresIn"`
}

// SendOTP handles POST /api/v1/auth/otp/send
func (h *AuthHandler) SendOTP(w http.ResponseWriter, r *http.Request) {
	var req sendOTPRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	normalized, err := phone.Normalize(req.Phone, "")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid phone number"))
		return
	}

	_, err = h.otpService.Generate(r.Context(), normalized)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, sendOTPResponse{ExpiresIn: 300})
}

type verifyOTPRequest struct {
	Phone    string `json:"phone"`
	Code     string `json:"code"`
	DeviceID string `json:"deviceId"`
}

type authResponse struct {
	AccessToken  string      `json:"accessToken"`
	RefreshToken string      `json:"refreshToken"`
	ExpiresAt    int64       `json:"expiresAt"`
	User         *model.User `json:"user"`
	IsNewUser    bool        `json:"isNewUser"`
}

// VerifyOTP handles POST /api/v1/auth/otp/verify
func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req verifyOTPRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	normalized, err := phone.Normalize(req.Phone, "")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid phone number"))
		return
	}

	if err := h.otpService.Verify(r.Context(), normalized, req.Code); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(err))
		return
	}

	h.completeAuth(w, r, normalized, req.DeviceID)
}

type initReverseOTPRequest struct {
	Phone string `json:"phone"`
}

type initReverseOTPResponse struct {
	SessionID string `json:"sessionId"`
	WANumber  string `json:"waNumber"`
	Code      string `json:"code"`
	ExpiresIn int    `json:"expiresIn"`
}

// InitReverseOTP handles POST /api/v1/auth/reverse-otp/init
func (h *AuthHandler) InitReverseOTP(w http.ResponseWriter, r *http.Request) {
	var req initReverseOTPRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	normalized, err := phone.Normalize(req.Phone, "")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid phone number"))
		return
	}

	session, err := h.reverseOTP.InitSession(r.Context(), normalized)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, initReverseOTPResponse{
		SessionID: session.SessionID,
		WANumber:  session.TargetWANumber,
		Code:      session.UniqueCode,
		ExpiresIn: int(time.Until(session.ExpiresAt).Seconds()),
	})
}

type checkReverseOTPRequest struct {
	SessionID string `json:"sessionId"`
	DeviceID  string `json:"deviceId"`
}

// CheckReverseOTP handles POST /api/v1/auth/reverse-otp/check
func (h *AuthHandler) CheckReverseOTP(w http.ResponseWriter, r *http.Request) {
	var req checkReverseOTPRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	result, err := h.reverseOTP.CheckVerification(r.Context(), req.SessionID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(err))
		return
	}

	if result.Status != "verified" {
		response.OK(w, map[string]string{"status": result.Status})
		return
	}

	h.completeAuth(w, r, result.Phone, req.DeviceID)
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// RefreshToken handles POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	tokens, err := h.tokenService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, tokens)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract access token from header
	accessToken := ""
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 {
		accessToken = authHeader[7:]
	}

	var req refreshRequest
	// Try to decode body for refresh token, but don't require it
	_ = DecodeJSON(r, &req)

	_ = h.tokenService.Revoke(r.Context(), accessToken, req.RefreshToken)

	// Also invalidate device session
	userID, err := GetUserID(r)
	if err == nil {
		_ = h.sessionService.Invalidate(r.Context(), userID)
	}

	response.NoContent(w)
}

// completeAuth finds or creates user and returns tokens.
func (h *AuthHandler) completeAuth(w http.ResponseWriter, r *http.Request, phoneNumber string, deviceID string) {
	ctx := r.Context()
	isNewUser := false

	// Find or create user
	user, err := h.userRepo.FindByPhone(ctx, phoneNumber)
	if err != nil {
		// Create new user
		input := model.CreateUserInput{
			Phone: phoneNumber,
			Name:  "User",
		}
		user, err = h.userRepo.Create(ctx, input)
		if err != nil {
			response.Error(w, apperror.Internal(err))
			return
		}
		isNewUser = true
	}

	// Generate tokens
	tokens, err := h.tokenService.Generate(ctx, user.ID)
	if err != nil {
		response.Error(w, apperror.Internal(err))
		return
	}

	// Register device session
	if deviceID != "" {
		_ = h.sessionService.Register(ctx, user.ID, deviceID, tokens.RefreshToken)
	}

	response.OK(w, authResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
		User:         user,
		IsNewUser:    isNewUser,
	})
}
