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

const createPairSharesTable = `CREATE TABLE candidate_pair_shares (
    id TEXT PRIMARY KEY,
    pairing_id TEXT NOT NULL,
    candidate_id TEXT NOT NULL,
    shared_by_user_id TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    cv_pdf_url TEXT,
    shared_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    unshared_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
)`

func setupPairShareTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(createPairSharesTable).Error)
	return db
}

func makePairShare(t *testing.T) *domain.CandidatePairShare {
	t.Helper()
	return &domain.CandidatePairShare{
		ID:             uuid.New().String(),
		PairingID:      uuid.New().String(),
		CandidateID:    uuid.New().String(),
		SharedByUserID: uuid.New().String(),
		IsActive:       true,
	}
}

func TestUpdateShareCVURL(t *testing.T) {
	db := setupPairShareTestDB(t)
	repo := &GormCandidatePairShareRepository{db: db}

	share := makePairShare(t)
	err := repo.Create(share)
	require.NoError(t, err)

	cvURL := "https://cdn.example.com/cvs/abc123.pdf"
	err = repo.UpdateCVURL(share.ID, cvURL)
	require.NoError(t, err)

	got, err := repo.GetActiveByPairingAndCandidate(share.PairingID, share.CandidateID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, cvURL, got.CVPDFURL)
}
