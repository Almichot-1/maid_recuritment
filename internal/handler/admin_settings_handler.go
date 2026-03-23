package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"maid-recruitment-tracking/internal/middleware"
	"maid-recruitment-tracking/internal/service"
	"maid-recruitment-tracking/pkg/utils"
)

type PlatformSettingsResponse struct {
	Settings any `json:"settings"`
}

type UpdatePlatformSettingsRequest struct {
	SelectionLockDurationHours         int    `json:"selection_lock_duration_hours"`
	RequireBothApprovals               bool   `json:"require_both_approvals"`
	AutoApproveAgencies                bool   `json:"auto_approve_agencies"`
	AutoExpireSelections               bool   `json:"auto_expire_selections"`
	EmailNotificationsEnabled          bool   `json:"email_notifications_enabled"`
	MaintenanceMode                    bool   `json:"maintenance_mode"`
	MaintenanceMessage                 string `json:"maintenance_message"`
	AgencyApprovalEmailTemplate        string `json:"agency_approval_email_template"`
	AgencyRejectionEmailTemplate       string `json:"agency_rejection_email_template"`
	SelectionNotificationEmailTemplate string `json:"selection_notification_email_template"`
	ExpiryNotificationEmailTemplate    string `json:"expiry_notification_email_template"`
}

type AdminSettingsHandler struct {
	settingsService *service.PlatformSettingsService
}

func NewAdminSettingsHandler(settingsService *service.PlatformSettingsService) *AdminSettingsHandler {
	return &AdminSettingsHandler{settingsService: settingsService}
}

func (h *AdminSettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settingsService.Get()
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load platform settings"})
		return
	}
	_ = utils.WriteJSON(w, http.StatusOK, PlatformSettingsResponse{Settings: settings})
}

func (h *AdminSettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req UpdatePlatformSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	adminID, ok := middleware.AdminIDFromContext(r.Context())
	if !ok || strings.TrimSpace(adminID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	settings, err := h.settingsService.Update(adminID, service.UpdatePlatformSettingsInput{
		SelectionLockDurationHours:         req.SelectionLockDurationHours,
		RequireBothApprovals:               req.RequireBothApprovals,
		AutoApproveAgencies:                req.AutoApproveAgencies,
		AutoExpireSelections:               req.AutoExpireSelections,
		EmailNotificationsEnabled:          req.EmailNotificationsEnabled,
		MaintenanceMode:                    req.MaintenanceMode,
		MaintenanceMessage:                 req.MaintenanceMessage,
		AgencyApprovalEmailTemplate:        req.AgencyApprovalEmailTemplate,
		AgencyRejectionEmailTemplate:       req.AgencyRejectionEmailTemplate,
		SelectionNotificationEmailTemplate: req.SelectionNotificationEmailTemplate,
		ExpiryNotificationEmailTemplate:    req.ExpiryNotificationEmailTemplate,
	}, clientIP(r))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPlatformSettings):
			_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid platform settings"})
		default:
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save platform settings"})
		}
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, PlatformSettingsResponse{Settings: settings})
}
