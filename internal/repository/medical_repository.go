package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

var ErrMedicalDataNotFound = errors.New("medical data not found")

type GormMedicalDataRepository struct {
	db *gorm.DB
}

func (r *GormMedicalDataRepository) DB() *gorm.DB {
	return r.db
}

func NewGormMedicalDataRepository(cfg *config.Config) (*GormMedicalDataRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &GormMedicalDataRepository{db: db}, nil
}

func (r *GormMedicalDataRepository) Upsert(data *domain.MedicalData) error {
	if data == nil {
		return fmt.Errorf("upsert medical data: data is nil")
	}
	if strings.TrimSpace(data.CandidateID) == "" {
		return fmt.Errorf("upsert medical data: candidate id is required")
	}
	if strings.TrimSpace(data.DocumentID) == "" {
		return fmt.Errorf("upsert medical data: document id is required")
	}

	if err := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "candidate_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"document_id":        data.DocumentID,
			"expiry_date":        data.ExpiryDate,
			"raw_text":           data.RawText,
			"extracted_at":       data.ExtractedAt,
			"warning_sent_flags": data.WarningSentFlags,
			"updated_at":         time.Now().UTC(),
		}),
	}).Create(data).Error; err != nil {
		return fmt.Errorf("upsert medical data: %w", err)
	}

	return nil
}

func (r *GormMedicalDataRepository) GetByCandidateID(candidateID string) (*domain.MedicalData, error) {
	var data domain.MedicalData
	if err := r.db.Where("candidate_id = ?", strings.TrimSpace(candidateID)).First(&data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMedicalDataNotFound
		}
		return nil, fmt.Errorf("get medical data by candidate id: %w", err)
	}
	return &data, nil
}

func (r *GormMedicalDataRepository) GetExpiringMedical(days int) ([]*domain.MedicalData, error) {
	if days <= 0 {
		return []*domain.MedicalData{}, nil
	}

	now := time.Now().UTC()
	deadline := now.AddDate(0, 0, days)

	results := make([]*domain.MedicalData, 0)
	if err := r.db.
		Where("expiry_date >= ? AND expiry_date <= ?", now, deadline).
		Order("expiry_date ASC").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("get expiring medical data: %w", err)
	}
	return results, nil
}

func (r *GormMedicalDataRepository) DeleteByCandidateID(candidateID string) error {
	result := r.db.Where("candidate_id = ?", strings.TrimSpace(candidateID)).Delete(&domain.MedicalData{})
	if result.Error != nil {
		return fmt.Errorf("delete medical data by candidate id: %w", result.Error)
	}
	return nil
}
