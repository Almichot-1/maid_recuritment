package domain

import "time"

const PlatformSettingsSingletonID = "default"

type PlatformSettings struct {
	ID                                 string    `json:"id" gorm:"primaryKey;size:32"`
	SelectionLockDurationHours         int       `json:"selection_lock_duration_hours" gorm:"not null;default:24"`
	RequireBothApprovals               bool      `json:"require_both_approvals" gorm:"not null;default:true"`
	AutoApproveAgencies                bool      `json:"auto_approve_agencies" gorm:"not null;default:false"`
	AutoExpireSelections               bool      `json:"auto_expire_selections" gorm:"not null;default:true"`
	EmailNotificationsEnabled          bool      `json:"email_notifications_enabled" gorm:"not null;default:true"`
	MaintenanceMode                    bool      `json:"maintenance_mode" gorm:"not null;default:false"`
	MaintenanceMessage                 string    `json:"maintenance_message" gorm:"type:text;not null;default:''"`
	AgencyApprovalEmailTemplate        string    `json:"agency_approval_email_template" gorm:"type:text;not null;default:''"`
	AgencyRejectionEmailTemplate       string    `json:"agency_rejection_email_template" gorm:"type:text;not null;default:''"`
	SelectionNotificationEmailTemplate string    `json:"selection_notification_email_template" gorm:"type:text;not null;default:''"`
	ExpiryNotificationEmailTemplate    string    `json:"expiry_notification_email_template" gorm:"type:text;not null;default:''"`
	CreatedAt                          time.Time `json:"created_at,omitempty" gorm:"not null;default:now()"`
	UpdatedAt                          time.Time `json:"updated_at,omitempty" gorm:"not null;default:now()"`
}

func (PlatformSettings) TableName() string {
	return "platform_settings"
}

func DefaultPlatformSettings() *PlatformSettings {
	return &PlatformSettings{
		ID:                                 PlatformSettingsSingletonID,
		SelectionLockDurationHours:         24,
		RequireBothApprovals:               true,
		AutoApproveAgencies:                false,
		AutoExpireSelections:               true,
		EmailNotificationsEnabled:          true,
		MaintenanceMode:                    false,
		MaintenanceMessage:                 "The platform is currently under scheduled maintenance. Please try again later.",
		AgencyApprovalEmailTemplate:        "Hello {full_name},\n\nYour account for {company_name} has been approved. You can now log in to the platform.",
		AgencyRejectionEmailTemplate:       "Hello {full_name},\n\nYour application for {company_name} was rejected.\nReason: {reason}",
		SelectionNotificationEmailTemplate: "Hello {full_name},\n\n{message}\n\nCandidate: {candidate_name}",
		ExpiryNotificationEmailTemplate:    "Hello {full_name},\n\n{message}\n\nCandidate: {candidate_name}",
	}
}

type PlatformSettingsRepository interface {
	Get() (*PlatformSettings, error)
	Upsert(settings *PlatformSettings) error
}
