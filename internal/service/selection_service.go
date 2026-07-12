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
	ErrCandidateNotLockedByYou              = errors.New("candidate is not locked by you")
)

type NotificationSender interface {
	IsForeignAgent(userID string) (bool, error)
	Send(userID, title, message, notificationType, relatedEntityType, relatedEntityID string) error
}

type SelectionUpdateSender interface {
	PushSelectionUpdate(selectionID, status, action, pairingID string)
}

type dbProvider interface {
	DB() *gorm.DB
}

type batchSelectionRepository interface {
	GetBySelectedByBatch(userID string) ([]*domain.Selection, error)
	GetBySelectedByAndPairingBatch(userID, pairingID string) ([]*domain.Selection, error)
	GetByCandidateOwnerBatch(userID string) ([]*domain.Selection, error)
	GetByCandidateOwnerAndPairingBatch(userID, pairingID string) ([]*domain.Selection, error)
}

type SelectionService struct {
	selectionRepository domain.SelectionRepository
	batchRepo           batchSelectionRepository
	candidateRepository domain.CandidateRepository
	notificationService NotificationSender
	platformSettings    PlatformSettingsReader
	storageService      StorageService
	pairingService      *PairingService
	selectionUpdates    SelectionUpdateSender
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

	batchRepo, _ := selectionRepository.(batchSelectionRepository)

	return &SelectionService{
		selectionRepository: selectionRepository,
		batchRepo:           batchRepo,
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

func (s *SelectionService) SetSelectionUpdateSender(sender SelectionUpdateSender) {
	s.selectionUpdates = sender
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
	expiresAt := time.Now().UTC().Add(30 * 24 * time.Hour)

	err = s.db.Transaction(func(tx *gorm.DB) error {
		var candidate domain.Candidate
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", candidateID).First(&candidate).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrCandidateNotFound
			}
			return fmt.Errorf("load candidate: %w", err)
		}

		canSelect := candidate.Status == domain.CandidateStatusAvailable
		isLockedByMe := candidate.Status == domain.CandidateStatusLocked &&
			candidate.LockedBy != nil &&
			strings.TrimSpace(*candidate.LockedBy) == strings.TrimSpace(selectedBy)
		if !canSelect && !isLockedByMe {
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

		// Only need to lock if not already locked by current user
		if canSelect {
			now := time.Now().UTC()
			lockExpiry := now.Add(30 * 24 * time.Hour)
			result := tx.Model(&domain.Candidate{}).Where("id = ?", candidateID).Updates(map[string]any{
				"status":          domain.CandidateStatusLocked,
				"locked_by":       selectedBy,
				"locked_at":       now,
				"lock_expires_at": lockExpiry,
			})
			if result.Error != nil {
				return fmt.Errorf("lock candidate: %w", result.Error)
			}
			if result.RowsAffected == 0 {
				return ErrCandidateNotAvailable
			}
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
			"You have successfully selected this candidate. The Ethiopian agency will review your selection.",
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

	if s.selectionUpdates != nil && selection != nil {
		s.selectionUpdates.PushSelectionUpdate(selection.ID, string(selection.Status), "select", selection.PairingID)
	}

	return selection, nil
}

func (s *SelectionService) LockCandidate(candidateID, userID, pairingID string) error {
	candidateID = strings.TrimSpace(candidateID)
	userID = strings.TrimSpace(userID)

	if candidateID == "" {
		return fmt.Errorf("candidate id is required")
	}
	if userID == "" {
		return fmt.Errorf("user id is required")
	}

	isForeign, err := s.notificationService.IsForeignAgent(userID)
	if err != nil {
		return fmt.Errorf("validate foreign agent role: %w", err)
	}
	if !isForeign {
		return ErrNotForeignAgent
	}

	var pairing *domain.AgencyPairing
	if s.pairingService != nil {
		pairing, err = s.pairingService.ResolveActivePairing(userID, string(domain.ForeignAgent), pairingID)
		if err != nil {
			return err
		}
	}

	var ownerID string
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

		now := time.Now().UTC()
		lockExpiry := now.Add(30 * 24 * time.Hour)
		result := tx.Model(&domain.Candidate{}).Where("id = ?", candidateID).Updates(map[string]any{
			"status":          domain.CandidateStatusLocked,
			"locked_by":       userID,
			"locked_at":       now,
			"lock_expires_at": lockExpiry,
		})
		if result.Error != nil {
			return fmt.Errorf("lock candidate: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return ErrCandidateNotAvailable
		}

		ownerID = candidate.CreatedBy
		return nil
	})
	if err != nil {
		return err
	}

	// Send notifications after successful commit — failures here don't roll back the lock
	_ = s.notificationService.Send(
		ownerID,
		"Candidate locked",
		"A foreign agent has placed a hold on one of your candidates. The candidate is reserved and cannot be selected by others.",
		"lock",
		"candidate",
		candidateID,
	)
	_ = s.notificationService.Send(
		userID,
		"Candidate locked",
		"You have placed a hold on this candidate. They are now reserved for your agency.",
		"lock",
		"candidate",
		candidateID,
	)

	return nil
}

func (s *SelectionService) UnlockCandidate(candidateID, userID string) error {
	candidateID = strings.TrimSpace(candidateID)
	userID = strings.TrimSpace(userID)

	if candidateID == "" {
		return fmt.Errorf("candidate id is required")
	}
	if userID == "" {
		return fmt.Errorf("user id is required")
	}

	var recipientID string
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var candidate domain.Candidate
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", candidateID).First(&candidate).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrCandidateNotFound
			}
			return fmt.Errorf("load candidate: %w", err)
		}

		if candidate.Status != domain.CandidateStatusLocked {
			return ErrCandidateNotAvailable
		}

		// Allow unlock by: the locker (foreign agent), or the candidate owner (Ethiopian agent)
		if candidate.LockedBy == nil || strings.TrimSpace(*candidate.LockedBy) == "" {
			return ErrCandidateNotLockedByYou
		}
		isLocker := strings.TrimSpace(*candidate.LockedBy) == strings.TrimSpace(userID)
		isOwner := strings.TrimSpace(candidate.CreatedBy) == strings.TrimSpace(userID)
		if !isLocker && !isOwner {
			return ErrNotAuthorized
		}

		if err := tx.Model(&domain.Candidate{}).Where("id = ?", candidateID).Updates(map[string]any{
			"status":          domain.CandidateStatusAvailable,
			"locked_by":       nil,
			"locked_at":       nil,
			"lock_expires_at": nil,
		}).Error; err != nil {
			return fmt.Errorf("unlock candidate: %w", err)
		}

		recipientID = candidate.CreatedBy
		if isOwner {
			recipientID = *candidate.LockedBy
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Send notifications after successful commit — failures here don't roll back the unlock
	if recipientID != "" && recipientID != userID {
		_ = s.notificationService.Send(
			recipientID,
			"Candidate released",
			"The hold on this candidate has been released. They are now available for selection.",
			"lock",
			"candidate",
			candidateID,
		)
	}
	_ = s.notificationService.Send(
		userID,
		"Candidate released",
		"You have released the hold on this candidate.",
		"lock",
		"candidate",
		candidateID,
	)

	return nil
}

func (s *SelectionService) BatchLockCandidates(candidateIDs []string, userID, pairingID string) (int, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return 0, fmt.Errorf("user id is required")
	}

	isForeign, err := s.notificationService.IsForeignAgent(userID)
	if err != nil {
		return 0, fmt.Errorf("validate foreign agent role: %w", err)
	}
	if !isForeign {
		return 0, ErrNotForeignAgent
	}

	var pairing *domain.AgencyPairing
	if s.pairingService != nil {
		pairing, err = s.pairingService.ResolveActivePairing(userID, string(domain.ForeignAgent), pairingID)
		if err != nil {
			return 0, err
		}
	}

	locked := 0
	for _, candidateID := range candidateIDs {
		candidateID = strings.TrimSpace(candidateID)
		if candidateID == "" {
			continue
		}

		err := s.db.Transaction(func(tx *gorm.DB) error {
			var candidate domain.Candidate
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", candidateID).First(&candidate).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				return fmt.Errorf("load candidate: %w", err)
			}

			if candidate.Status != domain.CandidateStatusAvailable {
				return nil
			}

			if pairing != nil {
				if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(pairing.EthiopianUserID) {
					return nil
				}
				isShared, err := s.pairingService.IsCandidateSharedWithPairing(candidate.ID, pairing.ID)
				if err != nil {
					return err
				}
				if !isShared {
					return nil
				}
			}

			now := time.Now().UTC()
			lockExpiry := now.Add(30 * 24 * time.Hour)
			result := tx.Model(&domain.Candidate{}).Where("id = ?", candidateID).Updates(map[string]any{
				"status":          domain.CandidateStatusLocked,
				"locked_by":       userID,
				"locked_at":       now,
				"lock_expires_at": lockExpiry,
			})
			if result.Error != nil || result.RowsAffected == 0 {
				return nil
			}

			if err := s.notificationService.Send(
				candidate.CreatedBy,
				"Candidate locked",
				"A foreign agent has placed a hold on one of your candidates.",
				"lock",
				"candidate",
				candidateID,
			); err != nil {
				return fmt.Errorf("notify candidate owner: %w", err)
			}

			return nil
		})
		if err == nil {
			locked++
		}
	}

	return locked, nil
}

func (s *SelectionService) BatchUnlockCandidates(candidateIDs []string, userID string) (int, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return 0, fmt.Errorf("user id is required")
	}

	unlocked := 0
	for _, candidateID := range candidateIDs {
		candidateID = strings.TrimSpace(candidateID)
		if candidateID == "" {
			continue
		}

		err := s.db.Transaction(func(tx *gorm.DB) error {
			var candidate domain.Candidate
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", candidateID).First(&candidate).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				return fmt.Errorf("load candidate: %w", err)
			}

			if candidate.Status != domain.CandidateStatusLocked {
				return nil
			}

			if candidate.LockedBy == nil || strings.TrimSpace(*candidate.LockedBy) == "" {
				return nil
			}
			isLocker := strings.TrimSpace(*candidate.LockedBy) == strings.TrimSpace(userID)
			isOwner := strings.TrimSpace(candidate.CreatedBy) == strings.TrimSpace(userID)
			if !isLocker && !isOwner {
				return nil
			}

			if err := tx.Model(&domain.Candidate{}).Where("id = ?", candidateID).Updates(map[string]any{
				"status":          domain.CandidateStatusAvailable,
				"locked_by":       nil,
				"locked_at":       nil,
				"lock_expires_at": nil,
			}).Error; err != nil {
				return fmt.Errorf("unlock candidate: %w", err)
			}

			return nil
		})
		if err == nil {
			unlocked++
		}
	}

	return unlocked, nil
}

func (s *SelectionService) UnlockSelection(selectionID, userID string) error {
	selectionID = strings.TrimSpace(selectionID)
	userID = strings.TrimSpace(userID)

	if selectionID == "" {
		return fmt.Errorf("selection id is required")
	}
	if userID == "" {
		return fmt.Errorf("user id is required")
	}

	var pairingID string

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var selection domain.Selection
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", selectionID).First(&selection).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrSelectionNotFound
			}
			return fmt.Errorf("load selection: %w", err)
		}

		if selection.Status != domain.SelectionPending {
			return ErrSelectionNotPending
		}

		pairingID = selection.PairingID

		var candidate domain.Candidate
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", selection.CandidateID).First(&candidate).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrCandidateNotFound
			}
			return fmt.Errorf("load candidate: %w", err)
		}

		if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(userID) {
			return ErrNotAuthorized
		}

		if err := tx.Model(&domain.Selection{}).Where("id = ?", selectionID).Update("status", domain.SelectionReleased).Error; err != nil {
			return fmt.Errorf("update selection released: %w", err)
		}

		if err := tx.Model(&domain.Candidate{}).Where("id = ?", selection.CandidateID).Updates(map[string]any{
			"status":          domain.CandidateStatusAvailable,
			"locked_by":       nil,
			"locked_at":       nil,
			"lock_expires_at": nil,
		}).Error; err != nil {
			return fmt.Errorf("unlock candidate: %w", err)
		}

		if err := s.notificationService.Send(
			candidate.CreatedBy,
			"Candidate unlocked",
			"You have released this candidate. They are now available for other selections.",
			"selection",
			"selection",
			selectionID,
		); err != nil {
			return fmt.Errorf("notify owner: %w", err)
		}

		if err := s.notificationService.Send(
			selection.SelectedBy,
			"Selection released",
			"The Ethiopian agency has released this candidate. The candidate is no longer selected by your agency.",
			"selection",
			"selection",
			selectionID,
		); err != nil {
			return fmt.Errorf("notify foreign agent: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if s.selectionUpdates != nil {
		s.selectionUpdates.PushSelectionUpdate(selectionID, string(domain.SelectionReleased), "release", pairingID)
	}

	return nil
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
	if s.batchRepo != nil {
		return s.batchRepo.GetBySelectedByBatch(userID)
	}
	return s.selectionRepository.GetBySelectedBy(userID)
}

func (s *SelectionService) GetSelectionsForUser(userID, role string) ([]*domain.Selection, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user id is required")
	}

	switch strings.TrimSpace(role) {
	case string(domain.ForeignAgent):
		if s.batchRepo != nil {
			return s.batchRepo.GetBySelectedByBatch(userID)
		}
		return s.selectionRepository.GetBySelectedBy(userID)
	case string(domain.EthiopianAgent):
		if s.batchRepo != nil {
			return s.batchRepo.GetByCandidateOwnerBatch(userID)
		}
		return s.selectionRepository.GetByCandidateOwner(userID)
	default:
		return nil, ErrForbidden
	}
}

func (s *SelectionService) ListSelections(filters domain.SelectionFilters) ([]*domain.Selection, int64, error) {
	total, err := s.selectionRepository.Count(filters)
	if err != nil {
		return nil, 0, err
	}
	selections, err := s.selectionRepository.List(filters)
	if err != nil {
		return nil, 0, err
	}
	return selections, total, nil
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
		if s.batchRepo != nil {
			return s.batchRepo.GetBySelectedByAndPairingBatch(userID, pairing.ID)
		}
		return s.selectionRepository.GetBySelectedByAndPairing(userID, pairing.ID)
	case string(domain.EthiopianAgent):
		if s.batchRepo != nil {
			return s.batchRepo.GetByCandidateOwnerAndPairingBatch(userID, pairing.ID)
		}
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

			if s.selectionUpdates != nil {
				s.selectionUpdates.PushSelectionUpdate(selectionID, string(domain.SelectionExpired), "expire", selection.PairingID)
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

	return strings.TrimSpace(selection.EmployerContractURL) != ""
}
