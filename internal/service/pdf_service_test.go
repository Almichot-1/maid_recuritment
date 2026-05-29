package service

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jung-kurt/gofpdf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"maid-recruitment-tracking/internal/domain"
)

func TestPDFService_NewAndGenerateValidation(t *testing.T) {
	svc := NewPDFService()
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(imgBytes)
		case "/bad-status":
			w.WriteHeader(http.StatusBadGateway)
		case "/bad-type":
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("x"))
		}
	}))
	defer server.Close()

	body, contentType, err := fetchImage(server.URL + "/ok")
	require.NoError(t, err)
	assert.Equal(t, "image/png", contentType)
	assert.NotEmpty(t, body)

	_, _, err = fetchImage(server.URL + "/bad-status")
	require.Error(t, err)

	_, _, err = fetchImage(server.URL + "/bad-type")
	require.ErrorIs(t, err, ErrMissingRequiredDocuments)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	require.NoError(t, addImageFromBytes(pdf, imgBytes, "image/png", "img1", 10, 10, 10, 10))
}

func TestPDFService_DrawHelpers(t *testing.T) {
	svc := NewPDFService()
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	svc.drawBrandingHeader(pdf, CandidateCVBranding{CompanyName: "Agency"})
	svc.drawHeader(pdf, CandidateCVBranding{CompanyName: "Agency"})
	assert.True(t, pdf.GetY() > 0)
}
