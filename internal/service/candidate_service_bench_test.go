package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"maid-recruitment-tracking/internal/domain"
)

type candidateRepoBenchMock struct {
	getByIDFn func(id string) (*domain.Candidate, error)
	listFn    func(filters domain.CandidateFilters) ([]*domain.Candidate, error)
}

func (m *candidateRepoBenchMock) Create(candidate *domain.Candidate) error { return nil }
func (m *candidateRepoBenchMock) GetByID(id string) (*domain.Candidate, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, errors.New("not configured")
}
func (m *candidateRepoBenchMock) GetByIDs(ids []string) ([]*domain.Candidate, error) {
	return nil, nil
}
func (m *candidateRepoBenchMock) GetByIDLean(id string) (*domain.Candidate, error) {
	return m.GetByID(id)
}
func (m *candidateRepoBenchMock) List(filters domain.CandidateFilters) ([]*domain.Candidate, error) {
	if m.listFn != nil {
		return m.listFn(filters)
	}
	return nil, nil
}
func (m *candidateRepoBenchMock) Update(candidate *domain.Candidate) error { return nil }
func (m *candidateRepoBenchMock) Delete(id string) error                   { return nil }
func (m *candidateRepoBenchMock) Lock(candidateID, lockedBy string, expiresAt time.Time) error {
	return nil
}
func (m *candidateRepoBenchMock) Unlock(candidateID string) error { return nil }
func (m *candidateRepoBenchMock) UpdateStatus(id string, status domain.CandidateStatus) error { return nil }

func newCandidateServiceForBenchmark(b *testing.B, repo *candidateRepoBenchMock) *CandidateService {
	b.Helper()
	svc, err := NewCandidateService(repo, &documentRepositoryMock{}, &storageServiceMock{}, &PDFService{}, &userRepositoryBehaviorMock{}, &candidatePairShareRepositoryBehaviorMock{}, &pairOverrideRepositoryBehaviorMock{}, nil, nil, nil, nil, nil)
	require.NoError(b, err)
	return svc
}

func BenchmarkCandidateService_ListCandidates(b *testing.B) {
	repo := &candidateRepoBenchMock{
		listFn: func(filters domain.CandidateFilters) ([]*domain.Candidate, error) {
			return []*domain.Candidate{}, nil
		},
	}
	svc := newCandidateServiceForBenchmark(b, repo)

	filters := domain.CandidateFilters{
		CreatedBy: "ethiopian-agent-1",
		Page:      1,
		PageSize:  20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.ListCandidates(string(domain.EthiopianAgent), filters)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCandidateService_GetCandidate(b *testing.B) {
	candidateID := "bench-candidate-id"
	repo := &candidateRepoBenchMock{
		getByIDFn: func(id string) (*domain.Candidate, error) {
			return &domain.Candidate{
				ID:        id,
				CreatedBy: "ethiopian-agent",
				FullName:  "Bench Candidate",
				Status:    domain.CandidateStatusAvailable,
			}, nil
		},
	}
	svc := newCandidateServiceForBenchmark(b, repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := svc.GetCandidate(candidateID)
		if err != nil {
			b.Fatal(err)
		}
	}
}
