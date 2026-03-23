package repository

import (
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
	ErrApprovalNotFound        = errors.New("approval not found")
	ErrInvalidApprovalDecision = errors.New("invalid approval decision")
)

type GormApprovalRepository struct {
	db *gorm.DB
}

func (r *GormApprovalRepository) DB() *gorm.DB {
	return r.db
}

func NewGormApprovalRepository(cfg *config.Config) (*GormApprovalRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &GormApprovalRepository{db: db}, nil
}

func (r *GormApprovalRepository) Create(approval *domain.Approval) error {
	if approval == nil {
		return fmt.Errorf("create approval: approval is nil")
	}
	if strings.TrimSpace(approval.SelectionID) == "" {
		return fmt.Errorf("create approval: selection id is required")
	}
	if strings.TrimSpace(approval.UserID) == "" {
		return fmt.Errorf("create approval: user id is required")
	}
	if !isValidApprovalDecision(approval.Decision) {
		return ErrInvalidApprovalDecision
	}

	if approval.ID == "" {
		approval.ID = uuid.NewString()
	}
	if approval.DecidedAt.IsZero() {
		approval.DecidedAt = time.Now().UTC()
	}

	if err := r.db.Create(approval).Error; err != nil {
		return fmt.Errorf("create approval: %w", err)
	}
	return nil
}

func (r *GormApprovalRepository) GetBySelectionID(selectionID string) ([]*domain.Approval, error) {
	approvals := make([]*domain.Approval, 0)
	if err := r.db.Preload("User").Where("selection_id = ?", selectionID).Order("decided_at DESC").Find(&approvals).Error; err != nil {
		return nil, fmt.Errorf("get approvals by selection id: %w", err)
	}
	return approvals, nil
}

func (r *GormApprovalRepository) GetBySelectionAndUser(selectionID, userID string) (*domain.Approval, error) {
	var approval domain.Approval
	if err := r.db.Preload("User").Where("selection_id = ? AND user_id = ?", selectionID, userID).First(&approval).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApprovalNotFound
		}
		return nil, fmt.Errorf("get approval by selection and user: %w", err)
	}
	return &approval, nil
}

func isValidApprovalDecision(decision domain.ApprovalDecision) bool {
	switch decision {
	case domain.ApprovalApproved, domain.ApprovalRejected:
		return true
	default:
		return false
	}
}
