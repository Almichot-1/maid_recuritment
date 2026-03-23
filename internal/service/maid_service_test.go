package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"maid-recruitment-tracking/internal/domain"
)

type maidRepoMock struct {
	listFn func(ctx context.Context) ([]domain.Maid, error)
}

func (m *maidRepoMock) List(ctx context.Context) ([]domain.Maid, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return []domain.Maid{}, nil
}

func TestMaidService_ListMaidProfiles(t *testing.T) {
	repo := &maidRepoMock{listFn: func(ctx context.Context) ([]domain.Maid, error) {
		return []domain.Maid{{ID: "1", Name: "A"}}, nil
	}}

	service := NewMaidService(repo)
	require.NotNil(t, service)

	items, err := service.ListMaidProfiles(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "A", items[0].Name)
}