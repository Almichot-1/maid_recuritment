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

var ErrStatusStepNotFound = errors.New("status step not found")

type GormStatusStepRepository struct {
	db *gorm.DB
}

func (r *GormStatusStepRepository) DB() *gorm.DB {
	return r.db
}

func NewGormStatusStepRepository(cfg *config.Config) (*GormStatusStepRepository, error) {
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

	return &GormStatusStepRepository{db: db}, nil
}

func (r *GormStatusStepRepository) Create(step *domain.StatusStep) error {
	if step == nil {
		return fmt.Errorf("create status step: step is nil")
	}
	if strings.TrimSpace(step.CandidateID) == "" {
		return fmt.Errorf("create status step: candidate id is required")
	}
	if strings.TrimSpace(step.StepName) == "" {
		return fmt.Errorf("create status step: step name is required")
	}
	if strings.TrimSpace(step.UpdatedBy) == "" {
		return fmt.Errorf("create status step: updated_by is required")
	}
	if step.ID == "" {
		step.ID = uuid.NewString()
	}
	if step.StepStatus == "" {
		step.StepStatus = domain.Pending
	}
	if step.CreatedAt.IsZero() {
		step.CreatedAt = time.Now().UTC()
	}
	if step.UpdatedAt.IsZero() {
		step.UpdatedAt = step.CreatedAt
	}

	if err := r.db.Create(step).Error; err != nil {
		return fmt.Errorf("create status step: %w", err)
	}
	return nil
}

func (r *GormStatusStepRepository) GetByCandidateID(candidateID string) ([]*domain.StatusStep, error) {
	steps := make([]*domain.StatusStep, 0)
	if err := r.db.Where("candidate_id = ?", candidateID).Order("created_at ASC").Find(&steps).Error; err != nil {
		return nil, fmt.Errorf("get status steps by candidate id: %w", err)
	}
	return steps, nil
}

func (r *GormStatusStepRepository) Update(step *domain.StatusStep) error {
	if step == nil {
		return fmt.Errorf("update status step: step is nil")
	}
	if strings.TrimSpace(step.ID) == "" {
		return fmt.Errorf("update status step: step id is required")
	}

	updates := map[string]any{
		"step_status":  step.StepStatus,
		"completed_at": step.CompletedAt,
		"notes":        step.Notes,
		"updated_by":   step.UpdatedBy,
	}

	result := r.db.Model(&domain.StatusStep{}).Where("id = ?", step.ID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("update status step: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrStatusStepNotFound
	}
	return nil
}
