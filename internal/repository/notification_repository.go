package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

var ErrNotificationNotFound = errors.New("notification not found")

type GormNotificationRepository struct {
	db *gorm.DB
}

func (r *GormNotificationRepository) GetByID(id string) (*domain.Notification, error) {
	var notification domain.Notification
	if err := r.db.Where("id = ?", id).First(&notification).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotificationNotFound
		}
		return nil, fmt.Errorf("get notification by id: %w", err)
	}
	return &notification, nil
}

func (r *GormNotificationRepository) CountByUserID(userID string, unreadOnly bool) (int64, error) {
	if strings.TrimSpace(userID) == "" {
		return 0, fmt.Errorf("count notifications: user id is required")
	}

	query := r.db.Model(&domain.Notification{}).Where("user_id = ?", userID)
	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count notifications by user id: %w", err)
	}
	return total, nil
}

func NewGormNotificationRepository(cfg *config.Config) (*GormNotificationRepository, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return nil, fmt.Errorf("database url is empty")
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	return &GormNotificationRepository{db: db}, nil
}

func (r *GormNotificationRepository) Create(notification *domain.Notification) error {
	if notification == nil {
		return fmt.Errorf("create notification: notification is nil")
	}
	if strings.TrimSpace(notification.UserID) == "" {
		return fmt.Errorf("create notification: user id is required")
	}
	if strings.TrimSpace(notification.Title) == "" {
		return fmt.Errorf("create notification: title is required")
	}
	if strings.TrimSpace(notification.Message) == "" {
		return fmt.Errorf("create notification: message is required")
	}
	if notification.ID == "" {
		notification.ID = uuid.NewString()
	}
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now().UTC()
	}

	if err := r.db.Create(notification).Error; err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	return nil
}

func (r *GormNotificationRepository) GetByUserID(userID string, unreadOnly bool) ([]*domain.Notification, error) {
	return r.GetByUserIDPaginated(userID, unreadOnly, 1, 20)
}

func (r *GormNotificationRepository) GetByUserIDPaginated(userID string, unreadOnly bool, page, pageSize int) ([]*domain.Notification, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("get notifications: user id is required")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := r.db.Where("user_id = ?", userID)
	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	offset := (page - 1) * pageSize
	items := make([]*domain.Notification, 0)
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("get notifications by user id: %w", err)
	}
	return items, nil
}

func (r *GormNotificationRepository) MarkAsRead(id string) error {
	result := r.db.Model(&domain.Notification{}).Where("id = ?", id).Update("is_read", true)
	if result.Error != nil {
		return fmt.Errorf("mark notification as read: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotificationNotFound
	}
	return nil
}

func (r *GormNotificationRepository) MarkAllAsRead(userID string) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("mark all notifications as read: user id is required")
	}
	if err := r.db.Model(&domain.Notification{}).Where("user_id = ?", userID).Update("is_read", true).Error; err != nil {
		return fmt.Errorf("mark all notifications as read: %w", err)
	}
	return nil
}
