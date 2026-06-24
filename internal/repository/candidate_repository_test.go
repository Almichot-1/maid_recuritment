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

const createCandidatesTable = `CREATE TABLE candidates (
	id TEXT PRIMARY KEY,
	created_by TEXT NOT NULL,
	full_name TEXT NOT NULL,
	nationality TEXT,
	date_of_birth TEXT,
	age INTEGER,
	place_of_birth TEXT,
	religion TEXT,
	marital_status TEXT,
	children_count INTEGER,
	education_level TEXT,
	experience_years INTEGER,
	country_of_experience TEXT,
	country_applied TEXT,
	salary_offered TEXT,
	languages TEXT NOT NULL DEFAULT '[]',
	skills TEXT NOT NULL DEFAULT '[]',
	status TEXT NOT NULL DEFAULT 'draft',
	locked_by TEXT,
	locked_at DATETIME,
	lock_expires_at DATETIME,
	cv_pdf_url TEXT,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at DATETIME
)`

const createDocumentsTable = `CREATE TABLE documents (
	id TEXT PRIMARY KEY,
	candidate_id TEXT NOT NULL,
	document_type TEXT NOT NULL,
	file_url TEXT NOT NULL,
	file_name TEXT,
	file_size INTEGER,
	uploaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
)`

func setupCandidateTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(createCandidatesTable).Error)
	require.NoError(t, db.Exec(createDocumentsTable).Error)
	return db
}

func makeCandidate(t *testing.T) *domain.Candidate {
	t.Helper()
	return &domain.Candidate{
		ID:        uuid.New().String(),
		FullName:  "Test Maid",
		Status:    domain.CandidateStatusDraft,
		CreatedBy: uuid.New().String(),
		Languages: json.RawMessage("[]"),
		Skills:    json.RawMessage("[]"),
	}
}

func TestCandidateRepository_CreateAndGetByID(t *testing.T) {
	db := setupCandidateTestDB(t)
	repo := &GormCandidateRepository{db: db}

	candidate := makeCandidate(t)
	candidate.FullName = "Test Maid"
	candidate.Nationality = "Ethiopian"

	err := repo.Create(candidate)
	require.NoError(t, err)

	got, err := repo.GetByID(candidate.ID)
	require.NoError(t, err)
	assert.Equal(t, candidate.FullName, got.FullName)
	assert.Equal(t, candidate.Nationality, got.Nationality)
	assert.Equal(t, candidate.Status, got.Status)
}

func TestCandidateRepository_List(t *testing.T) {
	db := setupCandidateTestDB(t)
	repo := &GormCandidateRepository{db: db}

	userID := uuid.New().String()
	for i := 0; i < 5; i++ {
		candidate := makeCandidate(t)
		candidate.CreatedBy = userID
		err := repo.Create(candidate)
		require.NoError(t, err)
	}

	filters := domain.CandidateFilters{
		Page:     1,
		PageSize: 20,
	}

	candidates, err := repo.List(filters)
	require.NoError(t, err)
	assert.Len(t, candidates, 5)
}

func TestCandidateRepository_GetByIDLean(t *testing.T) {
	db := setupCandidateTestDB(t)
	repo := &GormCandidateRepository{db: db}

	candidate := makeCandidate(t)

	err := repo.Create(candidate)
	require.NoError(t, err)

	lean, err := repo.GetByIDLean(candidate.ID)
	require.NoError(t, err)
	assert.Equal(t, candidate.ID, lean.ID)
	assert.Equal(t, candidate.Status, lean.Status)
}

func TestCandidateRepository_GetByIDs(t *testing.T) {
	db := setupCandidateTestDB(t)
	repo := &GormCandidateRepository{db: db}

	ids := make([]string, 3)
	for i := 0; i < 3; i++ {
		id := uuid.New().String()
		ids[i] = id
		candidate := makeCandidate(t)
		candidate.ID = id
		err := repo.Create(candidate)
		require.NoError(t, err)
	}

	candidates, err := repo.GetByIDs(ids)
	require.NoError(t, err)
	assert.Len(t, candidates, 3)
}

func TestCandidateRepository_NotFound(t *testing.T) {
	db := setupCandidateTestDB(t)
	repo := &GormCandidateRepository{db: db}

	_, err := repo.GetByID("nonexistent")
	assert.ErrorIs(t, err, ErrCandidateNotFound)
}

func TestCandidateRepository_Update(t *testing.T) {
	db := setupCandidateTestDB(t)
	repo := &GormCandidateRepository{db: db}

	candidate := makeCandidate(t)
	err := repo.Create(candidate)
	require.NoError(t, err)

	candidate.FullName = "Updated Name"
	candidate.Status = domain.CandidateStatusAvailable
	err = repo.Update(candidate)
	require.NoError(t, err)

	got, err := repo.GetByID(candidate.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", got.FullName)
	assert.Equal(t, domain.CandidateStatusAvailable, got.Status)
}

func TestCandidateRepository_Delete(t *testing.T) {
	db := setupCandidateTestDB(t)
	repo := &GormCandidateRepository{db: db}

	candidate := makeCandidate(t)
	err := repo.Create(candidate)
	require.NoError(t, err)

	err = repo.Delete(candidate.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(candidate.ID)
	assert.ErrorIs(t, err, ErrCandidateNotFound)
}

func TestCandidateRepository_Lock(t *testing.T) {
	db := setupCandidateTestDB(t)
	repo := &GormCandidateRepository{db: db}

	candidate := makeCandidate(t)
	candidate.Status = domain.CandidateStatusAvailable
	err := repo.Create(candidate)
	require.NoError(t, err)

	lockedBy := uuid.New().String()
	err = repo.Lock(candidate.ID, lockedBy, time.Now().Add(1*time.Hour))
	require.NoError(t, err)

	got, err := repo.GetByID(candidate.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.CandidateStatusLocked, got.Status)
	assert.NotNil(t, got.LockedBy)
	assert.Equal(t, lockedBy, *got.LockedBy)
}

func TestCandidateRepository_Unlock(t *testing.T) {
	db := setupCandidateTestDB(t)
	repo := &GormCandidateRepository{db: db}

	candidate := makeCandidate(t)
	candidate.Status = domain.CandidateStatusAvailable
	err := repo.Create(candidate)
	require.NoError(t, err)

	lockedBy := uuid.New().String()
	err = repo.Lock(candidate.ID, lockedBy, time.Now().Add(1*time.Hour))
	require.NoError(t, err)

	err = repo.Unlock(candidate.ID)
	require.NoError(t, err)

	got, err := repo.GetByID(candidate.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.CandidateStatusAvailable, got.Status)
}

func TestCandidateRepository_ListWithCreatedByFilter(t *testing.T) {
	db := setupCandidateTestDB(t)
	repo := &GormCandidateRepository{db: db}

	userID := uuid.New().String()
	otherUserID := uuid.New().String()

	for i := 0; i < 3; i++ {
		candidate := makeCandidate(t)
		candidate.CreatedBy = userID
		err := repo.Create(candidate)
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		candidate := makeCandidate(t)
		candidate.CreatedBy = otherUserID
		err := repo.Create(candidate)
		require.NoError(t, err)
	}

	filters := domain.CandidateFilters{
		CreatedBy: userID,
		Page:      1,
		PageSize:  20,
	}

	candidates, err := repo.List(filters)
	require.NoError(t, err)
	assert.Len(t, candidates, 3)
}
