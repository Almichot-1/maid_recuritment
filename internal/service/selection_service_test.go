package service

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/domain"
)

type selectionRepositoryMock struct {
	db *gorm.DB
}

type selectionTestCandidate struct {
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

func (selectionTestCandidate) TableName() string { return "candidates" }

type selectionTestSelection struct {
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

func (selectionTestSelection) TableName() string { return "selections" }

func (m *selectionRepositoryMock) DB() *gorm.DB                             { return m.db }
func (m *selectionRepositoryMock) Create(selection *domain.Selection) error { return nil }
func (m *selectionRepositoryMock) GetByID(id string) (*domain.Selection, error) {
	return nil, nil
}
func (m *selectionRepositoryMock) GetByCandidateID(candidateID string) (*domain.Selection, error) {
	return nil, nil
}
func (m *selectionRepositoryMock) GetByCandidateIDAndPairingID(candidateID, pairingID string) (*domain.Selection, error) {
	return nil, nil
}
func (m *selectionRepositoryMock) GetBySelectedBy(userID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *selectionRepositoryMock) GetBySelectedByAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *selectionRepositoryMock) GetByCandidateOwner(userID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *selectionRepositoryMock) GetByCandidateOwnerAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *selectionRepositoryMock) UpdateStatus(id string, status domain.SelectionStatus) error {
	return nil
}
func (m *selectionRepositoryMock) GetExpiredSelections() ([]*domain.Selection, error) {
	return nil, nil
}

type candidateRepositoryMock struct{}

func (m *candidateRepositoryMock) Create(candidate *domain.Candidate) error     { return nil }
func (m *candidateRepositoryMock) GetByID(id string) (*domain.Candidate, error) { return nil, nil }
func (m *candidateRepositoryMock) List(filters domain.CandidateFilters) ([]*domain.Candidate, error) {
	return nil, nil
}
func (m *candidateRepositoryMock) Update(candidate *domain.Candidate) error { return nil }
func (m *candidateRepositoryMock) Delete(id string) error                   { return nil }
func (m *candidateRepositoryMock) Lock(candidateID, lockedBy string, expiresAt time.Time) error {
	return nil
}
func (m *candidateRepositoryMock) Unlock(candidateID string) error { return nil }

type notificationSenderMock struct {
	mu           sync.Mutex
	foreignByID  map[string]bool
	sendCalls    int
	sentToUserID []string
	payloads     []notificationSendPayload
}

type notificationSendPayload struct {
	userID            string
	title             string
	notificationType  string
	relatedEntityType string
	relatedEntityID   string
}

func (m *notificationSenderMock) IsForeignAgent(userID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.foreignByID[userID], nil
}

func (m *notificationSenderMock) Send(userID, title, message, notificationType, relatedEntityType, relatedEntityID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendCalls++
	m.sentToUserID = append(m.sentToUserID, userID)
	m.payloads = append(m.payloads, notificationSendPayload{
		userID:            userID,
		title:             title,
		notificationType:  notificationType,
		relatedEntityType: relatedEntityType,
		relatedEntityID:   relatedEntityID,
	})
	return nil
}

func setupSelectionService(t *testing.T) (*SelectionService, *gorm.DB, *notificationSenderMock) {
	t.Helper()

	dsn := fmt.Sprintf("file:selection_test_%s?mode=memory&cache=shared", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&selectionTestCandidate{}, &selectionTestSelection{})
	require.NoError(t, err)

	err = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_selections_one_pending_per_candidate ON selections(candidate_id) WHERE status = 'pending'").Error
	require.NoError(t, err)

	notifier := &notificationSenderMock{foreignByID: map[string]bool{}}
	selectionRepo := &selectionRepositoryMock{db: db}
	candidateRepo := &candidateRepositoryMock{}

	service, err := NewSelectionService(selectionRepo, candidateRepo, notifier)
	require.NoError(t, err)

	return service, db, notifier
}

func seedCandidate(t *testing.T, db *gorm.DB, id, createdBy string, status domain.CandidateStatus) {
	t.Helper()
	err := db.Create(&selectionTestCandidate{
		ID:        id,
		CreatedBy: createdBy,
		FullName:  "Candidate",
		Status:    string(status),
		Languages: []byte("[]"),
		Skills:    []byte("[]"),
	}).Error
	require.NoError(t, err)
}

func TestSelectionService_SelectCandidate_SuccessLocksCandidate(t *testing.T) {
	service, db, notifier := setupSelectionService(t)

	seedCandidate(t, db, "cand-1", "ethiopian-owner", domain.CandidateStatusAvailable)
	notifier.foreignByID["foreign-1"] = true

	selection, err := service.SelectCandidate("cand-1", "foreign-1")
	require.NoError(t, err)
	require.NotNil(t, selection)
	assert.Equal(t, domain.SelectionPending, selection.Status)
	assert.Equal(t, "cand-1", selection.CandidateID)
	assert.Equal(t, "foreign-1", selection.SelectedBy)

	var candidate domain.Candidate
	err = db.Where("id = ?", "cand-1").First(&candidate).Error
	require.NoError(t, err)
	assert.Equal(t, domain.CandidateStatusLocked, candidate.Status)
	require.NotNil(t, candidate.LockedBy)
	assert.Equal(t, "foreign-1", *candidate.LockedBy)
	require.NotNil(t, candidate.LockExpiresAt)

	assert.Equal(t, 2, notifier.sendCalls)
	assert.ElementsMatch(t, []string{"ethiopian-owner", "foreign-1"}, notifier.sentToUserID)
	require.Len(t, notifier.payloads, 2)
	assert.Equal(t, "selection", notifier.payloads[0].relatedEntityType)
	assert.Equal(t, selection.ID, notifier.payloads[0].relatedEntityID)
	assert.Equal(t, "selection", notifier.payloads[1].relatedEntityType)
	assert.Equal(t, selection.ID, notifier.payloads[1].relatedEntityID)
}

func TestSelectionService_SelectCandidate_CandidateStateRules(t *testing.T) {
	tests := []struct {
		name          string
		candidateID   string
		candidateStat domain.CandidateStatus
		seedPending   bool
		expectedErr   error
	}{
		{
			name:          "cannot select already locked candidate",
			candidateID:   "cand-locked",
			candidateStat: domain.CandidateStatusLocked,
			expectedErr:   ErrCandidateNotAvailable,
		},
		{
			name:          "cannot select non available candidate",
			candidateID:   "cand-progress",
			candidateStat: domain.CandidateStatusInProgress,
			expectedErr:   ErrCandidateNotAvailable,
		},
		{
			name:          "cannot select candidate with active pending selection",
			candidateID:   "cand-pending",
			candidateStat: domain.CandidateStatusAvailable,
			seedPending:   true,
			expectedErr:   ErrAlreadySelected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, db, notifier := setupSelectionService(t)
			notifier.foreignByID["foreign-1"] = true

			seedCandidate(t, db, tt.candidateID, "owner-1", tt.candidateStat)

			if tt.seedPending {
				err := db.Create(&selectionTestSelection{
					ID:          "sel-existing",
					CandidateID: tt.candidateID,
					SelectedBy:  "foreign-2",
					Status:      string(domain.SelectionPending),
					ExpiresAt:   time.Now().UTC().Add(1 * time.Hour),
				}).Error
				require.NoError(t, err)
			}

			selection, err := service.SelectCandidate(tt.candidateID, "foreign-1")
			require.ErrorIs(t, err, tt.expectedErr)
			assert.Nil(t, selection)
		})
	}
}

func TestSelectionService_SelectCandidate_RoleValidation(t *testing.T) {
	tests := []struct {
		name         string
		selectedBy   string
		isForeign    bool
		expectErr    error
		expectCreate bool
	}{
		{
			name:         "foreign agent can select",
			selectedBy:   "foreign-1",
			isForeign:    true,
			expectCreate: true,
		},
		{
			name:       "ethiopian agent cannot select",
			selectedBy: "ethiopian-1",
			isForeign:  false,
			expectErr:  ErrNotForeignAgent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, db, notifier := setupSelectionService(t)
			seedCandidate(t, db, "cand-1", "owner-1", domain.CandidateStatusAvailable)
			notifier.foreignByID[tt.selectedBy] = tt.isForeign

			selection, err := service.SelectCandidate("cand-1", tt.selectedBy)
			if tt.expectErr != nil {
				require.ErrorIs(t, err, tt.expectErr)
				assert.Nil(t, selection)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, selection)
		})
	}
}

func TestSelectionService_SelectCandidate_ConcurrentAttempts(t *testing.T) {
	service, db, notifier := setupSelectionService(t)
	seedCandidate(t, db, "cand-race", "owner-race", domain.CandidateStatusAvailable)
	notifier.foreignByID["foreign-a"] = true
	notifier.foreignByID["foreign-b"] = true

	users := []string{"foreign-a", "foreign-b"}
	type result struct {
		selection *domain.Selection
		err       error
	}
	results := make([]result, len(users))

	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := range users {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			selection, err := service.SelectCandidate("cand-race", users[idx])
			results[idx] = result{selection: selection, err: err}
		}(i)
	}

	close(start)
	wg.Wait()

	successCount := 0
	errorCount := 0
	for _, res := range results {
		if res.err == nil {
			successCount++
			continue
		}
		errorCount++
		isSQLiteLockContention := strings.Contains(res.err.Error(), "database table is locked") || strings.Contains(res.err.Error(), "deadlocked")
		assert.True(t,
			res.err == ErrAlreadySelected || res.err == ErrCandidateNotAvailable || isSQLiteLockContention,
			fmt.Sprintf("unexpected error: %v", res.err),
		)
	}

	assert.Equal(t, 1, successCount)
	assert.Equal(t, 1, errorCount)

	var selections []domain.Selection
	err := db.Where("candidate_id = ? AND status = ?", "cand-race", domain.SelectionPending).Find(&selections).Error
	require.NoError(t, err)
	assert.Len(t, selections, 1)
}
