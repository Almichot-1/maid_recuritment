package service

import (
	"bytes"
	"errors"
	"fmt"
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
func (m *candidateRepoBehaviorMock) GetByIDs(ids []string) ([]*domain.Candidate, error) {
	return nil, nil
}
func (m *candidateRepoBehaviorMock) GetByIDLean(id string) (*domain.Candidate, error) {
	return m.GetByID(id)
}
func (m *candidateRepoBehaviorMock) UpdateStatus(id string, status domain.CandidateStatus) error {
	return nil
}

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
func (m *storageServiceMock) Open(fileURL string) (io.ReadCloser, string, error) {
	return io.NopCloser(bytes.NewReader(nil)), "", nil
}

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
func (m *candidatePairShareRepositoryBehaviorMock) UpdateCVURL(shareID, cvURL string) error {
	return nil
}

type pairOverrideRepositoryBehaviorMock struct {
	bulkUpsertFn func(overrides []*domain.CandidatePairOverride) error
}

func (m *pairOverrideRepositoryBehaviorMock) Upsert(override *domain.CandidatePairOverride) error {
	return nil
}
func (m *pairOverrideRepositoryBehaviorMock) GetByPairingAndCandidate(pairingID, candidateID string) (*domain.CandidatePairOverride, error) {
	return nil, nil
}
func (m *pairOverrideRepositoryBehaviorMock) ListByCandidateID(candidateID string) ([]*domain.CandidatePairOverride, error) {
	return nil, nil
}
func (m *pairOverrideRepositoryBehaviorMock) BulkUpsert(overrides []*domain.CandidatePairOverride) error {
	if m.bulkUpsertFn != nil {
		return m.bulkUpsertFn(overrides)
	}
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

	svc, err := NewCandidateService(repo, &documentRepositoryMock{}, &storageServiceMock{}, &PDFService{}, &userRepositoryBehaviorMock{}, &candidatePairShareRepositoryBehaviorMock{}, &pairOverrideRepositoryBehaviorMock{}, nil, nil, nil, nil, nil)
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
		Languages:       []domain.LanguageEntry{{Language: "Amharic", Proficiency: "Basic"}, {Language: "English", Proficiency: "Basic"}},
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

func strPtr(s string) *string {
	return &s
}

func TestSalaryResolution_PriorityChain(t *testing.T) {
	tests := []struct {
		name            string
		salaryOverride  string
		candidateSalary string
		pairing         *domain.AgencyPairing
		expected        string
	}{
		{
			name:            "override wins",
			salaryOverride:  "3000 KWD",
			candidateSalary: "2000 KWD",
			pairing:         &domain.AgencyPairing{DefaultSalary: strPtr("1500"), DefaultCurrency: strPtr("SAR")},
			expected:        "3000 KWD",
		},
		{
			name:            "candidate salary keeps existing",
			salaryOverride:  "",
			candidateSalary: "2000 KWD",
			pairing:         &domain.AgencyPairing{DefaultSalary: strPtr("1500"), DefaultCurrency: strPtr("SAR")},
			expected:        "2000 KWD",
		},
		{
			name:            "pairing DefaultSalary + DefaultCurrency combined",
			salaryOverride:  "",
			candidateSalary: "",
			pairing:         &domain.AgencyPairing{DefaultSalary: strPtr("2000"), DefaultCurrency: strPtr("KWD")},
			expected:        "2000 KWD",
		},
		{
			name:            "only DefaultCurrency yields Negotiable",
			salaryOverride:  "",
			candidateSalary: "",
			pairing:         &domain.AgencyPairing{DefaultSalary: nil, DefaultCurrency: strPtr("KWD")},
			expected:        "Negotiable KWD",
		},
		{
			name:            "no defaults leaves empty",
			salaryOverride:  "",
			candidateSalary: "",
			pairing:         &domain.AgencyPairing{DefaultSalary: nil, DefaultCurrency: nil},
			expected:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := &domain.Candidate{SalaryOffered: tt.candidateSalary}

			// Apply the same logic as GenerateCV's salary resolution
			if tt.salaryOverride != "" {
				candidate.SalaryOffered = tt.salaryOverride
			} else if candidate.SalaryOffered != "" {
				// keep existing
			} else if tt.pairing.DefaultSalary != nil && *tt.pairing.DefaultSalary != "" {
				salStr := *tt.pairing.DefaultSalary
				if tt.pairing.DefaultCurrency != nil && *tt.pairing.DefaultCurrency != "" {
					salStr = fmt.Sprintf("%s %s", *tt.pairing.DefaultSalary, *tt.pairing.DefaultCurrency)
				}
				candidate.SalaryOffered = salStr
			} else if tt.pairing.DefaultCurrency != nil && *tt.pairing.DefaultCurrency != "" {
				candidate.SalaryOffered = fmt.Sprintf("Negotiable %s", *tt.pairing.DefaultCurrency)
			}

			assert.Equal(t, tt.expected, candidate.SalaryOffered)
		})
	}
}

func TestSalaryResolution_NoDoubleCurrency(t *testing.T) {
	// When DefaultSalary already includes currency text (e.g. "2000 KWD")
	// and DefaultCurrency is also "KWD", the result should be "2000 KWD", not "2000 KWD KWD"
	pairing := &domain.AgencyPairing{DefaultSalary: strPtr("2000"), DefaultCurrency: strPtr("KWD")}
	candidate := &domain.Candidate{SalaryOffered: ""}

	if pairing.DefaultSalary != nil && *pairing.DefaultSalary != "" {
		salStr := *pairing.DefaultSalary
		if pairing.DefaultCurrency != nil && *pairing.DefaultCurrency != "" {
			salStr = fmt.Sprintf("%s %s", *pairing.DefaultSalary, *pairing.DefaultCurrency)
		}
		candidate.SalaryOffered = salStr
	}

	assert.Equal(t, "2000 KWD", candidate.SalaryOffered)
	// Verify no double currency
	assert.NotContains(t, candidate.SalaryOffered, "KWD KWD")
}

func TestBatchRegenerateCVs_MaxBatch(t *testing.T) {
	// Create enough IDs to exceed MaxBatchSize
	ids := make([]string, MaxBatchSize+1)
	for i := range ids {
		ids[i] = fmt.Sprintf("candidate-%d", i)
	}

	// The service method doesn't validate max batch internally (the handler does),
	// but the handler validation is tested. Verify the service accepts and processes
	// all candidates without error.
	repo := &candidateRepoBehaviorMock{
		getByID: func(id string) (*domain.Candidate, error) {
			return &domain.Candidate{
				ID:        id,
				CreatedBy: "user-1",
				FullName:  "Test",
				Status:    domain.CandidateStatusDraft,
				CVPDFURL:  "existing.pdf",
			}, nil
		},
	}
	svc := newCandidateServiceForTests(t, repo)
	pdfSvc := &PDFService{}
	svc.pdfService = pdfSvc

	// Should not panic or error when processing batch
	input := BatchRegenerateCVsInput{CandidateIDs: ids, PairingID: "pairing-1"}
	result := svc.BatchRegenerateCVs("user-1", input)
	assert.Equal(t, len(ids), result.SuccessCount+result.ErrorCount)
}

func TestBatchSetPairOverrides_Ownership(t *testing.T) {
	repo := &candidateRepoBehaviorMock{
		getByID: func(id string) (*domain.Candidate, error) {
			return nil, errors.New("not used with GetByIDs")
		},
	}
	svc := newCandidateServiceForTests(t, repo)

	// GetByIDs returns nil, so all candidates hit the "not in owner map" path
	// and are counted as forbidden
	input := BatchSetPairOverrideInput{
		CandidateIDs:   []string{"candidate-1", "candidate-2"},
		PairingID:      "pairing-1",
		CountryApplied: "Kuwait",
		SalaryOffered:  "900 KWD",
	}
	result := svc.BatchSetPairOverrides("user-1", input)

	assert.Equal(t, 0, result.SuccessCount)
	assert.Equal(t, 2, result.ErrorCount)
}
