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
	"gorm.io/gorm/clause"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

type statusStepTestCandidate struct {
	ID        string         `gorm:"primaryKey;column:id"`
	CreatedBy string         `gorm:"column:created_by"`
	FullName  string         `gorm:"column:full_name"`
	Status    string         `gorm:"column:status"`
	Languages []byte         `gorm:"column:languages"`
	Skills    []byte         `gorm:"column:skills"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (statusStepTestCandidate) TableName() string { return "candidates" }

type statusStepTestSelection struct {
	ID               string    `gorm:"primaryKey;column:id"`
	CandidateID      string    `gorm:"column:candidate_id"`
	SelectedBy       string    `gorm:"column:selected_by"`
	Status           string    `gorm:"column:status"`
	WarningSentFlags int       `gorm:"column:warning_sent_flags"`
	ExpiresAt        time.Time `gorm:"column:expires_at"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (statusStepTestSelection) TableName() string { return "selections" }

type statusStepTestStep struct {
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

func (statusStepTestStep) TableName() string { return "status_steps" }

type sqliteStatusStepRepo struct {
	db *gorm.DB
}

func (m *sqliteStatusStepRepo) DB() *gorm.DB { return m.db }

func (m *sqliteStatusStepRepo) Create(step *domain.StatusStep) error {
	if step == nil {
		return nil
	}
	record := statusStepTestStep{
		ID:          step.ID,
		CandidateID: step.CandidateID,
		StepName:    step.StepName,
		StepStatus:  string(step.StepStatus),
		CompletedAt: step.CompletedAt,
		Notes:       step.Notes,
		UpdatedBy:   step.UpdatedBy,
		CreatedAt:   step.CreatedAt,
		UpdatedAt:   step.UpdatedAt,
	}
	return m.db.Create(&record).Error
}

func (m *sqliteStatusStepRepo) GetByCandidateID(candidateID string) ([]*domain.StatusStep, error) {
	records := make([]statusStepTestStep, 0)
	if err := m.db.Where("candidate_id = ?", candidateID).Order("created_at asc").Find(&records).Error; err != nil {
		return nil, err
	}

	steps := make([]*domain.StatusStep, 0, len(records))
	for _, record := range records {
		record := record
		steps = append(steps, &domain.StatusStep{
			ID:          record.ID,
			CandidateID: record.CandidateID,
			StepName:    record.StepName,
			StepStatus:  domain.StepStatus(record.StepStatus),
			CompletedAt: record.CompletedAt,
			Notes:       record.Notes,
			UpdatedBy:   record.UpdatedBy,
			CreatedAt:   record.CreatedAt,
			UpdatedAt:   record.UpdatedAt,
		})
	}

	return steps, nil
}

func (m *sqliteStatusStepRepo) Update(step *domain.StatusStep) error {
	return m.db.Model(&statusStepTestStep{}).Where("id = ?", step.ID).Updates(map[string]any{
		"step_status":  string(step.StepStatus),
		"completed_at": step.CompletedAt,
		"notes":        step.Notes,
		"updated_by":   step.UpdatedBy,
		"updated_at":   time.Now().UTC(),
	}).Error
}

type sqliteStatusStepCandidateRepo struct {
	db *gorm.DB
}

func (m *sqliteStatusStepCandidateRepo) Create(candidate *domain.Candidate) error { return nil }
func (m *sqliteStatusStepCandidateRepo) GetByID(id string) (*domain.Candidate, error) {
	var record statusStepTestCandidate
	if err := m.db.Where("id = ?", id).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrCandidateNotFound
		}
		return nil, err
	}

	return &domain.Candidate{
		ID:        record.ID,
		CreatedBy: record.CreatedBy,
		FullName:  record.FullName,
		Status:    domain.CandidateStatus(record.Status),
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}, nil
}
func (m *sqliteStatusStepCandidateRepo) List(filters domain.CandidateFilters) ([]*domain.Candidate, error) {
	return nil, nil
}
func (m *sqliteStatusStepCandidateRepo) Update(candidate *domain.Candidate) error { return nil }
func (m *sqliteStatusStepCandidateRepo) Delete(id string) error                   { return nil }
func (m *sqliteStatusStepCandidateRepo) Lock(candidateID, lockedBy string, expiresAt time.Time) error {
	return nil
}
func (m *sqliteStatusStepCandidateRepo) Unlock(candidateID string) error { return nil }

type sqliteStatusStepSelectionRepo struct {
	db *gorm.DB
}

func (m *sqliteStatusStepSelectionRepo) Create(selection *domain.Selection) error { return nil }
func (m *sqliteStatusStepSelectionRepo) GetByID(id string) (*domain.Selection, error) {
	return nil, repository.ErrSelectionNotFound
}
func (m *sqliteStatusStepSelectionRepo) GetByCandidateID(candidateID string) (*domain.Selection, error) {
	var record statusStepTestSelection
	if err := m.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("candidate_id = ?", candidateID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrSelectionNotFound
		}
		return nil, err
	}
	return &domain.Selection{
		ID:          record.ID,
		CandidateID: record.CandidateID,
		SelectedBy:  record.SelectedBy,
		Status:      domain.SelectionStatus(record.Status),
		ExpiresAt:   record.ExpiresAt,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}, nil
}
func (m *sqliteStatusStepSelectionRepo) GetByCandidateIDAndPairingID(candidateID, pairingID string) (*domain.Selection, error) {
	return m.GetByCandidateID(candidateID)
}
func (m *sqliteStatusStepSelectionRepo) GetBySelectedBy(userID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *sqliteStatusStepSelectionRepo) GetBySelectedByAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *sqliteStatusStepSelectionRepo) GetByCandidateOwner(userID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *sqliteStatusStepSelectionRepo) GetByCandidateOwnerAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *sqliteStatusStepSelectionRepo) UpdateStatus(id string, status domain.SelectionStatus) error {
	return nil
}
func (m *sqliteStatusStepSelectionRepo) GetExpiredSelections() ([]*domain.Selection, error) {
	return nil, nil
}

func setupStatusStepServiceWithSQLite(t *testing.T) (*StatusStepService, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:status_step_test_%s?mode=memory&cache=shared", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&statusStepTestCandidate{}, &statusStepTestSelection{}, &statusStepTestStep{})
	require.NoError(t, err)

	service, err := NewStatusStepService(
		&sqliteStatusStepRepo{db: db},
		&sqliteStatusStepCandidateRepo{db: db},
		&sqliteStatusStepSelectionRepo{db: db},
		&notificationSenderMock{foreignByID: map[string]bool{}},
	)
	require.NoError(t, err)

	return service, db
}

type statusStepCandidateRepoMock struct {
	getByIDFn func(id string) (*domain.Candidate, error)
}

func (m *statusStepCandidateRepoMock) Create(candidate *domain.Candidate) error { return nil }
func (m *statusStepCandidateRepoMock) GetByID(id string) (*domain.Candidate, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, repository.ErrCandidateNotFound
}
func (m *statusStepCandidateRepoMock) List(filters domain.CandidateFilters) ([]*domain.Candidate, error) {
	return nil, nil
}
func (m *statusStepCandidateRepoMock) Update(candidate *domain.Candidate) error { return nil }
func (m *statusStepCandidateRepoMock) Delete(id string) error                   { return nil }
func (m *statusStepCandidateRepoMock) Lock(candidateID, lockedBy string, expiresAt time.Time) error {
	return nil
}
func (m *statusStepCandidateRepoMock) Unlock(candidateID string) error { return nil }

type statusStepRepoBehaviorMock struct {
	getByCandidateIDFn func(candidateID string) ([]*domain.StatusStep, error)
	updateFn           func(step *domain.StatusStep) error
}

func (m *statusStepRepoBehaviorMock) Create(step *domain.StatusStep) error { return nil }
func (m *statusStepRepoBehaviorMock) GetByCandidateID(candidateID string) ([]*domain.StatusStep, error) {
	if m.getByCandidateIDFn != nil {
		return m.getByCandidateIDFn(candidateID)
	}
	return nil, nil
}
func (m *statusStepRepoBehaviorMock) Update(step *domain.StatusStep) error {
	if m.updateFn != nil {
		return m.updateFn(step)
	}
	return nil
}

type statusStepSelectionRepoMock struct {
	getByCandidateIDFn func(candidateID string) (*domain.Selection, error)
}

func (m *statusStepSelectionRepoMock) Create(selection *domain.Selection) error { return nil }
func (m *statusStepSelectionRepoMock) GetByID(id string) (*domain.Selection, error) {
	return nil, repository.ErrSelectionNotFound
}
func (m *statusStepSelectionRepoMock) GetByCandidateID(candidateID string) (*domain.Selection, error) {
	if m.getByCandidateIDFn != nil {
		return m.getByCandidateIDFn(candidateID)
	}
	return nil, repository.ErrSelectionNotFound
}
func (m *statusStepSelectionRepoMock) GetByCandidateIDAndPairingID(candidateID, pairingID string) (*domain.Selection, error) {
	if m.getByCandidateIDFn != nil {
		return m.getByCandidateIDFn(candidateID)
	}
	return nil, repository.ErrSelectionNotFound
}
func (m *statusStepSelectionRepoMock) GetBySelectedBy(userID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *statusStepSelectionRepoMock) GetBySelectedByAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *statusStepSelectionRepoMock) GetByCandidateOwner(userID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *statusStepSelectionRepoMock) GetByCandidateOwnerAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *statusStepSelectionRepoMock) UpdateStatus(id string, status domain.SelectionStatus) error {
	return nil
}
func (m *statusStepSelectionRepoMock) GetExpiredSelections() ([]*domain.Selection, error) {
	return nil, nil
}

type statusStepDocumentRepoMock struct {
	documents []*domain.Document
	err       error
}

func (m *statusStepDocumentRepoMock) Create(document *domain.Document) error { return nil }
func (m *statusStepDocumentRepoMock) GetByCandidateID(candidateID string) ([]*domain.Document, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.documents, nil
}
func (m *statusStepDocumentRepoMock) Delete(id string) error { return nil }

func TestStatusStepService_UpdateStepRules(t *testing.T) {
	updated := false
	statusRepo := &statusStepRepoBehaviorMock{
		getByCandidateIDFn: func(candidateID string) ([]*domain.StatusStep, error) {
			return []*domain.StatusStep{
				{ID: "s1", CandidateID: candidateID, StepName: domain.MedicalTest, StepStatus: domain.Completed},
				{ID: "s2", CandidateID: candidateID, StepName: domain.LMISApproval, StepStatus: domain.InProgress},
			}, nil
		},
		updateFn: func(step *domain.StatusStep) error {
			updated = true
			return nil
		},
	}

	service := &StatusStepService{
		statusStepRepository: statusRepo,
		candidateRepository: &statusStepCandidateRepoMock{
			getByIDFn: func(id string) (*domain.Candidate, error) {
				return &domain.Candidate{ID: id, CreatedBy: "owner-1"}, nil
			},
		},
		selectionRepository: &statusStepSelectionRepoMock{},
		documentRepository: &statusStepDocumentRepoMock{
			documents: []*domain.Document{{CandidateID: "cand-1", DocumentType: domain.MedicalDocument, FileURL: "https://files.example/medical.pdf"}},
		},
		notificationService: &notificationSenderMock{foreignByID: map[string]bool{}},
	}

	err := service.UpdateStep("cand-1", domain.LMISApproval, "owner-1", domain.Completed, "done")
	require.NoError(t, err)
	assert.True(t, updated)
}

func TestStatusStepService_UpdateStepAuthorizationAndTransitions(t *testing.T) {
	steps := []*domain.StatusStep{{ID: "s1", CandidateID: "cand-1", StepName: domain.MedicalTest, StepStatus: domain.Pending}}
	statusRepo := &statusStepRepoBehaviorMock{
		getByCandidateIDFn: func(candidateID string) ([]*domain.StatusStep, error) { return steps, nil },
	}

	service := &StatusStepService{
		statusStepRepository: statusRepo,
		candidateRepository: &statusStepCandidateRepoMock{
			getByIDFn: func(id string) (*domain.Candidate, error) {
				return &domain.Candidate{ID: id, CreatedBy: "owner-1"}, nil
			},
		},
		selectionRepository: &statusStepSelectionRepoMock{},
		documentRepository: &statusStepDocumentRepoMock{
			documents: []*domain.Document{{CandidateID: "cand-1", DocumentType: domain.MedicalDocument, FileURL: "https://files.example/medical.pdf"}},
		},
		notificationService: &notificationSenderMock{foreignByID: map[string]bool{}},
	}

	err := service.UpdateStep("cand-1", domain.MedicalTest, "other-user", domain.InProgress, "")
	require.ErrorIs(t, err, ErrNotAuthorized)

	err = service.UpdateStep("cand-1", domain.MedicalTest, "owner-1", domain.Completed, "")
	require.NoError(t, err)
}

func TestStatusStepHelpers(t *testing.T) {
	assert.True(t, isValidStepStatus(domain.Pending))
	assert.True(t, isValidStepStatus(domain.Failed))
	assert.False(t, isValidStepStatus(domain.StepStatus("invalid")))

	assert.True(t, canTransitionStep(domain.Pending, domain.InProgress))
	assert.True(t, canTransitionStep(domain.InProgress, domain.Completed))
	assert.True(t, canTransitionStep(domain.InProgress, domain.Failed))
	assert.True(t, canTransitionStep(domain.Failed, domain.InProgress))
	assert.True(t, canTransitionStep(domain.Pending, domain.Completed))
	assert.True(t, canTransitionStep(domain.Completed, domain.Completed))
	assert.True(t, canTransitionStep(domain.Completed, domain.Pending))

	steps := predefinedStepNames()
	assert.Len(t, steps, len(predefinedStepNames()))
	assert.Equal(t, domain.MedicalTest, steps[0])
}

func TestStatusStepService_UpdateStepRequiresReasonWhenFailing(t *testing.T) {
	service := &StatusStepService{
		statusStepRepository: &statusStepRepoBehaviorMock{
			getByCandidateIDFn: func(candidateID string) ([]*domain.StatusStep, error) {
				return []*domain.StatusStep{{ID: "s1", CandidateID: candidateID, StepName: domain.Medical, StepStatus: domain.InProgress}}, nil
			},
		},
		candidateRepository: &statusStepCandidateRepoMock{
			getByIDFn: func(id string) (*domain.Candidate, error) {
				return &domain.Candidate{ID: id, CreatedBy: "owner-1", Status: domain.CandidateStatusInProgress}, nil
			},
		},
		selectionRepository: &statusStepSelectionRepoMock{},
		notificationService: &notificationSenderMock{foreignByID: map[string]bool{}},
	}

	err := service.UpdateStep("cand-1", domain.Medical, "owner-1", domain.Failed, "")
	require.ErrorIs(t, err, ErrStepFailureReasonRequired)
}

func TestStatusStepService_UpdateStepMarksCandidateCompletedWhenFinalStepFinishes(t *testing.T) {
	service, db := setupStatusStepServiceWithSQLite(t)

	require.NoError(t, db.Create(&statusStepTestCandidate{
		ID:        "cand-1",
		CreatedBy: "owner-1",
		FullName:  "Candidate",
		Status:    string(domain.CandidateStatusInProgress),
		Languages: []byte("[]"),
		Skills:    []byte("[]"),
	}).Error)

	now := time.Now().UTC()
	steps := []statusStepTestStep{
		{ID: "step-1", CandidateID: "cand-1", StepName: domain.Medical, StepStatus: string(domain.Completed), UpdatedBy: "owner-1", CreatedAt: now.Add(-9 * time.Minute), UpdatedAt: now.Add(-9 * time.Minute)},
		{ID: "step-2", CandidateID: "cand-1", StepName: domain.CoCPending, StepStatus: string(domain.Completed), UpdatedBy: "owner-1", CreatedAt: now.Add(-8 * time.Minute), UpdatedAt: now.Add(-8 * time.Minute)},
		{ID: "step-3", CandidateID: "cand-1", StepName: domain.CoCOnline, StepStatus: string(domain.Completed), UpdatedBy: "owner-1", CreatedAt: now.Add(-7 * time.Minute), UpdatedAt: now.Add(-7 * time.Minute)},
		{ID: "step-4", CandidateID: "cand-1", StepName: domain.LMISPending, StepStatus: string(domain.Completed), UpdatedBy: "owner-1", CreatedAt: now.Add(-6 * time.Minute), UpdatedAt: now.Add(-6 * time.Minute)},
		{ID: "step-5", CandidateID: "cand-1", StepName: domain.LMISIssued, StepStatus: string(domain.Completed), UpdatedBy: "owner-1", CreatedAt: now.Add(-5 * time.Minute), UpdatedAt: now.Add(-5 * time.Minute)},
		{ID: "step-6", CandidateID: "cand-1", StepName: domain.TicketPending, StepStatus: string(domain.Completed), UpdatedBy: "owner-1", CreatedAt: now.Add(-4 * time.Minute), UpdatedAt: now.Add(-4 * time.Minute)},
		{ID: "step-7", CandidateID: "cand-1", StepName: domain.TicketBooked, StepStatus: string(domain.Completed), UpdatedBy: "owner-1", CreatedAt: now.Add(-3 * time.Minute), UpdatedAt: now.Add(-3 * time.Minute)},
		{ID: "step-8", CandidateID: "cand-1", StepName: domain.TicketConfirmed, StepStatus: string(domain.Completed), UpdatedBy: "owner-1", CreatedAt: now.Add(-2 * time.Minute), UpdatedAt: now.Add(-2 * time.Minute)},
		{ID: "step-9", CandidateID: "cand-1", StepName: domain.Arrived, StepStatus: string(domain.InProgress), UpdatedBy: "owner-1", CreatedAt: now.Add(-1 * time.Minute), UpdatedAt: now.Add(-1 * time.Minute)},
	}
	for _, step := range steps {
		step := step
		require.NoError(t, db.Create(&step).Error)
	}

	require.NoError(t, db.Create(&statusStepTestSelection{
		ID:          "sel-1",
		CandidateID: "cand-1",
		SelectedBy:  "foreign-1",
		Status:      string(domain.SelectionApproved),
		ExpiresAt:   now.Add(24 * time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error)

	err := service.UpdateStep("cand-1", domain.Arrived, "owner-1", domain.Completed, "Completed arrival")
	require.NoError(t, err)

	var candidate statusStepTestCandidate
	require.NoError(t, db.Where("id = ?", "cand-1").First(&candidate).Error)
	assert.Equal(t, string(domain.CandidateStatusCompleted), candidate.Status)
}

func TestStatusStepService_UpdateStepAllowsApprovedCandidateToEnterProgress(t *testing.T) {
	service, db := setupStatusStepServiceWithSQLite(t)

	require.NoError(t, db.Create(&statusStepTestCandidate{
		ID:        "cand-approved",
		CreatedBy: "owner-1",
		FullName:  "Candidate",
		Status:    string(domain.CandidateStatusApproved),
		Languages: []byte("[]"),
		Skills:    []byte("[]"),
	}).Error)

	now := time.Now().UTC()
	require.NoError(t, db.Create(&statusStepTestStep{
		ID:          "step-medical",
		CandidateID: "cand-approved",
		StepName:    domain.Medical,
		StepStatus:  string(domain.Pending),
		UpdatedBy:   "owner-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error)

	require.NoError(t, db.Create(&statusStepTestSelection{
		ID:          "sel-approved",
		CandidateID: "cand-approved",
		SelectedBy:  "foreign-1",
		Status:      string(domain.SelectionApproved),
		ExpiresAt:   now.Add(24 * time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error)

	err := service.UpdateStep("cand-approved", domain.Medical, "owner-1", domain.InProgress, "Started medical processing")
	require.NoError(t, err)

	var candidate statusStepTestCandidate
	require.NoError(t, db.Where("id = ?", "cand-approved").First(&candidate).Error)
	assert.Equal(t, string(domain.CandidateStatusInProgress), candidate.Status)

	var step statusStepTestStep
	require.NoError(t, db.Where("candidate_id = ? AND step_name = ?", "cand-approved", domain.Medical).First(&step).Error)
	assert.Equal(t, string(domain.InProgress), step.StepStatus)
	assert.Equal(t, "Started medical processing", step.Notes)
}
