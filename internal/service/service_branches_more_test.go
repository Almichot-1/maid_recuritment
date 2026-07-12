package service

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

type candidateDocRepoBehaviorMock struct {
	getByCandidateFn func(candidateID string) ([]*domain.Document, error)
	createFn         func(document *domain.Document) error
}

func (m *candidateDocRepoBehaviorMock) Create(document *domain.Document) error {
	if m.createFn != nil {
		return m.createFn(document)
	}
	return nil
}
func (m *candidateDocRepoBehaviorMock) GetByCandidateID(candidateID string) ([]*domain.Document, error) {
	if m.getByCandidateFn != nil {
		return m.getByCandidateFn(candidateID)
	}
	return nil, nil
}
func (m *candidateDocRepoBehaviorMock) Delete(id string) error { return nil }

type candidateStorageBehaviorMock struct {
	uploadFn func(fileName, contentType string) (string, error)
	deleteFn func(url string) error
	openFn   func(fileURL string) (io.ReadCloser, string, error)
}

func (m *candidateStorageBehaviorMock) Upload(file io.Reader, fileName, contentType string) (string, error) {
	if m.uploadFn != nil {
		return m.uploadFn(fileName, contentType)
	}
	return "https://files/cv.pdf", nil
}
func (m *candidateStorageBehaviorMock) Delete(url string) error {
	if m.deleteFn != nil {
		return m.deleteFn(url)
	}
	return nil
}
func (m *candidateStorageBehaviorMock) Open(fileURL string) (io.ReadCloser, string, error) {
	if m.openFn != nil {
		return m.openFn(fileURL)
	}
	return io.NopCloser(bytes.NewReader(nil)), "", nil
}

func TestCandidateService_MoreBranches(t *testing.T) {
	imgBytes, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9WnR4x4AAAAASUVORK5CYII=")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(imgBytes)
	}))
	defer server.Close()

	repo := &candidateRepoBehaviorMock{}
	docRepo := &candidateDocRepoBehaviorMock{}
	storage := &candidateStorageBehaviorMock{}

	repo.getByID = func(id string) (*domain.Candidate, error) {
		return &domain.Candidate{ID: id, CreatedBy: "owner-1", FullName: "Name", Status: domain.CandidateStatusDraft, Languages: []byte(`["en"]`), Skills: []byte(`["s1"]`)}, nil
	}

	service, err := NewCandidateService(repo, docRepo, storage, NewPDFService(storage), &userRepositoryBehaviorMock{}, &candidatePairShareRepositoryBehaviorMock{}, &pairOverrideRepositoryBehaviorMock{}, nil, nil, nil, nil, nil)
	require.NoError(t, err)

	_, docs, err := service.GetCandidate("cand-1")
	require.NoError(t, err)
	assert.Nil(t, docs)

	_, err = service.PublishCandidate("cand-1", "other", PublishCandidateInput{})
	require.ErrorIs(t, err, ErrForbidden)

	repo.getByID = func(id string) (*domain.Candidate, error) {
		return &domain.Candidate{ID: id, CreatedBy: "owner-1", Status: domain.CandidateStatusAvailable, Languages: []byte("[]"), Skills: []byte("[]")}, nil
	}
	_, err = service.PublishCandidate("cand-1", "owner-1", PublishCandidateInput{})
	require.ErrorIs(t, err, repository.ErrInvalidStatusTransition)

	_, err = service.UploadCandidateDocument("cand-1", "owner-1", UploadCandidateDocumentInput{DocumentType: "photo", FileName: "p.png"})
	require.Error(t, err)

	repo.getByID = func(id string) (*domain.Candidate, error) {
		return &domain.Candidate{ID: id, CreatedBy: "owner-1", FullName: "Name", Status: domain.CandidateStatusDraft, Languages: []byte(`["en"]`), Skills: []byte(`["s1"]`)}, nil
	}
	docRepo.getByCandidateFn = func(candidateID string) ([]*domain.Document, error) {
		return []*domain.Document{
			{DocumentType: domain.Photo, FileURL: server.URL},
			{DocumentType: domain.Passport, FileURL: server.URL},
			{DocumentType: domain.Video, FileURL: "https://example/video.mp4"},
		}, nil
	}
	repo.updateFn = func(candidate *domain.Candidate) error { return nil }
	storage.uploadFn = func(fileName, contentType string) (string, error) {
		return "https://files/cv.pdf", nil
	}

	err = service.GenerateCV("cand-1", "owner-1", "", CandidateCVBranding{})
	require.Error(t, err)
}

func TestAuthService_RegisterInputValidation(t *testing.T) {
	repo := &authUserRepoMock{}
	svc, err := NewAuthService(repo, &config.Config{JWTSecret: "secret-key"})
	require.NoError(t, err)

	_, err = svc.Register("bad", "password123", "A", string(domain.ForeignAgent), "")
	require.ErrorIs(t, err, ErrInvalidCredentials)

	_, err = svc.Register("a@b.com", "short", "A", string(domain.ForeignAgent), "")
	require.ErrorIs(t, err, ErrInvalidCredentials)

	_, err = svc.Register("a@b.com", "password123", "A", "unknown", "")
	require.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestSelectionService_ProcessExpiredSelections_EmptyAndError(t *testing.T) {
	svc := &SelectionService{selectionRepository: &selectionQueryRepoMock{expired: []*domain.Selection{}}, db: nil}
	require.NoError(t, svc.ProcessExpiredSelections())

	svc = &SelectionService{selectionRepository: &selectionQueryRepoMock{expiredErr: errors.New("db")}}
	err := svc.ProcessExpiredSelections()
	require.Error(t, err)
}
