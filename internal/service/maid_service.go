package service

import (
	"context"

	"maid-recruitment-tracking/internal/domain"
)

type MaidService struct {
	repository domain.MaidRepository
}

func NewMaidService(repository domain.MaidRepository) *MaidService {
	return &MaidService{repository: repository}
}

func (s *MaidService) ListMaidProfiles(ctx context.Context) ([]domain.Maid, error) {
	return s.repository.List(ctx)
}
