package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

var ErrAgencyApprovalNotFound = errors.New("agency approval request not found")

type GormAgencyApprovalRepository struct {
	db *gorm.DB
}

func NewGormAgencyApprovalRepository(cfg *config.Config) (*GormAgencyApprovalRepository, error) {
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

	return &GormAgencyApprovalRepository{db: db}, nil
}

func (r *GormAgencyApprovalRepository) Create(request *domain.AgencyApprovalRequest) error {
	if request == nil {
		return fmt.Errorf("create approval request: request is nil")
	}
	if strings.TrimSpace(request.ID) == "" {
		request.ID = uuid.NewString()
	}
	if strings.TrimSpace(request.AgencyID) == "" {
		return fmt.Errorf("create approval request: agency id is required")
	}
	if request.Status == "" {
		request.Status = domain.AgencyApprovalPending
	}
	if err := r.db.Create(request).Error; err != nil {
		return fmt.Errorf("create approval request: %w", err)
	}
	return nil
}

func (r *GormAgencyApprovalRepository) GetByAgencyID(agencyID string) (*domain.AgencyApprovalRequest, error) {
	var request domain.AgencyApprovalRequest
	if err := r.db.Where("agency_id = ?", strings.TrimSpace(agencyID)).First(&request).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAgencyApprovalNotFound
		}
		return nil, fmt.Errorf("get approval request by agency id: %w", err)
	}
	return &request, nil
}

func (r *GormAgencyApprovalRepository) ListByStatus(status domain.AgencyApprovalStatus) ([]*domain.AgencyApprovalRequest, error) {
	requests := make([]*domain.AgencyApprovalRequest, 0)
	query := r.db.Model(&domain.AgencyApprovalRequest{})
	if strings.TrimSpace(string(status)) != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("created_at ASC").Find(&requests).Error; err != nil {
		return nil, fmt.Errorf("list approval requests: %w", err)
	}
	return requests, nil
}

func (r *GormAgencyApprovalRepository) Update(request *domain.AgencyApprovalRequest) error {
	if request == nil {
		return fmt.Errorf("update approval request: request is nil")
	}
	if strings.TrimSpace(request.ID) == "" {
		return fmt.Errorf("update approval request: id is required")
	}
	result := r.db.Model(&domain.AgencyApprovalRequest{}).
		Where("id = ?", request.ID).
		Updates(map[string]any{
			"status":           request.Status,
			"reviewed_by":      request.ReviewedBy,
			"reviewed_at":      request.ReviewedAt,
			"rejection_reason": request.RejectionReason,
			"admin_notes":      request.AdminNotes,
		})
	if result.Error != nil {
		return fmt.Errorf("update approval request: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrAgencyApprovalNotFound
	}
	return nil
}
