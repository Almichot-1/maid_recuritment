package service

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"maid-recruitment-tracking/internal/domain"
)

const maxDocumentFileSizeBytes int64 = 50 * 1024 * 1024

var (
	ErrInvalidDocumentType = errors.New("invalid document type")
	ErrFileTooLarge        = errors.New("file size exceeds 50MB")
	ErrInvalidFileType     = errors.New("invalid file type for document type")
)

type DocumentService struct {
	documentRepository domain.DocumentRepository
	storageService     StorageService
}

func NewDocumentService(documentRepository domain.DocumentRepository, storageService StorageService) (*DocumentService, error) {
	if documentRepository == nil {
		return nil, fmt.Errorf("document repository is nil")
	}
	if storageService == nil {
		return nil, fmt.Errorf("storage service is nil")
	}

	return &DocumentService{
		documentRepository: documentRepository,
		storageService:     storageService,
	}, nil
}

func (s *DocumentService) UploadDocument(candidateID, documentType string, file io.Reader, fileName string, fileSize int64) (*domain.Document, error) {
	if strings.TrimSpace(candidateID) == "" {
		return nil, fmt.Errorf("candidate id is required")
	}
	if file == nil {
		return nil, fmt.Errorf("file is required")
	}
	if strings.TrimSpace(fileName) == "" {
		return nil, fmt.Errorf("file name is required")
	}
	if fileSize > maxDocumentFileSizeBytes {
		return nil, ErrFileTooLarge
	}

	docType, err := parseDocumentType(documentType)
	if err != nil {
		return nil, err
	}

	bufferedFile, contentType, err := validateAndBufferUpload(file, fileName)
	if err != nil {
		return nil, err
	}
	if err := validateDocumentTypeContentType(docType, contentType); err != nil {
		return nil, err
	}

	fileURL, err := s.storageService.Upload(bufferedFile, fileName, contentType)
	if err != nil {
		return nil, fmt.Errorf("upload file: %w", err)
	}

	document := &domain.Document{
		ID:           uuid.NewString(),
		CandidateID:  candidateID,
		DocumentType: docType,
		FileURL:      fileURL,
		FileName:     fileName,
		FileSize:     fileSize,
		UploadedAt:   time.Now(),
	}

	if err := s.documentRepository.Create(document); err != nil {
		_ = s.storageService.Delete(fileURL)
		return nil, fmt.Errorf("save document: %w", err)
	}

	return document, nil
}

func parseDocumentType(value string) (domain.DocumentType, error) {
	switch strings.TrimSpace(value) {
	case string(domain.Passport):
		return domain.Passport, nil
	case string(domain.Photo):
		return domain.Photo, nil
	case string(domain.Video):
		return domain.Video, nil
	default:
		return "", ErrInvalidDocumentType
	}
}

func detectContentTypeFromFileName(fileName string) (string, error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg", nil
	case ".png":
		return "image/png", nil
	case ".mp4":
		return "video/mp4", nil
	case ".pdf":
		return "application/pdf", nil
	default:
		return "", ErrInvalidFileType
	}
}

func validateDocumentTypeContentType(documentType domain.DocumentType, contentType string) error {
	switch documentType {
	case domain.Passport:
		if contentType != "application/pdf" && contentType != "image/jpeg" && contentType != "image/png" {
			return ErrInvalidFileType
		}
	case domain.Photo:
		if contentType != "image/jpeg" && contentType != "image/png" {
			return ErrInvalidFileType
		}
	case domain.Video:
		if contentType != "video/mp4" {
			return ErrInvalidFileType
		}
	default:
		return ErrInvalidDocumentType
	}
	return nil
}
