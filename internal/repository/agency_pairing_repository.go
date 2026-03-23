package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

var (
	ErrAgencyPairingNotFound  = errors.New("agency pairing not found")
	ErrDuplicateAgencyPairing = errors.New("agency pairing already exists")
)

type GormAgencyPairingRepository struct {
	db *gorm.DB
}

func (r *GormAgencyPairingRepository) DB() *gorm.DB {
	return r.db
}

func NewGormAgencyPairingRepository(cfg *config.Config) (*GormAgencyPairingRepository, error) {
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

	return &GormAgencyPairingRepository{db: db}, nil
}

func (r *GormAgencyPairingRepository) Create(pairing *domain.AgencyPairing) error {
	if pairing == nil {
		return fmt.Errorf("create agency pairing: pairing is nil")
	}
	if strings.TrimSpace(pairing.ID) == "" {
		pairing.ID = uuid.NewString()
	}
	if pairing.Status == "" {
		pairing.Status = domain.AgencyPairingActive
	}

	if err := r.db.Create(pairing).Error; err != nil {
		if isDuplicateActivePairingError(err) {
			return ErrDuplicateAgencyPairing
		}
		return fmt.Errorf("create agency pairing: %w", err)
	}

	return nil
}

func (r *GormAgencyPairingRepository) GetByID(id string) (*domain.AgencyPairing, error) {
	var pairing domain.AgencyPairing
	if err := r.db.Where("id = ?", strings.TrimSpace(id)).First(&pairing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAgencyPairingNotFound
		}
		return nil, fmt.Errorf("get agency pairing by id: %w", err)
	}
	return &pairing, nil
}

func (r *GormAgencyPairingRepository) GetActiveByUsers(ethiopianUserID, foreignUserID string) (*domain.AgencyPairing, error) {
	var pairing domain.AgencyPairing
	err := r.db.
		Where("ethiopian_user_id = ? AND foreign_user_id = ? AND status = ?", strings.TrimSpace(ethiopianUserID), strings.TrimSpace(foreignUserID), domain.AgencyPairingActive).
		First(&pairing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAgencyPairingNotFound
		}
		return nil, fmt.Errorf("get active agency pairing by users: %w", err)
	}
	return &pairing, nil
}

func (r *GormAgencyPairingRepository) List(filters domain.AgencyPairingFilters) ([]*domain.AgencyPairing, error) {
	query := r.db.Model(&domain.AgencyPairing{})
	if strings.TrimSpace(filters.UserID) != "" {
		query = query.Where("ethiopian_user_id = ? OR foreign_user_id = ?", strings.TrimSpace(filters.UserID), strings.TrimSpace(filters.UserID))
	}
	if strings.TrimSpace(filters.EthiopianUserID) != "" {
		query = query.Where("ethiopian_user_id = ?", strings.TrimSpace(filters.EthiopianUserID))
	}
	if strings.TrimSpace(filters.ForeignUserID) != "" {
		query = query.Where("foreign_user_id = ?", strings.TrimSpace(filters.ForeignUserID))
	}
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	pairings := make([]*domain.AgencyPairing, 0)
	if err := query.Order("created_at DESC").Find(&pairings).Error; err != nil {
		return nil, fmt.Errorf("list agency pairings: %w", err)
	}
	return pairings, nil
}

func (r *GormAgencyPairingRepository) Update(pairing *domain.AgencyPairing) error {
	if pairing == nil {
		return fmt.Errorf("update agency pairing: pairing is nil")
	}
	result := r.db.Model(&domain.AgencyPairing{}).
		Where("id = ?", strings.TrimSpace(pairing.ID)).
		Updates(map[string]any{
			"status":               pairing.Status,
			"approved_by_admin_id": pairing.ApprovedByAdminID,
			"approved_at":          pairing.ApprovedAt,
			"ended_at":             pairing.EndedAt,
			"notes":                pairing.Notes,
		})
	if result.Error != nil {
		if isDuplicateActivePairingError(result.Error) {
			return ErrDuplicateAgencyPairing
		}
		return fmt.Errorf("update agency pairing: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrAgencyPairingNotFound
	}
	return nil
}

func isDuplicateActivePairingError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && strings.Contains(pgErr.Message, "idx_agency_pairings_unique_active_pair")
	}
	return false
}
