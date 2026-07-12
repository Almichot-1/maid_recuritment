package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"maid-recruitment-tracking/internal/domain"
)

const (
	MaxBatchSize       = 200
	maxConcurrentGenCV = 5
)

var (
	ErrCandidateLocked                 = errors.New("candidate is locked")
	ErrForbidden                       = errors.New("forbidden")
	ErrInvalidCandidateInput           = errors.New("invalid candidate input")
	ErrInvalidCandidateUpdateState     = errors.New("candidate cannot be updated in current status")
	ErrInvalidCandidateDeleteState     = errors.New("candidate cannot be deleted in current status")
	ErrInvalidCandidateDocumentType    = errors.New("invalid document type")
	ErrCandidateDocumentNotFound       = errors.New("candidate document not found")
	ErrPublishPairingSelectionRequired = errors.New("multiple active pairings require explicit pairing selection")
	ErrPairingDefaultsRequired         = errors.New("pairing defaults (country and currency) are required before publishing")
	ErrInvalidDefaultForeignPairing    = errors.New("default foreign pairing is invalid")
)

// isoToNationality maps ISO 3166-1 alpha-3 codes to full nationality names.
// The MRZ stores the ISO code (e.g. "ETH") but the candidate form uses full
// nationality adjectives (e.g. "Ethiopian").
var isoToNationality = map[string]string{
	"ETH": "Ethiopian",
	"ERI": "Eritrean",
	"KEN": "Kenyan",
	"SOM": "Somali",
	"SDN": "Sudanese",
	"SSD": "South Sudanese",
	"DJI": "Djiboutian",
	"ARE": "Emirati",
	"SAU": "Saudi",
	"QAT": "Qatari",
	"KWT": "Kuwaiti",
	"OMN": "Omani",
	"BHR": "Bahraini",
	"LBN": "Lebanese",
	"JOR": "Jordanian",
	"EGY": "Egyptian",
	"MAR": "Moroccan",
	"TUN": "Tunisian",
	"DZA": "Algerian",
	"LBY": "Libyan",
	"IRQ": "Iraqi",
	"SYR": "Syrian",
	"YEM": "Yemeni",
	"PSE": "Palestinian",
	"IND": "Indian",
	"PAK": "Pakistani",
	"BGD": "Bangladeshi",
	"LKA": "Sri Lankan",
	"NPL": "Nepali",
	"PHL": "Filipino",
	"IDN": "Indonesian",
	"MMR": "Myanmar",
	"KHM": "Cambodian",
	"VNM": "Vietnamese",
	"THA": "Thai",
	"CHN": "Chinese",
	"NGA": "Nigerian",
	"GHA": "Ghanaian",
	"ZAF": "South African",
	"UGA": "Ugandan",
	"TZA": "Tanzanian",
	"RWA": "Rwandan",
}

type CandidateInput struct {
	FullName            string
	Nationality         string
	DateOfBirth         *time.Time
	Age                 *int
	PlaceOfBirth        string
	PassportNumber      string
	IssueDate           *time.Time
	ExpiryDate          *time.Time
	Gender              string
	IssuingAuthority    string
	ExperienceAbroad    []domain.ExperienceEntry
	Religion            string
	MaritalStatus       string
	ChildrenCount       *int
	EducationLevel      string
	ExperienceYears     *int
	CountryOfExperience string
	Languages           []domain.LanguageEntry
	Skills              []string
	Remark              string
}

type UploadCandidateDocumentInput struct {
	DocumentType string
	File         io.Reader
	FileName     string
	FileSize     int64
}

type CandidateCVBranding struct {
	// Ethiopian agency (right side)
	CompanyName string
	LogoDataURL string

	// Foreign agency (left side)
	ForeignAgencyName        string
	ForeignAgencyLogoDataURL string
}

// SetCandidatePairOverrideInput carries the per-pairing country, salary, and
// logo overrides an Ethiopian agent wants to apply to a candidate's CV.
type SetCandidatePairOverrideInput struct {
	PairingID      string
	CandidateID    string
	CountryApplied string
	SalaryOffered  string
	LogoURL        string
}

type PublishCandidateInput struct {
	PairingID string
}

type PublishCandidateResult struct {
	AutoShared      bool
	SharedPairingID string
}

type CandidateService struct {
	candidateRepository    domain.CandidateRepository
	documentRepository     domain.DocumentRepository
	passportRepository     domain.PassportDataRepository
	medicalRepository      domain.MedicalDataRepository
	userRepository         domain.UserRepository
	pairOverrideRepository domain.CandidatePairOverrideRepository
	shareRepository        domain.CandidatePairShareRepository
	storageService         StorageService
	pdfService             *PDFService
	pairingService         *PairingService
	medicalService         *MedicalDocumentService
	passportOCRService     *PassportOCRService
}

func NewCandidateService(
	candidateRepository domain.CandidateRepository,
	documentRepository domain.DocumentRepository,
	storageService StorageService,
	pdfService *PDFService,
	userRepository domain.UserRepository,
	shareRepository domain.CandidatePairShareRepository,
	pairOverrideRepository domain.CandidatePairOverrideRepository,
	pairingService *PairingService,
	passportRepository domain.PassportDataRepository,
	medicalRepository domain.MedicalDataRepository,
	medicalService *MedicalDocumentService,
	passportOCRService *PassportOCRService,
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
		candidateRepository:    candidateRepository,
		documentRepository:     documentRepository,
		storageService:         storageService,
		pdfService:             pdfService,
		userRepository:         userRepository,
		shareRepository:        shareRepository,
		pairOverrideRepository: pairOverrideRepository,
		pairingService:         pairingService,
		passportRepository:     passportRepository,
		medicalRepository:      medicalRepository,
		medicalService:         medicalService,
		passportOCRService:     passportOCRService,
	}, nil
}

func (s *CandidateService) CreateCandidate(createdBy string, data CandidateInput) (*domain.Candidate, error) {
	if strings.TrimSpace(createdBy) == "" {
		return nil, ErrForbidden
	}
	if err := validateCandidateInput(data); err != nil {
		return nil, err
	}

	languages, err := marshalLanguages(data.Languages)
	if err != nil {
		return nil, fmt.Errorf("create candidate: marshal languages: %w", err)
	}
	skills, err := marshalStringSlice(data.Skills)
	if err != nil {
		return nil, fmt.Errorf("create candidate: marshal skills: %w", err)
	}

	candidate := &domain.Candidate{
		CreatedBy:           strings.TrimSpace(createdBy),
		FullName:            strings.TrimSpace(data.FullName),
		Nationality:         strings.TrimSpace(data.Nationality),
		DateOfBirth:         normalizeCandidateDate(data.DateOfBirth),
		Age:                 data.Age,
		PlaceOfBirth:        strings.TrimSpace(data.PlaceOfBirth),
		PassportNumber:      strings.TrimSpace(data.PassportNumber),
		IssueDate:           normalizeCandidateDate(data.IssueDate),
		ExpiryDate:          normalizeCandidateDate(data.ExpiryDate),
		Gender:              strings.ToUpper(strings.TrimSpace(data.Gender)),
		IssuingAuthority:    strings.TrimSpace(data.IssuingAuthority),
		ExperienceAbroad:    marshalExperienceAbroad(data.ExperienceAbroad),
		Religion:            strings.TrimSpace(data.Religion),
		MaritalStatus:       strings.TrimSpace(data.MaritalStatus),
		ChildrenCount:       data.ChildrenCount,
		EducationLevel:      strings.TrimSpace(data.EducationLevel),
		ExperienceYears:     data.ExperienceYears,
		CountryOfExperience: strings.TrimSpace(data.CountryOfExperience),
		Languages:           languages,
		Skills:              skills,
		Remark:              strings.TrimSpace(data.Remark),
		Status:              domain.CandidateStatusDraft,
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

	languages, err := marshalLanguages(data.Languages)
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
	candidate.PassportNumber = strings.TrimSpace(data.PassportNumber)
	candidate.IssueDate = normalizeCandidateDate(data.IssueDate)
	candidate.ExpiryDate = normalizeCandidateDate(data.ExpiryDate)
	candidate.Gender = strings.ToUpper(strings.TrimSpace(data.Gender))
	candidate.IssuingAuthority = strings.TrimSpace(data.IssuingAuthority)
	candidate.ExperienceAbroad = marshalExperienceAbroad(data.ExperienceAbroad)
	candidate.Religion = strings.TrimSpace(data.Religion)
	candidate.MaritalStatus = strings.TrimSpace(data.MaritalStatus)
	candidate.ChildrenCount = data.ChildrenCount
	candidate.EducationLevel = strings.TrimSpace(data.EducationLevel)
	candidate.ExperienceYears = data.ExperienceYears
	candidate.CountryOfExperience = strings.TrimSpace(data.CountryOfExperience)
	candidate.Languages = languages
	candidate.Skills = skills
	candidate.Remark = strings.TrimSpace(data.Remark)

	if err := s.candidateRepository.Update(candidate); err != nil {
		return err
	}

	// Auto-regenerate CV if one already existed
	if candidate.CVPDFURL != "" {
		idCopy, updatedByCopy := id, updatedBy
		go func() {
			log.Printf("update_candidate: auto-regenerating default CV for candidate=%s", idCopy)
			if err := s.GenerateCV(idCopy, updatedByCopy, "", CandidateCVBranding{}); err != nil {
				log.Printf("update_candidate: auto-regenerate default CV failed: %v", err)
			}
		}()

		// Also regenerate per-pairing CVs
		if s.shareRepository != nil {
			go func() {
				shares, err := s.shareRepository.ListByCandidateID(idCopy, true)
				if err != nil {
					log.Printf("update_candidate: listing shares failed: %v", err)
					return
				}
				for _, share := range shares {
					if err := s.GenerateCV(idCopy, updatedByCopy, share.PairingID, CandidateCVBranding{}); err != nil {
						log.Printf("update_candidate: auto-regenerate CV for pairing=%s failed: %v", share.PairingID, err)
					}
				}
			}()
		}
	}

	return nil
}

// UpdateCandidateStatus changes only the candidate's overall status
func (s *CandidateService) UpdateCandidateStatus(id, updatedBy string, status string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("candidate id is required")
	}
	if strings.TrimSpace(updatedBy) == "" {
		return ErrForbidden
	}

	candidate, err := s.candidateRepository.GetByID(id)
	if err != nil {
		return err
	}

	if candidate.CreatedBy != strings.TrimSpace(updatedBy) {
		return ErrForbidden
	}

	validStatuses := []domain.CandidateStatus{
		domain.CandidateStatusDraft,
		domain.CandidateStatusAvailable,
		domain.CandidateStatusLocked,
		domain.CandidateStatusInProgress,
		domain.CandidateStatusCompleted,
		domain.CandidateStatusRejected,
	}
	valid := false
	for _, vs := range validStatuses {
		if string(vs) == status {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid candidate status: %s", status)
	}

	return s.candidateRepository.UpdateStatus(id, domain.CandidateStatus(status))
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
		filters.Statuses = []domain.CandidateStatus{domain.CandidateStatusAvailable, domain.CandidateStatusLocked}
		filters.CurrentUserID = strings.TrimSpace(userID)
		filters.PairingID = pairing.ID
		filters.SharedOnly = true
	default:
		return nil, ErrForbidden
	}

	return s.candidateRepository.List(filters)
}

func (s *CandidateService) PublishCandidate(id, publishedBy string, input PublishCandidateInput) (*PublishCandidateResult, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("candidate id is required")
	}
	if strings.TrimSpace(publishedBy) == "" {
		return nil, ErrForbidden
	}

	candidate, err := s.candidateRepository.GetByID(id)
	if err != nil {
		log.Printf("publish_candidate: GetByID failed for candidate=%s: %v", id, err)
		return nil, err
	}
	log.Printf("publish_candidate: GetByID succeeded for candidate=%s, status=%s, created_by=%s", id, candidate.Status, candidate.CreatedBy)

	if candidate.CreatedBy != strings.TrimSpace(publishedBy) {
		return nil, ErrForbidden
	}
	if candidate.Status != domain.CandidateStatusDraft && candidate.Status != domain.CandidateStatusAvailable {
		return nil, fmt.Errorf("publish candidate: candidate status must be draft or available, got %s", candidate.Status)
	}

	log.Printf("publish_candidate: calling resolvePublishPairingTarget for user=%s", publishedBy)
	autoShareTarget, err := s.resolvePublishPairingTarget(strings.TrimSpace(publishedBy), strings.TrimSpace(input.PairingID))
	if err != nil {
		log.Printf("publish_candidate: resolvePublishPairingTarget failed: %v", err)
		return nil, err
	}
	log.Printf("publish_candidate: resolvePublishPairingTarget succeeded, autoShareTarget=%v", autoShareTarget != nil)

	if autoShareTarget != nil {
		if autoShareTarget.DefaultCountry == nil || *autoShareTarget.DefaultCountry == "" || autoShareTarget.DefaultCurrency == nil || *autoShareTarget.DefaultCurrency == "" {
			return nil, ErrPairingDefaultsRequired
		}

		// Check if a manual override already exists — if so, respect it
		existingOverride, lookupErr := s.pairOverrideRepository.GetByPairingAndCandidate(
			autoShareTarget.ID, candidate.ID,
		)
		if lookupErr == nil && existingOverride != nil {
			// Manual override exists — don't overwrite with auto-filled values
		} else {
			// Auto-fill empty country/salary from partner defaults
			countryVal := candidate.CountryApplied
			salVal := candidate.SalaryOffered

			if countryVal == "" {
				if autoShareTarget.DefaultCountry != nil && *autoShareTarget.DefaultCountry != "" {
					countryVal = *autoShareTarget.DefaultCountry
				}
			}

			if salVal == "" {
				if autoShareTarget.DefaultSalary != nil {
					salVal = *autoShareTarget.DefaultSalary
				}
				if autoShareTarget.DefaultCurrency != nil && *autoShareTarget.DefaultCurrency != "" {
					if salVal != "" {
						salVal = salVal + " " + *autoShareTarget.DefaultCurrency
					} else {
						salVal = *autoShareTarget.DefaultCurrency
					}
				}
			}

			if countryVal != candidate.CountryApplied || salVal != candidate.SalaryOffered {
				_ = s.pairOverrideRepository.BulkUpsert([]*domain.CandidatePairOverride{
					{
						PairingID:      autoShareTarget.ID,
						CandidateID:    candidate.ID,
						CountryApplied: countryVal,
						SalaryOffered:  salVal,
					},
				})
			}
		}
	}

	if candidate.Status != domain.CandidateStatusAvailable {
		candidate.Status = domain.CandidateStatusAvailable
		log.Printf("publish_candidate: calling Update for candidate=%s", id)
		if err := s.candidateRepository.Update(candidate); err != nil {
			log.Printf("publish_candidate: Update failed: %v", err)
			return nil, err
		}
		log.Printf("publish_candidate: Update succeeded")
	}

	result := &PublishCandidateResult{}
	if autoShareTarget == nil || s.pairingService == nil {
		log.Printf("publish_candidate: no auto-share needed, returning success")
		return result, nil
	}

	log.Printf("publish_candidate: calling ShareCandidate")
	if err := s.pairingService.ShareCandidate(candidate.ID, autoShareTarget.ID, publishedBy); err != nil && !errors.Is(err, ErrCandidateAlreadyShared) {
		log.Printf("publish_candidate: ShareCandidate failed: %v", err)
		return nil, err
	}
	log.Printf("publish_candidate: ShareCandidate succeeded")

	result.AutoShared = true
	result.SharedPairingID = autoShareTarget.ID

	// Auto-generate CV in background
	go func() {
		log.Printf("publish_candidate: auto-generating CV for candidate=%s pairing=%s", candidate.ID, autoShareTarget.ID)
		if err := s.GenerateCV(candidate.ID, publishedBy, autoShareTarget.ID, CandidateCVBranding{}); err != nil {
			log.Printf("publish_candidate: auto-generate CV failed: %v", err)
		}
	}()

	return result, nil
}

type BatchPublishCandidatesInput struct {
	CandidateIDs []string `json:"candidate_ids"`
	PairingIDs   []string `json:"pairing_ids"` // The pairings to share them with
}

type BatchPublishResult struct {
	SuccessCount int      `json:"success_count"`
	ErrorCount   int      `json:"error_count"`
	Errors       []string `json:"errors"`
}

func (s *CandidateService) BatchPublishCandidates(userID string, input BatchPublishCandidatesInput) (*BatchPublishResult, error) {
	if s.pairingService == nil {
		return nil, fmt.Errorf("pairing service unavailable")
	}

	result := &BatchPublishResult{}

	candidates, err := s.candidateRepository.GetByIDs(input.CandidateIDs)
	if err != nil {
		return nil, fmt.Errorf("batch publish: fetch candidates: %w", err)
	}

	ownerMap := make(map[string]*domain.Candidate, len(candidates))
	for _, c := range candidates {
		ownerMap[c.ID] = c
	}

	// Load target pairings
	var targetPairings []*domain.AgencyPairing
	for _, pid := range input.PairingIDs {
		pid = strings.TrimSpace(pid)
		if pid == "" {
			continue
		}
		p, err := s.pairingService.GetPairingByID(pid)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("pairing %s: %v", pid, err))
			continue
		}
		if p.EthiopianUserID != userID {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("pairing %s: forbidden", pid))
			continue
		}
		targetPairings = append(targetPairings, p)
	}

	for _, cid := range input.CandidateIDs {
		cid = strings.TrimSpace(cid)
		if cid == "" {
			continue
		}

		candidate, ok := ownerMap[cid]
		if !ok {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("candidate %s: not found", cid))
			continue
		}
		if candidate.CreatedBy != userID {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("candidate %s: forbidden", cid))
			continue
		}

		// Mark candidate as available on first publish
		needsStatusUpdate := candidate.Status == domain.CandidateStatusDraft

		for _, pairing := range targetPairings {
			pid := pairing.ID

			// Check if already shared
			existingShare, err := s.shareRepository.GetActiveByPairingAndCandidate(pid, cid)
			if err == nil && existingShare != nil {
				result.SuccessCount++
				continue
			}

			// Update status once
			if needsStatusUpdate {
				candidate.Status = domain.CandidateStatusAvailable
				if err := s.candidateRepository.Update(candidate); err != nil {
					result.ErrorCount++
					result.Errors = append(result.Errors, fmt.Sprintf("candidate %s: status update: %v", cid, err))
					continue
				}
				needsStatusUpdate = false
			}

			// Share with this pairing
			if err := s.pairingService.ShareCandidate(cid, pid, userID); err != nil {
				result.ErrorCount++
				result.Errors = append(result.Errors, fmt.Sprintf("candidate %s pairing %s: %v", cid, pid, err))
				continue
			}

			result.SuccessCount++

			// Auto-generate CV in background
			go func(candID, pairID string) {
				if err := s.GenerateCV(candID, userID, pairID, CandidateCVBranding{}); err != nil {
					log.Printf("batch_publish: auto-generate CV failed for candidate=%s pairing=%s: %v", candID, pairID, err)
				}
			}(cid, pid)
		}
	}

	return result, nil
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
		passportData, storedFromPreview, err := s.passportOCRService.StoreCachedPreview(candidateID, uploadedBy, input.FileName, bufferedBytes)
		switch {
		case err != nil:
			log.Printf("candidate_service: cached passport preview skipped for candidate %s: %v", candidateID, err)
			documentBytes := bytes.Clone(bufferedBytes)
			go s.processPassportDocument(candidateID, uploadedBy, input.FileName, documentBytes)
		case storedFromPreview:
			if err := s.applyPassportAutofill(candidateID, uploadedBy, passportData); err != nil {
				log.Printf("candidate_service: cached passport autofill skipped for candidate %s: %v", candidateID, err)
			}
		default:
			documentBytes := bytes.Clone(bufferedBytes)
			go s.processPassportDocument(candidateID, uploadedBy, input.FileName, documentBytes)
		}
	}

	if documentType == domain.MedicalDocument && s.medicalService != nil {
		documentBytes := bytes.Clone(bufferedBytes)
		documentCopy := *document
		go s.processMedicalDocument(candidateID, &documentCopy, input.FileName, contentType, documentBytes)
	}

	// Trigger auto-CV when both passport and photo exist
	if (documentType == domain.Passport || documentType == domain.Photo) && s.pdfService != nil {
		go s.tryAutoGenerateCV(candidateID, uploadedBy)
	}

	return document, nil
}

// GenerateCV generates a PDF CV for the candidate. When pairingID is non-empty
// the service resolves per-pairing overrides (country_applied, salary_offered)
// so each foreign agency gets a CV tailored to their specific posting.
func (s *CandidateService) GenerateCV(candidateID, generatedBy, pairingID string, branding CandidateCVBranding) error {
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

	if s.userRepository != nil {
		if ethiopianUser, err := s.userRepository.GetByID(candidate.CreatedBy); err == nil && ethiopianUser != nil {
			if branding.CompanyName == "" && ethiopianUser.CompanyName != "" {
				branding.CompanyName = ethiopianUser.CompanyName
			}
			if branding.LogoDataURL == "" && ethiopianUser.AvatarURL != "" {
				branding.LogoDataURL = ethiopianUser.AvatarURL
			}
		}
	}

	// Save originals before per-pairing overrides so global DB values aren't corrupted
	origCountry := candidate.CountryApplied
	origSalary := candidate.SalaryOffered

	// Apply per-pairing overrides when a pairingID was supplied.
	if trimmedPairing := strings.TrimSpace(pairingID); trimmedPairing != "" {
		// 1. Resolve overrides
		countryOverride, salaryOverride, logoOverride, err := s.resolveCVJobDetails(candidateID, trimmedPairing)
		if err != nil {
			log.Printf("generate_cv: could not resolve pair overrides for candidate=%s pairing=%s: %v", candidateID, trimmedPairing, err)
		} else {
			if countryOverride != "" {
				candidate.CountryApplied = countryOverride
			}
			if salaryOverride != "" {
				candidate.SalaryOffered = salaryOverride
			}
			if logoOverride != "" {
				branding.ForeignAgencyLogoDataURL = logoOverride
			}
		}

		// 2. Load pairing data for default salary/currency resolution
		var pairing *domain.AgencyPairing
		if s.pairingService != nil {
			p, err := s.pairingService.GetPairingByID(trimmedPairing)
			if err == nil && p != nil {
				pairing = p
			}
		}

		// 3. Apply pairing defaults if no manual override was set
		if pairing != nil {
			if countryOverride == "" && origCountry == "" && pairing.DefaultCountry != nil && *pairing.DefaultCountry != "" {
				candidate.CountryApplied = *pairing.DefaultCountry
			}

			// --- Resolve salary ---
			if salaryOverride != "" {
				candidate.SalaryOffered = salaryOverride
			} else if origSalary != "" {
				// Keep candidate's existing value (may have been pre-filled)
			} else if pairing.DefaultSalary != nil && *pairing.DefaultSalary != "" {
				salStr := *pairing.DefaultSalary
				if pairing.DefaultCurrency != nil && *pairing.DefaultCurrency != "" {
					salStr = fmt.Sprintf("%s %s", *pairing.DefaultSalary, *pairing.DefaultCurrency)
				}
				candidate.SalaryOffered = salStr
			} else if pairing.DefaultCurrency != nil && *pairing.DefaultCurrency != "" {
				candidate.SalaryOffered = fmt.Sprintf("Negotiable %s", *pairing.DefaultCurrency)
			}
		}

		// 4. Resolve Foreign Agent Metadata (Logo, Branding)
		if s.pairingService != nil && s.userRepository != nil {
			if pairing, err := s.pairingService.GetPairingByID(trimmedPairing); err == nil && pairing != nil {
				if foreignUser, err := s.userRepository.GetByID(pairing.ForeignUserID); err == nil && foreignUser != nil {
					if branding.ForeignAgencyName == "" && foreignUser.CompanyName != "" {
						branding.ForeignAgencyName = foreignUser.CompanyName
					}

					if branding.ForeignAgencyLogoDataURL == "" && pairing.PartnerLogoURL != nil && *pairing.PartnerLogoURL != "" {
						branding.ForeignAgencyLogoDataURL = *pairing.PartnerLogoURL
					} else if branding.ForeignAgencyLogoDataURL == "" && foreignUser.AvatarURL != "" {
						branding.ForeignAgencyLogoDataURL = foreignUser.AvatarURL
					}
				}
			}
		}
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

	timestamp := time.Now().Unix()
	var pdfFileName string
	if pairingID != "" {
		pdfFileName = fmt.Sprintf("candidates/%s/cv/%s_%d.pdf", candidateID, pairingID, timestamp)
	} else {
		pdfFileName = fmt.Sprintf("candidates/%s/cv/default_%d.pdf", candidateID, timestamp)
	}
	pdfURL, err := s.storageService.Upload(bytes.NewReader(pdfBytes), pdfFileName, "application/pdf")
	if err != nil {
		return fmt.Errorf("upload cv pdf: %w", err)
	}

	// Restore global fields before persisting so per-pairing overrides don't corrupt the DB
	candidate.CountryApplied = origCountry
	candidate.SalaryOffered = origSalary
	
	// If it's a default CV (no pairing), update the global candidate record
	if pairingID == "" {
		candidate.CVPDFURL = pdfURL
		if err := s.candidateRepository.Update(candidate); err != nil {
			_ = s.storageService.Delete(pdfURL)
			return err
		}
	} else {
		// If pairing-specific, ONLY store on the share record, don't overwrite global CV
		share, err := s.shareRepository.GetActiveByPairingAndCandidate(pairingID, candidateID)
		if err == nil && share != nil {
			_ = s.shareRepository.UpdateCVURL(share.ID, pdfURL)
		}
	}

	return nil
}

// SetPairOverride lets an Ethiopian agent save a per-pairing country/salary
// override for a candidate. The override is stored in candidate_pair_overrides
// and used the next time a CV is generated for that pairing.
func (s *CandidateService) SetPairOverride(callerID string, input SetCandidatePairOverrideInput) error {
	if strings.TrimSpace(callerID) == "" {
		return ErrForbidden
	}
	if strings.TrimSpace(input.PairingID) == "" {
		return fmt.Errorf("set pair override: pairing_id is required")
	}
	if strings.TrimSpace(input.CandidateID) == "" {
		return fmt.Errorf("set pair override: candidate_id is required")
	}
	if s.pairOverrideRepository == nil {
		return fmt.Errorf("set pair override: feature not available")
	}

	// Verify the caller owns the candidate.
	lean, err := s.candidateRepository.GetByIDLean(input.CandidateID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(lean.CreatedBy) != strings.TrimSpace(callerID) {
		return ErrForbidden
	}

	override := &domain.CandidatePairOverride{
		PairingID:      strings.TrimSpace(input.PairingID),
		CandidateID:    strings.TrimSpace(input.CandidateID),
		CountryApplied: strings.TrimSpace(input.CountryApplied),
		SalaryOffered:  strings.TrimSpace(input.SalaryOffered),
		LogoURL:        strings.TrimSpace(input.LogoURL),
	}
	return s.pairOverrideRepository.Upsert(override)
}

// GetPairOverridesForCandidate returns all per-pairing overrides for a candidate
// so the detail page can show the per-partner country/salary table.
func (s *CandidateService) GetPairOverridesForCandidate(candidateID string) ([]*domain.CandidatePairOverride, error) {
	if s.pairOverrideRepository == nil {
		return nil, nil
	}
	return s.pairOverrideRepository.ListByCandidateID(candidateID)
}

// resolveCVJobDetails looks up the per-pairing override first; if none exists
// it falls back to the candidate's global country_applied / salary_offered.
func (s *CandidateService) resolveCVJobDetails(candidateID, pairingID string) (countryApplied, salaryOffered, logoURL string, err error) {
	return s.ResolveCVJobDetails(candidateID, pairingID)
}

// ResolveCVJobDetails is an exported wrapper for resolveCVJobDetails, used by
// the handler layer to present per-pairing country/salary/logo to foreign agents.
func (s *CandidateService) ResolveCVJobDetails(candidateID, pairingID string) (countryApplied, salaryOffered, logoURL string, err error) {
	if s.pairOverrideRepository == nil {
		// No repository wired – return empty strings; caller will use candidate globals.
		return "", "", "", nil
	}
	override, err := s.pairOverrideRepository.GetByPairingAndCandidate(pairingID, candidateID)
	if err != nil {
		return "", "", "", err
	}
	if override != nil {
		return override.CountryApplied, override.SalaryOffered, override.LogoURL, nil
	}
	// No override row exists – signal to keep candidate globals as-is.
	return "", "", "", nil
}

func (s *CandidateService) tryAutoGenerateCV(candidateID, generatedBy string) {
	docs, err := s.documentRepository.GetByCandidateID(candidateID)
	if err != nil {
		log.Printf("try_auto_generate_cv: fetch docs failed for %s: %v", candidateID, err)
		return
	}
	hasPassport, hasPhoto := false, false
	for _, d := range docs {
		switch d.DocumentType {
		case domain.Passport:
			hasPassport = true
		case domain.Photo:
			hasPhoto = true
		}
	}
	if !hasPassport || !hasPhoto {
		return
	}

	log.Printf("try_auto_generate_cv: generating default CV for candidate=%s", candidateID)
	if err := s.GenerateCV(candidateID, generatedBy, "", CandidateCVBranding{}); err != nil {
		log.Printf("try_auto_generate_cv: failed for %s: %v", candidateID, err)
	}
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

func (s *CandidateService) RemoveCandidateDocument(candidateID, documentID, removedBy string) (*domain.Document, error) {
	if strings.TrimSpace(candidateID) == "" {
		return nil, fmt.Errorf("candidate id is required")
	}
	if strings.TrimSpace(documentID) == "" {
		return nil, fmt.Errorf("document id is required")
	}
	if strings.TrimSpace(removedBy) == "" {
		return nil, ErrForbidden
	}

	candidate, err := s.candidateRepository.GetByID(candidateID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(removedBy) {
		return nil, ErrForbidden
	}

	documents, err := s.documentRepository.GetByCandidateID(candidateID)
	if err != nil {
		return nil, err
	}

	var target *domain.Document
	for _, document := range documents {
		if document != nil && strings.TrimSpace(document.ID) == strings.TrimSpace(documentID) {
			target = document
			break
		}
	}
	if target == nil {
		return nil, ErrCandidateDocumentNotFound
	}

	if strings.TrimSpace(target.FileURL) != "" {
		if err := s.storageService.Delete(target.FileURL); err != nil {
			return nil, err
		}
	}
	if err := s.documentRepository.Delete(target.ID); err != nil {
		return nil, err
	}
	if target.DocumentType == domain.MedicalDocument {
		if s.medicalRepository != nil {
			if err := s.medicalRepository.DeleteByCandidateID(candidateID); err != nil {
				return nil, err
			}
		}
	}

	return target, nil
}

func (s *CandidateService) ApplyPassportAutofill(candidateID, updatedBy string, passportData *domain.PassportData) error {
	return s.applyPassportAutofill(candidateID, updatedBy, passportData)
}

func (s *CandidateService) processPassportDocument(candidateID, uploadedBy, fileName string, bufferedBytes []byte) {
	passportData, err := s.passportOCRService.ParseAndStore(candidateID, uploadedBy, bytes.NewReader(bufferedBytes), fileName)
	if err != nil {
		log.Printf("candidate_service: passport OCR skipped for candidate %s: %v", candidateID, err)
		return
	}
	if err := s.applyPassportAutofill(candidateID, uploadedBy, passportData); err != nil {
		log.Printf("candidate_service: passport autofill skipped for candidate %s: %v", candidateID, err)
	}
}

func (s *CandidateService) processMedicalDocument(candidateID string, document *domain.Document, fileName, contentType string, bufferedBytes []byte) {
	if _, err := s.medicalService.ParseAndStore(candidateID, document, fileName, contentType, bufferedBytes); err != nil {
		log.Printf("candidate_service: medical extraction skipped for candidate %s: %v", candidateID, err)
	}
}

func (s *CandidateService) applyPassportAutofill(candidateID, updatedBy string, passportData *domain.PassportData) error {
	if passportData == nil {
		return nil
	}
	if strings.TrimSpace(candidateID) == "" || strings.TrimSpace(updatedBy) == "" {
		return nil
	}

	// Lean ownership check — fetches only id/created_by/status with no
	// Documents JOIN. Fails fast on forbidden without paying the full GetByID cost.
	lean, err := s.candidateRepository.GetByIDLean(candidateID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(lean.CreatedBy) != strings.TrimSpace(updatedBy) {
		return ErrForbidden
	}

	// Ownership confirmed — now fetch the full candidate so Update can persist
	// all unchanged fields correctly.
	candidate, err := s.candidateRepository.GetByID(candidateID)
	if err != nil {
		return err
	}

	if holderName := strings.TrimSpace(passportData.HolderName); holderName != "" {
		candidate.FullName = holderName
	}
	if nationality := strings.TrimSpace(passportData.Nationality); nationality != "" {
		if full, ok := isoToNationality[nationality]; ok {
			candidate.Nationality = full
		} else {
			candidate.Nationality = nationality
		}
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
	if passportNumber := strings.TrimSpace(passportData.PassportNumber); passportNumber != "" {
		candidate.PassportNumber = passportNumber
	}
	if passportData.IssueDate != nil && !passportData.IssueDate.IsZero() {
		issueDate := passportData.IssueDate.UTC()
		candidate.IssueDate = &issueDate
	}
	if !passportData.ExpiryDate.IsZero() {
		expiryDate := passportData.ExpiryDate.UTC()
		candidate.ExpiryDate = &expiryDate
	}
	if gender := strings.TrimSpace(passportData.Gender); gender != "" {
		candidate.Gender = gender
	}
	if issuingAuthority := strings.TrimSpace(passportData.IssuingAuthority); issuingAuthority != "" {
		candidate.IssuingAuthority = issuingAuthority
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

	gender := strings.ToUpper(strings.TrimSpace(data.Gender))
	if gender != "" && gender != "M" && gender != "F" && gender != "MALE" && gender != "FEMALE" {
		return ErrInvalidCandidateInput
	}
	if data.IssueDate != nil && data.ExpiryDate != nil &&
		!data.IssueDate.IsZero() && !data.ExpiryDate.IsZero() &&
		data.IssueDate.UTC().After(data.ExpiryDate.UTC()) {
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

func marshalLanguages(entries []domain.LanguageEntry) (json.RawMessage, error) {
	normalized := make([]domain.LanguageEntry, 0, len(entries))
	for _, e := range entries {
		lang := strings.TrimSpace(e.Language)
		prof := strings.TrimSpace(e.Proficiency)
		if lang == "" {
			continue
		}
		if prof == "" {
			prof = "Basic"
		}
		normalized = append(normalized, domain.LanguageEntry{Language: lang, Proficiency: prof})
	}
	data, err := json.Marshal(normalized)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func marshalExperienceAbroad(entries []domain.ExperienceEntry) json.RawMessage {
	normalized := make([]domain.ExperienceEntry, 0, len(entries))
	for _, e := range entries {
		country := strings.TrimSpace(e.Country)
		if country == "" {
			continue
		}
		years := e.Years
		if years < 0 {
			years = 0
		}
		normalized = append(normalized, domain.ExperienceEntry{Country: country, Years: years})
	}
	data, err := json.Marshal(normalized)
	if err != nil {
		return json.RawMessage("[]")
	}
	return json.RawMessage(data)
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

func (s *CandidateService) resolvePublishPairingTarget(publishedBy, explicitPairingID string) (*domain.AgencyPairing, error) {
	if s.pairingService == nil || s.userRepository == nil {
		return nil, nil
	}

	user, err := s.userRepository.GetByID(strings.TrimSpace(publishedBy))
	if err != nil {
		return nil, err
	}

	activePairings, err := s.pairingService.ListActivePairingsForUser(strings.TrimSpace(publishedBy))
	if err != nil {
		return nil, err
	}

	findPairing := func(pairingID string) *domain.AgencyPairing {
		for _, pairing := range activePairings {
			if pairing != nil && strings.TrimSpace(pairing.ID) == strings.TrimSpace(pairingID) {
				return pairing
			}
		}
		return nil
	}

	if strings.TrimSpace(explicitPairingID) != "" {
		pairing := findPairing(explicitPairingID)
		if pairing == nil {
			return nil, ErrInvalidDefaultForeignPairing
		}
		return pairing, nil
	}

	if !user.AutoShareCandidates {
		return nil, nil
	}

	switch len(activePairings) {
	case 0:
		return nil, nil
	case 1:
		return activePairings[0], nil
	default:
		if user.DefaultForeignPairingID != nil {
			if pairing := findPairing(*user.DefaultForeignPairingID); pairing != nil {
				return pairing, nil
			}
			user.DefaultForeignPairingID = nil
			_ = s.userRepository.Update(user)
		}
		return nil, ErrPublishPairingSelectionRequired
	}
}

type BatchRegenerateCVsInput struct {
	CandidateIDs []string `json:"candidate_ids"`
	PairingID    string   `json:"pairing_id,omitempty"`
}

type BatchResult struct {
	SuccessCount int      `json:"success_count"`
	ErrorCount   int      `json:"error_count"`
	Errors       []string `json:"errors"`
}

func (s *CandidateService) BatchRegenerateCVs(userID string, input BatchRegenerateCVsInput) *BatchResult {
	result := &BatchResult{}
	if len(input.CandidateIDs) == 0 {
		return result
	}

	sem := make(chan struct{}, maxConcurrentGenCV)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, cid := range input.CandidateIDs {
		cid = strings.TrimSpace(cid)
		if cid == "" {
			continue
		}

		if input.PairingID == "" {
			sem <- struct{}{}
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				defer func() { <-sem }()

				if err := s.GenerateCV(id, userID, "", CandidateCVBranding{}); err != nil {
					mu.Lock()
					result.ErrorCount++
					result.Errors = append(result.Errors, fmt.Sprintf("candidate %s default: %v", id, err))
					mu.Unlock()
					return
				}

				shares, err := s.shareRepository.ListByCandidateID(id, true)
				if err != nil {
					mu.Lock()
					result.ErrorCount++
					result.Errors = append(result.Errors, fmt.Sprintf("candidate %s shares: %v", id, err))
					mu.Unlock()
					return
				}
				for _, share := range shares {
					if err := s.GenerateCV(id, userID, share.PairingID, CandidateCVBranding{}); err != nil {
						mu.Lock()
						result.ErrorCount++
						result.Errors = append(result.Errors, fmt.Sprintf("candidate %s pairing %s: %v", id, share.PairingID, err))
						mu.Unlock()
						return
					}
				}
				mu.Lock()
				result.SuccessCount++
				mu.Unlock()
			}(cid)
		} else {
			sem <- struct{}{}
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				defer func() { <-sem }()
				if err := s.GenerateCV(id, userID, input.PairingID, CandidateCVBranding{}); err != nil {
					mu.Lock()
					result.ErrorCount++
					result.Errors = append(result.Errors, fmt.Sprintf("candidate %s: %v", id, err))
					mu.Unlock()
					return
				}
				mu.Lock()
				result.SuccessCount++
				mu.Unlock()
			}(cid)
		}
	}
	wg.Wait()
	return result
}

type BatchSetPairOverrideInput struct {
	CandidateIDs   []string `json:"candidate_ids"`
	PairingID      string   `json:"pairing_id" validate:"required"`
	CountryApplied string   `json:"country_applied"`
	SalaryOffered  string   `json:"salary_offered"`
	LogoURL        string   `json:"logo_url"`
}

func (s *CandidateService) BatchSetPairOverrides(callerID string, input BatchSetPairOverrideInput) *BatchResult {
	result := &BatchResult{}
	if s.pairOverrideRepository == nil {
		result.ErrorCount = len(input.CandidateIDs)
		result.Errors = append(result.Errors, "pair override feature not available")
		return result
	}
	if len(input.CandidateIDs) == 0 {
		return result
	}

	candidates, err := s.candidateRepository.GetByIDs(input.CandidateIDs)
	if err != nil {
		result.ErrorCount = len(input.CandidateIDs)
		result.Errors = append(result.Errors, fmt.Sprintf("fetch candidates: %v", err))
		return result
	}

	ownerMap := make(map[string]bool, len(candidates))
	for _, c := range candidates {
		ownerMap[c.ID] = c.CreatedBy == callerID
	}

	var overrides []*domain.CandidatePairOverride
	for _, cid := range input.CandidateIDs {
		cid = strings.TrimSpace(cid)
		if cid == "" {
			continue
		}
		if !ownerMap[cid] {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("candidate %s: forbidden", cid))
			continue
		}
		overrides = append(overrides, &domain.CandidatePairOverride{
			PairingID:      input.PairingID,
			CandidateID:    cid,
			CountryApplied: input.CountryApplied,
			SalaryOffered:  input.SalaryOffered,
			LogoURL:        input.LogoURL,
		})
	}

	if len(overrides) == 0 {
		return result
	}

	if err := s.pairOverrideRepository.BulkUpsert(overrides); err != nil {
		result.ErrorCount = len(overrides)
		result.Errors = append(result.Errors, fmt.Sprintf("bulk upsert: %v", err))
		return result
	}

	result.SuccessCount = len(overrides)
	return result
}

type BulkPublishInput struct {
	CandidateIDs []string `json:"candidate_ids"`
	PairingIDs   []string `json:"pairing_ids"`
}

func (s *CandidateService) BulkPublish(userID string, input BulkPublishInput) *BatchResult {
	result := &BatchResult{}
	if len(input.CandidateIDs) == 0 {
		return result
	}

	// Load candidates
	candidates, err := s.candidateRepository.GetByIDs(input.CandidateIDs)
	if err != nil {
		result.ErrorCount = len(input.CandidateIDs)
		result.Errors = append(result.Errors, fmt.Sprintf("fetch candidates: %v", err))
		return result
	}

	ownerMap := make(map[string]*domain.Candidate, len(candidates))
	for _, c := range candidates {
		ownerMap[c.ID] = c
	}

	// Resolve target pairings
	var targetPairings []*domain.AgencyPairing
	if len(input.PairingIDs) > 0 {
		for _, pid := range input.PairingIDs {
			p, err := s.pairingService.GetPairingByID(pid)
			if err != nil {
				result.ErrorCount++
				result.Errors = append(result.Errors, fmt.Sprintf("pairing %s: %v", pid, err))
				continue
			}
			if p.EthiopianUserID != userID {
				result.ErrorCount++
				result.Errors = append(result.Errors, fmt.Sprintf("pairing %s: forbidden", pid))
				continue
			}
			targetPairings = append(targetPairings, p)
		}
	} else {
		// Publish to all active pairings
		// Load user to get pairings
		user, err := s.userRepository.GetByID(userID)
		if err != nil {
			result.ErrorCount = len(input.CandidateIDs)
			result.Errors = append(result.Errors, fmt.Sprintf("fetch user: %v", err))
			return result
		}
		_ = user // pairings loaded via pairing service
		pairings, err := s.pairingService.ListActivePairingsForUser(userID)
		if err != nil {
			result.ErrorCount = len(input.CandidateIDs)
			result.Errors = append(result.Errors, fmt.Sprintf("list pairings: %v", err))
			return result
		}
		for _, p := range pairings {
			if p.Status == domain.AgencyPairingActive {
				targetPairings = append(targetPairings, p)
			}
		}
	}

	if len(targetPairings) == 0 {
		result.ErrorCount = len(input.CandidateIDs)
		result.Errors = append(result.Errors, "no valid target pairings")
		return result
	}

	sem := make(chan struct{}, maxConcurrentGenCV)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, cid := range input.CandidateIDs {
		cid = strings.TrimSpace(cid)
		if cid == "" {
			continue
		}

		candidate, ok := ownerMap[cid]
		if !ok {
			mu.Lock()
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("candidate %s: not found", cid))
			mu.Unlock()
			continue
		}
		if candidate.CreatedBy != userID {
			mu.Lock()
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("candidate %s: forbidden", cid))
			mu.Unlock()
			continue
		}

		// Mark candidate as available before sharing
		if candidate.Status == domain.CandidateStatusDraft {
			candidate.Status = domain.CandidateStatusAvailable
			if err := s.candidateRepository.Update(candidate); err != nil {
				mu.Lock()
				result.ErrorCount++
				result.Errors = append(result.Errors, fmt.Sprintf("candidate %s: status update failed: %v", cid, err))
				mu.Unlock()
				continue
			}
		}

		for _, pairing := range targetPairings {
			defaultCountry, defaultSalary, defaultCurrency := "", "", ""
			if pairing.DefaultCountry != nil {
				defaultCountry = *pairing.DefaultCountry
			}
			if pairing.DefaultSalary != nil {
				defaultSalary = *pairing.DefaultSalary
			}
			if pairing.DefaultCurrency != nil {
				defaultCurrency = *pairing.DefaultCurrency
			}
			sem <- struct{}{}
			wg.Add(1)
			go func(candID, pairingID, country, salary, currency string) {
				defer wg.Done()
				defer func() { <-sem }()

				// Check if already shared
				existingShare, err := s.shareRepository.GetActiveByPairingAndCandidate(pairingID, candID)
				if err == nil && existingShare != nil {
					mu.Lock()
					result.SuccessCount++
					mu.Unlock()
					return
				}

				// Check if manual override already exists
				var countryVal, salVal string
				existingOverride, lookupErr := s.pairOverrideRepository.GetByPairingAndCandidate(pairingID, candID)
				if lookupErr == nil && existingOverride != nil {
					// Manual override exists, respect it (don't auto-fill)
					countryVal = existingOverride.CountryApplied
					salVal = existingOverride.SalaryOffered
				} else {
					// Auto-fill empty country/salary from partner defaults
					countryVal = ""
					if country != "" {
						countryVal = country
					}
					salVal = ""
					if salary != "" {
						salVal = salary
					}
					if currency != "" {
						if salVal != "" {
							salVal = salVal + " " + currency
						} else {
							salVal = currency
						}
					}
					if countryVal != "" || salVal != "" {
						_ = s.pairOverrideRepository.BulkUpsert([]*domain.CandidatePairOverride{
							{
								PairingID:      pairingID,
								CandidateID:    candID,
								CountryApplied: countryVal,
								SalaryOffered:  salVal,
							},
						})
					}
				}

				// Share the candidate
				shareErr := s.pairingService.ShareCandidate(candID, pairingID, userID)
				if shareErr != nil {
					mu.Lock()
					result.ErrorCount++
					result.Errors = append(result.Errors, fmt.Sprintf("candidate %s pairing %s: %v", candID, pairingID, shareErr))
					mu.Unlock()
					return
				}

				mu.Lock()
				result.SuccessCount++
				mu.Unlock()

				// Generate per-pairing CV (failure does not undo the share)
				if genErr := s.GenerateCV(candID, userID, pairingID, CandidateCVBranding{}); genErr != nil {
					mu.Lock()
					result.Errors = append(result.Errors, fmt.Sprintf("candidate %s pairing %s cv: %v", candID, pairingID, genErr))
					mu.Unlock()
				}
			}(cid, pairing.ID, defaultCountry, defaultSalary, defaultCurrency)
		}
	}

	wg.Wait()
	return result
}
