package service

import (
	"bytes"
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

	contentType, err := detectContentTypeFromFileName(fileName)
	if err != nil {
		return nil, err
	}
	if err := validateDocumentTypeContentType(docType, contentType); err != nil {
		return nil, err
	}

	fileURL, err := s.storageService.Upload(file, fileName, contentType)
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
	case string(domain.Medical):
		return domain.Medical, nil
	default:
		return "", ErrInvalidDocumentType
	}
}

func validateAndBufferUpload(file io.Reader, fileName string) (io.Reader, string, error) {
	if file == nil {
		return nil, "", fmt.Errorf("file is required")
	}
	if strings.TrimSpace(fileName) == "" {
		return nil, "", fmt.Errorf("file name is required")
	}

	expectedContentType, err := detectContentTypeFromFileName(fileName)
	if err != nil {
		return nil, "", err
	}

	const fileSignatureBytes = 512
	header := make([]byte, fileSignatureBytes)
	readBytes, readErr := io.ReadFull(file, header)
	switch readErr {
	case nil, io.EOF, io.ErrUnexpectedEOF:
	default:
		return nil, "", fmt.Errorf("read upload header: %w", readErr)
	}
	header = header[:readBytes]

	actualContentType, err := detectContentTypeFromBytes(header)
	if err != nil {
		return nil, "", err
	}
	if actualContentType != expectedContentType {
		return nil, "", ErrInvalidFileType
	}

	remaining, err := io.ReadAll(io.LimitReader(file, maxDocumentFileSizeBytes+1))
	if err != nil {
		return nil, "", fmt.Errorf("buffer upload body: %w", err)
	}
	if int64(len(header)+len(remaining)) > maxDocumentFileSizeBytes {
		return nil, "", ErrFileTooLarge
	}

	buffered := make([]byte, 0, len(header)+len(remaining))
	buffered = append(buffered, header...)
	buffered = append(buffered, remaining...)

	return bytes.NewReader(buffered), actualContentType, nil
}

func ValidateAndBufferUploadForProfile(file io.Reader, fileName string) (io.Reader, string, error) {
	buffered, contentType, err := validateAndBufferUpload(file, fileName)
	if err != nil {
		return nil, "", err
	}
	switch contentType {
	case "image/jpeg", "image/png", "image/webp":
		return buffered, contentType, nil
	default:
		return nil, "", ErrInvalidFileType
	}
}

func detectContentTypeFromBytes(header []byte) (string, error) {
	switch {
	case bytes.HasPrefix(header, []byte{0xFF, 0xD8, 0xFF}):
		return "image/jpeg", nil
	case bytes.HasPrefix(header, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}):
		return "image/png", nil
	case len(header) >= 12 && string(header[0:4]) == "RIFF" && string(header[8:12]) == "WEBP":
		return "image/webp", nil
	case bytes.HasPrefix(header, []byte("%PDF-")):
		return "application/pdf", nil
	case len(header) >= 12 && string(header[4:8]) == "ftyp":
		return "video/mp4", nil
	default:
		return "", ErrInvalidFileType
	}
}

func extensionForContentType(contentType string) string {
	switch strings.TrimSpace(contentType) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "video/mp4":
		return ".mp4"
	case "application/pdf":
		return ".pdf"
	default:
		return ""
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
	case domain.Medical:
		// Medical certificates may be submitted as PDF or image scans.
		if contentType != "application/pdf" && contentType != "image/jpeg" && contentType != "image/png" {
			return ErrInvalidFileType
		}
	default:
		return ErrInvalidDocumentType
	}
	return nil
}
