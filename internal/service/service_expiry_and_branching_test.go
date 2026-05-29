package service

import (
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

type approvalRepositoryWithDataMock struct {
	approvals []*domain.Approval
}

func (m *approvalRepositoryWithDataMock) DB() *gorm.DB                           { return nil }
func (m *approvalRepositoryWithDataMock) Create(approval *domain.Approval) error { return nil }
func (m *approvalRepositoryWithDataMock) GetBySelectionID(selectionID string) ([]*domain.Approval, error) {
	return m.approvals, nil
}
func (m *approvalRepositoryWithDataMock) GetBySelectionAndUser(selectionID, userID string) (*domain.Approval, error) {
	return nil, nil
}

type approvalRepositoryErrMock struct{}

func (m *approvalRepositoryErrMock) DB() *gorm.DB                           { return nil }
func (m *approvalRepositoryErrMock) Create(approval *domain.Approval) error { return nil }
func (m *approvalRepositoryErrMock) GetBySelectionID(selectionID string) ([]*domain.Approval, error) {
	return nil, errors.New("list approvals failed")
}
func (m *approvalRepositoryErrMock) GetBySelectionAndUser(selectionID, userID string) (*domain.Approval, error) {
	return nil, nil
}

type selectionRepoExpiredOnlyMock struct {
	expired []*domain.Selection
}

func (m *selectionRepoExpiredOnlyMock) DB() *gorm.DB                             { return nil }
func (m *selectionRepoExpiredOnlyMock) Create(selection *domain.Selection) error { return nil }
func (m *selectionRepoExpiredOnlyMock) GetByID(id string) (*domain.Selection, error) {
	return nil, nil
}
func (m *selectionRepoExpiredOnlyMock) GetByCandidateID(candidateID string) (*domain.Selection, error) {
	return nil, nil
}
func (m *selectionRepoExpiredOnlyMock) GetByCandidateIDAndPairingID(candidateID, pairingID string) (*domain.Selection, error) {
	return nil, nil
}
func (m *selectionRepoExpiredOnlyMock) GetBySelectedBy(userID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *selectionRepoExpiredOnlyMock) GetBySelectedByAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *selectionRepoExpiredOnlyMock) GetByCandidateOwner(userID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *selectionRepoExpiredOnlyMock) GetByCandidateOwnerAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *selectionRepoExpiredOnlyMock) UpdateStatus(id string, status domain.SelectionStatus) error {
	return nil
}
func (m *selectionRepoExpiredOnlyMock) GetExpiredSelections() ([]*domain.Selection, error) {
	return m.expired, nil
}

type statusStepRepoWithDBMock struct {
	db *gorm.DB
}

func (m *statusStepRepoWithDBMock) DB() *gorm.DB                         { return m.db }
func (m *statusStepRepoWithDBMock) Create(step *domain.StatusStep) error { return nil }
func (m *statusStepRepoWithDBMock) GetByCandidateID(candidateID string) ([]*domain.StatusStep, error) {
	return nil, nil
}
func (m *statusStepRepoWithDBMock) Update(step *domain.StatusStep) error { return nil }

func TestApprovalService_GetApprovals_ReturnsDataAndError(t *testing.T) {
	okSvc := &ApprovalService{approvalRepository: &approvalRepositoryWithDataMock{approvals: []*domain.Approval{{ID: "a1", SelectionID: "sel-1"}}}}
	approvals, err := okSvc.GetApprovals("sel-1")
	require.NoError(t, err)
	require.Len(t, approvals, 1)
	assert.Equal(t, "a1", approvals[0].ID)

	errSvc := &ApprovalService{approvalRepository: &approvalRepositoryErrMock{}}
	_, err = errSvc.GetApprovals("sel-1")
	require.Error(t, err)
}

func TestSelectionService_ProcessExpiredSelections_NilEntrySkipped(t *testing.T) {
	repo := &selectionRepoExpiredOnlyMock{expired: []*domain.Selection{nil}}
	svc := &SelectionService{selectionRepository: repo, notificationService: &notificationSenderMock{foreignByID: map[string]bool{}}}
	require.NoError(t, svc.ProcessExpiredSelections())
}

func TestSelectionService_ProcessExpiredSelections_SelectionMissingInTx(t *testing.T) {
	dsn := "file:sel_exp_miss_" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&selectionTestCandidate{}, &selectionTestSelection{}))

	repo := &selectionQueryRepoMock{db: db, expired: []*domain.Selection{{ID: "sel-missing", CandidateID: "cand-missing", SelectedBy: "foreign-1"}}}
	svc := &SelectionService{selectionRepository: repo, notificationService: &notificationSenderMock{foreignByID: map[string]bool{}}, db: db}
	require.NoError(t, svc.ProcessExpiredSelections())
}

func TestSelectionService_ProcessExpiredSelections_NotPendingOrNotExpiredBranches(t *testing.T) {
	dsn := "file:sel_exp_branches_" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&selectionTestCandidate{}, &selectionTestSelection{}))

	future := time.Now().UTC().Add(2 * time.Hour)
	require.NoError(t, db.Create(&selectionTestSelection{ID: "sel-future", CandidateID: "cand-x", SelectedBy: "foreign-1", Status: string(domain.SelectionPending), ExpiresAt: future}).Error)
	require.NoError(t, db.Create(&selectionTestSelection{ID: "sel-approved", CandidateID: "cand-y", SelectedBy: "foreign-1", Status: string(domain.SelectionApproved), ExpiresAt: time.Now().UTC().Add(-1 * time.Hour)}).Error)

	repo := &selectionQueryRepoMock{db: db, expired: []*domain.Selection{{ID: "sel-future", CandidateID: "cand-x", SelectedBy: "foreign-1"}, {ID: "sel-approved", CandidateID: "cand-y", SelectedBy: "foreign-1"}}}
	svc := &SelectionService{selectionRepository: repo, notificationService: &notificationSenderMock{foreignByID: map[string]bool{}}, db: db}
	require.NoError(t, svc.ProcessExpiredSelections())

	var selFuture domain.Selection
	require.NoError(t, db.Where("id = ?", "sel-future").First(&selFuture).Error)
	assert.Equal(t, domain.SelectionPending, selFuture.Status)

	var selApproved domain.Selection
	require.NoError(t, db.Where("id = ?", "sel-approved").First(&selApproved).Error)
	assert.Equal(t, domain.SelectionApproved, selApproved.Status)
}

func TestSelectionService_ProcessExpiredSelections_CandidateMissingDoesNotFail(t *testing.T) {
	dsn := "file:sel_exp_cand_missing_" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&selectionTestCandidate{}, &selectionTestSelection{}))

	require.NoError(t, db.Create(&selectionTestSelection{ID: "sel-1", CandidateID: "cand-missing", SelectedBy: "foreign-1", Status: string(domain.SelectionPending), ExpiresAt: time.Now().UTC().Add(-1 * time.Hour)}).Error)

	repo := &selectionQueryRepoMock{db: db, expired: []*domain.Selection{{ID: "sel-1", CandidateID: "cand-missing", SelectedBy: "foreign-1"}}}
	svc := &SelectionService{selectionRepository: repo, notificationService: &notificationSenderMock{foreignByID: map[string]bool{}}, db: db}
	require.NoError(t, svc.ProcessExpiredSelections())

	var sel domain.Selection
	require.NoError(t, db.Where("id = ?", "sel-1").First(&sel).Error)
	assert.Equal(t, domain.SelectionExpired, sel.Status)
}

func TestStatusStepService_InitializeSteps_CreatesOnlyMissing(t *testing.T) {
	dsn := "file:steps_init_more_" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&approvalTestCandidate{}, &approvalTestStatusStep{}))

	require.NoError(t, db.Create(&approvalTestCandidate{ID: "cand-1", CreatedBy: "owner-1", FullName: "C", Status: string(domain.CandidateStatusInProgress), Languages: []byte("[]"), Skills: []byte("[]")}).Error)
	require.NoError(t, db.Create(&approvalTestStatusStep{ID: "s-existing", CandidateID: "cand-1", StepName: domain.MedicalTest, StepStatus: string(domain.Pending), UpdatedBy: "owner-1", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}).Error)

	svc, err := NewStatusStepService(
		&statusStepRepoWithDBMock{db: db},
		&statusStepCandidateRepoMock{getByIDFn: func(id string) (*domain.Candidate, error) {
			return &domain.Candidate{ID: id, CreatedBy: "owner-1"}, nil
		}},
		&statusStepSelectionRepoMock{},
		&notificationSenderMock{foreignByID: map[string]bool{}},
	)
	require.NoError(t, err)

	require.NoError(t, svc.InitializeSteps("cand-1"))

	var rows []approvalTestStatusStep
	require.NoError(t, db.Where("candidate_id = ?", "cand-1").Find(&rows).Error)
	assert.Len(t, rows, len(predefinedStepNames()))
}

func TestStatusStepService_UpdateStep_CompletedSetsAndClearsTimestamp(t *testing.T) {
	var updated []*domain.StatusStep
	repo := &statusStepRepoBehaviorMock{
		getByCandidateIDFn: func(candidateID string) ([]*domain.StatusStep, error) {
			return []*domain.StatusStep{{ID: "s1", CandidateID: candidateID, StepName: domain.MedicalTest, StepStatus: domain.InProgress}}, nil
		},
		updateFn: func(step *domain.StatusStep) error {
			updated = append(updated, &domain.StatusStep{ID: step.ID, CandidateID: step.CandidateID, StepName: step.StepName, StepStatus: step.StepStatus, CompletedAt: step.CompletedAt})
			return nil
		},
	}

	svc := &StatusStepService{
		statusStepRepository: repo,
		candidateRepository: &statusStepCandidateRepoMock{getByIDFn: func(id string) (*domain.Candidate, error) {
			return &domain.Candidate{ID: id, CreatedBy: "owner-1"}, nil
		}},
		selectionRepository: &statusStepSelectionRepoMock{},
		documentRepository: &statusStepDocumentRepoMock{
			documents: []*domain.Document{{CandidateID: "cand-1", DocumentType: domain.MedicalDocument, FileURL: "https://files.example/medical.pdf"}},
		},
		notificationService: &notificationSenderMock{foreignByID: map[string]bool{}},
	}

	require.NoError(t, svc.UpdateStep("cand-1", domain.MedicalTest, "owner-1", domain.Completed, "done"))
	require.Len(t, updated, 1)
	require.NotNil(t, updated[0].CompletedAt)

	repo.getByCandidateIDFn = func(candidateID string) ([]*domain.StatusStep, error) {
		return []*domain.StatusStep{{ID: "s1", CandidateID: candidateID, StepName: domain.MedicalTest, StepStatus: domain.Completed, CompletedAt: updated[0].CompletedAt}}, nil
	}
	require.NoError(t, svc.UpdateStep("cand-1", domain.MedicalTest, "owner-1", domain.Pending, "reopen"))
	require.Len(t, updated, 2)
	require.Nil(t, updated[1].CompletedAt)
}

func TestNotificationService_NotifyRejectionUnknownPartyAndForeignCheckError(t *testing.T) {
	userRepo := &notificationUserRepoMock{users: map[string]*domain.User{
		"owner":    {ID: "owner", FullName: "Owner"},
		"selector": {ID: "selector", FullName: "Selector"},
	}}
	candidateRepo := &notificationCandidateRepoMock{byID: map[string]*domain.Candidate{"cand": {ID: "cand", CreatedBy: "owner", FullName: "Candidate"}}}
	selectionRepo := &notificationSelectionRepoMock{byID: map[string]*domain.Selection{"sel": {ID: "sel", CandidateID: "cand", SelectedBy: "selector"}}}
	repo := &notificationRepoErrMock{}

	svc, err := NewNotificationService(&config.Config{}, repo, &notificationEmailMock{}, userRepo, candidateRepo, selectionRepo)
	require.NoError(t, err)
	require.NoError(t, svc.NotifyRejection("sel", "unknown"))

	_, err = svc.IsForeignAgent("missing-user")
	require.Error(t, err)
}

type statusStepRepoNoDBMock struct{}

func (m *statusStepRepoNoDBMock) Create(step *domain.StatusStep) error { return nil }
func (m *statusStepRepoNoDBMock) GetByCandidateID(candidateID string) ([]*domain.StatusStep, error) {
	return nil, nil
}
func (m *statusStepRepoNoDBMock) Update(step *domain.StatusStep) error { return nil }

func TestNewStatusStepService_RepoWithoutTransactionDB(t *testing.T) {
	_, err := NewStatusStepService(
		&statusStepRepoNoDBMock{},
		&candidateRepositoryMock{},
		&selectionRepositoryMock{},
		&notificationSenderMock{foreignByID: map[string]bool{}},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not expose transaction db")
}

func TestApprovalService_LoadSelectionAndCandidateForDecision_Branches(t *testing.T) {
	dsn := "file:approval_load_branches_" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&approvalTestCandidate{}, &approvalTestSelection{}))

	svc := &ApprovalService{db: db}

	_, _, err = svc.loadSelectionAndCandidateForDecision(db, "missing")
	require.ErrorIs(t, err, repository.ErrSelectionNotFound)

	require.NoError(t, db.Create(&approvalTestSelection{ID: "sel-rejected", CandidateID: "cand-1", SelectedBy: "selector", Status: string(domain.SelectionRejected), ExpiresAt: time.Now().UTC().Add(2 * time.Hour)}).Error)
	_, _, err = svc.loadSelectionAndCandidateForDecision(db, "sel-rejected")
	require.ErrorIs(t, err, ErrSelectionNotPending)

	require.NoError(t, db.Create(&approvalTestSelection{ID: "sel-expired", CandidateID: "cand-1", SelectedBy: "selector", Status: string(domain.SelectionPending), ExpiresAt: time.Now().UTC().Add(-1 * time.Hour)}).Error)
	_, _, err = svc.loadSelectionAndCandidateForDecision(db, "sel-expired")
	require.ErrorIs(t, err, ErrSelectionNotPending)

	require.NoError(t, db.Create(&approvalTestSelection{ID: "sel-no-cand", CandidateID: "cand-missing", SelectedBy: "selector", Status: string(domain.SelectionPending), ExpiresAt: time.Now().UTC().Add(2 * time.Hour)}).Error)
	_, _, err = svc.loadSelectionAndCandidateForDecision(db, "sel-no-cand")
	require.ErrorIs(t, err, repository.ErrCandidateNotFound)

	require.NoError(t, db.Create(&approvalTestCandidate{ID: "cand-ok", CreatedBy: "owner", FullName: "C", Status: string(domain.CandidateStatusLocked), Languages: []byte("[]"), Skills: []byte("[]")}).Error)
	require.NoError(t, db.Create(&approvalTestSelection{ID: "sel-ok", CandidateID: "cand-ok", SelectedBy: "selector", Status: string(domain.SelectionPending), ExpiresAt: time.Now().UTC().Add(2 * time.Hour)}).Error)
	sel, cand, err := svc.loadSelectionAndCandidateForDecision(db, "sel-ok")
	require.NoError(t, err)
	require.NotNil(t, sel)
	require.NotNil(t, cand)
	assert.Equal(t, "cand-ok", sel.CandidateID)
	assert.Equal(t, "owner", cand.CreatedBy)
}

func TestCandidateParseDocumentType_AllSupportedValues(t *testing.T) {
	value, err := parseCandidateDocumentType(string(domain.Passport))
	require.NoError(t, err)
	assert.Equal(t, domain.Passport, value)

	value, err = parseCandidateDocumentType(string(domain.Photo))
	require.NoError(t, err)
	assert.Equal(t, domain.Photo, value)

	value, err = parseCandidateDocumentType(string(domain.Video))
	require.NoError(t, err)
	assert.Equal(t, domain.Video, value)
}

func TestNotificationService_NotifyExpiry_SecondSendFailure(t *testing.T) {
	userRepo := &notificationUserRepoMock{users: map[string]*domain.User{
		"owner":    {ID: "owner", FullName: "Owner"},
		"selector": {ID: "selector", FullName: "Selector"},
	}}
	candidateRepo := &notificationCandidateRepoMock{byID: map[string]*domain.Candidate{"cand": {ID: "cand", CreatedBy: "owner", FullName: "Candidate"}}}
	selectionRepo := &notificationSelectionRepoMock{byID: map[string]*domain.Selection{"sel": {ID: "sel", CandidateID: "cand", SelectedBy: "selector"}}}
	repo := &notificationRepoErrMock{failOnCall: 2}

	svc, err := NewNotificationService(&config.Config{}, repo, &notificationEmailMock{}, userRepo, candidateRepo, selectionRepo)
	require.NoError(t, err)
	err = svc.NotifyExpiry("sel")
	require.Error(t, err)
}

func TestS3StorageService_PublicURL_DefaultBranch(t *testing.T) {
	svc := &S3StorageService{bucket: "bucket", region: "us-east-1", endpoint: ""}
	url := svc.publicURL("documents/file.pdf")
	assert.Equal(t, "https://bucket.s3.us-east-1.amazonaws.com/documents/file.pdf", url)
}
