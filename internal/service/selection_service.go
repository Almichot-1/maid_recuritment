package service

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

var (
	ErrCandidateNotAvailable                = errors.New("candidate not available")
	ErrAlreadySelected                      = errors.New("candidate already selected")
	ErrNotForeignAgent                      = errors.New("only foreign agent can select candidate")
	ErrInvalidSelectionDocumentType         = errors.New("invalid selection document type")
	ErrSelectionSupportingDocumentsRequired = errors.New("employer contract and employer id are required before approval")
)

type NotificationSender interface {
	IsForeignAgent(userID string) (bool, error)
	Send(userID, title, message, notificationType, relatedEntityType, relatedEntityID string) error
}

type dbProvider interface {
	DB() *gorm.DB
}

type SelectionService struct {
	selectionRepository domain.SelectionRepository
	candidateRepository domain.CandidateRepository
	notificationService NotificationSender
	platformSettings    PlatformSettingsReader
	storageService      StorageService
	pairingService      *PairingService
	db                  *gorm.DB
}

func NewSelectionService(
	selectionRepository domain.SelectionRepository,
	candidateRepository domain.CandidateRepository,
	notificationService NotificationSender,
) (*SelectionService, error) {
	if selectionRepository == nil {
		return nil, fmt.Errorf("selection repository is nil")
	}
	if candidateRepository == nil {
		return nil, fmt.Errorf("candidate repository is nil")
	}
	if notificationService == nil {
		return nil, fmt.Errorf("notification service is nil")
	}

	dbSource, ok := selectionRepository.(dbProvider)
	if !ok || dbSource.DB() == nil {
		return nil, fmt.Errorf("selection repository does not expose transaction db")
	}

	return &SelectionService{
		selectionRepository: selectionRepository,
		candidateRepository: candidateRepository,
		notificationService: notificationService,
		db:                  dbSource.DB(),
	}, nil
}

func (s *SelectionService) SetPlatformSettingsReader(platformSettings PlatformSettingsReader) {
	s.platformSettings = platformSettings
}

func (s *SelectionService) SetStorageService(storageService StorageService) {
	s.storageService = storageService
}

func (s *SelectionService) SetPairingService(pairingService *PairingService) {
	s.pairingService = pairingService
}

type UploadSelectionDocumentInput struct {
	DocumentType string
	File         io.Reader
	FileName     string
	FileSize     int64
}

func (s *SelectionService) SelectCandidate(candidateID, selectedBy string) (*domain.Selection, error) {
	return s.SelectCandidateInPairing(candidateID, selectedBy, "")
}

func (s *SelectionService) SelectCandidateInPairing(candidateID, selectedBy, pairingID string) (*domain.Selection, error) {
	candidateID = strings.TrimSpace(candidateID)
	selectedBy = strings.TrimSpace(selectedBy)

	if candidateID == "" {
		return nil, fmt.Errorf("candidate id is required")
	}
	if selectedBy == "" {
		return nil, fmt.Errorf("selectedBy is required")
	}

	isForeign, err := s.notificationService.IsForeignAgent(selectedBy)
	if err != nil {
		return nil, fmt.Errorf("validate foreign agent role: %w", err)
	}
	if !isForeign {
		return nil, ErrNotForeignAgent
	}
	var pairing *domain.AgencyPairing
	if s.pairingService != nil {
		pairing, err = s.pairingService.ResolveActivePairing(selectedBy, string(domain.ForeignAgent), pairingID)
		if err != nil {
			return nil, err
		}
	}

	selection := &domain.Selection{}
	lockDurationHours := 24
	if s.platformSettings != nil {
		settings, err := s.platformSettings.Get()
		if err == nil && settings != nil && settings.SelectionLockDurationHours > 0 {
			lockDurationHours = settings.SelectionLockDurationHours
		}
	}
	expiresAt := time.Now().UTC().Add(time.Duration(lockDurationHours) * time.Hour)

	err = s.db.Transaction(func(tx *gorm.DB) error {
		var candidate domain.Candidate
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", candidateID).First(&candidate).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrCandidateNotFound
			}
			return fmt.Errorf("load candidate: %w", err)
		}

		if candidate.Status != domain.CandidateStatusAvailable {
			return ErrCandidateNotAvailable
		}
		if pairing != nil {
			if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(pairing.EthiopianUserID) {
				return ErrForbidden
			}
			isShared, err := s.pairingService.IsCandidateSharedWithPairing(candidate.ID, pairing.ID)
			if err != nil {
				return err
			}
			if !isShared {
				return ErrForbidden
			}
		}

		var existingPending domain.Selection
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("candidate_id = ? AND status = ?", candidateID, domain.SelectionPending).
			First(&existingPending).Error
		if err == nil {
			return ErrAlreadySelected
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("check existing pending selection: %w", err)
		}

		selection.ID = uuid.NewString()
		selection.CandidateID = candidateID
		if pairing != nil {
			selection.PairingID = pairing.ID
		}
		selection.SelectedBy = selectedBy
		selection.Status = domain.SelectionPending
		selection.ExpiresAt = expiresAt

		createQuery := tx
		if pairing == nil {
			createQuery = createQuery.Omit("PairingID")
		}
		if err := createQuery.Create(selection).Error; err != nil {
			if isSelectionConflictError(err) {
				return ErrAlreadySelected
			}
			return fmt.Errorf("create selection: %w", err)
		}

		now := time.Now().UTC()
		lockedBy := selectedBy
		result := tx.Model(&domain.Candidate{}).Where("id = ?", candidateID).Updates(map[string]any{
			"status":          domain.CandidateStatusLocked,
			"locked_by":       lockedBy,
			"locked_at":       now,
			"lock_expires_at": expiresAt,
		})
		if result.Error != nil {
			return fmt.Errorf("lock candidate: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return ErrCandidateNotAvailable
		}

		if err := s.notificationService.Send(
			candidate.CreatedBy,
			"Candidate selected",
			"A foreign agent has selected one of your candidates. Open the selection to review the approval request.",
			"selection",
			"selection",
			selection.ID,
		); err != nil {
			return fmt.Errorf("notify candidate owner: %w", err)
		}

		if err := s.notificationService.Send(
			selectedBy,
			"Selection confirmed",
			fmt.Sprintf("You have successfully selected this candidate for %d hours. Upload the contract package and continue from the selection page.", lockDurationHours),
			"selection",
			"selection",
			selection.ID,
		); err != nil {
			return fmt.Errorf("notify selecting agent: %w", err)
		}

		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrCandidateNotAvailable):
			return nil, ErrCandidateNotAvailable
		case errors.Is(err, ErrAlreadySelected):
			return nil, ErrAlreadySelected
		case errors.Is(err, ErrNotForeignAgent):
			return nil, ErrNotForeignAgent
		default:
			return nil, err
		}
	}

	return selection, nil
}

func (s *SelectionService) GetSelection(id string) (*domain.Selection, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("selection id is required")
	}
	return s.selectionRepository.GetByID(id)
}

func (s *SelectionService) GetMySelections(userID string) ([]*domain.Selection, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user id is required")
	}
	return s.selectionRepository.GetBySelectedBy(userID)
}

func (s *SelectionService) GetSelectionsForUser(userID, role string) ([]*domain.Selection, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user id is required")
	}

	switch strings.TrimSpace(role) {
	case string(domain.ForeignAgent):
		return s.selectionRepository.GetBySelectedBy(userID)
	case string(domain.EthiopianAgent):
		return s.selectionRepository.GetByCandidateOwner(userID)
	default:
		return nil, ErrForbidden
	}
}

func (s *SelectionService) GetSelectionsForWorkspace(userID, role, pairingID string) ([]*domain.Selection, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user id is required")
	}
	if s.pairingService == nil {
		return s.GetSelectionsForUser(userID, role)
	}

	pairing, err := s.pairingService.ResolveActivePairing(userID, role, pairingID)
	if err != nil {
		return nil, err
	}

	switch strings.TrimSpace(role) {
	case string(domain.ForeignAgent):
		return s.selectionRepository.GetBySelectedByAndPairing(userID, pairing.ID)
	case string(domain.EthiopianAgent):
		return s.selectionRepository.GetByCandidateOwnerAndPairing(userID, pairing.ID)
	default:
		return nil, ErrForbidden
	}
}

func (s *SelectionService) UploadSelectionDocument(selectionID, uploadedBy string, input UploadSelectionDocumentInput) (*domain.Selection, error) {
	if strings.TrimSpace(selectionID) == "" {
		return nil, fmt.Errorf("selection id is required")
	}
	if strings.TrimSpace(uploadedBy) == "" {
		return nil, fmt.Errorf("uploadedBy is required")
	}
	if s.storageService == nil {
		return nil, fmt.Errorf("storage service is not configured")
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

	documentType, err := parseSelectionDocumentType(input.DocumentType)
	if err != nil {
		return nil, err
	}

	bufferedFile, contentType, err := validateAndBufferUpload(input.File, input.FileName)
	if err != nil {
		return nil, err
	}
	if err := validateSelectionDocumentContentType(documentType, contentType); err != nil {
		return nil, err
	}

	var uploadedURL string
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var selection domain.Selection
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", selectionID).First(&selection).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrSelectionNotFound
			}
			return fmt.Errorf("load selection: %w", err)
		}

		if strings.TrimSpace(selection.SelectedBy) != strings.TrimSpace(uploadedBy) {
			return ErrNotAuthorized
		}
		if selection.Status != domain.SelectionPending {
			return ErrSelectionNotPending
		}

		uploadedURL, err = s.storageService.Upload(bufferedFile, input.FileName, contentType)
		if err != nil {
			return fmt.Errorf("upload selection document: %w", err)
		}

		now := time.Now().UTC()
		updates := map[string]any{}
		previousURL := ""
		switch documentType {
		case selectionDocumentTypeContract:
			previousURL = selection.EmployerContractURL
			updates["employer_contract_url"] = uploadedURL
			updates["employer_contract_file_name"] = strings.TrimSpace(input.FileName)
			updates["employer_contract_uploaded_at"] = now
		case selectionDocumentTypeEmployerID:
			previousURL = selection.EmployerIDURL
			updates["employer_id_url"] = uploadedURL
			updates["employer_id_file_name"] = strings.TrimSpace(input.FileName)
			updates["employer_id_uploaded_at"] = now
		default:
			return ErrInvalidSelectionDocumentType
		}

		if err := tx.Model(&domain.Selection{}).Where("id = ?", selectionID).Updates(updates).Error; err != nil {
			_ = s.storageService.Delete(uploadedURL)
			return fmt.Errorf("persist selection document: %w", err)
		}

		if strings.TrimSpace(previousURL) != "" && previousURL != uploadedURL {
			_ = s.storageService.Delete(previousURL)
		}

		return nil
	})
	if err != nil {
		if strings.TrimSpace(uploadedURL) != "" {
			_ = s.storageService.Delete(uploadedURL)
		}
		return nil, err
	}

	return s.selectionRepository.GetByID(selectionID)
}

func (s *SelectionService) ProcessExpiredSelections() error {
	if s.platformSettings != nil {
		settings, err := s.platformSettings.Get()
		if err == nil && settings != nil && !settings.AutoExpireSelections {
			return nil
		}
	}

	expiredSelections, err := s.selectionRepository.GetExpiredSelections()
	if err != nil {
		return fmt.Errorf("load expired selections: %w", err)
	}

	for _, selection := range expiredSelections {
		if selection == nil {
			continue
		}

		selectionID := selection.ID
		candidateID := selection.CandidateID
		selectedBy := selection.SelectedBy

		err := s.db.Transaction(func(tx *gorm.DB) error {
			var currentSelection domain.Selection
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", selectionID).First(&currentSelection).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				return fmt.Errorf("load selection: %w", err)
			}

			if currentSelection.Status != domain.SelectionPending || !currentSelection.ExpiresAt.Before(time.Now().UTC()) {
				return nil
			}

			if err := tx.Model(&domain.Selection{}).Where("id = ?", selectionID).Update("status", domain.SelectionExpired).Error; err != nil {
				return fmt.Errorf("update selection status: %w", err)
			}

			var candidate domain.Candidate
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", candidateID).First(&candidate).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				return fmt.Errorf("load candidate: %w", err)
			}

			if candidate.Status == domain.CandidateStatusLocked {
				updates := map[string]any{
					"status":          domain.CandidateStatusAvailable,
					"locked_by":       nil,
					"locked_at":       nil,
					"lock_expires_at": nil,
				}
				if err := tx.Model(&domain.Candidate{}).Where("id = ?", candidateID).Updates(updates).Error; err != nil {
					return fmt.Errorf("unlock candidate: %w", err)
				}
			}

			if err := s.notificationService.Send(
				candidate.CreatedBy,
				"Selection expired",
				"A pending selection has expired and candidate is available again.",
				"expiry",
				"selection",
				selectionID,
			); err != nil {
				return fmt.Errorf("notify owner on expiry: %w", err)
			}

			if err := s.notificationService.Send(
				selectedBy,
				"Selection expired",
				"Your pending selection has expired.",
				"expiry",
				"selection",
				selectionID,
			); err != nil {
				return fmt.Errorf("notify foreign agent on expiry: %w", err)
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func isSelectionConflictError(err error) bool {
	if errors.Is(err, repository.ErrActiveSelectionExists) {
		return true
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && strings.Contains(pgErr.Message, "idx_selections_one_pending_per_candidate")
	}
	return false
}

type selectionDocumentType string

const (
	selectionDocumentTypeContract   selectionDocumentType = "contract"
	selectionDocumentTypeEmployerID selectionDocumentType = "employer_id"
)

func parseSelectionDocumentType(value string) (selectionDocumentType, error) {
	switch strings.TrimSpace(value) {
	case string(selectionDocumentTypeContract):
		return selectionDocumentTypeContract, nil
	case string(selectionDocumentTypeEmployerID):
		return selectionDocumentTypeEmployerID, nil
	default:
		return "", ErrInvalidSelectionDocumentType
	}
}

func validateSelectionDocumentContentType(documentType selectionDocumentType, contentType string) error {
	switch documentType {
	case selectionDocumentTypeContract, selectionDocumentTypeEmployerID:
		if contentType != "application/pdf" && contentType != "image/jpeg" && contentType != "image/png" {
			return ErrInvalidFileType
		}
	default:
		return ErrInvalidSelectionDocumentType
	}
	return nil
}

func selectionHasRequiredSupportingDocuments(selection *domain.Selection) bool {
	if selection == nil {
		return false
	}

	return strings.TrimSpace(selection.EmployerContractURL) != "" && strings.TrimSpace(selection.EmployerIDURL) != ""
}
