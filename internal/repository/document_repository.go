package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

var ErrDocumentNotFound = errors.New("document not found")

type GormDocumentRepository struct {
	db *gorm.DB
}

func NewGormDocumentRepository(cfg *config.Config) (*GormDocumentRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &GormDocumentRepository{db: db}, nil
}

func (r *GormDocumentRepository) Create(document *domain.Document) error {
	if document == nil {
		return fmt.Errorf("create document: document is nil")
	}
	if document.ID == "" {
		document.ID = uuid.NewString()
	}
	if err := r.db.Create(document).Error; err != nil {
		return fmt.Errorf("create document: %w", err)
	}
	return nil
}

func (r *GormDocumentRepository) GetByCandidateID(candidateID string) ([]*domain.Document, error) {
	documents := make([]*domain.Document, 0)
	if err := r.db.Where("candidate_id = ?", candidateID).Order("uploaded_at DESC").Find(&documents).Error; err != nil {
		return nil, fmt.Errorf("get documents by candidate id: %w", err)
	}
	return documents, nil
}

func (r *GormDocumentRepository) Delete(id string) error {
	result := r.db.Delete(&domain.Document{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("delete document: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrDocumentNotFound
	}
	return nil
}
