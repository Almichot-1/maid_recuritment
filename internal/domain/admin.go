package domain

import (
	"encoding/json"
	"time"
)

type AdminRole string

const (
	SuperAdmin   AdminRole = "super_admin"
	SupportAdmin AdminRole = "support_admin"
)

type Admin struct {
	ID                  string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email               string    `gorm:"uniqueIndex;not null"`
	PasswordHash        string    `gorm:"not null"`
	FullName            string    `gorm:"not null"`
	Role                AdminRole `gorm:"type:admin_role;not null"`
	MFASecret           string    `gorm:"not null"`
	IsActive            bool      `gorm:"not null;default:true"`
	FailedLoginAttempts int       `gorm:"not null;default:0"`
	LockedUntil         *time.Time
	LastLogin           *time.Time
	ForcePasswordChange bool      `gorm:"not null;default:false"`
	CreatedAt           time.Time `gorm:"not null;default:now()"`
	UpdatedAt           time.Time `gorm:"not null;default:now()"`
}

func (Admin) TableName() string {
	return "admins"
}

type AdminSetupToken struct {
	ID        string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdminID   string `gorm:"type:uuid;not null;index"`
	TokenHash string `gorm:"not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

func (AdminSetupToken) TableName() string {
	return "admin_setup_tokens"
}

type AuditLog struct {
	ID         string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdminID    string `gorm:"type:uuid;not null"`
	Action     string `gorm:"not null"`
	TargetType string
	TargetID   *string         `gorm:"type:uuid"`
	Details    json.RawMessage `gorm:"type:jsonb;not null;default:'{}'::jsonb"`
	IPAddress  string
	CreatedAt  time.Time `gorm:"not null;default:now()"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

type AgencyApprovalStatus string

const (
	AgencyApprovalPending  AgencyApprovalStatus = "pending"
	AgencyApprovalApproved AgencyApprovalStatus = "approved"
	AgencyApprovalRejected AgencyApprovalStatus = "rejected"
)

type AgencyApprovalRequest struct {
	ID              string               `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgencyID        string               `gorm:"type:uuid;not null;uniqueIndex"`
	Status          AgencyApprovalStatus `gorm:"type:agency_approval_status;not null;default:pending"`
	ReviewedBy      *string              `gorm:"type:uuid"`
	ReviewedAt      *time.Time
	RejectionReason *string
	AdminNotes      *string
	CreatedAt       time.Time `gorm:"not null;default:now()"`
	UpdatedAt       time.Time `gorm:"not null;default:now()"`
}

func (AgencyApprovalRequest) TableName() string {
	return "agency_approval_requests"
}

type AdminRepository interface {
	Create(admin *Admin) error
	GetByEmail(email string) (*Admin, error)
	GetByID(id string) (*Admin, error)
	List() ([]*Admin, error)
	ListActive() ([]*Admin, error)
	Update(admin *Admin) error
}

type AdminSetupTokenRepository interface {
	Create(token *AdminSetupToken) error
	GetActiveByHash(tokenHash string) (*AdminSetupToken, error)
	InvalidateByAdminID(adminID string) error
	MarkUsed(id string, usedAt time.Time) error
}

type AuditLogRepository interface {
	Create(log *AuditLog) error
	List(filters AuditLogFilters) ([]*AuditLog, error)
}

type AuditLogFilters struct {
	AdminID    string
	Action     string
	TargetType string
	Page       int
	PageSize   int
}

type AgencyApprovalRequestRepository interface {
	Create(request *AgencyApprovalRequest) error
	GetByAgencyID(agencyID string) (*AgencyApprovalRequest, error)
	ListByStatus(status AgencyApprovalStatus) ([]*AgencyApprovalRequest, error)
	Update(request *AgencyApprovalRequest) error
}
