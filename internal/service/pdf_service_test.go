package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"testing"

	"github.com/jung-kurt/gofpdf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"maid-recruitment-tracking/internal/domain"
)

type pdfTestStorageMock struct{}

func (m *pdfTestStorageMock) Upload(file io.Reader, fileName, contentType string) (string, error) {
	return "https://files/test.pdf", nil
}
func (m *pdfTestStorageMock) Delete(url string) error {
	return nil
}
func (m *pdfTestStorageMock) Open(fileURL string) (io.ReadCloser, string, error) {
	switch fileURL {
	case "s3://bad-status":
		return nil, "", fmt.Errorf("not found")
	case "s3://bad-type":
		return io.NopCloser(bytes.NewReader([]byte("x"))), "text/plain", nil
	default:
		imgBytes, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9WnR4x4AAAAASUVORK5CYII=")
		return io.NopCloser(bytes.NewReader(imgBytes)), "image/png", nil
	}
}

func TestPDFService_NewAndGenerateValidation(t *testing.T) {
	svc := NewPDFService(&pdfTestStorageMock{})
	require.NotNil(t, svc)

	_, err := svc.GenerateCandidateCV(nil, nil, CandidateCVBranding{}, nil)
	require.Error(t, err)

	_, err = svc.GenerateCandidateCV(&domain.Candidate{FullName: "A"}, []*domain.Document{}, CandidateCVBranding{}, nil)
	require.ErrorIs(t, err, ErrMissingRequiredDocuments)
}

func TestPDFHelpers_DocumentPickingAndJSON(t *testing.T) {
	photo, passport, video := pickCandidateDocuments([]*domain.Document{
		{DocumentType: domain.Photo, FileURL: "p"},
		{DocumentType: domain.Passport, FileURL: "pp"},
		{DocumentType: domain.Video, FileURL: "v"},
	})
	require.NotNil(t, photo)
	require.NotNil(t, passport)
	require.NotNil(t, video)

	assert.Equal(t, []string{}, parseJSONList(nil))
	assert.Equal(t, []string{}, parseJSONList([]byte("invalid")))
	assert.Equal(t, []string{"en"}, parseJSONList([]byte(`["en"]`)))

	assert.Equal(t, "N/A", formatIntPointer(nil))
	v := 12
	assert.Equal(t, "12", formatIntPointer(&v))
}

func TestPDFHelpers_FetchImageAndAddImage(t *testing.T) {
	imgBytes, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9WnR4x4AAAAASUVORK5CYII=")

	mock := &pdfTestStorageMock{}
	svc := NewPDFService(mock)

	body, contentType, err := svc.fetchImage("s3://ok")
	require.NoError(t, err)
	assert.Equal(t, "image/png", contentType)
	assert.NotEmpty(t, body)

	_, _, err = svc.fetchImage("s3://bad-status")
	require.Error(t, err)

	_, _, err = svc.fetchImage("s3://bad-type")
	require.ErrorIs(t, err, ErrMissingRequiredDocuments)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	require.NoError(t, addImageFromBytes(pdf, imgBytes, "image/png", "img1", 10, 10, 10, 10))
}

func TestPDFService_DrawHelpers(t *testing.T) {
	svc := NewPDFService(&pdfTestStorageMock{})
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	svc.drawBrandingHeader(pdf, CandidateCVBranding{CompanyName: "Agency"})
	svc.drawHeader(pdf, CandidateCVBranding{CompanyName: "Agency"})
	assert.True(t, pdf.GetY() > 0)
}
