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
	ErrSelectionNotFound      = errors.New("selection not found")
	ErrActiveSelectionExists  = errors.New("active pending selection already exists for candidate")
	ErrInvalidSelectionStatus = errors.New("invalid selection status")
)

type GormSelectionRepository struct {
	db *gorm.DB
}

func (r *GormSelectionRepository) DB() *gorm.DB {
	return r.db
}

func NewGormSelectionRepository(cfg *config.Config) (*GormSelectionRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &GormSelectionRepository{db: db}, nil
}

func (r *GormSelectionRepository) Create(selection *domain.Selection) error {
	if selection == nil {
		return fmt.Errorf("create selection: selection is nil")
	}
	if strings.TrimSpace(selection.CandidateID) == "" {
		return fmt.Errorf("create selection: candidate id is required")
	}
	if strings.TrimSpace(selection.SelectedBy) == "" {
		return fmt.Errorf("create selection: selected_by is required")
	}

	if selection.ID == "" {
		selection.ID = uuid.NewString()
	}
	if selection.Status == "" {
		selection.Status = domain.SelectionPending
	}
	if !isValidSelectionStatus(selection.Status) {
		return ErrInvalidSelectionStatus
	}
	if selection.ExpiresAt.IsZero() {
		selection.ExpiresAt = time.Now().UTC().Add(24 * time.Hour)
	}

	if err := r.db.Select("ID", "CandidateID", "PairingID", "SelectedBy", "Status", "ExpiresAt").Create(selection).Error; err != nil {
		if isSelectionPendingUniqueViolation(err) {
			return ErrActiveSelectionExists
		}
		return fmt.Errorf("create selection: %w", err)
	}

	return nil
}

func (r *GormSelectionRepository) GetByID(id string) (*domain.Selection, error) {
	var selection domain.Selection
	if err := r.db.Where("id = ?", id).First(&selection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSelectionNotFound
		}
		return nil, fmt.Errorf("get selection by id: %w", err)
	}
	return &selection, nil
}

func (r *GormSelectionRepository) GetByCandidateID(candidateID string) (*domain.Selection, error) {
	var selection domain.Selection
	if err := r.db.Where("candidate_id = ?", candidateID).Order("created_at DESC").First(&selection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSelectionNotFound
		}
		return nil, fmt.Errorf("get selection by candidate id: %w", err)
	}
	return &selection, nil
}

func (r *GormSelectionRepository) GetByCandidateIDAndPairingID(candidateID, pairingID string) (*domain.Selection, error) {
	var selection domain.Selection
	if err := r.db.Where("candidate_id = ? AND pairing_id = ?", candidateID, pairingID).Order("created_at DESC").First(&selection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSelectionNotFound
		}
		return nil, fmt.Errorf("get selection by candidate id and pairing id: %w", err)
	}
	return &selection, nil
}

func (r *GormSelectionRepository) GetBySelectedBy(userID string) ([]*domain.Selection, error) {
	selections := make([]*domain.Selection, 0)
	if err := r.db.Where("selected_by = ?", userID).Order("created_at DESC").Find(&selections).Error; err != nil {
		return nil, fmt.Errorf("get selections by selected_by: %w", err)
	}
	return selections, nil
}

func (r *GormSelectionRepository) GetBySelectedByAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	selections := make([]*domain.Selection, 0)
	if err := r.db.Where("selected_by = ? AND pairing_id = ?", userID, pairingID).Order("created_at DESC").Find(&selections).Error; err != nil {
		return nil, fmt.Errorf("get selections by selected_by and pairing: %w", err)
	}
	return selections, nil
}

func (r *GormSelectionRepository) GetByCandidateOwner(userID string) ([]*domain.Selection, error) {
	selections := make([]*domain.Selection, 0)
	if err := r.db.
		Model(&domain.Selection{}).
		Joins("JOIN candidates ON candidates.id = selections.candidate_id").
		Where("candidates.created_by = ?", userID).
		Order("selections.created_at DESC").
		Find(&selections).Error; err != nil {
		return nil, fmt.Errorf("get selections by candidate owner: %w", err)
	}
	return selections, nil
}

func (r *GormSelectionRepository) GetByCandidateOwnerAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	selections := make([]*domain.Selection, 0)
	if err := r.db.
		Model(&domain.Selection{}).
		Joins("JOIN candidates ON candidates.id = selections.candidate_id").
		Where("candidates.created_by = ? AND selections.pairing_id = ?", userID, pairingID).
		Order("selections.created_at DESC").
		Find(&selections).Error; err != nil {
		return nil, fmt.Errorf("get selections by candidate owner and pairing: %w", err)
	}
	return selections, nil
}

func (r *GormSelectionRepository) UpdateStatus(id string, status domain.SelectionStatus) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("update selection status: id is required")
	}
	if !isValidSelectionStatus(status) {
		return ErrInvalidSelectionStatus
	}

	result := r.db.Model(&domain.Selection{}).Where("id = ?", id).Updates(map[string]any{
		"status": status,
	})
	if result.Error != nil {
		if isSelectionPendingUniqueViolation(result.Error) {
			return ErrActiveSelectionExists
		}
		return fmt.Errorf("update selection status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrSelectionNotFound
	}

	return nil
}

func (r *GormSelectionRepository) GetExpiredSelections() ([]*domain.Selection, error) {
	selections := make([]*domain.Selection, 0)
	if err := r.db.Where("status = ? AND expires_at < NOW()", domain.SelectionPending).Find(&selections).Error; err != nil {
		return nil, fmt.Errorf("get expired selections: %w", err)
	}
	return selections, nil
}

func isValidSelectionStatus(status domain.SelectionStatus) bool {
	switch status {
	case domain.SelectionPending, domain.SelectionApproved, domain.SelectionRejected, domain.SelectionExpired:
		return true
	default:
		return false
	}
}

func isSelectionPendingUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && strings.Contains(pgErr.Message, "idx_selections_one_pending_per_candidate")
	}
	return false
}
