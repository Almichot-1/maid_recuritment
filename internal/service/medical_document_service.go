package service

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	ledongpdf "github.com/ledongthuc/pdf"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/ocr"
	"maid-recruitment-tracking/internal/repository"
)

type MedicalDocumentService struct {
	repository   domain.MedicalDataRepository
	ocrProcessor *ocr.OCRProcessor
}

func NewMedicalDocumentService(cfg *config.Config, repository domain.MedicalDataRepository) (*MedicalDocumentService, error) {
	if repository == nil {
		return nil, fmt.Errorf("medical data repository is nil")
	}

	tesseractPath := ""
	ocrLanguage := "eng"
	if cfg != nil {
		tesseractPath = strings.TrimSpace(cfg.TesseractPath)
		if strings.TrimSpace(cfg.OCRLanguage) != "" {
			ocrLanguage = strings.TrimSpace(cfg.OCRLanguage)
		}
	}

	return &MedicalDocumentService{
		repository:   repository,
		ocrProcessor: ocr.NewOCRProcessor(tesseractPath, ocrLanguage),
	}, nil
}

func (s *MedicalDocumentService) ParseAndStore(candidateID string, document *domain.Document, fileName, contentType string, fileBytes []byte) (*domain.MedicalData, error) {
	if strings.TrimSpace(candidateID) == "" {
		return nil, fmt.Errorf("candidate id is required")
	}
	if document == nil {
		return nil, fmt.Errorf("document is required")
	}
	if len(fileBytes) == 0 {
		return nil, fmt.Errorf("file bytes are required")
	}

	extractedText, err := s.extractText(fileName, contentType, fileBytes)
	if err != nil {
		return nil, err
	}

	expiryDate, err := extractMedicalExpiryDate(extractedText)
	if err != nil {
		return nil, err
	}

	medicalData := &domain.MedicalData{
		CandidateID: candidateID,
		DocumentID:  document.ID,
		ExpiryDate:  expiryDate.UTC(),
		RawText:     extractedText,
		ExtractedAt: time.Now().UTC(),
	}

	if err := s.repository.Upsert(medicalData); err != nil {
		return nil, err
	}

	return s.repository.GetByCandidateID(candidateID)
}

func (s *MedicalDocumentService) GetByCandidateID(candidateID string) (*domain.MedicalData, error) {
	data, err := s.repository.GetByCandidateID(candidateID)
	if err != nil {
		if err == repository.ErrMedicalDataNotFound {
			return nil, err
		}
		return nil, err
	}
	return data, nil
}

func (s *MedicalDocumentService) extractText(fileName, contentType string, fileBytes []byte) (string, error) {
	switch strings.TrimSpace(contentType) {
	case "application/pdf":
		reader, err := ledongpdf.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
		if err != nil {
			return "", err
		}
		plainTextReader, err := reader.GetPlainText()
		if err != nil {
			return "", err
		}
		body, err := io.ReadAll(plainTextReader)
		if err != nil {
			return "", err
		}
		return string(body), nil
	case "image/jpeg", "image/png":
		tempFile, err := os.CreateTemp("", "medical-ocr-*"+strings.ToLower(filepath.Ext(fileName)))
		if err != nil {
			return "", err
		}
		tempPath := tempFile.Name()
		defer func() {
			_ = tempFile.Close()
			_ = os.Remove(tempPath)
		}()

		if _, err := io.Copy(tempFile, bytes.NewReader(fileBytes)); err != nil {
			return "", err
		}
		if err := tempFile.Close(); err != nil {
			return "", err
		}

		text, err := s.ocrProcessor.ExtractText(tempPath)
		if err != nil {
			return "", err
		}
		return text, nil
	default:
		return "", fmt.Errorf("medical document parsing only supports pdf, jpg, and png")
	}
}

var (
	medicalLabelPattern = regexp.MustCompile(`(?i)(expir(?:y|ation)|valid\s*(?:until|upto|to)|expires?|expiry\s*date|validity)`)
	medicalDatePatterns = []*regexp.Regexp{
		regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}\b`),
		regexp.MustCompile(`\b\d{2}[/-]\d{2}[/-]\d{4}\b`),
		regexp.MustCompile(`\b\d{1,2}\s+[A-Za-z]{3,9}\s+\d{4}\b`),
		regexp.MustCompile(`\b[A-Za-z]{3,9}\s+\d{1,2},\s+\d{4}\b`),
	}
)

func extractMedicalExpiryDate(rawText string) (time.Time, error) {
	text := strings.ReplaceAll(rawText, "\r", "\n")
	lines := strings.Split(text, "\n")

	candidates := make([]time.Time, 0, 8)

	for index, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		scanText := line
		if medicalLabelPattern.MatchString(line) {
			if index+1 < len(lines) {
				scanText += " " + strings.TrimSpace(lines[index+1])
			}
			candidates = append(candidates, extractDateCandidates(scanText)...)
		}
	}

	if len(candidates) == 0 {
		candidates = append(candidates, extractDateCandidates(text)...)
	}

	now := time.Now().UTC().Truncate(24 * time.Hour)
	var best time.Time
	for _, candidate := range candidates {
		candidate = candidate.UTC()
		if candidate.Before(now.AddDate(0, -1, 0)) {
			continue
		}
		if candidate.After(now.AddDate(5, 0, 0)) {
			continue
		}
		if best.IsZero() || candidate.Before(best) {
			best = candidate
		}
	}

	if best.IsZero() {
		return time.Time{}, fmt.Errorf("medical expiry date not found")
	}

	return best, nil
}

func extractDateCandidates(text string) []time.Time {
	candidates := make([]time.Time, 0, 6)
	for _, pattern := range medicalDatePatterns {
		for _, match := range pattern.FindAllString(text, -1) {
			if parsed, ok := parseMedicalDate(match); ok {
				candidates = append(candidates, parsed)
			}
		}
	}
	return candidates
}

func parseMedicalDate(value string) (time.Time, bool) {
	layouts := []string{
		"2006-01-02",
		"02-01-2006",
		"02/01/2006",
		"2 Jan 2006",
		"02 Jan 2006",
		"2 January 2006",
		"02 January 2006",
		"Jan 2, 2006",
		"January 2, 2006",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, strings.TrimSpace(value), time.UTC); err == nil {
			return parsed.UTC(), true
		}
	}
	return time.Time{}, false
}
