package repository

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/domain"
)

const createSelectionsTable = `CREATE TABLE selections (
	id TEXT PRIMARY KEY,
	candidate_id TEXT NOT NULL,
	pairing_id TEXT NOT NULL,
	selected_by TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'pending',
	employer_contract_url TEXT,
	employer_contract_file_name TEXT,
	employer_contract_uploaded_at DATETIME,
	employer_id_url TEXT,
	employer_id_file_name TEXT,
	employer_id_uploaded_at DATETIME,
	warning_sent_flags INTEGER DEFAULT 0,
	expires_at DATETIME NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
)`

func setupSelectionTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(createCandidatesTable).Error)
	require.NoError(t, db.Exec(createDocumentsTable).Error)
	require.NoError(t, db.Exec(createSelectionsTable).Error)
	return db
}

func createTestCandidate(db *gorm.DB, status domain.CandidateStatus) *domain.Candidate {
	candidate := &domain.Candidate{
		ID:        uuid.New().String(),
		FullName:  "Test Candidate",
		Status:    status,
		CreatedBy: uuid.New().String(),
		Languages: json.RawMessage("[]"),
		Skills:    json.RawMessage("[]"),
	}
	err := db.Create(candidate).Error
	if err != nil {
		panic(err)
	}
	return candidate
}

func TestSelectionRepository_CreateAndGetByID(t *testing.T) {
	db := setupSelectionTestDB(t)
	selectionRepo := &GormSelectionRepository{db: db}

	candidateID := uuid.New().String()
	candidate := &domain.Candidate{
		ID:        candidateID,
		FullName:  "Selection Test",
		Status:    domain.CandidateStatusAvailable,
		CreatedBy: uuid.New().String(),
		Languages: json.RawMessage("[]"),
		Skills:    json.RawMessage("[]"),
	}
	err := db.Create(candidate).Error
	require.NoError(t, err)

	selection := &domain.Selection{
		ID:          uuid.New().String(),
		CandidateID: candidateID,
		PairingID:   uuid.New().String(),
		SelectedBy:  uuid.New().String(),
		Status:      domain.SelectionPending,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	err = selectionRepo.Create(selection)
	require.NoError(t, err)

	got, err := selectionRepo.GetByID(selection.ID)
	require.NoError(t, err)
	assert.Equal(t, selection.ID, got.ID)
	assert.Equal(t, selection.Status, got.Status)
	assert.Equal(t, selection.CandidateID, got.CandidateID)
}

func TestSelectionRepository_GetBySelectedBy(t *testing.T) {
	db := setupSelectionTestDB(t)
	selectionRepo := &GormSelectionRepository{db: db}

	userID := uuid.New().String()
	for i := 0; i < 3; i++ {
		candidate := createTestCandidate(db, domain.CandidateStatusAvailable)

		selection := &domain.Selection{
			ID:          uuid.New().String(),
			CandidateID: candidate.ID,
			PairingID:   uuid.New().String(),
			SelectedBy:  userID,
			Status:      domain.SelectionPending,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
		}
		err := selectionRepo.Create(selection)
		require.NoError(t, err)
	}

	selections, err := selectionRepo.GetBySelectedBy(userID)
	require.NoError(t, err)
	assert.Len(t, selections, 3)
}

func TestSelectionRepository_GetBySelectedByBatch(t *testing.T) {
	db := setupSelectionTestDB(t)
	selectionRepo := &GormSelectionRepository{db: db}

	userID := uuid.New().String()
	for i := 0; i < 3; i++ {
		candidate := createTestCandidate(db, domain.CandidateStatusAvailable)

		selection := &domain.Selection{
			ID:          uuid.New().String(),
			CandidateID: candidate.ID,
			PairingID:   uuid.New().String(),
			SelectedBy:  userID,
			Status:      domain.SelectionPending,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
		}
		err := selectionRepo.Create(selection)
		require.NoError(t, err)
	}

	selections, err := selectionRepo.GetBySelectedByBatch(userID)
	require.NoError(t, err)
	assert.Len(t, selections, 3)
	for _, s := range selections {
		assert.NotNil(t, s.Candidate)
		assert.Equal(t, "Test Candidate", s.Candidate.FullName)
	}
}

func TestSelectionRepository_GetByCandidateID(t *testing.T) {
	db := setupSelectionTestDB(t)
	selectionRepo := &GormSelectionRepository{db: db}

	candidate := createTestCandidate(db, domain.CandidateStatusAvailable)

	selection := &domain.Selection{
		ID:          uuid.New().String(),
		CandidateID: candidate.ID,
		PairingID:   uuid.New().String(),
		SelectedBy:  uuid.New().String(),
		Status:      domain.SelectionPending,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}
	err := selectionRepo.Create(selection)
	require.NoError(t, err)

	got, err := selectionRepo.GetByCandidateID(candidate.ID)
	require.NoError(t, err)
	assert.Equal(t, selection.ID, got.ID)
}

func TestSelectionRepository_GetByCandidateIDAndPairingID(t *testing.T) {
	db := setupSelectionTestDB(t)
	selectionRepo := &GormSelectionRepository{db: db}

	candidate := createTestCandidate(db, domain.CandidateStatusAvailable)
	pairingID := uuid.New().String()

	selection := &domain.Selection{
		ID:          uuid.New().String(),
		CandidateID: candidate.ID,
		PairingID:   pairingID,
		SelectedBy:  uuid.New().String(),
		Status:      domain.SelectionPending,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}
	err := selectionRepo.Create(selection)
	require.NoError(t, err)

	got, err := selectionRepo.GetByCandidateIDAndPairingID(candidate.ID, pairingID)
	require.NoError(t, err)
	assert.Equal(t, selection.ID, got.ID)
}

func TestSelectionRepository_UpdateStatus(t *testing.T) {
	db := setupSelectionTestDB(t)
	selectionRepo := &GormSelectionRepository{db: db}

	candidate := createTestCandidate(db, domain.CandidateStatusAvailable)

	selection := &domain.Selection{
		ID:          uuid.New().String(),
		CandidateID: candidate.ID,
		PairingID:   uuid.New().String(),
		SelectedBy:  uuid.New().String(),
		Status:      domain.SelectionPending,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}
	err := selectionRepo.Create(selection)
	require.NoError(t, err)

	err = selectionRepo.UpdateStatus(selection.ID, domain.SelectionApproved)
	require.NoError(t, err)

	got, err := selectionRepo.GetByID(selection.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.SelectionApproved, got.Status)
}

func TestSelectionRepository_NotFound(t *testing.T) {
	db := setupSelectionTestDB(t)
	selectionRepo := &GormSelectionRepository{db: db}

	_, err := selectionRepo.GetByID("nonexistent")
	assert.ErrorIs(t, err, ErrSelectionNotFound)
}

func TestSelectionRepository_InvalidStatus(t *testing.T) {
	db := setupSelectionTestDB(t)
	selectionRepo := &GormSelectionRepository{db: db}

	err := selectionRepo.UpdateStatus("some-id", "invalid")
	assert.ErrorIs(t, err, ErrInvalidSelectionStatus)
}
