package repository

import (
	"encoding/json"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/domain"
)

func setupDocumentTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(createCandidatesTable).Error)
	require.NoError(t, db.Exec(createDocumentsTable).Error)
	return db
}

func TestDocumentRepository_CreateAndGetByCandidateID(t *testing.T) {
	db := setupDocumentTestDB(t)
	docRepo := &GormDocumentRepository{db: db}

	candidate := &domain.Candidate{
		ID:        uuid.New().String(),
		FullName:  "Doc Test",
		Status:    domain.CandidateStatusDraft,
		CreatedBy: uuid.New().String(),
		Languages: json.RawMessage("[]"),
		Skills:    json.RawMessage("[]"),
	}
	err := db.Create(candidate).Error
	require.NoError(t, err)

	doc := &domain.Document{
		ID:           uuid.New().String(),
		CandidateID:  candidate.ID,
		DocumentType: domain.Passport,
		FileURL:      "https://example.com/doc.pdf",
		FileName:     "passport.pdf",
		FileSize:     1024,
	}

	err = docRepo.Create(doc)
	require.NoError(t, err)

	docs, err := docRepo.GetByCandidateID(candidate.ID)
	require.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, doc.ID, docs[0].ID)
	assert.Equal(t, doc.DocumentType, docs[0].DocumentType)
}

func TestDocumentRepository_GetByCandidateID_Multiple(t *testing.T) {
	db := setupDocumentTestDB(t)
	docRepo := &GormDocumentRepository{db: db}

	candidate := &domain.Candidate{
		ID:        uuid.New().String(),
		FullName:  "Multi Doc",
		Status:    domain.CandidateStatusDraft,
		CreatedBy: uuid.New().String(),
		Languages: json.RawMessage("[]"),
		Skills:    json.RawMessage("[]"),
	}
	err := db.Create(candidate).Error
	require.NoError(t, err)

	docTypes := []domain.DocumentType{domain.Passport, domain.Photo, domain.MedicalDocument}
	for _, dt := range docTypes {
		doc := &domain.Document{
			ID:           uuid.New().String(),
			CandidateID:  candidate.ID,
			DocumentType: dt,
			FileURL:      "https://example.com/" + string(dt),
			FileName:     string(dt) + ".pdf",
		}
		err := docRepo.Create(doc)
		require.NoError(t, err)
	}

	docs, err := docRepo.GetByCandidateID(candidate.ID)
	require.NoError(t, err)
	assert.Len(t, docs, 3)
}

func TestDocumentRepository_Delete(t *testing.T) {
	db := setupDocumentTestDB(t)
	docRepo := &GormDocumentRepository{db: db}

	candidate := &domain.Candidate{
		ID:        uuid.New().String(),
		FullName:  "Delete Doc",
		Status:    domain.CandidateStatusDraft,
		CreatedBy: uuid.New().String(),
		Languages: json.RawMessage("[]"),
		Skills:    json.RawMessage("[]"),
	}
	err := db.Create(candidate).Error
	require.NoError(t, err)

	doc := &domain.Document{
		ID:           uuid.New().String(),
		CandidateID:  candidate.ID,
		DocumentType: domain.Passport,
		FileURL:      "https://example.com/doc.pdf",
	}
	err = docRepo.Create(doc)
	require.NoError(t, err)

	err = docRepo.Delete(doc.ID)
	require.NoError(t, err)

	docs, err := docRepo.GetByCandidateID(candidate.ID)
	require.NoError(t, err)
	assert.Len(t, docs, 0)
}

func TestDocumentRepository_DeleteNotFound(t *testing.T) {
	db := setupDocumentTestDB(t)
	docRepo := &GormDocumentRepository{db: db}

	err := docRepo.Delete("nonexistent")
	assert.ErrorIs(t, err, ErrDocumentNotFound)
}
