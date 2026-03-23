package service

import (
	"errors"
	"fmt"
	"testing"

	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/domain"
)

type statusStepCandidateInitMock struct {
	getByIDFn func(id string) (*domain.Candidate, error)
}

func (m *statusStepCandidateInitMock) Create(candidate *domain.Candidate) error { return nil }
func (m *statusStepCandidateInitMock) GetByID(id string) (*domain.Candidate, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, errors.New("not found")
}
func (m *statusStepCandidateInitMock) List(filters domain.CandidateFilters) ([]*domain.Candidate, error) {
	return nil, nil
}
func (m *statusStepCandidateInitMock) Update(candidate *domain.Candidate) error { return nil }
func (m *statusStepCandidateInitMock) Delete(id string) error                   { return nil }
func (m *statusStepCandidateInitMock) Lock(candidateID, lockedBy string, expiresAt time.Time) error {
	return nil
}
func (m *statusStepCandidateInitMock) Unlock(candidateID string) error { return nil }

func TestServiceConstructorsValidation(t *testing.T) {
	_, err := NewSelectionService(nil, &candidateRepositoryMock{}, &notificationSenderMock{foreignByID: map[string]bool{}})
	require.Error(t, err)
	_, err = NewSelectionService(&selectionRepositoryMock{}, nil, &notificationSenderMock{foreignByID: map[string]bool{}})
	require.Error(t, err)

	_, err = NewCandidateService(nil, &documentRepositoryMock{}, &storageServiceMock{}, &PDFService{})
	require.Error(t, err)
	_, err = NewCandidateService(&candidateRepoBehaviorMock{}, nil, &storageServiceMock{}, &PDFService{})
	require.Error(t, err)

	_, err = NewApprovalService(nil, &selectionRepositoryMock{}, &candidateRepositoryMock{}, &StatusStepService{}, &notificationSenderMock{foreignByID: map[string]bool{}})
	require.Error(t, err)
}

func TestStatusStepService_InitializeSteps(t *testing.T) {
	dsn := fmt.Sprintf("file:status_init_%s?mode=memory&cache=shared", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&approvalTestStatusStep{}))

	service, err := NewStatusStepService(
		&statusStepRepositoryMock{db: db},
		&statusStepCandidateInitMock{getByIDFn: func(id string) (*domain.Candidate, error) {
			return &domain.Candidate{ID: id, CreatedBy: "owner-1"}, nil
		}},
		&selectionRepositoryMock{db: db},
		&notificationSenderMock{foreignByID: map[string]bool{}},
	)
	require.NoError(t, err)

	require.NoError(t, service.InitializeSteps("cand-1"))

	var rows []approvalTestStatusStep
	require.NoError(t, db.Where("candidate_id = ?", "cand-1").Find(&rows).Error)
	require.Len(t, rows, len(predefinedStepNames()))

	err = service.InitializeSteps("")
	require.Error(t, err)
}
