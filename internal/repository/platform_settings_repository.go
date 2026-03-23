package repository

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

var ErrPlatformSettingsNotFound = errors.New("platform settings not found")

type GormPlatformSettingsRepository struct {
	db *gorm.DB
}

func NewGormPlatformSettingsRepository(cfg *config.Config) (*GormPlatformSettingsRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &GormPlatformSettingsRepository{db: db}, nil
}

func (r *GormPlatformSettingsRepository) Get() (*domain.PlatformSettings, error) {
	var settings domain.PlatformSettings
	if err := r.db.Where("id = ?", domain.PlatformSettingsSingletonID).First(&settings).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPlatformSettingsNotFound
		}
		return nil, fmt.Errorf("get platform settings: %w", err)
	}
	return &settings, nil
}

func (r *GormPlatformSettingsRepository) Upsert(settings *domain.PlatformSettings) error {
	if settings == nil {
		return fmt.Errorf("upsert platform settings: settings is nil")
	}
	if strings.TrimSpace(settings.ID) == "" {
		settings.ID = domain.PlatformSettingsSingletonID
	}
	if err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(settings).Error; err != nil {
		return fmt.Errorf("upsert platform settings: %w", err)
	}
	return nil
}
