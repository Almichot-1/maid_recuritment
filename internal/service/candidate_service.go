package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

var (
	ErrCandidateLocked              = errors.New("candidate is locked")
	ErrForbidden                    = errors.New("forbidden")
	ErrInvalidCandidateInput        = errors.New("invalid candidate input")
	ErrInvalidCandidateUpdateState  = errors.New("candidate cannot be updated in current status")
	ErrInvalidCandidateDeleteState  = errors.New("candidate cannot be deleted in current status")
	ErrInvalidCandidateDocumentType = errors.New("invalid document type")
)

type CandidateInput struct {
	FullName        string
	Nationality     string
	DateOfBirth     *time.Time
	Age             *int
	PlaceOfBirth    string
	Religion        string
	MaritalStatus   string
	ChildrenCount   *int
	EducationLevel  string
	ExperienceYears *int
	Languages       []string
	Skills          []string
}

type UploadCandidateDocumentInput struct {
	DocumentType string
	File         io.Reader
	FileName     string
	FileSize     int64
}

type CandidateCVBranding struct {
	CompanyName string
	LogoDataURL string
}

type CandidateService struct {
	candidateRepository domain.CandidateRepository
	documentRepository  domain.DocumentRepository
	passportRepository  domain.PassportDataRepository
	medicalRepository   domain.MedicalDataRepository
	storageService      StorageService
	pdfService          *PDFService
	pairingService      *PairingService
	medicalService      *MedicalDocumentService
	passportOCRService  *PassportOCRService
}

func NewCandidateService(
	candidateRepository domain.CandidateRepository,
	documentRepository domain.DocumentRepository,
	storageService StorageService,
	pdfService *PDFService,
) (*CandidateService, error) {
	if candidateRepository == nil {
		return nil, fmt.Errorf("candidate repository is nil")
	}
	if documentRepository == nil {
		return nil, fmt.Errorf("document repository is nil")
	}
	if storageService == nil {
		return nil, fmt.Errorf("storage service is nil")
	}
	if pdfService == nil {
		return nil, fmt.Errorf("pdf service is nil")
	}

	return &CandidateService{
		candidateRepository: candidateRepository,
		documentRepository:  documentRepository,
		storageService:      storageService,
		pdfService:          pdfService,
	}, nil
}

func (s *CandidateService) SetPairingService(pairingService *PairingService) {
	s.pairingService = pairingService
}

func (s *CandidateService) SetPassportRepository(passportRepository domain.PassportDataRepository) {
	s.passportRepository = passportRepository
}

func (s *CandidateService) SetMedicalDataRepository(medicalRepository domain.MedicalDataRepository) {
	s.medicalRepository = medicalRepository
}

func (s *CandidateService) SetMedicalDocumentService(medicalService *MedicalDocumentService) {
	s.medicalService = medicalService
}

func (s *CandidateService) SetPassportOCRService(passportOCRService *PassportOCRService) {
	s.passportOCRService = passportOCRService
}

func (s *CandidateService) CreateCandidate(createdBy string, data CandidateInput) (*domain.Candidate, error) {
	if strings.TrimSpace(createdBy) == "" {
		return nil, ErrForbidden
	}
	if err := validateCandidateInput(data); err != nil {
		return nil, err
	}

	languages, err := marshalStringSlice(data.Languages)
	if err != nil {
		return nil, fmt.Errorf("create candidate: marshal languages: %w", err)
	}
	skills, err := marshalStringSlice(data.Skills)
	if err != nil {
		return nil, fmt.Errorf("create candidate: marshal skills: %w", err)
	}

	candidate := &domain.Candidate{
		CreatedBy:       strings.TrimSpace(createdBy),
		FullName:        strings.TrimSpace(data.FullName),
		Nationality:     strings.TrimSpace(data.Nationality),
		DateOfBirth:     normalizeCandidateDate(data.DateOfBirth),
		Age:             data.Age,
		PlaceOfBirth:    strings.TrimSpace(data.PlaceOfBirth),
		Religion:        strings.TrimSpace(data.Religion),
		MaritalStatus:   strings.TrimSpace(data.MaritalStatus),
		ChildrenCount:   data.ChildrenCount,
		EducationLevel:  strings.TrimSpace(data.EducationLevel),
		ExperienceYears: data.ExperienceYears,
		Languages:       languages,
		Skills:          skills,
		Status:          domain.CandidateStatusDraft,
	}
	if candidate.Age == nil {
		candidate.Age = deriveAgePointer(candidate.DateOfBirth)
	}

	if err := s.candidateRepository.Create(candidate); err != nil {
		return nil, err
	}

	return candidate, nil
}

func (s *CandidateService) UpdateCandidate(id, updatedBy string, data CandidateInput) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("candidate id is required")
	}
	if strings.TrimSpace(updatedBy) == "" {
		return ErrForbidden
	}
	if err := validateCandidateInput(data); err != nil {
		return err
	}

	candidate, err := s.candidateRepository.GetByID(id)
	if err != nil {
		return err
	}

	if candidate.CreatedBy != strings.TrimSpace(updatedBy) {
		return ErrForbidden
	}

	if candidate.Status != domain.CandidateStatusDraft && candidate.Status != domain.CandidateStatusAvailable {
		return ErrInvalidCandidateUpdateState
	}

	if isLockedByAnotherUser(candidate, strings.TrimSpace(updatedBy)) {
		return ErrCandidateLocked
	}

	languages, err := marshalStringSlice(data.Languages)
	if err != nil {
		return fmt.Errorf("update candidate: marshal languages: %w", err)
	}
	skills, err := marshalStringSlice(data.Skills)
	if err != nil {
		return fmt.Errorf("update candidate: marshal skills: %w", err)
	}

	candidate.FullName = strings.TrimSpace(data.FullName)
	candidate.Nationality = strings.TrimSpace(data.Nationality)
	candidate.DateOfBirth = normalizeCandidateDate(data.DateOfBirth)
	candidate.Age = data.Age
	if candidate.Age == nil {
		candidate.Age = deriveAgePointer(candidate.DateOfBirth)
	}
	candidate.PlaceOfBirth = strings.TrimSpace(data.PlaceOfBirth)
	candidate.Religion = strings.TrimSpace(data.Religion)
	candidate.MaritalStatus = strings.TrimSpace(data.MaritalStatus)
	candidate.ChildrenCount = data.ChildrenCount
	candidate.EducationLevel = strings.TrimSpace(data.EducationLevel)
	candidate.ExperienceYears = data.ExperienceYears
	candidate.Languages = languages
	candidate.Skills = skills

	return s.candidateRepository.Update(candidate)
}

func (s *CandidateService) GetCandidate(id string) (*domain.Candidate, []*domain.Document, error) {
	candidate, err := s.candidateRepository.GetByID(id)
	if err != nil {
		return nil, nil, err
	}

	documents, err := s.documentRepository.GetByCandidateID(id)
	if err != nil {
		return nil, nil, err
	}

	return candidate, documents, nil
}

func (s *CandidateService) ListCandidates(role string, filters domain.CandidateFilters) ([]*domain.Candidate, error) {
	trimmedRole := strings.TrimSpace(role)

	switch trimmedRole {
	case string(domain.EthiopianAgent):
		if strings.TrimSpace(filters.CreatedBy) == "" {
			return nil, ErrForbidden
		}
	case string(domain.ForeignAgent):
		filters.CreatedBy = ""
		filters.Statuses = []domain.CandidateStatus{domain.CandidateStatusAvailable}
	default:
		return nil, ErrForbidden
	}

	return s.candidateRepository.List(filters)
}

func (s *CandidateService) ListCandidatesForWorkspace(role, userID, pairingID string, filters domain.CandidateFilters) ([]*domain.Candidate, error) {
	trimmedRole := strings.TrimSpace(role)

	switch trimmedRole {
	case string(domain.EthiopianAgent):
		filters.CreatedBy = strings.TrimSpace(userID)
		if filters.SharedOnly {
			if s.pairingService == nil {
				return nil, ErrForbidden
			}
			pairing, err := s.pairingService.ResolveActivePairing(strings.TrimSpace(userID), role, pairingID)
			if err != nil {
				if errors.Is(err, ErrForbidden) {
					return nil, ErrForbidden
				}
				return nil, err
			}
			filters.PairingID = pairing.ID
			filters.SharedOnly = true
		}
	case string(domain.ForeignAgent):
		if s.pairingService == nil {
			return nil, ErrForbidden
		}
		pairing, err := s.pairingService.ResolveActivePairing(strings.TrimSpace(userID), role, pairingID)
		if err != nil {
			if errors.Is(err, ErrForbidden) {
				return nil, ErrForbidden
			}
			return nil, err
		}
		filters.CreatedBy = ""
		filters.Statuses = []domain.CandidateStatus{domain.CandidateStatusAvailable}
		filters.PairingID = pairing.ID
		filters.SharedOnly = true
	default:
		return nil, ErrForbidden
	}

	return s.candidateRepository.List(filters)
}

func (s *CandidateService) PublishCandidate(id, publishedBy string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("candidate id is required")
	}
	if strings.TrimSpace(publishedBy) == "" {
		return ErrForbidden
	}

	candidate, err := s.candidateRepository.GetByID(id)
	if err != nil {
		return err
	}

	if candidate.CreatedBy != strings.TrimSpace(publishedBy) {
		return ErrForbidden
	}
	if candidate.Status != domain.CandidateStatusDraft {
		return repository.ErrInvalidStatusTransition
	}

	candidate.Status = domain.CandidateStatusAvailable
	return s.candidateRepository.Update(candidate)
}

func (s *CandidateService) UploadCandidateDocument(candidateID, uploadedBy string, input UploadCandidateDocumentInput) (*domain.Document, error) {
	if strings.TrimSpace(candidateID) == "" {
		return nil, fmt.Errorf("candidate id is required")
	}
	if strings.TrimSpace(uploadedBy) == "" {
		return nil, ErrForbidden
	}
	if input.File == nil {
		return nil, fmt.Errorf("file is required")
	}
	if strings.TrimSpace(input.FileName) == "" {
		return nil, fmt.Errorf("file name is required")
	}
	if input.FileSize > maxDocumentFileSizeBytes {
		return nil, ErrFileTooLarge
	}

	documentType, err := parseCandidateDocumentType(input.DocumentType)
	if err != nil {
		return nil, err
	}

	bufferedFile, contentType, err := validateAndBufferUpload(input.File, input.FileName)
	if err != nil {
		return nil, err
	}
	if err := validateDocumentTypeContentType(documentType, contentType); err != nil {
		return nil, err
	}

	candidate, err := s.candidateRepository.GetByID(candidateID)
	if err != nil {
		return nil, err
	}

	if candidate.CreatedBy != strings.TrimSpace(uploadedBy) {
		return nil, ErrForbidden
	}

	bufferedBytes, err := io.ReadAll(bufferedFile)
	if err != nil {
		return nil, fmt.Errorf("buffer validated upload: %w", err)
	}

	fileURL, err := s.storageService.Upload(bytes.NewReader(bufferedBytes), input.FileName, contentType)
	if err != nil {
		return nil, fmt.Errorf("upload file: %w", err)
	}

	document := &domain.Document{
		CandidateID:  candidateID,
		DocumentType: documentType,
		FileURL:      fileURL,
		FileName:     strings.TrimSpace(input.FileName),
		FileSize:     input.FileSize,
		UploadedAt:   time.Now().UTC(),
	}

	if err := s.documentRepository.Create(document); err != nil {
		_ = s.storageService.Delete(fileURL)
		return nil, err
	}

	if documentType == domain.Passport && s.passportOCRService != nil && (contentType == "image/jpeg" || contentType == "image/png") {
		if passportData, err := s.passportOCRService.ParseAndStore(candidateID, uploadedBy, bytes.NewReader(bufferedBytes), input.FileName); err == nil {
			_ = s.applyPassportAutofill(candidateID, uploadedBy, passportData)
		}
	}

	if documentType == domain.MedicalDocument && s.medicalService != nil {
		if _, err := s.medicalService.ParseAndStore(candidateID, document, input.FileName, contentType, bufferedBytes); err != nil {
			return document, nil
		}
	}

	return document, nil
}

func (s *CandidateService) GenerateCV(candidateID, generatedBy string, branding CandidateCVBranding) error {
	if strings.TrimSpace(candidateID) == "" {
		return fmt.Errorf("candidate id is required")
	}
	if strings.TrimSpace(generatedBy) == "" {
		return ErrForbidden
	}

	candidate, err := s.candidateRepository.GetByID(candidateID)
	if err != nil {
		return err
	}

	if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(generatedBy) {
		return ErrForbidden
	}

	documents, err := s.documentRepository.GetByCandidateID(candidateID)
	if err != nil {
		return err
	}

	var passportData *domain.PassportData
	if s.passportRepository != nil {
		if storedPassportData, err := s.passportRepository.GetByCandidateID(candidateID); err == nil {
			passportData = storedPassportData
		}
	}

	pdfBytes, err := s.pdfService.GenerateCandidateCV(candidate, documents, branding, passportData)
	if err != nil {
		return err
	}

	pdfFileName := fmt.Sprintf("candidate_%s_cv.pdf", candidateID)
	pdfURL, err := s.storageService.Upload(bytes.NewReader(pdfBytes), pdfFileName, "application/pdf")
	if err != nil {
		return fmt.Errorf("upload cv pdf: %w", err)
	}

	candidate.CVPDFURL = pdfURL
	if err := s.candidateRepository.Update(candidate); err != nil {
		_ = s.storageService.Delete(pdfURL)
		return err
	}

	return nil
}

func (s *CandidateService) DeleteCandidate(candidateID, deletedBy string) error {
	if strings.TrimSpace(candidateID) == "" {
		return fmt.Errorf("candidate id is required")
	}
	if strings.TrimSpace(deletedBy) == "" {
		return ErrForbidden
	}

	candidate, err := s.candidateRepository.GetByID(candidateID)
	if err != nil {
		return err
	}

	if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(deletedBy) {
		return ErrForbidden
	}
	if candidate.Status != domain.CandidateStatusDraft && candidate.Status != domain.CandidateStatusAvailable {
		return ErrInvalidCandidateDeleteState
	}

	return s.candidateRepository.Delete(candidateID)
}

func (s *CandidateService) ApplyPassportAutofill(candidateID, updatedBy string, passportData *domain.PassportData) error {
	return s.applyPassportAutofill(candidateID, updatedBy, passportData)
}

func (s *CandidateService) applyPassportAutofill(candidateID, updatedBy string, passportData *domain.PassportData) error {
	if passportData == nil {
		return nil
	}
	if strings.TrimSpace(candidateID) == "" || strings.TrimSpace(updatedBy) == "" {
		return nil
	}

	candidate, err := s.candidateRepository.GetByID(candidateID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(updatedBy) {
		return ErrForbidden
	}

	if holderName := strings.TrimSpace(passportData.HolderName); holderName != "" {
		candidate.FullName = holderName
	}
	if nationality := strings.TrimSpace(passportData.Nationality); nationality != "" {
		candidate.Nationality = nationality
	}
	if !passportData.DateOfBirth.IsZero() {
		dateOfBirth := passportData.DateOfBirth.UTC()
		candidate.DateOfBirth = &dateOfBirth
	}
	if derivedAge := passportData.Age(time.Now().UTC()); derivedAge > 0 {
		candidate.Age = &derivedAge
	}
	if placeOfBirth := strings.TrimSpace(passportData.PlaceOfBirth); placeOfBirth != "" {
		candidate.PlaceOfBirth = placeOfBirth
	}

	return s.candidateRepository.Update(candidate)
}

func validateCandidateInput(data CandidateInput) error {
	if strings.TrimSpace(data.FullName) == "" {
		return ErrInvalidCandidateInput
	}

	if data.Age != nil && (*data.Age < 18 || *data.Age > 65) {
		return ErrInvalidCandidateInput
	}
	if data.DateOfBirth != nil && data.DateOfBirth.UTC().After(time.Now().UTC()) {
		return ErrInvalidCandidateInput
	}
	if data.DateOfBirth != nil {
		if derivedAge := (&domain.PassportData{DateOfBirth: data.DateOfBirth.UTC()}).Age(time.Now().UTC()); derivedAge > 0 && (derivedAge < 18 || derivedAge > 65) {
			return ErrInvalidCandidateInput
		}
	}
	if data.ChildrenCount != nil && *data.ChildrenCount < 0 {
		return ErrInvalidCandidateInput
	}
	if data.ExperienceYears != nil && (*data.ExperienceYears < 0 || *data.ExperienceYears > 30) {
		return ErrInvalidCandidateInput
	}

	return nil
}

func isLockedByAnotherUser(candidate *domain.Candidate, userID string) bool {
	if candidate == nil || candidate.LockedBy == nil || strings.TrimSpace(*candidate.LockedBy) == "" {
		return false
	}
	if strings.TrimSpace(*candidate.LockedBy) == strings.TrimSpace(userID) {
		return false
	}
	if candidate.LockExpiresAt == nil {
		return true
	}
	return time.Now().UTC().Before(*candidate.LockExpiresAt)
}

func marshalStringSlice(values []string) (json.RawMessage, error) {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}

	data, err := json.Marshal(normalized)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(data), nil
}

func normalizeCandidateDate(value *time.Time) *time.Time {
	if value == nil || value.IsZero() {
		return nil
	}
	normalized := time.Date(value.UTC().Year(), value.UTC().Month(), value.UTC().Day(), 0, 0, 0, 0, time.UTC)
	return &normalized
}

func deriveAgePointer(dateOfBirth *time.Time) *int {
	if dateOfBirth == nil || dateOfBirth.IsZero() {
		return nil
	}
	years := (&domain.PassportData{DateOfBirth: dateOfBirth.UTC()}).Age(time.Now().UTC())
	if years <= 0 {
		return nil
	}
	return &years
}

func parseCandidateDocumentType(value string) (domain.DocumentType, error) {
	switch strings.TrimSpace(value) {
	case string(domain.Passport):
		return domain.Passport, nil
	case string(domain.Photo):
		return domain.Photo, nil
	case string(domain.Video):
		return domain.Video, nil
	case string(domain.MedicalDocument):
		return domain.MedicalDocument, nil
	default:
		return "", ErrInvalidCandidateDocumentType
	}
}
