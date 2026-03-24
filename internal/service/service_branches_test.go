package service

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

type selectionQueryRepoMock struct {
	byID          *domain.Selection
	byIDErr       error
	bySelected    []*domain.Selection
	bySelectedErr error
	expired       []*domain.Selection
	expiredErr    error
	db            *gorm.DB
}

func (m *selectionQueryRepoMock) DB() *gorm.DB                             { return m.db }
func (m *selectionQueryRepoMock) Create(selection *domain.Selection) error { return nil }
func (m *selectionQueryRepoMock) GetByID(id string) (*domain.Selection, error) {
	return m.byID, m.byIDErr
}
func (m *selectionQueryRepoMock) GetByCandidateID(candidateID string) (*domain.Selection, error) {
	return nil, repository.ErrSelectionNotFound
}
func (m *selectionQueryRepoMock) GetByCandidateIDAndPairingID(candidateID, pairingID string) (*domain.Selection, error) {
	return nil, repository.ErrSelectionNotFound
}
func (m *selectionQueryRepoMock) GetBySelectedBy(userID string) ([]*domain.Selection, error) {
	return m.bySelected, m.bySelectedErr
}
func (m *selectionQueryRepoMock) GetBySelectedByAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	return m.bySelected, m.bySelectedErr
}
func (m *selectionQueryRepoMock) GetByCandidateOwner(userID string) ([]*domain.Selection, error) {
	return m.bySelected, m.bySelectedErr
}
func (m *selectionQueryRepoMock) GetByCandidateOwnerAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	return m.bySelected, m.bySelectedErr
}
func (m *selectionQueryRepoMock) UpdateStatus(id string, status domain.SelectionStatus) error {
	return nil
}
func (m *selectionQueryRepoMock) GetExpiredSelections() ([]*domain.Selection, error) {
	return m.expired, m.expiredErr
}

func TestSelectionService_QueryMethodsAndHelpers(t *testing.T) {
	selection := &domain.Selection{ID: "sel-1", CandidateID: "cand-1", SelectedBy: "foreign-1"}
	repo := &selectionQueryRepoMock{byID: selection, bySelected: []*domain.Selection{selection}}
	svc := &SelectionService{selectionRepository: repo}

	g, err := svc.GetSelection("sel-1")
	require.NoError(t, err)
	assert.Equal(t, "sel-1", g.ID)

	my, err := svc.GetMySelections("foreign-1")
	require.NoError(t, err)
	assert.Len(t, my, 1)

	_, err = svc.GetSelection("  ")
	require.Error(t, err)
	_, err = svc.GetMySelections("")
	require.Error(t, err)

	assert.True(t, isSelectionConflictError(repository.ErrActiveSelectionExists))
	assert.True(t, isSelectionConflictError(&pgconn.PgError{Code: "23505", Message: "idx_selections_one_pending_per_candidate"}))
	assert.False(t, isSelectionConflictError(errors.New("x")))
}

func TestSelectionService_ProcessExpiredSelections(t *testing.T) {
	dsn := "file:sel_exp_" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&selectionTestCandidate{}, &selectionTestSelection{}))

	require.NoError(t, db.Create(&selectionTestCandidate{
		ID:        "cand-1",
		CreatedBy: "owner-1",
		FullName:  "Candidate",
		Status:    string(domain.CandidateStatusLocked),
		LockedBy:  ptrString("foreign-1"),
		Languages: []byte("[]"),
		Skills:    []byte("[]"),
	}).Error)

	require.NoError(t, db.Create(&selectionTestSelection{
		ID:          "sel-1",
		CandidateID: "cand-1",
		SelectedBy:  "foreign-1",
		Status:      string(domain.SelectionPending),
		ExpiresAt:   time.Now().UTC().Add(-1 * time.Hour),
	}).Error)

	repo := &selectionQueryRepoMock{db: db, expired: []*domain.Selection{{ID: "sel-1", CandidateID: "cand-1", SelectedBy: "foreign-1"}}}
	notif := &notificationSenderMock{foreignByID: map[string]bool{}}
	svc := &SelectionService{selectionRepository: repo, notificationService: notif, db: db}

	require.NoError(t, svc.ProcessExpiredSelections())

	var updatedSel domain.Selection
	require.NoError(t, db.Where("id = ?", "sel-1").First(&updatedSel).Error)
	assert.Equal(t, domain.SelectionExpired, updatedSel.Status)
}

func TestApprovalService_GetApprovals(t *testing.T) {
	repo := &approvalRepositoryMock{}
	svc := &ApprovalService{approvalRepository: repo}
	_, err := svc.GetApprovals("")
	require.Error(t, err)
}

func TestCandidateService_RemainingBranches(t *testing.T) {
	repo := &candidateRepoBehaviorMock{}
	docRepo := &documentRepositoryMock{}
	storage := &storageServiceMock{}
	svc, err := NewCandidateService(repo, docRepo, storage, &PDFService{})
	require.NoError(t, err)

	_, _, err = svc.GetCandidate("x")
	require.Error(t, err)

	repo.getByID = func(id string) (*domain.Candidate, error) {
		return &domain.Candidate{ID: id, CreatedBy: "owner", Status: domain.CandidateStatusDraft, Languages: []byte("[]"), Skills: []byte("[]")}, nil
	}
	repo.updateFn = func(candidate *domain.Candidate) error { return nil }
	require.NoError(t, svc.PublishCandidate("id", "owner"))

	_, err = svc.UploadCandidateDocument("id", "owner", UploadCandidateDocumentInput{DocumentType: "photo", File: bytes.NewReader(validPNGBytes()), FileName: "p.png", FileSize: int64(len(validPNGBytes()))})
	require.NoError(t, err)

	_, err = svc.UploadCandidateDocument("id", "owner", UploadCandidateDocumentInput{DocumentType: "bad", File: bytes.NewBufferString("x"), FileName: "bad.bin", FileSize: 1})
	require.Error(t, err)

	err = svc.GenerateCV("id", "owner", CandidateCVBranding{})
	require.Error(t, err)

	_, err = parseCandidateDocumentType("photo")
	require.NoError(t, err)
	_, err = parseCandidateDocumentType("bad")
	require.Error(t, err)
}

func TestNotificationService_RemainingBranches(t *testing.T) {
	_, err := NewNotificationService(nil, &notificationRepoMock{}, &notificationEmailMock{}, &notificationUserRepoMock{}, &notificationCandidateRepoMock{}, &notificationSelectionRepoMock{})
	require.Error(t, err)

	notifRepo := &notificationRepoMock{}
	userRepo := &notificationUserRepoMock{users: map[string]*domain.User{
		"owner-1":    {ID: "owner-1", Email: ""},
		"selector-1": {ID: "selector-1", Email: "", CompanyName: "Agency"},
	}}
	candidateRepo := &notificationCandidateRepoMock{byID: map[string]*domain.Candidate{"cand-1": {ID: "cand-1", FullName: "Candidate", CreatedBy: "owner-1"}}}
	selectionRepo := &notificationSelectionRepoMock{
		byID: map[string]*domain.Selection{"sel-1": {ID: "sel-1", CandidateID: "cand-1", SelectedBy: "selector-1"}},
	}

	svc, err := NewNotificationService(&config.Config{}, notifRepo, &notificationEmailMock{}, userRepo, candidateRepo, selectionRepo)
	require.NoError(t, err)
	require.NoError(t, svc.NotifySelection("cand-1", "selector-1"))
	require.NoError(t, svc.NotifyRejection("sel-1", "owner-1"))
	require.NoError(t, svc.NotifyExpiry("sel-1"))
	assert.True(t, len(notifRepo.created) >= 5)
}

func TestStatusStepService_Branches(t *testing.T) {
	svc := &StatusStepService{statusStepRepository: &statusStepRepoBehaviorMock{getByCandidateIDFn: func(candidateID string) ([]*domain.StatusStep, error) {
		return []*domain.StatusStep{}, nil
	}}}
	_, err := svc.GetCandidateProgress("")
	require.Error(t, err)

	_, err = svc.GetCandidateProgress("cand-1")
	require.NoError(t, err)
}

func TestPDFService_GenerateCV_ImageDecodeFailure(t *testing.T) {
	img, _ := base64.StdEncoding.DecodeString("/9j/4AAQSkZJRgABAQAAAQABAAD/2wCEAAkGBxISEhUTEhMWFhUXGBgYFxgXFxgYGBgYGBgXGBgYGBgYHSggGBolGxgYITEhJSkrLi4uGB8zODMsNygtLisBCgoKDg0OGxAQGy0lICYtLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLf/AABEIAAEAAQMBIgACEQEDEQH/xAAbAAADAQADAQAAAAAAAAAAAAAEBQYDAgABB//EADUQAAEDAgMFBQYEBwAAAAAAAAEAAgMEEQUSITFBUQYTImFxgZGhFCMyQrHB0fAjYnLxFf/EABkBAQADAQEAAAAAAAAAAAAAAAABAgMEBf/EACQRAQACAgICAgMBAAAAAAAAAAABAgMREiExBEETIlFhFDKB/9oADAMBAAIRAxEAPwD3iIiAiIgIiICIiAiIgIiICIiAiIgL//2Q==")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write(img)
	}))
	defer server.Close()

	age := 21
	exp := 2
	svc := NewPDFService()
	_, err := svc.GenerateCandidateCV(&domain.Candidate{FullName: "C", Age: &age, ExperienceYears: &exp, Languages: []byte(`["en"]`), Skills: []byte(`["s1"]`)}, []*domain.Document{
		{DocumentType: domain.Photo, FileURL: server.URL},
		{DocumentType: domain.Passport, FileURL: server.URL},
		{DocumentType: domain.Video, FileURL: "https://example.com/video.mp4"},
	}, CandidateCVBranding{})
	require.Error(t, err)
}
