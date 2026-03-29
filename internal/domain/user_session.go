package domain

import "time"

type UserSession struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     string    `gorm:"type:uuid;not null;index"`
	UserAgent  string    `gorm:"not null;default:''"`
	IPAddress  string    `gorm:"not null;default:''"`
	LastSeenAt time.Time `gorm:"not null;default:now()"`
	ExpiresAt  time.Time `gorm:"not null"`
	RevokedAt  *time.Time
	CreatedAt  time.Time `gorm:"not null;default:now()"`
	UpdatedAt  time.Time `gorm:"not null;default:now()"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}

type UserSessionRepository interface {
	Create(session *UserSession) error
	GetByID(id string) (*UserSession, error)
	ListActiveByUserID(userID string) ([]*UserSession, error)
	Touch(id string, lastSeenAt time.Time) error
	RevokeByID(userID, sessionID string) error
	RevokeAllByUserID(userID, exceptSessionID string) error
}
