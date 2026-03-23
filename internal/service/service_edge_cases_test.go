package service

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

type selectionNotifierErrMock struct {
	foreign       map[string]bool
	isForeignErr  error
	sendErr       error
	sendCallCount int
}

func (m *selectionNotifierErrMock) IsForeignAgent(userID string) (bool, error) {
	if m.isForeignErr != nil {
		return false, m.isForeignErr
	}
	return m.foreign[userID], nil
}

func (m *selectionNotifierErrMock) Send(userID, title, message, notificationType, relatedEntityType, relatedEntityID string) error {
	m.sendCallCount++
	return m.sendErr
}

type notificationRepoErrMock struct {
	created    []*domain.Notification
	createErr  error
	failOnCall int
	callCount  int
}

func (m *notificationRepoErrMock) Create(notification *domain.Notification) error {
	m.callCount++
	if m.failOnCall > 0 && m.callCount == m.failOnCall {
		return errors.New("create failed on call")
	}
	if m.createErr != nil {
		return m.createErr
	}
	m.created = append(m.created, notification)
	return nil
}
func (m *notificationRepoErrMock) GetByUserID(userID string, unreadOnly bool) ([]*domain.Notification, error) {
	return nil, nil
}
func (m *notificationRepoErrMock) MarkAsRead(id string) error        { return nil }
func (m *notificationRepoErrMock) MarkAllAsRead(userID string) error { return nil }

func newValidImageServer(t *testing.T) *httptest.Server {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	img.Set(1, 1, color.RGBA{G: 255, A: 255})

	buf := bytes.NewBuffer(nil)
	require.NoError(t, png.Encode(buf, img))

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(buf.Bytes())
	}))
}

func TestCandidateService_UpdateAndGenerateCV_ErrorBranches(t *testing.T) {
	repo := &candidateRepoBehaviorMock{}
	docRepo := &candidateDocRepoBehaviorMock{}
	storage := &candidateStorageBehaviorMock{}
	svc, err := NewCandidateService(repo, docRepo, storage, NewPDFService())
	require.NoError(t, err)

	err = svc.UpdateCandidate("", "owner", CandidateInput{FullName: "A"})
	require.Error(t, err)
	err = svc.UpdateCandidate("id", "", CandidateInput{FullName: "A"})
	require.ErrorIs(t, err, ErrForbidden)

	repo.getByID = func(id string) (*domain.Candidate, error) {
		return &domain.Candidate{ID: id, CreatedBy: "owner-1", Status: domain.CandidateStatusDraft}, nil
	}
	err = svc.UpdateCandidate("id", "other", CandidateInput{FullName: "A"})
	require.ErrorIs(t, err, ErrForbidden)

	repo.getByID = func(id string) (*domain.Candidate, error) {
		return &domain.Candidate{ID: id, CreatedBy: "owner-1", Status: domain.CandidateStatusInProgress}, nil
	}
	err = svc.UpdateCandidate("id", "owner-1", CandidateInput{FullName: "A"})
	require.ErrorIs(t, err, ErrInvalidCandidateUpdateState)

	lockOwner := "foreign-1"
	lockExpiry := time.Now().UTC().Add(1 * time.Hour)
	repo.getByID = func(id string) (*domain.Candidate, error) {
		return &domain.Candidate{ID: id, CreatedBy: "owner-1", Status: domain.CandidateStatusDraft, LockedBy: &lockOwner, LockExpiresAt: &lockExpiry}, nil
	}
	err = svc.UpdateCandidate("id", "owner-1", CandidateInput{FullName: "A"})
	require.ErrorIs(t, err, ErrCandidateLocked)

	repo.getByID = func(id string) (*domain.Candidate, error) {
		return &domain.Candidate{ID: id, CreatedBy: "owner-1", FullName: "C", Status: domain.CandidateStatusDraft, Languages: []byte(`["en"]`), Skills: []byte(`["s"]`)}, nil
	}
	docRepo.getByCandidateFn = func(candidateID string) ([]*domain.Document, error) {
		return nil, errors.New("docs fail")
	}
	err = svc.GenerateCV("id", "owner-1", CandidateCVBranding{})
	require.Error(t, err)

	server := newValidImageServer(t)
	defer server.Close()

	docRepo.getByCandidateFn = func(candidateID string) ([]*domain.Document, error) {
		return []*domain.Document{
			{DocumentType: domain.Photo, FileURL: server.URL},
			{DocumentType: domain.Passport, FileURL: server.URL},
			{DocumentType: domain.Video, FileURL: "https://example.com/video.mp4"},
		}, nil
	}

	storage.uploadFn = func(fileName, contentType string) (string, error) {
		return "", errors.New("upload fail")
	}
	err = svc.GenerateCV("id", "owner-1", CandidateCVBranding{})
	require.Error(t, err)

	deleted := false
	storage.uploadFn = func(fileName, contentType string) (string, error) {
		return "https://files/cv.pdf", nil
	}
	storage.deleteFn = func(url string) error {
		deleted = true
		return nil
	}
	repo.updateFn = func(candidate *domain.Candidate) error { return errors.New("update fail") }
	err = svc.GenerateCV("id", "owner-1", CandidateCVBranding{})
	require.Error(t, err)
	assert.True(t, deleted)
}

func TestApprovalAndSelection_DecisionErrorBranches(t *testing.T) {
	approvalSvc, db := setupApprovalService(t)
	seedApprovalScenario(t, db, domain.SelectionPending, time.Now().UTC().Add(1*time.Hour))

	err := approvalSvc.ApproveSelection("sel-1", "unrelated-user")
	require.ErrorIs(t, err, ErrNotAuthorized)

	err = db.Create(&domain.Approval{ID: "ap-rej", SelectionID: "sel-1", UserID: "owner-1", Decision: domain.ApprovalRejected, DecidedAt: time.Now().UTC()}).Error
	require.NoError(t, err)
	err = approvalSvc.ApproveSelection("sel-1", "owner-1")
	require.ErrorIs(t, err, ErrAlreadyDecided)

	selectionSvc, selectionDB, _ := setupSelectionService(t)
	seedCandidate(t, selectionDB, "cand-err", "owner-1", domain.CandidateStatusAvailable)
	notifier := &selectionNotifierErrMock{foreign: map[string]bool{"foreign-1": true}, isForeignErr: errors.New("lookup fail")}
	selectionSvc.notificationService = notifier

	_, err = selectionSvc.SelectCandidate("cand-err", "foreign-1")
	require.Error(t, err)

	notifier = &selectionNotifierErrMock{foreign: map[string]bool{"foreign-1": true}, sendErr: errors.New("send fail")}
	selectionSvc.notificationService = notifier
	_, err = selectionSvc.SelectCandidate("cand-err", "foreign-1")
	require.Error(t, err)
}

func TestNotificationService_EdgeFailures(t *testing.T) {
	userRepo := &notificationUserRepoMock{users: map[string]*domain.User{
		"owner":    {ID: "owner", Email: "", FullName: "Owner"},
		"selector": {ID: "selector", Email: "", FullName: "Selector"},
	}}
	candidateRepo := &notificationCandidateRepoMock{byID: map[string]*domain.Candidate{
		"cand-1": {ID: "cand-1", FullName: "Candidate", CreatedBy: "owner"},
	}}
	selectionRepo := &notificationSelectionRepoMock{
		byID: map[string]*domain.Selection{"sel-1": {ID: "sel-1", CandidateID: "cand-1", SelectedBy: "selector"}},
	}

	repo := &notificationRepoErrMock{}
	svc, err := NewNotificationService(&config.Config{}, repo, &notificationEmailMock{}, userRepo, candidateRepo, selectionRepo)
	require.NoError(t, err)

	err = svc.Send("", "t", "m", "selection", "candidate", "cand-1")
	require.Error(t, err)

	repo.createErr = errors.New("create fail")
	err = svc.Send("owner", "t", "m", "selection", "candidate", "cand-1")
	require.Error(t, err)
	repo.createErr = nil

	repo.failOnCall = 2
	err = svc.NotifyApproval("sel-1")
	require.Error(t, err)

	selectionRepo.byCandidateID = nil
	err = svc.NotifyStatusUpdate("cand-missing", "Medical")
	require.Error(t, err)

	selectionRepo.byCandidateID = map[string]*domain.Selection{"cand-1": {ID: "sel-1", CandidateID: "cand-1", SelectedBy: "selector"}}
	err = svc.NotifySelection("cand-1", "selector")
	require.NoError(t, err)
	assert.NotEmpty(t, repo.created)
	assert.Contains(t, repo.created[len(repo.created)-1].Message, "Selector")
}
