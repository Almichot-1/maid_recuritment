package service

import (
	"fmt"
	"strings"
	"time"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

// MedicalDocumentService records medical certificate uploads for a candidate.
type MedicalDocumentService struct {
	medicalRepository domain.MedicalDataRepository
}

func NewMedicalDocumentService(cfg *config.Config, medicalRepository domain.MedicalDataRepository) (*MedicalDocumentService, error) {
	_ = cfg
	if medicalRepository == nil {
		return nil, fmt.Errorf("medical repository is nil")
	}

	return &MedicalDocumentService{medicalRepository: medicalRepository}, nil
}

func (s *MedicalDocumentService) ParseAndStore(
	candidateID string,
	document *domain.Document,
	_ string,
	_ string,
	_ []byte,
) (*domain.MedicalData, error) {
	if strings.TrimSpace(candidateID) == "" {
		return nil, fmt.Errorf("candidate id is required")
	}
	if document == nil {
		return nil, fmt.Errorf("document is required")
	}

	record := &domain.MedicalData{
		CandidateID: candidateID,
		DocumentID:  document.ID,
		ExtractedAt: time.Now().UTC(),
	}
	if err := s.medicalRepository.Upsert(record); err != nil {
		return nil, fmt.Errorf("store medical data: %w", err)
	}

	return record, nil
}
