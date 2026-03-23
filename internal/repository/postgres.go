package repository

import (
	"context"

	"maid-recruitment-tracking/internal/domain"
)

type PostgresMaidRepository struct{}

func NewPostgresMaidRepository() *PostgresMaidRepository {
	return &PostgresMaidRepository{}
}

func (r *PostgresMaidRepository) List(ctx context.Context) ([]domain.Maid, error) {
	return []domain.Maid{}, nil
}
