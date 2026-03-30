package domain

import "time"

type UserRole string
type AccountStatus string

const (
	EthiopianAgent UserRole = "ethiopian_agent"
	ForeignAgent   UserRole = "foreign_agent"

	AccountStatusPendingApproval AccountStatus = "pending_approval"
	AccountStatusActive          AccountStatus = "active"
	AccountStatusRejected        AccountStatus = "rejected"
	AccountStatusSuspended       AccountStatus = "suspended"
)

type User struct {
	ID                      string `gorm:"type:uuid;primaryKey"`
	Email                   string `gorm:"uniqueIndex;not null"`
	PasswordHash            string `gorm:"not null"`
	FullName                string
	Role                    UserRole `gorm:"type:user_role;not null"`
	CompanyName             string
	AvatarURL               string
	AutoShareCandidates     bool
	DefaultForeignPairingID *string       `gorm:"type:uuid"`
	AccountStatus           AccountStatus `gorm:"type:account_status;not null;default:active"`
	IsActive                bool          `gorm:"not null;default:true"`
	CreatedAt               time.Time     `gorm:"not null;default:now()"`
	UpdatedAt               time.Time     `gorm:"not null;default:now()"`
}

func (User) TableName() string {
	return "users"
}

type UserRepository interface {
	Create(user *User) error
	GetByEmail(email string) (*User, error)
	GetByID(id string) (*User, error)
	Update(user *User) error
}
