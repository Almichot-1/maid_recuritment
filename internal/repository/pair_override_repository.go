package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"maid-recruitment-tracking/internal/domain"
)

// GormCandidatePairOverrideRepository implements domain.CandidatePairOverrideRepository
// using GORM + PostgreSQL.
type GormCandidatePairOverrideRepository struct {
	db *gorm.DB
}

func NewGormCandidatePairOverrideRepository(db *gorm.DB) *GormCandidatePairOverrideRepository {
	return &GormCandidatePairOverrideRepository{db: db}
}

// Upsert inserts or updates the override row for (pairing_id, candidate_id).
// On conflict it updates country_applied, salary_offered, and updated_at.
func (r *GormCandidatePairOverrideRepository) Upsert(override *domain.CandidatePairOverride) error {
	if override == nil {
		return fmt.Errorf("upsert pair override: override is nil")
	}
	if strings.TrimSpace(override.PairingID) == "" {
		return fmt.Errorf("upsert pair override: pairing_id is required")
	}
	if strings.TrimSpace(override.CandidateID) == "" {
		return fmt.Errorf("upsert pair override: candidate_id is required")
	}

	override.UpdatedAt = time.Now().UTC()

	result := r.db.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "pairing_id"},
				{Name: "candidate_id"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"country_applied",
				"salary_offered",
				"updated_at",
			}),
		}).
		Create(override)

	if result.Error != nil {
		return fmt.Errorf("upsert pair override: %w", result.Error)
	}
	return nil
}

// GetByPairingAndCandidate fetches the override for the given pairing + candidate.
// Returns nil, nil when no override exists (not an error – callers fall back to global defaults).
func (r *GormCandidatePairOverrideRepository) GetByPairingAndCandidate(pairingID, candidateID string) (*domain.CandidatePairOverride, error) {
	if strings.TrimSpace(pairingID) == "" || strings.TrimSpace(candidateID) == "" {
		return nil, fmt.Errorf("get pair override: pairing_id and candidate_id are required")
	}

	var override domain.CandidatePairOverride
	err := r.db.
		Where("pairing_id = ? AND candidate_id = ?", pairingID, candidateID).
		First(&override).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // no override – caller should fall back to global default
		}
		return nil, fmt.Errorf("get pair override: %w", err)
	}
	return &override, nil
}

// ListByCandidateID returns all overrides for a candidate across every pairing.
// Used to render the per-partner override table on the candidate detail page.
func (r *GormCandidatePairOverrideRepository) ListByCandidateID(candidateID string) ([]*domain.CandidatePairOverride, error) {
	if strings.TrimSpace(candidateID) == "" {
		return nil, fmt.Errorf("list pair overrides: candidate_id is required")
	}

	var overrides []*domain.CandidatePairOverride
	if err := r.db.
		Where("candidate_id = ?", candidateID).
		Order("created_at ASC").
		Find(&overrides).Error; err != nil {
		return nil, fmt.Errorf("list pair overrides: %w", err)
	}
	return overrides, nil
}

const pairOverrideBatchSize = 100

// BulkUpsert inserts or updates overrides for multiple candidates in one batch.
// Rows are chunked into batches of 100 to stay within SQL parameter limits.
func (r *GormCandidatePairOverrideRepository) BulkUpsert(overrides []*domain.CandidatePairOverride) error {
	if len(overrides) == 0 {
		return nil
	}

	now := time.Now().UTC()
	for _, o := range overrides {
		o.UpdatedAt = now
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}
	}

	for i := 0; i < len(overrides); i += pairOverrideBatchSize {
		end := i + pairOverrideBatchSize
		if end > len(overrides) {
			end = len(overrides)
		}
		batch := overrides[i:end]
		if err := r.db.
			Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "pairing_id"},
					{Name: "candidate_id"},
				},
				DoUpdates: clause.AssignmentColumns([]string{
					"country_applied", "salary_offered", "updated_at",
				}),
			}).
			Create(batch).Error; err != nil {
			return fmt.Errorf("bulk upsert pair overrides: %w", err)
		}
	}
	return nil
}
