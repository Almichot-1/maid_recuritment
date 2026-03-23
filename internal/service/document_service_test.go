package service

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"maid-recruitment-tracking/internal/domain"
)

type documentRepoMock struct {
	createFn func(document *domain.Document) error
}

func (m *documentRepoMock) Create(document *domain.Document) error {
	if m.createFn != nil {
		return m.createFn(document)
	}
	return nil
}

func (m *documentRepoMock) GetByCandidateID(candidateID string) ([]*domain.Document, error) {
	return nil, nil
}

func (m *documentRepoMock) Delete(id string) error { return nil }

type documentStorageMock struct {
	uploadFn func(fileName, contentType string) (string, error)
	deleteFn func(url string) error
}

func (m *documentStorageMock) Upload(file io.Reader, fileName, contentType string) (string, error) {
	if m.uploadFn != nil {
		return m.uploadFn(fileName, contentType)
	}
	return "https://example.com/file", nil
}

func (m *documentStorageMock) Delete(url string) error {
	if m.deleteFn != nil {
		return m.deleteFn(url)
	}
	return nil
}

func TestDocumentService_UploadDocument_Success(t *testing.T) {
	storage := &documentStorageMock{uploadFn: func(fileName, contentType string) (string, error) {
		assert.Equal(t, "passport.pdf", fileName)
		assert.Equal(t, "application/pdf", contentType)
		return "https://example.com/passport.pdf", nil
	}}
	repo := &documentRepoMock{}

	svc, err := NewDocumentService(repo, storage)
	require.NoError(t, err)

	doc, err := svc.UploadDocument("cand-1", string(domain.Passport), bytes.NewBufferString("file"), "passport.pdf", 100)
	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.Equal(t, domain.Passport, doc.DocumentType)
	assert.Equal(t, "https://example.com/passport.pdf", doc.FileURL)
}

func TestDocumentService_UploadDocument_ValidationErrors(t *testing.T) {
	svc, err := NewDocumentService(&documentRepoMock{}, &documentStorageMock{})
	require.NoError(t, err)

	_, err = svc.UploadDocument("", string(domain.Passport), bytes.NewBufferString("x"), "a.pdf", 1)
	require.Error(t, err)

	_, err = svc.UploadDocument("cand-1", string(domain.Passport), nil, "a.pdf", 1)
	require.Error(t, err)

	_, err = svc.UploadDocument("cand-1", string(domain.Passport), bytes.NewBufferString("x"), "", 1)
	require.Error(t, err)

	_, err = svc.UploadDocument("cand-1", string(domain.Passport), bytes.NewBufferString("x"), "a.pdf", maxDocumentFileSizeBytes+1)
	require.ErrorIs(t, err, ErrFileTooLarge)
}

func TestDocumentService_UploadDocument_SaveFailureDeletesUploadedFile(t *testing.T) {
	deleted := false
	storage := &documentStorageMock{
		uploadFn: func(fileName, contentType string) (string, error) {
			return "https://example.com/p.pdf", nil
		},
		deleteFn: func(url string) error {
			deleted = true
			assert.Equal(t, "https://example.com/p.pdf", url)
			return nil
		},
	}
	repo := &documentRepoMock{createFn: func(document *domain.Document) error {
		return errors.New("db fail")
	}}

	svc, err := NewDocumentService(repo, storage)
	require.NoError(t, err)

	_, err = svc.UploadDocument("cand-1", string(domain.Passport), bytes.NewBufferString("x"), "passport.pdf", 100)
	require.Error(t, err)
	assert.True(t, deleted)
}

func TestDocumentHelpers(t *testing.T) {
	dt, err := parseDocumentType(string(domain.Photo))
	require.NoError(t, err)
	assert.Equal(t, domain.Photo, dt)

	_, err = parseDocumentType("invalid")
	require.ErrorIs(t, err, ErrInvalidDocumentType)

	ct, err := detectContentTypeFromFileName("img.jpeg")
	require.NoError(t, err)
	assert.Equal(t, "image/jpeg", ct)

	_, err = detectContentTypeFromFileName("file.xyz")
	require.ErrorIs(t, err, ErrInvalidFileType)

	require.NoError(t, validateDocumentTypeContentType(domain.Passport, "application/pdf"))
	require.NoError(t, validateDocumentTypeContentType(domain.Passport, "image/png"))
	require.NoError(t, validateDocumentTypeContentType(domain.Passport, "image/jpeg"))
	require.NoError(t, validateDocumentTypeContentType(domain.Photo, "image/png"))
	require.NoError(t, validateDocumentTypeContentType(domain.Video, "video/mp4"))

	require.ErrorIs(t, validateDocumentTypeContentType(domain.Passport, "video/mp4"), ErrInvalidFileType)
	require.ErrorIs(t, validateDocumentTypeContentType(domain.DocumentType("x"), "image/png"), ErrInvalidDocumentType)
}

func TestNewDocumentServiceValidation(t *testing.T) {
	_, err := NewDocumentService(nil, &documentStorageMock{})
	require.Error(t, err)
	_, err = NewDocumentService(&documentRepoMock{}, nil)
	require.Error(t, err)
}
