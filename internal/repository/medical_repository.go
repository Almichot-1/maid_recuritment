package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

var ErrMedicalDataNotFound = errors.New("medical data not found")

type GormMedicalDataRepository struct {
	db *gorm.DB
}

func NewGormMedicalDataRepository(cfg *config.Config) (domain.MedicalDataRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&domain.MedicalData{}); err != nil {
		return nil, fmt.Errorf("migrate medical_data: %w", err)
	}

	return &GormMedicalDataRepository{db: db}, nil
}

func (r *GormMedicalDataRepository) Upsert(data *domain.MedicalData) error {
	if data == nil {
		return fmt.Errorf("upsert medical data: data is nil")
	}
	if strings.TrimSpace(data.CandidateID) == "" {
		return fmt.Errorf("upsert medical data: candidate_id is required")
	}

	if strings.TrimSpace(data.ID) == "" {
		data.ID = uuid.NewString()
	}
	data.UpdatedAt = time.Now().UTC()

	if err := r.db.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "candidate_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"document_id",
				"issue_date",
				"expiry_date",
				"extracted_at",
				"updated_at",
			}),
		}).
		Create(data).Error; err != nil {
		return fmt.Errorf("upsert medical data: %w", err)
	}

	return nil
}

func (r *GormMedicalDataRepository) GetByCandidateID(candidateID string) (*domain.MedicalData, error) {
	if strings.TrimSpace(candidateID) == "" {
		return nil, fmt.Errorf("get medical data: candidate_id is required")
	}

	var data domain.MedicalData
	if err := r.db.Where("candidate_id = ?", candidateID).First(&data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMedicalDataNotFound
		}
		return nil, fmt.Errorf("get medical data: %w", err)
	}

	return &data, nil
}
