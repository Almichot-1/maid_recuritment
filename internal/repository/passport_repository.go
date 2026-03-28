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

var ErrPassportDataNotFound = errors.New("passport data not found")

type GormPassportDataRepository struct {
	db *gorm.DB
}

func (r *GormPassportDataRepository) DB() *gorm.DB {
	return r.db
}

func NewGormPassportDataRepository(cfg *config.Config) (*GormPassportDataRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &GormPassportDataRepository{db: db}, nil
}

func (r *GormPassportDataRepository) Upsert(data *domain.PassportData) error {
	if data == nil {
		return fmt.Errorf("upsert passport data: data is nil")
	}
	if strings.TrimSpace(data.CandidateID) == "" {
		return fmt.Errorf("upsert passport data: candidate id is required")
	}

	if err := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "candidate_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"holder_name":                 data.HolderName,
			"passport_number":             data.PassportNumber,
			"country_code":                data.CountryCode,
			"nationality":                 data.Nationality,
			"date_of_birth":               data.DateOfBirth,
			"place_of_birth":              data.PlaceOfBirth,
			"gender":                      data.Gender,
			"issue_date":                  data.IssueDate,
			"expiry_date":                 data.ExpiryDate,
			"issuing_authority":           data.IssuingAuthority,
			"mrz_line_1":                  data.MRZLine1,
			"mrz_line_2":                  data.MRZLine2,
			"confidence":                  data.Confidence,
			"extracted_at":                data.ExtractedAt,
			"passport_warning_sent_flags": data.PassportWarningSentFlags,
			"updated_at":                  time.Now().UTC(),
		}),
	}).Create(data).Error; err != nil {
		return fmt.Errorf("upsert passport data: %w", err)
	}

	return nil
}

func (r *GormPassportDataRepository) GetByCandidateID(candidateID string) (*domain.PassportData, error) {
	var data domain.PassportData
	if err := r.db.Where("candidate_id = ?", strings.TrimSpace(candidateID)).First(&data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPassportDataNotFound
		}
		return nil, fmt.Errorf("get passport data by candidate id: %w", err)
	}
	return &data, nil
}

func (r *GormPassportDataRepository) GetExpiringPassports(days int) ([]*domain.PassportData, error) {
	if days <= 0 {
		return []*domain.PassportData{}, nil
	}
	now := time.Now().UTC()
	deadline := now.AddDate(0, 0, days)

	results := make([]*domain.PassportData, 0)
	if err := r.db.
		Where("expiry_date >= ? AND expiry_date <= ?", now, deadline).
		Order("expiry_date ASC").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("get expiring passports: %w", err)
	}
	return results, nil
}
