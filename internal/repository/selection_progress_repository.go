package repository

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

var ErrProgressNotFound = errors.New("selection progress not found")

type GormSelectionProgressRepository struct {
	db *gorm.DB
}

func (r *GormSelectionProgressRepository) DB() *gorm.DB {
	return r.db
}

func NewGormSelectionProgressRepository(cfg *config.Config) (*GormSelectionProgressRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}
	return &GormSelectionProgressRepository{db: db}, nil
}

func (r *GormSelectionProgressRepository) Create(progress *domain.SelectionProgress) error {
	if progress == nil {
		return fmt.Errorf("create progress: progress is nil")
	}
	if strings.TrimSpace(progress.SelectionID) == "" {
		return fmt.Errorf("create progress: selection_id is required")
	}
	if strings.TrimSpace(progress.UpdatedBy) == "" {
		return fmt.Errorf("create progress: updated_by is required")
	}

	return r.db.Create(progress).Error
}

func (r *GormSelectionProgressRepository) GetByID(id string) (*domain.SelectionProgress, error) {
	var progress domain.SelectionProgress
	if err := r.db.Where("id = ?", id).First(&progress).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProgressNotFound
		}
		return nil, fmt.Errorf("get progress by id: %w", err)
	}
	return &progress, nil
}

func (r *GormSelectionProgressRepository) GetBySelectionID(selectionID string) (*domain.SelectionProgress, error) {
	var progress domain.SelectionProgress
	if err := r.db.Where("selection_id = ?", selectionID).First(&progress).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProgressNotFound
		}
		return nil, fmt.Errorf("get progress by selection id: %w", err)
	}
	return &progress, nil
}

func (r *GormSelectionProgressRepository) Update(progress *domain.SelectionProgress) error {
	if progress == nil {
		return fmt.Errorf("update progress: progress is nil")
	}
	return r.db.Save(progress).Error
}

func (r *GormSelectionProgressRepository) Delete(selectionID string) error {
	return r.db.Where("selection_id = ?", selectionID).Delete(&domain.SelectionProgress{}).Error
}
