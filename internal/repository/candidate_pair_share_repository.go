package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

var (
	ErrCandidatePairShareNotFound  = errors.New("candidate pair share not found")
	ErrDuplicateCandidatePairShare = errors.New("candidate already shared to pairing")
)

type GormCandidatePairShareRepository struct {
	db *gorm.DB
}

func (r *GormCandidatePairShareRepository) DB() *gorm.DB {
	return r.db
}

func NewGormCandidatePairShareRepository(cfg *config.Config) (*GormCandidatePairShareRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &GormCandidatePairShareRepository{db: db}, nil
}

func (r *GormCandidatePairShareRepository) Create(share *domain.CandidatePairShare) error {
	if share == nil {
		return fmt.Errorf("create candidate pair share: share is nil")
	}
	if strings.TrimSpace(share.ID) == "" {
		share.ID = uuid.NewString()
	}
	if share.SharedAt.IsZero() {
		share.SharedAt = time.Now().UTC()
	}

	if err := r.db.Create(share).Error; err != nil {
		if isDuplicateCandidatePairShareError(err) {
			return ErrDuplicateCandidatePairShare
		}
		return fmt.Errorf("create candidate pair share: %w", err)
	}

	return nil
}

func (r *GormCandidatePairShareRepository) GetActiveByPairingAndCandidate(pairingID, candidateID string) (*domain.CandidatePairShare, error) {
	var share domain.CandidatePairShare
	err := r.db.
		Where("pairing_id = ? AND candidate_id = ? AND is_active = ?", strings.TrimSpace(pairingID), strings.TrimSpace(candidateID), true).
		First(&share).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCandidatePairShareNotFound
		}
		return nil, fmt.Errorf("get active candidate pair share: %w", err)
	}
	return &share, nil
}

func (r *GormCandidatePairShareRepository) ListByCandidateID(candidateID string, activeOnly bool) ([]*domain.CandidatePairShare, error) {
	query := r.db.Model(&domain.CandidatePairShare{}).Where("candidate_id = ?", strings.TrimSpace(candidateID))
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	shares := make([]*domain.CandidatePairShare, 0)
	if err := query.Order("created_at DESC").Find(&shares).Error; err != nil {
		return nil, fmt.Errorf("list candidate pair shares by candidate: %w", err)
	}
	return shares, nil
}

func (r *GormCandidatePairShareRepository) ListByPairingID(pairingID string, activeOnly bool) ([]*domain.CandidatePairShare, error) {
	query := r.db.Model(&domain.CandidatePairShare{}).Where("pairing_id = ?", strings.TrimSpace(pairingID))
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	shares := make([]*domain.CandidatePairShare, 0)
	if err := query.Order("created_at DESC").Find(&shares).Error; err != nil {
		return nil, fmt.Errorf("list candidate pair shares by pairing: %w", err)
	}
	return shares, nil
}

func (r *GormCandidatePairShareRepository) Deactivate(pairingID, candidateID string, unsharedAt time.Time) error {
	result := r.db.Model(&domain.CandidatePairShare{}).
		Where("pairing_id = ? AND candidate_id = ? AND is_active = ?", strings.TrimSpace(pairingID), strings.TrimSpace(candidateID), true).
		Updates(map[string]any{
			"is_active":   false,
			"unshared_at": unsharedAt,
		})
	if result.Error != nil {
		return fmt.Errorf("deactivate candidate pair share: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrCandidatePairShareNotFound
	}
	return nil
}

func isDuplicateCandidatePairShareError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && strings.Contains(pgErr.Message, "idx_candidate_pair_shares_unique_active")
	}
	return false
}
