package repository

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/domain"
)

const createPairOverridesTable = `CREATE TABLE candidate_pair_overrides (
    id TEXT PRIMARY KEY,
    pairing_id TEXT NOT NULL,
    candidate_id TEXT NOT NULL,
    country_applied TEXT,
    salary_offered TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (pairing_id, candidate_id)
)`

func setupPairOverrideTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(createPairOverridesTable).Error)
	return db
}

func makePairOverride(t *testing.T) *domain.CandidatePairOverride {
	t.Helper()
	return &domain.CandidatePairOverride{
		ID:             uuid.New().String(),
		PairingID:      uuid.New().String(),
		CandidateID:    uuid.New().String(),
		CountryApplied: "Kuwait",
		SalaryOffered:  "900 KWD",
	}
}

func TestBulkUpsert_PartialBatch(t *testing.T) {
	db := setupPairOverrideTestDB(t)
	repo := &GormCandidatePairOverrideRepository{db: db}

	const count = 250
	candidateID := uuid.New().String()
	overrides := make([]*domain.CandidatePairOverride, 0, count)
	for range count {
		overrides = append(overrides, &domain.CandidatePairOverride{
			ID:          uuid.New().String(),
			PairingID:   uuid.New().String(),
			CandidateID: candidateID,
		})
	}

	err := repo.BulkUpsert(overrides)
	require.NoError(t, err)

	result, err := repo.ListByCandidateID(candidateID)
	require.NoError(t, err)
	assert.Len(t, result, count)
}

func TestBulkUpsert_UpdatesExisting(t *testing.T) {
	db := setupPairOverrideTestDB(t)
	repo := &GormCandidatePairOverrideRepository{db: db}

	override := makePairOverride(t)
	override.CountryApplied = "Qatar"
	override.SalaryOffered = "1200 QAR"

	err := repo.Upsert(override)
	require.NoError(t, err)

	override.CountryApplied = "UAE"
	override.SalaryOffered = "2000 AED"

	err = repo.BulkUpsert([]*domain.CandidatePairOverride{override})
	require.NoError(t, err)

	got, err := repo.GetByPairingAndCandidate(override.PairingID, override.CandidateID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "UAE", got.CountryApplied)
	assert.Equal(t, "2000 AED", got.SalaryOffered)

	var rows []map[string]any
	err = db.Table("candidate_pair_overrides").
		Where("pairing_id = ? AND candidate_id = ?", override.PairingID, override.CandidateID).
		Find(&rows).Error
	require.NoError(t, err)
	assert.Len(t, rows, 1)
}
