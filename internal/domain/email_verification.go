package domain

import "time"

type EmailVerificationToken struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	TokenHash string    `gorm:"not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

func (EmailVerificationToken) TableName() string {
	return "email_verification_tokens"
}

type EmailVerificationTokenRepository interface {
	Create(token *EmailVerificationToken) error
	GetActiveByTokenHash(tokenHash string) (*EmailVerificationToken, error)
	GetLatestByUserID(userID string) (*EmailVerificationToken, error)
	InvalidateActiveByUserID(userID string) error
	MarkUsed(id string) error
	Delete(id string) error
}
