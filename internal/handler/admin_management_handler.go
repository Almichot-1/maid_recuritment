package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/middleware"
	"maid-recruitment-tracking/internal/repository"
	"maid-recruitment-tracking/internal/service"
	"maid-recruitment-tracking/pkg/utils"
)

type AdminManagementView struct {
	ID                  string  `json:"id"`
	Email               string  `json:"email"`
	FullName            string  `json:"full_name"`
	Role                string  `json:"role"`
	IsActive            bool    `json:"is_active"`
	FailedLoginAttempts int     `json:"failed_login_attempts"`
	ForcePasswordChange bool    `json:"force_password_change"`
	LastLogin           *string `json:"last_login,omitempty"`
	LockedUntil         *string `json:"locked_until,omitempty"`
	CreatedAt           string  `json:"created_at"`
}

type CreateAdminRequest struct {
	Email    string `json:"email" validate:"required,email"`
	FullName string `json:"full_name" validate:"required,min=2"`
	Role     string `json:"role" validate:"required,oneof=super_admin support_admin"`
}

type UpdateAdminRequest struct {
	Role                *string `json:"role,omitempty"`
	IsActive            *bool   `json:"is_active,omitempty"`
	ForcePasswordChange *bool   `json:"force_password_change,omitempty"`
}

type CreateAdminResponse struct {
	Admin             AdminManagementView `json:"admin"`
	SetupURL          string              `json:"setup_url"`
	InvitationWarning string              `json:"invitation_warning,omitempty"`
}

type AdminManagementHandler struct {
	adminRepository domain.AdminRepository
	auditRepository domain.AuditLogRepository
	authService     *service.AdminAuthService
	emailService    service.EmailService
	validator       *validator.Validate
}

func NewAdminManagementHandler(
	adminRepository domain.AdminRepository,
	auditRepository domain.AuditLogRepository,
	authService *service.AdminAuthService,
	emailService service.EmailService,
) *AdminManagementHandler {
	return &AdminManagementHandler{
		adminRepository: adminRepository,
		auditRepository: auditRepository,
		authService:     authService,
		emailService:    emailService,
		validator:       validator.New(),
	}
}

func (h *AdminManagementHandler) ListAdmins(w http.ResponseWriter, r *http.Request) {
	admins, err := h.adminRepository.List()
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	items := make([]AdminManagementView, 0, len(admins))
	for _, admin := range admins {
		if admin == nil {
			continue
		}
		items = append(items, mapAdminManagementView(admin))
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string][]AdminManagementView{"admins": items})
}

func (h *AdminManagementHandler) CreateAdmin(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.AdminIDFromContext(r.Context())
	if !ok {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req CreateAdminRequest
	if err := decodeJSONBody(w, r, &req, 16<<10); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := h.validator.Struct(req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed"})
		return
	}

	tempPassword, err := generateTemporaryAdminPassword()
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate temporary password"})
		return
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Maid Recruitment Platform",
		AccountName: strings.TrimSpace(strings.ToLower(req.Email)),
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
	})
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to provision mfa"})
		return
	}

	admin, err := h.authService.CreateAdmin(req.Email, tempPassword, req.FullName, req.Role, key.Secret())
	if err != nil {
		switch err {
		case service.ErrUserExists:
			_ = utils.WriteJSON(w, http.StatusConflict, map[string]string{"error": "admin email already exists"})
		case service.ErrWeakAdminPassword, service.ErrAdminInvalidCredentials:
			_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid admin details"})
		default:
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create admin"})
		}
		return
	}

	admin.ForcePasswordChange = true
	if err := h.adminRepository.Update(admin); err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to finalize admin setup"})
		return
	}

	setupURL, err := h.authService.CreateSetupInvitation(admin)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate admin setup link"})
		return
	}

	invitationWarning := "The new admin must use the one-time setup link to create a password and activate MFA."
	if h.emailService != nil {
		body := fmt.Sprintf(
			"Hello %s,\n\nYou have been added as an admin on the Maid Recruitment Platform.\n\nUse the secure one-time setup link below to create your password and add MFA to your authenticator app:\n\n%s\n\nThis setup link expires in 24 hours.",
			admin.FullName,
			setupURL,
		)
		if err := h.emailService.Send(admin.Email, "Your admin portal invitation", body); err != nil {
			invitationWarning = "Setup link generated, but invitation email delivery failed: " + err.Error()
		}
	}

	_ = h.createAudit(adminID, "create_admin", "admin", admin.ID, map[string]any{
		"email":     admin.Email,
		"full_name": admin.FullName,
		"role":      admin.Role,
	}, clientIP(r))

	_ = utils.WriteJSON(w, http.StatusCreated, CreateAdminResponse{
		Admin:             mapAdminManagementView(admin),
		SetupURL:          setupURL,
		InvitationWarning: invitationWarning,
	})
}

func (h *AdminManagementHandler) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.AdminIDFromContext(r.Context())
	if !ok {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	targetID := chi.URLParam(r, "id")
	admin, err := h.adminRepository.GetByID(targetID)
	if err != nil {
		if err == repository.ErrAdminNotFound {
			_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "admin not found"})
			return
		}
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	var req UpdateAdminRequest
	if err := decodeJSONBody(w, r, &req, 16<<10); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Role != nil {
		switch strings.TrimSpace(*req.Role) {
		case string(domain.SuperAdmin):
			admin.Role = domain.SuperAdmin
		case string(domain.SupportAdmin):
			admin.Role = domain.SupportAdmin
		default:
			_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid admin role"})
			return
		}
	}
	if req.IsActive != nil {
		admin.IsActive = *req.IsActive
	}
	if req.ForcePasswordChange != nil {
		admin.ForcePasswordChange = *req.ForcePasswordChange
	}

	if err := h.adminRepository.Update(admin); err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update admin"})
		return
	}

	_ = h.createAudit(adminID, "update_admin", "admin", admin.ID, map[string]any{
		"role":                  admin.Role,
		"is_active":             admin.IsActive,
		"force_password_change": admin.ForcePasswordChange,
	}, clientIP(r))

	_ = utils.WriteJSON(w, http.StatusOK, map[string]AdminManagementView{"admin": mapAdminManagementView(admin)})
}

func (h *AdminManagementHandler) createAudit(adminID, action, targetType, targetID string, details map[string]any, ipAddress string) error {
	payload, err := json.Marshal(details)
	if err != nil {
		return err
	}
	target := strings.TrimSpace(targetID)
	return h.auditRepository.Create(&domain.AuditLog{
		AdminID:    strings.TrimSpace(adminID),
		Action:     action,
		TargetType: strings.TrimSpace(targetType),
		TargetID:   &target,
		Details:    payload,
		IPAddress:  strings.TrimSpace(ipAddress),
	})
}

func mapAdminManagementView(admin *domain.Admin) AdminManagementView {
	view := AdminManagementView{
		ID:                  admin.ID,
		Email:               admin.Email,
		FullName:            admin.FullName,
		Role:                string(admin.Role),
		IsActive:            admin.IsActive,
		FailedLoginAttempts: admin.FailedLoginAttempts,
		ForcePasswordChange: admin.ForcePasswordChange,
		CreatedAt:           admin.CreatedAt.UTC().Format(time.RFC3339),
	}
	if admin.LastLogin != nil {
		value := admin.LastLogin.UTC().Format(time.RFC3339)
		view.LastLogin = &value
	}
	if admin.LockedUntil != nil {
		value := admin.LockedUntil.UTC().Format(time.RFC3339)
		view.LockedUntil = &value
	}
	return view
}

func generateTemporaryAdminPassword() (string, error) {
	const upper = "ABCDEFGHJKLMNPQRSTUVWXYZ"
	const lower = "abcdefghijkmnopqrstuvwxyz"
	const digits = "23456789"
	const special = "!@#$%^&*()-_=+"
	all := upper + lower + digits + special

	password := make([]byte, 0, 20)
	requiredSets := []string{upper, lower, digits, special}
	for _, charset := range requiredSets {
		value, err := randomPasswordChar(charset)
		if err != nil {
			return "", err
		}
		password = append(password, value)
	}

	for len(password) < 20 {
		value, err := randomPasswordChar(all)
		if err != nil {
			return "", err
		}
		password = append(password, value)
	}

	for index := len(password) - 1; index > 0; index-- {
		swapIndex, err := randomInt(index + 1)
		if err != nil {
			return "", err
		}
		password[index], password[swapIndex] = password[swapIndex], password[index]
	}

	return string(password), nil
}

func randomPasswordChar(charset string) (byte, error) {
	index, err := randomInt(len(charset))
	if err != nil {
		return 0, err
	}
	return charset[index], nil
}

func randomInt(max int) (int, error) {
	if max <= 0 {
		return 0, fmt.Errorf("invalid random bound")
	}
	value, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(value.Int64()), nil
}
