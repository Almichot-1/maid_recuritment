package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

type GormAuditLogRepository struct {
	db *gorm.DB
}

func NewGormAuditLogRepository(cfg *config.Config) (*GormAuditLogRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &GormAuditLogRepository{db: db}, nil
}

func (r *GormAuditLogRepository) Create(log *domain.AuditLog) error {
	if log == nil {
		return fmt.Errorf("create audit log: log is nil")
	}
	if strings.TrimSpace(log.ID) == "" {
		log.ID = uuid.NewString()
	}
	if strings.TrimSpace(log.AdminID) == "" {
		return fmt.Errorf("create audit log: admin id is required")
	}
	if strings.TrimSpace(log.Action) == "" {
		return fmt.Errorf("create audit log: action is required")
	}
	if len(log.Details) == 0 {
		log.Details = []byte("{}")
	}
	if err := r.db.Create(log).Error; err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

func (r *GormAuditLogRepository) List(filters domain.AuditLogFilters) ([]*domain.AuditLog, error) {
	query := r.db.Model(&domain.AuditLog{})

	if strings.TrimSpace(filters.AdminID) != "" {
		query = query.Where("admin_id = ?", strings.TrimSpace(filters.AdminID))
	}
	if strings.TrimSpace(filters.Action) != "" {
		query = query.Where("action = ?", strings.TrimSpace(filters.Action))
	}
	if strings.TrimSpace(filters.TargetType) != "" {
		query = query.Where("target_type = ?", strings.TrimSpace(filters.TargetType))
	}

	page := filters.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}

	offset := (page - 1) * pageSize
	logs := make([]*domain.AuditLog, 0)
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return logs, nil
		}
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	return logs, nil
}
