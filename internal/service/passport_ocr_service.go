package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/ocr"
	"maid-recruitment-tracking/internal/repository"
)

var (
	ErrPassportOCRRequiresImage = errors.New("passport OCR requires a JPG or PNG image")
	ErrPassportDataNotFound     = errors.New("passport data not found")
	ErrPassportOCRParseFailed   = errors.New("passport OCR could not parse the image")
	ErrPassportOCRUnavailable   = errors.New("passport OCR is unavailable")
)

type PassportOCRService struct {
	candidateRepository domain.CandidateRepository
	passportRepository  domain.PassportDataRepository
	ocrProcessor        *ocr.OCRProcessor
}

func NewPassportOCRService(
	cfg *config.Config,
	candidateRepository domain.CandidateRepository,
	passportRepository domain.PassportDataRepository,
) (*PassportOCRService, error) {
	if candidateRepository == nil {
		return nil, fmt.Errorf("candidate repository is nil")
	}
	if passportRepository == nil {
		return nil, fmt.Errorf("passport repository is nil")
	}

	tesseractPath := ""
	ocrLanguage := "eng"
	if cfg != nil {
		tesseractPath = strings.TrimSpace(cfg.TesseractPath)
		if strings.TrimSpace(cfg.OCRLanguage) != "" {
			ocrLanguage = strings.TrimSpace(cfg.OCRLanguage)
		}
	}

	return &PassportOCRService{
		candidateRepository: candidateRepository,
		passportRepository:  passportRepository,
		ocrProcessor:        ocr.NewOCRProcessor(tesseractPath, ocrLanguage),
	}, nil
}

func (s *PassportOCRService) ParseAndStore(candidateID, requestedBy string, file io.Reader, fileName string) (*domain.PassportData, error) {
	if strings.TrimSpace(candidateID) == "" {
		return nil, fmt.Errorf("candidate id is required")
	}
	if strings.TrimSpace(requestedBy) == "" {
		return nil, ErrForbidden
	}
	if file == nil {
		return nil, fmt.Errorf("file is required")
	}
	if strings.TrimSpace(fileName) == "" {
		return nil, fmt.Errorf("file name is required")
	}

	contentType, err := detectContentTypeFromFileName(fileName)
	if err != nil {
		return nil, err
	}
	if contentType != "image/jpeg" && contentType != "image/png" {
		return nil, ErrPassportOCRRequiresImage
	}

	candidate, err := s.candidateRepository.GetByID(candidateID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(requestedBy) {
		return nil, ErrForbidden
	}

	buffer, err := io.ReadAll(io.LimitReader(file, maxDocumentFileSizeBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read passport image: %w", err)
	}
	if int64(len(buffer)) > maxDocumentFileSizeBytes {
		return nil, ErrFileTooLarge
	}

	tempFile, err := os.CreateTemp("", "passport-ocr-*"+strings.ToLower(filepath.Ext(fileName)))
	if err != nil {
		return nil, fmt.Errorf("create temp passport image: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
	}()

	if _, err := io.Copy(tempFile, bytes.NewReader(buffer)); err != nil {
		return nil, fmt.Errorf("write temp passport image: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return nil, fmt.Errorf("close temp passport image: %w", err)
	}

	result, err := s.ocrProcessor.ExtractPassportData(tempPath)
	if err != nil {
		if isPassportOCRUnavailable(err) {
			return nil, fmt.Errorf("%w: %v", ErrPassportOCRUnavailable, err)
		}
		return nil, fmt.Errorf("%w: %v", ErrPassportOCRParseFailed, err)
	}

	holderName := strings.TrimSpace(strings.TrimSpace(result.GivenNames + " " + result.Surname))
	passportData := &domain.PassportData{
		CandidateID:      candidateID,
		HolderName:       holderName,
		PassportNumber:   strings.TrimSpace(result.PassportNumber),
		CountryCode:      strings.TrimSpace(result.CountryCode),
		Nationality:      strings.TrimSpace(result.Nationality),
		DateOfBirth:      result.DateOfBirth.UTC(),
		PlaceOfBirth:     strings.TrimSpace(result.PlaceOfBirth),
		Gender:           strings.TrimSpace(result.Sex),
		ExpiryDate:       result.DateOfExpiry.UTC(),
		IssuingAuthority: strings.TrimSpace(result.IssuingAuthority),
		MRZLine1:         strings.TrimSpace(result.MRZLine1),
		MRZLine2:         strings.TrimSpace(result.MRZLine2),
		Confidence:       result.Confidence,
		ExtractedAt:      time.Now().UTC(),
	}
	if !result.DateOfIssue.IsZero() {
		issueDate := result.DateOfIssue.UTC()
		passportData.IssueDate = &issueDate
	}

	if err := s.passportRepository.Upsert(passportData); err != nil {
		return nil, err
	}

	return s.passportRepository.GetByCandidateID(candidateID)
}

func (s *PassportOCRService) ParsePreview(file io.Reader, fileName string) (*domain.PassportData, error) {
	if file == nil {
		return nil, fmt.Errorf("file is required")
	}
	if strings.TrimSpace(fileName) == "" {
		return nil, fmt.Errorf("file name is required")
	}

	contentType, err := detectContentTypeFromFileName(fileName)
	if err != nil {
		return nil, err
	}
	if contentType != "image/jpeg" && contentType != "image/png" {
		return nil, ErrPassportOCRRequiresImage
	}

	buffer, err := io.ReadAll(io.LimitReader(file, maxDocumentFileSizeBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read passport image: %w", err)
	}
	if int64(len(buffer)) > maxDocumentFileSizeBytes {
		return nil, ErrFileTooLarge
	}

	tempFile, err := os.CreateTemp("", "passport-ocr-preview-*"+strings.ToLower(filepath.Ext(fileName)))
	if err != nil {
		return nil, fmt.Errorf("create temp passport image: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
	}()

	if _, err := io.Copy(tempFile, bytes.NewReader(buffer)); err != nil {
		return nil, fmt.Errorf("write temp passport image: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return nil, fmt.Errorf("close temp passport image: %w", err)
	}

	result, err := s.ocrProcessor.ExtractPassportData(tempPath)
	if err != nil {
		if isPassportOCRUnavailable(err) {
			return nil, fmt.Errorf("%w: %v", ErrPassportOCRUnavailable, err)
		}
		return nil, fmt.Errorf("%w: %v", ErrPassportOCRParseFailed, err)
	}

	holderName := strings.TrimSpace(strings.TrimSpace(result.GivenNames + " " + result.Surname))
	passportData := &domain.PassportData{
		HolderName:       holderName,
		PassportNumber:   strings.TrimSpace(result.PassportNumber),
		CountryCode:      strings.TrimSpace(result.CountryCode),
		Nationality:      strings.TrimSpace(result.Nationality),
		DateOfBirth:      result.DateOfBirth.UTC(),
		PlaceOfBirth:     strings.TrimSpace(result.PlaceOfBirth),
		Gender:           strings.TrimSpace(result.Sex),
		ExpiryDate:       result.DateOfExpiry.UTC(),
		IssuingAuthority: strings.TrimSpace(result.IssuingAuthority),
		MRZLine1:         strings.TrimSpace(result.MRZLine1),
		MRZLine2:         strings.TrimSpace(result.MRZLine2),
		Confidence:       result.Confidence,
		ExtractedAt:      time.Now().UTC(),
	}
	if !result.DateOfIssue.IsZero() {
		issueDate := result.DateOfIssue.UTC()
		passportData.IssueDate = &issueDate
	}

	return passportData, nil
}

func (s *PassportOCRService) GetByCandidateID(candidateID, requestedBy string) (*domain.PassportData, error) {
	if strings.TrimSpace(candidateID) == "" {
		return nil, fmt.Errorf("candidate id is required")
	}
	if strings.TrimSpace(requestedBy) == "" {
		return nil, ErrForbidden
	}

	candidate, err := s.candidateRepository.GetByID(candidateID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(requestedBy) {
		return nil, ErrForbidden
	}

	data, err := s.passportRepository.GetByCandidateID(candidateID)
	if err != nil {
		if errors.Is(err, repository.ErrPassportDataNotFound) {
			return nil, ErrPassportDataNotFound
		}
		return nil, err
	}
	return data, nil
}

func isPassportOCRUnavailable(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "executable file not found"):
		return true
	case strings.Contains(message, "system cannot find the file specified"):
		return true
	case strings.Contains(message, "tesseract is not recognized"):
		return true
	default:
		return false
	}
}
