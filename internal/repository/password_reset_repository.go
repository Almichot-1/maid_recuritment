package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

var ErrPasswordResetRequestNotFound = errors.New("password reset request not found")

type GormPasswordResetRequestRepository struct {
	db *gorm.DB
}

func NewGormPasswordResetRequestRepository(cfg *config.Config) (*GormPasswordResetRequestRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &GormPasswordResetRequestRepository{db: db}, nil
}

func (r *GormPasswordResetRequestRepository) Create(request *domain.PasswordResetRequest) error {
	if request == nil {
		return fmt.Errorf("create password reset request: request is nil")
	}
	if strings.TrimSpace(request.UserID) == "" {
		return fmt.Errorf("create password reset request: user id is required")
	}
	if strings.TrimSpace(request.CodeHash) == "" {
		return fmt.Errorf("create password reset request: code hash is required")
	}
	if request.ExpiresAt.IsZero() {
		return fmt.Errorf("create password reset request: expires at is required")
	}

	if strings.TrimSpace(request.ID) == "" {
		request.ID = uuid.NewString()
	}

	if err := r.db.Create(request).Error; err != nil {
		return fmt.Errorf("create password reset request: %w", err)
	}
	return nil
}

func (r *GormPasswordResetRequestRepository) GetLatestByUserID(userID string) (*domain.PasswordResetRequest, error) {
	var request domain.PasswordResetRequest
	if err := r.db.Where("user_id = ?", strings.TrimSpace(userID)).Order("created_at DESC").First(&request).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPasswordResetRequestNotFound
		}
		return nil, fmt.Errorf("get latest password reset request: %w", err)
	}
	return &request, nil
}

func (r *GormPasswordResetRequestRepository) GetLatestActiveByUserID(userID string) (*domain.PasswordResetRequest, error) {
	var request domain.PasswordResetRequest
	if err := r.db.
		Where("user_id = ? AND used_at IS NULL AND expires_at > ?", strings.TrimSpace(userID), time.Now().UTC()).
		Order("created_at DESC").
		First(&request).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPasswordResetRequestNotFound
		}
		return nil, fmt.Errorf("get latest active password reset request: %w", err)
	}
	return &request, nil
}

func (r *GormPasswordResetRequestRepository) IncrementAttempts(id string) error {
	result := r.db.Model(&domain.PasswordResetRequest{}).
		Where("id = ?", strings.TrimSpace(id)).
		UpdateColumn("attempt_count", gorm.Expr("attempt_count + 1"))
	if result.Error != nil {
		return fmt.Errorf("increment password reset attempts: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrPasswordResetRequestNotFound
	}
	return nil
}

func (r *GormPasswordResetRequestRepository) MarkUsed(id string) error {
	now := time.Now().UTC()
	result := r.db.Model(&domain.PasswordResetRequest{}).
		Where("id = ?", strings.TrimSpace(id)).
		Update("used_at", now)
	if result.Error != nil {
		return fmt.Errorf("mark password reset request used: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrPasswordResetRequestNotFound
	}
	return nil
}

func (r *GormPasswordResetRequestRepository) InvalidateActiveByUserID(userID string) error {
	now := time.Now().UTC()
	if err := r.db.Model(&domain.PasswordResetRequest{}).
		Where("user_id = ? AND used_at IS NULL AND expires_at > ?", strings.TrimSpace(userID), now).
		Update("used_at", now).Error; err != nil {
		return fmt.Errorf("invalidate active password reset requests: %w", err)
	}
	return nil
}

func (r *GormPasswordResetRequestRepository) Delete(id string) error {
	result := r.db.Where("id = ?", strings.TrimSpace(id)).Delete(&domain.PasswordResetRequest{})
	if result.Error != nil {
		return fmt.Errorf("delete password reset request: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrPasswordResetRequestNotFound
	}
	return nil
}
