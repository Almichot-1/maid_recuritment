package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

type notificationRepoMock struct {
	created []*domain.Notification
}

func (m *notificationRepoMock) Create(notification *domain.Notification) error {
	m.created = append(m.created, notification)
	return nil
}
func (m *notificationRepoMock) GetByUserID(userID string, unreadOnly bool) ([]*domain.Notification, error) {
	return nil, nil
}
func (m *notificationRepoMock) MarkAsRead(id string) error       { return nil }
func (m *notificationRepoMock) MarkAllAsRead(userID string) error { return nil }

type notificationEmailMock struct{}

func (m *notificationEmailMock) Send(to, subject, body string) error { return nil }

type notificationUserRepoMock struct {
	users map[string]*domain.User
}

func (m *notificationUserRepoMock) Create(user *domain.User) error { return nil }
func (m *notificationUserRepoMock) GetByEmail(email string) (*domain.User, error) {
	return nil, repository.ErrUserNotFound
}
func (m *notificationUserRepoMock) GetByID(id string) (*domain.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, repository.ErrUserNotFound
}
func (m *notificationUserRepoMock) Update(user *domain.User) error { return nil }

type notificationCandidateRepoMock struct {
	byID map[string]*domain.Candidate
}

func (m *notificationCandidateRepoMock) Create(candidate *domain.Candidate) error { return nil }
func (m *notificationCandidateRepoMock) GetByID(id string) (*domain.Candidate, error) {
	if candidate, ok := m.byID[id]; ok {
		return candidate, nil
	}
	return nil, repository.ErrCandidateNotFound
}
func (m *notificationCandidateRepoMock) List(filters domain.CandidateFilters) ([]*domain.Candidate, error) {
	return nil, nil
}
func (m *notificationCandidateRepoMock) Update(candidate *domain.Candidate) error { return nil }
func (m *notificationCandidateRepoMock) Delete(id string) error                   { return nil }
func (m *notificationCandidateRepoMock) Lock(candidateID, lockedBy string, expiresAt time.Time) error {
	return nil
}
func (m *notificationCandidateRepoMock) Unlock(candidateID string) error { return nil }

type notificationSelectionRepoMock struct {
	byID          map[string]*domain.Selection
	byCandidateID map[string]*domain.Selection
}

func (m *notificationSelectionRepoMock) Create(selection *domain.Selection) error { return nil }
func (m *notificationSelectionRepoMock) GetByID(id string) (*domain.Selection, error) {
	if selection, ok := m.byID[id]; ok {
		return selection, nil
	}
	return nil, repository.ErrSelectionNotFound
}
func (m *notificationSelectionRepoMock) GetByCandidateID(candidateID string) (*domain.Selection, error) {
	if selection, ok := m.byCandidateID[candidateID]; ok {
		return selection, nil
	}
	return nil, repository.ErrSelectionNotFound
}
func (m *notificationSelectionRepoMock) GetByCandidateIDAndPairingID(candidateID, pairingID string) (*domain.Selection, error) {
	return m.GetByCandidateID(candidateID)
}
func (m *notificationSelectionRepoMock) GetBySelectedBy(userID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *notificationSelectionRepoMock) GetBySelectedByAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *notificationSelectionRepoMock) GetByCandidateOwner(userID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *notificationSelectionRepoMock) GetByCandidateOwnerAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *notificationSelectionRepoMock) UpdateStatus(id string, status domain.SelectionStatus) error { return nil }
func (m *notificationSelectionRepoMock) GetExpiredSelections() ([]*domain.Selection, error) {
	return nil, nil
}

type realtimeNotifierMock struct {
	pushes int
}

func (m *realtimeNotifierMock) PushToUser(userID string, notification *domain.Notification) {
	m.pushes++
}

func TestNotificationService_SendAndLinks(t *testing.T) {
	notifRepo := &notificationRepoMock{}
	userRepo := &notificationUserRepoMock{
		users: map[string]*domain.User{
			"u1": {ID: "u1", Email: "", Role: domain.ForeignAgent},
		},
	}
	service, err := NewNotificationService(
		&config.Config{AppBaseURL: "https://app.example.com/"},
		notifRepo,
		&notificationEmailMock{},
		userRepo,
		&notificationCandidateRepoMock{byID: map[string]*domain.Candidate{}},
		&notificationSelectionRepoMock{byID: map[string]*domain.Selection{}, byCandidateID: map[string]*domain.Selection{}},
	)
	require.NoError(t, err)

	rt := &realtimeNotifierMock{}
	service.SetRealtimeNotifier(rt)

	err = service.Send("u1", "Title", "Body", string(domain.NotificationSelection), "candidate", "cand-1")
	require.NoError(t, err)
	require.Len(t, notifRepo.created, 1)
	assert.Equal(t, 1, rt.pushes)
	assert.Equal(t, "/candidates/cand-1", (&NotificationService{}).candidateLink("cand-1"))
	assert.Equal(t, "https://app.example.com/candidates/cand-1", service.candidateLink("cand-1"))
	assert.Equal(t, "https://app.example.com/selections/sel-1", service.selectionLink("sel-1"))
}

func TestNotificationService_IsForeignAgent(t *testing.T) {
	notifRepo := &notificationRepoMock{}
	userRepo := &notificationUserRepoMock{
		users: map[string]*domain.User{
			"foreign":   {ID: "foreign", Role: domain.ForeignAgent},
			"ethiopian": {ID: "ethiopian", Role: domain.EthiopianAgent},
		},
	}
	service, err := NewNotificationService(
		&config.Config{},
		notifRepo,
		&notificationEmailMock{},
		userRepo,
		&notificationCandidateRepoMock{byID: map[string]*domain.Candidate{}},
		&notificationSelectionRepoMock{byID: map[string]*domain.Selection{}, byCandidateID: map[string]*domain.Selection{}},
	)
	require.NoError(t, err)

	isForeign, err := service.IsForeignAgent("foreign")
	require.NoError(t, err)
	assert.True(t, isForeign)

	isForeign, err = service.IsForeignAgent("ethiopian")
	require.NoError(t, err)
	assert.False(t, isForeign)
}

func TestNotificationService_NotifyApprovalAndStatusUpdate(t *testing.T) {
	notifRepo := &notificationRepoMock{}
	userRepo := &notificationUserRepoMock{
		users: map[string]*domain.User{
			"owner-1":    {ID: "owner-1", Email: ""},
			"selector-1": {ID: "selector-1", Email: ""},
		},
	}
	candidateRepo := &notificationCandidateRepoMock{
		byID: map[string]*domain.Candidate{
			"cand-1": {ID: "cand-1", CreatedBy: "owner-1", FullName: "Candidate"},
		},
	}
	selectionRepo := &notificationSelectionRepoMock{
		byID: map[string]*domain.Selection{
			"sel-1": {ID: "sel-1", CandidateID: "cand-1", SelectedBy: "selector-1"},
		},
		byCandidateID: map[string]*domain.Selection{
			"cand-1": {ID: "sel-1", CandidateID: "cand-1", SelectedBy: "selector-1"},
		},
	}

	service, err := NewNotificationService(
		&config.Config{AppBaseURL: "https://app.example.com"},
		notifRepo,
		&notificationEmailMock{},
		userRepo,
		candidateRepo,
		selectionRepo,
	)
	require.NoError(t, err)

	require.NoError(t, service.NotifyApproval("sel-1"))
	require.NoError(t, service.NotifyStatusUpdate("cand-1", "Medical Test"))

	assert.Len(t, notifRepo.created, 3)
	assert.Equal(t, "owner-1", notifRepo.created[0].UserID)
	assert.Equal(t, "selector-1", notifRepo.created[1].UserID)
	assert.Equal(t, "selector-1", notifRepo.created[2].UserID)
}
