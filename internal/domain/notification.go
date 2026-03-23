package domain

import "time"

type NotificationType string

const (
	NotificationSelection    NotificationType = "selection"
	NotificationApproval     NotificationType = "approval"
	NotificationRejection    NotificationType = "rejection"
	NotificationStatusUpdate NotificationType = "status_update"
	NotificationExpiry       NotificationType = "expiry"
)

type Notification struct {
	ID                string           `gorm:"type:uuid;primaryKey"`
	UserID            string           `gorm:"type:uuid;not null;index"`
	Title             string           `gorm:"not null"`
	Message           string           `gorm:"not null"`
	Type              NotificationType `gorm:"type:notification_type;not null"`
	IsRead            bool             `gorm:"not null;default:false"`
	RelatedEntityType string
	RelatedEntityID   *string   `gorm:"type:uuid"`
	CreatedAt         time.Time `gorm:"not null;default:now()"`
}

func (Notification) TableName() string {
	return "notifications"
}

type NotificationRepository interface {
	Create(notification *Notification) error
	GetByUserID(userID string, unreadOnly bool) ([]*Notification, error)
	MarkAsRead(id string) error
	MarkAllAsRead(userID string) error
}
