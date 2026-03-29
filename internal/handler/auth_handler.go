package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/middleware"
	"maid-recruitment-tracking/internal/repository"
	"maid-recruitment-tracking/internal/service"
	"maid-recruitment-tracking/pkg/utils"
)

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	FullName    string `json:"full_name" validate:"required"`
	Role        string `json:"role" validate:"required,oneof=ethiopian_agent foreign_agent"`
	CompanyName string `json:"company_name"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Code        string `json:"code" validate:"required,len=6"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  AuthUserView `json:"user"`
}

type RegisterResponse struct {
	Message string       `json:"message"`
	User    AuthUserView `json:"user"`
}

type AuthUserView struct {
	ID               string `json:"id"`
	Email            string `json:"email"`
	FullName         string `json:"full_name"`
	Role             string `json:"role"`
	CompanyName      string `json:"company_name,omitempty"`
	AvatarURL        string `json:"avatar_url,omitempty"`
	AccountStatus    string `json:"account_status"`
	CurrentSessionID string `json:"current_session_id,omitempty"`
}

type AuthHandler struct {
	authService     *service.AuthService
	userRepository  domain.UserRepository
	approvalService *service.AgencyApprovalService
	inputValidator  *validator.Validate
}

func NewAuthHandler(authService *service.AuthService, userRepository domain.UserRepository, approvalService *service.AgencyApprovalService) *AuthHandler {
	return &AuthHandler{
		authService:     authService,
		userRepository:  userRepository,
		approvalService: approvalService,
		inputValidator:  validator.New(),
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.inputValidator.Struct(req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed"})
		return
	}

	_, err := h.authService.Register(req.Email, req.Password, req.FullName, req.Role, req.CompanyName)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserExists):
			_ = utils.WriteJSON(w, http.StatusConflict, map[string]string{"error": "email already exists"})
			return
		case errors.Is(err, service.ErrInvalidCredentials):
			_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed"})
			return
		default:
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
	}

	user, err := h.userRepository.GetByEmail(strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load user"})
		return
	}
	if h.approvalService != nil {
		if err := h.approvalService.RegisterPendingAgency(user); err != nil {
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create approval request"})
			return
		}
	}

	_ = utils.WriteJSON(w, http.StatusAccepted, RegisterResponse{
		Message: "Registration submitted and pending approval",
		User:    mapUserToAuthUserView(user, ""),
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.inputValidator.Struct(req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed"})
		return
	}

	token, sessionID, err := h.authService.LoginWithSession(req.Email, req.Password, r.UserAgent(), clientIP(r))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		case errors.Is(err, service.ErrAccountPendingApproval):
			_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "account pending approval", "account_status": string(domain.AccountStatusPendingApproval)})
			return
		case errors.Is(err, service.ErrAccountRejected):
			_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "account rejected", "account_status": string(domain.AccountStatusRejected)})
			return
		case errors.Is(err, service.ErrAccountSuspended):
			_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "account suspended", "account_status": string(domain.AccountStatusSuspended)})
			return
		case errors.Is(err, service.ErrAccountInactive):
			_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "account inactive"})
			return
		}
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	user, err := h.userRepository.GetByEmail(strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load user"})
		return
	}

	middleware.SetSessionCookie(w, r, middleware.UserSessionCookieName, token, middleware.UserSessionMaxAgeSeconds)
	_ = utils.WriteJSON(w, http.StatusOK, AuthResponse{
		Token: token,
		User:  mapUserToAuthUserView(user, sessionID),
	})
}

func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.inputValidator.Struct(req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed"})
		return
	}

	message, err := h.authService.RequestPasswordReset(req.Email)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPasswordResetServiceMissing):
			_ = utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "password reset is not available right now"})
		default:
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": message})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.inputValidator.Struct(req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed"})
		return
	}

	if err := h.authService.ResetPassword(req.Email, req.Code, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, service.ErrPasswordResetCodeInvalid), errors.Is(err, service.ErrPasswordResetCodeExpired):
			_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "The reset code is invalid or has expired. Please request a new code."})
		case errors.Is(err, repository.ErrInvalidPassword):
			_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Password must be at least 8 characters long."})
		case errors.Is(err, service.ErrPasswordResetServiceMissing):
			_ = utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "password reset is not available right now"})
		default:
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Password reset successfully. You can now sign in."})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	sessionID, _ := middleware.SessionIDFromContext(r.Context())
	_ = h.authService.RevokeSession(userID, sessionID)
	middleware.ClearSessionCookie(w, r, middleware.UserSessionCookieName)
	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	user, err := h.userRepository.GetByID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	sessionID, _ := middleware.SessionIDFromContext(r.Context())
	_ = utils.WriteJSON(w, http.StatusOK, map[string]AuthUserView{
		"user": mapUserToAuthUserView(user, sessionID),
	})
}

func mapUserToAuthUserView(user *domain.User, currentSessionID string) AuthUserView {
	return AuthUserView{
		ID:               user.ID,
		Email:            user.Email,
		FullName:         user.FullName,
		Role:             string(user.Role),
		CompanyName:      user.CompanyName,
		AvatarURL:        user.AvatarURL,
		AccountStatus:    string(user.AccountStatus),
		CurrentSessionID: currentSessionID,
	}
}
