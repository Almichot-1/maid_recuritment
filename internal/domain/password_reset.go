package domain

import "time"

type PasswordResetRequest struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       string    `gorm:"type:uuid;not null;index"`
	CodeHash     string    `gorm:"not null"`
	ExpiresAt    time.Time `gorm:"not null"`
	UsedAt       *time.Time
	AttemptCount int       `gorm:"not null;default:0"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`
}

func (PasswordResetRequest) TableName() string {
	return "password_reset_requests"
}

type PasswordResetRequestRepository interface {
	Create(request *PasswordResetRequest) error
	GetLatestByUserID(userID string) (*PasswordResetRequest, error)
	GetLatestActiveByUserID(userID string) (*PasswordResetRequest, error)
	IncrementAttempts(id string) error
	MarkUsed(id string) error
	InvalidateActiveByUserID(userID string) error
	Delete(id string) error
}
