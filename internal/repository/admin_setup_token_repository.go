package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
	"gorm.io/gorm"
)

var ErrAdminSetupTokenNotFound = errors.New("admin setup token not found")

type GormAdminSetupTokenRepository struct {
	db *gorm.DB
}

func NewGormAdminSetupTokenRepository(cfg *config.Config) (*GormAdminSetupTokenRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}
	return &GormAdminSetupTokenRepository{db: db}, nil
}

func (r *GormAdminSetupTokenRepository) Create(token *domain.AdminSetupToken) error {
	if token == nil {
		return fmt.Errorf("create admin setup token: token is nil")
	}
	if err := r.db.Create(token).Error; err != nil {
		return fmt.Errorf("create admin setup token: %w", err)
	}
	return nil
}

func (r *GormAdminSetupTokenRepository) GetActiveByHash(tokenHash string) (*domain.AdminSetupToken, error) {
	var token domain.AdminSetupToken
	now := time.Now().UTC()
	if err := r.db.
		Where("token_hash = ?", strings.TrimSpace(tokenHash)).
		Where("used_at IS NULL").
		Where("expires_at > ?", now).
		First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminSetupTokenNotFound
		}
		return nil, fmt.Errorf("get admin setup token: %w", err)
	}
	return &token, nil
}

func (r *GormAdminSetupTokenRepository) InvalidateByAdminID(adminID string) error {
	return r.db.Model(&domain.AdminSetupToken{}).
		Where("admin_id = ? AND used_at IS NULL", strings.TrimSpace(adminID)).
		Update("used_at", time.Now().UTC()).
		Error
}

func (r *GormAdminSetupTokenRepository) MarkUsed(id string, usedAt time.Time) error {
	result := r.db.Model(&domain.AdminSetupToken{}).
		Where("id = ?", strings.TrimSpace(id)).
		Update("used_at", usedAt.UTC())
	if result.Error != nil {
		return fmt.Errorf("mark admin setup token used: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrAdminSetupTokenNotFound
	}
	return nil
}
