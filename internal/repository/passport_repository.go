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

var ErrPassportDataNotFound = errors.New("passport data not found")

type GormPassportRepository struct {
	db *gorm.DB
}

func NewGormPassportDataRepository(cfg *config.Config) (domain.PassportDataRepository, error) {
	return NewGormPassportRepository(cfg)
}

func NewGormPassportRepository(cfg *config.Config) (*GormPassportRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}
	return &GormPassportRepository{db: db}, nil
}

// Upsert inserts a new PassportData row or updates the existing one for the
// same candidate_id. The unique index on candidate_id enforces one row per
// candidate. All non-zero fields in data are written.
func (r *GormPassportRepository) Upsert(data *domain.PassportData) error {
	if data == nil {
		return fmt.Errorf("upsert passport data: data is nil")
	}
	if strings.TrimSpace(data.CandidateID) == "" {
		return fmt.Errorf("upsert passport data: candidate_id is required")
	}

	if strings.TrimSpace(data.ID) == "" {
		data.ID = uuid.NewString()
	}
	data.UpdatedAt = time.Now().UTC()

	if err := r.db.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "candidate_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"holder_name",
				"passport_number",
				"nationality",
				"country_code",
				"date_of_birth",
				"place_of_birth",
				"gender",
				"issue_date",
				"expiry_date",
				"place_of_issue",
				"issuing_authority",
				"mrz_line1",
				"mrz_line2",
				"confidence",
				"extracted_at",
				"updated_at",
			}),
		}).
		Create(data).Error; err != nil {
		return fmt.Errorf("upsert passport data: %w", err)
	}

	return nil
}

// GetByCandidateID returns the PassportData for the given candidate.
// Returns ErrPassportDataNotFound when none exists.
func (r *GormPassportRepository) GetByCandidateID(candidateID string) (*domain.PassportData, error) {
	if strings.TrimSpace(candidateID) == "" {
		return nil, fmt.Errorf("get passport data: candidate_id is required")
	}

	var data domain.PassportData
	if err := r.db.
		Where("candidate_id = ?", candidateID).
		First(&data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPassportDataNotFound
		}
		return nil, fmt.Errorf("get passport data by candidate id: %w", err)
	}
	return &data, nil
}

// GetExpiringPassports returns PassportData rows whose expiry_date falls
// within the next daysAhead days and whose passport_warning_sent_flags does
// NOT yet have flagBit set. Only rows with a non-null expiry_date are returned.
func (r *GormPassportRepository) GetExpiringPassports(daysAhead int, flagBit int) ([]*domain.PassportData, error) {
	if daysAhead <= 0 {
		daysAhead = 30
	}

	cutoff := time.Now().UTC().AddDate(0, 0, daysAhead)
	rows := make([]*domain.PassportData, 0)

	if err := r.db.
		Where(
			"expiry_date IS NOT NULL AND expiry_date <= ? AND (passport_warning_sent_flags & ?) = 0",
			cutoff,
			flagBit,
		).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("get expiring passports: %w", err)
	}
	return rows, nil
}

// UpdateWarningSentFlags sets the passport_warning_sent_flags column for the
// given PassportData ID.
func (r *GormPassportRepository) UpdateWarningSentFlags(id string, flags int) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("update passport warning flags: id is required")
	}

	result := r.db.Model(&domain.PassportData{}).
		Where("id = ?", id).
		Update("passport_warning_sent_flags", flags)
	if result.Error != nil {
		return fmt.Errorf("update passport warning flags: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrPassportDataNotFound
	}
	return nil
}
