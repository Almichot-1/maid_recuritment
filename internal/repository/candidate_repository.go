package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

var (
	ErrCandidateNotFound       = errors.New("candidate not found")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

type GormCandidateRepository struct {
	db *gorm.DB
}

func (r *GormCandidateRepository) DB() *gorm.DB {
	return r.db
}

func NewGormCandidateRepository(cfg *config.Config) (*GormCandidateRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &GormCandidateRepository{db: db}, nil
}

func (r *GormCandidateRepository) Create(candidate *domain.Candidate) error {
	if candidate == nil {
		return fmt.Errorf("create candidate: candidate is nil")
	}

	if candidate.ID == "" {
		candidate.ID = uuid.NewString()
	}
	if candidate.Status == "" {
		candidate.Status = domain.CandidateStatusDraft
	}

	if err := r.db.Create(candidate).Error; err != nil {
		return fmt.Errorf("create candidate: %w", err)
	}

	return nil
}

func (r *GormCandidateRepository) GetByID(id string) (*domain.Candidate, error) {
	var candidate domain.Candidate
	if err := r.db.Preload("Documents").Where("id = ?", id).First(&candidate).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCandidateNotFound
		}
		return nil, fmt.Errorf("get candidate by id: %w", err)
	}

	return &candidate, nil
}

func (r *GormCandidateRepository) List(filters domain.CandidateFilters) ([]*domain.Candidate, error) {
	query, err := applyCandidateFilters(r.db.Model(&domain.Candidate{}), filters)
	if err != nil {
		return nil, err
	}

	page := filters.Page
	if page <= 0 {
		page = 1
	}

	pageSize := filters.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	var candidates []*domain.Candidate
	if err := query.Preload("Documents").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&candidates).Error; err != nil {
		return nil, fmt.Errorf("list candidates: %w", err)
	}

	return candidates, nil
}

func (r *GormCandidateRepository) Count(filters domain.CandidateFilters) (int64, error) {
	query, err := applyCandidateFilters(r.db.Model(&domain.Candidate{}), filters)
	if err != nil {
		return 0, err
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count candidates: %w", err)
	}

	return total, nil
}

func (r *GormCandidateRepository) Update(candidate *domain.Candidate) error {
	if candidate == nil {
		return fmt.Errorf("update candidate: candidate is nil")
	}
	if strings.TrimSpace(candidate.ID) == "" {
		return fmt.Errorf("update candidate: id is required")
	}

	var existing domain.Candidate
	if err := r.db.Where("id = ?", candidate.ID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCandidateNotFound
		}
		return fmt.Errorf("update candidate: load current candidate: %w", err)
	}

	if !isValidStatusTransition(existing.Status, candidate.Status) {
		return ErrInvalidStatusTransition
	}

	updates := map[string]any{
		"full_name":        candidate.FullName,
		"nationality":      candidate.Nationality,
		"date_of_birth":    candidate.DateOfBirth,
		"age":              candidate.Age,
		"place_of_birth":   candidate.PlaceOfBirth,
		"religion":         candidate.Religion,
		"marital_status":   candidate.MaritalStatus,
		"children_count":   candidate.ChildrenCount,
		"education_level":  candidate.EducationLevel,
		"experience_years": candidate.ExperienceYears,
		"languages":        candidate.Languages,
		"skills":           candidate.Skills,
		"status":           candidate.Status,
		"locked_by":        candidate.LockedBy,
		"locked_at":        candidate.LockedAt,
		"lock_expires_at":  candidate.LockExpiresAt,
		"cv_pdf_url":       candidate.CVPDFURL,
	}

	result := r.db.Model(&domain.Candidate{}).Where("id = ?", candidate.ID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("update candidate: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrCandidateNotFound
	}

	return nil
}

func (r *GormCandidateRepository) Delete(id string) error {
	var candidate domain.Candidate
	if err := r.db.Where("id = ?", id).First(&candidate).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCandidateNotFound
		}
		return fmt.Errorf("delete candidate: load candidate: %w", err)
	}

	if candidate.Status != domain.CandidateStatusDraft && candidate.Status != domain.CandidateStatusAvailable {
		return fmt.Errorf("delete candidate: only draft or available candidates can be deleted")
	}

	if err := r.db.Delete(&domain.Candidate{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete candidate: %w", err)
	}

	return nil
}

func (r *GormCandidateRepository) Lock(candidateID, lockedBy string, expiresAt time.Time) error {
	if strings.TrimSpace(candidateID) == "" {
		return fmt.Errorf("lock candidate: candidate id is required")
	}
	if strings.TrimSpace(lockedBy) == "" {
		return fmt.Errorf("lock candidate: locked by is required")
	}

	now := time.Now().UTC()

	result := r.db.Model(&domain.Candidate{}).
		Where("id = ?", candidateID).
		Where("status = ?", domain.CandidateStatusAvailable).
		Updates(map[string]any{
			"status":          domain.CandidateStatusLocked,
			"locked_by":       lockedBy,
			"locked_at":       now,
			"lock_expires_at": expiresAt,
		})
	if result.Error != nil {
		return fmt.Errorf("lock candidate: %w", result.Error)
	}
	if result.RowsAffected > 0 {
		return nil
	}

	var current domain.Candidate
	if err := r.db.Unscoped().Select("id", "status").Where("id = ?", candidateID).First(&current).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCandidateNotFound
		}
		return fmt.Errorf("lock candidate: check candidate existence: %w", err)
	}

	return ErrInvalidStatusTransition
}

func (r *GormCandidateRepository) Unlock(candidateID string) error {
	if strings.TrimSpace(candidateID) == "" {
		return fmt.Errorf("unlock candidate: candidate id is required")
	}

	result := r.db.Model(&domain.Candidate{}).
		Where("id = ?", candidateID).
		Where("status = ?", domain.CandidateStatusLocked).
		Updates(map[string]any{
			"status":          domain.CandidateStatusAvailable,
			"locked_by":       nil,
			"locked_at":       nil,
			"lock_expires_at": nil,
		})
	if result.Error != nil {
		return fmt.Errorf("unlock candidate: %w", result.Error)
	}
	if result.RowsAffected > 0 {
		return nil
	}

	var current domain.Candidate
	if err := r.db.Unscoped().Select("id", "status").Where("id = ?", candidateID).First(&current).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCandidateNotFound
		}
		return fmt.Errorf("unlock candidate: check candidate existence: %w", err)
	}

	return ErrInvalidStatusTransition
}

func isValidStatusTransition(from, to domain.CandidateStatus) bool {
	if from == "" || to == "" {
		return false
	}
	if from == to {
		return true
	}

	allowedTransitions := map[domain.CandidateStatus]map[domain.CandidateStatus]struct{}{
		domain.CandidateStatusDraft: {
			domain.CandidateStatusAvailable: {},
		},
		domain.CandidateStatusAvailable: {
			domain.CandidateStatusLocked:      {},
			domain.CandidateStatusUnderReview: {},
		},
		domain.CandidateStatusLocked: {
			domain.CandidateStatusAvailable:   {},
			domain.CandidateStatusUnderReview: {},
		},
		domain.CandidateStatusUnderReview: {
			domain.CandidateStatusApproved: {},
			domain.CandidateStatusRejected: {},
		},
		domain.CandidateStatusApproved: {
			domain.CandidateStatusInProgress: {},
		},
		domain.CandidateStatusInProgress: {
			domain.CandidateStatusCompleted: {},
		},
	}

	nextStatuses, ok := allowedTransitions[from]
	if !ok {
		return false
	}

	_, allowed := nextStatuses[to]
	return allowed
}

func applyCandidateFilters(query *gorm.DB, filters domain.CandidateFilters) (*gorm.DB, error) {
	if strings.TrimSpace(filters.PairingID) != "" {
		query = query.Where(
			`EXISTS (
				SELECT 1
				FROM candidate_pair_shares
				WHERE candidate_pair_shares.candidate_id = candidates.id
					AND candidate_pair_shares.pairing_id = ?
					AND candidate_pair_shares.is_active = ?
			)`,
			strings.TrimSpace(filters.PairingID),
			true,
		)
	} else if filters.SharedOnly {
		query = query.Where(
			`EXISTS (
				SELECT 1
				FROM candidate_pair_shares
				WHERE candidate_pair_shares.candidate_id = candidates.id
					AND candidate_pair_shares.is_active = ?
			)`,
			true,
		)
	}

	if len(filters.Statuses) > 0 {
		query = query.Where("status IN ?", filters.Statuses)
	}
	if filters.MinAge != nil {
		query = query.Where("age >= ?", *filters.MinAge)
	}
	if filters.MaxAge != nil {
		query = query.Where("age <= ?", *filters.MaxAge)
	}
	if filters.MinExperience != nil {
		query = query.Where("experience_years >= ?", *filters.MinExperience)
	}
	if filters.MaxExperience != nil {
		query = query.Where("experience_years <= ?", *filters.MaxExperience)
	}
	if trimmedSearch := strings.TrimSpace(filters.Search); trimmedSearch != "" {
		query = query.Where("LOWER(full_name) LIKE ?", "%"+strings.ToLower(trimmedSearch)+"%")
	}
	for _, language := range filters.Languages {
		trimmedLanguage := strings.TrimSpace(language)
		if trimmedLanguage == "" {
			continue
		}

		languageJSON, err := json.Marshal([]string{trimmedLanguage})
		if err != nil {
			return nil, fmt.Errorf("list candidates: marshal language filter: %w", err)
		}

		query = query.Where("languages @> ?::jsonb", string(languageJSON))
	}
	if strings.TrimSpace(filters.CreatedBy) != "" {
		query = query.Where("created_by = ?", strings.TrimSpace(filters.CreatedBy))
	}

	return query, nil
}
