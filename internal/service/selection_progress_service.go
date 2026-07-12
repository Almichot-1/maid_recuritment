package service

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

var (
	ErrProgressNotFound        = errors.New("progress not found")
	ErrProgressAlreadyExists   = errors.New("progress already exists for selection")
	ErrInvalidProgressDocument = errors.New("invalid progress document type")
	ErrProgressUpdateForbidden = errors.New("only ethiopian agents can update progress")
)

type ProgressUpdateSender interface {
	PushProgressUpdate(selectionID, stepName, status string)
}

// overall status constants that mirror CandidateStatus
const (
	candidateStatusApproved   = "approved"
	candidateStatusInProgress = "in_progress"
	candidateStatusCompleted  = "completed"
)

type SelectionProgressService struct {
	progressRepository  domain.SelectionProgressRepository
	selectionRepository domain.SelectionRepository
	candidateRepository domain.CandidateRepository
	notificationService NotificationSender
	storageService      StorageService
	progressUpdates     ProgressUpdateSender
	db                  *gorm.DB
}

func (s *SelectionProgressService) SetProgressUpdateSender(sender ProgressUpdateSender) {
	s.progressUpdates = sender
}

func NewSelectionProgressService(
	progressRepository domain.SelectionProgressRepository,
	selectionRepository domain.SelectionRepository,
	candidateRepository domain.CandidateRepository,
	notificationService NotificationSender,
) (*SelectionProgressService, error) {
	if progressRepository == nil {
		return nil, fmt.Errorf("progress repository is nil")
	}
	if selectionRepository == nil {
		return nil, fmt.Errorf("selection repository is nil")
	}
	if candidateRepository == nil {
		return nil, fmt.Errorf("candidate repository is nil")
	}
	if notificationService == nil {
		return nil, fmt.Errorf("notification service is nil")
	}

	dbSource, ok := progressRepository.(dbProvider)
	if !ok || dbSource.DB() == nil {
		return nil, fmt.Errorf("progress repository does not expose transaction db")
	}

	return &SelectionProgressService{
		progressRepository:  progressRepository,
		selectionRepository: selectionRepository,
		candidateRepository: candidateRepository,
		notificationService: notificationService,
		db:                  dbSource.DB(),
	}, nil
}

func (s *SelectionProgressService) SetStorageService(storageService StorageService) {
	s.storageService = storageService
}

type UpdateProgressInput struct {
	// COC fields
	COCStatus *string `json:"coc_status,omitempty"`
	COCType   *string `json:"coc_type,omitempty"`

	// Medical fields
	MedicalStatus *string `json:"medical_status,omitempty"`

	// Visa fields
	VisaStatus *string `json:"visa_status,omitempty"`

	// Ticket fields
	TicketStatus *string `json:"ticket_status,omitempty"`

	// Arrival fields
	ArrivalStatus      *string    `json:"arrival_status,omitempty"`
	ArrivalDate        *time.Time `json:"arrival_date,omitempty"`
	ArrivalCity        *string    `json:"arrival_city,omitempty"`
	DestinationCountry *string    `json:"destination_country,omitempty"`
	DepartureDate      *time.Time `json:"departure_date,omitempty"`
	Notes              *string    `json:"notes,omitempty"`
}

type UploadProgressDocumentInput struct {
	DocumentType string
	File         io.Reader
	FileName     string
	FileSize     int64
}

// CreateProgress creates a new progress record when a selection is approved
func (s *SelectionProgressService) CreateProgress(selectionID, updatedBy string) (*domain.SelectionProgress, error) {
	selectionID = strings.TrimSpace(selectionID)
	updatedBy = strings.TrimSpace(updatedBy)

	if selectionID == "" {
		return nil, fmt.Errorf("selection id is required")
	}
	if updatedBy == "" {
		return nil, fmt.Errorf("updated by is required")
	}

	// Check if progress already exists
	existing, err := s.progressRepository.GetBySelectionID(selectionID)
	if err != nil && !errors.Is(err, repository.ErrProgressNotFound) {
		return nil, fmt.Errorf("check existing progress: %w", err)
	}
	if existing != nil {
		return existing, nil // Already exists, return it
	}

	progress := &domain.SelectionProgress{
		ID:          uuid.NewString(),
		SelectionID: selectionID,
		UpdatedBy:   updatedBy,
		// All statuses default to "pending" (see domain constants)
		COCStatus:     domain.ProgressStatusPending,
		MedicalStatus: domain.ProgressStatusPending,
		VisaStatus:    domain.VisaStatusPending,
		TicketStatus:  domain.TicketStatusPending,
		ArrivalStatus: domain.ArrivalStatusNotArrived,
	}

	if err := s.progressRepository.Create(progress); err != nil {
		return nil, fmt.Errorf("create progress: %w", err)
	}

	return progress, nil
}

// GetProgress retrieves progress for a selection
func (s *SelectionProgressService) GetProgress(selectionID string) (*domain.SelectionProgress, error) {
	if strings.TrimSpace(selectionID) == "" {
		return nil, fmt.Errorf("selection id is required")
	}

	progress, err := s.progressRepository.GetBySelectionID(selectionID)
	if err != nil {
		if errors.Is(err, repository.ErrProgressNotFound) {
			return nil, ErrProgressNotFound
		}
		return nil, fmt.Errorf("get progress: %w", err)
	}

	return progress, nil
}

// UpdateProgress updates progress fields (Ethiopian agents only)
func (s *SelectionProgressService) UpdateProgress(selectionID, userID string, input UpdateProgressInput) (*domain.SelectionProgress, error) {
	selectionID = strings.TrimSpace(selectionID)
	userID = strings.TrimSpace(userID)

	if selectionID == "" {
		return nil, fmt.Errorf("selection id is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}

	var progress *domain.SelectionProgress
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Load selection to verify ownership
		selection, err := s.selectionRepository.GetByID(selectionID)
		if err != nil {
			if errors.Is(err, repository.ErrSelectionNotFound) {
				return repository.ErrSelectionNotFound
			}
			return fmt.Errorf("load selection: %w", err)
		}

		// Load candidate to check ownership
		candidate, err := s.candidateRepository.GetByID(selection.CandidateID)
		if err != nil {
			if errors.Is(err, repository.ErrCandidateNotFound) {
				return repository.ErrCandidateNotFound
			}
			return fmt.Errorf("load candidate: %w", err)
		}

		// Only Ethiopian agent (candidate owner) can update
		if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(userID) {
			return ErrProgressUpdateForbidden
		}

		// Get existing progress
		var existingProgress domain.SelectionProgress
		if err := tx.Where("selection_id = ?", selectionID).First(&existingProgress).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create progress directly inside the transaction to avoid deadlocks
				newProgress := &domain.SelectionProgress{
					ID:            uuid.NewString(),
					SelectionID:   selectionID,
					UpdatedBy:     userID,
					COCStatus:     domain.ProgressStatusPending,
					MedicalStatus: domain.ProgressStatusPending,
					VisaStatus:    domain.VisaStatusPending,
					TicketStatus:  domain.TicketStatusPending,
					ArrivalStatus: domain.ArrivalStatusNotArrived,
				}
				if err := tx.Create(newProgress).Error; err != nil {
					return fmt.Errorf("auto-create progress: %w", err)
				}
				existingProgress = *newProgress
			} else {
				return fmt.Errorf("load progress: %w", err)
			}
		}

		// Build updates map
		updates := make(map[string]interface{})

		// COC fields
		if input.COCStatus != nil {
			status := strings.TrimSpace(*input.COCStatus)
			if !isValidProgressStatus(status) {
				return fmt.Errorf("invalid coc status: %s", status)
			}
			updates["coc_status"] = status
		}
		if input.COCType != nil {
			cocType := strings.TrimSpace(*input.COCType)
			if cocType != "" && cocType != string(domain.COCTypeOnline) && cocType != string(domain.COCTypeOffline) {
				return fmt.Errorf("invalid coc type: %s", cocType)
			}
			if cocType == "" {
				updates["coc_type"] = nil
			} else {
				updates["coc_type"] = cocType
			}
		}

		// Medical fields
		if input.MedicalStatus != nil {
			status := strings.TrimSpace(*input.MedicalStatus)
			if !isValidProgressStatus(status) {
				return fmt.Errorf("invalid medical status: %s", status)
			}
			updates["medical_status"] = status
		}

		// Visa fields
		if input.VisaStatus != nil {
			status := strings.TrimSpace(*input.VisaStatus)
			if !isValidVisaStatus(status) {
				return fmt.Errorf("invalid visa status: %s", status)
			}
			updates["visa_status"] = status
		}

		// Ticket fields
		if input.TicketStatus != nil {
			status := strings.TrimSpace(*input.TicketStatus)
			if !isValidTicketStatus(status) {
				return fmt.Errorf("invalid ticket status: %s", status)
			}
			updates["ticket_status"] = status
		}

		// Arrival fields
		if input.ArrivalStatus != nil {
			status := strings.TrimSpace(*input.ArrivalStatus)
			if !isValidArrivalStatus(status) {
				return fmt.Errorf("invalid arrival status: %s", status)
			}
			updates["arrival_status"] = status
		}
		if input.ArrivalDate != nil {
			updates["arrival_date"] = *input.ArrivalDate
		}
		if input.ArrivalCity != nil {
			city := strings.TrimSpace(*input.ArrivalCity)
			if city == "" {
				updates["arrival_city"] = nil
			} else {
				updates["arrival_city"] = city
			}
		}
		if input.DestinationCountry != nil {
			country := strings.TrimSpace(*input.DestinationCountry)
			if country == "" {
				updates["destination_country"] = nil
			} else {
				updates["destination_country"] = country
			}
		}
		if input.DepartureDate != nil {
			updates["departure_date"] = *input.DepartureDate
		}

		if input.Notes != nil {
			updates["notes"] = strings.TrimSpace(*input.Notes)
		}

		if len(updates) == 0 {
			// No updates to apply
			progress = &existingProgress
			return nil
		}

		// Apply updates
		if err := tx.Model(&domain.SelectionProgress{}).Where("selection_id = ?", selectionID).Updates(updates).Error; err != nil {
			return fmt.Errorf("update progress: %w", err)
		}

		// Reload updated progress
		if err := tx.Where("selection_id = ?", selectionID).First(&existingProgress).Error; err != nil {
			return fmt.Errorf("reload progress: %w", err)
		}

		progress = &existingProgress

		// Update the candidate's overall status based on progress step statuses
		allCompleted := isAllProgressCompleted(&existingProgress)
		hasStarted := isAnyProgressStarted(&existingProgress)

		nextCandidateStatus := candidateStatusApproved
		if allCompleted {
			nextCandidateStatus = candidateStatusCompleted
		} else if hasStarted {
			nextCandidateStatus = candidateStatusInProgress
		}

		if err := tx.Model(&domain.Candidate{}).
			Where("id = ?", selection.CandidateID).
			Update("status", nextCandidateStatus).Error; err != nil {
			return fmt.Errorf("update candidate status after progress change: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Send specific "Arrived" notification when arrival status transitions to arrived
	if input.ArrivalStatus != nil && strings.TrimSpace(*input.ArrivalStatus) == domain.ArrivalStatusArrived {
		selection, selErr := s.selectionRepository.GetByID(selectionID)
		if selErr == nil && selection != nil {
			arrivalCity := ""
			if progress.ArrivalCity != nil {
				arrivalCity = strings.TrimSpace(*progress.ArrivalCity)
			}
			destCountry := ""
			if progress.DestinationCountry != nil {
				destCountry = strings.TrimSpace(*progress.DestinationCountry)
			}
			arrivalMsg := fmt.Sprintf("Candidate has arrived in %s.", arrivalCity)
			if arrivalCity != "" && destCountry != "" {
				arrivalMsg = fmt.Sprintf("Candidate has arrived in %s, %s.", arrivalCity, destCountry)
			}
			_ = s.notificationService.Send(
				progress.UpdatedBy,
				"Candidate arrived",
				arrivalMsg,
				string(domain.NotificationArrived),
				"selection",
				selectionID,
			)
			_ = s.notificationService.Send(
				selection.SelectedBy,
				"Candidate arrived",
				arrivalMsg,
				string(domain.NotificationArrived),
				"selection",
				selectionID,
			)
		}
	}

	if err != nil {
		return nil, err
	}

	// Send per-step update notification to foreign agent
	if selection, selErr := s.selectionRepository.GetByID(selectionID); selErr == nil && selection != nil {
		var parts []string
		if input.COCStatus != nil { parts = append(parts, "COC: "+strings.TrimSpace(*input.COCStatus)) }
		if input.MedicalStatus != nil { parts = append(parts, "Medical: "+strings.TrimSpace(*input.MedicalStatus)) }
		if input.VisaStatus != nil { parts = append(parts, "Visa: "+strings.TrimSpace(*input.VisaStatus)) }
		if input.TicketStatus != nil { parts = append(parts, "Ticket: "+strings.TrimSpace(*input.TicketStatus)) }
		if input.ArrivalStatus != nil { parts = append(parts, "Arrival: "+strings.TrimSpace(*input.ArrivalStatus)) }
		if len(parts) > 0 {
			_ = s.notificationService.Send(
				selection.SelectedBy,
				"Recruitment progress updated",
				"Progress updated: "+strings.Join(parts, ", "),
				string(domain.NotificationStatusUpdate),
				"selection",
				selectionID,
			)
		}
	}

	// Push real-time progress updates via WebSocket
	if s.progressUpdates != nil {
		if input.COCStatus != nil { s.progressUpdates.PushProgressUpdate(selectionID, "CoC", strings.TrimSpace(*input.COCStatus)) }
		if input.MedicalStatus != nil { s.progressUpdates.PushProgressUpdate(selectionID, "Medical", strings.TrimSpace(*input.MedicalStatus)) }
		if input.VisaStatus != nil { s.progressUpdates.PushProgressUpdate(selectionID, "Visa", strings.TrimSpace(*input.VisaStatus)) }
		if input.TicketStatus != nil { s.progressUpdates.PushProgressUpdate(selectionID, "Ticket", strings.TrimSpace(*input.TicketStatus)) }
		if input.ArrivalStatus != nil { s.progressUpdates.PushProgressUpdate(selectionID, "Arrival", strings.TrimSpace(*input.ArrivalStatus)) }
	}

	// Flight booked notification
	if input.TicketStatus != nil && strings.TrimSpace(*input.TicketStatus) == string(domain.TicketStatusConfirmed) {
		if selection, selErr := s.selectionRepository.GetByID(selectionID); selErr == nil && selection != nil {
			_ = s.notificationService.Send(
				selection.SelectedBy,
				"Flight booked",
				"The flight has been booked for this candidate.",
				string(domain.NotificationFlightBooked),
				"selection",
				selectionID,
			)
			candidate, candErr := s.candidateRepository.GetByID(selection.CandidateID)
			if candErr == nil && candidate != nil {
				_ = s.notificationService.Send(
					candidate.CreatedBy,
					"Flight booked",
					"The flight has been booked for this candidate.",
					string(domain.NotificationFlightBooked),
					"selection",
					selectionID,
				)
			}
		}
	}

	// Recruitment completed notification
	if isAllProgressCompleted(progress) {
		if selection, selErr := s.selectionRepository.GetByID(selectionID); selErr == nil && selection != nil {
			candidate, candErr := s.candidateRepository.GetByID(selection.CandidateID)
			if candErr == nil && candidate != nil {
				_ = s.notificationService.Send(
					candidate.CreatedBy,
					"Recruitment completed",
					"All recruitment steps have been completed for this candidate.",
					string(domain.NotificationStatusUpdate),
					"selection",
					selectionID,
				)
			}
			_ = s.notificationService.Send(
				selection.SelectedBy,
				"Recruitment completed",
				"All recruitment steps have been completed for this candidate.",
				string(domain.NotificationStatusUpdate),
				"selection",
				selectionID,
			)
		}
	}

	return progress, nil
}

type ProgressBatchUpdateResult struct {
	SelectionID string `json:"selection_id"`
	Error       error  `json:"error,omitempty"`
}

// BatchUpdateProgress applies the same update to multiple selections (Ethiopian agent only)
func (s *SelectionProgressService) BatchUpdateProgress(selectionIDs []string, userID string, input UpdateProgressInput) []ProgressBatchUpdateResult {
	results := make([]ProgressBatchUpdateResult, 0, len(selectionIDs))
	for _, selectionID := range selectionIDs {
		_, err := s.UpdateProgress(selectionID, userID, input)
		results = append(results, ProgressBatchUpdateResult{
			SelectionID: selectionID,
			Error:       err,
		})
	}
	return results
}

// UploadDocument uploads a document for a specific progress field
func (s *SelectionProgressService) UploadDocument(selectionID, userID string, input UploadProgressDocumentInput) (*domain.SelectionProgress, error) {
	selectionID = strings.TrimSpace(selectionID)
	userID = strings.TrimSpace(userID)

	if selectionID == "" {
		return nil, fmt.Errorf("selection id is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
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

	documentType, err := parseProgressDocumentType(input.DocumentType)
	if err != nil {
		return nil, err
	}

	bufferedFile, contentType, err := validateAndBufferUpload(input.File, input.FileName)
	if err != nil {
		return nil, err
	}
	if err := validateProgressDocumentContentType(contentType); err != nil {
		return nil, err
	}

	var uploadedURL string
	var progress *domain.SelectionProgress

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Load selection to verify ownership
		selection, err := s.selectionRepository.GetByID(selectionID)
		if err != nil {
			return fmt.Errorf("load selection: %w", err)
		}

		// Load candidate to check ownership
		candidate, err := s.candidateRepository.GetByID(selection.CandidateID)
		if err != nil {
			return fmt.Errorf("load candidate: %w", err)
		}

		// Only Ethiopian agent can upload
		if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(userID) {
			return ErrProgressUpdateForbidden
		}

		// Upload file
		uploadedURL, err = s.storageService.Upload(bufferedFile, input.FileName, contentType)
		if err != nil {
			return fmt.Errorf("upload document: %w", err)
		}

		// Load existing progress
		var existingProgress domain.SelectionProgress
		if err := tx.Where("selection_id = ?", selectionID).First(&existingProgress).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				_ = s.storageService.Delete(uploadedURL)
				return repository.ErrProgressNotFound
			}
			_ = s.storageService.Delete(uploadedURL)
			return fmt.Errorf("load progress: %w", err)
		}

		// Build updates based on document type
		updates := make(map[string]interface{})
		now := time.Now().UTC()
		var previousURL string

		switch documentType {
		case progressDocTypeCOC:
			previousURL = existingProgress.COCDocumentURL
			updates["coc_document_url"] = uploadedURL
			updates["coc_document_name"] = input.FileName
			updates["coc_uploaded_at"] = now
		case progressDocTypeMedical:
			previousURL = existingProgress.MedicalDocumentURL
			updates["medical_document_url"] = uploadedURL
			updates["medical_document_name"] = input.FileName
			updates["medical_uploaded_at"] = now
		case progressDocTypeVisa:
			previousURL = existingProgress.VisaDocumentURL
			updates["visa_document_url"] = uploadedURL
			updates["visa_document_name"] = input.FileName
			updates["visa_uploaded_at"] = now
		case progressDocTypeTicket:
			previousURL = existingProgress.TicketDocumentURL
			updates["ticket_document_url"] = uploadedURL
			updates["ticket_document_name"] = input.FileName
			updates["ticket_uploaded_at"] = now
		case progressDocTypeArrival:
			previousURL = existingProgress.ArrivalDocumentURL
			updates["arrival_document_url"] = uploadedURL
			updates["arrival_document_name"] = input.FileName
			updates["arrival_uploaded_at"] = now
		default:
			_ = s.storageService.Delete(uploadedURL)
			return ErrInvalidProgressDocument
		}

		// Apply updates
		if err := tx.Model(&domain.SelectionProgress{}).Where("selection_id = ?", selectionID).Updates(updates).Error; err != nil {
			_ = s.storageService.Delete(uploadedURL)
			return fmt.Errorf("update progress document: %w", err)
		}

		// Delete previous document if it exists
		if strings.TrimSpace(previousURL) != "" && previousURL != uploadedURL {
			_ = s.storageService.Delete(previousURL)
		}

		// Reload progress
		if err := tx.Where("selection_id = ?", selectionID).First(&existingProgress).Error; err != nil {
			return fmt.Errorf("reload progress: %w", err)
		}

		progress = &existingProgress
		return nil
	})

	if err != nil {
		if uploadedURL != "" {
			_ = s.storageService.Delete(uploadedURL)
		}
		return nil, err
	}

	return progress, nil
}

// DeleteDocument removes a document from a progress field
func (s *SelectionProgressService) DeleteDocument(selectionID, userID string, documentType string) (*domain.SelectionProgress, error) {
	selectionID = strings.TrimSpace(selectionID)
	userID = strings.TrimSpace(userID)

	if selectionID == "" {
		return nil, fmt.Errorf("selection id is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}

	docType, err := parseProgressDocumentType(documentType)
	if err != nil {
		return nil, err
	}

	var progress *domain.SelectionProgress
	err = s.db.Transaction(func(tx *gorm.DB) error {
		selection, err := s.selectionRepository.GetByID(selectionID)
		if err != nil {
			return fmt.Errorf("load selection: %w", err)
		}

		candidate, err := s.candidateRepository.GetByID(selection.CandidateID)
		if err != nil {
			return fmt.Errorf("load candidate: %w", err)
		}

		if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(userID) {
			return ErrProgressUpdateForbidden
		}

		var existingProgress domain.SelectionProgress
		if err := tx.Where("selection_id = ?", selectionID).First(&existingProgress).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrProgressNotFound
			}
			return fmt.Errorf("load progress: %w", err)
		}

		updates := make(map[string]interface{})

		switch docType {
		case progressDocTypeCOC:
			if existingProgress.COCDocumentURL != "" {
				_ = s.storageService.Delete(existingProgress.COCDocumentURL)
			}
			updates["coc_document_url"] = nil
			updates["coc_document_name"] = nil
			updates["coc_uploaded_at"] = nil
		case progressDocTypeMedical:
			if existingProgress.MedicalDocumentURL != "" {
				_ = s.storageService.Delete(existingProgress.MedicalDocumentURL)
			}
			updates["medical_document_url"] = nil
			updates["medical_document_name"] = nil
			updates["medical_uploaded_at"] = nil
		case progressDocTypeVisa:
			if existingProgress.VisaDocumentURL != "" {
				_ = s.storageService.Delete(existingProgress.VisaDocumentURL)
			}
			updates["visa_document_url"] = nil
			updates["visa_document_name"] = nil
			updates["visa_uploaded_at"] = nil
		case progressDocTypeTicket:
			if existingProgress.TicketDocumentURL != "" {
				_ = s.storageService.Delete(existingProgress.TicketDocumentURL)
			}
			updates["ticket_document_url"] = nil
			updates["ticket_document_name"] = nil
			updates["ticket_uploaded_at"] = nil
		case progressDocTypeArrival:
			if existingProgress.ArrivalDocumentURL != "" {
				_ = s.storageService.Delete(existingProgress.ArrivalDocumentURL)
			}
			updates["arrival_document_url"] = nil
			updates["arrival_document_name"] = nil
			updates["arrival_uploaded_at"] = nil
		default:
			return ErrInvalidProgressDocument
		}

		if err := tx.Model(&domain.SelectionProgress{}).Where("selection_id = ?", selectionID).Updates(updates).Error; err != nil {
			return fmt.Errorf("update progress document: %w", err)
		}

		if err := tx.Where("selection_id = ?", selectionID).First(&existingProgress).Error; err != nil {
			return fmt.Errorf("reload progress: %w", err)
		}

		progress = &existingProgress
		return nil
	})

	if err != nil {
		return nil, err
	}

	return progress, nil
}

// SendProgressNotification sends a summary notification to the foreign agent
func (s *SelectionProgressService) SendProgressNotification(selectionID string) error {
	if strings.TrimSpace(selectionID) == "" {
		return fmt.Errorf("selection id is required")
	}

	selection, err := s.selectionRepository.GetByID(selectionID)
	if err != nil {
		return fmt.Errorf("load selection: %w", err)
	}

	progress, err := s.progressRepository.GetBySelectionID(selectionID)
	if err != nil {
		return fmt.Errorf("load progress: %w", err)
	}

	// Build summary message
	message := fmt.Sprintf("Progress update: COC (%s), Medical (%s), Visa (%s), Ticket (%s), Arrival (%s)",
		progress.COCStatus,
		progress.MedicalStatus,
		progress.VisaStatus,
		progress.TicketStatus,
		progress.ArrivalStatus,
	)

	// Send to foreign agent
	if err := s.notificationService.Send(
		selection.SelectedBy,
		"Selection Progress Update",
		message,
		string(domain.NotificationStatusUpdate),
		"selection",
		selectionID,
	); err != nil {
		return fmt.Errorf("send notification: %w", err)
	}

	return nil
}

type progressDocumentType string

const (
	progressDocTypeCOC     progressDocumentType = "coc"
	progressDocTypeMedical progressDocumentType = "medical"
	progressDocTypeVisa    progressDocumentType = "visa"
	progressDocTypeTicket  progressDocumentType = "ticket"
	progressDocTypeArrival progressDocumentType = "arrival"
)

func parseProgressDocumentType(value string) (progressDocumentType, error) {
	switch strings.TrimSpace(value) {
	case string(progressDocTypeCOC):
		return progressDocTypeCOC, nil
	case string(progressDocTypeMedical):
		return progressDocTypeMedical, nil
	case string(progressDocTypeVisa):
		return progressDocTypeVisa, nil
	case string(progressDocTypeTicket):
		return progressDocTypeTicket, nil
	case string(progressDocTypeArrival):
		return progressDocTypeArrival, nil
	default:
		return "", ErrInvalidProgressDocument
	}
}

func validateProgressDocumentContentType(contentType string) error {
	if contentType != "application/pdf" && contentType != "image/jpeg" && contentType != "image/png" {
		return ErrInvalidFileType
	}
	return nil
}

func IsProgressStepCompleted(status string) bool {
	switch status {
	case domain.ProgressStatusDone, domain.VisaStatusApproved, domain.TicketStatusConfirmed, "arrived":
		return true
	default:
		return false
	}
}

func isProgressStepStarted(status string) bool {
	switch status {
	case domain.ProgressStatusInProgress, domain.ProgressStatusDone, domain.ProgressStatusFailed,
		domain.VisaStatusApproved, domain.VisaStatusRejected,
		domain.TicketStatusBooked, domain.TicketStatusConfirmed, "arrived",
		domain.ArrivalStatusInTransit:
		return true
	default:
		return false
	}
}

func isAllProgressCompleted(p *domain.SelectionProgress) bool {
	return IsProgressStepCompleted(p.COCStatus) &&
		IsProgressStepCompleted(p.MedicalStatus) &&
		IsProgressStepCompleted(p.VisaStatus) &&
		IsProgressStepCompleted(p.TicketStatus) &&
		IsProgressStepCompleted(p.ArrivalStatus)
}

func isAnyProgressStarted(p *domain.SelectionProgress) bool {
	return isProgressStepStarted(p.COCStatus) ||
		isProgressStepStarted(p.MedicalStatus) ||
		isProgressStepStarted(p.VisaStatus) ||
		isProgressStepStarted(p.TicketStatus) ||
		isProgressStepStarted(p.ArrivalStatus)
}

func isValidProgressStatus(status string) bool {
	switch status {
	case domain.ProgressStatusPending, domain.ProgressStatusInProgress, domain.ProgressStatusDone, domain.ProgressStatusFailed:
		return true
	default:
		return false
	}
}

func isValidVisaStatus(status string) bool {
	switch status {
	case domain.VisaStatusPending, domain.VisaStatusInProgress, domain.VisaStatusApproved, domain.VisaStatusRejected:
		return true
	default:
		return false
	}
}

func isValidTicketStatus(status string) bool {
	switch status {
	case domain.TicketStatusPending, domain.TicketStatusBooked, domain.TicketStatusConfirmed, domain.TicketStatusArrived:
		return true
	default:
		return false
	}
}

func isValidArrivalStatus(status string) bool {
	switch status {
	case domain.ArrivalStatusNotArrived, domain.ArrivalStatusInTransit, domain.ArrivalStatusArrived:
		return true
	default:
		return false
	}
}
