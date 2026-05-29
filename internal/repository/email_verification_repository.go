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

var ErrEmailVerificationTokenNotFound = errors.New("email verification token not found")

type GormEmailVerificationTokenRepository struct {
	db *gorm.DB
}

func NewGormEmailVerificationTokenRepository(cfg *config.Config) (*GormEmailVerificationTokenRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &GormEmailVerificationTokenRepository{db: db}, nil
}

func (r *GormEmailVerificationTokenRepository) Create(token *domain.EmailVerificationToken) error {
	if token == nil {
		return fmt.Errorf("create email verification token: token is nil")
	}
	if strings.TrimSpace(token.UserID) == "" {
		return fmt.Errorf("create email verification token: user id is required")
	}
	if strings.TrimSpace(token.TokenHash) == "" {
		return fmt.Errorf("create email verification token: token hash is required")
	}
	if token.ExpiresAt.IsZero() {
		return fmt.Errorf("create email verification token: expires at is required")
	}
	if strings.TrimSpace(token.ID) == "" {
		token.ID = uuid.NewString()
	}

	if err := r.db.Create(token).Error; err != nil {
		return fmt.Errorf("create email verification token: %w", err)
	}
	return nil
}

func (r *GormEmailVerificationTokenRepository) GetActiveByTokenHash(tokenHash string) (*domain.EmailVerificationToken, error) {
	var token domain.EmailVerificationToken
	if err := r.db.
		Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", strings.TrimSpace(tokenHash), time.Now().UTC()).
		Order("created_at DESC").
		First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmailVerificationTokenNotFound
		}
		return nil, fmt.Errorf("get active email verification token: %w", err)
	}
	return &token, nil
}

func (r *GormEmailVerificationTokenRepository) GetLatestByUserID(userID string) (*domain.EmailVerificationToken, error) {
	var token domain.EmailVerificationToken
	if err := r.db.
		Where("user_id = ?", strings.TrimSpace(userID)).
		Order("created_at DESC").
		First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmailVerificationTokenNotFound
		}
		return nil, fmt.Errorf("get latest email verification token: %w", err)
	}
	return &token, nil
}

func (r *GormEmailVerificationTokenRepository) InvalidateActiveByUserID(userID string) error {
	now := time.Now().UTC()
	if err := r.db.Model(&domain.EmailVerificationToken{}).
		Where("user_id = ? AND used_at IS NULL AND expires_at > ?", strings.TrimSpace(userID), now).
		Update("used_at", now).Error; err != nil {
		return fmt.Errorf("invalidate active email verification tokens: %w", err)
	}
	return nil
}

func (r *GormEmailVerificationTokenRepository) MarkUsed(id string) error {
	now := time.Now().UTC()
	result := r.db.Model(&domain.EmailVerificationToken{}).
		Where("id = ?", strings.TrimSpace(id)).
		Update("used_at", now)
	if result.Error != nil {
		return fmt.Errorf("mark email verification token used: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrEmailVerificationTokenNotFound
	}
	return nil
}

func (r *GormEmailVerificationTokenRepository) Delete(id string) error {
	result := r.db.Where("id = ?", strings.TrimSpace(id)).Delete(&domain.EmailVerificationToken{})
	if result.Error != nil {
		return fmt.Errorf("delete email verification token: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrEmailVerificationTokenNotFound
	}
	return nil
}
