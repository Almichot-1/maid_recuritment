package service

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/domain"
)

type selectionBenchCandidateRepo struct{}

func (m *selectionBenchCandidateRepo) Create(candidate *domain.Candidate) error { return nil }
func (m *selectionBenchCandidateRepo) GetByID(id string) (*domain.Candidate, error) { return nil, nil }
func (m *selectionBenchCandidateRepo) GetByIDs(ids []string) ([]*domain.Candidate, error) {
	return nil, nil
}
func (m *selectionBenchCandidateRepo) GetByIDLean(id string) (*domain.Candidate, error) {
	return nil, nil
}
func (m *selectionBenchCandidateRepo) List(filters domain.CandidateFilters) ([]*domain.Candidate, error) {
	return nil, nil
}
func (m *selectionBenchCandidateRepo) Update(candidate *domain.Candidate) error { return nil }
func (m *selectionBenchCandidateRepo) Delete(id string) error                   { return nil }
func (m *selectionBenchCandidateRepo) Lock(candidateID, lockedBy string, expiresAt time.Time) error {
	return nil
}
func (m *selectionBenchCandidateRepo) Unlock(candidateID string) error { return nil }

type selectionBenchRepo struct {
	db *gorm.DB
}

func (m *selectionBenchRepo) DB() *gorm.DB { return m.db }
func (m *selectionBenchRepo) Create(selection *domain.Selection) error { return nil }
func (m *selectionBenchRepo) GetByID(id string) (*domain.Selection, error) { return nil, nil }
func (m *selectionBenchRepo) GetByCandidateID(candidateID string) (*domain.Selection, error) {
	return nil, nil
}
func (m *selectionBenchRepo) GetByCandidateIDAndPairingID(candidateID, pairingID string) (*domain.Selection, error) {
	return nil, nil
}
func (m *selectionBenchRepo) GetBySelectedBy(userID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *selectionBenchRepo) GetBySelectedByAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *selectionBenchRepo) GetByCandidateOwner(userID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *selectionBenchRepo) GetByCandidateOwnerAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	return nil, nil
}
func (m *selectionBenchRepo) UpdateStatus(id string, status domain.SelectionStatus) error { return nil }
func (m *selectionBenchRepo) GetExpiredSelections() ([]*domain.Selection, error) { return nil, nil }
func (m *selectionBenchRepo) List(filters domain.SelectionFilters) ([]*domain.Selection, error) { return nil, nil }
func (m *selectionBenchRepo) Count(filters domain.SelectionFilters) (int64, error) { return 0, nil }

type selectionBenchNotifier struct {
	mu          sync.Mutex
	foreignByID map[string]bool
}

func (m *selectionBenchNotifier) IsForeignAgent(userID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.foreignByID[userID], nil
}
func (m *selectionBenchNotifier) Send(userID, title, message, notificationType, relatedEntityType, relatedEntityID string) error {
	return nil
}

type selectionBenchCandidate struct {
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

func (selectionBenchCandidate) TableName() string { return "candidates" }

type selectionBenchSelection struct {
	ID                         string         `gorm:"primaryKey;column:id"`
	CandidateID                string         `gorm:"column:candidate_id"`
	PairingID                  string         `gorm:"column:pairing_id"`
	SelectedBy                 string         `gorm:"column:selected_by"`
	Status                     string         `gorm:"column:status"`
	EmployerContractURL        string         `gorm:"column:employer_contract_url"`
	EmployerContractFileName   string         `gorm:"column:employer_contract_file_name"`
	EmployerContractUploadedAt *time.Time     `gorm:"column:employer_contract_uploaded_at"`
	EmployerIDURL              string         `gorm:"column:employer_id_url"`
	EmployerIDFileName         string         `gorm:"column:employer_id_file_name"`
	EmployerIDUploadedAt       *time.Time     `gorm:"column:employer_id_uploaded_at"`
	WarningSentFlags           int            `gorm:"column:warning_sent_flags"`
	ExpiresAt                  time.Time      `gorm:"column:expires_at"`
	CreatedAt                  time.Time      `gorm:"column:created_at"`
	UpdatedAt                  time.Time      `gorm:"column:updated_at"`
}

func (selectionBenchSelection) TableName() string { return "selections" }

func setupBenchSelectionService(b *testing.B) (*SelectionService, *gorm.DB) {
	b.Helper()

	dsn := fmt.Sprintf("file:selection_bench_%s?mode=memory&cache=shared", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(b, err)

	err = db.AutoMigrate(&selectionBenchCandidate{}, &selectionBenchSelection{})
	require.NoError(b, err)

	err = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_selections_one_pending_per_candidate ON selections(candidate_id) WHERE status = 'pending'").Error
	require.NoError(b, err)

	notifier := &selectionBenchNotifier{foreignByID: map[string]bool{}}
	selectionRepo := &selectionBenchRepo{db: db}

	service, err := NewSelectionService(selectionRepo, &selectionBenchCandidateRepo{}, notifier)
	require.NoError(b, err)

	return service, db
}

func BenchmarkSelectionService_GetSelectionsForUser(b *testing.B) {
	svc, _ := setupBenchSelectionService(b)

	userID := "test-user-id"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.GetSelectionsForUser(userID, string(domain.ForeignAgent))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSelectionService_GetSelectionsForEthiopianAgent(b *testing.B) {
	svc, _ := setupBenchSelectionService(b)

	userID := "test-owner-id"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.GetSelectionsForUser(userID, string(domain.EthiopianAgent))
		if err != nil {
			b.Fatal(err)
		}
	}
}
