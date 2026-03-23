package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

var ErrInvalidPlatformSettings = errors.New("invalid platform settings")

type PlatformSettingsReader interface {
	Get() (*domain.PlatformSettings, error)
}

type UpdatePlatformSettingsInput struct {
	SelectionLockDurationHours         int
	RequireBothApprovals               bool
	AutoApproveAgencies                bool
	AutoExpireSelections               bool
	EmailNotificationsEnabled          bool
	MaintenanceMode                    bool
	MaintenanceMessage                 string
	AgencyApprovalEmailTemplate        string
	AgencyRejectionEmailTemplate       string
	SelectionNotificationEmailTemplate string
	ExpiryNotificationEmailTemplate    string
}

type PlatformSettingsService struct {
	settingsRepository domain.PlatformSettingsRepository
	auditRepository    domain.AuditLogRepository
}

func NewPlatformSettingsService(settingsRepository domain.PlatformSettingsRepository, auditRepository domain.AuditLogRepository) (*PlatformSettingsService, error) {
	if settingsRepository == nil {
		return nil, fmt.Errorf("platform settings repository is nil")
	}
	if auditRepository == nil {
		return nil, fmt.Errorf("audit repository is nil")
	}
	return &PlatformSettingsService{
		settingsRepository: settingsRepository,
		auditRepository:    auditRepository,
	}, nil
}

func (s *PlatformSettingsService) Get() (*domain.PlatformSettings, error) {
	settings, err := s.settingsRepository.Get()
	if err != nil {
		if errors.Is(err, repository.ErrPlatformSettingsNotFound) {
			return normalizePlatformSettings(domain.DefaultPlatformSettings()), nil
		}
		return nil, fmt.Errorf("load platform settings: %w", err)
	}
	return normalizePlatformSettings(settings), nil
}

func (s *PlatformSettingsService) Update(adminID string, input UpdatePlatformSettingsInput, ipAddress string) (*domain.PlatformSettings, error) {
	if err := validatePlatformSettingsInput(input); err != nil {
		return nil, err
	}

	settings, err := s.Get()
	if err != nil {
		return nil, err
	}

	settings.SelectionLockDurationHours = input.SelectionLockDurationHours
	settings.RequireBothApprovals = input.RequireBothApprovals
	settings.AutoApproveAgencies = input.AutoApproveAgencies
	settings.AutoExpireSelections = input.AutoExpireSelections
	settings.EmailNotificationsEnabled = input.EmailNotificationsEnabled
	settings.MaintenanceMode = input.MaintenanceMode
	settings.MaintenanceMessage = strings.TrimSpace(input.MaintenanceMessage)
	settings.AgencyApprovalEmailTemplate = strings.TrimSpace(input.AgencyApprovalEmailTemplate)
	settings.AgencyRejectionEmailTemplate = strings.TrimSpace(input.AgencyRejectionEmailTemplate)
	settings.SelectionNotificationEmailTemplate = strings.TrimSpace(input.SelectionNotificationEmailTemplate)
	settings.ExpiryNotificationEmailTemplate = strings.TrimSpace(input.ExpiryNotificationEmailTemplate)

	settings = normalizePlatformSettings(settings)
	if err := s.settingsRepository.Upsert(settings); err != nil {
		return nil, fmt.Errorf("save platform settings: %w", err)
	}

	payload, err := json.Marshal(map[string]any{
		"selection_lock_duration_hours":         settings.SelectionLockDurationHours,
		"require_both_approvals":                settings.RequireBothApprovals,
		"auto_approve_agencies":                 settings.AutoApproveAgencies,
		"auto_expire_selections":                settings.AutoExpireSelections,
		"email_notifications_enabled":           settings.EmailNotificationsEnabled,
		"maintenance_mode":                      settings.MaintenanceMode,
		"maintenance_message":                   settings.MaintenanceMessage,
		"agency_approval_email_template":        settings.AgencyApprovalEmailTemplate,
		"agency_rejection_email_template":       settings.AgencyRejectionEmailTemplate,
		"selection_notification_email_template": settings.SelectionNotificationEmailTemplate,
		"expiry_notification_email_template":    settings.ExpiryNotificationEmailTemplate,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal settings audit payload: %w", err)
	}

	if strings.TrimSpace(adminID) != "" {
		if err := s.auditRepository.Create(&domain.AuditLog{
			AdminID:    strings.TrimSpace(adminID),
			Action:     "update_platform_settings",
			TargetType: "platform_settings",
			Details:    payload,
			IPAddress:  strings.TrimSpace(ipAddress),
		}); err != nil {
			return nil, fmt.Errorf("create settings audit log: %w", err)
		}
	}

	return settings, nil
}

func normalizePlatformSettings(settings *domain.PlatformSettings) *domain.PlatformSettings {
	defaults := domain.DefaultPlatformSettings()
	if settings == nil {
		return defaults
	}

	if strings.TrimSpace(settings.ID) == "" {
		settings.ID = defaults.ID
	}
	if settings.SelectionLockDurationHours <= 0 {
		settings.SelectionLockDurationHours = defaults.SelectionLockDurationHours
	}
	if strings.TrimSpace(settings.MaintenanceMessage) == "" {
		settings.MaintenanceMessage = defaults.MaintenanceMessage
	}
	if strings.TrimSpace(settings.AgencyApprovalEmailTemplate) == "" {
		settings.AgencyApprovalEmailTemplate = defaults.AgencyApprovalEmailTemplate
	}
	if strings.TrimSpace(settings.AgencyRejectionEmailTemplate) == "" {
		settings.AgencyRejectionEmailTemplate = defaults.AgencyRejectionEmailTemplate
	}
	if strings.TrimSpace(settings.SelectionNotificationEmailTemplate) == "" {
		settings.SelectionNotificationEmailTemplate = defaults.SelectionNotificationEmailTemplate
	}
	if strings.TrimSpace(settings.ExpiryNotificationEmailTemplate) == "" {
		settings.ExpiryNotificationEmailTemplate = defaults.ExpiryNotificationEmailTemplate
	}

	return settings
}

func validatePlatformSettingsInput(input UpdatePlatformSettingsInput) error {
	switch input.SelectionLockDurationHours {
	case 12, 24, 48:
	default:
		return ErrInvalidPlatformSettings
	}
	if strings.TrimSpace(input.MaintenanceMessage) == "" {
		return ErrInvalidPlatformSettings
	}
	return nil
}

func renderTemplate(template string, variables map[string]string) string {
	rendered := template
	for key, value := range variables {
		rendered = strings.ReplaceAll(rendered, "{"+strings.TrimSpace(key)+"}", strings.TrimSpace(value))
	}
	return rendered
}
