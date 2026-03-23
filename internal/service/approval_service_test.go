package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/domain"
)

type approvalRepositoryMock struct {
	db *gorm.DB
}

type approvalTestCandidate struct {
	ID            string         `gorm:"primaryKey;column:id"`
	CreatedBy     string         `gorm:"column:created_by"`
	FullName      string         `gorm:"column:full_name"`
	Status        string         `gorm:"column:status"`
	LockedBy      *string        `gorm:"column:locked_by"`
	LockedAt      *time.Time     `gorm:"column:locked_at"`
	LockExpiresAt *time.Time     `gorm:"column:lock_expires_at"`
	Languages     []byte         `gorm:"column:languages"`
	Skills        []byte         `gorm:"column:skills"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (approvalTestCandidate) TableName() string { return "candidates" }

type approvalTestSelection struct {
	ID                         string     `gorm:"primaryKey;column:id"`
	CandidateID                string     `gorm:"column:candidate_id"`
	SelectedBy                 string     `gorm:"column:selected_by"`
	Status                     string     `gorm:"column:status"`
	EmployerContractURL        string     `gorm:"column:employer_contract_url"`
	EmployerContractFileName   string     `gorm:"column:employer_contract_file_name"`
	EmployerContractUploadedAt *time.Time `gorm:"column:employer_contract_uploaded_at"`
	EmployerIDURL              string     `gorm:"column:employer_id_url"`
	EmployerIDFileName         string     `gorm:"column:employer_id_file_name"`
	EmployerIDUploadedAt       *time.Time `gorm:"column:employer_id_uploaded_at"`
	ExpiresAt                  time.Time  `gorm:"column:expires_at"`
	CreatedAt                  time.Time  `gorm:"column:created_at"`
	UpdatedAt                  time.Time  `gorm:"column:updated_at"`
}

func (approvalTestSelection) TableName() string { return "selections" }

type approvalTestApproval struct {
	ID          string    `gorm:"primaryKey;column:id"`
	SelectionID string    `gorm:"column:selection_id"`
	UserID      string    `gorm:"column:user_id"`
	Decision    string    `gorm:"column:decision"`
	DecidedAt   time.Time `gorm:"column:decided_at"`
}

func (approvalTestApproval) TableName() string { return "approvals" }

type approvalTestStatusStep struct {
	ID          string     `gorm:"primaryKey;column:id"`
	CandidateID string     `gorm:"column:candidate_id"`
	StepName    string     `gorm:"column:step_name"`
	StepStatus  string     `gorm:"column:step_status"`
	CompletedAt *time.Time `gorm:"column:completed_at"`
	Notes       string     `gorm:"column:notes"`
	UpdatedBy   string     `gorm:"column:updated_by"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
}

func (approvalTestStatusStep) TableName() string { return "status_steps" }

func (m *approvalRepositoryMock) DB() *gorm.DB                           { return m.db }
func (m *approvalRepositoryMock) Create(approval *domain.Approval) error { return nil }
func (m *approvalRepositoryMock) GetBySelectionID(selectionID string) ([]*domain.Approval, error) {
	return nil, nil
}
func (m *approvalRepositoryMock) GetBySelectionAndUser(selectionID, userID string) (*domain.Approval, error) {
	return nil, nil
}

type statusStepRepositoryMock struct {
	db *gorm.DB
}

func (m *statusStepRepositoryMock) DB() *gorm.DB                         { return m.db }
func (m *statusStepRepositoryMock) Create(step *domain.StatusStep) error { return nil }
func (m *statusStepRepositoryMock) GetByCandidateID(candidateID string) ([]*domain.StatusStep, error) {
	return nil, nil
}
func (m *statusStepRepositoryMock) Update(step *domain.StatusStep) error { return nil }

func setupApprovalService(t *testing.T) (*ApprovalService, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:approval_test_%s?mode=memory&cache=shared", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&approvalTestCandidate{}, &approvalTestSelection{}, &approvalTestApproval{}, &approvalTestStatusStep{})
	require.NoError(t, err)

	notifier := &notificationSenderMock{foreignByID: map[string]bool{}}
	statusStepService, err := NewStatusStepService(
		&statusStepRepositoryMock{db: db},
		&candidateRepositoryMock{},
		&selectionRepositoryMock{db: db},
		notifier,
	)
	require.NoError(t, err)

	service, err := NewApprovalService(
		&approvalRepositoryMock{db: db},
		&selectionRepositoryMock{db: db},
		&candidateRepositoryMock{},
		statusStepService,
		notifier,
	)
	require.NoError(t, err)

	return service, db
}

func seedApprovalScenario(t *testing.T, db *gorm.DB, selectionStatus domain.SelectionStatus, expiresAt time.Time) {
	t.Helper()
	now := time.Now().UTC()

	err := db.Create(&approvalTestCandidate{
		ID:        "cand-1",
		CreatedBy: "owner-1",
		FullName:  "Candidate",
		Status:    string(domain.CandidateStatusLocked),
		LockedBy:  ptrString("selector-1"),
		Languages: []byte("[]"),
		Skills:    []byte("[]"),
	}).Error
	require.NoError(t, err)

	err = db.Create(&approvalTestSelection{
		ID:                         "sel-1",
		CandidateID:                "cand-1",
		SelectedBy:                 "selector-1",
		Status:                     string(selectionStatus),
		EmployerContractURL:        "https://files.example/contract.pdf",
		EmployerContractFileName:   "contract.pdf",
		EmployerContractUploadedAt: &now,
		EmployerIDURL:              "https://files.example/employer-id.pdf",
		EmployerIDFileName:         "employer-id.pdf",
		EmployerIDUploadedAt:       &now,
		ExpiresAt:                  expiresAt,
	}).Error
	require.NoError(t, err)
}

func ptrString(value string) *string { return &value }

func TestApprovalService_SingleRejectionBlocksBothParties(t *testing.T) {
	service, db := setupApprovalService(t)
	seedApprovalScenario(t, db, domain.SelectionPending, time.Now().UTC().Add(2*time.Hour))

	err := service.RejectSelection("sel-1", "owner-1", "not suitable")
	require.NoError(t, err)

	var selection domain.Selection
	err = db.Where("id = ?", "sel-1").First(&selection).Error
	require.NoError(t, err)
	assert.Equal(t, domain.SelectionRejected, selection.Status)

	var candidate domain.Candidate
	err = db.Where("id = ?", "cand-1").First(&candidate).Error
	require.NoError(t, err)
	assert.Equal(t, domain.CandidateStatusAvailable, candidate.Status)
	assert.Nil(t, candidate.LockedBy)

	err = service.ApproveSelection("sel-1", "selector-1")
	require.ErrorIs(t, err, ErrSelectionNotPending)
}

func TestApprovalService_BothApprovalsNeededToProceed(t *testing.T) {
	service, db := setupApprovalService(t)
	seedApprovalScenario(t, db, domain.SelectionPending, time.Now().UTC().Add(2*time.Hour))

	err := service.ApproveSelection("sel-1", "owner-1")
	require.NoError(t, err)

	var selectionAfterFirst domain.Selection
	err = db.Where("id = ?", "sel-1").First(&selectionAfterFirst).Error
	require.NoError(t, err)
	assert.Equal(t, domain.SelectionPending, selectionAfterFirst.Status)

	err = service.ApproveSelection("sel-1", "selector-1")
	require.NoError(t, err)

	var selectionAfterSecond domain.Selection
	err = db.Where("id = ?", "sel-1").First(&selectionAfterSecond).Error
	require.NoError(t, err)
	assert.Equal(t, domain.SelectionApproved, selectionAfterSecond.Status)

	var candidate domain.Candidate
	err = db.Where("id = ?", "cand-1").First(&candidate).Error
	require.NoError(t, err)
	assert.Equal(t, domain.CandidateStatusInProgress, candidate.Status)
	assert.Nil(t, candidate.LockedBy)

	var steps []domain.StatusStep
	err = db.Where("candidate_id = ?", "cand-1").Find(&steps).Error
	require.NoError(t, err)
	assert.Len(t, steps, len(predefinedStepNames()))
}

func TestApprovalService_CannotApproveTwice(t *testing.T) {
	service, db := setupApprovalService(t)
	seedApprovalScenario(t, db, domain.SelectionPending, time.Now().UTC().Add(2*time.Hour))

	err := service.ApproveSelection("sel-1", "owner-1")
	require.NoError(t, err)

	err = service.ApproveSelection("sel-1", "owner-1")
	require.NoError(t, err)

	var approvals []domain.Approval
	err = db.Where("selection_id = ? AND user_id = ?", "sel-1", "owner-1").Find(&approvals).Error
	require.NoError(t, err)
	assert.Len(t, approvals, 1)
}

func TestApprovalService_CannotApproveExpiredSelection(t *testing.T) {
	service, db := setupApprovalService(t)
	seedApprovalScenario(t, db, domain.SelectionPending, time.Now().UTC().Add(-1*time.Hour))

	err := service.ApproveSelection("sel-1", "owner-1")
	require.ErrorIs(t, err, ErrSelectionNotPending)
}

func TestApprovalService_ApprovalInitializesStatusSteps(t *testing.T) {
	service, db := setupApprovalService(t)
	seedApprovalScenario(t, db, domain.SelectionPending, time.Now().UTC().Add(2*time.Hour))

	require.NoError(t, service.ApproveSelection("sel-1", "owner-1"))
	require.NoError(t, service.ApproveSelection("sel-1", "selector-1"))

	var steps []domain.StatusStep
	err := db.Where("candidate_id = ?", "cand-1").Order("created_at asc").Find(&steps).Error
	require.NoError(t, err)
	require.Len(t, steps, len(predefinedStepNames()))

	for _, step := range steps {
		assert.Equal(t, domain.Pending, step.StepStatus)
		assert.Equal(t, "owner-1", step.UpdatedBy)
	}
}
