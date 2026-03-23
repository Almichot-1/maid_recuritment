package service

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"maid-recruitment-tracking/internal/domain"
)

type candidateRepoBehaviorMock struct {
	createFn func(candidate *domain.Candidate) error
	getByID  func(id string) (*domain.Candidate, error)
	listFn   func(filters domain.CandidateFilters) ([]*domain.Candidate, error)
	updateFn func(candidate *domain.Candidate) error
	deleteFn func(id string) error
}

func (m *candidateRepoBehaviorMock) Create(candidate *domain.Candidate) error {
	if m.createFn != nil {
		return m.createFn(candidate)
	}
	return nil
}
func (m *candidateRepoBehaviorMock) GetByID(id string) (*domain.Candidate, error) {
	if m.getByID != nil {
		return m.getByID(id)
	}
	return nil, errors.New("not configured")
}
func (m *candidateRepoBehaviorMock) List(filters domain.CandidateFilters) ([]*domain.Candidate, error) {
	if m.listFn != nil {
		return m.listFn(filters)
	}
	return nil, nil
}
func (m *candidateRepoBehaviorMock) Update(candidate *domain.Candidate) error {
	if m.updateFn != nil {
		return m.updateFn(candidate)
	}
	return nil
}
func (m *candidateRepoBehaviorMock) Delete(id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}
func (m *candidateRepoBehaviorMock) Lock(candidateID, lockedBy string, expiresAt time.Time) error {
	return nil
}
func (m *candidateRepoBehaviorMock) Unlock(candidateID string) error { return nil }

type documentRepositoryMock struct{}

func (m *documentRepositoryMock) Create(document *domain.Document) error { return nil }
func (m *documentRepositoryMock) GetByCandidateID(candidateID string) ([]*domain.Document, error) {
	return nil, nil
}
func (m *documentRepositoryMock) Delete(id string) error { return nil }

type storageServiceMock struct{}

func (m *storageServiceMock) Upload(file io.Reader, fileName, contentType string) (string, error) {
	return "", nil
}
func (m *storageServiceMock) Delete(url string) error { return nil }

type userRepositoryBehaviorMock struct{}

func (m *userRepositoryBehaviorMock) Create(user *domain.User) error { return nil }
func (m *userRepositoryBehaviorMock) GetByEmail(email string) (*domain.User, error) {
	return nil, errors.New("not configured")
}
func (m *userRepositoryBehaviorMock) GetByID(id string) (*domain.User, error) {
	return nil, errors.New("not configured")
}
func (m *userRepositoryBehaviorMock) Update(user *domain.User) error { return nil }

type agencyPairingRepositoryBehaviorMock struct{}

func (m *agencyPairingRepositoryBehaviorMock) Create(pairing *domain.AgencyPairing) error { return nil }
func (m *agencyPairingRepositoryBehaviorMock) GetByID(id string) (*domain.AgencyPairing, error) {
	return nil, errors.New("not configured")
}
func (m *agencyPairingRepositoryBehaviorMock) GetActiveByUsers(ethiopianUserID, foreignUserID string) (*domain.AgencyPairing, error) {
	return nil, errors.New("not configured")
}
func (m *agencyPairingRepositoryBehaviorMock) List(filters domain.AgencyPairingFilters) ([]*domain.AgencyPairing, error) {
	return nil, nil
}
func (m *agencyPairingRepositoryBehaviorMock) Update(pairing *domain.AgencyPairing) error { return nil }

type candidatePairShareRepositoryBehaviorMock struct {
	listByCandidateIDFn func(candidateID string, activeOnly bool) ([]*domain.CandidatePairShare, error)
}

func (m *candidatePairShareRepositoryBehaviorMock) Create(share *domain.CandidatePairShare) error {
	return nil
}
func (m *candidatePairShareRepositoryBehaviorMock) GetActiveByPairingAndCandidate(pairingID, candidateID string) (*domain.CandidatePairShare, error) {
	return nil, errors.New("not configured")
}
func (m *candidatePairShareRepositoryBehaviorMock) ListByCandidateID(candidateID string, activeOnly bool) ([]*domain.CandidatePairShare, error) {
	if m.listByCandidateIDFn != nil {
		return m.listByCandidateIDFn(candidateID, activeOnly)
	}
	return nil, nil
}
func (m *candidatePairShareRepositoryBehaviorMock) ListByPairingID(pairingID string, activeOnly bool) ([]*domain.CandidatePairShare, error) {
	return nil, nil
}
func (m *candidatePairShareRepositoryBehaviorMock) Deactivate(pairingID, candidateID string, unsharedAt time.Time) error {
	return nil
}

type selectionRepositoryBehaviorMock struct{}

func (m *selectionRepositoryBehaviorMock) Create(selection *domain.Selection) error { return nil }
func (m *selectionRepositoryBehaviorMock) GetByID(id string) (*domain.Selection, error) {
	return nil, errors.New("not configured")
}
func (m *selectionRepositoryBehaviorMock) GetByCandidateID(candidateID string) (*domain.Selection, error) {
	return nil, errors.New("not configured")
}
func (m *selectionRepositoryBehaviorMock) GetByCandidateIDAndPairingID(candidateID, pairingID string) (*domain.Selection, error) {
	return nil, errors.New("not configured")
}
func (m *selectionRepositoryBehaviorMock) GetBySelectedBy(userID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *selectionRepositoryBehaviorMock) GetBySelectedByAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *selectionRepositoryBehaviorMock) GetByCandidateOwner(userID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *selectionRepositoryBehaviorMock) GetByCandidateOwnerAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *selectionRepositoryBehaviorMock) UpdateStatus(id string, status domain.SelectionStatus) error {
	return nil
}
func (m *selectionRepositoryBehaviorMock) GetExpiredSelections() ([]*domain.Selection, error) {
	return nil, nil
}

type auditLogRepositoryBehaviorMock struct{}

func (m *auditLogRepositoryBehaviorMock) Create(log *domain.AuditLog) error { return nil }
func (m *auditLogRepositoryBehaviorMock) List(filters domain.AuditLogFilters) ([]*domain.AuditLog, error) {
	return nil, nil
}

func newCandidateServiceForTests(t *testing.T, repo *candidateRepoBehaviorMock) *CandidateService {
	t.Helper()

	svc, err := NewCandidateService(repo, &documentRepositoryMock{}, &storageServiceMock{}, &PDFService{})
	require.NoError(t, err)
	return svc
}

func TestCandidateService_CreateCandidate_EthiopianAgentCanCreate(t *testing.T) {
	created := false
	repo := &candidateRepoBehaviorMock{
		createFn: func(candidate *domain.Candidate) error {
			created = true
			assert.Equal(t, "ethiopian-agent-1", candidate.CreatedBy)
			assert.Equal(t, domain.CandidateStatusDraft, candidate.Status)
			return nil
		},
	}
	svc := newCandidateServiceForTests(t, repo)

	age := 25
	exp := 3
	result, err := svc.CreateCandidate("ethiopian-agent-1", CandidateInput{
		FullName:        "Sara",
		Age:             &age,
		ExperienceYears: &exp,
		Languages:       []string{"Amharic", "English"},
		Skills:          []string{"Cooking"},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, created)
}

func TestCandidateService_CreateCandidate_ForeignAgentCannotCreate(t *testing.T) {
	repo := &candidateRepoBehaviorMock{}
	svc := newCandidateServiceForTests(t, repo)

	_, err := svc.CreateCandidate("", CandidateInput{FullName: "Any"})
	require.ErrorIs(t, err, ErrForbidden)
}

func TestCandidateService_UpdateCandidate_CannotUpdateLockedCandidate(t *testing.T) {
	lockOwner := "foreign-agent"
	expires := time.Now().UTC().Add(1 * time.Hour)

	repo := &candidateRepoBehaviorMock{
		getByID: func(id string) (*domain.Candidate, error) {
			return &domain.Candidate{
				ID:            id,
				CreatedBy:     "ethiopian-agent",
				FullName:      "Candidate",
				Status:        domain.CandidateStatusAvailable,
				LockedBy:      &lockOwner,
				LockExpiresAt: &expires,
			}, nil
		},
	}
	svc := newCandidateServiceForTests(t, repo)

	err := svc.UpdateCandidate("cand-1", "ethiopian-agent", CandidateInput{FullName: "Updated"})
	require.ErrorIs(t, err, ErrCandidateLocked)
}

func TestCandidateService_ListCandidates_VisibilityRules(t *testing.T) {
	tests := []struct {
		name              string
		role              string
		inputFilters      domain.CandidateFilters
		expectErr         error
		expectListCalled  bool
		expectedCreatedBy string
		expectedStatuses  []domain.CandidateStatus
	}{
		{
			name:              "ethiopian sees own candidates",
			role:              string(domain.EthiopianAgent),
			inputFilters:      domain.CandidateFilters{CreatedBy: "ethiopian-1"},
			expectListCalled:  true,
			expectedCreatedBy: "ethiopian-1",
		},
		{
			name:              "foreign sees only available candidates",
			role:              string(domain.ForeignAgent),
			inputFilters:      domain.CandidateFilters{CreatedBy: "should-be-ignored"},
			expectListCalled:  true,
			expectedCreatedBy: "",
			expectedStatuses:  []domain.CandidateStatus{domain.CandidateStatusAvailable},
		},
		{
			name:         "ethiopian without createdBy is forbidden",
			role:         string(domain.EthiopianAgent),
			inputFilters: domain.CandidateFilters{},
			expectErr:    ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listCalled := false
			var gotFilters domain.CandidateFilters

			repo := &candidateRepoBehaviorMock{
				listFn: func(filters domain.CandidateFilters) ([]*domain.Candidate, error) {
					listCalled = true
					gotFilters = filters
					return []*domain.Candidate{}, nil
				},
			}
			svc := newCandidateServiceForTests(t, repo)

			_, err := svc.ListCandidates(tt.role, tt.inputFilters)

			if tt.expectErr != nil {
				require.ErrorIs(t, err, tt.expectErr)
				assert.False(t, listCalled)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectListCalled, listCalled)
			assert.Equal(t, tt.expectedCreatedBy, gotFilters.CreatedBy)
			if tt.expectedStatuses != nil {
				assert.Equal(t, tt.expectedStatuses, gotFilters.Statuses)
			}
		})
	}
}

func TestCandidateService_DeleteCandidate_AvailableCandidateCanBeDeletedGlobally(t *testing.T) {
	deleted := false
	repo := &candidateRepoBehaviorMock{
		getByID: func(id string) (*domain.Candidate, error) {
			return &domain.Candidate{
				ID:        id,
				CreatedBy: "ethiopian-agent",
				Status:    domain.CandidateStatusAvailable,
			}, nil
		},
		deleteFn: func(id string) error {
			deleted = true
			return nil
		},
	}
	svc := newCandidateServiceForTests(t, repo)

	err := svc.DeleteCandidate("candidate-1", "ethiopian-agent")
	require.NoError(t, err)
	assert.True(t, deleted)
}
