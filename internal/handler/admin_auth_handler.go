package handler

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/middleware"
	"maid-recruitment-tracking/internal/service"
	"maid-recruitment-tracking/pkg/utils"
)

type AdminLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	OTPCode  string `json:"otp_code" validate:"required,len=6"`
}

type AdminChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=12"`
	NewPassword     string `json:"new_password" validate:"required,min=12"`
}

type AdminSetupPreviewRequest struct {
	Token string `json:"token" validate:"required"`
}

type AdminSetupCompleteRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=12"`
	OTPCode     string `json:"otp_code" validate:"required,len=6"`
}

type AdminUserView struct {
	ID                  string  `json:"id"`
	Email               string  `json:"email"`
	FullName            string  `json:"full_name"`
	Role                string  `json:"role"`
	LastLogin           *string `json:"last_login,omitempty"`
	ForcePasswordChange bool    `json:"force_password_change"`
}

type AdminLoginResponse struct {
	Admin     AdminUserView `json:"admin"`
	ExpiresAt string        `json:"expires_at"`
}

type AdminSetupPreviewResponse struct {
	Admin           AdminUserView `json:"admin"`
	ProvisioningURL string        `json:"provisioning_url"`
}

type AdminAuthHandler struct {
	authService     *service.AdminAuthService
	adminRepository domain.AdminRepository
	inputValidator  *validator.Validate
}

func NewAdminAuthHandler(authService *service.AdminAuthService, adminRepository domain.AdminRepository) *AdminAuthHandler {
	return &AdminAuthHandler{
		authService:     authService,
		adminRepository: adminRepository,
		inputValidator:  validator.New(),
	}
}

func (h *AdminAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req AdminLoginRequest
	if err := decodeJSONBody(w, r, &req, 16<<10); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := h.inputValidator.Struct(req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed"})
		return
	}

	admin, token, err := h.authService.Login(req.Email, req.Password, req.OTPCode, clientIP(r))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAdminInvalidCredentials), errors.Is(err, service.ErrAdminInvalidMFA):
			_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid admin credentials"})
		case errors.Is(err, service.ErrAdminAccountLocked):
			_ = utils.WriteJSON(w, http.StatusLocked, map[string]string{"error": "admin account locked"})
		case errors.Is(err, service.ErrAdminSetupRequired):
			_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "admin setup required"})
		case errors.Is(err, service.ErrAdminInactive):
			_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "admin account inactive"})
		default:
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	expiresAt := time.Now().UTC().Add(1 * time.Hour)
	middleware.SetSessionCookie(w, r, middleware.AdminSessionCookieName, token, middleware.AdminSessionMaxAgeSeconds)
	_ = utils.WriteJSON(w, http.StatusOK, AdminLoginResponse{
		Admin:     mapAdminUserView(admin),
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

func (h *AdminAuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.AdminIDFromContext(r.Context())
	if !ok || strings.TrimSpace(adminID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	admin, err := h.adminRepository.GetByID(adminID)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]AdminUserView{
		"admin": mapAdminUserView(admin),
	})
}

func (h *AdminAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.AdminIDFromContext(r.Context())
	if ok {
		_ = h.authService.LogLogout(adminID, clientIP(r))
	}
	middleware.ClearSessionCookie(w, r, middleware.AdminSessionCookieName)
	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "admin logged out"})
}

func (h *AdminAuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.AdminIDFromContext(r.Context())
	if !ok || strings.TrimSpace(adminID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req AdminChangePasswordRequest
	if err := decodeJSONBody(w, r, &req, 16<<10); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := h.inputValidator.Struct(req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed"})
		return
	}

	if err := h.authService.ChangePassword(adminID, req.CurrentPassword, req.NewPassword, clientIP(r)); err != nil {
		switch {
		case errors.Is(err, service.ErrAdminPasswordMismatch):
			_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "current password is incorrect"})
		case errors.Is(err, service.ErrWeakAdminPassword):
			_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "new password must be at least 12 characters and include uppercase, lowercase, number, and special character"})
		case errors.Is(err, service.ErrAdminInactive):
			_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "admin account inactive"})
		default:
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "admin password updated"})
}

func (h *AdminAuthHandler) PreviewSetup(w http.ResponseWriter, r *http.Request) {
	var req AdminSetupPreviewRequest
	if err := decodeJSONBody(w, r, &req, 8<<10); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := h.inputValidator.Struct(req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed"})
		return
	}

	preview, err := h.authService.PreviewSetup(req.Token)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAdminInvalidToken):
			_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid setup token"})
		case errors.Is(err, service.ErrAdminInactive):
			_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "admin account inactive"})
		default:
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, AdminSetupPreviewResponse{
		Admin:           mapAdminUserView(preview.Admin),
		ProvisioningURL: preview.ProvisioningURL,
	})
}

func (h *AdminAuthHandler) CompleteSetup(w http.ResponseWriter, r *http.Request) {
	var req AdminSetupCompleteRequest
	if err := decodeJSONBody(w, r, &req, 16<<10); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := h.inputValidator.Struct(req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed"})
		return
	}

	if err := h.authService.CompleteSetup(req.Token, req.NewPassword, req.OTPCode, clientIP(r)); err != nil {
		switch {
		case errors.Is(err, service.ErrAdminInvalidToken):
			_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid setup token"})
		case errors.Is(err, service.ErrAdminInvalidMFA):
			_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid mfa code"})
		case errors.Is(err, service.ErrWeakAdminPassword):
			_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "new password must be at least 12 characters and include uppercase, lowercase, number, and special character"})
		case errors.Is(err, service.ErrAdminInactive):
			_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "admin account inactive"})
		default:
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "admin setup completed"})
}

func mapAdminUserView(admin *domain.Admin) AdminUserView {
	view := AdminUserView{
		ID:                  admin.ID,
		Email:               admin.Email,
		FullName:            admin.FullName,
		Role:                string(admin.Role),
		ForcePasswordChange: admin.ForcePasswordChange,
	}
	if admin.LastLogin != nil {
		value := admin.LastLogin.UTC().Format(time.RFC3339)
		view.LastLogin = &value
	}
	return view
}

func clientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}
